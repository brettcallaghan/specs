# Epic Definition of Ready

The per-bead Definition of Ready says when a **bead** can be dispatched. This
says when an **epic** is fit to scope a whole run — the question an orchestrator
asks when it keys a run on one epic id.

Check it whenever you groom under an epic.

## The checks

- [ ] **It has children.** An epic with none scopes an empty run, which reads
      exactly like a finished one.
- [ ] **Every child meets the per-bead DoR**, or is explicitly out of scope and
      says so. A single un-ready child stalls a wave that has already started.
- [ ] **Dependency edges among the children are wired, not implied.** An order
      that lives only in prose is honoured only if someone reads the prose; an
      edge is honoured by the tracker at every dispatch.
- [ ] **Nothing the run needs is derivable-only-by-you.** Anything an
      orchestrator cannot recompute from the graph belongs on the epic.

## Flatness is preferred, not required

An epic whose children are all beads gets exact answers from
`bd swarm validate`, `bd swarm status` and `bd epic status`.

An epic containing **sub-epics** still dispatches correctly: the ready front
(`bd ready --parent <epic> --exclude-type=epic`) is transitive and honours
dependency edges that cross between sub-epics. But those three commands are
**one level deep** — they report the sub-epics as work items and never see
their leaves, and `bd epic status` counts sub-epics rather than beads.

So a nested epic costs the wave preview and the progress count, both of which
the orchestrator can recompute, and costs nothing else.

**Report which shape the epic is in; do not require flatness.** Requiring it
would push an author to move beads out of the epics that own them — and a bead
moved out of its epic makes that epic report complete when it is not, which is
a real loss traded for a display detail.
