package integration

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupTmplTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-tmpltests",
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
		fs.WithDir("a",
			fs.WithFiles(map[string]string{
				"pathtest.tpl": "{{ tmpl.Path }}\n{{ template `nested` }}",
				"a.tpl":        "{{ tmpl.PathDir }}",
				"b.tpl":        "{{ tmpl.Path }}",
			}),
		),
	)
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestTmpl_Inline(t *testing.T) {
	inOutTest(t, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- tpl "{{ add .first .second }}" $nums }}`,
		"15")

	inOutTest(t, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- $othernums := dict "first" 18 "second" -8 }}
		{{- tmpl.Inline "T" "{{ add .first .second }}" $nums }}
		{{- template "T" $othernums }}`,
		"1510")
}

func TestTmpl_Exec(t *testing.T) {
	tmpDir := setupTmplTest(t)

	_, _, err := cmd(t, "-i", `{{ tmpl.Exec "Nope" }}`).run()
	assert.ErrorContains(t, err, "Nope")
	assert.ErrorContains(t, err, " not defined")

	inOutTest(t, `{{define "T1"}}hello world{{end}}{{ tmpl.Exec "T1" | strings.ToUpper }}`, `HELLO WORLD`)

	o, e, err := cmd(t,
		"-c", "in=stdin:///in.json",
		"-t", "toyaml="+tmpDir.Join("toyaml.tmpl"),
		"-i", `foo:
{{ tmpl.Exec "toyaml" .in | strings.Indent 2 }}`).
		withStdin(`{"a":{"nested": "object"},"b":true}`).run()
	assertSuccess(t, o, e, err, `foo:
  a:
    nested: object
  b: true

`)

	outDir := tmpDir.Join("out")
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		assert.NilError(t, err)
	}
	o, e, err = cmd(t,
		"-d", "services="+tmpDir.Join("services.yaml"),
		"-i", `{{- define "config" }}{{ .config | data.ToJSONPretty " " }}{{ end }}
{{- range (ds "services").services -}}
{{- $outPath := path.Join .name "config.json" }}
{{- tmpl.Exec "config" . | file.Write $outPath }}
{{- end -}}`).
		withDir(outDir).run()
	assertSuccess(t, o, e, err, "")

	out, err := os.ReadFile(tmpDir.Join("out", "users", "config.json"))
	assert.NilError(t, err)
	assert.Equal(t, `{
 "replicas": 2
}`, string(out))
	out, err = os.ReadFile(tmpDir.Join("out", "products", "config.json"))
	assert.NilError(t, err)
	assert.Equal(t, `{
 "replicas": 18
}`, string(out))
}

func TestTmpl_Path(t *testing.T) {
	tmpDir := setupTmplTest(t)

	o, e, err := cmd(t,
		"-f", "a/pathtest.tpl",
		"-t", "nested=a/a.tpl",
	).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "a/pathtest.tpl\na")

	o, e, err = cmd(t,
		"-f", "a/a.tpl",
		"-f", "a/b.tpl",
		"-o", "-",
		"-o", "-",
	).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "aa/b.tpl")
}
