// Unit tests for `swarm ready`: the run-scope refusal, and a tracker
// response that is not a queue's shape reading as unanswered rather than
// as a length derived from it. Every scenario runs through the Beads
// Tracker's real argv-building against the fake process runner, so no
// test starts `bd`.
//
// AC-SWARM-DISP-002.1, .2, .7, .8, .9, .10 and .17 are declared e2e and
// live in cli/swarm/e2e against the core and dispatch-worktree fixtures.
package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

func readyDeps(fake *exectest.Fake) status.Deps {
	tr := beadstracker.Tracker{}
	return status.Deps{
		Tracker: tr,
		VCS:     gitvcs.VersionControl{},
		Store:   manifestbeads.Store{Tracker: tr},
		Runner:  plan.Runner{Exec: fake},
	}
}

// AC: AC-SWARM-DISP-002.3
//
// An unscoped invocation is refused with the malformed-invocation status,
// and nothing is read — asserted from the fake's own record.
func TestReadyRefusesAnEmptyScope(t *testing.T) {
	for _, args := range [][]string{
		{"ready"},
		{"ready", "--run", ""},
		{"ready", "--run", "   "},
	} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
		root := newRootCmd(streams)
		root.AddCommand(readyCmd(streams, readyDeps(fake)))

		code := dispatch(root, streams, args, &bytes.Buffer{}, &bytes.Buffer{})
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read something outside the scope: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-DISP-002.16
//
// A tracker response that is an error body rather than a list is not
// counted as a queue of that body's length; it fails to decode, and the
// command takes the stop status.
func TestReadyReportsANonListPayloadAsUnanswered(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:   []string{"bd", "ready", "--parent", "run-1", "--exclude-type", "epic", "--exclude-label", "human", "--limit", "0", "--json"},
		Stdout: []byte(`{"error":"database is locked"}`),
	})
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	code := dispatch(root, streams, []string{"ready", "--run", "run-1", "--json"}, &out, &errOut)
	if code != int(StatusStop) {
		t.Errorf("exit status = %d, want %d for a payload that is not a queue's shape", code, StatusStop)
	}
	if strings.Contains(out.String(), `"ready_count":1`) {
		t.Errorf("the error body's own length leaked into ready_count:\n%s", out.String())
	}
}

// breakerDeps builds the seam [applyBreaker] reads and writes through —
// only the manifest store and the runner, since the breaker touches
// neither the ready query nor version control.
func breakerDeps(fake *exectest.Fake) status.Deps {
	tr := beadstracker.Tracker{}
	return status.Deps{
		Store:  manifestbeads.Store{Tracker: tr},
		Runner: plan.Runner{Exec: fake},
	}
}

func showRecording(run string, metadata string) exectest.Recording {
	stdout := `[{"id":"` + run + `","status":"open","issue_type":"epic"`
	if metadata != "" {
		stdout += `,"metadata":` + metadata
	}
	stdout += `}]`
	return exectest.Recording{Argv: []string{"bd", "show", run, "--json"}, Stdout: []byte(stdout)}
}

func updateRecording(run, payload string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "update", run, "--metadata", payload, "--json"},
		Stdout: []byte(`[{"id":"` + run + `","status":"open","issue_type":"epic","metadata":` + payload + `}]`),
	}
}

// AC: AC-SWARM-DISP-002.4
//
// Consecutive dispatches that completed nothing stop the run: two are
// already recorded, and this dispatch return — reported with --completed
// 0 — is the third against a threshold of three.
func TestBreakerStopsTheRunAtItsThreshold(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":2}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":3}`),
	)

	facts, stopped, summary, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if !stopped {
		t.Errorf("stopped = false, want true at the threshold (summary: %q)", summary)
	}
	if summary == "" {
		t.Error("a stopped run names no summary")
	}
	if got := *facts[0].Value.Count; got != 3 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 3", got)
	}
}

// AC: AC-SWARM-DISP-002.5
//
// A dispatch that completed some of its work resets the breaker to zero,
// regardless of how many beads it also blocked.
func TestBreakerResetsToZeroWhenSomethingCompleted(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":2}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":0}`),
	)

	facts, stopped, _, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: true, Completed: 4}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if stopped {
		t.Error("a dispatch that completed 4 beads reads as stopping the run")
	}
	if got := *facts[0].Value.Count; got != 0 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 0", got)
	}
}

// AC: AC-SWARM-DISP-002.6 @negative
//
// One dispatch return in which a bead blocked and five were returned not
// reached advances the count by exactly one — the unit is the RETURN, not
// the beads inside it, so this is asserted by the fact that
// [breakerReport] carries only how many beads the return completed and
// nothing about how many it did not reach.
func TestBreakerCountsOneDispatchReturnNotItsBeads(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", ""), // no consecutive_dispatch_failures written yet
		updateRecording("run-1", `{"consecutive_dispatch_failures":1}`),
	)

	facts, stopped, _, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if stopped {
		t.Error("one failed return against a threshold of 3 reads as stopping the run")
	}
	if got := *facts[0].Value.Count; got != 1 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 1 (not 6)", got)
	}
}

// A bare re-query — no --completed — only reads the count; it neither
// advances nor resets it, even when the recorded count already meets the
// threshold.
func TestBreakerBareQueryReadsWithoutWriting(t *testing.T) {
	fake := exectest.NewFake(showRecording("run-1", `{"consecutive_dispatch_failures":3}`))

	facts, stopped, summary, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: false}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if !stopped {
		t.Errorf("a bare query against a count already at the threshold does not report stopped (summary: %q)", summary)
	}
	if got := *facts[0].Value.Count; got != 3 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 3", got)
	}
	if calls := fake.Argvs(); len(calls) != 1 {
		t.Errorf("a bare query wrote something: %v", calls)
	}
}

// An absent count reads as zero — the manifest's own "not yet written" is
// distinct from a resumed run's zero, and reconciling them is this
// command's decision.
func TestBreakerAbsentCountReadsAsZero(t *testing.T) {
	fake := exectest.NewFake(showRecording("run-1", ""))

	facts, stopped, _, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: false}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if stopped {
		t.Error("an absent count reads as stopping the run")
	}
	if got := *facts[0].Value.Count; got != 0 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 0", got)
	}
}

// A reset that would not change the recorded value writes nothing — the
// manifest's counter-backwards rule permits a write of exactly 0, but a
// write that changes nothing is still a write [[container-swarm-cli]]'s
// concurrent-write guard has no reason to risk.
func TestBreakerSkipsAWriteThatWouldNotChangeTheCount(t *testing.T) {
	fake := exectest.NewFake(showRecording("run-1", `{"consecutive_dispatch_failures":0}`))

	facts, stopped, _, err := applyBreaker(context.Background(), breakerDeps(fake),
		Request{Run: "run-1"}, breakerReport{Given: true, Completed: 5}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if stopped {
		t.Error("a freshly-reset count reads as stopping the run")
	}
	if got := *facts[0].Value.Count; got != 0 {
		t.Errorf("consecutive_dispatch_failures reported as %d, want 0", got)
	}
	if calls := fake.Argvs(); len(calls) != 1 {
		t.Errorf("a no-op reset wrote something: %v", calls)
	}
}
