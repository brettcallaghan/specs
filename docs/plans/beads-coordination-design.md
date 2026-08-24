# Coordination on the beads graph

Tests one hypothesis: **beads is already the orchestrator, and the sprint
skills are a second orchestrator built on top of it.** If that is true, most of
what the sprint system carries is not design — it is duplication.

Working document. Every capability claim below names the command run and the
output observed, per this repository's evidence rules. Claims taken from the
beads documentation rather than measured are marked **[doc]**; claims neither
measured nor documented are marked **[unverified]** with what would settle
them.

Measured against **bd 1.2.2 (dev)**, 2026-08-24, in a scratch database at
`scratchpad/bdlab` (prefix `lab`) so nothing touched this project's tracker.

---

## What this costs today

| | bytes |
|---|---|
| Ten sprint skills' bodies + their references | 210,232 |
| `bead-bonanza/scripts/` | 98,312 |

`bead-bonanza` alone is 96,911 bytes across a SKILL.md and twelve reference
files. The question is how much of that is coordination logic that `bd`
performs natively.

---

## 1 · The core loop works, and it is three lines

**Measured.** A root epic, two independent children, and a third bead depending
on both:

```
bd ready --parent lab-6af --json | jq -r '.[].id'
→ lab-6af.1
  lab-6af.2          # lab-6af.4 correctly absent: blocked by both

bd close lab-6af.1 lab-6af.2
bd ready --parent lab-6af --json | jq -r '.[].id'
→ lab-6af.4          # fan-in released automatically, no coordinator involved
```

Nothing computed that. Closing the blockers made the dependent claimable, which
is the entire wave mechanism the sprint system implements in prose.

So the coordinator's inner loop reduces to:

```
until bd ready --parent {EPIC} is empty and nothing is in_progress:
    dispatch what is ready, up to the concurrency cap
```

**`bd epic close-eligible`** ("Close epics where all children are complete") is
the termination signal, and it is a command rather than a judgement.

### Per-agent work queues are native

**Measured.** With `lab-6af.1` assigned to `dev-1` and `lab-6af.2` to `dev-2`:

```
bd ready --parent lab-6af --assignee dev-1 → lab-6af.1
bd ready --parent lab-6af --assignee dev-2 → lab-6af.2
```

That is the subagent loop you described, with no new machinery: a worker polls
its own queue and stops when it comes back empty.

---

## 2 · The constraint that shapes everything: claim XOR assign

**Measured, and it is a hard error rather than a preference:**

```
bd ready --parent lab-6af --assignee dev-1 --claim --json
→ {"error": "--claim cannot be combined with --assignee"}
```

**Measured:** `bd ready --claim` sets `status=in_progress` **and overwrites
`assignee` with the claiming actor** (observed: assignee became `Claude`, from
`$BEADS_ACTOR` / git `user.name`).

So beads offers two coordination models and will not let you blend them in one
call:

| | **Push** — coordinator assigns | **Pull** — workers self-select |
|---|---|---|
| who decides | the coordinator, at plan time | the worker, at claim time |
| worker loop | `bd ready --assignee me` → `bd update <id> --claim` **[doc]** | `bd ready --claim --json` |
| atomicity | two steps; exclusivity comes from the assignment | one step, first claim wins **[doc]** |
| suits | work that must go to a *particular* agent (a surface it holds context on) | a homogeneous pool of interchangeable work |

**Recommendation: push.** The sprint system's whole reason for batching is that
some beads must go to *the same* agent — one holding a surface's context, in
one worktree, committing serially. Pull cannot express "these five belong
together"; push can, and the two-step claim is not a real cost because the
assignment already guarantees exclusivity.

**Set `BEADS_ACTOR` to the agent id** in every dispatched subagent. Otherwise a
claim silently rewrites your assignment to the harness default and the
coordinator's view of who holds what decays. This is the answer to "what should
we use for agent IDs": one string, used as both the assignee and `BEADS_ACTOR`.

---

## 3 · Waves: delete them

**Waves are not a graph concept, and the graph makes them redundant.**

`bd ready` already returns the *complete* ready front. A wave is a decision to
freeze that front, dispatch it, wait for all of it, and re-query — which is
strictly worse than dispatching continuously from the re-queried front, because
the slowest bead in a wave holds up every bead that became ready behind it.

The sprint system's own supervision doc already concedes this shape: it
re-queries `bd ready` per wave and corrects the planned wave table against it.
The wave table is a prediction that the ready front then overrides.

**What waves genuinely carry that the graph does not: concurrency policy.** Two
beads that both change the same C4 component must not run at once even though
neither blocks the other. That is real, and it is the one thing `bd ready`
cannot know.

**But it is expressible as an edge.** "A and B collide, so serialize them" is
`bd dep add B A`. The wave packer already computes exactly this from
`c4_changes`; it just spends the answer on a table instead of writing it into
the graph.

So: **compile collisions into dependency edges at plan time, and delete the
wave.** The dispatch loop becomes a concurrency cap over the ready front, which
is one number rather than a planning artifact.

> **Should each wave be a swarm, with its own epic?** No. That re-adds the
> structure the graph already implies, and buys a progress view
> (`bd swarm status`) obtainable from `bd ready --parent` plus
> `bd epic status`. Nested epics also cost real accuracy — this repository has
> already measured that `bd list --parent` is direct-children-only and misses
> every leaf of a nested run epic.

---

## 4 · Batches: an assignment, not an epic and not a dep chain

You named the problem exactly: chaining a batch with dependency edges
**conflates sequencing with dependency**, and the two are different facts. Bead
B does not depend on bead A merely because one agent will do both.

The honest encoding is the one that carries no false claims:

- **A batch is a set of beads sharing an assignee.** Nothing more.
- Work happens sequentially because a subagent cannot spawn subagents — a
  property of the harness, not of the graph, so the graph should not assert it.
- Order within the batch is whatever `bd ready --assignee` returns
  (priority-sorted by default; `--sort` offers `priority|hybrid|oldest`).
  Order is free to vary precisely *because* batch members carry no real
  dependency on each other.

This also makes a batch queryable rather than declared. "Is `dev-1` finished?"
is `bd ready --assignee dev-1` returning empty with nothing `in_progress` — no
batch record, no return-line parsing, no dispatch-state metadata. Your question
"can bead assignments be queried in a monitor scenario" answers itself: that is
the only state there is.

**Where a batch genuinely IS ordered** — bead B builds on bead A's change — that
is a true dependency and deserves a real edge. The distinction stays visible
instead of being flattened into "everything in a batch is chained".

---

## 5 · Human decisions belong in the graph, and beads already models them

**Measured, both directions:**

```
bd gate create --type=human --blocks lab-6af.8 --reason "scope decision"
→ ✓ Created gate lab-r9e (type: human)
bd ready --parent lab-6af   → lab-6af.8 absent

bd gate resolve lab-r9e
bd ready --parent lab-6af   → lab-6af.8 present
```

A human decision is a gate. The bead leaves the ready front until a person
resolves it, and returns on resolution — with no coordinator branch, no
`decision-required` label, and no polling loop.

**Gate types, from `bd gate create --help`:** `human`, `timer` (auto-resolves
after `--timeout`), `gh:run` (waits for a GitHub Actions workflow), `gh:pr`
(waits for a PR merge). Cross-rig `bead` gates exist too, with `await_id`
formatted `<rig>:<bead-id>`. **[doc]** — of these only `human` was measured.

**`gh:run` and `gh:pr` are the ones you did not ask about and probably want
most.** A run that must wait for CI, or for a PR to merge, currently has no
representation at all; the coordinator polls or the user waits. As a gate it
becomes part of the graph, and `bd ready --gated` (measured: reports molecules
ready for gate-resume dispatch) tells the coordinator exactly when to move.

`decision` is also a first-class issue type (`bd types`: "Architecture decision
record (ADR)"), so a decision that deserves a durable record can be a bead
rather than only a gate.

---

## 6 · Merging

**The merge slot is already a queue, not just a mutex.** Measured:

```
BEADS_ACTOR=dev-1 bd merge-slot acquire       → ✓ Acquired. Holder: dev-1
BEADS_ACTOR=dev-2 bd merge-slot acquire       → Error: slot held by: dev-1
                                                 Hint: use --wait
BEADS_ACTOR=dev-2 bd merge-slot acquire --wait
                → ○ Slot held by dev-1, added to waiters queue (position 1)

bd show lab-merge-slot --json | jq '.[0].metadata'
→ {"holder": "dev-1", "waiters": ["dev-2"]}
```

`--wait` **registers and returns** rather than blocking, so a waiter polls its
position instead of holding a process open.

**Merge beads work, and they are the documented fan-in** **[doc]**:

```
bd dep add bd-merge bd-epic.1
bd dep add bd-merge bd-epic.2
```

A merge bead assigned to a merge agent, depending on every bead whose branch it
must land, becomes ready exactly when its inputs are done. That answers "could
we have merge beads assigned to the merge agent" — yes, and it is the shape the
beads docs themselves recommend.

**On a non-default base branch:** the slot is a mutex over a *name*, carrying no
notion of a branch at all, so the base is whatever the merging agent has checked
out. Merging into this run's feature branch needs nothing from beads.

**Caution — `merge-request` is not a usable type on 1.2.2.** `bd ready --type`'s
help advertises `merge-request` with an `mr` alias, but creation rejects it:

```
bd create --title "…" --type merge-request
→ Error: validation failed: invalid issue type: merge-request
```

Use `task`, or add one via `bd config set types.custom` **[unverified]** —
settled by setting it and creating a bead of that type.

---

## 7 · Ceremony as formulas and wisps

**Formulas** are TOML/JSON workflow templates in `.beads/formulas/`, compiled by
`bd cook` into protos, then instantiated by `bd mol pour` (persistent) or
`bd mol wisp` (ephemeral). **Measured:** this repository has *no* formulas —
`bd formula list` reports none across both search paths. The whole subsystem is
unused here.

This is the natural home for ceremony. Pre-flight, post-flight, and the close
sequence are a fixed DAG of steps instantiated per run — exactly what a proto
is. The ceremony stops being prose an agent must follow and becomes beads that
must close.

**Wisps** are ephemeral beads, `dolt_ignore`d so they never reach the synced
history; `bd promote` makes one permanent (preserving id, labels, deps, events,
comments) and `bd mol burn` discards a molecule. That is the right phase for
per-run ceremony: it should not accumulate in a tracker shared across every
future run, but a step that turned out to matter can be promoted rather than
retyped.

**Your pre-flight question — "if a pre-flight check fails do we create beads to
recover?"** — is where this pays off. A failed check becomes a bead blocking the
run epic; the run cannot proceed because the graph says so, not because a script
returned non-zero and a skill remembered to stop.

**`bd mol bond`** composes them, with bond types `sequential` (default),
`parallel`, and `conditional` ("B runs only if A fails"). `conditional` is a
genuine find: recovery paths become part of the template instead of prose.

---

## 8 · Wakeups: there is no push, and we already built the only substitute

**Measured.** A grep of bd's entire command surface for
`watch|notify|subscribe|poll` returns nothing. There is no mechanism to wake a
subagent when a bead becomes ready.

What exists is what this repository already uses: `bd` runs `.beads/hooks/on_update`
on every status change, and the harness tails that log with a Monitor. That is
the wakeup, and `progress-streaming.md` already specifies it.

So the honest answer to "can we wake up subagents when beads become ready" is
**no, and you already have the best available approximation.** Workers poll
their own queue; the coordinator watches the event stream. Design for polling.

---

## 9 · Harness integration — partly set up, and stale

**Measured:**

```
bd setup claude --check
→ ✓ Project hooks installed: .claude/settings.json
  ⚠ Claude Code integration installed but stale: CLAUDE.md
    Run: bd setup claude (to update)

bd hooks list
→ ✓ pre-commit, post-merge, pre-push, post-checkout, prepare-commit-msg
     (all installed, shim 1.1.2)
```

So the git hooks are in place and the Claude Code project hooks are wired, but
the CLAUDE.md guidance bd installs has drifted from what bd now ships. Worth
running and reviewing the diff.

**There is also a dedicated beads Claude Code plugin** you did not mention:
`/plugin marketplace add gastownhall/beads` then `/plugin install beads`. It
ships `/beads:*` slash commands, session lifecycle hooks, and — most relevant —
**`@task-agent`, "an autonomous agent that finds and completes ready tasks."**
That is this design's worker, already written. Worth evaluating before we write
our own. **[doc]**, not installed or tested here.

`prepare-commit-msg` adding "agent identity trailers for forensics" is also
directly useful for attributing a swarm's commits.

---

## 10 · What you did not ask about

- **`bd batch`** — "run multiple write operations in a single database
  transaction". Building one large graph is this design's central act, and doing
  it as N `bd create` calls is both slow and non-atomic.
- **`bd create` accepts markdown or graph JSON** (`bd create --help`: "or batch
  from markdown/graph JSON"). Planning could emit a graph document and load it
  in one call. **[unverified]** — settled by writing a graph JSON file and
  loading it.
- **Cross-repo dependencies** — `bd dep add bd-42 external:backend:api-ready`
  **[doc]**, plus `bd ship` ("publish a capability for cross-project
  dependencies") and `bd federation`. A run blocked on another repository's work
  becomes representable.
- **`bd query`** (a query language), **`bd count`**, **`bd graph`** — richer
  coordinator views than the `bd list`/`bd show` pair the skills use.
- **`bd kv`** — a key-value store. The run manifest currently lives in epic
  metadata, which forces read-modify-write merges; a kv namespace may be the
  better home. **[unverified]**.
- **`bd audit`** — append-only `.beads/interactions.jsonl` for "why did the
  agent do that", explicitly intended to be versioned and usable for
  fine-tuning datasets. A swarm generates exactly this data and currently
  discards it.
- **`bd set-state` / `bd state`** — operational state dimensions beyond
  `status`, as events plus labels. Possibly a cleaner home for `dispatch_state`
  than metadata keys. **[unverified]**.
- **`--readonly` mode** — "block write operations (for worker sandboxes)". A
  reviewing subagent can be handed a tracker it cannot mutate.
- **`bd ready --has-metadata-key` / `--metadata-field k=v` / `--label-any`** —
  the ready front is filterable enough to carve worker pools without assignments.

---

## 11 · What this would cut

Against the 210 KB measured above, the following exist to do things `bd` does:

| what | why it goes |
|---|---|
| the wave table, its producer, its overrides, per-wave re-query | the ready front is the wave |
| batch return-line parsing and verification | `bd ready --assignee X` empty is the return |
| `dispatch_state` / `dispatched_at` / `worktree` metadata bookkeeping | `status` + `assignee` already carry it |
| the acceptance-signal machinery for "did the subagent really close it" | closure is a tracker fact; testimony was never needed |
| `decision-required` and the human-escalation prose | gates |
| the merge-queue prose | merge beads + merge slot |
| much of `recovery.md` | there is no in-flight state outside the graph to reconstruct |

`sprint-wave` likely disappears entirely. `sprint-batch` becomes a worker loop
of a few lines. `sprint-plan` grows, because **planning becomes the whole job** —
which is your thesis, and it is the right place to spend.

**Estimate, not measurement: 40–60% of the 210 KB.** Settled only by doing it.

---

## 12 · Risks, honestly

- **Planning becomes a single point of failure.** Today a mediocre graph is
  rescued by a coordinator that re-reads and re-plans. Loop-on-the-graph has no
  such rescue: a missing edge is a race, a wrong edge is a stall. The
  Definition of Ready stops being paperwork and becomes load-bearing.
- **Claim-versus-assign must be chosen once**, and mixing them is a measured
  error, not a style question.
- **`--claim` rewrites `assignee`.** Any worker that claims without
  `BEADS_ACTOR` set corrupts the coordinator's ownership view silently.
- **JSON shapes are inconsistent.** Measured: `bd create --json` returns an
  **object**; `bd show --json` and `bd ready --json` return **arrays**. This is
  the same defect class as the `.metadata` array bug this repository's audit
  already found, and it will bite again.
- **Polling costs turns.** No push means the coordinator wakes on its own
  schedule; a tight loop burns context, a loose one adds latency.
- **Concurrency policy has no native representation.** Compiling C4 collisions
  into edges works, but the graph then encodes a *policy* as a *dependency* —
  the same conflation this document rejects for batches. It is the one place
  the compromise is worth it, and it should be labelled so a reader can tell
  the two apart.
- **Version risk.** Molecules, formulas, wisps and gates are the newest
  surfaces, this project uses none of them today, and `merge-request` shows the
  surface is not uniformly finished.

---

## 13 · What to settle next

1. Load a real graph via `bd create` from graph JSON, and via `bd batch` — the
   planning step's throughput depends on it.
2. Install and read the beads Claude Code plugin's `@task-agent`. If it is the
   worker loop, `sprint-batch` may not need to exist.
3. Test `gh:run` and `gh:pr` gates against a real PR — they are the highest-value
   unmeasured claim here.
4. Run `bd setup claude` and review the diff against the stale CLAUDE.md block.
5. Prototype: one root epic, collisions compiled to edges, three assigned
   workers, a merge bead, a human gate. Measure against a real feature.
