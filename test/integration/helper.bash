#!/bin/bash

function gomplate () {
  run bin/gomplate "$@"

  # Some debug information to make life easier. bats will only print it if the
  # test failed, in which case the output is useful.
  echo "gomplate $@ (status=$status):" >&2
  echo "$output" >&2
}

function gomplate_stdin () {
  run __gomplate_stdin "$@"

  # Some debug information to make life easier. bats will only print it if the
  # test failed, in which case the output is useful.
  in=$1
  shift
  echo "echo \"$in\" | gomplate $@ (status=$status):" >&2
  echo "$output" >&2
}

function __gomplate_stdin () {
  in=$1
  shift 1
  echo "$in" | bin/gomplate "$@"
}

function start_mirror_svc () {
  bin/mirror &
  wait_for_url http://127.0.0.1:8080/
}

function stop_mirror_svc () {
  wget -q -O - http://127.0.0.1:8080/quit
}

function start_meta_svc () {
  bin/meta &> /tmp/meta.log &
  wait_for_url http://127.0.0.1:8081/
}

function stop_meta_svc () {
  wget -q -O - http://127.0.0.1:8081/quit
}

function start_aws_svc () {
  bin/aws &
  wait_for_url http://127.0.0.1:8082/
}

function stop_aws_svc () {
  wget -q -O - http://127.0.0.1:8082/quit
}

function wait_for_url () {
  url=$1
  for i in {0..10}; do
    curl -o /dev/null -s -f $url && break || sleep 1
  done
}

function start_consul () {
  port=$1
  if [ -z $port ]; then
    port=8500
  fi
  PID_FILE=/tmp/gomplate-test-consul.pid
  rm -f $PID_FILE || true
  export CONSUL_ROOT_TOKEN=00000000-1111-2222-3333-444455556666
  echo "{\"acl_datacenter\": \"dc1\", \"acl_master_token\": \"${CONSUL_ROOT_TOKEN}\"}" >> /tmp/gomplate-test-consul.json
  consul agent -dev -config-file=/tmp/gomplate-test-consul.json -log-level=err -http-port=$port -pid-file=$PID_FILE >/dev/null &
  wait_for_url http://127.0.0.1:$port/v1/status/leader
}

function stop_consul () {
  PID_FILE=/tmp/gomplate-test-consul.pid
  kill $(cat $PID_FILE) &>/dev/null
  rm /tmp/gomplate-test-consul.json
}

function start_vault () {
  port=$1
  PID_FILE=/tmp/gomplate-test-vault.pid
  export VAULT_ROOT_TOKEN=00000000-1111-2222-3333-444455556666

  # back up any existing token so it doesn't get overridden
  if [ -f ~/.vault-token ]; then
    cp ~/.vault-token ~/.vault-token.bak
  fi

  vault server -dev -dev-root-token-id=${VAULT_ROOT_TOKEN} -log-level=err >&/dev/null &
  echo $! > $PID_FILE
  wait_for_url http://127.0.0.1:$port/sys/health
}

function stop_vault () {
  PID_FILE=/tmp/gomplate-test-vault.pid  
  kill $(cat $PID_FILE) &>/dev/null

  # restore old token if it was backed up
  if [ -f ~/.vault-token.bak ]; then
    mv ~/.vault-token.bak ~/.vault-token
  fi
}
