# Changelog

## 0.2.0

**Specs now write in a developer documentation style, with a new tech-writer skill.**

This release gives the plugin a writing style. A new `tech-writer` skill drafts, revises and audits documentation, and the `spec`, grooming and retro skills use it for everything they write.

### 🚀 Features

- **skills/tech-writer:** The new `tech-writer` skill drafts, revises and audits documentation in the project's style guide, which is the Google developer documentation style guide unless `AGENTS.md` or `CLAUDE.md` names another. ([#76](https://github.com/brettcallaghan/specs/pull/76), [c5a930b](https://github.com/brettcallaghan/specs/commit/c5a930b291b100e84d252d0555b3039fcad61240))

  It adds what the style guide leaves open: a reader profile that sets each document's voice and tone, sentence-case headings, timeless prose that describes how things are, and a rule that a rewrite never changes front matter, code, Gherkin, EARS or SHALL statements, ids or links. Call it directly on any document.

- **skills:** The `spec` skill writes every document with `tech-writer`, and the grooming and retro skills use it for what they write. ([#76](https://github.com/brettcallaghan/specs/pull/76), [c5a930b](https://github.com/brettcallaghan/specs/commit/c5a930b291b100e84d252d0555b3039fcad61240))

  Every spec template's header comment carries its document type's reader profile.

### 🪲 Bugs

- **skills:** Skills can be uploaded to claude.ai: `SKILL.md` files no longer carry a `version` key, which claude.ai rejects. ([#76](https://github.com/brettcallaghan/specs/pull/76), [c5a930b](https://github.com/brettcallaghan/specs/commit/c5a930b291b100e84d252d0555b3039fcad61240))

  The plugin's release version identifies the skills and templates instead.

## 0.1.0

**The project is now specs, installed from the agent-plugins marketplace, and the plugin provisions its own CLI.**

This release finishes the move to brettcallaghan/specs. The Go module path, the marketplace you install from and the retro workflow all change, so read the breaking changes before upgrading. The plugin now puts the matching specs CLI on the session's PATH by itself.

### 💥 Breaking changes

- **cli:** The Go module is now github.com/brettcallaghan/specs/cli, so the CLI installs from the specs repository instead of agent-plugins. ([#65](https://github.com/brettcallaghan/specs/pull/65), [ace86fe](https://github.com/brettcallaghan/specs/commit/ace86fe22d256f095efa0648b2e1ce17f9775863))

  Migration: Install with `go install github.com/brettcallaghan/specs/cli/specs@latest`, and update any import of github.com/brettcallaghan/agent-plugins/cli to the new path.

- **plugin:** The specs plugin is installed from the brettcallaghan/agent-plugins marketplace, which serves each published release; this repository no longer declares a marketplace of its own. ([#69](https://github.com/brettcallaghan/specs/pull/69), [2ca3220](https://github.com/brettcallaghan/specs/commit/2ca322056688c634f672c3d63f098943b79e9442))

  Migration: Run `/plugin marketplace add brettcallaghan/agent-plugins`, then `/plugin install specs@brettcallaghan`, and remove the old marketplace that pointed at this repository.

- **cli:** The `rfc` document type is removed: the CLI no longer recognises RFC frontmatter, and the RFC templates are gone in favour of the issue template, which also covers bug reports. ([#66](https://github.com/brettcallaghan/specs/pull/66), [bca3a6b](https://github.com/brettcallaghan/specs/commit/bca3a6be026ea199636b2770e8070c28f7765e2b))

  Migration: Move or delete any documents declaring `type: rfc`; file what they propose as issues (a bead, a GitHub issue, or a hand-over file) using the issue template.

### 🚀 Features

- **plugin:** When a session starts, the plugin puts the specs CLI release matching its own version first on the PATH, downloading and verifying it the first time and reusing it after. ([#74](https://github.com/brettcallaghan/specs/pull/74), [7f7c0f7](https://github.com/brettcallaghan/specs/commit/7f7c0f74df6cad704658610ec62d67e7809a3611))

  A different `specs` already installed is left alone and shadowed for that session only. Set `SPECS_CLI_PROVISION=off` to run your own `specs` instead.

- **skills:** retro-skill and retro-spec now report in the session, ending with a stack-ranked list of candidate issues for you to choose from, and commit no files to the repository. ([#66](https://github.com/brettcallaghan/specs/pull/66), [bca3a6b](https://github.com/brettcallaghan/specs/commit/bca3a6be026ea199636b2770e8070c28f7765e2b))

### 🪲 Bugs

- **groom-beads:** bd-health.sh validates only real epics, instead of running every nested dependency id through `bd swarm validate` and reporting failures that were not there. ([#67](https://github.com/brettcallaghan/specs/pull/67), [1d7f5f5](https://github.com/brettcallaghan/specs/commit/1d7f5f5086a06b013ebae70e1653c0a89a77f527))

## 0.0.0
