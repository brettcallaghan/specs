// Package git implements [vcs.VersionControl] over the `git` command
// line, with a decoder per subcommand — per the Git Version Control
// component in specs/architecture/container-swarm-cli.md.
//
// This is the only component that knows the word `git`. Everything above
// it asks in the run's vocabulary, so a git behaviour that changes
// changes one decoder here.
//
// Every operation returns plan steps: argv vectors and the decoders for
// what comes back. Nothing here starts a process.
package git

import (
	"fmt"
	"strconv"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// executable is the tool this implementation is coupled to. It is the one
// place the name appears, and a test asserts that no other non-test
// source file under cli/swarm carries it.
const executable = "git"

// VersionControl is [vcs.VersionControl] over `git`.
//
// Git is the executable to invoke, empty meaning the default. Dir is the
// repository or worktree to invoke it in, empty meaning the process's own
// directory.
type VersionControl struct {
	Git string
	Dir string
}

var _ vcs.VersionControl = VersionControl{}

func (v VersionControl) tool() string {
	if v.Git != "" {
		return v.Git
	}
	return executable
}

func (v VersionControl) step(name string, effect plan.Effect, repeatable bool, decode plan.Decoder, args ...string) plan.Step {
	as := make([]plan.Arg, len(args))
	for i, a := range args {
		as[i] = plan.Lit(a)
	}
	return plan.Step{
		Name:       name,
		Executable: v.tool(),
		Args:       as,
		Dir:        v.Dir,
		Decode:     decode,
		Effect:     effect,
		Repeatable: repeatable,
	}
}

// Worktrees clears stale registrations and then reads the standing ones.
//
// Two invocations, in that order, because a registration whose directory
// has been deleted is still listed at exit 0 — measured, with a
// `prunable gitdir file points to non-existent location` line — and reads
// as a live worktree to everything above. The prune is a write, so a dry
// run refuses it; the list's decoder drops prunable entries as well, which
// is what keeps the answer right in that one mode.
func (v VersionControl) Worktrees(prefix string) vcs.Steps {
	prune := v.step(prefix+"-prune", plan.Write, true, discard,
		"worktree", "prune")
	prune.Description = "clear registrations whose working tree is gone"

	list := v.step(prefix+"-list", plan.Read, true, decodeWorktreeList,
		"worktree", "list", "--porcelain")
	list.Description = "the standing working trees"

	return vcs.Steps{prune, list}
}

// Root reads the repository's working-tree root.
func (v VersionControl) Root(prefix string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeRoot,
		"rev-parse", "--show-toplevel")
	s.Description = "the repository's working-tree root"
	return vcs.Steps{s}
}

// Commits reads the most recent commits on a revision, newest first.
//
// The format is asked for with NUL separators rather than parsed out of a
// rendered log: a commit subject may contain anything, and this run's own
// subjects carry a bead id, a priority and a title.
func (v VersionControl) Commits(prefix, revision string, limit int) vcs.Steps {
	args := []string{"log", "--format=%H%x00%ct%x00%s"}
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	args = append(args, revision)

	s := v.step(prefix, plan.Read, true, decodeCommits, args...)
	s.Description = "the most recent commits on " + revision
	return vcs.Steps{s}
}

// Branches reads the local branches and where each points.
func (v VersionControl) Branches(prefix string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeBranches,
		"for-each-ref", "--format=%(refname:short)%00%(objectname)%00%(upstream:short)", "refs/heads/")
	s.Description = "the local branches"
	return vcs.Steps{s}
}

// Descends reports whether ancestor is an ancestor of descendant.
//
// The answer is the exit status: 0 yes, 1 no, 128 when a revision does
// not resolve. The step declares 1 as an answer so a legitimate "no" does
// not trip the ordering guard, and leaves 128 a failure so an
// unresolvable revision is not silently reported as "no" — which is the
// answer that would let a dispatch branch cut from the wrong base pass
// its own guard.
func (v VersionControl) Descends(prefix, ancestor, descendant string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeAncestry,
		"merge-base", "--is-ancestor", ancestor, descendant)
	s.Answers = []int{1}
	s.Description = fmt.Sprintf("is %s an ancestor of %s", ancestor, descendant)
	return vcs.Steps{s}
}

// Resolve turns a revision into the commit it names.
func (v VersionControl) Resolve(prefix, revision string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeRevision,
		"rev-parse", "--verify", revision+"^{commit}")
	s.Description = "resolve " + revision
	return vcs.Steps{s}
}

// CurrentBranch reads the branch the repository is on.
//
// `--abbrev-ref` answers the short name, and answers the literal string
// `HEAD` when the repository is on no branch — measured on git 2.43.0:
// `git rev-parse --abbrev-ref HEAD` prints `main` on a branch and `HEAD`
// after `git checkout --detach`, exit 0 both times. A caller has to read
// that literal as "no branch" rather than as a branch called HEAD, which
// is why the answer is documented on the interface.
func (v VersionControl) CurrentBranch(prefix string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeRoot,
		"rev-parse", "--abbrev-ref", "HEAD")
	s.Description = "the branch the repository is on"
	return vcs.Steps{s}
}

// Divergence compares a local branch against its counterpart on a named
// remote.
//
// Against `<remote>/<branch>` rather than `<branch>@{upstream}`, because
// the caller named the remote and the two can differ: measured, a branch
// with no upstream configured makes the upstream form exit 128, while
// `rev-list --count --left-right main...origin/main` answers `0\t0` for
// the same repository.
//
// A counterpart that does not exist makes this fail rather than answer
// zero: measured, `git rev-list --count --left-right
// claude/sprint...origin/claude/sprint` with no such remote-tracking ref
// exits 128 with `fatal: ambiguous argument`, and reporting that as "in
// step with its remote" would be the silent wrong answer.
func (v VersionControl) Divergence(prefix, branch, remote string) vcs.Steps {
	s := v.step(prefix, plan.Read, true, decodeDivergence,
		"rev-list", "--count", "--left-right", branch+"..."+remote+"/"+branch)
	s.Description = "how far " + branch + " stands from " + remote + "/" + branch
	return vcs.Steps{s}
}

// Merge folds a branch into the current one and reads back the commit it
// made.
//
// --no-ff so the merge is always a commit a run can find; a fast-forward
// would leave nothing naming what was merged. The merge stops the plan on
// failure, which is the ordering guard the whole container is built
// around: a worktree removal chained after a merge that exited non-zero
// is the incident this design exists to prevent.
//
// Re-running is not safe: a second merge of the same branch either
// reports nothing to do or conflicts, and neither is what the caller
// asked for.
func (v VersionControl) Merge(prefix, branch, message string) vcs.Steps {
	merge := v.step(prefix+"-merge", plan.Write, false, decodeMerge,
		"merge", "--no-ff", "-m", message, branch)
	merge.Description = "fold " + branch + " in"

	read := v.step(prefix+"-commit", plan.Read, true, decodeMerged,
		"log", "--format=%H%x00%ct%x00%s", "-n", "1", "HEAD")
	read.Description = "the merge commit"

	return vcs.Steps{merge, read}
}

// CreateWorktree stands a working tree up at path, on a new branch cut
// from base.
//
// Re-running is not safe. Measured on git 2.43.0: a second
// `git worktree add -b dev-1 <other-path> main` exits 255 with
// `fatal: a branch named 'dev-1' already exists` — the branch creation is
// what makes it single-shot, and that is the property worth keeping,
// since it is also what stops two dispatches sharing a branch.
func (v VersionControl) CreateWorktree(prefix, path, branch, base string) vcs.Steps {
	s := v.step(prefix, plan.Write, false, discard,
		"worktree", "add", "-b", branch, path, base)
	s.Description = "stand a worktree up at " + path + " on " + branch + ", cut from " + base
	return vcs.Steps{s}
}

// RemoveWorktree removes the working tree standing at path.
//
// No --force, and no parameter that could add one. The forcing form
// discards uncommitted work, and a dispatch's worktree is where an
// unfinished agent's work lives; a tree with work in it refusing removal
// is how an orchestrator learns the dispatch is unfinished.
//
// Re-running is not safe. Measured: removing an already-removed worktree
// exits 128 with `fatal: '<path>' is not a working tree`.
func (v VersionControl) RemoveWorktree(prefix, path string) vcs.Steps {
	s := v.step(prefix, plan.Write, false, discard,
		"worktree", "remove", path)
	s.Description = "remove the worktree at " + path
	return vcs.Steps{s}
}

// CreateBranch creates a branch at base without checking it out.
//
// Re-running is not safe. Measured: a second `git branch br-1 main`
// exits 128 with `fatal: a branch named 'br-1' already exists`.
func (v VersionControl) CreateBranch(prefix, name, base string) vcs.Steps {
	s := v.step(prefix, plan.Write, false, discard,
		"branch", name, base)
	s.Description = "create " + name + " at " + base
	return vcs.Steps{s}
}

// DeleteBranch deletes a merged branch.
//
// -d rather than -D: -D would take unmerged commits with it, and a
// dispatch branch that is not merged is one whose work has not landed.
//
// Re-running is not safe. Measured: `git branch -d br-1` exits 0 the
// first time (`Deleted branch br-1 (was 863ebb2).`) and 1 the second
// (`error: branch 'br-1' not found`).
func (v VersionControl) DeleteBranch(prefix, name string) vcs.Steps {
	s := v.step(prefix, plan.Write, false, discard,
		"branch", "-d", name)
	s.Description = "delete " + name
	return vcs.Steps{s}
}

// Rebase replays branch onto a named commit.
//
// Re-running is not safe, and the reason is the failure rather than the
// success: a rebase that conflicts leaves the repository mid-rebase, and
// a second invocation there does not start over. Recovery is resolving or
// aborting, which is a decision this container does not make for its
// caller. (A rebase that SUCCEEDED is a no-op on re-run — measured, exit
// 0, `Current branch clean is up to date.` — but Repeatable is a claim
// about the step, and the step's failure mode is the one recovery meets.)
func (v VersionControl) Rebase(prefix, branch, onto string) vcs.Steps {
	s := v.step(prefix, plan.Write, false, decodeRebase,
		"rebase", onto, branch)
	s.Description = "replay " + branch + " onto " + onto
	return vcs.Steps{s}
}

// Checkout puts the repository on a branch.
//
// The operation the concurrent case needs between the rebase and the
// fold: `git rebase <onto> <branch>` leaves the repository on <branch>
// — measured on git 2.43.0, HEAD reads `bb/dev-1` after
// `git rebase main bb/dev-1` run from `main` — so a `merge --ff-only`
// issued from there would advance the dispatch's own branch onto itself
// and leave the target where it was.
//
// No forcing form, and no parameter that could add one. Measured: a
// checkout whose target differs in a file the working tree has modified
// exits 1 with `error: Your local changes to the following files would
// be overwritten by checkout`, which is the refusal a run needs to see
// rather than to override.
//
// Re-running IS safe. Measured: checking out the branch already checked
// out exits 0 with `Already on 'main'`.
func (v VersionControl) Checkout(prefix, branch string) vcs.Steps {
	s := v.step(prefix, plan.Write, true, discard,
		"checkout", branch)
	s.Description = "put the repository on " + branch
	return vcs.Steps{s}
}

// MergeFastForward advances the current branch to branch, refusing
// anything that is not a fast forward.
//
// Re-running IS safe. Measured: a second `git merge --ff-only <branch>`
// after a successful one exits 0 with `Already up to date.`, and a
// REFUSED one leaves nothing behind — after
// `fatal: Not possible to fast-forward, aborting.` the working tree is
// clean, HEAD is unmoved and no MERGE_HEAD or rebase directory exists.
// That is what makes this the fold a recovery can simply re-run.
func (v VersionControl) MergeFastForward(prefix, branch string) vcs.Steps {
	s := v.step(prefix, plan.Write, true, decodeFastForward,
		"merge", "--ff-only", branch)
	s.Description = "fast-forward onto " + branch
	return vcs.Steps{s}
}

// Fetch updates the remote-tracking refs for a remote.
//
// A write: refs move, so a dry run refuses it. Re-running is safe —
// measured, two consecutive `git fetch origin` both exit 0.
func (v VersionControl) Fetch(prefix, remote string) vcs.Steps {
	s := v.step(prefix, plan.Write, true, discard,
		"fetch", remote)
	s.Description = "fetch " + remote
	return vcs.Steps{s}
}

// Push publishes a branch to a remote.
//
// No --force: measured, a push the remote refuses exits 1 with
// ` ! [rejected]        main -> main (non-fast-forward)`, which is a
// condition to report rather than to overwrite — the remote holds
// something this branch is not built on. A server refusing the ref
// outright reads the same way: measured against a bare repository whose
// pre-receive hook declines `refs/heads/claude/*`, exit 1 with
// `remote: refusing refs/heads/claude/*: …`,
// ` ! [remote rejected] claude/sprint -> claude/sprint (pre-receive hook declined)`
// and `error: failed to push some refs to '../remote.git'` — all on
// STDERR, with stdout empty.
//
// Re-running is safe. Measured: a second `git push origin main` with
// nothing new exits 0 with `Everything up-to-date`.
func (v VersionControl) Push(prefix, remote, branch string) vcs.Steps {
	s := v.step(prefix, plan.Write, true, pushDecoder(remote),
		"push", remote, branch)
	s.Description = "push " + branch + " to " + remote
	return vcs.Steps{s}
}

// discard is the decoder for a step whose output carries nothing the run
// reads. It still fails on a non-zero status, so the step is guarded like
// any other.
func discard(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git", res)
	}
	return nil, nil
}
