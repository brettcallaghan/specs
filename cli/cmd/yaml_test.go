// Tests for the YAML rendering [[cli-output-standards]]'s YAML Style
// section defines. Each case pins one rule from that document.
package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// roundTrips parses rendered YAML back, so a case proves the bytes say
// what they were meant to say rather than merely looking right.
func roundTrips(t *testing.T, rendered string) map[string]any {
	t.Helper()
	var back map[string]any
	if err := yaml.Unmarshal([]byte(rendered), &back); err != nil {
		t.Fatalf("the rendering does not parse as YAML: %v\n%s", err, rendered)
	}
	return back
}

func TestBlockStyleWithTwoSpaceIndentAndSequencesUnderTheirKey(t *testing.T) {
	type match struct {
		Path  string `json:"path"`
		Level uint64 `json:"level"`
	}
	value := struct {
		Matches []match `json:"matches"`
	}{Matches: []match{{Path: "a.md", Level: 2}, {Path: "b.md", Level: 3}}}

	got := renderYAML(value)
	want := "matches:\n" +
		"  - path: a.md\n" +
		"    level: 2\n" +
		"  - path: b.md\n" +
		"    level: 3\n"
	if got != want {
		t.Errorf("rendering =\n%q\nwant\n%q", got, want)
	}
	roundTrips(t, got)
}

func TestAnEmptySequenceRendersAsBracketsAndAnEmptyMappingAsBraces(t *testing.T) {
	value := struct {
		Matches []string          `json:"matches"`
		Extra   map[string]string `json:"extra"`
	}{Matches: []string{}, Extra: map[string]string{}}

	got := renderYAML(value)
	want := "matches: []\nextra: {}\n"
	if got != want {
		t.Errorf("rendering = %q, want %q", got, want)
	}

	back := roundTrips(t, got)
	// The point of the rule: empty is distinguishable from absent.
	if back["matches"] == nil {
		t.Error("an empty sequence parsed back as null, which the rule exists to prevent")
	}
}

func TestAMultiLineStringIsTrimmedAndRenderedAsABlockScalar(t *testing.T) {
	value := struct {
		Content string `json:"content"`
	}{Content: "\n\nfirst line\n\n  indented line\nlast line\n\n"}

	got := renderYAML(value)
	want := "content: |-\n" +
		"  first line\n" +
		"\n" +
		"    indented line\n" +
		"  last line\n"
	if got != want {
		t.Errorf("rendering =\n%q\nwant\n%q", got, want)
	}

	back := roundTrips(t, got)
	// Whitespace inside the value is content and is left alone.
	if back["content"] != "first line\n\n  indented line\nlast line" {
		t.Errorf("content parsed back as %q", back["content"])
	}
}

func TestAMultiLineStringCarryingAControlCharacterIsDoubleQuoted(t *testing.T) {
	value := struct {
		Content string `json:"content"`
	}{Content: "first\r\nsecond"}

	got := renderYAML(value)
	if strings.Contains(got, "|-") {
		t.Errorf("a carriage return was rendered as a block scalar: %q", got)
	}

	back := roundTrips(t, got)
	if back["content"] != "first\r\nsecond" {
		t.Errorf("content parsed back as %q, want the carriage return intact", back["content"])
	}
}

func TestAStringIsPlainWhereItRoundTripsAndQuotedWhereItDoesNot(t *testing.T) {
	plain := []string{
		"Overview",
		"the-profile-interface",
		"specs/adrs/adr-001.md",
		"A sentence with punctuation, and more.",
		"Ratio 1:2 written without a space",
	}
	quoted := []string{
		"",
		" leading",
		"trailing ",
		"- dash first",
		"#hash first",
		"key: value",
		"ADR-001: Example",
		"trailing colon:",
		"a # comment",
		"123",
		"1.5",
		"true",
		"NULL",
		"yes",
	}

	for _, s := range plain {
		got := renderYAML(struct {
			V string `json:"v"`
		}{V: s})
		if strings.Contains(got, `"`) {
			t.Errorf("%q was quoted: %s", s, got)
		}
		if back := roundTrips(t, got); back["v"] != s {
			t.Errorf("%q parsed back as %#v", s, back["v"])
		}
	}

	for _, s := range quoted {
		got := renderYAML(struct {
			V string `json:"v"`
		}{V: s})
		if !strings.Contains(got, `"`) {
			t.Errorf("%q was left plain: %s", s, got)
		}
		if back := roundTrips(t, got); back["v"] != s {
			t.Errorf("%q parsed back as %#v", s, back["v"])
		}
	}
}

func TestAdrOneColonExampleParsesBackAsWritten(t *testing.T) {
	// "ADR-001: Example" contains ": ", so it must be quoted — the case
	// the heading query hits on this repository's own corpus.
	got := renderYAML(struct {
		Text string `json:"text"`
	}{Text: "ADR-001: Example"})

	if back := roundTrips(t, got); back["text"] != "ADR-001: Example" {
		t.Errorf("text parsed back as %#v from %q", back["text"], got)
	}
}

func TestValuesAreRenderedAsTheirOwnTypes(t *testing.T) {
	value := struct {
		Count   uint64  `json:"count"`
		Ok      bool    `json:"ok"`
		Missing *string `json:"missing"`
	}{Count: 7, Ok: true}

	got := renderYAML(value)
	want := "count: 7\nok: true\nmissing: null\n"
	if got != want {
		t.Errorf("rendering = %q, want %q", got, want)
	}
}

func TestNestedMappingsIndentOneLevelUnderTheirKey(t *testing.T) {
	type span struct {
		Offset uint64 `json:"offset"`
		Length uint64 `json:"length"`
	}
	value := struct {
		Section span `json:"section"`
	}{Section: span{Offset: 10, Length: 20}}

	got := renderYAML(value)
	want := "section:\n  offset: 10\n  length: 20\n"
	if got != want {
		t.Errorf("rendering = %q, want %q", got, want)
	}
	roundTrips(t, got)
}

func TestTheOutputCarriesNoDocumentMarkersOrComments(t *testing.T) {
	value := struct {
		Matches []string `json:"matches"`
	}{Matches: []string{"one", "two"}}

	got := renderYAML(value)
	for _, forbidden := range []string{"---", "...", "#", "&", "*", "!!"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("rendering contains %q: %s", forbidden, got)
		}
	}
}

// A map is rendered in key order, so that one value renders as the same
// bytes on every run despite Go randomizing map iteration.
func TestAMapRendersInKeyOrder(t *testing.T) {
	value := struct {
		Content map[string]string `json:"content"`
	}{Content: map[string]string{"slug": "c", "title": "a", "description": "b"}}

	want := "content:\n  description: b\n  slug: c\n  title: a\n"
	for range 8 {
		if got := renderYAML(value); got != want {
			t.Fatalf("rendering = %q, want %q", got, want)
		}
	}
}

// A value that marshals itself renders from its own JSON rather than
// from its Go fields, so the two renderings carry the same keys in the
// same order. A union is the case that forces it: its Go shape is one
// field holding a variant, and its JSON is the variant's keys under a
// tag naming which variant it is.
func TestAValueThatMarshalsItselfRendersFromItsOwnJSON(t *testing.T) {
	value := struct {
		Reference specmodel.Reference `json:"reference"`
	}{Reference: specmodel.Reference{
		Path:   "docs/adrs/adr-800-current.md",
		Text:   "adr-900-old-approach",
		Slug:   "adr-900-old-approach",
		Origin: specmodel.NewReferenceOriginFrontmatterProperty(specmodel.ReferenceFrontmatterOrigin{Property: "superseded_by"}),
		Span:   specmodel.Span{Offset: 1, Length: 2, StartLine: 3, StartColumn: 4, EndLine: 5, EndColumn: 6},
		Resolution: specmodel.NewReferenceResolutionResolved(
			specmodel.ReferenceResolved{TargetPath: "docs/adrs/adr-900-old-approach.md"}),
	}}

	got := renderYAML(value)
	for _, want := range []string{
		"  origin:\n    kind: frontmatter-property\n    property: superseded_by\n",
		"  resolution:\n    state: resolved\n    target_path: docs/adrs/adr-900-old-approach.md\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendering does not contain %q:\n%s", want, got)
		}
	}
	// The Go field the union holds its variant in is not a key of the
	// answer, and its absence is the whole point of the rule.
	if strings.Contains(got, "Value:") {
		t.Errorf("rendering exposes the union's Go field:\n%s", got)
	}
	roundTrips(t, got)
}

// The rendering and the JSON one carry the same keys in the same order,
// which is what [[cli-output-standards]] requires of the two forms —
// checked over a value carrying two unions rather than asserted of the
// renderer in the abstract.
func TestTheTwoRenderingsCarryTheSameKeysInTheSameOrder(t *testing.T) {
	anchor := "Decision"
	value := struct {
		Reference specmodel.Reference `json:"reference"`
	}{Reference: specmodel.Reference{
		Path:   "docs/adrs/adr-800-current.md",
		Text:   "adr-900-old-approach#Decision",
		Slug:   "adr-900-old-approach",
		Anchor: &anchor,
		Origin: specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{}),
		Span:   specmodel.Span{Offset: 1, Length: 2, StartLine: 3, StartColumn: 4, EndLine: 5, EndColumn: 6},
		Resolution: specmodel.NewReferenceResolutionResolved(
			specmodel.ReferenceResolved{TargetPath: "docs/adrs/adr-900-old-approach.md"}),
		AnchorResolution: ambiguousAnchorResolution(),
	}}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshalling the value as JSON: %v", err)
	}
	if diff := keyOrderDiff(t, data, renderYAML(value)); diff != "" {
		t.Errorf("the two renderings disagree: %s", diff)
	}
}

// ambiguousAnchorResolution is an anchor that matched two headings — a
// union whose payload carries a sequence of mappings, so the comparison
// above covers nesting under a tag as well as beside one.
func ambiguousAnchorResolution() *specmodel.AnchorResolution {
	resolution := specmodel.NewAnchorResolutionAmbiguous(specmodel.AnchorAmbiguous{
		Matches: []specmodel.AnchorMatch{
			{Text: "Decision", Span: specmodel.Span{Offset: 7, Length: 8}},
			{Text: "Decision", Span: specmodel.Span{Offset: 9, Length: 10}},
		},
	})
	return &resolution
}

// keyOrderDiff reports how the keys of a JSON document and of a YAML one
// differ, comparing the flattened key path and leaf value of each. It
// returns the empty string where they agree.
func keyOrderDiff(t *testing.T, data []byte, rendered string) string {
	t.Helper()

	var fromJSON any
	if err := json.Unmarshal(data, &fromJSON); err != nil {
		t.Fatalf("parsing the JSON rendering: %v", err)
	}
	var fromYAML any
	if err := yaml.Unmarshal([]byte(rendered), &fromYAML); err != nil {
		t.Fatalf("parsing the YAML rendering: %v\n%s", err, rendered)
	}

	jsonKeys := strings.Join(sortedKeyPaths(fromJSON, ""), "\n")
	yamlKeys := strings.Join(sortedKeyPaths(fromYAML, ""), "\n")
	if jsonKeys == yamlKeys {
		return ""
	}
	return "\n--- json ---\n" + jsonKeys + "\n--- yaml ---\n" + yamlKeys
}

// sortedKeyPaths flattens a parsed document into its key paths. Both
// parsers hand back unordered maps, so the paths are sorted and the
// comparison is of the key SETS; the order within one rendering is
// pinned by the byte-exact cases above.
func sortedKeyPaths(value any, prefix string) []string {
	var paths []string
	switch v := value.(type) {
	case map[string]any:
		for key, nested := range v {
			paths = append(paths, sortedKeyPaths(nested, prefix+"."+key)...)
		}
	case map[any]any:
		for key, nested := range v {
			paths = append(paths, sortedKeyPaths(nested, fmt.Sprintf("%s.%v", prefix, key))...)
		}
	case []any:
		for i, nested := range v {
			paths = append(paths, sortedKeyPaths(nested, fmt.Sprintf("%s[%d]", prefix, i))...)
		}
	default:
		paths = append(paths, fmt.Sprintf("%s = %v", prefix, value))
	}
	sort.Strings(paths)
	return paths
}

// A value whose own marshalling fails is rendered from its Go fields
// rather than lost or fatal. A union holding no variant is the case that
// reaches it: its MarshalJSON returns an error, and the rendering still
// has to answer, because a YAML rendering that vanished would report the
// same defect --json already reports and answer nothing.
func TestAValueWhoseOwnMarshallingFailsFallsBackToItsFields(t *testing.T) {
	value := struct {
		Resolution specmodel.ReferenceResolution `json:"resolution"`
	}{}

	if _, err := json.Marshal(value); err == nil {
		t.Fatal("the fixture's JSON marshalling succeeded; it must fail for this case to mean anything")
	}

	got := renderYAML(value)
	if !strings.Contains(got, "resolution:") {
		t.Errorf("rendering dropped the value: %q", got)
	}
}
