// Tests for LoadLinkMap — reading the caller-supplied YAML link map that
// resolves a citation's slug to the path a link to it should use.
package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// writeLinkMap writes content to name under dir and returns its path.
func writeLinkMap(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
	return path
}

// AC: AC-SPECMD-005.4
func TestLoadLinkMapResolvesEachSlugToItsRecordsPath(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
template-authoring-standards:
  path: template-authoring-standards.md
ac-traceability-standards:
  path: ac-traceability-standards.md
`)

	links, findings := cmd.LoadLinkMap(path)
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
	want := cmd.LinkMap{
		"template-authoring-standards": "template-authoring-standards.md",
		"ac-traceability-standards":    "ac-traceability-standards.md",
	}
	if len(links) != len(want) {
		t.Fatalf("links = %+v, want %+v", links, want)
	}
	for slug, path := range want {
		if links[slug] != path {
			t.Errorf("links[%q] = %q, want %q", slug, links[slug], path)
		}
	}
}

// AC: AC-SPECMD-005.4
func TestLoadLinkMapPassesAPathThroughUninterpretedAndUnnormalized(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
template-authoring-standards:
  path: /standards/template-authoring/
`)

	links, findings := cmd.LoadLinkMap(path)
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
	if got := links["template-authoring-standards"]; got != "/standards/template-authoring/" {
		t.Errorf("path = %q, want it unresolved and unnormalized", got)
	}
}

// AC: AC-SPECMD-005.6
func TestLoadLinkMapIgnoresAnyPropertyOtherThanPath(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
template-authoring-standards:
  path: template-authoring-standards.md
  destinations:
    - plugin/skills/groom-feature-blueprint/references/shared/
    - plugin/skills/spec/references/shared/
`)

	links, findings := cmd.LoadLinkMap(path)
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
	if got := links["template-authoring-standards"]; got != "template-authoring-standards.md" {
		t.Errorf("path = %q, want template-authoring-standards.md", got)
	}
}

// AC: AC-SPECMD-005.7
func TestLoadLinkMapAcceptsAMapWrittenInJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.json", `{"template-authoring-standards": {"path": "template-authoring-standards.md"}}`)

	links, findings := cmd.LoadLinkMap(path)
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
	if got := links["template-authoring-standards"]; got != "template-authoring-standards.md" {
		t.Errorf("path = %q, want template-authoring-standards.md", got)
	}
}

// AC: AC-SPECMD-005.8
func TestLoadLinkMapReportsAnUnreadableFileAndResolvesNoMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.yaml")

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want nil (a map that fails to read resolves no map at all)", links)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v, want exactly one", findings)
	}
	if findings[0].Rule != cmd.RuleLinkMapInvalid {
		t.Errorf("Rule = %q, want %q", findings[0].Rule, cmd.RuleLinkMapInvalid)
	}
	if findings[0].Path != path {
		t.Errorf("Path = %q, want %q (the finding names the file)", findings[0].Path, path)
	}
}

// AC: AC-SPECMD-005.8
func TestLoadLinkMapReportsInvalidYAMLAndResolvesNoMap(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", "template-authoring-standards: [unterminated\n")

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want nil", links)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v, want exactly one", findings)
	}
	if findings[0].Rule != cmd.RuleLinkMapInvalid {
		t.Errorf("Rule = %q, want %q", findings[0].Rule, cmd.RuleLinkMapInvalid)
	}
}

// AC: AC-SPECMD-005.8
func TestLoadLinkMapReportsADuplicateSlugAsAParseFaultRatherThanOneWinning(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
template-authoring-standards:
  path: first.md
template-authoring-standards:
  path: second.md
`)

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want nil (neither record silently wins)", links)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v, want exactly one", findings)
	}
	if findings[0].Rule != cmd.RuleLinkMapInvalid {
		t.Errorf("Rule = %q, want %q", findings[0].Rule, cmd.RuleLinkMapInvalid)
	}
}

// AC: AC-SPECMD-005.8
func TestLoadLinkMapReportsARecordWithNoPathAndResolvesNoMap(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
template-authoring-standards:
  destinations:
    - plugin/skills/groom-feature-blueprint/references/shared/
`)

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want nil", links)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v, want exactly one", findings)
	}
	if findings[0].Rule != cmd.RuleLinkMapMissingPath {
		t.Errorf("Rule = %q, want %q", findings[0].Rule, cmd.RuleLinkMapMissingPath)
	}
	if findings[0].Path != path {
		t.Errorf("Path = %q, want %q (the finding names the file)", findings[0].Path, path)
	}
}

// AC: AC-SPECMD-005.8
func TestLoadLinkMapReportsEveryMissingPathRecordInOneRun(t *testing.T) {
	dir := t.TempDir()
	path := writeLinkMap(t, dir, "links.yaml", `
ac-traceability-standards:
  destinations: []
template-authoring-standards:
  destinations: []
`)

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want nil", links)
	}
	if len(findings) != 2 {
		t.Fatalf("findings = %+v, want exactly two (every defective record in one run)", findings)
	}
	// Deterministic despite Go's randomized map iteration: findings are
	// ordered by slug.
	if findings[0].Message == "" || findings[1].Message == "" {
		t.Fatalf("findings = %+v, want both to carry a message", findings)
	}
	first, ok := findings[0].Detail.AsMissingProperty()
	if !ok {
		t.Fatalf("findings[0].Detail is not a FindingMissingProperty: %+v", findings[0].Detail)
	}
	if first.Pointer != "/ac-traceability-standards" {
		t.Errorf("findings[0] pointer = %q, want /ac-traceability-standards (sorted by slug)", first.Pointer)
	}
	second, ok := findings[1].Detail.AsMissingProperty()
	if !ok {
		t.Fatalf("findings[1].Detail is not a FindingMissingProperty: %+v", findings[1].Detail)
	}
	if second.Pointer != "/template-authoring-standards" {
		t.Errorf("findings[1] pointer = %q, want /template-authoring-standards (sorted by slug)", second.Pointer)
	}
}
