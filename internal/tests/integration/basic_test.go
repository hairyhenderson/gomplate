package integration

import (
	"io/fs"
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	tfs "gotest.tools/v3/fs"
)

func setupBasicTest(t *testing.T) *tfs.Dir {
	t.Helper()

	tmpDir := tfs.NewDir(t, "gomplate-inttests",
		tfs.WithFile("one", "hi\n", tfs.WithMode(0o640)),
		tfs.WithFile("two", "hello\n"),
		tfs.WithFile("broken", "", tfs.WithMode(0o000)),
		tfs.WithDir("subdir",
			tfs.WithFile("f1", "first\n", tfs.WithMode(0o640)),
			tfs.WithFile("f2", "second\n"),
		),
	)
	t.Cleanup(tmpDir.Remove)
	return tmpDir
}

func TestBasic_ReportsVersion(t *testing.T) {
	o, e, err := cmd(t, "-v").run()
	assert.NilError(t, err)
	assert.Equal(t, "", e)
	assert.Assert(t, cmp.Contains(o, "gomplate version "))
}

func TestBasic_TakesStdinByDefault(t *testing.T) {
	o, e, err := cmd(t).withStdin("hello world").run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestBasic_TakesStdinWithFileFlag(t *testing.T) {
	o, e, err := cmd(t, "--file", "-").withStdin("hello world").run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestBasic_WritesToStdoutWithOutFlag(t *testing.T) {
	o, e, err := cmd(t, "--out", "-").withStdin("hello world").run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestBasic_IgnoresStdinWithInFlag(t *testing.T) {
	o, e, err := cmd(t, "--in", "hi").withStdin("hello world").run()
	assertSuccess(t, o, e, err, "hi")
}

func TestBasic_ErrorsWithInputOutputImbalance(t *testing.T) {
	tmpDir := setupBasicTest(t)

	_, _, err := cmd(t,
		"-f", tmpDir.Join("one"),
		"-f", tmpDir.Join("two"),
		"-o", tmpDir.Join("out"),
	).run()
	assert.ErrorContains(t, err, "must provide same number of 'outputFiles' (1) as 'in' or 'inputFiles' (2) options")
}

func TestBasic_RoutesInputsToProperOutputs(t *testing.T) {
	tmpDir := setupBasicTest(t)
	oneOut := tmpDir.Join("one.out")
	twoOut := tmpDir.Join("two.out")

	o, e, err := cmd(t,
		"-f", tmpDir.Join("one"),
		"-f", tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
	).run()
	assertSuccess(t, o, e, err, "")

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{oneOut, "hi\n", 0640},
		{twoOut, "hello\n", 0644},
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

func TestBasic_FlagRules(t *testing.T) {
	testdata := []struct {
		errmsg string
		args   []string
	}{
		{
			"only one of these options is supported at a time: 'in', 'inputFiles'",
			[]string{"-f", "-", "-i", "HELLO WORLD"},
		},
		{
			"these options must be set together: 'outputDir', 'inputDir'",
			[]string{"--output-dir", "."},
		},
		{
			"only one of these options is supported at a time: 'in', 'inputDir'",
			[]string{"--input-dir", ".", "--in", "param"},
		},
		{
			"only one of these options is supported at a time: 'inputFiles', 'inputDir'",
			[]string{"--input-dir", ".", "--file", "input.txt"},
		},
		{
			"only one of these options is supported at a time: 'outputFiles', 'outputDir'",
			[]string{"--output-dir", ".", "--out", "param"},
		},
		{
			"only one of these options is supported at a time: 'outputFiles', 'outputMap'",
			[]string{"--output-map", ".", "--out", "param"},
		},
	}

	for _, d := range testdata {
		_, _, err := cmd(t, d.args...).run()
		assert.ErrorContains(t, err, d.errmsg)
	}
}

func TestBasic_DelimsChangedThroughOpts(t *testing.T) {
	o, e, err := cmd(t,
		"--left-delim", "((",
		"--right-delim", "))",
		"-i", `foo((print "hi"))`,
	).run()
	assertSuccess(t, o, e, err, "foohi")
}

func TestBasic_DelimsChangedThroughEnvVars(t *testing.T) {
	o, e, err := cmd(t, "-i", `foo<<print "hi">>`).
		withEnv("GOMPLATE_LEFT_DELIM", "<<").
		withEnv("GOMPLATE_RIGHT_DELIM", ">>").
		run()
	assertSuccess(t, o, e, err, "foohi")
}

func TestBasic_UnknownArgErrors(t *testing.T) {
	_, _, err := cmd(t, "-in", "flibbit").run()
	assert.ErrorContains(t, err, `unknown command "flibbit" for "gomplate"`)
}

func TestBasic_ExecCommand(t *testing.T) {
	tmpDir := setupBasicTest(t)
	out := tmpDir.Join("out")
	o, e, err := cmd(t, "-i", `{{print "hello world"}}`,
		"-o", out,
		"--", "cat", out).run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestBasic_PostRunExecPipe(t *testing.T) {
	o, e, err := cmd(t,
		"-i", `{{print "hello world"}}`,
		"--exec-pipe",
		"--", "tr", "a-z", "A-Z").run()
	assertSuccess(t, o, e, err, "HELLO WORLD")
}

func TestBasic_EmptyOutputSuppression(t *testing.T) {
	tmpDir := setupBasicTest(t)
	out := tmpDir.Join("out")
	o, e, err := cmd(t, "-i", `{{print "\t  \n\n\r\n\t\t     \v\n"}}`,
		"-o", out).
		withEnv("GOMPLATE_SUPPRESS_EMPTY", "true").run()
	assertSuccess(t, o, e, err, "")

	_, err = os.Stat(out)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestBasic_RoutesInputsToProperOutputsWithChmod(t *testing.T) {
	tmpDir := setupBasicTest(t)
	oneOut := tmpDir.Join("one.out")
	twoOut := tmpDir.Join("two.out")

	o, e, err := cmd(t,
		"-f", tmpDir.Join("one"),
		"-f", tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
		"--chmod", "0600").
		withStdin("hello world").run()
	assertSuccess(t, o, e, err, "")

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{oneOut, "hi\n", 0600},
		{twoOut, "hello\n", 0600},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		assert.Equal(t, iohelpers.NormalizeFileMode(v.mode), info.Mode())
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestBasic_OverridesOutputModeWithChmod(t *testing.T) {
	tmpDir := setupBasicTest(t)
	out := tmpDir.Join("two")

	o, e, err := cmd(t,
		"-f", tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0600").
		withStdin("hello world").run()
	assertSuccess(t, o, e, err, "")

	testdata := []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{out, "hi\n", 0600},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(t, err)
		assert.Equal(t, iohelpers.NormalizeFileMode(v.mode), info.Mode())
		content, err := os.ReadFile(v.path)
		assert.NilError(t, err)
		assert.Equal(t, v.content, string(content))
	}
}

func TestBasic_AppliesChmodBeforeWrite(t *testing.T) {
	tmpDir := setupBasicTest(t)

	// 'broken' was created with mode 0000
	out := tmpDir.Join("broken")
	_, _, err := cmd(t,
		"-f", tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0644").run()
	assert.NilError(t, err)

	info, err := os.Stat(out)
	assert.NilError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), info.Mode())
	content, err := os.ReadFile(out)
	assert.NilError(t, err)
	assert.Equal(t, "hi\n", string(content))
}

func TestBasic_CreatesMissingDirectory(t *testing.T) {
	tmpDir := setupBasicTest(t)
	out := tmpDir.Join("foo/bar/baz")
	o, e, err := cmd(t, "-f", tmpDir.Join("one"), "-o", out).run()
	assertSuccess(t, o, e, err, "")

	info, err := os.Stat(out)
	assert.NilError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0640), info.Mode())
	content, err := os.ReadFile(out)
	assert.NilError(t, err)
	assert.Equal(t, "hi\n", string(content))

	out = tmpDir.Join("outdir")
	o, e, err = cmd(t,
		"--input-dir", tmpDir.Join("subdir"),
		"--output-dir", out,
	).run()
	assertSuccess(t, o, e, err, "")

	info, err = os.Stat(out)
	assert.NilError(t, err)

	assert.Equal(t, iohelpers.NormalizeFileMode(0o755|fs.ModeDir), info.Mode())
	assert.Equal(t, true, info.IsDir())
}
