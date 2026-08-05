package specindex_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// schemaPath is the generated bundle scripts/generate-schemas.sh emits
// from schemas/*.strata. The tests below read it rather than restating
// its contents, so a change to the Strata source that loosens the model
// fails here instead of passing unnoticed.
const schemaPath = "../schema/spec-model.schema.json"

// AC: AC-SPEC-002.2
func TestNewDeclaresTheSchemaVersion(t *testing.T) {
	index := specindex.New("agent-plugins")

	if index.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", index.SchemaVersion, "1.0")
	}
	if specindex.SchemaVersion != "1.0" {
		t.Errorf("specindex.SchemaVersion = %q, want %q", specindex.SchemaVersion, "1.0")
	}
	if index.Workspace != "agent-plugins" {
		t.Errorf("Workspace = %q, want %q", index.Workspace, "agent-plugins")
	}
}

// AC: AC-SPEC-002.2
func TestEmittedIndexCarriesTheSchemaVersion(t *testing.T) {
	encoded, err := json.Marshal(specindex.New("agent-plugins"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	version, ok := decoded["schema_version"]
	if !ok {
		t.Fatalf("emitted index has no schema_version property: %s", encoded)
	}
	if string(version) != `"1.0"` {
		t.Errorf("schema_version = %s, want %q", version, "1.0")
	}
}

// An empty index must emit empty collections rather than nulls, so that
// output stays comparable across platforms and a consumer can iterate
// without a nil check.
func TestNewEmitsEmptyCollectionsRatherThanNull(t *testing.T) {
	encoded, err := json.Marshal(specindex.New("agent-plugins"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	want := `{"schema_version":"1.0","workspace":"agent-plugins","entries":[],"exclusions":[]}`
	if string(encoded) != want {
		t.Errorf("emitted index =\n  %s\nwant\n  %s", encoded, want)
	}
}

// The schema version is required and pinned, so an index that omits it
// or declares a different value is not a valid index at all.
//
// AC: AC-SPEC-002.2
func TestSchemaRequiresAndPinsTheSchemaVersion(t *testing.T) {
	definition := loadDefinition(t, "SpecIndex")

	if !isRequired(definition, "schema_version") {
		t.Errorf("SpecIndex does not require schema_version; required = %v", definition["required"])
	}

	property := propertyOf(t, definition, "schema_version")
	pattern, _ := property["pattern"].(string)
	if pattern == "" {
		t.Fatalf("schema_version declares no pattern; it must be pinned")
	}

	matcher := regexp.MustCompile(pattern)
	if !matcher.MatchString(specindex.SchemaVersion) {
		t.Errorf("pattern %q rejects the emitted version %q", pattern, specindex.SchemaVersion)
	}
	for _, rejected := range []string{"", "1", "1.1", "2.0", "v1.0", "1.0.0"} {
		if matcher.MatchString(rejected) {
			t.Errorf("pattern %q accepts %q, but only %q is a valid version this sprint",
				pattern, rejected, specindex.SchemaVersion)
		}
	}
}

// The index root declares exactly the properties this sprint pins: four
// required ones — the version, the workspace, and the two ordered
// collections — and the optional input hash the warm cache is decided by.
//
// The hash is optional because absence is meaningful: an index that
// records no hash was not built by walking a corpus, and is never served
// warm.
//
// AC: AC-SPEC-002.3
func TestSchemaDeclaresTheIndexRootShape(t *testing.T) {
	definition := loadDefinition(t, "SpecIndex")

	required := []string{"schema_version", "workspace", "entries", "exclusions"}
	want := append(append([]string{}, required...), "input_hash")
	if got := propertyNames(t, definition); !reflect.DeepEqual(got, want) {
		t.Errorf("SpecIndex properties = %v, want %v", got, want)
	}
	for _, name := range required {
		if !isRequired(definition, name) {
			t.Errorf("SpecIndex does not require %q", name)
		}
	}
	if isRequired(definition, "input_hash") {
		t.Errorf("SpecIndex requires input_hash; an index built from no corpus carries none")
	}
	if additional, ok := definition["additionalProperties"].(bool); !ok || additional {
		t.Errorf("SpecIndex accepts additional properties; the index shape must be closed")
	}
}

// A member carries its repository-relative path and its parsed
// frontmatter, and the whole index round-trips through JSON unchanged.
//
// AC: AC-SPEC-002.1, AC-SPEC-002.3
func TestIndexCarriesEachMembersPathAndFrontmatter(t *testing.T) {
	index := specindex.New("agent-plugins")
	index.Entries = append(index.Entries, specmodel.IndexEntry{
		Path: "specs/architecture/module-document-core.md",
		Frontmatter: &specmodel.DocumentFrontmatter{
			Title:       "Module Blueprint: Document Core",
			Description: "The framework-agnostic engine.",
			Slug:        specmodel.Slug("module-document-core"),
			Type:        specmodel.DocumentTypeModuleBlueprint,
			Status:      specmodel.StatusApproved,
			UpdatedAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	})

	entry := index.Entries[0]
	if entry.Path != "specs/architecture/module-document-core.md" {
		t.Errorf("Path = %q", entry.Path)
	}
	if entry.Frontmatter == nil {
		t.Fatal("Frontmatter is nil")
	}
	if entry.Frontmatter.Slug != specmodel.Slug("module-document-core") {
		t.Errorf("Slug = %q", entry.Frontmatter.Slug)
	}
	if entry.ParseError != nil {
		t.Errorf("ParseError = %+v, want nil for a member that parsed", entry.ParseError)
	}

	assertRoundTrips(t, index)
}

// A member whose frontmatter could not be parsed is still indexed: it
// carries a located parse error in place of an envelope, and every other
// member is unaffected.
//
// AC: AC-SPEC-002.4
func TestIndexRecordsALocatedParseFailureAgainstAMember(t *testing.T) {
	index := specindex.New("agent-plugins")
	index.Entries = append(index.Entries, specmodel.IndexEntry{
		Path: "plugin/skills/groom-beads/references/work-order-template.md",
		ParseError: &specmodel.ParseError{
			Message: "mapping values are not allowed in this context",
			Position: specmodel.Span{
				Offset:      512,
				Length:      17,
				StartLine:   28,
				StartColumn: 3,
				EndLine:     28,
				EndColumn:   20,
			},
		},
	})

	entry := index.Entries[0]
	if entry.Frontmatter != nil {
		t.Errorf("Frontmatter = %+v, want nil for a member that failed to parse", entry.Frontmatter)
	}
	if entry.ParseError == nil {
		t.Fatal("ParseError is nil")
	}
	if entry.ParseError.Position.Offset != 512 || entry.ParseError.Position.Length != 17 {
		t.Errorf("Position extent = (%d, %d), want (512, 17)",
			entry.ParseError.Position.Offset, entry.ParseError.Position.Length)
	}

	assertRoundTrips(t, index)
}

// An excluded file is recorded with its reason, and an unknown declared
// type is reported verbatim so a reviewing agent can judge whether the
// profile or the document is wrong.
//
// AC: AC-SPEC-001.3, AC-SPEC-001.4
func TestIndexRecordsExclusionsWithTheirReason(t *testing.T) {
	declared := "work-order"
	index := specindex.New("agent-plugins")
	index.Exclusions = append(index.Exclusions,
		specmodel.IndexExclusion{
			Path:   "docs/session-bootstrap.md",
			Reason: specmodel.ExclusionReasonNoFrontmatter,
		},
		specmodel.IndexExclusion{
			Path:   "plugin/skills/groom-beads/references/work-order.md",
			Reason: specmodel.ExclusionReasonUnknownType,
			Type:   &declared,
		},
	)

	if index.Exclusions[0].Type != nil {
		t.Errorf("Type = %v, want nil when the reason is no-frontmatter", *index.Exclusions[0].Type)
	}
	if got := index.Exclusions[1].Type; got == nil || *got != declared {
		t.Errorf("Type = %v, want %q", got, declared)
	}

	assertRoundTrips(t, index)
}

// The two reason values are exactly the vocabulary REQ-SPEC-001 names,
// and nothing else decodes.
//
// AC: AC-SPEC-001.3, AC-SPEC-001.4
func TestExclusionReasonVocabularyIsClosed(t *testing.T) {
	for _, accepted := range []string{"no-frontmatter", "unknown-type"} {
		var reason specmodel.ExclusionReason
		if err := json.Unmarshal([]byte(`"`+accepted+`"`), &reason); err != nil {
			t.Errorf("decoding %q: %v", accepted, err)
		}
	}

	var reason specmodel.ExclusionReason
	if err := json.Unmarshal([]byte(`"not-a-reason"`), &reason); err == nil {
		t.Error("decoding an unregistered reason succeeded; the vocabulary must be closed")
	}
}

// assertRoundTrips encodes the index and decodes it back, so that a
// property the emitter writes but a consumer cannot read is caught here
// rather than in the consumer.
func assertRoundTrips(t *testing.T, index specmodel.SpecIndex) {
	t.Helper()

	encoded, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded specmodel.SpecIndex
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(index, decoded) {
		t.Errorf("round trip changed the index:\n got %+v\nwant %+v", decoded, index)
	}
}

// loadDefinition returns one named definition from the generated bundle.
func loadDefinition(t *testing.T, name string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.FromSlash(schemaPath))
	if err != nil {
		t.Fatalf("reading %s: %v", schemaPath, err)
	}

	var bundle struct {
		Definitions map[string]map[string]any `json:"definitions"`
	}
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatalf("parsing %s: %v", schemaPath, err)
	}

	definition, ok := bundle.Definitions[name]
	if !ok {
		t.Fatalf("%s declares no definition %q", schemaPath, name)
	}
	return definition
}

func propertyOf(t *testing.T, definition map[string]any, name string) map[string]any {
	t.Helper()

	properties, ok := definition["properties"].(map[string]any)
	if !ok {
		t.Fatalf("definition declares no properties object")
	}
	property, ok := properties[name].(map[string]any)
	if !ok {
		t.Fatalf("definition declares no property %q", name)
	}
	return property
}

// propertyNames returns the definition's property names in the order the
// schema lists them as required, followed by any that are not required.
// The generated bundle emits `properties` as an alphabetised JSON object,
// so `required` is where declaration order survives.
func propertyNames(t *testing.T, definition map[string]any) []string {
	t.Helper()

	properties, ok := definition["properties"].(map[string]any)
	if !ok {
		t.Fatalf("definition declares no properties object")
	}

	names := make([]string, 0, len(properties))
	seen := make(map[string]bool, len(properties))
	for _, required := range requiredNames(definition) {
		if _, ok := properties[required]; !ok {
			t.Errorf("definition requires %q but declares no such property", required)
			continue
		}
		names = append(names, required)
		seen[required] = true
	}
	for name := range properties {
		if !seen[name] {
			names = append(names, name)
		}
	}
	return names
}

func requiredNames(definition map[string]any) []string {
	required, _ := definition["required"].([]any)
	names := make([]string, 0, len(required))
	for _, entry := range required {
		if name, ok := entry.(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func isRequired(definition map[string]any, name string) bool {
	for _, required := range requiredNames(definition) {
		if required == name {
			return true
		}
	}
	return false
}
