---
title: "ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model"
description: Un-promoted elements carry a fenced record block validated by the same base schema a promoted document's frontmatter satisfies
slug: adr-008
type: adr
status: active
updated_at: 2026-09-29T00:00:00Z
---

# ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model

**Status:** active  
**Date:** 2026-07-29

## Context

The framework aims to let structured frontmatter become a queryable database of
architectural elements. But under the promotion model
([[adr-007]]), an un-promoted unit has no
document — and therefore no frontmatter record — so the database would be blind
to every unit that has not graduated. Lowering the promotion threshold to give
everything a document would reintroduce the per-trivial-unit maintenance tax
the promotion model exists to avoid. We needed database completeness without
coupling it to graduation.

## Decision

We will record every un-promoted child in a machine-readable **record block**:
a fenced `component` block in its parent's roster holding the element record
fields — `entity_name`, `slug`, `description`, and optionally `technology`. A
child is in exactly one state — a record block XOR a blueprint link, never both.
Promotion lifts the block verbatim into the new blueprint's frontmatter, which
adds the document's own `title`, and replaces the block with the link, in the
same change. The element-record fields and the document-envelope fields
(`title`, status, `updated_at`, and type-specific frontmatter) are defined as
two schemas merged with JSON Schema `allOf`, so one schema validates a child
whether it is embedded or promoted. The element's name is `entity_name`, not `title`:
`title` names a document, and a promoted element's blueprint carries both.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| The database covers only promoted documents | Simplest | Blind to every un-promoted unit;<br>the element inventory is incomplete | Defeats the database goal |
| Lower the promotion threshold so everything gets a document | Database-complete | Reintroduces the per-trivial-unit maintenance tax;<br>contradicts [[adr-007]] | Wrong trade-off |
| Embedded record blocks + a two-schema `allOf` model | Database-complete regardless of graduation<br>Promotion is a mechanical copy<br>Makes the element vocabulary authoritative (feeds groom-beads) and lets diagrams be generated from declarations | A small block to maintain per un-promoted unit<br>Validators must parse blocks, not just frontmatter | N/A — selected |

## Consequences

**Positive:**
- Database completeness is decoupled from whether a unit has graduated
- The C4 element vocabulary becomes authoritative — groom-beads's C4 collision
  keys can be validated precisely rather than scraped heuristically
- Diagrams can be generated from declarations; promotion is copy-paste

**Negative:**
- A small record block to maintain per un-promoted unit
- Tooling must extract and validate blocks in addition to frontmatter
- The block-XOR-link invariant must be enforced or a promoted unit could leave
  a stale block behind

## References

- [[adr-007]] — the model this completes
- `plugin/skills/spec/references/template-authoring-standards.md` — Embedded
  Element Records and the two-schema rule
- `references/c4-guide.md` — Element Declarations; templates:
  `container-blueprint-template.md`, `module-blueprint-template.md`
