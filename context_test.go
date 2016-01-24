package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapifiesEnvironment(t *testing.T) {
	c := &Context{}
	env := c.Env()
	assert.Equal(t, env["USER"], os.Getenv("USER"))
}

func TestEnvGetsUpdatedEnvironment(t *testing.T) {
	c := &Context{}
	assert.Empty(t, c.Env()["FOO"])
	os.Setenv("FOO", "foo")
	assert.Equal(t, c.Env()["FOO"], "foo")
}
