---
title: "RFC: Implementing anchor references — a renamed schema
  definition its derived consumer never heard about, an example
  illustrating a condition that cannot occur, and a shipped refusal no
  criterion asks for"
description: Name the derivation keyed on a schema definition's name,
  record which snapshot artifacts the bundled schema reaches, give the
  refusal and the unverified rule identifier criteria of their own,
  record why a criterion no corpus input can exercise is still load
  bearing, and settle whether the slug-less anchor form is legal
slug: rfc-0025-implementing-anchor-references
type: rfc
status: ready-for-review
updated_at: 2026-09-05T00:00:00Z
origin: agent-plugins
skill: retro-spec
skill_version: "1.5"
model: claude-opus-5
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.24.0"
---

# RFC: Implementing anchor references

## Summary

[[feature-spec-anchor-references]] was implemented as seven delta work
orders plus this Definition of Done item, under epic `agent-plugins-rc14`,
each by a separate agent holding no context but its own item. Every one
landed. None was bounced or blocked, and the finished corpus passes
`spec check --strict` with both new rules reporting `passed`.

The blueprint was accurate about the code it was going to change. Where
it cost time was at one remove: a delta renamed a schema definition and
did not name the code that derives its behaviour from that definition's
name, which broke the corpus checker outright; an FRD illustrated a
runtime condition with a case the implementation cannot reach; and two
behaviours that shipped — a refused flag combination and a severity
reported under a borrowed rule identifier — are covered by tests that
cite no criterion, because no criterion asks for them.

Six proposals follow. Five amend documents in this repository; one
settles a contradiction between [[adr-012]] and a shipped check that
predates this feature and that anchor resolution is what made visible.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Audience | project — resolved in this repository, which is what filing under `rfcs/inbox/feature-blueprint/` records; the generated `Rfc` schema has no `audience` property |
| Phase executed | Implemented a groomed plan — seven delta work orders plus this Definition of Done item, each in its own worktree |
| Documents exercised | [[feature-spec-anchor-references]]; [[frd-spec-anchor-references]]; [[frd-spec-markdown-links]]; [[frd-spec-structure]]; [[module-document-core]], [[module-spec-profile]], [[container-spec-cli]]; [[adr-012]], [[adr-016]]; `ac-traceability-standards.md` v0.4.0; `testing-standards.md` v0.1; `cli-output-standards.md` v0.1 |
| Documents produced | The folds into the three architecture blueprints; two content corrections to [[frd-spec-anchor-references]]; the retirement of [[feature-spec-anchor-references]] |
| Skill(s) followed | swarm v2.1 (orchestration), swarm-dispatch v0.4 (per-bead dispatch), retro-spec v1.5 (this document); groom-beads v1.3 and groom-feature-blueprint v1.3 produced the work orders in the prior phase |
| Templates / standards read | feature-blueprint-template v1.4; frd-template v1.3; `ac-traceability-standards.md` v0.4.0; `testing-standards.md` v0.1; `cli-output-standards.md` v0.1 |
| Plugin version | 0.24.0 |
| Session date | 2026-09-05 |
| Evidence channel | `retro:` comments recorded in flight on all seven delta items, harvested at this Definition of Done item per the findings capture protocol; 29 comments across `agent-plugins-rc14.1` through `.7` |

## Evidence

Findings are numbered for citation by the proposals. Each names the item
that recorded it; the two measured by this Definition of Done item are
marked as such.

1. **A delta renamed a schema definition and did not name the code that
   derives from that name (SREF, `agent-plugins-rc14.2`).** The SREF
   delta's Anchors line reads, in full: "`schemas/vocabulary.strata`
   (the `Slug` alias), `schemas/frontmatter.strata` (the
   reference-bearing properties), `scripts/generate-schemas.sh`."
   `cli/profile/registration.go` auto-registers every reference-bearing
   frontmatter property by matching the JSON schema `$ref`
   `"#/definitions/Slug"`. After the retype nothing matched, every
   document type's `ReferenceProperties` came back empty, and
   `corpus.ValidateRegistration` reported it as fatal — measured:
   `profile registration: "feature-blueprint.frd" is registered as a
   typed reference but not as a reference-bearing property`. The corpus
   checker was completely broken until the implementing agent found it.
   Nothing in the delta, the item, or [[module-spec-profile]] says that
   the definition's NAME is load bearing.

2. **An acceptance criterion's illustrative example named a condition
   the implementation cannot reach (measured by this item).**
   [[frd-spec-anchor-references]] illustrated AC-SPECANCH-002.4 with:
   "A target whose frontmatter failed to parse had no headings extracted
   from it, so a reference into it reports `unverified`". Both halves are
   false. `cli/core/corpus/corpus.go` sets `Headings: headingsOf(src,
   nil)` on the parse-error branch and `Headings: headingsOf(src, doc)`
   on the envelope-error branch, so a parse-error member does carry
   headings; and such a member decoded no slug, so a reference naming it
   fails the parent feature's slug rule and raises no anchor finding at
   all. Measured on a throwaway corpus: an entry with
   `parse_error=true` reported `headings=2`, and a reference into it
   came back `resolution={"state":"unresolved"}` with
   `anchor_resolution=null`. The condition that does reach `unverified`
   is a path the index carries no entry for — measured on the same
   corpus, `[[#Probe]]` written in a standalone `.feature` source gave
   `reference/unresolved-anchor unverified ... no headings are recorded
   for docs/probe.feature`. `agent-plugins-rc14.4` reached the same
   conclusion empirically, and recorded that it had to, because the FRD
   contradicted the code.

3. **A corpus-state figure in the FRD's Overview went stale during its
   own implementation (`agent-plugins-rc14.6`, re-measured by this
   item).** The Overview read "Nine references in the corpus name an
   anchor — four into `container-spec-cli#Command Surface`, five into
   component rosters of [[module-document-core]]". Measured on the
   finished tree from a freshly built index, filtered to references
   carrying an anchor: eleven of 697, every one
   `anchor_resolution.state == resolved`, six into
   `container-spec-cli#Command Surface` and five into
   [[module-document-core]] rosters. The count moved because blueprints
   written by this very feature added anchored references.

4. **Three agents hand-patched the same five snapshot fixtures, and the
   regeneration script already covers them (`agent-plugins-rc14.1`,
   `.2`, `.4`; measured by this item).** Every schema edit changes
   `opts.Profile.Schema()`'s bytes, which `cli/core/corpus/hash.go`
   folds into each index's `input_hash`, staling
   `cli/testdata/cases/{index-default,index-features,index-out-custom,index-out-stdout,index-unresolved-reference}/expected/manifest.json`.
   All three agents diffed a fresh run against the committed fixture and
   replaced the `input_hash` line by hand; the third recorded that the
   pattern was "three-for-three". Measured now:
   `scripts/regenerate-cli-snapshots.sh` writes `expected/manifest.json`
   for every case — its own header says so — and running it on the
   finished tree produced zero diff under `cli/testdata/cases`. The hand
   patching was avoidable, and nothing an implementing agent reads says
   so.

5. **A flag combination that names no set shipped a refusal no criterion
   asks for (`agent-plugins-rc14.7`).** The CMD delta states only "WHEN
   `refs` is given an argument qualified by an anchor, the CLI SHALL list
   every reference whose anchor resolves to that heading", and every
   AC-SPECANCH-003.x scenario is inbound. `--outbound` with an anchor is
   reachable from the command line and names no set: the references out
   of X into a heading of X. The implementer refused it with
   `errNoAnchorWithOutbound` in `cli/cmd/refs.go`, following
   `errNoEmitWithJSON`'s precedent, and covered it with the
   `refs-anchor-outbound-rejected` snapshot case — whose `case.toml`
   carries no `cites` key, because there is no criterion to cite.

6. **The `unverified` severity reports under a borrowed rule identifier,
   and no criterion says which (`agent-plugins-rc14.6`).**
   AC-SPECANCH-002.9 requires that the CLI "reports an anchor matching
   no heading and an ambiguous anchor under stable rule identifiers of
   their own" — two identifiers, for the two error cases. AC-SPECANCH-002.4
   requires the `unverified` report and names no identifier. The
   implementer raised it under `RuleUnresolvedAnchor`, making that
   constant the only one in `cli/core/finding` that reports at more than
   one severity, and rejected a third rule id as "the same check, no
   result" rather than a different defect. The decision is recorded in
   [[module-document-core]] and in the code; no criterion carries it, so
   a later reader is free to add the third rule.

7. **A criterion's two branches cannot be separated by any corpus input
   (`agent-plugins-rc14.5`).** `cli/core/references/anchor.go` resolves
   an anchor by `markdown.Identifier(heading.Text) == wanted` where
   `wanted := markdown.Identifier(*ref.Anchor)`. A resolved anchor's
   identifier is therefore by construction the heading's, so "derive the
   fragment from the heading" and "derive it from the anchor" emit the
   same bytes for every reference the matcher can produce. AC-SPECMD-001.3
   was proved with fixtures stating a heading whose text differs from the
   anchor — legitimate because `cli/core/mdlinks` consumes a settled
   resolution and never re-derives the match — and an end-to-end case
   through `spec links` was rejected as proving nothing. Nothing in
   [[frd-spec-markdown-links]] says the criterion is a guard against a
   future looser matcher rather than a description of observable
   behaviour.

8. **The slug-less anchor form [[adr-012]] mandates is reported as an
   error (measured by this item).** [[adr-012]] states: "A section of the
   referring document itself is referenced by an anchor with the slug
   omitted: `[[#Terminology]]`." AC-SPECANCH-002.6 resolves that form's
   anchor against the containing document, and does. The slug half is a
   separate check and fails every time. Measured on a throwaway corpus
   whose `adr-098` document contains `[[#Overview]]` above its own
   `## Overview` heading: `anchor_resolution={"state":"resolved",...}`
   and, from `spec check --strict`, `reference/unresolved error 14:19
   [[#Overview]] matches no document's slug`. This predates the anchor
   feature — an empty slug has never matched a declared one — but anchor
   resolution is what makes it visible, since the tool now reports the
   anchor resolved and the reference broken in the same run. No corpus
   member uses the form today, so no gate fails. Filed as
   `agent-plugins-clau`.

9. **[[module-document-core]] contradicts itself about Markdown
   Structure (`agent-plugins-rc14.5`).** Its relationships section says
   "The Corpus → Markdown Structure edge is still the intended shape
   rather than an existing one: nothing in Corpus calls
   `cli/core/markdown` yet", while its Components roster, twelve hundred
   lines above, says "Both edges from Corpus into Markdown Structure and
   Gherkin are exercised". `cli/core/corpus/corpus.go` calls
   `markdown.HeadingsFrom` and `cli/core/corpus/features.go` calls
   `markdown.Fenced`. Filed as `agent-plugins-bieu` during the run and
   deliberately left out of this feature's scope.

## Proposals

### P1 — Record that the reference-property derivation is keyed on a schema definition's name

**Problem.** Finding 1. [[module-spec-profile]] describes the derivation
accurately and says nothing about what breaks it: "`cli/profile/
registration.go` reads the reference properties and the status
vocabulary from the decoded schema at package initialization; only the
one typed reference (`frd` → `type: frd`) and the lifecycle's live/retired
split are authored, because the schema has the rest." A reader learns
that the registration is derived. They do not learn that it is derived by
matching one definition's name, so renaming that definition silently
empties the registration for every document type.

**Change.** Add one sentence to that paragraph, after the sentence
quoted above:

> The match is on the definition's name — `#/definitions/Citation` in
> the decoded schema — so renaming or replacing that definition in
> `schemas/vocabulary.strata` empties every type's registration until
> this file is retargeted with it, and `corpus.ValidateRegistration`
> reports the result as a fatal `config/invalid-registration` rather
> than as a missing reference.

**Backward compatibility.** Documentation only; no document or code
becomes invalid.

### P2 — State which snapshot artifacts the bundled schema's bytes reach, and how to refresh them

**Problem.** Finding 4. `specs/standards/testing-standards.md`'s
Snapshot Case Structure requires that "a script that regenerates every
expected artifact SHALL be provided" and does not say what invalidates
one. Three agents in one run therefore diagnosed the same failure
independently and repaired it by hand, three times, when
`scripts/regenerate-cli-snapshots.sh` already does it.

**Change.** Add to the Snapshot Case Structure section:

> A case's `expected/manifest.json` carries the index's `input_hash`,
> which `cli/core/corpus/hash.go` folds the bundled profile schema's
> bytes into. Any change under `schemas/` therefore stales every case
> that asserts a manifest, whether or not the change alters what those
> cases observe. `scripts/regenerate-cli-snapshots.sh` rewrites both
> `expected/checks.log` and `expected/manifest.json` for every case and
> is the way to refresh them; reviewing the resulting diff remains the
> author's obligation under the paragraph above.

**Backward compatibility.** Documentation only.

### P3 — Give the refused flag combination a criterion of its own

**Problem.** Finding 5. `spec refs --outbound '<slug>#<anchor>'` exits 1
with an error, is covered by a snapshot case, and is asked for by no
acceptance criterion — so the case's `case.toml` cites nothing, and the
behaviour is invisible to any coverage verification run against active
FRDs. [[ac-traceability-standards]] makes active FRDs plus test
citations the durable traceability pair; a behaviour outside that pair
can be deleted by a later change with no test naming what it was for.

**Change.** Append to FEAT-SPECANCH-003 in
[[frd-spec-anchor-references]], as the next unused identifier:

> ```gherkin
>   @AC-SPECANCH-003.5 @negative
>   Scenario: An outbound query takes no anchor
>     Given a query for the references OUT of a document
>     When the argument is qualified by an anchor
>     Then the CLI refuses the combination and exits non-zero
> ```

and cite it from `cli/testdata/cases/refs-anchor-outbound-rejected/
case.toml`'s `cites`. Add to the block's *Examples* paragraph: an anchor
names a heading of the cited document, and the outbound direction asks
what the given document cites, so the combination names no set.

**Backward compatibility.** Append-only identifier, per
[[ac-traceability-standards]]; no existing criterion is renumbered and
the shipped behaviour already satisfies it.

### P4 — Say which rule identifier carries the unverified anchor report

**Problem.** Finding 6. AC-SPECANCH-002.9 requires "stable rule
identifiers of their own" for the two error cases and is silent about
the unverified one; AC-SPECANCH-002.4 requires the unverified report and
is silent about its identifier. The implementation borrows
`reference/unresolved-anchor`, which is a defensible reading and not the
only one. A later agent adding `reference/unverified-anchor` would
satisfy every criterion as written while breaking every caller
selecting the anchor checks with `--only`.

**Change.** Add to AC-SPECANCH-002.4's Examples paragraph in
[[frd-spec-anchor-references]]:

> `reference/unresolved-anchor` carries two outcomes, not one.
>
> At severity `error` it means the anchor named no heading in the target
> document. At severity `unverified` it means the check could not answer:
> the target has no index entry, so there were no headings to match the
> anchor against.
>
> They share one identifier deliberately. Both are the same check
> reporting on the same anchor — one found the answer is no, the other
> found it could not ask. Giving the second a separate identifier such as
> `reference/unverified-anchor` would split them, and every caller that
> narrows to the anchor checks with `--only` would have to name both to
> keep the coverage one name gives it today.
>
> This is the only rule identifier in the corpus checks that reports at
> more than one severity.

**Backward compatibility.** States what ships; no test or document
changes meaning.

### P5 — Record why AC-SPECMD-001.3 is load bearing though no corpus input exercises it

**Problem.** Finding 7. The criterion reads as a description of
observable behaviour, and is not one: no reference today's matcher can
produce distinguishes its two fragment sources. An agent measuring that
would reasonably conclude the branch is dead and collapse
`anchorFragment` back onto the anchor as authored — which is exactly the
change the criterion exists to prevent, and which no test over corpus
input would catch. `cli/core/mdlinks/mdlinks.go` states this in a
comment; [[frd-spec-markdown-links]] does not.

**Change.** Add to AC-SPECMD-001.3's rationale in
[[frd-spec-markdown-links]]:

> The two sources cannot differ under the matching rule
> [[frd-spec-anchor-references]] fixes, which resolves an anchor by
> comparing its heading identifier with the heading's: a resolved
> anchor's identifier is the heading's. The criterion is therefore
> verified by fixture rather than by corpus input, and what it holds in
> place is the seam — a later matcher that admits a leading requirement
> code, or a heading named in part, would begin emitting fragments that
> reach nothing the moment the fragment is derived from the anchor
> instead.

**Backward compatibility.** Rationale only; the criterion is unchanged.

### P6 — Settle whether the slug-less anchor form is legal

**Problem.** Finding 8. [[adr-012]] mandates `[[#Terminology]]` for a
section of the referring document, and every use of it raises
`reference/unresolved` at severity `error`. A corpus that gates on
`spec check --strict` — this one does — cannot contain the form the ADR
requires. The two documents cannot both be followed.

**Change. Decided in review of this proposal: exempt the form.** A
reference whose anchor names a real heading of the document it sits in is
valid, and must not be reported. The work is tracked as
`agent-plugins-clau`.

- `reference/unresolved` skips a reference whose slug half is empty and
  whose anchor resolved against the containing document, and
  [[frd-spec-reference-integrity]] gains the criterion saying so.
- [[adr-012]] keeps the slug-less bullet; the form it mandates stays
  legal.
- AC-SPECANCH-002.6 stands. The anchor half already resolves such a
  reference deliberately, and reporting the slug half broken contradicts
  it — which is the contradiction this proposal exists to remove.

**Rejected: withdraw the form.** [[adr-012]] would drop the slug-less
bullet and name the fully qualified form instead, retiring
AC-SPECANCH-002.6 with it. Rejected because an anchor-only reference into
the enclosing document is a legitimate thing to write and the checker
already resolves it correctly; the defect is that a second rule then
calls the same reference broken.

**Backward compatibility.** Every existing document stays valid: no
corpus member uses the form today, measured on the finished tree — all
eleven anchored references carry a slug. The change can only turn a
reported error into silence, never the reverse.

## Considered and Not Proposed

- **A rule for how a delta's Anchors list is composed (finding 1).**
  The durable form of this is a constraint on the feature-blueprint
  TEMPLATE — "a delta that renames a declaration lists every consumer
  that names it" — which lands in the repository that vends the
  template, not in this project's documents. Routed to the skill lens;
  P1 takes only the half that lands here.
- **A rule against stating volatile corpus figures in an FRD (finding
  3).** `AGENTS.md` already carries it: "Where the fact is volatile — a
  file size, a warning count, a line number — record the command that
  measures it and no figure at all." The figure was corrected in place
  during this item. A second statement of an existing rule is not a
  proposal.
- **Exporting `references.AnchorTarget` as a documented seam.** The
  implementer of `agent-plugins-rc14.6` found that
  `finding.UnresolvedAnchor` needs a target path `specmodel.Reference`
  does not carry, and exported the private derivation rather than
  copying it into `cli/core/corpus/check.go`. That is the right outcome
  and it was reached without guidance; no document change would have
  improved it, and the code now carries the reason.
- **[[module-document-core]]'s self-contradiction about Markdown
  Structure (finding 9).** A content fix to one document, already filed
  as `agent-plugins-bieu`, and caused by the structure feature rather
  than by this one. Per the RFC template's own scope note, a wrong
  sentence is fixed, not RFC'd.
- **The tension between a bead's capability tier and its
  collision-chain assignment.** `agent-plugins-rc14.4` recorded that an
  `advanced` bead landed on a dispatch also holding two `standard`
  ones, because a shared collision mark serialized them, and asked
  whether tier is meant to gate assignment. That is a question about the
  planning skill, not about this project's documents. Routed to the
  skill lens.

## Non-Goals

- **Re-opening the matching rule.** Identifier equality is settled by
  AC-SPECANCH-002.2 and by [[adr-012]]'s choice of a notation that must
  survive Obsidian. P5 exists to keep a future change to it honest, not
  to invite one.
- **Nested heading paths, block anchors and terse requirement anchors.**
  [[frd-spec-anchor-references]] defers all three explicitly, and
  nothing in this run gave a reason to revisit them.
- **A coverage tool that reconciles citations against active FRDs.**
  [[ac-traceability-standards]] describes one as a SHOULD and none
  exists. P3 exists because the pair it depends on must be complete
  whether or not the tool is built; building it is separate work.

## Adoption

Each proposal is independently adoptable. P1, P2, P4 and P5 are
documentation changes with no code consequence and can land in any
order. P3 changes an active FRD and a snapshot case manifest together —
adding the criterion without the citation would leave it uncovered, which
is the state it exists to remove — so those two edits land in one change.

P6 is the only proposal that asks for a decision rather than an edit,
and it must be resolved before the form it concerns is used anywhere in
the corpus. Nothing in the corpus uses it today, so the decision is not
urgent; it is tracked as `agent-plugins-clau`.

## Resolution

Resolved by the repository owner in review of PR #38, 2026-09-05.

| | disposition | tracked as |
|---|---|---|
| **P1** — reference-property derivation is keyed on a definition's name | **adopted** | `agent-plugins-a7cg` |
| **P2** — what stales a snapshot manifest, and how to refresh it | **adopted** | `agent-plugins-u3ng` |
| **P3** — a criterion for the refused outbound/anchor combination | **adopted** | `agent-plugins-f6lh` |
| **P4** — which rule identifier carries the unverified report | **rephrased, awaiting review** | — |
| **P5** — why AC-SPECMD-001.3 is load bearing | **not yet reviewed** | — |
| **P6** — legality of the slug-less anchor form | **adopted: exempt the form** | `agent-plugins-clau`, landed |

**P4** was returned as unclear rather than judged on its merits: "I read
this a couple times and I don't understand it. Rephrase for clarity or
we'll decline this proposal." The proposed FRD text has been rewritten
above — four short paragraphs where there had been one sentence carrying
four ideas — with its substance unchanged. It awaits a second read.

**P6** was decided against the alternative this proposal had left open.
An anchor-only reference naming a real heading of the document it sits in
is valid and must not be reported; [[adr-012]] keeps its slug-less bullet
and AC-SPECANCH-002.6 stands. Implemented and closed as
`agent-plugins-clau`, which added `@AC-SPECREF-002.12` to
[[frd-spec-reference-integrity]].

**P6 took two follow-ups to complete.** Exempting the form from
`reference/unresolved` made it legal but left it invisible: two consumers
still keyed off the slug half resolving, and an empty slug never
resolves. Both were measured after the exemption landed.

- `agent-plugins-etyf` (P1) — **landed.** `spec refs` omitted such a
  reference from its answer and `spec links` emitted the wikilink
  unrewritten. A wrong answer rather than a failing one, seen by no gate,
  which is what made it worse than the check defect P6 fixed. Now
  `corpus.indexReferences` records the reference as inbound on the
  document holding it (`@AC-SPECREF-001.8`) and `mdlinks` rewrites it to
  the heading's fragment (`@AC-SPECMD-001.8`). All three consumers — the
  check, the graph and the rewrite — now ask one predicate,
  `references.ResolvedSelfAnchor`.
- `agent-plugins-oksv` (P2) — open. A slug-less anchor reaching no
  heading yields two findings at one span for one typo, contradicting
  AC-SPECANCH-002.8. Out of scope for P6, which settled whether the form
  reports rather than how many findings it makes.

**A wording defect this exposed, which append-only forbids repairing in
place.** `AC-SPECREF-001.3` derives inbound references "from the outbound
references of **all other** corpus members". Read strictly, that makes the
omission conformant. It is not a prohibition — its Then clause states
what must be included, not what must not — so `AC-SPECREF-001.8` extends
it rather than contradicting it, and says so in its own rationale. The
phrase itself cannot be corrected without renumbering.
