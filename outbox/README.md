# Issue outbox

**Issues this project reports about skills it does not own.**

A run's retro produces two kinds of finding. Ones about *this project* —
its specs, its standards, its checks — become beads in this tracker. Ones
about *a skill this project used* belong to whoever vends that skill, and
land here instead.

```
outbox/<plugin>/<skill>/<YYYY-MM-DD>-<slug>.md
```

`<plugin>` is the plugin that vends the skill — the namespace it is invoked
under (`specs:retro-skill` → `specs`, `swarm:swarm-retro` → `swarm`). Each
vendor sweeps only its own folder, so one plugin's backlog never reaches
another's.

One file per issue, in the issue template's format — the same template a
project-audience issue uses, since only the destination differs.

**Nothing here is ever filed in this project's own tracker.** A backlog of
proposals its owners cannot act on teaches them to ignore the channel, and
that is the failure the audience split exists to prevent.

**This directory is tracked, not ignored.** The issues are a run's
deliverable and have to survive a reclaimed container.

**It belongs to this project.** Emptying it, forwarding it, or ignoring it
entirely are all its owners' call — a run appends and never prunes.

## In this repository

**We vend the skills we run**, so we are our own downstream: `triage-issue`
sweeps `outbox/specs/` into our own tracker rather than anyone forwarding
anything. The mechanism is identical; the hop is just short.

`outbox/swarm/` is the `swarm` plugin's (brettcallaghan/swarm), not ours:
forward it there rather than triaging it here.
