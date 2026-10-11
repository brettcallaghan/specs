// Tests for the Schema Registry component.
//
// These run against a fixture union schema rather than the generated
// cli/core/schema/spec-model.schema.json bundle: the real profile that
// embeds that bundle lands with the PROF item, and this component's
// contract does not depend on which document types a profile happens to
// define. TestNoDocumentTypeInPackageSource is the one exception — it
// reads the real schema's DocumentType vocabulary, because what it
// guards against is this package's source naming a real type, not a
// fixture one.
package schemaregistry_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/finding"
	"github.com/brettcallaghan/specs/cli/core/frontmatter"
	"github.com/brettcallaghan/specs/cli/core/schemaregistry"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// fixtureSchema is a small union schema standing in for the real bundle
// a profile embeds. Its "definitions" are named the way the real
// generator names them — PascalCase, one per document type — so the
// PascalCase dispatch this package performs is exercised the same way
// it will be against the real bundle once the PROF item lands.
//
//   - Widget requires "name" and "released_at". "name" is pattern
//     constrained; "released_at" carries a format constraint, which is
//     what makes it the fixture for AC-SPEC-003.2.
//   - Gadget requires only "id", and exists to prove that one compiled
//     Registry dispatches correctly between more than one branch.
//   - BigWidget requires nothing and exists only to prove the
//     kebab-to-PascalCase transform handles more than one hyphen.
const fixtureSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "definitions": {
    "Widget": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "released_at"],
      "properties": {
        "name": {"type": "string", "pattern": "^[a-z]+$"},
        "released_at": {"type": "string", "format": "date-time"}
      }
    },
    "Gadget": {
      "type": "object",
      "required": ["id"],
      "properties": {
        "id": {"type": "string"}
      }
    },
    "BigWidget": {
      "type": "object",
      "properties": {}
    }
  }
}`

func mustRegistry(t *testing.T) *schemaregistry.Registry {
	t.Helper()
	registry, err := schemaregistry.New([]byte(fixtureSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return registry
}

func mustParse(t *testing.T, src string) *frontmatter.Document {
	t.Helper()
	doc, perr := frontmatter.Parse([]byte(src))
	if perr != nil {
		t.Fatalf("parsing fixture frontmatter: %s", perr.Message)
	}
	if doc == nil {
		t.Fatal("fixture carries no frontmatter block")
	}
	return doc
}

// A document satisfying its branch validates clean, and one compiled
// Registry serves more than one document type: the "one schema" a
// profile hands over is a union, not a registry keyed by type.
//
// AC: AC-SPEC-003.1
func TestValidateReportsNoFindingsForAConformingDocument(t *testing.T) {
	registry := mustRegistry(t)

	widget := mustParse(t, "---\n"+
		"name: crate\n"+
		"released_at: \"2026-07-29T00:00:00Z\"\n"+
		"---\n")
	findings, err := registry.Validate("widgets/crate.md", "widget", widget)
	if err != nil {
		t.Fatalf("Validate widget: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("Validate widget = %v, want no findings", findings)
	}

	gadget := mustParse(t, "---\nid: g-1\n---\n")
	findings, err = registry.Validate("gadgets/g-1.md", "gadget", gadget)
	if err != nil {
		t.Fatalf("Validate gadget: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("Validate gadget = %v, want no findings", findings)
	}
}

// Every violation in the document is reported in one pass, each with
// the JSON Pointer at which it occurred, rather than the check stopping
// at the first.
//
// AC: AC-SPEC-003.1
func TestValidateReportsEveryViolationWithItsPointer(t *testing.T) {
	registry := mustRegistry(t)

	// "name" violates its pattern; "released_at" is missing entirely.
	// Two independent defects, and both must come back.
	widget := mustParse(t, "---\nname: Not Lowercase\n---\n")
	findings, err := registry.Validate("widgets/bad.md", "widget", widget)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("Validate reported %d findings, want 2: %+v", len(findings), findings)
	}

	var sawPattern, sawMissing bool
	for _, f := range findings {
		switch {
		case f.Rule == finding.RuleSchemaViolation:
			violation, ok := f.Detail.AsSchemaViolation()
			if !ok {
				t.Errorf("schema-violation finding carries no FindingSchemaViolation detail: %+v", f)
				continue
			}
			if violation.Pointer != "/name" {
				t.Errorf("schema-violation pointer = %q, want %q", violation.Pointer, "/name")
			}
			if want := `pattern "^[a-z]+$"`; violation.Constraint != want {
				t.Errorf("schema-violation constraint = %q, want %q", violation.Constraint, want)
			}
			sawPattern = true
		case f.Rule == finding.RuleMissingProperty:
			missing, ok := f.Detail.AsMissingProperty()
			if !ok {
				t.Errorf("missing-property finding carries no FindingMissingProperty detail: %+v", f)
				continue
			}
			if missing.Property != "released_at" {
				t.Errorf("missing property = %q, want %q", missing.Property, "released_at")
			}
			sawMissing = true
		default:
			t.Errorf("unexpected rule %q", f.Rule)
		}
	}
	if !sawPattern {
		t.Error("no schema-violation finding for the pattern mismatch")
	}
	if !sawMissing {
		t.Error("no missing-property finding for released_at")
	}
}

// An undeclared property is reported once per offending key, each
// addressed at that property's own position with its own raw text as
// the value. The underlying library reports additionalProperties
// against the object as a whole, naming every offending key in one
// error; without this per-property fan-out, two different undeclared
// properties in the same document would be indistinguishable — both
// reported at the whole frontmatter block with no value at all — which
// fails the self-contained bar AC-SPEC-006.1 sets for every finding.
//
// AC: AC-SPEC-003.1
func TestValidateReportsOneFindingPerUndeclaredProperty(t *testing.T) {
	registry := mustRegistry(t)

	widget := mustParse(t, "---\n"+
		"name: crate\n"+
		"released_at: \"2026-07-29T00:00:00Z\"\n"+
		"extra: surplus\n"+
		"another: also-surplus\n"+
		"---\n")

	findings, err := registry.Validate("widgets/crate.md", "widget", widget)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("Validate reported %d findings, want 2: %+v", len(findings), findings)
	}

	extra, _ := widget.Property("extra")
	another, _ := widget.Property("another")
	wantSpan := map[string]specmodel.Span{"/extra": extra.Position, "/another": another.Position}
	wantValue := map[string]string{"/extra": "surplus", "/another": "also-surplus"}

	seen := map[string]bool{}
	for _, f := range findings {
		if f.Rule != finding.RuleSchemaViolation {
			t.Errorf("unexpected rule %q", f.Rule)
			continue
		}
		violation, ok := f.Detail.AsSchemaViolation()
		if !ok {
			t.Fatalf("finding carries no FindingSchemaViolation detail: %+v", f)
		}
		if violation.Constraint != "additionalProperties" {
			t.Errorf("constraint = %q, want %q", violation.Constraint, "additionalProperties")
		}
		if violation.Value != wantValue[violation.Pointer] {
			t.Errorf("%s value = %q, want %q", violation.Pointer, violation.Value, wantValue[violation.Pointer])
		}
		if f.Span != wantSpan[violation.Pointer] {
			t.Errorf("%s position = %+v, want the property's own span %+v",
				violation.Pointer, f.Span, wantSpan[violation.Pointer])
		}
		seen[violation.Pointer] = true
	}
	if !seen["/extra"] || !seen["/another"] {
		t.Errorf("saw pointers %v, want both /extra and /another", seen)
	}
}

// Where the branch constrains a property with a format keyword, the
// property's raw source text is validated in place of its YAML-coerced
// value. Writing digits with no punctuation is meant here as an author
// mistake, not a realistic date: goccy-yaml would parse it to a number,
// and if that coerced number were what got validated the failure would
// be a spurious type mismatch reporting a mangled numeric rendering of
// the value, not the format violation naming what the author actually
// typed.
//
// AC: AC-SPEC-003.2
func TestValidateChecksRawTextAgainstAFormatConstraint(t *testing.T) {
	registry := mustRegistry(t)

	widget := mustParse(t, "---\n"+
		"name: crate\n"+
		"released_at: 20260729\n"+
		"---\n")

	// Confirm the premise: goccy-yaml coerces the unquoted digits to a
	// number, not a string. If it did not, this test would not be
	// exercising the substitution at all.
	property, ok := widget.Property("released_at")
	if !ok {
		t.Fatal("fixture document has no released_at property")
	}
	if _, isString := property.Value.(string); isString {
		t.Fatalf("released_at Value = %#v (a string); this fixture needs a value YAML coerces away from string", property.Value)
	}
	if property.Raw != "20260729" {
		t.Fatalf("released_at Raw = %q, want %q", property.Raw, "20260729")
	}

	findings, err := registry.Validate("widgets/crate.md", "widget", widget)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("Validate reported %d findings, want 1: %+v", len(findings), findings)
	}

	violation, ok := findings[0].Detail.AsSchemaViolation()
	if !ok {
		t.Fatalf("finding carries no FindingSchemaViolation detail: %+v", findings[0])
	}
	if violation.Constraint != "format" {
		t.Errorf("constraint = %q, want %q (a coerced number would instead fail \"type\")", violation.Constraint, "format")
	}
	if violation.Value != "20260729" {
		t.Errorf("reported value = %q, want the exact source text %q, not a re-rendering of the coerced number", violation.Value, "20260729")
	}
}

// The PascalCase transform that selects a branch handles more than one
// hyphen, so a document type of arbitrary word count dispatches
// correctly without this package ever naming it.
func TestValidateDispatchesMultiWordDocumentTypes(t *testing.T) {
	registry := mustRegistry(t)

	doc := mustParse(t, "---\ntitle: anything\n---\n")
	findings, err := registry.Validate("things/x.md", "big-widget", doc)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("Validate = %v, want no findings against the permissive BigWidget branch", findings)
	}
}

// A document type the union schema defines no branch for is a caller
// error this package surfaces rather than papering over — Document
// Core's own unknown-type exclusion is expected to keep this
// unreachable for a real corpus walk.
func TestValidateReturnsAnErrorForAnUndefinedBranch(t *testing.T) {
	registry := mustRegistry(t)
	doc := mustParse(t, "---\nname: crate\n---\n")

	if _, err := registry.Validate("x.md", "unregistered-type", doc); err == nil {
		t.Error("Validate returned no error for a document type absent from the union schema")
	}
}

// CheckSlugPattern reports a mismatch naming both the actual slug and
// the expected pattern, and reports nothing when the slug matches.
//
// AC: AC-SPEC-003.5
func TestCheckSlugPatternReportsBothSlugAndPattern(t *testing.T) {
	pattern := `^widget-[0-9]{3}$`
	at := specmodel.Span{Offset: 4, Length: 6, StartLine: 2, StartColumn: 7, EndLine: 2, EndColumn: 13}

	f, err := schemaregistry.CheckSlugPattern("widgets/x.md", at, "widget-1", pattern)
	if err != nil {
		t.Fatalf("CheckSlugPattern: %v", err)
	}
	if f == nil {
		t.Fatal("CheckSlugPattern reported no finding for a mismatched slug")
	}
	if f.Rule != finding.RuleSlugPattern {
		t.Errorf("Rule = %q, want %q", f.Rule, finding.RuleSlugPattern)
	}
	slugPattern, ok := f.Detail.AsSlugPattern()
	if !ok {
		t.Fatalf("finding carries no FindingSlugPattern detail: %+v", *f)
	}
	if slugPattern.Slug != "widget-1" {
		t.Errorf("Slug = %q, want %q", slugPattern.Slug, "widget-1")
	}
	if slugPattern.Pattern != pattern {
		t.Errorf("Pattern = %q, want %q", slugPattern.Pattern, pattern)
	}

	clean, err := schemaregistry.CheckSlugPattern("widgets/x.md", at, "widget-007", pattern)
	if err != nil {
		t.Fatalf("CheckSlugPattern: %v", err)
	}
	if clean != nil {
		t.Errorf("CheckSlugPattern = %+v, want nil for a matching slug", *clean)
	}
}

// An invalid regular expression is a profile defect this package
// reports as an error rather than a Finding, since a Finding would
// blame the document instead of the pattern.
func TestCheckSlugPatternReturnsAnErrorForAnInvalidPattern(t *testing.T) {
	if _, err := schemaregistry.CheckSlugPattern("x.md", specmodel.Span{}, "anything", "("); err == nil {
		t.Error("CheckSlugPattern returned no error for an uncompilable pattern")
	}
}

// The extension point must not name a document type, because a
// component that knows one type by name is no longer framework-agnostic
// and an adopting repository with a different taxonomy cannot reuse it.
// specs/modules/document-core/document-core-module.md states this for every
// package under cli/core; cli/core/corpus carries the identical guard
// for its own source.
func TestNoDocumentTypeInPackageSource(t *testing.T) {
	vocabulary := documentTypeVocabulary(t)
	if len(vocabulary) == 0 {
		t.Fatal("read no document types from the generated schema; the check would pass vacuously")
	}

	for _, file := range goFilesUnder(t, ".") {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, file, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", file, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			if vocabulary[value] {
				t.Errorf("%s:%d names the document type %q",
					file, fileSet.Position(literal.Pos()).Line, value)
			}
			return true
		})
	}
}

func goFilesUnder(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	if len(files) == 0 {
		t.Fatalf("found no Go files under %s; the check would pass vacuously", root)
	}
	return files
}

// documentTypeVocabulary reads this framework's document types from the
// generated schema rather than restating them, so the check keeps
// working as the vocabulary grows.
func documentTypeVocabulary(t *testing.T) map[string]bool {
	t.Helper()

	raw, err := os.ReadFile(filepath.FromSlash("../schema/spec-model.schema.json"))
	if err != nil {
		t.Fatalf("reading the generated schema: %v", err)
	}

	var bundle struct {
		Definitions struct {
			DocumentType struct {
				Enum []string `json:"enum"`
			} `json:"DocumentType"`
		} `json:"definitions"`
	}
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatalf("parsing the generated schema: %v", err)
	}

	vocabulary := make(map[string]bool, len(bundle.Definitions.DocumentType.Enum))
	for _, value := range bundle.Definitions.DocumentType.Enum {
		vocabulary[value] = true
	}
	return vocabulary
}

// refSchema constrains two properties through $ref, the way the real
// bundle constrains slug and status through shared definitions: one by a
// pattern, one by an enum.
const refSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "definitions": {
    "Slug": {"type": "string", "pattern": "^[a-z][a-z0-9-]*$"},
    "Status": {"type": "string", "enum": ["draft", "approved", "retired"]},
    "Note": {
      "type": "object",
      "properties": {
        "slug": {"$ref": "#/definitions/Slug"},
        "status": {"$ref": "#/definitions/Status"}
      }
    }
  }
}`

// A violation reached through $ref is reported by the constraint it
// violated — the pattern, or every allowed value — rather than by the
// $ref keyword that led to it, and names the offending value.
//
// AC: AC-SPEC-003.1, AC-SPEC-006.1
func TestAViolationReachedThroughARefNamesItsLeafConstraint(t *testing.T) {
	registry, err := schemaregistry.New([]byte(refSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	doc := mustParse(t, "---\nslug: New_Thing\nstatus: actve\n---\n")
	findings, err := registry.Validate("notes/x.md", "note", doc)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("Validate reported %d findings, want 2: %+v", len(findings), findings)
	}

	want := map[string][]string{
		"/slug":   {`"New_Thing"`, `pattern "^[a-z][a-z0-9-]*$"`},
		"/status": {`"actve"`, "one of draft, approved, retired"},
	}
	for _, f := range findings {
		violation, ok := f.Detail.AsSchemaViolation()
		if !ok {
			t.Errorf("finding carries no schema-violation detail: %+v", f)
			continue
		}
		parts, ok := want[violation.Pointer]
		if !ok {
			t.Errorf("unexpected pointer %q: %+v", violation.Pointer, f)
			continue
		}
		delete(want, violation.Pointer)
		if strings.Contains(f.Message, "$ref") || strings.Contains(violation.Constraint, "$ref") {
			t.Errorf("finding names $ref: %q (constraint %q)", f.Message, violation.Constraint)
		}
		for _, part := range parts {
			if !strings.Contains(f.Message, part) {
				t.Errorf("message %q does not carry %s", f.Message, part)
			}
		}
	}
	for pointer := range want {
		t.Errorf("no finding at %s", pointer)
	}
}

// Enum answers the values a property's schema allows, resolving through
// $ref, and reports a property that is not an enum, or not declared, as
// having none.
//
// AC: AC-SPEC-006.1
func TestEnumResolvesAPropertysAllowedValuesThroughARef(t *testing.T) {
	registry, err := schemaregistry.New([]byte(refSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	values, ok, err := registry.Enum("note", "status")
	if err != nil || !ok || strings.Join(values, ",") != "draft,approved,retired" {
		t.Errorf("Enum(note, status) = %v, %v, %v; want draft, approved, retired", values, ok, err)
	}
	for _, property := range []string{"slug", "nope"} {
		if values, ok, err := registry.Enum("note", property); err != nil || ok {
			t.Errorf("Enum(note, %s) = %v, %v, %v; want no enum", property, values, ok, err)
		}
	}
}
