// Tests for the queue derivation. They are written against the same fake
// tracker the walk's tests use, because what is asserted is a property of
// the derivation — that both halves are required, and that the half no
// query can scope is narrowed to the run — rather than of any tracker's
// flags.
package tracker_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// queueFake replays the walk plus the two queue halves. The walk is what
// gives the held half a descendant set to be narrowed against.
func queueFake(children map[string][]string, nodes []string, ready, held string) *exectest.Fake {
	fake := tree(children, nodes...)
	return exectest.NewFake(append(fake.Recordings,
		exectest.Recording{Argv: []string{"tree", "ready", "run-1", "dev-1"}, Stdout: []byte(ready)},
		exectest.Recording{Argv: []string{"tree", "holding", "dev-1"}, Stdout: []byte(held)},
	)...)
}

var queueTree = map[string][]string{
	"run-1":       {"run-1.phase"},
	"run-1.phase": {"run-1.1", "run-1.2"},
}

var queueNodes = []string{"run-1", "run-1.phase", "run-1.1", "run-1.2"}

// The held half is the one no query can scope, so a bead the same
// dispatch id holds in ANOTHER run must not reach this run's queue.
func TestAnAgentsQueueDoesNotReachOutsideTheRun(t *testing.T) {
	fake := queueFake(queueTree, queueNodes, `["run-1.1"]`, `["run-1.2","other-9.1"]`)

	queue, _, err := tracker.AgentQueue(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1", "dev-1", tracker.ReadyFilter{})
	if err != nil {
		t.Fatalf("AgentQueue: %v", err)
	}

	if got := ids(queue.Ready); strings.Join(got, " ") != "run-1.1" {
		t.Errorf("Ready = %v, want [run-1.1]", got)
	}
	if got := ids(queue.InProgress); strings.Join(got, " ") != "run-1.2" {
		t.Errorf("InProgress = %v, want only run-1.2 — other-9.1 is another run's", got)
	}
	if queue.Drained() {
		t.Error("the queue reads drained though the agent holds two beads")
	}
}

// An agent holding one bead in progress and none ready is still holding
// work. The ready half alone reports it as finished, which is the reading
// that lets a dispatch be accepted mid-bead.
func TestAnAgentMidBeadIsNotDrained(t *testing.T) {
	fake := queueFake(queueTree, queueNodes, `[]`, `["run-1.2"]`)

	queue, _, err := tracker.AgentQueue(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1", "dev-1", tracker.ReadyFilter{})
	if err != nil {
		t.Fatalf("AgentQueue: %v", err)
	}

	if len(queue.Ready) != 0 {
		t.Errorf("Ready = %v, want none", ids(queue.Ready))
	}
	if queue.Drained() {
		t.Error("an agent holding a bead in progress reads drained; the ready half alone would say this")
	}
}

func TestAnAgentHoldingNothingIsDrained(t *testing.T) {
	fake := queueFake(queueTree, queueNodes, `[]`, `[]`)

	queue, _, err := tracker.AgentQueue(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1", "dev-1", tracker.ReadyFilter{})
	if err != nil {
		t.Fatalf("AgentQueue: %v", err)
	}
	if !queue.Drained() {
		t.Errorf("the queue is not drained: %+v", queue)
	}
}

// A half that could not be read is not an empty half. Reporting a drain
// from one query that answered and one that failed is an agent reported
// finished because a query broke.
func TestAHalfThatCouldNotBeReadIsNotAnEmptyHalf(t *testing.T) {
	fake := queueFake(queueTree, queueNodes, `[]`, `{"error": "the tracker fell over"}`)

	queue, records, err := tracker.AgentQueue(context.Background(), treeTracker{},
		plan.Runner{Exec: fake}, "run-1", "dev-1", tracker.ReadyFilter{})
	if err == nil {
		t.Fatalf("AgentQueue returned no error though a half could not be read: %+v", queue)
	}
	if queue.Drained() && len(queue.Ready) == 0 && len(queue.InProgress) == 0 {
		// The zero Queue reads drained by construction, so the error is
		// the only thing keeping a caller from acting on it — which is
		// why the error must not be droppable.
		if !strings.Contains(err.Error(), "queue") {
			t.Errorf("the error does not say a queue could not be built: %v", err)
		}
	}
	if len(records) == 0 {
		t.Error("no invocation was recorded, so the failure cannot be explained")
	}
}

// A scope or an agent nobody named would answer for the whole tracker or
// for every dispatch at once. Both are refused before anything runs.
func TestAnEmptyScopeOrAgentIsRefusedBeforeAnythingRuns(t *testing.T) {
	fake := exectest.NewFake()

	for _, c := range []struct{ scope, agent string }{{"", "dev-1"}, {"run-1", ""}} {
		_, records, err := tracker.AgentQueue(context.Background(), treeTracker{},
			plan.Runner{Exec: fake}, c.scope, c.agent, tracker.ReadyFilter{})
		if err == nil {
			t.Errorf("AgentQueue(scope=%q, agent=%q) returned no error", c.scope, c.agent)
		}
		if len(records) != 0 {
			t.Errorf("AgentQueue(scope=%q, agent=%q) ran %d invocations before refusing",
				c.scope, c.agent, len(records))
		}
	}
}
