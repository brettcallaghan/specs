// The `swarm dispatch accept` command, second half: the ancestry guard,
// the removal, the post-merge acceptance re-check and the landing. Per
// the "swarm dispatch accept" delta of [[container-swarm-cli]] (added —
// ancestry, removal and landing) and FEAT-SWARM-DISP-004 of
// specs/features/swarm-dispatch/swarm-dispatch-frd.md.
//
// The whole point of this half is that the removal sits BEHIND the
// merge's result rather than beside it. A shell conjunction is not a
// guard: a merge that exited non-zero followed by a removal that
// succeeded leaves a removed worktree beside an unmerged branch, which is
// the one state the in-flight signal asserts cannot occur. Here the
// ordering is the plan's ([[adr-024]], [[adr-025]]) in two different
// ways, and both are needed:
//
//   - The ancestry read is a plan of its own, because its ANSWER decides
//     whether a removal plan is built at all. A guard whose subject is
//     "should the next plan exist" cannot be a step of the plan it
//     guards.
//   - The removal and the branch deletion are two steps of ONE plan, so
//     the deletion does not run when the removal failed —
//     [plan.Runner]'s ordering guard rather than a conjunction this file
//     could forget (AC-SWARM-DISP-004.6, satisfied via the relied-upon
//     Plan Runner and asserted, not reimplemented).
//
// [[adr-026]] orders the irreversible writes last, and every ordering
// here is an instance of it: the removal goes behind the ancestry read,
// the deletion behind the removal, and the push behind every post-merge
// signal.
//
// # Nothing here forces
//
// [vcs.VersionControl.RemoveWorktree] takes no force argument, so there
// is no forcing form for a retry to reach for; and where a removal is
// refused because the worktree is LOCKED, the resolution this file names
// is the unlock invocation. A worktree that refuses removal is how an
// orchestrator learns the dispatch is unfinished, and discarding that is
// the failure the absent parameter exists to prevent.
package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// acceptanceScan bounds the commit window the "committed" signal reads
// on the target.
//
// Bounded rather than the whole history, so one dispatch's acceptance
// costs one bounded read on a repository of any age. It is far larger
// than a dispatch's own commit count — a batch is a handful of beads —
// so a bead whose close commit is not in the window is an anomaly worth
// reporting rather than a limit worth raising.
const acceptanceScan = 500

// findingPrefix is the run's finding convention: a comment recording a
// finding begins with it. Reading the convention here rather than
// counting comments is what makes the signal "the work was recorded"
// instead of "somebody said something".
const findingPrefix = "retro:"

// settlement is what the second half established: the facts, what to say
// about them, and — when something needs attention — the outcome and the
// resolution that go with it.
type settlement struct {
	facts      []Fact
	summary    string
	outcome    Outcome
	resolution string
}

// settleDispatch performs everything after the fold: the ancestry guard,
// the removal and the deletion, the post-merge acceptance re-check, and
// the landing.
//
// merged says whether the fold actually RAN. It is false only in a dry
// run — a fold that failed returns before this half is reached at all —
// and in that case there is nothing for the guard to guard and nothing
// to remove, delete, re-check or push (AC-SWARM-DISP-004.13).
func settleDispatch(
	ctx context.Context, tr tracker.Tracker, vc vcs.VersionControl,
	pr plan.Runner, runner swarmexec.Runner, req Request,
	target, dispatchID string, merged, noPush bool,
) (settlement, []plan.Record, error) {
	dispatchBranch := dispatchBranchName(dispatchID)
	var records []plan.Record
	s := settlement{outcome: OutcomeClear}

	ancestor, ancestorInv, recs, err := readAncestry(ctx, vc, pr, req, target, dispatchBranch)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}
	s.facts = append(s.facts, Flag("ancestor", ancestor, ancestorInv))

	if !merged {
		// AC-SWARM-DISP-004.13: every read above ran and no write did.
		// The steps below are reported as not attempted rather than
		// reported as having found nothing, which are different facts.
		s.facts = append(s.facts,
			Unestablished("worktree_removed",
				"nothing was merged, so no removal was attempted", ancestorInv),
			Unestablished("branch_deleted",
				"nothing was merged, so no branch was deleted", ancestorInv),
			Unestablished("landed",
				"nothing was merged, so no push was attempted", ancestorInv))
		s.summary = "nothing was removed, deleted or pushed"
		return s, records, nil
	}

	// AC-SWARM-DISP-004.1: the removal is behind this, and a branch that
	// is not an ancestor of the target leaves the worktree standing, the
	// branch undeleted, and the acceptance stopped.
	if !ancestor {
		return s, records, &acceptStop{
			refusal: RefuseUnrepairable(
				"%s is not an ancestor of %s, so its worktree was left standing and its branch was not deleted",
				dispatchBranch, target),
			facts: s.facts,
		}
	}

	standing, wtInv, recs, err := standingWorktree(ctx, vc, pr, req, dispatchBranch)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}

	removal, recs, err := removeAndDelete(ctx, vc, pr, req, dispatchID, dispatchBranch, standing, wtInv)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}
	s.facts = append(s.facts, removal.facts...)
	if removal.outcome != OutcomeClear {
		// AC-SWARM-DISP-004.10: the removal was refused, nothing was
		// retried with force, and the branch was not deleted — so no
		// re-check and no push follow.
		s.outcome = removal.outcome
		s.resolution = removal.resolution
		s.summary = removal.summary
		return s, records, nil
	}

	// AC-SWARM-DISP-004.11: every bead in the dispatch, re-checked with
	// no signal deferred — the merge and the removal have both happened,
	// so the two signals a merge decides are now readable.
	check, recs, err := recheckDispatchBeads(
		ctx, tr, vc, pr, req, target, dispatchID, dispatchBranch, ancestor, ancestorInv)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}
	s.facts = append(s.facts, check.facts...)
	if check.outcome != OutcomeClear {
		// AC-SWARM-DISP-004.12: a bead that failed its post-merge check
		// stops the acceptance before the push, with the failing signal
		// named.
		s.outcome = check.outcome
		s.resolution = check.resolution
		s.summary = removal.summary + "; " + check.summary
		return s, records, nil
	}

	land, recs, err := landTarget(ctx, vc, runner, req, target, noPush)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}
	s.facts = append(s.facts, land.facts...)
	if land.outcome != OutcomeClear {
		s.outcome = land.outcome
		s.resolution = land.resolution
	}
	s.summary = removal.summary + "; " + check.summary + "; " + land.summary
	return s, records, nil
}

// readAncestry answers whether the dispatch branch is an ancestor of the
// target — the guard the removal sits behind.
//
// Its own plan, because the answer decides whether a removal plan is
// built at all: a guard whose subject is "should the next plan exist"
// cannot be a step of the plan it guards.
func readAncestry(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request,
	target, dispatchBranch string,
) (bool, []string, []plan.Record, error) {
	steps := vc.Descends("ancestry", dispatchBranch, target)
	report, err := pr.Run(ctx, plan.Plan{
		Name:  "is " + dispatchBranch + " an ancestor of " + target,
		Steps: steps,
	}, req.Mode)
	if err != nil {
		return false, nil, report.Records, RefuseUnrepairable("%s", err)
	}
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return false, nil, report.Records, RefuseUnrepairable(
			"whether %s is an ancestor of %s could not be read: %s",
			dispatchBranch, target, describeFailure(rec))
	}
	ancestor, ok := rec.Value.(bool)
	if !ok {
		return false, nil, report.Records, RefuseUnrepairable(
			"the ancestry comparison produced %T, not a verdict", rec.Value)
	}
	return ancestor, argvOf(rec), report.Records, nil
}

// standingWorktree finds the worktree this dispatch's branch is checked
// out in, nil when none stands.
//
// Read BEFORE the removal is attempted, and that ordering is what makes
// a refusal nameable: a step's record carries a DIGEST of stderr rather
// than stderr, so "what the tool refused over" has to come from something
// this command measured. [vcs.Worktree] carries the lock and its reason,
// which is the one cause the delta names a specific resolution for.
func standingWorktree(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request, dispatchBranch string,
) (*vcs.Worktree, []string, []plan.Record, error) {
	steps := vc.Worktrees("accept-wt")
	report, err := pr.Run(ctx, plan.Plan{Name: "the standing worktrees", Steps: steps}, req.Mode)
	if err != nil {
		return nil, nil, report.Records, RefuseUnrepairable("%s", err)
	}
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return nil, nil, report.Records, RefuseUnrepairable(
			"the standing worktrees could not be read: %s", describeFailure(rec))
	}
	worktrees, ok := rec.Value.([]vcs.Worktree)
	if !ok {
		return nil, nil, report.Records, RefuseUnrepairable(
			"the worktree list produced %T, not a worktree set", rec.Value)
	}
	inv := invocationOf(steps[len(steps)-1], rec)
	for i := range worktrees {
		if worktrees[i].Branch == dispatchBranch {
			return &worktrees[i], inv, report.Records, nil
		}
	}
	return nil, inv, report.Records, nil
}

// removeAndDelete removes the dispatch's worktree and then deletes its
// branch, as two steps of ONE plan.
//
// The ordering obligation is the plan's: [plan.Step]'s zero failure
// policy is Stop, so a refused removal stops the plan and the deletion
// records as skipped. Nothing here re-runs the removal, and there is no
// forcing form to re-run it with.
func removeAndDelete(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request,
	dispatchID, dispatchBranch string, standing *vcs.Worktree, wtInv []string,
) (settlement, []plan.Record, error) {
	s := settlement{outcome: OutcomeClear}

	steps := plan.Steps{}
	var removeSteps vcs.Steps
	if standing != nil {
		removeSteps = vc.RemoveWorktree("remove-worktree", standing.Path)
		steps = append(steps, removeSteps...)
	}
	deleteSteps := vc.DeleteBranch("delete-branch", dispatchBranch)
	steps = append(steps, deleteSteps...)

	name := "delete " + dispatchBranch
	if standing != nil {
		name = "remove " + standing.Path + ", then delete " + dispatchBranch
	}
	report, err := pr.Run(ctx, plan.Plan{Name: name, Steps: steps}, req.Mode)
	if err != nil {
		return s, report.Records, RefuseUnrepairable("%s", err)
	}

	if standing == nil {
		// Nothing stood for this dispatch, so there was nothing to
		// remove — a different world from a removal that happened, and
		// only one of them means an agent's tree was taken down.
		//
		// False and satisfactory at once, so no reason is attached:
		// "nothing stood" is the world rather than something wrong with
		// the reading, and a fact carrying a reason is dropped from the
		// answer at a clear exit. The summary is what says which of the
		// two worlds this was.
		s.facts = append(s.facts, Flag("worktree_removed", false, wtInv))
	} else {
		removeRec, ok := report.Record(removeSteps.Answer())
		if !ok {
			return s, report.Records, RefuseUnrepairable("the removal step produced no record")
		}
		if removeRec.Failed() {
			return refusedRemoval(dispatchID, dispatchBranch, standing, removeRec), report.Records, nil
		}
		s.facts = append(s.facts, Flag("worktree_removed", true, invocationOf(removeSteps[0], removeRec)))
		s.facts = append(s.facts, Text("worktree_path", standing.Path, wtInv))
	}

	deleteRec, ok := report.Record(deleteSteps.Answer())
	if !ok {
		return s, report.Records, RefuseUnrepairable("the branch-deletion step produced no record")
	}
	if deleteRec.Failed() {
		// The branch outlived its worktree, which is the safe half of
		// the pair: nothing was lost, and a branch left standing is
		// recoverable where a deleted one is not.
		// [vcs.VersionControl.DeleteBranch] does not force, so a branch
		// carrying commits nothing else holds refuses deletion rather
		// than taking them with it — and that refusal reaches the caller
		// here rather than being retried away.
		s.facts = append(s.facts, Flag("branch_deleted", false, invocationOf(deleteSteps[0], deleteRec)))
		s.outcome = OutcomeAttention
		s.summary = fmt.Sprintf("%s could not be deleted: %s", dispatchBranch, describeFailure(deleteRec))
		if standing != nil {
			s.summary = fmt.Sprintf("%s's worktree at %s was removed and %s could not be deleted: %s",
				dispatchID, standing.Path, dispatchBranch, describeFailure(deleteRec))
		}
		s.resolution = fmt.Sprintf(
			"read what the branch still carries, and delete it by hand once nothing is lost by it: git log %s --not HEAD",
			dispatchBranch)
		return s, report.Records, nil
	}
	s.facts = append(s.facts, Flag("branch_deleted", true, invocationOf(deleteSteps[0], deleteRec)))

	s.summary = fmt.Sprintf("no worktree stood on %s, and it was deleted", dispatchBranch)
	if standing != nil {
		s.summary = fmt.Sprintf("%s's worktree at %s was removed and %s deleted",
			dispatchID, standing.Path, dispatchBranch)
	}
	return s, report.Records, nil
}

// refusedRemoval is the answer for a removal version control declined —
// AC-SWARM-DISP-004.10.
//
// Attention rather than stop, and no retry of any kind: the delta forbids
// the forcing form, and [vcs.VersionControl.RemoveWorktree] offers none,
// so the only thing left to do is name what the refusal was over and
// stop short of the deletion, which the plan already did.
//
// A LOCKED worktree is the one cause with a specific resolution, and it
// is named as the unlock invocation. Never `--force`: forcing discards
// the uncommitted work a lock exists to protect, and a run under time
// pressure is the caller most likely to take a resolution at its word.
func refusedRemoval(dispatchID, dispatchBranch string, standing *vcs.Worktree, rec plan.Record) settlement {
	inv := argvOf(rec)
	s := settlement{outcome: OutcomeAttention}

	if standing.Locked {
		reason := standing.LockReason
		if reason == "" {
			reason = "no reason was recorded"
		}
		s.summary = fmt.Sprintf(
			"removing %s's worktree at %s was refused and it is locked (%s); %s was not deleted and nothing was forced",
			dispatchID, standing.Path, reason, dispatchBranch)
		s.resolution = fmt.Sprintf("unlock it and accept the dispatch again: git worktree unlock %s", standing.Path)
	} else {
		s.summary = fmt.Sprintf(
			"removing %s's worktree at %s was refused (%s); %s was not deleted and nothing was forced",
			dispatchID, standing.Path, describeFailure(rec), dispatchBranch)
		s.resolution = fmt.Sprintf(
			"read what the tree still holds and settle it, then accept the dispatch again: git -C %s status --porcelain",
			standing.Path)
	}

	// The removal is reported as READ AND WRONG rather than as a plain
	// false, so a reader of the answer alone lands on the property that
	// failed and its own resolution, instead of on a value that looks
	// like an ordinary "no".
	refused := Invalid("worktree_removed", flagValue(false), s.summary, inv)
	refused.Resolution = s.resolution
	s.facts = []Fact{
		refused,
		Text("worktree_path", standing.Path, inv),
		Flag("worktree_locked", standing.Locked, inv),
		Flag("branch_deleted", false, inv),
	}
	return s
}

// flagValue is a yes-or-no fact value, for the one constructor that takes
// a value rather than building one.
func flagValue(v bool) swarmmodel.FactValue {
	return swarmmodel.FactValue{Kind: swarmmodel.FactKindFlag, Flag: &v}
}

// recheckDispatchBeads re-checks every bead the dispatch holds, with no
// signal deferred — AC-SWARM-DISP-004.11.
//
// The signal set is the run's own, and the two a merge decides are the
// reason this runs HERE rather than before the fold: whether the bead's
// work is on the target, and whether the worktree it was done in is gone.
// The other three read the same before and after.
//
//	closed     the bead is finished in the tracker
//	committed  a commit ON THE TARGET carries its id
//	merged     the dispatch branch is an ancestor of the target
//	worktree   nothing stands on the dispatch branch any more
//	findings   at least one comment records a finding
//
// `committed` reads the TARGET rather than the dispatch branch, for two
// reasons that both only hold after the fold: the branch has just been
// deleted, and a concurrent dispatch's commits were rewritten by the
// rebase, so the branch's own shas are not the ones that landed. Before
// the fold the same signal would have to read the branch — which is what
// makes it one of the two a merge decides.
//
// `merged` and `worktree` are properties of the dispatch rather than of
// one bead, so they are measured once and reported once; a per-bead
// re-reading of them would be the same answer several times over.
func recheckDispatchBeads(
	ctx context.Context, tr tracker.Tracker, vc vcs.VersionControl, pr plan.Runner,
	req Request, target, dispatchID, dispatchBranch string, ancestor bool, ancestorInv []string,
) (settlement, []plan.Record, error) {
	var records []plan.Record
	s := settlement{outcome: OutcomeClear}

	beneath, recs, err := tracker.Descendants(ctx, tr, pr, req.Run)
	records = append(records, recs...)
	if err != nil {
		return s, records, RefuseUnrepairable("%s", err)
	}
	beads := dispatchBeads(beneath, dispatchID)
	beadsInv := invocationsOf(recs)

	// The worktree signal, taken from state rather than from the removal
	// having reported success: "the removal exited 0" and "no worktree
	// stands" are different claims, and only the second is the signal.
	stillStanding, wtInv, recs, err := standingWorktree(ctx, vc, pr, req, dispatchBranch)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}

	committed, commitInv, recs, err := closedBeadsOn(ctx, vc, pr, req, target)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}

	findings, recs, err := findingsFor(ctx, tr, pr, req, beads)
	records = append(records, recs...)
	if err != nil {
		return s, records, err
	}

	var failures []string
	for _, b := range beads {
		if b.State != tracker.StateDone {
			failures = append(failures, b.ID+": closed")
		}
		if !committed[b.ID] {
			failures = append(failures, b.ID+": committed")
		}
		if !ancestor {
			failures = append(failures, b.ID+": merged")
		}
		if stillStanding != nil {
			failures = append(failures, b.ID+": worktree")
		}
		if !findings[b.ID] {
			failures = append(failures, b.ID+": findings")
		}
	}
	sort.Strings(failures)

	inv := [][]string{}
	inv = append(inv, beadsInv...)
	if commitInv != nil {
		inv = append(inv, commitInv)
	}
	if wtInv != nil {
		inv = append(inv, wtInv)
	}
	if ancestorInv != nil {
		inv = append(inv, ancestorInv)
	}

	s.facts = append(s.facts, Count("beads_checked", len(beads), inv...))

	failed := List("acceptance_failures", failures, inv...)
	if len(failures) > 0 {
		s.outcome = OutcomeAttention
		s.summary = fmt.Sprintf(
			"%d post-merge signal(s) failed across %s's %d bead(s) (%s); no push was attempted",
			len(failures), dispatchID, len(beads), strings.Join(failures, ", "))
		s.resolution = "put each named signal right — close the bead, or record the finding it is missing — then accept the dispatch again"
		// Read and wrong, rather than read: an empty list and a
		// non-empty one are the same shape, and only the reason tells a
		// reader which of them they are holding.
		failed.Reason = s.summary
		failed.Resolution = s.resolution
		s.facts = append(s.facts, failed)
		return s, records, nil
	}
	s.facts = append(s.facts, failed)
	s.summary = fmt.Sprintf("all %d of %s's bead(s) passed with no signal deferred", len(beads), dispatchID)
	return s, records, nil
}

// dispatchBeads narrows a run's descendants to the work one dispatch
// holds.
//
// The same filters the drain check applies, because the set that had to
// drain is the set that is re-checked: an epic is not work, and a bead
// awaiting a human answer was never this agent's to finish.
func dispatchBeads(beneath []tracker.Bead, dispatchID string) []tracker.Bead {
	var out []tracker.Bead
	for _, b := range beneath {
		if b.Assignee != dispatchID || b.Kind == epicKind {
			continue
		}
		if hasLabel(b, humanLabel) {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func hasLabel(b tracker.Bead, label string) bool {
	for _, l := range b.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// closedBeadsOn reads the beads whose close commits sit on a branch.
//
// One bounded read rather than one per bead: the subjects a run's close
// commits carry are read through [status.BeadIDFromCloseSubject], which
// is this run's ONE reading of a close commit's subject.
func closedBeadsOn(
	ctx context.Context, vc vcs.VersionControl, pr plan.Runner, req Request, branch string,
) (map[string]bool, []string, []plan.Record, error) {
	steps := vc.Commits("accepted-commits", branch, acceptanceScan)
	report, err := pr.Run(ctx, plan.Plan{
		Name:  "the beads closed on " + branch,
		Steps: steps,
	}, req.Mode)
	if err != nil {
		return nil, nil, report.Records, RefuseUnrepairable("%s", err)
	}
	rec, ok := report.Record(steps.Answer())
	if !ok || rec.Failed() {
		return nil, nil, report.Records, RefuseUnrepairable(
			"the commits on %s could not be read: %s", branch, describeFailure(rec))
	}
	commits, ok := rec.Value.([]vcs.Commit)
	if !ok {
		return nil, nil, report.Records, RefuseUnrepairable(
			"reading the commits on %s produced %T, not a commit set", branch, rec.Value)
	}
	closed := map[string]bool{}
	for _, c := range commits {
		if id := status.BeadIDFromCloseSubject(c.Subject); id != "" {
			closed[id] = true
		}
	}
	return closed, argvOf(rec), report.Records, nil
}

// findingsFor reads each bead's findings and reports which of them
// recorded one in the run's convention.
//
// A comment that is not a finding does not count: the signal is "the work
// was recorded", and counting every comment would let a bead pass on a
// note somebody left in passing.
func findingsFor(
	ctx context.Context, tr tracker.Tracker, pr plan.Runner, req Request, beads []tracker.Bead,
) (map[string]bool, []plan.Record, error) {
	recorded := map[string]bool{}
	if len(beads) == 0 {
		return recorded, nil, nil
	}

	steps := make([]plan.Step, 0, len(beads))
	names := make(map[string]string, len(beads))
	for i, b := range beads {
		name := fmt.Sprintf("findings-%d", i)
		names[name] = b.ID
		steps = append(steps, tr.Findings(name, b.ID))
	}
	report, err := pr.Run(ctx, plan.Plan{
		Name:  "the findings recorded on each of the dispatch's beads",
		Steps: steps,
	}, req.Mode)
	if err != nil {
		return nil, report.Records, RefuseUnrepairable("%s", err)
	}
	for _, rec := range report.Records {
		id, ok := names[rec.Step]
		if !ok {
			continue
		}
		if rec.Failed() {
			return nil, report.Records, RefuseUnrepairable(
				"the findings recorded on %s could not be read: %s", id, describeFailure(rec))
		}
		found, err := tracker.FindingsFrom(rec.Value)
		if err != nil {
			return nil, report.Records, RefuseUnrepairable("%s: %s", id, err)
		}
		for _, f := range found {
			if strings.HasPrefix(strings.TrimSpace(f.Text), findingPrefix) {
				recorded[id] = true
				break
			}
		}
	}
	return recorded, report.Records, nil
}

// landTarget pushes the sprint branch and proves it landed, which is
// `swarm branch push`'s implementation and not a second one — the delta
// requires this command to END by calling it.
//
// WHERE --no-push is passed every step before the push has already run,
// and the branch is reported as UNLANDED rather than pushed. Unlanded is
// the truth and the fact says so; the outcome stays clear because the
// caller asked for exactly this, and a flag whose use always exited
// non-zero would make a coordinator routing on the status read every
// accepted dispatch as needing attention.
func landTarget(
	ctx context.Context, vc vcs.VersionControl, runner swarmexec.Runner,
	req Request, target string, noPush bool,
) (settlement, []plan.Record, error) {
	if noPush {
		// The fact's invocation is the landing CHECK rather than the
		// push: an invocation on a fact is what reproduces it, and the
		// push would change it.
		unlanded := Flag("landed", false,
			[]string{"swarm", "branch", "verify-landed", "--branch", target})
		unlanded.Resolution = "land it: swarm branch push --branch " + target
		return settlement{
			outcome: OutcomeClear,
			facts:   []Fact{unlanded},
			summary: target + " was not pushed (--no-push), so it is unlanded",
		}, nil, nil
	}

	landed, err := verifyLanded(ctx, vc, runner, req, target, defaultRemote, true)
	if err != nil {
		return settlement{}, landed.Records, err
	}
	return settlement{
		facts:      landed.Facts,
		summary:    landed.Summary,
		outcome:    landed.Outcome,
		resolution: landed.Resolution,
	}, landed.Records, nil
}
