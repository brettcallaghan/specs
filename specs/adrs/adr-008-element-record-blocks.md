---
title: "ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model"
description: Un-promoted elements carry a fenced record block validated by the same base schema a promoted document's frontmatter satisfies
slug: adr-008
type: adr
status: active
---

# ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model

**Status:** active  
**Date:** 2026-07-29

## Context

The framework aims to let structured frontmatter become a queryable database of
architectural elements. Under the promotion model ([[adr-007]]), however, an
un-promoted unit has no document and therefore no frontmatter record, so the
database would be blind to every unit that has not graduated. Lowering the
promotion threshold to give everything a document would reintroduce the
per-trivial-unit maintenance tax the promotion model exists to avoid. We needed
database completeness without coupling it to graduation.

## Decision

We will record every un-promoted child in a machine-readable **record block**:
a fenced `component` block in its parent's roster holding the element record
fields — `entity_name`, `slug`, `description`, and optionally `technology`. An
element is identified by its `slug`, and may be declared more than once: by its
blueprint, by the record block in its parent's roster, or by blocks in several
documents. Promotion copies the block's fields into the new blueprint's
frontmatter, which adds the document's own `title`, and links the blueprint
from the roster; the block may stay. Where an element has a blueprint its
values are authoritative, and declarations that disagree with them are a
finding. The element-record fields and the document-envelope fields
(`title`, status, and type-specific frontmatter) are defined as
two schemas merged with JSON Schema `allOf`, so one schema validates every
declaration of an element. The element's name is `entity_name`, not `title`:
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
- Declarations of one element can disagree, so their agreement must be
  checked or a promoted unit could keep a stale block

## References

- [[adr-007]] — the model this completes
- `plugin/skills/spec/references/template-authoring-standards.md` — Embedded
  Element Records and the two-schema rule
- `references/c4-guide.md` — Element Declarations; templates:
  `container-blueprint-template.md`, `module-blueprint-template.md`
