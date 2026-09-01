// Tests for the Status Reporter: the wording it gives and the outcome it
// derives, against a fixed measurement set with no subprocess anywhere in
// the test. Separating [reportStatus] from [status.Collect] is what
// makes that possible — per the Status Reporter component in
// specs/architecture/container-swarm-cli.md.
package cli

import (
	"strings"
	"testing"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

func established() status.Unestablished { return status.Unestablished{Established: true} }

// clearMeasurements is a run reading clear: one worktree standing (the
// main checkout), no dispatch worktree, a fresh commit, an empty ready
// ready set and a fully closed, unnarrowed scope.
func clearMeasurements() status.Measurements {
	return status.Measurements{
		Run:        "run-1",
		Worktrees:  status.Worktrees{Unestablished: established(), Standing: 1, Trees: []vcs.Worktree{{Path: "/w/0", Branch: "main"}}},
		LastCommit: status.LastCommit{Unestablished: established(), SHA: "abc1234", Subject: "run-1.1 P1 done", When: time.Now().Add(-time.Minute), Age: time.Minute},
		Ready:      status.ReadySet{Unestablished: established(), Count: 0},
		Completion: status.Completion{Unestablished: established(), Total: 1, Done: 1},
		Scope:      status.Scope{Unestablished: established()},
	}
}

func req() Request { return Request{Command: "swarm status", Run: "run-1"} }

// sampleRecords is the set of invocation records a real [status.Collect]
// call produces for [clearMeasurements] — one per step it names, each
// having actually "run". Built by hand rather than run for real, because
// this reporter is tested with no subprocess anywhere.
func sampleRecords() []plan.Record {
	ran := func(step string, argv ...string) plan.Record {
		return plan.Record{Step: step, Status: plan.StepRan,
			Invocation: swarmexec.Invocation{Name: argv[0], Args: argv[1:]}}
	}
	return []plan.Record{
		ran("manifest", "bd", "show", "run-1", "--json"),
		ran("worktrees-prune", "git", "worktree", "prune"),
		ran("worktrees-list", "git", "worktree", "list", "--porcelain"),
		ran("last-commit", "git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", "sprint/run-1"),
		ran("ready", "bd", "ready", "--parent", "run-1", "--exclude-type", "epic", "--json"),
		ran("children-0-0", "bd", "list", "--parent", "run-1", "--all", "--limit", "0", "--json"),
	}
}

// AC: AC-SWARM-CORE-003.5
//
// A tracker query that failed is reported as unknown, on the one line it
// affected, and the command exits non-zero — whatever the rest of the
// run's live state would otherwise report.
func TestATrackerQueryThatCannotAnswerReportsUnknownAndExitsNonZero(t *testing.T) {
	m := clearMeasurements()
	m.Ready = status.ReadySet{Unestablished: status.Unestablished{Reason: "bd: database is locked"}}

	result := reportStatus(req(), m, nil, 30*time.Minute)

	if result.Outcome.Status() == StatusClear {
		t.Fatalf("outcome = %q, want a non-zero exit for a measurement that could not be taken", result.Outcome)
	}

	var found bool
	for _, f := range result.Facts {
		if f.Name != "ready_count" {
			continue
		}
		found = true
		if f.Established {
			t.Errorf("ready_count reads established though its query failed")
		}
		if !strings.Contains(f.Reason, "database is locked") {
			t.Errorf("reason = %q, want it to carry the query's own failure", f.Reason)
		}
	}
	if !found {
		t.Fatal("no fact named ready_count was reported")
	}

	// A completion line the query did not touch is unaffected.
	for _, f := range result.Facts {
		if f.Name == "completion_total" && !f.Established {
			t.Error("a measurement the failing query did not touch was disturbed")
		}
	}
}

// AC: AC-SWARM-CORE-003.10
//
// A manifest recording no admitted list is a legitimate state, not a
// failure, and the scope comparison is still reported as unestablished —
// never as "no additions" — and the command still exits non-zero.
func TestAnAbsentAdmittedListExitsNonZero(t *testing.T) {
	m := clearMeasurements()
	m.Scope = status.Scope{Unestablished: status.Unestablished{
		Reason: "the manifest records no admitted list, so there is nothing to compare the live queue against",
	}}

	result := reportStatus(req(), m, nil, 30*time.Minute)

	if result.Outcome.Status() == StatusClear {
		t.Fatalf("outcome = %q, want a non-zero exit over an unestablished scope comparison", result.Outcome)
	}

	var found bool
	for _, f := range result.Facts {
		if f.Name != "scope_additions" {
			continue
		}
		found = true
		if f.Established {
			t.Error("scope_additions reads established though the manifest records no admitted list")
		}
		if !strings.Contains(f.Reason, "admitted") {
			t.Errorf("reason = %q, want it to name the key it could not read", f.Reason)
		}
	}
	if !found {
		t.Fatal("no fact named scope_additions was reported")
	}
}

// AC: AC-SWARM-CORE-003.6
//
// The report states, on every call, that whether a subagent is still
// working cannot be observed from outside the harness, and names the
// three measurements it gives in that observation's place.
func TestTheReportNamesWhatItCannotObserveAndWhatItGivesInstead(t *testing.T) {
	result := reportStatus(req(), clearMeasurements(), nil, 30*time.Minute)

	var text string
	for _, f := range result.Facts {
		if f.Name == "observability" {
			text = *f.Value.Text
		}
	}
	if text == "" {
		t.Fatal("no fact named observability was reported")
	}
	if !strings.Contains(text, "cannot be observed") {
		t.Errorf("observability = %q, want it to say the state cannot be observed", text)
	}
	for _, want := range []string{"worktree", "commit", "ready"} {
		if !strings.Contains(text, want) {
			t.Errorf("observability = %q, want it to name the %q measurement given in place of the observation", text, want)
		}
	}
}

// AC: AC-SWARM-CORE-003.1, AC-SWARM-CORE-003.11
//
// No dispatch worktree stands, the ready set is not empty, and the last
// commit is older than the threshold in force: the report states the
// measurements and the threshold, exits with the attention status, and
// never says the run has stalled — the word appears nowhere in what it
// renders.
func TestNoProgressIsReportedAsMeasurementsNeverAsAStall(t *testing.T) {
	m := clearMeasurements()
	m.Worktrees = status.Worktrees{Unestablished: established(), Standing: 1,
		Trees: []vcs.Worktree{{Path: "/w/0", Branch: "main"}}} // no Dispatch was attributed this tree: no dispatch worktree stands
	m.Ready = status.ReadySet{Unestablished: established(), Count: 4}
	m.LastCommit = status.LastCommit{Unestablished: established(),
		SHA: "abc1234", Subject: "run-1.1 P1 done",
		When: time.Now().Add(-3*time.Hour - 12*time.Minute), Age: 3*time.Hour + 12*time.Minute}

	result := reportStatus(req(), m, nil, 30*time.Minute)

	if result.Outcome != OutcomeAttention {
		t.Fatalf("outcome = %q, want %q", result.Outcome, OutcomeAttention)
	}
	if strings.TrimSpace(result.Resolution) == "" {
		t.Fatal("an attention outcome names no resolution")
	}

	rendered := strings.ToLower(result.Summary + " " + result.Resolution)
	for _, f := range result.Facts {
		if f.Value != nil && f.Value.Text != nil {
			rendered += " " + strings.ToLower(*f.Value.Text)
		}
	}
	if strings.Contains(rendered, "stall") {
		t.Errorf("the report asserts a stall:\n%s", rendered)
	}

	for _, want := range []string{"4", "30m0s"} {
		if !strings.Contains(result.Summary, want) {
			t.Errorf("summary = %q, want it to contain %q", result.Summary, want)
		}
	}

	var threshold string
	for _, f := range result.Facts {
		if f.Name == "stall_after" {
			threshold = *f.Value.Text
		}
	}
	if threshold != "30m0s" {
		t.Errorf("stall_after = %q, want the threshold in force stated whether given or defaulted", threshold)
	}
}

// A run whose last commit is recent, or whose ready set is empty, or
// which still has a dispatch worktree standing, reads clear: the attention
// condition needs all three at once.
func TestProgressIsClearUnlessAllThreeConditionsHold(t *testing.T) {
	cases := []struct {
		name string
		edit func(*status.Measurements)
	}{
		{"a dispatch worktree stands", func(m *status.Measurements) {
			m.Ready.Count = 4
			m.LastCommit.Age = 3 * time.Hour
			m.Dispatches = append(m.Dispatches, status.Dispatch{
				Unestablished: established(), Agent: "dev-1", Drained: true, Worktree: "/w/1",
			})
		}},
		{"the ready set is empty", func(m *status.Measurements) {
			m.Ready.Count = 0
			m.LastCommit.Age = 3 * time.Hour
		}},
		{"the commit is recent", func(m *status.Measurements) {
			m.Ready.Count = 4
			m.LastCommit.Age = time.Minute
		}},
	}

	for _, c := range cases {
		m := clearMeasurements()
		c.edit(&m)
		result := reportStatus(req(), m, nil, 30*time.Minute)
		if result.Outcome != OutcomeClear {
			t.Errorf("%s: outcome = %q, want %q", c.name, result.Outcome, OutcomeClear)
		}
	}
}

// The threshold given on the command line is what the report states and
// compares against, not the default.
func TestAGivenThresholdIsTheOneStatedAndCompared(t *testing.T) {
	m := clearMeasurements()
	m.Ready.Count = 1
	m.LastCommit.Age = 2 * time.Hour

	clear := reportStatus(req(), m, nil, 3*time.Hour)
	if clear.Outcome != OutcomeClear {
		t.Errorf("2h old with a 3h threshold: outcome = %q, want %q", clear.Outcome, OutcomeClear)
	}

	attention := reportStatus(req(), m, nil, time.Hour)
	if attention.Outcome != OutcomeAttention {
		t.Errorf("2h old with a 1h threshold: outcome = %q, want %q", attention.Outcome, OutcomeAttention)
	}
	var threshold string
	for _, f := range attention.Facts {
		if f.Name == "stall_after" {
			threshold = *f.Value.Text
		}
	}
	if threshold != "1h0m0s" {
		t.Errorf("stall_after = %q, want the given threshold", threshold)
	}
}

// dispatchWorktreeCount reads the collector's own attribution — each
// dispatch's Worktree field — rather than the raw worktree count: a real
// repository's main checkout is always a standing worktree that no
// dispatch owns, so the raw count is never the right measure.
func TestDispatchWorktreeCountExcludesTheMainCheckout(t *testing.T) {
	dispatches := []status.Dispatch{
		{Agent: "dev-1", Worktree: ""},
		{Agent: "dev-2", Worktree: ""},
	}
	if n := dispatchWorktreeCount(dispatches); n != 0 {
		t.Errorf("dispatchWorktreeCount = %d, want 0 — neither dispatch was attributed a worktree", n)
	}

	dispatches = append(dispatches, status.Dispatch{Agent: "dev-3", Worktree: "/w/2"})
	if n := dispatchWorktreeCount(dispatches); n != 1 {
		t.Errorf("dispatchWorktreeCount = %d, want 1", n)
	}
}

// Every fact this reporter builds carries an invocation, whatever
// measurement it reports on — the manifest read is the fallback so this
// holds even for a fact with no measurement-specific invocation of its
// own, such as the threshold or a wholly unestablished line.
func TestEveryFactCarriesAnInvocation(t *testing.T) {
	m := clearMeasurements()
	m.Scope = status.Scope{Unestablished: status.Unestablished{Reason: "the manifest records no admitted list"}}

	result := reportStatus(req(), m, sampleRecords(), 30*time.Minute)
	if err := result.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}
