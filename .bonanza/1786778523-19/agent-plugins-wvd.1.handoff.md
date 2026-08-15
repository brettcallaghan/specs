# agent-plugins-wvd.1 handoff

Caveat: every emitted index now carries `"features": []` and a changed `input_hash` — the profile schema is folded into the fingerprint, so the four `index-*` snapshot manifests were regenerated and any index written before this change is cold.
