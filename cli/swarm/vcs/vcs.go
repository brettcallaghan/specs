// Package vcs states what a run needs from version control, in the run's
// own vocabulary, independent of which system implements it — per the
// Version Control component in
// specs/architecture/container-swarm-cli.md.
//
// The same seam as the Tracker, for the other tool. Every operation
// returns plan steps rather than performing a call, so a version-control
// read is visible to explain mode and guarded by the plan's ordering
// policy; and every operation is named for what a run means — a sprint
// branch, a dispatch's branch, a standing worktree — rather than for the
// command that happens to answer it.
package vcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// Steps is the plan fragment an operation here returns: an ordered slice
// whose last step carries the answer.
//
// An alias rather than a type of its own, so a fragment built here and
// one built anywhere else are the same value. Every operation in this
// interface returns one, because one of them genuinely needs two
// invocations — reading the standing worktrees means clearing stale
// registrations first — and an interface where one method returned a
// fragment and the rest returned single steps is one a caller has to
// remember the shape of.
type Steps = plan.Steps

// Worktree is one working tree registered against the repository.
type Worktree struct {
	// Path is where it stands on disk.
	Path string
	// Branch is the short name of the branch it has checked out, empty
	// when it is detached. A dispatch's worktree is matched to its agent
	// through this, which is why the short form is what is kept.
	Branch string
	// Head is the commit it is at.
	Head string
	// Detached is true when it has no branch.
	Detached bool
	// Locked is true when the registration is locked, with the reason
	// when one was given.
	Locked     bool
	LockReason string
}

// Commit is as much of a commit as a run needs: what it says, when it
// happened, and what it is.
type Commit struct {
	SHA string
	// When the commit was made. A run's progress measurements are ages,
	// so this is the raw material for all of them.
	When time.Time
	// Subject is the commit's first line. A run's commits carry the bead
	// id in it, which is how a bead's commit is found.
	Subject string
}

// Branch is one local branch and where it points.
type Branch struct {
	Name string
	Head string
	// Upstream is the short name of the remote-tracking branch, empty
	// when the branch has none.
	Upstream string
}

// Divergence is how far a branch stands from its remote.
type Divergence struct {
	// Ahead is the number of commits the local branch has that its
	// remote does not.
	Ahead int
	// Behind is the number the remote has that the local branch does not.
	Behind int
}

// Refusal is version control declining to fold one line of work into
// another, decoded into the run's vocabulary.
//
// It is a type rather than a message because the run branches on it: a
// dispatch whose rebase conflicts is handed back to its agent with the
// paths, and one whose merge is refused as a non-fast-forward is rebased
// and retried. A caller given the tool's text would have to re-derive
// both by matching on prose, which is the decoding this container exists
// to do once.
type Refusal struct {
	// Operation is what was refused, in the run's words: "merge" or
	// "rebase".
	Operation string
	// Conflicts are the repository-relative paths reported as
	// conflicting, in the order the tool reported them. Empty when the
	// refusal was not a conflict.
	Conflicts []string
	// NonFastForward is true when the refusal was that the fold could not
	// be a fast forward — the branch and its target have diverged.
	NonFastForward bool
	// Detail is what the tool said, kept for a report to show a human. It
	// is never the thing a caller branches on; Conflicts and
	// NonFastForward are.
	Detail string
}

func (r *Refusal) Error() string {
	switch {
	case r.NonFastForward:
		return fmt.Sprintf("%s refused: the branch and its target have diverged, so this cannot be a fast forward", r.Operation)
	case len(r.Conflicts) > 0:
		return fmt.Sprintf("%s refused: %d conflicting path(s): %s",
			r.Operation, len(r.Conflicts), strings.Join(r.Conflicts, ", "))
	default:
		return fmt.Sprintf("%s refused: %s", r.Operation, r.Detail)
	}
}

// VersionControl states what a run needs from version control.
//
// Each method takes a step-name prefix rather than a name: a fragment may
// hold more than one step, and their names have to be distinct within the
// plan they land in.
type VersionControl interface {
	// Worktrees reads the standing working trees. Stale registrations
	// are cleared first, because a registration whose directory is gone
	// still appears in the list and reads as a live worktree to
	// everything above. Answer: []Worktree.
	Worktrees(prefix string) Steps

	// Commits reads the most recent commits on a revision, newest first.
	// Answer: []Commit.
	Commits(prefix, revision string, limit int) Steps

	// Branches reads the local branches and where each points.
	// Answer: []Branch.
	Branches(prefix string) Steps

	// Descends reports whether one revision is an ancestor of another —
	// "was this branch cut from the sprint branch?", which is the guard
	// that stops a dispatch being written against a tree missing its
	// predecessors' commits. Answer: bool.
	Descends(prefix, ancestor, descendant string) Steps

	// Resolve turns a revision into the commit it names.
	// Answer: string, the full object name.
	Resolve(prefix, revision string) Steps

	// Divergence compares a local branch against its remote.
	// Answer: Divergence.
	Divergence(prefix, branch string) Steps

	// Merge folds one branch into the current one, without fast
	// forwarding, so the merge is a commit a run can find.
	// Answer: Commit, the merge commit.
	//
	// A refusal is a [*Refusal] on the step's error, not the tool's text.
	Merge(prefix, branch, message string) Steps

	// The writes the dispatch loop makes. Every one of them declares its
	// effect and whether re-running it is safe, per ADR-026: recovery
	// from a partial plan is re-running rather than undoing, so an
	// operation that cannot be re-run says so and the caller is told
	// recovery is not available for it.

	// CreateWorktree stands a working tree up at path, on a NEW branch
	// cut from base. The branch is created by this operation rather than
	// looked up, so a caller cannot attach a second worktree to a branch
	// another dispatch already holds. Answer: none.
	CreateWorktree(prefix, path, branch, base string) Steps

	// RemoveWorktree removes the working tree standing at path.
	//
	// There is deliberately NO force parameter, and the absence is the
	// guarantee rather than a default: the forcing form discards
	// uncommitted work, and a dispatch's worktree is exactly where an
	// unfinished agent's work lives. A bool here would put that discard
	// one argument away from every caller, and a run under time pressure
	// is the caller most likely to pass it. A tree with work in it is
	// meant to refuse removal — that refusal is how an orchestrator
	// learns the dispatch is unfinished. Answer: none.
	RemoveWorktree(prefix, path string) Steps

	// CreateBranch creates a branch at base without checking it out.
	// Answer: none.
	CreateBranch(prefix, name, base string) Steps

	// DeleteBranch deletes a branch that has been merged. It does not
	// force, so a branch carrying commits nothing else holds refuses
	// deletion rather than taking them with it. Answer: none.
	DeleteBranch(prefix, name string) Steps

	// Rebase replays branch onto a named commit.
	//
	// A conflict is a [*Refusal] carrying the paths. Answer: none.
	Rebase(prefix, branch, onto string) Steps

	// MergeFastForward advances the current branch to branch, and refuses
	// when that would not be a fast forward — the check a run makes
	// before landing a dispatch, since a fold that is not a fast forward
	// means the branch was not rebased onto the current target.
	//
	// A refusal is a [*Refusal] with NonFastForward set. Answer: none.
	MergeFastForward(prefix, branch string) Steps

	// Fetch updates the remote-tracking refs for a remote. It writes —
	// refs move — even though nothing in the working tree changes.
	// Answer: none.
	Fetch(prefix, remote string) Steps

	// Push publishes a branch to a remote. It does not force: a rejected
	// push means the remote holds something this branch is not built on,
	// which is a condition to report rather than to overwrite.
	// Answer: none.
	Push(prefix, remote, branch string) Steps
}
