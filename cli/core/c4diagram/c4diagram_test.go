package c4diagram_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/c4diagram"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

// parseBlock parses the content of the one mermaid block in doc.
func parseBlock(t *testing.T, doc string) c4diagram.Diagram {
	t.Helper()
	blocks := markdown.Fenced([]byte(doc), []string{"mermaid"})
	if len(blocks) != 1 {
		t.Fatalf("found %d mermaid blocks, want 1", len(blocks))
	}
	return c4diagram.Parse([]byte(doc), blocks[0].Inner)
}

// diagram wraps lines in a C4Component mermaid block, behind a heading
// so the block does not start on the document's first line.
func diagram(lines ...string) string {
	return "# Doc\n\n```mermaid\nC4Component\n" + strings.Join(lines, "\n") + "\n```\n"
}

func only(t *testing.T, d c4diagram.Diagram) c4diagram.Statement {
	t.Helper()
	if len(d.Errors) != 0 {
		t.Fatalf("unexpected parse errors: %+v", d.Errors)
	}
	if len(d.Statements) != 1 {
		t.Fatalf("got %d statements, want 1: %+v", len(d.Statements), d.Statements)
	}
	return d.Statements[0]
}

func param(t *testing.T, s c4diagram.Statement, name string) string {
	t.Helper()
	arg, ok := s.Param(name)
	if !ok {
		t.Fatalf("%s has no %s argument", s.Macro, name)
	}
	return arg.Value
}

// AC: AC-C4-DIAG-001.2
func TestAQuotedArgumentKeepsItsCommasAndParentheses(t *testing.T) {
	d := parseBlock(t, diagram(`Component(x, "X", "Go package (planned), internal", "desc")`))

	s := only(t, d)
	if s.Kind != c4diagram.KindElement {
		t.Errorf("Kind = %q, want %q", s.Kind, c4diagram.KindElement)
	}
	if got := param(t, s, c4diagram.ParamTechnology); got != "Go package (planned), internal" {
		t.Errorf("technology = %q, want the whole quoted text", got)
	}
	if got := param(t, s, c4diagram.ParamDescription); got != "desc" {
		t.Errorf("description = %q, want desc", got)
	}
	if got := param(t, s, c4diagram.ParamAlias); got != "x" {
		t.Errorf("alias = %q, want x", got)
	}
}

// AC: AC-C4-DIAG-004.1
func TestPositionalAndNamedTagsAreReadAlike(t *testing.T) {
	for _, statement := range []string{
		`Rel(a, b, "Uses", "HTTP", "", "", "async+audited")`,
		`Rel(a, b, "Uses", "HTTP", $tags="async+audited")`,
	} {
		t.Run(statement, func(t *testing.T) {
			s := only(t, parseBlock(t, diagram(statement)))
			tags := s.Relationship.Tags
			if !tags.Present {
				t.Fatal("tags argument not found")
			}
			if want := []c4diagram.ReservedTag{{Name: c4diagram.TagAsync}}; !reflect.DeepEqual(tags.Reserved, want) {
				t.Errorf("reserved = %+v, want %+v", tags.Reserved, want)
			}
			if want := []string{"audited"}; !reflect.DeepEqual(tags.Unreserved, want) {
				t.Errorf("unreserved = %q, want %q", tags.Unreserved, want)
			}
			if tags.Comma {
				t.Error("Comma set on a tags argument holding no comma")
			}
		})
	}
}

func TestReservedTagsSeparateTheContractSlug(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`Rel(a, b, "Uses", $tags="sync+contract:contract-orders+cached")`)))

	want := []c4diagram.ReservedTag{{Name: c4diagram.TagSync}, {Name: c4diagram.TagContract, Slug: "contract-orders"}}
	if got := s.Relationship.Tags.Reserved; !reflect.DeepEqual(got, want) {
		t.Errorf("reserved = %+v, want %+v", got, want)
	}
	if got := s.Relationship.Tags.Unreserved; !reflect.DeepEqual(got, []string{"cached"}) {
		t.Errorf("unreserved = %q, want [cached]", got)
	}
}

func TestACommaInTagsIsRecordedAndNotASeparator(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`Rel(a, b, "Uses", $tags="async,audited")`)))

	tags := s.Relationship.Tags
	if !tags.Comma {
		t.Error("Comma not recorded")
	}
	if len(tags.Reserved) != 0 {
		t.Errorf("reserved = %+v, want none: async,audited is one tag", tags.Reserved)
	}
	if want := []string{"async,audited"}; !reflect.DeepEqual(tags.Unreserved, want) {
		t.Errorf("unreserved = %q, want %q", tags.Unreserved, want)
	}
}

func TestRelBackIsNormalisedToDependencyDirection(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`Rel_Back(a, b, "Serves")`)))

	if s.Relationship.Source.Value != "b" || s.Relationship.Target.Value != "a" {
		t.Errorf("source, target = %q, %q; want b, a", s.Relationship.Source.Value, s.Relationship.Target.Value)
	}
	if s.Macro != "Rel_Back" || s.Canonical != "Rel" {
		t.Errorf("macro, canonical = %q, %q; want Rel_Back, Rel", s.Macro, s.Canonical)
	}
}

func TestDirectionalVariantsAreRelKeepingTheMacroAsWritten(t *testing.T) {
	for _, macro := range []string{"Rel_U", "Rel_Up", "Rel_D", "Rel_Down", "Rel_L", "Rel_Left", "Rel_R", "Rel_Right"} {
		s := only(t, parseBlock(t, diagram(macro+`(a, b, "Uses")`)))
		if s.Canonical != "Rel" || s.Macro != macro {
			t.Errorf("%s: macro, canonical = %q, %q", macro, s.Macro, s.Canonical)
		}
		if s.Relationship.Source.Value != "a" || s.Relationship.Target.Value != "b" || s.Relationship.Bidirectional {
			t.Errorf("%s: relationship = %+v, want a to b, one way", macro, s.Relationship)
		}
	}
}

func TestBiRelIsFlaggedBidirectional(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`BiRel(a, b, "Talks to")`)))

	if s.Kind != c4diagram.KindRelationship || !s.Relationship.Bidirectional {
		t.Errorf("kind = %q, bidirectional = %v; want a bidirectional relationship", s.Kind, s.Relationship.Bidirectional)
	}
}

func TestStatementsAreClassified(t *testing.T) {
	d := parseBlock(t, diagram(
		`title Components`,
		`Person(user, "User")`,
		`System_Ext(ext, "External")`,
		`ContainerDb(db, "Store", "Postgres")`,
		`ComponentQueue_Ext(q, "Queue")`,
		`Component(c, "C")`,
		`UpdateLayoutConfig($c4ShapeInRow="3")`,
		`Rel(c, db, "Reads")`,
	))
	if len(d.Errors) != 0 {
		t.Fatalf("unexpected errors: %+v", d.Errors)
	}
	type row struct {
		kind  c4diagram.Kind
		store bool
	}
	want := []row{
		{c4diagram.KindIgnored, false},
		{c4diagram.KindExternal, false},
		{c4diagram.KindExternal, false},
		{c4diagram.KindElement, true},
		{c4diagram.KindExternal, true},
		{c4diagram.KindElement, false},
		{c4diagram.KindIgnored, false},
		{c4diagram.KindRelationship, false},
	}
	var got []row
	for _, s := range d.Statements {
		got = append(got, row{s.Kind, s.Store})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("kinds = %+v\nwant    %+v", got, want)
	}
}

func TestSystemAndPersonTakeNoTechnology(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`System(s, "S", "What it does")`)))

	if got := param(t, s, c4diagram.ParamDescription); got != "What it does" {
		t.Errorf("description = %q, want the third argument", got)
	}
	if _, ok := s.Param(c4diagram.ParamTechnology); ok {
		t.Error("a System node reported a technology")
	}
}

func TestBoundariesOpenAndClose(t *testing.T) {
	d := parseBlock(t, diagram(
		`Container_Boundary(outer, "Outer") {`,
		`    Component(a, "A")`,
		`    Boundary(inner, "Inner", "group") {`,
		`        Component(b, "B")`,
		`    }`,
		`    Component(c, "C")`,
		`}`,
		`Component(d, "D")`,
	))
	if len(d.Errors) != 0 {
		t.Fatalf("unexpected errors: %+v", d.Errors)
	}
	parents := map[string]int{}
	index := map[string]int{}
	for i, s := range d.Statements {
		alias := param(t, s, c4diagram.ParamAlias)
		parents[alias] = s.Parent
		index[alias] = i
	}
	want := map[string]int{
		"outer": -1,
		"a":     index["outer"],
		"inner": index["outer"],
		"b":     index["inner"],
		"c":     index["outer"],
		"d":     -1,
	}
	if !reflect.DeepEqual(parents, want) {
		t.Errorf("parents = %v, want %v", parents, want)
	}
	if d.Statements[index["outer"]].Kind != c4diagram.KindBoundary {
		t.Error("Container_Boundary not classified as a boundary")
	}
}

func TestBoundaryErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		lines  []string
		reason string
	}{
		"close with nothing open": {[]string{`}`}, "not open"},
		"never closed":            {[]string{`System_Boundary(s, "S") {`, `Component(a, "A")`}, "never closed"},
		"no opening brace":        {[]string{`System_Boundary(s, "S")`}, "must end with `{`"},
		"brace on a node":         {[]string{`Component(a, "A") {`}, "unexpected text"},
	} {
		t.Run(name, func(t *testing.T) {
			d := parseBlock(t, diagram(tc.lines...))
			if len(d.Errors) != 1 {
				t.Fatalf("got %d errors, want 1: %+v", len(d.Errors), d.Errors)
			}
			if !strings.Contains(d.Errors[0].Reason, tc.reason) {
				t.Errorf("reason = %q, want it to mention %q", d.Errors[0].Reason, tc.reason)
			}
		})
	}
}

// A failed line ending in { still holds a place on the boundary stack, so
// its } does not close the enclosing boundary.
func TestAFailedOpeningLineKeepsTheBracePairing(t *testing.T) {
	for name, inner := range map[string]string{
		"unknown macro": `Deployment_Node(n, "N") {`,
		"bad arguments": `Container_Boundary(x) {`,
	} {
		t.Run(name, func(t *testing.T) {
			d := parseBlock(t, diagram(
				`Container_Boundary(o, "O") {`,
				inner,
				`Component(a, "A")`,
				`}`,
				`Component(b, "B")`,
				`}`,
			))
			if len(d.Errors) != 1 {
				t.Fatalf("got %d errors, want 1: %+v", len(d.Errors), d.Errors)
			}
			if len(d.Statements) != 3 {
				t.Fatalf("got %d statements, want 3", len(d.Statements))
			}
			for _, s := range d.Statements[1:] {
				if s.Parent != 0 {
					t.Errorf("%s parent = %d, want 0", param(t, s, c4diagram.ParamAlias), s.Parent)
				}
			}
		})
	}
}

func TestAFailedOpeningLineNeverClosedAddsNoSecondError(t *testing.T) {
	d := parseBlock(t, diagram(`Deployment_Node(n, "N") {`, `Component(a, "A")`))
	if len(d.Errors) != 1 {
		t.Fatalf("got %d errors, want 1: %+v", len(d.Errors), d.Errors)
	}
	if d.Statements[0].Parent != -1 {
		t.Errorf("parent = %d, want -1", d.Statements[0].Parent)
	}
}

// A parse error carries the span of its statement, in the coordinates of
// the whole document, and parsing continues with the next line.
func TestParseErrorsCarryTheStatementSpanAndParsingContinues(t *testing.T) {
	for name, tc := range map[string]struct {
		statement string
		reason    string
	}{
		"unknown macro":        {`Deployment_Node(n, "Node")`, "not a macro"},
		"not a macro call":     {`a --> b`, "not a statement"},
		"unbalanced quote":     {`Component(a, "A, "Go")`, "unbalanced quotes"},
		"unbalanced paren":     {`Component(a, "A"`, "unbalanced parentheses"},
		"extra closing paren":  {`Component(a, "A"))`, "unexpected text"},
		"too few arguments":    {`Rel(a, b)`, "at least 3"},
		"too few for a node":   {`Container(a)`, "at least 2"},
		"malformed named":      {`Rel(a, b, "x", $tags)`, "named argument"},
		"quote inside unquote": {`Component(a, A"b)`, "unbalanced quotes"},
	} {
		t.Run(name, func(t *testing.T) {
			doc := diagram("    "+tc.statement, `Component(ok, "OK")`)
			d := parseBlock(t, doc)
			if len(d.Errors) != 1 {
				t.Fatalf("got %d errors, want 1: %+v", len(d.Errors), d.Errors)
			}
			err := d.Errors[0]
			if !strings.Contains(err.Reason, tc.reason) {
				t.Errorf("reason = %q, want it to mention %q", err.Reason, tc.reason)
			}
			if got := string(span.Slice([]byte(doc), err.Span)); got != tc.statement {
				t.Errorf("span slices to %q, want %q", got, tc.statement)
			}
			if err.Statement != tc.statement {
				t.Errorf("Statement = %q, want %q", err.Statement, tc.statement)
			}
			// Line 1 heading, 2 blank, 3 fence, 4 C4Component, 5 the statement.
			if err.Span.StartLine != 5 || err.Span.StartColumn != 5 {
				t.Errorf("span starts at %d:%d, want 5:5", err.Span.StartLine, err.Span.StartColumn)
			}
			if len(d.Statements) != 1 || d.Statements[0].Macro != "Component" {
				t.Errorf("parsing did not continue to the next line: %+v", d.Statements)
			}
		})
	}
}

func TestStatementAndArgumentSpansAreInDocumentCoordinates(t *testing.T) {
	doc := diagram(`  Component(api, "API — gateway", "Go")`)
	s := only(t, parseBlock(t, doc))

	if got := string(span.Slice([]byte(doc), s.Span)); got != `Component(api, "API — gateway", "Go")` {
		t.Errorf("statement span slices to %q", got)
	}
	label, _ := s.Param(c4diagram.ParamLabel)
	if got := string(span.Slice([]byte(doc), label.Span)); got != `"API — gateway"` {
		t.Errorf("label span slices to %q", got)
	}
	tech, _ := s.Param(c4diagram.ParamTechnology)
	// The em dash is one column, not three bytes.
	if want := uint64(len(`  Component(api, "API — gateway", `)) - 1; tech.Span.StartColumn != want {
		t.Errorf("technology starts at column %d, want %d", tech.Span.StartColumn, want)
	}
}

func TestBlankLinesAndCommentsAreSkipped(t *testing.T) {
	d := parseBlock(t, "```mermaid\n%% leading comment\n\nC4Context\n\n   %% a comment\n\nPerson(p, \"P\")\n```\n")

	if d.Type != c4diagram.TypeContext {
		t.Errorf("Type = %q, want C4Context: a comment before the header is not the first statement", d.Type)
	}
	only(t, d)
}

func TestCRLFLinesParse(t *testing.T) {
	doc := "```mermaid\r\nC4Container\r\nContainer(c, \"C\", \"Go\")\r\n```\r\n"
	s := only(t, parseBlock(t, doc))
	if got := string(span.Slice([]byte(doc), s.Span)); got != `Container(c, "C", "Go")` {
		t.Errorf("span slices to %q: the carriage return must be trimmed", got)
	}
}

func TestOtherDiagramTypesReturnTheirNameOnly(t *testing.T) {
	for _, header := range []string{"C4Dynamic", "C4Deployment", "flowchart LR", "sequenceDiagram"} {
		d := parseBlock(t, "```mermaid\n"+header+"\n    a --> b\n```\n")
		if want := strings.Fields(header)[0]; d.Type != want {
			t.Errorf("Type = %q, want %q", d.Type, want)
		}
		if d.Read() || len(d.Statements) != 0 || len(d.Errors) != 0 {
			t.Errorf("%s: read = %v, %d statements, %d errors; want the type alone", header, d.Read(), len(d.Statements), len(d.Errors))
		}
	}
}

func TestAnEmptyBlockHasNoType(t *testing.T) {
	d := parseBlock(t, "```mermaid\n\n```\n")
	if d.Type != "" || d.Read() {
		t.Errorf("Type = %q, Read = %v; want empty and unread", d.Type, d.Read())
	}
}

func TestNamedArgumentsTakeNoPositionalSlot(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`Container(c, "C", $descr="Described", $link="https://x")`)))

	if len(s.Args) != 2 {
		t.Errorf("got %d positional arguments, want 2", len(s.Args))
	}
	if got := param(t, s, c4diagram.ParamDescription); got != "Described" {
		t.Errorf("description = %q, want the $descr value", got)
	}
	if _, ok := s.Param(c4diagram.ParamTechnology); ok {
		t.Error("a technology was read from a named argument's slot")
	}
}

// A required parameter given by name counts as supplied, and a named
// argument takes no positional slot.
func TestRequiredParametersMayBeGivenByName(t *testing.T) {
	s := only(t, parseBlock(t, diagram(`Rel(a, b, $label="Uses")`)))
	if got := param(t, s, c4diagram.ParamLabel); got != "Uses" {
		t.Errorf("label = %q, want Uses", got)
	}
	s = only(t, parseBlock(t, diagram(`Rel(a, $to=b, "x")`)))
	if got := param(t, s, c4diagram.ParamTo); got != "b" {
		t.Errorf("to = %q, want the $to value", got)
	}
	if got := param(t, s, c4diagram.ParamLabel); got != "x" {
		t.Errorf("label = %q, want the positional argument after the named slot", got)
	}
	only(t, parseBlock(t, diagram(`Component($alias=x, $label="X")`)))
}

func TestMissingRequiredParametersStillFail(t *testing.T) {
	for name, line := range map[string]string{
		"two positional":      `Rel(a, b)`,
		"unrelated named arg": `Rel(a, b, $sprite=s)`,
	} {
		t.Run(name, func(t *testing.T) {
			d := parseBlock(t, diagram(line))
			if len(d.Errors) != 1 || !strings.Contains(d.Errors[0].Reason, "needs at least 3 arguments, found 2") {
				t.Fatalf("errors = %+v, want one 'needs at least 3 arguments, found 2'", d.Errors)
			}
		})
	}
}

// Leading frontmatter and multi-line directives are skipped, not read as
// the diagram type.
func TestFrontmatterAndMultiLineDirectivesAreSkipped(t *testing.T) {
	doc := "# Doc\n\n```mermaid\n---\ntitle: X\n---\nC4Component\nComponent(a, \"A\")\n```\n"
	d := parseBlock(t, doc)
	if d.Type != "C4Component" || !d.Read() {
		t.Fatalf("type = %q, read = %v", d.Type, d.Read())
	}
	only(t, d)
	if got := string(span.Slice([]byte(doc), d.TypeSpan)); got != "C4Component" {
		t.Errorf("type span covers %q", got)
	}

	doc = "# Doc\n\n```mermaid\n%%{\ninit: {\"theme\": \"dark\"}\n}%%\nC4Container\n```\n"
	if d = parseBlock(t, doc); d.Type != "C4Container" {
		t.Errorf("type after a multi-line directive = %q", d.Type)
	}

	d = parseBlock(t, diagram(`%%{`, `init: {"theme": "dark"}`, `}%%`, `Component(a, "A")`))
	only(t, d)
}

func TestUnterminatedFrontmatterYieldsNoDiagram(t *testing.T) {
	d := parseBlock(t, "# Doc\n\n```mermaid\n---\ntitle: X\nC4Component\nComponent(a, \"A\")\n```\n")
	if d.Type != "" || len(d.Statements) != 0 || len(d.Errors) != 0 {
		t.Errorf("got type %q, %d statements, %d errors", d.Type, len(d.Statements), len(d.Errors))
	}
}

func TestUnterminatedDirectiveConsumesTheRestWithoutError(t *testing.T) {
	d := parseBlock(t, diagram(`%%{`, `init: {}`, `Component(a, "A")`))
	if len(d.Errors) != 0 || len(d.Statements) != 0 {
		t.Errorf("got %d statements, %+v", len(d.Statements), d.Errors)
	}
}

func TestContractSlugIsTrimmed(t *testing.T) {
	for tags, slug := range map[string]string{"contract: c-1": "c-1", "contract:": "", "contract:   ": ""} {
		s := only(t, parseBlock(t, diagram(`Rel(a, b, "Uses", $tags="`+tags+`")`)))
		want := []c4diagram.ReservedTag{{Name: c4diagram.TagContract, Slug: slug}}
		if got := s.Relationship.Tags.Reserved; !reflect.DeepEqual(got, want) {
			t.Errorf("%q: reserved = %+v, want %+v", tags, got, want)
		}
	}
}

// Every C4 diagram this repository draws parses with no error.
func TestThisRepositorysDiagramsParse(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	files := []string{
		"specs/containers/spec-cli/spec-cli-container.md",
		"specs/modules/document-core/document-core-module.md",
		"specs/modules/spec-profile/spec-profile-module.md",
	}
	diagrams := 0
	for _, file := range files {
		doc, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range markdown.Fenced(doc, []string{"mermaid"}) {
			d := c4diagram.Parse(doc, block.Inner)
			if !d.Read() {
				continue
			}
			diagrams++
			for _, e := range d.Errors {
				t.Errorf("%s:%d: %s: %s", file, e.Span.StartLine, e.Statement, e.Reason)
			}
			if len(d.Statements) == 0 {
				t.Errorf("%s: a %s diagram read no statements", file, d.Type)
			}
		}
	}
	if diagrams < 4 {
		t.Errorf("read %d C4 diagrams, want at least 4", diagrams)
	}
}
