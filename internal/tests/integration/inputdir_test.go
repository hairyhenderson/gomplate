package integration

import (
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	tassert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupInputDirTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("config.yml", "one: eins\ntwo: deux\n"),
		fs.WithFile("filemap.json", `{"eins.txt":"uno","deux.txt":"dos","drei.sh":"tres","vier.txt":"quatro"}`),
		fs.WithFile("out.t", `{{- /* .in may contain a directory name - we want to preserve that */ -}}
{{ $f := filepath.Base .in -}}
out/{{ .in | strings.ReplaceAll $f (index .filemap $f) }}.out
`),
		fs.WithDir("in",
			fs.WithFile("eins.txt", `{{ (ds "config").one }}`, fs.WithMode(0644)),
			fs.WithDir("inner",
				fs.WithFile("deux.txt", `{{ (ds "config").two }}`, fs.WithMode(0444)),
			),
			fs.WithFile("drei.sh", `#!/bin/sh\necho "hello world"\n`, fs.WithMode(0755)),
			fs.WithFile("vier.txt", `{{ (ds "config").two }} * {{ (ds "config").two }}`, fs.WithMode(0544)),
		),
		fs.WithDir("out"),
		fs.WithDir("bad_in",
			fs.WithFile("bad.tmpl", "{{end}}"),
		),
	)
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestInputDir_InputDir(t *testing.T) {
	tmpDir := setupInputDirTest(t)

	o, e, err := cmd(t,
		"--input-dir", tmpDir.Join("in"),
		"--output-dir", tmpDir.Join("out"),
		"-d", "config="+tmpDir.Join("config.yml"),
	).run()
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(tmpDir.Join("out"))
	assert.NilError(t, err)
	tassert.Len(t, files, 4)

	files, err = os.ReadDir(tmpDir.Join("out", "inner"))
	assert.NilError(t, err)
	tassert.Len(t, files, 1)

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{tmpDir.Join("out", "eins.txt"), "eins", 0644},
		{tmpDir.Join("out", "inner", "deux.txt"), "deux", 0444},
		{tmpDir.Join("out", "drei.sh"), `#!/bin/sh\necho "hello world"\n`, 0755},
		{tmpDir.Join("out", "vier.txt"), "deux * deux", 0544},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		m := iohelpers.NormalizeFileMode(v.mode)
		assert.Equal(t, m, info.Mode(), v.path)
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestInputDir_InputDirWithModeOverride(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, e, err := cmd(t,
		"--input-dir", tmpDir.Join("in"),
		"--output-dir", tmpDir.Join("out"),
		"--chmod", "0601",
		"-d", "config="+tmpDir.Join("config.yml"),
	).run()
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(tmpDir.Join("out"))
	assert.NilError(t, err)
	tassert.Len(t, files, 4)

	files, err = os.ReadDir(tmpDir.Join("out", "inner"))
	assert.NilError(t, err)
	tassert.Len(t, files, 1)

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{tmpDir.Join("out", "eins.txt"), "eins", 0601},
		{tmpDir.Join("out", "inner", "deux.txt"), "deux", 0601},
		{tmpDir.Join("out", "drei.sh"), `#!/bin/sh\necho "hello world"\n`, 0601},
		{tmpDir.Join("out", "vier.txt"), "deux * deux", 0601},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		m := iohelpers.NormalizeFileMode(v.mode)
		assert.Equal(t, m, info.Mode(), v.path)
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestInputDir_OutputMapInline(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, e, err := cmd(t,
		"--input-dir", tmpDir.Join("in"),
		"--output-map", `OUT/{{ strings.ToUpper .in }}`,
		"-d", "config.yml",
	).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(tmpDir.Join("OUT"))
	assert.NilError(t, err)
	tassert.Len(t, files, 4)

	files, err = os.ReadDir(tmpDir.Join("OUT", "INNER"))
	assert.NilError(t, err)
	tassert.Len(t, files, 1)

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{tmpDir.Join("OUT", "EINS.TXT"), "eins", 0644},
		{tmpDir.Join("OUT", "INNER", "DEUX.TXT"), "deux", 0444},
		{tmpDir.Join("OUT", "DREI.SH"), `#!/bin/sh\necho "hello world"\n`, 0755},
		{tmpDir.Join("OUT", "VIER.TXT"), "deux * deux", 0544},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		m := iohelpers.NormalizeFileMode(v.mode)
		assert.Equal(t, m, info.Mode(), v.path)
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestInputDir_OutputMapExternal(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, e, err := cmd(t,
		"--input-dir", tmpDir.Join("in"),
		"--output-map", `{{ template "out" . }}`,
		"-t", "out=out.t",
		"-c", "filemap.json",
		"-d", "config.yml",
	).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(tmpDir.Join("out"))
	assert.NilError(t, err)
	tassert.Len(t, files, 4)

	files, err = os.ReadDir(tmpDir.Join("out", "inner"))
	assert.NilError(t, err)
	tassert.Len(t, files, 1)

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{tmpDir.Join("out", "uno.out"), "eins", 0644},
		{tmpDir.Join("out", "inner", "dos.out"), "deux", 0444},
		{tmpDir.Join("out", "tres.out"), `#!/bin/sh\necho "hello world"\n`, 0755},
		{tmpDir.Join("out", "quatro.out"), "deux * deux", 0544},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		m := iohelpers.NormalizeFileMode(v.mode)
		assert.Equal(t, m, info.Mode(), v.path)
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestInputDir_DefaultOutputDir(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, e, err := cmd(t,
		"--input-dir", tmpDir.Join("in"),
		"-d", "config="+tmpDir.Join("config.yml"),
	).withDir(tmpDir.Join("out")).run()
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(tmpDir.Join("out"))
	assert.NilError(t, err)
	tassert.Len(t, files, 4)

	files, err = os.ReadDir(tmpDir.Join("out", "inner"))
	assert.NilError(t, err)
	tassert.Len(t, files, 1)

	content, err := os.ReadFile(tmpDir.Join("out", "eins.txt"))
	assert.NilError(t, err)
	assert.Equal(t, "eins", string(content))

	content, err = os.ReadFile(tmpDir.Join("out", "inner", "deux.txt"))
	assert.NilError(t, err)
	assert.Equal(t, "deux", string(content))

	content, err = os.ReadFile(tmpDir.Join("out", "drei.sh"))
	assert.NilError(t, err)
	assert.Equal(t, `#!/bin/sh\necho "hello world"\n`, string(content))

	content, err = os.ReadFile(tmpDir.Join("out", "vier.txt"))
	assert.NilError(t, err)
	assert.Equal(t, `deux * deux`, string(content))
}

func TestInputDir_ReportsFilenameWithBadInputFile(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, _, err := cmd(t,
		"--input-dir", tmpDir.Join("bad_in"),
		"--output-dir", tmpDir.Join("out"),
		"-d", "config="+tmpDir.Join("config.yml"),
	).run()
	assert.ErrorContains(t, err, "bad.tmpl:1: unexpected {{end}}")
	assert.Equal(t, "", o)
}

func TestInputDir_InputDirCwd(t *testing.T) {
	tmpDir := setupInputDirTest(t)
	o, e, err := cmd(t,
		"--input-dir", ".",
		"--include", "*.txt",
		"--output-map", `{{ .in | strings.ReplaceAll ".txt" ".out" }}`,
		"-d", "config="+tmpDir.Join("config.yml"),
	).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{tmpDir.Join("in", "eins.out"), "eins", 0644},
		{tmpDir.Join("in", "inner", "deux.out"), "deux", 0444},
		{tmpDir.Join("in", "vier.out"), "deux * deux", 0544},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		m := iohelpers.NormalizeFileMode(v.mode)
		assert.Equal(t, m, info.Mode(), v.path)
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}
