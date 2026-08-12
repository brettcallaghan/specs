# bead-bonanza Project Setup

The two blocks bead-bonanza reads from the project's `AGENTS.md`, and the
permissions a run needs. **Read this only when a block is missing** — once
both exist, Phase 0 loads them and never comes back here.

## Contents

- Gate profile — how this project validates a bead (`## bead-bonanza gates`)
- Model lineup — how capability tiers resolve to models (`## model lineup`)
- Command permissions a run needs

Both blocks are written into `AGENTS.md` so the next run hits the fast path.
Propose a draft, let the user confirm or correct it, then write it — never
write either block silently.

---


1. **Resolve the gate profile.** bead-bonanza learns how to validate a bead from
   a **gate profile** the project keeps in `AGENTS.md` under a `## bead-bonanza
   gates` heading. Read it first — every `<format>`, `<lint>`, `<build>`,
   `<test>`, and `<regen>` reference below resolves to a command from that
   block.
   - **If the block exists**, load it. Fast path — no prompt.
   - **If it's missing**, discover candidate commands from the rest of
     `AGENTS.md`/`CLAUDE.md` (build/test/lint/format sections, CI config),
     propose a draft profile to the user, let them confirm or correct it, then
     **write it into `AGENTS.md`** under `## bead-bonanza gates` (create
     `AGENTS.md` if neither doc exists) and offer to commit it. Future runs then
     hit the fast path.
   - The block is human-readable markdown: the core gate commands, an optional
     pre-test `<regen>` step (codegen/snapshot regeneration — many projects have
     none), a targeted-tests-by-surface list, and the environmental-failure
     ledger. Template:
     ```markdown
     ## bead-bonanza gates

     - **Format:** `<cmd>`
     - **Lint:** `<cmd>`
     - **Build:** `<cmd>`
     - **Test (full suite):** `<cmd>`
     - **Regenerate before tests** (codegen/snapshots; omit if the project has none): `<cmd>`

     ### Custom gates (optional)
     Any extra checks to run as part of the gate suite — a plain command, or an inline script the user describes:
     ```bash
     # e.g. a project-specific check the standard gates don't cover
     ./scripts/check-migrations.sh
     ```

     ### Targeted tests by surface
     When only part of the tree changed, run the narrowest suite that covers it:
     - `<path or surface>` → `<targeted test cmd>`

     ### Environmental-failure ledger
     Suites that fail — or falsely PASS — only in the local sandbox, diagnosed once and passed into every subagent prompt as fact — never re-diagnosed per wave.
     - `<suite>`: fails in sandbox — `<symptom>` → `<workaround / covered-by-CI note>`
     - `<suite>`: falsely GREEN — e.g. runs a stale prebuilt binary a plain test command doesn't rebuild → force a clean `<build>` before trusting the result
     ```
   - If the project genuinely has no build/test step (e.g. a pure-docs repo),
     record that: `<test>` is a no-op and the doc/chore fast lane is the only
     dispatch mode. Don't invent gates.
1a. **Resolve the model lineup.** Beads carry a capability `tier`, not a
   model name. Read the project's `## model lineup` block in `AGENTS.md`
   to resolve tier → model at dispatch:

   ```markdown
   ## model lineup

   Least to most capable:

   - **light:** `<model>`
   - **standard:** `<model>`
   - **advanced:** `<model>`
   - **frontier:** `<model>`
   ```

   - **If the block exists**, load it. Fast path — no prompt.
   - **If it's missing**, propose one from the models available in this
     harness, let the user confirm, then write it into `AGENTS.md` (same
     mechanics as the gate profile above).
   - **If the user declines**, dispatch without the `model` param — every
     subagent inherits the session model. Say this once, at run start, so
     the cost profile isn't a surprise.

2. **Common command permissions.** Confirm these are in
   `.claude/settings.local.json` allow list. If any are missing, ask the user
   once whether to add them, then add (don't add silently):
   - `Bash(bd:*)`, `Bash(bd create:*)`, `Bash(bd update:*)`, `Bash(bd close:*)`
   - `Bash(wait)`, `Bash(wait %1)`, `Bash(jq:*)`
   - This skill's bundled scripts — one entry covers every progress event,
     skip check, and close ceremony in the run:
     `Bash(<skill-dir>/scripts/pre-flight-check.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-progress.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-verify-skip.sh:*)`,
     `Bash(<skill-dir>/scripts/bb-close-bead.sh:*)`
   - The gate-profile commands — whatever
     `<format>`/`<lint>`/`<build>`/`<test>`/`<regen>` resolve to for this
     project (e.g. `Bash(cargo:*)`, `Bash(npm run test:*)`, `Bash(go build:*)`,
     `Bash(pytest:*)`, or the regen script's path)
   - `Read(/tmp/claude/**)` — and `Read($TMPDIR/**)` too WHEN `TMPDIR` is set
     (macOS sets it to a per-user dir; most Linux hosts leave it unset)