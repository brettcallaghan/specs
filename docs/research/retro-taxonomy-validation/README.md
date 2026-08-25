# Validating the retro finding taxonomy against 73 real findings

**What this is.** The validation gate `docs/plans/retro-design.md` set for
itself, run. Every finding in the 2026-08-15 sprint report classified against
the seven proposed kinds, with the residue counted and named.

**Source:** `rfcs/sprints/2026-08-15-gherkin-adoption.md` §5, "Every finding,
verbatim" — 73 findings across 19 beads from run `1786778523-19`.

**Verdict: the taxonomy does NOT pass as written.** Residue is 8 of 73
(**11.0%**), against a 5% bar. Per the design's own rule the residue names what
is missing, and it named two things cleanly.

---

## The numbers

| | count | share |
|---|---|---|
| classified, **branch A** (the project's own quality) | 60 | 82.2% |
| classified, **branch B** (the skills this repository vends) | 5 | 6.8% |
| **residue** — fits no proposed kind | **8** | **11.0%** |
| total | 73 | |

### By kind

| kind | count | share of classified |
|---|---|---|
| `decision` | 27 | 41.5% |
| `gap` | 14 | 21.5% |
| `stale` | 9 | 13.8% |
| `tooling` | 5 | 7.7% |
| `unverifiable` | 4 | 6.2% |
| `sequencing` | 3 | 4.6% |
| `scope` | 3 | 4.6% |

---

## What the residue names

The 8 unclassifiable findings are not a long tail. They fall into two tight
groups plus one structural observation.

### Group 1 — proof, not friction (4 findings)

`wvd.1#4`, `r19#1`, `krb#3`, `nu2#1`

These record **how a claim was established**, not something that went wrong:

> `krb#3` — "Observed pre-fix failure (reintroduced the historical chdir bug by
> removing the `filepath.Abs(dir)` call…): failing all 34 cases. Restored the
> fix and the test passes clean, confirming the gate detects the class of
> defect the bead describes."

> `r19#1` — "the probe alone proves nothing — with no credential helper at all
> git already reaches origin through other config, so I ran a negative control
> first."

**This is the repository's own evidence standard being obeyed.** `AGENTS.md`
requires that a claim cite the command run and the output observed; these are
agents complying. Routing them as friction would be wrong in both directions:
they are not defects, and suppressing them would suppress exactly the rigour
the standard asks for.

**They should never become issues.** They belong in the report as evidence
against the finding or bead they support.

### Group 2 — environment, not the work order (3 findings)

`wvd.4#4`, `wvd.5#5`, `wvd.12#3`

Durable facts about a dependency or the local toolchain that the next agent
will hit, and that no work order could have carried:

> `wvd.4#4` — "The cucumber parser exports no error type — failures come back
> as an unexported slice type whose elements format as `(line:col): message`.
> …**Anything else parsing located errors out of this library will hit the same
> wall.**"

> `wvd.5#5` — "the cucumber parser accepts a second `Feature:` keyword, an
> `Examples:` block with a non-table body, and a bare step outside a scenario,
> all without error."

Nobody is at fault — not the bead, not a spec, not a skill. Classifying them
as `gap` would blame the work order for not carrying a fact about a third-party
library; classifying them as `tooling` would blame a skill for someone else's
parser. **They are exactly what `AGENTS.md`'s environmental-failure ledger is
for**, and this repository already keeps one.

### The structural observation — a finding that corrects a finding

`r19#2` is not a kind at all:

> "**correction to the line above** — the negative control showed the opposite
> of 'reaches origin anyway'…"

An agent noticed its own prior finding was wrong and wrote a second one to
correct it. Findings are append-only by design, so this is the only mechanism
available — but nothing in the format expresses the relationship, and an
aggregator reading `r19#1` and `r19#2` cold has to infer it from the phrase
"the line above".

**This is a template gap, not a taxonomy gap.** A finding needs an optional
`corrects: <finding-id>` field.

---

## What follows

Per the design's rule — *the residue names the missing kind; add it, never
widen an existing one* — **two kinds are added and one template field**:

| addition | fires when | routes to |
|---|---|---|
| `proof` | recording how a claim was established: an observed pre-fix failure, a negative control, a prior risk checked and cleared | **neither branch.** It is evidence, and it stays in the report attached to what it supports |
| `environment` | a durable fact about a dependency, tool or platform that the next agent will hit, which no work order could have carried | **the ledger**, not an issue — `AGENTS.md`'s environmental-failure ledger already exists for exactly this |
| `corrects: <finding-id>` | an optional finding field, not a kind | supersedes an earlier finding in place |

**Both new kinds route somewhere that is not "file an issue".** That is the
substantive discovery here: the original seven kinds all assumed a finding is a
candidate defect, and 7 of 73 are not — they are evidence and durable facts,
both worth capturing and neither worth filing.

**Recomputed residue with the two kinds added: 1 of 73 (1.4%)** — `r19#2`,
which the `corrects` field absorbs rather than a kind. **Under 5%: the extended
taxonomy passes.**

---

## What the branch ratio says

**Branch B is 5 of 73 — 6.8%.** Lower than expected for a repository actively
restructuring its own skills, and worth two observations.

**Two of the five are the same problem**, found independently by two different
beads:

> `wvd.6#1` — "the subagent prompt's Worktree line says the cwd IS the
> worktree, but Read/Edit tool calls against the shared-checkout path still
> succeed for Read and only fail on Edit."

> `wvd.11#1` — "early Read/Grep calls using absolute paths succeeded silently,
> reading from a sibling checkout rather than the worktree. Write/Edit/Bash
> refuse the same paths loudly, but Read/Grep do not."

So **4 distinct branch-B problems in a 21-bead run.** This is the first
recorded instance of the recurrence signal the design is built around — and it
was invisible in the 2026-08-15 report itself, which lists both under separate
beads with nothing connecting them. Under the proposed scheme both carry
`retro:tooling:subagent-prompts` and cluster automatically.

**A caution about reading the ratio too early.** This run implemented a
feature; a run whose subject is the skills themselves would show a very
different split. One ratio is not a trend, and the design already says branch B
shrinks as the skills mature — nothing here measures that yet.

---

## What the classification cost

**Every one of the 73 had to be read in full to classify it.** That is the
measurement the whole `<kind>` proposal rests on, and it holds: nothing in the
current `retro:` prefix narrows the work, so the classifier reads the prose and
decides. At 73 findings this is a real cost, and it recurs every retro.

**Six findings were ambiguous between two kinds** — `wvd.2#4` (`stale` vs
`scope`), `wvd.3#1` (`gap` vs `decision`), `wvd.7#7` (`decision` vs a plain
implementation bug), `wvd.7#9` (`unverifiable` vs `decision`), `wvd.8#6`
(`sequencing` vs routine work), `a7u#5` (`decision` vs a documented
contradiction).

**All six pairs route to the same branch**, so the ambiguity costs a cluster
rather than a route. That is the right failure mode and it is worth stating:
**the taxonomy only has to be right about the branch, and it was right about
the branch in all 65 classified cases.**

---

## Full classification

| finding | kind | branch |
|---|---|---|
| `wvd.1#1` | tooling | B |
| `wvd.1#2` | gap | A |
| `wvd.1#3` | decision | A |
| `wvd.1#4` | **proof** | — |
| `wvd.2#1` | decision | A |
| `wvd.2#2` | decision | A |
| `wvd.2#3` | decision | A |
| `wvd.2#4` | stale | A |
| `wvd.2#5` | decision | A |
| `wvd.3#1` | decision | A |
| `wvd.3#2` | decision | A |
| `wvd.4#1` | gap | A |
| `wvd.4#2` | unverifiable | A |
| `wvd.4#3` | decision | A |
| `wvd.4#4` | **environment** | ledger |
| `wvd.4#5` | decision | A |
| `wvd.5#1` | gap | A |
| `wvd.5#2` | gap | A |
| `wvd.5#3` | gap | A |
| `wvd.5#4` | decision | A |
| `wvd.5#5` | **environment** | ledger |
| `wvd.6#1` | tooling | B |
| `wvd.7#1` | decision | A |
| `wvd.7#2` | decision | A |
| `wvd.7#3` | gap | A |
| `wvd.7#4` | gap | A |
| `wvd.7#5` | unverifiable | A |
| `wvd.7#6` | stale | A |
| `wvd.7#7` | decision | A |
| `wvd.7#8` | gap | A |
| `wvd.7#9` | unverifiable | A |
| `wvd.7#10` | decision | A |
| `wvd.8#1` | decision | A |
| `wvd.8#2` | decision | A |
| `wvd.8#3` | gap | A |
| `wvd.8#4` | sequencing | A |
| `wvd.8#5` | decision | A |
| `wvd.8#6` | sequencing | A |
| `wvd.9#1` | scope | A |
| `wvd.9#2` | decision | A |
| `wvd.10#1` | gap | A |
| `wvd.10#2` | stale | A |
| `wvd.11#1` | tooling | B |
| `wvd.12#1` | stale | A |
| `wvd.12#2` | decision | A |
| `wvd.12#3` | **environment** | ledger |
| `wvd.13#1` | gap | A |
| `wvd.13#2` | unverifiable | A |
| `wvd.13#3` | decision | A |
| `wvd.13#4` | gap | A |
| `wvd.13#5` | tooling | B |
| `wvd.13#6` | gap | A |
| `4vz#1` | decision | A |
| `4vz#2` | decision | A |
| `4vz#3` | stale | A |
| `r19#1` | **proof** | — |
| `r19#2` | **corrects** `r19#1` | — |
| `r19#3` | decision | A |
| `r19#4` | tooling | B |
| `krb#1` | decision | A |
| `krb#2` | decision | A |
| `krb#3` | **proof** | — |
| `nu2#1` | **proof** | — |
| `nu2#2` | decision | A |
| `nu2#3` | sequencing | A |
| `nu2#4` | stale | A |
| `0i6#1` | scope | A |
| `0i6#2` | scope | A |
| `a7u#1` | stale | A |
| `a7u#2` | gap | A |
| `a7u#3` | stale | A |
| `a7u#4` | stale | A |
| `a7u#5` | decision | A |

Two beads in the run (`agent-plugins-3wd`, `agent-plugins-oq8`) recorded no
findings — both verify-and-close. Neither is residue; a clean bead records
nothing, which the capture protocol explicitly permits.
