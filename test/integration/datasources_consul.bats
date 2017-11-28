#!/usr/bin/env bats

load helper

function setup () {
  start_consul 8501
  export CONSUL_HTTP_ADDR=http://127.0.0.1:8501
}

function teardown () {
  export CONSUL_HTTP_ADDR=http://127.0.0.1:8501
  consul kv delete foo
  vault unmount consul
  stop_consul
}

@test "Testing consul" {
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  gomplate -d consul=consul:// -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}

@test "Consul datasource works with MIME override" {
  consul kv put foo "{\"desc\":$BATS_TEST_DESCRIPTION}"
  gomplate -d consul=consul://?type=application/json -i '{{(datasource "consul" "foo").desc}}'
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

@test "Consul datasource works with Vault auth" {
  vault mount consul
  vault write consul/config/access address=127.0.0.1:8501 token=${CONSUL_ROOT_TOKEN}
  POLICY='key "" { policy = "read" }'
  vault write consul/roles/readonly policy=`echo $POLICY | base64`
  consul kv put foo "$BATS_TEST_DESCRIPTION"
  CONSUL_VAULT_ROLE=readonly gomplate -d consul=consul:// -i '{{(datasource "consul" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "$BATS_TEST_DESCRIPTION" ]]
}
