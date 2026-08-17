package corpus

import "github.com/brettcallaghan/agent-plugins/cli/core/specmodel"

// Feature returns the feature id names in c's index, and whether one was
// found. The lookup reads c.index.Features exactly as it stands — warm or
// freshly built — so it never re-walks the corpus (AC-GHK-007.6).
func (c *Corpus) Feature(id specmodel.FeatureId) (specmodel.Feature, bool) {
	for _, feature := range c.index.Features {
		if feature.Id == id {
			return feature, true
		}
	}
	return specmodel.Feature{}, false
}

// Features returns every feature in c's index, in the index's own
// deterministic order (path, then position within the path).
func (c *Corpus) Features() []specmodel.Feature {
	return c.index.Features
}

// FeaturesByCode returns every feature grouped under code, in the index's
// own order (AC-GHK-003.3, AC-GHK-007.3).
//
// An empty result is a valid selection, not an error: code names a group,
// not a specific record, so a code no feature carries answers "none",
// exactly as an unreferenced-but-declared slug does for [Corpus.Slug].
func (c *Corpus) FeaturesByCode(code specmodel.FeatureCode) []specmodel.Feature {
	var matched []specmodel.Feature
	for _, feature := range c.index.Features {
		if feature.Code == code {
			matched = append(matched, feature)
		}
	}
	return matched
}

// Criterion returns the criterion id names in c's index, and whether one
// was found. Every feature's criteria are scanned in the feature's own
// order, so the search order matches [Corpus.Features].
func (c *Corpus) Criterion(id specmodel.CriterionId) (specmodel.Criterion, bool) {
	for _, feature := range c.index.Features {
		for _, criterion := range feature.Criteria {
			if criterion.Id == id {
				return criterion, true
			}
		}
	}
	return specmodel.Criterion{}, false
}

// CriteriaOf returns every criterion beneath the feature named featureId,
// in that feature's own order (AC-GHK-007.3).
//
// A featureId naming no feature answers the same empty selection as a
// feature with no criteria: --of is a filter over the criterion
// collection, not an identifier lookup, so unlike [Corpus.Feature] and
// [Corpus.Criterion] it never distinguishes "unknown" from "empty".
func (c *Corpus) CriteriaOf(featureId specmodel.FeatureId) []specmodel.Criterion {
	feature, found := c.Feature(featureId)
	if !found {
		return nil
	}
	return feature.Criteria
}

// CriteriaByTag returns every criterion in c's index carrying the
// qualifier tag, per the active profile's registered vocabulary
// (AC-GHK-007.3).
//
// tag is matched against [FeatureTagRules.Qualifiers] exactly as it is
// registered — the host holds no qualifier spelling of its own — so a tag
// the active profile does not register matches nothing, the same empty
// selection a registered tag with no carriers answers.
func (c *Corpus) CriteriaByTag(tag string) []specmodel.Criterion {
	qualifier, ok := findQualifier(c.profile.FeatureTags().Qualifiers, tag)
	if !ok {
		return nil
	}
	var matched []specmodel.Criterion
	for _, feature := range c.index.Features {
		for _, criterion := range feature.Criteria {
			if qualifierMatches(criterion, qualifier) {
				matched = append(matched, criterion)
			}
		}
	}
	return matched
}

// findQualifier returns the registration among qualifiers whose Tag is
// tag, and whether one was found.
func findQualifier(qualifiers []QualifierTag, tag string) (QualifierTag, bool) {
	for _, q := range qualifiers {
		if q.Tag == tag {
			return q, true
		}
	}
	return QualifierTag{}, false
}

// qualifierMatches reports whether criterion carries the property
// qualifier's meaning names — the same three-way switch
// [specmodel.QualifierMeaning] admits everywhere else it is read, applied
// here in the read direction rather than the write direction
// `criterionTags` (cmd/feature.go) applies for `--emit`.
func qualifierMatches(criterion specmodel.Criterion, qualifier QualifierTag) bool {
	switch qualifier.Meaning {
	case specmodel.QualifierMeaningNegation:
		return criterion.Negative
	case specmodel.QualifierMeaningManualVerification:
		return criterion.Manual
	case specmodel.QualifierMeaningVerificationLevel:
		return criterion.Level != nil && *criterion.Level == qualifier.Level
	default:
		return false
	}
}
