`bead-bonanza`'s worktree-parallel mode assumes each subagent branches from the
run's working branch and that the orchestrator merges back into it. The
platform default is different.

Claude Code docs (`code.claude.com/docs/en/sub-agents`, frontmatter): setting
`isolation: worktree` gives the subagent "an isolated copy of the repository
**branched by default from your default branch rather than the parent
session's `HEAD`**". And `code.claude.com/docs/en/worktrees`: `"head"`: branch
from your current local `HEAD` ... **Use this when isolating subagents that
need to operate on in-progress work.**

So with the default `worktree.baseRef: "fresh"`, every worktree subagent
branches from `origin/main`. Consequences for a bead-bonanza run:

- Every merge back carries the delta between `main` and the working branch.
- **Any bead built on an earlier bead's commit in the same run silently loses
  that foundation.** Run 1786121603-ri would have been badly hit: 9kb.2-9kb.9
  all build on 9kb.1's model, and the wave plan explicitly sequenced them.

Neither `SKILL.md`'s dispatch-mode section nor the "Worktree parallel notes" in
`references/subagent-prompts.md` mentions `worktree.baseRef` at all.

This is a prerequisite for using worktree-parallel at all. The sibling bead
corrects the false claim that the mode is impossible; this one establishes
whether it is SAFE.


## Retro — implementation

**Findings** (also filed as `retro:` comments on this bead)

1. Conclusions 3 and 4 combine into a merge hazard this bead did not have a
   mandate to fix. All worktree agents share ONE database (conclusion 3), so
   the mandated `bd export > .beads/issues.jsonl` (conclusion 4) makes every
   worktree branch commit a near-identical FULL export of the same shared
   tracker. With no merge driver configured for that file (CLAUDE.md), two or
   more worktree branches in a wave will conflict textually on
   `.beads/issues.jsonl` at merge time — every time, by construction. The
   orchestrator should expect it and resolve by re-exporting on the merged
   tree, the same way a regen step resolves a fixture conflict. This belongs
   with the skip conditions owned by `agent-plugins-92q.9`, not here.

**What I tried that did not work** — nothing failed. No probes were re-run;
the measurements were handed over intact.

**What I decided silently**

- Took the optional pre-flight probe. `pre-flight-check.sh` now reads
  `.worktree.baseRef` from the user-scope settings file (the jq plumbing was
  already there for the telemetry deny) and reports it as an **advisory** line
  plus `BB_WORKTREE_BASEREF`, never a blocker — most runs never dispatch a
  worktree wave, and blocking them on a global setting they do not use would
  be wrong. Verified: self-test PASS (26 checks), live run emits
  `worktree.baseRef: head (advisory)` and `BB_WORKTREE_BASEREF='head'`.
- Cited the fact in SKILL.md through `BB_WORKTREE_BASEREF` rather than by
  naming the settings file and key, so Phase 0's "do not re-assert from memory
  what the script just measured" rule holds and the bullet stays 5 lines.
  Net SKILL.md growth: 43,621 -> 43,986 bytes.
- Structured the reference section as a two-row measured table plus a copy-
  pasteable verification instruction for the first worktree subagent, so the
  orchestrator can paste rather than compose it under load.
- Left Phase 2.5's "Skip this mode if..." list untouched per the handover;
  `agent-plugins-92q.9` owns it.

**Observed results** (measured by the orchestrator; quoted, not re-run)

Session HEAD `0ccc092` on `claude/bead-bonanza-improvements-j6jxjv`, 8 commits
ahead of default branch `main` at `5390f1d`. Two read-only `isolation:
"worktree"` probes. Discriminator: `scripts/pre-flight-check.sh` and the string
`After any compaction` in bead-bonanza's SKILL.md exist only on the branch.

Probe 1 — no `worktree` key in the user settings file:

    git rev-parse --short HEAD                -> 5390f1d   (DEFAULT BRANCH)
    git merge-base --is-ancestor 0ccc092 HEAD -> exit 1    (HEAD not an ancestor)
    git rev-list --count 5390f1d..HEAD        -> 0
    test -e .../scripts/pre-flight-check.sh   -> exit 1    (absent)
    wc -l .beads/issues.jsonl                 -> 105       (main checkout: 128)
    git rev-parse --git-common-dir            -> /home/user/agent-plugins/.git
    ls -d .beads/embeddeddolt                 -> No such file or directory
    bd where -> /home/user/agent-plugins/.beads
                database: /home/user/agent-plugins/.beads/embeddeddolt
    bd worktree info -> Main repo: /home/user/agent-plugins
                        Beads: local (no redirect)
    bd list                                   -> 63 issues

Probe 2 — after setting baseRef to head in user-scope settings:

    git rev-parse --short HEAD                -> 0ccc092   (SESSION HEAD)
    git merge-base --is-ancestor 0ccc092 HEAD -> exit 0
    git rev-list --count 5390f1d..HEAD        -> 8
    test -e .../scripts/pre-flight-check.sh   -> exit 0    (present)
    grep -c "After any compaction" .../SKILL.md -> 1
    wc -l .beads/issues.jsonl                 -> 128       (current)
    ls -d .beads/embeddeddolt                 -> No such file or directory
    bd where -> /home/user/agent-plugins/.beads (main checkout)

Acceptance outcome **(a)**: worktree-parallel is SAFE, conditional on
`baseRef: "head"` plus the per-wave ancestor check. The setting takes effect
immediately — probe 2 ran minutes after the write, in the same session.
Conclusion 3 independently reproduces `agent-plugins-t09` in this repository.

**Limits, stated not papered over**

- Project-root `.claude/settings.json` was NOT tested as a location for the
  key. Only user scope is verified; the docs say so rather than implying the
  project file works.
- This repository has no `.worktreeinclude` (`ls` -> No such file or
  directory), so gitignored-file copying was observed only in its default
  form: `.beads/embeddeddolt/` absent from both worktrees — which turned out
  not to matter, because of conclusion 3.
