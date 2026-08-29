# Output 

## Table of Contents

- [CommandMode](#commandmode)
- [CommandOutcome](#commandoutcome)
- [StepStatus](#stepstatus)
- [FactKind](#factkind)
- [FactValue](#factvalue)
- [Fact](#fact)
- [Invocation](#invocation)
- [CommandResult](#commandresult)

## CommandMode

Which of the three interpreters produced a result.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Explain | "explain" |  |
| DryRun | "dry-run" |  |
| Execute | "execute" |  |


## CommandOutcome

What a command concluded. The exit status is derived from this rather
than chosen beside it, so the two cannot disagree.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Clear | "clear" | It answered and its verdict is clear. Exit 0. |
| Attention | "attention" | It answered and what it found needs attention. Exit 1, and the resolution names what would resolve the condition. |
| Stop | "stop" | It found a condition the caller cannot repair in place. Exit 2. |
| Malformed | "malformed" | Its own invocation was malformed. Exit 3. |


## StepStatus

What became of one step of a command's plan.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Ran | "ran" | The interpreter attempted the invocation. |
| Refused | "refused" | A dry run declined to execute it, because it is not a read. |
| Skipped | "skipped" | It never ran: an earlier step stopped the plan, or an argument it needed came from a step that did not run. |


## FactKind

Which of four shapes a fact's value takes.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Text | "text" |  |
| Count | "count" |  |
| Flag | "flag" |  |
| List | "list" |  |


## FactValue

The value of an established fact.

A closed set of shapes rather than an open one: a consumer branches on
the kind and reads the field that kind carries, instead of guessing at
the JSON type it was handed.

Exactly one of the four value fields is present, and it is the one the
kind names. That is a rule the shape cannot state — a schema can say
which fields may appear, not which combination is coherent — so it is
enforced where the value is built, and a value carrying the wrong field
for its kind is refused rather than rendered.

| Property | Type | Documentation |
|----------|------|---------------|
| kind | [FactKind](#factkind) | Which field carries the value. |
| text | String? | Present when the kind is text. |
| count | Int? | Present when the kind is count. |
| flag | Bool? | Present when the kind is flag. |
| items | Array<String>? | Present when the kind is list. |


## Fact

One thing a command reports, with the invocations that established it.

An agent that can see the command behind a line can re-run it, argue with
it, and work around it when the underlying tool changes; an agent that
cannot has to trust a wrapper it did not write. So the invocations are
part of the fact rather than a separate section.

| Property | Type | Documentation |
|----------|------|---------------|
| name | String | What the fact is. A consumer keys on this, so it is a contract. |
| established | Bool | Whether the fact could be taken at all.<br><br>A fact that could not be taken is still reported — as unestablished, with a reason — rather than omitted or emitted as an empty value of its type. When this is false the value is absent and the reason says why; when it is true the value is present. |
| reason | String | Why an unestablished fact could not be taken. Empty when it was. |
| value | [FactValue](#factvalue)? | What was established. Absent when the fact was not. |
| invocations | Array<Array<String>> | The argv vectors a reader can re-run to reproduce the fact, each with its executable first. At least one. |


## Invocation

One step of the plan a command ran, as the audit line for it.

| Property | Type | Documentation |
|----------|------|---------------|
| step | String | The step's name within its plan. |
| argv | Array<String> | The resolved argv vector, executable first. Empty when the step did not run. |
| dir | String | The working directory it ran in. Empty means the caller's own. |
| status | [StepStatus](#stepstatus) | What became of the step. |
| reason | String | Why a refused or skipped step did not run. Empty otherwise. |
| exit_status | Int | The status the finished process reported. Meaningful only when the step ran; -1 when the process was terminated by a signal. |
| duration_ms | Int | How long the invocation took, in whole milliseconds.<br><br>A count rather than a formatted duration: a field whose meaning depends on the locale of the process that produced it is one this contract forbids. |
| stdout_digest | String | A digest of the whole standard output stream, as sha256:<hex>. |
| stderr_digest | String | A digest of the whole standard error stream, separately, because the streams are separate. |
| error | String | Why the step failed, when it did: the process could not be started, its decoder refused the result, or an argument could not be resolved. A non-zero exit alone is not an error. Empty when there was none. |


## CommandResult

The envelope every swarm command emits under --json.

Emitted on every exit status, including a non-zero one, so a consumer
parsing the result never has to fall back to reading text to learn why.

| Property | Type | Documentation |
|----------|------|---------------|
| contract_version | String | The version of the JSON contract this result was produced against, so a consumer branches on the version it was handed rather than on the binary's own version. |
| command | String | The invocation path, for example "swarm manifest get". |
| run | String | The run epic this invocation was scoped to. Empty for a command that is not run-scoped. |
| mode | [CommandMode](#commandmode) | Which interpreter ran. |
| outcome | [CommandOutcome](#commandoutcome) | What the command concluded. |
| status | Int in 0..3 | The exit status the outcome implies.<br><br>__Constraint:__ Values must be ≥ 0 and ≤ 3. |
| summary | String | One line saying what the command found. Empty when there is none. |
| resolution | String | What would resolve the condition. Present whenever the outcome is attention; empty otherwise. |
| facts | Array<[Fact](#fact)> | What the command established, in report order. |
| invocations | Array<[Invocation](#invocation)> | Every step the plan considered, in the order it considered them. |
| transcript | Array<String> | What the plan runner printed — the explained invocations in explain mode, and the executed ones otherwise. |


