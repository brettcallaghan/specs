---
name: groom-beads
description: >-
  Materialize and groom work orders as beads (issues in the Beads / `bd`
  git-native issue tracker) so each is Ready — dispatchable to an isolated,
  zero-context agent. Use this skill whenever the user wants to groom,
  refine, triage, prepare, clean up, or "get beads ready" — or mentions a
  backlog of beads, a `.beads/` directory, `bd ready`, running beads in
  batches, assigning capability tiers to beads, or wiring bead dependencies
  before a long-running or orchestrated session. Trigger even if the user
  just says "groom these" or "make these beads ready" in a repo that uses
  beads.
version: "1.6"
---

# Bead Grooming

## What this skill is

`groom-beads` is the **work-order Definition of Ready engine for the beads
backend**. A *work order* is this framework's portable, atomic unit of
executable work — everything a fresh agent needs to complete one job, with no
planning conversation to lean on. This skill materializes work orders as
**beads** and drives each to the **Definition of Ready**, so an orchestrator can
pull ready beads, group them into **batches**, and dispatch each to an
**isolated sub-agent** with zero session context.

Every sub-agent is a *fresh agent with zero context*, so the job is not "fill
in fields" — it is to **manufacture context-independence and correct
ordering**, the two things an orchestrator consumes. A bead that reads fine to
someone who was in the planning session and is unworkable cold is the central
failure mode, and it is invisible until it detonates three beads deep.

The north star, applied to every bead you touch:

> **A bead is groomed when a fresh sub-agent, with zero session context, can open
> `bd show <id>` and complete the work correctly — right approach, verifiable
> done-ness — without asking anyone anything.**

Two things live elsewhere and are NOT restated here, so authoring and grooming
can never disagree:

- **What "ready" means.** The criteria are the `spec` skill's. Load and enforce
  them; do not paraphrase or fork them:
  - `references/shared/work-order-definition-of-ready.md` — the checklist.
  - `references/shared/work-order-field-guide.md` — the three text fields kept
    distinct, and the tier-routing rubric.
  - `references/shared/work-order-template.md` — the work-order format this
    skill materializes into beads. Write each work order's prose with the
    `tech-writer` skill, taking the tone from the template's reader profile.
- **Where work orders come from.** A standalone bug/chore you groom directly, or
  a feature blueprint decomposed by `groom-feature-blueprint` — which hands its
  finishing pass (tier, C4, validation) to this skill.

What lives *here* is the beads-specific machinery: probing the `bd` CLI, mapping
the work-order format onto bead fields and metadata, wiring the edge graph,
validating the result, and previewing the partition.

## The work-order → bead mapping

One work order is exactly one bead. The format maps onto a bead like this:

| Work order | Bead |
|------------|------|
| `title` | `title` |
| `kind` (epic / feature / task / bug / chore) | `type` |
| `priority` | `priority` |
| `## Problem` | `description` |
| `## Design` | `design` |
| `## Acceptance` | `acceptance` |
| `depends_on` / `blocks` | `blocked-by` / `blocks` edges |
| `parent_slug` | `parent` edge |
| `related` / `discovered_from` | `related` / `discovered-from` edges |
| `tier`, `tier_reason`, `est_size`, `c4_*` | `metadata` JSON keys |
| `sources: [feature-<slug>]`, `delta` | a `feature-<slug>` label and a `delta:<KEY>` label/metadata |

A work order carries **no status field**: bead `status` and `grooming_status`
own the execution and grooming lifecycles respectively. This is deliberate — the
portable format leans on the tracker for status rather than keeping a second,
drifting copy. The authoritative on-disk shape for the beads materialization —
field names, edge names, enums, required fields — is `references/bead.schema.json`.

## Before you start: verify the CLI, don't assume it

`bd`'s flags move between versions and this skill must not hard-code guesses.
At the start of a grooming pass, spend a moment to ground yourself:

- Confirm `bd` exists and find the beads dir (`bd --version`, look for `.beads/`).
- Probe the flags you are about to use — `bd create/update/dep/lint --help`.
  Names like `--design`, `--acceptance`, `--set-metadata`, `--append-notes` are
  what current Beads uses; **adapt to what `--help` shows**. Where a capability
  is missing, degrade gracefully (fold acceptance into the description under an
  `## Acceptance` heading) and say so in the report rather than failing.
- **Where `bd` is unavailable and you emit bead files instead**, conform to
  `references/bead.schema.json` — it is authoritative for field and edge names,
  so do not improvise: `dependencies` not `deps`, `id` not `slug`, and
  parentage is a dependency of type `parent`, never a top-level field.
- Load durable project knowledge so grooming matches house style: `bd prime`,
  and skim existing `bd remember` insights. Past retros are the memory of what
  previous grooming got wrong — honour them.

Never invent bead IDs, component names, or dependencies. If you're unsure a
dependency is real, say so in the report instead of wiring a guess.

## The workflow

Copy this checklist into your response and check items off as you go.
(A checklist copied into your response is not a task tool — WHERE the project bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress visible.)

```
Grooming progress:
- [ ] 0. Probed the `bd` CLI; loaded `bd prime` / remembered insights
- [ ] 1. Scoped the grooming set; showed the user what will change
- [ ] 2. Every bead driven to the spec-owned Definition of Ready
  - [ ] 2a. No open decisions on any bead (DoR check 9)
  - [ ] 2b. Every bead reported ready / NOT-READY, including ones you changed nothing on
- [ ] 3. Dependency graph wired (every edge justified, none fabricated)
- [ ] 4. Tier + tier_reason on every bead (never null)
- [ ] 5. C4 blast radius tagged (or explicitly skipped: no C4 docs)
- [ ] 6. Validated — 0 errors, re-run after each fix
- [ ] 7. Reported: changes, still-not-ready + gaps, flags for the human
```


Work through these phases. For a small backlog you can do it inline; for a large
one, groom bead-by-bead and keep a running report.

### 1. Scope the grooming set

Find what actually needs grooming — newly created beads, and existing ones
missing the marks below. Useful starting points: `bd list` filtered to open work,
`bd lint` to surface under-specified beads, `bd ready` to see what the
orchestrator would pick up right now. Show the user the set and roughly what
you plan to change before making sweeping edits to a large backlog.

### 2. Bring each bead to the Definition of Ready

This is the core loop. For each bead, drive it to pass every check in
`references/shared/work-order-definition-of-ready.md` — read it; it is the gate
you enforce. Keep the three text fields distinct using
`references/shared/work-order-field-guide.md`: agents habitually collapse
`## Problem` / `## Design` / `## Acceptance` into one blob, and keeping them
apart (immutable problem statement / mutable approach / checkable done-ness) is
most of the quality battle.

Map as you go, per the table above: the work order's Problem, Design, and
Acceptance become the bead's `description`, `design`, and `acceptance`. If the
installed `bd` lacks separate `--design` / `--acceptance` flags, fold them into
the description under `## Design` and `## Acceptance` headings and note the
degradation — the *structure* matters more than the field boundary.

Prefer editing over rewriting: enrich a thin description, don't discard the
author's intent. When a description leans on session context, the fix is to
*inline the missing context*, not to delete the sentence.

**A bead that already passes every check is a VERDICT, not a silence.** Handed
a set to gate — which is how a sprint calls you — most beads may need nothing,
and reporting them `ready — no change` is the answer the caller asked for.
Confirm the checks against the bead; do not re-derive its Design, re-wire edges
that are already right, or edit it to prove you looked. **You are the only
holder of this checklist**, so a caller that skipped you has no verdict at all,
and one that gets silence cannot tell "ready" from "not examined".

**WHEN a bead cannot be brought to Ready, park it — which takes TWO writes,
and only the second one works.** `bd ready` reads `status`, not
`grooming_status`, so a bead flagged only with the metadata stays dispatchable
and reaches a fresh subagent looking handled:

```
bd update <id> --set-metadata grooming_status=blocked-needs-info
bd defer <id>          # status -> deferred; drops out of `bd ready`
```

Verify it took: the bead SHALL NOT appear in `bd ready` afterwards. Then flag
it in the report rather than shipping a half-groomed bead downstream silently.
Which `grooming_status` to use, and what each commits the next pass to:
`references/parking-a-bead.md`.

### 3. Wire the dependency graph

The orchestrator reads ordering **from the graph, not from prose or priority
numbers**. Encode real relationships with the right edge type (these realize the
work order's `depends_on` / `blocks` / `parent_slug` / `related` /
`discovered_from` fields):

- `blocks` / `blocked-by` — genuine sequential dependencies. This is what the
  ready set is computed from.
- `parent` — the bead belongs to an epic.
- `related` — relevant but non-blocking.
- `discovered-from` — work surfaced while implementing another bead. Preserve
  these; they're how emergent work stays attached to its origin.

Add only edges you can justify. A missing real edge is a dispatch-time race; a
fabricated edge needlessly serializes work the orchestrator could parallelize.

**A sequencing constraint you discover here is an edge, not a note.** WHEN one
bead consumes a contract another is changing — `c4_impacts` intersecting
another bead's `c4_changes`, or a stated prerequisite — wire
`bd dep add <consumer> <producer>`. An orchestrator dispatches from the
re-queried ready set, so an edge is honoured automatically, at any depth and
across epic boundaries; a line of prose in a preview is honoured only if
someone reads it, and nothing re-reads it after the first dispatch.

**`blocks` edges are type-segregated** — task↔task and epic↔epic are accepted,
task↔epic is refused in both directions (`tasks can only block other tasks, not
epics`). A constraint between a bead and an epic cannot be expressed this way;
record it on the epic and say so rather than leaving a failed command behind.

### 4. Assign a capability tier

Beads carry a **tier** (`light | standard | advanced | frontier`, least to most
capable), never a model name — the orchestrator resolves tier → model at
dispatch from the orchestrator's model lineup, so a bead groomed
today still routes correctly after the lineup changes.

```
bd update <id> --set-metadata tier=standard \
  --set-metadata tier_reason="localized CRUD + tests, low ambiguity" \
  --set-metadata est_size=S \
  --set-metadata execution_suggested_model=standard \
  --set-metadata execution_reasoning_effort=medium \
  --set-metadata execution_agent_type=worker
```

**The rubric is `references/shared/work-order-field-guide.md`'s.** Two rules
bind whatever it says:

- **Assign the LOWEST tier that can do the work.** Over-routing is invisible
  and paid for on every run; under-routing announces itself as an escalation
  the orchestrator reports back. Bias toward the cheaper call.
- **Always record `tier_reason`.** A bare tier cannot be audited, and the
  retro loop below needs it to recalibrate.

**The keys, why the `execution_*` ones are `bd`'s own convention rather than
ours, and how `--metadata` merges (shallowly) against `--set-metadata`'s
add-only, string-only writes:** `references/capability-tier.md`.

### 5. Tag C4 blast radius

If the repo documents architecture with the C4 model, record which
containers/components each bead touches — in metadata so it's queryable, not
buried in prose:

```
bd update <id> --metadata '{"tier":"standard","tier_reason":"...","est_size":"S",
  "c4_changes":["auth-service"],"c4_impacts":["web-spa","billing-service"],
  "c4_docs":["docs/c4/auth-service/component.md"]}'
```

These three fields are **arrays**, which is why this phase uses
`--metadata` and not `--set-metadata`: the latter would store
`["auth-service"]` as the literal string `"[\"auth-service\"]"`, and
anything reading `c4_changes` as a list — collision packing above
all — would not see a component name. `--metadata` merges at the top
level, so phase 4's tier keys survive without restating them.

- Distinguish **changes** (code this bead edits) from **impacts** (downstream
  components affected but not edited). Both matter; see below.
- Use a **controlled vocabulary**: read the actual C4 docs and only use component
  names that exist. If a bead targets a component with no doc, flag it in the
  report — that mismatch is either a stale doc or a misunderstanding, and both are
  worth surfacing.
- This pays off three ways: **dispatch isolation** (two beads that both *change*
  `auth-service` are a collision even when both are `bd ready` — dependency-free ≠
  safe-to-parallelize), **tier routing** (blast radius is a risk signal), and
  **doc freshness** (a bead that changes a component should carry an acceptance
  checkbox to update that component's C4 doc).

If the repo has no C4 docs, skip this phase and note it — don't invent an
architecture.

### 6. Validate

Before declaring the set ready, validate — don't eyeball it. Which tool depends
on whether `bd` is live:

- **With `bd` present**, run the health checks via `scripts/bd-health.sh`, which
  runs whatever of these exist and summarizes: `bd lint` (missing
  acceptance/repro/success criteria), `bd swarm validate` (DAG cycles, orphans,
  disconnected subgraphs per epic), `bd orphans`, `bd stale`, `bd doctor`.
  It exits 1 on hard problems (cycles, orphans) and 2 when no check could run
  at all. **Do not declare the set ready while it exits non-zero.** Fix what it reports and re-run; do not proceed while it fails. After two full failed cycles, stop and surface the blocker rather than continuing to retry.
  Read its summary line: when it reports that hard checks were *unavailable*
  in this `bd` build, graph integrity is unverified — say so in the report
  rather than treating a zero exit as a clean bill of health.
- **In the file-emit fallback**, run the bundled validator on what you wrote:

  ```
  python3 scripts/validate_beads.py <emitted.jsonl> --c4-docs <path/to/c4/docs...>
  ```

  It checks each bead against `references/bead.schema.json` **and** the things a
  per-record schema cannot: IDs resolve, no cycles, C4 names are in the
  vocabulary parsed from your docs, tier non-null on ready beads, ready
  features and bugs carry acceptance criteria. **Iterate emit → validate → fix
  until it reports 0 errors**; warnings are advisory. What each finding means:
  the script's own `--help`.

Either way: cycles, orphans and dangling references are hard problems — fix
them, don't report around them.

### 7. Report — and optionally preview the partition

Close the pass with a concise report:
- What you groomed and the notable changes per bead (especially descriptions you
  made context-independent, and beads you split).
- Anything **still failing** the Definition of Ready, and why — never pass a
  half-groomed bead downstream silently.
- Any flags for the human: invented-looking dependencies you declined to wire,
  C4 component mismatches, beads too ambiguous to size.

**Optional partition preview** (default on, non-authoritative): name every pair
of ready beads sharing a `c4_changes` component, and group each cluster into a
candidate batch in dependency order. Those pairs become **collision edges** in
the plan, so a pair never reaches the ready set together. The planner remains
the source of truth on packing; say explicitly that this is a preview.

## The retro loop (closing the circle)

Grooming quality compounds only if execution feeds back into it. Four signals a
retro can give you, and nothing else can:

- **Router calibration** — a bead whose `tier` was wrong is the signal to adjust
  the field guide's rubric. Note the asymmetry: escalations surface
  *under*-routing and nothing surfaces over-routing, so a large run with zero
  escalations is worth examining rather than celebrating.
- **Grooming misses** — beads reopened or blocked mid-run for missing context.
  Capture the *pattern* as `bd remember "<insight>"` so future grooming
  inherits it through `bd prime`, and propose it against the Definition of
  Ready, which the `spec` skill owns.
- **C4 drift** — components appearing in `c4_impacts` far more than expected: a
  coupling hotspot worth a refactor epic.
- **Discovered-work audit** — near-zero `discovered-from` across a big run is
  suspicious (agents hiding problems to save context); a flood from one epic
  means that epic was under-groomed.

Make the retro a tracked bead so it is never skipped.

## Anti-patterns to actively reject

- Session-dependent language ("as discussed", "the above", "finish the earlier
  refactor"). Hard fail — inline the context.
- An unquoted file:line citation — the position rots once an earlier order
  edits the file, leaving nothing to relocate by. Inline the content the line
  held.
- A `feature`/`task` with no `acceptance`, or a `bug` with no repro.
- Boulder beads masquerading as one unit of work → split into epic + children.
- Ordering that lives only in priority numbers, with no dependency edges.
- Two parallelizable beads that silently change the same C4 component.
- Model or C4 tags left "to figure out later" — later is inside an isolated agent
  that can't ask.
- Fabricated dependencies, IDs, or component names to make the graph look tidy.

## Reference files

Files under `references/shared/` are **generated copies** vendored in by
`scripts/sync-shared-references.sh` so this skill is self-contained. Read them
normally; never edit them — fix the source named in each file's header and
re-run the script.

- `references/shared/work-order-definition-of-ready.md` — the checklist you
  enforce in phase 2 (spec-owned; do not fork it).
- `references/shared/work-order-field-guide.md` — description vs. design vs.
  acceptance; the tier-routing rubric (spec-owned).
- `references/shared/work-order-template.md` — the work-order format this skill
  materializes into beads.
- `references/bead.schema.json` — authoritative JSON Schema for the beads
  materialization in the file-emit fallback. The single source of truth for
  bead field and edge names.
- `scripts/bd-health.sh` — composite `bd` graph/quality health check (phase 6, bd present).
- `scripts/validate_beads.py` — schema + graph/semantic validator for emitted
  JSONL (phase 6, file-emit fallback). Requires `python3`; uses `jsonschema` if
  installed and falls back to built-in checks otherwise.

## Retro this skill

WHEN a run of `groom-beads` struggled — you improvised around its
guidance, a command or script it prescribed misbehaved, or an
instruction here pointed you the wrong way — offer to run `retro-skill`
on `groom-beads` while the session is fresh, to capture the friction
as evidence-backed issues. This skill prescribes scripts
(`scripts/bd-health.sh`, `scripts/validate_beads.py`), so a misbehaving
check is a
prime `script-defect` finding. WHEN the friction was instead in what
"ready" means, that belongs against the `spec`-owned Definition of
Ready, not here. A clean run needs nothing; never run it unprompted.
