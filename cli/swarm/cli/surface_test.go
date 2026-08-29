// Tests for the command surface: the flags every subcommand carries, the
// mapping from an outcome to one of four exit statuses, the run-scope
// refusal, and the two output forms.
//
// No subcommand exists yet — this delta delivers the surface, and the
// commands arrive with the deltas that need them — so the tests drive a
// synthetic one built from the same [Spec] a real subcommand supplies.
// That is the subject: a real command contributes a plan and an
// interpretation and nothing else, so a surface tested through a
// synthetic one is tested through everything a real one would exercise.
//
// The process runner is the fake from cli/swarm/exec/exectest, so
// "nothing was read or written" is asserted from what the runner was
// asked for rather than assumed.
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// runSurface drives the real command tree with one synthetic subcommand
// registered, through the same dispatch the binary uses.
func runSurface(t *testing.T, spec Spec, runner swarmexec.Runner, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut, exitCode: int(StatusClear)}
	root := newRootCmd(streams)
	root.AddCommand(newSubcommand(spec, streams, runner))

	code = dispatch(root, streams, args, &out, &errOut)
	return out.String(), errOut.String(), code
}

// readPlan is a one-step plan whose single invocation is a read, so a
// dry run runs it and an execute runs it.
func readPlan() plan.Plan {
	return plan.Plan{Name: "probe", Steps: []plan.Step{{
		Name: "read", Executable: "bd", Effect: plan.Read, Repeatable: true,
		Args: []plan.Arg{plan.Lit("show"), plan.Lit("run-1"), plan.Lit("--json")},
	}}}
}

func writePlan() plan.Plan {
	return plan.Plan{Name: "probe", Steps: []plan.Step{{
		Name: "write", Executable: "bd", Effect: plan.Write,
		Args: []plan.Arg{plan.Lit("update"), plan.Lit("run-1"), plan.Lit("--claim")},
	}}}
}

// specReporting is a subcommand that runs one read and reports the given
// outcome, with one fact naming the invocation behind it.
func specReporting(outcome Outcome, resolution string) Spec {
	return Spec{
		Use:       "probe",
		Short:     "a synthetic subcommand",
		RunScoped: true,
		Plan:      func(Request) (plan.Plan, error) { return readPlan(), nil },
		Interpret: func(req Request, rep plan.Report) (Result, error) {
			return Result{
				Outcome:    outcome,
				Summary:    "the probe ran",
				Resolution: resolution,
				Facts: []Fact{
					Text("branch", "claude/branch-x", []string{"bd", "show", "run-1", "--json"}),
					Unestablished("review_passes", "the key is absent", []string{"bd", "show", "run-1", "--json"}),
				},
			}, nil
		},
	}
}

func fakeForRead() *exectest.Fake {
	return exectest.NewFake(exectest.Recording{
		Argv:   []string{"bd", "show", "run-1", "--json"},
		Stdout: []byte(`{"branch":"claude/branch-x"}`),
	})
}

// AC: AC-SWARM-CORE-001.1
//
// The status is derived from the outcome rather than chosen beside it,
// and the table below is the mapping FEAT-SWARM-CORE-001 fixes. Both
// halves are asserted: the mapping itself, and that a command reporting
// the outcome really exits with it.
func TestTheExitStatusSaysWhatTheCallerShouldDoNext(t *testing.T) {
	cases := []struct {
		outcome    Outcome
		resolution string
		want       Status
	}{
		{OutcomeClear, "", StatusClear},
		{OutcomeAttention, "close the bead or mark it blocked", StatusAttention},
		{OutcomeStop, "", StatusStop},
		{OutcomeMalformed, "", StatusMalformed},
	}

	if len(cases) != len(Outcomes()) {
		t.Fatalf("the table covers %d outcomes, but there are %d", len(cases), len(Outcomes()))
	}

	for _, c := range cases {
		if got := c.outcome.Status(); got != c.want {
			t.Errorf("Outcome(%q).Status() = %d, want %d", c.outcome, got, c.want)
		}

		_, _, code := runSurface(t, specReporting(c.outcome, c.resolution), fakeForRead(),
			"probe", "--run", "run-1")
		if code != int(c.want) {
			t.Errorf("a command reporting %q exited %d, want %d", c.outcome, code, c.want)
		}
	}
}

// AC: AC-SWARM-CORE-001.2
//
// A malformed invocation and a finding are different conditions, and the
// caller routes on the status alone — so the two must not share one.
func TestAMalformedInvocationIsDistinguishableFromAFinding(t *testing.T) {
	_, _, malformed := runSurface(t, specReporting(OutcomeClear, ""), fakeForRead(),
		"probe", "--run", "run-1", "--no-such-flag")
	if malformed != int(StatusMalformed) {
		t.Errorf("an unknown flag exited %d, want %d", malformed, StatusMalformed)
	}

	_, _, finding := runSurface(t, specReporting(OutcomeAttention, "close the bead"), fakeForRead(),
		"probe", "--run", "run-1")
	if finding != int(StatusAttention) {
		t.Errorf("a command reporting a finding exited %d, want %d", finding, StatusAttention)
	}
	if malformed == finding {
		t.Errorf("a malformed invocation and a finding share exit status %d", malformed)
	}
}

// AC: AC-SWARM-CORE-001.14
//
// Both halves: the status, and that nothing was read or written. The
// second is asserted from the process runner's own record, so a surface
// that checked the scope after building and running a plan would fail
// here even though its status was right.
func TestACommandThatMutatesRunStateRefusesAnEmptyScope(t *testing.T) {
	for _, args := range [][]string{
		{"probe"},                 // no --run at all
		{"probe", "--run", ""},    // present and empty
		{"probe", "--run", "   "}, // present and blank
	} {
		fake := fakeForRead()
		_, _, code := runSurface(t, specReporting(OutcomeClear, ""), fake, args...)

		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read or wrote something outside the scope: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-CORE-001.15
//
// A refusal takes the status matching what the caller can do about it,
// which is a different question from what went wrong.
func TestARefusalTakesTheStatusMatchingWhatTheCallerCanDoAboutIt(t *testing.T) {
	cases := []struct {
		reason  string
		refusal *Refusal
		want    Status
	}{
		{"the arguments it was given are not valid",
			RefuseArguments("--key is required"), StatusMalformed},
		{"the state it read is wrong and the caller can repair it",
			RefuseRepairable("close the bead or mark it blocked", "a bead is still claimed"), StatusAttention},
		{"the state it read is wrong and the caller cannot",
			RefuseUnrepairable("the run epic carries no manifest"), StatusStop},
	}

	for _, c := range cases {
		if got := c.refusal.Outcome.Status(); got != c.want {
			t.Errorf("%s: status = %d, want %d", c.reason, got, c.want)
		}

		refusal := c.refusal
		spec := Spec{
			Use: "probe", Short: "s", RunScoped: true,
			Plan:      func(Request) (plan.Plan, error) { return plan.Plan{}, refusal },
			Interpret: func(Request, plan.Report) (Result, error) { return Result{}, nil },
		}
		fake := fakeForRead()
		stdout, _, code := runSurface(t, spec, fake, "probe", "--run", "run-1")

		if code != int(c.want) {
			t.Errorf("%s: exited %d, want %d", c.reason, code, c.want)
		}
		if !strings.Contains(stdout, refusal.Reason) {
			t.Errorf("%s: the output does not say what was refused:\n%s", c.reason, stdout)
		}
		if len(fake.Argvs()) != 0 {
			t.Errorf("%s: a refusal at plan time still ran something: %v", c.reason, fake.Argvs())
		}
	}
}

// AC: AC-SWARM-CORE-001.8
//
// An agent that can see the command behind a line can re-run it, argue
// with it, and work around it when the underlying tool changes.
func TestEveryReportedFactNamesTheCommandThatEstablishedIt(t *testing.T) {
	stdout, _, _ := runSurface(t, specReporting(OutcomeClear, ""), fakeForRead(),
		"probe", "--run", "run-1")

	want := plan.Argv([]string{"bd", "show", "run-1", "--json"})
	for _, name := range []string{"branch", "review_passes"} {
		var found bool
		for _, line := range strings.Split(stdout, "\n") {
			if strings.HasPrefix(line, name+" =") {
				found = true
				if !strings.Contains(line, want) {
					t.Errorf("the line reporting %q does not name the command that established it:\n%s", name, line)
				}
			}
		}
		if !found {
			t.Errorf("no line reports the fact %q:\n%s", name, stdout)
		}
	}
}

// A fact nobody can reproduce is what this CLI exists to stop producing,
// so the surface refuses to render one rather than printing it bare.
func TestAFactNamingNoInvocationIsRefused(t *testing.T) {
	spec := Spec{
		Use: "probe", Short: "s", RunScoped: true,
		Plan: func(Request) (plan.Plan, error) { return readPlan(), nil },
		Interpret: func(Request, plan.Report) (Result, error) {
			return Result{Outcome: OutcomeClear, Facts: []Fact{stripInvocations(Text("branch", "x"))}}, nil
		},
	}

	stdout, _, code := runSurface(t, spec, fakeForRead(), "probe", "--run", "run-1")
	if code != int(StatusStop) {
		t.Errorf("exit status = %d, want %d for a result the surface cannot render", code, StatusStop)
	}
	if !strings.Contains(stdout, "branch") {
		t.Errorf("the refusal does not name the offending fact:\n%s", stdout)
	}
}

// An attention exit whose output does not name what would resolve it
// leaves a caller with a status and nothing to act on.
func TestAnAttentionOutcomeWithoutAResolutionIsRefused(t *testing.T) {
	stdout, _, code := runSurface(t, specReporting(OutcomeAttention, ""), fakeForRead(),
		"probe", "--run", "run-1")

	if code != int(StatusStop) {
		t.Errorf("exit status = %d, want %d", code, StatusStop)
	}
	if !strings.Contains(stdout, "resolve") {
		t.Errorf("the refusal does not say what is missing:\n%s", stdout)
	}
}

func TestAnAttentionOutcomeNamesWhatWouldResolveIt(t *testing.T) {
	const resolution = "close the bead or mark it blocked"

	stdout, _, code := runSurface(t, specReporting(OutcomeAttention, resolution), fakeForRead(),
		"probe", "--run", "run-1")

	if code != int(StatusAttention) {
		t.Errorf("exit status = %d, want %d", code, StatusAttention)
	}
	if !strings.Contains(stdout, resolution) {
		t.Errorf("the output does not name what would resolve the condition:\n%s", stdout)
	}
}

// AC: AC-SWARM-CORE-001.9
//
// One value, two renderings — so the assertion is that nothing the text
// form says is missing from the machine-readable one, field by field
// rather than by spot check.
func TestTheMachineReadableFormCarriesTheSameFactsAsTheTextForm(t *testing.T) {
	text, _, textCode := runSurface(t, specReporting(OutcomeAttention, "close the bead"), fakeForRead(),
		"probe", "--run", "run-1")
	machine, _, jsonCode := runSurface(t, specReporting(OutcomeAttention, "close the bead"), fakeForRead(),
		"probe", "--run", "run-1", "--json")

	if textCode != jsonCode {
		t.Errorf("the two forms exited differently: text %d, json %d", textCode, jsonCode)
	}

	var env swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(machine), &env); err != nil {
		t.Fatalf("the machine-readable form does not parse: %v\n%s", err, machine)
	}

	if env.ContractVersion != contractVersion {
		t.Errorf("contract_version = %q, want %q", env.ContractVersion, contractVersion)
	}
	if int(env.Status) != jsonCode {
		t.Errorf("the envelope's status (%d) is not the exit status (%d)", env.Status, jsonCode)
	}

	// Every fact the text form prints is in the envelope, with the same
	// name, the same rendered value and the same invocations.
	if len(env.Facts) != 2 {
		t.Fatalf("the envelope carries %d facts, want 2:\n%s", len(env.Facts), machine)
	}
	for _, f := range env.Facts {
		if !strings.Contains(text, f.Name) {
			t.Errorf("fact %q is in the envelope and not in the text form:\n%s", f.Name, text)
		}
		for _, argv := range f.Invocations {
			if !strings.Contains(text, plan.Argv(argv)) {
				t.Errorf("fact %q names %v in the envelope and not in the text form:\n%s", f.Name, argv, text)
			}
		}
	}

	// And the fields that are not facts.
	for name, value := range map[string]string{
		"command":    env.Command,
		"run":        env.Run,
		"mode":       string(env.Mode),
		"outcome":    string(env.Outcome),
		"summary":    env.Summary,
		"resolution": env.Resolution,
	} {
		if value == "" {
			t.Errorf("the envelope's %s is empty", name)
		}
		if !strings.Contains(text, value) {
			t.Errorf("the envelope's %s (%q) is not in the text form:\n%s", name, value, text)
		}
	}

	// The invocations the command ran, with each one's status and
	// duration, are in the envelope whatever the text form does with
	// them.
	if len(env.Invocations) != 1 {
		t.Fatalf("the envelope carries %d invocations, want 1", len(env.Invocations))
	}
	if got := strings.Join(env.Invocations[0].Argv, " "); got != "bd show run-1 --json" {
		t.Errorf("the recorded invocation = %q", got)
	}
	if env.Invocations[0].StdoutDigest == "" || env.Invocations[0].StderrDigest == "" {
		t.Errorf("the recorded invocation carries no stream digests: %+v", env.Invocations[0])
	}
}

// The envelope is emitted on every exit status, so a consumer never has
// to fall back to reading text to learn why a command failed.
func TestTheEnvelopeIsEmittedOnEveryExitStatus(t *testing.T) {
	seen := map[int]bool{}

	// A refusal at each of the three non-clear statuses, plus a clear
	// run, driven through the surface with --json.
	specs := []Spec{
		specReporting(OutcomeClear, ""),
		specReporting(OutcomeAttention, "close the bead"),
		specReporting(OutcomeStop, ""),
		{Use: "probe", Short: "s", RunScoped: true,
			Plan:      func(Request) (plan.Plan, error) { return plan.Plan{}, RefuseArguments("--key is required") },
			Interpret: func(Request, plan.Report) (Result, error) { return Result{}, nil }},
	}

	for _, spec := range specs {
		stdout, _, code := runSurface(t, spec, fakeForRead(), "probe", "--run", "run-1", "--json")
		seen[code] = true

		var env swarmmodel.CommandResult
		if err := json.Unmarshal([]byte(stdout), &env); err != nil {
			t.Errorf("exit %d: the result does not parse as the envelope: %v\n%s", code, err, stdout)
			continue
		}
		if int(env.Status) != code {
			t.Errorf("exit %d: the envelope says status %d", code, env.Status)
		}
	}

	for status := 0; status <= 3; status++ {
		if !seen[status] {
			t.Errorf("no case produced exit status %d, so the envelope is unproven there", status)
		}
	}
}

// AC: AC-SWARM-CORE-001.4
//
// Explain executes nothing it prints — asserted from the process
// runner's record, which is empty because nothing reached it.
func TestExplainModePrintsThePlanAndRunsNothing(t *testing.T) {
	fake := exectest.NewFake() // no recordings: anything executed would error
	stdout, _, code := runSurface(t, specReporting(OutcomeClear, ""), fake,
		"probe", "--run", "run-1", "--explain")

	if code != int(StatusClear) {
		t.Errorf("explain exited %d, want %d", code, StatusClear)
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("explain executed %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "bd show run-1 --json") {
		t.Errorf("explain did not print the invocation it would run:\n%s", stdout)
	}
}

// AC: AC-SWARM-CORE-001.6
func TestDryRunPerformsTheReadsAndRefusesTheWrites(t *testing.T) {
	spec := specReporting(OutcomeClear, "")
	spec.Plan = func(Request) (plan.Plan, error) { return writePlan(), nil }
	spec.Interpret = func(Request, plan.Report) (Result, error) {
		return Result{Outcome: OutcomeClear,
			Facts: []Fact{Flag("done", true, []string{"bd", "update", "run-1", "--claim"})}}, nil
	}

	fake := exectest.NewFake(exectest.Recording{Argv: []string{"bd", "update", "run-1", "--claim"}})
	_, _, code := runSurface(t, spec, fake, "probe", "--run", "run-1", "--dry-run")

	if code != int(StatusClear) {
		t.Errorf("dry run exited %d, want %d", code, StatusClear)
	}
	if calls := fake.Argvs(); len(calls) != 0 {
		t.Errorf("a dry run performed a write: %v", calls)
	}
}

// The two mode flags select different interpreters, so asking for both is
// a malformed invocation rather than a silent preference for one.
func TestExplainAndDryRunTogetherAreMalformed(t *testing.T) {
	fake := exectest.NewFake()
	stdout, _, code := runSurface(t, specReporting(OutcomeClear, ""), fake,
		"probe", "--run", "run-1", "--explain", "--dry-run")

	if code != int(StatusMalformed) {
		t.Errorf("exit status = %d, want %d", code, StatusMalformed)
	}
	if len(fake.Argvs()) != 0 {
		t.Errorf("a malformed invocation still ran something: %v", fake.Argvs())
	}
	if !strings.Contains(stdout, "--explain") || !strings.Contains(stdout, "--dry-run") {
		t.Errorf("the refusal does not name both flags:\n%s", stdout)
	}
}

// The four flags are declared by the surface rather than by each
// subcommand, which is what makes "carried by every subcommand" a
// property of the code rather than a checklist.
func TestEverySubcommandCarriesTheContractFlags(t *testing.T) {
	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	cmd := newSubcommand(Spec{Use: "probe", Short: "s"}, streams, exectest.NewFake())

	for _, name := range []string{"explain", "dry-run", "json", "run"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("--%s is not declared on the subcommand", name)
		}
	}
}

// A subcommand that is not run-scoped does not demand a scope. Without
// this, the scope check would be indistinguishable from one that always
// refuses.
func TestASubcommandThatIsNotRunScopedNeedsNoScope(t *testing.T) {
	spec := specReporting(OutcomeClear, "")
	spec.RunScoped = false

	_, _, code := runSurface(t, spec, fakeForRead(), "probe")
	if code != int(StatusClear) {
		t.Errorf("exit status = %d, want %d", code, StatusClear)
	}
}

// An outcome this package does not recognise is a condition the caller
// cannot repair in place, not a success.
func TestAnUnrecognisedOutcomeIsNotASuccess(t *testing.T) {
	if got := Outcome("something else").Status(); got != StatusStop {
		t.Errorf("an unrecognised outcome maps to %d, want %d", got, StatusStop)
	}
}

// The surface's help text is cobra's own rendering of the declarations,
// so a flag added here appears in it without a second edit.
func TestTheSubcommandHelpNamesTheContractFlags(t *testing.T) {
	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	cmd := newSubcommand(Spec{Use: "probe", Short: "s"}, streams, exectest.NewFake())
	cmd.InitDefaultHelpFlag()

	usage := cmd.UsageString()
	for _, want := range []string{"--explain", "--dry-run", "--json", "--run"} {
		if !strings.Contains(usage, want) {
			t.Errorf("the usage does not mention %s:\n%s", want, usage)
		}
	}
}

// A guard against the registration seam quietly acquiring a second wiring
// path: whatever register adds must be reachable from the tree Execute
// builds.
func TestRegisterAddsWhateverTheTreeOffers(t *testing.T) {
	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	root := newRootCmd(streams)
	before := len(root.Commands())
	register(root, streams, exectest.NewFake())

	if len(root.Commands()) < before {
		t.Error("register removed a command from the tree")
	}
	for _, c := range root.Commands() {
		if c.Flags().Lookup("json") == nil {
			t.Errorf("subcommand %q does not carry --json, so it was not built by newSubcommand", c.Name())
		}
	}
}

// stripInvocations returns a fact with no invocation behind it, for the
// one test that needs a result the surface must refuse to render.
func stripInvocations(f Fact) Fact {
	f.Invocations = nil
	return f
}
