# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records documenting significant design and technology choices for the specification-driven development framework.

## ADR Index

- [ADR-001: Use HTML Comment Blocks for Template Instructions](adr-001-html-comments-for-instructions.md) — HTML comments embed guidance without cluttering rendered output
- [ADR-002: Use EARS Notation for Requirements](adr-002-ears-for-requirements.md) — EARS provides standardized syntax patterns for clear, testable requirements
- [ADR-003: Use YAML Frontmatter for Document Metadata](adr-003-yaml-frontmatter-for-metadata.md) — YAML frontmatter separates structured metadata from document content
- [ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf](adr-004-module-blueprint-and-atomic-component.md) — split the overloaded component blueprint into an atomic component type and a new module type
- [ADR-005: Classify Composed Units by the Boundary of Independent Change](adr-005-independent-change-classification.md) — the component/module line is the same independent-change test as contract visibility (private/internal/public)
- [ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals](adr-006-visibility-and-kind-orthogonal-signals.md) — `visibility: public|internal` and `kind: standard|pattern` are separate axes, not one overloaded field
- [ADR-007: Create Blueprints by Promotion at a Seam](adr-007-promotion-based-blueprint-creation.md) — a unit starts as a record-block bullet in its container and graduates to its own blueprint at a self-contained seam
- [ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model](adr-008-element-record-blocks.md) — un-promoted elements carry a fenced record block validated by the same base schema a promoted document's frontmatter satisfies
- [ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges](adr-009-nodes-and-edges-ownership.md) — file every relationship at one altitude; leaves stay light on outward edges and code
- [ADR-010: Contracts Are Independent, Long-Lived Documents](adr-010-contract-independence.md) — contracts outlive and may precede implementations; source live schemas from the running service
- [ADR-011: Blueprints Declare Targeted Verification Gates](adr-011-targeted-verification-gates.md) — each blueprint names the minimal tests to run on change; agents run that set, CI runs the full suite
- [ADR-012: Wikilink and Anchor Reference Conventions](adr-012-reference-conventions.md) — `[[slug]]` for documents, heading anchors for un-promoted components, backticks for code; replaces `@` and the `#` element convention
- [ADR-013: Work Order as the Name for the Atomic Unit of Executable Work](adr-013-work-order-terminology.md) — the portable unit is a work order and its beads materialization is a bead; "work item" and "task" are retired as synonyms
- [ADR-015: Go as the Implementation Language for the Spec Corpus CLI](adr-015-go-for-the-spec-cli.md) — Go ties Rust on every discriminating measure and is the only candidate with an official in-process Gherkin parser
- [ADR-016: Strata as the Single Source for Schemas and Generated Types](adr-016-strata-for-schema-and-model-generation.md) — one definition emits JSON Schema, Go structs and Markdown; TypeSpec and Smithy are API IDLs that produce neither plain Go structs nor Markdown
- [ADR-017: The Spec CLI Is a Prerequisite of the sf Plugin, Not a Payload](adr-017-spec-cli-as-a-plugin-prerequisite.md) — the plugin stays pure markdown and names the binary as a prerequisite; skills call it in a single path rather than branching on its absence
- [ADR-018: References Are an Edge Collection in the Index](adr-018-references-as-index-edges.md) — one top-level `references` list carries each occurrence with its span, origin and resolution; both directions are derived from it rather than mirrored into each entry
- [ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers](adr-020-frd-lifecycle-and-ac-keyed-deltas.md) — an FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
- [ADR-021: Adopt Gherkin for Features and Acceptance Criteria](adr-021-gherkin-blocks-for-acceptance-criteria.md) — the AST yields typed records a central store can load where prose yields one string; `Scenario Outline` keeps the criterion/example split and makes examples queryable; EARS keeps every other normative statement
- [ADR-028: Default Directory Structure for the Spec Corpus](adr-028-specs-directory-structure.md) — the spec root mirrors the C4 containment hierarchy only where the containment edge has no child-side declaration, so components nest under their one owner while modules stay top-level; ADRs, standards, guides and contracts get one flat home each

ADR-019 and ADR-022 through ADR-027 moved to [brettcallaghan/swarm](https://github.com/brettcallaghan/swarm) with the `swarm` CLI and the `bead-bonanza` plugin. They decide things about that codebase and are maintained in its corpus.

ADR-014 is unused: a draft at that number deliberated the same question
ADR-021 settles, and was replaced by it rather than retained as a retired
record.

## About ADRs

Architecture Decision Records capture the context, options considered, and consequences of significant technical decisions. They create a decision log that explains "why we built it this way" for future maintainers and serve as institutional memory as the project evolves.

For the ADR template and usage guidelines, see [/plugin/skills/spec/references/adr-template.md](../../plugin/skills/spec/references/adr-template.md).
