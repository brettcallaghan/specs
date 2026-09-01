# Pre-flight

**The orchestrator runs this itself.** It is a script, not a skill: run it,
read it, done. Read this file only when it BLOCKS — it names each blocker and
the remedy, so there is nothing here to weigh in the normal case.

## The readiness gate is NOT here

It belongs to grooming. Readiness is a property of a groomed bead, so asking
before grooming answers a question about beads that are about to change.
`sprint-scope` does not carry it either.

## Running it

One script measures the environment; you act on what it returns. It runs
**twice**, and the split is not ceremony: `sprint-scope` resolves the run epic
by writing to the tracker, and a tracker write must not happen before the
script has shown that a closure can be pushed.

```bash
eval "$(scripts/pre-flight-check.sh)"              # environment
#   ... invoke sprint-scope; it returns BB_EPIC ...
eval "$(scripts/pre-flight-check.sh "$BB_EPIC")"   # + baseline
```

The second pass verifies the epic exists and IS an epic, counts what is ready
in scope, and captures the checks baseline. **The run
id is the epic id** — there is no generated id.

Eval the stdout — it sets `SCRIPTS` and the `BB_*` facts this
document refers to elsewhere. Read the stderr report once. Exit **0**
proceed; **1** blocking, named in the report and fixed before you invoke
`sprint-scope` — a missing declaration means read `project-setup.md` beside
this file and draft it WITH the user, never invent checks or a lineup from memory; **2** you
called it wrong.

Those facts are ledger lines with a command behind them. **Do not re-assert
from memory what the script just measured** — push mechanics in particular
vary by container. WHEN `BB_PUSH_NEEDS_REFSPEC=yes`, push with an explicit
refspec (`git push <remote> <branch>`) every time, and never let a pipe to
`tail` mask a push's exit status. Verifying that a push **landed** is not
conditional on that fact and is not `git status -sb` — it is the landing check
the orchestrator performs after every push.

**Commit trailers.** WHEN `CLAUDE.md`/`AGENTS.md` states a commit-trailer
convention (`Co-Authored-By`, a session URL, etc.), resolve it once here and
put it in the manifest's `trailers` key on the run epic — a session URL varies
per session, so it is per-run and never committed. **WHEN the project states
no convention, write no `trailers` key**: an absent key is what tells a dispatch to omit the
flag, and an empty one is not the same thing.

## What no script can settle

This is why the file exists. Everything below is invisible from a shell, so
the script cannot measure it and cannot warn you about it.

- **Pull requests are the project's business, not bead-bonanza's.** Whether a PR
  is opened, when it opens, and when its CI runs are the project's policy — and
  not every project uses a forge with pull requests at all. Do not ask, do not
  infer it from workflow files, and never treat a PR as cover for running fewer
  checks. **The project's declared checks run after every bead, always.** That
  is what it is for.

  WHERE the project does run CI on a pull request and the run ends with one
  becoming reviewable, the failures that surface then are the run's to fix, the
  same as any other gate failure.

- **Harness worktree directory.** Confirm it is gitignored: create the first
  worktree, then check `git status --porcelain` is clean. Checking only before
  that worktree exists tests nothing — the harness may place it inside the
  repo, and untracked worktrees there read as uncommitted work to anything
  that inspects `git status`.
- **Agent-tool availability — BLOCKING.** Whether you have an Agent/Task tool
  at all is a harness property, invisible from a shell. No script can check it,
  so **you** assert it: with no Agent tool, set `BB_ASSUME_NO_AGENT=1` in the
  pre-flight environment so the fact reaches the manifest like every other, and
  **stop**. Do not start a run.

  This is a blocking condition rather than a degraded mode because subtracting
  delegation leaves nothing bead-bonanza does: no subagents means no context
  isolation, which is its whole point, and with it go worktrees, batching, tier
  routing, concurrency, `sprint-plan` and the grooming subagent. What survives
  is the per-bead ceremony — genuinely valuable, and a different product.

  **Say so and stop.** There is no degraded mode to hand over to.

**Every other blocker is the script's to explain.** `BB_DOLT_WRITE=no` exits
non-zero with the cause and the remedy in the message. Restating them here
would give a reader two texts to reconcile and one of them a chance to go
stale — read the report.
