<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: plugin/skills/spec/references/blueprint-readiness-checklist.md (version 0.5.0)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Blueprint Readiness Checklist
description: The single definition of "ready to groom" for a feature
  blueprint — run by the spec skill as an authoring preflight and
  enforced by groom-feature-blueprint as its Phase 1 gate
slug: blueprint-readiness-checklist
type: standards
status: active
updated_at: 2026-07-29T00:00:00Z
version: "0.5.0"
---

# Blueprint Readiness Checklist

## Contents

- Readiness checks #1–#3 — is grooming premature? (FRD status, blueprint
  status, Open Questions)
- Content checks #4–#10 — is the approved design plannable? (AC Coverage
  rows, `[[slug]]` resolution, component references, design intent, delta
  keys, lockstep rules, Tracker Reconciliation)
- The decision sweep — the authoring preflight for check 7

One list, two consumers. The `spec` skill runs these checks as a
**preflight** before a feature blueprint is set to `ready-for-review`;
the `groom-feature-blueprint` skill enforces them as its **Phase 1 gate** before
deriving a work-item graph. The checks are defined only here — the
skills reference this document rather than restating it, so authoring
and grooming can never disagree about what "groomable" means.

What happens on failure differs by consumer and is defined in each
skill, not here: at authoring time a failed check is simply unfinished
work; at groom time a failed readiness check stops grooming without
modifying any document, and a failed content check bounces the gap
into the blueprint's Open Questions and demotes the blueprint to
`draft`.

---

## Readiness checks — is grooming premature?

1. The FRD's status is `approved` or `active`.
2. The blueprint's status is `approved`.
3. The blueprint's Open Questions section is absent, empty, or states
   only "None."

At authoring time these translate to a handoff obligation: WHEN a
feature blueprint is submitted for review, the author SHALL flag BOTH
status flips to the human reviewer — the blueprint's approval AND its
FRD's — so the FRD gate is not discovered for the first time at groom
time.

## Content checks — is the approved design plannable?

4. Every AC defined in the FRD appears at least once in the
   blueprint's AC Coverage table, and every row names at least one
   delta block that exists in the Element Deltas section. A row naming
   only a relied-upon element, a sequencing constraint, or a
   Definition of Done entry is invalid, and a row naming more than one
   delta without marking exactly one as the proving delta is invalid
   (see the AC Coverage row validity rules in the feature blueprint
   template).
5. Every `[[slug]]` reference resolves to an existing document; contracts
   listed in `creates` exist with status `draft`.
6. Every component reference resolves to exactly one declaration across
   the referenced containers' and modules' Components sections — by name,
   qualified in prose where ambiguous (see element declarations in
   `c4-guide.md`).
7. Every delta contains enough design intent to implement without
   making a design decision the blueprint does not record.
8. The AC Coverage legend assigns every delta block a unique short
   key, every legend key expands to an existing delta block, and no
   key from an earlier revision has been reused for a different delta.
   These keys are the stable identity work items carry through
   grooming and re-grooming.
9. For every lockstep-update rule in the owning blueprints' Change
   Guidance whose coupled artifacts span more than one delta block,
   the blueprint states how the rule is satisfied (in the delta or in
   Sequencing Constraints). Reconciling such a rule is a design
   decision — its absence is a content gap, not a groom-time judgment
   call.
10. The Tracker Reconciliation section exists and classifies every
    listed issue as subsumed, prerequisite, or out-of-scope — or
    states "None." (which also satisfies the check WHERE the project
    has no tracker).

### The decision sweep — authoring preflight for check 7

Check 7 is the check most likely to fail late — during grooming or
implementation — because unrecorded decisions are invisible until
someone has to make them. Before setting `ready-for-review`, walk each
delta asking: what would an implementer have to *choose* here?
Typical silent decisions:

- file/artifact paths and names
- default values and default resolution bases (cwd vs discovered root)
- runtime-vs-build-time access to generated artifacts
- v1 scope cuts (what is deliberately not handled)
- behavior at half-valid inputs (presence vs validity)
- coexistence rules (multiple instances of X at once)

For each: pin it as an EARS behavior line in the delta, or raise it as
an Open Question. Implementation *structure* (function signatures,
file layout, algorithms) stays unpinned — the sweep is for observable
behavior and on-disk surfaces only. The blueprint template's "anchors,
not signatures" rule guards against pinning too much; this sweep
guards against pinning too little.

## References

- `feature-blueprint-template.md` (same directory) — the document
  these checks apply to, including the AC Coverage row validity rules
- `c4-guide.md` — element declarations and reference rules
- [[ac-traceability-standards]] — the test citation obligations the AC
  Coverage table feeds
