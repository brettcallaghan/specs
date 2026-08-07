# agent-plugins-9kb.4 — Resolve references into the index graph with both directions derived

## Description

**Blueprint:** specs/features/reference-integrity/reference-integrity-blueprint.md — delta **RES**, section "References in [[module-document-core]] (added — resolved graph) — RES".
**FRD:** specs/features/reference-integrity/reference-integrity-frd.md.
**Delivery:** staged (stage 1 of 2); this feature's closure set is `bd list --label feature-spec-reference-integrity`.

Turn extracted references into a resolved graph and put it in the index.
Resolution is a lookup against slugs the corpus already holds, so the cost of
this work is the walk it rides on rather than a second pass; the graph it
produces is what makes a single-document check able to see the documents
pointing **at** it, and what the `refs` query reads.

This is the last item of stage 1. When it lands, `spec index` emits reference
edges over this repository's own corpus — no new check runs, no finding is
raised, and no exit status changes. That boundary is deliberate: it puts the
extraction and resolution rules, the ones that produce false positives, in
front of a reviewer before any of them can fail a build.

**Verification timing.** This item proves AC-SPECREF-001.3 (unit), -001.6
(integration), -002.2 (unit), -002.7 (unit) and -002.8 (unit), and owes those
citations. It also participates in AC-SPECREF-002.1 (proved by CHK) and
AC-SPECREF-003.1, -003.2 and -003.4 (proved by CMD), all outside this item's
dependency closure. Code work may proceed as soon as this item is ready;
verification of those four waits for CHK and CMD, and the final
Definition-of-Done item is the backstop.

AC-SPECREF-001.6 is verified at **integration** level against the existing
`index` subcommand, which already exists — it does not wait on CMD.

## Retro

Before closing this item, record retro findings on this issue —
one `bd comment` per finding, prefixed `retro:` — following the
findings capture protocol in `rfcs/findings-capture.md`. Friction
only: what happened plus the item or spec text that caused it (or
the gap where guidance was missing). No analysis, no RFCs — the
plan's final DoD item runs the aggregated retro.


## Design

Anchors: `cli/core/references`; `cli/core/corpus/corpus.go` (`Build`, the
`builder`, and the `Corpus` lookups); the emitted index `corpus.Write` produces
and `corpus.Open` serves — its `input_hash` already folds the active profile's
registration, so a profile that registers a new reference property invalidates
the warm index by construction.

- Resolve a reference's slug half against the slugs corpus members declare,
  compared **lowercased**, matching the case-insensitive slug uniqueness
  AC-SPEC-003.4 establishes (AC-SPECREF-002.8).
- WHEN a reference's slug matches a file the index recorded as an **exclusion**,
  or an entry recording a **parse failure**, record that outcome rather than
  "matches nothing", carrying the reason and the file's path
  (AC-SPECREF-002.7). The reader's next step differs, which is why the outcome
  differs.
- Resolve **only within the workspace** the run is against. A slug declared
  only in another repository's index is unresolved, per the FRD's stated
  boundary. The merged-graph case is a known follow-on, not a defect.
- Derive **inbound** references from the same collection the outbound ones are
  read from ([[adr-018]], AC-SPECREF-001.3), and do **not** store the reverse of
  an edge a second time.
- WHEN a reference resolves to no corpus member, record it in the index with its
  unresolved state, so `index` exits zero over a corpus whose links are broken
  (AC-SPECREF-001.6). Indexing never fails; reporting is the checker's job.
- Corpus extracts and resolves **during the walk that already reads each
  candidate's bytes**, so that no document is read twice.
- WHERE a corpus is served **warm**, Corpus reads the graph from the persisted
  index rather than parsing any document — which is what keeps the `refs` query
  and the inbound half of a single-document check inside the AC-SPEC-009.1 and
  AC-SPEC-009.2 budgets (AC-SPECREF-003.4).
- Corpus exposes lookups for **both directions** alongside the existing `Entry`
  and `Slug`, built once when the index is opened.
- Row order follows [[adr-018]]: sorted by path, then by position within the
  document. That satisfies AC-SPECREF-001.5's ordering and AC-SPEC-009.7's
  byte-identical output in one rule.


## Acceptance Criteria

- [ ] A reference's slug half resolves against corpus slugs compared lowercased (**proves AC-SPECREF-002.8**, unit).
- [ ] A reference carrying an anchor resolves on its slug portion alone, and no finding is raised about the anchor (**proves AC-SPECREF-002.2**, unit).
- [ ] A slug matching an excluded file or a parse-failure entry records that reason and the file's path, distinctly from "matches nothing" (**proves AC-SPECREF-002.7**, unit).
- [ ] Inbound references are derived for every document from the same collection, with no edge stored twice (**proves AC-SPECREF-001.3**, unit).
- [ ] A corpus containing unresolved references produces a complete index and `index` exits zero (**proves AC-SPECREF-001.6**, integration).
- [ ] A slug declared only outside the workspace resolves as unresolved.
- [ ] Extraction and resolution ride the existing corpus walk; no document is read twice.
- [ ] A warm corpus serves the graph from the persisted index without parsing any document.
- [ ] `Corpus` exposes inbound and outbound lookups built once at open, alongside `Entry` and `Slug`.
- [ ] The `references` rows are ordered by path then position, and the emitted index is byte-identical across repeated runs.
- [ ] All citations use the `AC:` marker adjacent to the test declaration per [[testing-standards]].
- [ ] [[module-document-core]] is folded: the References component records the resolution and graph half with its pinned semantics. Set `version` to 0.3 if not already set.

## Notes

tier: advanced — Reaches into the built Corpus build and warm-open paths: resolution, derived inbound edges with no edge stored twice, deterministic row ordering, and read-from-persisted-index on the warm path. Five ACs, and a determinism or warm-cache mistake is a silent wrong answer rather than a failure.

## Retro

### Approach

Resolution went into `cli/core/references/resolve.go` as `Resolve(refs, index)`,
a pure function of the reference rows and one `specmodel.SpecIndex`. The
References component owns the *rule* (the module blueprint says it "resolves
them against the corpus"); Corpus owns the *wiring*. The package still does not
import `core/specindex`, so the `agent-plugins-gz3` boundary test holds
unchanged — it forbids that package, and `specmodel.SpecIndex` is the shared
generated model, not that package.

Wiring in `cli/core/corpus`:

- `builder.classify` now threads the walk's `src` into `classifyDocument`, which
  calls `references.Extract` for every full member. Extraction therefore rides
  the read the walk already performed; nothing opens a file twice.
- `builder.finish` sorts the collected rows by path then `Position.Offset`,
  appends them to `index.References`, and calls `references.Resolve` once. It
  runs after the walk because resolution is a question about the *whole* corpus:
  the first file read may name the last.
- `Corpus` gained `outbound` and `inbound` — `map[string][]int` into the one
  `index.References` collection — built by `indexReferences()`, called from both
  `finish` (cold) and `fromIndex` (warm). No reverse edge is stored.
- `Corpus.OutboundReferences(path)` / `InboundReferences(path)` are the two
  public lookups, keyed by path and accepting the same path forms `Entry`,
  `Document` and `FindingsFor` accept.
- `inputHasher.sum` now folds each type's `ReferenceProperties`.

### Implicit decisions

1. **How a slug matches a non-member (AC-SPECREF-002.7): the file name without
   its extension, lowercased.** Neither the blueprint nor the FRD says how the
   match is made, and it has to be decided to satisfy the criterion at all. The
   index records nothing else about a non-member: an `IndexExclusion` carries
   path/reason/type and an entry with a `parse_error` carries no frontmatter —
   a parse failure by definition has no readable slug. The file *name* is the
   only handle both cases still have, and it is on the index, so the match is
   reproducible by anything reading the artifact rather than resting on
   knowledge only the build held. Recorded in `stem`'s doc comment and in the
   module blueprint: the consequence is one-directional (a non-member whose
   intended slug differs from its file name reports as an ordinary unresolved
   reference — the answer it got before the rule existed), and it lands the
   motivating case, the FRD's `work-order-template`.
2. **A member always wins over a file named like one**, and ties among
   non-members go to the first in path order, decided by comparison rather than
   by arrival order, so `Resolve` is order-independent with respect to how the
   index holds its collections. Tested directly
   (`TestResolutionDoesNotDependOnIndexOrder`).
3. **Only `ReferenceProperties` is folded into the input hash, not the whole
   registration**, even though the blueprint's RES anchors assert the hash
   "already folds the active profile's registration" — it did not; it folded
   only `SlugPattern` and `Numbered`. Reference properties decide the index's
   *content*, so omitting them would serve a stale index after a profile change.
   Typed references, pairings and the lifecycle reach nothing the index records
   and are read by checks against the live profile every invocation, so folding
   them would rebuild the corpus to produce identical bytes. Reasoned in the
   code comment and in the module blueprint. No `input-hash-version` bump: the
   new named field changes the digest for every corpus by itself, so every
   pre-existing index is invalidated automatically.
4. **Lookups keyed by path, not by slug.** `refs <slug>` (CMD) converts a slug
   to a path with the existing `Slug` lookup; keying by path keeps the two
   directions uniform and serves AC-SPECREF-002.9's single-document case
   directly.
5. **Names**: `OutboundReferences` / `InboundReferences` over `Outbound` /
   `Inbound` — the corpus has several lookups and a bare direction word says
   nothing about what is being looked up. Internal helpers: `targets`,
   `lookupKey`, `stem`, `orderedReferences`, `indexReferences`,
   `referencesAt`, `extractReferences`.
6. **File organisation**: resolution in its own `resolve.go` beside
   `references.go`; corpus-level graph tests in a new
   `cli/core/corpus/references_test.go` rather than appended to the 753-line
   `corpus_test.go`; one internal test file `corpus_internal_test.go` for the
   walk property.

### Verification

- AC-SPECREF-002.8 — `TestASlugResolvesComparedLowercased` (unit).
- AC-SPECREF-002.2 — `TestAnAnchoredReferenceResolvesOnItsSlugAlone` (unit),
  including the negative control that the unsplit text matches nothing.
- AC-SPECREF-002.7 — `TestASlugMatchingANonMemberRecordsThatReasonAndPath`
  (unit) and `TestAReferenceToANonMemberRecordsTheReasonAndPath` (corpus).
- AC-SPECREF-001.3 — `TestInboundReferencesAreDerivedFromTheSameCollection`
  (unit), which also asserts the "no edge stored twice" half by counting rows
  and comparing the outbound and inbound views of one edge.
- AC-SPECREF-001.6 — new snapshot case `index-unresolved-reference`
  (integration): `index` over a corpus with one broken link exits 0 and emits a
  complete index carrying the unresolved row. `cites` in `case.toml` is the
  citation form the testing standard requires of a data-driven case; the AC was
  also added to `TestSnapshotCases`'s own `AC:` line.
- AC-SPECREF-001.5 — `TestTheIndexCarriesEveryReferenceOrderedByPathThenPosition`,
  which also asserts two builds render byte-identically.
- Out-of-workspace boundary —
  `TestASlugDeclaredOnlyOutsideTheWorkspaceIsUnresolved`.
- No document read twice — `TestTheWalkOffersEachCandidateExactlyOnce`
  (internal, holds the read side directly) plus
  `TestAColdBuildWithAFullGraphMeetsTheSameBudget` (1000 documents, 3000
  references, same 1s budget as AC-SPEC-009.3's reference-free corpus).
- Warm path — `TestAWarmCorpusServesTheGraphFromThePersistedIndex` plants a
  reference row in the persisted index that no document on disk carries and
  asserts a warm open serves it; `TestTheGraphIsIdenticalWarmAndCold` holds
  warm/cold parity for both directions.

### Gate results

| Gate | Exit |
|---|---|
| `cd cli && gofmt -l .` | 0, printed nothing |
| `cd cli && go vet ./...` | 0 |
| `cd cli && go build ./...` | 0 |
| `cd cli && go test ./core/... ./profile/...` | 0 |
| `cd cli && go test -count=1 ./...` | 0 (13 packages) |
| `scripts/generate-schemas.sh` | not run — no `schemas/*.strata` changed |

### Regeneration diffstat

`scripts/regenerate-cli-snapshots.sh`: three existing cases changed
`input_hash` only (`index-default`, `index-out-custom`, `index-out-stdout`) —
the expected consequence of folding the reference registration into the hash;
their `references` collections stay `[]` because the shipped profile registers
no reference property until PREG. The new case
`index-unresolved-reference` gained its two expected artifacts. No other
snapshot moved.

### Surprises

- The blueprint's RES anchors state the input hash "already folds the active
  profile's registration". It did not. Making that true was a required part of
  this delta rather than a given; without it a profile that registered a new
  reference property would have been served the index built before it.
- `module-document-core.md` still carried "goldmark is not a dependency yet and
  arrives with the Markdown Structure component" in Key Decisions — stale since
  REX. Corrected in the same edit, per the repository's rule that specs read as
  statements of current behaviour.
- "No document is read twice" has no direct observable from outside the package;
  it is held by an internal test on the one walk plus a budget measurement.
  Recorded rather than papered over.
