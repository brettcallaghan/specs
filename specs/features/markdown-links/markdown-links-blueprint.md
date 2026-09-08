---
title: "Feature Blueprint: Spec Corpus Markdown Links"
description: Adds a Markdown Links component to Document Core that
  substitutes resolved paths for wikilinks, two reachability finding
  variants, and a links subcommand that reads an optional YAML link map
slug: feature-spec-markdown-links
type: feature-blueprint
status: retired
updated_at: 2026-08-14T07:30:00Z
delivery: single-pr
plan: agent-plugins-ia5
frd: frd-spec-markdown-links
---

# Spec Corpus Markdown Links

## Overview

**FRD:** [[frd-spec-markdown-links]]

A new **Markdown Links** component in [[module-document-core]] takes a
document's bytes and the references the existing References component
already extracted from them, and emits the same bytes with each
targetable wikilink replaced by a markdown link. It reuses extraction
rather than rescanning, which is what makes the code-region exclusion
AC-SPECMD-001.4 requires a property of the input rather than a rule this
component has to get right a second time.

Two things sit either side of it. Deciding *where a link points* is a
separate judgement — a caller's map entry, or the cited document's path
relative to the citing one — and it is the judgement that produces the
two reachability findings, so it carries its own delta and its own
finding variants. Reading the caller's map is a file-format concern the
[[container-spec-cli#Command Surface]] already owns for `specs.toml`,
and it is read the same way, through the same `finding.Report`.

Delta blocks in dependency order: **FND** (finding variants), **TGT**
(target resolution), **MDL** (link substitution), **MAP** (link-map
input), **CMD** (the `links` subcommand).

## Element Deltas

### Findings in [[module-document-core]] (modified — variants) — FND

**Intent:** The two ways a citation fails to reach its target are
distinct defects with distinct fixes, and AC-SPECMD-002.4 requires a
consumer to tell them apart. `FindingDetail` is a closed union generated
from `schemas/finding.strata`, so each gets its own variant rather than
borrowing one that nearly fits — borrowing is the defect
`agent-plugins-nu2` records, where a `corpus/empty` finding reports
`detail.kind` as `frontmatter-parse` and misleads every JSON consumer.

**Behavior:**
- The module SHALL carry a `FindingDetail` variant for a citation whose
  slug names a corpus document the link map does not cover, holding the
  slug, the citing document's path and the map's path.
- The module SHALL carry a `FindingDetail` variant for a citation whose
  slug names no corpus document, holding the slug and the citing
  document's path.
- The module SHALL derive each variant's message from the variant's own
  fields, so a consumer branches on structure rather than parsing prose.

**Anchors:** `schemas/finding.strata` for the union;
`cli/core/finding` for the constructor per variant.
`scripts/generate-schemas.sh` regenerates `cli/core/specmodel`, which it
rewrites from scratch on every run (`agent-plugins-a7m`).

### Markdown Links in [[module-document-core]] (added — targets) — TGT

**Intent:** One judgement decides, per citation, what a link points at
or why it cannot. Keeping it apart from substitution is what lets the
two modes FEAT-SPECMD-005 and FEAT-SPECMD-006 describe share one
substitution path, and it puts both reachability findings in one place
rather than spread across the rewrite.

**New components** (declared in [[module-document-core]]'s vocabulary):

- **Markdown Links** (Go package) — Wikilink-to-markdown-link
  substitution over an already-extracted reference set, and the target
  resolution that decides what each link points at. Consumes the
  References component's output rather than rescanning a document, so
  code-region exclusion and anchor splitting are properties of the
  input. Writes nothing.

  ```component
  title: Markdown Links
  slug: component-markdown-links
  description: Wikilink-to-markdown-link substitution and target resolution
  ```

**Behavior:**
- IF the caller supplies a link map AND a reference's slug appears in it,
  THEN the component SHALL resolve that reference's target to the mapped
  path exactly as the map holds it (AC-SPECMD-005.5).
- IF the caller supplies a link map AND a reference's slug names a
  corpus document the map does not cover, THEN the component SHALL
  report a finding of severity `error` under the rule id
  `links/unmapped`.
- IF the caller supplies no link map, THEN the component SHALL resolve
  a reference's target to the cited document's repository-relative path
  taken relative to the directory of the citing document's own
  repository-relative path, with no other normalization.
- IF the caller supplies a link map, THEN the component SHALL NOT
  derive a target for any slug (AC-SPECMD-006.3).
- IF a reference's slug names no corpus document AND no link map entry
  covers it, THEN the component SHALL report a finding of severity
  `warning` under the rule id `links/unresolved`.
- IF a reference's slug names no corpus document AND no link map entry
  covers it, THEN the component SHALL resolve no target for it.
- The component SHALL resolve every reference in one pass and report
  every finding it raised, rather than returning at the first
  (AC-SPECMD-002.5).
- WHEN a reference carries an anchor, the component SHALL derive the
  target fragment by lowercasing the anchor text, removing every
  character that is not a letter, digit, space, hyphen or underscore,
  and replacing each space with a hyphen — GitHub's heading-anchor
  rule, named because the copies are read there and by agents reading
  raw markdown, and because a derivation nobody named would be
  reinvented differently by the next reader.
- The component SHALL take the citing document's repository-relative
  path from the caller rather than deriving it, so a document that is
  not a corpus member is resolvable the same way one that is.

**Anchors:** a new `cli/core/mdlinks` package;
`cli/core/references` for the `specmodel.Reference` values it consumes,
including each one's `Resolution`; `cli/core/finding` for the report it
raises findings through.

### Markdown Links in [[module-document-core]] (added — rewrite) — MDL

**Intent:** Producing the rewritten bytes. Every guarantee here is about
what is *not* touched: substitution replaces the exact extent each
reference occupies and copies everything else through, which is what
makes AC-SPECMD-001.5's "alters nothing else" and AC-SPECMD-004.1's
idempotence the same property rather than two.

**Behavior:**
- WHEN a reference has a resolved target, the component SHALL replace
  the whole `[[…]]` construct with a markdown link whose text is the
  reference's text as written and whose target is the resolved path.
- WHEN a reference carries an anchor, the component SHALL keep that
  anchor in the emitted link's text (AC-SPECMD-001.2) and append the
  derived fragment to the emitted target (AC-SPECMD-001.3).
- The component SHALL copy through every byte no reference's span
  covers, unchanged.
- WHEN a reference has no resolved target, the component SHALL copy its
  `[[…]]` construct through unchanged (AC-SPECMD-002.7).
- WHEN no reference resolves a target, the component SHALL emit the
  document byte-identical to its input (AC-SPECMD-001.6).
- The component SHALL report the number of references it replaced.
- The component SHALL NOT read or write any file.

**Anchors:** `cli/core/mdlinks`; `cli/core/span` for the extents it
replaces — a reference's `Position` covers the text between the
brackets, so the construct's extent is derived from it rather than
rescanned.

### Command Surface in [[container-spec-cli]] (modified — map) — MAP

**Intent:** The link map is caller-supplied file input, which is the
[[container-spec-cli#Command Surface]]'s concern and not a module's, on
the same rule that puts `specs.toml` there. It is read the same way —
every defect reported through `finding.Report` in one run, never
defaulted around — so a malformed map reads like every other
configuration defect.

**Behavior:**
- WHEN the caller names a link map, the command surface SHALL read it as
  YAML: a mapping whose keys are slugs and whose value per slug is a
  mapping carrying a `path` string (AC-SPECMD-005.4).
- The command surface SHALL ignore every property of a record other than
  `path` (AC-SPECMD-005.6).
- The command surface SHALL decode the file with a YAML parser, so a
  map written in JSON is accepted unchanged (AC-SPECMD-005.7).
- IF the named file cannot be read or parsed, or a record declares no
  `path`, THEN the command surface SHALL report a finding of severity
  `error` naming the file and the fault (AC-SPECMD-005.8).
- IF the named file cannot be read or parsed, or a record declares no
  `path`, THEN the command surface SHALL resolve no map.
- IF a slug appears twice as a key, THEN the command surface SHALL
  report that as a parse fault rather than letting one silently win.
- The command surface SHALL NOT interpret, resolve or normalize a
  `path` value.

**Anchors:** `cli/cmd`, beside `Load` and the `specs.toml` handling it
already carries; `goccy/go-yaml`, already a module dependency
([[adr-015]]).

### Command Surface in [[container-spec-cli]] (modified — links) — CMD

**Intent:** Wiring only. Per [[container-spec-cli]]'s Change Guidance no
judgement lands here: the subcommand converts flags into one call
against Markdown Links and its result into two streams and an exit
status.

**Behavior:**
- The command surface SHALL expose a `links` subcommand taking exactly
  one operand, the repository-relative path of the markdown document to
  rewrite (AC-SPECMD-005.1).
- The `links` subcommand SHALL accept an optional `--map` option naming
  a link map file.
- The `links` subcommand SHALL run without a `--map` option
  (AC-SPECMD-005.10).
- The `links` subcommand SHALL write the rewritten document to standard
  output (AC-SPECMD-005.2) and every finding and the replacement count
  to standard error (AC-SPECMD-003.2).
- The `links` subcommand SHALL report the replacement count as the
  number of citations rewritten in that document (AC-SPECMD-003.1).
- IF any finding of severity `error` was reported, THEN the `links`
  subcommand SHALL write no rewritten document (AC-SPECMD-002.2,
  AC-SPECMD-005.9).
- IF any finding of severity `error` was reported, THEN the `links`
  subcommand SHALL exit 1.
- IF every finding reported is of severity `warning`, THEN the `links`
  subcommand SHALL write the rewritten document (AC-SPECMD-002.6).
- IF every finding reported is of severity `warning`, THEN the `links`
  subcommand SHALL exit 0 (AC-SPECMD-002.8).
- The `links` subcommand SHALL create, modify and delete no file
  (AC-SPECMD-005.3).
- IF the named document cannot be read, THEN the `links` subcommand
  SHALL report an error finding and write no rewritten document.
- The `links` subcommand SHALL open the corpus warm, as `check`,
  `refs`, `slug-exists` and `next-id` do.
- The `links` subcommand SHALL accept no `--strict` option: warnings
  exit zero unconditionally under AC-SPECMD-002.8, so an option that
  reversed it would contradict the criterion rather than narrow it.
- The `links` subcommand's `--help` text SHALL be derived from the same
  declaration that defines it, per [[container-spec-cli]]'s Change
  Guidance.
- The `links` subcommand's one-line synopsis SHALL distinguish it from
  `refs`: `refs` answers what cites what, `links` emits one document
  with its citations rewritten.

**Anchors:** `cli/cmd`, the `cobra` command tree `Execute` builds;
`cli/testdata/cases/` for the snapshot cases, replayed by
`cli/cmd/snapshot_test.go`.

### Relied-upon elements (unchanged)

- [[module-document-core]] References — supplies every reference with
  its span, its slug and anchor halves, and its resolution, and excludes
  code regions. AC-SPECMD-001.4 is satisfied entirely by relying on it.
- [[module-document-core]] Corpus — supplies the warm index the
  resolution state is read from, and each member's repository-relative
  path.
- [[module-document-core]] Spans — supplies the extents substitution
  replaces; no position is computed in the new component.
- [[module-document-core]] Findings — supplies the report, the severity
  vocabulary and the ordering both new rules report through.
- [[container-spec-cli]] Command Surface — supplies configuration
  discovery, profile activation and the rendering the new subcommand
  reuses rather than reimplements.
- [[contract-specs-configuration]] — unchanged; the link map is a
  per-invocation input, not configuration, and adds no key to
  `specs.toml`.

## Contracts

None. The link map is read from a path the caller names per invocation
rather than published for another codebase to build against, and no
existing contract changes. WHEN a second producer or an independent
consumer of the link map appears, that judgement is revisited under this
framework's filing rule.

## Tracker Reconciliation

- `agent-plugins-eaa.1` — out-of-scope: it backticks the ~50
  illustrative wikilinks in the documents bound for `specs/`. This
  feature must not edit anyone's prose; it reports each one as a
  `links/unresolved` warning, which is the evidence that work acts on.
  The two are complementary and neither blocks the other.
- `agent-plugins-eaa.2` — out-of-scope: it moves the 8 shipped standards
  into `specs/standards` and inverts the sync direction. That is what
  makes this repository's own documents worth rewriting; it is not what
  the capability is built against, and this feature is tested on
  fixtures rather than on the move.
- `agent-plugins-47b` — out-of-scope: it verifies that a committed
  markdown link lands on a co-located file. This feature checks
  reachability at rewrite time, which cannot guard what is already
  committed. Boundary: `47b` reads the rewritten copies; it does not
  reimplement resolution.
- `agent-plugins-nu2` — out-of-scope: an existing finding reuses an
  ill-fitting `FindingDetail` variant. The FND delta adds its own two
  variants rather than repeating that; it does not fix `corpus/empty`.
- `agent-plugins-a7m` — out-of-scope: `scripts/generate-schemas.sh`
  rewrites `cli/core/specmodel` from scratch, so the FND delta places no
  hand-written file there. It works within the constraint rather than
  lifting it.

## Sequencing Constraints

- The FND delta SHALL land before the TGT delta and before the MAP
  delta, because both raise findings through variants FND introduces.
- The TGT delta SHALL land before the MDL delta, because substitution
  replaces only references target resolution resolved.
- The MAP delta SHALL land before the CMD delta, because the subcommand
  passes a resolved map into the module.
- Delivery is `single-pr`: all five deltas land in one change, which is
  how [[container-spec-cli]]'s Change Guidance rule that a new command's
  `--help` derive from its own declaration is satisfied across the CMD
  delta and the snapshot cases the same delta regenerates.

## AC Coverage

Legend: FND = Findings delta (variants); TGT = Markdown Links delta
(targets); MDL = Markdown Links delta (rewrite); MAP = Command Surface
delta (map); CMD = Command Surface delta (links).

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-SPECMD-001.1 | TGT, **MDL (proves)** | unit |
| AC-SPECMD-001.2 | **MDL (proves)** | unit |
| AC-SPECMD-001.3 | **TGT (proves)**, MDL | unit |
| AC-SPECMD-001.4 | **MDL (proves)** — via relied-upon References | unit |
| AC-SPECMD-001.5 | **MDL (proves)** | unit |
| AC-SPECMD-001.6 | **MDL (proves)** | unit |
| AC-SPECMD-002.1 | FND, **TGT (proves)** | unit |
| AC-SPECMD-002.2 | TGT, **CMD (proves)** | integration |
| AC-SPECMD-002.3 | **FND (proves)**, TGT | unit |
| AC-SPECMD-002.4 | **FND (proves)**, TGT | unit |
| AC-SPECMD-002.5 | **TGT (proves)** | unit |
| AC-SPECMD-002.6 | TGT, **CMD (proves)** | integration |
| AC-SPECMD-002.7 | TGT, **MDL (proves)** | unit |
| AC-SPECMD-002.8 | **CMD (proves)** | integration |
| AC-SPECMD-003.1 | MDL, **CMD (proves)** | integration |
| AC-SPECMD-003.2 | **CMD (proves)** | integration |
| AC-SPECMD-004.1 | **MDL (proves)** | unit |
| AC-SPECMD-004.2 | **MDL (proves)** | unit |
| AC-SPECMD-005.1 | **CMD (proves)** | integration |
| AC-SPECMD-005.2 | **CMD (proves)** | integration |
| AC-SPECMD-005.3 | **CMD (proves)** | integration |
| AC-SPECMD-005.4 | **MAP (proves)** | unit |
| AC-SPECMD-005.5 | **TGT (proves)** | unit |
| AC-SPECMD-005.6 | **MAP (proves)** | unit |
| AC-SPECMD-005.7 | **MAP (proves)** | unit |
| AC-SPECMD-005.8 | FND, **MAP (proves)** | unit |
| AC-SPECMD-005.9 | MAP, **CMD (proves)** | integration |
| AC-SPECMD-005.10 | **CMD (proves)** | integration |
| AC-SPECMD-006.1 | **TGT (proves)** | unit |
| AC-SPECMD-006.2 | **TGT (proves)** | unit |
| AC-SPECMD-006.3 | **TGT (proves)** | unit |

## Key Decisions

None. Every choice made while authoring this blueprint is a behavior
pinned in a delta above rather than a decision with live alternatives:
the anchor derivation follows a named external rule, the command name
and the two rule ids are naming, and the module-versus-command-surface
split is [[container-spec-cli]]'s existing Change Guidance applied
rather than a new decision.

## Definition of Done

- [x] Every delta above is folded into its owning blueprint:
  - [x] [[module-document-core]] — Components roster gains the Markdown
    Links record block and its **Built** paragraph; the component
    diagram gains its edges to References, Spans and Findings; the
    Findings section's `FindingDetail` vocabulary gains both variants
  - [x] [[container-spec-cli]] — the Command Surface component's
    command-tree paragraph names `links` as a sixth subcommand, and the
    Container Context's read-only statement is re-affirmed against
    AC-SPECMD-005.3
- [x] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]] — all 31
  carry a citation: 22 from a `// AC:` comment on a unit test, 9 from a
  snapshot case manifest's `cites`, which [[testing-standards]]
  Acceptance-Criteria Traceability accepts equally. One exception:
  AC-SPECMD-005.3 is declared `integration` and is cited by
  `cli/cmd/links_test.go`'s `TestLinksCreatesModifiesAndDeletesNoFile`
  rather than by a snapshot case, because a case is pure data and
  cannot assert that no file appeared.
- [x] Snapshot cases under `cli/testdata/cases/` cover the `links`
  subcommand, regenerated by `scripts/regenerate-cli-snapshots.sh` with
  every resulting diff reviewed — nine of them: `links-no-map`,
  `links-map`, `links-map-empty`, `links-map-incomplete`,
  `links-unresolved-warning`, `links-unreadable-map`,
  `links-unreadable-operand`, `links-no-strict-flag`, `help-links`
- [x] `go test ./core/...` and `go test ./cmd/...` pass, and the profile
  self-application gate in `cli/profile` still runs clean
- [x] `specs check --strict` exits 0 over this repository's corpus
- [x] This blueprint's status → retired

## Open Questions

None.

## References

- `plugin/skills/spec/references/c4-guide.md` — the component and module
  definitions the delta blocks and component references are written
  against
- [[frd-spec-markdown-links]] — the requirements this design realizes
- [[container-spec-cli]] — the owning container, whose Change Guidance
  places the capability in a module and the wiring in the command
  surface
- [[module-document-core]] — the module the new component joins, and the
  boundary rule that keeps framework vocabulary out of it
- [[frd-spec-reference-integrity]] — the extraction and resolution this
  feature consumes rather than reimplements
- [[adr-018]] — references as one edge collection, which is what makes a
  reference's resolution available without re-reading a document
