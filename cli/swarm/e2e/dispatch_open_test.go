package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// AC: AC-SWARM-DISP-003.3
//
// A branch already carrying the name this run would use, with the
// manifest recording none, stops the run rather than being claimed for
// it — named, with an explicit --branch as the resolution — and nothing
// is created or recorded.
func TestDispatchOpenRefusesACollidingBranchTheManifestDoesNotRecord(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	// A run epic with no manifest at all — no "branch" key to reuse —
	// distinct from fixture-run so this test needs no key to be stripped
	// from an existing one.
	run := w.createID(t, "bd", "create", "A Second Run", "--type", "epic", "--json")

	// The branch this run would compute for itself, already occupied by
	// something else.
	w.mustRun(t, w.repo, "git", "branch", "sprint/"+run, "HEAD")

	stdout, stderr, code := w.swarm(t, "dispatch", "open", "--run", run, "--json")
	if code != 1 {
		t.Fatalf("dispatch open exited %d, want 1 (attention)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if !strings.Contains(stdout, "sprint/"+run) {
		t.Errorf("the refusal does not name the colliding branch:\n%s", stdout)
	}
	if !strings.Contains(stdout, "--branch") {
		t.Errorf("the refusal does not name an explicit --branch as the resolution:\n%s", stdout)
	}

	shown := w.mustCapture(t, w.repo, "bd", "show", run, "--json")
	var issues []struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(shown), &issues); err != nil || len(issues) != 1 {
		t.Fatalf("bd show %s did not answer one issue: %v\n%s", run, err, shown)
	}
	if _, present := issues[0].Metadata["branch"]; present {
		t.Errorf("the manifest recorded a branch despite the refusal: %v", issues[0].Metadata)
	}
}

// AC: AC-SWARM-DISP-003.5
//
// The sprint branch is created and the sha it was cut from is recorded
// in the same write, against a real `bd` and a real `git branch`.
func TestDispatchOpenCreatesAndRecordsTheSprintBranch(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	run := w.createID(t, "bd", "create", "A Second Run", "--type", "epic", "--json")
	headSHA := w.mustCapture(t, w.repo, "git", "rev-parse", "HEAD")
	branchName := "sprint/" + run + "-custom"

	stdout, stderr, code := w.swarm(t, "dispatch", "open", "--run", run,
		"--branch", branchName, "--base", "HEAD", "--json")
	if code != 0 {
		t.Fatalf("dispatch open exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		Branch  string `json:"branch"`
		BaseSHA string `json:"base_sha"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	if answer.Branch != branchName {
		t.Errorf("branch = %q, want %q", answer.Branch, branchName)
	}
	if answer.BaseSHA != headSHA {
		t.Errorf("base_sha = %q, want %q (HEAD at the time the command ran)", answer.BaseSHA, headSHA)
	}

	// The branch is real, and points at the sha the manifest now records.
	createdSHA := w.mustCapture(t, w.repo, "git", "rev-parse", branchName)
	if createdSHA != headSHA {
		t.Errorf("git rev-parse %s = %q, want %q", branchName, createdSHA, headSHA)
	}

	// The manifest carries both keys, read back through `bd` directly —
	// not just through the command that just wrote them.
	shown := w.mustCapture(t, w.repo, "bd", "show", run, "--json")
	var issues []struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(shown), &issues); err != nil || len(issues) != 1 {
		t.Fatalf("bd show %s did not answer one issue: %v\n%s", run, err, shown)
	}
	if got, _ := issues[0].Metadata["branch"].(string); got != branchName {
		t.Errorf("the tracker's own record of branch = %q, want %q", got, branchName)
	}
	if got, _ := issues[0].Metadata["base_sha"].(string); got != headSHA {
		t.Errorf("the tracker's own record of base_sha = %q, want %q", got, headSHA)
	}
}
