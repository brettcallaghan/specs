# Run record — spec corpus structure extraction

*Written by `retro-d7w5` on 2026-09-04. One document: the run's results, every
finding it captured, and what it could not establish.*

*Written in two passes, because the run re-entered its dispatch loop after the
first one. **Pass 1** recorded the run as it then stood — four work orders
closed, review pass 1's nine findings filed and none admitted. `AMENDMENT 3`
then admitted four of those findings, the run went back to groom & plan, and
`AMENDMENT 4` closed the loop at head `62e9e42`. This document was revised
there, at the run's true end.*

*Pass 1's content is **preserved, not rewritten**: every finding it recorded,
its routing, its candidate list and the filing decision it took stand exactly as
reviewed and committed. What was revised is Results, Scope, Validation, Pushes,
the routing and finding counts, and what could not be verified — plus the
seventeen findings the second half of the run produced and the four candidates
they raise.*

## The run

| | |
|---|---|
| run epic | "Swarm run — Spec corpus structure extraction" |
| target | "Spec corpus structure extraction", adopted whole |
| sprint branch | the sprint branch, merged into `claude/swarm-agent-plugins-5wxrfo` |
| `base_sha` → head | `70dddef` → `62e9e42` |
| queue | 7 beads under the target. **5 admitted at the gate** (`.1 .2 .3 .4 .7`), 2 excluded; AMENDMENT 1 dropped `.7`, leaving **4**. `AMENDMENT 3` then admitted four review findings (`d7w5.7 .8 .9 .11`), so `manifest.admitted` ends at **8 ids** |
| dispatches | three — `dev-d7w5-1` (batch `f6tl.1 → .2 → .3 → .4`), then `dev-d7w5-2` (`d7w5.9 → .8 → .7`) and `dev-d7w5-3` (`d7w5.11`) **concurrently** |
| partition | pass 1: a strict dependency chain — zero parallelism, zero collision edges. Pass 2: **two collision edges** in `.beads/plans/d7w5.pass2.edges.jsonl` (`.8`←`.9`, `.7`←`.8`), partitioning four beads into two batches — **the run's first real parallelism**. `manifest.partition_preview` records pass 1 only |
| declared checks | 12 — `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| model lineup | `advanced` tier → `opus` per `swarm.toml` `[models]`, for all three dispatches. **The manifest's `models` key is unset** and `effort` reads `unrecorded` — see `d7w5.2.1#3` |
| gate | one authorization gate, taken once |
| amendments | 4, all on the run epic — 1 narrowed the queue, 2 recorded review pass 1, 3 widened scope after the drain, 4 closed the loop |
| review passes | **2 of 3. The budget is unspent, not exhausted** |

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

**Eight beads closed across three dispatches, each in one commit, none blocked
and none skipped.** Four are the target's own work orders; four are review pass
1's P1 findings, admitted by `AMENDMENT 3` and worked in the second half.

| bead | status | agent | commit | what landed |
|---|---|---|---|---|
| `f6tl.1` | closed | `dev-d7w5-1` | `e4daa8c` | Declared the index model's `headings` and `record_blocks` structure properties in Strata and regenerated |
| `f6tl.2` | closed | `dev-d7w5-1` | `50f8835` | Built the heading tree, section spans and heading identifiers in `cli/core/markdown` |
| `f6tl.3` | closed | `dev-d7w5-1` | `1cf26e0` | Populated every member's headings during the corpus walk, from the body offset |
| `f6tl.4` | closed | `dev-d7w5-1` | `ef94bd1` | Added the `headings` query command with `--json` and `--emit` |
| `d7w5.11` | closed | `dev-d7w5-3` | `b9b9510` | Ported `encoding/json`'s `typeFields` into `renderYAML`, so both renderings resolve promoted fields and name conflicts identically |
| `d7w5.9` | closed | `dev-d7w5-2` | `85e3d20` | `prefixed()` stopped re-sniffing the raw line, so a heading in a container block gets a non-overlapping `Position` and a real `Section` |
| `d7w5.8` | closed | `dev-d7w5-2` | `f0bfadb` | `markdown.Identifier` stopped parsing heading text as a document, so an ordinal or list marker survives into the identifier |
| `d7w5.7` | closed | `dev-d7w5-2` | `cbfb438` | Exported `frontmatter.Locate` and gave the corpus walk one block-location rule, so an unparsable frontmatter block is no longer indexed as a heading |

One further commit, `62e9e42`, is the coordinator's single end-of-run coverage
regeneration and belongs to no bead.

- **8 completed, 0 blocked, 0 skipped, 0 tier escalations.**
- `manifest.consecutive_dispatch_failures` = 0 throughout; the circuit breaker
  never armed.
- **Pass 1 acceptance:** pre-merge checks 4/4 exit 0; `swarm dispatch accept`
  reported merged true, ancestor true, worktree removed, `beads_checked` 4,
  `acceptance_failures` `[]`.
- **Pass 2 acceptance:** pre-merge checks 4/4 exit 0; both accepts returned
  `acceptance_failures: []`, with `beads_checked` 3 (`dev-d7w5-2`) and 1
  (`dev-d7w5-3`). **The second accept failed on its first call** and needed a
  manual rebase inside the worktree first — the run's only tooling failure, and
  finding `d7w5.2.1#12`.
- `.7`, `.8` and `.9` each exist at two shas; they were rebased during
  acceptance and only the shas above are reachable from head.
- Measured on the finished tree: `git worktree list` returns one entry, the
  repository itself — **no dispatch worktree stands**. `git status --porcelain`
  is empty.
- Diff `70dddef..62e9e42`: **91 files, +5380/−81** (measured). The second half
  alone, `1bb35e7..62e9e42`, is 23 files, +2172/−87. (Pass 1's own figure, for
  the record it was taken in: `0f8d8b7..1bb35e7`, 74 files, +3288/−77.)

## What the run decided

Seventeen of the 45 findings are the four *work* beads' own — and **eleven of
them are `decision`**: a design question the work order or the blueprint left
open and the implementing agent had to settle. That was pass 1's dominant
signal, and `decision` remains the run's largest kind at **17 of 45**. The full
text of each is below; what pass 1's eleven decided, in one line each:

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

There were **no `sequencing` findings** in pass 1, consistent with a
strict-chain partition run behind one agent.

### What pass 2 decided

The four admitted review-finding beads produced **six more `decision` findings**,
plus one `scope`, one `gap`, one `stale` and six `proof`.

**Three of the six decisions are one decision taken three times.** `d7w5.7#3`,
`d7w5.8#4` and `d7w5.9#3` each record that `README.md`'s coverage figures are
**left stale by instruction**: the plan forbade per-bead regeneration precisely
so the two concurrent batches could not conflict on that file, and the
coordinator regenerated once at the end. That is the plan working, not friction —
and it is the only place in this run where a deliberate instruction had to be
repeated in three beads to hold.

**The other three are a single shape, and it is the second half's dominant
signal: in three of the four beads, what the bead REPORTED was right and what it
PRESCRIBED was wrong or unreachable.**

| finding | what the bead prescribed | what shipped, and why |
|---|---|---|
| `d7w5.11#2` | flatten an embedded struct on `f.Anonymous && !tagged && Kind()==Struct` | `encoding/json` flattens on an empty resolved tag **name**, not on the absence of a tag. The Design's `!tagged` would have rendered an embedded struct tagged `json:",omitempty"` under its type name — reintroducing the same class of divergence the bead exists to close, one case narrower |
| `d7w5.7#2` | "`Parse` and the exported locator share one implementation — the delimiter constants are read in exactly one function", with `Locate(src) (span.Region, bool)` | the two are unreachable together: that signature cannot distinguish "no block" from "a block opened and never closed", which `Parse` must. The AC's **substance** — one block-location rule, in `Locate` — holds and was verified; its literal wording was not met |
| `d7w5.8#4` | `Identifier("- Bullet")` is `-bullet`; the identifier of ``3. `## Design` present`` is `3-design-present` | `--bullet` and `3--design-present`. The AC was amended to the measured values and the documented character filter left alone, because changing the filter would silently move the anchor of every corpus heading carrying punctuation beside a space |

The `stale` finding `d7w5.8#2` names the mechanism behind the third row exactly:
the bead's **measured** table (render output) was correct byte for byte, and the
sentence **deriving** identifiers from it was an unmeasured inference, two of
whose four values were wrong. `d7w5.11#3` is the same defect from the other
side — the bead's three prescribed code fixes would have left the guarantee
still false, because two further divergences existed that its own reproduction
never reached, and the agent ported `typeFields` in full instead.

Still **no `sequencing` findings**, in a pass that did have a partition and did
run two agents at once.

## Scope

The run finished its **admitted** set — twice, because the set changed under it.
Scope was **narrowed twice at the start and widened once after the first
drain**, and the table below is the narrowing; the widening is `AMENDMENT 3`,
further down.

| bead | disposition | reason |
|---|---|---|
| `f6tl.5` | excluded, deferred | Stage 2: depends on the record-block tag registration `FEAT-SPECSTR-004` specifies in `frd-spec-structure-conformance`, which is **not yet a work order**. Excluded by the human at the gate |
| `f6tl.6` | excluded, deferred | Stage 2: a registration point no profile fills leaves it inert, per its own description. Excluded by the human at the gate |
| `f6tl.7` | admitted, then **dropped** | AMENDMENT 1: unreachable once `.5` and `.6` were excluded — `bd` counts a **deferred** dependency as blocking, and three of its ACs require stage 2 landed. Decided by `coordinator-d7w5` at groom & plan, not escalated; deferred and labelled `human` |

`manifest.moved` is `[]` — no bead was moved from another epic
(`move_foreign_beads=false`).

**The target epic is still open, and that is correct.**
Closing it would report the feature complete while its stage-2 deltas are
unlanded. `f6tl.7` carries the bounce record and two ways a human can clear it
(admit stage 2, or split a stage-1-only DoD verification out of it).

**Review pass 1 filed 9 findings and the run admitted none of them at the
time** — `admit_review_findings=false`, authorized at the gate. `d7w5.7` …
`d7w5.15` (4 P1, 4 P2, 1 P3; 6 in scope, 3 out) were all deferred. The loop did
not re-enter, because its condition is findings *admitted*, and zero were.

**`AMENDMENT 3` reversed that for four of them.** At the retro decision the
human overrode `admit_review_findings=false` after the fact and admitted the
four P1 findings — `d7w5.7`, `.8`, `.9`, `.11`. `manifest.admitted` went from
four ids to eight. The run re-entered at **groom & plan**, not at dispatch,
because review files and does not groom. This was a **scope widening after the
run's admitted set had already drained**, and it is why this record exists in
two passes. The other five findings (`.10`, `.12`, `.13`, `.14`, `.15`) stayed
deferred and unadmitted, and are still deferred at the run's end.

**Groom & plan pass 2 groomed all four to Ready and none bounced.** Each was
assigned a tier, and the pass produced **two collision edges** —
`.beads/plans/d7w5.pass2.edges.jsonl`, `.8`←`.9` and `.7`←`.8` — the first real
partition this run had. The ready-set assertion held both ways: all four ready
before the edges were loaded, exactly `.9` and `.11` after.

**Review pass 2** ran over `ec6b6fd..62e9e42` with the same declared reviewer
(`swarm.toml` `[review] skill = "code-review"`). It confirmed that all four
fixes fix the defect pass 1 found, each with a test that fails against the
pre-fix code for the stated reason, and filed **four further findings** —
two at P2 (the YAML-rendering performance regression and the `omitzero`
key divergence) and two at P3. **None was
admitted**, and all four were filed with `parent=null`, outside the run epic —
which is exactly the correction to the defect pass 1 committed and that
candidate S1 was filed against.

**`AMENDMENT 4` closed the loop.** `standing_auth.admit_review_findings`
returned to `false`: AMENDMENT 3's override was a one-time admission of pass 1's
four P1s, not a standing change. The loop condition — findings *admitted*, and
passes under budget — failed on its first clause, so the dispatch loop did not
re-enter. **`review_passes` = 2 of 3: the budget is unspent, not exhausted**,
and the run ends with a review pass in hand rather than out of them.

**Nine beads were filed outside the run and not one entered its dispatch set.**
`zka9` (P3) and `1j2i` (P2) by the dispatches; `yztv` (P2), `912t` (P2), `s5ok`
(P3) and `rk27` (P3) by review pass 2; `3oiv` (P1), `yqjh` (P1) and `nduj` (P2)
by this retro's first pass. Measured at the run's end from `bd export`: all nine
carry `parent=null` and all nine are open.

## Every finding, verbatim

**45 findings across 9 beads** — 28 harvested in pass 1 (below, unchanged) and
17 in pass 2, from the four admitted review-finding beads plus one further
`tooling` finding on the gate bead.

Counts verified against each bead's `comment_count` before aggregating. Pass 1:
`f6tl.1` 5, `f6tl.2` 4, `f6tl.3` 3, `f6tl.4` 5 — 17, every comment then a
finding — and `d7w5.2.1` 11 findings plus one gate record. Pass 2: `d7w5.7` 3,
`d7w5.8` 4, `d7w5.9` 4, `d7w5.11` 5, and `d7w5.2.1#12`. The `comment_count`s
read higher today — `f6tl.1` 6, `f6tl.3` 4, `f6tl.4` 7, `d7w5.2.1` 24 — because
pass 1's fifteen `retro:filed:` audit marks are comments too. They are marks,
not findings, and are excluded here.

Harvested with `bd export`, never `bd show --json`, for the reason
`d7w5.2.1#9` records: the JSON view returns `.comments` as `null` while the same
object reports a non-zero `comment_count`.

### `decision` — 17

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


**From pass 2 — the four admitted review-finding beads, and one further finding on the gate bead:**

**`d7w5.7#2`**

> retro:decision An acceptance criterion is met in substance but not in its literal wording, because the Design's own prescribed signature makes the literal reading unreachable. Reviewers should confirm this at the pull request.
> evidence: AC — "frontmatter.Parse and the exported locator share one implementation — the delimiter constants are read in exactly one function." Design — "export a locator ... e.g. func Locate(src []byte) (span.Region, bool)".
> With that signature Locate cannot distinguish "src carries no block" from "a block is opened and never closed", yet Parse must, because those are two of its three documented outcomes (a nil,nil versus a located error). Parse therefore keeps the bytes.HasPrefix(src, []byte(openingDelimiter)) test and the len(openingDelimiter) it measures the error span with — so openingDelimiter is read in two functions however the code is arranged, short of changing the signature the Design pinned.
> Measured on the finished tree: grep -rn "openingDelimiter|closingDelimiter" --include=*.go . returns, for this package, frontmatter.go:49 and :50 (the declarations), :137 inside Locate, and :163 and :166 inside Parse's unclosed-block error. The DELIMITER PAIR — the block-location rule, which is what the bug was — is read in exactly one function, Locate; Parse's two remaining reads are the opening delimiter alone, for reporting, and cannot close a block anywhere.
> The AC's substance holds and is verified: grep -rn "frontmatterDelimiter" . --include=*.go returns nothing (exit 1), so corpus.go declares no delimiter constant, and span.Find over the frontmatter delimiters appears once in the tree.
> REJECTED — widening Locate to return the opened-but-unclosed case (a third return value, or an error) so Parse reads no constant. It would satisfy the wording, and it changes an exported signature the Design fixed, to move a string used for an error message. Rejected as the more invasive answer to a criterion whose stated purpose — the Design says "so there is one implementation rather than two" — is already met.

**`d7w5.7#3`**

> retro:decision Coverage figures in README.md are LEFT STALE by this bead, by instruction; the run's coordinator must regenerate them before the run is finished.
> evidence: the bead's Design says "DO NOT run scripts/coverage.sh or commit README.md coverage figures. The run's coordinator regenerates coverage once at the end", and the dispatch prompt repeats it, naming concurrent dispatch dev-d7w5-3 under cli/**. This bead changes cli/core/corpus/corpus.go and cli/core/frontmatter/frontmatter.go and adds tests in both packages, so the committed figures no longer match a fresh measurement; AGENTS.md ("Checks") states cli.yml's coverage job runs scripts/coverage.sh --check and fails when they do not match. This is the third bead of dispatch dev-d7w5-2 to leave them stale — d7w5.9, d7w5.8 and d7w5.7 all touch cli/**. Verified no coverage run happened here: git status --porcelain lists only cli/core/corpus/corpus.go, cli/core/corpus/corpus_internal_test.go, cli/core/frontmatter/frontmatter.go, cli/core/frontmatter/frontmatter_test.go and specs/modules/document-core/document-core-module.md.

**`d7w5.8#3`**

> retro:decision AMENDED an acceptance criterion rather than the code: the AC's two literal expected strings were changed to the measured values, and the character filter was left exactly as documented. Reviewers should confirm this at the pull request.
> evidence: AC as written — "Identifier(\"- Bullet\") is -bullet, and the identifier of the heading text 3. `## Design` present is 3-design-present". Shipped instead: "--bullet" and "3--design-present" (see the retro:stale comment for the measurement). Everything the AC asserts SUBSTANTIVELY is met and tested: numbered siblings differ (1-overview vs 2-overview), the ordinal survives, no identifier derived from a digit-initial heading begins with a hyphen, and inline rendering is unchanged.
> REJECTED — make the AC's strings true by collapsing runs of hyphens and trimming leading ones. Cost, and why it is not this bead's call: (1) it contradicts the derivation documented in the Markdown Structure component block of specs/modules/document-core/document-core-module.md, which this bead is only meant to extend, not rewrite; (2) it breaks a committed assertion, headings_test.go's "corpus-in-module-document-core-modified--idx"; (3) the identifier is the whole key of the <document>#<heading> addressing grammar, so it would silently change the anchor of every corpus heading carrying punctuation beside a space — measured on one document alone, spec headings specs/standards/work-order-definition-of-ready.md --json shows "7. Tier assigned (+ reason)" -> 7-tier-assigned--reason, which predates this bead; (4) the bead's own Design says only rendered() changes and the existing character filter stays.
> FILED rather than decided silently: a P2 bead (discovered-from d7w5.8) puts the collapse-and-trim question, with the two answers and their costs. It records as UNVERIFIED the premise the bead's problem statement leans on — that GitHub renders these without the doubled hyphen — because no probe of a rendered GitHub anchor was run in this container, and names what would settle it.

**`d7w5.8#4`**

> retro:decision Coverage figures in README.md are LEFT STALE by this bead, by instruction; the run's coordinator must regenerate them before the run is finished.
> evidence: the bead's Design says "DO NOT run scripts/coverage.sh or commit README.md coverage figures. The run's coordinator regenerates coverage once at the end", and the dispatch prompt repeats it, naming concurrent dispatch dev-d7w5-3 under cli/**. This bead changes cli/core/markdown/headings.go and adds tests, so the committed figures no longer match a fresh measurement; AGENTS.md ("Checks") states cli.yml's coverage job runs scripts/coverage.sh --check and fails when they do not match. Rejected alternative: regenerate per bead — it would guarantee a README.md merge conflict with the sibling batch, which is why the instruction exists. Verified no coverage run happened here: git status --porcelain lists only cli/core/markdown/headings.go, cli/core/markdown/headings_test.go and specs/modules/document-core/document-core-module.md.

**`d7w5.9#3`**

> retro:decision Coverage figures in README.md are LEFT STALE by this bead, by instruction, so the run's coordinator must regenerate them before the run is finished.
> evidence: the bead's Design says "DO NOT run scripts/coverage.sh or commit README.md coverage figures. The run's coordinator regenerates coverage once at the end", and the dispatch prompt repeats it, naming concurrent dispatch dev-d7w5-3 under cli/**. This bead adds ~120 lines of test to cli/core/markdown/headings_test.go and changes cli/core/markdown/headings.go, so the committed figures no longer match a fresh measurement. AGENTS.md ("Checks") states that cli.yml's coverage job runs scripts/coverage.sh --check and FAILS when the committed figures do not match. Rejected alternative: regenerate per bead — it would guarantee a README.md merge conflict with the sibling batch, which is the reason the instruction exists. Verified in this worktree that no coverage run happened and README.md is untouched: git status --porcelain -> only cli/core/markdown/headings.go, cli/core/markdown/headings_test.go, specs/modules/document-core/document-core-module.md.

**`d7w5.11#2`**

> retro:decision The Design's proposed flatten guard (`f.Anonymous && !tagged && Kind()==Struct`) is not what encoding/json does; implemented the tag-NAME test instead.
>
> evidence: encoding/json flattens on an empty resolved tag NAME, not on the
> ABSENCE of a tag. Its typeFields reads
>
>     name, opts := parseTag(tag)
>     if !isValidTag(name) { name = "" }
>     ...
>     if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct { ... render as one key ... }
>     // else: explore the embedded struct
>
> So an embedded struct tagged `json:",omitempty"` — tagged, but with an empty
> name — still flattens, and one whose tag name is rejected by isValidTag
> flattens too. The Design's `!tagged` would have rendered the first as a key
> named after its type.
>
> Measured against the shipped code, case "embedded struct tagged with options
> only" in cli/cmd/yaml_promotion_test.go:
>
>     struct { PromotionTagged `json:",omitempty"`; B string `json:"b"` }
>     JSON: ["t" "b"]   YAML: ["t" "b"]
>
> REJECTED the Design's literal `!tagged` wording: it is a near-miss that
> reintroduces the same class of divergence the bead exists to close, one case
> narrower. The bead's INTENT — "make structFields mirror encoding/json for
> promoted fields" — is served by mirroring the actual condition. Cost of the
> choice: `jsonName` was replaced by `parseJSONTag` + `validTagName` (a port of
> encoding/json's isValidTag, including its punctuation set) rather than gaining
> a fourth return value; the acceptance criterion that speaks to this ("an
> embedded field that DOES carry a json tag still renders under that tag's name")
> is satisfied either way, since it is about a tag that NAMES the field.


### `scope` — 4

**`f6tl.4#4`**

> retro:scope Landing this item made cli-output-standards' own status text false, so the standards document had to change with it — the item's c4_docs names only spec-cli-container.md.
> evidence: The document read 'Status: approved, and not yet met' and its conformance table 'No command conforms today; the surface predates these standards.' Both become untrue the moment a conforming command exists, and the same document says 'A command specified after this document is written to conform from the start' — so the falsehood was built into it from the day it was written. Updated the status line, the table preamble and added a headings row rather than leaving a document that contradicts the tree; nothing else in it changed, and no other command's row was touched since none of them moved.

**`f6tl.3#2`**

> retro:scope Delivering this item required adding a function to cli/core/markdown (HeadingsFrom), which is the HEAD item's surface, not this one's. The item scopes itself to wiring — 'Corpus holds no parsing logic of its own' — and anchors only corpus.go and corpus.Open.
> evidence: Two constraints met in the same place. Headings must be recorded in document coordinates, as every other index position is, and the search must skip the frontmatter. Slicing the body in Corpus and shifting offsets afterwards would not work: span.Of derives line and column from the start of the slice it is given, so a body-relative slice yields a heading on line 1 of a file where it is on line 10, and re-deriving them in Corpus would put position logic in the one component the module says computes none. HeadingsFrom parses the body slice but calls span.Of against the whole source, which keeps both rules. The alternative — widening the HEAD item retroactively — was rejected because it had already closed; the change is additive (Headings is now HeadingsFrom at offset zero) and both the module blueprint's Markdown Structure and Corpus sections record it.

**`f6tl.1#2`**

> retro:scope Declaring the two properties as REQUIRED lists forced changes outside 'declare in Strata and regenerate' — the emitter, two hand-maintained input fixtures and three unit tests — none of which the bead's Design or Anchors anticipate.
> evidence: The bead's Design is 'Declare the shapes in schemas/index.strata and regenerate', Anchors name only schemas/index.strata, cli/core/schema/spec-model.schema.json and cli/core/specmodel. Measured after regenerating, before any further edit — 'cd cli && go test ./...' rc=1 with three distinct failures: (a) specindex_test.go:166 and :203 'unmarshal: IndexEntry: required field "headings" is missing or null' from entries built as Go literals and round-tripped; (b) cache_test.go:387 'Open rebuilt an index whose inputs are unchanged', because the planted entry marshalled headings as null and the warm read then rejected its own file; (c) portability_test.go:71 index-default, the emitted index carrying '"headings": null, "record_blocks": null'. Required-list was the right call — it is exactly how 'references' and 'features', the first two exercises of the additive-at-1.0 rule, are declared on SpecIndex, and the blueprint states 'No backward-compatibility window applies' — but the consequence is that a MODEL-only bead cannot land green without touching the emitter. The fix mirrors the rule cli/core/specindex/specindex.go:35 already states for the index's own collections ('Every collection is non-nil so that an index containing no members marshals its collections as [] rather than as null'), applied centrally in builder.addEntry so all four IndexEntry construction sites in corpus.go are covered.


**From pass 2 — the four admitted review-finding beads, and one further finding on the gate bead:**

**`d7w5.11#3`**

> retro:scope The bead's three named code fixes would have left the guarantee still false; two further divergences existed that the bead's reproduction did not reach.
>
> evidence: the Design named three changes (tag-presence in jsonName, the flatten
> guard, the PkgPath skip). Implementing only those leaves structFields as a
> DEPTH-FIRST walk with no conflict handling, while encoding/json resolves fields
> BREADTH-first and then applies dominance rules. Two shapes measured failing on
> this run's tree, neither mentioned in the bead:
>
> (a) A shallower field does not hide a promoted one of the same name.
>
>     type PromotionShadowed struct { Name string `json:"name"`; Deep string `json:"deep"` }
>     struct { PromotionShadowed; Name string `json:"name"` }
>
>     JSON: ["deep" "name"]                        (outer "name" wins; inner hidden)
>     YAML before the fix: ["PromotionShadowed" "name"]
>     YAML with only the three named fixes: would be ["name" "deep" "name"] —
>     the key "name" TWICE, carrying two different values.
>
> (b) Two embedded structs at one depth offering the same name do not annihilate.
>
>     struct { PromotionLeft; PromotionRight }   // both declare `Clash string`
>     JSON: ["left" "right"]                     (Clash dropped from both)
>     YAML before the fix: ["PromotionLeft" "PromotionRight"]
>
> DECIDED to port encoding/json's typeFields in full — breadth-first walk,
> `count`/`nextCount` duplicate-path bookkeeping, the name/depth/tagged sort,
> dominantField, and the return to index order — rather than patch the three
> named spots. The sentence this bead repairs in
> specs/containers/spec-cli/spec-cli-container.md promises "the same keys in the
> same order"; a partial fix would have left me writing a claim I had just
> measured false, which is the exact defect the bead was filed for.
>
> REJECTED the narrower reading of the bead's scope. Cost: cli/cmd/yaml.go grew
> by roughly 180 lines and now carries a port of stdlib logic that must be
> re-checked if the Go version's own rules change. The bead's tier_reason names
> "conflicts" among the intricacies, so this is within what it was sized for.


### `gap` — 3

**`f6tl.2#3`**

> retro:gap A heading nested inside a blockquote or a list item is reported like any other, and its section span is computed in document coordinates — neither the item nor the HEAD delta says whether it should be.
> evidence: The blueprint says only 'every heading of a document body in document order'. Measured against the implementation on body '# Top\n\n> ## Quoted\n>\n> text\n\n- item\n  ## In list\n\n## Real\n\nend\n': four headings are returned, and the blockquoted one's section slices to '> text\n\n- item\n' — carrying the '>' markers and running past the end of the blockquote into the list that follows. That is the honest consequence of section spans being document ranges rather than container-relative ones, but it means a caller emitting such a section (the CMD item's AC-SPECSTR-003.5) would print quote markers. Left as-is rather than special-cased: measured 0 headings inside a blockquote or list item across specs/ (the 3 grep hits are indented YAML comments inside fenced blocks, which goldmark correctly does not treat as headings), so nothing in the corpus exercises it and inventing a rule with no instance to check it against would be guessing.

**`f6tl.3#1`**

> retro:gap A whole-file heading search reports the entire frontmatter block as a level-2 heading, so the walk has to start at the body offset. The IDX delta calls this 'a wiring change, not a second parser' and its Anchors name only cli/core/corpus/corpus.go — the hazard appears nowhere in the item.
> evidence: A frontmatter block's closing '---' underlines the YAML above it, which is a setext heading in CommonMark. Measured by running the heading extraction over a real corpus file, specs/adrs/adr-008-element-record-blocks.md: 7 headings found, of which the first is level=2 with text 'title: "ADR-008: Machine-Readable Element Record Blocks with a Two-Schema Model"\ndescription: Un-promoted elements carry a fenced record block...\nslug: adr-008\ntype: adr\nstatus: active\nupdated_at: 2026-08-05T00:00:00Z' — the whole block. The blueprint does carry the fact, but only as a relied-upon element ('Frontmatter — the body offset a heading search starts from'), which reads as a convenience rather than as the thing that makes the output wrong. It also cuts against the nearest precedent: cli/core/corpus/features.go documents that 'The whole of src is offered to the fenced-block lookup, frontmatter included, so that every span a block yields is already in the document's own coordinates', which is safe for fences and not for headings. Pinned by TestASearchFromTheBodyOffsetLeavesTheFrontmatterBehind, whose negative control asserts the whole-file search still reports the frontmatter, so the two searches cannot be silently unified later.


**From pass 2 — the four admitted review-finding beads, and one further finding on the gate bead:**

**`d7w5.11#5`**

> retro:gap The sweep for copies of the retracted "cannot drift" claim reached every surface; one copy remains true after the fix and one further divergence was filed as a bead.
>
> evidence: grepped the distinctive phrasing BEFORE editing the source.
>
>     grep -rn "cannot drift" . | grep -v '^\./\.git/'
>       specs/containers/spec-cli/spec-cli-container.md:357   <- the retracted claim, rewritten
>       specs/containers/spec-cli/spec-cli-container.md:369   <- about cobra help text, unrelated
>       specs/adrs/adr-015-go-for-the-spec-cli.md:224         <- about clap/Rust --help, unrelated
>       specs/features/gherkin-adoption/gherkin-adoption-blueprint.md:309  <- unrelated
>       specs/standards/document-relationship-standards.md:48 <- unrelated
>       scripts/sync-shared-references.sh:46                  <- unrelated
>       cli/cmd/root.go:19                                    <- about cobra help text, unrelated
>       .claude/skills/strata/SKILL.md:165                    <- unrelated
>
>     grep -rn "same struct \`encoding/json\`|same field order|walks the same struct" .
>       specs/containers/spec-cli/spec-cli-container.md:355-356  <- rewritten
>       cli/cmd/yaml.go:15                                       <- file header, rewritten
>       docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md:87
>
> The retro-document copy reads: "The YAML renderer is hand-written over the same
> struct `encoding/json` marshals, not `goccy/go-yaml` configured — parity is by
> construction". That is a run record of a DECISION, and it is TRUE as of this
> bead: parity is now by construction. No retraction needed, and a run record is
> not a surface to edit after the fact.
>
>     ls docs/plans      -> absent
>     ls .sf/outbox      -> README.md and swarm/ only; no issue carrying the claim
>
> No durable trigger prompt or dispatch prompt in this session carries the claim.
>
> FILED, not fixed here: a P3 bead (discovered-from d7w5.11) —
> renderYAML ignores encoding/json's `,string` tag option, so the two renderings
> still differ in value TYPE where a field carries it. Measured on the fixed tree:
>
>     struct { N int `json:"n,string"` }{5}
>     JSON: {"n":"5"}    YAML: "n: 5\n"
>
> Keys agree, so this bead's acceptance criteria and the rewritten doc sentence
> (which speaks to keys and order, plus the deliberate collection difference)
> are honest as written. Adjacent, not blocking: no type on this command surface
> uses `,string`.


### `stale` — 1

New in pass 2. A claim in a bead that was true when written, or inferred from
something true, and is not true of the tree the bead ran against.

**`d7w5.8#2`**

> retro:stale The bead's Design states post-filter identifier values that the chosen fix does not produce. Its MEASURED table (render output) is correct; the sentence deriving identifiers from it is an unmeasured inference and two of its four values are wrong.
> evidence: the Design says "After Identifier's character filter these give 1-overview, 2-overview, -bullet and 3-design-present", and the Acceptance repeats "Identifier(\"- Bullet\") is -bullet, and the identifier of the heading text 3. `## Design` present is 3-design-present".
> Measured with the fix in place, via a temporary in-package probe calling rendered() and Identifier() (removed before commit):
>   "1. Overview"            -> rendered "1. Overview"            -> identifier "1-overview"     (Design correct)
>   "- Bullet"               -> rendered "- Bullet"               -> identifier "--bullet"       (Design said "-bullet")
>   "3. `## Design` present" -> rendered "3. ## Design present"   -> identifier "3--design-present" (Design said "3-design-present")
>   "## Not a heading here"  -> rendered "## Not a heading here"  -> identifier "-not-a-heading-here"
>   "> quoted"               -> rendered "> quoted"               -> identifier "-quoted"
> Every RENDER above matches the Design's measured table byte for byte, so the parser change is exactly what was specified; only the arithmetic applied to it afterwards is wrong.
> Cause: the documented derivation removes a disallowed character and then maps each REMAINING space to a hyphen, and does not collapse or trim. specs/modules/document-core/document-core-module.md: "every character that is not a letter, digit, hyphen, underscore or space removed and each remaining space replaced by a hyphen." So "- Bullet" contributes the literal "-" AND the space's "-". The Design's derivation silently dropped one of the two.
> The behaviour is not new and is already asserted in the committed tests: cli/core/markdown/headings_test.go asserts Identifier("Corpus in [[module-document-core]] (modified) — IDX") == "corpus-in-module-document-core-modified--idx" — a double hyphen from the dropped em dash beside a space.


### `proof` — 7

**`f6tl.1#5`**

> retro:proof The regenerated snapshots change every index case's input_hash even though no corpus file changed; verified this is by design rather than accepting the diff.
> evidence: git diff on cli/testdata/cases/index-default/expected/manifest.json shows input_hash sha256:3ae7797c... -> sha256:f9bb75b7... alongside the expected per-entry '"headings": []'. cli/core/corpus/hash.go's sum() folds in 'field(sum, "profile-schema", opts.Profile.Schema())' and its doc block states 'the active profile's schema and per-type table decide which candidates are members ... A change to any of them changes the index while every file on disk stands still, so all of them are folded in.' The profile's schema IS the regenerated cli/core/schema/spec-model.schema.json, so a model change necessarily moves the hash. Worth knowing for review: every future MODEL delta will rewrite input_hash in all five index snapshots, and that is correct, not snapshot churn to suppress.


**From pass 2 — the four admitted review-finding beads, and one further finding on the gate bead:**

**`d7w5.7#1`**

> retro:proof The defect was reproduced by test in BOTH packages before the fix, matching the bead's reported figures byte for byte, and the goldens were confirmed unchanged afterwards.
> evidence: tests added first, then run against the UNFIXED code.
> (1) cd cli && go test ./core/corpus/ -run 'UnparsableFrontmatterBlockIsNotIndexed|FallbackStillLocates' -> exit 1:
> --- FAIL: TestAnUnparsableFrontmatterBlockIsNotIndexedAsAHeading
>     corpus_internal_test.go:93: found 2 headings, want 1: [{Text:b
>         status: [unclosed Level:2 Position:{Offset:15 Length:24 StartLine:2 StartColumn:12 EndLine:5 EndColumn:1} Section:{Offset:39 Length:1 ...} RecordBlocks:[]} {Text:Real heading Level:1 Position:{Offset:40 Length:15 StartLine:6 ...} ...}]
> That is the bead's reported probe exactly: level=2, text "b\nstatus: [unclosed", Offset 15, Length 24, StartLine 2.
> The control in the same run, TestTheFallbackStillLocatesTheBodyOfAnOrdinaryDocument, PASSED before and after — unparsable frontmatter with no "---" in a value, a document with no frontmatter, and a "---" appearing later in the body. So the failure is confined to the delimiter-inside-a-value case and the fallback itself was never broken.
> (2) cd cli && go test ./core/frontmatter/ -run Locate -> exit 1, build failed: "undefined: frontmatter.Locate" at frontmatter_test.go:513, :525, :530, :535, :552 — the locator did not exist.
> After the fix: cd cli && go test ./core/corpus/ ./core/frontmatter/ -> both ok, exit 0. Full gate: bb-checks.py -> exit 0, all twelve, no line absent from the run's checks_baseline.
> (3) The AC's fixture invariant, measured on the finished tree: git diff --stat cli/testdata -> no output, no files. The bead predicted this (adr-002-broken.md carries no "---" inside a value) and it held.

**`d7w5.8#1`**

> retro:proof The defect was reproduced by test BEFORE the fix and the failure matched the bead's reported figures exactly; the fix was then confirmed on the real corpus document the AC names.
> evidence: tests added to cli/core/markdown/headings_test.go first, run against the UNFIXED rendered(). Command: cd cli && go test ./core/markdown/ -run 'BlockMarker|InlineRendering'
> Output (exit 1):
> --- FAIL: TestALeadingBlockMarkerIsPartOfTheHeadingsName
>     Identifier("1. Overview") = "overview", want "1-overview"
>     Identifier("2. Overview") = "overview", want "2-overview"
>     Identifier("- Bullet") = "bullet", want "-bullet"
>     Identifier("3. `## Design` present") = "-design-present", want "3-design-present"
>     Identifier("1. Self-contained Problem statement (the fresh-agent test)") = "self-contained-problem-statement-the-fresh-agent-test", want "1-..."
>     numbered siblings collide: both "overview"
>     Identifier("3. `## Design` present") = "-design-present", which begins with a hyphen
>     Identifier("4. `## Acceptance` criteria (`kind: bug`)") = "-acceptance-criteria-kind-bug", which begins with a hyphen
> TestInlineRenderingIsUnchangedByTheBlockMarkerRule passed BOTH before and after — the negative control that the change is confined to block syntax.
> After the fix: cd cli && go test ./core/markdown/ -> ok, exit 0.
> Corpus confirmation, the AC's own command, go build -C cli -o spec ./spec && spec headings specs/standards/work-order-definition-of-ready.md --json:
>   1. Self-contained Problem statement (the fresh-agent test) -> 1-self-contained-problem-statement-the-fresh-agent-test
>   2. Right size (single session)                             -> 2-right-size-single-session
>   3. `## Design` present (non-trivial, non-epic orders)      -> 3--design-present-non-trivial-non-epic-orders
>   4. `## Acceptance` criteria (`kind: bug` & `kind: feature`)-> 4--acceptance-criteria-kind-bug--kind-feature
>   5. Kind-appropriate completeness                           -> 5-kind-appropriate-completeness
>   6. Dependencies wired                                      -> 6-dependencies-wired
>   7. Tier assigned (+ reason)                                -> 7-tier-assigned--reason
>   8. C4 blast radius declared (IF the repo uses C4)          -> 8-c4-blast-radius-declared-if-the-repo-uses-c4
>   9. No open decisions                                       -> 9-no-open-decisions
> All nine carry their ordinal and none begins with a hyphen.

**`d7w5.9#1`**

> retro:proof The defect was reproduced by test BEFORE the fix, and the observed failure matches the bead's reported shape exactly.
> evidence: tests added to cli/core/markdown/headings_test.go first, then run against the UNFIXED prefixed(). Command: cd cli && go test ./core/markdown/ -run 'Container|ListItem|Overlap|Hash'
> Output (exit 1):
> --- FAIL: TestAPrefixedHeadingInsideAContainerBlockCoversOnlyItsOwnLine (0.00s)
>     headings_test.go:325: A's Position slices to "> ## A\n> ## B\n", want "> ## A\n"
>     headings_test.go:328: B's Position slices to "> ## B\n\n", want "> ## B\n"
>     headings_test.go:333: A's Position ends at 14 and B's begins at 7; the two overlap
> --- FAIL: TestAPrefixedHeadingInsideAListItemKeepsItsContentInTheSection (0.00s)
>     headings_test.go:352: Position slices to "- # Item\n  body line\n", want "- # Item\n"
>     headings_test.go:355: Section slices to "", want "  body line\n"
> --- FAIL: TestNoTwoHeadingPositionsOverlapForAnyForm (0.00s)
>     headings_test.go:405: "> ## A\n> ## B\n\ntail\n": heading 0's Position ends at 14 and heading 1's begins at 7; the two overlap
> FAIL
> Negative controls in the SAME run passed before the fix and after it: TestAnUnderlinedHeadingInsideAContainerBlockKeepsItsUnderline ("> A\n> ---\n> body\n") and TestAHashInTheHeadingsOwnTextDoesNotChangeItsForm ("## # Hashy", "a # b\n===\n") — the two cases the old forward scan already got right, so the change is narrowed to the container case. After the fix: cd cli && go test ./core/markdown/ -> ok, exit 0.

**`d7w5.9#2`**

> retro:proof The bead's REJECTED alternative — ask the AST node for the heading's written form — was re-measured at the point of use, because the code comment I shipped restates it and an inherited claim reads exactly like a measured one.
> evidence: goldmark v1.8.5 in this container, /root/go/pkg/mod/github.com/yuin/goldmark@v1.8.5.
> (1) sed -n '200,230p' ast/block.go ->
>     // A Heading struct represents headings like SetextHeading and ATXHeading.
>     type Heading struct {
>         BaseBlock
>         // Level returns a level of this heading.
>         Level int
>     }
>     ... var KindHeading = NewNodeKind("Heading")  // one kind for both forms
> (2) grep -rn "NewHeading(" parser/ ast/ -> parser/setext_headings.go:70, parser/atx_heading.go:95, parser/atx_heading.go:103 all call ast.NewHeading(level); no other constructor and no form argument.
> (3) grep -n "SetAttribute|Attribute" parser/setext_headings.go parser/atx_heading.go -> the only SetAttribute calls are attrNameID and parseLastLineAttributes, i.e. author-written {#id} attributes gated on the opt-in WithHeadingAttribute option; neither parser sets anything that names the written form.
> Conclusion: the public AST genuinely does not carry the distinction on the pinned version, so comparing the line-start and text-start offsets is not a workaround for a simpler API that exists. That is why the comment on prefixed() says so — without it the next reader "simplifies" it back to the forward scan that caused this bug.

**`d7w5.9#4`**

> retro:proof The acceptance criterion "every expected/ artifact under cli/testdata/cases/ is either unchanged or its diff is explained" is satisfied by UNCHANGED, and grooming's premise for that was re-measured rather than inherited.
> evidence: (1) after the fix and a full check run, git status --porcelain in the worktree lists no path under cli/testdata/. (2) The premise — no fixture input carries a heading inside a container block — measured with grep -rlnE '^\s*([>]|[-*+]|[0-9]+\.)\s+#' cli/testdata/cases/ -> no output, exit 1. So no fixture exercises the changed branch, which is also why the new coverage had to be written as unit tests in cli/core/markdown/headings_test.go rather than as a case fixture.

**`d7w5.11#1`**

> retro:proof The defect was reproduced by a failing differential test BEFORE any code changed — 7 of 16 subtests failed, on the shapes the bead named plus three it did not.
>
> evidence: cli/cmd/yaml_promotion_test.go written first, run against the unmodified cli/cmd/yaml.go.
> `cd cli && go test ./cmd/ -run 'TestRenderYAMLResolvesFieldsAsEncodingJSONDoes|TestNilCollections' -v` → exit 1:
>
>     --- FAIL: .../embedded_exported_struct
>           YAML: ["PromotionExported" "b"]   JSON: ["a" "b"]
>     --- FAIL: .../embedded_unexported_struct_type
>           YAML: ["b"]                       JSON: ["u" "b"]
>     --- FAIL: .../embedded_pointer_to_struct,_set
>           YAML: ["PromotionExported" "b"]   JSON: ["a" "b"]
>     --- FAIL: .../embedded_pointer_to_struct,_nil
>           YAML: ["PromotionExported" "b"]   JSON: ["b"]
>     --- FAIL: .../shallow_field_dominates_a_promoted_one
>           YAML: ["PromotionShadowed" "name"] JSON: ["deep" "name"]
>     --- FAIL: .../same-depth_conflict_annihilates_the_key
>           YAML: ["PromotionLeft" "PromotionRight"] JSON: ["left" "right"]
>     --- FAIL: .../tagged_field_wins_a_same-depth_conflict
>           YAML: ["PromotionTaggedClash" "PromotionUntaggedClash"] JSON: ["Clash" "only_tagged" "only_untagged"]
>     --- PASS: .../embedded_struct_with_a_json_tag
>     --- PASS: .../embedded_struct_tagged_with_options_only
>     --- PASS: .../embedded_struct_marked_json:"-"
>     --- PASS: .../field_marked_json:"-"
>     --- PASS: .../omitempty_on_an_empty_field
>     --- PASS: .../omitempty_on_a_populated_field
>     --- PASS: .../untagged_field_renders_under_its_Go_name
>     --- PASS: .../unexported_non-embedded_field_is_invisible_to_both
>     --- PASS: .../nil_slice_and_nil_map
>     --- PASS: TestNilCollectionsRenderPerTheOutputStandards
>
> After the fix the same command exits 0 with all 16 subtests passing.
>
> NEGATIVE CONTROL on the test itself: the first draft named its fixture types
> `promotionExported`, `promotionTagged` etc. in lower case. An embedded field
> takes its TYPE's name as its field name, so a lower-cased type name makes the
> embedded FIELD unexported and every case fell into structFields' `PkgPath !=
> ""` skip — case 1 "embedded exported struct" was silently testing the
> unexported-type path instead. Measured: with the lower-case names, JSON gave
> ["a" "b"] and YAML gave ["b"], i.e. the field vanished rather than rendering
> under the type name. The types were renamed to exported form, which is what
> produced the ["PromotionExported" "b"] evidence above. A differential test
> over embedded structs must use EXPORTED fixture type names or it tests a
> different rule than its case name claims.


### `tooling` — 13

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

> retro:tooling:swarm — `swarm status` reports ceremony beads as scope additions. scope_additions listed d7w5.2.1 (the run's own authorization gate task, closed) alongside the three genuinely-excluded beads. A ceremony bead is structurally never in manifest.admitted, so it will always read as creep and trains the reader to ignore the signal.

**`d7w5.2.1#9`** · target `swarm`

> retro:tooling:swarm — a retro that harvests findings from `bd show --json` will find none of them. Measured on bd 1.2.2 (dev): `bd show <gate task> --json | jq '.[0].comments'` -> null while the same object reports comment_count: 9 and the text view prints all nine bodies. The count says the findings exist and the array does not carry them, so a JSON-based harvest silently returns empty rather than erroring. swarm-retro must read the text view, or bd must populate the array.

**`d7w5.2.1#10`** · target `swarm`

> retro:tooling:swarm — swarm-review parents its findings under the RUN epic, which puts unadmitted work straight into the run's own dispatch set. Measured: after pass 1, `swarm ready --run <run epic>` reported ready_count 9, every one a findings bead with dispatch "" and held false, despite standing_auth.admit_review_findings=false. Only the coordinator's manual diff of the ready set against manifest.admitted caught it; an orchestrator that auto-chained on the ready set would have dispatched all nine against the gate's explicit authorization. Findings should be filed against their origin outside the run epic, or filed deferred, rather than relying on every coordinator to notice.

**`d7w5.2.1#11`** · target `swarm`

> retro:tooling:swarm — the phase chain deadlocks whenever the target epic legitimately cannot close. Measured: `bd close <review phase bead>` -> 'cannot close [review phase bead]: blocked by open issues [[target epic]] (use --force to override)'. f6tl is open only because .5 .6 .7 are deferred, and closing it would falsely report the feature complete — the false-complete failure swarm-scope explicitly warns about. So review, retro and post-flight each need --force to close, on a run that did everything right. A run whose target is a partially-scoped feature epic is the normal case, not the exception, and `swarm phase close` has no path through it.

**From pass 2 — the four admitted review-finding beads, and one further finding on the gate bead:**

**`d7w5.11#4`**

> retro:tooling `go vet`'s structtag analyzer makes one of the shapes under test impossible to write statically, so the test expresses that rule a different way.
>
> evidence: the same-depth annihilation case was first written with both embedded
> structs tagging the field `json:"clash"`. `bb-checks.py` failed:
>
>     === vet
>     # github.com/brettcallaghan/agent-plugins/cli/cmd
>     cmd/yaml_promotion_test.go:254:5: struct field Clash repeats json tag "clash" also at yaml_promotion_test.go:116
>     === vet FAILED (exit 1)
>
> vet forbids two promoted fields repeating one json tag — which is precisely the
> shape whose runtime behaviour (both keys dropped) the test needs to pin. There
> is no `//nolint`-style suppression for `go vet` short of excluding the package.
>
> Worked around by giving both fields the name from the Go FIELD name rather than
> from a tag (`Clash string`, untagged, in PromotionLeft and PromotionRight).
> dominantField compares resolved name, depth and taggedness, and does not care
> which of the two ways a name arrived, so the annihilation rule is still
> exercised; the tagged-vs-untagged tie-break is covered separately by
> PromotionTaggedClash / PromotionUntaggedClash, which vet accepts because only
> one side carries a tag. Verified after the change:
>
>     struct{PromotionLeft; PromotionRight}       JSON {"left":"L","right":"R"}   YAML left/right
>     struct{PromotionTaggedClash; PromotionUntaggedClash}
>                                                 JSON {"Clash":"tagged","only_tagged":"t","only_untagged":"u"}
>                                                 YAML Clash/only_tagged/only_untagged
>
> NOT reachable by this bead: the both-TAGGED same-depth conflict cannot be
> covered by a static test in this repository at all while `vet` is a declared
> check. It would need `reflect.StructOf` to build the type at runtime. Left
> uncovered deliberately; the untagged form exercises the same branch of
> dominantField.

**`d7w5.2.1#12`**

> retro:tooling:swarm — `swarm dispatch accept` cannot accept the SECOND of two concurrent dispatches: it tries to rebase the dispatch branch while that branch is still checked out in the worktree the command itself created. Measured, accepting dev-d7w5-2 after dev-d7w5-3 had advanced the sprint head to b9b9510: exit 2, 'rebasing bb/dev-d7w5-2 onto b9b9510 failed: git rebase exited 128: fatal: bb/dev-d7w5-2 is already used by worktree at /home/user/.bb-worktrees/dev-d7w5-2'. The first dispatch accepted cleanly because the sprint head had not moved, so this reproduces only under actual concurrency — the thing the whole worktree design exists to enable. Workaround: run `git rebase <sprint-head>` inside the worktree, then re-run accept, which then reports case: serialized and succeeds. The command should rebase inside the worktree, or remove the worktree before rebasing, since it owns both.

## Routing

Routed by the kind recorded at capture, not by reading the finding.

| kind | n | branch |
|---|---|---|
| `decision` | 17 | **project** — a bead in this tracker |
| `scope` | 4 | **project** |
| `gap` | 3 | **project** |
| `stale` | 1 | **project** |
| `proof` | 7 | **terminates here.** Recorded beside what it supports; no issue |
| `tooling`, target `swarm` | 12 | **skill** — the `swarm` plugin (`brettcallaghan/swarm`), which this repository does **not** vend. `.sf/outbox/swarm/` |
| `tooling`, **target unrecorded** | 1 | **project**, by the merge test — see below |
| `environment` | 0 | — |
| `unverifiable`, `sequencing` | 0 | — |
| **unrouted** | **0** | every finding carried a parseable kind |

**45 harvested, 45 routed — 26 project, 12 skill, 7 terminating at `proof`.
Zero unrouted**, in both passes. The taxonomy covered everything this run hit,
and pass 2 exercised two kinds pass 1 never produced: `stale`, and a `tooling`
finding carrying no target.

`swarm.toml`'s `[retros] skill_issues_dir` is deliberately unset, and its
comment carries the split this routing rests on: this repository vends the `sf`
skills (`spec`, `groom-beads`, `groom-feature-blueprint`, `retro-spec`,
`retro-skill`), so a finding against one of those is an ordinary bead here; it
no longer vends the swarm plugin, so a finding against `swarm` cannot be acted
on here and goes to `.sf/outbox/swarm/` for delivery.

**Twelve of the thirteen `tooling` findings target the swarm plugin.** None
targets an `sf` skill, so the outbox is this run's entire skill-audience output
and this repository's own tracker takes nothing from that branch.

**The thirteenth, `d7w5.11#4`, carries no target, and the merge test resolves
it to the PROJECT.** It reports that `go vet`'s structtag analyzer forbids two
promoted fields repeating one json tag — which is the exact shape whose runtime
behaviour a differential test needed to pin. Where would a fix land? Not in a
skill repository: `go vet` is the Go toolchain's, and the only thing anyone here
can change is this repository's own declared `vet` check and the test that works
around it, both of which the finding already documents. It is routed to the
project and raised as no candidate; the residual is recorded under "what the run
could NOT verify".

## Validation

- **Baseline** (`manifest.checks_baseline`, untouched tree): all 12 declared
  checks passed, exit status 0.
- **At head**: all 12 pass. Recorded twice, by two agents, and **not re-run by
  this retro** — the skill forbids a second pass over the run, so the claim is
  inherited, not measured here. Review pass 1 ran
  `python3 <swarm scripts>/bb-checks.py` → rc=0 against the pass-1 merged tree
  (recorded on the review phase bead with the placeholder `{SCRIPTS}`
  unresolved; the run resolved it to
  `/root/.cache/swarm-src/plugin/skills/swarm/scripts`). Review pass 2
  independently confirmed all 12 exit 0 at **`62e9e42`**, the current head. The
  coordinator reports pre-merge checks 4/4 exit 0 on each dispatch acceptance.
  **Same 12, same result as the baseline, at both ends of the run.**
- **Coverage** — not a declared check, but `AGENTS.md` requires it for a
  `cli/**` change and `cli.yml`'s `coverage` job enforces it. Regenerated
  **twice, once per half of the run, never per bead**: the plan forbade per-bead
  regeneration so the two concurrent pass-2 batches could not conflict on
  `README.md`, and three beads record that instruction as a `decision`
  (`d7w5.7#3`, `d7w5.8#4`, `d7w5.9#3`).
  - **After pass 1** (`1bb35e7`): 89.5% → **89.1%**, +313 statements;
    `cli/core/markdown` went 37 → 96 statements and 100% → 93.8%.
  - **After pass 2** (`62e9e42`): 89.1% → **89.5%**, 2,835 / 3,168 statements;
    `cli/cmd` 84.1% → **85.7%**. Read here from the committed `README.md`
    between its generated markers, which is what CI compares. That
    `scripts/coverage.sh --check` exits 0 against it is review pass 2's
    measurement and the coordinator's, inherited — this retro did not re-run
    the three instrumented test legs.
- **CI was not observed.** The branch is pushed; no workflow run was read back.

## Pushes

- `git log origin/claude/swarm-agent-plugins-5wxrfo..HEAD` is **empty** —
  nothing is stranded on the designated branch. Head is **`62e9e42`**. Measured
  at the run's end.
- **The sprint branch was never pushed, deliberately.** Only the
  designated branch may be pushed in this environment, so every dispatch was
  accepted with `--no-push` and merged by hand. Measured at the run's end: the
  sprint branch's tip is `cbfb438` and
  `git rev-list --left-right --count <sprint branch>...HEAD` → `0 1`,
  so it is an ancestor of the pushed head and carries nothing head does not.
  (Pass 1's own figure, for the record it was taken in: the tip was then
  `ef94bd1`, likewise an ancestor.)
- The pre-rebase shas of `.7`, `.8` and `.9` are reachable from no branch and
  were never pushed; the history behind head is linear.
- `merge_to_default=false` and `pull_request=none`, both authorized: **no pull
  request was opened and nothing reached the default branch.** Whatever happens
  to this branch is a human's next move, not the run's.

## What the run could NOT verify

Named, one per line. An unstated gap is indistinguishable from one that does
not exist.

- **SETTLED IN PASS 2 — whether the 4 P1 review findings are reachable or
  latent.** Pass 1 could not say; all four have since been fixed and reviewed,
  each with a test that fails against the pre-fix code, and the answer differs
  per bead. Two changed real corpus output — `d7w5.8`'s fix took
  `Identifier`-vs-`mdlinks.slugifyAnchor` from 9 disagreements to 0 over 612
  `specs/**` headings, and removed leading hyphens rather than adding them.
  **`d7w5.9`'s fix is preventive, not corrective on this corpus**: review pass 2
  reported the span invariants holding across all 291 `.md` files both before
  and after, because no corpus heading currently sits inside a blockquote or a
  list item. Re-measured here rather than inherited —
  `grep -rEn '^[[:space:]]*([>]|[-*+]|[0-9]+\.)[[:space:]]+#{1,6}[[:space:]]' specs/ --include='*.md'`
  → rc=1, 0 matches over 58 files. The defect was real in the code and unreached
  by the tree, which is a reason to keep the fix, not to discount it; it is
  recorded here so nobody later reads the four as one uniform class.
- **The pass-2 acceptance figures.** "Pre-merge checks 4/4 exit 0" and
  `acceptance_failures: []` with `beads_checked` 3 and 1 come from the
  coordinator's report and are carried into this record unmeasured: no bead or
  phase epic holds them — `d7w5.3` and `d7w5.6` both have
  `comment_count` 0, so the pass-2 groom & plan and post-flight phases left no
  durable record of their own. What **is** measured here: all four beads closed,
  one reachable commit each, the tree clean, no worktree standing, and nothing
  stranded. Settle the rest by re-running the acceptance command, which is no
  longer possible — the worktrees are gone.
- **Whether the `omitzero` key divergence is reachable on
  this command surface.** Filed P2 by review pass 2 by analogy with the `,string` bead,
  whose finding states no type on this surface uses `,string`; no equivalent
  probe for `omitzero` was run here.
- **The both-tagged same-depth promotion conflict cannot be covered by a static
  test in this repository at all** while `vet` is a declared check
  (`d7w5.11#4`). The untagged form exercises the same `dominantField` branch, so
  the *rule* is covered and the *shape* is not. Settling it would need
  `reflect.StructOf` to build the type at runtime.
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
- **Whether CI is green on head, `62e9e42`.** Checks pass locally; no workflow
  run was read back, in either half of the run.
- **The recurrence count on any of the 16 filed issues.** All 16 carry
  `recurrence: 0`, which is the template's starting value rather than a
  measurement: no window over prior runs' UNFILED findings was computed. This
  repository's earlier runs did report findings against the sprint skills while it
  still vended them, and those were filed as beads here — filed findings do not
  count toward recurrence — but that was not checked either. Settle it by
  harvesting `retro:` comments across the prior runs in the window and counting
  those with no `retro:filed:` mark for the same kind and target.
- **Whether the 12 declared checks pass at head, `62e9e42`, specifically.** The
  green result is inherited from the reviewer's and the coordinator's records —
  review pass 2 states it measured all 12 at exactly this commit — and this
  retro deliberately does not re-run checks. Settle it with
  `python3 <swarm scripts>/bb-checks.py` on the current tree. (Pass 1 recorded
  the same gap against the then-head, `1bb35e7`.)

## Candidates

Reasoned from the findings above, deduped against what is already filed, and
ranked. **This section is the list AS PUT to the human; the decision and what
was filed for each are recorded under "The decision, and what was filed" below.**
Each was put to a human once; a declined candidate stays in this document, where
a later run hitting the same friction raises it again with a higher recurrence
count.

**Nineteen candidates in two sets, and both sets are now answered.** The next
two subsections are **pass 1's fifteen**, answered at `AMENDMENT 3` and
unchanged; the third is **pass 2's four**, answered in a second decision round
on 2026-09-04 — **S11, S12 and P6 accepted and filed; A1 HELD**, neither filed
nor declined, pending a root-cause investigation already under way on
the YAML-rendering performance regression bead.

Provenance carried by every candidate: this run, branch
`claude/swarm-agent-plugins-5wxrfo`, record this file, model `opus` (advanced
tier per `swarm.toml` `[models]`) for the work-bead findings; the coordinator's
model is **unrecorded** — see the could-not-verify list.

### Pass 1, skill audience → `.sf/outbox/swarm/` (the swarm plugin)

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

### Pass 1, project audience → beads in this tracker

| # | candidate | findings | rank |
|---|---|---|---|
| P1 | **A model or schema change cannot be scoped to "declare and regenerate", and grooming does not say so.** Three of the four beads under-declared their surface in the same direction: a Strata declaration forced the emitter, two fixtures and three tests; a wiring-only item had to add a function to the previous item's package; a command item made a standards document's status text false. `c4_docs` is what the partitioner serializes on, so under-declaration silently defeats batch serialization — it cost nothing here only because the partition had zero parallelism. Proposed: `groom-beads` gains a rule that a bead touching a generated model declares its emitter, its fixtures, its snapshots and any status-bearing standards document in `c4_docs`. **`groom-beads` is vended by this repository, so this is a bead here, not an outbox file.** | `f6tl.1#2`, `f6tl.3#2`, `f6tl.4#4` | **P1** |
| P2 | **`.sf/outbox/README.md` and `.claude/skills/triage-issue/SKILL.md` both assert "we vend the skills we run, so we are our own downstream", which `swarm.toml`'s `[retros]` comment and an existing bead say is no longer true.** `triage-issue` sweeps `.sf/outbox/` into this tracker without anyone forwarding anything — so the ten swarm-plugin issues above would be swept straight back into the backlog this repository cannot act on, which is precisely the accumulation the audience split exists to prevent. Proposed: both documents state that `.sf/outbox/swarm/` is for **delivery to `brettcallaghan/swarm`** and is not swept here. | `d7w5.2.1#1`–`#11` (their existence is the evidence), plus the two documents quoted | **P1** |
| P3 | **`cli-output-standards.md` carries a hand-maintained conformance table and a status line that were false from the day they were written**, and nothing checks them. The document says "A command specified after this document is written to conform from the start" while asserting "No command conforms today" — so the first conforming command makes it untrue, and the only thing that catches it is an agent noticing. Proposed: either derive the table, or replace the corpus-wide assertion with a per-command row nothing else has to remember. | `f6tl.4#4` | P2 |
| P4 | **Three durable `bd` 1.2.2 facts belong in `AGENTS.md`'s environmental-failure ledger**, where the next agent looks: (a) `bd show <id> --json` returns `.comments` = **null** while the same object reports a non-zero `comment_count` and the text view prints every body — **re-measured by `retro-d7w5` on 2026-09-04**: `bd show <gate task> --json | jq '.[0].comments'` → `null`, `jq '.[0].comment_count'` → `12`; (b) `bd` counts a **deferred** dependency as blocking, so a bead whose blockers were deferred can never enter `bd ready` — **re-measured by `retro-d7w5` on 2026-09-04**: `bd blocked --parent <target epic>` → `f6tl.7: Blocked by 2 open dependencies: [f6tl.6 f6tl.5]`, and `bd export` reports both of those `deferred` (the count is 2 rather than AMENDMENT 1's 6 only because `.1`–`.4` have since closed); (c) `bd human respond` comments *and* closes and fails with `storage is nil` — **inherited from `f6tl.7`'s bounce comment and not re-measured here**, and it must be marked unverified in the ledger if filed. | `d7w5.2.1#9`, AMENDMENT 1, `f6tl.7` | P2 |
| P5 | **The two `gap` findings name exactly the two areas review then returned P1 defects against**, and both are places a blueprint declined to decide: whether a heading inside a blockquote or list item is reported and how its section span is bounded (`f6tl.2#3`), and that a frontmatter block's closing `---` is a setext heading so a whole-file search reports the block as a level-2 heading (`f6tl.3#1`). The already-filed `d7w5.7`, `d7w5.9` and `d7w5.12` fix the **code**; none of them says the module blueprint must state the rule. Proposed: `specs/modules/document-core/document-core-module.md` decides both cases explicitly. **Overlaps the three deferred review beads — decline if they are considered sufficient.** | `f6tl.2#3`, `f6tl.3#1` | P3 |

**Considered and not raised as candidates.** The eleven `decision` findings are
each already recorded where they belong — in `specs/modules/document-core/document-core-module.md`,
`specs/containers/spec-cli/spec-cli-container.md`, or a doc block beside the
code — by the agent that made them. Their volume is a signal about how much the
structure blueprint left open, but "eleven decisions were made at
implementation time" is a summary, not a change anyone can accept or decline.
They stay in this document so a second run reporting the same shape can be
counted against them.

### Pass 2 — four further candidates, PUT TO THE HUMAN AND ANSWERED

Reasoned from pass 2's seventeen findings, deduped against everything pass 1
filed and against the nine beads filed outside the run. **The list below is as
PUT**; the second decision round's answer and what was filed for each are under
"Pass 2's decision" in the next section. In summary: **S11, S12 and P6
accepted and filed on 2026-09-04**, their `retro:filed:` marks written, and the
manifest counters moved once as one merge; **A1 held** — the human answered
"investigate the cause but don't commit to a resolution yet", so
the YAML-rendering performance regression bead was not touched at all.

| # | audience | candidate | findings | rank |
|---|---|---|---|---|
| **S11** | skill → `.sf/outbox/swarm/` | **`swarm dispatch accept` cannot accept the SECOND of two concurrent dispatches.** It rebases the dispatch branch while that branch is still checked out in the worktree the command itself created. Measured, accepting `dev-d7w5-2` after `dev-d7w5-3` had advanced the sprint head: exit 2, `fatal: bb/dev-d7w5-2 is already used by worktree at /home/user/.bb-worktrees/dev-d7w5-2`. It reproduces **only under real concurrency** — the thing the whole worktree design exists to enable — and the first accept was clean only because the sprint head had not moved. Workaround used: rebase inside the worktree, then re-run accept, which reports `case: serialized` and succeeds. Proposed: rebase inside the worktree, or remove the worktree before rebasing, since the command owns both | `d7w5.2.1#12` | **P0** |
| **A1** | project → an **existing** bead | **The run introduced and shipped a performance regression against a specified target, and it is filed only as a P2.** The regression: `d7w5.11`'s port of `encoding/json`'s `typeFields` is faithful but drops the one thing `encoding/json` does around it — the per-`reflect.Type` memo — so the field set is re-resolved for every struct **value**, and `isBlock` pays for it a second time on the same value. `AC-SPEC-009.1` sets 50 ms for a warm corpus query; the YAML rendering alone now costs ~40 ms of that budget on this corpus. A memoized prototype is faster than the pre-run code. Proposed: raise that bead to **P1** and record this retro's independent reproduction on it. **Not a new bead — it already exists**, so this is a priority change and a comment, not a filing | `d7w5.11#3` (its origin), the YAML-rendering performance regression bead | **P1** |
| **P6** | project → a bead here | **A work order must not state a value or a signature it DERIVED by inference from a measurement.** In three of the four pass-2 beads what the bead *reported* was right and what it *prescribed* was wrong or unreachable: `d7w5.8`'s Design gave two of four post-filter identifiers incorrectly while its measured render table was byte-correct; `d7w5.7`'s acceptance criterion was unreachable given the signature its own Design pinned; `d7w5.11`'s three named fixes would have left the guarantee still false. Each cost the implementing agent an AC amendment or a mid-bead scope expansion, and each is exactly `AGENTS.md`'s "tool claims are measured at the point of use" one document earlier. Proposed: `specs/standards/work-order-definition-of-ready.md` — which `scripts/sync-shared-references.sh` vendors into both `groom-beads` and `spec`, so one edit reaches both — requires that a Design or AC stating a literal expected value cite the command that produced it, and marks a derived value **unverified** otherwise | `d7w5.8#2`, `d7w5.8#4`, `d7w5.7#2`, `d7w5.11#3` | **P1** |
| **S12** | skill → `.sf/outbox/swarm/` | **`swarm-review` files a finding as a bead carrying a prescribed remedy the reviewer did not measure.** The same four findings read at their origin rather than at the gate: the reviewer's *evidence* was correct in all four cases and its *suggested shape* was wrong or incomplete in three. A bead reads as already-decided work, so the prescription gets acted on rather than re-derived. Proposed: a review-filed bead states the defect and its evidence, and marks any suggested remedy explicitly unverified. **Overlaps P6** — P6 is the gate that would catch this, S12 is the source that produces it; decline either half if the other is judged sufficient | `d7w5.8#2`, `d7w5.8#4`, `d7w5.7#2`, `d7w5.11#3` | P2 |

#### A1's measurement, reproduced by this retro rather than inherited

`retro-d7w5` re-ran it on 2026-09-04 against two scratch trees extracted with
`git archive` — `ec6b6fd` (before `d7w5.11`) and `62e9e42` (head) — with one
identical benchmark added to each: `renderYAML` over a `headingsResult` of
2,434 `headingMatch` entries, this corpus's own heading count.
`go test ./cmd/ -run '^$' -bench BenchmarkRetroRenderYAML -benchtime 20x -count 3`,
go1.24.7, 4-core container:

| tree | ns/op, three runs | B/op | allocs/op |
|---|---|---|---|
| `ec6b6fd` — before `d7w5.11` | 24,555,037 / 23,476,837 / 24,732,037 | 11,450,768 | 130,988–990 |
| `62e9e42` — head, as shipped | 41,520,183 / 41,570,698 / 41,224,854 | 18,485,133 | 211,346–347 |
| head + a `sync.Map` memo on `typeFields` | 17,492,805 / 16,892,487 / 16,681,633 | 8,800,329 | 48,218–220 |

**~1.7x slower and 1.61x more allocations than before the run touched it**, and
the memoized prototype is **~1.4x faster than the pre-run code**. The allocation
counts reproduce review pass 2's reported 130,969 → 211,325 → 48,199 to within
the handful my fixture differs by; the wall-clock figures run higher than its
22.7 / 39.8 / 14.7 ms on the same class of container, so **the ratio is the
durable fact and the absolute numbers are indicative** — which is what the
`yztv` bead itself says. `go test ./cmd/` exits 0 with the memo in place, so the
differential suite `d7w5.11` shipped still holds. The mechanism is visible in
the source without a benchmark at all: `cli/cmd/yaml.go` contains no `sync.Map`
and no cache, and `structFields` is called once by `writeYAMLValue` and again by
`isBlock` for the same value.

**Considered in pass 2 and NOT raised as candidates.** `d7w5.11#4` — `go vet`'s
structtag analyzer forbidding the both-tagged same-depth shape — is a residual
already documented in the finding and beside the test, with nothing for anyone
to accept or decline; it is recorded under what the run could not verify. The
three identical coverage-left-stale `decision` findings record the plan working
as designed. The six `proof` findings terminate at routing. The five other beads
filed outside the run — `zka9`, `1j2i`, `912t`, `s5ok`, `rk27` — are already
filed, and re-filing them here would be a duplicate.


## The decision, and what was filed

**Pass 1's fifteen candidates** were put to the human once, at the retro human
channel, and answered there: **13 accepted, 2 declined.** Recorded on the run
epic as `AMENDMENT 3`, which also admitted the four P1 review findings that
produced the second half of this run. Filing was carried out by `retro-d7w5` on
2026-09-04; nothing below was filed before the decision. **Pass 2's four
candidates were put once and answered in a second round** — 3 accepted and
filed, 1 held; see "Pass 2's decision" at the end of this section.

`manifest.retro_filed` = **16**, `manifest.retro_declined` = **2**,
`manifest.retro_routed` = **45** — written as one merge with
`swarm manifest set --run <run epic>`, once, after the second round came
back. The figures cover both rounds together: 13 filed in pass 1 plus 3 in pass
2, the two pass-1 declines unchanged (a held candidate is not a decline), and
every one of the run's 45 findings routed.

### Accepted — skill audience, all 10, delivered to `brettcallaghan/swarm`

One file per issue in `.sf/outbox/swarm/`, in the issue template's shape. **None
of these is a bead in this tracker**: this repository does not vend the swarm
plugin, so a proposal about it cannot be acted on here.

| # | rank | finding | file (under `.sf/outbox/swarm/`) |
|---|---|---|---|
| S1 | P0 | `d7w5.2.1#10` | `2026-09-04-swarm-review-parents-findings-under-the-run-epic.md` |
| S2 | P0 | `d7w5.2.1#3` | `2026-09-04-pre-flight-check-never-writes-models-or-ledger.md` |
| S3 | P0 | `d7w5.2.1#11` | `2026-09-04-phase-close-has-no-path-through-an-unclosable-target-epic.md` |
| S4 | P1 | `d7w5.2.1#1` | `2026-09-04-authsource-has-no-human-member.md` |
| S5 | P1 | `d7w5.2.1#2` | `2026-09-04-standing-auth-has-no-home-for-dispatch-failure-threshold.md` |
| S6 | P1 | `d7w5.2.1#7` | `2026-09-04-phase-close-cannot-close-a-childless-ceremony-epic.md` |
| S7 | P1 | `d7w5.2.1#8` | `2026-09-04-swarm-status-counts-ceremony-beads-as-scope-creep.md` |
| S8 | P2 | `d7w5.2.1#4` | `2026-09-04-manifest-validate-renders-severitypending-as-not-ok.md` |
| S9 | P2 | `d7w5.2.1#5` | `2026-09-04-manifest-set-type-errors-name-the-go-type-not-the-shape.md` |
| S10 | P2 | `d7w5.2.1#6` | `2026-09-04-no-way-to-list-the-declared-manifest-keys.md` |

Every one carries its finding quoted inline, not only cited, and the run's
provenance: this run, branch `claude/swarm-agent-plugins-5wxrfo`,
this record, `skill_version` `swarm 0.1.0 @ 1a19e8e (2026-09-02)`. The `model`
field reads **unrecorded** on all ten, with the reason — `manifest.models` is
unset, which is candidate S2's own defect. `recurrence` is written as **0** on
all ten, and that figure is **not measured** — no recurrence window was computed
over prior runs' unfiled findings. It is the template's starting value for a
kind/target pair, not a finding that this pair is new; see the could-not-verify
list.

**Delivery is not filing.** `.sf/outbox/swarm/` is a delivery channel to
`brettcallaghan/swarm`; carrying it there is the outbox owner's call. Candidate
P2 below exists because two documents in this repository currently say the
opposite.

### Accepted — project audience, 3 beads in this tracker

**None is parented under the run epic.** Parenting findings under the run epic
is what put nine unadmitted beads into this run's own dispatch set (`d7w5.2.1#10`,
candidate S1); these carry `discovered-from` edges to their origin beads instead,
which records provenance without feeding dispatch. Each was groomed to the
work-order Definition of Ready — problem with durable anchors, design with the
choice settled, checkable acceptance criteria, dependencies, tier and reason, and
declared blast radius.

| # | rank | title |
|---|---|---|
| P1 | P1 | groom-beads has no rule that a "declare and regenerate" bead must declare the emitter, fixtures, snapshots and standards its regeneration reaches |
| P2 | P1 | The outbox README and triage-issue still say we vend every skill we run, so triage would sweep the swarm deliveries into this backlog |
| P3 | P2 | `cli-output-standards`' conformance table and status line are hand-maintained, unchecked, and false again |

Notes on each, where the filed bead differs from or extends the candidate:

- **The groom-beads regeneration bead** (findings `f6tl.1#2`, `f6tl.3#2`, `f6tl.4#4`) lands the
  rule in `specs/standards/work-order-definition-of-ready.md` check 8 rather than
  in `groom-beads`' own body — measured, `plugin/skills/groom-beads/SKILL.md` is
  19,825 bytes against `scripts/check-skills.sh`'s `BODY_BYTE_MAX = 20000`, so
  the rule does not fit there. `scripts/sync-shared-references.sh` vendors that
  standard into both `groom-beads` and `spec`, so one edit reaches both. Carries a
  `related` edge to the review finding `d7w5.15`, which it partially overlaps, and an
  acceptance criterion requiring that overlap to be resolved rather than left.
- **The outbox README bead** (findings `d7w5.2.1#1`–`#11`, whose existence is the
  evidence). **Both file locations were re-verified before filing, because a
  variant of this candidate in circulation named the wrong one.** Measured
  2026-09-04: `ls plugin/skills/triage-issue` → "No such file or directory", and
  `plugin/skills/` holds exactly `groom-beads`, `groom-feature-blueprint`,
  `retro-skill`, `retro-spec`, `spec` — five vended skills, `triage-issue` not
  among them. The file that carries the assertion is
  `.claude/skills/triage-issue/SKILL.md`, a real directory rather than a symlink
  into `plugin/skills/`, so `triage-issue` is repo-local and not vended; the text
  is its `## Inputs` bullet on `.sf/outbox/<skill>/*.md`. The second site,
  `.sf/outbox/README.md`'s "## In this repository" section, was confirmed as
  cited. The bead anchors on the quoted strings, not on line numbers.
- **The conformance-table bead** (finding `f6tl.4#4`). The candidate offered two
  remedies — derive the table, or replace the corpus-wide assertion — which is an
  open decision and would have failed the Definition of Ready's check 9, so the
  bead settles it: the table is **not** derived (that would mean asserting in code
  a conformance judgement nobody measured), the universal claims are deleted, and
  a new declared check enforces the part that is mechanizable — that the table's
  row set matches the command set the CLI registers. Measured while filing: the
  preamble is **false again today**, reading "Every command below predates these
  standards and none of them conforms yet" directly above a table whose `links`
  row reads "conforms". That is the second falsification of this document in one
  run, and it is on the bead.

### Declined — 2

Recorded, not deleted. A later run hitting the same friction raises it again with
a higher recurrence count, which is how a wrong decision corrects itself with
nobody tracking it.

| # | candidate | why it stays here |
|---|---|---|
| P4 | Three durable `bd` 1.2.2 facts for `AGENTS.md`'s environmental-failure ledger | Declined. The measurements stand and are recorded in the candidate above — including the two re-measured by `retro-d7w5` and the one that must be marked unverified if ever filed — but they are not being added to the ledger |
| P5 | The module blueprint should decide the two cases the `gap` findings name | Declined. The three already-filed review beads (`d7w5.7`, `d7w5.9`, `d7w5.12`) are considered sufficient; note that `.7` and `.9` have since been admitted by `AMENDMENT 3` and `.12` remains deferred |

The eleven `decision` findings were never candidates — each is already recorded
where it belongs, by the agent that made it — and the one `proof` finding
terminated at routing. Neither is a decision anyone declined.

### Pass 2's decision — 3 accepted and filed, 1 held

The four candidates above (**S11**, **A1**, **P6**, **S12**) were put to the
human at this retro's channel once, and answered on 2026-09-04. Filing was
carried out by `retro-d7w5` after the answer; nothing was filed before it.

| # | rank | decision | finding(s) | what was filed |
|---|---|---|---|---|
| S11 | P0 | **accepted** | `d7w5.2.1#12` | `.sf/outbox/swarm/2026-09-04-dispatch-accept-rebases-a-branch-still-checked-out-in-its-worktree.md` |
| S12 | P2 | **accepted** | `d7w5.9#2`, `d7w5.11#3` | `.sf/outbox/swarm/2026-09-04-review-filed-beads-prescribe-a-remedy-the-reviewer-never-ran.md` |
| P6 | P1 | **accepted** | `d7w5.9#2`, `d7w5.8#2`, `d7w5.7#2`, `d7w5.11#3` | bead (P1, `parent=null`) |
| A1 | P1 | **HELD** | `d7w5.11#3`, the YAML-rendering performance regression bead | nothing — see below |

**S11** and **S12** are skill-audience and went to `.sf/outbox/swarm/` for
delivery to `brettcallaghan/swarm`, in the issue template's shape and with the
same provenance the pass-1 ten carry: this run, branch
`claude/swarm-agent-plugins-5wxrfo`, this record, `skill_version`
`swarm 0.1.0 @ 1a19e8e (2026-09-02)`, `model` **unrecorded** with the reason,
and `recurrence: 0` as the template's starting value rather than a measurement.
S11 quotes `d7w5.2.1#12` in full and, beside it, the run epic's `PASS 2
ACCEPTANCE EVIDENCE` comment, which records both accepts side by side. S12
cites the three measured instances the human named: `d7w5.9`'s finding proposed
asking `heading.Type()` / the AST kind and goldmark v1.8.5's `ast.Heading`
carries only `BaseBlock` and `Level` (`d7w5.9#2`); the same finding mis-stated
part of its own defect, and `d7w5.9`'s Problem section records the
grooming correction — "A's Section is EMPTY" is correct behaviour, A stands over
nothing; and `d7w5.11`'s three named fixes left field dominance unhandled and
would have produced the key `name` twice carrying two different values
(`d7w5.11#3`).

**S12 and P6 overlap deliberately and both were accepted.** P6 is the readiness
gate that would catch such a work order at grooming; S12 is the source that
produces it. Each says so on its face, and neither is offered as a substitute
for the other.

**The P6 bead** requires that a work order state as fact
only what was measured: any value, signature, identifier or enumerated fix
derived by inference either carries the command and output that establish it, or
is marked unverified in place. It lands as a tenth check in
`specs/standards/work-order-definition-of-ready.md`, which
`scripts/sync-shared-references.sh` vendors into both `groom-beads` and `spec`,
so one edit reaches both — the bead's acceptance criteria require the version
bump, both regenerated copies and `--check` exiting 0. It carries the four
measured instances quoted inline: the unreachable AST field (`d7w5.9#2`), the
identifier arithmetic that gave `--bullet` and `3--design-present` where the
Design said `-bullet` and `3-design-present` (`d7w5.8#2`), the `Locate(src)
(span.Region, bool)` signature that could not distinguish "no block" from
"opened and never closed" (`d7w5.7#2`), and the three named fixes that asserted
a sufficiency they did not have (`d7w5.11#3`). Groomed to this repository's own
Definition of Ready — problem with durable anchors, design with the choice
settled and its alternatives recorded, checkable acceptance criteria,
`tier: advanced` with a reason, declared C4 blast radius. **`parent=null`**, with
`discovered-from` edges to `d7w5.7`, `.8`, `.9` and `.11` and a `related` edge to
the groom-beads regeneration bead, which edits the same standard: parenting a retro filing
under the run epic is the P0 defect this run reported as `d7w5.2.1#10`.

**A1 is HELD — neither filed nor declined.** The human answered "investigate the
cause but don't commit to a resolution yet". The regression bead was therefore
**not touched at all**: its priority still reads **P2**, no remedy was recorded
on it, and no comment was added — a separate agent is investigating its cause and
owns that bead. The measurement behind A1 stands and is recorded above, in
"A1's measurement, reproduced by this retro rather than inherited"; it is
evidence for whoever takes the investigation, not a decision anyone made. A held
candidate is not a decline and is not counted as one: `retro_declined` stays at
pass 1's **2**. If the investigation settles that the regression is worth fixing
at P1, this candidate is where the reproduction is written down.

### The audit closes both ways

Each filed issue is marked on the bead carrying its finding, as a
`retro:filed:<id> — <finding id>` comment. **Pass 1:** ten on
`d7w5.2.1` for the outbox deliveries, plus the groom-beads bead on `f6tl.1`,
`f6tl.3` and `f6tl.4`, the outbox README bead on `d7w5.2.1`, and the
conformance-table bead on `f6tl.4`. **Pass 2:**
S11's outbox file on `d7w5.2.1` (for `#12`); S12's outbox file on `d7w5.9` (for
`#2`) and on `d7w5.11` (for `#3`); and the P6 bead on all four pass-2
beads — `d7w5.9`, `d7w5.8`, `d7w5.7` and `d7w5.11`, each naming the finding it
rests on. Nothing was marked for **A1**, which was held rather than filed, so
`d7w5.11#3` remains unfiled for that candidate and keeps counting toward
recurrence. Recurrence counts unfiled findings only, so the pool shrinks as
issues are filed rather than growing without limit.

### Counts

**Pass 1, unchanged:** harvested 28, routed 28 (16 project, 11 skill, 1
terminating at `proof`), unrouted 0; 15 candidates, 13 filed, 2 declined.

**The run, at its end:**

- **Harvested 45** — 28 in pass 1, 17 in pass 2. Every count verified against
  its bead's `comment_count` before aggregating.
- **Routed 45** — **26 project**, **12 skill**, **7 terminating at `proof`**.
  **Unrouted 0.**
- **Candidates 19** — 15 in pass 1, 4 in pass 2. By audience: 12 skill, 7
  project.
- **Filed 16** — 13 in pass 1 (10 outbox files, 3 beads) and 3 in pass 2 (2
  outbox files, 1 bead). **Declined 2**, both pass 1's.
  **Held 1** — candidate A1, pending a root-cause investigation on
  the regression bead; a held candidate is neither filed nor declined and is
  counted as neither. **Pending 0.**
- **Written to the manifest as one merge**, covering both rounds:
  `retro_filed` **16**, `retro_declined` **2**, `retro_routed` **45**.
- Nothing was filed in this repository's tracker from the skill branch, and
  nothing in `brettcallaghan/swarm`'s tracker from here.
- Beads filed **outside** this retro, by the dispatches and by review pass 2,
  all `parent=null` and all open: `zka9`, `1j2i`, `yztv`, `912t`, `s5ok`,
  `rk27`. This retro re-files none of them.
