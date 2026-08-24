---
name: sprint-scope
description: Resolve a sprint's scope — turn whatever the user gave (an epic,
  several epics, bead ids, a label, a priority) into exactly one run epic with
  its ceremony epics and one flat swarmed target, and take the run's single
  authorization gate in the same interaction. Use when a sprint orchestrator is
  opening a run, when asked to set up or begin a sprint over a set of beads, or
  when a run must be re-scoped. Returns the run epic id and the admission
  record — which beads were admitted, which excluded and why, which moved and
  from where. It dispatches nothing and grooms nothing.
version: "0.2"
user-invocable: false
---

# sprint-scope

Resolve the **one epic** the run is scoped by, and get every decision the run
will need authorized in one interaction.

**You return facts; you dispatch nothing.** The orchestrator that invoked you
takes the run from there.

**You do not run pre-flight and you do not groom.** Pre-flight is a script the
orchestrator runs directly — before it invokes you, and again with your epic id
once you return it. Readiness is a property of a *groomed* bead, so the
readiness gate belongs to the grooming step, not here; asking about readiness
before grooming answers a question about beads that are about to change.

## Inputs

| | |
|---|---|
| the user's scope | an epic, several epics, bead ids, a label, a priority — or nothing, in which case ask once |
| standing authorization | what the user pre-authorized, WHEN they did |

## What you return

- `BB_EPIC` — the run epic. Its id **is** the run id, and
  `bd ready --parent <epic> --exclude-type=epic` **is** the queue.
- `BB_TARGET` — the flat target epic beneath it, and the id `bd swarm` is
  given. Never the run epic: pointed there, swarm counts the ceremony epics as
  the work and never sees a bead.
- The admission record: which beads were admitted, which were **excluded** and
  why, which were **moved** and from where.
- The authorization set, as decided — including every row left at its default.
- Anything that BLOCKED, named, with the run not started.

**Say what you could not establish** rather than omitting it. A run opened on
an unverified fact is the failure this step exists to prevent.

## Admission — resolve the run epic

**A run is one epic.** Its id is the run id, its descendants are the queue, and
it is what recovery reads. The user hands over whatever they have; admission
turns that into the epic.

If they gave no scope at all, ask once: "Which beads? (epic / label / priority
/ explicit ID list)" — the only question allowed before the gate.

**Resolve a label or priority to a bead set first** (`bd ready --label <l>`,
`bd ready --priority 1`, `-n 50` to raise the limit). Then every input is beads
and/or epics, and these rules apply:

| what they gave | what you do |
|---|---|
| **one epic** | use it. Create nothing, move nothing |
| **beads with no parent** | create the run epic, adopt them. Nothing is lost — they had no parent |
| **beads that already belong to an epic** | **ask first**, naming each bead and the epic it would leave. On decline, EXCLUDE those beads and run without them |
| **several epics** | create the run epic over them, adopting the EPICS themselves. Each keeps its own children, so no bead moves and no source epic is emptied. Confirm — it gives those epics a parent they did not have, which restructures the project's top level |

**Never move a bead to make the run tidier.** Moving a bead out of its epic
makes that epic report complete when it is not: a feature epic with one closed
and one open child reads `1/1 complete (100%) — eligible for close` once the
open one is taken. That is a project's structure damaged by a transient run,
and there is no undo step that survives a crash.

**Declining is a normal outcome, not a failure.** The run proceeds with what was
admitted; the excluded beads are recorded and named in the final report. A user
must never have to choose between running a sprint and keeping their
structure.

**Never flatten a nested epic**, and never adopt the leaf beads out of the
epics you were given. The queue is correct through nesting; only the swarm
preview and the progress count degrade, and both are recomputable. Flattening
trades a real structure for a display detail, and emptying a source epic of its
open work is the false-complete failure above.

**That is not in tension with the flat-target rule.** The target this run
CREATES is flat; an epic the user already owns is adopted whole and becomes its
own target, swarmed separately. Flatness is a rule about what the run builds,
never a licence to restructure what it was given.

**What a nested run epic costs, stated once at run start:** `bd swarm validate`
and `bd swarm status` are one level deep, so they report sub-epics rather than
leaf beads. The partition comes from `sprint-plan` instead, and the progress
count from `bb-status.sh run <epic>`, which walks the tree. Dispatch is
unaffected — the re-queried ready front is transitive and honours dependency
edges that cross between the adopted epics.

**Cross-epic dependency edges are the reason several epics are being run
together, and `swarm validate` reports them as warnings and acts on none of
them.** Surface every such warning at run start rather than dropping it: an
edge from a bead in one adopted epic to a bead in another is exactly what makes
the swarm preview untrustworthy, and it is what the ready front gets right.

Then print the scope: the epic id, the ready count, and any exclusions.

### Build the shape, not just the epic

**A run epic is a container of PHASE EPICS, one of which is the target.** Create
them in the same act as the run epic, chained `epic→epic` in order:

```
run epic  ──  pre-flight → scope → plan → TARGET → review → retro → report → post-flight
                            │                 └ the work, as flat TASK children
                            └ AUTHORIZATION GATE, a TASK child of scope
```

Three rules, and each of them is forced rather than chosen:

- **Phases are epics.** `bd dep add` refuses task→epic and epic→task in both
  directions, so nothing but an epic can block the target. One `epic→epic` edge
  then blocks every bead beneath it.
- **The target is FLAT — tasks only — and that binds admission, not just
  planning.** A sub-epic added mid-run breaks `bd swarm validate`'s count
  exactly as a planned one would: it is one level deep and reports the sub-epic
  as the work item. Adopted epics that are genuinely nested become a **second
  target** under the run epic, swarmed separately; they are never flattened and
  never adopted as leaves (below).
- **The authorization gate is a TASK child of the scope epic**, gated with
  `bd gate create`. Gating the scope epic itself would need the refused
  task→epic edge, and is unnecessary — the scope epic already blocks the plan
  epic, so holding its task child holds the chain.

The measurements behind all three are in the coordinator's run-record
reference, which is where the shape is specified; do not restate them here.

### Scope stays live, and the delta is reported

Membership of the epic IS the scope, so a bead added to it mid-run is in scope.
That is deliberate — it is why the run carries no frozen id list.

**On every start and restart:** recompute the queue, compare it against what
the manifest recorded, and state what changed before dispatching. Where the run
was admitted from a label or priority, re-run that filter too and offer new
matches for admission under the rules above. A run that silently grew is the
failure; one that grew and said so is the feature.

`bd` is the only tracker. A second progress store is a correctness hazard, not
redundancy.

## The authorization gate — one interaction, then silence

**This step carries the run's single human interaction.** It is deliberately
early: the human invokes, authorizes, and does not hear from the run again
until the pull request.

Ask the whole set at once, with the resolved value stated per row so a human
can accept the lot in a sentence.

**Resolve each row before asking**, in this order: `sf.toml`'s `[run]` if it
carries the key, else the skill default below. The skill defaults all say "do
the smaller thing and say so", which is the only kind of default safe to apply
without asking; a project that configures a row has said something louder than
that, and gets to.

| decision | `[run]` key | skill default |
|---|---|---|
| **Push mode** | `push_mode` | defer — the orchestrator pushes per batch |
| **Pull request** | `pull_request` | open as draft at the start, mark ready at the end |
| **Moving beads** that already belong to another epic | `move_foreign_beads` | **do not move** — exclude them and run narrower |
| **Scope growth** from review findings | `admit_review_findings` | file the beads, do **not** admit them into this run |
| **Tier escalation** on repeated failure | `tier_escalation` | escalate one tier, once, and report it |
| **Review loop budget** | `review_budget` | 2 loops (three passes total) |
| **Merge to the default branch** | `merge_to_default` | never — that is the pull request's, not the run's |

**Configuring a row changes what is confirmed, never whether it is.** A project
that sets all seven still gets the gate — as one confirmation of seven resolved
values rather than seven questions, which is the point. **A run that stops
asking is a run that authorized itself.**

Record the outcome in the manifest's `standing_auth` — every row's value **and
where it came from**, `sf.toml` or the skill:

```json
"standing_auth": {
  "push_mode":     {"value": "defer_push", "source": "sf.toml"},
  "review_budget": {"value": 2,            "source": "default"}
}
```

The source matters because the record exists to stop a row being re-litigated
mid-run, and a project-level default is one more place a decision can hide from
whoever reads the report. A row nobody remembers deciding is a row that will be
re-litigated mid-run, and mid-run is exactly when there is nobody to ask.

### After this, the run does not ask again

Every question the run could need must be askable **here**, because it gets no
second chance. After the gate, an unauthorized decision is never a stop: the
agent decides, records it, and the report names it.

That holds even for the four decisions that change what *correct* means — an
acceptance criterion needing amendment, a standard/requirement/AC tension, a
contract or interface change, a spec silent where a bead needs it to speak.
Those are recorded as **incidents**, not escalated. `run-bead` carries the
rule.

**The human may still ask you anything, at any time.** The asymmetry is
one-way: answering costs the run nothing and interrupts nothing, while a run
that stops to ask is not autonomous however reasonable the question.
