#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "supports BoltDB datasource file" {
  cp test/integration/config.db $tmpdir/config.db
  gomplate -d config=boltdb://$tmpdir/config.db#Bucket1 -i '{{(datasource "config" "foo")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar" ]]
}

@test "supports multi-bucket BoltDB datasource file" {
  cp test/integration/config.db $tmpdir/config.db
  gomplate -d config=boltdb://$tmpdir/config.db#Bucket1 -d config2=boltdb://$tmpdir/config.db#Bucket2 -i '{{(datasource "config" "foo")}}-{{(datasource "config2" "foobar")}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "bar-baz" ]]
}
