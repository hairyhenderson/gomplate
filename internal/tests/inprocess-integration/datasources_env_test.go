package integration

import (
	"os"

	. "gopkg.in/check.v1"
)

type EnvDatasourcesSuite struct {
}

var _ = Suite(&EnvDatasourcesSuite{})

func (s *EnvDatasourcesSuite) SetUpSuite(c *C) {
	os.Setenv("HELLO_WORLD", "hello world")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	os.Setenv("FOO", "bar")
	os.Setenv("foo", "baz")
}

func (s *EnvDatasourcesSuite) TearDownSuite(c *C) {
	os.Unsetenv("HELLO_WORLD")
	os.Unsetenv("HELLO_UNIVERSE")
	os.Unsetenv("FOO")
	os.Unsetenv("foo")
}

func (s *EnvDatasourcesSuite) TestEnvDatasources(c *C) {
	o, e, err := cmdTest(c, "-d", "foo=env:FOO", "-i", `{{ ds "foo" }}`)

	// Windows envvars are case-insensitive
	if isWindows {
		assertSuccess(c, o, e, err, "baz")
	} else {
		assertSuccess(c, o, e, err, "bar")
	}

	o, e, err = cmdTest(c, "-d", "foo=env:///foo", "-i", `{{ ds "foo" }}`)
	assertSuccess(c, o, e, err, "baz")

	o, e, err = cmdWithEnv(c, []string{"-d", "e=env:json_value?type=application/json",
		"-i", `{{ (ds "e").value}}`},
		map[string]string{"json_value": `{"value":"corge"}`})
	assertSuccess(c, o, e, err, "corge")
}
