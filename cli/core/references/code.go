package references

import (
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// parser is the markdown parser code regions are located with. It is the
// CommonMark configuration goldmark ships with and nothing else: the only
// question asked of it is where the code is.
var parser = goldmark.New().Parser()

// regions is a set of byte ranges within one document body.
type regions []regionRange

// regionRange is one half-open byte range, [start, stop).
type regionRange struct {
	start int
	stop  int
}

// covers reports whether any region overlaps [start, stop).
//
// Overlap rather than containment, because a reference that begins inside
// a code span and ends outside it — or the reverse — is still text the
// author wrote inside backticks, and reporting half of it is worse than
// reporting none of it.
func (r regions) covers(start, stop int) bool {
	for _, region := range r {
		if start < region.stop && region.start < stop {
			return true
		}
	}
	return false
}

// codeRegions returns the byte ranges of body that are code: fenced code
// blocks, indented code blocks, and inline code spans.
//
// This is the one thing a hand-rolled backtick scan cannot be trusted
// with, and the reason [[adr-015]]'s markdown parser arrives with this
// package. Where an inline code span ends depends on matching the length
// of its opening backtick run, and where a fence ends depends on the
// fence's own length and on its info string; a scanner that gets either
// wrong either masks prose that follows real code, or fails to mask an
// illustration and reports it as a broken reference — the false positive
// the exclusion exists to prevent.
//
// An HTML comment block is deliberately absent from this list. Only code
// is excluded, and the template standard requires a finished document's
// comment blocks to be deleted, so masking them would hide a reference in
// a document that should not be carrying the block at all.
func codeRegions(body []byte) regions {
	var found regions
	document := parser.Parse(text.NewReader(body))

	err := markdown.Walk(document, func(node markdown.Node, entering bool) (markdown.WalkStatus, error) {
		if !entering {
			return markdown.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *markdown.FencedCodeBlock:
			found = append(found, linesOf(typed)...)
			if info := typed.Info; info != nil {
				found = append(found, regionRange{info.Segment.Start, info.Segment.Stop})
			}
		case *markdown.CodeBlock:
			found = append(found, linesOf(typed)...)
		case *markdown.CodeSpan:
			found = append(found, spanOf(typed)...)
		}
		return markdown.WalkContinue, nil
	})
	if err != nil {
		// The walk function above returns no error, so this is
		// unreachable; it is handled rather than discarded so that a
		// future walk that can fail cannot fail silently.
		return found
	}
	return found
}

// linesOf returns the byte ranges of a block node's content lines. For a
// fenced block those are the lines between the fences, which is where an
// illustrative reference is written.
func linesOf(node markdown.Node) regions {
	lines := node.Lines()
	found := make(regions, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		found = append(found, regionRange{line.Start, line.Stop})
	}
	return found
}

// spanOf returns the byte ranges an inline code span covers. A code span
// holds its content as text segments rather than as one range, because
// its content may cross a line break.
func spanOf(node *markdown.CodeSpan) regions {
	var found regions
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		segment, ok := child.(*markdown.Text)
		if !ok {
			continue
		}
		found = append(found, regionRange{segment.Segment.Start, segment.Segment.Stop})
	}
	return found
}
