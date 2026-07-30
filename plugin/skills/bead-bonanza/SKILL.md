---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use when the user asks to work on multiple beads at once — e.g. "work on all open beads with label X", "address all P1 beads", "clear the bead backlog". Triages the queue, gates every bead for readiness (grooming under-specified beads via a subagent, or bouncing what can't be made ready), surfaces both explicit and implicit decisions up front, dispatches via the cheapest viable execution mode (doc fast lane / shared-surface batch / worktree parallel / sequential), streams per-bead progress, and lands each bead in a single commit.
version: "1.2"
---

# bead-bonanza

Batch-process beads via subagent orchestration. The orchestrator (you) stays small — you only hold the queue, the user's decisions, the dispatch plan, and one-line outcomes per bead. Each bead runs in a fresh subagent so context doesn't compound.

Read `CLAUDE.md` and `AGENTS.md` once at the start so you have repo conventions in scope before delegating.

**This skill is language-agnostic.** It never hard-codes build, test, lint, or format commands. Everything it needs to validate the work done for a bead comes from the project's **gate profile** (Phase 0, step 1). Throughout this document, `<format>`, `<lint>`, `<build>`, `<test>`, and `<regen>` are placeholders that resolve to that project's actual commands. The profile can also define **custom gates** — any extra command or inline script the project wants run as part of the gate suite (e.g. a schema check, a license-header linter, a `bash` block). Substitute the profile's values wherever you see a placeholder, and run any custom gates alongside them.

## Phase 0 — Pre-flight

Before reading the queue, verify the environment won't interrupt the run:

1. **Resolve the gate profile.** bead-bonanza learns how to validate a bead from a **gate profile** the project keeps in `AGENTS.md` under a `## bead-bonanza gates` heading. Read it first — every `<format>`, `<lint>`, `<build>`, `<test>`, and `<regen>` reference below resolves to a command from that block.
   - **If the block exists**, load it. Fast path — no prompt.
   - **If it's missing**, discover candidate commands from the rest of `AGENTS.md`/`CLAUDE.md` (build/test/lint/format sections, CI config), propose a draft profile to the user, let them confirm or correct it, then **write it into `AGENTS.md`** under `## bead-bonanza gates` (create `AGENTS.md` if neither doc exists) and offer to commit it. Future runs then hit the fast path.
   - The block is human-readable markdown: the core gate commands, an optional pre-test `<regen>` step (codegen/snapshot regeneration — many projects have none), a targeted-tests-by-surface list, and the environmental-failure ledger. Template:
     ```markdown
     ## bead-bonanza gates

     - **Format:** `<cmd>`
     - **Lint:** `<cmd>`
     - **Build:** `<cmd>`
     - **Test (full suite):** `<cmd>`
     - **Regenerate before tests** (codegen/snapshots; omit if the project has none): `<cmd>`

     ### Custom gates (optional)
     Any extra checks to run as part of the gate suite — a plain command, or an inline script the user describes:
     ```bash
     # e.g. a project-specific check the standard gates don't cover
     ./scripts/check-migrations.sh
     ```

     ### Targeted tests by surface
     When only part of the tree changed, run the narrowest suite that covers it:
     - `<path or surface>` → `<targeted test cmd>`

     ### Environmental-failure ledger
     Suites that fail — or falsely PASS — only in the local sandbox, diagnosed once and passed into every subagent prompt as fact — never re-diagnosed per wave.
     - `<suite>`: fails in sandbox — `<symptom>` → `<workaround / covered-by-CI note>`
     - `<suite>`: falsely GREEN — e.g. runs a stale prebuilt binary a plain test command doesn't rebuild → force a clean `<build>` before trusting the result
     ```
   - If the project genuinely has no build/test step (e.g. a pure-docs repo), record that: `<test>` is a no-op and the doc/chore fast lane is the only dispatch mode. Don't invent gates.
1a. **Resolve the model lineup.** Beads carry a capability `tier`, not a
   model name. Read the project's `## model lineup` block in `AGENTS.md`
   to resolve tier → model at dispatch:

   ```markdown
   ## model lineup

   Least to most capable:

   - **light:** `<model>`
   - **standard:** `<model>`
   - **advanced:** `<model>`
   - **frontier:** `<model>`
   ```

   - **If the block exists**, load it. Fast path — no prompt.
   - **If it's missing**, propose one from the models available in this
     harness, let the user confirm, then write it into `AGENTS.md` (same
     mechanics as the gate profile above).
   - **If the user declines**, dispatch without the `model` param — every
     subagent inherits the session model. Say this once, at run start, so
     the cost profile isn't a surprise.

2. **Common command permissions.** Confirm these are in `.claude/settings.local.json` allow list. If any are missing, ask the user once whether to add them, then add (don't add silently):
   - `Bash(bd:*)`, `Bash(bd create:*)`, `Bash(bd update:*)`, `Bash(bd close:*)`
   - `Bash(wait)`, `Bash(wait %1)`, `Bash(jq:*)`
   - The gate-profile commands — whatever `<format>`/`<lint>`/`<build>`/`<test>`/`<regen>` resolve to for this project (e.g. `Bash(cargo:*)`, `Bash(npm run test:*)`, `Bash(go build:*)`, `Bash(pytest:*)`, or the regen script's path)
   - `Read(/tmp/claude/**)` — and `Read($TMPDIR/**)` too WHEN `TMPDIR` is set
     (macOS sets it to a per-user dir; most Linux hosts leave it unset)
3. **Beads telemetry.** `bd` stores its data in Dolt, which calls `eventsapi.dolthub.com`. Just kill it — don't ask, don't suggest. Run with sandbox disabled (writing `~/.dolt/config_global.json` is sandbox-blocked):
   ```
   dolt config --global --add metrics.disabled true   # dangerouslyDisableSandbox: true
   ```
   Then add a permanent deny so the domain can't sneak back in: edit `~/.claude/settings.json`, append `"WebFetch(domain:eventsapi.dolthub.com)"` to `permissions.deny` (creating the array if needed). Tell the user in one line what was done. Skip if the config already shows `metrics.disabled = true` — check first with `dolt config --global --list`.
4. **Run scratch directory.** Pick a run id (e.g. `$(date +%s)-<N>` where N is the queue size), then resolve the run root ONCE and confirm it is writable before Phase 1:
   ```bash
   RUN_DIR="${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>"
   mkdir -p "$RUN_DIR" && touch "$RUN_DIR/.ok" && echo "run dir: $RUN_DIR"
   ```
   **Always write `${TMPDIR:-/tmp}`, never a bare `$TMPDIR`.** `TMPDIR` is set on macOS but unset on most Linux hosts and in most containers, where a bare `$TMPDIR/claude/...` expands to the absolute path `/claude/...` — which is unwritable, so every later write fails. Substitute the resolved `$RUN_DIR` wherever this document shows the path.

   Do not proceed while that command fails. This directory holds the decisions doc (when one is needed), the dispatch plan (when one is needed), per-bead progress files (one `.progress.jsonl` per bead), and per-bead retro files (one `.retro.md` per bead, written by the subagent before close) — and crash recovery reads from it, so a run started without it cannot be recovered.
5. **TaskCreate fallback.** This skill defaults to `TaskCreate`/`TaskUpdate` for orchestrator-side progress visibility. If the project's `CLAUDE.md`/`AGENTS.md` explicitly bans `TaskCreate` (e.g. "use bd for ALL task tracking"), skip it — the project rule wins. Use inline status updates in your text output instead. (`bd` already tracks the work; `TaskCreate` is purely orchestrator scaffolding for in-conversation visibility.)
6. **Push mechanics.** The sandbox blocks `.git/config` writes, so `--set-upstream` never persists and a bare `git push` exits 128 with "no upstream" (and `| tail` masks the failure — always propagate the pipe's exit status). Push with an explicit refspec every time: `git push origin <branch>`. Verify landings with `git rev-parse origin/<branch>` after fetch, not with `git status -sb`.
7. **Open a draft PR early (ask once).** If the repo has PR-triggered CI, ask the user at run start whether to open a draft PR from the working branch. Every wave push then gets the full CI matrix in the cloud — including toolchain-heavy suites the local sandbox can't run — which is what makes the slim local gate policy (below) safe. Note: `gh` fails TLS verification inside the sandbox; run `gh` commands with sandbox disabled.

## Phase 1 — Resolve the queue

Parse the user's filter from their request (label, priority, ID list).

1. Run `bd ready` with the filter — this gives dependency-aware ordering. Examples:
   - `bd ready --label <label>`
   - `bd ready --priority 1`
   - `bd ready -n 50` (raise the limit if needed)
2. If `bd ready` doesn't accept the filter, fall back to `bd list --status open --json` and filter the JSON yourself.
3. Only include beads whose status is `open`. Skip `in_progress` (may be claimed elsewhere), `closed`, `blocked`.
4. Print the queue to the user: count + ordered list of `id | P# | title`.
5. If the user gave no filter, ask once: "Which beads? (label / priority / explicit ID list)" — this is the only Phase-1 question allowed.

Use `TaskCreate` to track one task per bead so progress is visible (unless the project bans it — see Phase 0 step 5).

## Phase 1.5 — Readiness gate (grooming)

bead-bonanza dispatches **only groomed beads.** Every later phase already assumes grooming's outputs: Phase 1's `bd ready` ordering comes from the dependency graph grooming wires; Phase 2 reads the `decision-required` labels grooming sets; Phase 2.5 reads the `tier:` assignment grooming makes. An ungroomed queue makes all three unreliable — and handing a fresh, zero-context subagent an under-specified bead is the single most expensive failure mode this skill has: it detonates mid-wave, after later beads have built on it.

So before any decisions or dispatch, gate the queue against the **Definition of Ready** — the checklist owned by the `spec` skill (`../spec/references/work-order-definition-of-ready.md`) and enforced by `groom-beads`, which materializes work orders as beads. Those skills are the authority; do not restate the checklist here.

### 1. Cheap check (orchestrator, read-only)

Most runs pass in seconds because the user groomed first. Per bead, confirm the readiness signals (parallelize — this is read-only):

- `bd lint` is clean for it (no missing acceptance / repro / success criteria), if the installed `bd` has it.
- a `feature`/`bug` carries a `design` and `acceptance` criteria; a `bug` carries a repro.
- a `tier` is assigned (grooming never leaves it null).
- dependencies are wired (ordering must live in the graph, not in priority numbers or prose).
- no session-dependent language — "as discussed", "the above", "finish the earlier refactor" — the fresh-subagent test.

A queue that is entirely ready makes this phase a one-line no-op: say so and proceed to Phase 2.

### 2. Groom what fails (delegate — don't groom in the orchestrator)

If any bead fails the check, do NOT dispatch it and do NOT fix it inline — grooming reads source and wires graphs, which belongs in a subagent (see "don't read source files in the orchestrator"). Dispatch a **grooming subagent** over the ungroomed subset (template below); it runs `groom-beads` to the Definition of Ready and returns a per-bead readiness report.

When it returns, **re-run `bd ready`** with the run's filter: grooming may have wired new dependencies or split a boulder into an epic + children, changing both the ordering and the membership of the queue.

### 3. Bounce what grooming can't clear

Grooming, like this skill, bounces gaps rather than guessing. A bead it cannot bring to Ready — needs a human decision it can't make, can't be sized without a split, rests on an unexamined premise — does **not** enter the run. Exclude these from the queue and surface them to the user at the end of this phase (one line each + the gap). The run proceeds with the Ready beads; ungroomable beads are the user's to resolve, not a subagent's.

### 4. Admitting emergent work mid-run

During execution (Phase 3) subagents discover work. Some of it belongs in this run; some is scope creep wearing a matching label. Telling them apart is the point — pulling in the wrong thing is as real a failure as leaving a blocker unaddressed. This is where the autonomy the run was launched for lives, and where it stops.

**The test is necessity, not resemblance:** emergent work is in-scope only when a bead *already in the run* cannot be correctly and completely finished without it. Three shapes qualify:

- **an unblocking prerequisite** — the assigned bead genuinely cannot be completed until this is done;
- **an acceptance-criteria decomposition** — a piece the assigned bead's own AC cannot be satisfied without (a sign the bead was under-groomed, larger than it looked);
- **an expected test-coverage gap** — coverage the assigned scope already implied should exist and doesn't.

Everything else — adjacent cleanups, refactors "while I'm here", a new capability no assigned bead's AC demands — is **out of scope even if it matches the run's filter.** Matching the label is necessary but not sufficient; "required to finish a bead already in the run" is the bar.

**Mechanics.** The discovering subagent only *files* the bead (`bd create`) with a `discovered-from` edge to its origin and a one-line note of which shape it is (or "out-of-scope"); it never implements it. If its own bead is truly blocked by it, it takes the Blocked path naming the new id. The **orchestrator** then decides:

- **In-scope** → run it through this gate (groom → re-`bd ready`), wire the `blocks` edge so it orders ahead of the bead it unblocks, and dispatch it under the standing auto-chain authorization. No new prompt — this is the intended autonomy. (A groomed prerequisite that lands a `decision-required` label still stops for the user, per the standing-authorization terms.)
- **Out-of-scope** → leave it filed and linked, do NOT dispatch it, and surface it in the final report for the user to triage. The standing authorization finishes the launched scope; it never enlarges it.

**Ballooning is a stop signal.** If "prerequisites" multiply — a discovered blocker that is really a new epic, or more than a handful of new in-scope beads off a single bead — stop and surface it. That means the original scope was mis-groomed, which is a human's call, not an autonomous expansion.

### Grooming subagent prompt template

```
You are grooming {N} beads to the Definition of Ready so an orchestrator can dispatch each to an isolated, zero-context sub-agent. Use the `groom-beads` skill; it enforces the spec-owned Definition of Ready (`../spec/references/work-order-definition-of-ready.md`). Follow its full workflow (Definition of Ready, dependency wiring, model assignment, C4 blast radius, validate).

Beads to groom:
{ID_1} — {TITLE_1}
{ID_2} — {TITLE_2}
...

For each bead: bring it to Ready by editing it in place with `bd update` — enrich thin descriptions rather than discarding author intent; wire only dependencies you can justify; assign a `tier` with a `tier_reason`.

Do NOT implement any bead. Do NOT invent IDs, component names, or dependencies. Bounce, don't guess: any bead you cannot bring to Ready, leave as-is and report NOT-READY with the specific gap a human must close.

Return a report, one line per bead:
- `{ID}: ready — <notable change: description made context-independent / split into epic <id> + children <ids> / model assigned>`
- `{ID}: NOT-READY - <the gap that needs a human>`
```

## Phase 2 — Decisions

Run `bd show <id>` for every bead (parallelize — read-only). Classify each bead's decision state:

- **`explicit`** — bead has the `decision-required` label, a "DECISION REQUIRED" / "TBD" / "open question" marker, or a bulleted Options list with no chosen path.
- **`implicit`** — the description leaves an implementation choice open without flagging it as a decision. Look for: multiple acceptable acceptance-criteria phrasings ("either … or …"), under-specified output formats, fallback string formatting, default values, naming choices. The subagent will pick silently if you don't surface this.
- **`none`** — fully specified.

### Lean path: skip the decisions doc when nothing is pending

If **every** bead in the queue is `none` OR `implicit-with-trustworthy-default` AND there are no `explicit` decisions remaining (all previously-flagged decisions are already baked into bead descriptions with a SELECTED option), skip Step 2a entirely. State the implicit defaults inline in chat (one line each), issue a single consolidated implicit-decisions question (Step 2b's implicit flavor), and proceed. No `decisions.md` file. The doc is only worth writing when there are ≥3 implicit decisions worth tabulating, or any explicit decisions left to make.

For everything else (mixed bag, large queue, or any explicit decisions pending), use the full path below.

### Step 2a — Write the decisions doc

Write `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/decisions.md`. Two sections: **Explicit Decisions** and **Implicit Decisions**.

For each **explicit** decision:
- **Bead ID + title**
- **Problem in plain English** with file:line refs
- **Options table** — each option with: one-line summary, concrete code/output example, blast radius, trade-off
- **Recommendation** + one-sentence rationale
- **Status** — `[ ] pending`

For each **implicit** decision:
- **Bead ID + title**
- **The choice** — what the subagent will need to decide (e.g. "unknown-enum fallback: numeric code vs a `\"unknown\"` sentinel string vs raise an error?")
- **Subagent default** — what it will pick if you don't override (with rationale)
- **Status** — `[ ] subagent-default` (user can change to `[x] override: <value>`)

Read source files when needed for concrete examples. The user references this doc while answering.

Tell the user the doc's path explicitly when issuing AskUserQuestion.

### Step 2b — Batch ask

Issue `AskUserQuestion`. **Two flavors:**

- **Explicit decisions** — one multi-choice question per decision (max 4 per AskUserQuestion call; split into parallel calls if more). Recommended option first with " (Recommended)" suffix. Short labels — detail lives in the doc.
- **Implicit decisions** — ONE consolidated question: "Want to override any of the {N} implicit decisions? (defaults shown in the decisions doc)" with options: "Use all defaults (Recommended)" | "Override specific ones" (user replies with which + value via Other). Don't blast N implicit questions; most users will trust the defaults.

If the user replies with "clarify these questions" or pushes back, ask what they'd like clarified rather than guessing. After clarification, reissue with refined options.

### Step 2c — Bake decisions into beads

For each **explicit** decision answered:

1. Update the bead description: append (or modify) a `## DECISION (<YYYY-MM-DD>)` section. **Preserve all options** — mark only the selected one with " — SELECTED". Don't delete unselected options; future readers want to know what was considered.
2. Update acceptance criteria to refer to the selected option.
3. Remove the label: `bd update <id> --remove-label decision-required`.

For each **implicit** decision the user overrode: append the override + rationale to the bead description as a `## IMPLICIT DECISION OVERRIDE (<YYYY-MM-DD>)` section. Subagent-default decisions don't need a description update; the subagent will record its choice in the retro.

Do these `bd update` calls in parallel. Use `--body-file -` with HEREDOC for description updates.

Update `decisions.md` to mark each row `[x] <selected option>` or `[x] subagent-default`.

**After Phase 2c, do not ask the user any more questions.**

## Phase 2.5 — Triage & dispatch plan

**Bias toward batching.** Each "plain sequential" bead burns a full `<test>` run (often 10+ min). Batching N beads in one subagent saves N−1 of those. The skill's job is to find the largest safe batch, not to default to one-bead-per-subagent for caution. Sequential is the *exception*, not the rule.

### Lean path: skip the formal dispatch plan

Most runs don't need a markdown table. Skip the formal `dispatch.md` and just state the dispatch decision in one line in chat if **any** of:

- Queue size <5 beads.
- All beads land in one bucket regardless of dispatch mode (single-surface clusters, all-trivial batches).
- The user gave an explicit ID list rather than a label/priority filter — they've done the bucketing implicitly.

Examples of acceptable one-line dispatches:
- 1 bead: "Single bead, sequential, push at end."
- 3 beads same surface: "Single shared-batch subagent, defer-push, full gates at end."
- 4 doc-only beads: "Fast-lane subagent, defer-push."

Write the formal plan only if **both**: queue size ≥5 AND the queue spans multiple dispatch modes (e.g. "3 fast-lane + 4 shared-batch + 2 worktree-parallel"). Save to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md` in that case.

### Classify each bead

- **`trivial`** — pure documentation (md/mdx edits), single-line code swap, comment/regex-pin additions. No fixture regeneration needed; targeted unit tests suffice. Examples: "document a trade-off in docs/", "add an explanatory comment", "swap one equivalent stdlib call for another".
- **`isolated`** — touches files no other queued bead touches. No `BLOCKS` / `RELATED` link to another queued bead.
- **`shared-surface`** — touches files (especially snapshot/golden fixtures) that another queued bead also touches, or has explicit `RELATED` links to another queued bead.

Find file overlap by scanning each bead's description for fixture names + source paths, then bucketing by overlap.

### Choose a dispatch mode per bucket

Four modes. Pick the broadest one that's safe — you're trying to minimize total full-test-suite runs across the queue.

1. **Doc/chore fast lane** — `trivial` beads. ONE subagent processes the cluster sequentially. Targeted tests between beads, ONE full `<test>` run at the end. One commit per bead. Cap ~6 beads per fast-lane subagent.

2. **Shared-surface batch** — `shared-surface` beads that cluster on the same fixture/code area. ONE subagent processes the cluster sequentially. Between beads: `<format>`, `<lint>`, `<build>`, **plus surface-relevant targeted tests (non-negotiable, see Phase 3)**. ONE full `<test>` run at the END of the batch. One commit per bead. Saves N–1 full test-suite runs.

   **Lower the threshold:** use shared-surface batch when **≥2** beads share a surface and decisions are independent (bead A's choice doesn't change bead B's implementation). The previous "≥3" was too conservative. Even at 2 beads the savings are real and the bisect cost (one suspect commit) is low.

   **Don't over-split:** if 7 beads cluster on one surface and only 1–2 of them are "larger" (substantial new features), they can still go in one shared batch. Splitting buys you nothing if decisions don't conflict. Only split when bead A's implementation would conflict with bead B's, or when decisions overlap such that one bead's choice changes another bead's plan.

3. **Worktree parallel** — `isolated` beads with no surface overlap, where parallel wall-clock savings beat merge cost. Spawn each via `Agent` with `isolation: "worktree"`. Orchestrator merges results back to the parent branch one at a time. If a merge produces fixture/snapshot conflicts, run `<regen>` on the merged tree to settle (skip if the project defines no regen step — resolve the conflict manually). Skip this mode if there are <3 isolated beads (parallelism savings don't pay) or if the beads aren't actually isolated (you'll just hit merge conflicts).

4. **Plain sequential** — reserved for: a single bead with no batch peers, OR a bead whose surface diverges from every other queued bead, OR explicit user override. Not the default. If you find yourself defaulting here for >1 bead in a row, reconsider — they probably belong in a shared batch.

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
lineup` block in `AGENTS.md` (Phase 0 step 1a). Honor these:

- **Collect during Phase 2** — you're already running `bd show` per
  bead; record each bead's tier alongside its decision state.
  `bd list --label tier:standard --json` also works for bulk queries.
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
- **Record the routing** — when you write the formal dispatch plan, add
  a `Tier` column; on the lean path, name the tier in the one-line
  dispatch statement.

### Push policy

- **`push-per-bead`** — subagent pushes after each commit. Best for sequential mode (recovery is easier on a per-bead failure).
- **`defer-push`** — subagent commits per bead, orchestrator pushes once at bucket end. Default for fast-lane and shared-surface batch (one process; failures are observable).

For a single-bead "wave," `push-per-bead` and `defer-push` are equivalent — don't bother labeling. Just push when the subagent returns.

### When you write the formal plan

Print a table to chat AND save to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md`:

```
| Bead | Class | Bucket | Tier | Push | Reason |
|------|-------|--------|------|------|--------|
| demo-89s | trivial | fast-lane #1 | light | defer | docs only |
| demo-aqw | trivial | fast-lane #1 | light | defer | comment only |
| demo-9h0 | shared | shared-batch #1 | standard | defer | same serializer + fixture cluster |
| demo-04y | shared | shared-batch #1 | standard | defer | same cluster |
| demo-ojt | shared | shared-batch #1 | standard | defer | same cluster |
| demo-76y | shared | shared-batch #1 | standard | defer | same cluster |
| demo-2i5 | shared | shared-batch #1 | standard | defer | same cluster |
| demo-ckh | shared | shared-batch #1 | standard | defer | same cluster |
| demo-d3j | isolated | sequential | advanced | — | validator change; different surface |

(`Tier` = the bead's `tier:` label, resolved to a model at dispatch via the
project's `## model lineup`. Blank only when the project declares no lineup —
the subagent then inherits the session model.)
```

Don't ask for approval — just print and proceed. The user can interrupt if they disagree.

## Phase 3 — Execute

Execute buckets in this order (lowest blast radius first):

1. **Doc/chore fast lane** — single subagent.
2. **Shared-surface batches** — one subagent per batch, sequentially across batches.
3. **Worktree parallel waves** — spawn isolated beads concurrently in one message; merge each as it returns.
4. **Plain sequential** — one subagent per bead, in `bd ready` order.

After each subagent returns:

1. Append its summary lines to your log (one per bead processed).
2. Mark corresponding tasks complete via `TaskUpdate` (or skip if the project bans TaskCreate per Phase 0 step 5).
3. **Verify skip claims with a diff-stat.** For every bead the subagent returned as `skipped`:
   - Run `git diff --stat` and `git diff --staged --stat` (scoped to that bead's worktree if applicable).
   - **If both diffs are empty:** the skip is consistent. Confirm the subagent's summary line includes a one-line reason citing the bead/commit that already addressed the issue (e.g. `skipped - already addressed by demo-wxi @ a1b2c3d`). If the reason is missing, push back: re-prompt the subagent (or fill it in from `bd show <id>` history) before accepting the skip.
   - **If either diff is non-empty AND status=skipped:** this is an inconsistency. **Block the close.** The subagent either did work it didn't acknowledge, or left the tree dirty. Re-dispatch with explicit instructions to choose one of:
     - Commit the diff and change status to `completed` (run the full per-bead close ceremony), OR
     - Revert the diff (`git checkout -- <paths>` / `git restore --staged <paths>`) and re-confirm `skipped` with a fresh diff-stat showing clean.
   - Do NOT proceed to push or to the next bead until the skip is reconciled. This catches the case where the subagent's first read mistakenly concluded "already addressed" but it still produced edits — the orchestrator must not silently accept a contradictory state. (Origin: a near-miss where a subagent claimed a skip while leaving a non-empty diff.)
4. **Verify the working tree is clean** with `git status --porcelain`. If dirty, fall through to the recovery procedure below.
5. If the bucket used `defer-push`, run `git push` for the bucket now.
6. Move to the next bucket/bead.

**Circuit breaker:** if 3 consecutive beads return `blocked` or fail, stop the run and surface to the user. Something is systemically wrong (broken main, environment issue, bad decisions doc).

### Subagent crash / mid-step timeout recovery

If a subagent times out, returns an "API Error" instead of a summary, or otherwise dies before its summary line — check its progress JSONL to determine what state it left things in. Then:

1. **Read the progress file** (`<run-id>/<bead-id>.progress.jsonl`). The last phase tells you how far it got.
2. **If `gates-passed` was emitted but `committed` was not** — the work is done in the working tree, only the close+commit ceremony is incomplete. This is the **happy crash path**:
   - Read the retro file at `<run-id>/<bead-id>.retro.md`. If present, use it as the close body.
   - If the retro file is missing (subagent crashed before writing it), reconstruct minimally: the original bead description plus a short Retro noting "Subagent timed out before retro write; orchestrator finished close from working-tree state."
   - Run `bd update <id> --status closed --body-file <retro-file-or-tmp>`.
   - Run `bd export` into `.beads/issues.jsonl` first (the subagent may have crashed before exporting). Then stage source paths, any regenerated fixtures, `.beads/issues.jsonl`, `.beads/interactions.jsonl`. Inspect `git status --porcelain` carefully — only stage what belongs to this bead.
   - Commit with the canonical message format (`<id> P<priority> <title>`).
   - Push (if appropriate for the bucket).
   - Verify the bead is now `closed` via `bd show <id>`.
3. **If the last phase was `implementing` or earlier** — the work is partial. Stash with `git stash push -m "bead-bonanza recovery: <id>"`, mark the bead `blocked` with a recovery note, and surface to the user. Don't try to finish someone else's half-done work.
4. **If the progress file is empty or missing** — the subagent never started. Check `bd show <id>` to confirm status; if it's still `open`, retry the subagent. If it's `in_progress`, claim it back with `bd update <id> --status open` first.

The retro file is what makes the happy crash path mechanical instead of inferential — that's why it's non-negotiable.

### Mid-stream visibility (progress streaming)

Each subagent appends progress events to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.progress.jsonl`. Required event phases (one line each):

```jsonl
{"bead":"<id>","phase":"claimed"}
{"bead":"<id>","phase":"implementing"}
{"bead":"<id>","phase":"gates-running"}
{"bead":"<id>","phase":"gates-passed"}
{"bead":"<id>","phase":"retro-written"}
{"bead":"<id>","phase":"committed","sha":"<short>"}
{"bead":"<id>","phase":"pushed"}             # omitted in defer-push mode
{"bead":"<id>","phase":"blocked","reason":"..."}  # only on failure
```

**Events are FULLY LITERAL single-quoted strings** — `printf '%s\n' '{"bead":"demo-x","phase":"committed","sha":"abc1234"}' >> <file>` with the sha typed literally from the git output. Never embed `$(date ...)`, `"$var"`, or any runtime substitution: the permission system flags runtime-determined printf operands (zsh arith-eval injection heuristic) and prompts the user on EVERY event write. No timestamps — the orchestrator's monitor notification is timestamped on arrival, so `ts` never earned its keep.

Batch subagents (fast lane, shared-surface batch) write events for every bead they process to the same per-bead progress files (one file per bead-id, even from the same batch subagent).

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or actually does) run >5 min — true for any non-trivial bead given full test runs — the orchestrator must arm a `Monitor` on its progress JSONL so the user sees per-phase events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/<bead-id>.progress.jsonl | grep -E --line-buffered '"phase":"(committed|blocked)"'
```

For batch subagents, tail multiple files in one Monitor. **Filter to `committed|blocked` only** — those are the events the orchestrator acts on. Relaying every phase (implementing, gates-running, …) wakes the orchestrator per event and adds a conversation turn each time for narration nobody needs; the per-bead progress files still hold the full phase history for crash recovery.

### Crash-resilient close: write the retro to a file before bd update

After `gates-passed` and BEFORE the `bd update --status closed` call, the subagent **must** write its proposed Retro section to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.retro.md`. Then emit `{"phase":"retro-written"}`. Only then proceed to the close ceremony.

This is the recovery anchor. If the subagent times out, crashes, or otherwise dies between `gates-passed` and the commit, the orchestrator can:
1. Read the retro file.
2. Run `bd update <id> --status closed --body-file -` itself (concatenating original description + retro).
3. Stage, commit, push.

No information is lost; recovery is mechanical instead of inferential. **This is non-negotiable** — the cost is one extra file write per bead; the savings on a single recovery pay for years of runs.

### Per-bead subagent prompt template (sequential and worktree modes)

```
You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. The gate commands are in AGENTS.md under "## bead-bonanza gates".

Bead: {ID} (P{PRIORITY}) — {TITLE}
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` regen=`{REGEN or "none"}` custom=`{CUSTOM_GATES or "none"}`
Push mode: {push_per_bead | defer_push}
Progress file: ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.progress.jsonl
Retro file: ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md

## Workflow — follow in order, do not deviate

Append a JSONL event to the progress file at each labeled step (use `printf '%s\n' '<json>' >> <file>`).

1. Emit `{"phase":"claimed"}`. Claim: `bd update {ID} --status in_progress`
2. Read full details: `bd show {ID}`
3. Emit `{"phase":"implementing"}`. Implement the change. **If this bead fixes a defect or regression, first write (or locate) a test that reproduces it and observe it FAIL for the stated reason BEFORE you fix it** — a test written after the fix encodes the fix's own assumptions and can stay green while the defect is live. Record the observed failure; it goes in the retro. Stay strictly within this bead's scope. If you find work beyond it, file a new bead via `bd create` with a `discovered-from` link to {ID}, noting in one line whether it *blocks* this bead (a prerequisite, an acceptance-criteria decomposition, or an expected test-coverage gap) or is merely adjacent — the orchestrator decides whether it enters the run. Do NOT implement it or pull it into your own scope. If this bead genuinely cannot be finished without it, take the Blocked path naming the new bead as the blocker; otherwise continue with the original — do NOT expand scope inline. For any implicit decision you must make, record what you chose (and why) — you'll put it in the retro.
4. Emit `{"phase":"gates-running"}`. **The push is the gate** — this is a single-bead push, so full gates run here before pushing. If the change affects generated output the tests compare against (codegen, snapshots, golden files, schemas) AND the gate profile defines a regen step, run `{REGEN}` BEFORE the test step.
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile: `{CUSTOM_GATES}` (skip if none)
   On gate failure, fix the root cause and rerun. After two full failed cycles, follow the Blocked path — do not keep retrying. Emit `{"phase":"gates-passed"}` after success.
5. **Write the Retro to a file BEFORE closing the bead** (crash recovery anchor — non-negotiable).
   - Write `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md` containing the FULL bead description (original + a `## Retro` section appended). The Retro section covers:
     - Approach taken (which option, if a decision was made)
     - **For a defect/bug bead: the observed pre-fix failure** — the test and how it failed *before* the change (e.g. "test X failed with Y"), not merely "tests pass now"
     - **Implicit decisions you made** (one bullet each: what you chose + why)
     - Anything that differed from the original instructions
     - Anything surprising; ideas for improving prompts/skills/patterns
   - Emit `{"phase":"retro-written"}` after the file is on disk.
6. Atomic close — write description + status in ONE call, sourcing the retro file:
   ```
   bd update {ID} --status closed --body-file ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md
   ```
   Do NOT run `bd close` — `bd update --status closed` does both. This avoids a second commit.
7. **Refresh the durable tracker export, then stage.** `.beads/issues.jsonl` is a passive export — `bd` does not rewrite it on every write (Dolt auto-commit defaults off), so a closure that is not exported is lost if the container is reclaimed. Run `bd export` and make sure the result lands in `.beads/issues.jsonl` (redirect if your build of `bd` writes to stdout: `bd export > .beads/issues.jsonl`), and confirm the file actually changed. Then stage everything together: `git add <code paths>` AND `git add .beads/issues.jsonl` (also `.beads/interactions.jsonl` if modified).
8. Single commit (HEREDOC, no `--no-verify`, no `--amend`) with message exactly: `{ID} P{PRIORITY} {TITLE}`. Emit `{"phase":"committed","sha":"<short>"}`.
9. If push mode is `push_per_bead`: `git push` on the current branch (never main). Emit `{"phase":"pushed"}`. If `defer_push`: skip the push; orchestrator handles it.
10. Return one line as your final message:
    - Success → `{ID}: closed <short-sha>`
    - Blocked → `{ID}: blocked - <short reason>`
    - Skipped → `{ID}: skipped - <short reason>`

## Blocked path

If gates fail twice, the change is intractable, or the bead cannot be finished without a prerequisite you filed:
- Emit `{"phase":"blocked","reason":"<short>"}`.
- `bd update {ID} --status blocked --append-notes "<short reason>"` — if a filed prerequisite is the blocker, name its id in the reason.
- Stash any uncommitted changes: `git stash push -m "bead-bonanza blocked: {ID}"`
- Return the blocked summary line. Do not push.

## Skipped path

If the bead is already addressed, malformed, or duplicates a closed bead:

1. **Confirm your working tree is clean — this is a hard precondition, not a courtesy.** Before doing anything else, run the diff-stat self-check:

   ```bash
   git diff --stat            # unstaged changes
   git diff --staged --stat   # staged changes
   ```

   Both must produce empty output. If either shows changes:
   - You did work the bead should record. **This is not a skip.** Go back to step 4 (`gates-running`) of the per-bead workflow and finish the close ceremony as `completed`.
   - OR the changes are accidental / pre-existing / out-of-scope. Revert them: `git restore --staged <files>` then `git checkout -- <files>` (or `git stash push -m "out-of-scope before skip: {ID}"` if you want to preserve them for later inspection). Re-run the diff-stat check; it must be empty before you proceed.

   Do not skip this check. The orchestrator runs the same diff-stat verification on return (Phase 3 post-return step 3 above); a non-empty diff with a `skipped` claim will block the close and force a re-dispatch. The Phase 3 check is the safety net — your self-check here is the primary line of defense.

2. `bd update {ID} --status closed --append-notes "duplicate of <id>"` (or similar one-line rationale).

3. Return the skipped summary line **with a citation**: `{ID}: skipped - already addressed by <bead-id> @ <short-sha>` or `{ID}: skipped - duplicates <bead-id> (closed YYYY-MM-DD)`. "Already done" without a citation is not acceptable; the orchestrator can't verify it. To find the citation, use `git log --oneline --all | grep <keyword>` or `bd show <suspected-duplicate-id>`.

## Constraints

- Do not ask the user questions. Decisions are baked into this prompt.
- Do not modify code outside this bead's scope.
- Never reset or reclaim a bead's status that you did not set yourself. A bead already `in_progress` when you arrive belongs to another agent — treat it as untouchable, exactly like foreign working-tree changes; leave claim reconciliation to the orchestrator.
- Do not skip git hooks (`--no-verify`). If you realize a required step was missed (e.g. the tracker export) after committing, add a NEW follow-up commit — never amend a commit that may already be pushed, and never force-push a published branch. (Amending a local commit you are certain is unpushed, to fold in a missed step, is fine.)
- Only push the current branch — never push to main.
- Do not use `git reset --hard`, `git checkout .`, or other destructive recovery commands without first stashing.
```

### Doc/chore fast-lane subagent prompt template

```
You are processing {N} trivial beads in this repo. Read CLAUDE.md and AGENTS.md if you need conventions; gate commands are in AGENTS.md under "## bead-bonanza gates".

Beads (process in order; one commit per bead):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2}
...

Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` custom=`{CUSTOM_GATES or "none"}`
Push mode: {push_per_bead | defer_push}
Run id: {RUN_ID}

These are documentation, comments, single-line swaps, or other low-risk changes. For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push) exactly as in the sequential template, INCLUDING per-bead progress JSONL writes AND the per-bead retro file write to `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/<bead-id>.retro.md` BEFORE close.
2. Between beads: skip the full `{TEST}` run. Run `{FORMAT}` + `{BUILD}` per bead. **Targeted tests are non-negotiable** — identify which test file covers the touched surface (see the profile's targeted-tests-by-surface list) and run it. If the change is so trivial there's no plausible test (e.g. pure docs edit, a comment with zero output diff), say so explicitly in the bead's retro and continue. "I couldn't find a targeted test" without explanation is not acceptable.
3. After the LAST bead in this batch, run the full gate suite once (`{FORMAT}`, `{LINT}`, `{BUILD}`, `{TEST}`, plus any custom gates). **The push is the gate** — full gates must pass before the orchestrator pushes. If full gates fail, the most-recent bead is the suspect — investigate and fix it (a new commit), then continue.

Return ONE message containing N+1 lines: one per-bead summary, then a final `batch-end: full gates {passed|failed}`.

## Constraints

Same as the per-bead constraints. Each bead is its own commit. In `defer_push` mode, do not push at all — the orchestrator pushes after the batch.
```

### Shared-surface batch subagent prompt template

```
You are processing {N} related beads that share a code/fixture surface ({SURFACE_DESCRIPTION}, e.g. "the JSON serializer in src/serialize.* and the serializer golden fixtures").

Read CLAUDE.md and AGENTS.md if you need conventions; gate commands are in AGENTS.md under "## bead-bonanza gates".

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` regen=`{REGEN or "none"}` custom=`{CUSTOM_GATES or "none"}`
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}
Push mode: {push_per_bead | defer_push}
Run id: {RUN_ID}

These beads are NOT trivial — they touch real code. The reason they're batched is to avoid running the full `{TEST}` suite between them when later beads will rebuild the same surface. **The push at the end is the gate; per-bead validation in this batch is a regression-detector to prevent rework, not the final gate.**

For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push), INCLUDING progress JSONL writes per bead AND the per-bead retro file at `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/<bead-id>.retro.md` BEFORE close.
2. Between beads, run gates AT THIS REDUCED LEVEL — but **targeted tests are non-negotiable**:
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - If generated output was touched AND a regen step exists: run `{REGEN}` then the relevant fixture/snapshot test.
   - **Run the surface-relevant targeted tests** named above for the file(s) you touched. If none is listed for a surface you touched, consult the profile's targeted-tests-by-surface list; if there's still nothing, that's a signal the change is broader than expected — pause and reconsider scope, don't take it as permission to skip.

   **Why this matters:** without per-bead targeted tests, a regression introduced in bead #2 only surfaces at batch end, after beads #3 and #4 may have built on the broken assumption. You then bisect, rework #2, and possibly redo #3 and #4. Targeted tests prevent that cascade.
3. **After the LAST bead in this batch** (push gate). If the repo has a pre-push hook running a CI-mirror gauntlet AND a PR is open so pushes trigger full CI (ask the user once at run start — "slim gates"), SKIP this step entirely: the hook and CI *are* the full gate, and running the suite in-batch just duplicates them minutes apart. Otherwise run the full gate suite — must pass before orchestrator pushes:
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile
   If it fails, identify which bead introduced the regression (bisect via the per-bead commits — `git log --oneline | head` then `git checkout <sha>` and rerun the failing test) and follow the Blocked path for THAT bead — do not silently fix it across multiple commits.

Return ONE message containing N+1 lines: one per-bead summary, then a final `batch-end: full gates {passed|failed}`.

## Constraints

- Each bead is its own commit. Never combine bead changes into one commit.
- If a bead's implementation conflicts with an earlier bead in this batch (e.g. one removed a function the next was about to modify), follow the Blocked path for the conflicting bead and continue with the rest.
- Same per-bead constraints as the sequential template.
```

### Worktree parallel notes (orchestrator)

When dispatching a wave of `isolated` beads in parallel:

- Spawn all subagents in a single message with multiple `Agent` tool calls, each with `isolation: "worktree"`.
- Each subagent uses the per-bead template above, runs in its own worktree on its own branch.
- After all subagents return, merge their branches into the parent branch one at a time:
  - `git merge <branch>` — if clean, continue.
  - If a fixture/snapshot conflict and the profile defines a regen step: take both sides, run `<regen>` on the merged tree, stage and amend the merge commit. Without a regen step, resolve the conflict manually.
  - If non-fixture conflict on shared code: triage failure — the beads weren't actually isolated. Mark the unmerged bead as `blocked` with a note about the conflict; let the user resolve.
- If the bucket used `defer-push`, push once after all merges complete. Otherwise individual subagents pushed their branches; the orchestrator's merge commits still need a `git push` per merge or one `git push` at the end (your call — defer is fine here).

## Phase 4 — Final report

After the queue is drained (or the circuit breaker tripped), print:

1. A markdown table: `| ID | Status | Bucket | Commit | Notes |`
2. Totals: `N closed, N blocked, N skipped`.
3. **Implicit decisions made** — extract from the retros: which beads had implicit decisions, what the subagent chose. Surface to the user so they can audit non-obvious choices.
3a. **Tier escalations** — every bead whose tier you had to raise to get it done, one row each:

   ```
   | Bead | Assigned | Completed at | What failed at the assigned tier |
   |------|----------|--------------|----------------------------------|
   | demo-9h0 | light | standard | fixture regen wrong twice; gates failed both cycles |
   ```

   State `none` explicitly when there were no escalations — the null is
   the result, not the absence of one.

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
4. **Follow-up beads** — beads created mid-run via `bd create`, split into those admitted into the run (in-scope prerequisites / AC-decompositions / coverage gaps, with their outcome) and those left for the user (out-of-scope finds, filed with a `discovered-from` link, awaiting triage).
5. **Stashes** — if any were created during recovery, list with `git stash list`.
6. **Pushes** — confirm all closed beads' commits are on the remote.
7. Path to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/` so the user can review the decisions doc, dispatch plan, and progress files.

Do NOT create a PR or merge to main — that is a separate, explicit user action.

## Operating principles

- **Groom before you dispatch.** bead-bonanza executes *groomed* beads; it never fixes an under-specified bead inline. Phase 1.5 gates the queue against `groom-beads`'s Definition of Ready, delegates grooming of the rest to a subagent, and bounces what can't be made ready. This holds for beads pulled into scope mid-run too — a fresh subagent handed an ungroomed bead is this skill's most expensive failure mode.
- **Bias toward batching, not sequential.** Each plain-sequential bead burns a full `<test>` run. Find the largest safe batch — surface-overlap is the dominant signal, not bead complexity. Sequential is reserved for genuine surface divergence or single-bead queues, not "I'm being cautious."
- **The push is the gate.** Commits inside a branch are recoverable; only the push to origin is committed-to-the-world. Per-bead validation is regression-detection (catch issues before later beads build on them); the full `<test>` run happens once before the push.
- **Targeted tests per bead are non-negotiable.** "Reduced gates" between beads in a batch never means "no tests." Each bead must run surface-relevant targeted tests (from the gate profile's targeted-tests-by-surface list). If the subagent can't identify any, that's a signal the change is broader than expected — pause and reconsider scope, don't skip.
- **Red before green on defect beads.** A test written after its fix encodes the fix's own assumptions and can stay green while the bug is live. For a bead fixing a defect or regression, observe the reproducing test fail for the stated reason before fixing; the retro names that observed failure, not just "tests pass".
- **One commit per bead, always.** Every closed bead = exactly one commit on the branch. The retro file is written BEFORE the close ceremony (crash recovery anchor); the close happens via `bd update --status closed --body-file <retro-file>` BEFORE staging, so the close is folded into the implementation commit.
- **Lean path for small / simple queues.** Skip the formal decisions doc when all decisions are pre-baked. Skip the formal dispatch plan when queue size <5 or all beads share a bucket. State the decision in one line in chat and proceed. The ceremony is for runs that earn it.
- **Surface implicit decisions, don't bury them.** If the subagent will silently choose, name it — either inline in chat (small queues) or in the decisions doc (large queues). The user can override or trust the default, but should know it exists.
- **Visible progress beats opacity.** Subagents write JSONL progress events. For any subagent past 5 minutes, the orchestrator must arm `Monitor` on the progress file so the user sees per-phase events live. Long opaque runs erode trust.
- **Subagents discard context on return.** This is why the skill prefers fresh subagents over `/compact`. Batch subagents (fast lane, shared-surface) compromise — they hold context for N beads but free it on return.
- **Don't read source files in the orchestrator.** Almost all code reads happen inside subagents. Exception: Phase 2a may need to read source to make decision examples concrete.
- **Capture decisions in the bead, not just in chat.** The bead description (with SELECTED markers and preserved alternatives) is the durable record. Decisions doc and dispatch plan are session-scratch.
- **Fail loudly, recover quietly.** If a subagent blocks, log it and move on. If three consecutive block, stop and surface — that's a systemic problem. If a subagent crashes mid-step, follow the recovery procedure (retro file is the anchor).
- **Route by tier, and report every escalation.** A bead's `tier:` label is grooming's *prediction* of the capability the work needs; resolve it to a model via the project's `## model lineup` and dispatch there (Phase 2.5 → Tier selection). Assign nothing below the tier; escalate one tier on repeated failure; never downgrade to save cost mid-run. Every escalation is a required row in the Phase 4 report — it is the only routing evidence a run produces. Note the asymmetry: escalations expose under-routing, while over-routing succeeds silently and is paid for on every run, so a zero-escalation run is a prompt to examine the routing, not a clean bill of health. Beads marked "do not delegate" run in the orchestrator's own loop.
- **Project rules win over skill defaults.** The gate profile is the source of truth for build/test/lint/format commands and any custom gates the project adds. If `CLAUDE.md`/`AGENTS.md` bans `TaskCreate`, skip it (use inline status). If the project has its own commit message conventions, follow them. The skill is opinionated but not authoritative.
- **Auto-chain waves under standing authorization.** When dependencies mean closing one wave unblocks the next, ask ONCE (at run start or first chain point) for standing approval: "re-resolve the queue after each push and dispatch the next wave until dry, implicit decisions defaulted, explicit `decision-required` beads still stop." With it granted, waves flow without per-wave prompts; the user gets a report per wave instead of a question per wave. Newly-filed beads that enter the queue mid-run (subagents filing adjacent issues) are admitted only by the mid-run scope test (Phase 1.5 step 4): in-scope when a bead already in the run can't be finished without them, not merely when they match the filter. Admitted beads pass the grooming gate before joining a wave; out-of-scope finds are filed, linked, and surfaced for the user — never dispatched. The standing authorization finishes the launched scope; it never enlarges it.
- **Keep an environmental-failure ledger; never re-diagnose per wave.** Suites that fail — or silently PASS for environmental reasons — in the local sandbox get diagnosed ONCE, recorded in the gate profile's ledger (Phase 0 step 1), then the verdict is passed into every subsequent subagent prompt as fact. The silent-pass case is the dangerous one: a gate can go green without exercising your change (a harness running a stale prebuilt binary a plain test command didn't rebuild, a suite skipped by default, a cached result). When a gate *can* pass without touching your change, force a clean rebuild and confirm it actually ran before trusting it. Heavy suites (integration, codegen, anything needing sockets/toolchains the sandbox lacks) run locally only when a bead touches the surface they cover; when a PR is open, CI is the backstop.
- **Guard against foreign dirt and index races.** Multiple agents can share the tree (a diagnosis agent's uncommitted fixes, an in-flight batch's staged `.beads/issues.jsonl`). Batch prompts must name any known foreign files as untouchable; the orchestrator commits its own work by explicit pathspec (`git commit -- <files>`) so a concurrently-staged tracker ledger rides with the next bead commit instead of being swept up mid-ceremony. The same guard applies to tracker *claims*: a subagent may only reset a claim it made itself — a bead found `in_progress` that this agent didn't claim belongs to a live sibling — and the orchestrator (never a subagent) owns reconciling claims across waves.
- **Wire findings capture when the repo has it.** If `rfcs/findings-capture.md` exists (retro-spec protocol), every batch prompt includes: record `retro:`-prefixed `bd comment`s on the bead being worked, 1-3 sentences, friction only, zero analysis. The feature's final DoD bead runs the one aggregated retro-spec; closed-bead `## Retro` sections serve as the findings record for beads that predate the wiring.

## Retro this skill

WHEN a run of `bead-bonanza` struggled — you improvised around its
guidance, the gate profile it prescribes was wrong or missing, a command
it told a subagent to run misbehaved, or an instruction here pointed you
the wrong way — offer to run `retro-skill` on `bead-bonanza` while the
session is fresh, to capture the friction as an evidence-backed RFC. A
clean run needs nothing; never run it unprompted.
