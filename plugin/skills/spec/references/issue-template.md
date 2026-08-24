---
title: "<the change, not the symptom>"
description: <One sentence — what changes, and why it is worth changing>
slug: issue-<identifier>
                                     # No `type` property: see "Non-corpus"
                                     # below — do not add one.
audience: skill                      # project | skill — from routing, never
                                     # judged here. See "Which audience".
target: <skill or document name>     # what the fix lands in
priority: 2                          # 0 (critical) .. 4 (backlog), PROPOSED
updated_at: <YYYY-MM-DDTHH:MM:SSZ>

# Evidence — what this issue rests on
findings: []                         # finding ids, e.g. ["92u#2", "1ilc.7#1"]
recurrence: 0                        # prior runs in the window carrying an
                                     # UNFILED finding of this kind and target
report: <path>                       # the run report carrying those findings

# Provenance — who and what produced this
run: <run id>                        # the sprint run epic; absent for a solo retro
session: <session id>                # OPTIONAL — the authoring agent's session,
                                     # where the harness exposes one. This is the
                                     # IDENTITY field: it answers "was this the
                                     # same agent", which `model` cannot, since
                                     # two unrelated agents can share a model.
branch: <branch the run worked on>
skill_version: <what was being exercised>
model: <model that produced the findings>   # load-bearing; never omit
---

<!--
Template: Issue
Version: 1.0
Source: This specification framework. Reference conventions follow
  [[template-authoring-standards]]. The findings this issue cites are
  captured per [[findings-capture-protocol]].

Sections (complete each, then delete its instruction block)
  - What is wrong, and what it costs
  - Evidence
  - Proposed change

Purpose
  ONE issue is ONE actionable change, and the issue is the UNIT OF DECISION.
  That is why it is not bundled with others into a multi-proposal document:
  a container holding several independently-decidable things forces triage
  to decide inside it, and forces its own bookkeeping to count in a
  different unit than it files in.

  An issue is produced by a retro (from that run's own findings) or arrives
  from a project reporting one. Both use this template unchanged.

Which audience — apply the merge test
  Findings and the issues built from them serve two audiences, and the
  difference is not visible in the words. "The blueprint template has no
  rule for X" and "our blueprint has no rule for X" are the same sentence
  about different artifacts.

    Where would the fix land?
      - In this project's own repository  -> audience: project
      - In the repository that vends the skill -> audience: skill

  `audience: project` renders straight into the project's own tracker.
  `audience: skill` is written to `.sf/outbox/<target>/` for delivery to
  the repository that vends it, and is NEVER filed in the reporting
  project's own backlog — a project must not accumulate issues it cannot
  act on. Same fields, two encodings; it is one template.

Non-corpus: no `type` frontmatter property
  Like a work order, an issue materializes into a tracker record and stops
  existing as a document, so corpus membership does not apply. It carries no
  DocumentType and `scripts/check-template-drift.py` governs it as a
  template rather than as a corpus member.

On `recurrence`
  The count of prior runs in the recurrence window carrying an UNFILED
  finding of the same kind and target. It starts at 0 for a kind/target
  pair nothing has reported before, and that is not a defect — it is the
  difference between a real recurring problem and one bad afternoon, and it
  only becomes informative from the second run that reports the pair.

On `model`
  Required, and load-bearing rather than ceremonial: findings differ
  systematically by the model that produced them, so an issue that hides
  which model hit the friction cannot be weighed against one from a
  different capability tier. Where a harness withholds it, write
  `<withheld by session policy>` — never omit the key, because an omitted
  field is indistinguishable from a forgotten one.
-->

## What is wrong, and what it costs

<!--
One paragraph. Two obligations:

  - State the DEFECT plainly — what is wrong with the thing named in
    `target`, not what the symptom felt like.
  - State the COST explicitly. An issue that says what is wrong but not what
    it costs cannot be prioritised against anything, and a queue of them
    grows until nobody drains it.

Cost is concrete: work redone, a wrong result shipped, an agent blocked, a
decision made silently that a human would have made differently.
-->

<what is wrong, and what it costs>

## Evidence

<!--
One bullet per finding, quoting its headline.

EVERY claim in this issue traces to a finding id. A claim with no finding
behind it does not belong here — it belongs in a new finding, captured the
way the protocol says, and cited from the next version of this issue.

That rule is what makes an issue auditable in both directions: the issue
leads to its findings, and a filed finding leads to the issue it became.
-->

- `<finding-id>` — <the finding's headline, quoted>
- `<finding-id>` — <the finding's headline, quoted>

## Proposed change

<!--
The concrete edit, specific enough that someone who did not run the sprint
could make it.

  - For a DOCUMENT: the text to change, and what it becomes. Quote both.
  - For CODE or a SCRIPT: the behaviour to change, and what distinguishes
    done from not-done.

Propose ONE change. If the evidence supports two, that is two issues — the
issue is the unit of decision, and a reader must be able to accept one and
decline the other.

Do NOT include: alternatives considered, non-goals, an adoption plan, or a
resolution. Those are triage-workflow state and the tracker holds workflow
state; writing them here keeps a second copy of the tracker in prose.
-->

<the concrete edit>
