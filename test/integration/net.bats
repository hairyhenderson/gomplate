#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "'net.LookupIP'" {
  gomplate -i '{{ net.LookupIP "localhost" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "127.0.0.1" ]]
}
