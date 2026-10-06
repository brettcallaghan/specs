// Tests for the listing query: the summary a corpus member projects onto,
// what an undescribable member carries instead, what a non-member carries
// at all, and the two filters that narrow the answer.
//
// The filter fixtures register their lifecycle rather than relying on the
// shipped profile's spelling, and one of them registers a DIFFERENT
// status as the retired one. That is the point: a filter reading the
// registration answers the same way for a profile whose documents move
// from `current` to `sunset`, and a filter holding a spelling of its own
// could not.
package corpus_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// listingProfile registers three document types and a lifecycle whose
// retired set is the statuses named, so a test can decide which spelling
// means "no longer in effect" and watch the filter follow it.
func listingProfile(retired ...specmodel.Status) corpus.Profile {
	withheld := make([]string, 0, len(retired))
	for _, status := range retired {
		withheld = append(withheld, string(status))
	}
	return stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"adr":      {SlugPattern: `^adr-[0-9]{3}$`, Numbered: true},
			"frd":      {SlugPattern: `^frd-[a-z0-9-]+$`},
			"contract": {SlugPattern: `^contract-[a-z0-9-]+$`},
		},
		lifecycle: corpus.LifecycleRules{
			Statuses: []string{
				string(specmodel.StatusDraft), string(specmodel.StatusApproved),
				string(specmodel.StatusActive), string(specmodel.StatusRetired),
			},
			LiveStatuses:    []string{string(specmodel.StatusActive)},
			RetiredStatuses: withheld,
		},
	}
}

// listingCorpusAt builds a corpus over repo under listingProfile.
func listingCorpusAt(t *testing.T, repo string, retired ...specmodel.Status) *corpus.Corpus {
	t.Helper()

	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   listingProfile(retired...),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// documentAt renders a conforming document at a chosen status, which
// `document` fixes at approved.
func documentAt(docType, slug, title string, status specmodel.Status) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: %s
updated_at: "2026-08-05T00:00:00Z"
---

# %s
`, title, slug, docType, status, title)
}

// summarize runs the query and fails the test on the error it is not
// expecting, which is every test here but the unregistered-type ones.
func summarize(t *testing.T, built *corpus.Corpus, opts corpus.SummaryOptions) []corpus.DocumentSummary {
	t.Helper()

	summaries, err := built.DocumentSummaries(opts)
	if err != nil {
		t.Fatalf("DocumentSummaries(%+v): %v", opts, err)
	}
	return summaries
}

// summaryAt returns the summary for path, and fails the test where the
// listing carries none — the caller is always asserting about a member it
// wrote, so an absent one is a failure rather than a case to handle.
func summaryAt(t *testing.T, summaries []corpus.DocumentSummary, path string) corpus.DocumentSummary {
	t.Helper()

	for _, summary := range summaries {
		if summary.Path == path {
			return summary
		}
	}
	t.Fatalf("no summary for %s; the listing carries %v", path, summaryPaths(summaries))
	return corpus.DocumentSummary{}
}

// summaryPaths returns the path of every summary, in listing order.
func summaryPaths(summaries []corpus.DocumentSummary) []string {
	out := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, summary.Path)
	}
	return out
}

// AC: AC-SPECLIST-001.1
func TestEveryCorpusMemberIsSummarizedOnce(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))
	write(t, repo, "specs/frds/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))

	built := corpusAt(t, repo)
	summaries := summarize(t, built, corpus.SummaryOptions{})

	if got, want := summaryPaths(summaries), paths(built.Index()); !equal(got, want) {
		t.Fatalf("summaries = %v, want one per index entry in index order %v", got, want)
	}

	summary := summaryAt(t, summaries, "specs/frds/frd-checkout.md")
	if summary.Slug == nil || string(*summary.Slug) != "frd-checkout" {
		t.Errorf("slug = %v, want frd-checkout", summary.Slug)
	}
	if summary.Type == nil || string(*summary.Type) != "frd" {
		t.Errorf("type = %v, want frd", summary.Type)
	}
	if summary.Status == nil || string(*summary.Status) != "approved" {
		t.Errorf("status = %v, want approved", summary.Status)
	}
	if summary.Title == nil || *summary.Title != "Checkout" {
		t.Errorf("title = %v, want Checkout", summary.Title)
	}
	if summary.Description == nil || *summary.Description != "A fixture document." {
		t.Errorf("description = %v, want the document's own", summary.Description)
	}
	if summary.ParseError != nil {
		t.Errorf("a member that parsed carries a parse error: %s", summary.ParseError.Message)
	}
}

// The declaration order of the summary's fields is the order the command
// surface renders them in, so it is asserted here rather than left to the
// rendering to reveal.
func TestASummaryCarriesItsFieldsInTheContractsOrder(t *testing.T) {
	summaryType := reflect.TypeOf(corpus.DocumentSummary{})

	var got []string
	for i := range summaryType.NumField() {
		tag, _, _ := strings.Cut(summaryType.Field(i).Tag.Get("json"), ",")
		got = append(got, tag)
	}

	want := []string{"path", "slug", "type", "status", "title", "description", "parse_error"}
	if !equal(got, want) {
		t.Errorf("field order = %v, want %v", got, want)
	}
}

// AC: AC-SPECLIST-001.4
func TestASummarysPathIsRepositoryRelativeWithForwardSlashes(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/nested/adr-001.md", document("adr", "adr-001", "Nested"))

	summaries := summarize(t, corpusAt(t, repo), corpus.SummaryOptions{})

	if len(summaries) != 1 {
		t.Fatalf("summaries = %v, want exactly one", summaryPaths(summaries))
	}
	if got, want := summaries[0].Path, "specs/adrs/nested/adr-001.md"; got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if strings.ContainsRune(summaries[0].Path, '\\') {
		t.Errorf("path %q carries a backslash; paths are written with forward slashes on every platform",
			summaries[0].Path)
	}
	if strings.HasPrefix(summaries[0].Path, repo) {
		t.Errorf("path %q is absolute; it is relative to the repository root", summaries[0].Path)
	}
}

// AC: AC-SPECLIST-001.5
func TestAMemberRecordedWithAParseErrorIsSummarizedWithThatFailureAndNoEnvelope(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/templates/broken.md", "---\ntitle: [unclosed\n---\n\n# Broken\n")

	summaries := summarize(t, corpusAt(t, repo), corpus.SummaryOptions{})
	summary := summaryAt(t, summaries, "specs/templates/broken.md")

	if summary.ParseError == nil {
		t.Fatal("the unparsable member's summary carries no parse failure")
	}
	if summary.ParseError.Message == "" {
		t.Error("the parse failure carries no message")
	}
	if summary.ParseError.Span.StartLine == 0 {
		t.Error("the parse failure carries no position; lines are one-based")
	}
	if summary.Slug != nil {
		t.Errorf("slug = %v, want absent: an unparsable member declares none that is known", summary.Slug)
	}
	if summary.Type != nil {
		t.Errorf("type = %v, want absent", summary.Type)
	}
	if summary.Status != nil {
		t.Errorf("status = %v, want absent", summary.Status)
	}
	if summary.Title != nil {
		t.Errorf("title = %v, want absent", summary.Title)
	}
	if summary.Description != nil {
		t.Errorf("description = %v, want absent", summary.Description)
	}
}

// AC: AC-SPECLIST-001.6
func TestAFileRecordedAsAnExclusionIsNotSummarized(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/notes/plain.md", "# A note\n\nNo frontmatter at all.\n")
	write(t, repo, "specs/notes/foreign.md", document("runbook", "runbook-001", "Foreign"))

	built := corpusAt(t, repo)
	if len(built.Index().Exclusions) != 2 {
		t.Fatalf("the fixture recorded %v as exclusions, want both notes",
			excludedPaths(built.Index()))
	}

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{}))
	if want := []string{"specs/adrs/adr-001.md"}; !equal(got, want) {
		t.Errorf("summaries = %v, want %v: an excluded file has no summary", got, want)
	}
}

// A corpus with no members answers with an empty collection rather than a
// nil one, so the answer renders as an empty sequence and not as null
// (AC-SPECLIST-001.7, proved end to end by the list subcommand).
func TestAnEmptyCorpusSummarizesToAnEmptyCollection(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/.keep", "")

	summaries := summarize(t, corpusAt(t, repo), corpus.SummaryOptions{})

	if summaries == nil {
		t.Fatal("summaries = nil, want an empty collection")
	}
	if len(summaries) != 0 {
		t.Errorf("summaries = %v, want none", summaryPaths(summaries))
	}
}

// A summary holds its own copy of the parse failure, so a caller cannot
// reach through it into the index the corpus is still serving from.
func TestASummarysParseFailureIsACopy(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/templates/broken.md", "---\ntitle: [unclosed\n---\n\n# Broken\n")

	built := corpusAt(t, repo)
	summary := summaryAt(t, summarize(t, built, corpus.SummaryOptions{}), "specs/templates/broken.md")
	summary.ParseError.Message = "rewritten by a caller"

	entry, found := built.Entry("specs/templates/broken.md")
	if !found {
		t.Fatal("the unparsable member is absent from the index")
	}
	if entry.ParseError.Message == "rewritten by a caller" {
		t.Error("writing through a summary's parse failure reached the index entry")
	}
}

// listingFixture writes the corpus every filter test below runs over:
// two ADRs at different statuses, one FRD, and one contract that is no
// longer in effect.
func listingFixture(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", documentAt("adr", "adr-001", "First", specmodel.StatusActive))
	write(t, repo, "specs/adrs/adr-002.md", documentAt("adr", "adr-002", "Second", specmodel.StatusDraft))
	write(t, repo, "specs/frds/frd-checkout.md", documentAt("frd", "frd-checkout", "Checkout", specmodel.StatusApproved))
	write(t, repo, "specs/contracts/contract-old.md", documentAt("contract", "contract-old", "Old", specmodel.StatusRetired))
	return repo
}

// AC: AC-SPECLIST-002.1
func TestFilteringByOneRegisteredTypeReturnsEveryMemberOfItAndNoOther(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t), specmodel.StatusRetired)

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{
		Types:       []string{"adr"},
		AllStatuses: true,
	}))

	want := []string{"specs/adrs/adr-001.md", "specs/adrs/adr-002.md"}
	if !equal(got, want) {
		t.Errorf("summaries = %v, want %v", got, want)
	}
}

// AC: AC-SPECLIST-002.3
func TestARegisteredTypeMatchingNoMemberYieldsAnEmptyCollection(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", documentAt("adr", "adr-001", "First", specmodel.StatusActive))
	built := listingCorpusAt(t, repo, specmodel.StatusRetired)

	summaries, err := built.DocumentSummaries(corpus.SummaryOptions{Types: []string{"contract"}})
	if err != nil {
		t.Fatalf("a registered type with no members is an empty answer, not a failure: %v", err)
	}
	if summaries == nil {
		t.Fatal("summaries = nil, want an empty collection")
	}
	if len(summaries) != 0 {
		t.Errorf("summaries = %v, want none", summaryPaths(summaries))
	}
}

// An unregistered type is a malformed question rather than an empty
// category, so it is reported with the registered names and answered with
// no summaries at all (AC-SPECLIST-002.4, proved end to end by the list
// subcommand).
func TestAnUnregisteredTypeIsReportedWithTheRegisteredNamesAndNoSummaries(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t), specmodel.StatusRetired)

	summaries, err := built.DocumentSummaries(corpus.SummaryOptions{Types: []string{"adr", "runbook"}})

	if err == nil {
		t.Fatalf("an unregistered type was answered with %v rather than reported", summaryPaths(summaries))
	}
	if summaries != nil {
		t.Errorf("summaries = %v, want none alongside the report", summaryPaths(summaries))
	}
	if !errors.Is(err, corpus.ErrUnregisteredType) {
		t.Errorf("error %v is not an ErrUnregisteredType; a caller cannot tell it from any other failure", err)
	}
	if !strings.Contains(err.Error(), `"runbook"`) {
		t.Errorf("error %q does not name the value the caller asked for", err)
	}
	for _, registered := range []string{"adr", "contract", "frd"} {
		if !strings.Contains(err.Error(), registered) {
			t.Errorf("error %q does not name the registered type %q", err, registered)
		}
	}
}

// AC: AC-SPECLIST-002.5
func TestAMemberRecordedWithAParseErrorIsReturnedUnderNoTypeFilter(t *testing.T) {
	repo := listingFixture(t)
	write(t, repo, "specs/templates/broken.md", "---\ntitle: [unclosed\n---\n\n# Broken\n")
	built := listingCorpusAt(t, repo, specmodel.StatusRetired)

	unfiltered := summaryPaths(summarize(t, built, corpus.SummaryOptions{}))
	if !contains(unfiltered, "specs/templates/broken.md") {
		t.Fatalf("the unparsable member is missing from the unfiltered listing %v; the fixture proves nothing", unfiltered)
	}

	for _, docType := range []string{"adr", "frd", "contract"} {
		got := summaryPaths(summarize(t, built, corpus.SummaryOptions{
			Types:       []string{docType},
			AllStatuses: true,
		}))
		if contains(got, "specs/templates/broken.md") {
			t.Errorf("--type %s returned the unparsable member; it declares no type to match", docType)
		}
	}
}

// AC: AC-SPECLIST-003.1, AC-SPECLIST-003.3
func TestByDefaultRetiredSummariesAreOmittedAndEveryOtherStatusIsReturned(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t), specmodel.StatusRetired)

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{}))

	want := []string{
		"specs/adrs/adr-001.md",
		"specs/adrs/adr-002.md",
		"specs/frds/frd-checkout.md",
	}
	if !equal(got, want) {
		t.Errorf("summaries = %v, want %v: the withheld status is omitted and every other one returned", got, want)
	}
}

// Asking for every status returns the withheld ones too (AC-SPECLIST-003.2,
// proved end to end by the list subcommand).
func TestAskingForEveryStatusReturnsTheWithheldOnesToo(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t), specmodel.StatusRetired)

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{AllStatuses: true}))

	want := []string{
		"specs/adrs/adr-001.md",
		"specs/adrs/adr-002.md",
		"specs/contracts/contract-old.md",
		"specs/frds/frd-checkout.md",
	}
	if !equal(got, want) {
		t.Errorf("summaries = %v, want %v", got, want)
	}
}

// The decisive test for this delta: the same corpus, under a profile that
// registers a DIFFERENT status as the one no longer in effect, is filtered
// by that registration and not by any spelling the host holds. A filter
// testing for a literal would omit the same document under both
// registrations and pass every other test in this file.
func TestTheStatusFilterFollowsTheProfilesRegistrationRatherThanASpelling(t *testing.T) {
	repo := listingFixture(t)

	got := summaryPaths(summarize(t, listingCorpusAt(t, repo, specmodel.StatusDraft), corpus.SummaryOptions{}))

	want := []string{
		"specs/adrs/adr-001.md",
		"specs/contracts/contract-old.md",
		"specs/frds/frd-checkout.md",
	}
	if !equal(got, want) {
		t.Errorf("summaries = %v, want %v: the registered status is what the filter withholds", got, want)
	}
}

// A profile registering no retired statuses withholds nothing, which is
// the lifecycle saying its documents have none rather than a malformed
// registration.
func TestAProfileRegisteringNoRetiredStatusesWithholdsNothing(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t))

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{}))

	if len(got) != 4 {
		t.Errorf("summaries = %v, want all four members", got)
	}
}

// AC: AC-SPECLIST-003.4
func TestTheTypeAndStatusFiltersCompose(t *testing.T) {
	repo := listingFixture(t)
	write(t, repo, "specs/contracts/contract-new.md",
		documentAt("contract", "contract-new", "New", specmodel.StatusActive))
	built := listingCorpusAt(t, repo, specmodel.StatusRetired)

	got := summaryPaths(summarize(t, built, corpus.SummaryOptions{Types: []string{"contract"}}))

	if want := []string{"specs/contracts/contract-new.md"}; !equal(got, want) {
		t.Errorf("summaries = %v, want %v: only the summaries satisfying both filters", got, want)
	}
}

// An empty Types applies no filter at all, which is not the same as a
// filter that matches nothing.
func TestAnEmptyTypeSelectionAppliesNoFilter(t *testing.T) {
	built := listingCorpusAt(t, listingFixture(t), specmodel.StatusRetired)

	for _, types := range [][]string{nil, {}} {
		got := summaryPaths(summarize(t, built, corpus.SummaryOptions{
			Types:       types,
			AllStatuses: true,
		}))
		if len(got) != 4 {
			t.Errorf("Types=%#v returned %v, want every member", types, got)
		}
	}
}
