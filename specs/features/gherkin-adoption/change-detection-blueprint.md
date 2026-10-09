---
title: "Criterion Change Detection Against Git"
description: Replaces the persisted-index baseline of criterion --changed
  with the corpus at a git revision, adding a Revisions component to
  Document Core and --since, --exit-code and a structured answer to the
  command
slug: feature-gherkin-change-detection
type: feature-blueprint
status: retired
updated_at: 2026-10-08T12:00:00Z
delivery: single-pr
frd: frd-gherkin-adoption
---

# Criterion Change Detection Against Git

## Overview

**FRD:** [[frd-gherkin-adoption]]

This blueprint covers FEAT-GHK-008 only, as rewritten on 2026-10-08. The
rest of the FRD was built under `feature-gherkin-adoption`, now
retired, and its other changes of that date are carried as standalone
work orders (see Tracker Reconciliation).

Today `criterion --changed` compares the working tree against
`specs.json`, an artifact every query may rewrite, so its baseline is
whatever the last query left behind. The comparison moves to a git
revision. A new **Revisions** component in [[module-document-core]]
resolves the base commit and copies the spec root as that commit holds
it into a temporary directory, through the `git` executable
([[adr-031]]). The Corpus build already used for the working tree builds
that copy unchanged, and Command Surface compares the two builds' criteria
by identifier and content hash, as it does now. The persisted index takes
no part.

Delta blocks in dependency order: BASE, then DIFF.

## Element Deltas

### Revisions in [[module-document-core]] (added) — BASE

**Intent:** give Document Core the one capability the comparison lacks,
which is reading the spec root as a commit holds it. The capability is
framework-agnostic, since it knows git and directories and no document
type, so it belongs in the module rather than in the command surface
([[container-spec-cli]]'s Change Guidance). It returns a directory
rather than a parsed corpus so that the base and the working tree pass
through one build.

**New components:**
- **Revisions** (Go package) — Resolves a git revision to a commit and
  materializes a directory's tracked files at that commit. Runs the
  `git` executable as a subprocess: `rev-parse` to locate the repository
  and resolve revisions, `merge-base` for the base, `ls-tree` to list a
  path's tracked files, and `cat-file --batch` to read their blobs.
  Writes only into a temporary directory it creates and removes. It
  knows nothing of documents, criteria or the index.

  ```component
  entity_name: Revisions
  slug: component-revisions
  description: Resolves a git base revision and materializes a directory as a commit holds it
  ```

**Behavior:**
- WHEN asked for the base of a comparison that names no revision, the
  component SHALL resolve the default branch as the first of these that
  exists: the target of `refs/remotes/origin/HEAD`, then the branches
  `main`, `origin/main`, `master` and `origin/master`.
- IF none of those exists, THEN the component SHALL report an
  environment error naming each reference it tried and stating that
  `--since` names a base explicitly.
- WHEN asked for the base of a comparison against a named revision, the
  component SHALL accept any revision `git rev-parse --verify
  '<revision>^{commit}'` resolves: a branch, a tag, a commit or an
  expression such as `HEAD~3`.
- IF a named revision does not resolve, THEN the component SHALL report
  an invocation error naming the revision.
- The base SHALL be `git merge-base HEAD <revision>`.
- IF `git merge-base` finds no common ancestor, THEN the component SHALL
  report an environment error. WHERE `git rev-parse
  --is-shallow-repository` reports `true`, the error SHALL say that the
  clone is shallow and that deepening the fetch or naming `--since`
  resolves it.
- IF the working directory is not inside a git work tree, or the `git`
  executable cannot be run, THEN the component SHALL report an
  environment error naming which.
- WHEN materializing a directory at a commit, the component SHALL write
  every regular file (`100644` and `100755` entries) that `ls-tree -r`
  lists beneath it, byte for byte from its blob, at its
  repository-relative path under a fresh temporary directory, and SHALL
  write nothing else.
- The component SHALL NOT materialize symbolic links (`120000`) or
  submodules (`160000`), and SHALL NOT apply checkout filters,
  `.gitattributes` or `export-ignore`.
- WHERE the directory does not exist at the commit, the component SHALL
  materialize an empty directory rather than report an error, because a
  spec root added since the base holds only added criteria.
- The component SHALL remove its temporary directory before the command
  returns, on success and on every error path.
- The component SHALL NOT write to the repository or its `.git`
  directory, and SHALL NOT change HEAD, the index or the working tree.

**Anchors:** new package beside `cli/core/corpus`. Its tests build
throwaway repositories with `git init` in `t.TempDir()`, the way
`cli/specs/main_e2e_test.go` already drives subprocesses.

### Command Surface in [[container-spec-cli]] (modified — change listing) — DIFF

**Intent:** point `criterion --changed` at a git base and give it the
answer and the statuses [[cli-output-standards]] and
[[cli-exit-status-standards]] require of a query. The comparison itself
does not change: identifier and content hash only, never position.

**Behavior:**
- WHEN `criterion --changed` runs, the command SHALL build the current
  corpus from the working tree as `corpus.Build` does today, so that
  committed, staged, unstaged and untracked documents beneath the spec
  root are all part of the current side.
- WHEN `criterion --changed` runs, the command SHALL build the base
  corpus with `corpus.Build` over the directory BASE materialized, using
  the current configuration and profile, with `RepoRoot` and `Root`
  rebased onto that directory at the same repository-relative paths, so
  that both sides report identical paths for an unmoved document.
- The command SHALL NOT read, write or require `specs.json` on either
  side of the comparison.
- IF the configured spec root lies outside the git work tree, THEN the
  command SHALL report an environment error naming both.
- The command SHALL accept `--since <revision>` and pass it to BASE, and
  SHALL treat `--since` or `--exit-code` given without `--changed` as an
  invocation error.
- The command SHALL answer with a mapping holding `base`, then
  `criteria`. `base` holds `revision` (the revision named or the default
  branch resolved, as a reference name) and `commit` (the merge-base, as
  a full object name). `criteria` is a sequence holding one mapping per
  listed criterion with `id`, `classification` (`changed`, `added` or
  `removed`) and `path`, sorted by identifier.
- A criterion's `path` SHALL be its document's path in the current
  corpus, and for a removed criterion its path in the base.
- WHEN no criterion differs, the command SHALL answer with an empty
  `criteria` sequence and exit `0`.
- The command SHALL render the answer as YAML, and as JSON under
  `--json`. `--json` SHALL no longer be refused alongside `--changed`.
- WHEN `--exit-code` is given and `criteria` is non-empty, the command
  SHALL exit `1`. Otherwise an answered comparison SHALL exit `0`.
- The command SHALL report BASE's errors under the exit status
  [[cli-exit-status-standards]] assigns their class, and SHALL write
  nothing to standard output when it does.
- The "no persisted index" message and the `corpus.Load` read of the
  baseline SHALL be removed.

**Anchors:** `runCriterionChanged`, `changedCriteria`, `criterionHashes`
and `renderCriterionChanges` in `cli/cmd/criterion.go`; `corpusOptions`
in `cli/cmd/root.go`. The snapshot cases `criterion-changed`,
`criterion-changed-none` and `criterion-changed-no-index` under
`cli/testdata/cases/` depend on a persisted index as their baseline, and
a snapshot case's input cannot hold a git repository. They are replaced
by end-to-end tests that build their repository at test time.

### Relied-upon elements (unchanged)

- [[module-document-core#Corpus]] — `corpus.Build`, which builds both
  sides, and the CRLF normalization on its read path that makes a blob's
  bytes hash the same as the checked-out file's.
- [[module-document-core#Gherkin]] — the criterion content hash that
  ignores position, which AC-GHK-008.2 rests on.
- [[module-spec-profile]] — the feature tag registration that decides
  which blocks hold criteria, the same on both sides.

## Contracts

None. The answer's shape is fixed in the DIFF delta and reaches callers
only through the command surface.

## Tracker Reconciliation

- `specs-xqm.3` — subsumed by DIFF, with BASE: "criterion --changed
  compares against a git ref instead of the persisted index." Grooming
  blocks it on DIFF, and the Definition of Done closes it.
- `specs-xqm.20` — prerequisite: it implements the exit-status classes,
  including the environment status `3` that BASE's errors and the
  `--since` invocation error depend on.
- `specs-xqm.9` — out-of-scope: it renders `feature` and `criterion`
  queries as YAML lists. DIFF renders the change listing itself, and
  shares that work's YAML renderer without depending on its record
  shape.
- `specs-xqm.21` — out-of-scope: whether queries may write `specs.json`.
  DIFF stops `--changed` reading it, whatever that decides.

## Sequencing Constraints

- `delivery: single-pr`. BASE and DIFF land in the same change, with
  the documentation updates the Definition of Done lists.
- The BASE delta SHALL land before the DIFF delta.
- The `specs-xqm.20` work order SHALL land before the DIFF delta.

## AC Coverage

Legend: BASE = Revisions delta in [[module-document-core]];
DIFF = change listing delta in [[container-spec-cli]].

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-GHK-008.1 | BASE, **DIFF (proves)** | e2e |
| AC-GHK-008.2 | **DIFF (proves)**, BASE | e2e |
| AC-GHK-008.3 | **DIFF (proves)**, BASE | e2e |
| AC-GHK-008.4 | **BASE (proves)**, DIFF | unit |
| AC-GHK-008.5 | **BASE (proves)**, DIFF | unit |
| AC-GHK-008.6 | **DIFF (proves)** | e2e |
| AC-GHK-008.7 | **DIFF (proves)** | e2e |
| AC-GHK-008.8 | **DIFF (proves)** | e2e |
| AC-GHK-008.9 | BASE, **DIFF (proves)** | e2e |

## Key Decisions

- [[adr-031]] — base revisions are read through the `git` executable,
  with tracked blobs copied into a temporary directory.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — Revisions added to the Components
    roster, the diagram and its record block
  - [x] [[container-spec-cli]] — the `criterion --changed` paragraph of
    Command Surface rewritten for the git base, `--since` and
    `--exit-code`
- [x] [[adr-031]] status: draft → approved, and referenced from the Key
  Decisions of [[module-document-core]] and [[container-spec-cli]]
- [x] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]]
- [x] Every behaviour this blueprint decided that no FRD criterion asks
  for (the default-branch order, the answer's shape, `--since` without
  `--changed`, a spec root outside the work tree) is lifted into
  [[frd-gherkin-adoption]] as a criterion and cited from its test
- [x] `specs-xqm.3` is closed
- [x] No live document references this blueprint (`specs refs
  feature-gherkin-change-detection` names no live document), per
  [[document-relationship-standards]]
- [x] This blueprint's status → retired
- [x] With this blueprint already retired, `specs check --strict` over
  the corpus is green

## Open Questions

None.

## References

- [[c4-standards]] — the component and module definitions the delta
  blocks are written against
- `feature-gherkin-adoption` — the retired blueprint that built the
  rest of this FRD, including the index-based comparison this replaces
- [[adr-020]] — FRDs evolve in place, and deltas between versions are
  keyed on AC identifiers, which is the comparison this command performs
- [[cli-exit-status-standards]] — the classes BASE's errors map to
- [[cli-output-standards]] — the shape of DIFF's answer
