// Package span owns the position type every component reports through,
// and the one delimiter scanner every parser locates regions with.
//
// Two parsers that locate regions differently will disagree about where
// a finding is, and a caller that receives a bare line number cannot
// point at a value inside a line. Both problems are solved by having a
// single place where positions are computed: this one. A new parsing
// capability is built on [Find] rather than scanning independently.
//
// The Span type itself is generated from schemas/index.strata, because
// spans cross the JSON boundary inside the emitted index and — from the
// FIND delta onward — inside findings. This package aliases it rather
// than declaring a second one, so there is exactly one definition of a
// position in the system and no conversion between an in-memory span
// and a serialized one.
package span

import (
	"bytes"
	"unicode/utf8"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Span is a contiguous region of a document.
//
// Byte offset and byte length are the authoritative extent. Line and
// column are derived from the offset and carried for readability: lines
// and columns are one-based, and a column counts Unicode codepoints from
// the start of its line, so an em dash advances it by one rather than by
// its three UTF-8 bytes.
type Span = specmodel.Span

// Region is a delimited region of a document, located by [Find].
//
// Both extents are reported because both are needed: a parser reads
// Inner, and a caller that must replace or elide the whole construct —
// frontmatter block, comment block, fenced block — needs Outer.
type Region struct {
	// Outer covers the opening delimiter, the content, and the closing
	// delimiter.
	Outer Span
	// Inner covers the content between the delimiters. It is empty, not
	// absent, when the delimiters are adjacent.
	Inner Span
}

// Find locates the first region of src at or after the byte offset from
// that opens with opening and closes with closing.
//
// The search for the closing delimiter begins immediately after the
// opening one, so the two may be identical — "---" and "---" for
// frontmatter — as readily as distinct — "<!--" and "-->" for a comment
// block.
//
// The scanner does not nest. A second opening delimiter inside the
// region is content, and the first closing delimiter after the opening
// one ends the region. This matches the constructs it exists to locate:
// neither a frontmatter block nor an HTML comment nests, and a scanner
// that pretended otherwise would report a different region than the
// parsers reading the same file.
//
// Find reports false, rather than a partial region, when either
// delimiter is empty, when the opening delimiter does not occur at or
// after from, or when no closing delimiter follows it — an unterminated
// region is not a region. A caller that must report the unterminated
// opening delimiter as a defect locates it with its own [bytes.Index]
// and [Of]; conflating "no region" with "a region running to end of
// file" would silently swallow the malformed input.
//
// A negative from is treated as 0.
func Find(src []byte, opening, closing string, from int) (Region, bool) {
	if opening == "" || closing == "" {
		return Region{}, false
	}
	if from < 0 {
		from = 0
	}
	if from > len(src) {
		return Region{}, false
	}

	openAt := bytes.Index(src[from:], []byte(opening))
	if openAt < 0 {
		return Region{}, false
	}
	openAt += from

	contentAt := openAt + len(opening)
	closeAt := bytes.Index(src[contentAt:], []byte(closing))
	if closeAt < 0 {
		return Region{}, false
	}
	closeAt += contentAt

	return Region{
		Outer: Of(src, openAt, closeAt+len(closing)-openAt),
		Inner: Of(src, contentAt, closeAt-contentAt),
	}, true
}

// Of returns the Span covering length bytes of src from offset, with its
// line and column derived from src.
//
// Offset and length are clamped to src, so a caller cannot mint a span
// that does not address the document it came from.
func Of(src []byte, offset, length int) Span {
	if offset < 0 {
		offset = 0
	}
	if offset > len(src) {
		offset = len(src)
	}
	if length < 0 {
		length = 0
	}
	if offset+length > len(src) {
		length = len(src) - offset
	}

	startLine, startColumn := lineColumn(src, offset)
	endLine, endColumn := lineColumn(src, offset+length)

	return Span{
		Offset:      uint64(offset),
		Length:      uint64(length),
		StartLine:   uint64(startLine),
		StartColumn: uint64(startColumn),
		EndLine:     uint64(endLine),
		EndColumn:   uint64(endColumn),
	}
}

// Slice returns the bytes of src that s covers.
//
// The result aliases src rather than copying it; a caller that retains
// it beyond the life of src copies it itself.
func Slice(src []byte, s Span) []byte {
	if s.Offset > uint64(len(src)) {
		return nil
	}
	end := s.Offset + s.Length
	if end > uint64(len(src)) {
		end = uint64(len(src))
	}
	return src[s.Offset:end]
}

// lineColumn returns the one-based line and codepoint column of a byte
// offset within src.
func lineColumn(src []byte, offset int) (line, column int) {
	line = 1 + bytes.Count(src[:offset], []byte("\n"))
	lineStart := bytes.LastIndexByte(src[:offset], '\n') + 1
	column = 1 + utf8.RuneCount(src[lineStart:offset])
	return line, column
}
