# Changelog

## 0.3.0

**Blueprints that state rules, not code.**

This release reframes architecture blueprints around what an agent can't read from the code: boundaries, invariants, design patterns by C4 level and extension points. Templates now name everything their document type needs, and the C4 model's "drawing" is renamed "relationship statement", which changes one rule id and one output field.

### 💥 Breaking changes

- **cli:** The C4 model calls a relationship's evidence its statements: rule `c4/drawing-conflict` is now `c4/statement-conflict`, and a relationship in `specs c4` output lists `statements` instead of `drawn_in`. ([#92](https://github.com/brettcallaghan/specs/pull/92), [399df6a](https://github.com/brettcallaghan/specs/commit/399df6a8026a5ba38831a64cc2e2395220821a1d))

  The finding tag `c4-drawing-conflict` is now `c4-statement-conflict`, and the C4 model's `schema_version` moves from `0.1` to `0.2`.

  Migration: Replace `c4/drawing-conflict` with `c4/statement-conflict` wherever you select or filter rules, and `c4-drawing-conflict` with `c4-statement-conflict` for finding tags. Read a relationship's `statements` instead of `drawn_in`, and accept C4 model `schema_version` `0.2`.

### 🚀 Features

- **skills/spec:** Component Blueprints now state the rules an agent can't read from the code: what the component owns, hides and does not own, the patterns it follows, the invariants every change keeps, and where it is designed to grow. ([#85](https://github.com/brettcallaghan/specs/pull/85), [7185b05](https://github.com/brettcallaghan/specs/commit/7185b0518084c9ff995bfc4f575f5bc99e2b693d))

  The Component, Module, Container, System and Feature Blueprint templates are revised to match. Responsibility, Interface and Design use subheadings, verification gates split into Verify always and Verify conditionally, and the System Blueprint gains a Design section. Blueprints no longer restate code facts or the FRD's behavior and acceptance criteria. A new blueprint content standard holds the rules and ends with an audit the `spec` skill runs before review.

- **skills/spec:** The `spec` skill ships design-principles guides for every C4 level, and a blueprint names each pattern it adopts with its progenitor and the invariant it implies. ([#85](https://github.com/brettcallaghan/specs/pull/85), [7185b05](https://github.com/brettcallaghan/specs/commit/7185b0518084c9ff995bfc4f575f5bc99e2b693d), [#90](https://github.com/brettcallaghan/specs/pull/90), [12610ce](https://github.com/brettcallaghan/specs/commit/12610cecaddf206c1b655680669bbb657236b95c))

  There is one universal guide and one each for systems, containers and components. Each is a curated list that gives every entry one progenitor and the questions it answers, and the skill loads them only when an element's design changes.

- **skills/spec:** Each template now names everything its document type needs under "Load before authoring", so the `spec` skill loads less and reaches what it needs in one step. ([#87](https://github.com/brettcallaghan/specs/pull/87), [affa3a2](https://github.com/brettcallaghan/specs/commit/affa3a208fb4f51771208f14356dfa8a76a73766))

  The skill always loads only the `tech-writer` skill, the Specs CLI reference and the EARS guide. It also gains a generated reference describing every `specs check` rule, a standalone C4 guide beside the C4 standard, a standalone Gherkin guide, and a spec-authoring standard for line length, titles, slugs, prefixes and filenames. The readiness checklists are folded into the templates and the Feature Blueprint template.

- **cli:** Container Blueprints accept optional `source_paths` and `test_paths`, and `groom-beads` derives a work order's changes from them. ([#85](https://github.com/brettcallaghan/specs/pull/85), [7185b05](https://github.com/brettcallaghan/specs/commit/7185b0518084c9ff995bfc4f575f5bc99e2b693d))
- **skills/groom-feature-blueprint:** `groom-feature-blueprint` rejects a delta that restates FRD behavior, flags a delta that bypasses a declared extension point, and assigns verification gates by tier. ([#85](https://github.com/brettcallaghan/specs/pull/85), [7185b05](https://github.com/brettcallaghan/specs/commit/7185b0518084c9ff995bfc4f575f5bc99e2b693d))
- **skills/tech-writer:** `tech-writer` keeps prose stable and terms plain: it rules out counts, numbered citations and volatile figures, and it adds a glossary term only when two or more documents use it and you approve it. ([#85](https://github.com/brettcallaghan/specs/pull/85), [7185b05](https://github.com/brettcallaghan/specs/commit/7185b0518084c9ff995bfc4f575f5bc99e2b693d), [#90](https://github.com/brettcallaghan/specs/pull/90), [12610ce](https://github.com/brettcallaghan/specs/commit/12610cecaddf206c1b655680669bbb657236b95c), [#94](https://github.com/brettcallaghan/specs/pull/94), [b26df3b](https://github.com/brettcallaghan/specs/commit/b26df3b2b9f04777cfe1cd7f7cef1793feebd592))

  It may also make clarity edits to requirement statements without changing their modal, actor, trigger, condition, scope or EARS pattern, and evidence-backed factual fixes in blueprints. Terminology rules follow ASD-STE100. The skill is about a quarter smaller, with the human-voice guidance moved into its reader profiles.

- **skills/spec:** The `spec` skill has a copy-edit workflow for revising many documents at once: editors grouped by host blueprint, a diff-only reviewer for each editor, and new glossary terms shown to you for approval. ([#90](https://github.com/brettcallaghan/specs/pull/90), [12610ce](https://github.com/brettcallaghan/specs/commit/12610cecaddf206c1b655680669bbb657236b95c), [#94](https://github.com/brettcallaghan/specs/pull/94), [b26df3b](https://github.com/brettcallaghan/specs/commit/b26df3b2b9f04777cfe1cd7f7cef1793feebd592))

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
