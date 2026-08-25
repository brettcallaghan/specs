# Project Instructions for AI Agents

@AGENTS.md

<!--
Everything above this line is AGENTS.md, imported so the two files cannot
drift. Claude Code reads CLAUDE.md and NOT AGENTS.md, which is why the import
exists; the import is the mechanism the Claude Code memory docs prescribe for
exactly this case. Note it does not save context — an imported file is loaded
at launch just as inlined content would be. What it buys is one copy.

Add ONLY Claude-specific instructions below. Anything true for every agent
belongs in AGENTS.md.
-->

## Session bootstrap — check this first (Claude Code)

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
