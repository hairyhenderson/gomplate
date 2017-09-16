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

@test "reports version" {
  gomplate -v
  [ "$status" -eq 0 ]
  [[ "${output}" == "gomplate version "* ]]
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
  gomplate -f $tmpdir/one -f $tmpdir/two -o $tmpdir/out
  [ "$status" -eq 1 ]
  [[ "${lines[0]}" == "Error: Must provide same number of --out (1) as --file (2) options" ]]
}

@test "routes inputs to their proper outputs" {
  gomplate -f $tmpdir/one -f $tmpdir/two -o $tmpdir/one.out -o $tmpdir/two.out
  [ "$status" -eq 0 ]
  [[ "$(cat $tmpdir/one.out)" == "hi" ]]
  [[ "$(cat $tmpdir/two.out)" == "hello" ]]
}

@test "can't mix --in and --file" {
  gomplate -i 'HELLO WORLD' -f -
  [ "$status" -eq 1 ]
  [[ "${lines[0]}" == "Error: --in and --file may not be used together" ]]
}

@test "delimiters can be changed through opts" {
  gomplate --left-delim "((" --right-delim "))" -i '((print "hi"))'
  [ "$status" -eq 0 ]
  [[ "${output}" == "hi" ]]
}

@test "delimiters can be changed through envvars" {
  GOMPLATE_LEFT_DELIM="<<" GOMPLATE_RIGHT_DELIM=">>" gomplate -i '<<print "hi">>'
  [ "$status" -eq 0 ]
  [[ "${output}" == "hi" ]]
}

@test "unknown argument results in error" {
  gomplate -in flibbit
  [ "$status" -eq 1 ]
  [ "${lines[0]}" = 'Error: unknown command "flibbit" for "gomplate"' ]
}
