// Package exectest provides the fake [swarmexec.Runner] the unit tier
// substitutes for the real Process Runner, per ADR-025 and the Process
// Runner component in specs/architecture/container-swarm-cli.md.
//
// It replays a recorded output, error stream, exit status and duration
// for a declared argv, so a test of anything above the Process Runner
// needs neither `bd` nor `git` installed and never starts a process.
//
// # Test-only
//
// Nothing outside a _test.go file may import this package. That is not a
// convention here: exec_test.go in the parent package parses every
// non-test source file in the module and fails if any of them imports
// this path. The reason is that a fake reachable from production code is
// a way for a command to answer without having run anything, which is the
// silent-failure class this whole container exists to remove.
package exectest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
)

// Recording is one declared argv and what running it produces.
//
// The zero value is a recording of a command that writes nothing and
// exits 0, which is the common case and needs no fields set beyond Argv.
type Recording struct {
	// Argv is the whole invocation — executable first, then each
	// argument — which is what the fake matches against. It is one slice
	// rather than a name and an args slice so that a recording reads the
	// way the command reads.
	Argv []string
	// Dir, when non-empty, additionally requires the invocation to have
	// been made in that working directory. Empty means the recording
	// matches any directory, so a test only states the directory when the
	// directory is what it is asserting.
	Dir string
	// Stdout and Stderr are replayed verbatim into the result's separate
	// streams.
	Stdout []byte
	Stderr []byte
	// ExitStatus is replayed as the finished process's status.
	ExitStatus int
	// Duration is replayed as the invocation's elapsed time. The fake
	// does not sleep for it; replaying a duration is how a test drives
	// something that reads timings without spending the time.
	Duration time.Duration
	// Err, when non-nil, is returned instead of a result — for replaying
	// an invocation that never happened, such as
	// [swarmexec.ErrExecutableNotFound].
	Err error
}

// Fake is a [swarmexec.Runner] that replays [Recording] values instead of
// starting processes, and records every invocation it was asked for.
//
// An invocation matching no recording is a test defect rather than a
// scenario: the fake returns an error naming what was asked for and what
// was declared, because a fake that quietly returned an empty success
// would make a command that invoked the wrong tool look correct.
//
// A Fake is safe for concurrent use, so a component that runs steps in
// parallel can be tested against one.
type Fake struct {
	// Recordings is what the fake knows how to answer. Two recordings
	// with the same argv are matched in order: the first is replayed for
	// the first matching invocation, the second for the next, and the
	// last one repeats. That is what lets a test replay a command whose
	// answer changes after a write.
	Recordings []Recording

	mu    sync.Mutex
	calls []swarmexec.Invocation
	used  map[int]int
}

// Compile-time assurance that the fake is substitutable for the real
// runner. Without it, a change to the interface would be discovered by
// whichever test happened to use the fake rather than here.
var _ swarmexec.Runner = (*Fake)(nil)

// NewFake returns a Fake that replays the given recordings.
func NewFake(recordings ...Recording) *Fake {
	return &Fake{Recordings: recordings}
}

// Run replays the recording declared for inv's argv.
func (f *Fake) Run(_ context.Context, inv swarmexec.Invocation) (swarmexec.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, inv)

	i, ok := f.match(inv)
	if !ok {
		return swarmexec.Result{}, fmt.Errorf(
			"exectest: no recording for %s%s; declared: %s",
			formatArgv(argvOf(inv)), formatDir(inv.Dir), formatRecordings(f.Recordings))
	}

	if f.used == nil {
		f.used = map[int]int{}
	}
	f.used[i]++

	rec := f.Recordings[i]
	if rec.Err != nil {
		return swarmexec.Result{}, rec.Err
	}
	return swarmexec.Result{
		Invocation: inv,
		Stdout:     rec.Stdout,
		Stderr:     rec.Stderr,
		ExitStatus: rec.ExitStatus,
		Duration:   rec.Duration,
	}, nil
}

// match returns the index of the recording to replay for inv. Callers
// hold f.mu.
//
// Among the recordings whose argv matches, it prefers the first one that
// has not been replayed yet, and falls back to the last matching one, so
// a single recording answers repeated invocations while a sequence of
// them is consumed in order.
func (f *Fake) match(inv swarmexec.Invocation) (int, bool) {
	argv := argvOf(inv)

	last := -1
	for i, rec := range f.Recordings {
		if !equalArgv(rec.Argv, argv) {
			continue
		}
		if rec.Dir != "" && rec.Dir != inv.Dir {
			continue
		}
		last = i
		if f.used[i] == 0 {
			return i, true
		}
	}
	if last >= 0 {
		return last, true
	}
	return 0, false
}

// Calls returns every invocation the fake was asked to run, in order.
// A test asserting which commands a component builds reads this; it is a
// copy, so a caller cannot disturb the fake's record.
func (f *Fake) Calls() []swarmexec.Invocation {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]swarmexec.Invocation, len(f.calls))
	copy(out, f.calls)
	return out
}

// Argvs returns every invocation the fake was asked to run, as argv
// vectors, in order — the form a test usually wants to compare against.
func (f *Fake) Argvs() [][]string {
	calls := f.Calls()

	out := make([][]string, len(calls))
	for i, c := range calls {
		out[i] = argvOf(c)
	}
	return out
}

// Unused returns the recordings that were never replayed. A test that
// declared a command and never provoked it has usually not exercised
// what it thinks it has.
func (f *Fake) Unused() []Recording {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []Recording
	for i, rec := range f.Recordings {
		if f.used[i] == 0 {
			out = append(out, rec)
		}
	}
	return out
}

func argvOf(inv swarmexec.Invocation) []string {
	return append([]string{inv.Name}, inv.Args...)
}

func equalArgv(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func formatArgv(argv []string) string {
	quoted := make([]string, len(argv))
	for i, a := range argv {
		quoted[i] = fmt.Sprintf("%q", a)
	}
	return "[" + strings.Join(quoted, " ") + "]"
}

func formatDir(dir string) string {
	if dir == "" {
		return ""
	}
	return fmt.Sprintf(" in %q", dir)
}

func formatRecordings(recs []Recording) string {
	if len(recs) == 0 {
		return "(none)"
	}
	out := make([]string, len(recs))
	for i, rec := range recs {
		out[i] = formatArgv(rec.Argv) + formatDir(rec.Dir)
	}
	return strings.Join(out, ", ")
}
