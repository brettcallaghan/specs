// Tests for the store: the invocations it declares, and the conditional
// write it exists to provide.
//
// No test starts a process. The store reaches beads through the Beads
// Tracker, whose steps run through the fake process runner, so what the
// record would have been written with is read off the fake's own record
// of what it was asked for.
package beads_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
)

const runEpic = "agent-plugins-run1"

func store() manifestbeads.Store {
	return manifestbeads.Store{Tracker: beadstracker.Tracker{}}
}

func raw(values map[string]string) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(values))
	for k, v := range values {
		out[k] = json.RawMessage(v)
	}
	return out
}

// issue is what `bd show` returns for the run epic, carrying the metadata
// the record lives in.
func issue(metadata string) []byte {
	return []byte(`[{"id":"` + runEpic + `","status":"open","metadata":` + metadata + `}]`)
}

func showRecording(metadata string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "show", runEpic, "--json"},
		Stdout: issue(metadata),
	}
}

func argvs(fake *exectest.Fake) []string {
	var out []string
	for _, argv := range fake.Argvs() {
		out = append(out, strings.Join(argv, " "))
	}
	return out
}

func TestReadingTheRecordDeclaresOneInvocationAndYieldsADocument(t *testing.T) {
	fake := exectest.NewFake(showRecording(`{"branch":"claude/branch-x","review_passes":1}`))
	steps := store().Read("read", runEpic)

	if len(steps) != 1 {
		t.Fatalf("Read declares %d steps, want 1", len(steps))
	}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	value, err := steps.ValueFrom(report)
	if err != nil {
		t.Fatalf("ValueFrom: %v", err)
	}
	doc, ok := value.(manifest.Document)
	if !ok {
		t.Fatalf("the answer is %T, not a manifest document", value)
	}
	if got, _ := doc.Get("branch"); string(got) != `"claude/branch-x"` {
		t.Errorf("branch = %s", got)
	}
}

// AC: AC-SWARM-CORE-002.1
//
// The payload names only the keys the write covers. That is what makes
// "the other keys retain their values" structural rather than hopeful:
// the record's own store merges at the top level, so a key the payload
// does not mention cannot be touched, whoever wrote it.
func TestAWriteNamesOnlyTheKeysItCoversSoTheOthersCannotBeTouched(t *testing.T) {
	const before = `{"branch":"claude/branch-x","review_passes":1,"target":"run-1.target"}`
	fake := exectest.NewFake(
		showRecording(before),
		exectest.Recording{
			Argv:   []string{"bd", "update", runEpic, "--metadata", `{"review_passes":2}`, "--json"},
			Stdout: issue(`{"branch":"claude/branch-x","review_passes":2,"target":"run-1.target"}`),
		},
	)

	base := manifest.NewDocument(raw(map[string]string{
		"branch": `"claude/branch-x"`, "review_passes": `1`, "target": `"run-1.target"`,
	}))
	changes := raw(map[string]string{"review_passes": `2`})

	steps := store().Write("write", runEpic, changes, manifest.Expect(base, changes))
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.Failed() {
		t.Fatalf("the write failed: %+v", report.Records)
	}

	ran := argvs(fake)
	if len(ran) != 2 {
		t.Fatalf("the write ran %d invocations, want a guard and a write: %v", len(ran), ran)
	}
	if !strings.Contains(ran[1], `{"review_passes":2}`) {
		t.Errorf("the payload is not just the covered key: %s", ran[1])
	}
	for _, untouched := range []string{"branch", "target"} {
		if strings.Contains(ran[1], untouched) {
			t.Errorf("the payload names %q, which the write does not cover: %s", untouched, ran[1])
		}
	}
}

// AC: AC-SWARM-CORE-002.10
//
// A write that would clobber a concurrent change is refused and names the
// key that moved. The guard is a read whose decoder does the comparison,
// so the plan's own ordering policy is what stops the write — there is no
// conjunction to forget.
func TestAWriteThatWouldClobberAConcurrentChangeIsRefusedNamingTheKey(t *testing.T) {
	// The caller read review_passes as 1; by the time it writes, someone
	// else has made it 2.
	fake := exectest.NewFake(
		showRecording(`{"branch":"claude/branch-x","review_passes":2}`),
		// Declared, so a store ignoring the guard would find something to
		// run rather than failing for an unrelated reason.
		exectest.Recording{
			Argv:   []string{"bd", "update", runEpic, "--metadata", `{"review_passes":3}`, "--json"},
			Stdout: issue(`{"review_passes":3}`),
		},
	)

	base := manifest.NewDocument(raw(map[string]string{"review_passes": `1`}))
	changes := raw(map[string]string{"review_passes": `3`})

	steps := store().Write("write", runEpic, changes, manifest.Expect(base, changes))
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, argv := range argvs(fake) {
		if strings.Contains(argv, "--metadata") {
			t.Fatalf("the write ran despite the key having moved: %s", argv)
		}
	}
	if report.Stopped != "write-guard" {
		t.Errorf("Report.Stopped = %q, want the guard", report.Stopped)
	}

	guard, _ := report.Record("write-guard")
	if guard.Err == nil {
		t.Fatal("the guard records no reason")
	}
	if !strings.Contains(guard.Err.Error(), "review_passes") {
		t.Errorf("the refusal does not name the key that moved: %v", guard.Err)
	}
}

// The negative control for the guard: a record nobody else touched is
// written. Without it a store that refused every write would satisfy the
// test above.
func TestAWriteAgainstAnUnchangedRecordGoesThrough(t *testing.T) {
	fake := exectest.NewFake(
		showRecording(`{"review_passes":1}`),
		exectest.Recording{
			Argv:   []string{"bd", "update", runEpic, "--metadata", `{"review_passes":2}`, "--json"},
			Stdout: issue(`{"review_passes":2}`),
		},
	)

	base := manifest.NewDocument(raw(map[string]string{"review_passes": `1`}))
	changes := raw(map[string]string{"review_passes": `2`})

	steps := store().Write("write", runEpic, changes, manifest.Expect(base, changes))
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.Failed() {
		t.Fatalf("the write failed: %+v", report.Records)
	}

	value, err := steps.ValueFrom(report)
	if err != nil {
		t.Fatalf("ValueFrom: %v", err)
	}
	doc := value.(manifest.Document)
	if got, _ := doc.Get("review_passes"); string(got) != `2` {
		t.Errorf("the record as written reads %s", got)
	}
}

// A key that appeared after the caller read is a concurrent change too,
// not merely a value that differs.
func TestAKeyThatAppearedAfterTheReadIsAConcurrentChange(t *testing.T) {
	fake := exectest.NewFake(
		showRecording(`{"findings_path":"docs/retros/theirs.md"}`),
		exectest.Recording{
			Argv:   []string{"bd", "update", runEpic, "--metadata", `{"findings_path":"docs/retros/ours.md"}`, "--json"},
			Stdout: issue(`{"findings_path":"docs/retros/ours.md"}`),
		},
	)

	base := manifest.NewDocument(nil) // the caller saw no findings_path
	changes := raw(map[string]string{"findings_path": `"docs/retros/ours.md"`})

	steps := store().Write("write", runEpic, changes, manifest.Expect(base, changes))
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, argv := range argvs(fake) {
		if strings.Contains(argv, "--metadata") {
			t.Fatal("the write overwrote a key another writer had added")
		}
	}
	guard, _ := report.Record("write-guard")
	if guard.Err == nil || !strings.Contains(guard.Err.Error(), "findings_path") {
		t.Errorf("the refusal does not name the key that appeared: %v", guard.Err)
	}
}

// AC: AC-SWARM-CORE-002.13
//
// Several keys written in one invocation are one read and one write,
// however many keys there are.
func TestSeveralKeysInOneInvocationAreOneReadAndOneWrite(t *testing.T) {
	changes := raw(map[string]string{
		"branch":        `"claude/branch-x"`,
		"base_sha":      `"3a57dc7"`,
		"review_passes": `0`,
		"target":        `"run-1.target"`,
	})

	steps := store().Write("write", runEpic, changes, manifest.Document{})
	if len(steps) != 2 {
		t.Fatalf("a four-key write declares %d steps, want 2", len(steps))
	}

	// And the whole fragment is renderable before anything runs, which is
	// what explain mode needs: the payload does not depend on the read.
	var out strings.Builder
	if _, err := (plan.Runner{Out: &out}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExplain); err != nil {
		t.Fatalf("explain: %v", err)
	}
	for _, want := range []string{"bd show", "bd update", "--metadata"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the explained fragment does not carry %q:\n%s", want, out.String())
		}
	}
}

// The write is a write and the guard is a read, so a dry run performs the
// guard and refuses the write — a dry run that wrote the record would be
// the worst possible bug in this component.
func TestADryRunPerformsTheGuardAndRefusesTheWrite(t *testing.T) {
	fake := exectest.NewFake(
		showRecording(`{"review_passes":1}`),
		exectest.Recording{
			Argv:   []string{"bd", "update", runEpic, "--metadata", `{"review_passes":2}`, "--json"},
			Stdout: issue(`{"review_passes":2}`),
		},
	)

	base := manifest.NewDocument(raw(map[string]string{"review_passes": `1`}))
	changes := raw(map[string]string{"review_passes": `2`})

	steps := store().Write("write", runEpic, changes, manifest.Expect(base, changes))
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeDryRun)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}

	ran := argvs(fake)
	if len(ran) != 1 || !strings.Contains(ran[0], "bd show") {
		t.Fatalf("a dry run ran %v, want only the guard", ran)
	}
	write, _ := report.Record("write-write")
	if write.Status != plan.StepRefused {
		t.Errorf("the write's status = %q, want %q", write.Status, plan.StepRefused)
	}
}

// A run epic with no manifest yet reads as an empty record rather than as
// a failure — every run starts there.
func TestARunEpicWithNoManifestReadsAsAnEmptyRecord(t *testing.T) {
	fake := exectest.NewFake(showRecording(`{}`))
	steps := store().Read("read", runEpic)

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	value, err := steps.ValueFrom(report)
	if err != nil {
		t.Fatalf("ValueFrom: %v", err)
	}
	if names := value.(manifest.Document).Names(); len(names) != 0 {
		t.Errorf("an empty record read as %v", names)
	}
}

// A tracker that cannot answer is a failure, not an empty record. The
// difference matters: an empty record means "this run has written
// nothing yet", and reporting a failed read that way would let a run
// proceed as if it were new.
func TestATrackerThatCannotAnswerIsAFailureRatherThanAnEmptyRecord(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:       []string{"bd", "show", runEpic, "--json"},
		Stdout:     []byte(`{"error":"no issues found matching the provided IDs","schema_version":1}`),
		ExitStatus: 1,
	})

	steps := store().Read("read", runEpic)
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !report.Failed() {
		t.Fatal("a failed read is reported as a successful one")
	}
	if _, err := steps.ValueFrom(report); err == nil {
		t.Error("a failed read produced an answer")
	}
}
