# Per-bead subagent prompt

**One bead, dispatched alone into its own worktree.** The canonical ceremony —
the fast-lane and shared-surface batch prompts reference this workflow rather
than restating it, so read this one first.

Fill every `{PLACEHOLDER}` from
[prompt-placeholders.md](prompt-placeholders.md); create and later remove the
worktree per [worktree-mechanics.md](worktree-mechanics.md). Everything inside
the fence below is what you send.

```
You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. Validation comes from the checks the project declares in `sf.toml`.

Bead: {ID} (P{PRIORITY}) — {TITLE}
Run epic: {EPIC}            (record incidents here as well as on the bead)
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
    (Pass the ids covering what you touched while iterating; run it with NO
     arguments before your close. Exit 0 is the bar.)
Push mode: {push_per_bead | defer_push}
Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)
Scripts: {SCRIPTS}          (bb-verify-skip.sh, bb-close-bead.sh)
Trailers: {"on the run epic" if any}   (pass `--trailers-from {EPIC}` to bb-close-bead.sh; omit it when the run resolved none)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

Before your first action, record that you hold this bead:

```sh
bd update {ID} --set-metadata dispatch_id={DISPATCH_ID} \
  --set-metadata dispatch_state=running \
  --set-metadata worktree={WORKTREE} \
  --set-metadata dispatched_at="$(date -u +%FT%TZ)"
```

and as your last action, whatever the outcome:

```sh
bd update {ID} --set-metadata dispatch_state=returned \
  --set-metadata terminal_state=completed|blocked|skipped
```

That pair is what lets an orchestrator tell a live agent from a dead one
without guessing at your worktree. Write the second even when you are blocked —
especially then.

**Then invoke the `run-bead` skill.** It carries the ceremony — claim, read,
implement in scope, gate, caveat, findings, close-and-commit, push, return —
together with the blocked and skipped paths and its own acceptance criteria.
The template above supplies its inputs; it supplies the procedure.
```
