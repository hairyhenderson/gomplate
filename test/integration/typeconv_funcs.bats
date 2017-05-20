#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "'has' can handle sub-maps in nested maps" {
  gomplate -d config=$tmpdir/config.yml -i '{{ has ("foo:\n bar:\n  baz: qux" | yaml).foo.bar "baz"}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "true" ]]
}

@test "'toJSON' can handle nested maps" {
  gomplate -i '{{ "foo:\n bar:\n  baz: qux" | yaml | toJSON }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == '{"foo":{"bar":{"baz":"qux"}}}' ]]
}

@test "'toJSONPretty' can handle nested maps" {
  gomplate -i '{{ `{"foo":{"bar":{"baz":"qux"}}}` | json | toJSONPretty "   " }}
{{ toJSONPretty "" (`{"foo":{"bar":{"baz":"qux"}}}` | json) }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == '{
   "foo": {
      "bar": {
         "baz": "qux"
      }
   }
}
{
"foo": {
"bar": {
"baz": "qux"
}
}
}' ]]
}

@test "indent" {
  gomplate -i '{{ indent "   " "hello world" }}{{ "hello\nmultiline\nworld" | indent " " }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "   hello world hello
 multiline
 world" ]]
}

@test "join" {
  gomplate -i '{{ $a := `[1, 2, 3]` | jsonArray }}{{ join $a "-" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "1-2-3" ]]
}
