---
title: "RFC: <Title Spanning the Skill Improvements>"
description: <One-sentence summary of the skill defects/gaps fixed>
slug: rfc-<origin>-<short-description>  # numbered by the framework
                                        # repo at triage
type: rfc
status: draft
updated_at: <YYYY-MM-DDTHH:MM:SSZ>
# The fields below are the routing/weighting signals: they map 1:1 onto
# the triage issue's labels. retro-skill RECORDS them; triage DECIDES.
origin: self | <project>               # self = this repo critiqued its
                                       # own skill; else the run's project
skill: <skill-name>                    # the skill this RFC critiques
skill_version: <X.Y>                   # its SKILL.md frontmatter version
model: <e.g. claude-opus-4-8[1m]>      # the model that ran the skill
effort: <low | medium | high | xhigh | max>
runs_observed: <N>                     # runs standing behind the evidence
severity: <blocker | snag | polish>    # highest severity among proposals
plugin_version: "<x.y.z>"              # or "n/a (standalone install)"
related: []
---

<!--
Template: Skill-Improvement RFC
Version: 0.2
Source: spec-driven-dev framework (produced by the retro-skill skill)
Lineage: a skill-focused fork of retro-spec's rfc-template — drops
  audience (retro-skill RFCs are inherently skill-authoring), adds the
  label-level frontmatter (skill, skill_version, model, effort,
  runs_observed, severity), the three finding shapes, and a per-proposal
  Severity. Keep the two in sync where they overlap.

Sections (complete each, then delete its instruction block)
  - Summary
  - Provenance
  - Evidence
  - Proposals
  - Considered and Not Proposed
  - Non-Goals
  - Adoption
  - Resolution

Purpose
  Proposes evidence-backed improvements to ONE skill (its SKILL.md,
  references, scripts, or templates) from a real run of it. The RFC
  argues from observed run friction — misleading instructions, failed
  scripts, forced guesses — toward changes that make the next run
  succeed.

Scope — one RFC per skill
  Bundle all qualifying findings about ONE skill into one RFC; findings
  about a different skill get their own RFC. Within an RFC the proposals
  share one Provenance and one Evidence base; the PROPOSAL is the unit
  of decision, resolved individually at triage.

When to use / not use
  - Use: an agent ran a skill and its guidance could have guided better
    (a wrong action, a failed command, or a forced guess occurred).
  - Do not use: no run evidence ("seems better" is a discussion); or
    the fix is to a project's own document content (that is retro-spec
    or a direct edit), not to the skill.

Numbering and location
  - No number from the author. Downstream project:
    rfcs/outbox/rfc-<origin>-<slug>.md, copied into the framework repo's
    rfcs/inbox/. Inside the framework repo: write directly to
    rfcs/inbox/. Triage assigns the number on resolution.

Readiness gate — verify EVERY item before ready-for-review:
  [ ] Frontmatter sets model + effort (and Provenance names the
      harness) — a skill defect is model-relative
  [ ] Frontmatter runs_observed is set; no "every run" claim exceeds it
  [ ] Frontmatter skill + skill_version are set, and Provenance lists
      the versions of every reference/script/template invoked
  [ ] Frontmatter severity equals the highest severity among proposals
  [ ] Every finding is tagged with one shape (misleading-text /
      script-defect / gap) and one severity (blocker / snag / polish)
  [ ] Every misleading-text finding quotes the source verbatim with
      file:line, against the skill_version run (from the transcript if
      the tree has since moved); every gap names the step where the
      rule should live
  [ ] Every script-defect finding gives the exact command, observed vs
      expected, and a minimal repro
  [ ] Every Change is concrete enough to apply without further design
      (proposed wording, the script patch, the rule to insert)
  [ ] Proposals are ordered by severity, blockers first
  [ ] Findings that did NOT clear the bar appear under Considered and
      Not Proposed with the reason
  [ ] The Resolution section is present and empty

General instructions
  1. Complete each section per the comments, then delete all comment
     blocks.
  2. Keep lines within 80 characters per [[template-authoring-standards]].
  3. Quote skill text with repo-relative paths in backticks and line
     numbers; reference framework docs with [[slug]].
-->

# RFC: <Title Spanning the Skill Improvements>

## Summary

<!--
2-4 sentences: which skill was run, on what task, the class of friction
observed, and the one-line thesis ("the skill assumes one input file is
one RFC", "the harvest script's default filter hides the closure set").
-->

<What was run, what went wrong, what the proposals fix>

## Provenance

<!--
The label-level provenance — skill, skill_version, model, effort,
runs_observed, origin, severity — is in the FRONTMATTER (it maps onto
the triage issue's labels); do not repeat it here. This block holds the
human-context facts that do not become labels. MODEL and RUNS OBSERVED
(frontmatter) are load-bearing: a defect may be model-relative, and no
"every run" claim may exceed runs_observed.
-->

| Field | Value |
|-------|-------|
| Harness | <e.g. Claude Code CLI vX.Y> |
| Task the skill was applied to | <what the run was actually doing> |
| References / scripts / templates invoked | <name vX.Y, repo-relative path> |
| Run dates | <YYYY-MM-DD[, ...] — one per run behind runs_observed> |

## Evidence

<!--
The shared factual base every proposal cites. One numbered finding per
observed defect, each tagged [shape · severity], stated as what
HAPPENED:

  1. [misleading-text · blocker] `<file>:<line>` "<verbatim quote>" —
     <what the agent did because of it>.
  2. [script-defect · blocker] `<command>` — observed <X>, expected
     <Y>. Repro: <steps>.
  3. [gap · snag] No rule at <skill> step <N> for <situation> — the
     agent <what it improvised>.

Recurrence: where a finding recurred, cite each run; do not claim more
recurrence than Runs observed supports.
-->

1. <Finding>
2. <Finding>

## Proposals

<!--
One ### per proposal, ordered by severity (blockers first), IDs P1..Pn.
-->

### P1 — <Imperative title of the change> [severity]

**Problem.** <Why the current skill produces the defect. Quote the
misleading text with file:line, name the failing script + command, or
state the step where no rule exists.>

**Change.** <The concrete change: proposed SKILL.md wording, the script
patch (diff or exact new command), the rule to insert at step N. Ready
to apply.>

**Evidence.** <Finding number(s), and the expected effect on the next
run under the same model ("the split step is declared, so a multi-RFC
input no longer needs improvisation").>

## Considered and Not Proposed

<!--
Ideas examined and deliberately NOT proposed, each with the reason
(no run evidence, style-only, one-off carelessness the skill already
warns against, model-specific and unreproduced). Write "None." if empty.
-->

- <Idea> — <why not>

## Non-Goals

<!--
What these proposals deliberately do not change. Write "None." if empty.
-->

- <Non-goal>

## Adoption

<!--
Ordering and grouping: which proposals are independent, which must land
together, the files each touches, and backward compatibility (does the
change alter behavior for runs under other models or existing project
copies?). End with the expected overall effect on future runs.
-->

<Ordering, grouping, compatibility, expected effect>

## Resolution

<!--
FILLED BY triage-rfc — leave empty; keep the heading. Triage records,
per proposal: adopted (as proposed / with refinement), queued/watch
(bead id + theme), rejected (rationale), or returned. Plus date, files
changed, and version bumps.
-->
