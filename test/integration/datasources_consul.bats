#!/usr/bin/env bats

load helper

function setup () {
  tmpdir=$(mktemp -d)
}

function teardown () {
  rm -rf $tmpdir
  consul kv delete foo
}

@test "Testing consul" {
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  gomplate -d consul=consul:// -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}
