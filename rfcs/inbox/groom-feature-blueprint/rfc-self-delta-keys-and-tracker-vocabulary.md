---
title: "RFC: Two Phase 2 rules have no defined behavior for inputs the
  skill itself invites"
description: A work order derived from a sequencing constraint has no
  delta key to stamp, and a Tracker Reconciliation label outside the
  three-word vocabulary has no failure response short of demoting an
  approved blueprint
slug: rfc-self-delta-keys-and-tracker-vocabulary
type: rfc
status: ready-for-review
updated_at: 2026-08-04T00:00:00Z
origin: self
skill: groom-feature-blueprint
skill_version: "1.2"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.14.0"
---

# RFC: Two Phase 2 rules break on inputs the skill itself invites

## Summary

`groom-feature-blueprint` v1.2 was run against an approved feature
blueprint, producing an epic and 16 work orders. Two rules broke on
inputs the skill either explicitly invites or plausibly encounters: the
delta-key stamping rule assumes every work order derives from a delta
block, yet Phase 2 explicitly admits work orders derived from sequencing
constraints; and the Tracker Reconciliation rule enumerates exactly
three classifications with no stated response when a blueprint uses a
fourth word. Both forced unscripted improvisation, and in the second
case the agent knowingly deviated from the gate as written because the
prescribed failure response was disproportionate to the defect.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | Grooming `specs/features/spec-index/feature-spec-index.md` into beads — epic `agent-plugins-bge` plus 16 work orders, then handing the finishing pass to `groom-beads` |
| References / scripts / templates invoked | `blueprint-readiness-checklist.md` v0.5.0, `feature-blueprint-template.md` v1.3, `work-order-template.md` v1.2, `ac-traceability-standards.md` v0.2.0, `findings-capture-protocol.md` v0.2, `groom-beads` v1.1, `work-order-definition-of-ready.md` v1.1 |
| Run dates | 2026-08-04 |

The `model` frontmatter property is withheld: this session runs under a
harness policy forbidding the model identifier in repository artifacts.
The placeholder follows the convention already established by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, which
is the open RFC covering that gate conflict; this RFC does not restate
it. `effort` is recorded as `unknown` because the harness does not
expose it to the running agent.

## Evidence

1. **[gap · blocker] No rule for the delta key of a work order derived
   from a sequencing constraint.**
   `plugin/skills/groom-feature-blueprint/SKILL.md:107` states "every
   item must trace to a delta, a DoD entry, or a sequencing constraint",
   admitting three provenances. But
   `SKILL.md:222-223` states "Stamp each issue with its **delta key** as
   metadata or a `delta:<KEY>` label — the key is what re-grooming diffs
   by", and the task-file frontmatter at `SKILL.md:269-270` declares
   `delta: <KEY>  # delta key from the blueprint's AC Coverage legend`.
   A sequencing-constraint-derived item has no delta block and therefore
   no legend key. The blueprint under grooming directed exactly this
   case: "The `Profile` extension-point interface ... is a foundational
   sliver ... Grooming SHOULD lift it into its own early work item."
   The agent created that item and improvised, stamping it `delta:CO` —
   the key of the delta owning the package the interface lives in — and
   wrote a paragraph into the issue explaining that two items now share
   one delta key. Nothing in the skill sanctions or forbids that choice,
   and re-grooming (`SKILL.md:317-318`, "diffing by delta key") now sees
   an ambiguous key.

2. **[gap · blocker] No response defined for a Tracker Reconciliation
   label outside the three-word vocabulary.**
   `plugin/skills/spec/references/blueprint-readiness-checklist.md:82-85`
   (check 10) requires the section to classify "every listed issue as
   subsumed, prerequisite, or out-of-scope — or states 'None.'", and
   `SKILL.md:143-150` defines graph consequences for exactly those three
   words. The blueprint classified two issues as **"resolved"**
   (`agent-plugins-igc` and `agent-plugins-2l2`) — a word in neither
   list, whose meaning was nonetheless unambiguous in context: the
   blueprint had made the decision each issue tracked, and stated "The
   bead can be closed."

   Read literally, check 10 failed, and the skill's Phase 1 prescribes
   one response for a failed content check: write the gap into Open
   Questions and set the blueprint's status back to `draft`. **The agent
   did not do that.** It read "resolved" as a synonym for *subsumed*,
   applied the subsumed treatment (`SKILL.md:145-147`: blocked on the
   superseding work order plus a supersession note), and proceeded —
   because demoting an approved, otherwise-plannable blueprint to draft
   over a label synonym was plainly disproportionate to the defect. That
   deviation is recorded here rather than hidden: the skill offers no
   response between "pass" and "demote to draft", and the absence of a
   proportionate middle is what produced it.

## Proposals

### P1 — Define the delta key for non-delta-derived work orders [blocker]

**Problem.** `SKILL.md:107` admits work orders deriving from a delta, a
DoD entry, or a sequencing constraint, but `SKILL.md:222-223` and
`SKILL.md:269-270` both assume a delta key exists for every item and
name the AC Coverage legend as its only source. Two of the three
admitted provenances have no such key. The skill is silent at the point
of stamping, so the agent invents a convention, and `SKILL.md:317-318`
makes that convention load-bearing for re-grooming.

**Change.** Insert after `SKILL.md:223`:

> WHERE a work order derives from a sequencing constraint or a
> Definition of Done entry rather than from a delta block, it has no
> legend key. Stamp it `delta:none` and record its provenance instead:
> `derived:sequencing` or `derived:dod`, plus the section it traces to.
> Re-grooming diffs these items by title and provenance, not by delta
> key. A legend key SHALL NOT be borrowed from a neighbouring delta —
> two items sharing one key make the re-grooming diff ambiguous.

Mirror the same two labels in the task-file frontmatter block at
`SKILL.md:269-270`, changing the comment to note that `delta:` is
`none` for non-delta-derived items.

**Evidence.** Finding 1. The next run stamps the lifted `Profile` sliver
`delta:none` + `derived:sequencing` and leaves `delta:CO` uniquely
identifying the Corpus item, so a re-groom can tell the two apart
without reading their descriptions.

### P2 — Give Phase 1 a proportionate near-miss response [blocker]

**Problem.** Check 10
(`plugin/skills/spec/references/blueprint-readiness-checklist.md:82-85`)
admits three classification words; `SKILL.md:143-150` acts on those
three. Neither says what to do about a fourth word whose intent is
clear. Phase 1's only content-check failure response is to write an Open
Question and demote the blueprint to `draft`, which for a label synonym
costs a full re-approval cycle to fix a word. Faced with that, the agent
deviated from the gate rather than apply it (finding 2) — the outcome
the gate exists to prevent, produced by the gate's own rigidity.

**Change.** Add to `SKILL.md` Phase 2, after line 150:

> WHERE the Tracker Reconciliation section classifies an issue with a
> word outside {subsumed, prerequisite, out-of-scope} but states its
> disposition unambiguously (e.g. "resolved — the decision is X; the
> issue can be closed"), map it to the nearest of the three, apply that
> treatment, and record the mapping in the report and on the affected
> work order. Do NOT demote the blueprint for vocabulary alone. WHERE
> the intent is NOT unambiguous, that is a genuine check-10 failure and
> Phase 1's bounce applies.

State the general principle once in Phase 1, so it governs future
near-miss checks rather than only this one:

> A content check fails WHEN the design decision it guards is missing —
> not when a document expresses a present decision in unexpected words.
> Prefer mapping recognisable intent onto the vocabulary, reporting the
> mapping, over a bounce that costs a re-approval cycle.

**Evidence.** Finding 2. The next run applies the subsumed treatment to
a "resolved" issue by rule instead of by improvisation, reports the
mapping, and the human sees the vocabulary drift without the blueprint
losing its approved status.

## Considered and Not Proposed

- **Require the delta key to be globally unique across a feature's work
  orders.** This is the constraint P1 protects, but stating it as a
  validation rule would need a home in the Definition of Ready
  (`spec`-owned) and a check in `groom-beads`. P1's prohibition on
  borrowing a neighbour's key achieves the same outcome at the point of
  stamping, with no cross-skill change.
- **Widen check 10's vocabulary to include "resolved".** Rejected: the
  problem is not this particular word but the absence of any rule for
  unrecognised words. Adding a fourth leaves the fifth undefined.
- **The `manual` verified-by level with no `(manual)` marker in the
  FRD, and the readiness checklist's lack of a frontmatter-validity
  check.** Both are real findings from this run, but both concern
  `plugin/skills/spec/references/blueprint-readiness-checklist.md`,
  which the `spec` skill owns. Filed as a separate RFC against `spec`
  per the one-RFC-per-skill rule.
- **The skill's handling of a delta carrying no AC Coverage row.** The
  run hit this (the SP delta), but the fix belongs in the blueprint that
  omitted its verification basis, not in the grooming skill — filed in
  the project-audience RFC `rfcs/inbox/feature-blueprint/rfc-0010`.

## Non-Goals

- No change to Phase 3's beads serialization, the epic/child structure,
  or the `groom-beads` handoff — all three worked as written on this
  run.
- No change to the three-provenance rule at `SKILL.md:107`; P1 makes it
  consistent with stamping rather than narrowing it.
- Not weakening any content check. P2 narrows one failure response to
  vocabulary mismatches with unambiguous intent; a missing decision
  still bounces.

## Adoption

P1 and P2 are independent and touch disjoint regions of
`plugin/skills/groom-feature-blueprint/SKILL.md` — lines 222-223 and
269-270 for P1, the Phase 2 Tracker Reconciliation block and Phase 1 for
P2. Either can land alone.

Backward compatibility: P1 introduces two label values that no existing
plan uses, so previously-groomed plans stay valid; a re-groom of a plan
containing a borrowed key would re-stamp it, which is the intended
correction. P2 changes only a failure response, making the gate strictly
more permissive in one narrow case, so no blueprint that passes today
begins failing.

Expected effect: the two places this run had to improvise become
scripted, and one of them stops inviting an agent to deviate from the
gate to avoid a disproportionate bounce.

## Resolution
