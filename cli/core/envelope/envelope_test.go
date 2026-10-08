package envelope

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Every tag the generated union decodes must be one Decode discovers,
// selecting the same variant. Discovery derives a tag from the variant's
// name, so a variant whose tag broke that convention would be
// undecodable; this pins the convention against the generated decoder.
// The generated encoder cannot serve: it refuses a variant's zero value.
func TestEveryUnionVariantIsDiscoveredUnderItsGeneratedTag(t *testing.T) {
	found := variants()
	accessors := 0
	union := reflect.TypeOf(specmodel.TypedFrontmatter{})
	for i := 0; i < union.NumMethod(); i++ {
		if strings.HasPrefix(union.Method(i).Name, "As") {
			accessors++
		}
	}
	if accessors == 0 || len(found) != accessors {
		t.Fatalf("union has %d variants, %d discovered", accessors, len(found))
	}
	for tag, v := range found {
		raw := []byte(`{"type":"` + tag + `"}`)
		// The generated decoder enforces required fields, so it refuses
		// this value — naming the variant the tag selected.
		var generated specmodel.TypedFrontmatter
		err := json.Unmarshal(raw, &generated)
		if err == nil || !strings.HasPrefix(err.Error(), v.generated.Name()+":") {
			t.Errorf("tag %q: generated decoder says %v, want a %s error", tag, err, v.generated.Name())
		}
		decoded, err := Decode(raw)
		if err != nil {
			t.Fatalf("decode %q: %v", tag, err)
		}
		if got := reflect.TypeOf(decoded.Value); got != v.generated {
			t.Errorf("decode %q produced %v, want %v", tag, got, v.generated)
		}
		if got := Of(&decoded).Type; string(got) != tag {
			t.Errorf("envelope type = %q, want %q", got, tag)
		}
	}
}

// A document declaring a type the union does not define is refused with
// the type named.
func TestDecodeFrontmatterRefusesAnUnknownType(t *testing.T) {
	_, err := Decode([]byte(`{"type":"memo"}`))
	if err == nil || !strings.Contains(err.Error(), `"memo"`) {
		t.Fatalf("err = %v, want unknown document type \"memo\"", err)
	}
}

// A document missing a required property still decodes: the omission is
// the missing-property check's finding, not a parse failure.
func TestDecodeFrontmatterToleratesAMissingRequiredProperty(t *testing.T) {
	value, err := Decode([]byte(`{"type":"adr","slug":"adr-001"}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := Of(&value); got.Type != "adr" || got.Slug != "adr-001" || got.Status != "" {
		t.Errorf("envelope = %+v", got)
	}
}

// encodeCases are frontmatters covering every kind of property a variant
// holds — required and optional enums, timestamps, pointers to slices,
// citations, an integer — each fully declared.
var encodeCases = map[string]string{
	"adr": `{"type":"adr","title":"ADR-001: <Example> & co","description":"d","slug":"adr-001","status":"approved",` +
		`"updated_at":"2026-08-01T00:00:00Z","created_at":"2026-07-01T00:00:00Z","tags":["a","b"],"superseded_by":"adr-002"}`,
	"module-blueprint": `{"type":"module-blueprint","title":"M","description":"d","slug":"module-m","status":"draft",` +
		`"updated_at":"2026-08-01T00:00:00Z","entity_name":"M","kind":"standard","visibility":"internal","version":"1.0"}`,
	"contract": `{"type":"contract","title":"C","description":"d","slug":"contract-c","status":"approved",` +
		`"updated_at":"2026-08-01T00:00:00Z","owner":"container-x","visibility":"public","version":"1.0"}`,
	"rfc": `{"type":"rfc","title":"R","description":"d","slug":"rfc-r","status":"draft","updated_at":"2026-08-01T00:00:00Z",` +
		`"author":"a","created":"2026-08-01","repository":"r","related_spec":"s","rfc_type":"t","rfc_number":7,"severity":"snag","target":"x"}`,
}

// For a frontmatter with no empty enum, Encode's bytes are the generated
// encoder's: the index of a valid corpus is unchanged by encoding through
// it.
func TestEncodeMatchesTheGeneratedEncoderForAValidFrontmatter(t *testing.T) {
	for name, raw := range encodeCases {
		t.Run(name, func(t *testing.T) {
			f, err := Decode([]byte(raw))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			want, err := json.Marshal(f)
			if err != nil {
				t.Fatalf("the generated encoder refused a valid frontmatter: %v", err)
			}
			got, err := Encode(f)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("Encode =\n%s\nwant the generated encoder's\n%s", got, want)
			}
		})
	}
}

// A required enum left at its empty value — the document never declared
// it — is omitted, where the generated encoder refuses it; every other
// property is encoded as the generated encoder would.
func TestEncodeOmitsAnUndeclaredRequiredEnum(t *testing.T) {
	cases := []struct {
		name, raw, property, want string
	}{
		{"adr without status", strings.NewReplacer(`"status":"approved",`, "", "<Example> & co", "Example").Replace(encodeCases["adr"]), "status",
			`{"type":"adr","description":"d","slug":"adr-001","title":"ADR-001: Example",` +
				`"updated_at":"2026-08-01T00:00:00Z","created_at":"2026-07-01T00:00:00Z","tags":["a","b"],"superseded_by":"adr-002"}`},
		{"module blueprint without kind", strings.Replace(encodeCases["module-blueprint"], `"kind":"standard",`, "", 1), "kind",
			`{"type":"module-blueprint","description":"d","slug":"module-m","title":"M","status":"draft",` +
				`"updated_at":"2026-08-01T00:00:00Z","entity_name":"M","visibility":"internal","version":"1.0"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := Decode([]byte(tc.raw))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if _, err := json.Marshal(f); err == nil {
				t.Fatalf("the generated encoder accepted an empty %s; this test no longer covers a refusal", tc.property)
			}
			got, err := Encode(f)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("Encode =\n%s\nwant\n%s", got, tc.want)
			}
		})
	}
}

func TestEncodeRefusesANilVariant(t *testing.T) {
	if _, err := Encode(specmodel.TypedFrontmatter{}); err == nil {
		t.Error("Encode of a nil variant succeeded, want an error")
	}
}

// A property whose value its generated type refuses — a status outside
// the vocabulary — is named by the decode error, with the value as
// written, so the failure can be located at that property.
//
// AC: AC-SPEC-003.6, AC-SPEC-006.1
func TestADecodeFailureNamesTheOffendingProperty(t *testing.T) {
	raw := []byte(`{"type":"adr","title":"T","description":"D","slug":"adr-001","status":"actve","updated_at":"2026-08-01T00:00:00Z"}`)
	_, err := Decode(raw)
	var propertyErr *PropertyError
	if !errors.As(err, &propertyErr) {
		t.Fatalf("Decode error = %v (%T), want a *PropertyError", err, err)
	}
	if propertyErr.Property != "status" || propertyErr.Value != "actve" {
		t.Errorf("PropertyError = {%q, %q}, want {status, actve}", propertyErr.Property, propertyErr.Value)
	}
}
