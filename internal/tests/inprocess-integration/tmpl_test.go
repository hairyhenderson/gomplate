package integration

import (
	"io/ioutil"
	"os"

	. "gopkg.in/check.v1"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

type TmplSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&TmplSuite{})

func (s *TmplSuite) SetUpTest(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-tmpltests",
		fs.WithFiles(map[string]string{
			"toyaml.tmpl": `{{ . | data.ToYAML }}{{"\n"}}`,
			"services.yaml": `services:
  - name: users
    config:
      replicas: 2
  - name: products
    config:
      replicas: 18
`,
		}),
	)
}

func (s *TmplSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *TmplSuite) TestInline(c *C) {
	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- tpl "{{ add .first .second }}" $nums }}`,
		"15")

	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- $othernums := dict "first" 18 "second" -8 }}
		{{- tmpl.Inline "T" "{{ add .first .second }}" $nums }}
		{{- template "T" $othernums }}`,
		"1510")
}

func (s *TmplSuite) TestExec(c *C) {
	_, _, err := cmdTest(c, "-i", `{{ tmpl.Exec "Nope" }}`)
	assert.ErrorContains(c, err, `template "Nope" not defined`)

	inOutTest(c, `{{define "T1"}}hello world{{end}}{{ tmpl.Exec "T1" | strings.ToUpper }}`, `HELLO WORLD`)

	o, e, err := cmdWithStdin(c, []string{
		"-c", "in=stdin:///in.json",
		"-t", "toyaml=" + s.tmpDir.Join("toyaml.tmpl"),
		"-i", `foo:
{{ tmpl.Exec "toyaml" .in | strings.Indent 2 }}`},
		`{"a":{"nested": "object"},"b":true}`)
	assertSuccess(c, o, e, err, `foo:
  a:
    nested: object
  b: true

`)

	outDir := s.tmpDir.Join("out")
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		assert.NilError(c, err)
	}
	o, e, err = cmdWithDir(c, outDir,
		"-d", "services="+s.tmpDir.Join("services.yaml"),
		"-i", `{{- define "config" }}{{ .config | data.ToJSONPretty " " }}{{ end }}
{{- range (ds "services").services -}}
{{- $outPath := path.Join .name "config.json" }}
{{- tmpl.Exec "config" . | file.Write $outPath }}
{{- end -}}`)
	assertSuccess(c, o, e, err, "")

	out, err := ioutil.ReadFile(s.tmpDir.Join("out", "users", "config.json"))
	assert.NilError(c, err)
	assert.Equal(c, `{
 "replicas": 2
}`, string(out))
	out, err = ioutil.ReadFile(s.tmpDir.Join("out", "products", "config.json"))
	assert.NilError(c, err)
	assert.Equal(c, `{
 "replicas": 18
}`, string(out))
}
