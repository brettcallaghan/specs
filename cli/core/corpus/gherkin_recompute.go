package corpus

import (
	"fmt"

	"github.com/brettcallaghan/specs/cli/core/frontmatter"
	"github.com/brettcallaghan/specs/cli/core/gherkin"
)

// recomputeGherkin re-parses every Gherkin source the corpus holds —
// every corpus member's fenced blocks and every `.feature` file beneath
// the spec root — carrying the raw tags that identify every construct.
//
// Those tags are what identity.go's and content.go's checks need and
// what nothing else in this package retains: [specmodel.Feature] and
// [specmodel.Criterion] keep only the *resolved* identity (an empty `Id`
// where nothing matched), never the tag that failed to match or the
// second tag that matched alongside the first. A cold build discards
// that information at [builder.addFeature] exactly as a warm corpus
// never had it, so this recomputes unconditionally rather than branching
// on [Corpus.warm] — the same shape [duplicateSlugs]'s warm-path
// recomputation has in [fromIndex], widened here because the missing
// data costs a second parse rather than a second lookup.
//
// The walk mirrors [scanSources] exactly — [Build] and this read the
// same files the same way — but a fenced block is only extracted from a
// path this corpus already knows is a member with fully-decoded
// frontmatter ([Corpus.byPath] plus a non-nil [specmodel.IndexEntry.Frontmatter]),
// which is the same condition [builder.classifyDocument] gates
// [builder.extractBlockFeatures] on. Consulting the already-built index
// for that decision, rather than re-deriving type-membership from
// scratch, is what keeps this walk from becoming a second, divergeable
// implementation of corpus membership.
func (c *Corpus) recomputeGherkin() ([]*gherkin.Feature, []SourceFailure, error) {
	vocab := newVocabulary(c.profile.FeatureTags())

	var features []*gherkin.Feature
	var failures []SourceFailure

	err := scanSources(c.specRoot, c.repoRoot, func(kind sourceKind, reported string, src []byte, readErr error) {
		if kind == featureSource {
			if readErr != nil {
				failures = append(failures, failureOf(reported, readErr))
				return
			}
			parsed, fails := featureFileFeature(vocab, reported, src)
			failures = append(failures, fails...)
			if parsed != nil {
				features = append(features, parsed)
			}
			return
		}

		if readErr != nil {
			// An unreadable candidate is not a Gherkin source failure; the
			// corpus-level entry already carries that failure.
			return
		}
		i, member := c.byPath[reported]
		if !member || c.index.Entries[i].Frontmatter == nil {
			// Excluded, or a member whose frontmatter could not be fully
			// decoded: extractBlockFeatures never ran for this path at build
			// time either, so no fenced block here holds a feature.
			return
		}
		doc, perr := frontmatter.Parse(src)
		if perr != nil || doc == nil {
			// The build already proved this path decodes; a failure here
			// would mean the file changed beneath the walk, which the
			// caller's own re-check will surface elsewhere.
			return
		}
		found, fails := blockFeatures(vocab, reported, src, doc)
		failures = append(failures, fails...)
		features = append(features, found...)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("corpus: recomputing Gherkin structure: %w", err)
	}
	return features, failures, nil
}
