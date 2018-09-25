//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type NestedTemplatesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&NestedTemplatesSuite{})

func (s *NestedTemplatesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("hello.t", `Hello {{ . }}!`),
		fs.WithDir("templates",
			fs.WithFile("one.t", `{{ . }}`),
			fs.WithFile("two.t", `{{ range $n := (seq 2) }}{{ $n }}: {{ $ }} {{ end }}`),
		),
	)
}

func (s *NestedTemplatesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *NestedTemplatesSuite) TestNestedTemplates(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-t", "hello="+s.tmpDir.Join("hello.t"),
		"-i", `{{ template "hello" "World"}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Hello World!"})

	result = icmd.RunCmd(icmd.Cmd{
		Command: []string{
			GomplateBin,
			"-t", "hello.t",
			"-i", `{{ template "hello.t" "World"}}`,
		},
		Dir: s.tmpDir.Path(),
	},
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Hello World!"})

	result = icmd.RunCmd(icmd.Cmd{
		Command: []string{
			GomplateBin,
			"-t", "templates/",
			"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`,
		},
		Dir: s.tmpDir.Path(),
	},
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `one
1: two 2: two`})
}
