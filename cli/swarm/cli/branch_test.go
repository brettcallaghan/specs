// Unit tests for `swarm branch verify-landed` and `swarm branch push`:
// the fetch ordered before the comparison, the landing check needing no
// run scope, and a push refused for one path reported as a whole
// rejected push. Every scenario runs through the Git Version Control's
// real argv-building against the fake process runner, so no test starts
// `git`.
//
// AC-SWARM-DISP-005.1, .2, .3 and .4 are declared e2e and live in
// cli/swarm/e2e against the core and push-refused fixtures.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

const (
	landBranch = "sprint/run-1"
	landRemote = "origin"
)

// runBranch drives the real command tree against the fake and returns
// stdout, stderr and the exit status.
func runBranch(t *testing.T, fake *exectest.Fake, args ...string) (string, string, int) {
	t.Helper()

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(branchCmd(streams, gitvcs.VersionControl{}, fake))

	code := dispatch(root, streams, args, &out, &errOut)
	return out.String(), errOut.String(), code
}

// landedReads records the three invocations the landing check makes for a
// branch that is in step with its remote.
func landedReads(ahead, behind string) []exectest.Recording {
	return []exectest.Recording{
		{Argv: []string{"git", "fetch", landRemote}},
		{
			Argv:   []string{"git", "rev-parse", "--verify", landRemote + "/" + landBranch + "^{commit}"},
			Stdout: []byte(strings.Repeat("a", 40) + "\n"),
		},
		{
			Argv:   []string{"git", "rev-list", "--count", "--left-right", landBranch + "..." + landRemote + "/" + landBranch},
			Stdout: []byte(ahead + "\t" + behind + "\n"),
		},
	}
}

// AC: AC-SWARM-DISP-005.5
//
// The remote is read afresh before the comparison: the fetch is a step of
// the same plan, ordered ahead of both reads, so no caller can compare
// against a stale record of where the remote points.
func TestTheRemoteIsFetchedBeforeTheComparison(t *testing.T) {
	fake := exectest.NewFake(landedReads("0", "0")...)

	stdout, _, code := runBranch(t, fake,
		"branch", "verify-landed", "--branch", landBranch, "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	fetchIdx, resolveIdx, compareIdx := -1, -1, -1
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 2 && argv[1] == "fetch":
			fetchIdx = i
		case len(argv) >= 2 && argv[1] == "rev-parse":
			resolveIdx = i
		case len(argv) >= 2 && argv[1] == "rev-list":
			compareIdx = i
		}
	}
	if fetchIdx < 0 || resolveIdx < 0 || compareIdx < 0 {
		t.Fatalf("fetch@%d resolve@%d compare@%d; all three are required: %v",
			fetchIdx, resolveIdx, compareIdx, fake.Argvs())
	}
	if !(fetchIdx < resolveIdx && fetchIdx < compareIdx) {
		t.Errorf("the fetch ran at call %d, the reads at %d and %d; the fetch comes first",
			fetchIdx, resolveIdx, compareIdx)
	}
}

// AC: AC-SWARM-DISP-005.6
//
// The landing check needs no run scope: with no --run at all it still
// reports the branch's landing state, where every other command in this
// surface refuses the invocation outright.
func TestTheLandingCheckNeedsNoRunScope(t *testing.T) {
	for _, verb := range []string{"verify-landed"} {
		fake := exectest.NewFake(landedReads("0", "0")...)
		stdout, _, code := runBranch(t, fake, "branch", verb, "--branch", landBranch, "--json")
		if code != int(StatusClear) {
			t.Fatalf("%s with no --run exited %d, want %d\n%s", verb, code, StatusClear, stdout)
		}
		if !strings.Contains(stdout, `"landed": true`) {
			t.Errorf("%s did not report the branch's landing state:\n%s", verb, stdout)
		}
		if len(fake.Argvs()) == 0 {
			t.Errorf("%s ran nothing, so it refused the unscoped invocation after all", verb)
		}
	}

	// And the contrast: a run-scoped command in the same tree refuses.
	fake := exectest.NewFake()
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	register(root, streams, fake)
	if code := dispatch(root, streams, []string{"ready"}, &out, &errOut); code != int(StatusMalformed) {
		t.Errorf("`swarm ready` with no --run exited %d, want %d; the branch commands are the exception, not the rule",
			code, StatusMalformed)
	}
}

// AC: AC-SWARM-DISP-005.7
//
// A push the remote refuses for one path is reported as a whole rejected
// push: no commit in it is reported as landed, the refused ref is named,
// and the remote's own cause reaches the caller.
func TestAPushRefusedForOnePathLandsNothing(t *testing.T) {
	const stderr = `remote: refusing refs/heads/claude/*: this remote does not accept sprint branches
To ../remote.git
 ! [remote rejected] sprint/run-1 -> sprint/run-1 (pre-receive hook declined)
error: failed to push some refs to '../remote.git'
`
	fake := exectest.NewFake(exectest.Recording{
		Argv:       []string{"git", "push", landRemote, landBranch},
		Stderr:     []byte(stderr),
		ExitStatus: 1,
	})

	stdout, _, code := runBranch(t, fake,
		"branch", "push", "--branch", landBranch, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (stop)\n%s", code, StatusStop, stdout)
	}

	// Nothing landed, and the report says so rather than leaving it
	// unestablished.
	if !strings.Contains(stdout, `"landed"`) || strings.Contains(stdout, `"landed": true`) {
		t.Errorf("the report does not say that nothing landed:\n%s", stdout)
	}
	// The refused ref is named, and so is the remote's own reason.
	if !strings.Contains(stdout, "sprint/run-1") {
		t.Errorf("the report does not name the refused ref:\n%s", stdout)
	}
	if !strings.Contains(stdout, "this remote does not accept sprint branches") {
		t.Errorf("the report does not carry the rejection's cause:\n%s", stdout)
	}

	// A rejected push stops the plan: nothing after it ran, so no
	// comparison could report a landing.
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && (argv[1] == "fetch" || argv[1] == "rev-list") {
			t.Errorf("a step after the rejected push ran: %v", argv)
		}
	}
}

// `push` IS `verify-landed` with the push in front of it: an accepted
// push runs the same three invocations the check runs, in the same order,
// after the push.
func TestPushIsTheLandingCheckWithThePushInFrontOfIt(t *testing.T) {
	recs := append([]exectest.Recording{{Argv: []string{"git", "push", landRemote, landBranch}}},
		landedReads("0", "0")...)
	pushFake := exectest.NewFake(recs...)
	stdout, _, code := runBranch(t, pushFake, "branch", "push", "--branch", landBranch, "--json")
	if code != int(StatusClear) {
		t.Fatalf("push exited %d, want %d\n%s", code, StatusClear, stdout)
	}

	checkFake := exectest.NewFake(landedReads("0", "0")...)
	if _, _, code := runBranch(t, checkFake, "branch", "verify-landed", "--branch", landBranch, "--json"); code != int(StatusClear) {
		t.Fatalf("verify-landed exited %d, want %d", code, StatusClear)
	}

	pushed, checked := pushFake.Argvs(), checkFake.Argvs()
	if len(pushed) != len(checked)+1 {
		t.Fatalf("push ran %d invocation(s) and verify-landed %d; push is the check plus one\npush:  %v\ncheck: %v",
			len(pushed), len(checked), pushed, checked)
	}
	if pushed[0][1] != "push" {
		t.Errorf("push's first invocation is %v, want the push itself", pushed[0])
	}
	for i, argv := range checked {
		if strings.Join(pushed[i+1], " ") != strings.Join(argv, " ") {
			t.Errorf("after the push, invocation %d is %v; verify-landed's is %v", i, pushed[i+1], argv)
		}
	}
}

// WHERE --branch is absent the branch is the one checked out, and WHERE
// --remote is absent the remote is origin — both read from the
// repository, neither from configuration outside it.
func TestTheDefaultsAreTheCheckedOutBranchAndOrigin(t *testing.T) {
	recs := append([]exectest.Recording{{
		Argv:   []string{"git", "rev-parse", "--abbrev-ref", "HEAD"},
		Stdout: []byte(landBranch + "\n"),
	}}, landedReads("0", "0")...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runBranch(t, fake, "branch", "verify-landed", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	answer := acceptAnswer(t, stdout)
	if answer["branch"] != landBranch {
		t.Errorf("branch = %v, want the checked-out branch %q", answer["branch"], landBranch)
	}
	if answer["remote"] != landRemote {
		t.Errorf("remote = %v, want %q", answer["remote"], landRemote)
	}
}

// A repository on no branch is refused rather than compared against a
// remote ref named `HEAD`: `git rev-parse --abbrev-ref HEAD` answers the
// literal string `HEAD` when detached, and reading it as a branch name
// would produce a true report about the wrong question.
func TestADetachedHeadIsRefusedRatherThanCheckedAsABranch(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:   []string{"git", "rev-parse", "--abbrev-ref", "HEAD"},
		Stdout: []byte("HEAD\n"),
	})

	stdout, _, code := runBranch(t, fake, "branch", "verify-landed", "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	if !strings.Contains(stdout, "--branch") {
		t.Errorf("the refusal does not name what would resolve it:\n%s", stdout)
	}
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[1] == "fetch" {
			t.Errorf("a detached repository was fetched for anyway: %v", argv)
		}
	}
}

// Explain runs nothing at all, for either verb, and `push`'s explained
// sequence carries the push ahead of the check's own steps.
func TestBranchExplainRunsNothing(t *testing.T) {
	for _, verb := range []string{"verify-landed", "push"} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		stdout, _, code := runBranch(t, fake, "branch", verb, "--branch", landBranch, "--explain")
		if code != int(StatusClear) {
			t.Fatalf("%s --explain exited %d, want %d\n%s", verb, code, StatusClear, stdout)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%s --explain ran %d invocation(s): %v", verb, len(calls), calls)
		}
		for _, want := range []string{"fetch", "rev-list --count --left-right"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("%s --explain does not mention %q:\n%s", verb, want, stdout)
			}
		}
		if pushed := strings.Contains(stdout, "git push"); pushed != (verb == "push") {
			t.Errorf("%s --explain mentions a push: %v", verb, pushed)
		}
	}
}

// A dry run refuses the fetch — it is a write — while the two reads that
// compare still run, against whatever the local tracking ref happens to
// hold. So the comparison a dry run can make is exactly the one this
// command exists to prevent: a stale local record reporting a branch as
// landed.
//
// The reads below answer as they would for a branch in step with its
// counterpart, which is what a STALE ref looks like from inside the
// process — indistinguishable from a fresh one, which is why the verdict
// has to come from whether the fetch ran rather than from what the reads
// said.
func TestADryRunDoesNotReportLandedFromAnUnfetchedComparison(t *testing.T) {
	for _, verb := range []string{"verify-landed", "push"} {
		fake := exectest.NewFake(landedReads("0", "0")...)

		stdout, _, code := runBranch(t, fake,
			"branch", verb, "--branch", landBranch, "--dry-run", "--full-result", "--json")
		if code != int(StatusAttention) {
			t.Fatalf("%s --dry-run exited %d, want %d (attention)\n%s",
				verb, code, StatusAttention, stdout)
		}
		for _, argv := range fake.Argvs() {
			if len(argv) >= 2 && (argv[1] == "fetch" || argv[1] == "push") {
				t.Errorf("%s --dry-run ran a write: %v", verb, argv)
			}
		}

		env := envelopeOf(t, stdout)
		landed, ok := factNamed(env, "landed")
		if !ok {
			t.Fatalf("%s --dry-run reports no landing verdict at all:\n%s", verb, stdout)
		}
		if landed.Established {
			t.Errorf("%s --dry-run reports landed as established (%v) from a comparison made "+
				"without fetching; a stale tracking ref is exactly what that cannot distinguish",
				verb, landed.Value)
		}
		if !strings.Contains(landed.Reason, "fetch") {
			t.Errorf("%s --dry-run does not say the fetch is what is missing: %q", verb, landed.Reason)
		}
		if !strings.Contains(env.Resolution, "--dry-run") {
			t.Errorf("%s --dry-run does not name what would settle the question: %q",
				verb, env.Resolution)
		}
	}
}

// A fetch that really ran and failed leaves the tracking ref exactly as
// stale as a refused one does, and the two reads still compare against
// it. The guarantee is that the remote was consulted, not that a fetch
// step was present in the plan.
func TestAFailedFetchDoesNotReportLandedEither(t *testing.T) {
	recs := landedReads("0", "0")
	recs[0] = exectest.Recording{
		Argv:       []string{"git", "fetch", landRemote},
		ExitStatus: 128,
		Stderr:     []byte("fatal: unable to access 'origin': Could not resolve host"),
	}
	fake := exectest.NewFake(recs...)

	stdout, _, code := runBranch(t, fake,
		"branch", "verify-landed", "--branch", landBranch, "--full-result", "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	landed, ok := factNamed(envelopeOf(t, stdout), "landed")
	if !ok {
		t.Fatalf("no landing verdict at all:\n%s", stdout)
	}
	if landed.Established {
		t.Errorf("a failed fetch still produced an established verdict (%v); the comparison "+
			"rests on the same stale ref a refused fetch would leave", landed.Value)
	}
}

// The negative control: with the fetch actually run, the same reads
// produce the ordinary verdict. So the tests above assert that the fetch
// ran, and not that the command stopped reporting.
func TestAFetchThatRanStillProducesTheOrdinaryVerdict(t *testing.T) {
	fake := exectest.NewFake(landedReads("0", "0")...)

	stdout, _, code := runBranch(t, fake,
		"branch", "verify-landed", "--branch", landBranch, "--full-result", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	landed, ok := factNamed(envelopeOf(t, stdout), "landed")
	if !ok || !landed.Established || landed.Value == nil || landed.Value.Flag == nil || !*landed.Value.Flag {
		t.Errorf("a fetched comparison does not report landed: %v\n%s", landed, stdout)
	}
}
