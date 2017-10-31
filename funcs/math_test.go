package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	m := MathNS()
	assert.Equal(t, int64(12), m.Add(1, 1, 2, 3, 5))
	assert.Equal(t, int64(2), m.Add(1, 1))
	assert.Equal(t, int64(1), m.Add(1))
	assert.Equal(t, int64(0), m.Add(-5, 5))
}

func TestMul(t *testing.T) {
	m := MathNS()
	assert.Equal(t, int64(30), m.Mul(1, 1, 2, 3, 5))
	assert.Equal(t, int64(1), m.Mul(1, 1))
	assert.Equal(t, int64(1), m.Mul(1))
	assert.Equal(t, int64(-25), m.Mul("-5", 5))
	assert.Equal(t, int64(28), m.Mul(14, "2"))
}

func TestSub(t *testing.T) {
	m := MathNS()
	assert.Equal(t, int64(0), m.Sub(1, 1))
	assert.Equal(t, int64(-10), m.Sub(-5, 5))
	assert.Equal(t, int64(-41), m.Sub(true, "42"))
}

func mustDiv(a, b interface{}) int64 {
	m := MathNS()
	r, err := m.Div(a, b)
	if err != nil {
		return -1
	}
	return r
}

func TestDiv(t *testing.T) {
	m := MathNS()
	_, err := m.Div(1, 0)
	assert.Error(t, err)
	assert.Equal(t, int64(1), mustDiv(1, 1))
	assert.Equal(t, int64(-1), mustDiv(-5, 5))
	assert.Equal(t, int64(0), mustDiv(true, "42"))
}

func TestRem(t *testing.T) {
	m := MathNS()
	assert.Equal(t, int64(0), m.Rem(1, 1))
	assert.Equal(t, int64(2), m.Rem(5, 3.0))
	// assert.Equal(t, int64(1), m.Mod(true, "42"))
}

func TestPow(t *testing.T) {
	m := MathNS()
	assert.Equal(t, int64(4), m.Pow(2, "2"))
}
