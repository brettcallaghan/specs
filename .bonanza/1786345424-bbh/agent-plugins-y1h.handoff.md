# agent-plugins-y1h handoff

Caveat: with the last delegated file gone, `validate-corpus.py` no longer invokes `sync-shared-references.sh --check` as a subprocess (the `if delegated:` branch is now false), so that authority runs only as its own gate — the delegation machinery is intact and dormant, not dead, and re-arms the moment any GENERATED COPY file lands under `rfcs/**` or `plugin/skills/*/references/`.
