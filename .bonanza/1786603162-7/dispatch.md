# bead-bonanza run 1786603162-7

## Pre-flight (measured, not asserted)

```
RUN_DIR='/tmp/claude/bead-bonanza/1786603162-7'
SCRIPTS='/home/user/agent-plugins/plugin/skills/bead-bonanza/scripts'
BB_GATES_BLOCK=present
BB_LINEUP_BLOCK=present
BB_SUBAGENT_MODEL_OVERRIDE='unset'
BB_PR_CI=yes
BB_TREE_CLEAN=yes
BB_PUSH_NEEDS_REFSPEC=no
BB_GIT_CONFIG_WRITABLE=yes
BB_BRANCH='claude/validate-corpus-spec-cli-0t0cxo'
BB_REMOTE='origin'
BB_UPSTREAM='origin/claude/validate-corpus-spec-cli-0t0cxo'
BB_DEFAULT_BRANCH='main'
BB_BD=ok
BB_BD_DB='/home/user/agent-plugins/.beads/embeddeddolt'
BB_JQ=ok
BB_DOLT=absent
BB_DOLT_TELEMETRY=na
BB_DOLT_WRITE=yes
BB_WORKTREE_BASEREF='head'
BB_GATE_BASELINE=captured
BB_GATE_FORMAT=pass
BB_GATE_LINT=pass
BB_GATE_BUILD=pass
```

`BB_WORKTREE_BASEREF` was `unknown` on the first pre-flight run (exit 1,
BLOCKED). Fixed by writing `{"worktree":{"baseRef":"head"}}` to
/root/.claude/settings.json — the file did not previously exist — and
re-running pre-flight, which then exited 0. Re-verify before EACH wave: it is
user-global state no repository owns.

## Gate profile (AGENTS.md `## bead-bonanza gates`)

- **Format:** `cd cli && gofmt -l .` — must print nothing
- **Lint:** `scripts/check-skills.sh`; when `cli/` changed, `cd cli && go vet ./...`
- **Build:** `cd cli && go build ./...` (docs-only beads: no build step)
- **Test (full suite):** the three repo checks run INDEPENDENTLY, never chained
  with `&&` (check-skills.sh --strict is known red and would short-circuit the
  others):
  ```
  scripts/check-skills.sh --strict;           s1=$?
  scripts/sync-shared-references.sh --check;   s2=$?
  scripts/sync-strata-docs.sh --check;         s3=$?
  ```
  Plus, when `cli/` changed, `cd cli && go test ./...`
- **Custom gates:**
  - `scripts/generate-schemas.sh --check` — only after changing `schemas/*.strata`
  - `python3 scripts/validate-corpus.py` (needs `jsonschema`, `pyyaml`)
  - `scripts/check-template-drift.py` — added this branch; see baseline below

## Gate baseline — the bar is "no findings beyond this", not exit 0

| Gate | Baseline |
|------|----------|
| `cd cli && gofmt -l .` | pass (silent) |
| `scripts/check-skills.sh` | pass |
| `cd cli && go build ./...` | pass |
| `scripts/check-skills.sh --strict` | **exit 1 — 0 errors, 1 warning**: bead-bonanza/SKILL.md body 51921 bytes. KNOWN RED, pre-existing, editorial scope owned elsewhere. Do NOT fix mid-run. |
| `scripts/sync-shared-references.sh --check` | pass |
| `scripts/validate-corpus.py` | pass |
| `scripts/check-template-drift.py` | **exit 1 — 8 findings**: 7 against retro-spec/rfc-template.md (agent-plugins-3xx), 1 against work-order-template.md (agent-plugins-j5n). Both are beads IN THIS RUN; each should REDUCE this count, and the run should end at 0. |

## Model lineup (AGENTS.md `## model lineup`)

light: haiku · standard: sonnet · advanced: opus · frontier: fable

`CLAUDE_CODE_SUBAGENT_MODEL` is unset, so tier assignments ARE honoured and
tier escalations are meaningful evidence.

## Standing authorization

Granted by the user at run start: re-resolve the queue after each push and
dispatch the next wave until dry. Implicit decisions are defaulted and
recorded in each bead's `retro:` comment; a bead carrying an explicit
`decision-required` label still stops. Draft PR: approved. NOT opened by this run — PR #22 ("Deprecate
validate-corpus.py") already exists as a draft on this branch at head
f9afdc7, opened by the repository owner. Creating a second one returned 422.
Every wave push therefore lands on #22 and gets the full CI matrix, which is
the outcome the approval was for. Its title and body are the owner's and were
left untouched.

## Environmental ledger

- `check-skills.sh --strict` is KNOWN RED on a warning. Capture `$?` directly,
  never through a pipe — piping to `tail` reports the pager's status and reads
  as a pass.
- `generate-schemas.sh --check` fails when the installed `strata` predates
  `strata-lang/strata@8997153`. This session's SessionStart hook could NOT
  verify strata against upstream (fetch failed), so the installed binary may be
  stale. No bead in this queue touches `schemas/*.strata`, so this gate is NOT
  required for any of them — do not run it speculatively.
- The PAT cannot push `.github/workflows/**` (no `workflow` scope). No bead in
  this queue edits a workflow file. If one needs to, it must be routed through
  the GitHub App, not `git push`.
- `.beads/*.jsonl` is untracked and is NOT the tracker's transport
  (`origin refs/dolt/data` is). Never stage it.

## Queue (enumerated, not a filter)

agent-plugins-3xx, agent-plugins-j5n, agent-plugins-ziv, agent-plugins-k9j,
agent-plugins-lwu, agent-plugins-qz9.2, agent-plugins-qz9.3

## Measured before planning (removes a risk class)

`schemas/finding.strata:15` declares `alias RuleId: String` — an OPEN alias,
not a closed enum — and the rule constants live in `cli/core/finding/finding.go`,
which is hand-written, not generated. So **agent-plugins-qz9.2 adding a new rule
identifier requires NO `schemas/*.strata` edit and NO `generate-schemas.sh`
run**, and therefore never touches the possibly-stale strata binary the ledger
warns about. Do not run `generate-schemas.sh --check` for any bead in this run.

The regen step that IS relevant is `scripts/regenerate-cli-snapshots.sh`, for
agent-plugins-qz9.3, which changes text output that golden files capture.

## Wave table

Two waves, six dispatch units. `agent-plugins-j5n` and `agent-plugins-k9j` are
the one shared-surface batch (co-edit collision: same script, same directory,
different files); every other row is isolated/sequential or a solo trivial.
`agent-plugins-qz9.2` and `agent-plugins-qz9.3` are the run's only hard
collision (fact 1) and are kept in different waves outright, not merely
different concurrent slots, to match the orchestrator's instruction literally.

| Wave | Bead | Tier | Bucket | Push mode | Rationale |
|------|------|------|--------|-----------|-----------|
| 1 | agent-plugins-ziv | light | trivial fast-lane (solo) | defer | Trivial: one-line predicate-match fix (substring → suffix) in `scripts/validate-corpus.py`. No `c4_changes`, no file overlap with any queued bead. |
| 1 | agent-plugins-3xx | standard | isolated / sequential | — | No `c4_changes`, no file overlap with any queued bead (edits `plugin/skills/retro-spec/references/rfc-template.md` + `SKILL.md`, not vendored). Gate-accounting note, not a merge risk: contributes 7 of the baseline's 8 `check-template-drift.py` findings toward the run's 0 target; `agent-plugins-j5n` contributes the other 1, but the two share no file, so this is verified once at the run's final full test, not sequenced against j5n. |
| 1 | agent-plugins-j5n | standard | shared-surface batch #1 (with k9j), run first | defer | Co-edit collision: same script + same directory as `agent-plugins-k9j` — both run `scripts/sync-shared-references.sh` and commit into `plugin/skills/*/references/shared/` (different files: work-order-template.md's two vendored copies vs. ac-traceability-standards.md's one). Batched sequential rather than concurrent to avoid two worktrees regenerating that tree in parallel. Also edits `scripts/check-template-drift.py` (named exclusion) — the other half of driving that gate to 0. |
| 1 | agent-plugins-k9j | light | shared-surface batch #1 (with j5n), run second | defer | Co-edit collision: same script + same directory as `agent-plugins-j5n` (see j5n's row) — different files (`ac-traceability-standards.md` + its one vendored copy). Run after j5n in the same subagent. |
| 1 | agent-plugins-qz9.2 | standard | isolated / sequential | — | `c4_changes: [module-document-core, container-spec-cli]` — no other bead dispatched in wave 1 shares those components, so it is isolated *within this wave*. Deliberately sequenced ahead of `agent-plugins-qz9.3` (wave 2), which shares the same components — see qz9.3's row. |
| 1 | agent-plugins-lwu | standard | isolated / sequential | — | No `c4_changes`, no file overlap (authors `specs/adrs/ADR-020-*.md` only). `RELATED` to `agent-plugins-k9j`, no `BLOCKS` edge — not a hard dependency. Soft-sequenced into wave 2, after k9j (wave 1) has landed, so ADR-020's citation of the AC-append-only rule describes text that actually exists on the branch rather than a promise. This is an override of the "no dependency" reading of k9j's design note (see below) — see caller line 3. |
| 2 | agent-plugins-qz9.3 | standard | isolated / sequential | — | Co-edit collision: same `c4_changes` (`module-document-core`, `container-spec-cli`) as `agent-plugins-qz9.2` — per fact 1, must not share a wave, so placed in wave 2 outright. Also a sequencing benefit, not just an avoidance: qz9.3 regenerates CLI snapshots (`scripts/regenerate-cli-snapshots.sh`); running it after qz9.2's rule/report-accounting change has landed means the committed goldens capture both changes in one regen pass instead of two. |

### Overrides of an obvious partition

- `agent-plugins-k9j` looks trivial on its own metadata (tier `light`, size
  `XS`, prose addition) and an obvious partition would drop it into the
  trivial fast-lane next to `agent-plugins-ziv`. Pulled into the shared-surface
  batch with `agent-plugins-j5n` instead: both run
  `scripts/sync-shared-references.sh` against the same
  `plugin/skills/*/references/shared/` directory, which is a co-edit signal
  the trivial-lane definition ("no fixture regeneration needed") explicitly
  excludes.
- `agent-plugins-3xx` and `agent-plugins-j5n` look like batch candidates
  because they jointly own the `check-template-drift.py` finding count (7 + 1
  → 0). Kept apart: they share no file and no `c4_changes` — the shared gate
  is an accounting fact to verify at the end, not a merge-collision to batch
  against. Batching them would have cost a worktree-sequencing constraint for
  no real risk reduction.
- `agent-plugins-qz9.2` and `agent-plugins-qz9.3` are the kind of pair the
  shared-surface-batch worked example in `references/dispatch-plan.md`
  batches into one wave number (same `c4_changes`, real code overlap in
  `cli/core/finding`). Not batched, and not even given the same wave number:
  the orchestrator's fact 1 says plainly "they must NOT share a wave," and the
  literal reading was chosen over the framework's more permissive one to
  remove any ambiguity for whoever dispatches this table.
- `agent-plugins-lwu` has no `BLOCKS` edge to `agent-plugins-k9j` (only
  `RELATED`), so an obvious partition would treat them as fully independent
  and free to run in the same wave. Sequenced lwu into wave 2 anyway — see the
  sequencing-constraint line below.

### Sequencing constraint the orchestrator's list did not name

`agent-plugins-lwu` (ADR-020) is required to "name the dependency on AC
identifiers being append-only" — the rule `agent-plugins-k9j` adds to
`ac-traceability-standards.md`. `k9j`'s own Design states it does not need
`lwu` to land first, but says nothing about the reverse; there is no `BLOCKS`
edge between them, only `RELATED`. Treated it as a soft sequencing preference
rather than a hard block: `lwu` is planned into wave 2, after `k9j` (wave 1)
has landed, so the ADR cites the rule as it actually reads on the branch
instead of describing a change still in flight. Nothing in this run breaks if
a dispatcher ignores it and runs them concurrently — the ADR would still meet
its acceptance criteria — but the sequencing removes a small chance of the ADR
text and the standards text drifting during authorship.


## Amendments

### 2026-08-13 — wave 1 dispatched, three deviations from the template

1. **Batch j5n+k9j runs in ONE harness-managed worktree, not one per bead.**
   The shared-surface template has the batch subagent create a worktree per
   bead itself. Deviated because three other subagents are concurrently active
   in this repo, and having a batch agent run `git worktree add` against the
   shared checkout while they work is avoidable risk. The in-flight signal is
   preserved, only coarser: the worktree existing means the BATCH is
   unfinished. Both beads still land as separate commits.

2. **The base-ref ancestor check was given to ALL four wave-1 subagents,
   not only the first.** The template says give it to the first worktree
   subagent; with four dispatched in one message "first" is not well defined,
   and the check is two commands. Session HEAD for the check: f9afdc7.

3. **Per-bead drift-gate targets replace "exit 0".** `check-template-drift.py`
   holds 8 findings owned by two beads running CONCURRENTLY, so neither can
   reach exit 0 alone and both would otherwise reach into the other's files
   to try. Each prompt carries its own target instead:
   agent-plugins-3xx 8 -> 1, agent-plugins-j5n 8 -> 7, run end 0.
   Each is explicitly forbidden from touching the other's file.

Push mode for the whole wave: defer_push. The orchestrator pushes once after
all four units merge, so four concurrent pushes cannot race.

### 2026-08-13 — lwu resequenced out of wave 2 into wave 1

agent-plugins-lwu was planned into wave 2 for one reason: its ADR must cite the
append-only rule agent-plugins-k9j adds, so it needed k9j on the branch first.
k9j merged at c1c4c88, so that precondition is satisfied and the reason to hold
it is gone. Dispatched immediately rather than idling behind agent-plugins-qz9.2.

Safety re-checked before dispatch, not assumed:
  * lwu c4_changes = [] and qz9.2 c4_changes = [module-document-core,
    container-spec-cli] — no shared component, so the concurrency entry
    condition holds.
  * No file overlap: lwu writes specs/adrs/, qz9.2 writes cli/. lwu is
    explicitly forbidden from touching cli/ in its prompt.
  * worktree.baseRef re-verified as `head` before creating the worktree.
  * Ancestor check for lwu uses session HEAD 55cf378 (which contains k9j), so
    a worktree on the wrong base is caught before the ADR cites a rule it
    cannot see.

agent-plugins-qz9.3 is NOT resequenced. It shares both c4_changes components
with qz9.2, which is still in flight, so it stays in wave 2 behind it.

Wave table rows updated in place: lwu now reads wave 1.
