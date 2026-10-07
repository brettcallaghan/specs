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
	container.ReferenceProperties = []string{"modules", "contracts_consumed"}
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

// relationshipFixture draws relationships of every outcome: in alpha, a
// component diagram; in beta, a container diagram drawing one of alpha's
// relationships again; and two containers both named Twin, so a node
// naming Twin is ambiguous everywhere.
func relationshipFixture() map[string]string {
	return map[string]string{
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha") +
			"# Alpha\n\n## Components\n\n" +
			diagram(
				`Component(p, "Parser", "Go")`,
				`Component(l, "Lexer", "Go")`,
				`Component(x, "Ghost", "Go")`,
				`ContainerDb(db, "Store", "Files")`,
				`Person(u, "User")`,
				`Container(a, "Alpha", "Go")`,
				`Container(b, "Beta", "Go")`,
				`Container(t, "Twin", "Go")`,
				`Rel(p, l, "Tokenizes with", "Go", "streams tokens")`,
				`Rel_Back(l, p, "Tokenizes with")`,
				`Rel(u, p, "Uses")`,
				`Rel(p, db, "Writes")`,
				`Rel(p, nobody, "Calls")`,
				`Rel(p, x, "Calls")`,
				`Rel(t, p, "Calls")`,
				`Rel(t, u, "Mixed")`,
				`BiRel(p, l, "Talks")`,
				`Rel(a, b, "Calls", "HTTP", "", $tags="sync+contract:contract-api")`,
			) +
			"\n### Parser\n\n" + record("component-parser", "Parser") +
			"\n### Lexer\n\n" + record("component-lexer", "Lexer"),
		"beta.md": head("container-beta", "container-blueprint", "Beta") +
			"# Beta\n\n```mermaid\nC4Container\n" +
			"Container(a, \"Alpha\", \"Go\")\nContainer(b, \"Beta\", \"Go\")\n" +
			"Rel(a, b, \"Calls\", \"gRPC\")\n```\n",
		"twin1.md": head("container-twin-one", "container-blueprint", "Twin"),
		"twin2.md": head("container-twin-two", "container-blueprint", "Twin"),
	}
}

func TestRelationshipStatementsResolveToDrawings(t *testing.T) {
	set := build(t, relationshipFixture(), vocabulary())
	diagram := document(t, set, "alpha.md").Diagrams[0]

	var got []string
	for _, drawing := range diagram.Drawings {
		got = append(got, drawing.Source.Slug+">"+drawing.Target.Slug+":"+drawing.Label)
	}
	want := "component-parser>component-lexer:Tokenizes with," +
		"component-parser>component-lexer:Tokenizes with," +
		"container-alpha>container-beta:Calls"
	if strings.Join(got, ",") != want {
		t.Errorf("drawings = %v\nwant       %s", got, want)
	}

	first := diagram.Drawings[0]
	if first.Technology != "Go" || first.Description != "streams tokens" ||
		first.View != specmodel.C4ViewComponent || first.Path != "alpha.md" ||
		first.SourceAlias != "p" || first.TargetAlias != "l" {
		t.Errorf("first drawing = %+v", first)
	}
	last := diagram.Drawings[2]
	if len(last.Tags.Reserved) != 2 || last.Tags.Reserved[1].Slug != "contract-api" {
		t.Errorf("tags = %+v", last.Tags)
	}
}

// Rel_Back(l, p) draws p depending on l, and its aliases are reported in
// that direction.
func TestRelBackIsReadInDependencyDirection(t *testing.T) {
	set := build(t, relationshipFixture(), vocabulary())
	back := document(t, set, "alpha.md").Diagrams[0].Drawings[1]

	if back.Source.Slug != "component-parser" || back.Target.Slug != "component-lexer" {
		t.Errorf("Rel_Back drew %s > %s", back.Source.Slug, back.Target.Slug)
	}
	if back.SourceAlias != "p" || back.TargetAlias != "l" {
		t.Errorf("aliases = %s > %s", back.SourceAlias, back.TargetAlias)
	}
}

func TestUnresolvedStatementsCarryOneReasonInOrder(t *testing.T) {
	set := build(t, relationshipFixture(), vocabulary())
	unresolved := document(t, set, "alpha.md").Diagrams[0].Unresolved

	type outcome struct {
		reason specmodel.C4ExclusionReason
		source bool
	}
	want := []outcome{
		{specmodel.C4ExclusionReasonExternalEndpoint, true}, // Rel(u, p)
		{specmodel.C4ExclusionReasonPassiveStore, false},    // Rel(p, db)
		{specmodel.C4ExclusionReasonUndeclaredNode, false},  // Rel(p, nobody): no node
		{specmodel.C4ExclusionReasonUndeclaredNode, false},  // Rel(p, x): Ghost names nothing
		{specmodel.C4ExclusionReasonAmbiguousNode, true},    // Rel(t, p): two Twins
		// Rel(t, u): the source is ambiguous, but an external endpoint is
		// checked first, on either end.
		{specmodel.C4ExclusionReasonExternalEndpoint, false},
	}
	if len(unresolved) != len(want) {
		t.Fatalf("unresolved = %+v, want %d", unresolved, len(want))
	}
	for i, w := range want {
		if unresolved[i].Reason != w.reason || unresolved[i].SourceEndpoint != w.source {
			t.Errorf("unresolved %d (%s > %s) = %s source=%v, want %s source=%v",
				i, unresolved[i].SourceAlias, unresolved[i].TargetAlias,
				unresolved[i].Reason, unresolved[i].SourceEndpoint, w.reason, w.source)
		}
	}
}

func TestBiRelResolvesToNoDrawing(t *testing.T) {
	set := build(t, relationshipFixture(), vocabulary())
	diagram := document(t, set, "alpha.md").Diagrams[0]

	for _, drawing := range diagram.Drawings {
		if drawing.Label == "Talks" {
			t.Error("BiRel resolved to a drawing")
		}
	}
	for _, unresolved := range diagram.Unresolved {
		if unresolved.SourceAlias == "p" && unresolved.TargetAlias == "l" {
			t.Error("BiRel is reported unresolved")
		}
	}
}

// Alpha's and Beta's diagrams both draw Alpha calling Beta: one
// relationship with two drawings, ordered by path, each keeping its own
// technology and view.
func TestDrawingsGroupIntoRelationshipsAcrossDocuments(t *testing.T) {
	set := build(t, relationshipFixture(), vocabulary())

	var calls *c4elements.Relationship
	for i, relationship := range set.Relationships() {
		if relationship.Source.Slug == "container-alpha" && relationship.Label == "Calls" {
			calls = &set.Relationships()[i]
		}
	}
	if calls == nil {
		t.Fatalf("no Alpha calls Beta relationship among %+v", set.Relationships())
	}
	if len(calls.Drawings) != 2 || calls.Drawings[0].Path != "alpha.md" || calls.Drawings[1].Path != "beta.md" {
		t.Fatalf("drawings = %+v", calls.Drawings)
	}
	if calls.Drawings[0].Technology != "HTTP" || calls.Drawings[1].Technology != "gRPC" {
		t.Errorf("technologies = %q, %q", calls.Drawings[0].Technology, calls.Drawings[1].Technology)
	}
	if calls.Drawings[1].View != specmodel.C4ViewContainer {
		t.Errorf("beta's view = %s", calls.Drawings[1].View)
	}
	if got := len(set.Relationships()); got != 2 {
		t.Errorf("relationships = %d, want 2 (Parser>Lexer and Alpha>Beta)", got)
	}
}

func TestAnElementWithNoEntityNameIsNamedByNoNode(t *testing.T) {
	files := map[string]string{
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha") +
			"# Alpha\n\n```mermaid\nC4Container\n" +
			"Container(a, \"Alpha\", \"Go\")\nContainer(x, \"\", \"Go\")\n" +
			"Rel(a, x, \"Calls\")\n```\n",
		"nameless1.md": head("container-nameless-one", "container-blueprint", ""),
	}
	for name, extra := range map[string]map[string]string{
		"one nameless element":  {},
		"two nameless elements": {"nameless2.md": head("container-nameless-two", "container-blueprint", "")},
	} {
		t.Run(name, func(t *testing.T) {
			all := map[string]string{}
			for k, v := range files {
				all[k] = v
			}
			for k, v := range extra {
				all[k] = v
			}
			set := build(t, all, vocabulary())
			diagram := document(t, set, "alpha.md").Diagrams[0]
			for _, node := range diagram.Nodes {
				if node.Alias == "x" && len(node.Matches) != 0 {
					t.Errorf("node x matched %v, want none", slugs(node.Matches))
				}
			}
			if len(diagram.Unresolved) != 1 ||
				diagram.Unresolved[0].Reason != specmodel.C4ExclusionReasonUndeclaredNode {
				t.Errorf("unresolved = %+v, want one undeclared_node", diagram.Unresolved)
			}
		})
	}
}

func TestAnUnreadableMemberDeclaresANamelessElementAndDrawsNothing(t *testing.T) {
	files := map[string]string{
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha") +
			"# Alpha\n\n```mermaid\nC4Container\n" +
			"Container(a, \"Alpha\", \"Go\")\nContainer(e, \"\", \"Go\")\nContainer(g, \"Ghost\", \"Go\")\n" +
			"Rel(a, g, \"Calls\")\nRel(a, e, \"Calls\")\n```\n",
		"beta.md": head("container-beta", "container-blueprint", "Beta") +
			"# Beta\n\n```mermaid\nC4Container\nContainer(b, \"Beta\", \"Go\")\n```\n",
	}
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	built, err := corpus.Build(corpus.Options{Root: dir, Profile: stubProfile{architecture: vocabulary()}})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}
	set, err := c4elements.Build(c4elements.Input{
		Vocabulary: vocabulary(),
		Entries:    built.Index().Entries,
		References: built.Index().References,
		Source: func(path string) ([]byte, error) {
			if path == "beta.md" {
				return nil, os.ErrNotExist
			}
			return built.Source(path)
		},
		Decode: corpus.DecodeRecordBlock,
	})
	if err != nil || set == nil {
		t.Fatalf("Build = %v, %v, want a set and no error", set, err)
	}
	beta, ok := set.Element("container-beta")
	if !ok || beta.Name != "" {
		t.Fatalf("container-beta = %+v, %v, want a nameless element", beta, ok)
	}
	if got := document(t, set, "beta.md"); len(got.Diagrams) != 0 {
		t.Errorf("beta.md has %d diagrams, want none", len(got.Diagrams))
	}
	got := nodes(t, document(t, set, "alpha.md"))
	if got["a"] != "container-alpha" || got["e"] != "" || got["g"] != "" {
		t.Errorf("alpha.md nodes = %v, want a resolved and e, g unmatched", got)
	}
}

func TestMembersWithNoSlugDeclareElementsOfTheirOwn(t *testing.T) {
	diagramBody := "# %s\n\n```mermaid\nC4Container\nContainer(a, \"Alpha\", \"Go\")\nContainer(b, \"Beta\", \"Go\")\n```\n"
	files := map[string]string{
		"a.md": strings.Replace(head("x", "container-blueprint", "Alpha", `contracts_consumed:`, `  - "contract-one"`),
			"slug: x\n", "slug: \"\"\n", 1) + strings.Replace(diagramBody, "%s", "Alpha", 1),
		"b.md": strings.Replace(head("x", "container-blueprint", "Beta", `contracts_consumed:`, `  - "contract-two"`),
			"slug: x\n", "slug: \"\"\n", 1) + strings.Replace(diagramBody, "%s", "Beta", 1),
		"one.md": head("contract-one", "contract", "One"),
		"two.md": head("contract-two", "contract", "Two"),
	}
	for run := 0; run < 5; run++ {
		set := build(t, files, vocabulary())
		var named []*c4elements.Element
		for _, element := range set.Elements() {
			if element.Slug == "" {
				named = append(named, element)
			}
		}
		if len(named) != 2 || named[0].Path != "a.md" || named[0].Name != "Alpha" ||
			named[1].Path != "b.md" || named[1].Name != "Beta" {
			t.Fatalf("run %d: slugless elements = %+v, want Alpha at a.md then Beta at b.md", run, named)
		}
		if _, ok := set.Element(""); ok {
			t.Errorf("Element(\"\") found, want not found")
		}
		if !named[0].Consumes("contract-one") || named[0].Consumes("contract-two") ||
			!named[1].Consumes("contract-two") || named[1].Consumes("contract-one") {
			t.Errorf("consumed contracts crossed: %v / %v", named[0].ConsumedContracts, named[1].ConsumedContracts)
		}
		doc := document(t, set, "a.md")
		for _, node := range doc.Diagrams[0].Nodes {
			if node.Alias == "b" && (len(node.Matches) != 1 || node.Matches[0] != named[1]) {
				t.Errorf("node b matches %v, want exactly the second member's element", node.Matches)
			}
		}
	}
}

func TestADuplicatedNodeAliasMakesItsRelationshipsAmbiguous(t *testing.T) {
	type outcome struct {
		reason specmodel.C4ExclusionReason
		source bool
	}
	cases := []struct {
		name  string
		lines []string
		want  outcome
	}{
		{"target duplicated", []string{
			`Container(web, "Alpha", "Go")`, `Container(api, "Alpha", "Go")`, `Container(api, "Beta", "Go")`,
			`Rel(web, api, "calls")`}, outcome{specmodel.C4ExclusionReasonAmbiguousNode, false}},
		{"source duplicated", []string{
			`Container(api, "Alpha", "Go")`, `Container(api, "Beta", "Go")`, `Container(web, "Alpha", "Go")`,
			`Rel(api, web, "calls")`}, outcome{specmodel.C4ExclusionReasonAmbiguousNode, true}},
		{"external other endpoint takes precedence", []string{
			`Container(api, "Alpha", "Go")`, `Container(api, "Beta", "Go")`, `Container_Ext(ext, "Alpha", "Go")`,
			`Rel(api, ext, "calls")`}, outcome{specmodel.C4ExclusionReasonExternalEndpoint, false}},
		{"undeclared other endpoint takes precedence", []string{
			`Container(api, "Alpha", "Go")`, `Container(api, "Beta", "Go")`,
			`Rel(api, nobody, "calls")`}, outcome{specmodel.C4ExclusionReasonUndeclaredNode, false}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := map[string]string{
				"alpha.md": head("container-alpha", "container-blueprint", "Alpha") +
					"# Alpha\n\n" + strings.Replace(diagram(c.lines...), "C4Component", "C4Container", 1),
				"beta.md": head("container-beta", "container-blueprint", "Beta"),
			}
			d := document(t, build(t, files, vocabulary()), "alpha.md").Diagrams[0]
			if len(d.Drawings) != 0 {
				t.Errorf("drawings = %+v, want none", d.Drawings)
			}
			if len(d.Unresolved) != 1 || d.Unresolved[0].Reason != c.want.reason || d.Unresolved[0].SourceEndpoint != c.want.source {
				t.Errorf("unresolved = %+v, want %+v", d.Unresolved, c.want)
			}
		})
	}
}
