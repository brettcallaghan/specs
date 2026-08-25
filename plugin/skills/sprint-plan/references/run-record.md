# Run Record Format

**The run's record is its epic**, not a file: the manifest is `--metadata` JSON
on it and the amendments are `amend:`-prefixed comments on it. There is no
`dispatch.md` and no run directory — the tracker outlives a reclaimed
container, and a `/tmp` path does not.

Read this when you produce the edge file, and again whenever the coordinator
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
- The edge file — the one artifact a plan produces, its measured
  all-or-nothing loading, and the validators that run before it is handed over
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
  "target": "<the flat target epic; its task children are the work>",
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
  "collision_edges": "<path to the edge file this skill emitted>",
  "findings_path": "<sprint-retro's run record — the run's one document>",
  "retro_filed": 0,
  "retro_declined": 0,
  "retro_routed": 0,
  "ledger": ["<claim> — proved by <command>, observed <result>"]
}
```

**The four retro keys are `sprint-retro`'s, written at the end of the run.**
They are what lets a session that did not hold the run find its record from the
epic id alone. **An absent `findings_path` means the retro did not run** — a
different result from a retro that found nothing, and one a reader must not
have to guess at.

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
the ceremony roles (groom, plan, review, retro) are serialized by the graph,
never run alongside one another, and so share the one remaining slot between
them. Counting each ceremony role against the cap under-subscribes the
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

**Both ends of a collision edge ALREADY EXIST, so the edge goes in the SIDECAR
and never in the graph plan.** Admission created those beads and this skill
assigns them; a graph-plan edge resolves only against keys inside the plan, so
one naming an admitted bead aborts the whole load and creates nothing. "The
edge file" below is the mechanism; this section is only what the edge means.

**A collision edge is a POLICY expressed as a dependency**, and that conflation
is the price of putting it in the graph. Pay it explicitly: **both ends carry
`metadata.collision` naming the shared surface**, so a reader can tell a policy
edge from a real one. The mark cannot ride on the edge, in either artifact: a
line carries only `from` / `to` / `type`, and an unknown field on a
plan **edge** object — like one on a **node** or on the **plan** object itself
— is dropped with a warning and exit 0, so a mark placed there reaches the
tracker as nothing at all.

**Write the mark with `bd update`, on the bead itself**, in the same pass that
writes assignees. Both are writes against a bead that already exists, and
neither can come from the plan. Measured on bd 1.2.2:

```sh
bd update <id> --set-metadata collision=c4_changes:auth-service
# bd show <id> --json | jq -c '.[0].metadata'
#   -> {"collision":"c4_changes:auth-service"}
```

**Write the value BARE.** A JSON-quoted value stores the quotes literally —
measured, `--set-metadata collision='"c4_changes:auth-service"'` exits 0 and
`bd show --json` then returns `{"collision":"\"c4_changes:auth-service\""}`.
That matches no other bead's mark, so it defeats validator 3 while looking
correct in every listing.

Where a collision endpoint IS a node this plan creates, the mark rides in its
`metadata` like any other key:

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

### Name which class each edge is

The two classes have different consequences, so they must not read alike.
Both are lines in the edge file — the class
changes the mark, never which file the edge goes in:

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

## The edge file — the only artifact a plan produces

**You create no beads.** Every bead a sprint dispatches already exists when the
plan is written: `sprint-scope` admitted them under the target and built the
ceremony epics and the target itself. So a plan contributes exactly one thing —
`blocks` edges between beads that already exist — and assignees and collision
marks, which are `bd update` against those same beads.

Write it beside the run:

```
.beads/plans/<run-id>.edges.jsonl
```

One newline-delimited JSON object per line:

```json
{"from": "<the bead that WAITS>", "to": "<the bead it waits for>", "type": "blocks"}
```

**`from` is the WAITER**, and getting it backwards inverts a dependency at exit
0 with no warning — which is why the front is asserted after loading rather
than the file being trusted. Measured on bd 1.2.2: after
`{"from":"scr3-x88","to":"scr3-9c4","type":"blocks"}` loaded, `bd ready`
returned `scr3-9c4`, and `bd blocked` reported `scr3-x88: Blocked by 1 open
dependencies: [scr3-9c4]`.

Load it with `bd dep add --file <path>`. Per `bd dep add --help`, each line
must be an object with `from` and `to` and may include `type`; the aliases
`issue_id` / `depends_on_id` are accepted, and `--file -` reads stdin. Measured
on a one-line file: `✓ Added 1 dependencies`, exit 0.

**Validation is all-or-nothing, and that is what makes the file reviewable
before it loads.** A file whose second line names a nonexistent id wires
nothing — not even the valid first line:

```
$ bd dep add --file bad.jsonl ; echo $?
Error: bulk dependency validation failed:
  line 2: resolving dependency ID scr3-nope: resolving issue ID scr3-nope: no issue found matching "scr3-nope"
1
$ bd dep list scr3-x88
scr3-x88 has no dependencies
```

So a malformed file leaves no half-wired graph.

**Replay is safe.** Re-loading an identical file exits 0 and prints
`✓ Added 1 dependencies` again, with no duplicate-edge error, so a resumed run
may re-run it rather than reason about what already landed. The neighbouring
trap is not: `bd dep add` refuses to UPGRADE an existing edge's type, so a file
that would turn a `related` edge into a `blocks` fails rather than quietly
leaving the weaker one in place.

### Two validators, before the file is handed over

**1. Shape.** Every line is an object carrying `from`, `to` and optionally
`type`, and nothing else. Name the offending key and line in the failure.

**2. Every collision edge is marked — checked against the FILE.** For each line
you emitted as a collision, both endpoint beads must carry
`metadata.collision` with the same value, and no bead may carry the mark
without a line to justify it. **Read the marks back from the tracker**
(`bd show <id> --json | jq -c '.[0].metadata'`), not from the file: a mark is
written with `bd update --set-metadata` and does not appear in it. Unmarked, a
policy edge is indistinguishable from a real dependency the moment the file is
out of view — which is every reader after the run starts.

### Then assert the front

After the coordinator loads the file, `bd ready --parent <epic>
--exclude-type=epic` must contain neither end of any collision pair together.
Assert that rather than trusting the orientation: an inverted edge produces a
plausible-looking front and no error anywhere.

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
