---
title: "RFC: The reference-integrity blueprint states a verification
  condition as a landing order, and asks for a check it registers
  nothing to prove"
description: Split the one Sequencing Constraint that conflates a hard
  ordering with a verification-timing condition, and name the mechanism
  by which the pairing check is proved in a profile that registers no
  pairing
slug: rfc-0014
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: groom-feature-blueprint
skill_version: "1.2"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.15.0"
---

# RFC: Timing and fixtures in the reference-integrity blueprint

## Summary

[[feature-spec-reference-integrity]] was groomed into epic
`agent-plugins-9kb` and ten work orders. The Phase 1 readiness gate
passed all ten checks, and the document held up unusually well — all 24
acceptance criteria traced, all fourteen Tracker Reconciliation
classifications matched the tracker, and every lockstep rule spanning
delta blocks was already reconciled. Two things still forced the
groomer to decide rather than read: one Sequencing Constraint whose two
readings produce materially different dependency graphs, and one delta
that requires proving a check the shipped profile registers nothing to
exercise. P1 and P2 are narrow edits to the blueprint that remove those
judgment calls from the next agent to consume it. P3 comes from filing
this RFC rather than from grooming: the repository's `Rfc` schema can
only describe an RFC that critiques a skill, so every RFC it holds
about a *document* misuses one property.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Phase executed | Groomed [[feature-spec-reference-integrity]] into beads (epic `agent-plugins-9kb`, ten work orders) |
| Documents exercised | [[feature-spec-reference-integrity]]; [[frd-spec-reference-integrity]]; [[module-document-core]], [[module-spec-profile]], [[container-spec-cli]]; [[contract-specs-configuration]]; [[adr-012]], [[adr-015]], [[adr-016]], [[adr-018]]; [[testing-standards]], [[document-relationship-standards]]; `schemas/frontmatter.strata` |
| Documents produced | Epic `agent-plugins-9kb` and work orders `agent-plugins-9kb.1`–`.10`; `rfcs/findings-capture.md` |
| Skill(s) followed | groom-feature-blueprint v1.2; groom-beads v1.1 |
| Templates / standards read | feature-blueprint-template v1.4; blueprint-readiness-checklist v0.6.0; work-order-template v1.3; work-order-definition-of-ready v1.2; ac-traceability-standards v0.2.0 |
| Plugin version | 0.15.0 |
| Session date | 2026-08-07 |
| Model | Withheld: this session's harness supplies the model identifier under a policy that forbids writing it into any artifact pushed to a repository. Same handling as RFC-0010; the open question is tracked by `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`. |

## Evidence

1. Deriving the dependency graph, the groomer stopped at this
   Sequencing Constraint
   (`specs/features/reference-integrity/reference-integrity-blueprint.md`
   lines 592–594):

   > - The PREG delta SHALL land before LIFE, whose every check is
   >   inert without a lifecycle registration, and before REX's
   >   frontmatter extraction can be exercised against this
   >   repository's corpus.

   One bullet, one `SHALL land before`, two different obligations. The
   first is a hard ordering (PREG before LIFE). The second is a
   condition on *verification* — PREG must land before REX's
   frontmatter extraction can be exercised against the real corpus —
   which is not a statement about when REX may land. Read as a hard
   edge it yields `MODEL → XPT → PREG → REX → RES → CHK → CMD`, a
   fully serial chain. Read as a verification condition it yields
   `MODEL → {XPT, REX}` with `XPT → PREG` and `REX → RES` running in
   parallel, and the timing carried as a note on REX. The groomer took
   the second reading, wired no `PREG → REX` edge, put a
   verification-timing note on `agent-plugins-9kb.3`, and flagged the
   call to the human as a one-edge fix if the other reading was meant.
   The template's own instruction is what makes the ambiguity costly:
   `feature-blueprint-template.md` §Sequencing Constraints states that
   "only constraints listed here are guaranteed to become dependency
   edges in the derived plan" and asks for constraints "In EARS, one
   per line".

2. Writing the LIFE work order (`agent-plugins-9kb.7`), the groomer
   had to supply a test mechanism the blueprint does not name. The AC
   Coverage table declares AC-SPECREF-002.5 — the incomplete-pairing
   check — as proved by LIFE and verified at **unit** level (same
   file, line 642). The LIFE delta's anchors state the fixture need
   (lines 417–420):

   > **Anchors:** `cli/core/corpus/check.go`;
   > `cli/core/finding/finding.go`; snapshot cases under
   > `cli/testdata/cases/`, which need fixture corpora carrying a
   > retired document and a registered pairing — neither of which this
   > repository's own corpus supplies.

   and the PREG delta forbids the obvious way to satisfy it (line
   298):

   > - The profile SHALL register **no** pairing.

   Nothing in the blueprint says how a test registers a pairing when
   the shipped profile registers none. The groomer wrote the
   resolution into the work order's Design — register a test profile
   through the same `Registry` seam the command surface uses, rather
   than adding a pairing to the shipped profile — deriving it from
   [[container-spec-cli]]'s description of `Registry` as "the profile
   registry [[contract-specs-configuration]] names as the
   customization seam". An implementer resolving it the other way
   would break AC-SPECREF-004.2 and contradict PREG.

3. The blueprint's Definition of Done requires `version` bumped 0.2 →
   0.3 on [[module-document-core]] (line 686) and on
   [[module-spec-profile]] (line 692), one checkbox each. Grooming
   distributes the folds: six of the ten work orders fold some part of
   their delta into [[module-document-core]], because
   `groom-feature-blueprint` requires each item to carry "the fold of
   its delta into the owning architecture blueprint". No document says
   which item owns the single bump. The groomer wrote "Set `version`
   to 0.3 if not already set" into every folding item to make the act
   idempotent. This finding is recorded here for completeness but is
   routed to the skill lens — see Considered and Not Proposed.

4. Two Definition-of-Done entries were already satisfied when grooming
   ran. The blueprint requires "[[adr-018]] status: draft → approved"
   (line 702) and "`specs/adrs/README.md` carries an index entry for
   [[adr-018]]" (line 704); ADR-018 was already `status: approved` and
   already indexed at line 24 of that README. The groomer declined to
   create a near-empty work order and folded the verification into the
   final Definition-of-Done item instead. Routed to the skill lens —
   see Considered and Not Proposed.

5. Filing this RFC, its frontmatter had to be written twice. The
   `retro-spec` skill's `references/rfc-template.md` prescribes an
   `audience: project | skill-authors` property and, for
   `audience: project`, a frontmatter block of `title`, `description`,
   `slug`, `type`, `status`, `updated_at`, `audience`, `origin`,
   `plugin_version`. This repository's own `Rfc` struct
   (`schemas/frontmatter.strata` lines 408–453) declares no `audience`
   and makes six skill-critique properties required:

   > `origin: String` … `skill: String` … `skill_version: String` …
   > `model: String` … `effort: String` … `runs_observed: UInt >= 1`
   > … `severity: RfcSeverity`

   with `skill` documented as "The skill this RFC critiques." This RFC
   critiques a feature blueprint, not a skill. The first draft followed
   the template and did not conform; it was rewritten to match the
   schema, setting `skill: groom-feature-blueprint` — the skill whose
   run produced the evidence, not a skill under critique — by analogy
   with RFC-0013, which sets `skill: spec` while critiquing an FRD.
   RFC-0010 and RFC-0011 carry the same shape. Every project-audience
   RFC this repository has filed therefore misuses `skill`.

## Proposals

### P1 — Split the PREG constraint into an ordering and a timing statement

**Problem.** The bullet at lines 592–594 joins a hard ordering and a
verification-timing condition under one `SHALL land before`, in the one
section whose contents the planner turns into dependency edges. Its two
readings produce different graphs: one serializes five deltas, the
other parallelizes two pairs. `feature-blueprint-template.md`
§Sequencing Constraints already warns what is at stake — "only
constraints listed here are guaranteed to become dependency edges in
the derived plan" — but gives the author no separate place to record a
condition that is about verification rather than landing, so the
timing statement had nowhere else to go.

**Change.** Replace the single bullet with two statements, the second
marked as timing rather than ordering:

> - The PREG delta SHALL land before the LIFE delta, whose every check
>   is inert without a lifecycle registration.
> - **Verification timing, not an ordering constraint:** REX's
>   frontmatter extraction cannot be exercised against this
>   repository's own corpus until PREG registers this framework's
>   reference-bearing properties. REX SHALL be verified in isolation
>   against a profile registered by its own tests, and the profile
>   self-application gate is where the real corpus exercises it. This
>   implies no dependency edge between PREG and REX.

**Evidence.** Finding 1. The next agent to groom or re-groom this
blueprint reads one ordering and one timing statement instead of
choosing between two readings of one sentence, and the graph it derives
matches the one already built.

**Backward compatibility.** The blueprint is `status: in-progress` with
`plan: agent-plugins-9kb`. This proposal states the graph that was
actually built, so no work order changes and no edge is added or
removed. Adopting it makes the existing plan traceable to the document;
declining it leaves the plan resting on a groomer's reading.

### P2 — Name how the pairing check is proved without registering a pairing

**Problem.** AC-SPECREF-002.5 is declared proved by LIFE at unit level,
LIFE's anchors say the fixtures need "a registered pairing — neither of
which this repository's own corpus supplies", and PREG says "The
profile SHALL register **no** pairing." The three statements are
individually correct and jointly leave the implementer to invent the
bridge. No text in the blueprint names a mechanism. The wrong bridge —
adding a pairing to the shipped profile so the test has one — is the
cheapest one to reach and breaks AC-SPECREF-004.2.

**Change.** Add to the LIFE delta, after the fixture sentence in its
Anchors paragraph:

> The pairing and typed-reference fixtures SHALL be supplied by a
> profile registered for the test through the `Registry` seam
> [[container-spec-cli]] describes, never by relaxing PREG's
> no-pairing registration. That the shipped profile registers no
> pairing is an asserted property (PREG), so a test that changed it
> would invalidate AC-SPECREF-004.2 to prove AC-SPECREF-002.5.

**Evidence.** Finding 2. The implementer of `agent-plugins-9kb.7` reads
the mechanism in the blueprint rather than inheriting it from a work
order whose author derived it, and the conflict between PREG's
assertion and LIFE's fixture need is resolved in the document that
creates it.

**Backward compatibility.** Additive prose in one delta's anchors. No
acceptance criterion, delta key, or AC Coverage row changes; the work
order already carries the same instruction, so no re-grooming is
implied.

### P3 — Let the `Rfc` schema express an RFC that critiques a document

**Problem.** `schemas/frontmatter.strata` models every RFC as a skill
critique. `skill` is required and documented as "The skill this RFC
critiques"; `skill_version` is "The version recorded in the critiqued
skill's own frontmatter at the time of the run". An RFC produced by
`retro-spec` critiques *documents* — this one critiques a feature
blueprint, RFC-0013 critiques an FRD, RFC-0010 critiques a corpus
validator's specifications. There is no property that says so, and no
`audience`, which `retro-spec`'s `rfc-template.md` v1.2 makes the field
that decides who resolves the RFC. The result is that all four
document-critiquing RFCs in `rfcs/inbox/` set `skill` to whichever
skill happened to run, which is information about provenance dressed
as information about subject.

**Change.** On the `Rfc` struct:

> - Add `audience: RfcAudience`, a new enum of `project` and
>   `skill-authors`, matching `rfc-template.md`'s property of the same
>   name.
> - Make `skill` and `skill_version` optional (`String?`), REQUIRED
>   WHERE `audience` is `skill-authors` and otherwise recording the
>   skill whose run produced the evidence.
> - Add `documents_critiqued: [Slug]?`, the documents an
>   `audience: project` RFC argues about — which
>   [[frd-spec-reference-integrity]]'s reference checks then resolve,
>   making "which RFCs concern this blueprint?" a `spec refs` query.

**Evidence.** Finding 5. Triage reads what an RFC is about from a
property rather than inferring it from a directory name, and the
`skill` field stops carrying two meanings.

**Backward compatibility.** `audience` is new and required, so the four
RFCs in `rfcs/inbox/` and the ten in `rfcs/retired/` each need one line
added; all are this repository's own documents and none is referenced
by a test. Relaxing `skill` to optional invalidates nothing already
written. `documents_critiqued` is optional and may be left absent until
the reference-integrity feature makes it queryable. Note that
`rfcs/` sits outside the `specs/` corpus root, so no check enforces the
`Rfc` struct today and nothing breaks in CI either way — which is also
why the mismatch went unnoticed until now.

## Considered and Not Proposed

- **Who owns the `version` bump when folds are distributed (finding
  3).** Real and recurring, but the mismatch is between
  `feature-blueprint-template.md`'s Definition of Done — which models
  one checkbox per affected document, including its version bump — and
  `groom-feature-blueprint`'s rule that each work order carries the
  fold of its own delta. Any multi-delta feature folding into one
  blueprint hits it. That is machinery, not this blueprint's content,
  so it belongs to the skill lens and is being raised there rather
  than filed as a project proposal.
- **Definition-of-Done entries written as transitions (finding 4).**
  "[[adr-018]] status: draft → approved" reads as pending work when the
  transition has already happened, and `groom-feature-blueprint` says
  to create "One item per Definition of Done entry not tied to a single
  delta" — which would have produced a work order for finished work.
  The template models this shape itself (`- [ ] [[contract-<name>]]
  status: draft → active`), so the fix is to the template and the
  grooming rule, not to this blueprint. Routed to the skill lens.
- **Everything RFC-0013 already proposes.** That RFC retroed the
  authoring phase of the same FRD and is still `ready-for-review`: the
  unregistered lifecycle facts (its P1), the position semantics for a
  compiled-in registration (P2), and the stale `related` citations plus
  the `provider`/`consumers` omission (P3). This session confirmed all
  three from the consuming side — the blueprint's MODEL and XPT deltas
  do invent `live_statuses`, `retired_statuses` and
  `successor_property`, and PREG does record the `provider`/`consumers`
  divergence — but corroboration of an open proposal is not a new
  proposal. Noted here so triage can count this as a second
  observation.
- **Component reference form.** The blueprint writes
  `[[container-spec-cli#Command Surface]]` while the corpus also
  declares `slug: component-command-surface` in an element-record
  block, so two forms could address one component. Resolving check 6
  needed a moment's reasoning about whether an element-record `slug` is
  a promotion. It cost no wrong result and the gate passed cleanly;
  raising it would be taste rather than evidence. Worth remembering if
  a future corpus scan finds both forms in use.
- **The stale beads database.** Ten of the fourteen Tracker
  Reconciliation issues appeared not to exist until
  `.beads/issues.jsonl` was imported, because the SessionStart hook
  does not register in a multi-root workspace. `CLAUDE.md` documents
  this failure mode and its manual remedy in full, and the remedy
  worked. Environmental, already written down, no proposal.

## Non-Goals

- No change to the feature's scope, to any acceptance criterion, or to
  any delta key. P1 and P2 are clarifications of text that already
  governs work now in flight.
- No change to how RFCs are triaged, numbered, or filed. P3 changes
  what an RFC's frontmatter can say, not the pipeline that reads it.
- No change to the dependency graph already built. P1 documents the
  graph that exists rather than proposing a different one.
- No re-opening of anything RFC-0013 proposes; this RFC corroborates
  and does not restate it.

## Adoption

P1, P2 and P3 are independent of each other and of RFC-0013. P1 and P2
touch only
`specs/features/reference-integrity/reference-integrity-blueprint.md`;
P3 touches `schemas/frontmatter.strata` and adds one frontmatter line
to each existing RFC.

P3 additionally requires running `scripts/generate-schemas.sh` and
committing the regenerated artifacts, since it changes a Strata source.
It SHOULD NOT land in the same change as `agent-plugins-9kb.1`, which
regenerates the same outputs from `index.strata`, `finding.strata` and
`profile.strata` — two concurrent regenerations of
`cli/core/specmodel/` would conflict. Either order is fine; only
simultaneity is a problem.

P1 and P2 SHOULD land while the blueprint is `in-progress` and before the
work orders they concern are dispatched — `agent-plugins-9kb.3` (REX)
for P1 and `agent-plugins-9kb.7` (LIFE) for P2. Landing them later is
harmless but pointless: the work orders already carry both resolutions,
so the value is in making the blueprint the source of them rather than
a groomer's reasoning.

Backward compatibility is stated per proposal above; neither invalidates
an existing document, test, or citation, and neither requires
re-grooming.

Expected effect: a re-groom of this blueprint derives the same graph
without a judgment call, and the LIFE implementer cannot reach the
cheap wrong bridge for the pairing fixture.

## Resolution
