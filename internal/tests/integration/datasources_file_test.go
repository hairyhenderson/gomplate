package integration

import (
	"testing"

	"gotest.tools/v3/fs"
)

func setupDatasourcesFileTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
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

	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestDatasourcess_File(t *testing.T) {
	tmpDir := setupDatasourcesFileTest(t)

	o, e, err := cmd(t,
		"-d", "config="+tmpDir.Join("config.json"),
		"-i", `{{(datasource "config").foo.bar}}`).run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-d", "dir="+tmpDir.Path(),
		"-i", `{{ (datasource "dir" "config.json").foo.bar }}`).run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-d", "config=config.json",
		"-i", `{{ (ds "config").foo.bar }}`).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-d", "config.json",
		"-i", `{{ (ds "config").foo.bar }}`).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-i",
		`foo{{defineDatasource "config" `+"`"+tmpDir.Join("config.json")+"`"+`}}bar{{(datasource "config").foo.bar}}`).
		run()
	assertSuccess(t, o, e, err, "foobarbaz")

	o, e, err = cmd(t, "-i",
		`foo{{defineDatasource "config" `+"`"+tmpDir.Join("ajsonfile")+"?type=application/json`"+`}}bar{{(datasource "config").foo.bar}}`).
		run()
	assertSuccess(t, o, e, err, "foobarbaz")

	o, e, err = cmd(t, "-d", "config="+tmpDir.Join("config.yml"),
		"-i", `{{(datasource "config").foo.bar}}`).run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-d", "config="+tmpDir.Join("config2.yml"),
		"-i", `{{(ds "config").foo}}`).run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t, "-c", "config="+tmpDir.Join("config2.yml"),
		"-i", `{{ .config.foo}}`).run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t, "-c", ".="+tmpDir.Join("config2.yml"),
		"-i", `{{ .foo}} {{ (ds ".").foo }}`).run()
	assertSuccess(t, o, e, err, "bar bar")

	o, e, err = cmd(t, "-d", "config="+tmpDir.Join("config2.yml"),
		"-i", `{{ if (datasourceReachable "bogus") }}bogus!{{ end -}}
{{ if (datasourceReachable "config") -}}
{{ (ds "config").foo -}}
{{ end }}`).run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t, "-d", "csv="+tmpDir.Join("foo.csv"),
		"-i", `{{ index (index (ds "csv") 2) 1 }}`).run()
	assertSuccess(t, o, e, err, "foo\"\nbar")

	o, e, err = cmd(t, "-d", "config="+tmpDir.Join("config2.yml"),
		"-i", `{{ include "config" }}`).run()
	assertSuccess(t, o, e, err, "foo: bar\n")

	o, e, err = cmd(t, "-d", "dir="+tmpDir.Path()+"/",
		"-i", `{{ range (ds "dir") }}{{ . }} {{ end }}`).run()
	assertSuccess(t, o, e, err, "ajsonfile config.json config.yml config2.yml encrypted.json foo.csv sortorder test.env ")

	o, e, err = cmd(t, "-d", "enc="+tmpDir.Join("encrypted.json"),
		"-i", `{{ (ds "enc").password }}`).
		withEnv("EJSON_KEY", "553da5790efd7ddc0e4829b69069478eec9ddddb17b69eca9801da37445b62bf").
		run()
	assertSuccess(t, o, e, err, "swordfish")

	o, e, err = cmd(t, "-d", "core="+tmpDir.Join("sortorder", "core.yaml"),
		"-f", tmpDir.Join("sortorder", "template")).run()
	assertSuccess(t, o, e, err, `aws_zones = {
	zonea = "true"
	zoneb = "false"
	zonec = "true"
	zoned = "true"
	zonee = "false"
	zonef = "false"
}
`)

	o, e, err = cmd(t, "-d", "envfile="+tmpDir.Join("test.env"),
		"-i", `{{ (ds "envfile") | data.ToJSONPretty "  " }}`).run()
	assertSuccess(t, o, e, err, `{
  "BAR": "another value, exports are ignored",
  "BAZ": "variable expansion: a regular unquoted value",
  "FOO": "a regular unquoted value",
  "FOO.BAR": "values can be double-quoted, and shell\nescapes are supported",
  "QUX": "single quotes ignore $variables"
}`)
}
