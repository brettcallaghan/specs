---
title: "Feature Blueprint: Gherkin Features and Acceptance Criteria"
description: Builds a Gherkin component and the fenced-block half of
  Markdown Structure in Document Core, adds feature and criterion records to
  the index, widens the profile extension point to carry the tag vocabulary,
  and exposes the records as checks, queries and emitted feature files
slug: feature-gherkin-adoption
type: feature-blueprint
status: retired
updated_at: 2026-08-15T00:00:00Z
delivery: staged
plan: agent-plugins-wvd
frd: frd-gherkin-adoption
---

# Gherkin Features and Acceptance Criteria

## Overview

**FRD:** [[frd-gherkin-adoption]]

[[adr-021]] settles the notation and the parser; nothing reads either.
[[module-document-core]] declares a **Markdown Structure** component and
records that it is not built — there is no `cli/core/markdown` package
and no fenced block is extracted anywhere — so a ` ```gherkin ` block is
inert text that the References component actively masks. This feature
builds the fenced-block half of that component, adds a **Gherkin**
component beside it, and turns what the two produce into records the
index carries, checks the run reports, and queries an agent can ask.

The split that governs every delta below is the module boundary already
in force. Parsing Gherkin is general — an Obsidian vault with `.feature`
files could use it — so it lives in [[module-document-core]] and knows
tags only as strings. Which tag is an identifier, what shape it must
take, and which qualifiers mean anything are this framework's vocabulary
and arrive through the profile extension point, which this feature
widens for the second time. No `@AC-` literal appears in the host.

Two forms of source produce one shape of record. A `.feature` file is
discovered by a second pass beside the corpus walk rather than admitted
as a corpus member — AC-SPEC-001.6 stands unamended — and a fenced block
is found inside a member the walk already read.

The delta blocks in dependency order: **MODEL** (feature, criterion and
finding models in Strata) → **XPT** (the widened `Profile` extension
point) → **MDS** (fenced-block extraction) → **GHK** (the Gherkin
component) → **DISC** (discovery, ownership and the index collection) →
**PREG** (this framework's tag registrations) → **VTAG** (identity
checks) → **VCON** (content and diagnostic checks) → **GREF**
(references inside Gherkin) → **QRY** (feature and criterion queries) →
**EMIT** (feature file emission) → **CHG** (local change detection).

## Element Deltas

### Feature, criterion and finding models in Strata (added) — MODEL

**Intent:** Every other delta consumes generated types, so the models
grow first and at their source, per [[adr-016]]. The index gains
somewhere to put a feature and the criteria beneath it, and the finding
model gains the variants that make a Gherkin defect self-contained.
`schemas/feature.strata` is already authored on this feature's branch
and has never been compiled — compiling it, and fixing what the
transpiler reports, is the first act of this block.

**Data (Strata sources, `schemas/`):**

- `feature.strata` declares `FeatureCode`, `FeatureId`, `CriterionId`,
  `ContentHash`, `VerificationLevel`, `CriterionNotation`,
  `FeatureSource` (a `@discriminator("form")` union over
  `FeatureSourceBlock` and `FeatureSourceFile`), `Criterion` and
  `Feature`.
- `index.strata` gains one top-level `features` list on `SpecIndex`.
  Criteria are **nested** on `Feature` rather than held as a sibling
  collection — deliberately not the shape [[adr-018]] fixes for
  references, because a reference's attributes describe an edge while a
  criterion is a node belonging to exactly one feature, and its
  identifier already names which.
- `finding.strata` gains `FindingDetail` variants, each carrying enough
  to act on without reopening the document: a Gherkin grammar error with
  the parser's expectation; an untagged or malformed identifier with the
  offending text and the pattern it must match (AC-GHK-006.5); a
  duplicate identifier with the other occurrence's path and span; a
  partially tagged outline with the untagged blocks' spans; and an
  owning-document defect with the tag's slug and the containing
  document's.
- Behavior: `schema_version` SHALL remain `"1.0"`. The `features`
  collection is an addition alongside the existing shape, which
  [[container-spec-cli]]'s additive-at-1.0 rule permits, and the model
  and the emitted index change in one change as
  [[module-document-core]]'s lockstep rule requires. The satisfying act
  is the recorded no-bump, not an omission.
- Behavior: `features` SHALL be a list whose order the emitter fixes —
  by the path a feature occurs in, then by its position within that
  file — so AC-SPEC-009.7's byte-identical output is a property of how
  one list is filled rather than of a directory walk.
- Behavior: a `Criterion`'s `content_hash` SHALL be `sha256:` followed
  by the lowercase hexadecimal digest of the criterion's name and text
  and nothing else (AC-GHK-005.5).

**Anchors:** `schemas/feature.strata`, `schemas/index.strata`,
`schemas/finding.strata`; `scripts/generate-schemas.sh`, which
regenerates `cli/core/specmodel/`, `cli/core/schema/` and the plugin
markdown. Regeneration is the whole risk in this block — three
unverified constructs are named in `agent-plugins-dc6` under Tracker
Reconciliation.

### Profile extension point in [[module-document-core]] (modified) — XPT

**Intent:** Which tag identifies a feature, which identifies a
criterion, and which qualifiers carry meaning are facts about a document
vocabulary, not about Gherkin. The host must hold none of them, exactly
as it holds no status and no reference-bearing property name today. This
is that widening, reviewed on its own merits as
[[module-spec-profile]]'s Change Guidance requires, and framed so any
profile can express a tag vocabulary rather than so this one can.

**Behavior:**

- The `Profile` interface SHALL additionally supply a feature-tag
  registration: the pattern a feature identifier tag matches, the
  pattern a criterion identifier tag matches, the prefix an
  owning-document tag carries, and the qualifier tags the profile
  assigns meaning to together with what each one means (AC-GHK-002.1).
- The registration SHALL state how a criterion identifier derives from
  its feature's, so that the host can check the derivation without
  knowing either pattern's internal structure (AC-GHK-003.6).
- A tag matching no registered pattern and named by no qualifier SHALL
  be carried on the record and reported by nothing (AC-GHK-002.8).
- Document Core SHALL NOT contain the literal `AC`, `FEAT`, `frd`,
  `negative`, `manual`, or any registered level name. The existing
  boundary test — no non-test source in `cli/core/corpus` carries a
  value a profile registers — SHALL extend to the values this delta
  registers, so the widening cannot become a place to smuggle
  vocabulary into the host.
- WHEN a profile is activated, `ValidateRegistration` SHALL report a
  feature-tag registration whose patterns do not compile, and one whose
  criterion-derivation rule names a capture the identifier pattern does
  not define, naming the offending key rather than defaulting it.

**Anchors:** `cli/core/corpus/profile.go` (`Profile`,
`DocumentTypeRules`, `ValidateRegistration`); the boundary tests in
`cli/core/corpus`; the generated `specmodel.ProfileRegistration` that
MODEL extends.

### Markdown Structure in [[module-document-core]] (added — fences) — MDS

**Intent:** Build the fenced-block half of the component
[[module-document-core]] has declared and never built. This feature
needs one of its three capabilities: locate every fenced block whose
info string matches a registered tag, and hand back its content with
the two extents a caller needs. The heading tree and the comment-block
location stay unbuilt and belong to [[frd-spec-structure-conformance]];
this block does not start them.

**Behavior:**

- The component SHALL locate every fenced block in a document body whose
  info string matches a registered tag (AC-GHK-001.1).
- For each such block, the component SHALL return the content between the
  fences, an `Outer` span covering the whole construct including both
  fence lines, and an `Inner` span covering the content alone.
- Fence recognition SHALL come from `goldmark`, already a module
  dependency, rather than from a delimiter scan: where a fence ends
  depends on the fence character, its run length and its info string,
  and a scanner that guesses one of those wrong extracts a block no
  markdown reader would agree with. Every position SHALL still be
  computed by the Spans component, which is what the module's
  scanner rule is for — one way of computing a position, not one way of
  deciding where to look.
- The component SHALL report an offset within a block's content as an
  offset within the containing document, by adding the content's start
  (AC-GHK-004.2, AC-GHK-006.1).
- An unterminated fence SHALL yield no block, matching the Spans
  scanner's refusal to guess at a dangling delimiter.

**Anchors:** a new `cli/core/markdown` package; `cli/core/span`
(`Region`, `Of`); the `goldmark` usage already in
`cli/core/references`.

### Gherkin in [[module-document-core]] (added) — GHK

**Intent:** Parse a Gherkin source into the records the index carries.
This component holds the whole of what is true of Gherkin and none of
what is true of this framework: it reads tags as strings, attaches no
meaning to any of them, and reports what it could not parse. That is
what keeps it in the host rather than in the profile — the test
[[module-document-core]] states is whether a vault with nothing to do
with C4 could use it, and a vault with `.feature` files can.

**Behavior:**

- The component SHALL parse a source with the official Go Gherkin parser
  [[adr-021]] measured.
- The component SHALL produce one feature record per `Feature`, with its
  criteria nested beneath it in the order they are written
  (AC-GHK-005.4).
- A feature record SHALL carry its name, its description, its tags as
  written, and its span; a criterion record SHALL carry its name, its
  body as text, its own description, the name of the `Rule` it sits
  beneath where it sits beneath one, its tags, and its span
  (AC-GHK-005.1, AC-GHK-005.2).
- A criterion's text SHALL be its steps and any example table verbatim,
  with the block's common indentation removed (AC-GHK-005.3).
- A criterion's text SHALL NOT be decomposed into typed steps or keyed
  example rows.
- The component SHALL emit one criterion record per `Examples` block
  that carries an identifier tag, and one criterion record for a
  `Scenario Outline` whose `Examples` blocks carry none — in the latter
  case the record's text covers every block (AC-GHK-002.6,
  AC-GHK-002.9).
- WHERE a source's content begins with a keyword other than `Feature`,
  the component SHALL prepend a synthetic `Feature` header before parsing
  (AC-GHK-001.4).
- The component SHALL NOT report any position the synthetic header
  occupies (AC-GHK-001.5).
- The component SHALL report every position as its line in the containing
  document, computed as the source's start line plus the parser's line
  minus the number of lines the synthetic header added.
- The component SHALL compute a criterion's content hash over its name
  and its text alone, excluding every span and every parser node
  identifier — node identifiers are positional counters that shift
  whenever anything above them is edited, so a criterion that moved has
  not changed (AC-GHK-005.6, AC-GHK-005.7).
- The text a hash is taken over SHALL be normalized to LF line endings
  with the block's common indentation removed and trailing whitespace
  stripped from each line, so that a criterion reindented by a
  containing document's edit, or read from a CRLF checkout, hashes the
  same.
- A parse failure SHALL be returned as a located error carrying the
  parser's own expectation (AC-GHK-006.1).
- A parse failure SHALL NOT panic or abort a walk.
- Both source forms SHALL produce equal records but for their source
  location and their source form (AC-GHK-001.3).

**Anchors:** a new `cli/core/gherkin` package;
`github.com/cucumber/gherkin/go/v42` and
`github.com/cucumber/messages/go/v34`, the libraries [[adr-021]]
measured; `cli/core/span` for every position; the generated
`specmodel.Feature` and `specmodel.Criterion` from MODEL.

### Corpus in [[module-document-core]] (modified — discovery) — DISC

**Intent:** Wire the two sources into the walk and the index. A fenced
block is found inside a member the walk already read, so extraction
rides the same visit that reads its bytes, as reference extraction
already does. A `.feature` file is a second pass over the same tree.
This block also decides ownership, which is the one fact the two forms
establish differently.

**Behavior:**

- The component SHALL treat a regular file whose name ends in `.feature`,
  matched without regard to letter case, as a feature source beneath the
  spec root (AC-GHK-001.2).
- The component SHALL read a feature source through the same `os.Root`
  handle and the same CRLF normalization every other read passes
  through.
- A feature source SHALL produce neither an `IndexEntry` nor an
  `IndexExclusion`. It is not a `.md` candidate, so AC-SPEC-001.6's
  silent pass-over stands and no exception to corpus membership is
  introduced.
- The component SHALL attribute a feature written in a fenced block to
  the document the block is written in, requiring no tag to say so
  (AC-GHK-001.7), and a feature written in a feature source to the
  document its owning-document tag names (AC-GHK-001.2).
- The component SHALL record every extracted feature in `SpecIndex`'s
  `features` collection with its path, its source form, and the block's
  outer span where it was written as one (AC-GHK-005.8).
- `Fingerprint` SHALL fold every feature source's repository-relative
  path and content digest into the input hash, on the same walk, so that
  editing a `.feature` file invalidates a persisted index. A hash that
  covered only `.md` candidates would serve a stale answer to every
  query in this feature.
- The registered feature-tag patterns SHALL be folded into the input
  hash alongside the per-type table already there, because a changed
  pattern changes the index while the files stand still.
- A source that fails to parse SHALL be recorded as a failure against
  its own file and the walk SHALL continue, so indexing never fails as a
  whole (AC-SPEC-002.4).
- A feature source SHALL sit in the directory its owning document sits
  in; discovery searches the spec root and needs no location of its own
  (AC-GHK-001.6).

**Anchors:** `cli/core/corpus` (`Build`, `Fingerprint`, `Open`,
`Write`); `cli/core/markdown` from MDS and `cli/core/gherkin` from GHK.

### Tag registrations in [[module-spec-profile]] (modified) — PREG

**Intent:** Populate the extension point XPT widened with this
framework's vocabulary — the tag table [[frd-gherkin-adoption]]'s
AC-GHK-002.1 fixes, and nothing beyond it. As with every other
registration here, the rules are registered as data rather than
implemented as code, so that a reviewer can check a pattern against the
guide it mirrors.

**Behavior:**

- The module SHALL register the feature identifier pattern
  `^[A-Z]+-[A-Z0-9]+-[0-9]{3}$`, the criterion identifier pattern
  `^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`, and the owning-document tag prefix
  `frd-` (AC-GHK-002.1, AC-GHK-002.3).
- The module SHALL register the derivation rule that a criterion
  identifier's code and number are its feature's (AC-GHK-003.6).
- The module SHALL register `negative`, `manual` and the `level-`
  prefixed vocabulary — `unit`, `integration`, `e2e` — as the qualifiers
  it assigns meaning to, and nothing else (AC-GHK-002.1).
- The registration SHALL be checkable against the generated
  `CriterionId` and `FeatureId` patterns rather than transcribed
  independently, so that the model and the registration cannot drift.

**Anchors:** `cli/profile/profile.go`, `cli/profile/registration.go`;
the drift tests in `cli/profile`; the generated patterns from MODEL.

### Identity checks in [[module-document-core]] (added) — VTAG

**Intent:** The Gherkin parser validates its own grammar and has no
concept of an acceptance criterion, so every rule here is the CLI's.
This block covers the identity half: is each record addressed, is it
addressed once, and does its address agree with where it sits. These are
the checks that make a tag an addressing scheme rather than a
convention.

**Behavior:**

- The component SHALL report `gherkin/criterion-untagged` for a
  `Scenario` carrying no identifier tag, and `gherkin/feature-untagged`
  for a `Feature` carrying none, each at severity `error` (AC-GHK-002.2,
  AC-GHK-006.2).
- The component SHALL report `gherkin/identifier-malformed` for an
  identifier tag not matching its registered pattern, quoting both the
  identifier as written and the pattern (AC-GHK-002.3, AC-GHK-006.5).
- The component SHALL report `gherkin/identifier-multiple` for an
  element carrying more than one tag matching an identifier pattern,
  naming both (AC-GHK-002.5).
- The component SHALL report `gherkin/identifier-duplicate` against each
  of two records in the corpus declaring the same identifier, naming the
  other's location, for criterion identifiers and for feature
  identifiers alike (AC-GHK-002.4, AC-GHK-003.4).
- The component SHALL NOT report anything for several features declaring
  the same code with different numbers: a code groups and does not
  identify (AC-GHK-003.5).
- The component SHALL report `gherkin/criterion-code-mismatch` for a
  criterion whose identifier's code and number are not its feature's
  (AC-GHK-003.6).
- The component SHALL report `gherkin/owner-missing` for a feature in a
  feature source carrying no owning-document tag (AC-GHK-002.10), and
  `gherkin/owner-mismatch` for a feature in a fenced block carrying an
  owning-document tag that names a document other than the one it is
  written in, naming both (AC-GHK-002.11).
- The component SHALL report `gherkin/owner-unresolved` for an
  owning-document tag whose slug names no corpus member. The FRD
  requires ownership to be recorded and a tag that names nothing cannot
  record it; the defect reads as the unresolved-reference defect it is
  rather than producing a feature attributed to a document that does not
  exist.
- The component SHALL report no finding for a tag matching no registered
  pattern and named by no qualifier (AC-GHK-002.8).

**Anchors:** `cli/core/corpus` (`Check`, `CheckFor`);
`cli/core/finding` constructors; the registered patterns from PREG.

### Content and diagnostic checks in [[module-document-core]] (added) — VCON

**Intent:** The other half of what the parser accepts and a reader would
not: a criterion that says nothing, prose written where a criterion
belongs, and an outline whose blocks are tagged inconsistently. This
block also fixes what a Gherkin finding looks like, which is what makes
the whole set actionable by an agent that never opens the standard.

**Behavior:**

- Every finding this feature raises SHALL carry a rule identifier in the
  `gherkin/` namespace, a severity, a path, a span and a message stating
  the violated constraint and the offending value (AC-GHK-006.3).
- The component SHALL report `gherkin/parse` for a grammar error, at the
  line and column of the offending text in the containing document, with
  the parser's own expectation in the message (AC-GHK-006.1).
- The component SHALL report `gherkin/criterion-empty` for a criterion
  with no steps (AC-GHK-006.2).
- The component SHALL report `gherkin/scenario-prose` for a non-empty
  `Scenario` description, as a probable untagged criterion.
- The component SHALL NOT report a `Feature` description, which carries
  the user story and is exempt (AC-GHK-006.4).
- The component SHALL report `gherkin/examples-partially-tagged` for a
  `Scenario Outline` where some `Examples` blocks carry an identifier
  and others do not, naming the untagged blocks — the check is partial
  tagging, which is unambiguous, and not the multi-block case itself,
  which is equally a single criterion in named groups (AC-GHK-002.7,
  AC-GHK-006.2).
- Severity SHALL be `error` for an untagged criterion, a duplicated
  identifier and a criterion with no steps, and `warning` for an absent
  `Feature` description and a criterion carrying no verification level
  (AC-GHK-006.6).
- WHEN one document is checked, the component SHALL report that
  document's findings (AC-GHK-006.7).
- WHEN one document is checked, the component SHALL resolve identifier
  uniqueness against the whole corpus, reporting a duplicate wherever its
  counterpart sits. This follows the exception `reference/unresolved`
  already carries: a finding is reported against the file a reader must
  open.

**Anchors:** `cli/core/corpus` (`Check`, `CheckFor`);
`cli/core/finding` (`Report`, the `FindingDetail` variants from MODEL);
`cli/testdata/cases/` for the JSON assertions.

### References in [[module-document-core]] (modified — Gherkin) — GREF

**Intent:** A wikilink is a reference wherever it is written, and today
one written inside a ` ```gherkin ` block is not: the References
component masks every fenced block, precisely so a backticked
illustration is not reported. That exclusion is right for an
illustration and wrong for a feature, and the difference is the info
string. This block narrows the mask and extends extraction to feature
sources.

**Behavior:**

- The component SHALL NOT mask a fenced block whose info string names a
  registered Gherkin tag, so a wikilink written in a feature
  description, a scenario name, a step, a rule name or an examples cell
  is extracted like any other (AC-GHK-004.1). Every other fenced block
  and every inline code span stays masked.
- A reference extracted from a fenced block SHALL carry the span of the
  wikilink in the containing document (AC-GHK-004.2).
- The component SHALL extract references from a feature source and
  record them against that source's path, even though the file is not a
  corpus member. An edge's endpoints are paths, and the referring file
  is the one a reader must open.
- A reference from a Gherkin source SHALL resolve exactly as one written
  in prose, and an unresolved one SHALL be reported by the existing
  `reference/unresolved` rule rather than by a rule of its own
  (AC-GHK-004.3).
- The component SHALL split a reference on its first `#` as it does
  elsewhere.
- The component SHALL NOT resolve the anchor half against a `Feature`,
  `Rule` or `Scenario` name — a scenario name is not a heading, and
  anchors resolve against headings (AC-GHK-004.4).

**Anchors:** `cli/core/references` (`Extract`, the `goldmark` code-region
mask); `cli/core/markdown` from MDS for the fence info string.

### Command Surface in [[container-spec-cli]] (modified — queries) — QRY

**Intent:** The queries are what make the records worth building: an
agent asks for one criterion instead of reading an FRD. They answer from
the persisted index, so they inherit the warm-index behavior the
container already has and add no latency budget of their own.

**Behavior:**

- The command tree SHALL gain `spec feature` and `spec criterion`, each
  taking an identifier argument and each opening the corpus warm as
  `check`, `slug-exists`, `next-id` and `refs` already do.
- `spec feature <id>` SHALL return that feature's record including its
  criteria (AC-GHK-007.1); `spec criterion <id>` SHALL return that
  criterion's record (AC-GHK-007.2).
- `spec feature --code <CODE>` SHALL return every feature grouped under
  that code, `spec feature` with no argument every feature in the
  corpus, `spec criterion --of <feature-id>` every criterion beneath one
  feature, and `spec criterion --tag <qualifier>` every criterion
  carrying one qualifier (AC-GHK-003.3, AC-GHK-007.3).
- WHEN an identifier names no record, the command SHALL report that
  nothing declares it (AC-GHK-007.4).
- WHEN an identifier names no record, the command SHALL exit non-zero,
  which is what distinguishes an unknown name from an empty selection.
- `--json` SHALL write the matching records as the generated JSON to
  standard output (AC-GHK-007.5).
- WHERE the index's inputs are unchanged, both commands SHALL answer from
  the persisted index without re-walking the corpus (AC-GHK-007.6).
- Both commands SHALL return a feature by identifier regardless of which
  file it now sits in (AC-GHK-003.1).
- Their `--help` text SHALL be cobra's rendering of the same
  declarations that define them, as every other subcommand's is.

**Anchors:** `cli/cmd` (`Execute`, the subcommand declarations);
`cli/core/corpus` (`Open`); `cli/testdata/cases/` and
`scripts/regenerate-cli-snapshots.sh`.

### Feature file emission in [[container-spec-cli]] (modified) — EMIT

**Intent:** Emission is what makes the fenced form the default rather
than a compromise: a feature authored inside an FRD is still
publishable, so no second authoring format is needed for external
product and quality tooling.

**Behavior:**

- `spec feature <id> --emit` SHALL write a valid Gherkin feature file
  carrying that feature's tags, description and criteria, to standard
  output by default and to a named path where `--out` gives one
  (AC-GHK-009.1). One feature yields one file, as [[adr-021]]'s
  constraint 8 requires.
- An emitted file SHALL carry the owning-document tag, which the source
  fenced block does not: the emitted file is a feature source, and a
  feature source names its owner.
- Parsing an emitted file SHALL yield records equal to those extracted
  from the source but for their source location and their source form
  (AC-GHK-009.2).

**Anchors:** `cli/cmd`; `cli/core/gherkin` from GHK, which owns the
rendering as it owns the parsing.

### Local change detection in [[container-spec-cli]] (modified) — CHG

**Intent:** Content hashes make "which criteria changed" a comparison of
records rather than a diff of markdown, so a criterion that moved is not
reported and one whose wording changed is. [[frd-gherkin-adoption]]
declares this the lowest-priority requirement in the document, and it is
the last delta for that reason.

**Behavior:**

- `spec criterion --changed` SHALL compare the current corpus against the
  persisted index (AC-GHK-008.1).
- `spec criterion --changed` SHALL list every criterion whose content
  hash differs, classifying each as changed, added or removed by whether
  its identifier is present in both, only now, or only before
  (AC-GHK-008.3).
- A criterion whose content is unchanged and whose position moved SHALL
  NOT be listed (AC-GHK-008.2).
- WHERE no persisted index exists, the command SHALL say so rather than
  reporting every criterion as added.

**Anchors:** `cli/cmd`; `cli/core/corpus` (`Open`, `Build`) for the two
sides of the comparison.

### Relied-upon elements (unchanged)

- [[container-spec-cli#Command Surface]] — the flag parsing, output
  rendering and exit-status conventions every new subcommand follows.
- [[module-document-core#Spans]] — every position this feature reports
  is computed here, including the codepoint-counted column that matches
  Gherkin's own column semantics.
- [[module-document-core#Frontmatter]] — a containing document's slug,
  which is what a fenced block's ownership is derived from.
- [[module-document-core#Findings]] — the report envelope, the three
  severities, and the ordering that makes output byte-identical.
- [[module-document-core#Schema Registry]] — unchanged; a feature record
  is not frontmatter and is validated by no union branch.
- [[module-document-core#Numbering]] — unchanged; feature identifiers
  are not derived by a next-number query.
- [[contract-specs-configuration]] — the profile registration surface
  XPT widens the internal half of.

## Contracts

- [[contract-specs-configuration]] (revised, drafted in place) — the
  contract's description of the internal registration surface
  enumerates what a profile supplies and does not mention a tag
  vocabulary. The revision is additive and to the **internal** half
  only, which that contract's own change policy exempts from versioning
  WHILE both modules are `visibility: internal`; the public `specs.toml`
  surface is untouched, so no `version` bump is required and no adopting
  repository's file changes meaning.
- No new contract document. The index gains a `features` collection, and
  promotion of the index to a contract still waits on a second,
  independently-built consumer (`agent-plugins-h6c`). The generated JSON
  Schema remains its interim machine-readable specification. This is a
  scope decision, not an open question.

## Tracker Reconciliation

- `agent-plugins-dc6` — **subsumed by MODEL**: "compile
  `schemas/feature.strata` and regenerate derived artifacts." The file
  was authored on this branch while the strata CLI was unavailable and
  has never been transpiled. MODEL's first act is that compile, and the
  three constructs the bead names as unverified — the escaped-dot and
  brace-quantifier regex literals, the `@discriminator`/`@tag` union,
  and cross-file resolution of `Span` and `Slug` — are what MODEL has to
  prove. Grooming blocks the bead on the MODEL work order.
- `agent-plugins-3wd` — **subsumed by DISC**: "AC index: extract
  acceptance criteria into the corpus index." GHK extracts and DISC
  records them in `SpecIndex.features`, which is the whole of what the
  bead asks for.
- `agent-plugins-133` — **subsumed by VTAG**: "AC structural conformance
  checks: identifier well-formedness, uniqueness, sequence." VTAG covers
  well-formedness and uniqueness. Sequence — that a feature's criterion
  numbers run without gaps — is **not** covered: no criterion in
  [[frd-gherkin-adoption]] requires it, and this feature does not add
  it. The bead is re-scoped to the sequence check rather than closed.
- `agent-plugins-d1g` — **subsumed by EMIT**: "emit conforming
  `.feature` files from extracted Gherkin blocks."
- `agent-plugins-uoe` — **prerequisite**: PR #14 regressed the
  warm-index path roughly 1.8x by emitting `UnmarshalJSON` on every
  generated model. MODEL adds `Feature` and `Criterion` to that same
  generator and DISC puts them on the read path, so landing this feature
  first would compound a known regression underneath AC-GHK-007.6's
  answer-from-the-index requirement and make the cause harder to
  isolate.
- `agent-plugins-a7u` — **out-of-scope** (sequenced follow-on):
  "migrate acceptance criteria in the three FRDs to Gherkin." Nothing
  here reads a prose criterion ([[adr-021]] constraint 9), so until that
  bead lands, three of this corpus's four FRDs contribute no feature
  records and a corpus-wide query under-reports. That is a translation
  job, not a blocker on this feature, and it is named here so no later
  agent re-derives a compatibility reader. The boundary: no delta here
  rewrites a criterion in an existing FRD.
- `agent-plugins-8rr` — **out-of-scope** (coordination): Strata regex
  literals cannot contain a forward slash, even escaped. Neither pattern
  in `feature.strata` contains one, which is why the file could be
  authored at all; MODEL must not introduce one.
- `agent-plugins-a7m` — **out-of-scope** (coordination):
  `generate-schemas.sh` wipes `cli/core/specmodel`, so a test placed
  beside a generated model is destroyed on the next run. MODEL
  regenerates that directory; tests for the feature model therefore live
  in the consuming packages, as they already do.
- `agent-plugins-006` — **out-of-scope** (coordination): the schema is
  open (`allowAdditionalProperties: true`) pending correct union schema
  emission. `FeatureSource` is a discriminated union and will be
  affected by that fix; closing the schema later must not drop the
  union's tags.
- `agent-plugins-9jo` — **out-of-scope**: the generated models embed
  this framework's `DocumentType` vocabulary, coupling the
  framework-agnostic half to it. XPT widens the extension point without
  adding vocabulary to the host and extends the boundary test to the new
  registered values; it neither worsens nor fixes 9jo.
- `agent-plugins-h6c` — **out-of-scope** (deferred follow-up): "ADR:
  index schema and versioning." The `features` collection is additive
  under 1.0, which is what makes shipping before that ADR safe.
- `agent-plugins-krb` — **out-of-scope** (coordination):
  `regenerate-cli-snapshots.sh` needs a gate so it cannot silently
  regress. VTAG, VCON, QRY, EMIT and CHG all change observable output,
  so a silently-regressed regeneration script would let a stale
  expectation pass.
- `agent-plugins-bem` — **out-of-scope** (coordination): relational
  reference checks read documents off disk on the warm single-document
  path. GREF adds reference sources without changing that path's shape.
- `agent-plugins-gz3` — **out-of-scope**: index entries carry only the
  `DocumentFrontmatter` envelope. DISC derives a fenced block's owning
  slug from the parsed frontmatter the walk already holds, not from the
  typed envelope, so this feature neither depends on gz3 nor resolves
  it.
- `agent-plugins-n2k` — **out-of-scope** (this document): tracks
  authoring this blueprint.

## Sequencing Constraints

- `delivery: staged`. Three stages, each a reviewable increment:
  - **Stage 1 — records in the index:** MODEL, XPT, MDS, GHK, DISC,
    PREG. `spec index` emits feature and criterion records over this
    repository's own corpus; no new check runs, no finding is raised,
    and no exit status changes. The boundary is deliberate: it puts
    extraction and ownership — the rules that produce wrong records —
    in front of a reviewer before any of them can fail a build.
  - **Stage 2 — checks and queries:** VTAG, VCON, GREF, QRY. The checks
    report, the queries answer, and exit status becomes sensitive to
    Gherkin for the first time.
  - **Stage 3 — publication:** EMIT and CHG, the two deltas no other
    delta depends on.
- The MODEL delta SHALL land before every other delta: all of them
  consume types it generates.
- The XPT delta SHALL land before PREG, which populates the widened
  interface, and before VTAG, which reads it.
- The MDS delta SHALL land before GHK, which parses what it extracts,
  and before GREF, which reads a fence's info string.
- The GHK delta SHALL land before DISC, which records what it produces.
- The DISC delta SHALL land before VTAG, VCON, QRY, EMIT and CHG, all of
  which read the recorded features.
- The PREG delta SHALL land before VTAG and VCON, whose every check is
  inert without a registered vocabulary.
- The VTAG and VCON deltas SHALL land before or with QRY, which declares
  their rule identifiers to the run report.
- **Lockstep rules spanning delta blocks, and how they are satisfied:**
  - [[module-document-core]] requires the schema version and the emitted
    index to change together when the index model changes shape. Both
    are in MODEL, and the version is deliberately *not* bumped: the
    `features` collection is an addition under the additive-at-1.0 rule
    [[container-spec-cli]] states, so the satisfying act is the recorded
    no-bump.
  - [[module-document-core]] requires a new parsing capability to be
    built on the Spans delimiter scanner. MDS takes its fence
    recognition from `goldmark` and computes every position through
    Spans, which is the same narrowing References already carries and
    which folds into the module blueprint at retirement.
  - [[container-spec-cli]] requires the snapshot regeneration script to
    be run and every diff reviewed when observable output changes. VTAG,
    VCON, QRY, EMIT and CHG each change observable output, and each
    lands with its regenerated snapshots in the same change rather than
    a shared regeneration pass at the end of a stage.
  - [[module-spec-profile]] requires a registered pattern to change in
    the same change as the standard that defines it. PREG's patterns
    mirror `gherkin-guide.md` and the generated `feature.strata`
    patterns, and PREG lands with the drift test that holds them
    together.
- No backward-compatibility window applies. [[module-document-core]] and
  [[module-spec-profile]] are both `visibility: internal` and evolve in
  lockstep with their only consumer, and the widened extension point is
  the internal half of [[contract-specs-configuration]], which that
  contract's change policy exempts while both modules are internal.

## AC Coverage

Legend: MODEL = feature, criterion and finding models in Strata; XPT =
the widened `Profile` extension point in [[module-document-core]]; MDS =
Markdown Structure (fenced blocks); GHK = the Gherkin component; DISC =
Corpus (discovery and records); PREG = tag registrations in
[[module-spec-profile]]; VTAG = identity checks; VCON = content and
diagnostic checks; GREF = References (Gherkin sources); QRY = Command
Surface (queries); EMIT = Command Surface (emission); CHG = Command
Surface (change detection).

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-GHK-001.1 | **MDS (proves)**, DISC | unit |
| AC-GHK-001.2 | **DISC (proves)**, GHK | unit |
| AC-GHK-001.3 | **GHK (proves)**, DISC | unit |
| AC-GHK-001.4 | **GHK (proves)** | unit |
| AC-GHK-001.5 | **GHK (proves)**, MDS | unit |
| AC-GHK-001.6 | **DISC (proves)** | unit ‡ |
| AC-GHK-001.7 | **DISC (proves)** | unit |
| AC-GHK-002.1 | **PREG (proves)**, XPT, GHK | unit |
| AC-GHK-002.2 | **VTAG (proves)**, PREG | unit |
| AC-GHK-002.3 | **VTAG (proves)**, PREG | unit |
| AC-GHK-002.4 | **VTAG (proves)**, DISC | unit |
| AC-GHK-002.5 | **VTAG (proves)** | unit |
| AC-GHK-002.6 | **GHK (proves)** | unit |
| AC-GHK-002.7 | **VCON (proves)**, GHK | unit |
| AC-GHK-002.8 | **VTAG (proves)**, XPT | unit |
| AC-GHK-002.9 | **GHK (proves)** | unit |
| AC-GHK-002.10 | **VTAG (proves)**, DISC | unit |
| AC-GHK-002.11 | **VTAG (proves)**, DISC | unit |
| AC-GHK-003.1 | **QRY (proves)**, DISC | integration |
| AC-GHK-003.2 | **GHK (proves)**, MODEL | unit |
| AC-GHK-003.3 | **QRY (proves)**, DISC | integration |
| AC-GHK-003.4 | **VTAG (proves)**, DISC | unit |
| AC-GHK-003.5 | **VTAG (proves)** | unit |
| AC-GHK-003.6 | **VTAG (proves)**, PREG, XPT | unit |
| AC-GHK-004.1 | **GREF (proves)**, MDS | unit |
| AC-GHK-004.2 | **GREF (proves)**, MDS | unit |
| AC-GHK-004.3 | **GREF (proves)** | case + unit |
| AC-GHK-004.4 | **GREF (proves)** | unit |
| AC-GHK-005.1 | **GHK (proves)**, MODEL, DISC | unit |
| AC-GHK-005.2 | **GHK (proves)**, MODEL | unit |
| AC-GHK-005.3 | **MODEL (proves)**, GHK | unit |
| AC-GHK-005.4 | **MODEL (proves)**, GHK | unit |
| AC-GHK-005.5 | **GHK (proves)**, MODEL | unit |
| AC-GHK-005.6 | **GHK (proves)** | unit |
| AC-GHK-005.7 | **GHK (proves)** | unit |
| AC-GHK-005.8 | **DISC (proves)**, MODEL | integration |
| AC-GHK-006.1 | **VCON (proves)**, GHK, MDS | unit |
| AC-GHK-006.2 | **VCON (proves)**, VTAG | case + unit |
| AC-GHK-006.3 | **VCON (proves)**, MODEL | case |
| AC-GHK-006.4 | **VCON (proves)** | unit |
| AC-GHK-006.5 | **VCON (proves)**, PREG | unit |
| AC-GHK-006.6 | **VCON (proves)** | unit |
| AC-GHK-006.7 | **VCON (proves)**, DISC | unit |
| AC-GHK-007.1 | **QRY (proves)**, DISC | integration |
| AC-GHK-007.2 | **QRY (proves)**, DISC | integration |
| AC-GHK-007.3 | **QRY (proves)** | integration |
| AC-GHK-007.4 | **QRY (proves)** | case |
| AC-GHK-007.5 | **QRY (proves)**, MODEL | case |
| AC-GHK-007.6 | **QRY (proves)**, DISC | unit † |
| AC-GHK-008.1 | **CHG (proves)**, GHK | integration |
| AC-GHK-008.2 | **CHG (proves)**, GHK | integration |
| AC-GHK-008.3 | **CHG (proves)** | integration |
| AC-GHK-009.1 | **EMIT (proves)**, GHK | integration |
| AC-GHK-009.2 | **EMIT (proves)**, GHK | unit |
| AC-GHK-009.4 | **GHK (proves)**, MODEL | unit |

‡ AC-GHK-001.6 was declared `manual` at design time, on the reading that
it states an authoring convention rather than a behavior of the CLI: a
feature source sits beside the documents of its feature, and nothing in
the tool places it there. DISC was named as its proving delta because
discovery over the spec root is what makes the convention sufficient.
The level was corrected to `unit` at retirement — that observable half
is exactly what
`TestAFeatureSourceIsFoundWhereverBeneathTheSpecRootItSits` asserts, so
the criterion is verified automatically and nothing about it is left to
a human to check.

† AC-GHK-007.6 is verified by a Go test rather than a snapshot case,
because the state it is about — an index served warm, with nothing
re-walked — cannot be staged from a case's input corpus, and a snapshot
case is pure data by [[testing-standards]]'s harness-independence rule.
It joins AC-SPECREF-003.4, which is warm-path behavior for the same
reason.

## Key Decisions

- [[adr-021]] — Gherkin is the notation for features and acceptance
  criteria, the official Go parser is the implementation, and the
  fourteen binding constraints every delta here realizes.
- [[adr-016]] — Strata is the single source for generated models, which
  is why MODEL changes `.strata` sources rather than Go structs.
- [[adr-018]] — references are one top-level edge collection with both
  directions derived. Features deliberately do **not** follow it:
  criteria nest, because a criterion is a node belonging to exactly one
  feature rather than an edge describing neither endpoint.
- [[adr-015]] — Go, and the module composition model the MDS/GHK split
  follows.

No new ADR. The one boundary decision this blueprint makes — Gherkin
parsing in the host, tag vocabulary in the profile — is an application
of the one-way dependency [[module-document-core]] already states, and
its rationale folds into that blueprint's Gherkin component at
retirement rather than dying with this document.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — Components gains **Gherkin**;
        Markdown Structure carries a **Built — fenced blocks only**
        paragraph, with the heading tree and comment blocks still named
        as outstanding; Interface records the widened `Profile` as a
        fifth member, `FeatureTags()`, and the two new registration
        defects `ValidateRegistration` reports; Corpus's index
        vocabulary gains `Feature` and `Criterion`; References records
        the narrowed code mask; Findings records the new detail
        variants; Key Decisions' [[adr-021]] line names the parser the
        Gherkin component is built on rather than a follow-on.
        Reconciled at retirement: the Components prose still said
        nothing in Corpus called `cli/core/markdown` or
        `cli/core/gherkin`, which DISC made false.
  - [x] [[module-spec-profile]] — Interface and Profile Registration
        record the tag registration and the structural guard holding the
        registered patterns against the generated ones; Key Decisions'
        [[adr-021]] line names what Profile Registration registers.
  - [x] [[container-spec-cli]] — Command Surface records `feature` and
        `criterion`, the `--emit`/`--out`, `--code`, `--of`, `--tag`,
        `--changed` and `--json` flags and their mutual exclusions;
        Change Guidance's verification gates cover them under the
        general command-surface rule. Reconciled at retirement: Key
        Decisions' [[adr-021]] line still read "acceptance-criteria
        extraction is a follow-on feature, so no component here parses
        acceptance criteria yet".
- [x] [[contract-specs-configuration]] — the internal registration
      surface enumerates the tag vocabulary: the two identifier
      patterns, the owning-document tag prefix, the shared captures and
      the qualifier tags with their meanings. `version` stays `"1.0"`,
      the recorded no-bump being the satisfying act — the revision is
      additive and to the internal half only, which that contract's own
      change policy exempts WHILE both modules are `visibility:
      internal`.
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per `ac-traceability-standards.md`. All
      55 carry an `AC:` citation or a case manifest `cites` entry, and
      no citation names a criterion [[frd-gherkin-adoption]] does not
      define. Two — AC-GHK-002.8 and AC-GHK-003.5, both negative
      criteria — had a test each carrying the identifier in prose rather
      than behind the `AC:` marker, and were given proper citations at
      retirement.
- [x] Every issue listed as subsumed in Tracker Reconciliation is closed
      or re-scoped: `agent-plugins-dc6` (MODEL), `agent-plugins-3wd`
      (DISC) and `agent-plugins-d1g` (EMIT) closed; `agent-plugins-133`
      re-scoped to the criterion-number sequence check alone, which no
      criterion in [[frd-gherkin-adoption]] requires and this feature
      does not deliver.
- [x] This blueprint's status → retired.

## Open Questions

None.

## References

- `c4-guide.md` — the container, module and component definitions the
  delta blocks are written against
- [[frd-gherkin-adoption]] — the requirements these deltas realize
- [[adr-021]] — the decision, its measurements, and its binding
  constraints
- [[module-document-core]] — the host every parsing delta lands in
- [[module-spec-profile]] — the plugin PREG registers through
- [[container-spec-cli]] — the command surface QRY, EMIT and CHG extend
- [[frd-spec-validation-cli]] — the index, the finding record and the
  warm index these deltas build on
- [[frd-spec-reference-integrity]] — the reference extraction GREF
  narrows the code mask of
- [[gherkin-guide]] — the authoring rules
  PREG's registered patterns mirror
- `schemas/feature.strata` — the record contract MODEL compiles
