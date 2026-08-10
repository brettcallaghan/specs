# bead-bonanza Dispatch Plan Format

The format of `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md`, the
run's sprint plan. Read this at Phase 2.5 when you write the plan, and again
whenever you amend it. It is written for whoever picks the sprint back up —
which may be an agent with no memory of having written it.

## Contents

- Run manifest — the facts recovery cannot re-derive
- Wave table — the enumerated beads, with class, bucket, tier and push mode
- Amendments — the append-only log that keeps the plan true mid-run

Every run gets one, whatever its size. A two-bead run interrupted at bead one
needs the plan exactly as much as a twenty-bead run does.

## Run manifest

The header, before the wave table. Most of it is the stdout block from
`../scripts/pre-flight-check.sh`; this file is where that output lands durably.

```markdown
# Dispatch plan — run <run-id>

| Field | Value |
|-------|-------|
| Run id | `<run-id>` |
| RUN_DIR | `/tmp/claude/bead-bonanza/<run-id>` |
| Branch | `<branch>` |
| Base sha | `<sha at run start>` |
| Gate profile | format=`…` lint=`…` build=`…` test=`…` regen=`…` custom=`…` |
| Gate baseline | `<gate>=pass|fail` per gate, from pre-flight `--gate-baseline` |
| Model lineup | light=`…` standard=`…` advanced=`…` frontier=`…` |
| Standing authorization | what the user authorized without further asking (push mode, PR, scope) |

## Environmental ledger

Facts established this run that no wave should re-diagnose. One line each,
naming the command that proved it.

1. `<claim>` — proved by `<command>`, observed `<result>`
```

A gate recorded `fail` in the baseline is the bar for every bead in this run:
"no findings beyond the baseline", not "exit 0". A gate with no recorded
baseline is expected to exit 0.

## Wave table

**The plan names beads, never a filter.** The filter produced the queue; the
plan IS the queue. A filter re-evaluated later — at recovery, or at the next
wave — returns a different set than the one this run committed to, and admits
beads nobody scoped. Beads discovered mid-run enter only through Phase 1.5
step 4's scope test, and only as an amendment.

```markdown
| Wave | Bead | Class | Bucket | Tier | Push | Reason |
|------|------|-------|--------|------|------|--------|
| 1 | demo-89s | trivial | fast-lane #1 | light | defer | docs only |
| 1 | demo-aqw | trivial | fast-lane #1 | light | defer | comment only |
| 2 | demo-9h0 | shared | shared-batch #1 | standard | defer | same serializer + fixture cluster |
| 2 | demo-04y | shared | shared-batch #1 | standard | defer | same cluster |
| 2 | demo-ojt | shared | shared-batch #1 | standard | defer | same cluster |
| 3 | demo-d3j | isolated | sequential | advanced | — | validator change; different surface |
```

`Tier` is the bead's `tier:` label, resolved to a model at dispatch via the
project's `## model lineup`. Leave it blank only when the project declares no
lineup — the subagent then inherits the session model.

## Amendments

`dispatch.md` is **append-only and amendable**. WHEN the plan changes mid-run —
a wave resequenced, a bucket reclassified, a bead admitted or bounced, a tier
escalated — append a dated entry under `## Amendments` stating what changed and
why, AND update the wave table in place. The plan on disk SHALL describe the run
as it is actually being executed; a plan that silently stops describing the run
is worse than no plan, because recovery trusts it.

Mirror one line of each amendment to the tracker note that carries the
`RUN_DIR` pointer, so a reader who has only the tracker knows the plan moved.

```markdown
## Amendments

- **2026-08-10 21:41** — wave 5 (`demo-9kb.6`, `demo-9kb.7`, `demo-9kb.9`)
  re-planned from one shared batch to strictly sequential: .7's change to the
  shared fixture invalidates .9's expected bytes, so they cannot share a gate
  run. Wave table updated in place.
- **2026-08-10 22:05** — `demo-9kb.8` escalated standard → advanced after two
  failed gate cycles. Tier column updated in place.
```
