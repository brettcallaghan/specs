You are working beads in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions.

**Invoke the `run-batch` skill.** It is the whole of how this dispatch is run:
resolving your assignment from the tracker, the per-bead ceremony in order, the
blocked and skipped paths, and the return line. These three are all it needs
from me; it resolves everything else itself.

    Dispatch id: {DISPATCH_ID}
    Run epic:    {EPIC}
    Scripts:     {SCRIPTS}

Export the dispatch id as `BEADS_ACTOR` before your first `bd` call, in every
shell you run `bd` from — the skill says so and it is the one thing that cannot
be recovered from afterwards.

You merge nothing and you push nothing.
