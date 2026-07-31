---
title: "Contract: Specs Configuration"
description: The specs.toml format an adopting repository authors, and
  the in-process registration surface a profile uses to extend it
slug: contract-specs-configuration
type: contract
status: draft
updated_at: 2026-07-31T00:00:00Z
provider: container-spec-cli
visibility: public
consumers: []
specs:
  - "cli/core/schema/specs-config.schema.json"
version: "1.0"
related:
  - "frd-spec-validation-cli"
  - "module-document-core"
  - "module-spec-profile"
---

# Specs Configuration

## Overview

This contract governs how a repository tells the Spec CLI what its
specification corpus is and which rules apply to it. It has two
surfaces at different tiers, described together because they are two
views of one vocabulary.

The **public surface** is `specs.toml`, a file an adopting repository
authors at its root. It crosses a boundary of independent change: teams
outside this one write it, and their files must keep working as the CLI
evolves. The **internal surface** is the in-process registration
structure a profile hands to Document Core, which crosses the module
boundary between [[module-spec-profile]] and [[module-document-core]]
inside one process. Every key an adopter can write in `specs.toml`
resolves against a value a profile registered, so a change to one
surface is almost always a change to both.

Provided by [[container-spec-cli]]. Visibility: public, on the strength
of the `specs.toml` half; the registration surface is marked internal
where it appears below and SHALL NOT be treated as an extension point
by anything outside this repository.

## Specification

| Artifact | Location | Normative for |
|----------|----------|---------------|
| Configuration JSON Schema | `cli/core/schema/specs-config.schema.json` | Shape of `specs.toml` |
| `Profile` interface | `cli/core/corpus` (Go) | Shape of the internal registration surface |
| This document | — | Semantics, resolution order, change policy |

### The public surface: `specs.toml`

```toml
# specs.toml — at the repository root.

[workspace]
name = "agent-plugins"   # identifies this repository's documents when
                         # indexes from several repositories are merged
root = "docs"            # the spec root the CLI walks

[profile]
name = "spec-driven-dev" # which document vocabulary to activate

[conventions]
line_width = 80          # measured in Unicode codepoints

# Optional. Any key omitted here takes the active profile's value.
[types.adr]
slug_pattern = "^adr-[0-9]{3}$"
numbered = true
required_sections = ["Context", "Decision", "Alternatives Considered",
                     "Consequences"]

# Optional. Registers a fenced block tag for structured-data extraction.
[[blocks]]
tag = "component"
format = "yaml"
schema = ".spec/schemas/component-record.json"
```

`[workspace].name` and `[workspace].root` are required. Everything else
resolves from the active profile when absent.

### The internal surface: profile registration

**Internal — not an extension point outside this repository.** A profile
supplies, for each document type it registers: the type name, the JSON
Schema its frontmatter satisfies, the slug pattern, whether the type is
numbered, the headings it requires, and the named section sets a span
query can select. It additionally supplies the fenced block tags to
extract and any rule that cannot be expressed as data — the
record-XOR-blueprint-link invariant and supersession integrity are the
two current examples.

The registration surface is the union of what `specs.toml` can override
and what only a profile can supply. That relationship is the reason the
two are one contract: a profile that registers a type no adopter can
override, or an override key no profile supplies, is a defect in this
contract rather than in either module.

## Guarantees

- WHEN a key is absent from `specs.toml`, the CLI SHALL resolve it from
  the active profile's registered value for that key.
- WHEN a key is present in `specs.toml`, the CLI SHALL prefer it over
  the profile's registered value, so an adopting repository can diverge
  from a profile without forking it.
- The CLI SHALL apply exactly one profile per invocation.
- WHEN `specs.toml` is absent, the CLI SHALL report the path it looked
  for and exit non-zero, and SHALL NOT infer a spec root.
- IF `specs.toml` contains a key this contract does not define, THEN the
  CLI SHALL report a finding naming the key and its position in the
  file. A misspelled key is far more common than a key from the future,
  and silently ignoring it would let a repository believe it had
  configured something it had not.
- IF a value fails the configuration schema, THEN the CLI SHALL report
  the offending key and its position, and SHALL NOT fall back to a
  default for that key.
- The CLI SHALL resolve `[workspace].root` relative to the directory
  containing `specs.toml`, so a configuration file is portable with its
  repository.
- The CLI SHALL NOT read configuration from environment variables, a
  user home directory, or any location outside the repository. A
  corpus check SHALL produce the same findings for the same repository
  content on any machine.

## Operational Expectations

This is a file-format contract rather than a service interface, so
availability, latency, rate limits and authentication do not apply. What
does:

- **Consumer obligations:** an adopting repository SHALL commit
  `specs.toml` to version control, so that a check run in CI and a check
  run locally resolve the same configuration.
- **Consumer obligations:** a repository overriding a type's
  `slug_pattern` SHALL accept that documents conforming to its override
  may not conform to the profile's default, and that the two cannot both
  be enforced.
- **Provider obligations:** the CLI SHALL report configuration problems
  before any corpus work begins, so a misconfigured repository fails
  fast rather than producing a report derived from wrong settings.

## Versioning and Change Policy

- **Versioning scheme:** semantic versioning on this contract, tracked
  in the `version` frontmatter property and mirrored in the
  configuration JSON Schema.
- **Compatible changes:** the provider MAY add optional keys, add a
  document type to a profile, and relax a value constraint without a
  major version change. An existing `specs.toml` SHALL continue to
  resolve unchanged across a minor version.
- **Breaking changes:** WHEN a key is removed or renamed, or a value
  constraint is tightened, the provider SHALL bump the major version.
- **Deprecation:** WHEN a key is deprecated, it SHALL continue to
  resolve for one major version, and the CLI SHALL report a finding of
  severity `warning` naming the key and its replacement.
- The internal registration surface is exempt from this policy WHILE
  [[module-document-core]] and [[module-spec-profile]] are both
  `visibility: internal` — they change in lockstep in one repository.
  WHEN either is published, the registration surface becomes subject to
  the same policy as the public surface.

## Key Decisions

- [[adr-015]] — Go, and the in-process module composition that makes the
  internal surface an interface rather than a wire format.

## Open Questions

- [ ] Whether the two surfaces should be split into separate contract
      documents. They are one document because every public key resolves
      against a registered value, and splitting them would put that
      relationship in neither. The split becomes right if the
      registration surface is ever published on its own.
- [ ] Whether `[[blocks]]` schemas are resolved relative to `specs.toml`
      or to the spec root. The example above assumes relative to
      `specs.toml`; the alternative reads better when the spec root is
      deep. Affects portability of a configuration between repositories.
- [ ] Whether a repository may activate more than one profile — for
      example a base framework profile plus a house-rules profile. The
      guarantee above says exactly one, which is the simple answer;
      composition is the more useful one and has not been designed.

## References

- `c4-guide.md` — C4 mapping rules and contract visibility tiers
- [[container-spec-cli]] — the provider's blueprint
- [[module-document-core]] — the host the internal surface registers into
- [[module-spec-profile]] — the profile that populates it
- [[frd-spec-validation-cli]] — REQ-SPEC-008, the requirements this
  contract realizes
