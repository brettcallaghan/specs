package corpus

import (
	"errors"
	"regexp"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// featureExtension is the extension a feature source carries.
//
// Matched without regard to letter case for the same reason
// [candidateExtension] is: a corpus checked out on a case-preserving
// filesystem must yield the same features as the same corpus on a
// case-sensitive one.
//
// A file matching it is *not* a corpus candidate. It carries no
// frontmatter and is no document, so it produces neither an
// [specmodel.IndexEntry] nor an [specmodel.IndexExclusion]: AC-SPEC-001.6's
// silent pass-over of a non-`.md` file stands, and what this feature adds
// is a second kind of source the same walk offers, never an exception to
// corpus membership.
const featureExtension = ".feature"

// SourceFailure is one Gherkin source that would not parse, named by the
// file it was read from.
//
// A failure is carried rather than raised here because indexing never
// fails as a whole (AC-SPEC-002.4): the source is recorded against its own
// file, the walk continues, and the checks that turn a failure into a
// reported finding read it from the built corpus.
type SourceFailure struct {
	// Path is the file the source was read from, relative to the
	// repository root — the containing document for a fenced block, and
	// the feature source itself for a `.feature` file.
	Path string

	// Errors is every failure the Gherkin parser reported, in its order,
	// each located in the coordinates of Path.
	Errors []specmodel.ParseError
}

// vocabulary is a profile's [FeatureTagRules] compiled once per build.
//
// This is the seam between the registered patterns and the Gherkin
// component, and the reason that component reads tags as bare strings: it
// resolves no pattern of its own, so everything the profile registers
// meets it here and nowhere below.
//
// A pattern that does not compile leaves its half of the vocabulary
// unresolved rather than failing the build. [ValidateRegistration] is
// where a malformed registration is reported; a build under one still
// indexes every document, and records the identifiers it could not
// resolve as absent.
type vocabulary struct {
	feature    *regexp.Regexp
	criterion  *regexp.Regexp
	owner      string
	qualifiers map[string]QualifierTag
}

// newVocabulary compiles one profile's registered feature tags.
func newVocabulary(rules FeatureTagRules) vocabulary {
	compiled := vocabulary{owner: rules.OwnerPrefix}
	if rules.FeaturePattern != "" {
		compiled.feature, _ = regexp.Compile(rules.FeaturePattern)
	}
	if rules.CriterionPattern != "" {
		compiled.criterion, _ = regexp.Compile(rules.CriterionPattern)
	}
	if len(rules.Qualifiers) > 0 {
		compiled.qualifiers = make(map[string]QualifierTag, len(rules.Qualifiers))
		for _, qualifier := range rules.Qualifiers {
			compiled.qualifiers[qualifier.Tag] = qualifier
		}
	}
	return compiled
}

// options returns the identity functions the Gherkin component parses
// under.
//
// Both are closures over the compiled patterns rather than values the
// component could have read for itself, which is what keeps that
// component free of every pattern a profile registers.
func (v vocabulary) options() gherkin.Options {
	return gherkin.Options{
		FeatureIdentity: func(tags []string) (specmodel.FeatureId, specmodel.FeatureCode, bool) {
			id, ok := matched(v.feature, tags)
			if !ok {
				return "", "", false
			}
			feature := specmodel.FeatureId(id)
			return feature, featureCodeOf(feature), true
		},
		CriterionIdentity: func(tags []string) (specmodel.CriterionId, bool) {
			id, ok := matched(v.criterion, tags)
			return specmodel.CriterionId(id), ok
		},
	}
}

// matched returns the first tag pattern matches.
//
// An element carrying more than one matching tag is a defect for a check
// to report, and a check can only report what the records carry: the
// first is recorded, every tag is kept on the parsed record, and nothing
// here judges the second.
func matched(pattern *regexp.Regexp, tags []string) (string, bool) {
	if pattern == nil {
		return "", false
	}
	for _, tag := range tags {
		if pattern.MatchString(tag) {
			return tag, true
		}
	}
	return "", false
}

// featureCodeOf returns the code a feature identifier groups its feature
// under, read from the identifier itself.
//
// The identifier's composition is the generated model's — a prefix, the
// group code, and the number — so the code is the segment before the
// number, and neither the pattern that matched nor the profile that
// registered it is consulted. Deriving it from a named capture group
// would mean the host knowing what a profile called that group, which is
// exactly the knowledge the registration exists to keep out.
//
// An identifier carrying no such segment yields no code. That is an
// identifier a profile's pattern admitted and this model has no grouping
// for, which is worth recording as empty rather than guessing at.
func featureCodeOf(id specmodel.FeatureId) specmodel.FeatureCode {
	segments := strings.Split(string(id), "-")
	if len(segments) < 3 {
		return ""
	}
	return specmodel.FeatureCode(segments[len(segments)-2])
}

// owningSlug returns the document slug a feature's tags name, for a
// feature that has to name one.
//
// The tag *is* the slug; the registered prefix is only how that tag is
// told apart from every other tag on the feature. A feature carrying none
// is attributed to nothing here and reported by the check that requires
// one (AC-GHK-002.10).
func (v vocabulary) owningSlug(tags []string) specmodel.Slug {
	if v.owner == "" {
		return ""
	}
	for _, tag := range tags {
		if strings.HasPrefix(tag, v.owner) {
			return specmodel.Slug(tag)
		}
	}
	return ""
}

// qualify populates the criterion properties a profile's qualifier tags
// carry. A tag the profile registers nothing for reaches no property,
// which is what makes an unregistered tag one the vocabulary says nothing
// about (AC-GHK-002.8).
func (v vocabulary) qualify(record *specmodel.Criterion, tags []string) {
	for _, tag := range tags {
		qualifier, registered := v.qualifiers[tag]
		if !registered {
			continue
		}
		switch qualifier.Meaning {
		case specmodel.QualifierMeaningNegation:
			record.Negative = true
		case specmodel.QualifierMeaningManualVerification:
			record.Manual = true
		case specmodel.QualifierMeaningVerificationLevel:
			level := qualifier.Level
			record.Level = &level
		}
	}
}

// extractBlockFeatures records the features written in one member's
// fenced Gherkin blocks, from the bytes the walk has already read.
//
// Ownership needs no tag: a fenced block is written inside the document
// that owns it, so the owning slug is the one that document declares
// (AC-GHK-001.7). It is read from the parsed frontmatter rather than from
// the typed envelope, because the envelope's slug is a validated value
// and a document whose slug is absent or malformed still holds features
// that have to be attributed to what it actually wrote.
//
// The whole of src is offered to the fenced-block lookup, frontmatter
// included, so that every span a block yields is already in the
// document's own coordinates and needs no second arithmetic to become
// one.
func (b *builder) extractBlockFeatures(reported string, src []byte, doc *frontmatter.Document) {
	slug := declaredSlug(doc)
	for _, block := range markdown.Fenced(src, gherkin.FenceTags()) {
		parsed, err := gherkin.ParseRegion(src, block.Inner, b.vocabulary.options())
		if err != nil {
			b.recordFailure(reported, err)
			continue
		}
		if parsed == nil {
			continue
		}
		source := specmodel.NewFeatureSourceFencedBlock(specmodel.FeatureSourceBlock{Block: block.Outer})
		b.addFeature(reported, slug, source, parsed)
	}
}

// visitFeatureSource records the feature one `.feature` file holds.
//
// The file is the source, so the whole of it is parsed. Ownership is the
// one fact this form establishes differently: having no document around
// it to inherit from, a feature source names its owner in a tag
// (AC-GHK-001.2), and discovery therefore needs no location of its own —
// a feature source sits wherever the documents of its feature sit
// (AC-GHK-001.6).
func (b *builder) visitFeatureSource(reported string, src []byte, readErr error) {
	if readErr != nil {
		b.recordFailure(reported, readErr)
		return
	}
	parsed, err := gherkin.Parse(src, b.vocabulary.options())
	if err != nil {
		b.recordFailure(reported, err)
		return
	}
	if parsed == nil {
		return
	}
	source := specmodel.NewFeatureSourceFeatureFile(specmodel.FeatureSourceFile{})
	b.addFeature(reported, b.vocabulary.owningSlug(parsed.Tags), source, parsed)
}

// addFeature completes one parsed feature with the facts Gherkin does not
// establish — where it was written, which form it was written in, and
// which document owns it — and records it.
func (b *builder) addFeature(
	reported string,
	slug specmodel.Slug,
	source specmodel.FeatureSource,
	parsed *gherkin.Feature,
) {
	record := parsed.Record
	record.Path = reported
	record.DocumentSlug = slug
	record.Source = source
	record.Criteria = make([]specmodel.Criterion, 0, len(parsed.Criteria))
	for _, criterion := range parsed.Criteria {
		nested := criterion.Record
		b.vocabulary.qualify(&nested, criterion.Tags)
		record.Criteria = append(record.Criteria, nested)
	}
	b.features = append(b.features, record)
}

// recordFailure records a source that would not parse against the file it
// was read from, so that the walk continues and the failure is still
// answerable (AC-SPEC-002.4).
func (b *builder) recordFailure(reported string, err error) {
	failure := SourceFailure{Path: reported}
	var parse *gherkin.ParseError
	if errors.As(err, &parse) {
		failure.Errors = parse.Errors
	} else {
		failure.Errors = []specmodel.ParseError{{
			Message:  err.Error(),
			Position: span.Of(nil, 0, 0),
		}}
	}
	b.failures = append(b.failures, failure)
}

// FeatureFailures returns every Gherkin source that would not parse, in
// the order the walk met them.
//
// Empty for a corpus served warm from a persisted index: a parse failure
// is not on the index, and a warm corpus has parsed nothing. That is the
// same shape a duplicate-slug finding has, and the checks that report
// these recompute them for a warm corpus exactly as they do there.
func (c *Corpus) FeatureFailures() []SourceFailure {
	return c.failures
}
