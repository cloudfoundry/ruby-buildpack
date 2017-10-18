#!/usr/bin/env bash
set -euo pipefail

export ROOT=`dirname $(readlink -f ${BASH_SOURCE%/*})`
if [ ! -f $ROOT/.bin/ginkgo ]; then
  echo "Installing ginkgo"
  (cd $ROOT/src/ruby/vendor/github.com/onsi/ginkgo/ginkgo/ && go install)
fi

cd $ROOT/src/ruby/
ginkgo -r -skipPackage=brats,integration
