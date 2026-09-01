// End-to-end tests that run the real `swarm` binary as a process.
//
// The tests beside them in cli/swarm/cmd call [cmd.Execute] in-process,
// which exercises the command tree but stops one layer short of what a
// caller observes: the built binary, its argument vector, and the status
// the process exits with. That last layer is main.go, and it is the only
// place os.Exit is wired — so nothing else verifies that the status the
// command tree decided on is the status the process reports.
//
// The distinction is not academic here. The whole reason
// specs/features/swarm-core/swarm-core-frd.md makes the exit vocabulary a
// contract is that every caller routes on it without reading the output,
// and a main.go that dropped the return value or exited zero
// unconditionally would leave every in-process test green.
package main_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildSwarm builds the swarm binary once per invocation and returns its
// path. It builds the working tree's code rather than anything
// previously installed.
func buildSwarm(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping end-to-end binary test in -short mode (it compiles the binary)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("no go toolchain on PATH: %v", err)
	}

	// Windows will not execute a file without an executable extension,
	// however absolute the path handed to exec.Command — the failure is a
	// bare "executable file not found in %PATH%", which reads like a PATH
	// problem and is not one. go build honours whatever -o name it is
	// given, so the suffix has to be added here rather than assumed.
	name := "swarm"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the swarm binary: %v\n%s", err, out)
	}
	return bin
}

// run executes the binary and returns its stdout, its stderr and its exit
// status. A failure to start the process at all is fatal — that is a
// broken test, not a non-zero exit.
func run(t *testing.T, bin string, args ...string) (string, string, int) {
	t.Helper()

	var stdout, stderr strings.Builder
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running %s %v: %v", bin, args, err)
		}
		return stdout.String(), stderr.String(), exit.ExitCode()
	}
	return stdout.String(), stderr.String(), 0
}

// TestBinaryWithNoArgumentsPrintsTheCommandTreeAndExitsMalformed is the
// one observable behaviour the binary-skeleton delta specifies, asserted
// where it is actually specified: at the process boundary.
func TestBinaryWithNoArgumentsPrintsTheCommandTreeAndExitsMalformed(t *testing.T) {
	bin := buildSwarm(t)

	stdout, stderr, code := run(t, bin)

	if code != 3 {
		t.Errorf("process exit status = %d, want 3 (the malformed-invocation status)\n--- stderr ---\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "Usage:") || !strings.Contains(stderr, "swarm") {
		t.Errorf("the process did not print its command tree to stderr:\n%s", stderr)
	}
	if stdout != "" {
		t.Errorf("the process wrote to stdout, which is reserved for a command's answer:\n%s", stdout)
	}
}

// TestBinaryExitsZeroOnHelp is the negative control: without it a binary
// that exited 3 unconditionally would satisfy the test above.
func TestBinaryExitsZeroOnHelp(t *testing.T) {
	bin := buildSwarm(t)

	stdout, stderr, code := run(t, bin, "--help")

	if code != 0 {
		t.Errorf("`swarm --help` process exit status = %d, want 0\n--- stderr ---\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("`swarm --help` did not write help to stdout:\n%s", stdout)
	}
}
