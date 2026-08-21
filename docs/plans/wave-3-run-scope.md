# Wave 3, reconsidered — a run has a selector, not an epic

Supersedes wave 3 as filed. Transient; delete when the change lands.

---

## What changed

The filed plan said **a run is an epic**. That forces a project to organise
its epics around runs, and it cannot express the skill's own advertised entry
points — "all beads with label X", "all P1 beads" — because those cut across a
hierarchy and a bead has exactly one parent.

The reframe: **a run has a *selector*, and the selector's shape decides which
`bd` capabilities are available.** Epics stop being a requirement and become
the case where the most is possible.

## The selector

Three shapes. All are **live** — re-evaluated on every restart, never
snapshotted:

1. **One epic** — the scope is the whole epic, including beads added to it
   after the run started.
2. **Several epics** — the union.
3. **A filter** — label, priority, an explicit id list.

Locking scope to an enumeration was the wrong fix. An epic selector already
locks scope, to something durable and inspectable; a filter selector is
*meant* to move, and a user adding work between restarts is a supported
intent, not a hazard.

**The guard replaces the old rule.** "Enumerate, never filter" becomes:
re-evaluate the selector, diff it against what the manifest recorded, and
**report the delta before dispatching**. A run that silently grew is the
failure; a run that grew and said so is the feature.

## What is constant: the run bead

Every run has exactly one bead carrying its manifest, and **that bead's id is
the run id** — greppable, durable, survives compaction, unlike today's
`$(date +%s)-<count>`.

- **Selector is one epic** → that epic *is* the run bead. Nothing is created.
- **Otherwise** → the skill creates a run bead. It does **not** own the queued
  beads. No reparenting, ever. It records the selector, the manifest and the
  amendments, and nothing else.

Manifest is `--metadata` JSON on it; amendments are `amend:`-prefixed comments,
append-only and timestamped by construction. `$RUN_DIR`, the run root, the
`current` symlink and `dispatch.md` retire in all three cases, and recovery
reads the tracker.

A created run bead is opened **`in_progress` and claimed**, so `bd ready` never
returns it (measured: `bd ready` excludes `in_progress`), and closed at run
end. "Is a run live here?" becomes `bd list --status in_progress --label bb-run`.

## Capability matrix

| | one epic | several epics | filter |
|---|---|---|---|
| queue | `bd ready --parent <e> --exclude-type=epic` | per epic, unioned | `bd ready --label …` |
| dependency waves | `bd swarm validate` | per epic, merged | `plan-sprint`'s own partition |
| live view | `bd swarm status` | per epic | `bd list --status in_progress` |
| run-level DoD | `bd epic status` | per epic | closed count against the re-evaluated selector |
| registration | `bd swarm create` | one per epic | — |
| manifest home | the epic | the run bead | the run bead |

`--exclude-type=epic` is not optional: `--parent` is transitive and returns the
intermediate epics themselves, so without it a run dispatches its own sub-epics
as work.

## Answering "can we adopt swarm whenever given epics?"

**Yes, with two constraints.** Measured on this tracker (bd 1.2.2):

- `bd swarm validate agent-plugins-bir9` returned the dependency-wave partition
  directly — Wave 1 (2), Wave 2 (5), Wave 3 (1), max parallelism, and a warning
  per cross-epic dependency edge.
- `bd swarm status agent-plugins-bir9` returned the live view, including the
  in-progress claim held at the time. Its help states the status is **computed
  from beads, not stored** — so it tracks a scope that changes, which is what a
  live selector needs.

Constraints:

- **`bd swarm create` takes exactly ONE epic.** N epics means N swarms; there
  is no multi-epic form.
- **Given a non-epic issue, `swarm create` auto-wraps it** — creates an epic
  with that issue as its only child. That is a reparent. Never point it at a
  bare bead.

## Re-scoping the eight filed beads

| bead | filed as | becomes |
|---|---|---|
| `7vej` | a run is an epic; the epic id is the run id and the queue | a run's scope is its selector; every run has a run bead whose id is the run id |
| `x5ix` | attach the manifest to the epic; retire `$RUN_DIR` | attach it to the **run bead** — identical for a single-epic selector, and now defined for the other two |
| `205x` | take dependency waves from `swarm validate` | unchanged, **conditional**: epic selectors only; a filter selector keeps `plan-sprint`'s partition |
| `ddvt` | register with `swarm create`, `swarm status` as the live view | unchanged, **conditional**: epic selectors only |
| `kyyl` | run-level DoD via `bd epic status` | unchanged, **conditional**: epic selectors; a filter selector counts against the re-evaluated selector |
| `ccm9` | multi-epic runs via epic-of-epics | **shrinks or closes.** Several epics is now native to the selector, so no wrapper epic is needed — and building one would reparent. What survives is "run the per-epic commands N times and merge", which is one line of `205x` and `ddvt` |
| `67xf` | sequencing constraints as real `blocks` edges at grooming | unchanged — independent of the selector |
| `we2q` | epic-level Definition of Ready in `groom-beads` | unchanged, **conditional**: applies when an epic is the selector |

## The honest cost

This gives up **uniformity**. Three of the eight beads become conditional on
selector shape, so the skill carries two paths where the filed plan carried
one. In exchange no project has to reorganise its epics to use the skill, and
nothing reparents a bead the user did not ask to move.

It gives up nothing that was measured to work.

## Open for decision

**`ccm9` — shrink or close?** Closing it says several-epic support is just
"loop the per-epic commands", which is true and is already in the matrix above.
Shrinking it keeps a bead for merging `swarm validate` output across epics when
the epics have dependency edges between them — which `swarm validate` reports
as a warning and nothing currently acts on. My read: **shrink it to the
cross-epic dependency merge**, which is real work, and drop the epic-of-epics
framing entirely.
