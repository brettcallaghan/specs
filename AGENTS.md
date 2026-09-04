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

## Pipes hide exit statuses

**The symptom is a command that succeeded reporting failure, or one that failed
reporting success.** Both come from putting a pipe between a command and the
status you read, and both have bitten this repository more than once.

**Never read `$?` through a pipe.** `cmd | tail` gives you `tail`'s status, so a
failing command reads as a pass. Redirect to a file and read the status
directly:

```bash
cmd >out 2>&1; echo $?          # right
cmd 2>&1 | tail -5; echo $?     # WRONG — that is tail's status
```

**Under `set -o pipefail`, a reader that closes the pipe early inverts it.**
`grep -q`, `head` and `tail -n` exit as soon as they have what they need; the
writer then dies on SIGPIPE, and `pipefail` reports the whole pipeline as
failed. Measured here: `bd show <id> | grep -qi closed` reported a CLOSED bead
as not closed.

```bash
if bd show "$id" | grep -qi closed; then      # WRONG under pipefail
out="$(bd show "$id" --json)"                 # right — capture, then test
```

The fix is the same in both directions: **put the value in a variable or a
file, then test it.** A pipeline is for moving data, not for carrying a verdict.

## Tool claims are measured at the point of use

**A claim about what a command does binds only where somebody ran it.**
Inheriting one is not measuring it — not from a document in this repository,
not from a dispatch prompt, not from an epic comment, not from a prior run's
report. Prose records nothing about which of the two it was, so a claim
nobody ran reads exactly like one somebody did.

Before you **act on** such a claim, or restate it anywhere, run the command
and read the output; carry the invocation and its output with the claim
wherever you state it. Where you cannot run it, mark the claim **unverified**
and say what would settle it. This binds hardest while *correcting* someone
else's claim — the replacement is a new claim and carries the same burden.

This is the evidence rule from the environmental-failure ledger below, moved
one step earlier: it binds when a belief is **used or copied**, not only when
a ledger entry is admitted.

### Why the ledger rule alone did not bind

The ledger has required evidence throughout, and run `agent-plugins-so39`
violated it four times while correcting three other documents for exactly
this defect. Four structural reasons, each a thing to fix rather than a thing
to try harder at:

1. **Its subject is a document section, not an author.** It governs
   *admitting an entry to the ledger*. None of the four was a ledger entry —
   they were a paragraph in a skill reference, a sentence in a dispatch
   prompt, a line of shell in one of our own scripts, and a prompt armed on a
   trigger. An agent writing any of those is not performing the act the rule
   names, so the rule never comes up for application.

2. **Its placement scopes it.** It sits under `## Checks`, headed
   "Environmental-failure ledger". An agent that needs to state a fact about
   `bd` has no reason to open a section about container environment failures,
   and a rule filed under a heading is read as being about that heading.

3. **It governs origination; three of the four were propagation.** "Run it
   first, record it after" says what to do when you *make* a claim and
   nothing about relaying one — and relaying feels like citing a source
   rather than making a claim. The citation requirement is the only thing
   separating a measured claim from an inherited one, and it is demanded only
   of ledger entries, so every other surface loses the distinction.

4. **Correction is the moment of highest confidence, and nothing gates the
   replacement.** The sharpest instance: the bead whose job was removing a
   false claim about `bd update --metadata` shipped a different false claim
   about `--set-metadata` in its place. The evidence standard was applied to
   the sentence being deleted and not to the one being installed.

A fifth reason is not about placement: **the rule has no automated check.**
`specs/standards/skill-authoring-standards.md` requires every invariant a skill
asserts about itself to be enforced by a check; this one is enforced by
authorial memory, which is the condition it exists to correct. Stating it at
the point of use narrows the gap. It does not close it.

### The four instances

- **"`--set-metadata` stores every value as a string"** — false. Measured on
  bd 1.2.2: `num=5` → number, `b=true` → boolean, `f=1.5` → number,
  `n=null` → null; only arrays, objects and non-numeric text are stored as
  strings. Installed by `agent-plugins-4tjk`, which was removing a different
  false claim — "`bd update --metadata` replaces the entire metadata object";
  measured, it merges at the **top level** and replaces a **nested** value
  wholesale. The false universal survives in `capability-tier.md`
  (`agent-plugins-so39.5`, deferred).

- **`sync_dir`'s `rm -rf "$dest"` attributed to
  `scripts/sync-shared-references.sh`** in a dispatch prompt, with an argument
  built on top of it. Measured:
  `grep -rn 'sync_dir\|rm -rf' scripts/sync-shared-references.sh scripts/generate-schemas.sh`
  returns nothing from the first file and `generate-schemas.sh:183:sync_dir() {`
  plus `generate-schemas.sh:192:    rm -rf "$dest"` from the second;
  `render()` in `sync-shared-references.sh` ends `cat "$1"`. The receiving
  agent measured instead of believing, so the rule it shipped rests on the
  real mechanic.

- **"`bd search` is non-functional"** — false; it excludes closed issues by
  default. Measured: `bd search metadata` → `Found 2 issues`; the same query
  with `--status all` → 3, the extra one `closed`. The claim was relayed into
  an epic comment **and** into a durable recovery trigger's prompt, where a
  resuming agent would have inherited it as a hard constraint. The trigger was
  deleted before firing, because a report caught it.

- **`printf … | grep -qxF` under `set -o pipefail`** in this repository's own
  orphan guard, in `scripts/sync-shared-references.sh` — the trap
  "Pipes hide exit statuses" above already names. Measured by
  `agent-plugins-b9ft`: 2 failures in 30 `--check` runs on an unchanged tree,
  herestring control 100/100 clean. Intermittent by construction — a
  three-line reproduction (`printf '%s\n' a b c | grep -qxF a`) exits 0, so
  failing to trigger it is not evidence against it.

The common shape: a claim arrived as an inherited premise and was acted on
without being measured where it was used. Three of the four were then copied
onto a surface nobody re-reads.

### Surfaces that carry a copy of a belief

A retraction reaches the surface in front of you. It does not reach the
copies, and most of them are never read again by anybody positioned to notice.

| surface | lives | re-read by | audited by |
|---|---|---|---|
| **a durable trigger prompt** (a Routine, a recovery check-in) | past the session, the container and the run | a fresh agent with no memory of the retraction, as a hard constraint | **nobody.** `list_triggers` shows that a trigger exists; nothing reads its prompt text against reality. Only whoever armed it. |
| a dispatch prompt | dies with its agent — but goes to several agents at once | its recipient, once | nobody; it is not in the repo and no check sees it. The dispatcher, before sending. |
| a `.sf/outbox` issue | until triaged, possibly across runs and repositories | the triaging agent, who reads it as already human-approved | partially. `triage-issue` re-checks that the quoted evidence still *exists* in the source it cites; nothing re-runs a tool claim the issue asserts, so a false one passes triage intact and lands in the run record |
| a bead description or acceptance criterion | until the bead closes; longer through `discovered-from` edges | the implementing agent, cold, without the conversation that retracted it | grooming, once; after that nobody |
| an epic comment, a run record, a sprint report | permanent | the report and the retro | the report's author |
| a plan document under `docs/plans/` | until the change lands | whoever resumes the plan | nobody — transient by policy, so no check covers it |
| a committed document (`AGENTS.md`, `specs/**`, a `SKILL.md`) | permanent | every future agent | `scripts/ci.sh`, for shape — never for truth |
| a generated `references/shared/` copy | permanent | every agent that loads the consuming skill | `scripts/sync-shared-references.sh --check` propagates a corrected source, and reaches no hand-written citation of the same claim |
| this repository's own scripts and checks | permanent | the runtime, every run | the check itself, where one exists |

**The durable trigger prompt is the worst of them**, and the one the rest of
the corpus does not mention: it is the only surface that can deliver a
retracted claim to an agent with no way to learn it was retracted. A dispatch
prompt is the same defect one tier less durable.

### Retracting a claim found false

Not a checklist for its own sake. The point is that a retraction is a sweep,
not an edit.

1. **Record the measurement first** — the command and its output, before you
   change anything. It is what every replacement will cite.
2. **Grep the claim's distinctive phrasing across the tree BEFORE you fix the
   source.** Afterwards the string is gone, and the one place you were certain
   of it is the place you just destroyed. Include untracked files —
   `.sf/outbox/`, `docs/plans/`.
3. **Sweep in order of reach**: durable triggers first, because they can fire
   while you work; then beads and epic comments; then documents; then
   generated copies, via the sync script.
4. **Measure the replacement.** Do not install a corrected universal you have
   not run.
5. **Say where you could not reach.** A prompt already sent, a run record
   already written, a trigger in another account — an unreachable copy is a
   finding to file, not a silence.

## Checks

This repository is a documentation-and-skills corpus **plus** the Go module
under `cli/` that implements the Spec CLI. Run the Go checks from `cli/` if you
invoke them individually — that is the module root, so `./...` resolves there.

Two lists, and they are not the same list:

| | |
|---|---|
| `scripts/ci.sh` | everything a contributor should see green before pushing |
| `swarm.toml`'s `[[checks]]` | the quality bar for ONE COMMIT — what a bead runs before its own close |

They overlap today and will not stay that way: anything too slow to pay per
bead belongs in `ci.sh` and not in the declaration. Neither is the **release
bar** — the cross-compile matrix, the three-platform test matrix and the
busybox smoke test only mean something across the whole repository, and
`.github/workflows/` owns them.

One trap worth naming, because the obvious fix is wrong: **`format` is a file
list, not an exit status.** `gofmt -l` exits 0 whether or not it found
unformatted files, so a check reading only the status always passes. That is
why it points at `scripts/format-check`.

**Coverage belongs to none of those lists.** `scripts/coverage.sh` measures the
unit, snapshot and end-to-end layers separately and writes the merged figures
into `README.md` between generated markers; `cli.yml`'s `coverage` job runs
`scripts/coverage.sh --check` and fails when the committed figures do not match
a fresh measurement. A change under `cli/**` is therefore not finished until
that script has been run from the repository root and the README diff committed
with it. Regenerate with the Go version in `cli/go.mod` — statement counts come
from the toolchain, so a different one moves the numbers. It is in `ci.sh` and
deliberately not in `sf.toml`: three instrumented test legs is not a price
worth paying per bead.

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
  it: any hand-rolled close must verify the ledger — re-export a bounded number
  of times and refuse to commit a closure the export does not carry — rather
  than trusting one export. (`swarm`'s `bb-close-bead.sh` does exactly this;
  the trap is `bd`'s, not that script's, so it applies here too.)
- **The hook's proxy bypass is SCOPED to origin's URL, and must stay that way.**
  A host-wide `http.https://github.com/.proxy` reaches every other github.com
  remote in the container while the credential helper beside it does not, so
  those remotes lose the proxy's brokered credential and gain nothing in its
  place. Measured in a session holding `agent-plugins` and `swarm`, with the
  host-wide form in place: a push to `swarm` died with `could not read Username
  for 'https://github.com'`, and the same push succeeded under
  `git -c "http.https://github.com/.proxy=$HTTPS_PROXY" push`. After scoping,
  `git config --get-urlmatch http.proxy <url>` returns rc=0 (bypass applies)
  for `https://github.com/brettcallaghan/agent-plugins` and its `.git` form,
  and rc=1 (proxy stays on) for `https://github.com/brettcallaghan/swarm`;
  `git ls-remote origin HEAD` exits 0 in both clones with no `-c` override.
- **A push to THIS repository touching `.github/workflows/` is refused, and
  that is the bypass working as designed.** The bypass is what routes this
  remote to the real PAT, and `AGENT_PLUGINS_ACCESS_TOKEN` has no `workflow`
  scope. Measured: `! [remote rejected] ... refusing to allow a Personal Access
  Token to create or update workflow '.github/workflows/cli.yml' without
  'workflow' scope`; the same push succeeded through
  `git -c "http.https://github.com/.proxy=$HTTPS_PROXY" push`, which restores
  the proxy's brokered token for that one command. Git config cannot scope a
  bypass by refspec, so the bypass cannot apply to `refs/dolt/*` alone — the
  two remedies are that per-command override, or granting the PAT `workflow`
  scope. Do NOT "fix" it by widening the bypass back to the host.
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
  when the tracker reports zero issues.
- **`bd stats` exits non-zero when there is no database, and the hook runs under
  `set -euo pipefail`.** Until 2026-08-20 that combination made the hook die at
  `_n="$(_count)"` — silently, no message, exit 1 — on any clone without a
  database, which is the one case the bootstrap step exists for. Measured on a
  fresh clone against both the then-current hook and its predecessor: two lines
  of output, exit 1, no bootstrap attempted. `_count` now swallows the status
  internally. Any future edit there must keep doing so.
- **A database-name mismatch is the one tracker failure bd does NOT report.**
  Measured on bd 1.2.2 against a database holding 318 issues: setting
  `.beads/metadata.json`'s `dolt_database` to a name absent from
  `.beads/embeddeddolt/` makes bd CREATE an empty database under that name;
  `bd stats` then prints `Total Issues: 0` and `bd ready` prints "No open
  issues", both exit 0, no warning. The hook checks this **before** anything
  reads the tracker and treats it as fatal — ordered after the issue count, the
  measured sequence was worse than the bug itself: the count read 0, the hook
  moved the real database aside as stale, re-cloned under the wrong name, and
  the guard then saw the directory present and stayed silent.
- **Bootstrap cannot clone over an existing database, and says so.** Measured on
  1.2.2 with a genuinely initialised empty database in place: exit 1,
  `can't create database beads; database exists`. An empty *directory* at
  `.beads/embeddeddolt/` does not block it. This is why the hook moves a stale
  database aside rather than skipping bootstrap when the directory exists.
  1.2.2 also compares `metadata.json`'s `project_id` against the database's and
  refuses loudly on a mismatch — measured: `workspace identity mismatch
  detected`.
- **The issue prefix comes from the database, not `config.yaml`.** Measured on
  1.2.2: with `issue-prefix: "zzztest"` in `.beads/config.yaml` and a database
  configured `agent-plugins`, `bd create` minted `agent-plugins-4731`. The file
  key is ignored entirely.
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

## Citing another skill's reference

`scripts/check-skills.sh` makes one skill unable to *link* another skill's
reference file: rule 3 errors on any `../<skill>/references/` path, and rule 2
requires a backticked `*.md` path to resolve from the citing file's own
directory. That is deliberate — a skill is copied out of this repo on its own,
and a path reaching outside its directory dangles the moment it is. Two remedies
exist, and **which one is right is a rule, not a preference:**

> **Vendor when the citing skill needs the WHOLE cited file and only APPLIES
> it — a procedure, checklist, standard or template. Cite the owning skill by
> ROLE in prose, with no path, when you need one fact out of a larger document,
> or when the file is a REGISTRY that a second copy would give a second place
> to extend.**

**The two tests, in order.**

1. **Do you need the whole file?** `scripts/sync-shared-references.sh` vendors a
   file, never a section — `render` is `cat` over the source, so there is no way
   to generate part of one, and none to graft a section into a hand-authored
   file. Needing one fact out of a larger document therefore has no vendoring
   form: it would ship the citing skill a document it does not own and must not
   act on. **One fact out of a larger document → cite by role.**
2. **Is the file applied, or extended?** A procedure, checklist, standard or
   template is *applied*: it reads identically in every skill that ships it, and
   a generated copy is inert. A registry — a roster, an id table, an enum of
   kinds — is *extended*: its content asserts that it is the one place an entry
   may be added, and a second copy makes that claim false the moment somebody
   adds to the wrong one. The generator prevents drift between copies; it does
   not prevent a reader from treating a copy as a home. **Extended → cite by
   role. Applied → vendor.**

**VENDOR** means: add a `consumer<TAB>source` line to
`scripts/sync-shared-references.sh`'s `MAP`, run the script, and cite the
generated `references/shared/<file>.md` by path from the citing skill. The cost
you accept is one more generated file that `--check` must keep current forever.

**CITE BY ROLE** means: state the fact the reader needs, in the citing skill's
own words, and name the owner by its **role** rather than by a path — "the
coordinator's dispatch policy", "the grooming skill's Definition of Ready". No
path, because no path resolves for a reader who has only the citing skill. The
cost you accept is an unlinked hop, so say enough that a reader who never makes
the hop still behaves correctly.

**Both forms are already in the tree, and each is the right one for its case:**

| citation | whole file? | applied or extended? | form |
|---|---|---|---|
| `groom-beads` → the `spec` skill's work-order template | yes — the file *is* the shape a groomed bead takes | applied (a template) | **vendor** |
| `groom-beads` → the work-order Definition of Ready | yes — the file *is* the gate it runs | applied (a checklist) | **vendor** |
| a skill → one fact out of `bead.schema.json`'s tier enum | no | extended (a registry) | **cite by role** |

Where the two tests disagree, test 1 wins — a section cannot be vendored at
all.

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

- **Nothing.** The SessionStart hook bootstraps the database when it is empty
  and otherwise pulls; a `Stop` hook (`.claude/hooks/dolt-push.sh`) pushes at
  the end of every turn. Beads you create or close are durable without
  ceremony, in or out of the close ceremony. Pushing by hand is never wrong,
  just redundant.
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

Both re-measured 2026-08-20 on this repository, four pushes:

| proxy | token | target | result |
|---|---|---|---|
| bypassed | real PAT | `refs/dolt/*` | succeeds |
| enabled | real PAT | `refs/dolt/*` | **403** |
| enabled | garbage | `refs/heads/*` | **succeeds** |
| bypassed | garbage | `refs/dolt/*` | auth failed |

Row 3 is the one that settles it: a push authenticated with a deliberately
invalid token cannot succeed unless the proxy replaced the credential. So this
is not GitHub refusing `refs/dolt/*` writes — it is the proxy's brokered token
being scoped to `refs/heads`.

The credential helper is registered under **both** the suffixless and the
`.git` form of origin's URL because git matches the config key's path against
the request's path exactly, and only the matching form answers — measured
against a clone whose origin carries `.git`: suffixless key alone → `could not
read Username for 'https://github.com'`; `.git` key alone → push succeeds.
This repository's origin currently has no suffix, so the suffixless key is the
live one and the other is the spare; which is which depends on how the clone
was made.

See `.claude/hooks/session-start.sh` and `docs/session-bootstrap.md`. A skill
that needs a durable tracker never names a credential — it requires only "write
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
