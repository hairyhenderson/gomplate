#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "'regexp.Replace'" {
  gomplate -i '{{ "1.2.3-59" | regexp.Replace `-([0-9]*)` `.$1` }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "1.2.3.59" ]]
}
