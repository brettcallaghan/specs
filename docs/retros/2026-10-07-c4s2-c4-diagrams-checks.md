# Run record — C4 diagrams stage 2 (element resolution and `c4` checks)

## Read this first

**Every admitted bead closed, and the ten `c4/*` rules are live in `specs
check`.** 16 admitted, 15 landed as commits, 1 (`.15`) closed by
design at planning. Review ran two passes; the second admitted nothing, and
**12 review findings sit in the backlog** (see Scope) — 2 at P3, the rest P4.

**The tracker push was not held.** Brett asked for the final tracker push of
closures to wait for his word (`standing_auth.note`); `bb-close-bead.sh`
pushes on every close, so every closure was pushed as it happened
(`c4s2#4`). Candidate A.

**Accept needed hand repair twice**, both recurrences of issues already in
`outbox/swarm/swarm/` from the C4 diagrams stage 1 run (`c4s2#1`, `c4s2#3`).

**All candidates are PENDING** — put to Brett by the coordinator, no answer
recorded. Nothing has been filed and no `retro:filed:` comment written.

## The run

| | |
|---|---|
| run epic | "C4 diagrams stage 2: element resolution and c4 checks" (feature `feature-spec-c4-diagrams`, deltas XPT, PREG, ELEM, DRAW, CHKN, CHKR, DOC). The epic itself is still `open` |
| branch | the sprint branch, **not pushed** at the time of writing; the coordinator pushes after the retro |
| `base_sha` → head | `3e612b9` → `901c491` — 15 commits (before this record's commit) |
| queue | 8 admitted at the gate (`.1`–`.8`); +8 admitted by review pass 1 (`.9`–`.16`); pass 2 admitted 0. 16 / 16 closed |
| declared checks | `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| tiers (bead `metadata.tier`, set at plan) | advanced: `.1`, `.3`, `.5`; standard: `.2`, `.4`, `.6`, `.7`, `.8`, `.10`, `.11`, `.13`, `.14`; light: `.9`, `.12`, `.16`; `.15` none (not dispatched) |
| model lineup | **`manifest.models` is not set** (`swarm manifest get --key models` → `ok: false`). Models served, per the coordinator: `dev-c4s2-1` opus, `dev-c4s2-2` sonnet, `dev-c4s2-3` sonnet, `dev-c4s2-4` sonnet; plan, review and retro agents opus |
| effort | **`manifest.effort` is not set** — unrecorded |
| reviewer | declared `code-review` (`[review]`); `review_budget` 2 (default); `manifest.review_passes` 2 |
| concurrency cap | 20 |
| dispatches | `dev-c4s2-1` (`.1`, `.2`, `.3`, `.4`, `.5`, `.7`), `dev-c4s2-2` (`.6`), `dev-c4s2-3` (`.8`), `dev-c4s2-4` (`.9`–`.14`, `.16`) |
| escalations | none |

## Results

| bead | tier | agent | commit | what landed |
|---|---|---|---|---|
| `c4s2.6` | standard | `dev-c4s2-2` | `1bbfb40` | Amended `c4-standards` and the blueprint templates for the diagram rules; the CLI reference documents all ten `c4/*` rules |
| `c4s2.1` | advanced | `dev-c4s2-1` | `8f4d908` | Widened the Profile extension point with an architecture vocabulary; four new registration defects; added `schemaregistry.Registry.Declares` |
| `c4s2.2` | standard | `dev-c4s2-1` | `b29359a` | Registered this framework's architecture vocabulary in the profile, with a schema drift test |
| `c4s2.3` | advanced | `dev-c4s2-1` | `d6c1d6e` | New package `cli/core/c4elements`: element set and node resolution; vocabulary types moved there to break an import cycle; `Corpus.Source` added |
| `c4s2.4` | standard | `dev-c4s2-1` | `374ef8c` | Relationship statements resolve to drawings and group into relationships (reason-major unresolved ordering) |
| `c4s2.5` | advanced | `dev-c4s2-1` | `7501ceb` | `c4/parse`, `c4/undeclared-node`, `c4/ambiguous-node` raised by `specs check` |
| `c4s2.7` | standard | `dev-c4s2-1` | `8ac5138` | The remaining seven relationship and tag rules raised by `specs check` |
| `c4s2.8` | standard | `dev-c4s2-3` | `3bb889b` | Definition of Done verified and the C4 diagrams feature blueprint retired; module-document-core cites ADR-029/030 |
| `c4s2.9` | light | `dev-c4s2-4` | `cb1caaf` | An empty label no longer matches elements with no `entity_name` |
| `c4s2.14` | standard | `dev-c4s2-4` | `19a7545` | One unreadable architecture blueprint degrades per file instead of aborting `specs check` |
| `c4s2.10` | standard | `dev-c4s2-4` | `f3a18b1` | Blueprints with an empty slug no longer collapse into one element |
| `c4s2.11` | standard | `dev-c4s2-4` | `74cd502` | A duplicated node alias is ambiguous, not silently bound to the first node |
| `c4s2.16` | light | `dev-c4s2-4` | `43b05d0` | `scopeOf` no longer scans every reference per member; Path tie-break makes match order deterministic |
| `c4s2.12` | light | `dev-c4s2-4` | `e9ec23c` | `c4/drawing-conflict` compares contract tags case-insensitively |
| `c4s2.13` | standard | `dev-c4s2-4` | `901c491` | `ValidateRegistration` reports a roster tag with no heading and keys a heading-without-tag defect correctly |
| `c4s2.15` | — | — | — | **Closed by design at planning, no commit**: "The C4 blueprint requires diagrams to be read from source at run time and forbids recording them in the index … Reopen with a measurement as a Check-design change." |

### How the review passes went

- **Pass 1** admitted 8 run-introduced findings (`.9`–`.16`); 7 fixed, `.15`
  closed by design by `plan-c4s2` (grooming comment on `.15`). Held out
  two findings by the AC exception.
- **Pass 2** filed 10, admitted 0: eight held by the AC exception or as
  doc-only; two row 2. The review loop ended there (epic `amendment:` comments).
- `plan-c4s2` measured `.14`'s premise before admitting it and found the
  reviewer's "aborts the same way" claim true only for walk-level failures
  (grooming comment on `.14`) — admitted as a defect on that measurement.

## What the run decided

Ordered by how much a later reader needs it. Each marks a question the
feature blueprint or a bead did not answer.

- **Unresolved-endpoint ordering is reason-major** (`c4s2.4#1`). The
  blueprint's "checking in that order and reporting the source endpoint before
  the target" reads two ways; `Rel(t, u)` with an ambiguous source and an
  external target reports `external_endpoint` about the target. Pinned by a
  test, and now stated in `specs/modules/document-core/document-core-module.md`.
- **The C4 vocabulary types live in `cli/core/c4elements`**, aliased from
  `corpus`, because `c4elements` importing `corpus` would be a cycle
  (`c4s2.3#1`, gap). The level type is a Go type, not a strata enum, because
  the generated `C4Level` has no `module` (`c4s2.1#1`).
- **`Corpus.Source(path)` was added** — the blueprint required diagrams read
  from source at run time and the corpus exposed no source accessor
  (`c4s2.3#2`, gap).
- **`schemaregistry.Registry.Declares(type, property)` was added** — XPT
  required a property-declared check and named no mechanism (`c4s2.1#2`, gap).
- **Defect shapes for registration** (`c4s2.1#3`), **element naming, roster
  levels and roster span** (`c4s2.3#3`), **BiRel, duplicate aliases, view
  mapping** (`c4s2.4#2`), **finding constructors and failing the run on an
  unreadable drawing source** (`c4s2.5#2` — later reversed by `.14`), and **the
  scope of each relationship rule** (`c4s2.7#2`) were all unpinned and chosen
  by the implementer.
- **An AC could not hold literally**: "every Container and Component node …
  resolves to exactly one element" — `ContainerDb(corpus, …)` names no element;
  the repository test exempts passive stores and external nodes (`c4s2.3#4`).
- **Retiring the feature blueprint created 9 `reference/retired-target`
  warnings** the bead did not anticipate; links became plain-text slugs
  (`c4s2.8#2`).
- **The CLI reference documents all ten rules as live** from `.6`, before `.7`
  landed (`c4s2.6#1`).
- **`sameTags` tested directly** rather than through a new snapshot case
  (`c4s2.12#2`); **`scopeOf` gained a Path tie-break** (`c4s2.16#1`).

## Scope

- **Excluded:** none. **Moved:** none. **Admitted then dropped:** none —
  `.15` was admitted by review pass 1 and closed by design, which is a result,
  not a drop.
- **Held out of the run — review findings filed to the backlog, not worked:**

| P | why held | title |
|---|---|---|
| 4 | pass 1, AC exception | SystemProperty is registered and validated but no C4 check reads it |
| 4 | pass 1, AC exception | C4 node resolution trims entity_name but compares the label untrimmed |
| 3 | pass 2, AC exception / doc-only | An unreadable architecture blueprint yields c4/undeclared-node on the documents that draw it, and no finding names the unreadable file |
| 4 | pass 2 | An unreadable member sorting first on a duplicated slug suppresses the readable member's entity name |
| 3 | pass 2 | Relationships naming a duplicated alias are silently excluded when every copy resolves to the same element |
| 4 | pass 2 | A duplicated alias whose first node is external or undeclared is excluded with that reason, not ambiguous_node as the module doc states |
| 4 | pass 2 | ValidateRegistration reports only the missing heading for a roster tag that is also unregistered |
| 4 | pass 2 | architectureDefects' doc comment still says the roster defect is keyed by the roster tag |
| 4 | pass 2 | c4elements.Build keeps an error return that nothing can set |
| 4 | pass 2 | sameTags hand-copies c4elements' slugKey normalisation |
| 4 | pass 2, row 2 | c4elements.Build walks in.References twice where one grouped pass would do |
| 4 | pass 2, row 2 | Set.declare looks up the empty key and then discards the result |

Four of the pass-2 holds (`ah2`, `caa`, `5ug`, `n6h`) are defects in code
this run introduced or touched (`.14`, `.11`); the run is complete over its
admitted set, and those remain open against it.

## Every finding, verbatim

Ids are `<bead>#<ordinal>`, the `specs-` prefix dropped; `c4s2#n` is the run
epic's own. Ordinals count `retro:` comments only, by creation time.
`amendment:` comments and `plan-c4s2`'s grooming-decision comments (on `.14`
and `.15`) are not findings.

### `decision` — 13

**`c4s2.1#1`** (dev-c4s2-1)

> retro:decision The four C4 levels are a Go type corpus.ElementLevel with constants system/container/module/component declared in cli/core/corpus, not a generated strata enum. Rejected alternative: add a four-value enum to schemas/*.strata and regenerate specmodel (as QualifierMeaning is), which would have widened this bead into schema generation; the generated C4Level enum has only three values (no module) and describes relationship levels, so it could not be reused.
> evidence: python over cli/core/schema/spec-model.schema.json enums prints "C4Level ['system', 'container', 'component']"; the vocabulary scan (registeredVocabulary in core/corpus/profile_test.go) flags only document types, statuses, Slug-typed properties ['document_slug','modules'] and qualifiers, so the level literals pass TestTheExtensionPointNamesNoRegisteredVocabulary (measured: PASS). Caveat for later beads: the level 'component' shares its spelling with this framework's roster record-block tag, so a future scan that adds record-block tags to the vocabulary would flag LevelComponent.

**`c4s2.1#3`** (dev-c4s2-1)

> retro:decision Defect shapes not pinned by the spec: an undefined element/contract type is reported at profile level (DocumentType empty, Key = the type) and suppresses that type's level, property and roster defects; a level defect has Key = the level value; a roster is checked when either RosterHeading or RosterTag is set, so a heading with no tag reports Key "". A ProviderProperty registered with no ContractType, or properties registered with no element types, are vacuous but not reported, since the spec lists exactly four defects. Also: cli/profile's specDrivenDev returns the zero ArchitectureRules until the bead registering the architecture vocabulary fills it.
> evidence: cli/core/corpus/registration.go architectureDefects; tests in cli/core/corpus/architecture_registration_test.go.

**`c4s2.2#2`** (dev-c4s2-1)

> retro:decision The drift test reads each type's top-level PascalCase definition (e.g. ContainerBlueprint.properties) by its own untyped traversal, following registration_test.go's citationTypedPropertiesOf, rather than the discriminated TypedFrontmatter branch that schemaregistry.Declares selects — so the test and ValidateRegistration read two different shapes of the same schema and cannot agree by construction. The component-profile-registration record block's description in specs/modules/spec-profile/spec-profile-module.md was left unchanged (it does not mention the record-block or architecture registrations), to avoid drifting from any other declaration of that component.
> evidence: cli/profile/architecture_test.go declaredPropertiesOf; schema definitions ContainerBlueprint.properties include entity_name, system, modules, contracts_consumed.

**`c4s2.3#3`** (dev-c4s2-1)

> retro:decision Choices the ELEM delta did not pin: (1) an element declared only by roster record blocks gets level component (a roster lists a container's or module's components) — the vocabulary registers no level for roster children; (2) an element's name is its blueprint's entity-name property when a blueprint declares it, else the first record block's in path/span order — disagreement between declarations is Element Records' check, not resolution's; (3) a roster covers the roster heading line plus its section (markdown Heading.Section excludes the heading line, so a block attached to the roster heading itself would otherwise be missed); (4) the roster body-reference rule uses resolved body references whose span lies in the roster, and the composed-modules rule uses resolved frontmatter references whose origin property is the registered one, rather than re-reading the property's raw values; (5) added c4diagram.FenceTag = "mermaid"; (6) TestTheExtensionPointNamesNoRegisteredVocabulary now scans cli/core/c4elements as well as cli/core/corpus.
> evidence: cli/core/c4elements/c4elements.go (finishElements, readMembers, scopeOf); cli/core/markdown/headings.go Heading.Section: 'from the end of the heading line'.

**`c4s2.3#4`** (dev-c4s2-1)

> retro:decision AC 'every Container and Component node on its four C4 diagrams resolves to exactly one element' cannot hold literally: spec-cli-container.md draws ContainerDb(corpus, "Spec Corpus", ...), which names no element and is a passive store by the delta's own rule. The repository test (cli/profile/architecture_test.go TestEveryNodeOnThisRepositorysDiagramsResolvesToOneElement) exempts passive stores and external nodes; it reads 4 diagrams and resolves 18 nodes, each to exactly one element.
> evidence: go test ./profile/ -run TestEveryNodeOnThisRepositorysDiagramsResolvesToOneElement -v → 'resolved 18 nodes over 4 diagrams', PASS.

**`c4s2.4#1`** (dev-c4s2-1)

> retro:decision The DRAW delta says 'checking in that order and reporting the source endpoint before the target', which reads two ways: reason-major (each reason checked on source then target) or endpoint-major (every reason on the source, then the target). Implemented reason-major, so Rel(t, u) with an ambiguous source and an external target is external_endpoint about the target; Unresolved.SourceEndpoint records which end. Endpoint-major would have made it ambiguous_node about the source. A test pins the choice (TestUnresolvedStatementsCarryOneReasonInOrder, case 'Rel(t, u)').
> evidence: specs/features/spec-c4-diagrams/spec-c4-diagrams-blueprint.md 'C4 Elements ... (modified — drawings)': 'the component SHALL return the statement unresolved with that one reason, checking in that order and reporting the source endpoint before the target'. Mutation to endpoint-major failed: 'unresolved 5 (t > u) = ambiguous_node source=true, want external_endpoint source=false'.

**`c4s2.4#2`** (dev-c4s2-1)

> retro:decision Unpinned details: a BiRel is neither a drawing nor an Unresolved entry (C4ExclusionReason has no value for it; c4/bidirectional reports it from the statement); where a diagram defines one alias twice, the first node wins for endpoint lookup (c4/duplicate-edge and node checks own duplicates); a statement's view is mapped from the diagram type (C4Context/C4Container/C4Component → context/container/component); relationships are exposed via Set.Relationships() ordered by source slug, target slug, label.
> evidence: cli/core/c4elements/c4elements.go diagramsOf, views, group.

**`c4s2.5#2`** (dev-c4s2-1)

> retro:decision Added finding constructors finding.C4Parse, C4UnreadDiagram (rule c4/parse, severity unverified, detail C4UnreadDiagram), C4UndeclaredNode and C4AmbiguousNode (detail C4Node; Matches [] for undeclared, so the JSON carries an empty list not null). Corpus.check builds the element set for the whole corpus even under CheckFor and narrows afterwards, because scope crosses files; an error reading a drawing document's source fails the check run (returned as an error) rather than becoming a finding. Rejected: a per-document unverified finding for an unreadable source — no rule identifier exists for it.
> evidence: cli/core/corpus/check.go 'corpus: reading C4 diagrams'; cli/core/finding/finding.go C4* constructors.

**`c4s2.6#1`** (dev-c4s2-2)

> retro:decision spec-cli-reference now documents all ten c4/* rules as live although only some are raised until the relationship-check bead lands; I stated the rules as the blueprint specifies rather than marking unimplemented ones, since the sprint lands them together.
> evidence: specs/guides/spec-cli-reference.md 'C4 diagram checks'; rejected alternative: a per-rule 'not yet raised' marker, which would need removal in the next bead.

**`c4s2.7#2`** (dev-c4s2-1)

> retro:decision Unpinned details: (1) duplicate-edge ignores BiRel statements, which c4/bidirectional already reports; (2) interaction-conflict, contract-unresolved and tag-separator apply to every relationship statement including BiRel and those with unresolved endpoints, since each is judged from the statement alone; (3) contract-mismatch raises one finding per contract tag naming both failures when both hold; (4) tag equality for drawing-conflict compares parsed reserved and unreserved tag lists in order (so 'async+audited' and 'audited+async' conflict); (5) consumed contracts and a contract's providers are read from the index's frontmatter references (origin property = the registered property), so only the blueprint's own frontmatter counts and a record-only element consumes nothing; (6) a vocabulary that registers no contract type makes every contract: tag unresolved. c4elements gained Element.ConsumedContracts/Consumes, Contract/ProvidedBy and Set.Contract.
> evidence: cli/core/corpus/c4.go statementFindings, contractMismatchFindings, sameTags; cli/core/c4elements/c4elements.go contractsOf.

**`c4s2.8#2`** (dev-c4s2-3)

> retro:decision retiring the feature blueprint made 9 reference/retired-target warnings (TestProfileSelfApplicationAgainstThisRepositorysCorpus failed: 'Check returned 9 findings, want exactly 0')
> evidence: all in specs/features/spec-c4/spec-c4-blueprint.md citing [[feature-spec-c4-diagrams]]. Followed document-relationship-standards 'Nothing Durable Points at a Feature Blueprint' by replacing the links with plain-text slug mentions; rejected: leaving warnings (breaks zero-finding test). The bead did not anticipate this edit.

**`c4s2.12#2`** (dev-c4s2-4)

> retro:decision tested sameTags directly (internal unit test) rather than through specs check
> evidence: AC offered corpus or cli/cmd tests; a cmd test needs a new snapshot case (fixture corpus + expected log) for a one-function change. Rejected: new snapshot case, cost out of proportion; sameTags is the sole comparison point.

**`c4s2.16#1`** (dev-c4s2-4)

> retro:decision also added a Path tie-break to scopeOf's final sort
> evidence: elements with equal (empty) slugs from the empty-slug fix are ordered by map iteration in scopeOf (sort.Slice on Slug only); tie-break on Path makes Matches order deterministic. Alternative rejected: leave it, which keeps ambiguous_node match lists nondeterministic for slugless members.

### `gap` — 3

**`c4s2.1#2`** (dev-c4s2-1)

> retro:gap The XPT delta requires ValidateRegistration to check that the union schema declares a property on a type, but names no host mechanism for asking that, and corpus may not know where a schema keeps a type's properties. Decided: added schemaregistry.Registry.Declares(type, property), which selects the branch exactly as Validate does and follows $ref and allOf but not anyOf/oneOf. Rejected: re-parsing the schema JSON in registration.go, which would duplicate the branch-selection rule.
> evidence: blueprint specs/features/spec-c4-diagrams/spec-c4-diagrams-blueprint.md section 'Profile extension point in [[module-document-core]] (modified — C4)': 'a registered property that the union schema does not declare on a type it applies to'; anchors list only cli/core/corpus/profile.go and registration.go.

**`c4s2.3#1`** (dev-c4s2-1)

> retro:gap The ELEM delta puts C4 Elements in a new cli/core package that the corpus's checks (CHKN/CHKR, corpus/check.go) will call, but its input — the architecture vocabulary — was declared in cli/core/corpus by the Profile-vocabulary bead, so the new package importing corpus would make that call a cycle. Decided: moved ArchitectureRules, ElementTypeRules, ElementLevel and the level constants into cli/core/c4elements and made corpus name them as type aliases (corpus.ArchitectureRules = c4elements.ArchitectureRules), so Profile's signature and every implementation are unchanged; c4elements takes the index entries, references, a Source func and a Decode func (corpus.DecodeRecordBlock) as data, and corpus.Corpus.C4Elements() wires them. Rejected: (a) a parallel Vocabulary struct in c4elements with a conversion in corpus — two definitions of one registration; (b) putting the component inside package corpus — contradicts 'a new package under cli/core'.
> evidence: blueprint 'Corpus in [[module-document-core]] (modified — parse and node checks)' anchors the checks in cli/core/corpus/check.go; go build fails with an import cycle if c4elements imports corpus while corpus imports c4elements.

**`c4s2.3#2`** (dev-c4s2-1)

> retro:gap Diagrams must be 'read from the blueprint's source at run time', but Corpus exposed no way to read a member's source (only the parsed frontmatter Document). Added Corpus.Source(path), which reads through the same os.Root containment and line-ending normalization as the walk (readDocument now calls it), so index spans address the bytes returned.
> evidence: grep of cli/core/corpus exported methods before this bead: Entry, Document, Slug, OutboundReferences, InboundReferences — no source accessor; readUnder normalizes CRLF.

### `scope` — 0

None recorded.

### `sequencing` — 0

None recorded.

### `stale` — 2

**`c4s2.1#5`** (dev-c4s2-1)

> retro:stale specs/modules/document-core/document-core-module.md said 'The defects are five' while ValidateRegistration already found seven (the two record-block defects were described only in a later paragraph). Rewrote the list as all eleven.
> evidence: registration.go doc comment before this bead: 'Seven defects are found'.

**`c4s2.8#1`** (dev-c4s2-3)

> retro:stale module-document-core Key Decisions cited neither ADR
> evidence: awk over '## Key Decisions' for 'adr-029\|adr-030' returned nothing before this bead although the blueprint's DoD requires both; added two entries.

### `unverifiable` — 1

**`c4s2.8#3`** (dev-c4s2-3)

> retro:unverifiable AC 'retro-spec has run over the feature retro comments' not satisfied here; one DoD item naming a bead unverifiable (bd show: 'Issue not found', retired prefix). All 17 AC-C4-DIAG-001.1..004.5 verified cited (c4_check_test.go, c4diagram_test.go, snapshot cites); scripts/coverage.sh --check exit 0, README already current. The sprint retro (swarm-retro) is the natural home for the aggregated retro-spec pass.

### `tooling` — 7

**`c4s2#1`** (coordinator-c4s2)

> retro:tooling:swarm — `swarm dispatch accept` fast-forwarded the coordinator's CHECKED-OUT branch (claude/project-thread-w4yqh2) instead of the sprint branch when the coordinator checkout was not on the sprint branch; reported merged_sha=base, ancestor=false, exit 2, left worktree standing. `swarm dispatch open` cuts sprint/<epic> without checking it out. Repaired by `git branch -f <sprint branch> 1bbfb40`, re-running accept (exit 0), then checking out the sprint branch in the main checkout.

**`c4s2#2`** (coordinator-c4s2)

> retro:tooling:swarm — bb-close-bead.sh --trailers-from <epic> exits 2 when the manifest has no `trailers` key (dev-c4s2-2).

**`c4s2#3`** (coordinator-c4s2)

> retro:tooling:swarm — `swarm dispatch accept` (concurrent case) runs `git rebase <base> bb/<dispatch>` from the main checkout, which git refuses (exit 128, "already used by worktree") because the dispatch branch is checked out in its worktree. Worked around by `git -C <worktree> rebase <sprint-head>`, re-running bb-checks.py (exit 0), then re-running accept.

**`c4s2#4`** (coordinator-c4s2)

> retro:tooling:swarm — bb-close-bead.sh pushes the tracker on every close; the run's "hold the tracker push" note could not be honoured, and the c4elements bead's push printed a failure warning.

**`c4s2#5`** (coordinator-c4s2)

> retro:tooling:swarm-scope — scope recorded standing_auth but no `admitted` list; `swarm status` exited 1 at end of run until the coordinator wrote it.

**`c4s2#6`** (coordinator-c4s2)

> retro:tooling:swarm-dispatch — dev-c4s2-4 returned batch-end after its first bead while its queue held a bead that bead`s own close had released (collision chain .9 → .14 → …); the coordinator had to resume it. The dispatch skill should re-query its queue after each close until `swarm ready --agent` reports drained.

**`c4s2.7#3`** (dev-c4s2-1)

> retro:tooling runInFixture (cli/cmd/query_test.go) chdirs into the fixture and restores only at test cleanup, so the relative casesDir ('../testdata/cases') no longer resolves for a second run inside the same test; the first version of TestDrawingsThatDisagreeAreAWarningOnTheSecond failed with 'lstat ../testdata/cases/check-c4-relationship-defects/input: no such file or directory'. Fixed locally by resolving the cases directory from runtime.Caller in c4CheckFixture.
> evidence: go test ./cmd output quoted above.

### `proof` — 13

**`c4s2.1#4`** (dev-c4s2-1)

> retro:proof Negative control: with ValidateRegistration's 'return append(defects, architectureDefects(profile)...)' replaced by 'return defects', go test ./core/corpus/ -run 'Architecture|ALevel|RosterTag|Applies|UnparsableSchema' failed 6 tests (each of the four defect tests, the applies-only test and the unparsable-schema test); restored, all pass and bb-checks.py exits 0. scripts/coverage.sh was run (go1.24.7, go.mod go 1.24) and its README.md figures are in this commit.
> evidence: '--- FAIL: TestValidateRegistrationReportsAnUndefinedArchitectureType' ... '--- FAIL: TestValidateRegistrationRaisesNoPropertyDefectForAnUnparsableSchema'; bb-checks exit=0.

**`c4s2.2#1`** (dev-c4s2-1)

> retro:proof Negative control for the drift test: with SystemProperty changed to "system_of" in cli/profile/profile.go, go test ./profile/ failed three tests — TestTheProfileRegistersTheArchitectureVocabulary, TestEveryRegisteredArchitecturePropertyIsDeclaredWhereItApplies ('system property "system_of" is not declared on "container-blueprint" in the generated schema') and TestTheRegistrationHasNoDefects (ValidateRegistration reported {DocumentType:container-blueprint Key:system_of ...}), which also exercises the Profile-vocabulary bead's property defect against the real embedded schema. Restored, all pass; bb-checks.py exit 0; scripts/coverage.sh run and README figures committed.
> evidence: go test ./profile/ → neg=1 with the three FAIL lines above; restored=0.

**`c4s2.3#5`** (dev-c4s2-1)

> retro:proof Negative controls on the unit tests (cli/core/c4elements/c4elements_test.go): resolving nodes against every element instead of the drawing document's scope failed TestANodeResolvesWithinItsDrawingDocumentsScope ("alpha's Parser = component-alpha-parser,component-beta-parser"); matching the roster heading on raw text instead of markdown.Identifier failed TestABlueprintAndARecordBlockMergeUnderOneSlug and TestTheRosterHeadingIsMatchedWithoutRegardToCase. Restored, all pass; bb-checks.py exit 0; coverage.sh run and README figures committed. The fixtures use this framework's document types because IndexEntry.Frontmatter decodes only the generated union; the vocabulary itself comes from a stub profile (core may not import cli/profile).
> evidence: go test ./core/c4elements/ outputs quoted above.

**`c4s2.4#3`** (dev-c4s2-1)

> retro:proof Negative controls: dropping the BiRel filter failed TestRelationshipStatementsResolveToDrawings, TestBiRelResolvesToNoDrawing and TestDrawingsGroupIntoRelationshipsAcrossDocuments; the endpoint-major mutation above failed TestUnresolvedStatementsCarryOneReasonInOrder. Restored, all pass; bb-checks.py exit 0; coverage.sh run and README figures committed.
> evidence: go test ./core/c4elements/ FAIL lines as listed; restored 'ok'.

**`c4s2.5#1`** (dev-c4s2-1)

> retro:proof Integration proof: new snapshot case cli/testdata/cases/check-c4-node-defects (cites AC-C4-DIAG-001.1/.3/.4, 002.1-.5; added to TestSnapshotCases' AC line) and cli/cmd/c4_check_test.go (per-AC tests with AC: citations, plus --only selecting each of the three rules alone). Expected output: 'c4/ambiguous-node error 22:5 ... names 2 elements in scope: container-twin-one, container-twin-two', 'c4/undeclared-node error 23:5', 'c4/parse error 24:5 cannot read "Frobnicate(...)"', 'c4/parse unverified 46:1 diagram type "C4Dynamic"'; the fixture's feature-blueprint diagram, Person, System_Ext, ContainerDb and the two scoped Parsers raise nothing. Snapshot regeneration changed 21 existing expected files; every diff line was reviewed and is only the three new check rows (15x passed, 1x not-requested each) and the --only known-rule list. Self-application and build/specs check --strict over this repository report the three rules passed with no c4 finding.
> evidence: git diff -- cli/testdata | grep '^[-+]' | sort | uniq -c → only '+  c4/... passed', '+  c4/... not-requested', and the changed JSON/err lines carrying the same three rules; bb-checks.py exit 0 with 'c4/parse  passed' etc.

**`c4s2.7#1`** (dev-c4s2-1)

> retro:proof New snapshot cases cli/testdata/cases/check-c4-relationship-defects (cites 003.1-003.3, 004.2-004.5) and check-c4-drawing-conflict-path (single-document drawing-conflict, cites 003.3), plus per-AC tests with AC: citations and an --only test over all ten c4 rules in cli/cmd/c4_check_test.go. Output: bidirectional 24:5, duplicate-edge 26:5 (Rel_Back(b, a) after Rel(a, b); Rel(b, a) not reported), interaction-conflict 28:5, contract-unresolved 29:5, contract-mismatch warning 30:5 (both reasons), tag-separator warning 31:5, drawing-conflict warning at beta.md:20 naming alpha.md:25 — also under 'check docs/containers/beta.md'. Snapshot regeneration touched 22 existing files; a script compared every changed JSON report (findings identical, only 7 checks added) and every text diff line is a new c4 row (16x passed, 1x not-requested) or the --only known-list. build/specs check --strict over this repository: all ten c4 rules passed, no findings.
> evidence: git diff -- cli/testdata review output: 'findings same: True added: [passed] 7 changed: {}' x2, '[not-requested] 7' x2; bb-checks.py exit 0.

**`c4s2.9#1`** (dev-c4s2-4)

> retro:proof reproduced before fix: new test failed on unfixed resolve
> evidence: go test ./core/c4elements (guard stashed) -> 'node x matched [container-nameless-one], want none'; two nameless -> unresolved reason ambiguous_node; with guard -> ok

**`c4s2.10#1`** (dev-c4s2-4)

> retro:proof pre-fix failure observed
> evidence: go test ./core/c4elements -run NoSlug (fix stashed) -> 'run 0: slugless elements = [0xc000415180], want Alpha at a.md then Beta at b.md' (two members collapsed to one element); with fix passes

**`c4s2.11#1`** (dev-c4s2-4)

> retro:proof pre-fix failure observed
> evidence: go test ./core/c4elements -run Duplicated (fix stashed) -> 'drawings = [{... SourceAlias:web TargetAlias:api ...}], want none' and 'unresolved = [], want {ambiguous_node ...}' for target and source duplicated; with fix passes

**`c4s2.12#1`** (dev-c4s2-4)

> retro:proof pre-fix failure observed
> evidence: go test ./core/corpus -run SameTags (fix stashed) -> 'contract slugs differing in case: sameTags = false, want true'; with fix passes

**`c4s2.13#1`** (dev-c4s2-4)

> retro:proof pre-fix failures observed
> evidence: go test ./core/corpus -run 'RosterTagWithNoHeading|RosterHeadingWithNoTag' (fix stashed) -> 'defects = [], want one for wing keyed room' and 'defects = [{wing Key: Detail:is registered as the record-block tag ... "Components" ...}]' (empty Key); with fix both pass

**`c4s2.14#1`** (dev-c4s2-4)

> retro:proof pre-fix failure observed
> evidence: go test ./core/corpus -run Unreadable (readMembers fix stashed) -> 'Check returned corpus: reading C4 diagrams: c4elements: reading specs/beta.md: openat beta.md: no such file or directory, want no error'; with fix passes

**`c4s2.16#2`** (dev-c4s2-4)

> retro:proof pure refactor, no behaviour change to prove by a failing test
> evidence: go test ./core/c4elements and full bb-checks.py pass unchanged; no pre-fix failure applicable

### `environment` — 1

**`c4s2.5#3`** (dev-c4s2-1)

> retro:environment When a check run raises both an error and an unverified finding under one rule (c4/parse here, as reference/unresolved-anchor can), Report.Add sets the outcome to unverified on the later unverified finding but Model still reports the rule 'failed' because checkState.failed is sticky — observed: the fixture reports 'c4/parse  failed' with one error then one unverified finding. Correct for this rule; noted because Report.Add's comment reads as if order decides.
> evidence: cli/testdata/cases/check-c4-node-defects/expected/checks.log 'c4/parse  failed'; finding.go Report.Add sets state.outcome = CheckOutcomeUnverified on an unverified finding.

### kind missing (UNROUTED) — 0

None. Every `retro:` comment in scope carries a kind.

## Routing

| branch | count | findings |
|---|---|---|
| project (`decision`, `gap`, `stale`, `unverifiable`) | 19 | 13 `decision`, 3 `gap`, 2 `stale`, 1 `unverifiable`, as listed above |
| skill (`tooling`) | 7 | `c4s2#1`–`#4` (`swarm`), `c4s2#5` (`swarm-scope`), `c4s2#6` (`swarm-dispatch`) — all `swarm` plugin, so `outbox/swarm/`; and `c4s2.7#3`, which the merge test sends to the **project** (the fix lands in this repository's `cli/cmd/query_test.go`, not in any skill) |
| `proof` → this document | 13 | `c4s2.1#4`, `.2#1`, `.3#5`, `.4#3`, `.5#1`, `.7#1`, `.9#1`, `.10#1`, `.11#1`, `.12#1`, `.13#1`, `.14#1`, `.16#2` |
| `environment` → ledger | 1 | `c4s2.5#3` — **not written to the ledger by this retro** (instructed to commit the record and `log.md` only). By the merge test it is a project fact, not an environmental one: the fix is a comment in this repository's `cli/core/finding/finding.go` (`Report.Add`) |
| **unrouted** | **0** | — |

Unrouted is 0 of 40, against 9 of 29 on the C4 diagrams stage 1 run. The planner wrote no
`retro:` findings this run — its two grooming decisions were plain comments —
so the drop says nothing yet about whether `swarm-plan` now sets a kind.

## Validation

- **Baseline** (`manifest.checks_baseline`, untouched tree at `3e612b9`): all
  12 declared checks passed, exit 0.
- **At head**: every bead's close ran `bb-checks.py` per its findings
  (exit 0 cited in `.1#4`, `.2#1`, `.3#5`, `.4#3`, `.5#1`, `.7#1`, `.16#2`;
  `.1`–`.7` also ran `scripts/coverage.sh` and committed README figures; `.8`
  ran `coverage.sh --check`, exit 0). **Not re-run by this retro**, and no
  head-of-branch run is recorded on the epic.
- **CI was not observed**: the branch has not been pushed.

## Pushes

The sprint branch has **no remote counterpart**.
`git log --oneline origin/claude/project-thread-w4yqh2..HEAD` lists **15**
commits (that remote branch is at `3e612b9`, the base) — every commit of the
run is stranded locally until the coordinator pushes, which it does after this
retro. **Tracker:** closures were pushed by `bb-close-bead.sh` on each close,
against the standing note to hold them (`c4s2#4`); `.3`'s push printed a
failure warning, whose outcome this retro did not check.

## What this record could NOT verify

- **Which model served each dispatch** — `manifest.models` and
  `manifest.effort` are both unset; the lineup in "The run" is the
  coordinator's report.
- **Head-of-branch check results** — not on the epic, not re-run here.
- **CI** — branch unpushed.
- **Whether `.3`'s failed tracker push was later covered** by a subsequent
  close's push — the warning is reported in `c4s2#4`; no measurement.
- **The feature blueprint's DoD item "retro-spec has run over the feature
  retro comments"** — not done by `.8` (`c4s2.8#3`) and not done by this
  retro (swarm-retro is not retro-spec). Candidate G.
- **The DoD item naming a retired bead as closed** — the id is retired and
  `bd show` cannot resolve it (`c4s2.8#3`); the blueprint (now retired) still
  checks it as `[x]` at `specs/features/spec-c4-diagrams/spec-c4-diagrams-blueprint.md`.
- **`plan-c4s2`'s `.15` decision** rests on reading, and its cost claim is
  unmeasured by design ("Reopen with a measured cost").

## Candidates

**PENDING — put to Brett by the coordinator; no decision recorded.**

Ranked by how much each changes the next run. "Fold" means the issue already
exists in `outbox/swarm/`; the action is to attach this run's evidence and
raise its recurrence rather than file anew.

| # | candidate | audience → destination | findings | why it changes the next run | decision |
|---|---|---|---|---|---|
| A | **`bb-close-bead.sh` honours a run's hold on tracker pushes** — the close pushes the tracker only when the run's standing authorization permits it, and otherwise leaves the closure local for the coordinator's final push | skill → `outbox/swarm/swarm/<date>-close-pushes-the-tracker-despite-a-hold.md` (new) | `c4s2#4` | A human's explicit "hold the push" is silently overridden on every close; nothing the coordinator can do in-run honours it | pending |
| B | **`swarm-scope` records the `admitted` list with `standing_auth`** | skill → `outbox/swarm/swarm-scope/<date>-scope-records-no-admitted-list.md` (new) | `c4s2#5` | `swarm status` exits 1 at end of run until the coordinator hand-writes the manifest key | pending |
| C | **Fold: accept merges into the recorded sprint branch, not the checked-out one** — recurrence 1; add that `dispatch open` cuts `sprint/<epic>` without checking it out, which is what sets the trap | skill → existing `outbox/swarm/swarm/2026-10-07-accept-merges-into-the-current-branch.md` | `c4s2#1` | Second run in a row: first accept exits 2 and needs `git branch -f` | pending |
| D | **Fold: accept rebases inside the worktree that holds the branch** — recurrence 1 | skill → existing `outbox/swarm/swarm/2026-10-07-accept-rebases-a-branch-held-by-a-worktree.md` | `c4s2#3` | Second run in a row: every concurrent accept after the first fails exit 128 | pending |
| E | **Fold: a run resolves no `trailers`** — new evidence: `--trailers-from` is not merely unusable but fatal (`bb-close-bead.sh` exits 2 when the key is absent), so the issue's proposed "always pass `--trailers-from`" breaks every close until scope resolves the key | skill → existing `outbox/swarm/swarm/2026-10-07-a-run-resolves-no-trailers.md` (recurrence 2 → 3) | `c4s2#2` | Another run with no attribution trailers; the existing fix, applied alone, would turn that into a failed close | pending |
| F | **Fold: a dispatch returns only when its queue is drained** — recurrence 1; concrete trigger this time: a bead released by the agent's own close (collision chain `.9 → .14 → …`) | skill → existing `outbox/swarm/swarm-dispatch/2026-10-07-dispatch-returned-with-queue-undrained.md` | `c4s2#6` | Second run (after the C4 prerequisites run) where the coordinator had to resume a dispatch that returned early | pending |
| G | **Run `retro-spec` over the C4 diagrams feature's findings** — the feature blueprint's DoD required it and nobody ran it; 3 `gap` and the unpinned-detail `decision` findings (`.1#3`, `.3#3`, `.4#1`, `.4#2`, `.7#2`) are its input, aimed at the durable homes (`module-document-core`, `c4-standards`) now that the feature blueprint is retired | project → bead | `c4s2.8#3`, `c4s2.1#2`, `c4s2.3#1`, `c4s2.3#2`, `c4s2.1#3`, `c4s2.3#3`, `c4s2.4#1`, `c4s2.4#2`, `c4s2.7#2` | The implementer filled every silent spot by choice; unless those choices reach the durable specs, stage 3 and the held-out `ah2`/`5ug`/`n6h` beads are groomed against documents that still do not say what the code does | pending |
| H | **`runInFixture` leaves the working directory changed for the rest of the test** — resolve `casesDir` absolutely (as `c4CheckFixture` now does locally) or restore the directory after each run | project → bead (P4) | `c4s2.7#3` | The next cmd test that runs two fixtures in one test hits the same `lstat ../testdata/cases/...` failure; the local fix covers only `c4CheckFixture` | pending |

**Kept in this record, not candidates** (do not clear the bar on this run's
evidence): `c4s2.2#2` (drift test and `ValidateRegistration` read two shapes
of the schema — no disagreement observed); `c4s2.1#1`'s caveat that level
`component` shares a spelling with the roster tag; `c4s2.5#3` (`Report.Add`
comment reads as if order decides — a one-line comment fix, cheaper done than
filed); `c4s2.8#2` (blueprint retirement's retired-target warnings —
`document-relationship-standards` already covers it).

## What was filed

**Nothing yet** — decisions pending. No `retro:filed:` comment has been
written. `retro_filed`, `retro_declined` and `retro_routed` are not yet set on
the manifest; `findings_path` is.

## Counts

| | |
|---|---|
| harvested | 40 |
| routed — project | 19 (+1 `tooling` that the merge test sends to the project: `c4s2.7#3`) |
| routed — skill | 7 by kind; 6 after the merge test |
| routed — `proof` (this document) | 13 |
| routed — `environment` (ledger) | 1 (not yet written to the ledger) |
| unrouted | 0 |
| candidates | 8 (2 new skill issues, 4 folds into existing outbox issues, 2 project beads) |
| filed | 0 — pending |
| declined | 0 — pending |
