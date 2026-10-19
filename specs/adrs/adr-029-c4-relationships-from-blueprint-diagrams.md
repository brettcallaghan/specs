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

[[frd-spec-c4]] asks the CLI to emit the corpus's C4 model—its elements and
the relationships between them—as one structured artifact.
[[frd-spec-c4-diagrams]] asks the CLI to check the diagrams that those
relationships are drawn on. Elements have a machine-readable declaration:
blueprint frontmatter, or a `component` record block ([[adr-008]]).
Relationships don't. [[adr-009]] gives a neighbor edge two homes: the parent
that composes both endpoints, and an element's own Context view for its
outward edges. In either home, the only structured form that the edge takes is
a Mermaid `Rel(...)` line. The prose paragraph beside the line restates the
edge for a reader and carries no syntax that a tool can rely on. Meanwhile,
[[c4-standards]] declares prose the declaration format, calls diagrams views
of it, and says that a view yields to its declaration when the two disagree.

Three measured facts constrain the answer:

- **The subset is small.** Across `specs/` and the spec skill's templates,
  `grep` finds 35 `Rel` lines, all plain `Rel`, and no `BiRel`, `Rel_Back`,
  `RelIndex`, `$tags`, or `UpdateRelStyle`. [[c4-standards]] already requires
  every declaration to sit on one line.
- **The one Go library found with a C4 parser doesn't handle the subset.**
  Run over every `mermaid` block in the same files,
  `github.com/sammcj/mermaid-check` v0.5.1 returns
  `unrecognised C4 syntax in boundary` for two of this corpus's module
  diagrams. A minimal reproduction isolates the cause:
  `Component(x, "X", "Go package", "desc")` parses, and
  `Component(x, "X", "Go package (planned)", "desc")` doesn't. The library's
  `go.mod` declares `go 1.26.2`, `cli/go.mod` declares `go 1.24`, and adding
  the dependency raised a probe module's `go` line to 1.26.2.
- **Mermaid itself accepts that line.** Run under jsdom, `mermaid.parse` from
  Mermaid 12.0.0 accepts the line and rejects a malformed control. The failure
  is the library's, not the diagram's.

## Decision

We extract neighbor edges from the Mermaid C4 diagrams in the architecture
blueprints by using a C4 parser written in this repository's CLI. The parser
reads a parent's internal views and each element's Context view, which are the
two homes that [[adr-009]] gives an edge. We reserve the relationship tags
`sync`, `async`, and `contract:<slug>` for the edge properties that Mermaid
has no argument for. The parser matches diagram nodes to declared elements by
entity name within the drawing document's scope. It reports a node or edge
that it can't match, and never invents one.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| A fenced `relationship` record block beside each roster entry | Fully declared, with no diagram parsing<br>Consistent with the record block model | Every edge is written twice, once for the model and once on the diagram, and the two drift<br>Contradicts [[c4-standards]], which stops short of fenced declaration blocks | Doubles the authoring cost of the one fact that [[adr-009]] exists to keep single |
| Read edges from the parent's internal views only | One home per edge, so no merging | A standalone container has no parent, so none of its edges are admitted<br>Ignores the Context view that [[adr-009]] assigns an element's outward edges to | Drops edges that the corpus declares where [[adr-009]] says to |
| Contract edges only, from `Contract.provider` and `contracts_consumed` | Already declared, and needs no parser | An edge needs a consumer as well as a provider, and no blueprint in this corpus declares a consumed contract<br>The model is edgeless | Fails the FRD's purpose |
| Parse diagrams with `github.com/sammcj/mermaid-check` | An existing AST whose shape matches the model | Fails on real corpus diagrams (a parenthesis in a quoted string)<br>Forces Go 1.26.2 on the CLI module, which moves the committed coverage figures<br>Positions are diagram-relative lines, not the index's byte spans | The released version can't parse the corpus |
| Parse diagrams with a parser of our own, and reserve tags for the rest | Covers exactly the subset that [[c4-standards]] permits<br>Spans are the index's own<br>Parse errors become check findings at the offending line | A parser to maintain<br>Diagram lines gain model weight | N/A (selected) |

## Consequences

**Positive:**

- An edge is written only on the diagram that shows it, in a document that
  [[adr-009]] makes its home.
- Every edge in the emitted model carries the path and span that it was drawn
  at, so a disputed relationship is one hop from its source line.
- The CLI gains Mermaid C4 validation as a by-product. Each of the following
  becomes a finding: a line that the parser rejects, a node that matches no
  declared element, a bidirectional arrow, and a duplicated edge. Today,
  nothing checks a diagram at all.
- Reserved tags cost authors nothing at parse time: Mermaid 12.0.0's
  `mermaid.parse` accepts `$tags="async+contract:<slug>"`, positional tags,
  and comma-joined tags alike. Rendering a tagged diagram wasn't measured.

**Negative:**

- Diagrams stop being pure views. A `Rel` line is now the declaration of an
  edge, which inverts the authority rule in [[c4-standards]] for edges while
  leaving it intact for nodes. That standard's "prose is the declaration
  format" statement and its model-in-prose section need amending to say so.
- An edge drawn in both of its homes is two drawings that can disagree. The
  model keeps one relationship and the check reports the disagreement, but
  nothing prevents an author from creating it.
- The parser covers only the subset that the corpus uses, and must grow when
  an author reaches for a macro outside it. Until then, the parser reports the
  macro as a parse error instead of skipping it.
- Display names become load-bearing identifiers. If you rename an element
  without updating its diagram label, its edges no longer match. The check
  reports the mismatch, but only when it runs.
- `Rel_Back(a, b)` is read as `b -> a`. Mermaid 12.0.0's C4 renderer keeps
  `from` and `to` as written and, for `rel_b`, sets a start marker and no end
  marker, so the arrowhead is drawn at `a`. This was measured by reading the
  renderer source bundled with Mermaid. The source text reads in the opposite
  direction from the dependency, which a reviewer of the diff, rather than of
  the rendering, can misread.
- Reserved tags and the `technology` property are new authoring rules.
  [[c4-standards]] and the four architecture blueprint templates must state
  them before authors can be expected to follow them.

## References

- [[frd-spec-c4]] and [[frd-spec-c4-diagrams]]: the requirements that this
  decision serves.
- [[adr-009]]: an edge's two homes, which are the parent and the Context view.
- [[adr-008]]: record blocks, which declare the nodes that these edges join.
- [[c4-standards]]: the one-line declaration rule that this parser relies on,
  and the authority rule that this decision amends for edges.
- [[adr-015]]: Go as the CLI's language.
- Mermaid C4 syntax: `packages/mermaid/src/docs/syntax/c4.md` in
  `mermaid-js/mermaid`.
