// Whitebox tests for exitCodeFor and buildCheckReport: the exit-status
// derivation rule (AC-SPEC-006.6, AC-SPEC-006.7) and the --only
// accounting (AC-SPEC-006.8) that back the check subcommand.
//
// These run against a synthetic [finding.Report] rather than a real
// corpus, because no check registered in this sprint raises severity
// warning or unverified (cli/core/finding's own doc comment says so) —
// there is no corpus fixture that could exercise AC-SPEC-006.7
// end-to-end today. Testing the derivation rule directly is what proves
// the CLI already honors it correctly once such a check exists, rather
// than leaving it untested until one does.
package cmd

import (
	"testing"

	"github.com/brettcallaghan/specs/cli/core/finding"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

func errorFinding(rule specmodel.RuleId) specmodel.Finding {
	return specmodel.Finding{Rule: rule, Severity: finding.SeverityError, Path: "doc.md", Message: "boom"}
}

func warningFinding(rule specmodel.RuleId) specmodel.Finding {
	return specmodel.Finding{Rule: rule, Severity: finding.SeverityWarning, Path: "doc.md", Message: "hmm"}
}

func unverifiedFinding(rule specmodel.RuleId) specmodel.Finding {
	return specmodel.Finding{Rule: rule, Severity: finding.SeverityUnverified, Path: "doc.md", Message: "could not run"}
}

// AC: AC-SPEC-006.6
func TestExitCodeForIsNonZeroWhenAnyFindingIsAnError(t *testing.T) {
	report := finding.NewReport(finding.RuleSchemaViolation)
	report.Add(errorFinding(finding.RuleSchemaViolation))

	if got := exitCodeFor(report, false); got != 1 {
		t.Errorf("exitCodeFor(report, strict=false) = %d, want 1 for a run with an error finding", got)
	}
	if got := exitCodeFor(report, true); got != 1 {
		t.Errorf("exitCodeFor(report, strict=true) = %d, want 1: an error exits non-zero regardless of --strict", got)
	}
}

// AC: AC-SPEC-006.7
func TestExitCodeForWarningOrUnverifiedOnlyExitsZeroUnlessStrict(t *testing.T) {
	warningReport := finding.NewReport(finding.RuleSlugPattern)
	warningReport.Add(warningFinding(finding.RuleSlugPattern))

	unverifiedReport := finding.NewReport(finding.RuleSlugPattern)
	unverifiedReport.Add(unverifiedFinding(finding.RuleSlugPattern))

	for name, report := range map[string]*finding.Report{"warning": warningReport, "unverified": unverifiedReport} {
		if got := exitCodeFor(report, false); got != 0 {
			t.Errorf("%s: exitCodeFor(report, strict=false) = %d, want 0", name, got)
		}
		if got := exitCodeFor(report, true); got != 1 {
			t.Errorf("%s: exitCodeFor(report, strict=true) = %d, want 1", name, got)
		}
	}
}

func TestExitCodeForCleanReportIsAlwaysZero(t *testing.T) {
	report := finding.NewReport(finding.RuleSchemaViolation)
	report.Ran(finding.RuleSchemaViolation)

	if got := exitCodeFor(report, false); got != 0 {
		t.Errorf("exitCodeFor(clean, strict=false) = %d, want 0", got)
	}
	if got := exitCodeFor(report, true); got != 0 {
		t.Errorf("exitCodeFor(clean, strict=true) = %d, want 0: --strict has nothing to escalate", got)
	}
}

// AC: AC-SPEC-006.8, AC-SPEC-006.9
func TestBuildCheckReportOnlyRunsNamedChecks(t *testing.T) {
	findings := []specmodel.Finding{
		errorFinding(finding.RuleSlugPattern),
		errorFinding(finding.RuleMissingProperty),
	}

	report := buildCheckReport(findings, []string{string(finding.RuleSlugPattern)})
	model := report.Model()

	outcomes := map[specmodel.RuleId]specmodel.CheckOutcome{}
	for _, c := range model.Checks {
		outcomes[c.Rule] = c.Outcome
	}

	if got := outcomes[finding.RuleSlugPattern]; got != specmodel.CheckOutcomeFailed {
		t.Errorf("named check %q outcome = %q, want %q", finding.RuleSlugPattern, got, specmodel.CheckOutcomeFailed)
	}
	if got, listed := outcomes[finding.RuleMissingProperty]; listed {
		t.Errorf("un-named check %q is listed with outcome %q; a check the run was not asked for is absent from checks",
			finding.RuleMissingProperty, got)
	}
	if len(model.Checks) != 1 {
		t.Errorf("got %d checks, want exactly the 1 requested: %+v", len(model.Checks), model.Checks)
	}

	var sawUnrequestedFinding bool
	for _, f := range model.Findings {
		if f.Rule == finding.RuleMissingProperty {
			sawUnrequestedFinding = true
		}
	}
	if sawUnrequestedFinding {
		t.Error("a finding for a check that was not requested by --only was still reported")
	}
}

func TestBuildCheckReportWithNoOnlyRunsEveryCheck(t *testing.T) {
	report := buildCheckReport(nil, nil)
	model := report.Model()

	if len(model.Checks) != len(checkRules) {
		t.Fatalf("got %d checks, want %d (every known rule)", len(model.Checks), len(checkRules))
	}
	for _, c := range model.Checks {
		if c.Outcome != specmodel.CheckOutcomePassed {
			t.Errorf("rule %q outcome = %q, want %q for a clean run with no --only", c.Rule, c.Outcome, specmodel.CheckOutcomePassed)
		}
	}
}
