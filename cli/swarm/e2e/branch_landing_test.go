package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// acceptingRemote stands a bare repository up as `origin` and returns
// nothing — the world's clone has the bundle file as its origin, which no
// push reaches, so a case that pushes has to be given a remote that
// accepts.
//
// The refusing counterpart is the `push-refused` fixture's own, declared
// in its stamp and stood up by the harness; this is its opposite, and the
// only thing separating the two cases.
func acceptingRemote(t *testing.T, w world) {
	t.Helper()
	remote := w.dir + "/accepting-remote.git"
	w.mustRun(t, w.dir, "git", "init", "--quiet", "--bare", remote)
	w.mustRun(t, w.repo, "git", "remote", "remove", "origin")
	w.mustRun(t, w.repo, "git", "remote", "add", "origin", remote)
}

// landing is the answer both verbs print, in whichever of its two shapes
// the exit status selected.
type landing struct {
	Branch   string
	Remote   string
	Landed   bool
	Unpushed int
}

// landingOf reads the answer, which the exit status shapes: a clear exit
// prints each property's value directly, and every other status prints a
// verdict per property carrying that value under `value`. A case
// asserting across both statuses has to read both, which is what this
// does — and reading only the clear shape is why it would otherwise
// assert nothing about the interesting half.
func landingOf(t *testing.T, stdout string) landing {
	t.Helper()

	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	value := func(name string) any {
		v, ok := doc[name]
		if !ok {
			return nil
		}
		if verdict, isVerdict := v.(map[string]any); isVerdict {
			return verdict["value"]
		}
		return v
	}

	var answer landing
	if s, ok := value("branch").(string); ok {
		answer.Branch = s
	}
	if s, ok := value("remote").(string); ok {
		answer.Remote = s
	}
	if b, ok := value("landed").(bool); ok {
		answer.Landed = b
	}
	if n, ok := value("unpushed").(float64); ok {
		answer.Unpushed = int(n)
	}
	return answer
}

// AC: AC-SWARM-DISP-005.2
//
// A branch with no counterpart on the remote is not reported as landed:
// the accepting remote is empty, so the sprint branch has no counterpart
// there at all.
func TestVerifyLandedRefusesABranchWithNoCounterpart(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")
	acceptingRemote(t, w)

	stdout, stderr, code := w.swarm(t, "branch", "verify-landed", "--json")
	if code != 1 {
		t.Fatalf("verify-landed exited %d, want 1 (attention)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if strings.Contains(stdout, `"landed": true`) {
		t.Errorf("a branch with no counterpart was reported as landed:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-005.1, .4
//
// Unpushed commits are reported as unpushed with their count and the
// attention status; once pushed, the same check exits zero. Both halves
// in one case, because the second is only meaningful against the first:
// a check that reported "landed" for every branch would pass the second
// alone.
func TestVerifyLandedCountsUnpushedCommitsAndReportsCleanOnceTheyLand(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")
	acceptingRemote(t, w)

	// A counterpart exists but is one commit behind: push, then commit.
	w.mustRun(t, w.repo, "git", "push", "--quiet", "origin", "claude/bead-bonanza-fixture")
	w.mustRun(t, w.repo, "git", "commit", "--quiet", "--allow-empty", "-m", "fixture-run.9 P1 one more")

	stdout, stderr, code := w.swarm(t, "branch", "verify-landed", "--json")
	if code != 1 {
		t.Fatalf("verify-landed exited %d, want 1 (attention)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	answer := landingOf(t, stdout)
	if answer.Landed {
		t.Errorf("a branch one commit ahead of its remote was reported as landed:\n%s", stdout)
	}
	if answer.Unpushed != 1 {
		t.Errorf("unpushed = %d, want 1", answer.Unpushed)
	}
	if answer.Branch != "claude/bead-bonanza-fixture" || answer.Remote != "origin" {
		t.Errorf("branch/remote = %q/%q, want the checked-out branch and origin", answer.Branch, answer.Remote)
	}

	// AC-SWARM-DISP-005.4: pushed, the same check exits zero.
	stdout, stderr, code = w.swarm(t, "branch", "push", "--json")
	if code != 0 {
		t.Fatalf("branch push exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}
	if pushed := landingOf(t, stdout); !pushed.Landed || pushed.Unpushed != 0 {
		t.Errorf("after the push: landed=%v unpushed=%d, want true and 0\n%s",
			pushed.Landed, pushed.Unpushed, stdout)
	}

	stdout, stderr, code = w.swarm(t, "branch", "verify-landed", "--json")
	if code != 0 {
		t.Fatalf("verify-landed after the push exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if !landingOf(t, stdout).Landed {
		t.Errorf("a fully pushed branch was not reported as landed:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-005.3, .7
//
// A push the remote rejects is surfaced with its cause, and nothing in it
// is reported as landed — against a real bare repository whose
// pre-receive hook declines the sprint branch's namespace, so the
// refusal comes back through git the way a server's does.
func TestBranchPushSurfacesARejectionWithItsCause(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "push-refused")

	stdout, stderr, code := w.swarm(t, "branch", "push", "--json")
	if code != 2 {
		t.Fatalf("branch push exited %d, want 2 (stop)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}
	if strings.Contains(stdout, `"landed": true`) {
		t.Errorf("a rejected push reported something as landed:\n%s", stdout)
	}
	// The remote's own message is the cause.
	if !strings.Contains(stdout, "this remote does not accept sprint branches") {
		t.Errorf("the report does not carry the rejection's cause:\n%s", stdout)
	}
	// And the refused ref is named.
	if !strings.Contains(stdout, "claude/bead-bonanza-fixture") {
		t.Errorf("the report does not name the refused ref:\n%s", stdout)
	}
}
