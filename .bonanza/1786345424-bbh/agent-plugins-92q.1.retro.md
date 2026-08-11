`plugin/skills/bead-bonanza/SKILL.md` is 45,983 bytes / 869 lines. Claude Code
re-injects an invoked skill's body after compaction **capped at 5,000 tokens
per skill, truncating from the START of the file**
(`code.claude.com/docs/en/context-window`). Measured:

```
$ head -c 20000 plugin/skills/bead-bonanza/SKILL.md | wc -l
406            # line 406 == "### Choose a dispatch mode per bucket"
```

Everything after line 406 is absent from a compacted orchestrator's context,
including the three rules whose loss is most expensive:

- `SKILL.md:624-627` — "**Read that file at dispatch** and fill in every
  `{PLACEHOLDER}` before sending. Do not paraphrase a template from memory"
- `SKILL.md:616-618` — the retro anchor "is non-negotiable"
- `SKILL.md:833-837` — the three published script interfaces, published
  precisely because the scripts are declared "executed, not read"

`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=80` is set in this cloud environment
(verified in-session: `echo $CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` -> `80`), so
compaction fires whether or not the orchestrator plans for it.

In run `1786121603-ri` the three beads dispatched after the run's context loss
— `9kb.5`, `9kb.6`, `9kb.7` — are exactly the three that closed with no
progress ledger and no retro anchor, because their briefs were hand-written.
The rule forbidding hand-written briefs had been truncated away.

This bead is the cheap half of the fix and ships on its own, ahead of the
restructuring: a paragraph placed where truncation cannot reach it, telling a
compacted orchestrator that it has lost the procedure and must re-read it.
The restructuring that removes the need for this paragraph is tracked
separately; the paragraph should stay afterwards regardless, because a file
that fits today can grow again.


## Retro — implementation

**Findings:** None. The bead's Design section supplied exact wording and insertion point. No friction encountered.

**What you tried that did not work:** N/A.

**What you decided silently:** N/A.

**Pre-fix failures:** Not a defect bead.
