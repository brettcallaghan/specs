# Run Record Format

**The run's record is its epic**, not a file: the manifest is `--metadata` JSON
on it and the amendments are `amend:`-prefixed comments on it. There is no
`dispatch.md` and no run directory — the tracker outlives a reclaimed
container, and a `/tmp` path does not.

Read this when you produce the graph plan, and again whenever the orchestrator
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
- Collision edges — a concurrency policy expressed as a dependency, how it is
  marked so a reader can tell it from a real one, and why a co-edit collision
  batches rather than parallelizes
- Reading C4 overlap — compute co-edit and sequencing edges from bead metadata,
  not from file paths
- The graph plan — the `bd create --graph` document the plan emits, its
  measured schema, the three validators that run before it is loaded, and a
  worked example
- Amendments — `amend:` comments on the epic, append-only by construction

Every run gets one, whatever its size. A two-bead run interrupted at bead one
needs its record exactly as much as a twenty-bead run does.

## Run manifest

`--metadata` JSON on the run epic, **written by the orchestrator itself** from
`scripts/pre-flight-check.sh`'s output plus what admission decided. Retrieve it
whole with `bd show <epic-id> --json | jq '.[0].metadata'`.

**`branch` and `base_sha` are the orchestrator's, written once.** It creates
the sprint branch immediately before its first dispatch — after grooming, so
`base_sha` includes the grooming commits — and writes both keys then. A run
whose manifest already carries them **reuses** the branch and creates nothing;
that is what makes a resumed run land where the original did. Three things read
them and none may write them: the review scopes its diff `base_sha`..sprint
head, every worktree branches from that branch's HEAD, and the pull request
opens against it.

```json
{
  "target": "<the flat target epic; what `bd swarm` is given>",
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
  "partition_preview": "accepted | overridden — <why> | none (no C4 docs) | not returned",
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

Facts established this run that no dispatch should re-diagnose. Admitted on
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

## Collision edges

**The partition is edges in the graph, not a schedule.** Two beads that must
not run at once get a `blocks` edge between them, so they cannot both be in
`bd ready` and the coordinator can dispatch the front continuously without
consulting a plan. Beads discovered mid-run enter through the emergent-work
scope test and are recorded as an amendment; the graph absorbs them without a
re-plan, which a schedule could not.

**A collision edge is a POLICY expressed as a dependency**, and that conflation
is the price of putting it in the graph. Pay it explicitly: **both ends carry
`metadata.collision` naming the shared surface**, so a reader can tell a policy
edge from a real one. Nothing else in the plan can carry the mark — measured,
`bd create --graph` drops an unknown field on an **edge** object and an unknown
key on the **plan** object with a warning and exit 0, exactly as it does on a
node, so a label placed there reaches the tracker as nothing at all.

```json
{"key": "b", "title": "…", "type": "task", "parent_key": "target",
 "assignee": "dev-r7k-2", "metadata": {"c4_changes": "auth-service",
                                       "collision": "c4_changes:auth-service"}}
```

Two beads blocked apart by an edge and carrying the same `collision` value were
separated by policy. A blocked bead with no `collision` mark, or one whose mark
differs from its blocker's, has a real dependency.

**Chain a cluster, do not fan it.** N beads on one surface get N−1 edges in a
line, and the whole line takes ONE `assignee` — that is what a batch is now.
The assignee groups it; the edges order it; `bd ready --assignee dev-<id>`
returns its head. A fan (every bead blocked on one) would leave the rest ready
simultaneously and defeat the point.

**Order the chain by dependency first, then by priority.** Where two beads in a
cluster already have a real edge between them, the collision chain must not
contradict it — a cycle is refused, and an order that fights the real
dependency serializes the wrong way round.

### The merge bead

**One per batch, in the plan.** A task labelled `merge`, assigned to
`merge-<run-id>`, with a `blocks` edge from it to every bead in that batch:

```json
{"key": "m1", "title": "Merge batch 1", "type": "task", "parent_key": "target",
 "priority": 1, "assignee": "merge-<run-id>", "labels": ["merge"]}
```
```json
{"from_key": "m1", "to_key": "<each bead in the batch>", "type": "blocks"}
```

Remember the orientation: `from_key` **waits for** `to_key`, so the merge bead
is the `from_key` on every one of those edges.

Measured on a two-bead batch: `bd ready --assignee merge-<run-id>` returned
nothing while either input was open — including after only the first closed —
and returned the merge bead the moment the second closed. `bd ready --label
merge` answers the same, so nothing has to know the agent id to find it.

**A task plus a label, not a custom type.** `bd config set types.custom
merge-request` works and behaves as a task for segregation, but it lives in
LOCAL config, so it is a per-clone setup step and a clone that has not run it
cannot create or filter those beads. The coordinator's worktree-mechanics
reference carries that decision and its rationale in full; do not re-open it
here.

### Name which class each edge is

The two classes have different consequences, so they must not read alike:

- **Co-edit collision** — the shared surface: `c4_changes:auth-service`,
  `fixture:testdata/golden.json`. Marked with `metadata.collision`, removable
  without changing what any bead means.
- **Sequencing constraint** — the consumer waits for the producer, because it
  consumes a contract the producer is changing. It holds with zero file
  overlap, it is a real dependency, and it carries **no** `collision` mark.
- **Signal disagreement** — the file scan and the C4 tags disagree. Take the
  **union** as the collision set, mark it, and say so in the return: a
  disagreement usually means stale C4 tagging or a bead scoped wider than its
  tags say, so it is worth a line even when the packing is unaffected.

Tier is not an edge at all: it rides on the bead as
`metadata.execution_suggested_model` (else `metadata.tier`) and resolves to a
model at dispatch via the project's `sf.toml` `[models]` lineup. Where the
project declares no lineup the subagent inherits the session model.

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

A + B collide on `auth-service` even if they share no file: shared-surface, so
a **collision edge** chains them behind one assignee. C changes a different
component and shares no file with either, but consumes `auth-service` — a
sequencing constraint, so C gets a real `blocks` edge after both. D is isolated
on both signals and gets neither. Nothing in this reads a file path, which is
the point: the file scan would have parallelized A with B and seen no
relationship at all between C and them.

### Why a co-edit collision batches rather than parallelizes

**Worktrees isolate the agents but not the merge.** Two agents in two worktrees
never see each other's edits; then both branches come back to the orchestrator,
the second lands on a file the first already rewrote, and under `--ff-only` the
merge refuses. Batching removes that: the beads go behind one agent, in order,
each branching from the last, so what returns is already reconciled. Isolation
is never grounds to relax a batching rule — it prevents a collision that was
never the hazard.

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

**An edge carries nothing else** — which is why a collision edge's mark lives
in the waiting node's `metadata` and not on the edge.

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

**An unknown field is a warning, not an error, and the load still exits 0** —
at all three levels. Measured on one load: an unknown field on a **node**, on
an **edge**, and an unknown key on the **plan object itself** each printed
`graph plan … has unknown field(s): […] (silently dropped — see 'bd create
--graph' schema)` and the load reported `Created 3 issues`, exit 0. So a
mistyped `assignee` is a bead nobody owns, a mistyped edge key a dependency
that was never wired, and a sidecar list of collision edges reaches the tracker
as nothing at all. A *parse* error is the opposite — it aborts before creating
anything. The fail-open case is the one a validator has to cover.

### Three validators, before the plan is handed over

**1. Unknown fields.** Check every node and edge object — and the top-level
plan object — against the field tables above and **name the offending key** in
the failure. Do not rely on `bd`'s warning: it appears on a run that also
reports success, so nothing downstream treats it as a failure.

**2. Flatness of the swarmed target.** **No node whose `parent_key` is the
target epic may have `type: "epic"`.** A nested target makes `bd swarm
validate` count the child epic as one issue and never see the beads beneath
it — measured, a target with one child epic over two tasks reports `Total
issues: 1` and treats the *child epic* as the work item, while the same two
tasks parented directly report `Total issues: 2`. Both report `Swarmable: YES`.
Work that genuinely decomposes becomes a second target under the run epic,
swarmed separately.

**3. Every collision edge is marked.** For each edge you emitted as a
collision, both endpoints must carry `metadata.collision` with the same value,
and no node may carry the mark without an edge to justify it. Unmarked, the
policy is indistinguishable from a real dependency the moment the plan file is
out of view — which is every reader after the run starts.

### Worked example

Run epic as the container, ceremony epics and a flat target beneath it, work
beads assigned, dependencies and one collision as edges:

```json
{"nodes": [
  {"key": "run",       "title": "(EPIC) Run 3: the beads graph", "type": "epic", "priority": 1},
  {"key": "preflight", "title": "Pre-flight", "type": "epic", "parent_key": "run", "priority": 1},
  {"key": "target",    "title": "(TARGET) plan items 6-14", "type": "epic", "parent_key": "run", "priority": 1},
  {"key": "report",    "title": "Report", "type": "epic", "parent_key": "run", "priority": 2},
  {"key": "i6", "title": "Item 6: the graph plan", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-3", "priority": 1},
  {"key": "i7", "title": "Item 7: the run-epic shape", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-1", "priority": 1,
   "metadata": {"c4_changes": "bead-bonanza", "collision": "c4_changes:bead-bonanza"}},
  {"key": "i8", "title": "Item 8: collisions as edges", "type": "task", "parent_key": "target", "assignee": "dev-mpnf-1", "priority": 2,
   "metadata": {"c4_changes": "bead-bonanza", "collision": "c4_changes:bead-bonanza"}}],
 "edges": [
  {"from_key": "target", "to_key": "preflight", "type": "blocks"},
  {"from_key": "report", "to_key": "target",    "type": "blocks"},
  {"from_key": "i7",     "to_key": "i6",        "type": "blocks"},
  {"from_key": "i8",     "to_key": "i7",        "type": "blocks"}]}
```

`i7` and `i8` both edit `bead-bonanza`, so the last edge is a **collision**:
both ends carry the same `metadata.collision`, and both are assigned to
`dev-mpnf-1`, which is what makes them one batch worked in order. The
`i7 <- i6` edge above it carries no mark, so it is a real dependency.

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
diverges mid-flight — an edge added, a batch reclassified, a bead admitted or
bounced, a tier escalated — `bd comment <epic-id> "amend: …"`.

Append-only and timestamped by construction, so there is no read-modify-write
to lose and no ordering to maintain. There is no second copy to keep in step:
the graph is the plan, and the tracker is the only record.

```
bd comment <epic-id> "amend: collision edge added demo-9kb.9 <- demo-9kb.7 —
.7's change to the shared fixture invalidates .9's expected bytes, so they
cannot share a checks run."

bd comment <epic-id> "amend: demo-9kb.8 escalated standard -> advanced after
two failed checks cycles."
```

Read them back with `bd show <epic-id>`, or
`bd export | jq -r 'select(.id=="<epic-id>") | .comments[].text' | grep '^amend:'`.
