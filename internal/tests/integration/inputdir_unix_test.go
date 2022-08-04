//go:build !windows
// +build !windows

package integration

import (
	"fmt"
	"math"
	"os"
	"testing"

	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setFileUlimit(b uint64) error {
	ulimit := unix.Rlimit{
		Cur: b,
		Max: math.MaxInt64,
	}
	err := unix.Setrlimit(unix.RLIMIT_NOFILE, &ulimit)
	return err
}

func TestInputDir_RespectsUlimit(t *testing.T) {
	numfiles := 32
	flist := map[string]string{}
	for i := 0; i < numfiles; i++ {
		k := fmt.Sprintf("file_%d", i)
		flist[k] = fmt.Sprintf("hello world %d\n", i)
	}
	testdir := fs.NewDir(t, "ulimittestfiles",
		fs.WithDir("in", fs.WithFiles(flist)),
	)
	defer testdir.Remove()

	// we need another ~11 fds for other various things, so we'd be guaranteed
	// to hit the limit if we try to have all the input files open
	// simultaneously
	setFileUlimit(uint64(numfiles))
	defer setFileUlimit(8192)

	o, e, err := cmd(t, "--input-dir", testdir.Join("in"),
		"--output-dir", testdir.Join("out")).
		withDir(testdir.Path()).run()

	setFileUlimit(8192)
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(testdir.Join("out"))
	assert.NilError(t, err)
	assert.Equal(t, numfiles, len(files))

	for i := 0; i < numfiles; i++ {
		f := testdir.Join("out", fmt.Sprintf("file_%d", i))
		_, err := os.Stat(f)
		assert.NilError(t, err)

		content, err := os.ReadFile(f)
		assert.NilError(t, err)
		expected := fmt.Sprintf("hello world %d\n", i)
		assert.Equal(t, expected, string(content))
	}
}
