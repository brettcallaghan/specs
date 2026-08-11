Measured across run 1786121603-ri by tracing every quoted span >=40 chars in
`rfcs/inbox/feature-blueprint/rfc-0015-implementing-reference-integrity.md`
back to its source:

| Channel | Generated volume | Yield |
|---|---|---|
| 10 `.retro.md` files | 137,146 bytes (~34k output tokens) | 1 item, in *Considered and Not Proposed*. Zero proposals. |
| 34 `retro:` bd comments | 16,411 bytes (~4.1k output tokens) | 12 of 12 traceable Evidence quotes; 5 of 7 proposals |

**8.4x the cost for ~1/12 the findings.** `agent-plugins-9kb.6` — no retro
file at all, one 482-character comment — carries proposal P2. Conversely
`tix`, `ml9` and `j7f` wrote 38,205 bytes of retro files between them and are
cited zero times.

Three structural causes:

1. **The template asks for categories, not triggers.** `findings-capture.md`
   defines a trigger list (record WHEN you needed information the bead did not
   carry, WHEN the bead contradicted repo reality...), which filters.
   `references/subagent-prompts.md:82` asks for a category ("implicit decisions
   you made"), which enumerates. Implicit decisions are the largest bucket
   (32,892 bytes, 39% of retro prose) and produced zero RFC quotes on their own.

2. **The skill does not own the channel that works.** `grep -c
   "findings-capture\|retro:" references/subagent-prompts.md` returns **0**. The
   `retro:` comment instruction reached subagents only because the groomed
   `9kb.*` bead descriptions carried it themselves. `tix`, `ml9` and `j7f` were
   prerequisites outside that groomed plan, which is exactly why they produced
   files and no findings.

3. **The ceremony destroys the instruction.** `bb-close-bead.sh:58` runs `bd
   update --status closed --body-file <retro>`, and `--body-file` REPLACES the
   description. Agents hand-retype the description into the retro file and drop
   content. Verified on `agent-plugins-9kb.1`: the original description (from
   `git show f019922^:.beads/issues.jsonl`) ends with a `## Retro` section
   containing "one `bd comment` per finding, prefixed `retro:` — following the
   findings capture protocol in `rfcs/findings-capture.md`". The current
   description has a `## Retro` heading but no `findings-capture` and no
   `prefixed`. The collision is structural: the bead already has a `## Retro`
   section and the instruction says to append one with the same name.

Standing cost: 79% of description bytes in the committed `.beads/issues.jsonl`
are retro payload, rewritten and re-committed on every close.


## Retro — implementation

### Findings
- AC 8 ("measured retro output on the next run is materially smaller than the ~34k-token baseline") cannot be satisfied by the bead that changes the template — it is an observation about a later run. Closest available evidence, from this batch, which the orchestrator already dispatched under the new shape: 2 retro files totalling 8,552 bytes, ~4,276 bytes/bead against the baseline's 137,146/10 = 13,715 bytes/bead, a 3.2x reduction. That is a proxy, not the AC.
- The template's own prose tripped `check-skills.sh` rule 2: writing `` `bb-close-bead.sh` `` bare in backticks is read as a path that must resolve relative to the citing file, and it errored on both the plugin copy and its `.claude/skills` symlink. The file's existing convention is `{SCRIPTS}/bb-close-bead.sh`, which the regex skips because of the brace. Nothing in the bead or the file says so.
- Step 3 of the same template still says "For any implicit decision you must make, record what you chose (and why) — you'll put it in the retro", which is the broad category framing that the bead's cause 1 argues against. The bead scopes the change to step 5, so it is left as-is and the two now differ in width.

### What you tried that did not work
- Nothing abandoned on the design. The one reversal was mechanical: the first draft's bare backticked script name, replaced with the `{SCRIPTS}/`-prefixed form after the gate rejected it.

### What you decided silently
- The `retro:` comment obligation is stated inside finding item 1 rather than as its own bullet, so the instruction sits where the finding is written and cannot be read as an optional extra step.
- Added a clause naming WHY the comment channel matters ("what downstream retros and RFCs actually read; the file alone does not reach them"). The bead only required the obligation; without the reason an agent under time pressure treats it as duplication of the file it just wrote.
- Item 3 is given a worked contrast (helper name / guard-clause order vs. the approach the bead specified) rather than only the word "narrower" — the bead's measured complaint is that this bucket over-enumerates, and a boundary without an example does not move it.
- Kept the file's existing very-long-line paragraph style rather than wrapping at 80; wrapping only this step would make the diff read as a reflow.

### Observed pre-fix failure
Not a defect bead. The pre-change state is the measurement in the Description, reconfirmed here: `grep -c "findings-capture\|retro:" plugin/skills/bead-bonanza/references/subagent-prompts.md` returned `0` before the change and `1` after.
