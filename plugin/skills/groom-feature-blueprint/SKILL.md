---
name: groom-feature-blueprint
description: Translate an approved feature blueprint and its FRD into an
  executable implementation plan — beads if the project uses beads,
  otherwise a task-file directory. Enforces the readiness gate first and
  bounces spec gaps back into the blueprint's Open Questions instead of
  guessing. Use when a feature blueprint is ready for implementation
  planning, or when asked to groom, decompose, or plan a specified
  feature.
version: "1.2"
---

# Groom Feature Blueprint

**Role in the workflow:** this skill is the boundary between the
human-agent specification loop and autonomous execution. Upstream, the
`spec` skill authors the FRD, the feature blueprint, and the documents
they reference. This skill consumes them and decomposes the blueprint
into a graph of **work orders** — the framework's atomic, portable units
of executable work. It is a quality firewall, not just a translator: a
blueprint that cannot be planned mechanically is sent back, never patched
over with guesses.

It owns **decomposition and traceability** — deriving work orders from
deltas, wiring the graph, carrying AC coverage and the Definition of Done.
It does NOT own the **dispatch-readiness finishing pass** (assigning each
work order a model, tagging C4 blast radius, validating to the Definition
of Ready): that logic lives in one place, `groom-beads` (the work-order
DoR engine), which Phase 3 hands off to. In one line: this skill is the
blueprint-aware front-end that feeds the DoR engine.

Load alongside this skill (canonical copies live in the sibling
skills that own them — this skill deliberately keeps no copies, so
authoring and grooming can never disagree):

- `../spec/references/feature-blueprint-template.md` — the document
  being groomed
- `../spec/references/blueprint-readiness-checklist.md` — the Phase 1
  check definitions
- `../spec/references/work-order-template.md` — the format each work
  item is emitted in
- `../spec/references/ac-traceability-standards.md` — test citation
  obligations
- `../retro-spec/references/findings-capture-protocol.md` — the
  retro-evidence protocol Phase 3 vendors into the target project

---

## Phase 1 — Readiness Gate

The checks are defined in
`../spec/references/blueprint-readiness-checklist.md` — the single definition
of "ready to groom," which the `spec` skill runs as an authoring
preflight and this skill enforces as the gate. Do not restate or
paraphrase the checks; run them from that document. Resolve every
`[[slug]]` reference by reading the referenced document.

Two kinds of checks with different failure responses:

**Readiness checks** (checklist #1–#3) — is grooming premature?

WHEN a readiness check fails, report why grooming cannot start and
stop. Do NOT modify any document — a blueprint still being authored or
reviewed is not defective, just not ready; demoting it or injecting
questions would disrupt the specification loop mid-flight.

**Content checks** (checklist #4–#10) — is the approved design
plannable?

WHEN a content check fails, the skill SHALL NOT produce a plan.
Instead:

1. Write each gap as an entry in the blueprint's Open Questions section
   (create the section if absent): what is missing, what it blocks, who
   decides.
2. Set the blueprint's status back to `draft` and update `updated_at`.
3. Report the gaps to the user — this returns the feature to the
   human-agent specification loop.

Never resolve a gap by choosing an answer yourself, however obvious it
seems: an obvious answer is cheap for the human loop to confirm and
expensive to silently get wrong.

## Phase 2 — Decomposition

Derive work items from the blueprint — each item is one **work order**
(`../spec/references/work-order-template.md`). Do not invent design; every
item must trace to a delta, a DoD entry, or a sequencing constraint.

Decomposition produces the work order's *substance* — its Problem (the
delta and anchors), Design intent, Acceptance (the ACs), provenance
(`sources: [<feature-slug>]`, `delta:<KEY>`), and graph edges. It does NOT
set `tier`, `tier_reason`, or the C4 blast-radius fields: those are the
dispatch-readiness finishing pass, handed to `groom-beads` in Phase 3.
Deriving them here would fork routing logic that must live in one place.

**Work items:**

- **One item per element delta**, implementable and testable in
  isolation. The item includes its tests AND the fold of its delta into
  the owning architecture blueprint — code, tests, and documentation
  are delivered together as one unit, never as separate items.
  Documentation is not an upfront phase or a final pass; it travels
  with the code it describes.
- Each code item carries: the delta it implements (blueprint section
  reference) and its delta key from the AC Coverage legend, the ACs it
  covers with their `verified by` level (from the AC Coverage table),
  the test obligation (citation comments per the AC traceability
  standards), the code anchors from the delta, and the owning
  blueprint sections to update.
- **Citation obligations follow the proving delta.** WHEN an AC row
  names several deltas, the item derived from the marked proving delta
  owes the citation at the declared level; participating items list
  the AC without owing the citation.
- WHERE the owning blueprint declares a Change Guidance section, its
  constraints bind the items derived from that element's deltas:
  artifacts a lockstep-update rule couples are delivered within one
  item UNLESS the blueprint states how the rule is otherwise satisfied
  (e.g., single-PR delivery); that statement is a design decision the
  graph inherits, and its absence when coupled artifacts span deltas
  is a Phase 1 content failure (checklist #9), never the groomer's
  interpretation. Change Guidance verification gates inform each
  item's test obligation.
- **Tracker Reconciliation becomes graph structure.** For each issue
  the blueprint's Tracker Reconciliation section classifies: a
  *subsumed* issue is blocked on the superseding work item and gets a
  supersession note ("close or re-scope when <item> lands"); a
  *prerequisite* issue becomes a dependency edge into the items whose
  deltas build on it; an *out-of-scope* issue becomes a coordination
  note on the affected items. The final DoD item closes or re-scopes
  the subsumed issues.
- **Verification-timing notes.** For every item, check whether any of
  its ACs' proving deltas land on items outside its dependency
  closure. WHEN they do, the item carries a generated note: code work
  may proceed when the item is ready, but verification of the named
  ACs waits for the named items; the final DoD item is the backstop.
  Do NOT add dependency edges for this — serializing independent code
  work to protect a verification step trades real parallelism for a
  hazard the note and the DoD backstop already cover.
- **One item per Definition of Done entry not tied to a single delta**
  (flip draft contracts to `active`, reference new ADRs from the owning
  blueprints' Key Decisions), depending on the code items it concludes.
- **Every item ends with a Retro section** — the standing instruction
  for its implementing agent (often a spawned subagent that sees only
  the item) to record friction findings per the findings capture
  protocol (`../retro-spec/references/findings-capture-protocol.md`).
  Findings, not
  analysis: the implementer is a sensor for item quality —
  self-sufficiency gaps, unpinned decisions, stale anchors,
  unsatisfiable obligations. Phase 3 vendors the protocol into the
  project and states the section's exact text per backend.
- **A final item**: verify the Definition of Done — every delta folded,
  every AC cited at its declared level — set the blueprint's status
  to `retired`, and then run the `retro-spec` skill over the
  implementation phase, aggregating the findings recorded on the
  feature's items. Implementation is the phase that finally exercises
  the work items, so this is where their quality becomes evidence.
  The approved plan is the standing authorization for this retro (no
  further prompt needed); the RFCs it produces remain human-gated at
  review.

**Dependencies:**

- Every Sequencing Constraint in the blueprint becomes one or more
  dependency edges. Constraints are design decisions — the graph SHALL
  NOT contradict them.
- Add structural edges the deltas imply: a contract before its
  consumers, a schema migration before its readers, a shared component
  before the containers composing it.
- Beyond stated constraints and structural implications, prefer FEWER
  edges: independence is what lets agents parallelize.

**Sizing:** an item should be one reviewable unit of change. Split an
item only along boundaries the blueprint already draws (per delta, per
store, per contract) — never by inventing new design seams.

## Phase 3 — Serialization

Detect the backend, then emit the same graph in one of two formats.

First, for either backend, **vendor the findings capture protocol**:
copy `../retro-spec/references/findings-capture-protocol.md` to
`rfcs/findings-capture.md` in the target project (creating `rfcs/` if
absent; overwrite an older copy). Work items reference that
repo-local path, so any implementing agent — including one without
this plugin installed — can follow the protocol by reading a single
file inside the project, with no permission prompts and no skill
loading.

**Beads** — IF the project uses beads (a `.beads/` directory exists or
the project's docs say so), THEN serialize the canonical structure —
these conventions are normative, not per-session choices, so identical
inputs produce identical graphs:

- Create **one epic per feature**, titled from the blueprint, whose
  description links the blueprint and FRD paths and states the
  delivery unit (the blueprint's `delivery` declaration).
- Create one issue per work item via the `bd` CLI, each a **child of
  the epic**, with dependencies wired so ready/blocked states mirror
  the graph. Put the AC list, verified-by levels, blueprint section
  references, and any generated supersession / coordination /
  verification-timing notes in each issue's description.
- Stamp each issue with its **delta key** as metadata or a
  `delta:<KEY>` label — the key is what re-grooming diffs by.
- End every issue description with this Retro section, verbatim:

  ```markdown
  ## Retro

  Before closing this item, record retro findings on this issue —
  one `bd comment` per finding, prefixed `retro:` — following the
  findings capture protocol in `rfcs/findings-capture.md`. Friction
  only: what happened plus the item or spec text that caused it (or
  the gap where guidance was missing). No analysis, no RFCs — the
  plan's final DoD item runs the aggregated retro.
  ```
- Label the epic, every created issue, and every pre-existing issue
  named in Tracker Reconciliation with the **feature slug from the
  blueprint frontmatter** (e.g. `feature-local-packages`) — never an
  invented name. The feature's closure set is then exactly
  `bd list --label <feature-slug>`, and the final DoD item verifies
  completion by that label query.

Then run the **finishing pass**: invoke `groom-beads` over the epic's
children (`bd list --label <feature-slug>`) to assign each work order a
`tier` + `tier_reason`, tag its C4 blast radius (`c4_changes` /
`c4_impacts` / `c4_docs`), and validate the graph against the Definition
of Ready. Decomposition gave the beads their Problem / Design / Acceptance
and edges; this pass makes them dispatch-ready. Do NOT hand-roll model
routing or C4 tagging here — `groom-beads` is the single home for that
logic, and duplicating it is how the two skills drift.

**Task-file directory** — otherwise, write `plans/<feature-slug>/`:

<!--
This is the interim no-tracker fallback. A first-class file-backed
serialization of the work-order format
(`../spec/references/work-order-template.md`) is
deliberately not yet defined; when it is, it supersedes this shape and
the finishing pass applies here too. Until then, tier / C4 fields are
absent on this path — file-backed work is not the beads dispatch model.
-->

- `index.md` — the feature, links to FRD and blueprint, the item list
  in dependency order, and a mermaid graph of the edges.
- One file per work item, `NN-<item-slug>.md`, with frontmatter:

```yaml
---
id: <feature-slug>/NN-<item-slug>
title: <imperative summary>
delta: <KEY>            # delta key from the blueprint's AC Coverage
                        # legend — what re-grooming diffs by
status: todo            # todo | in-progress | done | blocked
depends_on: []          # ids of prerequisite items
acs: [AC-<PREFIX>-001.1]  # ACs this item covers
verify: [unit]          # levels from the AC Coverage table
---
```

  Body: what to build (referencing the blueprint delta), code anchors,
  out-of-scope notes, and done-criteria including the test citation
  obligation — ending with this Retro section, verbatim:

  ```markdown
  ## Retro

  Before setting this item done, record retro findings by appending
  them to a `## Retro findings` section at the end of this file,
  following the findings capture protocol in
  `rfcs/findings-capture.md`. Friction only: what happened plus the
  item or spec text that caused it (or the gap where guidance was
  missing). No analysis, no RFCs — the plan's final DoD item runs
  the aggregated retro.
  ```

Implementing agents claim an item by setting `status: in-progress`,
and an item is ready WHEN every `depends_on` id has `status: done`.

After serializing, set the blueprint's status to `in-progress` and
record the plan location in its frontmatter — `plan: <epic-id>` for
beads, `plan: plans/<feature-slug>/` for task files — so re-grooming
agents, DoD verifiers, and humans reading the blueprint mid-flight
find the graph without searching the tracker. The pointer retires
with the document.

---

## Rules

- The plan derives from the spec; it never extends it. A needed-but-
  missing design decision is a Phase 1 failure, even if discovered
  during Phase 2 or 3 — bounce it back.
- Do not prescribe function signatures or file layouts in work items;
  carry the blueprint's anchors and let implementers choose structure.
- The delivery unit is the blueprint's `delivery` declaration:
  `single-pr` (assumed when the field is absent) or `staged`, whose
  stage boundaries arrive as sequencing constraints. Read it; never
  assume a delivery model the reviewer did not approve.
- Re-grooming after a blueprint change: regenerate the plan, diffing
  by delta key — preserve the status of items whose delta and ACs are
  unchanged.

---

## After grooming — offer a retro

Grooming is the first downstream exercise of the blueprint, so it is
the moment retro evidence exists. WHEN Phase 3 completes — or WHEN
Phase 1 bounced the blueprint (every bounced gap is a finding about
the authoring side) — offer to run the `retro-spec` skill on the
blueprint, its FRD, and the skills and templates that produced them.
The human accepts or declines; never run the retro unprompted.

That offer is the **document lens** (`retro-spec`, project specs). If
`groom-feature-blueprint` itself misled you — a step here, the
readiness gate, or a prescribed script — also offer to run
`retro-skill` on `groom-feature-blueprint` for the **skill lens**. A
clean run needs nothing.
