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
  echo '{"foo": {"bar": "baz"}}' > $tmpdir/config.json
  gomplate -d config=$tmpdir/config.json -i '{{(datasource "config").foo.bar}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "baz" ]]
}

@test "supports YAML datasource file" {
  echo -e 'foo:\n bar: baz' > $tmpdir/config.yml
  gomplate -d config=$tmpdir/config.yml -i '{{(datasource "config").foo.bar}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "baz" ]]
}

@test "ds alias" {
  echo 'foo: bar' > $tmpdir/config.yml
  gomplate -d config=$tmpdir/config.yml -i '{{(ds "config").foo}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}

@test "supports CSV datasource file" {
  echo -e 'A,B\nA1,B1\nA2,"foo""\nbar"\n' > $tmpdir/foo.csv
  gomplate -d csv=$tmpdir/foo.csv -i '{{ index (index (ds "csv") 2) 1 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "foo\"
bar" ]]
}

@test "'include' doesn't parse file" {
  echo 'foo: bar' > $tmpdir/config.yml
  gomplate -d config=$tmpdir/config.yml -i '{{include "config"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "foo: bar" ]]
}
