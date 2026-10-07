// Package c4diagram parses Mermaid C4 diagram blocks into located
// statements.
//
// It reads a deliberately narrow subset of Mermaid's C4 notation — the
// statement table in [macros] — and turns every other statement into a
// [ParseError] rather than skipping it, so the subset grows only by an
// edit to that table. It is pure notation: a block of text in, statements
// with spans out. It never reads the corpus and never resolves a node to
// an element; both belong to its callers.
//
// Every position is a [span.Span] computed by [span.Of], in the
// coordinates of the whole document the block was read from, so a
// finding raised against a statement points at the line an author edits.
package c4diagram

import (
	"sort"
	"strconv"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// The diagram types this package parses. Every other first statement is
// reported by name and not parsed.
const (
	TypeContext   = "C4Context"
	TypeContainer = "C4Container"
	TypeComponent = "C4Component"
)

// Kind classifies a statement.
type Kind string

// The statement kinds. Every statement the parser reads is exactly one of
// them.
const (
	// KindElement is a system, container or component node — including
	// its Db and Queue forms — that names an element.
	KindElement Kind = "element"
	// KindExternal is a Person or any _Ext node: drawn, but never an
	// element of the corpus.
	KindExternal Kind = "external"
	// KindBoundary groups nodes and is not itself a node.
	KindBoundary Kind = "boundary"
	// KindRelationship is an arrow between two aliases.
	KindRelationship Kind = "relationship"
	// KindIgnored is a statement that is read but carries nothing the
	// checks examine: a title, or a style or layout directive.
	KindIgnored Kind = "ignored"
)

// Parameter names, in the vocabulary [Statement.Param] is asked in.
const (
	ParamAlias       = "alias"
	ParamLabel       = "label"
	ParamFrom        = "from"
	ParamTo          = "to"
	ParamTechnology  = "technology"
	ParamDescription = "description"
	ParamType        = "type"
	ParamSprite      = "sprite"
	ParamTags        = "tags"
	ParamLink        = "link"
)

// Diagram is what [Parse] read from one block.
type Diagram struct {
	// Type is the diagram type the block's first statement names — its
	// first whitespace-separated word, as written. Empty when the block
	// holds no statement at all.
	Type string

	// TypeSpan covers the first statement, trimmed of surrounding
	// whitespace. Zero-length at the block's start when Type is empty.
	TypeSpan span.Span

	// Statements are the statements read, in the order written, with the
	// diagram-type line excluded. Always empty unless [Diagram.Read].
	Statements []Statement

	// Errors are the statements that could not be read, in the order
	// written. Each one was skipped and parsing continued with the next
	// line.
	Errors []ParseError
}

// Read reports whether the diagram is one of the types this package
// parses. A diagram that is not has a [Diagram.Type] and nothing else.
func (d Diagram) Read() bool {
	switch d.Type {
	case TypeContext, TypeContainer, TypeComponent:
		return true
	}
	return false
}

// Statement is one line of a C4 diagram that the parser read.
type Statement struct {
	// Kind classifies the statement.
	Kind Kind

	// Macro is the macro as written — `Rel_Back`, `ContainerDb_Ext`,
	// `title`.
	Macro string

	// Canonical is the macro after normalisation. Every directional
	// relationship variant and `Rel_Back` is `Rel`; every other macro is
	// itself.
	Canonical string

	// Span covers the statement, trimmed of surrounding whitespace.
	Span span.Span

	// Args are the positional arguments, in the order written.
	Args []Argument

	// Named are the `$<name>=<value>` arguments, in the order written.
	// They occupy no positional slot.
	Named []Argument

	// Parent is the index into [Diagram.Statements] of the innermost
	// boundary open around this statement, or -1 at the top level.
	Parent int

	// Store marks a Db or Queue node — the only nodes that may stand for
	// a passive store rather than an element.
	Store bool

	// Relationship is set on a statement of [KindRelationship] and nil on
	// every other.
	Relationship *Relationship

	params []string
}

// Argument is one argument of a statement.
type Argument struct {
	// Name is the parameter name of a `$<name>=<value>` argument, without
	// the `$`, and empty for a positional one.
	Name string

	// Value is the argument's text: between the quotes for a quoted
	// argument, trimmed of surrounding whitespace for an unquoted one.
	Value string

	// Quoted reports whether the value was written in double quotes.
	Quoted bool

	// Span covers the value as written — including its quotes, when it
	// has them.
	Span span.Span
}

// Param returns the argument the macro's parameter of this name was
// given, positionally or as a named argument. A named argument wins over
// a positional one, since Mermaid applies it last.
func (s Statement) Param(name string) (Argument, bool) {
	if arg, ok := s.named(name); ok {
		return arg, true
	}
	slot := 0
	for _, param := range s.params {
		if _, ok := s.named(param); ok {
			continue // written by name: it takes no positional slot
		}
		if param == name {
			if slot < len(s.Args) {
				return s.Args[slot], true
			}
			return Argument{}, false
		}
		slot++
	}
	return Argument{}, false
}

// named returns the `$name` argument that names the parameter, if any.
func (s Statement) named(name string) (Argument, bool) {
	for _, arg := range s.Named {
		if namedAs(arg.Name, name) {
			return arg, true
		}
	}
	return Argument{}, false
}

// namedAs reports whether a `$name` argument names the parameter called
// param, accepting Mermaid's own short names for the two parameters that
// have them.
func namedAs(written, param string) bool {
	if written == param {
		return true
	}
	switch param {
	case ParamTechnology:
		return written == "techn"
	case ParamDescription:
		return written == "descr"
	}
	return false
}

// Relationship is the dependency a relationship statement states.
type Relationship struct {
	// Source is the dependent end: the first alias, or the second for
	// `Rel_Back`.
	Source Argument

	// Target is the end depended upon.
	Target Argument

	// Bidirectional marks a `BiRel`, which states no single direction.
	Bidirectional bool

	// Tags are the relationship's tags, read from its tags argument.
	Tags Tags
}

// Tags is a relationship's tags argument, read.
type Tags struct {
	// Present reports whether the statement has a tags argument at all.
	Present bool

	// Raw is the argument's value as written.
	Raw string

	// Span covers the argument as written.
	Span span.Span

	// Reserved are the tags with meaning to this framework, in the order
	// written.
	Reserved []ReservedTag

	// Unreserved are every other tag, in the order written.
	Unreserved []string

	// Comma reports that the argument contains a comma. Tags are split on
	// `+` only, so a comma-joined pair is read as one unreserved tag.
	Comma bool
}

// The reserved tag names.
const (
	TagSync     = "sync"
	TagAsync    = "async"
	TagContract = "contract"
)

// ReservedTag is one reserved relationship tag.
type ReservedTag struct {
	// Name is [TagSync], [TagAsync] or [TagContract].
	Name string

	// Slug is the contract slug written after `contract:`, and empty for
	// every other reserved tag.
	Slug string
}

// ParseError is a statement the parser could not read.
type ParseError struct {
	// Span covers the statement, trimmed of surrounding whitespace.
	Span span.Span

	// Statement is the statement's text, as Span covers it.
	Statement string

	// Reason says what was wrong with it.
	Reason string
}

// macro is one row of the statement table.
type macro struct {
	kind     Kind
	params   []string
	required int
	store    bool
}

var (
	relParams       = []string{ParamFrom, ParamTo, ParamLabel, ParamTechnology, ParamDescription, ParamSprite, ParamTags, ParamLink}
	techNodeParams  = []string{ParamAlias, ParamLabel, ParamTechnology, ParamDescription, ParamSprite, ParamTags, ParamLink}
	plainNodeParams = []string{ParamAlias, ParamLabel, ParamDescription, ParamSprite, ParamTags, ParamLink}
	boundaryParams  = []string{ParamAlias, ParamLabel, ParamTags, ParamLink}
	genericBoundary = []string{ParamAlias, ParamLabel, ParamType, ParamTags, ParamLink}
)

// macros is the statement subset this package reads. A statement whose
// macro is not here is a parse error; growing the subset is an edit to
// this table.
var macros = buildMacros()

func buildMacros() map[string]macro {
	table := map[string]macro{
		"Person":     {kind: KindExternal, params: plainNodeParams, required: 2},
		"Person_Ext": {kind: KindExternal, params: plainNodeParams, required: 2},

		"Boundary":            {kind: KindBoundary, params: genericBoundary, required: 2},
		"Enterprise_Boundary": {kind: KindBoundary, params: boundaryParams, required: 2},
		"System_Boundary":     {kind: KindBoundary, params: boundaryParams, required: 2},
		"Container_Boundary":  {kind: KindBoundary, params: boundaryParams, required: 2},

		"UpdateElementStyle": {kind: KindIgnored},
		"UpdateRelStyle":     {kind: KindIgnored},
		"UpdateLayoutConfig": {kind: KindIgnored},
	}
	for _, level := range []struct {
		name   string
		params []string
	}{
		{"System", plainNodeParams},
		{"Container", techNodeParams},
		{"Component", techNodeParams},
	} {
		for _, form := range []string{"", "Db", "Queue"} {
			store := form != ""
			table[level.name+form] = macro{kind: KindElement, params: level.params, required: 2, store: store}
			table[level.name+form+"_Ext"] = macro{kind: KindExternal, params: level.params, required: 2, store: store}
		}
	}
	for _, rel := range []string{"Rel", "Rel_U", "Rel_Up", "Rel_D", "Rel_Down", "Rel_L", "Rel_Left", "Rel_R", "Rel_Right", "Rel_Back", "BiRel"} {
		table[rel] = macro{kind: KindRelationship, params: relParams, required: 3}
	}
	return table
}

// Parse reads the region of doc that at covers — the content of one
// fenced `mermaid` block — as a Mermaid diagram.
//
// The diagram's type is its first statement's first word. Only
// `C4Context`, `C4Container` and `C4Component` are parsed; for any other
// type the result carries the type name and no statements. A blank line,
// and one whose first non-blank characters are `%%`, is skipped.
func Parse(doc []byte, at span.Span) Diagram {
	start := int(at.Offset)
	end := start + int(at.Length)
	if start > len(doc) {
		start = len(doc)
	}
	if end > len(doc) {
		end = len(doc)
	}

	p := &parser{doc: doc}
	diagram := Diagram{TypeSpan: span.Of(doc, start, 0)}
	typed := false
	var open []int

	for lineStart := start; lineStart < end; {
		lineEnd := lineStart
		for lineEnd < end && doc[lineEnd] != '\n' {
			lineEnd++
		}
		next := lineEnd + 1

		from, to := trim(doc, lineStart, lineEnd)
		lineStart = next
		if from == to {
			continue
		}
		text := string(doc[from:to])
		if strings.HasPrefix(text, "%%") {
			continue
		}
		at := span.Of(doc, from, to-from)

		if !typed {
			typed = true
			diagram.Type = firstWord(text)
			diagram.TypeSpan = at
			if !diagram.Read() {
				return diagram
			}
			continue
		}

		if text == "}" {
			if len(open) == 0 {
				diagram.Errors = append(diagram.Errors, p.fail(at, text, "closes a boundary that is not open"))
				continue
			}
			open = open[:len(open)-1]
			continue
		}

		statement, err := p.statement(text, from, at)
		if err != nil {
			diagram.Errors = append(diagram.Errors, *err)
			if strings.HasSuffix(text, "{") {
				// A placeholder keeps the brace pairing intact, so the
				// matching } closes this line and not the enclosing boundary.
				open = append(open, -1)
			}
			continue
		}
		statement.Parent = -1
		for i := len(open) - 1; i >= 0; i-- {
			if open[i] >= 0 {
				statement.Parent = open[i]
				break
			}
		}
		diagram.Statements = append(diagram.Statements, statement)
		if statement.Kind == KindBoundary {
			open = append(open, len(diagram.Statements)-1)
		}
	}

	for _, index := range open {
		if index < 0 {
			continue // a failed line already carries its own error
		}
		boundary := diagram.Statements[index]
		diagram.Errors = append(diagram.Errors, p.fail(boundary.Span, string(span.Slice(doc, boundary.Span)), "opens a boundary that is never closed"))
	}
	sortErrors(diagram.Errors)
	return diagram
}

// sortErrors orders errors by position, so an unclosed boundary reported
// after the walk sits among the others in document order.
func sortErrors(errors []ParseError) {
	sort.SliceStable(errors, func(i, j int) bool { return errors[i].Span.Offset < errors[j].Span.Offset })
}

// trim narrows [from, to) of doc to exclude surrounding whitespace,
// including the carriage return a CRLF checkout leaves on every line.
func trim(doc []byte, from, to int) (int, int) {
	for from < to && isSpace(doc[from]) {
		from++
	}
	for to > from && isSpace(doc[to-1]) {
		to--
	}
	return from, to
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '\v' || b == '\f'
}

func firstWord(text string) string {
	if i := strings.IndexAny(text, " \t"); i >= 0 {
		return text[:i]
	}
	return text
}

type parser struct {
	doc []byte
}

func (p *parser) fail(at span.Span, text, reason string) ParseError {
	return ParseError{Span: at, Statement: text, Reason: reason}
}

// statement reads one trimmed, non-blank line that begins at offset base
// of the document.
func (p *parser) statement(text string, base int, at span.Span) (Statement, *ParseError) {
	failure := func(reason string) (Statement, *ParseError) {
		err := p.fail(at, text, reason)
		return Statement{}, &err
	}

	if text == "title" || strings.HasPrefix(text, "title ") || strings.HasPrefix(text, "title\t") {
		return Statement{Kind: KindIgnored, Macro: "title", Canonical: "title", Span: at}, nil
	}

	// No escape sequences exist, so an odd number of quotes cannot pair
	// up however the arguments are divided.
	if strings.Count(text, `"`)%2 != 0 {
		return failure("unbalanced quotes: an odd number of double quotes")
	}

	open := strings.IndexByte(text, '(')
	if open < 0 {
		return failure("is not a statement this parser reads: expected a macro followed by its arguments in parentheses")
	}
	name := strings.TrimRight(text[:open], " \t")
	row, known := macros[name]
	if !known {
		return failure("`" + name + "` is not a macro this parser reads")
	}

	args, named, closing, reason := p.arguments(text, base, open+1)
	if reason != "" {
		return failure(reason)
	}

	rest := strings.TrimSpace(text[closing+1:])
	switch {
	case row.kind == KindBoundary && rest != "{":
		return failure("a boundary statement must end with `{`")
	case row.kind != KindBoundary && rest != "":
		return failure("unexpected text after the closing parenthesis: " + rest)
	}

	statement := Statement{
		Kind:      row.kind,
		Macro:     name,
		Canonical: name,
		Span:      at,
		Args:      args,
		Named:     named,
		Store:     row.store,
		params:    row.params,
	}
	if row.kind != KindIgnored {
		supplied := 0
		for _, param := range row.params[:min(row.required, len(row.params))] {
			if _, ok := statement.Param(param); ok {
				supplied++
			}
		}
		if supplied < row.required {
			return failure("`" + name + "` needs at least " + strconv.Itoa(row.required) + " arguments, found " + strconv.Itoa(supplied))
		}
	}
	if row.kind == KindRelationship {
		statement.Relationship = relationship(&statement)
	}
	return statement, nil
}

// relationship normalises a relationship statement to dependency
// direction and reads its tags.
func relationship(s *Statement) *Relationship {
	from, _ := s.Param(ParamFrom)
	to, _ := s.Param(ParamTo)
	rel := &Relationship{Source: from, Target: to}
	switch s.Macro {
	case "BiRel":
		rel.Bidirectional = true
	case "Rel_Back":
		rel.Source, rel.Target = to, from
		s.Canonical = "Rel"
	default:
		s.Canonical = "Rel"
	}
	if tags, ok := s.Param(ParamTags); ok {
		rel.Tags = readTags(tags)
	}
	return rel
}

// readTags splits a tags argument on `+` and separates the reserved tags
// from the rest.
func readTags(arg Argument) Tags {
	tags := Tags{Present: true, Raw: arg.Value, Span: arg.Span, Comma: strings.Contains(arg.Value, ",")}
	for _, tag := range strings.Split(arg.Value, "+") {
		tag = strings.TrimSpace(tag)
		switch {
		case tag == "":
		case tag == TagSync || tag == TagAsync:
			tags.Reserved = append(tags.Reserved, ReservedTag{Name: tag})
		case strings.HasPrefix(tag, TagContract+":"):
			tags.Reserved = append(tags.Reserved, ReservedTag{Name: TagContract, Slug: strings.TrimPrefix(tag, TagContract+":")})
		default:
			tags.Unreserved = append(tags.Unreserved, tag)
		}
	}
	return tags
}

// arguments reads the comma-separated arguments of text beginning at
// index from, just past the opening parenthesis, and returns the index of
// the closing parenthesis. A non-empty reason reports why the arguments
// could not be read.
func (p *parser) arguments(text string, base, from int) (args, named []Argument, closing int, reason string) {
	i := from
	// An empty argument list is written `Macro()`.
	if j := skipSpace(text, i); j < len(text) && text[j] == ')' {
		return nil, nil, j, ""
	}
	for {
		i = skipSpace(text, i)
		arg := Argument{}

		if i < len(text) && text[i] == '$' {
			eq := strings.IndexByte(text[i:], '=')
			stop := strings.IndexAny(text[i:], ",()\"")
			if eq < 0 || (stop >= 0 && stop < eq) {
				return nil, nil, 0, "a named argument must be written `$<name>=<value>`"
			}
			arg.Name = strings.TrimSpace(text[i+1 : i+eq])
			if arg.Name == "" {
				return nil, nil, 0, "a named argument must be written `$<name>=<value>`"
			}
			i = skipSpace(text, i+eq+1)
		}

		var r string
		i, r = p.value(text, base, i, &arg)
		if r != "" {
			return nil, nil, 0, r
		}
		if arg.Name != "" {
			named = append(named, arg)
		} else {
			args = append(args, arg)
		}

		i = skipSpace(text, i)
		switch {
		case i >= len(text):
			return nil, nil, 0, "unbalanced parentheses: the argument list is never closed"
		case text[i] == ',':
			i++
		case text[i] == ')':
			return args, named, i, ""
		default:
			return nil, nil, 0, "unexpected `" + string(text[i]) + "` after an argument"
		}
	}
}

// value reads one argument's value at index i into arg and returns the
// index just past it.
func (p *parser) value(text string, base, i int, arg *Argument) (int, string) {
	if i < len(text) && text[i] == '"' {
		end := strings.IndexByte(text[i+1:], '"')
		if end < 0 {
			return 0, "unbalanced quotes: a quoted argument is never closed"
		}
		end += i + 1
		arg.Value = text[i+1 : end]
		arg.Quoted = true
		arg.Span = span.Of(p.doc, base+i, end+1-i)
		return end + 1, ""
	}

	depth := 0
	j := i
scan:
	for ; j < len(text); j++ {
		switch text[j] {
		case '"':
			return 0, "unbalanced quotes: a quote inside an unquoted argument"
		case '(':
			depth++
		case ')':
			if depth == 0 {
				break scan
			}
			depth--
		case ',':
			if depth == 0 {
				break scan
			}
		}
	}
	if j >= len(text) {
		return 0, "unbalanced parentheses: the argument list is never closed"
	}
	from, to := i, j
	for to > from && (text[to-1] == ' ' || text[to-1] == '\t') {
		to--
	}
	arg.Value = text[from:to]
	arg.Span = span.Of(p.doc, base+from, to-from)
	return j, ""
}

func skipSpace(text string, i int) int {
	for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
		i++
	}
	return i
}
