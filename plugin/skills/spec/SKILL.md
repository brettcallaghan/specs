---
name: spec
description: Spec-first development using this framework's specification
  document types — ADRs (architecture decisions), FRDs (feature requirements),
  standards documents (cross-cutting standards, policies, and constraints), architecture
  blueprints (system, container, component), feature blueprints (how a
  single feature is built), and contracts (interface specifications). Use
  when starting new work, scoping or specifying a feature, recording a
  technology or architecture decision, defining standards, policies, or
  constraints, writing acceptance criteria, designing how a specified
  feature will be implemented, or documenting the architecture of a
  system, container, component, or interface. Covers choosing the right
  document type and authoring it from the framework templates using Gherkin
  and EARS syntax and the C4 model.
version: "1.10"
---

# Spec-First Development

**Core principle:** Write the spec before the code. Requirements,
constraints, and decisions are captured in version-controlled markdown
before implementation begins. Code records how something was built; specs
record what it must do and why — and they are the contract that design,
implementation, tests, and reviews trace back to.

Why spec first:

- **Disputes are cheap before code exists.** A wrong requirement costs a
  sentence edit at spec time and rework at implementation time.
- **Testable by construction.** An acceptance criterion written as a
  tagged scenario, and a requirement written in EARS, are both precise
  enough to write a failing test from — so implementation is verified
  against the spec rather than against intuition.
- **Decision memory.** Recorded decisions prevent relitigating settled
  questions and prevent accidental reversals by future maintainers (human
  or agent).
- **Shared contract.** Everyone working on the system — including coding
  agents — works from the same written source of truth instead of private
  understandings.

Skipping ahead creates rework: an FRD written before the constraints and
decisions it depends on will be rewritten once they land.

---

## Document Types

The specification document types are listed in the table below: some for
decisions, requirements, and cross-cutting standards, some for
architecture documentation, one — the Feature Blueprint — for feature
design, and one — the Work Order — for a single unit of executable work.
Use only these types; do not invent new ones. If work does not fit any of
them, say so rather than forcing it into the wrong type.

The Work Order is the one type this skill does not itself author. It is
the framework's portable, atomic unit of executable work — everything a
fresh agent needs to complete one job — and it is *generated* by a
grooming skill (`groom-beads`, or `groom-feature-blueprint` feeding
it), never hand-written through the authoring workflow below. The `spec`
skill owns its format and its readiness bar (the Definition of Ready);
grooming owns its production. What remains a derived, non-spec artifact
is the *plan*: the whole graph of work orders for a feature. Only the
atomic order is promoted to a type — not the aggregate, and not the
grooming that emits it. Improvement RFCs are likewise not specification
documents: they are process artifacts produced by the `retro-spec`
skill.

| Type | Captures | One per |
|------|----------|---------|
| **ADR** — Architecture Decision Record | A significant decision: the context, alternatives considered, the choice, and its consequences | Decision |
| **FRD** — Feature Requirements Document | The WHY (user story) and WHAT (Gherkin acceptance criteria) for a single feature | Feature |
| **Standards** | Cross-cutting standards, policies, invariants, or constraints that span features or systems (the type name is shorthand for this whole family) | Domain or scope |
| **System Blueprint** | A system's context, containers, public contracts, and deployment (C4 levels 1-2) | System |
| **Container Blueprint** | One deployable unit: its context, components, provided/consumed contracts, owned data, and change guidance for feature work | Container |
| **Component Blueprint** | An atomic component warranting standalone docs — an internal building block of one container, plus its interface and change guidance | Component |
| **Module Blueprint** | A unit that crosses a boundary of independent change: shared/published across containers (library, SDK, UI package, IaC module) and/or composite with internal components — plus its versioning and change rules | Module |
| **Feature Blueprint** | The HOW pairing 1:1 with an FRD's WHAT: a transient, diff-shaped design — element deltas, contracts, sequencing constraints, AC coverage, and a definition of done | Feature |
| **Contract** | The prose semantics companion to a machine-readable interface spec (API, event, data product) | Interface |
| **Work Order** | The atomic, self-contained unit of executable work — everything a fresh, zero-context agent needs to complete one job. Generated by grooming (not authored here); materialized into one tracker issue (bead) | Unit of work |

The architecture documents follow the C4 model with the framework's
documented extensions — load `references/shared/c4-standards.md` before authoring
any of them; it maps concrete elements (queues, shared libraries, IaC
modules, sidecars, data products) to the correct C4 level and document.

### Choosing the type

Ask these questions in order:

1. **Is this a choice between two or more real options** with trade-offs
   and consequences that are hard to reverse (technology, architectural
   pattern, API contract)? → **ADR**
2. **Is this a single feature** that delivers complete, independently
   testable user value? → **FRD**
3. **Is this the design for HOW an already-specified feature will be
   built** — the element deltas, contracts, and constraints that realize
   an FRD? → **Feature Blueprint**
4. **Is this a rule that applies across features or systems** — a coding
   or security standard, a compliance obligation, an architectural
   constraint, a system-wide invariant? → **Standards**
5. **Does this describe a system your team owns** — its context among
   neighbors, its containers, its deployment? → **System Blueprint**
6. **Does this describe one deployable unit** — an application or data
   store your team builds and runs? → **Container Blueprint**
7. **Does this describe an atomic component of a single container** — an
   internal building block with a defined interface, changing in lockstep
   with its container? → **Component Blueprint**
8. **Does this describe a unit that crosses a boundary of independent
   change** — shared or published across containers, and/or composite
   with internal components of its own? → **Module Blueprint**
9. **Does this specify an interface consumed across a boundary of
   independent change** — an API, event stream, or data product that
   other codebases evolve against? → **Contract**

WHEN two types both look plausible, or the obvious answer may be the
wrong one — component vs. module, a private schema vs. a Contract, a
work order vs. the FRD and blueprint upstream of it, an "architectural"
question that is really a requirements question, what may live only in a
feature blueprint — read `references/shared/document-type-boundaries-guide.md` and
decide from it rather than by analogy.

### Ordering

There is no rigid pipeline, but dependencies flow one way: cross-cutting
requirements and constraints inform decisions, and both inform feature
specs.

- Do not write an ADR before the requirements context that motivates it is
  understood — it will answer the wrong question.
- Do not write an FRD that silently depends on an unmade decision; write
  (or at least draft) the ADR first and reference it.
- FRDs comply with requirements documents by reference: "This feature
  SHALL comply with [[security-standards]]" — never by restating them.
- A feature blueprint is authored after its FRD is stable, and the
  contracts and ADRs it introduces are drafted alongside it — the
  blueprint references them; it never contains them.
- WHEN a feature blueprint reaches `approved` with no Open Questions,
  hand off to the `groom-feature-blueprint` skill — it translates the blueprint into
  an executable work-item graph. Authoring stops at the blueprint; task
  decomposition is grooming's job, not this skill's.

---

## Authoring Workflow

1. **Pick the template** from `references/` (table below) and copy it as
   the starting point for the new document. **Write it under the
   project's spec root**, which `specs.toml` declares as
   `[workspace].root` — read it rather than assuming. Where a project
   has no `specs.toml`, its `AGENTS.md` or `CLAUDE.md` names the root,
   and `docs/` is the fallback. Within the root, the default structure
   is:

   ```
   specs/
     adrs/          adr-<nnn>-<name>.md
     standards/     <name>-standards.md
     guides/        <name>-guide.md
     contracts/     <name>-contract.md
     systems/<system-name>/<system-name>-system.md
     containers/<container-name>/
       <container-name>-container.md
       components/<component-name>-component.md
     modules/<module-name>/
       <module-name>-module.md
       components/<component-name>-component.md
     features/<feature-name>/
       <feature-name>-frd.md
       <feature-name>-blueprint.md
       <subfeature-name>/
   ```

   Filenames carry the document type as a **suffix**; only the ADR leads
   with its type, because its number is how it is cited and sorted. Never
   double a suffix, and never stack one on a name already ending in a role
   noun (`-checklist`, `-protocol`, `-reference`).

   The tree mirrors containment only where the child declares no
   parent and has exactly one. A component blueprint carries no parent
   property and has one parent, so its directory is the only thing
   naming its owner — nest it. A module may be composed into several
   containers, each listing it in its own `modules`; no tree can hold
   that, so `modules/` stays at the top level. A container declares
   `system`, so it nests only when nesting earns its keep:

   | Systems documented | Where container directories live |
   |---|---|
   | None | `containers/<container-name>/` |
   | One | `containers/<container-name>/` |
   | Two or more | `systems/<system-name>/containers/<container-name>/` |

   ADRs, standards, guides and contracts get one flat top-level home
   each, never a copy under the level they govern — two directories
   that could each correctly receive a new standard destroy its claim
   to be the single place a rule is recorded. A contract is never
   filed under a provider: it outlives, and may precede, the
   implementations that realize it.

   WHEN the project has an established location for a type, follow it;
   the index does not depend on location, but humans do.
2. **Follow the embedded instructions.** Each template carries its own
   guidance in HTML comment blocks — when to use it, how to complete each
   section, ID and slug conventions. Read them, complete the section, then
   **delete the comment blocks** from the finished document.
3. **Write normative statements in the notation their surface uses.**
   Two notations, and which applies is decided by what the statement is,
   not by the document type:

   - **An FRD's features and acceptance criteria are tagged Gherkin** —
     one fenced ` ```gherkin ` block per requirement, the requirement on
     `Feature:`, its user story in that Feature's description, each
     criterion a `Scenario` or `Scenario Outline` carrying an
     `@AC-<CODE>-NNN.N` tag. Load
     `references/shared/gherkin-guide.md` before writing one; it holds
     the tag vocabulary, what belongs in each description slot, and the
     seven constructs that parse cleanly while losing meaning. Prefer a
     `Scenario Outline` with an `Examples` table where a criterion
     carries concrete example data.
   - **Every other normative statement is EARS** with capitalized RFC
     2119 modal verbs (SHALL / SHOULD / MAY) — standards documents,
     contracts, blueprints, and an FRD's own prose outside its blocks.
     Load `references/shared/ears-guide.md` before writing them.

   Either way the illustration is never normative: the criterion or the
   EARS sentence carries the requirement, and edge-case enumeration
   belongs in tests, not the spec.
4. **Follow the template-authoring standards** for frontmatter (title,
   description, slug, type, status, updated_at), kebab-case slugs with
   type prefixes, 80-character lines, and `[[document-slug]]`
   cross-references. `[[slug]]` is reserved for documents in the
   project's spec set: everything under `references/` — the templates,
   the guides, and the standards shipped with this skill — is cited by
   backticked path, never by wikilink. Cross-reference other documents
   instead of duplicating their content — duplication drifts.

   **Ask the corpus rather than reading it.** For a numbered type, the
   next number is `specs next-id <type>` — never the highest filename you
   happened to see, which misses a document whose filename and slug
   disagree. Before claiming any slug, `specs slug-exists <slug>`. To
   learn what a document depends on before editing it, `specs refs <slug>
   --outbound`. Each is one subprocess and each is checkable by a
   reviewer; a scan you did yourself is neither.
   `references/shared/spec-cli-reference.md` covers the rest of the command surface,
   including reading the whole index when you need many documents'
   metadata at once.
5. **Treat specs as living documents.** Update them when requirements
   change, new information arrives, or errors are found — and update
   `status` and `updated_at` when you do.

   **Before retiring or renaming a document, run `specs refs <slug>`.**
   It lists every document that cites this one, at the line and column
   of each citation — the blast radius, decided rather than estimated.
   Retiring a document leaves those references resolving to a retired
   target, which `specs check` reports as a warning naming the successor;
   declare `superseded_by` so it has one to name. Renaming a slug breaks
   them outright, and the rename is not finished until every line `refs`
   listed has been updated. Either way `specs check <path>` on the
   document you edited reports the findings against the documents citing
   it, so the narrow form is enough to see the blast radius after the
   fact — `refs` is how you see it before.
6. **Preflight feature blueprints.** Before setting a feature blueprint
   to `ready-for-review`, run the decision sweep and self-check against
   `references/shared/blueprint-readiness-checklist.md`, then re-run the checklist
   after fixing what it surfaced — a first pass that fixes check 4 often
   breaks check 8. Do not set `ready-for-review` while any check fails — the same checks
   `groom-feature-blueprint` enforces at its Phase 1 gate, so a first-pass
   blueprint lands groomable instead of discovering the checks by
   failing them. Authoring a feature blueprint includes surveying the
   project's issue tracker to populate its Tracker Reconciliation
   section — existing issues the feature subsumes, depends on, or must
   not absorb are design inputs, not groom-time discoveries. Flag BOTH
   status flips to the human reviewer: the blueprint's approval AND
   its FRD's (the FRD must be `approved` or `active` before grooming
   can start).
7. **Self-check the finished document.** Before setting any document to
   `ready-for-review`, run `specs check <path>` and then
   `references/shared/document-readiness-checklist.md` against it. The
   checklist marks which of its nine checks the CLI settles and which
   are yours; a clean `specs check` is not a passing document, and the
   checks it does not yet make are named there rather than left to
   silence. Fix what both report and re-run; do not set
   `ready-for-review` while any check fails. Feature blueprints
   additionally run the blueprint readiness checklist in step 6 — that
   one is about groomability, this one about conformance, and a
   blueprint passes both.
8. **Offer a retro.** WHEN an authored document consumed upstream
   documents (a feature blueprint from its FRD and architecture
   blueprints; any spec from prior specs) and the review loop has
   settled, offer to run the `retro-spec` skill on the documents this
   session exercised — the session's friction is the evidence, and it
   does not survive into later sessions. The human accepts or
   declines; never run the retro unprompted.

---

## References

Load the whole row. Where a row lists more than one file, those files are
a **bundle** — the template alone is not enough to author the document
correctly, and loading them together is one hop rather than a chain.

| Working on | Load together |
|------------|---------------|
| Checking a document, picking an identifier, querying the corpus | `references/shared/spec-cli-reference.md` |
| Choosing between two document types that both look plausible | `references/shared/document-type-boundaries-guide.md` |
| Reviewing a draft, or self-reviewing before `ready-for-review` | `references/shared/anti-patterns-guide.md` + `references/shared/document-readiness-checklist.md` |
| Recording an architecture or technology decision | `references/adr-template.md` |
| Specifying a feature (user stories + acceptance criteria) | `references/frd-template.md` + `references/shared/gherkin-guide.md` |
| Defining cross-cutting standards, policies, or constraints | `references/standards-template.md` + `references/shared/ears-guide.md` |
| Documenting a system (context, containers, deployment) | `references/system-blueprint-template.md` + `references/shared/c4-standards.md` |
| Documenting a deployable unit (app or owned data store) | `references/container-blueprint-template.md` + `references/shared/c4-standards.md` |
| Documenting an atomic component of one container | `references/component-blueprint-template.md` + `references/shared/c4-standards.md` |
| Documenting a shared/published or composite unit (a module) | `references/module-blueprint-template.md` + `references/shared/c4-standards.md` |
| Designing how a specified feature will be built (deltas, AC coverage) | `references/feature-blueprint-template.md` + `references/shared/ac-traceability-standards.md` + `references/shared/c4-standards.md` |
| Understanding, filling, or gating a work order | `references/work-order-template.md` + `references/shared/work-order-field-guide.md` + `references/shared/work-order-definition-of-ready.md` |
| Reporting one actionable change, from a retro or from another project | `references/issue-template.md` |
| Preflighting a feature blueprint before review handoff | `references/shared/blueprint-readiness-checklist.md` |
| Specifying an interface's semantics, SLOs, and change policy | `references/contract-template.md` + `references/shared/c4-standards.md` |
| Mapping any element to a C4 level; diagram conventions | `references/shared/c4-standards.md` |
| Writing a feature or an acceptance criterion (Gherkin) | `references/shared/gherkin-guide.md` |
| Writing any other normative statement (EARS) | `references/shared/ears-guide.md` |
| Slugs, instruction blocks, versioning, placeholders, validation | `references/shared/template-authoring-standards.md` |
| Frontmatter properties, status vocabulary, element records | `references/shared/frontmatter-standards.md` |
| `[[slug]]` wikilinks, heading anchors, backticked paths | `references/shared/cross-document-references-standards.md` |

Every file above is one hop from here. Read the bundle up front rather
than following a pointer out of a template mid-authoring — a reference
reached at the second hop tends to get previewed rather than read.

---

## Anti-Patterns

`references/shared/anti-patterns-guide.md` names the failure modes of spec-first
authoring and what each one costs — wrong document type, untestable prose
requirements, duplicated content, checks reported without being run, and
feature-blueprint deltas that never get folded back into the owning
blueprints. Read it before setting a document to `ready-for-review`, and
when reviewing someone else's.

## Retro this skill

WHEN a run of `spec` struggled — you improvised around its guidance, a
command or script it prescribed misbehaved, or an instruction here
pointed you the wrong way — offer to run `retro-skill` on `spec` while
the session is fresh, to capture the friction as an evidence-backed
RFC. A clean run needs nothing; never run it unprompted.
