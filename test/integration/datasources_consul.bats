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

@test "Consul datasource works with hostname in URL" {
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  unset CONSUL_HTTP_ADDR
  gomplate -d consul=consul+http://localhost:8500/ -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Consul datasource works with consul+http scheme" {
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  unset CONSUL_HTTP_ADDR
  gomplate -d consul=consul+http:// -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}
