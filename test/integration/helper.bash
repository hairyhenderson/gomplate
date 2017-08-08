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
}

function stop_mirror_svc () {
  wget -q http://127.0.0.1:8080/quit
}

function start_meta_svc () {
  bin/meta 3>/dev/null &
}

function stop_meta_svc () {
  wget -q http://127.0.0.1:8081/quit
}

function start_aws_svc () {
  bin/aws &
}

function stop_aws_svc () {
  wget -q http://127.0.0.1:8082/quit
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
