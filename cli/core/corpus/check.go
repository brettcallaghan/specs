package corpus

import (
	"fmt"
	"sync"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/schemaregistry"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Check runs, for every corpus member with parsed frontmatter, the Schema
// Registry's validation for the member's declared type and the active
// profile's per-type slug-pattern check, proving AC-SPEC-003.1,
// AC-SPEC-003.2 and AC-SPEC-003.5 at corpus level (the Schema Registry
// package proves them at unit level against a fixture schema).
//
// The result folds into one [finding.Report] together with the corpus's
// own duplicate-slug findings, so callers read one ordered list rather
// than reconciling a second channel — a defect discovered at Build and one
// discovered here are indistinguishable to a consumer.
//
// A member excluded from the corpus (no frontmatter, unknown type) or
// recorded against itself with a parse error has already been reported
// through the index's exclusions or the member's own parse_error and is
// not revisited here — there is no frontmatter to check it against.
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
	return narrowed, nil
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
