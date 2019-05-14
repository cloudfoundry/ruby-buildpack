#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source .envrc
./scripts/install_tools.sh

GINKGO_NODES=${GINKGO_NODES:-3}
GINKGO_ATTEMPTS=${GINKGO_ATTEMPTS:-2}
export CF_STACK=${CF_STACK:-cflinuxfs3}
DISK_LIMIT_ARG=""
MEM_LIMIT_ARG=""

if [ -n "${CF_BRATS_DISK_QUOTA+x}" ]; then
  DISK_LIMIT_ARG="-disk=$CF_BRATS_DISK_QUOTA"
fi

if [ -n "${CF_BRATS_MEM_QUOTA+x}" ]; then
  MEM_LIMIT_ARG="-memory=$CF_BRATS_MEM_QUOTA"
fi

cd src/*/brats

echo "Run Buildpack Runtime Acceptance Tests"
ginkgo -r -mod=vendor --flakeAttempts=$GINKGO_ATTEMPTS -nodes $GINKGO_NODES -- $DISK_LIMIT_ARG $MEM_LIMIT_ARG
