// Internal tests for the properties of the walk that are not observable
// from outside the package.
package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every source is offered to the builder exactly once, with its bytes,
// and each is offered as the kind of source it is.
//
// This is what "no document is read twice" rests on. [scanSources] is
// the only walk in the package, [Build] and [Fingerprint] both consume
// it, and reference and feature extraction are fed by the same offer that
// carries a file's bytes to classification — so a second read of any
// document would have to be a second walk, and a second walk would show
// up here as a second offer. That the feature sources ride this same walk
// is what keeps their bytes in the input hash.
func TestTheWalkOffersEachSourceExactlyOnce(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"adrs/adr-001.md":        "---\ntitle: One\n---\n\n[[adr-002]]\n",
		"adrs/nested/adr-002.md": "---\ntitle: Two\n---\n\n[[adr-001]]\n",
		"notes/loose.md":         "# No frontmatter\n",
		"notes/not-markdown.txt": "passed over silently\n",
		"features/one.feature":   "Feature: One\n",
		"features/UPPER.FEATURE": "Feature: Two\n",
	}
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", full, err)
		}
	}

	offers := map[string]int{}
	kinds := map[string]sourceKind{}
	err := scanSources(root, root, func(kind sourceKind, reported string, src []byte, readErr error) {
		offers[reported]++
		kinds[reported] = kind
		if readErr != nil {
			t.Errorf("%s could not be read: %v", reported, readErr)
			return
		}
		if string(src) != files[reported] {
			t.Errorf("%s was offered %q", reported, src)
		}
	})
	if err != nil {
		t.Fatalf("scanSources: %v", err)
	}

	want := map[string]sourceKind{
		"adrs/adr-001.md":        markdownCandidate,
		"adrs/nested/adr-002.md": markdownCandidate,
		"notes/loose.md":         markdownCandidate,
		"features/one.feature":   featureSource,
		"features/UPPER.FEATURE": featureSource,
	}
	if len(offers) != len(want) {
		t.Fatalf("offers = %v, want one per source: %v", offers, want)
	}
	for reported, kind := range want {
		if offers[reported] != 1 {
			t.Errorf("%s was offered %d times, want exactly once", reported, offers[reported])
		}
		if kinds[reported] != kind {
			t.Errorf("%s was offered as kind %d, want %d", reported, kinds[reported], kind)
		}
	}
}

// A member whose frontmatter failed to parse still carries its headings,
// and its envelope is not one of them — including when a value inside the
// block contains a run of three hyphens. The block's extent is the
// Frontmatter component's to decide, and closing it at the first "---"
// anywhere leaves the body starting mid-block, where the block's real
// closing delimiter underlines the YAML above it and CommonMark reads the
// envelope as a level-2 heading.
func TestAnUnparsableFrontmatterBlockIsNotIndexedAsAHeading(t *testing.T) {
	// The value carries "---" and the block does not parse: status opens
	// a sequence it never closes.
	src := []byte("---\ntitle: a---b\nstatus: [unclosed\n---\n\n# Real heading\n\nbody\n")

	found, _ := headingsOf(src, nil, nil)

	if len(found) != 1 {
		t.Fatalf("found %d headings, want 1: %+v", len(found), found)
	}
	if found[0].Text != "Real heading" {
		t.Errorf("Text = %q, want %q", found[0].Text, "Real heading")
	}
	if found[0].Level != 1 {
		t.Errorf("Level = %d, want 1", found[0].Level)
	}
	// The envelope sits before the body, so no heading may begin inside
	// it.
	if got := found[0].Span.Offset; got < 40 {
		t.Errorf("the heading begins at offset %d, inside the frontmatter block", got)
	}
}

// The fallback still finds the headings of a document whose frontmatter
// failed to parse without any "---" inside a value, and of one carrying
// no frontmatter at all — the behaviours the fallback exists for.
func TestTheFallbackStillLocatesTheBodyOfAnOrdinaryDocument(t *testing.T) {
	unparsable := []byte("---\nstatus: [unclosed\n---\n\n# Real heading\n\nbody\n")
	if found, _ := headingsOf(unparsable, nil, nil); len(found) != 1 || found[0].Text != "Real heading" {
		t.Errorf("unparsable frontmatter: got %+v, want the one heading %q", found, "Real heading")
	}

	none := []byte("# Only heading\n\nbody\n")
	if found, _ := headingsOf(none, nil, nil); len(found) != 1 || found[0].Text != "Only heading" {
		t.Errorf("no frontmatter: got %+v, want the one heading %q", found, "Only heading")
	}

	// A "---" further down a document with no block of its own opens
	// nothing: the whole source is the body.
	late := []byte("# Heading\n\nbody\n\n---\n\nmore\n")
	if found, _ := headingsOf(late, nil, nil); len(found) != 1 || found[0].Text != "Heading" {
		t.Errorf("a late delimiter: got %+v, want the one heading %q", found, "Heading")
	}
}

// A fence written inside frontmatter YAML is not a record block, and a
// record block in the body of a document with frontmatter is still
// recorded with a span counted from the start of the file.
func TestRecordBlocksAreSearchedFromTheBodyInWholeFileCoordinates(t *testing.T) {
	frontOnly := []byte("---\ntitle: X\ndescription: |\n  ```component\n  name: x\n  ```\n---\n\n# Body\n\ntext\n")
	found, unattached := headingsOf(frontOnly, nil, []string{"component"})
	if len(found) != 1 || len(unattached) != 0 || len(found[0].RecordBlocks) != 0 {
		t.Errorf("frontmatter fence: %d headings, %d unattached, %d attached; want 1, 0, 0",
			len(found), len(unattached), len(found[0].RecordBlocks))
	}

	inBody := []byte("---\ntitle: X\ndescription: |\n  ```component\n  name: x\n  ```\n---\n\n# Body\n\n```component\nname: y\n```\n")
	found, unattached = headingsOf(inBody, nil, []string{"component"})
	if len(unattached) != 0 || len(found) != 1 || len(found[0].RecordBlocks) != 1 {
		t.Fatalf("body fence: %d headings, %d unattached; want the block attached to its heading", len(found), len(unattached))
	}
	record := found[0].RecordBlocks[0]
	if want := uint64(strings.Index(string(inBody), "```component\nname: y")); record.Span.Offset != want {
		t.Errorf("Offset = %d, want %d", record.Span.Offset, want)
	}
	if record.Span.StartLine != 11 {
		t.Errorf("StartLine = %d, want 11 (frontmatter included)", record.Span.StartLine)
	}
}

func TestARecordTheIndexFoundFromTheBodyDecodesDespiteAFenceInFrontmatter(t *testing.T) {
	src := []byte("---\ntitle: X\ndescription: |\n  ```yaml\n  foo: 1\n---\n\n# Body\n\n```component\nname: y\n```\n")
	found, _ := headingsOf(src, nil, []string{"component"})
	if len(found) != 1 || len(found[0].RecordBlocks) != 1 {
		t.Fatalf("found %d headings; want one with one record block", len(found))
	}
	var v struct {
		Name string `yaml:"name"`
	}
	if err := DecodeRecordBlock(src, found[0].RecordBlocks[0], &v); err != nil {
		t.Fatalf("DecodeRecordBlock: %v", err)
	}
	if v.Name != "y" {
		t.Errorf("name = %q, want y", v.Name)
	}
}
