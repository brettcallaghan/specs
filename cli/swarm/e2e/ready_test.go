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
