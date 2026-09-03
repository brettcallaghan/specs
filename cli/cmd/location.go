package cmd

import (
	"regexp"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// tomlHeader matches a top-level table header line: `[name]`.
//
// Only a bare identifier is recognised. Array-of-tables (`[[name]]`), and
// a dotted or quoted table name, are outside the pinned v1 grammar
// specs-configuration-contract.md fixes ("That is the whole of it in
// v1."); a line using one of those forms simply produces no location
// below, and a finding about it falls back to an enclosing or whole-file
// span rather than a precise one.
var tomlHeader = regexp.MustCompile(`^\[([A-Za-z0-9_-]+)\]\s*(#.*)?$`)

// tomlKeyValue matches a flat `key = value` line within a table.
//
// The trailing `\s*$` means the captured value already has trailing
// whitespace removed; a `#` comment on the same line as a value is not
// stripped and becomes part of it; that is out of the pinned v1 grammar's
// tested shape (a bare scalar, quoted string, or integer, alone on its
// line) and is the scanner's one known gap.
var tomlKeyValue = regexp.MustCompile(`^([A-Za-z0-9_-]+)\s*=\s*(.*?)\s*$`)

// keyLocation is where one `key = value` line's key and value tokens sit
// in the source.
type keyLocation struct {
	key   specmodel.Span
	value specmodel.Span
}

// locations is the result of scanning a specs.toml source: every
// `key = value` pair's location, keyed the same way a JSON Pointer names
// it ("/workspace/name"), and every table header's own location, keyed
// by its pointer ("/workspace") — the same key set a JSON Schema
// violation's instanceLocation resolves against.
type locations struct {
	keys   map[string]keyLocation
	tables map[string]specmodel.Span
}

// locateConfig performs a line-oriented scan of a specs.toml source. It
// is deliberately shallow rather than a general TOML parser: one level of
// `[table]` headers, each holding flat `key = value` lines, is the whole
// of the pinned v1 grammar (specs-configuration-contract.md), and nothing
// else needs to be locatable precisely. BurntSushi/toml still decodes a
// file using a dotted key, an inline table, an array of tables, or a
// multi-line value as far as TOML's own grammar allows; this scanner
// reports no location for such a construct, and a finding about it falls
// back to its enclosing table or the whole file.
func locateConfig(src []byte) locations {
	loc := locations{keys: map[string]keyLocation{}, tables: map[string]specmodel.Span{}}

	table := ""
	offset := 0
	for _, line := range strings.Split(string(src), "\n") {
		lineStart := offset
		offset += len(line) + 1 // +1 for the newline Split consumed

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if m := tomlHeader.FindStringSubmatch(trimmed); m != nil {
			table = m[1]
			nameAt := lineStart + indent + strings.IndexByte(line[indent:], '[') + 1
			loc.tables["/"+table] = span.Of(src, nameAt, len(table))
			continue
		}

		eq := strings.IndexByte(line, '=')
		m := tomlKeyValue.FindStringSubmatch(trimmed)
		if m == nil || eq < 0 {
			continue
		}
		key, value := m[1], m[2]

		keyAt := lineStart + indent
		pointer := "/" + key
		if table != "" {
			pointer = "/" + table + "/" + key
		}

		valueStart := eq + 1
		for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
			valueStart++
		}

		loc.keys[pointer] = keyLocation{
			key:   span.Of(src, keyAt, len(key)),
			value: span.Of(src, lineStart+valueStart, len(value)),
		}
	}

	return loc
}

// valueSpan returns the span a finding about pointer's value should
// carry: the value's own span where the scanner located one, its
// enclosing table's span next, and finally the whole file — the same
// widening-fallback FindingsFor's caller-facing narrowing never needs,
// but a configuration defect at an unlocatable construct does.
func (l locations) valueSpan(pointer string, src []byte) specmodel.Span {
	if loc, ok := l.keys[pointer]; ok {
		return loc.value
	}
	if sp, ok := l.tables[pointer]; ok {
		return sp
	}
	if i := strings.LastIndex(pointer, "/"); i > 0 {
		return l.valueSpan(pointer[:i], src)
	}
	return span.Of(src, 0, len(src))
}

// keySpan is [locations.valueSpan]'s counterpart for a defect about a
// key's mere presence — an undeclared key — rather than about the value
// it carries.
func (l locations) keySpan(pointer string, src []byte) specmodel.Span {
	if loc, ok := l.keys[pointer]; ok {
		return loc.key
	}
	if sp, ok := l.tables[pointer]; ok {
		return sp
	}
	if i := strings.LastIndex(pointer, "/"); i > 0 {
		return l.keySpan(pointer[:i], src)
	}
	return span.Of(src, 0, len(src))
}

// text returns the exact source text of pointer's value — quotes and
// all — the same "report what the author wrote" discipline
// frontmatter.Property.Raw applies, or "" when the scanner located no
// value for pointer.
func (l locations) text(pointer string, src []byte) string {
	loc, ok := l.keys[pointer]
	if !ok {
		return ""
	}
	return string(span.Slice(src, loc.value))
}
