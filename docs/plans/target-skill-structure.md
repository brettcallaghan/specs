# Target state — the skill structure after the split

What `bead-bonanza` becomes once wave 6 finishes, expressed as the three-tier
team it already implies: **orchestrator → worker → bead.**

Transient. Delete when wave 6 lands.

---

## The shape in one screen

```
RUN TIER — the orchestrator. Owns the run. Never touches code.

  bead-bonanza ........... the conductor. Owns the loop and nothing else.
    ├─ start-sprint ...... ✅ open:  preflight, admission, readiness gate
    ├─ plan-sprint ....... ✅ pack:  the ready front → buckets
    ├─ run-wave .......... ⬜ NEW    dispatch a wave, verify every return
    ├─ merge-queue ....... ⬜ NEW    serialize integration behind the slot
    ├─ close-sprint ...... ⬜ sixr   report, retro, teardown
    └─ recover-sprint .... ✅ resume: re-derive the run from the tracker

WORKER TIER — a subagent. Owns a bucket and one worktree. Never merges.

  run-batch .............. ⬜ NEW    the worker's own open/work/close ceremony

BEAD TIER — one unit of work, claimed to committed.

  run-bead ............... ✅ exists

INVOKED, NOT A TIER
  groom-beads ............ ✅ called by start-sprint when a bead is not Ready
  retro-skill/retro-spec . ✅ called by close-sprint
```

## Why it reads as a team

Every tier has the same three beats, and that is the whole design:

| tier | **open** | **work** | **close** |
|---|---|---|---|
| **Run** | `start-sprint` — environment probed, run epic resolved, queue gated | waves, until the front is dry | `close-sprint` — report, retro, teardown |
| **Wave** | `plan-sprint` — pack the current ready front | dispatch N workers concurrently | `merge-queue` — drain, one at a time |
| **Worker** | `run-batch` — take the worktree, write the dispatch record | `run-bead` × N | return the record, hand the branch up |
| **Bead** | claim, read it in full | implement, targeted tests, checks | close + commit in one, `retro:` comments |

Three properties make it hold together:

**Each tier's state is a durable artifact, not a message.** The run is an epic
plus its manifest; the worker is a `dispatch_id` in bead metadata; the bead is
its own status and commit. Every one survives the container, so any tier can be
resumed by something that holds no context.

**Scope flows down, records flow up.** A tier is handed exactly what it needs
and returns exactly what happened. Nothing reaches past its neighbour: the
orchestrator never opens a source file, a worker never merges, a bead never
looks outside itself.

**No tier trusts a report it could verify.** `bb-status.sh` checks each tier's
acceptance from state — `run` for the run, `wave` for the wave, `bead` for the
bead, `slot` for the merge queue. Testimony is a convenience; state is the
verdict.

## The two new skills, and what is genuinely new about them

### `run-wave` — extraction, mostly

Phase 3 as it stands (14.6 KB): pick the dispatch mode, dispatch, verify skip
claims with a diff-stat, run the acceptance check on every return, verify the
tree is clean, trip the circuit breaker at 3 consecutive failures.

It becomes a skill because it runs **once per wave** and the orchestrator's
loop is then four lines: pack, dispatch, merge, re-query. What it must gain on
extraction is the re-query — a wave is the *current* ready front, not a
position in a stored list.

### `merge-queue` — a real design change, not an extraction

**Today the orchestrator merges inline**, and the merge slot exists only to
make that visible: "only the orchestrator merges, so within one run the
serialization already holds by construction."

Moving integration to its own worker changes two things:

- **Conflict resolution stops burning the orchestrator's context.** A merge
  that needs real work gets a fresh worker with the two branches and nothing
  else; the orchestrator keeps its budget for the run.
- **The slot stops being advisory.** With a merge worker it is what actually
  serializes integration, which is what `bd merge-slot` was built for.

**One merge worker at a time, strictly.** Concurrency here is the failure mode
bd's own documentation names — several agents racing to resolve conflicts and
cascading them. The queue is serial by construction; the parallelism stays in
the wave.

`bb-status.sh slot <epic>` already tells the orchestrator whether a held slot
is safe to reclaim, so the supervision for this exists before the tier does.

### `run-batch` — the missing middle

The worker tier has prompt templates and no skill. That is the same gap the
dispatch record closed at the data layer: a subagent had no id, no record, no
lifecycle. `run-batch` gives it a ceremony to match —

1. **Open** — write the dispatch record (`dispatch_id`, model, tier, worktree,
   time), take the worktree, read the manifest and the checks baseline.
2. **Work** — `run-bead` per bead, in order, one commit each.
3. **Close** — write the terminal state, hand the branch to the merge queue,
   return one line per bead.

The batch templates in `subagent-prompts.md` collapse into it, and the
three-case worktree prose disappears because the record answers it.

## What this costs, in bytes

The epic's criterion is a body under 20,000. Current: **54,425**.

| removed by | section | bytes |
|---|---|---|
| `run-wave` | Phase 3 | −14,621 |
| a Phase 2.5 bead | Triage & dispatch plan | −10,606 |
| `close-sprint` (`sixr`) | Phase 4 | −8,268 |
| `hx4t` | compaction defences, reordering | −2,000 est. |
| | pointers added back | +2,800 est. |
| | **projected** | **≈ 21,700** |

So the target is reachable, and **only** if the two unfiled beads exist —
`run-wave` and the Phase 2.5 extraction. `sixr` and `hx4t` alone land near
40,000. Trimming the last ~1,700 is `hx4t`'s reordering pass.

Note separately: `groom-beads/SKILL.md` (23,071) also trips `--strict`, and no
wave-6 bead owns it. The epic's "exits 0 with no warnings" criterion cannot be
met without a decision there.

## Sequencing

```
run-wave ─────┐
              ├─→ hx4t  (defences retire only once the body fits)
phase-2.5 ────┤
close-sprint ─┘        ← blocked on f8sa (wave 5 retro design)
merge-queue            ← independent; needs run-wave to hand off to
run-batch              ← independent; needs run-wave to dispatch it
```

`run-wave` is the keystone: it is the largest single cut, and both new tiers
attach to it.
