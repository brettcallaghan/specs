---
title: "RFC: Implementing reference integrity — retirement collides with
  the corpus gate, declared verification levels name harnesses that
  cannot reach them, and folded deltas leave their own contradictions
  standing"
description: Order retirement ahead of the corpus gate and stop durable
  documents declaring edges into feature blueprints; extend the
  reachability rule to the harness that owns a level; require a fold to
  retract what it falsifies; and correct two pending RFCs whose proposed
  wording this implementation phase has already disproved
slug: rfc-0015
type: rfc
status: ready-for-review
updated_at: 2026-08-07T23:30:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: "claude-opus-5"
effort: unknown
runs_observed: 1
severity: snag
plugin_version: "0.15.0"
---

# RFC: Implementing reference integrity

## Summary

The ten work orders of `agent-plugins-9kb` were implemented against
[[frd-spec-reference-integrity]], the (now retired) reference-integrity
feature blueprint, and the three architecture blueprints the deltas fold
into. The documents held up: every delta landed, all 24 acceptance
criteria carry citations, and the profile self-application gate is green.
What they did not do is survive their own completion. Retiring the
blueprint immediately broke the corpus gate, because the ADR that the
blueprint applies references it and `reference/retired-target` — a check
this very feature built — fires on the edge. Three AC Coverage rows
declared `integration` for criteria that the snapshot harness is
structurally incapable of reaching. Four "not built yet" statements were
still standing in the two module blueprints after every delta had folded
its own change. The proposals fix the documents that survive the
blueprint's retirement, and correct two pending RFCs whose proposed
wording this phase has already shown to be wrong.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins |
| Phase executed | Implemented the ten work orders of `agent-plugins-9kb` (stage 1: MODEL, XPT, REX, RES, PREG; stage 2: CHK, LIFE, CMD; contract revision; Definition-of-Done verification) |
| Documents exercised | [[frd-spec-reference-integrity]]; `specs/features/reference-integrity/reference-integrity-blueprint.md`; [[module-document-core]], [[module-spec-profile]], [[container-spec-cli]]; [[contract-specs-configuration]]; [[testing-standards]], [[document-relationship-standards]]; [[adr-012]], [[adr-015]], [[adr-016]], [[adr-018]] |
| Documents produced | The ten work orders' folds into the three architecture blueprints and the contract; this RFC |
| Skill(s) followed | bead-bonanza v1.3; retro-spec v1.3.0 |
| Templates / standards read | feature-blueprint-template v1.4; work-order-template v1.3; rfc-template v1.2; `testing-standards.md` v0.1; `document-relationship-standards.md` v0.1 |
| Plugin version | 0.15.0 |
| Session date | 2026-08-07 |

### Completeness of the evidence base

The findings below are harvested from `retro:` comments recorded on the
work orders per `rfcs/findings-capture.md`, plus the per-item retro
anchors written under `/tmp/claude/bead-bonanza/1786121603-ri/`, plus
the Definition-of-Done verification itself. **The base is not uniform,
and this RFC does not present it as though it were:**

- Seven items — `agent-plugins-9kb.1`, `.2`, `.3`, `.4`, `.8`, `.9`, and
  the prerequisites `-tix`, `-ml9`, `-j7f` — carry both a full retro
  anchor and their `retro:` comments.
- Three items — `agent-plugins-9kb.5` (PREG), `.6` (CHK), `.7` (LIFE) —
  have **no retro anchor at all**. Their dispatch briefs were
  hand-written rather than filled from the per-bead template, which cost
  them the ceremony. Each still left `retro:` comments (one, one and two
  respectively), so they are represented here — but their approach
  narratives, implicit-decision lists and gate records are gone. Their
  landing commits are `bb7ecc3` (with WIP `e3f2777`), `a43bcf9` and
  `da9d8bd`.
- `agent-plugins-koi` (authoring the blueprint and [[adr-018]]) recorded
  no findings; its phase is retroed by `rfcs/inbox/frd/rfc-0013`.

So: nothing from this feature contributed *nothing*, but three of ten
items contributed only their comments. Where a proposal below rests on a
single comment from one of those three, it says so.

## Evidence

1. Retiring the feature blueprint broke the profile self-application
   gate on the first run. `[[adr-018]]` is `status: approved`, which the
   shipped profile registers as live, and its References section carried
   `- [[feature-spec-reference-integrity]] — the feature blueprint that
   applies this decision` (`specs/adrs/ADR-018-references-as-index-edges.md`,
   References). Setting the blueprint to `status: retired` produced
   exactly one finding —
   `reference/retired-target ... "feature-spec-reference-integrity" is
   retired` at `ADR-018:103` — and
   `TestProfileSelfApplicationAgainstThisRepositorysCorpus` asserts
   exactly zero. Two Definition-of-Done items therefore contradict each
   other: "The profile self-application gate ... is green" and "This
   blueprint's status → retired". Nothing in either item, in
   [[document-relationship-standards]], or in the blueprint's Open
   Questions warns that the second falsifies the first.

2. Three rows of the blueprint's AC Coverage table declared
   `integration` for criteria the snapshot harness cannot reach.
   AC-SPECREF-002.9 needs a persisted index whose reference row
   disagrees with a document on disk, which `corpus.Open`'s
   all-or-nothing content fingerprint makes unreachable from files;
   AC-SPECREF-003.4 needs a warm-served row no document carries;
   AC-SPECREF-004.3 needs a profile whose registration is defective, and
   a profile is compiled in rather than authored in the corpus. All
   three are barred by the same rule — [[testing-standards]] requires a
   snapshot case to "consist of input files and expected artifacts, and
   SHALL NOT contain executable test code", so a case has no way to
   plant that state. The implementer of `agent-plugins-9kb.8` recorded
   the cost twice: "I nearly added the citation to
   `cmd/snapshot_test.go`'s own AC comment on the assumption every
   listed AC needed a CMD-owned test, then caught that snapshot cases
   assert content, not latency, and would have been a false claim"; and
   "`AC-SPECREF-004.3` ... cannot be exercised through the CLI's own
   snapshot-case harness ... The item's Anchors ... don't call out that
   this AC is unreachable end-to-end through the process boundary." The
   implementer of `agent-plugins-9kb.6` paid it a third time: working
   out that AC-SPECREF-002.9's mechanism "only fires for a Corpus held
   in memory across an edit (not through the CLI's own
   `Open()`+`CheckFor()` path) took real analysis the design text didn't
   supply." Definition-of-Done verification then had to correct all
   three rows.

3. Four statements asserting a component was unbuilt were still standing
   in the two module blueprints after every delta had folded. In
   `specs/modules/document-core/document-core-module.md`: "The emitted index
   carries the resolved reference graph; no check reads it yet, so a
   broken link is recorded and not reported" (Components); "There is no
   `cli/core/markdown` package and no markdown parser in the module's
   dependencies" (Markdown Structure) — `goldmark v1.8.5` has been a
   direct dependency since REX; and "Where one names Markdown Structure
   or References it describes the intended edge rather than an existing
   one, since neither component is built yet ... Today it delegates to
   Frontmatter alone" (component relationships) —
   `cli/core/corpus/corpus.go:19`
   imports `cli/core/references`. In
   `specs/modules/spec-profile/spec-profile-module.md`: "the checks that will
   read them belong in [[module-document-core]] ... (Those Document Core
   checks are themselves later work — `agent-plugins-9kb.6` and `.7` —
   which this module does not gate on)". This is the second occurrence
   of the same failure: `agent-plugins-9kb.4` recorded that
   "`document-core-module.md`'s Key Decisions still read 'goldmark is
   not a dependency yet and arrives with the Markdown Structure
   component', stale since REX. Folding this document for RES was the
   first thing to touch that line." Each delta added its own positive
   statement and none was obliged to retract the negative one it
   falsified.

4. `rfcs/inbox/feature-blueprint/rfc-0014` P2 proposes wording that this
   phase has already disproved. It would add to the LIFE delta: "The
   pairing and typed-reference fixtures SHALL be supplied by a profile
   registered for the test through the `Registry` seam
   [[container-spec-cli]] describes". The implementer of
   `agent-plugins-9kb.7` followed exactly that instruction from the work
   order and reported: "'Register a test profile through the same
   `Registry` seam the command surface uses' initially read as literally
   `cli/cmd`'s `Registry`/`BuiltinRegistry` (used only by `cmd.Execute`,
   which hardcodes `BuiltinRegistry()` with no test-injection point).
   That reading is a dead end ... Consider rewording the Design text in
   future work orders to say 'the `corpus.Profile` extension point'
   explicitly, since 'Registry' names a concrete, narrower type in this
   codebase."

5. `rfcs/inbox/frd/rfc-0013` P1 proposes a criterion narrower than the
   interface the implementation required. Its proposed AC-SPECREF-004.4
   asks the configuration to determine "which `status` values the
   profile treats as live, which it treats as retired, and which
   frontmatter property names a retired document's replacement" — three
   facts. The implementer of `agent-plugins-9kb.2` found three
   insufficient: "the AC 'All three defect classes are detected: ... a
   lifecycle status outside the profile's vocabulary' cannot be
   satisfied from those three alone — the host has no way to learn the
   profile's status vocabulary. I added a fourth member,
   `LifecycleRules.Statuses`, rather than derive it from the union
   schema's `#/definitions/Status` (which would pass vacuously for a
   profile that keeps its enum elsewhere)."

6. Two criteria could not be implemented from their own text because
   they name a match or a value the index does not record.
   `agent-plugins-9kb.4`: "Neither the blueprint's RES section ... nor
   AC-SPECREF-002.7 says what the slug is matched against for a
   non-member, and the index records no slug for one — an
   `IndexExclusion` carries path/reason/type, and a parse-failure entry
   has no readable frontmatter by definition. The matching rule (file
   name without extension, lowercased) had to be invented to satisfy the
   criterion." `agent-plugins-9kb.3`, on the same model: "The
   `Reference` model documents `Text` as 'the raw value of a
   reference-bearing frontmatter property', which is ambiguous between
   the YAML source text and the parsed scalar. Taken literally a quoted
   value would give `Text="\"adr-012\""` and a slug half carrying
   quotes, so every quoted frontmatter slug would report unresolved."
   The same item also recorded that "Nothing in the bead or blueprint
   fixes what a reference's `Span` covers — the text between the
   brackets or the brackets too."

7. A capability split across two deltas spent the interval between them
   built and unwired, with no document recording the interval.
   `agent-plugins-9kb.2`: "The bead says 'Interface shape and function
   signatures are yours to choose' but does not say where
   `ValidateRegistration` should be called from. I left it exported and
   uncalled — 'WHEN a profile is activated' is `cli/cmd`'s `Load` path,
   which is CMD (`agent-plugins-9kb.8`) — so between this bead and CMD
   the validation exists and nothing runs it." The mirror image landed
   on LIFE: "The acceptance bullet 'A check whose registration is empty
   is still requested, and reports passed rather than not-requested'
   cannot be proven end-to-end from this bead's own scope ... the
   requested/passed/not-requested accounting is entirely built in
   `cmd/check.go` ... and the Design's own Anchors list only
   `cli/core/corpus/check.go` and `cli/core/finding/finding.go`, not
   `cli/cmd`." And on REX: "the self-application gate passed unchanged
   because `cli/profile` still registers no reference property, so the
   fold is verified by nothing until PREG lands."

8. A delta's Anchors asserted a fact about existing code that was false.
   The RES anchors state that the emitted index's "`input_hash` already
   folds the active profile's registration, so a profile that registers
   a new reference property invalidates the warm index by construction."
   `agent-plugins-9kb.4`: "It did not — `cli/core/corpus/hash.go` folded
   only `SlugPattern` and `Numbered` per type. Folding
   `ReferenceProperties` had to be written as part of this delta rather
   than relied on as given." A second, smaller instance:
   `agent-plugins-9kb.2` found that "The bead's anchors name 'the
   boundary tests in `cli/core/corpus`' but not the second `Profile`
   test double in `cli/cmd/config_test.go`, which has the same shape and
   also had to gain `Lifecycle()`."

9. An acceptance criterion required a property with no observable
   outside the package that holds it. `agent-plugins-9kb.4`: "The AC
   'Extraction and resolution ride the existing corpus walk; no document
   is read twice' has no observable from outside `cli/core/corpus` —
   every read goes through one unexported walk with no seam. It is held
   by an internal test asserting each candidate is offered exactly once
   plus a 1000-document budget measurement, rather than by counting
   reads directly."

## Proposals

### P1 — Retire before gating, and keep durable edges out of blueprints

**Problem.** The blueprint's Definition of Done lists "The profile
self-application gate runs the built profile over this repository's own
`specs/` corpus and the run is green" and, four items later, "This
blueprint's status → retired". Finding 1 shows the second falsifies the
first whenever any live document references the blueprint. This is not
specific to this feature: a feature blueprint retires when its
Definition of Done is complete — the blueprint's own Open Questions say
so, "a feature blueprint retires when its Definition of Done is complete
rather than because something replaced it" — so *every* completed
feature in this repository will now trip `reference/retired-target` on
whatever cites its blueprint. [[document-relationship-standards]] has no
rule about which document types may be cited by a durable one; its
"Declare Once, Derive the Reverse" section governs duplication, not
lifetime.

**Change.** Two parts, both additive.

1. Add to `specs/standards/document-relationship-standards.md`, under
   "Choosing Which End Retains the Edge":

   > A document the profile registers at a live status SHALL NOT declare
   > a relationship to a feature blueprint, in prose or in frontmatter.
   > The feature blueprint declares its own edges — to its FRD, to the
   > blueprints it folds into, to the decisions it applies — and the
   > reverse direction is answered by `spec refs <slug>`.
   >
   > **Rationale:** a feature blueprint retires when its Definition of
   > Done is complete, not because something replaced it, so an edge
   > declared *into* one is an edge that becomes a
   > `reference/retired-target` warning on the day the feature ships. The
   > durable end of the relationship is the ADR, the FRD or the
   > architecture blueprint; those stay live, and the graph already
   > carries the direction that matters.

2. Reorder the Definition of Done in
   `plugin/skills/spec/references/feature-blueprint-template.md`'s
   generated shape *as this project uses it*: the retirement item moves
   above the corpus-gate item, and the corpus-gate item reads "…is green
   **with this blueprint already retired**". A gate run against a
   not-yet-retired blueprint tests a state that will never exist again.

**Evidence.** Finding 1. Applied to this feature: the offending bullet
was removed from `ADR-018`'s References at Definition-of-Done time, and
the gate is green with the blueprint retired. Part 1 prevents the next
one; part 2 makes the ordering that discovered it repeatable.

**Backward compatibility.** Part 1 is a new rule that the current corpus
already satisfies (`ADR-018` was the only violation, and it is fixed).
Part 2 reorders a checklist; no criterion or delta changes.

### P2 — Extend reachability from an unbuilt mechanism to an unreachable
harness

**Problem.** `rfcs/inbox/standards/rfc-0011` P3 proposes that "An
acceptance criterion's declared verification level SHALL name a level at
which a test can exist against the mechanism the criterion describes",
and gives the failure condition as "IF no test at the declared level can
exist **because the mechanism is unbuilt or unregistered**". Every one
of its four cited instances is of that shape. Finding 2 is a third
cause the proposed rule does not cover: the mechanism is built,
registered and working, and the level is still unreachable because the
harness that owns it is structurally forbidden from staging the input.
A snapshot case "SHALL NOT contain executable test code"
([[testing-standards]], Snapshot Case Structure), so it cannot plant a
warm index row, cannot make one file's bytes disagree with an aggregate
fingerprint, and cannot register a compiled-in profile that is
deliberately broken. Under rfc-0011 P3 as written, all three of these
criteria pass the rule and stay wrong.

**Change.** Amend the rule rfc-0011 P3 proposes for
`specs/standards/testing-standards.md`, Coverage, to read:

> An acceptance criterion's declared verification level SHALL name a
> level at which a test can exist against the mechanism the criterion
> describes, **using the harness that owns that level**.
>
> IF no test at the declared level can exist — because the mechanism is
> unbuilt or unregistered, **or because the state the criterion is about
> cannot be staged by that harness** — THEN the criterion SHALL either
> carry the `(manual)` marker or be re-declared at the highest reachable
> level, and the reason SHALL be recorded against the criterion.
>
> **Rationale:** `integration` in this project means a snapshot case,
> and a snapshot case is pure data. A criterion about a warm index that
> disagrees with disk, about latency, or about a compiled-in
> registration cannot be a case however completely it is built — the
> level is unreachable by construction rather than by schedule.

and add to the feature-blueprint AC Coverage table convention, in the
same standard:

> WHERE a criterion is declared at a level other than the one its
> proving delta's harness would suggest, the coverage table SHALL name
> the test that verifies it.

**Evidence.** Finding 2, plus finding 9 as the same shape one layer
down — an AC whose property has no observable outside the package that
holds it is a level-reachability question the standard has no vocabulary
for. Applied to this feature: the three rows were re-declared `unit` at
Definition-of-Done time with a footnote naming each test and the reason.
The next blueprint declares the reachable level the first time, and
`agent-plugins-9kb.8`'s near-miss — adding a citation that would have
been "a false claim of coverage" — does not recur.

**Backward compatibility.** Extends a rule that is itself still
`ready-for-review`; adopt with rfc-0011 P3 or after it. It constrains
future criteria and gives existing ones the same defined resolution
rfc-0011 P3 already offers.

### P3 — Require a fold to retract the statements it falsifies

**Problem.** `rfcs/inbox/standards/rfc-0011` P6 proposes that "Every
component in a module or container blueprint SHALL carry an explicit
build state" — which addresses a component declared and never built.
Finding 3 is the opposite direction and is not covered: a component that
*was* built during this plan, leaving four assertions of the previous
state alive in two `status: approved` blueprints after every delta had
folded. Each delta's fold instruction names what to add ("References
marked **built** with the package that realizes it") and nothing names
what to remove. The negative statements are not in the component's own
text, where P6 would put the build state — they are in the Components
summary, in a sibling component's section, and in a relationship
paragraph, which is exactly why eight consecutive folds walked past them.

**Change.** Add a third amendment to rfc-0011 P6, landing in the Change
Guidance of `specs/modules/document-core/document-core-module.md` and
`specs/modules/spec-profile/spec-profile-module.md`:

> WHEN a component's build state changes, every statement in this
> document that asserted the previous state SHALL be retracted in the
> same change — including statements outside that component's own
> section: the Components summary, the diagram's parenthetical, any
> sibling component's text, and the component-relationship paragraphs.
>
> **Rationale:** a fold adds the positive claim to one section, and the
> negations live in four others. A blueprint that says both is worse
> than one that says neither, because a reader who finds the negation
> first stops looking.

**Evidence.** Finding 3, and its first occurrence recorded by
`agent-plugins-9kb.4` one delta earlier. Applied to this feature: all
four statements were corrected at Definition-of-Done time. The next fold
retracts as it adds, rather than accumulating contradictions for the
final item to sweep.

**Backward compatibility.** Additive prose in two Change Guidance
sections. Adopt independently of rfc-0011 P6 or with it; the two are
complementary halves of one rule.

### P4 — Correct rfc-0014 P2: name the extension point, not `Registry`

**Problem.** Finding 4. `rfcs/inbox/feature-blueprint/rfc-0014` P2 is
still `ready-for-review`, and the wording it proposes — "through the
`Registry` seam [[container-spec-cli]] describes" — is the wording the
work order already carried, and it sent the implementer of
`agent-plugins-9kb.7` down a dead end: `cmd.Execute` hardcodes
`BuiltinRegistry()` and offers no test-injection point, so a test
profile can never reach a check through it. Adopting P2 as drafted would
write the dead end into the blueprint template's exemplar and reproduce
the same loss on the next feature that needs a fixture profile.

**Change.** Replace the quoted sentence in rfc-0014 P2's **Change**
block with:

> The pairing and typed-reference fixtures SHALL be supplied by a
> profile registered for the test through the `corpus.Profile`
> extension point [[module-document-core]] defines, never by relaxing
> PREG's no-pairing registration. `Registry` in [[container-spec-cli]]
> names the `[profile].name` lookup used by `cmd.Execute`, which
> hardcodes `BuiltinRegistry()` and has no test-injection seam; the
> extension point is what a test can substitute at.

**Evidence.** Finding 4. Zero-cost, and it prevents a known failure
rather than a hypothetical one — the failure has already been observed
once, by the agent following the instruction P2 would enshrine.

**Backward compatibility.** An edit to an unresolved RFC's proposed
text. Nothing in the corpus depends on it yet.

### P5 — Correct rfc-0013 P1: the lifecycle carries the status vocabulary

**Problem.** Finding 5. `rfcs/inbox/frd/rfc-0013` P1 is still
`ready-for-review` and proposes an AC-SPECREF-004.4 naming three
registered facts. The shipped implementation needs four: `LifecycleRules`
in `cli/core/corpus/profile.go` carries `Statuses` alongside
`LiveStatuses`, `RetiredStatuses` and `SuccessorProperty`, because
AC-SPECREF-004.3's third defect class — a lifecycle registration naming
a status outside the profile's own vocabulary — is undetectable without
it. [[contract-specs-configuration]] already documents all four ("the
full status vocabulary its documents may declare, which of those
statuses a document is in effect at (live) and which it is no longer in
effect at (retired), and the frontmatter property in which a document at
a retired status names its replacement"). Adopting P1 as drafted would
leave the FRD asking for three where the contract and the code carry
four.

**Change.** Amend rfc-0013 P1's proposed criterion to:

> **AC-SPECREF-004.4:** The configuration SHALL determine the full set
> of `status` values the profile's documents may declare, which of those
> values the profile treats as live, which it treats as retired, and
> which frontmatter property names a retired document's replacement.
> - *Example:* The shipped configuration registers this framework's six
>   statuses as its vocabulary, `retired` as its only retired status,
>   every other status as live, and `superseded_by` as the successor
>   property — so AC-SPECREF-002.6 can quote a replacement, and
>   AC-SPECREF-004.3 can report a lifecycle naming a status the profile
>   does not define, without the engine knowing any of the names.

**Evidence.** Finding 5. The FRD then matches the contract and the code
that already exist, and the vacuous-derivation trap
`agent-plugins-9kb.2` avoided ("derive it from the union schema's
`#/definitions/Status`, which would pass vacuously for a profile that
keeps its enum elsewhere") is closed by specification rather than by one
implementer's judgement.

**Backward compatibility.** An edit to an unresolved RFC's proposed
text, describing behaviour that already ships.

### P6 — A criterion matching on a value the index omits SHALL name the key

**Problem.** Finding 6. AC-SPECREF-002.7 requires the CLI to report that
a reference's slug "matched a non-member", and the index records no slug
for a non-member: an `IndexExclusion` carries path, reason and type, and
a parse-failure entry has no readable frontmatter by construction. The
criterion is therefore unimplementable as written, and the implementer
had to invent the match key (file name without extension, lowercased) to
satisfy it at all. The same shape recurred on `Reference.text`, whose
doc comment is ambiguous between raw YAML source and parsed scalar, and
on a reference's `Span`, whose extent no document fixed. All three
survive the blueprint's retirement, because the FRD and the Strata
sources are what [[frd-spec-anchor-references]] and
[[frd-spec-slug-rename]] will be authored against.

**Change.** Two document edits, both to surviving documents.

1. In `specs/features/reference-integrity/reference-integrity-frd.md`,
   extend AC-SPECREF-002.7's example to name the match key:

   > - *Example:* `[[work-order-template]]` matches
   >   `plugin/skills/spec/references/work-order-template.md`, which the
   >   index recorded as unparsable. A non-member has no declared slug —
   >   an exclusion carries only its path and reason, and a parse failure
   >   has no readable frontmatter — so the match is against the file's
   >   name without its extension, lowercased. The finding names that
   >   file and the reason it is not a member, rather than reporting a
   >   slug that matches nothing.

2. In `schemas/index.strata`, replace `Reference.text`'s doc comment
   with one that says which of the two readings applies to a frontmatter
   origin (the parsed scalar, not the quoted source) and state the span
   extents on `Reference.position` (the text between the brackets for a
   body reference; the value's own source text, quotes included, for a
   frontmatter one). [[module-document-core]]'s References component
   already records both; the generated model is where the next
   implementer reads them.

**Evidence.** Finding 6. The two sub-features that build on these rows
inherit a specified match key and an unambiguous `text`, instead of
re-deriving one implementer's invention from the code.

**Backward compatibility.** Both edits describe shipped behaviour.
Edit 2 regenerates `cli/core/specmodel`, `cli/core/schema` and the
plugin markdown via `scripts/generate-schemas.sh`; no field changes
shape, so no snapshot and no `input_hash` moves.

### P7 — Record the interval in which a capability is built and unwired

**Problem.** Finding 7. The blueprint's Sequencing Constraints state
which delta lands before which, and say nothing about the state of the
system in between. Three items paid for that separately: XPT left
`ValidateRegistration` exported and called by nothing until CMD; LIFE
owned a criterion whose accounting lives in `cli/cmd`, outside the
anchors LIFE declares; REX's fold was "verified by nothing until PREG
lands". Each is correct work order by work order and each left the
implementer deciding, alone, whether an unwired capability was their
defect or their successor's scope. The existing constraints ("The XPT
delta SHALL land before PREG … and before LIFE, which reads it") state
the order and imply nothing about the gap.

**Change.** Add to the Sequencing Constraints guidance the project
applies to a staged feature blueprint — recorded in
`specs/standards/testing-standards.md`, Coverage, since it is a
statement about what is verified when:

> WHERE a capability is split across delta blocks such that it is
> present but unreachable between them, the blueprint SHALL name the
> interval, the delta that closes it, and what verifies the capability
> while it is open.
>
> **Rationale:** an implementer who finds a capability built and called
> by nothing cannot tell an incomplete delta from a correctly-scoped
> one, and the cheapest resolution — wiring it up "while I am here" —
> is the one that breaks the delta boundary the staging exists to
> create.

**Evidence.** Finding 7, three independent instances in one feature.
This is adjacent to `rfcs/inbox/feature-blueprint/rfc-0014` P1, which
splits one constraint into an ordering and a verification-timing
statement; P1 fixes one constraint's ambiguity, this proposal adds the
class of statement that was missing from all of them.

**Backward compatibility.** Additive. Existing blueprints stay valid;
the rule binds the next staged one.

## Considered and Not Proposed

- **A delta's Anchors must cite the file and symbol for any claim about
  existing behaviour** (finding 8, the false `input_hash` claim and the
  missed second `Profile` stub). Real and costly, but the concrete
  change is a rule about how the *feature-blueprint template* asks
  Anchors to be written, which is the `retro-skill` lens rather than
  this one. Routed there rather than filed here; the specific claim is
  moot, since the blueprint is retired and `hash.go` now folds
  `ReferenceProperties`.
- **Record the `specmodel`-versus-extension-point trap in
  [[module-document-core]]'s Change Guidance.** `agent-plugins-9kb.9`
  lost time distinguishing "the real, live `corpus.Profile` Go
  interface" from "the generated `specmodel.go` data model that happens
  to share vocabulary (`DocumentTypeRegistration`, `Section`,
  `BlockTagRegistration` exist in both, but only one is actually wired
  to a consumer)". The item itself routed its suggestion to "a prompt or
  skill note", and the Interface section now names the four live members
  explicitly, which carries the same signal. Recorded here so the next
  retro does not re-derive it.
- **`.strata` doc comments and `[[slug]]` wikilinks** (`agent-plugins-9kb.1`:
  "this feature is precisely the one that makes an unresolvable wikilink
  in a doc comment matter"). One observation, and the generated markdown
  ships outside the `specs/` corpus root, so no check can see it either
  way. Not enough evidence for a rule.
- **`c4_docs` metadata and the acceptance-criteria list disagreeing on
  which documents a work order folds** (`agent-plugins-9kb.1`: "the two
  fields disagree with no note saying which governs"). A work-order
  template question — `retro-skill` lens.
- **The strata language reference is silent on whether a field-less
  struct is a valid union variant** (`agent-plugins-9kb.1` had to probe
  the compiler before it could choose between a discriminated union and
  a flat enum). A skill-reference gap; `rfcs/inbox/strata/` already
  holds an RFC against that skill.
- **How much of a reconciled component's documentation to remove**
  (`agent-plugins-9kb.5`: the AC "gave no guidance on how much of the
  existing `### Lifecycle Rules` subsection, its component fence, and
  the Components mermaid diagram to keep versus remove"). One
  observation, from an item with no retro anchor, and the agent resolved
  it by reading the blueprint's own Definition of Done — which is the
  intended fallback working.
- **`cmd/check.go`'s `checkRules` omission** (`agent-plugins-9kb.8`
  found three rule identifiers missing, masked by `finding.Report.Add`
  auto-registering any rule that produces a finding). An implementation
  defect the item found and fixed, now pinned by
  `testdata/cases/check-only-reference-rule`. Not a document defect.
- **Bumping `updated_at` where no acceptance criterion asks for it**
  (`agent-plugins-9kb.2`). The frontmatter convention already covers it;
  one-off.
- **This RFC's own frontmatter misuses `skill` and `skill_version` to
  describe a document-audience RFC.** Already the subject of
  `rfcs/inbox/feature-blueprint/rfc-0010` P2 and
  `rfcs/inbox/feature-blueprint/rfc-0014` P3; the local shape is
  followed here rather than relitigated.

## Non-Goals

- No change to the severities of the three `reference/*` rules, or to
  the exit-status contract they feed. The self-application gate is green
  at the severities these checks assign, and P1 fixes the corpus rather
  than loosening the gate or the lifecycle registration.
- No new document type, and no change to the retirement semantics of a
  feature blueprint. P1 keeps durable documents off the edge; it does
  not make a retired blueprint live again.
- No weakening of [[testing-standards]]'s snapshot-case purity rule. P2
  takes "a case is pure data" as a fixed constraint and moves the
  declared level to match it, not the reverse.
- No re-grooming of `agent-plugins-9kb`. Every proposal here lands on a
  document that survives the blueprint's retirement.

## Adoption

**Independent and adoptable alone:** P1, P3, P6, P7. Each is additive
text in a surviving document (`document-relationship-standards.md` and
the feature-blueprint DoD ordering; the two module blueprints' Change
Guidance; the FRD and `schemas/index.strata`; `testing-standards.md`).

**Bound to a pending RFC:** P2 amends the rule
`rfcs/inbox/standards/rfc-0011` P3 proposes — adopt it with P3 or
immediately after; adopting rfc-0011 P3 alone leaves finding 2's three
criteria passing a rule they violate in spirit. P4 and P5 are edits to
the proposed text of `rfcs/inbox/feature-blueprint/rfc-0014` P2 and
`rfcs/inbox/frd/rfc-0013` P1 respectively, and should be applied to
those RFCs **before** either is resolved; both are zero-cost if applied
first and require a follow-up edit to a landed document if not.

**Complementary:** P3 is the reverse direction of
`rfcs/inbox/standards/rfc-0011` P6. Adopting both gives one rule with
two halves — a component's build state is stated where it lives, and
every statement of the previous state goes when it changes.

**Touch points.** `specs/standards/document-relationship-standards.md`
(P1, P7 if filed there instead); `specs/standards/testing-standards.md`
(P2, P7); `specs/modules/document-core/document-core-module.md` and
`specs/modules/spec-profile/spec-profile-module.md` Change Guidance (P3);
`specs/features/reference-integrity/reference-integrity-frd.md` and
`schemas/index.strata` (P6); the three pending RFCs (P2, P4, P5).

**Expected effect.** The next feature this repository completes retires
its blueprint without breaking its own corpus gate, declares
verification levels its harnesses can reach, and folds its deltas
without leaving four contradictions for the final item to find — and the
two RFCs already queued for review carry wording that the implementation
phase has confirmed rather than wording it has disproved.

## Resolution
