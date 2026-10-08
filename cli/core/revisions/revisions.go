// Package revisions reads a directory as a git commit holds it.
//
// It resolves the base revision of a comparison — the merge-base of HEAD
// and the repository's default branch, or of HEAD and a named revision —
// and materializes a repository-relative directory's tracked files at a
// commit into a fresh temporary directory, so a caller can build that
// directory exactly as it builds the working tree.
//
// The package knows git and directories, and nothing of documents,
// criteria or the index. It runs the git executable as a subprocess and
// uses only the read-only subcommands rev-parse, merge-base, ls-tree and
// cat-file: it never writes to the repository or its .git directory, and
// never changes HEAD, the index or the working tree. The only thing it
// writes is the temporary directory it creates, which its caller removes
// through [Tree.Close].
package revisions

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Class says who acts on an [Error]. The package returns classes, never
// exit statuses; the command surface maps them.
type Class int

const (
	// Invocation: the caller asked for something that cannot be done as
	// asked, such as a revision that does not resolve.
	Invocation Class = iota + 1
	// Environment: the request was well formed but the repository or the
	// machine prevents it — no work tree, no git, no default branch, no
	// common ancestor.
	Environment
)

func (c Class) String() string {
	switch c {
	case Invocation:
		return "invocation"
	case Environment:
		return "environment"
	}
	return "Class(" + strconv.Itoa(int(c)) + ")"
}

// Error is every error this package returns for a condition a caller or
// maintainer can act on. Message begins with the offending value or
// condition and says what can be changed. Any other error is a state the
// package does not expect, such as git printing output it cannot parse.
type Error struct {
	Class   Class
	Message string
}

func (e *Error) Error() string { return e.Message }

func invocation(format string, args ...any) error {
	return &Error{Class: Invocation, Message: fmt.Sprintf(format, args...)}
}

func environment(format string, args ...any) error {
	return &Error{Class: Environment, Message: fmt.Sprintf(format, args...)}
}

// DefaultBranchCandidates are the branches tried, in order, after the
// target of refs/remotes/origin/HEAD, when a comparison names no revision.
var DefaultBranchCandidates = []string{"main", "origin/main", "master", "origin/master"}

// originHEAD is the symbolic reference whose target is tried first.
const originHEAD = "refs/remotes/origin/HEAD"

// Repository is a git work tree located by [Open].
type Repository struct {
	git  string
	root string
}

// Open locates the git work tree containing dir, using the git executable
// on PATH.
func Open(dir string) (*Repository, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, environment("git is not runnable: %v; the comparison reads the base revision through the git executable, so install git or put it on PATH", err)
	}
	return open(git, dir)
}

// open is [Open] with the git executable named, so a test can name one
// that cannot run.
func open(git, dir string) (*Repository, error) {
	r := &Repository{git: git, root: dir}
	out, err := r.run("rev-parse", "--show-toplevel")
	if err != nil {
		var gerr *gitError
		if errors.As(err, &gerr) {
			return nil, environment("%s is not inside a git work tree (git: %s); the comparison needs the repository the spec root is committed in",
				dir, gerr.stderr)
		}
		return nil, err
	}
	top := strings.TrimRight(string(out), "\r\n")
	if top == "" {
		return nil, environment("%s is not inside a git work tree: git rev-parse --show-toplevel printed nothing", dir)
	}
	r.root = filepath.Clean(filepath.FromSlash(top))
	return r, nil
}

// Root is the work tree's top-level directory, cleaned and in OS form.
func (r *Repository) Root() string { return r.root }

// Base is the base revision of a comparison.
type Base struct {
	// Ref is the reference the base was taken against: the revision as
	// the caller named it, or the default branch found (such as "main" or
	// "origin/main").
	Ref string
	// Commit is the merge-base of HEAD and Ref, as a full object name.
	Commit string
}

// Base resolves the base of a comparison against revision: the merge-base
// of HEAD and revision, or, when revision is empty, of HEAD and the
// default branch — the first of origin/HEAD's target and
// [DefaultBranchCandidates] that exists.
func (r *Repository) Base(revision string) (Base, error) {
	ref, commit, err := r.target(revision)
	if err != nil {
		return Base{}, err
	}
	if _, ok, err := r.commitOf("HEAD"); err != nil {
		return Base{}, err
	} else if !ok {
		return Base{}, environment("HEAD names no commit (the current branch has no commits yet), so it has no merge-base with %s; commit first, or run the comparison from a branch that has history", ref)
	}

	out, err := r.run("merge-base", "HEAD", commit)
	if err != nil {
		var gerr *gitError
		if !errors.As(err, &gerr) || gerr.code != 1 {
			return Base{}, err
		}
		shallow, serr := r.run("rev-parse", "--is-shallow-repository")
		if serr != nil {
			return Base{}, serr
		}
		if strings.TrimSpace(string(shallow)) == "true" {
			return Base{}, environment("HEAD and %s have no common ancestor in this clone: the clone is shallow, so its history stops before their merge-base; deepen the fetch (git fetch --unshallow, or --deepen) or name a base with --since", ref)
		}
		return Base{}, environment("HEAD and %s have no common ancestor: their histories are unrelated; name a base that HEAD descends from with --since", ref)
	}
	base := strings.TrimSpace(string(out))
	if base == "" {
		return Base{}, environment("HEAD and %s have no common ancestor: git merge-base printed nothing", ref)
	}
	return Base{Ref: ref, Commit: base}, nil
}

// target resolves the revision a base is taken against: the named one, or
// the default branch.
func (r *Repository) target(revision string) (ref, commit string, err error) {
	if revision != "" {
		if strings.HasPrefix(revision, "-") {
			return "", "", invocation("%s does not name a revision: a revision cannot begin with \"-\"", revision)
		}
		commit, ok, err := r.commitOf(revision)
		if err != nil {
			return "", "", err
		}
		if !ok {
			return "", "", invocation("%s does not resolve to a commit in %s; name a branch, a tag, a commit or an expression such as HEAD~3", revision, r.root)
		}
		return revision, commit, nil
	}

	tried := make([]string, 0, 1+len(DefaultBranchCandidates))
	candidates := make([]string, 0, 1+len(DefaultBranchCandidates))
	tried = append(tried, "origin/HEAD")
	if out, err := r.run("rev-parse", "--abbrev-ref", originHEAD); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" && name != originHEAD {
			candidates = append(candidates, name)
		}
	} else if !isGitError(err) {
		return "", "", err
	}
	tried = append(tried, DefaultBranchCandidates...)
	candidates = append(candidates, DefaultBranchCandidates...)

	for _, candidate := range candidates {
		commit, ok, err := r.commitOf(candidate)
		if err != nil {
			return "", "", err
		}
		if ok {
			return candidate, commit, nil
		}
	}
	return "", "", environment("no default branch found in %s: tried %s and none resolves to a commit; name a base explicitly with --since",
		r.root, strings.Join(tried, ", "))
}

// commitOf resolves rev to a commit's full object name, reporting false
// when it does not resolve.
func (r *Repository) commitOf(rev string) (string, bool, error) {
	out, err := r.run("rev-parse", "--verify", "--quiet", rev+"^{commit}")
	if err != nil {
		if isGitError(err) {
			return "", false, nil
		}
		return "", false, err
	}
	oid := strings.TrimSpace(string(out))
	return oid, oid != "", nil
}

// Tree is a directory materialized at a commit.
type Tree struct {
	root string
}

// Root is the temporary directory standing in for the repository's
// top-level directory: a file the commit holds at <dir>/a.md is at
// filepath.Join(Root(), dir, "a.md").
func (t *Tree) Root() string { return t.root }

// Close removes the temporary directory and everything in it.
func (t *Tree) Close() error {
	if t == nil || t.root == "" {
		return nil
	}
	err := os.RemoveAll(t.root)
	t.root = ""
	return err
}

// Materialize writes every regular file (modes 100644 and 100755) the
// commit holds beneath dir — a repository-relative, slash-separated path,
// "." for the whole tree — byte for byte from its blob, at its
// repository-relative path under a fresh temporary directory.
//
// Symbolic links (120000) and submodules (160000) are not written, and no
// checkout filter, .gitattributes rule or export-ignore applies. A
// directory the commit does not hold yields an empty tree. On any error
// the temporary directory is removed before Materialize returns.
func (r *Repository) Materialize(commit, dir string) (*Tree, error) {
	prefix, err := repoRelative(dir)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(commit, "-") || commit == "" {
		return nil, invocation("%q does not name a commit", commit)
	}
	oid, ok, err := r.commitOf(commit)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, invocation("%s does not resolve to a commit in %s", commit, r.root)
	}

	files, err := r.listFiles(oid, prefix)
	if err != nil {
		return nil, err
	}

	tmp, err := os.MkdirTemp("", "specs-base-*")
	if err != nil {
		return nil, environment("a temporary directory cannot be created: %v", err)
	}
	tree := &Tree{root: tmp}
	if err := r.writeBlobs(tmp, files); err != nil {
		_ = tree.Close()
		return nil, err
	}
	return tree, nil
}

// repoRelative cleans dir to a slash-separated path inside the work tree,
// returning "" for the whole tree.
func repoRelative(dir string) (string, error) {
	slashed := path.Clean(filepath.ToSlash(dir))
	if slashed == "." {
		return "", nil
	}
	if path.IsAbs(slashed) || filepath.IsAbs(dir) || !filepath.IsLocal(filepath.FromSlash(slashed)) {
		return "", invocation("%s is not a repository-relative directory: name it relative to the work tree's top level", dir)
	}
	return slashed, nil
}

// blobFile is one regular file to materialize.
type blobFile struct {
	path string // repository-relative, slash-separated
	oid  string
	exec bool
}

// listFiles lists the regular files commit holds beneath prefix ("" for
// the whole tree).
func (r *Repository) listFiles(commit, prefix string) ([]blobFile, error) {
	args := []string{"ls-tree", "-r", "-z", "--full-tree", commit}
	if prefix != "" {
		args = append(args, "--", prefix)
	}
	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}

	var files []blobFile
	for _, record := range bytes.Split(out, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		// <mode> SP <type> SP <object> TAB <path>
		meta, name, found := bytes.Cut(record, []byte{'\t'})
		if !found {
			return nil, fmt.Errorf("git ls-tree printed a record with no path: %q", record)
		}
		fields := strings.Fields(string(meta))
		if len(fields) != 3 {
			return nil, fmt.Errorf("git ls-tree printed a malformed record: %q", record)
		}
		mode, kind, oid, file := fields[0], fields[1], fields[2], string(name)
		if kind != "blob" || (mode != "100644" && mode != "100755") {
			continue // symbolic links, submodules and anything else
		}
		if prefix != "" && !strings.HasPrefix(file, prefix+"/") {
			continue // the pathspec matched the directory's name as a file
		}
		if !filepath.IsLocal(filepath.FromSlash(file)) {
			return nil, fmt.Errorf("git ls-tree listed %q, which is not a path inside the tree", file)
		}
		files = append(files, blobFile{path: file, oid: oid, exec: mode == "100755"})
	}
	return files, nil
}

// writeBlobs reads every file's blob through one git cat-file --batch
// process and writes it beneath root.
func (r *Repository) writeBlobs(root string, files []blobFile) (err error) {
	if len(files) == 0 {
		return nil
	}
	cmd := r.command("cat-file", "--batch")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("git cat-file --batch: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("git cat-file --batch: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return r.notRunnable(err)
	}
	defer func() {
		_ = stdin.Close()
		if err != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		werr := cmd.Wait()
		if err == nil && werr != nil {
			err = fmt.Errorf("git cat-file --batch: %v: %s", werr, strings.TrimSpace(stderr.String()))
		}
	}()

	reader := bufio.NewReader(stdout)
	for _, f := range files {
		if _, err := io.WriteString(stdin, f.oid+"\n"); err != nil {
			return fmt.Errorf("git cat-file --batch: requesting %s: %w", f.path, err)
		}
		content, err := readBlob(reader, f)
		if err != nil {
			return err
		}
		if err := writeFile(root, f, content); err != nil {
			return err
		}
	}
	return nil
}

// readBlob reads one cat-file --batch response:
// "<oid> <type> <size>\n<bytes>\n", or "<oid> missing\n".
func readBlob(reader *bufio.Reader, f blobFile) ([]byte, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("git cat-file --batch: reading %s: %w", f.path, err)
	}
	fields := strings.Fields(header)
	if len(fields) == 2 && fields[1] == "missing" {
		return nil, environment("%s: its blob %s is missing from the repository's object store", f.path, f.oid)
	}
	if len(fields) != 3 || fields[1] != "blob" {
		return nil, fmt.Errorf("git cat-file --batch: unexpected header for %s: %q", f.path, header)
	}
	size, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || size < 0 {
		return nil, fmt.Errorf("git cat-file --batch: unexpected size for %s: %q", f.path, header)
	}
	content := make([]byte, size)
	if _, err := io.ReadFull(reader, content); err != nil {
		return nil, fmt.Errorf("git cat-file --batch: reading %s: %w", f.path, err)
	}
	if b, err := reader.ReadByte(); err != nil || b != '\n' {
		return nil, fmt.Errorf("git cat-file --batch: %s's content is not followed by a newline", f.path)
	}
	return content, nil
}

// writeFile writes content at f's repository-relative path beneath root.
func writeFile(root string, f blobFile, content []byte) error {
	target := filepath.Join(root, filepath.FromSlash(f.path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return environment("%s cannot be written: %v", target, err)
	}
	perm := os.FileMode(0o644)
	if f.exec {
		perm = 0o755
	}
	if err := os.WriteFile(target, content, perm); err != nil {
		return environment("%s cannot be written: %v", target, err)
	}
	return nil
}

// gitError is a git invocation that ran and exited non-zero.
type gitError struct {
	args   []string
	code   int
	stderr string
}

func (e *gitError) Error() string {
	return fmt.Sprintf("git %s exited %d: %s", strings.Join(e.args, " "), e.code, e.stderr)
}

func isGitError(err error) bool {
	var gerr *gitError
	return errors.As(err, &gerr)
}

// command builds a git invocation against the work tree. Pathspecs are
// literal, so a directory name holding a glob character names only
// itself; no prompt can block; and no optional lock is taken, so even
// git's own index refresh does not write.
func (r *Repository) command(args ...string) *exec.Cmd {
	full := append([]string{"-C", r.root, "--literal-pathspecs"}, args...)
	cmd := exec.Command(r.git, full...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	return cmd
}

// run runs git and returns its standard output. A git that exits non-zero
// is a [*gitError]; one that cannot start is an environment [*Error].
func (r *Repository) run(args ...string) ([]byte, error) {
	cmd := r.command(args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return nil, &gitError{args: args, code: exit.ExitCode(), stderr: strings.TrimSpace(stderr.String())}
	}
	return nil, r.notRunnable(err)
}

func (r *Repository) notRunnable(err error) error {
	return environment("%s is not runnable: %v; the comparison reads the base revision through the git executable, so install git or put it on PATH", r.git, err)
}
