package corpus

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

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

// DocumentSummary is one corpus member as a listing reports it: a
// projection of [specmodel.IndexEntry] over the entry's path and the
// frontmatter envelope it parsed, carrying nothing else. It is not a new
// record type, so no index shape changes and `schema_version` stays
// "1.0".
//
// The field order — path, slug, type, status, title, description — is
// part of the answer's contract rather than a matter of taste: the
// command surface renders this struct directly, and both of its
// renderings report fields in declaration order.
//
// The five envelope fields are pointers because a member the index
// recorded with a parse error carries none of them, and absence is the
// honest answer — a null slug would assert the document declares none,
// which is not known. A pointer is what distinguishes that from a
// document declaring an empty value, which a bare string under
// `omitempty` could not.
type DocumentSummary struct {
	Path        string                  `json:"path"`
	Slug        *specmodel.Slug         `json:"slug,omitempty"`
	Type        *specmodel.DocumentType `json:"type,omitempty"`
	Status      *specmodel.Status       `json:"status,omitempty"`
	Title       *string                 `json:"title,omitempty"`
	Description *string                 `json:"description,omitempty"`
	// ParseError is the failure that kept this member's frontmatter from
	// being read, with its position. Present exactly when the five
	// envelope fields are absent, which is the same "exactly one of the
	// two" [specmodel.IndexEntry] carries.
	ParseError *specmodel.ParseError `json:"parse_error,omitempty"`
}

// SummaryOptions narrows what [Corpus.DocumentSummaries] returns.
//
// The zero value is the plain listing: every corpus member, at every
// status but the ones the active profile registers as no longer in
// effect. Both filters resolve against tables the profile registers,
// which is what lets them live here rather than in a caller — Document
// Core learns no vocabulary of its own by applying them.
type SummaryOptions struct {
	// Types, where it names one or more document types, narrows the
	// listing to members declaring one of them. Each value must be a key
	// of the active profile's [Profile.DocumentTypes]; one that is not is
	// reported rather than matched against nothing.
	//
	// An empty Types applies no type filter at all, which is not the same
	// as a filter matching nothing.
	Types []string

	// AllStatuses asks for summaries at every status, including the ones
	// the profile registers as retired. Without it those are omitted: a
	// document at such a status is no longer in effect, and an agent
	// orienting over the corpus that reads one has been misled rather
	// than merely charged for it.
	AllStatuses bool
}

// ErrUnregisteredType reports a type filter naming a value the active
// profile does not register. Callers test for it with [errors.Is]; the
// wrapped error carries the value and the type names that are registered.
var ErrUnregisteredType = errors.New("not a registered document type")

// DocumentSummaries returns one summary per corpus member the options
// select, in the index's own entry order — which the emitter fixes by
// repository-relative path, so two runs over an unchanged corpus answer
// identically (AC-SPECLIST-001.1). Nothing here sorts.
//
// Each path is passed through exactly as the index holds it: relative to
// the repository root, with forward slashes on every platform
// (AC-SPECLIST-001.4). Re-deriving one here would be a second notion of
// where a member lives.
//
// A file the index recorded as an EXCLUSION yields no summary
// (AC-SPECLIST-001.6) — it is not an entry, and whether a non-member
// ought to be one is a conformance judgement [Corpus.Check] already owns
// and reports as a finding. A member recorded with a PARSE ERROR is
// returned by an unfiltered listing regardless, carrying its path and
// that failure and no envelope values (AC-SPECLIST-001.5): the index
// recorded it as a member, so omitting it would make the listing disagree
// with the corpus. A TYPE filter is the one thing that drops it, because
// it declares no type for such a filter to match (AC-SPECLIST-002.5);
// the status filter does not, since it declares no status either and so
// is at none of the retired ones.
//
// The two filters compose: a call carrying both returns only the
// summaries satisfying both (AC-SPECLIST-003.4). A filter matching no
// member yields an empty collection rather than a failure
// (AC-SPECLIST-002.3), and so does a corpus holding no members
// (AC-SPECLIST-001.7) — the result is never nil.
//
// The one failure is a type the profile does not register, which is
// reported with the registered names and no summaries at all
// (AC-SPECLIST-002.4). An empty collection would make a typo
// indistinguishable from an empty category, which is a different answer
// to a different question.
func (c *Corpus) DocumentSummaries(opts SummaryOptions) ([]DocumentSummary, error) {
	wanted, err := c.resolveTypeFilter(opts.Types)
	if err != nil {
		return nil, err
	}
	withheld := withheldStatuses(c.profile.Lifecycle(), opts.AllStatuses)

	summaries := make([]DocumentSummary, 0, len(c.index.Entries))
	for _, entry := range c.index.Entries {
		summary := summarize(entry)
		if wanted != nil && (summary.Type == nil || !wanted[string(*summary.Type)]) {
			continue
		}
		if summary.Status != nil && withheld[string(*summary.Status)] {
			continue
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// resolveTypeFilter turns the requested type names into the set the
// listing matches against, and reports the first one the active profile
// does not register.
//
// A nil set means no type filter, which is what distinguishes "the caller
// asked for nothing in particular" from "the caller asked for a set that
// happens to be empty" — the first returns every member and the second
// could only return none.
func (c *Corpus) resolveTypeFilter(types []string) (map[string]bool, error) {
	if len(types) == 0 {
		return nil, nil
	}
	wanted := make(map[string]bool, len(types))
	for _, name := range types {
		if _, registered := c.types[name]; !registered {
			return nil, c.unregisteredType(name)
		}
		wanted[name] = true
	}
	return wanted, nil
}

// unregisteredType builds the error a type filter naming an unregistered
// value is reported with, listing what the active profile does register
// so the caller can correct the question rather than only learn it was
// wrong.
//
// The names are sorted because they come from a map, and an answer whose
// order depends on Go's map iteration would differ between two runs of
// one command.
func (c *Corpus) unregisteredType(name string) error {
	registered := make([]string, 0, len(c.types))
	for typeName := range c.types {
		registered = append(registered, typeName)
	}
	sort.Strings(registered)

	list := "none are registered"
	if len(registered) > 0 {
		list = strings.Join(registered, ", ")
	}
	return fmt.Errorf("corpus: %q is %w; registered: %s", name, ErrUnregisteredType, list)
}

// withheldStatuses is the set of statuses the listing omits: the ones the
// active profile registers as no longer in effect, or none at all where
// the caller asked for every status.
//
// The set is read from the registration and never spelled here. A profile
// whose documents move from `current` to `sunset` gets exactly this
// behavior, and a host holding the spelling one framework happens to use
// would be the single change that breaks that — which is why the
// extension point's own source scan fails on any string literal naming a
// registered status.
func withheldStatuses(lifecycle LifecycleRules, allStatuses bool) map[string]bool {
	if allStatuses {
		return nil
	}
	withheld := make(map[string]bool, len(lifecycle.RetiredStatuses))
	for _, status := range lifecycle.RetiredStatuses {
		withheld[status] = true
	}
	return withheld
}

// summarize projects one index entry onto the summary shape.
//
// The parse error is copied rather than aliased, so a caller holding a
// summary cannot reach into the index the corpus is still serving from.
func summarize(entry specmodel.IndexEntry) DocumentSummary {
	summary := DocumentSummary{Path: entry.Path}
	if envelope := entry.Frontmatter; envelope != nil {
		summary.Slug = ptrTo(envelope.Slug)
		summary.Type = ptrTo(envelope.Type)
		summary.Status = ptrTo(envelope.Status)
		summary.Title = ptrTo(envelope.Title)
		summary.Description = ptrTo(envelope.Description)
	}
	if entry.ParseError != nil {
		failure := *entry.ParseError
		summary.ParseError = &failure
	}
	return summary
}

// ptrTo returns a pointer to a copy of value, which is how a summary's
// envelope fields are set: they are pointers so that absence is
// representable, and every present one is taken from a value.
func ptrTo[T any](value T) *T {
	return &value
}
