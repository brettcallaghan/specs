---
title: "RFC: A stated constraint that must not become an edge, a
  Definition-of-Done entry already done, and one label query serving two
  different sets"
description: Give Phase 2 a rule for a Sequencing Constraint that
  expresses verification timing, for a DoD entry already satisfied at
  groom time, and for a single-valued artifact several folds touch; and
  scope the finishing-pass query to the work orders it is meant to reach
slug: rfc-self-timing-constraints-and-satisfied-dod
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: groom-feature-blueprint
skill_version: "1.2"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.15.0"
---

# RFC: Timing constraints, satisfied DoD entries, and a two-purpose label query

## Summary

`groom-feature-blueprint` v1.2 was run over
[[feature-spec-reference-integrity]], producing epic `agent-plugins-9kb`
and ten work orders. The Phase 1 gate passed all ten checks, so every
finding below comes from Phases 2 and 3. The skill's decomposition rules
held up well; what it lacks are rules for three situations this blueprint
presented — a Sequencing Constraint that expresses verification timing
rather than landing order, a Definition-of-Done entry already satisfied
before grooming began, and a single-valued document edit that several
distributed folds touch. A fourth finding is narrower: the query Phase 3
names for the finishing pass returns more than twice the beads that pass
is meant to reach.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Skill critiqued | `groom-feature-blueprint` v1.2 |
| Task | Groom the approved reference-integrity feature blueprint into an executable plan |
| Documents consumed | `specs/features/reference-integrity/feature-spec-reference-integrity.md`; `specs/features/reference-integrity/frd-spec-reference-integrity.md`; [[module-document-core]], [[module-spec-profile]], [[container-spec-cli]], [[contract-specs-configuration]], [[adr-018]] |
| Output produced | Epic `agent-plugins-9kb`, work orders `agent-plugins-9kb.1`–`.10` |
| References invoked | blueprint-readiness-checklist v0.6.0; feature-blueprint-template v1.4; work-order-template v1.3; ac-traceability-standards v0.2.0; findings-capture-protocol v0.2; work-order-definition-of-ready v1.2 |
| Downstream skill invoked | `groom-beads` v1.1 (finishing pass) |
| Harness | Claude Code on the web, remote execution container, multi-root workspace (`agent-plugins` + `strata`) |
| Model | Withheld: this session's harness supplies the model identifier under a policy forbidding it in any artifact pushed to a repository. Same handling as RFC-0010; tracked by `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md` |
| Runs observed | 1 |
| Run date | 2026-08-07 |
| Plugin version | 0.15.0 |

## Evidence

1. **[gap, blocker]** The blueprint's Sequencing Constraints contained
   this bullet
   (`specs/features/reference-integrity/feature-spec-reference-integrity.md`
   lines 592–594):

   > - The PREG delta SHALL land before LIFE, whose every check is
   >   inert without a lifecycle registration, and before REX's
   >   frontmatter extraction can be exercised against this
   >   repository's corpus.

   The second half is a condition on *verification*, not on landing.
   `plugin/skills/groom-feature-blueprint/SKILL.md:183` states the rule
   without qualification:

   > - Every Sequencing Constraint in the blueprint becomes one or more
   >   dependency edges. Constraints are design decisions — the graph
   >   SHALL NOT contradict them.

   Applied literally, this wires `PREG → REX` and yields the serial
   chain `MODEL → XPT → PREG → REX → RES → CHK → CMD`. The agent
   instead judged the clause to be a timing condition, wired no edge,
   and attached a verification-timing note to `agent-plugins-9kb.3` —
   contradicting the SHALL at line 184 on its own reading of the
   sentence's semantics. The skill already owns the concept it needed:
   `SKILL.md:151-158` defines verification-timing notes, but derives
   them only from AC coverage —

   > - **Verification-timing notes.** For every item, check whether any
   >   of its ACs' proving deltas land on items outside its dependency
   >   closure.

   — so an author-stated timing condition has no path to a note, and
   Sequencing Constraints is the only section available to state it in.
   No rule at Phase 2 for *a Sequencing Constraint that is not an
   ordering*.

2. **[gap, snag]** Two Definition-of-Done entries were already satisfied
   when grooming began. The blueprint requires "[[adr-018]] status:
   draft → approved" (line 702) and "`specs/adrs/README.md` carries an
   index entry for [[adr-018]]" (line 704). ADR-018 was already
   `status: approved` (`specs/adrs/ADR-018-references-as-index-edges.md`
   line 9) and already indexed (`specs/adrs/README.md` line 24).
   `SKILL.md:159` says:

   > - **One item per Definition of Done entry not tied to a single
   >   delta** (flip draft contracts to `active`, reference new ADRs
   >   from the owning blueprints' Key Decisions), depending on the code
   >   items it concludes.

   Followed literally this creates a work order for completed work. The
   agent improvised: it created no item, folded verification of both
   into the final DoD item, and recorded in that item's description that
   they were already satisfied. No rule at Phase 2 for *a DoD entry
   already satisfied at groom time*.

3. **[gap, snag]** `SKILL.md:119` requires each item to carry its own
   documentation fold:

   > isolation. The item includes its tests AND the fold of its delta
   > into the owning architecture blueprint — code, tests, and
   > documentation are delivered together as one unit, never as separate
   > items.

   Six of the ten work orders fold some part of their delta into
   [[module-document-core]]. Its Definition-of-Done entry names one
   version bump — "`version` bumped 0.2 → 0.3"
   (`feature-spec-reference-integrity.md` line 686) — and
   [[module-spec-profile]]'s names another (line 692). A version is
   single-valued; six distributed folds are not. Nothing in the skill
   says who owns it. The agent wrote "Set `version` to 0.3 if not
   already set" into every folding item to make the edit idempotent —
   an invented convention that appears nowhere in the skill or the
   template. No rule at Phase 2 for *a single-valued artifact edit that
   several distributed folds touch*.

4. **[misleading-text, snag]** `SKILL.md:243` names the finishing-pass
   input:

   > Then run the **finishing pass**: invoke `groom-beads` over the
   > epic's children (`bd list --label <feature-slug>`) to assign each
   > work order a `tier` + `tier_reason` …

   but `SKILL.md:239` assigns the same label to a wider set:

   > Label the epic, every created issue, and every pre-existing issue
   > named in Tracker Reconciliation with the **feature slug from the
   > blueprint frontmatter** … The feature's closure set is then exactly
   > `bd list --label <feature-slug>`

   The two purposes need different sets. Measured after labelling:
   `bd list --label feature-spec-reference-integrity` returns **24**
   issues — 10 work orders, 1 epic, and 13 pre-existing Tracker
   Reconciliation issues (`006`, `1ao`, `3md`, `9jo`, `9ko`, `a7m`,
   `dgn`, `e0t`, `gz3`, `h6c`, `j7f`, `krb`, `vg1`). Handing that set to
   `groom-beads` would tier and C4-tag fourteen beads the pass does not
   own, including a `deferred` bug and an umbrella epic. The agent
   narrowed the scope by hand to `agent-plugins-9kb.1`–`.10` when
   invoking the skill. The same query is also under-broad for its other
   stated purpose: `bd list` defaults to open issues, so the closed
   `agent-plugins-koi` — labelled by the Tracker Reconciliation rule —
   is absent from the 24, and a DoD item verifying completion by that
   query would not see it.

## Proposals

### P1 — Rule for a constraint that is not an ordering [blocker]

**Shape.** Gap — no rule at Phase 2 / Dependencies for a Sequencing
Constraint expressing verification timing.

**Problem.** `SKILL.md:183-185` makes every Sequencing Constraint an
edge and forbids contradicting one. `SKILL.md:151-158` defines
verification-timing notes but derives them only from AC coverage, so a
timing condition an author *states* has no route to a note. Finding 1's
bullet is exactly that, and the two readings differ by five serialized
deltas. The agent had to contradict an explicit SHALL to get the right
graph, with nothing in the skill authorizing it.

**Change.** Insert after `SKILL.md:185`, in the Dependencies list:

> - **A constraint about when something can be VERIFIED is not an
>   ordering constraint.** WHEN a Sequencing Constraint conditions a
>   later *verification* rather than a later *landing* — "X SHALL land
>   before Y can be exercised / tested / measured against Z" — it SHALL
>   become a verification-timing note on the affected item, never a
>   dependency edge. Read the constraint's object: an ordering names a
>   delta ("before the LIFE delta"), a timing condition names an act
>   ("before REX's extraction can be exercised"). WHERE one bullet
>   states both, split it and treat each half by its own kind.
> - IF the reading is genuinely ambiguous, THEN it is a Phase 1 content
>   failure (checklist #7) — bounce it rather than choose, because the
>   two readings produce different graphs and the choice is the
>   blueprint author's.

and extend the Verification-timing notes bullet at `SKILL.md:151` to say
these notes have two sources: derived from AC coverage, and stated by
the blueprint per the rule above.

**Evidence.** Finding 1.

**Backward compatibility.** Additive. No existing plan changes: a
blueprint whose constraints are all orderings grooms identically. The
one plan this would have changed — `agent-plugins-9kb` — was already
groomed to the outcome this rule prescribes.

### P2 — Give Phase 2 a rule for a DoD entry already satisfied [snag]

**Shape.** Gap — no rule at Phase 2 for a Definition-of-Done entry
completed before grooming ran.

**Problem.** `SKILL.md:159-161` says one item per DoD entry not tied to
a single delta. It assumes every entry names outstanding work. A
blueprint authored alongside an ADR routinely has that ADR approved and
indexed before grooming, as finding 2 shows. The literal rule produces
a work order whose acceptance criteria are already met — waste at best,
and at worst an agent "doing" it and touching a document that needed no
change.

**Change.** Append to the bullet at `SKILL.md:159`:

> WHEN a Definition-of-Done entry is already satisfied at groom time,
> do NOT create an item for it. Verify the claim against the repository
> rather than trusting the blueprint — a DoD entry is written before the
> work and may name a transition that has since happened — and record
> the already-satisfied entries in the final DoD item's description, so
> that item confirms them rather than re-doing them. An entry that is
> partly satisfied yields an item scoped to the remainder.

**Evidence.** Finding 2.

**Backward compatibility.** Additive; changes only the case where an
entry is already done, which today produces a redundant item.

### P3 — Say who owns a single-valued edit several folds touch [snag]

**Shape.** Gap — no rule at Phase 2 for a shared, single-valued
artifact edit under distributed folds.

**Problem.** `SKILL.md:119-122` distributes the documentation fold
across items, which is right — documentation travels with its code. But
a blueprint's `version` is one value, and finding 3's feature has six
items folding into one blueprint. The skill is silent, so the agent
invented an idempotent instruction. A different agent might assign the
bump to the last item by dependency order (fragile — "last" changes as
edges change), to the final DoD item (splitting documentation from
code, which `SKILL.md:119` forbids), or to all six non-idempotently
(a merge conflict per item).

**Change.** Append to the bullet at `SKILL.md:119`:

> WHERE several items fold into the same document, each item folds only
> its own delta's sections. A single-valued property of that document
> that the Definition of Done changes once — a `version` bump above all
> — SHALL be stated in every folding item as an idempotent instruction
> ("set `version` to <target> if it is not already"), never assigned to
> one nominated item. Nominating one item makes every other item's fold
> depend on it landing first, which is the serialization the
> distributed fold exists to avoid.

**Evidence.** Finding 3.

**Backward compatibility.** Additive, and it standardizes what this run
already did ad hoc. No existing work order needs rewriting.

### P4 — Scope the finishing-pass query to the work orders [snag]

**Shape.** Misleading text — `SKILL.md:243` names a query returning
2.4× the beads the finishing pass owns.

**Problem.** One query, `bd list --label <feature-slug>`, is named for
two different sets: the finishing-pass input (`SKILL.md:243`) and the
feature's closure set (`SKILL.md:239-241`). The labelling rule at
`SKILL.md:238` deliberately puts pre-existing Tracker Reconciliation
issues in that label, so the two sets cannot coincide. Measured: 24
returned, 10 owned (finding 4). An agent following the text literally
hands `groom-beads` a deferred bug, an umbrella epic, and twelve
unrelated issues to tier and C4-tag. The closure-set use has the
opposite defect — `bd list` defaults to open, so a closed member is
invisible.

**Change.** Two edits.

At `SKILL.md:243`, name the children rather than the label:

> Then run the **finishing pass**: invoke `groom-beads` over the work
> orders this grooming created — the epic's children, `bd list --parent
> <epic-id>`, NOT `bd list --label <feature-slug>`, which by the
> labelling rule above also returns the epic and every pre-existing
> Tracker Reconciliation issue. Only the created work orders take a
> `tier` and a C4 blast radius; a pre-existing issue keeps whatever
> grooming it already had.

At `SKILL.md:240`, make the closure-set query status-complete:

> `bd list --label <feature-slug> --status all`, and the final DoD item
> verifies completion by that query — `--status all` because `bd list`
> defaults to open issues and a closure set exists precisely to catch
> the members that are already closed.

Verify the exact flag spellings against `bd list --help` at
implementation time; `SKILL.md:0`'s own probe-the-CLI rule applies.

**Evidence.** Finding 4.

**Backward compatibility.** The parent query is a narrowing, so no bead
that should have been tiered stops being tiered. The `--status all`
change widens the closure set, which is its purpose; no DoD item has run
against the old form yet in this repository.

## Considered and Not Proposed

- **`bd-health.sh` appearing to exit 0 after running no checks.** Raised
  as a candidate script defect during this session and **retracted on
  verification**: the observation came from `echo "EXIT=$?"` placed after
  a pipeline, which captures `tail`'s status, not the script's. Measured
  directly, `bash bd-health.sh` from a directory with no `.beads/` exits
  **2**, exactly as `plugin/skills/groom-beads/SKILL.md:279` documents
  and as `bd-health.sh:26-27` implements. No defect; the error was the
  agent's measurement.
- **`bd create --type epic` doubling the `[epic]` title prefix.** The
  created epic rendered as "[epic] [epic] Spec Corpus Reference
  Integrity" and was fixed with `bd update --title`. The skill does not
  ask for the prefix — the agent copied it from the existing epic's
  stored title. House style, not a skill defect.
- **`bd` losing its workspace when the shell's cwd moves.** Two commands
  failed with "no beads database found" after a `cd` to a scratch
  directory; `export BEADS_DIR=<repo>/.beads` fixed it, and the error
  message itself named the remedy. `groom-beads` already tells the agent
  to locate the beads directory at step 0. Agent-side cwd management, not
  a skill gap.
- **The stale beads database on session start.** Ten of fourteen Tracker
  Reconciliation issues were missing until `.beads/issues.jsonl` was
  imported, because the SessionStart hook does not register in a
  multi-root workspace. Fully documented in the project's `CLAUDE.md`
  with the manual remedy, which worked. Environmental and already
  written down.
- **The mapping-table field name in `groom-beads`.** Real, but it is a
  finding about a different skill; filed separately as
  `rfcs/inbox/groom-beads/rfc-self-acceptance-field-name.md` per the
  one-RFC-per-skill rule.

## Non-Goals

- No change to the Phase 1 readiness gate. It passed all ten checks on
  this blueprint and none of these findings is a gate failure — P1's
  fallback deliberately routes the ambiguous case to the existing
  content check 7 rather than adding an eleventh check.
- No change to the division of labour with `groom-beads`. P4 corrects
  which beads are handed over, not what the finishing pass does to them.
- No change to the work-order format or the Definition of Ready; both
  held up.

## Adoption

P1–P4 are independent and each touches only
`plugin/skills/groom-feature-blueprint/SKILL.md`. P1 is the only one
that changes what a plan looks like; P2 and P3 change how many items a
plan has and what they say; P4 changes which beads a downstream skill
receives.

P1 SHOULD land first: it is the blocker, and it is the one whose absence
required contradicting an explicit SHALL to get a correct result. P4 is
the cheapest and is worth landing alongside it, because a wrong
finishing-pass scope silently mutates beads outside the feature.

All four are documentation-only edits to one file; none requires a
script change, a template change, or re-grooming any existing plan.
`groom-feature-blueprint` would go 1.2 → 1.3.

Expected effect: the next blueprint carrying a timing condition, an
already-satisfied DoD entry, or a multi-item fold grooms without the
agent inventing a convention, and the finishing pass touches only the
beads it owns.

## Resolution
