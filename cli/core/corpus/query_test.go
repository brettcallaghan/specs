// Tests for the listing query: the summary a corpus member projects onto,
// what an undescribable member carries instead, and what a non-member
// carries at all.
package corpus_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
)

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
	summaries := built.DocumentSummaries()

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

	summaries := corpusAt(t, repo).DocumentSummaries()

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

	summaries := corpusAt(t, repo).DocumentSummaries()
	summary := summaryAt(t, summaries, "specs/templates/broken.md")

	if summary.ParseError == nil {
		t.Fatal("the unparsable member's summary carries no parse failure")
	}
	if summary.ParseError.Message == "" {
		t.Error("the parse failure carries no message")
	}
	if summary.ParseError.Position.StartLine == 0 {
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

	got := summaryPaths(built.DocumentSummaries())
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

	summaries := corpusAt(t, repo).DocumentSummaries()

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
	summary := summaryAt(t, built.DocumentSummaries(), "specs/templates/broken.md")
	summary.ParseError.Message = "rewritten by a caller"

	entry, found := built.Entry("specs/templates/broken.md")
	if !found {
		t.Fatal("the unparsable member is absent from the index")
	}
	if entry.ParseError.Message == "rewritten by a caller" {
		t.Error("writing through a summary's parse failure reached the index entry")
	}
}
