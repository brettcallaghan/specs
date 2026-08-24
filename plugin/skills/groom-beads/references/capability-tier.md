# Recording a bead's capability tier

Read this at step 4, when you are writing tier metadata. The **rubric** for
which tier a bead needs is owned by `shared/work-order-field-guide.md`; this
file is how the answer gets recorded so an orchestrator that never read this
skill can act on it.

## Contents

- The metadata keys, and which are `bd`'s own convention rather than ours
- `--set-metadata` versus `--metadata` — additive versus merging (shallowly)
- The rules that bind whatever the rubric says

Record the tier in **metadata** so it rides with the bead and the orchestrator
can read it at dispatch:

```
bd update <id> --set-metadata tier=standard \
  --set-metadata tier_reason="localized CRUD + tests, low ambiguity" \
  --set-metadata est_size=S \
  --set-metadata execution_suggested_model=standard \
  --set-metadata execution_reasoning_effort=medium \
  --set-metadata execution_agent_type=worker
```

The `execution_*` keys are `bd`'s own documented convention (`docs/METADATA.md`),
and writing them is what makes a groomed bead portable to an orchestrator that
never read this skill. Three things that convention settles:

- **`execution_suggested_model` takes the TIER, not a model name.** The
  convention calls it "a capability-tier suggestion", and says a consumer on a
  different provider substitutes a model of the same tier rather than dropping
  the hint. So it carries the same value as `tier`.
- **`execution_reasoning_effort`** is a canonical advisory scale — `low`,
  `medium`, `high`, `xhigh`. Record it rather than leaving it unset: a consumer
  on a different scale maps to its nearest native level, but it cannot map an
  absent value.
- **`execution_agent_type`** is the worker class (`explorer`, `worker`,
  `mixed`).

`tier` and `tier_reason` stay. `tier_reason` has no counterpart in the
convention, and `tier` is what 205 of this tracker's beads already carry — the
orchestrator reads `execution_suggested_model` first and falls back to `tier`,
so old beads keep routing correctly and no bulk rewrite is required. Backfill
is optional and per-bead:

```
bd update <id> --set-metadata execution_suggested_model="$(bd show <id> --json | jq -r '.[0].metadata.tier')"
```

Two flags write metadata and they are not interchangeable — probe both
before a bulk pass:

- `--set-metadata key=value` is **repeatable and additive**: it leaves
  other keys intact, so it is what you want for touching one field on an
  already-groomed bead. It stores every value as a **string** — an array
  written this way reads back as the literal `["a","b"]`, not a list, so
  reach for `--metadata` instead when a value is an array or object.
- `--metadata '<json>'` takes real JSON types and **merges at the top
  level**: a key you do not restate survives. The merge is shallow,
  though — a **nested object value is replaced wholesale**, not merged,
  so a partial write to a nested key (e.g. a manifest's `standing_auth`)
  discards that object's other siblings unless you restate them too.

  Measured on bd 1.2.2 (dev) in a throwaway scratch database:

  ```
  bd update <id> --metadata '{"alpha":1,"beta":"two","gamma":["x","y"]}'
  bd update <id> --metadata '{"delta":9}'
  bd show <id> --json | jq -c '.[0].metadata'
  -> {"beta":"two","alpha":1,"delta":9,"gamma":["x","y"]}
  # all four keys survive — top-level merge, not replace

  bd update <id> --metadata '{"nest":{"a":1,"b":2}}'
  bd update <id> --metadata '{"nest":{"c":3}}'
  bd show <id> --json | jq -c '.[0].metadata'
  -> {"nest":{"c":3}}
  # "a" and "b" are gone — the nested object VALUE was replaced whole
  ```

Beads carry a **tier** (`light | standard | advanced | frontier`, least to
most capable), not a model name. The orchestrator resolves tier → model at
dispatch from the project's `sf.toml` `[models]` lineup, so a bead groomed today
still routes correctly after the lineup changes. The routing rubric is owned by
`shared/work-order-field-guide.md`. Apply it, and:

- **Assign the LOWEST tier that can do the work.** Over-routing is invisible
  and paid for on every run; under-routing announces itself as an escalation
  the orchestrator reports back. Bias toward the cheaper call.
- **Always record `tier_reason`** — a bare tier can't be audited, and the
  retro uses the reason to tell whether routing was calibrated.
- **Assign a default, never null** — a mis-groomed bead must not stall a run.
  WHEN genuinely unsure, `standard` and say so.
- Set a named `model` **only** as an override, when a specific model is
  genuinely required. Normal beads carry `tier` and no `model`.
