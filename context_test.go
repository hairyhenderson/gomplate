package gomplate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapifiesEnvironment(t *testing.T) {
	c := &context{}
	env := c.Env()
	assert.Equal(t, env["USER"], os.Getenv("USER"))
}

func TestEnvGetsUpdatedEnvironment(t *testing.T) {
	c := &context{}
	assert.Empty(t, c.Env()["FOO"])
	assert.NoError(t, os.Setenv("FOO", "foo"))
	assert.Equal(t, c.Env()["FOO"], "foo")
}
