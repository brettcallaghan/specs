---
title: "FRD: Spec Corpus Anchor References"
description: Resolves the anchor half of a wikilink — the heading it
  names — so a link into a section of a document is checked as
  rigorously as a link to the document
slug: frd-spec-anchor-references
type: frd
status: draft
updated_at: 2026-09-03T00:00:00Z
sources:
  - "frd-spec-reference-integrity"
---

# Spec Corpus Anchor References

## Overview

A wikilink can name a place inside a document as well as the document
itself: `[[container-spec-cli#Element Registry]]` points at one heading,
not at the whole blueprint. [[adr-012]] mandates that form for every
un-promoted component — an element with no blueprint of its own is
referenced by a heading anchor into its parent's roster — so anchors are
how a whole class of architectural element is addressed.

[[frd-spec-reference-integrity]] resolves the slug half of a reference
and stops there, splitting on the first `#` and ignoring what follows
(AC-SPECREF-002.2). This sub-feature specifies the other half, as its
own document because the two halves need different data: slug
resolution needs only the index, anchor resolution needs the headings
[[frd-spec-structure-conformance]] extracts.

Nine references in the corpus name an anchor — four into
`container-spec-cli#Command Surface`, five into component rosters of
[[module-document-core]]. Each names one heading in full, each reaches a
heading that exists, and nothing verifies either fact. That is the case
for settling the rules now: few enough that any rule chosen here costs
nothing to adopt, established enough that the next author will copy one
of them.

A reference should behave the same way in both environments this corpus
is read in — Obsidian, which [[adr-012]] chose the notation to survive,
and a rendered markdown copy, which is how the corpus travels into
skills and repositories that do not hold it. Deriving a heading
identifier the way a renderer does is what keeps a resolved reference
clickable in both.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Extracting headings and section spans.** Specified in
  [[frd-spec-structure-conformance]]. This feature consumes that data
  and does not re-derive it.
- **Resolving the slug half of a reference.** Specified in
  [[frd-spec-reference-integrity]]. A reference whose slug resolves to
  nothing is reported there; this feature says nothing further about it,
  so one broken reference never produces two findings.
- **Nested heading paths.** `#Components#Element Registry` is one
  anchor, which names no heading and therefore reports. Resolving a path
  of nested headings is deferred.
- **Block anchors.** A caret carries no special meaning: the caret in
  `#^decision-table` disappears in normalization, leaving an ordinary
  anchor. No feature extracts block identifiers to resolve one against.
- **Naming a requirement by its code alone.** An anchor names a whole
  heading, so a requirement is cited as
  `[[frd-spec-reference-integrity#FEAT-SPECREF-002: Reference Integrity
  Checks]]`. Matching against a leading identifier token is deferred.
- **Emitting the fragment for an exported copy.** Specified in
  [[frd-spec-markdown-links]], which this feature requires to derive
  that fragment from the heading an anchor resolved to.
- **Transclusion.** An embed (`![[slug#Heading]]`) is resolved and
  checked exactly as the equivalent link, per AC-SPECANCH-002.7. Pulling
  the target's content into the referring document is a rendering
  behavior this tool does not have and does not need.
- **Rewriting an anchor when its target heading is renamed.** The tool
  reports; a reviewing agent fixes. Consistent with the parent feature's
  no-suggestions rule.
- **Anchors into documents outside the specification set.** The same
  boundary the parent feature draws.

## Terminology

- **Anchor** — the portion of a reference after the first `#`, naming a
  place inside the target document rather than the document as a whole.
- **Heading identifier** — GitHub's rule for deriving a fragment from a
  heading: its rendered text lowercased, with every character that is
  not a letter, digit, hyphen, underscore or space removed, and each
  remaining space replaced by a hyphen. Deriving one from the anchor too
  is what makes an exported copy's fragment land where the anchor did.
- **Ambiguous anchor** — an anchor whose identifier is the identifier of
  more than one heading in the target document.
- **Display alias** — Obsidian's `[[slug|label]]` form, which
  [[cross-document-references]] prohibits. A vertical bar is therefore
  not a separator here.

## Requirements

A note on the blocks below. A wikilink written anywhere in a Gherkin
source is recorded as a reference ([[adr-021]] constraint 14), so this
feature writes its notation illustrations in backticked markdown beside
each block rather than in a step, where a literal one would enter the
index as a citation this document never made.

### FEAT-SPECANCH-001: Anchor Extraction in the Index

**Acceptance Criteria:**

```gherkin
@FEAT-SPECANCH-001
Feature: Anchor Extraction in the Index

  As an architecture owner
  I want the anchor half of every reference recorded in the index
  So that a query can tell which section of a document a reference points into, not merely which document

  Record what was written. Normalization is a matching rule and not an
  indexing one: an index that normalizes on the way in cannot report the
  text a fix has to edit.

  @AC-SPECANCH-001.1
  Scenario: An anchor is recorded alongside its reference's slug and span
    Given a reference carrying an anchor
    When the CLI builds the index
    Then the CLI records the anchor alongside that reference's slug and span

  @AC-SPECANCH-001.2
  Scenario: An anchor is recorded as it was authored
    Then the CLI records an anchor as it was authored, without normalization

  @AC-SPECANCH-001.5 @negative
  Scenario: A vertical bar does not divide a reference
    Given a reference containing a vertical bar
    When the CLI builds the index
    Then the CLI records the vertical bar as ordinary text of the slug or the anchor containing it

  @AC-SPECANCH-001.6
  Scenario: An anchor in a frontmatter property is recorded like any other
    Given a reference-bearing frontmatter property whose value carries an anchor
    When the CLI builds the index
    Then the CLI records that anchor exactly as it records an anchor written in the body
```

*Examples:* `` `[[container-spec-cli#Element Registry]]` `` records slug
`container-spec-cli` and anchor `Element Registry` on one reference
entry, at one span — not two references (AC-SPECANCH-001.1).
`#Element  Registry` keeps its double space, because the index has to
report the text a fix will edit (AC-SPECANCH-001.2).
`` `[[adr-012|the reference conventions ADR]]` `` records the whole of
`adr-012|the reference conventions ADR` as its slug, which matches no
corpus member and reports as unresolved — the alias
[[cross-document-references]] prohibits reaches nothing rather than
resolving quietly (AC-SPECANCH-001.5). `sources: ["adr-012#Decision"]`
records an anchor exactly as a body citation of the same target does: a
citation names a section wherever it is written (AC-SPECANCH-001.6).

*Rationale:* AC-SPECANCH-001.2 keeps the index a record of what the
author wrote — the parent feature records a reference's text and span
before anything resolves it (AC-SPECREF-001.1), and every question about
what an anchor matches is answered later, against corpus knowledge the
index does not hold. AC-SPECANCH-001.5 is stated as a prohibition
because the obvious handling of an alias, splitting it off and resolving
the rest, would quietly legalize a notation
[[cross-document-references]] forbids.

### FEAT-SPECANCH-002: Anchor Resolution Checks

**Acceptance Criteria:**

```gherkin
@FEAT-SPECANCH-002
Feature: Anchor Resolution Checks

  As a document author
  I want an anchor that names no real heading reported
  So that a reference into a section survives the section being renamed, or tells me it did not

  An anchor and a heading are compared as identifiers derived by the
  rule a markdown renderer applies, so a reference this feature resolves
  is a reference an exported copy can target.

  This feature reports errors and unverified findings only. A reader
  either reaches the section named or reaches nothing, and where the
  tool cannot yet see the headings it says so rather than guessing.

  @AC-SPECANCH-002.1
  Scenario: An anchor matching no heading is an error naming the document and the anchor
    Given a reference naming an anchor
    And the target document's slug resolves
    And the target's headings are available
    And no heading matches
    When the CLI checks the reference
    Then the CLI reports a finding of severity error naming the target document and the anchor

  @AC-SPECANCH-002.2
  Scenario: An anchor is matched on heading identifiers
    Then the CLI matches an anchor against a heading by deriving a heading identifier from each and comparing them

  @AC-SPECANCH-002.4
  Scenario: An anchor whose target headings are unavailable is unverified
    Given a reference naming an anchor
    And the target document's headings are not available
    When the CLI checks the reference
    Then the CLI reports that reference with severity unverified naming what is missing
    And the CLI does not report the anchor as resolved

  @AC-SPECANCH-002.5
  Scenario: An ambiguous anchor is an error naming every match
    Given an anchor whose identifier is the identifier of more than one heading in the target document
    When the CLI checks the reference
    Then the CLI reports a finding of severity error naming every matching heading and its span
    And the CLI does not report the anchor as resolved

  @AC-SPECANCH-002.6
  Scenario: An anchor with no slug resolves against the containing document
    Given a reference naming an anchor and no slug
    When the CLI resolves it
    Then the CLI resolves it against the containing document

  @AC-SPECANCH-002.7
  Scenario: An embed is resolved and checked as the equivalent link
    Given an embed
    When the CLI checks it
    Then the CLI resolves and checks it exactly as it resolves and checks the equivalent link

  @AC-SPECANCH-002.8 @negative
  Scenario: One mistake produces one finding
    Given a reference whose slug resolves to no corpus member
    When the CLI checks it
    Then the CLI reports no anchor finding for it

  @AC-SPECANCH-002.9
  Scenario: Every anchored reference in the corpus is checked
    When the CLI checks the corpus
    Then the CLI checks the anchor of every reference carrying one
    And the CLI reports an anchor matching no heading and an ambiguous anchor under stable rule identifiers of their own
```

*Examples:* `reference/unresolved-anchor  error … "no heading named
'Componets' in container-spec-cli"` states the fact and proposes no
correction, per the parent feature's rule (AC-SPECANCH-002.1).
`#element registry` matches `### Element Registry` because both yield
`element-registry`, which is also the fragment an exported copy targets
(AC-SPECANCH-002.2). Until [[frd-spec-structure-conformance]] supplies
headings there is nothing to match against, so every anchor reference
reports `unverified` per AC-SPEC-006.3 and AC-SPEC-006.4 — the same path
a target whose frontmatter failed to parse takes, and what lets this
feature ship first (AC-SPECANCH-002.4). [[module-document-core]] declares
two `References` headings, so `` `[[module-document-core#References]]` ``
names two sections and resolves to neither; the author disambiguates by
renaming one of them (AC-SPECANCH-002.5). `` `[[#Terminology]]` `` in
this FRD resolves to this FRD's own Terminology heading, where a
reference to a document whose slug is empty would report as broken every
time (AC-SPECANCH-002.6). `` `![[container-spec-cli#Element Registry]]` ``
produces the same finding as the same reference without the leading `!`
(AC-SPECANCH-002.7). `` `[[adr-099#Decision]]` `` where no `adr-099`
exists produces one finding, the parent's unresolved-slug error, because
two findings for one mistake trains a reader to skim
(AC-SPECANCH-002.8).

*Rationale:* AC-SPECANCH-002.2 adopts a normalization that already
exists rather than inventing one. A heading identifier is what a
rendered copy uses as a fragment, so matching on it collapses two
guarantees into one: an anchor this feature resolves is an anchor
[[frd-spec-markdown-links]] can target. Case-insensitivity follows from
the same rule.

AC-SPECANCH-002.5 offers no notation for selecting one of several
identical headings. A renderer disambiguates by position — the second
`References` becomes `references-1` — and a citation written that way
says nothing about which section was meant, breaks when a heading is
inserted above it, and obliges the author to count sections. Reporting
the ambiguity puts the fix where it lasts, in the heading.

AC-SPECANCH-002.9 makes the anchor rules part of an ordinary corpus
check rather than an option, each under its own rule identifier so a
caller can select it.

One consequence constrains this framework's most likely use of anchors:
a reference to a requirement names its heading in full, and the terser
`#FEAT-SPECREF-002` resolves to nothing. IF the terse form is wanted,
THEN the fix is to change what the heading says — a template decision
for [[frd-spec-structure-conformance]]'s required headings — rather than
a resolver only this tool understands, which would fail the moment an
author clicked the link in Obsidian.

### FEAT-SPECANCH-003: Anchor Impact Queries

**Acceptance Criteria:**

```gherkin
@FEAT-SPECANCH-003
Feature: Anchor Impact Queries

  As an authoring agent
  I want to ask what references point into a section before I rename or move it
  So that I learn the blast radius from the tool rather than from a reader who followed a broken link

  @AC-SPECANCH-003.1
  Scenario: A query qualified by an anchor lists every reference into that heading
    When the CLI is asked for the references to a document qualified by an anchor
    Then the CLI lists every reference whose anchor resolves to that heading, each with its referring document and span

  @AC-SPECANCH-003.2
  Scenario: A single-document check reports inbound anchor references that no longer resolve
    When the CLI checks a single document
    Then the CLI reports inbound anchor references to that document that no longer resolve, each cited at the span in the referring document

  @AC-SPECANCH-003.3
  Scenario: An anchor nothing references is an answer rather than a failure
    Given no reference names the given anchor
    When the query runs
    Then the CLI reports that
    And the CLI exits zero

  @AC-SPECANCH-003.4
  Scenario: An ambiguous reference is listed for every heading it matches
    Given a reference whose anchor matches more than one heading in the target document
    When the CLI is asked for the references to that document qualified by one of those headings
    Then the CLI lists that reference, marked as matching more than one heading
```

*Examples:* `spec refs 'container-spec-cli#Element Registry'` lists the
documents that would break if that heading were renamed — the question
[[adr-012]] created by making anchors the address of every un-promoted
component (AC-SPECANCH-003.1). An agent renames a `### Element Registry`
heading to `### Registry` and the write-time check on that one file
reports the four documents whose anchors now point at nothing, at their
own lines; this is AC-SPECREF-002.9 extended to the anchor half
(AC-SPECANCH-003.2). An empty answer is an answer, not a failure,
consistent with AC-SPECREF-003.5 and AC-SPEC-007.2 (AC-SPECANCH-003.3).
An ambiguous reference resolves to no heading, and the author renaming
one of the headings it matches is exactly the reader who has to see it,
so the query lists it rather than hiding it behind its own error
(AC-SPECANCH-003.4).

*Rationale:* AC-SPECANCH-003.2 is what makes anchors safe to adopt at
all. Without it a heading rename is the edit the tooling cannot see: the
author changes one document, every check against that document passes,
and the breakage lives entirely in files they did not touch. The parent
feature answers the same question for slugs the same way.

## References

- [[frd-spec-reference-integrity]] — the parent feature, which resolves
  the slug half of every reference and defers the anchor half here
- [[frd-spec-structure-conformance]] — supplies the headings and section
  spans this feature resolves an anchor against
- [[frd-spec-validation-cli]] — the index, finding engine, severities,
  and query surface both features extend
- [[frd-spec-markdown-links]] — rewrites a citation into a markdown
  link, and derives that link's fragment from the heading this feature
  resolved
- [[adr-012]] — the wikilink and heading-anchor conventions, and the
  requirement that un-promoted components be addressed by anchor
- [[cross-document-references]] — the reference notation this feature
  resolves, including the display alias it prohibits
- [[frd-spec-slug-rename]] — the sibling sub-feature, whose rewrites
  preserve the anchors this feature resolves
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
- `plugin/skills/spec/references/c4-guide.md` — element declarations and
  the roster headings an element anchor resolves against
