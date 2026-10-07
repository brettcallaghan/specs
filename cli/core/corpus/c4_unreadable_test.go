package corpus_test

import (
	"os"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
)

// containerProfile registers only container-blueprint as an element
// type, over the framework's own schema.
func containerProfile() stubProfile {
	return stubProfile{
		schema: schema.SpecModel,
		types: map[string]corpus.DocumentTypeRules{
			"container-blueprint": {SlugPattern: `^[a-z0-9-]+$`},
		},
		architecture: corpus.ArchitectureRules{
			Elements: map[string]corpus.ElementTypeRules{
				"container-blueprint": {Level: corpus.LevelContainer},
			},
			EntityNameProperty: "entity_name",
		},
	}
}

// One architecture blueprint that cannot be read when Check runs must
// not abort the check or drop the findings of the others.
func TestCheckSurvivesAnUnreadableArchitectureBlueprint(t *testing.T) {
	repo := t.TempDir()
	site := func(slug, body string) string {
		return "---\ntitle: " + slug + "\nslug: " + slug + "\ntype: container-blueprint\nstatus: draft\ndescription: d\n" +
			"updated_at: \"2026-10-07T00:00:00Z\"\nentity_name: " + slug + "\n---\n\n# " + slug + "\n\n" + body
	}
	write(t, repo, "specs/alpha.md", site("alpha",
		"```mermaid\nC4Container\nContainer(g, \"Ghost\", \"Go\")\nContainer(e, \"\", \"Go\")\n```\n"))
	beta := write(t, repo, "specs/beta.md", site("beta", "text\n"))

	built := buildChecked(t, repo, containerProfile())
	if err := os.Remove(beta); err != nil {
		t.Fatal(err)
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check returned %v, want no error", err)
	}
	got := findingsWithRule(findings, finding.RuleC4UndeclaredNode, "specs/alpha.md")
	if len(got) != 2 {
		t.Errorf("c4/undeclared-node findings on alpha.md = %d, want 2 (Ghost and the empty label): %+v", len(got), findings)
	}
}
