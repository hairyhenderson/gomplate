#!/usr/bin/env bats

load helper

function setup() {
  start_mirror_svc
}

function teardown() {
  stop_mirror_svc
}

@test "HTTP datasource with headers" {
  gomplate \
    -d foo=http://127.0.0.1:8080/ \
    -H foo=Foo:bar \
    -i '{{ index (datasource "foo").headers.Foo 0 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}
