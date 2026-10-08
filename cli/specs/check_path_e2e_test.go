// End-to-end tests for `check <path>` reporting the references elsewhere
// that the named document's change broke. The comparison base is git
// HEAD, which a snapshot case's input cannot hold, so each test builds
// its repository at test time with the git executable (TestMain in
// criterion_changed_e2e_test.go isolates it) and runs the real binary.
package main_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// adr renders a conforming ADR declaring slug, with body after its title.
func adr(slug, body string) string {
	return fmt.Sprintf("---\ntitle: \"%s\"\ndescription: A fixture ADR.\nslug: %s\ntype: adr\nstatus: approved\nupdated_at: 2026-10-08T00:00:00Z\n---\n\n# %s\n\n%s\n",
		slug, slug, slug, body)
}

const (
	citedPath   = "docs/adrs/adr-005-cited.md"
	citingB     = "docs/adrs/adr-006-citing.md"
	citingC     = "docs/adrs/adr-007-citing.md"
	unrelatedD  = "docs/adrs/adr-008-unrelated.md"
	headingBody = "## Decision\n\nThe decision.\n\n## Consequences\n\nSome."
)

// citationRepo commits a corpus where B and C cite adr-005 (C also by an
// anchor into its Decision heading) and D holds a dangling reference
// unrelated to adr-005.
func citationRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "specs.toml", specsTOML)
	writeFile(t, dir, citedPath, adr("adr-005", headingBody))
	writeFile(t, dir, citingB, adr("adr-006", "Builds on [[adr-005]] and again on [[ADR-005]]."))
	writeFile(t, dir, citingC, adr("adr-007", "Follows [[adr-005#Decision]]."))
	writeFile(t, dir, unrelatedD, adr("adr-008", "Cites [[adr-404]], which never existed."))
	commitAll(t, dir, "base")
	return dir
}

// checkAnswer is the part of check's JSON answer these tests read.
type checkAnswer struct {
	Findings []struct {
		Rule string `json:"rule"`
		Path string `json:"path"`
		Span struct {
			StartLine int `json:"start_line"`
		} `json:"span"`
	} `json:"findings"`
}

// located lists the answer's findings as "rule path:line", in order.
func located(t *testing.T, stdout string) []string {
	t.Helper()
	var answer checkAnswer
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("decoding %q: %v", stdout, err)
	}
	var out []string
	for _, f := range answer.Findings {
		out = append(out, fmt.Sprintf("%s %s:%d", f.Rule, f.Path, f.Span.StartLine))
	}
	return out
}

// corpusFindingsIn runs a whole-corpus check and keeps the findings
// located in paths, which is what a single-document check must reproduce
// exactly for them.
func corpusFindingsIn(t *testing.T, bin, dir string, paths ...string) []string {
	t.Helper()
	stdout, stderr, _ := runAll(t, bin, dir, "check", "--json")
	var kept []string
	for _, f := range located(t, stdout) {
		for _, p := range paths {
			if strings.Contains(f, " "+p+":") {
				kept = append(kept, f)
			}
		}
	}
	if len(kept) == 0 {
		t.Fatalf("the corpus check reported nothing in %v\n%s%s", paths, stdout, stderr)
	}
	return kept
}

// AC: AC-SPECREF-002.11, AC-SPECREF-002.9
func TestCheckPathReportsCitationsAnUncommittedSlugRenameBroke(t *testing.T) {
	bin := sharedSpec(t)
	dir := citationRepo(t)
	writeFile(t, dir, citedPath, adr("adr-905", headingBody))

	want := corpusFindingsIn(t, bin, dir, citingB, citingC)
	stdout, stderr, code := runAll(t, bin, dir, "check", "--json", citedPath)
	if code != 1 {
		t.Fatalf("exit %d, want 1\n%s%s", code, stdout, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want nothing inside a work tree", stderr)
	}
	got := located(t, stdout)
	if !slices.Equal(got, want) {
		t.Errorf("findings = %v, want the corpus check's findings in the citing documents %v", got, want)
	}
	if len(want) != 3 {
		t.Errorf("want 3 broken citations (two in B, the anchored one in C), the corpus check gave %v", want)
	}
	for _, f := range got {
		if strings.Contains(f, unrelatedD) {
			t.Errorf("the unrelated dangling reference in D was reported: %s", f)
		}
	}
}

// AC: AC-SPECREF-002.11
func TestCheckPathReportsAnchorsElsewhereAHeadingRemovalBroke(t *testing.T) {
	bin := sharedSpec(t)
	dir := citationRepo(t)
	writeFile(t, dir, citedPath, adr("adr-005", "## Consequences\n\nSome."))

	want := corpusFindingsIn(t, bin, dir, citingC)
	stdout, stderr, code := runAll(t, bin, dir, "check", "--json", citedPath)
	if code != 1 {
		t.Fatalf("exit %d, want 1\n%s%s", code, stdout, stderr)
	}
	got := located(t, stdout)
	if !slices.Equal(got, want) || len(got) != 1 || !strings.HasPrefix(got[0], "reference/unresolved-anchor ") {
		t.Errorf("findings = %v, want exactly C's unresolved anchor %v", got, want)
	}
}

// A committed rename is the base itself, so nothing elsewhere is reported
// as broken by this document's change; nor is anything for a document new
// since HEAD.
//
// AC: AC-SPECREF-002.9
func TestCheckPathReportsNothingElsewhereWithoutAChangeAgainstHEAD(t *testing.T) {
	bin := sharedSpec(t)
	dir := citationRepo(t)
	writeFile(t, dir, citedPath, adr("adr-905", headingBody))
	commitAll(t, dir, "rename")

	stdout, stderr, code := runAll(t, bin, dir, "check", "--json", citedPath)
	if code != 0 || len(located(t, stdout)) != 0 {
		t.Errorf("exit %d, findings %v; want 0 and none\n%s", code, located(t, stdout), stderr)
	}

	writeFile(t, dir, "docs/adrs/adr-009-new.md", adr("adr-009", "New."))
	stdout, stderr, code = runAll(t, bin, dir, "check", "--json", "docs/adrs/adr-009-new.md")
	if code != 0 || len(located(t, stdout)) != 0 || stderr != "" {
		t.Errorf("a new document: exit %d, findings %v, stderr %q; want 0, none, nothing", code, located(t, stdout), stderr)
	}
}

// AC: AC-SPECREF-002.13
func TestCheckPathOutsideAWorkTreeNotesWhatItSkipped(t *testing.T) {
	bin := sharedSpec(t)
	repo := citationRepo(t)
	dir := t.TempDir()
	for _, rel := range []string{"specs.toml", citedPath, citingB, citingC, unrelatedD} {
		content, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, dir, rel, string(content))
	}
	// adr-005 with a defect of its own, renamed: its own finding is
	// reported; what the rename broke elsewhere cannot be, with no HEAD.
	writeFile(t, dir, citedPath, adr("adr-905", headingBody+"\n\nSee [[adr-777]]."))

	stdout, stderr, code := runAll(t, bin, dir, "check", "--json", citedPath)
	if code != 1 {
		t.Errorf("exit %d, want 1 (the document's own error)\n%s%s", code, stdout, stderr)
	}
	got := located(t, stdout)
	if len(got) != 1 || !strings.Contains(got[0], citedPath) {
		t.Errorf("findings = %v, want only the document's own", got)
	}
	if lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n"); len(lines) != 1 ||
		!strings.HasPrefix(lines[0], "note: ") ||
		!strings.Contains(lines[0], "references elsewhere that this document's change broke are not checked") {
		t.Errorf("stderr = %q, want one note saying what was not checked", stderr)
	}
}
