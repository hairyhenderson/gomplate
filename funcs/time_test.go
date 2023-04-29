package funcs

import (
	"context"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTimeFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateTimeFuncs(ctx)
			actual := fmap["time"].(func() interface{})

			assert.Same(t, ctx, actual().(*TimeFuncs).ctx)
		})
	}
}

func TestParseNum(t *testing.T) {
	t.Parallel()

	i, f, _ := parseNum("42")
	assert.Equal(t, int64(42), i)
	assert.Equal(t, int64(0), f)

	i, f, _ = parseNum(42)
	assert.Equal(t, int64(42), i)
	assert.Equal(t, int64(0), f)

	i, f, _ = parseNum(big.NewInt(42))
	assert.Equal(t, int64(42), i)
	assert.Equal(t, int64(0), f)

	i, f, _ = parseNum(big.NewFloat(42.0))
	assert.Equal(t, int64(42), i)
	assert.Equal(t, int64(0), f)

	i, f, _ = parseNum(uint64(math.MaxInt64))
	assert.Equal(t, int64(uint64(math.MaxInt64)), i)
	assert.Equal(t, int64(0), f)

	i, f, _ = parseNum("9223372036854775807.999999999")
	assert.Equal(t, int64(9223372036854775807), i)
	assert.Equal(t, int64(999999999), f)

	i, f, _ = parseNum("999999999999999.123456789123")
	assert.Equal(t, int64(999999999999999), i)
	assert.Equal(t, int64(123456789), f)

	i, f, _ = parseNum("123456.789")
	assert.Equal(t, int64(123456), i)
	assert.Equal(t, int64(789000000), f)

	_, _, err := parseNum("bogus.9223372036854775807")
	assert.Error(t, err)

	_, _, err = parseNum("bogus")
	assert.Error(t, err)

	_, _, err = parseNum("1.2.3")
	assert.Error(t, err)

	_, _, err = parseNum(1.1)
	assert.Error(t, err)

	i, f, err = parseNum(nil)
	assert.Zero(t, i)
	assert.Zero(t, f)
	require.NoError(t, err)
}
