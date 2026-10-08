# Run record — `specs-qqnx`, Spec CLI QA round 2 fixes

## Read this first

**Every admitted bead closed, and three run-introduced defects remain open.**
`manifest.admitted` lists 20 beads, and all 20 are closed with commits on
`sprint/specs-qqnx`. That is the 13 planned beads (`.1`–`.13`), 2 amendments
admitted during dispatch (`.14`, `specs-jy7s`), 2 from review pass 1 (`.15`,
`.16`) and 3 from review pass 2 (`.17`–`.19`). **Review pass 3 was the last
the budget allowed. It filed `specs-qqnx.20`, `.21` and `.22` under the epic,
and all three are OPEN and unworked:**

- `specs-qqnx.20` (P2): "frontmatter/inline-comment: CRLF line endings give empty spans, wrong lines and the wrong owning property for every comment after the first"
- `specs-qqnx.21` (P3): "frontmatter/inline-comment: commentHash can locate the comment at a '#' inside an earlier quoted scalar when tabs precede it"
- `specs-qqnx.22` (P4): "frontmatter/inline-comment reports a comment after an explicit '? key' whose value is on the next line"

They are children of `specs-qqnx` and are not in `manifest.admitted`. **The
epic cannot close until they are worked or moved.**

**One rule took every review loop.** `specs-qqnx.9` added the
`frontmatter/inline-comment` warning. Every bead review produced inside the
epic (`.15`–`.22`) is a defect in that one rule. Review also filed 7 of its 15 backlog beads
against the same rule or the lexer beneath it. Patching each review finding as
it arrived did not converge in three passes. Candidate I.

**The swarm tooling needed hand repair five times.** The first `accept` moved
the checked-out branch and not the sprint branch (fifth run running). The
concurrent `accept` could not rebase a branch held by its worktree.
`bb-close-bead.sh` pushes the tracker although the dispatch prompt forbids it.
A two-bead dispatch failed `accept`'s signals only after the merge had deleted
its branch. The manifest refused to record the authorized third review pass.
Four new issues and three folds were written to `outbox/swarm/swarm/`.

**The project candidates are pending.** No human decision is recorded for
them here.

## The run

| | |
|---|---|
| run epic | `specs-qqnx` — "Spec CLI QA round 2 fixes". The epic is `open` |
| branch | `sprint/specs-qqnx`, cut from `origin/main` at `4676bc4`, not from the stale local `main` (manifest ledger entry 1) |
| `base_sha` → head | `4676bc4` → `74ee559`: 22 commits before this record's commit. 20 are bead commits, `c9b3260` is a follow-up for `.8`, and `c1af88a` regenerates coverage. No merge commits |
| queue | 13 planned (`.1`–`.13`). Amendments during dispatch added `.14` and `specs-jy7s`, both run-introduced. Review pass 1 added `.15` and `.16`, pass 2 added `.17`–`.19`, and pass 3 filed `.20`–`.22` (open, see above) |
| declared checks | 13 (`manifest.checks_ids`): `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `cli-conformance-table`, `skills-strict` |
| tiers (bead `metadata.tier`) | advanced: `.1`, `.10`. light: `.4`, `.11`, `.14`. standard: `.2`, `.3`, `.5`–`.9`, `.12`, `.13`, `jy7s`. No tier is recorded on `.15`–`.22` |
| model lineup | `manifest.models` maps the four tiers to models. Its values are in the manifest and are not restated here. Which model served each dispatch is not recorded |
| effort | `manifest.effort` reads `unrecorded` |
| standing authorization | Brett, 2026-10-08T20:26Z, thread reply "go". Every gate row took its default: `review_budget` 2, `tier_escalation` once, `move_foreign_beads` false, `merge_to_default` false, `dispatch_failure_threshold` 3 |
| review | `review_budget` 2. `review_passes` reads **2**, although three passes ran, because the manifest refuses 3 (`qqnx#5`) |
| partition | `manifest.partition_preview`: "dev-qqnx-1 .1; dev-qqnx-2 .3>.4>.5>.6; dev-qqnx-3 .8>.10>.9; dev-qqnx-4 .2>.12; dev-qqnx-5 .7; dev-qqnx-6 .11; dev-qqnx-7 .13". The planner packed beads by command file instead of by the C4 union, because every bead shares `container-spec-cli#Command Surface` |
| dispatches | `dev-qqnx-1` … `dev-qqnx-10`. `consecutive_dispatch_failures` 0 |
| concurrency cap | 20 |
| durable trigger / event-log hooks | **not armed.** The epic's first `amend:` says so |
| trailers | `manifest.trailers` was set by the coordinator after the first close (`qqnx#2`). All 22 commits carry both session trailers. `a22fd79`'s trailers were written by hand |

## Results

| bead | tier | agent | commit | outcome |
|---|---|---|---|---|
| `specs-qqnx.1` | advanced | `dev-qqnx-1` | `a22fd79` | closed — Update FRDs, ACs and standards for the CLI QA round 2 decisions |
| `specs-qqnx.2` | standard | `dev-qqnx-4` | `706e965` | closed — criterion --changed picks a stale local main over origin/main as its default base |
| `specs-qqnx.3` | standard | `dev-qqnx-2` | `f9165d7` | closed — headings matches carry a paste-ready reference; rename heading_path |
| `specs-qqnx.4` | light | `dev-qqnx-2` | `23014fb` | closed — headings --emit with zero matches is refused instead of printing nothing |
| `specs-qqnx.5` | standard | `dev-qqnx-2` | `3c5a6a2` | closed — headings omits retired documents unless --all |
| `specs-qqnx.6` | standard | `dev-qqnx-2` | `8f0610c` | closed — feature and headings answer compactly by default; --full returns every field |
| `specs-qqnx.7` | standard | `dev-qqnx-5` | `922f034` | closed — Compact refs --outbound groups by target and carries its path |
| `specs-qqnx.8` | standard | `dev-qqnx-3` | `4be06cf`, `c9b3260` | closed — check answers with findings first and a summary; passed rules only when verbose |
| `specs-qqnx.9` | standard | `dev-qqnx-3` | `01aa3a7` | closed — Warn on a frontmatter value cut short by an inline YAML comment; fix adr-012 |
| `specs-qqnx.10` | advanced | `dev-qqnx-3` | `8cc328a` | closed — check <path> also reports references elsewhere that no longer resolve to that document |
| `specs-qqnx.11` | light | `dev-qqnx-6` | `a4e625a` | closed — slug-exists echoes the slug as the member declares it |
| `specs-qqnx.12` | standard | `dev-qqnx-4` | `2f1dd19` | closed — criterion --changed carries each criterion's name; --full adds its record |
| `specs-qqnx.13` | standard | `dev-qqnx-7` | `418b592` | closed — specs --version reports the module version and VCS revision |
| `specs-qqnx.14` | light | `dev-qqnx-3` | `47d4215` | closed — spec-cli container still says slug-exists echoes slug as asked (run-introduced by `.11`) |
| `specs-jy7s` | standard | `dev-qqnx-8` | `0c5a41e` | closed — Restate the cli-output-standards conformance table's measurement base (run-introduced by `.8`) |
| `specs-qqnx.15` | — | `dev-qqnx-9` | `fc569da` | closed — inline-comment fires where no value is cut short (review pass 1) |
| `specs-qqnx.16` | — | `dev-qqnx-9` | `793b724` (empty close commit; the fix is in `fc569da`) | closed — inline-comment reports an own-line comment after a folded block scalar (review pass 1) |
| `specs-qqnx.17` | — | `dev-qqnx-10` | `f1b5b5f` | closed — inline-comment no longer reports a comment after a number, bool or null (review pass 2) |
| `specs-qqnx.18` | — | `dev-qqnx-10` | `f414f3c` | closed — inline-comment span starts at a tab, not the '#' (review pass 2) |
| `specs-qqnx.19` | — | `dev-qqnx-10` | `74ee559` (tests; the code fix is in `f1b5b5f`) | closed — inline-comment fires after an anchor whose collection is on the next lines (review pass 2) |
| `specs-qqnx.20` | — | — | — | **OPEN** — CRLF line endings break inline-comment spans, lines and owners (review pass 3, P2) |
| `specs-qqnx.21` | — | — | — | **OPEN** — commentHash can find a '#' inside an earlier quoted scalar after tabs (review pass 3, P3) |
| `specs-qqnx.22` | — | — | — | **OPEN** — inline-comment reports a comment after an explicit '? key' (review pass 3, P4) |

`c1af88a` ("specs-qqnx integration: regenerate coverage figures") belongs to no
bead.

### How the run's scope moved

These are the epic's amendment comments, verbatim:

- 20:43Z — "amend: coordinator did not install the bd event-log hooks or arm a durable out-of-session trigger; visibility is swarm status plus subagent completion notifications in this session. Sprint branch sprint/specs-qqnx cut from origin/main 4676bc4 (not local main, per ledger entry 1)."
- 20:59Z — "amend: admitted specs-qqnx.14 (run-introduced by specs-qqnx.11: the container doc still says slug-exists echoes the slug as asked). Filed against the run's own diff, so admitted; it goes through groom & plan before dispatch."
- 21:08Z — "amend: groomed specs-qqnx.14 inline (one-sentence doc fix, tier light) and assigned it to dev-qqnx-3 after specs-qqnx.9, since that batch also edits spec-cli-container.md."
- 21:09Z — "amend: admitted specs-jy7s (run-introduced by specs-qqnx.8: conformance table's measurement-base sentence). Parented under the run epic and sequenced after .9 and .14 so it runs once at integration, with scripts/coverage.sh."
- 21:28Z — "Amend: review pass 1 admitted specs-qqnx.15 and specs-qqnx.16 (run-introduced by specs-qqnx.9, both in cli/core/frontmatter/frontmatter.go). Groomed as one dispatch, dev-qqnx-9."
- 21:43Z — "Amend: review pass 2 admitted specs-qqnx.17, .18, .19 (inline-comment rule, cli/core/frontmatter). One dispatch, dev-qqnx-10."
- **Review pass 3 has no amendment comment on the epic.** Its record is what
  the tracker holds. `.20`–`.22` were created by `review-qqnx` at
  21:59–22:00Z under the epic. One comment on `.20` reads "Backlog twin
  (pre-existing Property.Raw/Position side of the same CRLF lexer miscount):
  specs-x8oo. One fix in absolute() may close both."

**Every amendment the run admitted was its own defect.** Two were documentation
sentences that an implementing bead left stale (`.14` after `.11`, `jy7s`
after `.8`), and `.8` also needed a follow-up commit for the same reason
(`.8#4`). The other five (`.15`–`.19`), and the three left open, are the
inline-comment rule from `.9`.

## What the run decided

The findings of kind `decision`, `gap`, `scope`, `sequencing`, `stale` and
`unverifiable` mark places where a document did not answer a question a bead
needed. All 19 appear verbatim below. These bear on durable documents:

- **Contradictory criteria resolved toward Brett's decision.** `.1#1` amended AC-SPECREF-002.9 because it "directly contradict[ed]" the amended .11. `.10#2` amended AC-SPECANCH-003.2 for the same reason. Both beads named only one of each pair, so grooming missed the sibling criterion that also needed amending.
- **Concept and field are separate.** `.1#2` renamed only the output field `heading_path` → `anchor`, and "heading path" stays the selection concept.
- **Where a bead's doc work stops.** `.1#6` left the module blueprint's Revisions, Corpus and Frontmatter sections to the implementing beads, against the `c4_impacts` metadata. `.8#4` missed a container paragraph its `c4_changes` named, because its ACs did not name it. `.11#2` and `.8#3` each left a stale sentence for a later bead (`.14`, `jy7s`).
- **The inline-comment rule's design.** `.9#3` chose the go-yaml lexer over the Design's `parser.ParseComments` so that Parse's spans could not move. Review then found eight edge cases in three passes (`.15`–`.22`), and `.17` moved the plain-scalar test to the parsed tree.
- **An AC that cannot be met as written.** `.6#1`: "AC 'specs headings module-document-core is under 2.5 KB' cannot be met with the pinned compact fields … 3241 bytes YAML". The bound is in the bead's own acceptance list, not in AC-SPECSTR-003.16, so the closed bead is the only place it lives.
- **Snapshot cases moved to `--verbose` or `--full`** to keep proving what they were written for (`.8#2`, `.6#2`). Alternative twins were rejected.
- **`--version`** is a pre-declared local root flag (`.13#2`). `--full` without `--changed` is refused with exit 2 (`.12#1`).
- **Ordering.** `.9#1`: a dependency closed in the tracker is not the same as its commit being in the dispatch branch. Filed to `outbox/swarm/` by the merge test.

## Scope

- **Excluded.** `manifest.excluded` is `[]`. The epic description puts several things out of scope and says they are filed separately: the AC identifier convention change, the coverage, context and edit-hook FRDs, skills adoption (`specs-nlhk`), and the feature/criterion rename.
- **Moved.** `manifest.moved` is `[]`. `specs-jy7s` was parented under the run epic when it was admitted (amendment at 21:09Z).
- **Superseded before the run.** `specs-8i2` (`--version`) carries a `supersedes` edge from `.13`. It was closed at 17:24Z, before the run began.
- **Admitted then dropped.** None.
- **Left open in the epic.** `.20`, `.21` and `.22`, filed by the final review pass. **The run finished its admitted set, but the epic is not finished.**
- **Filed to the backlog by review** (`discovered-from` edges, all `open`, all created by `review-qqnx`):
  - pass 1 (21:26Z): `specs-3cid`, `specs-gsuv` (`.10`), `specs-7457` (`.2`), `specs-25uh`, `specs-ve3r` (`.3`), `specs-hhw0` (`.7`), `specs-jexj` (epic), `specs-000q` (`.4`)
  - pass 2 (21:42Z): `specs-ejcl` (`.15`), `specs-kgpq` (`.16`), `specs-8jxr` (`.9`)
  - pass 3 (22:00Z): `specs-ihaq` (P2, epic), `specs-x8oo` (epic), `specs-oeya` (`.18`), `specs-j3sd` (`.17`)
- **Of those 15, 7 concern the inline-comment rule or the frontmatter lexer beneath it:** `ihaq`, `x8oo`, `oeya`, `j3sd`, `8jxr`, `kgpq`, `ejcl`.

## Every finding, verbatim

43 findings were harvested from the epic, its 22 children and `specs-jy7s`, with
`bd export | jq 'select(.id=="specs-qqnx" or (.id|startswith("specs-qqnx.")) or .id=="specs-jy7s") | .comments[] | select(.text|startswith("retro:"))'`.
`.20`–`.22` carry no `retro:` comment. Each finding is numbered
`<bead>#<ordinal>` by its position among that bead's `retro:` comments, with
the `specs-` prefix dropped. Newline markers are folded to spaces. `.15#1` and
`.16#1` are one comment posted to both beads of a single dispatch.

### `decision` — 13

- **`qqnx.1#1`** (dev-qqnx-1): retro:decision AC-SPECREF-002.9 (@negative) was amended alongside AC-SPECREF-002.11 and the D2 rationale, though the bead named only .11 and D2. evidence: before the edit, AC-SPECREF-002.9 read "And that holds for a reference held by another document that no longer resolves to the document named" — a direct contradiction of the amended .11. It now reads "reports no finding located in a document other than the one named, except those AC-SPECREF-002.11 names". The FEAT-SPECREF-002 description ("reports the findings located in that document and no others") and the anchor-references-frd example citing "AC-SPECREF-002.9 applied to the anchor half" were amended for the same reason. Rejected alternative: leaving .9 as written, which would have left specs-qqnx.10 building against two contradictory criteria. The outside-git-work-tree fallback is a new criterion, AC-SPECREF-002.13, rather than a clause of .11.
- **`qqnx.1#2`** (dev-qqnx-1): retro:decision Kept "heading path" as the concept term in the structure FRD and renamed only the field: AC-SPECSTR-003.9 now reads "its \"anchor\" is its heading path"; .10/.11 and anchor-references-frd keep "heading path" for the selection concept. evidence: grep -rni "heading path" specs/ shows the concept used in anchor-references-frd (AC-SPECANCH-002.5, .10) and document-core-module for selection semantics, none of which name the output field. Renaming the concept everywhere would have rewritten anchor-resolution criteria outside this bead. New structure criteria: AC-SPECSTR-003.12 (ref), .13 (emit, zero matches), .14/.15 (retired members, --all), .16 (compact view), .17 (--full with --emit). AC-SPECSTR-003.1 is now the full-view criterion; .6 and .8 are scoped "without the emit option".
- **`qqnx.1#3`** (dev-qqnx-1): retro:decision The criterion ids the implementation beads build against, for anyone checking citations: gherkin-adoption AC-GHK-007.10/.11/.12 (feature compact view, --full, --full with --emit), AC-GHK-008.10 (amended candidate order, table now expects origin/main for "main, origin/main, master"), AC-GHK-008.11 (name), AC-GHK-008.12 (adds a --full row), AC-GHK-008.14/.15 (stale-local warning, and its absence), AC-GHK-008.16 (record/base_record); spec-index AC-SPEC-003.9/.10 (inline comment), AC-SPEC-006.11/.12/.13 (summary, --verbose, exit status unchanged), AC-SPEC-007.2 (amended echo), AC-SPEC-011.7 (--version); reference-integrity AC-SPECREF-003.9 (outbound compact). evidence: build/specs check --strict on each edited FRD -> findings: [], exit 0. Consequence: the existing test citing AC-GHK-008.10 still asserts local main first, and specs-qqnx.2 has to change it; that test passes until then because tests exercise code, not FRD text.
- **`qqnx.1#4`** (dev-qqnx-1): retro:decision The conformance table check row does not mention specs-qqnx.10 (check <path> reporting broken references elsewhere). Only specs-qqnx.8 flips that row. evidence: specs-qqnx.8 and specs-qqnx.10 would otherwise both edit the check row, the merge conflict the reorder exists to prevent. specs-qqnx.10 changes which findings are reported, not the shape of the output, so the output standard has nothing to track for it. Row order as committed: check, list, feature, next-id, criterion, links, refs, index, slug-exists, c4, headings. scripts/check-cli-conformance-table.py -> "11 registered commands, each with a row".
- **`qqnx.12#1`** (dev-qqnx-4): retro:decision --full under criterion <id> (no --changed) is refused exit 2 like --since/--exit-code; record/base_record are pointers so absent ones are omitted. evidence: only snapshot changed was help-criterion (help text gained --full); regenerated, diff reviewed; e2e TestChangedFullCarriesRecords covers added/changed/removed with and without --full and --json.
- **`qqnx.13#2`** (dev-qqnx-7): retro:decision --version is a local root flag pre-declared in installVersionFlag (cli/cmd/version.go) with custom usage text; rejected cobra's auto flag (generic 'version for specs' help).
- **`qqnx.10#2`** (dev-qqnx-3): retro:decision AC-SPECANCH-003.2 contradicted the amended AC-SPECREF-002.11 (it still said a single-document check reports NO anchor finding elsewhere, "an anchor reference held by another document into a heading of the document named that no longer resolves" included). Followed 002.11 / Brett's decision and amended 003.2 to "except the unresolved-anchor findings AC-SPECREF-002.11 names", keeping ambiguous anchors elsewhere and anchors into other documents excluded; rewrote its rationale paragraph. Snapshot check-path-anchor-elsewhere flipped from exit 0/no findings to exit 1 with the referrer's reference/unresolved-anchor, now citing AC-SPECREF-002.11 and AC-SPECANCH-003.2. Rejected: leaving 003.2 as written (two criteria in direct conflict, one snapshot proving the retired behavior). evidence: specs/features/reference-integrity/anchors/anchor-references-frd.md at a22fd79: "Then the CLI reports no anchor finding located in a document other than the one named / And that holds for an anchor reference held by another document into a heading of the document named that no longer resolves"; its examples paragraph had already been amended by specs-qqnx.1 to say the write-time check reports them.
- **`qqnx.10#3`** (dev-qqnx-3): retro:decision Silent choices a diff-reader may not see as choices: (1) Corpus reads no revision; CheckFor(path, formerSlugs ...string) takes them, and cmd reads HEAD via new revisions.Repository.FileAt (ls-tree + cat-file blob) and new corpus.DeclaredSlug (parsed frontmatter, else the legible recovered slug line). (2) The repo is opened from the checked file's directory and the path made relative to git's toplevel after EvalSymlinks on both sides. (3) An unborn HEAD writes the note (FileAt returns an Environment error) rather than silently reporting nothing. (4) A former slug equal (case-insensitively) to the current slug contributes nothing. (5) Updated reference-integrity-blueprint.md's EARS bullet and test-name paragraph too (not in c4_docs) — it still described the round-0 "report every reference finding wherever written" rule and cited a test name that no longer exists (TestCheckForReportsTheReferencesARenamedSlugBroke); added an AC-SPECREF-002.13 coverage row. evidence: git diff a22fd79.. -- cli/core/corpus/check.go cli/core/revisions/revisions.go cli/cmd/check.go specs/features/reference-integrity/reference-integrity-blueprint.md
- **`qqnx.8#2`** (dev-qqnx-3): retro:decision Three snapshot cases whose stated purpose is the checks accounting (AC-SPEC-006.9: not-requested vs passed) now run with --verbose: check-only, check-only-anchor-rule, check-only-reference-rule. Without it their expected output would no longer contain the checks list they exist to prove. c4_check_test.go's c4RunCase likewise passes --verbose and decodes {findings, checks} into a FindingReport (specmodel.FindingReport's UnmarshalJSON rejects the unknown "summary" key and requires "checks"). Rejected alternative: leaving the cases default-shaped and adding separate --verbose twins, which would duplicate fixtures for no extra coverage. evidence: git diff -- cli/testdata/cases/check-only*/case.toml; check-only-schema-slug (cites AC-SPEC-006.8 for its findings, not checks) was left without --verbose.
- **`qqnx.9#3`** (dev-qqnx-3): retro:decision Detected comments with goccy/go-yaml's lexer (lexer.Tokenize over the block) instead of the design's parser.ParseBytes(..., parser.ParseComments): a comment is reported when another token ended on its line, a multi-line scalar counting as ending on its last line (its Origin's newlines). This leaves Parse's mode untouched, so no existing value or span can move — the risk the design asked to verify. Comments after a block-scalar indicator ("key: | # c") or a key whose value is nested are also reported (text after the # is still not part of anything); a "#" inside a quoted value, inside a block scalar, or with no space before it (http://x#y) is not a comment and is not reported. The owning property is the last top-level key at or before the "#". Regenerated schemas changed every index-* snapshot's input_hash (the profile schema is hashed); check-verbose gained a passed line and check-only-unknown-rule's known-rule list gained the rule. evidence: cli/core/frontmatter/frontmatter_test.go TestInlineCommentsAreTheCommentsTrailingAValue; git diff --stat cli/testdata/cases.
- **`qqnx.6#2`** (dev-qqnx-2): retro:decision existing cases feature-query-by-id(-json), headings-json, headings-document and query_test warm-index test gained --full to keep exercising the full record they were written for; feature-query-all and headings-all now cite the compact criteria. Alternative rejected: regenerating them compact, which would silently drop what they prove (criteria text, section span). evidence: cli/testdata/cases/*/case.toml args; cli/cmd/query_test.go TestFeatureQueryIsServedFromTheWarmIndexWithoutRewalking
- **`qqnx.5#1`** (dev-qqnx-2): retro:decision predicate exported as Corpus.InEffect (frontmatter-less or status-less member is in effect, matching list's parse-error survival); selectHeadings takes a nil-able filter func, nil under --all. Alternative rejected: filtering on a profile passed into cmd, which would put lifecycle logic in cli/cmd. evidence: cli/core/corpus/query.go InEffect; snapshots headings-retired-{omitted,all,named,emit-refusal,emit-live}; bb-checks.py exit 0
- **`qqnx.3#1`** (dev-qqnx-2): retro:decision ref for a member declaring no slug is <path>#<anchor> (the Design's decided default); alternative rejected: omit ref, which leaves --emit refusal entries unlisted for slugless members evidence: cli/cmd/headings.go selectHeadings refPrefix := envelope.DeclaredSlug(entry); if "" then entry.Path

### `gap` — 0

None.


### `scope` — 1

- **`qqnx.1#6`** (dev-qqnx-1): retro:scope The module blueprint sections Revisions (candidate order), Corpus (single-document selection) and Frontmatter/Findings (inline-comment rule) were not amended here. Only the heading_path rename was made in the module and container blueprints. evidence: the bead Design says "Docs only. Edit the FRDs and the standard named in the description", plus the rename sweep. The ACs of specs-qqnx.2, .9 and .10 each say "document-core-module.md › <section> states the new ... (or is confirmed already amended by specs-qqnx.1)", so those sections stay with the implementing beads. The c4_impacts metadata on this bead (module-document-core#Corpus, #Revisions) suggested otherwise.

### `sequencing` — 1

- **`qqnx.9#1`** (dev-qqnx-3): retro:sequencing The tracker reported specs-qqnx.9 ready (its dependency specs-qqnx.3 closed), but the dispatch worktree bb/dev-qqnx-3 does not contain .3's commit: closed in the tracker is not the same as present in this branch, and the bead's Ordering section depends on the latter. evidence: git merge-base --is-ancestor f9165d7 HEAD → exit 1 on bb/dev-qqnx-3 (HEAD 7b22b57); git log sprint/specs-qqnx shows f9165d7 specs-qqnx.3 P2 headings matches carry a paste-ready reference. swarm ready --run specs-qqnx --agent dev-qqnx-3 listed specs-qqnx.9 state: ready.

### `stale` — 3

- **`qqnx.8#3`** (dev-qqnx-3): retro:stale:project The cli-output-standards conformance table preamble still reads that "Prints today" is measured with a binary built at 4676bc4, but the check row now describes this branch's behavior. Not edited: other beads in this run flip rows in the same table concurrently, so the preamble is better fixed once at integration. evidence: specs/standards/cli-output-standards.md: "\"Prints today\" is measured with a binary built at `4676bc4`"; check row now "| `check [path]` | a YAML mapping of `summary` ... | conforms |".
- **`jy7s#1`** (dev-qqnx-8): retro:stale:specs/standards/cli-output-standards.md feature row said default prints full feature records; measured default is compact evidence: build/specs feature FEAT-GHK-001 prints id, code, name, path, criteria_count; --full adds document_slug, description etc. Every other row re-measured against binary built at 47d4215 and read true.
- **`qqnx.11#2`** (dev-qqnx-6): retro:stale:specs/standards/cli-output-standards.md slug-exists row. evidence: the bead's acceptance criterion says the row 'stays conforms', but the row read 'not yet conforming: echo slug as the member declares it'. Set to conforms and the echo wording updated, as the standards row now describes the amended AC-SPEC-007.2. Left the container sentence in specs/containers/spec-cli/spec-cli-container.md (line near 274, 'slug, as asked') for bead specs-qqnx.14.

### `unverifiable` — 1

- **`qqnx.6#1`** (dev-qqnx-2): retro:unverifiable AC 'specs headings module-document-core is under 2.5 KB' cannot be met with the pinned compact fields (anchor, ref, level, start_line, plus declared) on today's corpus: 23 matches, 3241 bytes YAML / 3109 bytes JSON evidence: build/specs headings module-document-core | wc -c -> 3241; ... | grep -c anchor: -> 23; --full was 8515 bytes before this bead. Fields are pinned by the Design and cli-output-standards, so I did not drop any; the 2.5 KB figure predates module-document-core growing. Feature AC met: specs feature -> 8271 bytes (< 15 KB); --full byte-identical to pre-change (cmp against binary built at 7ed1868, YAML and JSON).

### `tooling` — 7

- **`qqnx.1#7`** (dev-qqnx-1): retro:tooling:swarm bb-close-bead.sh has no way to skip its step-5 `bd dolt push`. This dispatch forbade pushing, so I ran steps 1-4 by hand and committed without the push. evidence: bb-close-bead.sh step 5 runs `if bd dolt remote list ... ; then bd dolt push` with no flag or environment variable to turn it off. The dispatch prompt said "You merge nothing and you push nothing (no git push, no bd dolt push)". Separately, `--trailers-from specs-qqnx` exited 2 with "specs-qqnx carries no `trailers` in its manifest". The session attribution trailers were therefore written into the commit message by hand. Commit a22fd79. The closure is in the local database only, and the coordinator owns the push.
- **`qqnx#1`** (coordinator-qqnx): retro:tooling:swarm swarm dispatch accept fast-forwarded the CHECKED-OUT branch rather than the run's target: with the coordinator checkout on claude/project-thread-nr65x0, accept for dev-qqnx-1 reported 'serialized ... fast-forwarded into sprint/specs-qqnx' but sprint/specs-qqnx stayed at 4676bc4 and claude/project-thread-nr65x0 moved to a22fd79; it exited 2 on the ancestry guard. Its --explain lists 'git rev-parse --verify HEAD^{commit} (the repository is on the target)' but nothing refuses when HEAD is not the target. Repaired by hand: git branch -f claude/project-thread-nr65x0 4676bc4; git checkout sprint/specs-qqnx; re-ran accept -> exit 0. Expected: accept refuses (exit 2, naming the target) when HEAD is not on the target branch, before any merge.
- **`qqnx#2`** (coordinator-qqnx): retro:tooling:swarm bb-close-bead.sh always ends with bd dolt push and has no flag to skip it, while the dispatch prompt says the dev agent pushes nothing; dev-qqnx-1 ran its steps 1-4 by hand. Also --trailers-from specs-qqnx exited 2 because the manifest carried no trailers; the coordinator has now set manifest.trailers.
- **`qqnx#3`** (coordinator-qqnx): retro:tooling:swarm swarm dispatch accept's concurrent case runs 'git rebase <target-sha> bb/<dispatch>' from the main checkout while bb/<dispatch> is still checked out in its worktree, so it always fails: dev-qqnx-7 accept exit 2, 'fatal: bb/dev-qqnx-7 is already used by worktree at /home/user/.bb-worktrees/dev-qqnx-7'. Workaround: git rebase sprint/specs-qqnx inside the worktree (exit 0, clean), then accept takes the serialized path. Expected: rebase inside the dispatch worktree (git -C <worktree> rebase <target-sha>).
- **`qqnx#4`** (Claude): retro:tooling:swarm A dispatch that fixes two beads in one commit fails accept's 'committed' signal (one '<id> P<n>' subject per bead) and 'findings' (no retro: comment) after the merge has already happened and the dispatch branch is deleted; re-accept then fails 'bb/dev-qqnx-9 could not be resolved'. evidence: accept exit 1 acceptance_failures [specs-qqnx.15: committed, findings, specs-qqnx.16: committed, findings]; second accept exit 2 rev-parse 128. Fixed by rewording to fc569da + empty close commit 793b724, adding retro:proof comments, recreating the branch at 793b724; third accept exit 0 landed true.
- **`qqnx#5`** (Claude): retro:tooling:swarm review_budget semantics: authorized as '2 loops = 3 passes' and recorded review_budget=2, but swarm manifest set refuses review_passes=3 ('would be 3 against a budget of 2'), i.e. it treats the budget as total passes. Pass 3 ran anyway as authorized (read-only review); no fix loop follows it. evidence: swarm manifest set --run specs-qqnx --key review_passes --value 3 → exit 1.
- **`qqnx.8#4`** (dev-qqnx-3): retro:tooling Missed the spec-cli container's Command Surface paragraph (still described checks-then-findings) though the bead's c4_changes named it; landed as a follow-up commit c592e7c rather than an amend, per the skill. evidence: specs/containers/spec-cli/spec-cli-container.md: "Its answer is the generated `specmodel.FindingReport` — a mapping of `checks` and `findings`, nothing else" — found while working specs-qqnx.10. The bead's acceptance criteria name only cli-output-standards and the CLI guide as docs to update, not the c4_docs, which is how it was missed.

### `proof` — 15

- **`qqnx.2#1`** (dev-qqnx-4): retro:proof stale-main test failed before the fix for the stated reason evidence: go test ./core/revisions -> 'Base = {Ref:main Commit:0ac232d... Behind:<nil>}, want origin/main at 79fe2b5...' (stub Behind field added first so it compiled); passes after reorder. No snapshot output changed (git status shows no snapshot diffs); module/container/guide docs amended, sync --check exit 0.
- **`qqnx.1#5`** (dev-qqnx-1): retro:proof I measured the "prints today" cells and the view tables in cli-output-standards against the current binary instead of copying them from the bead. evidence: build at 4676bc4: `build/specs slug-exists ADR-001` -> `slug: ADR-001`; a headings match carries path, text, identifier, heading_path, level, section (no heading-line span, so the compact `start_line` is documented as derived); `build/specs feature FEAT-GHK-001 --json | jq ".features[0]|keys_unsorted"` -> id,code,document_slug,path,name,description,source,span,criteria; `criterion --changed` entries carry id, classification, path; `refs adr-012 --outbound` groups under one `path: specs/adrs/adr-012...`. The first draft of the headings table used `id` where the field is `identifier`, and measuring caught it.
- **`qqnx.17#1`** (dev-qqnx-10): retro:proof Fix in f1b5b5f. Findings count before -> after: 'version: 1.2 # beta' 0->1 '# beta'; 'count: 5 # five' 0->1; 'draft: true # for now' 0->1; 'x: null # c' 0->1; 'x: 0x1F # c' 0->1; 'slug_extra: a # c' 1->1. Unchanged at 0: 'tags2: # list below'+list, '[a, b] # flow', '{a: b} # flow', '"a" # note', "'a' # note", '| # header', '>-' then own-line '#', full-line '#'; '  - x # y' 1->1. 'x: .inf # c' is a frontmatter/parse error in the CLI (json: unsupported value: +Inf), so 0 both sides there; unit test covers .inf -> 1. beforePrev guard removed (dead under the AST test; specs-kgpq left open). Checks at 74ee559: cli/ go vet ./... exit 0; go test -count=1 ./... exit 0; scripts/format-check exit 0; scripts/ci.sh exit 0 (RESULT: PASS) after scripts/coverage.sh (go1.24.7) regenerated README (frontmatter 225/203 90.2% -> 245/220 89.8%; total 91.4% unchanged). Measured: specs check <file> --only frontmatter/inline-comment --json, built binaries from 793b724 (before) and 74ee559 (after), on a copy of specs/ + specs.toml with one probe ADR per case.
- **`qqnx.16#1`** (Claude): retro:proof Fixed in inlineComments (cli/core/frontmatter/frontmatter.go). evidence: specs check --only frontmatter/inline-comment on a workspace copy, before/after: tags: # list below → 1→0; aliases: [a, b] # flow → 1→0; note: "a" # note → 1→0; >/>- followed by an own-line # comment → 1→0; description: a # b stays 1 at the #; | and |- stay 0. go test ./... exit 0; scripts/ci.sh PASS after coverage regen (frontmatter 89.7%→90.2%).
- **`qqnx.15#1`** (Claude): retro:proof Fixed in inlineComments (cli/core/frontmatter/frontmatter.go). evidence: specs check --only frontmatter/inline-comment on a workspace copy, before/after: tags: # list below → 1→0; aliases: [a, b] # flow → 1→0; note: "a" # note → 1→0; >/>- followed by an own-line # comment → 1→0; description: a # b stays 1 at the #; | and |- stay 0. go test ./... exit 0; scripts/ci.sh PASS after coverage regen (frontmatter 89.7%→90.2%).
- **`qqnx.10#1`** (dev-qqnx-3): retro:proof The new e2e test fails against the old selection and passes against the new; this repository reproduces the bead's five-citation scenario exactly. evidence: with cli/cmd/check.go temporarily calling c.CheckFor(path) (no former slug), go test ./specs -run TestCheckPathReportsCitations → "check_path_e2e_test.go:102: exit 0, want 1"; restored → PASS. On this repo, sed slug: adr-005 → adr-905 (uncommitted, reverted after): `specs check --json specs/adrs/adr-005-*.md` exit 1, summary errors 5, findings = reference/unresolved at adr-004:57, adr-004:63, adr-006:18, adr-006:60, adr-010:59 — line-for-line identical to the whole-corpus `specs check --json` (diff empty), stderr empty. No "own slug finding" appeared: adr-905 matches the adr slug pattern, so the AC's "own slug finding plus" holds only for a rename that also breaks the pattern. Full bb-checks.py rc=0.
- **`qqnx.8#1`** (dev-qqnx-3): retro:proof The declared corpus check text changes against the run baseline, as the bead predicted; everything else matches and all 13 checks exit 0. evidence: bb-checks.py (no args) rc=0; "=== corpus" now prints "summary:\n  errors: 0\n  warnings: 0\n  unverified: 0\nfindings: []" (5 lines) where the baseline printed 36 "- rule: ... outcome: passed" entries then "findings: []". go test ./... passes, including new TestCheckSummaryCountsTheFindingsItReports and TestCheckListsChecksOnlyUnderVerboseAndKeepsItsStatus (AC-SPEC-006.11/.12/.13) and the new snapshot case check-verbose.
- **`qqnx.9#2`** (dev-qqnx-3): retro:proof The rule finds exactly the one instance the bead names, and the fix clears it. evidence: before quoting, `build/specs check --only frontmatter/inline-comment` on this repo → 1 warning, specs/adrs/adr-012-reference-conventions.md 3:161-181, property "description", exit 0. After quoting: `specs check --strict` exit 0 (summary 0/0/0), `specs list --json` shows adr-012 description ending "replacing @ and the # element convention"; `specs check --config rfcs/specs.toml --only frontmatter/inline-comment` exit 0, no findings. New snapshot cases check-inline-comment (exit 0, one warning at line 4 col 16, the full-line comment on line 2 not reported) and check-inline-comment-strict (--strict exit 1). Full bb-checks.py rc=0 (after gofmt on the new test).
- **`qqnx.7#1`** (dev-qqnx-5): retro:proof outbound compact grouping verified evidence: bb-checks.py (all 13) exit 0; regenerated snapshots refs-outbound, refs-outbound-unresolved, help-refs show only the intended diff; new TestTheOutboundCompactRefsAnswerGroupsByCitedSlug. Moved the unresolved-coverage assertion out of the inbound-only compact test since inbound refs never carry unresolved.
- **`qqnx.4#1`** (dev-qqnx-2): retro:proof clean bead; built binary: headings 'module-document-core#nope' --emit -> 0 bytes stdout, exit 2; 'document-core#open-questions' without --emit -> declared: false, matches: [], exit 0. bb-checks.py exit 0. evidence: new snapshot cases headings-emit-no-match, headings-emit-undeclared-document cite AC-SPECSTR-003.13
- **`qqnx.3#2`** (dev-qqnx-2): retro:proof corpus test cli/cmd/headings_ref_internal_test.go passes over this repo's specs/ corpus; full bb-checks.py (no args) exit 0 after fixing my own over-broad 2.0->2.1 sed in specindex_test.go rejected-versions list evidence: go test ./cmd -run EveryHeadingsMatch -> ok; bb-checks.py -> 0
- **`qqnx.19#1`** (dev-qqnx-10): retro:proof Code fix landed in f1b5b5f (specs-qqnx.17: AST classification skips AnchorNode.Name); tests in 74ee559. Findings before -> after: 'a: &anc # note\n  - x' 1->0; 'a: &anc # note\n  k: v' 1->0; 'a: &anc v # c' 1->1 (anchored plain scalar still warns). New unit test TestACommentAfterAnAnchoredCollectionIsNotInline fails on the two anchored-collection cases against 793b724's frontmatter.go, passes at HEAD. Checks at 74ee559: cli/ go vet ./... exit 0; go test -count=1 ./... exit 0; scripts/format-check exit 0; scripts/ci.sh exit 0 (RESULT: PASS) after scripts/coverage.sh (go1.24.7) regenerated README (frontmatter 225/203 90.2% -> 245/220 89.8%; total 91.4% unchanged). Measured: specs check <file> --only frontmatter/inline-comment --json, built binaries from 793b724 (before) and 74ee559 (after), on a copy of specs/ + specs.toml with one probe ADR per case.
- **`qqnx.18#1`** (dev-qqnx-10): retro:proof Fix in f414f3c. Span text before -> after: 'description: a\t# b' '\t# b' -> '# b'; 'x: a\t\t\t# c' '\t\t\t# c' -> '# c'; 'x:\t\ta # c' 'a # c' -> '# c'. New unit test TestAnInlineCommentAfterTabsIsLocatedAtItsHash (4 cases incl. space-tab-space) fails 4/4 with frontmatter.go reverted, passes with it. Checks at 74ee559: cli/ go vet ./... exit 0; go test -count=1 ./... exit 0; scripts/format-check exit 0; scripts/ci.sh exit 0 (RESULT: PASS) after scripts/coverage.sh (go1.24.7) regenerated README (frontmatter 225/203 90.2% -> 245/220 89.8%; total 91.4% unchanged). Measured: specs check <file> --only frontmatter/inline-comment --json, built binaries from 793b724 (before) and 74ee559 (after), on a copy of specs/ + specs.toml with one probe ADR per case.
- **`qqnx.14#1`** (dev-qqnx-3): retro:proof Docs-only bead; the sentence was checked against the binary rather than the prose. evidence: `specs slug-exists ADR-004` → "slug: adr-004 / declared: true / path: specs/adrs/adr-004-module-blueprint-and-atomic-component.md"; `specs slug-exists NOPE-999` → "slug: NOPE-999 / declared: false / path: null". `grep -n "as asked" specs/containers/spec-cli/spec-cli-container.md` → only line 261 (check's "the run was asked for"), no slug-exists sentence. bb-checks.py rc=0. No test applies (pure docs edit).
- **`qqnx.11#1`** (dev-qqnx-6): retro:proof pre-fix failure observed on the new snapshot case. command: cd cli && go test ./cmd -run TestSnapshotCases ; result: --- FAIL: TestSnapshotCases/slug-exists-case-echo; got 'slug: ADR-001-EXAMPLE', want 'slug: adr-001-example'. Negative control slug-exists-case-undeclared (NOPE-999) passed before the fix. Fixed in cli/cmd/slug_exists.go; afterwards the full go test ./... exited 0.

### `environment` — 2

- **`qqnx.13#1`** (dev-qqnx-7): retro:environment go build in a git worktree stamps no VCS settings, so --version shows revision null here evidence: in /home/user/.bb-worktrees/dev-qqnx-7: 'go build -C cli -buildvcs=true -o ../build/specs ./specs' exits 0 and build/specs --version prints revision: null; 'go version -m' shows no vcs.* lines. Same build from /home/user/agent-plugins (main checkout) shows vcs=git, vcs.revision=a22fd79..., vcs.modified=false. AC1 revision check was verified via the main-checkout build, unit tests cover the rendering.
- **`qqnx.10#4`** (dev-qqnx-3): retro:environment Snapshot cases cannot exercise the git half: the harness copies each case input into t.TempDir() (cli/cmd/snapshot_test.go runSnapshotCaseWith), outside any work tree, so every check <path> case now writes the AC-SPECREF-002.13 note to stderr (not asserted, as no check-path case declares checks.err; the note embeds the temp path so it could not be snapshotted anyway). The bead design anticipated the opposite ("so the base read finds the committed fixture"). Git behavior is covered by cli/specs/check_path_e2e_test.go instead. evidence: TestCheckPathOutsideAWorkTreeNotesWhatItSkipped asserts one "note: ...references elsewhere that this document's change broke are not checked" line; snapshot outputs for check-path, check-path-reference-elsewhere, check-path-own-reference-findings unchanged after regeneration.

### kind missing (UNROUTED) — 0

None.

## Routing

| branch | count | findings |
|---|---|---|
| project (`decision`, `gap`, `scope`, `sequencing`, `stale`, `unverifiable`) | 19 | 13 `decision`, 0 `gap`, 1 `scope`, 1 `sequencing`, 3 `stale`, 1 `unverifiable`. The merge test sends `.9#1` (`sequencing`) to swarm instead, because its fix lands in `swarm ready` |
| skill (`tooling`) | 7 | merge test: **6 → swarm** (`qqnx#1`–`#5`, `.1#7`). **1 → project** (`.8#4`: the bead's ACs did not name its `c4_changes` doc, and this repository vends the grooming skill that writes ACs) |
| `proof` (this document) | 15 | as listed |
| `environment` (ledger) | 2 | `.13#1`: a `go build` in a dispatch worktree stamps no VCS settings. `.10#4`: snapshot cases run outside a git work tree. **Neither is admitted to `AGENTS.md`'s ledger by this retro.** `.10#4` is a fact about this repository's test harness, not the container. `.13#1` was measured in one worktree and would need re-measuring before admission |
| unrouted | **0** | none |

`[retros].skill_issues_dir` is unset. `swarm.toml`'s comment there directs
findings against `swarm`-plugin skills to `outbox/swarm/<skill>/`, and the
coordinator relayed the same standing instruction for this run.

## Validation

- **Baseline** (`manifest.checks_baseline`, untouched tree at `4676bc4`): all
  13 declared checks passed, exit 0.
- **The baseline changed under the run, as predicted.** `.8` changed the
  `corpus` check's output text from the 36 `- rule: … outcome: passed` lines to
  a `summary:` block. `.8#1` records this, and records that every check still
  exits 0.
- **At head**: `.17#1`, `.18#1` and `.19#1` record checks at `74ee559`: "cli/
  go vet ./... exit 0; go test -count=1 ./... exit 0; scripts/format-check exit
  0; scripts/ci.sh exit 0 (RESULT: PASS) after scripts/coverage.sh (go1.24.7)
  regenerated README". **This retro did not re-run them.**
- **Coverage**: `c1af88a` regenerated the README figures at integration.
  `fc569da` and `74ee559` regenerated them again for the frontmatter changes
  (both touch `README.md`).
- **CI**: not observed by this retro.

## Pushes

`git fetch origin sprint/specs-qqnx` followed by `git log --oneline
origin/sprint/specs-qqnx..HEAD` printed nothing before this record's commit,
so `origin/sprint/specs-qqnx` is at `74ee559`. This record's commit is local
and not pushed. Tracker pushes belong to the `Stop` hook and were not checked
here.

## What this record could NOT verify

- **Which model served each dispatch**: no dispatch records its model.
- **Effort**: `manifest.effort` reads `unrecorded`.
- **Review pass 3's own account**: there is no amendment comment, and `review_passes` reads 2. That `.20`–`.22` came from a third pass rests on the coordinator's report and the beads' creation by `review-qqnx` at 21:59Z, after pass 2's dispatch closed at 21:51Z.
- **Whether `.20`–`.22` are groomed.** They carry no tier.
- **Head-of-branch checks and CI**: not re-run by this retro (see Validation).
- **`.8#4`'s commit id**: the finding names `c592e7c`. The branch carries the same subject as `c9b3260`, and `c592e7c` still exists as an object. The likely cause is a rebase during accept, but that is **unverified**.
- **`.13#1` as a general fact**: whether every `go build` inside a git worktree omits `vcs.*` stamping was not re-measured.

## Candidates

"Fold" means the issue already exists. This run's finding is added to it and
its recurrence raised. Candidates are ranked by how much each changes the next
run.

| # | candidate | audience → destination | findings | why it changes the next run |
|---|---|---|---|---|
| A | **Fold: accept merges into the checked-out branch, not the recorded sprint branch.** Recurrence 0 → 4 (`c4s2`, `c4model`, `xqm`, `qqnx`) | skill → `outbox/swarm/swarm/2026-10-07-accept-merges-into-the-current-branch.md` | `qqnx#1` | Fifth run where the first accept exits 2 and needs `git branch -f`. Its report claims a fast-forward that did not happen |
| B | **New: accept checks its per-bead signals before it merges and deletes the branch** | skill → `outbox/swarm/swarm/2026-10-08-accept-checks-its-signals-after-the-merge.md` | `qqnx#4` | A failed signal now leaves a merged change and no branch to re-accept. Repair took a reword, an empty commit and a branch recreation |
| C | **Fold: accept rebases a branch held by its worktree.** Recurrence 0 → 4 (`d7w5`, `59ic`, `c4s2`, `qqnx`) | skill → `outbox/swarm/swarm/2026-10-07-accept-rebases-a-branch-held-by-a-worktree.md` | `qqnx#3` | Every concurrent dispatch accepted second fails and needs a hand rebase |
| D | **New: `review_budget` is counted in passes by the tool and in loops by the gate** | skill → `outbox/swarm/swarm/2026-10-08-review-budget-counts-passes-not-loops.md` | `qqnx#5` | The authorized third pass cannot be recorded, so a resumed run would believe a pass is still owed |
| E | **New: `swarm ready` hands out a bead whose dependency's commit is not in the dispatch branch** | skill → `outbox/swarm/swarm/2026-10-08-ready-ignores-whether-a-dependency-commit-is-in-the-dispatch-branch.md` | `.9#1` | The dependent bead is worked against a tree that lacks what it builds on |
| F | **New: `bb-close-bead.sh` pushes the tracker although the dispatch prompt forbids it** | skill → `outbox/swarm/swarm/2026-10-08-close-script-pushes-the-tracker-the-dispatch-may-not-push.md` | `.1#7`, `qqnx#2` | An agent that obeys the prompt has to run the close ceremony by hand |
| G | **Fold: a run resolves no `trailers`.** Recurrence 2 → 4 (`xqm`, `qqnx`) | skill → `outbox/swarm/swarm/2026-10-07-a-run-resolves-no-trailers.md` | `.1#7`, `qqnx#2` | The first close fails `--trailers-from` until the coordinator sets the trailers by hand |
| H | **New: when a bead names a document in `c4_changes` or `c4_docs`, its grooming writes that document's update into an AC.** `.8` missed a container paragraph its `c4_changes` named because the ACs named only two other docs. `.11` and `.8` each left a stale sentence that became a run-introduced amendment (`.14`, `jy7s`) | project → bead against `groom-beads` | `.8#4`, `.11#2`, `.8#3`, `.1#6` | Both amendments admitted outside review, and a follow-up commit, were doc sentences a bead's ACs did not name |
| I | **New: redesign `frontmatter/inline-comment` before working `.20`–`.22` one by one.** Gather `.20`, `.21`, `.22` and the backlog beads `x8oo`, `oeya`, `ihaq`, `kgpq`, `ejcl`, `j3sd`, `8jxr` under one bead. That bead classifies comments from the parsed tree and source columns (the approach `.17` began), and is proved against an edge-case corpus that includes CRLF, tabs, anchors, explicit keys and block scalars | project → one bead consolidating the open defects | `.9#3`, `.15#1`, `.16#1`, `.17#1`, `.18#1`, `.19#1`; review passes 1–3 | Three review passes each found new edge cases in the same lexer heuristic. Patching per finding will keep generating review loops |

**Kept in this record, not candidates:**

- `.1#1` and `.10#2` (sibling criteria left contradictory by grooming): each was resolved in-bead. A later run that hits it again can raise it.
- `.6#1` (`unverifiable`): the 2.5 KB bound sits in the closed bead's acceptance list, measured against an older corpus, and is not in AC-SPECSTR-003.16. No future verifier reads it. A grooming rule against pinning byte sizes to a growing corpus would need a second occurrence.
- `.13#1` and `.10#4` (`environment`): see Routing.
- The missing pass-3 amendment comment: the coordinator recorded no comment, and D covers why `review_passes` could not record the pass.

## What was filed

**Skill audience, A–G: written to `outbox/swarm/swarm/`.** They are 4 new issue
files (B, D, E, F) and 3 folds into existing files (A, C, G). The authority is
the coordinator's dispatch, which relays a standing project instruction that
swarm-tool defects are written to `outbox/swarm/`. No human answered per
candidate in this run. **Note:** at the `specs-xqm` retro, Brett declined the
swarm outbox candidates. Whoever forwards `outbox/swarm/` should know that
these went in under the standing instruction and not under a per-run yes. Each
file carries `model: "<withheld by session policy>"`.

**Project audience, H and I: pending.** Nothing was filed in the tracker. They
are decidable later from this record.

`retro:filed:<path>` comments were added to the epic and to `.1` and `.9`,
which hold the findings A–G rest on.

## Counts

| | |
|---|---|
| harvested | 43 |
| routed — project | 19 (+1 `tooling` that the merge test sends to the project, −1 `sequencing` it sends to swarm) |
| routed — skill | 6 `tooling` + 1 `sequencing` by the merge test, all swarm |
| routed — `proof` (this document) | 15 |
| routed — `environment` | 2 (neither admitted to the ledger) |
| unrouted | 0 |
| candidates | 9: 3 folds (A, C, G), 4 new skill issues (B, D, E, F), 2 project (H, I) |
| filed | 7 (A–G, `outbox/swarm/swarm/`) |
| declined | 0 |
| pending | 2 (H, I) |
