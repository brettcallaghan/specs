package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// The five exit statuses [[cli-exit-status-standards]] defines, and no
// others. Each names who has to act on a run that ends with it.
//
// Their numeric order is also their precedence: when a run meets the
// conditions of more than one class, the highest wins (4 > 3 > 2 > 1),
// so combining two statuses is max, never a lookup.
const (
	// statusOK: the command did what was asked. A query answered; a
	// gate passed.
	statusOK = 0
	// statusGateFailed: a gate ran to completion and the corpus failed
	// it. The corpus author acts. A gate signals it by setting
	// ioStreams.exitCode and returning nil — it is an answer, not an
	// error, and its findings stay on standard output.
	statusGateFailed = 1
	// statusInvocation: the command could not run as invoked — an
	// unknown command or flag, a missing or surplus argument, a value
	// outside a registered set, an operand the command cannot act on.
	// The caller acts.
	statusInvocation = 2
	// statusEnvironment: the invocation was correct but its
	// surroundings prevent it — no specs.toml, an invalid
	// configuration, an unreadable spec root, an output path that cannot
	// be written. The repository's maintainer acts.
	statusEnvironment = 3
	// statusInternal: this tool reached a state it does not handle — a
	// panic, an encoder refusing a value, a broken invariant. This
	// tool's maintainer acts.
	statusInternal = 4
)

// classified is implemented by every error that knows its own exit
// status: [*classifiedError], and sentinel types such as
// [*ErrConfigNotFound] whose class never varies with the call site.
type classified interface {
	error
	exitStatus() int
}

// classifiedError carries an exit-status class. Error is the message
// written to standard error, so it begins with the offending value or
// condition ([[cli-exit-status-standards]]'s "The reason names the
// class"); Unwrap exposes the cause, so [errors.Is] on a sentinel keeps
// matching through the classification.
type classifiedError struct {
	status int
	err    error
}

func (e *classifiedError) Error() string   { return e.err.Error() }
func (e *classifiedError) Unwrap() error   { return e.err }
func (e *classifiedError) exitStatus() int { return e.status }

// invocationError reports a run the caller invoked wrongly (status 2).
// It has [fmt.Errorf] semantics, %w included.
func invocationError(format string, args ...any) error {
	return &classifiedError{status: statusInvocation, err: fmt.Errorf(format, args...)}
}

// environmentError reports a run its surroundings prevented (status 3).
func environmentError(format string, args ...any) error {
	return &classifiedError{status: statusEnvironment, err: fmt.Errorf(format, args...)}
}

// internalError reports a state this tool does not handle (status 4).
func internalError(format string, args ...any) error {
	return &classifiedError{status: statusInternal, err: fmt.Errorf(format, args...)}
}

// prefixedError presents err with a leading package prefix removed, so a
// sentinel built inside cli/core ("corpus: ...") can be reported in
// words that begin with the offending value, while [errors.Is] still
// reaches the sentinel through Unwrap.
type prefixedError struct {
	msg string
	err error
}

func (e *prefixedError) Error() string { return e.msg }
func (e *prefixedError) Unwrap() error { return e.err }

// withoutPrefix returns err presented without prefix, or err itself when
// its message does not begin with it.
func withoutPrefix(err error, prefix string) error {
	msg := err.Error()
	trimmed, found := strings.CutPrefix(msg, prefix)
	if !found {
		return err
	}
	return &prefixedError{msg: trimmed, err: err}
}

// statusOf maps the error a command returned to its exit status:
//
//   - nil is statusOK;
//   - an error that classifies itself, anywhere in the chain, has its own
//     status — and an [errors.Join] of several has the highest among them;
//   - otherwise an [*fs.PathError] anywhere in the chain is
//     statusEnvironment: a file the repository should hold could not be
//     read or written;
//   - anything else is statusInternal. An error nobody classified is a
//     state this tool does not handle, and defaulting it to a plausible
//     class would hide the missing classification instead of surfacing it
//     as the bug it is.
func statusOf(err error) int {
	if err == nil {
		return statusOK
	}
	if status, ok := classifiedStatus(err); ok {
		return status
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return statusEnvironment
	}
	return statusInternal
}

// classifiedStatus walks err's chain — both single and multiple Unwrap —
// and returns the highest status any self-classifying error in it
// carries. The first classified error on a single-Unwrap path decides
// that path: an outer classification overrides whatever it wraps.
func classifiedStatus(err error) (int, bool) {
	if c, ok := err.(classified); ok {
		return c.exitStatus(), true
	}
	switch u := err.(type) {
	case interface{ Unwrap() []error }:
		best, found := 0, false
		for _, inner := range u.Unwrap() {
			if inner == nil {
				continue
			}
			if status, ok := classifiedStatus(inner); ok {
				found = true
				best = max(best, status)
			}
		}
		return best, found
	case interface{ Unwrap() error }:
		if inner := u.Unwrap(); inner != nil {
			return classifiedStatus(inner)
		}
	}
	return 0, false
}

// Execute parses args against the command tree and runs the selected
// subcommand, writing to stdout and stderr and returning the process exit
// status, per [[cli-exit-status-standards]].
//
// It is the one function main calls.
func Execute(args []string, stdout, stderr io.Writer) int {
	return execute(newRootCmd, args, stdout, stderr)
}

// execute is [Execute] over any command tree, so the stream and status
// rules below can be tested against a tree built for the purpose.
//
//   - Standard output is buffered and copied out only when the run ends
//     0 or 1. A run ending 2, 3 or 4 answers nothing, including whatever
//     half an answer a command wrote before it failed ("A failed
//     invocation answers nothing"). Status 1's findings are the answer,
//     so they are written.
//   - An error is written to stderr as its own message plus a newline:
//     cobra's "Error: " prefix is silenced, because the message begins
//     with the offending value or condition.
//   - An error raised before any RunE started — an unknown subcommand, a
//     flag that does not parse, an Args validator, a required flag — is
//     statusInvocation. An error a RunE returned is [statusOf] it.
//   - A panic is recovered and exits statusInternal, its value and stack
//     on stderr. An unrecovered Go panic exits 2, which would read as
//     "you called me wrong". The CLI starts no goroutines, so recovering
//     here covers every panic a command can raise.
//   - The final status is the highest of the error's class and the
//     exitCode a gate set, because precedence is numeric order.
func execute(build func(*ioStreams) *cobra.Command, args []string, stdout, stderr io.Writer) (status int) {
	var answer bytes.Buffer
	streams := &ioStreams{out: &answer, err: stderr}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(stderr, "internal error: %v\n%s", r, debug.Stack())
			status = statusInternal
		}
	}()

	root := build(streams)
	root.SetArgs(args)
	root.SetOut(&answer)
	root.SetErr(stderr)
	root.SilenceErrors = true
	root.SilenceUsage = true
	markEntry(root, streams)

	err := root.Execute()
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
	}

	status = statusOK
	if err != nil {
		status = statusInvocation
		if streams.entered {
			status = statusOf(err)
		}
	}
	status = max(status, streams.exitCode)

	if status <= statusGateFailed {
		if _, werr := stdout.Write(answer.Bytes()); werr != nil {
			fmt.Fprintf(stderr, "standard output cannot be written: %v\n", werr)
			return statusEnvironment
		}
	}
	return status
}

// markEntry wraps the RunE of cmd and every command beneath it so that
// streams.entered is set the moment a RunE starts. An error returned
// while it is still unset came from cobra's own parsing and validation,
// which run before RunE (flag parsing, Args, required flags and flag
// groups), and is the caller's.
//
// Wrapping RunE rather than installing a PersistentPreRun on the root is
// deliberate: cobra validates required flags and flag groups after the
// persistent pre-run, so a pre-run marker would misclassify those as a
// RunE's error, and a subcommand defining its own PersistentPreRun would
// shadow the root's.
func markEntry(cmd *cobra.Command, streams *ioStreams) {
	if run := cmd.RunE; run != nil {
		cmd.RunE = func(c *cobra.Command, args []string) error {
			streams.entered = true
			return run(c, args)
		}
	}
	for _, child := range cmd.Commands() {
		markEntry(child, streams)
	}
}
