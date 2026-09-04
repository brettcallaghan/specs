---
title: "FRD: Spec Corpus Anchor References"
description: Resolves the anchor half of a wikilink — heading, nested
  heading path, and block reference — so a link into a section of a
  document is checked as rigorously as a link to the document
slug: frd-spec-anchor-references
type: frd
status: draft
updated_at: 2026-08-15T00:00:00Z
sources:
  - "frd-spec-reference-integrity"
---

# Spec Corpus Anchor References

## Overview

A wikilink can name a place inside a document as well as the document
itself: `[[container-spec-cli#Element Registry]]` points at one heading,
not at the whole blueprint. [[adr-012]] makes this the mandated form for
every un-promoted component — an element with no blueprint of its own is
referenced by a heading anchor into its parent's roster — so anchors are
not an optional flourish in this framework, they are how a whole class
of architectural element is addressed at all.

[[frd-spec-reference-integrity]] resolves the slug half of a reference
and deliberately stops there, splitting on the first `#` and ignoring
what follows (AC-SPECREF-002.2). This sub-feature specifies the other
half. It exists as its own document because the two halves have
different dependencies and different urgency: slug resolution needs only
the index, while anchor resolution needs the heading data that
[[frd-spec-structure-conformance]] extracts, and a scan of this corpus
on 2026-08-06 found 256 body references of which **zero** name an
anchor. The 20 anchor-shaped strings in the corpus are all backticked
notation examples in [[adr-012]] and this FRD's own siblings —
illustrations of the convention, not uses of it, and excluded from the
count by the same rule (AC-SPECREF-001.2) that keeps them out of the
index.

That number is the reason this is a sub-feature rather than an
afterthought. A convention nobody has used yet is a convention with no
accumulated breakage to clean up and no installed base to keep
compatible, which makes now the cheapest moment to decide exactly what
an anchor means. It is also a warning: the un-promoted-component form
[[adr-012]] mandates has never actually been written, so the first
author to write one will be discovering the rules. This document is what
they will be discovering.

The wider goal is that the framework's references behave the way an
Obsidian author expects, because that is the environment [[adr-012]]
chose the notation to survive. Anchors are the largest remaining gap
between "we borrowed Obsidian's brackets" and "we support Obsidian's
links".

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Extracting headings and section spans.** Specified in
  [[frd-spec-structure-conformance]]. This feature consumes that data
  and does not re-derive it.
- **Resolving the slug half of a reference.** Specified in
  [[frd-spec-reference-integrity]]. A reference whose slug resolves to
  nothing is reported there; this feature says nothing further about it,
  so one broken reference never produces two findings.
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
- **Heading anchor** — an anchor naming heading text, optionally as a
  path of nested headings separated by further `#` characters.
- **Block anchor** — an anchor of the form `^identifier`, naming a
  specific block rather than a heading.
- **Anchor path** — the ordered sequence of heading texts in a nested
  anchor. `#Components#Element Registry` has the path
  `["Components", "Element Registry"]`.
- **Display text** — the portion of a reference after a `|`, shown to a
  reader in place of the link target. It is never part of the anchor.
- **Normalized heading text** — heading text compared with letter case
  ignored and runs of whitespace collapsed to a single space, the form
  in which an anchor is matched against a heading.

## Requirements

A note on the blocks below. A wikilink written anywhere in a Gherkin
source is recorded as a reference ([[adr-021]] constraint 14), so this
feature — whose subject is wikilink anchors — writes its notation
illustrations in backticked markdown beside each block rather than in a
step, where a literal one would enter the index as a citation this
document never made.

### FEAT-SPECANCH-001: Anchor Extraction in the Index

**Acceptance Criteria:**

```gherkin
@FEAT-SPECANCH-001
Feature: Anchor Extraction in the Index

  As an architecture owner
  I want the anchor half of every reference recorded in the index
  So that a query can tell which section of a document a reference points into, not merely which document

  Record what was written, and do not let a part of the syntax that is
  about presentation leak into the part that is about identity.
  Normalization is a matching rule, not an indexing one.

  @AC-SPECANCH-001.1
  Scenario: An anchor is recorded alongside its reference's slug and span
    Given a reference carrying an anchor
    When the CLI builds the index
    Then the CLI records the anchor alongside that reference's slug and span

  @AC-SPECANCH-001.2
  Scenario: An anchor is recorded as it was authored
    Then the CLI records an anchor as it was authored, without normalization

  @AC-SPECANCH-001.3
  Scenario: A nested anchor is recorded as an ordered path
    Given an anchor naming a path of nested headings
    When the CLI builds the index
    Then the CLI records the path as an ordered sequence

  @AC-SPECANCH-001.4
  Scenario: An anchor beginning with a caret is a block anchor
    Given an anchor beginning with a caret
    When the CLI builds the index
    Then the CLI records it as a block anchor, distinguishable in the index from a heading anchor

  @AC-SPECANCH-001.5 @negative
  Scenario: Display text is part of neither the anchor nor the slug
    Given a reference carrying display text
    When the CLI builds the index
    Then the CLI records that display text as part of neither an anchor nor a slug
```

*Examples:* `` `[[container-spec-cli#Element Registry]]` `` records slug
`container-spec-cli` and anchor `Element Registry` on one reference
entry, at one span — not two references (AC-SPECANCH-001.1).
`#Element  Registry` is recorded with its double space intact, because an
index that normalizes on the way in cannot report what the author
actually wrote, which is the text a fix has to edit (AC-SPECANCH-001.2).
`` `[[container-spec-cli#Components#Element Registry]]` `` records the
path `["Components", "Element Registry"]`, so a check can require the
second to be nested beneath the first rather than merely present
(AC-SPECANCH-001.3). `` `[[adr-012#^decision-table]]` `` records a block
anchor with identifier `decision-table`; the two kinds resolve against
different data, so a reader of the index must be able to tell them apart
without re-parsing the anchor text (AC-SPECANCH-001.4).
`` `[[adr-012|the reference conventions ADR]]` `` records slug `adr-012`
and no anchor, and `` `[[container-spec-cli#Element Registry|the
registry]]` `` records the anchor `Element Registry` — without this the
display text, free prose containing spaces and often a `#`, would be
matched against headings and never resolve (AC-SPECANCH-001.5).

*Rationale:* AC-SPECANCH-001.2 and AC-SPECANCH-001.5 are the same rule
as AC-SPECREF-001.2 and AC-SPECREF-001.4 in the parent feature, applied
one level down: record what was written, and do not let a part of the
syntax that is *about* presentation leak into the part that is *about*
identity. Both were free to get wrong at the moment the corpus contains
no instance of either form, and expensive to change once it does.

### FEAT-SPECANCH-002: Anchor Resolution Checks

**Acceptance Criteria:**

```gherkin
@FEAT-SPECANCH-002
Feature: Anchor Resolution Checks

  As a document author
  I want an anchor that names no real heading reported
  So that a reference into a section survives the section being renamed, or tells me it did not

  The severity split follows the parent feature's: error where a reader
  reaches nothing, warning where a reader reaches something real but
  unintended, unverified where the tool cannot yet see enough to judge.

  Where this framework's notation and Obsidian's overlap, Obsidian's
  behaviour governs. A resolver that disagrees with the authoring
  environment converts the notation choice into a liability.

  @AC-SPECANCH-002.1
  Scenario: An anchor matching no heading is an error naming the document and the anchor
    Given a reference naming a heading anchor
    And the target document's slug resolves
    And the target's headings are available
    And no heading matches
    When the CLI checks the reference
    Then the CLI reports a finding of severity error naming the target document and the anchor

  @AC-SPECANCH-002.2
  Scenario: An anchor is matched on normalized heading text
    Then the CLI matches an anchor against a heading on their normalized heading texts

  @AC-SPECANCH-002.3
  Scenario: A nested anchor resolves only where the nesting holds
    Given an anchor naming a path
    When the CLI resolves it
    Then the CLI resolves it only where each heading after the first is nested beneath its predecessor

  @AC-SPECANCH-002.4
  Scenario: An anchor whose target headings are unavailable is unverified
    Given a reference naming an anchor
    And the target document's headings are not available
    When the CLI checks the reference
    Then the CLI reports that reference with severity unverified naming what is missing
    And the CLI does not report the anchor as resolved

  @AC-SPECANCH-002.5
  Scenario: An ambiguous anchor is a warning naming every match
    Given an anchor matching more than one heading in the target document
    When the CLI checks the reference
    Then the CLI reports a finding of severity warning naming every matching heading and its span

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
```

*Examples:* `reference/unresolved-anchor  error … "no heading named
'Componets' in container-spec-cli"` — the finding states the fact and
proposes no correction, per the parent feature's rule
(AC-SPECANCH-002.1). `#element registry` matches the heading `### Element
Registry`, because Obsidian resolves heading links case-insensitively
(AC-SPECANCH-002.2). `#Components#Element Registry` resolves against an
`### Element Registry` under a `## Components`, and does not resolve
against an `### Element Registry` under a different `## Container
Diagram` (AC-SPECANCH-002.3). Before [[frd-spec-structure-conformance]]
ships there is no heading data at all, so every anchor reference reports
`unverified` per AC-SPEC-006.3 and AC-SPEC-006.4; the same path covers a
target whose frontmatter failed to parse and whose headings were
therefore never extracted, which is what lets this feature ship before
the heading data exists (AC-SPECANCH-002.4).
`specs/modules/document-core/document-core-module.md` declares two `References`
headings today, so `` `[[module-document-core#References]]` `` would be
ambiguous; a reader lands on a real section and Obsidian takes the
first, but which one is an accident of document order that a later edit
can silently change (AC-SPECANCH-002.5). `` `[[#Terminology]]` `` in this
FRD resolves to this FRD's own Terminology heading; read as a reference
to a document whose slug is empty it would report as broken every time
(AC-SPECANCH-002.6). `` `![[container-spec-cli#Element Registry]]` ``
produces the same finding as the same reference without the leading `!`
(AC-SPECANCH-002.7). `` `[[adr-099#Decision]]` `` where no `adr-099`
exists produces exactly one finding — the parent feature's
unresolved-slug error — because two findings for one mistake trains a
reader to skim (AC-SPECANCH-002.8).

*Rationale:* The severity split follows the parent feature's: `error`
where a reader reaches nothing (AC-SPECANCH-002.1), `warning` where a
reader reaches something real but unintended (AC-SPECANCH-002.5),
`unverified` where the tool cannot yet see enough to judge
(AC-SPECANCH-002.4).

AC-SPECANCH-002.2 through AC-SPECANCH-002.7 all resolve the same
question the same way: where this framework's notation and Obsidian's
overlap, Obsidian's behavior governs. [[adr-012]] chose wikilinks
specifically because they survive the authoring environment, and a
resolver that disagrees with that environment converts the choice into a
liability — every disagreement is a finding an author cannot reproduce
in the tool they are looking at, or a broken link the tool shows as
fine.

That principle has one consequence worth stating plainly, because it
constrains this framework's own most likely use of anchors. Obsidian
matches a heading anchor against the *whole* heading text, so a
reference to a requirement in an FRD must name the heading in full —
`[[frd-spec-reference-integrity#FEAT-SPECREF-002: Reference Integrity
Checks]]` — and the terser `#FEAT-SPECREF-002` resolves to nothing. The
terse form could be made to work by matching an anchor against a
leading identifier token, but that is a resolver that only this tool
understands, and it breaks the moment an author clicks the link in
Obsidian. IF the terse form is wanted, THEN the durable fix is to change
what the heading says, not what the resolver accepts — which is a
template decision for [[frd-spec-structure-conformance]]'s required
headings, not a rule this feature can adopt on its own.

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

*Rationale:* AC-SPECANCH-003.2 is the criterion that makes anchors safe
to adopt at all. Without it, a heading rename is exactly the edit the
tooling cannot see: the author changes one document, every check against
that document passes, and the breakage lives entirely in files they did
not touch. The parent feature made the same argument for slugs and
resolved it the same way.

## References

- [[frd-spec-reference-integrity]] — the parent feature, which resolves
  the slug half of every reference and defers the anchor half here
- [[frd-spec-structure-conformance]] — supplies the headings and section
  spans this feature resolves an anchor against
- [[frd-spec-index]] — the index, finding engine, severities,
  and query surface both features extend
- [[adr-012]] — the wikilink and heading-anchor conventions, and the
  requirement that un-promoted components be addressed by anchor
- [[frd-spec-slug-rename]] — the sibling sub-feature, whose rewrites
  preserve the anchors this feature resolves
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
- `plugin/skills/spec/references/c4-guide.md` — element declarations and
  the roster headings an element anchor resolves against
