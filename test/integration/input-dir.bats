#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
  mkdir -p $tmpdir/in/inner
  echo -n "{{ (datasource \"config\").one }}" > $tmpdir/in/eins.txt
  echo -n "{{ (datasource \"config\").two }}" > $tmpdir/in/inner/deux.txt

  cat <<"EOT"  > $tmpdir/config.yml
one: eins
two: deux
EOT
}

function teardown () {
  # rm -rf $tmpdir
  echo
}

@test "takes --input-dir and produces proper output files" {
  rm -rf $tmpdir/out || true
  gomplate --input-dir $tmpdir/in --output-dir $tmpdir/out -d config=$tmpdir/config.yml
  [ "$status" -eq 0 ]
  [[ "$(ls $tmpdir/out | wc -l)" == 2 ]]
  [[ "$(ls $tmpdir/out/inner | wc -l)" == 1 ]]
  [[ "$(cat $tmpdir/out/eins.txt)" == "eins" ]]
  [[ "$(cat $tmpdir/out/inner/deux.txt)" == "deux" ]]
}

@test "test . as default --output-dir param" {
  rm -rf $tmpdir/out_dot || true
  mkdir -p $tmpdir/out_dot
  g=$(pwd)/bin/gomplate
  cd $tmpdir/out_dot
  run $g --input-dir $tmpdir/in -d config=$tmpdir/config.yml
  [ "$?" -eq 0 ]
  [[ "$(ls | wc -l)" == 2 ]]
  [[ "$(ls inner | wc -l)" == 1 ]]
  [[ "$(cat eins.txt)" == "eins" ]]
  [[ "$(cat inner/deux.txt)" == "deux" ]]
}

@test "errors given --output-dir but no --input-dir" {
  gomplate --output-dir "."
  [ "$status" -eq 1 ]
  [[ "${output}" == "Error: --input-dir must be set when --output-dir is set"* ]]
}

@test "errors given both --input-dir and --in" {
  gomplate --input-dir "." --in "param"
  [ "$status" -eq 1 ]
  [[ "${output}" == "Error: --input-dir can not be used together with --in or --file"* ]]
}

@test "errors given both --input-dir and --file" {
  gomplate --input-dir "." --file input.txt
  [ "$status" -eq 1 ]
  [[ "${output}" == "Error: --input-dir can not be used together with --in or --file"* ]]
}

@test "errors given both --output-dir and --out" {
  gomplate --input-dir "." --output-dir /tmp --out out
  [ "$status" -eq 1 ]
  [[ "${output}" == "Error: --output-dir can not be used together with --out"* ]]
}

@test "errors with filename when using input dir and bad input file" {
  rm -rf $tmpdir/out || true
  echo -n "{{end}}" > $tmpdir/in/bad.tmpl
  gomplate --input-dir $tmpdir/in --output-dir $tmpdir/out -d config=$tmpdir/config.yml
  [ "$status" -eq 1 ]
  [[ "${output}" == "Error: template: $tmpdir/in/bad.tmpl:1: unexpected {{end}}"* ]]
}