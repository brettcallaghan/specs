# agent-plugins-wvd.10 handoff

Added the QRY query surface to `spec feature` and `spec criterion`,
extending the commands EMIT (`agent-plugins-wvd.11`) and CHG
(`agent-plugins-wvd.12`) left in place — `--emit`/`--out`/`--changed` are
untouched.

`cli/core/corpus/query.go` adds six read-only `*Corpus` methods —
`Feature`, `Features`, `FeaturesByCode`, `Criterion`, `CriteriaOf`,
`CriteriaByTag` — each a linear scan over the already-in-memory
`Index().Features`, so the lookup/selection logic lives in Document Core
rather than `cli/cmd` (the "no domain logic in cli/cmd" rule).
`CriteriaByTag` resolves a `--tag` string against the active profile's
`FeatureTagRules.Qualifiers`, the same registration `--emit`'s
`criterionTags` reads in the write direction.

`cli/cmd/feature.go` gained `--code` and `--json`; `cli/cmd/criterion.go`
gained `--of`, `--tag` and `--json`; both gained an optional positional
`<id>`. Both `Args` validators and `RunE` branch on the pre-existing mode
flag (`--emit`, `--changed`) before doing anything else, so the new query
path and the old path never share logic. An identifier lookup that finds
nothing writes `<id>: not found` to stderr and exits 1 (AC-GHK-007.4); a
`--code`/`--of`/`--tag` selection that matches nothing is a valid empty
answer and exits 0 — an implicit decision, since the FRD's "an identifier
names no record" scenario is worded only around the positional
identifier lookup, not a filter value.

`--json` marshals the matching `specmodel.Feature`/`specmodel.Criterion`
value (single lookup) or slice (selection) directly — "the generated
JSON" the blueprint names, no new result-envelope type. Text rendering
is an implicit decision, recorded in a `retro:` comment: tab-separated
`<id>\t<name>` per record, matching `criterion --changed`'s existing
tab-separated shape; a feature's criteria are listed the same way,
indented one leading tab.

AC-GHK-007.6 (warm-index, no re-walk) is proved by
`cli/cmd/query_test.go`: it primes a corpus, plants a feature/criterion
into the persisted index that exists nowhere on disk, and shows a second
query run returns it — only possible if the corpus was served warm.

12 new snapshot cases under `cli/testdata/cases/` cover both identifier
lookup, `--code`/`--of`/`--tag` selection (including two exit-0 empty
selections), `--json`, not-found (exit 1), and `--help` for both
commands. `scripts/regenerate-cli-snapshots.sh` was run before tests;
the only pre-existing artifact it touched is `help-root/expected/checks.log`
(the two commands' `Short` text), reviewed and correct.

`specs/architecture/container-spec-cli.md`'s Command Surface section is
updated in place for both commands' new corpus calls, exit-status rules
and flag-combination validation; Change Guidance's existing verification
gates already cover `cli/cmd` changes generically, so no edit was needed
there.

All required gates pass; `check-skills.sh --strict` reproduces the known
pre-existing baseline exactly (exit 1, 0 errors, 1 warning — the
bead-bonanza SKILL.md size warning, unrelated to this change).

Caveat: `--tag`'s value is validated only against the active profile's
registered qualifier vocabulary at query time (an unregistered tag
matches nothing, exit 0); it is not validated against any fixed enum, so
a typo'd qualifier is indistinguishable from a real one with zero
carriers — both report "no matching criteria".
