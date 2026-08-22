# The retro, end to end

**Status: DRAFT for review.** Everything else in wave 5 is decided inside
this document — the finding taxonomy, the candidate-issue flow, the audience
split in three skills, and what happens to the 27 queued RFCs.

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
**Destination:** **RFCs here.**

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
| the fix lands in **`agent-plugins`** | **branch B** — an RFC here |

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

Six measurements, each naming the command or source behind it. Read them
through the two-branch lens: several say something different about branch A
than about branch B.

| | measured | source |
|---|---|---|
| **1** | **240 findings** harvested across five sprints | `rfcs/retro-log.md` harvest counts (73 + 14 + 16 + 64 + 73) |
| **2** | **13 RFCs processed; 63 proposals adopted**, 4 deferred, 1 rejected. Processed RFCs were deleted, not archived | `rfcs/retro-log.md` outcome column, summed |
| **3** | **27 RFCs still queued**, 25 ready-for-review, 2 draft | `find rfcs/inbox -name '*.md' \| wc -l` |
| **4** | **All 27 carry `origin: self`**; 93 proposals processed with **no reporter outside this repository** | `grep -h '^origin:' rfcs/inbox/*/*.md \| sort \| uniq -c` → `27 origin: self`; `rfcs/retro-log.md` closing note |
| **5** | The protocol's **six categories are never recorded** — every retro re-derives each finding's kind by reading it | the `retro:` prefix carries no kind; `specs/standards/findings-capture-protocol.md` lists six triggers and no way to record which fired |
| **6** | A finding that became a bead took **five hops**: finding → report → RFC → triage → bead | `rfcs/retro-log.md` disposition column |

### What follows from them

**The RFC pipeline works — for branch B.** Measurement 2 is the one to lead
with: 63 adopted proposals have shipped into these skills through it. The
evidence-and-proposals structure and `triage-rfc`'s dedupe-and-defer machinery
have earned their cost. **Nothing here retires that**, and given branch B's
current urgency, nothing should.

**Its queue grew because branch A was routed into it too.** Measurements 3 and
6 together: 27 unprocessed proposals, and five hops from friction to action
for findings whose destination was determinable at hop one. A finding about
*this repository's own specs* went out through a cross-project proposal format
and back in through triage — when a bead was available the entire time.

**Measurement 4 is about reporters, not about value.** Ninety-three proposals
processed, none from an outside project. That does not mean the format failed;
it means **the cross-project accumulation the format was built for has not
started yet**, and until it does, `triage-rfc`'s recurrence machinery is
serving a single reporter. It stays, scoped honestly.

**Measurement 5 is why routing is expensive.** A finding's kind — and
therefore its branch — is knowable at the moment of capture, by the agent that
hit the friction, for the cost of one word. It is instead re-derived later by
an agent reading 73 findings cold.

---

## The insight this design turns on

> **This repository's project-level findings stop going through the RFC
> process. They become beads, through the sprint's own retro, exactly like
> any adopting project's.**

Everything else follows. The queue stops growing from branch A. `triage-rfc`
handles only what it was built for. And this repository finally runs the same
loop it ships — which is the only way its defects get found.

---

## The six stages

```
1 CAPTURE ..... implementing agent, at the moment of friction        MECHANICAL
2 HARVEST ..... one query over the run's closure set                 MECHANICAL
3 ROUTE ....... branch A or B, from the kind the capture recorded    MECHANICAL
        ─────────── the branches diverge here ───────────
4 REASON ...... findings → candidate issues; dedupe, cluster, rank   AGENTIC
5 DECIDE ...... the human, ONCE, at the point of maximum context     HUMAN
6 FILE ........ beads there, RFCs here                               MECHANICAL
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

#### The taxonomy — seven kinds

Six are the capture protocol's existing trigger list, which already enumerates
what an implementing agent notices. The seventh is the class it does not name.

| kind | fires when | branch |
|---|---|---|
| `gap` | information was needed that the work order and its references did not carry | **A** |
| `decision` | a design decision had to be made that the item or blueprint should have pinned | **A** |
| `stale` | the item contradicted repo reality — moved code, a dead anchor, a changed element | **A** |
| `unverifiable` | an AC or citation obligation could not be satisfied as declared | **A** |
| `sequencing` | a broken predecessor, or a blocker the dependency graph did not encode | **A** |
| `scope` | the item was mis-scoped — too large to review as one unit, trivial, or overlapping another | **A** |
| `tooling` | **the seventh** — a skill, script, template or protocol misled, failed, or had no rule where one was needed | **B** |

**`target` is what makes stage 3 a lookup.** For `tooling` it is the skill or
script (`retro:tooling:sprint-wave`). For the six branch-A kinds it is the
document or bead the finding is about (`retro:gap:FRD-014`). Omit it where
there genuinely is none; that costs a cluster, not a finding.

**When in doubt about the branch, apply the merge test** — where would the fix
land? An implementing agent can answer that in one beat.

**`incident` is not a kind.** An incident — a decision that changed what
"correct" means — is already recorded as `retro: incident —` by `run-bead` and
leads the sprint report. It reaches this flow at harvest and classifies as
`decision`; its *reporting* obligation belongs to the sprint, not the retro.

#### Validation gate — unrun

**This taxonomy is a proposal until it is tested against real findings.**
Classify all 73 findings in the 2026-08-15 report, count the residue that fits
no kind, and record the number here.

- Residue **≤ 5%** → adopt as written.
- Residue **> 5%** → the residue names the missing kind. **Add it; never widen
  an existing one.** A kind that absorbs anything routes nothing.

Also record the branch split of those 73 — it is the first real measurement of
how much of a run's friction is branch A versus branch B, and that ratio is
what tells us when branch B has shrunk.

**Do not skip this gate.** A taxonomy validated against its author's intuition
is measurement 5 again with more words.

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
kind == tooling         → branch B
kind is one of the six  → branch A
kind missing or unknown → UNROUTED — reported as a count and a list, never dropped
```

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
| proposed title | what the bead or RFC would be called |
| branch | A or B, from stage 3 |
| priority | proposed, P0–P4 |
| one-paragraph summary | what is wrong and what it costs |
| finding ids | **which findings it rests on** — what makes it auditable |
| recurrence | how many prior sprints carry an unfiled finding of the same kind and target |

**Candidates are proposals, not decisions.** The retro agent files nothing.

**No RFC is authored here, for either branch.** An RFC is a stage-6 artifact
and only for branch B. Authoring one at this stage is what made the old flow
five hops long.

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

### 6 · File — by branch, mechanically

**Branch A → a bead in the project's own tracker:**

```bash
bd create --title "<candidate title>" --priority <n> \
  --description "<summary>. Findings: <ids>. Report: <path>" \
  --discovered-from <origin-bead>
```

The `discovered-from` edge and the report citation let a reader six months
later ask "where did this come from" and get an answer. **No RFC.**

**Branch B → an RFC in `rfcs/inbox/<skill>/` in this repository.** The
evidence-and-proposals structure earns its cost here: it is a proposal about a
product shipped elsewhere, and `triage-rfc` is what accumulates recurrence
across the projects that eventually report.

**Never both, and never the other's.** A downstream project files no RFCs;
this repository files no beads in someone else's tracker.

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
| `retro-spec` | **SPLIT.** Findings about the spec *templates and standards this repository vends* are branch B and stay here as RFCs. Findings about a *project's own* FRDs, blueprints and ADRs are branch A and become beads there. The merge test decides |
| `triage-rfc` | **RETAINED, rescoped** to cross-project proposals about the skills — stated in the skill, so a reader does not infer its scope from whatever happens to arrive |
| The `origin: self` RFC path for **branch A** findings | **RETIRED.** This is the insight above: this repository's project-level findings become beads through the sprint's own retro, like any adopting project's |
| The `origin: self` RFC path for **branch B** findings | **RETAINED.** This repository is the upstream for its own skills, so a self-authored RFC about a skill is a genuine proposal that happens to originate at home — and it is the branch that matters most right now |

---

## The 27 queued RFCs

All 27 are `origin: self`. Under the split they divide by what they are about,
and **not by the directory they sit in** — `rfcs/inbox/feature-blueprint/`
holds proposals about the blueprint *template* (branch B), while some under
`rfcs/inbox/frd/` are about this repository's own FRDs (branch A).

Apply the merge test to each, then one of three verdicts:

| verdict | what happens |
|---|---|
| **superseded** by this renovation | closed, pointing at the bead that supersedes it |
| **still wanted** | branch A → filed as a bead; branch B → kept as an RFC. Never both |
| **rejected** | closed with a reason |

Recorded in `rfcs/retro-log.md`. Afterwards the inbox is empty, or every
remaining item carries a stated reason for waiting.

**Read the eight against `bead-bonanza` and its helpers first** — wave 8
restructured all of them, and several propose changes to sections that no
longer exist.

---

## Open — what this draft does not settle

1. **The taxonomy's residue is unmeasured.** The 73-finding validation has not
   been run. Until it is, seven kinds is a proposal — and so is the branch
   ratio it would also produce.

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
