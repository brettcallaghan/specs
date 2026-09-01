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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
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
		agentID    string
		threshold  int
		completed  int
	)

	cmd := &cobra.Command{
		Use:   "ready",
		Short: "Report what may be dispatched now",
		Long: "Reports the beads under the run that nothing is blocking, with the run's own " +
			"filter applied — no epic, no bead closed or deferred, none awaiting a human answer " +
			"— each named with the dispatch id it is assigned to, or unassigned when none is. A " +
			"bead assigned to a dispatch whose worktree is standing is reported held rather than " +
			"available, so a caller does not send a second agent after one already at work. With " +
			"--agent, the same query narrows to one dispatch's own queue, in dependency order, " +
			"together with its drain verdict.",
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

			agent := strings.TrimSpace(agentID)

			if req.Mode == plan.ModeExplain {
				return finish(streams, req, explainReady(req, agent))
			}

			var (
				result Result
				err    error
			)
			if agent != "" {
				result, err = reportAgentQueue(context.Background(), deps, req, agent)
			} else {
				report := breakerReport{Given: cmd.Flags().Changed("completed"), Completed: completed}
				result, err = reportReady(context.Background(), deps, req, report, threshold)
			}
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
	cmd.Flags().StringVar(&agentID, "agent", "",
		"narrow to one dispatch's own queue, in dependency order, with its drain verdict")
	cmd.Flags().IntVar(&threshold, "threshold", breakerDefaultThreshold,
		"the circuit breaker's threshold: the number of consecutive dispatch returns that completed no bead which stops the run")
	cmd.Flags().IntVar(&completed, "completed", 0,
		"report that a dispatch just returned having completed this many beads (0 counts as a failure); omitted, this invocation only reads the breaker's count")

	return cmd
}

// explainReady describes the measurements this command would take
// without taking any — the ready query is one invocation, but "held"
// requires the same worktree attribution `swarm status` performs, which
// is not a plan this package renders a second time.
func explainReady(req Request, agent string) Result {
	var lines []string
	if agent == "" {
		lines = []string{
			"swarm ready reads two things and joins them, then reads the circuit breaker's count, " +
				"so its steps are described here rather than printed as a plan:",
			"  bd ready --parent " + req.Run + " --exclude-type epic --exclude-label human --limit 0 --json  (the dispatch set)",
			"  the same worktree-to-dispatch attribution `swarm status` performs — the standing " +
				"worktrees, the run's completion walk, and which bead's assignee closed each worktree's " +
				"HEAD commit — to tell a bead already held from one nothing is working",
			"  bd show " + req.Run + " --json                              (the failure count; --completed also writes it back)",
		}
	} else {
		lines = []string{
			"swarm ready --agent " + agent + " reads the run's whole descendant set, then joins " +
				"two more queries against it, so its steps are described here rather than printed as " +
				"a plan:",
			"  one bd list --parent <bead> --all --limit 0 --json per level below " + req.Run +
				"  (every bead in the run, so a dispatch never given anything can be told from one that finished)",
			"  bd ready --parent " + req.Run + " --assignee " + agent +
				" --exclude-type epic --exclude-label human --limit 0 --json  (this dispatch's ready half)",
			"  bd list --assignee " + agent + " --status in_progress --limit 0 --json  (this dispatch's held half, narrowed to " + req.Run + ")",
		}
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
//
// breaker and threshold are the CB delta's: whether this invocation is
// reporting a dispatch return, and the count that stops the run. They
// apply only here — --agent narrows to one dispatch's own queue and
// carries no circuit-breaker semantics of its own.
func reportReady(ctx context.Context, deps status.Deps, req Request, breaker breakerReport, threshold int) (Result, error) {
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

	breakerFacts, stopped, stopSummary, err := applyBreaker(ctx, deps, req, breaker, threshold)
	if err != nil {
		return Result{}, err
	}
	result.Facts = append(result.Facts, breakerFacts...)
	if stopped {
		// The breaker overrides whatever the dispatch set itself
		// concluded: a stopped run has nothing left to dispatch, whether
		// or not beads remain in the ready set.
		result.Outcome = OutcomeStop
		result.Summary = stopSummary
	}
	return result, nil
}

// breakerDefaultThreshold is used WHERE --threshold is absent from the
// invocation, per the CB delta's own behavior line.
const breakerDefaultThreshold = 3

// breakerReport is what one invocation told the circuit breaker about the
// dispatch that just returned.
//
// Given is whether --completed was passed at all, and it is not the same
// question as whether Completed is zero: an invocation that omits
// --completed is a bare re-query between dispatches and must not
// silently advance or reset the count nobody reported a return for, while
// --completed 0 is a dispatch that returned having completed nothing and
// SHALL advance it.
type breakerReport struct {
	Given     bool
	Completed int
}

// applyBreaker reads consecutive_dispatch_failures, advances it by one or
// resets it to zero per report, and compares the result against
// threshold — the CB delta. The read always happens, so a bare `swarm
// ready` still reports the count and still stops a run already at
// threshold; the write happens only WHEN report.Given, since a query that
// is not reporting a dispatch return has nothing to record.
//
// It reaches the manifest only through [status.Deps.Store], per
// [[container-swarm-cli]]'s Manifest Store component — never the run
// epic's metadata directly.
func applyBreaker(ctx context.Context, deps status.Deps, req Request, report breakerReport, threshold int) ([]Fact, bool, string, error) {
	if deps.Store == nil {
		return nil, false, "", RefuseUnrepairable("no manifest store is configured")
	}

	readSteps := deps.Store.Read("failure-count", req.Run)
	readReport, err := deps.Runner.Run(ctx, plan.Plan{
		Name:  "read " + req.Run + "'s failure count",
		Steps: readSteps,
	}, plan.ModeExecute)
	if err != nil {
		return nil, false, "", RefuseUnrepairable("the manifest could not be read: %s", err)
	}
	doc, from, err := document(readSteps, readReport)
	if err != nil {
		return nil, false, "", err
	}

	var recorded int
	if present, decErr := doc.Decode("consecutive_dispatch_failures", &recorded); decErr != nil {
		return nil, false, "", RefuseUnrepairable("consecutive_dispatch_failures does not decode: %s", decErr)
	} else if !present {
		// MFK's rule that an absent count is absent, not zero, is the
		// record's own reading. Treating it as zero here is this
		// command's decision, per the delta's own behavior line, so a
		// fresh run starts at zero rather than being reported
		// unestablished.
		recorded = 0
	}

	next := recorded
	invocation := from
	if report.Given {
		if report.Completed > 0 {
			next = 0
		} else {
			next = recorded + 1
		}
		if next != recorded {
			nextJSON, merr := json.Marshal(next)
			if merr != nil {
				return nil, false, "", RefuseUnrepairable("%d does not encode as JSON: %s", next, merr)
			}
			changes := map[string]json.RawMessage{"consecutive_dispatch_failures": json.RawMessage(nextJSON)}
			writeSteps := deps.Store.Write("advance-failure-count", req.Run, changes, manifest.Rules(changes))
			writeReport, werr := deps.Runner.Run(ctx, plan.Plan{
				Name:  "record " + req.Run + "'s failure count",
				Steps: writeSteps,
			}, plan.ModeExecute)
			if werr != nil {
				return nil, false, "", RefuseUnrepairable("%s", werr)
			}
			if _, _, derr := document(writeSteps, writeReport); derr != nil {
				return nil, false, "", derr
			}
			invocation = argvOf(mustRecord(writeReport, writeSteps.Answer()))
		}
	}

	fact := Count("consecutive_dispatch_failures", next, invocation)
	if next >= threshold {
		return []Fact{fact}, true, fmt.Sprintf(
			"the run is stopped: %d consecutive dispatch(es) returned with no bead completed (threshold %d)",
			next, threshold), nil
	}
	return []Fact{fact}, false, "", nil
}

// queueEntry is one bead in an agent's own queue, rendered as
// [Structured] JSON: the id and whether it is ready to start or already
// claimed.
type queueEntry struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

const (
	queueStateReady      = "ready"
	queueStateInProgress = "in_progress"
)

// reportAgentQueue narrows the dispatch set to one agent's own queue —
// the AQ delta. It reuses [tracker.AgentQueue] for the drain derivation
// rather than restating the two-query join it performs, and separately
// walks the run's whole descendant set to tell a dispatch nobody ever
// gave a bead to from one that finished everything it was given.
// [tracker.AgentQueue] deliberately leaves that distinction to its
// caller — its own doc says telling the two apart "needs the run's whole
// assignment set", which is not either half a queue is built from.
func reportAgentQueue(ctx context.Context, deps status.Deps, req Request, agent string) (Result, error) {
	beneath, dRecords, err := tracker.Descendants(ctx, deps.Tracker, deps.Runner, req.Run)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	established := false
	for _, b := range beneath {
		if b.Assignee == agent {
			established = true
			break
		}
	}

	queue, qRecords, err := tracker.AgentQueue(ctx, deps.Tracker, deps.Runner, req.Run, agent, tracker.ReadyFilter{
		ExcludeKinds:  []string{epicKind},
		ExcludeLabels: []string{humanLabel},
	})
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	result := Result{Outcome: OutcomeClear}
	result.Records = append(result.Records, dRecords...)
	result.Records = append(result.Records, qRecords...)

	queueInv := invocationsOf(qRecords)
	establishedInv := invocationsOf(dRecords)

	// AC-SWARM-DISP-002.12: the ready half keeps the order the tracker
	// answered it in — its own dependency-satisfiability order — rather
	// than the alphabetical sort the whole-run set uses; re-sorting here
	// would throw that order away.
	entries := make([]queueEntry, 0, len(queue.Ready)+len(queue.InProgress))
	for _, b := range queue.Ready {
		entries = append(entries, queueEntry{ID: b.ID, State: queueStateReady})
	}
	for _, b := range queue.InProgress {
		entries = append(entries, queueEntry{ID: b.ID, State: queueStateInProgress})
	}

	result.Facts = append(result.Facts,
		Flag("established", established, establishedInv...),
		Structured("queue", marshal(entries), queueInv...),
		Count("queue_count", len(entries), queueInv...),
		Flag("drained", queue.Drained(), queueInv...))

	switch {
	case !established:
		// AC-SWARM-DISP-002.15: never given anything is a dispatch error,
		// not a completion — reported distinctly from drained so a
		// coordinator does not read a mistyped dispatch id as a finished
		// agent.
		result.Outcome = OutcomeAttention
		result.Resolution = fmt.Sprintf(
			"assign a bead in %s to %s, or dispatch a different agent id", req.Run, agent)
		result.Summary = fmt.Sprintf(
			"%s is not an established dispatch in %s: no bead anywhere in the run names it as assignee",
			agent, req.Run)
	case queue.Drained():
		result.Summary = fmt.Sprintf("%s has drained: nothing ready and nothing in progress in %s", agent, req.Run)
	default:
		result.Summary = fmt.Sprintf("%s holds %d bead(s) under %s: %d ready, %d in progress",
			agent, len(entries), req.Run, len(queue.Ready), len(queue.InProgress))
	}
	return result, nil
}

// invocationsOf reads every record's resolved invocation, skipping a step
// that never ran — so a fact cites every query that reached its answer
// rather than just the fragment's last one, and a reader can re-run any
// of them.
func invocationsOf(records []plan.Record) [][]string {
	out := make([][]string, 0, len(records))
	for _, r := range records {
		if argv := argvOf(r); argv != nil {
			out = append(out, argv)
		}
	}
	return out
}
