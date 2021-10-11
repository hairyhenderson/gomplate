package funcs

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEnvFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateEnvFuncs(ctx)
			actual := fmap["env"].(func() interface{})

			assert.Same(t, ctx, actual().(*EnvFuncs).ctx)
		})
	}
}

func TestEnvGetenv(t *testing.T) {
	t.Parallel()

	ef := &EnvFuncs{}
	expected := os.Getenv("USER")
	assert.Equal(t, expected, ef.Getenv("USER"))

	assert.Equal(t, "foo", ef.Getenv("bogusenvvar", "foo"))
}
