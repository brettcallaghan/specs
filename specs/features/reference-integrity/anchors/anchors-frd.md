---
title: "FRD: Spec Corpus Anchor References"
description: Resolves the anchor half of a wikilink — heading, nested
  heading path, and block reference — so a link into a section of a
  document is checked as rigorously as a link to the document
slug: frd-spec-anchor-references
type: frd
status: draft
updated_at: 2026-08-05T00:00:00Z
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

### FEAT-SPECANCH-001: Anchor Extraction in the Index

**User Story:** As an architecture owner, I want the anchor half of
every reference recorded in the index, so that a query can tell which
section of a document a reference points into, not merely which
document.

**Acceptance Criteria:**

- **AC-SPECANCH-001.1:** WHEN the CLI builds the index and a reference
  carries an anchor, the CLI SHALL record the anchor alongside that
  reference's slug and span.
  - *Example:* `[[container-spec-cli#Element Registry]]` records slug
    `container-spec-cli` and anchor `Element Registry` on one reference
    entry, at one span — not two references.
- **AC-SPECANCH-001.2:** The CLI SHALL record an anchor as it was
  authored, without normalization.
  - *Example:* `#Element  Registry` is recorded with its double space
    intact. Normalization is a matching rule (AC-SPECANCH-002.2), and an
    index that normalizes on the way in cannot report what the author
    actually wrote, which is the text a fix has to edit.
- **AC-SPECANCH-001.3:** WHEN an anchor names a path of nested headings,
  the CLI SHALL record the path as an ordered sequence.
  - *Example:* `[[container-spec-cli#Components#Element Registry]]`
    records the path `["Components", "Element Registry"]`, so a check
    can require the second to be nested beneath the first rather than
    merely present.
- **AC-SPECANCH-001.4:** WHEN an anchor begins with `^`, the CLI SHALL
  record it as a block anchor, distinguishable in the index from a
  heading anchor.
  - *Example:* `[[adr-012#^decision-table]]` records a block anchor with
    identifier `decision-table`. The two kinds resolve against different
    data, so a reader of the index must be able to tell them apart
    without re-parsing the anchor text.
- **AC-SPECANCH-001.5:** The CLI SHALL NOT record display text as part
  of an anchor or a slug.
  - *Example:* `[[adr-012|the reference conventions ADR]]` records slug
    `adr-012` and no anchor; `[[container-spec-cli#Element Registry|the
    registry]]` records the anchor `Element Registry`. Without this the
    display text — free prose, containing spaces and often a `#` — would
    be matched against headings and never resolve.

*Rationale:* AC-SPECANCH-001.2 and AC-SPECANCH-001.5 are the same rule
as AC-SPECREF-001.2 and AC-SPECREF-001.4 in the parent feature, applied
one level down: record what was written, and do not let a part of the
syntax that is *about* presentation leak into the part that is *about*
identity. Both were free to get wrong at the moment the corpus contains
no instance of either form, and expensive to change once it does.

### FEAT-SPECANCH-002: Anchor Resolution Checks

**User Story:** As a document author, I want an anchor that names no
real heading reported, so that a reference into a section survives the
section being renamed — or tells me it did not.

**Acceptance Criteria:**

- **AC-SPECANCH-002.1:** WHEN a reference names a heading anchor, the
  target document's slug resolves, and the target's headings are
  available, the CLI SHALL report a finding of severity `error` naming
  the target document and the anchor IF no heading matches.
  - *Example:* `reference/unresolved-anchor  error … "no heading named
    'Componets' in container-spec-cli"`. The finding states the fact and
    proposes no correction, per the parent feature's rule — an agent
    resolves a plain "not found" faster than a guess it must verify.
- **AC-SPECANCH-002.2:** The CLI SHALL match an anchor against a heading
  on their normalized heading texts.
  - *Example:* `#element registry` matches the heading `### Element
    Registry`. Obsidian resolves heading links case-insensitively, and a
    resolver stricter than the authoring tool would report findings that
    the tool the author is looking at shows as working links.
- **AC-SPECANCH-002.3:** WHEN an anchor names a path, the CLI SHALL
  resolve it only IF each heading after the first is nested beneath its
  predecessor.
  - *Example:* `#Components#Element Registry` resolves against a `###
    Element Registry` under a `## Components`, and does not resolve
    against an `### Element Registry` under a different `## Container
    Diagram` — the path is the author's assertion about where the
    section lives, and a resolver that ignores it silently accepts a
    reference to the wrong section.
- **AC-SPECANCH-002.4:** IF a reference names an anchor and the target
  document's headings are not available, THEN the CLI SHALL report that
  reference with severity `unverified` naming what is missing, and SHALL
  NOT report the anchor as resolved.
  - *Example:* Before [[frd-spec-structure-conformance]] ships there is
    no heading data at all, so every anchor reference reports
    `unverified` per AC-SPEC-006.3 and AC-SPEC-006.4. The same path
    covers a target whose frontmatter failed to parse and whose headings
    were therefore never extracted. This is the first use in the tool of
    a severity AC-SPEC-006.3 defines but nothing has yet emitted, which
    is what lets this feature ship before the heading data exists.
- **AC-SPECANCH-002.5:** WHEN an anchor matches more than one heading in
  the target document, the CLI SHALL report a finding of severity
  `warning` naming every matching heading and its span.
  - *Example:* `specs/architecture/module-document-core.md` declares two
    `References` headings today, so `[[module-document-core#References]]`
    would be ambiguous. It is a `warning` rather than an `error` because
    the reference does resolve — a reader lands on a real section, and
    Obsidian takes the first — but which one they land on is an accident
    of document order that a later edit can silently change.
- **AC-SPECANCH-002.6:** WHEN a reference names an anchor and no slug —
  the form `[[#Heading]]` — the CLI SHALL resolve it against the
  containing document.
  - *Example:* `[[#Terminology]]` in this FRD resolves to this FRD's own
    Terminology heading. Obsidian authors write this form constantly;
    read as a reference to a document whose slug is empty it would
    report as broken every time.
- **AC-SPECANCH-002.7:** The CLI SHALL resolve and check an embed
  (`![[…]]`) exactly as it resolves and checks the equivalent link.
  - *Example:* `![[container-spec-cli#Element Registry]]` produces the
    same finding as the same reference without the leading `!`. An embed
    is a link that renders its target; the target either exists or does
    not.
- **AC-SPECANCH-002.8:** The CLI SHALL NOT report an anchor finding for
  a reference whose slug resolves to no corpus member.
  - *Example:* `[[adr-099#Decision]]` where no `adr-099` exists produces
    exactly one finding — the parent feature's unresolved-slug error —
    and not a second one about the anchor. Two findings for one mistake
    trains a reader to skim.

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

**User Story:** As an authoring agent, I want to ask what references
point into a section before I rename or move it, so that I learn the
blast radius from the tool rather than from a reader who followed a
broken link.

**Acceptance Criteria:**

- **AC-SPECANCH-003.1:** WHEN asked for the references to a document
  qualified by an anchor, the CLI SHALL list every reference whose
  anchor resolves to that heading, each with its referring document and
  span.
  - *Example:* `spec refs 'container-spec-cli#Element Registry'` lists
    the documents that would break if that heading were renamed. This is
    the question [[adr-012]] created by making anchors the address of
    every un-promoted component: renaming a roster heading is a
    corpus-wide edit, and nothing today can tell an author how wide.
- **AC-SPECANCH-003.2:** WHEN the CLI checks a single document, the CLI
  SHALL report inbound anchor references to that document that no longer
  resolve, each cited at the span in the referring document.
  - *Example:* An agent renames a `### Element Registry` heading to `###
    Registry` and the write-time check on that one file reports the four
    documents whose anchors now point at nothing, at their own lines.
    This is AC-SPECREF-002.9 extended to the anchor half: the edit most
    likely to break an anchor is made in the document the anchor points
    *into*, which a check scoped to the edited file cannot otherwise
    see.
- **AC-SPECANCH-003.3:** WHEN no reference names the given anchor, the
  CLI SHALL report that and exit zero.
  - *Example:* Consistent with AC-SPECREF-003.5 and AC-SPEC-007.2 — an
    empty answer is an answer, not a failure.

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
- [[adr-012]] — the wikilink and heading-anchor conventions, and the
  requirement that un-promoted components be addressed by anchor
- [[frd-spec-slug-rename]] — the sibling sub-feature, whose rewrites
  preserve the anchors this feature resolves
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
- `plugin/skills/spec/references/c4-guide.md` — element declarations and
  the roster headings an element anchor resolves against
