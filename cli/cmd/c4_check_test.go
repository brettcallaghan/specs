// Tests for the c4 node checks that the snapshot case
// check-c4-node-defects does not express on its own: that --only selects
// each of the three rules alone, and the finding-by-finding claims each
// acceptance criterion makes, asserted against the run's JSON rather
// than against rendered text.
package cmd_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// c4CheckFixture copies the check-c4-node-defects case's input corpus
// into a fresh directory, so these tests and the snapshot read the same
// diagrams.
func c4CheckFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := copyTree(filepath.Join(casesDir, "check-c4-node-defects", "input"), dir); err != nil {
		t.Fatalf("copying the fixture: %v", err)
	}
	return dir
}

// c4Run runs check with args over the fixture and decodes its JSON report.
func c4Run(t *testing.T, args ...string) specmodel.FindingReport {
	t.Helper()
	code, stdout, stderr := runInFixture(t, c4CheckFixture(t), append([]string{"check", "--json"}, args...))
	if code != 1 {
		t.Fatalf("exit %d, want 1\n--- stderr ---\n%s", code, stderr)
	}
	var report specmodel.FindingReport
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("decoding the report: %v\n%s", err, stdout)
	}
	return report
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
// every other c4 rule is not requested, and only its findings appear.
func TestOnlySelectsEachC4NodeRule(t *testing.T) {
	rules := []string{"c4/parse", "c4/undeclared-node", "c4/ambiguous-node"}
	for _, rule := range rules {
		t.Run(rule, func(t *testing.T) {
			report := c4Run(t, "--only", rule)

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
