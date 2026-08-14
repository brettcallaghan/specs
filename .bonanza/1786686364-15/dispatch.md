# bead-bonanza run 1786686364-15

## Manifest

Run id: `1786686364-15`
Run dir: `/tmp/claude/bead-bonanza/1786686364-15`
Repo: `/home/user/agent-plugins`
Sprint branch: `claude/validate-corpus-spec-cli-0t0cxo`
Epic(s): `agent-plugins-ia5` (Spec Corpus Markdown Links), `agent-plugins-eaa`
(Shipped standards and guides live in specs/)

### Pre-flight stdout block (verbatim)

```
RUN_DIR='/tmp/claude/bead-bonanza/1786686364-15'
SCRIPTS='/home/user/agent-plugins/plugin/skills/bead-bonanza/scripts'
BB_GATES_BLOCK=present
BB_LINEUP_BLOCK=present
BB_SUBAGENT_MODEL_OVERRIDE='unset'
BB_PR_CI=yes
BB_TREE_CLEAN=yes
BB_PUSH_NEEDS_REFSPEC=no
BB_GIT_CONFIG_WRITABLE=yes
BB_BRANCH='claude/validate-corpus-spec-cli-0t0cxo'
BB_REMOTE='origin'
BB_UPSTREAM='origin/claude/validate-corpus-spec-cli-0t0cxo'
BB_DEFAULT_BRANCH='main'
BB_BD=ok
BB_BD_DB='/home/user/agent-plugins/.beads/embeddeddolt'
BB_JQ=ok
BB_DOLT=absent
BB_DOLT_TELEMETRY=na
BB_DOLT_WRITE=yes
BB_DENY_ENTRY=absent
BB_SETTINGS_LOCAL_ALLOW=absent
BB_WORKTREE_BASEREF='head'
BB_GATE_BASELINE=captured
BB_GATE_FORMAT=pass
BB_GATE_LINT=pass
BB_GATE_BUILD=pass
BB_GATE_TEST_FULL_SUITE=unparsed
```

### Gate profile

From `AGENTS.md` → `## bead-bonanza gates`. Placeholders resolve as:

- `<format>` — `cd cli && gofmt -l .` (must print nothing)
- `<lint>` — `scripts/check-skills.sh`; plus `cd cli && go vet ./...` when
  `cli/` changed
- `<build>` — `cd cli && go build ./...` (no build step for the docs corpus)
- `<test>` — the three docs checks run **independently, never chained with
  `&&`** (`check-skills.sh --strict` is known red and would short-circuit):

  ```bash
  scripts/check-skills.sh --strict;             s1=$?
  scripts/sync-shared-references.sh --check;    s2=$?
  scripts/sync-strata-docs.sh --check;          s3=$?
  ```

  plus `cd cli && go test ./...` when `cli/` changed.
- Custom gates — `scripts/generate-schemas.sh --check` after any
  `schemas/*.strata` change; `python3 scripts/validate-corpus.py` (needs
  `python3 -m pip install jsonschema pyyaml` once per container).

Targeted-tests-by-surface list is in the same AGENTS.md block; subagents read
it there.

### Gate baseline (measured this run, `--gate-baseline`)

| Gate | Baseline |
|---|---|
| Format (`cd cli && gofmt -l .`) | pass |
| Lint (`scripts/check-skills.sh`) | pass |
| Build (`cd cli && go build ./...`) | pass |
| Test (full suite) | `unparsed` — no single backticked command; run the three checks per the block above |

The bar for every bead is **no findings beyond this baseline**, not exit 0.
`scripts/check-skills.sh --strict` is KNOWN RED pre-existing (exits 1 on a
warning); measure its finding set, not its exit status, and capture `$?`
directly — never through a pipe.

### Model lineup

From `AGENTS.md` → `## model lineup`:

- light → `haiku`
- standard → `sonnet`
- advanced → `opus`
- frontier → `fable`

`CLAUDE_CODE_SUBAGENT_MODEL` is **unset**, so tier assignments are honoured and
tier escalations are meaningful.

### Effort setting

`unrecorded` — this harness exposes no effort setting to the orchestrator.

### Standing authorization

**Granted** by the user at run start: re-resolve the queue after each push and
dispatch the next wave until dry; implicit decisions defaulted and recorded in
each bead's `retro:` comment; explicit `decision-required` beads still stop.

### Decisions taken at the Phase 1.5 gate

1. **`agent-plugins-8ko`** was parked (`grooming_status: blocked-needs-info`,
   status `deferred`) on an owner either/or. User chose **(a) rename the
   document**: rename
   `rfcs/outbox/retro-spec/rfc-agent-plugins-findings-capture-harvest.md` to
   slug `rfc-self-findings-capture-harvest`, update every citation of the old
   slug, and **file a follow-up bead** for the surviving disagreement between
   `plugin/skills/retro-skill/references/skill-rfc-template.md` (which
   prescribes `rfc-<origin>-<description>`) and the profile pattern. Do NOT
   widen `cli/profile/profile.go`.
2. **Draft PR: yes.** Opened from the sprint branch so every wave push gets the
   full CI matrix (`cli.yml`: ubuntu/macOS/Windows Go suite, six-target
   cross-compile, busybox static-binary run).
3. **Auto-chain: yes** (see Standing authorization).

### Push policy

`BB_PUSH_NEEDS_REFSPEC=no` — upstream is configured, a bare `git push`
resolves. Verify landings with
`git rev-parse origin/claude/validate-corpus-spec-cli-0t0cxo` after fetch,
never `git status -sb`, and never let a pipe mask a push's exit status.

### Commit trailers

`$RUN_DIR/trailers.txt`, passed to every close as
`bb-close-bead.sh --trailers-file /tmp/claude/bead-bonanza/1786686364-15/trailers.txt`.

### Environmental ledger (carried into every dispatch prompt)

- `scripts/check-skills.sh --strict` is KNOWN RED, pre-existing, editorial
  scope owned elsewhere. Do not fix it mid-run. Bar = no findings beyond
  baseline.
- `bd`'s JSONL export lags the database; `bb-close-bead.sh` handles this. Never
  hand-roll a close.
- `scripts/generate-schemas.sh --check` fails against a `strata` older than
  `strata-lang/strata@8997153`. This session's hook could NOT verify the
  installed `strata` against upstream (fetch failed; stamp `ce18d66f`). If
  `--check` reports the committed artifacts stale with empty `properties` maps
  on inheriting types, **the binary is wrong, not the repo** — never commit the
  regenerated output.
- `api.github.com` and GitHub release assets return 403 through the egress
  proxy.
- `python3 scripts/validate-corpus.py` needs
  `python3 -m pip install jsonschema pyyaml` once per container.

## Wave table

Planned Phase 2.5 from the enumerated queue of 15 ids (14 dispatchable + 1
epic). **This table IS the queue** — no filter is re-resolvable into it.

Grooming wave preview: **overridden** — one row (`agent-plugins-eaa.3` moved
from wave 3 into the wave-5 shared batch; see Overrides below).

`Push` legend: `defer` = close, do not push; `push` = last bead of its wave or
batch, push here and re-resolve the queue.

| Wave | Bead | Class | Bucket | Tier | Push | Reason |
|------|------|-------|--------|------|------|--------|
| 1 | agent-plugins-ia5.1 | shared | concurrent-wave #1 | standard | defer | same c4_changes: module-document-core as ia5.2/ia5.3 — held apart by the blocks chain, so no wave peer here |
| 1 | agent-plugins-wu1 | shared | concurrent-wave #1 | standard | defer | co-edits plugin/skills/spec/references/c4-guide.md with eaa.1/eaa.3; none of them are in this wave |
| 1 | agent-plugins-8ko | isolated | concurrent-wave #1 | standard | defer | rfcs/outbox/ + rfcs/retired/rfc-0016 only; disjoint from every other queued surface |
| 1 | agent-plugins-cb9 | isolated | concurrent-wave #1 | light | push | scripts/check-template-drift.py only; no other bead touches that script |
| 2 | agent-plugins-ia5.2 | shared | concurrent-wave #2 | standard | defer | same c4_changes: module-document-core as ia5.1/ia5.3; both are in other waves |
| 2 | agent-plugins-ia5.4 | shared | concurrent-wave #2 | standard | defer | same c4_changes: container-spec-cli as ia5.5, which is in wave 4 |
| 2 | agent-plugins-eaa.1 | shared | concurrent-wave #2 | standard | push | co-edits plugin/skills/*/references/** with wu1 (wave 1) and eaa.2/eaa.3 (wave 5) |
| 3 | agent-plugins-ia5.3 | shared | sequential | standard | push | same c4_changes: module-document-core as ia5.1/ia5.2; no peer available — ia5.2→ia5.3→ia5.5 is a dependency bottleneck |
| 4 | agent-plugins-ia5.5 | shared | sequential | standard | push | same c4_changes: container-spec-cli as ia5.4 (wave 2); no peer available — three beads wait on it |
| 5 | agent-plugins-eaa.2 | shared | shared-batch #1 | standard | defer | both edit scripts/sync-shared-references.sh and move files out of plugin/skills/*/references/; runs FIRST in the batch |
| 5 | agent-plugins-eaa.3 | shared | shared-batch #1 | standard | push | same script + same tree as eaa.2; its Design says "follow the standards move exactly", so it runs SECOND in the same subagent |
| 6 | agent-plugins-ia5.6 | isolated | concurrent-wave #3 | advanced | defer | c4_changes empty; consumes module-document-core + container-spec-cli, both settled by wave 4 |
| 6 | agent-plugins-47b | shared | concurrent-wave #3 | standard | push | co-edits .github/workflows/checks.yml with 92u (wave 7); measures plugin/skills/*/references/**, settled by wave 5 |
| 6 | agent-plugins-eaa | — | epic closure | — | push | NOT dispatched — orchestrator closes it once eaa.1 (w2), eaa.2 and eaa.3 (w5) are all closed |
| 7 | agent-plugins-92u | shared | sequential | standard | push | co-edits .github/workflows/checks.yml with 47b (wave 6); also needs the eaa epic closed |

### Dependency edges verified against the tracker

All ten edges the orchestrator supplied were re-read from `bd show --json` and
match: `ia5.2←ia5.1`; `ia5.3←ia5.2`; `ia5.4←ia5.1`; `ia5.5←ia5.3+ia5.4`;
`ia5.6←ia5.5`; `eaa.2←eaa.1+ia5.5+wu1`; `eaa.3←eaa.1`; `47b←ia5.5`;
`92u←8ko+eaa`. No edge was added or fabricated during planning.

### C4 overlap, computed

Co-edit collisions (never share a concurrent wave):

- `ia5.1 + ia5.2 + ia5.3` — `module-document-core`. Waves 1 / 2 / 3.
- `ia5.4 + ia5.5` — `container-spec-cli`. Waves 2 / 4.

Both sets are already separated by the blocks chain, so the C4 constraint costs
no extra wave here.

Sequencing constraints (`c4_impacts` ∩ another bead's `c4_changes`):

- `ia5.6 after ia5.1, ia5.2, ia5.3` — consumes `module-document-core`.
- `ia5.6 after ia5.4, ia5.5` — consumes `container-spec-cli`.

Both hold: ia5.6 is in wave 6, all five producers are in waves 1–4.

**Signal disagreement, resolved.** The same computation emits `sequence ia5.1
after ia5.4` and `sequence ia5.1 after ia5.5` (ia5.1 carries `c4_impacts:
["container-spec-cli"]`, which ia5.4 and ia5.5 change). The wired dependency
runs the OTHER way — `ia5.4←ia5.1` is a real `blocks` edge, and ia5.5 sits
downstream of both. **The blocks edge wins**: ia5.1's impact tag records that
the CLI surfaces its two new finding variants, not that ia5.1 consumes a
contract ia5.4 is changing. Planned as ia5.1 (wave 1) → ia5.4 (wave 2) → ia5.5
(wave 4). No action needed; recorded so a re-plan does not "fix" it into a
cycle.

### Collision surface used per bucket (C4-untagged beads)

Nine of the fifteen queued beads carry `c4_changes: []`. An empty array is an
absence of signal, not proof of isolation, so collisions for these were judged
on the file surface each bead's Design names:

| Bucket | Surface used | Collides with |
|---|---|---|
| `wu1` | `plugin/skills/spec/references/c4-guide.md` frontmatter; `cli/profile/profile_test.go` (the `{"guide", "guide-c4", …}` slug/path row) | eaa.1, eaa.3 on `c4-guide.md` |
| `8ko` | `rfcs/outbox/retro-spec/`, `rfcs/retired/rfc-0016-close-ceremony-and-finding-harvest.md` | nothing |
| `cb9` | `scripts/check-template-drift.py` | nothing |
| `eaa.1` | `plugin/skills/*/references/**` (the 14 corpus-shaped documents) | wu1, eaa.2, eaa.3 |
| `eaa.2` | `plugin/skills/*/references/**` → `specs/standards/`; `scripts/sync-shared-references.sh` (the vendoring map) | eaa.1, eaa.3, wu1, 47b |
| `eaa.3` | `plugin/skills/*/references/**` → `specs/guides/`; `scripts/sync-shared-references.sh` (same map) | eaa.1, eaa.2, wu1, 47b |
| `47b` | new script under `scripts/`; `.github/workflows/checks.yml`; READS `plugin/skills/*/references/**` and records a file-count floor over it | 92u on `checks.yml`; eaa.2/eaa.3 on the tree it measures |
| `92u` | deletes `scripts/validate-corpus.py`; `.github/workflows/checks.yml`; adds a second `specs.toml` rooted at `rfcs/` | 47b on `checks.yml` |
| `ia5.6` | `specs/architecture/*.md` (verification of folds), `specs/features/markdown-links/markdown-links-blueprint.md` (retire), `cli/testdata/cases/` | nothing — disjoint from 47b, its wave peer |

`wu1` is in wave 1 and `eaa.1` in wave 2 specifically because they co-edit
`c4-guide.md`; the epic's own Design also wants wu1 first, since after the move
a slug becomes a compatibility surface.

### Overrides of the grooming wave preview

- **`agent-plugins-eaa.3`: wave 3 → wave 5, batched with `eaa.2`.** The preview
  placed it in wave 3 on the dependency graph alone (`eaa.3←eaa.1` is its only
  edge, satisfied in wave 2). Two things override that. (1) It co-edits
  `scripts/sync-shared-references.sh` with `eaa.2` — the same vendoring map,
  which is a merge collision the graph does not encode. (2) Its Design reads
  "Follow the standards move exactly — same GENERATED COPY mechanism, same
  inverted map", which requires `eaa.2` to have landed first; running it in
  wave 3 would invert the two. Batching them into one subagent that does eaa.2
  then eaa.3 fixes the ordering, removes the collision, and saves one full
  `<test>` run.

Every other row is the preview as given.

### Ledger addenda (Phase 2.5)

These extend the manifest's environmental ledger; the header itself was not
edited. Numbering is local to this section.

1. `.github/workflows/**` cannot be pushed with this container's credential —
   **unverified by this planning pass**. Asserted independently by
   `agent-plugins-47b` ("the credential in this container cannot push
   `.github/workflows/**`; that step must be applied through the GitHub App
   path, as `agent-plugins-9fx` also needs") and by `agent-plugins-92u`
   ("editing `.github/workflows/**` requires the GitHub App credential; the PAT
   in this container is rejected"). Would be settled by a probe push touching
   only that path. **Both wave-6 `47b` and wave-7 `92u` depend on it**, so
   expect the workflow half of each to be handed back rather than landed;
   neither bead's remaining work is blocked by it.
2. The `custom` gate `python3 scripts/validate-corpus.py` **stops existing
   after wave 7** — deleting that script is `agent-plugins-92u`'s primary
   acceptance criterion. Nothing follows wave 7 in this run, so no gate profile
   needs amending mid-run; recorded so a resumed or extended run does not read
   the missing script as a broken environment. Note `agent-plugins-8ko` (wave
   1) still carries `scripts/validate-corpus.py exits 0` as an acceptance
   criterion, which is correct — the script is present through wave 6.
3. `bd`-side graph health is clean for this queue — proved by
   `.claude/skills/groom-beads/scripts/bd-health.sh`, observed `RESULT: no hard
   problems`, exit 0 (no cycles, no orphans, no dangling dependency ids). Its
   `lint` advisories name only beads outside this run's queue.
4. `agent-plugins-eaa` is an **epic**: it is a closure step in wave 6, not a
   dispatch. Handing it to a subagent is a planning error. `agent-plugins-ia5`
   (the markdown-links epic) is deliberately NOT in the queue and must not be
   admitted by a re-resolve.

## Amendments

### 2026-08-14 — `agent-plugins-eaa.4` admitted mid-run; wave 6 split into 6a and 6b

**Admitted under the mid-run scope test as an acceptance-criteria
decomposition** of `agent-plugins-eaa`, whose AC-1 reads "The **14**
corpus-shaped documents live under `specs/`". `eaa.2` moved 8, `eaa.3` moved 5
— 13. The fourteenth, `plugin/skills/spec/references/c4-guide.md`, moved in
neither and has no entry in `scripts/sync-shared-references.sh`'s MAP.

Cause is a three-bead seam, and it is the run's clearest planning finding:
`wu1` (wave 1) retyped the document guide→standards; `eaa.2`'s Design had
already been written and declined it as "guides are separate"; `eaa.3`
confirmed it was standards-typed and therefore outside its 5 guides, recorded
the gap in a `retro:` comment, and **did not file a bead**. Each bead was
individually correct. The gap lives between them, which is exactly what an
AC-level check on the parent epic catches and a per-bead check does not.

**The epic cannot close until this lands** — AC-1 counts 13/14 and the
"vendored from `specs/` as the single source" criterion is unmet for this
document. Filing it is not scope creep; closing `eaa` without it would be
recording a false completion.

Edges wired: `parent = agent-plugins-eaa`; `discovered-from
agent-plugins-eaa.3`; **`agent-plugins-47b` now depends on it**.

**Wave 6 is therefore split**, because `47b` builds a script that records a
glob file-count floor over `plugin/skills/*/references/**` — the exact tree
`eaa.4` mutates. Running them concurrently would compute the floor against a
tree in motion. That collision is invisible to the C4 check (both are
`c4_changes: []`) and to a pure co-edit analysis (`47b` *reads* the tree rather
than editing it), which is a second `plan-sprint` finding: **a bead that
measures a surface collides with a bead that changes it, without co-editing a
single file.**

- **Wave 6a** — `ia5.6` (advanced) + `eaa.4` (standard), concurrent. Disjoint
  surfaces: `specs/architecture/` + the markdown-links blueprint + `cli/testdata`
  against `specs/standards/` + `scripts/sync-shared-references.sh` +
  `cli/profile`.
- **Wave 6b** — `47b` (standard), then close the `eaa` epic.
- **Wave 7** — `92u`, unchanged.

### 2026-08-14 — wave 2 concurrency produced a TGT/MAP interface hazard; carry it into ia5.5

`ia5.2` (TGT) and `ia5.4` (MAP) ran concurrently, so neither could see the
other's types. `ia5.2` recorded the consequence in a `retro:` comment: with no
link-map type in existence yet, it defined its own
`mdlinks.LinkMap` (`map[string]string`) and leans on Go's **nil vs non-nil map**
distinction to separate "no map supplied" from "a map supplied that resolves
zero entries" — AC-SPECMD-006.3 requires the latter to still raise
`links/unmapped`. `ia5.4` independently built `cli/cmd/linkmap.go`.

Nothing is broken at the merge: the two live in different packages and both
gate green. The hazard is at the seam, and `ia5.5` (CMD, wave 4) is the bead
that wires them. **Its dispatch prompt must carry this**: MAP must hand
`mdlinks` a **non-nil** map even when the YAML decodes to zero entries, or the
two components disagree about whether a map was given and
`links/unmapped` silently stops firing.

This is the C4-collision rule working as designed — the two beads were
correctly judged non-colliding on *files* — and also its limit: a shared
*interface* that does not exist yet is invisible to a co-edit analysis. Worth a
`plan-sprint` finding.

`ia5.4` separately split the map-reading fault into two rule ids
(`linkmap/invalid`, `linkmap/missing-path`) mirroring the package's existing
`config/schema` vs `config/missing-property` split, where the Design named the
three faults together. `ia5.5` should expect two rule ids, not one.

### 2026-08-14 — CI does not run on pushes to this branch (corrects the Phase 0 record)

The Phase 0 "Decisions" entry says the draft PR means "every wave push gets the
full CI matrix". **That is wrong.** Both `.github/workflows/checks.yml` and
`cli.yml` declare `push: branches: [main]` only, and each job's `if` skips
draft pull requests. A push to this working branch triggers nothing, draft or
not. `workflow_dispatch` against the branch ref is the only validation path.

Consequence for the run: the slim-gate policy the shared-surface batch template
offers (skip the in-batch full suite when "a PR is open so pushes trigger full
CI") **must not be taken** in this run. Every bead and every wave runs the full
local gate suite, which is what has happened so far.

`cli.yml` and `checks.yml` were dispatched manually at 06:22Z against head
`2157379` (wave 1 complete) — the first CI this sprint's work has seen. Not
re-dispatched per wave: the matrix is the most expensive thing the repo runs
(macOS 10x, Windows 2x on a private repo). One more dispatch when the run
drains.

### 2026-08-14 — push mode inside concurrent waves

The wave table's `Push` column marks one bead per wave as the wave's push
point. In a **concurrent** wave that cannot be the subagent's job: each
subagent works on its own bead branch, not the sprint branch, so a subagent
push would push a bead branch nobody merges from, and the wave's last *return*
is not knowable in advance. Every concurrent-wave subagent is therefore
dispatched `defer_push`, and the **orchestrator** pushes the sprint branch once,
after the wave's branches are all merged and the worktrees removed.

The column is unchanged and still reads correctly as "the orchestrator pushes
after this wave"; this entry records that the mechanism is the orchestrator's,
not the marked bead's. Sequential waves (3, 4, 7) and the wave-5 shared batch
are unaffected — the single subagent there can push its own last commit.

### 2026-08-14 — draft PR already existed

Phase 0 asked whether to open a draft PR (`BB_PR_CI=yes`); the user said yes.
None was created: `brettcallaghan/agent-plugins#22` is already open as a draft
from this sprint branch into `main`. Every wave push gets the full CI matrix
with no action taken.

### 2026-08-14 — durable trigger cadence is hourly, not 15–30 min

`references/progress-streaming.md` prefers a recurring out-of-session trigger
at 15–30 minute intervals. This harness's recurring-trigger minimum is
**hourly**, so the armed trigger (`trig_01LHGWHPcTNMMVdDoFEh5E9D`, cron
`3 * * * *`) is coarser than the reference asks for. Recorded rather than
substituting a re-armed one-shot, because the reference's stated reason for
preferring a recurring trigger — the agent forgetting to re-arm is the failure
mode being defended against — outweighs the interval gap.
