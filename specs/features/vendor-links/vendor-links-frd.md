---
title: "Vendored Reference Links"
description: The Spec CLI rewrites corpus wikilinks into relative markdown
  links when a document is copied out of the corpus, so a vendored copy
  cites its companions by a path that resolves where it lands
slug: frd-spec-vendor-links
type: frd
status: ready-for-review
updated_at: 2026-08-13T21:00:00Z
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
cites a document that does not ship, so every adopting repository already
holds a dangling citation.

This feature gives the Spec CLI one capability: rewrite a document's
wikilinks into relative markdown links pointing at where its companions
will sit after vendoring. The rewritten copy is self-contained — its
links resolve by path, on disk, with no index. The capability belongs to
the CLI rather than to the script that performs the copy because
deciding *what is a reference* is corpus knowledge the CLI already owns:
its extractor skips code spans, so an illustrative `` `[[module-<name>]]` ``
in a guide is not a citation. A rewriter that did not share that
judgement would corrupt the teaching prose it was meant to preserve.

## Terminology

- **Corpus document** — a document beneath the spec root, indexed, and
  citable by slug.
- **Vendored copy** — a generated copy of a corpus document written
  outside the spec root for distribution, carrying a header that marks
  it generated.
- **Vendor set** — the declared set of vendoring entries, and therefore
  the only documents a vendored copy may cite by path.
- **Vendoring entry** — one pairing of a corpus document's slug with the
  destination path its copy will occupy.
- **Wikilink** — a slug citation in `[[slug]]` or `[[slug#Heading Text]]`
  form, resolved through the index.
- **Anchor** — the heading-text fragment of a wikilink, per
  `plugin/skills/spec/references/cross-document-references.md`, which
  cites an un-promoted component as `[[container-<name>#Component Name]]`.

## Requirements

### REQ-SPECVND-001: Rewrite wikilinks to relative markdown links

The CLI converts slug citations into path citations, preserving what the
author wrote as the visible link text.

- **AC-SPECVND-001.1:** WHEN the CLI rewrites a document, it SHALL
  replace each wikilink that resolves to a document in the vendor set
  with a markdown link whose target is the path of that document's
  vendored location, relative to the rewritten document's own vendored
  location.

  Example: vendoring `ac-traceability-standards` and
  `template-authoring-standards` into the same directory, a citation of
  `[[template-authoring-standards]]` in the first becomes
  `[template-authoring-standards](template-authoring-standards.md)`.

- **AC-SPECVND-001.2:** WHEN a wikilink carries an anchor, the CLI SHALL
  preserve that anchor in the emitted link's text.

  Example: `[[container-api#Request Router]]` keeps `container-api#Request
  Router` as the visible text.

- **AC-SPECVND-001.3:** WHEN a wikilink carries an anchor, the CLI SHALL
  emit a target fragment derived from that anchor.

  Example: `[[container-api#Request Router]]` targets
  `container-api.md#request-router`.

- **AC-SPECVND-001.4:** The CLI SHALL NOT rewrite a wikilink that occurs
  inside a code span, an indented code block, or a fenced code block.

  Example: a guide explaining citation syntax writes
  `` `[[document-slug]]` `` as an illustration; it is unchanged by the
  rewrite.

- **AC-SPECVND-001.5:** The CLI SHALL NOT alter any content of the
  document other than the wikilinks it rewrites.

- **AC-SPECVND-001.6:** WHEN the CLI rewrites a document that contains no
  rewritable wikilink, it SHALL emit that document unchanged.

### REQ-SPECVND-002: Refuse a citation that leaves the vendor set

A vendored copy may only cite documents vendored alongside it. This is
the requirement that turns a packaging rule into a checkable one, and it
is enforced at rewrite time rather than discovered in an adopting
repository.

- **AC-SPECVND-002.1:** IF a wikilink resolves to a corpus document that
  is not in the vendor set, THEN the CLI SHALL report a finding.

  Example: `frontmatter-standards` cites
  `[[document-relationship-standards]]`; vendoring the first without the
  second reports that citation.

- **AC-SPECVND-002.2:** IF a wikilink resolves to a corpus document that
  is not in the vendor set, THEN the CLI SHALL NOT emit a rewritten
  document for the citing document.

- **AC-SPECVND-002.3:** WHEN reporting that finding, the CLI SHALL name
  the citing document, the unresolvable slug, and the position of the
  citation.

- **AC-SPECVND-002.4:** IF a wikilink resolves to no corpus document at
  all, THEN the CLI SHALL report a finding distinguishable from
  AC-SPECVND-002.1, because an absent document and an unvendored one
  call for different fixes.

- **AC-SPECVND-002.5:** The CLI SHALL report every such finding in one
  run rather than stopping at the first.

### REQ-SPECVND-003: Report what was rewritten

- **AC-SPECVND-003.1:** WHEN a rewrite completes, the CLI SHALL report
  how many wikilinks it rewrote in that document.

  Example: rewriting `frontmatter-standards` reports 3 rewritten
  citations.

- **AC-SPECVND-003.2:** The CLI SHALL report that count on a stream
  other than the one carrying the rewritten document, so that
  redirecting the document to a file does not capture the report.

### REQ-SPECVND-004: Rewriting is repeatable

- **AC-SPECVND-004.1:** WHEN the CLI rewrites a document whose wikilinks
  have already been rewritten, it SHALL emit that document unchanged.

  Example: running the vendoring twice with no source edit produces
  byte-identical output both times, so a repeat run cannot make a
  vendored copy differ from its source.

- **AC-SPECVND-004.2:** The CLI SHALL NOT rewrite an existing markdown
  link.

### REQ-SPECVND-005: Command surface and vendor-set input

The CLI needs two things it cannot derive from the corpus: where each
copy will land, and which copy it is rewriting right now. Relative
targets are computed between destinations, so the destinations are
input, not knowledge.

- **AC-SPECVND-005.1:** The CLI SHALL rewrite exactly one document per
  invocation.

- **AC-SPECVND-005.2:** The CLI SHALL write the rewritten document to
  standard output.

  Example: the vendoring script loops over its map, invoking the command
  once per entry and redirecting standard output to the destination.

- **AC-SPECVND-005.3:** The CLI SHALL NOT create, modify or delete a
  file at any destination path.

- **AC-SPECVND-005.4:** The CLI SHALL read the vendor set from a TOML
  file named by the caller.

- **AC-SPECVND-005.5:** The CLI SHALL NOT read the vendor set from
  `specs.toml`, which declares the corpus rather than any distribution
  of it.

- **AC-SPECVND-005.6:** The vendor set SHALL declare, for each entry,
  the slug of a corpus document and the destination path its copy will
  occupy relative to the repository root.

  Example:

  ```toml
  [[entry]]
  slug = "work-order-definition-of-ready"
  destination = "plugin/skills/groom-beads/references/shared/work-order-definition-of-ready.md"
  ```

- **AC-SPECVND-005.7:** The vendor set MAY declare more than one entry
  for the same slug, so that one document can be vendored to several
  destinations.

  Example: `work-order-definition-of-ready` ships to both `bead-bonanza`
  and `groom-beads` today.

- **AC-SPECVND-005.8:** The invocation SHALL identify which entry is
  being rewritten, so that a document vendored to several destinations
  resolves its targets from the destination it is being written to.

- **AC-SPECVND-005.9:** WHERE a cited slug has more than one entry, the
  CLI SHALL target the entry whose destination directory equals the
  citing entry's destination directory.

  Example: a copy in `groom-beads/references/shared/` citing
  `work-order-definition-of-ready` targets the copy in that same
  directory, not the one under `bead-bonanza/`.

- **AC-SPECVND-005.10:** IF a cited slug has entries but none shares the
  citing entry's destination directory, THEN the CLI SHALL report a
  finding rather than emitting a link that escapes the directory.

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
