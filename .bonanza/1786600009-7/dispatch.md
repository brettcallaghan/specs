# dispatch.md — run 1786600009-7

## Run manifest (orchestrator-authored, before grooming/planning dispatch)

| Field | Value |
|-------|-------|
| Run id | `1786600009-7` |
| Epic | `agent-plugins-61f` |
| Branch | `claude/triage-rfc-self-report-anchors-7vlncg` |
| Remote | `origin` (push needs explicit refspec) |
| Default branch | `main` |
| RUN_DIR | `/tmp/claude/bead-bonanza/1786600009-7` |
| SCRIPTS | `/home/user/agent-plugins/plugin/skills/bead-bonanza/scripts` |
| Trailers file | `$RUN_DIR/trailers.txt` |
| Base sha | `b8e9436` |
| Model lineup | light=`haiku` standard=`sonnet` advanced=`opus` frontier=`fable` |
| Wave preview | none (no C4 docs cover this surface) — bucketed on file paths; see "Measured file map" below |

## Enumerated bead list (the plan IS the queue — never a filter)

agent-plugins-61f.1, agent-plugins-61f.2, agent-plugins-61f.3,
agent-plugins-61f.4, agent-plugins-61f.5, agent-plugins-61f.6,
agent-plugins-61f.7

## Pre-flight stdout block

RUN_DIR='/tmp/claude/bead-bonanza/1786600009-7'
SCRIPTS='/home/user/agent-plugins/plugin/skills/bead-bonanza/scripts'
BB_GATES_BLOCK=present
BB_LINEUP_BLOCK=present
BB_SUBAGENT_MODEL_OVERRIDE='unset'
BB_PR_CI=yes
BB_TREE_CLEAN=yes
BB_PUSH_NEEDS_REFSPEC=yes
BB_GIT_CONFIG_WRITABLE=yes
BB_BRANCH='claude/triage-rfc-self-report-anchors-7vlncg'
BB_REMOTE='origin'
BB_UPSTREAM='none'
BB_DEFAULT_BRANCH='main'
BB_BD=ok
BB_JQ=ok
BB_DOLT_WRITE=yes
BB_WORKTREE_BASEREF='head'

## Gate profile (from AGENTS.md)

- Format: `cd cli && gofmt -l .`
- Lint: `scripts/check-skills.sh` (+ `cd cli && go vet ./...` when cli/ changed)
- Build: `cd cli && go build ./...`
- Test (full): the three checks run INDEPENDENTLY, never chained with `&&`
- Custom: `scripts/generate-schemas.sh --check` (only when schemas/*.strata change)

Targeted tests by surface, for this run's beads:
- `plugin/skills/*/SKILL.md` -> `scripts/check-skills.sh`
- `plugin/skills/*/references/**` -> `scripts/sync-shared-references.sh --check && scripts/check-skills.sh`

## Gate baseline (MEASURED this run, 1786600009-7 — not quoted)

| Gate | Exit | Result |
|------|------|--------|
| `cd cli && gofmt -l .` | 0 | pass |
| `scripts/check-skills.sh` | 0 | pass |
| `cd cli && go build ./...` | 0 | pass |
| `scripts/check-skills.sh --strict` | **1** | **0 error(s), 1 warning(s)** — KNOWN RED |
| `scripts/sync-shared-references.sh --check` | 0 | pass |
| `scripts/sync-strata-docs.sh --check` | 0 | pass |

The single warning: `plugin/skills/bead-bonanza/SKILL.md: body is 51921 bytes
(~12980 tokens); above ~20000 the tail is truncated from context after a
compaction`.

**The bar for every bead is 0 errors and 1 warning — NOT exit 0.** Five beads in
this run ADD text to that exact file, so the byte figure will rise and the
warning will persist. That is at baseline, not a regression. Do not attempt to
fix it; it is editorial scope owned elsewhere. Capture `$?` directly, never
through a pipe.

## Environmental ledger (carried into every dispatch prompt as fact)

- `check-skills.sh --strict` is KNOWN RED on a warning. See baseline above.
- Redirect gate output OUTSIDE the worktree — `bb-close-bead.sh` stages
  everything in the tree by design, so a file written there ships in the close
  commit. Use `$RUN_DIR/<id>.<gate>.out`.
- `bd`'s JSONL export lags the database; `bb-close-bead.sh` handles it. Never
  read `.beads/*.jsonl` as tracker state.
- No PR will be opened this run (repository instruction: never open a PR
  unless explicitly asked). CI backstop is therefore NOT available; local
  gates are the whole gate.

## Standing authorization

The user enumerated all seven beads and asked for them to be implemented. Waves
auto-chain to completion without per-wave prompts. Implicit decisions are
defaulted and recorded in `retro:` comments. No bead in this run carries a
`decision-required` label. Emergent work is admitted only by the Phase 1.5
step-4 necessity test.

## KNOWN LIVE DEFECT affecting this run's own dispatch

`references/subagent-prompts.md`'s "Shared batch preamble" tells every batch
subagent to create each worktree from `HEAD`. Inside a batch `HEAD` does not
move, so all beads would branch from the same commit and conflict on the shared
file. `agent-plugins-61f.2` fixes exactly this, and it is IN this run — so the
fix is not yet in effect while the batch that carries it is dispatched.

**Override, applied by the orchestrator to every batch prompt this run:** the
first bead branches from `HEAD`; each later bead branches from the PREVIOUS
bead's branch. Recorded here so recovery does not "correct" it back.

## Durable wake mechanisms

- In-session `Monitor`: armed per dispatch over the bead progress JSONL.
- Durable out-of-session trigger: `trig_01QnAgQNJ5zxJTZVT2xCxQQb`, fires
  2026-08-13T06:18:00Z, carries RUN_DIR + epic + the two facts recovery must not
  "correct". **Phase 4 step 6a MUST cancel it**, and the final report states its
  disposition.

## Amendments

(append-only)

---

# Wave plan (plan-sprint, Phase 2.5)

Everything below this rule is plan-sprint's. The manifest above is the
orchestrator's and was not edited.

## Signal inventory — what was available to bucket on

| Signal | Status |
|---|---|
| Wave preview | `none (no C4 docs for this surface)` |
| `c4_changes` / `c4_impacts` | **absent on all seven beads** — metadata keys are only `est_size, grooming_status, tier, tier_reason` |
| File paths (from each bead's Design) | **the whole signal** — primary and sole basis for bucketing |
| Dependency edges | present and load-bearing — two hard `blocks` chains |

The repo does carry C4 docs under `specs/architecture/`, but they describe the
Spec CLI / Document Core / Spec Profile. No bead in this run touches that
surface; all seven edit skill markdown under `plugin/skills/**`. Grooming
therefore skipped C4 tagging correctly, and **absent metadata here means "not
applicable", not "no collisions"** — the file scan below carries the full load.

## Measured file map (read from each bead's Design, not inferred)

| Bead | Files it edits |
|---|---|
| 61f.1 | `bead-bonanza/SKILL.md` |
| 61f.2 | `bead-bonanza/SKILL.md`, `bead-bonanza/references/subagent-prompts.md` |
| 61f.3 | `bead-bonanza/SKILL.md` |
| 61f.6 | `bead-bonanza/SKILL.md`, `bead-bonanza/references/progress-streaming.md` |
| 61f.7 | `bead-bonanza/SKILL.md` **(owns the single 1.4 → 1.5 bump)** |
| 61f.4 | `spec/references/work-order-definition-of-ready.md`, `spec/references/work-order-template.md` (sources) + 4 regenerated vendored copies |
| 61f.5 | `groom-beads/SKILL.md` **(owns the single 1.2 → 1.3 bump)** |

61f.4's regeneration writes exactly four destinations (from
`scripts/sync-shared-references.sh`'s `MAP`):
`bead-bonanza/references/shared/work-order-definition-of-ready.md`,
`groom-beads/references/shared/work-order-definition-of-ready.md`,
`groom-beads/references/shared/work-order-template.md`,
`groom-feature-blueprint/references/shared/work-order-template.md`.

**Chain A ∩ Chain B = ∅.** 61f.4 writes into `bead-bonanza/references/shared/`
while chain A edits `bead-bonanza/SKILL.md` and two files directly under
`bead-bonanza/references/` — same skill bundle, no shared file. The chains are
collision-free; they are sequenced for the reason in the wave 2 row, not
because they conflict.

## Wave table

| Wave | Bead | Class | Bucket | Tier | Push | Reason |
|------|------|-------|--------|------|------|--------|
| 1 | agent-plugins-61f.1 | shared | shared-batch #1 (chained) | standard | defer | co-edit: `bead-bonanza/SKILL.md`; chain head, branches from HEAD |
| 1 | agent-plugins-61f.2 | shared | shared-batch #1 (chained) | standard | defer | co-edit: `bead-bonanza/SKILL.md`; `blocks` from .1 — branch from .1's branch |
| 1 | agent-plugins-61f.3 | shared | shared-batch #1 (chained) | standard | defer | co-edit: `bead-bonanza/SKILL.md`; `blocks` from .2 — branch from .2's branch |
| 1 | agent-plugins-61f.6 | shared | shared-batch #1 (chained) | standard | defer | co-edit: `bead-bonanza/SKILL.md`; `blocks` from .3 — branch from .3's branch |
| 1 | agent-plugins-61f.7 | shared | shared-batch #1 (chained) | standard | defer | co-edit: `bead-bonanza/SKILL.md`; MUST land last — owns the sole 1.4 → 1.5 bump and its AC requires the other four edits present in the tree first |
| 2a | agent-plugins-61f.4 | shared | sequential #2a | **advanced** | push_per_bead | co-edit: `groom-beads` bundle with .5 (sources + regenerated `references/shared/`); chain head, branches from wave-1 sprint head |
| 2b | agent-plugins-61f.5 | shared | sequential #2b | standard | push_per_bead | co-edit: `groom-beads` bundle; `blocks` from .4 — the DoR rule must exist before this bullet rejects against it; branch from .4's branch |

Tier resolves at dispatch via the caller-supplied lineup (see the note under
"Planning inputs not present in the manifest header"): 61f.4 → `opus`, the
other six → `sonnet`.

### Dispatch mode, and why

**Both buckets are chained shared-surface batches. Neither is plain sequential,
and there is no concurrent wave.**

**Wave 1 — batch, not five sequential dispatches.** Five beads co-edit one file
under a hard `blocks` chain. Plain sequential would cost five full `<test>`
runs, each of which is three independently-invoked scripts; the batch costs one,
saving four. The batching threshold is met with room to spare: ≥2 beads on one
surface, and the decisions are independent — every bead's replacement text is
already drafted in its Design, and version-bump ownership is settled in advance
(four beads explicitly disclaim the bump, .7 owns it). The chain is pure
*ordering*, which a batch executes sequentially by construction.

**Wave 1 is dispatched under the orchestrator's chained-base override, and that
override is load-bearing rather than an optimization.** With the shipped
preamble (every worktree from `HEAD`) all five beads would branch from one
commit, the merge would fast-forward only 61f.7's branch, and the other four
beads' edits to `SKILL.md` would be silently discarded. The batch carries its
own tripwire: 61f.7's acceptance criterion requires confirming the other four
changes are present in `SKILL.md` **before** bumping, so a dropped override
surfaces as a failed bead rather than a quiet loss. A batch subagent holds one
worktree at a time; chaining is compatible with that, since bead N branches from
bead N−1's *branch*, which outlives its worktree.

At merge, fast-forward **61f.7's branch only** — it carries its four
predecessors. Do not merge the five branches in turn; that is exactly the merge
policy 61f.2 is adding to `SKILL.md` in this same wave.

**Wave 2 — sequenced after wave 1 rather than run concurrently with it.** The
two chains share no file, and `references/subagent-prompts.md` documents working
rebase-then-fast-forward mechanics for concurrent waves, so concurrency was
available. It was declined for a concrete reason, not caution: **61f.2 — the fix
for the batch-preamble defect — lands in wave 1.** Once wave 1 merges,
`references/subagent-prompts.md` contains the correct chaining instruction, so
wave 2's batch prompt is correct by construction and needs no hand-override. The
orchestrator MUST re-read the preamble from the sprint branch when assembling
wave 2's prompt for this to hold. Concurrency would have bought wall-clock only
while requiring the manual override to be applied correctly to two in-flight
batches at once — trading the run's one known defect for the run's only
non-mechanical dispatch step, twice over.

### Gates for this run

Per-bead between beads in a batch: `<format>`, `<lint>`, `<build>`, plus the
surface-targeted test the manifest names (`scripts/check-skills.sh`; add
`scripts/sync-shared-references.sh --check` for any bead touching
`references/**`). One full `<test>` at the END of each batch — the three checks
run **independently, never chained with `&&`**.

**The bar is the measured baseline: 0 errors / 1 warning, NOT exit 0.** All five
wave-1 beads add text to `plugin/skills/bead-bonanza/SKILL.md`, so the body-size
warning's byte figure rises and the warning persists. That is at baseline. Do
not attempt to fix it.

The custom gate `scripts/generate-schemas.sh --check` **does not apply to any
bead in this run** — no Design names a `.strata` file, `schemas/`, or the
generator. Do not run it.

## Environmental ledger (planning addenda)

Numbered `P<n>` so they do not collide with the orchestrator's bulleted entries
above. Cite as "ledger entry P1".

1. **P1** — The batch-preamble defect is live in the working tree, so wave 1's
   chained-base override is required. Proved by `grep -n -A14 "Shared batch
   preamble" plugin/skills/bead-bonanza/references/subagent-prompts.md`;
   observed line 178 reads `git -C <repo> worktree add -b <suffix>
   <repo>/../.bb-worktrees/<suffix> HEAD` with no predecessor-chaining clause.
   Conditions: working tree at `b8e9436`, clean.
2. **P2** — Chain A and chain B share no file. Proved by reading all seven
   Designs plus `scripts/sync-shared-references.sh`'s `MAP` table; observed
   61f.4's four regeneration destinations (listed above) are disjoint from chain
   A's three files. A cross-chain rebase therefore touches disjoint paths.
3. **P3** — No bead touches the schema surface. Proved by
   `grep -nE "\.strata|schemas/|generate-schemas"` over the full `bd show` text
   of all seven beads; observed zero matches. The custom gate is out of scope.
4. **P4** — No bead carries C4 metadata. Proved by `bd list --json | jq` over
   the seven; observed the union of metadata keys is exactly
   `est_size, grooming_status, tier, tier_reason`. C4 bucketing does not apply;
   file paths are the whole signal.
5. **P5** — Concurrent waves are mechanically supported (rebase onto the updated
   sprint head, then `merge --ff-only`). Proved by reading
   `plugin/skills/bead-bonanza/references/subagent-prompts.md` lines 340–358.
   Recorded because it establishes that wave 2's sequencing is a deliberate
   choice, not a capability limit.

## Planning inputs not present in the manifest header

Flagged, not fixed — the header is the orchestrator's and was left untouched.
The format in `references/dispatch-plan.md` expects these as manifest rows:

- **Model lineup** — absent from the header. Caller-supplied for this run:
  light=`haiku`, standard=`sonnet`, advanced=`opus`, frontier=`fable`. Tier
  resolution above uses these values.
- **Base sha** — absent from the header. Observed at planning time: `b8e9436`,
  tree clean.
- **Wave preview** — absent from the header. Correct value:
  `none (no C4 docs for this surface)`.

### 2026-08-13 — Monitor cap MEASURED; triage's dismissal of it was wrong

Arming the wave-1 Monitor produced a direct measurement that contradicts the
triage reasoning behind bead `agent-plugins-61f.6`:

    requested: timeout_ms = 3600000, persistent = true
    reported:  "Monitor started (task bhjo96kee, timeout 1800000ms)"

3,600,000 ms requested, 1,800,000 ms reported. That is the 30-minute cap the RFC
(`rfc-self-report-citations-and-batch-base.md`, P4) reported and that triage
declined to adopt, on the grounds that it could not be verified and that the
Monitor tool schema documents `timeout_ms` max 3600000. The schema documents the
accepted INPUT range; the runtime applies a lower cap. Both can be true, and the
RFC's observation was correct.

`61f.6`'s Design instructs the implementer to state NO numeric cap and says
triage "could not reconcile" the two figures. That premise is now false. The
bead was in flight when this was measured, so it is being allowed to land as
designed rather than redesigned mid-batch; a follow-up bead carries the
correction.

Open question this raises, NOT yet settled: `persistent: true` is documented as
"no timeout", yet the runtime still reported one. If the Monitor dies at ~30
minutes despite `persistent: true`, then `61f.6`'s core remedy is insufficient
and the re-arm fallback is the whole rule after all. Watch task `bhjo96kee`.

### 2026-08-13 — wave 2 SPLIT into two sequential dispatches (was: one chained batch)

The plan bucketed `61f.4` + `61f.5` as "shared-batch #2 (chained)" at tier
`advanced` (the higher of the two). Reclassified to two sequential single-bead
dispatches at their own tiers: `61f.4` -> `opus`, `61f.5` -> `sonnet`.

Reason, from the skill's tier-batching rule: batching exists to avoid repeated
full `<test>` runs, and "WHERE the gate baseline shows the suite completing in
seconds, the saving is negligible and the correct tier wins." This run's
measured suite is three shell scripts completing in seconds, so batching would
buy almost nothing while dragging `61f.5` a full tier above its assignment.

Second benefit: sequential means the orchestrator merges `61f.4` before `61f.5`
is dispatched, so repo HEAD genuinely advances and `61f.5`'s worktree branches
from HEAD correctly — no hand-applied chaining override, which is the run's one
non-mechanical dispatch step. Wave 1's fix (`61f.2`) is now merged, so the
preamble read from the sprint branch is correct for any future batch.

Wave table rows for `61f.4` and `61f.5` updated in place: Bucket
`sequential #2a` / `sequential #2b`, Tier `advanced` / `standard`,
Push `push_per_bead`.

