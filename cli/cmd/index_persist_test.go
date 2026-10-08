// A query or check persists the index best-effort, and a write that fails
// changes nothing the caller observes (AC-SPEC-010.7).
package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// observed is everything a caller sees of one invocation.
type observed struct {
	status int
	stdout string
	stderr string
}

// runObserved runs args with work as the working directory and returns
// its status and both streams.
func runObserved(t *testing.T, work string, args ...string) observed {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	var stdout, stderr bytes.Buffer
	status := cmd.Execute(args, &stdout, &stderr)
	return observed{status, stdout.String(), stderr.String()}
}

// A directory at the index path makes the write fail on every platform,
// including as root, where file permissions would not stop it. The same
// corpus with and without it must answer identically: same status,
// byte-identical standard output, identical standard error.
//
// AC: AC-SPEC-010.7
func TestAQueryThatCannotPersistTheIndexAnswersUnchanged(t *testing.T) {
	for _, args := range [][]string{{"check"}, {"list"}} {
		t.Run(args[0], func(t *testing.T) {
			writable := t.TempDir()
			unwritable := t.TempDir()
			for _, dir := range []string{writable, unwritable} {
				if err := copyTree(filepath.Join(casesDir, "index-default", "input"), dir); err != nil {
					t.Fatalf("copying the input corpus: %v", err)
				}
			}
			blocked := filepath.Join(unwritable, "docs", "specs.json")
			if err := os.Mkdir(blocked, 0o755); err != nil {
				t.Fatalf("creating a directory at the index path: %v", err)
			}

			want := runObserved(t, writable, args...)
			got := runObserved(t, unwritable, args...)

			// The writable copy really did persist, so the comparison is
			// between a write that succeeded and one that failed.
			if info, err := os.Stat(filepath.Join(writable, "docs", "specs.json")); err != nil || !info.Mode().IsRegular() {
				t.Fatalf("the writable copy did not persist the index: %v", err)
			}
			if info, err := os.Stat(blocked); err != nil || !info.IsDir() {
				t.Fatalf("the directory at the index path did not survive: %v", err)
			}

			if got.status != want.status {
				t.Errorf("status = %d, want %d", got.status, want.status)
			}
			if got.stdout != want.stdout {
				t.Errorf("stdout differs:\n got: %q\nwant: %q", got.stdout, want.stdout)
			}
			if got.stderr != want.stderr {
				t.Errorf("stderr = %q, want %q", got.stderr, want.stderr)
			}
			if got.stderr != "" {
				t.Errorf("stderr = %q, want it empty", got.stderr)
			}
		})
	}
}
