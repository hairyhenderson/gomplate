package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	assert.Empty(t, Getenv("FOOBARBAZ"))
	assert.Equal(t, os.Getenv("USER"), Getenv("USER"))
	assert.Equal(t, "default value", Getenv("BLAHBLAHBLAH", "default value"))
}

func TestExpandEnv(t *testing.T) {
	assert.Empty(t, ExpandEnv("${FOOBARBAZ}"))
	assert.Equal(t, os.Getenv("USER"), ExpandEnv("$USER"))
	assert.Equal(t, "something", ExpandEnv("something$BLAHBLAHBLAH"))
	assert.Equal(t, os.Getenv("USER")+": "+os.Getenv("HOME"),
		ExpandEnv("$USER: ${HOME}"))
}
