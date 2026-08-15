# agent-plugins-r19 handoff

Caveat: the scoped helper is keyed on `git remote get-url origin`, so a clone whose origin is an ssh URL now hits a loud `BB_DOLT_FATAL` instead of silently installing a bare global helper — intended, but it changes behaviour for any non-https checkout.
