// Package markdown locates a document body's headings and its fenced
// blocks by info string.
//
// These are two of the three capabilities the Markdown Structure
// component [[module-document-core]] declares: [Headings] builds the
// heading tree with the span of the content beneath each heading, and
// [Fenced] locates blocks by tag. HTML comment blocks are the third and
// stay unbuilt here; they belong to [[frd-spec-structure-conformance]].
//
// Both take recognition from goldmark, already a module dependency —
// References brought it in for the code-region mask described under that
// component. What looks like a heading or a fence inside a fenced block
// or an inline code span is not one, and only a markdown parser knows
// which regions are not to be scanned.
//
// Where a fence ends depends on the fence character, its run
// length and its info string, and a hand-rolled scan that guesses one of
// those wrong extracts a block no markdown reader would agree with.
// goldmark decides where a fence is and where its content lines fall;
// this package only turns those already-decided positions into [span.Span]
// values, through [span.Of], so a fenced block's position is computed the
// same way as every other position in the module.
//
// goldmark's public AST does not distinguish a fence that closed on a
// real closing delimiter from one that simply absorbed the rest of the
// document at EOF — both report the same node shape. This package tells
// the two apart itself: a fence whose content runs to the last byte of
// body, with nothing left for a closing line, is unterminated and yields
// no block, matching the Spans scanner's refusal to guess at a dangling
// delimiter.
package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	mdast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// parser is the markdown parser fenced blocks are located with. It is the
// CommonMark configuration goldmark ships with and nothing else: the only
// question asked of it is where a fence is and where its content lines
// fall.
var parser = goldmark.New().Parser()

// Block is one fenced block whose info string matched a registered tag.
type Block struct {
	// Tag is the registered tag the info string matched, exactly as
	// passed to [Fenced].
	Tag string

	// Content is the bytes between the fences: what [Block.Inner] covers,
	// sliced from body.
	Content []byte

	// Outer covers the whole construct: the opening fence line, the
	// content, and the closing fence line, each in full — a caller
	// eliding or replacing the block needs both fence lines gone with it.
	Outer span.Span

	// Inner covers the content alone, between the two fence lines. A
	// caller with an offset into Content reports it against the document
	// by adding Inner's own start to that offset, since Inner is already
	// expressed in document coordinates rather than block-relative ones.
	Inner span.Span
}

// Fenced returns every fenced block in body whose info string exactly
// matches one of tags, in the order the blocks are written.
//
// The info string is matched exactly as goldmark trims it — surrounding
// whitespace removed, nothing else — against the tag vocabulary tags
// carries. A block with no info string, or one that matches no tag,
// contributes nothing: this is a lookup by tag, not a full inventory of
// every fenced block in the document.
//
// A fenced block that never closes — its content runs to the end of body
// with no closing fence line — is not returned, matching the Spans
// scanner's refusal to guess at a dangling delimiter.
func Fenced(body []byte, tags []string) []Block {
	if len(tags) == 0 {
		return nil
	}
	registered := make(map[string]bool, len(tags))
	for _, tag := range tags {
		registered[tag] = true
	}

	var found []Block
	document := parser.Parse(text.NewReader(body))
	mdast.Walk(document, func(node mdast.Node, entering bool) (mdast.WalkStatus, error) {
		if !entering {
			return mdast.WalkContinue, nil
		}
		fenced, ok := node.(*mdast.FencedCodeBlock)
		if !ok || fenced.Info == nil {
			return mdast.WalkContinue, nil
		}
		tag := string(fenced.Info.Segment.Value(body))
		if !registered[tag] {
			return mdast.WalkContinue, nil
		}
		if block, ok := blockOf(body, fenced, tag); ok {
			found = append(found, block)
		}
		return mdast.WalkContinue, nil
	})
	return found
}

// blockOf computes tag's Block from the fenced code block node goldmark
// already located, reporting false for a fence with no closing line.
func blockOf(body []byte, node *mdast.FencedCodeBlock, tag string) (Block, bool) {
	openLineStart := lineStart(body, node.Info.Segment.Start)

	var contentStart, contentEnd int
	if lines := node.Lines(); lines.Len() > 0 {
		contentStart = lines.At(0).Start
		contentEnd = lines.At(lines.Len() - 1).Stop
	} else {
		// Empty content: if the fence closes at all, the closing line is
		// the one immediately following the opening line.
		contentStart = lineEnd(body, openLineStart)
		contentEnd = contentStart
	}

	if contentEnd >= len(body) {
		// Content ran to the end of body with nothing left for a closing
		// line: the fence never closed.
		return Block{}, false
	}
	closeLineEnd := lineEnd(body, contentEnd)

	return Block{
		Tag:     tag,
		Content: bytes.Clone(body[contentStart:contentEnd]),
		Outer:   span.Of(body, openLineStart, closeLineEnd-openLineStart),
		Inner:   span.Of(body, contentStart, contentEnd-contentStart),
	}, true
}

// lineStart returns the byte offset the line containing offset begins at.
func lineStart(body []byte, offset int) int {
	if idx := bytes.LastIndexByte(body[:offset], '\n'); idx >= 0 {
		return idx + 1
	}
	return 0
}

// lineEnd returns the byte offset immediately after the newline that
// terminates the line beginning at from, or len(body) when that line is
// the document's last and carries no terminating newline.
func lineEnd(body []byte, from int) int {
	if idx := bytes.IndexByte(body[from:], '\n'); idx >= 0 {
		return from + idx + 1
	}
	return len(body)
}
