`dispatch.md` is the sprint plan. Three defects make it unfit for the one job
that matters most — being the thing a recovering orchestrator reads instead of
re-deriving the run.

**1. It is optional.** `SKILL.md:370-391` skips the formal plan when the queue
is under 5 beads, when all beads land in one bucket, or when the user gave an
explicit ID list — replacing it with a one-line statement in chat. Chat is the
one store that does not survive. Recovery does not care how large the run was;
a two-bead run interrupted at bead one needs the plan exactly as much as a
twenty-bead run does.

**2. It identifies scope by filter, not by enumeration.** The plan is produced
from `bd ready --label <x>` and the table is described as the output of that
query. Two failures follow. Beads get added to an epic or gain the label during
the sprint, and a filter re-evaluated later silently pulls them in. And
`references/recovery.md`'s recovery path — reinforced by the orchestrator-death
procedure being added in a sibling bead — instructs a recovering orchestrator to
"read `dispatch.md`; this is the plan; do not re-derive it". **A filter
re-evaluated at recovery returns a different queue than the one the run
committed to.** That is not a stale plan; it is a silent scope change during the
least-supervised moment of the run.

**3. It is written once and never amended.** Observed twice in run
`1786121603-ri`: wave 5 was planned as one shared batch (`9kb.6`, `9kb.7`,
`9kb.9`) and re-decided mid-run as strictly sequential; the re-decision was
recorded in the epic's tracker notes because the skill offered no home for it.
Separately the plan's wave ordering and the commits' actual ordering diverged
(`.6` 21:34, `.7` 21:58, `.8` 22:17, `.9` 22:24 against a plan that had `.8`
alone after the batch). A plan that silently stops describing the run is worse
than no plan, because recovery trusts it.

Framing recorded 2026-08-10: the plan is the sprint plan, and it is written for
whoever picks the sprint back up — which may be an agent with no memory of
having written it.


## Retro — implementation

### Findings

- `scripts/check-skills.sh` treats any backticked token that looks like a
  filename as a path to resolve relative to the file it appears in. Naming
  `pre-flight-check.sh` in prose inside `references/dispatch-plan.md` was two
  hard errors, not a warning. Written as `../scripts/pre-flight-check.sh`
  instead. Nothing in the bead or in the skill's own conventions warns that
  prose mentions of a sibling script must be spelled as resolvable paths.
- The Design's line anchors (`SKILL.md:370-391`, `:534-558`) were already stale
  by the time this bead ran: the sibling bead in the same batch
  (`agent-plugins-92q.3`) rewrote Phase 0 first and shifted everything below
  it. Matched the regions by text instead. Line anchors in a bead that shares a
  file with another bead in the same batch cannot be trusted.

### What I tried that did not work

- First cut kept an abbreviated wave-table template inline in SKILL.md and
  moved only the manifest and amendments log to `references/`. That reproduces
  the format in two places, which is exactly the drift the amendments rule
  exists to prevent — and Design item 4 already says the format is a
  dispatch-time lookup. The whole format now lives in
  `references/dispatch-plan.md` and SKILL.md carries only the three
  load-bearing rules plus the pointer.

### What I decided silently

- Put the new `### The dispatch plan` section at the END of Phase 2.5, where
  the conditional "When you write the formal plan" section was, rather than at
  the start where the deleted lean path sat. The plan's columns (class, bucket,
  tier, push mode) are all decided by the subsections between those two points;
  an unconditional write instruction placed first would reference columns the
  reader has not met yet.
- Kept the decisions-doc lean path in the Operating principles bullet and
  narrowed the bullet to it, rather than deleting the bullet. Only the dispatch
  plan lost its exemption; nothing in this bead touches the decisions doc.
