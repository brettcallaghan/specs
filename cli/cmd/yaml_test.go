// Tests for the YAML rendering [[cli-output-standards]]'s YAML Style
// section defines. Each case pins one rule from that document.
package cmd

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
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
