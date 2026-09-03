# Plan: split the swarm CLI and the sprint skill family into `swarm`

**Target repository:** `git@github.com:brettcallaghan/swarm.git`
**Source repository:** this one (`brettcallaghan/agent-plugins`)

Transient — delete when the migration lands.

This is an execution plan. It is written for an agent with no memory of the
conversation that produced it: every file in this repository is classified,
every split is specified line-by-line where the split is not obvious, and every
claim about a tool was measured (the commands are quoted where they matter).

---

## 0. The end state

**In `swarm`:** the `swarm` Go CLI (source, tests, fixtures, e2e harness), its
Strata schemas, its specification corpus, the `swarm` Claude Code plugin
(`bead-bonanza` + six subskills + hooks + monitors), and the checks and CI that
gate all of it. It builds, tests and releases `swarm` with no reference back
here.

**In `agent-plugins`:** the `spec` CLI and the `sf` (spec-first) plugin — `spec`,
`groom-beads`, `groom-feature-blueprint`, `retro-spec`, `retro-skill` — plus the
repo-local `triage-issue` and `strata` skills, the specification corpus for all
of that, and the RFC archive.

**The dependency direction is one-way: `swarm` → `sf`.** Nothing in the
spec-first corpus or plugin references anything swarm-side. Measured: building
the reference graph from `spec index` and computing inbound edges to the
sixteen swarm corpus documents returns **zero** references from non-swarm
documents. The split has no cycle to break.

---

## 0.5 Preconditions — what must be true before an agent starts

Measured 2026-09-02. An executing agent SHOULD re-run the unchecked rows
rather than trusting this list.

| Precondition | State |
|---|---|
| `brettcallaghan/swarm` exists on GitHub | **done** |
| It is checked out locally beside this repo (`../swarm`) | **done** — 18 tracked files, `origin` = `git@github.com:brettcallaghan/swarm.git` |
| Its beads database is initialised and empty | **done** — `bd stats` → `Total Issues: 0`, `dolt_database: "swarm"`, its own `project_id` |
| Its tracker mints `swarm-` ids | **done, measured** — a probe `bd create` returned `swarm-45n` (since deleted). The prefix comes from the database and not from `config.yaml`, so this is the only check that settles it |
| A token with read/write on both repositories is on the environment | **done** — `AGENT_PLUGINS_ACCESS_TOKEN` |
| Every swarm branch that should be merged has been merged | **done** |
| No `bd remember` memories are outstanding | **done, by intent** — `bd memories` → `No memories stored`. Memories are not synced and would be lost silently |
| **`../swarm/.beads/config.yaml`'s `sync.remote` is usable** | **NOT DONE — see below** |
| The swarm marketplace is registered in Claude Code | **not done** — needed only when the plugin is installed, not to execute the migration |

**The one open precondition.** `../swarm/.beads/config.yaml` ends with:

```yaml
sync.remote: "git+ssh://git@github.com/brettcallaghan/swarm.git"
```

Two differences from this repository's working form, and the second is
load-bearing:

1. It is a **flat** `sync.remote:` key. Here the same setting is nested
   (`sync:` then an indented `remote:`).
2. It is **ssh**. Here it is `git+https://…`, and everything the
   `refs/dolt/data` write path depends on assumes https: the credential
   helper reads the PAT from the environment at call time, and the proxy
   bypass is configured for `https://github.com/`. An ssh remote uses none
   of it, and a cloud container generally has no ssh path to GitHub at all.

Change it to the nested https form before step 9, and verify with a real
probe push rather than by reading — a `bd dolt push` that exits 0 is the
only evidence that counts.

---

## 1. The two questions asked up front

### 1.1 Which spec-first skills does `bead-bonanza` depend on?

**Exactly one hard dependency: `groom-beads`.** `sprint-plan` invokes it as a
skill, in its own context, as step 1:

- `plugin/skills/sprint-plan/SKILL.md:100` — "### 1. Invoke `groom-beads` over
  the whole queue — in your own context"
- `plugin/skills/sprint-plan/SKILL.md:103` — "Invoke the `groom-beads` skill
  directly. Do not…"
- `plugin/skills/bead-bonanza/SKILL.md:45` and `:143` — the coordinator's map
  and the "one subagent, two jobs, one context" rule name the same call.

Everything else that looks like a dependency is not one:

| Apparent dependency | What it actually is |
|---|---|
| `spec` skill | Never invoked. Every `spec` hit in the sprint family is either the `spec` **CLI**, the word "spec" in prose, or a vendored `shared/` header naming its source file. |
| `retro-spec`, `groom-feature-blueprint` | Named once, illustratively, at `bead-bonanza/references/operating-principles.md:41` ("the same channel `retro-spec`, `groom-feature-blueprint` and the …"). No invocation. |
| `retro-skill` | Not invoked by `sprint-retro`. `sprint-retro` **writes files** into `[retros].skill_issues_dir` (`.sf/outbox/<skill>/`); `retro-skill` and `triage-issue` are the *vending* repo's consumers of that directory, running in `agent-plugins`, not in a swarm run. |
| `strata` | Build-time only, for `schemas/swarm/*.strata`. Not a runtime skill dependency. |

**One document is misfiled today and moves outright:**
`plugin/skills/groom-beads/references/epic-definition-of-ready.md`. An *epic* is
a tracker-and-run concept, not a work-order concept — spec-first grooming is
deliberately scoped to **work orders**, which map onto beads today and will
stand alone under a deterministic mapping shortly. The document was written for
`bead-bonanza`, two of its six checks invoke the swarm CLI directly
(`swarm manifest get --run <epic> --key concurrency_cap`, `swarm manifest
require`), and `swarm-cli-frd.md:308` already declares that the dispatch
planner **replaces its four mechanical checks**. It moves whole to
`sprint-plan`, and `groom-beads` sheds it (§5.2a).

**Three authored documents stay in `sf` and are vendored into `swarm`'s skills,
exactly as they are vendored today** — `specs/standards/findings-capture-protocol.md`,
`plugin/skills/spec/references/issue-template.md`, and
`specs/standards/work-order-definition-of-ready.md`. The declared plugin
dependency is what makes that honest: `sf` stays the single author, `swarm`
commits generated copies carrying the source header, and a sync script
re-derives them from an `sf` checkout (§5.2b).

**The corpus-wikilink half of the same problem is separate** and is resolved
differently (§5.2c). Vendoring a file into a skill's `references/` does not make
a `[[slug]]` inside a *specification document* resolve — and a vendored copy
cannot be a corpus member at all, because its `<!-- GENERATED COPY` header
precedes the frontmatter and `frontmatter/parse` requires the file to begin with
`---`.

### 1.2 Should the swarm plugin declare a dependency on the spec-first plugin?

**Yes.** Claude Code's plugin manifest supports it, and this was measured rather
than assumed:

```
$ cat ptest/.claude-plugin/plugin.json
{"name":"ptest","description":"d","version":"0.0.1","author":{"name":"x"},"dependencies":["sf@spec-driven-dev"]}
$ claude plugin validate ptest --strict
✔ Validation passed

# Control — an unrecognized key does warn, so the pass above is meaningful:
$ cat ptest/.claude-plugin/plugin.json
{"name":"ptest","description":"d","version":"0.0.1","author":{"name":"x"},"bogusField":["a"]}
$ claude plugin validate ptest --strict
⚠ bogusField: Unknown field 'bogusField'. Claude Code ignores it at load time.
✘ Validation failed (--strict treats warnings as errors)
```

`claude plugin --help` also lists `prune|autoremove — Remove auto-installed
dependencies that are no longer needed`, which only makes sense against a real
dependency mechanism.

So `swarm`'s `plugin/.claude-plugin/plugin.json` carries:

```json
"dependencies": ["sf@spec-driven-dev"]
```

**Why declare it rather than vendor `groom-beads` into the swarm plugin.**
`groom-beads` is a *registry-shaped* skill in the sense §"Citing another skill's
reference" in `AGENTS.md` uses: its `bead.schema.json`, its capability-tier
metadata keys and its `validate_beads.py` are the one place a bead's shape is
defined. A second copy gives a second place to extend it, and the two would
diverge on the first tier added. The whole-file test also fails — `sprint-plan`
needs `groom-beads` **executed**, not read.

**Consequence to accept:** installing `swarm` at the user level pulls `sf` to the
user level too. That is correct — `bead-bonanza` cannot complete a run without
`groom-beads`, wherever it runs.

**Belt and braces:** `plugin/skills/sprint-plan/SKILL.md` step 1 gains an
explicit precondition — if `groom-beads` is not available in the session, stop
and say so by name, rather than improvising grooming. A declared dependency that
the runtime silently fails to satisfy must not read as "no grooming was needed".

---

## 2. Layout of the target repository

Recommended, and the rest of this plan assumes it:

```
swarm/
├── go.mod                     module github.com/brettcallaghan/swarm
├── go.sum
├── cmd/swarm/                 main.go, main_e2e_test.go
├── internal/
│   ├── cli/                   <- cli/swarm/cli/
│   ├── contract/ exec/ manifest/ plan/ status/ swarmmodel/ tracker/ vcs/
│   ├── e2e/                   <- cli/swarm/e2e/
│   └── schema/                swarm-model.schema.json + embed.go (new home)
├── testdata/
│   ├── swarm-cases/
│   └── swarm-fixtures/
├── schemas/                   <- schemas/swarm/ (promoted to the root)
├── specs/                     the swarm corpus
├── specs.toml
├── sf.toml
├── plugin/                    the `swarm` Claude Code plugin
├── .claude-plugin/marketplace.json
├── .claude/  .beads/  .github/workflows/  scripts/  docs/
├── AGENTS.md  CLAUDE.md  README.md  .gitignore  .gitattributes
```

**Why the root module rather than mirroring `cli/`.** `go install
github.com/brettcallaghan/swarm/cmd/swarm@vX.Y.Z` yields a binary named `swarm`,
and the release tag becomes a plain `vX.Y.Z` instead of the `cli/vX.Y.Z` prefix
this repository needs because its module sits in a subdirectory. Mirroring `cli/`
would give either `…/swarm/cli/swarm` (redundant) or a binary named `cli`.

**Import rewrite** — one substitution over every migrated `.go` file:

```
github.com/brettcallaghan/agent-plugins/cli/swarm/  →  github.com/brettcallaghan/swarm/internal/
```

`cli/swarm/main.go` and `main_e2e_test.go` move to `cmd/swarm/`. The e2e harness's
relative fixture paths survive unchanged: `internal/e2e/harness_test.go` holds
`fixturesDir = "../../testdata/swarm-fixtures"`, which resolves to the repository
root from `internal/e2e/` exactly as it did from `cli/swarm/e2e/`.

**The one shared Go package.** `cli/swarm` imports exactly one non-swarm package
— `cli/core/schema` — from two test files:

- `cli/swarm/contract/contract_test.go:16`
- `cli/swarm/cli/envelope_test.go:18`

Both read `schema.SwarmModel`. Create `internal/schema/` in the target holding
`swarm-model.schema.json` and a generated `embed.go` exporting `SwarmModel`
only. **No other cross-boundary Go dependency exists** — the swarm tree's
complete third-party import set is `BurntSushi/toml`, `goccy/go-yaml`,
`santhosh-tekuri/jsonschema/v6`, `spf13/cobra`, `spf13/pflag`.

---

## 3. Audit — files that MIGRATE to `swarm` and are DELETED here

### 3.1 Go source (`cli/`) — 124 files + 47 testdata files

| Source path | Target path | Notes |
|---|---|---|
| `cli/swarm/main.go`, `cli/swarm/main_e2e_test.go` | `cmd/swarm/` | 2 files |
| `cli/swarm/cli/` | `internal/cli/` | 27 files |
| `cli/swarm/contract/` | `internal/contract/` | 2 |
| `cli/swarm/e2e/` | `internal/e2e/` | 11 |
| `cli/swarm/exec/` (incl. `exectest/`) | `internal/exec/` | 4 |
| `cli/swarm/manifest/` (incl. `beads/`) | `internal/manifest/` | 7 |
| `cli/swarm/plan/` | `internal/plan/` | 4 |
| `cli/swarm/status/` | `internal/status/` | 2 |
| `cli/swarm/swarmmodel/` | `internal/swarmmodel/` | 1 (generated) |
| `cli/swarm/tracker/` (incl. `beads/`) | `internal/tracker/` | 10 |
| `cli/swarm/vcs/` (incl. `git/`) | `internal/vcs/` | 5 |
| `cli/core/schema/swarm-model.schema.json` | `internal/schema/swarm-model.schema.json` | generated; **see §5.1 — `cli/core/schema/embed.go` is a SPLIT, not a move** |
| `cli/testdata/swarm-cases/` | `testdata/swarm-cases/` | 20 files, whole directory |
| `cli/testdata/swarm-fixtures/` | `testdata/swarm-fixtures/` | 27 files, whole directory (`repository.bundle` blobs included) |

**Delete here:** all of `cli/swarm/`, `cli/testdata/swarm-cases/`,
`cli/testdata/swarm-fixtures/`, and `cli/core/schema/swarm-model.schema.json`.

**Stays here untouched:** `cli/cmd/` (23), `cli/core/` minus the one schema file
(57), `cli/profile/` (5), `cli/spec/` (2), `cli/testdata/cases/` (299),
`cli/testdata/gen/` (2). `cli/testdata/gen` regenerates only `testdata/cases`
(`go run ./testdata/gen ./testdata/cases`) and has no swarm consumer — verified
by grep across `cli`, `scripts`, `.github`.

### 3.2 Strata schemas

| Source | Target |
|---|---|
| `schemas/swarm/manifest.strata` | `schemas/manifest.strata` |
| `schemas/swarm/output.strata` | `schemas/output.strata` |
| `schemas/swarm/strata.config.json` | `schemas/strata.config.json` |

`schemas/swarm/` carries its own `strata.config.json`, which already makes it an
independent Strata package root — it is carved out of the SpecModel package
today, so promoting it to the target's `schemas/` root is a directory move with
no schema edit. **Delete `schemas/swarm/` here.**

Stays here: `schemas/README.md`, `config.strata`, `feature.strata`,
`finding.strata`, `frontmatter.strata`, `index.strata`, `profile.strata`,
`vocabulary.strata`, `strata.config.json`, `overlays/specs-config.json`.
`schemas/README.md` needs the swarm paragraph removed (§4, modified-in-place).

### 3.3 Specification corpus (`specs/`) — 16 documents

Whole files, migrated and deleted here:

```
specs/adrs/ADR-019-skill-composed-subagent-orchestration.md
specs/adrs/ADR-022-project-declared-checks.md
specs/adrs/ADR-023-go-for-the-swarm-cli.md
specs/adrs/ADR-024-plans-are-values.md
specs/adrs/ADR-025-argv-not-shell.md
specs/adrs/ADR-026-no-rollback.md
specs/adrs/ADR-027-two-test-tiers-and-a-committed-fixture.md
specs/architecture/container-swarm-cli.md
specs/architecture/module-bead-bonanza.md
specs/contracts/contract-swarm-json.md
specs/features/swarm-cli/swarm-cli-frd.md
specs/features/swarm-core/swarm-core-blueprint.md
specs/features/swarm-core/swarm-core-frd.md
specs/features/swarm-dispatch/swarm-dispatch-blueprint.md
specs/features/swarm-dispatch/swarm-dispatch-frd.md
specs/standards/swarm-output-standards.md
```

**No inbound references from the documents that stay** — measured over the
index's `references` array: zero non-swarm documents cite any of these sixteen.
Deleting them leaves the remaining corpus clean.

**Outbound references that will dangle in the target**, with their citing
documents — every one must be resolved by §5.2c before `spec check --strict`
passes in `swarm`:

| Dangling target | Cited by |
|---|---|
| `adr-011` (targeted verification gates) | `module-bead-bonanza` |
| `adr-013` (work-order terminology) | `ADR-019`, `module-bead-bonanza` |
| `adr-015` (Go for the spec CLI) | `ADR-023` (×4) |
| `adr-016` (Strata for schemas) | `container-swarm-cli`, `contract-swarm-json`, `swarm-core-blueprint`, `swarm-dispatch-blueprint`, `swarm-output-standards` |
| `adr-017` (spec CLI as plugin prerequisite) | `ADR-019`, `ADR-022` |
| `container-spec-cli` | `swarm-core-blueprint` (shared `cli` module — the claim is now false, see §5.3) |
| `module-document-core`, `module-spec-profile` | `swarm-core-blueprint` (same) |
| `ac-traceability-standards` | `container-swarm-cli`, `swarm-core-blueprint`, `swarm-dispatch-blueprint` |
| `c4-guide` | `container-swarm-cli`, `swarm-core-blueprint`, `swarm-dispatch-blueprint` |
| `findings-capture-protocol` | `module-bead-bonanza`, `swarm-cli-frd` |
| `gherkin-guide` | `swarm-cli-frd`, `swarm-core-frd`, `swarm-dispatch-frd` |
| `skill-best-practices` | `module-bead-bonanza` |
| `testing-standards` | `ADR-027`, `container-swarm-cli`, `swarm-core-blueprint` |
| `work-order-definition-of-ready` | `module-bead-bonanza` |

### 3.4 Plugin skills — the sprint family

Whole directories, migrated and deleted here:

```
plugin/skills/bead-bonanza/      (24 files, incl. references/generated/ and scripts/)
plugin/skills/run-batch/         (2)
plugin/skills/sprint-plan/       (3)
plugin/skills/sprint-recovery/   (1)
plugin/skills/sprint-retro/      (3)
plugin/skills/sprint-review/     (1)
plugin/skills/sprint-scope/      (1)
```

Plus **one file out of a directory that stays** — the only such case in the
whole migration:

```
plugin/skills/groom-beads/references/epic-definition-of-ready.md
        ->  plugin/skills/sprint-plan/references/epic-definition-of-ready.md
```

authored there rather than generated, with `groom-beads/SKILL.md` edited here to
shed it (§4, §5.2a).

Inside the seven directories, note for §5.2b — the six `references/shared/*.md`
files are **generated copies** whose sources stay here:

| Generated copy (migrates) | Its source today |
|---|---|
| `bead-bonanza/references/shared/findings-capture-protocol.md` | `specs/standards/findings-capture-protocol.md` |
| `bead-bonanza/references/shared/issue-template.md` | `plugin/skills/spec/references/issue-template.md` |
| `run-batch/references/shared/findings-capture-protocol.md` | same as above |
| `sprint-retro/references/shared/findings-capture-protocol.md` | same as above |
| `sprint-retro/references/shared/issue-template.md` | same as above |
| `sprint-plan/references/shared/epic-definition-of-ready.md` | `plugin/skills/groom-beads/references/epic-definition-of-ready.md` — **deleted, not migrated**: the source moves and is authored at `sprint-plan/references/epic-definition-of-ready.md` (§5.2a) |

A **seventh** vendored copy is added in the target that does not exist here:
`bead-bonanza/references/shared/work-order-definition-of-ready.md`, so the
migrated `module-bead-bonanza.md` has a local file to cite (§5.2c).

And `bead-bonanza/references/generated/` (README.md, manifest.md, output.md) is
produced by `generate-schemas.sh` from `schemas/swarm/` — it follows the schemas
(§3.2, §5.5).

### 3.5 Plugin hooks and monitors — whole directories

```
plugin/hooks/bb-stop-ready-loop.sh
plugin/hooks/bb-subagent-stop-queue.sh
plugin/hooks/hooks.json
plugin/monitors/bb-events-monitor.sh
plugin/monitors/monitors.json
```

All five are `bead-bonanza`-only: `hooks.json` registers only the two `bb-*`
scripts; `monitors.json`'s single entry is gated `"when":
"on-skill-invoke:bead-bonanza"`. **Migrate all, delete all here.** After the
move, `plugin/hooks/` and `plugin/monitors/` do not exist in `agent-plugins` —
`plugin/.claude-plugin/plugin.json` needs no edit for that (both are convention
paths, absent = not registered), but confirm with `claude plugin validate plugin
--strict`.

### 3.6 Documentation

| File | Disposition |
|---|---|
| `docs/plans/bead-bonanza-renovation.md` | migrate |
| `docs/plans/beads-coordination-design.md` | migrate |
| `docs/plans/beads-coordination-rollout.md` | migrate |
| `docs/plans/retro-design.md` | migrate (designs `sprint-retro`, the finding taxonomy, `.sf/outbox`) |
| `docs/plans/sprint-skill-design.md` | migrate |
| `docs/plans/wave-3-run-scope.md` | migrate |
| `docs/research/bead-bonanza-design/README.md` | migrate |
| `docs/research/retro-taxonomy-validation/README.md` | migrate |
| `docs/research/sprint-findings/` (2 files) | migrate |
| `docs/research/swarm-cli/2026-08-25-capability-audit.md` | migrate |
| `docs/retros/` (5 files: README.md, log.md, and the three run records) | migrate — all three run records are swarm runs (`bd97-swarm-core`, `bguq-swarm-dispatch-defects`, `e6xt-swarm-dispatch`), `README.md` describes the `bead-bonanza`/`sprint-retro` output, and `log.md` is the run ledger. **See §5.6**: `sf.toml`'s `[retros] dir` points here, so `agent-plugins` needs a fresh empty `docs/retros/` after the move. |
| `docs/upstream-reports/` (2 files) | migrate — the one report is `bd swarm list` behaviour, hit by `bead-bonanza`; the README describes the channel, which follows the reports |
| `docs/plans/swarm-skill-migration.md` (this file) | delete on completion, per `AGENTS.md` ("Transient — delete when the change lands") |

**Stays here:** `docs/research/gherkin/` (2 files — EARS-vs-Gherkin and the
cucumber library findings, both behind ADR-021).
`docs/session-bootstrap.md` is a **split** (§5.7).

### 3.7 RFC archive (`rfcs/`) — the open swarm-family inbox only

Migrate (open proposals against skills that are moving):

```
rfcs/inbox/bead-bonanza/            (6 files)
rfcs/inbox/plan-sprint/rfc-self-collisions-the-c4-check-cannot-see.md
rfcs/inbox/recover-sprint/rfc-self-worktree-absence-is-not-completion.md
```

**Stay here, whole:** `rfcs/retired/` (13), `rfcs/sprints/` (5),
`rfcs/outbox/` (2), `rfcs/specs.toml`, `rfcs/inbox/.gitkeep`, and the remaining
21 `rfcs/inbox/` files (feature-blueprint, frd, groom-beads,
groom-feature-blueprint, retro-skill, spec, standards, strata).

`rfcs/` is the **retired** RFC pipeline's archive, superseded by `triage-issue`
and `.sf/outbox`. The retired and sprint records are this repository's own
history — including the entries about `bead-bonanza`, which were actioned here
— and moving them would break `rfcs/specs.toml`'s second corpus root for no
gain. Only the *open* inbox items follow the skills they propose changes to.

---

### 3.8 The tracker — 60 open beads move, and are deleted here

**Not a file migration.** The beads live in the Dolt database under
`.beads/`, which is gitignored; nothing in §3.1–§3.7 or §6 touches them, and
an agent that executes every other section leaves the swarm backlog behind
without producing an error. This section is the one whose omission is
silent.

#### What moves

**The 60 OPEN beads listed below. Closed and deferred beads stay here.**
Measured on the 576-issue database: 128 open, 409 closed, 39 deferred.

Three consequences of "open only", each deliberate:

- **The run history stays with the repository that produced it.** 222 closed
  beads under the swarm-family epics, carrying 550 comments, are the evidence
  base that `docs/retros/`'s three run records cite by id. They stay readable
  here; the migrating beads that cite them keep pointing at
  `agent-plugins-*` ids, which is now a cross-repository pointer and is
  correct.
- **Three finished run epics are closed here rather than migrated** —
  `agent-plugins-mpnf`, `agent-plugins-so39`, `agent-plugins-qozp`. Each is an
  open epic whose children are all closed: they are runs that ended without
  anybody closing the epic. Closing them here removes 81 boundary edges that
  migrating them would have broken, and loses nothing.
- **27 deferred swarm beads are OUT of the set as written.** Flagged rather
  than decided: a deferred bead is work postponed, not work finished, and
  leaving it here means it resurfaces one day in a repository that no longer
  holds the code it is about. **If the answer is "migrate them too", add
  `--status deferred` to the export filter and re-run the boundary
  measurement — the id list below changes.**

#### The set

Authored, not derived — the classification is a judgement per bead and a
grep cannot make it. **The reviewable copy is
`docs/plans/swarm-bead-migration-ids.txt`**, one id per line with its title
as a trailing `#` comment; it is the input to every command below. Review
and edit it before executing. The grid here is the same set, laid out to be
read.

```text
agent-plugins-y63n        agent-plugins-offx        agent-plugins-hg9t
agent-plugins-y63n.1      agent-plugins-offx.2      agent-plugins-hg9t.6
agent-plugins-y63n.2      agent-plugins-offx.3      agent-plugins-hg9t.7
agent-plugins-y63n.3      agent-plugins-offx.4      agent-plugins-hg9t.8
agent-plugins-y63n.4      agent-plugins-offx.6      agent-plugins-hg9t.9
agent-plugins-y63n.5      agent-plugins-offx.7      agent-plugins-hg9t.10
agent-plugins-y63n.6      agent-plugins-offx.8      agent-plugins-hg9t.12
agent-plugins-y63n.7      agent-plugins-offx.9
agent-plugins-y63n.8      agent-plugins-offx.10     agent-plugins-ia3y
agent-plugins-y63n.9      agent-plugins-offx.14     agent-plugins-ia3y.3
agent-plugins-y63n.10     agent-plugins-offx.15     agent-plugins-ia3y.4
agent-plugins-y63n.11     agent-plugins-offx.16     agent-plugins-ia3y.6
agent-plugins-y63n.12     agent-plugins-offx.17
agent-plugins-y63n.13     agent-plugins-offx.18     agent-plugins-gzcr.1
agent-plugins-y63n.14     agent-plugins-offx.20
agent-plugins-y63n.15     agent-plugins-offx.21
agent-plugins-y63n.16
agent-plugins-y63n.17     standalone, no swarm epic above them:
agent-plugins-y63n.18     agent-plugins-pg8s    agent-plugins-o76b
agent-plugins-y63n.23     agent-plugins-muto    agent-plugins-oyf
                          agent-plugins-h92f    agent-plugins-akz
                          agent-plugins-89xd    agent-plugins-93yx
                          agent-plugins-bhl     agent-plugins-b1gl
                          agent-plugins-bpon    agent-plugins-hhdf
```

**Four judgement calls recorded, so they are not re-litigated blind:**

- `agent-plugins-ia3y.1` ("retro-skill and retro-spec still author RFCs")
  **stays** — those two skills stay here, though its parent epic migrates.
- `agent-plugins-gzcr.1` ("rename bead-bonanza to run-swarm") **migrates**
  though its parent `gzcr` (the RFC-inbox epic) stays.
- The `bd` defects — `mcy`, `we50`, `2nb8`, `86s`, `9owj` — **stay**. They are
  tracker bugs that hit both repositories; this one keeps the ledger.
- The shared-script beads — `bru`, `lmrl`, `d29`, `6dbb` (`check-skills.sh`,
  pipefail audit) — **stay**. Both repositories carry copies; the authoring
  home is here.

#### Measured mechanics (bd 1.2.2, measured 2026-09-02)

Each was run, not inferred. Two contradict the tool's own documentation.

1. **`bd export` is lossless enough.** It emits `labels`, `metadata`,
   `comments`, `acceptance_criteria`, `design`, `notes`, `close_reason`,
   `closed_at`, `assignee` and `dependencies` (parent-child included),
   omitting a key when it is empty. Across all 576 records: `labels` 347,
   `acceptance_criteria` 335, `metadata` 296, `comments` 258. An early
   reading of one record suggested labels and metadata were dropped; that
   was the omit-when-empty behaviour, not loss.
2. **`bd import` preserves ids verbatim.** Measured: `swarm-dw9` and
   `swarm-dw9.1` imported into a database already minting `swarm-*` landed
   intact beside a native bead, hierarchy preserved. `bd import --help`
   confirms it "accepts every field `bd export` emits", upserts, and skips
   rows whose status is `tombstone`.
3. **`bd rename-prefix` does NOT rewrite text bodies.** Its `--help` says
   "This will update all issue IDs and all text references across all
   fields." Measured on a scratch database: `agent-plugins-dw9.1` →
   `swarm-dw9.1` with the parent link intact, and a description reading
   `blocked by agent-plugins-dw9 and see agent-plugins-dw9` came back
   **byte-identical**. Its only flags are `--dry-run` and `--repair`. So
   rewriting ids inside descriptions, comments and acceptance criteria is
   this plan's job, not the tool's.
4. **`bd delete` rewrites the beads that stay.** It removes dependency links
   in both directions, **replaces text references in directly connected
   issues with `[deleted:ID]`**, and deletes permanently. Without `-f` it
   prints a preview. It refuses by default when a dependent sits outside the
   deletion set; `--force` orphans them, `--cascade` deletes them too —
   **never use `--cascade` here**, it would take closed history with it.

#### The method: rewrite the JSONL, do not rename in place

Because import preserves ids and `rename-prefix` does not touch bodies, one
filtered-and-rewritten JSONL does both jobs in a single pass, and
`bd rename-prefix` is never needed.

**Rewrite only the ids in the migrating set.** A migrating bead that cites
`agent-plugins-92q.31` must keep citing it — that bead stays here, and
rewriting the citation would point it at a bead that does not exist in
`swarm`.

```sh
# 1. In agent-plugins: export, then filter to the reviewed id list.
bd export > /tmp/all.jsonl
python3 - <<'EOF'
import json, re
ids = {l.split()[0] for l in open('/tmp/swarm-ids.txt') if l.strip()}   # stripped, see below
rx = re.compile(r'\b(' + '|'.join(sorted(map(re.escape, ids), key=len, reverse=True)) + r')\b')
sub = lambda s: rx.sub(lambda m: 'swarm-' + m.group(1)[len('agent-plugins-'):], s)
out = open('/tmp/migrate.jsonl', 'w')
for line in open('/tmp/all.jsonl'):
    r = json.loads(line)
    if r.get('id') not in ids:
        continue
    out.write(sub(json.dumps(r)) + '\n')   # rewrites id, both ends of every
                                           # edge, and every body mention
EOF
wc -l /tmp/migrate.jsonl        # must equal the id count

# 2. In ../swarm: import, and prove the shape survived.
cd ../swarm && bd import /tmp/migrate.jsonl && bd stats
```

Rewriting the serialized JSON rather than the parsed fields is deliberate:
it reaches `description`, `design`, `notes`, `acceptance_criteria`, every
comment body and both ends of every dependency in one substitution, with no
field list to keep in agreement. The word-boundary anchor is what stops
`agent-plugins-y63n` inside `agent-plugins-y63n.14` from being rewritten
twice, and what leaves the repository name `agent-plugins` alone.

#### Before deleting here: re-parent what the deletion would orphan

Measured against the 60-bead set: **32 dependency edges cross the boundary**
(23 parent-child, 5 related, 2 blocks, 2 discovered-from), and **28 beads
that stay lose a link** (24 closed, 3 open, 1 deferred).

The damage is concentrated in one shape — **22 closed or deferred children
whose parent epic migrates**:

| epic that migrates | children left behind here |
|---|---|
| `agent-plugins-offx` | 7 |
| `agent-plugins-hg9t` | 6 |
| `agent-plugins-y63n` | 6 |
| `agent-plugins-ia3y` | 3 |

Left alone, `bd delete` strips their parent link and rewrites the epic's id
in their text to `[deleted:agent-plugins-offx]`. So **before the delete**,
create one archive epic here and re-parent the 22:

```sh
bd create --type epic -p 3 \
  --title "Swarm work moved to brettcallaghan/swarm" \
  --description "Closed children of the four swarm epics that migrated on <date>.
Their parents now live in the swarm repository under swarm-* ids."
# then, for each of the 22:
bd update <child-id> --parent <archive-epic-id>
```

#### Deleting here

`--from-file` reads one id per line and does not strip comments, so feed it
the bare column:

```sh
sed 's/#.*//' docs/plans/swarm-bead-migration-ids.txt | awk 'NF' > /tmp/swarm-ids.txt
wc -l < /tmp/swarm-ids.txt                             # must be 60

bd delete --from-file /tmp/swarm-ids.txt --dry-run     # preview, no -f
bd delete --from-file /tmp/swarm-ids.txt --force
```

Then close the three finished run epics rather than deleting them:

```sh
bd close agent-plugins-mpnf agent-plugins-so39 agent-plugins-qozp \
  --reason "Run complete; remaining swarm work moved to brettcallaghan/swarm"
```

#### Verification — the migration is not done until these agree

| check | in `swarm` | in `agent-plugins` |
|---|---|---|
| issue count | `bd stats` → 60 (or the reviewed count) | 576 − 60 |
| every id carries the new prefix | `bd list --status all` shows no `agent-plugins-` id | — |
| hierarchy survived | the four epics show their children under them | the archive epic shows 22 |
| comments survived | per-bead `comment_count` matches the source export | — |
| labels and tier metadata survived | spot-check a `tier:` bead | — |
| nothing dangles | `bd orphans` clean | `bd orphans` clean |
| it is durable | `bd dolt push` exits 0, and a fresh clone bootstraps to 60 | `bd dolt push` exits 0 |

The last row is the one that matters: until `bd dolt push` succeeds against
`refs/dolt/data` on the new remote, the migrated beads exist only on the
machine that imported them.

## 4. Audit — files MODIFIED IN PLACE here (not split, not migrated)

| File | Edit |
|---|---|
| `.claude-plugin/marketplace.json` | drop `bead-bonanza and its sprint-* subprocesses, run-batch` from the marketplace description; the `sf` plugin entry itself is unchanged |
| `plugin/.claude-plugin/plugin.json` | drop "orchestrating them as autonomous sprints" from `description`; bump `version` (0.23.0 → 0.24.0) |
| `README.md` | remove `/bead-bonanza` from the workflow mermaid and from "six skills"; point the implementation step at the `swarm` repository |
| `schemas/README.md` | remove the `schemas/swarm/` paragraph |
| `.claude/skills/bead-bonanza` … `sprint-scope` | delete the seven symlinks (`bead-bonanza`, `run-batch`, `sprint-plan`, `sprint-recovery`, `sprint-retro`, `sprint-review`, `sprint-scope`); keep `groom-beads`, `groom-feature-blueprint`, `retro-skill`, `retro-spec`, `spec` |
| `plugin/skills/groom-beads/SKILL.md`, `references/capability-tier.md`, `references/bead.schema.json`, `scripts/validate_beads.py` | the four `sf.toml [models]` citations stay — `[models]` lives in **both** repos' `sf.toml` (§5.6). No edit, listed so an executing agent does not "fix" them. |
| `plugin/skills/groom-beads/SKILL.md` | **delete** the section "An epic has its own Definition of Ready" (`:164-175`, including the "Flatness is preferred" paragraph) and checklist item `6a` (`:120`). Both belong to `sprint-plan` now. `check-skills.sh` rule 2 turns the dangling `references/epic-definition-of-ready.md` path into a hard error, so this is not optional. See §5.2a. |
| `AGENTS.md` | drop the `bead-bonanza`-only ledger entries and the `bd`-persistence section's sprint framing only where it is swarm-specific; keep everything about `bd` itself (this repo still uses beads) |
| `CLAUDE.md` | unchanged — the bootstrap section is about `bd` and `strata`, both of which stay |
| `.gitignore` | remove `.beads/plans/` (emitted by `sprint-plan`), `*.gate.lock*`, `.claude/worktrees/`, `.worktrees/` — all four are `bead-bonanza` run artifacts. Keep the rest. |
| `docs/retros/` | recreate empty with a `.gitkeep` after §3.6 (still `sf.toml`'s `[retros] dir`) |

---

## 5. Audit — files SPLIT into two

Seven splits. Each is specified so an agent can execute it without judgement
calls.

### 5.1 `cli/core/schema/embed.go` + the JSON bundles

**Here (`cli/core/schema/`)** — keeps `spec-model.schema.json`,
`specs-config.schema.json`, and an `embed.go` with the `SpecModel` and
`SpecsConfig` vars. **Delete the `SwarmModel` var and its `//go:embed
swarm-model.schema.json` line.**

**There (`internal/schema/`)** — a new package holding
`swarm-model.schema.json` and an `embed.go` with `SwarmModel` alone.

Both `embed.go` files are **generated** by their repository's
`generate-schemas.sh` heredoc (§5.5), not hand-edited. Update the two swarm test
imports to the new package.

**Also update:** `specs/contracts/contract-swarm-json.md` names
`cli/core/schema/` as the validating artifact's home; in the target that becomes
`internal/schema/`. Grep the migrated corpus for `cli/core/schema` and
`cli/swarm` before the first `spec check` there.

### 5.2 The cross-boundary authored documents

Two surfaces, resolved differently. Getting them mixed up is the single most
likely way to execute this section wrong: a **skill reference** is a file under
`plugin/skills/<skill>/references/` that an agent reads while running, and a
**corpus wikilink** is a `[[slug]]` inside a specification document that `spec
check` resolves against the index. Vendoring fixes only the first.

#### (a) `epic-definition-of-ready.md` — MOVES WHOLE to `sprint-plan`

Not a split. `groom-beads` is scoped to **work orders**; an *epic* is a
tracker-and-run construct that `sprint-plan` gates, and the document was written
for `bead-bonanza` in the first place. Two of its six checks already shell out
to the swarm CLI, and `swarm-cli-frd.md:308` lists "the four mechanical checks
of `groom-beads/references/epic-definition-of-ready.md`" among what the dispatch
planner **replaces** — so leaving it in `sf` would strand a document whose
successor lives in the other repository.

- **Move** `plugin/skills/groom-beads/references/epic-definition-of-ready.md` to
  `swarm`'s `plugin/skills/sprint-plan/references/epic-definition-of-ready.md`,
  **authored there**, whole and unedited.
- **Delete** the generated copy
  `plugin/skills/sprint-plan/references/shared/epic-definition-of-ready.md` and
  the MAP row `sprint-plan<TAB>groom-beads/references/epic-definition-of-ready.md`
  — in the target it is authored, not vendored. Update
  `sprint-plan/SKILL.md:331` from `references/shared/epic-definition-of-ready.md`
  to `references/epic-definition-of-ready.md`.
- **Update** the migrated `swarm-cli-frd.md:308` path to the new home.

**Three edits `groom-beads` needs, and they are mandatory, not cosmetic**
(`check-skills.sh` rule 2 errors on a backticked `*.md` path that no longer
resolves from the citing file's directory):

1. `plugin/skills/groom-beads/SKILL.md:164-175` — delete the whole section
   "### An epic has its own Definition of Ready", including the "Flatness is
   preferred, not required" paragraph. Both are now `sprint-plan`'s.
2. `plugin/skills/groom-beads/SKILL.md:120` — delete checklist item
   `6a. WHEN grooming under an epic: the epic itself meets the epic-level DoR`.
3. `scripts/sync-shared-references.sh` — drop the `sprint-plan` MAP row here
   (already listed in §5.5). The orphan guard will then flag the stale generated
   copy under `sprint-plan/`, which is the intended signal, not a failure.

Trap for whoever executes the move: the file carries measured `bd` behaviour
that must survive verbatim — read `assignee` from `--json`, **never** from `bd
show`'s `Owner:` line (measured on bd 1.2.2: it prints `created_by`, so a check
written against it passes on an epic where not one bead is assigned), and
`--status all` is not optional because `bd list` defaults to open issues.

#### (b) Three documents stay in `sf` and are VENDORED into `swarm`

`sf` remains the single author of all three. `swarm` commits generated copies
under its skills' `references/shared/`, carrying the `<!-- GENERATED COPY` header
naming the source — exactly the arrangement in force today, now spanning two
repositories. The declared plugin dependency (§1.2) is what makes that
defensible: `swarm` runs `sf`'s skills, so it is entitled to ship `sf`'s applied
documents, and `AGENTS.md`'s vendor test passes on all three (whole file, and
applied rather than extended).

| Source in `sf` (unchanged, stays) | Vendored into, in `swarm` |
|---|---|
| `specs/standards/findings-capture-protocol.md` | `bead-bonanza`, `run-batch`, `sprint-retro` |
| `plugin/skills/spec/references/issue-template.md` | `bead-bonanza`, `sprint-retro` |
| `specs/standards/work-order-definition-of-ready.md` | `bead-bonanza` — **new row**, see (c) |

**How the copies get re-derived.** `swarm`'s `scripts/sync-shared-references.sh`
gains a second map, `SF_MAP`, whose sources resolve from an `agent-plugins`
checkout rather than from its own tree. Mirror `scripts/sync-strata-docs.sh`
exactly — it already solves this problem for a document set owned by another
repository:

- **First match wins:** `$SF_SRC_DIR`, `../agent-plugins`, `~/agent-plugins`,
  `~/.cache/sf-src` (a clone the session hook maintains).
- **Stamp each copy with its source commit** (`git log -1 -- <path>`), per file
  rather than repository HEAD, so an unrelated `sf` commit does not make these
  look stale.
- **No checkout present → exit 2 and SKIP**, never a silent pass and never a
  failure. `ci.sh` labels it SKIP the way it does the strata checks today.
- The committed copies are still gated unconditionally by
  `check-reference-links.py` and `check-skills.sh`, which need no checkout.

Do **not** source these from the installed `sf` plugin directory instead. It
ships `findings-capture-protocol.md` only as a *generated copy* (under
`groom-feature-blueprint/` and `retro-spec/`), so vendoring from there would
copy a copy and stamp it with a "source of truth" header pointing at a file that
is itself generated.

#### (c) Corpus wikilinks — CITE BY ROLE, or by the vendored path

The migrated specification documents carry `[[slug]]` citations that will not
resolve in `swarm`'s corpus. Vendoring does not help: a generated copy opens
with an HTML comment, and `frontmatter/parse` requires a corpus member to begin
with `---`, so a vendored file cannot be a corpus member at all.

Two forms, and which one applies is determined by whether the target is also
vendored as a skill reference under (b):

- **Vendored under (b) → cite the vendored path.** The migrated corpus already
  does this for non-corpus material: `module-bead-bonanza.md:713` cites
  `shared/c4-guide.md` by path, and `container-swarm-cli.md:1245` cites
  `.claude/skills/spec/references/shared/c4-guide.md`. Follow that pattern.
  - `module-bead-bonanza.md:46,218,642` — three `[[work-order-definition-of-ready]]`
    → `plugin/skills/bead-bonanza/references/shared/work-order-definition-of-ready.md`.
    This is why (b) gains the third MAP row: without it there is no local file to
    point at.
  - `module-bead-bonanza.md:644` and `swarm-cli-frd.md:954` —
    `[[findings-capture-protocol]]` →
    `plugin/skills/bead-bonanza/references/shared/findings-capture-protocol.md`.
- **Not vendored → cite by role, in prose, with no path.** This covers the
  remaining twelve dangling targets in §3.3's table (`adr-011`, `adr-013`,
  `adr-015`, `adr-017`, `ac-traceability-standards`, `c4-guide`,
  `gherkin-guide`, `skill-best-practices`, `testing-standards`,
  `container-spec-cli`, `module-document-core`, `module-spec-profile`). Say
  enough that a reader who never makes the hop still behaves correctly —
  `AGENTS.md`'s "cite by role" form. `adr-016` is the exception and gets a real
  ADR in `swarm`'s own corpus (§5.3).

**One forward-looking note, not an action for this migration.**
`swarm-cli-frd.md:614` specifies `swarm finding record --kind <kind>` with the
nine kinds validated at write time, and states that the protocol's "Where to
record" and "Finding ids are derived" sections "become a validated write rather
than a rule an agent applies". When FEAT-SWARM-RUN-006 is implemented, the nine
kinds become a `FindingKind` enum in `swarm`'s `schemas/` — grep confirms no such
enum exists in `schemas/swarm/*.strata` today. At that point `swarm` owns the
normative form and the citation in (c) becomes a pointer to `sf`'s prose
rationale rather than to its authority. Do not pre-empt it here.
### 5.3 `specs/features/swarm-core/swarm-core-blueprint.md` — one claim goes false

Lines 489–495 state the swarm CLI "shares the `cli` module, its `go.mod`" with
the spec CLI, and cite `[[container-spec-cli]]`, `[[module-document-core]]`,
`[[module-spec-profile]]`. **After the split this is no longer true.** The
migrated blueprint must say the swarm CLI is its own module in its own
repository, and drop the three architecture citations. `ADR-023`'s "one module
on purpose" rationale needs the same treatment — record the supersession in the
migrating commit rather than silently deleting it.

Same sweep for the `[[adr-015]]`, `[[adr-016]]`, `[[adr-017]]`, `[[adr-011]]`,
`[[adr-013]]` citations in §3.3's table: each becomes either (i) a new ADR in
`swarm`'s own corpus where the decision is genuinely swarm's, or (ii) prose
naming the decision without a wikilink. **Recommended:** author
`swarm/specs/adrs/ADR-001-strata-for-schema-and-model-generation.md` (the
`adr-016` role — five documents cite it, and `swarm` regenerates its own
schemas) and cite the remainder by role.

### 5.4 `sf.toml`

The declaration is genuinely two declarations sharing a file.

**Rows that go to `swarm` only** — the sprint-scoped runtime config:

- `[run]` — the whole block (all eight commented keys: `pull_request`,
  `move_foreign_beads`, `admit_review_findings`, `tier_escalation`,
  `review_budget`, `merge_to_default`, `dispatch_failure_threshold`,
  `run_root`). Read by `sprint-scope`'s authorization gate.
- `[retros]` — `dir`, and the commented `skill_issues_dir` guidance. In `swarm`,
  **`skill_issues_dir` must be SET** (`.sf/outbox`) for `sprint-plan`/`run-batch`
  findings about `groom-beads` and `spec`, which `swarm` does not vend — and
  left **unset** for findings about `bead-bonanza`, which it does. The comment
  block explaining "unset because this repository vends them" is now only half
  true in each repo; rewrite it in both.
- `[review]` — `skill = "code-review"`. Read by `sprint-review`.
- `[[checks]] id = "script-self-tests"` — its command names eight paths, **all
  eight of which migrate** (`bead-bonanza/scripts/*` ×5, `plugin/hooks/*` ×2,
  `plugin/monitors/*` ×1). Move the whole check; **delete it here**.

**Rows that go to `agent-plugins` only:**

- `[[checks]] id = "corpus"` and `id = "ref-links"` and `id = "template-drift"` —
  `template-drift` reads `cli/core/schema/spec-model.schema.json` and the `spec`
  skill's templates; neither exists in `swarm`. `corpus` and `ref-links` get
  swarm-repo equivalents (§5.5) but not the same command strings.

**Rows that appear, adapted, in BOTH:**

- `[models]` — the whole capability-tier lineup. `groom-beads` (in `sf`) *writes*
  `metadata.tier`; `bead-bonanza` (in `swarm`) *resolves* tier→model here. Both
  repos need the block, identical.
- `[[checks]]` `format`, `vet`, `build`, `test-go` — same intent, adjusted paths
  (`cd cli && …` → `go vet ./...` at the root in `swarm`).
- `[[checks]]` `skills`, `skills-strict`, `shared-refs`, `strata-docs`,
  `schemas` — same scripts, per-repo content (§5.5).

### 5.5 `scripts/`

| Script | Disposition |
|---|---|
| `scripts/generate-schemas.sh` | **SPLIT.** Here: keep steps 1–5 (SpecModel ast/json-schema/specs-config/go/markdown) and the DocumentType shape assertion; **delete** the `SWARM_SRC`/`SWARM_GO_OUT`/`SWARM_MD_OUT` block, step 6, step 6.5, step 6.6, the `SwarmModel` heredoc lines in the generated `embed.go`, and the two `sync_dir "$WORK/swarm-*"` calls. There: a new script running only the swarm half — ast, json-schema (+ the RunManifest inheritance assertion), go → `internal/swarmmodel`, markdown → `plugin/skills/bead-bonanza/references/generated`, and the `embed.go` heredoc emitting `SwarmModel` alone into `internal/schema`. **Do not lose the RunManifest assertion** — the manifest is a chain of five inheriting structs and an older `strata` build empties them while exiting 0. |
| `scripts/sync-shared-references.sh` | **SPLIT, and the target's gains a second map.** Here: delete the six swarm rows (`bead-bonanza` ×2, `run-batch`, `sprint-retro` ×2, `sprint-plan`); nothing else changes. There: **`MAP`** (local sources, resolved in-tree) is empty at first — every remaining row's source is an `sf` file — and **`SF_MAP`** carries the six rows resolved from an `agent-plugins` checkout per §5.2b: `bead-bonanza ← specs/standards/findings-capture-protocol.md`, `bead-bonanza ← plugin/skills/spec/references/issue-template.md`, `bead-bonanza ← specs/standards/work-order-definition-of-ready.md`, `run-batch ← …findings-capture-protocol.md`, `sprint-retro ← …findings-capture-protocol.md`, `sprint-retro ← …issue-template.md`. **No `epic-definition-of-ready` row** — it is authored in `sprint-plan` there (§5.2a). Keep the orphan guard (it is what catches a vendored file whose row was deleted), and keep the herestring form at `:227` rather than `printf \| grep -qxF` (measured: 2 intermittent failures in 30 `--check` runs under `pipefail`, herestring control 100/100 clean). |
| `scripts/check-skills.sh` | **SPLIT.** The 210-line body is generic (globs `plugin/skills/*/SKILL.md` and `.claude/skills/*/SKILL.md`, skips symlinks and `<!-- GENERATED COPY` headers). Copy verbatim to `swarm`, then **delete lines 186–192 there** — the rule reading `plugin/skills/groom-beads/references/bead.schema.json` against `scripts/validate_beads.py`, neither of which exists in `swarm`. Keep those lines here. |
| `scripts/ci.sh` | **SPLIT.** Here: drop `script-self-tests`. There: drop `rfc-queue`, `template-drift`; retarget `corpus` at the swarm corpus; keep `script-self-tests` with all eight paths. |
| `scripts/check-reference-links.py` | **BOTH, byte-identical.** Globs `plugin/skills/*/references/**`. No edit either side. |
| `scripts/sync-strata-docs.sh` | **BOTH, byte-identical.** `swarm` authors `.strata` files, so it needs the `strata` skill and its generated reference. |
| `scripts/format-check` | **BOTH.** Here unchanged. In `swarm`, drop the `/cli` from the `cd` (root module). |
| `scripts/install-spec-skill.sh` | **BOTH, adapted.** Generic over `plugin/skills/*/SKILL.md`; only `MARKER`, `MANIFEST` and the error message name the project. In `swarm` rename to `scripts/install-swarm-skills.sh` with `MARKER=".swarm-installed"`. **The marker rename is load-bearing** — the script deletes stale installs carrying its own marker, and two scripts sharing one marker would each uninstall the other's skills from `~/.claude/skills`. |
| `scripts/check-template-drift.py` | **HERE ONLY.** Reads `cli/core/schema/spec-model.schema.json` and the `spec` skill's templates. |
| `scripts/regenerate-cli-snapshots.sh` | **HERE ONLY.** Regenerates `cli/testdata/cases` via `./testdata/gen`. `swarm-cases` have no generator — `internal/e2e/cases_test.go` replays them against the compiled binary and their `expected/stdout` are hand-committed. |

### 5.6 `.beads/`, `.claude/`, `.codex/`, `.cursor/`, `.agents/`, `.sf/`

| Path | Disposition |
|---|---|
| `.beads/config.yaml` | **BOTH, adapted** — `sync.remote` becomes `git+https://github.com/brettcallaghan/swarm.git`. The JSONL-off comment and the 403/proxy trap apply verbatim in both. |
| `.beads/metadata.json` | **NEITHER — regenerate.** `project_id` is a per-database UUID; `swarm` gets its own from `bd init`/`bd bootstrap`. Copying it causes the measured `workspace identity mismatch detected` failure. |
| `.beads/.gitignore`, `.beads/README.md`, `.beads/hooks/*` (5) | **BOTH, byte-identical** — `bd`'s own scaffolding, project-agnostic. |
| `.claude/hooks/session-start.sh` | **SPLIT** (§5.7) |
| `.claude/hooks/dolt-push.sh` | **BOTH, byte-identical** (25 lines, `bd dolt push` on Stop) |
| `.claude/settings.json` | **BOTH, byte-identical** — `worktree.baseRef: head` (load-bearing for `bead-bonanza` worktrees), the SessionStart hook, `bd prime`, the Stop hook. |
| `.claude/skills/*` (symlinks) | here: delete the seven sprint-family symlinks. There: create seven symlinks to `../../plugin/skills/<name>` for the migrated skills. |
| `.claude/skills/strata/` (4 files) | **BOTH** — generated by `sync-strata-docs.sh`; `swarm` regenerates rather than copies. |
| `.claude/skills/triage-issue/SKILL.md` | **HERE ONLY** — the vending repo triages `.sf/outbox` deliveries; `swarm` produces them for `sf` skills and files beads for its own. |
| `.codex/config.toml`, `.codex/hooks.json`, `.cursor/hooks.json`, `.cursor/rules/beads.mdc`, `.agents/skills/beads/SKILL.md`, `.agents/skills/beads/agents/openai.yaml` | **BOTH, byte-identical** — `bd`'s cross-harness integration scaffolding, project-agnostic. |
| `.sf/outbox/README.md` | **BOTH, byte-identical** — `swarm` needs the outbox for findings about `sf` skills (§5.4). |

### 5.7 `.claude/hooks/session-start.sh` + `docs/session-bootstrap.md`

The hook has five numbered sections. Split by section:

| Section | `agent-plugins` | `swarm` |
|---|---|---|
| 1. install `bd` | keep | keep (identical) |
| 2. beads telemetry + database bootstrap, prefix guard, write-access probe | keep | keep, with the remote and database name changed |
| 3. vendored references (`sync-shared-references.sh`) | keep | keep |
| 4. **build the `swarm` CLI from `cli/swarm`** | **delete** | keep, retargeted at `./cmd/swarm`, and installed on PATH |
| 5. build `strata` | keep | keep (identical) |

Section 4 is the whole reason the hook is swarm-coupled here: `bead-bonanza`
invokes `swarm manifest` and `swarm status` as required steps and this hook is
the only thing that provisions the binary. **After the split, `agent-plugins`
sessions have no `swarm` on PATH** — which is correct, because no skill left
here invokes it.

`docs/session-bootstrap.md` splits the same way: the `AGENT_PLUGINS_ACCESS_TOKEN`
/ proxy-bypass / `refs/dolt/data` material is needed verbatim in both (with a
`SWARM_ACCESS_TOKEN` rename there); the `swarm`-build paragraph goes to `swarm`
only.

### 5.8 `.github/workflows/`

| Workflow | Split |
|---|---|
| `checks.yml` | **SPLIT.** Here: keep both jobs; from the `checks` job's "Bundled scripts' self-tests" step, **delete the whole step** (all five paths are `bead-bonanza/scripts/*`). There: `checks` job with `check-skills.sh --strict`, the self-test step (all **eight** paths — the five scripts plus the two `plugin/hooks/*` and `plugin/monitors/*`, matching `sf.toml`), `sync-shared-references.sh --check`, `check-reference-links.py`; **omit** `check-template-drift.py`. The `corpus` job is duplicated with the swarm corpus; **the RFC-queue step is `agent-plugins`-only.** |
| `cli.yml` | **SPLIT.** Here: keep the three-OS matrix, `cross-compile`, `no-runtime` — but reduce every `for pkg in ./spec ./swarm` to `./spec`, drop the "Install dolt (Windows)" and "Install bd, CGO-enabled (Windows)" steps and `SWARM_E2E_REQUIRE_TRACKER`, restore the default `go test ./...` timeout (the 30m raise exists for `cli/swarm/e2e` on Windows), and delete the "Run swarm in an image…" busybox step. There: the mirror — `./swarm` only (as `./cmd/swarm`), **keep** the dolt + CGO-enabled `bd` Windows install, **keep** `SWARM_E2E_REQUIRE_TRACKER`, **keep** `-timeout 30m`, keep the busybox exit-3 assertion. Drop the `spec check --json` busybox step (no corpus fixture there). |
| `release.yml` | **SPLIT.** Here: `for tool in spec` only; tag prefix `cli/v*` stays (subdirectory module). There: `for tool in swarm`, tag `v*` (root module), `go install github.com/brettcallaghan/swarm/cmd/swarm@<tag>` in the notes, and keep the "`swarm` additionally needs `bd` and `git` on PATH" sentence. |

### 5.9 `specs.toml` / root docs

`specs.toml`, `AGENTS.md`, `CLAUDE.md`, `README.md`, `.gitignore`,
`.gitattributes` are all **BOTH, adapted**:

- `specs.toml` — `[workspace] name = "swarm"`, `root = "specs"`, same
  `[profile] name = "spec-driven-dev"` (the swarm corpus is authored under the
  same profile; `swarm` consumes the `spec` CLI as a released binary, per
  ADR-017's "spec CLI as a plugin prerequisite").
- `.gitattributes` — byte-identical; the CRLF rationale applies to
  `testdata/swarm-cases/**/expected/` exactly as it does to `cases/`.
- `.gitignore` — split by the four sprint-artifact lines named in §4; `swarm`
  keeps those four and drops `specs/specs.json`/`rfcs/specs.json` down to one
  root.
- `AGENTS.md` / `CLAUDE.md` — the "Tool claims are measured at the point of
  use", "Pipes hide exit statuses", "Historical prose: don't", and
  environmental-ledger sections are the operating discipline both repos run
  under. Copy them. The `strata`/`generate-schemas` ledger entries belong in
  both; the `bead-bonanza`-specific entries go to `swarm` only.

---

## 6. Coverage check — every tracked file is accounted for

825 tracked files at `77c41b7`. The classification is exhaustive and the
buckets are disjoint — 211 + 560 + 23 + 9 + 10 + 11 + 1 = 825, verified by
script against `git ls-files`:

| Bucket | Paths | Count |
|---|---|---|
| Migrate + delete here | `cli/swarm/` 77, `cli/testdata/swarm-cases/` 20, `cli/testdata/swarm-fixtures/` 27, `cli/core/schema/swarm-model.schema.json` 1, `schemas/swarm/` 3, 16 specs, 7 skill dirs 35, `groom-beads/references/epic-definition-of-ready.md` 1, `plugin/hooks/` 3, `plugin/monitors/` 2, docs 18, rfcs 8 | **211** |
| Stay here, unmodified — including the three §5.2b sources, which `swarm` vendors without editing them here | `cli/cmd` 23, `cli/core` 56, `cli/profile` 5, `cli/spec` 2, `cli/go.mod`+`go.sum` 2, `cli/testdata/cases` 299, `cli/testdata/gen` 2, `schemas/` 9, `specs/` 54 (incl. `findings-capture-protocol.md`, `work-order-definition-of-ready.md`), `plugin/skills/` (spec 34 incl. `issue-template.md`, groom-beads 8, groom-feature-blueprint 6, retro-skill 4, retro-spec 5), `rfcs/` 42, `docs/research/gherkin` 2, `.claude/skills/` 6 (triage-issue + the five spec-first symlinks), `scripts/` 2 (`check-template-drift.py`, `regenerate-cli-snapshots.sh`) | **560** |
| Copied to both, byte-identical | `scripts/check-reference-links.py`, `scripts/sync-strata-docs.sh`, `.claude/hooks/dolt-push.sh`, `.claude/settings.json`, `.claude/skills/strata/` 4, `.beads/.gitignore`, `.beads/README.md`, `.beads/hooks/` 5, `.codex/` 2, `.cursor/` 2, `.agents/` 2, `.gitattributes`, `.sf/outbox/README.md` | **23** |
| Copied to both, adapted | `scripts/check-skills.sh`, `scripts/format-check`, `scripts/install-spec-skill.sh`, `.beads/config.yaml`, `AGENTS.md`, `CLAUDE.md`, `README.md`, `.gitignore`, `specs.toml` | **9** |
| Split into two files | `sf.toml`, `scripts/generate-schemas.sh`, `scripts/sync-shared-references.sh`, `scripts/ci.sh`, `.claude/hooks/session-start.sh`, `docs/session-bootstrap.md`, `.github/workflows/checks.yml`, `.github/workflows/cli.yml`, `.github/workflows/release.yml`, `cli/core/schema/embed.go` | **10** |
| Modified in place, here only | `.claude-plugin/marketplace.json`, `plugin/.claude-plugin/plugin.json`, `schemas/README.md`, `plugin/skills/groom-beads/SKILL.md` (§5.2a), and the 7 sprint-family `.claude/skills/*` symlinks (deleted) | **11** |
| Regenerate, do not copy | `.beads/metadata.json` | **1** |

**The tracker is not in this census and cannot be.** `.beads/` is gitignored,
so `git ls-files` sees none of the 576 issues. §3.8 is the only accounting of
them, and step 9 the only step that moves them.

Run this to confirm nothing was missed when the migration branch is ready:

```sh
git ls-files > /tmp/before.txt          # at the pre-migration commit
git ls-files > /tmp/after.txt           # on the migration branch
comm -23 /tmp/before.txt /tmp/after.txt # every deletion — must match §3
```

`cli/go.mod` and `cli/go.sum` stay **unmodified**: all seven direct requires
remain used by the spec side after `cli/swarm` leaves (`BurntSushi/toml`,
`cucumber/gherkin`, `cucumber/messages`, `goccy/go-yaml`,
`santhosh-tekuri/jsonschema`, `spf13/cobra`, `yuin/goldmark` — measured by grep
over `cmd core profile spec`). Run `go mod tidy` anyway and expect no diff; a
diff means something else moved.

---

## 7. Execution order

Do it in this order; each step leaves both trees buildable.

1. **Create the target repository** with `go.mod` (`module
   github.com/brettcallaghan/swarm`, `go 1.24`), `.gitattributes`,
   `.gitignore`, and the five third-party requires (§2).
2. **Move the Go tree** (§3.1) preserving history where possible
   (`git format-patch` over `cli/swarm` and the two testdata dirs, or
   `git filter-repo --path`). Apply the one-line import rewrite. Create
   `internal/schema/`. `go build ./... && go vet ./... && go test ./...` must
   pass in the target before anything else moves.
3. **Move the schemas** (§3.2) and write the target's `generate-schemas.sh`
   (§5.5). `scripts/generate-schemas.sh --check` green in both repos.
4. **Move the plugin** (§3.4, §3.5), write `plugin/.claude-plugin/plugin.json`
   with `"dependencies": ["sf@spec-driven-dev"]` and a
   `.claude-plugin/marketplace.json` naming the `swarm` plugin. Move
   `epic-definition-of-ready.md` and apply the three `groom-beads` edits here
   (§5.2a). Write the target's `sync-shared-references.sh` with its `SF_MAP`
   (§5.2b, §5.5) and run it against an `agent-plugins` checkout. `claude plugin
   validate plugin --strict` green in both repos.
5. **Move the corpus** (§3.3) and resolve every dangling reference (§3.3 table,
   §5.2c, §5.3). `spec check --strict` green in the target — this is the step
   that will take the longest, and it is the one that proves the corpus is
   self-contained.
6. **Move the docs and the open RFC inbox** (§3.6, §3.7).
7. **Split the config and CI** (§5.4, §5.6, §5.7, §5.8). `scripts/ci.sh` green in
   both.
8. **Delete here** (§3), apply the in-place edits (§4), run `scripts/ci.sh` and
   `.github/workflows/checks.yml`'s job set locally. Confirm
   `sync-shared-references.sh --check` has no orphaned generated copies left —
   the orphan guard is what catches a vendored file whose MAP row was deleted.
9. **Migrate the tracker** (§3.8). The target is already initialised with its
   own `project_id` and already mints `swarm-` ids, so there is no `bd init`
   step: fix `sync.remote` (§0.5), re-parent the 22 orphans, export–rewrite–
   import the 60, verify, `bd dolt push` on **both** sides, then delete here
   and close the three finished run epics. Do this **after** step 8, so a
   failure anywhere earlier leaves the backlog untouched and the whole
   migration re-runnable.
10. **Close the migration bead** (`agent-plugins-xgci`), and **delete this
    plan file** together with `docs/plans/swarm-bead-migration-ids.txt`.

---

## 8. Risks and traps, named

- **The corpus reference sweep (step 5) is the real work.** Fourteen distinct
  dangling targets across ten documents. `spec check --strict` will name every
  one, so work from its output rather than from this table — this table was
  built from the index at `77c41b7` and will age.
- **`swarm-core-blueprint.md` and `ADR-023` assert a shared Go module.** That
  becomes false at step 2 and must be corrected at step 5, not left to a later
  reader. Record what the split retired in the migrating commit message.
- **The `install-spec-skill.sh` marker collision** (§5.5) silently uninstalls
  skills. Rename the marker in the target before running it once.
- **`.beads/metadata.json` must not be copied** — a `project_id` mismatch
  produces `workspace identity mismatch detected`, and a `dolt_database` name
  mismatch produces the one tracker failure `bd` does *not* report (an empty
  database, `Total Issues: 0`, exit 0).
- **The Windows CI legs are swarm's, not spec's.** The dolt download, the
  CGO-enabled `bd` install with `-tags gms_pure_go`, `SWARM_E2E_REQUIRE_TRACKER`
  and `-timeout 30m` all exist for `cli/swarm/e2e`. Moving them and *also*
  leaving them here wastes the most expensive job this repository runs; leaving
  them only here breaks the target's e2e tier.
- **`groom-beads` breaks if the epic DoR moves without its three edits.**
  `check-skills.sh` rule 2 errors on the now-dangling
  `references/epic-definition-of-ready.md` path at `SKILL.md:168`, and checklist
  item `6a` would instruct an agent to apply a document it no longer ships. Both
  repos go red until §5.2a is done in full.
- **`SF_MAP` must fail loudly, never quietly.** A cross-repository vendor whose
  source checkout is absent has three possible behaviours and only one is right:
  exit 2 and report SKIP. Passing silently lets the committed copies rot
  invisibly; failing hard makes every `swarm` CI run depend on a second
  repository being present.
- **`docs/retros/` is `sf.toml`'s `[retros] dir` in both repos.** Recreate it
  empty here after the move, or the first `sprint-retro` run in `agent-plugins`
  writes into a directory that does not exist.
- **Both repos keep using `bd`.** Nothing in the beads scaffolding, the session
  hooks or the Dolt write-access requirement is swarm-specific; only the remote
  URL and the database identity differ.
- **The tracker migration fails silently or not at all.** Every other step in
  this plan fails loudly — a build breaks, a check goes red, `spec check`
  names a dangling slug. Step 9 has no such property: an agent that skips it
  finishes with two green trees and 60 beads stranded in the wrong
  repository. §3.8's verification table is the only thing that catches it.
- **`bd rename-prefix`'s help text is wrong about text references** (§3.8,
  measured). An agent that trusts it will ship 60 beads whose bodies cite ids
  that resolve in neither tracker. The JSONL rewrite is what avoids it, and it
  is why `rename-prefix` is not used at all.
- **`bd delete --cascade` would destroy closed history.** The flag
  recursively deletes dependents, and 22 closed children hang off the
  migrating epics. Use the bare `--force` after re-parenting, never
  `--cascade`.
- **`sync.remote` in the target is ssh, and the Dolt write path is https.**
  The credential helper and the proxy bypass that make `refs/dolt/data`
  writable both key off `https://github.com/`. Until §0.5's open precondition
  is closed, step 9 can import beads locally and then fail to push them
  anywhere.
