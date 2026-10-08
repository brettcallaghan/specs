// Whitebox tests for the exit-status mechanism in exit.go: how statusOf
// classifies an error, and how execute applies the stream and status
// rules of specs/standards/cli-exit-status-standards.md to any command
// tree — a panic exits 4, a run exiting 2, 3 or 4 writes nothing to
// standard output, and a gate's findings survive a status of 1.
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var errSentinel = errors.New("sentinel")

func TestStatusOf(t *testing.T) {
	pathErr := &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist}

	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, statusOK},
		{"invocation", invocationError("bad flag"), statusInvocation},
		{"environment", environmentError("no config"), statusEnvironment},
		{"internal", internalError("broken"), statusInternal},
		{"classified then wrapped", fmt.Errorf("outer: %w", invocationError("inner")), statusInvocation},
		{"join takes the highest", errors.Join(invocationError("a"), environmentError("b")), statusEnvironment},
		{"join, highest first", errors.Join(internalError("a"), invocationError("b")), statusInternal},
		{"bare path error", pathErr, statusEnvironment},
		{"wrapped path error", fmt.Errorf("reading: %w", pathErr), statusEnvironment},
		{"config not found", &ErrConfigNotFound{Path: "specs.toml"}, statusEnvironment},
		{"unclassified", errors.New("nobody classified me"), statusInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusOf(tc.err); got != tc.want {
				t.Errorf("statusOf(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestClassifiedErrorKeepsTheSentinelReachable(t *testing.T) {
	err := invocationError("%w", errSentinel)
	if !errors.Is(err, errSentinel) {
		t.Errorf("errors.Is(invocationError(%%w, sentinel), sentinel) = false, want true")
	}
	if err.Error() != errSentinel.Error() {
		t.Errorf("Error() = %q, want the wrapped message %q", err.Error(), errSentinel.Error())
	}
}

func TestWithoutPrefixKeepsTheSentinelReachable(t *testing.T) {
	err := withoutPrefix(fmt.Errorf("corpus: %q is %w", "x", errSentinel), "corpus: ")
	if got, want := err.Error(), `"x" is sentinel`; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, errSentinel) {
		t.Error("errors.Is through withoutPrefix = false, want true")
	}
}

// treeRunning builds a one-subcommand tree whose RunE writes a partial
// answer and then does whatever run says.
func treeRunning(run func(*ioStreams) error) func(*ioStreams) *cobra.Command {
	return func(streams *ioStreams) *cobra.Command {
		root := &cobra.Command{Use: "specs"}
		root.AddCommand(&cobra.Command{
			Use:  "go",
			Args: cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				fmt.Fprintln(streams.out, "half an answer")
				return run(streams)
			},
		})
		return root
	}
}

func runTree(build func(*ioStreams) *cobra.Command, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	status := execute(build, args, &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func TestExecuteRecoversAPanicAsInternal(t *testing.T) {
	status, stdout, stderr := runTree(treeRunning(func(*ioStreams) error {
		panic("the invariant broke")
	}), "go")

	if status != statusInternal {
		t.Errorf("status = %d, want %d", status, statusInternal)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty: a panicking run answers nothing", stdout)
	}
	if !strings.Contains(stderr, "the invariant broke") {
		t.Errorf("stderr = %q, want it to carry the panic value", stderr)
	}
}

func TestExecuteDiscardsStdoutOnAFailedRun(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"invocation", invocationError("bad operand"), statusInvocation},
		{"environment", environmentError("no config"), statusEnvironment},
		{"internal", internalError("encoder refused"), statusInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, stdout, stderr := runTree(treeRunning(func(*ioStreams) error { return tc.err }), "go")
			if status != tc.want {
				t.Errorf("status = %d, want %d", status, tc.want)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if want := tc.err.Error() + "\n"; stderr != want {
				t.Errorf("stderr = %q, want exactly %q (no \"Error: \" prefix)", stderr, want)
			}
		})
	}
}

func TestExecuteKeepsAFailedGatesAnswer(t *testing.T) {
	status, stdout, _ := runTree(treeRunning(func(streams *ioStreams) error {
		streams.exitCode = statusGateFailed
		return nil
	}), "go")

	if status != statusGateFailed {
		t.Errorf("status = %d, want %d", status, statusGateFailed)
	}
	if stdout != "half an answer\n" {
		t.Errorf("stdout = %q, want the gate's findings intact", stdout)
	}
}

func TestExecuteTakesTheHigherOfGateAndError(t *testing.T) {
	status, stdout, _ := runTree(treeRunning(func(streams *ioStreams) error {
		streams.exitCode = statusGateFailed
		return environmentError("the spec root went away")
	}), "go")

	if status != statusEnvironment {
		t.Errorf("status = %d, want %d: 3 outranks 1", status, statusEnvironment)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func TestExecuteClassifiesCobrasOwnErrorsAsInvocation(t *testing.T) {
	build := treeRunning(func(*ioStreams) error { return nil })
	for _, args := range [][]string{{"nope"}, {"go", "--bogus"}, {"go", "surplus"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			status, stdout, stderr := runTree(build, args...)
			if status != statusInvocation {
				t.Errorf("status = %d, want %d", status, statusInvocation)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if strings.HasPrefix(stderr, "Error: ") {
				t.Errorf("stderr = %q, want no \"Error: \" prefix", stderr)
			}
		})
	}
}

func TestExecuteUnclassifiedRunEErrorIsInternal(t *testing.T) {
	status, _, _ := runTree(treeRunning(func(*ioStreams) error {
		return errors.New("nobody classified me")
	}), "go")
	if status != statusInternal {
		t.Errorf("status = %d, want %d", status, statusInternal)
	}
}
