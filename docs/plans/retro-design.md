# The retro, end to end

**Status: DRAFT for review.** Everything else in wave 5 is decided inside
this document — the finding taxonomy, the candidate-issue flow, the audience
split in three skills, the finding and issue templates that replace the RFC
format, the `.sf/outbox/` delivery path, `triage-rfc`'s future as
`triage-issue`, and what happens to the 27 queued RFCs.

---

## Objective

**Convert the friction encountered during a run into actions that make the
next run better.**

Not a report someone reads. Not a proposal queue someone drains. Actions —
filed where the people who can take them already look.

---

## Two branches, and they are not equally urgent

Every finding a run produces belongs to exactly one of two retrospectives.
They share stages 1–3 and diverge completely after that, because their
audiences are different people with different powers.

### Branch A — the project's own quality

**About:** the spec corpus, the standards, the environment, the checks, the
grooming — everything the project's own owners control.
**Audience:** the owners of the project the sprint ran in.
**Destination:** **beads in that project's tracker.**

**This is the branch that matters most to anyone who adopts the plugin.** It
is the whole return on running sprints this way: a corpus that gets measurably
better at answering the questions implementers actually ask, because every
sprint reports which questions it could not answer.

An adopting project should be able to get this branch and never think about
the other one.

### Branch B — the sprint skills themselves

**About:** the skills, scripts, templates and protocols this repository vends.
**Audience:** the owners of **this** repository — not the project the sprint
ran in, unless they are the same.
**Destination:** **issues delivered to this repository.**

**This branch shrinks as the skills mature, and it is critically important
right now.** The skills are young, they are being restructured, and every run
is the only source of evidence about where they mislead. Capturing that
feedback well is currently worth more than the marginal branch-A finding.

But it is *this repository's* feedback loop, and the design has to keep it
from leaking into anyone else's backlog.

### The test that separates them: who can merge the fix?

| | |
|---|---|
| the fix lands in **the project's own repository** | **branch A** — a bead there |
| the fix lands in **`agent-plugins`** | **branch B** — an issue delivered here |

**Use this test, not the words.** "The FRD template has no rule for X" and
"our FRD has no rule for X" are the same sentence about different artifacts.
The template ships from here; the FRD lives there. Only the merge question
tells them apart, and it tells them apart every time.

### Why this has been invisible

**This repository is both audiences at once.** Every finding had somewhere
plausible to go, so nothing ever forced the question. A second project using
the plugin would have surfaced it on day one — by accumulating, in its own
backlog, proposals about skills it cannot change.

---

## What the evidence says

Seven measurements, each naming the command or source behind it. Read them
through the two-branch lens: several say something different about branch A
than about branch B.

| | measured | source |
|---|---|---|
| **1** | **240 findings** harvested across five sprints | `rfcs/retro-log.md` harvest counts (73 + 14 + 16 + 64 + 73) |
| **2** | **13 RFCs processed; 63 proposals adopted**, 4 deferred, 1 rejected. Processed RFCs were deleted, not archived | `rfcs/retro-log.md` outcome column, summed |
| **3** | **27 RFCs still queued**, 25 ready-for-review, 2 draft | `find rfcs/inbox -name '*.md' \| wc -l` |
| **4** | **All 27 carry `origin: self`**; 93 proposals processed with **no reporter outside this repository** | `grep -h '^origin:' rfcs/inbox/*/*.md \| sort \| uniq -c` → `27 origin: self`; `rfcs/retro-log.md` closing note |
| **5** | The protocol's **six categories are never recorded** — every retro re-derives each finding's kind by reading it. **Confirmed by the validation gate**: all 73 had to be read in full to classify | the `retro:` prefix carries no kind; `specs/standards/findings-capture-protocol.md` lists six triggers and no way to record which fired; `docs/research/retro-taxonomy-validation/README.md` |
| **6** | A finding that became a bead took **five hops**: finding → report → RFC → triage → bead | `rfcs/retro-log.md` disposition column |
| **7** | Of 73 real findings, **60 are branch A, 5 branch B, and 7 belong to neither** — they are evidence or facts about a dependency | `docs/research/retro-taxonomy-validation/README.md` |

### What follows from them

**The pipeline works, and that is why its parts are kept rather than
replaced.** Measurement 2 is the one to lead with: 63 adopted proposals have
shipped into these skills through it. Evidence with citations, a concrete
change, and provenance all earned their cost, and all three carry forward into
the issue template unchanged. **What is retired is the RFC document, not the
practice it carried.**

**Its queue grew for two reasons, and both are structural.** First, branch A
was routed into it — measurements 3 and 6 together: 27 unprocessed proposals,
and five hops from friction to action
for findings whose destination was determinable at hop one. A finding about
*this repository's own specs* went out through a cross-project proposal format
and back in through triage — when a bead was available the entire time.

**Measurement 4 is about reporters, not about value.** Ninety-three proposals
processed, none from an outside project. That does not mean the format failed;
it means **the cross-project accumulation the format was built for has not
started yet**, and until it does, the recurrence machinery is serving a single
reporter. It stays, scoped honestly.

**Measurement 5 is why routing is expensive.** A finding's kind — and
therefore its branch — is knowable at the moment of capture, by the agent that
hit the friction, for the cost of one word. It is instead re-derived later by
an agent reading 73 findings cold.

---

## The insight this design turns on

> **One finding format and one issue format, for both branches. The branches
> differ in where the issue is delivered — nothing else.**

Two things follow, and they are the whole design:

**This repository's project-level findings stop going through the RFC
process.** They become beads, through the sprint's own retro, exactly like any
adopting project's. The queue stops growing from branch A, and this repository
finally runs the same loop it ships — which is the only way its own defects
get found.

**The RFC document itself is retired**, for both branches. Once an issue is
the unit, a document that bundles several of them is a container with nothing
to contain: its own log already counts per proposal rather than per RFC. What
the format got right — evidence with citations, a concrete change, provenance
including the model — moves into the issue template intact.

---

## The six stages

```
1 CAPTURE ..... implementing agent, at the moment of friction        MECHANICAL
2 HARVEST ..... one query over the run's closure set                 MECHANICAL
3 ROUTE ....... branch A or B, from the kind the capture recorded    MECHANICAL
        ─────────── the branches diverge here ───────────
4 REASON ...... findings → candidate issues; dedupe, cluster, rank   AGENTIC
5 DECIDE ...... the human, ONCE, at the point of maximum context     HUMAN
6 FILE ........ ONE issue format; only the destination differs       MECHANICAL
```

**Stages 1–3 never wait on a person and never require judgement. Stage 4 is
where all the reasoning happens. Stage 5 is the single human decision point.
Stage 6 is mechanical given 5.**

**The retro runs on every run, without exception** — including a run that
closed nothing. A sprint that bounced its whole queue produced no code and a
great deal of evidence about grooming, readiness and spec quality, which is
branch A's best material. The coordinator runs it whatever the outcome; a
retro that only fires on success cannot see the failures.

---

### 1 · Capture — record what only you can know

**Who:** the implementing agent, during the work.
**Produces:** one `bd comment` per finding, on the bead it happened to.

```
retro:<kind>[:<target>] <the finding>
```

#### Capture the evidence in full. Defer the analysis.

The old rule was "1–3 sentences". **That is the wrong axis.** It compresses
evidence and analysis together, and they have opposite economics:

| | recoverable later? | so |
|---|---|---|
| **evidence** — the exact quote, the file and line, the command and its output, what you expected versus what happened | **NO.** The session compacts, the worktree is removed, the tree moves on. Whoever aggregates this is reading it cold, weeks later | **record it completely, now** |
| **analysis** — root cause, proposal, priority, what should change | **YES**, from the evidence, at any time, by an agent with more perspective than you have mid-task | **do not record it** |

So the rule becomes:

> **Record everything about *what happened* that will not survive this
> session. Record nothing about *what it means*.**

The sensor framing is retained and is the thing most worth keeping — but a
sensor that under-reports its reading is not being disciplined, it is being
lossy. **A finding with a quote, a path and a command beats three sentences of
prose every time**, and it is usually shorter.

**Still forbidden:** root-causing, proposals, RFC prose, priority, and any
sentence beginning "we should". Those are stage 4's, and an implementing agent
writing them mid-task both derails its own work and produces worse analysis
than the aggregator would.

#### The taxonomy — nine kinds

Six are the capture protocol's existing trigger list, which already enumerates
what an implementing agent notices. The seventh is the class it does not name;
the last two were added by the validation gate below, which the first seven
failed.

| kind | fires when | branch |
|---|---|---|
| `gap` | information was needed that the work order and its references did not carry | **A** |
| `decision` | a design decision had to be made that the item or blueprint should have pinned | **A** |
| `stale` | the item contradicted repo reality — moved code, a dead anchor, a changed element | **A** |
| `unverifiable` | an AC or citation obligation could not be satisfied as declared | **A** |
| `sequencing` | a broken predecessor, or a blocker the dependency graph did not encode | **A** |
| `scope` | the item was mis-scoped — too large to review as one unit, trivial, or overlapping another | **A** |
| `tooling` | a skill, script, template or protocol misled, failed, or had no rule where one was needed | **B** |
| `proof` | recording **how a claim was established** — an observed pre-fix failure, a negative control, a prior risk checked and cleared | **neither** |
| `environment` | a durable fact about a **dependency, tool or platform** the next agent will hit, which no work order could have carried | **the ledger** |

**Two of these nine do not produce issues, and that is the point.** The first
seven all assume a finding is a candidate defect. Validation found that 7 of 73
real findings are not: they are evidence, and durable facts about someone
else's library. Both are worth capturing and neither is worth filing.

- **`proof` stays in the report**, attached to the finding or bead it supports.
  It is the repository's own evidence standard being obeyed — `AGENTS.md`
  requires a claim to cite the command run and the output observed — so routing
  it as friction would suppress exactly the rigour that standard asks for.
- **`environment` goes to the environmental-failure ledger**, which already
  exists in `AGENTS.md` for exactly this. Filing it as an issue would blame a
  work order for not carrying a fact about a third-party parser.

**`target` is what makes stage 3 a lookup.** For `tooling` it is the skill or
script (`retro:tooling:sprint-wave`). For the six branch-A kinds it is the
document or bead the finding is about (`retro:gap:FRD-014`). Omit it where
there genuinely is none; that costs a cluster, not a finding.

**When in doubt about the branch, apply the merge test** — where would the fix
land? An implementing agent can answer that in one beat.

**A finding may correct an earlier one.** Findings are append-only, so the only
way to withdraw a claim is to supersede it. An optional `corrects:
<finding-id>` field expresses that relationship, which prose cannot: validation
found an agent writing "correction to the line above", which an aggregator
reading the two cold has to infer from a phrase.

**`incident` is not a kind — it is a severity that crosses the kinds.** An
incident is a decision that changed what "correct" means. `run-bead` records it
as `retro: incident —` and it leads the sprint report, because a human needs to
see it before anything else.

**Incidents come in several types, and they map onto the kinds already here:**
an amended acceptance criterion is a `decision`; a standard/requirement/AC
tension is a `decision` against two targets; a contract or interface change is
a `scope`; a spec silent where a bead needed it to speak is a `gap`. So an
incident routes like any other finding and needs no eighth kind — what makes it
an incident is that the run **acted on it** rather than merely noticing it.

Record both: the kind routes it, the incident marker escalates it.

```
retro:gap:ADR-014 incident — <what was decided, and what was rejected>
```

Its *reporting* obligation belongs to the sprint, not the retro. Its *routing*
belongs here.

#### Validation gate — RUN, and the taxonomy was extended by what it found

All 73 findings in the 2026-08-15 report classified. Full working:
`docs/research/retro-taxonomy-validation/README.md`.

| | count | share |
|---|---|---|
| branch **A** — the project's own quality | 60 | 82.2% |
| branch **B** — the skills this repository vends | 5 | 6.8% |
| **residue against the seven proposed kinds** | **8** | **11.0%** |

**The seven kinds failed the 5% bar**, and the residue named what was missing
rather than sprawling: 4 findings recording *how a claim was established*, 3
recording a *dependency's behaviour*, 1 *correcting an earlier finding*. Hence
`proof`, `environment` and the `corrects` field above. **Recomputed residue
with those added: 1 of 73 (1.4%)** — and that one is what `corrects` absorbs.

**Six findings were ambiguous between two kinds, and all six pairs route to
the same branch.** That is the right failure mode: the taxonomy only has to be
right about the branch, and it was right about the branch in all 65 classified
cases. Ambiguity costs a cluster, not a route.

**The classification cost is the measurement the whole `<kind>` proposal rests
on, and it held**: every one of the 73 had to be read in full, because nothing
in the current `retro:` prefix narrows the work. That recurs every retro.

**Two branch-B findings turned out to be the same problem** — `wvd.6#1` and
`wvd.11#1`, both reporting that Read and Grep silently tolerate a
shared-checkout path where Edit and Write refuse it — found independently by
two beads, listed separately in the report, with nothing connecting them. Under
this scheme both carry `retro:tooling:subagent-prompts` and cluster
automatically. **That is the recurrence signal working, on real data, before
the mechanism exists.**

---

### 2 · Harvest — one query, and a zero is a bug

**Who:** the retro agent.
**Produces:** every `retro:` comment across the run's closure set, verbatim.

```bash
bd export | jq -r 'select(<run scope>) | .comments[]?
  | select(.text | startswith("retro:")) | .text'
```

`bd export`, never `bd list` plus `bd show --json`: `bd list` defaults to open
issues and a drained run is almost entirely closed, and `bd show --json` omits
comments entirely even though its text output prints them.

**A zero count across a drained run is a broken query, not a clean run.**
Stop, check the scope and the prefix spelling, and confirm the beads really
are closed, before aggregating anything. This has happened.

**A zero count across a run that closed nothing may be genuine** — and the
retro still runs, on whatever the grooming and readiness stages produced.

---

### 3 · Route — mechanical, and only routing

**Who:** nobody. It is a lookup on the kind.

```
kind == tooling          → branch B
kind == proof            → the report, attached to what it supports. No issue.
kind == environment      → the environmental-failure ledger. No issue.
kind is one of the six   → branch A
kind missing or unknown  → UNROUTED — reported as a count and a list, never dropped
```

**`proof` and `environment` terminate here.** They are captured, reported and
never reach stage 4 — there is nothing for a human to decide about a negative
control or a third-party parser's grammar. Measured at 7 of 73, so this is not
a rare path.

**Nothing about the finding's content is read here.** That is the whole point
of stage 1 carrying the kind: a routing step that requires reading is a
routing step that gets skipped on a 73-finding run.

**Unrouted findings are the taxonomy's own health metric.** Their count is
reported every run; a rising count means the taxonomy is drifting from what
agents actually hit, and the next validation pass has its material ready.

**Duplicate detection is NOT here.** Deciding that two findings are the same
problem requires reading and judging both — that is reasoning, and it belongs
to stage 4. Stage 3 sorts the mail; it does not read it.

---

### 4 · Reason — findings become candidate issues

**Who:** the retro agent. **This is the stage where all the judgement lives**,
and it runs separately for each branch because the two produce different
artifacts for different people.

**Input:** the routed findings for one branch.
**Output:** every finding verbatim, plus N candidate issues.

#### (a) Every finding, verbatim, grouped by kind

Not a summary, not a selection. **The complete set is the substrate recurrence
detection runs on** — a finding filtered out once can never be observed
appearing twice, and recurrence is the only signal that separates a real
defect from one bad afternoon.

#### (b) N candidate issues, on top of it

This is where reasoning converts findings into actions: **cluster** findings
that are the same problem seen from different beads, **dedupe** against what
is already filed, and **rank**.

| field | |
|---|---|
| proposed title | the change, not the symptom |
| branch | A or B, from stage 3 |
| priority | proposed, P0–P4 |
| one-paragraph summary | what is wrong and what it costs |
| finding ids | **which findings it rests on** — what makes it auditable |
| recurrence | how many prior sprints carry an unfiled finding of the same kind and target |

**Candidates are proposals, not decisions.** The retro agent files nothing.

**Nothing is filed here, for either branch.** Producing the artifact before
the human has decided is what made the old flow five hops long.

---

### 5 · Decide — one human, once, at maximum context

**Who:** the human.
**Produces:** a yes/no per candidate.

**Asked once**, immediately after the report, while the run's context is still
loaded and the evidence is on screen. Per *candidate*, not per finding — one
or two orders of magnitude fewer questions.

**Declining is a normal outcome and is recorded.** A declined candidate stays
in the report as findings; nothing is deleted, and a later sprint hitting the
same friction raises it again with a higher recurrence count. That is how a
decision the human got wrong corrects itself with nobody tracking it.

**This is the only stage that waits on a person.** Skipped — an unattended
run, a human who never answers — the report still exists complete and the
candidates are decidable later from it. **The run does not block.**

---

### 6 · File — one format, two destinations

**Both branches produce the same artifact: an issue.** The branches differ in
*where it goes*, not in what it is.

| branch | destination |
|---|---|
| **A** | a bead in the project's own tracker — filed directly |
| **B** | an issue file in **`.sf/outbox/<skill>/`** in the project, for delivery here |

```bash
# Branch A — the issue template rendered straight into the tracker
bd create --title "<title>" --priority <n> \
  --description "<the issue template body>" \
  --discovered-from <origin-bead>
```

The `discovered-from` edge and the report citation let a reader six months
later ask "where did this come from" and get an answer.

**Never both, and never the other's.** A downstream project files nothing in
this repository's tracker, and this repository files nothing in theirs.

#### Mark the findings a filed issue rests on

**The moment a candidate is filed, every finding it rests on is marked with
the id it became:**

```bash
bd comment <finding's bead> "retro:filed:<new-id> — <finding id>"
```

Two things depend on this, and neither works without it:

- **Recurrence stays bounded.** Stage 4 counts *unfiled* findings only, so the
  set it scans shrinks as issues get filed instead of growing without limit.
- **The audit closes.** A finding leads to the issue it produced, and an issue
  leads back to its findings.

---

## The two templates

**Two artifacts, and only two.** A finding is what an agent records at the
moment of friction. An issue is what a human decides to act on. Everything the
old pipeline had between them was ceremony.

**Both branches use both templates, unchanged.** That is the harmonization:
one shape for evidence and one shape for action, wherever the work lands.

### Both become formal template documents

**Yes — authored the way this framework authors every other template:**
frontmatter carrying the machine-readable fields, a fixed section list, and an
HTML instruction block per section that the author deletes once the section is
written. `[[template-authoring-standards]]` governs both, and
`scripts/check-template-drift.py` governs them like any other.

| template | lives in | governed by |
|---|---|---|
| **finding** | `specs/standards/` — it is a *format*, and one skill does not own it | the capture protocol standard, amended |
| **issue** | `plugin/skills/spec/references/issue-template.md`, vendored to the skills that produce one | the `spec` skill, like every other template |

**The finding template governs a grammar, not a document** — its instances are
`bd` comments, not files. That is not a special case: **the work-order template
already works this way**, and says so ("it materializes into a tracker issue
and stops existing as a document"). The template is the canonical FORMAT the
producing skill targets and validates against; the record lives in the tracker.

The full field lists follow, so the bead that writes these files has nothing
left to decide.

---

### Template 1 — the finding

Written by the implementing agent, mid-task, as one `bd comment`. It has to be
cheap enough to write while working and complete enough to read cold, weeks
later, by an agent that was never in the session.

```
retro:<kind>[:<target>] <one line: what happened>

evidence:
  <the quote, the path:line, the command and its output, or expected vs observed>
context:
  <what was being attempted when this was hit>          (optional)
```

| field | required | notes |
|---|---|---|
| `kind` | **yes** | one of the seven. It is what routes the finding, and it costs a word |
| `target` | where there is one | the skill/script (branch B) or the document/bead (branch A). Omitting it costs a cluster, not a finding |
| headline | **yes** | one line, what happened. Not what it means |
| `evidence` | **yes** | see below — this is the part that does not survive the session |
| `context` | no | what was being attempted. Include it when the headline alone would not locate the problem |

**Evidence is at least one of, and preferably the most specific available:**

- the **exact quote** of the instruction, AC or spec text that misled, with
  where it is;
- a **`path:line`** anchor;
- the **command run and the output observed**;
- **expected versus observed**, stated as two concrete values.

**No length limit, and no analysis.** The two are separate rules and this
design keeps them apart deliberately:

> **Record everything about *what happened* that will not survive this
> session. Record nothing about *what it means*.**

**Forbidden, still:** root cause, proposal, priority, "we should", and any
sentence about a fix. Those are the issue's, and an agent writing them
mid-task both derails its own work and reasons worse than the aggregator will
with the whole run in view.

#### Finding ids are derived, never authored

`<bead-suffix>#<ordinal>` — the ordinal being the finding's position among
that bead's `retro:` comments, ordered by `created_at`. So `92u#2` is the
second finding on bead `…-92u`.

- **The capturing agent assigns nothing.** It cannot: the comment's identity
  does not exist until the write lands.
- **The ordering is stable because findings are append-only** — the capture
  protocol already says findings are never deleted, which is what makes an
  ordinal safe to cite.
- **This scheme is already in use.** The 2026-08-14 sprint report cites `92u#2`
  and `92u#3`; nothing new is being invented, it is being made a rule.

`bd`'s own comment UUID is the durable underlying key where a machine needs
one; the ordinal is the human- and citation-facing id.

---

### Template 2 — the issue

Produced at stage 4, decided at stage 5, filed at stage 6. **One issue is one
actionable change.**

```markdown
---
title: "<the change, not the symptom>"
description: <one sentence — what changes and why>
slug: issue-<identifier>
                                     # No `type` property: like a work order,
                                     # an issue materializes into a tracker
                                     # record and stops being a document.
audience: skill                      # project | skill — from ROUTE, not judged
target: <skill or document name>     # what the fix lands in
priority: 2                          # 0 (critical) .. 4 (backlog), PROPOSED
updated_at: <YYYY-MM-DDTHH:MM:SSZ>

# Evidence — what this issue rests on
findings: []                         # finding ids, e.g. ["92u#2", "1ilc.15.7#1"]
recurrence: 0                        # prior sprints in the window carrying an
                                     # UNFILED finding of this kind and target
report: <path>                       # the sprint report that carries them

# Provenance — who and what produced this
run: <run epic id>
branch: <branch the run worked on>
skill_version: <what was being exercised>
model: <model that produced the findings>   # load-bearing; never omit
---

<!--
Template: Issue
Version: 1.0
Source: This specification framework. Conventions follow
  [[template-authoring-standards]].

Sections (complete each, then delete its instruction block)
  - What is wrong, and what it costs
  - Evidence
  - Proposed change

Purpose
  ONE issue is ONE actionable change. It is the unit of decision, which is
  why it is not bundled into a multi-proposal document.

Both audiences use this template unchanged. `audience: project` renders
straight into the project's tracker; `audience: skill` is written to
`.sf/outbox/<target>/` for delivery. Same fields, two encodings.
-->

## What is wrong, and what it costs

<!-- One paragraph. State the cost explicitly: an issue that says what is
wrong but not what it costs cannot be prioritised against anything, which is
how a queue reaches 27. -->

## Evidence

<!-- One bullet per finding, quoting its headline. EVERY claim in this issue
traces to a finding id. A claim with no finding behind it does not belong in
the issue — it belongs in a new finding. -->

- `<finding-id>` — <the finding's headline, quoted>

## Proposed change

<!-- The concrete edit. For a document: the text to change and what it
becomes. For code or a script: the behaviour to change. Specific enough that
someone who did not run the sprint could make it. -->
```

**Why each field is there:**

- **audience** — decides the destination, and is already computed by stage 3.
- **recurrence** — the difference between a real defect and one bad
  afternoon. Without it every issue looks equally urgent.
- **cost** — an issue that says what is wrong but not what it costs cannot be
  prioritised against anything, which is how a queue reaches 27.
- **evidence by finding id** — the audit runs both ways: an issue leads to its
  findings, and a filed finding leads to the issue it became.
- **provenance** — carried forward from the RFC format, where it earned its
  keep. **The model is load-bearing**: findings differ systematically by the
  model that produced them, and an issue that hides which model hit the
  friction cannot be weighed against one from a different tier.

**Branch A renders this straight into a bead** — title, priority, description,
`discovered-from`. **Branch B serializes it as a file** for delivery, because
the tracker that should hold it is in another repository. **Same fields, two
encodings.** It is one template.

---

### The RFC format is retired

**Not "retired for branch A" — retired.** The harmonization above leaves it
with nothing to do, and its own bookkeeping says why:

> "Outcome counts are per proposal, not per RFC — **the proposal is the unit
> of decision**." — `rfcs/retro-log.md`, line 7

**An RFC is a document containing several issues.** That is the structural
defect. Triage has to decide per-proposal inside a document, the log has to
count in a different unit than it files in, and a single RFC can be
simultaneously adopted, deferred and rejected. Making the issue the unit
collapses that entire layer.

**What is kept**, because it earned its place: Summary → *What is wrong and
what it costs*. Evidence with citations → *Evidence*. The per-proposal
**Change** block → *Proposed change*. Provenance → *Provenance*, model
included.

**What is dropped:** *Considered and Not Proposed*, *Non-Goals*, *Adoption*
and *Resolution*. All four are triage-workflow state, and a tracker already
holds workflow state — Resolution is the issue's close reason, Adoption is
the issue itself. Writing them into a document is keeping a second copy of
the tracker in prose.

---

### `triage-rfc` becomes `triage-issue`

**It survives as its own skill, rebranded and refined.** It stops parsing
multi-proposal documents — there are none — and takes issues instead.

**Its most critical role sits downstream of triage itself: translating a
reported issue into a groomed, high-quality work order.** An issue is written
by whoever hit the problem, in whatever detail they had. A bead has to be
dispatchable to a fresh agent with zero context. Closing that gap is real
work, and it is the same bar `groom-beads` already enforces — so
`triage-issue` invokes it rather than restating it.

Then **bundle the accepted set into an epic**, which is exactly what
`sprint-scope` takes and the sprint orchestrator swarms. That closes the loop
the whole design exists to close: friction in one run becomes groomed work in
the next.

**It accepts issues from any channel**, which is what keeps the delivery
mechanism from being load-bearing:

- `.sf/outbox/<skill>/` files swept from a downstream project;
- GitHub issues, where someone reports one directly;
- issues this repository's own retro produced;
- whatever channel is added later.

**Triage is stage 5 for issues that arrive from outside.** That is the
symmetry worth seeing: a retro decides which of *this run's* candidates make
the cut; triage decides which of the *incoming* ones do. Same judgement, same
human, same "declining is normal and is recorded" — the only difference is
where the issue came from.

So its existing machinery is retained because it earns its keep on any
channel: dedupe against prior work, defer with a reason, accumulate recurrence
across reporters. **What it does not do is decide alone** — a human still says
yes or no, and `triage-issue` makes each answer cheap by arriving with the
work order already drafted.

---

### Delivering a branch-B issue — `.sf/outbox/<skill>/`

A downstream project's run produces issues about **this** repository's skills.
It cannot file them here — it has no access and should not need any — so
something carries them.

**Decided: the project writes them to `.sf/outbox/<skill>/` and its owner
forwards them.** Same mechanism the RFC inbox used, which is the point: it is
known to work, needs no infrastructure, no token and no network, and gives a
downstream project's owner a chance to read what is being sent on their behalf
before it goes.

```
<project>/.sf/outbox/
  sprint-wave/
    2026-08-22-worktree-base-ref-check-is-unreachable.md
  groom-beads/
    2026-08-22-tier-reason-has-no-stated-bar.md
```

- **One file per issue**, in the issue template's format.
- **Named `<date>-<slug>.md`**, so a directory listing is chronological and a
  filename says what the issue is.
- **Under `<skill>/`**, so the target is legible without opening anything —
  the same shape as `rfcs/inbox/<skill>/`.
- **`.sf/outbox/` is the project's, not ours.** Emptying it, forwarding it, or
  ignoring it entirely are all the project's call.

**What was rejected, and why it matters:** a bead in the project's own tracker
labelled for export. It uses infrastructure already there and is the least
work, which is exactly why it is tempting — and it is the backlog pollution
the audience split exists to prevent. A project must never carry issues it
cannot act on.

**Not chosen yet, and deliberately deferred:** opening a GitHub issue directly
against `agent-plugins`. It removes the human hop, and it needs a token, a
network path, and a policy about what a downstream run may do on our behalf.
Worth doing later; not worth blocking on now. **`triage-issue` accepts issues
from any channel** (below), so adding one later changes no downstream skill.

**This repository is its own downstream today**, so its outbox is swept into
its own tracker in the same run.

---

## Recurrence at scale — bounded by construction

240 findings today; five sprints. Scanning all history every run does not
survive fifty sprints. Three mechanisms keep it bounded, and all three are
needed:

1. **Filed findings leave the pool.** The `retro:filed:` marker above. A
   finding that produced an issue is answered, and answered findings are not
   evidence of an unaddressed recurring problem.
2. **The scan is windowed.** Recurrence looks back **N sprints, not all
   history** (proposed: 5). A problem that last appeared twenty sprints ago
   and has not recurred since is not recurring.
3. **The match key is `<kind>:<target>`, not prose.** Counting is a group-by,
   not a semantic comparison over hundreds of records.

**This starts at zero and becomes useful from the second sprint after
adoption.** The 240 findings already on disk carry no kind and no target, so
they cannot be backfilled — worth stating plainly rather than discovering.

---

## What this supersedes, and what it retains

**Every piece of the current specification gets a verdict.** A rule nobody
retired is a rule some agent still follows.

| current | verdict |
|---|---|
| `specs/standards/findings-capture-protocol.md` — the finding, when to record, where to record | **RETAINED, amended.** The prefix gains `<kind>[:<target>]`; the six triggers become the six branch-A kinds; the 1–3 sentence cap is **replaced** by "complete evidence, no analysis". Friction-only and sensor-not-analyst are unchanged and are what is most worth keeping |
| — its "Aggregation" section | **SUPERSEDED** by stages 2–4. Aggregation is no longer a paragraph inside a capture protocol; it is three stages with contracts and a mechanical/agentic boundary |
| `bead-bonanza`'s "The sprint retro" — the five questions | **SUPERSEDED** by stages 2–6, and it moves out of the orchestrator into `sprint-retro`. Question 4 ("where did this skill mislead") becomes the `tooling` kind, recorded at capture rather than reconstructed at the end |
| `retro-skill` — the three finding shapes, the provenance rule, the bar | **RETAINED**, as branch B's stage-6 path. Its input becomes routed `tooling` findings rather than a fresh reading of the run. **The model-provenance rule is load-bearing and stays** |
| `retro-spec` | **SPLIT.** Findings about the spec *templates and standards this repository vends* are branch B; findings about a *project's own* FRDs, blueprints and ADRs are branch A. The merge test decides. Both now produce issues, not RFCs |
| **The RFC format** — `rfc-template.md`, `skill-rfc-template.md`, the 8-section document | **RETIRED**, for both branches. An RFC is a document containing several issues, which is why its own log counts per proposal and not per RFC. Summary, Evidence, the per-proposal Change block and Provenance carry forward into the issue template; the four workflow sections are dropped as a second copy of the tracker |
| The `origin: self` RFC path | **RETIRED** with the format. Branch A becomes beads in the project; branch B becomes issues delivered here |
| `triage-rfc` → `triage-issue` | **DONE.** It stops parsing multi-proposal documents and takes issues from any channel — outbox, GitHub, this repository's own retro. Its critical role becomes translating a reported issue into a groomed work order (via `groom-beads`) and bundling the accepted set into an epic the sprint can swarm. Dedupe, defer-with-a-reason and recurrence are retained |

---

## The 27 queued RFCs

All 27 are `origin: self`. Under the split they divide by what they are about,
and **not by the directory they sit in** — `rfcs/inbox/feature-blueprint/`
holds proposals about the blueprint *template* (branch B), while some under
`rfcs/inbox/frd/` are about this repository's own FRDs (branch A).

**They are the last documents in the old format**, so each is migrated as
well as decided. Apply the merge test, then one of three verdicts — **per
proposal, not per RFC**, since that was always the real unit:

| verdict | what happens |
|---|---|
| **superseded** by this renovation | closed, pointing at the bead that supersedes it |
| **still wanted** | rewritten as an **issue** and filed — branch A into the project's tracker, branch B here. Never both, and never left as an RFC |
| **rejected** | closed with a reason |

A single RFC may split across all three verdicts. That is not an awkward edge
case, it is the format's defect showing itself one last time.

Recorded in `rfcs/retro-log.md`. Afterwards the inbox is empty, or every
remaining item carries a stated reason for waiting.

**Read the eight against `bead-bonanza` and its helpers first** — wave 8
restructured all of them, and several propose changes to sections that no
longer exist.

---

## Open — what this draft does not settle

1. **One ratio is not a trend.** The gate measured branch A at 82% and branch B
   at 7% — on a run that implemented a *feature*. A run whose subject is the
   skills themselves would split very differently, and nothing here establishes
   what a normal ratio looks like or when branch B has shrunk enough to
   deprioritise.

2. **The recurrence window is a guess.** Five sprints is proposed, not
   measured. The right number is whatever spans "long enough that a real
   recurring problem shows up twice"; nothing on disk answers that yet.

3. **Nothing here sizes branch B's decline.** The design says branch B shrinks
   as the skills mature and that this should be visible in the branch ratio
   per run. What ratio would mean "mature enough to stop prioritising branch
   B" is not decided, and probably should not be until several runs have
   produced one.

4. **The merge test is clean on paper and untested under load.** It is
   unambiguous — where does the fix land — but whether an implementing agent
   applies it correctly mid-task, at the moment of friction, is exactly the
   kind of thing the first validated run should measure.

5. **Branch-B delivery has no mechanism.** Four candidates are tabled above
   and none is chosen. It is not blocking: this repository is its own
   downstream today, so its issues are filed here directly. It becomes
   load-bearing on the day a second project runs a sprint.

6. **`triage-issue` and stage 4 both dedupe.** Stage 4 dedupes a run's own
   candidates against what is already filed; `triage-issue` dedupes an arriving
   issue against prior work. Different scopes, probably the same query.
   Whether one is expressed in terms of the other or they share a script is
   not decided, and blocks neither.

7. **Nothing here says how an issue is revised.** An RFC could be edited in
   place before triage. A filed bead is edited in the tracker, which is fine
   for branch A — but a branch-B issue that has been delivered and then
   proves wrong has no stated correction path.
