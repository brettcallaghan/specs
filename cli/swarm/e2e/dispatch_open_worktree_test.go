package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// AC: AC-SWARM-DISP-003.4, .10
//
// Opening a fresh dispatch of a run whose sprint branch is already
// recorded stands a worktree up on a new branch named for the dispatch
// (.4), cut from the default base — the sprint branch's own head, since
// --base is not given (.10) — against a real `git worktree add`, not a
// hand-written one.
func TestDispatchOpenStandsAWorktreeUpForAFreshDispatch(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	headSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "HEAD")

	stdout, stderr, code := w.swarm(t, "dispatch", "open", "--run", "fixture-run", "--dispatch", "dev-fixture-9", "--json")
	if code != 0 {
		t.Fatalf("dispatch open exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		WorktreePath    string `json:"worktree_path"`
		DispatchBranch  string `json:"dispatch_branch"`
		WorktreeBaseSHA string `json:"worktree_base_sha"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}

	// AC-SWARM-DISP-003.4: exactly one branch, named for the dispatch.
	if answer.DispatchBranch != "bb/dev-fixture-9" {
		t.Errorf("dispatch_branch = %q, want %q", answer.DispatchBranch, "bb/dev-fixture-9")
	}
	// AC-SWARM-DISP-003.7: a sibling of the repository, never inside it.
	want := filepath.Join(filepath.Dir(w.repo), ".bb-worktrees", "dev-fixture-9")
	if !samePath(answer.WorktreePath, want) {
		t.Errorf("worktree_path = %q, want %q", answer.WorktreePath, want)
	}
	// AC-SWARM-DISP-003.1, .10: cut from the sprint branch's own head,
	// which --base was not given to override.
	if answer.WorktreeBaseSHA != headSHA {
		t.Errorf("worktree_base_sha = %q, want %q (the sprint branch's head)", answer.WorktreeBaseSHA, headSHA)
	}

	// The worktree and its branch are real.
	if _, err := os.Stat(answer.WorktreePath); err != nil {
		t.Errorf("no worktree stands at %s: %v", answer.WorktreePath, err)
	}
	createdSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-9")
	if createdSHA != headSHA {
		t.Errorf("git rev-parse bb/dev-fixture-9 = %q, want %q", createdSHA, headSHA)
	}

	registered := w.mustCapture(t, w.repo, "git", "worktree", "list", "--porcelain")
	if !worktreeListed(registered, answer.WorktreePath) {
		t.Errorf("git worktree list does not carry %s:\n%s", answer.WorktreePath, registered)
	}
}

// AC: AC-SWARM-DISP-003.1
//
// A worktree is created from the base the caller names, distinct from
// the sprint branch's own head: `origin/main` and the sprint branch
// diverge in the core fixture, so a worktree cut from
// `--base origin/main` landing on that sha rather than the sprint
// branch's is what tells the two apart.
func TestDispatchOpenCutsTheWorktreeFromAnExplicitBase(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	mainSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "origin/main")
	sprintSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "claude/bead-bonanza-fixture")
	if mainSHA == sprintSHA {
		t.Fatalf("the core fixture's origin/main and sprint branch heads coincide (%s); this case needs them to differ", mainSHA)
	}

	stdout, stderr, code := w.swarm(t, "dispatch", "open", "--run", "fixture-run",
		"--dispatch", "dev-fixture-8", "--base", "origin/main", "--json")
	if code != 0 {
		t.Fatalf("dispatch open exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		WorktreeBaseSHA string `json:"worktree_base_sha"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	if answer.WorktreeBaseSHA != mainSHA {
		t.Errorf("worktree_base_sha = %q, want %q (main, the named --base)", answer.WorktreeBaseSHA, mainSHA)
	}

	createdSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-8")
	if createdSHA != mainSHA {
		t.Errorf("git rev-parse bb/dev-fixture-8 = %q, want %q (main, not the sprint branch's head)", createdSHA, mainSHA)
	}
}

// AC: AC-SWARM-DISP-003.11
//
// A dispatch whose worktree already stands is reported and left
// untouched: the fixture's own worktree, on its own commit, is
// unmoved — recreating it would fail outright, since the branch it
// carries already exists.
func TestDispatchOpenLeavesAnExistingWorktreeUntouched(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "dispatch-worktree")

	standingPath := filepath.Join(w.dir, filepath.FromSlash(".bb-worktrees/dev-fixture-2"))
	beforeSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-2")

	stdout, stderr, code := w.swarm(t, "dispatch", "open", "--run", "fixture-run", "--dispatch", "dev-fixture-2", "--json")
	if code != 0 {
		t.Fatalf("dispatch open exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		WorktreePath   string `json:"worktree_path"`
		DispatchBranch string `json:"dispatch_branch"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	if !samePath(answer.WorktreePath, standingPath) {
		t.Errorf("worktree_path = %q, want the standing worktree's own path %q", answer.WorktreePath, standingPath)
	}
	if answer.DispatchBranch != "bb/dev-fixture-2" {
		t.Errorf("dispatch_branch = %q, want %q", answer.DispatchBranch, "bb/dev-fixture-2")
	}

	afterSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "bb/dev-fixture-2")
	if afterSHA != beforeSHA {
		t.Errorf("bb/dev-fixture-2 moved from %s to %s; the standing worktree was not left untouched", beforeSHA, afterSHA)
	}
}
