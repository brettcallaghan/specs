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
- **Test (full suite):** run the three checks independently — never chain them
  with `&&`, since `check-skills.sh --strict` is known red (see the ledger
  below) and would short-circuit the other two before they ever run:

  ```bash
  scripts/check-skills.sh --strict;            s1=$?
  scripts/sync-shared-references.sh --check;    s2=$?
  scripts/sync-strata-docs.sh --check;          s3=$?
  echo "check-skills --strict:        exit=$s1"
  echo "sync-shared-references --check: exit=$s2"
  echo "sync-strata-docs --check:      exit=$s3"
  [ "$s1" -eq 0 ] && [ "$s2" -eq 0 ] && [ "$s3" -eq 0 ]
  ```

  Report each check's exit status; the gate as a whole fails if any of the
  three failed. And, when `cli/` changed, `cd cli && go test ./...`.

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
  `check-reference-links.py`, `check-template-drift.py`. Its `corpus` job
  builds the Spec CLI and runs `spec check --strict` over `specs/`, then
  `spec check --config rfcs/specs.toml` over the RFC queue with a narrowed
  `--only` rule set.
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

**Entries are admitted on evidence, not on reasoning.** An entry asserting
what is impossible, unnecessary, unreachable, or already-verified — including
a claim of sufficiency ("this check verifies X") — cites the command run and
the output observed; run it first, record it after. A true premise plus a
plausible inference is not evidence.

State the conditions the command ran under when they could change the
result: attached repositories, credentials present, cache warmth, toolchain
version. A reachability or permission test run inside a session that already
holds a grant measures the grant, not the capability.

Where a claim cannot be tested cheaply, mark it **unverified** and state what
would settle it — an unverified entry is useful; one that reads as settled
and is not, is worse than absent.

Cite the measurement that established the entry, not the date someone last
re-ran it. "Re-verified `<date>`" is provenance for the audit rather than for
the claim: it ages on its own, and a reader who trusts it inherits whatever
the number happened to be that day. Where the fact is volatile — a file size,
a warning count, a line number — record the command that measures it and no
figure at all. Scope an entry to the repository, never to a session: "out of
scope for this run's beads" tells a future reader nothing they can act on.

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
  on a *warning*, not an error. Do not treat the warning as your change's
  failure and do not try to fix it mid-run; it is editorial scope owned
  elsewhere. The bar is **no findings beyond this baseline**, not exit 0.
  Measure the baseline at pre-flight rather than quoting a number from here.
  The warning is a body size that moves with every editorial bead, so any
  figure recorded in this file is stale on arrival — and a stale baseline read
  as current is how a real regression gets waved through. Capture `$?`
  directly, never through a pipe; piping to `tail` reports the pager's status
  and reads as a pass:

  ```bash
  scripts/check-skills.sh --strict >out 2>&1; echo $?
  ```
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

  **Both entries above are unverified**: no run has been cited showing the
  empty-`properties` failure. Settle by running
  `scripts/generate-schemas.sh --check` against a `strata` build older than
  `8997153` and confirming the emitted schema shows empty `properties` maps
  on inheriting types.
- `bd`: not preinstalled. `.claude/hooks/session-start.sh` installs it via the
  official installer, which needs `libicu-dev` for the Dolt/ICU cgo dependency
  — without those headers the build fails at `unicode/uregex.h: No such file`.
  **Partially re-verified 2026-08-12**: `dpkg -l | grep libicu` shows
  `libicu-dev` and `libicu74` both installed in this container, consistent
  with the hook needing them. The failure text itself is **unverified** —
  reproducing it means uninstalling a package this session depends on.
- `api.github.com` and `github.com/.../releases` return **403** through the
  egress proxy, so the installer's prebuilt-binary lookup always fails here
  and it falls back to a source build. `raw.githubusercontent.com` (200) and
  `proxy.golang.org` (in the proxy's no-proxy list) both work, which is why
  the fallback path succeeds. **Re-verified 2026-08-12** with `curl -o
  /dev/null -w '%{http_code}'`: `api.github.com/repos/strata-lang/strata/
  releases/latest` → 403; `github.com/strata-lang/strata/releases` → 403;
  `raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh` (the
  installer's actual URL) → 200; `proxy.golang.org/github.com/gastownhall/
  beads/@v/list` → 200 (unproxied, per `no_proxy`). Same result set as
  documented.
- **Measured install cost:** ~175s on a cold container (both Go caches empty),
  ~110s with a warm build cache. Budget ~3 minutes for first session start.
  If release access is ever restored, the installer will pick the prebuilt
  binary automatically and this drops to seconds — no hook change needed.
  **Unverified this session** — `bd` (`bd version 1.2.1`) was already
  installed when this bead started, so no fresh cold/warm timing was taken.
- The Dolt database (`.beads/embeddeddolt`) is gitignored, so **no clone ever
  has one**. The hook bootstraps it from the Dolt remote with `bd bootstrap`
  when both `.beads/embeddeddolt` and `.beads/dolt` are absent
  (`.claude/hooks/session-start.sh`, step "2b. The database"). **Corrected
  2026-08-12** — this entry previously said the hook runs `bd init` +
  `bd import`; verified by reading the current hook source that it does not.
- `strata`: not preinstalled either. The hook builds it from source — see
  "strata CLI" below.
- **Corrected 2026-08-12, was stale:** this entry previously read "`bd init`
  git-commits its own scaffolding; the hook records HEAD and unwinds that
  commit." The hook no longer runs `bd init` (see above) and contains no
  `git commit`, `git add`, or `git reset` call anywhere — verified by reading
  the current `.claude/hooks/session-start.sh` in full. The underlying rule
  still holds for any future hook logic: a session hook must not leave
  commits nobody asked for.
- `bd` has been observed once writing its global config to a **literal `./~/`**
  inside the repo rather than `$HOME`. Not reproducible on retry; the hook
  pins `HOME` and sweeps `./~` on exit, and `.gitignore` covers it.
  **Re-verified 2026-08-12** by reading the hook source: it still exports
  `HOME` and defines/traps a `sweep_literal_tilde` cleanup on exit.

### strata CLI

Local sessions already have `strata` on PATH; this section is about cloud
sessions, where `.claude/hooks/session-start.sh` provisions it because the
`strata` skill's entire method is compiling what it writes.

- **Built from source at HEAD, not from a release.** The published binaries
  lag `main` and the point is the latest code. Do not "optimize" this into a
  release download: both the release API and the release assets are blocked by
  the egress proxy, so a clone plus `cargo build` is the only path that works
  here. **Re-verified 2026-08-12** (updates "403 and 404 respectively, last
  measured 2026-08-09"): `api.github.com/repos/strata-lang/strata/
  releases/latest` → 403; `github.com/strata-lang/strata/releases` → 403. A
  direct download URL for a specific release asset returned 404, but that
  asset/tag may simply not exist — not distinguishable from a block with the
  URL tried, so the original "403 and 404 respectively" pairing is
  **unverified** as stated; the net effect (no release path works) is
  unchanged.
- The hook prefers a checkout the session already has — an attached `strata`,
  or `STRATA_SRC_DIR` — and clones only as a fallback. It never touches a
  checkout with work in it: it fast-forwards only a clean default branch, and
  otherwise builds what it finds. **Re-verified 2026-08-12** by reading the
  hook source (candidate checkout search, and the fast-forward-only-a-clean-
  default-branch guard).
- **Escape hatches:** `STRATA_CLI_SKIP=1` skips the step entirely;
  `STRATA_SRC_DIR` points at a checkout. Build output lands in
  `~/.cache/strata-cli/build.log` and the built commit in
  `~/.cache/strata-cli/built-commit`. **Re-verified 2026-08-12**: both
  environment variables appear in the current hook source, and both cache
  files exist in this container.
- **If `strata` is missing, read the hook's warning** — it names the cause.
  The common one is the clone: `strata-lang/strata` is private *today*, so a
  cloud session with no attached checkout needs `STRATA_ACCESS_TOKEN` set on
  the environment (`docs/session-bootstrap.md`). That requirement goes away
  when the repository goes public — `agent-plugins-6gn` tracks removing it.
  **Re-verified 2026-08-12**: the exact warning text is present in the hook
  source.
- **Cost:** ~1m50s and ~470 MB of `target/`, once per container. The
  SessionStart `timeout` in `.claude/settings.json` is 900s to cover it; at
  the 60s default the build is killed partway. **Re-verified 2026-08-12** from
  this session's own build: `~/.cache/strata-cli/build.log` shows `Finished
  \`release\` profile [optimized] target(s) in 2m 35s`, and
  `du -sh ~/.cache/strata-src/target` → `470M` (exact match). The time is
  higher than the documented ~1m50s on this run — cache warmth and container
  CPU allocation can move it; treat ~1m50s–2m35s as the observed range rather
  than a fixed number.

## model lineup

Least to most capable. Beads carry a `tier`; the orchestrator resolves it
here at dispatch.

- **light:** `haiku`
- **standard:** `sonnet`
- **advanced:** `opus`
- **frontier:** `fable`

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->
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
