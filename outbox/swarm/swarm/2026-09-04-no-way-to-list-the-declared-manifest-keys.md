---
title: "Add a way to list the declared manifest keys"
description: There is no command that names them, so a key has to be discovered by running `manifest validate` against an empty manifest and reading key names out of its failure lines.
slug: issue-no-way-to-list-the-declared-manifest-keys
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#6"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

The manifest's key set is closed — `swarm manifest set` refuses anything not
declared — and nothing on the surface says what the declared keys are. The
refusal message says a key is not declared without naming the ones that are, so
the only way to learn the set is to run `manifest validate` against an empty
manifest and read the key names out of its failure lines. That is a discovery
procedure built on an error path, and it is how `target` was found after
`target_epic` was guessed. The cost is a coordinator that guesses key names at
every stage boundary and, where a guess is wrong, either burns attempts or
records nothing.

## Evidence

- `d7w5.2.1#6` — "there is no way to list the declared manifest keys. I
  discovered 'target' (having guessed 'target_epic') by running `swarm manifest
  validate` against an empty manifest and reading the key names out of its
  failure lines. A `swarm manifest keys` subcommand, or naming the valid keys in
  the 'not a declared manifest key' error, would remove the guess."

## Proposed change

Add `swarm manifest keys`, listing every declared key with its accepted shape
and the stage at which it becomes required, rendered per the surface's own
output conventions. The "not a declared manifest key" rejection names the
command, and where the given key is close to a declared one, names that too.

Done is: `swarm manifest keys` lists the full declared set including `target`;
and `swarm manifest set --key target_epic ...` rejects with a message pointing
at `swarm manifest keys` and at `target`.
