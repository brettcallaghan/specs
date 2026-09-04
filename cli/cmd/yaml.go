package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// YAML rendering for this command surface, per
// [[cli-output-standards]]'s YAML Style section.
//
// The renderer walks the same struct `encoding/json` marshals, reading
// the same `json` tags in the same field order, because the standards
// require both renderings to carry "the same values in the same order".
// Deriving one rendering from the other's declaration is what makes that
// true by construction rather than by two authors agreeing; a value that
// appears in `--json` and not in the YAML would be a bug the type system
// cannot see.
//
// It is deliberately not a general YAML library. Every rule below removes
// a choice a library would offer — no flow style, no anchors, no folding,
// no document markers — so that two renderings of one value are the same
// bytes. A library configured into this shape would be a larger surface
// making the same guarantees less legibly.

// renderYAML writes value to a string as the standards' YAML: block
// style, two-space indent, and nothing a reader must resolve before
// knowing what the document says.
//
// The result always ends in a newline, and never carries a document-start
// or document-end marker.
func renderYAML(value any) string {
	var out strings.Builder
	writeYAMLValue(&out, reflect.ValueOf(value), 0, false)
	return out.String()
}

// writeYAMLValue writes one value at the given indent level.
//
// inSequence marks a value being written directly after a "- " marker,
// whose first line is therefore already indented and must not be
// indented again.
func writeYAMLValue(out *strings.Builder, v reflect.Value, indent int, inSequence bool) {
	v = indirect(v)

	// An invalid value is a nil pointer or interface: an absent value,
	// which the standards render as null.
	if !v.IsValid() {
		writeScalarLine(out, "null", indent, inSequence)
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		writeYAMLMapping(out, structFields(v), indent, inSequence)
	case reflect.Map:
		writeYAMLMapping(out, mapFields(v), indent, inSequence)
	case reflect.Slice, reflect.Array:
		writeYAMLSequence(out, v, indent, inSequence)
	default:
		writeScalar(out, v, indent, inSequence)
	}
}

// field is one key and its value, in the order the rendering reports it.
type field struct {
	key   string
	value reflect.Value
}

// writeYAMLMapping writes a mapping, or "{}" where it carries no key.
func writeYAMLMapping(out *strings.Builder, fields []field, indent int, inSequence bool) {
	if len(fields) == 0 {
		writeScalarLine(out, "{}", indent, inSequence)
		return
	}

	for i, f := range fields {
		// A mapping opening a sequence entry writes its first key on the
		// "- " line already emitted; every later key is indented to match.
		if i > 0 || !inSequence {
			out.WriteString(strings.Repeat(" ", indent))
		}
		out.WriteString(plainKey(f.key))
		out.WriteString(":")

		value := indirect(f.value)
		if isBlock(value) {
			// A mapping or sequence value goes on the following lines,
			// indented one level under the key introducing it.
			out.WriteString("\n")
			writeYAMLValue(out, value, indent+2, false)
			continue
		}
		out.WriteString(" ")
		writeYAMLValue(out, value, indent+2, true)
	}
}

// writeYAMLSequence writes a sequence, or "[]" where it carries no entry.
func writeYAMLSequence(out *strings.Builder, v reflect.Value, indent int, inSequence bool) {
	if v.Len() == 0 {
		writeScalarLine(out, "[]", indent, inSequence)
		return
	}

	for i := range v.Len() {
		out.WriteString(strings.Repeat(" ", indent))
		out.WriteString("- ")
		writeYAMLValue(out, v.Index(i), indent+2, true)
	}
}

// isBlock reports whether a value renders across lines beneath its key
// rather than after it. An empty collection does not: it renders as "[]"
// or "{}" on the key's own line.
func isBlock(v reflect.Value) bool {
	v = indirect(v)
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Struct:
		return len(structFields(v)) > 0
	case reflect.Map:
		return v.Len() > 0
	case reflect.Slice, reflect.Array:
		return v.Len() > 0
	default:
		return false
	}
}

// writeScalar writes one scalar value and its newline.
func writeScalar(out *strings.Builder, v reflect.Value, indent int, inSequence bool) {
	switch v.Kind() {
	case reflect.String:
		writeString(out, v.String(), indent, inSequence)
	case reflect.Bool:
		writeScalarLine(out, fmt.Sprintf("%t", v.Bool()), indent, inSequence)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeScalarLine(out, fmt.Sprintf("%d", v.Int()), indent, inSequence)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeScalarLine(out, fmt.Sprintf("%d", v.Uint()), indent, inSequence)
	case reflect.Float32, reflect.Float64:
		writeScalarLine(out, trimFloat(v.Float()), indent, inSequence)
	default:
		writeString(out, fmt.Sprintf("%v", v.Interface()), indent, inSequence)
	}
}

// writeString writes a string as a block scalar, a plain scalar or a
// double-quoted one, whichever the standards require of its content.
func writeString(out *strings.Builder, s string, indent int, inSequence bool) {
	if strings.Contains(s, "\n") && blockSafe(s) {
		writeBlockScalar(out, s, indent, inSequence)
		return
	}
	if plainSafe(s) {
		writeScalarLine(out, s, indent, inSequence)
		return
	}
	writeScalarLine(out, quote(s), indent, inSequence)
}

// writeBlockScalar writes a multi-line string as a literal block, trimmed
// of leading and trailing whitespace with the whitespace inside it
// untouched.
//
// The "-" chomping indicator follows from the trim: the value carries no
// trailing newline, and the default chomping would add one back.
func writeBlockScalar(out *strings.Builder, s string, indent int, inSequence bool) {
	if !inSequence {
		out.WriteString(strings.Repeat(" ", indent))
	}
	out.WriteString("|-\n")
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if line == "" {
			// An empty line carries no indentation: trailing whitespace
			// on a blank line is invisible and some parsers read it back
			// as content.
			out.WriteString("\n")
			continue
		}
		out.WriteString(strings.Repeat(" ", indent))
		out.WriteString(line)
		out.WriteString("\n")
	}
}

// writeScalarLine writes one already-rendered scalar, indenting it unless
// it follows a "- " or a "key: " already written on this line.
func writeScalarLine(out *strings.Builder, s string, indent int, inSequence bool) {
	if !inSequence {
		out.WriteString(strings.Repeat(" ", indent))
	}
	out.WriteString(s)
	out.WriteString("\n")
}

// blockSafe reports whether a block scalar carries s faithfully. It does
// not when s holds a control character other than a line break or a tab:
// a block's only structure is the line break, so a carriage return would
// fold into the break beside it and the value would come back shorter
// than it went in.
func blockSafe(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

// yamlIndicators are the characters that carry meaning at the start of a
// plain scalar.
const yamlIndicators = "-?:,[]{}#&*!|>'\"%@`"

// numberLike matches what a YAML reader takes for a number rather than
// for the text of one.
var numberLike = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$`)

// boolLike is every spelling a YAML reader may take for a boolean or a
// null, compared without regard to letter case.
var boolLike = map[string]bool{
	"true": true, "false": true, "yes": true, "no": true,
	"on": true, "off": true, "null": true, "~": true, "": true,
}

// plainSafe reports whether s round-trips as a plain scalar: it does not
// when it is empty, has leading or trailing whitespace, begins with a
// YAML indicator character, contains a control character, contains ": "
// or " #", ends with ":", or reads as a number, a boolean or a null.
func plainSafe(s string) bool {
	if s == "" || s != strings.TrimSpace(s) {
		return false
	}
	if strings.ContainsAny(s[:1], yamlIndicators) {
		return false
	}
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":") {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	if numberLike.MatchString(s) || boolLike[strings.ToLower(s)] {
		return false
	}
	return true
}

// quote renders s as a double-quoted scalar, escaping what a double-
// quoted scalar reads specially.
func quote(s string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&out, `\x%02x`, r)
				continue
			}
			out.WriteRune(r)
		}
	}
	out.WriteByte('"')
	return out.String()
}

// plainKey renders a mapping key, quoting it on the same terms a value is
// quoted. Every key this surface emits is a json tag and renders plain;
// the guard is here so that a map keyed by corpus content cannot produce
// a document that parses back differently.
func plainKey(key string) string {
	if plainSafe(key) {
		return key
	}
	return quote(key)
}

// trimFloat renders a float without an exponent or a trailing zero, so
// that a whole number reads as one.
func trimFloat(f float64) string {
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%f", f), "0"), ".")
}

// indirect follows pointers and interfaces to the value beneath, and
// reports an invalid value for a nil one.
func indirect(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// structFields returns a struct's rendered fields, in declaration order,
// reading the same json tags encoding/json reads so that the two
// renderings carry the same values in the same order.
func structFields(v reflect.Value) []field {
	var fields []field
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if f.PkgPath != "" {
			// Unexported, and invisible to encoding/json too.
			continue
		}
		name, opts, ok := jsonName(f)
		if !ok {
			continue
		}
		value := v.Field(i)
		if opts["omitempty"] && isEmptyValue(value) {
			continue
		}
		if f.Anonymous && indirect(value).Kind() == reflect.Struct && name == "" {
			// An embedded struct with no tag flattens into its parent,
			// exactly as encoding/json flattens it.
			fields = append(fields, structFields(indirect(value))...)
			continue
		}
		if name == "" {
			name = f.Name
		}
		fields = append(fields, field{key: name, value: value})
	}
	return fields
}

// jsonName reads a field's json tag: the name it renders under, its
// options, and whether it is rendered at all.
func jsonName(f reflect.StructField) (string, map[string]bool, bool) {
	tag, tagged := f.Tag.Lookup("json")
	if !tagged {
		return f.Name, nil, true
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "-" && len(parts) == 1 {
		return "", nil, false
	}
	opts := map[string]bool{}
	for _, opt := range parts[1:] {
		opts[opt] = true
	}
	return parts[0], opts, true
}

// isEmptyValue mirrors encoding/json's own omitempty test, so a field
// omitted from one rendering is omitted from the other.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// mapFields returns a map's entries ordered by key, so that a rendering
// of one map is the same bytes every run — Go randomizes map iteration,
// which would otherwise make output depend on nothing at all.
func mapFields(v reflect.Value) []field {
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
	})
	fields := make([]field, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, field{key: fmt.Sprintf("%v", key.Interface()), value: v.MapIndex(key)})
	}
	return fields
}
