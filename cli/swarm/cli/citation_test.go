// Unit tests for the guarantee every fact in this package carries: at
// least one non-empty invocation a reader can re-run.
//
// [Result.Validate] refuses a fact naming no invocation, and [finish]
// replaces the WHOLE result with a bare refusal when it does — so one
// unciteable fact costs every other fact beside it. These tests pin the
// five sites that built a fact from a record set nothing guarantees is
// citable, on the state their producers answer nothing for: a record
// whose Invocation.Name is empty, which is what [argvOf] returns nil for
// and what a step the runner refused or skipped produces.
package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"encoding/json"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
)

// uncitedRun is the run the guard test below scopes to.
const uncitedRun = "run-uncited"

// Why sites 1, 2 and 3 are covered at their guard and not end to end.
//
// Their records come back from [tracker.Descendants] and
// [tracker.AgentQueue], and NEITHER can answer successfully with records
// nothing cites. Three independent guards stand in the way, measured:
//
//   - A record that did not run carries a nil value, and BeadsFrom,
//     HeldFrom and AgentQueue's own answered map each refuse one — so a
//     skipped or refused record makes the helper return an error, and the
//     three sites return on it before the citation is built.
//   - Both helpers hardcode [plan.ModeExecute], so a dry run cannot
//     refuse their reads into the never-ran state either.
//   - A step declaring no executable — the other arrangement in which a
//     record that RAN carries an empty Invocation.Name, which is what
//     argvOf answers nil for — never runs at all: [plan.Plan.Validate]
//     refuses it with `plan: step %q has no executable`.
//
// So the never-ran-record path at those three is a defence rather than a
// reachable failure, and the guard it is made of is what these tests
// pin. Sites 4 and 5 are different: their steps ARE in scope, a dry run
// reaches their never-ran state, and they are covered end to end below.

// wantsStop asserts err is the unrepairable refusal, and that its reason
// says what could not be cited rather than reporting an answer.
func wantsStop(t *testing.T, err error, subject string) {
	t.Helper()
	if err == nil {
		t.Fatalf("no refusal: a read nothing cites was reported as an answer")
	}
	var refusal *Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %T (%v), want a *Refusal", err, err)
	}
	if refusal.Outcome != OutcomeStop {
		t.Errorf("outcome = %q, want %q — the caller cannot repair a read that never reached its answer",
			refusal.Outcome, OutcomeStop)
	}
	if !strings.Contains(refusal.Reason, subject) {
		t.Errorf("reason = %q, want it to name %q — the refusal has to say what could not be cited",
			refusal.Reason, subject)
	}
	if !strings.Contains(refusal.Reason, "no invocation") {
		t.Errorf("reason = %q, want it to say no invocation could be re-run", refusal.Reason)
	}
}

// The guard sites 1, 2 and 3 share. A record set nothing cites is refused
// rather than reported, and one that cites something is passed through
// whole.
func TestInvocationsRequiredRefusesAReadNoRecordCites(t *testing.T) {
	neverRan := []plan.Record{
		{Step: "children-0-0", Status: plan.StepSkipped},
		{Step: "children-0-1", Status: plan.StepRefused},
	}
	for _, rec := range neverRan {
		if argv := argvOf(rec); argv != nil {
			t.Fatalf("argvOf(%q) = %v, want nil — the scenario needs records nothing cites", rec.Step, argv)
		}
	}

	if _, err := invocationsRequired("the beads under "+uncitedRun, neverRan); err == nil {
		t.Errorf("invocationsRequired accepted a record set nothing cites")
	} else {
		wantsStop(t, err, "the beads under "+uncitedRun)
	}

	if _, err := invocationsRequired("an empty read", nil); err == nil {
		t.Errorf("invocationsRequired accepted an empty record set")
	}

	ran := append(neverRan, plan.Record{
		Step:       "children-0-2",
		Status:     plan.StepRan,
		Invocation: swarmexec.Invocation{Name: "bd", Args: []string{"list", "--json"}},
	})
	inv, err := invocationsRequired("the beads under "+uncitedRun, ran)
	if err != nil {
		t.Fatalf("invocationsRequired refused a record set one record cites: %v", err)
	}
	if len(inv) != 1 || len(inv[0]) == 0 {
		t.Errorf("invocations = %v, want the one record that ran", inv)
	}
}

// Sites 4 and 5, and the substitution that repairs them. A record that
// never ran carries no invocation of its own, and where the STEP is in
// scope the argv it would have run with is a citation a reader can
// re-run — which is what [invocationOf] answers and [argvOf] does not.
func TestInvocationOfCitesAStepThatNeverRan(t *testing.T) {
	step := beadstracker.Tracker{}.SetState("close-0", "run-1.a", tracker.StateDone)
	for _, rec := range []plan.Record{
		{Step: "close-0", Status: plan.StepRefused, Reason: "dry run: the step writes"},
		{Step: "close-0", Status: plan.StepSkipped, Reason: "an earlier step stopped the plan"},
	} {
		if argvOf(rec) != nil {
			t.Fatalf("argvOf answered for a %q record; the scenario needs one nothing cites", rec.Status)
		}
		argv := invocationOf(step, rec)
		if len(argv) == 0 {
			t.Fatalf("invocationOf(%q) = %v, want the argv the step would have run with", rec.Status, argv)
		}
		if argv[0] != "bd" || argv[1] != "update" || argv[2] != "run-1.a" {
			t.Errorf("invocationOf(%q) = %v, want the step's own declared invocation", rec.Status, argv)
		}
	}
}

// Sites 4 and 5 end to end. A dry run refuses every close, so each
// `close-N` record never ran — the state [closeEligible] used to skip
// without citing. Every fact the command reports still names an
// invocation a reader can re-run, and the whole result validates.
func TestPhaseCloseCitesEveryFactWhenNoCloseRan(t *testing.T) {
	tr := beadstracker.Tracker{}
	fake := exectest.NewFake(
		children("run-1", "["+epicRow("run-1.a")+","+epicRow("run-1.b")+"]"),
		children("run-1.a", "[]"),
		children("run-1.b", "[]"),
		readyAt("run-1", "[]"),
		closable(`[{"epic":`+epicRow("run-1.a")+`,"total_children":1,"closed_children":1,"eligible_for_close":true}]`),
		readyAt("run-1", "[]"),
	)

	req := Request{Command: "swarm phase close", Run: "run-1", Mode: plan.ModeDryRun}
	result, err := closePhases(context.Background(), tr, fake, req)
	if err != nil {
		t.Fatalf("closePhases: %v", err)
	}

	// The state the assertions are about: the close the eligible epic
	// would have taken was refused, so its record never ran.
	var closed plan.Record
	for _, rec := range result.Records {
		if rec.Step == "close-0" {
			closed = rec
		}
	}
	if closed.Status != plan.StepRefused {
		t.Fatalf("close-0 is %q, want %q — the scenario did not arise", closed.Status, plan.StepRefused)
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("the result does not validate, so finish replaces every fact in it with a bare refusal: %v", err)
	}
	for _, f := range result.Facts {
		if len(f.Invocations) == 0 {
			t.Errorf("fact %q names no invocation that established it", f.Name)
		}
		for i, argv := range f.Invocations {
			if len(argv) == 0 {
				t.Errorf("fact %q's invocation %d is empty", f.Name, i)
			}
		}
	}

	// closed_epics is site 4's own fact: the epic it would close comes
	// from the eligibility READ, and the citation now carries the write
	// each refused close would have made beside it.
	found := closePhaseFactNamed(t, result, "closed_epics")
	if got := found.Value.Items; got == nil || len(*got) != 1 || (*got)[0] != "run-1.a" {
		t.Errorf("closed_epics = %v, want [run-1.a]", got)
	}
	var citesTheClose bool
	for _, argv := range found.Invocations {
		if strings.Join(argv, " ") == "bd update run-1.a --status closed --json" {
			citesTheClose = true
		}
	}
	if !citesTheClose {
		t.Errorf("closed_epics cites %v, want the close it would have run among them", found.Invocations)
	}
}

// citesEveryFact asserts every fact in a JSON envelope names at least one
// non-empty invocation. It reads the emitted contract rather than the
// in-process Result, so a fact that reached a consumer uncited fails here
// even if the value it carried was right.
func citesEveryFact(t *testing.T, what, stdout string) {
	t.Helper()
	var env swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("%s: the result does not decode: %v\n%s", what, err, stdout)
	}
	if len(env.Facts) == 0 {
		t.Fatalf("%s reported no fact at all, so nothing was asserted:\n%s", what, stdout)
	}
	for _, f := range env.Facts {
		if len(f.Invocations) == 0 {
			t.Errorf("%s: fact %q names no invocation that established it", what, f.Name)
		}
		for i, cmd := range f.Invocations {
			if strings.TrimSpace(cmd) == "" {
				t.Errorf("%s: fact %q's invocation %d is empty", what, f.Name, i)
			}
		}
	}
}

// Sites 1 and 2 on the path that IS reachable. Both halves of `swarm
// dispatch accept` are driven — the drain check whose queueInv site 1
// cites, and the post-merge re-check whose beadsInv site 2 cites — and
// every fact either half reports has to name a command a reader can
// re-run.
func TestAcceptCitesEveryFactItReports(t *testing.T) {
	world := cleanSettle()

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(world)...)

	fake := exectest.NewFake(recs...)
	stdout, _, _ := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json", "--full-result")
	citesEveryFact(t, "swarm dispatch accept", stdout)
}

// Site 1's other arm. An agent still holding work reports three facts —
// drained, holding and holding_count — all cited from the same queue
// read, and nothing below the drain check runs to supply a citation if
// that one were empty.
func TestAcceptCitesEveryFactWhenTheDispatchHasNotDrained(t *testing.T) {
	fake := exectest.NewFake(drainedQueue("run-1.2")...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json", "--full-result")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	citesEveryFact(t, "swarm dispatch accept, not drained", stdout)
}

// Site 3. The four facts `swarm ready --agent` reports — established,
// queue, queue_count and drained — are cited from two tracker reads, and
// every one of them has to name a command a reader can re-run.
func TestReadyAgentCitesEveryFactItReports(t *testing.T) {
	_, stdout, code := runReadyAgent(t, []exectest.Recording{
		showRecording(acceptRun, `{"consecutive_dispatch_failures":0}`),
	}, "ready", "--run", acceptRun, "--agent", acceptAgent, "--full-result", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	citesEveryFact(t, "swarm ready --agent", stdout)
}
