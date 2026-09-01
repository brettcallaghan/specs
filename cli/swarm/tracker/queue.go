package tracker

import (
	"context"
	"fmt"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// AgentQueue returns one dispatch's queue beneath scope — the beads
// ready for it and the beads it already holds — together with the record
// of every invocation it made.
//
// **Why this is a derivation and not one query.** The partition takes two
// answers that no single tracker query gives together: a ready query
// excludes work in progress by construction, so an agent mid-bead has an
// empty ready set and reads as finished, while the in-progress set says
// nothing about what is waiting for it. [Tracker.Ready] and
// [Tracker.Holding] each declare a step, and this is where their two
// answers become one verdict — in one place, so a caller checking drain
// and a caller listing a queue cannot disagree about what drained means.
//
// **The two halves are scoped differently, and that is not an oversight.**
// The ready half is scoped by the query itself. The held half cannot be:
// see [Tracker.Holding] for the measurement, and [Held.Within] for where
// its narrowing happens. That narrowing needs the run's descendant set,
// which is why this walks — and the walk is also what makes an
// unreachable bead an error here rather than a queue that quietly omits
// it.
//
// filter carries the run's own filters; its Assignee is replaced with
// agent, since a queue for one dispatch is what this answers.
//
// A [Queue] whose halves are both empty is DRAINED, which is not the same
// as a dispatch nobody gave work to. Telling those apart needs the run's
// whole assignment set and is the caller's; this answers only what the
// named agent holds.
func AgentQueue(
	ctx context.Context,
	t Tracker,
	runner plan.Runner,
	scope, agent string,
	filter ReadyFilter,
) (Queue, []plan.Record, error) {
	if scope == "" {
		return Queue{}, nil, fmt.Errorf("tracker: AgentQueue needs a scope; an empty one would answer for the whole tracker")
	}
	if agent == "" {
		return Queue{}, nil, fmt.Errorf("tracker: AgentQueue needs an agent; an empty one would answer for every dispatch at once")
	}

	beneath, records, err := Descendants(ctx, t, runner, scope)
	if err != nil {
		return Queue{}, records, fmt.Errorf("tracker: %s's queue under %s: %w", agent, scope, err)
	}

	filter.Assignee = agent
	readyStep := t.Ready("ready", scope, filter)
	heldStep := t.Holding("held", scope, agent)

	report, err := runner.Run(ctx, plan.Plan{
		Name:  fmt.Sprintf("%s's queue under %s", agent, scope),
		Steps: []plan.Step{readyStep, heldStep},
	}, plan.ModeExecute)
	records = append(records, report.Records...)
	if err != nil {
		return Queue{}, records, fmt.Errorf("tracker: %s's queue under %s: %w", agent, scope, err)
	}

	queue := Queue{Scope: scope, Agent: agent}
	// Which halves actually produced a value. A step the runner SKIPPED
	// records no failure, so tracking failures alone would let a skipped
	// half arrive as an empty one — which is the drain verdict this
	// derivation exists to refuse.
	answered := map[string]bool{}
	for _, rec := range report.Records {
		// A half that did not answer must not become an empty half: a
		// drain verdict built on one is an agent reported finished
		// because a query failed.
		if rec.Failed() {
			return Queue{}, records, fmt.Errorf(
				"tracker: %s's queue under %s: step %q failed, and a queue is not reported from one half",
				agent, scope, rec.Step)
		}
		switch rec.Step {
		case readyStep.Name:
			beads, err := BeadsFrom(rec.Value)
			if err != nil {
				return Queue{}, records, fmt.Errorf("tracker: step %q: %w", rec.Step, err)
			}
			queue.Ready = beads
			answered[rec.Step] = true
		case heldStep.Name:
			held, err := HeldFrom(rec.Value)
			if err != nil {
				return Queue{}, records, fmt.Errorf("tracker: step %q: %w", rec.Step, err)
			}
			queue.InProgress = held.Within(beneath)
			answered[rec.Step] = true
		}
	}

	for _, step := range []string{readyStep.Name, heldStep.Name} {
		if !answered[step] {
			return Queue{}, records, fmt.Errorf(
				"tracker: %s's queue under %s: step %q produced no answer, and a queue is not reported from one half",
				agent, scope, step)
		}
	}

	return queue, records, nil
}
