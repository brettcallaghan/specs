package mdlinks

import (
	"bytes"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// openingDelimiter and closingDelimiter are the wikilink notation a body
// reference's Position sits between. references.Extract already located
// them by scanning; Substitute trusts that computed Position rather than
// rescanning, so these constants exist only to widen a span by the
// delimiters' own fixed width, never to search src for anything.
const (
	openingDelimiter = "[["
	closingDelimiter = "]]"
)

// Substitute rewrites src, replacing each body-origin reference in refs
// that settled a target with a markdown link, and copying every other
// byte through unchanged.
//
// refs and targets must be the slices [Resolve] was given and returned:
// targets[i] is read for refs[i], by index. A reference whose Origin is a
// frontmatter property, not a body wikilink, is left alone — its Position
// covers the property's YAML value rather than a `[[...]]` construct, and
// there is no such construct here to replace. refs is trusted to already
// exclude code regions (AC-SPECMD-001.4): that exclusion is a property of
// the References component's output, not a rule reimplemented here.
//
// count reports how many references were actually replaced
// (AC-SPECMD-003.1). A count of zero is always accompanied by out
// byte-identical to src (AC-SPECMD-001.6, AC-SPECMD-002.7).
func Substitute(src []byte, refs []specmodel.Reference, targets []*string) (out []byte, count int) {
	var buf bytes.Buffer
	buf.Grow(len(src))
	cursor := uint64(0)

	for i, ref := range refs {
		if _, isBody := ref.Origin.AsBody(); !isBody {
			continue
		}

		construct := constructSpan(ref.Position)
		buf.Write(src[cursor:construct.Offset])

		if target := targets[i]; target != nil {
			buf.WriteByte('[')
			buf.WriteString(ref.Text)
			buf.WriteString("](")
			buf.WriteString(*target)
			buf.WriteByte(')')
			count++
		} else {
			buf.Write(span.Slice(src, construct))
		}
		cursor = construct.Offset + construct.Length
	}
	buf.Write(src[cursor:])
	return buf.Bytes(), count
}

// constructSpan derives the extent of the `[[...]]` construct a body
// reference's Position sits inside: Position covers the text between the
// brackets (per the delta's Design), so the construct is that span
// widened by the delimiters' fixed width on each side.
func constructSpan(position span.Span) span.Span {
	position.Offset -= uint64(len(openingDelimiter))
	position.Length += uint64(len(openingDelimiter) + len(closingDelimiter))
	return position
}
