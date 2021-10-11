package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGCPFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateGCPFuncs(ctx)
			actual := fmap["gcp"].(func() interface{})

			assert.Same(t, ctx, actual().(*GcpFuncs).ctx)
		})
	}
}
