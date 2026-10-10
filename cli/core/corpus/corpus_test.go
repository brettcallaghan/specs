// Tests for corpus discovery and index assembly.
//
// The fixture profile below registers real DocumentType values ("adr",
// "frd") rather than invented ones, which is a constraint of the
// generated model rather than a choice: `specmodel.DocumentFrontmatter`
// types its `type` property as the closed `DocumentType` vocabulary, so
// a document declaring a type outside it cannot decode into an index
// entry at all (agent-plugins-9jo tracks that coupling). Nothing in the
// package under test names a document type — the extension point's own
// source scan in profile_test.go is what enforces that — so the host
// stays framework-agnostic whatever this fixture happens to register.
package corpus_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/envelope"
	"github.com/brettcallaghan/specs/cli/core/finding"
	"github.com/brettcallaghan/specs/cli/core/specindex"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// indexProfile is the active profile every build in this file runs
// under: two known document types and nothing else, so that a third type
// is genuinely unknown to it.
func indexProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {SlugPattern: `^adr-[0-9]{3}$`, Numbered: true},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
	}
}

// document renders a conforming document of the given type.
func document(docType, slug, title string) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: approved
---

# %s
`, title, slug, docType, title)
}

// write creates one file beneath dir, making its parent directories.
func write(t *testing.T, dir, rel, content string) string {
	t.Helper()

	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", full, err)
	}
	return full
}

// corpusAt builds a corpus whose repository root is repo and whose spec
// root is repo/specs, which is the shape every fixture here uses.
func corpusAt(t *testing.T, repo string) *corpus.Corpus {
	t.Helper()

	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   indexProfile(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// paths returns the path of every index entry, in index order.
func paths(index specmodel.SpecIndex) []string {
	out := make([]string, 0, len(index.Entries))
	for _, entry := range index.Entries {
		out = append(out, entry.Path)
	}
	return out
}

// excludedPaths returns the path of every exclusion, in index order.
func excludedPaths(index specmodel.SpecIndex) []string {
	out := make([]string, 0, len(index.Exclusions))
	for _, exclusion := range index.Exclusions {
		out = append(out, exclusion.Path)
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// exclusionFor returns the exclusion recorded against one path.
func exclusionFor(t *testing.T, index specmodel.SpecIndex, path string) specmodel.IndexExclusion {
	t.Helper()

	for _, exclusion := range index.Exclusions {
		if exclusion.Path == path {
			return exclusion
		}
	}
	t.Fatalf("no exclusion recorded for %s; exclusions are %v", path, excludedPaths(index))
	return specmodel.IndexExclusion{}
}

// AC: AC-SPEC-001.1
func TestBuildAdmitsMarkdownDeclaringAKnownType(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))

	index := corpusAt(t, repo).Index()

	want := []string{"specs/adrs/adr-001.md", "specs/features/frd-checkout.md"}
	if got := paths(index); len(got) != len(want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
	for _, path := range want {
		if !contains(paths(index), path) {
			t.Errorf("%s is not a member; entries are %v", path, paths(index))
		}
	}
	if len(index.Exclusions) != 0 {
		t.Errorf("conforming documents produced exclusions %v", excludedPaths(index))
	}
}

// AC: AC-SPEC-001.3
func TestBuildExcludesMarkdownWithNoFrontmatter(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/plans/plan.md", "# A plan\n\nNo frontmatter here.\n")

	index := corpusAt(t, repo).Index()

	if contains(paths(index), "specs/plans/plan.md") {
		t.Error("a file with no frontmatter was admitted as a member")
	}
	exclusion := exclusionFor(t, index, "specs/plans/plan.md")
	if exclusion.Reason != specmodel.ExclusionReasonNoFrontmatter {
		t.Errorf("reason = %q, want %q", exclusion.Reason, specmodel.ExclusionReasonNoFrontmatter)
	}
	if exclusion.Type != nil {
		t.Errorf("type = %q reported for a file that declared none", *exclusion.Type)
	}
}

// AC: AC-SPEC-001.4
func TestBuildExcludesAnUnknownTypeAndReportsTheDeclaredValue(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/cookbook/recipe.md", document("cookbook", "recipe-001", "Recipe"))

	index := corpusAt(t, repo).Index()

	if len(index.Entries) != 0 {
		t.Errorf("entries = %v, want none", paths(index))
	}
	exclusion := exclusionFor(t, index, "specs/cookbook/recipe.md")
	if exclusion.Reason != specmodel.ExclusionReasonUnknownType {
		t.Errorf("reason = %q, want %q", exclusion.Reason, specmodel.ExclusionReasonUnknownType)
	}
	if exclusion.Type == nil {
		t.Fatal("the declared type was not reported")
	}
	if *exclusion.Type != "cookbook" {
		t.Errorf("type = %q, want %q", *exclusion.Type, "cookbook")
	}
}

// A file whose frontmatter declares no type at all is excluded for the
// same reason as one declaring an unfamiliar type — it is not a document
// the active profile governs — but there is no declared value to report,
// and reporting an empty one would claim the file said something it did
// not.
func TestBuildExcludesFrontmatterDeclaringNoTypeWithoutInventingAValue(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/notes/note.md", "---\ntitle: \"A note\"\n---\n\n# A note\n")

	index := corpusAt(t, repo).Index()

	exclusion := exclusionFor(t, index, "specs/notes/note.md")
	if exclusion.Reason != specmodel.ExclusionReasonUnknownType {
		t.Errorf("reason = %q, want %q", exclusion.Reason, specmodel.ExclusionReasonUnknownType)
	}
	if exclusion.Type != nil {
		t.Errorf("type = %q reported for a file that declared none", *exclusion.Type)
	}
}

// AC: AC-SPEC-001.5
func TestBuildReadsNothingOutsideTheSpecRoot(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Inside"))
	// A sibling of the spec root, and a document above the repository
	// root entirely. Neither may reach the index by any route.
	write(t, repo, "elsewhere/adr-900.md", document("adr", "adr-900", "Sibling"))
	write(t, filepath.Dir(repo), "outside-the-repo.md", document("adr", "adr-901", "Outside"))

	built := corpusAt(t, repo)
	index := built.Index()

	if got := paths(index); len(got) != 1 || got[0] != "specs/adrs/adr-001.md" {
		t.Errorf("entries = %v, want only the document beneath the spec root", got)
	}
	if len(index.Exclusions) != 0 {
		t.Errorf("exclusions = %v, want none: a file outside the root is not a candidate", excludedPaths(index))
	}
	for _, slug := range []string{"adr-900", "adr-901"} {
		if _, found := built.Slug(slug); found {
			t.Errorf("slug %q resolved; a document outside the spec root was read", slug)
		}
	}
}

// The one route by which a walk could leave the spec root is a symlink,
// so it is closed explicitly rather than left to the walker's defaults.
func TestBuildDoesNotFollowASymlinkOutOfTheSpecRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks on Windows requires privileges this test does not need")
	}

	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Inside"))
	target := write(t, repo, "elsewhere/adr-900.md", document("adr", "adr-900", "Sibling"))

	link := filepath.Join(repo, "specs", "adrs", "linked.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are unavailable here: %v", err)
	}

	built := corpusAt(t, repo)
	index := built.Index()

	if contains(paths(index), "specs/adrs/linked.md") {
		t.Error("a symlink was followed out of the spec root")
	}
	if contains(excludedPaths(index), "specs/adrs/linked.md") {
		t.Error("a symlink was recorded as an exclusion rather than passed over")
	}
	if _, found := built.Slug("adr-900"); found {
		t.Error("a document outside the spec root was read through a symlink")
	}
}

// AC: AC-SPEC-001.6
func TestBuildPassesOverNonMarkdownFilesWithoutExcludingThem(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/specs.json", `{"schema_version":"1.0"}`)
	write(t, repo, "specs/adrs/README.rst", "Not markdown.\n")
	write(t, repo, "specs/adrs/diagram.mdx", "Not markdown either.\n")
	write(t, repo, "specs/adrs/notes.markdown", "Also not.\n")

	index := corpusAt(t, repo).Index()

	if got := paths(index); len(got) != 1 || got[0] != "specs/adrs/adr-001.md" {
		t.Errorf("entries = %v, want only the .md document", got)
	}
	if len(index.Exclusions) != 0 {
		t.Errorf("exclusions = %v, want none: a non-.md file is passed over silently", excludedPaths(index))
	}
}

// AC: AC-SPEC-002.1
func TestIndexCarriesEachMembersFrontmatterAndRepositoryRelativePath(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-011.md", document("adr", "adr-011", "Targeted verification gates"))

	built := corpusAt(t, repo)

	entry, found := built.Entry("specs/adrs/adr-011.md")
	if !found {
		t.Fatalf("no entry at specs/adrs/adr-011.md; entries are %v", paths(built.Index()))
	}
	if entry.ParseError != nil {
		t.Fatalf("parse error: %s", entry.ParseError.Message)
	}
	if entry.Frontmatter == nil {
		t.Fatal("the entry carries no frontmatter")
	}
	fm := entry.Frontmatter
	if envelope.Of(fm).Title != "Targeted verification gates" {
		t.Errorf("title = %q", envelope.Of(fm).Title)
	}
	if envelope.Of(fm).Description != "A fixture document." {
		t.Errorf("description = %q", envelope.Of(fm).Description)
	}
	if string(envelope.Of(fm).Slug) != "adr-011" {
		t.Errorf("slug = %q", envelope.Of(fm).Slug)
	}
	if string(envelope.Of(fm).Type) != "adr" {
		t.Errorf("type = %q", envelope.Of(fm).Type)
	}
	if string(envelope.Of(fm).Status) != "approved" {
		t.Errorf("status = %q", envelope.Of(fm).Status)
	}
}

// AC: AC-SPEC-002.3
func TestIndexDeclaresTheConfiguredWorkspaceAndSchemaVersion(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))

	index := corpusAt(t, repo).Index()

	if index.Workspace != "fixture-workspace" {
		t.Errorf("workspace = %q, want %q", index.Workspace, "fixture-workspace")
	}
	if index.SchemaVersion != specindex.SchemaVersion {
		t.Errorf("schema_version = %q, want %q", index.SchemaVersion, specindex.SchemaVersion)
	}
}

// An empty corpus still declares its workspace, and its collections
// marshal as [] rather than null, so a consumer never has to tell
// "empty" from "absent".
func TestAnEmptyCorpusStillDeclaresItsWorkspace(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "specs"), 0o755); err != nil {
		t.Fatalf("creating the spec root: %v", err)
	}

	index := corpusAt(t, repo).Index()

	if index.Workspace != "fixture-workspace" {
		t.Errorf("workspace = %q", index.Workspace)
	}
	encoded, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("marshalling the index: %v", err)
	}
	for _, want := range []string{`"entries":[]`, `"exclusions":[]`} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("index JSON %s does not contain %s", encoded, want)
		}
	}
}

// AC: AC-SPEC-002.4
func TestAMemberThatFailsToParseIsRecordedAndTheRestStillIndexed(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))
	write(t, repo, "specs/templates/broken.md", "---\ntitle: [unclosed\n---\n\n# Broken\n")

	built := corpusAt(t, repo)
	index := built.Index()

	for _, path := range []string{"specs/adrs/adr-001.md", "specs/adrs/adr-002.md"} {
		entry, found := built.Entry(path)
		if !found {
			t.Fatalf("%s is missing; one unparsable file aborted the index", path)
		}
		if entry.ParseError != nil {
			t.Errorf("%s: unexpected parse error %s", path, entry.ParseError.Message)
		}
	}

	broken, found := built.Entry("specs/templates/broken.md")
	if !found {
		t.Fatalf("the unparsable file is absent from the index; entries are %v", paths(index))
	}
	if broken.ParseError == nil {
		t.Fatal("the unparsable file carries no parse error")
	}
	if broken.Frontmatter != nil {
		t.Error("a member with a parse error also carries frontmatter; exactly one of the two is present")
	}
	if broken.ParseError.Message == "" {
		t.Error("the parse error carries no message")
	}
	if broken.ParseError.Span.StartLine == 0 {
		t.Error("the parse error carries no position; lines are one-based")
	}
}

// A block that parses as YAML but does not fit the frontmatter envelope
// — here a status outside the vocabulary — is the same kind of failure
// and is recorded the same way, rather than aborting the build.
func TestAMemberThatDoesNotFitTheEnvelopeIsRecordedAgainstItself(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/adrs/adr-002.md", strings.Replace(
		document("adr", "adr-002", "Second"), "status: approved", "status: half-baked", 1))

	built := corpusAt(t, repo)

	if _, found := built.Entry("specs/adrs/adr-001.md"); !found {
		t.Error("the conforming document is missing from the index")
	}
	entry, found := built.Entry("specs/adrs/adr-002.md")
	if !found {
		t.Fatal("the non-conforming document is absent from the index")
	}
	if entry.ParseError == nil {
		t.Fatal("the non-conforming document carries no parse error")
	}
	if entry.ParseError.Span.StartLine == 0 {
		t.Error("the parse error carries no position")
	}
}

// AC: AC-SPEC-003.3, AC-SPEC-003.4
func TestDuplicateSlugsAreReportedAgainstEachDocumentNamingTheOther(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/legacy/old-checkout.md", document("frd", "FRD-Checkout", "Checkout again"))
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Untouched"))

	built := corpusAt(t, repo)
	findings := built.Findings()

	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2: %+v", len(findings), findings)
	}

	pairs := map[string]string{}
	for _, f := range findings {
		if f.Rule != finding.RuleDuplicateSlug {
			t.Errorf("rule = %q, want %q", f.Rule, finding.RuleDuplicateSlug)
		}
		if f.Severity != finding.SeverityError {
			t.Errorf("severity = %q, want %q", f.Severity, finding.SeverityError)
		}
		detail, ok := f.Detail.Value.(specmodel.FindingDuplicateSlug)
		if !ok {
			t.Fatalf("detail = %T, want FindingDuplicateSlug", f.Detail.Value)
		}
		pairs[f.Path] = detail.OtherPath
		if f.Span.StartLine == 0 {
			t.Error("the finding carries no position")
		}
		if !strings.Contains(f.Message, detail.OtherPath) {
			t.Errorf("message %q does not name the other document", f.Message)
		}
	}

	want := map[string]string{
		"specs/features/frd-checkout.md": "specs/legacy/old-checkout.md",
		"specs/legacy/old-checkout.md":   "specs/features/frd-checkout.md",
	}
	for path, other := range want {
		if pairs[path] != other {
			t.Errorf("finding against %s names %q, want %q", path, pairs[path], other)
		}
	}

	// The slug each finding reports is the one its own document declared,
	// so the differing case is visible in the report rather than
	// normalised away.
	slugs := map[string]string{}
	for _, f := range findings {
		slugs[f.Path] = f.Detail.Value.(specmodel.FindingDuplicateSlug).Slug
	}
	if slugs["specs/features/frd-checkout.md"] != "frd-checkout" {
		t.Errorf("slug = %q, want %q", slugs["specs/features/frd-checkout.md"], "frd-checkout")
	}
	if slugs["specs/legacy/old-checkout.md"] != "FRD-Checkout" {
		t.Errorf("slug = %q, want %q", slugs["specs/legacy/old-checkout.md"], "FRD-Checkout")
	}
}

// Distinct slugs collide with nothing, so a clean corpus reports
// nothing — the check that keeps the one above from passing vacuously.
func TestDistinctSlugsRaiseNoFindings(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))

	if findings := corpusAt(t, repo).Findings(); len(findings) != 0 {
		t.Errorf("findings = %+v, want none", findings)
	}
}

// AC: AC-SPEC-001.2
func TestASingleCheckedFileResolvesAgainstTheFullCorpus(t *testing.T) {
	repo := t.TempDir()
	checked := write(t, repo, "specs/features/frd-checkout.md",
		document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/legacy/deep/nested/old-checkout.md",
		document("frd", "frd-checkout", "Checkout again"))

	built := corpusAt(t, repo)

	// The named file alone carries no defect. The collision only exists
	// relative to a document in a directory the caller never named, so a
	// check that read the named file in isolation would report nothing.
	findings := built.FindingsFor("specs/features/frd-checkout.md")
	if len(findings) != 1 {
		t.Fatalf("got %d findings for the checked file, want 1: %+v", len(findings), findings)
	}
	detail, ok := findings[0].Detail.Value.(specmodel.FindingDuplicateSlug)
	if !ok {
		t.Fatalf("detail = %T, want FindingDuplicateSlug", findings[0].Detail.Value)
	}
	if detail.OtherPath != "specs/legacy/deep/nested/old-checkout.md" {
		t.Errorf("other_path = %q, want the document discovered from the spec root", detail.OtherPath)
	}

	// The same narrowing is available from an absolute path, which is
	// what a caller invoked with a file argument actually holds.
	if got := built.FindingsFor(checked); len(got) != 1 {
		t.Errorf("got %d findings for the absolute path, want 1", len(got))
	}

	// Narrowing selects one file's findings and no other's.
	other := built.FindingsFor("specs/legacy/deep/nested/old-checkout.md")
	if len(other) != 1 {
		t.Fatalf("got %d findings for the other file, want 1", len(other))
	}
	if other[0].Path == findings[0].Path {
		t.Error("narrowing returned the same file's findings for two different paths")
	}
}

// AC: AC-SPEC-007.2
func TestSlugQueryReportsWhetherAMemberDeclaresItAndItsPath(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-004.md", document("adr", "adr-004", "Module blueprint"))
	write(t, repo, "specs/plans/plan.md", "# No frontmatter\n")

	built := corpusAt(t, repo)

	entry, found := built.Slug("adr-004")
	if !found {
		t.Fatal("slug adr-004 is in use and the query did not find it")
	}
	if entry.Path != "specs/adrs/adr-004.md" {
		t.Errorf("path = %q, want %q", entry.Path, "specs/adrs/adr-004.md")
	}

	// The query compares case-insensitively, the same way duplicate
	// detection does: a corpus cannot hold the two spellings as distinct
	// identifiers, so a query for either must find the other.
	if _, found := built.Slug("ADR-004"); !found {
		t.Error("the slug query is case-sensitive")
	}

	if entry, found := built.Slug("adr-999"); found {
		t.Errorf("slug adr-999 is declared by nothing and resolved to %q", entry.Path)
	}
}

// AC: AC-SPEC-009.8
func TestPathsAreForwardSlashedAndRepositoryRelative(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/nested/deeper/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/adrs/nested/deeper/adr-002.md", document("adr", "adr-001", "Colliding"))
	write(t, repo, "specs/adrs/nested/plan.md", "# No frontmatter\n")

	built := corpusAt(t, repo)
	index := built.Index()

	var reported []string
	reported = append(reported, paths(index)...)
	reported = append(reported, excludedPaths(index)...)
	for _, f := range built.Findings() {
		reported = append(reported, f.Path)
		reported = append(reported, f.Detail.Value.(specmodel.FindingDuplicateSlug).OtherPath)
	}
	if len(reported) == 0 {
		t.Fatal("nothing was reported; the check would pass vacuously")
	}

	for _, path := range reported {
		if strings.Contains(path, `\`) {
			t.Errorf("path %q uses a backslash separator", path)
		}
		if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
			t.Errorf("path %q is not relative to the repository root", path)
		}
		if strings.HasPrefix(path, "..") {
			t.Errorf("path %q escapes the repository root", path)
		}
		if !strings.HasPrefix(path, "specs/adrs/nested/") {
			t.Errorf("path %q is not relative to the repository root", path)
		}
	}
}

// Directory iteration order is a property of a filesystem, not of a
// corpus, so entries, exclusions and findings are ordered by the
// component rather than by the walk. Two builds of the same content must
// be byte-identical, and the order must be the sorted one rather than
// whatever a depth-first walk happens to yield.
func TestEntriesExclusionsAndFindingsAreOrderedDeterministically(t *testing.T) {
	repo := t.TempDir()
	// "specs/adrs.md" sorts before "specs/adrs/..." by path, and after it
	// in a depth-first walk, so a build that simply appended as it walked
	// would fail this.
	write(t, repo, "specs/adrs.md", document("adr", "adr-000", "Index page"))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/zeta.md", document("frd", "frd-zeta", "Zeta"))
	write(t, repo, "specs/alpha.md", document("frd", "frd-zeta", "Alpha"))
	write(t, repo, "specs/notes/b.md", "# No frontmatter\n")
	write(t, repo, "specs/notes/a.md", "# No frontmatter\n")

	first := corpusAt(t, repo)
	index := first.Index()

	wantEntries := []string{
		"specs/adrs.md",
		"specs/adrs/adr-001.md",
		"specs/adrs/adr-002.md",
		"specs/alpha.md",
		"specs/zeta.md",
	}
	if got := paths(index); !equal(got, wantEntries) {
		t.Errorf("entries = %v, want %v", got, wantEntries)
	}
	wantExclusions := []string{"specs/notes/a.md", "specs/notes/b.md"}
	if got := excludedPaths(index); !equal(got, wantExclusions) {
		t.Errorf("exclusions = %v, want %v", got, wantExclusions)
	}
	wantFindings := []string{"specs/alpha.md", "specs/zeta.md"}
	var findingPaths []string
	for _, f := range first.Findings() {
		findingPaths = append(findingPaths, f.Path)
	}
	if !equal(findingPaths, wantFindings) {
		t.Errorf("finding paths = %v, want %v", findingPaths, wantFindings)
	}

	encode := func(c *corpus.Corpus) string {
		t.Helper()
		payload := struct {
			Index    specmodel.SpecIndex `json:"index"`
			Findings []specmodel.Finding `json:"findings"`
		}{c.Index(), c.Findings()}
		out, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		return string(out)
	}
	if a, b := encode(first), encode(corpusAt(t, repo)); a != b {
		t.Errorf("two builds of one corpus differ:\n%s\n%s", a, b)
	}
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// AC: AC-SPEC-009.3, AC-SPEC-009.9
func TestAColdThousandDocumentBuildMeetsTheBudget(t *testing.T) {
	const (
		documents = 1000
		target    = time.Second
	)

	repo := t.TempDir()
	for i := range documents {
		// Spread the corpus over directories rather than piling it into
		// one, because a real corpus is nested and a flat one would not
		// exercise the walk.
		rel := fmt.Sprintf("specs/section-%02d/adr-%04d.md", i%20, i)
		write(t, repo, rel, document("adr", fmt.Sprintf("adr-%04d", i), fmt.Sprintf("Document %d", i)))
	}

	// The build is cold by construction: nothing here persists an index
	// or reads one, so this is the whole cost of discovery.
	started := time.Now()
	built := corpusAt(t, repo)
	elapsed := time.Since(started)

	if got := len(built.Index().Entries); got != documents {
		t.Fatalf("indexed %d documents, want %d", got, documents)
	}
	if len(built.Findings()) != 0 {
		t.Fatalf("findings = %+v, want none", built.Findings())
	}
	reportLatency(t, fmt.Sprintf("cold build of %d documents", documents), elapsed, target)
}

func TestBuildRejectsAnUnusableConfiguration(t *testing.T) {
	repo := t.TempDir()

	if _, err := corpus.Build(corpus.Options{Root: repo}); err == nil {
		t.Error("Build succeeded with no profile")
	}
	if _, err := corpus.Build(corpus.Options{Profile: indexProfile()}); err == nil {
		t.Error("Build succeeded with no spec root")
	}
	missing := filepath.Join(repo, "no-such-directory")
	if _, err := corpus.Build(corpus.Options{Root: missing, Profile: indexProfile()}); err == nil {
		t.Error("Build succeeded against a spec root that does not exist")
	}
}

// The repository root defaults to the spec root, so a caller that
// configures only a root gets paths relative to it rather than absolute
// ones.
func TestRepositoryRootDefaultsToTheSpecRoot(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))

	built, err := corpus.Build(corpus.Options{
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   indexProfile(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := paths(built.Index()); !equal(got, []string{"adrs/adr-001.md"}) {
		t.Errorf("entries = %v, want paths relative to the spec root", got)
	}
}
