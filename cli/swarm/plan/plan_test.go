// Tests for the Plan Runner.
//
// Not one of them starts a process: the Process Runner is substituted by
// the fake from cli/swarm/exec/exectest, which is the property ADR-024
// selected this design for — "a plan is a value, so a unit test can
// assert the sequence a command builds without running a single
// subprocess, which is what makes the guard itself testable rather than
// only its effects".
//
// The fake is also what makes "no tool is executed" assertable rather
// than merely likely: it records every invocation it was asked for, so a
// mode that ran something it should not have leaves evidence.
package plan_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// Digests of the bytes the recordings below replay, computed
// independently with sha256sum rather than by calling the same function
// the code under test calls.
const (
	digestOfValue1  = "sha256:bb849f0fb2d26f92c65cab1b66f79cca463e119cc8b56ead875e54329ec93f2b"
	digestOfWarning = "sha256:86a10149acbe95519ea3283d26d57491ede09bd3010cdb756358f66352236b79"
	digestOfEmpty   = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// commandLines returns the lines of transcript that are invocations
// rather than comments, which is what "the invocations it printed" means.
func commandLines(transcript string) []string {
	var out []string
	for _, line := range strings.Split(transcript, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func readStep(name string, argv ...string) plan.Step {
	args := make([]plan.Arg, len(argv))
	for i, a := range argv {
		args[i] = plan.Lit(a)
	}
	return plan.Step{Name: name, Executable: "bd", Args: args, Effect: plan.Read, Repeatable: true}
}

// AC: AC-SWARM-CORE-001.4
//
// The mode exists to be safe against a repository nobody has looked at,
// so the assertion is in two halves: the transcript carries every
// invocation in order, and the process runner was never reached.
func TestExplainPrintsEveryInvocationInOrderAndExecutesNone(t *testing.T) {
	// Deliberately a fake with NO recordings: were any step executed, it
	// would both error and leave a call on the record.
	fake := exectest.NewFake()
	var out strings.Builder

	p := plan.Plan{Name: "accept a dispatch", Steps: []plan.Step{
		readStep("read-epic", "show", "run-1", "--json"),
		readStep("read-front", "ready", "--parent", "run-1"),
		{Name: "claim", Executable: "bd", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("update"), plan.Lit("run-1.4"), plan.Lit("--claim")}},
	}}

	if _, err := (plan.Runner{Exec: fake, Out: &out}).Run(context.Background(), p, plan.ModeExplain); err != nil {
		t.Fatalf("explain: %v", err)
	}

	got := commandLines(out.String())
	want := []string{
		"bd show run-1 --json",
		"bd ready --parent run-1",
		"bd update run-1.4 --claim",
	}
	if len(got) != len(want) {
		t.Fatalf("printed %d invocations, want %d:\n%s", len(got), len(want), out.String())
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("invocation %d = %q, want %q", i, got[i], want[i])
		}
	}

	if calls := fake.Argvs(); len(calls) != 0 {
		t.Errorf("explain executed %d invocation(s): %v", len(calls), calls)
	}
}

// AC: AC-SWARM-CORE-001.5
//
// Explain cannot know the value without running the step it exists not to
// run, so the argument is printed as the step that produces it.
func TestExplainPrintsAnArgumentOnlyAPriorStepCouldSupplyAsThatStep(t *testing.T) {
	var out strings.Builder

	p := plan.Plan{Steps: []plan.Step{
		readStep("read-branch", "show", "run-1", "--json"),
		{Name: "merge", Executable: "git", Effect: plan.Write, Args: []plan.Arg{
			plan.Lit("merge"), plan.Lit("--no-ff"), plan.From("read-branch", nil),
		}},
	}}

	if _, err := (plan.Runner{Out: &out}).Run(context.Background(), p, plan.ModeExplain); err != nil {
		t.Fatalf("explain: %v", err)
	}

	transcript := out.String()
	if !strings.Contains(transcript, "read-branch") {
		t.Errorf("the reference does not name the step that produces it:\n%s", transcript)
	}

	lines := commandLines(transcript)
	if len(lines) != 2 {
		t.Fatalf("printed %d invocations, want 2:\n%s", len(lines), transcript)
	}
	if got, want := lines[1], `git merge --no-ff '<from step: read-branch>'`; got != want {
		t.Errorf("the referencing invocation printed as\n  %s\nwant\n  %s", got, want)
	}
}

// AC: AC-SWARM-CORE-001.7
func TestExecutePrintsEachInvocationAsItRunsIt(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "show", "run-1", "--json"}, Stdout: []byte("value-1\n")},
		exectest.Recording{Argv: []string{"bd", "ready", "--parent", "run-1"}},
	)
	var out strings.Builder

	p := plan.Plan{Steps: []plan.Step{
		readStep("read-epic", "show", "run-1", "--json"),
		readStep("read-front", "ready", "--parent", "run-1"),
	}}

	report, err := (plan.Runner{Exec: fake, Out: &out}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.Failed() {
		t.Fatalf("the plan reported a failure: %+v", report.Records)
	}

	got := commandLines(out.String())
	want := []string{"bd show run-1 --json", "bd ready --parent run-1"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("printed:\n%v\nwant:\n%v", got, want)
	}
	if n := len(fake.Argvs()); n != 2 {
		t.Errorf("the runner was asked for %d invocations, want 2", n)
	}
}

// The ordering guard, and the reason it is a field rather than a shell
// conjunction: the incident behind this whole feature is a worktree
// removed after a merge exited 128.
func TestAStepWhoseFailurePolicyIsStopPreventsEveryLaterStep(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge", "dev-1"}, ExitStatus: 128, Stderr: []byte("conflict\n")},
		// Declared, so that a runner which ignored the guard would find
		// something to run rather than erroring for an unrelated reason.
		exectest.Recording{Argv: []string{"git", "worktree", "remove", "dev-1"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		{Name: "merge", Executable: "git", Effect: plan.Write, OnFailure: plan.Stop,
			Args: []plan.Arg{plan.Lit("merge"), plan.Lit("dev-1")}},
		{Name: "remove-worktree", Executable: "git", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("worktree"), plan.Lit("remove"), plan.Lit("dev-1")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, argv := range fake.Argvs() {
		if strings.Join(argv, " ") == "git worktree remove dev-1" {
			t.Fatal("the step after a stop-on-failure failure was executed")
		}
	}
	if report.Stopped != "merge" {
		t.Errorf("Report.Stopped = %q, want %q", report.Stopped, "merge")
	}
	rec, ok := report.Record("remove-worktree")
	if !ok {
		t.Fatal("the report does not account for the step that never ran")
	}
	if rec.Status != plan.StepSkipped {
		t.Errorf("the later step's status = %q, want %q", rec.Status, plan.StepSkipped)
	}
}

// The negative control for the guard. Without it, an interpreter that
// stopped after every failure — or after every step — would satisfy the
// test above.
func TestAStepDeclaringContinueOnFailureDoesNotPreventLaterSteps(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "search", "nothing"}, ExitStatus: 1},
		exectest.Recording{Argv: []string{"bd", "ready"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		{Name: "search", Executable: "bd", Effect: plan.Read, OnFailure: plan.Continue,
			Args: []plan.Arg{plan.Lit("search"), plan.Lit("nothing")}},
		readStep("ready", "ready"),
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.Stopped != "" {
		t.Errorf("Report.Stopped = %q, want empty", report.Stopped)
	}
	if n := len(fake.Argvs()); n != 2 {
		t.Errorf("%d invocations ran, want 2: %v", n, fake.Argvs())
	}
	if !report.Failed() {
		t.Error("Report.Failed() = false, though a step exited 1")
	}
}

func TestThereAreExactlyThreeInterpreters(t *testing.T) {
	modes := plan.Modes()
	if len(modes) != 3 {
		t.Fatalf("Modes() has %d entries, want 3: %v", len(modes), modes)
	}
	want := []string{"explain", "dry-run", "execute"}
	for i, m := range modes {
		if m.String() != want[i] {
			t.Errorf("mode %d = %q, want %q", i, m.String(), want[i])
		}
	}
}

// AC: AC-SWARM-CORE-001.6
//
// Dry-run runs the reads — which is what makes its report reflect live
// state rather than an imagined one — and refuses everything else.
func TestADryRunExecutesTheReadsAndRefusesTheWrites(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "show", "run-1"}, Stdout: []byte("value-1\n")},
		// Declared on purpose: if the interpreter ran the write, this
		// answers it, so the test fails on the assertion about what ran
		// rather than on an unrelated "no recording" error.
		exectest.Recording{Argv: []string{"bd", "update", "run-1", "--claim"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		readStep("read", "show", "run-1"),
		{Name: "claim", Executable: "bd", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("update"), plan.Lit("run-1"), plan.Lit("--claim")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeDryRun)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}

	argvs := fake.Argvs()
	if len(argvs) != 1 || strings.Join(argvs[0], " ") != "bd show run-1" {
		t.Fatalf("a dry run executed %v, want only the read", argvs)
	}

	read, _ := report.Record("read")
	if read.Status != plan.StepRan {
		t.Errorf("the read's status = %q, want %q", read.Status, plan.StepRan)
	}
	claim, _ := report.Record("claim")
	if claim.Status != plan.StepRefused {
		t.Errorf("the write's status = %q, want %q", claim.Status, plan.StepRefused)
	}
	if claim.Reason == "" {
		t.Error("a refused step gives no reason")
	}
}

func TestExplainRendersBothArmsAndNamesTheConditionBetweenThem(t *testing.T) {
	var out strings.Builder

	p := plan.Plan{Steps: []plan.Step{
		{Name: "merge", Executable: "git", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("merge"), plan.Lit("dev-1")},
			Alternatives: []plan.Alternative{{
				Condition: "the merge left conflicts",
				Steps: []plan.Step{
					{Name: "abort", Executable: "git", Effect: plan.Write,
						Args: []plan.Arg{plan.Lit("merge"), plan.Lit("--abort")}},
				},
			}}},
		{Name: "push", Executable: "git", Effect: plan.Write, Args: []plan.Arg{plan.Lit("push")}},
	}}

	if _, err := (plan.Runner{Out: &out}).Run(context.Background(), p, plan.ModeExplain); err != nil {
		t.Fatalf("explain: %v", err)
	}

	transcript := out.String()
	if !strings.Contains(transcript, "the merge left conflicts") {
		t.Errorf("the transcript does not name the condition:\n%s", transcript)
	}
	if !strings.Contains(transcript, "otherwise") {
		t.Errorf("the transcript does not name the other arm:\n%s", transcript)
	}
	for _, want := range []string{"git merge dev-1", "git merge --abort", "git push"} {
		if !strings.Contains(transcript, want) {
			t.Errorf("the transcript is missing %q:\n%s", want, transcript)
		}
	}
}

func TestEachStepRecordsItsInvocationStatusDurationAndStreamDigests(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:       []string{"bd", "show", "run-1"},
		Stdout:     []byte("value-1\n"),
		Stderr:     []byte("warning\n"),
		ExitStatus: 3,
		Duration:   250 * time.Millisecond,
	})

	p := plan.Plan{Steps: []plan.Step{
		{Name: "read", Executable: "bd", Effect: plan.Read, OnFailure: plan.Continue,
			Args: []plan.Arg{plan.Lit("show"), plan.Lit("run-1")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	rec, ok := report.Record("read")
	if !ok {
		t.Fatal("no record for the step")
	}
	if rec.Invocation.Name != "bd" || strings.Join(rec.Invocation.Args, " ") != "show run-1" {
		t.Errorf("the record does not carry the invocation: %+v", rec.Invocation)
	}
	if rec.ExitStatus != 3 {
		t.Errorf("ExitStatus = %d, want 3", rec.ExitStatus)
	}
	if rec.Duration != 250*time.Millisecond {
		t.Errorf("Duration = %v, want 250ms", rec.Duration)
	}
	if rec.StdoutDigest != digestOfValue1 {
		t.Errorf("StdoutDigest = %s, want %s", rec.StdoutDigest, digestOfValue1)
	}
	if rec.StderrDigest != digestOfWarning {
		t.Errorf("StderrDigest = %s, want %s", rec.StderrDigest, digestOfWarning)
	}
	if rec.StdoutDigest == rec.StderrDigest {
		t.Error("the two streams share one digest, so they were not digested separately")
	}
	if !rec.Failed() {
		t.Error("Failed() = false for a step that exited 3")
	}
}

func TestAnEmptyStreamStillCarriesADigest(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{Argv: []string{"bd", "ready"}})

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: []plan.Step{readStep("ready", "ready")}}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	rec, _ := report.Record("ready")
	if rec.StdoutDigest != digestOfEmpty {
		t.Errorf("StdoutDigest = %s, want the digest of no bytes (%s)", rec.StdoutDigest, digestOfEmpty)
	}
}

func TestAReferenceIsResolvedToTheValueTheEarlierStepProduced(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "show", "run-1"}, Stdout: []byte("  claude/branch-x  \n")},
		exectest.Recording{Argv: []string{"git", "merge", "claude/branch-x"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		readStep("read-branch", "show", "run-1"),
		{Name: "merge", Executable: "git", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("merge"), plan.From("read-branch", nil)}},
	}}

	if _, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute); err != nil {
		t.Fatalf("execute: %v", err)
	}

	argvs := fake.Argvs()
	if len(argvs) != 2 {
		t.Fatalf("%d invocations ran, want 2: %v", len(argvs), argvs)
	}
	if got, want := strings.Join(argvs[1], " "), "git merge claude/branch-x"; got != want {
		t.Errorf("the resolved invocation = %q, want %q", got, want)
	}
}

// A step referencing one that failed is not handed whatever the failing
// tool happened to print.
func TestAReferenceToAStepThatFailedIsNotResolved(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "show", "run-1"}, Stdout: []byte("garbage\n"), ExitStatus: 1},
	)

	p := plan.Plan{Steps: []plan.Step{
		{Name: "read-branch", Executable: "bd", Effect: plan.Read, OnFailure: plan.Continue,
			Args: []plan.Arg{plan.Lit("show"), plan.Lit("run-1")}},
		{Name: "merge", Executable: "git", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("merge"), plan.From("read-branch", nil)}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, argv := range fake.Argvs() {
		if argv[0] == "git" {
			t.Fatalf("the referencing step ran with an unresolved argument: %v", argv)
		}
	}
	rec, _ := report.Record("merge")
	if rec.Status != plan.StepSkipped {
		t.Errorf("the referencing step's status = %q, want %q", rec.Status, plan.StepSkipped)
	}
}

func TestValidateRejectsAPlanThatCannotBeExplained(t *testing.T) {
	cases := map[string]plan.Plan{
		"a step with no name": {Steps: []plan.Step{
			{Executable: "bd", Effect: plan.Read},
		}},
		"a step with no executable": {Steps: []plan.Step{
			{Name: "read", Effect: plan.Read},
		}},
		"a duplicate step name": {Steps: []plan.Step{
			readStep("read", "ready"), readStep("read", "ready"),
		}},
		"an argument from a later step": {Steps: []plan.Step{
			{Name: "first", Executable: "git", Effect: plan.Read, Args: []plan.Arg{plan.From("second", nil)}},
			readStep("second", "ready"),
		}},
		"an argument from a step that does not exist": {Steps: []plan.Step{
			{Name: "only", Executable: "git", Effect: plan.Read, Args: []plan.Arg{plan.From("nowhere", nil)}},
		}},
		"an alternative naming no condition": {Steps: []plan.Step{
			{Name: "merge", Executable: "git", Effect: plan.Write, Alternatives: []plan.Alternative{{
				Steps: []plan.Step{readStep("abort", "abort")},
			}}},
		}},
	}

	for name, p := range cases {
		if err := p.Validate(); err == nil {
			t.Errorf("%s: Validate() returned no error", name)
		}
		// A malformed plan is refused before anything is printed or run,
		// not partway through.
		if _, err := (plan.Runner{}).Run(context.Background(), p, plan.ModeExplain); err == nil {
			t.Errorf("%s: Run(explain) returned no error", name)
		}
	}
}

func TestAValidPlanValidates(t *testing.T) {
	p := plan.Plan{Steps: []plan.Step{
		readStep("first", "ready"),
		{Name: "second", Executable: "git", Effect: plan.Read, Args: []plan.Arg{plan.From("first", nil)}},
	}}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() on a well-formed plan: %v", err)
	}
}

// The transcript is meant to be run by hand, so an argument that looks
// like shell syntax has to survive being pasted as one word.
func TestAnArgumentThatLooksLikeShellSyntaxIsQuotedInTheTranscript(t *testing.T) {
	var out strings.Builder

	p := plan.Plan{Steps: []plan.Step{
		{Name: "comment", Executable: "bd", Effect: plan.Write, Args: []plan.Arg{
			plan.Lit("comment"), plan.Lit("run-1"), plan.Lit("retro: it's a; rm -rf / $(echo hi)"),
		}},
	}}

	if _, err := (plan.Runner{Out: &out}).Run(context.Background(), p, plan.ModeExplain); err != nil {
		t.Fatalf("explain: %v", err)
	}

	line := commandLines(out.String())[0]
	want := `bd comment run-1 'retro: it'\''s a; rm -rf / $(echo hi)'`
	if line != want {
		t.Errorf("transcript line =\n  %s\nwant\n  %s", line, want)
	}
}

// Explain touches no process runner at all, which is why it is safe
// against a repository nobody has looked at. The two modes that do run
// steps say so rather than silently doing nothing.
func TestExplainNeedsNoProcessRunnerAndTheOtherModesSaySoWhenTheyHaveNone(t *testing.T) {
	p := plan.Plan{Steps: []plan.Step{readStep("read", "ready")}}

	if _, err := (plan.Runner{}).Run(context.Background(), p, plan.ModeExplain); err != nil {
		t.Errorf("explain with no process runner: %v", err)
	}
	for _, mode := range []plan.Mode{plan.ModeDryRun, plan.ModeExecute} {
		if _, err := (plan.Runner{}).Run(context.Background(), p, mode); err != plan.ErrNoProcessRunner {
			t.Errorf("%s with no process runner: err = %v, want ErrNoProcessRunner", mode, err)
		}
	}
}

// A missing executable is a failure of the step, not a panic and not a
// silent success — it reaches the report through the same Err field a
// decoder failure does.
func TestAnInvocationThatNeverRanIsReportedAsAFailure(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv: []string{"bd", "ready"},
		Err:  &swarmexec.NotFoundError{Name: "bd"},
	})

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: []plan.Step{readStep("ready", "ready")}}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	rec, _ := report.Record("ready")
	if rec.Err == nil {
		t.Error("the record carries no error for an invocation that never ran")
	}
	if !rec.Failed() {
		t.Error("Failed() = false for a step whose executable was missing")
	}
}

// An arm is entered only when its condition selects it, and the outer
// plan carries on afterwards.
func TestASelectedAlternativeRunsAndThenTheOuterPlanContinues(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge", "dev-1"}, ExitStatus: 1},
		exectest.Recording{Argv: []string{"git", "merge", "--abort"}},
		exectest.Recording{Argv: []string{"git", "status"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		{Name: "merge", Executable: "git", Effect: plan.Write, OnFailure: plan.Continue,
			Args: []plan.Arg{plan.Lit("merge"), plan.Lit("dev-1")},
			Alternatives: []plan.Alternative{{
				Condition: "the merge did not apply cleanly",
				Select:    func(r plan.Record, _ any) bool { return r.ExitStatus != 0 },
				Steps: []plan.Step{
					{Name: "abort", Executable: "git", Effect: plan.Write,
						Args: []plan.Arg{plan.Lit("merge"), plan.Lit("--abort")}},
				},
			}}},
		{Name: "status", Executable: "git", Effect: plan.Read, Args: []plan.Arg{plan.Lit("status")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	var ran []string
	for _, argv := range fake.Argvs() {
		ran = append(ran, strings.Join(argv, " "))
	}
	want := []string{"git merge dev-1", "git merge --abort", "git status"}
	if strings.Join(ran, " | ") != strings.Join(want, " | ") {
		t.Errorf("ran %v, want %v", ran, want)
	}

	abort, ok := report.Record("abort")
	if !ok {
		t.Fatal("the arm's step is not in the report")
	}
	if abort.Branch != "the merge did not apply cleanly" {
		t.Errorf("the arm's record does not name its condition: %q", abort.Branch)
	}
}

// A decoder whose output nothing can reach is not a decoder, so the
// decoded value is on the record a command reads and not only in the
// interpreter's internal state.
func TestTheRecordCarriesWhatTheDecoderMadeOfTheResult(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "ready"}, Stdout: []byte(`{"count":3}`)},
		exectest.Recording{Argv: []string{"bd", "blocked"}, Stdout: []byte("boom"), ExitStatus: 1},
	)

	decode := func(res swarmexec.Result) (any, error) {
		var v struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(res.Stdout, &v); err != nil {
			return nil, err
		}
		return v.Count, nil
	}

	p := plan.Plan{Steps: []plan.Step{
		{Name: "ready", Executable: "bd", Effect: plan.Read, Decode: decode,
			Args: []plan.Arg{plan.Lit("ready")}},
		{Name: "blocked", Executable: "bd", Effect: plan.Read, OnFailure: plan.Continue, Decode: decode,
			Args: []plan.Arg{plan.Lit("blocked")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	ready, _ := report.Record("ready")
	if ready.Value != 3 {
		t.Errorf("the record's Value = %v, want 3", ready.Value)
	}

	// A failed step publishes nothing, for the same reason a later step
	// referencing it is not resolved.
	blocked, _ := report.Record("blocked")
	if blocked.Value != nil {
		t.Errorf("a failed step published %v, want nil", blocked.Value)
	}
}

// A tool whose exit status IS the answer declares the statuses it answers
// with, so a legitimate "no" does not trip the ordering guard. `git
// merge-base --is-ancestor` is the case: measured on git 2.43.0, exit 0
// for yes, exit 1 for no, exit 128 when a revision does not resolve.
func TestAStepMayDeclareTheExitStatusesItAnswersWith(t *testing.T) {
	decode := func(res swarmexec.Result) (any, error) { return res.ExitStatus == 0, nil }

	ancestry := func(name string, status int) plan.Step {
		return plan.Step{
			Name: name, Executable: "git", Effect: plan.Read, Repeatable: true,
			Answers: []int{1}, Decode: decode,
			Args: []plan.Arg{plan.Lit("merge-base"), plan.Lit("--is-ancestor"), plan.Lit(name)},
		}
	}

	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge-base", "--is-ancestor", "yes"}},
		exectest.Recording{Argv: []string{"git", "merge-base", "--is-ancestor", "no"}, ExitStatus: 1},
		exectest.Recording{Argv: []string{"git", "merge-base", "--is-ancestor", "bad"}, ExitStatus: 128,
			Stderr: []byte("fatal: Not a valid object name deadbeef\n")},
		exectest.Recording{Argv: []string{"git", "status"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		ancestry("yes", 0), ancestry("no", 1), ancestry("bad", 128),
		{Name: "after", Executable: "git", Effect: plan.Read, Args: []plan.Arg{plan.Lit("status")}},
	}}

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	yes, _ := report.Record("yes")
	if yes.Failed() || yes.Value != true {
		t.Errorf("exit 0: Failed=%v Value=%v, want false/true", yes.Failed(), yes.Value)
	}

	// The point of the field: exit 1 is an answer, so the step did not
	// fail and its value was published.
	no, _ := report.Record("no")
	if no.Failed() {
		t.Error("exit 1: the step is reported as failed though it declared 1 as an answer")
	}
	if no.Value != false {
		t.Errorf("exit 1: Value = %v, want false", no.Value)
	}

	// A status it did NOT declare is still a failure, and the plan stops.
	bad, _ := report.Record("bad")
	if !bad.Failed() {
		t.Error("exit 128: the step is not reported as failed though it declared only 1 as an answer")
	}
	if report.Stopped != "bad" {
		t.Errorf("Report.Stopped = %q, want %q", report.Stopped, "bad")
	}
	for _, argv := range fake.Argvs() {
		if strings.Join(argv, " ") == "git status" {
			t.Error("the step after an undeclared status ran")
		}
	}
}

func TestAnUnselectedAlternativeIsNotRun(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge", "dev-1"}},
		exectest.Recording{Argv: []string{"git", "merge", "--abort"}},
	)

	p := plan.Plan{Steps: []plan.Step{
		{Name: "merge", Executable: "git", Effect: plan.Write,
			Args: []plan.Arg{plan.Lit("merge"), plan.Lit("dev-1")},
			Alternatives: []plan.Alternative{{
				Condition: "the merge did not apply cleanly",
				Select:    func(r plan.Record, _ any) bool { return r.ExitStatus != 0 },
				Steps: []plan.Step{
					{Name: "abort", Executable: "git", Effect: plan.Write,
						Args: []plan.Arg{plan.Lit("merge"), plan.Lit("--abort")}},
				},
			}}},
	}}

	if _, err := (plan.Runner{Exec: fake}).Run(context.Background(), p, plan.ModeExecute); err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, argv := range fake.Argvs() {
		if strings.Join(argv, " ") == "git merge --abort" {
			t.Error("an arm whose condition did not hold was run")
		}
	}
}
