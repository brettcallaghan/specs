# Repository layout

This repository contains the following directories:

| Path | Contents |
|:---|:---|
| `plugin/skills/` | The six skills, each with its templates, references, and scripts. |
| `plugin/hooks/` | The session hook that installs the `specs` CLI. |
| `specs/` | This repository's own specs, written with the framework. |
| `cli/` | The `specs` CLI, written in Go. |
| `schemas/` | [Strata](https://github.com/strata-lang/strata) schema sources for frontmatter, findings, the index, the profile, the C4 model, features, and configuration. |
| `scripts/` | Repository checks, the coverage report, release builds, and the skill installer. |
| `docs/` | User documentation, research notes, and temporary plans for changes to this repository. |

The FRDs, ADRs, blueprints, and Standards in `specs/` are the ones that
produced the CLI in `cli/`. See
[This repository specifies itself](../explanation/document-model.md#this-repository-specifies-itself).
