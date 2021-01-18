package integration

import (
	"os"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
)

type EnvvarsSuite struct{}

var _ = Suite(&EnvvarsSuite{})

func (s *EnvvarsSuite) TestNonExistantEnvVar(c *C) {
	os.Unsetenv("FOO")
	_, _, err := cmdTest(c, "-i", `{{ .Env.FOO }}`)
	assert.ErrorContains(c, err, "map has no entry for key")

	o, e, err := cmdTest(c, "-i", `{{ getenv "FOO" }}`)
	assertSuccess(c, o, e, err, "")

	o, e, err = cmdTest(c, "-i", `{{ getenv "FOO" "foo" }}`)
	assertSuccess(c, o, e, err, "foo")

	o, e, err = cmdTest(c, "-i", `{{env.ExpandEnv "${BAR}foo"}}`)
	assertSuccess(c, o, e, err, "foo")

	o, e, err = cmdWithEnv(c, []string{"-i", `{{ getenv "FOO" "foo" }}`},
		map[string]string{"FOO": ""})
	assertSuccess(c, o, e, err, "foo")
}

func (s *EnvvarsSuite) TestExistantEnvVar(c *C) {
	os.Unsetenv("FOO")

	env := map[string]string{"FOO": "foo"}
	expected := "foo"

	o, e, err := cmdWithEnv(c, []string{"-i", `{{ .Env.FOO }}`}, env)
	assertSuccess(c, o, e, err, expected)

	o, e, err = cmdWithEnv(c, []string{"-i", `{{ getenv "FOO" }}`}, env)
	assertSuccess(c, o, e, err, expected)

	o, e, err = cmdWithEnv(c, []string{"-i", `{{ env.Getenv "FOO" }}`}, env)
	assertSuccess(c, o, e, err, expected)

	o, e, err = cmdWithEnv(c, []string{"-i", `{{env.ExpandEnv "${FOO}"}}`}, env)
	assertSuccess(c, o, e, err, expected)
}
