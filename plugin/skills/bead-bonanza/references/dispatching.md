# Dispatching — send a batch out, verify what comes back

**The orchestrator's own procedure.** No subagent reads it: a batch subagent is
told three values and runs `run-batch`. Read this when you are about to
dispatch, and again when a batch returns.

## Contents

- Dispatch continuously — `swarm ready` is the dispatch set
- The prompt — the three values it carries, and why it carries nothing else
- Ceremony dispatches — the two writes that give a phase an identity
- Verify against state, never against testimony — skip claims, the pre-merge
  check, and accepting the dispatch
- Observing CI on a pull request
- The circuit breaker, and one dispatch's queue — both are `swarm ready`

## Dispatch continuously

**Do not freeze a cohort and wait for it.** Re-query as each dispatch returns,
so a bead that became ready behind a slow one goes out at once rather than
waiting for its neighbours:

```sh
swarm phase close --run <epic>   # first, every time (`supervision.md`)
swarm ready       --run <epic>   # the dispatch set
```

**`swarm ready` IS the dispatch set.** It applies the run's own filter, names
each bead with the dispatch id it is assigned to, and reports a bead whose
dispatch's worktree is still standing as **held** rather than available — so
you cannot send a second agent after one already at work. **Route on the exit
status**, and read what it prints; `swarm ready --help` names the flags and
`--explain` prints the queries it would run without running them.

**One thing it does not measure, and you still do:** diff the ready set
against `manifest.admitted` (`swarm status --run <epic>` names the difference)
and resolve any creep — admitted as an amendment, or deferred out — rather
than dispatching an unexamined bead.

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
below is read from the tracker or from git, and each command prints the
invocation that established each finding, so any single line can be re-run.
Where a step and this rule disagree, this rule wins.

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
3. **The pre-merge acceptance check — per bead, before you accept anything.**

   ```sh
   {SCRIPTS}/bb-status.sh bead <bead-id> --worktree <path> --pre-merge
   ```

   It defers the two signals a merge decides and checks the other three: the
   bead reads closed in the tracker, a commit carrying its id exists, and at
   least one `retro:` comment was recorded. **Read the exit code — it is the
   decision.** `0` is clear, `1` is repairable (close the bead, or add the
   missing `retro:` comment from the returned summary, and state the repair in
   the report), and `2` is stop.

   Applies **per bead, not per subagent**. A batch subagent that processed six
   beads is checked six times — that is exactly where a silent omission hides.

   **No `swarm` command performs this half yet**, which is why it is still
   here. The post-merge half — the same signals with none deferred — belongs
   to the acceptance below and is not yours to run.

   **A bead failing any signal is reported, not recorded closed.** Return it
   with the failing signal named (`acceptance-failed: bead not closed`) in
   place of a commit sha, and say so in the run totals rather than counting it
   among the closed.

   **A subagent that never returns is checked by these same signals.** A lost
   completion notification is not evidence of an unfinished bead, and it is
   common enough to plan for. The signals read state, not testimony, so run
   them unchanged. **Never re-dispatch on a missing return alone.**
4. **Accept the dispatch** — run it, do not perform it:

   ```sh
   swarm dispatch accept --run <epic> --dispatch <dispatch-id>
   ```

   The drain check, the serialized/concurrent case, the merge, the worktree
   removal, the per-bead re-check with no signal deferred and the push are all
   inside it, in the order its guards require. **The command owns them and
   this document does not restate them**: `--help` names the flags,
   `--explain` prints every invocation without running one, and `--no-push`
   stops in front of the push. **Route on the exit status.**
5. **Verify the working tree is clean** — `git status --porcelain`. If dirty,
   fall through to `recovery.md`.

## Observing CI on a pull request

**The push and its landing proof are `swarm dispatch accept`'s** (above), or
`swarm branch push --branch <branch>` on its own. You run neither by hand and
subagents run neither at all.

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

## The circuit breaker, and one dispatch's queue — both are `swarm ready`

**Report every dispatch's return to the breaker as you accept it:**

```sh
swarm ready --run <epic> --completed <n> [--threshold <n>]
```

`--completed 0` counts as a failure and `--completed <n>` clears the count. The
breaker's count lives in the manifest as `consecutive_dispatch_failures`, so a
resumed session does not restart it, and the command reports the run stopped
once the count reaches the threshold — **the threshold is a run-level
authorization row** (`sf.toml`'s `[run] dispatch_failure_threshold`, else 3),
resolved by `sprint-scope` and passed here. Omit `--completed` and the
invocation only reads the count.

**A threshold below 1 is refused, not honoured.** `swarm ready` rejects it in
every mode before reading anything, because 0 would stop the run on every
dispatch return rather than disabling the breaker — unlike `review_budget`,
where 0 does mean off. Nothing turns the breaker off. The default 3 is
currently written in two places, `breakerDefaultThreshold` in the command and
this row's default in `sprint-scope`; they must not drift, and collapsing them
is tracked separately.

**Tripping it stops the RUN, not just the batch.** Whatever it caught — a
broken base, a wrong assumption in the plan, an environment fault — the next
dispatch meets it too.

**One dispatch's own queue and its drain verdict:**

```sh
swarm ready --run <epic> --agent <dispatch-id>
```

It reports the dispatch's ready and in-progress beads in dependency order and
whether it has **drained**, and a dispatch id no bead in the run names is
reported as *not an established dispatch* rather than as a drained one — so a
mistyped id cannot read as completion. **Assignment IS the record**: one string is an agent's
whole identity — the bead's `assignee`, the agent's `$BEADS_ACTOR`, its
`{DISPATCH_ID}` — so nothing is written at dispatch time to make this work.

Read `bd list --assignee <id> --status blocked` beside it: drained with nothing
blocked is finished, drained with a blocked bead is finished-with-a-block, and
the two must not be reported alike. An agent drained while a bead it holds is
still `open` or `in_progress` died holding it — recover that bead
(`recovery.md`).

**A return line is a summary, not a signal.** A subagent's account of what it
did is worth reading and is never what you record: every verdict above comes
from the tracker or from git. Do not parse it.
