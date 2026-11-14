package conv

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzStringConversions(f *testing.F) {
	f.Add("0")
	f.Add("1")
	f.Add("-1")
	f.Add("0x1F")
	f.Add("0o17")
	f.Add("1,000,000")
	f.Add("1e3")
	f.Add("1e400")
	f.Add("+Inf")
	f.Add("NaN")
	f.Add("true")

	f.Fuzz(func(_ *testing.T, in string) {
		// arbitrary input should give an error, not a panic
		_, _ = ToInt64(in)
		_, _ = ToInt(in)
		_, _ = ToFloat64(in)
		_ = ToBool(in)
	})
}

func FuzzToInt64RoundTrip(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(-1))
	f.Add(int64(math.MaxInt64))
	f.Add(int64(math.MinInt64))

	f.Fuzz(func(t *testing.T, n int64) {
		for _, base := range []int{10, 16} {
			s := strconv.FormatInt(n, base)
			if base == 16 {
				if n < 0 {
					s = "-0x" + s[1:]
				} else {
					s = "0x" + s
				}
			}

			out, err := ToInt64(s)
			require.NoError(t, err, s)
			assert.Equal(t, n, out, s)
		}
	})
}

func FuzzToFloat64RoundTrip(f *testing.F) {
	f.Add(0.0)
	f.Add(-1.5)
	f.Add(1e20)
	f.Add(math.MaxFloat64)
	f.Add(math.SmallestNonzeroFloat64)

	f.Fuzz(func(t *testing.T, n float64) {
		if math.IsNaN(n) {
			t.Skip()
		}

		s := strconv.FormatFloat(n, 'g', -1, 64)
		out, err := ToFloat64(s)
		require.NoError(t, err, s)
		assert.Equal(t, n, out, s) //nolint:testifylint // the round trip must be exact
	})
}
