# agent-plugins-4sp handoff

Caveat: Rejected with no file changes — a `core.symlinks=false` checkout turns a symlinked reference into a 58-byte text file holding the path, and every gate in this repo still reports PASS, so the loss is silent; independently, `scripts/install-spec-skill.sh` uses `cp -RL` and materialises the link back into a duplicate file.
