package markdown_test

import (
	"reflect"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
)

// AC: AC-SPECSTR-003.9
func TestAHeadingPathIsTheIdentifiersOfItsEnclosingHeadingsThenItsOwn(t *testing.T) {
	outline := []markdown.Outline{
		{Text: "Document Core", Level: 1},
		{Text: "Components", Level: 2},
		{Text: "References", Level: 3},
		{Text: "Deep", Level: 6}, // skipped levels: still nested under References
		{Text: "Findings", Level: 3},
		{Text: "References", Level: 2}, // a sibling of Components, not nested in it
		{Text: "Appendix", Level: 1},
	}
	want := [][]string{
		{"document-core"},
		{"document-core", "components"},
		{"document-core", "components", "references"},
		{"document-core", "components", "references", "deep"},
		{"document-core", "components", "findings"},
		{"document-core", "references"},
		{"appendix"},
	}
	if got := markdown.HeadingPaths(outline); !reflect.DeepEqual(got, want) {
		t.Errorf("HeadingPaths =\n %v\nwant\n %v", got, want)
	}
	if got := markdown.JoinPath(want[2]); got != "document-core#components#references" {
		t.Errorf("JoinPath = %q", got)
	}
}

// A document that opens below the top level has paths beginning at its
// shallowest heading, and an empty outline has none.
//
// AC: AC-SPECSTR-003.9
func TestHeadingPathsOfADocumentWithNoTopLevelHeading(t *testing.T) {
	got := markdown.HeadingPaths([]markdown.Outline{{Text: "A", Level: 3}, {Text: "B", Level: 4}, {Text: "C", Level: 2}})
	want := [][]string{{"a"}, {"a", "b"}, {"c"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("HeadingPaths = %v, want %v", got, want)
	}
	if got := markdown.HeadingPaths(nil); len(got) != 0 {
		t.Errorf("HeadingPaths(nil) = %v, want none", got)
	}
}

// AC: AC-SPECSTR-003.2, AC-SPECSTR-003.10, AC-SPECSTR-003.11, AC-SPECANCH-002.10
func TestAPathQueryMatchesTheLastSegmentBeneathEachEarlierOneInOrder(t *testing.T) {
	roster := []string{"document-core", "components", "registry", "references"}
	cases := []struct {
		name  string
		query string
		path  []string
		want  bool
	}{
		{"single segment is an identifier comparison", "References", roster, true},
		{"single segment, case and punctuation ignored", "references!", roster, true},
		{"single segment naming an ancestor does not match", "Components", roster, false},
		{"direct parent", "Registry#References", roster, true},
		{"ancestor at any depth", "Components#References", roster, true},
		{"the full heading path pastes back", "document-core#components#registry#references", roster, true},
		{"gaps allowed between several segments", "Document Core#Registry#References", roster, true},
		{"out-of-order segments rejected", "Registry#Components#References", roster, false},
		{"a segment naming no ancestor rejected", "Findings#References", roster, false},
		{"last segment must be the heading itself", "Components#Registry", roster, false},
		{"the heading cannot be its own ancestor", "References#References", roster, false},
		{"an empty segment rejected", "Components##References", roster, false},
		{"a trailing separator rejected", "Components#", roster, false},
		{"a leading separator rejected", "#References", roster, false},
		{"no path, no match", "References", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := markdown.MatchesPath(c.query, c.path); got != c.want {
				t.Errorf("MatchesPath(%q, %v) = %v, want %v", c.query, c.path, got, c.want)
			}
		})
	}
}

// A single-segment query decides exactly what an identifier comparison
// decides, over every heading of a document — the reduction that keeps
// every existing selection and anchor answering as it did.
//
// AC: AC-SPECSTR-003.2
func TestASingleSegmentQueryIsTodaysIdentifierMatch(t *testing.T) {
	outline := []markdown.Outline{
		{Text: "Top", Level: 1}, {Text: "1. Overview", Level: 2}, {Text: "Overview", Level: 2},
		{Text: "`Code` Heading", Level: 3}, {Text: "Under_score", Level: 2},
	}
	paths := markdown.HeadingPaths(outline)
	for _, query := range []string{"overview", "1. Overview", "code heading", "Under_score", "TOP", "missing"} {
		for i, heading := range outline {
			want := markdown.Identifier(heading.Text) == markdown.Identifier(query)
			if got := markdown.MatchesPath(query, paths[i]); got != want {
				t.Errorf("MatchesPath(%q, %v) = %v, identifier comparison says %v", query, paths[i], got, want)
			}
		}
	}
}
