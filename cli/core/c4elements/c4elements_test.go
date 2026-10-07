// Tests for the C4 Elements package, over small corpora built on disk.
//
// The corpora use this framework's document types, because an index
// entry's frontmatter is the generated union and decodes only those; the
// architecture vocabulary itself is registered by the stub profile below
// and never assumed by the package under test. The profile is a stub
// rather than cli/profile, which Document Core may not import.
package c4elements_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/c4elements"
	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

type stubProfile struct {
	architecture corpus.ArchitectureRules
}

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
	container := any
	container.ReferenceProperties = []string{"modules"}
	types["container-blueprint"] = container
	return types
}
func (stubProfile) Conventions() specmodel.ConventionsConfig { return specmodel.ConventionsConfig{} }
func (stubProfile) Lifecycle() corpus.LifecycleRules         { return corpus.LifecycleRules{} }
func (stubProfile) FeatureTags() corpus.FeatureTagRules      { return corpus.FeatureTagRules{} }
func (stubProfile) RecordBlocks() []corpus.RecordBlockTag {
	return []corpus.RecordBlockTag{{Tag: "component"}}
}
func (p stubProfile) Architecture() corpus.ArchitectureRules { return p.architecture }

func vocabulary() corpus.ArchitectureRules {
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
		"description: test document",
		"slug: " + slug,
		"type: " + documentType,
		"status: draft",
		"updated_at: 2026-10-07T00:00:00Z",
	}
	if name != "" {
		lines = append(lines, "entity_name: "+name)
	}
	lines = append(lines, extra...)
	return strings.Join(append(lines, "---", ""), "\n")
}

func record(slug, name string) string {
	return "```component\nentity_name: " + name + "\nslug: " + slug + "\ndescription: d\n```\n"
}

func diagram(lines ...string) string {
	return "```mermaid\nC4Component\n" + strings.Join(lines, "\n") + "\n```\n"
}

// fixture is a corpus exercising every rule: two containers each
// declaring a component named Parser; one container composing a module
// and listing a component blueprint by reference under its roster; a
// module whose lower-case roster heading declares a component a
// blueprint also declares; a record block outside any roster; and a
// feature blueprint whose record block and diagram declare nothing.
func fixture() map[string]string {
	return map[string]string{
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha",
			"modules:", `  - "module-gamma"`) +
			"# Alpha\n\n## Overview\n\n" + record("component-stray", "Stray") +
			"\n## Components\n\n" +
			diagram(
				`Component(p, "Parser", "Go")`,
				`Component(g, "Gamma", "Go")`,
				`Component(s, "Solo", "Go")`,
				`Component(x, "Stray", "Go")`,
				`ContainerDb(db, "Store", "Files")`,
				`ContainerDb(alpha, "Alpha", "Files")`,
				`Person(u, "User")`,
				`Component_Ext(e, "Parser", "Go")`,
			) +
			"\n### Parser\n\n" + record("component-alpha-parser", "Parser"),
		"beta.md": head("container-beta", "container-blueprint", "Beta") +
			"# Beta\n\n## Components\n\n" +
			diagram(`Component(p, "Parser", "Go")`, `Component(s, "Solo", "Go")`, `Container(a, "Alpha", "Go")`) +
			"\n### Parser\n\n" + record("component-beta-parser", "Parser") +
			"\n### Solo\n\nSee [[component-solo]].\n",
		"gamma.md": head("module-gamma", "module-blueprint", "Gamma") +
			"# Gamma\n\n## components\n\n### Shared\n\n" + record("component-shared", "Shared"),
		"shared.md": head("component-shared", "component-blueprint", "Shared"),
		"solo.md":   head("component-solo", "component-blueprint", "Solo"),
		"feature.md": head("feature-delta", "feature-blueprint", "") +
			"# Delta\n\n## Components\n\n" + record("component-delta", "Delta") +
			diagram(`Component(d, "Delta", "Go")`),
	}
}

func build(t *testing.T, files map[string]string, architecture corpus.ArchitectureRules) *c4elements.Set {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	built, err := corpus.Build(corpus.Options{Root: dir, Profile: stubProfile{architecture: architecture}})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}
	set, err := built.C4Elements()
	if err != nil {
		t.Fatalf("C4Elements: %v", err)
	}
	return set
}

func slugs(elements []*c4elements.Element) []string {
	var out []string
	for _, element := range elements {
		out = append(out, element.Slug)
	}
	return out
}

func document(t *testing.T, set *c4elements.Set, path string) *c4elements.Document {
	t.Helper()
	for _, doc := range set.Documents() {
		if doc.Path == path {
			return doc
		}
	}
	t.Fatalf("no drawing document %s", path)
	return nil
}

// nodes maps each node's alias to the slugs it resolved to, with
// "external" and "store" standing for the two unresolved outcomes.
func nodes(t *testing.T, doc *c4elements.Document) map[string]string {
	t.Helper()
	if len(doc.Diagrams) != 1 {
		t.Fatalf("%s has %d diagrams, want 1", doc.Path, len(doc.Diagrams))
	}
	out := map[string]string{}
	for _, node := range doc.Diagrams[0].Nodes {
		switch {
		case node.External:
			out[node.Alias] = "external"
		case node.PassiveStore:
			out[node.Alias] = "store"
		default:
			out[node.Alias] = strings.Join(slugs(node.Matches), ",")
		}
	}
	return out
}

func TestTheElementSetHoldsBlueprintsAndRosterRecords(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	want := "component-alpha-parser,component-beta-parser,component-shared,component-solo," +
		"container-alpha,container-beta,module-gamma"
	if got := strings.Join(slugs(set.Elements()), ","); got != want {
		t.Errorf("elements = %s\nwant       %s", got, want)
	}
	alpha, _ := set.Element("CONTAINER-ALPHA")
	if alpha == nil || alpha.Name != "Alpha" || alpha.Level != corpus.LevelContainer || alpha.Path != "alpha.md" {
		t.Errorf("container-alpha = %+v", alpha)
	}
	parser, _ := set.Element("component-alpha-parser")
	if parser == nil || parser.Name != "Parser" || parser.Level != corpus.LevelComponent || parser.Path != "" {
		t.Errorf("component-alpha-parser = %+v", parser)
	}
}

func TestABlueprintAndARecordBlockMergeUnderOneSlug(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	shared, ok := set.Element("component-shared")
	if !ok {
		t.Fatal("component-shared is not an element")
	}
	if len(shared.Declarations) != 2 {
		t.Fatalf("declarations = %+v, want two", shared.Declarations)
	}
	if shared.Declarations[0].Path != "gamma.md" || shared.Declarations[1].Path != "shared.md" {
		t.Errorf("declarations are not ordered by path: %+v", shared.Declarations)
	}
	if shared.Path != "shared.md" || shared.Level != corpus.LevelComponent {
		t.Errorf("the blueprint does not define the element: %+v", shared)
	}
	if shared.Declarations[1].Span.Offset != 0 {
		t.Errorf("a blueprint's declaration is its frontmatter block, got %+v", shared.Declarations[1].Span)
	}
}

// The module's roster heading is written "components"; it matches the
// registered "Components" through the heading identifier.
func TestTheRosterHeadingIsMatchedWithoutRegardToCase(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	shared, _ := set.Element("component-shared")
	if shared == nil || shared.Declarations[0].Path != "gamma.md" {
		t.Fatalf("the lower-case roster's record block declared nothing: %+v", shared)
	}
}

func TestARecordBlockOutsideARosterDeclaresNothing(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	if _, ok := set.Element("component-stray"); ok {
		t.Error("a record block under Overview declared an element")
	}
	if _, ok := set.Element("component-delta"); ok {
		t.Error("a record block in a type with no roster declared an element")
	}
	if got := nodes(t, document(t, set, "alpha.md"))["x"]; got != "" {
		t.Errorf("Stray resolved to %q, want nothing", got)
	}
}

// Two containers each declare a component named Parser, and a node
// drawn in either resolves to its own.
func TestANodeResolvesWithinItsDrawingDocumentsScope(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	alpha := nodes(t, document(t, set, "alpha.md"))
	beta := nodes(t, document(t, set, "beta.md"))
	if alpha["p"] != "component-alpha-parser" {
		t.Errorf("alpha's Parser = %q", alpha["p"])
	}
	if beta["p"] != "component-beta-parser" {
		t.Errorf("beta's Parser = %q", beta["p"])
	}
	// Gamma is in alpha's scope through its composed modules.
	if alpha["g"] != "module-gamma" {
		t.Errorf("alpha's Gamma = %q", alpha["g"])
	}
	// Solo is listed by reference under beta's roster, not alpha's.
	if beta["s"] != "component-solo" || alpha["s"] != "" {
		t.Errorf("Solo: beta %q, alpha %q", beta["s"], alpha["s"])
	}
	// Every container-level element is in every document's scope.
	if beta["a"] != "container-alpha" {
		t.Errorf("beta's Alpha = %q", beta["a"])
	}

	want := "component-beta-parser,component-solo,container-alpha,container-beta"
	if got := strings.Join(slugs(document(t, set, "beta.md").Scope), ","); got != want {
		t.Errorf("beta's scope = %s, want %s", got, want)
	}
}

func TestPassiveStoresAndExternalNodes(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	alpha := nodes(t, document(t, set, "alpha.md"))
	if alpha["db"] != "store" {
		t.Errorf("an unmatched Db node = %q, want a passive store", alpha["db"])
	}
	if alpha["alpha"] != "container-alpha" {
		t.Errorf("a Db node naming an element = %q, want the element", alpha["alpha"])
	}
	if alpha["u"] != "external" || alpha["e"] != "external" {
		t.Errorf("external nodes = %q and %q", alpha["u"], alpha["e"])
	}
}

func TestOnlyRegisteredTypesDraw(t *testing.T) {
	set := build(t, fixture(), vocabulary())

	for _, doc := range set.Documents() {
		if doc.Path == "feature.md" {
			t.Error("a member whose type declares no element is a drawing document")
		}
	}
	if got := len(set.Documents()); got != 5 {
		t.Errorf("drawing documents = %d, want 5", got)
	}
}

func TestAProfileRegisteringNoArchitectureHasNoElements(t *testing.T) {
	set := build(t, fixture(), corpus.ArchitectureRules{})

	if len(set.Elements()) != 0 || len(set.Documents()) != 0 {
		t.Errorf("elements %v, documents %d; want none", slugs(set.Elements()), len(set.Documents()))
	}
}

// A non-C4 Mermaid block is kept with its type, so a check can name it.
func TestADiagramTheParserDoesNotReadIsKept(t *testing.T) {
	files := map[string]string{
		"beta.md": head("container-beta", "container-blueprint", "Beta") +
			"# Beta\n\n```mermaid\nflowchart LR\n  a --> b\n```\n",
	}
	set := build(t, files, vocabulary())

	doc := document(t, set, "beta.md")
	if len(doc.Diagrams) != 1 || doc.Diagrams[0].Parsed.Type != "flowchart" || doc.Diagrams[0].Parsed.Read() {
		t.Errorf("diagrams = %+v", doc.Diagrams)
	}
	if len(doc.Diagrams[0].Nodes) != 0 {
		t.Errorf("an unread diagram has nodes: %+v", doc.Diagrams[0].Nodes)
	}
}
