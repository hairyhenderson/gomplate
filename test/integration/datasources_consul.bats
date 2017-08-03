#!/usr/bin/env bats

load helper

function setup () {
  start_consul 8501
  export CONSUL_HTTP_ADDR=http://127.0.0.1:8501
}

function teardown () {
  consul kv delete foo
  unset CONSUL_HTTP_ADDR
  stop_consul
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
  gomplate -d consul=consul://127.0.0.1:8501/ -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Consul datasource works with consul+http scheme" {
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  unset CONSUL_HTTP_ADDR
  gomplate -d consul=consul+http://127.0.0.1:8501/ -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}
