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
composes, [[module-document-core]] and [[module-spec-profile]], are tested.
They apply to every change that touches the subjects of those three
documents. The Verification gates in each blueprint defer to them.

The organizing idea is that observable behavior is tested as data, not as
code. A snapshot case describes an input corpus and the artifacts the tool
must produce from it, with no test code inside it. A future HTTP or MCP
harness can then replay the same cases without rewriting them. That's why
these standards specify the case format instead of leaving it to the test
framework.

---

## Test layers

### Unit tests

Each component declared in a module blueprint SHALL carry unit tests
covering its exported surface.

**Rationale:** the component roster in a module blueprint is the unit of
architectural review, so it's also the unit of test accounting. A component
with no tests shows up as a gap without anyone reading the code.

---

Unit tests SHALL reside beside their subject.

**Rationale:** it's the Go convention. It also keeps a component's tests
inside the `source_paths` that its blueprint declares, so the targeted
verification gate for that module selects them without a separate path list.

---

Unit tests SHALL cite the acceptance criteria they verify using the
citation format defined in
`plugin/skills/spec/references/ac-traceability-standards.md`.

---

### Snapshot cases

The observable behavior of the CLI SHALL be verified by snapshot cases
rather than by unit tests of the command surface.

**Rationale:** the command surface is deliberately thin
([[container-spec-cli]]), so a unit test of it tests argument plumbing, not
behavior. The behavior worth asserting is what a corpus produces, and that's
what a snapshot case captures.

---

A snapshot case SHALL consist of input files and expected artifacts, and
SHALL NOT contain executable test code.

**Rationale:** this requirement is what makes a case independent of its
harness. Only a Go test runner can replay a case that contains Go test code.
An API harness, an MCP server, or a conformance suite for a third-party
reimplementation can replay a case that's pure data.

---

## Snapshot case structure

### Layout

Each snapshot case SHALL be a directory containing a manifest, an input
corpus, and the expected artifacts.

The case directory has this layout:

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

`manifest.json` holds the index that FEAT-SPEC-002 of [[frd-spec-index]]
defines. These standards call each artifact by the file name it's written
under. This document says *manifest* for the expected artifact and *case
manifest* for `case.toml`, and the qualifier is what tells the two apart.

A run produces two artifacts, and a case SHALL carry an expectation for
each one it exercises.

**Rationale:** the two artifacts fail independently. A change can leave
every finding correct while corrupting the manifest, or produce a sound
manifest while misreporting severity. Combining both in one expected file
would let either failure mask the other.

---

WHERE an argument vector produces only one of the two artifacts, the
case SHALL omit the other rather than record an empty file.

**Rationale:** an empty `checks.log` is a meaningful expectation: it asserts
a clean corpus. An absent one asserts that the command doesn't report checks
at all. Treating the two as the same loses a distinction the suite exists to
catch.

---

A case manifest SHALL declare the acceptance criteria the case verifies.

**Rationale:** a snapshot case has no code, so it has nowhere to put the
line comment that
`plugin/skills/spec/references/ac-traceability-standards.md` requires of
a test. The manifest is that citation in data form, and it's what makes a
data-driven case traceable at all.

---

A case manifest SHALL declare the argument vector the case is run with.

**Rationale:** the same input corpus produces different artifacts under
`check` and under `index`. Without the argument vector in the manifest, a
harness can't replay the case, which defeats the purpose of keeping cases
free of test code.

---

An input corpus SHALL contain its own configuration file.

**Rationale:** a case that inherits configuration from the repository it
happens to live in isn't portable. Its behavior would change whenever this
repository's own `specs.toml` changes.

---

A snapshot case SHALL NOT record acceptance-criteria identifiers in the
frontmatter of an input document.

**Rationale:** an input document is a specification document under test.
Adding a property the active profile doesn't define makes it
non-conforming, so the case would assert a finding it introduced itself.
The manifest exists so that the citation lives outside the fixture.

---

### Snapshots

A script that regenerates every expected artifact from the current
implementation SHALL be provided.

**Rationale:** without one, an author edits a legitimate output change by
hand across every affected case, which is slow and silently lossy.

---

WHEN a regenerated snapshot differs from the committed one, the author
SHALL confirm the difference was intended before committing it.

**Rationale:** a snapshot suite is valuable because an unintended behavior
change shows up as a diff. If authors regenerate without reading the diff,
the suite approves every change instead of catching the unintended ones.

---

A change that regenerates snapshots SHALL NOT also make unrelated
modifications.

**Rationale:** a reviewer can only judge a snapshot diff if the change
that produced it is the only change in view.

---

Expected artifacts SHALL be byte-identical across Linux, macOS, and
Windows for identical input.

**Rationale:** FEAT-SPEC-009.7 of [[frd-spec-index]] requires the CLI to
report the same findings in the same order on every platform. Snapshot
artifacts verify that requirement instead of only asserting it.

---

Expected artifacts SHALL NOT contain timestamps, absolute paths,
host names, durations, or any other value that varies between runs.

**Rationale:** a non-deterministic artifact produces a diff on every run.
That trains reviewers to regenerate without reading, which defeats the
requirement two above.

---

WHEN findings are emitted, the CLI SHALL order them deterministically.

**Rationale:** ordering is the most common source of spurious snapshot
diffs. It's the CLI's responsibility, not the harness's, because a consumer
that parses the JSON depends on it too.

---

## Coverage

### Statement coverage

Statement coverage SHALL be measured across every test layer and
reported as a single merged profile.

**Rationale:** the layers reach different code. Reporting one of them alone
understates what's tested. Reporting them side by side leaves the union
uncomputed, and only the union answers "is this statement tested at all."

---

The end-to-end layer SHALL be measured through the binary it runs,
rather than exempted from measurement because it crosses a process
boundary.

**Rationale:** `go test -coverprofile` observes only the test process. The
layer that verifies exit statuses and the argument vector would report zero
coverage, and its subject would look untested. A binary built with
`go build -cover` that writes counters to `GOCOVERDIR` closes that gap.

---

Machine-generated source SHALL be excluded from the reported figures,
and the exclusion SHALL be stated wherever the figures are published.

**Rationale:** generated code is a function of its generator and its schema,
not of the tests, and it's large enough here to dominate the total. Excluding
it silently would look the same as having a better-tested codebase. That's
why the exclusion is stated with the number.

---

The reported figures SHALL be committed, and CI SHALL fail when they do
not match a fresh measurement.

**Rationale:** a coverage figure regenerated only on demand records the last
time someone found it convenient to look, so a drop arrives with no change to
attribute it to. A committed figure makes the change visible in the diff that
caused it.

---

### Rule coverage

Every check the CLI implements SHALL be exercised by at least one
snapshot case.

**Rationale:** a check with no case can be broken by a refactor without
any test failing. This is the completeness property that runs from the
implementation outward.

---

### Acceptance-criteria traceability

Every acceptance criterion in an `approved` or `active` FRD SHALL be
cited by at least one snapshot case manifest or at least one unit test.

**Rationale:** this is the complement of rule coverage. It runs from the
specification inward, and it catches a criterion nobody implemented as well
as one nobody tested.

---

A map from rule identifier to acceptance-criteria identifiers SHALL be
maintained.

**Rationale:** AC-SPEC-006.1 of [[frd-spec-index]] already requires every
finding to carry a stable rule identifier, so the map is one table instead
of per-case metadata. It's what makes the cross-check below possible.

---

WHERE an acceptance criterion is verified by a finding, the coverage
check SHALL confirm that a citing case's expected artifacts contain a
finding whose rule maps to that criterion.

**Rationale:** a manifest is a claim, and unverified claims are how coverage
reports come to overstate what's tested. This check catches a case that cites
a criterion its output doesn't exercise.

**Note:** not every criterion is verified by a finding. Other artifacts or
unit tests verify exit status, JSON emission, index content, corpus queries,
configuration resolution, and latency. The cross-check doesn't apply to them,
so their citation in a manifest or a test comment stands on its own.

---

IF an acceptance criterion carries the `(manual)` marker defined by
`plugin/skills/spec/references/frd-template.md`, THEN it is exempt from
the citation obligation above.

---

## References

- [[frd-spec-index]]: the acceptance criteria these standards make
  traceable
- [[container-spec-cli]]: the container whose behavior snapshot cases
  verify
- [[module-document-core]] and [[module-spec-profile]]: the modules whose
  components carry unit tests
- `plugin/skills/spec/references/ac-traceability-standards.md`: the citation
  format unit tests use, which the case manifest mirrors
- `plugin/skills/spec/references/frd-template.md`: the `(manual)` marker and
  the acceptance-criteria identifier format
