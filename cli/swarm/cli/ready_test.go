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

// breakerPass runs one pass of the breaker and unpacks the three values
// these tests assert on, so each one states its scenario and its
// expectation and nothing about the shape of the return. The records the
// pass produced are asserted through the command instead, since it is the
// rendered result a consumer reads a refused write from.
func breakerPass(ctx context.Context, deps status.Deps, req Request, report breakerReport, threshold int) (
	[]Fact, bool, string, error,
) {
	outcome, err := applyBreaker(ctx, deps, req, report, threshold)
	return outcome.Facts, outcome.Stopped, outcome.Summary, err
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

	facts, stopped, summary, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: true, Completed: 0}, 3)
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

	facts, stopped, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: true, Completed: 4}, 3)
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

	facts, stopped, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: true, Completed: 0}, 3)
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

	facts, stopped, summary, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: false}, 3)
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

	facts, stopped, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: false}, 3)
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

	facts, stopped, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: true, Completed: 5}, 3)
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

// runReadyAgent drives `swarm ready --agent` through the real command
// tree against the fake, with the queue reads accept_test.go's own helper
// declares for the same run and dispatch id.
func runReadyAgent(t *testing.T, extra []exectest.Recording, args ...string) (*exectest.Fake, string, int) {
	t.Helper()

	fake := exectest.NewFake(append(drainedQueue(), extra...)...)
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	code := dispatch(root, streams, args, &out, &errOut)
	return fake, out.String(), code
}

// --agent narrows the answer; it does not discard the dispatch return.
// Reporting a return with --completed advances the breaker whether or not
// the same invocation also asked for one dispatch's queue — which is the
// invocation a coordinator most wants at the moment a dispatch comes
// back.
func TestTheAgentArmAppliesTheBreaker(t *testing.T) {
	fake, stdout, code := runReadyAgent(t, []exectest.Recording{
		showRecording(acceptRun, `{"consecutive_dispatch_failures":1}`),
		updateRecording(acceptRun, `{"consecutive_dispatch_failures":2}`),
	}, "ready", "--run", acceptRun, "--agent", acceptAgent, "--completed", "0",
		"--full-result", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	wrote := false
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			wrote = true
		}
	}
	if !wrote {
		t.Errorf("--agent --completed 0 never advanced the count: %v", fake.Argvs())
	}

	env := envelopeOf(t, stdout)
	count, ok := factNamed(env, "consecutive_dispatch_failures")
	if !ok {
		t.Fatalf("the --agent arm reports no breaker count at all:\n%s", stdout)
	}
	if !count.Established || count.Value == nil || count.Value.Count == nil || *count.Value.Count != 2 {
		t.Errorf("consecutive_dispatch_failures = %v, want an established 2", count.Value)
	}

	// The queue answer is still the --agent arm's own.
	if _, ok := factNamed(env, "drained"); !ok {
		t.Errorf("the breaker displaced the dispatch's own queue answer:\n%s", stdout)
	}
}

// A return that completed something resets the count, on this arm as on
// the other.
func TestTheAgentArmResetsTheBreakerOnACompletedReturn(t *testing.T) {
	fake, stdout, code := runReadyAgent(t, []exectest.Recording{
		showRecording(acceptRun, `{"consecutive_dispatch_failures":2}`),
		updateRecording(acceptRun, `{"consecutive_dispatch_failures":0}`),
	}, "ready", "--run", acceptRun, "--agent", acceptAgent, "--completed", "3",
		"--full-result", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	reset := false
	for _, argv := range fake.Argvs() {
		if len(argv) >= 4 && argv[1] == "update" && strings.Contains(argv[4], `"consecutive_dispatch_failures":0`) {
			reset = true
		}
	}
	if !reset {
		t.Errorf("--agent --completed 3 did not reset the count: %v", fake.Argvs())
	}
}

// The breaker's write obeys the request's mode on this arm too, so a dry
// run of it writes nothing.
func TestTheAgentArmsBreakerWritesNothingInADryRun(t *testing.T) {
	fake, stdout, _ := runReadyAgent(t, []exectest.Recording{
		showRecording(acceptRun, `{"consecutive_dispatch_failures":1}`),
		updateRecording(acceptRun, `{"consecutive_dispatch_failures":2}`),
	}, "ready", "--run", acceptRun, "--agent", acceptAgent, "--completed", "0",
		"--dry-run", "--full-result", "--json")

	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			t.Errorf("a dry run of the --agent arm wrote: %v\n%s", argv, stdout)
		}
	}
}

// Explain names what the invocation would do, so an --agent invocation
// carrying --completed says that the breaker is read and written back —
// otherwise the transcript itself would document the flag being dropped.
func TestExplainOfTheAgentArmNamesTheBreaker(t *testing.T) {
	fake := exectest.NewFake() // no recordings: explain runs nothing
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	code := dispatch(root, streams, []string{"ready", "--run", acceptRun, "--agent", acceptAgent,
		"--completed", "0", "--explain"}, &out, &errOut)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, out.String())
	}
	if calls := fake.Argvs(); len(calls) != 0 {
		t.Errorf("explain ran %d invocation(s): %v", len(calls), calls)
	}
	if !strings.Contains(out.String(), "bd show "+acceptRun) {
		t.Errorf("the --agent transcript does not name the breaker's read:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "--completed") {
		t.Errorf("the --agent transcript does not say --completed writes the count back:\n%s", out.String())
	}
}

// The breaker's write is this command's only write, and `swarm ready`'s
// own --dry-run help promises "it writes nothing either way". A dry run
// therefore refuses the write rather than performing it — [plan.Runner]'s
// own guarantee, once the mode reaches it.
//
// Asserted from the fake's own record, so a command that returned a
// clean-looking result while writing would still fail here.
func TestTheBreakerWritesNothingInADryRun(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":1}`),
		// Declared deliberately, and succeeding if it is reached: an
		// undeclared write would fail as a missing recording instead,
		// which would pass this test for the wrong reason.
		updateRecording("run-1", `{"consecutive_dispatch_failures":2}`),
	)

	facts, stopped, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeDryRun}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}

	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			t.Errorf("a dry run wrote: %v", argv)
		}
	}

	if len(facts) != 1 {
		t.Fatalf("applyBreaker reported %d fact(s), want 1", len(facts))
	}
	if facts[0].Established {
		t.Errorf("the count is reported as established at %v; a refused write leaves it unknown, "+
			"not at the value the write would have set", facts[0].Value)
	}
	if stopped {
		t.Error("a dry run stopped the run on the strength of a write it refused to make")
	}
}

// The negative control for the test above: the same invocation in execute
// mode does write and does report the new count, so what that test
// asserts is the mode reaching the runner rather than the breaker having
// been disabled.
func TestTheBreakerStillWritesWhenItReallyRuns(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":1}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":2}`),
	)

	facts, _, _, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeExecute}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}

	wrote := false
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			wrote = true
		}
	}
	if !wrote {
		t.Fatalf("execute mode issued no write: %v", fake.Argvs())
	}
	if !facts[0].Established || facts[0].Value == nil || *facts[0].Value.Count != 2 {
		t.Errorf("consecutive_dispatch_failures = %v (established %v), want an established 2",
			facts[0].Value, facts[0].Established)
	}
}

// The refused write reaches a --full-result consumer as a refused STEP
// record, which is what tells it the write was declined rather than never
// attempted. An early return guarded on the mode would skip the write and
// leave no record of it at all, which is why the mode is threaded into
// the runner instead.
func TestADryRunOfReadyRecordsTheBreakerWriteAsRefused(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{
			Argv: []string{"bd", "ready", "--parent", "run-1", "--exclude-type", "epic",
				"--exclude-label", "human", "--limit", "0", "--json"},
			Stdout: []byte("[]"),
		},
		showRecording("run-1", `{"consecutive_dispatch_failures":1}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":2}`),
	)

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	dispatch(root, streams, []string{"ready", "--run", "run-1", "--dry-run",
		"--completed", "0", "--full-result", "--json"}, &out, &errOut)

	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			t.Errorf("a dry run of the command wrote: %v", argv)
		}
	}
	if !strings.Contains(out.String(), "advance-failure-count") {
		t.Errorf("the refused write is not among the invocations the result reports:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "refused") {
		t.Errorf("the result does not report the write as refused:\n%s", out.String())
	}
}

// A dry run may not INVENT a stop from a number nothing wrote, and it
// must not HIDE one from a number something did. When the manifest
// already records the count at the threshold, the read alone establishes
// that the run is stopped — so the refused write withholds the projected
// advance and nothing else.
func TestADryRunReportsAStopTheRecordedCountAlreadyEstablishes(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":3}`),
		// Declared so an attempted write succeeds rather than failing as
		// a missing recording, which would pass this test for the wrong
		// reason.
		updateRecording("run-1", `{"consecutive_dispatch_failures":4}`),
	)

	facts, stopped, summary, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeDryRun}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if !stopped {
		t.Errorf("stopped = false at a RECORDED count of 3 against a threshold of 3; " +
			"the read established it and a refused write does not unestablish it")
	}
	if !strings.Contains(summary, "the run is stopped") {
		t.Errorf("summary = %q, want it to say the run is stopped", summary)
	}
	// Phrased from the recorded count, never from the projection the
	// refused write would have set.
	if strings.Contains(summary, "4 consecutive") {
		t.Errorf("summary = %q cites the projected 4; a refused write recorded nothing", summary)
	}

	if len(facts) != 1 {
		t.Fatalf("applyBreaker reported %d fact(s), want 1", len(facts))
	}
	if facts[0].Established {
		t.Errorf("the count is reported established at %v; a refused write leaves it unknown",
			facts[0].Value)
	}
	if !strings.Contains(facts[0].Reason, "would have become 4") {
		t.Errorf("the projection is not in the count's reason: %q", facts[0].Reason)
	}

	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			t.Errorf("a dry run wrote: %v", argv)
		}
	}
}

// The regression guard for the test above: BELOW the threshold a refused
// write still leaves the run clear, so what that test asserts is a stop
// resting on the RECORDED count rather than the stop test having been
// moved onto the projection.
func TestADryRunBelowTheThresholdStaysClear(t *testing.T) {
	fake := exectest.NewFake(
		showRecording("run-1", `{"consecutive_dispatch_failures":2}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":3}`),
	)

	_, stopped, summary, err := breakerPass(context.Background(), breakerDeps(fake),
		Request{Run: "run-1", Mode: plan.ModeDryRun}, breakerReport{Given: true, Completed: 0}, 3)
	if err != nil {
		t.Fatalf("applyBreaker: %v", err)
	}
	if stopped {
		t.Errorf("a dry run stopped the run from a recorded 2 against a threshold of 3 "+
			"(summary %q); the 3 it would have written is a number nothing wrote", summary)
	}
}

// The whole command, not the breaker alone: `--dry-run --completed 0` on
// a run the manifest already records at the threshold exits 2 and says
// so. The same dry run WITHOUT --completed already did; adding the flag
// must not suppress the verdict.
func TestReadyDryRunWithCompletedExitsStopAtARecordedThreshold(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{
			Argv: []string{"bd", "ready", "--parent", "run-1", "--exclude-type", "epic",
				"--exclude-label", "human", "--limit", "0", "--json"},
			Stdout: []byte("[]"),
		},
		showRecording("run-1", `{"consecutive_dispatch_failures":3}`),
		updateRecording("run-1", `{"consecutive_dispatch_failures":4}`),
	)

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	code := dispatch(root, streams, []string{"ready", "--run", "run-1", "--dry-run",
		"--completed", "0", "--full-result", "--json"}, &out, &errOut)
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (the run is stopped)\n%s", code, StatusStop, out.String())
	}
	if !strings.Contains(out.String(), "the run is stopped") {
		t.Errorf("the summary does not say the run is stopped:\n%s", out.String())
	}
}

// The --agent arm reaches the same applyBreaker, so it inherits this —
// asserted rather than assumed.
func TestTheAgentArmReportsAStopTheRecordedCountEstablishesInADryRun(t *testing.T) {
	_, stdout, code := runReadyAgent(t, []exectest.Recording{
		showRecording(acceptRun, `{"consecutive_dispatch_failures":3}`),
		updateRecording(acceptRun, `{"consecutive_dispatch_failures":4}`),
	}, "ready", "--run", acceptRun, "--agent", acceptAgent, "--completed", "0",
		"--dry-run", "--full-result", "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (the run is stopped)\n%s", code, StatusStop, stdout)
	}
	if !strings.Contains(stdout, "the run is stopped") {
		t.Errorf("the --agent arm does not say the run is stopped:\n%s", stdout)
	}
}
