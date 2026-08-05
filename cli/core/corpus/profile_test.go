// Tests for the Profile extension point.
//
// No test here carries an `AC:` citation. The extension point is a
// foundational sliver lifted out of the CO delta by the Sequencing
// Constraints of specs/features/spec-index/spec-index-blueprint.md, and
// no acceptance criterion names it: AC-SPEC-008.6 is proven by the
// profile that implements this interface, and AC-SPEC-003.1 by the
// registry that consumes its schema. What is verified here is that the
// interface is satisfiable and that the host stays framework-agnostic.
package corpus_test

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

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// modulePath is this Go module's import path. Anything under
// modulePath + "/profile" is a profile package, and Document Core may not
// import one.
const modulePath = "github.com/brettcallaghan/agent-plugins/cli"

// stubProfile stands in for the real profile, which lands with the PROF
// item. It exists to prove the interface is satisfiable from outside the
// package that declares it, using nothing but data.
type stubProfile struct {
	schema      []byte
	types       map[string]corpus.DocumentTypeRules
	conventions specmodel.ConventionsConfig
}

func (p stubProfile) Schema() []byte                                     { return p.schema }
func (p stubProfile) DocumentTypes() map[string]corpus.DocumentTypeRules { return p.types }
func (p stubProfile) Conventions() specmodel.ConventionsConfig           { return p.conventions }

// The interface is satisfiable, checked at compile time rather than only
// when a test runs.
var _ corpus.Profile = stubProfile{}

func newStubProfile() stubProfile {
	width := uint64(72)
	return stubProfile{
		schema: []byte(`{"oneOf":[]}`),
		types: map[string]corpus.DocumentTypeRules{
			"numbered-thing": {SlugPattern: `^thing-[0-9]{3}$`, Numbered: true},
			"named-thing":    {SlugPattern: `^[a-z0-9-]+$`, Numbered: false},
		},
		conventions: specmodel.ConventionsConfig{LineWidth: &width},
	}
}

func TestProfileSuppliesOneUnionSchema(t *testing.T) {
	var profile corpus.Profile = newStubProfile()

	schema := profile.Schema()
	if len(schema) == 0 {
		t.Fatal("Schema returned nothing")
	}
	var parsed map[string]any
	if err := json.Unmarshal(schema, &parsed); err != nil {
		t.Fatalf("Schema is not JSON: %v", err)
	}
}

func TestProfileSuppliesASlugPatternAndNumberedFlagPerType(t *testing.T) {
	var profile corpus.Profile = newStubProfile()

	types := profile.DocumentTypes()
	if len(types) != 2 {
		t.Fatalf("DocumentTypes returned %d types, want 2", len(types))
	}

	numbered, ok := types["numbered-thing"]
	if !ok {
		t.Fatal(`DocumentTypes has no entry keyed "numbered-thing"`)
	}
	if !numbered.Numbered {
		t.Error("Numbered = false, want true")
	}
	if numbered.SlugPattern != `^thing-[0-9]{3}$` {
		t.Errorf("SlugPattern = %q", numbered.SlugPattern)
	}

	if named := types["named-thing"]; named.Numbered {
		t.Error("a type the profile did not register as numbered reports Numbered = true")
	}

	if _, ok := types["absent"]; ok {
		t.Error("DocumentTypes reports a type the profile never registered")
	}
}

func TestProfileSuppliesConventionDefaults(t *testing.T) {
	var profile corpus.Profile = newStubProfile()

	conventions := profile.Conventions()
	if conventions.LineWidth == nil {
		t.Fatal("LineWidth default is absent")
	}
	if *conventions.LineWidth != 72 {
		t.Errorf("LineWidth = %d, want 72", *conventions.LineWidth)
	}
}

// A profile that supplies no convention at all is still a valid profile:
// the CLI resolves such a key from its own fallback rather than treating
// the profile as malformed.
func TestProfileMaySupplyNoConventionDefaults(t *testing.T) {
	var profile corpus.Profile = stubProfile{}

	if width := profile.Conventions().LineWidth; width != nil {
		t.Errorf("LineWidth = %d, want absent", *width)
	}
}

// Document Core is the host, and a host that imports its plugin is not a
// host. The Change Guidance in
// specs/architecture/module-document-core.md states this as a
// prohibition; this test is what makes it a gate.
func TestDocumentCoreImportsNoProfilePackage(t *testing.T) {
	forbidden := modulePath + "/profile"

	for _, file := range goFilesUnder(t, coreDir(t)) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsing %s: %v", file, err)
		}
		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquoting import %s: %v", file, imported.Path.Value, err)
			}
			if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
				t.Errorf("%s imports %s; Document Core may not depend on a profile", file, path)
			}
		}
	}
}

// The extension point must not name a document type, because a host that
// knows one type by name is no longer framework-agnostic and an adopting
// repository with a different taxonomy cannot reuse it.
func TestTheExtensionPointNamesNoConcreteDocumentType(t *testing.T) {
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

// coreDir returns the cli/core directory this package sits under.
func coreDir(t *testing.T) string {
	t.Helper()

	dir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolving cli/core: %v", err)
	}
	return dir
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
// generated schema rather than restating them, so the check keeps working
// as the vocabulary grows.
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
