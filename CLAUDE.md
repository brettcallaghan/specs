# Project Instructions for AI Agents

This file provides instructions and context for AI coding agents working on this project.

## Session bootstrap — check this first

**If `bd` or `strata` is missing from PATH, the SessionStart hook did not
run. Run it by hand before doing anything else:**

```bash
command -v bd >/dev/null && command -v strata >/dev/null || \
  CLAUDE_PROJECT_DIR="$(git rev-parse --show-toplevel)" \
  bash "$(git rev-parse --show-toplevel)/.claude/hooks/session-start.sh"
```

It is idempotent, takes a few minutes cold (it builds `bd` from source
and `strata` from a checkout), and is a no-op once both are installed.

**Why this is needed.** `.claude/settings.json` registers that hook, and
Claude Code loads hooks only from the **project root** — there is no
mechanism for a multi-root workspace to register hooks per repository
([hooks reference](https://code.claude.com/docs/en/hooks)). When a
session attaches this repository *alongside another*, both are cloned as
siblings under a shared parent, that parent becomes the project root,
and it contains no `.claude/` — so this repository's hook is never
registered. Nothing in the hook or the settings file is wrong, and
merging them to the default branch does not help; they are simply not
read. Observed 2026-08-01 in a session with `agent-plugins` and `strata`
both attached: no `bd` and no `strata`. The **skills are no longer
exposed to this**: they are committed symlinks under `.claude/skills/` →
`plugin/skills/<name>`, so they load from the working tree whether or not
the hook runs (and a global `~/.claude/skills` copy cannot shadow them).
Only `bd` and `strata` still depend on the hook or the setup script.

CLAUDE.md *is* still read from a nested repository, which is why this
instruction works where the hook does not.

**The durable fix is an environment setup script**, which runs before
Claude Code launches and does not depend on the project root. It is
configured per cloud environment at claude.ai/code rather than in this
repository — see `docs/session-bootstrap.md` for the script to paste and
its trade-offs.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:6cd5cc61 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

## How beads persist in THIS project

The managed block above states the upstream model — that sync uses
`refs/dolt/data` on your git remote. **That is not how this project works,
and following it will lose issues.** Read this section instead.

Sessions here run in ephemeral containers. `.beads/embeddeddolt/` is
gitignored, so the Dolt database is created fresh in each session and
destroyed with the container. `bd dolt push` cannot reach a remote: the
egress proxy returns HTTP 403, and `refs/dolt/data` does not exist on the
GitHub remote. Dolt is local and disposable, by design.

**`.beads/issues.jsonl` is the only thing that carries issues between
sessions and between branches.** It is committed to git. On a fresh
container `.claude/hooks/session-start.sh` detects the missing database,
runs `bd init`, unwinds the commit `bd init` makes, and imports this file.
Verified end to end on 2026-07-31: a fresh clone has no database at all,
`bd init` alone yields "No issues found", and `bd import .beads/issues.jsonl`
restores every issue with its original `agent-plugins-` id and status.

### What this requires of you

- **A bead that is not in `.beads/issues.jsonl` when the container dies is
  gone.** Not recoverable — the database it lived in no longer exists.
- Auto export is enabled in `.beads/config.yaml`, so creates and updates
  reach the file without being asked for. Do not rely on that alone.
- **Auto export is append/update-only. It does not propagate `bd delete`.**
  Measured 2026-07-31: a deleted issue was still present 90 seconds later,
  well past the 60s interval, and would have come back on the next import.
- **The beads pre-commit hook does not save you here.** It runs the same
  append/update path, so committing after a delete leaves the deleted record
  in the file. Measured the same day: create a probe and commit — the file
  gains it; delete the probe and commit again — the file still has it.
- **Before ending a session, run the explicit export and commit it:**

  ```bash
  bd export > .beads/issues.jsonl   # full rewrite; the only form that
                                    # reflects deletions
  git add .beads/issues.jsonl && git commit -m "..." && git push
  ```

- Diff the export before committing rather than trusting it. The intended
  change is usually two or three fields on known ids; anything wider means
  something else happened.
- `bd remember` memories are **excluded** from `bd export` by default, and
  the session-start hook imports without `--all`. Durable context belongs in
  an issue description, not a memory.
- Branches carry their own `.beads/issues.jsonl`. There is no merge driver
  configured for it, so beads created on two branches will conflict
  textually on merge. Prefer creating beads on one branch until that is
  resolved.

## Build & Test

_Add your build and test commands here_

```bash
# Example:
# npm install
# npm test
```

## Architecture Overview

_Add a brief overview of your project architecture_

## Conventions & Patterns

_Add your project-specific conventions here_
