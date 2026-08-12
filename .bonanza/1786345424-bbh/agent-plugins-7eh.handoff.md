# agent-plugins-7eh handoff

Caveat: validate-corpus.py now shells out to `scripts/sync-shared-references.sh --check` whenever a GENERATED COPY file falls in its SEARCH scope (one today: rfcs/findings-capture.md), so a stale vendored copy anywhere in the repo now fails both that gate and validate-corpus.py — duplicate CI output for one cause, accepted as the price of delegating instead of duplicating the orphan logic.
