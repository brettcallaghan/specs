# start-sprint Grooming Prompt Template

The prompt Phase 1.5 fills in and dispatches when beads need grooming. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the single
most common dispatch defect, because the subagent has no session context to infer
it from.

`{EPIC}` is the run epic this phase resolved; its id is the run id and
`bd ready --parent {EPIC} --exclude-type=epic` is the queue. `{RUN_ID}` does not
exist separately.

## Grooming subagent prompt template

```
You are grooming {N} beads to the Definition of Ready so an orchestrator can dispatch each to an isolated, zero-context sub-agent. Use the `groom-beads` skill and follow its full workflow (Definition of Ready, dependency wiring, tier assignment, C4 blast radius, validate). It loads the spec-owned Definition of Ready itself — do not go looking for that file by path.

Beads to groom:
{ID_1} — {TITLE_1}
{ID_2} — {TITLE_2}
...

For each bead: bring it to Ready by editing it in place with `bd update` — enrich thin descriptions rather than discarding author intent; wire only dependencies you can justify; assign a `tier` with a `tier_reason`; and where the repo has C4 docs, tag `c4_changes` / `c4_impacts` as arrays in metadata (the wave table you write below is bucketed off those two fields).

Do NOT implement any bead. Do NOT invent IDs, component names, or dependencies. Bounce, don't guess: any bead you cannot bring to Ready, leave as-is and report NOT-READY with the specific gap a human must close.

Return a report, one line per bead:
- `{ID}: ready — <notable change: description made context-independent / split into epic <id> + children <ids> / model assigned>`
- `{ID}: NOT-READY - <the gap that needs a human>`

Then produce the **wave preview** `groom-beads` produces, as its own block after the per-bead lines — `plan-sprint` consumes it below as a starting partition, so losing this block costs the whole preview:
- `wave N: {ID}, {ID}, … — <why they are safe together>`, one line per candidate wave, grouped under both constraints (dependencies, and no two beads sharing a `c4_changes` component in a wave)
- `sequencing: {ID} after {ID} — consumes <component>` for each pair whose `c4_impacts` intersects another bead's `c4_changes`. **Wire each one as a real `bd dep add` edge as you go** — the orchestrator dispatches from the re-queried ready front, so an edge is enforced and a preview line is not. Report any you could NOT wire (a task-to-epic constraint is refused by bd) so it is not mistaken for one that was
- `wave-preview: none (no C4 docs)` instead of the above WHEN the repo documents no architecture and you skipped C4 tagging — the reader must know the signal is absent, not empty
- `c4-untagged: {ID} — <why>` for any bead you could not tag (component with no doc, scope unclear)

FINALLY, as your last act, invoke the `plan-sprint` skill and follow it to produce the run's wave table. RETURN that table in your report — it is not written to a file, because it is derivable from the graph and a stored copy only drifts. You have just read every bead, which is why planning is yours and not a second agent's. Pass it:
- The run epic `{EPIC}` — its id IS the run id, `bd ready --parent {EPIC} --exclude-type=epic` IS the queue, and its metadata IS the manifest. Read the manifest; do NOT rewrite it
- The bead ids to plan: {PLAN_IDS}, which is `bd ready --parent {EPIC} --exclude-type=epic` as it stood when this prompt was written. Re-derive it from the epic rather than from any filter, and never add a bead the epic does not contain — admission decided that membership and may have excluded beads deliberately
- Checks: the declared ids `{CHECKS_IDS}`, run in full per bead; the baseline is `.metadata.checks_baseline` on `{EPIC}`
- Model lineup: {MODEL_LINEUP or "none declared"}
- Standing authorization: {STANDING_AUTH or "none"}
- Your own wave preview, above

Return the wave table it produces as the last block of your report, plus one line per row where you overrode your own preview and why.
```


---
