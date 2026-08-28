---
title: "ADR-025: Every External Invocation Is Argv Through One Gateway"
description: One component runs every `bd` and `git` invocation as an argv
  vector with separated streams, a status read from the process and a
  per-subcommand decoder, because a shell string loses all four
slug: adr-025
type: adr
status: ready-for-review
updated_at: 2026-08-28T00:00:00Z
---

# ADR-025: Every External Invocation Is Argv Through One Gateway

**Status:** ready-for-review  
**Date:** 2026-08-28

## Context

Every step in a plan ([[adr-024]]) is one invocation of an external tool,
and [[frd-swarm-core]] pins four properties of it: the invocation is
printed, its streams are never merged, its exit status is read from the
step itself rather than through a reader, and its duration is available
because the progress measurements are ages.

Two further constraints come from the tools rather than from the
requirements. The tracker writes warnings to its error stream while its
data goes to its output stream, so anything merging them corrupts the
payload it is about to parse. And its JSON is not uniform: measured
across nine subcommands, an empty result comes back as a bare `null`
from one, an empty array from several, and an object from others — all
at exit 0. A generic "treat any non-array as empty" guard, which the
skills currently prescribe, is measurably wrong for the
object-returning subcommands: it reports every run as absent.

Windows is a supported target ([[adr-023]]), and it has no `sh`.

## Decision

We will route every external invocation through one gateway component
that takes an argv vector — never a command string handed to a shell —
captures standard output and standard error into separate buffers, reads
the exit status from the finished process, records the elapsed time, and
hands the captured output to a decoder the step names. Each tracker
subcommand's decoder is the only place that knows that subcommand's
result shape, and it normalizes that subcommand's absent, null and empty
forms to one empty value.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| A command string through `sh -c` | Reads like the scripts being replaced; pipelines available | No `sh` on Windows; every constructed argument needs quoting that is one bug away from wrong; a pipeline reintroduces both measured status hazards | Fails the Windows target and the reason the CLI exists |
| Merged output and error streams | One buffer; simpler | The tracker's warnings land in the middle of the JSON being parsed | Contradicts AC-SWARM-CORE-001.12, and was observed corrupting a parse |
| One generic JSON guard for every subcommand | One rule to remember | Measured wrong: the guard the skills prescribe reports an object-returning subcommand as empty, so a live run reads as absent | Silently wrong in the direction that reads as "nothing there" |
| A third-party process library | Streaming, async, retries | The properties needed are what `os/exec` already gives; the extra surface is asynchronous execution and output streaming, neither of which a per-step synchronous invocation wants | A dependency for behaviour the standard library already has |
| One gateway over `os/exec` with per-subcommand decoders | All four properties are structural; the shape hazard is confined to one file per subcommand, so a tracker upgrade that changes a shape changes one place | A decoder per subcommand is more code than one `jq` expression; the gateway is a chokepoint every command depends on | N/A — selected |

## Consequences

**Positive:**

- The four required properties are the standard library's behaviour
  rather than conventions to maintain. Measured on Go 1.24.7 against a
  child writing to both streams and exiting 7: the status read back was
  `7` from the process, stdout `"data"` and stderr `"warning"` in
  separate buffers, and the elapsed time available per invocation.
- The `pipefail` inversion cannot occur, because there is no reader that
  can close a pipe early. A 2 MB standard output was collected whole in
  the same measurement.
- Injection through a constructed argument is structurally impossible.
  Measured: `printf` invoked with the single argument
  `a; rm -rf /; $(echo hi) | tee x` printed it literally.
- "The tool is not installed" and "the tool answered no" are
  distinguishable without parsing a message — measured, a missing
  executable returns `*exec.Error` while a non-zero exit returns
  `*exec.ExitError` — which is what lets a malformed invocation take a
  different exit status from a finding.
- A tracker upgrade that changes one subcommand's empty shape changes one
  decoder. The claim about that shape becomes a test rather than a
  sentence in a skill file.

**Negative:**

- Every subcommand the CLI uses needs a decoder before it can be used,
  which is a real cost per new call site and a temptation to reach past
  the gateway.
- The gateway is a chokepoint: a defect in it is a defect in every
  command, and its own tests carry proportionate weight.
- Pipelines are unavailable by construction, so anything a pipeline would
  have done — filtering, counting, joining — is Go code operating on
  decoded values. That is the intent, and it is more code than the shell
  one-liner it replaces.
- A decoder that normalizes an empty result hides the difference between
  the shapes it accepts, so a tracker that starts returning a *new* empty
  shape is absorbed silently unless the decoder asserts the shapes it
  knows.

## References

- [[frd-swarm-core]] — the four properties this gateway provides
- [[adr-023]] — Go, and the Windows target that rules out a shell
- [[adr-024]] — the plan whose steps this gateway runs
- `docs/research/swarm-cli/2026-08-25-capability-audit.md` §3.1 — the
  measured shape inconsistency across the tracker's subcommands
