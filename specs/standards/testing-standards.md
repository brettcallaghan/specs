---
title: Testing Standards for the Spec CLI
description: Unit testing for components, harness-independent snapshot
  cases for observable behavior, and the coverage obligations that tie
  both back to acceptance criteria
slug: testing-standards
type: standards
status: approved
updated_at: 2026-07-31T00:00:00Z
version: "0.1"
---

# Testing Standards for the Spec CLI

These standards govern how [[container-spec-cli]] and the modules it
composes — [[module-document-core]] and [[module-spec-profile]] — are
tested. They apply to every change touching those three documents'
subjects, and they are the authority the Verification gates in each
blueprint defer to.

The organizing idea is that **observable behavior is tested as data, not
as code**. A snapshot case describes an input corpus and the artifacts
the tool must produce from it, with no test code inside it. That is what
allows the same cases to be replayed by a future HTTP or MCP harness
without rewriting them, and it is why the case format is specified here
rather than left to the test framework.

---

## Test Layers

### Unit Tests

Each component declared in a module blueprint SHALL carry unit tests
covering its exported surface.

**Rationale:** the component roster in a module blueprint is the unit of
architectural review, so it is also the unit of test accounting. A
component with no tests is visible as a gap without reading the code.

---

Unit tests SHALL reside beside their subject.

**Rationale:** Go convention, and it keeps a component's tests inside the
`source_paths` its blueprint declares, so the targeted verification gate
for that module selects them without a separate path list.

---

Unit tests SHALL cite the acceptance criteria they verify using the
citation format defined in
`plugin/skills/spec/references/ac-traceability-standards.md`.

---

### Snapshot Cases

The observable behavior of the CLI SHALL be verified by snapshot cases
rather than by unit tests of the command surface.

**Rationale:** the command surface is deliberately thin
([[container-spec-cli]]), so unit-testing it tests argument plumbing
rather than behavior. The behavior worth asserting is what a corpus
produces, which is exactly what a snapshot case captures.

---

A snapshot case SHALL consist of input files and expected artifacts, and
SHALL NOT contain executable test code.

**Rationale:** this is the requirement that buys harness independence. A
case containing Go test code can only be replayed by a Go test runner; a
case that is pure data can be replayed by an API harness, an MCP server,
or a conformance suite for a third-party reimplementation.

---

## Snapshot Case Structure

### Layout

Each snapshot case SHALL be a directory containing a manifest, an input
corpus, and the expected artifacts.

The layout is:

```
cli/testdata/cases/<case-name>/
  case.toml               the case manifest — description, cited
                          criteria, argument vector
  input/                  a complete corpus, including its own specs.toml
  expected/
    checks.log            the check results the run must report
    manifest.json         the corpus manifest the run must produce,
                          whole or partial
```

`manifest.json` holds the index FEAT-SPEC-002 of
[[frd-spec-validation-cli]] defines. These standards use the file name
the artifact is written under; where this document says *manifest* of an
expected artifact and *case manifest* of `case.toml`, the qualifier is
load-bearing.

A run produces two artifacts, and a case SHALL carry an expectation for
each one it exercises.

**Rationale:** the two artifacts fail independently. A change can leave
every finding correct while corrupting the manifest, or produce a sound
manifest while mis-reporting severity. Folding both into one expected
file would let either failure mask the other.

---

WHERE an argument vector produces only one of the two artifacts, the
case SHALL omit the other rather than record an empty file.

**Rationale:** an empty `checks.log` is a meaningful expectation — it
asserts a clean corpus. An absent one asserts the command does not
report checks at all. Conflating the two loses a distinction the suite
exists to catch.

---

A case manifest SHALL declare the acceptance criteria the case verifies.

**Rationale:** a snapshot case has no code, so it has nowhere to put the
line comment that
`plugin/skills/spec/references/ac-traceability-standards.md` requires of
a test. The manifest is that citation in data form, and it is what makes
a data-driven case traceable at all.

---

A case manifest SHALL declare the argument vector the case is run with.

**Rationale:** the same input corpus produces different artifacts under
`check` and under `index`. Without the argument vector in the manifest,
a harness cannot replay the case, which defeats the purpose of keeping
cases free of test code.

---

An input corpus SHALL contain its own configuration file.

**Rationale:** a case that inherits configuration from the repository it
happens to live in is not portable, and would change behavior when this
repository's own `specs.toml` changes.

---

A snapshot case SHALL NOT record acceptance-criteria identifiers in the
frontmatter of an input document.

**Rationale:** an input document is a specification document under test.
Adding a property the active profile does not define makes it
non-conforming, so the case would assert a finding it introduced itself.
The manifest exists precisely so the citation lives outside the fixture.

---

### Snapshots

A script that regenerates every expected artifact from the current
implementation SHALL be provided.

**Rationale:** without one, a legitimate output change is edited by hand
across every affected case, which is slow and silently lossy.

---

WHEN a regenerated snapshot differs from the committed one, the author
SHALL confirm the difference was intended before committing it.

**Rationale:** a snapshot suite's whole value is that an unintended
behavior change shows up as a diff. Regenerating without reading the
diff converts the suite from a safety net into a rubber stamp.

---

A change that regenerates snapshots SHALL NOT also make unrelated
modifications.

**Rationale:** a reviewer can only judge a snapshot diff if the change
that produced it is the only change in view.

---

Expected artifacts SHALL be byte-identical across Linux, macOS, and
Windows for identical input.

**Rationale:** FEAT-SPEC-009.7 of [[frd-spec-validation-cli]] requires the
CLI to report the same findings in the same order on every platform.
Snapshot artifacts are how that requirement is verified rather than
asserted.

---

Expected artifacts SHALL NOT contain timestamps, absolute paths,
host names, durations, or any other value that varies between runs.

**Rationale:** a non-deterministic artifact produces a diff on every run,
which trains reviewers to regenerate without reading — defeating the
requirement two above.

---

WHEN findings are emitted, the CLI SHALL order them deterministically.

**Rationale:** ordering is the most common source of spurious snapshot
diffs, and it is the CLI's responsibility rather than the harness's,
because a consumer parsing the JSON depends on it too.

---

## Coverage

### Rule Coverage

Every check the CLI implements SHALL be exercised by at least one
snapshot case.

**Rationale:** a check with no case can be broken by a refactor without
any test failing. This is the completeness property that runs from the
implementation outward.

---

### Acceptance-Criteria Traceability

Every acceptance criterion in an `approved` or `active` FRD SHALL be
cited by at least one snapshot case manifest or at least one unit test.

**Rationale:** the complement of rule coverage — this property runs from
the specification inward, and catches a criterion nobody implemented as
well as one nobody tested.

---

A map from rule identifier to acceptance-criteria identifiers SHALL be
maintained.

**Rationale:** AC-SPEC-006.1 of [[frd-spec-validation-cli]] already
requires every finding to carry a stable rule identifier, so this is one
table rather than per-case metadata. It is what makes the cross-check
below possible.

---

WHERE an acceptance criterion is verified by a finding, the coverage
check SHALL confirm that a citing case's expected artifacts contain a
finding whose rule maps to that criterion.

**Rationale:** a manifest is a claim, and an unverified claim is how
coverage reports come to overstate what is tested. This is the check
that catches a case citing a criterion its output does not exercise.

**Note:** not every criterion is verified by a finding. Exit status,
JSON emission, index content, corpus queries, configuration resolution
and latency are verified by other artifacts or by unit tests, and the
cross-check does not apply to them — their citation in a manifest or a
test comment stands on its own.

---

IF an acceptance criterion carries the `(manual)` marker defined by
`plugin/skills/spec/references/frd-template.md`, THEN it is exempt from
the citation obligation above.

---

## References

- [[frd-spec-validation-cli]] — the acceptance criteria these standards
  make traceable
- [[container-spec-cli]] — the container whose behavior snapshot cases
  verify
- [[module-document-core]] and [[module-spec-profile]] — the modules
  whose components carry unit tests
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  citation format unit tests use, which the case manifest mirrors
- `plugin/skills/spec/references/frd-template.md` — the `(manual)`
  marker and the acceptance-criteria identifier format
