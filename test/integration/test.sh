#!/bin/bash
set -euo pipefail

source $(dirname $0)/helper.bash

function finish {
    stop_vault
}
trap finish EXIT

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_ROOT_TOKEN=00000000-1111-2222-3333-444455556666

start_vault 8200

bats $(dirname $0)
