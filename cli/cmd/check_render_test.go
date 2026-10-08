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

// AC: AC-SPEC-006.5
func TestCheckRendersTheSameValuesAsYAMLAndAsJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"check-error", []string{"check"}},
		{"check-only", []string{"check", "--only", "slug/pattern-mismatch"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yamlStatus, yamlOut := runIn(t, tc.name, tc.args...)
			jsonStatus, jsonOut := runIn(t, tc.name, append(tc.args, "--json")...)
			if yamlStatus != jsonStatus {
				t.Errorf("status = %d as YAML and %d as JSON", yamlStatus, jsonStatus)
			}
			if got, want := topLevelKeys(yamlOut), []string{"checks", "findings"}; !reflect.DeepEqual(got, want) {
				t.Errorf("YAML top-level keys = %v, want %v", got, want)
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
	_, out := runIn(t, "check-only", "check", "--json", "--only", "slug/pattern-mismatch")
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
