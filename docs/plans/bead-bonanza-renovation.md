# Plan: renovate bead-bonanza

Carries out the design review in
[`docs/research/bead-bonanza-design/README.md`](../research/bead-bonanza-design/README.md).
Transient — delete when the change lands.

Every item below is shaped as a bead-in-waiting: an id, what it does, what it
retires, what gates it, and a proposed title. Nothing here is a bead yet.

---

## How this plan is ordered, and why

Your sequencing rule, adopted as the spine:

> Eliminate steps we've decided to replicate. Renovate steps we've decided to
> improve. Mechanize workflows that can be handled by scripts instead of
> agents. THEN split things up into separate skills.

Two amendments the review forces:

- **Spikes come before eliminate.** Seven claims are unverified, and four of
  them gate an elimination. Deleting the progress ledger before knowing whether
  a polled Monitor works would trade a measured problem for an unmeasured one.
  `bd` has a built-in `spike` type — *timeboxed investigation to reduce
  uncertainty before committing to a story* — which is exactly this wave.
- **The epic pivot sits between renovate and mechanize**, where you put it.
  It is not a renovation of a step; it changes what a run *is*, and almost
  every mechanization below needs a stable run handle to key off. Doing it
  after wave 2 means the eliminations have already shrunk what has to move.

```
Wave 0  Spikes ─────────────┐
Wave 1  Eliminate ──────────┤
Wave 2  Renovate ───────────┤
Wave 3  Epic pivot ─────────┤──► Wave 4  Mechanize ──► Wave 6  Split
Wave 5  Retro rebuild ──────┘
```

Wave 5 is independent of 3 and 4 and can run alongside them; it is drawn last
only because its first item is a design decision you asked to hold.

---

## Decisions taken in review

Recorded here so the beads inherit them rather than re-litigating.

| # | Decision |
|---|---|
| D1 | **`bb-status.sh` is run-level, never bead-level.** The orchestrator runs it after each bead's acceptance check and at every wave boundary; a recovering session runs it as step 1; a woken durable trigger runs it. A **subagent never runs it** — it has no business auditing the run, and inside its own worktree the acceptance signals read differently by construction (its worktree exists because it is standing in it). |
| D2 | **The subagent record is an event bead, two per dispatch** (`dispatched`, `returned`) — not a wisp, not a row in the plan. Wisps are not synced, so they die with the container, which is the failure being designed against. A row in the plan makes a mutable blob with concurrent writers. An event bead is append-only, atomic, synced on the same push that makes closures durable, and independent of how many beads a subagent holds — which is the whole point, since a batch is where the current design breaks. Volume is ~2 per dispatch, not ~8 per bead. |
| D3 | **Claim with `bd update <id> --claim`, not `bd ready --claim`.** You are right and the review was wrong. `bd ready --claim` claims *the first ready issue matching the filters* — a pull model for a worker choosing its own work. An orchestrator dispatches a *named* bead, so it wants `bd update <id> --claim`: "atomically claim the issue (sets assignee to you, status to in_progress; idempotent if already claimed by you)". Same atomicity, right target. |
| D4 | **R13 is a spike, not a change.** Its output is a fact plus a follow-up bead, and the fact determines whether ~25 lines delete or stay. All seven unverified claims get the same treatment (wave 0). |
| D5 | **The retro drops the RFC format for self-authored findings.** Objective: identify high-quality issues, report each with a summary, let a human decide bead-or-not per issue. The RFC pipeline survives only for its original purpose — proposals arriving *from other projects* using the plugin, which `triage-rfc` exists to handle. |
| D6 | **R7 is held** until wave 5's end-to-end retro design lands. A finding taxonomy decided before the flow it feeds is a guess. |
| D7 | **"Ledger" means two different things and one of them is being deleted.** The *environmental-failure ledger* (`AGENTS.md`, and the dispatch manifest's copy) stays — R10(c) adds an entry to it. The *progress ledger* (`<id>.progress.jsonl`) is what R5 deletes. The collision is itself a defect; once R5 lands, "ledger" is unambiguous again and no rename is needed. |
| D8 | **A project config file is in scope, but for the project-level half only.** R8's tier keys are *per-bead metadata* and a config file does not touch them. What a config file does replace is the last of the `AGENTS.md` prose parsing: the `## model lineup` block, which the orchestrator reads as prose today. Pair it with R11's `checks` script and no executable or routing fact is parsed out of markdown any more. |

---

## Wave 0 — Spikes

Seven timeboxed investigations. All `type=spike`, all independent, all
parallelizable — this wave is a single concurrent dispatch. Each produces a
recorded fact and, where the fact says so, a follow-up bead.

| id | Spike | Gates | Tier |
|----|-------|-------|------|
| S1 | **What does `CLAUDE_CODE_SUBAGENT_MODEL` actually override?** Export it to a model outside the lineup, dispatch two subagents at two different resolved models, ask each what is serving it. | E6 — whether ~25 lines delete | light |
| S2 | **What do event beads and `bd set-state` cost at run volume?** Write 100 event beads and eight `set-state` transitions against a scratch bead; time `bd dolt push` before and after; then `bd delete` and confirm `bd doctor --check=pollution` is clean. | D2, M5 | light |
| S3 | **Does an epic-of-epics traverse?** Parent one epic to another; run `bd ready --parent <outer>` and `bd swarm validate <outer>` and see whether both reach the grandchildren. | P1, P2 — the multi-epic story | standard |
| S4 | **Does a Monitor over a polled command work as well as `tail -F`?** Arm one of each over a wave and compare latency and wake behaviour. | E1 — the ledger's only unique property | standard |
| S5 | **Does `bd` fire an issue-level script hook?** `OBSERVABILITY.md` documents a `hook.exec` span with `hook.event ∈ {create, update, close}` and a `hook.path`, but `CONFIG.md` documents no key and `bd config list` shows none. Grep the `bd` source for the hooks directory it resolves. | may supersede S4 entirely | standard |
| S6 | **Is `bd worktree create/remove` better than raw `git worktree` here?** The beads skill says always use it; this skill measured that `bd` already resolves the shared database through git's common directory, so the redirect files it configures may be moot. Its removal safety checks may still earn it. | N6 | standard |
| S7 | **Does embedded Dolt contend under concurrent subagents?** Dispatch a concurrent wave with `BD_OTEL_STDOUT=true` and read `bd_db_lock_wait_ms`. | how far concurrency scales | advanced |

**Acceptance for the wave:** every spike closes with a recorded measurement in
the form the repo's ledger admission rule requires — the command run, the
conditions, the result observed — and either a follow-up bead or an explicit
"no action".

---

## Wave 1 — Eliminate

Delete what the audit retired. First, so wave 2 does not renovate something
about to die.

| id | Item | Retires | Gated by |
|----|------|---------|----------|
| E1 | **Delete the progress ledger.** All eight phases, `bb-progress.sh`, the closed vocabulary, `progress-streaming.md`'s schema and writing sections, the "ledger is a hint" rule in four places, `recover-sprint` step 4's per-phase distrust table. Recovery re-runs gates. | 80-line script + ~150 lines of prose | S4 or S5 |
| E2 | **Delete `bb-backup-run.sh`** and with it the per-bead snapshot commits, the Phase 4 deletion commit, and the "the live directory is authoritative, the snapshot never is" rule in three documents. | 349 lines | E1 |
| E3 | **Delete `<id>.handoff.md`.** Fold its one `Caveat:` line onto the bead. Retires the write-before-close ordering rule and `recovery.md`'s crash-resilient-close section. | 1 file per bead + a non-negotiable rule | E1 |
| E4 | **Delete the `--gate-baseline` flag.** The baseline becomes unconditional; `--self-test` stays; no opt-out. | one way to get it wrong | N1 |
| E5 | **Delete the `tier:` label path** from `bead-bonanza`, including the `bd list --label tier:standard --json` example that returns 0 because `bd list` defaults to open. Metadata is the single home. | a documented example that cannot work | — |
| E6 | **Delete the `CLAUDE_CODE_SUBAGENT_MODEL` handling** — 4 script lines, 4 `SKILL.md` mentions, 2 of Phase 4's 3 special-case branches. | ~25 lines | S1 |
| E7 | **Delete `main-loop-fallback.md`** and make no-`Agent` a blocking pre-flight condition. | 54 lines describing a different product | X1 |

**Acceptance for the wave:** `SKILL.md` is measurably smaller, `scripts/`
holds three scripts rather than five, and no deleted mechanism is referenced by
anything that survives (`scripts/check-reference-links.py` proves the second
half).

---

## Wave 2 — Renovate

Replace hand-rolled procedure with a `bd` command or a script. Nothing
structural yet.

| id | Item | Replaces | Gated by |
|----|------|----------|----------|
| N1 | **Project-supplied `checks` script** (bead `agent-plugins-pq06`, **prioritized**). Pre-flight runs one known entry point instead of `bash -c` on strings parsed out of markdown bullets. | the gate-profile prose parser, and the whole "which bullets are commands" rule | — |
| N2 | **`bd update <id> --claim`** replaces `bd update <id> --status in_progress` at ceremony step 1. | a non-atomic claim, and half the reason the "never reset a claim you did not set" rule exists | — |
| N3 | **`bd merge-slot acquire/release`** around the merge. Serialized merging becomes a lock the whole run can see rather than a discipline the orchestrator maintains — the invariant that broke in run `1786778523-19`. | a prose merge policy | — |
| N4 | **Project config file** for the model lineup and run defaults (worktree root, push mode, run root). With N1 this retires the last executable or routing fact parsed out of `AGENTS.md`. | the `## model lineup` prose block | — |
| N5 | **Align bead metadata to `bd`'s documented convention** — `execution_suggested_model`, `execution_reasoning_effort`, `execution_parallel_group` alongside `tier`/`tier_reason`. Makes beads portable to any orchestrator and closes the `effort: unrecorded` hole in the manifest. Touches `groom-beads` as well. | a private key vocabulary | E5 |
| N6 | **`bd worktree create/remove`** if S6 says it earns its place. | raw `git worktree` plus a harness-placement trap | S6 |
| N7 | **`bd human`** for the `decision-required` bounce — a first-class list/respond/dismiss channel with a timeout. | a label plus a prose bounce procedure | — |
| N8 | **`bd gate --await gh:run:<id>`** for draft-PR CI, so "CI is the backstop" becomes an observed condition rather than an assertion. Also fixes the double-ask: `BB_PR_CI` is measured and then a second, uninformed ask decides the slim-gates policy. | an unobserved assumption | — |
| X1 | **Extract the per-bead ceremony into its own skill** (R15). Placed last in this wave deliberately: waves 1 and 2 both change the ceremony's content, and extracting first means editing it twice. What it buys — a subagent invokes it by name instead of receiving a paraphrase-prone template, a solo agent invokes it directly, and it becomes independently testable. | the do-not-paraphrase rule; one class of dispatch defect | E1–E3, N1–N2 |

---

## Wave 3 — The epic pivot

The structural change, and the one with the most downstream benefit. Stated as
a decision followed by everything the decision unlocks, because the value is
almost entirely in the second part.

### The decision

**A run is an epic. The epic is the queue.** Grooming creates or names a sprint
epic; the run's beads are its descendants.

### What follows from it

| id | Consequence | What it retires |
|----|-------------|-----------------|
| P1 | **The epic id is the run id.** No generated, unguessable `<timestamp>-<n>` that survives only in the orchestrator's transcript. Every artifact is found by an id a human can read, type, and `bd show`. | the run-id-is-unrecoverable problem, and the `current` symlink |
| P2 | **`bd ready --parent <epic>` is the queue** — scoped and dependency-ordered from one durable fact. | the "enumerate, never filter" rule, stated in four documents |
| P3 | **The dispatch plan attaches to the epic** — its `design` field or a child bead, decided in the bead. Recovery reads `bd show <epic>`. | `$RUN_DIR/dispatch.md`, and the run directory as a load-bearing store |
| P4 | **`bd swarm create <epic>`** registers the run, and `bd swarm list` shows every run in flight across the repo. | nothing today shows a run exists |
| P5 | **`bd swarm validate <epic>` supplies the dependency waves.** `plan-sprint` stops re-deriving them and keeps only what `bd` cannot compute: C4 co-edit collisions and dispatch-mode packing. It also inherits the integrity checks — direction, orphans, cycles, disconnected subgraphs — which `plan-sprint` has none of. | half of `plan-sprint`, plus a class of grooming defect nobody checks for |
| P6 | **`bd swarm status --json` is the live run view** — completed / active / ready / blocked, computed from beads. A run currently has no status view at all. | the run checklist as the only progress surface |
| P7 | **`bd epic status` is the run-level Definition of Done** (R12): all children closed is the mechanical half; one paragraph asserting the epic's success criteria are met, with the command that shows it, is the agentic half. | a run that ends when the queue drains rather than when the outcome is demonstrated |
| P8 | **An epic-level Definition of Ready in `groom-beads`.** Your point, and the cheaper half of the whole enumeration argument: a queue that changes mid-run is a grooming discipline failure, not a tracker limitation. `groom-beads` has a DoR per bead and none per epic. | the reason enumeration was introduced |
| P9 | **Grooming wires sequencing constraints as real `blocks` edges** rather than emitting them as advice a planner re-derives. A `blocks` edge is durable, visible to `bd ready`, honoured by `bd swarm validate`. Co-edit collisions stay with `plan-sprint` — there is no edge type for "must not share a wave". | an advisory line misread as a dependency in run `1786778523-19`, costing a wave |
| P10 | **Multi-epic runs via epic-of-epics**, if S3 confirms traversal. A cross-epic sprint becomes an outer epic parenting the feature epics. | the one property a sprint epic would otherwise cost |

**Acceptance for the wave:** a run can be started, recovered and reported on
using only an epic id; `$RUN_DIR` holds nothing recovery needs; `plan-sprint`'s
wave derivation is `bd swarm validate` plus a packing pass.

---

## Wave 4 — Mechanize

Answering §5.3's question directly: **mechanization is not one wave, it is
three**, because each tier depends on a different precondition. Separating them
is what keeps the first tier from waiting on the third.

### M-A — mechanizable today, no preconditions

Already shell one-liners; they only need collecting into a script.

| id | Item |
|----|------|
| A1 | The five acceptance signals — `bd show`, `git log --grep`, `git branch --contains`, `git worktree prune && git worktree list`, `bd export \| jq` |
| A2 | The skip check (`bb-verify-skip.sh` exists; wire its exit code into the status output rather than a prose procedure) |
| A3 | The landing check — `git fetch && git log <remote>/<branch>..HEAD` must be empty |
| A4 | Report completeness — grep the Phase 4 report for each required heading and for the explicit `none`/`N/A` nulls |
| A5 | Findings-harvest non-zero assertion |

### M-B — needs the epic pivot (wave 3)

Everything that must key off a stable run handle.

| id | Item |
|----|------|
| B1 | **`bb-status.sh <epic-id>`** — the run checklist with every mechanical item independently verified and the command that established it printed beside. Per D1: orchestrator and recovering sessions only. |
| B2 | Plan structural validation — every queued bead appears in exactly one wave row with tier, bucket and push mode filled |
| B3 | Readiness gate as a check — `bd lint` clean, no `decision-required`, `metadata.tier` non-null across `bd ready --parent <epic>` |
| B4 | `recover-sprint` collapses to "run `bb-status.sh`, then act" |
| B5 | **The subagent dispatch record** (R4/D2) — two event beads per dispatch, and `bb-status.sh` reconciles dispatched against returned |
| B6 | **`bd set-state <id> health=…`** as the watchdog signal. The one thing neither git nor status can express: a stuck agent, a degraded run, a wave that has stopped progressing |

### M-C — needs the checks script and config file (wave 2)

| id | Item |
|----|------|
| C1 | Gate execution and the baseline, through `checks` rather than parsed prose |
| C2 | Targeted-tests-by-surface as data the script resolves, not a list an agent reads |
| C3 | Tier → model resolution from config rather than a prose block |

---

## Wave 5 — Retro rebuild

Independent of waves 3 and 4. Its first item is the design you asked to hold
R7 behind.

| id | Item |
|----|------|
| T1 | **Design the retro end to end** — capture, harvest, classify, report, decide, file. One document, covering what each stage produces, who performs it, and where the output lives. Everything below is decided inside it. |
| T2 | **Replace self-authored RFCs with a triage list** (R6/D5). The sprint report keeps every finding verbatim — that is the substrate recurrence detection runs on. On top of it, the orchestrator produces N candidate issues, each with a proposed title, priority, one-paragraph summary and the finding ids it rests on, and asks once which to file. Accepted ones become beads with a `discovered-from` edge and a citation to the report. |
| T3 | **The finding taxonomy** (R7), decided inside T1's flow rather than ahead of it. |
| T4 | **Scope the RFC pipeline to its real purpose** — proposals arriving from other projects using the plugin. `triage-rfc` keeps that; the `origin: self` path retires. Decide what happens to the 27 already queued. |

---

## Wave 6 — Split

Last, per your sequencing. By now the phases have been shrunk by wave 1,
rewritten by wave 2, restructured by wave 3 and partly replaced by scripts in
wave 4 — so the split is a smaller and better-informed cut than it would have
been first.

| id | Item |
|----|------|
| X2 | **`start-sprint`** — Phases 0, 1 and 1.5 (pre-flight, queue, readiness gate) |
| X3 | **`close-sprint`** — Phase 4 (report, retro, teardown) |
| X4 | **What stays in `bead-bonanza`** — the Phase 3 loop, the acceptance check, the worktree lifecycle, the operating principles. The parts a compacted orchestrator genuinely needs re-injected |
| X5 | **Retire the compaction defences the split makes unnecessary** — the "re-read from Phase 0" instruction, the out-of-numeric-order apologia, and the `--strict` known-red exception in `AGENTS.md` |

**Acceptance:** `scripts/check-skills.sh --strict` exits 0 with no warnings —
the first time it has.

---

## What each wave unblocks

```
S1 ──► E6
S3 ──► P1, P2, P10
S4/S5 ──► E1 ──► E2, E3
S6 ──► N6
N1 ──► E4, C1
E1–E3, N1–N2 ──► X1 ──► E7
Wave 3 ──► B1–B6
Waves 1–4 ──► Wave 6
```

Wave 5 has no inbound edges.

---

## Open decisions this plan still needs

1. **Where the dispatch plan lives on the epic** (P3) — the `design` field, or a
   child bead. Depends on whether `bd` bounds field size; untested.
2. **Whether `$RUN_DIR` survives at all.** After E1–E3 and P3 it holds
   `trailers.txt`. That could move to the epic too, or stay as the one
   genuinely per-session file.
3. **What happens to the 27 queued RFCs** when the self-RFC path retires (T4).
4. **Whether wave 0 runs as a bead-bonanza sprint.** It is seven independent
   spikes with no shared surface — the best-shaped concurrent wave this
   backlog has produced, and a real test of the skill as it stands before any
   of it changes.

---

## The workflow, as it changes

**Where the live map lives.** The research document's §1 is the workflow **as
measured on 2026-08-19** — a frozen baseline, and it stays frozen, because a
research record that quietly rewrites itself stops being evidence. This section
is the **living map**: every wave that changes a step records the delta here,
keyed to the baseline's step numbers, so the two read side by side as
before-and-after.

Update this table in the same commit as the change, never after. An entry names
the step, the operation, and the wave item responsible.

### Delta log

*Empty — no wave has executed yet. Format below, with the deltas this plan
already implies, marked `planned` until their bead closes.*

| Step | Op | Change | By | Status |
|------|----|--------|----|--------|
| 0.1 | update | `--gate-baseline` removed; baseline is unconditional | E4 | planned |
| 0.1.5 | delete | gate-profile prose parsing — replaced by the `checks` script | N1 | planned |
| 0.1.7 | update | `$RUN_DIR` no longer load-bearing; `current` symlink retires | P1, P3 | planned |
| 0.1.12 | delete | `CLAUDE_CODE_SUBAGENT_MODEL` detection | E6 | planned |
| 0.1.16 | update | gate baseline runs via `checks`, always | N1, E4 | planned |
| 0.3 | update | `trailers.txt` — may move to the epic (open decision 2) | P3 | planned |
| 0.6 | update | becomes a **blocking** condition; `BB_ASSUME_NO_AGENT` asserts it | E7, R14 | planned |
| 1.2–1.4 | update | queue becomes `bd ready --parent <epic>` | P2 | planned |
| 1.5.1b | update | readiness gate becomes a mechanical check | B3 | planned |
| 1.5.2b | update | tier written to `execution_*` keys alongside `tier` | N5 | planned |
| 1.5.4 | update | bounce goes through `bd human` | N7 | planned |
| **1.5.x** | **add** | epic-level Definition of Ready | P8 | planned |
| **1.5.y** | **add** | grooming wires sequencing constraints as `blocks` edges | P9 | planned |
| 2.5.1 | update | manifest attaches to the epic, not `$RUN_DIR/dispatch.md` | P3 | planned |
| 2.5.2 | delete | "pass enumerated ids, never the filter" | P2 | planned |
| 2.5.3–2.5.5 | update | dependency waves from `bd swarm validate`; `plan-sprint` keeps co-edit packing only | P5 | planned |
| **2.5.z** | **add** | `bd swarm create <epic>` registers the run | P4 | planned |
| 3.0 | update | durable trigger carries the epic id alone | P1 | planned |
| 3.5 | update | Monitor source changes (polled `git log`, or a `bd` hook) | E1, S4/S5 | planned |
| **3.x** | **add** | dispatch record — two event beads per subagent | B5 | planned |
| 3.6.1 | update | `bd update <id> --claim` | N2 | planned |
| 3.6.3–3.6.4 | update | progress events removed; gates run via `checks` | E1, N1 | planned |
| 3.6.5 | delete | handoff file; caveat folds onto the bead | E3 | planned |
| 3.6.8–3.6.9 | delete | `committed` / `pushed` events | E1 | planned |
| 3.9 | update | five signals become `bb-status.sh` output | A1, B1 | planned |
| 3.9.3 | update | merge serialized by `bd merge-slot` | N3 | planned |
| 3.10 | delete | `bb-backup-run.sh` snapshot | E2 | planned |
| 3.12 | update | landing check inside `bb-status.sh` | A3 | planned |
| 4.1–4.7 | update | report sections checked mechanically | A4 | planned |
| 4.4–4.5 | update | both `N/A` branches delete if S1 says the env var is inert | E6 | planned |
| 4.8.6–4.8.7 | update | RFC route replaced by the triage list | T2 | planned |
| **4.x** | **add** | run-level Definition of Done via `bd epic status` | P7 | planned |
| 4.10 | update | trigger disposition read from the epic | P1 | planned |
| 4.11 | delete | `.bonanza/<run-id>/` removal commit | E2 | planned |
| X.1 | update | subagent crash recovery collapses to `bb-status.sh` | B4 | planned |
| X.3 | delete | main-loop fallback | E7 | planned |
| **skills** | **add** | per-bead ceremony skill | X1 | planned |
| **skills** | **add** | `start-sprint`, `close-sprint` | X2, X3 | planned |

**Counting the deltas:** 8 deletions of whole steps, 24 updates, 7 additions,
and 3 new skills — against a baseline of roughly 95 steps. The net is a shorter
procedure with more of it mechanized, which is the point.

### How to keep it current

- A bead that changes a step **updates its row and flips the status** in the
  same commit. A row still reading `planned` after its bead closes is a defect.
- A bead that changes a step this table does not list **adds a row**. The table
  is the record of drift between the baseline and the running procedure.
- When every row reads `done`, regenerate §1 of the research document as a
  second dated snapshot — not by editing the first — and this plan can be
  deleted, per the `docs/plans/` convention.

