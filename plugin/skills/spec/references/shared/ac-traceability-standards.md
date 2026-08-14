<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/ac-traceability-standards.md (version 0.3.0)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: AC Traceability Standards
description: How tests cite the acceptance criteria they verify, and how
  coverage is verified against active FRDs over time
slug: ac-traceability-standards
type: standards
status: active
updated_at: 2026-08-13T00:00:00Z
version: "0.3.0"
---

# AC Traceability Standards

## Contents

- Citation Format — how a test names the AC it verifies
- Coverage Obligations — which ACs need citations, at which level
- Lifecycle and Verification — keeping citations correct as ACs change
- References

Requirements binding automated tests to the acceptance criteria (ACs)
they verify. They apply to every project adopting this framework's FRDs
and feature blueprints.

The durable traceability pair is the **FRD** and the **annotated test
suite**. A feature blueprint's AC Coverage table declares intent at
design time and is reconciled at merge time (its Definition of Done),
but the blueprint retires; long-term coverage verification is a function
of active FRDs and test citations only. AC identifiers are never reused
(see `frd-template.md` ID conventions), which is what makes this
sound.

---

AC identifiers SHALL be append-only: an FRD SHALL NOT renumber an
existing acceptance criterion's identifier. WHEN a new criterion is
inserted conceptually between two existing identifiers, the FRD SHALL
assign it the next unused number in that AC's sequence rather than
shifting the surrounding identifiers (inserting between AC-SPEC-006.4
and .5 assigns the new criterion the next unused number; it does not
renumber .5 onward).

**Rationale:** A renumbered identifier silently retargets every existing
test citation and work-order source that pointed at it to different
criteria, and it makes any identifier-keyed delta between two versions
of an FRD meaningless.

---

## Citation Format

WHEN a test verifies an acceptance criterion, a citation comment SHALL
be placed on the line(s) immediately preceding the test declaration.

**Rationale:** Adjacent placement binds the citation to one specific
test, so tooling can attribute coverage without language-aware parsing.

---

A citation comment SHALL use the marker `AC:` followed by one or more
comma-separated AC identifiers, in the host language's line-comment
syntax:

```python
# AC: AC-CHK-003.2
def test_declined_card_returns_user_to_payment_form():
    ...
```

```typescript
// AC: AC-CHK-003.1, AC-CHK-003.3
it("shows order number and delivery date after payment", () => {
```

**Rationale:** A fixed marker makes citations machine-greppable:
`AC:\s*AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+`. Free-form mentions cannot be
verified.

---

AC identifiers in citations SHALL be verbatim copies of the identifiers
defined in the FRD.

---

A test MAY cite multiple acceptance criteria, and an acceptance
criterion MAY be cited by multiple tests.

---

WHEN a test is moved or renamed, its citation comment SHALL move with
the test declaration.

**Rationale:** Adjacent comments are the citation form most easily
orphaned by refactoring; this obligation, plus the verification tooling
below, is the mitigation.

---

## Coverage Obligations

Every acceptance criterion of an FRD SHALL be cited by at least one
test before the implementing change merges.

**Rationale:** This is the merge-time reconciliation of the feature
blueprint's AC Coverage table — the point where declared intent becomes
verifiable reality.

---

A test citing an acceptance criterion SHOULD verify it at the level
declared in the feature blueprint's AC Coverage table
(unit | integration | e2e | manual).

**Rationale:** The declared level is a design-time judgment about what
proof suffices. It is SHOULD, not SHALL, because the blueprint retires
and the level is not durably machine-checkable afterward.

---

IF an acceptance criterion can only be verified manually, THEN the FRD
SHALL record that designation by appending the marker `(manual)` to the
criterion text, and the criterion is exempt from the citation
obligation.

---

## Lifecycle and Verification

WHEN an acceptance criterion is retired in its FRD, citations of it
SHALL be removed or updated in the same change.

The test suite SHALL NOT cite an AC identifier that no `approved` or
`active` FRD defines.

**Rationale:** Stale citations are worse than none — they report
coverage of requirements that no longer exist. A retired FRD still
contains its identifiers, which is why existence is not the test:
citations must track the FRDs currently in force. `approved` is
included so tests written on a feature branch, before its FRD takes
effect, are not flagged.

---

WHEN verification tooling runs (CI or on demand), it SHOULD report:

- **Uncovered ACs** — defined in an active FRD, cited by no test
- **Orphaned citations** — citing an identifier no approved or active
  FRD defines

IF an implementing change would merge with uncovered ACs for its
feature, THEN the verification SHALL fail.

**Rationale:** Coverage is verifiable forever from two living artifacts
(active FRDs and the test suite), with no dependency on retired feature
blueprints.

---

## References

- `frd-template.md` (same directory) — AC identifier format, no-reuse
  rule, and the `(manual)` designation (templates are not documents
  and have no slugs; they are referenced by path)
- `feature-blueprint-template.md` (same directory) — AC Coverage
  table and Definition of Done
- [[template-authoring-standards]] — EARS and referencing conventions
