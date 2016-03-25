package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var g = NewGomplate()

func testTemplate(template string) string {
	in := strings.NewReader(template)
	var out bytes.Buffer
	g.RunTemplate(in, &out)
	return strings.TrimSpace(out.String())
}

func TestGetenvTemplates(t *testing.T) {
	assert.Empty(t, testTemplate(`{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(`{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(`{{getenv "BLAHBLAHBLAH" "default value"}}`))
}

func TestBoolTemplates(t *testing.T) {
	assert.Equal(t, "true", testTemplate(`{{bool "true"}}`))
	assert.Equal(t, "false", testTemplate(`{{bool "false"}}`))
	assert.Equal(t, "false", testTemplate(`{{bool "foo"}}`))
	assert.Equal(t, "false", testTemplate(`{{bool ""}}`))
}
