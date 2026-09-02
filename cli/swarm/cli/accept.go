// The `swarm dispatch accept` command, first half: the drain check, the
// serialized-or-concurrent case, and the merge. Per the "swarm dispatch
// accept" delta of [[container-swarm-cli]] (added — drain, case and
// merge) and FEAT-SWARM-DISP-004 of
// specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// The whole reason the case is named before the refusal is read: a
// refused fast-forward means the dispatch was cut from an earlier target
// and must be rebased, OR it means the target moved between the case
// being read and the fold being attempted — and the two call for
// opposite actions. Reading the refusal without the case is how a
// healthy run gets stopped and a raced one gets rewritten.
//
// Hand-declared rather than built through [Spec] and [newSubcommand],
// for the same reason [dispatchOpenCmd] is: which steps this command
// takes depends on what the drain query and the case comparison answer,
// so there is no single fixed plan to hand [newSubcommand] ahead of
// running it.
package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// The two cases a fast-forward refusal is read against, in the run's own
// words. They are reported as a fact so a caller can see which reading
// the command took rather than infer it from what the command then did.
const (
	caseSerialized = "serialized"
	caseConcurrent = "concurrent"
)

// dispatchAcceptCmd builds `swarm dispatch accept`.
func dispatchAcceptCmd(
	streams *ioStreams, store manifest.Store, tr tracker.Tracker,
	vc vcs.VersionControl, runner swarmexec.Runner,
) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
		dispatchID string
		noPush     bool
	)

	cmd := &cobra.Command{
		Use:   "accept",
		Short: "Check a returned dispatch has drained, name its case, and merge it",
		Long: "Checks that the dispatch holds nothing ready and nothing in progress before " +
			"anything else, and merges nothing when it does. Decides whether the dispatch was " +
			"serialized or concurrent by comparing the dispatch branch's merge-base with the " +
			"target against the target's current head — equal is serialized, anything else " +
			"concurrent — and reports which case it read before it reads a fast-forward " +
			"refusal. A concurrent branch is rebased onto the target as it stands at that " +
			"moment and then fast-forwarded; a serialized one whose fast-forward is refused " +
			"stops, and the merge is never re-attempted. A rebase that conflicts records a " +
			"sequencing finding naming both beads and every conflicting path, marks the " +
			"unmerged bead blocked, and resolves nothing itself. " +
			"The worktree is removed only when the dispatch branch is an ancestor of the target, " +
			"and its branch is deleted only after the removal succeeded; a removal the tool refuses " +
			"is never retried and never forced. Every bead is then re-checked with no signal " +
			"deferred, and the sprint branch is pushed and proved landed only once every one of " +
			"them has passed.",
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
			dispatch := strings.TrimSpace(dispatchID)
			if dispatch == "" {
				return finish(streams, req, refusal(req, RefuseArguments(
					"%s accepts one dispatch; pass --dispatch <dispatch-id>", req.Command)))
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
				return finish(streams, req, explainDispatchAccept(req, dispatch, noPush))
			}

			result, err := acceptDispatch(context.Background(), store, tr, vc, runner, req, dispatch, noPush)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe the invocations this command would run, and run none of them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"perform every read this command would perform, and merge nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")
	cmd.Flags().StringVar(&dispatchID, "dispatch", "",
		"the dispatch this invocation is accepting")
	cmd.Flags().BoolVar(&noPush, "no-push", false,
		"take every step before the push and report the branch as unlanded instead of pushing it")

	return cmd
}

// explainDispatchAccept describes the measurements and writes this
// command would take without taking any.
//
// Not a [plan.Plan]: whether the dispatch has drained decides whether
// anything after the drain check runs at all, and the case decides
// whether a rebase is in the sequence, so what is printed is a
// description with both arms named rather than a plan explain cannot
// render truthfully — the same reasoning [explainDispatchOpen] gives.
func explainDispatchAccept(req Request, dispatchID string, noPush bool) Result {
	branch := dispatchBranchName(dispatchID)
	lines := []string{
		"swarm dispatch accept derives whether the dispatch has drained, and then which case " +
			"it is in, before it can name a fixed sequence, so its steps are described here " +
			"rather than printed as a plan:",
		"  one bd list --parent <bead> --all --limit 0 --json per level below " + req.Run +
			"   (the run's descendant set, which narrows the held half)",
		"  bd ready --parent " + req.Run + " --assignee " + dispatchID +
			" --exclude-type epic --exclude-label human --limit 0 --json  (the ready half)",
		"  bd list --assignee " + dispatchID + " --status in_progress --exclude-type " + epicKind +
			" --exclude-label " + humanLabel + " --limit 0 --json  (the held half, under the same exclusions)",
		"  # if either half is non-empty: the dispatch has not drained, and nothing below runs",
		"  bd show " + req.Run + " --json                              (the manifest, for the target branch)",
		"  git rev-parse --verify <target>^{commit}            (the target's head, read now)",
		"  git rev-parse --verify HEAD^{commit}                (the repository is on the target)",
		"  git rev-parse --verify " + branch + "^{commit}",
		"  git merge-base --is-ancestor <target> " + branch + "   (equal merge-base and target head: serialized)",
		"  # serialized:",
		"  git merge --ff-only " + branch + "                     (a refusal here stops; it is never re-attempted)",
		"  # concurrent:",
		"  git rebase <target-sha> " + branch + "                 (<target-sha> is the head read above)",
		"  git checkout <target>                               (the rebase left the repository on " + branch + ")",
		"  git merge --ff-only " + branch,
		"  # if the rebase conflicts:",
		"  git log --format=... -n 1 " + branch + "               (the unmerged bead)",
		"  git log --format=... -n 1 <target>                  (the bead on the target it conflicts with)",
		"  bd update <bead> --status blocked --json            (the unmerged bead)",
		"  bd comment <bead> '<the sequencing finding>' --json",
		"  # once the fold has succeeded:",
		"  git merge-base --is-ancestor " + branch + " <target>   (the guard the removal sits behind)",
		"  # if it answers no: the worktree is left standing, no branch is deleted, and the command stops",
		"  git worktree prune && git worktree list --porcelain  (the dispatch's worktree, and whether it is locked)",
		"  git worktree remove <path>                          (never --force; a refusal stops the plan here)",
		"  git branch -d " + branch + "                           (only after the removal succeeded)",
		"  # the post-merge re-check, with no signal deferred:",
		"  one bd list --parent <bead> --all --limit 0 --json per level below " + req.Run,
		"  git log --format=... -n " + fmt.Sprint(acceptanceScan) + " <target>" +
			"   (the beads whose close commits are on the target)",
		"  git worktree prune && git worktree list --porcelain  (nothing stands on " + branch + " any more)",
		"  bd comments <bead> --json                           (one per bead the dispatch holds)",
	}
	if noPush {
		lines = append(lines,
			"  # --no-push: every step above runs and the branch is reported unlanded rather than pushed")
	} else {
		lines = append(lines,
			"  # only once every bead has passed, swarm branch push's own steps:",
			"  git push "+defaultRemote+" <target>",
			"  git fetch "+defaultRemote,
			"  git rev-parse --verify "+defaultRemote+"/<target>^{commit}",
			"  git rev-list --count --left-right <target>..."+defaultRemote+"/<target>")
	}
	return req.stamp(Result{
		Outcome: OutcomeClear,
		Summary: "explained; nothing was measured, merged or recorded", Transcript: lines,
	})
}

// acceptDispatch is the impure half. Every write goes through pr.Run with
// req.Mode, so a dry run performs every read here and merges nothing —
// [plan.Runner]'s own guarantee rather than a second one built by hand.
func acceptDispatch(
	ctx context.Context, store manifest.Store, tr tracker.Tracker,
	vc vcs.VersionControl, runner swarmexec.Runner, req Request, dispatchID string, noPush bool,
) (Result, error) {
	pr := plan.Runner{Exec: runner}
	var records []plan.Record

	// AC-SWARM-DISP-004.8: the drain check comes before anything else,
	// and an agent still holding work stops the acceptance here — nothing
	// below this point runs, so nothing is merged or removed.
	//
	// The derivation is [tracker.AgentQueue]'s, which `swarm ready
	// --agent` already reaches: the blueprint requires ONE
	// implementation of "drained", and a second reading of the two
	// queries here would be a second answer to disagree with it.
	queue, qRecords, err := tracker.AgentQueue(ctx, tr, pr, req.Run, dispatchID, tracker.ReadyFilter{
		ExcludeKinds:  []string{epicKind},
		ExcludeLabels: []string{humanLabel},
	})
	records = append(records, qRecords...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}
	queueInv := invocationsOf(qRecords)
	if !queue.Drained() {
		return notDrained(req, dispatchID, queue, records, queueInv), nil
	}

	target, targetInv, recs, err := acceptTarget(ctx, store, pr, req)
	records = append(records, recs...)
	if err != nil {
		return Result{}, err
	}
	dispatchBranch := dispatchBranchName(dispatchID)

	m, recs, err := readCase(ctx, vc, pr, req, target, dispatchBranch)
	records = append(records, recs...)
	if err != nil {
		return Result{}, err
	}

	result := req.stamp(Result{Outcome: OutcomeClear, Records: records})
	result.Facts = append(result.Facts,
		Flag("drained", true, queueInv...),
		Text("target", target, targetInv),
		Text("dispatch_branch", dispatchBranch, m.branchInv),
		Text("case", m.name(), m.caseInv))

	merged, recs, err := foldIn(ctx, tr, vc, pr, req, target, dispatchBranch, m)
	result.Records = append(result.Records, recs...)
	if err != nil {
		// A stop or an attention carrying facts of its own — the case is
		// among them, which is what AC-SWARM-DISP-004.2 asks for.
		var stopped *acceptStop
		if errors.As(err, &stopped) {
			return acceptStopped(result, stopped), nil
		}
		return Result{}, err
	}

	result.Facts = append(result.Facts, merged.facts...)
	result.Summary = merged.summary

	// The second half. It is reached only because the fold above did NOT
	// return an error, which is what puts the removal behind the merge's
	// result rather than beside it.
	settled, recs, err := settleDispatch(
		ctx, tr, vc, pr, runner, req, target, dispatchID, merged.ran, noPush)
	result.Records = append(result.Records, recs...)
	if err != nil {
		var stopped *acceptStop
		if errors.As(err, &stopped) {
			return acceptStopped(result, stopped), nil
		}
		return Result{}, err
	}
	result.Facts = append(result.Facts, settled.facts...)
	if settled.summary != "" {
		result.Summary = result.Summary + "; " + settled.summary
	}
	if settled.outcome != OutcomeClear {
		result.Outcome = settled.outcome
		result.Resolution = settled.resolution
	}
	return result, nil
}

// acceptStopped folds a refusal that carries facts of its own into the
// result so far, rather than discarding what was already established.
func acceptStopped(result Result, stopped *acceptStop) Result {
	result.Outcome = stopped.refusal.Outcome
	result.Resolution = stopped.refusal.Resolution
	result.Facts = append(result.Facts, stopped.facts...)
	if result.Summary == "" {
		result.Summary = stopped.refusal.Reason
	} else {
		result.Summary = result.Summary + "; " + stopped.refusal.Reason
	}
	return result
}

// notDrained is the answer for a dispatch that still holds work:
// attention, naming every bead it holds, with nothing merged.
func notDrained(req Request, dispatchID string, queue tracker.Queue, records []plan.Record, inv [][]string) Result {
	held := make([]string, 0, len(queue.Ready)+len(queue.InProgress))
	for _, b := range queue.Ready {
		held = append(held, b.ID)
	}
	for _, b := range queue.InProgress {
		held = append(held, b.ID)
	}
	sort.Strings(held)

	return req.stamp(Result{
		Outcome: OutcomeAttention,
		Records: records,
		Facts: []Fact{
			Flag("drained", false, inv...),
			List("holding", held, inv...),
			Count("holding_count", len(held), inv...),
		},
		Summary: fmt.Sprintf("%s has not drained: it still holds %d bead(s) in %s (%s); nothing was merged",
			dispatchID, len(held), req.Run, strings.Join(held, ", ")),
		Resolution: fmt.Sprintf(
			"let %s finish or mark what it holds blocked, then accept it again: swarm ready --run %s --agent %s",
			dispatchID, req.Run, dispatchID),
	})
}

// acceptTarget reads the sprint branch the run records — the branch this
// dispatch merges into. A run with none recorded has no target, which is
// repairable by opening a dispatch rather than by anything here.
func acceptTarget(ctx context.Context, store manifest.Store, pr plan.Runner, req Request) (string, []string, []plan.Record, error) {
	steps := store.Read("manifest", req.Run)
	report, err := pr.Run(ctx, plan.Plan{Name: "read " + req.Run + "'s manifest", Steps: steps}, req.Mode)
	if err != nil {
		return "", nil, report.Records, RefuseUnrepairable("the manifest could not be read: %s", err)
	}
	doc, from, err := document(steps, report)
	if err != nil {
		return "", nil, report.Records, err
	}
	var target string
	if present, decErr := doc.Decode("branch", &target); decErr != nil {
		return "", nil, report.Records, RefuseUnrepairable("the recorded branch does not decode: %s", decErr)
	} else if !present || strings.TrimSpace(target) == "" {
		return "", nil, report.Records, RefuseRepairable(
			fmt.Sprintf("establish it first: swarm dispatch open --run %s", req.Run),
			"%s records no sprint branch, so there is no target to merge into", req.Run)
	}
	return target, from, report.Records, nil
}

// mergeCase is what the case comparison read: the target's head, whether
// the dispatch was serialized, and the invocations behind each.
//
// It deliberately carries no branch sha. The dispatch branch's head IS
// resolved, as an early validity check, but a concurrent dispatch's
// commits are rewritten by the rebase, so the sha read before the fold
// names a commit that need not be reachable from the target at all —
// which is why [foldIn] resolves what landed from the TARGET afterwards.
// A populated field here would be the value the next reader reaches for.
type mergeCase struct {
	targetSHA  string
	serialized bool
	targetInv  []string
	branchInv  []string
	caseInv    []string
}

func (m mergeCase) name() string {
	if m.serialized {
		return caseSerialized
	}
	return caseConcurrent
}

// readCase decides serialized or concurrent, and does it BEFORE any
// fold is attempted (AC-SWARM-DISP-004.2).
//
// AC-SWARM-DISP-004.3: the comparison is the dispatch branch's
// merge-base with the target against the target's current head — equal
// is serialized. `git merge-base --is-ancestor <target> <branch>` IS
// that comparison: it is true exactly when the merge-base of the two is
// the target's own head. It is never inferred from how many worktrees
// are standing, which counts dispatches rather than measuring this one's
// base — nothing here reads the worktree list at all.
//
// The repository's own head is read in the same breath, because
// `merge --ff-only` advances the branch that is CHECKED OUT: a fold run
// from anywhere but the target would move the wrong branch and report
// the target merged.
func readCase(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request,
	target, dispatchBranch string,
) (mergeCase, []plan.Record, error) {
	targetStep := vc.Resolve("target-head", target)
	headStep := vc.Resolve("repo-head", "HEAD")
	branchStep := vc.Resolve("dispatch-head", dispatchBranch)
	caseStep := vc.Descends("merge-case", target, dispatchBranch)

	steps := plan.Steps{}
	for _, fragment := range []vcs.Steps{targetStep, headStep, branchStep, caseStep} {
		steps = append(steps, fragment...)
	}
	report, err := pr.Run(ctx, plan.Plan{
		Name:  "whether " + dispatchBranch + " was cut from " + target + " as it stands",
		Steps: steps,
	}, req.Mode)
	if err != nil {
		return mergeCase{}, report.Records, RefuseUnrepairable("%s", err)
	}

	targetSHA, err := revision(report, targetStep, target)
	if err != nil {
		return mergeCase{}, report.Records, err
	}
	headSHA, err := revision(report, headStep, "HEAD")
	if err != nil {
		return mergeCase{}, report.Records, err
	}
	// Resolved for the check and not for the value: this refuses early
	// when the dispatch branch does not resolve at all, rather than
	// letting that failure surface later inside the fold.
	if _, err := revision(report, branchStep, dispatchBranch); err != nil {
		return mergeCase{}, report.Records, err
	}

	if headSHA != targetSHA {
		return mergeCase{}, report.Records, RefuseRepairable(
			fmt.Sprintf("put the repository on the target first: git checkout %s", target),
			"the repository is at %s and the target %s is at %s, so a fast-forward here would advance a branch other than the target",
			headSHA, target, targetSHA)
	}

	caseRec, ok := report.Record(caseStep.Answer())
	if !ok || caseRec.Failed() {
		return mergeCase{}, report.Records, RefuseUnrepairable(
			"whether %s was cut from %s could not be read: %s",
			dispatchBranch, target, describeFailure(caseRec))
	}
	serialized, ok := caseRec.Value.(bool)
	if !ok {
		return mergeCase{}, report.Records, RefuseUnrepairable(
			"the case comparison produced %T, not a verdict", caseRec.Value)
	}

	return mergeCase{
		targetSHA:  targetSHA,
		serialized: serialized,
		targetInv:  argvOf(mustRecord(report, targetStep.Answer())),
		branchInv:  argvOf(mustRecord(report, branchStep.Answer())),
		caseInv:    argvOf(caseRec),
	}, report.Records, nil
}

// revision reads one resolved commit id out of a report, refusing a
// fragment that did not answer rather than carrying an empty id forward.
func revision(report plan.Report, steps vcs.Steps, what string) (string, error) {
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return "", RefuseUnrepairable("%q could not be resolved: %s", what, describeFailure(rec))
	}
	sha, ok := rec.Value.(string)
	if !ok || sha == "" {
		return "", RefuseUnrepairable("resolving %q produced %T, not a commit id", what, rec.Value)
	}
	return sha, nil
}

// acceptStop is a refusal that still carries facts — the case among
// them, so a caller reading the report can see which reading the
// command took before it stopped.
type acceptStop struct {
	refusal *Refusal
	facts   []Fact
}

func (a *acceptStop) Error() string { return a.refusal.Reason }

// mergeOutcome is what the fold produced.
type mergeOutcome struct {
	facts   []Fact
	summary string
	// ran is whether the fold was actually performed. It is false only
	// when an interpreter REFUSED it — the dry run — since a fold that
	// failed returns an error instead of an outcome.
	ran bool
}

// foldIn performs the merge for the case that was read.
//
// The serialized arm is one fast-forward and nothing else:
// AC-SWARM-DISP-004.5 requires a refusal there to stop WITHOUT the merge
// being re-attempted, and without it being re-attempted with
// fast-forward-only relaxed. That is structural rather than remembered —
// this file declares exactly one fold operation,
// [vcs.VersionControl.MergeFastForward], and never
// [vcs.VersionControl.Merge], so there is no relaxed form for a retry to
// reach for.
//
// The concurrent arm rebases onto the target AS IT STANDS AT MERGE TIME
// — m.targetSHA, resolved in [readCase] moments earlier, rather than the
// branch name, so the fold lands on the commit the case was decided
// against even if the target moves while the rebase runs — then puts the
// repository back on the target and fast-forwards.
func foldIn(
	ctx context.Context, tr tracker.Tracker, vc vcs.VersionControl, pr plan.Runner,
	req Request, target, dispatchBranch string, m mergeCase,
) (mergeOutcome, []plan.Record, error) {
	steps := plan.Steps{}
	if !m.serialized {
		steps = append(steps, vc.Rebase("rebase", dispatchBranch, m.targetSHA)...)
		steps = append(steps, vc.Checkout("return-to-target", target)...)
	}
	ffSteps := vc.MergeFastForward("fast-forward", dispatchBranch)
	steps = append(steps, ffSteps...)

	// What landed is read from the TARGET, after the fold, and never from
	// the dispatch branch: a concurrent dispatch's commits are rewritten
	// by the rebase above, so the branch sha [readCase] resolved before
	// any write names a commit that is not reachable from the target at
	// all. [recheckDispatchBeads] reads the target for the same reason.
	landedSteps := vc.Resolve("merged-head", target)
	steps = append(steps, landedSteps...)

	report, err := pr.Run(ctx, plan.Plan{
		Name:  "fold " + dispatchBranch + " into " + target + " (" + m.name() + ")",
		Steps: steps,
	}, req.Mode)
	records := report.Records
	if err != nil {
		return mergeOutcome{}, records, RefuseUnrepairable("%s", err)
	}

	if !m.serialized {
		rebaseRec, ok := report.Record("rebase")
		if ok && rebaseRec.Failed() {
			// AC-SWARM-DISP-004.4, .7: a conflict is recorded as a
			// sequencing finding, the unmerged bead is blocked, and
			// nothing here resolves or aborts the rebase — the working
			// tree it left is the evidence an agent holding that bead's
			// context is dispatched back into.
			if ref := asVCSRefusal(rebaseRec.Err); ref != nil && len(ref.Conflicts) > 0 {
				facts, recs, cerr := recordConflict(ctx, tr, vc, pr, req, target, dispatchBranch, argvOf(rebaseRec), ref)
				records = append(records, recs...)
				if cerr != nil {
					return mergeOutcome{}, records, cerr
				}
				return mergeOutcome{}, records, &acceptStop{
					refusal: RefuseUnrepairable(
						"rebasing %s onto %s conflicts on %d path(s): %s; the worktree is left standing and nothing was resolved",
						dispatchBranch, target, len(ref.Conflicts), strings.Join(ref.Conflicts, ", ")),
					facts: facts,
				}
			}
			return mergeOutcome{}, records, RefuseUnrepairable(
				"rebasing %s onto %s failed: %s", dispatchBranch, m.targetSHA, describeFailure(rebaseRec))
		}
		if coRec, ok := report.Record("return-to-target"); ok && coRec.Failed() {
			return mergeOutcome{}, records, RefuseUnrepairable(
				"the rebase left the repository on %s and it could not be put back on %s: %s",
				dispatchBranch, target, describeFailure(coRec))
		}
	}

	ffRec, ok := report.Record(ffSteps.Answer())
	if !ok {
		return mergeOutcome{}, records, RefuseUnrepairable(
			"the fold of %s into %s produced no record", dispatchBranch, target)
	}
	if ffRec.Failed() {
		return mergeOutcome{}, records, refusedFastForward(target, dispatchBranch, m, ffRec)
	}

	facts := []Fact{}
	summary := ""
	ran := ffRec.Status == plan.StepRan
	switch {
	case ffRec.Status != plan.StepRan:
		// A dry run refused the fold. Every read above still happened, so
		// the case is reported; nothing was written.
		facts = append(facts, Flag("merged", false, m.caseInv))
		summary = fmt.Sprintf("%s is %s against %s; nothing was merged (%s)",
			dispatchBranch, m.name(), target, req.Mode)
	default:
		facts = append(facts, Flag("merged", true, argvOf(ffRec)))

		// The invocation cited for the sha is the one that ESTABLISHED
		// it — the post-fold resolve — not the pre-fold read of the
		// branch, which established a different sha. It is read through
		// [plan.Report.Record] rather than [mustRecord] because this is
		// the one site where the record may legitimately be absent: an
		// absent record is among the ways the resolve below fails.
		//
		// [invocationOf] rather than [argvOf], because the fact must
		// carry an invocation on EVERY path through this branch. A
		// record that is absent, refused or skipped carries none of its
		// own, and a fact naming no invocation is one [Result.Validate]
		// refuses — which would replace the whole result with a bare
		// refusal, losing the merged fact below after a fast-forward
		// that cannot be undone. The step's own declaration names the
		// argv it would have run with, and every argument the resolve
		// builds is a literal, so it always answers here.
		rec, _ := report.Record(landedSteps.Answer())
		landedInv := [][]string{invocationOf(landedSteps[len(landedSteps)-1], rec)}

		// This read is taken AFTER the fast-forward, which is
		// irreversible. A failure of it therefore degrades the report and
		// never abandons the accept: returning an error here would skip
		// [settleDispatch] entirely — no ancestry guard, no post-merge
		// bead re-check, no worktree removal, no branch deletion, no push
		// — and, being a bare refusal, would carry no merged fact, so a
		// caller could not tell that the fold had landed. The fold DID
		// happen; only the sha is unknown, and an unknown value is
		// reported unestablished rather than dropped.
		landed, lerr := revision(report, landedSteps, target+" after the fold")
		if lerr != nil {
			facts = append(facts, Unestablished("merged_sha", lerr.Error(), landedInv...))
		} else {
			facts = append(facts, Text("merged_sha", landed, landedInv...))
		}
		summary = fmt.Sprintf("%s was %s against %s and fast-forwarded into it",
			dispatchBranch, m.name(), target)
		if !m.serialized {
			summary = fmt.Sprintf("%s was %s against %s, rebased onto %s and fast-forwarded into it",
				dispatchBranch, m.name(), target, m.targetSHA)
		}
	}
	return mergeOutcome{facts: facts, summary: summary, ran: ran}, records, nil
}

// refusedFastForward reads a refused fold against the case that was
// already named — AC-SWARM-DISP-004.2 and .5.
//
// Serialized is the one that stops flat: the branch WAS cut from the
// target as it stood when the case was read, so a refusal means the
// target moved underneath the fold, and re-attempting it would either
// fail the same way or, with fast-forward-only relaxed, write a merge
// nobody asked for. Concurrent reaching here means the same race after a
// rebase that succeeded, and stops for the same reason.
func refusedFastForward(target, dispatchBranch string, m mergeCase, rec plan.Record) error {
	facts := []Fact{Flag("merged", false, argvOf(rec))}
	if ref := asVCSRefusal(rec.Err); ref != nil && ref.NonFastForward {
		if m.serialized {
			return &acceptStop{
				refusal: RefuseUnrepairable(
					"%s was cut from %s as it stood at %s, and the fast-forward into it was still refused — the target moved; the merge was not re-attempted",
					dispatchBranch, target, m.targetSHA),
				facts: facts,
			}
		}
		return &acceptStop{
			refusal: RefuseUnrepairable(
				"%s was rebased onto %s and the fast-forward into %s was still refused — the target moved while the rebase ran; the merge was not re-attempted",
				dispatchBranch, m.targetSHA, target),
			facts: facts,
		}
	}
	return &acceptStop{
		refusal: RefuseUnrepairable("folding %s into %s failed: %s",
			dispatchBranch, target, describeFailure(rec)),
		facts: facts,
	}
}

// recordConflict writes the sequencing finding and blocks the unmerged
// bead — AC-SWARM-DISP-004.4.
//
// Both beads are read from the two branches' own head commits, through
// [status.BeadIDFromCloseSubject], which is the run's one reading of a
// close commit's subject: the dispatch branch's head is the bead that did
// not merge, and the target's head is the commit on the target it
// conflicts with.
//
// The two writes are ordered blocked-then-finding per [[adr-026]]:
// blocking is repeatable and recording a comment is not, so the
// irreversible one goes last and a recovery from a partial plan is a
// re-run rather than a duplicated finding.
func recordConflict(
	ctx context.Context, tr tracker.Tracker, vc vcs.VersionControl, pr plan.Runner,
	req Request, target, dispatchBranch string, rebaseInv []string, ref *vcs.Refusal,
) ([]Fact, []plan.Record, error) {
	var records []plan.Record

	unmergedSteps := vc.Commits("conflict-unmerged", dispatchBranch, 1)
	targetSteps := vc.Commits("conflict-target", target, 1)
	steps := append(append(plan.Steps{}, unmergedSteps...), targetSteps...)
	report, err := pr.Run(ctx, plan.Plan{
		Name:  "the beads " + dispatchBranch + " and " + target + " each last closed",
		Steps: steps,
	}, req.Mode)
	records = append(records, report.Records...)
	if err != nil {
		return nil, records, RefuseUnrepairable("the conflicting beads could not be read: %s", err)
	}
	unmerged := beadAtHead(report, unmergedSteps)
	landed := beadAtHead(report, targetSteps)

	sort.Strings(ref.Conflicts)
	facts := []Fact{
		List("conflicts", ref.Conflicts, rebaseInv),
		Count("conflict_count", len(ref.Conflicts), rebaseInv),
		Text("unmerged_bead", unmerged, argvOf(mustRecord(report, unmergedSteps.Answer()))),
		Text("landed_bead", landed, argvOf(mustRecord(report, targetSteps.Answer()))),
	}

	if unmerged == "" {
		// Nothing to block and nowhere to record: the branch's head is
		// not a close commit, so no bead is attributable to it. Reported
		// rather than guessed at — a finding written on the wrong bead is
		// worse than one a human has to place.
		facts = append(facts, Unestablished("finding_recorded",
			"the head of "+dispatchBranch+" is not shaped like a bead's close commit, so no bead could be named",
			argvOf(mustRecord(report, unmergedSteps.Answer()))))
		return facts, records, nil
	}

	blockStep := tr.SetState("block-unmerged", unmerged, tracker.StateBlocked)
	noteStep := tr.Note("sequencing-finding", unmerged, sequencingFinding(dispatchBranch, target, unmerged, landed, ref))
	writeReport, err := pr.Run(ctx, plan.Plan{
		Name:  "block " + unmerged + " and record the sequencing finding on it",
		Steps: []plan.Step{blockStep, noteStep},
	}, req.Mode)
	records = append(records, writeReport.Records...)
	if err != nil {
		return facts, records, RefuseUnrepairable("%s", err)
	}
	blockRec, _ := writeReport.Record(blockStep.Name)
	noteRec, _ := writeReport.Record(noteStep.Name)
	if blockRec.Failed() {
		return facts, records, RefuseUnrepairable("marking %s blocked failed: %s", unmerged, describeFailure(blockRec))
	}
	if noteRec.Failed() {
		return facts, records, RefuseUnrepairable("recording the sequencing finding on %s failed: %s", unmerged, describeFailure(noteRec))
	}

	facts = append(facts,
		Flag("blocked", blockRec.Status == plan.StepRan, invocationOf(blockStep, blockRec)),
		Flag("finding_recorded", noteRec.Status == plan.StepRan, invocationOf(noteStep, noteRec)))
	return facts, records, nil
}

// sequencingFinding is the finding's text, in the run's existing finding
// convention — a `retro:` line naming the kind, then the evidence.
//
// It goes through [tracker.Tracker.Note] because that is the channel a
// run's findings already travel on; no finding command exists and this
// feature waits on none, so adopting a dedicated one later is a change of
// writer rather than of content.
func sequencingFinding(dispatchBranch, target, unmerged, landed string, ref *vcs.Refusal) string {
	other := landed
	if other == "" {
		other = "the commit at the head of " + target
	}
	return fmt.Sprintf(
		"retro:sequencing %s could not be rebased onto %s: it conflicts with %s on %d path(s)\n"+
			"evidence: %s\nconflicting paths: %s\n"+
			"The rebase was left where it stopped and the worktree left standing; nothing was resolved here.",
		unmerged, target, other, len(ref.Conflicts),
		ref.Error(), strings.Join(ref.Conflicts, ", "))
}

// beadAtHead reads the bead a branch's head commit closed, empty when
// the fragment did not answer or the subject is not a close commit's.
func beadAtHead(report plan.Report, steps vcs.Steps) string {
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return ""
	}
	commits, ok := rec.Value.([]vcs.Commit)
	if !ok || len(commits) == 0 {
		return ""
	}
	return status.BeadIDFromCloseSubject(commits[0].Subject)
}

// asVCSRefusal reads a step's error as version control declining to fold
// one line of work into another, nil for anything else.
func asVCSRefusal(err error) *vcs.Refusal {
	var ref *vcs.Refusal
	if errors.As(err, &ref) {
		return ref
	}
	return nil
}
