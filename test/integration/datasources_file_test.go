//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type FileDatasourcesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&FileDatasourcesSuite{})

func (s *FileDatasourcesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("config.json", `{"foo": {"bar": "baz"}}`),
		fs.WithFile("config.yml", "foo:\n bar: baz\n"),
		fs.WithFile("config2.yml", "foo: bar\n"),
		fs.WithFile("foo.csv", `A,B
A1,B1
A2,"foo""
bar"
`),
	)
}

func (s *FileDatasourcesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *FileDatasourcesSuite) TestFileDatasources(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config.yml"),
		"-i", `{{(datasource "config").foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{(ds "config").foo}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "csv="+s.tmpDir.Join("foo.csv"),
		"-i", `{{ index (index (ds "csv") 2) 1 }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `foo"
bar`})

	result = icmd.RunCommand(GomplateBin,
		"-d", "config="+s.tmpDir.Join("config2.yml"),
		"-i", `{{ include "config" }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `foo: bar`})
}
