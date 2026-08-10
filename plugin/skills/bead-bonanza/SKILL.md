---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use
  when the user asks to work on multiple beads at once — e.g. "work on all open
  beads with label X", "address all P1 beads", "clear the bead backlog". Triages
  the queue, gates every bead for readiness (grooming under-specified beads via a
  subagent, or bouncing what can't be made ready), surfaces both explicit and
  implicit decisions up front, dispatches via the cheapest viable execution mode
  (doc fast lane / shared-surface batch / worktree parallel / sequential), streams
  per-bead progress, and lands each bead in a single commit.
version: "1.3"
---

# bead-bonanza

Batch-process beads via subagent orchestration. The orchestrator (you) stays
small — you only hold the queue, the user's decisions, the dispatch plan, and
one-line outcomes per bead. Each bead runs in a fresh subagent so context
doesn't compound.

Read `CLAUDE.md` and `AGENTS.md` once at the start so you have repo conventions
in scope before delegating.

**After any compaction, before dispatching anything:** re-read this file from
Phase 3 onward and re-read references/recovery.md. A compacted session re-injects
only this file's first ~5,000 tokens, so the per-bead ceremony, the published
script interfaces and the rule against paraphrasing a dispatch template are NOT
in your context even though you have "read the skill". Then ls "$RUN_DIR" and
re-run the queue query, to re-derive run state from the durable source rather
than from a summary of it.

**This skill is language-agnostic.** It never hard-codes build, test, lint, or
format commands. Everything it needs to validate the work done for a bead comes
from the project's **gate profile**, echoed by Phase 0's pre-flight script.
Throughout this document, `<format>`, `<lint>`, `<build>`, `<test>`, and
`<regen>` are placeholders that
resolve to that project's actual commands. The profile can also define **custom
gates** — any extra command or inline script the project wants run as part of
the gate suite (e.g. a schema check, a license-header linter, a `bash` block).
Substitute the profile's values wherever you see a placeholder, and run any
custom gates alongside them.

## Run checklist

Copy this into your response at run start and check items off as the run
proceeds. It is the one artifact that survives a long autonomous run.
(A checklist copied into your response is not a task tool — WHERE the project
bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress
visible.)

```
bead-bonanza run:
- [ ] 0. Pre-flight: pre-flight-check.sh run, RUN_DIR + SCRIPTS eval'd
- [ ] 1. Queue resolved and printed (count + ordered list)
- [ ] 1.5 Readiness gate: every bead groomed (decisions closed), or bounced
        with its gap
- [ ] 2. Dispatch plan: bucket + tier + push mode per bead
- [ ] 3. Execute buckets, verifying each return (skip claims, clean tree)
- [ ] 4. Final report: table, totals, implicit decisions, TIER ESCALATIONS,
        follow-up beads, stashes, pushes, run-dir path
```

## Phase 0 — Pre-flight

One script measures the environment; you act on what it returns. Run it from
this skill's own directory, with a run id of `$(date +%s)-<queue size>`:

```bash
eval "$(<this skill>/scripts/pre-flight-check.sh "<run-id>" --gate-baseline)"
```

Read its stderr report once. It echoes the two `AGENTS.md` blocks —
`## bead-bonanza gates`, the profile every `<format>`, `<lint>`, `<build>`,
`<test>` and `<regen>` placeholder in this document resolves to, and
`## model lineup`, how a bead's tier resolves to a model at dispatch — and
reports every other probe, including the gate baseline. Its stdout sets
`RUN_DIR`, `SCRIPTS` and the `BB_*` facts; substitute those wherever this
document shows a run-dir path or a scripts path. Exit **0** proceed;
**1** blocking, named in the report and fixed before Phase 1 — a missing
`AGENTS.md` block means read `references/project-setup.md` and draft it WITH
the user, never invent gates or a lineup from memory; **2** you called it
wrong.

Those facts are ledger lines with a command behind them. **Do not re-assert
from memory what the script just measured** — push mechanics in particular
vary by container. WHEN `BB_PUSH_NEEDS_REFSPEC=yes`, push with an explicit
refspec (`git push <remote> <branch>`) every time; verify landings with
`git rev-parse <remote>/<branch>` after fetch, not `git status -sb`; and never
let a pipe to `tail` mask a push's exit status.

Two things no script can settle:

- **Draft PR.** WHEN `BB_PR_CI=yes`, ask the user once whether to open a draft
  PR from the working branch — every wave push then gets the full CI matrix in
  the cloud, including toolchain-heavy suites the local sandbox can't run,
  which is what makes the slim local gate policy (below) safe. (`gh` fails TLS
  verification inside the sandbox; run it with sandbox disabled.) WHEN
  `BB_PR_CI=no`, don't ask.
- **Agent-tool availability.** Whether you have an Agent/Task tool at all is a
  harness property, invisible from a shell, and it is the trigger for the
  main-loop fallback. No script can check it — you must.

WHEN the run may use worktree-parallel, confirm `BB_WORKTREE_BASEREF` is `head`
before Phase 2.5. Unset, worktrees branch from the **default branch**, not
session HEAD, and beads building on each other lose their foundation. It is
user-scope global state this repo does not own, so re-verify per wave — see
`references/subagent-prompts.md`, "The base-ref trap".

## Phase 1 — Resolve the queue

Parse the user's filter from their request (label, priority, ID list).

1. Run `bd ready` with the filter — this gives dependency-aware ordering.
   Examples:
   - `bd ready --label <label>`
   - `bd ready --priority 1`
   - `bd ready -n 50` (raise the limit if needed)
2. If `bd ready` doesn't accept the filter, fall back to `bd list --status open
   --json` and filter the JSON yourself.
3. Only include beads whose status is `open`. Skip `in_progress` (may be claimed
   elsewhere), `closed`, `blocked`.
4. Print the queue to the user: count + ordered list of `id | P# | title`.
5. If the user gave no filter, ask once: "Which beads? (label / priority /
   explicit ID list)" — this is the only Phase-1 question allowed.

Use `TaskCreate` to track one task per bead so progress is visible (unless the
pre-flight reported `BB_TASKCREATE=banned`).

## Phase 1.5 — Readiness gate (grooming)

bead-bonanza dispatches **only groomed beads.** Every later phase already
assumes grooming's outputs: Phase 1's `bd ready` ordering comes from the
dependency graph grooming wires; this gate's own cheap check reads the
`decision-required` label grooming sets (DoR check 9); Phase 2.5 reads the
`tier:` assignment grooming makes. An ungroomed queue makes all three
unreliable — and handing a fresh, zero-context subagent an under-specified
bead is the single most expensive failure mode this skill has: it detonates
mid-wave, after later beads have built on it.

So before dispatch, gate the queue against the **Definition of
Ready** — the checklist owned by the `spec` skill
(`references/shared/work-order-definition-of-ready.md`) and enforced by
`groom-beads`, which materializes work orders as beads. Those skills are the
authority; do not restate the checklist here.

### 1. Cheap check (orchestrator, read-only)

Most runs pass in seconds because the user groomed first. Per bead, confirm the
readiness signals (parallelize — this is read-only):

- `bd lint` is clean for it (no missing acceptance / repro / success criteria),
  if the installed `bd` has it.
- a `feature`/`bug` carries a `design` and `acceptance` criteria; a `bug`
  carries a repro.
- a `tier` is assigned (grooming never leaves it null).
- dependencies are wired (ordering must live in the graph, not in priority
  numbers or prose).
- no session-dependent language — "as discussed", "the above", "finish the
  earlier refactor" — the fresh-subagent test.
- no open decision — no `decision-required` label, "DECISION REQUIRED"/"TBD"
  marker, or unresolved options list (DoR check 9); it is bounced, not
  questioned mid-run.

A queue that is entirely ready makes this phase a one-line no-op: say so and
proceed to Phase 2.5.

### 2. Groom what fails (delegate — don't groom in the orchestrator)

If any bead fails the check, do NOT dispatch it and do NOT fix it inline —
grooming reads source and wires graphs, which belongs in a subagent (see "don't
read source files in the orchestrator"). Dispatch a **grooming subagent** over
the ungroomed subset (template below); it runs `groom-beads` to the Definition
of Ready and returns a per-bead readiness report plus a **wave preview** — its
candidate partition under dependencies and C4 collisions. Carry the preview into
Phase 2.5, which takes it as a starting partition, not an instruction.

When it returns, **re-run `bd ready`** with the run's filter: grooming may have
wired new dependencies or split a boulder into an epic + children, changing both
the ordering and the membership of the queue.

### 3. Bounce what grooming can't clear

Grooming, like this skill, bounces gaps rather than guessing. A bead it cannot
bring to Ready — needs a human decision it can't make, can't be sized without a
split, rests on an unexamined premise — does **not** enter the run. Exclude
these from the queue and surface them to the user at the end of this phase (one
line each + the gap). The run proceeds with the Ready beads; ungroomable beads
are the user's to resolve, not a subagent's.

Bounced beads are the run's **single decision point**: surfaced with their
open question, resolved or groomed before dispatch, then the run proceeds
with the Ready remainder — an *admission* decision made once, up front, not
mid-flight.

### 4. Admitting emergent work mid-run

During execution (Phase 3) subagents discover work. Some of it belongs in this
run; some is scope creep wearing a matching label. Telling them apart is the
point — pulling in the wrong thing is as real a failure as leaving a blocker
unaddressed. This is where the autonomy the run was launched for lives, and
where it stops.

**The test is necessity, not resemblance:** emergent work is in-scope only when
a bead *already in the run* cannot be correctly and completely finished without
it. Three shapes qualify:

- **an unblocking prerequisite** — the assigned bead genuinely cannot be
  completed until this is done;
- **an acceptance-criteria decomposition** — a piece the assigned bead's own AC
  cannot be satisfied without (a sign the bead was under-groomed, larger than it
  looked);
- **an expected test-coverage gap** — coverage the assigned scope already
  implied should exist and doesn't.

Everything else — adjacent cleanups, refactors "while I'm here", a new
capability no assigned bead's AC demands — is **out of scope even if it matches
the run's filter.** Matching the label is necessary but not sufficient;
"required to finish a bead already in the run" is the bar.

**Mechanics.** The discovering subagent only *files* the bead (`bd create`) with
a `discovered-from` edge to its origin and a one-line note of which shape it is
(or "out-of-scope"); it never implements it. If its own bead is truly blocked by
it, it takes the Blocked path naming the new id. The **orchestrator** then
decides:

- **In-scope** → run it through this gate (groom → re-`bd ready`), wire the
  `blocks` edge so it orders ahead of the bead it unblocks, and dispatch it
  under the standing auto-chain authorization. No new prompt — this is the
  intended autonomy. Grooming closes decisions along with the rest of the
  Definition of Ready. (A groomed prerequisite that still lands a
  `decision-required` label stops for the user, per the standing-authorization
  terms.)
- **Out-of-scope** → leave it filed and linked, do NOT dispatch it, and surface
  it in the final report for the user to triage. The standing authorization
  finishes the launched scope; it never enlarges it.

**Ballooning is a stop signal.** If "prerequisites" multiply — a discovered
blocker that is really a new epic, or more than a handful of new in-scope beads
off a single bead — stop and surface it. That means the original scope was
mis-groomed, which is a human's call, not an autonomous expansion.

The grooming subagent prompt template is in
`references/subagent-prompts.md`.

## Phase 2.5 — Triage & dispatch plan

**Bias toward batching.** Each "plain sequential" bead burns a full `<test>` run
(often 10+ min). Batching N beads in one subagent saves N−1 of those. The
skill's job is to find the largest safe batch, not to default to
one-bead-per-subagent for caution. Sequential is the *exception*, not the rule.

### Classify each bead

- **`trivial`** — pure documentation (md/mdx edits), single-line code swap,
  comment/regex-pin additions. No fixture regeneration needed; targeted unit
  tests suffice. Examples: "document a trade-off in docs/", "add an explanatory
  comment", "swap one equivalent stdlib call for another".
- **`isolated`** — shares no C4 component and no file with another queued bead.
  No `BLOCKS` / `RELATED` link to another queued bead.
- **`shared-surface`** — shares a `c4_changes` component with another queued
  bead, or touches files (especially snapshot/golden fixtures) another queued
  bead also touches, or has explicit `RELATED` links to another queued bead.

**Overlap comes from C4 metadata first, file paths second.** Grooming already
computed the blast radius onto each bead — `c4_changes` (components it edits)
and `c4_impacts` (downstream components affected but not edited). Read it; don't
re-derive it from prose. The `jq` and a worked example are in
`references/dispatch-plan.md`.

1. **C4 metadata — primary.** Read `.metadata.c4_changes` and
   `.metadata.c4_impacts` from `bd list --json` for every queued bead.
   - Two beads sharing a `c4_changes` component are **shared-surface** whatever
     their file paths say — the same component through different files still
     collides. They MAY batch; they SHALL NOT share a worktree-parallel wave.
   - Bead B whose `c4_impacts` intersects bead A's `c4_changes` consumes a
     contract A is changing. That is a **sequencing** constraint, not a batching
     one, and it holds with zero file overlap: order B after A within the
     bucket, or wire a real `blocks` edge when the ordering must survive the run.
   - Disjoint changes with no impact intersection is isolated *architecturally*
     only — layer 2 still runs.
2. **File paths — fallback and corroborator.** Scan each bead's description for
   fixture names + source paths. It is the whole signal where `c4_*` is absent,
   and it runs alongside C4 where both exist. WHEN the two disagree — disjoint
   components, overlapping files — take the **union** as the collision set and
   note the discrepancy in the plan; do not silently prefer one signal.
3. **The grooming wave preview — starting partition.** WHEN Phase 1.5 returned
   one, start from it instead of re-deriving the partition, and record in the
   plan whether you accepted or overrode it and why. The orchestrator stays the
   source of truth on packing; the preview is an input, not an instruction.

Where the repo documents no C4 architecture, grooming skips the tagging and
`c4_*` is absent — file scanning is then the whole signal, and absent metadata
means *unknown*, never "no collisions".

The two collision classes read differently in the plan's `Reason` column: a
co-edit collision names what is shared (`same c4_changes: auth-service`); a
sequencing constraint names what is consumed and from whom (`consumes
auth-service, changed by demo-9h0`).

### Choose a dispatch mode per bucket

Four modes. Pick the broadest one that's safe — you're trying to minimize total
full-test-suite runs across the queue.

1. **Doc/chore fast lane** — `trivial` beads. ONE subagent processes the cluster
   sequentially. Targeted tests between beads, ONE full `<test>` run at the end.
   One commit per bead. Cap ~6 beads per fast-lane subagent.

2. **Shared-surface batch** — `shared-surface` beads that cluster on the same
   fixture/code area. ONE subagent processes the cluster sequentially. Between
   beads: `<format>`, `<lint>`, `<build>`, **plus surface-relevant targeted
   tests (non-negotiable, see Phase 3)**. ONE full `<test>` run at the END of
   the batch. One commit per bead. Saves N–1 full test-suite runs.

   **Threshold:** use shared-surface batch when **≥2** beads share a surface and
   decisions are independent (bead A's choice doesn't change bead B's
   implementation). Even at 2 beads the savings are real — one fewer full test
   run — and the bisect cost is one suspect commit.

   **Don't over-split:** if 7 beads cluster on one surface and only 1–2 of them
   are "larger" (substantial new features), they can still go in one shared
   batch. Splitting buys you nothing if decisions don't conflict. Only split
   when bead A's implementation would conflict with bead B's, or when decisions
   overlap such that one bead's choice changes another bead's plan.

3. **Worktree parallel** — `isolated` beads with no surface overlap, where
   parallel wall-clock savings beat merge cost. Spawn each via `Agent` with
   `isolation: "worktree"`. `bd` is not the obstacle here: it resolves the
   shared tracker database through git's common directory from any worktree,
   measured with `BEADS_DIR` unset (`references/subagent-prompts.md`, "Beads
   inside a worktree"). The obstacles are the exported ledger and the base
   ref. Each worktree agent SHALL run `bd export > .beads/issues.jsonl` in its
   own directory before committing — auto-export is best-effort and always
   writes the MAIN checkout's copy, so a worktree's committed ledger goes
   stale while `git status` reads clean (same reference, measured). Because
   `export.git-add: true` (`.beads/config.yaml`) stages those writes into the
   main index, keep commits pathspec-scoped (see "Guard against foreign dirt"
   above). A wave of 2+ worktree beads therefore commits near-identical full
   exports of one shared database with no merge driver configured for
   `.beads/issues.jsonl` — an N-way conflict there is guaranteed, not
   incidental; resolve it by re-exporting on the merged tree
   (`bd export > .beads/issues.jsonl`, stage, amend), not by hand-merging JSON
   and not via `<regen>` (that's for fixture/snapshot conflicts only, next).
   Orchestrator merges results back to the parent branch one at a time. If a
   merge instead produces fixture/snapshot conflicts, run `<regen>` on the
   merged tree to settle (skip if the project defines no regen step — resolve
   the conflict manually). Skip this mode if there are <3 isolated beads
   (parallelism savings don't pay), if the beads aren't actually isolated
   (you'll just hit merge conflicts), or if any bead builds on another in-run
   bead's work while `BB_WORKTREE_BASEREF` isn't confirmed `head` (Phase 0
   gate above; "The base-ref trap", `references/subagent-prompts.md`) — safe
   for that case is exactly what `agent-plugins-jh6` measured, not the
   default.

4. **Plain sequential** — reserved for: a single bead with no batch peers, OR a
   bead whose surface diverges from every other queued bead, OR explicit user
   override. Not the default. If you find yourself defaulting here for >1 bead
   in a row, reconsider — they probably belong in a shared batch.

### Tier selection per bead

Beads carry a **capability tier** from the grooming phase — the Phase 1.5
gate guarantees one (grooming assigns a default, never null): a
machine-readable label `tier:<name>` and/or a `Tier:` line in the bead's
notes (visible in `bd show`). The tiers are ordered, least to most
capable:

```
light  →  standard  →  advanced  →  frontier
```

Beads name a **tier**, not a model, so a backlog groomed months ago still
routes correctly. Resolve tier → model from the project's `## model
lineup` block in `AGENTS.md`, echoed by Phase 0. Honor these:

- **Collect during Phase 1.5** — you're already reading each bead's
  readiness signals; record its tier then. `bd list --label tier:standard
  --json` also works for bulk queries.
- **Dispatch at the resolved model** — pass `model: "<resolved>"` on the
  `Agent` tool call for that bead's subagent. Absent a lineup, omit the
  param (the subagent inherits the session model) and say so once.
- **A named `model` on a bead is an override** — grooming sets it only
  when a specific model is genuinely required. It wins over the tier.
- **Main-loop beads** — a bead at the orchestrator's own tier with a
  note saying "do not delegate" is implemented by the orchestrator
  directly, not dispatched. It still follows the full per-bead ceremony
  (claim, gates, retro, atomic close, single commit).
- **Batching across tiers** — prefer bucketing beads that share a tier.
  When a bucket mixes tiers, run the whole batch subagent at the
  **highest** tier among its members — never run a bead below its tier.
  If honoring that would drag several cheap beads up, split the bucket.
- **Escalate, never downgrade** — if a bead fails gates/review twice or
  its subagent visibly flails, re-dispatch that bead one tier up. Never
  re-dispatch below the assigned tier, and never "save cost" by
  downgrading mid-run.
- **Record every escalation** — the bead, the assigned tier, the tier
  that completed it, and what failed. This is a required line in the
  Phase 4 report (see "Tier escalations"), and it is the run's only
  direct evidence about whether grooming's routing is calibrated.
- **Record the routing** — the dispatch plan's wave table carries a
  `Tier` column for every bead.

### WHEN delegation is unavailable — the main-loop fallback

Everything above assumes you can spawn subagents. **Check that before
Phase 3, not during it:** some harnesses expose no `Agent` tool at all,
and some sessions forbid it. The check is simply whether an `Agent` tool
is available to you in this session.

WHEN it is not, the run does NOT abort — it degrades to a defined mode:

- **Implement every bead in the orchestrator's own loop**, in queue
  order, following the full per-bead ceremony unchanged (claim, gates,
  retro, atomic close, single commit). The ceremony is what makes a bead
  landing correct; delegation is only how the work is distributed.
- **Drop the parallel dispatch modes.** Worktree-parallel and
  shared-surface batching exist to isolate concurrent agents; with one
  agent there is no concurrency to isolate, so they add risk and no
  benefit. Run sequential.
- **Context is now the scarce resource**, because nothing is being
  isolated for you. Keep per-bead notes to one line, and re-read the run
  checklist rather than holding it in your head.
- **No tier routing occurs.** You cannot dispatch at a resolved model
  when you are the one doing the work — every bead runs at the session's
  own model regardless of its `tier`. Do not silently treat this as
  routing that happened to succeed.
- **Say so once, at the top of the run**, so the user knows the tier
  labels were not honored.

This is a real degradation, not an equivalent path: an under-tiered bead
that a `frontier` subagent would have handled may now fail, and a
`light` bead costs full session model. Both are acceptable; concealing
them is not.

### Push policy

- **`push-per-bead`** — subagent pushes after each commit. Best for sequential
  mode (recovery is easier on a per-bead failure).
- **`defer-push`** — subagent commits per bead, orchestrator pushes once at
  bucket end. Default for fast-lane and shared-surface batch (one process;
  failures are observable).

For a single-bead "wave," `push-per-bead` and `defer-push` are equivalent —
don't bother labeling. Just push when the subagent returns.

### The dispatch plan

Write `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md` **on every
run** — there is no queue-size or bucket-count exemption. It is the sprint plan,
written for whoever picks the sprint back up, which may be an agent with no
memory of having written it. Chat is the one store that does not survive.

Its format — run manifest, wave table, amendments — is in
`references/dispatch-plan.md`. Three of its rules are load-bearing enough to
state here:

- **The plan names beads, never a filter.** The filter produced the queue; the
  plan IS the queue. A filter re-evaluated later — at recovery, or at the next
  wave — returns a different set than the one this run committed to, and admits
  beads nobody scoped. Beads discovered mid-run enter only through Phase 1.5
  step 4's scope test, recorded as an amendment.
- **The manifest carries what recovery cannot re-derive** — run id, `RUN_DIR`,
  branch, base sha, gate profile, gate baseline, environmental ledger, resolved
  model lineup, and the standing authorization the user gave. Most of it is
  pre-flight's stdout block, landing durably.
- **The plan is append-only and amendable.** WHEN it changes mid-run — a wave
  resequenced, a bucket reclassified, a bead admitted or bounced, a tier
  escalated — append a dated entry under `## Amendments` stating what changed
  and why, AND update the wave table in place. The plan on disk SHALL describe
  the run as it is actually being executed. Mirror one line of each amendment to
  the tracker note carrying the `RUN_DIR` pointer, so a reader who has only the
  tracker knows the plan moved.

Print the wave table to chat as well. Don't ask for approval — just print and
proceed. The user can interrupt if they disagree.

## Phase 3 — Execute

Execute buckets in this order (lowest blast radius first):

1. **Doc/chore fast lane** — single subagent.
2. **Shared-surface batches** — one subagent per batch, sequentially across
   batches.
3. **Worktree parallel waves** — spawn isolated beads concurrently in one
   message; merge each as it returns.
4. **Plain sequential** — one subagent per bead, in `bd ready` order.

After each subagent returns:

1. Append its summary lines to your log (one per bead processed).
2. Mark corresponding tasks complete via `TaskUpdate` (or skip if the project
   reported `BB_TASKCREATE=banned`).
3. **Verify skip claims with a diff-stat.** For every bead the subagent returned
   as `skipped`:
   - Run `scripts/bb-verify-skip.sh` (scoped to that bead's worktree if
     applicable). It exits 0 when the skip is consistent, 1 when it contradicts
     the tree.
   - **If it exits 0:** the skip is consistent. Confirm the subagent's summary
     line includes a one-line reason citing the bead/commit that already
     addressed the issue (e.g. `skipped - already addressed by demo-wxi @
     a1b2c3d`). If the reason is missing, push back: re-prompt the subagent (or
     fill it in from `bd show <id>` history) before accepting the skip.
   - **If it exits 1 AND status=skipped:** this is an inconsistency. **Block the
     close.** The subagent either did work it didn't acknowledge, or left the
     tree dirty. Re-dispatch with explicit instructions to choose one of:
     - Commit the diff and change status to `completed` (run the full per-bead
       close ceremony), OR
     - Revert the diff (`git checkout -- <paths>` / `git restore --staged
       <paths>`) and re-confirm `skipped` with a fresh run showing CLEAN.
   - Do NOT proceed to push or to the next bead until the skip is reconciled.
     This catches the case where the subagent's first read mistakenly concluded
     "already addressed" but it still produced edits — the orchestrator must not
     silently accept a contradictory state.
4. **Verify the working tree is clean** with `git status --porcelain`. If dirty,
   fall through to the recovery procedure below.
5. If the bucket used `defer-push`, run `git push` for the bucket now.
6. Move to the next bucket/bead.

**Circuit breaker:** if 3 consecutive beads return `blocked` or fail, stop the
run and surface to the user. Something is systemically wrong (broken main,
environment issue, an under-groomed bead).

### Recovery and progress streaming

Three things live in `references/recovery.md`: the progress-event schema
every subagent writes (and the rule bounding what it proves), the procedure
for a subagent that dies mid-ceremony, and the entry procedure for resuming a
run after the *orchestrator* died. **Read it WHEN** a subagent times out,
returns an API error instead of a summary, or is expected to run past 5
minutes (any non-trivial bead, given full test runs) — at that point the
orchestrator must arm a `Monitor` on the progress file so the user sees
per-phase events live — **and WHEN** you are picking up a run whose state you
do not hold.

The one rule to carry without reading the file: after `gates-passed` and
BEFORE closing the bead, the subagent writes `$RUN_DIR/<bead-id>.handoff.md`
— the paths the close must stage, plus any close-specific caveat, and nothing
else. That file is the recovery anchor, and it is non-negotiable — without it,
recovery is guesswork instead of a procedure. Findings do not live there; they
are `retro:` comments on the bead.

### Subagent prompt templates

The four dispatch prompts — per-bead (sequential/worktree), doc/chore
fast lane, shared-surface batch, plus the orchestrator's worktree-parallel
merge notes — are in `references/subagent-prompts.md`. **Read that file at
dispatch** and fill in every `{PLACEHOLDER}` before sending. Do not
paraphrase a template from memory: the per-bead ceremony is ordered, and
the ordering is what makes the crash-recovery path mechanical.

## Phase 4 — Final report

After the queue is drained (or the circuit breaker tripped), print:

1. A markdown table: `| ID | Status | Bucket | Commit | Notes |`
2. Totals: `N closed, N blocked, N skipped`.
3. **Implicit decisions made** — extract from the beads' `retro:` comments: which beads had
   implicit decisions, what the subagent chose. Surface to the user so they can
   audit non-obvious choices.
3a. **Tier escalations** — every bead whose tier you had to raise to get it
    done, one row each:

   ```
   | Bead | Assigned | Completed at | What failed at the assigned tier |
   |------|----------|--------------|----------------------------------|
   | demo-9h0 | light | standard | fixture regen wrong twice; gates failed both cycles |
   ```

   State `none` explicitly when there were no escalations — the null is
   the result, not the absence of one.

   **`none` and "no routing occurred" are different results, and this
   table cannot tell them apart.** WHEN the run used the main-loop
   fallback because delegation was unavailable, do NOT write `none`.
   Write, in place of the table:

   ```
   Tier escalations: N/A — no routing occurred. No Agent tool was
   available, so every bead ran in the orchestrator's own loop at the
   session model, and no bead's tier was honored.
   ```

   An empty table read as `none` says "routing was tested and nothing
   needed raising", which is the opposite of what happened: routing was
   never exercised at all. Reporting the first as the second feeds
   `groom-beads`' retro loop a calibration signal that does not exist.

   This table is the whole point of routing by tier: it is the only
   direct evidence a run produces about whether grooming predicted
   capability correctly, and it is what `groom-beads`' retro loop
   consumes to recalibrate the rubric.

   **Say this alongside the table**, because the reader will otherwise
   draw the wrong conclusion from a short one: escalations reveal
   *under*-routing only. A bead groomed `frontier` that a `light` model
   would have finished never appears here — it succeeds, and the
   overspend is silent. So zero escalations across a large queue is not
   evidence of good calibration; it is equally consistent with routing
   everything too high. WHEN a run of ≥10 beads produces zero
   escalations, say so and suggest the next run try one tier lower on
   the beads whose `tier_reason` was thinnest.
4. **Follow-up beads** — beads created mid-run via `bd create`, split into those
   admitted into the run (in-scope prerequisites / AC-decompositions / coverage
   gaps, with their outcome) and those left for the user (out-of-scope finds,
   filed with a `discovered-from` link, awaiting triage).
5. **Stashes** — if any were created during recovery, list with `git stash
   list`.
6. **Pushes** — confirm all closed beads' commits are on the remote.
7. Path to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/` so the user can
   review the dispatch plan and progress files.

Do NOT create a PR or merge to main — that is a separate, explicit user action.

## Operating principles

- **Groom before you dispatch.** bead-bonanza executes *groomed* beads; it never
  fixes an under-specified bead inline. Phase 1.5 gates the queue against
  `groom-beads`'s Definition of Ready, delegates grooming of the rest to a
  subagent, and bounces what can't be made ready. This holds for beads pulled
  into scope mid-run too — a fresh subagent handed an ungroomed bead is this
  skill's most expensive failure mode.
- **Bias toward batching, not sequential.** Each plain-sequential bead burns a
  full `<test>` run. Find the largest safe batch — surface-overlap is the
  dominant signal, not bead complexity. Sequential is reserved for genuine
  surface divergence or single-bead queues, not "I'm being cautious."
- **The push is the gate.** Commits inside a branch are recoverable; only the
  push to origin is committed-to-the-world. Per-bead validation is
  regression-detection (catch issues before later beads build on them); the full
  `<test>` run happens once before the push.
- **Targeted tests per bead are non-negotiable.** "Reduced gates" between beads
  in a batch never means "no tests." Each bead must run surface-relevant
  targeted tests (from the gate profile's targeted-tests-by-surface list). If
  the subagent can't identify any, that's a signal the change is broader than
  expected — pause and reconsider scope, don't skip.
- **Red before green on defect beads.** A test written after its fix encodes the
  fix's own assumptions and can stay green while the bug is live. For a bead
  fixing a defect or regression, observe the reproducing test fail for the
  stated reason before fixing; the retro names that observed failure, not just
  "tests pass".
- **One commit per bead, always.** Every closed bead = exactly one commit on the
  branch. The handoff file is written BEFORE the close ceremony (crash recovery
  anchor); the close happens via a plain `bd update --status closed` BEFORE
  staging, so the close is folded into the implementation commit. Closing never
  rewrites the description — grooming owns it.
- **The dispatch plan has no lean path.** It is written every run, whatever
  its size, because it is what a recovering orchestrator reads instead of
  re-deriving the run.
- **Surface implicit decisions, don't bury them.** A bead with an open
  decision is bounced at Phase 1.5. A choice a subagent still makes silently
  — one the bead's Design didn't pin — belongs in its `retro:` comment, not
  the diff.
- **Visible progress beats opacity.** Subagents write JSONL progress events. For
  any subagent past 5 minutes, the orchestrator must arm `Monitor` on the
  progress file so the user sees per-phase events live. Long opaque runs erode
  trust.
- **Subagents discard context on return.** This is why the skill prefers fresh
  subagents over `/compact`. Batch subagents (fast lane, shared-surface)
  compromise — they hold context for N beads but free it on return.
- **Don't read source files in the orchestrator.** Code reads happen inside
  subagents — grooming (Phase 1.5) reads source to close decisions and wire
  dependencies before dispatch.
- **Capture decisions in the bead, not just in chat.** The bead description
  (selected options and alternatives, set by grooming) is the durable record;
  the dispatch plan is session-scratch.
- **Fail loudly, recover quietly.** If a subagent blocks, log it and move on. If
  three consecutive block, stop and surface — that's a systemic problem. If a
  subagent crashes mid-step, follow the recovery procedure (the handoff file is
  the anchor).
- **Route by tier, and report every escalation.** A bead's `tier:` label is
  grooming's *prediction* of the capability the work needs; resolve it to a
  model via the project's `## model lineup` and dispatch there (Phase 2.5 → Tier
  selection). Assign nothing below the tier; escalate one tier on repeated
  failure; never downgrade to save cost mid-run. Every escalation is a required
  row in the Phase 4 report — it is the only routing evidence a run produces.
  Note the asymmetry: escalations expose under-routing, while over-routing
  succeeds silently and is paid for on every run, so a zero-escalation run is a
  prompt to examine the routing, not a clean bill of health. Beads marked "do
  not delegate" run in the orchestrator's own loop.
- **Project rules win over skill defaults.** The gate profile is the source of
  truth for build/test/lint/format commands and any custom gates the project
  adds. If `CLAUDE.md`/`AGENTS.md` bans `TaskCreate`, skip it (use inline
  status). If the project has its own commit message conventions, follow them.
  The skill is opinionated but not authoritative.
- **Auto-chain waves under standing authorization.** When dependencies mean
  closing one wave unblocks the next, ask ONCE (at run start or first chain
  point) for standing approval: "re-resolve the queue after each push and
  dispatch the next wave until dry, implicit decisions defaulted, explicit
  `decision-required` beads still stop." With it granted, waves flow without
  per-wave prompts; the user gets a report per wave instead of a question per
  wave. Newly-filed beads that enter the queue mid-run (subagents filing
  adjacent issues) are admitted only by the mid-run scope test (Phase 1.5 step
  4): in-scope when a bead already in the run can't be finished without them,
  not merely when they match the filter. Admitted beads pass the grooming gate
  before joining a wave; out-of-scope finds are filed, linked, and surfaced for
  the user — never dispatched. The standing authorization finishes the launched
  scope; it never enlarges it.
- **Keep an environmental-failure ledger; never re-diagnose per wave.** Suites
  that fail — or silently PASS for environmental reasons — in the local sandbox
  get diagnosed ONCE, recorded in the gate baseline pre-flight captured,
  then the verdict is passed into every subsequent subagent prompt as fact. The
  silent-pass case is the dangerous one: a gate can go green without exercising
  your change (a harness running a stale prebuilt binary a plain test command
  didn't rebuild, a suite skipped by default, a cached result). When a gate
  *can* pass without touching your change, force a clean rebuild and confirm it
  actually ran before trusting it. Heavy suites (integration, codegen, anything
  needing sockets/toolchains the sandbox lacks) run locally only when a bead
  touches the surface they cover; when a PR is open, CI is the backstop.
- **Guard against foreign dirt and index races.** Multiple agents can share the
  tree (a diagnosis agent's uncommitted fixes, an in-flight batch's staged
  `.beads/issues.jsonl`). Batch prompts must name any known foreign files as
  untouchable; the orchestrator commits its own work by explicit pathspec (`git
  commit -- <files>`) so a concurrently-staged tracker ledger rides with the
  next bead commit instead of being swept up mid-ceremony. The same guard
  applies to tracker *claims*: a subagent may only reset a claim it made itself
  — a bead found `in_progress` that this agent didn't claim belongs to a live
  sibling — and the orchestrator (never a subagent) owns reconciling claims
  across waves.
- **Wire findings capture when the repo has it.** If `rfcs/findings-capture.md`
  exists (retro-spec protocol), every batch prompt includes: record
  `retro:`-prefixed `bd comment`s on the bead being worked, 1-3 sentences,
  friction only, zero analysis. The feature's final DoD bead runs the one
  aggregated retro-spec; closed-bead `## Retro` sections serve as the findings
  record for beads that predate the wiring.

## Reference files

Load the one that matches where you are in the run:

| Working on | Load |
|------------|------|
| Dispatching any subagent (Phase 1.5 grooming, Phase 3 execution) | `references/subagent-prompts.md` |
| Writing or amending the Phase 2.5 dispatch plan | `references/dispatch-plan.md` |
| A subagent that crashed, timed out, or is running past 5 minutes — or resuming a run whose state you do not hold | `references/recovery.md` |
| A project with no `## bead-bonanza gates` or `## model lineup` block | `references/project-setup.md` |
| Phase 0 pre-flight: probe the environment, resolve `RUN_DIR`/`SCRIPTS` | run `scripts/pre-flight-check.sh` |
| Emitting a progress event | run `scripts/bb-progress.sh` |
| Checking a `skipped` claim against the tree | run `scripts/bb-verify-skip.sh` |
| Closing a bead: close + export + stage + commit | run `scripts/bb-close-bead.sh` |
| The Definition of Ready the Phase 1.5 gate enforces | `references/shared/work-order-definition-of-ready.md` |

The reference files are one hop from here — read them in full rather
than previewing, and substitute every `{PLACEHOLDER}`, `{RUN_DIR}`, and
`{SCRIPTS}` before sending
a prompt.

The four `scripts/` entries are **executed, not read**. They exist because
each encodes an ordering, quoting or portability rule that is easy to state
and easy to get wrong under pressure — the close ceremony's four ordering
constraints, the progress event's quoting hazard, the two-sided skip check,
and pre-flight's ten probes across two platforms. Run them rather than
reproducing what they do.

Because they are not to be read, their interfaces are published here — an
agent that must open a script to learn how to call it has already been sent
to read it:

```
pre-flight-check.sh <run-id> [--gate-baseline] [--self-test]   # sandbox ON
bb-progress.sh   <run-dir> <bead-id> <phase> [key=value ...]
bb-verify-skip.sh [pathspec ...]        # no pathspec = whole working tree
bb-close-bead.sh <bead-id> <priority> <title> [path ...]
```

- **`scripts/pre-flight-check.sh`** — Phase 0, mechanized. Human report on
  stderr, an evalable `RUN_DIR=`/`SCRIPTS=`/`BB_*=` block on stdout. Exit 0
  proceed / 1 blocking / 2 bad invocation. **Run it with the sandbox ON** — the
  one write the sandbox refuses is the global Dolt telemetry config, which the
  script detects, skips, and reports as the single command to re-run with the
  sandbox off. `--gate-baseline` also runs the gate profile on the untouched
  tree, so a pre-existing failure is recorded as the baseline instead of being
  charged to the first bead. `--self-test` exercises the portability wrappers
  and takes no run id.
- **`scripts/bb-progress.sh`** — appends one event to
  `<run-dir>/<bead-id>.progress.jsonl`. `<phase>` is a closed vocabulary:
  `claimed`, `implementing`, `gates-running`, `gates-passed`,
  `retro-written`, `committed`, `pushed`, `blocked`. Extra fields are
  `key=value` pairs (`sha=a1b2c3d`, `reason="gates failed twice"`).
  Exit 2 on a bad invocation, and it writes nothing in that case.
- **`scripts/bb-verify-skip.sh`** — the two-sided skip check. With no argument it
  examines the whole working tree; pass pathspecs to scope it.
- **`scripts/bb-close-bead.sh`** — the close ceremony, in its required order.
  `<path ...>` are the code files this bead changed; WHEN omitted it stages
  everything currently modified, so pass explicit paths whenever another
  agent may be sharing the tree. It closes with a status-only `bd update` and
  never writes the description.

### Side effects — what each script leaves behind

These are the artifacts each script creates, writes, or repoints. They are
published for the same reason the usage lines are: the scripts are executed,
not read, so an agent that obeys that rule cannot discover a side effect by
opening the file.

- **`scripts/pre-flight-check.sh`** — creates `$RUN_DIR` if it does not
  exist, repoints `$RUN_ROOT/current` symlink at the start of every run,
  emits `BB_*` facts (`BB_LINT`, `BB_MTIME`, etc.) on stdout.
- **`scripts/bb-progress.sh`** — creates `<run-dir>` if it does not exist,
  appends one JSONL event to `<run-dir>/<bead-id>.progress.jsonl`.
- **`scripts/bb-verify-skip.sh`** — no side effects; exits 0 (skip
  verified), 1 (not skipped), or 2 (invocation error).
- **`scripts/bb-close-bead.sh`** — runs `bd update --status closed`,
  rewrites `.beads/issues.jsonl` by exporting the database, stages the
  specified paths plus `.beads/issues.jsonl` and
  `.beads/interactions.jsonl`, makes one commit.

Each script's own header carries the same usage line. WHEN a usage line
here disagrees with the script, the script is authoritative — and that
disagreement is a finding worth a retro.

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
the repo-level `scripts/sync-shared-references.sh` so this skill is
self-contained. Read them normally; never edit them — fix the source named
in each file's header and re-run that script.

## Retro this skill

WHEN a run of `bead-bonanza` struggled — you improvised around its
guidance, the gate profile it prescribes was wrong or missing, a command
it told a subagent to run misbehaved, or an instruction here pointed you
the wrong way — offer to run `retro-skill` on `bead-bonanza` while the
session is fresh, to capture the friction as an evidence-backed RFC. A
clean run needs nothing; never run it unprompted.
