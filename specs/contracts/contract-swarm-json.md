---
title: "Contract: Swarm CLI JSON Surfaces"
description: The machine-readable output every swarm command emits under
  --json, and the run manifest it reads and writes — one document over the
  Strata schemas that fix their shape
slug: contract-swarm-json
type: contract
status: draft
updated_at: 2026-08-28T00:00:00Z
provider: container-swarm-cli
visibility: internal
consumers:
  - "module-bead-bonanza"
specs:
  - "schemas/swarm-manifest.strata"
  - "schemas/swarm-output.strata"
version: "1.0"
---

# Swarm CLI JSON Surfaces

## Overview

This contract covers the two machine-readable surfaces of the swarm CLI:
the JSON every command emits under `--json`, and the **run manifest** —
the JSON object in a run epic's tracker metadata that the CLI reads,
validates and writes. Both are consumed by the `bead-bonanza` skill
family, which parses the first and, during migration, still writes the
second. Provided by [[container-swarm-cli]]. Visibility: internal.

The manifest is in this contract rather than in the container's Data
section for one reason: the container is not its only writer. Several
skills write keys into the same object today, which makes them
independently-built consumers of its shape, and a private contract is
defined as one with a single writer.

## Specification

Shape is fixed by Strata schemas, transpiled to JSON Schema and to Go
types by the same generator the rest of this repository uses
([[adr-016]]), so the declared key set the CLI validates against and the
schema a consumer validates against are generated from one source rather
than kept in step by hand.

| Artifact | Location | Normative for |
|----------|----------|---------------|
| Manifest schema | `schemas/swarm-manifest.strata` | Shape of the run manifest: every key, its type, and the stage at which it becomes required |
| Output schema | `schemas/swarm-output.strata` | Shape of each command's `--json` result, and of the shared envelope every result carries |
| Generated JSON Schema | `cli/core/schema/` via `scripts/generate-schemas.sh` | The validating artifact; regenerated, never hand-edited |
| This document | — | Semantics, guarantees, and change policy |

## Guarantees

- The provider SHALL emit, under `--json`, every fact its text form
  reports for the same invocation.
- The provider SHALL emit a result conforming to the output schema on
  every exit status, including a non-zero one, so a consumer parsing the
  result never has to fall back to reading text to learn why.
- Every result SHALL carry the invocations the command ran, in order,
  with each one's exit status and duration.
- WHEN a fact could not be established, the provider SHALL emit it as
  unestablished rather than omitting it or emitting an empty value.
- WHEN a manifest key is absent, the provider SHALL report it as absent
  rather than as an empty value of its declared type.
- The provider SHALL write the manifest by merging, so a key it did not
  set retains the value another writer gave it.
- IF a manifest write names a key outside the declared set, THEN the
  provider SHALL refuse the write and change nothing.
- IF a manifest write carries a value whose type is not its key's
  declared type, THEN the provider SHALL refuse the write and change
  nothing.
- The provider SHALL NOT emit a field whose meaning depends on the
  locale, the time zone or the terminal width of the process that
  produced it.
- Consumers SHALL tolerate unknown fields, so that adding one is not a
  breaking change.

## Operational Expectations

- **Availability / SLO:** not applicable. The provider is a binary
  invoked per command, not a service.
- **Latency:** bounded by the tracker and version-control invocations a
  command makes, which the result reports individually. The provider adds
  no budget of its own.
- **Rate limits / quotas:** none.
- **AuthN/AuthZ:** none of its own. The provider acts with the
  credentials of the process that invoked it, and inherits whatever
  access the tracker and the remote grant that process.
- **Consumer obligations:**
  - Consumers SHALL route on the exit status rather than on the presence
    or absence of a field.
  - Consumers SHALL tolerate unknown fields.
  - Consumers SHALL NOT parse the text form of any command that offers
    `--json`.

## Versioning and Change Policy

- **Versioning scheme:** semantic versioning on this contract, carried in
  the envelope every result emits, so a consumer can branch on the
  version it was handed rather than on the binary's own version.
- **Compatible changes:** the provider MAY add a field, add a manifest
  key, or widen an enumeration without a major version change.
- **Breaking changes:** WHEN a field or a manifest key is removed,
  renamed or retyped, the provider SHALL raise the major version. A
  renamed manifest key is the sharpest case in this contract: a reader
  gets absent rather than an error, and a run then proceeds without the
  value.
- **Deprecation:**
  - WHEN a field or key is deprecated, it SHALL continue to be emitted
    and accepted for one major version.
  - WHILE a field or key is deprecated, the schema SHALL mark it so, and
    a consumer generating types from the schema sees the deprecation.

## Key Decisions

- [[adr-016]] — Strata as the one source for schemas and generated
  models, which is why shape lives in `.strata` rather than in this
  document.
- [[adr-025]] — the per-subcommand decoders that produce the values this
  contract publishes.

## Open Questions

- [ ] Does this contract become public when the CLI is distributed
      outside this repository? It is internal while the CLI and its
      consumers ship together; the first consumer that upgrades
      independently makes it public, and its change policy tightens then.

## References

- [[container-swarm-cli]] — the provider
- [[frd-swarm-core]] — the requirements the emitted facts come from
- [[module-bead-bonanza]] — the consumer whose components parse this
  output and write this manifest
- `.claude/skills/spec/references/shared/c4-guide.md` — contract
  visibility tiers
