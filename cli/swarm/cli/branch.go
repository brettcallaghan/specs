// The `swarm branch verify-landed` and `swarm branch push` commands: one
// fetch and one comparison, and — for `push` — the push ordered in front
// of them. Per the "swarm branch verify-landed and swarm branch push"
// delta of [[container-swarm-cli]] and FEAT-SWARM-DISP-005 of
// specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// A rejected push leaves every other signal reading finished — the tree
// is clean, no worktrees stand, the beads are closed — so the comparison
// against the remote is the only thing separating work that landed from
// work stranded on a container about to be reclaimed.
//
// **`push` is `verify-landed` with the push in front of it, and that is
// structural here rather than a resemblance**: both verbs are built by
// [branchLandingCmd] and both run [verifyLanded], which takes the push as
// a flag on the same plan. There is no second implementation for the two
// to drift apart in.
//
// **Neither command refuses an absent run scope**, which is the one
// exception to the rule every other command in this surface follows.
// Both are the recovery path: a session whose first act is the landing
// check does not yet know which run it is in, and requiring a scope would
// defeat the purpose. `--run` is still declared, so the flag set every
// command carries is uniform, and it is carried into the result when
// given.
package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// defaultRemote is the remote used WHERE --remote is absent. It is a
// constant here and read from nowhere else: the delta requires both
// defaults to come from the repository or from this declaration, never
// from configuration outside the repository.
const defaultRemote = "origin"

// detachedHead is what `git rev-parse --abbrev-ref HEAD` answers when the
// repository is on no branch. It is a literal to be read as "no branch",
// never as a branch by that name.
const detachedHead = "HEAD"

// branchCmd builds the `branch` noun.
func branchCmd(streams *ioStreams, vc vcs.VersionControl, runner swarmexec.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "branch",
		Short:        "Prove a branch landed on its remote",
		SilenceUsage: true,
	}
	cmd.AddCommand(branchLandingCmd(streams, vc, runner, false))
	cmd.AddCommand(branchLandingCmd(streams, vc, runner, true))
	return cmd
}

// branchLandingCmd builds `verify-landed`, and — with pushFirst — `push`.
//
// One builder for both verbs, because the delta requires one
// implementation rather than two commands that resemble each other. The
// flag is the whole difference: it puts the push at the front of the same
// plan.
func branchLandingCmd(streams *ioStreams, vc vcs.VersionControl, runner swarmexec.Runner, pushFirst bool) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
		branchName string
		remoteName string
	)

	use, short := "verify-landed", "Report whether a branch has landed on its remote"
	long := "Fetches the remote and compares the branch against its counterpart there, so a " +
		"stale local record of where the remote points cannot report a branch as landed. A " +
		"branch ahead of its remote is reported with the number of unpushed commits, and one " +
		"with no counterpart on the remote is reported as unlanded rather than as clean — " +
		"both take the attention status. WHERE --branch is absent the branch is the one " +
		"currently checked out, and WHERE --remote is absent the remote is origin; both are " +
		"read from the repository. It needs no run scope, so a recovering session can run it " +
		"before it knows which run it is in."
	if pushFirst {
		use, short = "push", "Push a branch and prove it landed"
		long = "Pushes the branch to the remote and then performs the landing check, in that " +
			"order — the same check `verify-landed` performs, with the push in front of it. A " +
			"rejected push is reported with the remote's own cause and every ref it refused, " +
			"and no commit in that push is reported as landed. " + long
	}

	cmd := &cobra.Command{
		Use:          use,
		Short:        short,
		Long:         long,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req := Request{
				Command:    cmd.CommandPath(),
				Run:        strings.TrimSpace(runID),
				JSON:       asJSON,
				FullResult: fullResult || explain || dryRun,
			}

			// AC-SWARM-DISP-005.6: no run-scope refusal here. The scope
			// is carried when given and asked for never.

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

			remote := strings.TrimSpace(remoteName)
			if remote == "" {
				remote = defaultRemote
			}
			branch := strings.TrimSpace(branchName)

			if req.Mode == plan.ModeExplain {
				return finish(streams, req, explainBranchLanding(req, branch, remote, pushFirst))
			}

			result, err := verifyLanded(context.Background(), vc, runner, req, branch, remote, pushFirst)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe the invocations this command would run, and run none of them")
	dryRunHelp := "perform every read this command would perform, and fetch nothing"
	if pushFirst {
		dryRunHelp = "perform every read this command would perform, and push and fetch nothing"
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, dryRunHelp)
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this invocation belongs to; optional, since the landing check is the recovery path and needs no scope")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")
	cmd.Flags().StringVar(&branchName, "branch", "",
		"the branch to check; the one currently checked out when absent")
	cmd.Flags().StringVar(&remoteName, "remote", "",
		"the remote to check it against; "+defaultRemote+" when absent")

	return cmd
}

// explainBranchLanding describes the invocations this command would run
// without running any.
//
// Not a [plan.Plan]: the branch is not known until the repository has
// been asked which one it is on, so the first invocation's ANSWER is an
// argument to the rest — which is exactly what explain mode exists not to
// run.
func explainBranchLanding(req Request, branch, remote string, pushFirst bool) Result {
	named := branch
	if named == "" {
		named = "<branch>"
	}
	lines := []string{
		req.Command + " reads the branch it is checking before it can name the rest of its " +
			"invocations, so its steps are described here rather than printed as a plan:",
	}
	if branch == "" {
		lines = append(lines,
			"  git rev-parse --abbrev-ref HEAD                     (<branch>, since --branch is absent; `HEAD` here means the repository is on none)")
	}
	if pushFirst {
		lines = append(lines,
			"  git push "+remote+" "+named+"                          (ordered FIRST; a rejection stops the rest)")
	}
	lines = append(lines,
		"  git fetch "+remote+"                                 (before the comparison, so a stale record cannot report a branch as landed)",
		"  git rev-parse --verify "+remote+"/"+named+"^{commit}   (does the branch have a counterpart on "+remote+" at all)",
		"  git rev-list --count --left-right "+named+"..."+remote+"/"+named+"  (how far the branch stands from it)")
	return req.stamp(Result{
		Outcome: OutcomeClear,
		Summary: "explained; nothing was pushed, fetched or compared", Transcript: lines,
	})
}

// verifyLanded is the impure half, and the whole of both verbs.
//
// The ordering the delta fixes is the plan's: the push is the first step
// of the same plan the fetch and the comparison are in, so a rejected
// push stops the plan before anything reports a landing — [plan.Stop] is
// the step's own failure policy, not a conjunction this function could
// forget. Every write goes through pr.Run with req.Mode, so a dry run
// performs the reads and pushes and fetches nothing.
func verifyLanded(
	ctx context.Context, vc vcs.VersionControl, runner swarmexec.Runner,
	req Request, branch, remote string, pushFirst bool,
) (Result, error) {
	pr := plan.Runner{Exec: runner}
	var records []plan.Record

	branch, branchInv, recs, err := landingBranch(ctx, vc, pr, req, branch)
	records = append(records, recs...)
	if err != nil {
		return Result{}, err
	}

	steps := plan.Steps{}
	var pushSteps vcs.Steps
	if pushFirst {
		pushSteps = vc.Push("push", remote, branch)
		steps = append(steps, pushSteps...)
	}
	// AC-SWARM-DISP-005.5: the fetch is a step of this plan, ordered
	// before the two reads that compare — so the comparison cannot be
	// made against a stale record of where the remote points, and no
	// caller can reorder it.
	fetchSteps := vc.Fetch("fetch", remote)
	counterpartSteps := vc.Resolve("counterpart", remote+"/"+branch)
	divergeSteps := vc.Divergence("divergence", branch, remote)
	steps = append(steps, fetchSteps...)
	steps = append(steps, counterpartSteps...)
	steps = append(steps, divergeSteps...)

	name := "whether " + branch + " has landed on " + remote
	if pushFirst {
		name = "push " + branch + " to " + remote + ", then " + name
	}
	report, err := pr.Run(ctx, plan.Plan{Name: name, Steps: steps}, req.Mode)
	records = append(records, report.Records...)
	if err != nil {
		return Result{}, RefuseUnrepairable("%s", err)
	}

	// The counterpart probe's argv names the branch AND the remote, so it
	// is what establishes both — beside the read that supplied the branch
	// when the caller named none.
	counterpartInv := invocationOf(counterpartSteps[len(counterpartSteps)-1],
		mustRecord(report, counterpartSteps.Answer()))
	inv := [][]string{counterpartInv}
	if branchInv != nil {
		inv = append(inv, branchInv)
	}

	result := req.stamp(Result{Outcome: OutcomeClear, Records: records})
	result.Facts = append(result.Facts,
		Text("branch", branch, inv...),
		Text("remote", remote, inv...))

	if pushFirst {
		if rec, ok := report.Record(pushSteps.Answer()); ok && rec.Failed() {
			// AC-SWARM-DISP-005.3, .7: the cause reaches the caller, the
			// refused refs are named, and nothing in the push is reported
			// as landed.
			return rejected(result, remote, rec), nil
		}
	}

	landed, err := landingVerdict(report, fetchSteps, counterpartSteps, divergeSteps, branch, remote, pushFirst)
	if err != nil {
		return Result{}, err
	}
	result.Facts = append(result.Facts, landed.facts...)
	result.Outcome = landed.outcome
	result.Summary = landed.summary
	result.Resolution = landed.resolution
	return result, nil
}

// landingBranch answers which branch is being checked: the one the caller
// named, or the one the repository is on.
//
// AC-SWARM-DISP-005: the default is read from the repository. A detached
// HEAD is refused rather than checked, because `HEAD` is what git answers
// there and treating it as a branch name would compare the repository
// against a remote ref that does not exist and report "no counterpart" —
// a true statement about the wrong question.
func landingBranch(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request, named string,
) (string, []string, []plan.Record, error) {
	if named != "" {
		// Named by the caller, so nothing was read to establish it and
		// there is no invocation of this command's own to cite for it.
		return named, nil, nil, nil
	}

	steps := vc.CurrentBranch("current-branch")
	report, err := pr.Run(ctx, plan.Plan{Name: "the branch the repository is on", Steps: steps}, req.Mode)
	if err != nil {
		return "", nil, report.Records, RefuseUnrepairable("%s", err)
	}
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return "", nil, report.Records, RefuseUnrepairable(
			"the branch the repository is on could not be read: %s", describeFailure(rec))
	}
	branch, ok := rec.Value.(string)
	if !ok || branch == "" {
		return "", nil, report.Records, RefuseUnrepairable(
			"reading the current branch produced %T, not a branch name", rec.Value)
	}
	if branch == detachedHead {
		return "", nil, report.Records, RefuseRepairable(
			"name the branch: "+req.Command+" --branch <name>",
			"the repository is on no branch, so there is none to check against a remote")
	}
	return branch, argvOf(rec), report.Records, nil
}

// verdict is what the comparison concluded, and how it is reported.
type verdict struct {
	facts      []Fact
	outcome    Outcome
	summary    string
	resolution string
}

// landingVerdict reads the counterpart probe and the divergence count
// into the command's answer.
//
// pushFirst is which of the two verbs the caller invoked. `push` and
// `verify-landed` are one implementation rather than two commands that
// resemble each other, so a resolution telling the caller to run the
// command again has to name the one they ran: sending a caller of `push`
// to `verify-landed` has them rehearse a push and then perform only the
// check.
func landingVerdict(
	report plan.Report, fetchSteps, counterpartSteps, divergeSteps vcs.Steps, branch, remote string,
	pushFirst bool,
) (verdict, error) {
	tracking := remote + "/" + branch

	verb := "verify-landed"
	if pushFirst {
		verb = "push"
	}

	// The fetch is what makes the two reads below mean anything, and it
	// is the FIRST thing tested for that reason. It is a write, so a dry
	// run refuses it while the reads — both plain reads — still run,
	// against whatever the local tracking ref happens to hold. A stale
	// ref is indistinguishable from a fresh one from in here: it resolves,
	// it compares, and it reports a branch as landed. So the verdict rests
	// on whether the remote was actually consulted, never on what the
	// reads answered when it was not.
	fetchRec, ok := report.Record(fetchSteps.Answer())
	if !ok || fetchRec.Status != plan.StepRan || fetchRec.Failed() {
		return verdict{
			facts: []Fact{
				Unestablished("landed",
					remote+" was not fetched, so the comparison against "+tracking+
						" would rest on a possibly stale local record of where "+remote+" points",
					invocationOf(fetchSteps[len(fetchSteps)-1], fetchRec)),
			},
			outcome: OutcomeAttention,
			summary: fmt.Sprintf("whether %s has landed on %s was not established: %s was not fetched",
				branch, remote, remote),
			resolution: fmt.Sprintf(
				"run it again without --dry-run, or fetch first: swarm branch %s --branch %s --remote %s",
				verb, branch, remote),
		}, nil
	}

	counterpartRec, ok := report.Record(counterpartSteps.Answer())
	if !ok {
		return verdict{}, RefuseUnrepairable("the counterpart probe produced no record")
	}

	if counterpartRec.Failed() {
		// AC-SWARM-DISP-005.2: no counterpart is not "landed".
		return verdict{
			facts: []Fact{
				Flag("landed", false, invocationOf(counterpartSteps[len(counterpartSteps)-1], counterpartRec)),
			},
			outcome: OutcomeAttention,
			summary: fmt.Sprintf("%s has no counterpart on %s, so nothing on it has landed there",
				branch, remote),
			resolution: fmt.Sprintf("push it: swarm branch push --branch %s --remote %s", branch, remote),
		}, nil
	}

	divergeRec, ok := report.Record(divergeSteps.Answer())
	if !ok || divergeRec.Status != plan.StepRan {
		// The fetch ran, so this is a plan that stopped before the
		// comparison — reported rather than guessed at.
		//
		// It gets its own resolution rather than the fetch guard's. A dry
		// run cannot be here: the fetch is a write, a dry run refuses it,
		// and the guard above returns on exactly that — so "run it again
		// without --dry-run" would describe a mode the caller was not in,
		// and for `push` it would propose a WRITE as the remedy for a
		// read that did not execute. What is missing is one comparison,
		// so the resolution names that comparison and how to take it, for
		// both verbs alike.
		return verdict{
			facts: []Fact{
				Unestablished("landed",
					"the comparison against "+tracking+" did not run",
					invocationOf(divergeSteps[len(divergeSteps)-1], divergeRec)),
			},
			outcome: OutcomeAttention,
			summary: fmt.Sprintf("whether %s has landed on %s was not established", branch, remote),
			resolution: fmt.Sprintf(
				"the divergence comparison against %s did not run, so how far %s stands from it "+
					"is unknown; take it directly: git rev-list --count --left-right %s...%s",
				tracking, branch, branch, tracking),
		}, nil
	}
	if divergeRec.Failed() {
		return verdict{}, RefuseUnrepairable("%s could not be compared against %s: %s",
			branch, tracking, describeFailure(divergeRec))
	}
	divergence, ok := divergeRec.Value.(vcs.Divergence)
	if !ok {
		return verdict{}, RefuseUnrepairable(
			"comparing %s against %s produced %T, not a divergence", branch, tracking, divergeRec.Value)
	}
	inv := argvOf(divergeRec)

	if divergence.Ahead > 0 {
		// AC-SWARM-DISP-005.1: the count is named, and the status is
		// attention.
		return verdict{
			facts: []Fact{
				Flag("landed", false, inv),
				Count("unpushed", divergence.Ahead, inv),
			},
			outcome: OutcomeAttention,
			summary: fmt.Sprintf("%s is %d commit(s) ahead of %s; that much has not landed",
				branch, divergence.Ahead, tracking),
			resolution: fmt.Sprintf("push it: swarm branch push --branch %s --remote %s", branch, remote),
		}, nil
	}

	// AC-SWARM-DISP-005.4: every local commit is on the remote.
	summary := fmt.Sprintf("%s has landed: every commit on it is on %s", branch, tracking)
	if divergence.Behind > 0 {
		summary += fmt.Sprintf(" (and %s carries %d commit(s) this branch does not)", tracking, divergence.Behind)
	}
	return verdict{
		facts: []Fact{
			Flag("landed", true, inv),
			Count("unpushed", 0, inv),
		},
		outcome: OutcomeClear,
		summary: summary,
	}, nil
}

// rejected turns a refused push into the command's answer:
// AC-SWARM-DISP-005.3 and .7.
//
// `landed` is false rather than unestablished, because a rejected push is
// not a measurement that failed — git reports the WHOLE push as failed
// when the remote refuses any ref in it, so no commit in it reached the
// remote. Saying so is the point of the command.
func rejected(result Result, remote string, rec plan.Record) Result {
	inv := argvOf(rec)
	result.Facts = append(result.Facts, Flag("landed", false, inv))

	var rejection *vcs.Rejection
	if errors.As(rec.Err, &rejection) {
		result.Facts = append(result.Facts,
			List("refused", rejection.Refused, inv),
			Text("rejection", rejection.Reason, inv))
		result.Outcome = OutcomeStop
		result.Summary = rejection.Error() + "; no commit in that push landed"
		return result
	}

	result.Outcome = OutcomeStop
	result.Summary = fmt.Sprintf("the push to %s failed, so nothing in it landed: %s",
		remote, describeFailure(rec))
	return result
}
