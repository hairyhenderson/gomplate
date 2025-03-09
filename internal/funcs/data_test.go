package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDataFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateDataFuncs(ctx)
			actual := fmap["data"].(func() interface{})

			assert.Equal(t, ctx, actual().(*DataFuncs).ctx)
		})
	}
}
