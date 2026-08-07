package corpus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
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
// member (AC-SPECREF-002.1, AC-SPECREF-002.3, AC-SPECREF-002.7).
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
// not a substitute: `spec check` never shows the caller the index.
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
// It additionally reports every inbound reference to path that no longer
// resolves, against the referring document rather than path itself
// (AC-SPECREF-002.9) — see [Corpus.brokenInboundReferences].
//
// path may be absolute or repository-relative, and may use either
// platform's separator.
func (c *Corpus) CheckFor(path string) ([]specmodel.Finding, error) {
	wanted := c.normalize(path)
	return c.check(&wanted)
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
					entry.Path, entry.ParseError.Position, entry.ParseError.Message))
			}
			continue
		}
		doc, ok := c.document(entry.Path)
		if !ok {
			continue
		}

		docType := string(entry.Frontmatter.Type)

		violations, err := registry.Validate(entry.Path, docType, doc)
		if err != nil {
			return nil, fmt.Errorf("corpus: checking %s against the schema registry: %w", entry.Path, err)
		}
		report.Add(violations...)

		if rules, known := c.types[docType]; known && rules.SlugPattern != "" {
			mismatch, err := schemaregistry.CheckSlugPattern(
				entry.Path, slugSpan(doc), string(entry.Frontmatter.Slug), rules.SlugPattern)
			if err != nil {
				return nil, fmt.Errorf("corpus: checking the slug pattern for %s: %w", entry.Path, err)
			}
			if mismatch != nil {
				report.Add(*mismatch)
			}
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
	}

	// The relational half of reference integrity — what a resolved
	// reference *means* rather than whether it lands — walks the same
	// already-resolved reference collection, for the same reason: the
	// work is a lookup over data already in memory, not a second corpus
	// walk (AC-SPECREF-002.5, AC-SPECREF-002.6, AC-SPECREF-002.10).
	report.Add(c.retiredTargetFindings()...)
	report.Add(c.pairingIncompleteFindings()...)
	report.Add(c.typeMismatchFindings()...)

	findings := report.Model().Findings
	if only == nil {
		return findings, nil
	}

	var narrowed []specmodel.Finding
	for _, f := range findings {
		if f.Path == *only {
			narrowed = append(narrowed, f)
		}
	}
	// AC-SPECREF-002.9: a single-document check additionally reports every
	// inbound reference to *only that no longer resolves, against the
	// referring document rather than the one named — that is the file a
	// reader must open to fix it.
	narrowed = append(narrowed, c.brokenInboundReferences(*only)...)
	return narrowed, nil
}

// unresolvedReferenceFinding derives the reference/unresolved finding a
// reference's own recorded resolution calls for, or reports none for a
// reference that resolved to a corpus member.
func unresolvedReferenceFinding(ref specmodel.Reference) (specmodel.Finding, bool) {
	if nonMember, ok := ref.Resolution.AsNonMember(); ok {
		return finding.UnresolvedReference(ref.Path, ref.Position, ref.Text, ref.Slug, ref.Origin, &nonMember), true
	}
	if _, ok := ref.Resolution.AsUnresolved(); ok {
		return finding.UnresolvedReference(ref.Path, ref.Position, ref.Text, ref.Slug, ref.Origin, nil), true
	}
	return specmodel.Finding{}, false
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
		if !ok || referrer.Frontmatter == nil || !live[string(referrer.Frontmatter.Status)] {
			continue
		}
		target, ok := c.Entry(resolved.TargetPath)
		if !ok || target.Frontmatter == nil || !retired[string(target.Frontmatter.Status)] {
			continue
		}
		successor := c.successorOf(resolved.TargetPath, lifecycle.SuccessorProperty)
		findings = append(findings, finding.RetiredTarget(
			ref.Path, ref.Position, string(target.Frontmatter.Slug), successor))
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
		rules, known := c.types[string(declarer.Frontmatter.Type)]
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
				finding.IncompletePairing(ref.Path, ref.Position,
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
		rules, known := c.types[string(declarer.Frontmatter.Type)]
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
		actual := string(target.Frontmatter.Type)
		if actual == expected {
			continue
		}
		findings = append(findings, finding.ReferenceType(
			ref.Path, ref.Position, origin.Property, ref.Slug, expected, actual))
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

// brokenInboundReferences reports, for a single-document check of path,
// every inbound reference the warm index still records as resolved to it
// that no longer does — the graph's own edges were resolved when the
// corpus was built, but path's frontmatter may have changed since (the
// scenario AC-SPECREF-002.9 names: an agent renames a document's slug and
// the write-time check on that one file is what has to catch the
// referrers it just broke).
//
// This costs one lookup (InboundReferences, already built) and, when
// there is anything to compare against, one read of path itself — the
// same read a single-document check already pays for elsewhere — rather
// than a corpus walk, which is what keeps it inside AC-SPEC-009.2's
// single-document budget. A referrer whose reference still names path's
// current slug is left alone; every other one is reported against the
// referrer, at the referrer's own span, exactly as an ordinary unresolved
// reference is.
func (c *Corpus) brokenInboundReferences(path string) []specmodel.Finding {
	inbound := c.InboundReferences(path)
	if len(inbound) == 0 {
		return nil
	}

	currentSlug, haveSlug := c.currentSlug(path)

	var findings []specmodel.Finding
	for _, ref := range inbound {
		if haveSlug && slugsMatch(ref.Slug, currentSlug) {
			continue
		}
		findings = append(findings, finding.UnresolvedReference(
			ref.Path, ref.Position, ref.Text, ref.Slug, ref.Origin, nil))
	}
	return findings
}

// currentSlug reads the slug the document at path currently declares,
// through Corpus.document rather than the persisted index entry — the
// whole point being to see path's frontmatter as it stands now, which may
// differ from what the graph's edges were resolved against.
func (c *Corpus) currentSlug(path string) (string, bool) {
	doc, ok := c.document(path)
	if !ok {
		return "", false
	}
	return propertyText(doc, slugProperty)
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

// slugsMatch compares two slugs the way reference resolution itself does:
// trimmed and case-insensitively, so this check and the graph it is
// reading agree about what "the same slug" means (AC-SPECREF-002.8).
func slugsMatch(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
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
