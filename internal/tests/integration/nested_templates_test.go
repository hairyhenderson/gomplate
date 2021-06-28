package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/fs"
)

func setupNestedTemplatesTest(t *testing.T) (*fs.Dir, *httptest.Server) {
	one := `{{ . }}`
	two := `{{ range $n := (seq 2) }}{{ $n }}: {{ $ }} {{ end }}`

	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("hello.t", `Hello {{ . }}!`),
		fs.WithDir("templates",
			fs.WithFile("one.t", one),
			fs.WithFile("two.t", two),
		),
	)
	t.Logf("tmpDir created at %s", tmpDir.Path())
	t.Cleanup(tmpDir.Remove)

	h := http.NewServeMux()
	h.HandleFunc("/one.t", typeHandler("text/plain", one))
	h.HandleFunc("/two.t", typeHandler("text/plain", two))

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	return tmpDir, srv
}

func TestNestedTemplates(t *testing.T) {
	tmpDir, _ := setupNestedTemplatesTest(t)

	o, e, err := cmd(t,
		"-t", "hello="+tmpDir.Join("hello.t"),
		"-i", `{{ template "hello" "World"}}`,
	).run()
	assertSuccess(t, o, e, err, "Hello World!")

	o, e, err = cmd(t, "-t", "hello.t",
		"-i", `{{ template "hello.t" "World"}}`).
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "Hello World!")

	o, e, err = cmd(t, "-t", "templates/",
		"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`).
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "one\n1: two 2: two ")

	o, e, err = cmd(t, "-t", fmt.Sprintf("templates=file://%s/templates/", tmpDir.Path()),
		"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`).run()
	assertSuccess(t, o, e, err, `one
1: two 2: two `)

	// 	o, e, err = cmd(t,
	// 		"-t", fmt.Sprintf("one=%s/one.t", srv.URL),
	// 		"-t", fmt.Sprintf("two=%s/two.t", srv.URL),
	// 		"-i", `{{ template "one"}} {{ template "two"}}`).run()
	// 	assertSuccess(t, o, e, err, `one 1: two 2: two`)
}
