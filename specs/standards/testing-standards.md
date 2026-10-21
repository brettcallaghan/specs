---
title: Testing Standards for the Spec CLI
description: Unit testing for components, harness-independent snapshot
  cases for observable behavior, and the coverage obligations that tie
  both back to acceptance criteria
slug: testing-standards
type: standards
status: approved
---

# Testing standards for the Spec CLI

These standards govern how [[container-spec-cli]] and the modules it
composes ([[module-document-core]] and [[module-spec-profile]]) are
tested. They apply to every change that touches those three documents'
subjects, and they're the authority that the Verification gates in each
blueprint defer to.

The organizing idea is that **observable behavior is tested as data, not
as code**. A snapshot case describes an input corpus and the artifacts
the tool must produce from it, with no test code inside it. That's what
lets a future HTTP or MCP harness replay the same cases without
rewriting them, and it's why the case format is specified here instead
of being left to the test framework.

---

## Test layers

### Unit tests

Each component declared in a module blueprint SHALL carry unit tests
covering its exported surface.

**Rationale:** The component roster in a Module Blueprint is the unit of
architectural review, so it's also the unit of test accounting. A
component with no tests is visible as a gap without reading the code.

---

Unit tests SHALL reside beside their subject.

**Rationale:** It's Go convention, and it keeps a component's tests
inside the `source_paths` its blueprint declares, so the targeted
verification gate for that module selects them without a separate path
list.

---

Unit tests SHALL cite the acceptance criteria they verify using the
citation format defined in [[ac-traceability-standards]].

---

### Snapshot cases

The observable behavior of the CLI SHALL be verified by snapshot cases
rather than by unit tests of the command surface.

**Rationale:** The command surface is deliberately thin
([[container-spec-cli]]), so unit-testing it tests argument plumbing
instead of behavior. The behavior worth asserting is what a corpus
produces, which is exactly what a snapshot case captures.

---

A snapshot case SHALL consist of input files and expected artifacts, and
SHALL NOT contain executable test code.

**Rationale:** This is the requirement that buys harness independence.
A case containing Go test code can only be replayed by a Go test runner.
A case that's pure data can be replayed by an API harness, an MCP
server, or a conformance suite for a third-party reimplementation.

---

## Snapshot case structure

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

`manifest.json` holds the index that FEAT-SPEC-002 of
[[frd-spec-index]] defines. These standards use the file name the
artifact is written under. Where this document says *manifest* of an
expected artifact and *case manifest* of `case.toml`, the qualifier is
load-bearing.

A run produces two artifacts, and a case SHALL carry an expectation for
each one it exercises.

**Rationale:** The two artifacts fail independently. A change can leave
every finding correct while corrupting the manifest, or produce a sound
manifest while misreporting severity. Folding both into one expected
file would let either failure mask the other.

---

WHERE an argument vector produces only one of the two artifacts, the
case SHALL omit the other rather than record an empty file.

**Rationale:** An empty `checks.log` is a meaningful expectation: it
asserts a clean corpus. An absent one asserts that the command doesn't
report checks at all. Conflating the two loses a distinction the suite
exists to catch.

---

A case manifest SHALL declare the acceptance criteria the case verifies.

**Rationale:** A snapshot case has no code, so it has nowhere to put the
line comment that [[ac-traceability-standards]] requires of a test. The
manifest is that citation in data form, and it's what makes a
data-driven case traceable at all.

---

A case manifest SHALL declare the argument vector the case is run with.

**Rationale:** The same input corpus produces different artifacts under
`check` and under `index`. Without the argument vector in the manifest,
a harness can't replay the case, which defeats the purpose of keeping
cases free of test code.

---

An input corpus SHALL contain its own configuration file.

**Rationale:** A case that inherits configuration from the repository it
happens to live in isn't portable, and its behavior would change when
this repository's own `specs.toml` changes.

---

A snapshot case SHALL NOT record acceptance-criteria identifiers in the
frontmatter of an input document.

**Rationale:** An input document is a specification document under test.
Adding a property the active profile doesn't define makes it
non-conforming, so the case would assert a finding it introduced itself.
The manifest exists precisely so that the citation lives outside the
fixture.

---

### Snapshots

A script that regenerates every expected artifact from the current
implementation SHALL be provided.

**Rationale:** Without one, a legitimate output change is edited by hand
across every affected case, which is slow and silently lossy.

---

WHEN a regenerated snapshot differs from the committed one, the author
SHALL confirm the difference was intended before committing it.

**Rationale:** A snapshot suite's whole value is that an unintended
behavior change shows up as a diff. Regenerating without reading the
diff turns the suite from a safety net into a rubber stamp.

---

A change that regenerates snapshots SHALL NOT also make unrelated
modifications.

**Rationale:** A reviewer can judge a snapshot diff only if the change
that produced it is the only change in view.

---

Expected artifacts SHALL be byte-identical across Linux, macOS, and
Windows for identical input.

**Rationale:** AC-SPEC-009.7 of [[frd-spec-index]] requires the
CLI to report the same findings in the same order on every platform.
Snapshot artifacts are how that requirement is verified rather than
asserted.

---

Expected artifacts SHALL NOT contain timestamps, absolute paths,
host names, durations, or any other value that varies between runs.

**Rationale:** A non-deterministic artifact produces a diff on every
run, which trains reviewers to regenerate without reading. That defeats
the requirement three above.

---

WHEN findings are emitted, the CLI SHALL order them deterministically.

**Rationale:** Ordering is the most common source of spurious snapshot
diffs, and it's the CLI's responsibility instead of the harness's,
because a consumer parsing the JSON depends on it too.

---

## Coverage

### Statement coverage

Statement coverage SHALL be measured across every test layer and
reported as a single merged profile.

**Rationale:** The layers reach different code. Reporting one of them
alone understates what is tested, and reporting them side by side leaves
the union uncomputed. The union is the only figure that answers "is this
statement tested at all?"

---

The end-to-end layer SHALL be measured through the binary it runs,
rather than exempted from measurement because it crosses a process
boundary.

**Rationale:** `go test -coverprofile` observes only the test process,
so the layer that verifies exit statuses and the argument vector would
report as zero coverage, and its subject would appear untested. A binary
built with `go build -cover`, writing counters to `GOCOVERDIR`, closes
that gap.

---

Machine-generated source SHALL be excluded from the reported figures,
and the exclusion SHALL be stated wherever the figures are published.

**Rationale:** Generated code is a function of its generator and its
schema, not of the tests, and it's large enough here to dominate the
total. Excluding it silently would be indistinguishable from having a
better-tested codebase, which is why the exclusion travels with the
number.

---

The reported figures SHALL be committed, and CI SHALL fail when they do
not match a fresh measurement.

**Rationale:** A coverage figure regenerated only on demand records
whenever it was last convenient to look, so a drop arrives without an
author. A committed figure makes the change visible in the diff that
caused it.

---

### Rule coverage

Every check the CLI implements SHALL be exercised by at least one
snapshot case.

**Rationale:** A check with no case can be broken by a refactor without
any test failing. This is the completeness property that runs from the
implementation outward.

---

### Acceptance-criteria traceability

Every acceptance criterion in an `approved` or `active` FRD SHALL be
cited by at least one snapshot case manifest or at least one unit test.

**Rationale:** This is the complement of rule coverage: the property
runs from the specification inward, and catches a criterion nobody
implemented as well as one nobody tested.

---

A map from rule identifier to acceptance-criteria identifiers SHALL be
maintained.

**Rationale:** AC-SPEC-006.1 of [[frd-spec-index]] already requires
every finding to carry a stable rule identifier, so this is one table
instead of per-case metadata. It's what makes the cross-check below
possible.

---

WHERE an acceptance criterion is verified by a finding, the coverage
check SHALL confirm that a citing case's expected artifacts contain a
finding whose rule maps to that criterion.

**Rationale:** A manifest is a claim, and an unverified claim is how
coverage reports come to overstate what is tested. This is the check
that catches a case citing a criterion its output doesn't exercise.

**Note:** Not every criterion is verified by a finding. Exit status,
JSON emission, index content, corpus queries, configuration resolution,
and latency are verified by other artifacts or by unit tests, and the
cross-check doesn't apply to them. Their citation in a manifest or a
test comment stands on its own.

---

IF an acceptance criterion carries the `(manual)` marker defined by
`plugin/skills/spec/references/frd-template.md`, THEN it is exempt from
the citation obligation above.

---

## References

- [[frd-spec-index]] — the acceptance criteria these standards
  make traceable
- [[container-spec-cli]] — the container whose behavior snapshot cases
  verify
- [[module-document-core]] and [[module-spec-profile]] — the modules
  whose components carry unit tests
- [[ac-traceability-standards]] — the citation format unit tests use,
  which the case manifest mirrors
- `plugin/skills/spec/references/frd-template.md` — the `(manual)`
  marker and the acceptance-criteria identifier format
