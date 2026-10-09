---
title: "C4 Diagram Conformance"
description: How the Spec CLI learns to read the Mermaid C4 diagrams in
  architecture blueprints — a parser and an element resolver in Document
  Core, fed by an architecture vocabulary the profile registers — and
  reports every unreadable statement, unknown node, two-way arrow and
  malformed tag as a c4 finding
slug: feature-spec-c4-diagrams
type: feature-blueprint
status: retired
delivery: staged
frd: frd-spec-c4-diagrams
plan: specs-c4s2
---

# C4 Diagram Conformance

## Overview

**FRD:** [[frd-spec-c4-diagrams]]

The feature is a parser, a resolver and a family of checks, all in
[[module-document-core]], with the vocabulary they need arriving from
[[module-spec-profile]] per [[adr-030]]. A new **C4 Diagrams** component
parses each C4 diagram in an architecture blueprint into statements with
spans and knows nothing about the corpus. A new **C4 Elements** component
builds the corpus's element set from blueprints and roster record
blocks, computes each drawing document's scope, resolves a node to the
elements its label names, and resolves each relationship statement to a
drawing of one relationship. Corpus runs the `c4/*` rules over both, and
Findings carries their identifiers. Nothing here changes the index's
shape or adds a command: the rules are reached through `specs check` and
`--only` as every other rule is. [[frd-spec-c4]] builds its model on the
ELEM and DRAW deltas below.

Delta blocks in dependency order: PARSE and FIND, which depend on
nothing; then XPT, PREG, ELEM, DRAW, CHKN and CHKR in that order; DOC
alongside CHKR.

## Element Deltas

### Profile extension point in [[module-document-core]] (modified — C4)

**Intent:** Document Core must find elements without naming a document
type, a property or a heading, so the `Profile` interface gains one
member carrying that vocabulary, the way `FeatureTags()` carries the
Gherkin vocabulary ([[adr-030]]). The host learns where to look; the
profile says what is there.

**Behavior:**
- The `Profile` interface SHALL let a profile register an architecture
  vocabulary consisting of: the document types that declare an element,
  each with its C4 level (`system`, `container`, `module` or
  `component`); the frontmatter property naming an element's entity
  name; the property by which a container names its system; the property
  by which a container or module lists the modules composed into it; the
  property listing an element's consumed contracts; the contract document
  type and the property in which a contract names its provider; and, per
  document type that holds children, the heading text of its roster
  section and the record-block tag that declares a child there.
- The module SHALL NOT name any of those values. A profile that
  registers no architecture vocabulary SHALL yield a corpus with no
  elements, and SHALL NOT cause an error.
- Each registered property SHALL apply to a fixed set of registered
  types: the entity-name and consumed-contracts properties to every
  registered type; the system property to types at the `container`
  level; the composed-modules property to types at the `container` and
  `module` levels; the provider property to the contract type.
- WHEN a profile is activated, `ValidateRegistration` SHALL report as a
  defect: a registered type the profile does not define; a registered
  property that the union schema does not declare on a type it applies
  to; a level outside the four above; and a roster record-block tag that
  is not among the profile's registered record-block tags. Each defect
  SHALL name the offending key, as the existing defects do.

**Anchors:** `cli/core/corpus/profile.go` (`Profile`, `FeatureTagRules`
as the precedent); `cli/core/corpus/registration.go`
(`ValidateRegistration`, `RegistrationDefect`); the source-scan test
`TestTheExtensionPointNamesNoRegisteredVocabulary`, which stays green.

### Profile Registration in [[module-spec-profile]] (modified — C4)

**Intent:** This framework's architecture vocabulary is already stated
by `schemas/frontmatter.strata` and by [[c4-standards]]; the profile
transcribes it into the new member, and a drift test holds the two
together, as the reference registration's test does.

**Behavior:**
- The profile SHALL register `system-blueprint`, `container-blueprint`,
  `module-blueprint` and `component-blueprint` at levels `system`,
  `container`, `module` and `component` respectively.
- The profile SHALL register `entity_name` as the entity-name property,
  `system` as the system property, `modules` as the composed-modules
  property, `contracts_consumed` as the consumed-contracts property,
  `contract` as the contract type and `provider` as its provider
  property.
- The profile SHALL register `Components` as the roster heading and
  `component` as the roster record-block tag for `container-blueprint`
  and for `module-blueprint`, and no roster for the other two types.
- A test SHALL assert that the union schema declares every registered
  property on every type it applies to, so a renamed property fails the
  suite rather than silently resolving nothing.

**Anchors:** `cli/profile/profile.go`, `cli/profile/registration.go`;
`TestEveryCitationTypedPropertyInTheVocabularyIsRegistered` as the
drift-test precedent.

### C4 Diagrams in [[module-document-core]] (added)

**Intent:** [[adr-029]] gives the CLI its own parser, because the one Go
library measured fails on this corpus's own diagrams. It reads exactly
the statement subset the FRD's terminology lists and turns everything
else into a parse error, so the subset grows deliberately. It is pure
notation: a block of text in, statements with spans out. It neither
reads the corpus nor resolves a node.

**New components** (declared in module vocabulary):
- **C4 Diagrams** (Go package) — Parses Mermaid C4 diagram blocks into
  located statements. Reads the C4Context, C4Container and C4Component
  diagram types; reports every other Mermaid diagram type by name
  without parsing it; and yields, per statement, its kind, its macro,
  its arguments and its span, with relationship tags split into reserved
  and unreserved.

  ```component
  entity_name: C4 Diagrams
  slug: component-c4-diagrams
  description: Mermaid C4 diagram parsing into located statements, with relationship tags read
  ```

**Behavior:**
- The component SHALL take a fenced `mermaid` block's content and its
  position, and SHALL report every position as a `Span` computed by the
  Spans component, in the coordinates of the whole document.
- The component SHALL identify a diagram's type from its first statement
  and SHALL parse only `C4Context`, `C4Container` and `C4Component`
  diagrams. For any other first statement it SHALL return the type name
  and no statements.
- The component SHALL read each line as one statement. A blank line, and
  a line whose first non-blank characters are `%%`, SHALL be skipped.
- The component SHALL read the macros the FRD's terminology table lists,
  every `Db` and `Queue` form of `System`, `Container` and `Component`,
  and every `_Ext` form, and SHALL classify each statement as an element
  node, an external node, a boundary, a relationship or an ignored
  statement.
- A boundary statement SHALL end with `{`, and a line holding only `}`
  SHALL be read as closing the innermost open boundary.
- A quoted argument SHALL run from its opening double quote to the next
  double quote, so commas and parentheses inside it are text. There are
  no escape sequences. An unquoted argument SHALL be read as its text
  with surrounding whitespace removed.
- An argument written `$<name>=<value>` SHALL be read as the named
  parameter `<name>`, and SHALL occupy the position it is written at, as
  Mermaid's own C4 parser assigns it: a positional argument after it
  fills the parameter at its own position, never an earlier one.
- The component SHALL map positional arguments onto Mermaid's
  parameter order for each macro: for `Rel` and its variants, source
  alias, target alias, label, technology, description, sprite, tags and
  link; for a `Container` or `Component` node, alias, label, technology,
  description, sprite, tags and link; for a `System` or `Person` node,
  alias, label, description, sprite, tags and link.
- WHEN a statement is not one of the read macros, has unbalanced quotes
  or parentheses, or has fewer arguments than its macro requires, the
  component SHALL return a parse error carrying that statement's span,
  and SHALL continue with the next line.
- The component SHALL read a relationship's tags from its tags
  argument, positional or named, and SHALL split them on `+` only. The
  tags `sync` and `async`, and any tag beginning `contract:`, SHALL be
  returned as reserved, with the slug after `contract:` separated out;
  every other tag SHALL be returned as unreserved, in the order written.
- WHEN a tags argument contains a comma, the component SHALL record that
  it does, and SHALL still split on `+` only, so `async,audited` is one
  unreserved tag.
- The component SHALL normalise every relationship to dependency
  direction: `Rel_Back(a, b)` SHALL be returned with source `b` and target
  `a`; every directional variant (`Rel_U`, `Rel_Up`, `Rel_D`, `Rel_Down`,
  `Rel_L`, `Rel_Left`, `Rel_R`, `Rel_Right`) SHALL be returned as `Rel`
  with the macro as written kept on the statement; `BiRel` SHALL be
  returned as a relationship flagged bidirectional.

**Anchors:** a new package under `cli/core`; `cli/core/span` (`Of`) for
every position; `cli/core/markdown` (`Fenced`) is the caller's way to
the blocks and is not called from here.

### Findings in [[module-document-core]] (modified — c4 rules)

**Intent:** The rules are a published surface: `--only` selects them by
identifier, and the report accounts for each one by identifier whether
it ran or not. Each rule's detail is a variant of the generated
`FindingDetail` union, so a JSON consumer reads the alias, label or tag
the finding concerns without parsing a message.

**Behavior:**
- The module SHALL declare these rule identifiers: `c4/parse`,
  `c4/undeclared-node`, `c4/ambiguous-node`, `c4/bidirectional`,
  `c4/duplicate-edge`, `c4/drawing-conflict`, `c4/interaction-conflict`,
  `c4/contract-unresolved`, `c4/contract-mismatch` and
  `c4/tag-separator`.
- `c4/parse` SHALL be reportable at severity error and at severity
  unverified under the same identifier, as `reference/unresolved-anchor`
  already is.
- `c4/ambiguous-node`'s detail SHALL carry the slugs of every element
  the label matched; `c4/drawing-conflict`'s SHALL carry the path and
  span of the other drawing; `c4/parse`'s unverified detail SHALL carry
  the diagram type name.
- Each `c4/*` finding SHALL be narrowed by a single-document check to
  the file it sits in, like every other rule.
- This delta SHALL NOT add any `c4/*` identifier to `checkRules` in
  `cli/cmd/check.go`. A rule listed there is reported as passed whenever
  it is requested, so each identifier is added by the check delta that
  runs it, and never before.

**Anchors:** `cli/core/finding/finding.go` (the rule constants); `schemas/finding.strata` (`FindingDetail`),
regenerated with `scripts/generate-schemas.sh`, which needs the `strata`
compiler.

### C4 Elements in [[module-document-core]] (added — node resolution)

**Intent:** A node is valid only if it names a declared element, and
which elements it may name depends on where it is drawn. This component
owns that judgement once, so the checks here and the model in
[[frd-spec-c4]] cannot disagree about what a node is.

**New components** (declared in module vocabulary):
- **C4 Elements** (Go package) — The corpus's architectural elements,
  each drawing document's scope, and the resolution of diagram nodes and
  relationship statements to elements. Reads the index and the
  architecture vocabulary the profile registers; names no document type
  of its own.

  ```component
  entity_name: C4 Elements
  slug: component-c4-elements
  description: The corpus's architectural elements, each drawing document's scope, and node and relationship resolution
  ```

**Behavior:**
- The component SHALL declare an element for every corpus member whose
  type the profile registers as declaring one, taking the member's slug
  as the element's slug and its entity-name property as its name.
- The component SHALL declare an element for every record block whose
  tag is the registered roster tag of the document's type and which is
  attached to the registered roster heading or to a heading nested
  inside its section. The heading SHALL be matched on the identifier
  Markdown Structure derives, so its letter case does not decide a
  match.
- The component SHALL decode a roster record block's content through the
  record-block decoding function the structure feature provides. A block
  whose content does not decode as a mapping carrying a slug and an
  entity name SHALL declare no element; reporting the block is the
  structure conformance checks' job, not this component's.
- A record block in a document whose type declares no roster, such as a
  feature blueprint's delta block, SHALL declare no element.
- WHEN a blueprint and one or more record blocks carry the same slug, or
  two record blocks do, the component SHALL hold one element for that
  slug carrying every declaration, ordered by path and then by span.
- The scope of a drawing document SHALL be: its own element; the
  elements its roster declares, whether by record block or by a
  `[[slug]]` reference resolving to a member of a component level type;
  the elements its composed-modules property names; and every element
  at the `system` or `container` level in the corpus.
- A node SHALL resolve to every element in its drawing document's scope
  whose name equals the node's label exactly, compared as written.
- An external node SHALL NOT be resolved. A `Db` or `Queue` node whose
  label matches no element in scope SHALL resolve to a passive store
  rather than to nothing.
- WHEN a C4 diagram sits in a member whose type the profile does not
  register as declaring an element, the component SHALL return no node
  and no drawing for it.

**Anchors:** a new package under `cli/core`; `cli/core/corpus`
(`Corpus.Entry`, the index's `RecordBlock` and `Heading` records);
`cli/core/markdown` (`Identifier`, `Fenced`).

### C4 Elements in [[module-document-core]] (modified — drawings)

**Intent:** The drawing-conflict rule needs to know when two
relationship statements in different documents are the same
relationship, and the model needs the same grouping. Both read it from
here.

**Behavior:**
- The component SHALL resolve a relationship statement to a drawing
  WHEN both of its endpoint aliases name nodes in the same diagram that
  each resolve to exactly one element. A drawing SHALL carry the source
  and target elements in dependency direction, the label, the
  technology and description arguments, the tags as parsed, the
  diagram's view, the aliases as written, and the statement's path and
  span.
- WHEN an endpoint is an external node, a passive store, matches no
  element in scope, or matches more than one, the component SHALL
  return the statement unresolved with that one reason, checking in that
  order and reporting the source endpoint before the target.
- WHEN an alias names no node in its diagram, the component SHALL
  return the statement unresolved as `undeclared_node`.
- Two drawings SHALL be drawings of one relationship WHEN their source
  elements, target elements and labels are equal.
- A `BiRel` statement SHALL NOT resolve to a drawing.
- The component SHALL read diagrams from each architecture blueprint's
  source as it stands at the time of the run. The index SHALL NOT record
  diagrams, so this feature does not change the index's shape or its
  `schema_version`.

**Anchors:** the C4 Elements package the previous block adds; the C4
Diagrams package for statements.

### Corpus in [[module-document-core]] (modified — parse and node checks)

**Intent:** The first half of the rule family: everything that can be
judged from one statement and the scope it was drawn in.

**Behavior:**
- WHEN a C4 diagram in an architecture blueprint holds a statement the
  parser cannot read, the check SHALL report `c4/parse` at severity
  error at that statement's span.
- WHEN a `mermaid` block in an architecture blueprint is not a C4
  diagram the parser reads, the check SHALL report `c4/parse` at
  severity unverified at the block's span, naming its diagram type.
- WHEN a non-external node resolves to no element and is not a passive
  store, the check SHALL report `c4/undeclared-node` at severity error
  at the node's span.
- WHEN a node resolves to more than one element, the check SHALL report
  `c4/ambiguous-node` at severity error at the node's span, naming every
  element it matched.
- The checks SHALL NOT examine a `mermaid` block in a member whose type
  is not registered as declaring an element.
- WHERE the profile registers no architecture vocabulary, the checks
  SHALL report nothing.
- This delta SHALL add `c4/parse`, `c4/undeclared-node` and
  `c4/ambiguous-node` to `checkRules` in `cli/cmd/check.go`, so `--only`
  can name them and a clean run reports them as passed.

**Anchors:** `cli/core/corpus/check.go` (`Check`, `CheckFor`), beside
the reference and Gherkin checks; `checkRules` in `cli/cmd/check.go`;
snapshot cases under `cli/testdata/cases/` following
`check-only-anchor-rule`. Adding to `checkRules` changes the pass list
every existing `check` snapshot records.

### Corpus in [[module-document-core]] (modified — relationship checks)

**Intent:** The second half: every rule about an arrow and the tags on
it, including the one rule that compares drawings across documents.

**Behavior:**
- WHEN a C4 diagram holds a `BiRel` statement, the check SHALL report
  `c4/bidirectional` at severity error at its span.
- WHEN two relationship statements in one diagram have the same source
  alias and target alias after `Rel_Back` normalisation, the check SHALL
  report `c4/duplicate-edge` at severity error at the span of the later
  one. A statement from `a` to `b` and one from `b` to `a` SHALL NOT be
  reported.
- WHEN two drawings of one relationship differ in technology,
  description or tags, the check SHALL report `c4/drawing-conflict` at
  severity warning at the drawing that sorts second by path and then by
  span, naming the first. Each further disagreeing drawing SHALL be
  reported the same way against the first.
- WHEN a relationship statement carries both `sync` and `async`, the
  check SHALL report `c4/interaction-conflict` at severity error at its
  span.
- WHEN a `contract:` tag names a slug no corpus member of the registered
  contract type declares, the check SHALL report `c4/contract-unresolved`
  at severity error at the statement's span.
- WHEN a `contract:` tag names a contract the statement's source element
  does not list in its consumed-contracts property, or whose provider is
  not the target element, the check SHALL report `c4/contract-mismatch`
  at severity warning at the statement's span.
- WHERE either endpoint of a statement does not resolve to exactly one
  element, the check SHALL NOT report `c4/contract-mismatch` for it,
  since the node finding already covers the statement.
- WHEN a tags argument contains a comma, the check SHALL report
  `c4/tag-separator` at severity warning at the statement's span.
- WHEN `specs check` names one document, `c4/drawing-conflict` SHALL
  still compare that document's drawings against every architecture
  blueprint's, and SHALL then be narrowed to the named file like every
  other `c4/*` rule.
- This delta SHALL add the seven remaining `c4/*` identifiers to
  `checkRules` in `cli/cmd/check.go`.

**Anchors:** `cli/core/corpus/check.go`; `checkRules` in
`cli/cmd/check.go`; snapshot cases under `cli/testdata/cases/`.

### Documentation (modified — diagram rules)

**Intent:** The checks enforce rules [[c4-standards]] does not yet
state, and the templates do not show the `technology` property the
checks' sibling feature reads. Authors cannot be held to rules nobody
wrote down, so the documents change with the checks.

**Behavior:**
- [[c4-standards]] SHALL state that a relationship statement declares
  an edge while a node remains a view of its declaration, per
  [[adr-029]].
- [[c4-standards]] SHALL list the statement subset the parser reads,
  the reserved tags `sync`, `async` and `contract:<slug>`, that tags are
  joined with `+`, and that `Rel_Back(a, b)` declares `b` depending on
  `a`.
- The four architecture blueprint templates SHALL show the optional
  `technology` property in their frontmatter and in their component
  record block examples.
- [[spec-cli-reference]] SHALL list the `c4/*` rules, each with its
  severity and what it reports.

**Anchors:** `specs/standards/c4-standards.md`;
`plugin/skills/spec/references/*-blueprint-template.md` for the system,
container, module and component templates;
`specs/guides/spec-cli-reference.md`. `scripts/sync-shared-references.sh`
SHALL be re-run so the vendored copies follow.

### Relied-upon elements (unchanged)

- [[module-document-core#Markdown Structure]] — `Fenced` to locate
  `mermaid` blocks, `Identifier` to match the roster heading.
- [[module-document-core#Spans]] — every position the parser reports.
- The References component of [[module-document-core]] — the resolved
  body references that place a roster's component blueprint link under
  its roster heading.
- The index's per-type frontmatter, which carries `entity_name`,
  `system`, `modules` and `contracts_consumed` on each entry (measured:
  `specs index` on this repository emits `container-spec-cli` with
  `entity_name: Spec CLI` and both `modules`).
- The structure feature's record-block recording and its decoding
  function — not yet built; see Sequencing Constraints.
- `finding.NewReport` and the `--only` machinery in `cli/cmd/check.go`,
  which select and account for any rule `checkRules` declares. The only
  Command Surface change is the two `checkRules` additions the CHKN and
  CHKR deltas make.

## Contracts

None. The rules and their detail variants extend the finding JSON, which
is generated from `schemas/finding.strata` and is not yet a contract
document.

## Tracker Reconciliation

- `specs-f6tl.5` — prerequisite: the record-block tag
  registration point. XPT validates the roster tag against the
  registered tags and ELEM reads roster blocks by one; neither can land
  before this does.
- `specs-f6tl.6` — prerequisite: recording record blocks against
  their heading, and the decoding function ELEM uses. Without it no
  un-promoted component is an element, and every component node on this
  repository's diagrams is undeclared.
- `agent-plugins-7os0` — subsumed by the DOC delta, which amends
  [[c4-standards]] and the four templates exactly as that issue
  describes.
- `agent-plugins-gz3` — prerequisite, already met: the index carries
  per-type frontmatter since the change that closed
  `agent-plugins-23f1`. It blocks nothing; this feature does not close
  it.
- `agent-plugins-4me` — out-of-scope: stale `c4-guide.md` paths under
  `specs/features/`. This feature adds no such path and fixes none.

## Sequencing Constraints

- Delivery is staged. Stage 1 is PARSE and FIND, which depend on
  nothing outside this blueprint, add no rule to `checkRules` and change
  no output. Stage 2 is XPT, PREG, ELEM, DRAW, CHKN, CHKR and DOC.
- Stage 2 SHALL NOT begin until `specs-f6tl.5`,
  `specs-f6tl.6`, and the registration of the `component`
  record-block tag that FEAT-SPECSTR-004 of
  [[frd-spec-structure-conformance]] specifies, have landed. No work
  order for that registration exists yet.
- The XPT delta SHALL land before the PREG delta.
- The XPT, PREG and PARSE deltas SHALL land before the ELEM delta.
- The ELEM delta SHALL land before the DRAW delta.
- The FIND, PARSE and ELEM deltas SHALL land before the CHKN delta.
- The FIND and DRAW deltas SHALL land before the CHKR delta.
- WHEN the CHKN or CHKR delta lands, `specs check` over this repository
  SHALL report no `c4/*` finding, since the profile self-application
  test in `cli/profile` requires zero findings. Any diagram in this
  repository that a rule reports SHALL be corrected in the same change.
  Measured before this blueprint: every node on this repository's four
  C4 diagrams carries an entity name that a blueprint or a roster record
  block declares.
- The DOC delta SHALL land no later than the CHKR delta, so no rule is
  enforced before the standard states it, and SHALL NOT land before
  [[adr-029]] is approved, since it amends [[c4-standards]] on that
  ADR's authority.
- WHEN a delta changes observable output, the snapshot cases under
  `cli/testdata/cases/` SHALL be regenerated with
  `scripts/regenerate-cli-snapshots.sh` and every diff reviewed before
  commit, per [[container-spec-cli]]'s verification gates.
- Each delta SHALL be folded into its owning blueprint in the same work
  order that implements it.

## AC Coverage

Legend: XPT = Profile extension point delta; PREG = Profile Registration
delta; PARSE = C4 Diagrams delta; FIND = Findings delta; ELEM = C4
Elements delta (element set and node resolution); DRAW = C4 Elements
delta (drawings); CHKN = Corpus delta (parse and node checks); CHKR =
Corpus delta (relationship and tag checks); DOC = Documentation delta.

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-C4-DIAG-001.1 | PARSE, FIND, **CHKN (proves)** | integration |
| AC-C4-DIAG-001.2 | **PARSE (proves)**, CHKN | unit |
| AC-C4-DIAG-001.3 | PARSE, FIND, **CHKN (proves)** | integration |
| AC-C4-DIAG-001.4 | ELEM, **CHKN (proves)** | integration |
| AC-C4-DIAG-002.1 | XPT, PREG, ELEM, **CHKN (proves)** | integration |
| AC-C4-DIAG-002.2 | ELEM, **CHKN (proves)** | integration |
| AC-C4-DIAG-002.3 | ELEM, FIND, **CHKN (proves)** | integration |
| AC-C4-DIAG-002.4 | ELEM, FIND, **CHKN (proves)** | integration |
| AC-C4-DIAG-002.5 | PARSE, ELEM, **CHKN (proves)** | integration |
| AC-C4-DIAG-003.1 | PARSE, FIND, **CHKR (proves)** | integration |
| AC-C4-DIAG-003.2 | PARSE, FIND, **CHKR (proves)** | integration |
| AC-C4-DIAG-003.3 | DRAW, FIND, **CHKR (proves)** | integration |
| AC-C4-DIAG-004.1 | **PARSE (proves)** | unit |
| AC-C4-DIAG-004.2 | PARSE, FIND, **CHKR (proves)** | integration |
| AC-C4-DIAG-004.3 | PARSE, FIND, **CHKR (proves)** | integration |
| AC-C4-DIAG-004.4 | XPT, PREG, DRAW, **CHKR (proves)** | integration |
| AC-C4-DIAG-004.5 | PARSE, FIND, **CHKR (proves)** | integration |

## Key Decisions

- [[adr-029]] — relationships are read from blueprint diagrams by a
  parser this CLI owns, with reserved relationship tags.
- [[adr-030]] — the parser and the resolver live in Document Core, and
  the architecture vocabulary arrives as a profile registration.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — C4 Diagrams and C4 Elements added to
        the Components roster and diagram with their record blocks; the
        extension point's seventh member and its defects in Interface;
        Corpus's check list and Findings' rule list updated
  - [x] [[module-spec-profile]] — Profile Registration's member table
        gains the architecture vocabulary
  - [x] [[container-spec-cli]] — the Command Surface's account of `check`
        paragraph names the `c4/*` rules
- [x] [[c4-standards]], the four architecture blueprint templates and
      [[spec-cli-reference]] updated, and
      `scripts/sync-shared-references.sh --check` passes
- [x] [[adr-029]] and [[adr-030]] are referenced from
      [[module-document-core]]'s Key Decisions
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per [[ac-traceability-standards]]
- [x] `scripts/coverage.sh` has been run from the repository root and
      the README diff committed, since `cli/**` changed
- [x] `agent-plugins-7os0` is closed
- [x] This blueprint's status → retired

## Open Questions

None.

## References

- [[c4-standards]] — the diagram conventions these rules enforce, and
  the element definitions the deltas are written against
- [[frd-spec-c4-diagrams]] — the requirements this blueprint realizes
- [[frd-spec-c4]] — the model built on the ELEM and DRAW deltas
- [[container-spec-cli]] — the container whose Change Guidance binds
  every delta
- [[module-document-core]] and [[module-spec-profile]] — the modules
  the deltas change
- [[frd-spec-structure]] and [[frd-spec-structure-conformance]] — record
  block recording and registration, which stage 2 waits on
- [[ac-traceability-standards]] — the citation obligations the coverage
  table feeds
