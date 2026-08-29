# Dispatching — send a batch out, verify what comes back

**The orchestrator's own procedure.** No subagent reads it: a batch subagent is
told three values and runs `run-batch`. Read this when you are about to
dispatch, and again when a batch returns.

## Contents

- Dispatch continuously — the ready front is re-queried, never walked from a
  stored list
- The prompt — the three values it carries, and why it carries nothing else
- Ceremony dispatches — the two writes that give a phase an identity
- Verify against state, never against testimony — skip claims, the acceptance
  check, and what a standing worktree actually means
- The push, and proving it landed
- The circuit breaker
- What an agent is holding, and when it is done — as a query

## Dispatch continuously

**Do not freeze a cohort and wait for it.** `bd ready --parent <epic>
--exclude-type=epic --exclude-label human` already returns the complete set of
beads nothing is blocking, so a freeze-dispatch-wait-requery cycle adds only
the wait: the slowest bead in the cohort holds up every bead that became ready
behind it. Re-query the front, dispatch what is in it and not already held, and
re-query again as each batch returns.

**`--exclude-label human` is on the DISPATCH query and nowhere else.** A
`human`-labelled bead is the deferred-review channel — a question the run put
to a human, parented to a late ceremony phase so it surfaces at the end. The
label does **not** hold it out of `bd ready` (measured; only the parent phase
does), so once that phase opens it is ordinary-looking ready work and an
agent gets handed a question meant for a person. Do not copy the flag onto a
completion or queue count: a count that hides `human` beads reports a run as
drained while a human still owes it an answer. The channel, and the `bd`
behaviour behind it, are in `run-record.md`.

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
- Diff the front against `manifest.admitted` (`swarm status --run <epic>` names
  the difference) and resolve any creep — admitted as an amendment, or deferred
  out — rather than dispatching an unexamined bead.

## The prompt is three values

**A dispatch prompt carries what the subagent cannot look up, and nothing
else.** Three values, because `run-batch` resolves the rest from the tracker
and from the tree it is standing in:

| | resolved from |
|---|---|
| `{DISPATCH_ID}` | the roster — **derived** from `{EPIC}`, never minted (`dispatch-policy.md`) |
| `{EPIC}` | the run epic (`BB_EPIC`); also the run id |
| `{SCRIPTS}` | absolute path to the coordinator's `scripts/` directory |

Everything a prompt used to carry — the branch, the checks baseline, the check
ids, the trailers, the worktree path, and every bead's id, priority, title and
decision — is one query away for the agent and was a hand-copied second copy
here. **The assignee IS the batch**: `bd ready --assignee {DISPATCH_ID}` is the
queue, in order, and `bd show <id>` is the bead.

**So assign before you dispatch.** That write is now the whole handover; a bead
you meant to include and did not assign is a bead the agent will never see, and
an empty queue is the error it reports.

`prompt-dispatch.md` **IS the prompt** — send it whole, with the three values
filled in. There is no block to extract and nothing to summarise.

**There is no run directory.** The run's record is the epic: the manifest in
its metadata, the amendments in its comments.

## Ceremony dispatches carry an agent id too

A ceremony phase is dispatched the same way a batch is, and **the id is yours
to hand over, not the agent's to invent** — `dispatch-policy.md` has the roster
and the derivation rule, and it is their only copy.

Two writes, both yours, both before you invoke the phase:

- **Assign the phase's epic to its id** — `bd update <ceremony-epic>
  --assignee <role>-<run-id>`. `sprint-scope` does this for every ceremony epic
  as it creates them, so at dispatch time you are confirming it rather than
  doing it; a phase epic assigned to nobody is one `bd ready --assignee` cannot
  find, and the run then has a phase in flight that no query reports.
- **State the id in the prompt**, exactly as `{DISPATCH_ID}` is stated to a
  batch, with the instruction to export it as `BEADS_ACTOR` before its first
  `bd` call.

**Why the export and not just the assignment:** a ceremony agent claims. It
comments on beads and closes what it finished. With
`BEADS_ACTOR` unset those writes land under the harness default — `bd update
--claim` on an unassigned bead silently rewrites the assignee to git
`user.name` and exits 0 — and the phase's own queue stops answering for it.

## Verify against state, never against testimony

**A subagent saying a bead is closed is not evidence that it is.** Every signal
below is read from the tracker or from git. Each command that reads or writes
follows [[frd-swarm-core]]'s command contract (FEAT-SWARM-CORE-001), which
requires it to print the invocation that established each finding. Where a step
and this rule disagree, this rule wins.

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
   least one `retro:` comment was recorded. Each line follows the command
   contract's requirement to print the state and the command that established
   it, so any single check can be re-run without reconstructing it.

   **Read the exit code — it is the decision:** [[frd-swarm-core]]'s
   command contract (FEAT-SWARM-CORE-001) specifies the four exit statuses
   every command uses. For this check, `0` is clear, `1` is repairable
   (signals 1, 2, 5: close the bead, or add the missing `retro:` comment
   from the returned summary; state the repair in the report), and `2` is stop
   (not yours to repair inline — leave the worktree standing, since it is the
   in-flight signal, and surface it).

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
   dispatch before calling it a failure.**
   1. *Removal still pending.* Removal is your own final act of accepting, so
      between a dispatch's return and that act the worktree ALWAYS survives. Do
      the removal, then re-run `git worktree prune && git worktree list`. This
      is the normal state, not a failure.
   2. *A dispatch is still running.* A dispatch holds ONE worktree for every
      bead it carries, so a bead already closed inside it has a surviving
      worktree by construction. Once it returns and that one tree is removed,
      signal 4 becomes true for every bead in it at the same moment.
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

**You push, after each merge; subagents never do.** A push is not complete
until it is verified landed:

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

## What an agent is holding, and when it is done — as a query

**Assignment IS the record.** One string is an agent's whole identity — the
bead's `assignee`, the agent's `$BEADS_ACTOR`, its `{DISPATCH_ID}` — so the
tracker already answers every question a separate bookkeeping record would,
and answers it for a bead nobody remembered to annotate:

```sh
bd ready --assignee <id>                       # what it may start next
bd list --assignee <id> --status in_progress   # what it is working on now
bd list --assignee <id> --status blocked       # what it gave up on
```

**An agent has DRAINED when the first two are empty.** Both are required and
this is measured: `bd ready` excludes `in_progress`, so a bead being worked
right now makes the ready front empty and the ready query ALONE reports a live
agent as finished. `bd ready --json` prints `[]` when the front is empty —
measured on bd 1.2.2 in every configuration tried (empty tracker; issues
present with a filter matching none; every issue closed; `--parent`,
`--assignee` and `--exclude-type` filters), and `null` in none of them. Read
the emptiness test defensively all the same: count only a JSON array and treat
any other payload as empty. A bare `length` answers for whatever it is handed
— on a string it returns the character count — so an error body read back
unguarded reports a queue as long as its own text.

Then read the third: drained with nothing `blocked` is finished; drained with a
`blocked` bead is finished-with-a-block, and the two must not be reported alike.

**Nothing is written at dispatch time to make this work**, which is the point —
there is no key to forget, none to go stale, and a bead added to an agent's
queue mid-run is in the answer immediately.

### A standing worktree

Read it through its branch, not through a stored path. `git worktree list`
prints each registration's checked-out branch, and a dispatch's branch is
**named for the dispatch id** — so the branch names the agent directly, and the
agent's queue says whether it is still working:

| the agent's queue | reading |
|---|---|
| not drained | live agent, working — leave it alone |
| drained | finished; the tree is orphaned, remove it |

An agent that is drained while a bead it holds is still `open` or
`in_progress` died holding it — recover that bead (`recovery.md`).

`{SCRIPTS}/sprint-cleanup.sh` decides the same question and **removes only the
one unambiguous row** — a drained agent, belonging to this run, holding no
blocked bead. It leaves five cases standing and names each: a branch that
names no dispatch, a dispatch id belonging to another run, an agent still
holding a queue, an agent holding a BLOCKED bead (its worktree is the partial
work), and a tracker that could not answer. It never removes on a guess, and
it never force-removes to tidy up.

**A return line is a summary, not a signal.** A subagent's account of what it
did is worth reading and is never what you record: every verdict above comes
from the tracker or from git. Do not parse it.
