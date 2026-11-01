# Document types

The `spec` skill writes each of the following document types from a template
that it includes. The set of types is small and closed.

## Product documents

Product documents describe *what* to build. Product owns them.

| Document | Captures | Example | Template |
|:---|:---|:---|:---|
| **FRD** (feature requirements document) | The feature only: the user story (why) and the acceptance criteria (what), written as tagged Gherkin. An FRD never describes how to build the feature. | [FRD: Spec Corpus Index](../../specs/features/spec-index/spec-index-frd.md) | [`frd-template.md`](../../plugin/skills/spec/references/frd-template.md) |

Every feature starts with an FRD. Each FRD describes one feature that delivers
value that you can test on its own. A sub-feature gets its own FRD instead of
being nested inside another one.

## Architecture and engineering documents

Architecture and engineering documents describe *how* to build the system and
the constraints that the design must respect. These documents use the
[C4 model](https://c4model.com/), which describes software at the system,
container, and component levels.

| Document | Captures | Example | Template |
|:---|:---|:---|:---|
| **System Blueprint** | A system's context, its containers, its public contracts, and its deployment (C4 levels 1 and 2). | - | [`system-blueprint-template.md`](../../plugin/skills/spec/references/system-blueprint-template.md) |
| **Container Blueprint** | One deployable unit: its components, the contracts that it provides and consumes, the data that it owns, and change guidance. | [Container: Spec CLI](../../specs/containers/spec-cli/spec-cli-container.md) | [`container-blueprint-template.md`](../../plugin/skills/spec/references/container-blueprint-template.md) |
| **Component Blueprint** | An atomic building block of one container or module: what it owns, the design decision it hides, its interface guarantees, its invariants, and where it's designed to grow. | - | [`component-blueprint-template.md`](../../plugin/skills/spec/references/component-blueprint-template.md) |
| **Module Blueprint** | A unit that crosses a boundary of independent change, because it's shared or published across containers or because it's composite. Includes versioning and change rules. | [Module: Document Core](../../specs/modules/document-core/document-core-module.md) | [`module-blueprint-template.md`](../../plugin/skills/spec/references/module-blueprint-template.md) |
| **ADR** (architecture decision record) | One significant decision: its context, the alternatives, the choice, and the consequences. | [ADR-015: Go for the Spec CLI](../../specs/adrs/adr-015-go-for-the-spec-cli.md) | [`adr-template.md`](../../plugin/skills/spec/references/adr-template.md) |
| **Standards** | Cross-cutting standards, policies, invariants, and constraints. | [Testing Standards](../../specs/standards/testing-standards.md) | [`standards-template.md`](../../plugin/skills/spec/references/standards-template.md) |
| **Contract** | The meaning of an interface that's used across a boundary, such as an API, an event stream, or a data product. | [Contract: Specs Configuration](../../specs/contracts/specs-configuration-contract.md) | [`contract-template.md`](../../plugin/skills/spec/references/contract-template.md) |

An architecture blueprint states the rules that the code can't show and
links the code for everything else. To learn why, see
[How architecture blueprints work](../explanation/architecture-blueprints.md).

These documents cover systems, containers, modules, and components, and they
stop there. Code is the last level of the C4 hierarchy, and only Work Orders,
which are temporary, can refer to it.

## Documents that connect requirements to code

The following two document types connect a requirement to the code that
implements it. Both are temporary. You can keep them for audit purposes or
delete them after the change merges.

| Document | Captures | Example | Template |
|:---|:---|:---|:---|
| **Feature Blueprint** | How one FRD's feature maps onto your architecture: the components that it creates, updates, and deletes, plus contracts, sequencing, acceptance criteria coverage, and a definition of done. | [Blueprint: Spec Corpus Index](../../specs/features/spec-index/spec-index-blueprint.md) | [`feature-blueprint-template.md`](../../plugin/skills/spec/references/feature-blueprint-template.md) |
| **Work Order** | One atomic unit of work, similar to a development task. It contains everything that a new agent with no context needs to finish one job, and it connects component changes to code changes. | Stored as beads in this repository's tracker | [`work-order-template.md`](../../plugin/skills/spec/references/work-order-template.md) |

A Feature Blueprint refers to exactly one FRD and to every System, Container,
Component, or Module Blueprint that the change affects.

Nobody writes a Work Order by hand. Grooming generates each one to meet the
Definition of Ready. If you use [beads](https://github.com/gastownhall/beads),
each Work Order becomes one bead.

## Where each document type lives

The `spec` skill writes each document under the spec root that `specs.toml`
names, in the following layout by default:

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

## Related

- [How the document model works](../explanation/document-model.md): C4
  levels, where Standards and ADRs apply, and which documents are temporary.
