// Tests for the fake. It is a test double, but it is also the seam every
// unit test above the Process Runner rests on: a fake that matched
// loosely, or answered an invocation nobody declared, would make a
// command that ran the wrong tool look correct.
package exectest_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
)

func TestItReplaysTheRecordingDeclaredForAnArgv(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:       []string{"bd", "show", "x-1", "--json"},
		Stdout:     []byte(`[{"id":"x-1"}]`),
		Stderr:     []byte("warning: beads.role not configured\n"),
		ExitStatus: 2,
		Duration:   40 * time.Millisecond,
	})

	res, err := fake.Run(context.Background(), swarmexec.Invocation{
		Name: "bd", Args: []string{"show", "x-1", "--json"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := string(res.Stdout); got != `[{"id":"x-1"}]` {
		t.Errorf("Stdout = %q", got)
	}
	if got := string(res.Stderr); !strings.Contains(got, "beads.role") {
		t.Errorf("Stderr = %q", got)
	}
	if res.ExitStatus != 2 {
		t.Errorf("ExitStatus = %d, want 2", res.ExitStatus)
	}
	if res.Duration != 40*time.Millisecond {
		t.Errorf("Duration = %v, want 40ms", res.Duration)
	}
	if res.Invocation.Name != "bd" {
		t.Errorf("the result does not carry the invocation it answered: %+v", res.Invocation)
	}
}

func TestAnUndeclaredArgvIsAnErrorNamingWhatWasAskedAndWhatWasDeclared(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{Argv: []string{"git", "status"}})

	_, err := fake.Run(context.Background(), swarmexec.Invocation{Name: "git", Args: []string{"push"}})
	if err == nil {
		t.Fatal("Run returned no error for an argv no recording declares")
	}
	if !strings.Contains(err.Error(), "push") {
		t.Errorf("the error does not name what was asked for: %v", err)
	}
	if !strings.Contains(err.Error(), "status") {
		t.Errorf("the error does not name what was declared: %v", err)
	}
}

// Matching is on the whole argv, so a recording for a prefix of an
// invocation does not answer it. Without this, a test declaring
// `bd update` would silently satisfy `bd update --status closed`.
func TestMatchingIsOnTheWholeArgvNotAPrefix(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{Argv: []string{"bd", "update"}})

	if _, err := fake.Run(context.Background(), swarmexec.Invocation{
		Name: "bd", Args: []string{"update", "--status", "closed"},
	}); err == nil {
		t.Error("a recording for a prefix answered a longer invocation")
	}
}

func TestARecordingMayReplayAnInvocationThatNeverHappened(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv: []string{"bd"},
		Err:  &swarmexec.NotFoundError{Name: "bd"},
	})

	_, err := fake.Run(context.Background(), swarmexec.Invocation{Name: "bd"})
	if !errors.Is(err, swarmexec.ErrExecutableNotFound) {
		t.Errorf("errors.Is(err, ErrExecutableNotFound) = false for %v", err)
	}
}

// Two recordings of the same argv are consumed in order, and the last one
// repeats. That is what lets a test replay a query whose answer changes
// after a write, without the fake having to model state.
func TestRepeatedArgvsAreConsumedInOrderAndTheLastOneRepeats(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"bd", "ready"}, Stdout: []byte("first")},
		exectest.Recording{Argv: []string{"bd", "ready"}, Stdout: []byte("second")},
	)

	want := []string{"first", "second", "second"}
	for i, w := range want {
		res, err := fake.Run(context.Background(), swarmexec.Invocation{Name: "bd", Args: []string{"ready"}})
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if got := string(res.Stdout); got != w {
			t.Errorf("call %d: Stdout = %q, want %q", i, got, w)
		}
	}
}

// A recording that names a directory only answers an invocation made in
// it; one that names none answers any. A test asserting that git ran in a
// particular worktree depends on the first half.
func TestADeclaredDirectoryIsPartOfTheMatch(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv: []string{"git", "status"}, Dir: "/w/one", Stdout: []byte("in one"),
	})

	if _, err := fake.Run(context.Background(), swarmexec.Invocation{
		Name: "git", Args: []string{"status"}, Dir: "/w/two",
	}); err == nil {
		t.Error("a recording declaring one directory answered an invocation made in another")
	}

	res, err := fake.Run(context.Background(), swarmexec.Invocation{
		Name: "git", Args: []string{"status"}, Dir: "/w/one",
	})
	if err != nil {
		t.Fatalf("Run in the declared directory: %v", err)
	}
	if got := string(res.Stdout); got != "in one" {
		t.Errorf("Stdout = %q, want %q", got, "in one")
	}
}

func TestItRecordsEveryInvocationInOrderIncludingUnmatchedOnes(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{Argv: []string{"git", "status"}})

	_, _ = fake.Run(context.Background(), swarmexec.Invocation{Name: "git", Args: []string{"status"}})
	_, _ = fake.Run(context.Background(), swarmexec.Invocation{Name: "bd", Args: []string{"ready"}})

	got := fake.Argvs()
	want := [][]string{{"git", "status"}, {"bd", "ready"}}
	if len(got) != len(want) {
		t.Fatalf("Argvs() has %d entries, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if strings.Join(got[i], " ") != strings.Join(want[i], " ") {
			t.Errorf("call %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestUnusedReportsRecordingsNothingProvoked(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "status"}},
		exectest.Recording{Argv: []string{"git", "push"}},
	)

	if _, err := fake.Run(context.Background(), swarmexec.Invocation{Name: "git", Args: []string{"status"}}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	unused := fake.Unused()
	if len(unused) != 1 {
		t.Fatalf("Unused() has %d entries, want 1: %v", len(unused), unused)
	}
	if strings.Join(unused[0].Argv, " ") != "git push" {
		t.Errorf("Unused()[0].Argv = %v, want [git push]", unused[0].Argv)
	}
}
