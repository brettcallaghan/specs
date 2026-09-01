# Manifest 

## Table of Contents

- [RunStage](#runstage)
- [RunEffort](#runeffort)
- [AuthSource](#authsource)
- [TextAuthRow](#textauthrow)
- [FlagAuthRow](#flagauthrow)
- [CountAuthRow](#countauthrow)
- [StandingAuth](#standingauth)
- [TierModels](#tiermodels)
- [Exclusion](#exclusion)
- [Move](#move)
- [PreFlightKeys](#preflightkeys)
- [ScopeKeys](#scopekeys)
- [PlanKeys](#plankeys)
- [DispatchKeys](#dispatchkeys)
- [RetroKeys](#retrokeys)
- [RunManifest](#runmanifest)

## RunStage

The stages of a run, in the order a run reaches them.

A manifest key becomes required at the stage whose key struct declares it,
and the value here names that stage. The correspondence is structural:
the struct declaring a key IS the statement of its stage, so a key cannot
be given a type in one place and a stage in another.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| PreFlight | "pre-flight" |  |
| Scope | "scope" |  |
| Plan | "plan" |  |
| Dispatch | "dispatch" |  |
| Retro | "retro" |  |


## RunEffort

How much reasoning effort the harness was set to for this run.

A harness that exposes no such setting records it as unrecorded rather
than leaving the key absent, so a reader can tell "not offered" from
"not yet written".

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Low | "low" |  |
| Medium | "medium" |  |
| High | "high" |  |
| Xhigh | "xhigh" |  |
| Unrecorded | "unrecorded" |  |


## AuthSource

Where an authorization row's value came from.

A row resolved from the project's own configuration and a row that took
the skill's default are different facts, and the record exists to stop a
decision being re-litigated mid-run.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Project | "sf.toml" |  |
| Default | "default" |  |


## TextAuthRow

An authorization row whose resolved value is a word.

| Property | Type | Documentation |
|----------|------|---------------|
| value | String |  |
| source | [AuthSource](#authsource) |  |


## FlagAuthRow

An authorization row whose resolved value is yes or no.

| Property | Type | Documentation |
|----------|------|---------------|
| value | Bool |  |
| source | [AuthSource](#authsource) |  |


## CountAuthRow

An authorization row whose resolved value is a count.

| Property | Type | Documentation |
|----------|------|---------------|
| value | Int |  |
| source | [AuthSource](#authsource) |  |


## StandingAuth

The authorization gate's rows, each with its resolved value and where
that value came from.

Named fields rather than a map, because the rows are a fixed set and each
one's value has a fixed type: a map would have to carry a value of mixed
type, and there is no such thing in a closed schema. This is also what
lets the wire form stay exactly what the skills write today.

| Property | Type | Documentation |
|----------|------|---------------|
| pull_request | [TextAuthRow](#textauthrow) | draft, ready, or none — what the run opens at the end. |
| move_foreign_beads | [FlagAuthRow](#flagauthrow) | Whether a bead belonging to another epic is moved into this run. |
| admit_review_findings | [FlagAuthRow](#flagauthrow) | Whether review findings are worked in this run or only filed. |
| tier_escalation | [TextAuthRow](#textauthrow) | once or never — whether a repeated failure escalates a tier. |
| review_budget | [CountAuthRow](#countauthrow) | How many review loops the run may spend. |
| merge_to_default | [FlagAuthRow](#flagauthrow) | Whether the run may merge to the default branch. |


## TierModels

The model behind each capability tier, least to most capable.

| Property | Type | Documentation |
|----------|------|---------------|
| light | String |  |
| standard | String |  |
| advanced | String |  |
| frontier | String |  |


## Exclusion

A bead the run considered and did not admit.

| Property | Type | Documentation |
|----------|------|---------------|
| id | String |  |
| reason | String |  |


## Move

A bead the run moved into its own epic, and where it came from.

| Property | Type | Documentation |
|----------|------|---------------|
| id | String |  |
| from | String |  |


## PreFlightKeys

Keys the run's pre-flight establishes.

Every key here is optional on the wire: the manifest is a partial record
that grows as a run proceeds, and a key's absence before its stage is a
legitimate state rather than a defect. What this struct asserts is the
stage at which each key becomes required, which is the fact validation
reads.

| Property | Type | Documentation |
|----------|------|---------------|
| checks_ids | Array<String>? | The ids of the checks the project declares, in declaration order. |
| checks_baseline | String? | What the declared checks printed on the untouched tree, verbatim.<br><br>Recorded whole rather than summarized: the bar for every bead in the run is "no findings beyond this", which a summary cannot express. |
| checks_baseline_status | Int? | The status the baseline run exited with. Zero means every check is expected to pass, so any failure a bead sees is its own. |
| concurrency_cap | Int? | The most dispatches the run may have in flight at once. |
| models | [TierModels](#tiermodels)? | The model behind each capability tier. |
| effort | [RunEffort](#runeffort)? | The reasoning effort the harness was set to. |
| ledger | Array<String>? | Facts established this run that no dispatch should re-diagnose, each naming the command that proved it and the result observed.<br><br>Append-only: dispatch prompts cite entries by position, so an entry removed or reordered retargets every citation that addressed it. |


## ScopeKeys

Keys the run's scope and admission establish.

**Inherits from:** [PreFlightKeys](#preflightkeys)

| Property | Type | Documentation |
|----------|------|---------------|
| target | String? | The flat target epic whose task children are the run's work. |
| admitted | Array<String>? | Every bead admitted to the run. |
| excluded | Array<[Exclusion](#exclusion)>? | Every bead considered and not admitted, with why. |
| moved | Array<[Move](#move)>? | Every bead moved into the run's epic, and where it came from. |
| standing_auth | [StandingAuth](#standingauth)? | The authorization gate's resolved rows. |
| review_budget | Int? | How many review loops the run may spend. |


## PlanKeys

Keys the run's planning establishes.

**Inherits from:** [ScopeKeys](#scopekeys)

| Property | Type | Documentation |
|----------|------|---------------|
| collision_edges | String? | Where the edge document the run was built from was written. |
| partition_preview | String? | What became of the partition the planner proposed. |


## DispatchKeys

Keys the run establishes when it creates its branch, immediately
before its first dispatch.

**Inherits from:** [PlanKeys](#plankeys)

| Property | Type | Documentation |
|----------|------|---------------|
| branch | String? | The sprint branch. Written once: a run whose manifest already carries it reuses the branch and creates nothing, which is what makes a resumed run land where the original did. |
| base_sha | String? | The commit the sprint branch was cut from, recorded at creation and never derived later — a merge-base taken afterwards assumes the base has not moved, and on a long run it will have. |
| review_passes | Int? | How many review loops the run has spent. Durable because a resumed run that restarted the counter would loop forever. |


## RetroKeys

Keys the run's retrospective establishes.

An absent findings_path means the retro did not run, which is a different
result from a retro that found nothing.

**Inherits from:** [DispatchKeys](#dispatchkeys)

| Property | Type | Documentation |
|----------|------|---------------|
| findings_path | String? | Where the run's one human-facing document was written. |
| retro_filed | Int? | How many findings became issues. |
| retro_declined | Int? | How many findings a human declined. |
| retro_routed | Int? | How many findings were routed elsewhere. |


## RunManifest

The run manifest: the JSON object in a run epic's tracker metadata,
and the run's only durable record.

This is the whole declared key set. A key outside it is refused rather
than stored, and a value whose type is not its key's declared type is
refused, which is what the closed shape above expresses.

The stage at which a key becomes required is the first struct in the
chain that declares it — PreFlightKeys, then ScopeKeys, PlanKeys,
DispatchKeys, RetroKeys — so a key's name, its type and its stage all
come from one declaration.

**Inherits from:** [RetroKeys](#retrokeys)

| Property | Type | Documentation |
|----------|------|---------------|
| trailers | String? | The commit trailers the run resolved.<br><br>Declared here rather than in a stage struct, and that placement is what says it is never required: a run that resolved no trailer convention legitimately carries none at every stage, so an absent value means that rather than "not yet written". A key whose absence is always legitimate has no stage, and having no stage is expressed by having no stage struct. |


