---
title: "RFC: An unattended-run rule for the gate, and a dedupe that
  sees history"
description: The human gate has no rule for a session no human can
  answer, and the two bd commands step 2 and step 5 prescribe cannot
  return the data those steps need
slug: rfc-0002-triage-rfc-unattended-gate
type: rfc
status: retired
origin: self
skill: triage-rfc
skill_version: "1.0"
model: claude-opus-5
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.7.0"
---

# RFC: An unattended-run rule for the gate, and a dedupe that sees history

## Summary

`triage-rfc` v1.0 was run on a one-RFC inbox in a subagent session where
no human was reachable. Steps 1–3 executed cleanly and the provenance
check earned its keep — it caught a misattributed quote in the RFC under
triage. Step 4 then hard-stops for approval, and the skill has no rule
for a session that cannot supply one: everything after step 4 becomes
unreachable and the verified classification exists only in a transcript
that does not survive the session. Separately, both `bd` commands the
skill prescribes return output that cannot answer the question the
surrounding step asks, and the beads the skill mints are specified in
three fields of its own invention rather than against the framework's
own Definition of Ready. The thesis: `triage-rfc` assumes an attended
session with an interactive operator reading `bd` output, so the parts
it leaves unspecified are all handoffs — to the next session, to the
next triage, and to whoever eventually works a deferral.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code, remote (web) session |
| Task the skill was applied to | Triaging `rfcs/inbox/rfc-self-spec-document-self-check.md` (4 proposals: 1 blocker, 2 snags, 1 polish) in a subagent with no reachable human |
| References / scripts / templates invoked | `.claude/skills/triage-rfc/SKILL.md` v1.0 (the skill carries no reference files); `bd` CLI v1.1.2 (dev) |
| Run dates | 2026-07-30 |

## Evidence

1. **[gap · blocker]** No rule at `triage-rfc` step 4 for a session in
   which no human can answer. `.claude/skills/triage-rfc/SKILL.md:92-95`
   reads "4. **Present the recommendation and STOP for approval.** One
   table: proposal, classification, one-line rationale, files touched.
   Implement nothing before the human approves; the human may
   reclassify any row", reinforced at `SKILL.md:55`: "Step 4 is a hard
   stop. Do not tick 5 before the human has answered." Steps 5, 6 and 7
   are all gated behind that answer — implementation, the RFC's
   Resolution, the RFC's numbering and move, the retro log line, and
   the plugin version bump. In an unattended run none of them can
   execute, and the skill says nothing about what the run should leave
   behind instead. The agent had verified provenance for four proposals
   and classified each, and the skill offered no durable place to put
   that work: `SKILL.md:102-107` defines a bead only for an *approved
   deferral* ("File one bead per approved deferral"), which is a
   post-approval artifact. The run therefore had to invent a handoff —
   one bead per proposal carrying its recommended classification — on
   instruction from its caller, not from the skill. Absent that
   instruction the agent would have had to guess between at least three
   incompatible options: stop and report in prose (losing the work),
   write the classifications into the RFC's Resolution section (which
   `SKILL.md:98` reserves for triage's post-approval record and
   `SKILL.md:107-110` treats as a claim that the change landed), or
   treat "no human reachable" as implicit approval and implement.

2. **[script-defect · snag]** `bd list --label rfc-triage` — the dedupe
   command at `SKILL.md:70` — omits closed issues, so a theme that was
   deferred and later resolved is invisible to step 2. Observed: after
   closing a labelled bead, the command returns only the four open
   beads and reports "Total: 4 issues". Expected for a dedupe pass: the
   closed bead too, since `SKILL.md:79-81` requires exactly that case
   to be detected — "Prior ADOPTED match: verify the adoption actually
   covers this evidence; if it does, reject as already addressed."
   Minimal repro:

   ```
   bd create "TEMP dedupe probe" --type=task -p 3 \
     -l rfc-triage,theme:probe-only
   bd close <id>
   bd list --label rfc-triage                  # probe absent
   bd list --label rfc-triage --status closed  # probe present
   ```

3. **[script-defect · snag]** The same command prints no labels, so
   step 5's instruction to reuse theme keys has no command that can
   satisfy it. `SKILL.md:104-107` reads "labels `rfc-triage` and
   `theme:<key>`. Reuse an existing theme key WHERE one fits (check
   current labels before minting); themes are kebab-case topics".
   Observed: `bd list --label rfc-triage` prints id, status glyph,
   priority and title only — no label column — so the only `bd`
   invocation the skill names cannot show which theme keys are already
   in use. Expected: an enumeration of theme keys. Minimal repro:

   ```
   bd create "TEMP" --type=task -p 3 -l rfc-triage,theme:probe-only
   bd list --label rfc-triage   # no theme:* string anywhere in output
   bd label list-all            # theme:probe-only listed, with a count
   ```

   `bd label list-all` is the command that answers the step, and it
   counts closed beads too:

   ```
   $ bd label list-all
   🏷 All labels (6 unique):
     rfc-triage                       (5 issues)
     theme:document-conformance-gate  (1 issues)
     ...
   ```

   The run minted four theme keys against an empty database, so the
   gap cost nothing this time; it bites on the first triage that
   follows a deferral.

4. **[gap · snag]** No rule at `triage-rfc` step 5 for bringing a filed
   bead to the framework's own Definition of Ready. `SKILL.md:102-104`
   prescribes the entire bead in one clause — "File one bead per
   approved deferral: title from the proposal, description linking the
   RFC path, proposal ID, and evidence summary" — while
   `plugin/skills/spec/references/work-order-definition-of-ready.md`
   v1.1 states "This is the single definition of 'ready to dispatch.'
   The `spec` skill owns it; the grooming skills enforce it
   (`groom-beads` on any unit of work, `groom-feature-blueprint` on
   blueprint-decomposed work)". `triage-rfc` v1.0 contains no reference
   to that standard, to `groom-beads`, or to `bead-bonanza` — verified
   by `grep -n "groom-beads\|bead-bonanza\|Definition of Ready"
   .claude/skills/triage-rfc/SKILL.md`, which returns nothing. Observed
   in this run: filing four beads, the agent had to decide the
   description schema itself and wrote well past the three fields
   `SKILL.md:102-104` lists, adding the concrete change proposed, the
   files each touches, the version bumps implied, and an explicit
   "not implemented" marker. A bead filed to the literal schema would
   fail DoR check 1 — "A work order is **ready** when a fresh agent with
   zero session context can complete it correctly from the work order
   alone" — because an RFC path plus an evidence summary states the
   grievance without stating the job.

## Proposals

### P1 — Add an unattended-run rule to step 4 [blocker]

**Problem.** `SKILL.md:92-95` and `SKILL.md:55` make step 4 a hard stop
on a human answer, and no step covers a session that cannot produce
one. Finding 1: steps 5–7 become unreachable, and the skill defines no
artifact for a *recommended* classification — `SKILL.md:102` scopes the
bead to an "approved deferral". A verified triage then either evaporates
or gets written somewhere that misrepresents it.

**Change.** Append to step 4, after "the human may reclassify any row":

> **WHEN no human can answer in this session** (a subagent, batch, or
> headless invocation), the gate is not waived — it is deferred. Do not
> implement, do not write the Resolution, and do not number or move the
> RFC. Persist the recommendation instead, so it outlives the session:
> file one bead per proposal — every proposal, not only the deferrals —
> labelled `rfc-triage` and `theme:<key>`, whose description carries the
> RFC path, the proposal ID, the recommended classification, the
> evidence verified at step 1, and the version bumps the change would
> require. State in each that nothing was implemented. Leave the RFC in
> `rfcs/inbox/` with its Resolution empty, and report that steps 5–7 are
> blocked on approval, naming the beads. An attended run resumes from
> those beads rather than re-deriving the triage.

**Evidence.** Finding 1. Under the same model this converts the run's
forced three-way guess into a named path, and makes the unattended
result resumable instead of disposable — the classification survives in
the same store step 5 already uses for deferrals.

### P2 — Make the dedupe command see closed beads [snag]

**Problem.** Finding 2. `SKILL.md:70` names `bd list --label rfc-triage`
as the beads half of the dedupe, but `bd list` filters to open issues by
default. Step 2's "Prior ADOPTED match" and "Prior REJECTED match"
branches (`SKILL.md:75-81`) are precisely the cases whose beads are
closed, so the prescribed command cannot surface the matches the step
then tells the agent what to do about.

**Change.** Replace the parenthetical at `SKILL.md:70`:

> same theme, and the beads — `bd list --label rfc-triage` for open
> deferrals AND `bd list --label rfc-triage --status closed` for themes
> already resolved. `bd list` shows only open issues by default, so a
> deferral that was later adopted or rejected is invisible to the first
> form alone.

**Evidence.** Finding 2, with the repro above. The "already addressed"
check becomes executable rather than silently empty.

### P3 — Name the command that enumerates theme keys [snag]

**Problem.** Finding 3. `SKILL.md:105` says "check current labels before
minting" without naming a command, and the only `bd` command the skill
does name (`SKILL.md:70`) prints no labels. Theme keys are the
accumulation model's index — the skill's own anti-pattern table calls
deferring without one the thing that makes "recurrence across projects
uncountable" — so minting a near-duplicate key is a silent failure.

**Change.** Replace "(check current labels before minting)" at
`SKILL.md:105-106` with:

> (run `bd label list-all` before minting — it enumerates every theme
> key in use, closed beads included, which `bd list` does not show)

**Evidence.** Finding 3. One named command turns an unsatisfiable
instruction into a check, at no cost to the step's length.

### P4 — Bring step 5's beads up to the Definition of Ready [snag]

**Problem.** Finding 4. `SKILL.md:102-104` specifies the whole bead in
one clause — "File one bead per approved deferral: title from the
proposal, description linking the RFC path, proposal ID, and evidence
summary" — a bespoke schema that names no acceptance, no concrete
change, no files, and no tier. Meanwhile
`plugin/skills/spec/references/work-order-definition-of-ready.md` v1.1
declares itself "the single definition of 'ready to dispatch.' The
`spec` skill owns it; the grooming skills enforce it (`groom-beads` on
any unit of work". A deferral bead is a unit of work, and `triage-rfc`
mentions neither that standard nor `groom-beads` anywhere in the file.
So the skill that generates the framework's own backlog mints it
outside the framework's own gate — and does so at exactly the moment
the filing session's context is about to be lost.

**Change.** Replace the bead bullet at `SKILL.md:102-107` with:

> - File one bead per approved deferral, and take it to the framework's
>   own dispatch gate — [[work-order-definition-of-ready]] is the single
>   definition of ready for any unit of work, and the beads this skill
>   mints are no exception. Title from the proposal; description
>   carrying the RFC path, the proposal ID, the evidence summary, the
>   concrete change the proposal asks for, the files it touches, and the
>   version bumps it implies — enough that a fresh agent with no session
>   context can act on it (DoR check 1, the fresh-agent test). Labels
>   `rfc-triage` and `theme:<key>`. WHERE the deferral set is large, or
>   the descriptions came out thin, hand the beads to `groom-beads`
>   rather than grooming them by hand.

**Evidence.** Finding 4. The next run's deferrals survive the session
as dispatchable work rather than as pointers back to a transcript that
is gone. Ranked snag rather than blocker deliberately: the run
exercised the bead-filing half of step 5 but never its implementation
half, so the claim is that the skill was silent on a decision that
materially changes the outcome — not that it produced a wrong result.

## Considered and Not Proposed

- **The Role and Inputs sections name only `retro-spec` as the source
  of inbox items.** `SKILL.md:16` reads "Downstream retros (the
  `retro-spec` skill) deposit `audience: skill-authors` RFCs and retro
  notes into `rfcs/inbox/`", and `SKILL.md:31-32` describes received
  files as `rfc-<project>-<slug>.md`. The RFC actually triaged was
  produced by `retro-skill`, whose step 6 routes framework-repo retros
  "straight to `rfcs/inbox/`"; it carries no `audience` field and uses
  the skill-RFC frontmatter (`skill`, `skill_version`, `model`,
  `runs_observed`, `severity`) instead. The mismatch cost one
  verification hop to confirm the file was in scope, and produced no
  wrong action and no forced guess — below the bar. Worth folding into
  any later edit of that section.
- **Delegating step 5's implementation to `groom-beads` and
  `bead-bonanza`.** The broader form of finding 4: that step 5 applies
  changes, bumps versions and writes the Resolution in prose, when the
  plugin already ships a skill that brings work to a Definition of
  Ready and one that executes batches of beads — so the skill
  generating the framework's work does not use the framework. Not
  proposed, for two reasons. First, no run evidence: this run stopped
  at step 4 and never executed step 5's implementation half, so any
  claim about how that half should be carried out is design opinion,
  and the retro bar is explicit that a proposal must point at the step
  where the rule should live from something observed. Second, the
  separation is defensible on its own terms — step 5 runs immediately
  after a human has approved one specific table of changes, and
  interposing a grooming pass plus a batch-dispatch pass separates the
  approval from the application it authorized; for the typical adoption
  (a one-paragraph edit to one SKILL.md) that pipeline costs more than
  it returns. P4 takes the part of the observation the run actually
  evidenced: not who applies approved changes, but that the beads this
  skill mints should meet the same readiness gate as every other unit
  of work. What would promote the broader form: a run where step 5's
  approved set was large enough that inline application ran out of
  context or landed half-applied — the failure `SKILL.md:108-110`
  already warns about.
- **Checklist item 5 names only implementation.** The progress
  checklist's "5. Approved items implemented, verified, versions
  bumped" does not mention the Resolution, the RFC numbering, or the
  deferral beads that step 5's body also requires. No run evidence:
  this run stopped at step 4 and never exercised step 5.
- **Step 2 has no rule for an empty corpus.** `rfcs/` held no resolved
  RFCs and the beads database was empty, so the dedupe returned nothing
  on both halves. The agent recorded "no matches" and moved on; no cost,
  no guess. Not a finding.
- **Proposing that step 4 print the table to a file.** Would make the
  recommendation durable without beads, but it duplicates state the
  beads database already holds and the skill already writes to at step
  5. Rejected as the more expensive of two fixes; P1 reuses the
  existing store.

## Non-Goals

- Weakening the approval gate. P1 explicitly forbids implementing in an
  unattended run; it changes where the recommendation is written, not
  who decides.
- Changing the classification taxonomy (adopt-now / defer / reject) or
  the hybrid cadence.
- Introducing tooling. P2 and P3 are flag and command-name corrections
  to `bd` invocations the skill already makes.
- Moving step 5's implementation work into `groom-beads` or
  `bead-bonanza`. P4 changes what a deferral bead must contain, not who
  applies an approved change — see Considered and Not Proposed.

## Adoption

P1 is independent and carries the blocker; adopt it alone if nothing
else. P2 and P3 are one-line command corrections to step 2 and step 5
respectively — independent of P1 and of each other, though they are
adjacent enough to land together. P4 rewrites step 5's bead bullet and
is independent of all three, but it pairs naturally with P3 (same
bullet) and with P1 (which files beads under the same schema, so
adopting P1 without P4 propagates the thin schema to every unattended
run).

Files touched: `.claude/skills/triage-rfc/SKILL.md` only (P1–P4). P4
additionally makes the skill cite
[[work-order-definition-of-ready]], which already exists at
`plugin/skills/spec/references/` and is unchanged by this RFC.
All four change the skill's workflow, so its frontmatter `version`
bumps MINOR, 1.0 → 1.1. Note that `triage-rfc` lives only in
`.claude/skills/` and is not shipped under `plugin/skills/`, so no
`plugin/.claude-plugin/plugin.json` bump follows from these proposals.

Backward compatibility: additive. P1 adds a branch that no attended run
enters, so attended behaviour is unchanged. P2 and P3 widen what two
commands return and can only surface more prior art, never less. None
of the three depends on the model, though P1's value rises as agent
autonomy rises — a less capable agent facing the same silence is more
likely to resolve it by treating the absent human as approval.

Expected effect: an unattended triage ends with a resumable, labelled
recommendation instead of an evaporated one, and the dedupe step
actually reads the history it is told to read.

## Resolution

**Triaged 2026-07-31.** 4 proposals, all adopted and implemented.

| Proposal | Outcome |
|----------|---------|
| P1 — unattended-run rule at step 4 | **adopted.** Step 4 gained the unattended branch: classify and present, file one bead per proposal regardless of classification, implement nothing, leave RFCs in the inbox, and report which steps the gate blocked. The three tempting alternatives are named and forbidden. |
| P2 — dedupe command must see closed beads | **adopted.** Step 2 now uses `bd list --label rfc-triage --all`, with the reason the default hides closed and deferred issues, plus an explicit `--status` fallback. |
| P3 — name the theme-key enumeration command | **adopted.** Step 5 now names `bd label list-all \| grep 'theme:'` and points out that the per-key issue counts are themselves the recurrence signal. |
| P4 — bring step 5's beads to the Definition of Ready | **adopted.** Step 5 now defers to the spec-owned [[work-order-definition-of-ready]] instead of its own three-field schema, and requires deferrals carry status `deferred` plus a `grooming_status`. |

**Files changed:** `.claude/skills/triage-rfc/SKILL.md`.

**Version bumps:** `triage-rfc` 1.0 → 1.1. Note this skill is repo-local
and ships outside the plugin, so it does not by itself drive a plugin bump.
