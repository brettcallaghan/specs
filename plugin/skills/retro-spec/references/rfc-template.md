---
title: "RFC: <Title Spanning the Proposal Set>"
description: <One-sentence summary of what the proposals improve>
slug: rfc-<nnnn>-<short-description>  # numbered by the repo that
                                      # resolves the RFC — see
                                      # Numbering and location below
type: rfc
status: draft
updated_at: <YYYY-MM-DDTHH:MM:SSZ>
audience: project | skill-authors     # who resolves the proposals
origin: <project-name>                # the project whose session
                                      # produced the evidence
plugin_version: "<x.y.z>"             # spec-driven-dev plugin version
                                      # in use during that session
related: []                           # slugs of documents the proposals
                                      # would change (audience: project)
---

<!--
Template: Improvement RFC
Version: 1.1
Source: spec-driven-dev framework (produced by the retro-spec skill)

Purpose
  Proposes evidence-backed improvements to the documents, templates,
  or skills that guided a completed phase of spec-driven work. An RFC
  is a process artifact, not a specification document: it argues from
  observed friction in one real session toward changes that prevent
  the same friction in the next one.

Scope — one RFC per retro per audience
  An RFC bundles ALL qualifying proposals from one retro for one
  audience; do NOT split proposals into separate RFC files. The
  proposals share a single Evidence base and Provenance block — that
  sharing is the point. The RFC is the envelope; the PROPOSAL is the
  unit of decision, and review resolves each proposal individually
  (adopt / defer / reject) in the Resolution section.

When to use
  - A retro (retro-spec skill) after a downstream phase has exercised
    upstream documents found improvements that clear the bar below
  - Any session that surfaced concrete, citable defects in the
    project's specs (audience: project) or in the framework's skills,
    templates, or standards (audience: skill-authors)

When NOT to use
  - No concrete session evidence — an idea that "seems better" without
    an observed failure is a discussion, not an RFC
  - The fix is to one document's CONTENT (a wrong requirement, a stale
    diagram) — just fix the document; RFCs are for the machinery that
    produces documents, or for structural gaps across them
  - Recording a design decision — that is an ADR

Audience semantics
  - `project` — proposals change the origin project's own documents
    (its specs, blueprints, standards). Resolved in the origin
    repo. File: the origin repo's rfcs/ directory.
  - `skill-authors` — proposals change the spec-driven-dev framework
    (skills, templates, standards, guides). Resolved by the
    framework maintainers. File: the origin repo's rfcs/outbox/
    directory, from which the human copies it into the framework
    repo's rfcs/inbox/ for triage.

Numbering and location
  - An RFC is numbered by the repo that resolves it, when it is
    resolved — numbering marks ownership of the outcome.
  - audience: project — name it rfc-<nnnn>-<short-description>.md
    using the origin repo's next free number, in rfcs/.
  - audience: skill-authors — name it
    rfc-<origin-project>-<short-description>.md (NO number) in
    rfcs/outbox/. The framework repo's triage assigns its own number
    when the RFC is ingested and moves it into that repo's rfcs/.

Status lifecycle (controlled vocabulary per
template-authoring-standards)
  draft            under authorship by the retro agent
  ready-for-review submitted; awaiting triage / human review
  approved         triage accepted one or more proposals; work agreed
  in-progress      accepted proposals are being implemented
  retired          resolved: every proposal adopted, deferred
                   (to a tracker), or rejected — the Resolution
                   section is the record; deferred work lives on in
                   the tracker, not in the RFC
  An RFC whose every proposal is rejected goes straight from
  ready-for-review to retired, with the rationale in Resolution.

Readiness gate — verify EVERY item before setting ready-for-review;
an RFC failing any item stays draft:
  [ ] Every proposal cites at least one Evidence finding from the
      session (by finding number)
  [ ] Every Problem statement quotes the misleading source text
      verbatim with its location — or states explicitly that NO text
      exists where the needed rule should live
  [ ] Every Change is concrete enough to implement without further
      design work (proposed wording, rule, or mechanism — not "should
      be clearer")
  [ ] Proposals are independently adoptable, or Adoption states which
      must land together and why
  [ ] Backward compatibility of each proposal is stated (existing
      documents stay valid, or the migration is described)
  [ ] Provenance is complete, including the version of every skill,
      template, and standard cited — triage cannot judge "is this
      already fixed?" without them
  [ ] Findings that did NOT clear the bar appear under Considered and
      Not Proposed with the reason — never silently dropped
  [ ] The Resolution section is present and empty (the resolving
      review fills it)

General instructions
  1. Complete each section following the guidance in comments, then
     delete all HTML comment blocks.
  2. Keep lines within 80 characters per
     [[template-authoring-standards]].
  3. Reference framework documents with [[slug]]; quote file text with
     repo-relative paths in backticks.
-->

# RFC: <Title Spanning the Proposal Set>

## Summary

<!--
2-4 sentences: what phase was executed, what class of friction was
observed, and the one-line thesis of the proposals ("the authoring
skill does not encode its consumer's acceptance criteria", "the
template asks for X but its consumer needs Y"). A reader should know
from the Summary alone whether this RFC concerns them.
-->

<What was executed, what went wrong, and what the proposals fix>

## Provenance

<!--
The facts triage needs to reproduce your context and to check whether
a proposal is already fixed. Versions are load-bearing: cite the
version of every skill, template, and standard the session actually
followed (skill versions are in SKILL.md frontmatter; template and
standard versions are in their instruction headers or frontmatter).
-->

| Field | Value |
|-------|-------|
| Origin project | <project name / repo> |
| Phase executed | <e.g. groomed feature-x blueprint into beads> |
| Documents exercised | <paths/slugs of the docs the phase consumed> |
| Documents produced | <paths/slugs of what the phase created> |
| Skill(s) followed | <name vX.Y, name vX.Y> |
| Templates / standards read | <name vX.Y, name vX.Y> |
| Plugin version | <x.y.z> |
| Session date | <YYYY-MM-DD> |

## Evidence

<!--
The shared factual base every proposal cites into. One numbered
finding per observed defect or friction point, stated as what
HAPPENED, not what should change:

  1. <What the agent did or produced> — <the defect> — <root cause:
     the source text (quoted, located) that misled, or the gap where
     no guidance existed>.

Findings come from the session itself: human corrections during the
review loop, gate bounces, guesses an implementer would have had to
make, workarounds, abandoned syntax, anything re-done. Distinguish
carefully between defects CAUSED by the guidance (in scope) and
authoring carelessness the guidance already warns against (out of
scope — note it under Considered and Not Proposed if worth recording).
-->

1. <Finding>
2. <Finding>

## Proposals

<!--
One ### subsection per proposal, ordered by expected impact, IDs
P1..Pn. Each proposal follows the shape below. Keep proposals
independently adoptable; if two only make sense together, say so here
and in Adoption.
-->

### P1 — <Imperative title of the change>

**Problem.** <Why the current state produces the defect. Quote the
misleading source text verbatim with its location ("`SKILL.md`
Phase 2: 'One item per element delta …'"), or state that no text
exists where the rule should live.>

**Change.** <The concrete change: proposed wording for the template
or skill, the new rule in EARS where normative, the mechanism. Ready
to apply, not a direction to explore.>

**Evidence.** <Finding number(s) this traces to, plus the expected
effect on the next run ("the plan would have had three parallelizable
items instead of one oversized one").>

## Considered and Not Proposed

<!--
Ideas examined and deliberately NOT proposed, each with the reason
(insufficient evidence, cost exceeds benefit, conflicts with a design
principle, one-off carelessness rather than a guidance defect). This
section is the record that keeps future retros from re-deriving the
same rejected ideas. Write "None." if empty.
-->

- <Idea> — <why not>

## Non-Goals

<!--
What these proposals deliberately do not change, to bound triage
scope (e.g. "no new document types", "consumer checks stay as strict
as they are — the fix is visibility, not leniency"). Write "None." if
empty.
-->

- <Non-goal>

## Adoption

<!--
Grouping and ordering: which proposals are independent, which must
land together, expected touch points (files), and the backward
compatibility statement for each proposal not already covered above.
End with one sentence on the expected overall effect.
-->

<Ordering, grouping, compatibility, expected effect>

## Resolution

<!--
FILLED BY THE RESOLVING REVIEW (framework triage or the origin
project's own review) — leave empty; keep the section heading. The
review records, per proposal: adopted (as proposed / with refinement
— describe the refinement), deferred (tracker issue id + theme), or
rejected (rationale). Plus the date, the files changed, and any
version bumps applied.
-->
