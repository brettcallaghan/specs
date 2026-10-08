// Tests for narrowing the C4 model to one element's subtree.
package c4model_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// narrowed is the element-tree fixture with relationships drawn: in
// alpha.md, Parser calls Solo (both alpha's) and Beta (an orphaned
// container), and an undeclared node is excluded; in gamma.md, Own (the
// module's) uses Shared (placed under container-beta), and an external
// edge is excluded.
func narrowed() map[string]string {
	files := fixture()
	files["alpha.md"] += "\n## Relationships\n\n" + mermaid("C4Component",
		`Component(p, "Parser", "Go")`, `Component(s, "Solo", "Go")`, `Container(b, "Beta", "Go")`,
		`Component(x, "Nobody", "Go")`,
		`Rel(p, s, "Calls solo")`, `Rel(p, b, "Calls beta")`, `Rel(p, x, "Calls nobody")`)
	files["gamma.md"] += "\n" + mermaid("C4Component",
		`Component(o, "Own", "Go")`, `Component(sh, "Shared", "Go")`, `Person(u, "User")`,
		`Rel(o, sh, "Uses shared")`, `Rel(u, o, "Drives")`)
	return files
}

// shape summarises a model as the slugs each collection holds, nesting
// written with braces.
func shape(model *specmodel.C4Model) string {
	component := func(list []specmodel.C4Component) string { return componentSlugs(list) }
	container := func(c specmodel.C4Container) string {
		return string(c.Slug) + "{" + component(c.Components) + "|" + slugsOf(c.Modules) + "}"
	}
	var systems, orphans, modules []string
	for _, s := range model.SoftwareSystems {
		var held []string
		for _, c := range s.Containers {
			held = append(held, container(c))
		}
		systems = append(systems, string(s.Slug)+"{"+strings.Join(held, ",")+"}")
	}
	for _, c := range model.OrphanedContainers {
		orphans = append(orphans, container(c))
	}
	for _, m := range model.Modules {
		modules = append(modules, string(m.Slug)+"{"+component(m.Components)+"|"+slugsOf(m.Modules)+"}")
	}
	return "systems=" + strings.Join(systems, ",") +
		" orphaned_containers=" + strings.Join(orphans, ",") +
		" orphaned_components=" + component(model.OrphanedComponents) +
		" modules=" + strings.Join(modules, ",")
}

func edges(model *specmodel.C4Model) string {
	var out []string
	for _, r := range model.Relationships {
		out = append(out, r.Description)
	}
	for _, e := range model.ExcludedRelationships {
		out = append(out, "excluded:"+e.Path)
	}
	return strings.Join(out, ",")
}

func TestNarrowingKeepsTheRootsSubtreeAndItsAncestorChain(t *testing.T) {
	for _, example := range []struct {
		root, shape, edges string
	}{
		{
			// A system: its containers and everything they hold.
			"system-one",
			"systems=system-one{container-alpha{component-parser,component-solo,component-twin|module-gamma}}" +
				" orphaned_containers= orphaned_components= modules=module-gamma{component-own|}",
			"Calls solo,excluded:alpha.md,excluded:gamma.md",
		},
		{
			// A container in a system: the system holds it alone.
			"container-alpha",
			"systems=system-one{container-alpha{component-parser,component-solo,component-twin|module-gamma}}" +
				" orphaned_containers= orphaned_components= modules=module-gamma{component-own|}",
			"Calls solo,excluded:alpha.md,excluded:gamma.md",
		},
		{
			// An orphaned container stays in its orphaned collection.
			"container-beta",
			"systems= orphaned_containers=container-beta{component-shared|module-gamma}" +
				" orphaned_components= modules=module-gamma{component-own|}",
			"Uses shared,excluded:gamma.md",
		},
		{
			"container-delta",
			"systems= orphaned_containers=container-delta{|} orphaned_components= modules=",
			"",
		},
		{
			// A module has no tree ancestor.
			"module-gamma",
			"systems= orphaned_containers= orphaned_components= modules=module-gamma{component-own|}",
			"excluded:gamma.md",
		},
		{
			// A component under a container under a system: the chain holds
			// nothing else, not even the container's composed modules.
			"component-parser",
			"systems=system-one{container-alpha{component-parser|}} orphaned_containers= orphaned_components= modules=",
			"",
		},
		{
			// A component under a module appears under it in modules.
			"component-own",
			"systems= orphaned_containers= orphaned_components= modules=module-gamma{component-own|}",
			"",
		},
		{
			"component-lonely",
			"systems= orphaned_containers= orphaned_components=component-lonely modules=",
			"",
		},
		{
			// The root is compared without regard to letter case.
			"COMPONENT-LONELY",
			"systems= orphaned_containers= orphaned_components=component-lonely modules=",
			"",
		},
	} {
		t.Run(example.root, func(t *testing.T) {
			model, err := buildRooted(t, narrowed(), example.root)
			if err != nil {
				t.Fatalf("C4Model: %v", err)
			}
			if got := shape(model); got != example.shape {
				t.Errorf("shape:\n got  %s\n want %s", got, example.shape)
			}
			if got := edges(model); got != example.edges {
				t.Errorf("edges = %q, want %q", got, example.edges)
			}
		})
	}
}

// Derived properties keep their whole-corpus values in a narrowed model.
func TestNarrowingKeepsDerivedPropertiesWhole(t *testing.T) {
	model, err := buildRooted(t, narrowed(), "module-gamma")
	if err != nil {
		t.Fatal(err)
	}
	if got := slugsOf(model.Modules[0].ContainedBy); got != "container-alpha,container-beta" {
		t.Errorf("contained_by = %q, want both hosts though neither is in the narrowed model", got)
	}
}

// A root naming no element is an answer, not an error: a model holding
// no element and no relationship, with root_declared false. A root naming
// an element is declared, and a whole-corpus model names no root.
//
// AC: AC-C4-004.6
func TestARootNamingNoElementIsAnsweredAsNotDeclared(t *testing.T) {
	model, err := buildRooted(t, narrowed(), "component-missing")
	if err != nil {
		t.Fatalf("err = %v, want an answer", err)
	}
	if model.RootDeclared == nil || *model.RootDeclared {
		t.Errorf("root_declared = %v, want false", model.RootDeclared)
	}
	if len(model.SoftwareSystems)+len(model.OrphanedContainers)+len(model.OrphanedComponents)+
		len(model.Modules)+len(model.Relationships)+len(model.ExcludedRelationships) != 0 {
		t.Errorf("model holds elements or relationships: %+v", model)
	}
	if model.SoftwareSystems == nil || model.Relationships == nil || model.ExcludedRelationships == nil {
		t.Errorf("a collection is nil rather than empty: %+v", model)
	}

	known, err := buildRooted(t, narrowed(), "module-gamma")
	if err != nil {
		t.Fatal(err)
	}
	if known.RootDeclared == nil || !*known.RootDeclared {
		t.Errorf("root_declared = %v for a declared root, want true", known.RootDeclared)
	}

	whole, err := buildRooted(t, narrowed(), "")
	if err != nil {
		t.Fatal(err)
	}
	if whole.RootDeclared != nil {
		t.Errorf("root_declared = %v on a whole-corpus model, want absent", *whole.RootDeclared)
	}
}
