---
title: "`swarm manifest set` type errors state the accepted shape"
description: The error names the Go type and not the value it wants, costing three rejected writes in four attempts and a read of internal/swarmmodel to recover.
slug: issue-manifest-set-type-errors-name-the-go-type-not-the-shape
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.2.1#5"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm manifest set` rejects a badly-shaped value with an error naming the **Go
type** it failed to unmarshal into, which is not a description of what the
command accepts. An agent holding only the CLI has nowhere to go from
`[]swarmmodel.Exclusion` except into the package source — which is exactly what
happened here, at a cost of three rejected writes in four attempts on a
single-key write. The atomic refusal is right and should stay; the error text is
what is short. The cost compounds because a coordinator writes the manifest at
every stage boundary, so the same guess is paid repeatedly, and an agent that
gives up guessing writes nothing rather than writing the key.

## Evidence

- `d7w5.2.1#5` — "`swarm manifest set` type errors name the Go type but not the
  shape it accepts, costing 3 rejected writes on 4 attempts. 'excluded takes
  []swarmmodel.Exclusion, and the value given is not one: json: cannot unmarshal
  string into Go value of type map[string]json.RawMessage' does not say the
  element is {id, reason}; I had to read internal/swarmmodel/swarmmodel.go. Same
  for AuthSource's two-member enum. The atomic refusal ('storing none of it') is
  exactly right; only the error text is short."

## Proposed change

Every declared manifest key carries a one-line description of its accepted JSON
shape, and `swarm manifest set` prints it on a type rejection alongside a
minimal valid example — for `excluded`, `[{"id": "<bead id>", "reason":
"<why>"}]`; for an enum-valued field, the members it admits. The Go type name may
stay as a suffix; it may not be the whole message.

Done is: `swarm manifest set --run <epic> --key excluded --value '"nope"'` prints
the accepted element shape and one valid example, and a reader can write the key
correctly on the next attempt without opening `internal/swarmmodel`.
