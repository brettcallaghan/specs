// Tests for the retry that keeps an atomic replacement working on
// Windows. They are internal because the loop is unexported, and they
// exist because nothing else can reach it: [sharingConflict] is false by
// construction on every platform but Windows, so a Linux or macOS run
// executes not one line of the loop through the ordinary path. Left
// untested, a bounded retry is exactly the shape of code that degrades
// into a no-op or an unbounded wait without anyone noticing until the one
// platform that exercises it fails in CI.
package corpus

import (
	"errors"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// errTransient and transient stand in for the Windows conflict and
// [sharingConflict], so the loop's mechanics can be exercised anywhere.
var errTransient = errors.New("transient")

func transient(err error) bool { return errors.Is(err, errTransient) }

// A call that succeeds outright is made once and returns immediately.
func TestRetryWhileDoesNotRetryASuccess(t *testing.T) {
	calls := 0
	started := time.Now()

	err := retryWhile(time.Second, transient, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("op called %d times, want 1", calls)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Errorf("a successful call waited %s; it should not sleep at all", elapsed)
	}
}

// An error the predicate rejects is reported at once. This is the
// negative control for the test below: without it, a loop that retried
// everything would look identical from the outside, and a genuine
// permission failure on POSIX would be reported seconds late.
func TestRetryWhileDoesNotRetryANonRetryableError(t *testing.T) {
	permanent := errors.New("permanent")
	calls := 0
	started := time.Now()

	err := retryWhile(time.Second, transient, func() error {
		calls++
		return permanent
	})

	if !errors.Is(err, permanent) {
		t.Errorf("err = %v, want the permanent error", err)
	}
	if calls != 1 {
		t.Errorf("op called %d times, want 1", calls)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Errorf("a non-retryable error waited %s; it should not sleep at all", elapsed)
	}
}

// A conflict that clears is retried until it does, and the eventual
// success is what the caller sees.
func TestRetryWhileReturnsTheFirstSuccessAfterAConflict(t *testing.T) {
	calls := 0

	err := retryWhile(time.Second, transient, func() error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})

	if err != nil {
		t.Errorf("err = %v, want nil once the conflict cleared", err)
	}
	if calls != 3 {
		t.Errorf("op called %d times, want 3", calls)
	}
}

// A conflict that never clears is given up on, within the budget, and
// reports the condition that was waited on rather than a timeout of the
// loop's own invention. The budget is the guarantee that matters: without
// it this is an unbounded wait on a file some other process is holding.
func TestRetryWhileGivesUpWithinItsBudget(t *testing.T) {
	const budget = 50 * time.Millisecond
	calls := 0
	started := time.Now()

	err := retryWhile(budget, transient, func() error {
		calls++
		return errTransient
	})
	elapsed := time.Since(started)

	if !errors.Is(err, errTransient) {
		t.Errorf("err = %v, want the last error the op returned", err)
	}
	if calls < 2 {
		t.Errorf("op called %d times; the loop did not retry at all", calls)
	}
	// One backoff may straddle the deadline, and the cap on that backoff
	// is 50ms, so the budget plus one capped sleep plus slack is the
	// honest ceiling.
	if ceiling := budget + 50*time.Millisecond + 500*time.Millisecond; elapsed > ceiling {
		t.Errorf("the loop ran for %s against a %s budget", elapsed, budget)
	}
}

// sharingConflict recognises the two Windows codes a contended
// replacement produces, and nothing at all off Windows.
//
// AC: AC-SPEC-010.4
func TestSharingConflictRecognisesTheWindowsCodesOnly(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"access denied", &os.PathError{Op: "rename", Err: syscall.Errno(5)}, runtime.GOOS == "windows"},
		{"sharing violation", &os.PathError{Op: "open", Err: errorSharingViolation}, runtime.GOOS == "windows"},
		{"not exist", &os.PathError{Op: "open", Err: syscall.ENOENT}, false},
		{"unrelated", errors.New("something else"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharingConflict(tc.err); got != tc.want {
				t.Errorf("sharingConflict(%v) on %s = %t, want %t", tc.err, runtime.GOOS, got, tc.want)
			}
		})
	}
}
