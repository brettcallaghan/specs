# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records documenting significant design and technology choices for the specification-driven development framework.

## ADR Index

- [ADR-001: Use HTML Comment Blocks for Template Instructions](ADR-001-html-comments-for-instructions.md) — HTML comments embed guidance without cluttering rendered output
- [ADR-002: Use EARS Notation for Requirements](ADR-002-ears-for-requirements.md) — EARS provides standardized syntax patterns for clear, testable requirements
- [ADR-003: Use YAML Frontmatter for Document Metadata](ADR-003-yaml-frontmatter-for-metadata.md) — YAML frontmatter separates structured metadata from document content
- [ADR-004: Introduce a Module Blueprint; Component as the Atomic Leaf](ADR-004-module-blueprint-and-atomic-component.md) — split the overloaded component blueprint into an atomic component type and a new module type
- [ADR-005: Classify Composed Units by the Boundary of Independent Change](ADR-005-independent-change-classification.md) — the component/module line is the same independent-change test as contract visibility (private/internal/public)
- [ADR-006: Model Visibility and Kind as Orthogonal Frontmatter Signals](ADR-006-visibility-and-kind-orthogonal-signals.md) — `visibility: public|internal` and `kind: standard|pattern` are separate axes, not one overloaded field
- [ADR-007: Create Blueprints by Promotion at a Seam](ADR-007-promotion-based-blueprint-creation.md) — a unit starts as a record-block bullet in its container and graduates to its own blueprint at a self-contained seam
- [ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model](ADR-008-element-record-blocks.md) — un-promoted elements carry a fenced record block validated by the same base schema a promoted document's frontmatter satisfies
- [ADR-009: Nodes-vs-Edges Ownership — Parents Own Child Edges](ADR-009-nodes-and-edges-ownership.md) — file every relationship at one altitude; leaves stay light on outward edges and code
- [ADR-010: Contracts Are Independent, Long-Lived Documents](ADR-010-contract-independence.md) — contracts outlive and may precede implementations; source live schemas from the running service
- [ADR-011: Blueprints Declare Targeted Verification Gates](ADR-011-targeted-verification-gates.md) — each blueprint names the minimal tests to run on change; agents run that set, CI runs the full suite
- [ADR-012: Wikilink and Anchor Reference Conventions](ADR-012-reference-conventions.md) — `[[slug]]` for documents, heading anchors for un-promoted components, backticks for code; replaces `@` and the `#` element convention
- [ADR-013: Work Order as the Name for the Atomic Unit of Executable Work](ADR-013-work-order-terminology.md) — the portable unit is a work order and its beads materialization is a bead; "work item" and "task" are retired as synonyms
- [ADR-015: Go as the Implementation Language for the Spec Corpus CLI](ADR-015-go-for-the-spec-cli.md) — Go ties Rust on every discriminating measure and is the only candidate with an official in-process Gherkin parser
- [ADR-016: Strata as the Single Source for Schemas and Generated Types](ADR-016-strata-for-schema-and-model-generation.md) — one definition emits JSON Schema, Go structs and Markdown; TypeSpec and Smithy are API IDLs that produce neither plain Go structs nor Markdown
- [ADR-017: The Spec CLI Is a Prerequisite of the sf Plugin, Not a Payload](ADR-017-spec-cli-as-a-plugin-prerequisite.md) — the plugin stays pure markdown and names the binary as a prerequisite; skills call it in a single path rather than branching on its absence
- [ADR-018: References Are an Edge Collection in the Index](ADR-018-references-as-index-edges.md) — one top-level `references` list carries each occurrence with its span, origin and resolution; both directions are derived from it rather than mirrored into each entry
- [ADR-019: Orchestrate Work Orders with a Skill Composing Native Subagents](ADR-019-skill-composed-subagent-orchestration.md) — `bead-bonanza` composes the `Agent` primitive rather than adopting agent teams, background sessions, `/batch` or dynamic workflows, whose durability and input models do not fit an hours-long run that commits
- [ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers](ADR-020-frd-lifecycle-and-ac-keyed-deltas.md) — an FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
- [ADR-021: Adopt Gherkin for Features and Acceptance Criteria](ADR-021-gherkin-blocks-for-acceptance-criteria.md) — the AST yields typed records a central store can load where prose yields one string; `Scenario Outline` keeps the criterion/example split and makes examples queryable; EARS keeps every other normative statement
- [ADR-022: Projects Declare Their Checks in sf.toml](ADR-022-project-declared-checks.md) — a project declares its checks as an array of id-and-command entries in a plugin-scoped `sf.toml` and a plugin-shipped runner executes them, because prose cannot express a multi-command check, a suite-wide exit code cannot express a known-red baseline, and asking every project to write its own runner puts the same code in every repository
- [ADR-023: Go for the Swarm CLI, in the Existing Module](ADR-023-go-for-the-swarm-cli.md) — Go for the same measured reasons as the spec CLI, and a shell language is disqualified by the pipeline hazards the swarm CLI exists to remove; a second binary in the existing module inherits the three-OS matrix and the six-target cross-compile
- [ADR-024: A Command Builds a Plan; Three Interpreters Read It](ADR-024-plans-are-values.md) — a command produces an ordered plan of typed step values that explain, dry-run and execute all read, because a second code path for explain drifts and a closure cannot be rendered without being run
- [ADR-025: Every External Invocation Is Argv Through One Gateway](ADR-025-argv-not-shell.md) — one component runs every `bd` and `git` invocation as an argv vector with separated streams, a status read from the process and a per-subcommand decoder, because a shell string loses all four and Windows has no `sh`
- [ADR-026: No Rollback — Validate, Order, Declare Idempotence](ADR-026-no-rollback.md) — a failed plan stops and reports where it stopped rather than compensating, because a push and a worktree removal cannot be reversed and a guarantee that holds for some steps is not a guarantee
- [ADR-027: Two Test Tiers, and a Committed Fixture for the Upper One](ADR-027-two-test-tiers-and-a-committed-fixture.md) — unit tests assert the plan and replay recorded tool output, end-to-end tests execute the compiled binary against a repository and tracker restored from committed artifacts, because faking the tools tests the wrapper and not the wrapping and a fixture generator is code that can be wrong

ADR-014 is unused: a draft at that number deliberated the same question
ADR-021 settles, and was replaced by it rather than retained as a retired
record.

## About ADRs

Architecture Decision Records capture the context, options considered, and consequences of significant technical decisions. They create a decision log that explains "why we built it this way" for future maintainers and serve as institutional memory as the project evolves.

For the ADR template and usage guidelines, see [/plugin/skills/spec/references/adr-template.md](../../plugin/skills/spec/references/adr-template.md).
