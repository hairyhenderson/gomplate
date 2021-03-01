package integration

import (
	"net"
	"net/http"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
)

type MergeDatasourceSuite struct {
	tmpDir *fs.Dir
	l      *net.TCPListener
}

var _ = Suite(&MergeDatasourceSuite{})

func (s *MergeDatasourceSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"config.json": `{"foo": {"bar": "baz"}, "isDefault": false, "isOverride": true}`,
			"default.yml": "foo:\n  bar: qux\nother: true\nisDefault: true\nisOverride: false\n",
		}),
	)

	var err error
	s.l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	handle(c, err)

	http.HandleFunc("/foo.json", typeHandler("application/json", `{"foo": "bar"}`))
	http.HandleFunc("/1.env", typeHandler("application/x-env", "FOO=1\nBAR=2\n"))
	http.HandleFunc("/2.env", typeHandler("application/x-env", "FOO=3\n"))
	go http.Serve(s.l, nil)
}

func (s *MergeDatasourceSuite) TearDownSuite(c *C) {
	s.l.Close()
	s.tmpDir.Remove()
}

func (s *MergeDatasourceSuite) TestMergeDatasource(c *C) {
	o, e, err := cmdTest(c,
		"-d", "user="+s.tmpDir.Join("config.json"),
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:user|default",
		"-i", `{{ ds "config" | toJSON }}`,
	)
	assertSuccess(c, o, e, err, `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`)

	o, e, err = cmdTest(c,
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:user|default",
		"-i", `{{ defineDatasource "user" `+"`"+s.tmpDir.Join("config.json")+"`"+` }}{{ ds "config" | toJSON }}`,
	)
	assertSuccess(c, o, e, err, `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`)

	o, e, err = cmdTest(c,
		"-d", "default="+s.tmpDir.Join("default.yml"),
		"-d", "config=merge:http://"+s.l.Addr().String()+"/foo.json|default",
		"-i", `{{ ds "config" | toJSON }}`,
	)
	assertSuccess(c, o, e, err, `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`)

	o, e, err = cmdTest(c,
		"-c", "merged=merge:http://"+s.l.Addr().String()+"/2.env|http://"+s.l.Addr().String()+"/1.env",
		"-i", `FOO is {{ .merged.FOO }}`,
	)
	assertSuccess(c, o, e, err, `FOO is 3`)
}
