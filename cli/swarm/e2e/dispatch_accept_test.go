package e2e

import (
	"encoding/json"
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
func standDispatchBranch(t *testing.T, w world, branch string) {
	t.Helper()
	w.mustRun(t, w.repo, "git", "branch", branch, "origin/"+branch)
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

	stdout, stderr, code := w.swarm(t, "dispatch", "accept",
		"--run", "fixture-run", "--dispatch", "dev-fixture-2", "--json")
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

	branchAfter := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-2")
	if branchAfter == branchBefore {
		t.Errorf("bb/dev-fixture-2 is still at %s; a concurrent dispatch is rebased", branchBefore)
	}
	targetAfter := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	if targetAfter != branchAfter {
		t.Errorf("the target is at %s and the branch at %s; the fold is a fast-forward onto the branch",
			targetAfter, branchAfter)
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
