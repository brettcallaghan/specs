# agent-plugins-wvd.4 handoff

Caveat: DISC must bridge `corpus.FeatureTagRules` to `gherkin.Options`'s two identity functions, and must derive a `FeatureCode` from a matched feature identifier itself — `cli/core/gherkin` reads tags as strings and resolves no pattern of its own.
