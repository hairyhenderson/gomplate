//+build integration

package integration

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
	. "gopkg.in/check.v1"
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
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{ tmpl.Exec "Nope" }}`,
	))
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: `template "Nope" not defined`})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{define "T1"}}hello world{{end}}{{ tmpl.Exec "T1" | strings.ToUpper }}`,
	))
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `HELLO WORLD`})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "in=stdin:///in.json",
		"-t", "toyaml="+s.tmpDir.Join("toyaml.tmpl"),
		"-i", `foo:
{{ tmpl.Exec "toyaml" .in | strings.Indent 2 }}`,
	), func(cmd *icmd.Cmd) {
		in := bytes.NewBufferString(`{"a":{"nested": "object"},"b":true}`)
		cmd.Stdin = in
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `foo:
  a:
    nested: object
  b: true
`})

	outDir := s.tmpDir.Join("out")
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		assert.NilError(c, err)
	}
	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "services="+s.tmpDir.Join("services.yaml"),
		"-i", `{{- define "config" }}{{ .config | data.ToJSONPretty " " }}{{ end }}
{{- range (ds "services").services -}}
{{- $outPath := path.Join .name "config.json" }}
{{- tmpl.Exec "config" . | file.Write $outPath }}
{{- end -}}`,
	), func(cmd *icmd.Cmd) {
		cmd.Dir = outDir
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	assert.Equal(c, "", result.Stdout())
	assert.Equal(c, "", result.Stderr())

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
