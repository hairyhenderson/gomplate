package integration

import (
	. "gopkg.in/check.v1"
)

type TypeconvSuite struct{}

var _ = Suite(&TypeconvSuite{})

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

func (s *TypeconvSuite) TestTypeconvFuncs(c *C) {
	inOutTest(c, `{{ has ("`+testYAML+`" | yaml).foo.bar "baz"}}`,
		"true")
}

func (s *TypeconvSuite) TestJSON(c *C) {
	inOutTest(c, `{{ "`+testYAML+`" | yaml | toJSON }}`, testJSON)

	inOutTest(c, `{{ `+"`"+testJSON+"`"+` | json | toJSONPretty "   " }}
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

func (s *TypeconvSuite) TestJoin(c *C) {
	inOutTest(c, `{{ $a := "[1, 2, 3]" | jsonArray }}{{ join $a "-" }}`,
		"1-2-3")
}

func (s *TypeconvSuite) TestCSV(c *C) {
	inOutTest(c, `{{ $c := `+"`"+testCsv+"`"+` | csv -}}
{{ index (index $c 0) 1 }}`,
		"keywords")

	inOutTest(c, `{{ $c := `+"`"+testCsv+"`"+` | csvByRow -}}
{{ range $c }}{{ .lang }} has {{ .keywords }} keywords.
{{end}}`,
		`C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
`)

	inOutTest(c, `{{ $c := `+"`"+testTsv+"`"+` | csvByColumn "\t" -}}
Languages are: {{ join $c.lang " and " }}`,
		"Languages are: C and Go and COBOL")
}

func (s *TypeconvSuite) TestTOML(c *C) {
	inOutTest(c, `{{ $t := `+"`"+`# comment
foo = "bar"

[baz]
qux = "quux"`+"`"+` | toml -}}
{{ $t.baz.qux }}`, "quux")

	inOutTest(c, `{{ "foo:\n bar:\n  baz: qux" | yaml | toTOML }}`,
		`[foo]
  [foo.bar]
    baz = "qux"
`)
}

func (s *TypeconvSuite) TestDict(c *C) {
	inOutTest(c, `{{ $d := dict true false "foo" "bar" }}{{ data.ToJSON $d }}`,
		`{"foo":"bar","true":false}`)
}
