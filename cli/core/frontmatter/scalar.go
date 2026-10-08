package frontmatter

import (
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// TopLevelScalar reads the value of the top-level property name from
// src's frontmatter block by line, without parsing the block as YAML. It
// exists for a block [Parse] rejected: one malformed value elsewhere in
// the block must not make a property written correctly on its own line
// unreadable.
//
// A value is found only where all of these hold, and none is guessed
// otherwise:
//
//   - the block is delimited — [Locate] finds it. An unterminated block
//     has no extent, so nothing inside it is read;
//   - a line begins with the key itself, unindented, followed by a colon.
//     An indented key belongs to some other mapping;
//   - its value is a single-line scalar: plain, single-quoted or
//     double-quoted, a trailing comment allowed after a plain value or a
//     closing quote;
//   - the scalar is not empty.
//
// The first line declaring the key decides: a duplicate key is a YAML
// error in its own right and never makes a second value legible.
func TopLevelScalar(src []byte, name string) (string, bool) {
	value, _, found := TopLevelScalarAt(src, name)
	return value, found
}

// TopLevelScalarAt is [TopLevelScalar] that also returns where the value
// sits in src: the value's source text on its line, from its first
// non-blank byte to its last, which is what [Property.Position] covers
// for the same line in a block that parsed.
func TopLevelScalarAt(src []byte, name string) (string, specmodel.Span, bool) {
	region, found := Locate(src)
	if !found {
		return "", specmodel.Span{}, false
	}
	prefix := name + ":"
	offset := int(region.Inner.Offset)
	for _, line := range strings.SplitAfter(string(span.Slice(src, region.Inner)), "\n") {
		start := offset
		offset += len(line)
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		rest, ok := strings.CutPrefix(line, prefix)
		if !ok {
			continue
		}
		if rest != "" && rest[0] != ' ' && rest[0] != '\t' {
			// "slugs:" or "slug_x:" — a different key sharing the prefix.
			continue
		}
		text := strings.TrimSpace(rest)
		value, ok := singleLineScalar(text)
		if !ok || value == "" {
			return "", specmodel.Span{}, false
		}
		valueStart := start + len(prefix) + (len(rest) - len(strings.TrimLeft(rest, " \t")))
		return value, span.Of(src, valueStart, len(text)), true
	}
	return "", specmodel.Span{}, false
}

// singleLineScalar reads text as one YAML scalar written on one line.
func singleLineScalar(text string) (string, bool) {
	if text == "" {
		return "", false
	}
	switch text[0] {
	case '"':
		end := closingDoubleQuote(text)
		if end < 0 || !onlyComment(text[end+1:]) {
			return "", false
		}
		// Unescaped by the YAML library rather than strconv.Unquote: YAML's
		// escape set is not Go's (`\/`, `\ `, `\e`, `\N` …), and the
		// parser that decides whether the block parsed is the one that
		// should decide what a quoted value says.
		var value string
		if err := yaml.Unmarshal([]byte(text[:end+1]), &value); err != nil {
			return "", false
		}
		return value, true
	case '\'':
		var b strings.Builder
		for i := 1; i < len(text); i++ {
			if text[i] != '\'' {
				b.WriteByte(text[i])
				continue
			}
			if i+1 < len(text) && text[i+1] == '\'' {
				b.WriteByte('\'')
				i++
				continue
			}
			if !onlyComment(text[i+1:]) {
				return "", false
			}
			return b.String(), true
		}
		return "", false
	case '[', '{', '&', '*', '!', '|', '>', '%', '@', '`', '#':
		// A collection, an anchor, alias, tag or block scalar, a directive,
		// a reserved indicator or a comment: none is a plain scalar.
		return "", false
	}
	if i := strings.Index(text, " #"); i >= 0 {
		text = strings.TrimSpace(text[:i])
	}
	if i := strings.Index(text, "\t#"); i >= 0 {
		text = strings.TrimSpace(text[:i])
	}
	if strings.Contains(text, ": ") || strings.HasSuffix(text, ":") {
		// A nested mapping on one line is no scalar.
		return "", false
	}
	return text, true
}

// closingDoubleQuote returns the index of the quote closing the
// double-quoted scalar text opens, or -1 where it is never closed.
func closingDoubleQuote(text string) int {
	for i := 1; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}
	return -1
}

// onlyComment reports whether rest, what follows a closing quote, is
// nothing or a comment.
func onlyComment(rest string) bool {
	rest = strings.TrimSpace(rest)
	return rest == "" || strings.HasPrefix(rest, "#")
}
