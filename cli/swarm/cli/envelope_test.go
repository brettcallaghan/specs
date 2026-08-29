// Tests that what this surface emits under --json conforms to the schema
// that fixes it.
//
// Without them the generated schema and the code that emits it are one
// edit apart from disagreeing, and a consumer generating types from the
// schema would be the one to find out. The schema is read as the
// generated artifact a consumer would validate against, not as the
// .strata source it came from.
package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// validatorFor compiles the generated swarm bundle against one of its
// definitions.
func validatorFor(t *testing.T, definition string) *jsonschema.Schema {
	t.Helper()

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.SwarmModel))
	if err != nil {
		t.Fatalf("reading the generated swarm schema: %v", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("swarm-model.schema.json", doc); err != nil {
		t.Fatalf("adding the swarm schema: %v", err)
	}
	compiled, err := compiler.Compile("swarm-model.schema.json#/definitions/" + definition)
	if err != nil {
		t.Fatalf("compiling %s: %v", definition, err)
	}
	return compiled
}

// specWithEveryFactKind reports one fact of each shape the contract
// declares, plus an unestablished one, so a single run exercises the
// whole of the envelope a command can fill in.
func specWithEveryFactKind() Spec {
	from := []string{"bd", "show", "run-1", "--json"}
	return Spec{
		Use: "probe", Short: "a synthetic subcommand", RunScoped: true,
		Plan: func(Request) (plan.Plan, error) { return readPlan(), nil },
		Interpret: func(Request, plan.Report) (Result, error) {
			return Result{
				Outcome:    OutcomeAttention,
				Summary:    "the probe ran",
				Resolution: "close the bead or mark it blocked",
				Facts: []Fact{
					Text("branch", "claude/branch-x", from),
					Count("ready_front", 3, from),
					Flag("drained", false, from),
					List("admitted", []string{"run-1.1", "run-1.2"}, from),
					Unestablished("review_passes", "the key is absent", from),
				},
			}, nil
		},
	}
}

// AC: AC-SWARM-CORE-001.9
//
// The contract's guarantee is that the envelope holds on every exit
// status and in every mode, so every mode and a refusal are all put
// through the validator rather than one happy path.
func TestWhatTheSurfaceEmitsConformsToTheEnvelopeSchema(t *testing.T) {
	compiled := validatorFor(t, "CommandResult")

	for _, args := range [][]string{
		{"probe", "--run", "run-1", "--json"},
		{"probe", "--run", "run-1", "--json", "--explain"},
		{"probe", "--run", "run-1", "--json", "--dry-run"},
		{"probe", "--json"}, // refused: no run scope
	} {
		stdout, _, code := runSurface(t, specWithEveryFactKind(), fakeForRead(), args...)

		var value any
		if err := json.Unmarshal([]byte(stdout), &value); err != nil {
			t.Errorf("%v (exit %d): the result does not parse: %v\n%s", args, code, err, stdout)
			continue
		}
		if err := compiled.Validate(value); err != nil {
			t.Errorf("%v (exit %d): the emitted envelope does not conform to the schema:\n%v\n%s",
				args, code, err, stdout)
		}
	}
}

// Each of the four fact shapes conforms to the union the schema declares,
// so the surface cannot emit a value a consumer branching on the kind
// would not recognise.
func TestEveryFactKindTheSurfaceCanEmitIsOneTheSchemaDeclares(t *testing.T) {
	compiled := validatorFor(t, "FactValue")

	from := []string{"bd", "show", "run-1"}
	for _, f := range []Fact{
		Text("a", "x", from),
		Count("b", 1, from),
		Flag("c", true, from),
		List("d", []string{"x"}, from),
	} {
		data, err := json.Marshal(f.Value)
		if err != nil {
			t.Errorf("%s: marshalling: %v", f.Name, err)
			continue
		}
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Errorf("%s: %v", f.Name, err)
			continue
		}
		if err := compiled.Validate(value); err != nil {
			t.Errorf("%s: %s does not conform to FactValue:\n%v", f.Name, data, err)
		}
	}
}

// The negative control for the two tests above: a document the schema
// should reject really is rejected, so a validator that accepted anything
// could not make them pass.
func TestTheEnvelopeSchemaRejectsWhatItShould(t *testing.T) {
	compiled := validatorFor(t, "CommandResult")

	for name, body := range map[string]string{
		"an unknown key":        `{"contract_version":"1.0","command":"c","run":"","mode":"execute","outcome":"clear","status":0,"summary":"","resolution":"","facts":[],"invocations":[],"transcript":[],"surprise":1}`,
		"an unknown outcome":    `{"contract_version":"1.0","command":"c","run":"","mode":"execute","outcome":"maybe","status":0,"summary":"","resolution":"","facts":[],"invocations":[],"transcript":[]}`,
		"a status out of range": `{"contract_version":"1.0","command":"c","run":"","mode":"execute","outcome":"clear","status":9,"summary":"","resolution":"","facts":[],"invocations":[],"transcript":[]}`,
		"a missing field":       `{"contract_version":"1.0","command":"c"}`,
	} {
		var value any
		if err := json.Unmarshal([]byte(body), &value); err != nil {
			t.Fatalf("%s: the test's own document does not parse: %v", name, err)
		}
		if err := compiled.Validate(value); err == nil {
			t.Errorf("%s: the schema accepted it", name)
		}
	}
}

// A run manifest as the skills write it today validates against the
// declared key set. This is the check that the schema describes the
// record that exists rather than one that would be nicer.
func TestAManifestOfTheShapeTheSkillsWriteTodayValidates(t *testing.T) {
	compiled := validatorFor(t, "RunManifest")

	const manifest = `{
      "target": "run-1.target",
      "branch": "claude/bead-bonanza-run-1",
      "base_sha": "3a57dc733e8e8ee0efbd4cb58e1880a492f4e502",
      "checks_ids": ["format", "vet", "build"],
      "checks_baseline": "# Checks baseline\n\nEverything passed.\n",
      "checks_baseline_status": 0,
      "concurrency_cap": 20,
      "models": {"light": "haiku", "standard": "sonnet", "advanced": "opus", "frontier": "fable"},
      "effort": "unrecorded",
      "standing_auth": {
        "pull_request": {"value": "draft", "source": "default"},
        "review_budget": {"value": 2, "source": "default"},
        "tier_escalation": {"value": "once", "source": "default"},
        "merge_to_default": {"value": false, "source": "default"},
        "move_foreign_beads": {"value": false, "source": "default"},
        "admit_review_findings": {"value": false, "source": "default"}
      },
      "admitted": ["run-1.1", "run-1.2"],
      "excluded": [],
      "moved": [],
      "partition_preview": "accepted — 6 edges validated independently",
      "review_passes": 0,
      "review_budget": 2,
      "collision_edges": ".beads/plans/run-1.edges.jsonl",
      "ledger": ["a claim — proved by a command, observed a result"]
    }`

	var value any
	if err := json.Unmarshal([]byte(manifest), &value); err != nil {
		t.Fatalf("the sample manifest does not parse: %v", err)
	}
	if err := compiled.Validate(value); err != nil {
		t.Errorf("a manifest of the shape the skills write today does not validate:\n%v", err)
	}

	// A partial manifest is legitimate: the record grows as a run
	// proceeds, and a key's absence before its stage is a state rather
	// than a defect.
	var partial any
	if err := json.Unmarshal([]byte(`{"checks_ids": ["format"], "checks_baseline_status": 0}`), &partial); err != nil {
		t.Fatalf("the partial manifest does not parse: %v", err)
	}
	if err := compiled.Validate(partial); err != nil {
		t.Errorf("a manifest partway through a run does not validate:\n%v", err)
	}

	// A key outside the declared set is refused rather than stored, and a
	// value of the wrong type for its key is refused — both by the closed
	// shape rather than by a rule somebody has to remember.
	for name, body := range map[string]string{
		"an undeclared key": `{"circuit_breaker": 3}`,
		"a wrong type":      `{"review_passes": "two"}`,
		"a wrong row type":  `{"standing_auth": {"review_budget": {"value": "two", "source": "default"}}}`,
	} {
		var bad any
		if err := json.Unmarshal([]byte(body), &bad); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := compiled.Validate(bad); err == nil {
			t.Errorf("%s: the schema accepted it", name)
		}
	}
}

// The one rule the schema cannot state: a value carries exactly the field
// its kind names. A schema can say which fields may appear; which
// combination is coherent is a rule only code can hold, so the surface
// refuses an incoherent value rather than rendering it.
func TestAValueCarryingTheWrongFieldForItsKindIsRefused(t *testing.T) {
	text := "x"
	count := int64(1)

	for name, value := range map[string]swarmmodel.FactValue{
		"a text kind carrying a count":      {Kind: swarmmodel.FactKindText, Count: &count},
		"a count kind carrying text":        {Kind: swarmmodel.FactKindCount, Text: &text},
		"two fields at once":                {Kind: swarmmodel.FactKindText, Text: &text, Count: &count},
		"no field at all":                   {Kind: swarmmodel.FactKindText},
		"a kind this surface does not know": {Kind: swarmmodel.FactKind("colour"), Text: &text},
	} {
		v := value
		result := Result{
			Command: "swarm probe", Outcome: OutcomeClear,
			Facts: []Fact{{
				Name: "branch", Established: true, Value: &v,
				Invocations: [][]string{{"bd", "show", "run-1"}},
			}},
		}
		if err := result.Validate(); err == nil {
			t.Errorf("%s: Validate accepted it", name)
		}
	}

	// And the four the constructors build are all coherent, so the rule
	// cannot be satisfied by refusing everything.
	from := []string{"bd", "show", "run-1"}
	for _, f := range []Fact{
		Text("a", "x", from), Count("b", 1, from),
		Flag("c", true, from), List("d", []string{"x"}, from),
	} {
		result := Result{Command: "swarm probe", Outcome: OutcomeClear, Facts: []Fact{f}}
		if err := result.Validate(); err != nil {
			t.Errorf("%s: Validate refused a value a constructor built: %v", f.Name, err)
		}
	}
}
