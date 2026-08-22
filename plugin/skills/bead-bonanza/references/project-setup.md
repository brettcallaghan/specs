# bead-bonanza Project Setup

What a project must provide for bead-bonanza to run: declared checks, a model
lineup and a code reviewer in `sf.toml`, and a few command permissions. **Read
this only when one is missing** — once they exist, Phase 0 loads them and never
comes back here.

## Contents

- The `checks` script — how this project validates a bead
- Model lineup — how capability tiers resolve to models (`sf.toml` `[models]`)
- Code reviewer — which reviewer the run dispatches (`sf.toml` `[review]`)
- Issue outbox — where findings about someone else's skills go
  (`.sf/outbox/`)
- Command permissions a run needs

Propose a draft, let the user confirm or correct it, then write it — never
write any of it silently.

---

1. **Resolve the checks script.** bead-bonanza learns how to validate a bead
   by running the project's **`checks` script**.

   **It is the quality bar for a single commit, not the release bar.** Every
   bead runs it before its own close, so what belongs in it is the checks that
   give real confidence in one bead's change and are cheap enough to pay every
   time. A check that takes minutes, or that only means anything across the
   whole repository at release time, belongs in CI — a `checks` script a bead
   cannot afford to run is one agents will find reasons to skip.

   They are declared in `sf.toml`, a plugin-scoped config file at the
   repository root — an array of tables, two keys each, and no others accepted:

   ```toml
   [[checks]]
   id = "test"
   command = 'cargo test'

   [[checks]]
   id = "lint"
   command = 'cargo clippy -- -D warnings'
   ```

   `id` is a slug (`[a-z0-9][a-z0-9-]*`), unique. `command` runs through
   `sh -c` with the repository root as its working directory. Prefer TOML
   *literal* strings (single quotes) — they do no escape processing, so a shell
   one-liner survives verbatim. `$BB_SF_TOML` moves the file; nothing else
   overrides it, and Phase 0 blocks the run when the declaration is missing or
   invalid.

   **The runner ships with this skill**: `scripts/bb-checks.py`. No project
   writes one.

   ```
   bb-checks.py                    every declared check
   bb-checks.py <id>...            only these, still in declaration order
   bb-checks.py --ids              what is declared
   bb-checks.py --report <path>    also write `<id><TAB><exit>` per check there
   ```

   Exit 0 all passed, 1 something failed, 3 the declaration itself is broken —
   and a 3 runs nothing, because a bad declaration must not half-execute.

   **Two rules the runner enforces, so no project has to remember them:**
   execution is serial in declaration order, so a step that must precede
   another is sequenced by where it sits in the file; and every requested check
   runs even after one fails. Chaining with `&&` breaks the second, which is
   how a known-red check masks everything behind it.

   Phase 0 captures what the checks print on the untouched tree as the
   **checks baseline**, and a bead compares its own run against that text — a
   line in both was already failing, a line only in the bead's is the bead's.

   - **If `[[checks]]` exists**, Phase 0 runs it. Fast path — no prompt.
   - **If it's missing**, discover candidate commands from `AGENTS.md` /
     `CLAUDE.md` (build/test/lint/format sections, CI config), propose a draft
     declaration to the user, let them confirm or correct it, then write it and
     offer to commit it. **Propose the fast, broadly-applicable ones and say
     which CI-only candidates you left out** — the user decides, but they
     cannot correct an omission they were never shown. Future runs hit the
     fast path.
   - If the project genuinely has no build/test step (e.g. a pure-docs repo),
     declare whatever it does have. Don't invent checks.
   - A check whose verdict is its OUTPUT rather than its exit status — `gofmt
     -l` prints filenames and exits 0 — needs a command that converts one to
     the other, or a small script the `command` points at. Pointing at a script
     is the escape hatch for anything a one-liner would mangle.

   The **environmental-failure ledger** stays in `AGENTS.md` as prose: it is
   read by agents, not executed, and it is the one part of the old block that
   was never a command.

1a. **Resolve the model lineup.** Beads carry a capability `tier`, not a
   model name, so a backlog groomed months ago still routes correctly. The
   lineup lives in `sf.toml` beside the checkpoint declaration:

   ```toml
   [models]
   light = "<model>"
   standard = "<model>"
   advanced = "<model>"
   frontier = "<model>"
   ```

   Pre-flight emits each as a `BB_MODEL_*` fact. Tiers resolve independently,
   so a project may declare only the ones it uses.

   - **If it exists**, Phase 0 reads it. Fast path — no prompt.
   - **If it's missing**, propose one from the models available in this
     harness, let the user confirm, then write it (same mechanics as the
     checks declaration above).
   - **If the user declines**, dispatch without the `model` param — every
     subagent inherits the session model. Pre-flight reports
     `BB_LINEUP=undeclared` and says so; state it once at run start so the
     cost profile isn't a surprise. An undeclared tier is never guessed.

   `[run]` in the same file carries project-level run defaults — `push_mode`,
   and `run_root` when the default run root is wrong for this project. Omit a
   key to leave that decision to the run.

1b. **Resolve the code reviewer.** A sprint reviews its own diff before the
   human does, and it runs whatever the project declares:

   ```toml
   [review]
   skill = "code-review"     # dispatch a skill by name
   # command = 'my-linter --diff'   # or run a command; stdout is the report
   ```

   Exactly one of `skill` or `command`. **We ship no reviewer** — harnesses
   provide one and the ecosystem is full of them. What the sprint owns is what
   happens to the findings: scoping them to this run's diff, filing each as a
   bead, classifying in-scope against the authorization gate, and feeding the
   review loop.

   - **If it exists**, the review step runs it.
   - **If it's missing**, the run **skips the review step and says so in its
     report**. This is not an error and does not block the run. Do not prompt
     for it mid-run; propose it at setup like the others, or leave it out.

   A silently skipped review reads exactly like a clean one, which is why the
   skip is reported rather than defaulted away.

1c. **Create the issue outbox.** A run reports findings about the skills
   it used as well as about the project itself. The second kind cannot be
   filed here — this project does not own those skills — so they are
   written to a directory its owner forwards:

   ```
   <project>/.sf/outbox/<skill>/<YYYY-MM-DD>-<slug>.md
   ```

   One file per issue, in the issue template's format. **Nothing about a
   skill this project does not own is ever filed in this project's own
   tracker**, under any label: a backlog of proposals its owners cannot
   act on teaches them to ignore the channel, and that is the failure the
   whole split exists to prevent.

   - **Track the directory, do not gitignore it.** The issues are the
     deliverable of a run and should survive a reclaimed container. An
     ignored outbox loses them silently.
   - **The outbox is the project's.** Emptying it, forwarding it, or
     ignoring it entirely are all its owners' call. A run appends; it
     never prunes.
   - **Where this repository vends the skills it runs**, the outbox is
     swept into its own tracker in the same run — it is its own
     downstream.

2. **Common command permissions.** Confirm these are in
   `.claude/settings.local.json` allow list. If any are missing, ask the user
   once whether to add them, then add (don't add silently):
   - `Bash(bd:*)` — subsumes `bd create`, `bd update`, `bd close`, and every
     other `bd` subcommand, so list it alone.
   - `Bash(wait)`, `Bash(wait %1)`, `Bash(jq:*)`
   - This skill's bundled scripts — one entry per script covers every
     skip check and close ceremony in the run:
     `Bash(<skill-dir>/scripts/pre-flight-check.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-status.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-verify-skip.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-close-bead.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-checks.py:*)`,
     `Bash(<skill-dir>/scripts/sprint-cleanup.sh:*)`
   - Whatever the declared `command`s invoke underneath (e.g. `Bash(cargo:*)`,
     `Bash(npm run test:*)`, `Bash(go build:*)`, `Bash(pytest:*)`) — the runner
     is one entry, but each check is its own subprocess
   - `Read(/tmp/claude/**)` — and `Read($TMPDIR/**)` too WHEN `TMPDIR` is set
     (macOS sets it to a per-user dir; most Linux hosts leave it unset)