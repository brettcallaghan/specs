// Tests for the transitive walk. They are written against a fake tracker
// rather than the beads implementation, because what is being asserted is
// a property of the walk — every bead below the scope, however deep —
// rather than of any tracker's flags.
//
// The beads implementation's own tests pin the flags, and the measurement
// that makes this walk necessary is recorded on [Descendants]: `bd list
// --parent` answers with direct children and `bd ready --parent` answers
// transitively but only with what is ready, so neither is "every bead in
// this run".
package tracker_test

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// treeTracker is a [tracker.Tracker] whose Children step invokes a
// fictional tool and decodes the JSON the fake replays. Only Children is
// exercised by the walk; the rest satisfy the interface.
type treeTracker struct{}

func (treeTracker) Children(name, parent string) plan.Step {
	return plan.Step{
		Name: name, Executable: "tree", Args: []plan.Arg{plan.Lit("children"), plan.Lit(parent)},
		Effect: plan.Read, Repeatable: true,
		Decode: func(res swarmexec.Result) (any, error) {
			body := strings.TrimSpace(string(res.Stdout))
			if body == "" || body == "null" {
				return []tracker.Bead{}, nil
			}
			var ids []string
			if err := json.Unmarshal([]byte(body), &ids); err != nil {
				return nil, err
			}
			beads := make([]tracker.Bead, 0, len(ids))
			for _, id := range ids {
				beads = append(beads, tracker.Bead{ID: id, Parent: parent, State: tracker.StateOpen})
			}
			return beads, nil
		},
	}
}

func (treeTracker) Bead(name, id string) plan.Step { return plan.Step{Name: name, Executable: "tree"} }

// Ready and Holding are the two halves [tracker.AgentQueue] is built
// from, so unlike the rest they decode what the fake replays. Each
// answers with a JSON list of ids.
func (treeTracker) Ready(name, scope string, filter tracker.ReadyFilter) plan.Step {
	return plan.Step{
		Name: name, Executable: "tree",
		Args:   []plan.Arg{plan.Lit("ready"), plan.Lit(scope), plan.Lit(filter.Assignee)},
		Effect: plan.Read, Repeatable: true,
		Decode: func(res swarmexec.Result) (any, error) {
			return decodeIDs(res, scope)
		},
	}
}
func (treeTracker) Holding(name, scope, agent string) plan.Step {
	return plan.Step{
		Name: name, Executable: "tree",
		Args:   []plan.Arg{plan.Lit("holding"), plan.Lit(agent)},
		Effect: plan.Read, Repeatable: true,
		Decode: func(res swarmexec.Result) (any, error) {
			beads, err := decodeIDs(res, "")
			if err != nil {
				return nil, err
			}
			return tracker.Held{Scope: scope, Agent: agent, Beads: beads}, nil
		},
	}
}

// decodeIDs turns the fake's JSON id list into beads.
func decodeIDs(res swarmexec.Result, parent string) ([]tracker.Bead, error) {
	body := strings.TrimSpace(string(res.Stdout))
	if body == "" || body == "null" {
		return []tracker.Bead{}, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(body), &ids); err != nil {
		return nil, err
	}
	beads := make([]tracker.Bead, 0, len(ids))
	for _, id := range ids {
		beads = append(beads, tracker.Bead{ID: id, Parent: parent, State: tracker.StateOpen})
	}
	return beads, nil
}
func (treeTracker) ClosableEpics(name, scope string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) Findings(name, id string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) Claim(name, id, agent string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) SetState(name, id string, _ tracker.State) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) AddLabel(name, id, label string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) Note(name, id, body string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) BlockedBy(name, bead, blocker string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}
func (treeTracker) DiscoveredFrom(name, bead, origin string) plan.Step {
	return plan.Step{Name: name, Executable: "tree"}
}

// tree builds the recordings for a parent -> children map. A parent with
// no entry answers with an empty list, so a leaf needs no declaration.
func tree(edges map[string][]string, nodes ...string) *exectest.Fake {
	var recordings []exectest.Recording
	for _, node := range nodes {
		body, _ := json.Marshal(edges[node])
		if edges[node] == nil {
			body = []byte("[]")
		}
		recordings = append(recordings, exectest.Recording{
			Argv:   []string{"tree", "children", node},
			Stdout: body,
		})
	}
	return exectest.NewFake(recordings...)
}

func ids(beads []tracker.Bead) []string {
	out := make([]string, 0, len(beads))
	for _, b := range beads {
		out = append(out, b.ID)
	}
	sort.Strings(out)
	return out
}

// AC: AC-SWARM-CORE-003.3
//
// A bead nested below the run epic is in the set. This is the whole point
// of the walk: a run whose work sits under phase epics has its beads
// invisible to a direct-children query.
func TestEveryBeadBelowTheScopeIsFoundHoweverDeep(t *testing.T) {
	edges := map[string][]string{
		"run-1":         {"run-1.phase-a", "run-1.phase-b"},
		"run-1.phase-a": {"run-1.1", "run-1.2"},
		"run-1.phase-b": {"run-1.epic"},
		"run-1.epic":    {"run-1.3"},
	}
	fake := tree(edges, "run-1", "run-1.phase-a", "run-1.phase-b", "run-1.epic",
		"run-1.1", "run-1.2", "run-1.3")

	found, records, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1")
	if err != nil {
		t.Fatalf("Descendants: %v", err)
	}

	want := []string{"run-1.1", "run-1.2", "run-1.3", "run-1.epic", "run-1.phase-a", "run-1.phase-b"}
	if got := ids(found); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("found %v, want %v", got, want)
	}
	// run-1.3 is two levels below the scope. A direct-children query
	// would have returned the two phase epics and stopped.
	if len(records) != len(want)+1 {
		t.Errorf("the walk recorded %d invocations, want one per bead visited plus the scope (%d)",
			len(records), len(want)+1)
	}
	for _, rec := range records {
		if rec.Invocation.Name != "tree" {
			t.Errorf("a recorded invocation is not the tracker's: %+v", rec.Invocation)
		}
	}
}

func TestAScopeWithNoChildrenYieldsNothingAndSaysSoWithoutError(t *testing.T) {
	found, records, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: tree(nil, "run-1")}, "run-1")
	if err != nil {
		t.Fatalf("Descendants: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("found %v, want nothing", ids(found))
	}
	if len(records) != 1 {
		t.Errorf("the walk made %d invocations, want the one that asked", len(records))
	}
}

// A cycle terminates. A tracker's graph should not contain one, but a
// walk that loops forever on a malformed graph is worse than one that
// reports what it found.
func TestACycleTerminates(t *testing.T) {
	edges := map[string][]string{
		"run-1": {"a"},
		"a":     {"b"},
		"b":     {"a", "run-1"},
	}
	found, _, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: tree(edges, "run-1", "a", "b")}, "run-1")
	if err != nil {
		t.Fatalf("Descendants: %v", err)
	}
	if got := ids(found); strings.Join(got, " ") != "a b" {
		t.Errorf("found %v, want [a b] once each", got)
	}
}

// An empty scope would walk the whole tracker. The refusal is the same
// class as the command surface's run-scope refusal: scope is an argument,
// not a convention.
func TestAnEmptyScopeIsRefused(t *testing.T) {
	_, _, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: exectest.NewFake()}, "")
	if err == nil {
		t.Fatal("Descendants accepted an empty scope")
	}
	if !strings.Contains(err.Error(), "scope") {
		t.Errorf("the refusal does not say what was missing: %v", err)
	}
}

// A level that failed makes the whole set incomplete, and an incomplete
// set reported as if it were whole is exactly the silent failure this
// container exists to remove.
func TestAFailedLevelIsAnErrorRatherThanAShorterAnswer(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"tree", "children", "run-1"}, Stdout: []byte(`["a","b"]`)},
		exectest.Recording{Argv: []string{"tree", "children", "a"}, Stdout: []byte(`[]`)},
		exectest.Recording{Argv: []string{"tree", "children", "b"}, ExitStatus: 1, Stderr: []byte("boom")},
	)

	found, records, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1")
	if err == nil {
		t.Fatalf("Descendants returned no error though a level failed; it found %v", ids(found))
	}
	if !strings.Contains(err.Error(), "incomplete") {
		t.Errorf("the error does not say the set is incomplete: %v", err)
	}
	// What it did read is still reported, and so is every invocation, so
	// a caller can say how far it got.
	if len(found) == 0 {
		t.Error("the error discarded what the walk had already read")
	}
	if len(records) == 0 {
		t.Error("the error discarded the record of what was invoked")
	}
}

// Every invocation the walk makes goes through a plan step, so each one
// is printed, timed and digested exactly as any other.
func TestEveryInvocationTheWalkMakesIsARecordedPlanStep(t *testing.T) {
	edges := map[string][]string{"run-1": {"a"}}
	var out strings.Builder

	_, records, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: tree(edges, "run-1", "a"), Out: &out}, "run-1")
	if err != nil {
		t.Fatalf("Descendants: %v", err)
	}

	for _, rec := range records {
		if rec.Status != plan.StepRan {
			t.Errorf("step %q status = %q", rec.Step, rec.Status)
		}
		if rec.StdoutDigest == "" || rec.StderrDigest == "" {
			t.Errorf("step %q carries no stream digests", rec.Step)
		}
	}
	for _, want := range []string{"tree children run-1", "tree children a"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the transcript does not carry %q:\n%s", want, out.String())
		}
	}
}

// The step names are unique across the whole walk. Two steps sharing a
// name in one plan is a validation failure, and a walk that produced one
// would fail on a wide level rather than on a narrow one — the kind of
// defect that shows up only on a real run.
func TestTheWalkNamesEveryStepUniquely(t *testing.T) {
	edges := map[string][]string{
		"run-1": {"a", "b", "c"},
		"a":     {"a1", "a2"},
		"b":     {"b1"},
	}
	nodes := []string{"run-1", "a", "b", "c", "a1", "a2", "b1"}

	_, records, err := tracker.Descendants(context.Background(), treeTracker{},
		plan.Runner{Exec: tree(edges, nodes...)}, "run-1")
	if err != nil {
		t.Fatalf("Descendants: %v", err)
	}

	seen := map[string]bool{}
	for _, rec := range records {
		if seen[rec.Step] {
			t.Errorf("step name %q is used twice", rec.Step)
		}
		seen[rec.Step] = true
	}
	if len(seen) != len(nodes) {
		t.Errorf("the walk made %d distinct steps, want %d", len(seen), len(nodes))
	}
}

// BeadsFrom and BeadFrom separate "the step returned something else" from
// "the step returned nothing", which is a distinction this container
// refuses to conflate.
func TestReadingAStepsValueDistinguishesWrongTypeFromNoValue(t *testing.T) {
	if _, err := tracker.BeadsFrom(nil); err == nil {
		t.Error("BeadsFrom(nil) returned no error")
	}
	if _, err := tracker.BeadsFrom("not a bead set"); err == nil {
		t.Error("BeadsFrom of the wrong type returned no error")
	}
	beads, err := tracker.BeadsFrom([]tracker.Bead{{ID: "x"}})
	if err != nil || len(beads) != 1 {
		t.Errorf("BeadsFrom of a bead set: %v, %v", beads, err)
	}

	if _, err := tracker.BeadFrom(nil); err == nil {
		t.Error("BeadFrom(nil) returned no error")
	}
	if _, err := tracker.BeadFrom([]tracker.Bead{}); err == nil {
		t.Error("BeadFrom of a set returned no error")
	}
	if bead, err := tracker.BeadFrom(tracker.Bead{ID: "x"}); err != nil || bead.ID != "x" {
		t.Errorf("BeadFrom of a bead: %v, %v", bead, err)
	}
}
