---
title: "ADR-029: C4 Relationships Are Read From Blueprint Diagrams"
description: The C4 model's neighbor edges are extracted from the Mermaid C4
  diagrams in the architecture blueprints that ADR-009 makes their home, by
  a parser this project writes, with reserved relationship tags carrying
  what Mermaid cannot express
slug: adr-029
type: adr
status: approved
---

# ADR-029: C4 relationships are read from blueprint diagrams

**Status:** approved  
**Date:** 2026-09-26

## Context

[[frd-spec-c4]] asks the CLI to emit the corpus's C4 model (its elements
and the relationships between them) as one structured artifact.
[[frd-spec-c4-diagrams]] asks it to check the diagrams those relationships
are drawn on. Elements have a machine-readable declaration: blueprint
frontmatter, or a `component` record block ([[adr-008]]). Relationships
don't.

[[adr-009]] gives a neighbor edge two homes: the parent that composes both
endpoints, and an element's own Context view for its outward edges. In
either home, though, the only structured form the edge takes is a Mermaid
`Rel(...)` line. The prose paragraph beside it restates the edge for a
reader and carries no syntax a tool can rely on. Meanwhile,
[[c4-standards]] declares prose the declaration format, calls diagrams
views of it, and says a view yields to its declaration when the two
disagree.

Three measured facts constrain the answer:

- The subset is small. Across `specs/` and the spec skill's templates,
  `grep` finds 35 `Rel` lines, all plain `Rel`, and no `BiRel`,
  `Rel_Back`, `RelIndex`, `$tags`, or `UpdateRelStyle`. [[c4-standards]]
  already requires every declaration to sit on one line.
- The one Go library found with a C4 parser doesn't handle it.
  `github.com/sammcj/mermaid-check` v0.5.1, run over every `mermaid` block
  in the same files, returns `unrecognised C4 syntax in boundary` for two
  of this corpus's module diagrams. A minimal reproduction isolates the
  cause: `Component(x, "X", "Go package", "desc")` parses and
  `Component(x, "X", "Go package (planned)", "desc")` doesn't. Its
  `go.mod` declares `go 1.26.2`, while `cli/go.mod` declares `go 1.24`.
  Adding the dependency raised a probe module's `go` line to 1.26.2.
- Mermaid itself accepts that line. `mermaid.parse` from Mermaid 12.0.0,
  run under jsdom, accepts it and rejects a malformed control run
  alongside it. The failure is the library's, not the diagram's.

## Decision

We will extract neighbor edges from the Mermaid C4 diagrams in the
architecture blueprints, using a C4 parser written in this repository's
CLI. The diagrams read are a parent's internal views and each element's
Context view, the two homes [[adr-009]] gives an edge. We will reserve the
relationship tags `sync`, `async`, and `contract:<slug>` for the edge
properties Mermaid has no argument for. Diagram nodes are matched to
declared elements by entity name within the drawing document's scope. A
node or edge that can't be matched is reported, never invented.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| A fenced `relationship` record block beside each roster entry | Fully declared; no diagram parsing<br>Consistent with the record block model | Every edge written twice (once for the model, once on the diagram), and the two drift<br>Contradicts [[c4-standards]], which stops short of fenced declaration blocks | Doubles the authoring cost of the one fact [[adr-009]] exists to keep single |
| Read edges from the parent's internal views only | One home per edge, so no merging | A standalone container has no parent, so none of its edges would be admitted<br>Ignores the Context view [[adr-009]] assigns an element's outward edges to | Drops edges the corpus declares where [[adr-009]] says to |
| Contract edges only, from `Contract.provider` and `contracts_consumed` | Already declared; needs no parser | An edge needs a consumer as well as a provider, and no blueprint in this corpus declares a consumed contract<br>The model would be edgeless | Fails the FRD's purpose |
| Parse diagrams with `github.com/sammcj/mermaid-check` | An existing AST whose shape matches the model | Fails on real corpus diagrams (parenthesis in a quoted string)<br>Forces Go 1.26.2 on the CLI module, which moves the committed coverage figures<br>Positions are diagram-relative lines, not the index's byte spans | The corpus can't be parsed with it as released |
| Parse diagrams with a parser of our own, reserved tags for the rest | Covers exactly the subset [[c4-standards]] permits<br>Spans are the index's own<br>Parse errors become check findings at the offending line | A parser to maintain<br>Diagram lines gain model weight | N/A (selected) |

## Consequences

**Positive:**
- An edge is written on the diagram that shows it, in a document
  [[adr-009]] makes its home, and nowhere else.
- Every edge in the emitted model carries the path and span it was drawn
  at, so a disputed relationship is one hop from its source line.
- The CLI gains Mermaid C4 validation as a by-product. A line the parser
  rejects, a node matching no declared element, a bidirectional arrow, and
  a duplicated edge each become a finding, where today nothing checks a
  diagram at all.
- Reserved tags cost authors nothing at parse time: Mermaid 12.0.0's
  `mermaid.parse` accepts `$tags="async+contract:<slug>"`, positional
  tags, and comma-joined tags alike. Rendering a tagged diagram wasn't
  measured.

**Negative:**
- Diagrams stop being pure views. A `Rel` line is now the declaration of
  an edge, which inverts [[c4-standards]]' authority rule for edges while
  leaving it intact for nodes. Its "prose is the declaration format"
  statement and its model-in-prose section need amending to say so.
- An edge drawn in both of its homes is two drawings that can disagree.
  The model keeps one relationship and the check reports the
  disagreement, but nothing prevents an author from writing it.
- The parser covers only the subset the corpus uses and must grow when an
  author reaches for a macro outside it. Until then, the macro is reported
  as a parse error, not skipped.
- Display names become load-bearing identifiers. Renaming an element
  without updating its diagram label unmatches its edges. The check
  reports it, but only when run.
- `Rel_Back(a, b)` is read as `b -> a`. Mermaid 12.0.0's C4 renderer
  keeps `from` and `to` as written and, for `rel_b`, sets a start marker
  and no end marker (measured by reading its bundled renderer source). So
  the arrowhead is drawn at `a`. The source text reads the other way round
  from the dependency, and a reviewer who reads the diff rather than the
  rendering can misread it.
- Reserved tags and the `technology` property are new authoring rules.
  [[c4-standards]] and the four architecture blueprint templates must
  state them before authors can be expected to follow them.

## References

- [[frd-spec-c4]] and [[frd-spec-c4-diagrams]]: the requirements this
  decision serves
- [[adr-009]]: an edge's two homes, the parent and the Context view
- [[adr-008]]: record blocks, which declare the nodes these edges join
- [[c4-standards]]: the one-line declaration rule this parser relies on,
  and the authority rule this decision amends for edges
- [[adr-015]]: Go as the CLI's language
- Mermaid C4 syntax: `packages/mermaid/src/docs/syntax/c4.md` in
  `mermaid-js/mermaid`
