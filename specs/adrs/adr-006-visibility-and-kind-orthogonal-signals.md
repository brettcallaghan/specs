---
title: "ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals"
description: Record a module's publication tier and its pattern-ness as separate frontmatter fields, not one overloaded kind
slug: adr-006
type: adr
status: active
---

# ADR-006: Model visibility and kind as orthogonal frontmatter signals

**Status:** active  
**Date:** 2026-07-29

## Context

Given the three-tier classification ([[adr-005]]), we needed to encode it in
module frontmatter. Early drafts overloaded a single `kind` field with the
values shared, internal, and pattern. An interim version used a
`shared: true|false` boolean. Both conflated independent axes. A unit's
publication status, whether it's a repeated pattern, and whether it's composite
are three separate facts. For example, a published plugin SDK is public, a
pattern, and composite at the same time. The frontmatter is also intended to
feed a queryable database of architectural elements, which needs clean columns
that each have a single meaning.

## Decision

We model these facts as orthogonal frontmatter signals:

- `visibility: public | internal` records publication, and it gates the rules
  for backwards compatibility, deprecation, and change history.
- `kind: standard | pattern` records whether the document governs one instance
  or many. It's uniform across the component and module types.
- Composite-ness is expressed by the presence of the optional Components
  roster, not by a field.

Components are always private and in-container, so they carry only `kind`.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| `kind: shared \| internal \| pattern` | One field. | Conflates the publication axis with the pattern axis.<br>Can't express a public pattern.<br>`kind` means different things at different values. | Overloaded, and doesn't give clean database columns. |
| `shared: true \| false` boolean | Simple. | Loses the distinction between published and in-repo, which governs change rules.<br>Still needs a separate pattern signal. | Too coarse. |
| Orthogonal `visibility` and `kind`; composite through section presence | Each field means one thing everywhere.<br>Clean database columns.<br>`visibility: public` is a precise anchor for change rules that you can find with `grep`. | Two fields to set instead of one.<br>The composite signal is implicit (section presence). | Not applicable; selected. |

## Consequences

**Positive:**

- `kind` means the same thing on every document type.
- `visibility` is a single anchor for change rules ("if public, do X") that you
  can query and find with `grep`.
- The element database gets clean columns that each have a single meaning.

**Negative:**

- Authors set two fields rather than one.
- Composite-ness is inferred from the presence of the Components roster rather
  than declared as an explicit boolean. This trade-off is accepted for
  simplicity.

## References

- [[adr-005]]: the tiers that this ADR encodes
- [[adr-011]]: visibility also gates the change rules
- Template: `module-blueprint-template.md`;
  `plugin/skills/spec/references/template-authoring-standards.md`
