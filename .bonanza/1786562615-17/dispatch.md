# Dispatch plan — run 1786562615-17

| Field | Value |
|-------|-------|
| Run id | `1786562615-17` |
| RUN_DIR | `/tmp/claude/bead-bonanza/1786562615-17` |
| Branch | `claude/bead-bonanza-refinement-261qvk` |
| Base sha | `b8b36d106f81e62e7379f39e79fc46360f70ca6d` (`b8b36d1`) |
| Draft PR | https://github.com/brettcallaghan/agent-plugins/pull/20 (opened at run start; every wave push gets full CI) |
| Gate profile | format=`cd cli && gofmt -l .` lint=`scripts/check-skills.sh` (+ `cd cli && go vet ./...` when `cli/` changed) build=`cd cli && go build ./...` test=`scripts/check-skills.sh --strict` + `scripts/sync-shared-references.sh --check` + `scripts/sync-strata-docs.sh --check` (+ `cd cli && go test ./...` when `cli/` changed) regen=`scripts/generate-schemas.sh` custom=`scripts/generate-schemas.sh --check`, `python3 scripts/validate-corpus.py` |
| Gate baseline | format=pass lint=pass build=pass **test=FAIL** (known-red `check-skills.sh --strict`) |
| Model lineup | light=`haiku` standard=`sonnet` advanced=`opus` frontier=`fable` |
| Standing authorization | Auto-chain waves until dry; re-resolve queue after each push; implicit decisions defaulted and reported; `decision-required` still stops. Scope fixed to the 17 enumerated beads. Draft PR authorized and opened. |
| Wave preview | `not returned` — queue passed the Phase 1.5 cheap check, so no grooming subagent ran. Partition derived by the orchestrator from C4 metadata. |

## Environmental ledger

Facts established this run. No wave re-diagnoses these.

1. `check-skills.sh --strict` is **red at baseline on a warning, not an error** — proved by `scripts/check-skills.sh --strict` at pre-flight, observed `0 error(s), 1 warning(s)` (bead-bonanza SKILL.md 859 lines vs an "under 500" guideline), exit 1. **The bar for every bead in this run is "no findings beyond this baseline", not exit 0.**
2. The documented full-test gate `--strict && sync-shared-references --check && sync-strata-docs --check` **short-circuits**: the first command always exits 1, so the two `--check` steps never execute — proved by reading AGENTS.md:52 against ledger entry 1. **This is bead `txq`, wave 1.** Until it lands, every wave must run the three checks *independently* rather than as the documented `&&` chain, or two thirds of the gate is unverified.
3. Dolt remote write access is **available** — proved by `pre-flight-check.sh`'s probe push to `refs/dolt/preflight-probe-$$` and its deletion, observed exit 0 (`BB_DOLT_WRITE=yes`). Closures are durable; `bb-close-bead.sh` pushes at every close.
4. A bare `git push` resolves — proved by `pre-flight-check.sh` reading `origin/claude/bead-bonanza-refinement-261qvk` as configured upstream (`BB_PUSH_NEEDS_REFSPEC=no`). Verify landings with `git rev-parse origin/<branch>` after fetch, never `git status -sb`.
5. `worktree.baseRef` was **unset at pre-flight and has been set to `head`** in `/root/.claude/settings.json` — without it every worktree branches from `main` and silently drops the sprint branch. It is user-global state this repo does not own: **re-verify before every wave.**
6. `strata` is installed but its freshness against upstream is **unverified this session** (SessionStart hook could not fetch; stamp matches local HEAD `ce18d66f`). `scripts/generate-schemas.sh --check` may therefore report committed-correct artifacts as stale. **Never "fix" that by committing regenerated output** — no bead in this queue touches `schemas/*.strata`, so this gate should not fire at all.
7. `sync-strata-docs.sh --check` reports SKIP (exit 0) when no strata checkout is present — a SKIP is not a pass.

## Wave table

The plan names beads, never a filter.

| Wave | Bead | Class | Bucket | Tier | Model | Push | Reason |
|------|------|-------|--------|------|-------|------|--------|
| 1 | txq | shared | shared-batch #1 | standard | sonnet | defer | co-edit: `AGENTS.md` with d6x. Runs first — it is the blocker on epic 41z and closes ledger entry 2 |
| 1 | d6x | shared | shared-batch #1 | standard | sonnet | defer | co-edit: `AGENTS.md` |
| 1 | d70 | isolated | sequential | advanced | opus | defer | zero co-edit edges with any queued bead; concurrent with shared-batch #1. Blocks 9ko |
| 2A | q1h | shared | shared-batch #2 | standard | sonnet | defer | co-edit: `scripts/pre-flight-check.sh` |
| 2A | td1 | shared | shared-batch #2 | standard | sonnet | defer | co-edit: `scripts/pre-flight-check.sh` (over-routed light→standard, batch tier) |
| 2A | 9p4 | shared | shared-batch #2 | standard | sonnet | defer | co-edit: `scripts/pre-flight-check.sh` + `SKILL.md` |
| 2A | 8mv | shared | shared-batch #2 | standard | sonnet | defer | co-edit: `scripts/pre-flight-check.sh`. Must precede b13 — b13 discloses what 8mv leaves behind |
| 2B | b13 | shared | shared-batch #3 | standard | sonnet | defer | co-edit: `SKILL.md` + `bb-close-bead.sh`; sequencing: after 8mv/9p4/q1h/td1, consumes `pre-flight-check.sh` |
| 2B | 0z0 | shared | shared-batch #3 | standard | sonnet | defer | co-edit: `SKILL.md`; sequencing: after b13, consumes `bb-close-bead.sh` (over-routed light→standard, batch tier) |
| 2B | viy | shared | shared-batch #3 | standard | sonnet | defer | co-edit: `references/operating-principles.md` with 8ef.1 |
| 3 | 8ef.1 | shared | sequential | advanced | opus | defer | epic 41z decision 4 binds children to non-concurrent dispatch; `blocks` edge → 8ef.2 |
| 3 | 8ef.2 | shared | sequential | advanced | opus | defer | `blocks` edge: after 8ef.1; co-edit `SKILL.md` |
| 3 | 8ef.3 | shared | sequential | standard | sonnet | defer | `blocks` edge: after 8ef.2; co-edit `SKILL.md` + `subagent-prompts.md` |
| 4 | new | shared | shared-batch #4 | advanced | opus | defer | co-edit: `SKILL.md` + `references/recovery.md`. Dedup (wave 3) runs first so extraction moves already-deduped text |
| 4 | b5z | shared | shared-batch #4 | advanced | opus | defer | co-edit: `SKILL.md` + `references/dispatch-plan.md` |
| 5 | 9ko | isolated | fast-lane #1 | light | haiku | defer | zero co-edit edges; `blocks` edges: after 41z (wave 3) and d70 (wave 1) |
| 5 | x8r | shared | fast-lane #1 | light | haiku | defer | co-edit: five `SKILL.md` frontmatters. **Strictly last** — its own note: must run after every other skill-changing bead or it needs doing twice |

Between wave 3 and wave 4, epic **41z** closes (all three children done, its own blocker `txq` done), which is what unblocks `x8r` and `9ko`.

### Why no concurrent wave beyond wave 1

Fifteen of the seventeen beads carry a co-edit edge on
`plugin/skills/bead-bonanza/SKILL.md`, its references, or its scripts. Epic
41z's design decision 4 is binding and independently reaches the same verdict
for its children: *"These MUST run as one shared-surface batch or strictly
sequentially — never as a concurrent worktree wave. Dispatching them
concurrently will produce conflicting edits to the same regions."*

Only `d70` and `9ko` have zero co-edit edges. `d70` is dispatched concurrently
with wave 1; `9ko` is dependency-pinned to wave 5, where batching it with `x8r`
costs one gate run instead of two.

### The line-number hazard

`8ef.1`, `8ef.2`, `b5z` and `new` cite exact `SKILL.md` / `recovery.md` line
ranges in their designs (`:570-642`, `:871-879`, `:403-411`, `:294`, `:25-31`,
`:251-262`). Every wave before them shifts those lines. **Every dispatch prompt
for a line-citing bead carries an explicit instruction: line numbers are as-of
grooming; locate the passage by its content, and report the drift in a `retro:`
comment.** This is not optional — a subagent that trusts a stale line number
edits the wrong passage silently.

## Amendments

- **2026-08-12** — **Dependency repair before wave 1.** `agent-plugins-8ef.1`
  carried a blocking edge to `agent-plugins-41z`, its own parent epic. That is a
  deadlock: an epic cannot complete until its children do, so 8ef.1 could never
  become ready and the whole 41z subtree (8ef.1→8ef.2→8ef.3, plus x8r and 9ko
  downstream) was undispatchable. Removed that edge and wired `8ef.1 depends on
  txq` — which is exactly what 41z itself depends on, so the intent is preserved.
  Removing the edge also severed the parent-child link (bd surfaces parent-child
  as a dependency); it was restored with `bd update --parent`, and 41z reads 0/3
  children again. `bd dep cycles` clean. Verified: 8ef.1 now blocked by txq only.
- **2026-08-12** — **Readiness-gate exception, `agent-plugins-b13`.** Its
  `design` field is empty, which the Definition of Ready requires for a bug.
  Admitted without a grooming round-trip: its description enumerates all five
  defects with file-and-line citations, it carries repro steps, three acceptance
  criteria, a tier and a tier_reason, and est_size S. It passes the fresh-agent
  test on description+AC alone. Recorded here rather than silently.
- **2026-08-12** — **Tier resolved from `metadata.tier`, not the `tier:` label,
  where they disagree.** Four beads carry a `tier:standard` label; two of them
  (`8ef.1`, `8ef.2`) carry `metadata.tier=advanced` with a `tier_reason` reading
  "judgment-heavy prose surgery", and one (`txq`) carries `metadata.tier=light`.
  The `tier_reason` text matches `metadata.tier` in every case, so the labels are
  stale. Applied the rule "never run a bead below its tier" — took the **higher**
  of the two everywhere: 8ef.1/8ef.2 → advanced, txq → standard. Worth an RFC to
  `groom-beads`: two tier channels that can disagree is a routing hazard.
