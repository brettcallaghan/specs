# Run Record Format

**The run's record is its epic**, not a file: the manifest is `--metadata` JSON
on it, the amendments are `amend:`-prefixed comments on it, and the wave table
is returned rather than stored. There is no `dispatch.md` and no run directory —
the tracker outlives a reclaimed container, and a `/tmp` path does not.

Read this when you produce a wave table, and again whenever the orchestrator
amends the run. It is written for whoever picks the sprint back up — which may
be an agent with no memory of having written it.

**The manifest is the orchestrator's, not yours.** It is already on the epic
when this skill is invoked. Read it; never rewrite it.

## Contents

- Run manifest — `--metadata` on the run epic; the facts recovery cannot
  re-derive. A check whose baseline is `fail` sets "no new findings", not
  exit 0, as the bar for every bead
- Wave table — returned to the orchestrator, never stored; names beads, and
  every row carries class, bucket, tier and push mode, with its Reason naming
  which collision class applies
- Reading C4 overlap — compute co-edit and sequencing edges from bead metadata,
  not from file paths
- Amendments — `amend:` comments on the epic, append-only by construction

Every run gets one, whatever its size. A two-bead run interrupted at bead one
needs its record exactly as much as a twenty-bead run does.

## Run manifest

`--metadata` JSON on the run epic, **written by the orchestrator itself** from
`scripts/pre-flight-check.sh`'s output plus what admission decided. Retrieve it
whole with `bd show <epic-id> --json | jq .metadata`.

**`branch` and `base_sha` are the orchestrator's, written once.** It creates
the sprint branch immediately before dispatching wave 1 — after grooming, so
`base_sha` includes the grooming commits — and writes both keys then. A run
whose manifest already carries them **reuses** the branch and creates nothing;
that is what makes a resumed run land where the original did. Three things read
them and none may write them: the review scopes its diff `base_sha`..sprint
head, every worktree branches from that branch's HEAD, and the pull request
opens against it.

```json
{
  "branch": "<branch>",
  "base_sha": "<sha at branch creation, after grooming>",
  "checks_ids": "format vet build test-go …",
  "checks_baseline": "<verbatim text — inline this into EVERY subagent prompt>",
  "checks_baseline_status": 1,
  "models": {"light": "…", "standard": "…", "advanced": "…", "frontier": "…"},
  "effort": "low|medium|high|xhigh|unrecorded",
  "standing_auth": "what the user authorized without further asking",
  "trailers": "<commit trailers, or absent when the run resolved none>",
  "admitted": ["<bead-id>", "…"],
  "excluded": [{"id": "<bead-id>", "reason": "declined move out of <epic-id>"}],
  "moved": [{"id": "<bead-id>", "from": "<previous parent>"}],
  "wave_preview": "accepted | overridden — <why> | none (no C4 docs) | not returned",
  "ledger": ["<claim> — proved by <command>, observed <result>"]
}
```

**Merge, never replace.** The epic already carries grooming's keys and
pre-flight writes the baseline into the same object: read it, update your keys,
write it back.

`admitted` / `excluded` / `moved` are the admission record. A run that finished
its admitted set while beads sat out is done, and must say so rather than
reading as complete over a scope nobody remembers narrowing.

## Environmental ledger

Facts established this run that no wave should re-diagnose. Admitted on
evidence, not on reasoning: every entry — including a claim of sufficiency
("this check verifies X"), not only a claim of impossibility — names the
command that proved it and the result observed; run it first, record it
after. Note the conditions the command ran under when they could change the
result (attached repos, credentials, cache warmth, toolchain version). A
claim that cannot be tested cheaply is recorded as `unverified` with what
would settle it, not omitted. Keep the ORDER — dispatch prompts cite entries
by position ("ledger entry 1", "ledger entry 2"), so append rather than
reorder.

1. `<claim>` — proved by `<command>`, observed `<result>`
1. `<claim>` — unverified; would be settled by `<command>`
```

A gate recorded `fail` in the baseline is the bar for every bead in this run:
"no findings beyond the baseline", not "exit 0". A gate with no recorded
baseline is expected to exit 0.

## Wave table

**The plan names beads, and records the epic they came from.** The queue is
re-derivable from that one id — `bd ready --parent <epic-id> --exclude-type=epic`
— which is why the table is returned rather than stored: a stored copy can only
drift out of agreement with the graph it describes. Beads discovered mid-run
enter through Phase 1.5 step 4's scope test, and are recorded as an amendment.

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

Amendments are **`amend:`-prefixed comments on the run epic**. WHEN the run
diverges mid-flight — a wave resequenced, a bucket reclassified, a bead
admitted or bounced, a tier escalated — `bd comment <epic-id> "amend: …"`.

Append-only and timestamped by construction, so there is no read-modify-write
to lose and no ordering to maintain. There is no second copy to keep in step:
the wave table is re-derived, not stored, and the tracker is the only record.

```
bd comment <epic-id> "amend: wave 5 (demo-9kb.6, .7, .9) re-planned from one
shared batch to strictly sequential — .7's change to the shared fixture
invalidates .9's expected bytes, so they cannot share a checks run."

bd comment <epic-id> "amend: demo-9kb.8 escalated standard -> advanced after
two failed checks cycles."
```

Read them back with `bd show <epic-id>`, or
`bd export | jq -r 'select(.id=="<epic-id>") | .comments[].text' | grep '^amend:'`.
