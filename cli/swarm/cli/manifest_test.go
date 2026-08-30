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

	"github.com/goccy/go-yaml"

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

// runManifestStreams is runManifest with the two streams kept apart, so
// a test can assert that the answer is on stdout and that nothing else
// reached the caller.
func runManifestStreams(t *testing.T, fake *exectest.Fake, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut, exitCode: int(StatusClear)}
	root := newRootCmd(streams)
	root.AddCommand(manifestCmd(streams, manifestbeads.Store{Tracker: beadstracker.Tracker{}}, fake))

	code = dispatch(root, streams, args, &out, &errOut)
	return out.String(), errOut.String(), code
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
		"manifest", "get", "--run", testRun, "--json", "--full-result",
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
		"manifest", "get", "--run", testRun, "--json", "--full-result", "--key", "checks_baseline")
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

// AC: AC-SWARM-CORE-002.20
//
// A clear answer is the properties and their values and nothing else: no
// header, no transcript, no summary, no outcome line.
func TestAClearAnswerIsThePropertiesAndTheirValues(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

	stdout, stderr, code := runManifestStreams(t, fake, "manifest", "get", "--run", testRun, "--key", "branch")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s\n%s", code, StatusClear, stdout, stderr)
	}
	if stdout != "branch: claude/branch-x\n" {
		t.Errorf("stdout = %q, want the property and its value alone", stdout)
	}
	if stderr != "" {
		t.Errorf("the answer wrote to stderr: %q", stderr)
	}
}

// AC: AC-SWARM-CORE-002.20
//
// A value that a shell would mangle has to arrive intact: whitespace kept,
// an embedded newline kept, a leading dash neither eaten nor escaped.
//
// Each case asserts the exact bytes AND parses them back with a YAML
// reader, so the rendering is checked against what it means rather than
// only against what it looks like.
func TestAValueAShellWouldMangleSurvivesTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stored string // as JSON, the way the tracker holds it
		value  string // what it holds
		want   string // the exact stdout, newline included
	}{
		{"inner and outer whitespace", `"  two  words\t "`, "  two  words\t ",
			"findings_path: \"  two  words\\t \"\n"},
		{"an embedded newline", `"first\nsecond"`, "first\nsecond",
			"findings_path: |-\n  first\n  second\n"},
		{"a leading dash", `"--not-a-flag"`, "--not-a-flag",
			"findings_path: \"--not-a-flag\"\n"},
		{"a trailing newline of its own", `"trailing\n"`, "trailing\n",
			"findings_path: |\n  trailing\n"},
		{"two trailing newlines", `"trailing\n\n"`, "trailing\n\n",
			"findings_path: |+\n  trailing\n\n"},
		{"a lone dash", `"-"`, "-", "findings_path: \"-\"\n"},
		{"quotes and a backslash", `"say \"hi\" \\ bye"`, `say "hi" \ bye`,
			"findings_path: say \"hi\" \\ bye\n"},
		{"a word that reads as a number", `"1.0"`, "1.0", "findings_path: \"1.0\"\n"},
		{"a word that reads as a boolean", `"no"`, "no", "findings_path: \"no\"\n"},
		{"a line whose first character is a space", `"  indented\nplain"`, "  indented\nplain",
			"findings_path: |2-\n    indented\n  plain\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := exectest.NewFake(reading(`{"findings_path":` + tc.stored + `}`))

			stdout, stderr, code := runManifestStreams(t, fake,
				"manifest", "get", "--run", testRun, "--key", "findings_path")
			if code != int(StatusClear) {
				t.Fatalf("exit status = %d, want %d\n%s\n%s", code, StatusClear, stdout, stderr)
			}
			if stdout != tc.want {
				t.Errorf("stdout = %q, want %q", stdout, tc.want)
			}

			var got map[string]any
			if err := yaml.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("the answer does not parse as YAML: %v\n%s", err, stdout)
			}
			if got["findings_path"] != tc.value {
				t.Errorf("the value read back as %q, want %q", got["findings_path"], tc.value)
			}
		})
	}
}

// --json is a syntax and not a depth: the same answer, parseable, with no
// envelope around it.
func TestTheAnswerAsJSONCarriesTheSameValue(t *testing.T) {
	fake := exectest.NewFake(reading(`{"findings_path":"a b\nc"}`))

	stdout, _, code := runManifestStreams(t, fake,
		"manifest", "get", "--run", testRun, "--key", "findings_path", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("the answer does not parse as JSON: %v\n%s", err, stdout)
	}
	if got["findings_path"] != "a b\nc" {
		t.Errorf("the value round-tripped as %q", got["findings_path"])
	}
}

// A value is rendered as its own type: a number unquoted, a list as a
// sequence, a boolean as true or false.
func TestAValueRendersAsItsOwnType(t *testing.T) {
	for _, tc := range []struct{ key, stored, want string }{
		{"review_passes", `2`, "review_passes: 2\n"},
		{"admitted", `["run-1.1","run-1.2"]`, "admitted:\n  - run-1.1\n  - run-1.2\n"},
		{"excluded", `[]`, "excluded: []\n"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			fake := exectest.NewFake(reading(`{"` + tc.key + `":` + tc.stored + `}`))

			stdout, _, code := runManifestStreams(t, fake,
				"manifest", "get", "--run", testRun, "--key", tc.key)
			if code != int(StatusClear) {
				t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
			}
			if stdout != tc.want {
				t.Errorf("stdout = %q, want %q", stdout, tc.want)
			}
		})
	}
}

// AC: AC-SWARM-CORE-002.21
//
// No key asks for the record, so the answer is every key it carries.
func TestGetAnswersWithTheWholeRecordWhenNoKeyIsGiven(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x","review_passes":2}`))

	stdout, stderr, code := runManifestStreams(t, fake, "manifest", "get", "--run", testRun)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s\n%s", code, StatusClear, stdout, stderr)
	}

	var got map[string]any
	if err := yaml.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("the answer does not parse as YAML: %v\n%s", err, stdout)
	}
	if len(got) != 2 || got["branch"] != "claude/branch-x" || got["review_passes"] != uint64(2) {
		t.Errorf("the answer is not the record: %#v", got)
	}
}

// Several keys ask for part of the record, so the answer carries exactly
// those keys.
func TestGetAnswersWithJustTheKeysAskedFor(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x","review_passes":2,"target":"t.target"}`))

	stdout, _, code := runManifestStreams(t, fake,
		"manifest", "get", "--run", testRun, "--key", "branch", "--key", "review_passes")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	var got map[string]any
	if err := yaml.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("the answer does not parse as YAML: %v\n%s", err, stdout)
	}
	if len(got) != 2 {
		t.Fatalf("the answer carries %d keys, want the 2 that were asked for: %v", len(got), got)
	}
	if _, unwanted := got["target"]; unwanted {
		t.Errorf("a key nobody asked for is in the answer: %v", got)
	}
}

// AC: AC-SWARM-CORE-002.22
//
// The three answers stay distinguishable by exit status, and the shape
// follows the status: clear renders the properties, anything else renders
// a verdict per property, and a refusal that named no property renders
// `issues` alone. All three on stdout, so one redirect captures whichever
// arrives.
func TestTheThreeAnswersStayApart(t *testing.T) {
	for _, tc := range []struct {
		name    string
		key     string
		want    Status
		wantOut string
	}{
		{"established", "branch", StatusClear, "branch: claude/branch-x\n"},
		{"declared and not written", "checks_baseline", StatusAttention,
			"checks_baseline:\n  ok: false\n  value: null\n  issues:\n" +
				"    - the key is not written\n" +
				"    - \"write it with: swarm manifest set --run agent-plugins-run1" +
				" --key checks_baseline --value <value>\"\n"},
		{"not declared at all", "circuit_breaker", StatusMalformed,
			"issues:\n  - \"circuit_breaker: not a declared manifest key\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

			stdout, stderr, code := runManifestStreams(t, fake,
				"manifest", "get", "--run", testRun, "--key", tc.key)
			if code != int(tc.want) {
				t.Fatalf("exit status = %d, want %d\n%s\n%s", code, tc.want, stdout, stderr)
			}
			if stdout != tc.wantOut {
				t.Errorf("stdout = %q, want %q", stdout, tc.wantOut)
			}
			if stderr != "" {
				t.Errorf("the answer wrote to stderr: %q", stderr)
			}
		})
	}
}

// AC: AC-SWARM-CORE-002.22
//
// --full-result is the way back to the provenance, and it renders the
// CommandResult the contract declares: the summary, the fact and the
// invocation behind it.
func TestFullResultRestoresTheReport(t *testing.T) {
	fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

	stdout, _, code := runManifestStreams(t, fake,
		"manifest", "get", "--run", testRun, "--key", "branch", "--full-result", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	env := envelopeOf(t, stdout)
	fact, ok := factNamed(env, "branch")
	if !ok {
		t.Fatal("the full result carries no fact for the key")
	}
	if fact.Value == nil || fact.Value.Text == nil || *fact.Value.Text != "claude/branch-x" {
		t.Errorf("the fact does not carry the value: %v", fact.Value)
	}
	if len(fact.Invocations) == 0 {
		t.Error("the full result carries no invocation behind the fact")
	}
}

// Every subcommand carries --full-result. Each now has an answer smaller
// than its report, so each has a use for the flag, and a caller of any of
// them has a way back to the provenance.
func TestEverySubcommandCarriesFullResult(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"get", []string{"manifest", "get", "--run", testRun, "--key", "branch"}},
		{"require", []string{"manifest", "require", "--run", testRun, "--key", "branch"}},
		{"validate", []string{"manifest", "validate", "--run", testRun, "--stage", "pre-flight"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

			stdout, stderr, code := runManifestStreams(t, fake, append(tc.args, "--full-result", "--json")...)
			if code == int(StatusMalformed) {
				t.Fatalf("`%s --full-result` was refused as malformed\n%s\n%s", tc.name, stdout, stderr)
			}
			if env := envelopeOf(t, stdout); env.ContractVersion == "" {
				t.Errorf("`%s --full-result` did not render the command result", tc.name)
			}
		})
	}
}

// --explain and --dry-run imply --full-result. Both ask what the command
// would do, and what it would do is the plan — which lives in the full
// result. In explain mode nothing runs, so the answer alone would be an
// empty mapping.
func TestExplainAndDryRunImplyTheFullResult(t *testing.T) {
	for _, mode := range []string{"--explain", "--dry-run"} {
		t.Run(mode, func(t *testing.T) {
			fake := exectest.NewFake(reading(`{"branch":"claude/branch-x"}`))

			stdout, stderr, code := runManifestStreams(t, fake,
				"manifest", "get", "--run", testRun, "--key", "branch", mode, "--json")
			if code != int(StatusClear) {
				t.Fatalf("exit status = %d, want %d\n%s\n%s", code, StatusClear, stdout, stderr)
			}
			env := envelopeOf(t, stdout)
			if env.ContractVersion == "" {
				t.Fatalf("%s did not render the command result:\n%s", mode, stdout)
			}
			if len(env.Transcript) == 0 {
				t.Errorf("%s rendered no transcript, which is the only thing it produces", mode)
			}
		})
	}
}

// A key nothing declares is a malformed invocation, refused before
// anything is read — the caller has to fix the invocation, not the run.
func TestGetRefusesAnUndeclaredKeyWithoutReadingAnything(t *testing.T) {
	fake := exectest.NewFake(reading(`{}`))

	stdout, stderr, code := runManifestStreams(t, fake,
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
	if stderr != "" {
		t.Errorf("the refusal wrote to stderr: %q", stderr)
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
		"manifest", "set", "--run", testRun, "--json", "--full-result",
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

	stdout, code := runManifest(t, fake, "manifest", "require", "--run", testRun, "--json", "--full-result",
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
		if !ok || fact.Established || fact.Value != nil {
			t.Errorf("%q is not reported as not written: %v", missing, fact)
		}
	}
	// A property is the thing itself and never a verdict about it: the
	// key that IS written reports its VALUE, which is what `get` would
	// report, and whether every key is written is the exit status.
	present, ok := factNamed(env, "branch")
	if !ok || present.Value == nil || present.Value.Text == nil || *present.Value.Text != "claude/branch-x" {
		t.Errorf("the key that IS written does not report its value: %v", present)
	}

	// Every key present is a clear answer, so the command is not simply
	// "always attention".
	whole := exectest.NewFake(reading(`{"branch":"b","base_sha":"s","target":"t"}`))
	if _, code := runManifest(t, whole, "manifest", "require", "--run", testRun,
		"--key", "branch", "--key", "base_sha", "--key", "target"); code != int(StatusClear) {
		t.Errorf("exit status = %d with every key present, want %d", code, StatusClear)
	}
}

// AC: AC-SWARM-CORE-002.19
//
// A key with no stage is served by get, require and set like any other.
// Having no stage says its absence is legitimate at every stage; refusing
// it as undeclared would leave the run unable to ask the record for
// something the record is declared to carry. `trailers` is the one such
// key today, and it is named here so an empty derivation could not pass.
func TestTheVerbsServeAKeyThatHasNoStage(t *testing.T) {
	written := exectest.NewFake(reading(`{"trailers":"Co-Authored-By: Someone"}`))
	stdout, code := runManifest(t, written,
		"manifest", "get", "--run", testRun, "--json", "--full-result", "--key", "trailers")
	if code != int(StatusClear) {
		t.Fatalf("get --key trailers exited %d, want %d\n%s", code, StatusClear, stdout)
	}
	fact, ok := factNamed(envelopeOf(t, stdout), "trailers")
	if !ok || fact.Value == nil || fact.Value.Text == nil || *fact.Value.Text != "Co-Authored-By: Someone" {
		t.Errorf("get did not report the trailers it read: %v", fact)
	}

	// Required and written: clear. Required and unwritten: attention that
	// says how to write it — not "not a declared manifest key".
	if _, code := runManifest(t, exectest.NewFake(reading(`{"trailers":"x"}`)),
		"manifest", "require", "--run", testRun, "--key", "trailers"); code != int(StatusClear) {
		t.Errorf("require --key trailers exited %d with the key written, want %d", code, StatusClear)
	}
	stdout, code = runManifest(t, exectest.NewFake(reading(`{}`)),
		"manifest", "require", "--run", testRun, "--json", "--key", "trailers")
	if code != int(StatusAttention) {
		t.Fatalf("require --key trailers exited %d with the key absent, want %d\n%s",
			code, StatusAttention, stdout)
	}
	if strings.Contains(stdout, "not a declared manifest key") {
		t.Errorf("require reported trailers as undeclared: %s", stdout)
	}

	setting := exectest.NewFake(
		reading(`{}`),
		exectest.Recording{
			Argv: []string{"bd", "update", testRun, "--metadata",
				`{"trailers":"Co-Authored-By: Someone"}`, "--json"},
			Stdout: epic(`{"trailers":"Co-Authored-By: Someone"}`),
		},
	)
	stdout, code = runManifest(t, setting, "manifest", "set", "--run", testRun, "--json",
		"--key", "trailers", "--value", `"Co-Authored-By: Someone"`)
	if code != int(StatusClear) {
		t.Fatalf("set --key trailers exited %d, want %d\n%s", code, StatusClear, stdout)
	}

	// The declared type is still enforced, so serving the key is not the
	// same as accepting anything for it.
	stdout, code = runManifest(t, exectest.NewFake(reading(`{}`)),
		"manifest", "set", "--run", testRun, "--key", "trailers", "--value", `3`)
	if code != int(StatusMalformed) {
		t.Errorf("a numeric trailers value exited %d, want %d\n%s", code, StatusMalformed, stdout)
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
		"--json", "--full-result", "--stage", "scope")
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
		"--json", "--full-result", "--stage", "scope")
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

	stdout, code := runManifest(t, fake, "manifest", "validate", "--run", testRun, "--json", "--full-result")
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
