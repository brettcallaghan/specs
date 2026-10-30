# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd prime` for full workflow context.

> **Architecture in one line:** Issues live in a local Dolt database
> (`.beads/embeddeddolt/`); cross-machine sync uses `bd dolt push/pull`
> against the `beads/data` branch of this repository's git remote (prefix
> `specs-`), set up per clone by `scripts/beads-setup.sh`.
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

The ledger has required evidence throughout, and one sprint run
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
  strings. Installed by a change that was removing a different
  false claim — "`bd update --metadata` replaces the entire metadata object";
  measured, it merges at the **top level** and replaces a **nested** value
  wholesale. The false universal survives in `capability-tier.md`.

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
  "Pipes hide exit statuses" above already names. Measured:
  2 failures in 30 `--check` runs on an unchanged tree,
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
| an issue handed to another project (a GitHub issue, a Markdown file) | until triaged, possibly across runs and repositories | the triaging agent, who reads it as already human-approved | partially. `triage-issue` re-checks that the quoted evidence still *exists* in the source it cites; nothing re-runs a tool claim the issue asserts, so a false one passes triage intact and lands in a bead |
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
   `docs/plans/`, and any issue file still waiting to be handed over.
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
between generated markers: the tables into `docs/code-coverage.md`, the badge
into `README.md`; `cli.yml`'s `coverage` job runs
`scripts/coverage.sh --check` and fails when the committed figures do not match
a fresh measurement. A change under `cli/**` is therefore not finished until
that script has been run from the repository root and the diff of both files
committed with it. Regenerate with the Go version in `cli/go.mod` — statement counts come
from the toolchain, so a different one moves the numbers. It is in `ci.sh` and
deliberately not in `swarm.toml`: three instrumented test legs is not a price
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

- **`bd dolt pull` deadlocks the container, and `bd`'s own push-failure message
  recommends it.** It takes the shared embedded-Dolt write lock, does not
  return, ignores `SIGTERM`, and starves every other `bd` call in the container
  until `kill -9`. The trap is a chain: a `bd dolt push` that loses a race to a
  concurrent writer fails non-fast-forward and ends `hint: Integrate the remote
  changes (e.g. 'dolt pull ...') before pushing again`. **Do not follow that
  hint.** A closure already committed to the local database is not lost; the
  repository's `Stop` hook (`.claude/hooks/dolt-push.sh`) is the documented
  pusher, and in a swarm run tracker durability is the coordinator's.
  Measured twice, independently, in one sprint run: during one pull
  `ps -eo pid,etimes,time,stat,args` showed `bd dolt pull` resident with three
  sibling `bd` calls stacked behind it at **zero CPU** for 337s/211s/209s;
  `/proc/<pid>/wchan` read `futex_do_wait` with no open sockets, so it looks
  hung and is not doing network I/O; `free -m` showed 14656 MB of 16075 free
  and `dmesg` carried no OOM kill; `kill -TERM` returned 0 and the process was
  still alive 8s later; `kill -9` cleared it and all three siblings completed
  immediately, with no committed data lost. **A third measurement, taken at
  that run's post-flight with NO concurrent agent running, settles the scope:
  contention is not the cause.** `ps -eo pid,args | grep -E '[b]d |[d]olt'`
  showed no other `bd` or `dolt` process; `timeout 100 bd dolt pull` was still
  resident at 126s at 8.9% CPU, so `timeout(1)`'s `SIGTERM` was ignored exactly
  as with siblings present; `kill -9` cleared it and `bd stats` then read 654
  issues with every bead of that run still closed. So a lone `bd dolt pull`
  deadlocks on its own — losing a push race is what *leads you to run it*, not
  what makes it hang. Three consequences worth carrying: a `bd` call that
  appears hung should be checked against `ps` for any `bd dolt` before being
  treated as a failure; a 120s tool timeout is too short for any `bd` call made
  while a pull may be running; and **a non-fast-forward `bd dolt push` has no
  in-container remedy** — `.claude/hooks/dolt-push.sh` carries no pull, merge or
  retry path, so a diverged Dolt remote leaves the tracker local-only until
  a human reconciles it. Measured in that state: `bd dolt push` → exit 1,
  `! [rejected] main -> main (non-fast-forward)`, remote at `3c3103a`.
- **`bd`'s JSONL export lags the database.** A close reaches the Dolt DB
  seconds before it reaches `.beads/issues.jsonl`, so an export taken
  immediately after `bd update --status closed` can still return pre-close
  bytes. Measured 2026-08-10 on two consecutive beads of one run (commits
  `d5a6de5` and `905aee4`), each committed without `.beads/issues.jsonl`;
  on the clean tree straight afterwards
  `jq -r 'select(.id=="<bead-id>") | .status' .beads/issues.jsonl`
  printed `open` while `bd export | jq -r 'select(...) | .status'` printed
  `closed`, for both beads. Every other bead in the run exported
  correctly, so this is intermittent, not a broken build. Do not re-diagnose
  it: any hand-rolled close must verify the ledger — re-export a bounded number
  of times and refuse to commit a closure the export does not carry — rather
  than trusting one export. (`swarm`'s `bb-close-bead.sh` does exactly this;
  the trap is `bd`'s, not that script's, so it applies here too.)
- **`strata` must be v0.8.0 or later.**
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
  has one**. `scripts/beads-setup.sh`, which the hook runs, clones it from the
  `beads/data` branch when the tracker reports zero issues.
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
  reads the tracker and moves the whole `.beads/embeddeddolt/` aside, so the
  named database is cloned fresh — ordered after the issue count, the
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
  configured `agent-plugins`, `bd create` minted an `agent-plugins-` id. The file
  key is ignored entirely.
- `strata`: not preinstalled either. The hook installs the release binary —
  see "strata CLI" below.
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

`scripts/install-strata-cli.sh` installs the latest `strata-lang/strata`
release binary to `/usr/local/bin/strata`; the session-start hook runs it in
cloud sessions, and anyone can run it locally. It is a no-op when
`~/.cache/strata-cli/installed-release` already names the latest tag.

- **Needs `STRATA_ACCESS_TOKEN`** while `strata-lang/strata` is private
  (`docs/session-bootstrap.md`).
  Without it the release lookup fails: measured, unauthenticated direct
  `api.github.com/repos/strata-lang/strata/releases/latest` → 403.
- **Its GitHub requests fall back to bypassing the egress proxy, and must.**
  The proxy replaces the credential: measured with `STRATA_ACCESS_TOKEN` set
  and strata not attached to the session, the release lookup through the proxy
  → 403 "GitHub access to this repository is not enabled for this session";
  the same request under `curl --noproxy '*'` → 200, and the asset download
  (redirected to `release-assets.githubusercontent.com`) → 200. The proxy is
  tried first so an environment that does grant access through it still works.
- **Verified before install**: against the `sha256:` digest GitHub publishes
  for the asset, then by running `strata --version`. Pin a release with an
  argument (`scripts/install-strata-cli.sh v0.8.0`) or `STRATA_VERSION`.
- **v0.8.0 publishes no `aarch64-unknown-linux-gnu` asset** (measured from the
  release's asset list); the script names the missing asset and exits 1.
- **A failed install is fatal in the hook**, because a session without the
  compiler can edit a `.strata` file but not compile it. `STRATA_CLI_SKIP=1`
  starts a session that deliberately does no schema work.
- **Cost:** measured 1.3s for a cold install through the hook's section.

## Where documents live

Four homes, distinguished by what the document is *for* rather than by
what it is about.

| Home | Holds | Lifespan |
|---|---|---|
| `specs/` | The specification corpus — ADRs, FRDs, standards, blueprints, contracts. Indexed, validated, referenced by `[[slug]]` | Durable |
| `docs/{tutorials,how-to,reference,explanation}/` | User documentation, one Diátaxis type per directory, indexed by `docs/README.md` and linked from the root `README.md`. Write it with the `tech-writer` skill | Durable |
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

**Do not cite bead ids in any committed file.** Code, specs, skills, docs,
retros and RFC records alike. The tracker is kept lean and beads are not
preserved forever, so an id in a file becomes a reference nobody can look up
— and skills are exported to other projects, where `<this-project>-abc.12`
never resolved at all. Cross-reference the bead in the COMMIT message that
makes the change, and in the file state the fact the id stood for.

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

**Architecture in one line:** issues live in a local Dolt DB; sync uses the `beads/data` branch on the git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

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

**The `beads/data` branch is the source of truth.** Issues live in a Dolt
database synced to the ordinary branch `beads/data` on this repository's git
remote, with the `specs-` prefix. Dolt also keeps its bookkeeping branch
`__dolt_remote_info__` there. `.beads/*.jsonl` is **not tracked** and is not
how issues travel.

**Ids of the form `agent-plugins-…` are retired.** Until 2026-10-06 the
tracker lived under `refs/dolt/data` with that prefix. Open work moved to this
database under the same suffix with the `specs-` prefix; closed beads were
not carried over.

### What this requires of you

- **Outside a cloud session, run `./scripts/beads-setup.sh` once after
  cloning.** It needs `bd`, `git` and `python3`, and no token. In a cloud
  session the SessionStart hook runs it for you.
- **Otherwise nothing.** A `Stop` hook (`.claude/hooks/dolt-push.sh`) pushes
  at the end of every turn. Beads you create or close are durable without
  ceremony. Pushing by hand (`bd dolt push`) is never wrong, just redundant.
- **Never `bd bootstrap` or `bd init` by hand.** Both read Dolt's default ref,
  `refs/dolt/data`, which does not exist here, and `bd` has no config key for
  any other ref. The failure is silent: measured on bd 1.3.1 against a remote
  holding only `beads/data`, `bd bootstrap --yes` exits 0, prints "Created
  fresh database", and `bd stats` reads 0 issues. `scripts/beads-setup.sh` is the only supported way to
  hydrate a clone. Refresh a broken one with
  `rm -rf .beads/embeddeddolt && ./scripts/beads-setup.sh`, after
  `bd dolt push` if it holds beads not yet pushed.
- **Never `bd export` as a persistence step.** It writes an untracked file and
  persists nothing. Export only for a viewer or for interchange.
- **`bd remember` memories are excluded from `bd export`.** Durable context
  belongs in an issue description, not a memory.

### Write access to the tracker is write access to the repository

A Claude Code cloud session's GitHub credential is accepted on `refs/heads/*`
and refused on `refs/dolt/*`, which is why the data lives on a branch. Anyone
who can push a branch here can push the tracker; anyone who can clone can read
it. No PAT, proxy bypass, Dolt credential or extra network host is involved.

- **Never put a force-push block on `beads/**` or `__dolt_remote_info__`.**
  Dolt syncs with `--force-with-lease`; a rule forbidding force pushes breaks
  every `bd dolt push`. Restricting who may push, and forbidding deletion, is
  fine.
- **Never merge, rebase or open a pull request from `beads/data`.** Its
  commits are Dolt's storage, not source files.
- **`refs/dolt/data` must stay absent.** A clone that reached it would get a
  tracker that is populated, plausible and wrong. `dolt-push.sh` refuses to
  push unless the Dolt remote targets `refs/heads/beads/data`.
- **swarm's pre-flight probe pushes to `refs/dolt/*` and will report the
  tracker as not durable.** That is a false positive here. Verify the real
  path with `bd dolt push` (exit 0, "Push complete.") and proceed.

The hook exports `BD_NO_REMOTE_ADOPT=1` so `bd` never adds a Dolt remote
without the branch ref, and `.beads/config.yaml` keeps `dolt.auto-push`,
`import.auto` and `backup.enabled` off.

A skill that needs a durable tracker never names a credential — it requires
only "write access to the tracker's Dolt remote", and the mechanism is this
repository's business.

### Two hazards that remain

- **The id prefix travels with the database.** A bootstrapped clone mints ids
  from the database's configured `issue_prefix`, not from `.beads/config.yaml`.
  The hook warns when it is not `specs`. Do not repair it with
  `bd rename-prefix --repair`, which rewrites ids to random hashes when it
  misdetects the prefix.
- **The hook is not always registered.** Claude Code loads hooks from the
  project root only, so a session attaching this repo alongside another never
  runs it — and then nothing hydrates the database. Run
  `./scripts/beads-setup.sh` by hand. The manual fallback is at the top of `CLAUDE.md`; the durable fix
  is the per-environment setup script in `docs/session-bootstrap.md`.
