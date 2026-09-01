// Package exec is the Swarm CLI's Process Runner: the one component that
// starts an external process, per the component of that name in
// specs/architecture/container-swarm-cli.md and ADR-025.
//
// Four properties are the whole reason it exists, and each guards a
// failure this repository has observed in the shell pipelines this CLI
// replaces:
//
//   - An invocation is an argv vector. Nothing here builds a command
//     string for a shell interpreter, so an argument that looks like
//     shell syntax is one argument. Windows is a supported target and has
//     no `sh` (ADR-023), so a shell was never available anyway.
//   - Standard output and standard error go to separate buffers. The bead
//     tracker writes warnings to its error stream while its data goes to
//     its output stream, so anything merging the two corrupts the payload
//     it is about to parse.
//   - The exit status is read from the finished process. Reading it
//     through a pipe returns the reader's status, so a failing step reads
//     as a pass; there is no pipe here to read it through.
//   - A reader that stops consuming early cannot invert a success into a
//     failure, because there is no incremental reader: each stream is
//     collected whole into a buffer.
//
// It knows nothing about what it is running. The argv and the meaning of
// the bytes that come back both arrive from the caller, which is what
// confines the tracker's and git's command surfaces to one component
// each (container-swarm-cli.md, Change Guidance: "A component other than
// those two implementations SHALL NOT reach the Process Runner").
package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	osexec "os/exec"
	"time"
)

// Invocation is one external command: an executable and the arguments it
// is given, as a vector.
//
// There is deliberately no field holding a command line. A command line
// would have to be split by something, and whatever split it is the
// quoting bug this type exists to make impossible.
type Invocation struct {
	// Name is the executable to run. A bare name is resolved against
	// PATH; a name containing a path separator is used as given.
	Name string
	// Args are the arguments after the executable, one element per
	// argument. An element is passed through verbatim: no splitting on
	// whitespace, no glob expansion, no variable substitution.
	Args []string
	// Dir is the working directory to run in. Empty means the calling
	// process's own, which is what os/exec does. It is here because this
	// CLI reads git from a specific worktree, and the alternative — the
	// caller changing its own process's directory — is shared mutable
	// state across concurrent invocations.
	Dir string
}

// Result is what one finished invocation produced.
//
// A non-zero ExitStatus is data, not an error: a tool answering "no" is a
// successful invocation of it, and the caller decides what the answer
// means. The error return is reserved for an invocation that never
// happened.
type Result struct {
	// Invocation is what was run, carried back so a caller reporting a
	// fact can name the command that established it without holding onto
	// the request separately.
	Invocation Invocation
	// Stdout is the whole of the process's standard output.
	Stdout []byte
	// Stderr is the whole of the process's standard error, never merged
	// into Stdout.
	Stderr []byte
	// ExitStatus is the status the finished process reported. It is -1
	// when the process was terminated by a signal, which is what
	// os.ProcessState reports and is not flattened here into a number a
	// process could also have exited with.
	ExitStatus int
	// Duration is how long the invocation took, measured around the run.
	// The run's progress measurements are ages, so this is required
	// rather than diagnostic.
	Duration time.Duration
}

// Succeeded reports whether the process exited zero.
func (r Result) Succeeded() bool { return r.ExitStatus == 0 }

// ErrExecutableNotFound reports that the named executable could not be
// found at all, which is a different condition from a tool that ran and
// answered no. Callers separate the two with errors.Is; the exit
// vocabulary in specs/features/swarm-core/swarm-core-frd.md needs that
// separation, because a missing tool is a malformed environment and a
// non-zero exit is a finding.
var ErrExecutableNotFound = errors.New("executable not found")

// NotFoundError is the error a Runner returns when the executable could
// not be resolved. It matches [ErrExecutableNotFound] under errors.Is,
// and unwraps to whatever the toolchain reported, so a caller that wants
// the underlying *os/exec.Error can still reach it.
type NotFoundError struct {
	// Name is the executable that could not be found.
	Name string
	// Err is the error the standard library returned.
	Err error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("swarm/exec: executable not found: %s", e.Name)
}

// Is makes errors.Is(err, ErrExecutableNotFound) true for this error.
func (e *NotFoundError) Is(target error) bool { return target == ErrExecutableNotFound }

// Unwrap exposes the standard library's own error.
func (e *NotFoundError) Unwrap() error { return e.Err }

// Runner starts one external process and returns what it produced.
//
// This is the seam the unit tier replaces. A test that would otherwise
// need `bd` or `git` installed substitutes a fake implementing this
// interface, which is why every component above declares its dependency
// as a Runner rather than as the concrete [ProcessRunner].
type Runner interface {
	// Run invokes inv and returns its result. A non-zero exit is
	// reported through [Result.ExitStatus] with a nil error; a non-nil
	// error means the invocation did not produce a status at all.
	Run(ctx context.Context, inv Invocation) (Result, error)
}

// ProcessRunner is the [Runner] that starts real processes, over
// os/exec. It holds no state, so the zero value is usable and one value
// may be shared.
type ProcessRunner struct{}

// Run invokes inv as an argv vector and returns what the finished
// process produced.
//
// The three error outcomes are deliberately distinct:
//
//   - The executable could not be resolved: [NotFoundError], matching
//     [ErrExecutableNotFound]. ADR-025 records the measurement this rests
//     on — a missing executable comes back as *os/exec.Error while a
//     non-zero exit comes back as *os/exec.ExitError, so the two are
//     separable without parsing a message.
//   - The process ran and exited non-zero: no error, and the status is in
//     the result. This is the case a shell pipeline loses.
//   - Anything else that stopped the process from running — a working
//     directory that does not exist, a cancelled context: a wrapped
//     error, with whatever was captured before the failure.
func (ProcessRunner) Run(ctx context.Context, inv Invocation) (Result, error) {
	var stdout, stderr bytes.Buffer

	// CommandContext, not Command: a cancelled context has to be able to
	// stop a tool that is waiting on something, and adding it later would
	// change every call site.
	cmd := osexec.CommandContext(ctx, inv.Name, inv.Args...)
	cmd.Dir = inv.Dir
	// Two buffers, never one. bytes.Buffer collects each stream whole, so
	// nothing here is a reader that can stop early and make a successful
	// writer die on SIGPIPE.
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	result := Result{
		Invocation: inv,
		Stdout:     stdout.Bytes(),
		Stderr:     stderr.Bytes(),
		Duration:   elapsed,
	}

	if err != nil {
		var lookErr *osexec.Error
		if errors.As(err, &lookErr) {
			// Nothing ran, so there is no result to report — returning a
			// zero-status Result beside this error would read as success
			// to anyone who checked the status first.
			return Result{}, &NotFoundError{Name: inv.Name, Err: err}
		}

		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitStatus = exitErr.ExitCode()
			return result, nil
		}

		return result, fmt.Errorf("swarm/exec: running %q: %w", inv.Name, err)
	}

	result.ExitStatus = cmd.ProcessState.ExitCode()
	return result, nil
}

// Compile-time assurance that the concrete runner satisfies the seam
// every component above depends on.
var _ Runner = ProcessRunner{}
