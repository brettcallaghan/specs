// Tests for feature-source discovery: the two admissible forms, the
// ownership each of them establishes, and the input hash that keeps a
// persisted index from outliving either.
package corpus_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// featureTags is the tag vocabulary every build in this file runs under.
//
// The patterns mirror the shape a real registration has — an identifier
// of a prefix, a group code and a number — because the code recorded on a
// feature is derived from a matched identifier rather than supplied, and
// a fixture pattern admitting some other shape would not exercise that.
func featureTags() corpus.FeatureTagRules {
	return corpus.FeatureTagRules{
		FeaturePattern:   `^FEAT-(?P<code>[A-Z0-9]+)-(?P<number>[0-9]{3})$`,
		CriterionPattern: `^AC-(?P<code>[A-Z0-9]+)-(?P<number>[0-9]{3})\.[0-9]+$`,
		OwnerPrefix:      "frd-",
		SharedCaptures:   []string{"code", "number"},
		Qualifiers: []corpus.QualifierTag{
			{Tag: "negative", Meaning: specmodel.QualifierMeaningNegation},
			{Tag: "manual", Meaning: specmodel.QualifierMeaningManualVerification},
			{
				Tag:     "level-unit",
				Meaning: specmodel.QualifierMeaningVerificationLevel,
				Level:   specmodel.VerificationLevelUnit,
			},
		},
	}
}

// featureProfile is indexProfile with that vocabulary registered.
func featureProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {SlugPattern: `^adr-[0-9]{3}$`, Numbered: true},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
		featureTags: featureTags(),
	}
}

// featureOptions is the build options every fixture here uses: repo as
// the repository root, repo/specs as the spec root.
func featureOptions(repo string, profile corpus.Profile) corpus.Options {
	return corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   profile,
	}
}

func featureCorpusAt(t *testing.T, repo string) *corpus.Corpus {
	t.Helper()

	built, err := corpus.Build(featureOptions(repo, featureProfile()))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// gherkinSource is one feature, written as the bare Gherkin both forms
// carry: a `.feature` file holds it as written, and a fenced block holds
// it between two fence lines.
const gherkinSource = `@FEAT-CHK-001 @frd-checkout
Feature: Checkout

  A fixture feature.

  @AC-CHK-001.1 @negative @level-unit
  Scenario: A basket with nothing in it cannot be paid for
    Given an empty basket
    Then payment is refused
`

// blockDocument renders a document of the given type carrying source in a
// fenced Gherkin block.
func blockDocument(docType, slug, title, source string) string {
	return document(docType, slug, title) + "\n```" + "gherkin\n" + source + "```\n"
}

// featureNamed returns the one feature the index carries for path.
func featureNamed(t *testing.T, index specmodel.SpecIndex, path string) specmodel.Feature {
	t.Helper()

	for _, feature := range index.Features {
		if feature.Path == path {
			return feature
		}
	}
	t.Fatalf("no feature recorded for %s; features are %v", path, featurePaths(index))
	return specmodel.Feature{}
}

func featurePaths(index specmodel.SpecIndex) []string {
	out := make([]string, 0, len(index.Features))
	for _, feature := range index.Features {
		out = append(out, feature.Path)
	}
	return out
}

// A `.feature` file beneath the spec root is discovered, parsed whole,
// and attributed to the document its owning-document tag names.
//
// AC: AC-GHK-001.2
func TestAFeatureFileBeneathTheSpecRootIsAFeatureSource(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", gherkinSource)

	index := featureCorpusAt(t, repo).Index()

	feature := featureNamed(t, index, "specs/features/checkout.feature")
	if feature.Id != "FEAT-CHK-001" {
		t.Errorf("Id = %q, want %q", feature.Id, "FEAT-CHK-001")
	}
	if feature.Code != "CHK" {
		t.Errorf("Code = %q, want %q, derived from the matched identifier", feature.Code, "CHK")
	}
	if feature.DocumentSlug != "frd-checkout" {
		t.Errorf("DocumentSlug = %q, want the slug its tag names", feature.DocumentSlug)
	}
	if _, ok := feature.Source.AsFeatureFile(); !ok {
		t.Errorf("Source = %#v, want the feature-file form", feature.Source)
	}
	if len(feature.Criteria) != 1 {
		t.Fatalf("Criteria has %d entries, want 1", len(feature.Criteria))
	}
	if feature.Criteria[0].Id != "AC-CHK-001.1" {
		t.Errorf("criterion Id = %q, want %q", feature.Criteria[0].Id, "AC-CHK-001.1")
	}
}

// Discovery searches the whole spec root, so a feature source needs no
// location of its own: it sits with the documents of its feature and is
// found there.
//
// AC: AC-GHK-001.6
func TestAFeatureSourceIsFoundWhereverBeneathTheSpecRootItSits(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout/deeply/nested.feature", gherkinSource)

	index := featureCorpusAt(t, repo).Index()

	feature := featureNamed(t, index, "specs/features/checkout/deeply/nested.feature")
	if feature.DocumentSlug != "frd-checkout" {
		t.Errorf("DocumentSlug = %q, want the slug its tag names", feature.DocumentSlug)
	}
}

// A fenced block is owned by the document it is written in, and needs no
// tag to say so.
//
// AC: AC-GHK-001.7
func TestAFencedBlockIsOwnedByTheDocumentItIsWrittenIn(t *testing.T) {
	repo := t.TempDir()
	untagged := strings.Replace(gherkinSource, " @frd-checkout", "", 1)
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout", untagged))

	index := featureCorpusAt(t, repo).Index()

	feature := featureNamed(t, index, "specs/features/frd-checkout.md")
	if feature.DocumentSlug != "frd-checkout" {
		t.Errorf("DocumentSlug = %q, want the containing document's slug", feature.DocumentSlug)
	}
	block, ok := feature.Source.AsFencedBlock()
	if !ok {
		t.Fatalf("Source = %#v, want the fenced-block form", feature.Source)
	}
	if block.Block.Offset >= feature.Position.Offset {
		t.Errorf("the fence's outer extent begins at %d and the Gherkin at %d; the fence must open first",
			block.Block.Offset, feature.Position.Offset)
	}
}

// A feature source is no corpus candidate: it produces neither an entry
// nor an exclusion, so AC-SPEC-001.6's silent pass-over of a non-`.md`
// file stands.
//
// AC: AC-GHK-001.2
func TestAFeatureSourceIsNeitherAnEntryNorAnExclusion(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", gherkinSource)

	index := featureCorpusAt(t, repo).Index()

	const source = "specs/features/checkout.feature"
	if contains(paths(index), source) {
		t.Errorf("entries = %v, want no entry for the feature source", paths(index))
	}
	if contains(excludedPaths(index), source) {
		t.Errorf("exclusions = %v, want no exclusion for the feature source", excludedPaths(index))
	}
}

// The qualifier tags a profile registers reach the criterion properties
// they mean, and a tag it registers nothing for reaches none.
//
// AC: AC-GHK-005.8
func TestRegisteredQualifiersReachTheCriterionRecord(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		strings.Replace(gherkinSource, "@level-unit", "@level-unit @wip", 1))

	index := featureCorpusAt(t, repo).Index()

	criterion := featureNamed(t, index, "specs/features/checkout.feature").Criteria[0]
	if !criterion.Negative {
		t.Error("Negative = false, want the registered negation tag to have set it")
	}
	if criterion.Manual {
		t.Error("Manual = true, want false: no tag on this criterion means manual verification")
	}
	if criterion.Level == nil || *criterion.Level != specmodel.VerificationLevelUnit {
		t.Errorf("Level = %v, want the level the registered tag records", criterion.Level)
	}
}

// Both forms produce equal records but for their source location and
// their source form, which is what makes the choice between them an
// authoring one.
//
// AC: AC-GHK-001.3
func TestBothFormsProduceEqualRecords(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout", gherkinSource))
	write(t, repo, "specs/features/checkout.feature", gherkinSource)

	index := featureCorpusAt(t, repo).Index()

	block := featureNamed(t, index, "specs/features/frd-checkout.md")
	file := featureNamed(t, index, "specs/features/checkout.feature")

	if block.Id != file.Id || block.Code != file.Code || block.Name != file.Name {
		t.Errorf("identity differs: block %q/%q/%q, file %q/%q/%q",
			block.Id, block.Code, block.Name, file.Id, file.Code, file.Name)
	}
	if block.DocumentSlug != file.DocumentSlug {
		t.Errorf("DocumentSlug differs: block %q, file %q", block.DocumentSlug, file.DocumentSlug)
	}
	if len(block.Criteria) != 1 || len(file.Criteria) != 1 {
		t.Fatalf("criteria counts differ: block %d, file %d", len(block.Criteria), len(file.Criteria))
	}
	if block.Criteria[0].ContentHash != file.Criteria[0].ContentHash {
		t.Errorf("content hashes differ: block %q, file %q",
			block.Criteria[0].ContentHash, file.Criteria[0].ContentHash)
	}
}

// Features are ordered by the path they are written in and then by their
// position within that file, so the emitted index is a function of corpus
// content rather than of the filesystem's iteration order.
//
// AC: AC-SPEC-009.7
func TestFeaturesAreOrderedByPathThenPosition(t *testing.T) {
	repo := t.TempDir()
	second := strings.NewReplacer("CHK-001", "CHK-002", "Checkout", "Refunds").Replace(gherkinSource)
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout", gherkinSource)+
			"\n```"+"gherkin\n"+second+"```\n")
	write(t, repo, "specs/adrs/adr-001.md",
		blockDocument("adr", "adr-001", "First", gherkinSource))

	index := featureCorpusAt(t, repo).Index()

	if got := featurePaths(index); !equal(got, []string{
		"specs/adrs/adr-001.md",
		"specs/features/frd-checkout.md",
		"specs/features/frd-checkout.md",
	}) {
		t.Fatalf("feature paths = %v, want them ordered by path", got)
	}
	first, later := index.Features[1], index.Features[2]
	if first.Position.Offset >= later.Position.Offset {
		t.Errorf("features within one file are at offsets %d then %d, want ascending",
			first.Position.Offset, later.Position.Offset)
	}
}

// A source that will not parse is recorded against its own file and the
// walk continues: indexing never fails as a whole.
//
// AC: AC-SPEC-002.4
func TestASourceThatWillNotParseDoesNotStopTheWalk(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/broken.feature",
		"@FEAT-CHK-002\nFeature: Broken\n\n  @AC-CHK-002.1\n")
	write(t, repo, "specs/features/checkout.feature", gherkinSource)

	built := featureCorpusAt(t, repo)

	if _, ok := built.Entry("specs/features/frd-checkout.md"); !ok {
		t.Error("the document beside the unparsable source is not indexed")
	}
	featureNamed(t, built.Index(), "specs/features/checkout.feature")
	if contains(featurePaths(built.Index()), "specs/features/broken.feature") {
		t.Error("the unparsable source contributed a feature")
	}

	failures := built.FeatureFailures()
	if len(failures) != 1 {
		t.Fatalf("FeatureFailures = %v, want exactly the one unparsable source", failures)
	}
	if failures[0].Path != "specs/features/broken.feature" {
		t.Errorf("failure recorded against %q, want the source's own path", failures[0].Path)
	}
	if len(failures[0].Errors) == 0 {
		t.Error("the failure carries no located error")
	}
}

// Editing a feature source invalidates a persisted index.
//
// This is the property the whole warm-cache path rests on for features: a
// hash that covered only `.md` candidates would serve an index built
// before the edit, which is a silent wrong answer to every query over
// features rather than a visible failure.
//
// AC: AC-SPEC-010.6
func TestEditingAFeatureSourceChangesTheInputHash(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	source := write(t, repo, "specs/features/checkout.feature", gherkinSource)
	opts := featureOptions(repo, featureProfile())

	before := fingerprint(t, opts)
	if err := os.WriteFile(source, []byte(strings.Replace(
		gherkinSource, "Then payment is refused", "Then payment is declined", 1)), 0o644); err != nil {
		t.Fatalf("rewriting the feature source: %v", err)
	}
	after := fingerprint(t, opts)

	if before == after {
		t.Errorf("the input hash is %s both before and after the edit", before)
	}
}

// Adding a feature source invalidates a persisted index, and the hash the
// build records is the one Fingerprint computes — both are fed by the one
// walk.
//
// AC: AC-SPEC-010.5
func TestTheBuildRecordsTheHashFingerprintComputesOverBothKindsOfSource(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	opts := featureOptions(repo, featureProfile())

	without := fingerprint(t, opts)
	write(t, repo, "specs/features/checkout.feature", gherkinSource)
	with := fingerprint(t, opts)

	if without == with {
		t.Errorf("the input hash is %s with and without a feature source", with)
	}

	built, err := corpus.Build(opts)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	recorded := built.Index().InputHash
	if recorded == nil || *recorded != with {
		t.Errorf("the built index records %v, want the fingerprint %s", recorded, with)
	}
}

// A changed feature-tag registration invalidates a persisted index, for
// the same reason a changed reference-bearing property does: the index's
// content moves while every file on disk stands still.
//
// AC: AC-SPEC-010.6
func TestChangingTheRegisteredFeatureTagsChangesTheInputHash(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/checkout.feature", gherkinSource)

	registered := featureProfile()
	before := fingerprint(t, featureOptions(repo, registered))

	for name, changed := range map[string]corpus.FeatureTagRules{
		"a widened feature pattern": {
			FeaturePattern:   `^[A-Z]+-[A-Z0-9]+-[0-9]{3}$`,
			CriterionPattern: featureTags().CriterionPattern,
			OwnerPrefix:      featureTags().OwnerPrefix,
			Qualifiers:       featureTags().Qualifiers,
		},
		"another owner prefix": {
			FeaturePattern:   featureTags().FeaturePattern,
			CriterionPattern: featureTags().CriterionPattern,
			OwnerPrefix:      "owner-",
			Qualifiers:       featureTags().Qualifiers,
		},
		"one fewer qualifier": {
			FeaturePattern:   featureTags().FeaturePattern,
			CriterionPattern: featureTags().CriterionPattern,
			OwnerPrefix:      featureTags().OwnerPrefix,
			Qualifiers:       featureTags().Qualifiers[:1],
		},
	} {
		t.Run(name, func(t *testing.T) {
			profile := stubProfile{
				schema:      []byte(`{"definitions":{}}`),
				types:       registered.DocumentTypes(),
				featureTags: changed,
			}
			if after := fingerprint(t, featureOptions(repo, profile)); after == before {
				t.Errorf("the input hash is %s under both registrations", after)
			}
		})
	}
}

// fingerprint is Fingerprint with its error handled once.
func fingerprint(t *testing.T, opts corpus.Options) string {
	t.Helper()

	hash, err := corpus.Fingerprint(opts)
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	return hash
}
