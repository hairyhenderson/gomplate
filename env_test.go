package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	e := &Env{}
	assert.Empty(t, e.Getenv("FOOBARBAZ"))
	assert.Equal(t, os.Getenv("USER"), e.Getenv("USER"))
	assert.Equal(t, "default value", e.Getenv("BLAHBLAHBLAH", "default value"))
}
