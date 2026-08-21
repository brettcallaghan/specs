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

## Swarm is one level deep — but it is not the dispatch mechanism

**Measured.** A run epic holding a sub-epic (two leaves) plus a direct leaf:

```
bd swarm status   →  Ready: sub-epic-A, direct-leaf     Progress: 0/2
bd swarm validate →  Total issues: 2   Wave 1: sub-epic-A, direct-leaf
```

Neither leaf appears and the sub-epic is listed as a dispatchable work item.

**Running swarm on the leaf epics instead does not fix it — it produces wrong
answers.** With `a2` (in E1) blocking `b1` (in E2):

```
bd swarm validate E2 →  Wave 1: b1, b2        ⚠ b1 depends on a2 (outside epic)
bd swarm status   E2 →  Ready: b1, b2         Blocked: (none)
```

`b1` is not ready. Its blocker has not started. Swarm scopes dependency
analysis to the epic it was given, notes the outside edge as a *warning*, and
then computes waves and readiness as if the edge were not there. The warning
makes it detectable; the outputs are still unsafe to dispatch from.

**The re-queried ready front is correct where swarm is not.** Same graph,
`bd ready --parent <run-epic> --exclude-type=epic` after each close:

| after | front |
|---|---|
| — | `a1`, `b2` |
| `a1` closed | `a2`, `b2` |
| `+ b2` | `a2` |
| `+ a2` | **`b1`** — the cross-epic edge released correctly |
| `+ b1` | (empty) |

It walks the whole DAG, through nesting and across epic boundaries, with no
precomputed table. **This is the dispatch mechanism.** `swarm validate`'s waves
are a lookahead for planning and reporting, not something to dispatch from.

### What nesting therefore costs

Not correctness. Two display signals:

- **The wave lookahead.** Valid on a flat epic; on a nested one, `plan-sprint`
  partitions, or the run simply shows the current front and no preview.
- **The progress count.** `bd epic status` counts direct children — the probe
  read `0/2` with all four leaves closed.

A transitive count has to be walked, because **`--parent` does not mean the
same thing on both commands** — measured against `agent-plugins-1ilc`:
`bd ready --parent` returns descendants at any depth, while
`bd list --parent` returns **direct children only** (10 leaf beads for an epic
with 32 descendants). Its own help says so; the two flags simply differ.

**So flatness is a quality-of-reporting preference, not a precondition.** An
earlier revision of this plan said the run epic must be flat and that
epic-of-epics was therefore dead. That was wrong: it followed from testing
swarm and assuming swarm was how work gets dispatched.

## Why membership cannot be a dependency edge

A run epic that *depends on* the beads instead of owning them would be fully
non-destructive. **bd refuses to build it.**

```
bd dep add <run-epic> --depends-on <task>
  → Error: epics can only block other epics, not tasks
bd dep add <task> --depends-on <run-epic>
  → Error: tasks can only block other tasks, not epics
```

`blocks` edges are type-segregated in both directions. Epic↔epic is allowed and
task↔task is allowed; epic↔task is not.

The non-blocking types are not segregated — `bd dep add <run-epic> --type
related <task>` succeeds — but such an edge creates nothing usable:
`bd ready --parent <run-epic>` returns no work and `bd swarm validate` reports
`Total issues: 0`. Same for `tracks`.

**Only `parent-child` — the parent field — creates membership.** That is the
single mechanism, which is why the hybrid ("adopt the free beads as children,
link the owned ones as dependencies") cannot be built either: the second half
has no implementation.

## Admission — turning the input into the run epic

Resolve a label or priority filter to a bead set first; then every input is a
set of beads and/or epics, and these rules apply.

| input | what happens |
|---|---|
| **one flat epic** | use it. Nothing created, nothing moved. |
| **one nested epic** | use it as-is. The queue is correct through nesting; say once that the wave preview and progress count are degraded. **Never flatten** — that moves beads with homes for a display signal. |
| **beads with no parent** | create the run epic and adopt them. Nothing is lost: they had no parent, and the membership is provenance afterwards. |
| **beads that already belong to an epic** | **ask before moving.** On decline, **exclude that subset from the run** and say so in the manifest and the report. |
| **several epics** | create the epic-of-epics and adopt the epics themselves. Each keeps its own children, so no bead moves and no source epic is emptied — only the epics gain a parent, which restructures the top level, so confirm that. Declining leaves them out of the run. |

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
| `205x` | **narrowed to a lookahead.** Dispatch reads the re-queried ready front, which is correct at any depth; `swarm validate` supplies a wave preview only on a flat epic, and must never be dispatched from where cross-epic edges exist |
| `67xf` | sequencing constraints as real `blocks` edges at grooming. Unchanged |
| `ccm9` | **survives.** Epic-of-epics is how several epics become one run, and the queue is correct through it. What it owns: the top-level-restructure confirmation, and the degraded-lookahead notice |
| `ddvt` | `bd swarm create` for registration, `bd swarm status` as the live view. Same flat-epic caveat |
| `kyyl` | run-level DoD. `bd epic status` is exact on a flat epic only; a nested run epic needs a walked count, since `bd list --parent` is direct-children-only |
| `we2q` | epic-level Definition of Ready in `groom-beads`. It should **prefer** flatness and state what a nested epic costs, not require it |

## The honest cost

A user handing over loose or explicitly-listed beads that already belong to an
epic must accept a move or accept that those beads sit out the run. That case
is genuinely unavoidable: the bead can only have one parent.

Whole epics are not affected — they are adopted intact, so the several-epics
case costs a confirmation about top-level structure and nothing else. What a
nested run epic gives up is the wave preview and the progress count, both of
which are recomputable.
