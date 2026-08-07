# agent-plugins-9kb.2 — Widen the Profile extension point for reference and lifecycle registration

## Bead description (as given)

**Blueprint:** specs/features/reference-integrity/reference-integrity-blueprint.md
— delta **XPT**, section "Profile extension point in [[module-document-core]]
(modified) — XPT".
**FRD:** specs/features/reference-integrity/reference-integrity-frd.md.
**Delivery:** staged (stage 1 of 2); this feature's closure set is
`bd list --label feature-spec-reference-integrity`.

Every fact this feature checks is vocabulary, and Document Core must not hold
any of it. The Profile interface has carried three members since it was defined
— a union schema, a per-type slug and numbered table, and the convention
defaults — and its own documentation names reference properties and lifecycle
rules as the capabilities that would widen it. This is that widening, reviewed
on its own merits as [[module-spec-profile]]'s Change Guidance requires, and
framed so that **any** profile can express these rules rather than so that this
one can.

**Verification timing.** This item proves AC-SPECREF-004.1 (unit) and owes that
citation. It also participates in AC-SPECREF-002.5 (proved by LIFE) and
AC-SPECREF-004.3 (proved by CMD), both of which land on items outside this
item's dependency closure. Code work may proceed as soon as this item is ready;
verification of those two criteria waits for LIFE and CMD, and the final
Definition-of-Done item is the backstop.

**Coordination — agent-plugins-9jo** (out of scope, do not fix here): the
generated models embed this framework's DocumentType vocabulary, coupling the
framework-agnostic half to it. This item widens the extension point without
adding vocabulary to the host and extends the existing boundary test to the new
registered values; it neither worsens nor fixes 9jo.

### DESIGN (as given)

Anchors: cli/core/corpus/profile.go (Profile, DocumentTypeRules); the boundary
tests in cli/core/corpus; the generated specmodel.DocumentTypeRegistration and
specmodel.ProfileRegistration that MODEL extends.

- DocumentTypeRules carries, per document type, the names of its
  reference-bearing frontmatter properties, its typed reference properties as a
  mapping from property name to the type its target must declare, and its
  property pairings (AC-SPECREF-004.1).
- The Profile interface additionally supplies the **lifecycle registration**:
  the statuses at which a document is live, the statuses at which it is retired,
  and the name of the property in which a retired document names its
  replacement.
- Document Core contains **no** literal retired, superseded_by, or any
  registered property name. The existing boundary test — no non-test source in
  cli/core/corpus carries a document-type literal — extends to the values this
  delta registers, so the widening cannot become a place to smuggle vocabulary
  into the host.
- WHEN a profile is activated, the module validates its registration
  (AC-SPECREF-004.3). The defects are exactly three: a pairing or a typed
  reference naming a property the same type does not register as
  reference-bearing; a typed reference naming a type the profile does not
  define; and a lifecycle registration naming a status outside the profile's own
  vocabulary.
- IF the registration carries a defect, the module reports **every** one it
  finds, naming the offending key, rather than falling back to a default for
  that key.
- The validation **returns** its defects to the caller rather than rendering
  them, because a compiled-in registration has no file position of its own. The
  command surface supplies the position — that is CMD's job, not this item's.

Interface shape and function signatures are yours to choose; the members and the
validation semantics above are not.

### Acceptance criteria (as given)

- DocumentTypeRules carries per-type reference-bearing properties, typed
  reference properties (property → required target type), and pairings
  (**proves AC-SPECREF-004.1**, unit — cite per [[testing-standards]]).
- The Profile interface supplies the lifecycle registration: live statuses,
  retired statuses, and the successor property name.
- Activating a profile validates its registration and returns every defect
  found, each naming the offending key, with no fallback to a default for that
  key.
- All three defect classes are detected: an unregistered property named by a
  pairing or typed reference; a typed reference naming an undefined type; a
  lifecycle status outside the profile's vocabulary.
- Validation returns defects to the caller and renders nothing.
- The cli/core/corpus boundary test extends to the newly registered values, and
  no non-test source there carries retired, superseded_by, or any registered
  property name.
- [[module-document-core]] is folded: the Interface section records the widened
  Profile and DocumentTypeRules. Set version to 0.3 if not already set.
- [[module-spec-profile]] is folded: Interface records the widened Profile
  members. Set version to 0.3 if not already set.

## Retro

### Approach

Four changes, in order:

1. `cli/core/corpus/profile.go` — `DocumentTypeRules` gains
   `ReferenceProperties []string`, `TypedReferenceProperties map[string]string`
   and `Pairings []specmodel.PropertyPairing`; `Profile` gains
   `Lifecycle() LifecycleRules`; `LifecycleRules` is a new exported struct.
2. `cli/core/corpus/registration.go` (new) — `RegistrationDefect` and
   `ValidateRegistration(Profile) []RegistrationDefect`, plus two unexported
   helpers and a `sortedKeys` generic.
3. `cli/core/corpus/registration_test.go` (new) — the AC-SPECREF-004.1 citation
   and six tests over the three defect classes, the no-fallback rule, the
   renders-nothing rule, and determinism. `cli/core/corpus/profile_test.go` —
   the boundary test extended and renamed.
4. Folds: `module-document-core.md` Interface (already at version 0.3, left
   there), `module-spec-profile.md` Interface + Profile Registration table,
   version 0.2 → 0.3, `updated_at` → 2026-08-07.

`cli/profile` and the two `stubProfile` test doubles (`cli/core/corpus`,
`cli/cmd`) gained the `Lifecycle()` member. `cli/profile` returns the **zero**
registration — filling it is PREG (agent-plugins-9kb.5), explicitly out of
scope here.

### Implicit decisions, and why

- **The lifecycle registration carries the profile's whole status vocabulary
  (`LifecycleRules.Statuses`), not only the live and retired sets.** This is the
  one place where the shape I chose adds something the blueprint's three-member
  list does not literally name, and it is forced by an AC in the same bead: "a
  lifecycle status outside the profile's vocabulary" must be *detected*, and
  the host cannot detect it without being told what the vocabulary is. The union
  of live and retired is not the vocabulary — `schemas/profile.strata` says so
  explicitly ("a status in neither is one the lifecycle says nothing about,
  which is the right answer for a draft"). The alternative was deriving it from
  the union schema at `#/definitions/Status`, which bakes a schema-layout
  convention into the host and passes **vacuously** for any profile that keeps
  its status enum elsewhere — the exact failure AC-SPECREF-004.3 exists to
  prevent. Recorded in the `LifecycleRules.Statuses` doc comment and in
  module-document-core's Interface section so a later reader does not re-open it.
- **Statuses are `[]string`, not `[]specmodel.Status`.** The generated `Status`
  is this framework's six-value enum; putting it in the host would mean an
  adopting profile could not express a status of its own without failing
  `Status.Valid()`. That would worsen agent-plugins-9jo, which the bead forbids.
  Plain strings match the established pattern: `DocumentTypes()` is already keyed
  by `string`, not by `DocumentType`.
- **Pairings reuse `specmodel.PropertyPairing` rather than a corpus twin.** The
  interface already returns a generated type (`Conventions() specmodel.ConventionsConfig`),
  `PropertyPairing` is two property names and carries no framework vocabulary,
  and a parallel `corpus.PropertyPairing` would force PREG to convert between
  two identical shapes for nothing. Per adr-016 the Strata source is the single
  description of this surface.
- **One method, `Lifecycle()`, rather than three.** Three sibling methods would
  let a profile register retired statuses and no successor property with no
  single place to read "the lifecycle registration" as one value, and
  `ValidateRegistration` wants it as one value.
- **Validation lives in a new file, `registration.go`, not in `profile.go`.**
  `profile.go` is the declaration of the extension point; validation is behavior
  over it. Same package, so nothing about the surface changes.
- **`RegistrationDefect` has no `Error()` and no `String()`.** "Returns its
  defects rather than rendering them" is easier to hold if the type cannot
  render itself. `DocumentType` / `Key` / `Detail` is enough for CMD to compose
  a finding, and `Detail` deliberately does not repeat `Key`.
- **A profile registering lifecycle statuses but an empty vocabulary gets one
  defect per status**, rather than being waved through. An empty vocabulary
  containing nothing is the honest reading, and it keeps the check from passing
  vacuously on the registration most likely to be half-written. A profile
  registering *no* lifecycle at all has no defect — that is the zero value, and
  the checks that read it simply do not run.
- **Live/retired disjointness is NOT checked.** The blueprint says "the defects
  are exactly three"; a fourth class would be inventing scope.
- **The boundary test was renamed** `TestTheExtensionPointNamesNoConcreteDocumentType`
  → `TestTheExtensionPointNamesNoRegisteredVocabulary`, and now forbids three
  kinds of literal: document types, statuses, and reference-bearing property
  names. All three are read from the generated schema, never listed: statuses
  from `definitions.Status.enum`, reference-bearing properties as every property
  whose value is `$ref: Slug`, `items.$ref: Slug`, or a `oneOf` including it.
  That derivation yields exactly PREG's registered set (`consumers`,
  `contained_by`, `contracts_consumed`, `contracts_provided`, `frd`, `provider`,
  `sources`, `superseded_by`, `system`) with no transcription, so it will cover
  a property added to the schema before anything registers it.
- **`slug` is excluded from the derived property set**, with the reason stated
  in a named constant (`identityProperty`): it names the document declaring it
  rather than another document, so it is not reference-bearing, and
  `corpus.go`'s existing `const slugProperty = "slug"` is not a leak. Without
  the exclusion the extended boundary test fails on pre-existing, correct code.
- **The test asserts non-vacuity per kind**, not just overall: it fails if the
  schema yields zero document types, zero statuses, or zero reference properties.
- **`ValidateRegistration` is not called from `corpus` itself.** "WHEN a profile
  is activated" — activation happens in `cli/cmd`, and positioning the defect at
  `[profile].name` is CMD's job (agent-plugins-9kb.8). Exported and tested here.

### Gate results

All run as independent steps from `/home/user/agent-plugins`.

| Gate | Exit |
|---|---|
| `cd cli && gofmt -l .` | 0, printed nothing |
| `cd cli && go vet ./...` | 0 |
| `cd cli && go build ./...` | 0 |
| `cd cli && go test ./core/... ./profile/... ./cmd/...` | 0 |
| `cd cli && go test ./...` (whole module, for confidence) | 0 |

No `schemas/*.strata` changed, so `generate-schemas.sh` was not run and no
generated artifact moved. No regeneration diffstat for this bead.

### Anything surprising

- The generated models from MODEL (agent-plugins-9kb.1) already carry
  `ProfileRegistration.{live_statuses,retired_statuses,successor_property}` and
  `DocumentTypeRegistration.{reference_properties,typed_reference_properties,pairings}`.
  Nothing in this bead needed a schema change — the Go-side extension point was
  the only thing missing, which is exactly how the MODEL/XPT split was drawn.
- The `cli/cmd` package has a **second** `stubProfile` (in `config_test.go`)
  with the same shape as the one in `cli/core/corpus`. Widening the interface
  meant editing both. Neither the bead nor the blueprint mentions the second.
- The `slug` exclusion above is the only place where the extended boundary test
  needed a hand-written literal. It is worth watching: if a future envelope
  property is both slug-typed and legitimately known to the host, the same
  tension recurs.
