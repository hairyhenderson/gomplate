package integration

import (
	"testing"

	"gotest.tools/v3/fs"
)

func setupNestedTemplatesTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("hello.t", `Hello {{ . }}!`),
		fs.WithDir("templates",
			fs.WithFile("one.t", `{{ . }}`),
			fs.WithFile("two.t", `{{ range $n := (seq 2) }}{{ $n }}: {{ $ }} {{ end }}`),
		),
	)
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestNestedTemplates(t *testing.T) {
	tmpDir := setupNestedTemplatesTest(t)

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
}
