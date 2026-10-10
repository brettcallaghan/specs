#!/usr/bin/env bash
# Builds the `specs` release archives for every supported target, with the
# version stamped into each binary, plus a SHA256SUMS file covering them.
#
#   scripts/build-release.sh <version> <out-dir>
#
# <version> is the tag without its cli/ prefix, e.g. v0.26.0; `specs
# --version` prints it. Archives are named specs_<version>_<os>_<arch>.tar.gz
# (.zip on windows) and hold the binary alone. The release workflow publishes
# them; the cli workflow runs this on every pull request so a target that
# stops building fails there rather than at release time.
#
# Every binary is cross-compiled on one machine from one checkout
# (AC-SPEC-009.7). CGO_ENABLED=0 makes each a single static file with no libc
# to find at run time (AC-SPEC-009.4); -trimpath strips the builder's paths so
# the same source gives the same bytes from any checkout directory.
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <version> <out-dir>" >&2
  exit 2
fi
version="$1"
out="$2"

targets="
linux/amd64 linux/arm64
darwin/amd64 darwin/arm64
windows/amd64 windows/arm64
"

root="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$out"
out="$(cd "$out" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

for target in $targets; do
  os="${target%/*}"
  arch="${target#*/}"
  exe=specs
  [ "$os" = windows ] && exe=specs.exe
  name="specs_${version}_${os}_${arch}"
  echo "==> $target"
  mkdir -p "$work/$name"
  (cd "$root/cli" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath \
    -ldflags="-s -w -X github.com/brettcallaghan/agent-plugins/cli/cmd.version=$version" \
    -o "$work/$name/$exe" ./specs)
  if [ "$os" = windows ]; then
    (cd "$work/$name" && zip -q -X "$out/$name.zip" "$exe")
  else
    # --sort=name and the pinned mtime/owner keep the archive itself
    # reproducible, not just the binary inside it.
    tar --sort=name --owner=0 --group=0 --numeric-owner \
      --mtime='UTC 2020-01-01' -C "$work/$name" -czf "$out/$name.tar.gz" "$exe"
  fi
done

(cd "$out" && sha256sum specs_* > SHA256SUMS)
cat "$out/SHA256SUMS"
