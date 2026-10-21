---
title: "Spec Corpus Anchor References"
description: Resolves the anchor half of a wikilink — the heading it
  names — so a link into a section of a document is checked as
  rigorously as a link to the document
slug: frd-spec-anchor-references
type: frd
status: approved
sources:
  - "frd-spec-reference-integrity"
---

# Spec corpus anchor references

## Overview

A wikilink can name a place inside a document as well as the document
itself: `[[container-spec-cli#Element Registry]]` points at one heading,
not at the whole Container Blueprint. [[adr-012]] mandates that form for
every un-promoted component. An element with no blueprint of its own is
referenced by a heading anchor into its parent's roster, so anchors are
how a whole class of architectural element is addressed.

[[frd-spec-reference-integrity]] resolves the slug half of a reference
and stops there: it splits on the first `#` and ignores what follows
(AC-SPECREF-002.2). This sub-feature specifies the other half. It's a
separate document because the two halves need different data: slug
resolution needs only the index, and anchor resolution needs the headings
[[frd-spec-structure]] records.

References in the corpus name anchors, for example into
`container-spec-cli#Command Surface` and into component rosters of
[[module-document-core]]. Each names one heading in full, and each
reaches a heading that exists.

A reference should behave the same way in both environments the corpus
is read in: Obsidian, which [[adr-012]] chose the notation to survive,
and a rendered Markdown copy, which is how the corpus travels into skills
and repositories that don't hold it. Deriving a heading identifier the
way a renderer does keeps a resolved reference clickable in both.

**Out of scope for this feature.** Each of these is deferred on purpose,
not overlooked:

- **Extracting headings and section spans.** [[frd-spec-structure]]
  specifies it. This feature consumes that data and doesn't re-derive
  it.
- **Resolving the slug half of a reference.**
  [[frd-spec-reference-integrity]] specifies it and reports a reference
  whose slug resolves to nothing. This feature says nothing further
  about that reference, so one broken reference never produces two
  findings.
- **Block anchors.** A caret carries no special meaning: the caret in
  `#^decision-table` disappears in normalization and leaves an ordinary
  anchor. No feature extracts block identifiers to resolve one against.
- **Naming a requirement by its code alone.** An anchor names a whole
  heading, so a requirement is cited as
  `[[frd-spec-reference-integrity#FEAT-SPECREF-002: Reference Integrity
  Checks]]`. Matching against a leading identifier token is deferred.
- **Emitting the fragment for an exported copy.**
  [[frd-spec-markdown-links]] specifies it, and this feature requires
  that feature to derive the fragment from the heading an anchor resolved
  to.
- **Transclusion.** An embed (`![[slug#Heading]]`) is resolved and
  checked exactly as the equivalent link, per AC-SPECANCH-002.7. Pulling
  the target's content into the referring document is rendering behavior
  this tool doesn't have and doesn't need.
- **Rewriting an anchor when its target heading is renamed.** The tool
  reports, and a reviewing agent fixes, consistent with the parent
  feature's no-suggestions rule.
- **Anchors into documents outside the specification set.** The parent
  feature draws the same boundary.

## Terminology

- **Anchor** — the portion of a reference after the first `#`, naming a
  place inside the target document rather than the document as a whole.
- **Heading identifier** — GitHub's rule for deriving a fragment from a
  heading: lowercase its rendered text, remove every character that isn't
  a letter, digit, hyphen, underscore, or space, and replace each
  remaining space with a hyphen. Deriving one from the anchor as well
  makes an exported copy's fragment land where the anchor did.
- **Ambiguous anchor** — an anchor whose identifier is the identifier of
  more than one heading in the target document.
- **Display alias** — Obsidian's `[[slug|label]]` form, which
  [[cross-document-references]] prohibits. No slug admits a vertical
  bar, so an aliased reference names no document. Where an aliased
  reference also carries an anchor, the bar falls inside the anchor.

## Requirements

A wikilink written anywhere in a Gherkin source is recorded as a
reference ([[adr-021]] constraint 14). So this feature writes its
notation illustrations in backticked Markdown beside each block, not in a
step, where a literal wikilink would enter the index as a citation this
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
  Scenario: A vertical bar does not divide an anchor
    Given a reference whose anchor contains a vertical bar
    When the CLI builds the index
    Then the CLI records the vertical bar as ordinary text of that anchor

  @AC-SPECANCH-001.6
  Scenario: An anchor in a frontmatter property is recorded like any other
    Given a reference-bearing frontmatter property whose value carries an anchor
    When the CLI builds the index
    Then the CLI records that anchor exactly as it records an anchor written in the body
```

*Examples:* `` `[[container-spec-cli#Element Registry]]` `` records slug
`container-spec-cli` and anchor `Element Registry` on one reference
entry, at one span, not as two references (AC-SPECANCH-001.1).
`#Element  Registry` keeps its double space because the index has to
report the text a fix will edit (AC-SPECANCH-001.2).
`` `[[container-spec-cli#Element Registry|the registry]]` `` records the
anchor `Element Registry|the registry`, which is the identifier of no
heading and reports (AC-SPECANCH-001.5). `sources: ["adr-012#Decision"]`
records an anchor exactly as a body citation of the same target does: a
citation names a section wherever it's written (AC-SPECANCH-001.6).

*Rationale:* AC-SPECANCH-001.2 keeps the index a record of what the
author wrote: the parent feature records a reference's text and span
before anything resolves it (AC-SPECREF-001.1), and every question about
what an anchor matches is answered later, against corpus knowledge the
index doesn't hold. AC-SPECANCH-001.5 covers the one half an alias can
reach. A vertical bar isn't a slug character, so `` `[[adr-012|label]]` ``
names no document whatever this feature says. An anchor, though, is free
text, and splitting one on a bar would quietly legalize a notation
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
    Then the CLI reports a finding of severity error naming every matching heading with its span and the heading path that selects it alone
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

  @AC-SPECANCH-002.10
  Scenario: A nested anchor resolves through its enclosing headings
    Given a reference whose anchor holds more than one segment separated by "#"
    When the CLI checks the reference
    Then the CLI resolves it to the heading whose heading path is those segments, where one heading's is
    And otherwise the CLI resolves it to the heading named by the last segment whose enclosing headings include, in order, a heading named by each earlier segment
    And the CLI reports an unresolved anchor where no heading satisfies every segment, and an ambiguous anchor where more than one heading has that heading path or, none having it, more than one satisfies every segment
```

*Examples:* `reference/unresolved-anchor  error … "no heading named
'Componets' in container-spec-cli"` states the fact and proposes no
correction, per the parent feature's rule (AC-SPECANCH-002.1).
`#element registry` matches `### Element Registry` because both yield
`element-registry`, which is also the fragment an exported copy targets
(AC-SPECANCH-002.2). Headings are unavailable where the index carries no
entry for the target path at all. A standalone `.feature` source is the
concrete case: its references are recorded against its own path, though
the file isn't a corpus member. So a `` `[[#Heading]]` `` written there
reports `unverified` per AC-SPEC-006.3 and AC-SPEC-006.4, not resolved or
broken. The tool says it couldn't look, which isn't the same claim as
looking and finding nothing (AC-SPECANCH-002.4). An entry that exists and
records no matching heading isn't this case: the tool looked, so it
reports an error. [[module-document-core]] declares two `References`
headings, so `` `[[module-document-core#References]]` `` names two
sections and resolves to neither. The author disambiguates by renaming
one of them (AC-SPECANCH-002.5). `` `[[#Terminology]]` `` in this FRD
resolves to this FRD's own Terminology heading. Read as a reference to a
document whose slug is empty, it would report as broken every time
(AC-SPECANCH-002.6). `` `![[container-spec-cli#Element Registry]]` ``
produces the same finding as the same reference without the leading `!`
(AC-SPECANCH-002.7). `` `[[adr-099#Decision]]` `` where no `adr-099`
exists produces one finding, the parent's unresolved-slug error, because
two findings for one mistake train a reader to skim (AC-SPECANCH-002.8).
`` `[[module-document-core#document-core#References]]` `` resolves to the
top-level `References` alone. `document-core#components#references` also
satisfies every segment, but a heading whose full heading path is the
anchor's segments takes precedence (AC-SPECANCH-002.10). The precedence
needs more than one segment: the one-segment
`` `[[module-document-core#References]]` `` stays ambiguous, as
AC-SPECANCH-002.5 states.

*Rationale:* AC-SPECANCH-002.2 adopts a normalization that already
exists instead of inventing one. A rendered copy uses a heading
identifier as a fragment, so matching on it collapses two guarantees into
one: an anchor this feature resolves is an anchor
[[frd-spec-markdown-links]] can target. Case-insensitivity follows from
the same rule.

AC-SPECANCH-002.5 offers no notation for selecting one of several
identical headings. A renderer disambiguates by position (the second
`References` becomes `references-1`). A citation written that way says
nothing about which section was meant, breaks when a heading is inserted
above it, and makes the author count sections. Reporting the ambiguity
puts the fix where it lasts: in the heading.

AC-SPECANCH-002.9 makes the anchor rules part of an ordinary corpus
check, not an option, and puts each under its own rule identifier so a
caller can select it.

One consequence constrains this framework's most likely use of anchors:
a reference to a requirement names its heading in full, and the terser
`#FEAT-SPECREF-002` resolves to nothing. If the terse form is wanted, the
fix is to change what the heading says (a template decision for
[[frd-spec-structure-conformance]]'s required headings). The fix isn't a
resolver only this tool understands, which would fail the moment an
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

  @AC-SPECANCH-003.2 @negative
  Scenario: A single-document check reports no anchor finding located in another document beyond those its change broke
    When the CLI checks a single document
    Then the CLI reports no anchor finding located in a document other than the one named, except the unresolved-anchor findings AC-SPECREF-002.11 names
    And that holds for an ambiguous anchor held by another document, and for any anchor held by another document into a different document

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

  @AC-SPECANCH-003.5 @negative
  Scenario: An outbound query takes no anchor
    Given a query for the references out of a document
    When the argument is qualified by an anchor
    Then the CLI refuses the combination and exits with the invocation-error status

  @AC-SPECANCH-003.6
  Scenario: An anchor naming no heading is distinguishable from a heading nobody cites
    When the CLI is asked for the references to a document qualified by an anchor
    Then the CLI states in the answer whether the anchor names a heading of that document
    And where it names none, the CLI answers with no references and exits zero
```

*Examples:* `specs refs 'container-spec-cli#Element Registry'` lists the
documents that would break if that heading were renamed. That's the
question [[adr-012]] raises by making anchors the address of every
un-promoted component (AC-SPECANCH-003.1). An agent renames a
`### Element Registry` heading to `### Registry`. The query asked before
the rename names the four documents whose anchors would point at nothing
(AC-SPECANCH-003.2). After the rename, the write-time check on that one
file reports them, located in the documents holding them
(AC-SPECREF-002.11). An empty answer is an answer, not a failure,
consistent with AC-SPECREF-003.5 and AC-SPEC-007.2 (AC-SPECANCH-003.3).
An ambiguous reference resolves to no heading, and the author renaming
one of the headings it matches is exactly the reader who has to see it.
So the query lists it instead of hiding it behind its own error
(AC-SPECANCH-003.4). An anchor names a heading of the cited document, and
the outbound direction asks what the given document cites. That's why
`specs refs 'adr-012#Decision' --outbound` names no set and is refused
with status `2`, the invocation error of [[cli-exit-status-standards]]
(AC-SPECANCH-003.5). `specs refs 'adr-012#Decison'` answers with no
references and states that the anchor isn't declared (as
`anchor_declared: false`, for instance). An uncited heading of `adr-012`
answers with no references and states that it is declared. Both exit
`0`, because a subject the corpus doesn't hold is an answer
(AC-SPECANCH-003.6).

*Rationale:* AC-SPECANCH-003.1 is what makes anchors safe to adopt at
all. A heading rename is the edit a single-document check can't see: the
breakage lives entirely in files the author didn't touch. The write-time
check on the renamed document reports exactly that breakage (references
into it that no longer resolve). AC-SPECANCH-003.2 keeps every other
anchor finding elsewhere out of that check, as the parent feature does
for slugs, so a write-time hook never fails on breakage the author didn't
cause. The query asked before the rename surfaces the blast radius in
advance.

AC-SPECANCH-003.6 exists because, without it, a misspelled anchor and an
uncited heading return the same empty answer, and an agent checking the
blast radius of a rename reads a typo as "safe to rename."

## References

- [[frd-spec-reference-integrity]] — the parent feature, which resolves
  the slug half of every reference and defers the anchor half here
- [[frd-spec-structure]] — supplies the headings and section spans this
  feature resolves an anchor against
- [[frd-spec-index]] — the index, finding engine, severities,
  and query surface both features extend
- [[frd-spec-markdown-links]] — rewrites a citation into a Markdown
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
- [[c4-standards]] — element declarations and the roster headings an
  element anchor resolves against
