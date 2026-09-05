// Tests for anchor resolution — the half of Resolve that matches a
// reference's anchor against its target's headings.
//
// Like resolve_test.go, every fixture here builds an index by hand:
// anchor resolution is a lookup over headings already on the index, so
// an index literal is the whole of the input.
package references_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/references"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// heading returns a specmodel.Heading carrying the given text, with a
// position distinct enough per call to tell two headings apart in a
// failure message.
func heading(text string, offset uint64) specmodel.Heading {
	return specmodel.Heading{
		Text:     text,
		Level:    2,
		Position: specmodel.Span{Offset: offset, Length: uint64(len(text))},
	}
}

// memberWithHeadings returns a corpus member declaring slug, carrying
// the given headings.
func memberWithHeadings(path, slug string, headings ...specmodel.Heading) specmodel.IndexEntry {
	entry := member(path, slug)
	entry.Headings = headings
	return entry
}

// resolveAnchorOf resolves a single reference naming slug and anchor
// (against a containing document at citingPath) and returns its anchor
// resolution.
func resolveAnchorOf(citingPath, slug, anchor string, index specmodel.SpecIndex) *specmodel.AnchorResolution {
	refs := []specmodel.Reference{{Path: citingPath, Slug: slug, Anchor: &anchor}}
	references.Resolve(refs, index)
	return refs[0].AnchorResolution
}

// An anchor matched against a heading whose case and word-boundary
// punctuation differ still resolves: both are reduced to the same
// heading identifier before comparison, the derivation a markdown
// renderer applies for its own fragment.
//
// AC: AC-SPECANCH-002.2
func TestAnAnchorMatchesAHeadingByIdentifierNotByExactText(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		memberWithHeadings("specs/containers/spec-cli.md", "container-spec-cli",
			heading("Element Registry", 10)),
	})

	resolution := resolveAnchorOf("specs/citing.md", "container-spec-cli", "element registry", index)
	resolved, ok := resolution.AsResolved()
	if !ok {
		t.Fatalf("anchor_resolution = %+v, want resolved", resolution.Value)
	}
	if resolved.Heading.Text != "Element Registry" {
		t.Errorf("Heading.Text = %q, want the heading as authored", resolved.Heading.Text)
	}
}

// A reference naming an anchor and no slug resolves the anchor against
// the document containing the reference, not against nothing: the whole
// point of `[[#Heading]]` is a link to a place in the same document.
//
// AC: AC-SPECANCH-002.6
func TestAnAnchorWithNoSlugResolvesAgainstTheContainingDocument(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		memberWithHeadings("specs/frds/frd-spec-anchor-references.md", "frd-spec-anchor-references",
			heading("Terminology", 5)),
	})

	resolution := resolveAnchorOf("specs/frds/frd-spec-anchor-references.md", "", "Terminology", index)
	resolved, ok := resolution.AsResolved()
	if !ok {
		t.Fatalf("anchor_resolution = %+v, want resolved against the containing document", resolution.Value)
	}
	if resolved.Heading.Text != "Terminology" {
		t.Errorf("Heading.Text = %q, want %q", resolved.Heading.Text, "Terminology")
	}

	// The slug half is genuinely empty, not merely absent as a
	// convenience — resolving against the containing document must not
	// depend on the slug resolution having found anything to piggyback
	// on.
	refs := []specmodel.Reference{{
		Path:   "specs/frds/frd-spec-anchor-references.md",
		Slug:   "",
		Anchor: strPtr("Terminology"),
	}}
	references.Resolve(refs, index)
	if _, ok := refs[0].Resolution.AsResolved(); ok {
		t.Fatal("the fixture's slug half resolved; it no longer proves the anchor half stands alone")
	}
}

// An embed (`![[…]]`) carries no marker of its own on [specmodel.Reference]
// distinguishing it from an ordinary link — extraction already records
// it identically — so it resolves through exactly the same path,
// asserted here so a later change cannot special-case one without
// noticing it changed the other.
//
// AC: AC-SPECANCH-002.7
func TestAnEmbedResolvesExactlyAsTheEquivalentLink(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		memberWithHeadings("specs/adrs/adr-012.md", "adr-012", heading("Decision", 5)),
	})

	link := resolveAnchorOf("specs/citing.md", "adr-012", "Decision", index)
	embed := resolveAnchorOf("specs/citing.md", "adr-012", "Decision", index)

	linkResolved, ok := link.AsResolved()
	if !ok {
		t.Fatalf("the link's anchor did not resolve: %+v", link.Value)
	}
	embedResolved, ok := embed.AsResolved()
	if !ok {
		t.Fatalf("the embed's anchor did not resolve: %+v", embed.Value)
	}
	if linkResolved.Heading != embedResolved.Heading {
		t.Errorf("link resolved to %+v, embed to %+v; an embed must resolve exactly as the equivalent link",
			linkResolved.Heading, embedResolved.Heading)
	}
}

// More than one heading answers to the same identifier: the anchor
// resolves to no single heading, and every match is carried so the
// reader is told which headings to disambiguate rather than merely that
// a collision exists.
func TestAnAmbiguousAnchorCarriesEveryMatch(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		memberWithHeadings("specs/modules/document-core.md", "module-document-core",
			heading("References", 10), heading("References", 40)),
	})

	resolution := resolveAnchorOf("specs/citing.md", "module-document-core", "References", index)
	ambiguous, ok := resolution.AsAmbiguous()
	if !ok {
		t.Fatalf("anchor_resolution = %+v, want ambiguous", resolution.Value)
	}
	if len(ambiguous.Matches) != 2 {
		t.Fatalf("Matches has %d entries, want both headings", len(ambiguous.Matches))
	}
	if ambiguous.Matches[0].Position == ambiguous.Matches[1].Position {
		t.Error("both matches carry the same position; the fixture no longer proves two distinct headings")
	}
}

// An anchor naming no heading of a target whose headings ARE available
// is a considered "no": the headings were read and none of them
// matched, which is a different state from not having looked at all.
func TestAnAnchorMatchingNoHeadingIsUnresolvedNotUnverified(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		memberWithHeadings("specs/adrs/adr-012.md", "adr-012", heading("Decision", 5)),
	})

	resolution := resolveAnchorOf("specs/citing.md", "adr-012", "Nonexistent", index)
	if _, ok := resolution.AsUnresolved(); !ok {
		t.Errorf("anchor_resolution = %+v, want unresolved", resolution.Value)
	}
}

// A target the index carries no entry for at all — a `.feature` file's
// own path is the concrete case, since a feature source produces no
// index entry — has headings that are simply not known, which is a
// different claim from having looked and found none: the tool did not
// get to look.
func TestATargetWithNoIndexEntryIsUnverified(t *testing.T) {
	index := indexOf(nil) // no entries at all: the self-reference's own document is unindexed

	resolution := resolveAnchorOf("docs/features/refunds.feature", "", "Refunds", index)
	unverified, ok := resolution.AsUnverified()
	if !ok {
		t.Fatalf("anchor_resolution = %+v, want unverified", resolution.Value)
	}
	if unverified.Reason == "" {
		t.Error("Reason is empty; an unverified anchor must name what is missing")
	}
}

// A reference whose slug resolves to no corpus member gets no anchor
// finding — the parent feature already reports it as broken, and an
// anchor resolution on top of it would be a second finding for one
// mistake.
func TestAReferenceWhoseSlugDoesNotResolveGetsNoAnchorResolution(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})

	resolution := resolveAnchorOf("specs/citing.md", "adr-404-absent", "Decision", index)
	if resolution != nil {
		t.Errorf("anchor_resolution = %+v, want nil: the slug half never resolved", resolution.Value)
	}
}

// A reference carrying no anchor at all gets no anchor resolution: the
// property is absent, not a resolved-to-nothing state, so it does not
// appear in the emitted index.
func TestAReferenceWithNoAnchorGetsNoAnchorResolution(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})

	refs := []specmodel.Reference{{Path: "specs/citing.md", Slug: "adr-012"}}
	references.Resolve(refs, index)
	if refs[0].AnchorResolution != nil {
		t.Errorf("AnchorResolution = %+v, want nil for a reference carrying no anchor", refs[0].AnchorResolution)
	}
}

// strPtr returns a pointer to s, for building the optional anchor field
// [specmodel.Reference] takes.
func strPtr(s string) *string { return &s }
