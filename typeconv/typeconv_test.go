package typeconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustParseBool(t *testing.T) {
	for _, b := range []string{"1", "t", "T", "true", "TRUE", "True"} {
		assert.True(t, MustParseBool(b))
	}
	for _, b := range []string{"0", "f", "F", "false", "FALSE", "False", "", "gibberish", "12345"} {
		assert.False(t, MustParseBool(b))
	}
}

func TestMustParseInt(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0, int(MustParseInt(i, 10, 64)))
	}
	assert.Equal(t, 1, int(MustParseInt("1", 10, 64)))
	assert.Equal(t, -1, int(MustParseInt("-1", 10, 64)))
}

func TestMustAtoi(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0, MustAtoi(i))
	}
	assert.Equal(t, 1, MustAtoi("1"))
	assert.Equal(t, -1, MustAtoi("-1"))
}

func TestMustParseUint(t *testing.T) {
	for _, i := range []string{"0", "-0", "-1", "foo", "", "*&^%"} {
		assert.Equal(t, uint64(0), MustParseUint(i, 10, 64))
	}
	assert.Equal(t, uint64(1), MustParseUint("1", 10, 64))
}

func TestMustParseFloat(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0.0, MustParseFloat(i, 64))
	}
	assert.Equal(t, 1.0, MustParseFloat("1", 64))
	assert.Equal(t, -1.0, MustParseFloat("-1", 64))
}
