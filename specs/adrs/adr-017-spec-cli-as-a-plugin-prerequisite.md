---
title: "ADR-017: The specs Plugin Provisions the Spec CLI Version It Pins"
description: A SessionStart hook in the specs plugin puts the Spec CLI release
  matching the plugin's version first on the session's PATH; skills call it
  directly and degrade to hand-checking only when provisioning failed
slug: adr-017
type: adr
status: draft
---

# ADR-017: The specs plugin provisions the Spec CLI version it pins

**Status:** draft
**Date:** 2026-08-05

## Context

The `specs` plugin is distributed by clone through the `brettcallaghan`
marketplace (`brettcallaghan/agent-plugins`), and its skills are Markdown.
The Spec CLI is a Go binary. `.github/workflows/release.yml` publishes it
through two channels:

- The Go module proxy (`go install …/cli/specs@cli/vX.Y.Z`).
- Per-platform release archives, with a `SHA256SUMS` file.

The skills call the CLI throughout. `specs check` decides three of the eight
checks in `document-readiness-checklist.md`. `specs next-id` and
`specs slug-exists` answer questions that an agent would otherwise answer by
reading the corpus. Reading the corpus is expensive and unverifiable,
because an agent that scanned the corpus and an agent that assumed the
answer produce identical output.

The skills are written against one release of the CLI: they cite its
subcommands, flags, output shapes, and exit statuses. A skill that cites a
subcommand that an older binary lacks fails at the point of use. One release
cut versions both artifacts. The `cli/vX.Y.Z` tag marks the CLI, and the
same cut writes `X.Y.Z` into `plugin/.claude-plugin/plugin.json`. The plugin
version therefore names exactly one CLI release: the one that its skills
were written against.

A Markdown plugin and a compiled binary have different distribution
mechanics, so the two can't be one artifact. The choice is how the binary
reaches a session, and what the skills can assume about it.

## Decision

The plugin pins the CLI release whose version equals its own. A
`SessionStart` hook (`plugin/hooks/install-cli`) provisions that release
before any skill runs:

1. If `SPECS_CLI_PROVISION=off` is set, the hook does nothing. The skills
   run the user's own `specs`, if any.
2. If a `specs` on the PATH reports the pinned version from `--version`, the
   hook uses it as is. The hook downloads nothing and doesn't change the
   PATH.
3. Otherwise, the hook downloads the pinned release archive for the platform
   from the public release assets on github.com. It verifies the archive
   against the release's `SHA256SUMS`, keeps it under
   `${CLAUDE_PLUGIN_DATA}/bin/<version>/`, and prepends that directory to
   the session's PATH through `CLAUDE_ENV_FILE`. The hook doesn't download a
   version that it already keeps there.

On a user's machine, the hook shadows a different `specs` on the PATH and
never replaces it. The user's `specs` stays installed and in use everywhere
outside the session. The hook adds one line to the session's context that
names the other `specs`, its version, and the opt-out. In a cloud session
(`CLAUDE_CODE_REMOTE=true`), the container is disposable and the pinned
binary is the only right answer. There, the hook also links the binary as
`specs` into `/usr/local/bin`, replacing whatever is there, and reports
nothing.

The hook never fails a session. A failure, such as no release for the
platform, no network, or a checksum mismatch, produces one line in the
session's context. That line says that the CLI wasn't installed and what to
do, and the hook exits 0. If another `specs` is still first on the PATH, the
line names it and says not to run it, because its commands aren't the ones
that the skills cite. The skills then take the hand-checking fallback.

**The checksum detects a corrupted download; it doesn't authenticate the
binary.** `SHA256SUMS` comes from the same release as the archive, so anyone
who can replace one can replace both. Doing so requires write access to this
repository, which also serves the plugin and therefore the hook itself.
Someone with that access could change the hook just as easily. The
repository is the trust root, as it is for every skill that the plugin
ships. A signature verified by a key or verifier that the plugin carries
would rest on the same root.

Skills call `specs` directly, in a single code path. They don't branch on
whether it's installed. The fallback is stated once, where the checks are
defined (`document-readiness-checklist.md`). Every check stays defined
independently of the tool, so you can run all of them by hand, and the agent
SHALL say which method it used. "`specs check` reports no findings" and "I
read the frontmatter and it looks right" are different claims, and a
reviewer can re-run only the first.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| The hook provisions the pinned release, and shadows another version for the session (selected). | The skills meet the binary that they were written against, with no install step. A user's own installation is never modified outside a cloud container. | The plugin downloads and runs a binary at session start. A Markdown plugin otherwise lacks this trust surface. The surface is bounded: the hook fetches only the release that the plugin's own version names, from the repository that serves the plugin. | N/A: selected. |
| The user installs the prerequisite, and skills use a single path. | The plugin stays Markdown only. | A consumer who skips the installation gets degraded checking on every document, and nothing fails loudly. Version skew between the plugin and the binary is unmanaged. | Every skill instruction assumes a binary that nothing provides. |
| Install on demand, when a skill first needs the CLI. | Nothing runs for a session that never touches a document. | Every skill has to carry the provisioning step before its first command, and the first command of a session pays for the download. | The skills call `specs` in nearly every procedure. Provisioning up front is simpler and makes the binary an invariant rather than a step. |
| Use whatever `specs` is on the PATH, and warn on a mismatch. | It respects a deliberate local installation. | A mismatched binary fails at the point of use, mid-procedure, on a subcommand or output shape that the skill assumed. | The opt-out covers a deliberate choice. The default should be the version that the skills were written against. |
| Replace a different `specs` on the PATH. | One binary everywhere. | It modifies the user's own installation, which other tools and other projects might depend on. | Outside a disposable container, this is out of the plugin's remit. |
| Vendor per-platform binaries in the plugin directory. | No network access at run time. | Every clone carries static binaries for six platforms, and every plugin update re-clones them. | Every consumer pays the cost in order to serve the consumers who work offline. |

## Consequences

**Positive:**
- The skills meet the CLI release that they were written against, so a
  subcommand that they cite exists, and its output has the shape that they
  parse.
- A reviewer can re-run a check that an agent claims to have run, which is
  what gives the claim its value.
- The plugin never modifies a user's own `specs`, and
  `SPECS_CLI_PROVISION=off` makes it the one that the skills run.
- An update downloads once. A downgrade, or a second project on an earlier
  plugin version, downloads nothing.

**Negative:**
- The first session after an installation or update depends on reaching
  github.com. Offline, the skills fall back to hand-checking and say so.
- A plugin version with no matching CLI release provisions nothing. The
  release workflow keeps the two in step.
- To decide whether the first `specs` on the PATH is already the pinned one,
  the hook runs `specs --version` on it, whatever it is.
- Nothing authenticates the binary independently of the repository that
  serves the plugin. The checksum only rejects a corrupted download.

## References

- [[adr-015]]: Go for the Spec CLI. The single static binary with no runtime
  is what makes provisioning a download and a rename.
- [[container-spec-cli]]: the container that this decision distributes,
  including the Coding Agent interaction that this decision serves.
- `plugin/hooks/install-cli`: the hook, with its lookup order and output
  contract.
- `plugin/skills/spec/references/shared/spec-cli-reference.md`: the command
  surface that the skills call.
- `plugin/skills/spec/references/document-readiness-checklist.md`: where the
  split between mechanized and human checks is recorded.
