// The check command's two renderings carry the same values in the same
// order, per [[cli-output-standards]]: YAML by default, JSON under --json,
// both the generated finding report.
package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// runIn runs args against a copy of a snapshot case's input corpus and
// returns its status and standard output.
func runIn(t *testing.T, caseName string, args ...string) (int, string) {
	t.Helper()
	work := t.TempDir()
	if err := copyTree(filepath.Join(casesDir, caseName, "input"), work); err != nil {
		t.Fatalf("copying the input corpus: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	var stdout, stderr bytes.Buffer
	status := cmd.Execute(args, &stdout, &stderr)
	return status, stdout.String()
}

// generic decodes a rendering into plain values, normalized through JSON
// so that a YAML integer and a JSON number compare equal.
func generic(t *testing.T, decode func([]byte, any) error, data string) any {
	t.Helper()
	var v any
	if err := decode([]byte(data), &v); err != nil {
		t.Fatalf("decoding %q: %v", data, err)
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("re-encoding: %v", err)
	}
	var normalized any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		t.Fatalf("re-decoding: %v", err)
	}
	return normalized
}

// topLevelKeys lists the YAML document's unindented keys in order.
func topLevelKeys(doc string) []string {
	var keys []string
	for _, line := range strings.Split(doc, "\n") {
		if line == "" || line[0] == ' ' || line[0] == '-' {
			continue
		}
		key, _, _ := strings.Cut(line, ":")
		keys = append(keys, key)
	}
	return keys
}

// AC: AC-SPEC-006.5, AC-SPEC-006.11, AC-SPEC-006.12
func TestCheckRendersTheSameValuesAsYAMLAndAsJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		keys []string
	}{
		{"check-error", []string{"check"}, []string{"summary", "findings"}},
		{"check-only", []string{"check", "--only", "slug/pattern-mismatch"}, []string{"summary", "findings"}},
		{"check-error", []string{"check", "--verbose"}, []string{"summary", "findings", "checks"}},
		{"check-only", []string{"check", "--verbose", "--only", "slug/pattern-mismatch"}, []string{"summary", "findings", "checks"}},
	} {
		t.Run(strings.Join(append([]string{tc.name}, tc.args[1:]...), " "), func(t *testing.T) {
			yamlStatus, yamlOut := runIn(t, tc.name, tc.args...)
			jsonStatus, jsonOut := runIn(t, tc.name, append(tc.args, "--json")...)
			if yamlStatus != jsonStatus {
				t.Errorf("status = %d as YAML and %d as JSON", yamlStatus, jsonStatus)
			}
			if got := topLevelKeys(yamlOut); !reflect.DeepEqual(got, tc.keys) {
				t.Errorf("YAML top-level keys = %v, want %v", got, tc.keys)
			}
			if got := jsonKeys(t, jsonOut); !reflect.DeepEqual(got, tc.keys) {
				t.Errorf("JSON top-level keys = %v, want %v", got, tc.keys)
			}
			fromYAML := generic(t, yaml.Unmarshal, yamlOut)
			fromJSON := generic(t, json.Unmarshal, jsonOut)
			if !reflect.DeepEqual(fromYAML, fromJSON) {
				t.Errorf("the renderings differ.\n--- YAML ---\n%s\n--- JSON ---\n%s", yamlOut, jsonOut)
			}
		})
	}
}

// AC: AC-SPEC-006.8, AC-SPEC-006.9
func TestCheckOnlyListsExactlyTheRequestedCheck(t *testing.T) {
	_, out := runIn(t, "check-only", "check", "--json", "--verbose", "--only", "slug/pattern-mismatch")
	var report struct {
		Checks []struct {
			Rule    string `json:"rule"`
			Outcome string `json:"outcome"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("decoding: %v\n%s", err, out)
	}
	if len(report.Checks) != 1 || report.Checks[0].Rule != "slug/pattern-mismatch" {
		t.Errorf("checks = %+v, want exactly slug/pattern-mismatch", report.Checks)
	}
	if strings.Contains(out, "not-requested") {
		t.Errorf("a not-requested outcome appears in the answer:\n%s", out)
	}
}

// jsonKeys lists a JSON object's top-level keys in document order.
func jsonKeys(t *testing.T, doc string) []string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(doc))
	if _, err := dec.Token(); err != nil {
		t.Fatalf("decoding %q: %v", doc, err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("decoding %q: %v", doc, err)
		}
		keys = append(keys, tok.(string))
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			t.Fatalf("decoding %q: %v", doc, err)
		}
	}
	return keys
}

// checkSummaryAnswer is the part of check's JSON answer the summary
// tests read.
type checkSummaryAnswer struct {
	Summary  map[string]int `json:"summary"`
	Findings []struct {
		Severity string `json:"severity"`
	} `json:"findings"`
	Checks *json.RawMessage `json:"checks"`
}

// AC: AC-SPEC-006.11
func TestCheckSummaryCountsTheFindingsItReports(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"check-error", []string{"check", "--json"}},
		{"check-strict-warning-only", []string{"check", "--json"}},
		{"check-only", []string{"check", "--json", "--only", "slug/pattern-mismatch"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, out := runIn(t, tc.name, tc.args...)
			var answer checkSummaryAnswer
			if err := json.Unmarshal([]byte(out), &answer); err != nil {
				t.Fatalf("decoding: %v\n%s", err, out)
			}
			want := map[string]int{"errors": 0, "warnings": 0, "unverified": 0}
			for _, f := range answer.Findings {
				switch f.Severity {
				case "error":
					want["errors"]++
				case "warning":
					want["warnings"]++
				case "unverified":
					want["unverified"]++
				default:
					t.Errorf("unexpected severity %q", f.Severity)
				}
			}
			if !reflect.DeepEqual(answer.Summary, want) {
				t.Errorf("summary = %v, want %v (counted from the findings)", answer.Summary, want)
			}
		})
	}
}

// AC: AC-SPEC-006.12, AC-SPEC-006.13
func TestCheckListsChecksOnlyUnderVerboseAndKeepsItsStatus(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"check-error", []string{"check", "--json"}},
		{"check-strict-warning-only", []string{"check", "--json"}},
		{"check-strict-warning-only", []string{"check", "--json", "--strict"}},
	} {
		t.Run(strings.Join(append([]string{tc.name}, tc.args[1:]...), " "), func(t *testing.T) {
			plainStatus, plain := runIn(t, tc.name, tc.args...)
			verboseStatus, verbose := runIn(t, tc.name, append(tc.args, "--verbose")...)
			if plainStatus != verboseStatus {
				t.Errorf("status = %d without --verbose and %d with it", plainStatus, verboseStatus)
			}
			var a, b checkSummaryAnswer
			if err := json.Unmarshal([]byte(plain), &a); err != nil {
				t.Fatalf("decoding: %v\n%s", err, plain)
			}
			if err := json.Unmarshal([]byte(verbose), &b); err != nil {
				t.Fatalf("decoding: %v\n%s", err, verbose)
			}
			if a.Checks != nil {
				t.Errorf("checks present without --verbose:\n%s", plain)
			}
			if b.Checks == nil {
				t.Errorf("checks absent under --verbose:\n%s", verbose)
			}
		})
	}
}
