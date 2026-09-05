// Tests for target resolution and the two reachability findings.
//
// Every case builds its [specmodel.Reference] fixtures directly rather than
// through [references.Extract] and [references.Resolve]: this package
// consumes a reference's Resolution as already-settled input, so a fixture
// only needs to state that outcome, not reproduce the corpus that would
// produce it.
package mdlinks_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/mdlinks"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// at is a stand-in position, mirroring cli/core/finding's own tests: this
// package neither derives nor validates a span, so a fixed extent serves
// every case.
func at(t *testing.T) specmodel.Span {
	t.Helper()
	return span.Of([]byte("The decision is [[adr-012]], refining it."), 17, 7)
}

// resolvedRef builds a reference whose slug reached a corpus member at
// targetPath.
func resolvedRef(t *testing.T, citingPath, slug, targetPath string, anchor *string) specmodel.Reference {
	t.Helper()
	return specmodel.Reference{
		Path:     citingPath,
		Text:     slug,
		Slug:     slug,
		Anchor:   anchor,
		Position: at(t),
		Resolution: specmodel.NewReferenceResolutionResolved(
			specmodel.ReferenceResolved{TargetPath: targetPath}),
	}
}

// anchorResolvedTo builds the anchor resolution of a reference whose
// anchor matched exactly one heading, whose text is heading.
func anchorResolvedTo(t *testing.T, heading string) *specmodel.AnchorResolution {
	t.Helper()
	resolution := specmodel.NewAnchorResolutionResolved(specmodel.AnchorResolved{
		Heading: specmodel.AnchorMatch{Text: heading, Position: at(t)},
	})
	return &resolution
}

// unresolvedRef builds a reference whose slug reached nothing at all.
func unresolvedRef(t *testing.T, citingPath, slug string) specmodel.Reference {
	t.Helper()
	return specmodel.Reference{
		Path:       citingPath,
		Text:       slug,
		Slug:       slug,
		Position:   at(t),
		Resolution: specmodel.NewReferenceResolutionUnresolved(specmodel.ReferenceUnresolved{}),
	}
}

// nonMemberRef builds a reference whose slug matched a file the index
// recorded as excluded rather than as a corpus member — reachable but not
// a corpus document, per this package's own vocabulary.
func nonMemberRef(t *testing.T, citingPath, slug, matchedPath string) specmodel.Reference {
	t.Helper()
	reason := specmodel.ExclusionReasonNoFrontmatter
	return specmodel.Reference{
		Path:     citingPath,
		Text:     slug,
		Slug:     slug,
		Position: at(t),
		Resolution: specmodel.NewReferenceResolutionNonMember(
			specmodel.ReferenceNonMember{Path: matchedPath, Reason: &reason}),
	}
}

// AC: AC-SPECMD-005.5
func TestAMappedSlugIsUsedExactlyAsHeld(t *testing.T) {
	ref := resolvedRef(t, "specs/adrs/adr-011.md", "template-authoring-standards",
		"specs/standards/template-authoring-standards.md", nil)

	targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{
		Map:     mdlinks.LinkMap{"template-authoring-standards": "../../weird//../path.md"},
		MapPath: "links.yaml",
	})

	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
	if targets[0] == nil || *targets[0] != "../../weird//../path.md" {
		t.Fatalf("target = %v, want the mapped path with no normalization", targets[0])
	}
}

// A mapped slug wins even when it does not resolve to a corpus document at
// all — the map is consulted first, per the delta's ordering.
func TestAMappedSlugWinsOverAnUnresolvedSlug(t *testing.T) {
	ref := unresolvedRef(t, "specs/adrs/adr-011.md", "external-doc")

	targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{
		Map:     mdlinks.LinkMap{"external-doc": "https://example.com/doc"},
		MapPath: "links.yaml",
	})

	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
	if targets[0] == nil || *targets[0] != "https://example.com/doc" {
		t.Fatalf("target = %v, want the mapped path", targets[0])
	}
}

// AC: AC-SPECMD-006.1, AC-SPECMD-006.2
func TestWithNoMapATargetIsDerivedRelativeToTheCitingDirectory(t *testing.T) {
	cases := []struct {
		name       string
		citingPath string
		targetPath string
		want       string
	}{
		{
			name:       "sibling in the same directory",
			citingPath: "specs/adrs/adr-011.md",
			targetPath: "specs/adrs/adr-012.md",
			want:       "adr-012.md",
		},
		{
			name:       "a different directory",
			citingPath: "specs/adrs/adr-011.md",
			targetPath: "specs/features/markdown-links/markdown-links-frd.md",
			want:       "../features/markdown-links/markdown-links-frd.md",
		},
		{
			name:       "citing document at the repository root",
			citingPath: "README.md",
			targetPath: "specs/adrs/adr-012.md",
			want:       "specs/adrs/adr-012.md",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ref := resolvedRef(t, c.citingPath, "adr-012", c.targetPath, nil)

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			if targets[0] == nil || *targets[0] != c.want {
				t.Fatalf("target = %v, want %q", targets[0], c.want)
			}
		})
	}
}

// AC: AC-SPECMD-006.3
func TestAGivenMapDerivesNothingForAnOmittedSlug(t *testing.T) {
	ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", nil)

	targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{
		Map:     mdlinks.LinkMap{"unrelated-slug": "unrelated.md"},
		MapPath: "links.yaml",
	})

	if targets[0] != nil {
		t.Fatalf("target = %q, want none — a given map must not fall back to derivation", *targets[0])
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %v, want exactly one", findings)
	}
	if findings[0].Rule != finding.RuleUnmappedCitation {
		t.Errorf("rule = %q, want %q", findings[0].Rule, finding.RuleUnmappedCitation)
	}
}

// AC: AC-SPECMD-002.1
func TestAnUnmappedCorpusDocumentRaisesAnErrorFinding(t *testing.T) {
	ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", nil)

	_, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{
		Map:     mdlinks.LinkMap{},
		MapPath: "links.yaml",
	})

	if len(findings) != 1 {
		t.Fatalf("findings = %v, want exactly one", findings)
	}
	f := findings[0]
	if f.Rule != finding.RuleUnmappedCitation {
		t.Errorf("rule = %q, want %q", f.Rule, finding.RuleUnmappedCitation)
	}
	if f.Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", f.Severity)
	}
	detail, ok := f.Detail.AsUnmappedCitation()
	if !ok {
		t.Fatalf("detail is not an unmapped citation: %#v", f.Detail.Value)
	}
	if detail.Slug != "adr-012" || detail.CitingPath != "specs/adrs/adr-011.md" || detail.MapPath != "links.yaml" {
		t.Errorf("detail = %#v", detail)
	}
}

// A slug naming no corpus document at all, uncovered by the map, raises a
// warning finding and resolves no target — the checklist item this bead
// owes with no cited AC number of its own.
func TestASlugNamingNoCorpusDocumentRaisesAWarningFinding(t *testing.T) {
	cases := []struct {
		name string
		ref  specmodel.Reference
	}{
		{name: "unresolved", ref: unresolvedRef(t, "specs/adrs/adr-011.md", "no-such-slug")},
		{name: "non-member", ref: nonMemberRef(t, "specs/adrs/adr-011.md", "no-such-slug", "specs/adrs/draft.md")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			targets, findings := mdlinks.Resolve([]specmodel.Reference{c.ref}, mdlinks.Options{})

			if targets[0] != nil {
				t.Fatalf("target = %q, want none", *targets[0])
			}
			if len(findings) != 1 {
				t.Fatalf("findings = %v, want exactly one", findings)
			}
			f := findings[0]
			if f.Rule != finding.RuleUnresolvedCitation {
				t.Errorf("rule = %q, want %q", f.Rule, finding.RuleUnresolvedCitation)
			}
			if f.Severity != finding.SeverityWarning {
				t.Errorf("severity = %q, want warning", f.Severity)
			}
			detail, ok := f.Detail.AsUnresolvedCitation()
			if !ok {
				t.Fatalf("detail is not an unresolved citation: %#v", f.Detail.Value)
			}
			if detail.Slug != "no-such-slug" || detail.CitingPath != "specs/adrs/adr-011.md" {
				t.Errorf("detail = %#v", detail)
			}
		})
	}
}

// AC: AC-SPECMD-002.5
func TestEveryFindingInARunIsReportedNotJustTheFirst(t *testing.T) {
	refs := []specmodel.Reference{
		resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", nil),  // unmapped
		unresolvedRef(t, "specs/adrs/adr-011.md", "no-such-slug"),                         // unresolved
		resolvedRef(t, "specs/adrs/adr-011.md", "mapped-one", "specs/adrs/other.md", nil), // clean
	}

	_, findings := mdlinks.Resolve(refs, mdlinks.Options{
		Map:     mdlinks.LinkMap{"mapped-one": "other.md"},
		MapPath: "links.yaml",
	})

	if len(findings) != 2 {
		t.Fatalf("findings = %v, want exactly two — one per broken reference", findings)
	}
	rules := map[specmodel.RuleId]bool{}
	for _, f := range findings {
		rules[f.Rule] = true
	}
	if !rules[finding.RuleUnmappedCitation] || !rules[finding.RuleUnresolvedCitation] {
		t.Errorf("rules raised = %v, want both links/unmapped and links/unresolved", rules)
	}
}

// An anchor that reached no heading still yields a fragment, derived
// from the anchor as the author wrote it: the rewrite stays total, and
// the copy names the section the source named rather than naming
// nothing at all.
//
// AC: AC-SPECMD-001.7
func TestAnAnchorReachingNoHeadingYieldsAFragmentFromTheAnchorAsAuthored(t *testing.T) {
	cases := []struct {
		anchor string
		want   string
	}{
		{anchor: "Simple Heading", want: "simple-heading"},
		{anchor: "Punctuation, and (Parens)!", want: "punctuation-and-parens"},
		{anchor: "Already-Hyphenated_and_Underscored", want: "already-hyphenated_and_underscored"},
		{anchor: "Multiple   Spaces", want: "multiple---spaces"},
	}

	for _, c := range cases {
		t.Run(c.anchor, func(t *testing.T) {
			anchor := c.anchor
			ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", &anchor)

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			want := "adr-012.md#" + c.want
			if targets[0] == nil || *targets[0] != want {
				t.Fatalf("target = %v, want %q", targets[0], want)
			}
		})
	}
}

// A reference whose anchor resolved to a heading takes its fragment from
// that heading's text, not from the anchor's.
//
// Every fixture below states a heading whose text is NOT the anchor's,
// which the matcher that settles a resolution would never produce: it
// resolves an anchor by comparing the two identifiers, so under today's
// matching rule the heading and the anchor always derive alike and no
// corpus input can tell the two sources apart. Stating a divergent
// heading is what makes the seam observable — and it is a fair fixture
// because this package is a CONSUMER of a settled resolution and never
// re-derives or re-checks the match. The distinction is the point of the
// criterion: the emitted link has to land on the heading a reader is
// being sent to, so a looser matcher — one that admits a leading
// requirement code, or a heading named in part — cannot silently start
// emitting fragments that reach nothing.
//
// AC: AC-SPECMD-001.3
func TestAResolvedAnchorTakesItsFragmentFromTheResolvedHeading(t *testing.T) {
	cases := []struct {
		name    string
		anchor  string
		heading string
		want    string
	}{
		{
			name:    "the anchor names the heading in full and the two agree",
			anchor:  "Request Router",
			heading: "Request Router",
			want:    "request-router",
		},
		{
			name:    "the heading carries markup the anchor does not",
			anchor:  "The Profile interface",
			heading: "The `Profile` interface",
			want:    "the-profile-interface",
		},
		{
			name:    "the heading is spelled differently from the anchor",
			anchor:  "Reference Integrity Checks",
			heading: "FEAT-SPECREF-002: Reference Integrity Checks",
			want:    "feat-specref-002-reference-integrity-checks",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			anchor := c.anchor
			ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", &anchor)
			ref.AnchorResolution = anchorResolvedTo(t, c.heading)

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			want := "adr-012.md#" + c.want
			if targets[0] == nil || *targets[0] != want {
				t.Fatalf("target = %v, want %q", targets[0], want)
			}
		})
	}
}

// Every anchor resolution that is not a resolved one falls back to the
// anchor as authored — a considered "no heading", a collision the
// resolver refused to pick between, and a target whose headings were
// never recorded alike. None of the three names a heading to derive
// from, so none can supply one.
//
// AC: AC-SPECMD-001.7
func TestAnAnchorResolutionThatNamesNoHeadingFallsBackToTheAnchor(t *testing.T) {
	unresolved := specmodel.NewAnchorResolutionUnresolved(specmodel.AnchorUnresolved{})
	ambiguous := specmodel.NewAnchorResolutionAmbiguous(specmodel.AnchorAmbiguous{
		Matches: []specmodel.AnchorMatch{
			{Text: "Overview Of The Thing", Position: at(t)},
			{Text: "Overview of the thing", Position: at(t)},
		},
	})
	unverified := specmodel.NewAnchorResolutionUnverified(specmodel.AnchorUnverified{
		Reason: "no headings are recorded for specs/adrs/adr-012.md",
	})

	cases := []struct {
		name       string
		resolution *specmodel.AnchorResolution
	}{
		{name: "no heading matched", resolution: &unresolved},
		// Ambiguous carries matches, and it is still the anchor that the
		// fragment comes from: the resolver declined to choose one, and
		// this package must not choose for it.
		{name: "several headings matched", resolution: &ambiguous},
		{name: "the target's headings were never recorded", resolution: &unverified},
		{name: "no resolution was computed at all", resolution: nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			anchor := "Overview Of The Thing"
			ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", &anchor)
			ref.AnchorResolution = c.resolution

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			const want = "adr-012.md#overview-of-the-thing"
			if targets[0] == nil || *targets[0] != want {
				t.Fatalf("target = %v, want %q", targets[0], want)
			}
		})
	}
}

// A mapped slug's target gains the same fragment a derived one does: the
// fragment is a property of the reference's anchor rather than of how
// its path was settled.
//
// AC: AC-SPECMD-001.3
func TestAMappedTargetCarriesTheResolvedHeadingsFragmentToo(t *testing.T) {
	anchor := "Request Router"
	ref := resolvedRef(t, "specs/adrs/adr-011.md", "container-api", "specs/containers/container-api.md", &anchor)
	ref.AnchorResolution = anchorResolvedTo(t, "The `Request` Router")

	targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{
		Map:     mdlinks.LinkMap{"container-api": "container-api.md"},
		MapPath: "links.yaml",
	})

	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
	const want = "container-api.md#the-request-router"
	if targets[0] == nil || *targets[0] != want {
		t.Fatalf("target = %v, want %q", targets[0], want)
	}
}

// A citing document that is not itself a corpus member resolves exactly
// as one that is: the citing path is taken from the caller and used as
// given, never looked up. The checklist item this bead owes with no cited
// AC number of its own.
func TestACitingDocumentOutsideTheCorpusResolvesTheSameWay(t *testing.T) {
	ref := resolvedRef(t, "notes/scratch/idea.md", "adr-012", "specs/adrs/adr-012.md", nil)

	targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
	want := "../../specs/adrs/adr-012.md"
	if targets[0] == nil || *targets[0] != want {
		t.Fatalf("target = %v, want %q", targets[0], want)
	}
}
