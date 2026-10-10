package corpus

import (
	"errors"
	"regexp"
	"strings"

	"github.com/brettcallaghan/specs/cli/core/frontmatter"
	"github.com/brettcallaghan/specs/cli/core/gherkin"
	"github.com/brettcallaghan/specs/cli/core/markdown"
	"github.com/brettcallaghan/specs/cli/core/references"
	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
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
// carry. A tag the profile registers nothing for reaches no property; the
// check reports it as gherkin/tag-unregistered (AC-GHK-002.8), so a
// misspelt qualifier is not lost in silence.
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

// failureOf builds one recorded failure from a Gherkin parse error,
// locating every error the parser reported when err wraps
// [gherkin.ParseError], and a single unlocated one otherwise (an I/O
// failure reading the source, which carries no position of its own).
//
// A free function rather than a builder method so that a check
// recomputing Gherkin structure against a warm corpus (identity.go,
// content.go) records a failure the identical way [Build] does, without
// needing a builder to do it.
func failureOf(reported string, err error) SourceFailure {
	failure := SourceFailure{Path: reported}
	var parse *gherkin.ParseError
	if errors.As(err, &parse) {
		failure.Errors = parse.Errors
	} else {
		failure.Errors = []specmodel.ParseError{{
			Message: err.Error(),
			Span:    span.Of(nil, 0, 0),
		}}
	}
	return failure
}

// blockFeatures parses every fenced Gherkin block in one document's
// bytes, attributing each to the document's own slug (AC-GHK-001.7). It
// is read from the parsed frontmatter rather than from the typed
// envelope, because the envelope's slug is a validated value and a
// document whose slug is absent or malformed still holds features that
// have to be attributed to what it actually wrote.
//
// The whole of src is offered to the fenced-block lookup, frontmatter
// included, so that every span a block yields is already in the
// document's own coordinates and needs no second arithmetic to become
// one.
//
// A free function, not a builder method, so that [builder.extractBlockFeatures]
// (a cold build) and a check recomputing Gherkin structure against a warm
// corpus (identity.go, content.go — MODEL's Feature and Criterion keep
// only the resolved identity, never the tags that produced it, so both
// paths reach for the source) perform the identical parse rather than
// two implementations of it that could drift apart.
func blockFeatures(vocab vocabulary, reported string, src []byte, doc *frontmatter.Document) (features []*gherkin.Feature, failures []SourceFailure) {
	slug := declaredSlug(doc)
	for _, block := range markdown.Fenced(src, gherkin.FenceTags()) {
		parsed, err := gherkin.ParseRegion(src, block.Inner, vocab.options())
		if err != nil {
			failures = append(failures, failureOf(reported, err))
			continue
		}
		if parsed == nil {
			continue
		}
		parsed.Record.Path = reported
		parsed.Record.DocumentSlug = slug
		parsed.Record.Source = specmodel.NewFeatureSourceFencedBlock(specmodel.FeatureSourceBlock{BlockSpan: block.Outer})
		features = append(features, parsed)
	}
	return features, failures
}

// featureFileFeature parses one `.feature` file's whole content: the
// Gherkin half of what [builder.visitFeatureSource] does at build time,
// factored out for the same reason [blockFeatures] is. Ownership is read
// from the parsed tags — having no document around it to inherit from, a
// feature source names its owner in a tag (AC-GHK-001.2).
func featureFileFeature(vocab vocabulary, reported string, src []byte) (*gherkin.Feature, []SourceFailure) {
	parsed, err := gherkin.Parse(src, vocab.options())
	if err != nil {
		return nil, []SourceFailure{failureOf(reported, err)}
	}
	if parsed == nil {
		return nil, nil
	}
	parsed.Record.Path = reported
	parsed.Record.DocumentSlug = vocab.owningSlug(parsed.Tags)
	parsed.Record.Source = specmodel.NewFeatureSourceFeatureFile(specmodel.FeatureSourceFile{})
	return parsed, nil
}

// extractBlockFeatures records the features written in one member's
// fenced Gherkin blocks, from the bytes the walk has already read.
func (b *builder) extractBlockFeatures(reported string, src []byte, doc *frontmatter.Document) {
	features, failures := blockFeatures(b.vocabulary, reported, src, doc)
	b.failures = append(b.failures, failures...)
	for _, f := range features {
		b.addFeature(f)
	}
}

// visitFeatureSource records the feature one `.feature` file holds, and
// the references its Gherkin makes.
//
// References are extracted against the file's own path even though a
// feature source is not a corpus member (GREF): an edge's endpoints are
// paths, and the referring file is the one a reader must open to fix a
// broken one. Extraction runs whether or not the Gherkin itself goes on
// to parse — a wikilink is a reference on the strength of what is
// written, not of whether the rest of the file is well-formed — and
// [references.Options.Frontmatter] is left nil, so the whole file is read
// as body and its info-string mask never comes into play: a `.feature`
// file carries no fences of its own.
func (b *builder) visitFeatureSource(reported string, src []byte, readErr error) {
	if readErr != nil {
		b.failures = append(b.failures, failureOf(reported, readErr))
		return
	}
	b.references = append(b.references, references.Extract(references.Options{
		Path:   reported,
		Source: src,
	})...)

	parsed, failures := featureFileFeature(b.vocabulary, reported, src)
	b.failures = append(b.failures, failures...)
	if parsed != nil {
		b.addFeature(parsed)
	}
}

// addFeature completes one parsed feature with the qualifier properties
// its criteria's tags carry, and records it. Path, DocumentSlug and
// Source are already set on parsed.Record by [blockFeatures] or
// [featureFileFeature].
//
// Each criterion also carries where it lives — its feature's path and
// identifier — and every tag written on it, so one criterion's record
// locates it without a second query (AC-GHK-005.2).
func (b *builder) addFeature(parsed *gherkin.Feature) {
	record := parsed.Record
	record.Criteria = make([]specmodel.Criterion, 0, len(parsed.Criteria))
	for _, criterion := range parsed.Criteria {
		nested := criterion.Record
		nested.Path = record.Path
		nested.FeatureId = record.Id
		nested.Tags = append([]string{}, criterion.Tags...)
		b.vocabulary.qualify(&nested, criterion.Tags)
		record.Criteria = append(record.Criteria, nested)
	}
	b.features = append(b.features, record)
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
