package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// YAML rendering for this command surface, per
// [[cli-output-standards]]'s YAML Style section.
//
// The renderer walks the same struct `encoding/json` marshals and
// resolves its fields the same way — the same `json` tags, the same
// promotion of embedded structs, the same dominance rules between two
// fields of one name — so the two renderings carry the same keys in the
// same order. The standards require that ("the same values in the same
// order"), and deriving one rendering from the other's declaration is
// what makes it true by construction rather than by two authors
// agreeing; a value that appears in `--json` and not in the YAML would
// be a bug the type system cannot see.
//
// One difference is deliberate: an empty or nil sequence renders `[]`
// and an empty or nil mapping `{}`, per the standards' YAML Style
// section, where `encoding/json` writes `null` for a nil one.
//
// A value that marshals ITSELF — every union and enum
// `schemas/*.strata` generates does — is rendered from the JSON its own
// MarshalJSON writes rather than from its Go fields, because those two
// are not the same document. A union is one Go field holding a variant
// interface, and its JSON is the variant's own keys under a tag naming
// which variant it is; walking the fields would print the Go shape
// (`Value:`) and lose the tag, so the two renderings would carry
// different keys for one value. See [marshaledValue].
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
	v = marshaledValue(v)

	// An invalid value is a nil pointer or interface: an absent value,
	// which the standards render as null.
	if !v.IsValid() {
		writeScalarLine(out, "null", indent, inSequence)
		return
	}

	if mapping, ok := asJSONMapping(v); ok {
		writeYAMLMapping(out, mapping, indent, inSequence)
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

		value := marshaledValue(f.value)
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
	v = marshaledValue(v)
	if !v.IsValid() {
		return false
	}
	if mapping, ok := asJSONMapping(v); ok {
		return len(mapping) > 0
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

// jsonMapping is a JSON object decoded from a value's own MarshalJSON,
// holding its keys in the order that JSON wrote them.
//
// A Go map cannot stand in for it: the standards require a stable order
// and a map has none, so the order the value itself chose would be
// replaced by one this renderer invented.
type jsonMapping []field

// asJSONMapping reports whether v is a decoded JSON object, and returns
// it when so. It is checked before the Kind switch because a jsonMapping
// is a slice and would otherwise render as a sequence.
func asJSONMapping(v reflect.Value) (jsonMapping, bool) {
	if !v.IsValid() || v.Type() != reflect.TypeOf(jsonMapping(nil)) {
		return nil, false
	}
	mapping, ok := v.Interface().(jsonMapping)
	return mapping, ok
}

// marshaledValue resolves v to what this renderer walks: the value
// beneath a pointer or interface, and — for a value that marshals itself
// — the tree its own JSON carries rather than its Go fields.
//
// A value whose marshalling fails, or whose JSON cannot be decoded, is
// returned unchanged and rendered from its fields. The rendering stays
// total either way: `--json` is where a marshalling failure is reported,
// and a YAML rendering that panicked or vanished would report it twice
// and answer neither question.
func marshaledValue(v reflect.Value) reflect.Value {
	v = indirect(v)
	if !v.IsValid() || !v.CanInterface() {
		return v
	}
	if _, already := v.Interface().(jsonMapping); already {
		return v
	}
	marshaler, ok := v.Interface().(json.Marshaler)
	if !ok {
		return v
	}
	data, err := marshaler.MarshalJSON()
	if err != nil {
		return v
	}
	decoded, err := decodeOrderedJSON(data)
	if err != nil {
		return v
	}
	return reflect.ValueOf(decoded)
}

// decodeOrderedJSON decodes one JSON document into the generic tree this
// renderer walks: a [jsonMapping] for an object, a []any for an array,
// and a Go scalar otherwise.
//
// Numbers are decoded through [json.Number] and converted back to an
// integer where they are one, so that a whole number renders unquoted
// and without a decimal point it was not written with.
func decodeOrderedJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	value, err := decodeOrderedValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("cmd: trailing content after a value's own JSON")
	}
	return value, nil
}

// decodeOrderedValue decodes the next value the decoder offers.
func decodeOrderedValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	return decodeOrderedToken(decoder, token)
}

// decodeOrderedToken turns one token, and whatever it opens, into a
// value.
func decodeOrderedToken(decoder *json.Decoder, token json.Token) (any, error) {
	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '{':
			return decodeOrderedMapping(decoder)
		case '[':
			return decodeOrderedSequence(decoder)
		}
		return nil, fmt.Errorf("cmd: unexpected %q in a value's own JSON", t)
	case json.Number:
		return orderedNumber(t), nil
	default:
		// A string, a boolean, or nil for a JSON null.
		return t, nil
	}
}

// decodeOrderedMapping decodes an object whose opening brace has already
// been read, keeping its keys in order.
func decodeOrderedMapping(decoder *json.Decoder) (jsonMapping, error) {
	mapping := jsonMapping{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("cmd: a non-string key in a value's own JSON")
		}
		value, err := decodeOrderedValue(decoder)
		if err != nil {
			return nil, err
		}
		mapping = append(mapping, field{key: name, value: reflect.ValueOf(value)})
	}
	// The closing brace.
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return mapping, nil
}

// decodeOrderedSequence decodes an array whose opening bracket has
// already been read.
func decodeOrderedSequence(decoder *json.Decoder) ([]any, error) {
	sequence := []any{}
	for decoder.More() {
		value, err := decodeOrderedValue(decoder)
		if err != nil {
			return nil, err
		}
		sequence = append(sequence, value)
	}
	// The closing bracket.
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return sequence, nil
}

// orderedNumber narrows a JSON number to the Go type that renders it as
// written: an integer where it is one, a float otherwise, and the
// number's own text where it fits neither — a value too large for either
// is still reported, in the digits it arrived in.
func orderedNumber(number json.Number) any {
	if i, err := strconv.ParseInt(number.String(), 10, 64); err == nil {
		return i
	}
	if u, err := strconv.ParseUint(number.String(), 10, 64); err == nil {
		return u
	}
	if f, err := number.Float64(); err == nil {
		return f
	}
	return number.String()
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

// structFields returns a struct's rendered fields: the same keys, under
// the same names, in the same order `encoding/json` writes them.
//
// The field set is resolved by typeFields below; this walks each
// resolved index path to the value, skipping a field reached through a
// nil embedded pointer exactly as `encoding/json` skips it.
func structFields(v reflect.Value) []field {
	var fields []field
resolve:
	for _, spec := range typeFields(v.Type()) {
		fv := v
		for _, i := range spec.index {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					// The embedded struct is absent, and so is every
					// field promoted through it.
					continue resolve
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}
		if spec.omitEmpty && isEmptyValue(fv) {
			continue
		}
		fields = append(fields, field{key: spec.key, value: fv})
	}
	return fields
}

// fieldSpec is one visible field of a struct type: the key it renders
// under, the index path reaching it through any embedded structs,
// whether a json tag NAMED it, and whether that tag asked for
// omitempty.
type fieldSpec struct {
	key       string
	index     []int
	tagged    bool
	omitEmpty bool
}

// embeddedType is a struct type awaiting exploration, with the index
// path that reached it.
type embeddedType struct {
	typ   reflect.Type
	index []int
}

// typeFields resolves a struct type's visible fields the way
// `encoding/json` resolves them: a breadth-first walk over the type and
// the struct types it embeds, then the dominance rules that decide which
// of two fields sharing a name is visible, and finally a return to
// declaration order.
//
// Reproducing the walk is what makes the two renderings agree. A
// shallower field hides a promoted one of the same name; two fields at
// the same depth hide each other unless exactly one is named by a tag;
// and an embedded field is flattened only when no tag names it — the
// rules a hand-rolled reading of the tags gets wrong in the direction of
// silently dropping a value.
func typeFields(t reflect.Type) []fieldSpec {
	var fields []fieldSpec
	current := []embeddedType{}
	next := []embeddedType{{typ: t}}
	var count, nextCount map[reflect.Type]int
	visited := map[reflect.Type]bool{}

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, e := range current {
			if visited[e.typ] {
				continue
			}
			visited[e.typ] = true

			for i := range e.typ.NumField() {
				sf := e.typ.Field(i)
				if sf.Anonymous {
					et := sf.Type
					if et.Kind() == reflect.Pointer {
						et = et.Elem()
					}
					if !sf.IsExported() && et.Kind() != reflect.Struct {
						// An embedded field of an unexported NON-struct
						// type is invisible. One of an unexported STRUCT
						// type is not: its exported fields are promoted
						// through it, and the type's own name never
						// appears.
						continue
					}
				} else if !sf.IsExported() {
					// Unexported, and invisible to encoding/json too.
					continue
				}

				tag := sf.Tag.Get("json")
				if tag == "-" {
					continue
				}
				name, opts := parseJSONTag(tag)
				if !validTagName(name) {
					name = ""
				}

				index := make([]int, len(e.index)+1)
				copy(index, e.index)
				index[len(e.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}

				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					fields = append(fields, fieldSpec{
						key:       name,
						index:     index,
						tagged:    tagged,
						omitEmpty: opts["omitempty"],
					})
					if count[e.typ] > 1 {
						// The embedding type was reached by more than
						// one path, so record a duplicate for the
						// dominance rules to annihilate.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// An embedded struct no tag names: explore it in the
				// next round, one level deeper.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, embeddedType{typ: ft, index: index})
				}
			}
		}
	}

	// Group by name, shallowest first, and within one depth the
	// tag-named field ahead of the one named after its Go field.
	sort.Slice(fields, func(i, j int) bool {
		x, y := fields[i], fields[j]
		if x.key != y.key {
			return x.key < y.key
		}
		if len(x.index) != len(y.index) {
			return len(x.index) < len(y.index)
		}
		if x.tagged != y.tagged {
			return x.tagged
		}
		return lessByIndex(x, y)
	})

	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		fi := fields[i]
		for advance = 1; i+advance < len(fields); advance++ {
			if fields[i+advance].key != fi.key {
				break
			}
		}
		if advance == 1 {
			out = append(out, fi)
			continue
		}
		if dominant, ok := dominantField(fields[i : i+advance]); ok {
			out = append(out, dominant)
		}
	}
	fields = out

	sort.Slice(fields, func(i, j int) bool { return lessByIndex(fields[i], fields[j]) })
	return fields
}

// dominantField picks the visible field from a group sharing one name,
// already ordered shallowest-first and tag-named-first. The first is
// therefore the dominant one unless the second ties it on both depth and
// tagging, in which case Go's embedding rules hide both and the name
// renders in neither rendering.
func dominantField(fields []fieldSpec) (fieldSpec, bool) {
	if len(fields) > 1 &&
		len(fields[0].index) == len(fields[1].index) &&
		fields[0].tagged == fields[1].tagged {
		return fieldSpec{}, false
	}
	return fields[0], true
}

// lessByIndex orders two fields by their index paths, which restores
// declaration order with a promoted field sitting where its embedding
// field was declared.
func lessByIndex(x, y fieldSpec) bool {
	for k, xk := range x.index {
		if k >= len(y.index) {
			return false
		}
		if xk != y.index[k] {
			return xk < y.index[k]
		}
	}
	return len(x.index) < len(y.index)
}

// parseJSONTag splits a json tag into the name it asks for and the
// options that follow it. An absent tag yields an empty name, which is
// what lets an untagged embedded struct be told from one tagged with a
// name.
func parseJSONTag(tag string) (string, map[string]bool) {
	name, rest, _ := strings.Cut(tag, ",")
	opts := map[string]bool{}
	for rest != "" {
		var opt string
		opt, rest, _ = strings.Cut(rest, ",")
		opts[opt] = true
	}
	return name, opts
}

// tagNamePunctuation is the punctuation `encoding/json` accepts inside a
// tag name; a backslash or a quote is reserved.
const tagNamePunctuation = "!#$%&()*+-./:;<=>?@[]^_{|}~ "

// validTagName reports whether a json tag names a field, on
// `encoding/json`'s terms. A name it rejects is ignored rather than
// honoured, so the field renders under its Go name — or, for an embedded
// struct, flattens.
func validTagName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if strings.ContainsRune(tagNamePunctuation, r) {
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
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
