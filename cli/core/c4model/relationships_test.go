// Tests for the C4 Model package's relationships and excluded
// relationships.
package c4model_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

func mermaid(diagramType string, lines ...string) string {
	return "```mermaid\n" + diagramType + "\n" + strings.Join(lines, "\n") + "\n```\n"
}

// drawings is a corpus whose diagrams draw every kind of statement the
// model reads:
//
//   - alpha.md, a container in system-one composing module-mu, draws
//     every level of relationship, the directional variants, Rel_Back, a
//     BiRel and one statement for each exclusion reason, and holds a
//     dynamic and a deployment diagram;
//   - beta.md, an orphaned container, draws in its Context view one
//     relationship alpha.md also draws, with a different technology, and
//     one only it draws;
//   - feature.md, a feature blueprint, draws a diagram of its own.
func drawings() map[string]string {
	return map[string]string{
		"one.md": head("system-one", "system-blueprint", "One"),
		"two.md": head("system-two", "system-blueprint", "Two"),
		"alpha.md": head("container-alpha", "container-blueprint", "Alpha",
			"system: system-one", "modules:", `  - "module-mu"`) +
			"# Alpha\n\n## Components\n\n" +
			mermaid("C4Component",
				`Component(p, "P", "Go")`,
				`Component(q, "Q", "Go")`,
				`Component(m, "Mu", "Go")`,
				`Container(a, "Alpha", "Go")`,
				`Container(b, "Beta", "Go")`,
				`System(s1, "One")`,
				`System(s2, "Two")`,
				`Person(u, "User")`,
				`ContainerDb(db, "Store", "Files")`,
				`Component(g, "Ghost", "Go")`,
				`Component(dup, "Dup", "Go")`,
				`Rel(p, q, "Reads documents")`,
				`Rel_U(q, p, "Writes back")`,
				`Rel_Back(p, q, "Notifies")`,
				`Rel(p, m, "Uses")`,
				`Rel(p, a, "Lives in")`,
				`Rel(a, s1, "Belongs to")`,
				`Rel(s1, s2, "Talks to")`,
				`Rel(a, b, "Calls", "HTTP", "On demand")`,
				`BiRel(p, q, "Chats with")`,
				`Rel(u, p, "Drives")`,
				`Rel(p, db, "Stores in")`,
				`Rel(p, g, "Haunts")`,
				`Rel(p, dup, "Doubles")`,
				`Rel(p, , "Broken"`,
			) +
			mermaid("C4Dynamic", `Component(p, "P", "Go")`, `Component(q, "Q", "Go")`, `Rel(p, q, "Dynamically")`) +
			mermaid("C4Deployment", `Component(p, "P", "Go")`, `Component(q, "Q", "Go")`, `Rel(p, q, "Deployed")`) +
			"\n### P\n\n" + record("component-p", "P", "p") +
			"\n### Q\n\n" + record("component-q", "Q", "q") +
			"\n### Dup one\n\n" + record("component-dup-one", "Dup", "d") +
			"\n### Dup two\n\n" + record("component-dup-two", "Dup", "d"),
		"beta.md": head("container-beta", "container-blueprint", "Beta") +
			"# Beta\n\n## Context\n\n" +
			mermaid("C4Context",
				`Container(a, "Alpha", "Go")`,
				`Container(b, "Beta", "Go")`,
				`Rel_Back(b, a, "Calls", "gRPC")`,
				`Rel(b, a, "Reports to")`,
			),
		"mu.md": head("module-mu", "module-blueprint", "Mu", "kind: standard", "visibility: internal"),
		"feature.md": head("feature-delta", "feature-blueprint", "") +
			"# Delta\n\n" +
			mermaid("C4Container", `Container(a, "Alpha", "Go")`, `Container(b, "Beta", "Go")`,
				`Rel(a, b, "Featured")`, `Person(u, "User")`, `Rel(u, a, "Featured too")`),
	}
}

func relationship(t *testing.T, model *specmodel.C4Model, source, target, description string) specmodel.C4Relationship {
	t.Helper()
	for _, r := range model.Relationships {
		if string(r.Source) == source && string(r.Target) == target && r.Description == description {
			return r
		}
	}
	var held []string
	for _, r := range model.Relationships {
		held = append(held, string(r.Source)+" -> "+string(r.Target)+" "+r.Description)
	}
	t.Fatalf("no relationship %s -> %s %q; the model holds:\n%s", source, target, description, strings.Join(held, "\n"))
	return specmodel.C4Relationship{}
}

func text(value *string) string {
	if value == nil {
		return "<absent>"
	}
	return *value
}

// AC: AC-C4-002.1
func TestRelArgumentsBecomeRelationshipProperties(t *testing.T) {
	for _, example := range []struct {
		statement, technology, detail string
	}{
		{`Rel(a, b, "Reads documents")`, "<absent>", "<absent>"},
		{`Rel(a, b, "Reads documents", "filesystem")`, "filesystem", "<absent>"},
		{`Rel(a, b, "Reads documents", "filesystem", "At start")`, "filesystem", "At start"},
		{`Rel(a, b, "Reads documents", "", "")`, "<absent>", "<absent>"},
	} {
		t.Run(example.statement, func(t *testing.T) {
			model := build(t, map[string]string{
				"a.md": head("container-a", "container-blueprint", "A") + "# A\n\n" +
					mermaid("C4Container", `Container(a, "A", "Go")`, `Container(b, "B", "Go")`, example.statement),
				"b.md": head("container-b", "container-blueprint", "B"),
			})
			r := relationship(t, model, "container-a", "container-b", "Reads documents")
			if got := text(r.Technology); got != example.technology {
				t.Errorf("technology = %s, want %s", got, example.technology)
			}
			if got := text(r.Detail); got != example.detail {
				t.Errorf("detail = %s, want %s", got, example.detail)
			}
		})
	}
}

// AC: AC-C4-002.2
func TestDirectionalLayoutVariantsAreTheSameRelationship(t *testing.T) {
	model := build(t, drawings())

	r := relationship(t, model, "component-q", "component-p", "Writes back")
	if len(r.DrawnIn) != 1 || r.DrawnIn[0].SourceAlias != "q" || r.DrawnIn[0].TargetAlias != "p" {
		t.Errorf("drawn_in = %+v, want one drawing from q to p", r.DrawnIn)
	}
}

// AC: AC-C4-002.3
func TestRelBackReversesTheDependency(t *testing.T) {
	model := build(t, drawings())

	r := relationship(t, model, "component-q", "component-p", "Notifies")
	if got := r.DrawnIn[0]; got.SourceAlias != "q" || got.TargetAlias != "p" {
		t.Errorf("aliases = %s -> %s, want q -> p after normalisation", got.SourceAlias, got.TargetAlias)
	}
}

// AC: AC-C4-002.4
func TestARelationshipDrawnInBothOfItsHomesIsOneRelationship(t *testing.T) {
	files := drawings()
	model := build(t, files)

	count := 0
	for _, r := range model.Relationships {
		if r.Source == "container-alpha" && r.Target == "container-beta" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("container-alpha -> container-beta is held %d times, want once", count)
	}
	r := relationship(t, model, "container-alpha", "container-beta", "Calls")
	if len(r.DrawnIn) != 2 {
		t.Fatalf("drawn_in holds %d drawings, want 2", len(r.DrawnIn))
	}
	want := []struct {
		path   string
		view   specmodel.C4View
		source string
		target string
		prefix string
	}{
		{"alpha.md", specmodel.C4ViewComponent, "a", "b", "Rel(a, b,"},
		{"beta.md", specmodel.C4ViewContext, "a", "b", "Rel_Back(b, a,"},
	}
	for i, w := range want {
		got := r.DrawnIn[i]
		if got.Path != w.path || got.View != w.view || got.SourceAlias != w.source || got.TargetAlias != w.target {
			t.Errorf("drawn_in[%d] = %s %s %s->%s, want %s %s %s->%s",
				i, got.Path, got.View, got.SourceAlias, got.TargetAlias, w.path, w.view, w.source, w.target)
		}
		written := files[got.Path][got.Span.Offset : got.Span.Offset+got.Span.Length]
		if !strings.HasPrefix(written, w.prefix) {
			t.Errorf("drawn_in[%d] span covers %q, want the statement %s…", i, written, w.prefix)
		}
	}
}

// AC: AC-C4-002.5
func TestARelationshipDrawnOnlyInAContextViewIsAdmitted(t *testing.T) {
	model := build(t, drawings())

	r := relationship(t, model, "container-beta", "container-alpha", "Reports to")
	if len(r.DrawnIn) != 1 || r.DrawnIn[0].View != specmodel.C4ViewContext {
		t.Errorf("drawn_in = %+v, want one Context drawing", r.DrawnIn)
	}
}

// AC: AC-C4-002.6
func TestARelationshipRecordsItsLevelFromItsEndpoints(t *testing.T) {
	model := build(t, drawings())

	for _, example := range []struct {
		source, target, description string
		level                       specmodel.C4Level
	}{
		{"system-one", "system-two", "Talks to", specmodel.C4LevelSystem},
		{"container-alpha", "system-one", "Belongs to", specmodel.C4LevelContainer},
		{"container-alpha", "container-beta", "Calls", specmodel.C4LevelContainer},
		{"component-p", "container-alpha", "Lives in", specmodel.C4LevelComponent},
		{"component-p", "module-mu", "Uses", specmodel.C4LevelComponent},
		{"component-p", "component-q", "Reads documents", specmodel.C4LevelComponent},
	} {
		r := relationship(t, model, example.source, example.target, example.description)
		if r.Level != example.level {
			t.Errorf("%s -> %s level = %s, want %s", example.source, example.target, r.Level, example.level)
		}
	}
}

// AC: AC-C4-002.9
func TestDrawingsOfOneRelationshipThatDisagreeKeepTheFirst(t *testing.T) {
	model := build(t, drawings())

	r := relationship(t, model, "container-alpha", "container-beta", "Calls")
	if got := text(r.Technology); got != "HTTP" {
		t.Errorf("technology = %s, want HTTP from alpha.md, which sorts first", got)
	}
	if got := text(r.Detail); got != "On demand" {
		t.Errorf("detail = %s, want On demand from alpha.md", got)
	}
	if len(r.DrawnIn) != 2 {
		t.Errorf("drawn_in holds %d drawings, want both", len(r.DrawnIn))
	}
}

// AC: AC-C4-002.10
func TestEdgesTheModelDoesNotAdmitAreExcluded(t *testing.T) {
	files := drawings()
	model := build(t, files)

	want := map[string]specmodel.C4ExclusionReason{
		`Rel(u, p, "Drives")`:       specmodel.C4ExclusionReasonExternalEndpoint,
		`Rel(p, db, "Stores in")`:   specmodel.C4ExclusionReasonPassiveStore,
		`Rel(p, g, "Haunts")`:       specmodel.C4ExclusionReasonUndeclaredNode,
		`Rel(p, dup, "Doubles")`:    specmodel.C4ExclusionReasonAmbiguousNode,
		`BiRel(p, q, "Chats with")`: specmodel.C4ExclusionReasonBidirectional,
	}
	got := map[string]specmodel.C4ExclusionReason{}
	for _, e := range model.ExcludedRelationships {
		if e.Path != "alpha.md" {
			t.Errorf("excluded %q from %s, which draws no excluded statement", e.Text, e.Path)
			continue
		}
		if written := files[e.Path][e.Span.Offset : e.Span.Offset+e.Span.Length]; written != e.Text {
			t.Errorf("text %q does not match its span %q", e.Text, written)
		}
		got[e.Text] = e.Reason
	}
	for statement, reason := range want {
		if got[statement] != reason {
			t.Errorf("%s excluded with %q, want %q", statement, got[statement], reason)
		}
	}
	if len(got) != len(want) {
		t.Errorf("excluded %d statements, want %d: %v", len(got), len(want), got)
	}
	for i := 1; i < len(model.ExcludedRelationships); i++ {
		a, b := model.ExcludedRelationships[i-1], model.ExcludedRelationships[i]
		if a.Path > b.Path || (a.Path == b.Path && a.Span.Offset > b.Span.Offset) {
			t.Errorf("excluded_relationships out of order at %d", i)
		}
	}
	for _, r := range model.Relationships {
		for _, label := range []string{"Drives", "Stores in", "Haunts", "Doubles", "Chats with", "Broken"} {
			if r.Description == label {
				t.Errorf("relationships include the excluded statement %q", label)
			}
		}
	}
}

// AC: AC-C4-002.11
func TestDynamicAndDeploymentDiagramsContributeNoRelationships(t *testing.T) {
	model := build(t, drawings())

	for _, r := range model.Relationships {
		if r.Description == "Dynamically" || r.Description == "Deployed" {
			t.Errorf("relationship %q recorded from a dynamic or deployment diagram", r.Description)
		}
	}
	for _, e := range model.ExcludedRelationships {
		if strings.Contains(e.Text, "Dynamically") || strings.Contains(e.Text, "Deployed") {
			t.Errorf("excluded relationship %q recorded from a dynamic or deployment diagram", e.Text)
		}
	}
}

// AC: AC-C4-002.12
func TestDiagramsOutsideArchitectureBlueprintsContributeNoRelationships(t *testing.T) {
	model := build(t, drawings())

	for _, r := range model.Relationships {
		for _, d := range r.DrawnIn {
			if d.Path == "feature.md" {
				t.Errorf("relationship %q drawn in the feature blueprint", r.Description)
			}
		}
	}
	for _, e := range model.ExcludedRelationships {
		if e.Path == "feature.md" {
			t.Errorf("excluded relationship %q from the feature blueprint", e.Text)
		}
	}
}

// Relationships are ordered by source slug, then target slug, then
// description.
func TestRelationshipsAreOrderedBySourceTargetAndDescription(t *testing.T) {
	model := build(t, drawings())

	if len(model.Relationships) < 2 {
		t.Fatalf("model holds %d relationships; the check would pass vacuously", len(model.Relationships))
	}
	for i := 1; i < len(model.Relationships); i++ {
		a, b := model.Relationships[i-1], model.Relationships[i]
		key := func(r specmodel.C4Relationship) string {
			return string(r.Source) + "\x00" + string(r.Target) + "\x00" + r.Description
		}
		if key(a) > key(b) {
			t.Errorf("relationships out of order: %s before %s", key(a), key(b))
		}
	}
}
