# Specs: spec-first development

<!-- coverage:badge:start -->
[![coverage](https://img.shields.io/badge/cli%20coverage-91.4%25-brightgreen)](docs/code-coverage.md)
<!-- coverage:badge:end -->

**Agency for the agents.**

Specs is a Claude Code plugin for spec-first development with agents. Its
skills write feature requirements, C4 architecture blueprints, architecture
decision records (ADRs), and Standards. They then turn an approved design into
Work Orders that agents can run without asking questions. Its `specs` CLI
checks, numbers, and traces those documents, so a reviewer can rerun any check
the CLI made.

Use it when agents build your software and you need them to make the right
call when the plan meets reality, instead of guessing or stopping to ask.

## What's included

- **Skills:** `spec`, `groom-feature-blueprint`, `groom-beads`, `retro-spec`,
  `retro-skill`, and `tech-writer`. See [Skills](docs/reference/skills.md).
- **Document types:** a small, closed set of ten, from the FRD to the Work
  Order. See [Document types](docs/reference/document-types.md).
- **The `specs` CLI:** eleven deterministic commands that index, check, and
  query a spec corpus. See [The `specs` CLI](docs/reference/cli.md).

## Get started

1. In Claude Code, add the marketplace and install the plugin:

   ```
   /plugin marketplace add brettcallaghan/agent-plugins
   /plugin install specs@brettcallaghan
   ```

   When a session starts, the plugin puts the matching `specs` CLI on the
   session's `PATH`, and downloads it the first time.

2. Add a `specs.toml` file at the root of your repository to tell the skills
   and the CLI where your specs live:

   ```toml
   [workspace]
   name = "my-project"
   root = "specs"

   [profile]
   name = "specs"
   ```

3. Describe a feature to the `spec` skill. For example:

   ```
   /spec Users can export a report as a CSV file.
   ```

   The skill chooses the document type, writes it under `specs/` from a
   template, and runs its readiness checklist before it hands the document to
   you for review.

For a guided first run, follow
[Specify your first feature](docs/tutorials/specify-your-first-feature.md).

## Documentation

| If you want to | Read |
|:---|:---|
| Learn by writing a first spec | [Specify your first feature](docs/tutorials/specify-your-first-feature.md) |
| Install or configure the plugin and the CLI | [Install the plugin and the CLI](docs/how-to/install.md) |
| Take a feature from requirements to Work Orders | [Plan a feature for agents](docs/how-to/plan-a-feature.md) |
| Find out what went wrong in a phase or a skill | [Run a retrospective](docs/how-to/run-a-retrospective.md) |
| Look up a document type, skill, command, or path | [Reference](docs/README.md#reference) |
| Understand why specs, and how the pieces fit | [Explanation](docs/README.md#explanation) |

This repository specifies itself: the FRDs, ADRs, blueprints, and Standards in
[`specs/`](specs/) are the ones that produced the CLI in [`cli/`](cli/), so
they're also worked examples of every document type.

## License

This project is licensed under the [MIT License](LICENSE).
