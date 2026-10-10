// Tests for reference extraction.
//
// Every fixture here is a whole document — frontmatter block and body —
// because half the rules under test are about the boundary between the
// two, and a fixture that starts at the body cannot exercise them.
package references_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/frontmatter"
	"github.com/brettcallaghan/specs/cli/core/references"
	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// extract parses src's frontmatter and extracts its references, which is
// what every test here does first.
func extract(t *testing.T, src string, properties ...string) ([]specmodel.Reference, []byte) {
	t.Helper()

	source := []byte(src)
	front, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parsing the fixture's frontmatter: %s", perr.Message)
	}
	return references.Extract(references.Options{
		Path:                "specs/fixture.md",
		Source:              source,
		Frontmatter:         front,
		ReferenceProperties: properties,
	}), source
}

// slugs returns the slug half of each reference, which is what most of
// these tests assert on.
func slugs(found []specmodel.Reference) []string {
	out := make([]string, 0, len(found))
	for _, reference := range found {
		out = append(out, reference.Slug)
	}
	return out
}

// AC: AC-SPECREF-001.1
func TestEveryBodyReferenceIsRecordedWithItsSpan(t *testing.T) {
	found, source := extract(t, `---
title: A document
---

The decision is [[adr-012]], refining [[adr-009]].

See also [[module-document-core#Components]].
`)

	if want := []string{"adr-012", "adr-009", "module-document-core"}; !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}

	for _, reference := range found {
		if reference.Path != "specs/fixture.md" {
			t.Errorf("%q carries path %q", reference.Text, reference.Path)
		}
		if got := string(span.Slice(source, reference.Span)); got != reference.Text {
			t.Errorf("the span of %q covers %q", reference.Text, got)
		}
		if reference.Span.StartLine == 0 || reference.Span.StartColumn == 0 {
			t.Errorf("%q carries no line or column: %+v", reference.Text, reference.Span)
		}
		if _, ok := reference.Origin.AsBody(); !ok {
			t.Errorf("%q is not recorded as a body reference", reference.Text)
		}
	}

	// The position is the whole reference, at the line and column a
	// reader opens, rather than a bare offset.
	if first := found[0].Span; first.StartLine != 5 {
		t.Errorf("the first reference is on line %d, want 5", first.StartLine)
	}
}

// AC: AC-SPECREF-001.2
func TestAReferenceInsideCodeIsNotRecorded(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"Write `[[security-standards]]` to cite a standard.\n\n"+
		"```markdown\nSee [[adr-004]] and [[adr-006]].\n```\n\n"+
		"    [[adr-indented]]\n\n"+
		"A real one: [[adr-012]].\n\n"+
		"``a code span with a [[double-backtick]] fence``\n\n"+
		"``` \nfenced with a trailing space [[adr-007]]\n```\n")

	if want := []string{"adr-012"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

// A fence's info string is code too, so a reference written there is an
// illustration on the same terms as one written inside the fence.
func TestAReferenceInAFenceInfoStringIsNotRecorded(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"```[[adr-004]]\ncontent\n```\n\nReal: [[adr-012]].\n")

	if want := []string{"adr-012"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

// A fenced block tagged `gherkin` is Gherkin, not an illustration, so its
// content is not masked: a wikilink written in a feature description, a
// rule name, a scenario name, a step or an examples cell is extracted
// like any other body reference.
//
// AC: AC-GHK-004.1
func TestAWikilinkInsideAGherkinFenceIsExtracted(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"```gherkin\n"+
		"Feature: Refunds cite [[gk-description]]\n"+
		"\n"+
		"  Rule: [[gk-rule]] governs partial refunds\n"+
		"\n"+
		"    Scenario: A wikilink in the name [[gk-scenario]]\n"+
		"      Given a step naming [[gk-step]]\n"+
		"      When nothing else happens\n"+
		"      Then done\n"+
		"\n"+
		"    Scenario Outline: parameterized citing <method>\n"+
		"      Given a method\n"+
		"      Examples:\n"+
		"        | method          |\n"+
		"        | [[gk-examples]] |\n"+
		"```\n")

	want := []string{"gk-description", "gk-rule", "gk-scenario", "gk-step", "gk-examples"}
	if !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}
	for _, reference := range found {
		if _, ok := reference.Origin.AsBody(); !ok {
			t.Errorf("%q is not recorded as a body reference", reference.Text)
		}
	}
}

// A renderer shows a gherkin fence literally, so the filter a rendering
// caller applies drops a reference written inside one and keeps the same
// slug written in prose, plus every frontmatter reference, in order —
// while Extract itself still returns all of them for the index.
//
// AC: AC-SPECMD-001.4, AC-GHK-004.1
func TestOutsideGherkinFencesDropsOnlyFencedBodyReferences(t *testing.T) {
	src := "---\ntitle: A document\nsupersedes: front-ref\n---\n\n" +
		"Prose cites [[shared]] first.\n\n" +
		"```gherkin\n" +
		"Feature: Refunds follow [[shared]] and [[fenced-only]]\n" +
		"```\n\n" +
		"```text\n[[illustration]]\n```\n\n" +
		"Prose cites [[after]] last.\n"
	source := []byte(src)
	front, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parsing the fixture's frontmatter: %s", perr.Message)
	}
	options := references.Options{
		Path:                "specs/fixture.md",
		Source:              source,
		Frontmatter:         front,
		ReferenceProperties: []string{"supersedes"},
	}
	all := references.Extract(options)
	if want := []string{"front-ref", "shared", "shared", "fenced-only", "after"}; !equal(slugs(all), want) {
		t.Fatalf("Extract slugs = %v, want %v", slugs(all), want)
	}

	kept := references.OutsideGherkinFences(options, all)
	if want := []string{"front-ref", "shared", "after"}; !equal(slugs(kept), want) {
		t.Fatalf("OutsideGherkinFences slugs = %v, want %v", slugs(kept), want)
	}
	if kept[1].Span != all[1].Span {
		t.Errorf("the prose reference kept is not the first one written: span %+v, want %+v", kept[1].Span, all[1].Span)
	}
}

// A document with no gherkin fence comes back unchanged.
//
// AC: AC-SPECMD-001.4
func TestOutsideGherkinFencesKeepsADocumentWithoutAFence(t *testing.T) {
	src := "---\ntitle: A document\n---\n\nProse cites [[one]] and [[two]].\n"
	source := []byte(src)
	front, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parsing the fixture's frontmatter: %s", perr.Message)
	}
	options := references.Options{Path: "specs/fixture.md", Source: source, Frontmatter: front}
	kept := references.OutsideGherkinFences(options, references.Extract(options))
	if want := []string{"one", "two"}; !equal(slugs(kept), want) {
		t.Fatalf("slugs = %v, want %v", slugs(kept), want)
	}
}

// A reference extracted from inside a fenced gherkin block still carries
// the span of the wikilink in the whole document, exactly as one written
// in prose does — the fence changes what is masked, not how a position
// is computed.
//
// AC: AC-GHK-004.2
func TestAGherkinFenceReferenceIsPositionedInTheContainingDocument(t *testing.T) {
	found, source := extract(t, "---\ntitle: A document\n---\n\n"+
		"Prose before the block.\n\n"+
		"```gherkin\n"+
		"Feature: Refunds\n"+
		"\n"+
		"  Scenario: A step names a reference\n"+
		"    Given a step citing [[gk-positioned]]\n"+
		"    When nothing else happens\n"+
		"    Then done\n"+
		"```\n")

	if len(found) != 1 {
		t.Fatalf("found %d references, want 1: %v", len(found), slugs(found))
	}
	reference := found[0]
	if got := string(span.Slice(source, reference.Span)); got != "gk-positioned" {
		t.Errorf("the span covers %q, want the reference text", got)
	}
	if reference.Span.StartLine == 0 || reference.Span.StartColumn == 0 {
		t.Errorf("%q carries no line or column: %+v", reference.Text, reference.Span)
	}
}

// A reference written inside Gherkin resolves through the same lookup a
// prose reference does: this package extracts it the same way, and
// [references.Resolve] never asks where a reference was written.
//
// AC: AC-GHK-004.3
func TestAReferenceInsideGherkinResolvesLikeAProseReference(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"```gherkin\n"+
		"Feature: Refunds\n"+
		"\n"+
		"  Scenario: A step names references\n"+
		"    Given a step citing [[adr-012]] and [[adr-999-absent]]\n"+
		"    When nothing else happens\n"+
		"    Then done\n"+
		"```\n")

	if want := []string{"adr-012", "adr-999-absent"}; !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}

	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})
	references.Resolve(found, index)

	if got := targetOf(t, found[0].Resolution); got != "specs/adrs/adr-012.md" {
		t.Errorf("the reference resolved to %q, want the member it names", got)
	}
	if _, ok := found[1].Resolution.AsUnresolved(); !ok {
		t.Errorf("a reference naming no corpus member resolved as %+v, want unresolved — "+
			"the report a prose reference to the same absent slug would receive", found[1].Resolution.Value)
	}
}

// An anchor written on a wikilink inside Gherkin is split from the slug
// on the first `#`, as it is everywhere else, and recorded without being
// interpreted: nothing resolves it against the Feature, Rule or Scenario
// name it happens to read like, because a scenario name is not a
// heading.
//
// AC: AC-GHK-004.4
func TestAGherkinReferenceAnchorIsSplitButNotResolvedAgainstGherkinStructure(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"```gherkin\n"+
		"Feature: Refunds\n"+
		"\n"+
		"  Scenario: A step names an anchored reference\n"+
		"    Given a step citing [[adr-012#Scenario: A step names an anchored reference]]\n"+
		"    When nothing else happens\n"+
		"    Then done\n"+
		"```\n")

	if len(found) != 1 {
		t.Fatalf("found %d references, want 1: %v", len(found), slugs(found))
	}
	reference := found[0]
	if reference.Slug != "adr-012" {
		t.Errorf("slug = %q, want adr-012", reference.Slug)
	}
	const wantAnchor = "Scenario: A step names an anchored reference"
	if reference.Anchor == nil || *reference.Anchor != wantAnchor {
		t.Errorf("anchor = %v, want %q recorded verbatim", reference.Anchor, wantAnchor)
	}

	// Resolution reads the slug half alone. Nothing here — or anywhere
	// resolution runs — treats the anchor as naming a Feature, Rule or
	// Scenario, so an anchor that happens to read like one plays no part
	// in what the reference resolves to.
	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})
	references.Resolve(found, index)
	if got := targetOf(t, found[0].Resolution); got != "specs/adrs/adr-012.md" {
		t.Errorf("resolved to %q; the anchor should play no part in resolution", got)
	}
}

// AC: AC-SPECREF-001.4
func TestAReferenceInAFrontmatterValueIsNotRecorded(t *testing.T) {
	// The description names the notation rather than a document. It is
	// the only unresolved reference a scan of this repository's corpus
	// produces, and it is a false one.
	found, _ := extract(t, `---
title: Reference conventions
description: References use [[slug]] wikilinks rather than paths
superseded_by: adr-013
---

The body cites [[adr-012]].
`, "superseded_by")

	if want := []string{"adr-013", "adr-012"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

// AC: AC-SPECREF-001.5
func TestRepeatedReferencesAreSeparateOccurrencesInSourceOrder(t *testing.T) {
	found, source := extract(t, `---
title: A document
---

The requirement cites [[adr-002]] in its Requirements.

And again in its References: [[adr-002]].
`)

	if len(found) != 2 {
		t.Fatalf("found %d references, want 2: %v", len(found), slugs(found))
	}
	if found[0].Span.Offset >= found[1].Span.Offset {
		t.Errorf("references are out of source order: %d then %d",
			found[0].Span.Offset, found[1].Span.Offset)
	}
	if found[0].Span == found[1].Span {
		t.Error("both occurrences carry the same span")
	}
	for _, reference := range found {
		if got := string(span.Slice(source, reference.Span)); got != "adr-002" {
			t.Errorf("a span covers %q", got)
		}
	}
}

// AC: AC-SPECREF-001.7
func TestRegisteredFrontmatterPropertiesContributeReferences(t *testing.T) {
	found, source := extract(t, `---
title: A module
contained_by:
  - "container-spec-cli"
frd: frd-spec-reference-integrity
tags:
  - not-a-reference
---

The body cites [[adr-012]].
`, "contained_by", "frd")

	if want := []string{"container-spec-cli", "frd-spec-reference-integrity", "adr-012"}; !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}

	origins := map[string]string{}
	for _, reference := range found {
		if property, ok := reference.Origin.AsFrontmatterProperty(); ok {
			origins[reference.Slug] = property.Property
			continue
		}
		if _, ok := reference.Origin.AsBody(); !ok {
			t.Errorf("%q carries neither origin", reference.Slug)
		}
	}
	if origins["container-spec-cli"] != "contained_by" {
		t.Errorf(`the sequence value names property %q`, origins["container-spec-cli"])
	}
	if origins["frd-spec-reference-integrity"] != "frd" {
		t.Errorf(`the scalar value names property %q`, origins["frd-spec-reference-integrity"])
	}
	if _, named := origins["adr-012"]; named {
		t.Error("the body reference names a frontmatter property")
	}

	// The span of a frontmatter reference covers the value as written,
	// quoting included, because that is the extent a rewrite replaces.
	if got := string(span.Slice(source, found[0].Span)); got != `"container-spec-cli"` {
		t.Errorf("the sequence element's span covers %q", got)
	}
	if got := string(span.Slice(source, found[1].Span)); got != "frd-spec-reference-integrity" {
		t.Errorf("the scalar's span covers %q", got)
	}
}

func TestASequenceValuedPropertyYieldsOneReferencePerItem(t *testing.T) {
	found, source := extract(t, `---
title: A component
contracts_consumed:
  - contract-specs-configuration
  - "contract-payments-api"
---

Body.
`, "contracts_consumed")

	if want := []string{"contract-specs-configuration", "contract-payments-api"}; !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}
	if found[0].Span == found[1].Span {
		t.Fatal("both items carry the span of the whole list")
	}
	if got := string(span.Slice(source, found[0].Span)); got != "contract-specs-configuration" {
		t.Errorf("the first item's span covers %q", got)
	}
	if found[0].Span.StartLine != 4 || found[1].Span.StartLine != 5 {
		t.Errorf("items are on lines %d and %d, want 4 and 5",
			found[0].Span.StartLine, found[1].Span.StartLine)
	}
}

// AC: AC-SPECANCH-001.1
func TestAnAnchorIsSplitOnTheFirstHashAndNotInterpreted(t *testing.T) {
	found, source := extract(t, `---
title: A document
---

[[container-spec-cli#Element Registry]] and [[adr-012#Decision#Nested]]
and [[adr-009]] and [[slug|text]].
`)

	if want := []string{"container-spec-cli", "adr-012", "adr-009", "slug|text"}; !equal(slugs(found), want) {
		t.Fatalf("slugs = %v, want %v", slugs(found), want)
	}

	if found[0].Anchor == nil || *found[0].Anchor != "Element Registry" {
		t.Errorf("anchor = %v, want the heading text", found[0].Anchor)
	}
	// One entry at one span: the anchor is carried alongside the slug on
	// the very reference the slug is recorded on — not a second
	// reference, and not a span of its own.
	if got := string(span.Slice(source, found[0].Span)); got != "container-spec-cli#Element Registry" {
		t.Errorf("the span covers %q, want slug and anchor together in one span", got)
	}
	// Split on the FIRST hash: what follows is recorded whole and not
	// read as a second anchor.
	if found[1].Anchor == nil || *found[1].Anchor != "Decision#Nested" {
		t.Errorf("anchor = %v, want everything after the first hash", found[1].Anchor)
	}
	if found[2].Anchor != nil {
		t.Errorf("a reference naming no anchor carries anchor %q", *found[2].Anchor)
	}
	// The display alias the reference standard forbids is not a
	// separator: the whole of it is the slug half, so it reports as
	// unresolved rather than resolving to something the standard bans.
	if found[3].Anchor != nil {
		t.Errorf("the display alias was read as an anchor: %q", *found[3].Anchor)
	}
	if found[3].Text != "slug|text" {
		t.Errorf("text = %q, want the reference as written", found[3].Text)
	}
}

// The index has to report the text a fix will edit, so an anchor is
// recorded exactly as authored: neither its internal whitespace nor its
// case is touched on the way in. Normalization is a matching question
// anchor resolution answers later, against corpus knowledge extraction
// does not have.
//
// AC: AC-SPECANCH-001.2
func TestAnAnchorIsRecordedWithNoNormalization(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"[[container-spec-cli#Element  Registry]] and [[adr-012#miXedCase]].\n")

	if len(found) != 2 {
		t.Fatalf("found %d references, want 2: %v", len(found), slugs(found))
	}
	const wantDoubleSpace = "Element  Registry"
	if found[0].Anchor == nil || *found[0].Anchor != wantDoubleSpace {
		t.Errorf("anchor = %v, want %q with its double space preserved", found[0].Anchor, wantDoubleSpace)
	}
	const wantMixedCase = "miXedCase"
	if found[1].Anchor == nil || *found[1].Anchor != wantMixedCase {
		t.Errorf("anchor = %v, want %q with its case preserved", found[1].Anchor, wantMixedCase)
	}
}

// The anchor separator is `#` alone. A vertical bar written inside an
// anchor is not read as the display alias delimiter — that notation
// [[cross-document-references]] prohibits reaches only the slug half —
// so the bar is ordinary text of the anchor, recorded whole.
//
// AC: AC-SPECANCH-001.5
func TestAVerticalBarInsideAnAnchorIsOrdinaryText(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"[[container-spec-cli#Element Registry|the registry]]\n")

	if len(found) != 1 {
		t.Fatalf("found %d references, want 1: %v", len(found), slugs(found))
	}
	if found[0].Slug != "container-spec-cli" {
		t.Errorf("slug = %q, want container-spec-cli — the bar plays no part in the slug half", found[0].Slug)
	}
	const want = "Element Registry|the registry"
	if found[0].Anchor == nil || *found[0].Anchor != want {
		t.Errorf("anchor = %v, want %q — the bar is ordinary text of the anchor", found[0].Anchor, want)
	}
}

// An anchor written in a reference-bearing frontmatter property is
// recorded exactly as one written in the body: same split on the first
// `#`, same lack of normalization, same span covering the value as
// written.
//
// This item participates in AC-SPECANCH-001.6; the SREF delta that lets
// a frontmatter citation carry an anchor owns the citation for it.
func TestAnAnchorInAFrontmatterPropertyIsRecordedLikeABodyAnchor(t *testing.T) {
	found, source := extract(t, `---
title: A document
sources:
  - "adr-012#Decision"
---

Body.
`, "sources")

	if len(found) != 1 {
		t.Fatalf("found %d references, want 1: %v", len(found), slugs(found))
	}
	reference := found[0]
	if reference.Slug != "adr-012" {
		t.Errorf("slug = %q, want adr-012", reference.Slug)
	}
	if reference.Anchor == nil || *reference.Anchor != "Decision" {
		t.Errorf("anchor = %v, want %q", reference.Anchor, "Decision")
	}
	if _, ok := reference.Origin.AsFrontmatterProperty(); !ok {
		t.Errorf("%q is not recorded as a frontmatter reference", reference.Text)
	}
	if got := string(span.Slice(source, reference.Span)); got != `"adr-012#Decision"` {
		t.Errorf("the span covers %q, want the value as written, quotes included", got)
	}
}

// An embed is written with a leading `!` before the delimiters, but
// extraction does not special-case it: nothing here looks for the `!`,
// so an embed is recorded exactly as the equivalent link is — same
// slug, same anchor, same text, same body origin.
//
// This item participates in AC-SPECANCH-002.7; the ARES delta that
// resolves and checks an embed owns the citation for it.
func TestAnEmbedIsRecordedAsTheEquivalentLink(t *testing.T) {
	embed, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"![[container-spec-cli#Element Registry]]\n")
	link, _ := extract(t, "---\ntitle: A document\n---\n\n"+
		"[[container-spec-cli#Element Registry]]\n")

	if len(embed) != 1 || len(link) != 1 {
		t.Fatalf("found %d embed reference(s) and %d link reference(s), want 1 each", len(embed), len(link))
	}
	if embed[0].Slug != link[0].Slug {
		t.Errorf("embed slug = %q, link slug = %q, want them equal", embed[0].Slug, link[0].Slug)
	}
	if (embed[0].Anchor == nil) != (link[0].Anchor == nil) ||
		(embed[0].Anchor != nil && *embed[0].Anchor != *link[0].Anchor) {
		t.Errorf("embed anchor = %v, link anchor = %v, want them equal", embed[0].Anchor, link[0].Anchor)
	}
	if embed[0].Text != link[0].Text {
		t.Errorf("embed text = %q, link text = %q, want them equal", embed[0].Text, link[0].Text)
	}
	if _, ok := embed[0].Origin.AsBody(); !ok {
		t.Error("the embed is not recorded as a body reference")
	}
}

func TestAnHTMLCommentBlockIsReadAsBodyText(t *testing.T) {
	found, _ := extract(t, `---
title: A document
---

<!-- Instructions: cite the decision as [[adr-012]] and delete this block. -->

Real prose.
`)

	if want := []string{"adr-012"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

func TestADocumentWithNoFrontmatterIsAllBody(t *testing.T) {
	source := []byte("No frontmatter here, but a reference to [[adr-012]].\n")
	front, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parsing: %s", perr.Message)
	}
	if front != nil {
		t.Fatal("the fixture parsed a frontmatter block it does not have")
	}

	found := references.Extract(references.Options{Path: "specs/fixture.md", Source: source})
	if want := []string{"adr-012"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

// An unregistered property holds no reference however slug-shaped its
// value is: what is a reference is registered, never guessed.
func TestAnUnregisteredPropertyContributesNothing(t *testing.T) {
	found, _ := extract(t, `---
title: A document
superseded_by: adr-013
---

Body.
`)

	if len(found) != 0 {
		t.Errorf("found %v with no property registered", slugs(found))
	}
}

// A registered property whose value is not text contributes nothing.
// Whether such a value belongs there is the schema's question.
func TestANonTextValueContributesNoReference(t *testing.T) {
	found, _ := extract(t, `---
title: A document
frd: 42
sources:
  - key: value
  - a-real-slug
---

Body.
`, "frd", "sources")

	if want := []string{"a-real-slug"}; !equal(slugs(found), want) {
		t.Errorf("slugs = %v, want %v", slugs(found), want)
	}
}

// Extraction says nothing about what a reference reached. Leaving the
// resolution unset is what makes an unresolved reference unserializable,
// so one cannot reach the index without having been resolved on the way.
func TestExtractionResolvesNothing(t *testing.T) {
	found, _ := extract(t, "---\ntitle: A document\n---\n\n[[adr-012]]\n")

	if len(found) != 1 {
		t.Fatalf("found %d references, want 1", len(found))
	}
	if found[0].Resolution.Value != nil {
		t.Errorf("extraction recorded a resolution: %+v", found[0].Resolution.Value)
	}
	if _, err := found[0].Resolution.MarshalJSON(); err == nil {
		t.Error("an unresolved reference serialized; it must not reach the index")
	}
}

// Every position this package reports is a Span from the Spans
// component, checked at compile time as well as by every span assertion
// above.
var _ span.Span = specmodel.Reference{}.Span

// The index's entries carry each document's typed frontmatter values but
// not their positions, which a reference needs. Reading one through an
// entry would lose them, so this package reads the parsed block instead — enforced here rather
// than left to review (agent-plugins-gz3).
func TestExtractionReadsNoIndexEntry(t *testing.T) {
	const forbidden = "github.com/brettcallaghan/specs/cli/core/specindex"

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	sources := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		sources++
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.FromSlash(name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquoting %s: %v", name, imported.Path.Value, err)
			}
			if path == forbidden {
				t.Errorf("%s imports %s; a reference-bearing value must not be reached "+
					"through an index entry's frontmatter", name, path)
			}
		}
	}
	if sources == 0 {
		t.Fatal("found no non-test source in this package; the check would pass vacuously")
	}
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
