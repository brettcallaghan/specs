---
title: "RFC: Two standards describe one citation surface, and two
  measurements are taken where they cannot be trusted"
description: Give the traceability standard the manifest citation form
  its sibling already sanctions, and give the testing standard the two
  facts a measurement needs to be trustworthy — that a snapshot case
  runs once, and that a coverage figure is only valid on the tree
  carrying every change it measures
slug: rfc-0027-two-standards-one-citation-surface
type: rfc
status: ready-for-review
updated_at: 2026-09-07T00:00:00Z
origin: agent-plugins
skill: retro-spec
skill_version: "1.5.0"
model: "claude-opus-5"
effort: high
runs_observed: 1
severity: snag
plugin_version: "0.24.0"
---

# RFC: Two Standards Describe One Citation Surface

## Summary

Implementing [[frd-spec-list]] exercised three standards, and each held
a hole that let a claim read as verified when it was not.

[[ac-traceability-standards]] defines the citation obligation, and
defines exactly one citation form: a line comment adjacent to a test
declaration. [[testing-standards]] sanctions a second form — a citation
in a snapshot case's manifest — and this repository uses it in all 90
of its cases. The two never reference each other in that direction, so
an implementer reading the standard that *owns* the obligation cannot
learn that the manifest counts. The one who hit this wrote both forms,
which is how a single `// AC:` comment came to carry 34 identifiers and
bind to none of them.

The other two proposals are the same defect in a different register: a
measurement whose preconditions are unstated. A snapshot case replays
its argument vector once, so a criterion about two runs agreeing cannot
be verified by one — yet nothing says so, and such a criterion has been
declared `e2e` twice. And a coverage figure is only meaningful on a tree
carrying every change it measures, which under parallel dispatch is not
the tree the script runs in.

None of the three is a defect in the feature. All three are places where
a document let a reasonable agent believe it had proved something.

## Provenance

| | |
|---|---|
| Phase executed | Implementation of [[frd-spec-list]] — five work orders under epic `agent-plugins-kgir`, dispatched in parallel across three agents |
| Documents consumed | [[frd-spec-list]]; the feature blueprint `specs/features/spec-list/spec-list-blueprint.md`; [[container-spec-cli]]; [[module-document-core]]; [[ac-traceability-standards]]; [[testing-standards]]; [[frontmatter-standards]]; [[cli-output-standards]] |
| Documents produced | The `list` subcommand and its query, the two folded blueprints, `specs/guides/spec-cli-reference.md`'s `list` section |
| Findings source | 18 `retro:` comments across `agent-plugins-kgir.1`–`.5`, plus this item's own Definition-of-Done verification |

Versions at the time of the run:

| Document / artifact | Version | Status |
|---|---|---|
| [[ac-traceability-standards]] | `0.4.0` | active |
| [[testing-standards]] | `0.1` | approved |
| [[frontmatter-standards]] | `0.2` | active |
| [[module-document-core]] | `0.3` | approved |
| [[container-spec-cli]] | unversioned | approved |
| `specs/features/spec-list/spec-list-blueprint.md` | unversioned | in-progress → retired by this item |
| `retro-spec` skill | `1.5.0` | — |
| `sf` plugin | `0.24.0` | — |

## Evidence

### Finding 1 — the standard owning the obligation does not know
the manifest form exists

[[ac-traceability-standards]] `0.4.0`, "Citation Format", states:

> WHEN a test verifies an acceptance criterion, a citation comment SHALL
> be placed on the line(s) immediately preceding the test declaration.
>
> **Rationale:** Adjacent placement binds the citation to one specific
> test, so tooling can attribute coverage without language-aware
> parsing.

It gives two worked examples, both line comments, and never mentions a
manifest, a snapshot case, or a data-driven suite. Measured:
`grep -nic 'manifest\|case.toml\|data-driven'
specs/standards/ac-traceability-standards.md` → `0`.

[[testing-standards]] `0.1` treats the manifest citation as first-class,
twice:

> The manifest exists precisely so the citation lives outside the
> fixture.

> ...their citation in a manifest or a test comment stands on its own.

The cross-reference runs one way only. Measured:
`grep -c 'ac-traceability' specs/standards/testing-standards.md` → `3`;
`grep -c 'testing-standards'
specs/standards/ac-traceability-standards.md` → `0`.

The repository has settled the question in practice and neither document
records it: all 90 snapshot cases declare a `cites` key. Measured:
`ls cli/testdata/cases/ | wc -l` → `90`;
`grep -rl 'cites = ' cli/testdata/cases/*/case.toml | wc -l` → `90`,
spanning seven AC families (`AC-GHK`, `AC-SPEC`, `AC-SPECANCH`,
`AC-SPECLIST`, `AC-SPECMD`, `AC-SPECREF`, `AC-SPECSTR`).

The implementer of `agent-plugins-kgir.3`, reading both, recorded the
friction directly:

> a snapshot case is TOML with no test of its own, so a case-proved AC
> has TWO citation sites and neither document says which is
> authoritative: the case's `cites` key, and the shared `// AC:` comment
> on `TestSnapshotCases`. I wrote both. The shared comment is now 34
> identifiers on one line and grows with every feature, which is a
> citation nobody can read back to a specific test.

Measured on the shipped tree:
line 67 of `cli/cmd/snapshot_test.go`, piped through
`grep -o 'AC-[A-Z]*-[0-9]*\.[0-9]*' | wc -l`, yields `34`. That
comment is adjacent to `TestSnapshotCases`, which replays all 90
cases, so the standard's own stated rationale — "adjacent placement
binds the citation to one specific test" — is defeated by the form the
standard requires.

### Finding 2 — the prescribed verification command reports half
the coverage that exists

The Definition-of-Done work order (`agent-plugins-kgir.6`) prescribed
`grep -rn 'AC: AC-SPECLIST' cli/` as the way to confirm citations, which
follows [[ac-traceability-standards]] exactly. Measured, it returns 10
lines covering 11 criteria:

```
AC-SPECLIST-001.1, 001.4, 001.5, 001.6, 002.1, 002.3, 002.5,
003.1, 003.3, 003.4, 004.3
```

The feature has 23 criteria, of which 21 are non-`@manual`. Measured:
`(cd cli && go build -o ../build/spec ./spec)` then
`for f in 001 002 003 004 005; do build/spec criterion --of
"FEAT-SPECLIST-$f"; done` from the repository root → 23 criteria;
`build/spec criterion --of FEAT-SPECLIST-005 --json` → both `005.1` and
`005.2` carry `"manual":true`.

The remaining 10 are cited, in manifests:

```
cli/testdata/cases/list-all/case.toml:        001.2 001.3 003.2
                                              004.1 004.4 004.5
cli/testdata/cases/list-empty-corpus/case.toml:      001.7
cli/testdata/cases/list-type-repeated/case.toml:     002.2
cli/testdata/cases/list-type-comma/case.toml:        002.2
cli/testdata/cases/list-type-unregistered/case.toml: 002.4
cli/testdata/cases/list-json/case.toml:              004.1 004.2
cli/testdata/cases/list-all-statuses/case.toml:      003.2
```

The union is all 21. An agent that ran only the prescribed command would
have reported 10 uncited criteria and then either blocked the feature or
"fixed" it by adding redundant comment citations — growing the blanket
comment of finding 1 by another ten identifiers.

### Finding 3 — a criterion was declared at a level the harness cannot express

The feature blueprint's AC Coverage table declares AC-SPECLIST-004.3
("Members are reported in the order the index holds them", whose text
requires two runs over an unchanged corpus to agree) as `e2e`. The
snapshot harness cannot verify it. Measured: `cli/cmd/snapshot_test.go`
materializes each case into a fresh `t.TempDir()` (line 124) and calls
`cmd.Execute` exactly once (line 148), so a second run is not
expressible as a case.

The implementer proved it with a Go test in `cli/cmd/list_test.go`
instead, and recorded that this had happened before:
`cli/cmd/query_test.go` states the identical limit for AC-GHK-007.6 —
"the harness in `cli/cmd/snapshot_test.go` can only compare what one run
produces... only a Go test can". Two features, two implementers, the
same discovery made independently, and no document carries it.

[[testing-standards]]' "Snapshots" section states what a case must
contain and that a regeneration script must exist. It does not state
what a case cannot assert.

### Finding 4 — a coverage figure is measured on whichever tree
the script sits in

[[testing-standards]] requires the figures be committed and CI fail on a
mismatch:

> The reported figures SHALL be committed, and CI SHALL fail when they
> do not match a fresh measurement.

It does not say which tree the measurement is taken on. Measured:
`scripts/coverage.sh` resolves
`REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"` — the root
of whatever checkout the script itself lives in. Under parallel dispatch
that is a per-agent worktree that does not carry a sibling's merged
`cli/**` commits.

The implementer of `agent-plugins-kgir.5` caught it only by checking by
hand:

> the repository root you are standing in is a per-dispatch worktree
> that does not carry any sibling's merged work. Nothing in the item,
> the blueprint or the script warns of that, and the figures it writes
> are wrong without a word of warning if a sibling has already landed a
> `cli/**` change on the sprint branch.

It did not bite this run only by luck: the one commit absent from that
worktree, `8bff463`, touched four markdown files and no `cli/**` path.
Had it touched code, the committed figures would have been wrong and
CI's `--check` would have failed on a number no one could reproduce.

## Proposals

### P1 — Name the manifest as a conforming citation site, and
forbid the blanket comment

**Problem.** Finding 1 and finding 2. [[ac-traceability-standards]]
`0.4.0` defines one citation form; [[testing-standards]] `0.1` sanctions
a second; the repository uses the second in 90 of 90 cases; and the
document that owns the obligation is silent on it. The cost is not
theoretical — it produced a 34-identifier comment that satisfies the
letter of the standard while defeating the rationale printed directly
beneath it, and it makes the standard's own greppable-citation promise
report half the coverage that exists.

**Change.** Add to [[ac-traceability-standards]], "Citation Format",
after the existing marker rule:

> WHERE a test is data-driven — one harness replaying a set of cases
> declared as data — the citation SHALL be recorded against the
> individual case rather than against the harness, in whatever
> per-case manifest the suite defines.
>
> **Rationale:** the obligation is that a citation bind to one specific
> verification. For a data-driven suite the case is that unit and the
> harness is not, so a comment on the harness binds to every case and
> therefore to none. This repository's snapshot cases record it as a
> `cites` array in `cli/testdata/cases/<case>/case.toml`;
> [[testing-standards]] governs that manifest's shape.
>
> A citation comment SHALL NOT enumerate identifiers verified by cases
> the comment's own test declaration does not individually assert.

And add [[testing-standards]] to that document's References section,
noting that it defines the manifest form for this repository's snapshot
cases.

Also amend the "Coverage Obligations" section so the greppability claim
stays true, replacing the single regex with a statement that a coverage
query SHALL read both sites — the comment marker and the per-case
manifest key.

**Backward compatibility.** Every existing comment citation stays
conforming. The one artifact made non-conforming is the blanket comment
at `cli/cmd/snapshot_test.go:67`, whose 34 identifiers are already
carried by the individual cases; removing it is a deletion, and finding
2 shows the manifests already hold the full set. No case file changes.

**Evidence.** Findings 1, 2.

### P2 — State what a snapshot case cannot assert

**Problem.** Finding 3. A snapshot case replays one argument vector
once, so no case can assert a property *between* two runs —
determinism, idempotence, warm-versus-cold agreement. Nothing in
[[testing-standards]] says so, and the omission has now cost two
features: AC-GHK-007.6 and AC-SPECLIST-004.3 were both declared at a
level the harness cannot reach, and both were rescued by an implementer
who noticed at coding time and wrote a Go test instead.

This is adjacent to `rfcs/inbox/standards/rfc-0011` P3, which requires a
declared level to be *reachable*. That proposal addresses a level
unreachable because the mechanism is unbuilt or unregistered. This one
is different: the mechanism is built and working, and the harness at the
declared level cannot express the assertion. P3 tells an author to
re-declare the level; it does not tell them which levels are impossible
here, which is the fact that was missing both times.

**Change.** Add to [[testing-standards]], "Snapshots":

> A snapshot case SHALL replay one argument vector once. A criterion
> asserting a relationship between two runs — that repeated invocation
> agrees, that an operation is idempotent, that a warm run matches a
> cold one — SHALL NOT be declared at the snapshot layer, and SHALL be
> verified by a test that can invoke the subject more than once.
>
> **Rationale:** the harness materializes each case into a fresh
> temporary directory and invokes the subject a single time, so the
> second run a cross-run property needs does not exist. A criterion
> declared there reads as verified by the case that cites it while
> nothing has compared anything.

**Backward compatibility.** Constrains future declarations. The two
existing criteria in this class are already verified by Go tests, so no
current case or criterion changes; the rule records what those two
implementers each worked out alone.

**Evidence.** Finding 3.

### P3 — State the tree a coverage figure is measured on

**Problem.** Finding 4. [[testing-standards]] requires a committed
figure that CI re-measures, and says nothing about the tree the
measurement is taken on. `scripts/coverage.sh` measures whatever
checkout it lives in, which under parallel dispatch is a worktree
missing every sibling's merged work. The failure is silent at the point
of measurement and surfaces in CI as a mismatch with no visible cause.

**Change.** Add to [[testing-standards]], "Coverage", beside the
committed-figures rule:

> A coverage figure SHALL be measured on a tree that carries every
> change the figure reports on, and the measurement SHALL fail rather
> than report when it cannot establish that.
>
> **Rationale:** the figures are a property of the integrated change,
> not of one contributor's checkout. A measurement taken on a tree
> missing a sibling's commits is wrong in the direction that reads as
> correct — it produces a plausible number that CI cannot reproduce,
> and the diff that caused it is not the diff that carries it.

**Backward compatibility.** No existing figure changes; the committed
figures were verified correct for this run
(`scripts/coverage.sh --check` exits 0). The rule constrains when a
regeneration is trustworthy. Whether `scripts/coverage.sh` grows an
assertion enforcing it is an implementation question for the bead that
adopts this, not part of the standard's text.

**Evidence.** Finding 4.

## Considered and Not Proposed

| Finding | Why not proposed |
|---|---|
| `bd dolt pull` takes the shared Dolt write lock, never returns, ignores `SIGTERM`, and starves every concurrent `bd` call. Recorded on `kgir.1` and `kgir.4` with process-level evidence; cost this run roughly 20 minutes twice. | Real and expensive, but it is a fact about the tracker and the dispatch harness, not about a document this phase consumed. Its home is the environmental-failure ledger in `AGENTS.md` and the `swarm` skill lens; it belongs to the run's own retro, not to a spec RFC. |
| Vendored `references/shared/` copies cite sibling shared files with a `shared/` prefix that resolves in neither home, and `scripts/check-skills.sh` skips vendored copies entirely. | Already filed as `agent-plugins-ru6x` with the measurement attached. A bead is the actionable artifact; an RFC would restate it. |
| The DOC work order documented `spec list` on a branch where the command did not exist, so `specs/guides/spec-cli-reference.md` shipped sourced from the blueprint rather than from a binary, and nothing re-checked it afterwards. | The durable fix is a Definition-of-Done shape — a documentation delta written ahead of its surface must be re-verified against the built surface before the feature closes — which lives in the feature-blueprint template this framework vends, not in this project's documents. Routed to `retro-skill`. (Verified by hand in this item: `build/spec list`, `--type adr`, and `--type bogus` all match the guide's text.) |
| The FILT work order changed the CORE work order's published query signature rather than adding beside it. | The plan sequenced CORE → FILT → CMD under `single-pr` delivery precisely so this is invisible outside the branch. Working as designed, and the implementer recorded it as such. |
| The blueprint pinned the summary's fields and their order but never named the query identifier or its Go shape, so CORE chose both. | Implementation-level detail correctly left to the implementer; [[container-spec-cli]] now names `corpus.Corpus.DocumentSummaries` alongside every other query, so the gap is closed where it mattered. |
| The FILT delta's Anchors named `query.go` and `profile.go` but not the pre-existing `TestTheExtensionPointNamesNoRegisteredVocabulary` guard that already enforced the delta's own constraint. | Genuine friction, but `agent-plugins-kgir.2` fixed it at the source: [[module-document-core]] now names that guard beside the extension point, so the next reader finds it. A standing rule about what Anchors must enumerate is a template concern, routed to `retro-skill`. |
| The Go representation of an absent frontmatter envelope field (pointers with `omitempty` versus bare strings) was undecided by the blueprint. | A blueprint that pinned Go types would be specifying the implementation. The implementer decided it with a negative control and recorded the rejected alternative; that is the process working. |

## Non-Goals

- **Building the coverage checker.** `rfcs/inbox/standards/rfc-0011` P1
  already proposes `scripts/ac-coverage.py` reading both citation
  sites. P1 here supplies the standards text that checker would
  enforce; it does not duplicate the tooling proposal, and the two are
  independently adoptable in either order.
- **Changing any snapshot case, criterion, or coverage figure.** All
  three proposals are document changes. The single code artifact
  touched by adopting P1 is the deletion of one comment line.
- **Revisiting the manifest's shape.** Whether `cites` is the right key
  name, and what else a manifest should carry, is
  [[testing-standards]]' existing business.
- **Re-litigating verification-level reachability in general.** That is
  `rfc-0011` P3; P2 here adds one concrete impossibility to the
  catalogue rather than restating the principle.

## Adoption

Each proposal is independently adoptable; none depends on another.

- **P1** touches [[ac-traceability-standards]] plus one comment-line
  deletion in `cli/cmd/snapshot_test.go`. Both vendored copies of the
  standard under `plugin/skills/*/references/shared/` must be refreshed
  by `scripts/sync-shared-references.sh` in the same change, which
  `scripts/ci.sh` verifies.
- **P2** and **P3** touch [[testing-standards]] only, in two different
  sections, and can land in one edit or two. That document is not
  currently vendored into any skill, so no sync step applies.
- Adopting **P1** before `rfc-0011` P1 is the cheaper order: the
  checker then has a standard to implement rather than a practice to
  infer. The reverse order also works and loses nothing.

## Resolution

<!-- Filled by the resolving review: adopt / defer / reject per
     proposal, with rationale and any resulting issue ids. -->
