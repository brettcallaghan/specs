---
title: "FRD: Spec Corpus Markdown Links"
description: The Spec CLI rewrites corpus wikilinks into markdown links,
  targeting them from a caller-supplied map or, given none, from the
  corpus's own layout, so a copy read without an index still resolves
slug: frd-spec-markdown-links
type: frd
status: approved
updated_at: 2026-08-15T00:00:00Z
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

A note on the blocks below. A wikilink written anywhere in a Gherkin
source is recorded as a reference ([[adr-021]] constraint 14), so this
feature — whose whole subject is wikilink syntax — writes its
illustrations in backticked markdown beside each block rather than in a
step, where a literal citation would enter the index as a reference this
document never made.

### FEAT-SPECMD-001: Rewrite wikilinks to markdown links

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-001
Feature: Rewrite wikilinks to markdown links

  As an agent reading a reference shipped inside a skill
  I want every citation in it to be a link I can follow
  So that I can reach the cited document without the index it was written against

  Deciding what is a reference is corpus knowledge the CLI already owns:
  it reads references through the same extractor the check uses, which
  excludes any wikilink overlapping a code span, an indented block or a
  fenced block.

  @AC-SPECMD-001.1
  Scenario: Each targetable wikilink becomes a markdown link
    When the CLI rewrites a document
    Then the CLI replaces each wikilink it is able to target with a markdown link whose text is the wikilink's own text and whose target is the path settled for that slug

  @AC-SPECMD-001.2
  Scenario: An anchor stays in the emitted link's text
    Given a wikilink carrying an anchor
    When the CLI rewrites the document
    Then the emitted link's text keeps the anchor, so that the text remains the citation as the author wrote it

  @AC-SPECMD-001.3
  Scenario: An anchor becomes a fragment on the settled path
    Given a wikilink carrying an anchor
    When the CLI rewrites the document
    Then the CLI appends to the settled path a fragment derived from that anchor

  @AC-SPECMD-001.4 @negative
  Scenario Outline: A wikilink inside a code region is left as written
    Given a wikilink occurring inside "<region>"
    When the CLI rewrites the document
    Then the CLI does not rewrite that wikilink

    Examples:
      | region                 |
      | a code span            |
      | an indented code block |
      | a fenced code block    |

  @AC-SPECMD-001.5 @negative
  Scenario: Nothing but the rewritten wikilinks changes
    When the CLI rewrites a document
    Then the CLI alters no content of the document other than the wikilinks it rewrites

  @AC-SPECMD-001.6
  Scenario: A document with no rewritable wikilink is emitted unchanged
    Given a document containing no rewritable wikilink
    When the CLI rewrites it
    Then the CLI emits that document unchanged
```

*Examples:* with `template-authoring-standards` mapped to
`template-authoring-standards.md`, a citation of
`` `[[template-authoring-standards]]` `` becomes
`[template-authoring-standards](template-authoring-standards.md)`
(AC-SPECMD-001.1). With `container-api` mapped to `container-api.md`,
`` `[[container-api#Request Router]]` `` emits
`container-api#Request Router` as its visible text (AC-SPECMD-001.2) and
targets `container-api.md#request-router` (AC-SPECMD-001.3). A guide
explaining citation syntax writes `` `[[document-slug]]` `` as an
illustration, and the rewrite leaves it alone (AC-SPECMD-001.4).

### FEAT-SPECMD-002: Report a citation the copy cannot reach

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-002
Feature: Report a citation the copy cannot reach

  As a framework maintainer
  I want the rewrite to tell me which citations will not resolve where the copy is read
  So that I learn it before the copy reaches an adopting repository rather than after

  A citation fails to reach its target two ways, and they call for
  different answers, which is what the two rules below separate.

  A slug naming a corpus document the caller left out of the link map is
  a packaging mistake: the caller stated the shipped set and this
  citation leaves it. A slug naming no corpus document at all is usually
  not a citation — the templates and guides are written to teach citation
  syntax, so their prose is full of illustrations no author ever meant to
  resolve. Refusing those would refuse most of the documents this feature
  exists to serve. The report is still worth making: the same signal
  covers a genuine mistyped slug.

  @AC-SPECMD-002.3
  Scenario: A finding names the citation a reader must open
    When the CLI reports a finding about a citation it cannot target
    Then the CLI names the citing document, the slug, and the position of the citation

  @AC-SPECMD-002.5
  Scenario: Every such finding is reported in one run
    When the CLI rewrites a document carrying several such citations
    Then the CLI reports every one of them rather than stopping at the first

  @AC-SPECMD-002.8
  Scenario: Warnings alone do not fail the run
    Given a run whose every finding is of severity warning
    Then the CLI exits zero, so that illustrative prose does not fail a build that has nothing wrong with it

  Rule: A citation leaving a declared shipped set is a packaging mistake, and the run refuses to produce it

    @AC-SPECMD-002.1
    Scenario: A slug the map omits is an error
      Given the caller names a link map
      And a wikilink's slug names a corpus document the map does not cover
      Then the CLI reports a finding of severity error

    @AC-SPECMD-002.2 @negative
    Scenario: A slug the map omits stops the rewritten document
      Given the caller names a link map
      And a wikilink's slug names a corpus document the map does not cover
      Then the CLI does not emit a rewritten document

  Rule: A slug naming no corpus document was probably never a citation, so the run reports it and carries on

    @AC-SPECMD-002.4
    Scenario: A slug naming no corpus document is a warning
      Given a wikilink whose slug names no corpus document
      And no link map entry covers it
      Then the CLI reports a finding of severity warning, distinguishable from the one a shipped-set omission reports, because the two call for different fixes

    @AC-SPECMD-002.6
    Scenario: A slug naming no corpus document still yields a rewritten document
      Given a wikilink whose slug names no corpus document
      And no link map entry covers it
      Then the CLI emits the rewritten document notwithstanding that finding

    @AC-SPECMD-002.7
    Scenario: A slug naming no corpus document is left as the author wrote it
      Given a wikilink whose slug names no corpus document
      And no link map entry covers it
      Then the CLI leaves that wikilink exactly as the author wrote it
```

*Examples:* `frontmatter-standards` cites
`` `[[document-relationship-standards]]` ``, a corpus document under
`specs/standards/` the plugin does not ship; a link map omitting it
reports both citations as errors (AC-SPECMD-002.1). `c4-guide` writes
`` `[[component-<name>]]` `` and `` `[[module-<name>]]` `` in prose about
element records, and each is reported as a warning (AC-SPECMD-002.4).

*Note:* AC-SPECMD-002.1 is the reason a given map is never completed by
derivation — see AC-SPECMD-006.3.

### FEAT-SPECMD-003: Report what was rewritten

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-003
Feature: Report what was rewritten

  As a framework maintainer
  I want the rewrite to tell me how many citations it changed
  So that I can tell a document that was processed from one that merely had nothing to rewrite

  @AC-SPECMD-003.1
  Scenario: A completed rewrite reports how many citations it changed
    Given a document carrying six wikilinks outside code regions, citing three documents
    When the rewrite completes
    Then the CLI reports six rewritten citations for that document

  @AC-SPECMD-003.2
  Scenario: The count does not travel with the document
    Then the CLI reports that count on a stream other than the one carrying the rewritten document, so that redirecting the document to a file does not capture the report
```

### FEAT-SPECMD-004: Rewriting is repeatable

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-004
Feature: Rewriting is repeatable

  As a framework maintainer
  I want re-running the rewrite to be a no-op
  So that I can run it in a check that fails only on real drift

  @AC-SPECMD-004.1
  Scenario: A rewritten copy fed back through the rewrite comes out byte-identical
    Given a copy produced by an earlier run, whose citations are already markdown links and in which no wikilink remains
    When the CLI rewrites it
    Then the CLI emits that document unchanged

  @AC-SPECMD-004.2 @negative
  Scenario: An existing markdown link is not rewritten
    Given a document containing a markdown link
    When the CLI rewrites it
    Then the CLI does not rewrite that link
```

### FEAT-SPECMD-005: Command surface and link-map input

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-005
Feature: Command surface and link-map input

  As the author of the script that performs the copy
  I want to state the link paths myself
  So that I can write copies into a layout the CLI knows nothing about

  The CLI cannot know where a copy will live, and does not try to. The
  caller states, per slug, the path a link should use; the CLI
  substitutes it. Every question about how that path was constructed —
  absolute or relative, to what base, on a filesystem or a website —
  belongs to the caller and never reaches the CLI.

  The map is keyed by slug because that is what it is: one record per
  document. Each record may carry whatever else the caller needs about
  that document, so that adding a document to the shipped set is one edit
  to one file rather than two edits to two files that nothing keeps in
  agreement.

  @AC-SPECMD-005.1
  Scenario: One document per invocation
    Then the CLI rewrites exactly one document per invocation

  @AC-SPECMD-005.2
  Scenario: The rewritten document goes to standard output
    Then the CLI writes the rewritten document to standard output

  @AC-SPECMD-005.3 @negative
  Scenario: The rewrite touches no file
    Then the CLI creates, modifies and deletes no file

  @AC-SPECMD-005.4
  Scenario: A link map is a YAML mapping from slug to a record carrying a path
    Given the caller names a link map
    When the CLI reads it
    Then the CLI reads a YAML file: a mapping whose keys are corpus document slugs, and whose value for each slug is a record declaring, in a "path" property, the path a link to that document is to use
      """
      template-authoring-standards:
        path: template-authoring-standards.md
        destinations:
          - plugin/skills/groom-feature-blueprint/references/shared/
      ac-traceability-standards:
        path: ac-traceability-standards.md
        destinations:
          - plugin/skills/groom-feature-blueprint/references/shared/
          - plugin/skills/spec/references/shared/
      """

  @AC-SPECMD-005.5
  Scenario: A mapped path is used exactly as written
    Then the CLI uses a mapped path without resolving it, normalizing it, or making it relative to anything

  @AC-SPECMD-005.6
  Scenario: A record's other properties are the caller's business
    Then the CLI ignores any property of a record other than "path", so that the link map can carry the caller's own per-document data without the CLI having to know what it means

  @AC-SPECMD-005.7
  Scenario: A link map may be written in JSON
    Then the CLI accepts a link map written in JSON, JSON being valid YAML, so that a generated map need not be emitted in YAML syntax

  @AC-SPECMD-005.8
  Scenario: A link map that cannot be read is named
    Given a link map that cannot be read, or a record that declares no "path"
    Then the CLI reports a finding naming the fault

  @AC-SPECMD-005.9 @negative
  Scenario: A link map that cannot be read stops the rewritten document
    Given a link map that cannot be read, or a record that declares no "path"
    Then the CLI does not emit a rewritten document

  @AC-SPECMD-005.10
  Scenario: An invocation naming no link map is accepted
    Then the CLI accepts an invocation that names no link map
```

*Example:* the map above is written for copies that land beside each
other, and `destinations` belongs to the script that performs the copy.
The same documents published to a docs site are the same map with
different paths, and no other change:

```yaml
template-authoring-standards:
  path: /standards/template-authoring/
```

The copying script loops over its map, invoking the command once per
output file and redirecting standard output to it (AC-SPECMD-005.2).

### FEAT-SPECMD-006: Target the corpus's own layout when no map is given

**Acceptance Criteria:**

```gherkin
@FEAT-SPECMD-006
Feature: Target the corpus's own layout when no map is given

  As someone publishing the corpus as a browsable set of pages
  I want citations to become links between those pages
  So that I do not have to write a map whose only content is the layout the CLI can already see

  Exporting the corpus whole is the case where the caller has nothing to
  add: every cited document is present, at a path the index already
  records, and the copy keeps the corpus's shape. Deriving the target is
  correct there for exactly the reason it is wrong under a map — the
  destination is the corpus layout itself, so there is no external layout
  to guess at.

  @AC-SPECMD-006.1
  Scenario: With no map, a citation targets the cited document's corpus-relative path
    Given the caller names no link map
    When the CLI rewrites a citation
    Then the CLI targets it with the path of the cited document relative to the directory holding the citing document

  @AC-SPECMD-006.2
  Scenario: With no map, the cited document's file name is used as the corpus holds it
    Given the caller names no link map
    When the CLI rewrites a citation
    Then the CLI uses the cited document's file name as the corpus holds it, changing neither its spelling nor its extension

  @AC-SPECMD-006.3 @negative
  Scenario: Under a map, no path is ever derived
    Given the caller names a link map
    Then the CLI derives a path for no slug, so that a slug the map omits is the finding AC-SPECMD-002.1 requires rather than a link the caller never asked for
```

*Example:* this FRD cites `` `[[adr-018]]` ``, whose document is
`specs/adrs/adr-018-references-as-index-edges.md`; with no map the
emitted target is `../../adrs/adr-018-references-as-index-edges.md`
(AC-SPECMD-006.1), keeping that file name rather than one derived from
the `adr-018` slug (AC-SPECMD-006.2).

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
