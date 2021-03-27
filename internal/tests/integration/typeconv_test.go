package integration

import (
	"fmt"
	"testing"
)

const (
	testYAML = "foo:\\n bar:\\n  baz: qux"
	testJSON = `{"foo":{"bar":{"baz":"qux"}}}`
	testCsv  = `lang,keywords
C,32
Go,25
COBOL,357`
	testTsv = `lang	keywords
C	32
Go	25
COBOL	357`
)

func TestTypeconv_TypeconvFuncs(t *testing.T) {
	inOutTest(t, `{{ has ("`+testYAML+`" | yaml).foo.bar "baz"}}`,
		"true")
}

func TestTypeconv_JSON(t *testing.T) {
	inOutTest(t, `{{ "`+testYAML+`" | yaml | toJSON }}`, testJSON)

	inOutTest(t, `{{ `+"`"+testJSON+"`"+` | json | toJSONPretty "   " }}
{{ toJSONPretty "" (`+"`"+testJSON+"`"+` | json) }}`,
		`{
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
}`)
}

func TestTypeconv_Join(t *testing.T) {
	inOutTest(t, `{{ $a := "[1, 2, 3]" | jsonArray }}{{ join $a "-" }}`,
		"1-2-3")
}

func TestTypeconv_CSV(t *testing.T) {
	inOutTest(t, fmt.Sprintf(`{{ $c := %q | csv -}}
{{ index (index $c 0) 1 }}`, testCsv),
		"keywords")

	inOutTest(t, fmt.Sprintf(`{{ $c := %q | csvByRow -}}
{{ range $c }}{{ .lang }} has {{ .keywords }} keywords.
{{end}}`, testCsv),
		`C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
`)

	inOutTest(t, fmt.Sprintf(`{{ $c := %q | csvByColumn "\t" -}}
Languages are: {{ join $c.lang " and " }}`, testTsv),
		"Languages are: C and Go and COBOL")
}

func TestTypeconv_TOML(t *testing.T) {
	tomlIn := `# comment
foo = "bar"

[baz]
qux = "quux"`
	inOutTest(t, fmt.Sprintf(`{{ $t := %q | toml }}{{ $t.baz.qux }}`, tomlIn), "quux")

	inOutTest(t, `{{ "foo:\n bar:\n  baz: qux" | yaml | toTOML }}`,
		`[foo]
  [foo.bar]
    baz = "qux"
`)
}

func TestTypeconv_Dict(t *testing.T) {
	inOutTest(t, `{{ $d := dict true false "foo" "bar" }}{{ data.ToJSON $d }}`,
		`{"foo":"bar","true":false}`)
}
