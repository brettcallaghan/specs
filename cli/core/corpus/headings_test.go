// Tests for the heading half of the corpus walk: that every member's
// entry carries the headings Markdown Structure extracts, in document
// order and in the document's own coordinates, and that a warm corpus
// answers for them from the index it opened.
package corpus_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// headingsOf returns the entry for path, failing the test when there is
// none.
func entryFor(t *testing.T, index specmodel.SpecIndex, path string) specmodel.IndexEntry {
	t.Helper()
	for _, entry := range index.Entries {
		if entry.Path == path {
			return entry
		}
	}
	t.Fatalf("no entry for %q in %v", path, paths(index))
	return specmodel.IndexEntry{}
}

// texts returns each heading's text, for comparing order and content in
// one assertion.
func texts(headings []specmodel.Heading) []string {
	out := make([]string, 0, len(headings))
	for _, heading := range headings {
		out = append(out, heading.Text)
	}
	return out
}

// A built corpus carries every member's headings on its index entry, in
// document order, with the span of the content beneath each — and the
// frontmatter block is not among them.
//
// AC: AC-SPECSTR-001.1
func TestABuiltCorpusCarriesEveryMembersHeadings(t *testing.T) {
	repo := t.TempDir()
	body := `---
title: "With headings"
description: "A fixture document."
slug: adr-001
type: adr
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

# Title

intro

## Context

why

### Detail

deeper

## Decision

what
`
	write(t, repo, "specs/adrs/adr-001.md", body)

	index := corpusAt(t, repo).Index()
	entry := entryFor(t, index, "specs/adrs/adr-001.md")

	want := []string{"Title", "Context", "Detail", "Decision"}
	if got := texts(entry.Headings); !equal(got, want) {
		t.Fatalf("headings = %v, want %v", got, want)
	}

	levels := []uint64{1, 2, 3, 2}
	for i, level := range levels {
		if entry.Headings[i].Level != level {
			t.Errorf("heading %d (%q) level = %d, want %d",
				i, entry.Headings[i].Text, entry.Headings[i].Level, level)
		}
	}

	// The spans address the file on disk, frontmatter included, so a
	// consumer slices the section straight out of the bytes it read.
	src, err := os.ReadFile(filepath.Join(repo, "specs", "adrs", "adr-001.md"))
	if err != nil {
		t.Fatalf("reading the fixture back: %v", err)
	}
	title := entry.Headings[0]
	if got, want := string(span.Slice(src, title.Position)), "# Title\n"; got != want {
		t.Errorf("the first heading's position slices to %q, want %q", got, want)
	}
	context := entry.Headings[1]
	if got, want := string(span.Slice(src, context.Section)), "\nwhy\n\n### Detail\n\ndeeper\n\n"; got != want {
		t.Errorf("Context's section slices to %q, want %q", got, want)
	}
}

// A member whose frontmatter could not be parsed still carries its
// headings: a heading is markdown and does not depend on an envelope.
//
// AC: AC-SPECSTR-001.1
func TestAMemberWhoseFrontmatterFailedToParseStillCarriesItsHeadings(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-002.md", `---
title: "Unterminated
description: "The quoted string above never closes."
slug: adr-002
type: adr
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

# Broken but readable

## Still a heading
`)

	index := corpusAt(t, repo).Index()
	entry := entryFor(t, index, "specs/adrs/adr-002.md")

	if entry.ParseError == nil {
		t.Fatal("the fixture was expected to fail frontmatter parsing")
	}
	want := []string{"Broken but readable", "Still a heading"}
	if got := texts(entry.Headings); !equal(got, want) {
		t.Errorf("headings = %v, want %v", got, want)
	}
}

// A document with no heading carries an empty collection rather than a
// null one, so the emitted index compares byte for byte and a consumer
// iterating it never distinguishes "empty" from "absent".
//
// AC: AC-SPECSTR-001.1
func TestAMemberWithNoHeadingCarriesAnEmptyCollection(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-003.md", `---
title: "No headings"
description: "A fixture document."
slug: adr-003
type: adr
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

Just prose, and no heading anywhere.
`)

	index := corpusAt(t, repo).Index()
	entry := entryFor(t, index, "specs/adrs/adr-003.md")

	if entry.Headings == nil {
		t.Error("Headings is nil; want an empty collection")
	}
	if len(entry.Headings) != 0 {
		t.Errorf("headings = %v, want none", texts(entry.Headings))
	}
	if entry.RecordBlocks == nil {
		t.Error("RecordBlocks is nil; want an empty collection")
	}
}

// Two builds over an unchanged corpus emit the same headings in the same
// order, which is what keeps the emitted index byte-identical.
//
// AC: AC-SPEC-009.7
func TestHeadingsAreOrderedIdenticallyAcrossBuilds(t *testing.T) {
	repo := t.TempDir()
	for _, name := range []string{"adr-001", "adr-002", "adr-003"} {
		write(t, repo, "specs/adrs/"+name+".md", `---
title: "`+name+`"
description: "A fixture document."
slug: `+name+`
type: adr
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

# One

## Two

### Three

## Four
`)
	}

	first := rendered(t, corpusAt(t, repo).Index())
	second := rendered(t, corpusAt(t, repo).Index())

	if string(first) != string(second) {
		t.Error("two builds over an unchanged corpus rendered different bytes")
	}

	index := corpusAt(t, repo).Index()
	want := []string{"One", "Two", "Three", "Four"}
	for _, entry := range index.Entries {
		if got := texts(entry.Headings); !equal(got, want) {
			t.Errorf("%s: headings = %v, want %v", entry.Path, got, want)
		}
	}
}

// A warm corpus answers for headings from the index it opened rather than
// by reading the document off disk.
//
// The proof is a planted heading: the persisted index is doctored to name
// a heading the document does not contain, leaving the recorded input
// hash alone. A rebuild could not produce that heading, so an answer
// carrying it can only have come from the file.
//
// AC: AC-SPECSTR-001.1
func TestAWarmCorpusAnswersForHeadingsWithoutReadingTheDocument(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))

	opts := indexOptions(repo)
	dest := buildAndWrite(t, repo)

	persisted := readIndex(t, dest)
	if len(persisted.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(persisted.Entries))
	}
	persisted.Entries[0].Headings = append(persisted.Entries[0].Headings, specmodel.Heading{
		Text:         "Planted in the persisted index and nowhere else",
		Level:        2,
		Position:     span.Of(nil, 0, 0),
		Section:      span.Of(nil, 0, 0),
		RecordBlocks: []specmodel.RecordBlock{},
	})
	if err := corpus.Write(persisted, dest, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	opened, warm, err := corpus.Open(opts, dest)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !warm {
		t.Fatal("Open rebuilt an index whose inputs are unchanged")
	}

	entry := entryFor(t, opened.Index(), "specs/adrs/adr-001.md")
	if got := texts(entry.Headings); len(got) != 2 || got[1] != "Planted in the persisted index and nowhere else" {
		t.Errorf("headings = %v; the planted heading is absent, so the document was re-read rather than served warm", got)
	}
}
