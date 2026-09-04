# Run record — `agent-plugins-d7w5`, spec corpus structure extraction

*Written by `retro-d7w5` on 2026-09-04. One document: the run's results, every
finding it captured, and what it could not establish.*

## The run

| | |
|---|---|
| run epic | `agent-plugins-d7w5` — "Swarm run — Spec corpus structure extraction" |
| target | `agent-plugins-f6tl` — "Spec corpus structure extraction", adopted whole |
| sprint branch | `sprint/agent-plugins-d7w5`, merged into `claude/swarm-agent-plugins-5wxrfo` |
| `base_sha` → head | `70dddef` → `1bb35e7` |
| queue | 7 beads under the target. **5 admitted at the gate** (`.1 .2 .3 .4 .7`), 2 excluded; AMENDMENT 1 dropped `.7`, leaving **4** |
| dispatches | one — `dev-d7w5-1`, batch `f6tl.1 → .2 → .3 → .4` |
| partition | strict dependency chain, **zero parallelism, zero collision edges** |
| declared checks | 12 — `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| model lineup | `advanced` tier → `opus` per `swarm.toml` `[models]`. **The manifest's `models` key is unset** and `effort` reads `unrecorded` — see `d7w5.2.1#3` |
| gate | `agent-plugins-987r` — one authorization gate, taken once |

**Standing authorization** (`manifest.standing_auth`), six rows:

| row | value | source |
|---|---|---|
| `pull_request` | `none` | recorded `default`; **actually the human's choice against a skill default of `draft`** — see `d7w5.2.1#1` |
| `move_foreign_beads` | `false` | default |
| `admit_review_findings` | `false` | default (the human declined it explicitly when offered) |
| `tier_escalation` | `once` | default |
| `review_budget` | `2` | default |
| `merge_to_default` | `false` | default |

`dispatch_failure_threshold` resolved to `3` at the gate and **is not in the
manifest at all** — the model has no field for it (`d7w5.2.1#2`).

## Results

All four admitted beads closed, each in one commit, on one dispatch.

| bead | status | agent | commit | what landed |
|---|---|---|---|---|
| `f6tl.1` | closed | `dev-d7w5-1` | `e4daa8c` | Declared the index model's `headings` and `record_blocks` structure properties in Strata and regenerated |
| `f6tl.2` | closed | `dev-d7w5-1` | `50f8835` | Built the heading tree, section spans and heading identifiers in `cli/core/markdown` |
| `f6tl.3` | closed | `dev-d7w5-1` | `1cf26e0` | Populated every member's headings during the corpus walk, from the body offset |
| `f6tl.4` | closed | `dev-d7w5-1` | `ef94bd1` | Added the `headings` query command with `--json` and `--emit` |

- 4 completed, 0 blocked, 0 skipped, **0 tier escalations**.
- `manifest.consecutive_dispatch_failures` = 0; the circuit breaker never armed.
- Pre-merge acceptance check: **4/4 exit 0**. `swarm dispatch accept`: merged
  true, ancestor true, worktree removed, `beads_checked` 4,
  `acceptance_failures` `[]`.
- Diff `0f8d8b7..1bb35e7`: **74 files, +3288/−77** (measured; the wider range
  `70dddef..HEAD` including the `.gitignore` commit is 75 files, +3291/−77).

## What the run decided

Seventeen of the 28 findings are the four work beads’ own — and **eleven of them
are `decision`**: a design question the work order or the blueprint left open
and the implementing agent had to settle. That is the run's dominant signal.
The full text of each is below; what they decided, in one line each:

| finding | what was decided, that a document should have pinned |
|---|---|
| `f6tl.1#1` | `RecordBlock.content` is `[String: String]` — Strata is closed-shape and has no arbitrary-value type. Residual limit: a tag registering non-scalar values cannot be represented |
| `f6tl.1#3` | `RecordBlock.position` covers the fenced block **including** its fences, so a promotion rewrite does not orphan them |
| `f6tl.1#4` | `Heading` carries **no** identifier property; it is a pure function of the text. The cost lands on the CMD item, which re-derives per match |
| `f6tl.2#1` | `Identifier` renders inline content through goldmark rather than filtering raw text. The two readings diverge only on markdown links, of which the corpus has zero |
| `f6tl.2#2` | A setext heading's `Position` covers its underline line, so `Position` ends exactly where `Section` begins |
| `f6tl.2#4` | `Identifier` is a package function, not a field — the MODEL item's AC list had already excluded it |
| `f6tl.3#3` | A heading carries `RecordBlocks` as `[]`, never null, so the generated reader accepts the index |
| `f6tl.4#1` | `--emit` writes only the outermost matched section; a nested one is a byte-identical repeat |
| `f6tl.4#2` | The YAML renderer is hand-written over the same struct `encoding/json` marshals, not `goccy/go-yaml` configured — parity is by construction |
| `f6tl.4#3` | The answer carries only `matches`; the selection is not echoed back. An empty result is `matches: []` |
| `f6tl.4#5` | `--json` and `--emit` are rejected together rather than one silently winning, following `feature`'s existing rule |

**Three `scope` findings, and they are one shape**: the work order's declared
surface was smaller than the change it required.

- `f6tl.1#2` — a "declare in Strata and regenerate" bead had to touch the
  emitter, two hand-maintained fixtures and three unit tests. Anchors named
  three paths, none of them those.
- `f6tl.3#2` — delivering a wiring-only item required adding `HeadingsFrom` to
  `cli/core/markdown`, which is the **previous** item's surface.
- `f6tl.4#4` — landing the command made `cli-output-standards.md`'s own status
  text false, so a standards document had to change with it; `c4_docs` named
  only the container doc.

Review independently filed the third of those as `d7w5.15`, framed as a
collision-graph risk: `c4_docs` is what the partitioner serializes on, so an
under-declared surface cannot be serialized. **This run had zero parallelism,
so it cost nothing here.**

**Two `gap` findings, and both predicted where review then found defects.**
`f6tl.3#1` (frontmatter's closing `---` is a setext heading, so the walk must
start at the body offset) and `f6tl.2#3` (a heading inside a blockquote or list
item gets a document-coordinate section span carrying quote markers) name the
two areas review returned `d7w5.7`, `d7w5.9` and `d7w5.12` against.

There were **no `sequencing` findings**, consistent with a strict-chain
partition run behind one agent.

## Scope

The run finished its **admitted** set. It did so over a scope narrowed twice.

| bead | disposition | reason |
|---|---|---|
| `f6tl.5` | excluded, deferred | Stage 2: depends on the record-block tag registration `FEAT-SPECSTR-004` specifies in `frd-spec-structure-conformance`, which is **not yet a work order**. Excluded by the human at the gate |
| `f6tl.6` | excluded, deferred | Stage 2: a registration point no profile fills leaves it inert, per its own description. Excluded by the human at the gate |
| `f6tl.7` | admitted, then **dropped** | AMENDMENT 1: unreachable once `.5` and `.6` were excluded — `bd` counts a **deferred** dependency as blocking, and three of its ACs require stage 2 landed. Decided by `coordinator-d7w5` at groom & plan, not escalated; deferred and labelled `human` |

`manifest.moved` is `[]` — no bead was moved from another epic
(`move_foreign_beads=false`).

**The target epic `agent-plugins-f6tl` is still open, and that is correct.**
Closing it would report the feature complete while its stage-2 deltas are
unlanded. `f6tl.7` carries the bounce record and two ways a human can clear it
(admit stage 2, or split a stage-1-only DoD verification out of it).

**Review filed 9 findings, and this run admitted none of them** —
`admit_review_findings=false`, authorized at the gate. `d7w5.7` … `d7w5.15`
(4 P1, 4 P2, 1 P3; 6 in scope, 3 out) are all deferred. `review_passes` = 1 of
3: **the budget is unspent, not exhausted.** The loop did not re-enter because
its condition is findings *admitted*, and zero were.

## Every finding, verbatim

28 findings across 5 beads. Counts verified against each bead's
`comment_count` before aggregating: `f6tl.1` 5, `f6tl.2` 4, `f6tl.3` 3,
`f6tl.4` 5 (17, every comment a finding); `d7w5.2.1` 12 comments of which 11
are findings and one is the gate record.

### `decision` — 11

**`f6tl.4#1`**

> retro:decision --emit skips a section already contained in one just written. The two ACs it sits between do not settle this and, read literally, produce duplicated output.
> evidence: AC-SPECSTR-003.4 requires a matched heading to be returned with every heading nested inside its section, and AC-SPECSTR-003.5 requires the emit form to return 'the text of each match's section'. Applied literally to a match on a heading with children, the parent's span already covers the children's text, so each nested section is written twice. Measured before the fix, on the headings-emit snapshot case ('headings adr-001-first#Decision --emit'): the expected artifact contained 'The part that only matters once you have decided.' twice — once inside Decision's section and once as Detail's own. Decided to write only the outermost matched section, since what is skipped is always a byte-identical copy of text already emitted, and this query exists to spend a reader's attention once on the text it selected (the FRD: 'I review only the normative text of many documents instead of reading them whole'). Rejected: emitting every span faithfully, whose cost is a reader paying twice for the same passage and having no way to tell the repeat from a real repetition in the document. Recorded in specs/containers/spec-cli/spec-cli-container.md so the behaviour is not read as a bug later.

**`f6tl.4#2`**

> retro:decision The YAML renderer is hand-written and reflects over the same struct encoding/json marshals, rather than being goccy/go-yaml (already a module dependency) configured into shape.
> evidence: cli-output-standards' YAML Style section fixes every choice a library would offer — block style only, two-space indent, '[]' and '{}' for empty collections, block scalars trimmed at the edges with '|-', plain scalars only where they round-trip, no comments, anchors, aliases, tags or document markers, no folding. It also requires that 'Both renderings SHALL carry the same values in the same order', which reflecting over the json tags makes true by construction: a field cannot appear in --json and not in the YAML. A library would have to be constrained into all of that and still could not give the second guarantee. Rejected: goccy/go-yaml with options, whose cost is a larger surface making the same promises less legibly and no parity guarantee. The renderer is ~250 lines in cli/cmd/yaml.go with cli/cmd/yaml_test.go pinning each style rule, and every case parses its own output back through goccy/go-yaml so the tests check what the bytes MEAN and not only how they look.

**`f6tl.4#3`**

> retro:decision The answer carries only 'matches' — the selection is not echoed back — which differs from the nearest sibling on the surface.
> evidence: cli-output-standards: 'A command's answer SHALL be a YAML mapping whose keys are the properties the command established ... and SHALL contain nothing else. No header naming the command — the caller typed it.' refs does echo its input (refsResult carries Slug and Direction), but refs is listed in that document's own Current conformance table as not conforming, so it is not the precedent to copy. An empty result is therefore 'matches: []', which satisfies AC-SPECSTR-003.6's 'reports that and exits zero' without a message: the empty collection IS the report, per 'An empty answer is an answer'.

**`f6tl.4#5`**

> retro:decision --json and --emit are rejected together rather than one silently winning.
> evidence: The item specifies both flags and does not say what their combination means. --emit writes a document payload as itself, which by cli-output-standards has no envelope and therefore no JSON form to select, so the pair names no output. feature already draws the same line for the same reason — cli/cmd/feature.go returns 'spec feature: --code and --json cannot be combined with --emit' — so this follows an established rule on the surface rather than inventing one. Rejected: letting --emit win, whose cost is a caller passing --json and getting bytes no JSON parser accepts, with nothing to tell them why.

**`f6tl.2#1`**

> retro:decision Identifier renders inline content through goldmark rather than filtering the raw text; the blueprint says 'its rendered inline content lowercased' but the two readings only diverge on links, so the cheaper one would have passed unnoticed.
> evidence: The spec's own character rule ('every character that is not a letter, digit, hyphen, underscore or space removed') already strips backticks, asterisks and wikilink brackets, so for 'The `Profile` interface' and 'Corpus in [[module-document-core]] (modified) — IDX' a filter-only implementation and a rendering one agree exactly. They disagree on a markdown link: measured with python over the spec's rule, filtering 'See [the guide](http://example.com/guide)' gives 'see-the-guidehttpexamplecomguide' — the URL leaking into the identifier — where rendering gives 'see-the-guide'. Measured the corpus: 623 headings under specs/, of which 0 contain a markdown link and 9 contain backticks, so nothing in the tree distinguishes the two today and a filter-only implementation would have shipped green. cli/core/markdown/headings_test.go pins the link case as a negative control so a later simplification to filtering fails rather than silently changing anchors.

**`f6tl.2#2`**

> retro:decision An underlined (setext) heading's Position covers its underline line as well as its text line. The item says 'the span of the heading line itself', which does not say where a heading written across two lines ends.
> evidence: Decided so that Position ends exactly where Section begins in both written forms — otherwise the underline would fall inside the section, and a caller slicing a section back out of the body would get '==========' as the first line of the content. Measured on goldmark: an underlined heading's Lines() segment covers only the text line (probe on 'Setext H2\n---------' returned seg (66,75) "Setext H2" at level 2), so the underline has to be consumed deliberately; cli/core/markdown/headings.go does that via the prefixed() check on the first non-blank byte of the heading line. Pinned by TestAnUnderlinedHeadingCarriesItsUnderlineInThePositionSpan. No document in specs/ uses the form today, so this is unexercised by the corpus.

**`f6tl.2#4`**

> retro:decision Identifier is exposed as a package function taking a text, not as a field on Heading, and the module blueprint now records why.
> evidence: The item's AC says 'The heading identifier of a text is exposed, is a pure function of it', and the preceding MODEL item's AC list gives Heading exactly text, level, heading-line span and section span with no identifier — so storing one would have contradicted the model that had just landed. specs/modules/document-core/document-core-module.md now states 'It is exposed as a function rather than recorded on a Heading, and the index model carries no identifier property for the same reason: it is recomputable from the text and is therefore not state.' The cost lands on the consumers rather than here: the CMD item returns an identifier per match (AC-SPECSTR-003.1) and anchor resolution matches on one, so both derive per call instead of reading a stored value.

**`f6tl.3#3`**

> retro:decision A heading carries RecordBlocks as an empty collection during the walk, not null, even though no profile registers a record-block tag yet and the REC item that fills them has not landed.
> evidence: Required by the model the MODEL item declared: record_blocks is a required list on both Heading and IndexEntry, so a null marshals into an index the generated reader rejects. Consistent with the rule cli/core/specindex/specindex.go:35 already states — 'Every collection is non-nil so that an index containing no members marshals its collections as [] rather than as null ... the emitted index is compared byte for byte across platforms'. Visible in the regenerated snapshots: every heading in cli/testdata/cases/index-default/expected/manifest.json carries "record_blocks": []. Pinned by TestAMemberWithNoHeadingCarriesAnEmptyCollection, which asserts non-nil rather than merely empty.

**`f6tl.1#1`**

> retro:decision RecordBlock.content is typed [String: String]; the bead and the blueprint say only 'its parsed content' and Strata has no arbitrary-value type, so the type was mine to pick.
> evidence: .claude/skills/strata/references/strata-llm-reference.md:16 — 'Closed shapes only. No additionalProperties: true, no untagged unions, no Bytes, no Any'; and its porting table — 'additionalProperties: true / open-shape records | Not supported. Strata is closed-shape'. Measured what the corpus actually holds: awk over every fenced 'component' block in specs/ gives 13 blocks / 1991 bytes, each exactly title+slug+description, all scalar strings — matching the blueprint's own 'thirteen blocks carry 1,991 bytes'. Rejected: a recursive JSON-value union. It is expressible (language-reference.md:707 'Union Recursion') but every union in this model carries a discriminator, so parsed content would reach specs.json as tagged variants rather than as the plain data a catalogue reads — the cost being an index no consumer can read an element declaration out of without unwrapping. The residual limit is that a tag registering a format with non-scalar values (BlockFormat admits json and toml) cannot be represented; the property's doc block in schemas/index.strata states this so the REC item cannot miss it.

**`f6tl.1#3`**

> retro:decision RecordBlock.position covers the fenced block INCLUDING its opening and closing fences; the bead says only 'the span of the block as written', which does not settle it.
> evidence: Decided on ADR-008's promotion mechanic — 'Promotion lifts the block verbatim into the new document's frontmatter and replaces it with the link, in the same change' — so a rewrite addressing only the content would leave the fences orphaned. Follows the precedent already in this model: schemas/feature.strata:104 FeatureSourceBlock.block is 'The extent of the fenced block, opening and closing fences included', recorded because 'a caller that replaces or extracts the whole construct needs the fence too'. Rejected: an inner span over the content alone, which would have cost every promotion rewrite a second derivation of the fence extent.

**`f6tl.1#4`**

> retro:decision Heading carries NO identifier property, and I recorded that as a rule in the module blueprint rather than leaving it unstated.
> evidence: The bead's AC lists exactly 'text, level, heading-line span and section span', and the blueprint's HEAD delta makes the identifier the component's — 'The derivation SHALL be a pure function of the text, deriving no positional suffix for a repeated heading' — so it is recomputable and not state. But CMD's AC-SPECSTR-003.1 returns the identifier on every match, so a later agent may reasonably expect to read it off the index; specs/modules/document-core/document-core-module.md now says 'A heading carries no identifier, because the identifier is a pure function of the text and is derived by the Markdown Structure component on demand' so the omission reads as a decision rather than a gap. Flagging because the cost lands on CMD, not here: the query derives an identifier per heading per run instead of reading one.


### `scope` — 3

**`f6tl.4#4`**

> retro:scope Landing this item made cli-output-standards' own status text false, so the standards document had to change with it — the item's c4_docs names only spec-cli-container.md.
> evidence: The document read 'Status: approved, and not yet met' and its conformance table 'No command conforms today; the surface predates these standards.' Both become untrue the moment a conforming command exists, and the same document says 'A command specified after this document is written to conform from the start' — so the falsehood was built into it from the day it was written. Updated the status line, the table preamble and added a headings row rather than leaving a document that contradicts the tree; nothing else in it changed, and no other command's row was touched since none of them moved.

**`f6tl.3#2`**

> retro:scope Delivering this item required adding a function to cli/core/markdown (HeadingsFrom), which is the HEAD item's surface, not this one's. The item scopes itself to wiring — 'Corpus holds no parsing logic of its own' — and anchors only corpus.go and corpus.Open.
> evidence: Two constraints met in the same place. Headings must be recorded in document coordinates, as every other index position is, and the search must skip the frontmatter. Slicing the body in Corpus and shifting offsets afterwards would not work: span.Of derives line and column from the start of the slice it is given, so a body-relative slice yields a heading on line 1 of a file where it is on line 10, and re-deriving them in Corpus would put position logic in the one component the module says computes none. HeadingsFrom parses the body slice but calls span.Of against the whole source, which keeps both rules. The alternative — widening the HEAD item retroactively — was rejected because it had already closed; the change is additive (Headings is now HeadingsFrom at offset zero) and both the module blueprint's Markdown Structure and Corpus sections record it.

**`f6tl.1#2`**

> retro:scope Declaring the two properties as REQUIRED lists forced changes outside 'declare in Strata and regenerate' — the emitter, two hand-maintained input fixtures and three unit tests — none of which the bead's Design or Anchors anticipate.
> evidence: The bead's Design is 'Declare the shapes in schemas/index.strata and regenerate', Anchors name only schemas/index.strata, cli/core/schema/spec-model.schema.json and cli/core/specmodel. Measured after regenerating, before any further edit — 'cd cli && go test ./...' rc=1 with three distinct failures: (a) specindex_test.go:166 and :203 'unmarshal: IndexEntry: required field "headings" is missing or null' from entries built as Go literals and round-tripped; (b) cache_test.go:387 'Open rebuilt an index whose inputs are unchanged', because the planted entry marshalled headings as null and the warm read then rejected its own file; (c) portability_test.go:71 index-default, the emitted index carrying '"headings": null, "record_blocks": null'. Required-list was the right call — it is exactly how 'references' and 'features', the first two exercises of the additive-at-1.0 rule, are declared on SpecIndex, and the blueprint states 'No backward-compatibility window applies' — but the consequence is that a MODEL-only bead cannot land green without touching the emitter. The fix mirrors the rule cli/core/specindex/specindex.go:35 already states for the index's own collections ('Every collection is non-nil so that an index containing no members marshals its collections as [] rather than as null'), applied centrally in builder.addEntry so all four IndexEntry construction sites in corpus.go are covered.


### `gap` — 2

**`f6tl.2#3`**

> retro:gap A heading nested inside a blockquote or a list item is reported like any other, and its section span is computed in document coordinates — neither the item nor the HEAD delta says whether it should be.
> evidence: The blueprint says only 'every heading of a document body in document order'. Measured against the implementation on body '# Top\n\n> ## Quoted\n>\n> text\n\n- item\n  ## In list\n\n## Real\n\nend\n': four headings are returned, and the blockquoted one's section slices to '> text\n\n- item\n' — carrying the '>' markers and running past the end of the blockquote into the list that follows. That is the honest consequence of section spans being document ranges rather than container-relative ones, but it means a caller emitting such a section (the CMD item's AC-SPECSTR-003.5) would print quote markers. Left as-is rather than special-cased: measured 0 headings inside a blockquote or list item across specs/ (the 3 grep hits are indented YAML comments inside fenced blocks, which goldmark correctly does not treat as headings), so nothing in the corpus exercises it and inventing a rule with no instance to check it against would be guessing.

**`f6tl.3#1`**

> retro:gap A whole-file heading search reports the entire frontmatter block as a level-2 heading, so the walk has to start at the body offset. The IDX delta calls this 'a wiring change, not a second parser' and its Anchors name only cli/core/corpus/corpus.go — the hazard appears nowhere in the item.
> evidence: A frontmatter block's closing '---' underlines the YAML above it, which is a setext heading in CommonMark. Measured by running the heading extraction over a real corpus file, specs/adrs/adr-008-element-record-blocks.md: 7 headings found, of which the first is level=2 with text 'title: "ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model"\ndescription: Un-promoted elements carry a fenced record block...\nslug: adr-008\ntype: adr\nstatus: active\nupdated_at: 2026-08-05T00:00:00Z' — the whole block. The blueprint does carry the fact, but only as a relied-upon element ('Frontmatter — the body offset a heading search starts from'), which reads as a convenience rather than as the thing that makes the output wrong. It also cuts against the nearest precedent: cli/core/corpus/features.go documents that 'The whole of src is offered to the fenced-block lookup, frontmatter included, so that every span a block yields is already in the document's own coordinates', which is safe for fences and not for headings. Pinned by TestASearchFromTheBodyOffsetLeavesTheFrontmatterBehind, whose negative control asserts the whole-file search still reports the frontmatter, so the two searches cannot be silently unified later.


### `proof` — 1

**`f6tl.1#5`**

> retro:proof The regenerated snapshots change every index case's input_hash even though no corpus file changed; verified this is by design rather than accepting the diff.
> evidence: git diff on cli/testdata/cases/index-default/expected/manifest.json shows input_hash sha256:3ae7797c... -> sha256:f9bb75b7... alongside the expected per-entry '"headings": []'. cli/core/corpus/hash.go's sum() folds in 'field(sum, "profile-schema", opts.Profile.Schema())' and its doc block states 'the active profile's schema and per-type table decide which candidates are members ... A change to any of them changes the index while every file on disk stands still, so all of them are folded in.' The profile's schema IS the regenerated cli/core/schema/spec-model.schema.json, so a model change necessarily moves the hash. Worth knowing for review: every future MODEL delta will rewrite input_hash in all five index snapshots, and that is correct, not snapshot churn to suppress.


### `tooling` — 11

**`d7w5.2.1#1`** · target `swarm`

> retro:tooling:swarm — swarmmodel.AuthSource admits only "swarm.toml" and "default", so a row the HUMAN changed at the authorization gate cannot be recorded truthfully. This run's pull_request resolved to the skill default "draft" and the human chose "none"; the manifest now reads value=none source=default, which is false provenance. The gate exists for a human to change rows, so the enum needs a third member (e.g. "human").

**`d7w5.2.1#2`** · target `swarm`

> retro:tooling:swarm — swarm-scope's gate table lists seven rows including dispatch_failure_threshold, but swarmmodel.StandingAuth declares only six fields and has no home for it. The threshold is resolved at the gate and then passed to `swarm ready --threshold`, so a resumed coordinator cannot recover the authorized value from the manifest and silently falls back to the flag default of 3.

**`d7w5.2.1#3`** · target `swarm`

> retro:tooling:swarm — pre-flight-check.sh never writes two keys its own stage requires. Measured: `grep -c ledger pre-flight-check.sh` -> 0, and the only two manifest_merge calls (lines 764, 807) write checks_baseline/checks_baseline_status/checks_ids and concurrency_cap. Its header at line 21 claims it writes 'the model lineup' to the manifest, but no 'models' key exists on the run epic. Consequence: `swarm manifest validate` reports models and ledger absent forever, and because stage is derived from the keys present, this run's stage is pinned at 'pre-flight' even with branch, base_sha, admitted and standing_auth all written and two beads already committed.

**`d7w5.2.1#4`** · target `swarm`

> retro:tooling:swarm — `swarm manifest validate` renders SeverityPending as ok:false, so a healthy pre-dispatch manifest reports 6 failures of which 4 are just 'not written yet; becomes required at the retro stage'. A caller routing on ok cannot tell a pending key from a broken one, which is the distinction the RuleNotYetWritten severity exists to draw.

**`d7w5.2.1#5`** · target `swarm`

> retro:tooling:swarm — `swarm manifest set` type errors name the Go type but not the shape it accepts, costing 3 rejected writes on 4 attempts. 'excluded takes []swarmmodel.Exclusion, and the value given is not one: json: cannot unmarshal string into Go value of type map[string]json.RawMessage' does not say the element is {id, reason}; I had to read internal/swarmmodel/swarmmodel.go. Same for AuthSource's two-member enum. The atomic refusal ('storing none of it') is exactly right; only the error text is short.

**`d7w5.2.1#6`** · target `swarm`

> retro:tooling:swarm — there is no way to list the declared manifest keys. I discovered 'target' (having guessed 'target_epic') by running `swarm manifest validate` against an empty manifest and reading the key names out of its failure lines. A `swarm manifest keys` subcommand, or naming the valid keys in the 'not a declared manifest key' error, would remove the guess.

**`d7w5.2.1#7`** · target `swarm`

> retro:tooling:swarm — `swarm phase close` cannot close a childless ceremony epic. The pre-flight phase epic has no children by construction, so it reported 'the tracker does not report this epic eligible, and no open child was found among its own' and stayed open; I closed it with `bd close` by hand. A phase epic left open blocks the chain, which is the failure mode the command exists to prevent.

**`d7w5.2.1#8`** · target `swarm`

> retro:tooling:swarm — `swarm status` reports ceremony beads as scope additions. scope_additions listed agent-plugins-d7w5.2.1 (the run's own authorization gate task, closed) alongside the three genuinely-excluded beads. A ceremony bead is structurally never in manifest.admitted, so it will always read as creep and trains the reader to ignore the signal.

**`d7w5.2.1#9`** · target `swarm`

> retro:tooling:swarm — a retro that harvests findings from `bd show --json` will find none of them. Measured on bd 1.2.2 (dev): `bd show agent-plugins-d7w5.2.1 --json | jq '.[0].comments'` -> null while the same object reports comment_count: 9 and the text view prints all nine bodies. The count says the findings exist and the array does not carry them, so a JSON-based harvest silently returns empty rather than erroring. swarm-retro must read the text view, or bd must populate the array.

**`d7w5.2.1#10`** · target `swarm`

> retro:tooling:swarm — swarm-review parents its findings under the RUN epic, which puts unadmitted work straight into the run's own dispatch set. Measured: after pass 1, `swarm ready --run agent-plugins-d7w5` reported ready_count 9, every one a findings bead with dispatch "" and held false, despite standing_auth.admit_review_findings=false. Only the coordinator's manual diff of the ready set against manifest.admitted caught it; an orchestrator that auto-chained on the ready set would have dispatched all nine against the gate's explicit authorization. Findings should be filed against their origin outside the run epic, or filed deferred, rather than relying on every coordinator to notice.

**`d7w5.2.1#11`** · target `swarm`

> retro:tooling:swarm — the phase chain deadlocks whenever the target epic legitimately cannot close. Measured: `bd close agent-plugins-d7w5.4` -> 'cannot close agent-plugins-d7w5.4: blocked by open issues [agent-plugins-f6tl] (use --force to override)'. f6tl is open only because .5 .6 .7 are deferred, and closing it would falsely report the feature complete — the false-complete failure swarm-scope explicitly warns about. So review, retro and post-flight each need --force to close, on a run that did everything right. A run whose target is a partially-scoped feature epic is the normal case, not the exception, and `swarm phase close` has no path through it.

## Routing

Routed by the kind recorded at capture, not by reading the finding.

| kind | n | branch |
|---|---|---|
| `decision` | 11 | **project** — a bead in this tracker |
| `scope` | 3 | **project** |
| `gap` | 2 | **project** |
| `proof` | 1 | **terminates here.** Recorded beside what it supports; no issue |
| `tooling`, target `swarm` | 11 | **skill** — the `swarm` plugin (`brettcallaghan/swarm`), which this repository does **not** vend. `.sf/outbox/swarm/` |
| `environment` | 0 | — |
| `stale`, `unverifiable`, `sequencing` | 0 | — |
| **unrouted** | **0** | every finding carried a parseable kind |

**Zero unrouted.** The taxonomy covered everything this run hit.

`swarm.toml`'s `[retros] skill_issues_dir` is deliberately unset, and its
comment carries the split this routing rests on: this repository vends the `sf`
skills (`spec`, `groom-beads`, `groom-feature-blueprint`, `retro-spec`,
`retro-skill`), so a finding against one of those is an ordinary bead here; it
no longer vends the swarm plugin, so a finding against `swarm` cannot be acted
on here and goes to `.sf/outbox/swarm/` for delivery.

**All eleven `tooling` findings target the swarm plugin.** None targets an `sf`
skill. So the outbox is this run's entire skill-audience output, and this
repository's own tracker takes nothing from that branch.

## Validation

- **Baseline** (`manifest.checks_baseline`, untouched tree): all 12 declared
  checks passed, exit status 0.
- **At head**: all 12 pass on the merged tree — `python3 <swarm scripts>/bb-checks.py` (recorded with the placeholder `{SCRIPTS}` unresolved on `d7w5.4`; the run resolved it to `/root/.cache/swarm-src/plugin/skills/swarm/scripts`)
  → rc=0, recorded on `agent-plugins-d7w5.4` by the reviewer against the
  merged tree at review pass 1, and reported again by the coordinator. **Same
  12, same result as the baseline.** This retro did **not** re-run them — the
  claim is inherited from those two records, not measured here, and the tree
  gained one commit (`1bb35e7`, the coverage regeneration) after review.
- **Coverage** — not a declared check, but `AGENTS.md` requires it for a
  `cli/**` change and `cli.yml`'s `coverage` job enforces it. Regenerated by
  the coordinator: **89.5% → 89.1%**, +313 statements. `cli/core/markdown` went
  37 → 96 statements and 100% → 93.8%. `scripts/coverage.sh --check` exits 0,
  so the committed README figures match a fresh measurement.
- **CI was not observed.** The branch is pushed; no workflow run was read back.

## Pushes

- `git log origin/claude/swarm-agent-plugins-5wxrfo..HEAD` is **empty** —
  nothing is stranded on the designated branch. Head is `1bb35e7`.
- **`sprint/agent-plugins-d7w5` was never pushed, deliberately.** Only the
  designated branch may be pushed in this environment, so each dispatch was
  accepted with `--no-push` and merged by hand. The sprint branch's tip
  (`ef94bd1`) is an ancestor of the pushed head, so nothing is lost by that;
  `git log origin/claude/swarm-agent-plugins-5wxrfo..sprint/agent-plugins-d7w5`
  is empty.
- `merge_to_default=false` and `pull_request=none`, both authorized: **no pull
  request was opened and nothing reached the default branch.** Whatever happens
  to this branch is a human's next move, not the run's.

## What the run could NOT verify

Named, one per line. An unstated gap is indistinguishable from one that does
not exist.

- **Whether the 4 P1 review findings are reachable in production or latent.**
  The reviewer established reachability for `d7w5.7` specifically and marked
  `d7w5.11` latent-but-specified. `d7w5.8` and `d7w5.9` were **not probed**
  either way. All four are deferred, so nothing downstream has settled it.
- **The batch agent's finding count.** Its return claimed 21 `retro:` comments
  across the four work beads. Measured here from `bd export`: **17** — `f6tl.1`
  5, `f6tl.2` 4, `f6tl.3` 3, `f6tl.4` 5, each exactly matching that bead's
  `comment_count`, and every comment on those four beads is a finding. Adding
  the coordinator's 11 gives 28, which is not 21 either. **21 matches nothing
  measurable from the tracker.** The batch agent's transcript is gone, so what
  it counted cannot be recovered; this is recorded as unreconciled rather than
  resolved in favour of 17.
- **`models` and `effort` for this run.** `manifest.models` is unset and
  `effort` reads `unrecorded`, because `pre-flight-check.sh` never writes
  either key (`d7w5.2.1#3`). The dispatch tier is recoverable from
  `partition_preview` — advanced, which `swarm.toml` maps to `opus` — but the
  **coordinator's own** model and the run's effort setting are not recorded
  anywhere and are not recoverable from the epic.
- **The true provenance of `pull_request=none`.** The manifest says
  `source: default`. It was the human's choice at the gate, against a skill
  default of `draft`. The schema has no `human` source (`d7w5.2.1#1`), so the
  manifest is false here and only the gate comment on `d7w5.2.1` says so.
- **The authorized `dispatch_failure_threshold`.** Resolved to 3 at the gate,
  passed to `swarm ready --threshold`, and stored nowhere. A resumed
  coordinator would fall back to the flag default and could not tell that it
  had (`d7w5.2.1#2`).
- **`bd human respond` failing with `storage is nil` on bd 1.2.2.** Asserted in
  `f6tl.7`'s bounce comment and **not re-measured here** — the command comments
  *and* closes, so probing it would corrupt a bead the human still has to
  decide. Settle it on a throwaway bead.
- **Whether CI is green on `1bb35e7`.** Checks pass locally; no workflow run was
  read.
- **Whether the 12 declared checks pass at `1bb35e7` specifically.** The green
  result is inherited from the reviewer's and the coordinator's records, both
  taken before or at the coverage commit; this retro deliberately does not
  re-run checks. Settle it with `python3 <swarm scripts>/bb-checks.py` on the
  current tree.

## Candidates

Reasoned from the findings above, deduped against what is already filed, and
ranked. **Nothing here has been filed.** Each is put to a human once; a
declined candidate stays in this document, where a later run hitting the same
friction raises it again with a higher recurrence count.

Provenance carried by every candidate: run `agent-plugins-d7w5`, branch
`claude/swarm-agent-plugins-5wxrfo`, record this file, model `opus` (advanced
tier per `swarm.toml` `[models]`) for the work-bead findings; the coordinator's
model is **unrecorded** — see the could-not-verify list.

### Skill audience → `.sf/outbox/swarm/` (the swarm plugin)

Ten candidates from eleven `tooling` findings. One finding deduped out.

| # | candidate | findings | rank |
|---|---|---|---|
| S1 | **`swarm-review` files findings under the run epic, putting unadmitted work in the dispatch set.** File them against their origin outside the run epic, or file them deferred, instead of relying on each coordinator to diff the ready set against `manifest.admitted` | `d7w5.2.1#10` | **P0** |
| S2 | **`pre-flight-check.sh` never writes `models` or `ledger`, so `manifest validate` reports them absent forever and the derived stage stays pinned at `pre-flight`.** Its own header claims it writes the model lineup | `d7w5.2.1#3` | **P0** |
| S3 | **`swarm phase close` has no path through a target epic that legitimately cannot close.** A partially-scoped feature epic is the normal case; review, retro and post-flight each needed `--force` on a run that did everything right | `d7w5.2.1#11` | **P0** |
| S4 | **`swarmmodel.AuthSource` admits only `swarm.toml` and `default`, so a row the human changed at the gate cannot be recorded truthfully.** Add a third member | `d7w5.2.1#1` | P1 |
| S5 | **`StandingAuth` has six fields for a seven-row gate; `dispatch_failure_threshold` has no home.** A resumed coordinator silently falls back to the flag default | `d7w5.2.1#2` | P1 |
| S6 | **`swarm phase close` cannot close a childless ceremony epic**, which every pre-flight phase is by construction | `d7w5.2.1#7` | P1 |
| S7 | **`swarm status` counts ceremony beads as scope additions**, so the creep signal is always dirty and trains the reader to ignore it | `d7w5.2.1#8` | P1 |
| S8 | **`swarm manifest validate` renders `SeverityPending` as `ok:false`**, so a caller routing on `ok` cannot tell a pending key from a broken one | `d7w5.2.1#4` | P2 |
| S9 | **`swarm manifest set` type errors name the Go type, not the shape it accepts** — 3 rejected writes in 4 attempts | `d7w5.2.1#5` | P2 |
| S10 | **No way to list the declared manifest keys.** `target` was found by running `validate` against an empty manifest and reading key names out of failure lines | `d7w5.2.1#6` | P2 |

**Deduped out — `d7w5.2.1#9`** (a JSON-based harvest of findings silently
returns empty). The remedy it asks for is **already in the skill that ran this
retro**: `swarm-retro` §2 states "`bd export`, never `bd list` plus
`bd show --json`… `bd show --json` omits comments entirely even though its text
output prints them." The skill half needs no issue. The `bd` half is a durable
platform fact and is candidate **P4** below.

### Project audience → beads in this tracker

| # | candidate | findings | rank |
|---|---|---|---|
| P1 | **A model or schema change cannot be scoped to "declare and regenerate", and grooming does not say so.** Three of the four beads under-declared their surface in the same direction: a Strata declaration forced the emitter, two fixtures and three tests; a wiring-only item had to add a function to the previous item's package; a command item made a standards document's status text false. `c4_docs` is what the partitioner serializes on, so under-declaration silently defeats batch serialization — it cost nothing here only because the partition had zero parallelism. Proposed: `groom-beads` gains a rule that a bead touching a generated model declares its emitter, its fixtures, its snapshots and any status-bearing standards document in `c4_docs`. **`groom-beads` is vended by this repository, so this is a bead here, not an outbox file.** | `f6tl.1#2`, `f6tl.3#2`, `f6tl.4#4` | **P1** |
| P2 | **`.sf/outbox/README.md` and `.claude/skills/triage-issue/SKILL.md` both assert "we vend the skills we run, so we are our own downstream", which `swarm.toml`'s `[retros]` comment and bead `agent-plugins-g990` say is no longer true.** `triage-issue` sweeps `.sf/outbox/` into this tracker without anyone forwarding anything — so the ten swarm-plugin issues above would be swept straight back into the backlog this repository cannot act on, which is precisely the accumulation the audience split exists to prevent. Proposed: both documents state that `.sf/outbox/swarm/` is for **delivery to `brettcallaghan/swarm`** and is not swept here. | `d7w5.2.1#1`–`#11` (their existence is the evidence), plus the two documents quoted | **P1** |
| P3 | **`cli-output-standards.md` carries a hand-maintained conformance table and a status line that were false from the day they were written**, and nothing checks them. The document says "A command specified after this document is written to conform from the start" while asserting "No command conforms today" — so the first conforming command makes it untrue, and the only thing that catches it is an agent noticing. Proposed: either derive the table, or replace the corpus-wide assertion with a per-command row nothing else has to remember. | `f6tl.4#4` | P2 |
| P4 | **Three durable `bd` 1.2.2 facts belong in `AGENTS.md`'s environmental-failure ledger**, where the next agent looks: (a) `bd show <id> --json` returns `.comments` = **null** while the same object reports a non-zero `comment_count` and the text view prints every body — **re-measured by `retro-d7w5` on 2026-09-04**: `bd show agent-plugins-d7w5.2.1 --json | jq '.[0].comments'` → `null`, `jq '.[0].comment_count'` → `12`; (b) `bd` counts a **deferred** dependency as blocking, so a bead whose blockers were deferred can never enter `bd ready` — **re-measured by `retro-d7w5` on 2026-09-04**: `bd blocked --parent agent-plugins-f6tl` → `agent-plugins-f6tl.7: Blocked by 2 open dependencies: [agent-plugins-f6tl.6 agent-plugins-f6tl.5]`, and `bd export` reports both of those `deferred` (the count is 2 rather than AMENDMENT 1's 6 only because `.1`–`.4` have since closed); (c) `bd human respond` comments *and* closes and fails with `storage is nil` — **inherited from `f6tl.7`'s bounce comment and not re-measured here**, and it must be marked unverified in the ledger if filed. | `d7w5.2.1#9`, AMENDMENT 1, `f6tl.7` | P2 |
| P5 | **The two `gap` findings name exactly the two areas review then returned P1 defects against**, and both are places a blueprint declined to decide: whether a heading inside a blockquote or list item is reported and how its section span is bounded (`f6tl.2#3`), and that a frontmatter block's closing `---` is a setext heading so a whole-file search reports the block as a level-2 heading (`f6tl.3#1`). The already-filed `d7w5.7`, `d7w5.9` and `d7w5.12` fix the **code**; none of them says the module blueprint must state the rule. Proposed: `specs/modules/document-core/document-core-module.md` decides both cases explicitly. **Overlaps the three deferred review beads — decline if they are considered sufficient.** | `f6tl.2#3`, `f6tl.3#1` | P3 |

**Considered and not raised as candidates.** The eleven `decision` findings are
each already recorded where they belong — in `specs/modules/document-core/document-core-module.md`,
`specs/containers/spec-cli/spec-cli-container.md`, or a doc block beside the
code — by the agent that made them. Their volume is a signal about how much the
structure blueprint left open, but "eleven decisions were made at
implementation time" is a summary, not a change anyone can accept or decline.
They stay in this document so a second run reporting the same shape can be
counted against them.

## What was filed

**Nothing yet.** The candidate list above has not been put to a human, and this
skill files nothing without one. `manifest.retro_filed`, `retro_declined` and
`retro_routed` are therefore **unwritten** — they land as one merge when the
decision comes back, and writing `retro_filed: 0` now would report a decision
nobody made.

Routing itself is complete and does not depend on the decision: 28 harvested,
16 to the project branch, 11 to the skill branch, 1 terminating at `proof`,
**0 unrouted**.
