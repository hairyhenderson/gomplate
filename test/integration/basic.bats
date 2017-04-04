#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
  echo 'hi' > $tmpdir/one
  echo 'hello' > $tmpdir/two
}

function teardown () {
  rm -rf $tmpdir
}

@test "takes stdin by default" {
  gomplate_stdin "hello world"
  [ "$status" -eq 0 ]
  [[ "${output}" == "hello world" ]]
}

@test "takes stdin with --file -" {
  gomplate_stdin "hello world" --file -
  [ "$status" -eq 0 ]
  [[ "${output}" == "hello world" ]]
}

@test "writes to stdout with --out -" {
  gomplate_stdin "hello world" --out -
  [ "$status" -eq 0 ]
  [[ "${output}" == "hello world" ]]
}

@test "ignores stdin with --in" {
  gomplate_stdin "hello world" --in "hi"
  [ "$status" -eq 0 ]
  [[ "${output}" == "hi" ]]
}

@test "errors given more inputs than outputs" {
  skip
  gomplate -f $tmpdir/one -f $tmpdir/two -o $tmpdir/out
  [ "$status" -eq 2 ]
  [[ "${output}" == "still need to make this real..." ]]
}

@test "routes inputs to their proper outputs" {
  gomplate -f $tmpdir/one -f $tmpdir/two -o $tmpdir/one.out -o $tmpdir/two.out
  [ "$status" -eq 0 ]
  [[ "$(cat $tmpdir/one.out)" == "hi" ]]
  [[ "$(cat $tmpdir/two.out)" == "hello" ]]
}

@test "number of input files irrelevant given input string" {
  gomplate -i 'HELLO WORLD' -f foo -f bar
  [ "$status" -eq 0 ]
  [[ "${output}" == "HELLO WORLD" ]]
}
