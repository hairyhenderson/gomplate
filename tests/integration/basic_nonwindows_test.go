//+build integration
//+build !windows

package integration

import (
	"bytes"
	"io/ioutil"
	"os"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/icmd"
)

func (s *BasicSuite) TestRoutesInputsToProperOutputsWithChmod(c *C) {
	oneOut := s.tmpDir.Join("one.out")
	twoOut := s.tmpDir.Join("two.out")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
		"--chmod", "0600"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Success)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{oneOut, 0600, "hi\n"},
		{twoOut, 0600, "hello\n"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.mode, info.Mode())
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *BasicSuite) TestOverridesOutputModeWithChmod(c *C) {
	out := s.tmpDir.Join("two")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0600"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Success)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{out, 0600, "hi\n"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.mode, info.Mode())
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}
