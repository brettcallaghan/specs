---
name: sprint-review
description: Review a sprint's own diff and turn what comes back into beads —
  resolve a reviewer (the one declared in sf.toml's [review], else one found in
  the session, else a bare agent), scope it to this run's diff, file each
  finding as a bead against its origin, and classify each as in or out of the
  run's authorized scope. Use when a sprint
  orchestrator reaches its review step, or when asked to review a sprint's
  accumulated changes, triage review findings into beads, or decide which
  findings a run is allowed to take on. It does not review code itself and does
  not decide whether to loop.
version: "0.2"
user-invocable: false
---

# sprint-review

**We do not ship a code reviewer, and this skill is not one.** Harnesses
provide one, the ecosystem is full of them, and reviewing a diff is the most
generic thing in a sprint. What is not generic — and is entirely the sprint's —
is what happens to the findings.

**That does not mean a run may go unreviewed.** Shipping no reviewer is a
statement about what this framework builds, not about what a run does. Every
run gets a review; step 1 is how one is found.

So: run whatever the project declared, and route what comes back.

## Inputs

| | |
|---|---|
| `{EPIC}` | the run epic. Its manifest carries `base_sha` and `branch` |
| the authorization set | what the run was authorized to take on, from the gate |
| the pass number | which review pass this is. You do not decide whether there is another |
| `{DISPATCH_ID}` | your agent id — `review-<run-id>`, handed to you, never minted |

**Before your first `bd` call: `export BEADS_ACTOR={DISPATCH_ID}`.** You file
beads and comment on them, so you write; unset, those writes land under the
harness default and `bd ready --assignee review-<run-id>` stops answering for
your phase. `<run-id>` is `{EPIC}` with the project prefix dropped. The roster
of agent ids and the derivation rule are the coordinator's dispatch policy's,
their only copy — take yours from there rather than inventing one.

## What you return

- **The findings**, as beads that now exist — id, one line, and in-scope or
  out-of-scope.
- **The counts**: filed, admitted into this run, held out of it.
- **`reviewer_source`** — `declared`, `discovered`, or `none`, and the name of
  whatever ran. The report needs it: a review by the project's declared
  reviewer and a review by a bare agent are not the same evidence, and a report
  that cannot tell them apart lets a weak review read like a strong one.

You do not return a recommendation about looping. The orchestrator owns the
counter and the budget.

## 1 · Resolve the reviewer

Read `[review]` from `sf.toml`. Exactly one of two keys:

```toml
[review]
skill   = "code-review"          # dispatch a skill by name
# command = 'my-linter --diff'   # or run a command; its stdout is the report
```

**Undeclared is not an error, not a pass, and not a skip.** Take the first of
these that resolves, and report which one:

| | `reviewer_source` | |
|---|---|---|
| 1 | `declared` | `[review]` names a skill or command. Use it. |
| 2 | `discovered` | Undeclared: look for a review skill available in this session and use it. Name the one you picked. |
| 3 | `none` | No skill available: dispatch a plain agent to review the diff with no skill, and say the review ran without one. |

**Never prompt for a declaration mid-run** — the run has one human touchpoint
and it is already past.

**Why not skip.** A skipped review reads exactly like a clean one, which is why
skipping was never safe; but returning nothing also threw away a review the
session could have performed. Falling back keeps the review and keeps the
distinction, as long as `reviewer_source` reaches the report — which is the
condition, not a nicety.

## 2 · Scope it to this run's diff

```sh
base="$(bd show {EPIC} --json | jq -r '.[0].metadata.base_sha')"
git diff "$base"...HEAD
```

**`base_sha` from the manifest, never a merge-base you compute.** The manifest
records what the run branched from at the moment it branched; a merge-base
recomputed now assumes the base branch has not moved, and on a long run it
will have — which silently widens the review to include other people's work.

Where the manifest carries no `base_sha`, say so and review nothing. A review
of an unknown range is worse than no review, because it reports with the same
confidence.

## 3 · File every finding as a bead

One `bd create` per finding, against its origin:

```sh
bd create --title "<finding>" --description "<what, where, why it matters>" \
  --type=bug --priority=<n> --parent {EPIC} --discovered-from <origin-bead-id>
```

`--discovered-from` is what keeps a finding attached to the bead whose change
produced it, so a later reader can ask "what did this bead cost us" and get an
answer. Where a finding is not attributable to one bead, parent it to the run
epic and say so in its description.

**A finding is filed whether or not this run acts on it.** Filing is free;
losing a finding because the run declined to fix it is not.

**Cite a durable anchor, never a position.** A path plus the quoted string, or
a class, function or component name, is what a dispatched agent can still find;
a line number rots between filing and dispatch, and has — two of six in one
run's findings had moved by the time the bead was dispatched, both because a
sibling bead in the same batch edited the file. Where a line number helps,
write it as a hint beside the string, never as the handle. **Quote the evidence
INTO the bead** rather than citing a document that holds it: a bead whose only
evidence is a path stops being workable the moment that path moves.

## 4 · Classify each finding against three criteria

Not against your judgement of what would be nice to fix. Take the first row
that matches:

| the finding is | verdict |
|---|---|
| a defect **this run's own merged diff introduced** — it did not exist at `base_sha` | **IN SCOPE**, whatever the gate says |
| a defect in code this run **touched** but did not introduce | in scope **only if** `standing_auth.admit_review_findings` is true |
| anything else — including a real, serious defect the run neither introduced nor touched | **out of scope** |

**Row 1 does not consult the gate, and that is deliberate.** A defect the run
shipped is not scope growth; it is the run's own deliverable not working, and a
run that merges a regression it found in its own diff has failed at the thing
it was for. Across three runs on this corpus the gate's row was never once
answered by a human, took its `false` default every time, and was overridden by
the coordinator in the two most recent runs — a default overridden every time
it is reached is not a default.

**Two things row 1 does NOT license.** It does not admit a finding whose only
remedy is prose clarifying a document that has not yet misled anybody — file
it, because the value of a documentation nit is not worth a dispatch and a
review pass. And it does not admit a finding against a document the run is
already replacing: fixing text that is about to be deleted spends a review loop
on nothing.

**Everything else defaults to out of scope**: file the bead, do not admit it,
and let the record name it. A run that quietly grew to fix everything it
noticed is a run whose scope nobody agreed to, and its pull request is
unreviewable for it.

Admitted findings enter the run the way every other bead does — through
`sprint-plan`, groomed and gated like anything else. **You file; you do not
groom.** That is why the loop re-enters at grooming rather than dispatching
your findings directly: a review finding earns its readiness the same way, or
bounces the same way.

**Say which row decided each finding**, so the record can report how much of
the run's own diff it fixed and how much it filed.

## Two things that are not yours

- **The loop.** The pass counter, the budget and the decision to re-enter are
  routing, and routing is the coordinator's only job. You report what you
  found; it decides whether there is another pass. **Findings admitted under
  row 1 still consume a pass** — the budget bounds how many times the machine
  reviews itself, not how serious the findings were.
- **Fixing anything.** You file beads. A review that also implements is a
  review nobody can audit.
