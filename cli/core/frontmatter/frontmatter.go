// Package frontmatter splits the frontmatter block off a markdown
// document and parses it to a generic model that keeps each property's
// raw source text.
//
// Two consumers read the result and they need different things. Schema
// validation needs a generic value tree, because the schema arrives at
// runtime from whichever profile is active and no Go type corresponds to
// it. Typed decoding needs JSON, because the generated records carry
// json tags and no YAML ones. Both are served from one parse of the
// source rather than two, which is what keeps the position of a property
// and the value of that property from ever disagreeing.
//
// The raw text is the reason this package exists in the shape it does. A
// YAML parser resolves `updated_at: 2026-07-29` to the string
// "2026-07-29" or to a timestamp depending on the library and the tag
// resolution it applies, and a format check run against the resolved
// value tests what the parser inferred rather than what the author
// wrote. Keeping the source text of every property alongside its parsed
// value lets the Schema Registry check the written form (AC-SPEC-003.2).
//
// The block is located with the Spans delimiter scanner, never with an
// independent search for the delimiters, so a position reported here is
// computed the same way as a position reported by any other parser.
package frontmatter

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// The delimiters of a frontmatter block.
//
// The closing delimiter carries its leading newline so that only a `---`
// beginning a line ends the block: a run of three hyphens inside a value
// is content, which is the same rule every markdown frontmatter
// convention applies.
const (
	openingDelimiter = "---"
	closingDelimiter = "\n---"
)

// Property is one top-level frontmatter property.
type Property struct {
	// Name is the property's key, with any quoting resolved.
	Name string

	// Key is where the key itself sits in the document, so a finding
	// about a property that should not be there can point at the name
	// rather than at the value.
	Key specmodel.Span

	// Position is where the value sits in the document.
	Position specmodel.Span

	// Raw is the value's source text exactly as written, quotes and all,
	// with surrounding whitespace removed.
	//
	// This is what a format constraint is checked against. For a value
	// spanning several lines — a sequence, a nested mapping, a block
	// scalar — it is the whole of that value's source.
	Raw string

	// Value is the parsed value: a string, a bool, a number, a
	// []any, or a map[string]any. It is JSON-compatible, so the whole
	// property set converts to JSON without a second pass over the
	// source.
	Value any
}

// Document is one markdown document's parsed frontmatter.
type Document struct {
	// Region is where the block sits in the document: Outer covers both
	// delimiters, Inner covers the YAML between them.
	Region span.Region

	// Properties are the block's top-level properties, in the order the
	// document declares them.
	Properties []Property

	values map[string]any
	index  map[string]int
}

// Parse returns the frontmatter of src.
//
// The three outcomes are distinct and the caller acts on each one
// differently:
//
//   - a document and no error — the block was found and parsed;
//   - no document and no error — src carries no frontmatter block, which
//     is not a defect here: the caller records the file as excluded with
//     reason no-frontmatter;
//   - no document and a located error — the block is there and could not
//     be read. Parse never panics on malformed input, so the caller
//     records the failure against this document and carries on indexing
//     the rest of the corpus (AC-SPEC-002.4).
//
// A block is only recognised at the very start of the document. A `---`
// further down is content.
func Parse(src []byte) (*Document, *specmodel.ParseError) {
	region, found := span.Find(src, openingDelimiter, closingDelimiter, 0)
	if !found || region.Outer.Offset != 0 {
		if bytes.HasPrefix(src, []byte(openingDelimiter)) {
			return nil, located(
				"the frontmatter block is opened and never closed",
				span.Of(src, 0, len(openingDelimiter)))
		}
		return nil, nil
	}

	inner := span.Slice(src, region.Inner)
	file, err := parser.ParseBytes(inner, 0)
	if err != nil {
		return nil, syntaxError(src, region, err)
	}

	doc := &Document{
		Region: region,
		values: map[string]any{},
		index:  map[string]int{},
	}

	body := documentBody(file)
	if body == nil {
		// An empty block declares no properties. That is a document with
		// nothing in its frontmatter, not a document with no frontmatter,
		// and the two are reported differently.
		return doc, nil
	}

	pairs, ok := mappingPairs(body)
	if !ok {
		return nil, located(
			fmt.Sprintf("the frontmatter block is %s rather than a mapping of properties", describe(body)),
			region.Inner)
	}

	for i, pair := range pairs {
		property, perr := readProperty(src, region, pairs, i, pair)
		if perr != nil {
			return nil, perr
		}
		doc.index[property.Name] = len(doc.Properties)
		doc.Properties = append(doc.Properties, property)
		doc.values[property.Name] = property.Value
	}
	return doc, nil
}

// Property returns the named property.
func (d *Document) Property(name string) (Property, bool) {
	i, ok := d.index[name]
	if !ok {
		return Property{}, false
	}
	return d.Properties[i], true
}

// Values returns the generic value tree the block declares, keyed by
// property name.
//
// This is what a JSON Schema is validated against. The map is the
// document's own; a caller that modifies it corrupts every later read of
// the same document, so callers treat it as read-only.
func (d *Document) Values() map[string]any {
	return d.values
}

// JSON returns the value tree as JSON, which is how a block is decoded
// into one of the generated records: they carry json tags and no yaml
// ones, so JSON is the crossing point between the generic model and the
// typed one.
func (d *Document) JSON() ([]byte, error) {
	return json.Marshal(d.values)
}

// readProperty builds one property from a key/value pair, resolving both
// its position in the whole document and its raw source text.
func readProperty(src []byte, region span.Region, pairs []*ast.MappingValueNode, i int, pair *ast.MappingValueNode) (Property, *specmodel.ParseError) {
	keyToken := pair.Key.GetToken()
	if keyToken == nil {
		return Property{}, located("a frontmatter property has no name", region.Inner)
	}

	keyStart := absolute(region, src, keyToken)
	valueStart := absolute(region, src, valueToken(pair.Value))

	// The property runs to the start of the next one, or to the end of
	// the block. Trimming the trailing whitespace off that extent is what
	// turns it into the value's own source text, and it is the only form
	// that is correct for a multi-line value as well as a scalar.
	end := int(region.Inner.Offset) + int(region.Inner.Length)
	if i+1 < len(pairs) {
		if next := pairs[i+1].Key.GetToken(); next != nil {
			end = absolute(region, src, next)
		}
	}
	if end < valueStart {
		end = valueStart
	}
	raw := strings.TrimRight(string(src[valueStart:end]), " \t\r\n")

	var value any
	if err := yaml.NodeToValue(pair.Value, &value); err != nil {
		return Property{}, located(
			fmt.Sprintf("property %q could not be read: %s", keyToken.Value, err),
			span.Of(src, valueStart, len(raw)))
	}

	return Property{
		Name:     keyToken.Value,
		Key:      span.Of(src, keyStart, tokenWidth(keyToken)),
		Position: span.Of(src, valueStart, len(raw)),
		Raw:      raw,
		Value:    value,
	}, nil
}

// documentBody returns the body of the first YAML document in the block,
// or nil when the block is empty.
func documentBody(file *ast.File) ast.Node {
	for _, doc := range file.Docs {
		if doc.Body != nil {
			return doc.Body
		}
	}
	return nil
}

// mappingPairs returns the top-level key/value pairs of a mapping body.
//
// A block declaring exactly one property parses to a single pair rather
// than to a mapping of one, so both shapes are accepted.
func mappingPairs(body ast.Node) ([]*ast.MappingValueNode, bool) {
	switch node := body.(type) {
	case *ast.MappingNode:
		return node.Values, true
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{node}, true
	default:
		return nil, false
	}
}

// valueToken returns the token a value's source text starts at.
//
// A mapping reports the colon of its first pair rather than that pair's
// key, which would place the value's span in the middle of its own first
// property; every other node type reports its first token, which is
// where its source begins.
func valueToken(value ast.Node) *token.Token {
	if mapping, ok := value.(*ast.MappingNode); ok && len(mapping.Values) > 0 {
		if first := mapping.Values[0].Key.GetToken(); first != nil {
			return first
		}
	}
	return value.GetToken()
}

// absolute converts a token's position within the parsed block to a byte
// offset in the whole document.
//
// tok.Position.Offset is one-based, but it is not a byte offset: the
// underlying YAML scanner (goccy/go-yaml) advances it one unit per
// Unicode codepoint rather than per byte — verified against
// v1.19.2's scanner, which progresses position by
// len([]rune(value)) rather than len(value). Treating it as a byte
// offset directly is correct only while the block is pure ASCII; a
// block whose title or description contains an em dash, a curly
// quote, or any other multi-byte character makes every token
// reported after it land short of its true byte position, by the
// accumulated byte/rune width difference of everything before it in
// the block. That is a silent, self-application-only failure mode: no
// unit fixture in this package's own tests uses non-ASCII prose, so it
// went undetected until the profile self-application gate ran the
// same code over this repository's real documents (several of which
// use em dashes in their titles and descriptions) and it corrupted
// the raw text schemaregistry.instanceValue checks a format constraint
// against.
//
// The fix re-derives the byte offset from tok.Position by counting
// codepoints into region.Inner's own bytes — the exact slice the YAML
// parser tokenized — rather than trusting the library's Offset field.
func absolute(region span.Region, src []byte, tok *token.Token) int {
	if tok == nil {
		return int(region.Inner.Offset)
	}
	runeOffset := tok.Position.Offset - 1
	if runeOffset < 0 {
		runeOffset = 0
	}
	inner := span.Slice(src, region.Inner)
	return int(region.Inner.Offset) + byteOffsetOfRune(inner, runeOffset)
}

// byteOffsetOfRune returns the byte offset of the runeOffset-th Unicode
// codepoint (0-based) within data, or len(data) if data has fewer than
// runeOffset codepoints.
func byteOffsetOfRune(data []byte, runeOffset int) int {
	if runeOffset <= 0 {
		return 0
	}
	count := 0
	for i := range string(data) {
		if count == runeOffset {
			return i
		}
		count++
	}
	return len(data)
}

// tokenWidth returns the width in bytes of a token's source text,
// including any quoting, so a span over a key covers what the author
// wrote rather than what the parser resolved.
func tokenWidth(tok *token.Token) int {
	if width := len(strings.TrimSpace(tok.Origin)); width > 0 {
		return width
	}
	if width := len(tok.Value); width > 0 {
		return width
	}
	return 1
}

// syntaxError converts a YAML parse failure into a located parse error,
// using the position the parser reports where it has one.
func syntaxError(src []byte, region span.Region, err error) *specmodel.ParseError {
	var syntax *yaml.SyntaxError
	if stderrors.As(err, &syntax) {
		if tok := syntax.GetToken(); tok != nil {
			return located(syntax.GetMessage(), span.Of(src, absolute(region, src, tok), tokenWidth(tok)))
		}
		return located(syntax.GetMessage(), region.Inner)
	}
	return located(err.Error(), region.Inner)
}

// located pairs a message with the position it is about.
func located(message string, at specmodel.Span) *specmodel.ParseError {
	return &specmodel.ParseError{Message: message, Position: at}
}

// describe names a node type in the words a document author would use,
// so the message about a block that is not a mapping says what it is.
func describe(node ast.Node) string {
	switch node.(type) {
	case *ast.SequenceNode:
		return "a sequence"
	case *ast.StringNode, *ast.LiteralNode:
		return "a bare string"
	default:
		return "a bare value"
	}
}
