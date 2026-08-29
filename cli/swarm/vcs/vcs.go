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
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// Steps is an ordered fragment of a plan whose LAST step's value is the
// fragment's answer.
//
// Every operation here returns one, because one of them genuinely needs
// two invocations — reading the standing worktrees means clearing stale
// registrations first — and an interface where one method returns a
// fragment and the rest return single steps is one a caller has to
// remember the shape of. The Tracker's operations are each a single
// invocation and return a bare step; if one ever needs two, this is the
// type it takes.
type Steps []plan.Step

// Answer is the name of the step carrying the fragment's answer.
func (s Steps) Answer() string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1].Name
}

// ValueFrom reads the fragment's answer out of a report.
func (s Steps) ValueFrom(report plan.Report) (any, error) {
	name := s.Answer()
	if name == "" {
		return nil, fmt.Errorf("vcs: an empty fragment has no answer")
	}
	record, ok := report.Record(name)
	if !ok {
		return nil, fmt.Errorf("vcs: the report has no record for step %q", name)
	}
	if record.Err != nil {
		return nil, fmt.Errorf("vcs: step %q: %w", name, record.Err)
	}
	if record.Value == nil {
		return nil, fmt.Errorf("vcs: step %q produced no value (%s)", name, record.Status)
	}
	return record.Value, nil
}

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
	Merge(prefix, branch, message string) Steps
}
