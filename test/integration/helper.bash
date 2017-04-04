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
