package corpus

import (
	"regexp"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// gherkinIdentityFindings reports every VTAG identity defect over
// features: whether a feature or a criterion is addressed, addressed
// once, and whether its address agrees with where it sits.
//
// features comes from [Corpus.recomputeGherkin] rather than
// [Corpus.index].Features, because the persisted record keeps only a
// construct's *resolved* identity — an empty Id where nothing matched —
// and several of these checks turn on the tags that produced (or failed
// to produce) that resolution, which the persisted shape does not retain
// either warm or cold. The recompute is the caller's (check.go), shared
// with content.go's VCON checks, so that one Check or CheckFor call
// parses the corpus's Gherkin once rather than twice.
func (c *Corpus) gherkinIdentityFindings(vocab vocabulary, rules FeatureTagRules, features []*gherkin.Feature) []specmodel.Finding {
	var findings []specmodel.Finding
	findings = append(findings, untaggedOrMalformedFindings(vocab, rules, features)...)
	findings = append(findings, multipleFindings(vocab, rules, features)...)
	findings = append(findings, duplicateIdentifierFindings(features)...)
	findings = append(findings, codeMismatchFindings(vocab, rules, features)...)
	findings = append(findings, ownershipFindings(c, vocab, features)...)
	return findings
}

// untaggedSpans returns the set of criterion positions this feature's
// [gherkin.PartialOutline]s already account for, so [untaggedFindings]
// does not report a partially tagged Examples block a second time under
// gherkin/criterion-untagged: the FRD frames the block-level defect as
// gherkin/examples-partially-tagged specifically (VCON), not as an
// ordinary untagged Scenario, and the two rules would otherwise double
// -report the identical block.
func untaggedSpans(feature *gherkin.Feature) map[span.Span]bool {
	if len(feature.PartialOutlines) == 0 {
		return nil
	}
	set := make(map[span.Span]bool)
	for _, outline := range feature.PartialOutlines {
		for _, s := range outline.Untagged {
			set[s] = true
		}
	}
	return set
}

// literalPrefix returns the longest run of literal characters an
// identifier pattern opens with, after its leading `^` anchor — the
// marker [malformedFindings] compares a candidate tag's own prefix
// against to tell a near-miss identifier apart from an unrelated tag
// (AC-GHK-002.8). It stops at the first regular-expression
// metacharacter, so `^AC-(?P<code>...)` yields `"AC-"` without this
// function reading any more of the pattern's structure than that — a
// pattern with no literal run before its first capture, such as
// `^[A-Z]+-(?P<code>...)`, yields the empty string, and no tag is ever
// treated as a near-miss under it.
func literalPrefix(pattern string) string {
	pattern = strings.TrimPrefix(pattern, "^")
	for i, r := range pattern {
		if strings.ContainsRune(`\.+*?()[]{}|^$`, r) {
			return pattern[:i]
		}
	}
	return pattern
}

// candidateTags returns every tag on the element that shares pattern's
// literal prefix but does not itself match pattern in full — the
// near-miss tags [malformedFindings] and [multipleFindings]'s "which
// tag" question both need. Returns nil, never reported, when the pattern
// carries no literal prefix to test against (see [literalPrefix]).
func candidateTags(pattern *regexp.Regexp, prefix string, tags []string) []string {
	if prefix == "" {
		return nil
	}
	var candidates []string
	for _, tag := range tags {
		if !strings.HasPrefix(tag, prefix) {
			continue
		}
		if pattern != nil && pattern.MatchString(tag) {
			continue
		}
		candidates = append(candidates, tag)
	}
	return candidates
}

// untaggedOrMalformedFindings reports gherkin/feature-untagged,
// gherkin/criterion-untagged and gherkin/identifier-malformed
// (AC-GHK-002.2, AC-GHK-002.3, AC-GHK-006.2, AC-GHK-006.5) — the three
// rules that all turn on the same first question, "did this element
// resolve an identifier", and so are decided together rather than by
// three independent walks that could each answer differently for one
// element: a feature or criterion resolving none is untagged when it
// carries no tag resembling one, and malformed when it does.
//
// A record that DID resolve an identifier is not visited by either rule
// even if it also carries a near-miss tag: [multipleFindings] is what a
// second matching tag means, and a tag that merely resembles the pattern
// without matching it is judged by nothing once one tag has already
// resolved the identity.
func untaggedOrMalformedFindings(vocab vocabulary, rules FeatureTagRules, features []*gherkin.Feature) []specmodel.Finding {
	featurePrefix := literalPrefix(rules.FeaturePattern)
	criterionPrefix := literalPrefix(rules.CriterionPattern)

	var findings []specmodel.Finding
	for _, feature := range features {
		if feature.Record.Id == "" {
			if candidates := candidateTags(vocab.feature, featurePrefix, feature.Tags); len(candidates) > 0 {
				findings = append(findings, finding.IdentifierMalformed(
					feature.Record.Path, feature.Record.Position, candidates[0], rules.FeaturePattern))
			} else {
				findings = append(findings, finding.FeatureUntagged(
					feature.Record.Path, feature.Record.Position, rules.FeaturePattern))
			}
		}

		excluded := untaggedSpans(feature)
		for _, criterion := range feature.Criteria {
			if criterion.Record.Id != "" {
				continue
			}
			if excluded[criterion.Record.Position] {
				continue
			}
			candidates := candidateTags(vocab.criterion, criterionPrefix, criterion.Tags)
			if len(candidates) > 0 {
				findings = append(findings, finding.IdentifierMalformed(
					feature.Record.Path, criterion.Record.Position, candidates[0], rules.CriterionPattern))
			} else {
				findings = append(findings, finding.CriterionUntagged(
					feature.Record.Path, criterion.Record.Position, rules.CriterionPattern))
			}
		}
	}
	return findings
}

// matchedAll returns every tag pattern matches, in the order written —
// widening [matched], which only the first of these ever reaches.
func matchedAll(pattern *regexp.Regexp, tags []string) []string {
	if pattern == nil {
		return nil
	}
	var matches []string
	for _, tag := range tags {
		if pattern.MatchString(tag) {
			matches = append(matches, tag)
		}
	}
	return matches
}

// multipleFindings reports gherkin/identifier-multiple: an element
// carrying more than one tag matching its identifier pattern, naming all
// of them (AC-GHK-002.5).
func multipleFindings(vocab vocabulary, rules FeatureTagRules, features []*gherkin.Feature) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, feature := range features {
		if matches := matchedAll(vocab.feature, feature.Tags); len(matches) > 1 {
			findings = append(findings, finding.IdentifierMultiple(
				feature.Record.Path, feature.Record.Position, matches, rules.FeaturePattern))
		}
		for _, criterion := range feature.Criteria {
			matches := matchedAll(vocab.criterion, criterion.Tags)
			if len(matches) <= 1 {
				continue
			}
			findings = append(findings, finding.IdentifierMultiple(
				feature.Record.Path, criterion.Record.Position, matches, rules.CriterionPattern))
		}
	}
	return findings
}

// identifierOccurrence is one record's identifier, addressable enough for
// [duplicateIdentifierFindings] to report against.
type identifierOccurrence struct {
	id       string
	path     string
	position span.Span
}

// duplicateIdentifierFindings reports gherkin/identifier-duplicate for
// every pair of feature identifiers, and every pair of criterion
// identifiers, declared identically elsewhere in the recomputed corpus
// (AC-GHK-002.4, AC-GHK-003.4). Features and criteria are grouped
// separately — the two vocabularies practically never collide, since
// they are shaped differently, but nothing here assumes it.
func duplicateIdentifierFindings(features []*gherkin.Feature) []specmodel.Finding {
	var featureOccurrences, criterionOccurrences []identifierOccurrence
	for _, feature := range features {
		if feature.Record.Id != "" {
			featureOccurrences = append(featureOccurrences, identifierOccurrence{
				id: string(feature.Record.Id), path: feature.Record.Path, position: feature.Record.Position,
			})
		}
		for _, criterion := range feature.Criteria {
			if criterion.Record.Id == "" {
				continue
			}
			criterionOccurrences = append(criterionOccurrences, identifierOccurrence{
				id: string(criterion.Record.Id), path: feature.Record.Path, position: criterion.Record.Position,
			})
		}
	}

	var findings []specmodel.Finding
	findings = append(findings, duplicatesAmong(featureOccurrences)...)
	findings = append(findings, duplicatesAmong(criterionOccurrences)...)
	return findings
}

// duplicatesAmong reports every ordered pair within each group of
// occurrences sharing an identifier, so that checking any one of them
// surfaces the collision — the same fan-out [duplicateSlugs] uses for
// two documents sharing a slug.
func duplicatesAmong(occurrences []identifierOccurrence) []specmodel.Finding {
	groups := map[string][]int{}
	for i, occ := range occurrences {
		groups[occ.id] = append(groups[occ.id], i)
	}

	var findings []specmodel.Finding
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		for _, i := range group {
			for _, j := range group {
				if i == j {
					continue
				}
				findings = append(findings, finding.IdentifierDuplicate(
					occurrences[i].path, occurrences[i].position,
					occurrences[i].id, occurrences[j].path, occurrences[j].position))
			}
		}
	}
	return findings
}

// captureValues returns the named capture values pattern extracts from
// id, keyed by capture name, or nil when the pattern is unset or id does
// not match at all — either of which leaves nothing for
// [codeMismatchFindings] to compare, and both are already reported by
// [untaggedFindings] or [malformedFindings].
func captureValues(pattern *regexp.Regexp, id string) map[string]string {
	if pattern == nil {
		return nil
	}
	match := pattern.FindStringSubmatch(id)
	if match == nil {
		return nil
	}
	names := pattern.SubexpNames()
	values := make(map[string]string, len(names))
	for i, name := range names {
		if name == "" {
			continue
		}
		values[name] = match[i]
	}
	return values
}

// codeMismatchFindings reports gherkin/criterion-code-mismatch: a
// criterion whose identifier's code and number are not its feature's
// (AC-GHK-003.6), read from the named captures the profile registers as
// shared between the two patterns — never from [specmodel.Feature.Code],
// which groups several features under one code by design (AC-GHK-003.5)
// and is therefore the wrong value to require a criterion agree with.
func codeMismatchFindings(vocab vocabulary, rules FeatureTagRules, features []*gherkin.Feature) []specmodel.Finding {
	if len(rules.SharedCaptures) == 0 {
		return nil
	}

	var findings []specmodel.Finding
	for _, feature := range features {
		featureCaptures := captureValues(vocab.feature, string(feature.Record.Id))
		if featureCaptures == nil {
			continue
		}
		for _, criterion := range feature.Criteria {
			criterionCaptures := captureValues(vocab.criterion, string(criterion.Record.Id))
			if criterionCaptures == nil {
				continue
			}
			mismatched := false
			for _, name := range rules.SharedCaptures {
				if featureCaptures[name] != criterionCaptures[name] {
					mismatched = true
					break
				}
			}
			if !mismatched {
				continue
			}
			findings = append(findings, finding.CriterionCodeMismatch(
				feature.Record.Path, criterion.Record.Position,
				string(criterion.Record.Id), string(feature.Record.Id), rules.CriterionPattern))
		}
	}
	return findings
}

// ownershipFindings reports gherkin/owner-missing, gherkin/owner-mismatch
// and gherkin/owner-unresolved (AC-GHK-002.10, AC-GHK-002.11): whether a
// feature's owning document is stated, whether — for a fenced block —
// it agrees with the document it is written in, and whether the slug it
// names resolves to a corpus member at all.
//
// vocab.owningSlug is asked of feature.Tags directly rather than trusted
// from feature.Record.DocumentSlug: for a fenced block that value is
// always the containing document's own slug (AC-GHK-001.7), which is
// what lets a contradicting tag be detected in the first place.
func ownershipFindings(c *Corpus, vocab vocabulary, features []*gherkin.Feature) []specmodel.Finding {
	if vocab.owner == "" {
		return nil
	}

	var findings []specmodel.Finding
	for _, feature := range features {
		tagged := vocab.owningSlug(feature.Tags)

		if _, isFile := feature.Record.Source.AsFeatureFile(); isFile {
			switch {
			case tagged == "":
				findings = append(findings, finding.OwnerMissing(
					feature.Record.Path, feature.Record.Position, vocab.owner))
			default:
				if _, ok := c.Slug(string(tagged)); !ok {
					findings = append(findings, finding.OwnerUnresolved(
						feature.Record.Path, feature.Record.Position, string(tagged)))
				}
			}
			continue
		}

		if tagged != "" && tagged != feature.Record.DocumentSlug {
			findings = append(findings, finding.OwnerMismatch(
				feature.Record.Path, feature.Record.Position,
				string(tagged), string(feature.Record.DocumentSlug)))
		}
	}
	return findings
}
