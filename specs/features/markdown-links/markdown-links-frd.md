---
title: "FRD: Spec Corpus Markdown Links"
description: The Spec CLI rewrites corpus wikilinks into markdown links,
  targeting them from a caller-supplied map or, given none, from the
  corpus's own layout, so a copy read without an index still resolves
slug: frd-spec-markdown-links
type: frd
status: approved
updated_at: 2026-08-14T01:00:00Z
sources: [adr-018, frd-spec-reference-integrity]
---

# Spec Corpus Markdown Links

## Overview

Specification documents cite each other by slug — `[[frd-checkout]]` —
and that citation resolves through the corpus index. The index is built
by walking a spec root, so a slug only means something to a reader
holding the corpus.

Documents get read where that index is absent, in two ways. This
framework ships standards and guides inside its plugin: a copy is
written into a skill's reference directory, travels to an adopting
repository, and is read there by a human or an agent that has no index
and no spec root. A corpus published as a browsable set of pages has the
same problem from the other direction — every file is present, but a
slug is not something a markdown renderer can follow. Either way the
wikilink travels and the thing it resolves against does not. The reader
sees a citation that points nowhere, and nothing reports it — `spec
check` never sees the copy, because the copy is outside the corpus by
construction.

This is not hypothetical. The shipped copy of `frontmatter-standards.md`
cites `[[document-relationship-standards]]` twice. That document is a
corpus member under `specs/standards/` and is not among the copies the
plugin ships, so every adopting repository already holds two citations
that resolve to nothing.

This feature gives the Spec CLI one capability: rewrite a document's
wikilinks into markdown links. Where those links point is settled one of
two ways. Given a link map, the CLI substitutes the path the caller
supplies for each slug and derives nothing — whoever performs the copy
knows where things land, and a CLI that guessed would encode one
layout's assumptions into everyone's. Given no map, the corpus keeps its
own shape: each link is the cited document's path relative to the citing
document, which is what makes an exported corpus browsable without a map
that would do nothing but restate a layout the CLI already knows.

The capability belongs to the CLI rather than to that caller because
deciding *what is a reference* is corpus knowledge the CLI already
owns: it reads references through the same extractor `spec check` uses,
which excludes any wikilink overlapping a code span, an indented block
or a fenced block. A rewriter built on a regular expression would not
draw that distinction, and would corrupt the teaching prose it was
meant to preserve.

## Terminology

- **Corpus document** — a document beneath the spec root, indexed, and
  citable by slug.
- **Rewritten copy** — a copy of a corpus document whose wikilinks have
  been replaced by markdown links, written outside the spec root for
  distribution or publication. Whatever marks the copy as generated is
  written by the caller performing the copy, not by the rewrite.
- **Link map** — an optional caller-supplied mapping from a document's
  slug to the path a link to it should use. When one is given it is the
  whole of what the CLI knows about where copies live: the CLI never
  derives, resolves or normalizes a mapped path, and never learns the
  destination of the document it is rewriting. A slug absent from a
  given map is a citation the copy may not make. The map may carry
  further per-slug data for the caller's own use; the CLI reads the path
  and ignores the rest.
- **Corpus-relative target** — the path of the cited document relative
  to the directory holding the citing document, which is what a link
  targets when no map is given.
- **Wikilink** — a slug citation in `[[slug]]` or `[[slug#Heading Text]]`
  form, resolved through the index.
- **Anchor** — the heading-text fragment of a wikilink, per
  `plugin/skills/spec/references/cross-document-references.md`, which
  cites an un-promoted component as `[[container-<name>#Component Name]]`.

## Requirements

### REQ-SPECMD-001: Rewrite wikilinks to markdown links

**User Story:** As an agent reading a reference shipped inside a skill,
I want every citation in it to be a link I can follow, so that I can
reach the cited document without the index it was written against.

**Acceptance Criteria:**

- **AC-SPECMD-001.1:** WHEN the CLI rewrites a document, it SHALL
  replace each wikilink it is able to target with a markdown link whose
  text is the wikilink's own text and whose target is the path settled
  for that slug.
  - *Example:* with `template-authoring-standards` mapped to
    `template-authoring-standards.md`, a citation of
    `[[template-authoring-standards]]` becomes
    `[template-authoring-standards](template-authoring-standards.md)`.
- **AC-SPECMD-001.2:** WHEN a wikilink carries an anchor, the CLI SHALL
  keep the anchor in the emitted link's text, so that the text remains
  the citation as the author wrote it.
  - *Example:* `[[container-api#Request Router]]` emits
    `container-api#Request Router` as the visible text.
- **AC-SPECMD-001.3:** WHEN a wikilink carries an anchor, the CLI SHALL
  append to the settled path a fragment derived from that anchor.
  - *Example:* with `container-api` mapped to `container-api.md`,
    `[[container-api#Request Router]]` targets
    `container-api.md#request-router`.
- **AC-SPECMD-001.4:** The CLI SHALL NOT rewrite a wikilink that occurs
  inside a code span, an indented code block, or a fenced code block.
  - *Example:* a guide explaining citation syntax writes
    `` `[[document-slug]]` `` as an illustration; it is unchanged by the
    rewrite.
- **AC-SPECMD-001.5:** The CLI SHALL NOT alter any content of the
  document other than the wikilinks it rewrites.
- **AC-SPECMD-001.6:** WHEN the CLI rewrites a document that contains no
  rewritable wikilink, it SHALL emit that document unchanged.

### REQ-SPECMD-002: Report a citation the copy cannot reach

**User Story:** As a framework maintainer, I want the rewrite to tell me
which citations will not resolve where the copy is read, so that I learn
it before the copy reaches an adopting repository rather than after.

A citation fails to reach its target two ways, and they call for
different answers. A slug naming a corpus document the caller left out
of the link map is a packaging mistake: the caller stated the shipped
set and this citation leaves it, so the copy would carry a link that is
knowingly broken and the run refuses to produce one. It is the reason a
given map is never completed by derivation — see AC-SPECMD-006.3.

A slug naming no corpus document at all is usually not a citation. The
templates and guides are written to teach citation syntax, so their
prose is full of `[[adr-<nnn>]]` and `[[frd-checkout]]` that no author
ever meant to resolve. Refusing those would refuse most of the documents
this feature exists to serve, so the run reports each one and carries
on, leaving the wikilink as written. The report is still worth making:
the same signal covers a genuine mistyped slug.

**Acceptance Criteria:**

- **AC-SPECMD-002.1:** IF the caller names a link map AND a wikilink's
  slug names a corpus document the map does not cover, THEN the CLI
  SHALL report a finding of severity error.
  - *Example:* `frontmatter-standards` cites
    `[[document-relationship-standards]]`, a corpus document under
    `specs/standards/` the plugin does not ship; a link map omitting it
    reports both citations as errors.
- **AC-SPECMD-002.2:** IF the caller names a link map AND a wikilink's
  slug names a corpus document the map does not cover, THEN the CLI
  SHALL NOT emit a rewritten document.
- **AC-SPECMD-002.3:** WHEN reporting any such finding, the CLI SHALL
  name the citing document, the slug, and the position of the citation.
- **AC-SPECMD-002.4:** IF a wikilink's slug names no corpus document AND
  no link map entry covers it, THEN the CLI SHALL report a finding of
  severity warning, distinguishable from AC-SPECMD-002.1's, because a
  citation leaving a declared shipped set and an illustration that was
  never a citation call for different fixes.
  - *Example:* `c4-guide` writes `[[component-<name>]]` and
    `[[module-<name>]]` in prose about element records; each is reported
    as a warning.
- **AC-SPECMD-002.5:** The CLI SHALL report every such finding in one
  run rather than stopping at the first.
- **AC-SPECMD-002.6:** IF a wikilink's slug names no corpus document AND
  no link map entry covers it, THEN the CLI SHALL emit the rewritten
  document notwithstanding that finding.
- **AC-SPECMD-002.7:** IF a wikilink's slug names no corpus document AND
  no link map entry covers it, THEN the CLI SHALL leave that wikilink
  exactly as the author wrote it.
- **AC-SPECMD-002.8:** WHEN every finding a run reports is of severity
  warning, the CLI SHALL exit zero, so that illustrative prose does not
  fail a build that has nothing wrong with it.

### REQ-SPECMD-003: Report what was rewritten

**User Story:** As a framework maintainer, I want the rewrite to tell me
how many citations it changed, so that I can tell a document that was
processed from one that merely had nothing to rewrite.

**Acceptance Criteria:**

- **AC-SPECMD-003.1:** WHEN a rewrite completes, the CLI SHALL report
  how many wikilinks it rewrote in that document.
  - *Example:* `document-readiness-checklist` carries six wikilinks
    outside code regions, citing three documents; rewriting it reports
    six rewritten citations.
- **AC-SPECMD-003.2:** The CLI SHALL report that count on a stream
  other than the one carrying the rewritten document, so that
  redirecting the document to a file does not capture the report.

### REQ-SPECMD-004: Rewriting is repeatable

**User Story:** As a framework maintainer, I want re-running the rewrite
to be a no-op, so that I can run it in a check that fails only on real
drift.

**Acceptance Criteria:**

- **AC-SPECMD-004.1:** WHEN the CLI rewrites a document whose wikilinks
  have already been rewritten, it SHALL emit that document unchanged.
  - *Example:* a copy of `ac-traceability-standards.md` produced by an
    earlier run, fed back through the rewrite, comes out byte-identical:
    its citations are already markdown links and no wikilink remains.
- **AC-SPECMD-004.2:** The CLI SHALL NOT rewrite an existing markdown
  link.

### REQ-SPECMD-005: Command surface and link-map input

**User Story:** As the author of the script that performs the copy, I
want to state the link paths myself, so that I can write copies into a
layout the CLI knows nothing about.

The CLI cannot know where a copy will live, and does not try to. The
caller states, per slug, the path a link should use; the CLI substitutes
it. Every question about how that path was constructed — absolute or
relative, to what base, on a filesystem or a website — belongs to the
caller and never reaches the CLI.

The map is keyed by slug because that is what it is: one record per
document. Each record may carry whatever else the caller needs about
that document — where its copies are written, for instance — so that
adding a document to the shipped set is one edit to one file, rather
than two edits to two files that nothing keeps in agreement. The CLI
reads the path out of each record and ignores everything beside it.

**Acceptance Criteria:**

- **AC-SPECMD-005.1:** The CLI SHALL rewrite exactly one document per
  invocation.
- **AC-SPECMD-005.2:** The CLI SHALL write the rewritten document to
  standard output.
  - *Example:* the copying script loops over its map, invoking the
    command once per output file and redirecting standard output to it.
- **AC-SPECMD-005.3:** The CLI SHALL NOT create, modify or delete any
  file.
- **AC-SPECMD-005.4:** WHEN the caller names a link map, the CLI SHALL
  read it from a YAML file: a mapping whose keys are corpus document
  slugs, and whose value for each slug is a record declaring, in a
  `path` property, the path a link to that document is to use.
  - *Example:* for copies that land beside each other. `destinations`
    here belongs to the script that performs the copy; the CLI reads
    `path` alone and never interprets the rest:

    ```yaml
    template-authoring-standards:
      path: template-authoring-standards.md
      destinations:
        - plugin/skills/groom-feature-blueprint/references/shared/
    ac-traceability-standards:
      path: ac-traceability-standards.md
      destinations:
        - plugin/skills/groom-feature-blueprint/references/shared/
        - plugin/skills/spec/references/shared/
    ```

    The same documents published to a docs site are the same map with
    different paths, and no other change:

    ```yaml
    template-authoring-standards:
      path: /standards/template-authoring/
    ```
- **AC-SPECMD-005.5:** The CLI SHALL use a mapped path exactly as
  written, without resolving it, normalizing it, or making it relative
  to anything.
- **AC-SPECMD-005.6:** The CLI SHALL ignore any property of a record
  other than `path`, so that the link map can carry the caller's own
  per-document data without the CLI having to know what it means.
- **AC-SPECMD-005.7:** The CLI SHALL accept a link map written in JSON,
  JSON being valid YAML, so that a generated map need not be emitted in
  YAML syntax.
- **AC-SPECMD-005.8:** IF the link map cannot be read, or a record
  declares no `path`, THEN the CLI SHALL report a finding naming the
  fault.
- **AC-SPECMD-005.9:** IF the link map cannot be read, or a record
  declares no `path`, THEN the CLI SHALL NOT emit a rewritten document.
- **AC-SPECMD-005.10:** The CLI SHALL accept an invocation that names no
  link map.

### REQ-SPECMD-006: Target the corpus's own layout when no map is given

**User Story:** As someone publishing the corpus as a browsable set of
pages, I want citations to become links between those pages, so that I
do not have to write a map whose only content is the layout the CLI can
already see.

Exporting the corpus whole is the case where the caller has nothing to
add: every cited document is present, at a path the index already
records, and the copy keeps the corpus's shape. Deriving the target is
correct there for exactly the reason it is wrong under a map — the
destination is the corpus layout itself, so there is no external layout
to guess at.

**Acceptance Criteria:**

- **AC-SPECMD-006.1:** IF the caller names no link map, THEN the CLI
  SHALL target each rewritten citation with the path of the cited
  document relative to the directory holding the citing document.
  - *Example:* `specs/features/markdown-links/markdown-links-frd.md`
    cites `[[adr-018]]`, whose document is
    `specs/adrs/ADR-018-references-as-index-edges.md`; the emitted
    target is `../../adrs/ADR-018-references-as-index-edges.md`.
- **AC-SPECMD-006.2:** IF the caller names no link map, THEN the CLI
  SHALL use the cited document's file name as the corpus holds it,
  changing neither its spelling nor its extension.
  - *Example:* the target above keeps
    `ADR-018-references-as-index-edges.md`, rather than a name derived
    from the `adr-018` slug.
- **AC-SPECMD-006.3:** IF the caller names a link map, THEN the CLI
  SHALL NOT derive a path for any slug, so that a slug the map omits is
  the finding AC-SPECMD-002.1 requires rather than a link the caller
  never asked for.

## References

- `plugin/skills/spec/references/cross-document-references.md` — the
  wikilink and heading-anchor conventions this feature reads, and the
  rule that documents outside the specification set are cited by
  backticked path
- [[adr-018]] — references as index edges, the model that makes a slug
  resolvable in the first place
- [[frd-spec-reference-integrity]] — the reference extraction and
  resolution this feature reuses rather than reimplements, including the
  code-region exclusion AC-SPECMD-001.4 depends on
