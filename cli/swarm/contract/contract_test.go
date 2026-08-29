// Tests that the generated schema really is one declaration of the
// manifest's keys, and that what the command surface emits conforms to
// the envelope that schema fixes.
//
// Both are read from the generated artifact rather than from the .strata
// source: the artifact is what a consumer validates against, so it is
// what a claim about the contract has to be made of.
package contract_test

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

type definition struct {
	Properties map[string]json.RawMessage `json:"properties"`
}

func definitions(t *testing.T) map[string]definition {
	t.Helper()

	var bundle struct {
		Definitions map[string]definition `json:"definitions"`
	}
	if err := json.Unmarshal(schema.SwarmModel, &bundle); err != nil {
		t.Fatalf("the generated swarm schema does not parse: %v", err)
	}
	if len(bundle.Definitions) == 0 {
		t.Fatal("the generated swarm schema carries no definitions")
	}
	return bundle.Definitions
}

func keysOf(d definition) []string {
	out := make([]string, 0, len(d.Properties))
	for k := range d.Properties {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// AC: AC-SWARM-CORE-002.19
//
// A key's name, its type and the stage at which it becomes required all
// come from one declaration: the stage struct it is written in. This test
// is what makes that structural rather than conventional — a key added to
// the manifest without a stage, or given a different type at its stage
// than in the manifest, fails here.
func TestEveryManifestKeyIsDeclaredExactlyOnceWithItsStage(t *testing.T) {
	defs := definitions(t)

	manifest, ok := defs["RunManifest"]
	if !ok {
		t.Fatal("the schema has no RunManifest definition")
	}
	if len(manifest.Properties) == 0 {
		t.Fatal("RunManifest declares no keys")
	}

	// Which stage first declares each key. Inheritance flattens, so a key
	// appears in its own stage's struct and in every later one; the first
	// stage carrying it is the stage it becomes required at.
	stageOf := map[string]swarmmodel.RunStage{}
	var previous []string

	for _, stage := range contract.Stages() {
		name := contract.StageStructName(stage)
		def, ok := defs[name]
		if !ok {
			t.Fatalf("stage %q names definition %q, which the schema does not have", stage, name)
		}

		keys := keysOf(def)
		for _, key := range keys {
			if _, seen := stageOf[key]; !seen {
				stageOf[key] = stage
			}
		}

		// Each stage carries everything the one before it did. Without
		// this the chain could be re-parented and a key's stage would
		// change without anyone editing its declaration.
		for _, key := range previous {
			if _, ok := def.Properties[key]; !ok {
				t.Errorf("%s does not carry %q, which an earlier stage declares", name, key)
			}
		}
		previous = keys

		// A key's type at its stage is its type in the manifest. One
		// declaration means one type, not two that happen to agree today.
		for _, key := range keys {
			if !bytes.Equal(def.Properties[key], manifest.Properties[key]) {
				t.Errorf("key %q is declared differently in %s and in RunManifest:\n  %s\n  %s",
					key, name, def.Properties[key], manifest.Properties[key])
			}
		}
	}

	// And nothing is in the manifest that no stage declares — which is
	// what a key added straight to RunManifest would look like.
	for _, key := range keysOf(manifest) {
		if _, ok := stageOf[key]; !ok {
			t.Errorf("key %q is in RunManifest and belongs to no stage", key)
		}
	}
	if len(stageOf) != len(manifest.Properties) {
		t.Errorf("the stages declare %d keys and RunManifest carries %d", len(stageOf), len(manifest.Properties))
	}

	// A spot check that the chain is really carrying the run's own keys,
	// so a schema that lost its contents could not pass the rules above
	// by being uniformly empty.
	for key, want := range map[string]swarmmodel.RunStage{
		"checks_baseline": swarmmodel.RunStagePreFlight,
		"admitted":        swarmmodel.RunStageScope,
		"collision_edges": swarmmodel.RunStagePlan,
		"base_sha":        swarmmodel.RunStageDispatch,
		"findings_path":   swarmmodel.RunStageRetro,
	} {
		if got := stageOf[key]; got != want {
			t.Errorf("key %q becomes required at stage %q, want %q", key, got, want)
		}
	}
}

// The stage a struct name is derived from, and back again. A stage whose
// struct name did not round-trip would silently have no keys.
func TestAStageNamesItsOwnStruct(t *testing.T) {
	want := map[swarmmodel.RunStage]string{
		swarmmodel.RunStagePreFlight: "PreFlightKeys",
		swarmmodel.RunStageScope:     "ScopeKeys",
		swarmmodel.RunStagePlan:      "PlanKeys",
		swarmmodel.RunStageDispatch:  "DispatchKeys",
		swarmmodel.RunStageRetro:     "RetroKeys",
	}
	for stage, name := range want {
		if got := contract.StageStructName(stage); got != name {
			t.Errorf("StageStructName(%q) = %q, want %q", stage, got, name)
		}
	}
	if len(contract.Stages()) != len(want) {
		t.Errorf("Stages() has %d entries, want %d", len(contract.Stages()), len(want))
	}
}
