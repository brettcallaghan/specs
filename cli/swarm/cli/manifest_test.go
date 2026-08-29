// Tests for `swarm manifest get`, `set`, `require` and `validate`.
//
// They drive the real command tree with a fake process runner, so what
// the tracker would have been asked for is read off the fake's own record
// and no test starts a process.
//
// The end-to-end cases these commands owe — AC-SWARM-CORE-001.6 and
// AC-SWARM-CORE-002.16, both declared as e2e — are not here. The
// blueprint's sequencing says so: "WHILE the FIX delta is outstanding, a
// command delta's acceptance rests on its unit tests, and its end-to-end
// case is added when the fixture exists." The fixture is the next item.
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
)

const testRun = "agent-plugins-run1"

// runManifest drives the real tree, with the manifest commands registered
// against a fake process runner.
func runManifest(t *testing.T, fake *exectest.Fake, args ...string) (stdout string, code int) {
	t.Helper()

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut, exitCode: int(StatusClear)}
	root := newRootCmd(streams)
	root.AddCommand(manifestCmd(streams, manifestbeads.Store{Tracker: beadstracker.Tracker{}}, fake))

	code = dispatch(root, streams, args, &out, &errOut)
	return out.String(), code
}

func epic(metadata string) []byte {
	return []byte(`[{"id":"` + testRun + `","status":"open","metadata":` + metadata + `}]`)
}

func reading(metadata string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "show", testRun, "--json"},
		Stdout: epic(metadata),
	}
}

// A manifest far enough along to satisfy every rule at the scope stage.
const wholeRecord = `{
  "checks_ids": ["format"],
  "checks_baseline": "everything passed",
  "checks_baseline_status": 0,
  "concurrency_cap": 20,
  "models": {"light":"haiku","standard":"sonnet","advanced":"opus","frontier":"fable"},
  "effort": "unrecorded",
  "ledger": ["a claim — proved by a command"],
  "target": "run-1.target",
  "admitted": ["run-1.1","run-1.2"],
  "excluded": [],
  "moved": [],
  "standing_auth": {"pull_request":{"value":"draft","source":"default"},
    "move_foreign_beads":{"value":false,"source":"default"},
    "admit_review_findings":{"value":false,"source":"default"},
    "tier_escalation":{"value":"once","source":"default"},
    "review_budget":{"value":2,"source":"default"},
    "merge_to_default":{"value":false,"source":"default"}},
  "review_budget": 2
}`

func envelopeOf(t *testing.T, stdout string) swarmmodel.CommandResult {
	t.Helper()

	var env swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("the result does not parse as the envelope: %v\n%s", err, stdout)
	}
	return env
}

func factNamed(env swarmmodel.CommandResult, name string) (swarmmodel.Fact, bool) {
	for _, f := range env.Facts {
		if f.Name == name {
			return f, true
		}
	}
	return swarmmodel.Fact{}, false
}

func TestGetReportsAKeyWithTheKindItsValueHas(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x","review_passes":2,
	  "admitted":["run-1.1","run-1.2"],"models":{"light":"haiku","standard":"sonnet",
	  "advanced":"opus","frontier":"fable"}}`))

	stdout, code := runManifest(t, fake,
		"manifest", "get", "--run", testRun, "--json",
		"--key", "branch", "--key", "review_passes", "--key", "admitted", "--key", "models")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	env := envelopeOf(t, stdout)
	for name, wantKind := range map[string]swarmmodel.FactKind{
		"branch":        swarmmodel.FactKindText,
		"review_passes": swarmmodel.FactKindCount,
		"admitted":      swarmmodel.FactKindList,
		// An object has no fact kind of its own, so it is reported as text
		// carrying its JSON rather than as a fifth kind the schema does
		// not declare.
		"models": swarmmodel.FactKindText,
	} {
		fact, ok := factNamed(env, name)
		if !ok {
			t.Errorf("no fact for %q", name)
			continue
		}
		if fact.Value == nil || fact.Value.Kind != wantKind {
			t.Errorf("%q is reported as %v, want kind %q", name, fact.Value, wantKind)
		}
	}
}

// AC: AC-SWARM-CORE-002.3
//
// A key that is not written is reported as unestablished, and the command
// exits asking for attention with what would resolve it — never as an
// empty value.
func TestGetReportsAnAbsentKeyAsUnestablished(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

	stdout, code := runManifest(t, fake,
		"manifest", "get", "--run", testRun, "--json", "--key", "checks_baseline")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusAttention, stdout)
	}

	env := envelopeOf(t, stdout)
	fact, ok := factNamed(env, "checks_baseline")
	if !ok {
		t.Fatal("the absent key is not reported at all")
	}
	if fact.Established {
		t.Error("an absent key is reported as established")
	}
	if fact.Value != nil {
		t.Errorf("an absent key carries a value: %v", fact.Value)
	}
	if !strings.Contains(env.Resolution, "swarm manifest set") {
		t.Errorf("the result does not name what would resolve it: %q", env.Resolution)
	}
}

// A key nothing declares is a malformed invocation, refused before
// anything is read — the caller has to fix the invocation, not the run.
func TestGetRefusesAnUndeclaredKeyWithoutReadingAnything(t *testing.T) {
	fake := exectest.NewFake(reading(`{}`))

	stdout, code := runManifest(t, fake,
		"manifest", "get", "--run", testRun, "--key", "circuit_breaker")
	if code != int(StatusMalformed) {
		t.Errorf("exit status = %d, want %d", code, StatusMalformed)
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("an undeclared key still read the tracker: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "circuit_breaker") {
		t.Errorf("the refusal does not name the key: %s", stdout)
	}
}

// AC: AC-SWARM-CORE-002.13
//
// Several keys written in one invocation are applied as one merge: one
// read and one write, whatever the number of keys.
func TestSetAppliesSeveralKeysAsOneMerge(t *testing.T) {
	fake := exectest.NewFake(
		reading(`{}`),
		exectest.Recording{
			Argv: []string{"bd", "update", testRun, "--metadata",
				`{"base_sha":"3a57dc7","branch":"claude/branch-x","review_passes":0}`, "--json"},
			Stdout: epic(`{"branch":"claude/branch-x","base_sha":"3a57dc7","review_passes":0}`),
		},
	)

	stdout, code := runManifest(t, fake,
		"manifest", "set", "--run", testRun, "--json",
		"--key", "branch", "--value", `"claude/branch-x"`,
		"--key", "base_sha", "--value", `"3a57dc7"`,
		"--key", "review_passes", "--value", `0`)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	var writes int
	for _, argv := range fake.Argvs() {
		if strings.Join(argv, " ") != "" && contains(argv, "--metadata") {
			writes++
		}
	}
	if writes != 1 {
		t.Errorf("a three-key set made %d writes, want 1", writes)
	}
	if len(fake.Argvs()) != 2 {
		t.Errorf("a three-key set made %d invocations, want a guard and a write: %v",
			len(fake.Argvs()), fake.Argvs())
	}

	env := envelopeOf(t, stdout)
	for _, name := range []string{"branch", "base_sha", "review_passes"} {
		if _, ok := factNamed(env, name); !ok {
			t.Errorf("the result does not report %q", name)
		}
	}
}

// AC: AC-SWARM-CORE-002.14
//
// A multi-key write that fails on one key stores none of them, and the
// refusal names the offending key. The refusal happens before anything is
// read, because a value that is not of its key's type is an argument
// error rather than a state the run is in.
func TestSetRefusesTheWholeWriteWhenOneKeyIsRefused(t *testing.T) {
	fake := exectest.NewFake(reading(`{}`))

	stdout, code := runManifest(t, fake,
		"manifest", "set", "--run", testRun,
		"--key", "branch", "--value", `"claude/branch-x"`,
		"--key", "review_passes", "--value", `"two"`)

	if code != int(StatusMalformed) {
		t.Errorf("exit status = %d, want %d", code, StatusMalformed)
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("a refused write still touched the tracker: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "review_passes") {
		t.Errorf("the refusal does not name the offending key: %s", stdout)
	}
	if !strings.Contains(stdout, "storing none of it") {
		t.Errorf("the refusal does not say nothing was stored: %s", stdout)
	}
}

// A rule only visible against the record is checked where the record is
// read, and stops the write — the branch is written once.
func TestSetRefusesAWriteBreakingARuleOnlyVisibleAgainstTheRecord(t *testing.T) {
	fake := exectest.NewFake(
		reading(`{"branch":"claude/bead-bonanza-run1"}`),
		// Declared, so a command ignoring the guard would find something
		// to run rather than failing for an unrelated reason.
		exectest.Recording{
			Argv:   []string{"bd", "update", testRun, "--metadata", `{"branch":"claude/other"}`, "--json"},
			Stdout: epic(`{"branch":"claude/other"}`),
		},
	)

	stdout, code := runManifest(t, fake,
		"manifest", "set", "--run", testRun, "--key", "branch", "--value", `"claude/other"`)

	if code == int(StatusClear) {
		t.Errorf("a second branch was accepted\n%s", stdout)
	}
	for _, argv := range fake.Argvs() {
		if contains(argv, "--metadata") {
			t.Fatalf("the write ran despite the rule: %v", argv)
		}
	}
	if !strings.Contains(stdout, "written once") {
		t.Errorf("the refusal does not say why: %s", stdout)
	}
}

// --if-absent writes only when nobody has, which is what makes a resumed
// run reuse the branch it already recorded rather than race for it.
func TestSetIfAbsentRefusesAKeyAlreadyWritten(t *testing.T) {
	fake := exectest.NewFake(
		reading(`{"branch":"claude/bead-bonanza-run1"}`),
		exectest.Recording{
			Argv:   []string{"bd", "update", testRun, "--metadata", `{"branch":"claude/bead-bonanza-run1"}`, "--json"},
			Stdout: epic(`{"branch":"claude/bead-bonanza-run1"}`),
		},
	)

	stdout, code := runManifest(t, fake, "manifest", "set", "--run", testRun,
		"--if-absent", "--key", "branch", "--value", `"claude/bead-bonanza-run1"`)

	if code == int(StatusClear) {
		t.Errorf("--if-absent wrote a key that was already written\n%s", stdout)
	}
	for _, argv := range fake.Argvs() {
		if contains(argv, "--metadata") {
			t.Fatalf("--if-absent still wrote: %v", argv)
		}
	}
	if !strings.Contains(stdout, "already written") {
		t.Errorf("the refusal does not say why: %s", stdout)
	}
}

// AC: AC-SWARM-CORE-002.18
//
// A precondition check names every absent key, not the first.
func TestRequireNamesEveryAbsentKey(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

	stdout, code := runManifest(t, fake, "manifest", "require", "--run", testRun, "--json",
		"--key", "branch", "--key", "base_sha", "--key", "target")

	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusAttention, stdout)
	}

	env := envelopeOf(t, stdout)
	for _, missing := range []string{"base_sha", "target"} {
		if !strings.Contains(env.Summary, missing) {
			t.Errorf("the summary does not name %q: %q", missing, env.Summary)
		}
		fact, ok := factNamed(env, missing)
		if !ok || fact.Value == nil || fact.Value.Flag == nil || *fact.Value.Flag {
			t.Errorf("%q is not reported as not written: %v", missing, fact)
		}
	}
	present, ok := factNamed(env, "branch")
	if !ok || present.Value == nil || present.Value.Flag == nil || !*present.Value.Flag {
		t.Errorf("the key that IS written is not reported as written: %v", present)
	}

	// Every key present is a clear answer, so the command is not simply
	// "always attention".
	whole := exectest.NewFake(reading(`{"branch":"b","base_sha":"s","target":"t"}`))
	if _, code := runManifest(t, whole, "manifest", "require", "--run", testRun,
		"--key", "branch", "--key", "base_sha", "--key", "target"); code != int(StatusClear) {
		t.Errorf("exit status = %d with every key present, want %d", code, StatusClear)
	}
}

// AC: AC-SWARM-CORE-002.16
//
// A manifest whose keys are all present, correctly typed and mutually
// consistent validates clean. Asserted here against the real command with
// a faked tracker; the end-to-end case against the committed fixture is
// the fixture item's.
func TestValidateIsCleanForARecordThatSatisfiesEveryRule(t *testing.T) {
	fake := exectest.NewFake(reading(wholeRecord))

	stdout, code := runManifest(t, fake, "manifest", "validate", "--run", testRun,
		"--json", "--stage", "scope")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	env := envelopeOf(t, stdout)
	stage, ok := factNamed(env, "stage")
	if !ok || stage.Value == nil || stage.Value.Text == nil || *stage.Value.Text != "scope" {
		t.Errorf("the result does not say which stage it validated against: %v", stage)
	}
}

// AC: AC-SWARM-CORE-002.15
//
// Validation reports every rule that fails, not the first.
func TestValidateReportsEveryFailingRule(t *testing.T) {
	fake := exectest.NewFake(reading(`{
	  "admitted": ["run-1.1"],
	  "excluded": [{"id":"run-1.1","reason":"declined"}],
	  "review_passes": 5,
	  "review_budget": 2
	}`))

	stdout, code := runManifest(t, fake, "manifest", "validate", "--run", testRun,
		"--json", "--stage", "scope")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusAttention, stdout)
	}

	env := envelopeOf(t, stdout)
	for _, want := range []string{"admission-overlap", "counter-over-budget"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the result does not cite %s:\n%s", want, stdout)
		}
	}
	if !strings.Contains(env.Summary, "rule(s) broken") {
		t.Errorf("the summary does not say how many rules broke: %q", env.Summary)
	}
}

// With no --stage the command validates against the stage the record has
// reached, and says so — a caller reading a clean result has to know what
// it was clean against.
func TestValidateDerivesTheStageAndSaysThatItDid(t *testing.T) {
	fake := exectest.NewFake(reading(wholeRecord))

	stdout, code := runManifest(t, fake, "manifest", "validate", "--run", testRun, "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	env := envelopeOf(t, stdout)
	stage, _ := factNamed(env, "stage")
	if stage.Value == nil || stage.Value.Text == nil || *stage.Value.Text != "scope" {
		t.Errorf("the derived stage is %v, want scope", stage.Value)
	}
	if _, ok := factNamed(env, "stage_source"); !ok {
		t.Error("the result does not say the stage was derived")
	}

	// A stage the vocabulary does not know is an argument error, refused
	// before anything is read.
	empty := exectest.NewFake(reading(wholeRecord))
	out, code := runManifest(t, empty, "manifest", "validate", "--run", testRun, "--stage", "halfway")
	if code != int(StatusMalformed) {
		t.Errorf("an unknown stage exited %d, want %d", code, StatusMalformed)
	}
	if len(empty.Argvs()) != 0 {
		t.Errorf("an unknown stage still read the tracker: %v", empty.Argvs())
	}
	if !strings.Contains(out, "pre-flight") {
		t.Errorf("the refusal does not name the stages there are: %s", out)
	}
}

// AC: AC-SWARM-CORE-001.6
//
// A dry run reads live state and writes nothing: the guard runs, the
// write is refused, and the report reflects what the read returned.
func TestADryRunOfSetReadsAndWritesNothing(t *testing.T) {
	fake := exectest.NewFake(
		reading(`{"review_passes":1,"review_budget":2}`),
		exectest.Recording{
			Argv:   []string{"bd", "update", testRun, "--metadata", `{"review_passes":2}`, "--json"},
			Stdout: epic(`{"review_passes":2}`),
		},
	)

	stdout, code := runManifest(t, fake, "manifest", "set", "--run", testRun, "--dry-run",
		"--key", "review_passes", "--value", `2`)

	ran := fake.Argvs()
	if len(ran) != 1 || !contains(ran[0], "show") {
		t.Fatalf("a dry run ran %v, want only the read", ran)
	}
	if code == int(StatusMalformed) {
		t.Errorf("a dry run was refused as malformed\n%s", stdout)
	}
	if !strings.Contains(stdout, "refused") {
		t.Errorf("the report does not say the write was refused:\n%s", stdout)
	}
}

// Explain runs nothing at all, and prints both invocations the write
// would make — including the payload, which is known before anything is
// read.
func TestExplainOfSetPrintsBothInvocationsAndRunsNothing(t *testing.T) {
	fake := exectest.NewFake() // no recordings: anything executed errors

	stdout, code := runManifest(t, fake, "manifest", "set", "--run", testRun, "--explain",
		"--key", "review_passes", "--value", `2`)

	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("explain executed %v", fake.Argvs())
	}
	for _, want := range []string{"bd show " + testRun, "bd update " + testRun, "--metadata"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the transcript does not carry %q:\n%s", want, stdout)
		}
	}
}

// Every one of the four commands refuses an empty run scope before it
// reads or writes anything.
func TestEveryManifestCommandRefusesAnEmptyScope(t *testing.T) {
	for _, verb := range [][]string{
		{"get"},
		{"set", "--key", "branch", "--value", `"x"`},
		{"require", "--key", "branch"},
		{"validate"},
	} {
		fake := exectest.NewFake(reading(`{}`))
		args := append([]string{"manifest"}, verb...)

		_, code := runManifest(t, fake, args...)
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if len(fake.Argvs()) != 0 {
			t.Errorf("%v read or wrote something outside the scope: %v", args, fake.Argvs())
		}
	}
}

func contains(argv []string, want string) bool {
	for _, a := range argv {
		if a == want {
			return true
		}
	}
	return false
}
