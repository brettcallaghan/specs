---
title: "RFC: A malformed query is not an empty answer, and a measured
  maximum is not a bound"
description: Give the output standards a rule separating an unrecognized
  query term from an empty result — which `check --only` currently
  conflates into a clean bill of health — and stop the Gherkin guide
  letting a freshly measured number become an acceptance criterion's
  limit
slug: rfc-0026-a-malformed-query-is-not-an-empty-answer
type: rfc
status: ready-for-review
updated_at: 2026-09-07T00:00:00Z
origin: agent-plugins
skill: retro-spec
skill_version: "1.5.0"
model: "claude-opus-5"
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.24.0"
---

# RFC: A Malformed Query Is Not an Empty Answer

## Summary

One session authored [[frd-spec-list]] and [[feature-spec-list]] and
groomed the blueprint into a plan. Two documents it consumed left a
decision unmade that the session had to make and argue for itself.

The blocker is live in shipped code. [[cli-output-standards]] rules that
a query matching nothing is an answer and exits zero, and says nothing
about a query naming something that does not exist. `spec check --only`
resolves that silence in the dangerous direction: an unrecognized rule
name is accepted, every rule is marked `not-requested`, and the command
prints `no findings` and exits zero. A caller who mistypes a rule name
receives a clean bill of health over a corpus that was never checked.

The second proposal is smaller and is about how a criterion is written.
The Gherkin guide teaches measurement and says nothing about what may be
written down from one, so an acceptance criterion in this session's FRD
pinned a number that was only the corpus maximum on the day it was
measured. A human reviewer caught it.

## Provenance

| Field | Value |
|-------|-------|
| Origin | agent-plugins (this repository critiqued its own documents) |
| Phase | Authored [[feature-spec-list]] from [[frd-spec-list]]; groomed it into epic `agent-plugins-kgir` |
| Skill | `retro-spec` v1.5.0 |
| Model | `claude-opus-5` |
| Effort | high |
| Harness | Claude Code, remote cloud session |
| Runs observed | 1 |
| Run date | 2026-09-07 |
| Documents consumed | [[cli-output-standards]] v0.1; [[gherkin-guide]] v1.2.0; [[frontmatter-standards]] v0.2; [[blueprint-readiness-checklist]] v0.7.0; [[work-order-definition-of-ready]] v1.5; [[ac-traceability-standards]]; [[container-spec-cli]]; [[module-document-core]] |
| Skills invoked | `spec` v1.10; `groom-feature-blueprint` v1.3; `groom-beads` v1.3 |
| Templates invoked | `frd-template.md` v1.3; `feature-blueprint-template.md` v1.4 |
| Tools invoked | `spec check`, `spec feature`, `spec criterion`, `spec slug-exists`, `spec index` (built from `cli/` at working-tree HEAD); `bd create/update/dep/lint/swarm validate` |
| Documents produced | [[frd-spec-list]], [[feature-spec-list]] |
| Plugin version | 0.24.0 |

## Evidence

1. **[missing-rule | blocker]** [[cli-output-standards]] governs the
   empty-result case and no other. `specs/standards/cli-output-standards.md`
   states, under "An empty answer is an answer":

   > WHERE a query matches nothing, the CLI SHALL render the empty
   > collection and SHALL exit zero.
   >
   > **Rationale:** nothing found is a fact about the corpus, not a
   > failure of the invocation. Exiting non-zero would make a caller
   > branch on the shape of the world rather than on whether the command
   > worked.

   No rule anywhere in that document governs a query whose *terms* are
   unrecognized — a filter value naming no registered thing. Searching
   the document for `malformed`, `unknown`, `unrecognized` and `invalid`
   returns nothing.

   The rationale quoted above is about the world; an unrecognized term is
   a fact about the invocation, which is the case the rule was explicitly
   scoped away from and which nothing else picks up.

2. **[defect | blocker]** `spec check --only` resolves that silence by
   accepting anything. Measured this session against the current working
   tree:

   ```
   $ spec check --only=totally-bogus-rule-name; echo "EXIT=$?"
   checked 59 documents
   ...
   no findings

   checks:
     corpus/empty  not-requested
     frontmatter/missing-property  not-requested
     frontmatter/parse  not-requested
     frontmatter/schema  not-requested
   EXIT=0
   ```

   Every rule is `not-requested`, the report says `no findings`, and the
   exit status is zero. The output is indistinguishable — to a script
   branching on the exit status — from a corpus that was checked and
   conformed. `check` is the gate the `spec` skill's step 7 makes
   authoritative, so a typo in a CI invocation disables it silently.

3. **[missing-rule | snag]** [[gherkin-guide]] teaches measurement but
   sets no limit on what may be written down from one. Its "Writing good
   scenarios" section says:

   > **State constraints, not justification.** "Membership is by content,
   > not path" tells a reader what they may not change.

   and, on data:

   > **Prefer `Scenario Outline` where a criterion varies over data.**

   Nothing addresses a value that is *measured rather than required*.
   Searching the document for `measur`, `observed`, `maximum` and `bound`
   returns only its own prose about Gherkin's parser, none of it about
   criteria.

   This session wrote AC-SPECLIST-001.3 as:

   > Given a member whose description is 280 characters long

   280 was simply the longest description in the corpus that day
   (`document-relationship-standards`; n=58, min 68, median 173).
   Measured this session, nothing enforces any limit: the generated JSON
   Schema for `description` is `{"type": "string"}` with no `maxLength`,
   and `grep -rn 'maxLength\|MaxLength' cli/core cli/profile schemas/`
   returns nothing outside tests. The criterion therefore read as a bound
   that does not exist, and would have been falsified by the next longer
   description anyone wrote. A human reviewer caught it in review of
   pull request 39.

   The pressure toward this mistake is created by guidance working as
   intended: the `spec` skill instructs "Ask the corpus rather than
   reading it", so an author arrives at criterion-writing time holding
   fresh measurements. Nothing then says which of them may be written
   into a criterion.

## Proposals

### P1 — Rule that an unrecognized query term is not an empty result [blocker]

**Problem.** [[cli-output-standards]]' "An empty answer is an answer"
governs a query that *matches* nothing and is silent on a query that
*names* nothing, and the silence is being resolved wrongly in shipped
code. Finding 1, Finding 2.

Both cases render as an empty collection and exit zero today, so a caller
cannot distinguish "this corpus has no contracts" from "you misspelled
the filter". The first is information; the second is the caller's bug,
and returning zero hides it at exactly the moment a human would want to
know.

**Change.** Add a rule to `specs/standards/cli-output-standards.md`,
directly after "An empty answer is an answer", with this wording:

> ### An unrecognized query term is not an empty answer
>
> WHERE a command's argument or flag value names a member of a
> registered set — a document type, a rule identifier, a tag — and the
> value names no member of that set, the CLI SHALL report the
> unrecognized value together with the registered names, and SHALL exit
> non-zero.
>
> **Rationale:** the rule above is about the world, and this one is
> about the invocation. A corpus containing no contracts is a fact worth
> reporting as an empty collection; a filter value that names nothing is
> a caller's mistake, and answering it with an empty collection makes a
> typo indistinguishable from an empty category. Naming the registered
> values makes the report self-correcting, which is what the profile
> registry already does for an unregistered `[profile].name`.

**Backward compatibility.** One shipped behavior changes: `check --only`
begins rejecting rule names it currently accepts. That is the defect in
Finding 2, so the change is the fix rather than a cost. No committed
document or snapshot asserts the current behavior — `cli/testdata/cases/`
contains a `check-only` case, which exercises valid rule names only and
is unaffected. Commands with no registered-set argument are untouched.

**Evidence.** Findings 1 and 2. [[frd-spec-list]] already had to make
this exact ruling for `spec list --type` and argue it in prose beneath
FEAT-SPECLIST-002, because no standard supplied it; with this rule the
FRD would have cited the standard in one line instead.

### P2 — Forbid a measured value standing as a criterion's bound [snag]

**Problem.** [[gherkin-guide]] gives no rule about writing a measurement
into a criterion, and an author who has just measured the corpus — as
the `spec` skill instructs — has a number to hand at exactly the moment
they are writing criteria. Finding 3.

**Change.** Add to `specs/standards/gherkin-guide.md`, in "Writing good
scenarios", after "State constraints, not justification":

> **A measurement is not a requirement.** A number you obtained by
> measuring the corpus states what is true today; written into a
> criterion it reads as a limit, and the next document that exceeds it
> falsifies the criterion rather than violating anything. WHERE a
> criterion needs a concrete extreme, name it relative to the corpus —
> "the longest description in the corpus" — which stays testable and
> asserts no bound. Write a number only where something actually
> enforces it, and say what does.

**Backward compatibility.** Documentation only; no existing criterion
becomes invalid. AC-SPECLIST-001.3 in [[frd-spec-list]] has already been
rewritten to the corpus-relative form and is the worked example.

**Evidence.** Finding 3.

## Considered and Not Proposed

- **The `fb-` versus `feature-` slug disagreement.** Authoring
  [[feature-spec-list]] produced `fb-spec-list` by following
  `feature-blueprint-template.md`, and `spec check` rejected it against
  the profile's `^feature-[a-z0-9]+(-[a-z0-9]+)*$`. **Not re-proposed,
  because it is already filed** as
  `rfcs/inbox/spec/rfc-self-slug-prefix-and-unrunnable-checks.md`, P1,
  severity blocker, status `ready-for-review` since 2026-08-07.
  Verified still live this session:
  `specs/standards/template-authoring-standards.md:135` still carries a
  `fb-` row in its slug-prefix table for Feature Blueprint,
  the template still says `fb-` at lines 5 and 131-132, and
  `cli/profile/profile.go:163` still enforces `feature-`. Recorded here
  as **recurrence evidence**: that RFC declares `runs_observed: 1` and
  the defect has now cost a second run, one month and nine plugin
  versions later (0.15.0 → 0.24.0, `spec` 1.7 → 1.10). Triage should
  raise its recurrence count rather than open a duplicate.
- **Acceptance criteria that assert what a document says, without
  `@manual`.** AC-SPECLIST-005.1 and 005.2 were authored untagged and so
  carried a test-citation obligation under [[ac-traceability-standards]]
  that nothing could discharge; caught while mapping the coverage table
  and fixed. Not proposed: [[gherkin-guide]] already states the rule
  ("IF an acceptance criterion can only be verified manually, THEN tag it
  `@manual`"). The fix for ignoring an existing rule is not another rule.
- **Reading an exit status through a pipe.** `bd-health.sh` was first run
  as `... | tail`, reporting `EXIT=0` for a script that had exited 1. Not
  proposed: `AGENTS.md` documents this trap at length under "Pipes hide
  exit statuses", with this exact remedy. Carelessness against existing
  guidance.
- **`groom-feature-blueprint` directing acceptance criteria into a bead's
  description.** Its Phase 3 says to put "the AC list, verified-by
  levels, blueprint section references ... in each issue's description",
  while `groom-beads`' work-order-to-bead mapping sends `## Acceptance` to
  the bead's `acceptance` field. Following the first produced six beads
  the second's Definition of Ready then flagged. **Routed to
  `retro-skill`, not filed here**: the fix lands in
  `plugin/skills/groom-feature-blueprint/SKILL.md`, a skill this
  repository vends, which is the skill lens rather than the document lens.
- **Two candidate findings eliminated by measurement.**
  [[blueprint-readiness-checklist]] directs enumerating ACs with
  `spec feature <id>`; run this session, it lists every criterion
  correctly, so the instruction is right. And
  [[spec-cli-reference]] was suspected of not saying how to obtain the
  binary when it is absent from `PATH`; it does, under "When it is not
  installed", naming `go install`. Neither is a defect.

## Non-Goals

- Changing the CLI's exit-status vocabulary. P1 requires only that the
  status be non-zero, leaving which non-zero status to
  [[frd-spec-index]]'s existing definitions.
- Re-litigating whether feature blueprint slugs should be `fb-` or
  `feature-`. That choice belongs to the open RFC named above.
- Adding a length limit to `description`. [[frd-spec-list]]'s
  FEAT-SPECLIST-005 argues deliberately against one — a length rule would
  pass a useless description of the right size — and P2 is about how a
  criterion is written, not about constraining the property.

## Adoption

P1 and P2 are independently adoptable; they touch different documents and
neither depends on the other.

P1 lands in two parts, and they SHALL land together: the rule in
[[cli-output-standards]], and the `check --only` change that makes the
shipped surface conform to it. Adopting the rule alone would leave the
standard documenting behavior the CLI contradicts, which is the condition
this RFC exists to remove. The `--only` defect is filed as
`agent-plugins-js5h` so it is not lost if P1 is deferred.

P2 is documentation only and lands in one edit. Both documents are
vendored into the skills by `scripts/sync-shared-references.sh`, which
SHALL be re-run so the generated copies under
`plugin/skills/*/references/shared/` do not drift.

## Resolution

<!-- Filled by the resolving review. -->
