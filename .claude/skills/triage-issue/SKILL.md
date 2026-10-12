---
name: triage-issue
description: Triage issues reported against the skills, templates and
  standards this repository vends — verify their evidence against
  current versions, dedupe against prior resolutions and deferred
  beads, decide adopt / defer / reject with a human, and turn what is
  accepted into groomed work orders bundled for a sprint. Use when
  asked to triage reported issues or bug reports (GitHub issues, or an
  issue file handed over from another project), review a reported
  issue, or check on deferred themes.
version: "3.0"
---

# Issue Triage

**Role: this repo is the upstream of every project using the plugin,
and this skill is how a report from downstream becomes work here.**

**Triage is the decide stage for issues that arrive from outside.** A
retro decides, with its user, which of *that run's* candidate issues
are filed; this decides which of the *incoming* ones are. Same judgement, same
human, same rule that declining is normal and is recorded — the only
difference is where the issue came from.

**Its most critical role sits downstream of that decision: turning a
reported issue into a groomed work order.** An issue is written by
whoever hit the problem, in whatever detail they had. A bead has to be
dispatchable to a fresh agent with zero context. Closing that gap is
the work, and it ends with an epic a sprint can swarm.

**Cadence — hybrid:** obvious, cheap, well-evidenced issues are adopted
immediately; plausible-but-unproven or expensive ones are deferred as
beads so recurrence across projects can surface priority. Rejection is
a first-class outcome and is recorded so the same idea is not
relitigated without new evidence.

**One issue is one actionable change**, which is the unit of decision.
Nothing here triages a document containing several of them.

## Inputs

**Issues, from any channel.** The channel does not change the work:

- **GitHub issues** on this repository — a project that ran one of our
  skills and reported a finding or a bug against it.
- **An issue file handed over in conversation** — the standalone
  Markdown a retro writes, outside source control, when it cannot reach
  the vendor's tracker. One file may hold several issues, one section
  each.
- **A report typed in conversation.**

This repository's **own** retros need no triage pass: we vend the
skills we run, so a retro here files the issues its user chooses as
beads directly, and that choice is the triage decision.

Every issue arrives in the shape the issue template defines — the one
the `spec` skill ships and every producer vendors. One that does not is
normalized into it before triage, never triaged as prose. **Nothing
received is committed to the repository**: the issue becomes a bead or
is declined, and the tracker records which.

**Also read** the beads database — prior outcomes and deferred issues
(`bd` CLI; label `issue-triage`). It is the only dedupe source.

## Process

Copy this checklist into your response and check items off as you go.
(A checklist copied into your response is not a task tool — WHERE the project bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress visible.)

```
Issue triage progress:
- [ ] 1. Provenance checked against current versions (per issue)
- [ ] 2. Deduped against resolved history and issue-triage beads
- [ ] 3. Every issue classified (adopt-now / defer / reject /
      already-implemented)
- [ ] 4. Recommendation table presented — STOPPED for human approval
- [ ] 5. Approved issues groomed into work orders and bundled
- [ ] 6. Received issues closed out at their source
- [ ] 7. Plugin version bumped IF any shipped skill/template/standard
      changed
```

Step 4 is a hard stop. Do not tick 5 before the human has answered.

**The issue is the unit throughout.** Classification, the table's rows,
the counts reported and the beads filed are all per issue. A received
file holding several issues is triaged one issue at a time.

Then, per issue, in received order:

1. **Provenance check.** Read the issue's provenance frontmatter —
   `run`, `branch`, `skill_version`, `model`. Compare the cited version
   against the current one. WHEN the source has changed since, verify
   the quoted evidence still exists: an issue against text already
   fixed or removed is rejected "already addressed in <file> v<X.Y>"
   without further debate. An issue missing its versions is still
   triaged, but say so; the citation burden was the reporter's.

   **Read the findings it cites.** Every claim in an issue traces to a
   finding id — that is what the template requires — so an issue whose
   evidence section names no finding is not verifiable and is returned
   to its reporter rather than triaged on plausibility.

   **Disclose self-approval, and test IDENTITY rather than model.** Take
   the first of these the issue carries: `session`, else `run`. WHEN it
   matches this triage session's own, the record SHALL state that the
   issue was authored and triaged by the same agent, and name any change
   that ratifies a decision its own author made. This does not
   disqualify it — self-report is how this repository improves — but an
   approval recorded without the disclosure reads as independent
   corroboration when there was none.

   **`model` is not an identity and SHALL NOT be used for this test.**
   Two unrelated agents, weeks apart, can run the same model — testing
   on it reports a conflict that does not exist, and misses the real one
   whenever the two sessions happen to differ. `model` stays in the
   template as provenance for reproducing a finding; it answers a
   different question.

   WHERE the issue carries neither `session` nor `run`, the record SHALL
   say the check **could not be made**. Silence there reads as
   independent corroboration, which is the outcome this rule exists to
   prevent.
2. **Dedupe.** Search the beads for prior outcomes on the same theme:

   ```
   bd list --label issue-triage --all
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

   Match on each bead's recorded outcome, not its title.
   - Prior REJECTED match: flag it with the prior rationale; recommend
     reconsideration only if this issue carries new evidence.
   - Prior DEFERRED match (a deferred bead): this is recurrence — the
     accumulation signal working. Link the new issue in the bead,
     bump its priority, and recommend promotion to adopt-now when a
     second independent project has hit the same theme.
   - Prior ADOPTED match: verify the adoption actually covers this
     evidence; if it does, reject as already addressed.
3. **Classify each issue:**
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
     this triage. Record it on the bead as "adopted — already
     implemented", name where the change lives, and bump nothing.

     The test against `reject — already addressed` is whether triage is
     ruling on the proposal's merits at all. Use **already-implemented**
     WHEN you would have adopted it had the tree been untouched: the
     proposal was right, and this triage says so. Use **reject —
     already addressed** WHEN there is nothing to rule on — the text it
     quotes is gone or rewritten, so its premise fails and triage takes
     no position on whether the idea was ever good.

     The difference is load-bearing at step 2, which reads prior
     outcomes. "Adopted" tells a future triage the idea is settled
     and should not be relitigated; "rejected — already addressed"
     tells it only that the evidence was stale, so the same idea backed
     by fresh evidence gets a fresh hearing.

   Before classifying an issue **adopt-now**, run two checks its
   reporter could not have run. Both are triage's, and each has a
   recorded outcome:

   - **Extent.** WHEN the evidence cites occurrences of a repeatable
     pattern — a syntax, a phrase, a defect class — enumerate that
     pattern across the whole surface before adopting. A cited list is
     a sample from one run's path, not a census: the reporting run only
     opened the files its task required. Put the enumerating command
     and its result in the bead's recorded outcome.
   - **Sufficiency.** Read the proposed Change against the finding it
     cites. WHEN the change resolves less than its own evidence
     describes, widen it and adopt with refinement.

   Failing either check does not sink an issue — it resizes it. An
   issue whose extent was understated is adopted with refinement over
   the full set, not rejected.
4. **Present the recommendation and STOP for approval.** One table:
   issue, classification, one-line rationale, files touched.

   A triage MAY originate a change no issue reported — one the evidence
   implies. Table it as its own row, numbered `T1..Tn` so it is never
   confused with a reported issue, with its origin given as triage. It
   carries the same evidence bar and needs the same approval; what it
   lacks is a reporter, so the rationale column carries triage's own
   reasoning and says so.

   Implement nothing before the human approves; the human may
   reclassify any row.

   **WHEN no human can answer** — an unattended, scheduled, or
   batch run — the gate does NOT dissolve, and absence is NOT
   approval. Take the unattended branch instead:
   - Classify every proposal as usual, and present the table. The
     analysis is the valuable part and it does not require a human.
   - **File one bead per issue, whatever its classification** —
     adopt-now included — each meeting the bar in step 5, labelled
     `issue-triage` and `theme:<key>`, with the recommended
     classification recorded in the bead. The beads ARE the handoff
     artifact: they survive the session, and a human can approve
     later by working the queue.
   - **Implement nothing, and bundle nothing.** Not the obvious ones,
     not the one-line ones.
   - **Leave the issues where they arrived.** Do not close them and
     do not mark them accepted. Acceptance records what a human
     approved; writing it here would be a record of an approval that
     never happened.
   - **Say in the report which steps the gate blocked** — namely 5
     through 7 — and name the beads filed, so the next attended
     session resumes from them rather than re-triaging the inbox.

   Do not resolve the ambiguity any other way. The three tempting
   alternatives — discarding the analysis, recording an outcome on
   the issue, or reading an absent human as consent — lose work,
   falsify the record, and defeat the gate respectively.
5. **Groom the approved issues into work orders, and bundle them.**
   **This is the step the whole skill exists for.**

   An issue is written by whoever hit the problem, in whatever detail
   they had. A bead has to be **dispatchable to a fresh agent with zero
   context** — right approach, verifiable done-ness, nothing left in a
   conversation that will not exist at dispatch. Those are different
   documents, and closing the gap is real work.

   - **Run `groom-beads` over every approved issue.** Do not restate its
     bar and do not approximate it: the Definition of Ready is
     spec-owned ([[work-order-definition-of-ready]]) and `groom-beads`
     is what enforces it. **The fresh-agent test is the one that
     bites** — a bead saying "adopt the reported issue about tier
     reasons" fails it, because the reader has neither the issue open
     nor the session that produced it.

     Each groomed bead SHALL carry: the issue's path or id, the
     evidence summary **inlined rather than linked**, the finding ids
     it rests on, and labels `issue-triage` plus `theme:<key>`.

   - **Sequence by FILE, not by issue.** WHEN two approved issues change
     the same file, wire a real `blocks` edge between them; the one that
     owns that file's version bump goes last in its chain. Issues from
     different reporters routinely land in the same file, and two beads
     with no edge between them are still a collision when they edit the
     same lines.

   - **Bundle the accepted set into one epic.** That epic is exactly
     what `sprint-scope` takes and a sprint swarms — which is what
     closes the loop: friction reported by one run becomes groomed work
     in the next.

     WHERE the approved set is one or two small issues touching
     disjoint files, implement inline and say so; an epic for a
     two-line change is overhead. **Say which you did, either way.**

   - Apply any inline changes; keep every touched file within its own
     conventions (80-column lines, EARS for normative statements).
   - Bump the version of every changed template (instruction-block
     `Version:`), skill (frontmatter `version`), and standard
     (frontmatter `version`) per [[template-authoring-standards]].
     **Bump each file exactly once for the whole triage, no matter how
     many proposals changed it** — the version announces the release,
     not the proposal count. WHEN several proposals touch one file, the
     bump belongs to whichever lands last; verify the bump after the
     final change to that file, never before.

     WHEN a changed file carries no version property at all, do not
     bump — add one at `0.1`, and let this triage's change land inside
     that first version. A higher initial value implies a predecessor a
     reader could diff against; for a file that has never been
     versioned there is none, and step 1's provenance check would then
     be asking about a version that never existed. Place the property
     per [[frontmatter-standards]].
   - File one bead per approved **deferral** to the same bar — deferred
     work is read months later by an agent with no memory of this
     triage, so it needs grooming exactly as much as adopted work does.

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
     recorded outcome will claim it landed.
   - **Record the outcome on the bead, not in a document.** Adopted /
     adopted-with-refinement + the refinement / deferred / rejected +
     rationale, with the files changed and versions bumped. **The
     tracker holds workflow state**; writing a resolution section into
     a separate document keeps a second copy of it in prose.

     A `T` change carries the same record, marked triage-originated
     rather than as some reporter's testimony.
6. **Close out each issue at its source.** A GitHub issue is closed
   with a pointer to its bead, or with the rejection rationale; a
   handed-over file needs nothing — it was never committed. The bead is
   the record from here on, and an issue left open behind a filed bead
   gets triaged twice. Report the outcome as counts per issue
   (`N adopted, N deferred, N rejected`) — the issue is the unit of
   decision, so it is the unit the counts are in too. **Write no log
   file**: the tracker already holds every outcome.
7. **Publish.** WHEN any skill, template, or standard **that ships
   inside the plugin** changed, bump the plugin version in
   `plugin/.claude-plugin/plugin.json` (MINOR for normative changes).
   A skill ships in the plugin IFF it has a source directory under
   `plugin/skills/`; a tracked skill living only under
   `.claude/skills/` (such as this skill itself) is repo-local, and
   changing it does NOT on its own justify a bump, because installs
   never receive it. WHEN a triage changed only repo-local skills, say
   so and skip the bump. Distribution is plugin-only: installs pick
   up the change on plugin update once a release carrying it is
   published, because the marketplace serves the plugin from the
   `stable` branch the publish workflow moves; a session run with
   `claude --plugin-dir ./plugin` sees it after `/reload-plugins`.

## Anti-Patterns

| Pattern | Problem |
|---------|---------|
| Implementing before the human approves the classification table | The gate exists because adoption changes every downstream project |
| Filing an accepted issue as a bead without grooming it | An issue is what the reporter had; a bead must survive dispatch to an agent with no context. Skipping the gap between them is how a queue fills with work nobody can start |
| Adopting on plausibility without verifying the cited findings | The pipeline's currency is verified evidence; spending it on taste devalues every future report |
| Silent rejection (no recorded outcome) | The next reporter re-derives and re-reports the same idea |
| Deferring without a theme label | Recurrence across projects becomes uncountable — the accumulation model stops working |
| Forgetting version bumps on changed files | Issues cite versions; stale versions break the "already fixed?" check |
| Editing a received issue's evidence or proposed change | The issue is the reporter's testimony; triage records an outcome, it does not rewrite the report |
| Adopting a self-reported issue without recording that it was self-reported | An undisclosed self-adoption is indistinguishable later from an independently reviewed one |
| Treating an issue's cited occurrences as the complete set | The reporter lists what its run touched; a defect class usually reaches files that run never opened, and the record then says fixed about a class that is still live |
| Writing a resolution document or log beside the tracker | The tracker holds workflow state. A second copy in prose drifts from it and counts in a different unit than it files in |
| Committing a received issue file to the repository | The repository is not a tracker; the issue becomes a bead or is declined |
