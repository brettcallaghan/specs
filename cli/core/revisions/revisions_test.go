// Tests for the Revisions component over throwaway repositories built with
// the git executable in t.TempDir(). TestMain isolates every git this
// package and its tests run from the machine's configuration and from any
// GIT_* variable the test process inherited (a hook running go test sets
// GIT_DIR, which would point every invocation at the wrong repository).
package revisions

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(runIsolated(m))
}

func runIsolated(m *testing.M) int {
	for _, kv := range os.Environ() {
		if name, _, _ := strings.Cut(kv, "="); strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	home, err := os.MkdirTemp("", "revisions-home-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(home)
	global := filepath.Join(home, "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	os.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	os.Setenv("GIT_CONFIG_GLOBAL", global)
	return m.Run()
}

// git runs git in dir for a test, failing the test on a non-zero exit.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return gitInput(t, dir, nil, args...)
}

func gitInput(t *testing.T, dir string, stdin []byte, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir,
		"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)
	cmd := exec.Command("git", full...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

// newRepo initializes a repository on branch main.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	return dir
}

// commit commits whatever the index holds, allowing an empty commit.
func commit(t *testing.T, dir, message string) string {
	t.Helper()
	git(t, dir, "commit", "-q", "--allow-empty", "-m", message)
	return git(t, dir, "rev-parse", "HEAD")
}

// stage records content at path in the index with the given mode, through
// hash-object --no-filters, so the blob holds exactly these bytes whatever
// .gitattributes says and whatever the platform supports (a symbolic link
// or a submodule needs no support from the file system this way).
func stage(t *testing.T, dir, mode, path string, content []byte) string {
	t.Helper()
	oid := gitInput(t, dir, content, "hash-object", "-w", "--no-filters", "--stdin")
	git(t, dir, "update-index", "--add", "--cacheinfo", mode+","+oid+","+path)
	return oid
}

func open_(t *testing.T, dir string) *Repository {
	t.Helper()
	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%s): %v", dir, err)
	}
	return repo
}

// wantClass asserts err is an *Error of class want whose message begins
// with prefix and contains every one of contains.
func wantClass(t *testing.T, err error, want Class, prefix string, contains ...string) {
	t.Helper()
	var rerr *Error
	if !errors.As(err, &rerr) {
		t.Fatalf("error = %v (%T), want a *revisions.Error", err, err)
	}
	if rerr.Class != want {
		t.Errorf("class = %v, want %v (message %q)", rerr.Class, want, rerr.Message)
	}
	if !strings.HasPrefix(rerr.Message, prefix) {
		t.Errorf("message = %q, want it to begin with %q", rerr.Message, prefix)
	}
	for _, s := range contains {
		if !strings.Contains(rerr.Message, s) {
			t.Errorf("message = %q, want it to contain %q", rerr.Message, s)
		}
	}
}

func TestOpenReportsTheTopLevelInOSForm(t *testing.T) {
	dir := newRepo(t)
	sub := filepath.Join(dir, "specs", "adrs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	repo := open_(t, sub)

	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Root() = %q, want %q", got, want)
	}
	if repo.Root() != filepath.Clean(repo.Root()) || strings.Contains(repo.Root(), "/") && filepath.Separator != '/' {
		t.Errorf("Root() = %q is not clean and in OS form", repo.Root())
	}
}

// divergedHistory builds a repository whose current branch "work" is one
// commit ahead of a shared first commit, and returns that commit.
func divergedHistory(t *testing.T) (dir, shared string) {
	t.Helper()
	dir = newRepo(t)
	shared = commit(t, dir, "shared")
	git(t, dir, "branch", "-q", "-m", "main", "work")
	commit(t, dir, "work only")
	return dir, shared
}

// branchFrom creates branch (a local or an origin/ remote-tracking name)
// at a new commit whose parent is from, so HEAD lacks the branch's tip and
// the merge-base is from itself.
func branchFrom(t *testing.T, dir, branch, from string) {
	t.Helper()
	tree := git(t, dir, "rev-parse", from+"^{tree}")
	own := git(t, dir, "commit-tree", "-p", from, "-m", branch+" only", tree)
	git(t, dir, "update-ref", qualify(branch), own)
}

// qualify spells a short branch name as the full reference the test
// creates for it.
func qualify(branch string) string {
	if rest, ok := strings.CutPrefix(branch, "origin/"); ok {
		return "refs/remotes/origin/" + rest
	}
	return "refs/heads/" + branch
}

// The default branch is the first that exists of origin/HEAD's target,
// main, origin/main, master and origin/master, and the base is the
// merge-base of HEAD and it.
//
// AC: AC-GHK-008.4, AC-GHK-008.10
func TestDefaultBranchIsOriginHEADThenMainOriginMainMasterOriginMaster(t *testing.T) {
	order := []string{"origin/trunk", "main", "origin/main", "master", "origin/master"}

	// Each case holds the candidates from position i onward, so the one
	// at position i must win over every later one.
	for i, want := range order {
		t.Run(want, func(t *testing.T) {
			dir, shared := divergedHistory(t)
			for _, branch := range order[i:] {
				branchFrom(t, dir, branch, shared)
			}
			if want == "origin/trunk" {
				git(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
			}

			base, err := open_(t, dir).Base("")
			if err != nil {
				t.Fatalf("Base(\"\"): %v", err)
			}
			if base.Ref != want {
				t.Errorf("Ref = %q, want %q", base.Ref, want)
			}
			if base.Commit != shared {
				t.Errorf("Commit = %s, want the merge-base %s", base.Commit, shared)
			}
		})
	}

	t.Run("origin/HEAD names a branch that does not exist", func(t *testing.T) {
		dir, shared := divergedHistory(t)
		branchFrom(t, dir, "master", shared)
		git(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/gone")
		base, err := open_(t, dir).Base("")
		if err != nil {
			t.Fatalf("Base(\"\"): %v", err)
		}
		if base.Ref != "master" || base.Commit != shared {
			t.Errorf("Base = %+v, want master at %s", base, shared)
		}
	})
}

// A named branch, tag, commit and expression each resolve, and the base is
// the merge-base of HEAD and that revision.
//
// AC: AC-GHK-008.5
func TestNamedRevisionResolvesToItsMergeBaseWithHEAD(t *testing.T) {
	dir := newRepo(t)
	first := commit(t, dir, "first")
	second := commit(t, dir, "second")
	git(t, dir, "tag", "v1", first)
	git(t, dir, "tag", "-a", "-m", "annotated", "v2", second)
	branchFrom(t, dir, "side", first)
	head := commit(t, dir, "third")

	cases := []struct {
		revision string
		want     string
	}{
		{"side", first},       // a branch that diverged at first
		{"v1", first},         // a lightweight tag
		{"v2", second},        // an annotated tag, peeled to its commit
		{second, second},      // a full commit
		{second[:10], second}, // an abbreviated commit
		{"HEAD~1", second},    // an expression
		{"HEAD", head},        // HEAD itself
	}
	repo := open_(t, dir)
	for _, tc := range cases {
		t.Run(tc.revision, func(t *testing.T) {
			base, err := repo.Base(tc.revision)
			if err != nil {
				t.Fatalf("Base(%q): %v", tc.revision, err)
			}
			if base.Ref != tc.revision {
				t.Errorf("Ref = %q, want the revision as named, %q", base.Ref, tc.revision)
			}
			if base.Commit != tc.want {
				t.Errorf("Commit = %s, want %s", base.Commit, tc.want)
			}
		})
	}
}

func TestUnknownRevisionIsAnInvocationErrorNamingIt(t *testing.T) {
	dir := newRepo(t)
	commit(t, dir, "first")
	repo := open_(t, dir)
	for _, revision := range []string{"no-such-branch", "HEAD~5", "--output=x"} {
		t.Run(revision, func(t *testing.T) {
			_, err := repo.Base(revision)
			wantClass(t, err, Invocation, revision)
		})
	}
}

func TestNotAWorkTreeIsAnEnvironmentError(t *testing.T) {
	dir := t.TempDir()
	_, err := Open(dir)
	wantClass(t, err, Environment, dir, "not inside a git work tree")

	t.Run("inside .git", func(t *testing.T) {
		repo := newRepo(t)
		gitDir := filepath.Join(repo, ".git")
		_, err := Open(gitDir)
		wantClass(t, err, Environment, gitDir, "not inside a git work tree")
	})
}

func TestGitNotRunnableIsAnEnvironmentErrorSayingWhich(t *testing.T) {
	dir := newRepo(t)

	t.Run("not on PATH", func(t *testing.T) {
		t.Setenv("PATH", "")
		_, err := Open(dir)
		wantClass(t, err, Environment, "git is not runnable", "PATH")
	})

	t.Run("named executable cannot run", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-git-here")
		_, err := open(missing, dir)
		wantClass(t, err, Environment, missing, "not runnable")
	})
}

func TestNoDefaultBranchIsAnEnvironmentErrorNamingEveryReferenceTried(t *testing.T) {
	dir := newRepo(t)
	commit(t, dir, "first")
	git(t, dir, "branch", "-q", "-m", "main", "trunk")

	_, err := open_(t, dir).Base("")
	wantClass(t, err, Environment, "no default branch",
		"origin/HEAD, main, origin/main, master, origin/master", "--since")
}

func TestUnrelatedHistoriesAreAnEnvironmentError(t *testing.T) {
	dir := newRepo(t)
	commit(t, dir, "main")
	git(t, dir, "checkout", "-q", "--orphan", "island")
	commit(t, dir, "island")

	_, err := open_(t, dir).Base("main")
	wantClass(t, err, Environment, "HEAD and main", "no common ancestor", "--since")
	var rerr *Error
	if errors.As(err, &rerr) && strings.Contains(rerr.Message, "shallow") {
		t.Errorf("message = %q mentions a shallow clone, and this clone is not shallow", rerr.Message)
	}
}

func TestNoMergeBaseInAShallowCloneSaysSo(t *testing.T) {
	src := newRepo(t)
	shared := commit(t, src, "shared")
	commit(t, src, "main only")
	git(t, src, "branch", "-q", "feature", shared)
	git(t, src, "checkout", "-q", "feature")
	commit(t, src, "feature only")
	git(t, src, "checkout", "-q", "main")

	clone := filepath.Join(t.TempDir(), "clone")
	url := "file://" + filepath.ToSlash(src)
	if runtime.GOOS == "windows" {
		url = "file:///" + filepath.ToSlash(src)
	}
	git(t, src, "clone", "-q", "--depth", "1", "--no-single-branch", url, clone)

	_, err := open_(t, clone).Base("origin/feature")
	wantClass(t, err, Environment, "HEAD and origin/feature", "shallow", "deepen", "--since")
}

func TestUnbornHEADIsAnEnvironmentError(t *testing.T) {
	src := newRepo(t)
	commit(t, src, "first")
	dir := newRepo(t)
	// A branch exists, HEAD's does not: the state a fresh git init that
	// has fetched but not yet committed is in.
	git(t, dir, "fetch", "-q", src, "main:refs/remotes/origin/main")

	_, err := open_(t, dir).Base("")
	wantClass(t, err, Environment, "HEAD", "no commit", "origin/main")
}

// materialized lists every file beneath root, slash-separated and
// relative to it, with its bytes.
func materialized(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			t.Errorf("%s is not a regular file (%v)", p, d.Type())
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = content
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// isolateTemp points the temporary directory at a fresh one the test can
// inspect, so it can see whether Materialize left anything behind.
func isolateTemp(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, tmp)
	}
	return tmp
}

func entries(t *testing.T, dir string) []string {
	t.Helper()
	list, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(list))
	for _, e := range list {
		names = append(names, e.Name())
	}
	return names
}

func TestMaterializeWritesTheDirectoryAsTheCommitHoldsIt(t *testing.T) {
	dir := newRepo(t)
	crlf := []byte("---\r\ntitle: CRLF\r\n---\r\n\r\nBody\r\n")
	script := []byte("#!/bin/sh\necho hi\n")
	stage(t, dir, "100644", ".gitattributes", []byte("* text eol=lf\nspecs/ignored.md export-ignore\n"))
	stage(t, dir, "100644", "README.md", []byte("outside\n"))
	stage(t, dir, "100644", "specs-other/a.md", []byte("outside, sharing a prefix\n"))
	stage(t, dir, "100644", "specs/a.md", []byte("plain\n"))
	stage(t, dir, "100644", "specs/crlf.md", crlf)
	stage(t, dir, "100755", "specs/tools/run.sh", script)
	stage(t, dir, "100644", "specs/ignored.md", []byte("export-ignored\n"))
	stage(t, dir, "100644", "specs/deep/er/b.md", []byte("nested\n"))
	stage(t, dir, "100644", "specs/[glob]/c.md", []byte("a glob character\n"))
	stage(t, dir, "120000", "specs/link.md", []byte("a.md"))
	sub := commit(t, newRepo(t), "a submodule's commit")
	git(t, dir, "update-index", "--add", "--cacheinfo", "160000,"+sub+",specs/module")
	head := commit(t, dir, "tree")

	tmp := isolateTemp(t)
	tree, err := open_(t, dir).Materialize(head, "specs")
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}

	got := materialized(t, tree.Root())
	want := map[string][]byte{
		"specs/a.md":         []byte("plain\n"),
		"specs/crlf.md":      crlf,
		"specs/tools/run.sh": script,
		"specs/ignored.md":   []byte("export-ignored\n"),
		"specs/deep/er/b.md": []byte("nested\n"),
		"specs/[glob]/c.md":  []byte("a glob character\n"),
	}
	for name, content := range want {
		if g, ok := got[name]; !ok {
			t.Errorf("%s is missing", name)
		} else if !bytes.Equal(g, content) {
			t.Errorf("%s = %q, want byte for byte %q", name, g, content)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("%s was written and is not a regular file beneath specs", name)
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(tree.Root(), "specs", "tools", "run.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o100 == 0 {
			t.Errorf("run.sh mode = %v, want it executable", info.Mode())
		}
	}

	if !strings.HasPrefix(tree.Root(), tmp) {
		t.Errorf("Root() = %q, want it under the temporary directory %q", tree.Root(), tmp)
	}
	if err := tree.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if left := entries(t, tmp); len(left) != 0 {
		t.Errorf("after Close the temporary directory holds %v, want nothing", left)
	}
}

func TestMaterializeTheWholeTree(t *testing.T) {
	dir := newRepo(t)
	stage(t, dir, "100644", "a.md", []byte("a\n"))
	stage(t, dir, "100644", "specs/b.md", []byte("b\n"))
	head := commit(t, dir, "tree")

	tree, err := open_(t, dir).Materialize(head, ".")
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	defer tree.Close()
	got := slicesKeys(materialized(t, tree.Root()))
	if want := []string{"a.md", "specs/b.md"}; !slices.Equal(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
}

func slicesKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func TestMaterializeAnAbsentDirectoryGivesAnEmptyTree(t *testing.T) {
	dir := newRepo(t)
	stage(t, dir, "100644", "README.md", []byte("no spec root yet\n"))
	stage(t, dir, "100644", "specs", []byte("a file named like the directory\n"))
	head := commit(t, dir, "tree")

	tree, err := open_(t, dir).Materialize(head, "specs")
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	defer tree.Close()
	if got := materialized(t, tree.Root()); len(got) != 0 {
		t.Errorf("files = %v, want none", slicesKeys(got))
	}
}

func TestMaterializeRemovesItsTemporaryDirectoryOnError(t *testing.T) {
	dir := newRepo(t)
	stage(t, dir, "100644", "specs/a.md", []byte("written before the failure\n"))
	// A tree entry whose blob the object store does not hold: cat-file
	// answers "missing" partway through the run.
	absent := strings.Repeat("de", 20) // no object of this name was ever written
	git(t, dir, "update-index", "--add", "--info-only", "--cacheinfo", "100644,"+absent+",specs/z.md")
	treeOID := git(t, dir, "write-tree", "--missing-ok")
	head := git(t, dir, "commit-tree", "-m", "broken", treeOID)

	tmp := isolateTemp(t)
	_, err := open_(t, dir).Materialize(head, "specs")
	wantClass(t, err, Environment, "specs/z.md", "missing")
	if left := entries(t, tmp); len(left) != 0 {
		t.Errorf("after the error the temporary directory holds %v, want nothing", left)
	}
}

func TestMaterializeRefusesADirectoryOutsideTheTree(t *testing.T) {
	dir := newRepo(t)
	head := commit(t, dir, "first")
	repo := open_(t, dir)
	for _, d := range []string{"../elsewhere", "specs/../../x", "/abs"} {
		t.Run(d, func(t *testing.T) {
			_, err := repo.Materialize(head, d)
			wantClass(t, err, Invocation, d, "repository-relative")
		})
	}
	t.Run("unknown commit", func(t *testing.T) {
		_, err := repo.Materialize("no-such", "specs")
		wantClass(t, err, Invocation, "no-such")
	})
}

// repositoryState is everything an operation of this package must leave
// as it found it.
func repositoryState(t *testing.T, dir string) string {
	t.Helper()
	return strings.Join([]string{
		git(t, dir, "rev-parse", "HEAD"),
		git(t, dir, "symbolic-ref", "-q", "HEAD"),
		git(t, dir, "status", "--porcelain", "--untracked-files=all"),
		git(t, dir, "worktree", "list", "--porcelain"),
		git(t, dir, "for-each-ref"),
	}, "\n--\n")
}

func TestOperationsLeaveTheRepositoryUntouched(t *testing.T) {
	dir := newRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "specs", "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "specs/a.md")
	first := commit(t, dir, "first")
	git(t, dir, "tag", "v1")
	git(t, dir, "branch", "-q", "-m", "main", "work")
	git(t, dir, "update-ref", "refs/remotes/origin/main", first)
	if err := os.WriteFile(filepath.Join(dir, "specs", "a.md"), []byte("modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "specs", "new.md"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commit(t, dir, "second")

	before := repositoryState(t, dir)
	repo := open_(t, dir)
	for _, revision := range []string{"", "v1", "HEAD~1", "nope"} {
		if base, err := repo.Base(revision); err == nil {
			tree, err := repo.Materialize(base.Commit, "specs")
			if err != nil {
				t.Fatalf("Materialize: %v", err)
			}
			if err := tree.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if after := repositoryState(t, dir); after != before {
		t.Errorf("repository state changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
