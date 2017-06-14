#!/usr/bin/env bats

load helper

tmpdir=$(mktemp -u)

function setup () {
  mkdir -p $tmpdir
}

function teardown () {
  rm -rf $tmpdir || true
}

@test "'strings.Indent'" {
  gomplate -i '{{ strings.Indent "   " "hello world" }}
{{ "hello\nmultiline\nworld" | indent 2 "-" }}
{{ "foo\nbar" | strings.Indent 2 }}
    {{"hello\nworld" | strings.Indent 5 | strings.TrimSpace }}
'
  [ "$status" -eq 0 ]
  [[ "${output}" == "   hello world
--hello
--multiline
--world
  foo
  bar
    hello
     world" ]]
}
