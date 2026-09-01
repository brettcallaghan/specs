package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// AC: AC-SWARM-DISP-001.1 — proving delta
// AC: AC-SWARM-DISP-001.2
// AC: AC-SWARM-DISP-001.6 — proving delta
// AC: AC-SWARM-DISP-001.7
// AC: AC-SWARM-DISP-001.8 — proving delta
//
// One world proves all five together, against real `bd`:
//
//   - Closing fixture-run.2 makes the target epic eligible on the same
//     rule as a ceremony phase (.7), and its own id is reported (.2).
//   - Phase B is blocked-by Phase A at the EPIC level (`bd dep add
//     fixture-run.phase-b fixture-run.phase-a --type blocks` — the only
//     shape bd's own dependency rule allows toward an epic: measured on
//     this bd, `tasks can only block other tasks, not epics` and `epics
//     can only block other epics, not tasks`). Phase B's own child stays
//     open, so Phase B is still not eligible, and its open child is named
//     as the reason (.6).
//   - Closing Phase A resolves that epic-level block, and
//     fixture-run.phase-b.1 — ready neither before this run nor by
//     virtue of any change to itself — enters the ready set as a result.
//     Reported as released (.8).
//   - A foreign epic outside fixture-run, made eligible the same way, is
//     never touched and never named (.1).
func TestPhaseCloseClosesOnlyItsOwnRunAndNamesWhatItReleased(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "phase-epics")

	// The target's second child closes, so the target is eligible on the
	// same rule as a ceremony phase.
	w.mustRun(t, w.repo, "bd", "update", "fixture-run.2", "--status", "closed", "--json")

	// Phase B is gated on Phase A at the epic level.
	w.mustRun(t, w.repo, "bd", "dep", "add", "fixture-run.phase-b", "fixture-run.phase-a",
		"--type", "blocks", "--json")

	// A foreign epic, outside fixture-run entirely, made eligible the
	// same way an in-run one would be.
	foreignEpic := w.createID(t, "bd", "create", "Foreign Epic", "--type", "epic", "--json")
	foreignChild := w.createID(t, "bd", "create", "Foreign child", "--type", "task",
		"--parent", foreignEpic, "--json")
	w.mustRun(t, w.repo, "bd", "update", foreignChild, "--status", "closed", "--json")

	// Baseline: before this command runs, phase-b.1 is not ready — the
	// epic-level block already gates it, which is what makes "released
	// by THIS run" provable rather than assumed.
	before := w.mustCapture(t, w.repo, "bd", "ready", "--parent", "fixture-run",
		"--exclude-type", "epic", "--limit", "0", "--json")
	if strings.Contains(before, "fixture-run.phase-b.1") {
		t.Fatalf("fixture-run.phase-b.1 already reads ready before the command ran:\n%s", before)
	}

	stdout, stderr, code := w.swarm(t, "phase", "close", "--run", "fixture-run", "--json")
	if code != 0 {
		t.Fatalf("phase close exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var answer struct {
		ClosedEpics []string `json:"closed_epics"`
		LeftOpen    []struct {
			Epic      string `json:"epic"`
			OpenChild string `json:"open_child"`
		} `json:"left_open"`
		ReleasedReady []string `json:"released_ready"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}

	// AC-SWARM-DISP-001.2, .7: both closed, named.
	if got := strings.Join(answer.ClosedEpics, ","); got != "fixture-run.phase-a,fixture-run.target" {
		t.Errorf("closed_epics = %v, want [fixture-run.phase-a fixture-run.target]", answer.ClosedEpics)
	}

	// AC-SWARM-DISP-001.1: the foreign epic — eligible, outside the run —
	// is named nowhere in the output.
	if strings.Contains(stdout, foreignEpic) {
		t.Errorf("the foreign epic %s appears in output scoped to fixture-run:\n%s", foreignEpic, stdout)
	}
	foreignState := w.mustCapture(t, w.repo, "bd", "show", foreignEpic, "--json")
	var shown []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(foreignState), &shown); err != nil || len(shown) != 1 {
		t.Fatalf("bd show %s did not answer one issue: %v\n%s", foreignEpic, err, foreignState)
	}
	if shown[0].Status != "open" {
		t.Errorf("the foreign epic's status = %q, want open", shown[0].Status)
	}

	// AC-SWARM-DISP-001.6: Phase B stays open, its own open child named.
	var phaseB *struct {
		Epic      string `json:"epic"`
		OpenChild string `json:"open_child"`
	}
	for i := range answer.LeftOpen {
		if answer.LeftOpen[i].Epic == "fixture-run.phase-b" {
			phaseB = &answer.LeftOpen[i]
		}
	}
	if phaseB == nil {
		t.Fatalf("fixture-run.phase-b is not reported left open:\n%s", stdout)
	}
	if phaseB.OpenChild != "fixture-run.phase-b.1" {
		t.Errorf("phase-b's reason = %q, want fixture-run.phase-b.1", phaseB.OpenChild)
	}

	// AC-SWARM-DISP-001.8: the epic-level block Phase A's closure
	// resolved releases fixture-run.phase-b.1 into the ready set.
	if got := strings.Join(answer.ReleasedReady, ","); got != "fixture-run.phase-b.1" {
		t.Errorf("released_ready = %v, want [fixture-run.phase-b.1]", answer.ReleasedReady)
	}
}

// AC: AC-SWARM-DISP-001.4
//
// The core fixture's only epic under the run has one child still in
// progress, so nothing is eligible — reported as such, and clear. The
// case-driven equivalent is
// cli/testdata/swarm-cases/phase-close-none-eligible; this is the same
// scenario as a second confirmation that a real `bd` agrees.
func TestPhaseCloseWithNothingEligibleReportsSoAndExitsClear(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	stdout, stderr, code := w.swarm(t, "phase", "close", "--run", "fixture-run", "--json")
	if code != 0 {
		t.Fatalf("phase close exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}
	var answer struct {
		ClosedEpics []string `json:"closed_epics"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	if len(answer.ClosedEpics) != 0 {
		t.Errorf("closed_epics = %v, want none: nothing under fixture-run is eligible\n%s",
			answer.ClosedEpics, stdout)
	}
}
