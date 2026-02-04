#!/bin/bash

set -e
set -u
set -o pipefail

function main() {
  echo "-----> [DEBUG] install_go.sh called, CF_STACK=${CF_STACK:-NOT_SET}"
  
  # Allow cflinuxfs3, cflinuxfs4, and cflinuxfs5
  if [[ "${CF_STACK:-}" != "cflinuxfs3" && "${CF_STACK:-}" != "cflinuxfs4" && "${CF_STACK:-}" != "cflinuxfs5" ]]; then
    echo "       **ERROR** Unsupported stack"
    echo "                 See https://docs.cloudfoundry.org/devguide/deploy-apps/stacks.html for more info"
    exit 1
  fi

  local version expected_sha dir
  if [[ "${CF_STACK}" == "cflinuxfs5" ]]; then
    version="1.23.12"
    expected_sha="ff5ab3b02743246897bfc65d7c205e58eabba313d6a07587bec97d812fb3c6c0"
  else
    version="1.22.5"
    expected_sha="ddb12ede43eef214c7d4376761bd5ba6297d5fa7a06d5635ea3e7a276b3db730"
  fi
  dir="/tmp/go${version}"

  mkdir -p "${dir}"

  echo "-----> [DEBUG] install_go.sh: CF_STACK=${CF_STACK:-not set}, Go version=${version}"

  if [[ ! -f "${dir}/bin/go" ]]; then
    local url
    # Use stack-based Go dependency URL
    if [[ "${CF_STACK}" == "cflinuxfs5" ]]; then
      echo "-----> [DEBUG] Using GitHub URL for cflinuxfs5"
      url="https://github.com/ivanovac/go-buildpack/releases/download/v1.11.00-beta-cflinuxfs5/go_${version}_linux_x64_cflinuxfs5_${expected_sha:0:8}.tgz"
    else
      echo "-----> [DEBUG] Using buildpacks.cloudfoundry.org URL for ${CF_STACK}"
      url="https://buildpacks.cloudfoundry.org/dependencies/go/go_${version}_linux_x64_${CF_STACK}_${expected_sha:0:8}.tgz"
    fi

    echo "-----> Download go ${version}"
    echo "-----> [DEBUG] URL: ${url}"
    curl "${url}" \
      --silent \
      --location \
      --retry 15 \
      --retry-delay 2 \
      --output "/tmp/go.tgz"

    local sha
    sha="$(shasum -a 256 /tmp/go.tgz | cut -d ' ' -f 1)"

    if [[ "${sha}" != "${expected_sha}" ]]; then
      echo "       **ERROR** SHA256 mismatch: got ${sha}, expected ${expected_sha}"
      exit 1
    fi

    tar xzf "/tmp/go.tgz" -C "${dir}"
    rm "/tmp/go.tgz"
  fi

  if [[ ! -f "${dir}/bin/go" ]]; then
    echo "       **ERROR** Could not download go"
    exit 1
  fi

  GoInstallDir="${dir}"
  export GoInstallDir
}

main "${@:-}"
