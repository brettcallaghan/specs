package numbering_test

import (
	"errors"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/numbering"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// rules is the numbered-type table this test file exercises Next
// against: one numbered type ("adr") and one type the profile does not
// register as numbered ("frd"), mirroring the shape cli/profile
// supplies in production.
var rules = map[string]corpus.DocumentTypeRules{
	"adr": {SlugPattern: `^adr-[0-9]{3,}(-[a-z0-9]+(-[a-z0-9]+)*)?$`, Numbered: true},
	"frd": {SlugPattern: `^frd-[a-z0-9]+(-[a-z0-9]+)*$`, Numbered: false},
}

// entry builds an index entry declaring the given type, slug and path,
// with a parsed frontmatter envelope — enough for Next, which reads
// only Type and Slug from it.
func entry(path, docType, slug string) specmodel.IndexEntry {
	return specmodel.IndexEntry{
		Path: path,
		Frontmatter: &specmodel.DocumentFrontmatter{
			Type: specmodel.DocumentType(docType),
			Slug: specmodel.Slug(slug),
		},
	}
}

// AC: AC-SPEC-007.1
func TestNextOnAnEmptyCorpusReturnsOne(t *testing.T) {
	index := specindexOf()

	next, err := numbering.Next(index, "adr", rules)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != 1 {
		t.Errorf("Next = %d, want 1", next)
	}
}

// AC: AC-SPEC-007.1
func TestNextIsOneGreaterThanTheHighestAcrossAGapInTheSequence(t *testing.T) {
	index := specindexOf(
		entry("specs/adrs/ADR-001-x.md", "adr", "adr-001-x"),
		entry("specs/adrs/ADR-003-x.md", "adr", "adr-003-x"),
		entry("specs/adrs/ADR-010-x.md", "adr", "adr-010"),
	)

	next, err := numbering.Next(index, "adr", rules)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != 11 {
		t.Errorf("Next = %d, want 11 (one greater than the highest, gaps notwithstanding)", next)
	}
}

// AC: AC-SPEC-007.1
func TestNextIsIndependentOfDirectory(t *testing.T) {
	index := specindexOf(
		entry("specs/adrs/ADR-005-x.md", "adr", "adr-005-x"),
		entry("archive/old/ADR-013-y.md", "adr", "adr-013-y"),
		entry("elsewhere/nested/deep/ADR-009-z.md", "adr", "adr-009-z"),
	)

	next, err := numbering.Next(index, "adr", rules)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != 14 {
		t.Errorf("Next = %d, want 14; the highest number (13) lives outside the conventional adrs/ directory "+
			"and must still be found", next)
	}
}

// AC: AC-SPEC-007.1
func TestNextOnlyCountsMembersOfTheRequestedType(t *testing.T) {
	index := specindexOf(
		entry("specs/adrs/ADR-002-x.md", "adr", "adr-002-x"),
		entry("specs/frds/frd-checkout.md", "frd", "frd-checkout"),
	)

	next, err := numbering.Next(index, "adr", rules)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != 3 {
		t.Errorf("Next = %d, want 3; an frd member must not influence the adr count", next)
	}
}

// A type the active profile does not register as numbered — present in
// the table but with Numbered false — is refused rather than answered
// from whatever slugs happen to match.
//
// AC: AC-SPEC-007.1
func TestNextRefusesATypeNotRegisteredAsNumbered(t *testing.T) {
	index := specindexOf(entry("specs/frds/frd-checkout.md", "frd", "frd-checkout"))

	_, err := numbering.Next(index, "frd", rules)
	if !errors.Is(err, numbering.ErrNotNumbered) {
		t.Errorf("Next err = %v, want ErrNotNumbered", err)
	}
}

// A type absent from the table altogether is refused the same way as
// one registered with Numbered false — both mean the profile offers no
// next-id query for it.
//
// AC: AC-SPEC-007.1
func TestNextRefusesATypeAbsentFromTheTable(t *testing.T) {
	index := specindexOf()

	_, err := numbering.Next(index, "guide", rules)
	if !errors.Is(err, numbering.ErrNotNumbered) {
		t.Errorf("Next err = %v, want ErrNotNumbered", err)
	}
}

// A member whose frontmatter failed to parse carries no Type or Slug to
// read, and must not panic or otherwise disrupt the count.
//
// AC: AC-SPEC-007.1
func TestNextSkipsEntriesWithNoParsedFrontmatter(t *testing.T) {
	index := specindexOf(
		entry("specs/adrs/ADR-004-x.md", "adr", "adr-004-x"),
		specmodel.IndexEntry{
			Path: "specs/adrs/ADR-broken.md",
			ParseError: &specmodel.ParseError{
				Message: "mapping values are not allowed in this context",
			},
		},
	)

	next, err := numbering.Next(index, "adr", rules)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if next != 5 {
		t.Errorf("Next = %d, want 5", next)
	}
}

func specindexOf(entries ...specmodel.IndexEntry) specmodel.SpecIndex {
	index := specmodel.SpecIndex{
		SchemaVersion: "1.0",
		Workspace:     "agent-plugins",
		Entries:       []specmodel.IndexEntry{},
		Exclusions:    []specmodel.IndexExclusion{},
	}
	index.Entries = append(index.Entries, entries...)
	return index
}
