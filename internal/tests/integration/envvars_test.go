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
	_, _, err = cmd(t, "-i", `{{ env.Env.FOO }}`).run()
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
		`{{ env.Env.FOO }}`,
		`{{ (env.Env).FOO }}`,
	}
	for _, in := range data {
		o, e, err := cmd(t, "-i", in).
			withEnv("FOO", "foo").run()
		assertSuccess(t, o, e, err, "foo")
	}
}

func TestEnvVars_HasEnv(t *testing.T) {
	os.Unsetenv("FOO")

	t.Run("non-existent var returns false", func(t *testing.T) {
		inOutTest(t, `{{ env.HasEnv "FOO" }}`, "false")
	})

	t.Run("empty var still returns true", func(t *testing.T) {
		o, e, err := cmd(t, "-i", `{{ env.HasEnv "FOO" }}`).
			withEnv("FOO", "").run()
		assertSuccess(t, o, e, err, "true")
	})

	t.Run("existent var (including empty) returns true", func(t *testing.T) {
		data := []string{
			`{{ env.HasEnv "FOO" }}`,
			`{{ "FOO" | env.HasEnv }}`,
		}
		for _, in := range data {
			o, e, err := cmd(t, "-i", in).withEnv("FOO", "bar").run()
			assertSuccess(t, o, e, err, "true")
		}
	})
}
