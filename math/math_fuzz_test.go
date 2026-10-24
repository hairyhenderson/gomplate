package math

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzSeq(f *testing.F) {
	f.Add(int64(1), int64(10), int64(1))
	f.Add(int64(0), int64(-10), int64(-5))
	f.Add(int64(4), int64(1), int64(1))
	f.Add(int64(1), int64(-1), int64(0))
	f.Add(int64(-1), int64(math.MaxInt64), int64(math.MaxInt64))
	f.Add(int64(math.MinInt64), int64(math.MaxInt64), int64(math.MaxInt64))
	f.Add(int64(math.MaxInt64), int64(math.MinInt64), int64(math.MinInt64))
	f.Add(int64(math.MinInt64), int64(math.MaxInt64), int64(1))

	f.Fuzz(func(t *testing.T, start, end, step int64) {
		if step == 0 {
			assert.Empty(t, Seq(start, end, step))
			return
		}

		// work out the expected length with big.Int, so it can't overflow
		dist := new(big.Int).Sub(big.NewInt(end), big.NewInt(start))
		stride := new(big.Int).Abs(big.NewInt(step))
		n := new(big.Int).Quo(new(big.Int).Abs(dist), stride)
		n.Add(n, big.NewInt(1))

		if !n.IsInt64() || n.Int64() > math.MaxInt {
			assert.Panics(t, func() { Seq(start, end, step) })
			return
		}

		// keep the sequence small - long ones only test the allocator
		if n.Int64() > 10000 {
			t.Skip()
		}

		seq := Seq(start, end, step)
		require.Len(t, seq, int(n.Int64()))
		assert.Equal(t, start, seq[0])

		// each value is one stride further towards end
		if dist.Sign() < 0 {
			stride.Neg(stride)
		}
		for i := 1; i < len(seq); i++ {
			d := new(big.Int).Sub(big.NewInt(seq[i]), big.NewInt(seq[i-1]))
			require.Zero(t, d.Cmp(stride), "step from %d to %d", seq[i-1], seq[i])
		}

		// the last value isn't past end, and is less than a step short of it
		rest := new(big.Int).Sub(big.NewInt(end), big.NewInt(seq[len(seq)-1]))
		assert.GreaterOrEqual(t, rest.Sign()*dist.Sign(), 0, "last value %d is past %d", seq[len(seq)-1], end)
		assert.Negative(t, new(big.Int).Abs(rest).Cmp(new(big.Int).Abs(stride)), "last value %d is a whole step short of %d", seq[len(seq)-1], end)
	})
}
