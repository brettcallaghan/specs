package e2e

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// A worktree is attributed to a dispatch by the bead its HEAD commit
// closed and that bead's assignee, never by its branch name — against a
// REAL `git worktree list --porcelain` capture, not a hand-written one.
// The harness that actually creates a dispatch's worktree in this
// environment does not name it `bb/<dispatch-id>` (see
// cli/swarm/status/status.go's package doc), so a worktree checked out
// under an arbitrary name still has to be recognised.
func TestAWorktreeIsAttributedByItsCommitNotItsBranchName(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	// dev-fixture-1's own commit — "fixture-run.2 P1 the second bead of
	// the run" — lives on origin/dev-fixture-1 in the core fixture's
	// history. Checking it out under a branch that carries no dispatch id
	// at all is what this asserts survives.
	extra := filepath.Join(w.dir, "extra-worktree")
	w.mustRun(t, w.repo, "git", "worktree", "add",
		"-b", "worktree-agent-notadispatchname", extra, "origin/dev-fixture-1")

	stdout, stderr, code := w.swarm(t, "status", "--run", "fixture-run", "--json", "--full-result")
	// Either exit is fine here — this test is about attribution, not
	// about the verdict a particular fixture's ready front implies.
	if code != 0 && code != 1 {
		t.Fatalf("status exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var env swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("status --json did not parse: %v\n%s", err, stdout)
	}

	var dispatchLines []string
	for _, f := range env.Facts {
		if f.Name == "dispatches" && f.Value != nil && f.Value.Items != nil {
			dispatchLines = *f.Value.Items
		}
	}
	if len(dispatchLines) == 0 {
		t.Fatalf("no dispatches were reported at all:\n%s", stdout)
	}

	var found string
	for _, line := range dispatchLines {
		if strings.HasPrefix(line, "dev-fixture-1:") {
			found = line
		}
	}
	if found == "" {
		t.Fatalf("dev-fixture-1 was not named among the dispatches:\n%v", dispatchLines)
	}
	if !strings.Contains(found, extra) {
		t.Errorf("dev-fixture-1's worktree was not attributed by its commit, though the branch carries "+
			"no dispatch id at all:\n%s", found)
	}

	// The negative control: the SAME commit, on a branch shaped like the
	// old convention this replaces, is attributed identically — the
	// branch's name plays no part either way.
	other := restore(t, binary, "core")
	otherExtra := filepath.Join(other.dir, "bb-worktree")
	other.mustRun(t, other.repo, "git", "worktree", "add",
		"-b", "bb/dev-fixture-1", otherExtra, "origin/dev-fixture-1")

	stdout2, stderr2, code2 := other.swarm(t, "status", "--run", "fixture-run", "--json", "--full-result")
	if code2 != 0 && code2 != 1 {
		t.Fatalf("status exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code2, stdout2, stderr2)
	}
	var env2 swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(stdout2), &env2); err != nil {
		t.Fatalf("status --json did not parse: %v\n%s", err, stdout2)
	}
	var dispatchLines2 []string
	for _, f := range env2.Facts {
		if f.Name == "dispatches" && f.Value != nil && f.Value.Items != nil {
			dispatchLines2 = *f.Value.Items
		}
	}
	var found2 string
	for _, line := range dispatchLines2 {
		if strings.HasPrefix(line, "dev-fixture-1:") {
			found2 = line
		}
	}
	if !strings.Contains(found2, otherExtra) {
		t.Errorf("dev-fixture-1's worktree was not attributed on a bb/-shaped branch either:\n%s", found2)
	}
}
