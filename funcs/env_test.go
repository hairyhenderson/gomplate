package funcs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvGetenv(t *testing.T) {
	ef := &EnvFuncs{}
	expected := os.Getenv("USER")
	assert.Equal(t, expected, ef.Getenv("USER"))

	assert.Equal(t, "foo", ef.Getenv("bogusenvvar", "foo"))
}
