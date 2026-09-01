// Tests for the Process Runner. Four of them are the four properties
// ADR-025 selected this design for, and each one is written against a
// real child process, because a fake cannot demonstrate that a real
// process's status was read from the process.
//
// The child is this test binary re-invoked with a mode argument — the
// standard library's own idiom for testing os/exec. It needs no Go
// toolchain at run time and no shell, so the tests run unchanged on the
// windows-latest leg of .github/workflows/cli.yml, where a shell script
// or a /bin/sh helper would not.
package exec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
)

const (
	// bigOutputBytes is deliberately larger than a pipe buffer (64 KiB on
	// Linux), so a design that streamed through a pipe would have to
	// block or drop; collecting it whole is what makes the reader-stops-
	// early inversion impossible.
	bigOutputBytes = 2 << 20 // 2 MiB
	// helperSleep is what the "slow" mode sleeps for. Assertions against
	// it use a much lower bound, so a slow or coarse-grained timer does
	// not make the test flaky.
	helperSleep = 150 * time.Millisecond
)

// helperInvocation builds an Invocation that re-runs this test binary as
// a child in the given mode, passing extra through to it.
func helperInvocation(mode string, extra ...string) swarmexec.Invocation {
	args := []string{"-test.run=^TestHelperProcess$", "--", mode}
	return swarmexec.Invocation{Name: os.Args[0], Args: append(args, extra...)}
}

// helperArgs returns the arguments after the "--" separator, or false
// when this process was not invoked as a helper child.
func helperArgs() ([]string, bool) {
	for i, a := range os.Args {
		if a == "--" {
			return os.Args[i+1:], true
		}
	}
	return nil, false
}

// TestHelperProcess is not a test. It is the child process the tests in
// this file run: invoked with a "--" separator and a mode, it writes what
// the mode calls for and exits. It calls os.Exit before the testing
// package can print anything, which is what keeps its stdout exactly the
// bytes the mode wrote.
//
// Invoked normally — by `go test` with no separator — it returns
// immediately and passes.
func TestHelperProcess(t *testing.T) {
	args, ok := helperArgs()
	if !ok || len(args) == 0 {
		return
	}

	switch mode := args[0]; mode {
	case "streams":
		// The shape the bead tracker actually produces: a warning on the
		// error stream while the payload goes to the output stream.
		fmt.Fprint(os.Stdout, `{"ok":true}`)
		fmt.Fprint(os.Stderr, "warning: beads.role not configured\n")
		os.Exit(0)

	case "output-then-fail":
		fmt.Fprint(os.Stdout, "line one\nline two\nline three\n")
		fmt.Fprint(os.Stderr, "it did not work\n")
		os.Exit(7)

	case "big-output":
		if _, err := os.Stdout.Write(bytes.Repeat([]byte("x"), bigOutputBytes)); err != nil {
			os.Exit(98)
		}
		os.Exit(0)

	case "echo-args":
		for _, a := range args[1:] {
			fmt.Fprintf(os.Stdout, "%s\n", a)
		}
		os.Exit(0)

	case "slow":
		time.Sleep(helperSleep)
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "unknown helper mode %q\n", mode)
		os.Exit(99)
	}
}

// AC: AC-SWARM-CORE-001.10
//
// The shell form this replaces is `cmd | tail`, whose status is tail's:
// a failing command reads as a pass. Here the child writes output *and*
// exits 7, and the status the runner reports comes from the finished
// process rather than from anything downstream of it.
func TestAFailingCommandIsReportedAsFailedEvenThoughItProducedOutput(t *testing.T) {
	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("output-then-fail"))
	if err != nil {
		t.Fatalf("Run returned an error for a process that ran and exited non-zero: %v", err)
	}

	if res.ExitStatus != 7 {
		t.Errorf("ExitStatus = %d, want 7 (the status the child exited with)", res.ExitStatus)
	}
	if res.Succeeded() {
		t.Error("Succeeded() = true for a child that exited 7")
	}
	if !strings.Contains(string(res.Stdout), "line three") {
		t.Errorf("the child's output was not collected; got %q", res.Stdout)
	}
}

// AC: AC-SWARM-CORE-001.11
//
// The inversion this replaces is `set -o pipefail` with a reader that
// stops early: the writer dies on SIGPIPE and a successful command is
// reported as failed. There is no incremental reader here — each stream
// is collected whole — so a child that writes far more than a pipe buffer
// holds and exits 0 is reported as having succeeded, with every byte
// present.
func TestASucceedingCommandWithMoreOutputThanAPipeHoldsIsNotReportedAsFailed(t *testing.T) {
	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("big-output"))
	if err != nil {
		t.Fatalf("Run returned an error for a child that exited 0: %v", err)
	}

	if res.ExitStatus != 0 {
		t.Errorf("ExitStatus = %d, want 0", res.ExitStatus)
	}
	if !res.Succeeded() {
		t.Error("Succeeded() = false for a child that exited 0")
	}
	if len(res.Stdout) != bigOutputBytes {
		t.Errorf("collected %d bytes of stdout, want %d — the output was truncated", len(res.Stdout), bigOutputBytes)
	}
}

// AC: AC-SWARM-CORE-001.12
//
// The tracker writes warnings to its error stream while its data goes to
// its output stream, so a runner that merged them would hand a parser a
// buffer with a warning in the middle of the payload.
func TestTheOutputAndErrorStreamsAreNeverMerged(t *testing.T) {
	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("streams"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var payload struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(res.Stdout, &payload); err != nil {
		t.Fatalf("the output stream did not parse as the payload the child wrote: %v\nstdout: %q", err, res.Stdout)
	}
	if !payload.OK {
		t.Errorf("parsed payload = %+v, want ok=true", payload)
	}
	if bytes.Contains(res.Stdout, []byte("warning")) {
		t.Errorf("the warning written to the error stream is present in what was parsed: %q", res.Stdout)
	}
	if !bytes.Contains(res.Stderr, []byte("warning")) {
		t.Errorf("the error stream was not captured at all; got %q", res.Stderr)
	}
}

// An argument that looks like shell syntax is one argument. This is the
// property that makes quoting bugs structurally impossible rather than
// something every call site has to get right, and it is why nothing here
// builds a command string.
func TestAnArgumentThatLooksLikeShellSyntaxIsPassedThroughLiterally(t *testing.T) {
	const hostile = `a; rm -rf /; $(echo hi) | tee x`

	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("echo-args", hostile, "second"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := strings.Split(strings.TrimSuffix(string(res.Stdout), "\n"), "\n")
	want := []string{hostile, "second"}
	if len(got) != len(want) {
		t.Fatalf("the child saw %d arguments, want %d: %q", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("argument %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// Every invocation's elapsed time is recorded, because the run's progress
// measurements are ages rather than counts.
func TestEachInvocationRecordsItsElapsedTime(t *testing.T) {
	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("slow"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// A third of what the child sleeps for: enough to prove the duration
	// came from the run rather than being left at zero, low enough that a
	// coarse timer cannot fail it.
	if floor := helperSleep / 3; res.Duration < floor {
		t.Errorf("Duration = %v, want at least %v for a child sleeping %v", res.Duration, floor, helperSleep)
	}
}

// A missing executable and a tool that answered no are different
// conditions, and the exit vocabulary needs them separable without
// parsing a message. No acceptance criterion covers this — it is the
// Process Runner delta's own "IF the named executable cannot be found,
// THEN the Process Runner SHALL report that distinctly from a non-zero
// exit" — so the citation is to the delta rather than to an AC.
func TestAMissingExecutableIsReportedDistinctlyFromANonZeroExit(t *testing.T) {
	missing := swarmexec.Invocation{Name: "swarm-no-such-executable-6f1c2b"}

	res, err := swarmexec.ProcessRunner{}.Run(context.Background(), missing)
	if err == nil {
		t.Fatalf("Run returned no error for an executable that does not exist; result: %+v", res)
	}
	if !errors.Is(err, swarmexec.ErrExecutableNotFound) {
		t.Errorf("errors.Is(err, ErrExecutableNotFound) = false for %v", err)
	}
	if !strings.Contains(err.Error(), missing.Name) {
		t.Errorf("the error does not name the executable it looked for: %v", err)
	}

	// It also unwraps to what the standard library reported, which is the
	// measurement ADR-025 rests on.
	var lookErr *osexec.Error
	if !errors.As(err, &lookErr) {
		t.Errorf("the error does not unwrap to *os/exec.Error: %v", err)
	}

	// The negative control. Without it, a runner that returned
	// ErrExecutableNotFound for everything would satisfy the assertions
	// above.
	_, err = swarmexec.ProcessRunner{}.Run(context.Background(), helperInvocation("output-then-fail"))
	if err != nil {
		t.Fatalf("a child that ran and exited 7 returned an error: %v", err)
	}
}

// A working directory that does not exist stops the process from running
// without the executable being missing, and must not be mistaken for one.
func TestAnUnusableWorkingDirectoryIsNotReportedAsAMissingExecutable(t *testing.T) {
	inv := helperInvocation("streams")
	inv.Dir = filepath.Join(t.TempDir(), "no-such-directory")

	_, err := swarmexec.ProcessRunner{}.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("Run returned no error for a working directory that does not exist")
	}
	if errors.Is(err, swarmexec.ErrExecutableNotFound) {
		t.Errorf("a bad working directory was reported as a missing executable: %v", err)
	}
}

// TestTheFakeIsNotReachableFromProductionCode enforces the "test-only
// package" the Process Runner delta specifies. Go has no mechanism for
// it, so this is the mechanism: no non-test source file in the module may
// import the fake, because a fake reachable from production code is a way
// for a command to report an answer it never obtained.
func TestTheFakeIsNotReachableFromProductionCode(t *testing.T) {
	const fakeImport = `"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"`

	// The module root, from this package's directory.
	root := filepath.Join("..", "..")

	fset := token.NewFileSet()
	var scanned, testImporters int

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// testdata holds committed corpora, not module source.
			if d.Name() == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		isTest := strings.HasSuffix(path, "_test.go")
		if !isTest {
			scanned++
		}
		for _, imp := range file.Imports {
			if imp.Path.Value != fakeImport {
				continue
			}
			if isTest {
				testImporters++
				continue
			}
			t.Errorf("%s imports the test-only fake package; the fake must not be reachable from production code", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the module: %v", err)
	}

	// Two guards against a vacuous pass: the walk has to have found
	// source at all, and the import path string has to match something
	// real, which is what proves a rename would be caught rather than
	// silently turning this test into a no-op.
	if scanned == 0 {
		t.Fatalf("scanned no non-test source files under %s; the walk is not reaching the module", root)
	}
	if testImporters == 0 {
		t.Errorf("no _test.go file imports %s; either the fake is unused or this test's import path is stale", fakeImport)
	}
}
