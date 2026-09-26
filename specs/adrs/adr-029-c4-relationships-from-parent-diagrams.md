---
title: "ADR-029: C4 Relationships Are Read From the Parent's Diagram"
description: The C4 model's neighbour edges are extracted from the Mermaid C4
  diagrams in the parent blueprint that owns them, by a parser this project
  writes, with reserved relationship tags carrying what Mermaid cannot express
slug: adr-029
type: adr
status: draft
updated_at: 2026-09-26T00:00:00Z
---

# ADR-029: C4 Relationships Are Read From the Parent's Diagram

**Status:** draft  
**Date:** 2026-09-26

## Context

[[frd-spec-c4]] asks the CLI to emit the corpus's C4 model — its elements
and the relationships between them — as one structured artifact. Elements
are already machine-readable: blueprint frontmatter and `component` record
blocks ([[adr-008]]) declare every node, and the index carries both.
Relationships are not. [[adr-009]] files every neighbour edge with the
parent that composes its endpoints, but the only structured form an edge
takes today is a Mermaid `Rel(...)` line in that parent's C4 diagram; the
prose paragraph beside it restates the edge for a reader and carries no
syntax a tool can rely on. [[c4-standards]] meanwhile declares prose the
declaration format, calls diagrams views of it, and says a view yields to
its declaration when the two disagree.

Three facts constrain the answer. The corpus's C4 diagrams use a small
subset of Mermaid — measured across `specs/` and the spec skill's
templates, 35 `Rel` lines and no `BiRel`, `Rel_Back`, `RelIndex`,
`$tags` or `UpdateRelStyle`; [[c4-standards]] already requires every
declaration to sit on one line. The one Go library found with a C4
parser does not handle that subset: `github.com/sammcj/mermaid-check`
v0.5.1 rejects
`Component(x, "X", "Go package (planned)", "desc")` — a parenthesis inside
a quoted argument — which fails two of this corpus's module diagrams, and
it requires Go 1.26.2 where `cli/go.mod` pins 1.24. And Mermaid itself
(12.0.0, measured with `mermaid.parse`) accepts that line, so the failure
is the library's, not the diagram's.

## Decision

We will extract neighbour edges from the Mermaid C4 diagrams in the
parent blueprint that owns them, using a C4 parser written in this
repository's CLI, and we will reserve relationship tags — `sync`,
`async` and `contract:<slug>` — for the edge properties Mermaid has no
argument for. Diagram nodes are matched to declared elements by their
display name; a node or edge that cannot be matched is reported, never
invented.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| A fenced `relationship` record block beside each roster entry | Fully declared; no diagram parsing<br>Consistent with the record block model | Every edge written twice — once for the model, once on the diagram — and the two drift<br>Contradicts [[c4-standards]], which stops short of fenced declaration blocks | Doubles the authoring cost of the one fact [[adr-009]] exists to keep single |
| Contract edges only, from `contracts_provided` and `contracts_consumed` | Already declared; needs no parser | An edge needs a consumer as well as a provider, and no blueprint in this corpus declares a consumed contract<br>The model would be edgeless | Fails the FRD's purpose |
| Parse diagrams with `github.com/sammcj/mermaid-check` | An existing AST whose shape matches the model | Fails on real corpus diagrams (parenthesis in a quoted string)<br>Forces Go 1.26.2 on the CLI module, which moves the committed coverage figures<br>Positions are diagram-relative lines, not the index's byte spans | The corpus cannot be parsed with it as released |
| Parse diagrams with a parser of our own, reserved tags for the rest | Covers exactly the subset [[c4-standards]] permits<br>Spans are the index's own<br>Parse errors become check findings at the offending line | A parser to maintain<br>Diagram lines gain model weight | N/A — selected |

## Consequences

**Positive:**
- An edge is written once, on the diagram that shows it, in the document
  [[adr-009]] makes its owner.
- Every edge in the emitted model carries the path and span it was drawn
  at, so a disputed relationship is one hop from its source line.
- The CLI gains Mermaid C4 validation as a by-product: a line the parser
  rejects, a node matching no declared element, a bidirectional arrow and
  a duplicated edge each become a finding, where today nothing checks a
  diagram at all.
- Reserved tags cost authors nothing at parse time: Mermaid 12.0.0's
  `mermaid.parse` accepts `$tags="async+contract:<slug>"`, positional
  tags, and comma-joined tags alike. Rendering a tagged diagram was not
  measured.

**Negative:**
- Diagrams stop being pure views. A `Rel` line is now the declaration of
  an edge, which inverts [[c4-standards]]' authority rule for edges while
  leaving it intact for nodes; that standard needs amending to say so.
- The parser covers only the subset the corpus uses and must grow when an
  author reaches for a macro outside it. Until then the macro is reported
  as a parse error, not skipped.
- Display names become load-bearing identifiers. Renaming an element
  without updating its diagram label unmatches its edges; the check
  reports it, but only when run.
- `Rel_Back(a, b)` is read as `b -> a`. Mermaid 12.0.0's C4 renderer
  keeps `from` and `to` as written and, for `rel_b`, sets only a start
  marker, so the arrowhead is drawn at `a`. The source text reads the
  other way round from the dependency, which a reviewer of the diff
  rather than the rendering can misread.
- Reserved tags and the `technology` property are new authoring rules.
  [[c4-standards]] and the four architecture blueprint templates must
  state them before authors can be expected to follow them.

## References

- [[frd-spec-c4]] — the requirements this decision serves
- [[adr-009]] — parents own the edges among their children
- [[adr-008]] — record blocks, which declare the nodes these edges join
- [[c4-standards]] — the one-line declaration rule this parser relies on,
  and the authority rule this decision amends for edges
- [[adr-015]] — Go as the CLI's language
- Mermaid C4 syntax: `packages/mermaid/src/docs/syntax/c4.md` in
  `mermaid-js/mermaid`
