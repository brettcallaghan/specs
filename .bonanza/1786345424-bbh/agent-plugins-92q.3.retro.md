bead-bonanza's Phase 0 is `SKILL.md:54-142` — **89 lines, ~4,600 bytes**, all of
it inside the ~20,000-byte window that survives compaction, and most of it
mechanical. It spends that budget describing shell commands rather than
judgment, while the judgment-bearing phases sit past the truncation point.

Audit of the six steps, done 2026-08-10:

| # | Step | Mechanical? |
|---|---|---|
| 1 | Load `## bead-bonanza gates` + `## model lineup` | detection only |
| 2 | Beads telemetry (`SKILL.md:74-98`, 25 lines) | fully |
| 3 | Run scratch directory + `SCRIPTS` path (`:99-123`) | fully |
| 4 | TaskCreate fallback (`:124-129`) | fully, as detection |
| 5 | Push mechanics (`:130-135`) | fully — and it is currently an untested *claim* |
| 6 | Draft-PR ask (`:136-142`) | detection only |

Steps 2, 3 and 5 together are **3,047 bytes** of pure mechanics.

Step 5 is the sharpest case. It is not an action at all today — it is an
assertion ("the sandbox blocks `.git/config` writes, so `--set-upstream` never
persists") carried forward as prose. Bead `agent-plugins-d6x` establishes that
ledger claims must cite the command that proved them; a pre-flight script is
the mechanism that makes that true for the environment, by measuring at run
start what is currently asserted from memory.

Six checks belong in pre-flight and exist nowhere today:

- `CLAUDE_CODE_SUBAGENT_MODEL` — it overrides the per-invocation `model`
  parameter, so if it is set every tier assignment in the dispatch plan is
  silently ignored and the Phase 4 escalation table is fiction. Verified NOT
  set in this container, so it is a latent trap rather than an observed failure.
- `bd` present, `bd where` resolving to a real database.
- `jq` present (the close script, the harvest and the C4 bucketing all need it).
- **Working tree clean at t=0.** A tree already dirty at run start poisons every
  later `bb-verify-skip.sh` result and every pathspec guard.
- Gate baseline capture — a gate that already fails on an untouched tree must be
  recorded as the baseline, or every bead is measured against "exit 0" and
  reports a failure it did not cause.
- The command permissions in `.claude/settings.local.json`.

One check is deliberately NOT in scope: whether an `Agent` tool exists. That is
a harness property, not observable from a shell, and it is the trigger for the
main-loop fallback — it stays an instruction in SKILL.md.

Environment for portability: the script must run on macOS and Linux.


## Retro — implementation

### Findings

- AGENTS.md's `## bead-bonanza gates` block runs 163 lines here and its
  `- **Label:** \`cmd\`` bullet shape continues into `###` prose subsections
  (the KNOWN RED note, the strata install notes). The bead specified "run the
  gate profile's commands" without saying where the profile ends; a naive
  parse over the whole block produced eleven non-commands. Bounded to the
  bullets before the first `###`.
- AC "--self-test exits 0 on Linux and on macOS, and the retro records the
  observed output of both runs" cannot be satisfied from this container —
  there is no macOS host. Linux run: 26 checks, RESULT: PASS, exit 0. The
  macOS half is unverified; the BSD `stat -f` branch is exercised only by the
  probe's fallback logic, not by a BSD `stat`.
- `dolt` is not on PATH here, so only the absent-dolt branch of the telemetry
  step ran. The present-dolt branch (including the sandbox-refusal detection)
  is unexercised.
- The reduction is smaller than the bead's 3,047-byte estimate: SKILL.md went
  46,485 → 44,617 bytes (-1,868), because publishing the new script's
  interface and its bullet adds ~900 bytes back in the Reference-files
  section, which the estimate did not net out.

### What I tried that did not work

- First cut echoed only the four gate bullets instead of the two AGENTS.md
  blocks. That is not what "load the two blocks" means for a reader who then
  still has to open AGENTS.md for the KNOWN RED exception — so it now echoes
  both blocks whole and the read is genuinely replaced.
- CI detection first used one `ls .github/workflows/*.yml *.yaml`. `ls` exits
  nonzero when either glob matches nothing, so a repo with only `.yml`
  workflows reported `BB_PR_CI=no` — wrong here, where two workflows are
  `pull_request`-triggered. Replaced with a per-file loop.
- The portability self-lint first ran over the whole source and flagged the
  header's own prose ("no `stat -c`, no `sed -i`…"). It now strips whole-line
  comments first, so the header stays free to name the constructs it bans.

### What I decided silently

- **Sandbox:** the script runs WITH the sandbox on. It detects the refused
  `dolt config --global` write, skips it, and prints the single command to
  re-run with the sandbox off. Rationale: the alternative would attach
  `dangerouslyDisableSandbox: true` to an invocation that also runs the whole
  gate suite, to buy one telemetry write that is a no-op wherever `dolt` is
  absent. Stated on the published usage line (`# sandbox ON`) and in the
  script header.
- Added `--gate-baseline` as an opt-in flag (the Design permitted guarding it)
  rather than making baseline capture unconditional; `--self-test` takes no
  run id; a run-id containing `/` or a space exits 2.
- Swept the stale `Phase 0 step 1/1a/4/5` cross-references in SKILL.md,
  recovery.md and subagent-prompts.md to non-numeric ones. They were already
  off by one before this change and renumbering would have made them worse.
