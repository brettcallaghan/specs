---
title: "C4 Model"
description: How the Spec CLI assembles the corpus's C4 model — an element
  tree and one edge list built in Document Core from the elements and
  drawings the diagram conformance feature resolves — and emits it from a
  new c4 subcommand
slug: feature-spec-c4
type: feature-blueprint
status: retired
delivery: single-pr
frd: frd-spec-c4
plan: specs-c4model
---

# C4 Model

## Overview

**FRD:** [[frd-spec-c4]]

The model is a projection, assembled in [[module-document-core]] by a new
**C4 Model** component from two things Document Core's **C4 Elements**
component already builds: the element set with its declarations, and the drawings
each relationship statement resolves to. C4 Model nests the elements by
the containment edges the corpus declares, merges drawings into
relationships, reads the reserved tags into properties, and narrows the
result to one subtree on request. It returns the generated
`specmodel.C4Model`, whose shape `schemas/c4.strata` already declares;
the Command Surface in [[container-spec-cli]] gains a `c4` subcommand
that renders it. One enum value is added to the schema. The index is not
touched.

Delta blocks in dependency order: MODEL, then TREE, then REL, then TAGS
and ROOT, then CMD; DOC lands with CMD.

## Element Deltas

### `C4ExclusionReason` in `schemas/c4.strata` (modified)

**Intent:** [[frd-spec-c4]] records every relationship statement the
model does not admit, so that its absence is visible. A `BiRel` is such
a statement — [[c4-standards]] forbids it and the diagram checks report
it — and the four reasons the schema names do not describe it. Without a
fifth, a `BiRel` would vanish from the model with no trace.

**Behavior:**
- `C4ExclusionReason` SHALL gain the value `bidirectional`, documented
  as a `BiRel` statement, which states no single dependency.
- `C4Model.schema_version` SHALL stay `0.1`: no model of this shape has
  been emitted, so no consumer has read the enum this extends.

**Anchors:** `schemas/c4.strata`; the generated
`cli/core/schema/spec-model.schema.json` and `cli/core/specmodel`,
regenerated with `scripts/generate-schemas.sh`, which needs the `strata`
compiler.

### C4 Model in [[module-document-core]] (added — element tree)

**Intent:** The tree is the element set nested by the edges the corpus
declares: a container's system property, a host's modules list, and a
child's place in its parent's roster. Nothing is inferred from directory
layout or from a diagram. Living in Document Core keeps the assembly
available to any wrapper, per [[container-spec-cli]]'s Change Guidance
and [[adr-030]].

**New components** (declared in module vocabulary):
- **C4 Model** (Go package) — Assembles the corpus's C4 model from the
  resolved elements and drawings: the element tree, the module
  collection, the relationships and the excluded statements, optionally
  narrowed to one element's subtree. Returns the generated model and
  renders nothing.

  ```component
  entity_name: C4 Model
  slug: component-c4-model
  description: Assembles the element tree, modules, relationships and exclusions of the corpus's C4 model
  ```

**Behavior:**
- The component SHALL place every `system` level element in
  `software_systems`, every `module` level element in `modules`, and
  every other element under its parent or in an orphaned collection.
- A container SHALL be placed among the containers of the system its
  system property names. WHEN that property is absent, or names no
  member at the `system` level, the container SHALL be placed in
  `orphaned_containers`.
- A component SHALL be placed among the components of the container or
  module whose roster declares it, by record block or by a reference to
  its blueprint. WHEN more than one roster declares it, it SHALL be
  placed under the parent whose roster declaration sorts first by path
  and then by span, and nowhere else.
- A component no roster declares SHALL be placed in
  `orphaned_components`.
- A container's and a module's `modules` SHALL list the slugs its
  composed-modules property names. A module's `contained_by` SHALL list
  every container and module whose composed-modules property names it.
- An element's `contracts_provided` SHALL list every member of the
  registered contract type whose provider property names the element.
- Derived properties — `contained_by` and `contracts_provided` — SHALL
  be computed over the whole corpus, including when the model is
  narrowed to a subtree.
- An element's `declarations` SHALL list every declaration C4 Elements
  holds for its slug, each with its document's path and the span of the
  frontmatter block or of the record block, fences included.
- WHERE an element has a blueprint, its properties SHALL be taken from
  that blueprint's frontmatter. Otherwise they SHALL be taken from the
  declaration whose path sorts first, and then whose span does.
- The properties copied SHALL be `entity_name`, `description`,
  `technology`, `contracts_consumed` and `tags`, and for a module `kind`,
  `visibility` and `version`, and for a component with a blueprint
  `kind`. A property the source declaration does not carry SHALL be
  absent.
- No element SHALL be created from a diagram node.
- Every collection of elements SHALL be ordered by slug, and every
  collection of slugs likewise.

**Anchors:** a new package under `cli/core`, beside the C4 Elements
package (`cli/core/c4elements`); `specmodel.C4Model` and its
member types, generated from `schemas/c4.strata`.

### C4 Model in [[module-document-core]] (modified — relationships)

**Intent:** A relationship is read from every drawing of it, in both of
the homes [[adr-009]] gives an edge, and recorded once. The grouping and
the endpoint resolution are C4 Elements'; this delta turns them into
the model's edge list and exclusion list.

**Behavior:**
- The component SHALL record one relationship for each group of
  drawings C4 Elements reports as one relationship, whatever document
  and view each drawing sits in.
- A relationship's `description`, `technology` and `detail` SHALL be the
  label, technology argument and description argument of its drawing
  that sorts first by path and then by span. An absent argument, or an
  argument written as the empty string, SHALL leave the property absent.
- A relationship's `drawn_in` SHALL list every drawing, ordered by path
  and then by span, each carrying its path, its span, its view and its
  source and target aliases after `Rel_Back` normalisation.
- A relationship's `level` SHALL be the finer of its two endpoints'
  levels, where `system` is coarser than `container`, and `container` is
  coarser than both `module` and `component`, which are recorded as
  `component`.
- A statement C4 Elements returns unresolved SHALL be recorded in
  `excluded_relationships` with the reason it gives, its path, its span
  and its text as written.
- A `BiRel` statement SHALL be recorded in `excluded_relationships` with
  reason `bidirectional`.
- A statement the parser could not read SHALL contribute nothing to the
  model.
- `C4Dynamic` and `C4Deployment` diagrams, and every diagram outside a
  member whose type declares an element, SHALL contribute no
  relationship and no excluded relationship.
- `relationships` SHALL be ordered by source slug, then target slug,
  then description; `excluded_relationships` by path, then span.

**Anchors:** the C4 Model package; the C4 Elements drawings API.

### C4 Model in [[module-document-core]] (modified — reserved tags)

**Intent:** Mermaid has no argument for whether a call waits or which
contract governs it, so [[adr-029]] reserves tags for both. Whether a
tag is well formed is the diagram checks' question; this delta records
what a tag means once it is read.

**Behavior:**
- WHEN the first drawing carries `sync` and not `async`, the
  relationship's `interaction` SHALL be `sync`; WHEN it carries `async`
  and not `sync`, `async`. WHEN it carries both or neither, the
  `interaction` SHALL be absent.
- WHEN the first drawing carries a `contract:` tag, the relationship's
  `contract` SHALL carry the slug that tag names, with source `tag`,
  whether or not that slug resolves. WHERE it carries several, the first
  as written SHALL be used.
- WHEN the first drawing carries no `contract:` tag and exactly one
  contract in the source element's `contracts_consumed` names the target
  element as its provider, the relationship's `contract` SHALL carry
  that contract's slug, with source `derived`.
- WHEN no contract tag is present and zero or several contracts fit,
  the relationship SHALL carry no `contract`.
- A relationship's `tags` SHALL be the first drawing's unreserved tags in
  the order written, and SHALL be absent when there are none.

**Anchors:** the C4 Model package; the tag split the C4 Diagrams
component returns.

### C4 Model in [[module-document-core]] (modified — subtree)

**Intent:** An agent planning a change to one element wants that
element's part of the architecture, not the whole. Narrowing is the
query's, so a wrapper gets the same answer as the command.

**Behavior:**
- WHERE a root slug is given, the component SHALL return a model holding
  only that element and its descendants: for a system, its containers
  and everything they hold; for a container, its components, the modules
  composed into it, and theirs; for a module, its components, the
  modules composed into it, and theirs; for a component, nothing else.
- The root's ancestors SHALL appear only as the chain from the top of
  the tree down to the root, each holding no other child. A module root
  has no ancestor in the tree and SHALL appear in `modules`; a component
  whose parent is a module SHALL appear under that module in `modules`.
- An orphaned container or component given as the root SHALL appear in
  its orphaned collection.
- The narrowed model SHALL hold only the relationships whose source and
  target are both in the narrowed set, and only the excluded
  relationships whose statement sits in a blueprint of an element in
  that set.
- A narrowed model SHALL carry `root_declared`, true where the root slug
  names an element. WHERE it names no element, the component SHALL
  return a model with `root_declared` false and every collection empty,
  and no error (AC-C4-004.6).

**Anchors:** the C4 Model package.

### Command Surface in [[container-spec-cli]] (modified — `c4`)

**Intent:** A new subcommand, wired like `list`: flags in, one Document
Core call, one rendering out. It decides nothing about the model.

**Behavior:**
- The subcommand SHALL be named `c4` and SHALL accept no positional
  arguments.
- The subcommand SHALL accept `--root <slug>` and `--json`, and no
  other flag.
- The subcommand SHALL open the corpus warm through the shared
  `openCorpus` helper, as `check` and `list` do.
- The subcommand SHALL write the model as YAML by default and as JSON
  under `--json`, both carrying the same values in the same order, per
  [[cli-output-standards]].
- The model SHALL carry `schema_version` `0.1` and the workspace name
  the configuration resolves.
- WHEN the model holds excluded relationships, or `specs check` would
  report `c4/*` findings for the same corpus, the subcommand SHALL still
  exit zero.
- WHERE `--root` names no element, the subcommand SHALL write that model,
  `root_declared: false` with every collection empty, write nothing to
  standard error, and exit zero (AC-C4-004.6).
- Every flag's `--help` text SHALL come from its declaration, per the
  container's Change Guidance.

**Anchors:** a new `cli/cmd/c4.go` beside `cli/cmd/list.go`, which is
the shape to follow; registration in `newRootCmd` in `cli/cmd/root.go`;
`renderYAML` in `cli/cmd/yaml.go`; snapshot cases under
`cli/testdata/cases/`.

### Documentation (modified — `c4`)

**Intent:** The command is not discoverable until the reference lists
it, and [[frd-spec-c4]] says the output standards' conformance table
gains a row when it lands.

**Behavior:**
- [[spec-cli-reference]] SHALL document `specs c4`, its two flags, the
  model's top-level collections, and the excluded-relationship reasons.
- [[cli-output-standards]]' current-conformance table SHALL gain a `c4`
  row recording that the command conforms.

**Anchors:** `specs/guides/spec-cli-reference.md`;
`specs/standards/cli-output-standards.md`.
`scripts/sync-shared-references.sh` SHALL be re-run so the vendored
copies follow.

### Relied-upon elements (unchanged)

- [[module-document-core]]'s C4 Elements component — the element
  set, each element's declarations, node resolution, and drawings
  grouped into relationships with unresolved statements and their
  reasons. This feature adds no resolution of its own.
- [[module-document-core]]'s C4 Diagrams component — statements,
  views and the reserved-tag split.
- The architecture vocabulary the profile registers per [[adr-030]],
  read through Document Core's extension point.
- The index's per-type frontmatter, which carries every property TREE
  copies.
- `renderYAML` and `openCorpus` in `cli/cmd`, which the CMD delta uses
  as `list` does.

## Contracts

None. The model's shape is declared in `schemas/c4.strata` and emitted
as JSON Schema with every other generated model. Like the index
([[container-spec-cli]], Open Questions), it is not promoted to a
contract document until a consumer exists that is built independently.

## Tracker Reconciliation

- `specs-f6tl.5` and `specs-f6tl.6` — prerequisite,
  transitively, already met: Document Core's C4 Elements component
  rests on both, and this feature rests on that component.
- `agent-plugins-7os0` — out-of-scope: subsumed by
  the [[c4-standards]] amendment that landed with Document Core's C4
  Elements component. This feature does not amend
  [[c4-standards]].
- `agent-plugins-h6c` — out-of-scope: the index-schema and contract
  ADR. This feature emits a separate model and changes no index shape,
  so it neither advances nor blocks that decision.
- `agent-plugins-gz3` — prerequisite, already met: the index carries
  per-type frontmatter since the change that closed
  `agent-plugins-23f1`.

## Sequencing Constraints

- Document Core's C4 Elements component, with its drawings, SHALL be in
  place before the TREE delta; it is.
- The MODEL delta SHALL land before the REL delta.
- The TREE delta SHALL land before the REL, TAGS and ROOT deltas.
- The REL delta SHALL land before the TAGS and ROOT deltas.
- The TREE, REL, TAGS and ROOT deltas SHALL land before the CMD delta.
- Delivery is `single-pr`: the DOC delta and the fold of every delta
  into [[module-document-core]] and [[container-spec-cli]] SHALL land in
  the same change as the code, so the main branch never carries a
  command the container blueprint does not list.
- WHEN the CMD delta changes observable output, the snapshot cases under
  `cli/testdata/cases/` SHALL be regenerated with
  `scripts/regenerate-cli-snapshots.sh` and every diff reviewed before
  commit, per [[container-spec-cli]]'s verification gates.

## AC Coverage

Legend: MODEL = `C4ExclusionReason` delta; TREE = C4 Model delta
(element tree); REL = C4 Model delta (relationships); TAGS = C4 Model
delta (reserved tags); ROOT = C4 Model delta (subtree); CMD = Command
Surface `c4` delta; DOC = Documentation delta.

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-C4-001.1 | **TREE (proves)** | unit |
| AC-C4-001.2 | **TREE (proves)** | unit |
| AC-C4-001.3 | **TREE (proves)** | unit |
| AC-C4-001.4 | **TREE (proves)** | unit |
| AC-C4-001.5 | **TREE (proves)** | unit |
| AC-C4-001.6 | **TREE (proves)** | unit |
| AC-C4-001.7 | **TREE (proves)** | unit |
| AC-C4-001.8 | **TREE (proves)** | unit |
| AC-C4-001.9 | **TREE (proves)** | unit |
| AC-C4-001.10 | **TREE (proves)** | unit |
| AC-C4-001.11 | **TREE (proves)** | unit |
| AC-C4-002.1 | **REL (proves)** | unit |
| AC-C4-002.2 | **REL (proves)** — via relied-upon C4 Diagrams | unit |
| AC-C4-002.3 | **REL (proves)** — via relied-upon C4 Diagrams | unit |
| AC-C4-002.4 | **REL (proves)** | unit |
| AC-C4-002.5 | **REL (proves)** | unit |
| AC-C4-002.6 | **REL (proves)** | unit |
| AC-C4-002.9 | **REL (proves)** | unit |
| AC-C4-002.10 | MODEL, **REL (proves)** — via relied-upon C4 Elements | unit |
| AC-C4-002.11 | **REL (proves)** | unit |
| AC-C4-002.12 | **REL (proves)** | unit |
| AC-C4-003.1 | **TAGS (proves)** | unit |
| AC-C4-003.2 | **TAGS (proves)** | unit |
| AC-C4-003.3 | **TAGS (proves)** | unit |
| AC-C4-003.4 | **TAGS (proves)** | unit |
| AC-C4-003.5 | **TAGS (proves)** | unit |
| AC-C4-004.1 | TREE, **CMD (proves)** | e2e |
| AC-C4-004.2 | TREE, **CMD (proves)** | e2e |
| AC-C4-004.3 | **CMD (proves)** | e2e |
| AC-C4-004.4 | TREE, REL, **CMD (proves)** | integration |
| AC-C4-004.5 | ROOT, **CMD (proves)** | e2e |
| AC-C4-004.6 | ROOT, **CMD (proves)** | e2e |
| AC-C4-004.7 | REL, **CMD (proves)** | e2e |

AC-C4-004.4 is declared `integration` because the snapshot harness
replays each case once and cannot compare two runs; a Go test in
`cli/cmd` runs the command twice and compares the bytes, as
`list_test.go` does for AC-SPECLIST-004.3.

## Key Decisions

- [[adr-029]] — relationships are read from blueprint diagrams, with
  reserved relationship tags.
- [[adr-030]] — the model is assembled in Document Core from vocabulary
  the profile registers.
- [[adr-009]] — an edge's two homes, which is why one relationship can
  carry several drawings.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — C4 Model added to the Components
        roster and diagram with its record block, including the subtree
        query
  - [x] [[container-spec-cli]] — the Command Surface's subcommand list
        gains `c4`, with its flags and the Document Core call it makes
- [x] `schemas/c4.strata` carries `bidirectional`, and the generated
      schema and Go model are regenerated from it
- [x] [[spec-cli-reference]] documents `specs c4`, [[cli-output-standards]]
      lists it, and `scripts/sync-shared-references.sh --check` passes
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per [[ac-traceability-standards]]
- [x] `scripts/coverage.sh` has been run from the repository root and
      the README diff committed, since `cli/**` changed
- [x] This blueprint's status → retired

## Open Questions

None.

## References

- [[c4-standards]] — the levels, containment and arrow rules the model
  follows
- [[frd-spec-c4]] — the requirements this blueprint realizes
- [[container-spec-cli]] — the container whose Change Guidance binds
  every delta
- [[module-document-core]] — the module the model is assembled in, whose
  C4 Diagrams and C4 Elements components are the parser, element set and
  drawings this model is built from
- [[cli-output-standards]] — the rendering rules the CMD delta follows
- [[ac-traceability-standards]] — the citation obligations the coverage
  table feeds
