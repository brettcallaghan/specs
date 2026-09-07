---
title: "Feature Blueprint: Corpus Listing"
description: How the Spec CLI grows a listing surface — a summary query in
  Document Core, a thin list subcommand over it, and the authoring
  guidance that makes description a consumed interface
slug: feature-spec-list
type: feature-blueprint
status: retired
updated_at: 2026-09-07T00:00:00Z
delivery: single-pr
frd: frd-spec-list
plan: agent-plugins-kgir
---

# Corpus Listing

**FRD:** [[frd-spec-list]]

Every value this feature reports is already in the index. The design is
therefore a query, not a new artifact: Document Core gains a summary
query over the entries it already holds, and the Command Surface gains a
subcommand that turns flags into that query's options and hands the
result to the YAML renderer two other subcommands already use. Nothing
parses a document, nothing writes a file, and the index's shape does not
change.

The layering is the whole of the design. [[container-spec-cli]]'s Change
Guidance forbids a capability living in the command surface, so the
filtering and the summary shape are Document Core's, and `cli/cmd` only
declares flags and selects a rendering. Two profile-registered tables —
the document types and the lifecycle's retired statuses — are what let
that filtering live in Document Core without it learning any framework
vocabulary.

Delta blocks in dependency order: CORE, then FILT, then CMD, with DOC
independent of all three.

## Element Deltas

### `Corpus` in [[module-document-core]] (modified — summary query)

**Intent:** The listing needs one query returning a summary per corpus
member. It belongs beside `Features` and `Criteria` in the same query
surface, for the same reason those do: the Command Surface may hold no
domain logic, and a future API or MCP wrapper needs the same answer
without going through cobra.

The summary is a projection of `IndexEntry`, not a new record. It reads
the entry's path and its parsed frontmatter envelope and carries nothing
else, so no index shape changes and `schema_version` stays `1.0` — the
Change Guidance rule about the emitted index is not triggered by this
feature.

**Behavior:**
- The query SHALL return one summary per corpus member, carrying the
  member's path, slug, type, status, title and description in that
  order.
- WHERE the index recorded a member with a parse error, the query SHALL
  return a summary carrying that member's path and the parse error's
  message and position, and SHALL NOT carry a slug, type, status, title
  or description for it — an absent value here would assert the document
  declares none, which is not known.
- The query SHALL NOT return anything for a file the index recorded as
  an exclusion.
- The query SHALL return summaries in the order the index holds its
  entries, which the emitter already fixes by repository-relative path.
- WHERE the corpus holds no members, the query SHALL return an empty
  collection rather than an error.
- The query SHALL report each path exactly as the index holds it —
  relative to the repository root, with forward slashes.

**Anchors:** `cli/core/corpus/query.go`, beside `Corpus.Features` and
`Corpus.CriteriaOf`; the entry shape is
`cli/core/specmodel`'s `IndexEntry`, generated from
`schemas/index.strata`.

### `Corpus` in [[module-document-core]] (modified — listing filters)

**Intent:** Filtering is what keeps the answer proportionate as a corpus
grows, so it belongs in the query rather than in a caller that discards
rows after receiving them — a caller-side filter would make every
consumer reimplement it, and would make the command surface hold the
policy.

Both filters resolve against tables the profile registers, which is what
keeps this in Document Core without importing framework vocabulary. The
type filter validates against `Profile.DocumentTypes`; the retired
filter reads `LifecycleRules.RetiredStatuses` rather than testing for the
literal string `retired`, so a profile whose documents are `current` and
`sunset` gets the same behavior. Hard-coding the spelling here is the
one change that would violate the module's own dependency rule.

**Behavior:**
- WHERE the caller names one or more types, the query SHALL return only
  summaries whose type is among them, and SHALL return every such
  summary.
- WHERE the caller names a value that is not a key of the active
  profile's registered document types, the query SHALL report that value
  and the registered type names, and SHALL return no summaries.
- WHERE a type filter is in effect, the query SHALL NOT return a summary
  for a member recorded with a parse error, since such a member declares
  no type for a filter to match.
- WHERE the caller does not ask for every status, the query SHALL omit
  summaries whose status is among the active profile's registered
  retired statuses, and SHALL return summaries at every other status.
- WHERE the caller asks for every status, the query SHALL return
  summaries at every status, retired ones included.
- The two filters SHALL compose: a query carrying both SHALL return only
  summaries satisfying both.
- WHERE a filter matches no member, the query SHALL return an empty
  collection rather than reporting a failure.

**Anchors:** `cli/core/corpus/query.go` for the filtering;
`cli/core/corpus/profile.go` for `Profile.DocumentTypes` and
`LifecycleRules.RetiredStatuses`, both of which already exist and are
relied upon unchanged.

### `Command Surface` in [[container-spec-cli]] (modified — `list`)

**Intent:** A tenth subcommand, wired like the nine before it: flags in,
one Document Core call, one rendering out. It decides nothing about what
a summary is or which members qualify.

`list` opens the corpus warm through the shared `openCorpus` helper
rather than building fresh, for the reason `check` does — a listing is a
query, and the persisted index at the default location answers it
without re-walking the corpus.

**Behavior:**
- The subcommand SHALL be named `list` and SHALL accept no positional
  arguments.
- The subcommand SHALL accept `--type`, repeatable and accepting a
  comma-separated value, naming the document types to return.
- The subcommand SHALL accept `--all`, which asks for summaries at every
  status including retired ones; without it, retired summaries are
  omitted.
- The subcommand SHALL accept `--json`.
- The answer SHALL be a mapping carrying one key, `documents`, whose
  value is the sequence of summaries.
- The subcommand SHALL render the answer as YAML, and as JSON under
  `--json`, both carrying the same values in the same order.
- WHERE the query reports an unregistered type, the subcommand SHALL
  write that report to standard error and SHALL exit non-zero.
- The subcommand SHALL write no count of the summaries, no heading
  naming the command and no trailing summary.
- Every flag's `--help` text SHALL come from the flag declaration
  itself, per the container's Change Guidance.

**Anchors:** a new `cli/cmd/list.go` beside `cli/cmd/headings.go`, which
is the first subcommand written against [[cli-output-standards]] and is
the shape to follow; registration in `newRootCmd` in `cli/cmd/root.go`;
snapshot cases under `cli/testdata/cases/`.

### Documentation guidance (modified)

**Intent:** `description` acquires a second audience the day this command
ships. Both documents that tell an author what to write in it, and a
reader what the CLI offers, are stale until they say so.

**Behavior:**
- [[frontmatter-standards]]' guidance on `description` SHALL state that
  the listing command reports it and that it is written to be triaged
  from rather than read as an introduction.
- [[spec-cli-reference]] SHALL document the `list` subcommand, its
  filters, and the fields a summary carries.

**Anchors:** `specs/standards/frontmatter-standards.md` (the required
properties table and the prose beneath it);
`specs/guides/spec-cli-reference.md`. Both are vendored into the `spec`
skill by `scripts/sync-shared-references.sh`, which SHALL be re-run so
the generated copies under `plugin/skills/*/references/shared/` do not
drift.

### Relied-upon elements (unchanged)

- [[cli-output-standards]] — the rendering rules the CMD delta conforms
  to rather than restates; `renderYAML` in `cli/cmd/yaml.go` already
  implements the YAML style rules, including the double-quoting that
  AC-SPECLIST-004.4 depends on.
- [[frd-spec-index]] — the index this feature reads. Its entry order,
  its repository-relative forward-slash paths, and its recording of
  parse failures and exclusions are all relied upon and none change.
- `Profile.DocumentTypes` and `LifecycleRules.RetiredStatuses` in
  `cli/core/corpus/profile.go` — the registered tables both filters
  resolve against.
- [[module-spec-profile]] — supplies those tables for this repository's
  profile; unchanged, since it already registers every document type and
  the retired status.
- `openCorpus` in `cli/cmd/root.go` — the warm-open helper the
  subcommand uses.

## Contracts

None. This feature reads the index and introduces no interface of its
own. The index's shape is unchanged, so the deferred question of
promoting it to a contract document is untouched.

## Tracker Reconciliation

- `agent-plugins-gz3` — out-of-scope: index entries carry only the
  `DocumentFrontmatter` envelope, losing per-type properties. Every
  field a summary carries is on that envelope, so this feature neither
  depends on that issue nor is blocked by it. The boundary: `list` SHALL
  NOT be extended to report per-type properties as part of this feature.
- `agent-plugins-1ao` — out-of-scope: a feature blueprint for the spec
  CLI as a whole, against `frd-spec-validation-cli`. This blueprint
  realizes `frd-spec-list` only and must not absorb that scope.
- `agent-plugins-h6c` — out-of-scope: the index schema and versioning
  ADR. This feature makes no index shape change, so it neither advances
  nor blocks that decision.
- `agent-plugins-lyqk` — subsumed by this feature's upstream FRD, and
  already closed.

## Sequencing Constraints

- The CORE delta SHALL land before the FILT delta, which extends the
  query CORE introduces.
- The CORE and FILT deltas SHALL both land before the CMD delta, which
  invokes the query they define.
- Delivery is `single-pr`: the DOC delta and the Definition of Done's
  fold of the CMD delta into [[container-spec-cli]] SHALL land in the
  same change as the code, so the main branch never carries a command
  the container blueprint does not list.
- WHEN the CMD delta changes observable output, the snapshot cases under
  `cli/testdata/cases/` SHALL be regenerated with
  `scripts/regenerate-cli-snapshots.sh` and every resulting diff reviewed
  before commit, per the container's verification gates.

## AC Coverage

Legend: CORE = `Corpus` summary-query delta; FILT = `Corpus`
listing-filters delta; CMD = `Command Surface` `list` subcommand delta;
DOC = documentation guidance delta.

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPECLIST-001.1 | **CORE (proves)** | unit |
| AC-SPECLIST-001.2 | CORE, **CMD (proves)** | e2e |
| AC-SPECLIST-001.3 | CORE, **CMD (proves)** | e2e |
| AC-SPECLIST-001.4 | **CORE (proves)** | unit |
| AC-SPECLIST-001.5 | **CORE (proves)** | unit |
| AC-SPECLIST-001.6 | **CORE (proves)** | unit |
| AC-SPECLIST-001.7 | CORE, **CMD (proves)** | e2e |
| AC-SPECLIST-002.1 | **FILT (proves)** | unit |
| AC-SPECLIST-002.2 | FILT, **CMD (proves)** | e2e |
| AC-SPECLIST-002.3 | **FILT (proves)** | unit |
| AC-SPECLIST-002.4 | FILT, **CMD (proves)** | e2e |
| AC-SPECLIST-002.5 | **FILT (proves)** | unit |
| AC-SPECLIST-003.1 | **FILT (proves)** | unit |
| AC-SPECLIST-003.2 | FILT, **CMD (proves)** | e2e |
| AC-SPECLIST-003.3 | **FILT (proves)** | unit |
| AC-SPECLIST-003.4 | **FILT (proves)** | unit |
| AC-SPECLIST-004.1 | **CMD (proves)** | e2e |
| AC-SPECLIST-004.2 | **CMD (proves)** | e2e |
| AC-SPECLIST-004.3 | CORE, **CMD (proves)** | e2e |
| AC-SPECLIST-004.4 | **CMD (proves)** — via relied-upon `renderYAML` | e2e |
| AC-SPECLIST-004.5 | **CMD (proves)** | e2e |
| AC-SPECLIST-005.1 | **DOC (proves)** | manual |
| AC-SPECLIST-005.2 | **DOC (proves)** | manual |

## Key Decisions

None. Every choice this design makes is either dictated by an existing
approved document — the layering rule and verification gates in
[[container-spec-cli]]'s Change Guidance, the rendering rules in
[[cli-output-standards]] — or is a behavior pinned in a delta above.
Nothing here has alternatives whose trade-offs outlive the feature, so
nothing warrants an ADR of its own.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[container-spec-cli]] — the Command Surface component's
        subcommand vocabulary grows by `list`, with its flags and the
        Document Core call it makes
  - [x] [[module-document-core]] — the query surface gains the summary
        query and its filters
- [x] [[frontmatter-standards]] states `description`'s second audience,
      and `scripts/sync-shared-references.sh` has been re-run so the
      vendored copies match
- [x] [[spec-cli-reference]] documents `list`
- [x] Every AC in the coverage table is verified at its declared level
      and cited from its tests per [[ac-traceability-standards]]; the two
      `manual` rows are exempt from the citation obligation
- [x] `scripts/coverage.sh` has been run from the repository root and the
      README diff committed with the change, since `cli/**` changed
- [x] Every issue listed as subsumed in Tracker Reconciliation is closed
      or re-scoped
- [x] This blueprint's status → retired

**Verification note (recorded at retirement).** Every entry above was
verified, with one declared-level deviation. AC-SPECLIST-004.3 is
declared `e2e` in the AC Coverage table, but the snapshot harness
replays each case once and cannot express a two-run comparison; it is
verified instead by a Go test in `cli/cmd/list_test.go`. Citation is
satisfied and the level obligation is a SHOULD in
[[ac-traceability-standards]], so this is a permitted deviation rather
than a gap — recorded here, and proposed as a standards change in
`rfcs/inbox/standards/rfc-0027-two-standards-one-citation-surface.md`.

Citations for the 21 non-`@manual` criteria live on two surfaces: `// AC:`
comments under `cli/` for 11 of them, and the `cites` key of the snapshot
manifests under `cli/testdata/cases/` for the other 10. A grep for the
comment form alone reports half the coverage that exists.

## Open Questions

None.

## References

- `c4-guide.md` — the container, module and component definitions the
  delta blocks are written against
- [[frd-spec-list]] — the requirements this blueprint realizes
- [[container-spec-cli]] — the container whose Change Guidance binds
  every delta above
- [[module-document-core]] — the module the query lives in
- [[cli-output-standards]] — the rendering rules the CMD delta conforms
  to
- [[ac-traceability-standards]] — the citation obligations the coverage
  table feeds
