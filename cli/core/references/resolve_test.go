// Tests for reference resolution.
//
// Every fixture here builds an index by hand rather than by walking a
// corpus. Resolution is a function of the index and nothing else — no
// document is read and no frontmatter is parsed — so an index literal is
// the whole of the input, and a test that had to write files to state
// its case would be testing the corpus walk instead.
package references_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/references"
	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// member returns an index entry for a corpus member declaring one slug.
func member(path, slug string) specmodel.IndexEntry {
	return specmodel.IndexEntry{
		Path:        path,
		Frontmatter: &specmodel.DocumentFrontmatter{Slug: specmodel.Slug(slug)},
	}
}

// indexOf returns an index holding the given entries and exclusions,
// built through the same constructor an emitter uses so its collections
// are shaped the way a real index's are.
func indexOf(entries []specmodel.IndexEntry, exclusions ...specmodel.IndexExclusion) specmodel.SpecIndex {
	index := specindex.New("fixture-workspace")
	index.Entries = append(index.Entries, entries...)
	index.Exclusions = append(index.Exclusions, exclusions...)
	return index
}

// resolveOne resolves a single reference carrying the given slug half and
// returns its resolution.
func resolveOne(slug string, index specmodel.SpecIndex) specmodel.ReferenceResolution {
	refs := []specmodel.Reference{{Path: "specs/citing.md", Text: slug, Slug: slug}}
	references.Resolve(refs, index)
	return refs[0].Resolution
}

// targetOf returns the path a resolution resolved to, failing the test
// when it resolved to anything else.
func targetOf(t *testing.T, resolution specmodel.ReferenceResolution) string {
	t.Helper()

	resolved, ok := resolution.AsResolved()
	if !ok {
		t.Fatalf("resolution = %+v, want a resolved reference", resolution.Value)
	}
	return resolved.TargetPath
}

// AC: AC-SPECREF-002.8
func TestASlugResolvesComparedLowercased(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})

	for _, written := range []string{"adr-012", "ADR-012", "Adr-012", " adr-012 "} {
		if got := targetOf(t, resolveOne(written, index)); got != "specs/adrs/adr-012.md" {
			t.Errorf("[[%s]] resolved to %q", written, got)
		}
	}

	// The corpus's own slug is compared the same way, so a member that
	// declared it in capitals is found by a reference written in lower
	// case — the two halves of AC-SPEC-003.4's case-insensitive
	// uniqueness have to agree.
	shouting := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "ADR-012")})
	if got := targetOf(t, resolveOne("adr-012", shouting)); got != "specs/adrs/adr-012.md" {
		t.Errorf("a reference in lower case resolved to %q against a member declaring ADR-012", got)
	}
}

// A reference carrying an anchor resolves on its slug portion alone. The
// anchor is recorded and not interpreted, and nothing about it is
// reported: whether the target has that heading is a question this
// feature does not ask.
//
// AC: AC-SPECREF-002.2
func TestAnAnchoredReferenceResolvesOnItsSlugAlone(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{
		member("specs/architecture/container-spec-cli.md", "container-spec-cli"),
	})

	anchor := "Element Registry"
	refs := []specmodel.Reference{{
		Path:   "specs/citing.md",
		Text:   "container-spec-cli#Element Registry",
		Slug:   "container-spec-cli",
		Anchor: &anchor,
	}}
	references.Resolve(refs, index)

	if got := targetOf(t, refs[0].Resolution); got != "specs/architecture/container-spec-cli.md" {
		t.Errorf("resolved to %q, want the document declaring the slug half", got)
	}
	if refs[0].Anchor == nil || *refs[0].Anchor != anchor {
		t.Errorf("anchor = %v, want it recorded unchanged", refs[0].Anchor)
	}

	// The whole reference read as a slug matches nothing, which is what
	// this criterion exists to prevent: every anchored reference in a
	// corpus would otherwise report as broken.
	if _, ok := resolveOne("container-spec-cli#Element Registry", index).AsUnresolved(); !ok {
		t.Error("the unsplit text resolved; the fixture no longer proves anything")
	}
}

// A slug matching a file the index recorded as an exclusion, or as a
// member whose frontmatter could not be read, is held apart from a slug
// matching nothing at all: the document is right there, and what is wrong
// is its frontmatter rather than the reference.
//
// AC: AC-SPECREF-002.7
func TestASlugMatchingANonMemberRecordsThatReasonAndPath(t *testing.T) {
	parseError := &specmodel.ParseError{Message: "could not find end character of double-quoted text"}
	index := indexOf(
		[]specmodel.IndexEntry{
			member("specs/adrs/adr-012.md", "adr-012"),
			{Path: "specs/adrs/adr-099-unreadable.md", ParseError: parseError},
		},
		specmodel.IndexExclusion{
			Path:   "specs/templates/work-order-template.md",
			Reason: specmodel.ExclusionReasonUnknownType,
		},
	)

	unreadable, ok := resolveOne("adr-099-unreadable", index).AsNonMember()
	if !ok {
		t.Fatalf("a slug matching a parse failure did not resolve as a non-member")
	}
	if unreadable.Path != "specs/adrs/adr-099-unreadable.md" {
		t.Errorf("path = %q, want the matched file's own path", unreadable.Path)
	}
	if unreadable.ParseError == nil || unreadable.ParseError.Message != parseError.Message {
		t.Errorf("parse_error = %+v, want the failure the index recorded", unreadable.ParseError)
	}
	if unreadable.Reason != nil {
		t.Errorf("reason = %v, want none: the file is an entry, not an exclusion", *unreadable.Reason)
	}

	excluded, ok := resolveOne("work-order-template", index).AsNonMember()
	if !ok {
		t.Fatalf("a slug matching an exclusion did not resolve as a non-member")
	}
	if excluded.Path != "specs/templates/work-order-template.md" {
		t.Errorf("path = %q, want the excluded file's own path", excluded.Path)
	}
	if excluded.Reason == nil || *excluded.Reason != specmodel.ExclusionReasonUnknownType {
		t.Errorf("reason = %v, want the reason the index recorded", excluded.Reason)
	}
	if excluded.ParseError != nil {
		t.Errorf("parse_error = %+v, want none: the file is an exclusion, not a failed entry", excluded.ParseError)
	}

	// The distinction is the point of the criterion, so it is asserted
	// rather than assumed: a slug matching nothing at all is a different
	// state, not the same one with empty fields.
	if _, ok := resolveOne("adr-404", index).AsUnresolved(); !ok {
		t.Error("a slug matching nothing did not resolve as unresolved")
	}
}

// A member always wins. A file beneath the spec root named like a
// document that exists must not shadow the document itself.
func TestAMemberWinsOverAFileNamedLikeIt(t *testing.T) {
	index := indexOf(
		[]specmodel.IndexEntry{member("specs/adrs/adr-012-reference-conventions.md", "adr-012")},
		specmodel.IndexExclusion{
			Path:   "specs/drafts/adr-012.md",
			Reason: specmodel.ExclusionReasonNoFrontmatter,
		},
	)

	if got := targetOf(t, resolveOne("adr-012", index)); got != "specs/adrs/adr-012-reference-conventions.md" {
		t.Errorf("resolved to %q, want the member declaring the slug", got)
	}
}

// Resolution is only ever within the workspace the run is against. A slug
// another repository's index declares is unresolved here, per the FRD's
// stated boundary; the merged-graph case is a known follow-on.
func TestASlugDeclaredOnlyOutsideTheWorkspaceIsUnresolved(t *testing.T) {
	here := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})

	elsewhere := specindex.New("another-workspace")
	elsewhere.Entries = append(elsewhere.Entries, member("specs/adrs/adr-500.md", "adr-500"))

	if _, ok := resolveOne("adr-500", here).AsUnresolved(); !ok {
		t.Error("a slug declared only in another workspace's index resolved")
	}
	// Resolved against the index that does declare it, so the fixture is
	// proving a boundary rather than a typo.
	if got := targetOf(t, resolveOne("adr-500", elsewhere)); got != "specs/adrs/adr-500.md" {
		t.Errorf("the other workspace's own index resolved adr-500 to %q", got)
	}
}

// Every reference gets a resolution, including the ones that reached
// nothing: an unresolved reference is recorded rather than dropped, which
// is what lets the index describe a corpus whose links are broken.
func TestEveryReferenceIsGivenAResolution(t *testing.T) {
	index := indexOf([]specmodel.IndexEntry{member("specs/adrs/adr-012.md", "adr-012")})

	refs := []specmodel.Reference{
		{Path: "specs/a.md", Slug: "adr-012"},
		{Path: "specs/a.md", Slug: "adr-404"},
		{Path: "specs/a.md", Slug: ""},
	}
	references.Resolve(refs, index)

	for _, reference := range refs {
		if reference.Resolution.Value == nil {
			t.Errorf("%q was left unresolved, which cannot be serialized", reference.Slug)
			continue
		}
		if _, err := reference.Resolution.MarshalJSON(); err != nil {
			t.Errorf("%q resolved to something unserializable: %v", reference.Slug, err)
		}
	}
	if _, ok := refs[2].Resolution.AsUnresolved(); !ok {
		t.Error("a reference with an empty slug half did not resolve as unresolved")
	}
}

// The answer does not depend on the order the index holds its
// collections in: where two files answer to one key, the first in path
// order wins by comparison rather than by arrival.
func TestResolutionDoesNotDependOnIndexOrder(t *testing.T) {
	first := specmodel.IndexExclusion{
		Path:   "specs/a/duplicate.md",
		Reason: specmodel.ExclusionReasonNoFrontmatter,
	}
	second := specmodel.IndexExclusion{
		Path:   "specs/b/duplicate.md",
		Reason: specmodel.ExclusionReasonUnknownType,
	}

	forward, ok := resolveOne("duplicate", indexOf(nil, first, second)).AsNonMember()
	if !ok {
		t.Fatal("the slug did not resolve as a non-member")
	}
	reversed, ok := resolveOne("duplicate", indexOf(nil, second, first)).AsNonMember()
	if !ok {
		t.Fatal("the slug did not resolve as a non-member when the exclusions were reversed")
	}
	if forward.Path != reversed.Path || forward.Path != first.Path {
		t.Errorf("resolved to %q and %q, want %q either way", forward.Path, reversed.Path, first.Path)
	}
}
