---
name: triage-rfc
description: Triage improvement RFCs received from projects using the
  spec-driven-dev plugin — verify their evidence against current
  versions, dedupe against prior resolutions and deferred beads,
  classify each proposal (adopt-now, defer, reject), and implement
  adopted changes after human approval. Use when asked to triage the
  RFC inbox, process rfcs/inbox/, review a received RFC, or check on
  deferred RFC themes.
version: "1.0"
---

# RFC Triage

**Role:** this repo is the upstream of every project using the
spec-driven-dev plugin. Downstream retros (the `retro-spec` skill)
deposit `audience: skill-authors` RFCs and retro notes into
`rfcs/inbox/`. This skill turns that inbox into resolutions: changes
to the skills, templates, and standards; beads for deferred themes;
and a complete retro log. It is human-gated throughout — the skill
recommends and, only after approval, implements.

**Cadence — hybrid:** obvious, cheap, well-evidenced proposals are
adopted immediately; plausible-but-unproven or expensive ones are
deferred as beads so recurrence across projects can surface priority.
Rejection is a first-class outcome and is recorded so the same idea is
not relitigated without new evidence.

## Inputs

- `rfcs/inbox/*.md` — received RFCs (`rfc-<project>-<slug>.md`) and
  retro notes (`retro-note-<project>-<date>.md`)
- `rfcs/*.md` — this repo's resolved RFC history (dedupe source)
- The beads database — deferred proposals (`bd` CLI; label
  `rfc-triage`). IF `.beads/` does not exist yet, THEN run `bd init`
  on first use.

## Process

Copy this checklist into your response and check items off as you go.
(A checklist copied into your response is not a task tool — WHERE the project bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress visible.)

```
Triage progress:
- [ ] Retro notes appended to the log and deleted
- [ ] 1. Provenance checked against current versions (per RFC)
- [ ] 2. Deduped against resolved RFCs and rfc-triage beads
- [ ] 3. Every proposal classified (adopt-now / defer / reject)
- [ ] 4. Recommendation table presented — STOPPED for human approval
- [ ] 5. Approved items implemented, verified, versions bumped
- [ ] 6. Retro log appended
- [ ] 7. Plugin version bumped IF any skill/template/standard changed
```

Step 4 is a hard stop. Do not tick 5 before the human has answered.


Retro notes first: append each to the retro log (step 6), then delete
the note. Then, per RFC, in received order:

1. **Provenance check.** Read the RFC's Provenance section. For every
   proposal, compare the cited skill/template/standard versions
   against the current ones. WHEN the cited source has changed since,
   verify the quoted text still exists — a proposal against text that
   is already fixed or removed is resolved "rejected — already
   addressed in <file> v<X.Y>" without further debate. An RFC missing
   provenance versions is still triaged, but say so in the
   Resolution; the citation burden was the author's.
2. **Dedupe.** Search this repo's `rfcs/` for prior proposals on the
   same theme, and the beads (`bd list --label rfc-triage`). Match on
   the outcome section of each resolved RFC whatever it is titled —
   older RFCs may use a different heading than the current template's,
   so search the content rather than assuming one heading.
   - Prior REJECTED match: flag it with the prior rationale; recommend
     reconsideration only if this RFC carries new evidence.
   - Prior DEFERRED match (an open bead): this is recurrence — the
     accumulation signal working. Link the new RFC in the bead,
     bump its priority, and recommend promotion to adopt-now when a
     second independent project has hit the same theme.
   - Prior ADOPTED match: verify the adoption actually covers this
     evidence; if it does, reject as already addressed.
3. **Classify each proposal** (per proposal, not per RFC):
   - **adopt-now** — concrete change, verbatim evidence, small blast
     radius, no conflict with the framework's design principles
     (quality-firewall strictness, small taxonomy, no duplication,
     backward compatibility).
   - **defer** — plausible but wants corroboration from another
     project, or the change is large enough that one session's
     evidence should not drive it alone.
   - **reject** — bar not met, conflicts with design principles,
     already addressed, or fixes carelessness rather than guidance.
4. **Present the recommendation and STOP for approval.** One table:
   proposal, classification, one-line rationale, files touched.
   Implement nothing before the human approves; the human may
   reclassify any row.
5. **Implement** the approved adopt-now items:
   - Apply the changes; keep every touched file within its own
     conventions (80-column lines, EARS for normative statements).
   - Bump the version of every changed template (instruction-block
     `Version:`), skill (frontmatter `version`), and standard
     (frontmatter `version`) per [[template-authoring-standards]].
   - File one bead per approved deferral: title from the proposal,
     description linking the RFC path, proposal ID, and evidence
     summary; labels `rfc-triage` and `theme:<key>`. Reuse an
     existing theme key WHERE one fits (check current labels before
     minting); themes are kebab-case topics, e.g.
     `theme:ac-coverage`, `theme:reference-syntax`.
   - **Verify before moving on.** Re-read each file you changed: the
     change is present, its version bumped, and its own conventions held.
     A half-applied adoption is worse than a deferral, because the
     Resolution will claim it landed.
   - Write the RFC's Resolution section: per-proposal outcome
     (adopted / adopted-with-refinement + the refinement / deferred +
     bead id / rejected + rationale), date, files changed, version
     bumps.
   - Number and move the RFC: rename to
     `rfcs/rfc-<nnnn>-<short-description>.md` using this repo's next
     free number, update its slug to match, and set status `retired`
     (the Resolution is the record; deferred work lives in beads).
6. **Log.** Append one line per ingested item to `rfcs/retro-log.md`:
   date, origin project, phase retroed, item (RFC number or "retro
   note"), outcome (counts adopted/deferred/rejected, or "null
   retro").
7. **Publish.** WHEN any skill, template, or standard changed, bump
   the plugin version in `plugin/.claude-plugin/plugin.json` (MINOR
   for normative changes). Distribution is plugin-only: installs pick
   up the change on plugin update once the bump lands on the
   marketplace (repo-root `.claude-plugin/marketplace.json`); a
   session run with `claude --plugin-dir ./plugin` sees it after
   `/reload-plugins`.

## Anti-Patterns

| Pattern | Problem |
|---------|---------|
| Implementing before the human approves the classification table | The gate exists because adoption changes every downstream project |
| Triaging per-RFC instead of per-proposal | RFCs bundle proposals of very different quality; the grain of decision is the proposal |
| Adopting on plausibility without verifying the quoted evidence | The whole pipeline's currency is verified evidence; spending it on taste devalues every future RFC |
| Silent rejection (no Resolution entry) | The next retro re-derives and re-proposes the same idea |
| Deferring without a theme label | Recurrence across projects becomes uncountable — the accumulation model stops working |
| Forgetting version bumps on changed files | Future RFCs cite versions; stale versions break the "already fixed?" check |
| Editing a received RFC's Evidence or Proposals | The RFC is the reporter's testimony; triage writes only the Resolution |
