package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// standDispatchBranch creates the local dispatch branch from the
// remote-tracking ref the bundle carries.
//
// A clone of a bundle checks out HEAD and creates a local branch for that
// alone, so `bb/dev-fixture-2` arrives only as `origin/bb/dev-fixture-2`.
// In a real run `swarm dispatch open --dispatch` created the local
// branch; here the world carries the commits and this puts the ref where
// that command would have left it.
//
// `--no-track` is what makes it the same ref that command leaves.
// `dispatch open` cuts the branch at a COMMIT (`git worktree add -b
// <branch> <path> <sha>`), which sets no upstream; starting one from a
// remote-tracking ref sets one by default, and that difference is not
// cosmetic. Measured against this fixture: with the upstream set,
// `git branch -d bb/dev-fixture-2` after a rebase exits 1 — the rebased
// head is not in `origin/bb/dev-fixture-2` — where `git branch -d`
// checks HEAD when there is no upstream, and the fold has just put the
// branch there.
func standDispatchBranch(t *testing.T, w world, branch string) {
	t.Helper()
	w.mustRun(t, w.repo, "git", "branch", "--no-track", branch, "origin/"+branch)
}

// recordFinding puts a `retro:` comment on a bead, which is what the
// dispatch's own close ceremony leaves and what the post-merge
// acceptance re-check reads back.
//
// Written by the case rather than carried by the fixture because a
// tracker export does not carry comments: `bd import` restores issues,
// and a world declaring a finding would have to declare it somewhere
// `bd` never reads.
func recordFinding(t *testing.T, w world, bead string) {
	t.Helper()
	w.mustRun(t, w.repo, "bd", "comment", bead,
		"retro:proof the world was measured before it was accepted")
}

// AC: AC-SWARM-DISP-004.8
//
// A dispatch whose agent still holds a bead is not accepted: the command
// takes the attention status naming what the agent holds, and merges
// nothing — asserted from the target branch being where it was.
//
// The drain derivation is `swarm ready --agent`'s, reused rather than
// restated, so this case also proves the two agree: dev-fixture-1 holds
// fixture-run.2 in progress and nothing ready, which is exactly the half
// a ready query alone would miss.
func TestDispatchAcceptRefusesADispatchThatHasNotDrained(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")
	before := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")

	stdout, stderr, code := w.swarm(t, "dispatch", "accept",
		"--run", "fixture-run", "--dispatch", "dev-fixture-1", "--json")
	if code != 1 {
		t.Fatalf("dispatch accept exited %d, want 1 (attention)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if !strings.Contains(stdout, "fixture-run.2") {
		t.Errorf("the refusal does not name the bead the agent still holds:\n%s", stdout)
	}

	after := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	if after != before {
		t.Errorf("the target moved from %s to %s; an undrained dispatch must merge nothing", before, after)
	}
}

// AC: AC-SWARM-DISP-004.3
//
// A dispatch branch cut from an earlier target than the current one is
// rebased onto the target as it stands at that moment, and the resulting
// merge is a fast-forward — asserted against the real git, so the
// concurrent case is proved by the target and the branch ending at one
// commit rather than by what the command said it did.
func TestDispatchAcceptRebasesAConcurrentBranchAndFastForwards(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "dispatch-stale-base")
	standDispatchBranch(t, w, "bb/dev-fixture-2")

	targetBefore := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	branchBefore := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-2")

	// The world is the concurrent one: the merge-base is not the target's
	// head, which is what the command is about to read.
	base := w.mustCapture(t, w.repo, "git", "merge-base", "claude/bead-bonanza-fixture", "bb/dev-fixture-2")
	if base == targetBefore {
		t.Fatalf("the fixture's merge-base (%s) IS the target's head; this case needs them to differ", base)
	}

	// What the dispatch's own close ceremony would have left behind, and
	// what the acceptance re-check reads: the finding recorded on the
	// bead it closed. The fixture carries the commits and the closed
	// bead; a comment is not something a tracker export puts in a world.
	recordFinding(t, w, "fixture-run.3")

	stdout, stderr, code := w.swarm(t, "dispatch", "accept",
		"--run", "fixture-run", "--dispatch", "dev-fixture-2", "--no-push", "--json")
	if code != 0 {
		t.Fatalf("dispatch accept exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		Case   string `json:"case"`
		Merged bool   `json:"merged"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	if answer.Case != "concurrent" {
		t.Errorf("case = %q, want %q", answer.Case, "concurrent")
	}
	if !answer.Merged {
		t.Errorf("merged = false after a clear exit:\n%s", stdout)
	}

	// The branch is gone by now — the acceptance deletes it once the fold
	// has landed — so the rebase is proved from the TARGET: it carries
	// the dispatch's commit, at a sha the rebase minted rather than the
	// one the branch arrived with.
	targetAfter := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	if targetAfter == targetBefore {
		t.Errorf("the target is still at %s; the dispatch was not folded in", targetBefore)
	}
	if targetAfter == branchBefore {
		t.Errorf("the target is at the branch's original sha %s; a concurrent dispatch is rebased first", branchBefore)
	}
	if subject := w.mustCapture(t, w.repo, "git", "log", "-1", "--format=%s", targetAfter); subject != "fixture-run.3 P1 the third bead of the run" {
		t.Errorf("the target's head is %q; the dispatch's own commit is what landed", subject)
	}

	// A fast-forward and not a merge commit: the new head has one parent.
	parents := w.mustCapture(t, w.repo, "git", "rev-list", "--parents", "-n", "1", targetAfter)
	if fields := strings.Fields(parents); len(fields) != 2 {
		t.Errorf("the target's head has %d parent(s) (%q); a fast-forward leaves one",
			len(fields)-1, parents)
	}

	// And the repository is back on the target, not left on the branch
	// the rebase checked out.
	if head := w.mustCapture(t, w.repo, "git", "rev-parse", "--abbrev-ref", "HEAD"); head != "claude/bead-bonanza-fixture" {
		t.Errorf("the repository is on %q, want the target; the rebase left it on the dispatch branch", head)
	}
}

// AC: AC-SWARM-DISP-004.4, .7
//
// A rebase that conflicts records a sequencing finding naming both beads
// and the conflicting files, marks the unmerged bead blocked, and
// resolves nothing — the evidence is left exactly where the rebase
// stopped.
func TestDispatchAcceptBlocksTheBeadAndRecordsAFindingOnAConflict(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "dispatch-rebase-conflict")
	standDispatchBranch(t, w, "bb/dev-fixture-2")

	targetBefore := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")

	stdout, stderr, code := w.swarm(t, "dispatch", "accept",
		"--run", "fixture-run", "--dispatch", "dev-fixture-2", "--json")
	if code != 2 {
		t.Fatalf("dispatch accept exited %d, want 2 (stop)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if !strings.Contains(stdout, "work.txt") {
		t.Errorf("the report does not name the conflicting path:\n%s", stdout)
	}

	// The unmerged bead is blocked.
	shown := w.mustCapture(t, w.repo, "bd", "show", "fixture-run.3", "--json")
	var beads []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(shown), &beads); err != nil || len(beads) == 0 {
		t.Fatalf("bd show fixture-run.3 did not answer a bead: %v\n%s", err, shown)
	}
	if beads[0].Status != "blocked" {
		t.Errorf("fixture-run.3 is %q, want blocked", beads[0].Status)
	}

	// And the sequencing finding is on it, naming both beads and the path.
	comments := w.mustCapture(t, w.repo, "bd", "comments", "fixture-run.3", "--json")
	if !strings.Contains(comments, "retro:sequencing") {
		t.Errorf("no sequencing finding was recorded on fixture-run.3:\n%s", comments)
	}
	for _, want := range []string{"fixture-run.3", "fixture-run.4", "work.txt"} {
		if !strings.Contains(comments, want) {
			t.Errorf("the finding does not name %q:\n%s", want, comments)
		}
	}

	// Nothing was resolved and nothing landed: the target is where it was.
	targetAfter := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	if targetAfter != targetBefore {
		t.Errorf("the target moved from %s to %s on a conflicting rebase", targetBefore, targetAfter)
	}
	unmerged := w.mustCapture(t, w.repo, "git", "diff", "--name-only", "--diff-filter=U")
	if !strings.Contains(unmerged, "work.txt") {
		t.Errorf("the conflict was not left standing for an agent to resolve; unmerged paths: %q", unmerged)
	}
}

// AC: AC-SWARM-DISP-004.11
//
// Every bead in the dispatch is re-checked after the merge and the
// removal, with no signal deferred — and the two a merge decides are the
// reason the check runs THERE. Proved against a world with a worktree
// really standing, so "the worktree is gone" is read from git's own
// registration and from the directory rather than from the removal
// having exited zero.
//
// The dispatch's bead is closed and its finding recorded here, which is
// what its own close ceremony would have left; the fixture carries the
// commit, the branch and the standing worktree, which a case cannot
// build.
func TestDispatchAcceptRechecksEveryBeadAfterTheMergeAndTheRemoval(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "dispatch-worktree")
	worktree := filepath.Join(w.dir, ".bb-worktrees", "dev-fixture-2")
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("the fixture's worktree is not standing at %s: %v", worktree, err)
	}

	w.mustRun(t, w.repo, "bd", "update", "fixture-run.3", "--status", "closed")
	recordFinding(t, w, "fixture-run.3")

	stdout, stderr, code := w.swarm(t, "dispatch", "accept",
		"--run", "fixture-run", "--dispatch", "dev-fixture-2", "--no-push", "--json")
	if code != 0 {
		t.Fatalf("dispatch accept exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		Merged             bool     `json:"merged"`
		Ancestor           bool     `json:"ancestor"`
		WorktreeRemoved    bool     `json:"worktree_removed"`
		BranchDeleted      bool     `json:"branch_deleted"`
		BeadsChecked       int      `json:"beads_checked"`
		AcceptanceFailures []string `json:"acceptance_failures"`
		Landed             bool     `json:"landed"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	switch {
	case !answer.Merged:
		t.Errorf("merged = false after a clear exit:\n%s", stdout)
	case !answer.Ancestor:
		t.Errorf("ancestor = false after a clear exit:\n%s", stdout)
	case !answer.WorktreeRemoved:
		t.Errorf("worktree_removed = false, but one was standing:\n%s", stdout)
	case !answer.BranchDeleted:
		t.Errorf("branch_deleted = false after a clear exit:\n%s", stdout)
	}
	if answer.BeadsChecked != 1 {
		t.Errorf("beads_checked = %d, want 1 (fixture-run.3 is dev-fixture-2's only bead):\n%s",
			answer.BeadsChecked, stdout)
	}
	if len(answer.AcceptanceFailures) != 0 {
		t.Errorf("acceptance_failures = %v, want none:\n%s", answer.AcceptanceFailures, stdout)
	}
	if answer.Landed {
		t.Errorf("landed = true under --no-push:\n%s", stdout)
	}

	// The signals the re-check read, re-read here from the world itself.
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Errorf("the worktree at %s is still on disk (%v)", worktree, err)
	}
	list := w.mustCapture(t, w.repo, "git", "worktree", "list", "--porcelain")
	if strings.Contains(list, "refs/heads/bb/dev-fixture-2") {
		t.Errorf("a worktree is still registered on the dispatch branch:\n%s", list)
	}
	if _, _, code := w.run(t, w.repo, "git", "rev-parse", "--verify", "bb/dev-fixture-2"); code == 0 {
		t.Errorf("bb/dev-fixture-2 still exists after a clear acceptance")
	}
	head := w.mustCapture(t, w.repo, "git", "log", "-1", "--format=%s", "claude/bead-bonanza-fixture")
	if head != "fixture-run.3 P1 the third bead of the run" {
		t.Errorf("the target's head is %q; the dispatch's own commit is what landed", head)
	}
}

// `merged_sha` names a commit that is ON the target, on both arms.
//
// The concurrent arm is where this can go wrong and did: the rebase
// rewrites the dispatch's commits, so the branch sha read before the fold
// names a commit that is not reachable from the target afterwards, and a
// coordinator verifying the merge by that sha finds nothing. The
// serialized arm is checked beside it because the two must not disagree
// about what the fact means.
//
// The property asserted is ancestry rather than equality with the
// target's head, because it is the property a caller actually uses — the
// commit is on the target — and it holds whatever else lands afterwards.
func TestDispatchAcceptReportsAMergedShaThatIsOnTheTarget(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	const target = "claude/bead-bonanza-fixture"

	for _, c := range []struct {
		name    string
		fixture string
		arrange func(t *testing.T, w world)
	}{
		{
			name:    "concurrent",
			fixture: "dispatch-stale-base",
			arrange: func(t *testing.T, w world) {
				standDispatchBranch(t, w, "bb/dev-fixture-2")
				recordFinding(t, w, "fixture-run.3")
			},
		},
		{
			name:    "serialized",
			fixture: "dispatch-worktree",
			arrange: func(t *testing.T, w world) {
				w.mustRun(t, w.repo, "bd", "update", "fixture-run.3", "--status", "closed")
				recordFinding(t, w, "fixture-run.3")
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := restore(t, binary, c.fixture)
			c.arrange(t, w)

			stdout, stderr, code := w.swarm(t, "dispatch", "accept",
				"--run", "fixture-run", "--dispatch", "dev-fixture-2", "--no-push", "--json")
			if code != 0 {
				t.Fatalf("dispatch accept exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s",
					code, stdout, stderr)
			}

			var answer struct {
				Case      string `json:"case"`
				Merged    bool   `json:"merged"`
				MergedSHA string `json:"merged_sha"`
			}
			if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
				t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
			}
			if answer.Case != c.name {
				t.Fatalf("case = %q, want %q — this fixture is not the arm it is here to cover",
					answer.Case, c.name)
			}
			if !answer.Merged {
				t.Fatalf("merged = false after a clear exit:\n%s", stdout)
			}
			if answer.MergedSHA == "" {
				t.Fatalf("no merged_sha was reported:\n%s", stdout)
			}

			if _, _, code := w.run(t, w.repo,
				"git", "merge-base", "--is-ancestor", answer.MergedSHA, target); code != 0 {
				targetHead := w.mustCapture(t, w.repo, "git", "rev-parse", target)
				t.Errorf("merged_sha %s is not on %s (which is at %s); "+
					"the fact names a commit that did not land",
					answer.MergedSHA, target, targetHead)
			}
		})
	}
}
