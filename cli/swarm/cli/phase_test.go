// Unit tests for `swarm phase close`: the run-scope refusal, a dry run
// closing nothing, and a failed eligibility query reading as unanswered
// rather than as nothing eligible. Every scenario runs through the Beads
// Tracker's real argv-building against the fake process runner, so no
// test starts `bd`.
//
// AC-SWARM-DISP-001.1, .2, .4, .6, .7 and .8 are declared e2e and live in
// cli/swarm/e2e against the phase-epics and core fixtures.
package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
)

// children records `bd list --parent <parent> --all --limit 0 --json`.
func children(parent string, rows string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "list", "--parent", parent, "--all", "--limit", "0", "--json"},
		Stdout: []byte(rows),
	}
}

// readyAt records `bd ready --parent <run> --exclude-type epic --limit 0
// --json`, twice-matchable: the fake replays recordings sharing this
// argv in the order they were declared, so a before/after pair is two
// calls of this helper with different rows.
func readyAt(run, rows string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "ready", "--parent", run, "--exclude-type", "epic", "--limit", "0", "--json"},
		Stdout: []byte(rows),
	}
}

func closable(rows string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "epic", "status", "--eligible-only", "--json"},
		Stdout: []byte(rows),
	}
}

func setClosed(id string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "update", id, "--status", "closed", "--json"},
		Stdout: []byte(`[{"id":"` + id + `","status":"closed","issue_type":"epic"}]`),
	}
}

func epicRow(id string) string {
	return `{"id":"` + id + `","status":"open","issue_type":"epic"}`
}

// AC: AC-SWARM-DISP-001.5
//
// An unscoped invocation is refused with the malformed-invocation status,
// and nothing is read or written — asserted from the fake's own record,
// so a surface that checked the scope after running something would fail
// here even though its status was right.
func TestPhaseCloseRefusesAnEmptyScope(t *testing.T) {
	for _, args := range [][]string{
		{"phase", "close"},
		{"phase", "close", "--run", ""},
		{"phase", "close", "--run", "   "},
	} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		tr := beadstracker.Tracker{}
		streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
		root := newRootCmd(streams)
		root.AddCommand(phaseCmd(streams, tr, fake))

		code := dispatch(root, streams, args, &bytes.Buffer{}, &bytes.Buffer{})
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read or wrote something outside the scope: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-DISP-001.3
//
// A dry run performs every read and refuses the close — [plan.Runner]'s
// own guarantee — and the epic it would close is still reported, because
// that answer comes from the eligibility READ, never from the write.
func TestPhaseCloseDryRunPerformsTheReadsAndClosesNothing(t *testing.T) {
	tr := beadstracker.Tracker{}
	fake := exectest.NewFake(
		children("run-1", "["+epicRow("run-1.a")+"]"),
		children("run-1.a", "[]"),
		readyAt("run-1", "[]"),
		closable(`[{"epic":`+epicRow("run-1.a")+`,"total_children":1,"closed_children":1,"eligible_for_close":true}]`),
		readyAt("run-1", "[]"),
		// Deliberately no recording for `bd update run-1.a --status
		// closed --json`: a dry run must never reach it, and the fake
		// errors any invocation it was not told about.
	)

	req := Request{Command: "swarm phase close", Run: "run-1", Mode: plan.ModeDryRun}
	result, err := closePhases(context.Background(), tr, fake, req)
	if err != nil {
		t.Fatalf("closePhases: %v", err)
	}

	if wrote := writeCallsMade(fake); len(wrote) != 0 {
		t.Errorf("a dry run wrote: %v", wrote)
	}

	found := closePhaseFactNamed(t, result, "closed_epics")
	if got := found.Value.Items; got == nil || len(*got) != 1 || (*got)[0] != "run-1.a" {
		t.Errorf("closed_epics = %v, want [run-1.a] — the dry run's own eligibility read", got)
	}
	if !strings.Contains(result.Summary, "would close") {
		t.Errorf("summary = %q, want it to say what would close", result.Summary)
	}
	if strings.Contains(result.Summary, "released") {
		t.Errorf("summary = %q, a dry run wrote nothing so nothing was released", result.Summary)
	}
}

// AC: AC-SWARM-DISP-001.9
//
// A tracker that cannot answer the eligibility query is reported as
// unanswered, never as nothing eligible, and the command takes the stop
// status.
func TestClosePhasesReportsAFailedEligibilityQueryAsUnanswered(t *testing.T) {
	tr := beadstracker.Tracker{}
	fake := exectest.NewFake(
		children("run-1", "[]"),
		readyAt("run-1", "[]"),
		exectest.Recording{
			Argv:       []string{"bd", "epic", "status", "--eligible-only", "--json"},
			ExitStatus: 1,
			Stderr:     []byte("bd: database is locked"),
		},
	)

	req := Request{Command: "swarm phase close", Run: "run-1", Mode: plan.ModeExecute}
	_, err := closePhases(context.Background(), tr, fake, req)
	if err == nil {
		t.Fatal("closePhases: want an error when the eligibility query fails")
	}

	refusal := asRefusal(err)
	if refusal.Outcome != OutcomeStop {
		t.Errorf("outcome = %q, want %q", refusal.Outcome, OutcomeStop)
	}
	if strings.Contains(refusal.Reason, "nothing eligible") || strings.Contains(refusal.Reason, "no epic") {
		t.Errorf("a failed query must not be reported as nothing eligible: %q", refusal.Reason)
	}
	if !strings.Contains(refusal.Reason, "unanswered") {
		t.Errorf("reason = %q, want it to say the query is unanswered", refusal.Reason)
	}
}

// writeCallsMade is every `bd update` the fake actually saw — the write
// half of the invocation record, read from the fake rather than from the
// result, so a dry run that fabricated a clean-looking result would still
// be caught here.
func writeCallsMade(fake *exectest.Fake) [][]string {
	var out [][]string
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update" {
			out = append(out, argv)
		}
	}
	return out
}

// closePhaseFactNamed reads one fact out of a [Result] by name, failing
// the test if it is absent. Named apart from manifest_test.go's own
// factNamed, which reads the rendered [swarmmodel.CommandResult] rather
// than the pre-render [Result] this package's own tests build.
func closePhaseFactNamed(t *testing.T, result Result, name string) Fact {
	t.Helper()
	for _, f := range result.Facts {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("the result carries no fact named %q", name)
	return Fact{}
}
