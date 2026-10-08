package references

import (
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
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
// blocks, indented code blocks, and inline code spans — except a fenced
// block's content lines where the fence's info string names a registered
// Gherkin tag ([gherkin.FenceTags]), which is Gherkin rather than an
// illustration and is left unmasked so a wikilink written inside it is
// extracted like any other.
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
// A fence's info string is masked whether or not its tag is registered:
// it is notation naming the block, never a place Gherkin content is
// written, so nothing about narrowing the content mask bears on it.
//
// An HTML comment block is deliberately absent from this list. Only code
// is excluded, and the template standard requires a finished document's
// comment blocks to be deleted, so masking them would hide a reference in
// a document that should not be carrying the block at all.
func codeRegions(body []byte) regions {
	code, _ := scanRegions(body)
	return code
}

// gherkinFenceRegions returns the byte ranges of body that are the content
// lines of a fenced block whose info string names a registered Gherkin tag
// — the lines [codeRegions] deliberately leaves unmasked.
func gherkinFenceRegions(body []byte) regions {
	_, fenced := scanRegions(body)
	return fenced
}

// scanRegions walks body once and returns both what is code and what is a
// Gherkin fence's content, so the two can never disagree about where a
// fence begins or ends.
func scanRegions(body []byte) (code, gherkinFenced regions) {
	gherkinFence := tagSet(gherkin.FenceTags())

	document := parser.Parse(text.NewReader(body))

	err := markdown.Walk(document, func(node markdown.Node, entering bool) (markdown.WalkStatus, error) {
		if !entering {
			return markdown.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *markdown.FencedCodeBlock:
			var tag string
			if info := typed.Info; info != nil {
				code = append(code, regionRange{info.Segment.Start, info.Segment.Stop})
				tag = string(info.Segment.Value(body))
			}
			if gherkinFence[tag] {
				gherkinFenced = append(gherkinFenced, linesOf(typed)...)
			} else {
				code = append(code, linesOf(typed)...)
			}
		case *markdown.CodeBlock:
			code = append(code, linesOf(typed)...)
		case *markdown.CodeSpan:
			code = append(code, spanOf(typed)...)
		}
		return markdown.WalkContinue, nil
	})
	if err != nil {
		// The walk function above returns no error, so this is
		// unreachable; it is handled rather than discarded so that a
		// future walk that can fail cannot fail silently.
		return code, gherkinFenced
	}
	return code, gherkinFenced
}

// OutsideGherkinFences returns refs without every body reference written
// inside a fenced block tagged with a registered Gherkin tag, preserving
// order. Frontmatter references pass through.
//
// [Extract] keeps such a reference, because a wikilink in Gherkin is a
// reference the index and the checks must see. A caller rendering the
// document is asking a different question: whether the text is a citation
// a reader of the rendered copy follows. A renderer shows a fenced block
// literally, so for that caller a Gherkin fence is code like any other,
// and this filter is what makes it so. options must describe the same
// document refs were extracted from.
func OutsideGherkinFences(options Options, refs []specmodel.Reference) []specmodel.Reference {
	start := bodyStart(options)
	if start >= len(options.Source) {
		return refs
	}
	fenced := gherkinFenceRegions(options.Source[start:])
	if len(fenced) == 0 {
		return refs
	}

	kept := make([]specmodel.Reference, 0, len(refs))
	for _, reference := range refs {
		if _, body := reference.Origin.AsBody(); body {
			from := int(reference.Span.Offset) - start
			if fenced.covers(from, from+int(reference.Span.Length)) {
				continue
			}
		}
		kept = append(kept, reference)
	}
	return kept
}

// tagSet builds a lookup set from a fence tag vocabulary, so membership is
// a map access per fenced block rather than a scan of tags for each one.
func tagSet(tags []string) map[string]bool {
	set := make(map[string]bool, len(tags))
	for _, tag := range tags {
		set[tag] = true
	}
	return set
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
