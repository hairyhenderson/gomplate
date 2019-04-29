//+build integration

package integration

import (
	"io/ioutil"
	"os"
	"runtime"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
	tassert "github.com/stretchr/testify/assert"
)

type InputDirSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&InputDirSuite{})

func (s *InputDirSuite) SetUpTest(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
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
}

func (s *InputDirSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *InputDirSuite) TestInputDir(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-dir", s.tmpDir.Join("out"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	)
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 4)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{s.tmpDir.Join("out", "eins.txt"), 0644, "eins"},
		{s.tmpDir.Join("out", "inner", "deux.txt"), 0444, "deux"},
		{s.tmpDir.Join("out", "drei.sh"), 0755, `#!/bin/sh\necho "hello world"\n`},
		{s.tmpDir.Join("out", "vier.txt"), 0544, "deux * deux"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		// chmod support on Windows is pretty weak for now
		if runtime.GOOS != "windows" {
			assert.Equal(c, v.mode, info.Mode(), v.path)
		}
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *InputDirSuite) TestInputDirWithModeOverride(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-dir", s.tmpDir.Join("out"),
		"--chmod", "0601",
		"-d", "config="+s.tmpDir.Join("config.yml"),
	)
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 4)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{s.tmpDir.Join("out", "eins.txt"), 0601, "eins"},
		{s.tmpDir.Join("out", "inner", "deux.txt"), 0601, "deux"},
		{s.tmpDir.Join("out", "drei.sh"), 0601, `#!/bin/sh\necho "hello world"\n`},
		{s.tmpDir.Join("out", "vier.txt"), 0601, "deux * deux"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		// chmod support on Windows is pretty weak for now
		if runtime.GOOS != "windows" {
			assert.Equal(c, v.mode, info.Mode())
		}
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *InputDirSuite) TestOutputMapInline(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-map", `OUT/{{ strings.ToUpper .in }}`,
		"-d", "config.yml",
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("OUT"))
	assert.NilError(c, err)
	tassert.Len(c, files, 4)

	files, err = ioutil.ReadDir(s.tmpDir.Join("OUT", "INNER"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{s.tmpDir.Join("OUT", "EINS.TXT"), 0644, "eins"},
		{s.tmpDir.Join("OUT", "INNER", "DEUX.TXT"), 0444, "deux"},
		{s.tmpDir.Join("OUT", "DREI.SH"), 0755, `#!/bin/sh\necho "hello world"\n`},
		{s.tmpDir.Join("OUT", "VIER.TXT"), 0544, "deux * deux"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		// chmod support on Windows is pretty weak for now
		if runtime.GOOS != "windows" {
			assert.Equal(c, v.mode, info.Mode())
		}
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *InputDirSuite) TestOutputMapExternal(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-map", `{{ template "out" . }}`,
		"-t", "out=out.t",
		"-c", "filemap.json",
		"-d", "config.yml",
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 4)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{s.tmpDir.Join("out", "uno.out"), 0644, "eins"},
		{s.tmpDir.Join("out", "inner", "dos.out"), 0444, "deux"},
		{s.tmpDir.Join("out", "tres.out"), 0755, `#!/bin/sh\necho "hello world"\n`},
		{s.tmpDir.Join("out", "quatro.out"), 0544, "deux * deux"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		// chmod support on Windows is pretty weak for now
		if runtime.GOOS != "windows" {
			assert.Equal(c, v.mode, info.Mode())
		}
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *InputDirSuite) TestDefaultOutputDir(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Join("out")
	})
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 4)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	content, err := ioutil.ReadFile(s.tmpDir.Join("out", "eins.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "eins", string(content))

	content, err = ioutil.ReadFile(s.tmpDir.Join("out", "inner", "deux.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "deux", string(content))

	content, err = ioutil.ReadFile(s.tmpDir.Join("out", "drei.sh"))
	assert.NilError(c, err)
	assert.Equal(c, `#!/bin/sh\necho "hello world"\n`, string(content))

	content, err = ioutil.ReadFile(s.tmpDir.Join("out", "vier.txt"))
	assert.NilError(c, err)
	assert.Equal(c, `deux * deux`, string(content))
}

func (s *InputDirSuite) TestReportsFilenameWithBadInputFile(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--input-dir", s.tmpDir.Join("bad_in"),
		"--output-dir", s.tmpDir.Join("out"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	)
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "template: " + s.tmpDir.Join("bad_in", "bad.tmpl") + ":1: unexpected {{end}}",
	})
}
