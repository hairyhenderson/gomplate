//go:build !windows
// +build !windows

package integration

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setFileUlimit(b uint64) error {
	ulimit := unix.Rlimit{}
	err := unix.Getrlimit(unix.RLIMIT_NOFILE, &ulimit)
	if err != nil {
		return err
	}

	ulimit.Cur = b
	err = unix.Setrlimit(unix.RLIMIT_NOFILE, &ulimit)
	return err
}

func checkFileUlimit(t *testing.T, b uint64) {
	ulimit := unix.Rlimit{}
	err := unix.Getrlimit(unix.RLIMIT_NOFILE, &ulimit)
	assert.NilError(t, err)
	assert.Equal(t, b, ulimit.Cur)
}

func TestInputDir_RespectsUlimit(t *testing.T) {
	numfiles := uint32(32)
	flist := map[string]string{}
	for i := range int(numfiles) {
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

	// make sure the ulimit is actually set correctly - the capability may not
	// be available in some environments, and we don't want to pass if we can't
	// actually test the behavior
	checkFileUlimit(t, uint64(numfiles))

	o, e, err := cmd(t, "--input-dir", testdir.Join("in"),
		"--output-dir", testdir.Join("out")).
		withDir(testdir.Path()).run()

	setFileUlimit(8192)
	checkFileUlimit(t, 8192)
	assertSuccess(t, o, e, err, "")

	files, err := os.ReadDir(testdir.Join("out"))
	assert.NilError(t, err)
	assert.Equal(t, int(numfiles), len(files))

	for i := range int(numfiles) {
		f := testdir.Join("out", fmt.Sprintf("file_%d", i))
		_, err := os.Stat(f)
		assert.NilError(t, err)

		content, err := os.ReadFile(f)
		assert.NilError(t, err)
		expected := fmt.Sprintf("hello world %d\n", i)
		assert.Equal(t, expected, string(content))
	}
}
