---
name: retro-spec
description: Retrospective on a completed phase of spec-driven work —
  examines the project documents the phase depended on and produces
  evidence-backed improvement RFCs for the project's own specs. Use
  after a downstream phase has exercised upstream specs (a feature
  blueprint was authored from an FRD and architecture blueprints; a
  blueprint was groomed into a plan; a plan was implemented), when asked
  to run a retro or retrospective on a project's specifications, or to
  capture lessons learned about the documents a phase consumed. To
  critique the SKILLS, templates, or scripts that guided the phase, use
  the retro-skill skill instead.
version: "1.7.0"
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
better result next time? Findings that clear the bar (below) become one
RFC; a retro that finds nothing produces a null note.

The parallel question — what improvements to the SKILLS, templates, and
scripts that guided the phase, or the creation of the spec itself, would
help ANY project — is the domain of the `retro-skill` skill. The same completed phase can warrant both
retros: run this one for the document lens, `retro-skill` for the skill
lens. When friction traces to the guidance rather than the project's
documents, route it to `retro-skill` rather than filing it here.

## Whose problem is it — apply the merge test

**This skill's findings serve the project whose documents were exercised.**
Its output is issues in that project's own tracker — the audience has a
tracker and is the same people, so no proposal document is needed to tell them
about their own repository.

**The boundary with `retro-skill` is sharper than the words are.** A spec
*template* this framework vends and a spec *document* in a project are
described by identical sentences. One question separates them:

> **Where would the fix land?**
>
> - In **this project's own repository** → yours. An issue in its tracker.
> - In the **repository that vends the template or standard** → `retro-skill`'s.
>   It leaves through `outbox/<plugin>/` and is never filed in this project's
>   backlog.

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
  Never run a full retro per work order: per-item RFCs restate the
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
- [ ] 1. Scope and provenance established (phase, docs, skill/template versions)
- [ ] 2. Evidence harvested from the session / in-flight findings
- [ ] 3. Each finding root-caused (project-doc / skill-template / one-off)
- [ ] 4. Bar applied — below-bar items to Considered and Not Proposed
- [ ] 5. Project RFC authored IF anything cleared the bar; gate passed
- [ ] 6. Null note recorded IF nothing did
- [ ] 7. Handed off; retro-skill flagged IF guidance friction surfaced
```


1. **Establish scope and provenance.** Identify the phase just
   executed, the documents it consumed, the documents it produced, and
   the skill(s) and template(s) that guided it. Record the version of
   each (SKILL.md frontmatter; template instruction headers; standards
   frontmatter) and the plugin version — the RFC's Provenance section
   needs them, and triage cannot judge "already fixed?" without them.
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
     misleading (feeds this project RFC)
   - **skill/template/script gap** — the guidance itself was missing or
     misleading — NOT a finding for this retro; hand it to `retro-skill`
     (note it so the skill lens can pick it up)
   - **one-off** — carelessness the existing guidance already warns
     against; not an RFC finding (record under Considered and Not
     Proposed if worth remembering)
4. **Apply the bar.** A finding supports a proposal only if ALL hold:
   - it traces to concrete session evidence (step 2), not to taste
   - recurrence is plausible — the same guidance would mislead the
     next competent agent the same way
   - a specific change would have prevented it, and you can state
     that change concretely
   Findings that fail the bar go to Considered and Not Proposed with
   the reason. IF no finding clears the bar for either question, THEN
   skip to step 6 — do not pad an RFC with marginal proposals; a
   clean retro is signal that the guidance has converged.
5. **Author the project RFC** from `references/rfc-template.md` (it
   cites `[[template-authoring-standards]]`, whose canonical copy is the
   sibling spec skill's
   `references/shared/template-authoring-standards.md`), bundling every
   qualifying finding into the project variant (a numbered slug —
   see the template's Numbering and location). File it under
   `rfcs/inbox/<spec-type>/`, named
   `rfc-<nnnn>-<short-description>.md` with the repo's next free
   number. `<spec-type>` is the document type the findings most concern
   — `frd`, `adr`, `container-blueprint`, `standards` — so one directory
   accumulates every RFC about one document type and recurrence is
   visible by listing it. Complete every section, run the template's
   readiness gate. Fix what it reports and re-run; do not proceed while it fails. After two full failed cycles, stop and surface the blocker rather than continuing to retry. Set status to
   `ready-for-review` only once the gate passes. Order proposals
   by expected impact. (Skill-lens findings go to `retro-skill`, which
   authors its own skill-authors variant RFC — do not file one from
   here.)
6. **Record a null result.** WHEN no finding clears the bar, write a
   retro note to
   `rfcs/inbox/<spec-type>/retro-note-<phase>-<YYYY-MM-DD>.md`: one short
   paragraph naming the phase executed, the documents examined, and the
   outcome ("no project-doc finding cleared the bar"). A phase that
   retroed clean is information — it says the consumed documents held up.
7. **Hand off.** Report what was produced; the project RFC awaits the
   project's own review in `rfcs/inbox/<spec-type>/`. IF the phase also surfaced
   guidance friction (any finding root-caused to skill/template/script),
   THEN remind the human to run `retro-skill` for the skill lens — its
   RFCs travel to the framework repo's `rfcs/inbox/` for triage.

## The bar, restated

"High probability of meaningful improvement" is the test. Concretely:
prefer five proposals with verbatim citations and observed failures
over fifteen plausible ideas. An RFC is an argument from evidence; the
moment a proposal's Problem section cannot quote the text that caused
the defect (or point at the hole where text should be), it is opinion
and belongs in Considered and Not Proposed at most.

## Anti-Patterns

| Pattern | Problem |
|---------|---------|
| Retroing documents nothing has consumed yet | No evidence exists; the retro produces taste, not findings |
| Running the retro from memory in a later session | The friction detail is gone; findings degrade to vague impressions — capture in-session, or in-flight via the findings protocol |
| Running a full retro per work order | N near-identical RFCs per feature; the implementer records findings, the final DoD item runs the ONE aggregated retro |
| Recording analysis or proposals as findings | Capture is a sensor's job; root-causing without cross-item perspective bakes in wrong conclusions |
| Proposals without verbatim source citations | Triage cannot verify the defect or check whether it is already fixed |
| Blaming guidance for authoring carelessness | The fix for ignoring an existing warning is not another warning |
| Omitting versions from Provenance | "Is this already fixed?" becomes archaeology at triage time |
| Padding a thin retro to justify an RFC | Dilutes trust in every RFC; a null retro note is the honest artifact |
| Re-proposing an idea a prior RFC's Resolution rejected, without new evidence | Relitigates settled questions; check the project's `rfcs/` first |
| Weakening a consumer's checks to make authoring easier | The checks catch real defects; move the knowledge upstream instead |
| Editing the project's documents, skills, or templates directly during the retro | The retro's authority ends at the RFC; resolution belongs to the doc owners (project review or framework triage) |

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
while the session is fresh, to capture the friction as an
evidence-backed RFC. (That is distinct from step 7, which retros the
documents and skills of the phase just executed; this retros
`retro-spec` itself.) A clean run needs nothing; never run it
unprompted.
