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

			// Built once, ahead of the fork: --agent narrows WHICH beads
			// are reported, and a narrower answer is no reason to discard
			// the dispatch return the same invocation carried. The
			// blueprint states the breaker on the command and calls
			// --agent "the same query with one more filter".
			report := breakerReport{Given: cmd.Flags().Changed("completed"), Completed: completed}

			var (
				result Result
				err    error
			)
			if agent != "" {
				result, err = reportAgentQueue(context.Background(), deps, req, agent, report, threshold)
			} else {
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
				"two more queries against it and reads the circuit breaker's count, so its steps are " +
				"described here rather than printed as a plan:",
			"  one bd list --parent <bead> --all --limit 0 --json per level below " + req.Run +
				"  (every bead in the run, so a dispatch never given anything can be told from one that finished)",
			"  bd ready --parent " + req.Run + " --assignee " + agent +
				" --exclude-type epic --exclude-label human --limit 0 --json  (this dispatch's ready half)",
			"  bd list --assignee " + agent + " --status in_progress --exclude-type " + epicKind +
				" --exclude-label " + humanLabel + " --limit 0 --json  (this dispatch's held half, " +
				"under the same exclusions as the ready half, narrowed to " + req.Run + ")",
			"  bd show " + req.Run + " --json                              (the failure count; --completed also writes it back)",
		}
	}
	return req.stamp(Result{
		Outcome: OutcomeClear,
		Summary: "explained; nothing was measured", Transcript: lines,
	})
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
// reporting a dispatch return, and the count that stops the run. Both
// arms carry them, through the one [applyBreaker] — --agent narrows
// which beads are reported and changes nothing about the dispatch return
// the same invocation reported. Where the two arms differ is in what a
// stop supersedes, and that is stated on [reportAgentQueue], which is
// the arm it is a property of.
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

	result := req.stamp(Result{Outcome: OutcomeClear})
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

	breaking, err := applyBreaker(ctx, deps, req, breaker, threshold)
	result.Records = append(result.Records, breaking.Records...)
	if err != nil {
		return breakerUnread(result, breaking, err), nil
	}
	result.Facts = append(result.Facts, breaking.Facts...)
	if breaking.Stopped {
		// The breaker overrides whatever the dispatch set itself
		// concluded: a stopped run has nothing left to dispatch, whether
		// or not beads remain in the ready set.
		result.Outcome = OutcomeStop
		result.Summary = breaking.Summary
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
//
// Both plans run under req.Mode, so the write this command's only write
// is refused in a dry run rather than performed — which is what makes
// --dry-run's own help, "it writes nothing either way", true. The mode is
// threaded rather than the write being skipped behind a guard, because a
// refused STEP is what a --full-result consumer reads to see that the
// write was declined; a skipped one leaves no record at all.
func applyBreaker(ctx context.Context, deps status.Deps, req Request, report breakerReport, threshold int) (breakerOutcome, error) {
	if deps.Store == nil {
		return breakerOutcome{}, RefuseUnrepairable("no manifest store is configured")
	}

	var records []plan.Record

	readSteps := deps.Store.Read("failure-count", req.Run)
	readReport, err := deps.Runner.Run(ctx, plan.Plan{
		Name:  "read " + req.Run + "'s failure count",
		Steps: readSteps,
	}, req.Mode)
	records = append(records, readReport.Records...)
	if err != nil {
		return breakerOutcome{Records: records}, RefuseUnrepairable("the manifest could not be read: %s", err)
	}
	doc, from, err := document(readSteps, readReport)
	if err != nil {
		return breakerOutcome{Records: records}, err
	}

	var recorded int
	if present, decErr := doc.Decode("consecutive_dispatch_failures", &recorded); decErr != nil {
		return breakerOutcome{Records: records},
			RefuseUnrepairable("consecutive_dispatch_failures does not decode: %s", decErr)
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
				return breakerOutcome{Records: records},
					RefuseUnrepairable("%d does not encode as JSON: %s", next, merr)
			}
			changes := map[string]json.RawMessage{"consecutive_dispatch_failures": json.RawMessage(nextJSON)}
			writeSteps := deps.Store.Write("advance-failure-count", req.Run, changes, manifest.Rules(changes))
			writeReport, werr := deps.Runner.Run(ctx, plan.Plan{
				Name:  "record " + req.Run + "'s failure count",
				Steps: writeSteps,
			}, req.Mode)
			records = append(records, writeReport.Records...)
			if werr != nil {
				return breakerOutcome{Records: records}, RefuseUnrepairable("%s", werr)
			}

			// A dry run performs the guard read and refuses the write, so
			// there is no post-write count: the value the write WOULD have
			// set is a projection, and reporting it as the count would be
			// the same class of untruth as the write itself. The fact is
			// reported unestablished, citing the read that established
			// what the count stands at, and the projection goes in the
			// reason where nothing can mistake it for a measurement.
			if refused(writeReport, writeSteps) {
				refusal := breakerOutcome{
					Facts: []Fact{Unestablished("consecutive_dispatch_failures", fmt.Sprintf(
						"the write was refused: %s. The count stands at %d and would have become %d "+
							"(threshold %d)", refusalReason(writeReport, writeSteps), recorded, next, threshold),
						from)},
					Records: records,
				}
				// A dry run may not INVENT a stop from a projection; it
				// must not HIDE one the manifest already carries — that
				// is the breaker failing open in the mode meant to be
				// safe to rehearse. Which of the two applies is decided
				// by the DIRECTION the report goes, not by the refused
				// write:
				//
				//   --completed 0 ADVANCES the count. The advance is a
				//   projection nothing wrote, so the verdict rests on
				//   recorded — a run the manifest already has at the
				//   threshold is still reported stopped, and one below it
				//   stays clear.
				//
				//   --completed N>0 RESETS it to zero. The reset is not a
				//   projection: it is fixed by the report itself, which
				//   the caller supplied and which execute mode applies
				//   unconditionally. Holding the verdict on recorded here
				//   reports a stop execute would never produce, and puts
				//   "would have become 0" in one result beside a stop
				//   summary.
				//
				// So the verdict follows the same value execute would
				// compare, whenever that value is one the report settles
				// rather than one the write would have invented.
				verdict := recorded
				if report.Completed > 0 {
					verdict = next
				}
				if verdict >= threshold {
					refusal.Stopped = true
					refusal.Summary = stoppedSummary(verdict, threshold)
				}
				return refusal, nil
			}

			if _, _, derr := document(writeSteps, writeReport); derr != nil {
				return breakerOutcome{Records: records}, derr
			}
			invocation = argvOf(mustRecord(writeReport, writeSteps.Answer()))
		}
	}

	outcome := breakerOutcome{
		Facts:   []Fact{Count("consecutive_dispatch_failures", next, invocation)},
		Records: records,
	}
	if next >= threshold {
		outcome.Stopped = true
		outcome.Summary = stoppedSummary(next, threshold)
	}
	return outcome, nil
}

// stoppedSummary is what the breaker says when the run is stopped, from
// the count the verdict rests on — the value a write established, or the
// recorded value when the write was refused. One function because the two
// paths must not drift into two sentences.
func stoppedSummary(count, threshold int) string {
	return fmt.Sprintf(
		"the run is stopped: %d consecutive dispatch(es) returned with no bead completed (threshold %d)",
		count, threshold)
}

// breakerOutcome is what one pass of the circuit breaker established: the
// count it reports, the steps it ran or refused, and whether the run is
// stopped.
//
// A struct rather than a tuple because the records are only reportable if
// they reach the result, and a fifth positional return is how a caller
// drops them without noticing.
type breakerOutcome struct {
	// Facts is the count, established from the read or from the write
	// that advanced it — or unestablished when the write was refused.
	Facts []Fact
	// Records are the breaker's own steps, so a refused write is visible
	// to a --full-result consumer rather than only to this function.
	Records []plan.Record
	// Stopped is whether the run is stopped, which is a statement about
	// the count that WAS recorded — never about the one a write would
	// have set. A refused write therefore withholds only the PROJECTED
	// advance: a run the manifest already records at the threshold is
	// still reported stopped, while a run below it stays clear because
	// the number that would have crossed it is a number nothing wrote.
	Stopped bool
	// Summary is what to say when Stopped, and empty otherwise.
	Summary string
}

// breakerUnread folds a breaker refusal into the answer the invocation
// had ALREADY established, instead of discarding it.
//
// [applyBreaker] refuses three ways — no manifest store is configured,
// the manifest read failed, and consecutive_dispatch_failures does not
// decode — and it is right to refuse each. What must not happen is the
// CALLER throwing away a queue or ready set that its own queries already
// answered: a dispatched worker runs `swarm ready --agent` to learn what
// it holds, and a manifest fault is not that worker's concern.
//
// The outcome becomes attention: not clear, which would assert the run is
// dispatchable when the breaker's verdict is unknown, and not stop, which
// nothing established.
//
// The count is reported unestablished, citing the breaker's own steps —
// which is also what makes [breakerOutcome.Records]' promise, that a
// refused step reaches a --full-result consumer, true on this path. It is
// omitted only when there are no steps to cite, since a store that was
// never configured ran nothing and [Result.Validate] refuses a fact
// naming no invocation; the reason still reaches the reader through the
// resolution, which every path sets.
//
// An existing resolution is kept and appended to rather than replaced: a
// dispatch id no bead names and a breaker that could not be read are two
// unsatisfied properties, and answering one of them erases the other.
func breakerUnread(result Result, breaking breakerOutcome, err error) Result {
	if inv := invocationsOf(breaking.Records); len(inv) > 0 {
		result.Facts = append(result.Facts,
			Unestablished("consecutive_dispatch_failures", err.Error(), inv...))
	}
	result.Outcome = OutcomeAttention

	resolution := fmt.Sprintf(
		"the circuit breaker could not be read (%s), so the run's stop verdict is unknown; "+
			"repair the run epic's manifest and re-run — the answer above stands, it is the "+
			"count and the stop verdict that are missing", err)
	if result.Resolution == "" {
		result.Resolution = resolution
	} else {
		result.Resolution += "; " + resolution
	}
	return result
}

// refusalReason is the reason the plan runner gave for refusing a
// fragment's answer step, for a caller that has already established the
// step was refused.
func refusalReason(report plan.Report, steps plan.Steps) string {
	if record, ok := report.Record(steps.Answer()); ok && record.Reason != "" {
		return record.Reason
	}
	return "the step writes"
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
// It carries the circuit breaker exactly as the whole-run arm does: the
// filter narrows which beads are reported and changes nothing about the
// dispatch return the same invocation reported.
//
// The breaker's OVERRIDE is where the two arms differ, and the difference
// is in this arm's own answer rather than in the breaker. This one can
// already have concluded the attention outcome before the breaker runs —
// a dispatch id no bead in the run names — and that conclusion carries a
// resolution. A stop supersedes it, so the override both carries the
// unestablished id forward into the summary it replaces and clears the
// resolution, which would otherwise name a repair that does not lift the
// stop. [reportReady] sets no resolution at all before its own override,
// so it has none to supersede and needs no counterpart.
func reportAgentQueue(
	ctx context.Context, deps status.Deps, req Request, agent string,
	breaker breakerReport, threshold int,
) (Result, error) {
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

	result := req.stamp(Result{Outcome: OutcomeClear})
	result.Records = append(result.Records, dRecords...)
	result.Records = append(result.Records, qRecords...)

	establishedInv, err := invocationsRequired("the beads under "+req.Run, dRecords)
	if err != nil {
		return Result{}, err
	}
	queueInv, err := invocationsRequired(agent+"'s queue under "+req.Run, qRecords)
	if err != nil {
		return Result{}, err
	}

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

	breaking, err := applyBreaker(ctx, deps, req, breaker, threshold)
	result.Records = append(result.Records, breaking.Records...)
	if err != nil {
		return breakerUnread(result, breaking, err), nil
	}
	result.Facts = append(result.Facts, breaking.Facts...)
	if breaking.Stopped {
		// Same override as the whole-run arm: a stopped run has nothing
		// left to dispatch, and a coordinator reading only this
		// dispatch's queue is the one most likely to send more work into
		// it. The queue's own facts are still reported beside it.
		result.Outcome = OutcomeStop
		result.Summary = breaking.Summary
		if !established {
			// The stop replaces the summary, and the mistyped-dispatch
			// signal must survive that replacement: a dispatch id no
			// bead names may never read as anything else, and the
			// established fact alone is not what a caller reads first.
			result.Summary = fmt.Sprintf("%s; %s is not an established dispatch in %s",
				breaking.Summary, agent, req.Run)
		}
		// No resolution un-stops a run, so the one belonging to the
		// condition this stop replaced is cleared rather than left
		// pointing at a repair that would not lift the stop. This is the
		// half [reportReady] needs no counterpart for: the whole-run arm
		// sets no resolution before its own override, so it has none to
		// supersede.
		result.Resolution = ""
	}
	return result, nil
}

// invocationsOf reads every record's resolved invocation, skipping a step
// that never ran — so a fact cites every query that reached its answer
// rather than just the fragment's last one, and a reader can re-run any
// of them.
// invocationsRequired is [invocationsOf] with the guarantee every fact
// built from the answer needs: at least one non-empty invocation, or a
// refusal naming what could not be cited.
//
// The three call sites take their records from [tracker.AgentQueue] and
// [tracker.Descendants], which answer with records but not with the steps
// that produced them — so where no record ran there is nothing to cite
// and nothing to invent, and the read never reached its answer. Reporting
// a fact from it would be the silent success this container exists to
// remove; the alternative, widening the two tracker helpers to also
// return their steps, puts a change into three packages for a path
// nothing reaches today.
//
// what names the read, and becomes the subject of the refusal.
func invocationsRequired(what string, records []plan.Record) ([][]string, error) {
	inv := invocationsOf(records)
	if len(inv) == 0 {
		return nil, RefuseUnrepairable(
			"%s names no invocation a reader could re-run, so its answer is not reported", what)
	}
	return inv, nil
}

func invocationsOf(records []plan.Record) [][]string {
	out := make([][]string, 0, len(records))
	for _, r := range records {
		if argv := argvOf(r); argv != nil {
			out = append(out, argv)
		}
	}
	return out
}
