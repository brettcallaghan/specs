// The `swarm ready` command: what may be dispatched now, across every
// dispatch in the run. Per the "swarm ready" delta of
// [[container-swarm-cli]] (added — the dispatch set) and
// FEAT-SWARM-DISP-002 of specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// It carries no noun before the verb — the same reasoning
// specs/features/swarm-core/swarm-core-frd.md gives for `swarm status`:
// every command here concerns one run's graph, so naming that graph
// would say nothing, and `bd ready` is the tracker's own word for the
// same question over the whole repository.
//
// `--agent` (one dispatch's own queue, in dependency order, with its
// drain verdict) and the consecutive-failure breaker are a later delta's
// — AQ and CB narrow and extend the filter set built here rather than
// restating it.
package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// humanLabel is the label grooming attaches to a bead awaiting a human
// decision. Excluded here and nowhere else: a completion count that hid
// it would report a run as drained while a human still owes it an
// answer — AC-SWARM-DISP-002.1 and .2.
const humanLabel = "human"

// readyCmd builds `swarm ready`.
func readyCmd(streams *ioStreams, deps status.Deps) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
	)

	cmd := &cobra.Command{
		Use:   "ready",
		Short: "Report what may be dispatched now",
		Long: "Reports the beads under the run that nothing is blocking, with the run's own " +
			"filter applied — no epic, no bead closed or deferred, none awaiting a human answer " +
			"— each named with the dispatch id it is assigned to, or unassigned when none is. A " +
			"bead assigned to a dispatch whose worktree is standing is reported held rather than " +
			"available, so a caller does not send a second agent after one already at work.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req := Request{
				Command:    cmd.CommandPath(),
				Run:        runID,
				JSON:       asJSON,
				FullResult: fullResult || explain || dryRun,
			}

			if strings.TrimSpace(runID) == "" {
				return finish(streams, req, refusal(req, RefuseArguments(
					"%s is scoped to one run; pass --run <run-epic-id>", req.Command)))
			}

			switch {
			case explain && dryRun:
				return finish(streams, req, refusal(req, RefuseArguments(
					"--explain and --dry-run select different interpreters; pass one")))
			case explain:
				req.Mode = plan.ModeExplain
			case dryRun:
				req.Mode = plan.ModeDryRun
			default:
				req.Mode = plan.ModeExecute
			}

			if req.Mode == plan.ModeExplain {
				return finish(streams, req, explainReady(req))
			}

			result, err := reportReady(context.Background(), deps, req)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe the invocations this command would run, and run none of them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"perform every read this command would perform — it writes nothing either way")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")

	return cmd
}

// explainReady describes the measurements this command would take
// without taking any — the ready query is one invocation, but "held"
// requires the same worktree attribution `swarm status` performs, which
// is not a plan this package renders a second time.
func explainReady(req Request) Result {
	lines := []string{
		"swarm ready reads two things and joins them, so its steps are described here rather " +
			"than printed as a plan:",
		"  bd ready --parent " + req.Run + " --exclude-type epic --exclude-label human --limit 0 --json  (the dispatch set)",
		"  the same worktree-to-dispatch attribution `swarm status` performs — the standing " +
			"worktrees, the run's completion walk, and which bead's assignee closed each worktree's " +
			"HEAD commit — to tell a bead already held from one nothing is working",
	}
	return Result{
		Command: req.Command, Run: req.Run, Mode: req.Mode, Outcome: OutcomeClear,
		Summary: "explained; nothing was measured", Transcript: lines,
	}
}

// readyEntry is one bead in the dispatch set, rendered as [Structured]
// JSON: the id, the dispatch it is assigned to (empty, never omitted,
// when it is assigned to none — AC-SWARM-DISP-002.17), and whether that
// dispatch's worktree is standing.
type readyEntry struct {
	ID       string `json:"id"`
	Dispatch string `json:"dispatch"`
	Held     bool   `json:"held"`
}

// reportReady runs the dispatch query and joins it against the same
// worktree ownership [status.Collect] measures, so the algorithm that
// matches a standing worktree to its dispatch — by branch, falling back
// to the bead its HEAD commit closed — is the one implementation
// [[container-swarm-cli]] declares, not a second copy of it living here.
func reportReady(ctx context.Context, deps status.Deps, req Request) (Result, error) {
	step := deps.Tracker.Ready("ready", req.Run, tracker.ReadyFilter{
		ExcludeKinds:  []string{epicKind},
		ExcludeLabels: []string{humanLabel},
	})
	report, err := deps.Runner.Run(ctx, plan.Plan{
		Name:  "the dispatch set under " + req.Run,
		Steps: []plan.Step{step},
	}, req.Mode)
	if err != nil {
		return Result{}, RefuseUnrepairable("the dispatch set could not run: %s", err)
	}
	rec, ok := report.Record("ready")
	if !ok || rec.Failed() {
		// AC-SWARM-DISP-002.16: a tracker response that is not a queue's
		// shape fails to decode here, rather than a length this package
		// derived from whatever it was.
		return Result{}, RefuseUnrepairable("the dispatch set could not be read: %s", describeFailure(rec))
	}
	beads, err := tracker.BeadsFrom(rec.Value)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}
	readyInv := argvOf(rec)

	m, records, err := status.Collect(ctx, deps, req.Run)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	held := map[string]bool{}
	for _, d := range m.Dispatches {
		if d.Worktree != "" {
			held[d.Agent] = true
		}
	}

	entries := make([]readyEntry, 0, len(beads))
	for _, b := range beads {
		entries = append(entries, readyEntry{ID: b.ID, Dispatch: b.Assignee, Held: held[b.Assignee]})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })

	result := Result{Outcome: OutcomeClear}
	result.Records = append(result.Records, report.Records...)
	result.Records = append(result.Records, records...)
	result.Facts = append(result.Facts,
		Structured("ready", marshal(entries), readyInv),
		Count("ready_count", len(entries), readyInv))

	switch len(entries) {
	case 0:
		result.Summary = fmt.Sprintf("nothing in %s is ready to dispatch", req.Run)
	default:
		var heldCount int
		for _, e := range entries {
			if e.Held {
				heldCount++
			}
		}
		result.Summary = fmt.Sprintf("%d bead(s) ready under %s, %d already held by a standing dispatch",
			len(entries), req.Run, heldCount)
	}
	return result, nil
}
