# The retro, end to end

**Status: DRAFT for review.** Everything else in wave 5 is decided inside
this document — the finding taxonomy, the candidate-issue flow, the audience
split in three skills, and what happens to the 27 queued RFCs.

---

## Objective

A sprint should get measurably better at running sprints, and a spec corpus
measurably better at answering questions, **as a by-product of doing the
work** — not as a second project someone has to staff.

Today it does neither, and the reason is not that findings are missing. There
are 240 of them. The reason is that every stage after capture is a human
decision point, and the queue behind those points is now longer than anyone
will ever drain.

---

## What the evidence says

Six measurements. Each names where it came from; the whole design follows from
them, so they come first.

| | measured | source |
|---|---|---|
| **1** | **240 findings** harvested across five sprints | `rfcs/retro-log.md` harvest counts (73 + 14 + 16 + 64 + 73) |
| **2** | **27 RFCs queued**, 25 ready-for-review, 2 draft | `find rfcs/inbox -name '*.md' \| wc -l` |
| **3** | **All 27 carry `origin: self`** | `grep -h '^origin:' rfcs/inbox/*/*.md \| sort \| uniq -c` → `27 origin: self` |
| **4** | **93 proposals processed, zero from outside this repository** | `rfcs/retro-log.md`, closing note |
| **5** | The protocol's **six categories are never recorded** — every retro re-derives each finding's kind by reading it | the `retro:` prefix carries no kind; `specs/standards/findings-capture-protocol.md` lists six triggers and no way to record which fired |
| **6** | Findings that became a bead did so through **five hops**: finding → report → RFC → triage → bead | `rfcs/retro-log.md` disposition column |

### What follows from them

**Measurement 4 is the one that reframes the whole pipeline.** The RFC format
exists to accumulate recurrence *across projects* — that is what earns its
evidence-and-proposals structure and `triage-rfc`'s dedupe-and-defer
machinery. Ninety-three proposals in, it has never carried one. So the
cross-project machinery is not being under-used; **it is being used for a
purpose it was not built for**, on findings that had a tracker available the
whole time.

That does not mean retire it. It means **stop routing this repository's own
project-level findings through it** — which is D11's audience split, now with
a number behind it.

**Measurement 5 is why routing is expensive.** A finding's kind is knowable at
the moment of capture, by the agent that hit the friction, at a cost of one
word. It is instead re-derived later, by an agent reading 73 findings cold.
Every downstream stage pays for that omission.

**Measurement 6 is where the latency lives.** Five hops, four of them human
decision points, for a finding whose destination was determinable at hop one.

---

## The organising principle: whose problem is it?

D11, stated as the axis the whole flow turns on.

| the finding is about… | audience | destination | why |
|---|---|---|---|
| **the user's project** — a spec that was wrong, a defect the sprint surfaced, work the sprint implied | the project itself | **a bead in that project's tracker** | the audience has a tracker and is the same people. No proposal document is needed to tell them about their own repository |
| **a skill this repository vends** — a rule that misled, a script that failed, a template with a gap | every project using the plugin | **an RFC in this repository** | it is a proposal about a *product shipped elsewhere*, it needs the evidence-and-proposals structure, and recurrence across projects is exactly what `triage-rfc` accumulates |

**Two consequences worth stating plainly:**

1. **A downstream project must never accumulate RFCs about the plugin's own
   skills.** Only this repository can act on those, and a backlog full of
   proposals its owners cannot implement is noise that trains them to ignore
   the channel.
2. **This repository is both audiences at once**, which is why the split has
   been invisible here. Every finding had somewhere plausible to go, so
   nothing forced the question. A second project using the plugin would have
   surfaced it immediately.

---

## The six stages

```
1 CAPTURE ....... the implementing agent, per finding, at the moment of friction
2 HARVEST ....... one query over the run's closure set
3 CLASSIFY ...... mechanical, from the kind and target the capture recorded
4 REPORT ........ every finding verbatim + N candidate issues
5 DECIDE ........ the human, ONCE, at the point of maximum context
6 FILE .......... beads here, RFCs there — by audience, not by judgement
```

**Stages 1–4 are mechanical or agentic and never wait on a person. Stage 5 is
the single human decision point. Stage 6 is mechanical given 5.**

That is the whole design. What follows is each stage's contract.

---

### 1 · Capture — the agent is a sensor

**Who:** the implementing agent, during the work.
**Produces:** one `bd comment` per finding on the bead it happened to.
**Cost:** one line, at the moment the friction is felt.

Unchanged from the Findings Capture Protocol in one respect and changed in
one. Unchanged: **1–3 sentences, friction only, no root-causing, no
proposals.** The capturing agent has the context and lacks the perspective;
asking it to analyse produces worse findings *and* fewer of them.

Changed: **the prefix carries the kind and, where there is one, the target.**

```
retro:<kind>[:<target>] <what happened + what caused it>
```

Because measurement 5: the kind is free at capture and expensive afterwards.

#### The taxonomy — seven kinds

Six are the protocol's existing trigger list, which already enumerates what an
implementing agent notices. The seventh is the class it does not name.

| kind | fires when | audience |
|---|---|---|
| `gap` | information was needed that the work order and its references did not carry | project |
| `decision` | a design decision had to be made that the item or blueprint should have pinned | project |
| `stale` | the item contradicted repo reality — moved code, a dead anchor, a changed element | project |
| `unverifiable` | an AC or citation obligation could not be satisfied as declared | project |
| `sequencing` | a broken predecessor, or a blocker the dependency graph did not encode | project |
| `scope` | the item was mis-scoped — too large to review as one unit, trivial, or overlapping another | project |
| `tooling` | **the seventh** — a skill, script, template or protocol misled, failed, or had no rule where one was needed | **skill** |

**`target` is what makes stage 3 mechanical.** For a `tooling` finding it is
the skill or script name (`retro:tooling:sprint-wave`). For the six project
kinds it is the document or bead the finding is about, where there is one
(`retro:gap:FRD-014`). Omit it when there genuinely is not one; an omitted
target costs a classification, not a finding.

**`incident` is not a kind here.** An incident — a decision that changed what
"correct" means — is already recorded as `retro: incident —` by `run-bead` and
routed to the report's first section. It reaches this flow at harvest like any
other finding and classifies as `decision`, but its *reporting* obligation is
the sprint's, not the retro's.

#### Validation gate for this taxonomy

**This design is not settled until the taxonomy is tested against real
findings.** Classify all 73 findings in the 2026-08-15 report, count the
residue that fits no kind, and record the number here.

- Residue **≤ 5%** → adopt as written.
- Residue **> 5%** → the residue names the missing kind. Add it, do not widen
  an existing one; a kind that absorbs anything routes nothing.

**Do not skip this gate.** A taxonomy validated against its own author's
intuition is measurement 5 again with more words.

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
are closed, before aggregating anything. This has actually happened.

---

### 3 · Classify — mechanical, or it does not happen

**Who:** nobody. It is a lookup.
**Produces:** each finding tagged `project` or `skill`, with its target.

```
kind == tooling            → skill audience,   target = the skill/script
kind is one of the six     → project audience, target = the document/bead
kind missing or unknown    → UNCLASSIFIED — surfaced in the report, never dropped
```

**Unclassified findings are reported as a count and a list.** They are the
measurement that tells the next run whether the taxonomy is holding, and
dropping them silently is how a taxonomy stays wrong.

**Nothing about the finding's *content* is read at this stage.** That is the
point of stage 1 carrying the kind: classification that requires reading is
classification that gets skipped on a 73-finding run.

---

### 4 · Report — every finding verbatim, then the candidates

**Who:** the retro agent.
**Produces:** two things, in the sprint report.

**(a) Every harvested finding, verbatim, grouped by kind.** Not a summary, not
a selection. **The complete set is the substrate recurrence detection runs
on** — a finding filtered out once can never be observed appearing twice, and
recurrence is the only signal that distinguishes a real defect from one bad
afternoon.

**(b) N candidate issues, on top of it.** Each carries:

| field | |
|---|---|
| proposed title | what the bead or RFC would be called |
| audience | `project` or `skill`, from stage 3 |
| priority | proposed, P0–P4 |
| one-paragraph summary | what is wrong and what it costs |
| finding ids | **which findings it rests on** — this is what makes it auditable |
| recurrence | how many prior sprints carry a finding of the same kind and target |

**Candidates are proposals, not decisions.** The retro agent does not file
them, and does not rank them beyond a proposed priority.

**No RFC is authored at this stage, for either audience.** An RFC is a stage-6
artifact and only for the skill audience. Authoring one here is what made the
old flow five hops long.

---

### 5 · Decide — one human, once, at maximum context

**Who:** the human.
**Produces:** a yes/no per candidate.

**Asked once**, immediately after the report, while the run's context is still
loaded and the evidence is on screen. Not per finding — per *candidate*, which
is one or two orders of magnitude fewer.

**Declining is a normal outcome and is recorded.** A declined candidate stays
in the report as a finding; it is not deleted, and a later sprint that hits the
same friction will raise it again with a higher recurrence count. That is the
mechanism by which a decision the human got wrong corrects itself without
anyone tracking it.

**This is the only stage that waits on a person.** If it is skipped — an
unattended run, a human who never answers — the report still exists complete,
and the candidates can be decided later from it. **The run does not block.**

---

### 6 · File — by audience, mechanically

**Who:** the retro agent, from stage 5's answers.

**Project audience → a bead in the project's tracker:**

```bash
bd create --title "<candidate title>" --priority <n> \
  --description "<summary>. Findings: <ids>. Report: <path>" \
  --discovered-from <origin-bead>
```

The `discovered-from` edge and the report citation are what let a reader six
months later ask "where did this come from" and get an answer. **No RFC.**

**Skill audience → an RFC in *this* repository's `rfcs/inbox/<skill>/`.**
The evidence-and-proposals structure earns its cost here: the proposal is
about a product shipped to other projects, and `triage-rfc` accumulates
recurrence across them.

**Never both, and never the other one's.** A downstream project files no RFCs;
this repository files no beads in someone else's tracker.

---

## What this supersedes, and what it retains

**Every piece of the current specification gets a verdict.** Nothing is left
in an undecided state, because a rule nobody retired is a rule some agent
still follows.

| current | verdict |
|---|---|
| `specs/standards/findings-capture-protocol.md` — the finding, when to record, where to record | **RETAINED**, amended: the prefix gains `<kind>[:<target>]`, and the six triggers become the six named project kinds. The 1–3 sentence, friction-only, sensor-not-analyst rule is unchanged and is the thing most worth keeping |
| — its "Aggregation" section | **SUPERSEDED** by stages 2–4. Aggregation is no longer one paragraph inside a capture protocol; it is four stages with contracts |
| `bead-bonanza`'s "The sprint retro" — the five questions | **SUPERSEDED** by stages 2–6, and it moves out of the orchestrator into `sprint-retro`. Question 4 ("where did this skill mislead") becomes the `tooling` kind, recorded at capture rather than reconstructed at the end |
| `retro-skill` — the three finding shapes, the provenance rule, the bar | **RETAINED**, and it becomes stage 6's skill-audience path. Its input becomes the classified `tooling` findings rather than a fresh reading of the run. **The model-provenance rule is load-bearing and stays** |
| `retro-spec` — the workflow, the RFC structure | **RETAINED** for the skill audience — specifically, findings about the *spec templates and standards this repository vends*. **Its project-audience use is superseded**: a finding about the user's own FRD is now a bead in the user's tracker |
| `triage-rfc` | **RETAINED, and rescoped**: cross-project proposals only. Stated in the skill, so a reader does not have to infer it from what happens to arrive |
| The `origin: self` RFC path for **project-level** findings | **RETIRED.** It is measurement 4 and measurement 6 together: a cross-project format carrying findings that never left the project, at four decision points of cost |
| The `origin: self` RFC path for **skill-level** findings | **RETAINED.** This repository *is* the upstream for its own skills, so a self-authored RFC about a skill is a genuine cross-project proposal that happens to originate at home |

---

## The 27 queued RFCs

All 27 are `origin: self` (measurement 3). Under the split they divide by what
they are about, and **the division is not the directory they are filed in** —
`rfcs/inbox/feature-blueprint/` holds proposals about the blueprint *template*,
which is skill audience, and `rfcs/inbox/frd/` holds some that are about this
repository's own FRDs, which is project audience.

**Each gets one of three verdicts**, and the queue is not empty until every one
has had one:

| verdict | what happens |
|---|---|
| **superseded** by this renovation | closed, with a pointer to the bead that supersedes it |
| **still wanted** | filed as a bead (project audience) or kept as an RFC (skill audience) — never both |
| **rejected** | closed with a reason |

Recorded in `rfcs/retro-log.md`. The inbox afterwards is empty, or every
remaining item carries a stated reason for waiting.

**The eight against `bead-bonanza` and its helpers deserve reading first**:
wave 8 restructured all of them, and several propose changes to sections that
no longer exist.

---

## Open — what this draft does not settle

1. **The taxonomy's residue is unmeasured.** The 73-finding validation in
   stage 1 has not been run. Until it is, seven kinds is a proposal.

2. **Recurrence needs a query and nobody has written it.** Stage 4 asks for
   "how many prior sprints carry a finding of the same kind and target", which
   is only cheap once `<kind>:<target>` exists — so it cannot be measured
   against the 240 findings already on disk, none of which carry one.
   **Recurrence therefore starts at zero and becomes useful from the second
   sprint after adoption.** Worth stating rather than discovering.

3. **`retro-skill` and `retro-spec` overlap at the boundary.** A finding about
   a spec *template* this repository vends is skill audience; a finding about a
   spec *document* in a project is project audience. Same words, different
   destinations. The rule is clear; whether an agent applies it correctly under
   load is not established.

4. **Nothing here says when the retro runs on a run that closed nothing.** A
   sprint that bounced its whole queue produced no `retro:` comments and has
   no findings, but it is exactly the kind of run worth learning from. Probably
   a `null result` report; not decided.
