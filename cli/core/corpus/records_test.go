// Tests for recording registered record blocks: that the index carries a
// block's tag and span and nothing of its content, that each block hangs
// off the innermost heading whose section contains it, and that one
// function decodes a recorded block's record.
package corpus_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

const recordFixture = `---
title: "With records"
description: "A fixture document."
slug: adr-001
type: adr
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

` + "```widget\nname: preface\n```" + `

# Title

## Roster

` + "```widget\nname: alpha\nkind: [one, two]\n```" + `

### Detail

` + "```widget\nname: beta\n```" + `

` + "```widget\nname: [unterminated\n```" + `

` + "```other\nname: ignored\n```" + `
`

func recordCorpus(t *testing.T, tags []corpus.RecordBlockTag) (*corpus.Corpus, string) {
	t.Helper()
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", recordFixture)
	profile := indexProfile().(stubProfile)
	profile.recordBlocks = tags
	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   profile,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built, repo
}

func headingNamed(t *testing.T, entry specmodel.IndexEntry, text string) specmodel.Heading {
	t.Helper()
	for _, heading := range entry.Headings {
		if heading.Text == text {
			return heading
		}
	}
	t.Fatalf("no heading %q", text)
	return specmodel.Heading{}
}

// A registered block is recorded with its tag and its span, fences
// included, attached to the innermost heading whose section contains it
// and to the entry where no heading does. Each appears exactly once,
// whether or not its content is valid, and no finding is raised.
//
// AC: AC-SPECSTR-001.2
func TestRegisteredRecordBlocksAreRecordedAndAttached(t *testing.T) {
	built, repo := recordCorpus(t, []corpus.RecordBlockTag{{Tag: "widget"}})
	entry := entryFor(t, built.Index(), "specs/adrs/adr-001.md")

	if len(entry.RecordBlocks) != 1 {
		t.Fatalf("entry carries %d record blocks, want the 1 before any heading", len(entry.RecordBlocks))
	}
	if got := len(headingNamed(t, entry, "Title").RecordBlocks); got != 0 {
		t.Errorf("Title carries %d blocks, want 0", got)
	}
	roster := headingNamed(t, entry, "Roster").RecordBlocks
	if len(roster) != 1 {
		t.Fatalf("Roster carries %d blocks, want 1", len(roster))
	}
	detail := headingNamed(t, entry, "Detail").RecordBlocks
	if len(detail) != 2 {
		t.Fatalf("Detail carries %d blocks, want 2 (the innermost heading owns them)", len(detail))
	}

	src, err := os.ReadFile(filepath.Join(repo, "specs/adrs/adr-001.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range append(append(roster, detail...), entry.RecordBlocks...) {
		if record.Tag != "widget" {
			t.Errorf("recorded tag %q, want widget", record.Tag)
		}
		text := string(src[record.Span.Offset : record.Span.Offset+record.Span.Length])
		if !strings.HasPrefix(text, "```widget\n") || !strings.HasSuffix(strings.TrimRight(text, "\n"), "```") {
			t.Errorf("span does not cover the block with its fences: %q", text)
		}
	}

	for _, f := range built.Findings() {
		if strings.Contains(f.Message, "widget") {
			t.Errorf("indexing raised a finding about a record: %+v", f)
		}
	}
}

// With no tag registered, every collection is empty, never null.
//
// AC: AC-SPECSTR-001.2
func TestNoRegisteredTagRecordsNoBlocks(t *testing.T) {
	built, _ := recordCorpus(t, nil)
	entry := entryFor(t, built.Index(), "specs/adrs/adr-001.md")

	if entry.RecordBlocks == nil || len(entry.RecordBlocks) != 0 {
		t.Errorf("entry RecordBlocks = %#v, want empty non-nil", entry.RecordBlocks)
	}
	for _, heading := range entry.Headings {
		if heading.RecordBlocks == nil || len(heading.RecordBlocks) != 0 {
			t.Errorf("%q RecordBlocks = %#v, want empty non-nil", heading.Text, heading.RecordBlocks)
		}
	}
}

// One function decodes a recorded block into the caller's own type.
//
// AC: AC-SPECSTR-001.2
func TestDecodeRecordBlockDecodesIntoACallerType(t *testing.T) {
	built, repo := recordCorpus(t, []corpus.RecordBlockTag{{Tag: "widget"}})
	entry := entryFor(t, built.Index(), "specs/adrs/adr-001.md")
	src, err := os.ReadFile(filepath.Join(repo, "specs/adrs/adr-001.md"))
	if err != nil {
		t.Fatal(err)
	}

	var widget struct {
		Name string   `yaml:"name"`
		Kind []string `yaml:"kind"`
	}
	if err := corpus.DecodeRecordBlock(src, headingNamed(t, entry, "Roster").RecordBlocks[0], &widget); err != nil {
		t.Fatalf("DecodeRecordBlock: %v", err)
	}
	if widget.Name != "alpha" || len(widget.Kind) != 2 {
		t.Errorf("decoded %+v, want alpha with two kinds", widget)
	}

	// Content that does not decode is reported by the decoder, not by
	// the index.
	bad := headingNamed(t, entry, "Detail").RecordBlocks[1]
	if err := corpus.DecodeRecordBlock(src, bad, &widget); err == nil {
		t.Error("an unterminated flow sequence decoded without error")
	}
	if err := corpus.DecodeRecordBlock(src[:10], bad, &widget); err == nil {
		t.Error("a span outside the document decoded without error")
	}
}

// A record inside a container block decodes from the content the parser
// reads, not from the raw lines with their prefixes.
//
// AC: AC-SPECSTR-001.2
func TestDecodeRecordBlockDecodesRecordsInsideContainers(t *testing.T) {
	cases := map[string]string{
		"blockquote": "# T\n\n> ```component\n> name: x\n> kind: y\n> ```\n",
		"list item":  "# T\n\n- item\n\n  ```component\n  name: x\n  kind: y\n  ```\n",
		"top level":  "# T\n\n```component\nname: x\nkind: y\n```\n",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			src := []byte(doc)
			found := markdown.Fenced(src, []string{"component"})
			if len(found) != 1 {
				t.Fatalf("found %d blocks, want 1", len(found))
			}
			block := specmodel.RecordBlock{Tag: "component", Span: found[0].Outer}
			var got struct {
				Name string `yaml:"name"`
				Kind string `yaml:"kind"`
			}
			if err := corpus.DecodeRecordBlock(src, block, &got); err != nil {
				t.Fatalf("DecodeRecordBlock: %v", err)
			}
			if got.Name != "x" || got.Kind != "y" {
				t.Errorf("decoded %+v, want name x kind y", got)
			}
		})
	}
}

// A span that does not cover a fenced block of the block's tag is an
// error, not an empty decode.
//
// AC: AC-SPECSTR-001.2
func TestDecodeRecordBlockRejectsASpanThatIsNotABlockOfItsTag(t *testing.T) {
	src := []byte("# T\n\n```component\nname: x\n```\n\n```other\nname: y\n```\n")
	found := markdown.Fenced(src, []string{"component"})
	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	var got map[string]any
	if err := corpus.DecodeRecordBlock(src, specmodel.RecordBlock{Tag: "component", Span: found[0].Inner}, &got); err == nil {
		t.Error("an inner-content span decoded without error")
	}
	if err := corpus.DecodeRecordBlock(src, specmodel.RecordBlock{Tag: "other", Span: found[0].Outer}, &got); err == nil {
		t.Error("a span of a block with another tag decoded without error")
	}
}

// The registered tags decide what the index records, so changing them
// changes the input hash.
func TestChangingTheRecordBlockTagsChangesTheInputHash(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", recordFixture)
	hash := func(tags []corpus.RecordBlockTag) string {
		profile := indexProfile().(stubProfile)
		profile.recordBlocks = tags
		got, err := corpus.Fingerprint(corpus.Options{
			RepoRoot: repo, Root: filepath.Join(repo, "specs"),
			Workspace: "fixture-workspace", Profile: profile,
		})
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	if hash(nil) == hash([]corpus.RecordBlockTag{{Tag: "widget"}}) {
		t.Error("registering a record-block tag left the hash unchanged")
	}
}
