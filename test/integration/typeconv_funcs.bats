#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "'has' can handle sub-maps in nested maps" {
  gomplate -d config=$tmpdir/config.yml -i '{{ has ("foo:\n bar:\n  baz: qux" | yaml).foo.bar "baz"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "true" ]]
}
