#!/usr/bin/env bash
#
# Regenerate every artifact derived from schemas/*.strata.
#
# The .strata files are the single source of truth. Everything this
# script writes is generated and must not be hand-edited — each output
# carries a generated-by banner saying so.
#
# Usage:
#   scripts/generate-schemas.sh          # regenerate in place
#   scripts/generate-schemas.sh --check  # fail if any output is stale
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

SRC="schemas"
SCHEMA_OUT="cli/core/schema"
GO_OUT="cli/core/specmodel"
MD_OUT="plugin/skills/spec/references/generated"

# The swarm CLI's JSON surfaces are a second Strata package, not more
# files in the first. schemas/swarm carries its own strata.config.json,
# which makes it a package root and carves it out of the SpecModel
# package: without that, its types would land in the spec model's bundle,
# in cli/core/specmodel, and in the spec skill's generated Markdown
# reference — three places none of them belongs.
#
# Its JSON Schema joins the others under cli/core/schema because
# specs/contracts/contract-swarm-json.md names that directory as the
# validating artifact's home. Its Go types do not: they are the swarm
# binary's model, not the document corpus's.
SWARM_SRC="schemas/swarm"
SWARM_GO_OUT="cli/swarm/swarmmodel"
SWARM_MD_OUT="plugin/skills/bead-bonanza/references/generated"

if ! command -v strata >/dev/null 2>&1; then
  echo "error: strata not found on PATH." >&2
  echo "  The session-start hook installs it; see .claude/skills/strata/SKILL.md" >&2
  echo "  to build it by hand from a strata-lang/strata checkout." >&2
  exit 1
fi

CHECK=0
[[ "${1:-}" == "--check" ]] && CHECK=1

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# 1. Validate. --target ast runs the parser and every semantic check but
#    emits only JSON, so it is the cheapest correctness gate. Nothing is
#    written to the tree until this passes.
strata build "$SRC" --target ast --output "$WORK/ast" --error-format json

# 2. Bundled JSON Schema covering every type. The profile registers each
#    document type against a JSON Pointer into this bundle.
strata build "$SRC" --target json-schema --output "$WORK/schema"
mv "$WORK/schema/SpecModel.schema.json" "$WORK/schema/spec-model.schema.json"

# 2.5. Sanity assertion: every document-type struct must retain a non-empty
#      `required` list. strata commit 44c1955 (predating strata-lang/strata's
#      fix in 8997153, "Fix: allowAdditionalProperties: false silently
#      dropped struct inheritance (#289)") emitted structs like
#      `Adr: DocumentFrontmatter {}` as
#      {"additionalProperties": false, "properties": {}} — every inherited
#      envelope property AND the required list silently gone — while strata
#      itself still exited 0. `strata --version` cannot tell a good build
#      from a bad one (both report strata-cli 0.7.0), so this checks the
#      emitted SHAPE instead of the version string, which catches any future
#      regression of this kind rather than only this one commit
#      (agent-plugins-j7f).
#
#      The DocumentType enum (kebab-case values, e.g. "system-blueprint") is
#      used to derive the document-type struct names (PascalCase, e.g.
#      "SystemBlueprint") rather than hardcoding them, so the check stays
#      correct as document types are added or renamed in schemas/*.strata.
#      The enum itself is a plain string enum with no struct inheritance, so
#      it is not vulnerable to the bug this check guards against and remains
#      a reliable anchor even when the structs it names have been corrupted.
python3 - "$WORK/schema/spec-model.schema.json" <<'PYEOF'
import json
import sys

path = sys.argv[1]
with open(path) as f:
    schema = json.load(f)

defs = schema.get("definitions", {})

doc_type_enum = defs.get("DocumentType", {}).get("enum")
if not doc_type_enum:
    print(
        "error: generated schema has no definitions.DocumentType enum to "
        "derive the document-type list from",
        file=sys.stderr,
    )
    sys.exit(1)

missing = []
broken = []
for value in doc_type_enum:
    name = "".join(word.capitalize() for word in value.split("-"))
    definition = defs.get(name)
    if definition is None:
        missing.append(name)
        continue
    if not definition.get("required"):
        broken.append(name)

if missing:
    print(
        "error: generated schema is missing definitions for document "
        f"type(s): {', '.join(missing)}",
        file=sys.stderr,
    )
    sys.exit(1)

if broken:
    print(
        "error: generated schema has document type(s) with an empty or "
        f"missing 'required' list: {', '.join(broken)}\n"
        "  This is the shape of the regression fixed upstream by "
        "strata-lang/strata commit 8997153 ('Fix: allowAdditionalProperties: "
        "false silently dropped struct inheritance (#289)') — struct "
        "inheritance (e.g. 'struct Adr: DocumentFrontmatter {}') silently "
        "drops every inherited property and the required list, while strata "
        "still exits 0 and 'strata --version' still reports strata-cli "
        "0.7.0 either way.\n"
        "  Check the strata build on PATH: cat "
        "~/.cache/strata-cli/built-commit names the commit it was built "
        "from. See .claude/skills/strata/SKILL.md to rebuild if needed.",
        file=sys.stderr,
    )
    sys.exit(1)
PYEOF

# 3. The standalone specs.toml schema, filtered to SpecsConfig and its
#    transitive dependencies. This is the artifact contract-specs-
#    configuration.md names as normative for the shape of specs.toml.
strata build "$SRC" --target json-schema \
  --config "$SRC/overlays/specs-config.json" \
  --output "$WORK/cfg"
mv "$WORK/cfg/SpecModel.schema.json" "$WORK/schema/specs-config.schema.json"

# 3.5. A small Go wrapper embedding both bundles as a package, so a
#      consumer that needs a schema at runtime (the spec-driven-dev
#      profile) reads no file. cli/core/schema has no .strata source of
#      its own — the file below is generated by this script rather than
#      by strata — but it is still generated content and is written here
#      so it stays in cli/core/schema's synced output rather than a
#      hand-maintained file a regeneration could silently leave stale.
#      go:embed cannot reach outside its own package directory, which is
#      why the embedding lives beside the JSON rather than in the
#      package that consumes it.
cat >"$WORK/schema/embed.go" <<'GOFILE'
// Code generated by scripts/generate-schemas.sh. DO NOT EDIT.

// Package schema embeds the JSON Schema bundles this package holds, so a
// consumer needs no file access at runtime to read them.
package schema

import _ "embed"

// SpecModel is the union schema covering every document type this
// framework defines: spec-model.schema.json.
//
//go:embed spec-model.schema.json
var SpecModel []byte

// SpecsConfig is the standalone schema for specs.toml, filtered to
// SpecsConfig and its transitive dependencies: specs-config.schema.json.
//
//go:embed specs-config.schema.json
var SpecsConfig []byte

// SwarmModel is the schema for the swarm CLI's two JSON surfaces — the
// run manifest and the result every command emits under --full-result:
// swarm-model.schema.json. It is normative for
// specs/contracts/contract-swarm-json.md.
//
//go:embed swarm-model.schema.json
var SwarmModel []byte
GOFILE

# 4. Go structs for the CLI and its modules.
strata build "$SRC" --target go --output "$WORK/go"

# 5. Markdown reference shipped with the plugin, describing the shape of
#    frontmatter to anyone authoring a document by hand.
strata build "$SRC" --target markdown --output "$WORK/md"

cat >"$WORK/md/README.md" <<'BANNER'
# Generated schema reference

**Generated — do not edit.** Every file in this directory is produced by
`scripts/generate-schemas.sh` from the `.strata` sources in `schemas/`.
Edit those and re-run the script; a change made here is lost on the next
regeneration.

These pages describe the shape of the frontmatter and configuration a
document author writes. The normative artifacts derived from the same
sources are the JSON Schemas under `cli/core/schema/`.
BANNER

# 6. The swarm CLI's schemas, from the nested package. Validated the same
#    way the first package is — --target ast runs every semantic check and
#    writes nothing to the tree — before either artifact is emitted.
strata build "$SWARM_SRC" --target ast --output "$WORK/swarm-ast" --error-format json

strata build "$SWARM_SRC" --target json-schema --output "$WORK/swarm-schema"
mv "$WORK/swarm-schema/SwarmModel.schema.json" "$WORK/schema/swarm-model.schema.json"

strata build "$SWARM_SRC" --target go --output "$WORK/swarm-go"

# 6.6. Markdown reference for the swarm CLI's own model — the run
#      manifest's key set and the command-result envelope — describing
#      what a `swarm manifest` or `--json` reader will find, from the
#      same docstrings the Go and JSON Schema targets draw on. Vendored
#      into bead-bonanza rather than left as a command to run: the key
#      set is read while deciding whether to run a command at all, and a
#      run-record document that pointed at nothing but `swarm manifest
#      get --run <epic>` would need a live run in flight to be useful.
strata build "$SWARM_SRC" --target markdown --output "$WORK/swarm-md"

cat >"$WORK/swarm-md/README.md" <<'BANNER'
# Generated swarm model reference

**Generated — do not edit.** Every file in this directory is produced by
`scripts/generate-schemas.sh` from the `.strata` sources in
`schemas/swarm/`. Edit those and re-run the script; a change made here is
lost on the next regeneration.

`manifest.md` describes the run manifest's key set: which keys exist,
their types, the stage each becomes required at (the struct that first
declares it), and what each key means. `output.md` describes the
result every `swarm` command emits under `--full-result`. The normative
artifact derived from the same sources is the JSON Schema under
`cli/core/schema/`.
BANNER

# 6.5. The same shape assertion the spec model gets, for the reason it
#      gets it: the strata regression this guards against emits an
#      inheriting struct with an empty properties map while exiting 0, and
#      the manifest is a chain of five inheriting structs. An empty
#      RunManifest would make every manifest key read as unknown.
python3 - "$WORK/schema/swarm-model.schema.json" <<'PYEOF'
import json
import sys

with open(sys.argv[1]) as f:
    schema = json.load(f)

defs = schema.get("definitions", {})
manifest = defs.get("RunManifest")
if manifest is None:
    print("error: the swarm schema has no RunManifest definition", file=sys.stderr)
    sys.exit(1)

properties = manifest.get("properties") or {}
# Every key the chain declares has to survive inheritance into the leaf.
for stage in ("PreFlightKeys", "ScopeKeys", "PlanKeys", "DispatchKeys", "RetroKeys"):
    stage_def = defs.get(stage)
    if stage_def is None:
        print(f"error: the swarm schema has no {stage} definition", file=sys.stderr)
        sys.exit(1)
    missing = sorted(set(stage_def.get("properties") or {}) - set(properties))
    if missing:
        print(
            f"error: RunManifest is missing key(s) declared by {stage}: "
            f"{', '.join(missing)}\n"
            "  This is the shape of the strata regression the spec model's own "
            "assertion guards against (strata-lang/strata 8997153): struct "
            "inheritance silently drops inherited properties while strata still "
            "exits 0. Check ~/.cache/strata-cli/built-commit.",
            file=sys.stderr,
        )
        sys.exit(1)
PYEOF

sync_dir() {
  local src="$1" dest="$2"
  if [[ $CHECK -eq 1 ]]; then
    if ! diff -r -q "$src" "$dest" >/dev/null 2>&1; then
      echo "error: $dest is stale — run scripts/generate-schemas.sh" >&2
      diff -r "$src" "$dest" >&2 || true
      return 1
    fi
  else
    rm -rf "$dest"
    mkdir -p "$(dirname "$dest")"
    cp -r "$src" "$dest"
  fi
}

rc=0
sync_dir "$WORK/schema" "$SCHEMA_OUT" || rc=1
sync_dir "$WORK/go" "$GO_OUT" || rc=1
sync_dir "$WORK/md" "$MD_OUT" || rc=1
sync_dir "$WORK/swarm-go" "$SWARM_GO_OUT" || rc=1
sync_dir "$WORK/swarm-md" "$SWARM_MD_OUT" || rc=1

if [[ $rc -ne 0 ]]; then
  exit 1
fi

if [[ $CHECK -eq 1 ]]; then
  echo "generated artifacts are up to date"
else
  echo "regenerated:"
  echo "  $SCHEMA_OUT  (JSON Schema)"
  echo "  $GO_OUT  (Go)"
  echo "  $MD_OUT  (Markdown)"
  echo "  $SWARM_GO_OUT  (Go, swarm)"
  echo "  $SWARM_MD_OUT  (Markdown, swarm)"
fi
