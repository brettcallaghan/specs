# Scripts — interfaces and side effects

**Scoped to what the COORDINATOR runs.** Three of the bundled scripts are its
own — `scripts/pre-flight-check.sh`, `scripts/bb-status.sh`,
`scripts/sprint-cleanup.sh` — plus `scripts/bb-close-bead.sh` on exactly one
path, finishing a bead a crashed subagent left behind (`recovery.md`). The
skip-verification script belongs to the **`sprint-wave`** and **`sprint-batch`**
skills, and the checks runner to **`run-bead`**; each is documented where it is
called.

**The scripts are executed, not read.** Each encodes an ordering, quoting or
portability rule that is easy to state and easy to get wrong under pressure.
Run them rather than reproducing what they do — and read this file rather than
opening one to learn how to call it.

Each script's own header carries the same usage line. WHEN a usage line here
disagrees with the script, **the script is authoritative**, and the
disagreement is a finding worth recording.


```
pre-flight-check.sh [<epic-id>] [--json] [--self-test]          # sandbox ON
bb-status.sh bead <id> [--worktree <path>] [--pre-merge] | wave | run <epic> | harvest
sprint-cleanup.sh <epic-id> [--dry-run]
bb-close-bead.sh [--trailers-from <epic>] <bead-id> <priority> <title>
```
The last one is the recovery path only — see its entry below.

- **`scripts/pre-flight-check.sh`** — the environment probe. Human report on
  stderr, an evalable `SCRIPTS=`/`BB_*=` block on stdout. Exit 0
  proceed / 1 blocking / 2 bad invocation. **Run it with the sandbox ON** — the
  one write the sandbox refuses is the global Dolt telemetry config, which the
  script detects, skips, and reports as the single command to re-run with the
  sandbox off. It always runs the project's declared checks on the untouched
  tree, so a pre-existing failure is recorded as the baseline instead of being
  charged to the first bead — there is no opt-out, because the run where a
  baseline is most expensive to skip is exactly the slow one somebody would
  want to skip it on. `--self-test` exercises the portability wrappers and
  takes no run id; it runs no probe, so it touches neither git nor the network.
  `--json` emits the same facts as one JSON object instead of the eval block —
  an additional mode for an agent that wants structure without eval, never a
  replacement, since a shell cannot eval JSON.
- **`scripts/bb-status.sh`** — read-only; repairs nothing. `bead` verifies the
  five acceptance signals in order; `wave` reports tree cleanliness, the
  landing check (unpushed commits against the upstream), surviving worktrees
  and the merge slot's holder; `harvest` asserts a run's `retro:` total is
  non-zero, since a zero total is a broken pipeline rather than a flawless run.
  Exit 0 clear, 1 repairable, 2 stop-the-wave, 3 invocation error. **Run-level:
  a subagent never runs it** — inside its own worktree signal 4 reads
  differently by construction.
- **`scripts/bb-close-bead.sh`** — **`run-bead` owns this**; the coordinator runs
  it on one path only, finishing a bead a crashed subagent left uncommitted.
  There, from inside that worktree: `--trailers-from <epic>` (optional, before
  `<bead-id>`) reads the manifest's `trailers` key, and passing **no paths**
  stages everything modified, which is correct because everything uncommitted
  in that tree is the bead's. It **verifies the close against the database**
  before staging anything, so a `bd update` that silently did not take cannot
  produce a commit claiming a closure that never happened. A `bd dolt push`
  failure afterwards is a WARNING — the bead IS closed and committed, and the
  repair is another push, never a second `bd update`.

## Side effects — what each script leaves behind


These are the artifacts each script creates, writes, or repoints.

- **`scripts/pre-flight-check.sh`** — **writes the checks baseline into the
  run epic's manifest**, merging rather than replacing so grooming's and
  admission's keys survive; emits `BB_*` facts (`BB_CHECKS_IDS`,
  `BB_MODEL_*`, etc.) on stdout. It creates no run directory. It also: runs
  `git worktree prune` at startup; **pushes to the remote** — probes tracker
  write access with `git push <remote> HEAD:refs/dolt/preflight-probe-$$`,
  then deletes that ref, both against the same remote the bead's own work
  will later push to; and **executes the project's declared checks**, which
  run whatever commands that project declared. It does not write user-global state — where
  Dolt telemetry suppression or the `eventsapi.dolthub.com` deny entry is
  absent, it reports that and prints the command to run yourself.
- **`scripts/bb-close-bead.sh`** — runs `bd update --status closed`, stages
  the specified paths (or, with none given, everything modified **except
  `.beads/*.jsonl`**), makes one commit, then runs `bd dolt push` when a Dolt remote
  is configured. It does **not** write or stage `.beads/issues.jsonl`: the
  tracker's transport is `origin refs/dolt/data`, the JSONL exports are
  untracked, and a close commit carries code only.

- **`scripts/sprint-cleanup.sh`** — removes worktrees whose bead reads
  `dispatch_state=returned`, runs `git worktree prune`, releases the merge
  slot, closes the run's swarm molecule, sets the epic's terminal health, and
  runs `bd dolt push`. **It leaves anything ambiguous standing** — a worktree
  whose bead still reads `running`, or one with no dispatch record — and exits
  1 naming each. `--dry-run` performs none of it and prints what it would do.

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill, vendored in by the repo-level
`scripts/sync-shared-references.sh` so this skill is self-contained. Read them
normally; never edit them — fix the source named in each file's header and
re-run that script.
