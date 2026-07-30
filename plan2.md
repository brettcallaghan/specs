# Beyond Best Practices — A Step-Back Review

`plan.md` measures these skills against
[Anthropic's skill-authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices).
This document does something different: it reads the skills for **intent**
— what is this framework trying to be? — and asks where it falls short of
its own ambition, or where a structural change would make it markedly
better. Nothing here is a best-practice violation. Several items are
things the best-practices doc has no opinion about.

**What I think this framework is.** A spec-first development discipline
for a world where the implementers are agents, not people. The thesis is
sharp and unusual: agents fail not because they cannot code but because
they cannot ask, so the fix is to manufacture context-independence and
correct ordering *upstream* of execution. Everything follows from that —
EARS because a requirement must be testable by a machine that cannot
seek clarification; the Definition of Ready because a fresh subagent has
no session to lean on; the C4 blast radius because wave isolation needs a
collision key; the retro loop because the only honest evidence about a
spec's quality comes from the agent that consumed it.

That thesis is good, and the execution is more rigorous than most
internal frameworks I have seen. The gaps are not conceptual. They are
that **the framework does not yet apply its own discipline to itself.**

Ranked by leverage.

---

## Contents

1. Nothing enforces the invariants the skills assert
2. The self-improvement loop has never run
3. There is no skill-authoring standard, so the skills drift from house style
4. Model tiering is coupled to model names — the root cause of a whole bug class
5. The atomic unit of work has two names
6. The framework has no on-ramp
7. The frontmatter index is specified but does not exist — three checks are blocked on it
8. Durable specs have no drift gate, so they rot silently
9. `retro-skill` is the least mature skill and the most load-bearing
10. The `spec` skill is a router that costs 294 lines to route
11. No adoption ladder: the floor is as high as the ceiling
12. Version compatibility is unmapped
13. This repo does not dogfood `bead-bonanza`
14. Smaller items

---

## 1. Nothing enforces the invariants the skills assert

**The observation.** This framework's central belief is that quality
comes from gates, not diligence. It says so repeatedly and well:

- `groom-beads/SKILL.md:212` — *"Before declaring the set ready, validate
  — don't eyeball it."*
- `work-order-definition-of-ready.md:20` — *"Treat these as a gate, not a
  suggestion."*
- `groom-feature-blueprint/SKILL.md:20-22` — *"It is a quality firewall,
  not just a translator."*

And yet every invariant the framework asserts *about itself* is enforced
by nothing but authorial memory:

| Invariant | Asserted at | Enforced by |
|---|---|---|
| Model vocabulary is one closed set | 4 files (see plan.md P1) | nothing — and it has already broken |
| Spec-owned files are never forked | `groom-beads:45-57`, `groom-feature-blueprint:32-34` | nothing |
| The RFC template exists once | `spec:272` anti-pattern | nothing — and it has already broken (plan.md P5) |
| Normative change bumps the version | `template-authoring-standards:436-444` | nothing |
| Every `[[slug]]` resolves | `template-authoring-standards:717-720`, checklist #5 | nothing — ADR-012 admits this |
| 80-column lines | `template-authoring-standards:24` | nothing (plan.md P19) |
| No volatile facts / counts | `template-authoring-standards:666-680` | nothing — `spec:47` violates it |
| Backticked paths resolve | ADR-012:56-57 | nothing (plan.md P4) |

Two of these eight have *already* silently broken. That is the framework
diagnosing itself: prose invariants without a gate decay, which is
precisely the argument it makes for the Definition of Ready.

**Why this is the highest-leverage item.** Every Tier 1 and Tier 5 finding
in `plan.md` is an instance of this one. Fixing them individually leaves
the mechanism that produced them intact.

**Proposal.** Add `scripts/check-consistency.sh` — the framework's own
`bd-health.sh`, pointed at itself — and run it in CI and as a pre-commit
hook. Eight checks, none of them hard:

1. **Model vocabulary** — the enum in `bead.schema.json`, the set in
   `validate_beads.py`, and the list in `work-order-field-guide.md` are
   identical; no other file names a model outside it.
2. **No forked canon** — no file outside `spec/references/` contains a
   heading or first-line matching a spec-owned reference file (catches a
   copy before it drifts).
3. **TOC present** — every `*/references/*.md` over 100 lines has a
   `## Contents`.
4. **Path resolution** — no file under `*/references/` contains the
   literal `` `references/ ``; every backticked `*.md` path resolves from
   the containing file's directory.
5. **Cross-skill paths** — every `../<skill>/references/<f>` target
   exists.
6. **Version bump** — any commit touching a template, skill, or standard
   also changes its `Version:`/`version` field. (Warn, not fail — some
   commits are genuinely editorial.)
7. **No volatile counts** — flag `\b(two|three|...|ten|[0-9]+)\s+(types|
   phases|checks|documents|skills)\b` in normative prose.
8. **Line length** — prose lines ≤80 columns outside tables and fences.

Follow the framework's own script conventions: exit 0 clean / 1 hard /
2 bad invocation, verbose per-check output, `--strict` to promote
warnings — the same contract `bd-health.sh` and `validate_beads.py`
already use, so it feels native.

**Expected effect.** This is the change that makes every other change
stick. It also gives `triage-rfc` step 5 ("keep every touched file within
its own conventions") something mechanical to lean on instead of care.

---

## 2. The self-improvement loop is fully specified and has never run

**The observation.** The retro pipeline is the most original thing in this
repo:

```
run a skill → retro-skill → rfcs/outbox/ → (human) → rfcs/inbox/
           → triage-rfc → adopt / defer as a bead / reject → rfcs/retro-log.md
```

It has a severity rubric, a bar for what counts as evidence, provenance
that records model and effort, dedupe against prior resolutions, theme
labels so recurrence across projects accumulates into priority, and
rejection as a first-class recorded outcome. It is a genuinely good
design for learning from agent behavior at scale.

It has produced nothing. Measured:

| `triage-rfc`'s declared input | State |
|---|---|
| `rfcs/inbox/*.md` | directory does not exist |
| `rfcs/*.md` (resolved history, the dedupe source) | does not exist |
| `rfcs/retro-log.md` | does not exist |
| beads labelled `rfc-triage` | `.beads/interactions.jsonl` is 0 bytes |

`triage-rfc:50-52` refers to "rfc-0001 through rfc-0003" as existing
history. They are not in the repo. And `retro-skill:174-177` routes
in-repo retros "straight to `rfcs/inbox/` for triage" — into a directory
that is not there.

**Why it matters more than it looks.** Everything in this framework is
justified by the retro loop. `retro-skill`'s bar, the version-stamping
obligation in `template-authoring-standards:446-453`, the provenance
blocks, the theme labels — all of it exists so that triage can answer "is
this already fixed?" and "has a second project hit this?". None of that
machinery has ever been exercised, which means none of it has been
*tested*, which means the framework's own quality signal is unvalidated.
It is the one part of the system with no evidence behind it — in a
framework whose thesis is that evidence is the only thing worth arguing
from.

**Proposal.**

1. Create the skeleton so the declared inputs exist: `rfcs/inbox/.gitkeep`,
   `rfcs/outbox/.gitkeep`, and `rfcs/retro-log.md` with its column header.
2. Reconcile `triage-rfc:50-52` with reality: either recover
   rfc-0001..0003 or drop the reference (and with it the "older heading
   Disposition" note — see plan.md P15).
3. **Run one real dogfood end to end.** The findings in `plan.md` are
   already a retro of these skills; the cheapest first RFC is
   `plan.md` P1 (the model-vocabulary contradiction) written up on
   `skill-rfc-template.md` and triaged through `triage-rfc`. That
   simultaneously fixes a blocker, validates the template, validates the
   triage workflow, and produces the first `retro-log.md` line.
4. Whatever that dogfood run finds hard is a `retro-skill` finding about
   `retro-skill` — which is exactly the loop the repo says it wants
   (`retro-skill:207-212`).

---

## 3. There is no skill-authoring standard, so the skills drift from house style

**The observation.** This repo has a 737-line
`template-authoring-standards.md` governing every detail of how a
*document template* is written — line length, frontmatter, slug prefixes,
instruction-block placement, EARS, placeholders, reference syntax,
versioning, validation. It is thorough to the point of being the largest
file in the repo.

There is **no equivalent standard for skills** — and skills are the
actual product. The consequence is visible in the spread:

| Dimension | Best | Worst |
|---|---|---|
| 80-column compliance | `groom-feature-blueprint`: 0 lines over | `bead-bonanza`: 207 lines over, longest 1,070 |
| Progressive disclosure | `spec`: 17 bundled references | `bead-bonanza`: 0, body 615 lines |
| Reference-path convention | `groom-feature-blueprint`: consistent | `spec/references/`: two conventions two lines apart |
| Version scheme | `retro-spec`: `1.2.0` | everyone else: `1.2` / `1.0` / `0.1` |
| Model vocabulary | `spec`, `groom-beads`: one set | `bead-bonanza`: a different set |
| Changelog in prose | absent in 5 skills | `bead-bonanza:244` argues with its own past |

`README.md:10` already diagnoses this, about a different skill: *"Refine
the groom-beads skill, adopting the conventions (e.g. c4, references,
etc) of the surrounding skills. **Make it belong.**"* The instinct is
right; what is missing is a written definition of what belonging means,
so "make it belong" is a taste judgment per skill rather than a check.

**Proposal.** Write `docs/skill-authoring-standards.md` — deliberately
short, maybe 120 lines, the opposite of the template standard's
maximalism. It should settle, normatively:

- **Frontmatter**: required `name`, `description`, `version`; version
  scheme `MAJOR.MINOR` (pick one and fix `retro-spec`); description
  budget (~600 chars, hard ceiling 1,024) with the guidance that the
  description's job is triggers, not mechanism.
- **Naming**: one pattern for the collection (plan.md P16 recommends
  verb-object), and the rule that a skill name states the activity.
- **Body budget**: ≤400 lines with a stated split policy at 300 —
  tighter than the platform's 500 so there is headroom.
- **Bundling**: reference files live under the skill; the resolution root
  for backticked paths is the file that contains them; cross-skill
  sharing goes through the vendoring mechanism, not `../`.
- **Structure**: which sections a skill has and in what order —
  Core principle → When to run → Workflow (with a copyable checklist) →
  References → Anti-patterns → Retro this skill. Five of the seven skills
  already converge on roughly this; write it down and the sixth follows.
- **Prose rules**: 80 columns; no changelogs; no volatile counts; every
  prescribed command is executable as written.
- **Testing**: three evals minimum before a skill ships, and the model
  tiers it is tested against.

Then add a conformance section to `triage-rfc` step 5 so adopted changes
are checked against it, and wire the mechanical parts into
`check-consistency.sh` (§1).

**A note on tone.** The right move is *not* to make skills as
rule-governed as templates. Skills are prose for a reader that reasons;
the template standard's density would be counterproductive. The standard
should settle the handful of things that cause real divergence and stay
silent elsewhere.

---

## 4. Model tiering is coupled to model names — the root cause, not just the bug

**The observation.** `plan.md` P1 documents the contradiction: four files
say `opus | sonnet | fable`, `bead-bonanza` says
`haiku | sonnet | opus | fable`, and the two disagree about whether
`fable` is the cheapest tier or the most capable.

Stepping back, the contradiction is a *symptom*. The design decision that
caused it is that **the work order's routing field stores a model name.**
Consequences:

- Every model-lineup change is a coordinated edit across a JSON Schema
  enum, a Python constant, two prose lists, and one skill — five files,
  no gate (§1).
- A work order groomed today carries a name that may not exist when it is
  dispatched next quarter. The framework's whole premise is that a work
  order is *portable* and survives to a fresh agent with zero context; a
  hard-coded model name is the one field that does not travel.
- Two skills can hold coherent-but-incompatible tier orderings, because
  nothing writes the ordering down as an ordering.
- `work-order-field-guide.md:91-92` already senses the problem and
  half-solves it: *"If the project's model lineup differs, adopt its names
  but preserve the tiering (top / mid / cheap-fast) and the reasoning."*
  So the framework already knows the tier is the durable concept and the
  name is the local binding. It just stores the wrong one of the two.

**Proposal.** Invert it. Store the tier; resolve the name at dispatch.

1. Work orders carry `tier: cheap-fast | mid | top` (closed, ordered,
   model-agnostic) plus `tier_reason` — the current `model_reason`, which
   is the genuinely valuable field because it is what a retro audits.
2. The project declares its binding once, e.g. in `AGENTS.md` beside the
   gate profile:
   ```markdown
   ## model lineup
   - cheap-fast: haiku
   - mid: sonnet
   - top: opus
   ```
3. `bead-bonanza` resolves `tier` → model at dispatch, and its escalation
   ladder becomes `cheap-fast → mid → top` — which is unambiguous, cannot
   invert, and needs no edit when the lineup changes.
4. `bead.schema.json` and `validate_beads.py` validate the *tier* enum, a
   three-value set that will not churn. Keep `model` as an optional
   override for the rare bead that genuinely needs a named model.
5. Retain a compatibility read: a bead carrying only `model` maps to its
   tier via the project lineup, so existing backlogs keep working.

**Expected effect.** One five-file class of bug becomes structurally
impossible, work orders become portable across model generations, and the
routing rubric in `work-order-field-guide.md:60-92` — which is already
written in tier language — stops needing a translation layer.

---

## 5. The atomic unit of work has two names, and one skill uses both

**The observation.** `README.md:12` asks the question directly: *"Adopt
"work order" or "dev task" or "task" or "bead"?"* It is unresolved, and
the corpus shows it. Occurrences per SKILL.md:

| Skill | "work order" | "work item" | "bead" |
|---|---|---|---|
| `spec` | 11 | 0 | 3 |
| `groom-beads` | 9 | 0 | 66 |
| `groom-feature-blueprint` | 6 | **8** | 9 |
| `retro-spec` | 0 | **6** | 0 |
| `bead-bonanza` | 1 | 0 | 167 |

`groom-feature-blueprint` uses both, more often the non-canonical one —
in the skill whose entire job is *producing* the thing. `retro-spec` uses
only "work item". Meanwhile `spec/SKILL.md:79` promotes **Work Order** to
one of the ten first-class document types, and `spec:117-123` is careful
about what is and is not a promoted type.

**Why it is more than a style nit.** The framework already draws a real
distinction it could use here:

- **Work order** — the portable, tracker-agnostic *format* and its
  Definition of Ready. Owned by `spec`.
- **Bead** — one materialization of a work order in the `bd` tracker.
  Owned by `groom-beads`, which even publishes the mapping table
  (`groom-beads:63-80`).

That is a clean two-term vocabulary. "Work item" is a third word for the
first concept, and "task" appears as a fourth. The three-way slippage
undermines `spec:79`'s claim that Work Order is a *type*, and it is
exactly the consistency problem the best-practices doc flags — but the
deeper cost is conceptual: a reader cannot tell whether "work item" means
the portable order or its bead.

**Proposal.**

1. Ratify: **work order** = the portable unit and format; **bead** = its
   beads materialization; retire "work item" and stop using "task"
   except as the `kind: task` enum value.
2. Sweep `groom-feature-blueprint` (8 sites) and `retro-spec` (6) — pure
   find-and-replace, no semantic change.
3. Record the decision. This is genuinely an ADR: it has real
   alternatives, and `docs/adrs/` is where the repo puts choices like
   this. `ADR-013-work-order-terminology`.
4. Add a terminology check to §1's script: `work item` count == 0.

---

## 6. The framework has no on-ramp

**The observation.** `README.md` is not a README. It is a private,
unstructured TODO list — CAB persona notes, open questions about Gherkin,
C4 conventions still being drafted, "Config" ideas. A newcomer (human or
agent) arriving at this repo learns nothing about what the framework is,
how to install it, or which skill to invoke when.

For a 7-skill framework with a *directional* pipeline
(`spec` → `groom-feature-blueprint` → `groom-beads` → `bead-bonanza` →
`retro-spec`/`retro-skill` → `triage-rfc`), the ordering is the hardest
thing to learn and the least documented. Each skill describes its own
handoffs well — `groom-feature-blueprint:15-30` is excellent about its
neighbors — but no artifact shows the whole shape. An agent must read
seven skills to find out that the second one exists.

**Proposal.** Three artifacts, in value order:

1. **A real `README.md`**: what the framework is (three sentences), the
   pipeline diagram, install (`scripts/install-spec-skill.sh` or the
   marketplace), and a "which skill do I want?" table keyed by what the
   user is trying to do. Move the current contents to
   `docs/backlog.md` — or better, file them as beads, which is what this
   repo's own tooling is for.
2. **One worked example threaded end to end**, under `docs/example/`: a
   small feature, its FRD, its blueprint, the groomed work orders, and the
   retro that came out. This is the single most valuable missing artifact.
   The framework asks a project to adopt ten document types, EARS, C4,
   wikilinks, delta keys, AC-coverage tables with proving deltas, and
   verified-by levels — and shows no complete instance of the result. The
   best-practices doc's own advice applies: *"Examples convey the desired
   style and level of detail to Claude more clearly than descriptions
   alone."* It doubles as eval fixtures (plan.md P10).
3. **A pipeline diagram** — a small mermaid graph of skills, the artifact
   each consumes and produces, and the human gates. Put it in the README
   and reference it from each skill's role paragraph rather than
   re-describing the neighborhood in six places.

---

## 7. The frontmatter index is specified but does not exist — and three checks depend on it

**The observation.** `template-authoring-standards.md:319-325` specifies
it:

> "A document system SHOULD maintain a machine-readable index of all
> documents' frontmatter (a JSON object written to disk). The index is the
> authority for slug-uniqueness checks, next-number derivation for
> numbered types, `[[slug]]` reference validation, and dangling-reference
> detection after a promotion or reparent."

Four capabilities, one artifact, and it does not exist. ADR-012 concedes
the cost in its own Consequences (`:96-97`): *"Reference validation
depends on the planned frontmatter-index script existing; **until it
does, `[[ ]]` links are unchecked.**"*

Downstream, things assert what it would have proved:

- `blueprint-readiness-checklist.md:54` check #5 — *"Every `[[slug]]`
  reference resolves to an existing document"* — is a gate
  `groom-feature-blueprint` Phase 1 must *enforce*, by hand, across a
  whole blueprint.
- `template-authoring-standards.md:237-241` claims slug uniqueness "is
  enforced mechanically against the frontmatter index… so a collision is
  a commit-time error, not a silent overwrite." Nothing enforces it.
- `:274-276` — the next ADR number "is derived from the frontmatter index
  (`max + 1`), never tracked by hand." It is tracked by hand.

**Why this is high leverage.** It is one small script — walk the tree,
parse YAML frontmatter, emit JSON — that converts three hand-checks into
machine checks and unblocks the reference validation ADR-012 was designed
around. Perhaps 60–80 lines of Python. The ratio of specified-value to
implementation-cost is the best in the repo.

**Proposal.** `scripts/build-frontmatter-index.py` → `.spec-index.json`,
plus `scripts/check-references.py` that reads it and reports dangling
`[[slug]]`s, duplicate slugs, and slugs whose prefix does not match their
`type` per the registry at `template-authoring-standards.md:296-307`.
Wire both into §1's consistency script. Then change
`blueprint-readiness-checklist.md` check #5 from a prose obligation to
*"run `scripts/check-references.py <blueprint>`; it must report zero
dangling references"* — a machine gate where there is currently a
careful-reading gate.

---

## 8. Durable specs have no drift gate, so they rot silently

**The observation.** The framework is careful about which documents are
transient and which are durable —
`spec/SKILL.md:152-158`:

> "Architecture blueprints and contracts describe the system **as it is**
> and are updated in the same change that alters the architecture."

And it builds real machinery to keep that promise at *merge* time:
`groom-feature-blueprint:100-107` puts the fold-into-the-owning-blueprint
work inside the same work order as the code, and `spec:284` makes
"Merging a feature without folding its deltas into the owning blueprints"
an anti-pattern.

But every gate in the framework fires **before** implementation:

| Gate | Fires |
|---|---|
| Blueprint readiness checklist | before grooming |
| Work-order Definition of Ready | before dispatch |
| `bd-health.sh` / `validate_beads.py` | before dispatch |
| `retro-spec` / `retro-skill` | after, on *guidance* quality |

Nothing ever asks: *is `container-ledger-api`'s blueprint still true of
`services/ledger`?* Architecture blueprints are the framework's only
permanently-durable specs, they are the ones most expensive to re-derive,
and they are the only ones with no verification at all. A single feature
merged without its fold — the anti-pattern the framework names — is
invisible forever.

The container and component templates even carry the raw material:
`component-blueprint-template.md:9-10` declares `source_paths: []` and
`test_paths: []`. Those fields exist to make drift checkable. Nothing
checks them.

`README.md:64-66` proposes a `spec-review` skill. I think that is the
right instinct pointed at the wrong moment — reviewing a PR for
compliance catches new violations; it does not find the blueprint that
quietly stopped being true three features ago.

**Proposal.** Two complementary pieces:

1. **`scripts/check-spec-drift.py`** (cheap, mechanical, high signal):
   for every blueprint, assert `source_paths` and `test_paths` exist;
   report blueprints whose `source_paths` have git commits since the
   blueprint's `updated_at`; report source roots that no blueprint
   claims. None of this proves a blueprint is *correct*, but the last
   check finds undocumented containers and the middle one produces a
   ranked staleness list — which is exactly the triage input a human
   needs.
2. **A `spec-review` skill** as the README proposes, but scoped to a
   *diff*: given a PR, does it touch a `source_paths` root whose
   blueprint it did not update? Is every new normative statement in EARS?
   Do new ACs carry citations per `ac-traceability-standards.md`? This is
   the "verification gate" pattern the framework already believes in
   (`ADR-011-targeted-verification-gates`), applied to the one phase that
   lacks one.

Sequence: the script first. It is small, it produces a worklist, and the
worklist tells you whether the skill is worth writing.

---

## 9. `retro-skill` is the least mature skill and the most load-bearing

**The observation.** Version inventory:

| Skill | Version |
|---|---|
| `spec` | 1.3 |
| `groom-feature-blueprint` | 1.2 |
| `bead-bonanza` | 1.2 |
| `retro-spec` | 1.2.0 |
| `groom-beads` | 1.0 |
| `triage-rfc` | 1.0 |
| **`retro-skill`** | **0.1** |

`retro-skill` is the *only* mechanism by which this framework learns
anything. Every other skill's "Retro this skill" footer routes to it —
`spec:288-294`, `groom-beads:302-312`, `retro-spec:175-184`,
`bead-bonanza:608-615`, `groom-feature-blueprint:314-318`. It is the
entry point to the entire improvement pipeline, it is at 0.1, and
(per §2) it has never produced an artifact.

Its content is good — the three finding shapes at `:56-79` are a genuinely
useful taxonomy, and the insistence on model/effort provenance at
`:96-110` is more sophisticated than most retro tooling. The maturity gap
is not quality; it is exercise.

**Proposal.** Treat `retro-skill` as the framework's highest-priority
skill rather than its newest, and specifically:

1. Run it. §2's dogfood is the natural first exercise, and `plan.md` is
   pre-harvested evidence.
2. Give it the worked examples its siblings have. `:71-79`'s three shapes
   have illustrative fragments, not input/output pairs; one complete
   finding per shape would make it usable by a smaller model (plan.md
   P11).
3. Resolve the small internal staleness: `:26-29` says its scope is
   *"A plugin skill in this repo (`spec`, `groom-beads` and `retro-skill`
   itself)"* — a partial enumeration of a 7-skill repo that will keep
   going stale. Say "any skill in this repo" and drop the list.
4. Reconcile the routing at `:170-180` with §2's directory skeleton.

---

## 10. The `spec` skill is a router that costs 294 lines to route

**The observation.** `spec/SKILL.md`'s job is mostly *classification*:
given a piece of work, pick one of ten document types and load the right
template. That decision occupies `:45-183` — a 10-row type table, a
9-question decision tree, 12 boundary cases, and an ordering section —
and every token of it loads before the agent knows which single template
it needs. The References table at `:233-256` (the actual routing table)
is the last thing in the file.

The 22-row Anti-Patterns table at `:262-286` is excellent content, but
most rows are type-specific: an agent authoring an ADR loads seventeen
anti-patterns about container blueprints, contracts, and feature
blueprints.

This is the one place the repo does not apply its own progressive
disclosure. Ironically `spec` is the skill that bundles the most
references (17) — it disclosed the *templates* well and kept the
*routing* monolithic.

**Proposal.** Restructure `spec/SKILL.md` as a thin router, ~120 lines:

- Core principle and why spec-first (`:20-41`) — keep, it is short and
  load-bearing.
- The type table — keep; it is the routing key.
- **Move** the decision tree + boundary cases (`:86-163`) to
  `references/choosing-a-document-type.md`, loaded only when the type is
  genuinely unclear. Keep 3–4 lines in the body for the common cases plus
  "if the choice is not obvious, load it".
- **Move** the anti-patterns table to `references/anti-patterns.md`, or
  better: distribute each row into the template it concerns, where it
  fires at the moment of relevance. (Careful: this must not become
  duplication — each row lands in exactly one place.)
- Keep the References table and the authoring workflow, and promote the
  References table nearer the top.

Net: the common path ("I know I need an FRD") loads ~120 lines instead of
294, and the hard path loads exactly the disambiguation help it needs.
This is the same conditional-details move `plan.md` P6 proposes for
`bead-bonanza`, applied to the other end of the pipeline.

---

## 11. No adoption ladder: the floor is as high as the ceiling

**The observation.** Adopting this framework as written means adopting:
ten document types; six status values with a state-vs-transient
distinction; EARS with RFC 2119 modals; C4 with four documented
extensions; `[[slug]]` wikilinks plus heading anchors plus backticked
paths; a slug prefix registry; delta keys as stable identity; AC coverage
tables with proving-delta marking; verified-by levels; test citation
comments; Tracker Reconciliation; a Definition of Ready; and a retro
protocol. That is ~9,300 lines of framework.

For a team already committed to spec-driven development at scale, this is
a strength — the rigor is the product. But there is no smaller version.
A project that wants only "make my backlog dispatchable to agents" — which
is the framework's most immediately compelling promise, and the one
`groom-beads` alone could deliver — has no documented path that does not
route through blueprints, C4, and delta keys.

The skills are individually more modular than the framework advertises.
`groom-beads` degrades gracefully when `bd` lacks flags (`:96-99`) and
skips C4 cleanly when there are no C4 docs (`:207-208`).
`groom-feature-blueprint` supports a no-tracker fallback (`:233-241`).
The graceful degradation is *built* and simply not surfaced as an adoption
story.

**Proposal.** Document a three-rung ladder in the README (§6), and be
explicit that each rung is a legitimate resting place, not a failure to
finish:

| Rung | Adopt | Skills | Get |
|---|---|---|---|
| 1 | Work orders + Definition of Ready | `groom-beads`, `bead-bonanza` | a backlog you can dispatch to isolated agents |
| 2 | + FRDs in EARS, + retros | `spec` (FRD only), `retro-spec` | testable requirements and a learning loop |
| 3 | + architecture blueprints, C4, feature blueprints, contracts | all | full traceability from requirement to test citation |

Then audit each skill for a rung-1 path that does not mention rung-3
concepts. `groom-beads` is nearly there already; its dependence on
`../spec/references/*` (plan.md P9) is the main coupling, which §4 of
`plan.md` addresses anyway.

**A caution.** Do not weaken any gate to make a lower rung possible.
`retro-spec:172` names the anti-pattern precisely — *"Weakening a
consumer's checks to make authoring easier."* The ladder is about which
document *types* a project adopts, never about which checks apply to the
types it did adopt.

---

## 12. Version compatibility is unmapped

**The observation.** Four independent version axes:

- plugin: `0.7.0` (`plugin/.claude-plugin/plugin.json`)
- skills: 7 values from `0.1` to `1.3`, one of them three-part
- templates: `Version:` in each instruction block (adr 1.2, frd 1.1,
  standards 1.1, rfc 1.1, …)
- standards docs: frontmatter `version` (template-authoring 0.11,
  blueprint-readiness 0.5.0, findings-capture 0.2, ac-traceability
  0.2.0, DoR 1.1, field guide 1.1)

Note the format inconsistency even within one axis: `0.11` vs `0.5.0` vs
`1.1`. `template-authoring-standards:439` permits both ("`MAJOR.MINOR`
with an optional `.PATCH`"), so this is legal — but it makes comparison
awkward for the one consumer that must compare them.

That consumer is `triage-rfc:39-47`, whose entire provenance check is:

> "For every proposal, compare the cited skill/template/standard versions
> against the current ones."

There is no artifact mapping plugin `0.7.0` → the skill and template
versions it contains. So a downstream project reporting "I was on plugin
0.7.0" gives triage nothing directly usable; triage must resolve four
version axes by hand. `triage-rfc:102-108` bumps the plugin MINOR on any
normative change, which is right, but the bump records *that* something
changed, not *what*.

**Proposal.**

1. Generate `VERSIONS.md` (or `.spec-versions.json`) at release: plugin
   version → every skill, template, and standards version it ships. One
   script, appended to `triage-rfc` step 7.
2. Standardize on `MAJOR.MINOR` for skills and normalize `retro-spec`'s
   `1.2.0`; leave standards docs alone if `.PATCH` is genuinely used
   there.
3. Simplify `triage-rfc`'s provenance check to: given the RFC's
   `plugin_version`, look up that row and diff. This turns a four-axis
   manual reconciliation into a table lookup — and it is what
   `rfc-template.md`'s `plugin_version` frontmatter field was already
   designed for.

---

## 13. This repo does not dogfood `bead-bonanza`

**The observation.** `bead-bonanza` Phase 0 step 1 requires a
`## bead-bonanza gates` block in `AGENTS.md`, and describes a slow path
when it is absent: discover candidate commands, propose a draft, prompt
the user, write it back.

This repo's `AGENTS.md` has no such block (verified: zero matches). So a
`bead-bonanza` run *here* — the most likely first run, since this is
where the skill is developed — takes the slow path and prompts, on the
one repo whose maintainer wrote the skill.

The gap is trivial to close and the answer is interesting: this is a
docs-only repo, so `<test>` is a no-op — the case
`bead-bonanza:48` explicitly handles (*"If the project genuinely has no
build/test step (e.g. a pure-docs repo), record that: `<test>` is a
no-op and the doc/chore fast lane is the only dispatch mode"*). Writing
that block would exercise the skill's least-tested branch.

**Proposal.** Add to `AGENTS.md`:

```markdown
## bead-bonanza gates

- **Format:** *(none — markdown only)*
- **Lint:** `scripts/check-consistency.sh`   # see plan2 §1
- **Build:** *(none)*
- **Test (full suite):** *(no-op — documentation repo)*

### Targeted tests by surface
- `plugin/skills/*/SKILL.md` → `scripts/check-consistency.sh`
- `plugin/skills/groom-beads/scripts/` → `python3 .../validate_beads.py evals/fixtures/*.jsonl`

### Environmental-failure ledger
- (none recorded yet)
```

This closes the loop nicely: §1's consistency script becomes this repo's
lint gate, which means `bead-bonanza` runs here would enforce the
invariants from §1 automatically.

---

## 14. Smaller items

**a. The marketplace description is stale.**
`.claude-plugin/marketplace.json` advertises *"(spec,
groom-feature-blueprint, retro-spec)"* — three of the six plugin skills.
`groom-beads`, `retro-skill`, and `bead-bonanza` are invisible to anyone
browsing the marketplace. The `plugins[0].description` is current; only
the top-level `metadata.description` is stale. One-line fix, and worth
adding to §1's checks since it will go stale again.

**b. `retro-skill`'s self-scope enumeration.** `:26-29` lists three
skills as its scope. Generalize (see §9.3).

**c. `templates/` has one file, and it is a stale duplicate.** Covered as
`plan.md` P5; the structural question is whether `templates/` should exist
at all. It should not — every template in this framework belongs to a
skill.

**d. `docs/adrs/` is exemplary and under-referenced.** Twelve ADRs
recording real decisions with alternatives tables — this is the framework
practising what it preaches, and it is the most convincing artifact in
the repo for a skeptical adopter. But no skill references them, and
`docs/adrs/README.md` is 24 lines. Consider linking the ADR index from
the main README (§6) as evidence, and citing the relevant ADR from the
convention it justifies — e.g. `template-authoring-standards`'s reference
section should cite `[[adr-012]]`, which currently cites *it* one-way.

**e. `AGENTS.md` has a duplicated Quick Reference.** Lines 16-24
(hand-written) and 55-62 (inside the managed `<!-- BEGIN BEADS
INTEGRATION -->` block) are near-identical. The managed block will be
regenerated by tooling; the hand-written copy will drift. Delete the
hand-written one — the same drift argument as `plan.md` P5, one level up.

**f. `retro-skill:27` has trailing whitespace mid-sentence** — a
cosmetic artifact of reflowing. Worth a pass if §1 adds a whitespace
check.

---

## What I would do first

If only three things happen from this document:

1. **§1 — the consistency script.** It is the framework applying its own
   thesis to itself, and it is what prevents every `plan.md` Tier 1 and
   Tier 5 finding from silently recurring. Half a day.
2. **§2 — run the retro loop once, end to end**, using `plan.md` P1 as
   the first RFC. It validates the framework's most original machinery,
   fixes a blocker, and produces the first `retro-log.md` entry. A day.
3. **§6 — the worked end-to-end example.** It is the missing artifact that
   makes the framework legible to a newcomer, and it doubles as eval
   fixtures for `plan.md` P10. Two days, and it will surface real defects
   in the templates as you fill them in.

§4 (tier-not-model) and §7 (frontmatter index) are the two highest-value
*structural* changes, and both are small. I would take them next.
