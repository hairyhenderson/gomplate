package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupFileTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFile("one", "hi\n"),
		fs.WithFile("two", "hello\n"))
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestFile_ReadsFile(t *testing.T) {
	tmpDir := setupFileTest(t)

	inOutTest(t, "{{ file.Read `"+tmpDir.Join("one")+"`}}", "hi\n")
}

func TestFile_Write(t *testing.T) {
	tmpDir := setupFileTest(t)

	outDir := tmpDir.Join("writeOutput")
	os.MkdirAll(outDir, 0755)
	o, e, err := cmd(t, "-i", `{{ "hello world" | file.Write "./out" }}`).
		withDir(outDir).run()
	assertSuccess(t, o, e, err, "")

	out, err := ioutil.ReadFile(filepath.Join(outDir, "out"))
	assert.NilError(t, err)
	assert.Equal(t, "hello world", string(out))
}
