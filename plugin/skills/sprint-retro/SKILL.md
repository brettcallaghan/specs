---
name: sprint-retro
description: Turn a run's friction into actions — harvest every finding the
  run recorded, route each by kind, reason over them into candidate issues,
  put those to a human once, and file what they accept. Use when a sprint has
  finished and its findings need to become work, or when asked to run a
  retrospective over a set of closed beads, harvest retro comments, or decide
  which findings deserve an issue. Serves two audiences by the merge test —
  the project's own quality, and the skills it runs on — and files nothing
  without a human's yes.
version: "0.1"
user-invocable: false
---

# sprint-retro

**Convert the friction encountered during a run into actions that make the
next run better.** Not a report someone reads; not a queue someone drains.

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

## What you return

- **The findings**, every one of them, verbatim and grouped by kind — written
  into the run's findings document, which you create.
- **N candidate issues**, in the issue template's shape.
- **The human's decision per candidate**, and what you filed for each yes.
- **Counts**: harvested, routed per branch, unrouted, filed, declined.

## Run it on every run

**Including one that closed nothing.** A sprint that bounced its whole queue
produced no code and a great deal of evidence about grooming, readiness and
spec quality — which is the project branch's best material. A retro that only
fires on success cannot see the failures.

## Do not re-derive the run

The inputs are the harvested findings, the epic's manifest, its `amend:`
comments and `git log` on the sprint branch. **Do NOT re-read diffs,
re-run checks, or re-open the beads' work.** A retro that costs a second pass
over the run is a retro that gets skipped.

---

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

**Incidents arrive through this same query.** An incident is a severity, not a
kind — an ordinary `retro:<kind>:<target>` finding whose text carries the word
` incident ` after the prefix. It routes by its kind like any other; carry the
marker through to the candidate issue, because each one marks a place the
corpus failed to answer a question a bead needed answered.

**Findings are numbered `<bead-id>#<ordinal>`**, by position within that
bead's own findings. Cite that form everywhere. The id is derived, never
authored — `references/shared/findings-capture-protocol.md` says why.

## 3 · Route — a lookup, and nothing more

```
kind == tooling          → the SKILL audience
kind == proof            → the findings document, beside what it supports. No issue.
kind == environment      → the project's environmental-failure ledger. No issue.
kind is one of the six    → the PROJECT audience
kind missing or unknown  → UNROUTED — a count and a list, never dropped
```

**Read no finding's content here.** That is the point of the kind being
recorded at capture: a routing step that requires reading is one that gets
skipped on a seventy-finding run.

**`proof` and `environment` terminate here.** There is nothing for a human to
decide about a negative control or a third-party parser's grammar. Measured at
7 of 73 on a real run, so this is not a rare path.

**Unrouted findings are the taxonomy's health metric.** Report the count every
run. A rising one means the kinds are drifting from what agents actually hit.

**When a finding's kind is present but its audience is unclear, apply the
merge test** — where would the fix land? In the project's own repository, or
in the repository that vends the skill? The words will not tell you apart;
that question always will.

## 4 · Reason — where all the judgement is

Runs **separately per audience**, because the two produce different artifacts
for different people.

### (a) Every finding, verbatim, grouped by kind

**Into the run's findings document, which is yours to write and yours alone.**
You write no part of the report and the report writes no part of this:

```sh
path="rfcs/sprints/<YYYY-MM-DD>-<epic-slug>-findings.md"
bd update {EPIC} --set-metadata findings_path="$path"
```

Recording the path is what lets `sprint-report` cite the document without being
handed anything — and what makes the two skills order-independent, since
neither reads the other's output directly. Not a summary and not a selection.

**The complete set is the substrate recurrence runs on.** A finding filtered
out once can never be observed appearing twice, and recurrence is the only
signal separating a real defect from one bad afternoon. Filtering the document
down to what looks actionable breaks that accumulation silently.

### (b) N candidate issues

**Cluster** findings that are the same problem seen from different beads,
**dedupe** against what is already filed, **rank**. Each candidate is the
issue template's frontmatter plus its three sections —
`references/shared/issue-template.md` is the format, and this skill does not
restate it.

**The bar: it changes how the next run goes.** "One batch needed three cycles" is
a summary. "The `schema`-labelled beads were batched shared-surface but each
regenerated the same fixture, so the batch serialized anyway — batch that
label sequential" is a candidate. Only findings that clear that bar become
candidates; the rest stay in the document, where recurrence can still find them.

**Incidents are the best material.** Each marks a place the corpus failed to
answer a question a bead needed answered — a candidate against a *document*
rather than a complaint about the run.

## 5 · Decide — one human, once

**Ask once**, with the candidate list on screen — that is the evidence, and it
is yours, not the report's. Per **candidate**, not per finding — one or two
orders of magnitude fewer questions.

**Declining is normal and is recorded.** A declined candidate stays in the
findings document; nothing is deleted. A later sprint hitting the same
friction raises it again with a higher recurrence count, which is how a
decision that was wrong corrects itself with nobody tracking it.

**If nobody answers, the run does not block.** The findings document is
complete and the candidates are decidable later from it.

## 6 · File — one format, two destinations

| audience | destination |
|---|---|
| **project** | a bead in this project's tracker |
| **skill** | an issue file at `.sf/outbox/<skill>/<YYYY-MM-DD>-<slug>.md`, for delivery to the repository that vends it |

```bash
# project audience
bd create --title "<title>" --priority <n> \
  --description "<the issue body>. Findings: <ids>. Report: <path>" \
  --discovered-from <origin-bead>
```

**Never both, and never the other's.** A project files nothing in the vending
repository's tracker, and never accumulates issues about skills it cannot
change.

**Append to the outbox; never prune it.** It belongs to the project, and what
happens to it — forwarded, swept, ignored — is its owner's call. **Where this
repository vends the skills it runs it is its own downstream**, so
`triage-issue` sweeps the outbox into this tracker in the same run.

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

Append one row for the sprint to `rfcs/retro-log.md`, so one index covers all
retro activity, and commit it with the findings document and anything filed.

**Record the outcome on the epic**, so `sprint-report` can name what you
produced from the manifest alone:

```sh
bd update {EPIC} --set-metadata retro_filed=<n> \
  --set-metadata retro_declined=<n> --set-metadata retro_outbox=<n>
```

**Every issue carries the run's model lineup and effort from the manifest into
its provenance.** The manifest records them precisely so this step can, and an
issue without them ships with an unfillable field.
