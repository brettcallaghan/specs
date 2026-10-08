// Package finding owns the one shape every check reports through, the
// severity vocabulary that ranks it, and the two renderings of a run's
// report.
//
// A check that formatted its own output would be a check no other
// consumer could read, and a report that listed only what went wrong
// would be indistinguishable from a report of checks that never ran.
// Both problems are solved the same way: a check produces [Finding]
// values and nothing else, and the [Report] accounts for every check it
// knows about — including the ones the caller did not ask for.
//
// The shapes themselves are generated from schemas/finding.strata,
// because the finding JSON is a published interface an agent parses.
// This package owns the rules about them: which severities exist, how a
// check that could not run is recorded, how a message is derived from a
// variant, and the order findings are emitted in. It computes no
// positions — a [specmodel.Span] arrives already built by the Spans
// component, which is the only place one is constructed.
package finding

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// The severities a finding may carry.
//
// Re-exported from the generated model so a caller reasoning about
// severity need not reach past this package for the vocabulary that
// governs it.
const (
	// SeverityError marks a conformance failure.
	SeverityError = specmodel.SeverityError
	// SeverityWarning marks a probable problem that does not by itself
	// fail a build.
	SeverityWarning = specmodel.SeverityWarning
	// SeverityUnverified marks a check that was requested and could not
	// run. It is not a milder failure; it is the absence of a result.
	SeverityUnverified = specmodel.SeverityUnverified
)

// The rule identifiers of the checks built so far.
//
// They are namespaced kebab-case so the family a check belongs to is
// legible without a lookup, and they are constants because they are a
// published surface: a caller naming a subset of checks to run names
// them by these values, and reads each one's outcome back under the same
// value.
//
// [RuleUnresolvedAnchor] is the one identifier here that reports at more
// than one severity: an anchor matching no heading is an error, and an
// anchor whose target document has no recorded headings is
// [SeverityUnverified] under the same identifier — the tool could not
// look, which is not the claim that it looked and found nothing.
const (
	// RuleSchemaViolation covers a frontmatter value that violates the
	// schema for its document type.
	RuleSchemaViolation specmodel.RuleId = "frontmatter/schema"
	// RuleMissingProperty covers a required frontmatter property the
	// document does not declare.
	RuleMissingProperty specmodel.RuleId = "frontmatter/missing-property"
	// RuleSlugPattern covers a slug that does not match the pattern the
	// active profile supplies for the document's type.
	RuleSlugPattern specmodel.RuleId = "slug/pattern-mismatch"
	// RuleDuplicateSlug covers a slug declared by more than one corpus
	// member.
	RuleDuplicateSlug specmodel.RuleId = "slug/duplicate"
	// RuleFrontmatterParse covers a frontmatter block that could not be
	// read as a document envelope, so no other check could examine the
	// document at all.
	RuleFrontmatterParse specmodel.RuleId = "frontmatter/parse"
	// RuleUnresolvedReference covers a reference — written in a
	// document's body or held by a frontmatter property the active
	// profile registers as reference-bearing — whose slug matches no
	// corpus member. One identifier covers all three ways this can
	// happen (a body reference, a frontmatter reference, and a slug
	// matching a file that exists but is not a member): they are the
	// same defect reported from three places, not three defects.
	RuleUnresolvedReference specmodel.RuleId = "reference/unresolved"
	// RuleUnresolvedAnchor covers a reference's anchor matching no
	// heading of its target document, once the target's slug has
	// resolved and its headings are known. Distinct from
	// [RuleUnresolvedReference] so `--only` can select the anchor checks
	// without selecting slug resolution.
	RuleUnresolvedAnchor specmodel.RuleId = "reference/unresolved-anchor"
	// RuleAmbiguousAnchor covers an anchor whose identifier is the
	// identifier of more than one heading of its target document.
	RuleAmbiguousAnchor specmodel.RuleId = "reference/ambiguous-anchor"
	// RuleRetiredTarget covers a reference from a document at a status
	// the active profile's lifecycle registration treats as live to one
	// at a registered retired status. It is the first rule this package
	// declares at [SeverityWarning] rather than [SeverityError]: the
	// target still resolves, so this is not a conformance failure, but
	// it is the one finding that tells an author what to cite instead.
	RuleRetiredTarget specmodel.RuleId = "reference/retired-target"
	// RuleIncompletePairing covers a pairing the active profile
	// registers — a property declared from one end and a counterpart
	// the resolved target must declare in return — where the target
	// does not declare it.
	RuleIncompletePairing specmodel.RuleId = "reference/pairing-incomplete"
	// RuleTypeMismatch covers a reference-bearing property the active
	// profile requires a specific target document `type` for, whose
	// resolved target declares a different one.
	RuleTypeMismatch specmodel.RuleId = "reference/type-mismatch"
	// RuleCorpusEmpty covers a check run whose corpus recorded no
	// members at all: the spec root exists but discovery found nothing
	// beneath it to check. A report with zero findings otherwise looks
	// identical whether the run examined a clean corpus or examined
	// nothing, and only this rule tells the two apart.
	RuleCorpusEmpty specmodel.RuleId = "corpus/empty"
	// RuleUnmappedCitation covers a citation whose slug names a corpus
	// document the caller's link map does not cover. The slug resolves;
	// the map is incomplete.
	RuleUnmappedCitation specmodel.RuleId = "links/unmapped"
	// RuleUnresolvedCitation covers a citation whose slug names no
	// corpus document at all, with or without a link map. Distinct from
	// [RuleUnmappedCitation]: nothing resolves the slug, so this may be
	// a mistyped reference or illustrative prose that was never meant to
	// resolve.
	RuleUnresolvedCitation specmodel.RuleId = "links/unresolved"

	// RuleCriterionUntagged covers a Scenario carrying no identifier tag.
	RuleCriterionUntagged specmodel.RuleId = "gherkin/criterion-untagged"
	// RuleFeatureUntagged covers a Feature carrying no identifier tag.
	RuleFeatureUntagged specmodel.RuleId = "gherkin/feature-untagged"
	// RuleIdentifierMalformed covers an identifier tag that does not
	// match the pattern registered for identifiers of its kind.
	RuleIdentifierMalformed specmodel.RuleId = "gherkin/identifier-malformed"
	// RuleIdentifierMultiple covers a Feature or a Scenario carrying more
	// than one tag matching its identifier pattern.
	RuleIdentifierMultiple specmodel.RuleId = "gherkin/identifier-multiple"
	// RuleIdentifierDuplicate covers a feature or a criterion identifier
	// declared by two records in the corpus — the Gherkin analogue of
	// [RuleDuplicateSlug].
	RuleIdentifierDuplicate specmodel.RuleId = "gherkin/identifier-duplicate"
	// RuleCriterionCodeMismatch covers a criterion whose identifier's
	// code and number are not its feature's.
	RuleCriterionCodeMismatch specmodel.RuleId = "gherkin/criterion-code-mismatch"
	// RuleOwnerMissing covers a feature in a standalone feature source
	// carrying no owning-document tag.
	RuleOwnerMissing specmodel.RuleId = "gherkin/owner-missing"
	// RuleOwnerMismatch covers a feature in a fenced block carrying an
	// owning-document tag that names a document other than the one it is
	// written in.
	RuleOwnerMismatch specmodel.RuleId = "gherkin/owner-mismatch"
	// RuleOwnerUnresolved covers an owning-document tag whose slug names
	// no corpus member.
	RuleOwnerUnresolved specmodel.RuleId = "gherkin/owner-unresolved"

	// RuleGherkinParse covers a Gherkin source that could not be parsed.
	RuleGherkinParse specmodel.RuleId = "gherkin/parse"
	// RuleCriterionEmpty covers a criterion with no steps.
	RuleCriterionEmpty specmodel.RuleId = "gherkin/criterion-empty"
	// RuleScenarioProse covers a non-empty Scenario description, reported
	// as a probable untagged criterion rather than the rationale a
	// well-formed criterion carries alongside its steps.
	RuleScenarioProse specmodel.RuleId = "gherkin/scenario-prose"
	// RuleExamplesPartiallyTagged covers a Scenario Outline whose
	// Examples blocks carry an identifier on some and not others.
	RuleExamplesPartiallyTagged specmodel.RuleId = "gherkin/examples-partially-tagged"

	// The c4 rules check the Mermaid C4 diagrams in architecture
	// blueprints. They are declared here before any check raises them,
	// so the identifiers and their detail variants exist as one
	// published surface; none is corpus-wide, since each finding is
	// narrowed to the file it sits in, and none is listed among the
	// rules a check run reports until the check that raises it lands.

	// RuleC4Parse covers a C4 diagram statement the parser cannot read,
	// at [SeverityError]. Like [RuleUnresolvedAnchor] it also reports at
	// [SeverityUnverified] under the same identifier: a `mermaid` block
	// whose diagram type the parser does not read was not looked inside,
	// which is not the claim that it was looked inside and found sound.
	RuleC4Parse specmodel.RuleId = "c4/parse"
	// RuleC4UndeclaredNode covers a non-external node that resolves to
	// no element in the scope of its drawing document and is not a
	// passive store.
	RuleC4UndeclaredNode specmodel.RuleId = "c4/undeclared-node"
	// RuleC4AmbiguousNode covers a node whose label matches the entity
	// names of more than one element in scope.
	RuleC4AmbiguousNode specmodel.RuleId = "c4/ambiguous-node"
	// RuleC4Bidirectional covers a `BiRel` statement: an arrow states
	// one dependency, in one direction.
	RuleC4Bidirectional specmodel.RuleId = "c4/bidirectional"
	// RuleC4DuplicateEdge covers two relationship statements in one
	// diagram with the same source and target after `Rel_Back`
	// normalisation.
	RuleC4DuplicateEdge specmodel.RuleId = "c4/duplicate-edge"
	// RuleC4DrawingConflict covers two drawings of one relationship,
	// anywhere in the corpus, that differ in technology, description or
	// tags.
	RuleC4DrawingConflict specmodel.RuleId = "c4/drawing-conflict"
	// RuleC4InteractionConflict covers a relationship tagged both `sync`
	// and `async`.
	RuleC4InteractionConflict specmodel.RuleId = "c4/interaction-conflict"
	// RuleC4ContractUnresolved covers a `contract:` tag whose slug no
	// corpus member of the registered contract type declares.
	RuleC4ContractUnresolved specmodel.RuleId = "c4/contract-unresolved"
	// RuleC4ContractMismatch covers a `contract:` tag naming a contract
	// the source element does not consume or the target element does not
	// provide.
	RuleC4ContractMismatch specmodel.RuleId = "c4/contract-mismatch"
	// RuleC4TagSeparator covers a relationship tags argument that joins
	// its tags with a comma rather than `+`.
	RuleC4TagSeparator specmodel.RuleId = "c4/tag-separator"
)

// rootPointer is what an empty JSON Pointer is rendered as, so a message
// about the frontmatter object itself does not begin with a bare colon.
const rootPointer = "/"

// SchemaViolation reports a frontmatter value that fails the schema for
// its document type.
//
// The value is the raw source text rather than the parsed value, because
// the constraint that failed is usually a textual one and reporting what
// a YAML parser inferred would quote a value the author never wrote.
func SchemaViolation(path string, at specmodel.Span, pointer, constraint, value string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleSchemaViolation,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("%s: %q does not satisfy %s",
			renderPointer(pointer), value, constraint),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    pointer,
			Constraint: constraint,
			Value:      value,
		}),
	}
}

// MissingProperty reports a required frontmatter property the document
// does not declare.
func MissingProperty(path string, at specmodel.Span, pointer, property string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleMissingProperty,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("%s: required property %q is not declared",
			renderPointer(pointer), property),
		Detail: specmodel.NewFindingDetailMissingProperty(specmodel.FindingMissingProperty{
			Pointer:  pointer,
			Property: property,
		}),
	}
}

// SlugPattern reports a slug that does not match the pattern the active
// profile supplies for the document's type.
//
// Both the slug and the pattern are quoted in the message as well as
// carried in the detail, so the finding is actionable whether it is read
// or parsed.
func SlugPattern(path string, at specmodel.Span, slug, pattern string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleSlugPattern,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("%q does not match pattern %q", slug, pattern),
		Detail: specmodel.NewFindingDetailSlugPattern(specmodel.FindingSlugPattern{
			Slug:    slug,
			Pattern: pattern,
		}),
	}
}

// DuplicateSlug reports a slug declared by two corpus members, against
// one of them and naming the other.
//
// The caller raises one of these against each document, so that checking
// either one surfaces the collision.
func DuplicateSlug(path string, at specmodel.Span, slug, otherPath string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleDuplicateSlug,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("slug %q is also declared by %s", slug, otherPath),
		Detail: specmodel.NewFindingDetailDuplicateSlug(specmodel.FindingDuplicateSlug{
			Slug:      slug,
			OtherPath: otherPath,
		}),
	}
}

// FrontmatterParse reports a frontmatter block that could not be read as
// a document envelope.
//
// It carries [SeverityError] rather than [SeverityUnverified], even
// though the practical consequence is that other checks could not run.
// Unverified means the tool was unable to perform a check — no validator
// installed, a dependency missing — and says nothing about the document.
// This says something definite about the document: its frontmatter is
// wrong, and that is a conformance failure in its own right, whoever
// else was prevented from looking at it.
//
// message is the parser's own wording, taken from the index entry's
// parse error, so the finding and the index agree about what happened
// rather than describing it twice.
func FrontmatterParse(path string, at specmodel.Span, message string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleFrontmatterParse,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("frontmatter could not be read: %s", message),
		Detail: specmodel.NewFindingDetailFrontmatterParse(specmodel.FindingFrontmatterParse{
			Message: message,
		}),
	}
}

// UnresolvedReference reports a reference whose slug matches no corpus
// member, or matches a file the index recorded as excluded or as a parse
// failure rather than as a member.
//
// The message reads differently depending on origin, matching how the
// defect is best fixed: a body reference is quoted back exactly as the
// author wrote it, brackets included, while a frontmatter reference names
// the property that holds it rather than making the reader search the
// document for the offending value (AC-SPECREF-002.1, AC-SPECREF-002.3).
// nonMember is nil for an ordinary unresolved slug and set when the slug
// matched a non-member file, in which case the message states that
// reason instead of "matches nothing" (AC-SPECREF-002.7).
func UnresolvedReference(
	path string, at specmodel.Span, text, slug string,
	origin specmodel.ReferenceOrigin, nonMember *specmodel.ReferenceNonMember,
) specmodel.Finding {
	outcome := "matches no document's slug"
	if nonMember != nil {
		outcome = fmt.Sprintf("target exists but is not a corpus member: %s", nonMemberReason(*nonMember))
	}

	var message string
	if fm, ok := origin.AsFrontmatterProperty(); ok {
		message = fmt.Sprintf("%s: %q %s", fm.Property, text, outcome)
	} else {
		message = fmt.Sprintf("[[%s]] %s", text, outcome)
	}

	return specmodel.Finding{
		Rule:     RuleUnresolvedReference,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  message,
		Detail: specmodel.NewFindingDetailUnresolvedReference(specmodel.FindingUnresolvedReference{
			Slug:      slug,
			Origin:    origin,
			NonMember: nonMember,
		}),
	}
}

// nonMemberReason states why a matched file is not a corpus member, for
// [UnresolvedReference]'s message: the position of the parse failure that
// kept it out, or the exclusion reason recorded against it. Exactly one
// of the two is present on a [specmodel.ReferenceNonMember], per its own
// doc comment.
func nonMemberReason(nonMember specmodel.ReferenceNonMember) string {
	if nonMember.ParseError != nil {
		return fmt.Sprintf("parse failure at line %d", nonMember.ParseError.Span.StartLine)
	}
	if nonMember.Reason != nil {
		return fmt.Sprintf("excluded: %s", *nonMember.Reason)
	}
	return "not a corpus member"
}

// UnresolvedAnchor reports a reference's anchor matching no heading of
// its target document, once the target's slug has resolved and its
// headings are known (AC-SPECANCH-002.1). Raised against the referring
// document, matching [UnresolvedReference] and every other
// reference-integrity finding — the reader's next step is to open the
// target named in the message, not the document this finding is
// attached to.
func UnresolvedAnchor(path string, at specmodel.Span, anchor, targetPath string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleUnresolvedAnchor,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("no heading named %q in %s", anchor, targetPath),
		Detail: specmodel.NewFindingDetailUnresolvedAnchor(specmodel.FindingUnresolvedAnchor{
			Anchor:     anchor,
			TargetPath: targetPath,
		}),
	}
}

// AmbiguousAnchor reports an anchor whose identifier is the identifier of
// more than one heading of its target document, naming every match
// (AC-SPECANCH-002.5). Raised against the referring document, like
// [UnresolvedAnchor].
func AmbiguousAnchor(path string, at specmodel.Span, anchor, targetPath string, matches []specmodel.AnchorMatch) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleAmbiguousAnchor,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("%q matches %d headings in %s", anchor, len(matches), targetPath),
		Detail: specmodel.NewFindingDetailAmbiguousAnchor(specmodel.FindingAmbiguousAnchor{
			Anchor:     anchor,
			TargetPath: targetPath,
			Matches:    matches,
		}),
	}
}

// RetiredTarget reports a reference from a document at a live status to
// one at a registered retired status.
//
// It carries [SeverityWarning] rather than [SeverityError]: the target
// still resolves — the reader reaches a real document that says it has
// been replaced — so failing a build on it would punish a corpus for
// retiring anything (AC-SPECREF-002.6). successor is nil where the
// target declares none; not every retirement has a successor, and a
// message that implied otherwise would send the reader looking for one.
func RetiredTarget(path string, at specmodel.Span, slug string, successor *string) specmodel.Finding {
	message := fmt.Sprintf("%q is retired", slug)
	if successor != nil {
		message += fmt.Sprintf(", superseded by %q", *successor)
	}
	return specmodel.Finding{
		Rule:     RuleRetiredTarget,
		Severity: SeverityWarning,
		Path:     path,
		Span:     at,
		Message:  message,
		Detail: specmodel.NewFindingDetailRetiredTarget(specmodel.FindingRetiredTarget{
			Slug:      slug,
			Successor: successor,
		}),
	}
}

// IncompletePairing reports one half of a paired relationship the
// resolved target does not declare the counterpart of.
//
// The caller raises this twice, once against sourcePath and once against
// targetPath, both times with the same property, counterpart, sourcePath
// and targetPath — the same structured detail from both calls, so the
// two findings describe the relationship in one direction rather than
// each from its own point of view (AC-SPECREF-002.5). at is the
// declared value's span when path is sourcePath, and the target's
// frontmatter-block span when path is targetPath: an absent property has
// no position of its own.
func IncompletePairing(path string, at specmodel.Span, property, counterpart, sourcePath, targetPath string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleIncompletePairing,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("%s declares %q naming %s, which does not declare the paired property %q back",
			sourcePath, property, targetPath, counterpart),
		Detail: specmodel.NewFindingDetailIncompletePairing(specmodel.FindingIncompletePairing{
			Property:    property,
			Counterpart: counterpart,
			SourcePath:  sourcePath,
			TargetPath:  targetPath,
		}),
	}
}

// ReferenceType reports a reference-bearing property whose resolved
// target declares a document `type` other than the one the active
// profile requires for that property (AC-SPECREF-002.10).
func ReferenceType(path string, at specmodel.Span, property, slug, expectedType, actualType string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleTypeMismatch,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("%s expects type %q, target %q declares type %q",
			property, expectedType, slug, actualType),
		Detail: specmodel.NewFindingDetailReferenceType(specmodel.FindingReferenceType{
			Property:     property,
			Slug:         slug,
			ExpectedType: expectedType,
			ActualType:   actualType,
		}),
	}
}

// UnmappedCitation reports a citation whose slug names a corpus document
// the caller's link map does not cover.
//
// It carries [SeverityError]: the slug resolves to a real document, so a
// build that renders the map's target set incomplete must not pass
// silently (AC-SPECMD-002.1). citingPath is the document that wrote the
// citation, and the finding is raised against it; mapPath is carried
// alongside so the finding is actionable without the caller having the
// map open.
func UnmappedCitation(citingPath string, at specmodel.Span, slug, mapPath string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleUnmappedCitation,
		Severity: SeverityError,
		Path:     citingPath,
		Span:     at,
		Message:  fmt.Sprintf("[[%s]] is not covered by link map %s", slug, mapPath),
		Detail: specmodel.NewFindingDetailUnmappedCitation(specmodel.FindingUnmappedCitation{
			Slug:       slug,
			CitingPath: citingPath,
			MapPath:    mapPath,
		}),
	}
}

// UnresolvedCitation reports a citation whose slug names no corpus
// document at all.
//
// It carries [SeverityWarning] rather than [SeverityError]: a slug that
// resolves to nothing is as often illustrative prose — a template
// teaching citation syntax — as it is a mistyped reference, and the two
// cannot be told apart, so a build with no other finding must not fail
// on it (AC-SPECMD-002.4). citingPath is the document that wrote the
// citation, and the finding is raised against it.
func UnresolvedCitation(citingPath string, at specmodel.Span, slug string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleUnresolvedCitation,
		Severity: SeverityWarning,
		Path:     citingPath,
		Span:     at,
		Message:  fmt.Sprintf("[[%s]] matches no corpus document", slug),
		Detail: specmodel.NewFindingDetailUnresolvedCitation(specmodel.FindingUnresolvedCitation{
			Slug:       slug,
			CitingPath: citingPath,
		}),
	}
}

// CorpusEmpty reports a check run whose corpus recorded no members: the
// configured spec root exists but discovery found no document beneath
// it to check.
//
// It carries [SeverityError], not a lesser severity, because a check
// that examined nothing is a broken configuration — a typo'd root, a
// moved directory, an over-broad ignore rule — not a corpus that
// happens to be clean; a caller must not need --strict to see it. There
// is no document to point the finding at, so Path names the spec root
// itself (relative to the repository root, matching every other
// finding's path) and Position is the zero span. root is repeated in
// Path, in the detail and quoted in Message so the finding reads the
// same whether a consumer looks at the structured field or the rendered
// text.
//
// Its detail is the corpus-level variant rather than any document-level
// one: a consumer keying off detail.kind reads a kind that matches the
// rule, and is not sent looking for a file this finding does not have.
func CorpusEmpty(root string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleCorpusEmpty,
		Severity: SeverityError,
		Path:     root,
		Message: fmt.Sprintf(
			"spec root %q holds no corpus members: the run checked zero documents", root),
		Detail: specmodel.NewFindingDetailCorpusEmpty(specmodel.FindingCorpusEmpty{
			Root: root,
		}),
	}
}

// Unverified reports a check that was requested and could not run.
//
// It carries [SeverityUnverified] rather than a warning, because a
// consumer must be able to tell "this was checked and is fine" from
// "this was not checked". Prefer [Report.Unverified], which records the
// same fact against the check's outcome as well as raising this finding;
// a bare finding leaves the report claiming the check is outstanding.
func Unverified(rule specmodel.RuleId, path string, at specmodel.Span, reason string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     rule,
		Severity: SeverityUnverified,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("check %s could not run: %s", rule, reason),
		Detail: specmodel.NewFindingDetailUnverified(specmodel.FindingUnverified{
			Reason: reason,
		}),
	}
}

// FeatureUntagged reports a Feature carrying no identifier tag
// (AC-GHK-002.2).
func FeatureUntagged(path string, at specmodel.Span, pattern string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleFeatureUntagged,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("Feature carries no identifier tag matching %q", pattern),
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Pattern: pattern,
		}),
	}
}

// CriterionUntagged reports a Scenario carrying no identifier tag
// (AC-GHK-002.2, AC-GHK-006.2).
func CriterionUntagged(path string, at specmodel.Span, pattern string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleCriterionUntagged,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("Scenario carries no identifier tag matching %q", pattern),
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Pattern: pattern,
		}),
	}
}

// IdentifierMalformed reports an identifier tag that does not match the
// pattern registered for identifiers of its kind, quoting both
// (AC-GHK-002.3, AC-GHK-006.5).
func IdentifierMalformed(path string, at specmodel.Span, identifier, pattern string) specmodel.Finding {
	id := identifier
	return specmodel.Finding{
		Rule:     RuleIdentifierMalformed,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("%q does not match pattern %q", identifier, pattern),
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Identifier: &id,
			Pattern:    pattern,
		}),
	}
}

// IdentifierMultiple reports a Feature or a Scenario carrying more than
// one tag matching its identifier pattern, naming every one of them
// (AC-GHK-002.5).
func IdentifierMultiple(path string, at specmodel.Span, tags []string, pattern string) specmodel.Finding {
	joined := strings.Join(tags, ", ")
	return specmodel.Finding{
		Rule:     RuleIdentifierMultiple,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("carries %d tags matching pattern %q: %s", len(tags), pattern, joined),
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Identifier: &joined,
			Pattern:    pattern,
		}),
	}
}

// IdentifierDuplicate reports a feature or a criterion identifier
// declared by two records in the corpus, against one of them and naming
// the other's location (AC-GHK-002.4, AC-GHK-003.4).
//
// The caller raises one of these against each record, so that checking
// either one surfaces the collision — the same shape [DuplicateSlug]
// has for two documents.
func IdentifierDuplicate(path string, at specmodel.Span, identifier, otherPath string, otherPosition specmodel.Span) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleIdentifierDuplicate,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("identifier %q is also declared by %s", identifier, otherPath),
		Detail: specmodel.NewFindingDetailDuplicateIdentifier(specmodel.FindingDuplicateIdentifier{
			Identifier: identifier,
			OtherPath:  otherPath,
			OtherSpan:  otherPosition,
		}),
	}
}

// CriterionCodeMismatch reports a criterion whose identifier's code and
// number are not its feature's (AC-GHK-003.6). pattern is the registered
// criterion pattern criterionId matches — this is not a malformed
// identifier, so the pattern is quoted for context rather than as the
// constraint that failed; the message states the derivation that did.
func CriterionCodeMismatch(path string, at specmodel.Span, criterionId, featureId, pattern string) specmodel.Finding {
	id := criterionId
	return specmodel.Finding{
		Rule:     RuleCriterionCodeMismatch,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("%q does not carry feature %q's code and number", criterionId, featureId),
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Identifier: &id,
			Pattern:    pattern,
		}),
	}
}

// OwnerMissing reports a feature in a standalone feature source carrying
// no owning-document tag (AC-GHK-002.10). Both halves of the detail are
// absent: there is no tag to quote, and a standalone feature file is no
// document, so it names none of its own to disagree with.
func OwnerMissing(path string, at specmodel.Span, ownerPrefix string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleOwnerMissing,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("carries no owning-document tag (expected one prefixed %q)", ownerPrefix),
		Detail:   specmodel.NewFindingDetailOwningDocument(specmodel.FindingOwningDocument{}),
	}
}

// OwnerMismatch reports a feature in a fenced block carrying an
// owning-document tag that names a document other than the one it is
// written in, naming both (AC-GHK-002.11).
func OwnerMismatch(path string, at specmodel.Span, taggedSlug, documentSlug string) specmodel.Finding {
	tagged, doc := taggedSlug, documentSlug
	return specmodel.Finding{
		Rule:     RuleOwnerMismatch,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("owning-document tag names %q, but is written in %q", taggedSlug, documentSlug),
		Detail: specmodel.NewFindingDetailOwningDocument(specmodel.FindingOwningDocument{
			Slug:         &tagged,
			DocumentSlug: &doc,
		}),
	}
}

// OwnerUnresolved reports an owning-document tag whose slug names no
// corpus member. The FRD requires ownership to be recorded, and a tag
// that names nothing cannot record it — this reads as the unresolved
// reference it is rather than an ordinary owner mismatch.
func OwnerUnresolved(path string, at specmodel.Span, slug string) specmodel.Finding {
	s := slug
	return specmodel.Finding{
		Rule:     RuleOwnerUnresolved,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("owning-document tag names %q, which matches no corpus member", slug),
		Detail: specmodel.NewFindingDetailOwningDocument(specmodel.FindingOwningDocument{
			Slug: &s,
		}),
	}
}

// GherkinParse reports a Gherkin source that could not be parsed, at the
// line and column of the offending text in the containing document, with
// the parser's own expectation quoted in the message (AC-GHK-006.1).
func GherkinParse(path string, at specmodel.Span, expected string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleGherkinParse,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("Gherkin source could not be parsed: %s", expected),
		Detail: specmodel.NewFindingDetailGherkinParse(specmodel.FindingGherkinParse{
			Expected: expected,
		}),
	}
}

// CriterionEmpty reports a criterion with no steps (AC-GHK-006.2). It
// carries no structured detail beyond the shared envelope: unlike a
// malformed identifier, an empty criterion has no offending value to
// quote — only its own absence, which the message states.
func CriterionEmpty(path string, at specmodel.Span) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleCriterionEmpty,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  "criterion carries no steps",
		Detail: specmodel.NewFindingDetailMissingProperty(specmodel.FindingMissingProperty{
			Property: "steps",
		}),
	}
}

// ScenarioProse reports a non-empty Scenario description as a probable
// untagged criterion — the same shape [CriterionUntagged] carries,
// because the fix this rule points toward is the same act: write the
// description as a tagged criterion of its own, or delete it if it was
// never meant to be one.
func ScenarioProse(path string, at specmodel.Span, pattern string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleScenarioProse,
		Severity: SeverityWarning,
		Path:     path,
		Span:     at,
		Message:  "Scenario description reads as a normative statement, which is a probable untagged criterion",
		Detail: specmodel.NewFindingDetailIdentifierPattern(specmodel.FindingIdentifierPattern{
			Pattern: pattern,
		}),
	}
}

// ExamplesPartiallyTagged reports a Scenario Outline whose Examples
// blocks carry an identifier on some and not others, naming the
// untagged blocks (AC-GHK-002.7, AC-GHK-006.2). Position is the first
// untagged block's own span — every one of them is carried in the
// detail, and the message points a reader at the first if they open
// nothing else.
func ExamplesPartiallyTagged(path string, untagged []specmodel.Span) specmodel.Finding {
	var at specmodel.Span
	if len(untagged) > 0 {
		at = untagged[0]
	}
	return specmodel.Finding{
		Rule:     RuleExamplesPartiallyTagged,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf(
			"Scenario Outline is partially tagged: %d Examples block(s) carry no identifier while at least one sibling does",
			len(untagged)),
		Detail: specmodel.NewFindingDetailPartiallyTaggedOutline(specmodel.FindingPartiallyTaggedOutline{
			Untagged: untagged,
		}),
	}
}

// C4Parse reports a statement of a C4 diagram the parser could not read,
// quoting the statement and the parser's reason (AC-C4-DIAG-001.1).
func C4Parse(path string, at specmodel.Span, statement, reason string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4Parse,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("cannot read %q: %s", statement, reason),
		Detail: specmodel.NewFindingDetailC4Parse(specmodel.FindingC4Parse{
			Statement: statement,
			Reason:    reason,
		}),
	}
}

// C4UnreadDiagram reports a `mermaid` block in an architecture blueprint
// whose diagram type the parser does not read, naming the type
// (AC-C4-DIAG-001.3). It carries [SeverityUnverified] under
// [RuleC4Parse]: the block was not looked inside, which is not the claim
// that it was and is sound.
func C4UnreadDiagram(path string, at specmodel.Span, diagramType string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4Parse,
		Severity: SeverityUnverified,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("diagram type %q is not a C4 diagram this check reads", diagramType),
		Detail: specmodel.NewFindingDetailC4UnreadDiagram(specmodel.FindingC4UnreadDiagram{
			DiagramType: diagramType,
		}),
	}
}

// C4UndeclaredNode reports a non-external node whose label names no
// element in the scope of its drawing document and which is not a
// passive store (AC-C4-DIAG-002.3).
func C4UndeclaredNode(path string, at specmodel.Span, alias, label string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4UndeclaredNode,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("node %q labelled %q names no element in scope", alias, label),
		Detail: specmodel.NewFindingDetailC4Node(specmodel.FindingC4Node{
			Alias:   alias,
			Label:   label,
			Matches: []string{},
		}),
	}
}

// C4AmbiguousNode reports a node whose label names more than one element
// in scope, naming every one in slug order (AC-C4-DIAG-002.4).
func C4AmbiguousNode(path string, at specmodel.Span, alias, label string, matches []string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4AmbiguousNode,
		Severity: SeverityError,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("node %q labelled %q names %d elements in scope: %s",
			alias, label, len(matches), strings.Join(matches, ", ")),
		Detail: specmodel.NewFindingDetailC4Node(specmodel.FindingC4Node{
			Alias:   alias,
			Label:   label,
			Matches: matches,
		}),
	}
}

// C4Bidirectional reports a `BiRel` statement: an arrow states one
// dependency, in one direction (AC-C4-DIAG-003.1).
func C4Bidirectional(path string, at specmodel.Span, source, target string) specmodel.Finding {
	return c4Relationship(RuleC4Bidirectional, SeverityError, path, at, source, target,
		fmt.Sprintf("BiRel(%s, %s) states no direction; draw the dependency as Rel", source, target))
}

// C4DuplicateEdge reports the later of two relationship statements in
// one diagram from the same source alias to the same target alias, after
// `Rel_Back` normalisation (AC-C4-DIAG-003.2).
func C4DuplicateEdge(path string, at specmodel.Span, source, target string) specmodel.Finding {
	return c4Relationship(RuleC4DuplicateEdge, SeverityError, path, at, source, target,
		fmt.Sprintf("%s -> %s is drawn more than once in this diagram", source, target))
}

// C4InteractionConflict reports a relationship statement tagged both
// `sync` and `async` (AC-C4-DIAG-004.2).
func C4InteractionConflict(path string, at specmodel.Span, source, target string) specmodel.Finding {
	return c4Relationship(RuleC4InteractionConflict, SeverityError, path, at, source, target,
		fmt.Sprintf("%s -> %s is tagged both sync and async", source, target))
}

// c4Relationship builds a finding carrying the relationship-defect
// detail: the statement's two aliases, in dependency direction.
func c4Relationship(rule specmodel.RuleId, severity specmodel.Severity, path string, at specmodel.Span, source, target, message string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     rule,
		Severity: severity,
		Path:     path,
		Span:     at,
		Message:  message,
		Detail: specmodel.NewFindingDetailC4RelationshipDefect(specmodel.FindingC4Relationship{
			Source: source,
			Target: target,
		}),
	}
}

// C4DrawingConflict reports a drawing of a relationship that differs in
// technology, description or tags from the relationship's first drawing
// by path and then by span, which it names (AC-C4-DIAG-003.3). It is a
// warning: the relationship is declared, and which drawing is right is
// the author's call.
func C4DrawingConflict(path string, at specmodel.Span, source, target, otherPath string, otherSpan specmodel.Span, differs string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4DrawingConflict,
		Severity: SeverityWarning,
		Path:     path,
		Span:     at,
		Message: fmt.Sprintf("%s -> %s differs in %s from its drawing at %s:%d",
			source, target, differs, otherPath, otherSpan.StartLine),
		Detail: specmodel.NewFindingDetailC4DrawingConflict(specmodel.FindingC4DrawingConflict{
			Source:    source,
			Target:    target,
			OtherPath: otherPath,
			OtherSpan: otherSpan,
		}),
	}
}

// C4ContractUnresolved reports a `contract:` tag naming a slug no member
// of the registered contract type declares (AC-C4-DIAG-004.3).
func C4ContractUnresolved(path string, at specmodel.Span, slug, source, target string) specmodel.Finding {
	return c4Contract(RuleC4ContractUnresolved, SeverityError, path, at, slug, source, target,
		fmt.Sprintf("contract:%s names no contract", slug))
}

// C4ContractMismatch reports a `contract:` tag naming a contract the
// source element does not consume or the target element does not
// provide (AC-C4-DIAG-004.4). why says which.
func C4ContractMismatch(path string, at specmodel.Span, slug, source, target, why string) specmodel.Finding {
	return c4Contract(RuleC4ContractMismatch, SeverityWarning, path, at, slug, source, target,
		fmt.Sprintf("contract:%s on %s -> %s: %s", slug, source, target, why))
}

// c4Contract builds a finding carrying the contract detail.
func c4Contract(rule specmodel.RuleId, severity specmodel.Severity, path string, at specmodel.Span, slug, source, target, message string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     rule,
		Severity: severity,
		Path:     path,
		Span:     at,
		Message:  message,
		Detail: specmodel.NewFindingDetailC4Contract(specmodel.FindingC4Contract{
			Slug:   slug,
			Source: source,
			Target: target,
		}),
	}
}

// C4TagSeparator reports a tags argument containing a comma: tags are
// joined with `+`, so a comma-joined pair reads as one unreserved tag
// (AC-C4-DIAG-004.5).
func C4TagSeparator(path string, at specmodel.Span, tags string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleC4TagSeparator,
		Severity: SeverityWarning,
		Path:     path,
		Span:     at,
		Message:  fmt.Sprintf("tags %q contain a comma; join tags with +", tags),
		Detail: specmodel.NewFindingDetailC4TagSeparator(specmodel.FindingC4TagSeparator{
			Tags: tags,
		}),
	}
}

// renderPointer renders a JSON Pointer for a message, substituting the
// root pointer for the empty string the JSON Pointer spec uses for it.
func renderPointer(pointer string) string {
	if pointer == "" {
		return rootPointer
	}
	return pointer
}

// Report accumulates one run's findings and accounts for every check the
// run knows about.
//
// The accounting is the half that is easy to omit and expensive to lack.
// A report that listed findings alone would let "the slug check passed"
// and "the slug check was never run" produce identical output, which is
// exactly the confusion AC-SPEC-006.9 exists to prevent.
//
// A Report is not safe for concurrent use; a run that checks documents in
// parallel collects into one Report from a single goroutine.
type Report struct {
	// order is the checks the report knows about, in the order they were
	// first named. It exists so a rule added by [Report.Add] for a check
	// that was never declared is still accounted for.
	order []specmodel.RuleId
	// results is the current outcome of each known check.
	results map[specmodel.RuleId]checkState
	// findings is every finding recorded, unordered until rendering.
	findings []specmodel.Finding

	// documentCount and hasDocumentCount back [Report.SetDocumentCount].
	// hasDocumentCount distinguishes "never told" from "told, and it was
	// zero" — a caller that never calls SetDocumentCount (a specs.toml
	// validation, a refs query) gets no count line from [Report.Text] at
	// all, rather than a misleading "checked 0 documents".
	documentCount    int
	hasDocumentCount bool

	// exclusions backs [Report.SetExclusions]. Nil (the default) renders
	// no excluded section, exactly like documentCount's default.
	exclusions []specmodel.IndexExclusion
}

// SetDocumentCount records how many documents the run examined, so
// [Report.Text] prints a count line — on a clean run and a failing run
// alike, which is what makes a zero-document run auditable by a human
// reading output rather than only by [RuleCorpusEmpty]'s finding.
//
// Only a corpus-wide run has a meaningful count to report; a caller
// checking one document leaves this unset.
func (r *Report) SetDocumentCount(n int) {
	r.documentCount = n
	r.hasDocumentCount = true
}

// SetExclusions records the corpus's excluded candidate files, so
// [Report.Text] lists each with its reason rather than letting a file
// excluded for missing frontmatter or an unrecognized `type` read as
// indistinguishable from one that was checked and passed.
func (r *Report) SetExclusions(exclusions []specmodel.IndexExclusion) {
	r.exclusions = exclusions
}

// checkState is the mutable half of one check's accounting.
type checkState struct {
	outcome specmodel.CheckOutcome
	reason  string
	// failed records whether any finding of severity error or warning has
	// been raised for this check, so that a later [Report.Ran] cannot
	// overwrite a failure with a pass.
	failed bool
}

// NewReport returns a report that knows about the given checks and has
// been asked to run none of them.
//
// Every named check starts as [specmodel.CheckOutcomeNotRequested], which
// is the honest starting position: nothing is known about a check until
// it has been requested and has reported.
func NewReport(known ...specmodel.RuleId) *Report {
	r := &Report{results: make(map[specmodel.RuleId]checkState, len(known))}
	for _, rule := range known {
		r.declare(rule)
	}
	return r
}

// Request marks a check as requested for this run.
//
// Until the check reports — through [Report.Ran], [Report.Add] or
// [Report.Unverified] — its outcome is [specmodel.CheckOutcomeUnverified].
// That default is deliberate and is what AC-SPEC-006.4 asks for: a check
// that starts and never finishes must not be reported as having passed,
// so silence reads as "not known" rather than as "fine".
func (r *Report) Request(rule specmodel.RuleId) {
	r.declare(rule)
	state := r.results[rule]
	if state.outcome != specmodel.CheckOutcomeNotRequested {
		return
	}
	state.outcome = specmodel.CheckOutcomeUnverified
	state.reason = "the check was requested and reported no result"
	r.results[rule] = state
}

// Ran records that a requested check completed.
//
// Its outcome becomes passed, unless a finding has already been recorded
// against it, in which case it stays failed. Calling this for a check
// that was never requested requests it first, so a caller cannot report a
// result for a check the report does not know about.
func (r *Report) Ran(rule specmodel.RuleId) {
	r.Request(rule)
	state := r.results[rule]
	if state.failed {
		state.outcome = specmodel.CheckOutcomeFailed
	} else {
		state.outcome = specmodel.CheckOutcomePassed
	}
	state.reason = ""
	r.results[rule] = state
}

// Add records findings, marking each one's check failed.
//
// A finding of severity [SeverityUnverified] is recorded as the check
// being unverified rather than failed, so that a caller which builds such
// a finding by hand still produces a coherent report. Prefer
// [Report.Unverified] for that case.
//
// A finding naming a check the report does not know about registers that
// check rather than discarding the finding: losing a finding is worse
// than an unexpected entry in the accounting.
func (r *Report) Add(findings ...specmodel.Finding) {
	for _, f := range findings {
		r.Request(f.Rule)
		state := r.results[f.Rule]
		if f.Severity == SeverityUnverified {
			state.outcome = specmodel.CheckOutcomeUnverified
			if state.reason == "" {
				state.reason = f.Message
			}
		} else {
			state.failed = true
			state.outcome = specmodel.CheckOutcomeFailed
			state.reason = ""
		}
		r.results[f.Rule] = state
		r.findings = append(r.findings, f)
	}
}

// Unverified records that a requested check could not run, both as the
// check's outcome and as one finding carrying the reason.
//
// This is the whole of the unverified mechanism. No check registered in
// this sprint calls it; it exists so that the first one that needs it
// reports through the same shape as every other check rather than
// inventing one.
func (r *Report) Unverified(rule specmodel.RuleId, path string, at specmodel.Span, reason string) {
	r.Request(rule)
	r.findings = append(r.findings, Unverified(rule, path, at, reason))
	state := r.results[rule]
	state.outcome = specmodel.CheckOutcomeUnverified
	state.reason = reason
	r.results[rule] = state
}

// declare registers a check the report does not yet know about, as not
// requested.
func (r *Report) declare(rule specmodel.RuleId) {
	if r.results == nil {
		r.results = make(map[specmodel.RuleId]checkState)
	}
	if _, known := r.results[rule]; known {
		return
	}
	r.order = append(r.order, rule)
	r.results[rule] = checkState{outcome: specmodel.CheckOutcomeNotRequested}
}

// HasErrors reports whether any finding carries [SeverityError].
//
// It is the question the process exit status is derived from, asked here
// rather than by the command surface so that every consumer of a report
// answers it the same way.
func (r *Report) HasErrors() bool {
	for _, f := range r.findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Model returns the report as the generated envelope, with both
// collections ordered deterministically and never nil.
//
// Checks are ordered by rule identifier and findings by path, then
// position, then rule, then message. Determinism is this package's
// responsibility rather than the caller's: a consumer diffing two runs
// depends on it, and so does every snapshot case.
func (r *Report) Model() specmodel.FindingReport {
	checks := make([]specmodel.CheckResult, 0, len(r.order))
	for _, rule := range r.order {
		state := r.results[rule]
		result := specmodel.CheckResult{Rule: rule, Outcome: state.outcome}
		if state.outcome == specmodel.CheckOutcomeUnverified && state.reason != "" {
			reason := state.reason
			result.Reason = &reason
		}
		checks = append(checks, result)
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Rule < checks[j].Rule })

	findings := make([]specmodel.Finding, len(r.findings))
	copy(findings, r.findings)
	sort.SliceStable(findings, func(i, j int) bool {
		return less(findings[i], findings[j])
	})

	return specmodel.FindingReport{Checks: checks, Findings: findings}
}

// less orders two findings: by path, then by where in the file they are,
// then by rule and message so that two findings at one position still
// have a fixed order.
func less(a, b specmodel.Finding) bool {
	if a.Path != b.Path {
		return a.Path < b.Path
	}
	if a.Span.Offset != b.Span.Offset {
		return a.Span.Offset < b.Span.Offset
	}
	if a.Span.Length != b.Span.Length {
		return a.Span.Length < b.Span.Length
	}
	if a.Rule != b.Rule {
		return a.Rule < b.Rule
	}
	return a.Message < b.Message
}

// JSON renders the report as the generated finding JSON.
//
// This is the machine rendering AC-SPEC-006.5 requires the command
// surface to write; the command surface selects it rather than
// formatting findings itself.
func (r *Report) JSON() ([]byte, error) {
	return json.Marshal(r.Model())
}

// Text renders the report for a human reader.
//
// When [Report.SetDocumentCount] was called, a count line opens the
// report, on a clean run and a failing run alike, followed — when
// [Report.SetExclusions] recorded any — by the excluded files and their
// reasons. Neither section appears for a report that never received
// them, so a report built for something other than a corpus scan (a
// specs.toml validation, a refs query) renders exactly as before.
//
// Findings follow, grouped under a header naming the file they are
// against — [Report.Model] already sorts them by path first, so a run of
// findings sharing a file is contiguous and this only inserts the
// header:
//
//	specs/adrs/adr-011.md:
//	  slug/pattern-mismatch  error  4:7  "adr-011-x" does not match pattern "^adr-[0-9]{3}$"
//
// The check accounting follows last, so a reader who sees no findings can
// tell at a glance which checks that verdict actually covers.
func (r *Report) Text() string {
	model := r.Model()
	var b strings.Builder

	if r.hasDocumentCount {
		fmt.Fprintf(&b, "checked %d document%s\n", r.documentCount, plural(r.documentCount))
	}
	if len(r.exclusions) > 0 {
		b.WriteString("\nexcluded:\n")
		for _, ex := range r.exclusions {
			fmt.Fprintf(&b, "  %s  %s", ex.Path, ex.Reason)
			if ex.Type != nil && *ex.Type != "" {
				fmt.Fprintf(&b, " (declared type %q)", *ex.Type)
			}
			b.WriteString("\n")
		}
	}
	if r.hasDocumentCount || len(r.exclusions) > 0 {
		b.WriteString("\n")
	}

	writeFindingsByFile(&b, model.Findings)

	if len(model.Checks) == 0 {
		return b.String()
	}
	b.WriteString("\nchecks:\n")
	for _, c := range model.Checks {
		fmt.Fprintf(&b, "  %s  %s", c.Rule, c.Outcome)
		if c.Reason != nil {
			fmt.Fprintf(&b, " (%s)", *c.Reason)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// writeFindingsByFile renders findings grouped by the file they are
// against: rule grouping (the checks section) answers "is this
// systemic", file grouping answers "what do I fix next", and the text
// report is read by whoever is fixing things.
//
// findings arrives sorted by path first, so a run of entries sharing a
// Path is already contiguous; this only inserts the header between runs.
func writeFindingsByFile(b *strings.Builder, findings []specmodel.Finding) {
	if len(findings) == 0 {
		b.WriteString("no findings\n")
		return
	}
	last := ""
	for i, f := range findings {
		if i == 0 || f.Path != last {
			if i > 0 {
				b.WriteString("\n")
			}
			fmt.Fprintf(b, "%s:\n", f.Path)
			last = f.Path
		}
		fmt.Fprintf(b, "  %s  %s  %d:%d  %s\n",
			f.Rule, f.Severity, f.Span.StartLine, f.Span.StartColumn, f.Message)
	}
}

// plural returns "" for a count of exactly one and "s" otherwise, so a
// count line reads "1 document" rather than "1 documents".
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
