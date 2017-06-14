package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	re := &ReFuncs{}
	assert.Equal(t, "hello world", re.Replace("i", "ello", "hi world"))
}

func TestMatch(t *testing.T) {
	re := &ReFuncs{}
	assert.True(t, re.Match(`i\ `, "hi world"))
}
