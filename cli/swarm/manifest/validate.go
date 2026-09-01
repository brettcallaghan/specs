package manifest

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// The rules below are the ones no single write can check on its own: a
// branch written twice, a review counter that went backwards, a ledger
// whose entries were reordered under citations that address them by
// position, a bead in two admission lists at once. Every one of them is
// stated here rather than as prose in five skills.

// Severity is what a finding amounts to.
type Severity string

const (
	// SeverityPending — a key is not written yet, at a stage that does
	// not require it. Reported, because a reader has to be able to tell
	// "not yet" from "never", and not a failure.
	SeverityPending Severity = "pending"
	// SeverityFailure — a rule the record breaks.
	SeverityFailure Severity = "failure"
)

// Rule identifies what was checked, so a caller can act on a finding
// without reading its message.
type Rule string

const (
	// RuleAbsent — a key required by this stage is not written.
	RuleAbsent Rule = "manifest/absent"
	// RuleNotYetWritten — a key not written, at a stage before its own.
	RuleNotYetWritten Rule = "manifest/not-yet-written"
	// RuleType — a key's value is not of its declared type.
	RuleType Rule = "manifest/type"
	// RuleUndeclared — a key outside the declared set.
	RuleUndeclared Rule = "manifest/undeclared"
	// RuleWriteOnce — a key that may be written once was written again
	// with a different value.
	RuleWriteOnce Rule = "manifest/write-once"
	// RuleCounterBackwards — the review counter went down.
	RuleCounterBackwards Rule = "manifest/counter-backwards"
	// RuleCounterOverBudget — the review counter passed its budget.
	RuleCounterOverBudget Rule = "manifest/counter-over-budget"
	// RuleLedgerRewritten — a ledger entry was removed or reordered.
	RuleLedgerRewritten Rule = "manifest/ledger-rewritten"
	// RuleAdmissionOverlap — a bead is in more than one admission list.
	RuleAdmissionOverlap Rule = "manifest/admission-overlap"
	// RuleEffortUnrecorded — the effort key is absent where it should be
	// present as unrecorded.
	RuleEffortUnrecorded Rule = "manifest/effort-unrecorded"
)

// Finding is one thing validation found.
type Finding struct {
	Rule     Rule
	Severity Severity
	// Keys are what the finding concerns — one key for most rules, a
	// bead id for the admission overlap.
	Keys []string
	// Message says what is wrong, in the run's own words.
	Message string
}

// Report is everything validation found, in rule order.
type Report struct {
	Findings []Finding
}

// Failed reports whether any finding is a failure. A report of pending
// keys and nothing else is a run partway through, not a broken record.
func (r Report) Failed() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityFailure {
			return true
		}
	}
	return false
}

// Failures returns only the findings that are failures.
func (r Report) Failures() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Severity == SeverityFailure {
			out = append(out, f)
		}
	}
	return out
}

// Error renders the report as one line naming every failure, empty when
// there are none.
func (r Report) Error() string {
	failures := r.Failures()
	if len(failures) == 0 {
		return ""
	}
	messages := make([]string, 0, len(failures))
	for _, f := range failures {
		messages = append(messages, f.Message)
	}
	return strings.Join(messages, "; ")
}

func (r *Report) add(rule Rule, severity Severity, keys []string, format string, a ...any) {
	r.Findings = append(r.Findings, Finding{
		Rule: rule, Severity: severity, Keys: keys,
		Message: fmt.Sprintf(format, a...),
	})
}

// stageOrder is each stage's position in a run.
var stageOrder = func() map[swarmmodel.RunStage]int {
	out := map[swarmmodel.RunStage]int{}
	for i, stage := range contract.Stages() {
		out[stage] = i
	}
	return out
}()

// Validate checks a record as it stands, at a stage a run has reached.
//
// It reports EVERY rule that fails rather than the first, because a
// caller that fixed one and re-ran to find the next would be doing the
// work this is for. And it separates a key that is absent at a stage
// which does not yet require it — reported as pending — from one that is
// wrong, because the first is a run partway through and the second is a
// defect.
func Validate(doc Document, stage swarmmodel.RunStage) Report {
	var report Report

	reached, known := stageOrder[stage]
	if !known {
		report.add(RuleUndeclared, SeverityFailure, []string{string(stage)},
			"%q is not a stage of a run", stage)
		return report
	}

	for _, key := range Keys() {
		value, present := doc.Get(key.Name)
		switch {
		case present:
			if err := checkType(key, value); err != nil {
				report.add(RuleType, SeverityFailure, []string{key.Name}, "%s", err)
			}
		case key.Stage == "":
			// A key that becomes required at no stage. Absent by design
			// at every one of them, so not reported at all: a reader told
			// "pending" would wait for something that is never coming.
		case stageOrder[key.Stage] <= reached:
			report.add(RuleAbsent, SeverityFailure, []string{key.Name},
				"%q is required by the %s stage and is not written", key.Name, stage)
		default:
			report.add(RuleNotYetWritten, SeverityPending, []string{key.Name},
				"%q is not written yet; it becomes required at the %s stage", key.Name, key.Stage)
		}
	}

	// A key the record carries that nothing declares.
	for _, name := range doc.Names() {
		if _, ok := Lookup(name); !ok {
			report.add(RuleUndeclared, SeverityFailure, []string{name},
				"%q is not a declared manifest key", name)
		}
	}

	validateEffort(doc, reached, &report)
	validateCounter(doc, &report)
	validateAdmission(doc, &report)

	return report
}

func checkType(key Key, value json.RawMessage) error {
	target := newValue(key)
	if err := json.Unmarshal(value, target); err != nil {
		return fmt.Errorf("%q takes %s, and the value stored is not one: %v", key.Name, key.Type, err)
	}
	return nil
}

// validateEffort: a harness that exposes no effort setting records the
// key as unrecorded rather than leaving it absent, so a reader can tell
// "not offered" from "not yet written". An absent key past the stage that
// establishes it is therefore a failure with its own rule, saying what to
// write rather than only that something is missing.
func validateEffort(doc Document, reached int, report *Report) {
	key, ok := Lookup("effort")
	if !ok {
		return
	}
	if _, present := doc.Get("effort"); present {
		return
	}
	if stageOrder[key.Stage] > reached {
		return
	}
	// Replace the generic absence finding with one that says what to do.
	for i, f := range report.Findings {
		if f.Rule == RuleAbsent && len(f.Keys) == 1 && f.Keys[0] == "effort" {
			report.Findings[i] = Finding{
				Rule: RuleEffortUnrecorded, Severity: SeverityFailure, Keys: []string{"effort"},
				Message: `"effort" is absent; a harness that exposes no effort setting records it as "unrecorded", so that a reader can tell "not offered" from "not yet written"`,
			}
			return
		}
	}
}

// validateCounter: the review counter never exceeds the budget recorded
// beside it. A resumed run that spent its budget must not be handed a
// further pass.
func validateCounter(doc Document, report *Report) {
	passes, okPasses := readInt(doc, "review_passes")
	budget, okBudget := readInt(doc, "review_budget")
	if !okPasses || !okBudget {
		return
	}
	if passes > budget {
		report.add(RuleCounterOverBudget, SeverityFailure, []string{"review_passes"},
			`"review_passes" is %d against a budget of %d`, passes, budget)
	}
}

// validateAdmission: the three admission lists do not overlap. A bead in
// two of them makes the run's own account of its scope self-contradictory,
// and the finding names the bead rather than the lists.
func validateAdmission(doc Document, report *Report) {
	lists := map[string][]string{}

	var admitted []string
	if present, err := doc.Decode("admitted", &admitted); present && err == nil {
		lists["admitted"] = admitted
	}
	var excluded []swarmmodel.Exclusion
	if present, err := doc.Decode("excluded", &excluded); present && err == nil {
		for _, e := range excluded {
			lists["excluded"] = append(lists["excluded"], e.Id)
		}
	}
	var moved []swarmmodel.Move
	if present, err := doc.Decode("moved", &moved); present && err == nil {
		for _, m := range moved {
			lists["moved"] = append(lists["moved"], m.Id)
		}
	}

	in := map[string][]string{}
	for _, name := range []string{"admitted", "excluded", "moved"} {
		for _, id := range lists[name] {
			in[id] = append(in[id], name)
		}
	}

	var overlapping []string
	for id, names := range in {
		if len(names) > 1 {
			overlapping = append(overlapping, id)
		}
	}
	sort.Strings(overlapping)

	for _, id := range overlapping {
		report.add(RuleAdmissionOverlap, SeverityFailure, []string{id},
			"%s is in more than one of the admitted, excluded and moved lists (%s)",
			id, strings.Join(in[id], ", "))
	}
}

// CheckWrite reports every reason a proposed change may not be applied to
// the record as it stands.
//
// It covers what [Check] does — a key outside the declared set, a value
// of the wrong type — and the rules that can only be seen by comparing
// the change to what is already there.
func CheckWrite(current Document, changes map[string]json.RawMessage) Report {
	var report Report

	names := make([]string, 0, len(changes))
	for name := range changes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		key, ok := Lookup(name)
		if !ok {
			report.add(RuleUndeclared, SeverityFailure, []string{name},
				"%q is not a declared manifest key", name)
			continue
		}
		if err := checkType(key, changes[name]); err != nil {
			report.add(RuleType, SeverityFailure, []string{name}, "%s", err)
			continue
		}
		checkWriteOnce(current, name, changes[name], &report)
		checkCounter(current, name, changes[name], &report)
		checkLedger(current, name, changes[name], &report)
	}

	// The admission lists as they would stand after the write.
	if touchesAdmission(changes) {
		validateAdmission(current.Merged(changes), &report)
	}

	return report
}

// writeOnce names the keys a run writes once and never again. The branch
// and the commit it was cut from are the run's identity: three things
// read them and none may change them, because a run that repointed either
// mid-flight would leave its review scoped to one tree and its worktrees
// branched from another.
var writeOnce = map[string]bool{"branch": true, "base_sha": true}

func checkWriteOnce(current Document, name string, value json.RawMessage, report *Report) {
	if !writeOnce[name] {
		return
	}
	existing, present := current.Get(name)
	if !present || sameJSON(existing, value) {
		return
	}
	report.add(RuleWriteOnce, SeverityFailure, []string{name},
		"%q is written once and already reads %s; refusing to change it to %s",
		name, existing, value)
}

func checkCounter(current Document, name string, value json.RawMessage, report *Report) {
	if name != "review_passes" {
		return
	}

	var next int
	if err := json.Unmarshal(value, &next); err != nil {
		return
	}
	if spent, ok := readInt(current, "review_passes"); ok && next < spent {
		report.add(RuleCounterBackwards, SeverityFailure, []string{name},
			`"review_passes" already reads %d; refusing to set it to %d, because a resumed run that restarted the counter would loop forever`,
			spent, next)
	}
	if budget, ok := readInt(current, "review_budget"); ok && next > budget {
		report.add(RuleCounterOverBudget, SeverityFailure, []string{name},
			`"review_passes" would be %d against a budget of %d`, next, budget)
	}
}

// checkLedger: the environmental ledger is append-only. Dispatch prompts
// cite entries by position, so an entry removed or reordered retargets
// every citation that addressed it — and nothing downstream can tell.
func checkLedger(current Document, name string, value json.RawMessage, report *Report) {
	if name != "ledger" {
		return
	}

	var next []string
	if err := json.Unmarshal(value, &next); err != nil {
		return
	}
	var existing []string
	present, err := current.Decode("ledger", &existing)
	if !present || err != nil {
		return
	}

	if len(next) < len(existing) {
		report.add(RuleLedgerRewritten, SeverityFailure, []string{name},
			"the ledger is append-only: the write drops %d of its %d entries",
			len(existing)-len(next), len(existing))
		return
	}
	for i, entry := range existing {
		if next[i] != entry {
			report.add(RuleLedgerRewritten, SeverityFailure, []string{name},
				"the ledger is append-only: entry %d would change, and dispatch prompts cite entries by position",
				i+1)
			return
		}
	}
}

func touchesAdmission(changes map[string]json.RawMessage) bool {
	for _, name := range []string{"admitted", "excluded", "moved"} {
		if _, ok := changes[name]; ok {
			return true
		}
	}
	return false
}

// Require reports every one of the named keys that is not set to a valid
// value.
//
// "Valid", not merely "present": a caller saying "I need this key now" is
// asking whether it can USE the value, and a key written with a value of
// the wrong type fails that as surely as an unwritten one. The two are
// still distinguishable — they carry different rules — so a caller that
// cares which can tell.
//
// Every failing key, not the first: a caller told about one missing
// precondition at a time discovers the rest one invocation at a time.
//
// It overrides optionality, which is what separates it from [Validate]:
// a key belonging to no stage the record has reached is never required by
// Validate and is always required here.
func Require(doc Document, keys ...string) Report {
	var report Report

	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)

	for _, name := range sorted {
		key, ok := Lookup(name)
		if !ok {
			report.add(RuleUndeclared, SeverityFailure, []string{name},
				"%q is not a declared manifest key", name)
			continue
		}
		value, present := doc.Get(name)
		if !present {
			report.add(RuleAbsent, SeverityFailure, []string{name},
				"%q is required and is not written", name)
			continue
		}
		if err := checkType(key, value); err != nil {
			report.add(RuleType, SeverityFailure, []string{name}, "%s", err)
		}
	}
	return report
}

func readInt(doc Document, key string) (int, bool) {
	var n int
	present, err := doc.Decode(key, &n)
	if !present || err != nil {
		return 0, false
	}
	return n, true
}
