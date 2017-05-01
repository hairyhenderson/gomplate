#!/bin/bash
set -euo pipefail

# This is useful for killing vault after the script exits, but causes the CircleCI
# build to fail, so... ¯\_(ツ)_/¯
# trap "exit" INT TERM
# trap "kill 0" EXIT

# TODO: export these in a bats helper, as well as only launch vault in a vault helper
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_ROOT_TOKEN=00000000-1111-2222-3333-444455556666

# fire up vault in dev mode for the vault tests
vault server -dev -dev-root-token-id=${VAULT_ROOT_TOKEN} -log-level=err >&/dev/null &

bats $(dirname $0)
