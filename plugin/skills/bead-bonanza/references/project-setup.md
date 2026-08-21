# bead-bonanza Project Setup

What a project must provide for bead-bonanza to run: an executable `checks`
script, a model lineup in `sf.toml`, and a few command permissions. **Read
this only when one is missing** — once they exist, Phase 0 loads them and never
comes back here.

## Contents

- The `checks` script — how this project validates a bead
- Model lineup — how capability tiers resolve to models (`sf.toml` `[models]`)
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

   Phase 0 resolves it in this order, first hit wins, and blocks the run when
   none is found:

   | | |
   |---|---|
   | `$BB_CHECKS` | explicit override |
   | `sf.toml` | `[checkpoints]` `checks = "<repo-root-relative path>"` |
   | `scripts/checks` | the conventional location, when `sf.toml` declares none |

   `sf.toml` is a plugin-scoped config file at the repository root. Reading one
   named key out of it is not the thing this replaced: the retired mechanism
   parsed *prose* for a command, where this reads a config file for a path.

   It must be **executable**. A present-but-not-executable file is treated as a
   hard error rather than as absent, because that is a mistake to report, not a
   state to work around.

   **The contract, in full: run it.** Exit 0 means every check passed;
   anything else means something failed, and the output says what. There is no
   required output format, no ids, and nothing to enumerate.

   Phase 0 captures what it prints on the untouched tree as the **checks
   baseline**, and a bead compares its own run against that text — a line in
   both was already failing, a line only in the bead's is the bead's. That
   comparison never needs a name for anything.

   **Run every check, even after one fails.** Chaining with `&&` stops at the
   first failure and hides the rest, so a run sees one problem at a time and a
   known-red check masks everything behind it.

   - **If the script exists**, Phase 0 runs it. Fast path — no prompt.
   - **If it's missing**, discover candidate commands from `AGENTS.md` /
     `CLAUDE.md` (build/test/lint/format sections, CI config), propose a draft
     script to the user, let them confirm or correct it, then write it and
     offer to commit it. **Propose the fast, broadly-applicable ones and say
     which CI-only candidates you left out** — the user decides, but they
     cannot correct an omission they were never shown. Future runs hit the
     fast path.
   - If the project genuinely has no build/test step (e.g. a pure-docs repo),
     the script still exists and runs whatever it does have. Don't invent
     checks.

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
     checks script above).
   - **If the user declines**, dispatch without the `model` param — every
     subagent inherits the session model. Pre-flight reports
     `BB_LINEUP=undeclared` and says so; state it once at run start so the
     cost profile isn't a surprise. An undeclared tier is never guessed.

   `[run]` in the same file carries project-level run defaults — `push_mode`,
   and `run_root` when the default run root is wrong for this project. Omit a
   key to leave that decision to the run.

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
     `Bash(<skill-dir>/scripts/bb-close-bead.sh:*)`
   - The `checks` script's path (e.g. `Bash(scripts/checks:*)`), plus whatever
     the checks it runs invoke underneath (e.g. `Bash(cargo:*)`,
     `Bash(npm run test:*)`, `Bash(go build:*)`, `Bash(pytest:*)`)
   - `Read(/tmp/claude/**)` — and `Read($TMPDIR/**)` too WHEN `TMPDIR` is set
     (macOS sets it to a per-user dir; most Linux hosts leave it unset)