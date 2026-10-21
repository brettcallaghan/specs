---
title: "Specs Configuration"
description: The specs.toml format an adopting repository authors, and
  the in-process registration surface a profile uses to extend it
slug: contract-specs-configuration
type: contract
status: approved
provider: container-spec-cli
visibility: public
specs:
  - "cli/core/schema/specs-config.schema.json"
version: "1.1"
---

# Specs configuration

## Overview

This contract governs how a repository tells the Spec CLI what its
specification corpus is and which rules apply to it. It has two
surfaces at different tiers, described together because they are two
views of one vocabulary.

The **public surface** is `specs.toml`, a file an adopting repository
authors at its root. It crosses a boundary of independent change: teams
outside this one write it, and their files must keep working as the CLI
evolves. The **internal surface** is the in-process registration
structure a profile hands to Document Core. It crosses the module boundary
between [[module-spec-profile]] and [[module-document-core]] inside one
process.

**In v1 the public surface selects a profile; it does not modify one.**
An adopting repository says which repository it is, where its documents
live, which profile is active, and a small number of house conventions.
The active profile supplies everything about a document type (its
schema, its slug pattern, and the sections it requires), and
`specs.toml` can't override it. The vocabulary below therefore appears
once, in the internal surface, and the public surface names it instead
of redefining it.

The two surfaces are one contract because they're one vocabulary seen
from two sides, and because the boundary between them is the thing most
likely to move: a key promoted from internal to public in a later version
is a change to both halves at once. This document is where that promotion
is decided.

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
# specs.toml — found in the working directory or a directory above it, up to the repository root.

[workspace]
name = "agent-plugins"   # identifies this repository's documents when
                         # indexes from several repositories are merged
root = "docs"            # the spec root the CLI walks

[profile]
name = "specs" # which document vocabulary to activate

[conventions]
line_width = 80          # measured in Unicode codepoints
```

That's the whole file in v1. `[workspace].name`, `[workspace].root`, and
`[profile].name` are required. `[conventions]` keys resolve from the
active profile when they're absent.

There's deliberately no way to override a document type, add one,
register a fenced block tag, or supply a schema. Those belong to the
internal surface. Holding them back is what keeps v1's compatibility
promise cheap: a key that was never public can't be broken by changing
it, and the framework's vocabulary is still moving. Promotion is one-way
and additive: a repository that never wrote a key it couldn't write is
unaffected by a later version that lets it.

### The internal surface: profile registration

**Internal: not an extension point outside this repository, and not
reachable from `specs.toml` in v1.** For each document type it
registers, a profile supplies the type name, the JSON Schema its
frontmatter satisfies, the slug pattern, whether the type is numbered,
the section hierarchy it requires, and the named section sets a span
query can select. It also supplies the fenced block tags to extract and
any rule that can't be expressed as data. The agreement of an element's
declarations and supersession integrity are the two current examples.

Required sections are a **hierarchy, not a flat list.** A document's
headings nest, and the rules worth enforcing are about that nesting. A
missing `## Consequences` and a `### Consequences` misplaced under
`## Decision` are different defects, and a flat list can't tell them
apart. Nor can a flat list say that `## Alternatives considered` holds
one subsection per alternative, which is the shape most likely to be
gotten wrong. The registered form is therefore a tree, each node naming a
heading, its level relative to its parent, whether it's required, and
its own children:

```go
type Section struct {
    Title    string
    Required bool
    Repeated bool      // one occurrence per instance, e.g. per alternative
    Children []Section
}
```

`Repeated` distinguishes a heading that appears exactly once from one
whose count depends on content. Without it, a tree can only express
fixed documents, and the templates this profile mirrors aren't all fixed.

A profile also supplies, per document type, which of its frontmatter
properties are reference-bearing, which of those are typed, and which are
paired. A JSON Schema can't express these three facts, because what a
slug-shaped value resolves to, and what its target must be, are questions
about the corpus rather than about the value itself:

- **Reference-bearing properties** are the frontmatter properties of the
  type whose value is a document slug. A value in a property absent from
  this list contributes no reference, however slug-shaped it looks: what
  counts as a reference is registered, never inferred from a value's
  shape.
- **Typed reference properties** narrow a reference-bearing property
  further, to the one document type its target must declare. For
  example, a Feature Blueprint's `frd` property must resolve to a
  document of type `frd`. They're registered as a mapping from the
  property's name to the required type. Every key also names a
  reference-bearing property, because a typed property that isn't
  reference-bearing has no target to check the type of.
- **Pairings** are the reference-bearing properties that must be
  declared from both ends. A document that names a counterpart in one of
  them, where the target declares nothing in return, is a half-built
  edge. Each pairing names the property on this type and the counterpart
  property the target must declare in return. A profile MAY register
  none. A vocabulary whose relationships are all recoverable by querying
  inbound references, rather than by requiring a reverse declaration,
  registers an empty pairing set, and the check this feeds is inert for
  it by design.

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

A profile separately supplies one **lifecycle registration** for its
whole document vocabulary, not per type, because status is a property of
the shared document envelope rather than of any one type. The
registration names the full status vocabulary its documents may declare,
the statuses at which a document is in effect (live), the statuses at
which it's no longer in effect (retired), and the frontmatter property in
which a document at a retired status names its replacement. A profile
MAY register no lifecycle (the zero value), which describes a corpus
whose documents have none. The checks that read the registration don't
run for it.

```go
type LifecycleRules struct {
    Statuses          []string
    LiveStatuses      []string
    RetiredStatuses   []string
    SuccessorProperty string
}
```

A profile separately supplies one **tag vocabulary** for all of its
documents, again not per type, because a feature isn't a document type.
A feature is written inside a document of any type, or in a file that is
no document at all, and the tags that identify it are the same in both.
The vocabulary states the pattern a feature identifier matches, the pattern
a criterion identifier matches, and the prefix of the tag in which a
feature written outside any document names its owner. It also states the
capture groups a criterion identifier shares with its feature's
identifier, and the tags that qualify a criterion, together with what
each of them means. Neither pattern includes the leading marker that
makes an identifier a tag: the identifier is the value, and the tag is
only how it's written down.

`SharedCaptures` is the derivation rule, stated as data. A criterion
identifier is derived from its feature's identifier. The engine checks
that derivation by compiling both patterns, capturing both identifiers,
and comparing only the groups named here, so it knows nothing of either
pattern's internal structure. Every name must be a capture group that
both patterns define. A name that isn't is reported rather than
defaulted, because a comparison of two values that nothing captured
passes.

A qualifier's meaning is drawn from a fixed set rather than invented by
the profile, because a qualifier can only mean one of the things a
criterion record has somewhere to put: that the criterion states a
prohibition, that it can only be verified by hand, or that it records
the level it's intended to be verified at. A profile chooses each tag's
spelling and which of those meanings it carries. A tag that matches
neither pattern and that no qualifier names is carried on the record and
reported by nothing.

Every member is independent, and a profile MAY register none of them.
The zero value describes a corpus with no tagged features, and the
checks that read the registration don't run for it.

```go
type FeatureTagRules struct {
    FeaturePattern   string
    CriterionPattern string
    OwnerPrefix      string
    SharedCaptures   []string // capture groups both identifier patterns define
    Qualifiers       []QualifierTag
}

type QualifierTag struct {
    Tag     string
    Meaning specmodel.QualifierMeaning  // negation, manual verification, verification level
    Level   specmodel.VerificationLevel // for a verification-level tag and no other
}
```

A profile separately registers **record-block tags**: the fenced-block
info strings that mark a block as a record. A record is YAML, so a tag
carries no content format. Instead, it carries an optional JSON Pointer
fragment (for example `#/definitions/Name`) that addresses the definition
in the profile's own union schema that the record is validated against.
A profile MAY register none, which describes documents with no record
blocks. Registration validation reports a pointer that resolves to
nothing in the profile's schema instead of skipping it. A tag with no
pointer is valid, and its records aren't schema-validated.

```go
type RecordBlockTag struct {
    Tag    string // fenced-block info string
    Schema string // optional "#/definitions/<Name>" into Profile.Schema()
}
```

Because the public surface doesn't overlap the internal one, the
relationship that binds them into a single contract is naming, not
override: `[profile].name` selects one of these registrations, and
`[conventions]` supplies values a registration reads. A profile whose
name no adopter can select, or a convention no registration consumes, is
a defect in this contract rather than in either module.

The internal surface described above, including the reference-bearing,
typed-reference, pairing, lifecycle, tag-vocabulary, and record-block tag
registrations, doesn't change `version`, because the Versioning and
change policy section below exempts the internal registration surface
from versioning. That exemption holds while [[module-document-core]] and
[[module-spec-profile]] are both `visibility: internal`, which they are.
These registrations add to that internal half only. They leave the public
`specs.toml` surface unchanged, so no adopting repository's file changes
meaning because of them.

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
thing an adopting repository will get wrong, and the thing it has least
help with. So the CLI reports it through the same machinery as every
other defect, not as a bare startup failure.

- WHEN the CLI validates `specs.toml`, it SHALL emit findings carrying
  the rule identifier, severity, path and span that AC-SPEC-006.1 of
  [[frd-spec-index]] requires of every finding, so a
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
- WHERE no `--config` option names a file, the CLI SHALL search the
  working directory and then each directory above it, stopping at the
  repository root — the nearest directory holding a `.git` entry — or,
  outside any repository, at the filesystem root, and SHALL use the
  first `specs.toml` it finds. WHERE `--config` names a file, the CLI
  SHALL read that file and search for no other.
- WHEN `specs.toml` is absent, the CLI SHALL report every directory it
  searched — or, for a file `--config` named, that path — and exit with
  the environment-error status, and SHALL NOT infer a spec root. This is the
  one configuration failure that is not a finding, because there is no
  file to carry a span into.
- The configuration checks SHALL be exercised by snapshot cases under
  [[testing-standards]] on the same terms as corpus checks, since each
  input corpus already carries its own `specs.toml`.

## Operational expectations

This is a file-format contract, not a service interface, so
availability, latency, rate limits, and authentication don't apply. What
does apply:

- **Consumer obligations:** an adopting repository SHALL commit
  `specs.toml` to version control, so that a check run in CI and a check
  run locally resolve the same configuration.
- **Consumer obligations:** a repository needing a document vocabulary
  the active profile does not define SHALL raise it as a change to the
  profile rather than expect a local override. In v1 there's no
  override, by design: the alternative is a fork of the vocabulary,
  which this framework exists to prevent.
- **Provider obligations:** the CLI SHALL report configuration problems
  before any corpus work begins, so a misconfigured repository fails
  fast rather than producing a report derived from wrong settings.

## Versioning and change policy

- **Versioning scheme:** semantic versioning on this contract, tracked
  in the `version` frontmatter property and mirrored in the
  configuration JSON Schema.
- **Compatible changes:** the provider MAY add optional keys, add a
  document type to a profile, promote an internal key to the public
  surface, and relax a value constraint without a major version change.
  An existing `specs.toml` SHALL continue to resolve unchanged across a
  minor version. Promotion is compatible in this direction only: a key
  no repository could previously write can't be written wrongly by an
  existing file, which is why v1 holds the type vocabulary back.
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

## Key decisions

- [[adr-015]] — Go, and the in-process module composition that makes the
  internal surface an interface rather than a wire format.

## Open questions

- [ ] Whether the two surfaces should be split into separate contract
      documents. Because v1's public surface doesn't overlap the
      internal one, what holds them together is the promotion path, not
      a shared key set, which strengthens the case for splitting. The
      split becomes right if the registration surface is ever published
      on its own.
- [ ] Which internal keys are promoted first, and on what evidence. The
      honest answer is "whichever an adopting repository is blocked on",
      and no adopting repository exists yet. Deciding before then would
      be guessing at demand.
- [ ] Whether a repository may activate more than one profile, for
      example a base framework profile plus a house-rules profile. The
      guarantee above says exactly one, which is the simple answer.
      Composition is the more useful answer and hasn't been designed.
      It's also the shape a house-rules override would take if promotion
      is ever asked for, so the two questions resolve together.
- [ ] Whether the required-section hierarchy expresses ordering. The
      tree above captures nesting and repetition but says nothing about
      whether `## Decision` must precede `## Consequences`. Templates
      imply an order. Whether that order is normative hasn't been
      decided, and enforcing it prematurely would reject documents that
      are merely unconventional.

## References

- [[c4-standards]] — C4 mapping rules and contract visibility tiers
- [[container-spec-cli]] — the provider's blueprint
- [[module-document-core]] — the host the internal surface registers into
- [[module-spec-profile]] — the profile that populates it
- [[frd-spec-index]] — FEAT-SPEC-008, the requirements this
  contract realizes, and AC-SPEC-006.1, the finding shape configuration
  checks emit
- [[testing-standards]] — the snapshot cases that exercise the
  configuration checks
