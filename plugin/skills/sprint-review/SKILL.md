---
name: sprint-review
description: Review a sprint's own diff and turn what comes back into beads —
  invoke the reviewer the project declared in sf.toml's [review], scope it to
  this run's diff, file each finding as a bead against its origin, and classify
  each as in or out of the run's authorized scope. Use when a sprint
  orchestrator reaches its review step, or when asked to review a sprint's
  accumulated changes, triage review findings into beads, or decide which
  findings a run is allowed to take on. It does not review code itself and does
  not decide whether to loop.
version: "0.1"
user-invocable: false
---

# sprint-review

**We do not ship a code reviewer, and this skill is not one.** Harnesses
provide one, the ecosystem is full of them, and reviewing a diff is the most
generic thing in a sprint. What is not generic — and is entirely the sprint's —
is what happens to the findings.

So: run whatever the project declared, and route what comes back.

## Inputs

| | |
|---|---|
| `{EPIC}` | the run epic. Its manifest carries `base_sha` and `branch` |
| the authorization set | what the run was authorized to take on, from the gate |
| the pass number | which review pass this is. You do not decide whether there is another |

## What you return

- **The findings**, as beads that now exist — id, one line, and in-scope or
  out-of-scope.
- **The counts**: filed, admitted into this run, held out of it.
- **`review skipped — no reviewer declared`**, where that is the case, as the
  whole return.

You do not return a recommendation about looping. The orchestrator owns the
counter and the budget.

## 1 · Resolve the reviewer

Read `[review]` from `sf.toml`. Exactly one of two keys:

```toml
[review]
skill   = "code-review"          # dispatch a skill by name
# command = 'my-linter --diff'   # or run a command; its stdout is the report
```

**Undeclared is not an error — and not a pass.** Return
`review skipped — no reviewer declared` and stop. The orchestrator names the
skip in the report, because a silently skipped review reads exactly like a
clean one. Do not substitute your own review, and do not prompt for a
declaration mid-run.

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

## 4 · Classify in-scope against the authorization gate

Not against your judgement of what would be nice to fix.

| | |
|---|---|
| **in scope** | the finding is a defect in what this run's beads changed, and the gate authorized scope growth from review findings |
| **out of scope** | anything else — including a real, serious defect the run did not cause |

**The conservative default is out of scope**: file the bead, do not admit it
into this run, and let the report name it. A run that quietly grew to fix
everything it noticed is a run whose scope nobody agreed to, and its
pull request is unreviewable for it.

Admitted findings enter the run the way every other bead does — through
`sprint-plan`, groomed and gated like anything else. **You file; you do not
groom.** That is why the loop re-enters at grooming rather than dispatching
your findings directly: a review finding earns its readiness the same way, or
bounces the same way.

## Two things that are not yours

- **The loop.** The pass counter, the budget and the decision to re-enter are
  routing, and routing is the orchestrator's only job. You report what you
  found; it decides whether there is another pass.
- **Fixing anything.** You file beads. A review that also implements is a
  review nobody can audit.
