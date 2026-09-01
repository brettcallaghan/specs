// The `swarm dispatch open` command: establishes the sprint branch and,
// when a dispatch id is given, the dispatch's own worktree cut from it.
// Per the "swarm dispatch open" delta of [[container-swarm-cli]] (added —
// the sprint branch; added — the worktree) and FEAT-SWARM-DISP-003 of
// specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// The sprint branch is established first, and nowhere else — "the branch
// SHALL be created and recorded before the worktree is created" is a fact
// about ORDER, and this file is what makes that order structural: the
// worktree half runs only after the sprint-branch half has returned a
// branch to cut from.
//
// Hand-declared rather than built through [Spec] and [newSubcommand], for
// the same reason [phaseCloseCmd] is: whether the sprint branch needs
// creating at all is not known until the manifest has been read, so
// there is no single fixed plan to hand [newSubcommand] ahead of running
// it.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// dispatchCmd builds the `dispatch` noun.
func dispatchCmd(
	streams *ioStreams, store manifest.Store, tr tracker.Tracker,
	vc vcs.VersionControl, runner swarmexec.Runner,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dispatch",
		Short:        "Open and accept a run's dispatches",
		SilenceUsage: true,
	}
	cmd.AddCommand(dispatchOpenCmd(streams, store, vc, runner))
	cmd.AddCommand(dispatchAcceptCmd(streams, store, tr, vc, runner))
	return cmd
}

// dispatchOpenCmd builds `swarm dispatch open`.
func dispatchOpenCmd(streams *ioStreams, store manifest.Store, vc vcs.VersionControl, runner swarmexec.Runner) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
		dispatchID string
		base       string
		branchName string
	)

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Establish the sprint branch and, with --dispatch, its worktree",
		Long: "Creates the sprint branch and records it — together with the commit it was cut " +
			"from — in one manifest write, the first time any dispatch of this run is opened. " +
			"Every dispatch after the first reuses the recorded branch and creates none, and " +
			"--branch is ignored once one is recorded. A branch already carrying the name this " +
			"run would use, with the manifest recording none, stops the run rather than being " +
			"claimed for it. With --dispatch, stands a worktree up at " +
			"<parent>/.bb-worktrees/<dispatch-id> on a new branch bb/<dispatch-id>, cut from " +
			"--base or, absent that, the sprint branch's head at that moment; a dispatch whose " +
			"worktree already stands is reported and left untouched.",
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
				return finish(streams, req, explainDispatchOpen(req, dispatchID))
			}

			result, err := openDispatch(context.Background(), store, vc, runner, req, branchName, base, dispatchID)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe the invocations this command would run, and run none of them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"perform every read this command would perform, and create or record nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")
	cmd.Flags().StringVar(&dispatchID, "dispatch", "",
		"the dispatch this invocation is opening; absent, only the sprint branch is established and no worktree is created")
	cmd.Flags().StringVar(&base, "base", "",
		"the ref the sprint branch is cut from, when it does not exist yet (HEAD when absent), and the ref the dispatch's worktree is cut from (the sprint branch's own head when absent)")
	cmd.Flags().StringVar(&branchName, "branch", "",
		"the sprint branch's name, when it does not exist yet; sprint/<run-epic-id> when absent; ignored once the manifest records one")

	return cmd
}

// explainDispatchOpen describes the measurements and writes this command
// would take without taking any — whether the sprint branch needs
// creating at all is not known until the manifest has been read, so
// there is no fixed plan to render truthfully, the same reasoning
// [explainPhaseClose] gives.
func explainDispatchOpen(req Request, dispatchID string) Result {
	lines := []string{
		"swarm dispatch open derives whether the sprint branch already exists before it can " +
			"name a fixed sequence, so its steps are described here rather than printed as a plan:",
		"  bd show " + req.Run + " --json                              (the run manifest, for a recorded branch)",
		"  # if the manifest already records a branch: reused as-is, nothing else runs",
		"  # otherwise:",
		"  git for-each-ref refs/heads/                       (does a branch already carry the name this run would use)",
		"  git rev-parse --verify <base>^{commit}              (<base> is --base, or HEAD when absent)",
		"  git branch <name> <sha>                             (<name> is --branch, or sprint/" + req.Run + " when absent)",
		"  bd show " + req.Run + " --json                              (guard: the record still permits the write)",
		"  bd update " + req.Run + " --metadata '{...}' --json          (branch and base_sha, merged in one write)",
	}
	if strings.TrimSpace(dispatchID) != "" {
		lines = append(lines,
			"  # --dispatch is given, so the worktree half runs too:",
			"  git rev-parse --show-toplevel                       (the repository root, for the worktree's sibling path)",
			"  git worktree prune                                  (clear stale registrations before the list is read)",
			"  git worktree list --porcelain                       (does a worktree for "+dispatchID+" already stand)",
			"  # if one does: reported and left untouched, nothing else runs",
			"  # otherwise:",
			"  git rev-parse --verify <base>^{commit}              (<base> is --base, or the sprint branch's head when absent)",
			"  git worktree add -b bb/"+dispatchID+" <path> <sha>   (<path> is <parent>/.bb-worktrees/"+dispatchID+")",
		)
	}
	return Result{
		Command: req.Command, Run: req.Run, Mode: req.Mode, Outcome: OutcomeClear,
		Summary: "explained; nothing was measured or written", Transcript: lines,
	}
}

// defaultBranchName is the form the skills already create and resume
// onto, used whenever --branch is absent and no branch is recorded.
func defaultBranchName(run string) string { return "sprint/" + run }

// openSprintBranch is the impure half: it reads the manifest, and either
// reports the branch it already records or creates one and records it —
// every write going through pr.Run with req.Mode, so a dry run performs
// every read here and creates and records nothing, [plan.Runner]'s own
// guarantee.
func openSprintBranch(
	ctx context.Context, store manifest.Store, vc vcs.VersionControl, runner swarmexec.Runner,
	req Request, branchFlag, baseFlag string,
) (Result, error) {
	pr := plan.Runner{Exec: runner}
	var records []plan.Record

	readSteps := store.Read("manifest", req.Run)
	readReport, err := pr.Run(ctx, plan.Plan{Name: "read " + req.Run + "'s manifest", Steps: readSteps}, req.Mode)
	records = append(records, readReport.Records...)
	if err != nil {
		return Result{}, RefuseUnrepairable("the manifest could not be read: %s", err)
	}
	doc, from, err := document(readSteps, readReport)
	if err != nil {
		return Result{}, err
	}

	// AC-SWARM-DISP-003.2: a branch the manifest already records is this
	// run's, reused without a check — a name colliding with something
	// else is exactly what recording it in the manifest rules out.
	var recordedBranch string
	if present, decErr := doc.Decode("branch", &recordedBranch); decErr != nil {
		return Result{}, RefuseUnrepairable("the recorded branch does not decode: %s", decErr)
	} else if present {
		result := Result{Outcome: OutcomeClear, Records: records}
		result.Summary = fmt.Sprintf("reusing the sprint branch already recorded for %s: %s", req.Run, recordedBranch)
		result.Facts = append(result.Facts, Text("branch", recordedBranch, from))
		var recordedBase string
		if present, decErr := doc.Decode("base_sha", &recordedBase); decErr == nil && present {
			result.Facts = append(result.Facts, Text("base_sha", recordedBase, from))
		} else {
			result.Facts = append(result.Facts, Unestablished("base_sha",
				"the manifest records a branch but no base_sha for it", from))
		}
		return result, nil
	}

	branchName := branchFlag
	if branchName == "" {
		branchName = defaultBranchName(req.Run)
	}

	// AC-SWARM-DISP-003.3: a branch by this name that the manifest does
	// not record was put there by something other than this run, and
	// nothing here distinguishes which — so it stops the run rather than
	// being claimed for it.
	branchesStep := vc.Branches("branches")
	baseRef := baseFlag
	if baseRef == "" {
		baseRef = "HEAD"
	}
	resolveStep := vc.Resolve("resolve-base", baseRef)
	checkReport, err := pr.Run(ctx, plan.Plan{
		Name:  "whether " + branchName + " already exists, and what " + baseRef + " resolves to",
		Steps: append(append(plan.Steps{}, branchesStep...), resolveStep...),
	}, req.Mode)
	records = append(records, checkReport.Records...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	branchesRec, ok := checkReport.Record(branchesStep.Answer())
	if !ok || branchesRec.Failed() {
		return Result{}, RefuseUnrepairable("the local branches could not be read: %s", describeFailure(branchesRec))
	}
	branches, ok := branchesRec.Value.([]vcs.Branch)
	if !ok {
		return Result{}, RefuseUnrepairable("the branch list produced %T, not a branch set", branchesRec.Value)
	}
	for _, b := range branches {
		if b.Name == branchName {
			return Result{}, RefuseRepairable(
				fmt.Sprintf("pass an explicit branch name: %s --branch <name>", req.Command),
				"a branch named %q already exists and the manifest records no sprint branch for %s",
				branchName, req.Run)
		}
	}

	resolveRec, ok := checkReport.Record(resolveStep.Answer())
	if !ok || resolveRec.Failed() {
		return Result{}, RefuseUnrepairable("%q could not be resolved: %s", baseRef, describeFailure(resolveRec))
	}
	sha, ok := resolveRec.Value.(string)
	if !ok || sha == "" {
		return Result{}, RefuseUnrepairable("resolving %q produced %T, not a commit id", baseRef, resolveRec.Value)
	}

	// AC-SWARM-DISP-003.5, .6: created and recorded — the commit it was
	// cut from goes into the SAME merge as the branch name, since a
	// branch recorded without its base sha is a run whose worktrees have
	// nothing to compare their own base against.
	branchJSON, err := json.Marshal(branchName)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s does not encode as JSON: %s", branchName, err)
	}
	shaJSON, err := json.Marshal(sha)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s does not encode as JSON: %s", sha, err)
	}
	changes := map[string]json.RawMessage{
		"branch":   json.RawMessage(branchJSON),
		"base_sha": json.RawMessage(shaJSON),
	}
	createStep := vc.CreateBranch("create-branch", branchName, sha)
	writeSteps := store.Write("manifest-write", req.Run, changes, manifest.Rules(changes))

	writeReport, err := pr.Run(ctx, plan.Plan{
		Name:  "create " + branchName + " at " + sha + " and record it",
		Steps: append(append(plan.Steps{}, createStep...), writeSteps...),
	}, req.Mode)
	records = append(records, writeReport.Records...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}
	createRec, ok := writeReport.Record(createStep.Answer())
	if !ok {
		return Result{}, RefuseUnrepairable("the branch-creation step produced no record")
	}
	if createRec.Status == plan.StepRan && createRec.Failed() {
		return Result{}, RefuseUnrepairable("creating %s failed: %s", branchName, describeFailure(createRec))
	}
	if _, _, err := document(writeSteps, writeReport); err != nil {
		return Result{}, err
	}

	result := Result{Outcome: OutcomeClear, Records: records}
	result.Summary = fmt.Sprintf("created the sprint branch %s at %s", branchName, sha)
	writeInv := argvOf(mustRecord(writeReport, writeSteps.Answer()))
	result.Facts = append(result.Facts,
		Text("branch", branchName, writeInv),
		Text("base_sha", sha, writeInv))
	return result, nil
}

// mustRecord reads a named record out of a report the caller has already
// established carries it — used only after [document] has already
// succeeded against the same report, so the record is known to exist.
func mustRecord(report plan.Report, name string) plan.Record {
	rec, _ := report.Record(name)
	return rec
}

// openDispatch runs the sprint-branch half and, when a dispatch id is
// given, the worktree half — in that order, so the worktree is never cut
// before the branch it comes from has been established. Absent a
// dispatch id, only the sprint branch is addressed, unchanged from
// before this half existed.
func openDispatch(
	ctx context.Context, store manifest.Store, vc vcs.VersionControl, runner swarmexec.Runner,
	req Request, branchFlag, baseFlag, dispatchID string,
) (Result, error) {
	result, err := openSprintBranch(ctx, store, vc, runner, req, branchFlag, baseFlag)
	if err != nil {
		return Result{}, err
	}

	dispatchID = strings.TrimSpace(dispatchID)
	if dispatchID == "" {
		return result, nil
	}

	var branchName string
	for _, f := range result.Facts {
		if f.Name == "branch" && f.Value != nil && f.Value.Text != nil {
			branchName = *f.Value.Text
			break
		}
	}
	if branchName == "" {
		return Result{}, RefuseUnrepairable("the sprint branch was not established, so %s's worktree has nothing to cut from", dispatchID)
	}

	wtFacts, wtRecords, wtSummary, err := openDispatchWorktree(ctx, vc, runner, req, branchName, baseFlag, dispatchID)
	if err != nil {
		return Result{}, err
	}

	result.Records = append(result.Records, wtRecords...)
	result.Facts = append(result.Facts, wtFacts...)
	result.Summary = result.Summary + "; " + wtSummary
	return result, nil
}

// dispatchBranchName is the form the skills already create and clean
// up, so a worktree the coordinator stands is attributable by branch
// name alone. Per AC-SWARM-DISP-003.8: derived from the dispatch id
// alone, so two sessions holding the same id compute the same value.
func dispatchBranchName(dispatchID string) string { return "bb/" + dispatchID }

// dispatchWorktreePath is the pinned form: a sibling of the repository
// root, never inside it (AC-SWARM-DISP-003.7), derived from the
// repository root and the dispatch id alone (AC-SWARM-DISP-003.8).
func dispatchWorktreePath(repoRoot, dispatchID string) string {
	return filepath.Join(filepath.Dir(repoRoot), ".bb-worktrees", dispatchID)
}

// openDispatchWorktree stands the dispatch's worktree up, cut from
// branchName's current head unless baseFlag names another ref, and
// reports an existing one untouched (AC-SWARM-DISP-003.11) rather than
// recreating it.
func openDispatchWorktree(
	ctx context.Context, vc vcs.VersionControl, runner swarmexec.Runner,
	req Request, branchName, baseFlag, dispatchID string,
) (facts []Fact, records []plan.Record, summary string, err error) {
	pr := plan.Runner{Exec: runner}
	dispatchBranch := dispatchBranchName(dispatchID)

	// AC-SWARM-DISP-003.9: the prune runs before the list is read —
	// [vcs.VersionControl.Worktrees] bundles the two as one fragment, so
	// no caller here can read the list without clearing stale
	// registrations first.
	wtSteps := vc.Worktrees("dispatch-wt")
	wtReport, err := pr.Run(ctx, plan.Plan{Name: "the standing worktrees", Steps: wtSteps}, req.Mode)
	records = append(records, wtReport.Records...)
	if err != nil {
		return nil, records, "", RefuseUnrepairable("%s", err)
	}
	wtRec, ok := wtReport.Record(wtSteps.Answer())
	if !ok || wtRec.Failed() {
		return nil, records, "", RefuseUnrepairable("the standing worktrees could not be read: %s", describeFailure(wtRec))
	}
	worktrees, ok := wtRec.Value.([]vcs.Worktree)
	if !ok {
		return nil, records, "", RefuseUnrepairable("the worktree list produced %T, not a worktree set", wtRec.Value)
	}
	wtInv := invocationOf(wtSteps[len(wtSteps)-1], wtRec)

	// AC-SWARM-DISP-003.11: a worktree already standing for this
	// dispatch is reported and left untouched rather than recreated.
	for _, w := range worktrees {
		if w.Branch == dispatchBranch {
			facts = append(facts,
				Text("worktree_path", w.Path, wtInv),
				Text("dispatch_branch", dispatchBranch, wtInv))
			summary = fmt.Sprintf("reusing the worktree already standing for %s: %s on %s",
				dispatchID, w.Path, dispatchBranch)
			return facts, records, summary, nil
		}
	}

	// AC-SWARM-DISP-003.7, .8: the path is a sibling of the repository
	// root, derived from the root and the dispatch id alone — read only
	// now, since the fast path above never needed it.
	rootStep := vc.Root("dispatch-root")
	rootReport, err := pr.Run(ctx, plan.Plan{Name: "the repository's working-tree root", Steps: rootStep}, req.Mode)
	records = append(records, rootReport.Records...)
	if err != nil {
		return nil, records, "", RefuseUnrepairable("%s", err)
	}
	rootRec, ok := rootReport.Record(rootStep.Answer())
	if !ok || rootRec.Failed() {
		return nil, records, "", RefuseUnrepairable("the repository root could not be read: %s", describeFailure(rootRec))
	}
	root, ok := rootRec.Value.(string)
	if !ok || root == "" {
		return nil, records, "", RefuseUnrepairable("resolving the repository root produced %T, not a path", rootRec.Value)
	}
	path := dispatchWorktreePath(root, dispatchID)

	// AC-SWARM-DISP-003.1, .10, .12: the base is the caller's --base, or
	// the sprint branch's head read now — never a global setting.
	baseRef := baseFlag
	if baseRef == "" {
		baseRef = branchName
	}
	resolveStep := vc.Resolve("dispatch-base", baseRef)
	resolveReport, err := pr.Run(ctx, plan.Plan{Name: "resolve " + baseRef, Steps: resolveStep}, req.Mode)
	records = append(records, resolveReport.Records...)
	if err != nil {
		return nil, records, "", RefuseUnrepairable("%s", err)
	}
	resolveRec, ok := resolveReport.Record(resolveStep.Answer())
	if !ok || resolveRec.Failed() {
		return nil, records, "", RefuseUnrepairable("%q could not be resolved: %s", baseRef, describeFailure(resolveRec))
	}
	baseSHA, ok := resolveRec.Value.(string)
	if !ok || baseSHA == "" {
		return nil, records, "", RefuseUnrepairable("resolving %q produced %T, not a commit id", baseRef, resolveRec.Value)
	}

	// AC-SWARM-DISP-003.4: exactly one worktree and one branch, named
	// for the dispatch — [vcs.VersionControl.CreateWorktree] cuts the
	// branch as part of standing the worktree up, so there is no second
	// call that could create a second branch.
	createStep := vc.CreateWorktree("dispatch-wt-create", path, dispatchBranch, baseSHA)
	createReport, err := pr.Run(ctx, plan.Plan{Name: "stand a worktree up at " + path, Steps: createStep}, req.Mode)
	records = append(records, createReport.Records...)
	if err != nil {
		return nil, records, "", RefuseUnrepairable("%s", err)
	}
	createRec, ok := createReport.Record(createStep.Answer())
	if !ok {
		return nil, records, "", RefuseUnrepairable("the worktree-creation step produced no record")
	}
	if createRec.Status == plan.StepRan && createRec.Failed() {
		return nil, records, "", RefuseUnrepairable("standing a worktree up at %s failed: %s", path, describeFailure(createRec))
	}

	createInv := invocationOf(createStep[len(createStep)-1], createRec)
	facts = append(facts,
		Text("worktree_path", path, createInv),
		Text("dispatch_branch", dispatchBranch, createInv),
		Text("worktree_base_sha", baseSHA, createInv))
	summary = fmt.Sprintf("stood a worktree up for %s at %s on %s, cut from %s",
		dispatchID, path, dispatchBranch, baseSHA)
	return facts, records, summary, nil
}

// invocationOf reports the argv a step ran with, or — for a write a dry
// run refused — the argv it would have run with. A refused step's own
// record carries no invocation ([argvOf] answers nil for it), and every
// argument this half's steps build is a literal, so the step's own
// declaration already names it.
func invocationOf(step plan.Step, rec plan.Record) []string {
	if argv := argvOf(rec); argv != nil {
		return argv
	}
	argv := []string{step.Executable}
	for _, a := range step.Args {
		if a.IsRef() {
			return nil
		}
		argv = append(argv, a.Literal())
	}
	return argv
}
