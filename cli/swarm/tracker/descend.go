package tracker

import (
	"context"
	"fmt"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// maxDepth bounds the walk. A run's tree is a run epic, its phase epics
// and their task children — three levels, with room to spare. The bound
// exists so a malformed graph costs a bounded number of queries rather
// than an unbounded one; a graph deeper than this is reported, never
// silently truncated.
const maxDepth = 32

// Descendants returns every bead below scope, transitively, together with
// the record of every invocation it made.
//
// **Why this is a walk and not a query.** A tracker's parent-scoped
// queries answer different questions, and none of them is "every bead in
// this run". Measured against bd 1.2.2 on this repository's own database,
// with a run epic whose work sits under a nested epic:
//
//	bd list  --parent <run> --json  -> the run's DIRECT children only
//	bd ready --parent <run> --json  -> transitive, but only what is ready
//
// The grandchildren appear in one and not the other, and the two flags
// share a name without sharing a meaning. So the transitive set is
// derived here, level by level, rather than trusted to either.
//
// Every invocation still goes through a plan step, so each one is
// printed, timed and recorded exactly as any other. What cannot be a
// single plan is the whole walk: its next query depends on the last one's
// answer, and a plan is renderable before anything runs precisely because
// it does not do that. A caller explaining a command that walks says so —
// one children query per level of the run's tree — rather than printing a
// plan it cannot know the shape of.
//
// A bead reached twice is walked once, so a cycle terminates.
func Descendants(
	ctx context.Context,
	t Tracker,
	runner plan.Runner,
	scope string,
) ([]Bead, []plan.Record, error) {
	if scope == "" {
		return nil, nil, fmt.Errorf("tracker: Descendants needs a scope; an empty one would walk the whole tracker")
	}

	var (
		found   []Bead
		records []plan.Record
		seen    = map[string]bool{scope: true}
		level   = []string{scope}
	)

	for depth := 0; len(level) > 0; depth++ {
		if depth >= maxDepth {
			return found, records, fmt.Errorf(
				"tracker: the tree below %s is deeper than %d levels; walked %d beads and stopped rather than truncating silently",
				scope, maxDepth, len(found))
		}

		steps := make([]plan.Step, 0, len(level))
		for i, parent := range level {
			steps = append(steps, t.Children(fmt.Sprintf("children-%d-%d", depth, i), parent))
		}

		report, err := runner.Run(ctx, plan.Plan{
			Name:  fmt.Sprintf("the beads under %s, level %d", scope, depth),
			Steps: steps,
		}, plan.ModeExecute)
		records = append(records, report.Records...)
		if err != nil {
			return found, records, fmt.Errorf("tracker: walking below %s: %w", scope, err)
		}

		var next []string
		for _, rec := range report.Records {
			if rec.Failed() {
				return found, records, fmt.Errorf(
					"tracker: reading the beads under %s failed at step %q; the set below %s is incomplete and is not reported as if it were whole",
					scope, rec.Step, scope)
			}
			children, err := BeadsFrom(rec.Value)
			if err != nil {
				return found, records, fmt.Errorf("tracker: step %q: %w", rec.Step, err)
			}
			for _, b := range children {
				if seen[b.ID] {
					continue
				}
				seen[b.ID] = true
				found = append(found, b)
				next = append(next, b.ID)
			}
		}
		level = next
	}

	return found, records, nil
}
