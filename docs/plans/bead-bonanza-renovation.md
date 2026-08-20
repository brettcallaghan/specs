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

### Wave 0 results — CLOSED 2026-08-20

All seven settled. Full measurements live in each bead; the verdicts and what
they change:

| id | Verdict | Effect on the plan |
|----|---------|--------------------|
| S1 `epju` | **Split.** The per-invocation `model` parameter demonstrably routes when the variable is unset — two subagents served by `claude-haiku-4-5-20251001` and `claude-opus-5`, read from their transcripts. The variable was `NOTSET` and **cannot be injected into a running session**, so the precedence claim is unverified. | **E6 BLOCKED.** New gate `pr8l.1` probes it with the variable set at launch. Do not delete the routing tables until then. |
| S2 `wyun` | Event beads cost 0.68s each and take `bd dolt push` from 0.63s → 5.36s (8.5x); deleting them returns it only to 1.56s, because Dolt keeps history — **the weight is permanent**. `set-state` is a replacing label, no churn. Metadata takes nested objects, arrays and 100 KB values, and merges across writes. | **D2 needs a volume budget or a cheaper carrier.** M5 cheap. **D9 works — but written via `bd update --metadata`, not through the graph file**, whose importer is `map[string]string`. |
| S3 `—` | `bd ready --parent` traverses two levels; `bd swarm validate` does not. | P4 runs per wave epic. |
| S4 `0len` | **Superseded by S5.** | Not run. |
| S5 `89gy` | **Issue hooks exist** — `.beads/hooks/{on_create,on_update,on_close}`, executable, `hook <id> <event>` with issue JSON on stdin, async and timeout-bounded. No config key, which is why the docs search found nothing. **`on_close` does NOT fire on `bd update --status closed`** — this project's own close path — but `on_update` does. | **E1 unblocked.** The push signal comes from the tracker write itself. Use `on_update`. Hazard: `.beads/hooks/` is bd's git-hook directory, is tracked here, and `bd init` rewrites it. |
| S6 `7axi` | Database sharing is **identical** both ways (`no redirect` in both). The one real difference: on a clean worktree holding **unpushed commits**, `git worktree remove` deletes it silently while `bd worktree remove` refuses. | **N6 adopts it, on removal safety only.** Two hazards to handle: it edits the main repo's tracked `.gitignore`, and a bare name nests the worktree inside the repo. |
| S7 `xf70` | **No failures at any concurrency — it serializes instead.** 1/2/4/8 writers all sustain 3.3–4.0 writes/sec while wall time scales linearly; a tracker write is ~0.28s of globally exclusive time. `BD_OTEL_STDOUT` emits traces, not the metrics (1.6 MB per single write) — outcome measured directly instead. | Budget tracker writes per bead; prefer one write carrying several facts. Compounds S2's verdict on D2. |

Two facts worth carrying beyond this plan: a subagent **cannot** learn its own
model from `get_session` (it reports the parent session's), and `bd doctor` is
**unavailable in embedded mode**, which is what this project runs.

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

### Wave 1 results — 4 of 5 done 2026-08-20; E6 blocked

`SKILL.md` **58,392 → 53,599 bytes** (−8.2%). `scripts/` holds **three**
scripts, down from five. `check-skills.sh` and
`sync-shared-references.sh --check` both pass, with only the pre-existing
body-size warning.

| id | Bead | Outcome |
|----|------|---------|
| E1 | Progress ledger | **Done.** `bb-progress.sh`, the eight-phase vocabulary, the schema and writing sections, the "ledger is a hint" rule in every place, and `recover-sprint`'s per-phase distrust table all deleted. Monitor and durable trigger retained, re-sourced onto `bd`'s `on_update` hook. Recovery and `recover-sprint` now classify from worktree + tracker and **re-run gates** instead of trusting a recorded pass. |
| E2 | `bb-backup-run.sh` | **Done.** Script, Phase 3 step 5, Phase 4 step 6b, the "live directory is authoritative" rule in all three documents, and `recover-sprint`'s snapshot route all deleted. The "read the merge target at merge time" rule survives with a new cause: the head still moves, now from the bead merges themselves. |
| E3 | Per-bead handoff file | **Done.** Folded to a `caveat:` comment on the bead, written only when there is one. `bb-close-bead.sh` loses `--run-dir` and the minimal-handoff block with it (9,960 → 8,277 bytes) — the flag existed solely to serve that block. |
| E5 | `tier:` label path | **Done.** Metadata is the single home, matching what grooming already wrote. The broken example is replaced by one measured here: `bd list --status all --json \| jq '[.[] \| select(.metadata.tier == "standard")] \| length'` → **110** (320 beads, 205 tiered). The old `--label` form returns **0**. |
| E6 | `CLAUDE_CODE_SUBAGENT_MODEL` | **BLOCKED** on `pr8l.1` — see Wave 0, S1. Not executable from inside a session. |
| E4, E7 | | Not in this wave's filed beads; carried forward. |

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

## Decisions resolved in the second review

### D9 — Where the plan lives on the epic

**The wave table is not stored at all — it is derived.** Once P5 takes the
dependency waves from `bd swarm validate` and P9 wires sequencing constraints as
real `blocks` edges, the wave table is a *view over the graph*, recomputable at
any moment and stale the instant it is written down. Storing it creates the
exact drift the amendments log exists to police. That collapses the question
from "where does the plan live" to "where do the two things that are *not*
derivable live":

| Artifact | Home | Why |
|---|---|---|
| **The manifest** — gate profile, gate baseline, model lineup, effort, standing authorization, environmental ledger, base sha | `--metadata` JSON on the epic | Structured, queryable, small, read by machines. `bd show <epic> --json \| jq .metadata` is the whole retrieval. |
| **The amendments log** | `bd comment` on the epic, prefixed `amend:` | Append-only and timestamped by construction — no concurrent-write blob, no ordering rule to state. It is the same channel `retro:` findings already use, which means one harvest pipeline covers both. |
| **The wave table** | derived, `bd swarm validate <epic>` | Recomputed, never recorded |

Options considered and rejected:

- **The `design` field.** Free text, so it holds the manifest and the amendments
  together — but it is a *mutable single blob*. Every amendment is a
  read-modify-write of the whole field, which is precisely the concurrency
  hazard the plan's own rules avoid elsewhere, and it makes the structured half
  unqueryable.
- **A child bead holding the plan.** Durable and appendable, but it puts a
  non-work item in the epic's child set, which corrupts `bd epic status`,
  `bd ready --parent` and `bd swarm validate` — every graph command the pivot
  exists to unlock. Self-defeating.
- **A committed file** (`docs/sprints/<epic>.md`). Durable and diffable, but
  re-introduces a second store the tracker has to agree with, and a run's plan
  is not a durable project artifact — the sprint report already is.
- **A poured molecule** (`bd mol pour`). The most interesting alternative: it
  makes the plan *be* the bead graph. Rejected for now only because it requires
  a formula and the epic already carries the graph; worth revisiting once
  wave 3 lands.

*Unverified:* whether `bd` bounds metadata size, and whether `--metadata`
round-trips a nested object. `bd create --graph` types metadata as
`map[string]string`, so **array-valued metadata cannot go through the graph
format** and must be set afterwards with `bd update --metadata '<json>'`. Folded
into S2.

### D10 — `$RUN_DIR` holds more than `trailers.txt`, and that is the point

You are right, and the plan understated it. The run-level facts injected into
**every** child prompt are the gate profile (`{FORMAT}` `{LINT}` `{BUILD}`
`{TEST}` `{REGEN}` `{CUSTOM_GATES}`), the **gate baseline**
(`{GATE_BASELINE}`), the **environmental-failure ledger** — the known-red
verdicts, diagnosed once and passed forward as fact — plus the model lineup,
the standing authorization, and `{TRAILERS_FILE}`.

All of those live in `dispatch.md`'s manifest today. So they are not an
argument for keeping `$RUN_DIR`; they are the **contents of D9's manifest**, and
D9 already moves them to the epic. Once it does:

- gate profile → resolved by the `checks` script (N1), not stored
- gate baseline, environmental ledger, lineup, standing authorization →
  epic metadata (D9)
- trailers → epic metadata, except the session URL, which
  `bd update <id> --session` records natively and better (this also closes bead
  `agent-plugins-oyf`)

**`$RUN_DIR` then holds nothing, and the concept retires** along with the run
root, the `current` symlink, and the run id. That is a cleaner outcome than the
plan claimed, and it makes P3 load-bearing rather than tidy-up.

### D11 — The RFC path is partitioned by audience, not retired

Correcting D5. The split is **whose problem is it**:

- **`bead-bonanza`'s retro finds issues in the user's project** — a spec that
  was wrong, a defect the sprint surfaced, work the sprint implied. Those
  become **beads in that project**, decided one at a time by a human. No RFC:
  the audience is the project itself, and it has a tracker.
- **`retro-skill` and `retro-spec` find issues with the skills this project
  vends.** Those stay **RFCs**, because they are proposals about a product
  shipped to other projects, they need the evidence-and-proposals structure,
  and `triage-rfc`'s dedupe-and-defer machinery is what accumulates recurrence
  across projects. Only this repository cares about them, and they must not be
  folded into a downstream project's backlog.

So the retro's first act becomes a **routing decision**, and it is mechanical
once a finding names its target (R7/T3). T4 changes from "retire the self path"
to "make the audience split explicit in each skill", plus a review pass over
the 27 queued RFCs.

### D12 — Wave 0 is run by hand, using the commands we intend to adopt

Adopting your inversion. The skill as it stands has been exercised over five
sprints; another run teaches little. Wave 0 instead becomes a **deliberate
trial of the target state**: create the epic, nest the wave epics, claim with
`bd update <id> --claim`, drive status with `bd swarm status --json`, validate
with `bd swarm validate`, and record what each command actually does.

This has a useful side effect: **filing this plan's own beads exercises S3
immediately.** The bead graph below is a root epic parenting seven wave epics
parenting the work — so `bd ready --parent <root>` and
`bd swarm validate <root>` either traverse two levels or they do not, and we
find out before anything depends on it.

---

## Filing status — FILED

**52 beads filed 2026-08-20** with `bd create --graph`, after the tracker was
repaired (see below). Root epic **`agent-plugins-1ilc`**. The graph file and
its generator were deleted once consumed — they are a second copy of issue
content, they cannot be re-applied without minting duplicates, and the tracker
is the source of truth. `git show 5709a7b` has them if the construction ever
needs auditing.

| Wave | Epic | Leaves |
|---|---|---|
| 0 Spikes | `agent-plugins-c3zl` | 7 |
| 1 Eliminate | `agent-plugins-pr8l` | 5 |
| 2 Renovate | `agent-plugins-1qc8` | 12 |
| 3 Epic pivot | `agent-plugins-bir9` | 8 |
| 4 Mechanize | `agent-plugins-ob33` | 4 |
| 5 Retro rebuild | `agent-plugins-bmhe` | 5 |
| 6 Split | `agent-plugins-8t8v` | 3 |

Verified: **52 created, 0 dependency cycles, 0 `bd lint` warnings**, tracker
pushed.

### The review pass found two defects, both fixed

1. **All seven spikes linted dirty.** `bd lint` requires `## Goal` and
   `## Findings` for `type=spike` — not `## Acceptance Criteria`. Every spike
   was restructured to Goal / Context / Method / Findings, with the Findings
   section carrying the ledger admission rule as its placeholder.
2. **The root epic still said `## Acceptance Criteria`.** The pre-flight fix
   reached the seven wave epics but missed the root. `bd lint` wants
   `## Success Criteria` on `type=epic`.

Both are the same underlying gap: **`bd lint`'s per-type section requirements
are not documented anywhere the authoring skills read**, so a graph is written
against the wrong headings and only linting reveals it. Worth a bead against
`groom-beads`.

### Spike S3 is already answered — by the act of filing

Recorded and closed as `agent-plugins-iwxu`. The graph is deliberately two levels
deep, so filing it *was* the experiment. **Split verdict:**

- `bd ready --parent <root>` returned 12 items **including grandchildren** —
  **it traverses.** But it also returns the wave epics themselves, which are not
  workable; `--exclude-type=epic` gives the real queue.
- `bd swarm validate <root>` reported **"Total issues: 7"** and analysed only
  the direct children — **it does not traverse.** One level down, on a wave
  epic, it correctly resolved 12 leaves into 3 ready fronts.

**Consequences:** P4 must run `swarm validate` **per wave epic**, never on the
sprint root, where it returns a useless answer that looks valid. P8 is viable
for the queue but not for wave derivation — a cross-epic sprint needs
`swarm validate` per member epic, composed.

## How the tracker was repaired

The blocker was never `sync.remote`, and never the schema cursor alone. A
v1.2.x migration moved the `events` audit table into `dolt_ignore`; ignored
tables are never staged or committed, so the table never travelled to any
clone, while the v65 cursor — which lives in the tracked `schema_migrations`
— did. Every `bd` write that records an audit event therefore failed with
`table not found: events` in every clone, while reads and `bd comment` worked.

**bd's official mechanism cannot self-heal this.** `bd migrate schema` is
cursor-driven (`INSERT IGNORE INTO <cursor> (version, content_hash)`), so
migration `0005_create_events` is recorded as applied and skipped even though
its table is gone. There is no drift detection. That is a bd defect
independent of the v1.2.1 accident, and it is why the repair had to be manual.

The repair, in two parts:

1. **Cursor rollback** (the guide's recommended fix) — restored the schema
   version and replicated correctly. Verified: a fresh clone stopped warning
   about skew.
2. **Re-create and track the `events` table** — the part the guide only
   covers for a machine that still *has* the table. No reachable copy did, so
   `DOLT_ADD -f events` had nothing to stage. The table was created from
   **bd's own DDL**, obtained by running `bd init` into an empty directory and
   reading `SHOW CREATE TABLE events` from the result, then `DOLT_ADD -f` +
   `DOLT_COMMIT` + push.

**Getting the DDL from bd rather than by hand mattered.** A reconstruction
composed here from the migration chain matched on every column, index and
foreign key — and omitted
`ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin`. That
collation is *binary*; the server default is case-insensitive, and the column
carrying it is the foreign key into `issues(id)`. Never hand-compose a schema
when the tool will emit it.

### One stray artifact

A one-word `probe` comment landed on `agent-plugins-92q` while establishing
which write paths work (`bd comment` is the one that succeeds). `bd comments`
has no delete subcommand, so it remains. Harmless, and named here rather than
left for someone to wonder about.
