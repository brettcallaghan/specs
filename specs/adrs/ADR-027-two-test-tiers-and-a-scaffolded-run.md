---
title: "ADR-027: Two Test Tiers, and a Scaffolded Run for the Upper One"
description: Unit tests assert the plan a command builds and replay recorded
  tool output; end-to-end tests execute the compiled binary against a real
  repository and tracker that a scaffold command creates
slug: adr-027
type: adr
status: ready-for-review
updated_at: 2026-08-28T00:00:00Z
---

# ADR-027: Two Test Tiers, and a Scaffolded Run for the Upper One

**Status:** ready-for-review  
**Date:** 2026-08-28

## Context

This CLI's entire job is invoking `bd` and `git` correctly, so a test
suite that fakes both tests almost nothing that matters: every trap it
exists to guard is a property of the real tools. Its exit status is part
of its contract ([[frd-swarm-core]], AC-SWARM-CORE-001.1), and a status is
not observable from inside the process that would return it.

Against that, the tests must run on macOS, Linux and Windows
([[adr-023]]), and the tracker is a heavier dependency than git: it
carries an embedded database, and its availability on every runner is not
established.

This repository already has the machinery for the upper tier. Cases under
`cli/testdata/cases/` are pure data — a `case.toml` naming the argv and
the expected exit status, an `input/` corpus, and committed `expected/`
artifacts — replayed by one harness, and replayed a second time from a
CRLF working tree so a line-ending difference is caught rather than
assumed away.

There is no fixture a swarm case can point at, though. A run is a git
repository with a branch and commits, a tracker holding a run epic, a
phase chain, a flat target and beads under it, and possibly worktrees.
Nothing today creates one.

## Decision

We will test in two tiers.

**Unit** — components are exercised against a fake tool gateway that
replays recorded tool output. Because a command builds a plan value
([[adr-024]]), the plan itself is asserted without running anything, and
each decoder is tested against the recorded shapes it must accept.

**End-to-end** — the compiled `swarm` binary is executed as a subprocess
against a real git repository and, where the case needs one, a real
tracker, both created by a new `swarm scaffold` command. Output and exit
status are compared against committed snapshots, through a declared
normalization list.

`swarm scaffold` ships in the binary. A case declares which fixtures it
needs; a case needing the tracker fails where the tracker is expected and
is reported as skipped-with-a-reason elsewhere, never skipped silently.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Unit tests only, everything faked | Fast; runs anywhere | Every trap the CLI exists to guard is a property of the real tools, so the suite would pass while the thing it guards is broken | Tests the wrapper and not the wrapping |
| End-to-end only, real tools throughout | Highest fidelity | Slow per case; requires the tracker on every runner; a decoder's handling of a rare empty shape is nearly impossible to provoke through the real tool | Cannot reach the cases that matter most |
| In-process invocation of the command tree | Fast; no build step; the existing spec-CLI harness does this | Cannot observe a real exit status, which is a first-class part of this CLI's contract; shares the test process's working directory and environment, so a command that mishandles either passes | Cannot test the contract's most-routed-on property |
| Recorded tool output as the only fidelity | Deterministic; no tool needed at test time | A recording rots silently when the tool changes, which is exactly the failure class this CLI exists to catch — the corpus has four retracted claims about tool behaviour to prove it | Reproduces the failure being fixed, one layer down |
| Two tiers, recordings below and real tools above | Each tier covers the other's weakness: recordings reach the rare shapes, real tools catch a tool that changed under a recording | Two harnesses; a fixture generator to build and keep honest; snapshots that must be normalized without normalizing away the thing under test | N/A — selected |

## Consequences

**Positive:**

- The exit status, the printed invocations and the stream separation are
  tested as a caller sees them, because a subprocess is what a caller
  runs.
- A tracker upgrade that changes a subcommand's shape fails the
  end-to-end tier, which is the guard the recorded tier cannot provide.
- Rare shapes — a bare null where an array was expected, an object where
  a count was expected — are reachable in the unit tier, where the real
  tool cannot easily be made to produce them.
- The fixture is a command rather than a test helper, so a human
  debugging a command by hand builds the same run a case does, and does
  it with one invocation.
- The existing CRLF replay extends to these cases at no cost, so a
  Windows line-ending difference is caught by machinery that already
  exists.

**Negative:**

- Snapshot normalization is a standing hazard in both directions. Shas,
  timestamps, durations and temporary paths must be normalized or every
  case fails on every run; normalize too broadly and the case stops
  asserting what it was written for. The substitution list is declared
  per case and is itself something to review.
- A fixture generator is code that can be wrong, and a case passing
  against a fixture that does not resemble a real run is a false pass
  that nothing else catches.
- `swarm scaffold` is a development affordance shipped in the production
  binary. It creates repositories and tracker databases, so it is the one
  command whose blast radius is a directory the caller names rather than
  a run.
- Cases needing the tracker do not run everywhere, so the platform
  guarantee for those commands rests on the unit tier plus the
  cross-compile, not on an end-to-end run per platform.
- Two harnesses is more surface than one, and a case can be written in
  the wrong tier — the cheap tier is the tempting one.

## References

- [[frd-swarm-core]] — the contract these tiers verify, and the exit
  status an in-process test cannot observe
- [[adr-023]] — the three-OS matrix and the six-target cross-compile
  these tiers run in
- [[adr-024]] — the plan value the unit tier asserts without executing
- [[adr-025]] — the gateway the fake replaces, and the per-subcommand
  decoders the recorded shapes exercise
