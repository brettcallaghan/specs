# Run record — `specs-c4model`, C4 model and the `specs c4` command

## Read this first

**Every admitted bead closed, and `specs c4` ships.** 12 admitted, 12 closed,
12 commits (plus one merge commit from dispatch `dev-c4model-2`). Review ran
all three passes the default budget allows; pass 1 admitted 3 run-introduced
bugs (`.9`–`.11`), pass 2 admitted 1 (`.12`, **a regression introduced by
`.10`'s fix**), pass 3 admitted nothing. **11 review findings sit in the
backlog**, one at P2 (`specs-i20`, the `c4/contract-mismatch` check's twin of
the `.12` defect).

**The `.10` → `.12` regression reversed a decision an earlier bead had made
and recorded only as a retro comment** (`specs-c4model.4#1`: read
`contracts_consumed` from the source declaration, because
`Element.ConsumedContracts` holds no record-block values). Candidate F.

**Accept needed hand repair again** — third run in a row (`specs-c4model#3`).

**All candidates are PENDING** — returned to the coordinator to put to a human.
Nothing has been filed and no `retro:filed:` comment written.

## The run

| | |
|---|---|
| run epic | `specs-c4model` — "C4 model: element tree, relationships and the specs c4 command" (feature `feature-spec-c4`). The epic itself is still `open` |
| branch | `sprint/specs-c4model`; also pushed as `claude/swarm-c4model-iqetxu` |
| `base_sha` → head | `0a44644` → `dc2cb1c` — 13 commits before this record's commit (12 bead commits, 1 merge `3079922`) |
| queue | 8 admitted at the gate (`.1`–`.8`); +3 admitted by review pass 1 (`.9`–`.11`); +1 by pass 2 (`.12`); pass 3 admitted 0. 12 / 12 closed |
| declared checks | `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| tiers (bead `metadata.tier`) | advanced: `.2`; standard: `.3`, `.4`, `.5`, `.6`, `.8`, `.9`; light: `.1`, `.7`, `.10`, `.11`, `.12` |
| model lineup (`manifest.models`) | advanced → `opus`, frontier → `fable`, light → `haiku`, standard → `sonnet`. Which model served each *dispatch* is not recorded on the epic |
| effort | `manifest.effort` = `unrecorded` |
| review | `review_budget` 2 (default), `review_passes` 2 → 3 passes, budget spent |
| concurrency cap | 20 |
| dispatches | 5, 0 blocked, 0 skipped; circuit breaker (`consecutive_dispatch_failures`) never above 0 |
| escalations | none |
| out-of-session recovery trigger | not armed — the session was attended |

## Results

| bead | tier | agent | commit | outcome |
|---|---|---|---|---|
| `.1` | light | `dev-c4model-1` | `1f1bb73` | closed — `bidirectional` exclusion reason added to the C4 model schema |
| `.2` | advanced | `dev-c4model-2` | `beced1d` | closed — C4 Model component assembles the element tree |
| `.3` | standard | `dev-c4model-2` | `9154026` | closed — relationships and exclusions |
| `.4` | standard | `dev-c4model-2` | `e86cfd9` | closed — reserved relationship tags, derived contract link |
| `.5` | standard | `dev-c4model-2` | `41126e0` | closed — narrowing to one element's subtree (`--root`) |
| `.6` | standard | `dev-c4model-3` | `19e9384` | closed — `specs c4` subcommand |
| `.7` | light | `dev-c4model-3` | `d882b41` | closed — CLI reference and output standards |
| `.8` | standard | `dev-c4model-3` | `685fd7c` | closed — DoD verified, C4 model blueprint retired |
| `.9` | standard | `dev-c4model-4` | `47b421e` | closed — `--json` no longer fails on a module blueprint missing `kind`/`visibility` |
| `.10` | light | `dev-c4model-4` | `8371d3c` | closed — anchored `contracts_consumed` citation derives the contract |
| `.11` | light | `dev-c4model-4` | `b3d938d` | closed — `modules` list folds case-variant duplicates |
| `.12` | light | `dev-c4model-5` | `dc2cb1c` | closed — record-block `contracts_consumed` derives the contract again (regression from `.10`) |

### How the review passes went

From the epic's `amend:` comments, verbatim:

- "review pass 1 admitted run-introduced specs-c4model.9, .10, .11; backlog: specs-lce, specs-8l2, specs-nra, specs-7zb, specs-3g1. Re-entering groom & plan (review_passes=1 of budget 2)."
- "pass-2 collision edges loaded from /tmp/claude-0/specs-c4model.pass2.edges.jsonl (.10<-.9, .11<-.10); batch dev-c4model-4."
- "review pass 2 admitted run-introduced specs-c4model.12 (regression from .10); backlog: specs-56e, specs-35f, specs-1x7, specs-m3h, specs-l4n. Re-entering groom & plan (review_passes=2 of budget 2; pass 3 is the last)."
- "review pass 3 admitted nothing (1 backlog bead specs-i20, pre-existing check-side gap). Review loop ends with budget spent (2 loops, 3 passes)."

Pass 2's only admission was a defect pass 1's own fix introduced, so one of the
two loops the budget allows was spent repairing the other.

## What the run decided

Findings of kind `decision`, `scope`, `gap` and `sequencing` — each a place a
document did not answer a question a bead needed.

- **`specs-c4model.2#1`** — C4 Model reads nesting from C4 Elements (new `System`, `ComposedModules`, `Roster`, `Set.Contracts()`) rather than re-deriving rosters.
- **`specs-c4model.2#2`** — non-identity properties copied by JSON field name via reflection; wrong-typed YAML values left absent, not coerced; slugless elements left out.
- **`specs-c4model.2#3`** — `Corpus.C4Model(workspace)` added as the Document Core entry point; `.5` extended it (`specs-c4model.5#2`, `scope`).
- **`specs-c4model.3#2`** — `BiRel` exclusions read directly from diagram statements; C4 Elements left unchanged so the `c4` checks do not see `BiRel` as an endpoint failure.
- **`specs-c4model.4#1`** — the derived contract reads `contracts_consumed` from the source's declaration (blueprint, else first record block), **not** `Element.ConsumedContracts`, "which holds only blueprint references and so would never derive a contract for a record-only component". **`.10`'s fix switched to `source.ConsumedContracts` (`specs-c4model.10#1`) and `.12` restored record-block derivation.**
- **`specs-c4model.5#1`** — "each ancestor holding no other child" read as no other component, container or composed module; a narrowed ancestor shows `modules: []`.
- **`specs-c4model.5#2`** (`scope`) — entry-point signature changed to `C4Model(workspace, root)` with `*c4model.UnknownRootError`.
- **`specs-c4model.1#1`** (`gap`) — the ACs did not say that editing `spec-model.schema.json` changes the `input_hash` in six index snapshot manifests; `go test ./...` failed until `scripts/regenerate-cli-snapshots.sh` ran.

`sequencing`: none recorded.

## Scope

- **excluded:** none (`manifest.excluded` = `[]`).
- **moved:** none (`manifest.moved` = `[]`).
- **admitted then dropped:** none.
- **Held out to the backlog by review** (not admitted, by the default
  `admit_review_findings` rule for out-of-scope findings): pass 1 —
  `specs-lce` (P3, `--root` under a module omits the container/system chain the
  CLI reference promises), `specs-8l2` (P4), `specs-nra` (P4), `specs-7zb`
  (P4), `specs-3g1` (P4); pass 2 — `specs-56e`, `specs-35f`, `specs-1x7`,
  `specs-m3h`, `specs-l4n` (all P4); pass 3 — `specs-i20` (P2,
  `c4/contract-mismatch` ignores a record-block element's
  `contracts_consumed`). The run finished its admitted set; it did **not**
  narrow scope.

## Every finding, verbatim

**23 records across 13 issues** (the epic and `.1`–`.12`; `.7` and `.8` carry
only kind-less "clean bead" markers). Harvested with one query over the run's
closure set:

```bash
bd export | jq -r 'select(.id=="specs-c4model" or (.id|startswith("specs-c4model.")))
  | .comments[]? | select(.text|startswith("retro:")) | .text'
```

### `decision` — 6

> **`specs-c4model.2#1`** retro:decision Exposed what C4 Model nests by from C4 Elements rather than recomputing it: Element gained System and ComposedModules (read from the same frontmatter-property references as ConsumedContracts), Document gained Roster []Child (record blocks plus roster-heading references to component blueprints, with spans), and Set gained Contracts(). Rejected: re-deriving rosters/refs inside cli/core/c4model, which would have duplicated the roster judgement C4 Elements owns.
> evidence: cli/core/c4elements/c4elements.go "func rosterOf(" and "func (s *Set) Contracts() []*Contract"; bead metadata c4_impacts: ["component-c4-elements"].

> **`specs-c4model.2#2`** retro:decision Properties other than entity_name/contracts_consumed (description, technology, tags, kind, visibility, version) are copied by the model fields JSON names via struct tags + reflection, not string literals, so cli/core/c4model passes the vocabulary source scan, which I extended to cover the package. A value of the wrong YAML type (e.g. unquoted version: 0.3 read as a float) is left absent rather than coerced. Elements with an empty slug are left out of the model (schema needs a slug).
> evidence: cli/core/corpus/profile_test.go "goFilesUnder(t, filepath.Join(\"..\", \"c4model\"))"; cli/core/c4model/c4model.go "func copyProperties(".

> **`specs-c4model.2#3`** retro:decision Added Corpus.C4Model(workspace string) as the Document Core entry point (the CMD bead wants one Document Core call); relationships and excluded_relationships are emitted as empty lists until the REL bead fills them. The ROOT bead will need to extend this signature with a root slug.
> evidence: cli/core/corpus/c4.go "func (c *Corpus) C4Model(workspace string) (*specmodel.C4Model, error)".

> **`specs-c4model.3#2`** retro:decision BiRel exclusions are read by the C4 Model package directly from each drawing document diagram statements (Kind relationship, Relationship.Bidirectional), since C4 Elements deliberately leaves BiRel out of both Drawings and Unresolved; I did not change C4 Elements to report it as Unresolved, which would have made the c4 checks see it as an endpoint failure. A relationship whose endpoint element has no slug is left out, matching the tree.
> evidence: cli/core/c4elements/c4elements.go "A `BiRel` statement is in neither"; cli/core/c4model/c4model.go "func (a *assembly) excluded()".

> **`specs-c4model.4#1`** retro:decision The derived contract reads the source element contracts_consumed from its source declaration (blueprint frontmatter, else the first record block), the same value the model emits, rather than c4elements Element.ConsumedContracts, which holds only blueprint references and so would never derive a contract for a record-only component. A contract listed twice counts once; a contract: tag with an empty slug is skipped as if absent (well-formedness is the c4 checks).
> evidence: cli/core/c4model/c4model.go "func (a *assembly) contractOf("; cli/core/c4elements/c4elements.go "A record block declares none."

> **`specs-c4model.5#1`** retro:decision "Each [ancestor] holding no other child" read as: no other component, container or composed module. A component root under container-alpha yields container-alpha with modules: [] even though the container composes module-gamma. Rejected keeping the ancestor modules list, which would name modules absent from the narrowed model. "The narrowed set" for relationship/exclusion filtering is the root plus descendants, not the ancestor chain.
> evidence: cli/core/c4model/narrow.go "container.Components, container.Modules = components, []specmodel.Slug{}"; narrow_test.go case component-parser. Negative control: keeping the ancestor modules fails that case.

### `gap` — 1

> **`specs-c4model.1#1`** retro:gap:cli-snapshots The bead's ACs do not mention that editing spec-model.schema.json changes the input_hash embedded in six index snapshot manifests; go test ./... failed (TestSnapshotCases, TestCRLFCheckoutProducesIdenticalOutput, TestBinaryWritesTheIndexToStandardOutput) until scripts/regenerate-cli-snapshots.sh ran. Evidence: go test output showed input_hash sha256:98b011cc... expected vs sha256:07e70149... produced; the regenerated diff is six one-line input_hash changes only.

### `scope` — 1

> **`specs-c4model.5#2`** retro:scope Changed the Document Core entry point signature from Corpus.C4Model(workspace) (added by specs-c4model.2) to Corpus.C4Model(workspace, root string); an unknown root returns *c4model.UnknownRootError ("<slug> names no element") and a nil model. The CMD bead (specs-c4model.6) should call this and print the error on stderr.
> evidence: cli/core/corpus/c4.go "func (c *Corpus) C4Model(workspace, root string)".

### `sequencing` — 0

### `stale` — 2

> **`specs-c4model.2#4`** retro:stale document-core-module.md said "Ten of the eleven are built" while its diagram already listed twelve components including the built C4 Elements; corrected to "Eleven of the thirteen" with C4 Elements and C4 Model named.
> evidence: git diff specs/modules/document-core/document-core-module.md, the "Components" paragraph after the diagram.

> **`specs-c4model.9#1`** retro:stale An invalid kind/visibility value (e.g. visibility: bogus) never reaches the module loop: the blueprint fails frontmatter schema validation and the module is absent from the model on the pre-fix binary too (exit 0, modules: []). Only a MISSING kind/visibility reproduced the failure.
> evidence: cp c4-json input, set visibility: bogus, kind: standard; pre-fix binary: specs c4 --json rc=0, top-level modules []. Added e2e cases for both and unit tests for both.

### `unverifiable` — 0

(`specs-c4model.11#1`, filed as `proof`, carries one unverified sub-claim; see
"What this record could NOT verify".)

### `tooling` — 1

> **`specs-c4model#3`** retro:tooling:swarm swarm dispatch accept fast-forwarded the dispatch branch into the CURRENTLY CHECKED-OUT branch (claude/swarm-c4model-iqetxu) rather than the reported target sprint/specs-c4model, then reported ancestor=false and exit 2. Measured: after accept, `git worktree list` showed the main checkout at 1f1bb73 and `git log -1 sprint/specs-c4model` at 0a44644. Repaired by `git branch -f sprint/specs-c4model 1f1bb73` and checking out the sprint branch. The command should check out the target or refuse when HEAD is not on it.

### `proof` — 8

> **`specs-c4model.2#5`** retro:proof Negative control: making a later roster declaration overwrite the parent (sed "if _, placed := parents[element]; !placed {" -> "if true {") fails TestComponentsNestUnderTheParentWhoseRosterDeclaresThem ("component-shared placed under \"module-gamma\", want \"container-beta\"") and TestRecordBlocksInTwoDocumentsYieldOneElement; restored, all pass. Over this repo: container-spec-cli is in orphaned_containers with modules [module-document-core, module-spec-profile], module-document-core.contained_by = [container-spec-cli] (TestThisRepositorysC4ModelNestsItsContainerAndModules in cli/profile, since cli/core may not import the profile). scripts/coverage.sh not run: the DoD bead specs-c4model.8 owns it.

> **`specs-c4model.3#1`** retro:proof C4Dynamic/C4Deployment exclusion needed no code: c4diagram.Parse returns no statements for a type Read() rejects ("Statements ... Always empty unless [Diagram.Read]"), so c4elements yields no drawings or unresolved from them; the tests prove it end to end. Negative controls: making present() keep empty strings fails TestRelArgumentsBecomeRelationshipProperties ("technology = , want <absent>"); suppressing BiRel handling fails TestEdgesTheModelDoesNotAdmitAreExcluded. Restored, all pass.
> evidence: cli/core/c4diagram/c4diagram.go "Always empty unless [Diagram.Read]"; cli/core/c4model/relationships_test.go.

> **`specs-c4model.4#2`** retro:proof Negative control: relaxing "if len(fits) != 1" to "== 0" fails TestNoContractIsDerivedWhenMoreThanOneFits ("contract = &{Slug:contract-one Source:derived}, want none"); restored, passes. AC-C4-003.1 also tests sync+async and an unreserved-only tag reading as absent.

> **`specs-c4model.6#1`** retro:proof validated build/specs c4 --json over this repo and the c4-default fixture against definitions/C4Model of cli/core/schema/spec-model.schema.json with python jsonschema: both valid; scripts/coverage.sh run, README diff committed with the bead, --check exits 0.

> **`specs-c4model.9#2`** retro:proof Missing-kind failure reproduced by reviewer's recipe; post-fix: c4-module-missing-kind-json expected has module-gamma with visibility and no kind key. index-*/manifest.json input_hash changed legitimately because the embedded spec-model schema changed (go run ./testdata/gen).

> **`specs-c4model.10#1`** retro:proof Pre-fix failure observed: TestADerivedContractIgnoresTheAnchorOfTheConsumedCitation -> 'contract = <nil>, want contract-three with source derived'; passes after reading source.ConsumedContracts.

> **`specs-c4model.11#1`** retro:proof Pre-fix failure observed: 'container-alpha modules = " ghost-mod ,Ghost-Mod,Module-Gamma,module-gamma"'; 'module-gamma modules = "Module-Delta,module-delta"'. Not checked whether corpus checks reject case-variant slugs (model folds regardless): unverified; would settle with specs check on such a corpus.

> **`specs-c4model.12#1`** retro:proof pre-fix failure observed for the record-block source test
> evidence: go test -run RecordBlockSource ./cli/core/c4model -> 'contract = <nil>, want contract-x with source derived'; passes after fix. Note: record block must sit under '## Components' or the diagram node is undeclared_node.

### `environment` — 0

### kind missing (UNROUTED) — 4

> **`specs-c4model#1`** (author `plan-c4model`) retro: groom-beads scripts/bd-health.sh enumerates epics by sed-scraping every "id" key in 'bd list --type epic --json', so ids nested in epic metadata (e.g. excluded[].id) are passed to 'bd swarm validate' as epic ids: 39 scraped vs 9 real epics (jq -r '.[].id'), yielding 30+ false 'epic not found' hard failures and exit 1. 'bd swarm validate specs-c4model' run directly exits 0, Swarmable: YES.

> **`specs-c4model#2`** (author `plan-c4model`) retro: swarm-plan's partitioning reference gives no path for a plan that needs no new edges; 'bd dep add --file - </dev/null' on bd 1.3.1 exits 1 with 'Error: no dependency edges found', so an honest empty edge file fails to load. This plan's queue was already fully ordered by real blocks edges, so the edge file restates five existing intra-batch edges instead.

> **`specs-c4model.7#1`** (author `dev-c4model-3`) retro: clean bead, no friction.

> **`specs-c4model.8#1`** (author `dev-c4model-3`) retro: clean bead; DoD walked: all 33 AC-C4 ids cited (comm of FRD ids vs 'AC:' tags and case.toml cites is empty), generate-schemas.sh --check exit 0, sync --check exit 0, coverage --check exit 0, strict corpus check no findings.

Related `amend:` on the epic (not a finding): "edge file /tmp/claude-0/c4model-edges.jsonl not loaded — its 5 lines restate existing blocks edges; partition accepted as planned (3 dispatches)."

## Routing

| branch | count | findings |
|---|---|---|
| project (`decision`, `gap`, `scope`, `stale`) | 10 | 6 `decision`, 1 `gap`, 1 `scope`, 2 `stale`, as listed above |
| skill (`tooling`) | 1 | `specs-c4model#3` (`swarm` plugin → `outbox/swarm/`) |
| `proof` → this document | 8 | `.2#5`, `.3#1`, `.4#2`, `.6#1`, `.9#2`, `.10#1`, `.11#1`, `.12#1` |
| `environment` → ledger | 0 | — |
| **unrouted** | **4** | `specs-c4model#1`, `specs-c4model#2`, `specs-c4model.7#1`, `specs-c4model.8#1` |

**Unrouted is 4 of 23 (17%)**, against 0 on `specs-c4s2` and 9 of 29 on
`specs-c4s1`. Same two writers as `specs-c4s1`, not taxonomy drift: the
planner (`plan-c4model`) again wrote both its findings with no kind, and a
developer (`dev-c4model-3`) again wrote "clean bead" markers the protocol says
not to write. By the merge test, read for the candidates only:
`specs-c4model#1` is tooling against `groom-beads`, which this repository vends
→ project; `specs-c4model#2` is tooling against `swarm-plan` → `outbox/swarm/`;
`.7#1` is not a finding, `.8#1` is a `proof`.

## Validation

- **Baseline** (`manifest.checks_baseline`, untouched tree at `0a44644`): all
  12 declared checks passed, exit 0 (`checks_baseline_status: 0`).
- **At head**: each bead's close ran the declared checks through
  `bb-close-bead.sh`; `.6#1` and `.8#1` cite `coverage.sh --check` exit 0,
  `generate-schemas.sh --check` exit 0 and a strict corpus check with no
  findings. **Not re-run by this retro**, and no head-of-branch run is recorded
  on the epic.
- **CI**: not observed by this retro.

## Pushes

`origin/claude/swarm-c4model-iqetxu` is at `dc2cb1c`, the same commit as
`sprint/specs-c4model`'s head before this record:
`git log --oneline origin/claude/swarm-c4model-iqetxu..HEAD` was empty. This
record's own commit is local until the coordinator pushes. No
`origin/sprint/specs-c4model` ref was observed. Tracker pushes are the
`Stop` hook's and `bb-close-bead.sh`'s; not checked by this retro.

## What this record could NOT verify

- **Which model served each dispatch** — `manifest.models` maps tiers to
  models, but `dev-c4model-4` held a standard bead (`.9`) and two light ones
  (`.10`, `.11`); which model ran the batch is not on the epic.
- **Effort** — `manifest.effort` reads `unrecorded`.
- **Head-of-branch check results and CI** — not on the epic, not re-run here.
- **Whether corpus checks reject case-variant slugs** — `.11#1`: "unverified;
  would settle with specs check on such a corpus."
- **Why `.10`'s fix switched to `source.ConsumedContracts`** — `.10#1` records
  only the change and its passing test; whether the `.10` agent saw `.4#1` is
  not recorded anywhere.

## Candidates

**PENDING — returned to the coordinator to put to a human; no decision
recorded.**

Ranked by how much each changes the next run. "Fold" means the issue already
exists; the action is to attach this run's evidence and raise its recurrence
rather than file anew.

| # | candidate | audience → destination | findings | why it changes the next run | decision |
|---|---|---|---|---|---|
| A | **Fold: accept merges into the recorded sprint branch, not the checked-out one** — recurrence: `specs-c4s1`, `specs-c4s2` (its fold is itself pending), now `specs-c4model` | skill → existing `outbox/swarm/swarm/2026-10-07-accept-merges-into-the-current-branch.md` | `specs-c4model#3` | Third run in a row the first accept exits 2 and needs `git branch -f` | pending |
| B | **`swarm-plan`: a plan that needs no new edges returns no edge file** — say so in `references/partitioning.md` and have the coordinator skip `bd dep add --file` when there is none, because `bd dep add --file` on an empty input exits 1 (`Error: no dependency edges found`, measured on bd 1.3.1 here and in run `agent-plugins-w85b`) | skill → new `outbox/swarm/swarm-plan/2026-10-07-no-path-for-a-plan-with-no-new-edges.md` | `specs-c4model#2`; epic `amend:` 19:27:51Z | The planner padded the file with five restated `blocks` edges and the coordinator had to judge and discard it; the next zero-edge plan does the same | pending |
| C | **`swarm-dispatch`: make "a clean bead records none" unambiguous** — the instruction reads "A clean bead records none — say so rather than inventing filler.", and "say so" is being read as "write a `retro:` comment saying clean" | skill → new `outbox/swarm/swarm-dispatch/2026-10-07-clean-bead-instruction-invites-a-marker.md` | `specs-c4model.7#1`, `.8#1`; prior: `specs-c4s1.10#1`, `.11#1`, `specs-f6tl.5#3` | Third run with kind-less clean markers inflating the unrouted count — the taxonomy's health metric | pending |
| D | **Fold: planner findings carry no kind** — recurrence 1 | skill → existing `outbox/swarm/swarm-plan/2026-10-07-planner-findings-carry-no-kind.md` | `specs-c4model#1`, `#2` | Both planner findings again routed nowhere by kind | pending |
| E | **Fold: `bd-health.sh` validates non-epics** — recurrence comment (39 scraped vs 9 real epics) | project → comment on existing bead `specs-84yu` | `specs-c4model#1` | Every plan pass still has to ignore the health script's exit 1 and re-run `bd swarm validate` by hand | pending |
| F | **Pin the C4 model's unpinned rules in the durable specs** — run `retro-spec` (or a direct spec edit) over this feature's `decision` findings into `module-document-core` / the C4 Elements and C4 Model component docs and the CLI reference, first among them `.4#1`'s rule that the derived contract reads `contracts_consumed` from the source declaration (blueprint, else first record block) | project → bead (could be merged with `specs-c4s2`'s pending candidate G, the same ask for the C4 diagrams feature) | `.2#1`, `.2#2`, `.2#3`, `.3#2`, `.4#1`, `.5#1`, `.10#1`, `.12#1` | The feature blueprint is retired, so these rules now live only in retro comments and code; `.10`'s review fix reversed `.4#1` and cost a whole review loop (`.12`), and `specs-i20` is the same question on the check side | pending |
| G | **Fold: a schema edit's snapshot regeneration is invisible to the bead** — recurrence on `specs-xoe` (the snapshot failure should name `scripts/regenerate-cli-snapshots.sh`) | project → comment on existing bead `specs-xoe` | `specs-c4model.1#1` (also `.9#2` regenerated the same hashes) | Schema-touching beads keep failing `go test` once before finding the script; runs `agent-plugins-59ic`, `specs-sra`, `specs-c4s1` hit it too | pending |

**Kept in this record, not candidates** (do not clear the bar on this run's
evidence): `.2#4` and `.9#1` (`stale`, both fixed in-bead); `.5#2` (the
signature change was the plan's own sequencing, handled inside one dispatch);
`.11#1`'s unverified case-variant-slug question; `.12#1`'s note that a record
block must sit under `## Components` (already the corpus rule).

## What was filed

**Nothing yet** — decisions pending. No `retro:filed:` comment has been
written. `retro_filed`, `retro_declined` and `retro_routed` are not yet set on
the manifest; `findings_path` is.

## Counts

| | |
|---|---|
| harvested | 23 |
| routed — project | 10 (+1 unrouted that the merge test sends to the project: `specs-c4model#1`) |
| routed — skill | 1 (+1 unrouted that the merge test sends to `swarm-plan`: `specs-c4model#2`) |
| routed — `proof` (this document) | 8 |
| routed — `environment` (ledger) | 0 |
| unrouted | 4 |
| candidates | 7 (2 new skill issues, 2 folds into existing outbox issues, 2 folds onto existing project beads, 1 new project bead) |
| filed | 0 — pending |
| declined | 0 — pending |
