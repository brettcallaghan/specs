# agent-plugins-wvd.9 handoff

Narrowed the References component's code-region mask by fence info
string instead of removing it, and extended extraction to `.feature`
sources.

`cli/core/references/code.go`: `codeRegions` now looks up each fenced
block's info string against `gherkin.FenceTags()` (a `map[string]bool`
built once per call via a new `tagSet` helper). When the tag matches, the
block's content lines are left out of the mask; the block's info-string
segment is still always masked, regardless of tag, since it is notation
naming the block and never a place Gherkin content is written. Every
other fenced block, every indented code block, and every inline code
span are masked exactly as before.

`cli/core/corpus/features.go`: `visitFeatureSource` now calls
`references.Extract` with `Path: reported, Source: src` (no
`Frontmatter`) before attempting `gherkin.Parse`, so extraction runs
whether or not the Gherkin itself parses, and a `.feature` file's whole
content is read as body — it has no frontmatter to skip. Its references
flow into `b.references` and are resolved and reported the same way
every other reference is, via the existing `reference/unresolved` rule.
No change was needed to anchor handling: this codebase never resolves an
anchor against a heading anywhere, so "not resolved against a Feature,
Rule or Scenario name" was already true and only needed proving.

Tests added: four unit tests in `cli/core/references/references_test.go`
citing AC-GHK-004.1 through .4 (wikilink positions inside a `gherkin`
fence, span in the containing document, resolution parity with prose via
`references.Resolve`, and anchor split without Gherkin-structure
interpretation), plus a new snapshot case
`cli/testdata/cases/check-gherkin-reference-unresolved` (citing
AC-GHK-004.3, the "case" half of that AC's "case + unit" verification)
proving a wikilink inside a `.feature` file is reported
`reference/unresolved` against that file's own path.

Folded into `specs/architecture/module-document-core.md`'s References
component: two new bullets under "the extraction half" (mask narrowing
by `gherkin.FenceTags`; `.feature` source extraction against its own
path), a short addition to the existing anchor-split bullet, and updated
summary prose — plus the "Seven semantics" counter corrected to "Nine".

Gates: gofmt clean; `go build ./...` clean; `go test ./...` green;
`scripts/sync-shared-references.sh --check` and
`scripts/sync-strata-docs.sh --check` and
`scripts/generate-schemas.sh --check` all green; `check-skills.sh
--strict` at the documented pre-existing baseline (exit 1, 0 errors / 1
warning — bead-bonanza SKILL.md size, not touched by this change). A
full `spec check` run against this repository's own corpus after the
doc edit reports no findings.

Caveat: none.
