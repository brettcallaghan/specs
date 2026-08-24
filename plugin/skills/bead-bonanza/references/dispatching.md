# Dispatching — send a batch out, verify what comes back

**The orchestrator's own procedure.** No subagent reads it: a batch subagent is
told its placeholders and runs `sprint-batch`. Read this when you are about to
dispatch, and again when a batch returns.

## Contents

- Dispatch continuously — the ready front is re-queried, never walked from a
  stored list
- Filling in the prompt — every placeholder, and where each is resolved from
- Verify against state, never against testimony — skip claims, the acceptance
  check, and what a standing worktree actually means
- The push, and proving it landed
- The circuit breaker
- The dispatch record — what an agent is holding, as a query

## Dispatch continuously

**Do not freeze a cohort and wait for it.** `bd ready --parent <epic>
--exclude-type=epic` already returns the complete set of beads nothing is
blocking, so a freeze-dispatch-wait-requery cycle adds only the wait: the
slowest bead in the cohort holds up every bead that became ready behind it.
Re-query the front, dispatch what is in it and not already held, and re-query
again as each batch returns.

**Concurrency policy is in the graph, not in a schedule.** Two beads that
share a `c4_changes` component must not run at once; the plan expresses that as
a `blocks` edge between them, so they cannot both be in the ready front. That
is why nothing here needs to remember which beads may run together — the front
is already safe by construction. Those **collision edges** are policy rather
than real dependencies, and both ends carry `metadata.collision` naming the
shared surface so a reader can tell them apart: `run-record.md`.

**Re-verify at every re-query**, whatever the plan said:

- `bd epic close-eligible` first, scoped by hand — it takes **no** positional
  epic, so the bare command closes every eligible epic in the database.
  Omitting it is a silent deadlock: a finished ceremony epic that never closes
  keeps the beads it gates out of the front forever
  (`supervision.md`).
- `BB_WORKTREE_BASEREF` is `head` before the next worktree is created. It is
  global user state no repository owns, so it can change between dispatches.
- Diff the front against `manifest.admitted` (`bb-status.sh run <epic>` names
  the difference) and resolve any creep — admitted as an amendment, or deferred
  out — rather than dispatching an unexamined bead.

## Fill in the prompt before you send it

The prompt files are the prompt and nothing else — send one whole, with no
block to extract. **You** resolve every `{PLACEHOLDER}` first; a subagent has no
session context to infer one from, and an unsubstituted placeholder is the
single most common dispatch defect.

| resolved | placeholder | from |
|---|---|---|
| once per run | `{EPIC}` | the run epic (emitted as `BB_EPIC`); also the run id |
| | `{BRANCH}` | the sprint branch, from the manifest |
| | `{SCRIPTS}` | absolute path to the orchestrator's `scripts/` directory |
| | `{CHECKS_IDS}` | the ids the project declared (`BB_CHECKS_IDS`) |
| | `{CHECKS_BASELINE_TEXT}` | `bd show {EPIC} --json \| jq -r '.[0].metadata.checks_baseline'` |
| per dispatch | `{DISPATCH_ID}` | the subagent's **agent id**, DERIVED from `{EPIC}` and never minted — one per subagent, shared by every bead it holds. It is also the bead's `assignee` and the subagent's `$BEADS_ACTOR`: one string, which is what makes "what is that agent holding" a query and lets a resumed session recompute it. The roster of agent kinds and the derivation rule are in `dispatch-policy.md`, their only copy — read the id off it, do not invent one |
| | `{WORKTREE}` | the absolute path to the worktree you created for it |
| | `{SURFACE}` | what the batch's beads have in common, or "documentation and chores; no shared code surface" |
| | `{SURFACE_TARGETED_TESTS}` | the tests covering that surface |
| per bead | `{ID}`, `{PRIORITY}`, `{TITLE}`, `{DECISION}`, `{IMPLICIT_NOTES}` | the bead itself; numbered (`{ID_1}`, …) in the batch prompt, one line per bead in processing order |

**There is no run directory.** The run's record is the epic: the manifest in
its metadata, the amendments in its comments.

Batches go out with the batch prompt (`prompt-batch.md`), one
subagent per batch; a bead with no batch peers goes out alone with
`prompt-per-bead.md`. Each gets one worktree
(`worktree-mechanics.md`).

## Verify against state, never against testimony

**A subagent saying a bead is closed is not evidence that it is.** Every signal
below is read from the tracker or from git, and each script prints the command
that established it so any single line can be re-run. Where a step and this
rule disagree, this rule wins.

After a subagent returns:

1. Append its summary lines to your log — one per bead processed.
2. **Verify skip claims with a diff-stat.** For every bead returned `skipped`,
   run `{SCRIPTS}/bb-verify-skip.sh` in that bead's worktree. It exits 0 when
   the skip is consistent, 1 when it contradicts the tree.
   - **Exit 0:** confirm the summary line cites the bead or commit that already
     addressed it (`skipped - already addressed by demo-wxi @ a1b2c3d`). A
     missing citation is pushed back on — re-prompt, or fill it in from
     `bd show <id>` — before the skip is accepted.
   - **Exit 1 with status `skipped`:** an inconsistency. **Block the close.**
     The subagent either did work it did not acknowledge or left the tree
     dirty. Re-dispatch instructing it to pick one: commit the diff and close
     as `completed` under the full ceremony, or revert it and re-confirm
     `skipped` with a fresh CLEAN run.
   - Do not proceed to the merge, the push, or the next batch until it is
     reconciled.
3. **Acceptance check — every bead, every dispatch mode.** Run it, do not
   perform it:

   ```sh
   {SCRIPTS}/bb-status.sh bead <bead-id> --worktree <path> --pre-merge
   ```

   It verifies the five signals in the required order and stops at the first
   that fails: the bead reads closed in the tracker; a commit carrying the id
   exists; that commit is an ancestor of HEAD; the worktree is gone; and at
   least one `retro:` comment was recorded. Each line prints the state **and
   the command that established it**, so any single check can be re-run without
   reconstructing it.

   **Read the exit code — it is the decision:**

   | | |
   |---|---|
   | `0` | all clear — record the bead closed |
   | `1` | **repairable in place.** Signals 1, 2, 5: close the bead, or add the missing `retro:` comment from the returned summary. State the repair in the report. |
   | `2` | **stop.** Not yours to repair inline — leave the worktree standing, since it is the in-flight signal, and surface it. |

   **`--pre-merge` is not optional on this run.** You have not merged yet, so
   signals 3 (merged into HEAD) and 4 (worktree gone) would fail on **every
   bead, by construction**. The flag defers both. **Re-run the check without it
   after merging and removing the worktree** — that second run is what actually
   verifies them, and skipping it leaves nothing checking that either happened.

   Applies **per bead, not per subagent**. A batch subagent that processed six
   beads is checked six times — that is exactly where a silent omission hides.

   **A bead failing any signal is reported, not recorded closed.** Return it
   with the failing signal named (`acceptance-failed: worktree survives`) in
   place of a commit sha, and say so in the run totals rather than counting it
   among the closed.

   **A standing worktree is one of three things — read signal 4 against the
   dispatch mode before calling it a failure.**
   1. *Removal still pending.* Removal is your own final act of accepting, so
      between a batch's return and that act the worktree ALWAYS survives. Do
      the removal, then re-run `git worktree prune && git worktree list`. This
      is the normal state, not a failure.
   2. *A batch is still running.* A batch holds ONE worktree for every bead it
      carries, so a bead already closed inside it has a surviving worktree by
      construction. Once the batch returns and that one tree is removed, signal
      4 becomes true for every bead in it at the same moment.
   3. *A real signal-4 failure.* Removal itself failed, or a worktree you never
      created survives. Leave it standing and surface it.

   **A subagent that never returns is checked by these same five signals.** A
   lost completion notification is not evidence of an unfinished bead, and it
   is common enough to plan for. The signals read state, not testimony, so they
   hold with or without a return — run them unchanged. To decide whether a bead
   is still *working* rather than finished, read `bd show <id>` together with
   whether its worktree still exists: a closed tracker record and a removed
   worktree is conclusive. Not `bd list --status in_progress` — it is a query
   over a moving set and has disagreed with both durable signals. **Never
   re-dispatch on a missing return alone.**
4. **Verify the working tree is clean** — `git status --porcelain`. If dirty,
   fall through to `recovery.md`.

## The push, and proving it landed

Push per the batch's push mode (`dispatch-policy.md`). **A push is
not complete until it is verified landed:**

```sh
git fetch <remote> <branch>
git log --oneline <remote>/<branch>..HEAD    # must be empty
```

A push can be rejected for reasons that leave the tree clean, the worktrees
removed and the beads closed — every other signal reads "finished". This is the
only one that distinguishes a landed push from a rejected one, and it costs one
fetch. A credential refused for `.github/workflows/**` rejects the WHOLE push,
so one workflow file strands every commit behind it. `sprint-recovery` runs the
same check from the other side; keep the two in the same form.
(`{SCRIPTS}/bb-status.sh branch` runs this comparison too.)

**WHERE a PR is open, OBSERVE its CI rather than assuming it.** "CI is the
backstop" is otherwise an assertion the run never checks, and a run that
slimmed its local gates on that basis has nothing behind it. Open a gate on the
push and let `bd` watch it:

```sh
bd gate create -t gh:run --blocks <bead-or-epic-id> --await-id <run-id> \
  --reason "CI for the push"
bd gate check          # resolves it when the run finishes; re-run to poll
```

`bd gate discover` finds the run id by branch or commit sha when you do not
have it. The gated issue leaves `bd ready` until the gate resolves, so the wait
is expressed as tracker state rather than as a polling loop to remember.

**Precondition, measured: `bd gate check` shells out to the `gh` CLI.** With
`gh` absent it reports `error checking - gh CLI not installed` and resolves
nothing — the gate stays open and blocks its issue indefinitely. `gh` is not
installed in every environment this skill runs in, and fails TLS inside the
sandbox where it is. So create `gh:run` gates only after confirming `gh` runs;
otherwise keep full local gates and say once that CI is unobserved.

## Circuit breaker

**Three consecutive beads returning `blocked` or failing stops the RUN, not
just the batch.** That is systemic — a broken base, a wrong assumption in the
plan, an environment fault — and the next dispatch meets it too. Stopping one
batch alone hands the same failure to a fresh set of agents.

## The dispatch record — what an agent is holding

A bead has a branch, a worktree, a commit and a comment thread. A **subagent**
has none of that on its own: no id, no record, no way to ask which agent is
running and what it is holding.

**The record is metadata on the beads, not new beads.** On dispatch:

```sh
bd update <bead> --set-metadata dispatch_id=<id> \
  --set-metadata dispatch_state=running \
  --set-metadata dispatch_model=<model> --set-metadata dispatch_tier=<tier> \
  --set-metadata worktree=<abs path> \
  --set-metadata dispatched_at="$(date -u +%FT%TZ)"
```

On return, the same keys plus `dispatch_state=returned` and
`terminal_state=completed|blocked|skipped`. `--set-metadata` merges at the top
level, so each write adds without disturbing grooming's `tier` and `c4_*` keys.

**One `dispatch_id` per subagent, not per bead.** A batch's beads all carry the
same one, which is what makes "what is it holding" answerable:

```sh
bd list --parent <epic> --metadata-field dispatch_state=running   # who is live
bd list --parent <epic> --metadata-field dispatch_id=<id>         # what it holds
```

**This is what a standing worktree means**, without prose to interpret:

| worktree | `dispatch_state` | reading |
|---|---|---|
| present | `running` | live agent, working — leave it alone |
| present | `returned` | finished; the tree is orphaned, remove it |
| absent | `running` | the agent died. Recover the bead |
| absent | `returned` | done and cleaned up |

**Why not event beads.** Measured: 100 event beads took `bd dolt push` from
0.63s to 5.36s, and *deleting them* left it at 1.56s — Dolt is versioned, so
the weight is permanent and is charged to every future clone. A 12-bead
dispatch would be 24 such beads. Metadata was measured cheap on the same run:
nested values round-trip, successive writes merge, and a 100,000-character
value was accepted.
