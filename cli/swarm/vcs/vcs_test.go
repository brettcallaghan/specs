// Tests for the fragment type and for the one behaviour that only shows
// up when a fragment is actually interpreted: a dry run refuses the
// prune, and the standing-worktree answer has to still be right.
package vcs_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

// worktreeList holds one live registration and one stale one, in the form
// measured from `git worktree list --porcelain` in a scratch repository
// whose second worktree directory had been deleted.
const worktreeList = `worktree /w/main
HEAD f927a8fe0f263541ce8e7e8e594e9c5839166c6d
branch refs/heads/main

worktree /w/dev-2
HEAD f927a8fe0f263541ce8e7e8e594e9c5839166c6d
branch refs/heads/dev-2
prunable gitdir file points to non-existent location
`

func worktreeFake() *exectest.Fake {
	return exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "worktree", "prune"}},
		exectest.Recording{Argv: []string{"git", "worktree", "list", "--porcelain"}, Stdout: []byte(worktreeList)},
	)
}

func TestAFragmentsAnswerIsReadOutOfTheReport(t *testing.T) {
	steps := git.VersionControl{}.Worktrees("wt")

	report, err := (plan.Runner{Exec: worktreeFake()}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	value, err := steps.ValueFrom(report)
	if err != nil {
		t.Fatalf("ValueFrom: %v", err)
	}
	trees, ok := value.([]vcs.Worktree)
	if !ok {
		t.Fatalf("the answer is %T, not a worktree set", value)
	}
	if len(trees) != 1 || trees[0].Branch != "main" {
		t.Errorf("got %+v, want only the live registration", trees)
	}
}

// The prune is a write, so a dry run refuses it — and the standing-
// worktree answer still has to be right, because a stale registration
// counted as live is what the operation exists to prevent. This is why
// the decoder filters prunable entries as well as the prune removing
// them.
func TestADryRunRefusesThePruneAndStillDropsTheStaleRegistration(t *testing.T) {
	steps := git.VersionControl{}.Worktrees("wt")
	fake := worktreeFake()

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeDryRun)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}

	for _, argv := range fake.Argvs() {
		if strings.Join(argv, " ") == "git worktree prune" {
			t.Fatal("a dry run performed the prune")
		}
	}
	prune, _ := report.Record("wt-prune")
	if prune.Status != plan.StepRefused {
		t.Errorf("the prune's status = %q, want %q", prune.Status, plan.StepRefused)
	}

	value, err := steps.ValueFrom(report)
	if err != nil {
		t.Fatalf("ValueFrom: %v", err)
	}
	trees := value.([]vcs.Worktree)
	if len(trees) != 1 {
		t.Errorf("a dry run reported %d standing worktrees, want 1; the stale one was counted as live", len(trees))
	}
}

// A fragment whose answering step never ran says so, rather than handing
// back a zero value that reads as an answer.
func TestAFragmentWhoseAnswerStepDidNotRunSaysSo(t *testing.T) {
	steps := git.VersionControl{}.Merge("m", "dev-1", "merge dev-1")

	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge", "--no-ff", "-m", "merge dev-1", "dev-1"},
			ExitStatus: 128, Stderr: []byte("CONFLICT (content): Merge conflict in f.txt\n")},
		// Declared so a runner ignoring the guard would find something to
		// run rather than erroring for an unrelated reason.
		exectest.Recording{Argv: []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", "HEAD"}},
	)

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if report.Stopped != "m-merge" {
		t.Errorf("Report.Stopped = %q, want the merge", report.Stopped)
	}
	for _, argv := range fake.Argvs() {
		if argv[1] == "log" {
			t.Error("the step after a failed merge ran")
		}
	}
	if _, err := steps.ValueFrom(report); err == nil {
		t.Error("ValueFrom returned an answer for a fragment whose last step never ran")
	}
}

// A refusal has to survive the trip through the Plan Runner: the decoder
// builds it, the record carries it, and the caller branches on it. An
// error that arrived as a *Refusal and left as prose would put the
// caller back to matching on git's text.
func TestARefusalReachesTheCallerThroughTheReport(t *testing.T) {
	steps := git.VersionControl{}.MergeFastForward("ff", "dev-1")

	fake := exectest.NewFake(exectest.Recording{
		Argv:       []string{"git", "merge", "--ff-only", "dev-1"},
		ExitStatus: 128,
		Stderr:     []byte("fatal: Not possible to fast-forward, aborting.\n"),
	})

	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: steps}, plan.ModeExecute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	record, ok := report.Record("ff")
	if !ok {
		t.Fatal("the report holds no record for the fast-forward")
	}
	var refused *vcs.Refusal
	if !errors.As(record.Err, &refused) {
		t.Fatalf("the record's error is %T, not a refusal: %v", record.Err, record.Err)
	}
	if !refused.NonFastForward {
		t.Error("the caller cannot tell a non-fast-forward from a conflict")
	}
}

// Every write the dispatch loop makes is declared as one, so a dry run
// performs none of them. A write mis-declared as a read is a worktree a
// preview removes.
func TestADryRunPerformsNoneOfTheDispatchLoopsWrites(t *testing.T) {
	v := git.VersionControl{}
	writes := map[string]vcs.Steps{
		"create-worktree": v.CreateWorktree("create-worktree", "/w/dev-1", "dev-1", "sprint"),
		"remove-worktree": v.RemoveWorktree("remove-worktree", "/w/dev-1"),
		"create-branch":   v.CreateBranch("create-branch", "dev-1", "sprint"),
		"delete-branch":   v.DeleteBranch("delete-branch", "dev-1"),
		"rebase":          v.Rebase("rebase", "dev-1", "sprint"),
		"fast-forward":    v.MergeFastForward("fast-forward", "dev-1"),
		"fetch":           v.Fetch("fetch", "origin"),
		"push":            v.Push("push", "origin", "sprint"),
	}

	var all []plan.Step
	for _, steps := range writes {
		all = append(all, steps...)
	}

	// An empty fake: any invocation at all is a dry run that wrote.
	fake := exectest.NewFake()
	report, err := (plan.Runner{Exec: fake}).Run(context.Background(),
		plan.Plan{Steps: all}, plan.ModeDryRun)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}

	if got := fake.Argvs(); len(got) != 0 {
		t.Errorf("a dry run ran %v", got)
	}
	for name := range writes {
		record, ok := report.Record(name)
		if !ok {
			t.Errorf("the report holds no record for %q", name)
			continue
		}
		if record.Status != plan.StepRefused {
			t.Errorf("%q was %q in a dry run, want %q; it is not declared a write",
				name, record.Status, plan.StepRefused)
		}
	}
}

func TestValueFromReportsAMissingStepRatherThanAZeroValue(t *testing.T) {
	steps := git.VersionControl{}.Branches("br")

	if _, err := steps.ValueFrom(plan.Report{}); err == nil {
		t.Error("ValueFrom returned an answer from a report holding no record")
	}
	if _, err := (vcs.Steps{}).ValueFrom(plan.Report{}); err == nil {
		t.Error("an empty fragment returned an answer")
	}
}
