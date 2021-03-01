package integration

import (
	"os"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
)

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpTest(c *C) {
}

func (s *TestSuite) TearDownTest(c *C) {
}

func (s *TestSuite) TestFail(c *C) {
	_, _, err := cmdTest(c, "-i", "{{ fail }}")
	assert.ErrorContains(c, err, `template generation failed`)

	_, _, err = cmdTest(c, "-i", "{{ fail `some message` }}")
	assert.ErrorContains(c, err, `some message`)
}

func (s *TestSuite) TestRequired(c *C) {
	os.Unsetenv("FOO")
	_, _, err := cmdTest(c, "-i", `{{getenv "FOO" | required "FOO missing" }}`)
	assert.ErrorContains(c, err, "FOO missing")

	o, e, err := cmdWithEnv(c, []string{"-i", `{{getenv "FOO" | required "FOO missing" }}`},
		map[string]string{"FOO": "bar"})
	assertSuccess(c, o, e, err, "bar")

	_, _, err = cmdWithStdin(c, []string{"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required "foo should not be null" }}`},
		`foo: null`)
	assert.ErrorContains(c, err, "foo should not be null")

	o, e, err = cmdWithStdin(c, []string{
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`},
		`foo: []`)
	assertSuccess(c, o, e, err, "[]")

	o, e, err = cmdWithStdin(c, []string{
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`},
		`foo: {}`)
	assertSuccess(c, o, e, err, "map[]")

	o, e, err = cmdWithStdin(c, []string{
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`},
		`foo: 0`)
	assertSuccess(c, o, e, err, "0")

	o, e, err = cmdWithStdin(c, []string{
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`},
		`foo: false`)
	assertSuccess(c, o, e, err, "false")
}
