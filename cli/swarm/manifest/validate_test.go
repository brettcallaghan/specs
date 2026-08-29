// Tests for the rules no single write can check on its own.
package manifest_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// doc builds a record from raw JSON values.
func doc(t *testing.T, values map[string]string) manifest.Document {
	t.Helper()
	return manifest.NewDocument(raw(t, values))
}

// rules returns the rules a report's failures cite, so a test asserts
// what was found rather than how it was worded.
func rules(report manifest.Report) []manifest.Rule {
	var out []manifest.Rule
	for _, f := range report.Failures() {
		out = append(out, f.Rule)
	}
	return out
}

func hasRule(report manifest.Report, want manifest.Rule) bool {
	for _, r := range rules(report) {
		if r == want {
			return true
		}
	}
	return false
}

// AC: AC-SWARM-CORE-002.6
//
// The branch and the commit it was cut from are the run's identity: the
// review scopes its diff from the base, every worktree branches from the
// branch, and the pull request opens against it. A run that repointed
// either mid-flight would leave those three describing different trees.
func TestTheBranchAndItsBaseAreWrittenOnce(t *testing.T) {
	current := doc(t, map[string]string{
		"branch":   `"claude/bead-bonanza-run1"`,
		"base_sha": `"3a57dc733e8e8ee0efbd4cb58e1880a492f4e502"`,
	})

	for name, change := range map[string]map[string]string{
		"a different branch": {"branch": `"claude/other"`},
		"a different base":   {"base_sha": `"deadbeef"`},
	} {
		report := manifest.CheckWrite(current, raw(t, change))
		if !report.Failed() {
			t.Errorf("%s: the write was accepted", name)
			continue
		}
		if !hasRule(report, manifest.RuleWriteOnce) {
			t.Errorf("%s: findings are %v, want a write-once refusal", name, rules(report))
		}
		// The refusal names what is already recorded, so a caller can see
		// it is not the value it expected rather than only that it lost.
		if !strings.Contains(report.Error(), "claude/bead-bonanza-run1") &&
			!strings.Contains(report.Error(), "3a57dc73") {
			t.Errorf("%s: the refusal does not name the recorded value: %s", name, report.Error())
		}
	}

	// Writing the SAME value is not a change, so a resumed run that
	// re-writes what it already recorded is not refused — which is what
	// makes a run reuse its branch rather than fail on finding one.
	same := manifest.CheckWrite(current, raw(t, map[string]string{
		"branch": `"claude/bead-bonanza-run1"`,
	}))
	if same.Failed() {
		t.Errorf("re-writing the recorded branch was refused: %s", same.Error())
	}

	// And a record with no branch yet accepts one, so the rule is
	// write-ONCE rather than write-never.
	first := manifest.CheckWrite(manifest.Document{}, raw(t, map[string]string{
		"branch": `"claude/bead-bonanza-run1"`,
	}))
	if first.Failed() {
		t.Errorf("the first write of the branch was refused: %s", first.Error())
	}
}

// AC: AC-SWARM-CORE-002.4
// AC: AC-SWARM-CORE-002.7
//
// The counter is durable for one reason: a resumed run that restarted it
// would loop forever. So it reads back what was spent, never decreases,
// and never passes the budget recorded beside it.
func TestTheReviewCounterSurvivesAResumedRunAndOnlyGoesForward(t *testing.T) {
	current := doc(t, map[string]string{"review_passes": `2`, "review_budget": `2`})

	// It reads back what was spent, not zero.
	var spent int
	present, err := current.Decode("review_passes", &spent)
	if err != nil || !present || spent != 2 {
		t.Fatalf("a resumed run read review_passes as %d (present=%v, err=%v), want 2", spent, present, err)
	}

	for name, change := range map[string]map[string]string{
		"a lower count": {"review_passes": `1`},
		"a third pass":  {"review_passes": `3`},
	} {
		report := manifest.CheckWrite(current, raw(t, change))
		if !report.Failed() {
			t.Errorf("%s: the write was accepted", name)
		}
	}

	if report := manifest.CheckWrite(current, raw(t, map[string]string{"review_passes": `1`})); !hasRule(report, manifest.RuleCounterBackwards) {
		t.Errorf("a lower count cited %v, want a backwards refusal", rules(report))
	}
	if report := manifest.CheckWrite(current, raw(t, map[string]string{"review_passes": `3`})); !hasRule(report, manifest.RuleCounterOverBudget) {
		t.Errorf("a third pass cited %v, want an over-budget refusal", rules(report))
	}

	// The negative control: spending the budget's last pass is allowed,
	// so the rule is not "refuse every write to the counter".
	partway := doc(t, map[string]string{"review_passes": `1`, "review_budget": `2`})
	if report := manifest.CheckWrite(partway, raw(t, map[string]string{"review_passes": `2`})); report.Failed() {
		t.Errorf("spending the last pass was refused: %s", report.Error())
	}
}

// AC: AC-SWARM-CORE-002.8
//
// Dispatch prompts cite ledger entries by position, so an entry removed
// or reordered retargets every citation that addressed it — and nothing
// downstream can tell.
func TestTheEnvironmentalLedgerIsAppendOnly(t *testing.T) {
	current := doc(t, map[string]string{
		"ledger": `["first — proved by a", "second — proved by b", "third — proved by c"]`,
	})

	for name, change := range map[string]map[string]string{
		"an entry removed":   {"ledger": `["first — proved by a", "second — proved by b"]`},
		"entries reordered":  {"ledger": `["second — proved by b", "first — proved by a", "third — proved by c"]`},
		"an entry rewritten": {"ledger": `["first — proved by z", "second — proved by b", "third — proved by c"]`},
		"emptied":            {"ledger": `[]`},
	} {
		report := manifest.CheckWrite(current, raw(t, change))
		if !report.Failed() {
			t.Errorf("%s: the write was accepted", name)
			continue
		}
		if !hasRule(report, manifest.RuleLedgerRewritten) {
			t.Errorf("%s: findings are %v, want a ledger refusal", name, rules(report))
		}
	}

	// Appending is the whole point, so it is accepted — twice over, to
	// show the rule is about the prefix rather than about the length.
	for _, change := range []string{
		`["first — proved by a", "second — proved by b", "third — proved by c", "fourth — proved by d"]`,
		`["first — proved by a", "second — proved by b", "third — proved by c"]`,
	} {
		if report := manifest.CheckWrite(current, raw(t, map[string]string{"ledger": change})); report.Failed() {
			t.Errorf("an append was refused: %s", report.Error())
		}
	}
}

// AC: AC-SWARM-CORE-002.12
//
// A bead in two admission lists makes the run's own account of its scope
// self-contradictory, and the finding names the bead rather than the
// lists — the bead is what somebody has to decide about.
func TestABeadInMoreThanOneAdmissionListIsNamed(t *testing.T) {
	record := doc(t, map[string]string{
		"admitted": `["run-1.1", "run-1.2", "run-1.3"]`,
		"excluded": `[{"id": "run-1.2", "reason": "declined move out of other-epic"}]`,
		"moved":    `[{"id": "run-1.3", "from": "other-epic"}]`,
	})

	report := manifest.Validate(record, swarmmodel.RunStageScope)
	if !report.Failed() {
		t.Fatal("a bead in two admission lists validated clean")
	}

	var named []string
	for _, f := range report.Failures() {
		if f.Rule == manifest.RuleAdmissionOverlap {
			named = append(named, f.Keys...)
		}
	}
	if len(named) != 2 {
		t.Fatalf("the overlap findings name %v, want run-1.2 and run-1.3", named)
	}
	for _, want := range []string{"run-1.2", "run-1.3"} {
		if !strings.Contains(strings.Join(named, " "), want) {
			t.Errorf("no finding names %s", want)
		}
	}

	// The negative control: three lists that do not overlap are clean.
	clean := doc(t, map[string]string{
		"admitted": `["run-1.1"]`,
		"excluded": `[{"id": "run-1.2", "reason": "declined"}]`,
		"moved":    `[{"id": "run-1.3", "from": "other-epic"}]`,
	})
	for _, f := range manifest.Validate(clean, swarmmodel.RunStageScope).Failures() {
		if f.Rule == manifest.RuleAdmissionOverlap {
			t.Errorf("non-overlapping lists produced %v", f)
		}
	}

	// And a write that would create the overlap is refused before it
	// happens, not only reported afterwards.
	if report := manifest.CheckWrite(clean, raw(t, map[string]string{
		"admitted": `["run-1.1", "run-1.3"]`,
	})); !hasRule(report, manifest.RuleAdmissionOverlap) {
		t.Errorf("a write creating an overlap cited %v", rules(report))
	}
}

// AC: AC-SWARM-CORE-002.11
//
// A harness that exposes no effort setting records it as unrecorded
// rather than leaving the key absent, so a reader can tell "not offered"
// from "not yet written".
func TestASettingTheHarnessDoesNotExposeIsRecordedAsUnrecorded(t *testing.T) {
	absent := doc(t, map[string]string{"checks_ids": `["format"]`})

	report := manifest.Validate(absent, swarmmodel.RunStagePreFlight)
	if !hasRule(report, manifest.RuleEffortUnrecorded) {
		t.Errorf("an absent effort key cited %v, want the unrecorded rule", rules(report))
	}
	if !strings.Contains(report.Error(), "unrecorded") {
		t.Errorf("the finding does not say what to write instead: %s", report.Error())
	}

	// Present as "unrecorded" is the state the rule asks for, and it is
	// clean.
	recorded := doc(t, map[string]string{"effort": `"unrecorded"`})
	for _, f := range manifest.Validate(recorded, swarmmodel.RunStagePreFlight).Failures() {
		if f.Rule == manifest.RuleEffortUnrecorded {
			t.Errorf(`effort:"unrecorded" was reported as absent: %v`, f)
		}
	}

	// A value outside the declared set is a type failure, not an
	// absence: "frantic" is not one of the efforts a run can record.
	wrong := doc(t, map[string]string{"effort": `"frantic"`})
	if !hasRule(manifest.Validate(wrong, swarmmodel.RunStagePreFlight), manifest.RuleType) {
		t.Error("an effort value outside the declared set was accepted")
	}
}

// AC: AC-SWARM-CORE-002.17
//
// A key not written at a stage that does not yet require it is a run
// partway through, not a defect. Reporting it as a failure would make
// every early-stage manifest look broken.
func TestValidationSeparatesAKeyNotYetWrittenFromOneThatIsWrong(t *testing.T) {
	// A record that has only reached pre-flight.
	record := doc(t, map[string]string{
		"checks_ids":             `["format", "vet"]`,
		"checks_baseline":        `"everything passed"`,
		"checks_baseline_status": `0`,
		"concurrency_cap":        `20`,
		"models":                 `{"light":"haiku","standard":"sonnet","advanced":"opus","frontier":"fable"}`,
		"effort":                 `"unrecorded"`,
		"ledger":                 `[]`,
	})

	report := manifest.Validate(record, swarmmodel.RunStagePreFlight)
	if report.Failed() {
		t.Errorf("a complete pre-flight record failed validation: %s", report.Error())
	}

	// The later keys are reported, and reported as pending.
	pending := map[string]bool{}
	for _, f := range report.Findings {
		if f.Rule == manifest.RuleNotYetWritten && f.Severity == manifest.SeverityPending {
			pending[f.Keys[0]] = true
		}
	}
	for _, want := range []string{"branch", "base_sha", "admitted", "findings_path"} {
		if !pending[want] {
			t.Errorf("%q is not reported as not yet written", want)
		}
	}

	// The same record at the dispatch stage: the same keys are now
	// failures, because the stage requires them.
	later := manifest.Validate(record, swarmmodel.RunStageDispatch)
	if !later.Failed() {
		t.Fatal("a record missing its dispatch keys passed at the dispatch stage")
	}
	failed := map[string]bool{}
	for _, f := range later.Failures() {
		if f.Rule == manifest.RuleAbsent {
			failed[f.Keys[0]] = true
		}
	}
	for _, want := range []string{"branch", "base_sha", "admitted"} {
		if !failed[want] {
			t.Errorf("%q is not a failure at the dispatch stage", want)
		}
	}
	// And a key of a LATER stage is still only pending there.
	for _, f := range later.Findings {
		if len(f.Keys) == 1 && f.Keys[0] == "findings_path" && f.Severity != manifest.SeverityPending {
			t.Errorf("findings_path is a %s at the dispatch stage", f.Severity)
		}
	}

	// A key whose absence is legitimate at every stage is not reported at
	// all — a reader told "pending" would wait for something that is
	// never coming.
	for _, f := range later.Findings {
		if len(f.Keys) == 1 && f.Keys[0] == "trailers" {
			t.Errorf("trailers is reported as %s, though its absence is always legitimate", f.Severity)
		}
	}

	// And having no stage does not make the key undeclared: a record that
	// carries it validates clean, and a wrong value for it still fails.
	withTrailers := doc(t, map[string]string{"trailers": `"Co-Authored-By: Someone"`})
	report = manifest.Validate(withTrailers, swarmmodel.RunStagePreFlight)
	for _, f := range report.Failures() {
		if f.Rule == manifest.RuleUndeclared && len(f.Keys) == 1 && f.Keys[0] == "trailers" {
			t.Error("a record carrying trailers reports it as a key nothing declares")
		}
	}
	wrongType := doc(t, map[string]string{"trailers": `3`})
	if !hasRule(manifest.Validate(wrongType, swarmmodel.RunStagePreFlight), manifest.RuleType) {
		t.Error("a numeric trailers value was accepted, though the key takes a string")
	}
}

// AC: AC-SWARM-CORE-002.15
// AC: AC-SWARM-CORE-002.16
//
// Validation reports every rule that fails, not the first — and a record
// that satisfies every rule validates clean, so "reports everything" is
// not satisfied by reporting everything always.
func TestValidationReportsEveryFailingRuleAndACleanRecordPasses(t *testing.T) {
	broken := doc(t, map[string]string{
		"admitted":      `["run-1.1"]`,
		"excluded":      `[{"id": "run-1.1", "reason": "declined"}]`,
		"review_passes": `5`,
		"review_budget": `2`,
		"effort":        `"frantic"`,
		"surprise":      `1`,
	})

	report := manifest.Validate(broken, swarmmodel.RunStageScope)
	found := map[manifest.Rule]bool{}
	for _, r := range rules(report) {
		found[r] = true
	}
	for _, want := range []manifest.Rule{
		manifest.RuleAdmissionOverlap,
		manifest.RuleCounterOverBudget,
		manifest.RuleType,
		manifest.RuleUndeclared,
	} {
		if !found[want] {
			t.Errorf("the report does not cite %s; it cites %v", want, rules(report))
		}
	}

	clean := doc(t, map[string]string{
		"checks_ids":             `["format"]`,
		"checks_baseline":        `"everything passed"`,
		"checks_baseline_status": `0`,
		"concurrency_cap":        `20`,
		"models":                 `{"light":"haiku","standard":"sonnet","advanced":"opus","frontier":"fable"}`,
		"effort":                 `"unrecorded"`,
		"ledger":                 `["a claim — proved by a command"]`,
		"target":                 `"run-1.target"`,
		"admitted":               `["run-1.1"]`,
		"excluded":               `[]`,
		"moved":                  `[]`,
		"standing_auth": `{"pull_request":{"value":"draft","source":"default"},
		  "move_foreign_beads":{"value":false,"source":"default"},
		  "admit_review_findings":{"value":false,"source":"default"},
		  "tier_escalation":{"value":"once","source":"default"},
		  "review_budget":{"value":2,"source":"default"},
		  "merge_to_default":{"value":false,"source":"default"}}`,
		"review_budget": `2`,
	})
	if clean := manifest.Validate(clean, swarmmodel.RunStageScope); clean.Failed() {
		t.Errorf("a record satisfying every rule failed: %s", clean.Error())
	}
}

// AC: AC-SWARM-CORE-002.18
//
// A precondition check names every absent key. A caller told about one
// missing precondition at a time discovers the rest one invocation at a
// time.
func TestAPreconditionCheckNamesEveryMissingKey(t *testing.T) {
	record := doc(t, map[string]string{"branch": `"claude/branch-x"`})

	report := manifest.Require(record, "branch", "base_sha", "target")
	if !report.Failed() {
		t.Fatal("Require passed with two of three keys absent")
	}

	named := map[string]bool{}
	for _, f := range report.Failures() {
		named[f.Keys[0]] = true
	}
	for _, want := range []string{"base_sha", "target"} {
		if !named[want] {
			t.Errorf("%q is not named as missing; the report says %s", want, report.Error())
		}
	}
	if named["branch"] {
		t.Error("a key that IS present was named as missing")
	}

	// Every key present is clean.
	whole := doc(t, map[string]string{"branch": `"x"`, "base_sha": `"y"`, "target": `"z"`})
	if report := manifest.Require(whole, "branch", "base_sha", "target"); report.Failed() {
		t.Errorf("Require failed with every key present: %s", report.Error())
	}

	// A key nothing declares cannot be required, because a caller
	// requiring a misspelt key would otherwise wait forever for it.
	if report := manifest.Require(whole, "brnach"); !hasRule(report, manifest.RuleUndeclared) {
		t.Errorf("requiring an undeclared key cited %v", rules(report))
	}
}

// A stage this vocabulary does not know is refused rather than treated as
// the earliest, which would report every unwritten key as pending.
func TestAStageTheVocabularyDoesNotKnowIsRefused(t *testing.T) {
	report := manifest.Validate(manifest.Document{}, swarmmodel.RunStage("halfway"))
	if !report.Failed() {
		t.Error("an unknown stage validated clean")
	}
}
