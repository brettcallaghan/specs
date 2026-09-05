// Tests for the heading tree, the heading half of Markdown Structure.
// Every fixture is body text: a heading search runs within a document
// body, and frontmatter separation is a caller concern this package does
// not take on.
package markdown_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// Every heading of a body is returned in document order, each carrying
// its level and its text as authored, and each with the span of the
// content beneath it.
//
// AC: AC-SPECSTR-001.1
func TestEveryHeadingIsReturnedInDocumentOrderWithItsSectionSpan(t *testing.T) {
	body := []byte("# Top\n\nintro\n\n## Middle\n\nbody\n\n### Deep\n\nleaf\n")

	found := markdown.Headings(body)

	if len(found) != 3 {
		t.Fatalf("found %d headings, want 3: %+v", len(found), found)
	}

	want := []struct {
		text    string
		level   int
		section string
	}{
		{"Top", 1, "\nintro\n\n## Middle\n\nbody\n\n### Deep\n\nleaf\n"},
		{"Middle", 2, "\nbody\n\n### Deep\n\nleaf\n"},
		{"Deep", 3, "\nleaf\n"},
	}
	for i, w := range want {
		got := found[i]
		if got.Text != w.text {
			t.Errorf("heading %d: Text = %q, want %q", i, got.Text, w.text)
		}
		if got.Level != w.level {
			t.Errorf("heading %d: Level = %d, want %d", i, got.Level, w.level)
		}
		if slice := string(span.Slice(body, got.Section)); slice != w.section {
			t.Errorf("heading %d: Section slices to %q, want %q", i, slice, w.section)
		}
	}

	if slice := string(span.Slice(body, found[0].Position)); slice != "# Top\n" {
		t.Errorf("Position slices to %q, want %q", slice, "# Top\n")
	}
}

// A section ends at the next heading of the same or a higher level. A
// deeper heading is content within the section rather than the end of it,
// and the last heading's section runs to the end of the document.
//
// AC: AC-SPECSTR-001.1
func TestASectionEndsAtTheNextHeadingOfTheSameOrAHigherLevel(t *testing.T) {
	body := []byte("## First\n\na\n\n### Nested\n\nb\n\n## Second\n\nc\n")

	found := markdown.Headings(body)

	if len(found) != 3 {
		t.Fatalf("found %d headings, want 3", len(found))
	}

	// The deeper heading falls inside the shallower one's section rather
	// than closing it.
	if got, want := string(span.Slice(body, found[0].Section)), "\na\n\n### Nested\n\nb\n\n"; got != want {
		t.Errorf("First's section = %q, want %q", got, want)
	}
	// A sibling at the same level closes it.
	if got, want := string(span.Slice(body, found[2].Section)), "\nc\n"; got != want {
		t.Errorf("Second's section = %q, want %q", got, want)
	}
	// The nested heading's own section ends at the next same-or-higher
	// heading, not at the end of the document.
	if got, want := string(span.Slice(body, found[1].Section)), "\nb\n\n"; got != want {
		t.Errorf("Nested's section = %q, want %q", got, want)
	}
}

// A heading that stands over nothing carries an empty span rather than an
// absent one.
//
// AC: AC-SPECSTR-001.1
func TestAHeadingWithNoContentBeneathItCarriesAnEmptySection(t *testing.T) {
	body := []byte("# Only\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1", len(found))
	}
	if found[0].Section.Length != 0 {
		t.Errorf("Section length = %d, want 0: %+v", found[0].Section.Length, found[0].Section)
	}
	if got := string(span.Slice(body, found[0].Section)); got != "" {
		t.Errorf("Section slices to %q, want empty", got)
	}
}

// What looks like a heading inside a fenced block or an inline code span
// is not a heading, because only a markdown parser knows which regions
// are not to be scanned.
//
// AC: AC-SPECSTR-001.1
func TestAHeadingInsideAFencedBlockOrACodeSpanIsNotReported(t *testing.T) {
	// The code span opens the line, so its content is the only thing on
	// that line that could be mistaken for a heading.
	body := []byte("# Real\n\n```sh\n# not a heading\n```\n\n`## not a heading` is written inline.\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if found[0].Text != "Real" {
		t.Errorf("Text = %q, want %q", found[0].Text, "Real")
	}
}

// A document with no heading yields an empty result rather than an error.
//
// AC: AC-SPECSTR-001.1
func TestADocumentWithNoHeadingYieldsAnEmptyResult(t *testing.T) {
	body := []byte("Just prose.\n\nAnd more of it.\n")

	if found := markdown.Headings(body); len(found) != 0 {
		t.Fatalf("found %d headings, want 0: %+v", len(found), found)
	}

	if found := markdown.Headings(nil); len(found) != 0 {
		t.Fatalf("an empty body found %d headings, want 0", len(found))
	}
}

// Every span the component returns is computed through Spans, so a
// heading's position carries the line and column a report reads and
// slices back to the text it covers.
//
// AC: AC-SPECSTR-001.1
func TestEverySpanIsComputedThroughSpans(t *testing.T) {
	body := []byte("intro\n\n## Heading\n\nbody\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1", len(found))
	}
	heading := found[0]

	if got, want := span.Of(body, int(heading.Position.Offset), int(heading.Position.Length)), heading.Position; got != want {
		t.Errorf("Position = %+v, want the span.Of value %+v", heading.Position, got)
	}
	if got, want := span.Of(body, int(heading.Section.Offset), int(heading.Section.Length)), heading.Section; got != want {
		t.Errorf("Section = %+v, want the span.Of value %+v", heading.Section, got)
	}
	if heading.Position.StartLine != 3 {
		t.Errorf("StartLine = %d, want 3", heading.Position.StartLine)
	}
	// The heading line ends exactly where its section begins.
	if got, want := heading.Position.Offset+heading.Position.Length, heading.Section.Offset; got != want {
		t.Errorf("Position ends at %d, Section begins at %d; want them equal", got, want)
	}
}

// A heading underlined rather than prefixed is a heading, and its
// underline belongs to the heading rather than to the section beneath it.
//
// AC: AC-SPECSTR-001.1
func TestAnUnderlinedHeadingCarriesItsUnderlineInThePositionSpan(t *testing.T) {
	body := []byte("Underlined\n==========\n\nbody\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if found[0].Level != 1 {
		t.Errorf("Level = %d, want 1", found[0].Level)
	}
	if got, want := string(span.Slice(body, found[0].Position)), "Underlined\n==========\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
	if got, want := string(span.Slice(body, found[0].Section)), "\nbody\n"; got != want {
		t.Errorf("Section slices to %q, want %q", got, want)
	}
}

// The text is recorded as authored: the markers that open and close the
// line are not part of the heading's name.
//
// AC: AC-SPECSTR-001.1
func TestTheTextIsRecordedAsAuthoredWithoutItsMarkers(t *testing.T) {
	body := []byte("## Closed as well ##\n\nbody\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1", len(found))
	}
	if got, want := found[0].Text, "Closed as well"; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}
}

// The heading identifier is the rendered inline content lowercased, with
// every character that is not a letter, digit, hyphen, underscore or
// space removed and each remaining space replaced by a hyphen.
//
// AC: AC-SPECSTR-001.1
func TestTheHeadingIdentifierIsDerivedFromTheRenderedInlineContent(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"Overview", "overview"},
		{"Key Decisions", "key-decisions"},
		{"The `Profile` interface", "the-profile-interface"},
		{"**Bold** and *italic*", "bold-and-italic"},
		{"Corpus in [[module-document-core]] (modified) — IDX", "corpus-in-module-document-core-modified--idx"},
		{"AC-SPEC-001.1: what it covers", "ac-spec-0011-what-it-covers"},
		{"under_scores kept", "under_scores-kept"},
		{"", ""},
		// A link discriminates rendering from filtering: stripping
		// punctuation out of the raw text would leave the URL behind as
		// "see-the-guidehttpexamplecomguide".
		{"See [the guide](http://example.com/guide)", "see-the-guide"},
	}
	for _, c := range cases {
		if got := markdown.Identifier(c.text); got != c.want {
			t.Errorf("Identifier(%q) = %q, want %q", c.text, got, c.want)
		}
	}
}

// The derivation is a pure function of the text and derives no positional
// suffix, so two headings that read the same yield the same identifier.
//
// AC: AC-SPECSTR-001.1
func TestTheHeadingIdentifierIsPureAndDerivesNoPositionalSuffix(t *testing.T) {
	body := []byte("## Overview\n\na\n\n## Overview\n\nb\n")

	found := markdown.Headings(body)

	if len(found) != 2 {
		t.Fatalf("found %d headings, want 2", len(found))
	}
	first, second := markdown.Identifier(found[0].Text), markdown.Identifier(found[1].Text)
	if first != second {
		t.Errorf("repeated heading identifiers differ: %q and %q", first, second)
	}
	if first != "overview" {
		t.Errorf("Identifier = %q, want %q with no positional suffix", first, "overview")
	}
	if again := markdown.Identifier("Overview"); again != first {
		t.Errorf("Identifier is not a pure function of the text: %q then %q", first, again)
	}
}

// A search that starts at a document's body leaves the frontmatter block
// behind the offset, and every span it returns is still in the document's
// own coordinates.
//
// The negative control is the whole-file search below: a frontmatter
// block's closing "---" underlines the YAML above it, which CommonMark
// reads as a setext heading, so a search over the whole file reports the
// entire block as a level-2 heading.
//
// AC: AC-SPECSTR-001.1
func TestASearchFromTheBodyOffsetLeavesTheFrontmatterBehind(t *testing.T) {
	src := []byte("---\ntitle: Example\nslug: adr-001\n---\n\n# Real\n\nbody\n")
	bodyAt := len("---\ntitle: Example\nslug: adr-001\n---\n")

	found := markdown.HeadingsFrom(src, bodyAt)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if found[0].Text != "Real" {
		t.Errorf("Text = %q, want %q", found[0].Text, "Real")
	}
	// The span addresses the document, not the slice: the heading is on
	// line 6 of the file and slices back out of src.
	if got, want := string(span.Slice(src, found[0].Position)), "# Real\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
	if found[0].Position.StartLine != 6 {
		t.Errorf("StartLine = %d, want 6 — the span must count from the start of the file", found[0].Position.StartLine)
	}
	if got, want := string(span.Slice(src, found[0].Section)), "\nbody\n"; got != want {
		t.Errorf("Section slices to %q, want %q", got, want)
	}

	// The control: without the offset, the frontmatter itself is reported
	// as a heading.
	whole := markdown.Headings(src)
	if len(whole) != 2 || whole[0].Level != 2 {
		t.Fatalf("expected the whole-file search to report the frontmatter as a level-2 heading, got %+v", whole)
	}
	if whole[0].Text != "title: Example\nslug: adr-001" {
		t.Errorf("the control did not reproduce the frontmatter-as-heading case: %q", whole[0].Text)
	}
}

// A prefixed heading inside a container block — a blockquote, a list
// item — is the prefixed form, so its Position covers its own line and no
// more. The container's own marker opens the line, and reading that
// marker as the heading's first character would make an ATX heading look
// underlined and extend its Position over the line beneath it.
//
// AC: AC-SPECSTR-001.1
func TestAPrefixedHeadingInsideAContainerBlockCoversOnlyItsOwnLine(t *testing.T) {
	body := []byte("> ## A\n> ## B\n\ntail\n")

	found := markdown.Headings(body)

	if len(found) != 2 {
		t.Fatalf("found %d headings, want 2: %+v", len(found), found)
	}
	if got, want := string(span.Slice(body, found[0].Position)), "> ## A\n"; got != want {
		t.Errorf("A's Position slices to %q, want %q", got, want)
	}
	if got, want := string(span.Slice(body, found[1].Position)), "> ## B\n"; got != want {
		t.Errorf("B's Position slices to %q, want %q", got, want)
	}
	// The pair is the reason the case matters: an overlapping pair of
	// heading positions is not a structure.
	if end, next := found[0].Position.Offset+found[0].Position.Length, found[1].Position.Offset; end > next {
		t.Errorf("A's Position ends at %d and B's begins at %d; the two overlap", end, next)
	}
}

// A prefixed heading inside a list item keeps the item's remaining
// content in its section: reading the list marker as the heading's first
// character would treat the heading as underlined, swallow the content
// line into Position and leave the section without it.
//
// AC: AC-SPECSTR-001.1
func TestAPrefixedHeadingInsideAListItemKeepsItsContentInTheSection(t *testing.T) {
	body := []byte("- # Item\n  body line\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if got, want := string(span.Slice(body, found[0].Position)), "- # Item\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
	if got, want := string(span.Slice(body, found[0].Section)), "  body line\n"; got != want {
		t.Errorf("Section slices to %q, want %q", got, want)
	}
}

// An underlined heading inside a container block is still underlined: its
// underline belongs to the heading, so Position ends exactly where
// Section begins.
//
// AC: AC-SPECSTR-001.1
func TestAnUnderlinedHeadingInsideAContainerBlockKeepsItsUnderline(t *testing.T) {
	body := []byte("> A\n> ---\n> body\n")

	found := markdown.Headings(body)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if got, want := string(span.Slice(body, found[0].Position)), "> A\n> ---\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
	if got, want := found[0].Position.Offset+found[0].Position.Length, found[0].Section.Offset; got != want {
		t.Errorf("Position ends at %d, Section begins at %d; want them equal", got, want)
	}
}

// The two written forms are told apart by where the heading's text sits
// on its line, which holds inside a container block and does not mistake
// a "#" in the heading's own text for a marker. No two headings of any
// body carry overlapping Position spans.
//
// AC: AC-SPECSTR-001.1
func TestNoTwoHeadingPositionsOverlapForAnyForm(t *testing.T) {
	bodies := []string{
		"# A\nbody\n",                 // plain prefixed
		"   ## Indented\nbody\n",      // prefixed behind indentation
		"A\n---\nbody\n",              // plain underlined
		"a # b\n===\nbody\n",          // underlined, "#" in its own text
		"## # Hashy\nbody\n",          // prefixed, "#" in its own text
		"> ## A\n> ## B\n\ntail\n",    // prefixed inside a blockquote
		"- # Item\n  body line\n",     // prefixed inside a list item
		"> A\n> ---\n> body\n",        // underlined inside a blockquote
		"# A\n\n## B\n\n### C\n\nx\n", // a nested run
		"# One\nx\n\n# Two\ny\n",      // siblings
	}
	for _, source := range bodies {
		body := []byte(source)
		found := markdown.Headings(body)
		for i := 1; i < len(found); i++ {
			previous, current := found[i-1], found[i]
			if end := previous.Position.Offset + previous.Position.Length; end > current.Position.Offset {
				t.Errorf("%q: heading %d's Position ends at %d and heading %d's begins at %d; the two overlap",
					source, i-1, end, i, current.Position.Offset)
			}
		}
		// Every heading's own line ends where its section begins, in
		// both written forms.
		for i, heading := range found {
			if end := heading.Position.Offset + heading.Position.Length; end != heading.Section.Offset {
				t.Errorf("%q: heading %d's Position ends at %d but its Section begins at %d",
					source, i, end, heading.Section.Offset)
			}
		}
	}
}

// The forms that already worked are unchanged: a prefixed heading whose
// own text begins with "#" is prefixed, and an underlined heading whose
// text contains one is underlined.
//
// AC: AC-SPECSTR-001.1
func TestAHashInTheHeadingsOwnTextDoesNotChangeItsForm(t *testing.T) {
	prefixed := []byte("## # Hashy\nbody\n")
	found := markdown.Headings(prefixed)
	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if got, want := string(span.Slice(prefixed, found[0].Position)), "## # Hashy\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
	if got, want := found[0].Text, "# Hashy"; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}

	underlined := []byte("a # b\n===\nbody\n")
	found = markdown.Headings(underlined)
	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if got, want := string(span.Slice(underlined, found[0].Position)), "a # b\n===\n"; got != want {
		t.Errorf("Position slices to %q, want %q", got, want)
	}
}

// A heading's text is rendered as inline content, so a block-level
// marker that opens it — an ordinal, a bullet, a blockquote marker — is a
// literal character of the heading's name rather than syntax. Two
// numbered siblings that read the same therefore address differently.
//
// AC: AC-SPECSTR-003.2
func TestALeadingBlockMarkerIsPartOfTheHeadingsName(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"1. Overview", "1-overview"},
		{"2. Overview", "2-overview"},
		// The derivation removes a disallowed character and maps each
		// remaining space to a hyphen, and does neither collapse a run
		// of hyphens nor trim one: dropped punctuation beside a space
		// leaves both hyphens, here and in the "modified — IDX" case
		// above.
		{"- Bullet", "--bullet"},
		{"3. `## Design` present", "3--design-present"},
		{"## Not a heading here", "-not-a-heading-here"},
		{"> quoted", "-quoted"},
		{"1. Self-contained Problem statement (the fresh-agent test)",
			"1-self-contained-problem-statement-the-fresh-agent-test"},
	}
	for _, c := range cases {
		if got := markdown.Identifier(c.text); got != c.want {
			t.Errorf("Identifier(%q) = %q, want %q", c.text, got, c.want)
		}
	}

	// The collision is the point: an ordinal is neither letter case nor
	// punctuation, so two numbered siblings must not share an anchor.
	if first, second := markdown.Identifier("1. Overview"), markdown.Identifier("2. Overview"); first == second {
		t.Errorf("numbered siblings collide: both %q", first)
	}
	// An identifier derived from a heading whose text begins with a digit
	// does not begin with a hyphen.
	for _, text := range []string{"3. `## Design` present", "4. `## Acceptance` criteria (`kind: bug`)"} {
		if got := markdown.Identifier(text); strings.HasPrefix(got, "-") {
			t.Errorf("Identifier(%q) = %q, which begins with a hyphen", text, got)
		}
	}
}

// Rendering inline content is unchanged by that: markup is still removed
// rather than retained, and a link's target still does not leak into the
// identifier.
//
// AC: AC-SPECSTR-003.2
func TestInlineRenderingIsUnchangedByTheBlockMarkerRule(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"The `Profile` interface", "the-profile-interface"},
		{"**Bold** and *italic*", "bold-and-italic"},
		{"*emph* and _under_", "emph-and-under"},
		{"See [the guide](http://example.com/guide)", "see-the-guide"},
		{"A [link](http://x) here", "a-link-here"},
	}
	for _, c := range cases {
		if got := markdown.Identifier(c.text); got != c.want {
			t.Errorf("Identifier(%q) = %q, want %q", c.text, got, c.want)
		}
	}
}
