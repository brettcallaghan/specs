// Whitebox tests for the swarm command surface's substrate: what a bare
// invocation prints, where it prints it, and the status it exits with.
//
// The binary-skeleton delta of
// specs/features/swarm-core/swarm-core-blueprint.md specifies exactly one
// observable behaviour — "WHEN the binary is invoked with no arguments,
// it SHALL print its command tree and exit with the malformed-invocation
// status" — and these are the tests for it. The four-status mapping as a
// whole belongs to the Swarm Command Surface delta and its acceptance
// criteria; nothing here claims to verify that.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
)

// run executes the command tree in-process against args and returns
// stdout, stderr and the exit status.
func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	code := Execute(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// usageString is cobra's own rendering of the command tree, which is what
// a bare invocation is required to print. Deriving the expectation from
// the same source the code prints from is deliberate: it keeps this test
// asserting *that the tree is printed* as subcommands are added, rather
// than pinning today's empty tree as the answer forever.
func usageString(t *testing.T) string {
	t.Helper()

	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	root := newRootCmd(streams)
	// The same tree the binary builds, subcommands included: what a bare
	// invocation prints is the tree as it stands, not an empty one.
	register(root, streams, exectest.NewFake())
	// cobra registers --help during its own Execute, not at construction,
	// so a root that has never been executed renders a usage block with no
	// Flags section. Registering it here is what makes the two renderings
	// comparable.
	// Same for the `help` command, which cobra adds during Execute once
	// the tree has subcommands.
	root.InitDefaultHelpFlag()
	root.InitDefaultHelpCmd()
	return root.UsageString()
}

func TestBareInvocationPrintsTheCommandTreeAndExitsMalformed(t *testing.T) {
	stdout, stderr, code := run(t)

	if code != int(StatusMalformed) {
		t.Errorf("exit status = %d, want %d (the malformed-invocation status)", code, int(StatusMalformed))
	}
	if want := usageString(t); stderr != want {
		t.Errorf("a bare invocation did not print the command tree.\n--- got (stderr) ---\n%s\n--- want ---\n%s", stderr, want)
	}
	if stdout != "" {
		t.Errorf("a bare invocation wrote to stdout, which is reserved for a command's answer:\n%s", stdout)
	}
	if !strings.Contains(stderr, "swarm") {
		t.Errorf("the printed command tree does not name the binary:\n%s", stderr)
	}
}

func TestAnUnknownFlagExitsMalformed(t *testing.T) {
	stdout, stderr, code := run(t, "--no-such-flag")

	if code != int(StatusMalformed) {
		t.Errorf("exit status = %d, want %d for an unknown flag", code, int(StatusMalformed))
	}
	if stdout != "" {
		t.Errorf("an argument error wrote to stdout:\n%s", stdout)
	}
	if !strings.Contains(stderr, "no-such-flag") {
		t.Errorf("the failure does not name the flag it rejected:\n%s", stderr)
	}
}

// TestHelpExitsClear is the negative control for the two tests above.
// Without it a surface that exited 3 unconditionally would satisfy both
// of them and nothing would notice.
func TestHelpExitsClear(t *testing.T) {
	stdout, _, code := run(t, "--help")

	if code != int(StatusClear) {
		t.Errorf("`swarm --help` exit status = %d, want %d", code, int(StatusClear))
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("`swarm --help` did not write help to stdout, where a requested answer belongs:\n%s", stdout)
	}
}
