package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// One decoder per subcommand, as the tracker implementation has. What
// they confine is not JSON — git emits none — but the shape each
// subcommand's output takes and the way each reports "no".
//
// Measured on git 2.43.0 in this repository, and in a scratch repository
// for the stale-registration case:
//
//	git worktree list --porcelain
//	    exit 0. Records separated by a blank line; each is `worktree
//	    <path>`, `HEAD <sha>`, and then `branch <ref>` or `detached`,
//	    plus `locked [reason]` and `prunable <reason>` where they apply.
//	    A registration whose directory has been deleted is STILL LISTED,
//	    at exit 0, with `prunable gitdir file points to non-existent
//	    location`. That is the hazard: it reads as a live worktree.
//	git worktree prune          -> exit 0, and the stale entry is gone
//	git log --format=... -n     -> exit 0; the format is asked for with
//	    NUL separators so nothing parses a rendered log
//	git merge-base --is-ancestor A B
//	    -> exit 0 yes, exit 1 no, exit 128 when a revision does not
//	    resolve. Two of those are answers and one is a failure.
//	git rev-parse <missing>
//	    -> exit 128, `fatal: ambiguous argument` on stderr — AND THE
//	    ARGUMENT ECHOED ON STDOUT. A decoder reading stdout without
//	    checking the status takes "no-such-branch" for a commit id.
//	git rev-list --count --left-right <b>...<b>@{upstream}
//	    -> exit 128 with `fatal: no upstream configured` when there is
//	    none, which is a condition rather than a crash.
//	git for-each-ref --format=... refs/heads/
//	    -> exit 0; a branch with no upstream yields an empty field.

// failure turns a non-zero result into an error naming the subcommand and
// what git said. git puts its diagnostics on stderr; the last non-empty
// line is the message.
func failure(subcommand string, res swarmexec.Result) error {
	if text := strings.TrimSpace(string(res.Stderr)); text != "" {
		return fmt.Errorf("git: %s exited %d: %s", subcommand, res.ExitStatus, lastLine(text))
	}
	return fmt.Errorf("git: %s exited %d and said nothing", subcommand, res.ExitStatus)
}

func lastLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			return trimmed
		}
	}
	return text
}

// nonEmptyLines splits output into lines, dropping empty ones. git's
// list-shaped output is one record per line and ends with a newline, so
// the trailing empty element is never data.
func nonEmptyLines(b []byte) []string {
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if trimmed := strings.TrimRight(line, "\r"); strings.TrimSpace(trimmed) != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// decodeWorktreeList decodes `git worktree list --porcelain`.
//
// A prunable registration is dropped rather than reported. Pruning runs
// first and normally removes it, but a dry run refuses the prune — it is
// a write — so the filter is what keeps the answer right in the one mode
// where the prune did not happen.
func decodeWorktreeList(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git worktree list", res)
	}

	trees := []vcs.Worktree{}
	var current vcs.Worktree
	var open, prunable bool

	flush := func() {
		if open && !prunable {
			trees = append(trees, current)
		}
		current, open, prunable = vcs.Worktree{}, false, false
	}

	for _, line := range strings.Split(string(res.Stdout), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			flush()
			continue
		}
		key, value, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			flush()
			current.Path, open = value, true
		case "HEAD":
			current.Head = value
		case "branch":
			// The porcelain form gives the full ref; a run matches a
			// dispatch to its worktree by the short branch name.
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "detached":
			current.Detached = true
		case "locked":
			current.Locked, current.LockReason = true, value
		case "prunable":
			prunable = true
		}
	}
	flush()

	return trees, nil
}

// decodeCommits decodes `git log --format=%H%x00%ct%x00%s`.
//
// A requested format with NUL separators rather than a rendered log: a
// subject may contain anything, and splitting a rendered line on
// whitespace or a dash is how a commit subject carrying one becomes two
// fields.
func decodeCommits(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git log", res)
	}

	commits := []vcs.Commit{}
	for _, line := range nonEmptyLines(res.Stdout) {
		fields := strings.SplitN(line, "\x00", 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("git: a log record has %d fields, want 3: %q", len(fields), line)
		}
		seconds, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("git: a log record's timestamp is not a number: %q", fields[1])
		}
		commits = append(commits, vcs.Commit{
			SHA: fields[0],
			// UTC, not the committer's zone: the contract forbids a field
			// whose meaning depends on the producing process's time zone.
			When:    time.Unix(seconds, 0).UTC(),
			Subject: fields[2],
		})
	}
	return commits, nil
}

// decodeBranches decodes
// `git for-each-ref --format=%(refname:short)%00%(objectname)%00%(upstream:short) refs/heads/`.
func decodeBranches(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git for-each-ref", res)
	}

	branches := []vcs.Branch{}
	for _, line := range nonEmptyLines(res.Stdout) {
		fields := strings.SplitN(line, "\x00", 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("git: a branch record has %d fields, want 3: %q", len(fields), line)
		}
		branches = append(branches, vcs.Branch{Name: fields[0], Head: fields[1], Upstream: fields[2]})
	}
	return branches, nil
}

// decodeAncestry decodes `git merge-base --is-ancestor`, whose answer is
// its exit status: 0 yes, 1 no. 128 is not an answer — the step declares
// only 1 beyond 0, so an unresolvable revision reaches the decoder as a
// failure rather than as "no".
func decodeAncestry(res swarmexec.Result) (any, error) {
	switch res.ExitStatus {
	case 0:
		return true, nil
	case 1:
		return false, nil
	default:
		return nil, failure("git merge-base --is-ancestor", res)
	}
}

// decodeRevision decodes `git rev-parse <revision>`.
//
// The status check is not a formality here: measured, rev-parse echoes
// its argument on stdout when it cannot resolve it, so a decoder reading
// stdout first would return the branch name as if it were a commit id.
func decodeRevision(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git rev-parse", res)
	}
	lines := nonEmptyLines(res.Stdout)
	if len(lines) != 1 {
		return nil, fmt.Errorf("git: rev-parse printed %d lines, want one commit id", len(lines))
	}
	return strings.TrimSpace(lines[0]), nil
}

// decodeDivergence decodes
// `git rev-list --count --left-right <branch>...<branch>@{upstream}`,
// whose output is two counts separated by a tab: local-only, then
// remote-only.
func decodeDivergence(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("git rev-list", res)
	}
	fields := strings.Fields(strings.TrimSpace(string(res.Stdout)))
	if len(fields) != 2 {
		return nil, fmt.Errorf("git: rev-list --left-right printed %d counts, want 2: %q", len(fields), res.Stdout)
	}
	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, fmt.Errorf("git: the ahead count is not a number: %q", fields[0])
	}
	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("git: the behind count is not a number: %q", fields[1])
	}
	return vcs.Divergence{Ahead: ahead, Behind: behind}, nil
}

// decodeMerged decodes the `git log -1` that follows a merge, which is
// how the merge commit is read back. The merge itself prints a summary
// this container does not parse.
func decodeMerged(res swarmexec.Result) (any, error) {
	value, err := decodeCommits(res)
	if err != nil {
		return nil, err
	}
	commits, _ := value.([]vcs.Commit)
	if len(commits) != 1 {
		return nil, fmt.Errorf("git: reading the merge commit returned %d commits, want 1", len(commits))
	}
	return commits[0], nil
}
