# agent-plugins-92q.22 handoff

Paths: plugin/skills/bead-bonanza/scripts/bb-close-bead.sh, plugin/skills/bead-bonanza/references/subagent-prompts.md, plugin/skills/bead-bonanza/references/recovery.md, plugin/skills/bead-bonanza/SKILL.md, scripts/check-skills.sh

Caveat: close with the NEW four-argument form this bead introduces (no body-file argument); the close script does not write commit trailers, so `git commit --amend` them on immediately afterwards and read the sha from the branch.
