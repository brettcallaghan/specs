# Dispatch Plan Format

The format of `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md`, the
run's sprint plan. Read this when you write the plan, and again whenever the
orchestrator amends it. It is written for whoever picks the sprint back up —
which may be an agent with no memory of having written it.

**The run manifest is the orchestrator's, not yours.** It is already on disk
when this skill is invoked; write the wave table below it and leave the header
alone. The manifest's format is documented here so both authors write the same
file.

## Contents

- Run manifest — the facts recovery cannot re-derive; a gate whose baseline is
  `fail` sets "no new findings", not exit 0, as the bar for every bead
- Wave table — names beads, never a filter; every row carries class, bucket,
  tier and push mode, and its Reason names which collision class applies
- Reading C4 overlap — compute co-edit and sequencing edges from bead metadata,
  not from file paths
- Amendments — append-only; when the run diverges from the plan, log the change
  AND update the wave table in place

Every run gets one, whatever its size. A two-bead run interrupted at bead one
needs the plan exactly as much as a twenty-bead run does.

## Run manifest

The header, before the wave table, **written by the orchestrator itself**.
Most of it is the stdout block from bead-bonanza's `scripts/pre-flight-check.sh`;
this file is where that output lands durably.

```markdown
# Dispatch plan — run <run-id>

| Field | Value |
|-------|-------|
| Run id | `<run-id>` |
| RUN_DIR | `/tmp/claude/bead-bonanza/<run-id>` |
| Branch | `<branch>` |
| Base sha | `<sha at run start>` |
| Gate profile | format=`…` lint=`…` build=`…` test=`…` regen=`…` custom=`…` |
| Gate baseline | `<gate>=pass|fail` per gate, from pre-flight (always captured). The full text, including each known-red gate's own output, is at `$RUN_DIR/gate-baseline.md` — carry that path, and inline the file's contents into every subagent prompt. A baseline that stops here is one the agent running the gate never sees. |
| Model lineup | light=`…` standard=`…` advanced=`…` frontier=`…` |
| Standing authorization | what the user authorized without further asking (push mode, PR, scope) |
| Wave preview | `accepted` \| `overridden — <why>` \| `none (no C4 docs)` \| `not returned` |

## Environmental ledger

Facts established this run that no wave should re-diagnose. Admitted on
evidence, not on reasoning: every entry — including a claim of sufficiency
("this check verifies X"), not only a claim of impossibility — names the
command that proved it and the result observed; run it first, record it
after. Note the conditions the command ran under when they could change the
result (attached repos, credentials, cache warmth, toolchain version). A
claim that cannot be tested cheaply is recorded as `unverified` with what
would settle it, not omitted. Keep the numbering — dispatch prompts cite
entries by number ("ledger entry 1", "ledger entry 2").

1. `<claim>` — proved by `<command>`, observed `<result>`
1. `<claim>` — unverified; would be settled by `<command>`
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
| 2 | demo-1kq | shared | shared-batch #2 | standard | defer | same c4_changes: auth-service as demo-4tp (no file overlap) |
| 2 | demo-4tp | shared | shared-batch #2 | standard | defer | same c4_changes: auth-service |
| 3 | demo-7wb | isolated | sequential | standard | — | consumes auth-service, changed by demo-1kq |
| 3 | demo-d3j | isolated | sequential | advanced | — | validator change; different surface |
```

`Tier` is the bead's `metadata.execution_suggested_model` (else
`metadata.tier`), resolved to a model at dispatch via the
project's `sf.toml` `[models]` lineup. Leave it blank only when the project declares no
lineup — the subagent then inherits the session model.

### The Reason column names the collision class

The two collision classes have different consequences, so they must not read
alike. Spell out which one a row is:

- **Co-edit collision** — names the shared surface: `same c4_changes:
  auth-service`, `same fixture cluster`, `both edit testdata/golden.json`. These
  beads may share a bucket and a batch; they may never share a
  concurrent wave.
- **Sequencing constraint** — names what is consumed and whose change it
  consumes: `consumes auth-service, changed by demo-1kq`. This is an *ordering*
  edge, not a batching one; it can hold with zero file overlap. The consumer
  goes in a later wave, or after the producer inside one bucket. WHEN the
  ordering must survive recovery or a re-plan, wire a real `blocks` edge and say
  so in the row.
- **Signal disagreement** — say both signals and that you took the union:
  `union: disjoint c4, overlapping fixture testdata/golden.json`. A
  disagreement usually means stale C4 tagging or a bead scoped wider than its
  tags say, so it is worth a line even when the packing is unaffected.

## Reading C4 overlap

Phase 2.5's primary overlap signal. `bd`'s JSON gives `metadata` as an object,
so `c4_changes` / `c4_impacts` come back as real arrays:

```bash
bd list --status open --json | jq -r '
  [ .[] | {id, ch:(.metadata.c4_changes // []), im:(.metadata.c4_impacts // [])} ] as $b
  | ( [ $b[] as $x | $b[] as $y | select($x.id < $y.id)
        | ($x.ch - ($x.ch - $y.ch)) as $s | select($s|length > 0)
        | "co-edit  \($x.id) + \($y.id): \($s|join(","))" ]
    + [ $b[] as $a | $b[] as $c | select($a.id != $c.id)
        | ($c.im - ($c.im - $a.ch)) as $s | select($s|length > 0)
        | "sequence \($c.id) after \($a.id): consumes \($s|join(","))" ] ) | .[]'
```

`$x.ch - ($x.ch - $y.ch)` is array intersection in `jq`; `// []` is what keeps
an untagged bead from erroring out instead of reading as "no components".
Filter the list to the run's queue first — a repo-wide `bd list` will pair
queued beads with beads nobody scoped.

Worked example, four beads:

| Bead | `c4_changes` | `c4_impacts` |
|------|--------------|--------------|
| A | `auth-service` | `billing-service` |
| B | `auth-service` | — |
| C | `web-spa` | `auth-service` |
| D | `reporting-service` | — |

A + B collide on `auth-service` even if they share no file: shared-surface, one
batch, never the same parallel wave. C changes a different component and shares
no file with either, but consumes `auth-service` — a sequencing constraint, so C
runs after both. D is isolated on both signals. Nothing in this reads a file
path, which is the point: the file scan would have parallelized A with B and
seen no relationship at all between C and them.

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
