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

@test "join" {
  gomplate -i '{{ $a := `[1, 2, 3]` | jsonArray }}{{ join $a "-" }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "1-2-3" ]]
}

@test "'csv'" {
  gomplate -i '{{ $c := `lang,keywords
C,32
Go,25
COBOL,357` | csv -}}
{{ index (index $c 0) 1 }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "keywords" ]]
}

@test "'csvByRow' with default settings" {
  gomplate -i '{{ $c := `lang,keywords
C,32
Go,25
COBOL,357` | csvByRow }}{{ range $c }}{{ .lang }} has {{ .keywords }} keywords.
{{end}}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords." ]]
}

@test "'csvByColumn' (tab-separated)" {
  gomplate -i '{{ $c := `lang	keywords
C	32
Go	25
COBOL	357` | csvByColumn "\t" -}}
Languages are: {{ join $c.lang " and " }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "Languages are: C and Go and COBOL" ]]
}

@test "'toml'" {
  gomplate -i '{{ $t := `# comment
foo = "bar"

[baz]
qux = "quux"` | toml -}}
{{ $t.baz.qux }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "quux" ]]
}

@test "'toTOML'" {
  gomplate -i '{{ "foo:\n bar:\n  baz: qux" | yaml | toTOML }}'
  [ "$status" -eq 0 ]
  [[ "${output}" == "[foo]
  [foo.bar]
    baz = \"qux\"" ]]
}