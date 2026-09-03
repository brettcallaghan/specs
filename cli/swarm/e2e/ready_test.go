package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// AC: AC-SWARM-DISP-002.1
// AC: AC-SWARM-DISP-002.2 (via the relied-upon Status Collector)
// AC: AC-SWARM-DISP-002.8
//
// A bead labelled for a human and a deferred one are both excluded from
// the dispatch set — but the human-labelled one still counts as
// outstanding work in `swarm status`'s completion tally, which is what
// AC-SWARM-DISP-002.2 asks: the exclusion is the dispatch query's
// convention alone, never the completion count's.
func TestReadyExcludesHumanAndDeferredBeadsButStatusStillCountsTheHumanOne(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	// The core fixture's baseline completion count, read before either
	// bead exists, so the count after can be compared against it rather
	// than against a hand-derived expectation.
	before := completionTotal(t, w)

	w.mustRun(t, w.repo, "bd", "create", "Needs a human", "--type", "task",
		"--parent", "fixture-run.target", "-l", "human", "--json")
	w.mustRun(t, w.repo, "bd", "create", "Deferred work", "--type", "task",
		"--parent", "fixture-run.target", "--defer", "+1d", "--json")

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--json")
	if code != 0 {
		t.Fatalf("ready exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}
	if strings.Contains(stdout, "target.1") || strings.Contains(stdout, "human") {
		t.Errorf("the human-labelled bead reached the dispatch set:\n%s", stdout)
	}
	if strings.Contains(stdout, "target.2") {
		t.Errorf("the deferred bead reached the dispatch set:\n%s", stdout)
	}

	after := completionTotal(t, w)
	if after != before+2 {
		t.Errorf("swarm status's completion_total = %d after adding 2 beads (was %d), want %d — "+
			"the human-labelled bead must still be counted as outstanding work",
			after, before, before+2)
	}
}

// completionTotal reads swarm status's own completion_total, which is
// what AC-SWARM-DISP-002.2 asks the human-labelled bead still be part of
// — read through the real command rather than assumed, so a change to
// how the Status Collector counts is caught here too.
//
// The core fixture reads a stale commit against `swarm status`'s own
// threshold, so its answer takes the attention shape — a verdict object
// per property, `{ok, value, issues}` — rather than the clear shape's
// bare value; both are decoded here since which one this run gets is not
// this test's concern.
func completionTotal(t *testing.T, w world) int {
	t.Helper()

	stdout, stderr, code := w.swarm(t, "status", "--run", "fixture-run", "--json")
	if code != 0 && code != 1 {
		t.Fatalf("status exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &fields); err != nil {
		t.Fatalf("the status answer does not parse: %v\n%s", err, stdout)
	}
	raw, ok := fields["completion_total"]
	if !ok {
		t.Fatalf("the status answer carries no completion_total:\n%s", stdout)
	}

	var direct int
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct
	}
	var verdict struct {
		Value int `json:"value"`
	}
	if err := json.Unmarshal(raw, &verdict); err != nil {
		t.Fatalf("completion_total is neither a bare value nor a verdict: %v\n%s", err, stdout)
	}
	return verdict.Value
}

// readyAgentQueue reads a `swarm ready --agent` answer, at either exit
// status: the clear shape carries each property's bare value, and the
// attention shape carries {ok, value, issues} per property — both are
// read here so a test asserts against the same fields either way.
type readyAgentQueue struct {
	Established bool
	Drained     bool
	IDs         []string
	States      map[string]string
}

func readAgentQueue(t *testing.T, stdout string) readyAgentQueue {
	t.Helper()

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &fields); err != nil {
		t.Fatalf("the ready --agent answer does not parse: %v\n%s", err, stdout)
	}

	out := readyAgentQueue{States: map[string]string{}}
	out.Established = readAgentFlag(t, fields, "established", stdout)
	out.Drained = readAgentFlag(t, fields, "drained", stdout)

	raw, ok := fields["queue"]
	if !ok {
		t.Fatalf("the answer carries no queue:\n%s", stdout)
	}
	var entries []struct {
		ID    string `json:"id"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(raw, &entries); err != nil {
		// The attention shape wraps it as {value: [...]}.
		var verdict struct {
			Value []struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"value"`
		}
		if err := json.Unmarshal(raw, &verdict); err != nil {
			t.Fatalf("queue is neither a bare list nor a verdict: %v\n%s", err, stdout)
		}
		entries = verdict.Value
	}
	for _, e := range entries {
		out.IDs = append(out.IDs, e.ID)
		out.States[e.ID] = e.State
	}
	return out
}

func readAgentFlag(t *testing.T, fields map[string]json.RawMessage, name, stdout string) bool {
	t.Helper()

	raw, ok := fields[name]
	if !ok {
		t.Fatalf("the answer carries no %s:\n%s", name, stdout)
	}
	var direct bool
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct
	}
	var verdict struct {
		Value bool `json:"value"`
	}
	if err := json.Unmarshal(raw, &verdict); err != nil {
		t.Fatalf("%s is neither a bare flag nor a verdict: %v\n%s", name, err, stdout)
	}
	return verdict.Value
}

// AC: AC-SWARM-DISP-002.12
//
// A dispatch holding three beads in a collision chain: only the head of
// the chain is ready, since a bead blocked by an unfinished dependency
// never reaches the ready set at all — which is what "in an order their
// dependencies allow" comes down to. Neither blocked member of the chain
// is reported.
func TestReadyAgentReturnsOnlyItsBeadsInDependencyOrder(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	a := w.createID(t, "bd", "create", "chain a", "--type", "task",
		"--parent", "fixture-run.target", "--assignee", "dev-chain", "--json")
	b := w.createID(t, "bd", "create", "chain b", "--type", "task",
		"--parent", "fixture-run.target", "--assignee", "dev-chain", "--json")
	c := w.createID(t, "bd", "create", "chain c", "--type", "task",
		"--parent", "fixture-run.target", "--assignee", "dev-chain", "--json")
	w.mustRun(t, w.repo, "bd", "dep", "add", b, a, "--type", "blocks", "--json")
	w.mustRun(t, w.repo, "bd", "dep", "add", c, b, "--type", "blocks", "--json")

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--agent", "dev-chain", "--json")
	if code != 0 {
		t.Fatalf("ready --agent exited %d, want 0\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	queue := readAgentQueue(t, stdout)
	if !queue.Established {
		t.Errorf("dev-chain is not reported established, though it holds three beads:\n%s", stdout)
	}
	if queue.Drained {
		t.Errorf("dev-chain is reported drained, though its head bead is ready:\n%s", stdout)
	}
	if strings.Join(queue.IDs, ",") != a {
		t.Errorf("queue = %v, want only %s — %s and %s are still blocked", queue.IDs, a, b, c)
	}
}

// AC: AC-SWARM-DISP-002.13
//
// An agent holding one bead in progress and none ready — the core
// fixture's own dev-fixture-1, which holds fixture-run.2 in progress and
// fixture-run.1 closed — is reported as still holding work rather than
// drained. The ready half alone would report it finished.
func TestReadyAgentMidBeadIsNotDrained(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--agent", "dev-fixture-1", "--json")
	if code != 0 {
		t.Fatalf("ready --agent exited %d, want 0\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	queue := readAgentQueue(t, stdout)
	if !queue.Established {
		t.Errorf("dev-fixture-1 is not reported established:\n%s", stdout)
	}
	if queue.Drained {
		t.Errorf("dev-fixture-1 is reported drained though it holds fixture-run.2 in progress:\n%s", stdout)
	}
	if queue.States["fixture-run.2"] != "in_progress" {
		t.Errorf("fixture-run.2 is not reported in_progress: %v\n%s", queue.States, stdout)
	}
	if _, ready := queue.States["fixture-run.1"]; ready {
		t.Errorf("fixture-run.1 (closed) reached the queue:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-002.14
//
// The same dispatch id holding a bead in this run and a bead in another
// run: the foreign bead never reaches the queue, even though the held
// half's own query is not scoped by run and would otherwise return it.
func TestReadyAgentQueueNeverReachesOutsideItsRun(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	foreignRun := w.createID(t, "bd", "create", "A Second Run", "--type", "epic", "--json")
	own := w.createID(t, "bd", "create", "in this run", "--type", "task",
		"--parent", "fixture-run.target", "--assignee", "dev-cross-run", "--json")
	foreign := w.createID(t, "bd", "create", "in the other run", "--type", "task",
		"--parent", foreignRun, "--assignee", "dev-cross-run", "--json")

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--agent", "dev-cross-run", "--json")
	if code != 0 {
		t.Fatalf("ready --agent exited %d, want 0\n--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	queue := readAgentQueue(t, stdout)
	if !queue.Established {
		t.Errorf("dev-cross-run is not reported established, though it holds a bead in this run:\n%s", stdout)
	}
	if strings.Join(queue.IDs, ",") != own {
		t.Errorf("queue = %v, want only %s — %s belongs to %s", queue.IDs, own, foreign, foreignRun)
	}
}

// AC: AC-SWARM-DISP-002.15
//
// A dispatch id with no bead assigned to it anywhere in the run is
// reported not established rather than drained, and the command exits
// with the attention status — a dispatch error, not a completion.
func TestReadyAgentNeverGivenAnythingIsNotReportedAsDrained(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--agent", "dev-never-given", "--json")
	if code != 1 {
		t.Fatalf("ready --agent exited %d, want 1 (attention)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}

	queue := readAgentQueue(t, stdout)
	if queue.Established {
		t.Errorf("dev-never-given is reported established, though no bead names it:\n%s", stdout)
	}
	if !strings.Contains(stdout, "not established") && !strings.Contains(stdout, "dev-never-given") {
		t.Errorf("the answer does not explain that the dispatch was never given anything:\n%s", stdout)
	}
}

// AC: AC-SWARM-DISP-002.11
//
// A resumed run's recorded failure count survives a bare re-query: two
// consecutive dispatches completing nothing were recorded before this
// session started, and querying the ready set with no --completed reads
// that count back rather than resetting it to zero.
func TestReadyBreakerCountSurvivesAResumedRun(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	setOut, setErr, setCode := w.swarm(t, "manifest", "set", "--run", "fixture-run",
		"--key", "consecutive_dispatch_failures", "--value", "2", "--json")
	if setCode != 0 {
		t.Fatalf("manifest set exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", setCode, setOut, setErr)
	}

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--json")
	if code != 0 {
		t.Fatalf("ready exited %d, want 0 (a count of 2 does not yet meet the default threshold of 3)\n"+
			"--- stdout ---\n%s\n--- stderr ---\n%s", code, stdout, stderr)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &fields); err != nil {
		t.Fatalf("the ready answer does not parse: %v\n%s", err, stdout)
	}
	raw, ok := fields["consecutive_dispatch_failures"]
	if !ok {
		t.Fatalf("the answer carries no consecutive_dispatch_failures:\n%s", stdout)
	}
	var count int
	if err := json.Unmarshal(raw, &count); err != nil {
		t.Fatalf("consecutive_dispatch_failures does not parse as a number: %v\n%s", err, stdout)
	}
	if count != 2 {
		t.Errorf("consecutive_dispatch_failures = %d, want 2 (the resumed run's own recorded count)", count)
	}
}

// The full CLI wiring: --completed 0 reports a dispatch return that
// completed nothing, against a run already two returns into the default
// threshold of three — the third return trips the breaker, the command
// exits with the stop status, and the write lands in the tracker rather
// than only in this invocation's own answer.
func TestReadyBreakerStopsTheRunThroughTheRealCommand(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	w := restore(t, binary, "core")

	setOut, setErr, setCode := w.swarm(t, "manifest", "set", "--run", "fixture-run",
		"--key", "consecutive_dispatch_failures", "--value", "2", "--json")
	if setCode != 0 {
		t.Fatalf("manifest set exited %d\n--- stdout ---\n%s\n--- stderr ---\n%s", setCode, setOut, setErr)
	}

	stdout, stderr, code := w.swarm(t, "ready", "--run", "fixture-run", "--completed", "0", "--json")
	if code != 2 {
		t.Fatalf("ready --completed 0 exited %d, want 2 (stop)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, stdout, stderr)
	}

	shown := w.mustCapture(t, w.repo, "bd", "show", "fixture-run", "--json")
	var issues []struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(shown), &issues); err != nil || len(issues) != 1 {
		t.Fatalf("bd show fixture-run did not answer one issue: %v\n%s", err, shown)
	}
	got, ok := issues[0].Metadata["consecutive_dispatch_failures"]
	if !ok {
		t.Fatalf("the tracker records no consecutive_dispatch_failures after the write: %v", issues[0].Metadata)
	}
	if n, ok := got.(float64); !ok || n != 3 {
		t.Errorf("consecutive_dispatch_failures = %v, want 3 (2 already recorded, advanced by this return)", got)
	}
}
