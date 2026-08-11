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

This repository is a documentation-and-skills corpus **plus** the Go module
under `cli/` that implements the Spec CLI. The gates are the repo's own
consistency checks and, for anything under `cli/`, the Go toolchain.

Run the Go gates from `cli/` — that is the module root, so `./...` and the
`go test ./core/...` form the work orders cite both resolve there.

- **Format:** `cd cli && gofmt -l .` — must print nothing (`gofmt -w .` to fix)
- **Lint:** `scripts/check-skills.sh` and, when `cli/` changed, `cd cli && go vet ./...`
- **Build:** `cd cli && go build ./...` *(no build step for the docs corpus)*
- **Test (full suite):** `scripts/check-skills.sh --strict && scripts/sync-shared-references.sh --check && scripts/sync-strata-docs.sh --check` and, when `cli/` changed, `cd cli && go test ./...`

### Custom gates

```bash
scripts/generate-schemas.sh --check   # committed artifacts under cli/core/{schema,specmodel}
                                      # match the schemas/*.strata sources
```

Run it after any change to `schemas/*.strata`. The artifacts it checks are
**generated — never hand-edit them**; regenerate with the same script without
`--check` and commit the result in the same change.

### What CI runs

Three workflows under `.github/workflows/`, all on push and pull request
except the last:

- `checks.yml` — the docs-corpus gates on one Linux runner:
  `check-skills.sh` (not `--strict`), `sync-shared-references.sh --check`,
  `validate-corpus.py`.
- `cli.yml` — the Go gates. `go vet`, `go build` and `go test ./...` on
  ubuntu, macOS and Windows (the matrix is the cross-platform half of
  AC-SPEC-009.7: one committed set of expected bytes, three platforms);
  `gofmt -l` on ubuntu only; a cross-compile of all six release targets;
  and a run of the static binary inside a busybox image, which is
  AC-SPEC-009.4's worked example executed rather than asserted.
- `release.yml` — on a `cli/vX.Y.Z` tag. Cross-compiles the six targets
  on one Linux runner, packages one archive each plus `SHA256SUMS`, and
  publishes a GitHub Release.

Neither `check-skills.sh --strict` nor `generate-schemas.sh --check` runs
in CI — the first is known red (below) and the second needs the Strata
compiler. Both remain local gates.

`.gitattributes` pins the working tree to LF on every platform. It has to:
the expected artifacts under `cli/testdata/cases/**/expected/` are compared
byte for byte, and Git's `core.autocrlf` default on Windows would rewrite
them on checkout. The CRLF checkout AC-SPEC-009.7 names is reproduced
synthetically instead, on every platform, by `cli/cmd/portability_test.go`.

### Targeted tests by surface

- `cli/**` → `cd cli && gofmt -l . && go vet ./... && go test ./...`
- `schemas/*.strata` → `scripts/generate-schemas.sh --check`, then
  `cd cli && go build ./...` (the generated Go must still compile)

- `plugin/skills/*/SKILL.md` → `scripts/check-skills.sh`
- `plugin/skills/*/references/**` → `scripts/sync-shared-references.sh --check && scripts/check-skills.sh`
- `.claude/skills/*/SKILL.md` → `scripts/check-skills.sh` (only the
  project-native skills `triage-rfc` and `strata`; the committed symlinks
  to `plugin/skills/` are skipped and checked at their source)
- `.claude/skills/strata/references/**` → `scripts/sync-strata-docs.sh --check`.
  These are generated verbatim from `strata-lang/strata`; never hand-edit
  them. `--check` reports SKIP (exit 0) when no strata checkout is present.
- `plugin/skills/groom-beads/scripts/**` → `python3 plugin/skills/groom-beads/scripts/validate_beads.py <fixture>.jsonl`
- `plugin/skills/bead-bonanza/scripts/**` → run the script's own guards (see its header)
- `scripts/**` → `bash -n <file>` then run it in its `--check` mode

### Environmental-failure ledger

- **`bd`'s JSONL export lags the database.** A close reaches the Dolt DB
  seconds before it reaches `.beads/issues.jsonl`, so an export taken
  immediately after `bd update --status closed` can still return pre-close
  bytes. Measured 2026-08-10 in run `1786345424-bbh` on two consecutive
  beads: `agent-plugins-92q.20` (`d5a6de5`) and `agent-plugins-92q.23`
  (`905aee4`) each committed without `.beads/issues.jsonl`, and on the clean
  tree straight afterwards
  `jq -r 'select(.id=="agent-plugins-92q.20") | .status' .beads/issues.jsonl`
  printed `open` while `bd export | jq -r 'select(...) | .status'` printed
  `closed` — same for `92q.23`. Every other bead in the run exported
  correctly, so this is intermittent, not a broken build. Do not re-diagnose
  it: `scripts/bb-close-bead.sh` now re-exports a bounded number of times and
  refuses to commit a closure the ledger does not carry, and any hand-rolled
  close must verify the ledger the same way rather than trusting one export.
- **KNOWN RED, pre-existing — `scripts/check-skills.sh --strict`.** It exits 1
  on a *warning*: `plugin/skills/bead-bonanza/SKILL.md` is 859 lines against a
  "under 500" guideline. Zero errors. Do not treat this as your change's
  failure and do not try to fix it mid-run; shortening that skill is editorial
  work with its own scope. The bar is **no new findings** — compare against
  this baseline (`0 error(s), 1 warning(s)`) rather than expecting exit 0.
- **`scripts/generate-schemas.sh --check` fails when the installed `strata`
  predates `strata-lang/strata@8997153`.** That commit fixes
  `allowAdditionalProperties: false` silently dropping struct inheritance.
  An older binary regenerates `cli/core/schema/spec-model.schema.json` with
  `"properties": {}` on every inheriting type, so `--check` reports the
  committed (correct) artifacts as stale — the binary is wrong, not the repo.
  **Never "fix" this by committing the regenerated output**; that would delete
  the properties. Fix the binary:

  ```bash
  cd /path/to/strata && cargo build --release -p strata-cli   # ~1m20s
  cp -f target/release/strata /usr/local/bin/strata
  ```

  **`strata --version` cannot identify a build** — every build in the range
  that matters reports `strata-cli 0.7.0`. Check
  `~/.cache/strata-cli/built-commit` instead.
- **Build `strata` from `strata-lang/strata` HEAD**, against the manifest as
  committed in this repo. A build that is merely "recent enough" is not
  sufficient: the supported build tracks HEAD, and both older and mismatched
  builds fail — older ones *silently*, by emitting document-type structs with
  an empty `properties` map while exiting 0.

  That silent mode is why `scripts/generate-schemas.sh` asserts on the emitted
  shape (every document-type struct retains a non-empty `required` list)
  rather than on a version string. Keep the assertion: it catches this class
  of corruption from any build, not one known-bad commit.
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
