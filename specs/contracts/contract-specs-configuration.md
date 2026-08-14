---
title: "Contract: Specs Configuration"
description: The specs.toml format an adopting repository authors, and
  the in-process registration surface a profile uses to extend it
slug: contract-specs-configuration
type: contract
status: approved
updated_at: 2026-08-07T00:00:00Z
provider: container-spec-cli
visibility: public
consumers: []
specs:
  - "cli/core/schema/specs-config.schema.json"
version: "1.0"
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
inside one process.

**In v1 the public surface selects a profile; it does not modify one.**
An adopting repository says which repository it is, where its documents
live, which profile is active, and a small number of house conventions.
Everything about a document type — its schema, its slug pattern, the
sections it requires — is supplied by the active profile and is not
overridable from `specs.toml`. The vocabulary below therefore appears
once, in the internal surface, and the public surface names it rather
than redefining it.

The two surfaces are one contract because they are one vocabulary seen
from two sides, and because the boundary between them is the thing most
likely to move: a key promoted from internal to public in a later
version is a change to both halves at once, and this document is where
that promotion is decided.

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
```

That is the whole of it in v1. `[workspace].name`, `[workspace].root`
and `[profile].name` are required; `[conventions]` keys resolve from the
active profile when absent.

There is deliberately no way to override a document type, add one,
register a fenced block tag, or supply a schema. Those are the internal
surface, and holding them back is what keeps v1's compatibility promise
cheap: a key that was never public cannot be broken by changing it, and
the framework's vocabulary is still moving. Promotion is one-way and
additive — a repository that never wrote a key it could not write is
unaffected by a later version that lets it.

### The internal surface: profile registration

**Internal — not an extension point outside this repository, and not
reachable from `specs.toml` in v1.** A profile supplies, for each
document type it registers: the type name, the JSON Schema its
frontmatter satisfies, the slug pattern, whether the type is numbered,
the section hierarchy it requires, and the named section sets a span
query can select. It additionally supplies the fenced block tags to
extract and any rule that cannot be expressed as data — the
record-XOR-blueprint-link invariant and supersession integrity are the
two current examples.

Required sections are a **hierarchy, not a flat list.** A document's
headings nest, and the rules worth enforcing are about that nesting: a
missing `## Consequences` and a `### Consequences` misplaced under
`## Decision` are different defects, and a flat list cannot tell them
apart. It also cannot say that `## Alternatives Considered` holds one
subsection per alternative — the shape most likely to be got wrong. The
registered form is therefore a tree, each node naming a heading, its
level relative to its parent, whether it is required, and its own
children:

```go
type Section struct {
    Title    string
    Required bool
    Repeated bool      // one occurrence per instance, e.g. per alternative
    Children []Section
}
```

`Repeated` is what distinguishes a heading that appears exactly once
from one whose count is determined by content. Without it a tree can
only express fixed documents, and the templates this profile mirrors are
not all fixed.

A profile also supplies, per document type, which of its frontmatter
properties are reference-bearing, which of those are typed, and which
are paired — three facts a JSON Schema cannot express, because what a
slug-shaped value resolves to, and what its target must be, are
questions about the corpus rather than about the value itself:

- **Reference-bearing properties** are the frontmatter properties of the
  type whose value is a document slug. A value in a property absent
  from this list contributes no reference however slug-shaped it looks:
  what is a reference is registered, never inferred from a value's
  shape.
- **Typed reference properties** narrow a reference-bearing property
  further, to the one document type its target must declare — a
  feature blueprint's `frd` property must resolve to a document of type
  `frd`, for example. Registered as a mapping from the property's name
  to the required type; every key also names a reference-bearing
  property, since a typed property that is not reference-bearing has no
  target to check the type of.
- **Pairings** are the reference-bearing properties that must be
  declared from both ends: a document naming a counterpart in one of
  them, where the target declares nothing in return, is a half-built
  edge. Each pairing names the property on this type and the
  counterpart property the target must declare in return. A profile MAY
  register none — a vocabulary whose relationships are all recoverable
  by querying inbound references, rather than by requiring a reverse
  declaration, registers an empty pairing set, and the check this feeds
  is inert for it by design.

```go
type DocumentTypeRules struct {
    ReferenceProperties      []string
    TypedReferenceProperties map[string]string // property name -> required type
    Pairings                 []PropertyPairing
}

type PropertyPairing struct {
    Property    string // this type's property
    Counterpart string // the property the target must declare in return
}
```

A profile separately supplies one **lifecycle registration** for the
whole of its document vocabulary, not per type, since status is a
property of the shared document envelope rather than of any one type:
the full status vocabulary its documents may declare, which of those
statuses a document is in effect at (live) and which it is no longer in
effect at (retired), and the frontmatter property in which a document
at a retired status names its replacement. A profile MAY register no
lifecycle — the zero value — describing a corpus whose documents have
none; the checks that read the registration do not run for it.

```go
type LifecycleRules struct {
    Statuses          []string
    LiveStatuses      []string
    RetiredStatuses   []string
    SuccessorProperty string
}
```

Because the public surface no longer overlaps the internal one, the
relationship binding them into a single contract is naming rather than
override: `[profile].name` selects one of these registrations, and
`[conventions]` supplies values a registration reads. A profile whose
name no adopter can select, or a convention no registration consumes, is
a defect in this contract rather than in either module.

`version` stays `"1.0"` for the whole of the internal surface described
above, including the reference-bearing, typed-reference, pairing and
lifecycle registrations: the Versioning and Change Policy section below
exempts the internal registration surface from versioning while
[[module-document-core]] and [[module-spec-profile]] are both
`visibility: internal`, which they are, and this vocabulary is additive
to that internal half only. The public `specs.toml` surface above is
unchanged, so no adopting repository's file changes meaning under this
version.

## Guarantees

### Resolution

- WHEN a `[conventions]` key is absent from `specs.toml`, the CLI SHALL
  resolve it from the active profile's registered value for that key.
- The CLI SHALL apply exactly one profile per invocation.
- The CLI SHALL NOT accept a document-type definition, a block-tag
  registration or a schema path from `specs.toml` in v1. WHERE such a
  key appears, it is an undefined key and the checking guarantees below
  apply to it.
- The CLI SHALL resolve `[workspace].root` relative to the directory
  containing `specs.toml`, so a configuration file is portable with its
  repository.
- The CLI SHALL NOT read configuration from environment variables, a
  user home directory, or any location outside the repository. A
  corpus check SHALL produce the same findings for the same repository
  content on any machine.

### Checking the configuration

**Validating `specs.toml` is a check the CLI performs, not an
error path it falls down.** A malformed configuration file is the first
thing an adopting repository will get wrong and the thing it has least
help with, so it is reported through the same machinery as every other
defect rather than as a bare startup failure.

- WHEN the CLI validates `specs.toml`, it SHALL emit findings carrying
  the rule identifier, severity, path and span that AC-SPEC-006.1 of
  [[frd-spec-validation-cli]] requires of every finding, so a
  configuration defect is consumable by the same tooling as a corpus
  defect.
- WHEN the CLI validates `specs.toml`, it SHALL report every defect in
  the file rather than stopping at the first, so one run tells an author
  everything that needs fixing.
- IF `specs.toml` contains a key this contract does not define, THEN the
  CLI SHALL report a finding naming the key and its span. A misspelled
  key is far more common than a key from the future, and silently
  ignoring it would let a repository believe it had configured something
  it had not.
- IF a value fails the configuration schema, THEN the CLI SHALL report a
  finding naming the offending key and its span, and SHALL NOT fall back
  to a default for that key.
- IF `[profile].name` names no registered profile, THEN the CLI SHALL
  report a finding naming the value and listing the registered profile
  names.
- WHEN `specs.toml` is absent, the CLI SHALL report the path it looked
  for and exit non-zero, and SHALL NOT infer a spec root. This is the
  one configuration failure that is not a finding, because there is no
  file to carry a span into.
- The configuration checks SHALL be exercised by snapshot cases under
  [[testing-standards]] on the same terms as corpus checks, since each
  input corpus already carries its own `specs.toml`.

## Operational Expectations

This is a file-format contract rather than a service interface, so
availability, latency, rate limits and authentication do not apply. What
does:

- **Consumer obligations:** an adopting repository SHALL commit
  `specs.toml` to version control, so that a check run in CI and a check
  run locally resolve the same configuration.
- **Consumer obligations:** a repository needing a document vocabulary
  the active profile does not define SHALL raise it as a change to the
  profile rather than expect a local override. In v1 there is no
  override, by design — the alternative is a fork of the vocabulary,
  which is what this framework exists to prevent.
- **Provider obligations:** the CLI SHALL report configuration problems
  before any corpus work begins, so a misconfigured repository fails
  fast rather than producing a report derived from wrong settings.

## Versioning and Change Policy

- **Versioning scheme:** semantic versioning on this contract, tracked
  in the `version` frontmatter property and mirrored in the
  configuration JSON Schema.
- **Compatible changes:** the provider MAY add optional keys, add a
  document type to a profile, promote an internal key to the public
  surface, and relax a value constraint without a major version change.
  An existing `specs.toml` SHALL continue to resolve unchanged across a
  minor version. Promotion is compatible in this direction only: a key
  no repository could previously write cannot be written wrongly by an
  existing file, which is the reason v1 holds the type vocabulary back.
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
      documents. Now that v1's public surface does not overlap the
      internal one, the case for splitting is stronger than it was —
      what holds them together is the promotion path rather than a
      shared key set. The split becomes right if the registration
      surface is ever published on its own.
- [ ] Which internal keys are promoted first, and on what evidence. The
      honest answer is "whichever an adopting repository is blocked on",
      and no adopting repository exists yet. Deciding before then would
      be guessing at demand.
- [ ] Whether a repository may activate more than one profile — for
      example a base framework profile plus a house-rules profile. The
      guarantee above says exactly one, which is the simple answer;
      composition is the more useful one and has not been designed. It
      is also the shape a house-rules override would take if promotion
      is ever asked for, so the two questions resolve together.
- [ ] Whether the required-section hierarchy expresses ordering. The
      tree above captures nesting and repetition but says nothing about
      whether `## Decision` must precede `## Consequences`. Templates
      imply an order; whether that order is normative has not been
      decided, and enforcing it prematurely would reject documents that
      are merely unconventional.

## References

- `c4-guide.md` — C4 mapping rules and contract visibility tiers
- [[container-spec-cli]] — the provider's blueprint
- [[module-document-core]] — the host the internal surface registers into
- [[module-spec-profile]] — the profile that populates it
- [[frd-spec-validation-cli]] — FEAT-SPEC-008, the requirements this
  contract realizes, and AC-SPEC-006.1, the finding shape configuration
  checks emit
- [[testing-standards]] — the snapshot cases that exercise the
  configuration checks
