---
name: retro-spec
description: Retrospective on a completed phase of spec-driven work —
  examines the project documents the phase depended on and reports, in
  the session, a stack-ranked list of evidence-backed candidate issues
  against the project's own specs for the user to file. Use
  after a downstream phase has exercised upstream specs (a feature
  blueprint was authored from an FRD and architecture blueprints; a
  blueprint was groomed into a plan; a plan was implemented), when asked
  to run a retro or retrospective on a project's specifications, or to
  capture lessons learned about the documents a phase consumed. To
  critique the SKILLS, templates, or scripts that guided the phase, use
  the retro-skill skill instead.
---

# Spec Retro

**Core principle:** a document's quality is only observable through its
consumers. A spec can look complete and still mislead the next agent —
the defects surface when a downstream phase exercises it. So the retro
runs AFTER a dependent phase executes, and it examines the
*dependencies* of that phase: the documents the phase consumed and the
skill and templates that guided it. Do not retro a document that
nothing has exercised yet — there is no evidence to argue from.

The retro answers one question: what improvements to the documents this
phase consumed — the project's own specs — would help agents produce a
better result next time? Findings that clear the bar (below) become
candidate issues in a report presented to the user, who chooses which to
file in the project's tracker; a retro that finds nothing says so.

**Nothing this skill produces is committed to the repository** — no
retro document, no proposal file, no note, no log. What is worth
addressing becomes an issue in the tracker; what is not, is not kept.

The parallel question — what improvements to the SKILLS, templates, and
scripts that guided the phase, or the creation of the spec itself, would
help ANY project — is the domain of the `retro-skill` skill. The same completed phase can warrant both
retros: run this one for the document lens, `retro-skill` for the skill
lens. When friction traces to the guidance rather than the project's
documents, route it to `retro-skill` rather than filing it here.

## Whose problem is it — apply the merge test

**This skill's findings serve the project whose documents were exercised.**
Its output is issues in that project's own tracker — the audience has a
tracker and is the same people, so no document is needed to tell them about
their own repository.

**The boundary with `retro-skill` is sharper than the words are.** A spec
*template* this framework vends and a spec *document* in a project are
described by identical sentences. One question separates them:

> **Where would the fix land?**
>
> - In **this project's own repository** → yours. An issue in its tracker.
> - In the **repository that vends the template or standard** → `retro-skill`'s.
>   It goes to that vendor and is never filed in this project's backlog.

So: a gap in the FRD *template* this framework ships is `retro-skill`'s. A
gap in this project's own FRD-014 is yours. Both would be reported in the same
words by an agent that hit them.

## When to run

| Phase just executed | Project documents to retro |
|---------------------|----------------------------|
| Authored a feature blueprint (spec skill) | The FRD it realizes; the architecture blueprints and standards it referenced |
| Groomed a blueprint into a plan (groom-feature-blueprint skill) | The feature blueprint and its FRD |
| Implemented a plan's work orders (retro run by the final DoD item's agent, aggregating per-item findings — see the findings capture protocol) | The work orders produced; the blueprint and FRD behind them |
| Authored any spec from upstream docs | The upstream docs it consumed |

(The skills, templates, and scripts these phases followed are the
subject of `retro-skill`, not of this table.)

Evidence does not survive being recalled from memory later, so it must
be captured while it exists:

- **Single-session phases** (authoring, grooming): run the retro in
  the SAME session, or with the transcript at hand — the session is
  the findings record. A groom-feature-blueprint Phase 1 bounce is prime retro
  input: every bounced gap is a candidate finding about the authoring
  side.
- **Multi-session / multi-agent phases** (implementing a plan, where
  each work order may go to a different spawned agent): no one session
  holds the evidence. Each implementing agent records findings
  in-flight per the findings capture protocol below, and ONE
  aggregated retro runs at the end of the phase — by the agent
  executing the plan's final DoD item, which reads every item anyway.
  Never run a full retro per work order: per-item retros restate the
  same systemic issues N times, and the item's implementer lacks the
  cross-item perspective to tell a one-off from a pattern.

## Findings capture protocol

Defined once in `references/shared/findings-capture-protocol.md` — load and
follow it from that document; do not restate it. In brief: the
capturing agent is a sensor that records findings (friction only,
zero analysis) as `retro:`-prefixed `bd comment`s on its own item
(task-file backend: an appended `## Retro findings` section), and
the final DoD item's agent harvests them across the feature's
closure set, adds its own DoD-verification friction, and proceeds
from Workflow step 3 with that pool as the step-2 output. Findings
stay on closed items permanently as raw provenance.

The protocol document's authored copy lives under `specs/standards/`; it
is the single statement of this channel, and every skill that captures
or harvests findings — including this one — vendors it into its own
`references/shared/` and points at that copy. It is never copied into a
target project, and never restated per item — an item carries only the
short standing instruction its grooming skill stamps, which is
self-sufficient, so an implementing agent that sees nothing but its own
item still records findings without loading this skill.

## Workflow

Copy this checklist into your response and check items off as you go.
(A checklist copied into your response is not a task tool — WHERE the project bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress visible.)

```
Retro progress:
- [ ] 1. Scope and provenance established (phase, docs, plugin release)
- [ ] 2. Evidence harvested from the session / in-flight findings
- [ ] 3. Each finding root-caused (project-doc / skill-template / one-off)
- [ ] 4. Bar applied — below-bar items to Considered and Not Proposed
- [ ] 5. Report presented with candidate issues stack-ranked — STOPPED
- [ ] 6. Issues the user chose filed in the project's tracker
- [ ] 7. Handed off; retro-skill flagged IF guidance friction surfaced
```

Step 5 is a hard stop. Do not tick 6 before the user has answered.

1. **Establish scope and provenance.** Identify the phase just
   executed, the documents it consumed, the documents it produced, and
   the skill(s) and template(s) that guided it. Record the plugin's
   release version, which identifies its skills, templates, and
   standards — the report's Provenance block needs it, and a later
   reader cannot judge "already fixed?" without it.
2. **Harvest evidence from the session.** Walk the session for
   friction, and state each item as a finding — what happened, the
   defect, the root cause. Sources of findings:
   - corrections the human made during the review loop
   - gate failures and bounces (readiness checks, content checks)
   - places the agent guessed, improvised, or worked around guidance
   - syntax or sections that could not be used as documented
   - design decisions discovered missing at consume time
   - anything written twice
   - findings recorded in-flight on work orders (the findings capture
     protocol) — for an implementation-phase retro these ARE the
     session evidence
   For each finding, capture the source text that caused it —
   verbatim, with location — or note explicitly that no guidance
   existed where it was needed.
3. **Root-cause each finding** to exactly one of:
   - **project-doc gap** — the consumed document was missing or
     misleading (a candidate issue for this retro)
   - **skill/template/script gap** — the guidance itself was missing or
     misleading — NOT a finding for this retro; hand it to `retro-skill`
     (note it so the skill lens can pick it up)
   - **one-off** — carelessness the existing guidance already warns
     against; not an issue (record under Considered and Not Proposed if
     worth remembering)
4. **Apply the bar.** A finding supports a candidate issue only if ALL
   hold:
   - it traces to concrete session evidence (step 2), not to taste
   - recurrence is plausible — the same document would mislead the
     next competent agent the same way
   - a specific change would have prevented it, and you can state
     that change concretely
   Findings that fail the bar go to Considered and Not Proposed with
   the reason. Shape each finding that clears it into a candidate issue
   per `references/shared/issue-template.md`, `audience: project`,
   `target` naming the document. **One issue is one change** — the
   issue is the unit of decision, so the user must be able to accept
   one and decline the next. Write each issue with the `tech-writer`
   skill, taking the tone from the template's reader profile. Before
   raising one, search the project's tracker for an issue already open
   or closed on the same document and theme: an open one gains this finding as recurrence, and a declined
   one is re-raised only with new evidence.
5. **Present the report, and STOP.** In the session, in this order:
   - **Provenance** — the phase, the documents examined, and the
     release version from step 1.
   - **Candidate issues, stack-ranked** by expected impact, highest
     first — `priority` 0 (critical) to 4 (backlog), ties broken by
     recurrence. One numbered entry each: title, target document,
     priority, the evidence it rests on, and the proposed change.
     Render each in full as the issue template's body, so the text the
     user approves is the text filed.
   - **Considered and Not Proposed** — each below-bar item and why.
   - **Guidance friction** — findings root-caused to a skill, template
     or script in step 3, listed for `retro-skill`.

   Then ask which to file. The user may file any subset, re-rank, or
   file none. Do not pad the list to justify the retro — a short list
   of cited issues is worth more than a long one of plausible ideas.

   **Null result** — WHEN nothing cleared the bar, say so in the
   report: name the phase, the documents examined, and the outcome
   ("no project-doc finding cleared the bar"). A phase that retroed
   clean is information — it says the consumed documents held up — and
   it needs no file.

   **Absence is not approval.** WHEN no human can answer — the final
   DoD item of a groomed plan, run by an implementing agent — file
   nothing. Record the report as a comment on that item in the tracker
   (task-file backend: the item's final output), so a human can choose
   from it later, and say in the item's close that the candidates await
   a decision.
6. **File what the user chose** in the project's own tracker — one
   bead per issue where the project uses beads, groomed to its
   Definition of Ready; otherwise one issue in whatever tracker it
   keeps. The issue's title, priority and sections become the record's
   title, priority and description.
7. **Hand off.** Report what was filed (ids or links) and what the user
   declined. IF the phase also surfaced guidance friction (any finding
   root-caused to skill/template/script), THEN remind the human to run
   `retro-skill` for the skill lens — it reports those to whoever vends
   the skill.

## The bar, restated

"High probability of meaningful improvement" is the test. Concretely:
prefer five issues with verbatim citations and observed failures over
fifteen plausible ideas. An issue is an argument from evidence; the
moment its Problem cannot quote the text that caused the defect (or
point at the hole where text should be), it is opinion and belongs in
Considered and Not Proposed at most.

## Anti-Patterns

| Pattern | Problem |
|---------|---------|
| Retroing documents nothing has consumed yet | No evidence exists; the retro produces taste, not findings |
| Running the retro from memory in a later session | The friction detail is gone; findings degrade to vague impressions — capture in-session, or in-flight via the findings protocol |
| Running a full retro per work order | N near-identical issues per feature; the implementer records findings, the final DoD item runs the ONE aggregated retro |
| Recording analysis or proposals as findings | Capture is a sensor's job; root-causing without cross-item perspective bakes in wrong conclusions |
| Issues without verbatim source citations | Nobody can verify the defect or check whether it is already fixed |
| Blaming guidance for authoring carelessness | The fix for ignoring an existing warning is not another warning |
| Omitting the release version from Provenance | "Is this already fixed?" becomes archaeology |
| Padding a thin retro with marginal issues | Dilutes trust in every issue; saying the retro came up clean is the honest result |
| Bundling several changes into one issue | The issue is the unit of decision; the user cannot accept half of it |
| Re-raising an issue the project already declined, without new evidence | Relitigates settled questions; search the tracker first |
| Filing before the user chooses | The ranking is a recommendation; filing is the user's decision |
| Committing the report, a retro note, or an issue file to the repository | The repository is not a tracker. What is worth doing is an issue in the tracker; the rest is not kept |
| Weakening a consumer's checks to make authoring easier | The checks catch real defects; move the knowledge upstream instead |
| Editing the project's documents, skills, or templates directly during the retro | The retro's authority ends at the issue; resolution belongs to the doc owners |

## Reference files

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
`scripts/sync-shared-references.sh` so this skill is self-contained. Read
them normally; never edit them. Fix the source named in each file's header
and re-run the script.

## Retro this skill

WHEN a run of `retro-spec` struggled — you improvised around its
guidance, a command it prescribed misbehaved, or an instruction here
pointed you the wrong way — offer to run `retro-skill` on `retro-spec`
while the session is fresh, to capture the friction as evidence-backed
issues. (That is distinct from step 7, which retros the
documents and skills of the phase just executed; this retros
`retro-spec` itself.) A clean run needs nothing; never run it
unprompted.
