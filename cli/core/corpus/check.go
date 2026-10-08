package corpus

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/brettcallaghan/agent-plugins/cli/core/envelope"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/references"
	"github.com/brettcallaghan/agent-plugins/cli/core/schemaregistry"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Check runs, for every corpus member with parsed frontmatter, the Schema
// Registry's validation for the member's declared type and the active
// profile's per-type slug-pattern check, proving AC-SPEC-003.1,
// AC-SPEC-003.2 and AC-SPEC-003.5 at corpus level (the Schema Registry
// package proves them at unit level against a fixture schema). It also
// reports every reference the graph recorded as unresolved or as reaching
// a non-member — [finding.RuleUnresolvedReference], one rule whether the
// reference was written in a document's body, held by a registered
// frontmatter property, or matched a file that exists but is not a corpus
// member (AC-SPECREF-002.1, AC-SPECREF-002.3, AC-SPECREF-002.7). The one
// reference it does not report is the slug-less anchor form whose anchor
// resolved against the document holding it (AC-SPECREF-002.12).
//
// The result folds into one [finding.Report] together with the corpus's
// own duplicate-slug findings, so callers read one ordered list rather
// than reconciling a second channel — a defect discovered at Build and one
// discovered here are indistinguishable to a consumer.
//
// A member excluded from the corpus (no frontmatter, unknown type) has
// already been reported through the index's exclusions and is not
// revisited here: a file with no frontmatter is not claiming to be a
// specification, so it has nothing to conform to.
//
// A member recorded with a parse error is the opposite case and IS
// reported here, as [finding.RuleFrontmatterParse]. That document did
// claim to be a specification and got the claim wrong. Skipping it —
// which this function did until agent-plugins-5xo — left the run
// reporting every rule as `passed` for the one document in the corpus
// nothing had been able to read, so the worse a document's frontmatter
// was, the quieter the tool got. The index recording the parse error is
// not a substitute: `specs check` never shows the caller the index.
//
// An error is returned only when the active profile's schema does not
// compile; a document's own defects are always findings, never errors.
func (c *Corpus) Check() ([]specmodel.Finding, error) {
	return c.check(nil)
}

// CheckFor narrows [Corpus.Check] to one file, resolving against the full
// corpus exactly as [Corpus.FindingsFor] does (AC-SPEC-001.2): the
// corpus's own findings — duplicate slugs among them — still see every
// member, while schema and slug-pattern checking runs only for the named
// file, which is what keeps a single-document check from paying for the
// whole corpus.
//
// Every finding is narrowed to the named document, with no exception:
// what an edit breaks elsewhere is answered by a corpus check, or by
// `refs` before the edit (AC-SPECREF-002.9). The rules that need the
// whole corpus still run over it, so a finding located in the named
// document — an unresolved, retired-target, type-mismatched or
// incompletely paired reference out of it, a duplicate identifier or slug
// it shares with another document — is reported exactly as a corpus check
// reports it (AC-SPECREF-002.11).
//
// It returns [ErrNotAMember] for a path no corpus member occupies, so
// that a mistyped path, a file outside the spec root and a file the
// index excluded are all distinguishable from a document that passed.
// Selecting nothing and reporting every rule passed is the vacuous
// success this tool exists to catch elsewhere (`agent-plugins-ltd`).
//
// path may be absolute or repository-relative. On Windows it may use
// either separator; on other platforms a backslash is an ordinary
// character in a filename and is left alone.
func (c *Corpus) CheckFor(path string) ([]specmodel.Finding, error) {
	wanted := c.normalize(path)
	if _, member := c.byPath[wanted]; !member {
		return nil, c.notAMember(wanted)
	}
	return c.check(&wanted)
}

// ErrNotAMember reports that a path a caller named is not a corpus
// member. Callers test for it with [errors.Is]; the wrapped error carries
// the path and, where the index recorded one, the reason it was excluded.
var ErrNotAMember = errors.New("not a corpus member")

// notAMember builds the error [Corpus.CheckFor] returns for a path that
// occupies no entry, distinguishing the two cases the index can already
// tell apart: a candidate that was discovered and excluded — where the
// recorded reason is the answer the caller needs, and the same reason the
// index reports — and a path discovery never saw at all.
func (c *Corpus) notAMember(normalized string) error {
	for _, excluded := range c.index.Exclusions {
		if excluded.Path != normalized {
			continue
		}
		if excluded.Type != nil && *excluded.Type != "" {
			return fmt.Errorf("corpus: %s is %w: %s (declared type %q)",
				normalized, ErrNotAMember, excluded.Reason, *excluded.Type)
		}
		return fmt.Errorf("corpus: %s is %w: %s",
			normalized, ErrNotAMember, excluded.Reason)
	}
	return fmt.Errorf("corpus: %s is %w: no document at that path was discovered under the spec root",
		normalized, ErrNotAMember)
}

// check is the shared implementation behind Check and CheckFor. only,
// when non-nil, is a normalized path: schema and slug-pattern checking run
// for that member alone, and the result is narrowed to it after folding.
func (c *Corpus) check(only *string) ([]specmodel.Finding, error) {
	registry, err := c.schemaRegistry()
	if err != nil {
		return nil, err
	}

	report := finding.NewReport()
	report.Add(c.findings...)

	// A whole-corpus run that indexed no members at all is a broken
	// configuration, not a pass: nothing below has anything to check
	// against, so every rule would otherwise report "passed" for a run
	// that examined zero documents (agent-plugins-qz9.2). This is
	// deliberately narrowed to only == nil: CheckFor resolves a single
	// named member against the same index, and a one-member selection is
	// not what "empty corpus" means.
	if only == nil && len(c.index.Entries) == 0 {
		report.Add(finding.CorpusEmpty(c.normalize(c.specRoot)))
	}

	for _, entry := range c.index.Entries {
		if only != nil && entry.Path != *only {
			continue
		}
		if entry.Frontmatter == nil {
			// No parsed envelope, so nothing below can run against this
			// document. Report why rather than passing over it in
			// silence; see this function's doc comment.
			if entry.ParseError != nil {
				report.Add(finding.FrontmatterParse(
					entry.Path, entry.ParseError.Span, entry.ParseError.Message))
			}
			continue
		}
		doc, ok := c.document(entry.Path)
		if !ok {
			continue
		}

		docType := string(envelope.Of(entry.Frontmatter).Type)

		violations, err := registry.Validate(entry.Path, docType, doc)
		if err != nil {
			return nil, fmt.Errorf("corpus: checking %s against the schema registry: %w", entry.Path, err)
		}

		var mismatch *specmodel.Finding
		if rules, known := c.types[docType]; known && rules.SlugPattern != "" {
			mismatch, err = schemaregistry.CheckSlugPattern(
				entry.Path, slugSpan(doc), string(envelope.Of(entry.Frontmatter).Slug), rules.SlugPattern)
			if err != nil {
				return nil, fmt.Errorf("corpus: checking the slug pattern for %s: %w", entry.Path, err)
			}
		}
		if mismatch != nil {
			// One slug mistake is one finding. The type's pattern is the
			// stricter statement of what the slug must be (AC-SPEC-003.5),
			// so the shared schema's pattern failure on the same slug is
			// not reported beside it — under --only frontmatter/schema
			// too, since the fault is slug/pattern-mismatch's. A slug that
			// passes its type's pattern keeps any schema finding it earns.
			violations = withoutSlugPatternViolation(violations)
		}
		report.Add(violations...)
		if mismatch != nil {
			report.Add(*mismatch)
		}
	}

	// Every reference the graph recorded as unresolved or as reaching a
	// non-member is reported here, regardless of only: the references are
	// already computed and held in memory, so processing all of them costs
	// nothing extra, and the narrowing below scopes the result to one file
	// exactly as it already does for the corpus's own duplicate-slug
	// findings (AC-SPECREF-002.1, AC-SPECREF-002.3, AC-SPECREF-002.7).
	for _, ref := range c.index.References {
		if f, ok := unresolvedReferenceFinding(ref); ok {
			report.Add(f)
		}
		report.Add(anchorFindings(ref)...)
	}

	// The relational half of reference integrity — what a resolved
	// reference *means* rather than whether it lands — walks the same
	// already-resolved reference collection rather than re-walking the
	// corpus. It is not, though, a lookup over data already in memory on
	// a warm corpus: retiredTargetFindings' successorOf and
	// pairingIncompleteFindings each call Corpus.document for a resolved
	// reference's target, which on a warm corpus reads and parses that
	// one target off disk (and caches it for later callers). All three
	// checks below also run corpus-wide, unnarrowed by only, exactly as
	// the unresolved-reference loop above does (AC-SPECREF-002.5,
	// AC-SPECREF-002.6, AC-SPECREF-002.10).
	report.Add(c.retiredTargetFindings()...)
	report.Add(c.pairingIncompleteFindings()...)
	report.Add(c.typeMismatchFindings()...)

	// The Gherkin identity (VTAG) and content (VCON) checks share one
	// recomputed structure from source — see [Corpus.recomputeGherkin] —
	// rather than reading from the index or from c.findings, so a Check or
	// CheckFor call parses the corpus's Gherkin once for both rather than
	// twice. Both run here regardless of only: gherkin/identifier-duplicate
	// needs the whole corpus even on a single-document check, gherkin/parse
	// can only be found by a full recompute in the first place (a source
	// that fails to parse produces no feature and so is invisible to any
	// narrower walk), and the narrowing below scopes every other Gherkin
	// finding to one file exactly as it does for every other rule.
	rules := c.profile.FeatureTags()
	vocab := newVocabulary(rules)
	gherkinFeatures, gherkinFailures, err := c.recomputeGherkin()
	if err != nil {
		return nil, err
	}
	report.Add(c.gherkinIdentityFindings(vocab, rules, gherkinFeatures)...)
	report.Add(gherkinContentFindings(rules, gherkinFeatures, gherkinFailures)...)

	// The C4 checks read the element set from the index and each
	// drawing document's diagrams from its source; resolving a node needs
	// the whole corpus's elements, so the set is built whatever only
	// names, and the narrowing below scopes the findings to one file.
	elements, err := c.C4Elements()
	if err != nil {
		return nil, fmt.Errorf("corpus: reading C4 diagrams: %w", err)
	}
	report.Add(c4NodeFindings(elements)...)
	report.Add(c4RelationshipFindings(elements)...)

	findings := report.Model().Findings
	if only == nil {
		return findings, nil
	}

	// A single-document check reports the named document's own findings
	// and nothing located elsewhere (AC-SPECREF-002.9).
	var narrowed []specmodel.Finding
	for _, f := range findings {
		if f.Path == *only {
			narrowed = append(narrowed, f)
		}
	}
	return narrowed, nil
}

// unresolvedReferenceFinding derives the reference/unresolved finding a
// reference's own recorded resolution calls for, or reports none for a
// reference that resolved to a corpus member or is a resolved
// self-anchor (AC-SPECREF-002.12).
func unresolvedReferenceFinding(ref specmodel.Reference) (specmodel.Finding, bool) {
	if references.ResolvedSelfAnchor(ref) {
		return specmodel.Finding{}, false
	}
	if nonMember, ok := ref.Resolution.AsNonMember(); ok {
		// A reference that reached a member whose frontmatter could not be
		// read depends on that member's parse failure, which is already
		// reported against it. The citing document is not what is wrong,
		// so it raises nothing (AC-SPEC-003.8, AC-SPECREF-002.7).
		if nonMember.ParseError != nil {
			return specmodel.Finding{}, false
		}
		return finding.UnresolvedReference(ref.Path, ref.Span, ref.Text, ref.Slug, ref.Origin, &nonMember), true
	}
	if _, ok := ref.Resolution.AsUnresolved(); ok {
		return finding.UnresolvedReference(ref.Path, ref.Span, ref.Text, ref.Slug, ref.Origin, nil), true
	}
	return specmodel.Finding{}, false
}

// anchorFindings derives the findings a reference's own recorded anchor
// resolution calls for: none where the anchor resolved, or where there
// is no anchor resolution at all (AC-SPECANCH-002.9).
//
// An absent resolution is the one-mistake-one-finding rule already
// applied: a reference carrying no anchor has nothing to check, and one
// whose slug did not resolve is already reported as an unresolved
// reference by the loop this is called from, so an anchor finding on top
// of it would be a second finding for one mistake (AC-SPECANCH-002.8).
//
// Like the unresolved-reference derivation above, this is a lookup over
// what resolution already settled: it opens no document, re-reads no
// heading and re-derives no identifier, so it costs nothing on the warm
// single-document path that the relational checks below do pay for.
//
// Every finding is raised against the REFERRING document at the
// reference's own span, which is the document a reader must open to fix
// it — the target named in the message is where they look next
// (AC-SPECANCH-003.2).
func anchorFindings(ref specmodel.Reference) []specmodel.Finding {
	if ref.AnchorResolution == nil || ref.Anchor == nil {
		return nil
	}
	targetPath, checkable := references.AnchorTarget(ref)
	if !checkable {
		return nil
	}

	switch {
	case isAnchorResolved(*ref.AnchorResolution):
		return nil
	case isAnchorUnresolved(*ref.AnchorResolution):
		return []specmodel.Finding{
			finding.UnresolvedAnchor(ref.Path, ref.Span, *ref.Anchor, targetPath),
		}
	}
	if ambiguous, ok := ref.AnchorResolution.AsAmbiguous(); ok {
		return []specmodel.Finding{
			finding.AmbiguousAnchor(ref.Path, ref.Span, *ref.Anchor, targetPath, ambiguous.Matches),
		}
	}
	if unverified, ok := ref.AnchorResolution.AsUnverified(); ok {
		// The tool could not look, which is not the claim that it looked
		// and found nothing (AC-SPECANCH-002.4). It is reported under
		// the same rule identifier the anchor check owns, at
		// [finding.SeverityUnverified], so a caller reads "this was not
		// checked" rather than "this is fine".
		return []specmodel.Finding{
			finding.Unverified(finding.RuleUnresolvedAnchor, ref.Path, ref.Span, unverified.Reason),
		}
	}
	return nil
}

// isAnchorResolved and isAnchorUnresolved name the two variants the
// switch above tests without binding a payload it does not read.
func isAnchorResolved(resolution specmodel.AnchorResolution) bool {
	_, ok := resolution.AsResolved()
	return ok
}

func isAnchorUnresolved(resolution specmodel.AnchorResolution) bool {
	_, ok := resolution.AsUnresolved()
	return ok
}

// retiredTargetFindings reports, for every resolved reference from a
// document at a status the active profile's lifecycle registers as live
// to one at a registered retired status, a warning against the
// referring document at the reference's own position, naming the
// retired target's slug and its successor value where the target
// declares one (AC-SPECREF-002.6).
//
// A profile that registers no lifecycle — [corpus.LifecycleRules]'s zero
// value — has an empty live set and an empty retired set, so this loop
// runs its full length and matches nothing rather than being skipped: a
// corpus whose documents have no lifecycle is a legitimate registration,
// not a reason to special-case the walk.
func (c *Corpus) retiredTargetFindings() []specmodel.Finding {
	lifecycle := c.profile.Lifecycle()
	live := stringSet(lifecycle.LiveStatuses)
	retired := stringSet(lifecycle.RetiredStatuses)

	var findings []specmodel.Finding
	for _, ref := range c.index.References {
		resolved, ok := ref.Resolution.AsResolved()
		if !ok {
			continue
		}
		referrer, ok := c.Entry(ref.Path)
		if !ok || referrer.Frontmatter == nil || !live[string(envelope.Of(referrer.Frontmatter).Status)] {
			continue
		}
		target, ok := c.Entry(resolved.TargetPath)
		if !ok || target.Frontmatter == nil || !retired[string(envelope.Of(target.Frontmatter).Status)] {
			continue
		}
		successor := c.successorOf(resolved.TargetPath, lifecycle.SuccessorProperty)
		findings = append(findings, finding.RetiredTarget(
			ref.Path, ref.Span, string(envelope.Of(target.Frontmatter).Slug), successor))
	}
	return findings
}

// successorOf reads the text of the successor property the active
// profile registers, from the document at path, returning nil where the
// profile registers no successor property or the document does not
// declare it — nothing requires a retired document to declare one.
func (c *Corpus) successorOf(path, property string) *string {
	if property == "" {
		return nil
	}
	doc, ok := c.document(path)
	if !ok {
		return nil
	}
	value, ok := propertyText(doc, property)
	if !ok {
		return nil
	}
	return &value
}

// pairingIncompleteFindings reports, for every pairing the active
// profile registers whose declaring document names a resolved target
// that does not declare the paired counterpart in return, one finding
// against each of the two documents, carrying the same structured detail
// (AC-SPECREF-002.5). The finding against the declaring document sits at
// the declared value's span; the finding against the target sits at the
// span of its frontmatter block, because an absent property has no
// position of its own.
//
// A type that registers no pairing contributes nothing here, which is
// this loop finding zero matches over real data rather than being
// skipped — the same reasoning [retiredTargetFindings] documents.
func (c *Corpus) pairingIncompleteFindings() []specmodel.Finding {
	var findings []specmodel.Finding
	for _, ref := range c.index.References {
		resolved, ok := ref.Resolution.AsResolved()
		if !ok {
			continue
		}
		origin, ok := ref.Origin.AsFrontmatterProperty()
		if !ok {
			continue
		}
		declarer, ok := c.Entry(ref.Path)
		if !ok || declarer.Frontmatter == nil {
			continue
		}
		rules, known := c.types[string(envelope.Of(declarer.Frontmatter).Type)]
		if !known {
			continue
		}
		for _, pairing := range rules.Pairings {
			if pairing.Property != origin.Property {
				continue
			}
			targetDoc, ok := c.document(resolved.TargetPath)
			if !ok {
				continue
			}
			if _, declared := targetDoc.Property(pairing.Counterpart); declared {
				continue
			}
			findings = append(findings,
				finding.IncompletePairing(ref.Path, ref.Span,
					pairing.Property, pairing.Counterpart, ref.Path, resolved.TargetPath),
				finding.IncompletePairing(resolved.TargetPath, targetDoc.Region.Inner,
					pairing.Property, pairing.Counterpart, ref.Path, resolved.TargetPath),
			)
		}
	}
	return findings
}

// typeMismatchFindings reports, for every reference-bearing property the
// active profile requires a target document `type` for, a resolved
// reference whose target declares a different one (AC-SPECREF-002.10).
//
// A type that registers no typed reference contributes nothing here, for
// the same reason [pairingIncompleteFindings] does not special-case an
// empty registration.
func (c *Corpus) typeMismatchFindings() []specmodel.Finding {
	var findings []specmodel.Finding
	for _, ref := range c.index.References {
		resolved, ok := ref.Resolution.AsResolved()
		if !ok {
			continue
		}
		origin, ok := ref.Origin.AsFrontmatterProperty()
		if !ok {
			continue
		}
		declarer, ok := c.Entry(ref.Path)
		if !ok || declarer.Frontmatter == nil {
			continue
		}
		rules, known := c.types[string(envelope.Of(declarer.Frontmatter).Type)]
		if !known {
			continue
		}
		expected, typed := rules.TypedReferenceProperties[origin.Property]
		if !typed {
			continue
		}
		target, ok := c.Entry(resolved.TargetPath)
		if !ok || target.Frontmatter == nil {
			continue
		}
		actual := string(envelope.Of(target.Frontmatter).Type)
		if actual == expected {
			continue
		}
		findings = append(findings, finding.ReferenceType(
			ref.Path, ref.Span, origin.Property, ref.Slug, expected, actual))
	}
	return findings
}

// stringSet converts values to a lookup set, so a status membership test
// reads as a map access rather than a linear scan repeated per reference.
func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

// propertyText returns the text of one frontmatter property: its parsed
// value where that is a string, and its raw source text otherwise — the
// same fallback declaredType applies for a document's type, so a value
// that is not a plain string is still reported as what the author wrote
// rather than silently dropped.
func propertyText(doc *frontmatter.Document, name string) (string, bool) {
	property, ok := doc.Property(name)
	if !ok {
		return "", false
	}
	if value, isString := property.Value.(string); isString {
		return value, true
	}
	return property.Raw, true
}

// registryState is the lazily-built, once-compiled Schema Registry for one
// Corpus's active profile.
type registryState struct {
	once     sync.Once
	registry *schemaregistry.Registry
	err      error
}

// schemaRegistry returns this corpus's Schema Registry, compiling the
// active profile's union schema on first use and reusing the compilation
// afterward — the same once-per-branch caching [schemaregistry.Registry]
// itself applies to individual branches, extended here to the registry as
// a whole so that a repeated Check or CheckFor call does not recompile the
// schema every time.
func (c *Corpus) schemaRegistry() (*schemaregistry.Registry, error) {
	c.registryState.once.Do(func() {
		c.registryState.registry, c.registryState.err = schemaregistry.New(c.profile.Schema())
		if c.registryState.err != nil {
			c.registryState.err = fmt.Errorf("corpus: compiling the active profile's schema: %w", c.registryState.err)
		}
	})
	return c.registryState.registry, c.registryState.err
}

// withoutSlugPatternViolation returns violations less any schema
// violation of the slug's pattern: the finding a slug/pattern-mismatch on
// the same slug already states more precisely.
func withoutSlugPatternViolation(violations []specmodel.Finding) []specmodel.Finding {
	kept := violations[:0:0]
	for _, violation := range violations {
		if detail, ok := violation.Detail.AsSchemaViolation(); ok &&
			violation.Rule == finding.RuleSchemaViolation &&
			detail.Pointer == "/slug" && strings.HasPrefix(detail.Constraint, "pattern") {
			continue
		}
		kept = append(kept, violation)
	}
	return kept
}
