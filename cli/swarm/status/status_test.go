// Tests for the Status Collector: what it measures, what it walks, and
// what it refuses to report as a number.
//
// No test starts a process. The collector reaches `bd` and `git` through
// the Beads Tracker and the Git Version Control, whose steps run through
// the fake process runner, so every scenario below is a set of recorded
// outputs rather than a repository and a database.
package status_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

const (
	runEpic = "demo-run1"
	branch  = "sprint/run1"
)

// now is the instant every age in these tests is taken against, fixed so
// an age is a value rather than a moving target.
var now = time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

func deps(recordings ...exectest.Recording) (status.Deps, *exectest.Fake) {
	fake := exectest.NewFake(recordings...)
	tr := beadstracker.Tracker{}
	return status.Deps{
		Tracker: tr,
		VCS:     gitvcs.VersionControl{},
		Store:   manifestbeads.Store{Tracker: tr},
		Runner:  plan.Runner{Exec: fake},
		Now:     func() time.Time { return now },
	}, fake
}

func collect(t *testing.T, recordings ...exectest.Recording) status.Measurements {
	t.Helper()
	d, _ := deps(recordings...)
	m, _, err := status.Collect(context.Background(), d, runEpic)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	return m
}

// --- recordings -----------------------------------------------------

// show records `bd show <epic> --json`, which is how the manifest is read.
func show(metadata string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "show", runEpic, "--json"},
		Stdout: []byte(`[{"id":"` + runEpic + `","status":"open","issue_type":"epic","metadata":` + metadata + `}]`),
	}
}

// manifestWith is the smallest record that names a branch and an
// admission list.
func manifestWith(admitted ...string) string {
	quoted := make([]string, len(admitted))
	for i, id := range admitted {
		quoted[i] = `"` + id + `"`
	}
	return `{"branch":"` + branch + `","admitted":[` + strings.Join(quoted, ",") + `]}`
}

type bead struct {
	id       string
	status   string
	kind     string
	assignee string
}

func issues(beads ...bead) []byte {
	rows := make([]string, len(beads))
	for i, b := range beads {
		kind := b.kind
		if kind == "" {
			kind = "task"
		}
		rows[i] = fmt.Sprintf(
			`{"id":%q,"title":%q,"status":%q,"issue_type":%q,"assignee":%q}`,
			b.id, b.id, b.status, kind, b.assignee)
	}
	return []byte("[" + strings.Join(rows, ",") + "]")
}

// children records the walk's query for one parent.
func children(parent string, beads ...bead) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "list", "--parent", parent, "--all", "--limit", "0", "--json"},
		Stdout: issues(beads...),
	}
}

// ready records the ready front under the run epic.
func ready(beads ...bead) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"bd", "ready", "--parent", runEpic, "--exclude-type", "epic", "--json"},
		Stdout: issues(beads...),
	}
}

func prune() exectest.Recording {
	return exectest.Recording{Argv: []string{"git", "worktree", "prune"}}
}

// worktrees records `git worktree list --porcelain` for the given
// branches, the first entry standing for the main checkout.
func worktrees(branches ...string) exectest.Recording {
	var b strings.Builder
	for i, br := range branches {
		fmt.Fprintf(&b, "worktree /w/%d\nHEAD %040d\nbranch refs/heads/%s\n\n", i, i, br)
	}
	return exectest.Recording{
		Argv:   []string{"git", "worktree", "list", "--porcelain"},
		Stdout: []byte(b.String()),
	}
}

// commit records the newest commit on the sprint branch, ago before the
// fixed now.
func commit(sha, subject string, ago time.Duration) exectest.Recording {
	when := now.Add(-ago).Unix()
	return exectest.Recording{
		Argv:   []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", branch},
		Stdout: []byte(fmt.Sprintf("%s\x00%d\x00%s\n", sha, when, subject)),
	}
}

// ownerHead is the HEAD [worktrees] fabricates for the tree at index i —
// the same `%040d` it writes into the porcelain output — so a test naming
// the commit at that HEAD does not hand-transcribe it.
func ownerHead(i int) string { return fmt.Sprintf("%040d", i) }

// ownerCommit records the commit a worktree's HEAD sits at, for the
// worktree-ownership lookup: the same invocation [commit] makes, keyed on
// the commit itself rather than on the sprint branch.
func ownerCommit(head, subject string) exectest.Recording {
	return exectest.Recording{
		Argv:   []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", head},
		Stdout: []byte(fmt.Sprintf("%s\x00%d\x00%s\n", head, now.Unix(), subject)),
	}
}

// --- tests ----------------------------------------------------------

// AC: AC-SWARM-CORE-003.2
//
// The run epic carries a health value saying the run is healthy, and a
// state that says otherwise: nothing standing, work ready, and a commit
// hours old. Every measurement reflects the state.
func TestProgressIsNeverReadFromAStoredValue(t *testing.T) {
	m := collect(t,
		show(`{"branch":"`+branch+`","health":"healthy","progress":"on track","admitted":["demo-run1.1"]}`),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.0 P1 something", 3*time.Hour+12*time.Minute),
		ready(bead{id: "demo-run1.1", status: "open"}),
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		children("demo-run1.1"),
	)

	if !m.Worktrees.Established || m.Worktrees.Standing != 1 {
		t.Errorf("standing worktrees = %d (established %v), want the measured 1",
			m.Worktrees.Standing, m.Worktrees.Established)
	}
	if !m.LastCommit.Established || m.LastCommit.Age != 3*time.Hour+12*time.Minute {
		t.Errorf("last commit age = %v (established %v), want 3h12m",
			m.LastCommit.Age, m.LastCommit.Established)
	}
	if !m.ReadyFront.Established || m.ReadyFront.Depth != 1 {
		t.Errorf("ready depth = %d (established %v), want 1",
			m.ReadyFront.Depth, m.ReadyFront.Established)
	}

	// The stored keys reach no field. The measurement set has nowhere to
	// put them, which is the structural half of the guarantee; this
	// asserts the other half — that reading them changed no answer.
	if m.Completion.Total != 1 || m.Completion.Open != 1 {
		t.Errorf("completion = %+v, want the one open bead the tracker reported", m.Completion)
	}
}

// AC: AC-SWARM-CORE-003.3
//
// The run's beads are grandchildren of the run epic, under a phase epic.
// A parent-scoped list would report one child — the epic — and a ready
// query would omit the closed ones. The walk counts all three.
func TestCompletionCountsEveryBeadHoweverNested(t *testing.T) {
	m := collect(t,
		show(manifestWith("demo-run1.1", "demo-run1.2", "demo-run1.3")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.2 P1 done", time.Minute),
		ready(bead{id: "demo-run1.3", status: "open"}),
		children(runEpic, bead{id: "demo-target", status: "open", kind: "epic"}),
		children("demo-target",
			bead{id: "demo-run1.1", status: "closed"},
			bead{id: "demo-run1.2", status: "closed"},
			bead{id: "demo-run1.3", status: "open"},
		),
		children("demo-run1.1"),
		children("demo-run1.2"),
		children("demo-run1.3"),
	)

	if !m.Completion.Established {
		t.Fatalf("completion unestablished: %s", m.Completion.Reason)
	}
	if m.Completion.Total != 3 {
		t.Errorf("total = %d, want 3 — every bead in the run, however nested", m.Completion.Total)
	}
	if m.Completion.Done != 2 || m.Completion.Open != 1 {
		t.Errorf("done/open = %d/%d, want 2/1", m.Completion.Done, m.Completion.Open)
	}
	// The phase epic is a container, not work: counting it would inflate
	// every completion verdict by the number of phases a run has.
	for _, b := range m.Completion.Beads {
		if b.ID == "demo-target" {
			t.Errorf("the phase epic %q was counted as work", b.ID)
		}
	}
}

// AC: AC-SWARM-CORE-003.4
//
// Two dispatches. dev-1 holds a bead it has claimed and has nothing
// ready — the state the ready query alone reports as finished. dev-2
// holds nothing in either query.
func TestEachAgentHoldingWorkIsNamedAndTheDrainedOneIsReportedDrained(t *testing.T) {
	m := collect(t,
		show(manifestWith("demo-run1.1", "demo-run1.2")),
		prune(),
		// The dispatch's own worktree carries no `bb/`-prefixed branch —
		// nothing here does, matching the harness that actually creates
		// one — and is still attributed, through the bead its HEAD commit
		// closed.
		worktrees(branch, "worktree-agent-notadispatchname"),
		commit("abc1234", "demo-run1.2 P1 done", time.Minute),
		ownerCommit(ownerHead(1), "demo-run1.1 P1 in progress"),
		ready(),
		children(runEpic,
			bead{id: "demo-run1.1", status: "in_progress", assignee: "dev-1"},
			bead{id: "demo-run1.2", status: "closed", assignee: "dev-2"},
		),
		children("demo-run1.1"),
		children("demo-run1.2"),
	)

	if len(m.Dispatches) != 2 {
		t.Fatalf("dispatches = %+v, want one per agent", m.Dispatches)
	}
	one, two := m.Dispatches[0], m.Dispatches[1]
	if one.Agent != "dev-1" || two.Agent != "dev-2" {
		t.Fatalf("agents = %q, %q; want dev-1, dev-2 in order", one.Agent, two.Agent)
	}

	// The whole point of the second query: dev-1's ready queue is empty,
	// and it is not drained.
	if len(one.Ready) != 0 {
		t.Errorf("dev-1 ready = %v, want none — the fixture gives it none", one.Ready)
	}
	if one.Drained {
		t.Error("dev-1 reads drained though it holds a bead in progress; the ready query alone would say this")
	}
	if got := strings.Join(one.InProgress, ","); got != "demo-run1.1" {
		t.Errorf("dev-1 in progress = %q, want demo-run1.1", got)
	}
	if one.Worktree != "/w/1" {
		t.Errorf("dev-1 worktree = %q, want the tree on its own branch", one.Worktree)
	}
	if !two.Drained {
		t.Error("dev-2 holds neither ready nor in-progress work and is not reported drained")
	}
	if two.Worktree != "" {
		t.Errorf("dev-2 worktree = %q, want none standing", two.Worktree)
	}
}

// AC: AC-SWARM-CORE-003.7
//
// The run holds a bead the admitted list does not name — work that
// entered after admission — and the admitted list names one the run no
// longer holds.
func TestTheLiveQueueIsComparedAgainstTheAdmittedSet(t *testing.T) {
	m := collect(t,
		show(manifestWith("demo-run1.1", "demo-run1.9")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		ready(),
		children(runEpic,
			bead{id: "demo-run1.1", status: "closed"},
			bead{id: "demo-run1.4", status: "open"},
		),
		children("demo-run1.1"),
		children("demo-run1.4"),
	)

	if !m.Scope.Established {
		t.Fatalf("scope unestablished: %s", m.Scope.Reason)
	}
	if got := strings.Join(m.Scope.Additions, ","); got != "demo-run1.4" {
		t.Errorf("additions = %q, want demo-run1.4 named as scope the run grew", got)
	}
	if got := strings.Join(m.Scope.Missing, ","); got != "demo-run1.9" {
		t.Errorf("missing = %q, want the admitted bead the run no longer holds", got)
	}
}

// AC: AC-SWARM-CORE-003.7
//
// An absent admitted list is unestablished, never "no additions". The two
// read the same in a report and mean opposite things.
func TestAnAbsentAdmittedListIsUnestablishedRatherThanNoAdditions(t *testing.T) {
	m := collect(t,
		show(`{"branch":"`+branch+`"}`),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		ready(),
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		children("demo-run1.1"),
	)

	if m.Scope.Established {
		t.Fatal("the scope comparison is established though the manifest records no admitted list")
	}
	if !strings.Contains(m.Scope.Reason, "admitted") {
		t.Errorf("reason = %q, want it to name the key it could not read", m.Scope.Reason)
	}
	if len(m.Scope.Additions) != 0 {
		t.Errorf("additions = %v, want none reported from a comparison that was not made", m.Scope.Additions)
	}
}

// AC: AC-SWARM-CORE-003.8
//
// The excluded beads are carried beside the completion verdict, so a run
// that finished its admitted set while beads sat out cannot read as done
// over the whole scope.
func TestBeadsExcludedAtAdmissionAreCarriedBesideTheVerdict(t *testing.T) {
	m := collect(t,
		show(`{"branch":"`+branch+`","admitted":["demo-run1.1"],`+
			`"excluded":[{"id":"demo-run1.7","reason":"declined move"},`+
			`{"id":"demo-run1.8","reason":"declined move"}]}`),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		ready(),
		children(runEpic, bead{id: "demo-run1.1", status: "closed"}),
		children("demo-run1.1"),
	)

	if got := strings.Join(m.Scope.Excluded, ","); got != "demo-run1.7,demo-run1.8" {
		t.Errorf("excluded = %q, want both beads admission left out", got)
	}
	if m.Completion.Done != m.Completion.Total {
		t.Fatalf("fixture: the admitted work should all be closed, got %+v", m.Completion)
	}
}

// AC: AC-SWARM-CORE-003.9
//
// A blocked bead whose dispatch's worktree still stands is an unresolved
// conflict: the partial work is on disk, and the standing tree is the
// evidence it holds.
func TestABlockedBeadWhoseWorktreeStandsIsReportedUnresolved(t *testing.T) {
	m := collect(t,
		show(manifestWith("demo-run1.1", "demo-run1.2")),
		prune(),
		worktrees(branch, "worktree-agent-notadispatchname"),
		commit("abc1234", "demo-run1.2 P1 done", time.Minute),
		ownerCommit(ownerHead(1), "demo-run1.1 P1 partial work"),
		ready(),
		children(runEpic,
			bead{id: "demo-run1.1", status: "blocked", assignee: "dev-1"},
			bead{id: "demo-run1.2", status: "blocked", assignee: "dev-2"},
		),
		children("demo-run1.1"),
		children("demo-run1.2"),
	)

	if len(m.Conflicts) != 1 {
		t.Fatalf("conflicts = %+v, want only the blocked bead whose tree still stands", m.Conflicts)
	}
	c := m.Conflicts[0]
	if c.Bead != "demo-run1.1" || c.Agent != "dev-1" || c.Worktree != "/w/1" {
		t.Errorf("conflict = %+v, want demo-run1.1 held by dev-1 at /w/1", c)
	}
}

// AC: AC-SWARM-CORE-003.5
//
// One query fails and the rest still answer. The affected line is
// unknown, with the reason on it, and no measurement it did not touch is
// disturbed — which is why each measurement is its own plan.
func TestATrackerQueryThatCannotAnswerCostsOnlyItsOwnLine(t *testing.T) {
	failing := ready()
	failing.Stdout = nil
	failing.Stderr = []byte("bd: database is locked")
	failing.ExitStatus = 1

	m := collect(t,
		show(manifestWith("demo-run1.1")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", 90*time.Second),
		failing,
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		children("demo-run1.1"),
	)

	if m.ReadyFront.Established {
		t.Fatal("the ready front reads established though its query failed")
	}
	if m.ReadyFront.Depth != 0 || m.ReadyFront.Reason == "" {
		t.Errorf("ready front = %+v, want depth unreported and a reason given", m.ReadyFront)
	}
	if !m.Worktrees.Established || !m.LastCommit.Established || !m.Completion.Established {
		t.Errorf("a failed ready query disturbed other measurements: worktrees %v, commit %v, completion %v",
			m.Worktrees.Established, m.LastCommit.Established, m.Completion.Established)
	}
	if m.LastCommit.Age != 90*time.Second {
		t.Errorf("last commit age = %v, want 90s", m.LastCommit.Age)
	}
}

// A drain verdict needs both queries, so one of them failing makes every
// dispatch's verdict unestablished rather than "drained".
func TestDrainIsUnestablishedWhenEitherQueryCouldNotAnswer(t *testing.T) {
	failing := ready()
	failing.Stderr = []byte("bd: database is locked")
	failing.Stdout = nil
	failing.ExitStatus = 1

	m := collect(t,
		show(manifestWith("demo-run1.1")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		failing,
		children(runEpic, bead{id: "demo-run1.1", status: "in_progress", assignee: "dev-1"}),
		children("demo-run1.1"),
	)

	if len(m.Dispatches) != 1 {
		t.Fatalf("dispatches = %+v, want dev-1 named from the claimed set", m.Dispatches)
	}
	d := m.Dispatches[0]
	if d.Established {
		t.Error("dev-1's drain state reads established though the ready query failed")
	}
	if d.Drained {
		t.Error("dev-1 reads drained on a query that could not answer")
	}
}

// A walk that cannot finish makes the count unestablished rather than a
// number over an incomplete set — a partial walk reporting a total is how
// a run reads as finished when it is not.
func TestAWalkThatCannotFinishIsUnestablishedRatherThanAPartialCount(t *testing.T) {
	broken := children("demo-run1.1")
	broken.Stderr = []byte("bd: database is locked")
	broken.ExitStatus = 1

	m := collect(t,
		show(manifestWith("demo-run1.1")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		ready(),
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		broken,
	)

	if m.Completion.Established {
		t.Fatalf("completion reads established over an incomplete walk: %+v", m.Completion)
	}
	if m.Completion.Total != 0 {
		t.Errorf("total = %d, want no count reported at all", m.Completion.Total)
	}
	if !strings.Contains(m.Completion.Reason, "incomplete") {
		t.Errorf("reason = %q, want it to say the set is incomplete", m.Completion.Reason)
	}
}

// A run with no branch in its manifest has no sprint history, and that is
// unestablished rather than measured against whatever the process happens
// to have checked out.
func TestWithNoBranchRecordedTheCommitAgeIsUnestablishedRatherThanTakenFromHEAD(t *testing.T) {
	d, _ := deps(
		show(`{"admitted":["demo-run1.1"]}`),
		prune(),
		worktrees(branch),
		ready(),
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		children("demo-run1.1"),
	)
	m, records, err := status.Collect(context.Background(), d, runEpic)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if m.LastCommit.Established {
		t.Fatal("a commit age was established though the manifest records no branch")
	}
	if !strings.Contains(m.LastCommit.Reason, "branch") {
		t.Errorf("reason = %q, want it to name the missing branch", m.LastCommit.Reason)
	}
	// The last-commit measurement specifically never ran — worktree
	// ownership is a separate measurement that reads a commit for a
	// different reason (attributing a standing tree) and is unaffected by
	// the manifest recording no branch, so its own step is free to run.
	for _, r := range records {
		if r.Step == "last-commit" {
			t.Errorf("the last-commit step ran anyway: %+v", r)
		}
	}
}

// An empty run scope is refused rather than measured: it would walk the
// whole tracker and report another run's beads as this one's.
func TestAnEmptyRunScopeIsRefused(t *testing.T) {
	d, fake := deps()
	if _, _, err := status.Collect(context.Background(), d, ""); err == nil {
		t.Fatal("an empty run scope was accepted")
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("it ran %v before refusing", fake.Argvs())
	}
}

// Every invocation the collector made comes back with the measurements,
// so a caller can print what established each line.
func TestItReturnsTheRecordOfEveryInvocationItMade(t *testing.T) {
	d, fake := deps(
		show(manifestWith("demo-run1.1")),
		prune(),
		worktrees(branch),
		commit("abc1234", "demo-run1.1 P1 done", time.Minute),
		ready(),
		children(runEpic, bead{id: "demo-run1.1", status: "open"}),
		children("demo-run1.1"),
	)
	_, records, err := status.Collect(context.Background(), d, runEpic)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(records) != len(fake.Argvs()) {
		t.Errorf("%d records for %d invocations; every one has to be reportable",
			len(records), len(fake.Argvs()))
	}
	for _, r := range records {
		if r.Step == "" {
			t.Errorf("a record carries no step name: %+v", r)
		}
	}
}
