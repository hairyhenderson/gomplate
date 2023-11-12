package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSockaddrFuncs(t *testing.T) {
	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateSockaddrFuncs(ctx)
			actual := fmap["sockaddr"].(func() interface{})

			assert.Equal(t, ctx, actual().(*SockaddrFuncs).ctx)
		})
	}
}
