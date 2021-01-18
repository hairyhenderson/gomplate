package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
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
	o, e, err := cmdTest(c,
		"-d", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{(datasource "config").foo.bar}}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdTest(c, "-d", "dir="+s.tmpDir.Path(),
		"-i", `{{ (datasource "dir" "config.json").foo.bar }}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdWithDir(c, s.tmpDir.Path(), "-d", "config=config.json",
		"-i", `{{ (ds "config").foo.bar }}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdWithDir(c, s.tmpDir.Path(), "-d", "config.json",
		"-i", `{{ (ds "config").foo.bar }}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdTest(c, "-i",
		`foo{{defineDatasource "config" `+"`"+s.tmpDir.Join("config.json")+"`"+`}}bar{{(datasource "config").foo.bar}}`)
	assertSuccess(c, o, e, err, "foobarbaz")

	o, e, err = cmdTest(c, "-i",
		`foo{{defineDatasource "config" `+"`"+s.tmpDir.Join("ajsonfile")+"?type=application/json`"+`}}bar{{(datasource "config").foo.bar}}`)
	assertSuccess(c, o, e, err, "foobarbaz")

	o, e, err = cmdTest(c, "-d", "config="+s.tmpDir.Join("config.yml"),
		"-i", `{{(datasource "config").foo.bar}}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdTest(c, "-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{(ds "config").foo}}`)
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c, "-c", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ .config.foo}}`)
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c, "-c", ".="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ .foo}} {{ (ds ".").foo }}`)
	assertSuccess(c, o, e, err, "bar bar")

	o, e, err = cmdTest(c, "-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ if (datasourceReachable "bogus") }}bogus!{{ end -}}
{{ if (datasourceReachable "config") -}}
{{ (ds "config").foo -}}
{{ end }}`)
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c, "-d", "csv="+s.tmpDir.Join("foo.csv"),
		"-i", `{{ index (index (ds "csv") 2) 1 }}`)
	assertSuccess(c, o, e, err, "foo\"\nbar")

	o, e, err = cmdTest(c, "-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ include "config" }}`)
	assertSuccess(c, o, e, err, "foo: bar\n")

	o, e, err = cmdTest(c, "-d", "dir="+s.tmpDir.Path()+"/",
		"-i", `{{ range (ds "dir") }}{{ . }} {{ end }}`)
	assertSuccess(c, o, e, err, "ajsonfile config.json config.yml config2.yml encrypted.json foo.csv sortorder test.env ")

	o, e, err = cmdWithEnv(c, []string{"-d", "enc=" + s.tmpDir.Join("encrypted.json"),
		"-i", `{{ (ds "enc").password }}`}, map[string]string{
		"EJSON_KEY": "553da5790efd7ddc0e4829b69069478eec9ddddb17b69eca9801da37445b62bf"})
	assertSuccess(c, o, e, err, "swordfish")

	o, e, err = cmdTest(c, "-d", "core="+s.tmpDir.Join("sortorder", "core.yaml"),
		"-f", s.tmpDir.Join("sortorder", "template"))
	assertSuccess(c, o, e, err, `aws_zones = {
	zonea = "true"
	zoneb = "false"
	zonec = "true"
	zoned = "true"
	zonee = "false"
	zonef = "false"
}
`)

	o, e, err = cmdTest(c, "-d", "envfile="+s.tmpDir.Join("test.env"),
		"-i", `{{ (ds "envfile") | data.ToJSONPretty "  " }}`)
	assertSuccess(c, o, e, err, `{
  "BAR": "another value, exports are ignored",
  "BAZ": "variable expansion: a regular unquoted value",
  "FOO": "a regular unquoted value",
  "FOO.BAR": "values can be double-quoted, and shell\nescapes are supported",
  "QUX": "single quotes ignore $variables"
}`)
}
