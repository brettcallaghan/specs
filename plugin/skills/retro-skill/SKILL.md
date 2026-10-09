---
name: retro-skill
description: Evidence-based retrospective on a SKILL that an agent just
  ran — identifies skill defects, refinements, and gaps that would make
  future runs succeed. Cites the exact instruction that misled, names
  the script that failed, flags friction seen on every run, and records
  the model that ran it. Use after running any skill (a `specs` plugin
  skill, this repo's own skills, or a project-local skill),
  especially one that struggled, improvised, or produced a wrong result,
  or when asked to retro a skill, critique a skill, or capture how a
  skill could be improved.
version: "0.5"
---

# Skill Retro

**Core principle:** a skill's quality is only observable when an agent
runs it. A SKILL.md can read clean and still mislead the agent following
it — the defect surfaces only in a real run, as the instruction that
pointed the wrong way, the script that errored, or the step where the
skill went silent and the agent had to improvise. So retro-skill runs
AFTER a skill has been exercised, and it argues from that run: verbatim
text, exact commands, observed behavior. Its currency is run evidence,
never opinion about how the skill reads.

**Scope — any skill.** A plugin skill in this repo (`spec`, `groom-beads`
and `retro-skill` itself), or a project's local skill. If an agent followed 
a skill's instructions and those instructions could have guided it better, 
that is a retro-skill finding.

**Distinct from `retro-spec`.** `retro-spec` critiques a phase's
*documents*; `retro-skill` critiques the *skills, templates, and
scripts*. A completed phase can warrant both.

## Whose problem is it — apply the merge test

**Findings here serve the repository that VENDS the skill**, not the project
the skill ran in. Those are the same people only when a repository retros its
own skills, and assuming they always are is what puts proposals a project
cannot act on into that project's backlog.

The words do not separate the cases — "the blueprint template has no rule for
X" and "our blueprint has no rule for X" are the same sentence about different
artifacts. One question does:

> **Where would the fix land?**
>
> - In **this project's own repository** → it is not a `retro-skill` finding.
>   It belongs to the project's own retro, as an issue in its tracker.
> - In the **repository that vends the skill** → it is yours.

A finding captured in-flight already carries the answer: `retro:tooling:<name>`
is this skill's audience; the six work-order kinds are the project's.

**A project SHALL NOT accumulate issues about skills it cannot change.** Where
this skill runs in a project that did not author the skill, its output leaves
through that project's `outbox/<plugin>/<skill>/` and is never filed in its
backlog.

## When to run

- After a skill runs **to completion** — while the session is fresh and
  is itself the evidence record. Long, multi-step skills yield the most.
- After a run that **struggled**: the agent improvised, guessed at a
  silent step, hit a script error, or produced a wrong result by
  following the guidance literally.
- When the **same friction recurs across runs** of a skill — recurrence
  is the strongest signal there is. Say how many runs and cite each.

Do NOT retro a skill nothing has exercised yet (no evidence — you would
produce taste), and do NOT open a retro to relitigate style preferences.

**Capture while fresh.** Skill friction evaporates from memory faster
than the wording that caused it. Run the retro in the same session, or
with the transcript at hand. For a skill run across multiple agents
(e.g. a plan executed by spawned agents), have each agent record skill
friction in-flight per the findings-capture protocol —
`references/shared/findings-capture-protocol.md` — and aggregate
once at the end; never one retro per agent.

## What counts as evidence — the three finding shapes

Every finding is exactly one of these, or it is not a retro-skill
finding:

1. **Misleading text** — an instruction that led the agent to the wrong
   action. Quote it **verbatim with `file:line`**, and state what the
   agent did because of it.
   > e.g. `triage-issue/SKILL.md:0` "normalize the item into a conforming
   > RFC" — assumed one file = one RFC; the real input was three
   > concatenated RFCs, so the agent had to invent a split step.
2. **Script / tool defect** — a command, script, or check the skill
   tells you to run that failed, returned wrong output, or silently
   no-op'd. Name the script, give the **exact command**, show
   **observed vs expected**, and a minimal repro.
   > e.g. `bd list --label <slug>` returns ~nothing at aggregation time
   > because it defaults to open issues; the closed closure set needs
   > `--status closed`. Repro: label two closed issues, run both forms.
3. **Gap / silence** — the skill gave no useful guidance where a
   decision was forced: it said nothing, *or* it said something that
   cost time/tokens without resolving the choice, *or* the run hit a
   situation the skill author never considered. Quote the surrounding
   step and state **"no rule at step N for `<situation>`"**, plus what
   the agent did instead.

Not every shape applies to every skill: a prose-only skill (nothing to
run) yields no script-defect findings — tag the shapes that occurred,
do not manufacture the others.

**Cite against the version you ran.** Quote the text as it stood in the
`skill_version` you exercised — from the transcript when the working
tree has since moved on. You are NOT required to match the skill's
current source: clients run different versions, and whether a proposal
still applies to the latest version is the triage side's check, not
yours. Record `skill_version` faithfully so triage can make that check.

A finding that is none of these — "reads awkwardly", "could be tighter"
— is not a finding; it is at most a note under Considered and Not
Proposed.

## Provenance — record the model (load-bearing)

A skill behaves differently under different models, efforts, and
harnesses. A defect real under one model may not reproduce under
another, and the triage side cannot scope the fix without knowing.
**retro-skill records this faithfully; it does not decide.** Whether a
finding is adopted, and how recurrence across runs is weighted, is the
triage side's call — never the retro's. Record:

- **skill** and **skill_version** — the skill critiqued and its
  SKILL.md frontmatter version.
- **model** and **effort** — exact model id and reasoning effort
  (e.g. `claude-opus-4-8[1m]`, effort high). Not "Claude".
- **session** — the harness's session identifier, WHERE it exposes one.
  Optional, and it is not a substitute for `model`: `model` says what
  behaviour to reproduce, `session` says *which agent* produced the
  finding, which is what lets triage disclose self-approval. Two
  unrelated agents can share a model, so `model` cannot answer it.
- **runs_observed** — how many runs stand behind this evidence. This is
  the denominator for any "every run" claim; never assert more
  recurrence than you observed.
- **origin** — `self` when this repo critiqued its own skill, otherwise
  the project the run happened in.
- **severity** — the highest severity among the RFC's proposals.
- **harness**, the **task** the skill was applied to, the **run
  dates**, and the versions of every reference/script/template invoked.

The first six are RFC **frontmatter** — they map 1:1 onto the triage
issue's labels; the rest live in the Provenance block. See
`references/skill-rfc-template.md`.

## The bar

A finding supports a proposal only if ALL hold:

- it cites concrete run evidence in one of the three shapes, not taste;
- recurrence is plausible — the same text / script / silence would
  mislead the next competent agent under the same model;
- a specific change would have prevented it, stated concretely (the
  proposed wording, the script fix, the rule to add at step N) — not
  "make it clearer".

Rank every surviving finding by **severity**:

| Severity | Test |
|----------|------|
| **blocker** | forced an unscripted improvisation, or the skill as written produced a wrong result |
| **snag** | the skill was silent on a decision that materially changed the outcome |
| **polish** | it worked, but a change would make the next run clearer, faster, or safer |

Lead with blockers; prefer a handful of cited blockers over a long tail
of polish. A run that surfaced nothing that clears the bar is a real,
valuable result — record the null (Workflow step 6). A converged skill
is the goal, not a quota of findings.

## Workflow

Copy this checklist into your response and check items off as you go.
(A checklist copied into your response is not a task tool — WHERE the project bans `TodoWrite`/`TaskCreate`, this is the compatible way to keep progress visible.)

```
Retro progress:
- [ ] 1. Scope & provenance recorded (skill, version, model, effort, runs)
- [ ] 2. Run friction harvested — verbatim text, exact commands, locations
- [ ] 3. Each finding classified by shape and severity
- [ ] 4. Bar applied — below-bar items moved to Considered and Not Proposed
- [ ] 5. One RFC per skill authored; readiness gate passed
- [ ] 6. Null result recorded IF nothing cleared the bar
- [ ] 7. Handed off — output routed, skill NOT edited
```


1. **Scope & provenance.** Name the skill exercised and the task it was
   applied to; fill the Provenance block above — model, effort, harness,
   runs observed, versions, dates.
2. **Harvest run friction.** Walk the session (and any in-flight
   findings) for the three shapes. Capture verbatim text and exact
   commands with locations as you go. Separate skill defects from the
   agent's own carelessness that the skill already warned against — the
   latter is out of scope.
3. **Classify & rank.** Assign each finding a shape and a severity.
   Non-findings and below-bar items drop to Considered and Not Proposed
   with the reason — never silently.
4. **Apply the bar.** IF no finding clears it, THEN write a null result
   (step 6) and stop — do not pad.
5. **Author one RFC per skill** from `references/skill-rfc-template.md`.
   Group the findings by the skill they target; author one RFC per
   skill (a run that surfaced findings about more than one skill
   produces more than one RFC — do not force them into one).

   **WHERE a finding is a property of several skills rather than of
   one** — the same gap present in each skill's text, not one skill's
   defect — do NOT file near-identical RFCs against each. Propose it as
   a requirement in the skill-authoring standard
   (standards/skill-best-practices.md, under the project's spec root),
   and file ONE RFC against
   the skill whose text most directly exhibits it, naming the other
   affected skills and their `file:line` in Evidence. A standards
   requirement is how the rule reaches skills not yet written; the
   single RFC is how it reaches the ones that exist. For each,
   set the frontmatter labels (skill, skill_version, model, effort,
   runs_observed, origin, severity); order proposals by severity, each
   carrying its finding shape, cited evidence, and a concrete change.
   Complete every section, then run the template's readiness gate. Fix what it reports and re-run; do not proceed while it fails. After two full failed cycles, stop and surface the blocker rather than continuing to retry.
   Set status `ready-for-review` only once the gate passes.
6. **Route the output** into a directory named for the skill critiqued,
   under one named for the plugin that vends it. `<target>` is the
   issue's `target` frontmatter value; `<plugin>` is the namespace the
   skill is invoked under (`specs:spec` → `specs`). One directory
   accumulates every issue about one skill — and recurrence,
   the strongest signal this pipeline has, becomes visible by listing
   it.
   - Retroing a skill **from a project that does not vend it** →
     `outbox/<plugin>/<target>/<YYYY-MM-DD>-<slug>.md`, one file per issue.
     The project's owner forwards them to the vending repository.
     **They are NEVER filed in this project's own tracker**: a backlog
     of proposals its owners cannot act on teaches them to ignore the
     channel.
   - Retroing a skill **inside the repository that vends it**
     (dogfooding, or this repo's own skills) → the outbox is swept into
     that repository's own tracker in the same run, since it is its own
     downstream. This closes the self-improvement loop: retro-skill →
     outbox → `triage-issue` → groomed beads → a sprint.
   - **Null result** → say so in the report and log it; a skill that
     retroed clean is information. Write no issue file — an empty
     issue is not a null result, it is a file someone has to open to
     discover it says nothing.

   **The outbox belongs to the project, not to us.** Emptying it,
   forwarding it, or ignoring it entirely are all its owners' call.
7. **Hand off.** Report what was produced and where it routes.

   Every factual claim in that report SHALL cite what it derives from —
   a finding number, a `file:line`, or a command and its output — or be
   marked explicitly as your own inference. You SHALL NOT state that a
   check was performed unless it was performed. A claim you cannot cite
   does not belong in the report, for exactly the reason it would not
   have cleared the bar as a finding: the bar governs the artifact, and
   a report that escapes it is the one layer with no gate.

   This is not a call to be more careful — it is mechanically checkable.
   A sentence either names its source or it does not.

   Do NOT edit the skill yourself — retro-skill's authority ends at the
   RFC; `triage-issue` resolves it after human approval.

## The bar, restated

An RFC is an argument from a run. The moment a proposal's Problem cannot
quote the instruction that misled, name the script that failed, or point
at the step where the rule should live, it is opinion — and opinion
about a skill belongs in Considered and Not Proposed at most.

## Anti-Patterns

| Pattern | Problem |
|---------|---------|
| Retroing a skill nothing has run | No run evidence exists; the output is taste, not findings |
| Running the retro from memory in a later session | The wording that misled is gone; findings degrade to vague impressions |
| "Reads awkwardly" with no observed defect | Style opinion, not a finding — needs a wrong action, a failed command, or a forced guess |
| Omitting the model / effort from Provenance | A model-specific defect cannot be reproduced or scoped; the fix may be wrong for another model |
| Asserting "every run" from a single run | Recurrence is a claim about a denominator; state runs observed, don't inflate one into many |
| Reporting a script bug without a command or repro | Triage cannot verify it; an unreproducible bug is a rumor |
| Proposing "be more careful" / "read carefully" | The fix for ignoring a warning is not another warning; find the misleading text or the gap instead |
| Padding with polish while blockers exist | Dilutes the RFC; rank by severity and lead with blockers |
| One RFC per finding, or one RFC spanning several skills | Bundle all findings for ONE skill into one RFC (its proposals are the units of decision); findings about a different skill get their own RFC. The exception is a finding that is a property of several skills at once — route it per step 5: a standards requirement plus ONE RFC against the skill that most exhibits it |
| N near-identical RFCs for one cross-cutting finding | The workaround the rule above invites, and the worse failure: triage then decides the same proposal N times. Use step 5's cross-skill route instead |
| Editing the skill during the retro | The retro's authority ends at the RFC; resolution belongs to `triage-issue` |

## Reference files

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
`scripts/sync-shared-references.sh` so this skill is self-contained. Read
them normally; never edit them. Fix the source named in each file's header
and re-run the script.

## Retro this skill

retro-skill applies to itself. WHEN these steps or the
`skill-rfc-template` fought you while authoring an RFC, offer to run
`retro-skill` on `retro-skill` while it is fresh (as the first
dogfood did). A clean run needs nothing; never run it unprompted.
