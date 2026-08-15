# agent-plugins-wvd.6 handoff

Populated `corpus.Profile.FeatureTags()` for the spec-driven-dev profile
(`cli/profile`) with this framework's Gherkin tag vocabulary, per the
blueprint's PREG delta:

- Feature identifier pattern `^[A-Z]+-[A-Z0-9]+-[0-9]{3}$` and criterion
  identifier pattern `^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`, registered as
  `featureIDPattern`/`criterionIDPattern` in `cli/profile/profile.go`,
  each the generated `FeatureId`/`CriterionId` pattern
  (`schemas/feature.strata`) widened with named capture groups `code`
  and `number` — naming a group changes nothing a pattern matches.
- Owning-document tag prefix `frd-`.
- The derivation rule (a criterion's code and number are its feature's)
  registered as `SharedCaptures: []string{"code", "number"}`.
- Qualifiers `negative`, `manual`, `level-unit`, `level-integration`,
  `level-e2e`, mapped to `specmodel.QualifierMeaning`/`VerificationLevel`
  — a literal Go slice of `corpus.QualifierTag`, the implicit-decision
  data shape (the vocabulary is five fixed, authored tags with no
  generated source to derive a table from, unlike the reference-property
  and status derivations elsewhere in this package).

Added `TestFeatureTagPatternsMatchTheGeneratedIdentifiers` in
`cli/profile/registration_test.go`: it decodes the embedded schema
independently (matching this file's existing untyped-traversal
convention), strips named-capture-group naming from the registered
patterns, and asserts the result equals the generated `FeatureId`/
`CriterionId` pattern exactly — the drift guard the bead required.
Added two further coverage tests for the owner prefix / shared captures
and for the qualifier table. `TestTheRegistrationHasNoDefects` already
exercises the new registration through `corpus.ValidateRegistration`,
which now runs `featureTagDefects` against non-empty `FeatureTagRules`
for the first time.

Folded the delta into `specs/architecture/module-spec-profile.md`:
Interface gained a guarantee bullet for the tag registration and a
paragraph on `FeatureTags()`'s content and its drift guard; Profile
Registration's member table and mermaid diagram gained the fifth
member; Key Decisions' `adr-021` line no longer says no component here
registers AC rules.

No disagreement found between `specs/standards/gherkin-guide.md` and the
registered patterns — the guide states the tag shapes in prose
(`@<PREFIX>-<CODE>-<nnn>`, `@AC-<CODE>-<nnn>.<n>`) and the generated
`FeatureId`/`CriterionId` patterns already match them.

Gates: `gofmt -l .` clean; `cd cli && go build ./...` clean; `cd cli &&
go test ./...` all green; `scripts/check-skills.sh --strict` exits 1
with 0 errors / 1 warning (bead-bonanza SKILL.md body size) — the
documented pre-existing baseline, unchanged by this bead;
`scripts/sync-shared-references.sh --check` exit 0;
`scripts/sync-strata-docs.sh --check` exit 0;
`scripts/generate-schemas.sh --check` exit 0 (no schema touched).

Caveat: none.
