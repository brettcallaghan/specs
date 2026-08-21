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

## Do we move beads that already belong to another epic? No.

**Measured, and it decides the design.** A probe feature epic with two children
(one closed, one open) reported `◐ 1/2 complete (50%)`. Moving the open child
into a run epic changed it to:

```
✓ 1/1 complete (100%) — eligible for close
```

Restoring the parent put it back to 1/2. So a run that adopts a bead with a
home makes that home falsely report finished, and eligible for
`bd epic close-eligible`. A transient run must not be able to corrupt the
structure a project maintains.

**The rule: a bead that already belongs to an epic is never moved.** Not with
confirmation, not with restore-on-completion — a crashed run would leave the
lie in place, and recovery is the case this wave exists to make reliable.

## Two kinds of run

That rule collapses the four input shapes into two.

### Owned run — the run's scope IS an epic

Three ways to get one, and none of them moves a bead that has a home:

1. **One epic given** — use it. Nothing is created.
2. **Several ROOT epics given** — create the epic-of-epics. This gives each
   root a parent, which restructures the project's top level, so **confirm
   first**. Nothing under-reports: a root epic's own status is computed from
   its own children and is unaffected.
3. **Loose beads given, none with a parent** — create a run epic and adopt
   them. Benign: they had no parent to lose, and the membership is decent
   provenance afterwards.

Full swarm applies, the manifest lives on the epic, and the epic id is the run
id.

Measured: `bd ready --parent <run-epic> --exclude-type=epic` reaches through an
adopted epic to its grandchildren, so the epic-of-epics queue works.

**But `bd epic status` on an epic-of-epics counts child EPICS, not leaf beads**
— the probe read `0/1`, not `0/2`. So run-level Definition of Done for case 2
needs `swarm status` or a per-epic roll-up, not `epic status` alone. That is a
correction to `kyyl` as filed.

### Referencing run — the scope is a selector over beads the run does not own

A label, a priority, or an explicit id list containing beads that already live
in other epics. **Case 4 is not its own case**: an explicit list that crosses
epics is inherently cross-cutting, which is the same shape as a filter, so it
lands here.

The record is **a plain bead, and calling it anything else was the mistake**.
An epic that owns nothing is a type error — `bd epic status` on it reads 0/0,
and it invites exactly the "why not adopt the beads" question that the
measurement above forbids. So: `--type task`, label `bb-run`, opened
`in_progress` and claimed so `bd ready` never returns it, carrying the selector
and manifest as `--metadata` and amendments as `amend:` comments. Closed at run
end.

It is honestly just a durable store for run context, and that is enough to earn
it: it is what retires `$RUN_DIR`, and it gives the run a greppable id that
survives compaction. No swarm; `plan-sprint` partitions; DoD is a count against
the re-evaluated selector.

**Alternatives rejected.** `bd remember` — memories are excluded from
`bd export` and are not synced, so they are not durable across machines.
A molecule — `bd mol`'s own help says "protos are template epics", so a
molecule is an epic underneath and inherits the same type error. A swarm
molecule — this *is* bd's native record-that-points-at-work-it-does-not-own,
but `bd swarm create` requires an epic, so it cannot serve the case that has
none.

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
| `x5ix` | attach the manifest to the epic; retire `$RUN_DIR` | attach it to **the epic for an owned run, the run bead for a referencing run** — identical to as-filed for the single-epic case |
| `205x` | take dependency waves from `swarm validate` | unchanged, **conditional**: epic selectors only; a filter selector keeps `plan-sprint`'s partition |
| `ddvt` | register with `swarm create`, `swarm status` as the live view | unchanged, **conditional**: epic selectors only |
| `kyyl` | run-level DoD via `bd epic status` | **corrected**: `epic status` counts child epics, not leaf beads, so an epic-of-epics run needs `swarm status` or a per-epic roll-up. A referencing run counts against the re-evaluated selector |
| `ccm9` | multi-epic runs via epic-of-epics | **survives, narrowed.** Epic-of-epics is the right answer for several ROOT epics, and is measured to work — the queue reaches through it. What it must add is the confirmation step (it restructures the top level) and the cross-epic dependency merge that `swarm validate` reports as a warning and nothing acts on |
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
