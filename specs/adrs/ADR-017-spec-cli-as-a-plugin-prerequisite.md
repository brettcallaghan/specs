---
title: "ADR-017: The Spec CLI Is a Prerequisite of the sf Plugin, Not a Payload"
description: The sf plugin stays pure markdown and names the spec binary as
  a prerequisite; skills call it directly and degrade to hand-checking
  rather than branching on its absence
slug: adr-017
type: adr
status: draft
updated_at: 2026-08-05T00:00:00Z
---

# ADR-017: The Spec CLI Is a Prerequisite of the sf Plugin, Not a Payload

**Status:** draft
**Date:** 2026-08-05

## Context

The `sf` plugin is distributed by clone through
`.claude-plugin/marketplace.json` and consists entirely of markdown:
six skills and their bundled references. The Spec CLI is a Go binary,
published by `.github/workflows/release.yml` through two channels — the
Go module proxy (`go install …/cli/spec@cli/vX.Y.Z`) and per-platform
release archives.

The skills now want to call it. `spec check` decides three of the eight
checks in `document-readiness-checklist.md`, and `spec next-id` and
`spec slug-exists` answer questions the skill previously told an agent to
answer by reading the corpus. Those instructions were both expensive and
unverifiable: an agent that scanned the corpus and one that assumed the
answer produce identical output.

A markdown plugin and a compiled binary have different distribution
mechanics, so the two cannot simply be shipped as one artifact. The
choice is what the skills may assume, and nothing about it is
reversible cheaply — every skill's instructions are written against the
answer.

## Decision

The plugin ships markdown only. The Spec CLI is a **named prerequisite**,
installed by the adopting user or project through either published
channel.

Skills call `spec` directly, in a single code path. They do not branch on
whether it is installed, and they do not carry a parallel set of
instructions for its absence.

The fallback is stated once, where the checks are already defined
(`document-readiness-checklist.md`), and it is a fallback rather than a
second procedure: every check remains defined independently of the tool,
so all of them can be run by hand, and the agent SHALL say which happened.
"`spec check` reports no findings" and "I read the frontmatter and it
looks right" are different claims, and only the first can be re-run by a
reviewer.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Prerequisite; single-path skills (selected) | The skill body shrinks, because there is one procedure instead of two; every instruction is a command a reviewer can re-run | A user who has not installed it gets degraded checking, and must be told so | N/A — selected |
| Optional accelerator; skills branch on `command -v spec` | Nothing to install; works for every consumer today | The skills get *longer*: two procedures for the same eight checks, and the hand path is the one that was already too vague to follow. Directly defeats the reason for building the CLI | Adding a tool that makes the instructions longer is a net loss |
| Plugin provisions the binary via a `SessionStart` hook | Zero-touch; the skill can assume the binary unconditionally | A plugin advertised as markdown would download and execute a binary at session start. That is a trust and support surface the plugin does not currently have, and it is orthogonal to what the skills assume | Deferred, not rejected — see Consequences |
| Vendor per-platform binaries in the plugin directory | No install step, no network at run time | Six platforms of static binary in every clone of a markdown plugin, re-cloned on every plugin update | Cost falls on every consumer to serve the ones who skip an install |

The third option is worth separating from this decision precisely
because it does not change it: whether the user installs the binary by
hand or a hook installs it for them, the skills are written the same way.
Provisioning can therefore land later without reopening this ADR.

## Consequences

**Positive:**
- The skills' verification guidance collapses to a command plus the
  checks the command does not make. The mechanized rules are named in
  one table with who decides each.
- A check an agent claims to have run is re-runnable by a reviewer,
  which is what makes the claim worth anything.
- The plugin stays clonable, inspectable markdown with no install-time
  execution.
- The two publication channels already exist and need nothing further.

**Negative:**
- A consumer who installs the plugin and not the CLI gets weaker
  checking than the skill describes, and depends on the agent saying so.
- The prerequisite is stated in the skill's references rather than
  enforced by anything; nothing fails loudly at install time.
- Version skew between the plugin and the binary is unmanaged. A skill
  citing a subcommand an older binary does not have fails at the point
  of use, and no compatibility floor is declared today.

**Follow-on work, deliberately not decided here:**
- Whether a `SessionStart` hook in the plugin provisions the binary, and
  if so with what platform detection, checksum verification, and opt-out.
- Whether the plugin declares a minimum CLI version, and how a skill
  states the requirement it depends on.

## References

- [[adr-015]] — Go for the Spec CLI; the single static binary with no
  runtime is what makes an install-it-yourself prerequisite cheap
- [[container-spec-cli]] — the container this distributes, including the
  Coding Agent interaction this decision serves
- `plugin/skills/spec/references/spec-cli.md` — the command surface the
  skills call, and the install instructions this decision points users at
- `plugin/skills/spec/references/document-readiness-checklist.md` — where
  the mechanized/human split is recorded
