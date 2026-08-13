---
title: "FRD: Vendored Reference Links"
description: The Spec CLI rewrites corpus wikilinks into markdown links
  using a caller-supplied slug-to-path map, so a copy taken out of the
  corpus cites its companions by a path rather than by an index it lacks
slug: frd-spec-vendor-links
type: frd
status: ready-for-review
updated_at: 2026-08-13T23:10:00Z
sources: [adr-018, frd-spec-reference-integrity]
---

# Vendored Reference Links

## Overview

Specification documents cite each other by slug — `[[frd-checkout]]` —
and that citation resolves through the corpus index. The index is built
by walking a spec root, so a slug only means something to a reader who
has the corpus.

Some documents leave the corpus. This framework ships standards and
guides inside its plugin: a copy is written into a skill's reference
directory, travels to an adopting repository, and is read there by a
human or an agent that has no index and no spec root. The wikilink
travels with the copy; the thing it resolves against does not. The
reader sees a citation that points nowhere, and nothing reports it —
`spec check` never sees the copy, because the copy is outside the
corpus by construction.

This is not hypothetical. The shipped copy of `frontmatter-standards.md`
cites `[[document-relationship-standards]]` twice. That document is a
corpus member under `specs/standards/` and is not among the copies the
plugin ships, so every adopting repository already holds two citations
that resolve to nothing.

This feature gives the Spec CLI one capability: rewrite a document's
wikilinks into markdown links, substituting a path the caller supplies
per slug. The rewritten copy is self-contained — its links are paths,
resolvable without an index. The CLI does not decide what those paths
are: whoever performs the copy knows where things land, and a CLI that
guessed would encode one layout's assumptions into everyone's. The
capability belongs to the CLI rather than to that caller because
deciding *what is a reference* is corpus knowledge the CLI already
owns: it reads references through the same extractor `spec check` uses,
which excludes any wikilink overlapping a code span, an indented block
or a fenced block. A rewriter built on a regular expression would not
draw that distinction, and would corrupt the teaching prose it was
meant to preserve.

## Terminology

- **Corpus document** — a document beneath the spec root, indexed, and
  citable by slug.
- **Vendored copy** — a copy of a corpus document written outside the
  spec root for distribution. Whatever marks the copy as generated is
  written by the caller performing the copy, not by the rewrite.
- **Link map** — the caller-supplied mapping from a document's slug to
  the path a link to it should use. It is the whole of what the CLI
  knows about where copies live: the CLI never derives, resolves or
  normalizes a path, and never learns the destination of the document it
  is rewriting. A slug absent from the map is a citation the copy may
  not make. The map may carry further per-slug data for the caller's own
  use; the CLI reads the path and ignores the rest.
- **Wikilink** — a slug citation in `[[slug]]` or `[[slug#Heading Text]]`
  form, resolved through the index.
- **Anchor** — the heading-text fragment of a wikilink, per
  `plugin/skills/spec/references/cross-document-references.md`, which
  cites an un-promoted component as `[[container-<name>#Component Name]]`.

## Requirements

### REQ-SPECVND-001: Rewrite wikilinks to markdown links

**User Story:** As an agent reading a reference shipped inside a skill,
I want every citation in it to be a link I can follow, so that I can
reach the cited document without the index it was written against.

**Acceptance Criteria:**

- **AC-SPECVND-001.1:** WHEN the CLI rewrites a document, it SHALL
  replace each wikilink whose slug appears in the link map with a
  markdown link whose text is the wikilink's own text and whose target
  is the mapped path, used verbatim.
  - *Example:* with `template-authoring-standards` mapped to
    `template-authoring-standards.md`, a citation of
    `[[template-authoring-standards]]` becomes
    `[template-authoring-standards](template-authoring-standards.md)`.
- **AC-SPECVND-001.2:** WHEN a wikilink carries an anchor, the CLI SHALL
  keep the anchor in the emitted link's text, so that the text remains
  the citation as the author wrote it.
  - *Example:* `[[container-api#Request Router]]` emits
    `container-api#Request Router` as the visible text.
- **AC-SPECVND-001.3:** WHEN a wikilink carries an anchor, the CLI SHALL
  append to the mapped path a fragment derived from that anchor.
  - *Example:* with `container-api` mapped to `container-api.md`,
    `[[container-api#Request Router]]` targets
    `container-api.md#request-router`.
- **AC-SPECVND-001.4:** The CLI SHALL NOT rewrite a wikilink that occurs
  inside a code span, an indented code block, or a fenced code block.
  - *Example:* a guide explaining citation syntax writes
    `` `[[document-slug]]` `` as an illustration; it is unchanged by the
    rewrite.
- **AC-SPECVND-001.5:** The CLI SHALL NOT alter any content of the
  document other than the wikilinks it rewrites.
- **AC-SPECVND-001.6:** WHEN the CLI rewrites a document that contains no
  rewritable wikilink, it SHALL emit that document unchanged.

### REQ-SPECVND-002: Refuse a citation the link map does not cover

**User Story:** As a framework maintainer, I want a citation of a
document I have not shipped to fail the vendoring, so that I find out
before the copy reaches an adopting repository rather than after.

A copy may only cite documents the caller has said it can reach. The
link map is that statement, so it doubles as the shipped set: omitting a
document from it makes every citation of that document a finding. This
turns a packaging rule into a checkable one, enforced at rewrite time
rather than discovered in an adopting repository.

**Acceptance Criteria:**

- **AC-SPECVND-002.1:** IF a wikilink's slug is absent from the link
  map, THEN the CLI SHALL report a finding.
  - *Example:* `frontmatter-standards` cites
    `[[document-relationship-standards]]`, a corpus document the plugin
    does not ship; a link map omitting it reports both citations.
- **AC-SPECVND-002.2:** IF a wikilink's slug is absent from the link
  map, THEN the CLI SHALL NOT emit a rewritten document.
- **AC-SPECVND-002.3:** WHEN reporting that finding, the CLI SHALL name
  the citing document, the unresolvable slug, and the position of the
  citation.
- **AC-SPECVND-002.4:** IF a wikilink's slug names no corpus document at
  all, THEN the CLI SHALL report a finding distinguishable from
  AC-SPECVND-002.1, because a slug nobody has written and a document
  deliberately left out of the map call for different fixes.
- **AC-SPECVND-002.5:** The CLI SHALL report every such finding in one
  run rather than stopping at the first.

### REQ-SPECVND-003: Report what was rewritten

**User Story:** As a framework maintainer, I want the rewrite to tell me
how many citations it changed, so that I can tell a document that was
processed from one that merely had nothing to rewrite.

**Acceptance Criteria:**

- **AC-SPECVND-003.1:** WHEN a rewrite completes, the CLI SHALL report
  how many wikilinks it rewrote in that document.
  - *Example:* `document-readiness-checklist` carries six wikilinks
    outside code regions, citing three documents; rewriting it reports
    six rewritten citations.
- **AC-SPECVND-003.2:** The CLI SHALL report that count on a stream
  other than the one carrying the rewritten document, so that
  redirecting the document to a file does not capture the report.

### REQ-SPECVND-004: Rewriting is repeatable

**User Story:** As a framework maintainer, I want re-running the
vendoring to be a no-op, so that I can run it in a check that fails only
on real drift.

**Acceptance Criteria:**

- **AC-SPECVND-004.1:** WHEN the CLI rewrites a document whose wikilinks
  have already been rewritten, it SHALL emit that document unchanged.
  - *Example:* the vendored copy of `ac-traceability-standards.md`, fed
    back through the rewrite, comes out byte-identical: its citations
    are already markdown links and no wikilink remains.
- **AC-SPECVND-004.2:** The CLI SHALL NOT rewrite an existing markdown
  link.

### REQ-SPECVND-005: Command surface and link-map input

**User Story:** As the author of the script that performs the copy, I
want to state the link paths myself, so that I can vendor into a layout
the CLI knows nothing about.

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

- **AC-SPECVND-005.1:** The CLI SHALL rewrite exactly one document per
  invocation.
- **AC-SPECVND-005.2:** The CLI SHALL write the rewritten document to
  standard output.
  - *Example:* the vendoring script loops over its map, invoking the
    command once per output file and redirecting standard output to it.
- **AC-SPECVND-005.3:** The CLI SHALL NOT create, modify or delete any
  file.
- **AC-SPECVND-005.4:** The CLI SHALL read the link map from a YAML file
  named by the caller: a mapping whose keys are corpus document slugs,
  and whose value for each slug is a record declaring, in a `path`
  property, the path a link to that document is to use.
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
- **AC-SPECVND-005.5:** The CLI SHALL use a mapped path exactly as
  written, without resolving it, normalizing it, or making it relative
  to anything.
- **AC-SPECVND-005.6:** The CLI SHALL ignore any property of a record
  other than `path`, so that the link map can carry the caller's own
  per-document data without the CLI having to know what it means.
- **AC-SPECVND-005.7:** The CLI SHALL accept a link map written in JSON,
  JSON being valid YAML, so that a generated map need not be emitted in
  YAML syntax.
- **AC-SPECVND-005.8:** IF the link map cannot be read, or a record
  declares no `path`, THEN the CLI SHALL report a finding naming the
  fault.
- **AC-SPECVND-005.9:** IF the link map cannot be read, or a record
  declares no `path`, THEN the CLI SHALL NOT emit a rewritten document.

## References

- `plugin/skills/spec/references/cross-document-references.md` — the
  wikilink and heading-anchor conventions this feature reads, and the
  rule that documents outside the specification set are cited by
  backticked path
- [[adr-018]] — references as index edges, the model that makes a slug
  resolvable in the first place
- [[frd-spec-reference-integrity]] — the reference extraction and
  resolution this feature reuses rather than reimplements, including the
  code-region exclusion AC-SPECVND-001.4 depends on
