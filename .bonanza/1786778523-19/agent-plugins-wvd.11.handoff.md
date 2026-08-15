# agent-plugins-wvd.11 handoff

## What shipped

- `cli/core/gherkin/render.go` — `Render(RenderFeature) []byte`, `Parse`'s
  inverse. `RenderFeature`/`RenderCriterion` carry tags, name, description,
  rule and text as strings; the package holds no vocabulary, matching
  `Options`'s own separation. Every criterion becomes its own top-level
  `Scenario`/`Scenario Outline` construct (retro decision below) —
  `Scenario Outline` chosen wherever `Text` carries a line opening an
  `Examples` block, detected via the same dialect the parser reads.
- `cli/core/gherkin/render_test.go` — `TestRenderIsParsesInverse`
  round-trips a fixture covering a Feature description, a Rule-grouped
  criterion with its own description and a doc string, a plain criterion,
  and an outline whose two `Examples` blocks are separately identified;
  compares every `Feature`/`Criterion` field but `Position` (AC-GHK-009.2,
  unit). Two more unit tests pin the owning-tag line and the
  Scenario/Scenario Outline keyword choice (AC-GHK-009.1).
- `cli/cmd/feature.go` — the `feature <id>` subcommand, carrying only
  `--emit` and `--out`: looks the feature up in the warm corpus index the
  same way `refs`/`slug-exists` do, converts it to tags via the profile's
  `FeatureTagRules` (identifier tags plus registered qualifier tags for
  `Negative`/`Manual`/`Level`), and writes `gherkin.Render`'s bytes to
  stdout or the named path. `spec feature <id>` with no `--emit` reports
  that querying is not yet supported rather than guessing at QRY's surface.
- `cli/testdata/cases/feature-emit-fenced-block` and
  `feature-emit-not-found` — snapshot cases; the first proves the
  owning-document tag is added even though its source is a fenced block
  that carries none (AC-GHK-009.1, integration), the second the not-found
  path exits 1 with a distinct message.
- `specs/architecture/container-spec-cli.md` — Command Surface records
  `feature`, `--emit` and `--out`, and the command-tree paragraph and the
  bare-invocation paragraph no longer say "six subcommands".
- `scripts/regenerate-cli-snapshots.sh` run; `help-root`'s expected
  `checks.log` picked up the new `feature` line, nothing else changed.

## Retro decision (internal shape)

Each `specmodel.Criterion` renders as its own independent
`Scenario`/`Scenario Outline`, even where several criteria came from one
`Scenario Outline`'s separately-tagged `Examples` blocks in the source.
`specmodel.Criterion` carries no link back to a shared outline once
extracted, so nothing could recover that grouping; duplicating the shared
steps across N single-block scenarios costs nothing a corpus tool reads
(`Text` already stands alone per criterion) and — checked directly, not
just argued — reparsing never re-divides them, because every one of a
criterion's tags is written on its own `Scenario`/`Scenario Outline` line
rather than split between an outline header and a block, so the
`Examples` block on re-parse always carries none of its own and the
criterion resolves from the scenario-level tags alone.

## Verification

- `cd cli && gofmt -l .` — clean
- `cd cli && go build ./...` — clean
- `scripts/check-skills.sh --strict` — exit 1, 0 errors / 1 warning
  (pre-existing baseline, untouched by this change)
- `scripts/sync-shared-references.sh --check` — pass
- `scripts/sync-strata-docs.sh --check` — pass
- `cd cli && go test ./...` — pass
- `scripts/generate-schemas.sh --check` — pass
- `scripts/regenerate-cli-snapshots.sh` — ran; only `help-root` diffed,
  reviewed and expected

Caveat: `spec feature` as a plain query (no `--emit`) is not built — it
reports "querying without --emit is not yet supported" and exits 1. QRY
(agent-plugins-wvd.10) extends this same `newFeatureCmd`/`feature.go`
with the identifier-only lookup, `--code`, `--of`, `--tag` and `--json`;
it should not need to touch `--emit`/`--out` or `gherkin.Render`.
