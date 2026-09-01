# Epic Definition of Ready

The per-bead Definition of Ready says when a **bead** can be dispatched. This
says when an **epic** is fit to scope a whole run — the question an orchestrator
asks when it keys a run on one epic id.

Check it whenever you groom under an epic.

## The checks

- [ ] **It has children.** An epic with none scopes an empty run, which reads
      exactly like a finished one.
- [ ] **Every child meets the per-bead DoR**, or is explicitly out of scope and
      says so. A single un-ready child stalls a run that has already started.
- [ ] **Dependency edges among the children are wired, not implied.** An order
      that lives only in prose is honoured only if someone reads the prose; an
      edge is honoured by the tracker at every dispatch.
- [ ] **Nothing the run needs is derivable-only-by-you.** Anything an
      orchestrator cannot recompute from the graph belongs on the epic.
- [ ] **Every child carries an assignee, and the check NAMES the ones that do
      not.** An unowned bead is not idle — it is claimed by whoever gets there
      first, and `bd update <id> --claim` writes the claimer into an empty
      `assignee` and exits 0. Nothing reports it, so the run's ownership view
      decays silently.

      **The producer is `sprint-plan`'s assign step**, which writes
      `dev-<run-id>-<batch-id>` onto every planned bead once the partition
      that value derives from exists. Nothing earlier in a run writes an
      assignee, so this check runs AFTER that step. Run before it and it names
      every child of every epic on every run — a check that always fires is
      indistinguishable from one that never fires.

      ```sh
      bd list --parent <epic> --status all --json \
        | jq -r '.[] | select((.assignee // "") == "") | .id'
      ```

      **Read `assignee` out of `--json`, never off `bd show`'s `Owner:` line —
      that line prints `created_by`.** Measured on bd 1.2.2: `bd show <id>`
      printed `Owner: dev-4` for a bead whose `--json` returned
      `assignee: null`, `created_by: dev-4`. A check written against the Owner
      line passes on an epic where not one bead is assigned, which is the exact
      condition it exists to catch.

      `--status all` is not optional either: `bd list` defaults to open issues,
      so a check written without it silently skips every closed and deferred
      child.

- [ ] **The dev pool fits the concurrency cap, and the check NAMES the count
      and the cap.** Require `N_dev <= concurrency_cap - 1` — the coordinator
      holds a slot of its own. Above that the surplus agents queue behind the
      cap and their queues never drain; the run looks alive and makes no progress.

      ```sh
      cap=$(swarm manifest get --run <epic> --key concurrency_cap --json \
              | jq -r .concurrency_cap)
      n=$(bd list --parent <epic> --status all --json \
            | jq '[.[].assignee | select(. != null) | select(startswith("dev-"))]
                  | unique | length')
      echo "dev pool $n, cap $cap, budget $((cap - 1))"
      ```

      **Count the DEV POOL, not every distinct assignee.** Only the dev pool is
      concurrent; the ceremony agents — groom, plan, merge, review, retro,
      report — run one at a time in a chain the dependency graph already
      serializes, so they cost one slot between them rather than one each.
      Counting all distinct assignees under-subscribes the dev pool by the
      number of ceremony roles the run happens to have used.

      WHERE the epic's manifest carries no `concurrency_cap` — `swarm manifest
      require --run <epic> --key concurrency_cap` answers that as a status —
      report the check as **unmeasurable** and say so. An epic that was never
      pre-flighted has
      no cap to plan against, and a default substituted here is a number
      nobody measured reading as one somebody did.

## Flatness is preferred, not required

An epic whose children are all beads gets exact answers from
`bd swarm validate`, `bd epic status` and `bd list --parent`.

An epic containing **sub-epics** still dispatches correctly: the ready set
(`bd ready --parent <epic> --exclude-type=epic`) is transitive and honours
dependency edges that cross between sub-epics. But those three commands are
**one level deep** — they report the sub-epics as work items and never see
their leaves, and `bd epic status` counts sub-epics rather than beads.

So a nested epic costs the progress count and the graph preview, both of which
an orchestrator can recompute by walking the tree, and costs nothing else.

**Report which shape the epic is in; do not require flatness.** Requiring it
would push an author to move beads out of the epics that own them — and a bead
moved out of its epic makes that epic report complete when it is not, which is
a real loss traded for a display detail.
