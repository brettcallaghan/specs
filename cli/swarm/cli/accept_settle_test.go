// Unit tests for `swarm dispatch accept`'s second half: the ancestry
// guard the removal sits behind, the removal-then-deletion ordering, a
// refusal that is neither retried nor forced, the post-merge re-check
// gating the push, and a dry run that writes nothing.
//
// Every case drives the real command tree — the Beads Tracker's and the
// Git Version Control's own argv-building — against the fake process
// runner, and asserts what the command DID from the fake's record rather
// than what its report claimed. The two are different, and only the first
// would have caught a removal chained after a failed merge.
//
// AC-SWARM-DISP-004.1 and .11 are declared e2e and live in cli/swarm/e2e
// against the dispatch-stale-base fixture.
package cli

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
)

// ran reports whether the fake was asked for an invocation beginning with
// prefix. It is the assertion every ordering case below makes.
func ran(fake *exectest.Fake, prefix ...string) bool {
	for _, argv := range fake.Argvs() {
		if len(argv) < len(prefix) {
			continue
		}
		match := true
		for i, want := range prefix {
			if argv[i] != want {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// indexOf returns the call index of the first invocation beginning with
// prefix, -1 when there is none.
func indexOf(fake *exectest.Fake, prefix ...string) int {
	for i, argv := range fake.Argvs() {
		if len(argv) < len(prefix) {
			continue
		}
		match := true
		for j, want := range prefix {
			if argv[j] != want {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// AC: AC-SWARM-DISP-004.1
//
// A worktree is not removed unless its branch is an ancestor of the
// target: the fold reports success and the ancestry read answers no, so
// the worktree is left in place, no branch is deleted, and the command
// takes the stop status.
//
// Declared e2e in the blueprint's coverage table and proved HERE instead,
// with the reason recorded as a finding: the world the criterion
// describes cannot be built against a real git. A fast-forward that
// succeeds has by definition put the branch on the target, so nothing a
// single-actor end-to-end world can do produces a successful fold whose
// branch is not an ancestor — that state exists only when another actor
// moves the target between the two reads, which is the race the guard is
// there for. The fake is the only place the two answers can be posed
// independently.
func TestAWorktreeIsNotRemovedUnlessItsBranchIsAnAncestorOfTheTarget(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})

	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"
	world.ancestorStatus = 1 // the target moved: the fold's branch is not on it
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (stop)\n%s", code, StatusStop, stdout)
	}
	if ran(fake, "git", "worktree", "remove") {
		t.Errorf("a worktree was removed with the branch not an ancestor of the target: %v", fake.Argvs())
	}
	if ran(fake, "git", "branch", "-d") {
		t.Errorf("a branch was deleted with the branch not an ancestor of the target: %v", fake.Argvs())
	}
	if ran(fake, "git", "push") {
		t.Errorf("a push was attempted with the branch not an ancestor of the target: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "not an ancestor") {
		t.Errorf("the report does not say why nothing was removed:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-004.6
//
// A step does not run when the step before it failed: the merge exits
// non-zero, and no worktree is removed, no branch is deleted and no push
// is attempted.
//
// The guarantee is the relied-upon Plan Runner's — a step whose failure
// policy is Stop ends its plan, and the removal is not even in the fold's
// plan; it is in one that is only BUILT when the fold returned no error.
// This asserts the property holds through this command rather than
// reimplementing it.
func TestAFailedMergeRemovesDeletesAndPushesNothing(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge", "--ff-only", acceptBranch},
		Stderr:     []byte("fatal: Not possible to fast-forward, aborting.\n"),
		ExitStatus: 128,
	})
	// Declared and expected to go unused: recordings the command must
	// not reach.
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (stop)\n%s", code, StatusStop, stdout)
	}
	for _, forbidden := range [][]string{
		{"git", "worktree", "remove"},
		{"git", "branch", "-d"},
		{"git", "push"},
	} {
		if ran(fake, forbidden...) {
			t.Errorf("%v ran after a merge that exited non-zero: %v", forbidden, fake.Argvs())
		}
	}
}

// AC: AC-SWARM-DISP-004.9
//
// The dispatch branch is not deleted before the fast-forward has
// succeeded. The rebase completes, the fold that follows it is refused,
// and the deletion never runs — so a run interrupted at that point still
// has the branch.
func TestTheDispatchBranchIsNotDeletedBeforeTheFastForwardSucceeds(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(1)...) // concurrent
	recs = append(recs,
		exectest.Recording{Argv: []string{"git", "rebase", acceptTargetSH, acceptBranch}},
		exectest.Recording{Argv: []string{"git", "checkout", acceptTargetBr}},
		exectest.Recording{
			Argv:       []string{"git", "merge", "--ff-only", acceptBranch},
			Stderr:     []byte("fatal: Not possible to fast-forward, aborting.\n"),
			ExitStatus: 128,
		},
	)
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (stop)\n%s", code, StatusStop, stdout)
	}
	if !ran(fake, "git", "rebase") {
		t.Fatalf("the rebase never ran, so this is not the world the case describes: %v", fake.Argvs())
	}
	if ran(fake, "git", "branch", "-d") {
		t.Errorf("%s was deleted with the fast-forward refused: %v", acceptBranch, fake.Argvs())
	}
	if ran(fake, "git", "worktree", "remove") {
		t.Errorf("a worktree was removed with the fast-forward refused: %v", fake.Argvs())
	}
}

// The removal comes before the deletion, and both come after the fold —
// the delta's second behavior line, asserted as the order the fake was
// asked for them in rather than as a claim in the report.
func TestTheWorktreeIsRemovedBeforeItsBranchIsDeleted(t *testing.T) {
	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--no-push", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	foldIdx := indexOf(fake, "git", "merge", "--ff-only")
	removeIdx := indexOf(fake, "git", "worktree", "remove")
	deleteIdx := indexOf(fake, "git", "branch", "-d")
	if foldIdx < 0 || removeIdx < 0 || deleteIdx < 0 {
		t.Fatalf("fold@%d remove@%d delete@%d; all three are required: %v",
			foldIdx, removeIdx, deleteIdx, fake.Argvs())
	}
	if !(foldIdx < removeIdx && removeIdx < deleteIdx) {
		t.Errorf("order was fold@%d, remove@%d, delete@%d; the removal follows the fold and the deletion follows the removal",
			foldIdx, removeIdx, deleteIdx)
	}
}

// AC: AC-SWARM-DISP-004.10
//
// A worktree whose removal the tool refuses is never force-removed: the
// removal is not retried at all, the command takes the attention status
// naming what the tool refused over, and the branch is left undeleted.
func TestARefusedRemovalIsNeitherRetriedNorForced(t *testing.T) {
	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"
	world.removalStatus = 128

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}

	var removals [][]string
	for _, argv := range fake.Argvs() {
		if len(argv) >= 3 && argv[0] == "git" && argv[1] == "worktree" && argv[2] == "remove" {
			removals = append(removals, argv)
		}
	}
	if len(removals) != 1 {
		t.Errorf("the removal was attempted %d time(s), want exactly once: %v", len(removals), removals)
	}
	for _, argv := range removals {
		if contains(argv, "--force") || contains(argv, "-f") {
			t.Errorf("the removal was forced: %v", argv)
		}
	}
	if ran(fake, "git", "branch", "-d") {
		t.Errorf("the branch was deleted after a refused removal: %v", fake.Argvs())
	}
	if ran(fake, "git", "push") {
		t.Errorf("a push was attempted after a refused removal: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, world.worktree) {
		t.Errorf("the report does not name what the tool refused over:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-004.10
//
// WHERE the refusal is over a LOCK, the resolution the command names is
// the unlock invocation and never the forcing form. Asserted on the text
// a caller acts on, because that is the only place a forcing form could
// reach an operator from: the operation itself has no such parameter.
func TestALockedWorktreeIsResolvedByUnlockingItAndNeverByForcing(t *testing.T) {
	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"
	world.locked = "dev-1 is still working in it"
	world.removalStatus = 128

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json", "--full-result")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	if want := "worktree unlock " + world.worktree; !strings.Contains(stdout, want) {
		t.Errorf("the resolution does not name %q:\n%s", want, stdout)
	}
	if strings.Contains(stdout, "--force") {
		t.Errorf("the report offers the forcing form:\n%s", stdout)
	}
	if !strings.Contains(stdout, world.locked) {
		t.Errorf("the report does not name the lock's reason:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-004.12
//
// The push runs only after every bead has passed its post-merge check. A
// bead that recorded no finding fails one signal, and that is enough: no
// push is attempted and the failing signal is named.
func TestNoPushIsAttemptedWhenABeadFailsItsPostMergeCheck(t *testing.T) {
	world := cleanSettle()
	world.findings = nil // run-1.1 recorded none

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	if ran(fake, "git", "push") {
		t.Errorf("a push was attempted with a bead failing its post-merge check: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "run-1.1: findings") {
		t.Errorf("the failing signal is not named:\n%s", stdout)
	}
}

// A worktree that survives its own removal fails the post-merge check
// too, which is what makes that signal a reading of state rather than of
// the removal's exit status.
func TestASurvivingWorktreeFailsThePostMergeCheck(t *testing.T) {
	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"
	world.survives = true

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	if !strings.Contains(stdout, "run-1.1: worktree") {
		t.Errorf("a worktree that outlived its removal did not fail the signal:\n%s", stdout)
	}
	if ran(fake, "git", "push") {
		t.Errorf("a push was attempted with the worktree still standing: %v", fake.Argvs())
	}
}

// WHERE --no-push is passed, every step before the push is taken and the
// branch is reported as unlanded rather than pushed.
func TestNoPushTakesEveryStepBeforeThePushAndReportsTheBranchUnlanded(t *testing.T) {
	world := cleanSettle()
	world.worktree = "/dispatches/dev-1"

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--no-push", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	for _, required := range [][]string{
		{"git", "merge", "--ff-only"},
		{"git", "worktree", "remove"},
		{"git", "branch", "-d"},
		{"bd", "comments"},
	} {
		if !ran(fake, required...) {
			t.Errorf("%v did not run, so --no-push skipped a step BEFORE the push: %v", required, fake.Argvs())
		}
	}
	if ran(fake, "git", "push") || ran(fake, "git", "fetch") {
		t.Errorf("--no-push pushed or fetched: %v", fake.Argvs())
	}
	if answer := acceptAnswer(t, stdout); answer["landed"] != false {
		t.Errorf("landed = %v, want false; --no-push leaves the branch unlanded:\n%s", answer["landed"], stdout)
	}
}

// AC: AC-SWARM-DISP-004.13
//
// A dry run performs every read it reaches and writes nothing: no merge,
// no removal, no deletion and no push. The ancestry read still runs — it
// is a read — and the steps behind a fold that did not happen are
// reported as not attempted rather than as having found nothing.
func TestAcceptDryRunReadsEverythingAndWritesNothing(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge-base", "--is-ancestor", acceptBranch, acceptTargetBr},
		ExitStatus: 1,
	})

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--dry-run", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	for _, forbidden := range [][]string{
		{"git", "merge", "--ff-only"},
		{"git", "rebase"},
		{"git", "worktree", "prune"},
		{"git", "worktree", "remove"},
		{"git", "branch", "-d"},
		{"git", "push"},
		{"git", "fetch"},
	} {
		if ran(fake, forbidden...) {
			t.Errorf("a dry run ran %v: %v", forbidden, fake.Argvs())
		}
	}
	if !ran(fake, "git", "merge-base", "--is-ancestor", acceptBranch, acceptTargetBr) {
		t.Errorf("a dry run skipped the ancestry read, which is a read: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "nothing was merged") {
		t.Errorf("the report does not say the writes were not attempted:\n%s", stdout)
	}
}
