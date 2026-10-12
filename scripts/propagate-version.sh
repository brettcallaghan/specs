#!/usr/bin/env sh
#
# propagate-version.sh — this repository's commands.propagate.
#
# `release cut` runs it in the stage checkout with RELEASE_VERSION set and
# commits what it changes, provided every path is in release_paths. It writes
# the version into the specs plugin manifest, so the plugin and the `specs`
# CLI released from the same commit carry one version. The marketplace that
# vends the plugin, in brettcallaghan/agent-plugins, serves it from the stable
# branch, which the publish workflow moves to each release.
set -eu

: "${RELEASE_VERSION:?RELEASE_VERSION is not set; release cut sets it}"
manifest="plugin/.claude-plugin/plugin.json"

python3 -I - "$manifest" "$RELEASE_VERSION" <<'PY'
import json, os, re, shutil, sys, tempfile
manifest, version = sys.argv[1], sys.argv[2]

# A regex substitution rather than a JSON round-trip, so the file's own
# layout survives; the result is parsed back before it replaces the target.
text = open(manifest, encoding="utf-8").read()
new, n = re.subn(r'("version"\s*:\s*")[^"]*(")', r"\g<1>" + version + r"\g<2>", text, count=1)
if n != 1:
    sys.exit(f"{manifest}: no version field to update")
if json.loads(new)["version"] != version:
    sys.exit(f"{manifest}: version was not written")

fd, tmp = tempfile.mkstemp(prefix=".propagate-", dir=os.path.dirname(manifest) or ".")
try:
    with os.fdopen(fd, "w", encoding="utf-8") as f:
        f.write(new)
    shutil.copymode(manifest, tmp)
    os.replace(tmp, manifest)
except BaseException:
    try:
        os.unlink(tmp)
    except FileNotFoundError:
        pass
    raise
PY
