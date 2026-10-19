<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/blueprint-readiness-checklist.md (version unversioned)
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
---

# Blueprint readiness checklist

## Contents

- Readiness checks 1 to 3, "is grooming premature?": FRD status,
  blueprint status, and Open questions
- Content checks 4 to 10, "is the approved design plannable?": AC
  Coverage rows, `[[slug]]` resolution, component references, design
  intent, delta keys, lockstep rules, and Tracker reconciliation
- The decision sweep: the authoring preflight for check 7, and which
  notation a delta's behavior lines use

This checklist has two consumers. The `spec` skill runs these checks as a
preflight before a feature blueprint is set to `ready-for-review`. The
`groom-feature-blueprint` skill enforces them as its Phase 1 gate before
it derives a work-item graph. The checks are defined only here. The skills
reference this document rather than restating it, so authoring and
grooming can't disagree about what "groomable" means.

Each skill, not this document, defines what happens when a check fails,
and the two consumers differ. At authoring time, a failed check is
unfinished work. At groom time, a failed readiness check stops grooming
without modifying any document. A failed content check bounces the gap
into the blueprint's Open questions and demotes the blueprint to `draft`.

---

## Readiness checks — is grooming premature?

1. The FRD's status is `approved` or `active`.
2. The blueprint's status is `approved`.
3. The blueprint's Open questions section is absent, empty, or states
   only "None."

At authoring time these translate to a handoff obligation: WHEN a
feature blueprint is submitted for review, the author SHALL flag BOTH
status flips to the human reviewer — the blueprint's approval AND its
FRD's — so the FRD gate is not discovered for the first time at groom
time.

## Content checks — is the approved design plannable?

4. Every AC defined in the FRD appears at least once in the blueprint's
   AC coverage table. Every row names at least one delta block that
   exists in the Element deltas section. A row is invalid if it names
   only a relied-upon element, a sequencing constraint, or a Definition
   of Done entry. A row is also invalid if it names more than one delta
   without marking exactly one as the proving delta (see the AC coverage
   row validity rules in the feature blueprint template).

   Enumerate the FRD's ACs with `specs feature <id>` rather than by
   reading the FRD. An acceptance criterion is a tagged Gherkin
   `Scenario` ([[adr-021]]), so the set is a query. Reading for them by
   eye misses the case the notation handles worst: several `Examples`
   blocks under one outline. Each tagged block is its own criterion, but
   only the first looks like a criterion in a diff.
5. Every `[[slug]]` reference resolves to an existing document.
6. Every component reference resolves to exactly one declaration across
   the referenced containers' and modules' Components sections. It
   resolves by name, qualified in prose where the name is ambiguous (see
   element declarations in `c4-guide.md`).
7. Every delta contains enough design intent to implement without
   making a design decision the blueprint doesn't record.
8. The AC coverage legend assigns every delta block a unique short key,
   and every legend key expands to an existing delta block. No key from
   an earlier revision has been reused for a different delta. These keys
   are the stable identity work orders carry through grooming and
   re-grooming.
9. This check covers each lockstep-update rule in the owning blueprints'
   Change guidance whose coupled artifacts span more than one delta
   block. For each such rule, the blueprint states how the rule is
   satisfied (in the delta or in Sequencing constraints). Reconciling
   such a rule is a design decision, so a missing reconciliation is a
   content gap, not a groom-time judgment call.
10. The Tracker reconciliation section exists, and it either classifies
    every listed issue as subsumed, prerequisite, or out-of-scope, or
    states "None." When the project has no tracker, "None." also
    satisfies the check.

### The decision sweep — authoring preflight for check 7

Check 7 is the check most likely to fail late, during grooming or
implementation, because unrecorded decisions are invisible until someone
has to make them. Before you set `ready-for-review`, walk through each
delta and ask what an implementer would have to choose there. These are
typical silent decisions:

- file and artifact paths and names
- default values and default resolution bases (cwd versus discovered
  root)
- runtime versus build-time access to generated artifacts
- v1 scope cuts (what is deliberately not handled)
- behavior at half-valid inputs (presence versus validity)
- coexistence rules (multiple instances of X at once)

Pin each one as an EARS behavior line in the delta, or raise it as an
Open Question. Implementation structure (function signatures, file
layout, algorithms) stays unpinned. The sweep is for observable behavior
and on-disk surfaces only. The blueprint template's "anchors, not
signatures" rule guards against pinning too much, and this sweep guards
against pinning too little.

A delta's behavior lines are EARS, not Gherkin. A feature blueprint holds
no features and no acceptance criteria. Those live in its FRD, in tagged
Gherkin blocks, and a blueprint that restates them duplicates the
document it traces to. [[adr-002]] governs everything a blueprint writes.
[[adr-021]] governs nothing in it.

## References

- `feature-blueprint-template.md` (same directory): the document these
  checks apply to, including the AC coverage row validity rules
- `c4-guide.md`: element declarations and reference rules
- [[ac-traceability-standards]]: the test citation obligations the AC
  Coverage table feeds
- [[gherkin-guide]]: the notation the FRD's acceptance criteria are
  written in, which check 4 enumerates against
