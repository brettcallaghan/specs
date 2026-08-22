# Pre-flight

**The orchestrator runs this itself.** It is a script, not a skill: run it,
read it, done. Read this file only when it BLOCKS — it names each blocker and
the remedy, so there is nothing here to weigh in the normal case.

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
conditional on that fact and is not `git status -sb` — it is the end-of-wave
check the orchestrator performs on every wave.

**Commit trailers.** WHEN `CLAUDE.md`/`AGENTS.md` states a commit-trailer
convention (`Co-Authored-By`, a session URL, etc.), resolve it once here and
put it in the manifest's `trailers` key on the run epic — a session URL varies
per session, so it is per-run and never committed. Every dispatch prompt then
passes `--trailers-from {EPIC}`, and the close ceremony fetches the text from
the tracker and appends it to the commit body, so no subagent needs
`git commit --amend` to add them. Omit the flag entirely when the run resolved
no trailers.
WHEN the project states no trailer convention, leave `{TRAILERS_FILE}` unset
and omit the flag; the close ceremony then appends no trailers.

What no script can settle:

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
  routing, concurrency, `plan-sprint` and the grooming subagent. What survives
  is the per-bead ceremony — genuinely valuable, and a different product.

  **Say so and hand over**: the **`run-bead`** skill takes one bead from
  claimed to committed with no orchestrator at all. Working a queue is then
  `bd ready`, then `run-bead` per bead, which is an honest description of what
  a no-Agent session can do.

**`BB_DOLT_WRITE` is a blocking fact where it reads `no`.** It means the
project syncs beads to a Dolt remote and this container cannot write it, so
every closure the run makes dies with the container — the exact loss the close
ceremony exists to prevent. Stop and fix the environment; do not run and hope.
`none` means no Dolt remote is configured, which is not a fault (the project
may track beads another way) and `yes` is the normal case. Pre-flight measures
it even where a project's own session hook also does; a green hook is not a
substitute for this fact.

`BB_WORKTREE_BASEREF` is the same kind of blocking fact — pre-flight exits
non-zero when it is not `head`. See "Every bead runs in its own worktree" in
the skill.
