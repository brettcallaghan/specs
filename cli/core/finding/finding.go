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
// No constant here names a check that reports [SeverityUnverified]. That
// mechanism exists — see [Report.Unverified] — but this sprint registers
// no check that uses it.
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
)

// referenceRules is every rule whose finding is raised against the
// document that *holds* a reference rather than the document the
// reference is about.
//
// That distinction is what a single-document check turns on: an edit to
// one document breaks references held by others, so these findings are
// reported wherever they were written rather than narrowed to the file a
// caller named (AC-SPECREF-002.9, AC-SPECREF-002.11). Membership is a
// property of where a rule's findings land, not of its `reference/`
// prefix — a future rule reported against the referenced document would
// not belong here however it is named.
var referenceRules = map[specmodel.RuleId]bool{
	RuleUnresolvedReference: true,
	RuleRetiredTarget:       true,
	RuleIncompletePairing:   true,
	RuleTypeMismatch:        true,
}

// IsReferenceRule reports whether findings of this rule are raised
// against the referring document — see [referenceRules].
func IsReferenceRule(rule specmodel.RuleId) bool {
	return referenceRules[rule]
}

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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		return fmt.Sprintf("parse failure at line %d", nonMember.ParseError.Position.StartLine)
	}
	if nonMember.Reason != nil {
		return fmt.Sprintf("excluded: %s", *nonMember.Reason)
	}
	return "not a corpus member"
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
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
		Position: at,
		Message:  fmt.Sprintf("check %s could not run: %s", rule, reason),
		Detail: specmodel.NewFindingDetailUnverified(specmodel.FindingUnverified{
			Reason: reason,
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
	if a.Position.Offset != b.Position.Offset {
		return a.Position.Offset < b.Position.Offset
	}
	if a.Position.Length != b.Position.Length {
		return a.Position.Length < b.Position.Length
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
			f.Rule, f.Severity, f.Position.StartLine, f.Position.StartColumn, f.Message)
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
