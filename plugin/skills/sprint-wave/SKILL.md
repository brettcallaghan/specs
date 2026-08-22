---
name: sprint-wave
description: Dispatch one wave of a sprint and verify what comes back — take a
  wave's batches, dispatch a subagent per batch, verify every return against
  the tracker and git rather than against the subagent's own account, and
  return one line per bead. Use when a sprint orchestrator is executing a wave
  from its wave table, or when asked to run a batch of beads, verify a
  subagent's claimed completions, check a skip claim, or decide whether a
  standing worktree means a live agent or a dead one. Merges nothing and
  creates no branch.
version: "0.1"
user-invocable: false
---

# sprint-wave

One wave in, one line per bead out. **You are the orchestrator here** — this
skill is the procedure you follow while dispatching, not an agent you dispatch.
The agents you dispatch run `sprint-batch`.

## Inputs

| | |
|---|---|
| the wave's batches | from the wave table, with each bead's class, tier and push mode |
| `{EPIC}` | the run epic — the manifest, and where incidents land |
| `{SCRIPTS}` | the absolute path to the orchestrator's `scripts/` directory |

The sprint branch already exists and the manifest already carries it. **You do
not create one**, and you never dispatch from the default branch.

## What you return

One line per bead: `completed` / `blocked` / `skipped`, each with a sha. Plus
the wave's push verdict, and the circuit breaker's state if it tripped.

Nothing else. The orchestrator that invoked you merges each branch, and it
reads your lines against state before it does.

## Verify against state, never against testimony

**A subagent saying a bead is closed is not evidence that it is.** Every
acceptance signal below is read from the tracker or from git, and each script
prints the command that established it so any single line can be re-run. This
is the rule the rest of this file implements; where a step and this rule
disagree, this rule wins.

## Dispatch the batches

Execute batches in this order (lowest blast radius first):

1. **Doc/chore fast lane** — single subagent, one worktree at a time.
2. **Shared-surface batches** — one subagent per batch, sequentially across
   batches, one worktree at a time within each.
3. **Plain sequential** — one subagent per bead, in `bd ready` order.

Within a batch, dispatch `isolated` beads concurrently where the planning step's entry
conditions hold; otherwise one at a time. Either way, merge and remove each
bead's worktree as it is accepted — concurrently dispatched beads are merged
one at a time, in the order they return.

After each subagent returns:

1. Append its summary lines to your log (one per bead processed).
2. Mark items on the run checklist complete as you go.
3. **Verify skip claims with a diff-stat.** For every bead the subagent returned
   as `skipped`:
   - Run `{SCRIPTS}/bb-verify-skip.sh` (scoped to that bead's worktree if
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
4. **Acceptance check — every bead, every dispatch mode.** Run it, do not
   perform it:

   ```sh
   {SCRIPTS}/bb-status.sh bead <bead-id> --worktree <path> [--batch]
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
   | `2` | **stop the wave.** Signals 3, 4: the merge or the acceptance never happened. Not yours to repair inline — leave the worktree standing, since it is the in-flight signal, and surface it. |

   Pass `--batch` for a bead processed inside a batch's shared worktree, which
   makes signal 4 skip rather than fail; read signal 4 once, against the batch,
   when it returns.

   Applies **per bead, not per subagent**. A fast-lane subagent that processed
   six beads is checked six times — that is exactly where a silent omission
   hides.

   **A bead failing any signal is reported, not recorded closed.** Return it with the failing signal named (`acceptance-failed: worktree
   survives`) in place of a commit sha, and say so in the run totals rather
   than counting it among the closed.

   **A standing worktree is one of three things — read signal 4 against the
   dispatch mode before calling it a failure.**
   1. *Removal still pending.* Removal is your own final act, so
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
   3. *A real signal-4 failure.* Removal itself failed, or a worktree you
      never created survives. Leave it standing and surface it.


   **A subagent that never returns is checked by these same five signals.** A
   lost completion notification is not evidence of an unfinished bead, and it
   is common enough to plan for. The five signals read state, not testimony,
   so they hold with or without a return — run them unchanged. To decide
   whether a bead is still *working* rather than finished, read `bd show <id>`
   together with whether its worktree still exists: a closed tracker record and
   a removed worktree is conclusive. Not `bd list --status in_progress` — it is
   a query over a moving set and has disagreed with both durable signals.
   **Never re-dispatch on a missing return alone.**
5. **Verify the working tree is clean** with `git status --porcelain`. If dirty,
   fall through to the recovery procedure below.
6. If the batch used `defer-push`, run `git push` for the batch now.

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
   strands every commit behind it. `sprint-recovery` step 3b runs the same
   check from the other side; keep the two in the same form.
   (`{SCRIPTS}/bb-status.sh wave` runs this comparison too.)

   **WHERE a PR is open, OBSERVE its CI rather than assuming it.** "CI is the
   backstop" is otherwise an assertion the run never checks, and a wave that
   slimmed its local gates on that basis has nothing behind it. Open a gate on
   the push and let `bd` watch it:

   ```sh
   bd gate create -t gh:run --blocks <bead-or-epic-id> --await-id <run-id> \
     --reason "CI for the wave push"
   bd gate check          # resolves it when the run finishes; re-run to poll
   ```

   `bd gate discover` finds the run id by branch or commit sha when you do not
   have it. The gated issue leaves `bd ready` until the gate resolves, so the
   wait is expressed as tracker state rather than as a polling loop the
   orchestrator has to remember.

   **Precondition, measured: `bd gate check` shells out to the `gh` CLI.** With
   `gh` absent it reports `error checking - gh CLI not installed` and resolves
   nothing — the gate stays open and blocks its issue indefinitely. `gh` is not
   installed in every environment this skill runs in, and fails TLS inside the
   sandbox where it is. So create gh:run gates only after confirming `gh` runs;
   otherwise keep full local gates and say once that CI is unobserved.
8. Move to the next batch/bead.

## Circuit breaker

**If 3 consecutive beads return `blocked` or fail, stop the wave** and return
saying so. Something is systemically wrong (broken main,
environment issue, an under-groomed bead).

## The dispatch record — what a subagent is and what it holds

A bead has a branch, a worktree, a commit and a comment thread. A **subagent**
had none of that: no id, no record, no way to ask which agent is running and
what it is holding. That gap is why a batch holding one worktree for six beads
needed three paragraphs to disambiguate.

**The record is metadata on the beads, not new beads.** On dispatch:

```sh
bd update <bead> --set-metadata dispatch_id=<id> \
  --set-metadata dispatch_state=running \
  --set-metadata dispatch_model=<model> --set-metadata dispatch_tier=<tier> \
  --set-metadata worktree=<abs path> \
  --set-metadata dispatched_at="$(date -u +%FT%TZ)"
```

On return, the same keys plus `dispatch_state=returned` and
`terminal_state=completed|blocked|skipped`. `--set-metadata` merges, so each
write adds without disturbing grooming's `tier` and `c4_*` keys.

**One `dispatch_id` per subagent, not per bead.** A batch's beads all carry the
same one, which is what makes "what is it holding" answerable:

```sh
bd list --parent <epic> --metadata-field dispatch_state=running   # who is live
bd list --parent <epic> --metadata-field dispatch_id=<id>         # what it holds
```

**This replaces the worktree-disambiguation prose.** A standing worktree no
longer has to be interpreted:

| worktree | `dispatch_state` | reading |
|---|---|---|
| present | `running` | live agent, working — leave it alone |
| present | `returned` | finished; the tree is orphaned, remove it |
| absent | `running` | the agent died. Recover the bead |
| absent | `returned` | done and cleaned up |

**Why not event beads.** D2 proposed two per dispatch. Measured: 100 event
beads took `bd dolt push` from 0.63s to 5.36s, and *deleting them* left it at
1.56s — Dolt is versioned, so the weight is permanent and is charged to every
future clone. A 12-bead wave would be 24 such beads. Metadata was measured
cheap on the same run: nested values round-trip, successive writes merge, and a
100,000-character value was accepted.
