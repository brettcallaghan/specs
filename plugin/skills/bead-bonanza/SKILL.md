---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use
  when the user asks to work on multiple beads at once — e.g. "work on all open
  beads with label X", "address all P1 beads", "clear the bead backlog". Triages
  the queue, gates every bead for readiness (grooming under-specified beads via a
  subagent, or bouncing what can't be made ready), surfaces both explicit and
  implicit decisions up front, dispatches via the cheapest viable execution mode
  (doc fast lane / shared-surface batch / sequential), runs every bead in its
  own git worktree, streams per-bead progress, and lands each bead in a single
  commit.
version: "1.6"
---

# bead-bonanza

Batch-process beads via subagent orchestration. The orchestrator (you) stays
small — you only hold the queue, the user's decisions, the dispatch plan, and
one-line outcomes per bead. Each bead runs in a fresh subagent so context
doesn't compound.

Read `CLAUDE.md` and `AGENTS.md` once at the start so you have repo conventions
in scope before delegating.

**After any compaction, before dispatching anything:** re-read this file from
Phase 0 onward, then invoke the **`recover-sprint`** skill. A compacted session
re-injects only this file's first ~5,000 tokens, so the per-bead ceremony, the
published script interfaces and the rule against paraphrasing a dispatch
template are NOT in your context even though you have "read the skill".
`recover-sprint` re-derives run state from the durable sources (the branch,
`dispatch.md`, `git worktree list`, the tracker) rather than from a summary of
it, and a skill name survives truncation where a relative path does not. Then
re-run the queue query.

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
(It lives in your response, not in a task tool — nothing external to render,
nothing to fall out of sync with `bd`.)

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

**How this document is ordered.** The executable core comes first — worktrees
and merges, Phase 3's per-bead ceremony, the reference and script interfaces —
because only this file's opening survives a compaction, so the phases appear
out of numeric order.

## Every bead runs in its own worktree

**This is universal, not opt-in.** Each dispatched bead gets a git worktree
named for its hash suffix (`jh6`, `92q.14` — not the full id), created from the
sprint branch and carrying a branch of the same name. Nothing is implemented in
the shared checkout except by the orchestrator itself.

**Presence is the in-flight signal.** A worktree that still exists means its
bead is not yet accepted — the orchestrator removes it as the final act of
accepting the bead, after the branch merges. `git worktree list` is the
complete answer to "what is still in flight": there is no mtime heuristic
anywhere, and no need to re-derive which paths a bead touched.

**Hard gate — every worktree must branch from the sprint's own branch, never
the default branch.** Verify `BB_WORKTREE_BASEREF` is `head` **before the first
worktree of each wave, and do NOT dispatch otherwise.** It is global user state
that no repository owns, so anything could change between waves: re-verify
before **each** wave, not once per run.

**Merge policy.** The **orchestrator** merges each bead's branch into the
sprint branch as it is accepted, one at a time; subagents never merge. A
chained batch is merged by fast-forwarding the LAST bead's branch, which
carries its predecessors — not by merging each bead's branch in turn. Either
way the result is linear history and one commit per bead. What a `--ff-only`
refusal means differs between a sequential and a concurrent wave, and reading
it wrong stops a healthy wave — take that rule from the lifecycle named below
before you merge anything.

Where the worktree lives, how `bd` behaves inside one, the base-ref remedy, and
the create/merge/remove lifecycle are all in `references/subagent-prompts.md`,
"Worktree mechanics" — read it at dispatch, which is when the first worktree of
the run gets created.

## Phase 3 — Execute

Execute buckets in this order (lowest blast radius first):

1. **Doc/chore fast lane** — single subagent, one worktree at a time.
2. **Shared-surface batches** — one subagent per batch, sequentially across
   batches, one worktree at a time within each.
3. **Plain sequential** — one subagent per bead, in `bd ready` order.

Within a bucket, dispatch `isolated` beads concurrently where Phase 2.5's entry
conditions hold; otherwise one at a time. Either way, merge and remove each
bead's worktree as it is accepted — concurrently dispatched beads are merged
one at a time, in the order they return.

After each subagent returns:

1. Append its summary lines to your log (one per bead processed).
2. Mark items on the run checklist complete as you go.
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
4. **Acceptance check — every bead, every dispatch mode.** A bead the subagent
   returned as `completed` is not closed until **every signal below** agrees.
   Verify them in this order and stop at the first that fails:

   1. **The bead reads closed in the tracker** — `bd show <id>` shows
     `CLOSED`.
   2. **A commit exists on the bead's branch** with the subject
     `<bead-id> P<n> <title>` — `git log --oneline --grep=<bead-id> -1`.
   3. **The branch merged into the sprint branch.** The merge is `--ff-only`,
     so a refusal is not a conflict to resolve here but an invariant that
     broke (see "Merge policy") — stop the wave. After the fact, confirm with
     `git branch --contains <sha>` and look for the sprint branch.
   4. **The worktree is GONE** — `git worktree prune && git worktree list` no
     longer lists it. **This is the load-bearing signal.** Removal is the
     orchestrator's own final act of accepting the bead, so a surviving
     worktree is a bead not accepted, whatever the subagent reported: it is
     state rather than testimony, and no subagent summary can forge it.
   5. **Findings were recorded** — at least one `retro:` comment on the bead,
     or the subagent's explicit "clean bead, no findings". A silent bead is
     neither, and is not accepted. Check with the export pipeline:

     ```
     bd export | jq -r 'select(.id=="<bead-id>")
       | [.comments[]?|select(.text|startswith("retro:"))] | length'
     ```

     Not `bd show <id> --json` — it reports `comment_count: 0` for a bead that
     has comments, so it fails every bead in the run at once.

   Applies **per bead, not per subagent**. A fast-lane subagent that
   processed six beads is checked six times — that is exactly where a silent
   omission hides.

   **A bead failing any signal is reported, not recorded closed.** Give it a
   Phase 4 row naming the failing signal (`acceptance-failed: worktree
   survives`) in place of a commit sha, and say so in the run totals rather
   than counting it among the closed. Signals 1, 2 and 5 are repairable in
   place — close the bead, or add the missing `retro:` comment from the
   returned summary — and the repair is stated in the report. Failures of 3
   or 4 are not the orchestrator's to repair inline: the merge or the
   acceptance never happened, so leave the worktree standing (it is the
   in-flight signal) and surface it.

   **A standing worktree is one of three things — read signal 4 against the
   dispatch mode before calling it a failure.**
   1. *Removal still pending.* Removal is the orchestrator's own final act, so
      between a bead's return and that act the worktree ALWAYS survives. Do the
      removal (`git worktree remove <path>`), then re-run
      `git worktree prune && git worktree list`. This is the normal state, not
      a failure.
   2. *A batch is still running.* A batch dispatched into one
      harness-provisioned worktree holds it for EVERY bead it carries, so beads
      already closed inside it have a surviving worktree by construction. Accept them on
      signals 1–3 and 5, and read signal 4 once, against the batch, when it
      returns. Charging each bead for a worktree its siblings are still using
      fails the whole batch but the last bead.
   3. *A real signal-4 failure.* Removal itself failed, or a worktree the
      orchestrator never created survives. Leave it standing and surface it.


   **Signals this check deliberately does NOT use.** A
   `"phase":"committed"` ledger event is written by the agent doing the work
   and stops when that agent stops — see `references/progress-streaming.md`,
   "the ledger is a hint, never the authority" — so its absence proves nothing and
   its presence proves only that a line was appended, not that the sha in it
   is reachable. And there is no per-bead retro FILE: findings are `retro:`
   comments on the bead, and `<bead-id>.handoff.md` is a recovery anchor that
   carries no findings and is not evidence of acceptance.

   **A subagent that never returns is checked by these same five signals.** A
   lost completion notification is not evidence of an unfinished bead, and it
   is common enough to plan for. The five signals read state, not testimony,
   so they hold with or without a return — run them unchanged. To decide
   whether a bead is still *working* rather than finished, read its
   `<bead-id>.progress.jsonl` last phase **together with** `bd show <id>`: a
   terminal phase (`committed`/`pushed`/`blocked`) plus a closed tracker
   record is conclusive. The ledger is still never an acceptance signal: it is
   read here only to separate "still running" from "finished", and only
   alongside tracker state. Not `bd list --status in_progress` — it is a query
   over a moving set and has disagreed with both durable signals.
   **Never re-dispatch on a missing return alone.**

   **Note a missing terminal event, do not act on it.** WHERE a subagent
   returned `completed` but its progress file carries no `committed` event,
   the Monitor never fired for that bead — `references/progress-streaming.md`
   filters to `committed|blocked` — and the user saw silence for its whole
   run. This does not affect acceptance, since the check above reads state
   rather than the ledger, but say so in the Phase 4 report: an unremarked
   gap trains the reader to trust a stream that has silently stopped covering
   part of the run.
5. **Snapshot the run directory** — `scripts/bb-backup-run.sh "$RUN_DIR"`, run
   from your own checkout on the sprint branch and **never from a bead's
   worktree**. Run it after **every** bead's acceptance check, not once per
   wave — a copy refreshed once mid-run is indistinguishable from a complete
   one, and it prints `unchanged` without committing on a quiet bead.
6. **Verify the working tree is clean** with `git status --porcelain`. If dirty,
   fall through to the recovery procedure below.
7. If the bucket used `defer-push`, run `git push` for the bucket now.

   **A wave is not complete until its push is verified landed.** After
   pushing, fetch and confirm nothing is left behind:

   ```sh
   git fetch <remote> <branch>
   git log --oneline <remote>/<branch>..HEAD    # must be empty
   ```

   A push can be rejected for reasons that leave the tree clean, the
   worktrees removed and the beads closed — every other signal reads
   "finished". This is the only one that distinguishes a landed wave from a
   rejected one, and it costs one fetch. A credential refused for
   `.github/workflows/**` rejects the WHOLE push, so one workflow file
   strands every commit behind it. `recover-sprint` step 3b runs the same
   check from the other side; keep the two in the same form.
8. Move to the next bucket/bead.

**Circuit breaker:** if 3 consecutive beads return `blocked` or fail, stop the
run and surface to the user. Something is systemically wrong (broken main,
environment issue, an under-groomed bead).

### Recovery and progress streaming

Three situations, three loads — take the one you are in, not the whole corpus.
A subagent past 5 minutes or any question about the progress events:
`references/progress-streaming.md`. A subagent that crashed or timed out
mid-ceremony: `references/recovery.md`. Resuming a run whose state you do not
hold, because the *orchestrator* died or compacted: invoke the
**`recover-sprint`** skill — a skill rather than a file, because the session
that needs it may be fresh, with no bundle loaded and no path to resolve.

Without being sent anywhere, the orchestrator arms a `Monitor` on the progress
file for any subagent expected to run past 5 minutes (any non-trivial bead,
given full test runs) so the user sees per-phase events live — armed
`persistent`, stopped with `TaskStop` when the wave drains — and, before the
first dispatch, arms the durable out-of-session trigger
`references/progress-streaming.md`'s "Back the Monitor with a durable trigger"
specifies, cancelling it in Phase 4.

### Subagent prompt templates

The dispatch prompts — per-bead (sequential/worktree), doc/chore
fast lane, shared-surface batch, plus the orchestrator's worktree
merge notes — are in `references/subagent-prompts.md`. **Read that file at
dispatch** and fill in every `{PLACEHOLDER}` before sending. Do not
paraphrase a template from memory: the per-bead ceremony is ordered, and
the ordering is what makes the crash-recovery path mechanical.

## Reference files

Load the one that matches where you are in the run:

| Working on | Load |
|------------|------|
| Dispatching any subagent (Phase 1.5 grooming, Phase 3 execution) | `references/subagent-prompts.md` |
| Partitioning the queue into waves and buckets, and writing the Phase 2.5 dispatch plan | the grooming subagent invokes the `plan-sprint` skill |
| No `Agent` tool in this session — the degraded, nothing-delegated run | `references/main-loop-fallback.md` |
| A subagent running past 5 minutes; arming the Monitor and the durable trigger; reading progress events | `references/progress-streaming.md` |
| A subagent that crashed or timed out mid-ceremony | `references/recovery.md` |
| Resuming a run whose state you do not hold (the orchestrator died or compacted) | invoke the `recover-sprint` skill |
| A project with no `## bead-bonanza gates` or `## model lineup` block | `references/project-setup.md` |
| Resuming a bead interrupted mid-implementation (its worktree still exists) | `references/operating-principles.md` |
| Phase 0 pre-flight: probe the environment, resolve `RUN_DIR`/`SCRIPTS` | run `scripts/pre-flight-check.sh` |
| Emitting a progress event | run `scripts/bb-progress.sh` |
| Checking a `skipped` claim against the tree | run `scripts/bb-verify-skip.sh` |
| Closing a bead: close + verify + stage + commit + push the tracker | run `scripts/bb-close-bead.sh` |
| Snapshotting the run directory after each bead's acceptance check | run `scripts/bb-backup-run.sh` |
| The Definition of Ready the Phase 1.5 gate enforces | `references/shared/work-order-definition-of-ready.md` |

The reference files are one hop from here — read them in full rather
than previewing.

The `scripts/` entries are **executed, not read**. Each encodes an ordering,
quoting or portability rule that is easy to state and easy to get wrong under
pressure. Run them rather than reproducing what they do.

Because they are not to be read, their interfaces are published here — an
agent that must open a script to learn how to call it has already been sent
to read it:

```
pre-flight-check.sh <run-id> [--gate-baseline] [--self-test]   # sandbox ON
bb-progress.sh   <run-dir> <bead-id> <phase> [key=value ...]
bb-verify-skip.sh [pathspec ...]        # no pathspec = whole working tree
bb-close-bead.sh [--run-dir <path>] [--trailers-file <path>] <bead-id> <priority> <title> [path ...]
bb-backup-run.sh <run-dir> [<dest>]     # <dest> defaults to <repo>/.bonanza/<run-id>
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
  `--trailers-file <path>` is optional and must precede `<bead-id>`; its
  contents are appended to the commit body after a blank line, so a project's
  required trailers (`Co-Authored-By`, a session URL) land in the one commit
  this script makes and no caller needs `git commit --amend` to add them.
  Omitting it appends no trailers.
  `<path ...>` are the code files this bead changed; WHEN omitted it stages
  everything currently modified, so pass explicit paths whenever another
  agent may be sharing the tree. It closes with a status-only `bd update` and
  never writes the description. It then **verifies the close against the
  database** — `bd show <id>` must read `closed` — and refuses to stage or
  commit anything when it does not, so a `bd update` that silently did not
  take cannot produce a commit claiming a closure that never happened. After
  the commit it runs `bd dolt push`; a push failure is a WARNING, not an
  error — the bead IS closed and committed, and the repair is another push,
  never a second `bd update`.
- **`scripts/bb-backup-run.sh`** — commits a self-describing copy of the live
  run directory to `.bonanza/<run-id>/` on the run's branch, so `dispatch.md`
  and the per-bead progress and handoff files survive a container reclaim.
  Run it per bead, after each acceptance check, from the orchestrator's own
  checkout — it refuses to commit to `main`/`master`, and a snapshot taken
  inside a bead's worktree lands on a branch that is deleted after merge.
  `<run-dir>` may be the `current` symlink; it is resolved to its physical
  path so the snapshot is named for the run. Idempotent: each invocation
  replaces the copy wholesale, and prints `unchanged` without committing when
  no run file moved. Exit 0 committed (short sha on stdout) or already
  current (`unchanged`), 1 a step failed, 2 bad invocation. **The live run
  directory is authoritative and this copy never is** — read the snapshot only
  WHEN `$RUN_DIR` is gone, and then state its as-of time in whatever you
  conclude from it.

## Phase 0 — Pre-flight

One script measures the environment; you act on what it returns. Run it from
this skill's own directory, with a run id of `$(date +%s)-<queue size>`:

```bash
eval "$(<this skill>/scripts/pre-flight-check.sh "<run-id>" --gate-baseline)"
```

Eval the stdout — it sets `RUN_DIR`, `SCRIPTS` and the `BB_*` facts this
document refers to elsewhere. Read the stderr report once. Exit **0**
proceed; **1** blocking, named in the report and fixed before Phase 1 — a
missing `AGENTS.md` block means read `references/project-setup.md` and draft
it WITH the user, never invent gates or a lineup from memory; **2** you
called it wrong.

Those facts are ledger lines with a command behind them. **Do not re-assert
from memory what the script just measured** — push mechanics in particular
vary by container. WHEN `BB_PUSH_NEEDS_REFSPEC=yes`, push with an explicit
refspec (`git push <remote> <branch>`) every time, and never let a pipe to
`tail` mask a push's exit status. Verifying that a push **landed** is not
conditional on that fact and is not `git status -sb` — it is the end-of-wave
check at Phase 3 step 7, which every run performs.

**Commit trailers.** WHEN `CLAUDE.md`/`AGENTS.md` states a commit-trailer
convention (`Co-Authored-By`, a session URL, etc.), resolve it once here and
write it to `$RUN_DIR/trailers.txt` — a session URL varies per session, so
this file is per-run and never committed. Substitute that path as
`{TRAILERS_FILE}` into every dispatch prompt; the close ceremony
(`bb-close-bead.sh --trailers-file {TRAILERS_FILE} ...`) appends its contents
to the commit body, so no subagent needs `git commit --amend` to add them.
WHEN the project states no trailer convention, leave `{TRAILERS_FILE}` unset
and omit the flag; the close ceremony then appends no trailers.

What no script can settle:

- **Draft PR.** WHEN `BB_PR_CI=yes`, ask the user once whether to open a draft
  PR from the working branch — every wave push then gets the full CI matrix in
  the cloud, including toolchain-heavy suites the local sandbox can't run,
  which is what makes the slim local gate policy (below) safe. (`gh` fails TLS
  verification inside the sandbox; run it with sandbox disabled.) WHEN
  `BB_PR_CI=no`, don't ask.
- **Harness worktree directory.** Confirm it is gitignored: create the first
  worktree, then check `git status --porcelain` is clean. Checking only before
  that worktree exists tests nothing — the harness may place it inside the
  repo, and untracked worktrees there read as uncommitted work to anything
  that inspects `git status`.
- **Agent-tool availability.** Whether you have an Agent/Task tool at all is a
  harness property, invisible from a shell. No script can check it — you must.
  WHEN you have none, nothing can be delegated (not grooming, not planning, not
  the beads): read `references/main-loop-fallback.md` before Phase 3.

**`BB_DOLT_WRITE` is a blocking fact where it reads `no`.** It means the
project syncs beads to a Dolt remote and this container cannot write it, so
every closure the run makes dies with the container — the exact loss the close
ceremony exists to prevent. Stop and fix the environment; do not run and hope.
`none` means no Dolt remote is configured, which is not a fault (the project
may track beads another way) and `yes` is the normal case. Pre-flight measures
it even where a project's own session hook also does; a green hook is not a
substitute for this fact.

`BB_WORKTREE_BASEREF` is the same kind of blocking fact — pre-flight now exits
non-zero when it is not `head`. See "Every bead runs in its own worktree"
above.

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

The Run checklist copied into your response (above) is the progress-visibility
mechanism, and it is sufficient. `bd` is the only tracker: a second progress
store is a correctness hazard, not redundancy.

## Phase 1.5 — Readiness gate (grooming)

bead-bonanza dispatches **only groomed beads.** Every later phase already
assumes grooming's outputs: Phase 1's `bd ready` ordering comes from the
dependency graph grooming wires; this gate's own cheap check reads the
`decision-required` label grooming sets (DoR check 9); Phase 2.5 reads the
`tier:` assignment grooming makes. An ungroomed queue makes every one of them
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
Definition of Ready's checks hold — read them from
`references/shared/work-order-definition-of-ready.md`, not from memory — plus
one signal it does not carry: `bd lint` is clean for the bead, WHERE the
installed `bd` has it. Parallelize; this is read-only. A bead failing check 9
(an open decision) is bounced, not questioned mid-run.

A queue that is entirely ready makes this phase a one-line no-op for *grooming*:
say so and proceed to Phase 2.5, which still dispatches one subagent — for
planning alone.

### 2. Groom what fails (delegate — don't groom in the orchestrator)

If any bead fails the check, do NOT dispatch it and do NOT fix it inline —
grooming reads source and wires graphs, which belongs in a subagent (see "don't
read source files in the orchestrator"). Dispatch a **grooming subagent** over
the ungroomed subset (template below); it runs `groom-beads` to the Definition
of Ready and returns a per-bead readiness report plus a **wave preview** — its
candidate partition under dependencies and C4 collisions. **Its final act is to
invoke `plan-sprint`**, which turns that preview into the run's wave table:
it has already read every bead, so a second agent would only re-derive the same
partition and leave you two to reconcile (Phase 2.5).

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
it. These shapes qualify:

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
  terms.) Grooming and wiring the edge are tracker writes, which need not wait
  for a wave boundary (`references/operating-principles.md`).
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

**Planning is delegated; the manifest is not.** Classifying every queued bead,
computing C4 overlap and packing buckets is read-heavy analytic work across the
whole queue — it belongs in a subagent ("Don't read source files in the
orchestrator"), and Phase 1.5's grooming subagent has already read every bead.
So that subagent invokes the **`plan-sprint`** skill as its final act, rather
than a second agent re-partitioning the same queue, and returns you the wave
table plus one line per override of its own wave preview. **One subagent, never
two** — WHERE the queue needed no grooming at all, dispatch one for planning
alone.

Three things stay yours:

1. **The run manifest.** Write
   `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/dispatch.md`'s header — the
   pre-flight stdout block, gate profile, gate baseline, model lineup, the
   run's effort setting, standing authorization and environmental ledger —
   **before** you dispatch grooming. WHERE the harness exposes no effort
   setting, write `unrecorded` — never omit the field and never guess a
   value; an omitted field is indistinguishable from a forgotten one, and a
   guessed value is worse than either.
   Recovery reads `dispatch.md` instead of re-deriving the run, so the run's
   most load-bearing artifact is not authored one delegation hop away.
   `plan-sprint` appends the wave table under your header and leaves it alone.
2. **The enumerated bead list.** Pass `plan-sprint` explicit bead ids, never
   the filter that produced them. The filter produced the queue; the plan IS
   the queue, and a filter re-evaluated later returns a different set than the
   one this run committed to and admits beads nobody scoped.
3. **Amendments.** WHEN the run diverges from the plan mid-flight — a wave
   resequenced, a bucket reclassified, a bead admitted or bounced, a tier
   escalated — append a dated entry under `## Amendments` in `dispatch.md`
   **and** update the wave table in place, yourself; never re-invoke
   `plan-sprint` to re-partition a run in flight. The plan on disk SHALL
   describe the run as it is actually being executed; a plan that silently
   stops describing the run is worse than no plan, because recovery trusts it.

**`dispatch.md` is written on every run** — there is no queue-size or
bucket-count exemption. Chat is the one store that does not survive.

Print the returned wave table to chat. Don't ask for approval — just print and
proceed. The user can interrupt if they disagree.

WHEN no `Agent` tool is available, nothing can be delegated, planning included:
read `references/main-loop-fallback.md` before Phase 3.

### Concurrent dispatch

**Concurrent dispatch is ENABLED.** Several beads may be dispatched at once
within a wave, each in its own worktree, merged as each returns.

Note what is and is not concurrent. **A wave is the unit of parallelism**, and
waves remain sequential — a wave boundary is where dependency edges resolve and
where the base ref advances. Concurrency is *within* a wave, across beads that
the dispatch plan established are independent.

Entry conditions, both required and both about **safety**: **no shared
`c4_changes` component** between any two beads in the wave, and the base-ref
check re-verified for this wave. Both are yours to check at dispatch, whatever
the plan says.

**There is no minimum bead count, and do not reintroduce one** — a count in a
list of safety conditions reads as one, and nothing about two concurrent beads
is less safe than three. Whether a wave is *worth* dispatching concurrently is
ordinary judgment; whether it is *safe* is the entry conditions above.

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
  Split the bucket WHEN more than one bead would be dragged above its
  tier, OR when any bead would be dragged more than one tier. Weigh that
  against what batching saves: batching exists to avoid repeated full
  `<test>` runs, so WHERE Phase 0's gate baseline (`scripts/pre-flight-check.sh`)
  shows the suite completing in seconds, the saving is negligible and the
  correct tier wins; WHERE it takes minutes, one bead one tier high is
  the cheaper error.
- **Escalate, never downgrade** — if a bead fails gates/review twice or
  its subagent visibly flails, re-dispatch that bead one tier up. Never
  re-dispatch below the assigned tier, and never "save cost" by
  downgrading mid-run.
- **Record every escalation** — the bead, the assigned tier, the tier
  that completed it, and what failed. This is a required line in the
  Phase 4 report (see "Tier escalations"), and it is the run's only
  direct evidence about whether grooming's routing is calibrated. WHEN
  `CLAUDE_CODE_SUBAGENT_MODEL` is set, report it as Phase 4's "Tier
  escalations" section directs.
- **Record the routing** — the dispatch plan's wave table carries a
  `Tier` column for every bead.

### Push policy

- **`push-per-bead`** — subagent pushes after each commit. Best for sequential
  mode (recovery is easier on a per-bead failure).
- **`defer-push`** — subagent commits per bead, orchestrator pushes once at
  bucket end. Default for fast-lane and shared-surface batch (one process;
  failures are observable).

For a single-bead "wave," `push-per-bead` and `defer-push` are equivalent —
don't bother labeling. Just push when the subagent returns.

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

   **`CLAUDE_CODE_SUBAGENT_MODEL` override and no routing are different
   results.** WHEN `CLAUDE_CODE_SUBAGENT_MODEL` env var is set, every
   subagent runs at that forced model regardless of tier, so no routing
   occurs and tier escalations mean nothing — omit the table and write:

   ```
   Tier escalations: N/A — CLAUDE_CODE_SUBAGENT_MODEL is set, so every
   bead ran at that model and no tier routing occurred.
   ```

   **`none` and "no routing occurred" are different results, and this
   table cannot tell them apart.** WHEN the run used the main-loop
   fallback because delegation was unavailable, do NOT write `none`.
   Write, in place of the table:

   ```
   Tier escalations: N/A — no routing occurred. No Agent tool was
   available, so every bead ran in the orchestrator's own loop at the
   session model, and no bead's tier was honored.
   ```

   **Say this alongside the table**, because the reader will otherwise
   draw the wrong conclusion from a short one: escalations reveal
   *under*-routing only. A bead groomed `frontier` that a `light` model
   would have finished never appears here — it succeeds, and the
   overspend is silent. So zero escalations across a large queue is not
   evidence of good calibration; it is equally consistent with routing
   everything too high (see the over-routing table below). WHEN a run of
   ≥10 beads produces zero escalations, say so and suggest the next run
   try one tier lower on the beads whose `tier_reason` was thinnest.
3b. **Deliberate over-routes** — every bead you knowingly ran a tier above
    its assigned one (e.g. batched into a higher-tier bucket per the
    tier-batching rule), one row each:

   ```
   | Bead | Assigned | Ran at | Why raised |
   |------|----------|--------|------------|
   | demo-9h0 | light | standard | batched with a standard bead; gate suite completes in seconds |
   ```

   State `none` explicitly when there were no over-routes — the null is
   the result, not the absence of one. This table and the escalation table
   above are together the run's complete routing evidence, and
   `groom-beads`' retro loop needs both to recalibrate the rubric.
4. **Follow-up beads** — beads created mid-run via `bd create`, split into those
   admitted into the run (in-scope prerequisites / AC-decompositions / coverage
   gaps, with their outcome) and those left for the user (out-of-scope finds,
   filed with a `discovered-from` link, awaiting triage).
5. **Stashes** — if any were created during recovery, list with `git stash
   list`.
5a. **Sprint retro** — synthesize the run's own evidence into how the next
    sprint runs better. Specified in "The sprint retro" below. It runs on
    every run that closed at least one bead, and it runs *here*: after the
    report, before 6b deletes the run snapshot it reads.
6. **Pushes** — confirm all closed beads' commits are on the remote.
6a. **Cancel the durable out-of-session trigger** armed at first dispatch
    (`references/progress-streaming.md`, "Back the Monitor with a durable
    trigger") —
    a trigger left armed after the run drains wakes an agent into a finished
    sprint, which trains the reader to ignore it. **State its disposition in
    the report**: `Durable trigger: cancelled (<id>)` or `Durable trigger:
    none armed` — both branches are stateable, and an omitted line is
    indistinguishable from a forgotten cancellation.
6b. **Delete the run snapshot** — `git rm -r .bonanza/<run-id>/` as its own
    commit, **after** the report above is written and committed, never before.
    It is a working artifact of the run and must not reach the PR that carries
    the run's actual work. Nothing is lost: the commits that added it stay in
    the branch history, found with
    `git log --diff-filter=A -- '.bonanza/<run-id>/'` and read back with
    `git show <sha>:<path>`.
7. Path to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/` so the user can
   review the dispatch plan and progress files.

Do NOT create a PR or merge to main — that is a separate, explicit user action.

### The sprint retro

A run is a sprint; this is its retrospective. It runs on **every run that
closed at least one bead**, not only on a run that struggled.

**Read the run's records; do not re-derive the run.** The inputs are the
harvested `retro:` comments, the tables you just printed (3a escalations, 3b
over-routes, 4 follow-up and admitted/bounced beads), `dispatch.md`'s
`## Amendments` log, the per-bead `<bead-id>.handoff.md` and
`<bead-id>.progress.jsonl` under `$RUN_DIR`, and `git log` on the sprint
branch. Do NOT re-read diffs, re-run gates, or re-open the beads' work. A
retro that costs a second pass over the run is a retro that gets skipped.

**Harvest first, then assert.** Findings are `retro:` comments on the beads —
there is no per-bead retro file. Harvest with the `bd export | jq` pipeline in
`references/operating-principles.md` ("Findings harvest"), once filtered to
closed and once unfiltered. A zero count across a drained run is a broken
query, not a clean run: **stop, re-run it**, check the label spelling and
confirm the beads really are closed, before aggregating anything.

**Answer every question below. A question with no evidence gets `null result`
in writing** — omit it and an unasked question is indistinguishable from an
answered one.

1. **Grooming** — which beads were under-groomed, and how did it surface
   (bounced at the readiness gate, an AC unsatisfiable as written, a
   description that failed the fresh-agent test in practice, a wrong or
   missing dependency edge)? Which tier predictions were wrong, in which
   direction? *From:* the bounce list, 3a, 3b, `retro:` comments.
2. **Specs** — which spec documents did an implementer find wrong, thin or
   self-contradictory? *From:* `retro:` comments naming a document.
3. **Plan** — what did `## Amendments` record (resequencings,
   reclassifications, admissions, bounces), and what would have produced a
   better initial bucketing? *From:* the amendments log read against the wave
   table it amended.
4. **Skill** — where did `bead-bonanza` itself mislead, and what did a
   subagent improvise around? *From:* `retro:` comments citing this skill,
   its references or its scripts.

**Write one sprint report per run** at
`rfcs/sprints/<YYYY-MM-DD>-<epic-slug>.md`, dated and slugged rather than
named for the run id. It carries the answers and **every harvested finding
verbatim, including the tail that
warrants no RFC**, and it says so: `triage-rfc` defers on recurrence, so a
finding recorded once and filtered out the next time can never be observed
appearing twice. Filtering the report down to RFC-worthy findings breaks that
accumulation silently.

**Number findings per bead, not globally: `<bead-id>#<ordinal>`** (e.g.
`8ef.1#3`), the ordinal given by position within that finding's own
`### agent-plugins-<id>` heading. Cite findings by this form everywhere the
report's analysis references one. A per-bead id is stable the moment a
finding is harvested; a global sequence number does not exist until the
whole report is assembled, so analysis written before assembly — which is
how this report gets written — would be citing a number that isn't there
yet.

**The bar for anything past the report: it changes how the next run is
dispatched.** "Wave 2 needed three cycles" is a summary. "The `schema`-labelled
beads were bucketed shared-surface but each regenerated the same fixture, so
the batch serialized anyway — bucket that label sequential" is a finding. Only
findings that clear that bar become RFCs.

**Route them, as offers.** Grooming and tier-calibration findings →
`groom-beads`; spec findings → `retro-spec`; skill findings → `retro-skill`.
Each route is one offer the user accepts or declines, naming the findings it
would carry. RFCs land in `rfcs/inbox/<skill>/` and **cite the sprint report
instead of re-inlining its evidence**, so report and RFC cannot disagree: the
report is the testimony, the RFC is the proposal built on it. Every routed RFC
SHALL carry the run's model lineup and effort setting from `dispatch.md`'s
manifest header into its Provenance — the manifest records them precisely so
this step can, and an RFC without them ships with an unfillable provenance
field.

The report itself is not an offer. Write it, append one row for the sprint to
`rfcs/retro-log.md` so one index covers all retro activity, and commit both
with any RFCs in a single commit before 6b's deletion commit.

## Operating principles

- **Groom before you dispatch.** bead-bonanza executes *groomed* beads; it never
  fixes an under-specified bead inline. Phase 1.5 is that gate, and it holds for
  beads pulled into scope mid-run too.
- **Bias toward batching, not sequential.** Find the largest safe batch —
  surface-overlap is the dominant signal, not bead complexity. Sequential is
  reserved for genuine surface divergence or single-bead queues, not "I'm being
  cautious."
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
- **One commit per bead, always — no exceptions.** An interrupted bead's work
  stays uncommitted in its own worktree, so there is no `WIP` commit to make
  and nothing to squash. Full rule and resume path:
  `references/operating-principles.md`.
- **The dispatch plan has no lean path.** It is written every run, whatever
  its size.
- **Surface implicit decisions, don't bury them.** A bead with an open
  decision is bounced at Phase 1.5. A choice a subagent still makes silently
  — one the bead's Design didn't pin — belongs in its `retro:` comment, not
  the diff.
- **Visible progress beats opacity, and the fast path alone doesn't survive an
  interrupt.** Subagents write JSONL progress events; the orchestrator arms
  both wake mechanisms over them ("Recovery and progress streaming" above).
  Long opaque runs erode trust; a run whose only wake mechanism is in-session
  cannot survive an interrupt that lands on the orchestrator.
- **Subagents discard context on return.** This is why the skill prefers fresh
  subagents over `/compact`. Batch subagents (fast lane, shared-surface)
  compromise — they hold context for N beads but free it on return.
- **Don't read source files in the orchestrator.** Code reads happen inside
  subagents — grooming (Phase 1.5) reads source to close decisions and wire
  dependencies before dispatch.
- **Capture decisions in the bead, not just in chat.** The bead description
  (selected options and alternatives, set by grooming) is the durable record;
  the dispatch plan is session-scratch.
- **Fail loudly, recover quietly.** If a subagent blocks, log it and move on;
  the circuit breaker above is what stops the run. If a subagent crashes
  mid-step, follow the recovery procedure (the handoff file is the anchor).
- **Route by tier, and report every escalation.** A bead's `tier:` label is
  grooming's *prediction* of the capability the work needs; resolve it to a
  model via the project's `## model lineup` and dispatch there (Phase 2.5 → Tier
  selection). Assign nothing below the tier; escalate one tier on repeated
  failure; never downgrade to save cost mid-run. Every escalation is a required
  row in the Phase 4 report — it is the only routing evidence a run produces.
  Beads marked "do not delegate" run in the orchestrator's own loop.
- **Project rules win over skill defaults.** The gate profile is the source of
  truth for build/test/lint/format commands and any custom gates the project
  adds. Progress visibility defaults to the run checklist (above, Phase 0); task
  tools are the exception for projects with no durable tracker of their own —
  read the project's `CLAUDE.md`/`AGENTS.md` once at session start. If the
  project has its own commit message conventions, follow them. The skill is
  opinionated but not authoritative.
- **Auto-chain waves under standing authorization.** When dependencies mean
  closing one wave unblocks the next, ask ONCE (at run start or first chain
  point) for standing approval: "re-resolve the queue after each push and
  dispatch the next wave until dry, implicit decisions defaulted, explicit
  `decision-required` beads still stop." With it granted, waves flow without
  per-wave prompts: emit the wave's report and **dispatch the next wave in
  the same turn**. The report is a progress notice, not a checkpoint — under
  standing authorization the turn ends only when the queue is dry, a
  `decision-required` bead is reached, or the circuit breaker trips (Phase 3:
  3 consecutive beads `blocked` or failed). Ending a turn with a non-empty
  queue and none of those causes is a defect, not a pause: nobody is waiting
  to restart it. Newly-filed beads that enter the queue mid-run (subagents
  filing adjacent issues) are admitted only by the mid-run scope test
  (Phase 1.5 step 4), never merely because they match the filter.
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
- **Guard against foreign dirt; tracker writes need no deferring.** Multiple
  agents can share the tree, so name known foreign files as untouchable in
  batch prompts and commit your own work by explicit pathspec
  (`git commit -- <files>`). Tracker writes are exempt: groom a discovered
  bead, bake a decision, or wire a dependency edge at any time, including
  mid-wave. Why, and the claim-reset rule:
  `references/operating-principles.md`.
- **Wire findings capture into every dispatch prompt unconditionally — no
  project file gates it.** Full wiring and the harvest pipeline:
  `references/operating-principles.md`.

## Side effects — what each script leaves behind

These are the artifacts each script creates, writes, or repoints.

- **`scripts/pre-flight-check.sh`** — creates `$RUN_DIR` if it does not
  exist, repoints `$RUN_ROOT/current` symlink at the start of every run,
  emits `BB_*` facts (`BB_LINT`, `BB_MTIME`, etc.) on stdout. It also: runs
  `git worktree prune` at startup; **pushes to the remote** — probes tracker
  write access with `git push <remote> HEAD:refs/dolt/preflight-probe-$$`,
  then deletes that ref, both against the same remote the bead's own work
  will later push to; and, under `--gate-baseline`, **executes arbitrary
  commands** by running `bash -c` on every command parsed out of AGENTS.md's
  `## bead-bonanza gates` block. It does not write user-global state — where
  Dolt telemetry suppression or the `eventsapi.dolthub.com` deny entry is
  absent, it reports that and prints the command to run yourself.
- **`scripts/bb-progress.sh`** — creates `<run-dir>` if it does not exist,
  appends one JSONL event to `<run-dir>/<bead-id>.progress.jsonl`.
- **`scripts/bb-verify-skip.sh`** — no side effects; exits 0 (skip
  verified), 1 (not skipped), or 2 (invocation error).
- **`scripts/bb-close-bead.sh`** — runs `bd update --status closed`, stages
  the specified paths (or, with none given, everything modified **except
  `.beads/*.jsonl`**), makes one commit, then runs `bd dolt push` when a Dolt remote
  is configured. It does **not** write or stage `.beads/issues.jsonl`: the
  tracker's transport is `origin refs/dolt/data`, the JSONL exports are
  untracked, and a close commit carries code only.
- **`scripts/bb-backup-run.sh`** — creates and replaces
  `<repo>/.bonanza/<run-id>/` (or `<dest>`), writes a
  `.bonanza/<run-id>/SNAPSHOT.md` inside it, and makes one commit of that path
  alone. It stages by pathspec, so it cannot swallow unrelated staged work in
  the orchestrator's checkout. It does not push, and it deletes nothing —
  Phase 4 removes the directory.

Each script's own header carries the same usage line. WHEN a usage line
here disagrees with the script, the script is authoritative — and that
disagreement is a finding worth a retro.

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
the repo-level `scripts/sync-shared-references.sh` so this skill is
self-contained. Read them normally; never edit them — fix the source named
in each file's header and re-run that script.

## Retro this skill

Question 4 of Phase 4's sprint retro ("The sprint retro") owns this. It asks
where `bead-bonanza` misled and what a subagent improvised around, answers it
from the run's harvested `retro:` comments, and offers `retro-skill` when
those comments carry a finding. Do not gate a retro of this skill on whether
the run felt like it struggled — the comments are the evidence, and a run that
felt clean can still have produced several.
