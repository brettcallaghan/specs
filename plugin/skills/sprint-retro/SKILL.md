---
name: sprint-retro
description: Write a run's record and turn its friction into actions — harvest
  every finding the run recorded, report the results and what could not be
  verified, route each finding by kind, reason over them into candidate issues,
  put those to a human once, and file what they accept. Use when a sprint has
  finished, tripped its circuit breaker or otherwise reached its end and needs
  its one human-facing document, or when asked to run a retrospective over a
  set of closed beads, harvest retro comments, or decide which findings deserve
  an issue. Serves two audiences by the merge test — the project's own quality,
  and the skills it runs on — and files nothing without a human's yes.
version: "0.2"
user-invocable: false
---

# sprint-retro

**Write the run's record, and convert the friction it captured into actions
that make the next run better.** One document, not two: the results and the
findings are the same evidence read for two purposes, and a separate report
buys a second thing to keep in agreement with the first.

## Inputs

| | |
|---|---|
| `{EPIC}` | the run epic. Its manifest, its `amend:` comments, and its descendants are the whole of what you need |
| `{DISPATCH_ID}` | your agent id — `retro-<run-id>`, handed to you, never minted |

**Before your first `bd` call: `export BEADS_ACTOR={DISPATCH_ID}`.** You file
beads for what the human accepts, so you write; unset, those writes land under
the harness default and `bd ready --assignee retro-<run-id>` stops answering
for your phase. `<run-id>` is `{EPIC}` with the project prefix dropped. The
roster of agent ids and the derivation rule are the coordinator's dispatch
policy's, their only copy — take yours from there rather than inventing one.

**Everything else comes off the epic**, which is what lets a session that did
not hold the run write its record.

## What you return

- **The run's record** — one document, at the path the project configured.
- **N candidate issues**, in the issue template's shape.
- **The human's decision per candidate**, and what you filed for each yes.
- **Counts**: harvested, routed per branch, unrouted, filed, declined.

## Run it on every run

**Including one that closed nothing**, and one that stopped early. A sprint
that bounced its whole queue produced no code and a great deal of evidence
about grooming, readiness and spec quality — which is the project branch's best
material. A retro that only fires on success cannot see the failures.

**A run that ends without this document has failed**, whatever else it
achieved.

## Do not re-derive the run

The inputs are the harvested findings, the epic's manifest, its `amend:`
comments and `git log` on the sprint branch. **Do NOT re-read diffs, re-run
checks, or re-open the beads' work.** A retro that costs a second pass over the
run is a retro that gets skipped.

---

## 1 · Where the document goes

**The project configures it; the skill only defaults.** Read `sf.toml`:

```toml
[reports]
# Where a run's record is written. Default: docs/reports/
dir = "docs/reports"
# Where issues against the SKILLS a run used go. UNSET means file them as beads
# in this project's tracker instead — which is right where this repository is
# the one that vends those skills.
skill_issues_dir = ".sf/outbox"
# One index of all retro activity. Default: rfcs/retro-log.md
log = "rfcs/retro-log.md"
```

```sh
path="<[reports].dir>/<YYYY-MM-DD>-<epic-slug>.md"
bd update {EPIC} --set-metadata findings_path="$path"
```

**Recording the path is what makes the document findable from the epic id
alone** — the one thing a resumed session, a triage pass or a later retro has.

## 2 · Harvest

One query over the run's closure set:

```bash
bd export | jq -r 'select(<run scope>) | .comments[]?
  | select(.text | startswith("retro:")) | .text'
```

`bd export`, never `bd list` plus `bd show --json`: `bd list` defaults to open
issues and a drained run is almost entirely closed, while `bd show --json`
omits comments entirely even though its text output prints them.

**A zero count across a DRAINED run is a broken query, not a clean run.**
Stop, check the scope and the prefix spelling, confirm the beads really are
closed, and re-run before aggregating anything. A zero count across a run that
closed nothing may be genuine — proceed on what grooming and readiness
produced.

**Findings are numbered `<bead-id>#<ordinal>`**, by position within that
bead's own findings. Cite that form everywhere. The id is derived, never
authored — `references/shared/findings-capture-protocol.md` says why.

## 3 · Route — a lookup, and nothing more

```
kind == tooling          → the SKILL audience
kind == proof            → the document, beside what it supports. No issue.
kind == environment      → the project's environmental-failure ledger. No issue.
kind is one of the six   → the PROJECT audience
kind missing or unknown  → UNROUTED — a count and a list, never dropped
```

**Read no finding's content here.** That is the point of the kind being
recorded at capture: a routing step that requires reading is one that gets
skipped on a seventy-finding run.

**`proof` and `environment` terminate here.** There is nothing for a human to
decide about a negative control or a third-party parser's grammar. Measured at
7 of 73 on a real run, so this is not a rare path.

**Unrouted findings are the taxonomy's health metric.** Report the count every
run. A rising one means the kinds are drifting from what agents actually hit,
and a finding routed nowhere reached its audience only if a human happened to
read it — which is not a mechanism.

**When a finding's kind is present but its audience is unclear, apply the
merge test** — where would the fix land? In the project's own repository, or
in the repository that vends the skill? The words will not tell you apart;
that question always will.

## 4 · Write the record

Everything below comes from the epic, `git log` and the harvest. **State a null
explicitly** wherever a section is empty: a missing section reads like a clean
run rather than an unreported one.

| section | source |
|---|---|
| **The run** — epic, branch, `base_sha`→head, queue size, declared checks, model lineup, effort | `manifest` |
| **Results** — every admitted bead: status, agent, commit, one clause | `manifest.admitted` × `bd show`, `git log <base_sha>..HEAD` |
| **What the run decided** — findings of kind `decision`, `scope`, `gap` and `sequencing`, first, because each marks a place a document failed to answer a question a bead needed | the harvest |
| **Scope** — `excluded` with its reason per entry, `moved`, and anything admitted then dropped. A run that finished its admitted set while beads sat out is done over a NARROWED scope and must say so | `manifest` |
| **Every finding, verbatim, grouped by kind** | the harvest |
| **Routing** — the counts per branch, and the unrouted list | §3 |
| **Validation** — the checks baseline, the checks at head, and whether CI was observed | `manifest.checks_baseline`, the checks runner |
| **Pushes** — `git log <remote>/<branch>..HEAD` empty, or what is stranded | git |
| **What the run could NOT verify** — named, one per line | the harvest's `unverifiable`, plus anything above that had no source |
| **Candidates and what was filed** | §5, §6 |

**The complete finding set is the substrate recurrence runs on.** A finding
filtered out once can never be observed appearing twice, and recurrence is the
only signal separating a real defect from one bad afternoon. Filtering the
document down to what looks actionable breaks that accumulation silently.

**An unstated gap is indistinguishable from one that does not exist**, which is
why the could-not-verify section is not optional and not a footnote.

## 5 · Reason, then decide — one human, once

**Cluster** findings that are the same problem seen from different beads,
**dedupe** against what is already filed, **rank**. Each candidate is the
issue template's frontmatter plus its three sections —
`references/shared/issue-template.md` is the format, and this skill does not
restate it.

**The bar: it changes how the next run goes.** "One batch needed three cycles"
is a summary. "The `schema`-labelled beads were batched shared-surface but each
regenerated the same fixture, so the batch serialized anyway — batch that label
sequential" is a candidate. Only findings that clear that bar become
candidates; the rest stay in the document, where recurrence can still find them.

**Ask once**, with the candidate list on screen — per **candidate**, not per
finding, which is one or two orders of magnitude fewer questions.

**Declining is normal and is recorded.** A declined candidate stays in the
document; nothing is deleted. A later sprint hitting the same friction raises
it again with a higher recurrence count, which is how a decision that was wrong
corrects itself with nobody tracking it.

**If nobody answers, the run does not block.** The document is complete and the
candidates are decidable later from it.

## 6 · File — one format, two destinations

| audience | destination |
|---|---|
| **project** | a bead in this project's tracker |
| **skill** | a file under `[reports].skill_issues_dir`, for delivery to the repository that vends the skill — or, where that key is UNSET, a bead here, because this project vends them |

```bash
# project audience
bd create --title "<title>" --priority <n> \
  --description "<the issue body>. Findings: <ids>. Record: <path>" \
  --discovered-from <origin-bead>
```

**Never both, and never the other's.** A project that does not vend a skill
files nothing in the vending repository's tracker, and never accumulates issues
about skills it cannot change.

**A filed issue carries its evidence INLINE, never only a path.** A bead whose
only evidence is a document reference stops being workable the moment that
document moves — quote the finding, the command and its output into the bead
itself, and cite the path beside them rather than instead of them.

**Cite durable anchors.** A path plus a quoted string, a class, a function or a
component name survives the edits between filing and dispatch; a line number
does not. Where one helps, mark it as a hint.

**Append to a skill-issues directory; never prune it.** It belongs to the
project, and what happens to it — forwarded, swept, ignored — is its owner's
call.

### Mark what a filed issue rests on

The moment a candidate is filed:

```bash
bd comment <finding's bead> "retro:filed:<new-id> — <finding id>"
```

**Two things depend on it.** Recurrence counts *unfiled* findings only, so the
pool shrinks as issues get filed rather than growing without limit — and the
audit closes both ways: a finding leads to the issue it produced, an issue
leads back to its findings.

---

## Close the loop

Append one row for the run to `[reports].log`, so one index covers all retro
activity, and commit it with the record and anything filed.

**Record the outcome on the epic:**

```sh
bd update {EPIC} --set-metadata retro_filed=<n> \
  --set-metadata retro_declined=<n> --set-metadata retro_routed=<n>
```

**Every issue carries the run's model lineup and effort from the manifest into
its provenance.** The manifest records them precisely so this step can, and an
issue without them ships with an unfillable field.

**Read the document before you hand it back.** You own that it is true, not
that it is written.
