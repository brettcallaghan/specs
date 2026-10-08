// Tests for the anchor half of the reference checks: what a corpus check
// reports about a reference's anchor, once resolution has settled what
// that anchor reached.
//
// The fixtures are check_test.go's — the same profile, document renderer
// and build helper the slug-half reference checks use — because an
// anchored reference is an ordinary reference with a second half, not a
// second kind of thing.
package corpus_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// sectioned renders a conforming document whose body carries the given
// headings, each standing over a line of prose, so that a reference into
// it has something to resolve against.
func sectioned(docType, slug, title string, headings ...string) string {
	var body strings.Builder
	for _, heading := range headings {
		body.WriteString("\n" + heading + "\n\nProse beneath it.\n")
	}
	return citing(docType, slug, title, body.String())
}

// anchorFindingsOf returns every finding of the two anchor rules, in the
// order the check reported them.
func anchorFindingsOf(findings []specmodel.Finding) []specmodel.Finding {
	var out []specmodel.Finding
	for _, f := range findings {
		if f.Rule == finding.RuleUnresolvedAnchor || f.Rule == finding.RuleAmbiguousAnchor {
			out = append(out, f)
		}
	}
	return out
}

// checkedAnchors builds a corpus under repo with the reference-bearing
// fixture profile and returns its findings.
func checkedAnchors(t *testing.T, repo string) []specmodel.Finding {
	t.Helper()
	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	return findings
}

// An anchor that matches no heading of a target whose headings are known
// is an error, raised against the REFERRING document at the reference's
// own span and naming both the target document and the anchor as
// written.
//
// AC: AC-SPECANCH-002.1
func TestAnAnchorMatchingNoHeadingIsAnErrorNamingTheTargetAndTheAnchor(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-002.md", sectioned("adr", "adr-002", "Second", "## Decision", "## Consequences"))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002#Rationale]]."))

	got := anchorFindingsOf(checkedAnchors(t, repo))

	if len(got) != 1 {
		t.Fatalf("got %d anchor findings, want 1: %+v", len(got), got)
	}
	if got[0].Rule != finding.RuleUnresolvedAnchor {
		t.Errorf("rule = %q, want %q", got[0].Rule, finding.RuleUnresolvedAnchor)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
	if got[0].Path != "specs/adrs/adr-001.md" {
		t.Errorf("path = %q, want the referring document", got[0].Path)
	}
	if got[0].Span.Length == 0 {
		t.Error("finding carries an empty span")
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedAnchor)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedAnchor", got[0].Detail.Value)
	}
	if detail.Anchor != "Rationale" {
		t.Errorf("anchor = %q, want %q", detail.Anchor, "Rationale")
	}
	if detail.TargetPath != "specs/adrs/adr-002.md" {
		t.Errorf("target_path = %q, want the target document", detail.TargetPath)
	}
	if !strings.Contains(got[0].Message, "Rationale") || !strings.Contains(got[0].Message, "adr-002") {
		t.Errorf("message = %q, want it to name the anchor and the target", got[0].Message)
	}
}

// An anchor whose identifier is the identifier of more than one heading
// is an error carrying every match with its span, so the reader can see
// which headings to disambiguate rather than only that a collision
// exists. The anchor resolves to none of them.
//
// AC: AC-SPECANCH-002.5
func TestAnAmbiguousAnchorIsAnErrorNamingEveryMatchingHeadingAndItsSpan(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-002.md",
		sectioned("adr", "adr-002", "Second", "## References", "## Decision", "### references"))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002#References]]."))

	findings := checkedAnchors(t, repo)
	got := anchorFindingsOf(findings)

	if len(got) != 1 {
		t.Fatalf("got %d anchor findings, want 1: %+v", len(got), got)
	}
	if got[0].Rule != finding.RuleAmbiguousAnchor {
		t.Errorf("rule = %q, want %q", got[0].Rule, finding.RuleAmbiguousAnchor)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingAmbiguousAnchor)
	if !ok {
		t.Fatalf("detail = %T, want FindingAmbiguousAnchor", got[0].Detail.Value)
	}
	if len(detail.Matches) != 2 {
		t.Fatalf("got %d matches, want 2 — the anchor is matched on identifiers, so letter case does not decide one: %+v", len(detail.Matches), detail.Matches)
	}
	for i, match := range detail.Matches {
		if match.Span.Length == 0 {
			t.Errorf("match %d carries an empty span", i)
		}
	}
	if detail.Matches[0].Span.Offset == detail.Matches[1].Span.Offset {
		t.Error("both matches carry the same span; each names its own heading")
	}

	// The reference itself is not also reported as unresolved: the slug
	// half landed, and the ambiguity is one mistake.
	if slug := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); len(slug) != 0 {
		t.Errorf("got %d reference/unresolved findings, want none: %+v", len(slug), slug)
	}
}

// A reference into a document whose headings were never recorded is
// unverified rather than resolved or broken: the tool says it could not
// look, which is not the same claim as looking and finding nothing.
//
// A `.feature` file is the concrete case. Its references are extracted
// against its own path, and that path is a feature source rather than a
// corpus member, so it appears in no index entry and has no recorded
// headings for an anchor naming no slug to be matched against.
//
// AC: AC-SPECANCH-002.4
func TestAnAnchorWhoseTargetHasNoRecordedHeadingsIsUnverified(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		strings.Replace(gherkinSource, "A fixture feature.", "A fixture feature, refining [[#Decision]].", 1))

	got := anchorFindingsOf(checkedAnchors(t, repo))

	if len(got) != 1 {
		t.Fatalf("got %d anchor findings, want 1: %+v", len(got), got)
	}
	if got[0].Severity != finding.SeverityUnverified {
		t.Fatalf("severity = %q, want unverified", got[0].Severity)
	}
	if got[0].Rule != finding.RuleUnresolvedAnchor {
		t.Errorf("rule = %q, want the anchor check's own identifier", got[0].Rule)
	}
	if got[0].Path != "specs/features/checkout.feature" {
		t.Errorf("path = %q, want the referring file", got[0].Path)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnverified)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnverified", got[0].Detail.Value)
	}
	if !strings.Contains(detail.Reason, "specs/features/checkout.feature") {
		t.Errorf("reason = %q, want it to name what is missing", detail.Reason)
	}
}

// A reference whose slug resolves to no corpus member produces the
// parent feature's unresolved-reference finding and no anchor finding on
// top of it: two findings for one mistake trains a reader to skim.
//
// AC: AC-SPECANCH-002.8
func TestAReferenceWhoseSlugResolvesToNothingProducesNoAnchorFinding(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-099#Decision]]."))

	findings := checkedAnchors(t, repo)

	if got := anchorFindingsOf(findings); len(got) != 0 {
		t.Errorf("got %d anchor findings, want none: %+v", len(got), got)
	}
	if got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want exactly 1: %+v", len(got), findings)
	}
}

// Every reference carrying an anchor is checked by an ordinary corpus
// check, not behind an option, and the two ways an anchor can fail are
// reported under stable rule identifiers of their own — which is what
// lets a caller select either one without selecting slug resolution.
//
// AC: AC-SPECANCH-002.9
func TestEveryAnchoredReferenceIsCheckedUnderItsOwnRuleIdentifier(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-002.md",
		sectioned("adr", "adr-002", "Second", "## Decision", "## References", "### references"))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First",
		"It refines [[adr-002#Decision]], revisits [[adr-002#Rationale]],\n"+
			"cites [[adr-002#References]], and embeds ![[adr-002#Consequences]].",
	))

	got := anchorFindingsOf(checkedAnchors(t, repo))

	// Four anchored references, one of which resolves; the embed is
	// checked exactly as the equivalent link would be.
	rules := map[specmodel.RuleId]int{}
	for _, f := range got {
		rules[f.Rule]++
	}
	if rules[finding.RuleUnresolvedAnchor] != 2 {
		t.Errorf("got %d reference/unresolved-anchor findings, want 2 (#Rationale and the #Consequences embed): %+v",
			rules[finding.RuleUnresolvedAnchor], got)
	}
	if rules[finding.RuleAmbiguousAnchor] != 1 {
		t.Errorf("got %d reference/ambiguous-anchor findings, want 1 (#References): %+v",
			rules[finding.RuleAmbiguousAnchor], got)
	}
	if len(got) != 3 {
		t.Errorf("got %d anchor findings, want 3 — #Decision resolves and is not reported: %+v", len(got), got)
	}
}

// A single-document check on the document whose heading was renamed
// reports no anchor finding located in the referring document: it covers
// the named document only, and what the rename broke elsewhere is a
// corpus check's to report.
//
// AC-SPECANCH-003.2 is cited at the integration level it is declared at,
// by the snapshot case check-path-anchor-elsewhere. This is its
// unit-level companion, which is why it carries no citation of its own.
func TestCheckForReportsNoAnchorFindingLocatedElsewhere(t *testing.T) {
	repo := t.TempDir()
	// The heading the referrer names has been renamed to "Outcome".
	write(t, repo, "specs/adrs/adr-002.md", sectioned("adr", "adr-002", "Second", "## Outcome"))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002#Decision]]."))

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))
	findings, err := built.CheckFor("specs/adrs/adr-002.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	for _, f := range findings {
		if f.Path != "specs/adrs/adr-002.md" {
			t.Errorf("a check of adr-002 reported a finding located in %s: %s %s", f.Path, f.Rule, f.Message)
		}
	}
}

// The same broken anchor is reported by a check of the document that
// holds it, at the span of the reference.
func TestCheckForReportsTheNamedDocumentsOwnBrokenAnchor(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-002.md", sectioned("adr", "adr-002", "Second", "## Outcome"))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002#Decision]]."))

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))
	findings, err := built.CheckFor("specs/adrs/adr-001.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}

	got := anchorFindingsOf(findings)
	if len(got) != 1 {
		t.Fatalf("got %d anchor findings for the referring document, want 1: %+v", len(got), findings)
	}
	if got[0].Path != "specs/adrs/adr-001.md" {
		t.Errorf("path = %q, want the referring document", got[0].Path)
	}
	if got[0].Span.Length == 0 {
		t.Error("finding carries an empty span")
	}
}

// An anchor written as a path of headings resolves through its enclosing
// headings: it selects the heading named by its last segment beneath each
// earlier one, so it can name one of two same-named headings that a
// single-segment anchor finds ambiguous. A path no heading satisfies is
// unresolved, and one that several satisfy is ambiguous, each match named
// by its full heading path and line.
//
// AC: AC-SPECANCH-002.10, AC-SPECANCH-002.5
func TestANestedAnchorResolvesThroughItsEnclosingHeadings(t *testing.T) {
	target := sectioned("adr", "adr-002", "Second",
		"## Components", "### References", "### Findings", "#### Detail", "## References")

	cases := []struct {
		name   string
		anchor string
		rule   specmodel.RuleId
		names  []string
	}{
		{"a path naming one of two same-named headings resolves", "Components#References", "", nil},
		{"an intermediate ancestor may be left out", "Second#Detail", "", nil},
		{"the full heading path resolves", "second#components#references", "", nil},
		{"a path no heading satisfies is unresolved", "Components#Nope", finding.RuleUnresolvedAnchor, nil},
		{"out-of-order segments are unresolved", "References#Components", finding.RuleUnresolvedAnchor, nil},
		{"a path several headings satisfy is ambiguous", "Second#References", finding.RuleAmbiguousAnchor,
			[]string{"second#components#references (line ", "second#references (line "}},
		{"a single segment several match names every heading path", "References", finding.RuleAmbiguousAnchor,
			[]string{"second#components#references (line ", "second#references (line "}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := t.TempDir()
			write(t, repo, "specs/adrs/adr-002.md", target)
			write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "See [[adr-002#"+c.anchor+"]]."))

			got := anchorFindingsOf(checkedAnchors(t, repo))
			if c.rule == "" {
				if len(got) != 0 {
					t.Fatalf("[[adr-002#%s]] raised %+v, want it to resolve", c.anchor, got)
				}
				return
			}
			if len(got) != 1 || got[0].Rule != c.rule {
				t.Fatalf("[[adr-002#%s]] raised %+v, want one %s", c.anchor, got, c.rule)
			}
			for _, name := range c.names {
				if !strings.Contains(got[0].Message, name) {
					t.Errorf("message %q does not name %q", got[0].Message, name)
				}
			}
			if c.rule == finding.RuleAmbiguousAnchor {
				detail, _ := got[0].Detail.AsAmbiguousAnchor()
				for _, match := range detail.Matches {
					if !strings.HasPrefix(match.HeadingPath, "second#") {
						t.Errorf("match %+v carries no heading path", match)
					}
				}
			}
		})
	}
}
