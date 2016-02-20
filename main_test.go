package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTemplate(template string) string {
	in := strings.NewReader(template)
	var out bytes.Buffer
	RunTemplate(in, &out)
	return strings.TrimSpace(out.String())
}

func TestGetenv(t *testing.T) {
	assert.Empty(t, Getenv("FOOBARBAZ"))
	assert.Empty(t, testTemplate(`{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, Getenv("USER"), os.Getenv("USER"))
	assert.Equal(t, os.Getenv("USER"), testTemplate(`{{getenv "USER"}}`))
}

func TestBool(t *testing.T) {
	assert.False(t, Bool(""))
	assert.False(t, Bool("asdf"))
	assert.False(t, Bool("1234"))
	assert.False(t, Bool("False"))
	assert.False(t, Bool("0"))
	assert.False(t, Bool("false"))
	assert.False(t, Bool("F"))
	assert.False(t, Bool("f"))
	assert.True(t, Bool("true"))
	assert.True(t, Bool("True"))
	assert.True(t, Bool("t"))
	assert.True(t, Bool("T"))
	assert.True(t, Bool("1"))
}
