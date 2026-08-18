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

If it skips `strata`, read the warning it logs — `AGENTS.md` ("strata
CLI") covers the causes. Everything else still lands either way.

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

**The Dolt remote is the source of truth.** Issues live in a Dolt database
synced to `refs/dolt/data` on the GitHub remote. `.beads/*.jsonl` is **not
tracked** and is not how issues travel.

This is beads' documented model, and this project used to violate it. Until
2026-08-11 the database was per-session and disposable, `.beads/issues.jsonl`
was committed, and this section told you to run `bd export` before the
container died or lose your work. `agent-plugins-92q.31` retired all of that.

### What this requires of you

- **Nothing, on the happy path.** The SessionStart hook bootstraps the database
  from the remote, and the close ceremony (`bb-close-bead.sh`) pushes after
  every close. Beads you create or close are durable without ceremony.
- **If you create or close beads outside the close ceremony, run `bd dolt
  push`.** That is the one manual step. A closure sitting only in the local
  database dies with the container, exactly as an unexported one used to.
- **Never `bd export` as a persistence step.** It writes an untracked file and
  persists nothing. Export only for a viewer or for interchange.
- **`bd remember` memories are excluded from `bd export`** and are not
  synced. Durable context belongs in an issue description, not a memory.

### The environment must provide write access to `refs/dolt/data`

Not a preference — a precondition, and deliberately without a JSONL fallback.
A session that can read but not write the tracker can close a bead and lose it.
The hook therefore **verifies** write access on every session with a real probe
push and **fails loudly** (exit 1) when it cannot, rather than degrading.

Two things the environment supplies:

1. `AGENT_PLUGINS_ACCESS_TOKEN` — a GitHub PAT with push access, set at
   container creation. Environment variables cannot be added to a running
   container.
2. **Direct, unproxied egress to `github.com`.** The egress proxy substitutes a
   brokered GitHub App token scoped to `refs/heads`, which returns 403 on
   `refs/dolt/*`. The hook configures `http.https://github.com/.proxy` empty
   and a credential helper that reads the PAT from the environment at call
   time, so the token is never written to disk.

This was misdiagnosed for a full session as "GitHub refuses `refs/dolt/*`
writes". It does not. The proof that settled it: a push with a **deliberately
invalid** token succeeded, which is only possible if the proxy replaced the
credential.

See `.claude/hooks/session-start.sh` and `docs/session-bootstrap.md`. The
`bead-bonanza` skill itself never names a credential — it requires only "write
access to `refs/dolt/data`", and the mechanism is this repository's business.

### The one thing that must never be committed

**`sync.remote` must stay UNSET in `.beads/config.yaml`.** The file is tracked,
so the key breaks every clone at once, and the failure is silent: on a fresh
clone `bd bootstrap` refuses the value —

```
error: sync.remote "git+https://github.com/…" looks like a git code-repository
URL, not a Dolt remote — skipping clone
✓ Database already exists: …/.beads    Nothing to do.
```

— and then matches the tracked `.beads` DIRECTORY with its "already exists"
probe, exits 0, and leaves a session with an empty tracker that reports itself
healthy. Every bead closed into it dies with the container.

`bd` fights this: **`bd bootstrap` writes the key back** after a successful
clone, with the exact value its own clone path refuses next time. The writer and
the validator disagree — a `bd` bug (measured on 1.2.1). Until it is fixed, the
SessionStart hook strips the key on every session. Do not commit it, and do not
"restore" it when you see the hook remove it. The remote it names is not lost:
bootstrap derives the same one from git origin's `refs/dolt/data` and stores it
in the Dolt config, which is where `bd dolt remote list` reads it from and which
is not tracked.

Stripping it is not cosmetic. `bead-bonanza`'s close ceremony stages everything
dirty in the tree, so a re-added key rides into a bead's close commit and
re-breaks every future clone.

### Verify the tracker by issue count, never by directory existence

Three things independently report a healthy tracker when it is empty: `bd
bootstrap`'s exit code, its "database already exists" probe (which matches the
tracked directory), and the presence of `.beads/embeddeddolt/` (which can
survive into a new container, stale and empty). `bd stats` reporting a non-zero
issue count is the only check that distinguishes them, and it is what the hook
uses.

### Two hazards that remain

- **The id prefix travels with the database.** A bootstrapped clone mints ids
  from the database's configured `issue_prefix`, not from `.beads/config.yaml`
  (measured: `config.yaml` said `agent-plugins`, `bd where` reported
  `agent-plugins`, and a new bead came out `spec-driven-dev-w0n`). It was
  repaired at the source with `bd rename-prefix agent-plugins --repair` and
  pushed; the hook re-checks it each session as a regression guard.
- **The hook is not always registered.** Claude Code loads hooks from the
  project root only, so a session attaching this repo alongside another never
  runs it — and then nothing bootstraps the database or establishes write
  access. The manual fallback is at the top of this file; the durable fix is
  the per-environment setup script in `docs/session-bootstrap.md`.

## Where documents live

Three homes, distinguished by what the document is *for* rather than by
what it is about.

| Home | Holds | Lifespan |
|---|---|---|
| `specs/` | The specification corpus — ADRs, FRDs, standards, blueprints, contracts. Indexed, validated, referenced by `[[slug]]` | Durable |
| `docs/research/<topic>/` | Research findings, recorded objectively | Kept while useful |
| `docs/plans/` | Plans for carrying out a change in this repository | Transient — delete when the change lands |

**A research document is not an ADR.** Research records what was
measured and what the answer turned out to be, without deciding
anything. An ADR decides something, and cites research where research is
what settled it. The two are separable on one test: if the document would
still be true had the decision gone the other way, it is research.

Keep a research document when it answers a question someone will ask
again, or when it supports a decision not yet made. Sometimes the whole
value is having asked a question and confirmed the answer — that is worth
keeping and is not worth an ADR.

Research documents are not corpus members. They live outside `specs/`,
carry no slug, and are cited by path.

## Historical prose: don't

Every skill file, reference, config comment and script header is loaded into an
agent's context window. Tokens spent narrating how a rule came to exist are
tokens not spent on the rule. Write the current state; leave the story out.

**The test:** would a reader who has never seen this project's history behave
differently for having read this sentence? If no, delete it.

**State the fact and its failure mode. Drop the account of how it was learned.**

- Yes: "`worktree.baseRef` must be `head`. Unset, a worktree branches from the
  default branch instead of session HEAD, and the bead loses every commit the
  sprint branch carries."
- No: "This was misdiagnosed for a full session as X. The proof that settled it
  was Y. It was then repaired at the source and pushed, and the hook re-checks
  it each session as a regression guard."

Both convey the rule. The second spends 4x the tokens on a story that changes
nothing a reader does.

**Where the history goes instead.** It is not lost, it moves to where it is
retrievable but not re-read every session:

| Content | Home |
|---|---|
| Why a rule exists; what was tried and rejected | the bead, and the commit message |
| A measurement that justified a decision | the bead |
| What a change retired, and why | the commit message |
| The rule itself, and what breaks without it | the skill file |

A commit message is the right place for a retirement narrative: it is attached
to the change, `git log` finds it, and nothing loads it into context.

**Three narrow exceptions** — keep it short even then:

1. **A rule that looks removable but is not.** One sentence saying why the
   obvious simplification is wrong, so it is not "fixed" every six months.
2. **A trap with a plausible wrong answer.** Where an agent would otherwise
   guess confidently and wrongly, name the wrong answer.
3. **A non-obvious measured constraint.** State it as a current fact, not as a
   discovery narrative — "`bd` resolves the shared database from any worktree
   through git's common directory", not "we measured that…".

**Do not cross-reference issue ids in skill files.** Skills are exported to
other projects, where `<this-project>-abc.12` resolves to nothing. A reader
there sees a dangling reference and cannot look it up. Cross-reference the bead
in the COMMIT that makes the change; keep the skill file self-contained.
This applies to `plugin/skills/**` and anything vendored from it. Files that
never leave this repo — `CLAUDE.md`, `AGENTS.md`, `docs/`, `.beads/config.yaml`
— may cite ids, and should still keep the narrative short.

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
