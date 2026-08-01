---
name: triage-rfc
description: Triage improvement RFCs received from projects using the
  spec-driven-dev plugin — verify their evidence against current
  versions, dedupe against prior resolutions and deferred beads,
  classify each proposal (adopt-now, defer, reject), and implement
  adopted changes after human approval. Use when asked to triage the
  RFC inbox, process rfcs/inbox/, review a received RFC, or check on
  deferred RFC themes.
version: "1.3"
---

# RFC Triage

**Role:** this repo is the upstream of every project using the
spec-driven-dev plugin. Downstream retros (the `retro-spec` skill)
deposit `audience: skill-authors` RFCs and retro notes into
`rfcs/inbox/<skill-name>/`, and `retro-spec` deposits its
`audience: project` RFCs into `rfcs/inbox/<spec-type>/`. This skill turns that inbox into resolutions: changes
to the skills, templates, and standards; beads for deferred themes;
and a complete retro log. It is human-gated throughout — the skill
recommends and, only after approval, implements.

**Cadence — hybrid:** obvious, cheap, well-evidenced proposals are
adopted immediately; plausible-but-unproven or expensive ones are
deferred as beads so recurrence across projects can surface priority.
Rejection is a first-class outcome and is recorded so the same idea is
not relitigated without new evidence.

## Inputs

- `rfcs/inbox/**/*.md` — received RFCs (`rfc-<project>-<slug>.md`) and
  retro notes (`retro-note-<project>-<date>.md`), filed in a
  subdirectory naming their subject: `<skill-name>/` for items from
  `retro-skill`, `<spec-type>/` for items from `retro-spec`. Read the
  whole tree — a glob that stops at the top level finds nothing.
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
- [ ] 3. Every proposal classified (adopt-now / defer / reject /
      already-implemented)
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

   WHEN the RFC's `origin` is `self` AND the model triaging it is the
   model named in its `model` field, the Resolution SHALL state that
   the RFC was authored and triaged by the same agent, and name any
   proposal that ratifies a decision its own author made. This does
   not disqualify the proposal — self-retro is how this repository
   improves — but an approval recorded without the disclosure reads as
   independent corroboration when there was none.
2. **Dedupe.** Search this repo's `rfcs/` for prior proposals on the
   same theme, and the beads:

   ```
   bd list --label rfc-triage --all
   ```

   The `--all` is load-bearing. `bd list` defaults to open work and
   hides `closed` and `deferred` issues, which are exactly the two
   states a dedupe must see: a closed bead is a prior ADOPTED or
   REJECTED outcome, and a deferred bead is the accumulation signal.
   Without it the command returns only open beads and the prior-adopted
   and prior-rejected branches below can never fire. (WHERE the
   installed `bd` has no `--all`, use the explicit form
   `--status open,in_progress,blocked,deferred,closed`; note that
   repeating `--status` overwrites rather than accumulates.)

   Match on the outcome section of each resolved RFC whatever it is
   titled — older RFCs may use a different heading than the current
   template's, so search the content rather than assuming one heading.
   - Prior REJECTED match: flag it with the prior rationale; recommend
     reconsideration only if this RFC carries new evidence.
   - Prior DEFERRED match (a deferred bead): this is recurrence — the
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
   - **already-implemented** — the proposal is sound and its change is
     already present in the tree, landed by earlier work rather than by
     this triage. Distinct from `reject — already addressed`, which is
     for a proposal whose premise no longer holds: this one was right
     and nothing about it is being turned down. Record it in the
     Resolution as "adopted — already implemented", name where the
     change lives, and bump nothing.
4. **Present the recommendation and STOP for approval.** One table:
   proposal, classification, one-line rationale, files touched.
   Implement nothing before the human approves; the human may
   reclassify any row.

   **WHEN no human can answer** — an unattended, scheduled, or
   batch run — the gate does NOT dissolve, and absence is NOT
   approval. Take the unattended branch instead:
   - Classify every proposal as usual, and present the table. The
     analysis is the valuable part and it does not require a human.
   - **File one bead per proposal, whatever its classification** —
     adopt-now included — each meeting the bar in step 5, labelled
     `rfc-triage` and `theme:<key>`, with the recommended
     classification recorded in the bead. The beads ARE the handoff
     artifact: they survive the session, and a human can approve
     later by working the queue.
   - **Implement nothing.** Not the obvious ones, not the one-line
     ones.
   - **Leave the RFCs in `rfcs/inbox/`.** Do not write Resolutions,
     do not renumber, do not retire. The Resolution section records
     what a human approved; writing it here would be a record of an
     approval that never happened.
   - **Say in the report which steps the gate blocked** — namely 5
     through 7 — and name the beads filed, so the next attended
     session resumes from them rather than re-triaging the inbox.

   Do not resolve the ambiguity any other way. The three tempting
   alternatives — discarding the analysis, writing into the
   Resolution section, or reading an absent human as consent — lose
   work, falsify the record, and defeat the gate respectively.
5. **Implement** the approved adopt-now items:
   - **Sequence before you edit.** Count the approved proposals and the
     files they touch. WHEN more than one approved proposal changes the
     same file, or more than about five are approved in total, do not
     implement inline: file one bead per approved proposal (to the same
     bar as a deferral bead, below) and run `groom-beads` over them to
     wire the dependency graph. Two proposals editing one file SHALL be
     serialized by a real edge; the proposal that owns that file's
     version bump SHALL be last in its chain. Implement in the
     resulting wave order.

     The collision key is the FILE, not the RFC — proposals from
     different RFCs routinely land in the same file, and two beads with
     no dependency between them are still a collision when they edit
     the same lines.

     WHERE the approved set is small and touches disjoint files,
     implement inline and say so; grooming a two-file change is
     overhead.

     `bead-bonanza` MAY then execute the graph when the queue is large
     enough to warrant orchestration, but this step does not require
     it — the sequencing is the point, not the dispatch.
   - Apply the changes; keep every touched file within its own
     conventions (80-column lines, EARS for normative statements).
   - Bump the version of every changed template (instruction-block
     `Version:`), skill (frontmatter `version`), and standard
     (frontmatter `version`) per [[template-authoring-standards]].
     **Bump each file exactly once for the whole triage, no matter how
     many proposals changed it** — the version announces the release,
     not the proposal count. WHEN several proposals touch one file, the
     bump belongs to whichever lands last; verify the bump after the
     final change to that file, never before.
   - File one bead per approved deferral **that meets the Definition of
     Ready**. The bar is the spec-owned
     [[work-order-definition-of-ready]] — do not restate or fork it
     here; run the bead against it, and use `groom-beads` when a bead
     needs to be brought up to it. Deferred work is read months later by
     an agent with no memory of this triage, so the fresh-agent test
     (DoR check 1) is the one that actually bites: a bead saying "adopt
     P4 from the RFC" fails it, because the reader has neither the RFC
     open nor the session that produced it.

     Beyond the DoR, each bead SHALL carry: the RFC path and proposal ID
     (so the testimony is recoverable), the evidence summary inlined
     rather than linked, and labels `rfc-triage` plus `theme:<key>`.
     Reuse an existing theme key WHERE one fits. Enumerate them before
     minting a new one:

     ```
     bd label list-all | grep 'theme:'
     ```

     The issue count beside each key is the recurrence signal itself —
     a theme reaching two independent projects is the promotion trigger
     in step 2 — so read the counts, not just the names. Keep themes
     kebab-case topics, e.g. `theme:ac-coverage`,
     `theme:reference-syntax`. A bead without a theme label is
     invisible to the accumulation model.

     A deferral is not dispatchable work: set the bead's status to
     `deferred` so it does not surface in `bd ready`, and record why in
     `grooming_status` (`blocked-needs-info` when a specific fact is
     missing, `blocked-ungroomable` when it needs redesign).
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
6. **Log.** Append one row per ingested item to `rfcs/retro-log.md`.
   WHEN the file does not exist, create it with this header and no
   other content; do not restyle an existing log to match:

   ```markdown
   # Retro Log

   | Date | Origin | Retroed | Item | Outcome |
   |------|--------|---------|------|---------|
   ```

   One row per item, newest at the bottom: date (`YYYY-MM-DD`), origin
   project (`self` for this repo), the skill or phase retroed, the item
   (`RFC <nnnn>` or `retro note`), and the outcome as per-proposal
   counts (`N adopted, N deferred, N rejected`) or `null retro`. Counts
   are per proposal, never per RFC. Any commentary goes below the table
   under `## Run notes`, never inside a row.
7. **Publish.** WHEN any skill, template, or standard **that ships
   inside the plugin** changed, bump the plugin version in
   `plugin/.claude-plugin/plugin.json` (MINOR for normative changes).
   A skill ships in the plugin IFF it has a source directory under
   `plugin/skills/`; a tracked skill living only under
   `.claude/skills/` (such as `triage-rfc` itself) is repo-local, and
   changing it does NOT on its own justify a bump, because installs
   never receive it. WHEN a triage changed only repo-local skills, say
   so and skip the bump. Distribution is plugin-only: installs pick
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
| Adopting a self-authored proposal without recording that it was self-authored | The Resolution is the durable record; an undisclosed self-adoption is indistinguishable later from an independently reviewed one |
