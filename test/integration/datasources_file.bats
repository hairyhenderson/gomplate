#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "supports json datasource file" {
  echo '{"foo": "bar"}' > $tmpdir/config.json
  gomplate -d config=$tmpdir/config.json -i '{{(datasource "config").foo}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}

@test "supports YAML datasource file" {
  echo 'foo: bar' > $tmpdir/config.yml
  gomplate -d config=$tmpdir/config.yml -i '{{(datasource "config").foo}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}

@test "ds alias" {
  echo 'foo: bar' > $tmpdir/config.yml
  gomplate -d config=$tmpdir/config.yml -i '{{(ds "config").foo}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}
