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
not at the whole blueprint. [[adr-012]] makes this the mandated form for
every un-promoted component — an element with no blueprint of its own is
referenced by a heading anchor into its parent's roster — so anchors are
not an optional flourish in this framework, they are how a whole class
of architectural element is addressed at all.

[[frd-spec-reference-integrity]] resolves the slug half of a reference
and deliberately stops there, splitting on the first `#` and ignoring
what follows (AC-SPECREF-002.2). This sub-feature specifies the other
half. It exists as its own document because the two halves have
different dependencies: slug resolution needs only the index, while
anchor resolution needs the heading data that
[[frd-spec-structure-conformance]] extracts.

The corpus has begun using the convention. A scan on 2026-09-03 found
551 body references, of which nine name an anchor: four citations of
`container-spec-cli#Command Surface` across three feature blueprints,
and five naming component rosters of [[module-document-core]]. Every one
of the nine resolves against a heading that exists today, and every one
names a single heading in full. They are the un-promoted-component form
[[adr-012]] mandates, written by hand and correct by luck rather than by
verification, because nothing checks them.

Nine is the number that makes this worth settling now. It is small
enough that any rule chosen here costs nothing to adopt, and large
enough that the convention is no longer hypothetical: the next author to
write an anchor will copy one of these, so whatever this document
decides is what gets copied.

The wider goal is that a reference behaves the same way in the two
environments this corpus is read in — Obsidian, which [[adr-012]] chose
the notation to survive, and a rendered markdown copy, which is how the
corpus travels into skills and repositories that do not hold it.
Matching an anchor the way a markdown renderer derives a heading
identifier is what keeps a reference this tool calls resolved clickable
in both.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Extracting headings and section spans.** Specified in
  [[frd-spec-structure-conformance]]. This feature consumes that data
  and does not re-derive it.
- **Resolving the slug half of a reference.** Specified in
  [[frd-spec-reference-integrity]]. A reference whose slug resolves to
  nothing is reported there; this feature says nothing further about it,
  so one broken reference never produces two findings.
- **Nested heading paths.** `#Components#Element Registry` is recorded
  as authored and matched as a single anchor, which names no heading and
  therefore reports. Resolving a path of nested headings is deferred.
- **Block anchors.** A caret carries no special meaning:
  `#^decision-table` is an ordinary anchor whose caret disappears in
  normalization. No feature extracts block identifiers, so there is
  nothing to resolve one against.
- **Naming a requirement by its code alone.** An anchor names a whole
  heading, so a requirement is cited as
  `[[frd-spec-reference-integrity#FEAT-SPECREF-002: Reference Integrity
  Checks]]`. Matching an anchor against a leading identifier token is
  deferred.
- **Emitting the fragment for an exported copy.** Specified in
  [[frd-spec-markdown-links]], which this feature requires to derive
  that fragment from the heading an anchor resolved to rather than from
  the anchor's own text.
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
- **Heading identifier** — the identifier a markdown renderer derives
  from a heading: its rendered text lowercased, with every character
  that is not a letter, digit, hyphen, underscore or space removed, and
  each remaining space replaced by a hyphen. This is the rule GitHub
  applies, and applying the same rule to an anchor is what makes a
  fragment in an exported copy land where the anchor did.
- **Ambiguous anchor** — an anchor whose identifier is the identifier of
  more than one heading in the target document.
- **Display alias** — Obsidian's `[[slug|label]]` form, which
  [[cross-document-references]] prohibits. A vertical bar is therefore
  not a separator in this notation.

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
    Then the CLI records the vertical bar and everything after it as part of the slug or the anchor it occurs in, dividing the reference at no vertical bar

  @AC-SPECANCH-001.6
  Scenario: An anchor in a frontmatter property is recorded like any other
    Given a reference-bearing frontmatter property whose value carries an anchor
    When the CLI builds the index
    Then the CLI records that anchor exactly as it records an anchor written in the body
```

*Examples:* `` `[[container-spec-cli#Element Registry]]` `` records slug
`container-spec-cli` and anchor `Element Registry` on one reference
entry, at one span — not two references (AC-SPECANCH-001.1).
`#Element  Registry` is recorded with its double space intact, because an
index that normalizes on the way in cannot report what the author
actually wrote, which is the text a fix has to edit (AC-SPECANCH-001.2).
`` `[[adr-012|the reference conventions ADR]]` `` records the slug
`adr-012|the reference conventions ADR`, which matches no corpus member
and reports as an unresolved reference — the display alias
[[cross-document-references]] prohibits resolves to nothing rather than
resolving quietly to something the standard forbids
(AC-SPECANCH-001.5). A document declaring
`sources: ["adr-012#Decision"]` records an anchor on that reference just
as a body citation of the same target would, because a citation names a
section wherever it is written (AC-SPECANCH-001.6).

*Rationale:* AC-SPECANCH-001.2 keeps the index a record of what the
author wrote. The parent feature already records a reference's text and
span before anything resolves it (AC-SPECREF-001.1); normalizing the
anchor on the way in would lose the one string a fix has to edit, and
every question about what an anchor matches is answered later, against
corpus knowledge the index does not itself hold.

AC-SPECANCH-001.5 is stated as a prohibition rather than left to
silence because the display alias is the one bracket form whose obvious
handling — split it off, resolve the rest — would quietly legalize a
notation [[cross-document-references]] forbids. Leaving it in the slug
makes the finding say what is wrong.

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
    Then the CLI checks the anchor of every reference carrying one, under a stable rule identifier for an anchor matching no heading and a second for an ambiguous anchor
```

*Examples:* `reference/unresolved-anchor  error … "no heading named
'Componets' in container-spec-cli"` — the finding states the fact and
proposes no correction, per the parent feature's rule
(AC-SPECANCH-002.1). `#element registry` matches the heading `### Element
Registry` because both yield the identifier `element-registry`, which is
also the fragment an exported copy of the citing document targets
(AC-SPECANCH-002.2). Before [[frd-spec-structure-conformance]] ships
there is no heading data at all, so every anchor reference reports
`unverified` per AC-SPEC-006.3 and AC-SPEC-006.4; the same path covers a
target whose frontmatter failed to parse and whose headings were
therefore never extracted, which is what lets this feature ship before
the heading data exists (AC-SPECANCH-002.4).
`specs/architecture/module-document-core.md` declares two `References`
headings today, so `` `[[module-document-core#References]]` `` names two
sections and resolves to neither: the author disambiguates by renaming a
heading, which is a durable fix, rather than by counting sections
(AC-SPECANCH-002.5). `` `[[#Terminology]]` `` in this FRD resolves to
this FRD's own Terminology heading; read as a reference to a document
whose slug is empty it would report as broken every time
(AC-SPECANCH-002.6). `` `![[container-spec-cli#Element Registry]]` ``
produces the same finding as the same reference without the leading `!`
(AC-SPECANCH-002.7). `` `[[adr-099#Decision]]` `` where no `adr-099`
exists produces exactly one finding — the parent feature's
unresolved-slug error — because two findings for one mistake trains a
reader to skim (AC-SPECANCH-002.8).

*Rationale:* AC-SPECANCH-002.2 settles normalization by adopting a rule
that already exists rather than inventing one. A heading identifier is
what a rendered copy of the corpus uses as a fragment, so matching on it
makes one guarantee out of two that would otherwise have to be kept in
step by hand: an anchor this feature resolves is an anchor
[[frd-spec-markdown-links]] can target. Case-insensitivity falls out of
the same rule rather than being asserted separately.

AC-SPECANCH-002.5 makes ambiguity an error, and deliberately offers no
notation for selecting one of several identical headings. A renderer
disambiguates by position — the second `References` becomes
`references-1` — and a reference written that way says nothing about
which section was meant, breaks when a heading is inserted above it, and
obliges an author to count sections to write it. Reporting the
ambiguity instead puts the fix where it lasts, in the heading.

AC-SPECANCH-002.9 exists because a check nobody runs is not a check. The
anchor rules are part of an ordinary check run over the whole corpus,
not an option, and each reports under its own rule identifier so a
caller can select it.

That principle has one consequence worth stating plainly, because it
constrains this framework's own most likely use of anchors. An anchor
names a whole heading, so a reference to a requirement in an FRD must
name that heading in full —
`[[frd-spec-reference-integrity#FEAT-SPECREF-002: Reference Integrity
Checks]]` — and the terser `#FEAT-SPECREF-002` resolves to nothing. The
terse form could be made to work by matching an anchor against a leading
identifier token, but that is a resolver only this tool understands, and
it breaks the moment an author clicks the link in Obsidian. IF the terse
form is wanted, THEN the durable fix is to change what the heading says,
not what the resolver accepts — which is a template decision for
[[frd-spec-structure-conformance]]'s required headings, not a rule this
feature can adopt on its own.

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
An ambiguous reference resolves to no heading, and an author renaming
one of the headings it matches is exactly the reader who has to see it,
so the query lists it rather than hiding it behind its own error
(AC-SPECANCH-003.4).

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
