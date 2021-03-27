package integration

import (
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestTest_Fail(t *testing.T) {
	_, _, err := cmd(t, "-i", "{{ fail }}").run()
	assert.ErrorContains(t, err, `template generation failed`)

	_, _, err = cmd(t, "-i", "{{ fail `some message` }}").run()
	assert.ErrorContains(t, err, `some message`)
}

func TestTest_Required(t *testing.T) {
	os.Unsetenv("FOO")
	_, _, err := cmd(t, "-i", `{{getenv "FOO" | required "FOO missing" }}`).run()
	assert.ErrorContains(t, err, "FOO missing")

	o, e, err := cmd(t, "-i", `{{getenv "FOO" | required "FOO missing" }}`).
		withEnv("FOO", "bar").run()
	assertSuccess(t, o, e, err, "bar")

	_, _, err = cmd(t, "-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required "foo should not be null" }}`).
		withStdin(`foo: null`).run()
	assert.ErrorContains(t, err, "foo should not be null")

	o, e, err = cmd(t, "-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`).
		withStdin(`foo: []`).run()
	assertSuccess(t, o, e, err, "[]")

	o, e, err = cmd(t, "-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`).
		withStdin(`foo: {}`).run()
	assertSuccess(t, o, e, err, "map[]")

	o, e, err = cmd(t, "-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`).
		withStdin(`foo: 0`).run()
	assertSuccess(t, o, e, err, "0")

	o, e, err = cmd(t, "-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`).
		withStdin(`foo: false`).run()
	assertSuccess(t, o, e, err, "false")
}
