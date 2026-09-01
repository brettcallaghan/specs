// The `swarm phase close` command: closes the epics beneath a run epic
// that the tracker judges eligible to close, and reports the beads that
// closure released as ready. Per the "swarm phase close" delta of
// [[container-swarm-cli]] and FEAT-SWARM-DISP-001 of
// specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// The tracker's own epic-closing command ignores its positional argument
// and closes every eligible epic in the database at exit 0
// (`bd epic status --eligible-only --json`, measured against 12 rows
// with or without one — see [tracker.Tracker.ClosableEpics]). This
// command exists so that a scope is an argument the plan carries, and so
// that closing nothing is a reported outcome rather than a silence.
//
// It does not go through [Spec] and [newSubcommand], for the same reason
// [statusCmd] does not: the scope this command may act on is not known
// until the run's descendants have been walked, so there is no single
// fixed plan to hand [newSubcommand] ahead of running it. See
// [tracker.Descendants]'s own doc comment.
package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// epicKind is the tracker's own type for an epic. An epic is not work
// this command closes for its own sake — the run's descendant epics are.
const epicKind = "epic"

// phaseCmd builds the `phase` noun.
func phaseCmd(streams *ioStreams, tr tracker.Tracker, runner swarmexec.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "phase",
		Short:        "Advance a run's phase epics",
		SilenceUsage: true,
	}
	cmd.AddCommand(phaseCloseCmd(streams, tr, runner))
	return cmd
}

// phaseCloseCmd builds `swarm phase close`.
func phaseCloseCmd(streams *ioStreams, tr tracker.Tracker, runner swarmexec.Runner) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
	)

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close this run's finished phase epics",
		Long: "Closes the epics beneath the named run epic that the tracker judges eligible " +
			"to close — open, with every child closed — and reports the beads that closure " +
			"released as ready. An epic outside the run is never touched and never reported, " +
			"and eligibility is never re-derived here: it is the tracker's own rule, only " +
			"narrowed to this run's own descendants.",
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
				return finish(streams, req, explainPhaseClose(req))
			}

			result, err := closePhases(context.Background(), tr, runner, req)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe the invocations this command would run, and run none of them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"perform every read this command would perform, and close nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")

	return cmd
}

// explainPhaseClose describes the measurements and writes this command
// would take without taking any.
//
// Not a [plan.Plan]: which epics it may close, and which of those have an
// open child worth naming, is not known until the run's descendants have
// been walked and read back, so what is printed is a description of the
// sequence rather than a plan explain cannot render truthfully — the same
// reasoning [explainStatus] gives.
func explainPhaseClose(req Request) Result {
	lines := []string{
		"swarm phase close derives its scope before it can name a fixed sequence, so its " +
			"steps are described here rather than printed as a plan:",
		"  bd list --parent <epic> --all --limit 0 --json      (one query per level below " +
			req.Run + "; the depth is not known until each level answers)",
		"  bd ready --parent " + req.Run + " --exclude-type epic --json  (the ready set, before closing anything)",
		"  bd epic status --eligible-only --json                (every epic the tracker judges eligible, database-wide; narrowed to " +
			req.Run + "'s own descendants)",
		"  bd list --parent <epic> --all --limit 0 --json      (for an epic left open, its children — to name the open one as the reason)",
		"  bd update <epic> --status closed --json              (one per epic found eligible, in the order the walk found them)",
		"  bd ready --parent " + req.Run + " --exclude-type epic --json  (the ready set again, to name what a closure released)",
	}
	return Result{
		Command: req.Command, Run: req.Run, Mode: req.Mode, Outcome: OutcomeClear,
		Summary: "explained; nothing was measured or closed", Transcript: lines,
	}
}

// closePhases is the impure half: it runs the walk, the eligibility
// query, the per-epic reason queries and the closing writes, and turns
// what came back into the result. Every write step goes through pr.Run
// with req.Mode, so a dry run performs every read here and refuses every
// close — [plan.Runner]'s own guarantee, not a second one built by hand.
func closePhases(ctx context.Context, tr tracker.Tracker, runner swarmexec.Runner, req Request) (Result, error) {
	pr := plan.Runner{Exec: runner}
	var records []plan.Record

	descendants, recs, err := tracker.Descendants(ctx, tr, pr, req.Run)
	records = append(records, recs...)
	if err != nil {
		return Result{}, RefuseUnrepairable("the run's beads could not be walked: %s", err)
	}

	before, beforeInv, recs, err := readyBeads(ctx, pr, tr, req.Mode, "ready-before", req.Run)
	records = append(records, recs...)
	if err != nil {
		return Result{}, RefuseUnrepairable("the ready set could not be read before closing: %s", err)
	}

	closableStep := tr.ClosableEpics("closable", req.Run)
	closableReport, rerr := pr.Run(ctx, plan.Plan{
		Name:  "the epics eligible to close under " + req.Run,
		Steps: []plan.Step{closableStep},
	}, req.Mode)
	records = append(records, closableReport.Records...)
	if rerr != nil {
		return Result{}, RefuseUnrepairable("the eligibility query could not run: %s", rerr)
	}
	closableRec, ok := closableReport.Record("closable")
	if !ok || closableRec.Failed() {
		// AC-SWARM-DISP-001.9: a failed eligibility query is reported as
		// unanswered, never as "nothing eligible".
		return Result{}, RefuseUnrepairable(
			"the eligibility query failed, so eligibility under %s is unanswered: %s",
			req.Run, describeFailure(closableRec))
	}
	closableAnswer, err := tracker.ClosableFrom(closableRec.Value)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}
	closableInv := argvOf(closableRec)

	epics := epicsOf(descendants)
	eligible := closableAnswer.Within(descendants)
	eligibleIDs := make(map[string]bool, len(eligible))
	for _, e := range eligible {
		eligibleIDs[e.Epic.ID] = true
	}

	reasons, reasonInvs, recs, err := reasonsFor(ctx, pr, tr, req.Mode, openEpicsNotEligible(epics, eligibleIDs))
	records = append(records, recs...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	closeInvs, recs, err := closeEligible(ctx, pr, tr, req.Mode, eligible)
	records = append(records, recs...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	after, afterInv, recs, err := readyBeads(ctx, pr, tr, req.Mode, "ready-after", req.Run)
	records = append(records, recs...)
	if err != nil {
		return Result{}, RefuseUnrepairable("the ready set could not be read after closing: %s", err)
	}
	released := newlyReady(before, after)

	closedIDs := make([]string, 0, len(eligible))
	for _, e := range eligible {
		closedIDs = append(closedIDs, e.Epic.ID)
	}
	sort.Strings(closedIDs)

	result := Result{Outcome: OutcomeClear, Records: records}
	result.Facts = append(result.Facts,
		List("closed_epics", closedIDs, append([][]string{closableInv}, closeInvs...)...))

	leftOpenInv := reasonInvs
	if len(leftOpenInv) == 0 {
		leftOpenInv = [][]string{closableInv}
	}
	result.Facts = append(result.Facts, Structured("left_open", leftOpenDocument(reasons), leftOpenInv...))
	result.Facts = append(result.Facts, List("released_ready", released, beforeInv, afterInv))

	switch {
	case len(eligible) == 0:
		result.Summary = fmt.Sprintf("no epic in %s is eligible to close", req.Run)
	case req.Mode == plan.ModeDryRun:
		result.Summary = fmt.Sprintf("would close %d epic(s): %s; nothing was written",
			len(eligible), strings.Join(closedIDs, ", "))
	default:
		result.Summary = fmt.Sprintf("closed %d epic(s): %s", len(eligible), strings.Join(closedIDs, ", "))
		if len(released) > 0 {
			result.Summary += fmt.Sprintf("; released %d bead(s) ready: %s",
				len(released), strings.Join(released, ", "))
		}
	}

	return result, nil
}

// readyBeads reads the ready set under run, excluding epics — the same
// filter [measureReadySet] in status.go uses, since an epic is never
// work a dispatch takes.
func readyBeads(ctx context.Context, pr plan.Runner, tr tracker.Tracker, mode plan.Mode, name, run string) (
	[]tracker.Bead, []string, []plan.Record, error,
) {
	step := tr.Ready(name, run, tracker.ReadyFilter{ExcludeKinds: []string{epicKind}})
	report, err := pr.Run(ctx, plan.Plan{Name: "the ready set under " + run, Steps: []plan.Step{step}}, mode)
	if err != nil {
		return nil, nil, report.Records, fmt.Errorf("reading the ready set: %w", err)
	}
	rec, ok := report.Record(name)
	if !ok || rec.Failed() {
		return nil, nil, report.Records, fmt.Errorf("the ready set could not be read: %s", describeFailure(rec))
	}
	beads, err := tracker.BeadsFrom(rec.Value)
	if err != nil {
		return nil, nil, report.Records, err
	}
	return beads, argvOf(rec), report.Records, nil
}

// openReason is one epic under scope that stayed open, and why.
type openReason struct {
	Epic      string
	OpenChild string
}

// reasonsFor reads each ineligible epic's children, so the open one can
// be named as the reason it stays open (AC-SWARM-DISP-001.6). One plan,
// one query per epic — a second round after eligibility is known, the
// same shape [tracker.Descendants] takes for its own per-level queries.
func reasonsFor(ctx context.Context, pr plan.Runner, tr tracker.Tracker, mode plan.Mode, epics []tracker.Bead) (
	[]openReason, [][]string, []plan.Record, error,
) {
	if len(epics) == 0 {
		return nil, nil, nil, nil
	}

	steps := make([]plan.Step, 0, len(epics))
	for i, b := range epics {
		steps = append(steps, tr.Children(fmt.Sprintf("reason-%d", i), b.ID))
	}
	report, err := pr.Run(ctx, plan.Plan{Name: "why an epic under scope stays open", Steps: steps}, mode)
	if err != nil {
		return nil, nil, report.Records, fmt.Errorf("checking why an epic stays open: %w", err)
	}

	var reasons []openReason
	var invs [][]string
	for i, b := range epics {
		rec, ok := report.Record(fmt.Sprintf("reason-%d", i))
		if !ok || rec.Failed() {
			return nil, nil, report.Records, fmt.Errorf(
				"could not read %s's children to name why it stays open: %s", b.ID, describeFailure(rec))
		}
		if argv := argvOf(rec); argv != nil {
			invs = append(invs, argv)
		}
		children, err := tracker.BeadsFrom(rec.Value)
		if err != nil {
			return nil, nil, report.Records, fmt.Errorf("%s: %w", b.ID, err)
		}
		reasons = append(reasons, openReason{Epic: b.ID, OpenChild: firstOpenChild(children)})
	}
	return reasons, invs, report.Records, nil
}

// firstOpenChild returns the lowest-id child not counted as done, empty
// when every child is — which the caller reads as "the tracker's
// eligibility answer disagrees with what its own children show".
func firstOpenChild(children []tracker.Bead) string {
	var ids []string
	for _, c := range children {
		if c.State != tracker.StateDone {
			ids = append(ids, c.ID)
		}
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

// closeEligible closes every epic the eligibility query returned, one
// write step each. A single epic's write failing does not withhold the
// others — [plan.Continue] — since each is an independent close and one
// tracker error is not evidence the rest would fail too.
func closeEligible(ctx context.Context, pr plan.Runner, tr tracker.Tracker, mode plan.Mode, eligible []tracker.EpicClosure) (
	[][]string, []plan.Record, error,
) {
	if len(eligible) == 0 {
		return nil, nil, nil
	}

	steps := make([]plan.Step, 0, len(eligible))
	for i, e := range eligible {
		s := tr.SetState(fmt.Sprintf("close-%d", i), e.Epic.ID, tracker.StateDone)
		s.OnFailure = plan.Continue
		steps = append(steps, s)
	}
	report, err := pr.Run(ctx, plan.Plan{Name: "close the run's finished phase epics", Steps: steps}, mode)
	if err != nil {
		return nil, report.Records, fmt.Errorf("closing the run's finished phase epics: %w", err)
	}

	var invs [][]string
	for i, e := range eligible {
		rec, ok := report.Record(fmt.Sprintf("close-%d", i))
		if !ok || rec.Status != plan.StepRan {
			// Refused (dry run) or skipped: nothing closed, nothing to cite.
			continue
		}
		if rec.Failed() {
			return nil, report.Records, fmt.Errorf("closing %s failed: %s", e.Epic.ID, describeFailure(rec))
		}
		if argv := argvOf(rec); argv != nil {
			invs = append(invs, argv)
		}
	}
	return invs, report.Records, nil
}

// newlyReady returns the ids present in after and not in before, sorted
// — the beads a closure released, AC-SWARM-DISP-001.8.
func newlyReady(before, after []tracker.Bead) []string {
	prior := make(map[string]bool, len(before))
	for _, b := range before {
		prior[b.ID] = true
	}
	var out []string
	for _, b := range after {
		if !prior[b.ID] {
			out = append(out, b.ID)
		}
	}
	sort.Strings(out)
	return out
}

// epicsOf narrows a bead set to epics, sorted by id so the walk order
// does not decide report order.
func epicsOf(beads []tracker.Bead) []tracker.Bead {
	var out []tracker.Bead
	for _, b := range beads {
		if b.Kind == epicKind {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// openEpicsNotEligible is the run's own epics that are neither already
// closed nor eligible — the ones AC-SWARM-DISP-001.6 asks a reason for.
func openEpicsNotEligible(epics []tracker.Bead, eligibleIDs map[string]bool) []tracker.Bead {
	var out []tracker.Bead
	for _, b := range epics {
		if b.State == tracker.StateDone || eligibleIDs[b.ID] {
			continue
		}
		out = append(out, b)
	}
	return out
}

// leftOpenEntry is one epic under scope that stayed open, rendered as
// [Structured] JSON.
type leftOpenEntry struct {
	Epic      string `json:"epic"`
	OpenChild string `json:"open_child,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func leftOpenDocument(reasons []openReason) string {
	entries := make([]leftOpenEntry, 0, len(reasons))
	for _, r := range reasons {
		e := leftOpenEntry{Epic: r.Epic, OpenChild: r.OpenChild}
		if e.OpenChild == "" {
			e.Reason = "the tracker does not report this epic eligible, and no open child was found among its own"
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Epic < entries[j].Epic })
	// marshal is defined in status.go, in this same package.
	return marshal(entries)
}

// argvOf reads a record's resolved invocation as an argv, nil for a step
// that never ran (refused by a dry run, or skipped).
func argvOf(rec plan.Record) []string {
	if rec.Invocation.Name == "" {
		return nil
	}
	return append([]string{rec.Invocation.Name}, rec.Invocation.Args...)
}

// describeFailure is why a record counts as [plan.Record.Failed] — its
// own error when it has one, its exit status otherwise.
func describeFailure(rec plan.Record) string {
	if rec.Err != nil {
		return rec.Err.Error()
	}
	return fmt.Sprintf("exit status %d", rec.ExitStatus)
}
