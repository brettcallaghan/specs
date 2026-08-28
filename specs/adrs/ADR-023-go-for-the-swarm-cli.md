---
title: "ADR-023: Go for the Swarm CLI, in the Existing Module"
description: Go is chosen for the same measured reasons as the spec CLI, and
  a shell language is disqualified by the very hazards this CLI exists to
  remove; the binary ships from the existing cli module rather than its own
slug: adr-023
type: adr
status: ready-for-review
updated_at: 2026-08-28T00:00:00Z
---

# ADR-023: Go for the Swarm CLI, in the Existing Module

**Status:** ready-for-review  
**Date:** 2026-08-28

## Context

[[frd-swarm-core]] is approved and specifies a CLI whose whole job is to
invoke `bd` and `git` correctly and report what it did. Two properties of
that job bear on the language choice more than performance does.

**Its reason to exist is that shell pipelines lose information.** Two of
the hazards its criteria pin are properties of shell semantics rather
than of any particular script: an exit status read through a pipe is the
reader's status, so a failing step reads as a pass; and a reader that
closes a pipe early makes a successful writer report failure under
`pipefail`. Both have been observed in this repository's own scripts. A
language whose native idiom is the pipeline reintroduces them at every
call site.

**It must run where a run runs.** Coordinators run on Linux containers,
contributors on macOS, and the repository already commits to Windows —
`.github/workflows/cli.yml` runs `go test` on `ubuntu-latest`,
`macos-latest` and `windows-latest`, and cross-compiles six
`GOOS`/`GOARCH` targets. Whatever this CLI is written in inherits that
obligation.

[[adr-015]] already chose Go for the spec CLI on measured startup and a
single-artifact distribution requirement, against Rust, Python and
TypeScript. That measurement is not re-run here; what this decision adds
is a candidate that ADR-015 never considered — a shell language — and a
second question ADR-015 did not face: whether a second binary shares the
first one's module.

## Decision

We will write the swarm CLI in Go, as a second binary at `cli/swarm`
within the existing `github.com/brettcallaghan/agent-plugins/cli` module,
built and tested by the matrices that already build and test `spec`.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| PowerShell 7 | Cross-platform since 6; object pipeline rather than text; native on Windows | Requires a runtime install on Linux and macOS, which [[adr-015]]'s distribution constraint forbids; a pipeline language makes the two measured status hazards the default idiom rather than an avoidable mistake; no in-process typed access to `bd`'s JSON without a parse step per call | Reintroduces the failure class the CLI exists to remove |
| Bash | Zero install on Linux and macOS; the scripts it replaces are already Bash | Not present on Windows without a compatibility layer; the same pipeline hazards; every existing bundled script carries a portability wrapper for `stat`, `readlink`, `mapfile` and `grep -P` precisely because Bash is not one language | Fails the Windows target outright |
| Python | Rich standard library; readable | A runtime install on the target machine, and 33.30 ms interpreter startup measured in [[adr-015]] | Same distribution constraint that rejected it for the spec CLI |
| Rust | Ties Go on startup and artifact size ([[adr-015]]) | A second toolchain in CI and in every contributor's environment, for a repository whose only compiled artifact is Go | No benefit to offset a second toolchain |
| Go, in a new module | The swarm CLI stays vendable on its own; its dependency set never enters the spec CLI's | A second CI matrix, a second toolchain pin, a second `go.sum` to keep current, and the snapshot-case harness duplicated or extracted before it has a second user | Cost paid now for portability nobody has asked for; reversible later, and cheaply, because a module split is mechanical |
| Go, in the existing module | Inherits the three-OS test matrix, the six-target cross-compile, the format and vet gates, and the snapshot-case harness; one `go test ./...` covers both binaries | The two binaries share one dependency graph and one release cadence | N/A — selected |

## Consequences

**Positive:**

- The Windows obligation is met by infrastructure that already exists and
  already runs, rather than by a matrix written for this container.
- The four properties the command contract needs are stdlib behaviour
  rather than conventions to maintain. Measured in this environment on
  Go 1.24.7, running a child that writes to both streams and exits 7:
  the exit status read back is `7` from the process itself, stdout is
  `"data"` and stderr `"warning"` in separate buffers, and the elapsed
  time is available per invocation. A 2 MB stdout is collected whole,
  because nothing in the arrangement is a reader that can close a pipe
  early — so the `pipefail` inversion has no way to occur.
- Arguments are passed as an argv vector, not a command line: `printf`
  invoked with the single argument `a; rm -rf /; $(echo hi) | tee x`
  printed it literally. Quoting bugs in constructed commands become
  impossible rather than guarded against.
- A missing executable is `*exec.Error` while a non-zero exit is
  `*exec.ExitError`, so "the tool is not installed" and "the tool
  answered no" are distinguishable without parsing a message — which is
  what the malformed-invocation status needs to be separable from a
  finding.
- Contributors and agents need one toolchain to build everything this
  repository ships.

**Negative:**

- The two binaries share a dependency graph. A library the swarm CLI
  needs is in the spec CLI's module graph whether or not `spec` imports
  it, and `go.sum` growth is shared.
- They share a release cadence: a tag builds both, and a swarm-only fix
  ships a new `spec` binary with an unchanged tree.
- Anyone wanting the swarm CLI alone vendors a module containing the spec
  CLI as well.
- The cross-compile leg and the release workflow must build a second
  target; a build that silently kept compiling only `./spec` would leave
  the Windows guarantee unproven for this binary while CI stayed green.

## References

- [[adr-015]] — Go for the spec CLI, and the startup and artifact
  measurements this decision rests on rather than repeating
- [[frd-swarm-core]] — the requirements this CLI implements
- `.github/workflows/cli.yml` — the three-OS test matrix and six-target
  cross-compile this binary joins
