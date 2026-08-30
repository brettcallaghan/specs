package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// The YAML the command surface emits, and the style rules that fix it.
//
// A hand-written emitter rather than a library one, because the style is
// the point: every rule below removes a choice, so two renderings of the
// same value are the same bytes. A general emitter would wrap long lines
// at a width nobody is looking at, choose its own quoting, and pick a
// chomping indicator that does not round-trip.
//
// Three node kinds and no more. Everything a command reports — an answer,
// or the whole result — is built from them, so there is one place that
// decides how a value looks.

// yamlNode is one node of an emitted document. yamlLines renders it with
// no leading indentation; a container indents what its children return,
// which is what makes indentation compose without any node knowing its
// own depth.
type yamlNode interface {
	yamlLines() []string
}

// yScalar is a leaf. A string carries its value in str so the quoting
// rules can be applied to it; everything else arrives already rendered in
// lit, because a number, a boolean and a null have exactly one rendering.
type yScalar struct {
	lit string
	str *string
}

func yStr(s string) yScalar { return yScalar{str: &s} }
func yNull() yScalar        { return yScalar{lit: "null"} }
func yBool(b bool) yScalar  { return yScalar{lit: strconv.FormatBool(b)} }
func yInt(n int64) yScalar  { return yScalar{lit: strconv.FormatInt(n, 10)} }

// ySeq is a block sequence.
type ySeq []yamlNode

// yPair is one entry of a mapping.
type yPair struct {
	key   string
	value yamlNode
}

// yMap is a block mapping. A slice and not a map: the order a command
// established its properties in is part of what it reports, and a Go map
// would randomise it.
type yMap []yPair

// yamlIndicators are the characters that begin a YAML construct. A plain
// scalar starting with one of them parses as that construct instead.
const yamlIndicators = "-?:,[]{}#&*!|>'\"%@`"

func (s yScalar) yamlLines() []string {
	if s.str == nil {
		return []string{s.lit}
	}
	v := *s.str
	if strings.Contains(v, "\n") && blockScalarRoundTrips(v) {
		return blockScalar(v)
	}
	if plainRoundTrips(v) {
		return []string{v}
	}
	return []string{quoteYAML(v)}
}

func (s ySeq) yamlLines() []string {
	if len(s) == 0 {
		return []string{"[]"}
	}
	var out []string
	for _, item := range s {
		lines := item.yamlLines()
		out = append(out, "- "+lines[0])
		for _, line := range lines[1:] {
			out = append(out, indentLine(line))
		}
	}
	return out
}

func (m yMap) yamlLines() []string {
	if len(m) == 0 {
		return []string{"{}"}
	}
	var out []string
	for _, pair := range m {
		key := yamlKey(pair.key)
		lines := pair.value.yamlLines()
		_, leaf := pair.value.(yScalar)
		// An empty collection has no block form — a key with nothing
		// after it is a key whose value is null — so it is the one place
		// flow style is emitted, and it is forced rather than chosen.
		inline := leaf || (len(lines) == 1 && (lines[0] == "[]" || lines[0] == "{}"))
		if inline {
			out = append(out, key+": "+lines[0])
			for _, line := range lines[1:] {
				out = append(out, indentLine(line))
			}
			continue
		}
		out = append(out, key+":")
		for _, line := range lines {
			out = append(out, indentLine(line))
		}
	}
	return out
}

// indentLine indents one rendered line. An empty line stays empty: a
// blank line inside a block scalar that carried indentation would carry
// trailing whitespace, and the value would no longer round-trip.
func indentLine(line string) string {
	if line == "" {
		return ""
	}
	return "  " + line
}

// blockScalarRoundTrips reports whether a literal block can carry the
// value exactly. A value it cannot is double-quoted instead: less
// readable, and the only rendering that survives.
//
// Two cases, and the second is the trap. A block's only structure is the
// line break, so any OTHER control character in it is lost or
// reinterpreted — a carriage return most of all, which a reader folds
// into the line break beside it. And a block's last content line may end
// in white space by the specification, but a reader may drop it: measured
// on goccy/go-yaml, `v: |-` over `a` then `b ` reads back `a\nb`, while
// the same document through PyYAML reads back `a\nb `. The renderer
// cannot know which reader it is writing for, so it does not write the
// document they disagree about.
func blockScalarRoundTrips(v string) bool {
	for _, r := range v {
		if r == '\n' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	lines := strings.Split(strings.TrimRight(v, "\n"), "\n")
	last := lines[len(lines)-1]
	return !strings.HasSuffix(last, " ") && !strings.HasSuffix(last, "\t")
}

// blockScalar renders a multi-line string as a literal block, with the
// chomping indicator that reproduces the value exactly: `|-` when it ends
// with no newline, `|` when it ends with exactly one, `|+` when it ends
// with more.
//
// The explicit indentation indicator is not decoration either. A block's
// indentation is otherwise detected from its first non-empty line, so a
// value whose first line begins with a space would silently lose it.
func blockScalar(v string) []string {
	trailing := len(v) - len(strings.TrimRight(v, "\n"))
	header := "|-"
	switch {
	case trailing == 1:
		header = "|"
	case trailing > 1:
		header = "|+"
	}

	body := strings.Split(v, "\n")
	// The element after the final newline is not a line of the value.
	if len(body) > 0 && body[len(body)-1] == "" {
		body = body[:len(body)-1]
	}
	if len(body) > 0 && strings.HasPrefix(body[0], " ") {
		header = "|2" + strings.TrimPrefix(header, "|")
	}

	out := make([]string, 0, len(body)+1)
	out = append(out, header)
	out = append(out, body...)
	return out
}

// plainRoundTrips reports whether a one-line string parses back as
// itself when written unquoted. The list is the exact set of cases where
// it does not, so a reader can tell that a quoted value is quoted
// because it had to be.
func plainRoundTrips(v string) bool {
	if v == "" {
		return false
	}
	if strings.TrimSpace(v) != v {
		return false
	}
	for _, r := range v {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	if strings.ContainsRune(yamlIndicators, rune(v[0])) {
		return false
	}
	if strings.Contains(v, ": ") || strings.Contains(v, " #") {
		return false
	}
	if strings.HasSuffix(v, ":") {
		return false
	}
	return !readsAsScalar(v)
}

// readsAsScalar reports whether a plain string would be parsed as a
// number, a boolean or a null rather than as text. The boolean set is the
// wider YAML 1.1 one rather than the 1.2 core one, because a consumer's
// parser may be either and quoting one word too many costs nothing.
func readsAsScalar(v string) bool {
	switch strings.ToLower(v) {
	case "null", "~", "true", "false", "yes", "no", "on", "off", "y", "n",
		".inf", "-.inf", "+.inf", ".nan":
		return true
	}
	if _, err := strconv.ParseInt(v, 0, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return true
	}
	return false
}

// quoteYAML renders a string double-quoted. Go's own quoting is a subset
// of YAML's double-quoted escapes, so it is reused rather than
// reimplemented.
func quoteYAML(v string) string { return strconv.Quote(v) }

// yamlKey renders a mapping key. A key is a scalar like any other, with
// one addition: a colon anywhere in it, not only before a space, is
// ambiguous in key position.
func yamlKey(k string) string {
	if strings.Contains(k, ":") || !plainRoundTrips(k) {
		return quoteYAML(k)
	}
	return k
}

// writeYAML emits a document. No document-start marker, no document-end
// marker, no comments, no anchors: each is a feature a reader would have
// to resolve before they know what the document says.
func writeYAML(w io.Writer, n yamlNode) error {
	for _, line := range n.yamlLines() {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// --- the same nodes as JSON -----------------------------------------

// The two renderings come from one node tree, which is what keeps them
// carrying the same values in the same order. Ordering is why these are
// hand-marshalled: a map[string]any would sort the keys and lose the
// order the command reported them in.

func (s yScalar) MarshalJSON() ([]byte, error) {
	if s.str != nil {
		return json.Marshal(*s.str)
	}
	return []byte(s.lit), nil
}

func (s ySeq) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('[')
	for i, item := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		b.Write(encoded)
	}
	b.WriteByte(']')
	return b.Bytes(), nil
}

func (m yMap) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, pair := range m {
		if i > 0 {
			b.WriteByte(',')
		}
		key, err := json.Marshal(pair.key)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')
		value, err := json.Marshal(pair.value)
		if err != nil {
			return nil, err
		}
		b.Write(value)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// writeJSON emits the same node tree as indented JSON.
func writeJSON(w io.Writer, n yamlNode) error {
	encoded, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return fmt.Errorf("cli: rendering the answer as JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", encoded)
	return err
}

// --- a generated struct as nodes ------------------------------------

// nodeOf projects a value of a generated contract type onto the node
// tree, in declaration order and under the field names its JSON tags fix.
//
// Reflection rather than a hand-written projection: the shape lives in
// schemas/swarm/output.strata and reaches this package as generated Go, so
// a field added there reaches both renderings without anyone editing a
// third place.
func nodeOf(v reflect.Value) yamlNode {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return yNull()
		}
		return nodeOf(v.Elem())
	case reflect.Struct:
		m := yMap{}
		t := v.Type()
		for i := range t.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			name, opts, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			value := v.Field(i)
			if strings.Contains(opts, "omitempty") && isEmpty(value) {
				continue
			}
			m = append(m, yPair{name, nodeOf(value)})
		}
		return m
	case reflect.Slice, reflect.Array:
		s := ySeq{}
		for i := range v.Len() {
			s = append(s, nodeOf(v.Index(i)))
		}
		return s
	case reflect.String:
		return yStr(v.String())
	case reflect.Bool:
		return yBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return yInt(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return yInt(int64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return yScalar{lit: strconv.FormatFloat(v.Float(), 'g', -1, 64)}
	default:
		return yNull()
	}
}

// isEmpty is encoding/json's omitempty test, so a field this renderer
// drops is exactly a field the JSON rendering drops.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return false
	}
}
