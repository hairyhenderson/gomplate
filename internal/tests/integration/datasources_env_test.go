package integration

import (
	"testing"
)

func TestDatasources_Env(t *testing.T) {

	o, e, err := cmd(t, "-d", "foo=env:FOO", "-i", `{{ ds "foo" }}`).
		withEnv("HELLO_WORLD", "hello world").
		withEnv("HELLO_UNIVERSE", "hello universe").
		withEnv("FOO", "bar").
		withEnv("foo", "baz").
		run()
	// Windows envvars are case-insensitive
	if isWindows {
		assertSuccess(t, o, e, err, "baz")
	} else {
		assertSuccess(t, o, e, err, "bar")
	}

	o, e, err = cmd(t,
		"-d", "foo=env:///foo", "-i", `{{ ds "foo" }}`).
		withEnv("HELLO_WORLD", "hello world").
		withEnv("HELLO_UNIVERSE", "hello universe").
		withEnv("FOO", "bar").
		withEnv("foo", "baz").
		run()
	assertSuccess(t, o, e, err, "baz")

	o, e, err = cmd(t, "-d", "e=env:json_value?type=application/json",
		"-i", `{{ (ds "e").value}}`).
		withEnv("json_value", `{"value":"corge"}`).
		run()
	assertSuccess(t, o, e, err, "corge")
}
