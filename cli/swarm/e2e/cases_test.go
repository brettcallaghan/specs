package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// AC: AC-SWARM-CORE-001.6
// AC: AC-SWARM-CORE-002.16
//
// Replays every case under cli/testdata/swarm-cases against the compiled
// binary and a restored fixture. A case is pure data — case.toml and the
// expected artifacts it names — so adding one is a reviewable diff rather
// than more Go.
func TestSwarmCases(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", casesDir, err)
	}
	var cases int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cases++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			runCase(t, binary, filepath.Join(casesDir, entry.Name()))
		})
	}
	if cases == 0 {
		t.Fatalf("no cases under %s, so this tier asserts nothing", casesDir)
	}
}

func runCase(t *testing.T, binary, dir string) {
	t.Helper()

	var manifest caseManifest
	if _, err := toml.DecodeFile(filepath.Join(dir, "case.toml"), &manifest); err != nil {
		t.Fatalf("reading the case: %v", err)
	}
	if manifest.Fixture == "" || len(manifest.Args) == 0 {
		t.Fatal("a case declares a fixture and an argument vector")
	}

	w := restore(t, binary, manifest.Fixture)

	stdout, stderr, code := w.swarm(t, manifest.Args...)
	if code != manifest.Exit {
		t.Errorf("exit status = %d, want %d\n--- stdout ---\n%s\n--- stderr ---\n%s",
			code, manifest.Exit, stdout, stderr)
	}
	compare(t, dir, "stdout", normalized(t, stdout, manifest.Normalize))

	for i, after := range manifest.After {
		stdout, stderr, code := w.swarm(t, after.Args...)
		if code != after.Exit {
			t.Errorf("after[%d] exit status = %d, want %d\n--- stdout ---\n%s\n--- stderr ---\n%s",
				i, code, after.Exit, stdout, stderr)
		}
		compare(t, dir, "after-"+itoa(i)+".stdout", normalized(t, stdout, manifest.Normalize))
	}
}

// compare holds actual output to the artifact the case committed.
//
// A missing artifact is a failure that prints what would have been
// written, so recording a new case is reading the failure rather than
// running a command that writes whatever the code currently does.
func compare(t *testing.T, dir, name, actual string) {
	t.Helper()

	path := filepath.Join(dir, "expected", name)
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v\n--- what the run produced ---\n%s", path, err, actual)
	}
	if actual != string(want) {
		t.Errorf("%s differs from what the case committed.\n--- got ---\n%s\n--- want ---\n%s",
			name, actual, want)
	}
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

// AC: AC-SWARM-CORE-001.13
//
// Explain output is a runnable transcript: the invocations it prints,
// run by hand in order, perform what the command would have performed.
//
// "Run by hand" is what the test does — it executes each printed
// invocation as an argv vector, exactly as a reader pasting them would —
// and then asserts the world is the one an ordinary execute would have
// left. Without the second half the test would only prove the transcript
// runs, not that it does the same thing.
func TestExplainOutputIsARunnableTranscript(t *testing.T) {
	requireTools(t)
	binary := buildBinary(t)

	args := []string{
		"manifest", "set", "--run", "fixture-run",
		"--key", "findings_path", "--value", `"docs/retros/fixture.md"`,
	}

	// One world, explained and then run by hand.
	byHand := restore(t, binary, "core")
	transcript, stderr, code := byHand.swarm(t, append(args, "--explain", "--json")...)
	if code != 0 {
		t.Fatalf("explain exited %d\n%s\n%s", code, transcript, stderr)
	}

	invocations := runnableLines(t, transcript)
	if len(invocations) < 2 {
		t.Fatalf("the transcript prints %d invocations, want the read and the write:\n%s",
			len(invocations), transcript)
	}
	for _, line := range invocations {
		byHand.mustRun(t, byHand.repo, "sh", "-c", line)
	}

	// The other world, run the ordinary way.
	executed := restore(t, binary, "core")
	if _, stderr, code := executed.swarm(t, args...); code != 0 {
		t.Fatalf("execute exited %d\n%s", code, stderr)
	}

	// The two worlds now agree about the key the command wrote, which is
	// what "perform what the command would have performed" means.
	byHandValue := readKey(t, byHand, "findings_path")
	executedValue := readKey(t, executed, "findings_path")
	if byHandValue != executedValue {
		t.Errorf("running the transcript by hand left %q; running the command left %q",
			byHandValue, executedValue)
	}
	if byHandValue != "docs/retros/fixture.md" {
		t.Errorf("running the transcript by hand left %q, want the value the command writes", byHandValue)
	}

	// The negative control: a world where neither happened does NOT carry
	// the key, so the comparison above is not two empty answers agreeing.
	untouched := restore(t, binary, "core")
	if value := readKey(t, untouched, "findings_path"); strings.Contains(value, "docs/retros/fixture.md") {
		t.Errorf("the untouched fixture already carries the key: %q", value)
	}
}

// runnableLines is the transcript with its comments and blank lines
// removed, read out of the command result --explain renders.
//
// The result rather than the printed lines: --explain implies
// --full-result, so the transcript arrives as a field of a document
// rather than as the whole of stdout, and reading the field is what makes
// this independent of how that document is laid out.
func runnableLines(t *testing.T, result string) []string {
	t.Helper()

	var env swarmmodel.CommandResult
	if err := json.Unmarshal([]byte(result), &env); err != nil {
		t.Fatalf("the explained result does not parse: %v\n%s", err, result)
	}
	var out []string
	for _, line := range env.Transcript {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

// readKey reads one manifest key out of a world, through the binary
// itself, and returns the value.
//
// The value alone, not the whole result: the result also carries the
// elapsed time of the read and a digest of what the tracker printed, and
// two worlds that agree about the key will differ in both. Comparing the
// value is what the claim is about.
//
// A key nobody has written exits 1 and its answer carries no value, which
// is returned as the empty string rather than failing: the negative
// control below reads a key it expects to be absent.
func readKey(t *testing.T, w world, key string) string {
	t.Helper()

	stdout, stderr, code := w.swarm(t, "manifest", "get", "--run", "fixture-run", "--key", key, "--json")
	switch code {
	case 0:
	case 1:
		return ""
	default:
		t.Fatalf("`manifest get --key %s` exited %d\n%s\n%s", key, code, stdout, stderr)
	}

	var answer map[string]any
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	value, present := answer[key]
	if !present {
		return ""
	}
	return fmt.Sprint(value)
}

// Every fixture records the version of the contract it was recorded
// against, and the kind it is says how that version must compare with the
// current one.
//
// The core's must EQUAL it: a core fixture whose stamp has fallen behind
// is a re-record that was missed, and the cases would be asserting a
// shape the code no longer writes. A legacy fixture's must be OLDER: one
// whose version has caught up is a legacy fixture somebody re-recorded,
// which destroys the only evidence that reading an older shape works.
func TestEveryFixtureStampAgreesWithItsKind(t *testing.T) {
	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", fixturesDir, err)
	}

	var cores int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var stamp fixtureStamp
		path := filepath.Join(fixturesDir, entry.Name(), "fixture.toml")
		if _, err := toml.DecodeFile(path, &stamp); err != nil {
			t.Errorf("reading %s: %v", path, err)
			continue
		}

		switch stamp.Kind {
		case "core":
			cores++
			if stamp.ContractVersion != contract.Version {
				t.Errorf("%s is a core fixture recorded against contract %s, and the current version is %s; re-record it",
					entry.Name(), stamp.ContractVersion, contract.Version)
			}
		case "variant":
			if stamp.ContractVersion != contract.Version {
				t.Errorf("%s is a variant recorded against contract %s, and the current version is %s",
					entry.Name(), stamp.ContractVersion, contract.Version)
			}
		case "legacy":
			if stamp.ContractVersion == contract.Version {
				t.Errorf("%s is a legacy fixture whose contract version has caught up with the current one (%s); "+
					"a legacy fixture is never re-recorded, because re-recording it destroys the only evidence "+
					"that reading an older shape works",
					entry.Name(), contract.Version)
			}
		default:
			t.Errorf("%s declares kind %q; a fixture is core, variant or legacy", entry.Name(), stamp.Kind)
		}
	}

	if cores != 1 {
		t.Errorf("there are %d core fixtures; there is one canonical run", cores)
	}
}
