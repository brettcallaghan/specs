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
- `concurrency_cap` — the size of the harness's worker pool, and the
  N_dev ≤ cap − 1 rule the plan is checked against
- Wave table — returned to the orchestrator, never stored; names beads, and
  every row carries class, batch, tier and push mode, with its Reason naming
  which collision class applies
- Reading C4 overlap — compute co-edit and sequencing edges from bead metadata,
  not from file paths; and why a co-edit collision batches rather than
  parallelizes
- Trusting `bd swarm validate`'s preview — the two measured conditions that
  void its partition
- The graph plan — the `bd create --graph` document the plan emits, its
  measured schema, the two validators that run before it is loaded, and a
  worked example
- Amendments — `amend:` comments on the epic, append-only by construction

Every run gets one, whatever its size. A two-bead run interrupted at bead one
needs its record exactly as much as a twenty-bead run does.

## Run manifest

`--metadata` JSON on the run epic, **written by the orchestrator itself** from
`scripts/pre-flight-check.sh`'s output plus what admission decided. Retrieve it
whole with `bd show <epic-id> --json | jq '.[0].metadata'`.

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
  "checks_ids": ["format", "vet", "build", "test-go", "…"],
  "checks_baseline": "<verbatim text — inline this into EVERY subagent prompt>",
  "checks_baseline_status": 1,
  "concurrency_cap": 20,
  "models": {"light": "…", "standard": "…", "advanced": "…", "frontier": "…"},
  "effort": "low|medium|high|xhigh|unrecorded",
  "standing_auth": {"<row>": {"value": "<resolved>", "source": "sf.toml|default"}},
  "trailers": "<commit trailers, or absent when the run resolved none>",
  "admitted": ["<bead-id>", "…"],
  "excluded": [{"id": "<bead-id>", "reason": "declined move out of <epic-id>"}],
  "moved": [{"id": "<bead-id>", "from": "<previous parent>"}],
  "wave_preview": "accepted | overridden — <why> | none (no C4 docs) | not returned",
  "review_passes": 0,
  "review_budget": 2,
  "graph_plan": "<path to the graph plan this skill emitted>",
  "findings_path": "<sprint-retro's findings document>",
  "retro_filed": 0,
  "retro_declined": 0,
  "retro_outbox": 0,
  "report_path": "<sprint-report's report>",
  "ledger": ["<claim> — proved by <command>, observed <result>"]
}
```

**The four retro keys and `report_path` are written by their own skills, at the
end of the run, in that order.** They are why `sprint-report`'s only input is
the epic id: the retro runs first and leaves its outcome here, so the report
names what the retro produced without being handed it, and a session that did
not run the sprint can still write one. **An absent retro key means the retro
did not run** — which the report states rather than omits, since a skipped
retro and one that found nothing are different results.

**Merge, never replace.** The epic already carries grooming's keys and
pre-flight writes the baseline into the same object: read it, update your keys,
write it back.

**`review_passes` is durable for one reason:** a resumed run that restarted the
counter would loop forever. `review_budget` records what the authorization gate
decided, so a run resumed by a session that never met the human still honours
it.

**`standing_auth` records each gate row's SOURCE as well as its value.** A row
resolved from `sf.toml`'s `[run]` and a row that took the skill default are
different facts, and the record exists to stop a decision being re-litigated
mid-run — a project-level default is one more place one can hide from whoever
reads the report.

`admitted` / `excluded` / `moved` are the admission record. A run that finished
its admitted set while beads sat out is done, and must say so rather than
reading as complete over a scope nobody remembers narrowing.

## `concurrency_cap`

**The size of the harness's worker pool, not a dispatch rhythm.** Its value is
`CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` where the session sets it, and **20** —
Claude Code's documented default — where it does not. It is recorded rather
than assumed because it varies by harness and by session: the Workflow tool
caps at `min(16, CPUs − 2)`, which is 2 on a 4-CPU container, and a run resumed
on a differently-configured session has to know what the plan assumed. The
orchestrator resolves it once, with the rest of the manifest and before
grooming is dispatched — `${CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS:-20}`.

It bounds **simultaneous** agents, so the planning rule is **N_dev ≤ cap − 1**:
the ceremony roles (groom, plan, review, retro, report, merge) are serialized by
the graph, never run alongside one another, and so share the one remaining slot
between them. Counting each ceremony role against the cap under-subscribes the
dev pool by the number of roles; counting none of them oversubscribes it by one.

**Exceeding it does not fail loudly.** A spawn past the cap is refused, and the
beads assigned to a dev that was never spawned then sit ready, correctly
assigned and permanently unstarted — from the tracker's point of view nothing is
wrong. Breadth beyond N_dev is therefore expressed as more beads per dev, never
as more devs. Check it at plan time, counting dev assignees only:

```sh
bd list --parent <run-epic> --json \
  | jq -r '[.[].assignee] | map(select(. != null and . != "")) | unique | length'
# must be <= manifest.concurrency_cap - 1
```

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
| Wave | Bead | Class | Batch | Tier | Push | Reason |
|------|------|-------|--------|------|------|--------|
| 1 | demo-89s | trivial | batch #1 | light | defer | docs only |
| 1 | demo-aqw | trivial | batch #1 | light | defer | comment only |
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
  beads may share a batch; they may never share a concurrent wave.
- **Sequencing constraint** — names what is consumed and whose change it
  consumes: `consumes auth-service, changed by demo-1kq`. This is an *ordering*
  edge, not a batching one; it can hold with zero file overlap. The consumer
  goes in a later wave, or after the producer inside one batch. WHEN the
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

### Why a co-edit collision batches rather than parallelizes

**Worktrees isolate the agents but not the merge.** Two agents in two worktrees
never see each other's edits; then both branches come back to the orchestrator,
the second lands on a file the first already rewrote, and under `--ff-only` the
merge refuses and stops the wave. Batching removes that: the beads go behind one
agent, in order, each branching from the last, so what returns is already
reconciled. Isolation is never grounds to relax a batching rule — it prevents a
collision that was never the hazard.

## Trusting `bd swarm validate`'s preview

`bd swarm validate <epic>` gives a wave preview worth starting from, under two
measured conditions:

- **It scopes dependency analysis to the epic given.** An edge to a bead outside
  that epic is reported as a warning and then ignored — observed placing a
  blocked bead in Wave 1. If it warns about an outside dependency, do not use
  its partition.
- **It is one level deep.** On a nested run epic it lists sub-epics as work
  items and never sees their leaves. If the run epic is nested, do not use its
  partition.

Under either, derive packing from the re-queried ready front instead, and say in
the return which one was used.

## The graph plan

**One JSON document builds the run.** Instead of describing a shape for the
orchestrator to construct bead by bead, the plan emits the nodes and edges and
the orchestrator loads them with a single `bd create --graph <path>`. Write it
where the run can find it again — `.beads/plans/<run-id>.graph.json` — and
**return the path**; `graph_plan` is a manifest key and the manifest is the
orchestrator's to write, not yours.

**Being reviewable before it is loaded is most of the value.** Emit it, diff
it, validate it, then load it. A malformed plan fails partway through and
leaves half a graph behind.

### Schema

```json
{"nodes": [{"key": "i6", "title": "…", "type": "task",
            "parent_key": "target", "assignee": "dev-<run-id>-1",
            "priority": 1, "description": "…",
            "labels": ["human"], "metadata": {"tier": "advanced"}}],
 "edges": [{"from_key": "i7", "to_key": "i6", "type": "blocks"}]}
```

Every field below is measured against `bd create --graph`; nothing else is
accepted, and what is not accepted is **dropped rather than refused**.

| Node field | Meaning |
|---|---|
| `key` | the plan-local handle other nodes and edges reference. Not an issue id |
| `title` | the bead's title |
| `type` | `epic` or `task` |
| `parent_key` | another node's `key` **in this plan** — this is what builds the run's hierarchy |
| `parent_id` | an id that **already exists** in the tracker, for hanging a plan off a live epic |
| `assignee` | minted from the coordinator's agent roster, never spelled by hand |
| `priority` | integer, 0–4 |
| `description`, `labels`, `metadata` | as on `bd create`. `metadata` values are **strings**; a number or nested object is a parse error |

`acceptance`, `design`, `notes`, `estimate`, `external_ref`, `status`, `deps`
and `gate` are **not** node fields — a gate in particular cannot be created
from a plan, so anything needing one creates it against the loaded bead
afterwards.

| Edge field | Meaning |
|---|---|
| `from_key` | the node that **waits** |
| `to_key` | the node it waits for |
| `type` | `blocks` |

### Three traps, each of which fails silently

**`from_key` waits for `to_key` — the opposite of how `"type": "blocks"`
reads.** `{"from_key": "a", "to_key": "b", "type": "blocks"}` puts **b** in
`bd ready` and **a** in `bd blocked`. Getting it backwards inverts every
dependency in the run, exits 0, and produces a ready front that looks
plausible. Assert the front after loading rather than trusting the file.

**`parent_key` and `parent_id` are not interchangeable.** `parent_id` holding a
plan key is accepted with **no warning** and creates an orphan — the node lands
outside the run epic entirely, and every `--parent` query then under-reports the
queue. Use `parent_key` for nodes this plan creates.

**An unknown field is a warning, not an error, and the load still exits 0.**
`bd create --graph` prints `graph plan node["x"] has unknown field(s): […]
(silently dropped)` and carries on, so a mistyped `assignee` is a bead nobody
owns and a mistyped edge key is a dependency that was never wired. A *parse*
error is the opposite — it aborts before creating anything. So the fail-open
case is the one a validator has to cover.

### Two validators, before the plan is handed over

**1. Unknown fields.** Check every node and edge object against the field
tables above and **name the offending key** in the failure. Do not rely on
`bd`'s warning: it appears on a run that also reports success, so nothing
downstream treats it as a failure.

**2. Flatness of the swarmed target.** **No node whose `parent_key` is the
target epic may have `type: "epic"`.** A nested target makes `bd swarm
validate` count the child epic as one issue and never see the beads beneath
it — measured, a target with one child epic over two tasks reports `Total
issues: 1` and puts the *child epic* in Wave 1, while the same two tasks
parented directly report `Total issues: 2`. Both report `Swarmable: YES`. Work
that genuinely decomposes becomes a second target under the run epic, swarmed
separately.

### Worked example

Run epic as the container, ceremony epics and a flat target beneath it, work
beads assigned, dependencies as edges:

```json
{"nodes": [
  {"key": "run",       "title": "(EPIC) Run 3: the beads graph", "type": "epic", "priority": 1},
  {"key": "preflight", "title": "Pre-flight", "type": "epic", "parent_key": "run", "priority": 1},
  {"key": "target",    "title": "(TARGET) plan items 6-14", "type": "epic", "parent_key": "run", "priority": 1},
  {"key": "report",    "title": "Report", "type": "epic", "parent_key": "run", "priority": 2},
  {"key": "i6", "title": "Item 6: the graph plan", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-3", "priority": 1},
  {"key": "i7", "title": "Item 7: the run-epic shape", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-1", "priority": 1},
  {"key": "i8", "title": "Item 8: collisions as edges", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-2", "priority": 2}],
 "edges": [
  {"from_key": "target", "to_key": "preflight", "type": "blocks"},
  {"from_key": "report", "to_key": "target",    "type": "blocks"},
  {"from_key": "i7",     "to_key": "i6",        "type": "blocks"},
  {"from_key": "i8",     "to_key": "i6",        "type": "blocks"}]}
```

Loading that prints `Created 7 issues` and one `key -> id` line each. Assert
four things against the database, not against the file: the node count, every
edge (`bd dep list <id>` per node), an assignee on every work bead, and the
ready front — here `bd ready` returns the run epic and pre-flight, and `bd
blocked` returns the other five.

**A blocked parent blocks its children**, which is what makes one `epic→epic`
edge order a whole phase: with the target waiting on pre-flight, `bd blocked`
reports `i6` as blocked by the *target*, not by anything in its own row of the
plan. That is also why the queue is always
`bd ready --parent <epic> --exclude-type=epic` — the run epic answers `bd
ready` on its own account.

## Amendments

Amendments are **`amend:`-prefixed comments on the run epic**. WHEN the run
diverges mid-flight — a wave resequenced, a batch reclassified, a bead
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
