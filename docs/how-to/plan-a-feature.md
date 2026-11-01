# Plan a feature for agents

This guide takes one feature from requirements to a plan of Work Orders that
agents can run. Each step uses a skill from the Specs plugin. For the order
the steps run in and who reviews what, see [Workflow](../explanation/workflow.md).

Before you start, install the plugin and add a `specs.toml`. See
[Install the plugin and the CLI](install.md).

## Write the FRD

1. Describe the feature to the `spec` skill:

   ```
   /spec <what the feature does, and for whom>
   ```

   The skill writes an FRD: the user story and the acceptance criteria, as
   tagged Gherkin. Each FRD describes one feature that delivers value you can
   test on its own. A sub-feature gets its own FRD instead of being nested in
   another one.

2. Review the acceptance criteria, and then set the FRD's `status` to
   `approved`.

## Document the architecture the feature touches

A Feature Blueprint maps the FRD onto architecture that's already written
down. If the systems, containers, components, or modules that the feature
changes have no blueprint yet, write them first. This track can run alongside
the FRD.

1. Describe each one to the `spec` skill. It chooses among System, Container,
   Component, and Module Blueprints.
2. Record each significant decision the design depends on as an ADR, each
   cross-cutting rule as Standards, and each interface used across a boundary
   as a Contract.

For what each type holds, see [Document types](../reference/document-types.md).

## Write the Feature Blueprint

1. Ask the `spec` skill for a Feature Blueprint for the approved FRD.

   The blueprint names the components the feature creates, updates, and
   deletes, plus contracts, sequencing, acceptance criteria coverage, and a
   definition of done. While it writes the blueprint, the skill surveys your
   issue tracker for existing issues the feature subsumes or depends on.

   Before it sets the blueprint to `ready-for-review`, the skill runs the
   readiness checks in the Feature Blueprint template, the same checks that
   grooming enforces. It then flags two status changes for you to make.

2. Review the blueprint. Answer or remove everything in its Open questions
   section.
3. Set the blueprint's `status` to `approved`. Make sure that the FRD's
   `status` is `approved` or `active`.

## Groom the blueprint into Work Orders

1. Run the `groom-feature-blueprint` skill on the approved blueprint:

   ```
   /groom-feature-blueprint <path to the Feature Blueprint>
   ```

2. Read the result. It's one of the following:

   - **Grooming can't start:** the FRD or the blueprint isn't approved, or the
     blueprint still has open questions. The skill says which, and changes
     nothing. Fix the status or the questions, and run it again.
   - **The design has gaps:** the skill writes each gap into the blueprint's
     Open questions section, sets the blueprint back to `draft`, and stops.
     It never picks an answer for you. Answer the questions, approve the
     blueprint again, and run it again.
   - **A plan:** the skill breaks the blueprint's changes into a graph of Work
     Orders and sets the blueprint to `in-progress`. Every Work Order traces
     to a change, a definition-of-done entry, or a sequencing constraint in
     the blueprint.

Where the plan goes depends on your project:

- **If you use [beads](https://github.com/gastownhall/beads):** each Work
  Order becomes one bead. The skill then runs `groom-beads`, which assigns
  each bead a capability tier, tags what it changes, and checks the graph
  against the Definition of Ready.
- **Otherwise:** the skill writes a task-file directory,
  `plans/<feature-slug>/`.

## Implement the Work Orders

Hand the Work Orders to agents. Each Work Order updates the blueprint that
owns its component along with the code and the tests, so your architecture
specs stay current after the change merges.

The last Work Order checks the definition of done: every architecture change
reached the blueprint that owns it, and every acceptance criterion is cited.
It then sets the Feature Blueprint's `status` to `retired` and runs
`retro-spec` over the implementation. The retrospective files nothing on its
own. It records its candidate issues on that Work Order for you to choose
from. To run one yourself, see [Run a retrospective](run-a-retrospective.md).

## Groom beads that didn't come from a blueprint

To bring existing beads up to the Definition of Ready, run `groom-beads` on
them directly:

```
/groom-beads
```

The skill reports each bead as ready or not ready, with the gaps that keep a
bead from being ready.
