package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
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
	o, e, err := cmdTest(c,
		"-t", "hello="+s.tmpDir.Join("hello.t"),
		"-i", `{{ template "hello" "World"}}`,
	)
	assertSuccess(c, o, e, err, "Hello World!")

	o, e, err = cmdWithDir(c, s.tmpDir.Path(),
		"-t", "hello.t",
		"-i", `{{ template "hello.t" "World"}}`)
	assertSuccess(c, o, e, err, "Hello World!")

	o, e, err = cmdWithDir(c, s.tmpDir.Path(),
		"-t", "templates/",
		"-i", `{{ template "templates/one.t" "one"}}
{{ template "templates/two.t" "two"}}`,
	)
	assertSuccess(c, o, e, err, "one\n1: two 2: two ")
}
