// Tests for the architecture vocabulary this profile registers.
//
// As in registration_test.go, the schema is read here through an untyped
// traversal of its own, so the drift guard cannot agree with a misread
// schema by sharing the code that misread it.
package profile_test

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/profile"
)

// The registration is exactly the vocabulary the C4 diagrams blueprint's
// Profile Registration section lists, transcribed here on purpose: a test
// that read it back from the same place the profile does would assert
// nothing about what the profile registers.
func TestTheProfileRegistersTheArchitectureVocabulary(t *testing.T) {
	roster := func(level corpus.ElementLevel) corpus.ElementTypeRules {
		return corpus.ElementTypeRules{Level: level, RosterHeading: "Components", RosterTag: "component"}
	}
	want := corpus.ArchitectureRules{
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

	if got := profile.New().Architecture(); !reflect.DeepEqual(got, want) {
		t.Errorf("Architecture() =\n  %+v\nwant\n  %+v", got, want)
	}
}

// Every registered property is declared, in the embedded union schema, on
// every type it applies to: the entity-name and consumed-contracts
// properties on every element type, the system property on the
// container-level types, the composed-modules property on the container-
// and module-level types, and the provider property on the contract type.
// A property renamed in the vocabulary fails here rather than silently
// resolving nothing.
func TestEveryRegisteredArchitecturePropertyIsDeclaredWhereItApplies(t *testing.T) {
	rules := profile.New().Architecture()
	checked := 0

	expect := func(documentType, role, property string) {
		t.Helper()
		if property == "" {
			t.Errorf("the profile registers no %s", role)
			return
		}
		checked++
		if !declaredPropertiesOf(t, documentType)[property] {
			t.Errorf("%s %q is not declared on %q in the generated schema", role, property, documentType)
		}
	}

	for documentType, element := range rules.Elements {
		expect(documentType, "entity-name property", rules.EntityNameProperty)
		expect(documentType, "consumed-contracts property", rules.ConsumedContractsProperty)
		if element.Level == corpus.LevelContainer {
			expect(documentType, "system property", rules.SystemProperty)
		}
		if element.Level == corpus.LevelContainer || element.Level == corpus.LevelModule {
			expect(documentType, "composed-modules property", rules.ComposedModulesProperty)
		}
	}
	expect(rules.ContractType, "provider property", rules.ProviderProperty)

	if checked == 0 {
		t.Fatal("checked no property; the guard would pass vacuously")
	}
}

// declaredPropertiesOf returns the properties the generated schema
// declares on one document type's definition.
func declaredPropertiesOf(t *testing.T, documentType string) map[string]bool {
	t.Helper()

	var bundle map[string]any
	if err := json.Unmarshal(schema.SpecModel, &bundle); err != nil {
		t.Fatalf("reading the generated schema: %v", err)
	}
	definitions, ok := bundle["definitions"].(map[string]any)
	if !ok {
		t.Fatal("the generated schema has no definitions object")
	}
	name := pascalCase(documentType)
	branch, ok := definitions[name].(map[string]any)
	if !ok {
		t.Fatalf("the generated schema declares no branch %q for document type %q", name, documentType)
	}
	properties, ok := branch["properties"].(map[string]any)
	if !ok {
		t.Fatalf("the %q branch declares no properties", name)
	}

	declared := make(map[string]bool, len(properties))
	for property := range properties {
		declared[property] = true
	}
	return declared
}

// Over this repository's own corpus, every Container and Component node
// drawn on its C4 diagrams names exactly one element. A Db or Queue node
// naming nothing is a passive store — the Spec Corpus the container
// diagram draws — and is not an element to resolve.
func TestEveryNodeOnThisRepositorysDiagramsResolvesToOneElement(t *testing.T) {
	root := repositoryRoot(t)
	built, err := corpus.Build(corpus.Options{
		RepoRoot:  root,
		Root:      filepath.Join(root, "specs"),
		Workspace: "agent-plugins",
		Profile:   profile.New(),
	})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}
	set, err := built.C4Elements()
	if err != nil {
		t.Fatalf("C4Elements: %v", err)
	}

	diagrams, resolved := 0, 0
	for _, document := range set.Documents() {
		for _, diagram := range document.Diagrams {
			if !diagram.Parsed.Read() {
				continue
			}
			diagrams++
			for _, node := range diagram.Nodes {
				macro := diagram.Parsed.Statements[node.Statement].Macro
				if node.External || node.PassiveStore ||
					!(strings.HasPrefix(macro, "Container") || strings.HasPrefix(macro, "Component")) {
					continue
				}
				if len(node.Matches) != 1 {
					t.Errorf("%s: %s node %q resolves to %d elements, want 1",
						document.Path, macro, node.Label, len(node.Matches))
					continue
				}
				resolved++
			}
		}
	}
	t.Logf("resolved %d nodes over %d diagrams", resolved, diagrams)
	if diagrams != 4 {
		t.Errorf("read %d C4 diagrams, want this repository's 4", diagrams)
	}
	if resolved == 0 {
		t.Fatal("resolved no node; the check would pass vacuously")
	}
}
