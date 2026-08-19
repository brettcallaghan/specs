# bead-bonanza design review

A step-by-step map of the `bead-bonanza` skill as it stands at version 1.7,
and what the map shows. Research, not a decision: it records what was measured
and what the answers turned out to be. Nothing here is adopted until a bead
says so.

Published as an artifact for reading:
<https://claude.ai/code/artifact/14b4aece-c866-4b03-860a-7a539d64ed55>

**Evidence rule.** Every substantive claim is marked *measured* (a command was
run in the session that produced this document, and its output is in the
ledger at the end) or *unverified* (an inference or a design judgment, stated
with what would settle it). Claims taken from a prior run name the run.

**Conditions for every measured entry.** This container, 2026-08-19, branch
`claude/bead-bonanza-design-review-ne1plu`, `bd version 1.2.2 (dev)`, project
database at schema v65 — twelve migrations ahead of the binary — so every `bd`
call was made with `BD_IGNORE_SCHEMA_SKEW=1`.

**Scope.** The whole bundle (`SKILL.md`, seven references, five scripts), the
three sibling skills the workflow delegates to (`groom-beads`, `plan-sprint`,
`recover-sprint`), the five committed sprint reports, the 27 RFCs in
`rfcs/inbox/`, the beads 1.2.2 source documentation, and live probes against
this repository's own tracker.

---

## 1. The workflow, step by step

Every step a run executes, nested. The tag names who performs it — **[O]**
orchestrator, **[S]** subagent, **[X]** script, **[K]** a separate skill. That
distinction is the argument of section 2.

### Entry — before Phase 0

- E.1 **[O]** Read `CLAUDE.md` and `AGENTS.md` once, for repo conventions
- E.2 **[O]** Copy the six-line run checklist into the response (it lives in
  chat, not in a task tool)
- E.3 **[O]** If resuming after a compaction: re-read `SKILL.md` from Phase 0,
  invoke `recover-sprint`, re-run the queue query

### Phase 0 — Pre-flight

- 0.1 **[O]** Run `pre-flight-check.sh <run-id> --gate-baseline` with the
  sandbox ON and `eval` its stdout
  - 0.1.1 **[X]** Confirm cwd is inside a git work tree; resolve the repo root
  - 0.1.2 **[X]** `git worktree prune` — drop registrations whose directory a
    reclaimed container deleted
  - 0.1.3 **[X]** Confirm `## bead-bonanza gates` exists in `AGENTS.md`; echo
    it so the orchestrator never opens the file
  - 0.1.4 **[X]** Confirm `## model lineup` exists; echo it
  - 0.1.5 **[X]** Parse the gate profile: the ``- **Label:** `cmd` `` bullets
    before the first `###` subsection
  - 0.1.6 **[X]** Check Dolt telemetry suppression and the
    `eventsapi.dolthub.com` deny entry; report, never write user-global state
  - 0.1.7 **[X]** Create `$RUN_DIR` under
    `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>`; repoint `current`
  - 0.1.8 **[X]** Measure push mechanics — branch, remote, `.git/config`
    writability, upstream, default branch → `BB_PUSH_NEEDS_REFSPEC`
  - 0.1.9 **[X]** **Block** if HEAD is the default branch
  - 0.1.10 **[X]** Check `bd` on PATH and that `bd where` resolves a database;
    check `jq`
  - 0.1.11 **[X]** **Block** if the working tree is dirty at t=0
  - 0.1.12 **[X]** Report `CLAUDE_CODE_SUBAGENT_MODEL` — if set, every tier
    assignment is silently ignored
  - 0.1.13 **[X]** Count `.claude/settings.local.json` allow entries
  - 0.1.14 **[X]** **Block** if `worktree.baseRef` is not `head`
  - 0.1.15 **[X]** Detect `pull_request`-triggered workflows → `BB_PR_CI`
  - 0.1.16 **[X]** Under `--gate-baseline`: run every parsed gate on the
    untouched tree, record pass/fail per gate
  - 0.1.17 **[X]** Probe tracker durability with a real push to
    `refs/dolt/preflight-probe-$$`, delete the ref; **block** on failure
  - 0.1.18 **[X]** Emit the verdict and the evalable
    `RUN_DIR=` / `SCRIPTS=` / `BB_*=` block
- 0.2 **[O]** Read the stderr report once; fix any blocking condition
- 0.3 **[O]** Resolve the commit-trailer convention → `$RUN_DIR/trailers.txt`
- 0.4 **[O]** If `BB_PR_CI=yes`, ask the user once about a draft PR
- 0.5 **[O]** Confirm the harness worktree directory is gitignored — *after*
  the first worktree exists, not before
- 0.6 **[O]** Determine whether an `Agent` tool exists; if not, read
  `main-loop-fallback.md`

### Phase 1 — Resolve the queue

- 1.1 **[O]** Parse the user's filter (label, priority, id list)
- 1.2 **[O]** `bd ready` with the filter, for dependency-aware ordering
- 1.3 **[O]** Fall back to `bd list --status open --json` + jq if unsupported
- 1.4 **[O]** Admit only `open` beads
- 1.5 **[O]** Print the queue: count + ordered `id | P# | title`
- 1.6 **[O]** If no filter was given, ask once — the only Phase 1 question

### Phase 1.5 — Readiness gate (grooming)

- 1.5.1 **[O]** Cheap read-only check per bead, parallelized
  - 1.5.1a The nine Definition-of-Ready checks, read from
    `shared/work-order-definition-of-ready.md`, not from memory
  - 1.5.1b `bd lint` clean for the bead, where the installed `bd` has it
  - 1.5.1c A `decision-required` bead is bounced, not questioned mid-run
- 1.5.2 **[O]** Dispatch **one** grooming subagent over the ungroomed subset
  - 1.5.2a **[S]** Run `groom-beads` to the Definition of Ready
  - 1.5.2b **[S]** Assign `tier` + `tier_reason` + `est_size` into metadata
  - 1.5.2c **[S]** Tag `c4_changes` / `c4_impacts` as metadata arrays
  - 1.5.2d **[S]** Return one line per bead: `ready` or `NOT-READY — <gap>`
  - 1.5.2e **[S]** Return the wave preview
  - 1.5.2f **[S]** As its final act, invoke `plan-sprint` — Phase 2.5 runs
    inside this same subagent
- 1.5.3 **[O]** Re-run `bd ready`; grooming may have changed the queue
- 1.5.4 **[O]** Bounce what grooming could not clear, one line + gap each
- 1.5.5 Standing rule for the rest of the run — the mid-run scope test
  - 1.5.5a **[S]** A subagent *files* discovered work with a
    `discovered-from` edge; it never implements it
  - 1.5.5b **[O]** In scope only if a queued bead cannot be finished without
    it: unblocking prerequisite, AC decomposition, implied coverage gap
  - 1.5.5c **[O]** In scope → groom, re-`bd ready`, wire `blocks`, dispatch
  - 1.5.5d **[O]** Out of scope → leave filed and linked, surface in the report
  - 1.5.5e **[O]** Ballooning prerequisites is a stop signal

### Phase 2.5 — Triage and dispatch plan

- 2.5.1 **[O]** Write `dispatch.md`'s manifest header *before* dispatching
  grooming — run id, RUN_DIR, branch, base sha; gate profile and per-gate
  baseline; model lineup and effort (`unrecorded` where none is exposed);
  standing authorization; environmental ledger
- 2.5.2 **[O]** Pass `plan-sprint` the **enumerated** bead ids, never a filter
- 2.5.3 **[S]** Classify each bead `trivial` / `isolated` / `shared-surface`
- 2.5.4 **[S]** Compute overlap: C4 metadata primary, file paths fallback,
  union on disagreement
- 2.5.5 **[S]** Separate co-edit collisions from sequencing constraints
  (`c4_impacts` ∩ `c4_changes`)
- 2.5.6 **[S]** Choose a dispatch mode per bucket
- 2.5.7 **[S]** Append the wave table under the manifest, Reason column naming
  the collision class
- 2.5.8 **[S]** Return the table plus one line per override of the preview
- 2.5.9 **[O]** Print the wave table; do not ask for approval
- 2.5.10 **[O]** Resolve each bead's `tier` → model; honor a `model` override
- 2.5.11 **[O]** Ask once for standing authorization to auto-chain waves
- 2.5.12 **[O]** Amend `dispatch.md` in place on any divergence — append to
  `## Amendments` *and* update the table

### Phase 3 — Execute

- 3.0 **[O]** Before the first dispatch, arm the durable out-of-session trigger
  carrying RUN_DIR, the epic id, and "invoke `recover-sprint`"
- 3.1 **[O]** Per wave, re-verify `BB_WORKTREE_BASEREF=head`
- 3.2 **[O]** Buckets lowest-blast-radius first: fast lane → shared-surface
  batches → plain sequential
- 3.3 **[O]** Per bead, create the worktree from HEAD as a sibling of the repo
- 3.4 **[O]** Read `subagent-prompts.md` at dispatch, fill every
  `{PLACEHOLDER}`; never paraphrase from memory
- 3.5 **[O]** Arm a `Monitor` (`persistent`) on the progress JSONL for any
  subagent past five minutes, filtered to `committed|blocked`
- 3.6 The per-bead ceremony, inside the worktree
  - 3.6.1 **[S]** `… claimed`; `bd update --status in_progress`
  - 3.6.2 **[S]** `bd show <id>`
  - 3.6.3 **[S]** `… implementing`; implement. Defect beads: observe the
    reproducing test FAIL first
  - 3.6.4 **[S]** `… gates-running`; regen → format → lint → build → test →
    custom. Two failed cycles ⇒ Blocked path
  - 3.6.5 **[S]** `… gates-passed`; write `<id>.handoff.md` — a heading plus
    one `Caveat:` line — **before** the close
  - 3.6.6 **[S]** `… retro-written`; one `retro:` `bd comment` per friction
    point, 1–3 sentences, friction only
  - 3.6.7 **[X]** `bb-close-bead.sh` — atomic close, verify against the
    database, stage code only, one commit, `bd dolt push`
  - 3.6.8 **[S]** `… committed sha=$(git rev-parse --short HEAD)` — read from
    the branch, not the script's printout
  - 3.6.9 **[S]** If `push_per_bead`: push the branch, then `… pushed`
  - 3.6.10 **[S]** Return exactly one line: closed / blocked / skipped
- 3.7 **[O]** On return, append the summary line to the log
- 3.8 **[O]** Verify every `skipped` claim with `bb-verify-skip.sh`
- 3.9 **[O]** The five acceptance signals, per bead, in order
  - 3.9.1 The bead reads `CLOSED` in the tracker
  - 3.9.2 A commit subjected `<id> P<n> <title>` exists on the bead's branch
  - 3.9.3 The branch merged into the sprint branch (`--ff-only`;
    rebase-then-ff for the 2nd+ branch of a concurrent wave)
  - 3.9.4 The worktree is **gone** — the load-bearing signal, and the
    orchestrator's own final act
  - 3.9.5 At least one `retro:` comment, harvested via `bd export | jq`
- 3.10 **[X]** `bb-backup-run.sh "$RUN_DIR"` after *every* acceptance check,
  from the orchestrator's own checkout
- 3.11 **[O]** `git status --porcelain` clean, else fall through to recovery
- 3.12 **[O]** If `defer-push`: push, then `git log remote/branch..HEAD` empty
- 3.13 **[O]** Circuit breaker: three consecutive blocked/failed beads stop
  the run
- 3.14 **[O]** Under standing authorization, dispatch the next wave *in the
  same turn* as the wave report

### Phase 4 — Report, retro, teardown

- 4.1 **[O]** Results table `| ID | Status | Bucket | Commit | Notes |`
- 4.2 **[O]** Totals: N closed / N blocked / N skipped
- 4.3 **[O]** Implicit decisions, extracted from the `retro:` comments
- 4.4 **[O]** Tier escalations table, or the explicit `N/A` branch
- 4.5 **[O]** Deliberate over-routes; `none` stated explicitly
- 4.6 **[O]** Follow-up beads, admitted vs left for the user
- 4.7 **[O]** Stashes created during recovery
- 4.8 The sprint retro
  - 4.8.1 Harvest twice — filtered to closed, and unfiltered. A zero count is
    a broken query, not a clean run
  - 4.8.2 Q1 Grooming — who was under-groomed, which tier predictions were
    wrong and in which direction
  - 4.8.3 Q2 Specs — which documents an implementer found wrong or thin
  - 4.8.4 Q3 Plan — what `## Amendments` recorded, what would have bucketed
    better
  - 4.8.5 Q4 Skill — where `bead-bonanza` itself misled
  - 4.8.6 Write `rfcs/sprints/<date>-<epic-slug>.md`, every finding verbatim,
    numbered `<bead>#<n>`
  - 4.8.7 Offer routes: grooming → `groom-beads`, specs → `retro-spec`, skill
    → `retro-skill`
  - 4.8.8 Append a row to `rfcs/retro-log.md`; commit report + RFCs together
- 4.9 **[O]** Confirm every closed bead's commit is on the remote
- 4.10 **[O]** Cancel the durable trigger and state its disposition
- 4.11 **[O]** `git rm -r .bonanza/<run-id>/` as its own commit, after the
  report is committed
- 4.12 **[O]** Print `$RUN_DIR`. Do *not* open a PR or merge to main

### Exception paths — entered by condition, not by sequence

- X.1 **[O]** Subagent crash → `recovery.md`
- X.2 **[K]** Orchestrator death or compaction → `recover-sprint`
- X.3 **[O]** No `Agent` tool → `main-loop-fallback.md`
- X.4 **[O]** Missing `AGENTS.md` block → `project-setup.md`
- X.5 **[O]** Interrupted bead → `operating-principles.md`: the work stays
  uncommitted in its worktree; there is no WIP commit

Counting leaves, the happy path is about **95 discrete steps**: 18 inside one
script, 10 in the per-bead subagent ceremony, 5 in the acceptance check, and
roughly 60 of orchestrator judgment expressed as prose.

---

## 2. What the map shows

### 2.1 The procedure is 2.9× its own context budget

*Measured.* `SKILL.md` is 58,392 bytes. The repo's own gate
(`check-skills.sh`, `BODY_BYTE_MAX = 20000`) derives that cap from the
~5,000-token skill re-injection limit after a compaction. **34.3% of the file
survives.** `head -c 20000` lands at line 349, mid-sentence, inside the
`bb-close-bead.sh` interface description — before Phase 0, Phase 1, Phase 1.5,
Phase 2.5, Phase 4, the sprint retro, and every operating principle.

Epic `agent-plugins-92q` was written when `SKILL.md` was 45,983 bytes and named
the overflow as failure #1. Thirty-nine children later, all closed, the file is
27% larger. The epic's fix was to *reorder* so the executable core came first,
which is real mitigation — but reordering buys a fixed 20 KB of survivable
text, and every subsequent hardening bead spends more of it. The trend is the
finding, not the number.

Almost every defensive mechanism in the skill exists because of this cliff: the
"after any compaction, re-read from Phase 0" instruction, the whole
`recover-sprint` skill, the "a skill name survives truncation where a relative
path does not" rule, the published script interfaces, and the paragraph
explaining why the document is ordered out of numeric sequence. That is a lot
of machinery defending a constraint that decomposition removes outright. The
repo already knows how: `plan-sprint` and `recover-sprint` are exactly this
move, and both are under cap.

### 2.2 The run writes thirteen classes of artifact to five stores

| Artifact | Store | Per run | Lifetime | Read by |
|---|---|---|---|---|
| `dispatch.md` | `$RUN_DIR` (tmp) | 1 | run | recovery, retro |
| `trailers.txt` | `$RUN_DIR` | 1 | run | close script |
| `<id>.progress.jsonl` | `$RUN_DIR` | N | run | Monitor, recovery |
| `<id>.handoff.md` | `$RUN_DIR` | N | run | crash recovery only |
| `.bonanza/<run-id>/` | git branch | N commits | run, then deleted | recovery after reclaim |
| `retro:` comments | bd (Dolt) | ~3.5N | permanent | sprint retro, RFCs |
| bead status / notes | bd (Dolt) | N | permanent | everything |
| run checklist | chat | 1 | session | the user |
| per-wave reports | chat | W | session | the user |
| Phase 4 report | chat | 1 | session | the user |
| `rfcs/sprints/<date>.md` | git | 1 | permanent | triage-rfc |
| `rfcs/retro-log.md` row | git | 1 | permanent | triage-rfc |
| `rfcs/inbox/<skill>/*.md` | git | 0–4 | until triaged | triage-rfc |

For the 21-bead Gherkin run that is 44 files in `$RUN_DIR`, ~21 snapshot
commits, one deletion commit, 73 tracker comments, a 615-line report, three
RFCs and a log row — plus the 21 implementation commits that are the work.

**Verified defect — the snapshot's fallback does not survive a squash merge.**
*Measured.* `bb-backup-run.sh` and Phase 4 both assert that deleting
`.bonanza/<run-id>/` loses nothing, because
`git log --diff-filter=A -- '.bonanza/<run-id>/'` still finds the adding
commits. In this clone no commit reachable from any ref touches `.bonanza/` at
all, and the four most recent PR merges on `main` each have a single parent —
they were squash-merged, which discards exactly that history. The recovery
artifact is durable only for the life of the un-merged branch. Either say so,
or stop deleting it and let the squash discard it.

### 2.3 The process is not organized around the hierarchy it describes

Orchestrator → subagent → task is the right framing, and the skill
half-implements it. What breaks down is that **the middle tier has no artifacts
of its own.** A bead has a progress ledger, a handoff file, a branch, a
worktree, a commit and a comment thread. A *subagent* has nothing: no id, no
record, no lifecycle event, no way to ask "which agent is running, since when,
and what is it holding". Which is why:

- a batch subagent holding one worktree for six beads breaks acceptance signal
  4, and the skill needs three paragraphs to disambiguate a standing worktree;
- lost completion notifications require a whole sub-procedure, because the only
  evidence a subagent existed is the bead it touched;
- "a bead can be worked without ever being claimed" is stated as a hazard
  rather than fixed — the claim is per-bead, so a subagent that dies before
  claiming is invisible;
- tier escalation is recorded against the bead, so the run cannot separate "did
  this agent flail" from "was this bead mis-tiered".

### 2.4 Nearly every rule is a paragraph; very few are a program

Five scripts cover pre-flight, progress, skip verification, close and backup —
a real achievement, and where the reliability comes from. But the
orchestrator's own work, about 60 of the ~95 steps, is prose it must execute
correctly under compaction pressure. The five acceptance signals are the
clearest case: all five are pure git and `bd` queries, they run per bead every
time, they have a documented failure taxonomy, they must not be reordered, and
they are a numbered list the orchestrator performs by hand.

Run `1786778523-19` recorded two orchestrator conformance failures in one
sprint — a rebase onto a stale target, and stopping to narrate instead of
dispatching. Both are the kind of failure a script does not have.

### 2.5 The retro loop terminates in a queue, not in an action

*Measured.* Five sprint reports have produced 240 harvested findings.
`rfcs/inbox/` holds 27 RFCs — eight against `bead-bonanza` and its two helper
skills — of which 25 are `ready-for-review` and 2 are drafts. The pipeline
runs: findings → report → RFC → triage → beads → a future sprint. Each hop is a
human decision point, and the queue is where the latency lives.

---

## 3. The unused half of beads

`bd` is not just a tracker the skill under-uses. `bd` ships an orchestration
control plane, and `bead-bonanza` reimplements about eight of its primitives in
prose.

### 3.1 `bd swarm validate` is `plan-sprint`

*Measured.* Run against this repo's own closed epic `agent-plugins-9kb`, with
no argument beyond the epic id, `bd swarm validate` produced the six-wave
partition, max parallelism 3, and an estimate of 10 worker-sessions — computed
from the dependency graph.

What it does *not* compute is C4 co-edit overlap, which is the part
`plan-sprint` adds and `bd` cannot know. So the right decomposition is:
**`bd swarm validate` produces the dependency waves; `plan-sprint` subdivides
them by surface collision and picks a dispatch mode.** Today `plan-sprint` does
both, and the first half is a re-derivation.

`bd swarm validate` also reports what `plan-sprint` has no check for: wrong
dependency direction, orphaned issues, cycles, disconnected subgraphs. Those
are grooming defects, and the 2026-08-15 retro found one by hand.

### 3.2 Inventory

| bd primitive | What it does | What bead-bonanza does instead | Status |
|---|---|---|---|
| `bd swarm create/status/validate` | Wraps an epic as coordinated parallel work; ready fronts, max parallelism, worker-session estimate; `status --json` is a live view | `plan-sprint` derives waves by hand; no live status view | unused |
| `bd ready --claim` | Atomically claims the first ready issue | `bd update --status in_progress`, non-atomic; a "never reset a claim you did not set" rule covers the race | unused |
| `bd ready --parent <epic>` | Scopes ready work to an epic's descendants | "Pass enumerated ids, never the filter" — a rule repeated in four documents to solve exactly this | unused |
| `bd merge-slot acquire/release` | Exclusive lock serializing merge/conflict resolution, with a priority waiter queue | "The orchestrator merges one at a time", held only in the orchestrator's head | unused |
| `bd gate create --await gh:run:<id>` | Blocks until a GitHub Actions run finishes, with a timeout, no polling | Draft-PR CI is asked about once, then never observed | unused |
| `bd gate --await human:` · `bd human` | First-class human-decision channel: list, respond, dismiss, timeout | The `decision-required` label plus a prose bounce procedure | unused |
| `--type event` beads · `--wisp-type heartbeat` | Durable, synced, TTL-compacted event records with actor/target/category/payload | `<id>.progress.jsonl` in `/tmp` plus `bb-backup-run.sh` keeping it alive | unused |
| `bd set-state <id> health=…` | Atomic operational-state dimension: event bead plus lookup label | No health signal exists at all | unused |
| `bd worktree create/remove` | Worktree management with beads config, gitignore, and removal safety checks | Raw `git worktree add`, plus a documented trap about harness placement | unevaluated |
| `bd formula` / `bd mol pour` | A TOML DAG of steps instantiated as beads with per-step checkpoint recovery | The six-line run checklist in chat, plus `recover-sprint` | unevaluated |
| `bd batch` | Many writes in one transaction, one Dolt commit, rolled back on error | N separate `bd` invocations | minor |
| `bd update --session <id>` | Records the Claude Code session that closed a bead | Open bead `agent-plugins-oyf` exists to add exactly this | unused |
| `bd query "…"` | Compound query language with booleans and relative dates | `bd list --json \| jq` pipelines | minor |
| `bd dep cycles` · `bd graph --html` | Cycle detection; interactive D3 dependency view | A markdown wave table | minor |
| `bd find-duplicates` | Mechanical or AI duplicate detection | The Skipped path asks the subagent to `git log \| grep` | minor |

### 3.3 Two places the skill is arguably mis-using beads

**Tier is stored twice, by two skills.** *Measured.* `groom-beads` writes tier
into metadata (`--set-metadata tier=standard`); `bead-bonanza` reads a
`tier:<name>` label and offers `bd list --label tier:standard --json` as a bulk
query. Across this database 153 beads carry `metadata.tier` and 63 carry a
`tier:` label; the sets do not coincide. The documented bulk query **returns 0**
— `bd list` defaults to open issues, so it silently misses everything groomed
and closed. That is the same trap the findings harvest hit and fixed; it
survives here unfixed. No currently-open bead carries a tier by either
mechanism (`{with_label: 0, with_meta: 0, total: 51}`).

Beads has a documented convention for this data — `execution_agent_type`,
`execution_suggested_model`, `execution_reasoning_effort`,
`execution_parallel_group`, in `docs/METADATA.md` — with the instruction that
"parent/orchestrator agents must consume these keys before spawning subagents".
Adopting those names makes the beads portable to any orchestrator, and
`execution_reasoning_effort` fills the `effort: unrecorded` hole in the
dispatch manifest.

**A stale `bd show --json` workaround.** *Measured.* The skill says
`bd show <id> --json` "reports `comment_count: 0` for a bead that has
comments". On `bd` 1.2.2 it reports `comment_count: 5` correctly. The
conclusion still holds — `--json` carries no `comments` array, so
`bd export | jq` remains the right harvest — but the stated reason is now
wrong, and a reader who checks it will conclude the rule is stale and drop it.

**One precondition to check before scaling concurrency.** *Unverified.* Beads'
documented default storage mode is embedded: single-writer, file-locked;
multi-writer needs an external `dolt sql-server`. Concurrent worktree subagents
each run `bd update`, `bd comment` and `bd dolt push` against one database.
Nothing in five runs has reported a lock failure, so this may be fine at 2–4
agents. It would be settled by dispatching a concurrent wave with
`BD_OTEL_STDOUT=true` and reading `bd_db_lock_wait_ms`, which `bd` already
emits.

---

## 4. Prior art

### Gas Town — the same design, built as software

Gas Town is the multi-agent workspace manager built by the beads authors, and
it is `bead-bonanza`'s nearest neighbour: an AI coordinator dispatching a
git-backed work ledger to 20–30 agents, with watchdog recovery and a bisecting
merge queue that lands green batches. Its vocabulary maps almost one-to-one — a
*rig* is a repo, a *polecat* is a worker agent, a *witness* is a per-rig
lifecycle manager that monitors polecats, detects stuck agents and triggers
recovery, and a *swarm* is the workers currently on a convoy's issues.

Three ideas transfer without adopting Gas Town itself:

- **Beads as the control plane, with no separate orchestrator state.** Every
  coordination fact lives in the tracker, so any session can reconstruct the
  run by querying it. `bead-bonanza` keeps its most load-bearing artifact
  (`dispatch.md`) in `/tmp` and then builds a script to make it durable — a
  script that exists because the state is in the wrong store.
- **A watchdog as a separate role.** The witness is not the coordinator. The
  Monitor and durable trigger are the same idea, but both are owned by the
  orchestrator, so an interrupt that kills the orchestrator kills the thing
  that would have reported it.
- **The merge queue as a first-class component**, not a step in the
  coordinator's procedure. `bd merge-slot` is its primitive, already installed.

### The commodity orchestrator tier

The 2026 open-source orchestrators (Agent Orchestrator, Orca, Emdash) converge
on isolated worktrees per agent, a dashboard, and human-retained merge control.
The one capability they emphasize that `bead-bonanza` lacks is **routing
failures back to the responsible agent** — CI failures, review comments and
merge conflicts re-dispatched to whoever produced them. `bead-bonanza`'s
subagents are fire-and-forget, which is why the shared-surface batch template
has to describe bisecting by hand.

They also confirm what the design gets right: bucketing by predicted collision
*before* parallel execution. Doing it from architecture metadata rather than
file paths is ahead of that tier, not behind it.

### Anthropic's long-running-agent harness

The pattern is a two-agent split: an **initializer agent** that sets up the
environment once — `init.sh`, a progress log, a `feature_list.json` of
end-to-end cases all marked `"passes": false`, and an initial git commit — and
a **coding agent** that makes incremental progress each session and leaves
clear artifacts for the next. The reported failure modes are the ones
`bead-bonanza` was built against: one-shotting the whole app, half-implemented
features, declaring victory early.

Two rules from it that `bead-bonanza` does not hold:

- *"Work on one feature at a time. Mark a feature complete only after
  end-to-end verification confirms it works, not when the code is written."*
  The acceptance check is stronger than this at the bead level. The **run** has
  no equivalent: nothing verifies the epic's stated outcome was achieved, only
  that N beads closed.
- *"Evaluators are tuned to be strict and objective, overcoming the model's
  natural bias to self-approve."* Run `1786778523-19` caught this exactly — a
  handoff reported a gate clean that exits 1 with 222 findings. The skill's own
  conclusion, "a gate claim in a handoff is testimony, not evidence", is now a
  rule. It is not yet a check.

---

## 5. The memory-tool frame

### 5.1 The order is initializer-first, and setup differs from resume

The memory-tool documentation puts the initializer first and has it produce the
other three: *"The first session sets up the memory files before any
substantive work begins. This includes a progress log, a feature checklist, and
a reference to any startup or initialization script."* Then *"each new session
opens by reading those memory files"*, and *"before a session ends, it updates
the progress log."*

So there are two orders, and they differ:

| | Setup order (build) | Resume order (read) |
|---|---|---|
| 1 | Initializer — measures the environment, creates everything below | Checklist — what is the scope, what is done |
| 2 | Checklist — the scope, every item false | Progress log — where the last session stopped |
| 3 | Progress log — empty, ready to append | Git / durable state — what actually landed |
| 4 | Git commit — the baseline recovery point | Initializer — re-run to restore the environment, not the state |

`bead-bonanza` conflates them. `pre-flight-check.sh` is the initializer and
runs first, correctly. But the checklist is created in chat at entry, *before*
pre-flight, and the plan (`dispatch.md`) is created at Phase 2.5, two phases
after the measuring is done. On resume, `recover-sprint` reads git before the
plan — which is the right order, for a reason the memory-tool docs do not have
to state: the log is written by the agent doing the work and stops when it
stops.

### 5.2 The frame makes the three levels consistent

Each level gets the same four artifacts. Today only the task level has all four.

| Level | Checklist | Initializer | Progress log | Recovery |
|---|---|---|---|---|
| **Orchestrator** (the sprint) | run checklist — *chat only* | `pre-flight-check.sh` — have | `dispatch.md` `## Amendments` — *tmp* | `recover-sprint` — have |
| **Subagent** (the developer) | **missing** | worktree creation — *implicit* | **missing** | `recovery.md` — *bead-scoped* |
| **Task** (the bead) | the ten-step ceremony — have | the dispatch prompt — have | `<id>.progress.jsonl` — have | `<id>.handoff.md` + worktree — have |

Read across: the task level is fully instrumented and the subagent level is
empty, which is 2.3 arriving from another direction. The orchestrator level has
all four, but two live in stores that do not survive the thing they exist to
survive — chat and `/tmp`.

### 5.3 Mechanical versus agentic checklist items

Almost every item is mechanical. **The distinguishing test: does the item
assert a fact about durable state, or a judgment about content?** Facts about
state are mechanical and belong in a script any session can re-run. Judgments
about content are agentic and must be recorded as claims with evidence
attached, because nothing can re-derive them.

| Checklist item | Verifier | Kind |
|---|---|---|
| 0. Pre-flight run, RUN_DIR + SCRIPTS eval'd | `[ -d "$RUN_DIR" ]` and the script's exit code | mechanical |
| 1. Queue resolved and printed | `dispatch.md` carries an enumerated wave table | mechanical |
| 1.5 Every bead groomed or bounced | `bd lint <ids>` clean · no `decision-required` label · `metadata.tier` non-null on every queued bead | mechanical |
| 1.5 …the *quality* of a groomed description | the fresh-agent test | agentic |
| 2. Bucket + tier + push mode per bead | every enumerated id in exactly one row with all three columns filled | mechanical |
| 2. …whether the bucketing is *right* | C4 overlap is computed; whether the surfaces really collide is judgment | agentic |
| 3. Acceptance signals 1–5, per bead | `bd show` · `git log --grep` · `git branch --contains` · `git worktree list` · `bd export \| jq` | mechanical |
| 3. Skip claims verified | `bb-verify-skip.sh` exit code | mechanical |
| 3. Wave landed on the remote | `git log remote/branch..HEAD` empty | mechanical |
| 3. Whether emergent work is in scope | the necessity test | agentic |
| 4. Report sections present, nulls stated | grep for each required heading and for `none`/`N/A` | mechanical |
| 4. Findings harvested, count non-zero | the `bd export \| jq` pipeline | mechanical |
| 4. What the findings *mean* | the four retro questions | agentic |
| 4. Durable trigger cancelled | the trigger API's own list | mechanical |
| 4. Snapshot removed | `[ ! -d .bonanza/<run-id> ]` | mechanical |

That test yields a concrete artifact: **a `bb-status.sh` that takes a RUN_DIR
or epic id and prints the checklist with every mechanical item independently
verified.** It is the answer to the restart question — a recovering session
does not need to be told where it is; it runs one command and is told. It also
turns `recover-sprint`'s six manual steps into "run this, then act".

---

## 6. Consolidating the artifacts

### 6.1 Handoff versus report

They are not duplicates, and the **handoff** is the weak one.

*Measured, run `1786345424-bbh`.* Across 31 beads the handoff file was written
23 times and needed **once** — on the one bead that crashed, where it was
**absent**, because the agent died between emitting `retro-written` and writing
the file. The mitigation already landed (`bb-close-bead.sh --run-dir` writes a
minimal handoff if none exists), which means the file is now written by the
close ceremony anyway. Its remaining payload is one `Caveat:` line that is
`none` in the overwhelming majority of cases.

**Proposal:** delete `<id>.handoff.md` and make the caveat a field on the
`gates-passed` progress event —
`bb-progress.sh … gates-passed caveat="regen before staging"`. One artifact
class, one write, no ordering hazard between a phase and a file, and the caveat
becomes queryable rather than parseable. This also retires the "write the
handoff BEFORE the close" rule and the crash-resilient-close section of
`recovery.md`.

The sprint **report** should stay: it is the run's testimony, cited by RFCs,
and the substrate `triage-rfc`'s recurrence detection runs on. But the Phase 4
chat report and the sprint report are largely the same content written twice.
The chat version should summarise and link, not re-author.

### 6.2 One log, or one store?

**Not one log — one store.** Per-bead files are correct; a shared append-only
log across concurrent worktree agents is a contention and interleaving hazard
for no gain. The problem is not the split, it is that the store is `/tmp` and
therefore needs `bb-backup-run.sh` committing ~N snapshot commits per run to
stay alive.

Beads already has the right store. An event bead
(`bd create --type event --event-category bb.gates-passed --event-target <id>
--event-payload '{"caveat":"…"}'`) is durable, synced on the same push that
makes the closure durable, queryable, and visible to any session without a run
directory. That single move retires `bb-backup-run.sh`, the per-bead snapshot
commits, the deletion commit, the "live directory is authoritative" rule in
three documents, and the squash-merge defect in 2.2.

*Unverified:* the write cost. N×8 event beads per run against a Dolt database,
plus their sync. Settled by writing 100 event beads and timing `bd dolt push`
before and after. The fallback is wisps (`--ephemeral --wisp-type heartbeat`),
designed for high-volume operational records and TTL-compacted — but wisps are
*not synced*, which puts them back in `/tmp`'s durability class.

### 6.3 Structured formats

Two files are prose that wants to be data:

- **`dispatch.md`'s wave table** — written by `plan-sprint`, amended by the
  orchestrator, reconstructed by `recover-sprint`. Three consumers parsing
  markdown. A `dispatch.json` with a rendered view would let `bb-status.sh`
  verify "every enumerated id appears in exactly one row with tier, bucket and
  push mode" mechanically. Same argument for the manifest header.
- **The `AGENTS.md` gate profile** — `pre-flight-check.sh` parses it with awk
  plus a documented rule about which bullets are commands and which are
  commentary. Open bead `agent-plugins-pq06` already proposes replacing it with
  project-supplied checkpoint scripts; that bead is the right answer.

### 6.4 The proposed artifact set

| Today | Proposed | Why |
|---|---|---|
| `dispatch.md` in /tmp | `dispatch.json` + rendered view, attached to the sprint epic in `bd` | durable without a backup script; mechanically checkable; found by epic id, not an unguessable run id |
| `<id>.progress.jsonl` | event beads on the bead | durable, synced, queryable, survives reclaim |
| `<id>.handoff.md` | *deleted* — caveat becomes a field on `gates-passed` | written 23×, needed 1×, absent when needed |
| `.bonanza/<run-id>/` + N commits + deletion commit | *deleted* | only exists to make /tmp durable; its own fallback does not survive a squash merge |
| `trailers.txt` | keep | per-session, correctly not committed |
| run checklist in chat | `bb-status.sh` output, echoed to chat | every mechanical item verified rather than asserted |
| sprint report + retro-log row | keep | the run's testimony; the recurrence substrate |
| RFCs in `rfcs/inbox/` | keep, but see §7 | 27 queued is the symptom |

Net: **from 13 artifact classes across 5 stores to 6 classes across 3**, and
the two stores that disappear are the two that do not survive the failures the
design exists to survive.

---

## 7. The retro, reframed

### 7.1 Triage findings in the same turn

Yes, with one boundary. The orchestrator is holding the run, the evidence is
free, and a finding routed into an RFC queue waits for a human decision the
same human could make in the same turn. It also completes the sprint metaphor:
a retro that ends in a backlog item is a retro; one that ends in a document
nobody has read is a status report.

The boundary: **the sprint report must still be written in full, every finding
verbatim.** That is the substrate `triage-rfc`'s recurrence rule runs on, and
the skill states the reason precisely — "a finding recorded once and filtered
out the next time can never be observed appearing twice". Triage decides what
becomes a bead now; the report preserves what did not, so it can be seen
recurring.

The shape that follows:

1. Harvest and write the sprint report as today — every finding, verbatim,
   numbered.
2. Answer the four questions as today.
3. **New:** classify every finding into one of four buckets, mechanically where
   possible — *defect in this repo*, *defect in a skill*, *spec gap*,
   *observation only*.
4. **New:** present the non-observation-only findings as one numbered list with
   a proposed title and priority each, and ask which to file. One question, one
   turn, at the point of maximum context.
5. File the accepted ones with `bd create`, each carrying a `discovered-from`
   edge to the bead that produced the finding and a citation to the report
   section.
6. Keep the RFC route as an *escalation*: a finding proposing a change to a
   skill's *design*, rather than a fix to its text, still deserves the RFC's
   evidence-and-proposals structure and `triage-rfc`'s deferral machinery. Most
   findings are not that.

### 7.2 Is 2–3 sentences enough?

Right for the *capture*, wrong for the *triage* — and the Findings Capture
Protocol already says why: "the capturing agent is a sensor, and analysis is
the aggregated retro's job". Lengthening the finding pushes analysis into the
agent least able to do it, one that has seen exactly one bead.

But 2–3 *free-form* sentences is the problem. *Measured:* across 240 harvested
findings the protocol's own six trigger categories are never recorded, so the
retro re-derives every finding's kind by reading it. A minimal structured
prefix costs the sensor nothing and makes step 3 above mostly mechanical:

```
retro:<kind>[:<target>] <1-3 sentences: what happened + the text that caused it>

kind ∈ { gap           — information the bead and its references did not carry
         decision      — something the bead or spec should have pinned
         contradiction — the bead or spec contradicted repo reality
         unprovable    — an AC or citation obligation that could not be satisfied
         sequencing    — inherited breakage, or a dependency the graph did not encode
         misscoped     — too large, trivially small, or overlapping another item
         tooling       — the skill, a script, or the harness misled or blocked }

target = the document slug, skill name, or script the finding is about,
         where the finding is about one. Omitted otherwise.
```

Seven kinds, six lifted verbatim from the protocol's existing "when to record"
list plus `tooling` for the skill-and-script class it does not name. The
`target` is what makes routing mechanical: `retro:tooling:bead-bonanza` routes
itself. Everything after the prefix stays as it is.

*Unverified:* whether the seven kinds partition the real population cleanly.
Settled by classifying the 73 findings in the 2026-08-15 report against them
and counting the residue.

---

## 8. Recommendations

Ranked by expected reduction in failure rate per unit of work.

### Tier 1 — changes what the skill can survive

**R1. Split `SKILL.md` into invoked skills until the body is under 20 KB.**
Move Phase 0/1/1.5 into a `start-sprint` skill and Phase 4 into a
`close-sprint` skill, matching the existing `plan-sprint` / `recover-sprint`
pattern. What stays is the Phase 3 loop, the acceptance check, the worktree
lifecycle and the operating principles — the parts a compacted orchestrator
genuinely needs re-injected. This is the enabling change for most of the rest:
it removes the compaction cliff, retires the "re-read from Phase 0" instruction
and the out-of-numeric-order apologia, and turns the `--strict` known-red
baseline in `AGENTS.md` green. *Evidence:* 34.3% survival, measured. *Effort:*
high.

**R2. Write `bb-status.sh`: the run checklist, mechanically verified.** One
script, one argument, printing every checklist item with a verified state and
the command that established it. Every item marked mechanical in §5.3 is
already a one-line shell or `bd` query; the five acceptance signals are the
core. Run it after every bead instead of performing the signals by hand, at
every wave boundary, and as `recover-sprint`'s first step. *Evidence:* two
orchestrator conformance failures in one run. *Effort:* medium.

**R3. Make the sprint an epic, and the epic the queue.** Create (or require) an
epic per run and make the queue its children. Then `bd ready --parent <epic>`
*is* the enumerated queue — membership is durable in the tracker, so a filter
re-evaluated at recovery cannot admit beads nobody scoped. That retires the
"enumerate, never filter" rule from all four documents carrying it. It also
unlocks `bd swarm validate` for the dependency-wave half of Phase 2.5,
`bd swarm status --json` as a live run view, and `bd epic status` as the
run-level completion signal §4 says is missing. Attach `dispatch.json` to the
epic and the run directory stops being load-bearing. *Evidence:* `swarm
validate` reproduced `9kb`'s six waves, measured. *Effort:* medium.

**R4. Give the subagent a first-class record.** At dispatch, create a durable
record — an event bead, a wisp, or a row in `dispatch.json` — carrying dispatch
id, model, tier, worktree path, assigned beads, dispatch time and terminal
state. Close it on return. This is the missing middle of the hierarchy: it
makes "which agent is running and what is it holding" a query rather than an
inference, disambiguates a batch's shared worktree from a failed removal
without the three-case prose rule, makes a lost notification a comparison
between dispatched and returned, and separates "this agent flailed" from "this
bead was mis-tiered" in the routing evidence. *Evidence:* notification loss in
three waves of one run. *Effort:* medium.

### Tier 2 — removes ceremony

**R5. Move the progress ledger into the tracker; delete the backup script.**
Replace `<id>.progress.jsonl` with event beads, and delete `bb-backup-run.sh`,
its ~N snapshot commits, its deletion commit, and the "live directory is
authoritative" rule in three documents. Fold the handoff caveat into the
`gates-passed` event and delete `<id>.handoff.md`. Measure the write cost
first — this is the one recommendation with an unverified precondition.

**R6. Triage the retro into beads in the same turn.** §7.1's six-step shape.
Keep the report complete; make the RFC an escalation rather than the default;
ask once, at the point of maximum context.

**R7. Structure the retro finding prefix.** `retro:<kind>[:<target>]`, seven
kinds, six already in the Findings Capture Protocol. Keep 1–3 sentences and
keep analysis out. Validate the taxonomy against the 73 findings on disk first.

**R8. Fix the tier double-store, and adopt beads' metadata convention.** Pick
metadata as the single home, delete the `tier:` label path from
`bead-bonanza`, and delete the `bd list --label tier:standard --json` example —
it returns 0. Then rename toward `execution_suggested_model` /
`execution_reasoning_effort` / `execution_parallel_group` so the beads are
portable and the `effort: unrecorded` hole closes.

**R9. Adopt `bd ready --claim` and `bd merge-slot`.** Atomic claiming removes
the race the "never reset a claim you did not set" rule papers over. The merge
slot makes serialized merging a lock the whole run can see rather than a
discipline the orchestrator maintains — which is exactly the invariant that
broke in the 2026-08-15 rebase incident.

**R10. Correct three stale or unsurvivable claims.** (a) The `.bonanza`
snapshot's "nothing is lost" fallback does not survive a squash merge — state
it or stop deleting. (b) `bd show --json` now reports `comment_count`
correctly; the rule to use `bd export` is right but its stated reason is wrong.
(c) Add a ledger entry for the v65/v53 schema skew and
`BD_IGNORE_SCHEMA_SKEW=1`, which every `bd` call in this container needs.

### Tier 3 — worth doing, nothing depends on it

**R11. Replace gate-profile prose parsing with project checkpoint scripts.**
Open bead `agent-plugins-pq06` already proposes this. Parsing executable
commands out of markdown bullets is the most fragile step in the pre-flight.

**R12. Add a run-level Definition of Done.** The harness pattern's rule — mark
complete only after end-to-end verification — holds at the bead level and
nowhere else. A run ends when the queue drains, not when the epic's outcome is
demonstrated. `bd epic status` gives the mechanical half; the agentic half is
one paragraph asserting the epic's success criteria are met, with the command
that shows it.

### What not to change

Stated explicitly so a later reader does not mistake omission for oversight.

- **The five acceptance signals.** State, not testimony, ordered cheapest
  first. The worktree-is-the-in-flight-signal invariant is the strongest idea
  in the design and is why lost notifications cost nothing. Mechanize them
  (R2); do not weaken them.
- **C4 overlap as the primary bucketing signal.** Ahead of the commodity
  orchestrators, which bucket on file paths. Separating co-edit collisions from
  sequencing constraints is a real distinction `bd` cannot compute — and it is
  `plan-sprint`'s reason to exist once R3 takes the dependency waves.
- **One commit per bead, in its own worktree.** Landed 23 consecutive beads
  across a container restart with zero loss. "An interrupted bead's work stays
  uncommitted in its worktree" removes an entire class of WIP-commit ceremony.
- **Per-bead files, not one shared log.** Concurrent agents appending to one
  file is a contention hazard for no gain. The store is the problem, not the
  split.

---

## 9. Evidence ledger

| Claim | Command | Observed |
|---|---|---|
| `SKILL.md` is 58,392 bytes; 34.3% survives the re-injection cap | `wc -c plugin/skills/bead-bonanza/SKILL.md` · `head -c 20000 … \| wc -l` | `58392`; truncation lands at line 349, mid-sentence in the `bb-close-bead.sh` interface |
| The cap is 20,000 bytes and is the repo's own gate | `grep -n BODY_BYTE_MAX scripts/check-skills.sh` | `BODY_BYTE_MAX = 20000`, derived from ~5,000 tokens at ~4 bytes/token |
| bead-bonanza is the only skill over cap | `for f in plugin/skills/*/SKILL.md; do wc -c $f; done \| sort -rn` | 58,392; next largest 20,648 (`groom-beads`); six others below |
| `bd swarm validate` computes waves, parallelism and worker-sessions | `bd swarm validate agent-plugins-9kb` | six waves reproduced from the dependency graph; max parallelism 3; ten worker-sessions |
| Tier is stored in two places; the documented bulk query returns 0 | `bd list --label tier:standard --json \| jq length` · `bd export \| jq 'select(.metadata.tier!=null)'` | `0` (default status filter is open); 153 beads with `metadata.tier`, 63 with a `tier:` label; `--status all` returns 37 |
| No open bead carries a tier by either mechanism | `bd export \| jq -s` over `status=="open"` | `{with_label: 0, with_meta: 0, total: 51}` |
| `bd show --json` reports comment_count correctly but carries no comments array | `bd show agent-plugins-9v1 --json \| jq '.[0]\|{comment_count, has_comments:(.comments\|length)}'` | `{comment_count: 5, has_comments: 0}` |
| No `.bonanza/` path survives on any ref; recent PRs are squash merges | `git log --all --oneline -- .bonanza \| wc -l` · `git rev-list --parents -n1 <sha>` | `0`; four recent PR merge commits each have exactly one parent |
| 27 RFCs queued; 8 against bead-bonanza and its helpers | `find rfcs/inbox -type f -name '*.md' \| wc -l` · `grep -h '^status:' rfcs/inbox/*/*.md \| sort \| uniq -c` | 27 files; 25 `ready-for-review`, 2 `draft`; 6 `bead-bonanza`, 1 `plan-sprint`, 1 `recover-sprint` |
| 240 findings across five sprint reports | `rfcs/retro-log.md` harvest counts | 73 + 14 + 16 + 64 + 73 |
| The database is twelve migrations ahead of the binary | `bd stats` | `database is at v65, binary knows up to v53`; recovery guide printed |
| Handoff written 23×, needed 1×, absent then; retro file cost 14× the ledger | prior run `1786345424-bbh`, file-level audit of 67 files / 328 KB | recorded in `rfcs/inbox/bead-bonanza/rfc-self-run-directory-as-evidence.md`; ledger tail accurate 29/31 |
| Two orchestrator conformance failures in one run | prior run `1786778523-19`, amendments 3 and 6 | stale rebase target requiring object-store recovery; stopped to narrate a wave instead of dispatching it |

### Unverified claims, and what would settle each

- **Event-bead write cost at run volume** — write 100 event beads, time
  `bd dolt push` before and after. Gates R5.
- **Embedded-Dolt lock contention under concurrent subagents** — dispatch a
  concurrent wave with `BD_OTEL_STDOUT=true` and read `bd_db_lock_wait_ms`.
  Gates how far concurrency can scale.
- **Whether the seven finding kinds partition the real population** — classify
  the 73 findings in the 2026-08-15 report and count the residue. Gates R7.
- **Whether `bd worktree create/remove` is better than raw git here** — the
  beads skill says always use it; `bead-bonanza` measured that `bd` already
  resolves the shared database through git's common directory, so the redirect
  files it configures may be moot. Try it once and compare.
- **Whether `bd mol pour`'s per-step checkpoint recovery is a better run
  checklist than `bb-status.sh`** — pour a three-step formula, kill the
  session, and see what resumes.

### Sources for the external prior art

- Anthropic memory-tool documentation, "Multisession software development
  pattern" (platform.claude.com).
- Beads 1.2.2 source documentation, read from the local module cache:
  `docs/MOLECULES.md`, `docs/METADATA.md`, `docs/OBSERVABILITY.md`,
  `docs/EXCLUSIVE_LOCK.md`, and the bundled beads skill's
  `resources/RESUMABILITY.md`, `resources/WORKTREES.md`,
  `resources/ASYNC_GATES.md`.
- Gas Town and beads project documentation, and published surveys of 2026
  open-source agent orchestrators, via web search.
