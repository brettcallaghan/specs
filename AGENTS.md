# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd prime` for full workflow context.

> **Architecture in one line:** Issues live in a local Dolt database
> (`.beads/dolt/`); cross-machine sync uses `bd dolt push/pull` (a
> git-compatible protocol), stored under `refs/dolt/data` on your git
> remote — separate from `refs/heads/*` where your code lives.
> `.beads/issues.jsonl` is a passive export, not the wire protocol.
>
> See [SYNC_CONCEPTS.md](https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md)
> for the one-screen overview and anti-patterns (don't treat JSONL as the
> source of truth; don't `bd import` during normal operation; don't
> reach for third-party Dolt hosting before trying the default).

## Non-Interactive Shell Commands

**ALWAYS use non-interactive flags** with file operations to avoid hanging on confirmation prompts.

Shell commands like `cp`, `mv`, and `rm` may be aliased to include `-i` (interactive) mode on some systems, causing the agent to hang indefinitely waiting for y/n input.

**Use these forms instead:**
```bash
# Force overwrite without prompting
cp -f source dest           # NOT: cp source dest
mv -f source dest           # NOT: mv source dest
rm -f file                  # NOT: rm file

# For recursive operations
rm -rf directory            # NOT: rm -r directory
cp -rf source dest          # NOT: cp -r source dest
```

**Other commands that may prompt:**
- `scp` - use `-o BatchMode=yes` for non-interactive
- `ssh` - use `-o BatchMode=yes` to fail instead of prompting
- `apt-get` - use `-y` flag
- `brew` - use `HOMEBREW_NO_AUTO_UPDATE=1` env var

## bead-bonanza gates

This is a documentation and skills repository — there is no compiler and no
test suite. The gates are the repo's own consistency checks.

- **Format:** *(none — markdown and shell only)*
- **Lint:** `scripts/check-skills.sh`
- **Build:** *(none)*
- **Test (full suite):** `scripts/check-skills.sh --strict && scripts/sync-shared-references.sh --check && scripts/dogfood-install.sh --check && scripts/sync-strata-docs.sh --check`

### Targeted tests by surface

- `plugin/skills/*/SKILL.md` → `scripts/check-skills.sh`
- `plugin/skills/*/references/**` → `scripts/sync-shared-references.sh --check && scripts/check-skills.sh`
- `.claude/skills/*/SKILL.md` → `scripts/check-skills.sh` (project-native
  skills — `triage-rfc`, `strata` — are checked alongside the plugin's)
- `.claude/skills/strata/references/**` → `scripts/sync-strata-docs.sh --check`.
  These are generated verbatim from `strata-lang/strata`; never hand-edit
  them. `--check` reports SKIP (exit 0) when no strata checkout is present.
- `plugin/skills/groom-beads/scripts/**` → `python3 plugin/skills/groom-beads/scripts/validate_beads.py <fixture>.jsonl`
- `plugin/skills/bead-bonanza/scripts/**` → run the script's own guards (see its header)
- `scripts/**` → `bash -n <file>` then run it in its `--check` mode

### Environmental-failure ledger

- `bd`: not preinstalled. `.claude/hooks/session-start.sh` installs it via the
  official installer, which needs `libicu-dev` for the Dolt/ICU cgo dependency
  — without those headers the build fails at `unicode/uregex.h: No such file`.
- `api.github.com` and `github.com/.../releases` return **403** through the
  egress proxy, so the installer's prebuilt-binary lookup always fails here
  and it falls back to a source build. `raw.githubusercontent.com` (200) and
  `proxy.golang.org` (in the proxy's no-proxy list) both work, which is why
  the fallback path succeeds.
- **Measured install cost:** ~175s on a cold container (both Go caches empty),
  ~110s with a warm build cache. Budget ~3 minutes for first session start.
  If release access is ever restored, the installer will pick the prebuilt
  binary automatically and this drops to seconds — no hook change needed.
- The Dolt database (`.beads/embeddeddolt`) is gitignored, so **no clone ever
  has one**. Every fresh container must run `bd init` + `bd import` before any
  `bd` command works; the hook does this.
- `strata`: not preinstalled either. The hook builds it from source — see
  "strata CLI" below.

### strata CLI

`.claude/hooks/session-start.sh` provisions the `strata` compiler because the
`strata` skill's entire method is compiling what it writes. Notes measured in
this environment:

- **Source, not a release.** The published binaries lag `main`, and the point
  is the latest code. `https://github.com/strata-lang/strata.git` clones fine
  through the egress proxy (unauthenticated, over HTTPS).
- **Measured build cost:** 1m50s for `cargo build --release -p strata-cli` on
  a cold container, and ~470 MB in `target/`. The hook stamps the built commit
  in `~/.cache/strata-cli/built-commit` and skips the build when it matches, so
  only the first session of a container pays. The SessionStart hook's `timeout`
  in `.claude/settings.json` is raised to 900s to cover it — at the 60s default
  the build is killed partway.
- **It never touches a checkout that has work in it.** An attached strata clone
  is fast-forwarded only when it is on the default branch *and* clean;
  otherwise it is built exactly as it sits.
- **Escape hatches:** `STRATA_CLI_SKIP=1` skips the whole step,
  `STRATA_SRC_DIR` points it at a specific checkout. Build output lands in
  `~/.cache/strata-cli/build.log`.
- **Prebuilt binaries are not an option here, so don't "optimize" to them.**
  Measured 2026-08-01: `api.github.com` answers **200** (so a release *lookup*
  succeeds and lists the Linux asset), but the asset download itself returns
  **404** through the proxy, at both the `/releases/latest/download/` redirect
  and the tagged URL. Release metadata being reachable is not the same as
  release bytes being reachable — that is the trap. This also refines the 403
  entry above: the bd installer's prebuilt lookup fails at the download, not
  at the API.
- **The clone is per-container, not shared.** `~/.cache/strata-src` and
  `~/.cache/strata-cli` live on the session's ephemeral disk, so a new
  container pays the full ~1m45s build once. The stamp only saves repeat
  SessionStart fires within one container (resume, `/clear`), which cost 0.4s.
- `bd init` **git-commits its own scaffolding**. The hook records HEAD and
  unwinds that commit — a session hook must not leave commits nobody asked for.
- `bd` has been observed once writing its global config to a **literal `./~/`**
  inside the repo rather than `$HOME`. Not reproducible on retry; the hook
  pins `HOME` and sweeps `./~` on exit, and `.gitignore` covers it.

## model lineup

Least to most capable. Beads carry a `tier`; the orchestrator resolves it
here at dispatch.

- **light:** `haiku`
- **standard:** `sonnet`
- **advanced:** `opus`
- **frontier:** `fable`

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->
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
   bd dolt push
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


<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
