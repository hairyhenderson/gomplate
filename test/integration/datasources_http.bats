#!/usr/bin/env bats

load helper

@test "HTTP datasource with headers" {
  gomplate \
    -d foo=http://httpbin.org/get \
    -H foo=Foo:bar \
    -i '{{ (datasource "foo").headers.Foo }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}
