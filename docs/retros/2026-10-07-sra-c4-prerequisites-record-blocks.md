# Run record — C4 prerequisites with structure record blocks

*Written by `retro-sra` on 2026-10-07. One document: the run's results, every
finding it captured, what it routed where, and what it could not establish.*

## Read this first

- **The review budget is EXHAUSTED, not unspent.** Three passes ran (budget 2 =
  three passes). Pass 3 filed `sra.8` and `sra.9` as run-introduced
  defects and they are **left open under the run epic, not worked**. The run
  finished over a narrower scope than its review produced.
- **Every tracker write this run was local-only until post-retro** (reconciled at 06:00Z; see Decisions and filings). The tracker's Dolt remote
  diverged mid-run (another session pushed `beads/data` at 05:05–05:07Z), so
  every `bd dolt push` failed non-fast-forward (manifest ledger entry 3). The
  closures, comments and this run's manifest keys exist only in this
  container's database until a human reconciles. Measured again at retro time:
  `git ls-remote origin beads/data` → `6c9e616…`, which is neither the ledger's
  `1e19020` nor the local tracking ref `6d389c6` — the remote has moved again
  since the ledger entry was written.
- **`f6tl.7` was excluded (deferred)**, so the structure-extraction epic is not closed.

## The run

| | |
|---|---|
| run epic | "Run: C4 prerequisites with structure record blocks" (open) |
| targets | the C4 prerequisites epic (closed) and the structure-extraction epic (open — `.7` deferred) |
| sprint branch | `claude/project-thread-es7x8o` — the session's designated branch, **recorded directly** on the manifest because `swarm dispatch open` refuses to adopt an existing branch (`amend:` 05:06:32Z) |
| `base_sha` → head | `b626128` → `66d22ec`, **ten commits** |
| queue | 3 admitted at the gate (`c4pre.1`, `f6tl.5`, `f6tl.6`) + 7 review findings admitted across two loops (`sra.1`–`.7`) = 10 admitted, 10 closed |
| dispatches | `dev-sra-1` (gate beads), `dev-sra-2` (`sra.1`–`.5`), `dev-sra-3` (`sra.6`, `.7`), plus `review-sra`, `coordinator-sra`, `retro-sra`. The first dev dispatch returned after **one** bead with its queue not drained and had to be resumed (coordinator report) |
| `concurrency_cap` | 20 |
| declared checks | 12 — `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| tiers | 7 `standard`, 3 `light` among the closed set (`f6tl.7`, excluded, is `advanced`) |
| model lineup | `manifest.models`: `frontier`→`fable`, `advanced`→`opus`, `standard`→`sonnet`, `light`→`haiku`. **Which model served each dispatch is recorded nowhere.** `manifest.effort` = `unrecorded` |
| review passes | `review_passes: 2` loops → **3 of 3 passes; budget exhausted** |
| circuit breaker | `consecutive_dispatch_failures` 0 at every accept (`swarm ready --completed` after each dispatch) — never tripped |
| trailers | `swarm manifest get --run <run epic> --key trailers` → `ok: false, value: null`. Closes ran **without** `--trailers-from`; `git log --format='%h\|%(trailers:only,unfold)\|' b626128..HEAD` prints ten empty trailer blocks |

**Standing authorization** (`manifest.standing_auth`): `scope` from the human
("run epic [this run] over [the C4 prerequisites and structure-extraction epics], chosen by Brett on
decision card 2026-10-07T05:02Z"); `review_budget` 2, `tier_escalation` once,
`merge_to_default` false, `move_foreign_beads` false,
`dispatch_failure_threshold` 3 — all `default`.

**Ledger** (`manifest.ledger`, three entries): the tracker remote is
`refs/heads/beads/data`, never `refs/dolt/*`; never `bd dolt pull` / `bd
bootstrap` / `bd init`; and the mid-run divergence (`git fetch origin
beads/data: c1056c6..1e19020`; `bd dolt push` at 05:12Z rejected, tip behind
remote).

## Results

| bead | P | status | agent | commit | what |
|---|---|---|---|---|---|
| `f6tl.5` | 1 | closed | `dev-sra-1` | `e602237` | A profile registers record-block tags (`Profile.RecordBlocks()`), pointers `#/definitions/<Name>` |
| `c4pre.1` | 1 | closed | `dev-sra-1` | `07fa867` | Registered the `component` record-block tag in the spec-driven-dev profile |
| `f6tl.6` | 1 | closed | `dev-sra-1` | `4638d28` | Records registered record blocks and attaches them to their heading; tags folded into the input hash |
| `sra.1` | 2 | closed | `dev-sra-2` | `56fa3db` | `DecodeRecordBlock` decodes records inside containers (blockquote) |
| `sra.2` | 2 | closed | `dev-sra-2` | `3503343` | Record blocks searched from the body, not frontmatter |
| `sra.3` | 3 | closed | `dev-sra-2` | `d505122` | Unparsable profile schema no longer reported as every pointer unresolved |
| `sra.4` | 3 | closed | `dev-sra-2` | `899ead6` | `resolvePointer` follows RFC 6901 / URI-fragment rules, arrays included |
| `sra.5` | 3 | closed | `dev-sra-2` | `e6d4938` | `ValidateRegistration` reports a tag registered twice |
| `sra.6` | 3 | closed | `dev-sra-3` | `80a4b5f` | `DecodeRecordBlock` re-locates from the body, as the index does |
| `sra.7` | 4 | closed | `dev-sra-3` | `66d22ec` | Frontmatter-fence test no longer indexes `found[0]` in its failure message |
| `sra.8` | 4 | **open** | — | — | Record-block decode test indexes `RecordBlocks` without a length guard (review pass 3; not worked) |
| `sra.9` | 4 | **open** | — | — | `DecodeRecordBlock` bounds guard overflows on `Offset+Length` (review pass 3; not worked) |

Review held further findings in the backlog, outside the run:
four from pass 1, six from pass 2 and two from pass 3.

**Shape of the review loop.** All nine run-introduced defects (`sra.1`–`.9`)
are in the two files the gate beads wrote — `cli/core/corpus/records.go` and
`registration.go` — and each pass found more. The loop stopped on budget, not on
a clean pass.

## What the run decided

Findings of kind `decision`, `scope`, `gap`, `sequencing`, first. **Zero `gap`,
zero `scope`, zero `sequencing`.**

- `f6tl.5#1` — pointer form `#/definitions/<Name>`, resolved by an
  in-package walker (rejected: compiling the schema with the jsonschema lib).
- `f6tl.5#2` and `f6tl.6#1` — **a decision reversed inside the
  run.** `.5` chose not to fold record-block tags into the corpus input hash
  ("nothing reads them yet"); `.6` folded them, because a stale warm index would
  otherwise be served, and hand-edited six snapshot `expected/manifest.json`
  files to absorb the new hash. Neither bead nor the structure blueprint pinned
  whether registered tags are a hash input.
- `sra.1#2`, `sra.6#1` — **both rest on a misreading, measured
  here.** Each agent reported that the review-filed bead's example had lost its
  fences (`"> component\n> name: x"`; `description: |\n  yaml\n  foo: 1`) and
  substituted a fenced variant. The stored text **has** the fences:
  `bd show sra.6 --json | jq -r '.[0].description' | grep -c '```'` → `2`,
  and the description reads `` description: |\n  ```yaml\n  foo: 1 ``. It is
  `bd show`'s **text** rendering on bd 1.3.1 that drops them: `bd show
  sra.6 > file` shows `description: |\n  yaml\n  foo: 1`. So `sra.6#1`'s
  "the bead's reproduction source does not reproduce the defect" is false —
  the variant it used is the stored reproduction. `sra.4#2` (`stale`) is
  the same mechanism: stored `"m~n"`, rendered `"m~~n"` (measured the same two
  ways). See candidate **B**.

## Scope

- **Excluded:** `f6tl.7` — "deferred; Definition-of-Done verification of
  the structure blueprint is not required by [the C4 prerequisites epic]". It still carries the
  2026-09-04 bounce from `plan-d7w5`.
- **Moved:** `[]`.
- **Admitted then dropped:** none. **Filed and left open:** `sra.8`,
  `sra.9` (budget exhausted).
- **The run is done over a NARROWED scope**: its admitted set is drained, but
  the structure-extraction epic stays open on `.7`, and two run-introduced defects stand.

## Every finding, verbatim, grouped by kind

Numbered `<bead>#<ordinal>` by position among that bead's `retro:` comments.
16 harvested from `bd export` over the run epic, the two target epics
and their children.

### `decision` (5)

> **`f6tl.5#1`** retro:decision Schema pointer form is '#/definitions/<Name>' (JSON Pointer URI fragment), resolved by a small in-package walker over Profile.Schema() JSON.
> evidence: grep -n '"definitions"\|C4ElementRecord' cli/core/schema/spec-model.schema.json -> line 3 '"definitions": {', line 544 '"C4ElementRecord": {' (draft-07 style, not $defs). Rejected: compiling the schema with jsonschema lib (heavier, needs the host to load a validator for a registration check).

> **`f6tl.5#2`** retro:decision Added Profile.RecordBlocks() []RecordBlockTag{Tag,Schema}; did not fold record tags into the corpus input hash (hash.go foldFeatureTags) since nothing reads them yet.
> evidence: cli/core/corpus/hash.go folds FeatureTags only; REC (f6tl.6) is where record blocks start affecting the index, so hash folding belongs there.

> **`f6tl.6#1`** retro:decision Folded the registered record-block tags into the corpus input hash (hash.go), reversing the earlier choice in f6tl.5 not to; now that tags change what the index records, a stale warm index would otherwise be served.
> evidence: go test ./cmd -run TestSnapshotCases failed on 6 index-* cases with only input_hash differing; no regeneration tool exists, so I replaced the input_hash in each cli/testdata/cases/index-*/expected/manifest.json by hand with the 'got' value from the test output. Alternative rejected: not hashing, which serves stale indexes after a profile registers a tag.

> **`sra.1#2`** retro:decision AC example text lacks fences ("> component\n> name: x"); tested the fenced form from the Steps to Reproduce instead. Also added Block.Text (not Content) as designed; scripts/coverage.sh run, README diff committed.
> evidence: README.md modified by scripts/coverage.sh exit 0

> **`sra.6#1`** retro:decision the bead's reproduction source does not reproduce the defect; I used a variant that does
> evidence: bead src has frontmatter 'description: |\n  yaml\n  foo: 1' (no fence). With records.go reverted, go test -run ARecordTheIndex PASSED. Changed the block scalar to '  ```yaml\n  foo: 1' (unterminated fence in frontmatter): pre-fix FAIL 'record block span [60,85) is not a fenced "component" block'; post-fix ok. Also the bead text lost the opening backticks of the component fence (```component). Rejected: keeping the bead's literal src (green pre-fix, proves nothing).

### `stale` (1)

> **`sra.4#2`** retro:stale AC rows render escapes garbled: "m~~n", "a~~1b" (doubled tilde) in the bead text. Read as RFC 6901 keys "m~n" and pointer tokens "a~1b" / "m~0n".
> evidence: bead text: `"m~~n"`, `"#/definitions/a~~1b" -> true`, `"#/definitions/m~0n" -> true`. Fixture uses key "m~n" and key "a/b"; test rows #/definitions/a~1b and #/definitions/m~0n both true.

### `tooling` (1)

> **`f6tl.6#2`** retro:tooling cli/testdata snapshot cases have no regenerate command; a hash-affecting change needs hand-edited expected/manifest.json files.
> evidence: grep -rn 'UPDATE\|-update' cli/cmd/snapshot_test.go -> nothing.

*Contradicted by the tree:* `scripts/regenerate-cli-snapshots.sh` exists and
its header says it writes every case's `expected/manifest.json`. The finding's
grep looked only in `snapshot_test.go`, which names no regeneration path. By the
merge test this is a **project** problem, not a skill one — see candidate **A**.

### `proof` (8)

> **`c4pre.1#1`** retro:proof Clean bead. Pointer resolution against the real embedded schema is proved by TestTheRegistrationHasNoDefects (ValidateRegistration now resolves #/definitions/C4ElementRecord); go test ./... and bb-checks.py exit 0.

> **`sra.1#1`** retro:proof pre-fix failure observed, then fixed
> evidence: with old records.go and new tests: TestDecodeRecordBlockDecodesRecordsInsideContainers/blockquote -> "[1:1] invalid header option: \" name: x\""; span-rejection test -> "an inner-content span decoded without error". After fix: go test ./core/markdown ./core/corpus ok; bb-checks.py rc=0.

> **`sra.2#1`** retro:proof pre-fix failure observed
> evidence: with recordBlocksOf(src, 0, ...) (whole-file search) the new test fails: "frontmatter fence: 1 headings, 1 unattached, 0 attached; want 1, 0, 0". With from passed: go test ./core/markdown ./core/corpus ok, bb-checks.py rc=0, coverage.sh rc=0 (README committed).

> **`sra.3#1`** retro:proof pre-fix failure observed
> evidence: with registration.go reverted, new test fails: "got 1 defects for an unparsable schema, want none: [{... Key:widget Detail:is registered as a record block validated against \"#/definitions/Widget\", which resolves to nothing ...}]". After fix: go test ./core/corpus ok; bb-checks.py rc=0.

> **`sra.4#1`** retro:proof pre-fix failures observed
> evidence: with registration.go reverted: resolvePointer("#definitions/W") = true, want false; ("#/allOf/0") = false, want true; ("#") = false, want true; ("#/definitions/sp%20ace") = false, want true; ValidateRegistration with "#definitions/Widget" returned [] (no defect). After fix all pass; bb-checks.py rc=0.

> **`sra.5#1`** retro:proof pre-fix failure observed
> evidence: with registration.go reverted: "same pointer twice: got [], want one defect keyed c"; "two pointers: got [], want one defect". After fix pass; bb-checks.py rc=0.

> **`sra.6#2`** retro:proof pre-fix failure observed and post-fix pass; coverage.sh run, README coverage diff committed; bb-checks.py all pass

> **`sra.7#1`** retro:proof clean bead: test-only guard; the body-fence block's message does not index found[0], so only the frontmatter-fence block changed. Targeted test and bb-checks.py pass. Not reproduced by forcing a panic (unverified).

### `environment` (0)

None captured by a dispatch. The run's one environmental fact — the diverged
Dolt remote — arrived through the manifest ledger, not a finding.

### No kind (1) — unrouted

> **`f6tl.5#3`** retro: n/a

A "clean bead" marker written with no kind. The protocol says a clean bead
records nothing; `retro:proof` is the form `c4pre.1` and `sra.7` used.

### Not findings, recorded from the run epic and the coordinator

These reached the retro as `amend:` comments, manifest keys and the
coordinator's report rather than as `retro:` comments, so they carry no finding
id. Each is the source of a candidate or a dedupe below.

1. **Designated branch recorded by hand** — `amend:` 05:06:32Z: "the harness
   permits pushing only that branch and swarm dispatch open refuses to adopt an
   existing branch."
2. **No `trailers` on the manifest** — measured above; closes ran without
   `--trailers-from`. Third consecutive run (anchor references,
   corpus listing, now this one).
3. **`groom-beads`' `bd-health.sh` false failure on bd 1.3.1.** Re-measured by
   this retro: `bash plugin/skills/groom-beads/scripts/bd-health.sh` → exit 1,
   `RESULT: hard problems found in: swarm validate`; it ran `bd swarm
   validate` on `f6tl.7` ("is not an epic or molecule (type: task)") and
   on **bead titles** ("epic 'Let a profile register record-block tags' not
   found") — the nested-key enumeration an existing `bd-health.sh` bead describes, now matching
   title-valued ids on bd 1.3.1.
4. **First dev dispatch returned after one bead, queue not drained**, and was
   resumed (coordinator report; `dev-sra-1` then closed `c4pre.1` and `f6tl.6`).
5. **Dolt remote divergence** — ledger entry 3.

## Routing

| branch | count | findings |
|---|---|---|
| project (six work-order kinds) | 6 | `f6tl.5#1`, `f6tl.5#2`, `f6tl.6#1`, `sra.1#2`, `sra.6#1`, `sra.4#2` |
| skill (`tooling`) | 1 | `f6tl.6#2` |
| `proof` → this document | 8 | listed above |
| `environment` → ledger | 0 | — |
| **unrouted** | **1** | `f6tl.5#3` (`retro: n/a`, no kind) |

Two routings change under the merge test, and the candidates follow the merge
test rather than the kind: `f6tl.6#2` (`tooling`) lands in this project's own
test harness (candidate A, project); `sra.1#2`, `sra.6#1`, `sra.4#2` (project
kinds) land in the `swarm-dispatch` skill's reading instruction (candidate B,
skill).

## Validation

- **Baseline:** all 12 declared checks passed on the untouched tree
  (`checks_baseline_status: 0`).
- **At head:** not re-run by the retro (the skill forbids re-running checks).
  Every closing bead's `proof` finding reports `bb-checks.py` rc=0; `sra.1`,
  `.2`, `.6` report `scripts/coverage.sh` run and README committed.
- **CI:** not observed.

## Pushes

- `git fetch origin claude/project-thread-es7x8o` then
  `git log --oneline origin/claude/project-thread-es7x8o..HEAD` → **empty**.
  Nothing on the code branch is stranded; head `66d22ec`.
- **The tracker is stranded.** Every `bd dolt push` this run failed
  non-fast-forward; closures of all ten beads, every `retro:` comment, the
  `amend:` comments, `sra.8`/`.9` and this run's manifest (including
  `findings_path`) are local-only. Remote `beads/data` at retro time:
  `6c9e616`.

## What the run could NOT verify

- `sra.7#1`: the guarded panic was "Not reproduced by forcing a panic (unverified)".
- Which model served each dispatch — only the tier→model map is recorded; `effort` is `unrecorded`.
- The first dev dispatch's early return — coordinator-reported; no transcript is durable.
- `swarm dispatch open` refusing an existing branch — recorded only as an `amend:` comment; not reproduced by the retro.
- The checks at head as a whole suite, and CI.
- Whether the tracker writes will survive reconciliation — depends on a human's merge of the diverged `beads/data`.

## Candidates

Each is the issue template's shape; frontmatter common to all:
`run: <this run>`, `branch: claude/project-thread-es7x8o`,
`report: docs/retros/2026-10-07-sra-c4-prerequisites-record-blocks.md`,
`model: dispatches resolved through manifest.models (standard→sonnet,
light→haiku); served model unrecorded; effort unrecorded`,
`updated_at: 2026-10-07T06:00:00Z`.

### A — Make the snapshot test's stale-manifest failure name the regeneration script
`audience: project` · `target: cli/cmd/snapshot_test.go` · `priority: 2` ·
`findings: ["f6tl.6#2", "f6tl.6#1"]` · `recurrence: 1`
(anchor references run: `rc14.1#4`, `rc14.2#3`, `rc14.4#4`)
· destination: **bead**, `discovered-from:f6tl.6`

**What is wrong, and what it costs.** When an input-hash change stales snapshot
cases, `TestSnapshotCases` prints only expected vs got; nothing on that path
names `scripts/regenerate-cli-snapshots.sh`. An agent searching the test file
for a regeneration hook finds none and hand-edits fixtures. Second run in a
row: six `expected/manifest.json` files hand-patched here, three times over in
`59ic`. testing-standards documents the fact; nobody reads that
at the moment of failure.

**Evidence.** `f6tl.6#2` — "cli/testdata snapshot cases have no
regenerate command" (`grep -rn 'UPDATE\|-update' cli/cmd/snapshot_test.go` →
nothing); `f6tl.6#1` — "replaced the input_hash in each
cli/testdata/cases/index-*/expected/manifest.json by hand". Measured by the
retro: `ls scripts/ | grep -i regen` → `regenerate-cli-snapshots.sh`;
`grep -n regenerate cli/cmd/snapshot_test.go` → nothing.

**Proposed change.** When a snapshot case's expected artifact differs, the
failure message ends with `regenerate with scripts/regenerate-cli-snapshots.sh,
then review the diff`. Done when `grep -n regenerate-cli-snapshots
cli/cmd/snapshot_test.go` matches the failure path.

### B — swarm-dispatch: read a bead with `bd show --json`, not the text rendering
`audience: skill` · `target: swarm-dispatch` · `priority: 1` ·
`findings: ["sra.6#1", "sra.1#2", "sra.4#2"]` ·
`recurrence: 0` · destination: **`outbox/swarm/swarm-dispatch/2026-10-07-read-beads-as-json-not-rendered-text.md`**

**What is wrong, and what it costs.** `swarm-dispatch` says "Read each bead in
full with `bd show <id>`". On bd 1.3.1 the text output renders Markdown: it
drops triple backticks inside inline code and doubles `~`. Three of seven
review-filed beads were misread this run; `sra.6`'s agent concluded a correct
reproduction "does not reproduce" and rewrote it, and `sra.4`'s agent had to
guess RFC 6901 tokens from garbled rows. A bead's code examples are exactly the
text that must survive verbatim.

**Evidence.** The three findings quoted above. Measured by the retro, bd 1.3.1:
`bd show sra.6 > f` → `description: |\n  yaml\n  foo: 1`, while
`bd show sra.6 --json | jq -r '.[0].description'` carries
`` ```yaml `` (2 fences). `bd show sra.4` → `"m~~n"`; `--json` → `"m~n"`.

**Proposed change.** In `swarm-dispatch` SKILL.md, replace "Read it in full —
`bd show <id>`" with `bd show <id> --json | jq -r '.[0] | .description,
.design, .acceptance_criteria, .notes'` (comments via `--include-comments`),
and say why in one sentence: the text form renders Markdown and corrupts code.

### C — AGENTS.md: document how a human reconciles a diverged `beads/data`
`audience: project` · `target: AGENTS.md (environmental-failure ledger /
"How beads persist")` · `priority: 1` · `findings: []` (manifest ledger entry
3) · `recurrence: 1` (the corpus listing run measured the same
non-fast-forward state) · destination: **bead**, `discovered-from` the run epic

**What is wrong, and what it costs.** The ledger says a non-fast-forward `bd
dolt push` "leaves the tracker local-only until a human reconciles it" and
forbids `bd dolt pull`, but gives the human no procedure. This run's ten
closures, all its findings, two new beads and its manifest are stranded in one
container; if it is reclaimed first, they are lost and the next session sees
ten beads open that were closed and committed.

**Evidence.** Ledger entry 3: "bd dolt push exits 1 non-fast-forward … Proved
by bd dolt push at 05:12Z: rejected, tip behind remote." Retro: `git ls-remote
origin beads/data` → `6c9e616`, ledger names `1e19020` — the remote kept moving.

**Proposed change.** Add to "How beads persist in THIS project" a measured
reconcile procedure for a diverged `beads/data` that does not use `bd dolt
pull` in-container (or states that none exists and the safe fallback, e.g.
replaying closures from `git log` commit subjects), with the command and its
output. Done when the procedure has been run once and its output is cited.

### D — `swarm dispatch open` should adopt an existing, harness-designated branch
`audience: skill` · `target: swarm` (CLI `dispatch open`) · `priority: 2` ·
`findings: []` (`amend:` on the run epic, 05:06:32Z) · `recurrence: 1`
(the structure-extraction run hand-merged onto a designated branch) · destination:
**`outbox/swarm/swarm/2026-10-07-dispatch-open-cannot-adopt-a-designated-branch.md`**

**What is wrong, and what it costs.** In a cloud session the harness permits
pushing only one designated branch; `swarm dispatch open` refuses to adopt an
existing branch, so the coordinator hand-writes `branch`/`base_sha` onto the
manifest, bypassing whatever `open` validates.

**Evidence.** "sprint branch is the session's designated branch
claude/project-thread-es7x8o (recorded directly, base_sha b626128…), because
the harness permits pushing only that branch and swarm dispatch open refuses to
adopt an existing branch." Not reproduced by the retro.

**Proposed change.** `swarm dispatch open --branch <existing>` adopts an
existing branch, recording its current tip as `base_sha`. Done when that
invocation exits 0 and the manifest's `branch`/`base_sha` match.

### E — A run resolves no `trailers`, so `--trailers-from` is never usable
`audience: skill` · `target: swarm` (scope step / `bb-close-bead.sh`) ·
`priority: 2` · `findings: []` (manifest) · `recurrence: 2`
(anchor references run candidate B, corpus listing run) · destination:
**`outbox/swarm/swarm/2026-10-07-a-run-resolves-no-trailers.md`** — the
previous copy lived in `.sf/outbox/swarm/`, which commit `732ea44` removed, so
there is no file here to bump.

**What is wrong, and what it costs.** Third run running: the manifest never
carries `trailers`, closes run without `--trailers-from`, and no commit carries
the session's required attribution trailers.

**Evidence.** `swarm manifest get --run <run epic> --key trailers` → `ok: false,
value: null`; `git log --format='%h|%(trailers:only,unfold)|' b626128..HEAD` →
ten empty blocks.

**Proposed change.** As filed for `59ic`: the scope step resolves `trailers`
onto the manifest, and the close step always passes `--trailers-from`. Done
when a run's commits carry them.

### F — swarm-dispatch: a dispatch returns only when its queue is drained
`audience: skill` · `target: swarm-dispatch` · `priority: 2` · `findings: []`
(coordinator report) · `recurrence: 0` · destination:
**`outbox/swarm/swarm-dispatch/2026-10-07-dispatch-returned-with-queue-undrained.md`**

**What is wrong, and what it costs.** `dev-sra-1` returned after one of its
three beads; the coordinator had to notice and resume it, costing a round trip
on the critical path of the gate beads.

**Evidence.** Coordinator-reported; no durable transcript (see could-not-verify).
Commits: `e602237` (`f6tl.5`) 05:10, then `07fa867` / `4638d28` after resume.

**Proposed change.** `swarm-dispatch`'s hand-back step requires `bd list
--assignee <dispatch-id> --status open,in_progress` to be empty (or each
remaining bead blocked/skipped with its reason) before the agent returns.

### Deduped — no new issue

- **`bd-health.sh` false failure** → two existing beads, both
  open. Proposed action: a recurrence comment on the first quoting the bd
  1.3.1 measurement above.
- **`f6tl.5#2`/`.6#1` hash reversal** — not a candidate: the code now
  folds tags, and `f6tl.7` retires the blueprint. Stays here for
  recurrence.
- **`f6tl.5#3` (`retro: n/a`)** — stays as the unrouted count.

## Decisions and filings

Put to Brett in the project thread at 05:55Z; answered "yes" at 06:00Z to the
recommendation to file all six. Filed 6, declined 0.

| | filed as |
|---|---|
| A | bead, `discovered-from:f6tl.6` |
| B | `outbox/swarm/swarm-dispatch/2026-10-07-read-beads-as-json-not-rendered-text.md` |
| C | bead, `discovered-from` the run epic — re-scoped: see below |
| D | `outbox/swarm/swarm/2026-10-07-dispatch-open-cannot-adopt-a-designated-branch.md` |
| E | `outbox/swarm/swarm/2026-10-07-a-run-resolves-no-trailers.md` |
| F | `outbox/swarm/swarm-dispatch/2026-10-07-dispatch-returned-with-queue-undrained.md` |

Dedupe: recurrence comment added to the existing `bd-health.sh` bead.

**The tracker is no longer local-only.** With Brett's approval, after a tar
backup of `.beads/embeddeddolt`: `timeout -s KILL 240 bd dolt pull` → exit 0,
"Pull complete."; a second pull then `bd dolt push` → "Push complete.";
`git ls-remote origin beads/data` → `9e1382b`. The pull did not deadlock, so C
was filed as "document the reconcile procedure and re-measure the `bd dolt
pull` deadlock ledger entry" rather than "no in-container remedy exists".
