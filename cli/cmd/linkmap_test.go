// Tests for LoadLinkMap — reading the caller-supplied YAML link map that
// resolves a citation's slug to the path a link to it should use.
package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
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
	// A syntax fault stays a syntax fault, told apart from a map of the
	// wrong shape (AC-SPECMD-005.13).
	if v, ok := findings[0].Detail.AsSchemaViolation(); !ok || v.Constraint != "yaml-syntax" {
		t.Errorf("detail = %+v, want constraint yaml-syntax", findings[0].Detail)
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

// shapeViolations returns the link-map-shape findings among findings.
func shapeViolations(t *testing.T, findings []specmodel.Finding) []specmodel.Finding {
	t.Helper()
	var out []specmodel.Finding
	for _, f := range findings {
		if v, ok := f.Detail.AsSchemaViolation(); ok && v.Constraint == "link-map-shape" {
			out = append(out, f)
		}
	}
	return out
}

// A map whose top level is not a mapping is reported with the shape it
// should have, not with the parser's account of the mismatch.
//
// AC: AC-SPECMD-005.13
func TestLoadLinkMapReportsATopLevelListWithTheExpectedShape(t *testing.T) {
	path := writeLinkMap(t, t.TempDir(), "links.yaml", "- adr-012\n- adr-013\n")

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want no map resolved", links)
	}
	shape := shapeViolations(t, findings)
	if len(findings) != 1 || len(shape) != 1 {
		t.Fatalf("findings = %+v, want one link-map-shape finding", findings)
	}
	if shape[0].Rule != cmd.RuleLinkMapInvalid || !strings.Contains(shape[0].Message, `record declaring "path"`) {
		t.Errorf("finding = %s %q, want linkmap/invalid naming the record-with-path shape", shape[0].Rule, shape[0].Message)
	}
	if strings.Contains(shape[0].Message, "where mapping is expected") {
		t.Errorf("message %q carries the parser's wording", shape[0].Message)
	}
}

// An entry that is not a record, or whose path is not a string, is
// reported per slug with what was found and the expected shape, every one
// in one run.
//
// AC: AC-SPECMD-005.13, AC-SPECMD-005.8
func TestLoadLinkMapReportsEveryEntryOfTheWrongShape(t *testing.T) {
	path := writeLinkMap(t, t.TempDir(), "links.yaml", `
adr-013:
  path: [a, b]
adr-012: https://x
adr-014:
  path: ok.md
`)

	links, findings := cmd.LoadLinkMap(path)
	if links != nil {
		t.Errorf("links = %+v, want no map resolved", links)
	}
	shape := shapeViolations(t, findings)
	if len(shape) != 2 || len(findings) != 2 {
		t.Fatalf("findings = %+v, want two link-map-shape findings", findings)
	}
	for i, want := range []struct{ slug, found string }{{"adr-012", "a string"}, {"adr-013", "a list"}} {
		f := shape[i]
		for _, part := range []string{want.slug, want.found, `"path"`} {
			if !strings.Contains(f.Message, part) {
				t.Errorf("finding %d message %q does not carry %q", i, f.Message, part)
			}
		}
		if v, _ := f.Detail.AsSchemaViolation(); v.Pointer != "/"+want.slug {
			t.Errorf("finding %d pointer = %q, want /%s", i, v.Pointer, want.slug)
		}
	}
}
