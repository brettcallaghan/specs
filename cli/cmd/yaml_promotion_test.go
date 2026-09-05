// Differential tests holding renderYAML's field resolution to
// `encoding/json`'s. [[cli-output-standards]] requires the YAML and JSON
// renderings of one answer to carry "the same values in the same order",
// and the only way that is true by construction is for the YAML renderer
// to resolve a struct's fields exactly as `encoding/json` resolves them —
// json tags, embedded-field promotion, and the dominance rules that pick
// between two fields of the same name.
//
// Each case marshals one value both ways and compares the KEY SEQUENCE.
// A shape whose two renderings disagree on a key, or on the order of
// two keys, is the drift these tests exist to catch.
package cmd

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// jsonKeys reads the top-level object keys `encoding/json` writes, in the
// order it writes them.
func jsonKeys(t *testing.T, value any) []string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encoding/json refused the value: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		t.Fatalf("expected a JSON object, got %v (%v)", tok, err)
	}
	var keys []string
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("reading %s: %v", encoded, err)
		}
		if delim, ok := tok.(json.Delim); ok {
			switch delim {
			case '{', '[':
				depth++
			case '}', ']':
				if depth == 0 {
					return keys
				}
				depth--
			}
			continue
		}
		if depth > 0 {
			continue
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("expected a key in %s, got %v", encoded, tok)
		}
		keys = append(keys, key)
		// Consume the value; a composite one is counted by the delimiter
		// branch above.
		if !dec.More() {
			continue
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("reading the value of %q in %s: %v", key, encoded, err)
		}
	}
}

// yamlKeys reads the top-level mapping keys renderYAML writes, in the
// order it writes them. A top-level key is the only thing the renderer
// puts in column zero.
func yamlKeys(t *testing.T, value any) []string {
	t.Helper()
	var keys []string
	for _, line := range strings.Split(renderYAML(value), "\n") {
		if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") {
			continue
		}
		key, _, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

// PromotionExported is embedded under an EXPORTED name, so the embedded
// field itself is exported.
type PromotionExported struct {
	A string `json:"a"`
}

type promotionUnexported struct {
	U string `json:"u"`
	// An unexported field of an embedded unexported type is invisible to
	// both renderings.
	hidden string //nolint:unused // asserted absent from both renderings
}

type PromotionTagged struct {
	T string `json:"t"`
}

type PromotionShadowed struct {
	Name string `json:"name"`
	Deep string `json:"deep"`
}

// PromotionLeft and PromotionRight both offer a field resolving to the
// name "Clash" at the same depth, neither of them named by a tag, so
// Go's embedding rules hide both and neither rendering carries the key.
//
// The name comes from the Go field rather than from a json tag on both
// sides because `go vet`'s structtag analyzer rejects two promoted
// fields repeating one json tag — the shape cannot be written
// statically, and the rule under test does not distinguish the two ways
// a field arrives at its name.
type PromotionLeft struct {
	Clash string
	Left  string `json:"left"`
}

type PromotionRight struct {
	Clash string
	Right string `json:"right"`
}

// PromotionTaggedClash and PromotionUntaggedClash both offer a field
// resolving to the name "Clash" at the same depth — one from a json tag,
// one from the Go field name. encoding/json lets the tagged one win.
type PromotionTaggedClash struct {
	Clash string `json:"Clash"`
	OnlyT string `json:"only_tagged"`
}

type PromotionUntaggedClash struct {
	Clash string
	OnlyU string `json:"only_untagged"`
}

func TestRenderYAMLResolvesFieldsAsEncodingJSONDoes(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{
			// An embedded exported struct flattens into its parent.
			name: "embedded exported struct",
			value: struct {
				PromotionExported
				B string `json:"b"`
			}{PromotionExported{A: "x"}, "y"},
		},
		{
			// An embedded struct of an UNEXPORTED type still has its
			// exported fields promoted: the type is invisible, the
			// fields are not.
			name: "embedded unexported struct type",
			value: struct {
				promotionUnexported
				B string `json:"b"`
			}{promotionUnexported{U: "x"}, "y"},
		},
		{
			// An embedded field carrying a json NAME is not flattened;
			// it renders as one key under that name.
			name: "embedded struct with a json tag",
			value: struct {
				PromotionTagged `json:"tagged"`
				B               string `json:"b"`
			}{PromotionTagged{T: "x"}, "y"},
		},
		{
			// A tag with options but no name leaves the name empty, and
			// an embedded field with an empty name still flattens.
			name: "embedded struct tagged with options only",
			value: struct {
				PromotionTagged `json:",omitempty"`
				B               string `json:"b"`
			}{PromotionTagged{T: "x"}, "y"},
		},
		{
			name: "embedded pointer to struct, set",
			value: struct {
				*PromotionExported
				B string `json:"b"`
			}{&PromotionExported{A: "x"}, "y"},
		},
		{
			// A nil embedded pointer contributes no key at all.
			name: "embedded pointer to struct, nil",
			value: struct {
				*PromotionExported
				B string `json:"b"`
			}{nil, "y"},
		},
		{
			name: "embedded struct marked json:\"-\"",
			value: struct {
				PromotionExported `json:"-"`
				B                 string `json:"b"`
			}{PromotionExported{A: "x"}, "y"},
		},
		{
			name: "field marked json:\"-\"",
			value: struct {
				A string `json:"-"`
				B string `json:"b"`
			}{"x", "y"},
		},
		{
			name: "omitempty on an empty field",
			value: struct {
				A string   `json:"a,omitempty"`
				B []string `json:"b,omitempty"`
				C string   `json:"c"`
			}{"", nil, "z"},
		},
		{
			name: "omitempty on a populated field",
			value: struct {
				A string   `json:"a,omitempty"`
				B []string `json:"b,omitempty"`
				C string   `json:"c"`
			}{"x", []string{"y"}, "z"},
		},
		{
			name: "untagged field renders under its Go name",
			value: struct {
				Untagged string
				B        string `json:"b"`
			}{"x", "y"},
		},
		{
			name: "unexported non-embedded field is invisible to both",
			value: struct {
				a string //nolint:unused // asserted absent from both renderings
				B string `json:"b"`
			}{"x", "y"},
		},
		{
			// A field at the shallower depth hides the promoted one of
			// the same name, and keeps its own declaration position.
			name: "shallow field dominates a promoted one",
			value: struct {
				PromotionShadowed
				Name string `json:"name"`
			}{PromotionShadowed{Name: "inner", Deep: "d"}, "outer"},
		},
		{
			// Two embedded structs at the same depth both offering
			// "clash", neither tagged more than the other: the key is
			// dropped from both renderings.
			name: "same-depth conflict annihilates the key",
			value: struct {
				PromotionLeft
				PromotionRight
			}{PromotionLeft{Clash: "l", Left: "L"}, PromotionRight{Clash: "r", Right: "R"}},
		},
		{
			// Same depth, but only one of the two carries a json tag:
			// the tagged one wins.
			name: "tagged field wins a same-depth conflict",
			value: struct {
				PromotionTaggedClash
				PromotionUntaggedClash
			}{
				PromotionTaggedClash{Clash: "tagged", OnlyT: "t"},
				PromotionUntaggedClash{Clash: "untagged", OnlyU: "u"},
			},
		},
		{
			// Nil collections: the KEYS agree. The values deliberately
			// do not, and TestNilCollectionsRenderPerTheOutputStandards
			// below pins that difference.
			name: "nil slice and nil map",
			value: struct {
				S []string          `json:"s"`
				M map[string]string `json:"m"`
				B string            `json:"b"`
			}{nil, nil, "y"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := jsonKeys(t, tc.value)
			got := yamlKeys(t, tc.value)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("key sequence drifted\n  YAML: %q\n  JSON: %q\n\nrendering:\n%s",
					got, want, renderYAML(tc.value))
			}
		})
	}
}

// TestNilCollectionsRenderPerTheOutputStandards pins the one difference
// between the two renderings that is required rather than accidental:
// [[cli-output-standards]] says an empty sequence renders `[]` and an
// empty mapping `{}`, where encoding/json writes null for a nil one.
func TestNilCollectionsRenderPerTheOutputStandards(t *testing.T) {
	value := struct {
		S []string          `json:"s"`
		M map[string]string `json:"m"`
	}{nil, nil}

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encoding/json refused the value: %v", err)
	}
	if want := `{"s":null,"m":null}`; string(encoded) != want {
		t.Fatalf("JSON = %s, want %s", encoded, want)
	}
	if got, want := renderYAML(value), "s: []\nm: {}\n"; got != want {
		t.Errorf("YAML = %q, want %q", got, want)
	}
}
