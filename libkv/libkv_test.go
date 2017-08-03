package libkv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustParseBool(t *testing.T) {
	for _, b := range []string{"1", "t", "T", "true", "TRUE", "True"} {
		assert.True(t, mustParseBool(b))
	}
	for _, b := range []string{"0", "f", "F", "false", "FALSE", "False", "", "gibberish", "12345"} {
		assert.False(t, mustParseBool(b))
	}
}

func TestMustParseInt(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0, int(mustParseInt(i)))
	}
	assert.Equal(t, 1, int(mustParseInt("1")))
	assert.Equal(t, -1, int(mustParseInt("-1")))
}
