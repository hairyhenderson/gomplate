//+build !windows

package integration

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"

	. "gopkg.in/check.v1"

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

func (s *InputDirSuite) TestInputDirRespectsUlimit(c *C) {
	numfiles := 32
	flist := map[string]string{}
	for i := 0; i < numfiles; i++ {
		k := fmt.Sprintf("file_%d", i)
		flist[k] = fmt.Sprintf("hello world %d\n", i)
	}
	testdir := fs.NewDir(c, "ulimittestfiles",
		fs.WithDir("in", fs.WithFiles(flist)),
	)
	defer testdir.Remove()

	// we need another ~11 fds for other various things, so we'd be guaranteed
	// to hit the limit if we try to have all the input files open
	// simultaneously
	setFileUlimit(uint64(numfiles))
	defer setFileUlimit(8192)

	o, e, err := cmdWithDir(c, testdir.Path(),
		"--input-dir", testdir.Join("in"),
		"--output-dir", testdir.Join("out"),
	)
	setFileUlimit(8192)
	assertSuccess(c, o, e, err, "")

	files, err := ioutil.ReadDir(testdir.Join("out"))
	assert.NilError(c, err)
	assert.Equal(c, numfiles, len(files))

	for i := 0; i < numfiles; i++ {
		f := testdir.Join("out", fmt.Sprintf("file_%d", i))
		_, err := os.Stat(f)
		assert.NilError(c, err)

		content, err := ioutil.ReadFile(f)
		assert.NilError(c, err)
		expected := fmt.Sprintf("hello world %d\n", i)
		assert.Equal(c, expected, string(content))
	}
}
