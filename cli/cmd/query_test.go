// Tests for `feature` and `criterion`'s query surface that the snapshot
// cases under cli/testdata/cases/ cannot express: the harness in
// cli/cmd/snapshot_test.go can only compare what one run produces, and
// AC-GHK-007.6 is a claim about a *second* run reading the persisted
// index rather than re-walking the corpus — a case cannot stage "nothing
// was re-walked", only a Go test can plant a record that exists solely
// in the persisted index and prove it was served.
package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// writeQueryFixture materializes a minimal corpus carrying one feature,
// FEAT-QRY-001 with one criterion AC-QRY-001.1, under dir.
func writeQueryFixture(t *testing.T, dir string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(dir, "specs.toml"), `[workspace]
name = "acme"
root = "docs"

[profile]
name = "specs"
`)
	mustWriteFile(t, filepath.Join(dir, "docs", "features", "frd-warm.md"), `---
title: "Warm"
description: A feature requirements document used by a warm-index test.
slug: frd-warm
type: frd
status: approved
updated_at: 2026-08-01T00:00:00Z
---

# Warm

`+"```gherkin"+`
@FEAT-QRY-001
Feature: Warm lookups

  As an agent
  I want a query answered from the persisted index
  So that it costs nothing beyond the index read

  @AC-QRY-001.1
  Scenario: A lookup is served warm
    Given a warm index
    When a query runs
    Then it is served from the index
`+"```"+`
`)
}

// plantedFeature is a feature record present in nowhere but the persisted
// index — [dir]'s corpus never carries FEAT-PLANTED-999 on disk. A
// command that answers it can only have read the persisted index; a
// command that re-walked the corpus would report it not found, exactly
// as [runFeatureQuery] and [runCriterionQuery] do for any other unknown
// identifier.
func plantedFeature() specmodel.Feature {
	return specmodel.Feature{
		Id:           "FEAT-PLANTED-999",
		Code:         "PLANTED",
		DocumentSlug: "frd-warm",
		Path:         "docs/features/frd-warm.md",
		Name:         "Planted only in the index",
		Source:       specmodel.NewFeatureSourceFeatureFile(specmodel.FeatureSourceFile{}),
		Criteria: []specmodel.Criterion{
			{
				Id:          "AC-PLANTED-999.1",
				Path:        "docs/features/frd-warm.md",
				FeatureId:   "FEAT-PLANTED-999",
				Tags:        []string{"AC-PLANTED-999.1"},
				Text:        "Given the index is warm\nWhen a query runs\nThen the planted record is returned",
				Notation:    specmodel.CriterionNotationGherkin,
				ContentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}
}

// plantFeature reads the persisted index at path, appends feature, and
// writes it back byte-for-byte through the same [specmodel.SpecIndex]
// shape corpus.Write renders, leaving InputHash untouched — the fixture's
// source files do not change, so the fingerprint openWarm compares
// against still matches and the corpus is still servable warm.
func plantFeature(t *testing.T, path string, feature specmodel.Feature) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var index specmodel.SpecIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("decoding %s: %v", path, err)
	}
	index.Features = append(index.Features, feature)

	rewritten, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatalf("re-encoding %s: %v", path, err)
	}
	if err := os.WriteFile(path, append(rewritten, '\n'), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// runInFixture runs cmd.Execute against a fresh copy of the warm-index
// fixture and returns its exit code and streams.
func runInFixture(t *testing.T, dir string, args []string) (int, string, string) {
	t.Helper()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%s): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("restoring the working directory: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	exitCode := cmd.Execute(args, &stdout, &stderr)
	return exitCode, stdout.String(), stderr.String()
}

// AC: AC-GHK-007.6
//
// A first run has nowhere warm to read from, so it builds the corpus and
// persists the result (openCorpus's existing best-effort write, the same
// path check, slug-exists, next-id and refs already take). This test
// then plants a feature into that persisted index which exists nowhere
// on disk, and shows a second `feature <id>` run returns it — the only
// way it could, since a re-walk would find no such feature and report it
// not found.
func TestFeatureQueryIsServedFromTheWarmIndexWithoutRewalking(t *testing.T) {
	dir := t.TempDir()
	writeQueryFixture(t, dir)

	if exitCode, _, stderr := runInFixture(t, dir, []string{"feature", "FEAT-QRY-001"}); exitCode != 0 {
		t.Fatalf("priming run: exit code = %d, stderr: %s", exitCode, stderr)
	}

	indexPath := filepath.Join(dir, "docs", "specs.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("the priming run did not persist an index at %s: %v", indexPath, err)
	}
	plantFeature(t, indexPath, plantedFeature())

	exitCode, stdout, stderr := runInFixture(t, dir, []string{"feature", "FEAT-PLANTED-999", "--full"})
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", exitCode, stderr)
	}
	if !strings.HasPrefix(stdout, "declared: true\nfeatures:\n") ||
		!strings.Contains(stdout, "FEAT-PLANTED-999") || !strings.Contains(stdout, "AC-PLANTED-999.1") {
		t.Errorf("stdout = %q, want declared: true and the planted feature and its criterion; a re-walked corpus answers declared: false", stdout)
	}
}

// AC: AC-GHK-007.6
//
// The same proof as [TestFeatureQueryIsServedFromTheWarmIndexWithoutRewalking],
// for `criterion <id>`: the planted feature's nested criterion exists
// nowhere on disk, so only a warm-served answer can return it.
func TestCriterionQueryIsServedFromTheWarmIndexWithoutRewalking(t *testing.T) {
	dir := t.TempDir()
	writeQueryFixture(t, dir)

	if exitCode, _, stderr := runInFixture(t, dir, []string{"feature", "FEAT-QRY-001"}); exitCode != 0 {
		t.Fatalf("priming run: exit code = %d, stderr: %s", exitCode, stderr)
	}

	indexPath := filepath.Join(dir, "docs", "specs.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("the priming run did not persist an index at %s: %v", indexPath, err)
	}
	plantFeature(t, indexPath, plantedFeature())

	exitCode, stdout, stderr := runInFixture(t, dir, []string{"criterion", "AC-PLANTED-999.1"})
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", exitCode, stderr)
	}
	if !strings.HasPrefix(stdout, "declared: true\ncriteria:\n") || !strings.Contains(stdout, "AC-PLANTED-999.1") {
		t.Errorf("stdout = %q, want declared: true and the planted criterion; a re-walked corpus answers declared: false", stdout)
	}
}
