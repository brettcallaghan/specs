// Tests for the C4 Model package's element tree, over a small corpus
// built on disk.
//
// The corpus uses this framework's document types, because an index
// entry's frontmatter is the generated union and decodes only those; the
// architecture vocabulary itself is registered by the stub profile below
// and never assumed by the package under test. The profile is a stub
// rather than cli/profile, which Document Core may not import.
package c4model_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/schema"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

type stubProfile struct{}

func (stubProfile) Schema() []byte { return schema.SpecModel }
func (stubProfile) DocumentTypes() map[string]corpus.DocumentTypeRules {
	any := corpus.DocumentTypeRules{SlugPattern: `^[a-z0-9-]+$`}
	types := map[string]corpus.DocumentTypeRules{}
	for _, name := range []string{
		"system-blueprint", "container-blueprint", "module-blueprint",
		"component-blueprint", "feature-blueprint", "contract",
	} {
		types[name] = any
	}
	with := func(properties ...string) corpus.DocumentTypeRules {
		rules := any
		rules.ReferenceProperties = properties
		return rules
	}
	types["container-blueprint"] = with("system", "modules", "contracts_consumed")
	types["module-blueprint"] = with("modules", "contracts_consumed")
	types["component-blueprint"] = with("contracts_consumed")
	types["contract"] = with("provider")
	return types
}
func (stubProfile) Conventions() specmodel.ConventionsConfig { return specmodel.ConventionsConfig{} }
func (stubProfile) Lifecycle() corpus.LifecycleRules         { return corpus.LifecycleRules{} }
func (stubProfile) FeatureTags() corpus.FeatureTagRules      { return corpus.FeatureTagRules{} }
func (stubProfile) RecordBlocks() []corpus.RecordBlockTag {
	return []corpus.RecordBlockTag{{Tag: "component"}}
}
func (stubProfile) Architecture() corpus.ArchitectureRules {
	roster := func(level corpus.ElementLevel) corpus.ElementTypeRules {
		return corpus.ElementTypeRules{Level: level, RosterHeading: "Components", RosterTag: "component"}
	}
	return corpus.ArchitectureRules{
		Elements: map[string]corpus.ElementTypeRules{
			"system-blueprint":    {Level: corpus.LevelSystem},
			"container-blueprint": roster(corpus.LevelContainer),
			"module-blueprint":    roster(corpus.LevelModule),
			"component-blueprint": {Level: corpus.LevelComponent},
		},
		EntityNameProperty:        "entity_name",
		SystemProperty:            "system",
		ComposedModulesProperty:   "modules",
		ConsumedContractsProperty: "contracts_consumed",
		ContractType:              "contract",
		ProviderProperty:          "provider",
	}
}

// head is a document's frontmatter, with extra lines appended.
func head(slug, documentType, name string, extra ...string) string {
	lines := []string{
		"---",
		"title: " + slug,
		"description: " + slug + " blueprint",
		"slug: " + slug,
		"type: " + documentType,
		"status: draft",
	}
	if name != "" {
		lines = append(lines, "entity_name: "+name)
	}
	lines = append(lines, extra...)
	return strings.Join(append(lines, "---", ""), "\n")
}

func record(slug, name, description string) string {
	return "```component\nentity_name: " + name + "\nslug: " + slug + "\ndescription: " + description + "\n```\n"
}

// fixture is a corpus exercising every placement rule:
//
//   - system-one, holding container-alpha, which composes module-gamma,
//     consumes contract-api, provides it, and declares a component by
//     record block, a component blueprint by link, and a twice-declared
//     record-only component;
//   - container-beta, naming no system, composing module-gamma, and
//     declaring by record block a component a blueprint and module-gamma
//     also declare, and the twice-declared component again;
//   - container-delta, whose system property names a component;
//   - module-gamma, declaring one component of its own;
//   - component-lonely, a blueprint no roster declares;
//   - a diagram node naming nothing declared.
func fixture() map[string]string {
	return map[string]string{
		"sys.md": head("system-one", "system-blueprint", "One"),
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha",
			"technology: Go, cobra",
			"system: system-one",
			"modules:", `  - "module-gamma"`,
			"contracts_consumed:", `  - "contract-api"`,
			"tags:", `  - "cli"`) +
			"# Alpha\n\n## Components\n\n" +
			"```mermaid\nC4Component\nComponent(g, \"Ghost\", \"Go\")\n```\n" +
			"\n### Parser\n\n" + record("component-parser", "Parser", "Parses") +
			"\n### Twin\n\n" + record("component-twin", "Twin", "Twin from alpha") +
			"\n### Solo\n\nSee [[component-solo]].\n",
		"beta.md": head("container-beta", "container-blueprint", "Beta",
			"modules:", `  - "module-gamma"`) +
			"# Beta\n\n## Components\n\n" +
			"### Shared\n\n" + record("component-shared", "Shared", "Shared from a record") +
			"\n### Twin\n\n" + record("component-twin", "Twin", "Twin from beta"),
		"delta.md": head("container-delta", "container-blueprint", "Delta",
			"system: component-solo"),
		"gamma.md": head("module-gamma", "module-blueprint", "Gamma",
			"kind: standard", "visibility: internal", `version: "1.0"`) +
			"# Gamma\n\n## Components\n\n" +
			"### Own\n\n" + record("component-own", "Own", "Gamma's own") +
			"\n### Shared\n\n" + record("component-shared", "Shared", "Shared again"),
		"shared.md": head("component-shared", "component-blueprint", "Shared", "kind: standard"),
		"solo.md":   head("component-solo", "component-blueprint", "Solo"),
		"lonely.md": head("component-lonely", "component-blueprint", "Lonely"),
		"api.md":    head("contract-api", "contract", "", "provider: container-alpha"),
	}
}

func build(t *testing.T, files map[string]string) *specmodel.C4Model {
	t.Helper()
	model, err := buildRooted(t, files, "")
	if err != nil {
		t.Fatalf("C4Model: %v", err)
	}
	return model
}

// buildRooted builds the model over files narrowed to root, returning
// C4Model's error rather than failing on it.
func buildRooted(t *testing.T, files map[string]string, root string) (*specmodel.C4Model, error) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	built, err := corpus.Build(corpus.Options{Root: dir, Profile: stubProfile{}})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}
	return built.C4Model("test", root)
}

func system(t *testing.T, model *specmodel.C4Model, slug string) specmodel.C4System {
	t.Helper()
	for _, s := range model.SoftwareSystems {
		if string(s.Slug) == slug {
			return s
		}
	}
	t.Fatalf("no system %s", slug)
	return specmodel.C4System{}
}

func module(t *testing.T, model *specmodel.C4Model, slug string) specmodel.C4Module {
	t.Helper()
	for _, m := range model.Modules {
		if string(m.Slug) == slug {
			return m
		}
	}
	t.Fatalf("no module %s", slug)
	return specmodel.C4Module{}
}

func container(t *testing.T, model *specmodel.C4Model, slug string) specmodel.C4Container {
	t.Helper()
	for _, s := range model.SoftwareSystems {
		for _, c := range s.Containers {
			if string(c.Slug) == slug {
				return c
			}
		}
	}
	for _, c := range model.OrphanedContainers {
		if string(c.Slug) == slug {
			return c
		}
	}
	t.Fatalf("no container %s", slug)
	return specmodel.C4Container{}
}

// components returns where each component is placed: the parent's slug,
// or "orphaned".
func components(model *specmodel.C4Model) map[string][]string {
	placed := map[string][]string{}
	add := func(parent string, list []specmodel.C4Component) {
		for _, c := range list {
			placed[string(c.Slug)] = append(placed[string(c.Slug)], parent)
		}
	}
	for _, s := range model.SoftwareSystems {
		for _, c := range s.Containers {
			add(string(c.Slug), c.Components)
		}
	}
	for _, c := range model.OrphanedContainers {
		add(string(c.Slug), c.Components)
	}
	for _, m := range model.Modules {
		add(string(m.Slug), m.Components)
	}
	add("orphaned", model.OrphanedComponents)
	return placed
}

func component(t *testing.T, model *specmodel.C4Model, slug string) specmodel.C4Component {
	t.Helper()
	var all [][]specmodel.C4Component
	for _, s := range model.SoftwareSystems {
		for _, c := range s.Containers {
			all = append(all, c.Components)
		}
	}
	for _, c := range model.OrphanedContainers {
		all = append(all, c.Components)
	}
	for _, m := range model.Modules {
		all = append(all, m.Components)
	}
	all = append(all, model.OrphanedComponents)
	for _, list := range all {
		for _, c := range list {
			if string(c.Slug) == slug {
				return c
			}
		}
	}
	t.Fatalf("no component %s", slug)
	return specmodel.C4Component{}
}

func slugsOf[T ~string](items []T) string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = string(item)
	}
	return strings.Join(out, ",")
}

func componentSlugs(list []specmodel.C4Component) string {
	out := make([]string, len(list))
	for i, c := range list {
		out[i] = string(c.Slug)
	}
	return strings.Join(out, ",")
}

// AC: AC-C4-001.1
func TestContainersNestUnderTheSystemTheyDeclare(t *testing.T) {
	model := build(t, fixture())

	if got := len(model.SoftwareSystems); got != 1 {
		t.Fatalf("software_systems holds %d systems, want 1", got)
	}
	one := system(t, model, "system-one")
	var held []string
	for _, c := range one.Containers {
		held = append(held, string(c.Slug))
	}
	if got := strings.Join(held, ","); got != "container-alpha" {
		t.Errorf("system-one holds %q, want container-alpha", got)
	}
}

// AC: AC-C4-001.2
func TestAContainerThatDeclaresNoSystemIsOrphaned(t *testing.T) {
	model := build(t, fixture())

	var orphaned []string
	for _, c := range model.OrphanedContainers {
		orphaned = append(orphaned, string(c.Slug))
	}
	// container-delta's system property names a component, which is no
	// member at the system level.
	if got := strings.Join(orphaned, ","); got != "container-beta,container-delta" {
		t.Errorf("orphaned_containers = %q, want container-beta,container-delta", got)
	}
}

// AC: AC-C4-001.3
func TestComponentsNestUnderTheParentWhoseRosterDeclaresThem(t *testing.T) {
	model := build(t, fixture())
	placed := components(model)

	for slug, want := range map[string]string{
		"component-parser": "container-alpha", // record block
		"component-solo":   "container-alpha", // blueprint link
		"component-own":    "module-gamma",    // a module's roster
		// Declared by beta.md's and gamma.md's rosters: beta.md sorts
		// first, and the component is placed there and nowhere else.
		"component-shared": "container-beta",
	} {
		if got := strings.Join(placed[slug], ","); got != want {
			t.Errorf("%s placed under %q, want %q", slug, got, want)
		}
	}
	if got := componentSlugs(container(t, model, "container-alpha").Components); got != "component-parser,component-solo,component-twin" {
		t.Errorf("container-alpha's components = %q", got)
	}
}

// AC: AC-C4-001.4
func TestAModuleAppearsOnceAndItsHostsAreDerived(t *testing.T) {
	model := build(t, fixture())

	if got := len(model.Modules); got != 1 {
		t.Fatalf("modules holds %d modules, want 1", got)
	}
	for _, host := range []string{"container-alpha", "container-beta"} {
		if got := slugsOf(container(t, model, host).Modules); got != "module-gamma" {
			t.Errorf("%s modules = %q, want module-gamma", host, got)
		}
	}
	gamma := module(t, model, "module-gamma")
	if got := slugsOf(gamma.ContainedBy); got != "container-alpha,container-beta" {
		t.Errorf("module-gamma contained_by = %q, want container-alpha,container-beta", got)
	}
	if gamma.Kind == nil || *gamma.Kind != "standard" || gamma.Visibility == nil || *gamma.Visibility != "internal" ||
		gamma.Version == nil || *gamma.Version != "1.0" {
		t.Errorf("module-gamma kind/visibility/version = %v/%v/%v", gamma.Kind, gamma.Visibility, gamma.Version)
	}
}

// A module blueprint that omits kind or visibility carries no such
// property, and the model still marshals.
// AC: AC-C4-004.1
func TestAModuleWithAMissingKindOrVisibilityOmitsThem(t *testing.T) {
	for name, props := range map[string][]string{
		"missing kind":       {"visibility: internal"},
		"missing visibility": {"kind: standard"},
	} {
		t.Run(name, func(t *testing.T) {
			files := fixture()
			files["gamma.md"] = head("module-gamma", "module-blueprint", "Gamma", props...) + "# Gamma\n"
			gamma := module(t, build(t, files), "module-gamma")
			out, err := json.Marshal(gamma)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			var keys map[string]any
			if err := json.Unmarshal(out, &keys); err != nil {
				t.Fatal(err)
			}
			wantKind := name != "missing kind"
			if _, has := keys["kind"]; has != wantKind {
				t.Errorf("kind key present = %v for %s", has, name)
			}
			if _, has := keys["visibility"]; has == wantKind {
				t.Errorf("visibility key present = %v for %s", has, name)
			}
		})
	}
}

// A module blueprint whose kind or visibility is outside its enum is
// invalid frontmatter; building the model neither fails nor invents a
// value, and the model marshals.
// AC: AC-C4-004.1
func TestAModuleWithAnInvalidKindOrVisibilityDoesNotBreakTheModel(t *testing.T) {
	for name, props := range map[string][]string{
		"kind":       {"kind: bogus", "visibility: internal"},
		"visibility": {"kind: standard", "visibility: bogus"},
	} {
		t.Run(name, func(t *testing.T) {
			files := fixture()
			files["gamma.md"] = head("module-gamma", "module-blueprint", "Gamma", props...) + "# Gamma\n"
			model := build(t, files)
			if _, err := json.Marshal(model); err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
		})
	}
}

// AC: AC-C4-001.5
func TestAnElementListsEveryDeclaration(t *testing.T) {
	files := fixture()
	model := build(t, files)

	shared := component(t, model, "component-shared")
	var got []string
	for _, d := range shared.Declarations {
		got = append(got, d.Path)
	}
	if strings.Join(got, ",") != "beta.md,gamma.md,shared.md" {
		t.Fatalf("declarations = %v, want beta.md, gamma.md, shared.md", got)
	}
	for _, d := range shared.Declarations {
		text := files[d.Path][d.Span.Offset : d.Span.Offset+d.Span.Length]
		if d.Path == "shared.md" {
			if !strings.HasPrefix(text, "---\n") || !strings.HasSuffix(strings.TrimRight(text, "\n"), "---") {
				t.Errorf("shared.md declaration span = %q, want the frontmatter block", text)
			}
			continue
		}
		if !strings.HasPrefix(text, "```component\n") || !strings.HasSuffix(strings.TrimRight(text, "\n"), "```") {
			t.Errorf("%s declaration span = %q, want the record block with its fences", d.Path, text)
		}
	}
}

// AC: AC-C4-001.6
func TestAnElementWithABlueprintTakesItsPropertiesFromTheBlueprint(t *testing.T) {
	model := build(t, fixture())

	shared := component(t, model, "component-shared")
	if shared.Description != "component-shared blueprint" {
		t.Errorf("description = %q, want the blueprint's", shared.Description)
	}
	if shared.Kind == nil || *shared.Kind != "standard" {
		t.Errorf("kind = %v, want the blueprint's standard", shared.Kind)
	}
	if parser := component(t, model, "component-parser"); parser.Kind != nil {
		t.Errorf("a component with no blueprint carries kind %q", *parser.Kind)
	}
}

// AC: AC-C4-001.7
func TestElementPropertiesAreCopiedFromTheDeclaration(t *testing.T) {
	model := build(t, fixture())

	alpha := container(t, model, "container-alpha")
	if alpha.EntityName != "Alpha" || alpha.Description != "container-alpha blueprint" {
		t.Errorf("entity_name/description = %q/%q", alpha.EntityName, alpha.Description)
	}
	if alpha.Technology == nil || *alpha.Technology != "Go, cobra" {
		t.Errorf("technology = %v, want Go, cobra", alpha.Technology)
	}
	if alpha.ContractsConsumed == nil || slugsOf(*alpha.ContractsConsumed) != "contract-api" {
		t.Errorf("contracts_consumed = %v, want contract-api", alpha.ContractsConsumed)
	}
	if alpha.Tags == nil || strings.Join(*alpha.Tags, ",") != "cli" {
		t.Errorf("tags = %v, want cli", alpha.Tags)
	}

	// A property the declaration does not carry is absent.
	beta := container(t, model, "container-beta")
	if beta.Technology != nil || beta.ContractsConsumed != nil || beta.Tags != nil {
		t.Errorf("container-beta carries properties it does not declare: %v %v %v",
			beta.Technology, beta.ContractsConsumed, beta.Tags)
	}
	raw, err := json.Marshal(beta)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"technology"`) {
		t.Errorf("container-beta emits technology: %s", raw)
	}
}

// AC: AC-C4-001.8
func TestNoElementIsCreatedFromADiagramAlone(t *testing.T) {
	model := build(t, fixture())

	raw, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "Ghost") {
		t.Errorf("the model holds an element for the diagram node Ghost: %s", raw)
	}
}

// AC: AC-C4-001.9
func TestRecordBlocksInTwoDocumentsYieldOneElement(t *testing.T) {
	model := build(t, fixture())

	if got := components(model)["component-twin"]; strings.Join(got, ",") != "container-alpha" {
		t.Fatalf("component-twin placed under %v, want container-alpha once", got)
	}
	twin := component(t, model, "component-twin")
	var paths []string
	for _, d := range twin.Declarations {
		paths = append(paths, d.Path)
	}
	if strings.Join(paths, ",") != "alpha.md,beta.md" {
		t.Errorf("declarations = %v, want alpha.md, beta.md", paths)
	}
	if twin.Description != "Twin from alpha" {
		t.Errorf("description = %q, want the one from alpha.md, whose path sorts first", twin.Description)
	}
}

// AC: AC-C4-001.10
func TestProvidedContractsAreDerivedFromEachContractsProvider(t *testing.T) {
	model := build(t, fixture())

	if got := slugsOf(container(t, model, "container-alpha").ContractsProvided); got != "contract-api" {
		t.Errorf("container-alpha contracts_provided = %q, want contract-api", got)
	}
	if got := container(t, model, "container-beta").ContractsProvided; got == nil || len(got) != 0 {
		t.Errorf("container-beta contracts_provided = %v, want an empty list", got)
	}
}

// AC: AC-C4-001.11
func TestAComponentNoRosterDeclaresIsOrphaned(t *testing.T) {
	model := build(t, fixture())

	if got := componentSlugs(model.OrphanedComponents); got != "component-lonely" {
		t.Errorf("orphaned_components = %q, want component-lonely", got)
	}
}

// The model is the generated shape, with every collection present.
func TestTheModelCarriesItsVersionWorkspaceAndEmptyCollections(t *testing.T) {
	model := build(t, map[string]string{})

	if model.SchemaVersion != "0.1" || model.Workspace != "test" {
		t.Errorf("schema_version/workspace = %q/%q", model.SchemaVersion, model.Workspace)
	}
	raw, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var back specmodel.C4Model
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Errorf("the empty model does not read back as a C4Model: %v\n%s", err, raw)
	}
	if strings.Contains(string(raw), "null") {
		t.Errorf("the empty model emits null: %s", raw)
	}
}

// A composed slug resolves as the element set resolves it: a case variant
// of a module's slug emits that module's own slug, once; a slug naming no
// module is kept, trimmed, once per case-folded key.
// AC: AC-C4-001.4
func TestComposedModulesAreFoldedLikeTheRestOfTheModel(t *testing.T) {
	files := fixture()
	files["alpha.md"] = head("container-alpha", "container-blueprint", "Alpha",
		"system: system-one",
		"modules:", `  - "Module-Gamma"`, `  - "module-gamma"`, `  - "Ghost-Mod"`, `  - " ghost-mod "`)
	files["gamma.md"] = head("module-gamma", "module-blueprint", "Gamma",
		"kind: standard", "visibility: internal",
		"modules:", `  - "Module-Delta"`, `  - "module-delta"`) + "# Gamma\n"
	files["moddelta.md"] = head("module-delta", "module-blueprint", "Delta",
		"kind: standard", "visibility: internal") + "# Delta\n"
	model := build(t, files)

	if got := slugsOf(container(t, model, "container-alpha").Modules); got != "ghost-mod,module-gamma" {
		t.Errorf("container-alpha modules = %q, want module-gamma and ghost-mod once each, malformed spellings left out", got)
	}
	if got := slugsOf(module(t, model, "module-gamma").Modules); got != "module-delta" {
		t.Errorf("module-gamma modules = %q, want module-delta once", got)
	}
}

func TestAComposedModuleThatIsNotASlugIsLeftOut(t *testing.T) {
	files := fixture()
	files["alpha.md"] = head("container-alpha", "container-blueprint", "Alpha",
		"system: system-one",
		"modules:", `  - "Module Gamma"`, `  - "module-gamma"`)
	model := build(t, files)

	if got := slugsOf(container(t, model, "container-alpha").Modules); got != "module-gamma" {
		t.Errorf("container-alpha modules = %q, want only module-gamma", got)
	}
}
