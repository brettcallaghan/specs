// Tests for fenced-block location, the fenced-block half of Markdown
// Structure. Every fixture is body text: MDS is a lookup within a
// document body, and frontmatter separation is a caller concern this
// package does not take on.
package markdown_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// AC: AC-GHK-001.1
func TestAFencedBlockTaggedWithARegisteredTagIsLocated(t *testing.T) {
	body := []byte("Intro text.\n\n```gherkin\nFeature: Checkout\n```\n\nMore text.\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	block := found[0]

	if block.Tag != "gherkin" {
		t.Errorf("Tag = %q, want %q", block.Tag, "gherkin")
	}
	if got, want := string(block.Content), "Feature: Checkout\n"; got != want {
		t.Errorf("Content = %q, want %q", got, want)
	}
	if got, want := string(span.Slice(body, block.Inner)), "Feature: Checkout\n"; got != want {
		t.Errorf("Inner slices to %q, want %q", got, want)
	}
	if got, want := string(span.Slice(body, block.Outer)), "```gherkin\nFeature: Checkout\n```\n"; got != want {
		t.Errorf("Outer slices to %q, want %q", got, want)
	}
	if block.Inner.StartLine == 0 || block.Outer.StartLine == 0 {
		t.Errorf("a span carries no line: inner=%+v outer=%+v", block.Inner, block.Outer)
	}
}

func TestAFencedBlockWhoseInfoStringMatchesNoRegisteredTagIsNotReturned(t *testing.T) {
	body := []byte("```python\nprint('hi')\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 0 {
		t.Fatalf("found %d blocks, want 0", len(found))
	}
}

func TestAFencedBlockWithNoInfoStringIsNotReturned(t *testing.T) {
	body := []byte("```\ncode\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 0 {
		t.Fatalf("found %d blocks, want 0", len(found))
	}
}

func TestNoRegisteredTagsReturnsNothing(t *testing.T) {
	body := []byte("```gherkin\nFeature: X\n```\n")

	found := markdown.Fenced(body, nil)

	if found != nil {
		t.Fatalf("found %v, want nil", found)
	}
}

func TestMultipleRegisteredTagsEachMatch(t *testing.T) {
	body := []byte("```gherkin\nFeature: X\n```\n\n```strata\nmodel Y {}\n```\n")

	found := markdown.Fenced(body, []string{"gherkin", "strata"})

	if len(found) != 2 {
		t.Fatalf("found %d blocks, want 2", len(found))
	}
	if found[0].Tag != "gherkin" || found[1].Tag != "strata" {
		t.Errorf("tags = %q, %q, want gherkin, strata", found[0].Tag, found[1].Tag)
	}
}

func TestBlocksAreReturnedInDocumentOrder(t *testing.T) {
	body := []byte("```gherkin\nFirst\n```\n\nBetween.\n\n```gherkin\nSecond\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 2 {
		t.Fatalf("found %d blocks, want 2", len(found))
	}
	if got, want := string(found[0].Content), "First\n"; got != want {
		t.Errorf("first block content = %q, want %q", got, want)
	}
	if got, want := string(found[1].Content), "Second\n"; got != want {
		t.Errorf("second block content = %q, want %q", got, want)
	}
	if found[0].Outer.Offset >= found[1].Outer.Offset {
		t.Errorf("blocks not in document order: %+v then %+v", found[0].Outer, found[1].Outer)
	}
}

// An unterminated fence yields no block, matching the Spans scanner's
// refusal to guess at a dangling delimiter.
func TestAnUnterminatedFenceYieldsNoBlock(t *testing.T) {
	cases := map[string]string{
		"content, never closed":       "```gherkin\nFeature: X\n",
		"opening line only, with EOL": "```gherkin\n",
		"opening line only, no EOL":   "```gherkin",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			found := markdown.Fenced([]byte(src), []string{"gherkin"})
			if len(found) != 0 {
				t.Fatalf("found %d blocks for %q, want 0", len(found), src)
			}
		})
	}
}

func TestAnEmptyFencedBlockThatClosesIsLocated(t *testing.T) {
	body := []byte("```gherkin\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	block := found[0]
	if len(block.Content) != 0 {
		t.Errorf("Content = %q, want empty", block.Content)
	}
	if block.Inner.Length != 0 {
		t.Errorf("Inner.Length = %d, want 0", block.Inner.Length)
	}
	if got, want := string(span.Slice(body, block.Outer)), "```gherkin\n```\n"; got != want {
		t.Errorf("Outer slices to %q, want %q", got, want)
	}
}

func TestAFenceWithNoTrailingDocumentNewlineStillClosesOuterAtEOF(t *testing.T) {
	body := []byte("```gherkin\nFeature: X\n```")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	if got, want := string(span.Slice(body, found[0].Outer)), "```gherkin\nFeature: X\n```"; got != want {
		t.Errorf("Outer slices to %q, want %q", got, want)
	}
}

func TestAFalseClosingRunShorterThanTheOpeningIsContent(t *testing.T) {
	// goldmark's own rule: a closing run must be at least as long as the
	// opening one. A shorter run is content, not a closer.
	body := []byte("```gherkin\nline one\n``\nline three\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	if got, want := string(found[0].Content), "line one\n``\nline three\n"; got != want {
		t.Errorf("Content = %q, want %q", got, want)
	}
}

func TestAnIndentedFencedBlockIsStillLocated(t *testing.T) {
	body := []byte("- item\n\n  ```gherkin\n  Feature: X\n  ```\n")

	found := markdown.Fenced(body, []string{"gherkin"})

	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	if got, want := string(span.Slice(body, found[0].Outer)), "  ```gherkin\n  Feature: X\n  ```\n"; got != want {
		t.Errorf("Outer slices to %q, want %q", got, want)
	}
}

// The Behavior bullet under MDS requires that an offset within a block's
// content be reportable as an offset within the containing document by
// adding the content's start (AC-GHK-004.2, AC-GHK-006.1 — verified where
// those items consume this capability, per the item's Verification
// timing note). Inner.Offset already is that content start, expressed in
// document coordinates, so a caller adds it directly without this
// package's help.
func TestInnerOffsetIsExpressedInDocumentCoordinates(t *testing.T) {
	body := []byte("Intro.\n\n```gherkin\nFeature: X\n```\n")

	found := markdown.Fenced(body, []string{"gherkin"})
	if len(found) != 1 {
		t.Fatalf("found %d blocks, want 1", len(found))
	}
	block := found[0]

	// An offset of 8 within Content ("Feature: X\n") lands on "X". Adding
	// Inner's own document offset must land on the same byte in body.
	const withinContent = 9 // "Feature: " is 9 bytes; content[9] == 'X'
	if block.Content[withinContent] != 'X' {
		t.Fatalf("fixture drifted: content[%d] = %q, want 'X'", withinContent, block.Content[withinContent])
	}
	absolute := int(block.Inner.Offset) + withinContent
	if body[absolute] != 'X' {
		t.Errorf("body[%d] = %q, want 'X'", absolute, body[absolute])
	}
}
