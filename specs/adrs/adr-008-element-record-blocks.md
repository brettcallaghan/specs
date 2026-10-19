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

The framework aims to make structured frontmatter a queryable database of
architectural elements. Under the promotion model ([[adr-007]]), though, an
un-promoted unit has no document and so no frontmatter record. The database
would then be blind to every unit that hasn't graduated. Lowering the
promotion threshold to give everything a document would bring back the
maintenance cost for each trivial unit that the promotion model exists to
avoid. We needed a complete database without tying completeness to
graduation.

## Decision

We will record every un-promoted child in a machine-readable **record
block**. A record block is a fenced `component` block in the parent's
roster. It holds the element record fields: `entity_name`, `slug`,
`description`, and optionally `technology`.

An element is identified by its `slug` and may be declared more than once:
by its blueprint, by the record block in its parent's roster, or by blocks in
several documents. Promotion copies the block's fields into the new
blueprint's frontmatter, where the document's own `title` joins them, and
links the blueprint from the roster. The block may stay. Where an element
has a blueprint, the blueprint's values are authoritative, and declarations
that disagree with them are a finding.

The element-record fields and the document-envelope fields (`title`, status,
and type-specific frontmatter) are defined as two schemas merged with JSON
Schema `allOf`, so one schema validates every declaration of an element. The
element's name is `entity_name`, not `title`: `title` names a document, and
a promoted element's blueprint carries both.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| The database covers only promoted documents | Simplest | Blind to every un-promoted unit;<br>the element inventory is incomplete | Defeats the database goal |
| Lower the promotion threshold so everything gets a document | Database-complete | Brings back the maintenance cost for each trivial unit;<br>contradicts [[adr-007]] | Wrong trade-off |
| Embedded record blocks and a two-schema `allOf` model | Database-complete regardless of graduation<br>Promotion is a mechanical copy<br>Makes the element vocabulary authoritative (feeds groom-beads) and lets tools generate diagrams from declarations | A small block to maintain for each un-promoted unit<br>Validators must parse blocks, not just frontmatter | N/A — selected |

## Consequences

**Positive:**
- Database completeness doesn't depend on whether a unit has graduated
- The C4 element vocabulary becomes authoritative, so groom-beads's C4
  collision keys can be validated precisely rather than scraped heuristically
- Diagrams can be generated from declarations, and promotion is copy-paste

**Negative:**
- A small record block to maintain for each un-promoted unit
- Tooling must extract and validate blocks in addition to frontmatter
- Declarations of one element can disagree, so tooling must check that they
  agree, or a promoted unit could keep a stale block

## References

- [[adr-007]] — the model this decision completes
- `plugin/skills/spec/references/template-authoring-standards.md` — Embedded
  Element Records and the two-schema rule
- `references/c4-guide.md` — Element Declarations; templates:
  `container-blueprint-template.md`, `module-blueprint-template.md`
