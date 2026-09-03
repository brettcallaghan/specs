// Package span_test covers the Spans component.
//
// No test here carries an `AC:` citation, and that is deliberate rather
// than an omission: no acceptance criterion in
// specs/features/spec-index/frd-spec-index.md names Spans. The component
// is proven through the components that consume it — Frontmatter, the
// Schema Registry and Corpus — and its own verification obligation comes
// from the Change Guidance in specs/modules/document-core/module-document-core.md,
// which requires property tests over the delimiter scanner. Those live in
// span_property_test.go.
package span_test

import (
	"bytes"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// The Spans component reports positions in the generated Span type
// rather than a second, parallel one.
var _ specmodel.Span = span.Span{}

func TestFindLocatesAFrontmatterBlock(t *testing.T) {
	src := []byte("---\ntitle: Spans\nslug: component-spans\n---\n\n# Spans\n")

	region, ok := span.Find(src, "---", "---", 0)
	if !ok {
		t.Fatal("Find reported no region")
	}

	if got, want := string(span.Slice(src, region.Inner)), "\ntitle: Spans\nslug: component-spans\n"; got != want {
		t.Errorf("Inner = %q, want %q", got, want)
	}
	if got, want := string(span.Slice(src, region.Outer)), "---\ntitle: Spans\nslug: component-spans\n---"; got != want {
		t.Errorf("Outer = %q, want %q", got, want)
	}
	if region.Outer.Offset != 0 || region.Outer.StartLine != 1 || region.Outer.StartColumn != 1 {
		t.Errorf("Outer starts at offset %d, %d:%d; want 0, 1:1",
			region.Outer.Offset, region.Outer.StartLine, region.Outer.StartColumn)
	}
	if region.Outer.EndLine != 4 || region.Outer.EndColumn != 4 {
		t.Errorf("Outer ends at %d:%d, want 4:4", region.Outer.EndLine, region.Outer.EndColumn)
	}
}

func TestFindLocatesACommentBlockWithDistinctDelimiters(t *testing.T) {
	src := []byte("before <!-- scaffolding --> after")

	region, ok := span.Find(src, "<!--", "-->", 0)
	if !ok {
		t.Fatal("Find reported no region")
	}

	if got, want := string(span.Slice(src, region.Inner)), " scaffolding "; got != want {
		t.Errorf("Inner = %q, want %q", got, want)
	}
	if got, want := string(span.Slice(src, region.Outer)), "<!-- scaffolding -->"; got != want {
		t.Errorf("Outer = %q, want %q", got, want)
	}
}

func TestFindReportsNoRegionForAnUnterminatedOpening(t *testing.T) {
	src := []byte("---\ntitle: never closed\n")

	if _, ok := span.Find(src, "---", "---", 0); ok {
		t.Error("Find accepted an unterminated region; it must report none")
	}
}

func TestFindReportsNoRegionWhenTheOpeningIsAbsent(t *testing.T) {
	src := []byte("# Spans\n\nNo frontmatter here.\n")

	if _, ok := span.Find(src, "---", "---", 0); ok {
		t.Error("Find located a region in a document with no opening delimiter")
	}
}

func TestFindReportsAnEmptyRegionForAdjacentDelimiters(t *testing.T) {
	src := []byte("------\n")

	region, ok := span.Find(src, "---", "---", 0)
	if !ok {
		t.Fatal("Find reported no region")
	}
	if region.Inner.Length != 0 {
		t.Errorf("Inner.Length = %d, want 0", region.Inner.Length)
	}
	if region.Inner.Offset != 3 {
		t.Errorf("Inner.Offset = %d, want 3", region.Inner.Offset)
	}
	if region.Outer.Length != 6 {
		t.Errorf("Outer.Length = %d, want 6", region.Outer.Length)
	}
}

func TestFindDoesNotNest(t *testing.T) {
	src := []byte("<!-- outer <!-- inner --> tail -->")

	region, ok := span.Find(src, "<!--", "-->", 0)
	if !ok {
		t.Fatal("Find reported no region")
	}

	if got, want := string(span.Slice(src, region.Inner)), " outer <!-- inner "; got != want {
		t.Errorf("Inner = %q, want %q — the first closing delimiter ends the region", got, want)
	}
}

func TestFindResumesAfterAPrecedingRegion(t *testing.T) {
	src := []byte("<!-- one --> middle <!-- two -->")

	first, ok := span.Find(src, "<!--", "-->", 0)
	if !ok {
		t.Fatal("Find reported no first region")
	}

	second, ok := span.Find(src, "<!--", "-->", int(first.Outer.Offset+first.Outer.Length))
	if !ok {
		t.Fatal("Find reported no second region")
	}
	if got, want := string(span.Slice(src, second.Inner)), " two "; got != want {
		t.Errorf("second Inner = %q, want %q", got, want)
	}
}

func TestFindRejectsAnEmptyDelimiter(t *testing.T) {
	src := []byte("---\nbody\n---\n")

	if _, ok := span.Find(src, "", "---", 0); ok {
		t.Error("Find accepted an empty opening delimiter")
	}
	if _, ok := span.Find(src, "---", "", 0); ok {
		t.Error("Find accepted an empty closing delimiter")
	}
}

func TestFindTreatsANegativeStartAsZeroAndAStartPastTheEndAsNoMatch(t *testing.T) {
	src := []byte("<!-- x -->")

	fromZero, ok := span.Find(src, "<!--", "-->", 0)
	if !ok {
		t.Fatal("Find reported no region from 0")
	}
	fromNegative, ok := span.Find(src, "<!--", "-->", -100)
	if !ok {
		t.Fatal("Find reported no region from a negative start")
	}
	if fromNegative != fromZero {
		t.Errorf("a negative start gave %+v, want %+v", fromNegative, fromZero)
	}

	if _, ok := span.Find(src, "<!--", "-->", len(src)+1); ok {
		t.Error("Find located a region starting past the end of the document")
	}
}

// A column counts codepoints, so a value that follows a multibyte
// character is reported at the column a reader would count to — not at
// its byte distance from the line start.
func TestColumnsCountCodepointsRatherThanBytes(t *testing.T) {
	src := []byte("héllo — <!-- x -->\n")

	region, ok := span.Find(src, "<!--", "-->", 0)
	if !ok {
		t.Fatal("Find reported no region")
	}

	if region.Outer.StartLine != 1 {
		t.Errorf("StartLine = %d, want 1", region.Outer.StartLine)
	}
	// "héllo — " is 8 codepoints, so the region opens at column 9 even
	// though its byte offset is 11.
	if region.Outer.StartColumn != 9 {
		t.Errorf("StartColumn = %d, want 9", region.Outer.StartColumn)
	}
	if region.Outer.Offset != 11 {
		t.Errorf("Offset = %d, want 11", region.Outer.Offset)
	}
}

func TestOfClampsToTheDocument(t *testing.T) {
	src := []byte("abc")

	past := span.Of(src, 10, 5)
	if past.Offset != 3 || past.Length != 0 {
		t.Errorf("Of past the end = (%d, %d), want (3, 0)", past.Offset, past.Length)
	}

	overlong := span.Of(src, 1, 99)
	if overlong.Offset != 1 || overlong.Length != 2 {
		t.Errorf("Of with an overlong length = (%d, %d), want (1, 2)", overlong.Offset, overlong.Length)
	}

	negative := span.Of(src, -4, -4)
	if negative.Offset != 0 || negative.Length != 0 {
		t.Errorf("Of with negative arguments = (%d, %d), want (0, 0)", negative.Offset, negative.Length)
	}
}

func TestSliceReturnsTheBytesTheSpanCovers(t *testing.T) {
	src := []byte("alpha beta")

	if got := span.Slice(src, span.Of(src, 6, 4)); !bytes.Equal(got, []byte("beta")) {
		t.Errorf("Slice = %q, want %q", got, "beta")
	}
	if got := span.Slice(src, span.Span{Offset: 99, Length: 1}); got != nil {
		t.Errorf("Slice of an out-of-range span = %q, want nil", got)
	}
}
