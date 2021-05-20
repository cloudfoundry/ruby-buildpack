#!/usr/bin/env bash

set -e
set -u
set -o pipefail

ROOTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly ROOTDIR

function main() {
  local src
  src="$(find "${ROOTDIR}/src/" -type d -depth 1)"

  for name in supply finalize; do
    GOOS=linux \
      go build \
        -mod vendor \
        -ldflags="-s -w" \
        -o "${ROOTDIR}/bin/${name}" \
          "${src}/${name}/cli"
  done
}

main "${@:-}"
