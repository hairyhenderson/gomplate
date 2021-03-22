//+build integration

package integration

import (
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type MergeDatasourceSuite struct {
	tmpDir *fs.Dir
	srv    *httptest.Server
}

var _ = Suite(&MergeDatasourceSuite{})

func (s *MergeDatasourceSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"config.json": `{"foo": {"bar": "baz"}, "isDefault": false, "isOverride": true}`,
			"default.yml": "foo:\n  bar: qux\nother: true\nisDefault: true\nisOverride: false\n",
		}),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/foo.json", typeHandler("application/json", `{"foo": "bar"}`))
	mux.HandleFunc("/1.env", typeHandler("application/x-env", "FOO=1\nBAR=2\n"))
	mux.HandleFunc("/2.env", typeHandler("application/x-env", "FOO=3\n"))

	s.srv = httptest.NewServer(mux)
}

func (s *MergeDatasourceSuite) TearDownSuite(c *C) {
	s.srv.Close()
	s.tmpDir.Remove()
}

func (s *MergeDatasourceSuite) TestMergeDatasource(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "user="+s.tmpDir.Join("config.json"),
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:user|default",
		"-i", `{{ ds "config" | toJSON }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`})

	result = icmd.RunCommand(GomplateBin,
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:user|default",
		"-i", `{{ defineDatasource "user" `+"`"+s.tmpDir.Join("config.json")+"`"+` }}{{ ds "config" | toJSON }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`})

	result = icmd.RunCommand(GomplateBin,
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:"+s.srv.URL+"/foo.json|default",
		"-i", `{{ ds "config" | toJSON }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`})

	result = icmd.RunCommand(GomplateBin,
		"-c", "merged=merge:"+s.srv.URL+"/2.env|"+s.srv.URL+"/1.env",
		"-i", `FOO is {{ .merged.FOO }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `FOO is 3`})
}
