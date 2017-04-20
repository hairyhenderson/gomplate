#!/usr/bin/env bats

load helper

# function setup () {
# }

# function teardown () {
#   # rm -rf $tmpdir
# }

@test "errors with non-existant env var using .Env" {
  gomplate -i '{{.Env.FOO}}'
  [ "$status" -eq 2 ]
  [[ "${lines[0]}" == *"map has no entry for key"* ]]
}

@test "empty string with non-existant env var using getenv" {
  gomplate -i '{{getenv "FOO" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "" ]]
}

@test "default string with non-existant env var using getenv" {
  gomplate -i '{{getenv "FOO" "foo"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "foo" ]]
}

@test "existant env var using .Env" {
  gomplate -i '{{.Env.HOME}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "${HOME}" ]]
}

@test "existant env var using getenv" {
  gomplate -i '{{getenv "HOME"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "${HOME}" ]]
}
