# Wave 3, reconsidered — one run epic, and how the user's input becomes it

Supersedes wave 3 as filed. Transient; delete when the change lands.

---

## The decision

**A run is a single epic.** The epic id is the run id, `bd ready --parent` is
the queue, the manifest lives on the epic, and the `swarm` commands apply.

What the filed plan got wrong was not the epic — it was assuming the user would
hand one over. The user hands over *whatever they have*: an epic, several
epics, a pile of bead ids, a label, a priority. **Admission** is the step that
turns any of those into the one epic, and it is where every consequence we
measured has to be handled.

## The constraint that shapes everything: swarm is one level deep

**Measured.** A run epic with one sub-epic (two leaves, one closed) and one
direct leaf:

```
bd swarm status   →  Ready: sub-epic-A, direct-leaf     Progress: 0/2
bd swarm validate →  Total issues: 2      Wave 1: sub-epic-A, direct-leaf
```

Neither leaf under the sub-epic appears, and the closed one is not counted.
`bd swarm validate` lists the **sub-epic itself as a dispatchable work item**.

`bd ready --parent` behaves the opposite way — it is transitive, returning
descendants at any depth (which is why `--exclude-type=epic` is mandatory).

So the queue reaches through nesting and `swarm` does not. **The run epic must
be FLAT: its children are the beads to run.** A nested run epic silently
degrades every swarm-derived signal — the wave partition, the live view and the
progress count — and would dispatch a sub-epic as if it were a bead.

This retires the epic-of-epics idea for multi-epic runs. It was measured to
work for the *queue* and that is not enough.

## Admission — turning the input into the run epic

Resolve a label or priority filter to a bead set first; then every input is a
set of beads and/or epics, and these rules apply.

| input | what happens |
|---|---|
| **one flat epic** | use it. Nothing created, nothing moved. |
| **one nested epic** | use it, and **say that swarm signals are one level deep** so the wave table comes from `plan-sprint` instead. Do not flatten it — flattening moves beads that have homes. |
| **beads with no parent** | create the run epic and adopt them. Nothing is lost: they had no parent, and the membership is provenance afterwards. |
| **beads that already belong to an epic** | **ask before moving.** On decline, **exclude that subset from the run** and say so in the manifest and the report. |
| **several epics** | ask before moving. Adopting the epics gives a nested run epic (degraded swarm); adopting their leaf beads empties the source epics of open work. Both are moves, so both need the same confirmation, and declining excludes them. |

**Decline is a first-class outcome, not a failure.** The run proceeds with the
beads that could be admitted, and the excluded set is recorded on the epic and
named in the report. That is what makes "always one epic" safe to insist on:
the user never has to choose between running the skill and keeping their
structure.

## Why moving is confirmed rather than assumed

**Measured.** A feature epic with two children, one closed and one open, reads
`◐ 1/2 complete (50%)`. Move the open child into a run epic and it becomes:

```
✓ 1/1 complete (100%) — eligible for close
```

Restoring the parent puts it back. So an unconfirmed move makes a project's
feature epic falsely report finished and eligible for `bd epic close-eligible`.

The confirmation must therefore name **which beads move and which epic each one
leaves** — not "some beads will be reparented". And there is no
restore-on-completion: a crashed run would leave the lie in place, and recovery
is the case this wave exists to make reliable.

## Scope is the epic, and it stays live

Membership of the run epic is the scope. A bead added to it mid-run is in
scope — that is the point, and it is why the run does not carry an enumerated
list.

The old "enumerate, never filter" rule is replaced by a **delta report**: on
every start and restart, recompute the queue, diff it against what the manifest
recorded, and state what changed before dispatching. Where the run was admitted
from a filter, re-run the filter too and offer any new matches for admission
under the same rules above. A run that silently grew is the failure; one that
grew and said so is the feature.

## Capability matrix

Now single-column, which is the point of the decision:

| | mechanism |
|---|---|
| run id | the epic id |
| queue | `bd ready --parent <epic> --exclude-type=epic` |
| dependency waves | `bd swarm validate` — flat epic only; `plan-sprint` partitions a nested one |
| live view | `bd swarm status` — same caveat |
| run-level DoD | `bd epic status` — counts direct children, so it is exact on a flat epic |
| registration | `bd swarm create <epic>` |
| manifest home | `--metadata` JSON on the epic |
| amendments | `amend:`-prefixed comments on the epic |
| excluded beads | recorded on the epic, named in the report |

`bd swarm create` takes exactly one epic, and given a non-epic issue it
auto-wraps by creating an epic with that issue as its only child — a reparent.
Never point it at a bare bead.

## Re-scoping the eight filed beads

| bead | becomes |
|---|---|
| `7vej` | a run is one epic; **admission** turns the user's input into it, with the move-confirmation and exclusion rules above. The run root, the `current` symlink and the generated run id retire |
| `x5ix` | manifest as `--metadata` on the epic, amendments as `amend:` comments, `$RUN_DIR` and `dispatch.md` retired, recovery reads the epic. As filed |
| `205x` | waves from `bd swarm validate`, **with the flat-epic precondition stated and the `plan-sprint` fallback for a nested one** |
| `67xf` | sequencing constraints as real `blocks` edges at grooming. Unchanged |
| `ccm9` | **closes as won't-do.** Epic-of-epics degrades every swarm signal to one level. Several epics is handled by admission — confirm, adopt or exclude — not by a wrapper |
| `ddvt` | `bd swarm create` for registration, `bd swarm status` as the live view. Same flat-epic caveat |
| `kyyl` | run-level DoD via `bd epic status`. Exact on a flat epic, which admission now guarantees for every run the skill creates |
| `we2q` | epic-level Definition of Ready in `groom-beads` — and it should now require **flatness**, since that is what the swarm signals depend on |

## The honest cost

A user whose work spans several epics must accept a move, or accept that those
beads sit out the run. There is no third option that keeps their structure
intact *and* gives the run one epic — that is what the measurements above
establish, not a design preference.
