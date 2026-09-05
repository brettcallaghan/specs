---
title: swarm-retro's filing snippet uses `bd create --discovered-from`, which does not exist
skill: swarm-retro
kind: defect
priority: 2
reported_by: run agent-plugins-59ic (agent-plugins)
recurrence: 0
---

## What happened

`swarm-retro`'s instructions show filing a project-audience finding with
`bd create --discovered-from <id>`. That flag does not exist, so the snippet
fails as written on every run that files anything.

## Evidence

Measured on bd 1.2.2, in the repository `brettcallaghan/agent-plugins`:

```
$ bd create --title="zzz probe" --discovered-from=agent-plugins-rc14.1 --type=task
Error: unknown flag: --discovered-from
(exit 1, nothing created)
```

`bd create --help` does not list it as a flag. The string `discovered-from`
appears exactly once in the help output, inside the description of a
different flag:

```
--deps strings   Dependencies in format 'type:id' or 'id'
                 (e.g., 'discovered-from:bd-20,blocks:bd-15' or 'bd-20')
```

## The working form

The relationship is expressed through `--deps`, with the title positional:

```
bd create --title="..." --type=bug --priority=2 --deps 'discovered-from:<id>'
```

Verified: all three project-audience beads this run filed were created that
way, and the edge reads back correctly —

```
$ bd show agent-plugins-flvl
DISCOVERED FROM
  ◊ ✓ agent-plugins-rc14.2: Let a frontmatter citation carry an anchor ● P1
```

## Why it matters

The snippet is in the one step that files, so an agent following it literally
gets an `unknown flag` error at the moment it is turning a run's findings into
issues. This run's retro agent worked around it by discovering the `--deps`
form itself; a less careful one could drop the provenance edge rather than the
command, leaving filed beads with no link back to the bead that discovered
them.

## Suggested remedy

Replace the snippet with the `--deps 'discovered-from:<id>'` form. If the
skill asserts the flag anywhere in prose as well, correct that too — a
corrected snippet beside an uncorrected sentence is the failure mode the
citing project's own standards call out.
