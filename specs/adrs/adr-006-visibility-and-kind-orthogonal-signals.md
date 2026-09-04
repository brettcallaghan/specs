---
title: "ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals"
description: Record a module's publication tier and its pattern-ness as separate frontmatter fields, not one overloaded kind
slug: adr-006
type: adr
status: active
updated_at: 2026-08-05T00:00:00Z
---

# ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals

**Status:** active  
**Date:** 2026-07-29

## Context

Given the three-tier classification
([[adr-005]]), we needed to encode it in module
frontmatter. Early drafts overloaded a single `kind` field with the values
shared, internal, and pattern; an interim version used a `shared: true|false`
boolean. Both conflated independent axes. A unit's publication status, whether
it is a repeated pattern, and whether it is composite are three separate facts —
a published plugin SDK is simultaneously public, a pattern, and composite. The
frontmatter is also intended to feed a queryable database of architectural
elements, which wants clean, single-meaning columns.

## Decision

We will model these as orthogonal frontmatter signals. `visibility: public |
internal` records publication and gates the backwards-compatibility,
deprecation, and change-history rules. `kind: standard | pattern` records
whether the document governs one instance or many, and is uniform across the
component and module types. Composite-ness is expressed by the presence of the
optional Components roster, not a field. Components — always private and
in-container — carry only `kind`.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| `kind: shared \| internal \| pattern` | One field | Conflates the publication axis with the pattern axis<br>Cannot express a public pattern<br>`kind` means different things at different values | Overloaded and not database-clean |
| `shared: true \| false` boolean | Simple | Loses the published-vs-in-repo distinction that governs change rules<br>Still needs a separate pattern signal | Too coarse |
| Orthogonal `visibility` + `kind`; composite via section presence | Each field means one thing everywhere<br>Clean database columns<br>`visibility: public` is a precise, greppable change-rule anchor | Two fields to set instead of one<br>The composite signal is implicit (section presence) | N/A — selected |

## Consequences

**Positive:**
- `kind` means the same thing on every document type
- `visibility` is a single, queryable, greppable anchor for change rules ("if
  public, do X")
- The element database gets clean, single-meaning columns

**Negative:**
- Authors set two fields rather than one
- Composite-ness is inferred from the presence of the Components roster rather
  than declared as an explicit boolean (accepted for simplicity)

## References

- [[adr-005]] — the tiers this encodes
- [[adr-011]] — visibility also gates the change rules
- Template: `module-blueprint-template.md`;
  `plugin/skills/spec/references/template-authoring-standards.md`
