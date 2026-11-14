package random

import (
	"math"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzStringRE(f *testing.F) {
	f.Add(1, "")
	f.Add(10, "[a-z]")
	f.Add(5, "[[:alpha:]]")
	f.Add(3, `\x01`)
	f.Add(0, ".")
	f.Add(-1, "")

	f.Fuzz(func(t *testing.T, count int, match string) {
		// keep the output small - long strings only test the allocator
		if count > 1000 {
			t.Skip()
		}

		// every pattern is new while fuzzing, so don't let the cache grow
		t.Cleanup(matchCache.Clear)

		out, err := StringRE(count, match)
		if err != nil {
			return
		}

		require.True(t, utf8.ValidString(out))
		assert.Equal(t, max(count, 0), utf8.RuneCountInString(out))
	})
}

func FuzzStringBounds(f *testing.F) {
	f.Add(1, int32('a'), int32('z'))
	f.Add(10, int32(' '), int32('~'))
	f.Add(1, int32(0), int32(31))
	f.Add(1, int32('z'), int32('a'))
	f.Add(-1, int32(' '), int32('~'))
	f.Add(1, int32(0), int32(math.MaxInt32))

	f.Fuzz(func(t *testing.T, count int, lower, upper int32) {
		if count > 1000 {
			t.Skip()
		}

		out, err := StringBounds(count, lower, upper)
		if err != nil {
			return
		}

		require.True(t, utf8.ValidString(out))
		for _, r := range out {
			assert.GreaterOrEqual(t, r, lower)
			assert.LessOrEqual(t, r, upper)
		}
	})
}
