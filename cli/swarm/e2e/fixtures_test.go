package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every fixture restores, and every world exhibits the condition it was
// recorded for.
//
// The stamp test beside this one reads fixture.toml and never opens the
// artifacts, so a corrupt bundle, an export the tracker will not import,
// or a worktree declaration naming a branch the bundle does not carry all
// pass it. Nothing else would catch them until a case in a later change
// tripped over one, and the failure would arrive attributed to that case
// rather than to the fixture.
//
// This restores each fixture through the same path a case does — the real
// git and the real tracker — and asserts the world came up.
func TestEveryFixtureRestores(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", fixturesDir, err)
	}

	var restored int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		restored++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			w := restore(t, binary, entry.Name())

			// The repository came up on the branch every case starts from.
			head := w.mustCapture(t, w.repo, "git", "rev-parse", "--abbrev-ref", "HEAD")
			if head != "claude/bead-bonanza-fixture" {
				t.Errorf("the restored repository is on %q, want the sprint branch", head)
			}

			// The tracker imported the export, rather than coming up empty
			// — an empty tracker reads as a run with no work, which is a
			// world several commands would answer about quite happily.
			out := w.mustCapture(t, w.repo, "bd", "show", w.stamp.Run, "--json")
			if !strings.Contains(out, w.stamp.Run) {
				t.Errorf("the tracker does not carry the run epic %q: %s", w.stamp.Run, out)
			}

			assertWorktreesStand(t, w)
			assertRemoteRefuses(t, w)
		})
	}

	if restored == 0 {
		t.Fatalf("no fixtures under %s", fixturesDir)
	}
}

// assertWorktreesStand: a declared worktree is registered and checked out
// on the branch it names.
func assertWorktreesStand(t *testing.T, w world) {
	t.Helper()

	if len(w.stamp.Worktree) == 0 {
		return
	}
	list := w.mustCapture(t, w.repo, "git", "worktree", "list", "--porcelain")
	for _, decl := range w.stamp.Worktree {
		if !strings.Contains(list, "branch refs/heads/"+decl.Branch) {
			t.Errorf("no worktree stands on %q; git worktree list says:\n%s", decl.Branch, list)
		}
		path := filepath.Join(w.dir, filepath.FromSlash(decl.Path))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("the worktree declared at %s is not there: %v", decl.Path, err)
		}
	}
}

// assertRemoteRefuses: a declared refusing remote really refuses.
//
// The push is made here rather than trusted, because a hook that is not
// executable, or a pattern that matches nothing, produces a world that
// silently accepts pushes — and a case written against it would pass
// while proving the opposite of what it says.
func assertRemoteRefuses(t *testing.T, w world) {
	t.Helper()

	decl := w.stamp.RefusingRemote
	if decl == nil {
		return
	}
	stdout, stderr, code := w.run(t, w.repo, "git", "push", decl.Name, "claude/bead-bonanza-fixture")
	if code == 0 {
		t.Errorf("the push to the refusing remote succeeded:\n%s\n%s", stdout, stderr)
	}
	if decl.Message != "" && !strings.Contains(stderr, decl.Message) {
		t.Errorf("the refusal does not carry the declared message %q:\n%s", decl.Message, stderr)
	}
}
