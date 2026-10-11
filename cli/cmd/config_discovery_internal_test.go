// Tests for discoverConfig: where the search for specs.toml starts, which
// file wins, and where the search stops.
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// tree creates every path under root; a path ending "/" is a directory,
// any other an empty file.
func tree(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(root, filepath.FromSlash(p))
		if p[len(p)-1] == '/' {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// resolvedTempDir is a temporary directory with symbolic links resolved,
// so a path the search returns compares equal to one built here.
func resolvedTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// AC: AC-SPEC-008.7
func TestDiscoveryFindsTheSpecsTomlAtTheRepositoryRootFromBelowIt(t *testing.T) {
	root := resolvedTempDir(t)
	tree(t, root, "repo/.git/", "repo/specs.toml", "repo/a/b/")

	got, err := discoverConfig(filepath.Join(root, "repo", "a", "b"))
	if err != nil {
		t.Fatalf("discoverConfig: %v", err)
	}
	if want := filepath.Join(root, "repo", "specs.toml"); got != want {
		t.Errorf("found %s, want %s", got, want)
	}
}

// AC: AC-SPEC-008.7
func TestDiscoveryPrefersTheNearestSpecsToml(t *testing.T) {
	root := resolvedTempDir(t)
	tree(t, root, "repo/.git/", "repo/specs.toml", "repo/notes/specs.toml", "repo/notes/drafts/")

	got, err := discoverConfig(filepath.Join(root, "repo", "notes", "drafts"))
	if err != nil {
		t.Fatalf("discoverConfig: %v", err)
	}
	if want := filepath.Join(root, "repo", "notes", "specs.toml"); got != want {
		t.Errorf("found %s, want %s", got, want)
	}
}

// AC: AC-SPEC-008.7, AC-SPEC-008.3
//
// A specs.toml above the repository root belongs to something else, and
// the search does not reach it. A .git file — a worktree's or a
// submodule's — marks the root exactly as a .git directory does.
func TestDiscoveryStopsAtTheRepositoryRoot(t *testing.T) {
	for name, marker := range map[string]string{"a .git directory": "outer/repo/.git/", "a .git file": "outer/repo/.git"} {
		t.Run(name, func(t *testing.T) {
			root := resolvedTempDir(t)
			tree(t, root, "outer/specs.toml", marker, "outer/repo/docs/")

			_, err := discoverConfig(filepath.Join(root, "outer", "repo", "docs"))
			var notFound *ErrConfigNotFound
			if !errors.As(err, &notFound) {
				t.Fatalf("err = %v, want *ErrConfigNotFound", err)
			}
			repo := filepath.Join(root, "outer", "repo")
			if want := []string{filepath.Join(repo, "docs"), repo}; !reflect.DeepEqual(notFound.Searched, want) {
				t.Errorf("Searched = %v, want %v", notFound.Searched, want)
			}
			if notFound.RepoRoot != repo {
				t.Errorf("RepoRoot = %q, want %q", notFound.RepoRoot, repo)
			}
			if statusOf(err) != statusEnvironment {
				t.Errorf("status = %d, want %d", statusOf(err), statusEnvironment)
			}
		})
	}
}

// AC: AC-SPEC-008.9
func TestDiscoveryOutsideARepositoryRunsToTheFilesystemRoot(t *testing.T) {
	root := resolvedTempDir(t)
	tree(t, root, "top/specs.toml", "top/unpacked/archive/")

	got, err := discoverConfig(filepath.Join(root, "top", "unpacked", "archive"))
	if err != nil {
		t.Fatalf("discoverConfig: %v", err)
	}
	if want := filepath.Join(root, "top", "specs.toml"); got != want {
		t.Errorf("found %s, want %s", got, want)
	}
}

// AC: AC-SPEC-008.3
//
// A failed search names every directory it checked and the remedy.
func TestAFailedSearchNamesEveryDirectorySearched(t *testing.T) {
	err := &ErrConfigNotFound{Searched: []string{"/a/b/c", "/a/b", "/a"}, RepoRoot: "/a"}
	want := "no specs.toml found in /a/b/c, /a/b, /a (searched up to the repository root /a); pass --config to name one"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q\nwant      %q", got, want)
	}
}

// A specs.toml that is a directory is not a configuration file, and the
// search goes on past it.
func TestDiscoverySkipsASpecsTomlThatIsNotAFile(t *testing.T) {
	root := resolvedTempDir(t)
	tree(t, root, "repo/.git/", "repo/specs.toml", "repo/sub/specs.toml/")

	got, err := discoverConfig(filepath.Join(root, "repo", "sub"))
	if err != nil {
		t.Fatalf("discoverConfig: %v", err)
	}
	if want := filepath.Join(root, "repo", "specs.toml"); got != want {
		t.Errorf("found %s, want %s", got, want)
	}
}
