# Run manifest — 1786778523-19

## Pre-flight (stdout block, verbatim)

    RUN_DIR='/tmp/claude/bead-bonanza/1786778523-19'
    BB_GATES_BLOCK=present          BB_LINEUP_BLOCK=present
    BB_SUBAGENT_MODEL_OVERRIDE='unset'
    BB_PR_CI=yes                    BB_TREE_CLEAN=yes
    BB_PUSH_NEEDS_REFSPEC=no        BB_GIT_CONFIG_WRITABLE=yes
    BB_BRANCH='claude/gherkin-adoption-strategy-fikj4i'
    BB_REMOTE='origin'              BB_DEFAULT_BRANCH='main'
    BB_BD=ok                        BB_JQ=ok
    BB_DOLT_WRITE=yes               BB_WORKTREE_BASEREF='head'
    BB_GATE_BASELINE=captured

Verdict: OK — safe to proceed.

## Gate profile

From AGENTS.md `## bead-bonanza gates`. Run Go gates from `cli/`.

- Format: `cd cli && gofmt -l .`   (must print nothing)
- Lint: `scripts/check-skills.sh`; when `cli/` changed, `cd cli && go vet ./...`
- Build: `cd cli && go build ./...`
- Test: the three repo checks run INDEPENDENTLY (never chained with `&&`),
  plus `cd cli && go test ./...` when `cli/` changed.
- Custom: `scripts/generate-schemas.sh --check` after any `schemas/*.strata`
  change.

## Gate baseline (measured on the untouched tree, this run)

| Gate | Result |
|---|---|
| Format | pass |
| Lint (`check-skills.sh`) | pass |
| Build | pass |
| `check-skills.sh --strict` | **exit 1 — 0 errors, 1 warning** (bead-bonanza SKILL.md body size) |
| `generate-schemas.sh --check` | **stale** — committed artifacts lag `schemas/*.strata`; MODEL (wvd.1) is the bead that resolves this |

The bar is **no findings beyond this baseline**, not exit 0.

## Environmental ledger

- `strata` IS available this run: `/usr/local/bin/strata`, `strata-cli 0.7.0`,
  built commit `ce18d66f`. Fixed under `agent-plugins-9v1` (`a02e131`) — the
  clone was losing to an unguarded global credential helper. Do not
  re-diagnose.
- `schemas/feature.strata` COMPILES. Verified this run via
  `generate-schemas.sh --check`, which emits `Feature`, `Criterion`, a
  `features` collection on `SpecIndex`, and `feature.md`. The three
  constructs `agent-plugins-dc6` flagged unverified all work.
- `bd`'s JSONL export lags the DB; `bb-close-bead.sh` already handles it.
- `api.github.com` and GitHub release pages return 403 through the egress
  proxy. Expected; not a fault.

## Model lineup

light `haiku` · standard `sonnet` · advanced `opus` · frontier `fable`

## Run effort setting

unrecorded — the harness exposes no effort setting to this session.

## Commit trailers

`$RUN_DIR/trailers.txt` (Co-Authored-By + Claude-Session). Passed to every
close via `bb-close-bead.sh --trailers-file`.

## Standing authorization

Granted by the user 2026-08-15: re-resolve the queue after each push and
dispatch the next wave until dry. Implicit decisions defaulted and recorded in
`retro:` comments. A bead labelled `decision-required` still stops for the
user. Emergent work is admitted only by the Phase 1.5 step-4 necessity test.

## Amendments

_(appended as the run diverges from the plan)_


---

# Dispatch plan (Phase 2.5)

| Field | Value |
|-------|-------|
| Run id | `1786778523-19` |
| RUN_DIR | `/tmp/claude/bead-bonanza/1786778523-19` |
| Branch | `claude/gherkin-adoption-strategy-fikj4i` |
| Base sha | `a02e1310e8a778283b2f8b6184892fd30af718ad` |
| Queue | 20 beads, enumerated below. **18 dispatchable, 2 blocked out-of-queue.** |
| Wave preview | `overridden — 4 changes; see "Overrides" below` |

## Environmental ledger (appended by planning — continues the header's list)

The header's ledger is the orchestrator's and is unchanged. These are facts
established during Phase 1.5 grooming that a wave must not re-diagnose.

5. **The unguarded global git credential helper is still installed in this
   container** — header ledger entry 1 says the strata clone failure was
   "fixed under `agent-plugins-9v1`", and a reader can mistake that for the
   helper itself having been fixed. It was not; 9v1 fixed the strata *call
   sites*. Proved by `git config --global --get-all credential.helper`,
   observed `!f() { echo username=x-access-token; echo
   "password=$AGENT_PLUGINS_ACCESS_TOKEN"; }; f` — no `get` guard, no host
   scoping, no empty-token guard. `agent-plugins-r19` is the bead that fixes
   it. Do not close r19 as already-done.
6. **`agent-plugins-9v1` is closed, so it does not block `agent-plugins-wvd.1`**
   — proved by `bd show agent-plugins-9v1 --json`, observed `"status":
   "closed"`. `bd list --json` returns only OPEN issues, so a naive blocker
   scan that treats "absent from that list" as "open" will wrongly strand
   wvd.1 and, transitively, the whole `wvd` epic.
7. **`agent-plugins-3wd` ("AC index: extract acceptance criteria into the
   corpus index") is OPEN and is NOT in this run's committed queue** — proved
   by `bd show agent-plugins-3wd --json`, observed `"status": "open"`, and by
   `bd blocked`, observed `agent-plugins-a7u ... Blocked by 3 open
   dependencies: [agent-plugins-wvd.7 agent-plugins-3wd agent-plugins-wvd.8]`.
   This strands `a7u` and, through it, `oq8`. Admitting 3wd is a scope
   decision for the user, not a planning default.
8. **Graph integrity verified this run** — proved by
   `.claude/skills/groom-beads/scripts/bd-health.sh`, observed `swarm
   validate: clean (6 epic(s))`, `orphans: none`, `RESULT: no hard problems`,
   exit 0; and by `bd dep cycles`, observed `No dependency cycles detected`.

## Wave table

18 dispatchable beads across 9 waves. `defer` = commit only, the bucket pushes
at its end; `push` = this bead closes its bucket and the wave pushes here.

| Wave | Bead | Class | Bucket | Tier | Push | Reason |
|------|------|-------|--------|------|------|--------|
| 1 | agent-plugins-krb | shared | fast-lane #1 | light | defer | union: no c4 tags, overlapping snapshot harness (`scripts/regenerate-cli-snapshots.sh`) with 0i6 |
| 1 | agent-plugins-0i6 | shared | fast-lane #1 | standard | push | same snapshot-harness cluster as krb; bucket runs at its highest tier (standard) |
| 1 | agent-plugins-4vz | isolated | sequential | advanced | push | requirements/CI-gate surface; no c4 tags, no file overlap with any queued bead |
| 1 | agent-plugins-r19 | isolated | sequential | advanced | push | `.claude/hooks/session-start.sh` only; c4_changes deliberately `[]` (session-bootstrap infra maps to no C4 component) |
| 2 | agent-plugins-wvd.1 | shared | sequential | advanced | push | same c4_changes: component-corpus, component-findings as wvd.2/.5/.7/.8; regenerates `cli/core/specmodel` and clears the stale `generate-schemas.sh --check` baseline |
| 3 | agent-plugins-wvd.2 | shared | shared-batch #1 | advanced | defer | same c4_changes: component-corpus as wvd.5/.7/.8 |
| 3 | agent-plugins-nu2 | shared | shared-batch #1 | standard | push | union: no c4 tags, but edits the findings model wvd.1 regenerates — same corpus/findings surface as wvd.2 |
| 3 | agent-plugins-wvd.3 | isolated | sequential | standard | push | changes component-markdown-structure; disjoint from wave 3's corpus surface |
| 4 | agent-plugins-wvd.4 | isolated | sequential | advanced | push | changes component-gherkin; only queued co-changer is wvd.11, three waves later |
| 5 | agent-plugins-wvd.5 | shared | shared-batch #2 | advanced | defer | same c4_changes: component-corpus as wvd.2/.7/.8 |
| 5 | agent-plugins-wvd.6 | isolated | shared-batch #2 | standard | push | consumes component-corpus, changed by wvd.5 — ordered after it inside the bucket rather than costing a wave |
| 6 | agent-plugins-wvd.7 | shared | shared-batch #3 | standard | defer | same c4_changes: component-corpus, component-findings as wvd.8 |
| 6 | agent-plugins-wvd.8 | shared | shared-batch #3 | standard | push | same c4_changes: component-corpus, component-findings as wvd.7 |
| 7 | agent-plugins-wvd.11 | shared | shared-batch #4 | standard | defer | same c4_changes: component-command-surface as wvd.12 (and wvd.10, wave 8) |
| 7 | agent-plugins-wvd.12 | shared | shared-batch #4 | standard | push | same c4_changes: component-command-surface as wvd.11; consumes component-corpus, changed by wvd.7/.8 in wave 6 |
| 7 | agent-plugins-wvd.9 | isolated | sequential | standard | push | changes component-references; consumes component-corpus, changed by wvd.7/.8 in wave 6 |
| 8 | agent-plugins-wvd.10 | shared | sequential | standard | push | same c4_changes: component-command-surface as wvd.11/.12 — must not share their wave |
| 9 | agent-plugins-wvd.13 | isolated | sequential | advanced | push | c4_changes `[]`; consumes all seven components, so it runs last by construction |

### Blocked — in the committed queue, not dispatchable this run

| Bead | Class | Tier | Blocker | Reason |
|------|-------|------|---------|--------|
| agent-plugins-a7u | shared | advanced | `agent-plugins-3wd` (OPEN, **not in this run's queue**) | also blocked by wvd.7 + wvd.8, which wave 6 clears; 3wd is the one that does not |
| agent-plugins-oq8 | isolated | light | `agent-plugins-a7u` | blocked transitively through a7u |

**Decision for the user, not defaulted:** admitting `agent-plugins-3wd` would
unblock `a7u` (wave 9+) and then `oq8` (wave 10+). It is not in the committed
queue, so planning does not add it. Leave both blocked, or amend the queue.

### C4 sequencing constraints

Edges where one bead's `c4_impacts` intersects another's `c4_changes` and the
`blocks` DAG does not already order the pair. All are honoured by the wave
order above except the one marked, which is overridden below.

- `agent-plugins-wvd.4` after `agent-plugins-wvd.2` — consumes component-corpus (wave 4 after 3 ✓)
- `agent-plugins-wvd.9` after `agent-plugins-wvd.2` — consumes component-corpus (wave 7 after 3 ✓)
- `agent-plugins-wvd.12` after `agent-plugins-wvd.2` — consumes component-corpus (wave 7 after 3 ✓)
- `agent-plugins-wvd.6` after `agent-plugins-wvd.5` — consumes component-corpus (same bucket, ordered after ✓)
- `agent-plugins-wvd.9` after `agent-plugins-wvd.7` — consumes component-corpus (wave 7 after 6 ✓)
- `agent-plugins-wvd.9` after `agent-plugins-wvd.8` — consumes component-corpus (wave 7 after 6 ✓)
- `agent-plugins-wvd.12` after `agent-plugins-wvd.7` — consumes component-corpus (wave 7 after 6 ✓)
- `agent-plugins-wvd.12` after `agent-plugins-wvd.8` — consumes component-corpus (wave 7 after 6 ✓)
- `agent-plugins-wvd.8` after `agent-plugins-wvd.11` — consumes component-command-surface (**overridden**)
- `agent-plugins-wvd.8` after `agent-plugins-wvd.12` — consumes component-command-surface (**overridden**)

`wvd.8` and `wvd.12` consume each other's changed component, so the two edges
cannot both hold. Broken in favour of corpus-before-command-surface: the
corpus is the lower layer and the command surface reads it, so producers run
first. `wvd.8` is a check over corpus records that surfaces through the CLI;
`wvd.11`/`wvd.12` are CLI commands over corpus records. Running the CLI
commands first would build them against a corpus wvd.8 has not yet extended.

### C4-untagged beads

Grooming did not tag these; file paths are the whole collision signal for them,
and absent metadata reads as *unknown*, never "no collisions".

- `agent-plugins-4vz`, `agent-plugins-nu2`, `agent-plugins-0i6`,
  `agent-plugins-krb`, `agent-plugins-a7u`, `agent-plugins-oq8` — no
  `c4_changes` / `c4_impacts` key. Pre-groomed before this pass and not
  re-groomed. `nu2`, `0i6` and `krb` are placed by file-path overlap
  (union rule); `4vz`, `a7u`, `oq8` show no file overlap with any queued bead.
- `agent-plugins-r19`, `agent-plugins-wvd.13` — `c4_changes: []` **deliberately**,
  not missing. r19 touches `.claude/hooks/session-start.sh`, repository
  session-bootstrap infrastructure that no C4 component in `specs/architecture/`
  covers. wvd.13 edits no component; it folds deltas and impacts all seven.

### Gate bar for every bead in this run

Not exit 0 — **no findings beyond the recorded baseline**:

- `scripts/check-skills.sh --strict` exits 1 with 0 errors / 1 warning
  (bead-bonanza SKILL.md body size). A second warning or any error is a fail.
- `scripts/generate-schemas.sh --check` reports the committed artifacts stale.
  `agent-plugins-wvd.1` (wave 2) is the bead that clears it. **From wave 3 on,
  this gate is expected to pass** — a still-stale result after wave 2 is a
  regression, not the baseline.
- The three repo checks run INDEPENDENTLY, never chained with `&&`:
  `scripts/check-skills.sh --strict`; `scripts/sync-shared-references.sh --check`;
  `scripts/sync-strata-docs.sh --check`.
- `cd cli && go test ./...` additionally, whenever `cli/` changed.

### Overrides of the grooming wave preview

| # | Change | Why |
|---|--------|-----|
| 1 | `krb` + `0i6` merged into one fast-lane bucket (preview had them as four independent wave-1 beads) | both edit the snapshot-harness cluster; one bucket, one full test run instead of two, and it removes an unmetadata'd merge-collision risk |
| 2 | `nu2` batched with `wvd.2` (preview co-scheduled them concurrently in wave 3) | `nu2` is c4-untagged but edits the findings model `wvd.1` regenerates — union rule makes it a corpus/findings co-editor, so concurrent is unsafe |
| 3 | `wvd.6` moved from wave 4 into the wave-5 bucket after `wvd.5` (preview had wvd.4 + wvd.6 concurrent in wave 4) | honours the `wvd.6 consumes component-corpus, changed by wvd.5` sequencing edge the preview left unenforced, at no extra wave and one fewer test run |
| 4 | `wvd.7` + `wvd.8` batched, and `wvd.11` + `wvd.12` batched (preview split each pair across two waves) | both pairs are co-edit collisions, which a shared bucket resolves without spending a wave; collapses the preview's 9 waves' worth of L6 work into 2 |

### 2026-08-15 — amendment 1: admit agent-plugins-3wd

The planner flagged `a7u` as blocked by `3wd`, which is OPEN and was not in the
committed queue. Verified: the `a7u -> 3wd` edge is pre-existing (not wired by
this sprint's grooming), and `3wd` depends on `wvd.5` (DISC).

`3wd` is one of the four beads the blueprint's Tracker Reconciliation records as
**subsumed by DISC** — GHK extracts and DISC records the criteria, which is the
whole of what `3wd` asks for. Its remaining work is therefore verify-and-close,
not implementation.

ADMITTED under the standing authorization, by the Phase 1.5 step-4 necessity
test: `a7u` is in the run and cannot be finished until `3wd` closes. Queue is
now 21 beads. `3wd` slots after `wvd.5` (wave 5) and before `a7u`; `a7u` and
`oq8` become reachable in waves 8-9 rather than blocked.

### 2026-08-15 — amendment 2: c4 metadata gap on six pre-groomed beads

The planner reports `4vz`, `nu2`, `0i6`, `krb`, `a7u`, `oq8` carry no
`c4_changes`/`c4_impacts` key. Correct — this orchestrator set only
tier/tier_reason/est_size/sprint on those six and tagged C4 on the thirteen
`wvd.*` children alone. Not backfilled mid-run: the planner placed all six by
file-path signal instead, which is sound for beads this small, and re-grooming
them now would invalidate a wave table already written. Recorded as a grooming
finding for the sprint retro.

### 2026-08-15 — amendment 3: orchestrator merge error on wave 1 (recovered)

Merging `4vz` I rebased its branch onto `135ceab` — the sha `r19` merged at —
rather than onto current HEAD, which the run-directory snapshot commit had
already advanced to `0becc94`. The `--ff-only` correctly refused, and I then
deleted the branch before re-checking, leaving the rebased commit unreferenced.

Recovered: the commit was still in the object store, re-rebased onto the actual
HEAD and fast-forwarded. Sprint branch is linear, one commit per bead, no work
lost.

Cause: `bb-backup-run.sh` commits to the sprint branch, so HEAD moves between a
bead's acceptance and the next bead's merge. The rebase target in a concurrent
wave is therefore always `git rev-parse HEAD` read at merge time, never a sha
noted earlier in the wave. Remaining waves take it that way.

### 2026-08-15 — amendment 4: wvd.6 pulled forward from wave 5 into wave 4

The plan put `wvd.6` (PREG) in wave 5, batched behind `wvd.5` (DISC), to honour
the advisory C4 edge "wvd.6 consumes component-corpus, changed by wvd.5".

`wvd.6`'s only dependency edge is `wvd.2` (XPT), which closed in wave 3, so it
is dispatchable now. Its `c4_changes` is `component-profile-registration`;
`wvd.4`'s is `component-gherkin`. Disjoint, so the concurrency entry conditions
hold and the two run together.

The consumes-edge is advisory, not a dependency: PREG registers patterns into
the extension point XPT widened and validates them against the generated
`CriterionId`/`FeatureId` from MODEL. Neither is DISC's work. Saves one wave.

Wave table updated in place: wave 4 is now `wvd.4` + `wvd.6`; wave 5 is
`wvd.5` alone.
