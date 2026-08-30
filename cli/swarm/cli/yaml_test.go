// Tests for the YAML the surface emits.
//
// The output standard's conformance table is a list of properties, and a
// table is not a check. These are the checks: every style rule below is
// asserted on the bytes AND, where a value is at stake, on what those
// bytes parse back to — so a rendering that looks right and means
// something else fails here.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func rendered(t *testing.T, n yamlNode) string {
	t.Helper()

	var out bytes.Buffer
	if err := writeYAML(&out, n); err != nil {
		t.Fatalf("writing the document: %v", err)
	}
	return out.String()
}

// A multi-line string is a block scalar, with the chomping indicator that
// reproduces the value exactly. Plain `|` clips to one trailing newline,
// so using it on a value that has none adds a byte.
func TestAMultiLineValueIsABlockScalarThatRoundTrips(t *testing.T) {
	for _, tc := range []struct{ name, value, want string }{
		{"no trailing newline", "a\nb", "v: |-\n  a\n  b\n"},
		{"exactly one", "a\nb\n", "v: |\n  a\n  b\n"},
		{"two", "a\nb\n\n", "v: |+\n  a\n  b\n\n"},
		{"three", "a\n\n\n", "v: |+\n  a\n\n\n"},
		{"a blank line inside", "a\n\nb", "v: |-\n  a\n\n  b\n"},
		{"a first line that begins with a space", " a\nb", "v: |2-\n   a\n  b\n"},
		{"a line with trailing whitespace", "a \nb", "v: |-\n  a \n  b\n"},
		{"a line of only whitespace", "a\n   \nb", "v: |-\n  a\n     \n  b\n"},
		{"a tab inside", "a\n\tb", "v: |-\n  a\n  \tb\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := rendered(t, yMap{{"v", yStr(tc.value)}})
			if got != tc.want {
				t.Errorf("rendered %q, want %q", got, tc.want)
			}

			var back map[string]string
			if err := yaml.Unmarshal([]byte(got), &back); err != nil {
				t.Fatalf("the rendering does not parse: %v\n%s", err, got)
			}
			if back["v"] != tc.value {
				t.Errorf("the value read back as %q, want %q", back["v"], tc.value)
			}
		})
	}
}

// A block scalar carries the line break and nothing else, so a value
// with any other control character in it — or with white space at the end
// of its last content line, which readers disagree about — is
// double-quoted instead. The rendering a reader would hand back changed
// is not one this renderer writes.
func TestAValueABlockScalarCannotCarryIsQuoted(t *testing.T) {
	for _, value := range []string{"a\r\nb", "a\nb\r", "a\n\x00b", "a\nb ", "a\nb\t"} {
		got := rendered(t, yMap{{"v", yStr(value)}})
		if strings.Contains(got, "|") {
			t.Errorf("%q was rendered as a block scalar: %q", value, got)
		}

		var back map[string]string
		if err := yaml.Unmarshal([]byte(got), &back); err != nil {
			t.Fatalf("the rendering does not parse: %v\n%s", err, got)
		}
		if back["v"] != value {
			t.Errorf("the value read back as %q, want %q", back["v"], value)
		}
	}
}

// An empty sequence renders `[]` and an empty mapping `{}`. There is no
// block rendering of an empty collection — it would be a key with nothing
// after it, which is a key whose value is null — so this is the single
// forced exception to block style, and it is what makes an empty list
// distinguishable from an empty string and from an unset key.
func TestAnEmptyCollectionIsDistinguishableFromNothing(t *testing.T) {
	got := rendered(t, yMap{
		{"empty_list", ySeq{}},
		{"empty_map", yMap{}},
		{"empty_string", yStr("")},
		{"unset", yNull()},
	})
	want := "empty_list: []\nempty_map: {}\nempty_string: \"\"\nunset: null\n"
	if got != want {
		t.Fatalf("rendered %q, want %q", got, want)
	}

	var back map[string]any
	if err := yaml.Unmarshal([]byte(got), &back); err != nil {
		t.Fatalf("the rendering does not parse: %v\n%s", err, got)
	}
	if list, ok := back["empty_list"].([]any); !ok || len(list) != 0 {
		t.Errorf("empty_list read back as %#v, want an empty sequence", back["empty_list"])
	}
	if back["empty_string"] != "" {
		t.Errorf("empty_string read back as %#v, want the empty string", back["empty_string"])
	}
	if back["unset"] != nil {
		t.Errorf("unset read back as %#v, want null", back["unset"])
	}
}

// A string is plain where it round-trips plain and double-quoted
// otherwise. Quoting everything is safe and unreadable; quoting nothing
// is readable and wrong, so a reader can tell that a quoted value is
// quoted because it had to be.
func TestAStringIsQuotedExactlyWhereItHasToBe(t *testing.T) {
	for _, tc := range []struct {
		name   string
		value  string
		quoted bool
	}{
		{"a word", "branch", false},
		{"a sentence", "2 key(s) written", false},
		{"a path", "docs/retros/fixture.md", false},
		{"an inner quote", `say "hi"`, false},
		{"an inner colon with no space", "a:b", false},
		{"empty", "", true},
		{"leading whitespace", " a", true},
		{"trailing whitespace", "a ", true},
		{"a leading indicator", "--not-a-flag", true},
		{"a lone dash", "-", true},
		{"a colon and a space", "write it with: swarm manifest set", true},
		{"a space and a hash", "a #b", true},
		{"a trailing colon", "a:", true},
		{"a tab", "a\tb", true},
		{"reads as an integer", "12", true},
		{"reads as a float", "1.0", true},
		{"reads as a boolean", "no", true},
		{"reads as null", "~", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := rendered(t, yMap{{"v", yStr(tc.value)}})
			body := strings.TrimSuffix(strings.TrimPrefix(got, "v: "), "\n")
			if quoted := strings.HasPrefix(body, `"`); quoted != tc.quoted {
				t.Errorf("rendered %q; quoted = %v, want %v", got, quoted, tc.quoted)
			}

			var back map[string]string
			if err := yaml.Unmarshal([]byte(got), &back); err != nil {
				t.Fatalf("the rendering does not parse: %v\n%s", err, got)
			}
			if back["v"] != tc.value {
				t.Errorf("the value read back as %q, want %q", back["v"], tc.value)
			}
		})
	}
}

// A value is rendered as its own type, so a consumer parsing the YAML
// gets the type the fact carries and a human reading it sees the same
// thing.
func TestAValueIsRenderedAsItsOwnType(t *testing.T) {
	got := rendered(t, yMap{
		{"count", yInt(20)},
		{"flag", yBool(false)},
		{"absent", yNull()},
		{"text", yStr("unrecorded")},
	})
	want := "count: 20\nflag: false\nabsent: null\ntext: unrecorded\n"
	if got != want {
		t.Errorf("rendered %q, want %q", got, want)
	}
}

// Block style throughout, two spaces per level, a sequence indented one
// level under the key that introduces it — including a sequence of
// sequences, which is what the full result's invocations are.
func TestTheOutputIsBlockStyleThroughout(t *testing.T) {
	got := rendered(t, yMap{
		{"invocations", ySeq{ySeq{yStr("bd"), yStr("show"), yStr("fixture-run")}}},
		{"fact", yMap{{"name", yStr("branch")}, {"established", yBool(true)}}},
	})
	want := "invocations:\n  - - bd\n    - show\n    - fixture-run\n" +
		"fact:\n  name: branch\n  established: true\n"
	if got != want {
		t.Errorf("rendered %q, want %q", got, want)
	}
}

// Nothing is wrapped, folded, truncated or elided: agents read long
// lines, and a wrapping decision depends on a terminal width that is not
// a property of the data.
func TestALongValueIsRenderedWhole(t *testing.T) {
	long := strings.Repeat("word ", 200) + "end"

	got := rendered(t, yMap{{"observability", yStr(long)}})
	if lines := strings.Count(got, "\n"); lines != 1 {
		t.Fatalf("a long value was rendered over %d lines:\n%s", lines, got)
	}
	if !strings.HasSuffix(got, "end\n") {
		t.Errorf("the value does not end where it should: %q", got[len(got)-40:])
	}
	if strings.Contains(got, ">") {
		t.Errorf("the value was folded, which changes its bytes:\n%s", got)
	}
}

// No comments, anchors, aliases, explicit tags, document-start marker or
// document-end marker: each is a feature a reader would have to resolve
// before they know what the document says. Asserted over a document that
// repeats a value, which is what would otherwise become an alias.
func TestTheDocumentCarriesNothingToResolve(t *testing.T) {
	repeated := ySeq{yStr("the same line"), yStr("the same line")}

	got := rendered(t, yMap{{"a", repeated}, {"b", repeated}})
	for _, forbidden := range []string{"---", "...", "&", "*", "!!", " #"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("the document carries %q:\n%s", forbidden, got)
		}
	}
}

// A key is a scalar like any other, with one addition: a colon anywhere
// in it, not only before a space, is ambiguous in key position. Validate
// names its rules by joining the keys they cover, so a key with a comma,
// parentheses or a colon in it is not hypothetical.
func TestAKeyIsQuotedWhereItHasToBe(t *testing.T) {
	got := rendered(t, yMap{
		{"checks_ids (manifest/absent)", yStr("x")},
		{"a, b", yStr("x")},
		{"a:b", yStr("x")},
	})

	var back map[string]string
	if err := yaml.Unmarshal([]byte(got), &back); err != nil {
		t.Fatalf("the rendering does not parse: %v\n%s", err, got)
	}
	for _, key := range []string{"checks_ids (manifest/absent)", "a, b", "a:b"} {
		if back[key] != "x" {
			t.Errorf("the key %q did not round-trip:\n%s", key, got)
		}
	}
}

// The two renderings come from one node tree, which is what keeps them
// carrying the same values in the same order rather than being two
// renderers kept in agreement.
func TestTheJSONRenderingKeepsTheOrderTheCommandReported(t *testing.T) {
	node := yMap{
		{"stall_after", yStr("30m0s")},
		{"worktrees_standing", yInt(1)},
		{"conflicts", ySeq{}},
	}

	var out bytes.Buffer
	if err := writeJSON(&out, node); err != nil {
		t.Fatalf("writing the JSON: %v", err)
	}
	want := "{\n  \"stall_after\": \"30m0s\",\n  \"worktrees_standing\": 1,\n  \"conflicts\": []\n}\n"
	if out.String() != want {
		t.Errorf("rendered %q, want %q", out.String(), want)
	}
}
