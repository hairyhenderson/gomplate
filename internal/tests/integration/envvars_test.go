package integration

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestEnvVars_NonExistent(t *testing.T) {
	os.Unsetenv("FOO")
	_, _, err := cmd(t, "-i", `{{ .Env.FOO }}`).run()
	assert.ErrorContains(t, err, "map has no entry for key")

	inOutTest(t, `{{ getenv "FOO" }}`, "")
	inOutTest(t, `{{ getenv "FOO" "foo" }}`, "foo")
	inOutTest(t, `{{env.ExpandEnv "${BAR}foo"}}`, "foo")

	o, e, err := cmd(t, "-i", `{{ getenv "FOO" "foo" }}`).
		withEnv("FOO", "").run()
	assertSuccess(t, o, e, err, "foo")
}

func TestEnvVars_Existent(t *testing.T) {
	os.Unsetenv("FOO")

	data := []string{
		`{{ .Env.FOO }}`,
		`{{ getenv "FOO" }}`,
		`{{ env.Getenv "FOO" }}`,
		`{{env.ExpandEnv "${FOO}"}}`,
	}
	for _, in := range data {
		o, e, err := cmd(t, "-i", in).
			withEnv("FOO", "foo").run()
		assertSuccess(t, o, e, err, "foo")
	}
}
