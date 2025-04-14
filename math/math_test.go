package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMath(t *testing.T) {
	assert.Equal(t, int64(10), AddInt(1, 2, 3, 4))
	assert.Equal(t, int64(12), MulInt(3, 4, 1))
}

func TestSeq(t *testing.T) {
	assert.Equal(t, []int64{1, 2, 3}, Seq(1, 3, 1))
	assert.Equal(t, []int64{1, 3}, Seq(1, 3, 2))
	assert.Equal(t, []int64{0, 2}, Seq(0, 3, 2))
	assert.Equal(t, []int64{0, 2, 4}, Seq(0, 4, 2))
	assert.Equal(t, []int64{0, -5, -10}, Seq(0, -10, -5))
	assert.Equal(t, []int64{4, 3, 2, 1}, Seq(4, 1, 1))
	assert.Equal(t, []int64{-2, -1, 0}, Seq(-2, 0, 1))
	assert.Equal(t, []int64{-1, 0, 1}, Seq(-1, 1, 1))
	assert.Equal(t, []int64{-1, 0, 1}, Seq(-1, 1, -1))
	assert.Equal(t, []int64{1, 0, -1}, Seq(1, -1, 1))
	assert.Equal(t, []int64{1, 0, -1}, Seq(1, -1, -1))
	assert.Equal(t, []int64{}, Seq(1, -1, 0))
	assert.Equal(t, []int64{1}, Seq(1, 10000, 10000))
	assert.Equal(t, []int64{1, 0, -1}, Seq(1, -1, -1))
}
