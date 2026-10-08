// Tests for the c4 checks that the snapshot cases check-c4-node-defects
// and check-c4-relationship-defects do not express on their own: that
// --only selects each rule alone, and the finding-by-finding claims each
// acceptance criterion makes, asserted against the run's JSON rather
// than against rendered text.
package cmd_test

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// c4CheckFixture copies a snapshot case's input corpus into a fresh
// directory, so these tests and the snapshot read the same diagrams.
func c4CheckFixture(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	// Located from this file rather than the working directory, which a
	// run before this one in the same test has moved.
	_, self, _, _ := runtime.Caller(0)
	cases := filepath.Join(filepath.Dir(self), filepath.FromSlash(casesDir))
	if err := copyTree(filepath.Join(cases, name, "input"), dir); err != nil {
		t.Fatalf("copying the fixture: %v", err)
	}
	return dir
}

const (
	nodeCase         = "check-c4-node-defects"
	relationshipCase = "check-c4-relationship-defects"
)

// c4Run runs check with args over the node fixture and decodes its JSON
// report.
func c4Run(t *testing.T, args ...string) specmodel.FindingReport {
	t.Helper()
	return c4RunCase(t, nodeCase, 1, args...)
}

// c4RunCase runs check --verbose with args over a case's fixture,
// expecting exit status code, and decodes its JSON answer's checks and
// findings.
func c4RunCase(t *testing.T, name string, code int, args ...string) specmodel.FindingReport {
	t.Helper()
	got, stdout, stderr := runInFixture(t, c4CheckFixture(t, name), append([]string{"check", "--json", "--verbose"}, args...))
	if got != code {
		t.Fatalf("exit %d, want %d\n--- stderr ---\n%s", got, code, stderr)
	}
	var answer struct {
		Findings []specmodel.Finding     `json:"findings"`
		Checks   []specmodel.CheckResult `json:"checks"`
	}
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("decoding the report: %v\n%s", err, stdout)
	}
	return specmodel.FindingReport{Checks: answer.Checks, Findings: answer.Findings}
}

// c4Findings returns the report's c4 findings as "rule severity
// line:column label-or-type" lines, in report order.
func c4Findings(report specmodel.FindingReport) []string {
	var out []string
	for _, f := range report.Findings {
		if !strings.HasPrefix(string(f.Rule), "c4/") {
			continue
		}
		out = append(out, f.Path+" "+string(f.Rule)+" "+string(f.Severity)+" "+f.Message)
	}
	return out
}

func contains(lines []string, fragment string) bool {
	for _, line := range lines {
		if strings.Contains(line, fragment) {
			return true
		}
	}
	return false
}

// AC: AC-C4-DIAG-001.1, AC-C4-DIAG-001.3
func TestC4ParseReportsAnUnreadableStatementAndAnUnreadDiagram(t *testing.T) {
	findings := c4Findings(c4Run(t))

	if !contains(findings, `c4/parse error cannot read "Frobnicate(x, \"Not a macro\")"`) {
		t.Errorf("no c4/parse error for the unreadable statement in %q", findings)
	}
	if !contains(findings, `c4/parse unverified diagram type "C4Dynamic"`) {
		t.Errorf("no unverified c4/parse naming C4Dynamic in %q", findings)
	}
}

// AC: AC-C4-DIAG-001.4
func TestADiagramInAFeatureBlueprintIsNotChecked(t *testing.T) {
	findings := c4Findings(c4Run(t))

	if contains(findings, "docs/features/") || contains(findings, "Nowhere Widget") {
		t.Errorf("a feature blueprint's diagram raised c4 findings: %q", findings)
	}
}

// AC: AC-C4-DIAG-002.1, AC-C4-DIAG-002.2, AC-C4-DIAG-002.5
func TestResolvedNodesExternalNodesAndPassiveStoresRaiseNothing(t *testing.T) {
	findings := c4Findings(c4Run(t))

	for _, label := range []string{
		`"Spec CLI"`,    // a container node naming its own blueprint
		`"Parser"`,      // two containers declare a Parser; each resolves in scope
		`"Spec Author"`, // a Person node
		`"CI Pipeline"`, // a System_Ext node
		`"Spec Corpus"`, // a ContainerDb node naming nothing: a passive store
	} {
		if contains(findings, label) {
			t.Errorf("a finding names %s: %q", label, findings)
		}
	}
}

// AC: AC-C4-DIAG-002.3, AC-C4-DIAG-002.4
func TestUndeclaredAndAmbiguousNodesAreErrors(t *testing.T) {
	report := c4Run(t)

	var undeclared, ambiguous *specmodel.Finding
	for i, f := range report.Findings {
		switch f.Rule {
		case "c4/undeclared-node":
			undeclared = &report.Findings[i]
		case "c4/ambiguous-node":
			ambiguous = &report.Findings[i]
		}
	}
	if undeclared == nil || undeclared.Severity != "error" || undeclared.Span.StartLine != 23 {
		t.Errorf("undeclared-node finding = %+v, want an error at line 23", undeclared)
	}
	if ambiguous == nil || ambiguous.Severity != "error" || ambiguous.Span.StartLine != 22 {
		t.Fatalf("ambiguous-node finding = %+v, want an error at line 22", ambiguous)
	}
	node, ok := ambiguous.Detail.AsC4Node()
	if !ok || strings.Join(node.Matches, ",") != "container-twin-one,container-twin-two" {
		t.Errorf("ambiguous-node detail = %+v, want both Twins named", ambiguous.Detail)
	}
}

// --only names each c4 rule alone: that rule is requested and failed,
// every other rule is not requested, and only its findings appear.
func TestOnlySelectsEachC4Rule(t *testing.T) {
	cases := map[string]string{
		"c4/parse":                nodeCase,
		"c4/undeclared-node":      nodeCase,
		"c4/ambiguous-node":       nodeCase,
		"c4/bidirectional":        relationshipCase,
		"c4/duplicate-edge":       relationshipCase,
		"c4/drawing-conflict":     relationshipCase,
		"c4/interaction-conflict": relationshipCase,
		"c4/contract-unresolved":  relationshipCase,
		"c4/contract-mismatch":    relationshipCase,
		"c4/tag-separator":        relationshipCase,
	}
	for rule, name := range cases {
		t.Run(rule, func(t *testing.T) {
			code := 1
			if severity := c4RuleSeverity[rule]; severity == "warning" {
				code = 0
			}
			report := c4RunCase(t, name, code, "--only", rule)

			for _, check := range report.Checks {
				want := specmodel.CheckOutcomeNotRequested
				if string(check.Rule) == rule {
					want = specmodel.CheckOutcomeFailed
				}
				if check.Outcome != want {
					t.Errorf("%s = %s, want %s", check.Rule, check.Outcome, want)
				}
			}
			if len(report.Findings) == 0 {
				t.Fatal("no finding")
			}
			for _, f := range report.Findings {
				if string(f.Rule) != rule {
					t.Errorf("a %s finding under --only %s", f.Rule, rule)
				}
			}
		})
	}
}

// c4RuleSeverity is the severity each relationship rule reports at; a
// warning-only run exits zero without --strict.
var c4RuleSeverity = map[string]string{
	"c4/drawing-conflict":  "warning",
	"c4/contract-mismatch": "warning",
	"c4/tag-separator":     "warning",
}

// relationshipFinding returns the one finding of rule in the relationship
// fixture's full run.
func relationshipFinding(t *testing.T, report specmodel.FindingReport, rule string) specmodel.Finding {
	t.Helper()
	var found []specmodel.Finding
	for _, f := range report.Findings {
		if string(f.Rule) == rule {
			found = append(found, f)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d %s findings, want 1: %+v", len(found), rule, found)
	}
	return found[0]
}

// at asserts a finding's file, line and severity.
func at(t *testing.T, f specmodel.Finding, path string, line uint64, severity string) {
	t.Helper()
	if f.Path != path || f.Span.StartLine != line || string(f.Severity) != severity {
		t.Errorf("%s at %s:%d %s, want %s:%d %s",
			f.Rule, f.Path, f.Span.StartLine, f.Severity, path, line, severity)
	}
}

// AC: AC-C4-DIAG-003.1
func TestABidirectionalArrowIsAnError(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	at(t, relationshipFinding(t, report, "c4/bidirectional"), "docs/containers/alpha.md", 24, "error")
}

// The second a -> b statement — written Rel_Back(b, a) — is the
// duplicate; Rel(b, a) is a different pair and is not.
//
// AC: AC-C4-DIAG-003.2
func TestAnEdgeDrawnTwiceInOneDiagramIsAnError(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	f := relationshipFinding(t, report, "c4/duplicate-edge")
	at(t, f, "docs/containers/alpha.md", 26, "error")
	if detail, _ := f.Detail.AsC4RelationshipDefect(); detail.Source != "a" || detail.Target != "b" {
		t.Errorf("detail = %+v, want a -> b", detail)
	}
}

// AC: AC-C4-DIAG-003.3
func TestDrawingsThatDisagreeAreAWarningOnTheSecond(t *testing.T) {
	for _, args := range [][]string{nil, {"docs/containers/beta.md"}} {
		code := 1
		if args != nil {
			code = 0 // the named file holds only the warning
		}
		report := c4RunCase(t, relationshipCase, code, args...)
		f := relationshipFinding(t, report, "c4/drawing-conflict")
		at(t, f, "docs/containers/beta.md", 20, "warning")
		detail, _ := f.Detail.AsC4DrawingConflict()
		if detail.OtherPath != "docs/containers/alpha.md" || detail.OtherSpan.StartLine != 25 {
			t.Errorf("args %v: detail = %+v, want alpha.md:25 named", args, detail)
		}
	}
}

// AC: AC-C4-DIAG-004.2
func TestContradictoryInteractionTagsAreAnError(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	at(t, relationshipFinding(t, report, "c4/interaction-conflict"), "docs/containers/alpha.md", 28, "error")
}

// AC: AC-C4-DIAG-004.3
func TestAContractTagNamingNoContractIsAnError(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	f := relationshipFinding(t, report, "c4/contract-unresolved")
	at(t, f, "docs/containers/alpha.md", 29, "error")
	if detail, _ := f.Detail.AsC4Contract(); detail.Slug != "contract-missing" {
		t.Errorf("detail = %+v", detail)
	}
}

// Rel(b, g) carries contract-good, which Beta does not consume and Gamma
// does not provide. Rel(a, b) carries it too, and Alpha consumes it and
// Beta provides it, so that one raises nothing.
//
// AC: AC-C4-DIAG-004.4
func TestAContractTagTheEndpointsDoNotDeclareIsAWarning(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	f := relationshipFinding(t, report, "c4/contract-mismatch")
	at(t, f, "docs/containers/alpha.md", 30, "warning")
	if detail, _ := f.Detail.AsC4Contract(); detail.Slug != "contract-good" || detail.Source != "b" || detail.Target != "g" {
		t.Errorf("detail = %+v", detail)
	}
}

// AC: AC-C4-DIAG-004.5
func TestTagsJoinedWithACommaAreAWarning(t *testing.T) {
	report := c4RunCase(t, relationshipCase, 1)
	f := relationshipFinding(t, report, "c4/tag-separator")
	at(t, f, "docs/containers/alpha.md", 31, "warning")
	if detail, _ := f.Detail.AsC4TagSeparator(); detail.Tags != "async,audited" {
		t.Errorf("detail = %+v", detail)
	}
}
