//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type FileDatasourcesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&FileDatasourcesSuite{})

func (s *FileDatasourcesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"config.json": `{"foo": {"bar": "baz"}}`,
			"ajsonfile":   `{"foo": {"bar": "baz"}}`,
			"encrypted.json": `{
  "_public_key": "dfcf98785869cdfc4a59273bbdfe1bfcf6c44850a11ea9d84db21c89a802c057",
  "password": "EJ[1:Cb1AY94Dl76xwHHrnJyh+Y+fAeovijPlFQZXSAuvZBc=:oCGZM6lbeXXOl2ONSKfLQ0AgaltrTpNU:VjegqQPPkOK1hSylMAbmcfusQImfkHCWZw==]"
}`,
			"config.yml":  "foo:\n bar: baz\n",
			"config2.yml": "foo: bar\n",
			"foo.csv": `A,B
A1,B1
A2,"foo""
bar"
`,
			"test.env": `FOO=a regular unquoted value
export BAR=another value, exports are ignored

# comments are totally ignored, as are blank lines
FOO.BAR = "values can be double-quoted, and shell\nescapes are supported"

BAZ = "variable expansion: ${FOO}"
QUX='single quotes ignore $variables'
`,
			"multidoc.yaml": `---
# empty document
---
foo: bar
---
foo: baz
...
`,
		}),
		fs.WithDir("sortorder", fs.WithFiles(map[string]string{
			"template": `aws_zones = {
	{{- range $key, $value := (ds "core").cloud.aws.zones }}
	{{ $key }} = "{{ $value }}"
	{{- end }}
}
`,
			"core.yaml": `cloud:
  aws:
    zones:
      zonea: true
      zoneb: false
      zonec: true
      zoned: true
      zonee: false
      zonef: false
`,
		})),
	)
}

func (s *FileDatasourcesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *FileDatasourcesSuite) TestFileDatasources(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "dir="+s.tmpDir.Path(),
		"-i", `{{ (datasource "dir" "config.json").foo.bar }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "config=config.json",
		"-i", `{{ (ds "config").foo.bar }}`,
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "config.json",
		"-i", `{{ (ds "config").foo.bar }}`,
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-i", `foo{{defineDatasource "config" `+"`"+s.tmpDir.Join("config.json")+"`"+`}}bar{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foobarbaz"})

	result = icmd.RunCommand(GomplateBin,
		"-i", `foo{{defineDatasource "config" `+"`"+s.tmpDir.Join("ajsonfile")+"?type=application/json`"+`}}bar{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foobarbaz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config.yml"),
		"-i", `{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{(ds "config").foo}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-c", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ .config.foo}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-c", ".="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ .foo}} {{ (ds ".").foo }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar bar"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ if (datasourceReachable "bogus") }}bogus!{{ end -}}
{{ if (datasourceReachable "config") -}}
{{ (ds "config").foo -}}
{{ end }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "csv="+s.tmpDir.Join("foo.csv"),
		"-i", `{{ index (index (ds "csv") 2) 1 }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `foo"
bar`})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ include "config" }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `foo: bar`})

	result = icmd.RunCommand(GomplateBin,
		"-d", "dir="+s.tmpDir.Path()+"/",
		"-i", `{{ range (ds "dir") }}{{ . }} {{ end }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `config.json config.yml config2.yml encrypted.json foo.csv`})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "enc="+s.tmpDir.Join("encrypted.json"),
		"-i", `{{ (ds "enc").password }}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"EJSON_KEY=553da5790efd7ddc0e4829b69069478eec9ddddb17b69eca9801da37445b62bf",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "swordfish"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "core="+s.tmpDir.Join("sortorder", "core.yaml"),
		"-f", s.tmpDir.Join("sortorder", "template"),
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `aws_zones = {
	zonea = "true"
	zoneb = "false"
	zonec = "true"
	zoned = "true"
	zonee = "false"
	zonef = "false"
}
`})
	result = icmd.RunCommand(GomplateBin,
		"-d", "envfile="+s.tmpDir.Join("test.env"),
		"-i", `{{ (ds "envfile") | data.ToJSONPretty "  " }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `{
  "BAR": "another value, exports are ignored",
  "BAZ": "variable expansion: a regular unquoted value",
  "FOO": "a regular unquoted value",
  "FOO.BAR": "values can be double-quoted, and shell\nescapes are supported",
  "QUX": "single quotes ignore $variables"
}`})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d=multidoc.yaml",
		"-i", `{{ range $d := (include "multidoc" | data.YAMLStream) }}---{{"\n"}}{{ range $k, $v := $d }}{{ $k }}: {{ $v }}{{"\n"}}{{ end }}{{ end }}`,
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "---\n---\nfoo: bar\n---\nfoo: baz\n"})
}
