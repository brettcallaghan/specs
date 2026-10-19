---
title: "ADR-017: The specs Plugin Provisions the Spec CLI Version It Pins"
description: A SessionStart hook in the specs plugin puts the Spec CLI release
  matching the plugin's version first on the session's PATH; skills call it
  directly and degrade to hand-checking only when provisioning failed
slug: adr-017
type: adr
status: draft
---

# ADR-017: The specs Plugin Provisions the Spec CLI Version It Pins

**Status:** draft
**Date:** 2026-08-05


## Context

Users install the `specs` plugin by cloning it through the
`brettcallaghan` marketplace (`brettcallaghan/agent-plugins`), and its
skills are Markdown. The Spec CLI is a Go binary that
`.github/workflows/release.yml` publishes through two channels: the Go
module proxy (`go install …/cli/specs@cli/vX.Y.Z`), and per-platform
release archives with a `SHA256SUMS` file.

The skills call the CLI throughout. `specs check` decides three of the
eight checks in `document-readiness-checklist.md`. `specs next-id` and
`specs slug-exists` answer questions that an agent would otherwise answer
by reading the corpus. Reading the corpus is expensive and unverifiable,
because an agent that scanned the corpus and one that assumed the answer
produce identical output.

The skills are written against one release of the CLI. They cite its
subcommands, flags, output shapes, and exit statuses. If a skill cites a
subcommand that an older binary lacks, the skill fails at the point of
use. One release cut versions both: `cli/vX.Y.Z` tags the CLI, and the
same cut writes `X.Y.Z` into `plugin/.claude-plugin/plugin.json`. The
plugin version therefore names exactly one CLI release, the one its skills
were written against.

A Markdown plugin and a compiled binary are distributed differently, so
they can't be one artifact. The decision is how the binary reaches a
session, and what the skills can assume about it.

## Decision

The plugin pins the CLI release whose version equals its own. Before any
skill runs, a `SessionStart` hook (`plugin/hooks/install-cli`) provisions
that release:

1. If `SPECS_CLI_PROVISION=off` is set, the hook does nothing. The skills
   run the user's own `specs`, if there is one.
2. If a `specs` on the PATH reports the pinned version from `--version`,
   the hook uses it as is. The hook downloads nothing and doesn't change
   the PATH.
3. Otherwise, the hook does the following:
   - Downloads the pinned release archive for the platform from the
     public release assets on github.com.
   - Verifies the archive against the release's `SHA256SUMS`.
   - Keeps the binary under `${CLAUDE_PLUGIN_DATA}/bin/<version>/`.
   - Prepends that directory to the session's PATH through
     `CLAUDE_ENV_FILE`.

   If that directory already holds the version, the hook doesn't
   download it again.

**On a user's machine, the hook shadows a different `specs` on the PATH
and never replaces it.** That binary stays installed and in use
everywhere outside the session. The hook adds one line to the session's
context that names the binary, its version, and the opt-out. In a cloud
session (`CLAUDE_CODE_REMOTE=true`), the container is disposable and the
pinned binary is the only correct choice. There, the hook also links the
pinned binary as `specs` into `/usr/local/bin`, replaces whatever is
there, and reports nothing.

The hook never fails a session. If provisioning fails (for example, no
release exists for the platform, the network is unavailable, or the
checksum doesn't match), the hook adds one line to the session's context.
That line says the CLI wasn't installed and what to do, and the hook exits
with status 0. If another `specs` is still first on the PATH, the line
names it and says not to run it, because its commands aren't the ones the
skills cite. The skills then use the hand-checking fallback.

**The checksum detects a corrupted download, but it doesn't authenticate
the binary.** `SHA256SUMS` comes from the same release as the archive, so
anyone who can replace one can replace both. Doing so requires write
access to this repository. The repository also serves the plugin, and so
the hook itself, so the same person could change the hook. The repository
is the root of trust, as it is for every skill the plugin ships. A
signature checked by a key or verifier that the plugin carries would rest
on the same root.

Skills call `specs` directly, in a single code path, and don't branch on
whether it's installed. The fallback is stated once, where the checks are
defined (`document-readiness-checklist.md`). Every check stays defined
independently of the tool, so an agent can run all of them by hand, and
the agent SHALL say which way it ran them. "`specs check` reports no
findings" and "I read the frontmatter and it looks right" are different
claims, and a reviewer can re-run only the first.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Hook provisions the pinned release; another version is shadowed for the session (selected) | The skills meet the binary they were written against, with no install step; a user's own install is never modified outside a cloud container | The plugin downloads and runs a binary at session start, which adds a trust surface that a Markdown plugin otherwise lacks. The hook limits that surface by fetching only the release that the plugin's own version names, from the repository that serves the plugin | N/A (selected) |
| Prerequisite installed by the user; skills single-path | The plugin stays Markdown only | A consumer who skips the install gets degraded checking on every document, and nothing reports a visible failure; nothing manages version skew between plugin and binary | Every skill instruction assumes a binary that nothing provides |
| Install on demand, when a skill first needs it | Nothing runs in a session that never touches a document | Every skill must carry the provisioning step before its first command, and the first command of a session pays for the download | The skills call `specs` in nearly every procedure. Installing at session start is simpler, and makes the binary an invariant instead of a step |
| Use whatever `specs` is on the PATH, and warn on a mismatch | Respects a deliberate local install | A mismatched binary fails at the point of use, partway through a procedure, on a subcommand or output shape that the skill assumed | The opt-out covers a deliberate choice; the default should be the version the skills were written against |
| Replace a different `specs` on the PATH | One binary everywhere | Modifies the user's own install, which other tools and projects might depend on | Outside the plugin's responsibility, except in a disposable container |
| Vendor per-platform binaries in the plugin directory | No network access at run time | Static binaries for six platforms in every clone, cloned again on every plugin update | Every consumer pays the cost so that offline consumers are served |

## Consequences

**Positive:**
- The skills meet the CLI release they were written against, so every
  subcommand they cite exists and its output has the shape they parse.
- A reviewer can re-run any check that an agent claims to have run, which
  is what gives the claim its value.
- The hook never modifies a user's own `specs`, and
  `SPECS_CLI_PROVISION=off` makes the skills run that binary.
- An update downloads once. A downgrade, or a second project on an
  earlier plugin version, downloads nothing.

**Negative:**
- The first session after an install or update depends on reaching
  github.com. Offline, the skills fall back to hand-checking and say so.
- A plugin version with no matching CLI release provisions nothing. The
  release workflow keeps the two versions aligned.
- The hook runs `specs --version` on whatever `specs` is first on the
  PATH, to decide whether it's already the pinned version.
- Nothing authenticates the binary independently of the repository that
  serves the plugin. The checksum only rejects a corrupted download.

## References

- [[adr-015]] — Go for the Spec CLI; the single static binary with no
  runtime makes provisioning a download and a rename
- [[container-spec-cli]] — the container this decision distributes,
  including the Coding Agent interaction it serves
- `plugin/hooks/install-cli` — the hook, with its lookup order and output
  contract
- `plugin/skills/spec/references/shared/spec-cli-reference.md` — the
  command surface the skills call
- `plugin/skills/spec/references/document-readiness-checklist.md` — where
  the split between mechanized and human checks is recorded
