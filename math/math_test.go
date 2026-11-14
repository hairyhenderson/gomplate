package math

import (
	"math"
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

	// these used to overflow
	assert.Equal(t, []int64{-1, math.MaxInt64 - 1}, Seq(-1, math.MaxInt64, math.MaxInt64))
	assert.Equal(t, []int64{math.MinInt64, -1, math.MaxInt64 - 1}, Seq(math.MinInt64, math.MaxInt64, math.MaxInt64))
	assert.Equal(t, []int64{math.MaxInt64, 0, math.MinInt64 + 1}, Seq(math.MaxInt64, math.MinInt64, math.MaxInt64))
	assert.Equal(t, []int64{0, math.MinInt64}, Seq(0, math.MinInt64, math.MinInt64))
	assert.Equal(t, []int64{math.MinInt64, 0}, Seq(math.MinInt64, math.MaxInt64, math.MinInt64))
	assert.Equal(t, []int64{math.MaxInt64}, Seq(math.MaxInt64, math.MaxInt64, 1))

	assert.Panics(t, func() { Seq(math.MinInt64, math.MaxInt64, 1) })
}
