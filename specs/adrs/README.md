# Architecture decision records

This directory contains architecture decision records (ADRs). Each ADR
documents a significant design or technology choice for the
specification-driven development framework.

## ADR index

- [ADR-001: Use HTML Comment Blocks for Template Instructions](adr-001-html-comments-for-instructions.md):
  HTML comments embed guidance without cluttering rendered output.
- [ADR-002: Use EARS Notation for Requirements](adr-002-ears-for-requirements.md):
  EARS provides standardized syntax patterns for clear, testable
  requirements.
- [ADR-003: Use YAML Frontmatter for Document Metadata](adr-003-yaml-frontmatter-for-metadata.md):
  YAML frontmatter separates structured metadata from document content.
- [ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf](adr-004-module-blueprint-and-atomic-component.md):
  Split the overloaded component blueprint into an atomic component type and
  a new module type.
- [ADR-005: Classify Composed Units by the Boundary of Independent Change](adr-005-independent-change-classification.md):
  The component/module line is the same independent-change test as contract
  visibility (private, internal, or public).
- [ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals](adr-006-visibility-and-kind-orthogonal-signals.md):
  `visibility: public|internal` and `kind: standard|pattern` are separate
  axes, not one overloaded field.
- [ADR-007: Create Blueprints by Promotion at a Seam](adr-007-promotion-based-blueprint-creation.md):
  A unit starts as a record-block bullet in its container and graduates to its
  own blueprint at a self-contained seam.
- [ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model](adr-008-element-record-blocks.md):
  Un-promoted elements carry a fenced record block that's validated by the
  same base schema that a promoted document's frontmatter satisfies.
- [ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges](adr-009-nodes-and-edges-ownership.md):
  File every relationship at one altitude; leaves stay light on outward edges
  and code.
- [ADR-010: Contracts Are Independent, Long-Lived Documents](adr-010-contract-independence.md):
  Contracts outlive and might precede implementations; source live schemas
  from the running service.
- [ADR-011: Blueprints Declare Targeted Verification Gates](adr-011-targeted-verification-gates.md):
  Each blueprint names the minimal tests to run on change; agents run that
  set, and CI runs the full suite.
- [ADR-012: Wikilink and Anchor Reference Conventions](adr-012-reference-conventions.md):
  `[[slug]]` for documents, heading anchors for un-promoted components, and
  backticks for code; replaces `@` and the `#` element convention.
- [ADR-013: Work Order as the Name for the Atomic Unit of Executable Work](adr-013-work-order-terminology.md):
  The portable unit is a work order, and its beads materialization is a bead;
  "work item" and "task" are retired as synonyms.
- [ADR-015: Go as the Implementation Language for the Spec Corpus CLI](adr-015-go-for-the-spec-cli.md):
  Go ties Rust on every discriminating measure and is the only candidate with
  an official in-process Gherkin parser.
- [ADR-016: Strata as the Single Source for Schemas and Generated Types](adr-016-strata-for-schema-and-model-generation.md):
  One definition emits JSON Schema, Go structs, and Markdown. TypeSpec and
  Smithy are API IDLs that produce neither plain Go structs nor Markdown.
- [ADR-017: The specs Plugin Provisions the Spec CLI Version It Pins](adr-017-spec-cli-as-a-plugin-prerequisite.md):
  A SessionStart hook puts the CLI release that matches the plugin's version
  first on the session's `PATH`, shadowing any other version. Skills call the
  CLI by a single path.
- [ADR-018: References Are an Edge Collection in the Index](adr-018-references-as-index-edges.md):
  One top-level `references` list carries each occurrence with its span,
  origin, and resolution. Both directions are derived from that list instead
  of being mirrored into each entry.
- [ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers](adr-020-frd-lifecycle-and-ac-keyed-deltas.md):
  An FRD is a living document that's edited in place under a same-feature test
  keyed on its slug, and is never amended by a successor FRD. Changes between
  versions are diffed by AC identifier, not by line.
- [ADR-021: Adopt Gherkin for Features and Acceptance Criteria](adr-021-gherkin-blocks-for-acceptance-criteria.md):
  The AST yields typed records that a central store can load, where prose
  yields one string. `Scenario Outline` keeps the criterion/example split and
  makes examples queryable. EARS keeps every other normative statement.
- [ADR-028: Default Directory Structure for the Spec Corpus](adr-028-specs-directory-structure.md):
  The spec root mirrors the C4 containment hierarchy only where the
  containment edge has no child-side declaration, so components nest under
  their one owner while modules stay top-level. ADRs, standards, guides, and
  contracts each get one flat home.
- [ADR-029: C4 Relationships Are Read From Blueprint Diagrams](adr-029-c4-relationships-from-blueprint-diagrams.md):
  Neighbor edges are extracted from the Mermaid C4 diagrams in the blueprints
  that ADR-009 makes their home. A parser that this CLI owns does the
  extraction, with reserved `sync`, `async`, and `contract:` tags for what
  Mermaid can't express.
- [ADR-030: C4 Vocabulary Reaches Document Core Through the Profile](adr-030-c4-vocabulary-through-the-profile.md):
  The C4 parser, element resolution, and model live in Document Core. The
  architecture vocabulary that they read arrives as a profile registration,
  as the Gherkin tag vocabulary does.
- [ADR-031: Base Revisions Are Read Through the git Executable](adr-031-git-executable-for-base-revisions.md):
  `criterion --changed` reads its base commit by running `git` and copying the
  spec root's tracked blobs into a temporary directory, instead of linking a
  git library or using `git archive`.

ADR-019 and ADR-022 through ADR-027 moved to
[brettcallaghan/swarm](https://github.com/brettcallaghan/swarm) with the
`swarm` CLI and its plugin. They decide things about that codebase and are
maintained in its corpus.

ADR-014 is unused. A draft at that number deliberated the same question that
ADR-021 settles, and ADR-021 replaced it instead of retaining it as a retired
record.

## About ADRs

ADRs capture the context, options considered, and consequences of significant
technical decisions. They create a decision log that explains why the project
is built the way it is, for future maintainers, and they serve as
institutional memory as the project evolves.

For the ADR template and usage guidelines, see
[/plugin/skills/spec/references/adr-template.md](../../plugin/skills/spec/references/adr-template.md).
