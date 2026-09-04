package markdown

import (
	"bytes"
	"strings"
	"unicode"

	mdast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// Heading is one heading of a document body, with the extent of the
// content beneath it.
//
// The section span is what makes a place inside a document addressable
// without the document's prose being copied anywhere: a caller that wants
// the text under a heading slices it back out of body through Section.
type Heading struct {
	// Text is the heading's text as authored, with the markers that open
	// and close the line removed and surrounding whitespace trimmed —
	// "Last" for "## Last ##". Inline markup is left as written, so a
	// caller comparing headings by name uses [Identifier] rather than
	// this.
	Text string

	// Level is 1 for a top-level heading through 6 for the deepest, for a
	// heading written in either markdown form.
	Level int

	// Position covers the heading line itself, in full: its markers, its
	// text and its terminating newline. For a heading underlined rather
	// than prefixed, it covers the underline line too, so that Position
	// ends exactly where Section begins in both forms.
	Position span.Span

	// Section covers the content beneath the heading: from the end of the
	// heading line to the start of the next heading of the same or a
	// higher level, and to the end of the document where there is none.
	// A heading that stands over nothing carries an empty span rather
	// than an absent one.
	Section span.Span
}

// Headings returns every heading of body in document order, each with its
// level, its text as authored, and the span of the content beneath it.
//
// Heading recognition comes from goldmark for the same reason fence
// recognition does: what looks like a heading inside a fenced block or an
// inline code span is not one, and only a markdown parser knows which
// regions are not to be scanned. This function turns the positions
// goldmark has already decided into [span.Span] values through
// [span.Of], so a heading's position is comparable to a frontmatter,
// reference or fenced-block position.
//
// A body carrying no heading yields an empty result rather than an error:
// a document without headings is a document, not a defect.
func Headings(body []byte) []Heading {
	located := locate(body)
	if len(located) == 0 {
		return nil
	}

	headings := make([]Heading, 0, len(located))
	for i, current := range located {
		// The section ends where the next heading that is not nested
		// beneath this one begins. A deeper heading is part of this
		// section, so it is skipped; the first heading at this level or
		// above closes it, and the last such heading's section runs to
		// the end of the document.
		sectionEnd := len(body)
		for _, next := range located[i+1:] {
			if next.heading.Level <= current.heading.Level {
				sectionEnd = next.lineStart
				break
			}
		}

		heading := current.heading
		heading.Section = span.Of(body, current.sectionStart, sectionEnd-current.sectionStart)
		headings = append(headings, heading)
	}
	return headings
}

// placed is one heading with the byte offsets [Headings] needs to compute
// sections, which the Span it reports has already converted away from.
type placed struct {
	heading Heading
	// lineStart is where the heading's own line begins: where the section
	// of every shallower heading above it ends.
	lineStart int
	// sectionStart is where the content beneath the heading begins.
	sectionStart int
}

// locate walks body for headings, recording each one's extent without yet
// knowing where its section ends — that depends on the headings after it.
func locate(body []byte) []placed {
	var located []placed
	document := parser.Parse(text.NewReader(body))
	mdast.Walk(document, func(node mdast.Node, entering bool) (mdast.WalkStatus, error) {
		if !entering {
			return mdast.WalkContinue, nil
		}
		heading, ok := node.(*mdast.Heading)
		if !ok {
			return mdast.WalkContinue, nil
		}
		lines := heading.Lines()
		if lines.Len() == 0 {
			// A heading with no text segment — "##" alone on a line —
			// has no authored text to record and no line to point at.
			return mdast.WalkContinue, nil
		}
		first, last := lines.At(0), lines.At(lines.Len()-1)

		start := lineStart(body, first.Start)
		end := lineEnd(body, last.Start)
		if !prefixed(body, start) {
			// An underlined heading: the line of dashes or equals signs
			// below the text is part of the heading rather than the
			// first line of its section.
			end = lineEnd(body, end)
		}

		located = append(located, placed{
			heading: Heading{
				Text:     string(bytes.TrimSpace(body[first.Start:last.Stop])),
				Level:    heading.Level,
				Position: span.Of(body, start, end-start),
			},
			lineStart:    start,
			sectionStart: end,
		})
		return mdast.WalkContinue, nil
	})
	return located
}

// prefixed reports whether the heading line beginning at start is written
// with leading markers rather than underlined beneath its text.
func prefixed(body []byte, start int) bool {
	for i := start; i < len(body); i++ {
		if body[i] != ' ' && body[i] != '\t' {
			return body[i] == '#'
		}
	}
	return false
}

// Identifier returns the heading identifier of text: its rendered inline
// content lowercased, with every character that is not a letter, digit,
// hyphen, underscore or space removed, and each remaining space replaced
// by a hyphen.
//
// The derivation is a pure function of text and derives no positional
// suffix for a repeated heading, so two documents whose headings read the
// same produce the same identifier and a caller resolving an anchor
// matches against this one derivation rather than a second one.
//
// Inline content is rendered rather than filtered: the identifier of
// "The `Profile` interface" is "the-profile-interface", the backticks
// being markup rather than characters of the heading's name.
func Identifier(text string) string {
	var identifier strings.Builder
	for _, r := range strings.ToLower(rendered(text)) {
		switch {
		case r == ' ':
			identifier.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			identifier.WriteRune(r)
		}
	}
	return identifier.String()
}

// rendered returns source's inline content as a reader sees it, with the
// markup that produced it removed.
func rendered(source string) string {
	body := []byte(source)
	var out []byte
	mdast.Walk(parser.Parse(text.NewReader(body)), func(node mdast.Node, entering bool) (mdast.WalkStatus, error) {
		if !entering {
			return mdast.WalkContinue, nil
		}
		switch value := node.(type) {
		case *mdast.Text:
			out = append(out, value.Segment.Value(body)...)
			if value.SoftLineBreak() || value.HardLineBreak() {
				// A break renders as a space; without one the last word
				// of a line and the first of the next would join.
				out = append(out, ' ')
			}
		case *mdast.String:
			out = append(out, value.Value...)
		}
		return mdast.WalkContinue, nil
	})
	return string(out)
}
