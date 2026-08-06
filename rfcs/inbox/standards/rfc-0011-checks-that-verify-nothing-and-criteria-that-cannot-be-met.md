---
title: "RFC: Three checks that reported success while verifying nothing,
  and four acceptance criteria no mechanism could satisfy"
description: Build the acceptance-criteria coverage check this project's
  testing standard already mandates; make a harness that discovers
  nothing fail; require an acceptance criterion's declared verification
  level to be reachable; and record the generation-pipeline, rename and
  determinism constraints six work orders had to rediscover
slug: rfc-0011
type: rfc
status: ready-for-review
updated_at: 2026-08-05T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: "claude-opus-5"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.14.0"
---

# RFC: Checks that verify nothing, and criteria that cannot be met

## Summary

Sixteen work orders implementing [[feature-spec-validation-cli]] shipped
the Spec CLI end to end. Aggregating the 53 findings its implementers
recorded in flight produces two systemic classes, each with independent
instances that no single item could see.

**Checks that report success while verifying nothing.** Three of them,
found by three unrelated mechanisms: an entire item's snapshot suite
asserted only exit status for a full wave before a *later* item's
negative control exposed it; the shipped CLI answered queries over an
empty corpus because `specs.toml` still pointed at `docs/`; and the
acceptance-criteria coverage check that
[[testing-standards]] normatively mandates **was never built**, so an
acceptance criterion reached this feature's Definition of Done uncited
through fifteen closed items and a green self-application gate.

**Acceptance criteria written without checking that a mechanism exists
to satisfy them.** Four, three of which became beads rather than being
faked: a criterion requiring a `version` bump on a document type whose
schema declares no `version`; one requiring a gate to surface a file
outside the corpus root that gate scans; one declared at `integration`
level that no input corpus can reach; and one whose text is in tension
with the very decision the blueprint pinned to satisfy it.

The proposals build the missing coverage check, make vacuous discovery a
failure, and write down the constraints — pipeline, rename, determinism
— that six items each rediscovered alone.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Phase executed | Implemented the 16 work orders of epic `agent-plugins-bge` (Spec Corpus Index and Core Conformance Checks), stages 1–3, plus 5 beads discovered mid-run |
| Documents exercised | `specs/features/spec-index/spec-index-blueprint.md`, `specs/features/spec-index/spec-index-frd.md`, `specs/standards/testing-standards.md`, `specs/architecture/{module-document-core,module-spec-profile,container-spec-cli}.md`, `specs/contracts/contract-specs-configuration.md`, `schemas/README.md` |
| Documents produced | `cli/` (11 Go packages, 11 snapshot cases), `schemas/{index,finding}.strata`, `.github/workflows/{cli,release}.yml`; blueprint retired |
| Skill(s) followed | `bead-bonanza` v1.3, `retro-spec` v1.3.0 |
| Templates / standards read | `testing-standards` v0.1, `rfc-template.md` v1.1, `findings-capture-protocol.md` (vendored as `rfcs/findings-capture.md`) |
| Plugin version | 0.14.0 |
| Session date | 2026-08-04 – 2026-08-05 |
| Findings aggregated | 53, across 15 of the closure set's beads |

Two provenance caveats. `agent-plugins-bge.5` contributes **no**
findings: its implementing task was killed by a container restart after
the code was written but before its retro step, and the loss is recorded
on the bead. It is an absent data point, not a clean run. And the CI
workflows of `agent-plugins-bge.13` could not be executed in this
environment (no Actions runner, no Docker daemon), so findings about
them rest on static validation.

## Evidence

1. **An entire item's snapshot suite asserted nothing but exit status,
   and stayed green.** `agent-plugins-bge.12` delivered 11 snapshot
   cases. `agent-plugins-bge.13`, two items later, ran a negative
   control on a new portability test and it *passed when it should have
   failed*. Cause: `cli/cmd/snapshot_test.go` resolved each case
   directory as a relative path and then `os.Chdir`'d into a temp dir
   before reading `expected/checks.log` and `expected/manifest.json`, so
   every read hit the "`os.IsNotExist` → this case carries no such
   artifact" branch. The branch exists because [[testing-standards]]
   requires it: "WHERE an argument vector produces only one of the two
   artifacts, the case SHALL omit the other rather than record an empty
   file" (`specs/standards/testing-standards.md`, Snapshot Case
   Structure). Absence is therefore load-bearing intent, and a harness
   cannot distinguish "case carries no manifest" from "harness cannot
   find the manifest". Filed as `agent-plugins-0i6`.

2. **The shipped CLI answered over an empty corpus.**
   `agent-plugins-8cj`: `specs.toml` set `[workspace].root` to `docs/`,
   which the `specs/` reorg had emptied of specification documents. The
   CLI reported success over zero members. This is the same defect
   `rfcs/inbox/feature-blueprint/rfc-0010` documents for
   `scripts/validate-corpus.py` — but in a *second, independently built*
   tool, discovered separately.

3. **The acceptance-criteria coverage check this project mandates does
   not exist.** [[testing-standards]] states three normative rules:
   "Every acceptance criterion in an `approved` or `active` FRD SHALL be
   cited by at least one snapshot case manifest or at least one unit
   test"; "A map from rule identifier to acceptance-criteria identifiers
   SHALL be maintained"; and "the coverage check SHALL confirm that a
   citing case's expected artifacts contain a finding whose rule maps to
   that criterion." No such map and no such check exist anywhere in
   `scripts/` or `.github/workflows/` — verified by grep during this
   Definition-of-Done audit.

4. **A consequence of finding 3, measured.** Auditing the blueprint's
   51-row AC Coverage table by hand during this item found
   **AC-SPEC-008.6 entirely uncited**, after fifteen closed work orders,
   a green `go test ./...`, a green `check-skills`, and a green
   self-application run over 26 documents. It is substantively verified
   — `TestBuiltinRegistryRegistersSpecDrivenDev` and every snapshot
   case's `specs.toml` exercise it — but nothing declared it, so nothing
   could have noticed. Closed by adding the citation to
   `cli/testdata/cases/check-json/case.toml`.

5. **A second consequence, also from the hand audit.**
   **AC-SPEC-006.7** is declared `integration` in the AC Coverage table
   but is cited only from a unit test of `exitCodeFor`. The declared
   level is *unreachable*: every finding constructor in
   `cli/core/finding/finding.go` sets `SeverityError` except one setting
   `SeverityUnverified`, and none sets `SeverityWarning`, so no input
   corpus can produce the warning-only run the criterion describes.
   `agent-plugins-bge.12` recorded exactly this in flight ("no check
   registered anywhere in `cli/core/finding` or `cli/core/corpus` today
   raises severity warning or unverified") and worked around it. Filed
   as `agent-plugins-3md`.

6. **An acceptance criterion requiring a property its schema forbids.**
   `agent-plugins-bge.15`'s AC required "`version` bumped on each" of
   three blueprints. `ContainerBlueprint` in
   `cli/core/schema/spec-model.schema.json` declares no `version`
   property and sets `additionalProperties: false`, so bumping
   `container-spec-cli.md` would have *failed the corpus check*. Filed
   as `agent-plugins-e0t`; the criterion was met for the two module
   blueprints and recorded as unsatisfiable for the container.

7. **An acceptance criterion requiring a gate to see outside its own
   root.** `agent-plugins-bge.14`'s AC4 asked that `agent-plugins-6n4`
   be "confirmed as surfaced by the gate". 6n4's file
   (`plugin/skills/spec/references/work-order-template.md`) sits outside
   `specs/`, the corpus root the gate scans, so the gate cannot report
   it. Filed as `agent-plugins-mvx`.

8. **An acceptance criterion in tension with the decision pinned to
   satisfy it.** AC-SPEC-010.5 requires the CLI answer "without
   re-walking the corpus", while the content-hash staleness decision
   (`agent-plugins-igc`, pinned in the blueprint's CACHE delta) makes
   reading every candidate file unavoidable — staleness cannot be
   detected by content without reading content. What the warm path
   actually skips is parsing, decoding, classification and checking.
   Recorded by `agent-plugins-bge.10`.

9. **A ninth, procedural instance of the same class.**
   `agent-plugins-bge.8` found its item's "Verification timing" note
   ("AC-SPEC-009.3 is proved here at integration level with CACHE
   participating … Both land later") contradicting its own acceptance
   box, which required the 1,000-document budget be met in that item.
   The implementer read the acceptance box as authoritative.

10. **The generation pipeline forbids things the blueprints' Anchors
    sections named.** `agent-plugins-bge.7`'s Design specified embedding
    `cli/core/schema/spec-model.schema.json`, but Go's `//go:embed`
    cannot cross a package boundary (confirmed empirically by that
    implementer, "including that a symlink also fails with 'cannot embed
    irregular file'"). Resolved by extending
    `scripts/generate-schemas.sh` to emit a generated `embed.go`.

11. **Four more pipeline constraints, each rediscovered alone.**
    `agent-plugins-bge.1`: `scripts/generate-schemas.sh`'s `sync_dir()`
    runs `rm -rf` over `cli/core/specmodel/` every run, deleting any
    test file placed there — which makes [[testing-standards]]'s "Unit
    tests SHALL reside beside their subject" unsatisfiable for generated
    code (a sibling package `cli/core/specindex` was invented). Same
    item: Strata cannot emit a constant, so a hand-written Go file was
    required to have a version constant to test at all; and a backticked
    filename inside a `.strata` doc block turned `scripts/check-skills.sh`
    from its baseline 1 warning to **2 errors**, at a generated location
    under `plugin/skills/spec/references/generated/` that cannot be
    hand-edited. `agent-plugins-bge.4`: Strata flattens inheritance in
    every target, so the blueprint's "a base `Finding` struct" plus "a
    discriminated union of variants" shape would have duplicated five
    properties into every variant; nesting was chosen against the letter
    of the design.

12. **A schema regeneration invalidated snapshot fixtures that never
    mentioned the schema.** `agent-plugins-tff` added one optional
    `FeatureBlueprint` property; because `corpus.inputHasher` folds the
    whole `Profile.Schema()` in, the corpus `input_hash` changed and
    three `cli/cmd` snapshot fixtures (`index-default`,
    `index-out-custom`, `index-out-stdout`) failed although their corpus
    content never mentioned `plan` or `FeatureBlueprint`.

13. **The `specs/` reorg left `docs/`-era assumptions in six places, no
    two found by the same mechanism.** `scripts/validate-corpus.py`
    (rfc-0010 P1); `specs.toml` (finding 2); the blueprint's Definition
    of Done (corrected during `agent-plugins-bge.14`); the blueprint's
    Tracker Reconciliation (`agent-plugins-mvx`);
    `module-spec-profile`'s verification gate, "still named the `docs/`
    corpus after `agent-plugins-8cj` moved the root to `specs/`"
    (`agent-plugins-bge.15`); and the bodies of two tracker issues —
    `agent-plugins-dgn` cites `docs/frds/FRD-spec-validation-cli.md`,
    `agent-plugins-vg1` cites `docs/blueprints/*.md` (`agent-plugins-mvx`).

14. **Determinism had environmental preconditions nobody recorded.**
    `agent-plugins-bge.13` was dispatched believing cross-platform
    determinism was already implemented and only the assertion remained.
    It was not: `check --json` was identical across line endings but the
    index was not — "span offsets 11 to 12, length 158 to 163, and a
    different `input_hash` on a CRLF tree". CRLF normalization had to be
    added to `cli/core/corpus/corpus.go` `readUnder` before AC-SPEC-009.7
    could be asserted at all. Separately, putting `windows-latest` in the
    CI matrix required adding a `.gitattributes` pinning the tree to LF,
    because Git's `core.autocrlf` default there rewrites the committed
    byte-compared goldens on checkout.

15. **`go build`/`vet`/`test ./...` silently skip the snapshot
    regenerator.** `agent-plugins-bge.12` found that Go's tooling
    excludes any directory literally named `testdata` from `...`
    wildcard matching, so `cli/testdata/gen` — real, checked-in Go code
    — is invisible to every gate the work orders listed. It had to be
    verified by hand with `go vet ./testdata/gen`.

16. **Architecture blueprints carried claims about the implementation
    that nothing verified until one item read them end to end.**
    `agent-plugins-bge.15` found: `module-document-core` cited
    `goldmark` as an implementation choice "this module is built on"
    (absent from `go.mod`), deferred `samber/oops` "to implementation"
    (implementation shipped without it), and named fixture tests under
    `cli/core/testdata/` (does not exist; cases live at
    `cli/testdata/cases/`); its Interface guaranteed the module "SHALL
    NOT write to the filesystem outside a cache location the caller
    supplies" when no such caller-supplied location exists — the emitted
    index *is* the cache, at `corpus.IndexPath`.

17. **A blueprint's component inventory and its own C4 diagram
    disagreed.** `agent-plugins-bge.15`'s Design enumerated six
    components of `module-document-core` to confirm as built; the
    blueprint's diagram declares **eight**. Markdown Structure and
    References have no package in `cli/core` and no delta ever marked
    them unbuilt. The item's AC ("no blueprint statement contradicts the
    shipped code") pulled them into scope; its Design did not.

18. **An item could expose a contradiction it was forbidden to fix.**
    `agent-plugins-bge.3` recorded the extension point as supplying one
    union schema, which left `module-document-core` self-contradictory —
    its Schema Registry component still read "A profile registers one
    schema per type" — but that text belonged to `agent-plugins-bge.6`.
    Relatedly, `agent-plugins-bge.10` could not tell which half of
    `module-document-core` it owned versus `agent-plugins-bge.15`: its
    ACs required recording the cache there, while its Design said that
    fold "belongs to the blueprint-reconciliation item".

19. **Running the CLI to gather evidence dirtied the tree.**
    `agent-plugins-bge.15` ran `spec index` to measure rebuild cost and
    it wrote `specs/specs.json`, which was not gitignored, leaving an
    untracked artifact to remove by hand before committing. Filed and
    fixed as `agent-plugins-9tu`.

20. **Unpinned data locations became schema changes discovered
    mid-item.** Because every generated model sets
    `additionalProperties: false`, a datum the blueprint names in prose
    but does not place is a schema edit plus a regeneration plus a test
    update. `agent-plugins-bge.10`: "Neither the item nor the
    blueprint's CACHE delta says where the hash is persisted", so
    `input_hash` had to be added to `schemas/index.strata`, three
    artifacts regenerated, and a shape test updated. Same item: nothing
    said what *feeds* the hash beyond "the index's inputs", so the
    implementer decided it — and the decision that makes the
    fresh-CI-checkout condition hold (excluding absolute paths) "was not
    written down anywhere". `agent-plugins-bge.1`: the entry's
    `frontmatter` and `parse_error` both had to become optional against
    a Design that read as frontmatter-required, and the *position type*
    for a parse error had no stated source, so `Span` was declared in
    `schemas/index.strata` as an unguided decision — which then made
    `agent-plugins-bge.2`'s "Define the `Span` type" an
    alias-vs-duplicate choice neither the item nor
    `module-document-core` ("Owns the `Span` type") anticipated.

## Proposals

### P1 — Build the AC coverage check the standard already requires

**Problem.** [[testing-standards]] mandates three things that do not
exist: "Every acceptance criterion in an `approved` or `active` FRD
SHALL be cited by at least one snapshot case manifest or at least one
unit test"; "A map from rule identifier to acceptance-criteria
identifiers SHALL be maintained"; "the coverage check SHALL confirm that
a citing case's expected artifacts contain a finding whose rule maps to
that criterion" (`specs/standards/testing-standards.md`, Coverage). A
normative rule with no enforcement is indistinguishable from no rule: the
uncited AC-SPEC-008.6 of finding 4 survived fifteen items because nothing
looked.

**Change.** Add `scripts/ac-coverage.py`, and a step invoking it to
`.github/workflows/checks.yml`. It SHALL:

- parse acceptance-criteria identifiers from every FRD under `specs/`
  whose frontmatter `status` is `approved` or `active`;
- collect citations from two sources — `// AC: <ids>` comments in
  `cli/**/*_test.go` (comma-separated, all ids on the line) and the
  `cites = [...]` array in `cli/testdata/cases/*/case.toml`;
- exempt any criterion whose FRD text carries the `(manual)` marker;
- exit 1 on an uncited criterion or an orphaned citation (an id cited by
  a test but absent from every FRD);
- exit 2 when its own parse cannot be trusted — an FRD yielding zero
  criteria, or a tree yielding zero citations — because both otherwise
  read as better coverage than exists.

The exit-2 rule is the finding-2 lesson applied to the checker itself.
The sibling `strata` repository runs an equivalent `scripts/ac_coverage.py`
with exactly this exit contract; this proposal is not novel design.

The rule→AC map and the finding-content cross-check the standard also
mandates are deliberately **deferred**, not built here: no rule→AC map
exists to drive them, and authoring one is a separate decision. P3
amends the standard so this deferral is recorded rather than silently
unmet.

**Evidence.** Findings 3, 4, 5. The next run gets AC-SPEC-008.6's class
of defect at commit time instead of at Definition-of-Done, and the
level-mismatch of finding 5 becomes visible rather than needing a hand
audit of 51 rows.

### P2 — Make a harness that discovers nothing fail

**Problem.** Three checks reported success while verifying nothing
(findings 1, 2, 3), and in the first case the standard is the proximate
cause. [[testing-standards]] requires "WHERE an argument vector produces
only one of the two artifacts, the case SHALL omit the other rather than
record an empty file", which makes file *absence* the encoding of
intent. A harness reading that encoding cannot separate "this case
carries no manifest" from "this harness cannot find the manifest" — and
for a full wave, it did not.

**Change.** Two amendments to `specs/standards/testing-standards.md`,
Snapshot Case Structure, and the corresponding harness change.

Add, after the "SHALL omit the other" rule:

> A case manifest SHALL declare which expected artifacts the case
> carries.
>
> **Rationale:** omission encodes intent, so a harness that infers
> intent from a failed read cannot distinguish a case that carries no
> artifact from a case whose artifact it could not locate. An explicit
> declaration makes the second an error. Observed: an entire wave's
> snapshot suite asserted only exit status because every expected-artifact
> read resolved a relative path after the harness had changed directory.

Add, to Coverage:

> A test harness or corpus-scanning check SHALL fail when it discovers
> zero units of the thing it exists to check.
>
> **Rationale:** a scan that finds nothing and reports success is
> indistinguishable from a scan that passes, and it is the failure mode
> this project has now observed three times in independently built
> tools.

Harness side: `cli/cmd/snapshot_test.go` gains an `expects` field on
`caseManifest`; a declared artifact that cannot be read is a hard
failure. (`agent-plugins-0i6` already tracks the path-resolution half.)

**Evidence.** Findings 1, 2, 3. Backward compatible if `expects` defaults
to the currently-inferred set for cases that omit it — though this RFC
recommends populating all eleven cases in the same change, since there
are eleven.

### P3 — Require a criterion's declared verification level to be reachable

**Problem.** Four criteria in this feature named a target or a level that
no mechanism could satisfy (findings 5, 6, 7, 8), plus one item-level
contradiction (finding 9). Three became beads —
`agent-plugins-3md`, `agent-plugins-e0t`, `agent-plugins-mvx` — because
the implementers refused to fake them, which is the right outcome but a
late one. Nothing in [[testing-standards]] governs the *declared level*
at all: its Acceptance-Criteria Traceability section requires a citation
and exempts `(manual)`, and stops there. There is no text stating that a
level must correspond to a test that can exist.

`rfcs/inbox/feature-blueprint/rfc-0010` P3 and P5 address adjacent gaps
— missing `(manual)` markers, and a delta with no AC row. Neither covers
a criterion whose declared level is provably unreachable; that is this
proposal.

**Change.** Add to `specs/standards/testing-standards.md`, Coverage:

> An acceptance criterion's declared verification level SHALL name a
> level at which a test can exist against the mechanism the criterion
> describes.
>
> IF no test at the declared level can exist because the mechanism is
> unbuilt or unregistered, THEN the criterion SHALL either carry the
> `(manual)` marker or be re-declared at the highest reachable level,
> and the reason SHALL be recorded against the criterion.
>
> **Rationale:** a level nobody can reach is not a weaker guarantee, it
> is an absent one that reads as present. Observed four times in one
> feature: a criterion requiring a `version` bump on a document type
> whose schema forbids the property; one requiring a gate to surface a
> file outside its corpus root; one declared `integration` against a
> severity no registered check emits; and one whose text ("without
> re-walking the corpus") contradicts the decision pinned to satisfy it.

Also record, in the same section, that the rule→AC map and finding-content
cross-check the standard mandates are **not yet built** (P1), so the
standard stops overstating what this project enforces.

**Evidence.** Findings 5, 6, 7, 8, 9. Backward compatible: it constrains
future criteria and gives existing unsatisfiable ones a defined
resolution instead of a bead each.

### P4 — Record what the generation pipeline forbids, in `schemas/README.md`

**Problem.** Six work orders each independently hit a constraint of the
Strata → Go → JSON Schema pipeline that no document states, and two of
them hit a *blueprint anchor that named an impossible location*.
`schemas/README.md` has 111 lines under "The sources" and "Design
decisions" and documents none of it.

**Change.** Add a "Constraints this pipeline imposes" section to
`schemas/README.md`, stating at minimum:

- `//go:embed` cannot cross a package boundary, and a symlink does not
  work around it — a package needing generated bytes gets a generated
  `embed.go` emitted into the generated package (finding 10).
- `scripts/generate-schemas.sh`'s `sync_dir()` runs `rm -rf` over its
  output directories, so **no hand-written file may live in a generated
  package** — including tests. [[testing-standards]]'s "Unit tests SHALL
  reside beside their subject" therefore does not apply to generated
  code; the convention is a sibling package (`cli/core/specindex` for
  `cli/core/specmodel`). Amend the testing standard's Unit Tests section
  to carry this exception explicitly (finding 11).
- Strata cannot emit a constant; a version constant is hand-written Go
  and is part of the deliverable, not an afterthought (finding 11).
- Strata flattens inheritance in every target, so a "base struct plus
  discriminated union" design shape duplicates the base properties into
  every variant; model such shapes by nesting (finding 11).
- A regex literal in Strata cannot contain a forward slash, escaped or
  in a character class (`agent-plugins-8rr`, filed upstream).
- Regenerating any schema changes the corpus `input_hash`, because
  `corpus.inputHasher` folds the whole `Profile.Schema()` in — so a
  schema change invalidates every snapshot fixture carrying a manifest,
  including fixtures whose corpus content is unrelated (finding 12).
- Prose inside a `.strata` doc comment reaches
  `plugin/skills/spec/references/generated/` and is subject to
  `scripts/check-skills.sh`, at a location that cannot be hand-edited
  (finding 11).

**Evidence.** Findings 10, 11, 12. Each is a constraint that cost one
item real time and would cost the next item the same.

### P5 — Define the rename sweep: four surfaces, not just the files that moved

**Problem.** The `specs/` reorg predates this feature, and its residue
was still being found in the feature's *final* items — six places, no two
found by the same mechanism, and none by any of the others (finding 13).
One of them meant the shipped CLI answered over an empty corpus. No
document says what a directory rename obliges.

**Change.** Add a short standard —
`specs/standards/repository-change-standards.md`, or a section in an
existing standard if the reviewer prefers — stating:

> WHEN a directory or document path changes, the change SHALL sweep four
> surfaces before it is complete:
>
> 1. **Specification prose** — every `specs/**/*.md` mentioning the old
>    path, including Definition of Done and Tracker Reconciliation
>    sections, which are read late and therefore rot unobserved.
> 2. **Tooling configuration** — `specs.toml`, and any script with a
>    hard-coded root.
> 3. **Scripts** — `scripts/*.py`, `scripts/*.sh`.
> 4. **Tracker records** — the bodies of open beads citing the old path.
>
> **Rationale:** the `specs/` reorg was applied to the files that moved
> and to nothing that referenced them. Six residual sites were found
> across a later feature, each by a different accident.

Pair it with the mechanical half: extend `scripts/check-skills.sh` (or
add a step to `checks.yml`) that greps `specs/`, `scripts/` and
`specs.toml` for `docs/`-prefixed repository paths and fails on a path
that does not exist. This is a cheap grep, not the reference-checking
feature `agent-plugins-dgn` tracks.

**Evidence.** Finding 13, plus finding 2 as the cost. This RFC does not
restate rfc-0010's P1 (repoint `validate-corpus.py`); P5 is the general
rule that would have prevented all six sites, and it assumes P1 lands.

### P6 — Separate built from aspirational in the architecture blueprints

**Problem.** One item reading three blueprints end to end found four
implementation claims that were simply false, a guarantee describing a
mechanism that does not exist, and two components in a C4 diagram that
no delta ever addressed and no document marked unbuilt (findings 16,
17). A further item exposed a contradiction it was structurally
forbidden to fix, and another could not tell which half of a document it
owned (finding 18). These blueprints are `status: approved`; nothing
distinguishes a component that ships from one that is intended.

**Change.** Two amendments, both to the project's architecture
blueprints rather than to a template:

1. Every component in a module or container blueprint SHALL carry an
   explicit build state. Where a component is declared in the diagram
   but has no implementing package, the blueprint SHALL say so in the
   component's own text — not leave it to a reader to notice the count
   mismatch. Apply now to `module-document-core`'s Markdown Structure
   and References, and to `module-spec-profile`'s Element Records and
   Lifecycle Rules.
2. An implementation choice named in a blueprint (`goldmark`,
   `samber/oops`, a fixture path) SHALL be either present in the code or
   marked as a candidate. The four false claims of finding 16 were all
   of this shape.

**Evidence.** Findings 16, 17, 18. Backward compatible — it is additive
text on three approved documents, and `agent-plugins-bge.15` already
corrected the specific instances; this prevents the next set.

### P7 — Record determinism's environmental preconditions

**Problem.** Byte-identical cross-platform output (AC-SPEC-009.7) is
stated in the FRD and the blueprint as a property of the emitter. It is
not: it also requires input normalization and a checkout-line-ending
policy, neither of which any document mentions. An item dispatched to
"add the assertion" instead had to implement the normalization
(finding 14), and adding Windows to CI required a `.gitattributes` that
nothing anticipated.

**Change.** Add to `specs/architecture/module-document-core.md`, in the
Corpus component:

> Determinism has two preconditions outside the emitter. Input is
> normalized to LF at read time (`readUnder`), because a CRLF checkout
> otherwise shifts every span offset and changes the corpus
> `input_hash`; and the repository pins its working tree to LF via
> `.gitattributes`, because a byte-comparing snapshot suite is otherwise
> invalidated by Git's `core.autocrlf` default on Windows checkouts.

And note in [[testing-standards]], Snapshots, that a byte-comparing
suite constrains checkout line endings.

**Evidence.** Finding 14. The next agent asked to assert determinism
learns it is a three-part property before assuming two parts are done.

## Considered and Not Proposed

- **A rule→AC map and the finding-content cross-check.** Mandated by
  [[testing-standards]] and genuinely unbuilt, but building it needs a
  design decision (where the map lives, whether rules or criteria own
  it) that a retro should not make. P1 builds the reachable half and P3
  records the deferral, which is the honest treatment.
- **A mutation-testing gate for property tests.** `agent-plugins-bge.2`
  ran three mutations by hand to confirm its property tests had teeth
  and noted "no gate asks for that, so the next agent will likely skip
  it". Correct observation, but a standing mutation gate is a large cost
  for one item's worth of evidence; the honest recurrence claim is 1.
- **A "Go tooling gotchas" document** covering `testdata` exclusion from
  `./...` (finding 15). Folded into P4's pipeline-constraints section
  instead of a second document — one place to look beats two.
- **Correcting the slug prefix registry.** `agent-plugins-bge.7` found
  `template-authoring-standards.md`'s registry contradicting repo
  reality in two places with no row for the `guide` type. That file is a
  **vendored copy of a framework document** (`references/shared/`), so
  the fix belongs to the framework, not to this project's specs.
  Routed to `retro-skill`.
- **The dual RFC slug scheme** (`rfc-0007` vs `rfc-self-`), which makes
  the `numbered` flag ill-defined for the `rfc` type
  (`agent-plugins-bge.7`). Already tracked as an open concern in
  `rfc-self-delta-keys-and-tracker-vocabulary.md` under
  `rfcs/inbox/groom-feature-blueprint/`; re-proposing it here would
  relitigate a live question.
- **Strata's forward-slash regex limitation** (`agent-plugins-bge.4`).
  A defect in the Strata language, filed as `agent-plugins-8rr`; not a
  document of this project. Recorded in P4 only as a constraint to know.
- **A proposal about unpinned data locations** (finding 20). The
  evidence is strong — three items paid for it — but the concrete fix is
  a rule about how *feature blueprints* are authored, and the feature
  blueprint template is a framework document. Routed to `retro-skill`;
  P4 captures only the pipeline half (that `additionalProperties: false`
  makes an unpinned datum a schema change).
- **Reopening content hash vs mtime.** Settled as `agent-plugins-igc`
  and closed with the rationale recorded. Finding 8 is about the AC's
  wording, not the decision.
- **`agent-plugins-bge.5`'s missing findings.** An infrastructure
  failure (container restart), not a document defect. Recorded in
  Provenance so the absence is not read as a clean run.

## Non-Goals

- No new document types, and no change to the `Rfc` model —
  rfc-0010 P2 already proposes that.
- No weakening of any existing check. P2 makes the snapshot suite and
  the corpus scan **stricter**; the fix for a vacuous gate is never
  leniency.
- No change to the retired `specs/features/spec-index/` blueprint or its
  FRD beyond the corrections already committed. The proposals target
  durable documents.
- No re-decision of anything settled in this feature's Tracker
  Reconciliation.

## Adoption

**Independent and adoptable alone:** P4, P6, P7 — each is additive text
on documents this project owns, with no code dependency.

**Ordered pair:** P1 then P3. P1 builds the coverage check; P3 amends
the standard to describe what that check does and does not enforce.
Adopting P3 without P1 restores the current state — a normative rule
with no enforcement — which is the defect finding 3 records.

**Depends on rfc-0010 P1:** P5's sweep rule assumes
`scripts/validate-corpus.py` has been repointed at `specs/`; its grep
check would otherwise flag the validator as one of the six sites and
duplicate that RFC's proposal.

**Largest single change:** P2, which touches
`specs/standards/testing-standards.md`, `cli/cmd/snapshot_test.go` and
all eleven `case.toml` files. It is backward compatible if `expects`
defaults to the inferred set, but this RFC recommends populating the
eleven in the same change so no case relies on inference.

Backward compatibility holds for every proposal: no existing
specification document becomes non-conforming, and no generated artifact
changes shape. P1 and P5 add gates that will fail on current state —
that is their purpose, and each names the specific current failures
(AC-SPEC-006.7's level; `validate-corpus.py`'s root) so adopting them is
not open-ended.

Expected effect: the three vacuous-success defects of this phase become
detectable at commit time rather than by accident two work orders later,
and the four unsatisfiable criteria become a defined resolution at
authoring time rather than four beads filed during implementation.

## Resolution
