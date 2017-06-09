#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir
}

@test "'base64.Encode'" {
  gomplate -i '{{ "foo" | base64.Encode }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "Zm9v" ]]
}

@test "'base64.Decode'" {
  gomplate -i '{{ "Zm9v" | base64.Decode }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "foo" ]]
}