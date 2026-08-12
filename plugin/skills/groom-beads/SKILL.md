---
name: groom-beads
description: >-
  Materialize and groom work orders as beads (issues in the Beads / `bd`
  git-native issue tracker) so each is Ready — dispatchable to an isolated,
  zero-context agent. Use this skill whenever the user wants to groom,
  refine, triage, prepare, clean up, or "get beads ready" — or mentions a
  backlog of beads, a `.beads/` directory, `bd ready`, running beads in
  waves, assigning capability tiers to beads, or wiring bead dependencies
  before a long-running or orchestrated session. Trigger even if the user
  just says "groom these" or "make these beads ready" in a repo that uses
  beads.
version: "1.2"
---

# Bead Grooming

## What this skill is

`groom-beads` is the **work-order Definition of Ready engine for the beads
backend**. A *work order* is this framework's portable, atomic unit of
executable work — everything a fresh agent needs to complete one job, with no
planning conversation to lean on. This skill materializes work orders as
**beads** and drives each to the **Definition of Ready**, so an orchestrator can
pull ready beads, group them into **waves**, and dispatch each to an **isolated
sub-agent** with zero session context.

That execution model is the whole reason grooming matters. Every sub-agent is a
*fresh agent with zero context*. So the job of grooming is not "fill in fields"
— it is to **manufacture context-independence and correct ordering**, because
those are the two things the orchestrator actually consumes. A bead that reads
fine to someone who was in the planning session but is unworkable cold is the
central failure mode, and it is invisible until it detonates three waves deep.

The north star, applied to every bead you touch:

> **A bead is groomed when a fresh sub-agent, with zero session context, can open
> `bd show <id>` and complete the work correctly — right approach, verifiable
> done-ness — without asking anyone anything.**

Two things live elsewhere and are deliberately NOT restated here (canonical
copies live in the `spec` skill, so authoring and grooming can never disagree):

- **What "ready" means.** The criteria are owned by the `spec` skill. Load and
  enforce them; do not paraphrase or fork them:
  - `references/shared/work-order-definition-of-ready.md` — the checklist.
  - `references/shared/work-order-field-guide.md` — the three text fields kept
    distinct, and the tier-routing rubric.
  - `references/shared/work-order-template.md` — the work-order format this
    skill materializes into beads.
- **Where work orders come from.** A standalone bug/chore you groom directly, or
  a feature blueprint decomposed by `groom-feature-blueprint` — which hands its
  finishing pass (tier, C4, validation) to this skill.

What lives *here* is the beads-specific machinery: probing the `bd` CLI, mapping
the work-order format onto bead fields and metadata, wiring the edge graph,
validating the result, and previewing waves.

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
- Probe the flags you're about to use: `bd create --help`, `bd update --help`,
  `bd dep --help` (or `bd link`/`--deps`), `bd lint --help`. Field and flag names
  like `--design`, `--acceptance`, `--set-metadata`, `--append-notes`,
  `--deps discovered-from:<id>` are what current Beads uses, but adapt to what
  `--help` actually shows. If a capability is missing, degrade gracefully (e.g.
  fold acceptance criteria into the description under an `## Acceptance` heading)
  and note it in your report rather than failing.
- **When `bd` is unavailable and you fall back to emitting bead files,** emit
  JSONL that conforms to `references/bead.schema.json` — that schema is the
  authoritative source for field and edge names, so don't improvise alternatives
  (it's `dependencies` not `deps`, `id` not `slug`, and parentage is a dependency
  of type `parent`, never a top-level `parent` field). Validate what you emit with
  `scripts/validate_beads.py` and fix until it passes (phase 6).
- Load durable project knowledge so grooming matches house style: `bd prime` if
  available, and skim existing `bd remember` insights. Past retros are the memory
  of what previous grooming got wrong — honor them.

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

WHEN a bead cannot be brought to Ready, do not dispatch it. Parking it takes
**two** writes, and the second is the one that actually works:

1. Record WHY, for the human and the next grooming pass — set
   `grooming_status` to `blocked-needs-info` (a specific fact is missing, or
   the Design leaves a choice open — that is the human's call, never yours to
   guess) or `blocked-ungroomable` (it needs redesign).
2. Record THAT, for the orchestrator — change the bead's `status` so it
   leaves the ready queue:

   ```
   bd update <id> --set-metadata grooming_status=blocked-needs-info
   bd defer <id>          # status -> deferred; drops out of `bd ready`
   ```

`bd ready` reads `status`, **not** `grooming_status`. A bead flagged only
with `grooming_status` stays dispatchable and will be handed to a fresh
subagent exactly as if it were ready — which is the failure the flag exists
to prevent, made worse by looking handled. Verify the park took effect: the
bead SHALL NOT appear in `bd ready` afterwards.

Then flag it in the report rather than shipping a half-groomed bead
downstream silently.

### 3. Wire the dependency graph

The orchestrator reads ordering **from the graph, not from prose or priority
numbers**. Encode real relationships with the right edge type (these realize the
work order's `depends_on` / `blocks` / `parent_slug` / `related` /
`discovered_from` fields):

- `blocks` / `blocked-by` — genuine sequential dependencies. This is what defines
  wave boundaries via `bd ready`.
- `parent` — the bead belongs to an epic.
- `related` — relevant but non-blocking.
- `discovered-from` — work surfaced while implementing another bead. Preserve
  these; they're how emergent work stays attached to its origin.

Add only edges you can justify. A missing real edge is a wave-time race; a
fabricated edge needlessly serializes work the orchestrator could parallelize.

### 4. Assign a capability tier

Record the tier in **metadata** so it rides with the bead and the orchestrator
can read it at dispatch:

```
bd update <id> --set-metadata tier=standard \
  --set-metadata tier_reason="localized CRUD + tests, low ambiguity" \
  --set-metadata est_size=S
```

Two flags write metadata and they are not interchangeable — probe both
before a bulk pass:

- `--set-metadata key=value` is **repeatable and additive**: it leaves
  other keys intact, so it is what you want for touching one field on an
  already-groomed bead. It stores every value as a **string**.
- `--metadata '<json>'` takes real JSON types but **replaces the entire
  metadata object**. Use it to write a bead's metadata in one shot (at
  creation); do NOT reach for it to add a key, or you will silently drop
  every key you did not restate.

Beads carry a **tier** (`light | standard | advanced | frontier`, least to
most capable), not a model name. The orchestrator resolves tier → model at
dispatch from the project's `## model lineup` block, so a bead groomed today
still routes correctly after the lineup changes. The routing rubric is owned
by `references/shared/work-order-field-guide.md`. Apply it, and:

- **Assign the LOWEST tier that can do the work.** Over-routing is invisible
  and paid for on every run; under-routing announces itself as an escalation
  the orchestrator reports back. Bias toward the cheaper call.
- **Always record `tier_reason`** — a bare tier can't be audited, and the
  retro uses the reason to tell whether routing was calibrated.
- **Assign a default, never null** — a mis-groomed bead must not stall a wave.
  WHEN genuinely unsure, `standard` and say so.
- Set a named `model` **only** as an override, when a specific model is
  genuinely required. Normal beads carry `tier` and no `model`.

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
anything reading `c4_changes` as a list — wave-collision packing above
all — would not see a component name. Because `--metadata` replaces the
whole object, restate the tier keys phase 4 already wrote, as shown.

- Distinguish **changes** (code this bead edits) from **impacts** (downstream
  components affected but not edited). Both matter; see below.
- Use a **controlled vocabulary**: read the actual C4 docs and only use component
  names that exist. If a bead targets a component with no doc, flag it in the
  report — that mismatch is either a stale doc or a misunderstanding, and both are
  worth surfacing.
- This pays off three ways: **wave isolation** (two beads that both *change*
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

  Requires `python3`. It uses the `jsonschema` package for full structural
  validation when installed and falls back to equivalent built-in checks
  otherwise, so it never fails for a missing dependency.

  It checks each bead against `references/bead.schema.json` (canonical field/edge
  names, enums, required fields) **and** the things a per-record schema can't:
  dependency/parent IDs resolve, no cycles in the blocks graph, C4 names are in
  the vocabulary parsed from your docs, tier is non-null on ready beads, ready
  features/bugs carry acceptance criteria, plus a session-phrase lint and a
  wave-collision advisory. **Iterate emit → validate → fix until it reports 0
  errors** (warnings are advisory — a greenfield C4 component or a deliberately
  `blocked-ungroomable` bead will warn, and that's fine).

Either way: cycles and orphans and dangling references are hard problems — fix
them, don't report around them.

### 7. Report — and optionally preview waves

Close the pass with a concise report:
- What you groomed and the notable changes per bead (especially descriptions you
  made context-independent, and beads you split).
- Anything **still failing** the Definition of Ready, and why — never pass a
  half-groomed bead downstream silently.
- Any flags for the human: invented-looking dependencies you declined to wire,
  C4 component mismatches, beads too ambiguous to size.

**Optional wave preview** (default on, non-authoritative): group the ready beads
into candidate waves under *two* constraints — dependencies **and** no two beads
sharing a `c4_changes` component in the same wave — and present it as a
suggestion for the orchestrator to accept or ignore. The orchestrator remains
the source of truth on wave-packing; this just hands it the C4-collision insight
for free. Say explicitly that it's a preview.

## The retro loop (closing the circle)

Grooming quality compounds only if execution feeds back into it. When the user
runs a retro (or asks you to help with one), the highest-value outputs are:
- **Router calibration** — beads whose assigned `tier` was wrong are the signal
  to adjust the field guide's routing rubric. The orchestrator reports every
  escalation (bead, assigned tier, tier that completed it, what failed); those
  are the run's only direct routing evidence. Note the asymmetry: escalations
  surface under-routing, but nothing surfaces over-routing — so a large run
  with zero escalations is worth examining, not celebrating.
- **Grooming misses** — beads reopened or blocked mid-wave for missing context.
  Capture the *pattern* as `bd remember "<insight>"` so future grooming (via
  `bd prime`) inherits it, and ideally propose it as an RFC against the
  Definition of Ready (which the `spec` skill owns).
- **C4 drift** — changes that shipped without a doc update; components appearing
  in `c4_impacts` far more than expected (a coupling hotspot worth a refactor
  epic).
- **Discovered-work audit** — near-zero `discovered-from` across a big run is
  suspicious (agents hiding problems to save context); a flood from one epic
  means that epic was under-groomed.

Prefer making the retro itself a tracked bead (type `chore`, or a `retro` type)
so it's never skipped.

## Anti-patterns to actively reject

- Session-dependent language ("as discussed", "the above", "finish the earlier
  refactor"). Hard fail — inline the context.
- A `feature`/`task` with no `acceptance`, or a `bug` with no repro.
- Boulder beads masquerading as one unit of work → split into epic + children.
- Ordering that lives only in priority numbers, with no dependency edges.
- Two parallelizable beads that silently change the same C4 component.
- Model or C4 tags left "to figure out later" — later is inside an isolated agent
  that can't ask.
- Fabricated dependencies, IDs, or component names to make the graph look tidy.

## Reference files

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
`scripts/sync-shared-references.sh` so this skill is self-contained. Read
them normally; never edit them. Fix the source named in each file's header
and re-run the script.


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
as an evidence-backed RFC. This skill prescribes scripts
(`scripts/bd-health.sh`, `scripts/validate_beads.py`), so a misbehaving
check is a
prime `script-defect` finding. WHEN the friction was instead in what
"ready" means, that belongs against the `spec`-owned Definition of
Ready, not here. A clean run needs nothing; never run it unprompted.
