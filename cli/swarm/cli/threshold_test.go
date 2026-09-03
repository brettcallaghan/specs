// Unit tests for `swarm ready`'s two integer flags. They are the only
// integer flags in the whole command surface — measured:
// `grep -rn 'IntVar\|IntSliceVar\|Int64Var' cli/swarm/cli/` returns
// exactly the two lines that bind them — and each has a value the
// breaker cannot read as a report.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
)

// runReadyArgs drives `swarm ready` through the real command tree against
// a fake with NO recordings, so anything the command reads or writes is a
// test failure rather than a scenario. It returns stdout, stderr, the
// exit status and the fake, and is the shape every case below asserts
// from: a refusal that had already read something would pass a
// status-only check.
func runReadyArgs(t *testing.T, args ...string) (string, string, int, *exectest.Fake) {
	t.Helper()

	fake := exectest.NewFake()
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	code := dispatch(root, streams, args, &out, &errOut)
	return out.String(), errOut.String(), code, fake
}

// A non-positive --threshold makes the breaker's `next >= threshold`
// comparison unconditionally true against a count that is never negative,
// so every run stops — including one whose dispatch just completed work
// and whose count therefore reset to zero. The value is refused at the
// surface instead, before the breaker reads or writes anything.
//
// Every mode, explain included: a malformed argument is not made
// well-formed by asking the command to only describe itself.
func TestReadyRefusesANonPositiveThreshold(t *testing.T) {
	for _, value := range []string{"0", "-1", "-3"} {
		for _, mode := range [][]string{nil, {"--dry-run"}, {"--explain"}} {
			args := append([]string{"ready", "--run", "run-1", "--threshold", value}, mode...)

			stdout, stderr, code, fake := runReadyArgs(t, args...)
			if code != int(StatusMalformed) {
				t.Errorf("%v exited %d, want %d (malformed)\n%s%s", args, code, StatusMalformed, stdout, stderr)
			}
			if calls := fake.Argvs(); len(calls) != 0 {
				t.Errorf("%v read or wrote something before refusing: %v", args, calls)
			}
			said := stdout + stderr
			if !strings.Contains(said, "--threshold") {
				t.Errorf("%v does not name the flag it refused:\n%s", args, said)
			}
			if !strings.Contains(said, "1 or greater") {
				t.Errorf("%v does not say what a valid value is:\n%s", args, said)
			}
		}
	}
}

// A negative --completed is read by `report.Completed > 0` as false, so
// it is silently treated as --completed 0 and ADVANCES the failure count:
// an operator reporting a malformed value gets a breaker advance rather
// than a refusal. It is refused instead.
func TestReadyRefusesANegativeCompleted(t *testing.T) {
	for _, value := range []string{"-1", "-7"} {
		for _, mode := range [][]string{nil, {"--dry-run"}, {"--explain"}} {
			args := append([]string{"ready", "--run", "run-1", "--completed", value}, mode...)

			stdout, stderr, code, fake := runReadyArgs(t, args...)
			if code != int(StatusMalformed) {
				t.Errorf("%v exited %d, want %d (malformed)\n%s%s", args, code, StatusMalformed, stdout, stderr)
			}
			if calls := fake.Argvs(); len(calls) != 0 {
				t.Errorf("%v read or wrote something before refusing: %v", args, calls)
			}
			said := stdout + stderr
			if !strings.Contains(said, "--completed") {
				t.Errorf("%v does not name the flag it refused:\n%s", args, said)
			}
			if !strings.Contains(said, "0 or greater") {
				t.Errorf("%v does not say what a valid value is:\n%s", args, said)
			}
		}
	}
}

// The guards bound the malformed values and nothing else. An omitted
// --completed keeps its current meaning — the invocation only reads the
// breaker's count — and a positive --threshold is untouched, which is
// what the existing breaker tests in ready_test.go assert in full.
func TestReadyAcceptsTheValidEdgeOfEachFlag(t *testing.T) {
	for _, args := range [][]string{
		{"ready", "--run", "run-1", "--threshold", "1", "--explain"},
		{"ready", "--run", "run-1", "--completed", "0", "--explain"},
		{"ready", "--run", "run-1", "--explain"},
	} {
		stdout, stderr, code, _ := runReadyArgs(t, args...)
		if code == int(StatusMalformed) {
			t.Errorf("%v was refused as malformed:\n%s%s", args, stdout, stderr)
		}
	}
}
