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

The `specs` plugin is distributed by clone through the `brettcallaghan`
marketplace (`brettcallaghan/agent-plugins`) and its skills are markdown.
The Spec CLI is a Go binary, published by `.github/workflows/release.yml`
through two channels — the Go module proxy
(`go install …/cli/specs@cli/vX.Y.Z`) and per-platform release archives
with a `SHA256SUMS` file.

The skills call it throughout. `specs check` decides three of the eight
checks in `document-readiness-checklist.md`, and `specs next-id` and
`specs slug-exists` answer questions an agent would otherwise answer by
reading the corpus — expensively and unverifiably, since an agent that
scanned the corpus and one that assumed the answer produce identical
output.

The skills are written against one release of the CLI: they cite its
subcommands, flags, output shapes and exit statuses. A skill citing a
subcommand an older binary lacks fails at the point of use. One release
cut versions both: `cli/vX.Y.Z` tags the CLI, and the same cut writes
`X.Y.Z` into `plugin/.claude-plugin/plugin.json`. The plugin version
therefore names exactly one CLI release, the one its skills were written
against.

A markdown plugin and a compiled binary have different distribution
mechanics, so the two cannot be one artifact. The choice is how the binary
reaches a session, and what the skills may assume about it.

## Decision

The plugin pins the CLI release whose version equals its own, and a
`SessionStart` hook (`plugin/hooks/install-cli`) provisions it before any
skill runs:

1. With `SPECS_CLI_PROVISION=off` the hook does nothing; the user's own
   `specs`, if any, is what the skills run.
2. A `specs` on the PATH whose `--version` reports the pinned version is
   used as it is. Nothing is downloaded and the PATH is not changed.
3. Otherwise the hook downloads the pinned release archive for the
   platform from the public release assets on github.com, verifies it
   against the release's `SHA256SUMS`, keeps it under
   `${CLAUDE_PLUGIN_DATA}/bin/<version>/`, and prepends that directory to
   the session's PATH through `CLAUDE_ENV_FILE`. A version already kept
   there is not downloaded again.

**A different `specs` on the PATH is shadowed, never replaced.** On a
user's machine it stays installed and in use everywhere outside the
session; the hook adds one line to the session's context naming it, its
version, and the opt-out. In a cloud session (`CLAUDE_CODE_REMOTE=true`)
the container is disposable and the pinned binary is the only right
answer, so the hook also links it as `specs` into `/usr/local/bin`,
replacing whatever is there, and says nothing.

The hook never fails a session. A failure — no release for the platform,
no network, a checksum mismatch — is one line in the session's context
saying the CLI was not installed and what to do, and the hook exits 0.

Skills call `specs` directly, in a single code path. They do not branch on
whether it is installed. The fallback is stated once, where the checks
are defined (`document-readiness-checklist.md`): every check remains
defined independently of the tool, so all of them can be run by hand, and
the agent SHALL say which happened. "`specs check` reports no findings"
and "I read the frontmatter and it looks right" are different claims, and
only the first can be re-run by a reviewer.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Hook provisions the pinned release; another version is shadowed for the session (selected) | The skills meet the binary they were written against, with no install step; the user's own install is never modified | The plugin downloads and runs a binary at session start, a trust surface a markdown plugin otherwise lacks — bounded by fetching only the release that the plugin's own version names, checksum-verified | N/A — selected |
| Prerequisite installed by the user; skills single-path | Plugin stays markdown only | A consumer who skips the install gets degraded checking on every document, and nothing fails loudly; version skew between plugin and binary is unmanaged | Every skill instruction assumes a binary that nothing provides |
| Install on demand, when a skill first needs it | Nothing runs for a session that never touches a document | Every skill has to carry the provisioning step before its first command, and the first command of a session pays the download | The skills call `specs` in nearly every procedure; up front is simpler and makes the binary an invariant rather than a step |
| Use whatever `specs` is on the PATH, and warn on a mismatch | Respects a deliberate local install | A mismatched binary fails at the point of use, mid-procedure, on a subcommand or output shape the skill assumed | The opt-out covers a deliberate choice; the default should be the version the skills were written against |
| Replace a different `specs` on the PATH | One binary everywhere | Modifies the user's own install, which other tools and other projects may depend on | Out of the plugin's remit outside a disposable container |
| Vendor per-platform binaries in the plugin directory | No network at run time | Six platforms of static binary in every clone, re-cloned on every plugin update | Cost falls on every consumer to serve the ones offline |

## Consequences

**Positive:**
- The skills meet the CLI release they were written against, so a
  subcommand they cite exists and its output has the shape they parse.
- A check an agent claims to have run is re-runnable by a reviewer,
  which is what makes the claim worth anything.
- A user's own `specs` is never modified, and `SPECS_CLI_PROVISION=off`
  makes it the one the skills run.
- An update downloads once; a downgrade, or a second project on an
  earlier plugin version, downloads nothing.

**Negative:**
- The first session after an install or update depends on reaching
  github.com. Offline, the skills fall back to hand-checking and say so.
- A plugin version with no matching CLI release provisions nothing; the
  release workflow is what keeps the two in step.
- The hook runs `specs --version` on whatever `specs` is first on the
  PATH, to decide whether it is already the pinned one.

## References

- [[adr-015]] — Go for the Spec CLI; the single static binary with no
  runtime is what makes provisioning a download and a rename
- [[container-spec-cli]] — the container this distributes, including the
  Coding Agent interaction this decision serves
- `plugin/hooks/install-cli` — the hook, with its lookup order and output
  contract
- `plugin/skills/spec/references/shared/spec-cli-reference.md` — the
  command surface the skills call
- `plugin/skills/spec/references/document-readiness-checklist.md` — where
  the mechanized/human split is recorded
