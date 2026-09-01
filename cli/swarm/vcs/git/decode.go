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
//
// The dispatch loop's writes, measured the same way:
//
//	git worktree add -b <branch> <path> <base>
//	    -> exit 0. A second add of the same branch exits 255,
//	    `fatal: a branch named 'dev-1' already exists`.
//	git worktree remove <path>
//	    -> exit 0; removing it again exits 128,
//	    `fatal: '<path>' is not a working tree`.
//	git branch <name> <base>   -> exit 0; again exit 128, `already exists`
//	git branch -d <name>       -> exit 0, `Deleted branch br-1 (was 863ebb2).`;
//	    again exit 1, `error: branch 'br-1' not found`
//	git merge --no-ff -m <m> <b>
//	    -> a conflict exits 1 with the CONFLICT lines ON STDOUT and
//	    nothing on stderr. That is the trap: a decoder reading stderr for
//	    the reason finds an empty stream.
//	git merge --ff-only <b>
//	    -> diverged branches exit 128 with `fatal: Not possible to
//	    fast-forward, aborting.` on stderr, behind several `hint:` lines.
//	    Nothing is left behind: the working tree is clean, HEAD is
//	    unmoved, and there is no MERGE_HEAD.
//	    Already merged -> exit 0, `Already up to date.`
//	git rebase <onto> <branch>
//	    -> a conflict exits 1 with the CONFLICT lines ON STDOUT and
//	    `error: could not apply ...` plus hints on stderr. The repository
//	    is left mid-rebase.
//	    Nothing to do -> exit 0, `Current branch clean is up to date.`
//	git fetch <remote>   -> exit 0, silent; a missing remote exits 128
//	git push <remote> <branch>
//	    -> exit 0; nothing new -> exit 0, `Everything up-to-date`;
//	    refused -> exit 1 with ` ! [rejected]        main -> main
//	    (non-fast-forward)` on stderr.
//
// The conflict-line forms measured, which is what the path decoder is
// held to:
//
//	CONFLICT (content): Merge conflict in f.txt
//	CONFLICT (add/add): Merge conflict in sp ace.txt
//	CONFLICT (modify/delete): del.txt deleted in dev and modified in HEAD.  Version HEAD of del.txt left in tree.
//	CONFLICT (rename/rename): a.txt renamed to c.txt in HEAD and to b.txt in dev.
//
// A path is printed unquoted, spaces and all, so the path in the first
// two forms is everything after `Merge conflict in ` — never a
// whitespace split.

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

// conflictMarker introduces every conflict line git prints, on both the
// merge and the rebase paths.
const conflictMarker = "CONFLICT ("

// mergeConflictIn is the form that names a path unambiguously: whatever
// follows it, to the end of the line, IS the path — spaces included,
// because git does not quote it here.
const mergeConflictIn = "Merge conflict in "

// refusal reads a failed merge or rebase as a [vcs.Refusal], or reports
// an ordinary failure when the result is not a refusal at all.
//
// Both streams are scanned for conflict lines. Measured, git puts them on
// stdout for both subcommands and the diagnostics on stderr; scanning
// both costs nothing and cannot produce a wrong answer, since the marker
// appears in no other output either subcommand emits.
func refusal(operation string, res swarmexec.Result) error {
	lines := append(nonEmptyLines(res.Stdout), nonEmptyLines(res.Stderr)...)

	var conflictLines, paths []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, conflictMarker) {
			continue
		}
		conflictLines = append(conflictLines, trimmed)
		if path := conflictPath(trimmed); path != "" {
			paths = append(paths, path)
		}
	}

	nonFastForward := strings.Contains(string(res.Stderr), "Not possible to fast-forward")

	if len(conflictLines) == 0 && !nonFastForward {
		// Not a refusal — git could not run the operation at all, which
		// is an ordinary failure and reads better as one.
		return failure("git "+operation, res)
	}

	detail := strings.Join(conflictLines, "\n")
	if detail == "" {
		detail = lastLine(strings.TrimSpace(string(res.Stderr)))
	}
	return &vcs.Refusal{
		Operation:      operation,
		Conflicts:      paths,
		NonFastForward: nonFastForward,
		Detail:         detail,
	}
}

// conflictPath pulls the path out of one conflict line, empty when the
// line's form does not name one unambiguously.
//
// Three forms are decoded, all measured — the two above the package
// comment records and the rename. A form naming no path still reaches the
// caller: the line itself is kept in [vcs.Refusal.Detail], so a conflict
// this cannot parse is visible rather than silently absent.
func conflictPath(line string) string {
	_, rest, ok := strings.Cut(line, "): ")
	if !ok {
		return ""
	}
	if _, path, found := strings.Cut(rest, mergeConflictIn); found {
		return path
	}
	// `<path> deleted in <branch> and modified in HEAD. …` and
	// `<path> renamed to <other> in HEAD and to …` — the path is what
	// precedes the verb. A path containing the verb phrase itself would
	// be cut short; nothing in git's output distinguishes that case,
	// which is why the whole line is kept in Detail regardless.
	for _, verb := range []string{" deleted in ", " renamed to "} {
		if path, _, found := strings.Cut(rest, verb); found {
			return path
		}
	}
	return ""
}

// decodeMerge decodes `git merge --no-ff`. A refusal comes back as a
// [*vcs.Refusal] so a caller branches on which paths conflict rather than
// on git's prose; anything else is an ordinary failure.
func decodeMerge(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, refusal("merge", res)
	}
	return nil, nil
}

// decodeFastForward decodes `git merge --ff-only`, whose characteristic
// refusal is the non-fast-forward: the branch and its target diverged, so
// the dispatch was not rebased onto the current target.
func decodeFastForward(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, refusal("merge", res)
	}
	return nil, nil
}

// decodeRebase decodes `git rebase`. A conflict is a [*vcs.Refusal]
// carrying the paths; the repository is left mid-rebase, which the
// caller resolves — this container does not abort on its behalf.
func decodeRebase(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, refusal("rebase", res)
	}
	return nil, nil
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
