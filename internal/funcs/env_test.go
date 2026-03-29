package funcs

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEnvFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			fmap := CreateEnvFuncs(ctx)
			actual := fmap["env"].(func() any)
			ns := actual().(*EnvFuncs)

			assert.Equal(t, ctx, ns.ctx)
			assert.NotNil(t, ns.Env())
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

func TestEnvEnv(t *testing.T) {
	t.Setenv("GOMPLATE_TEST_ENV", "testvalue")

	t.Run("should return a map with expected values", func(t *testing.T) {
		ef := &EnvFuncs{}

		fn := ef.Env()
		assert.Equal(t, "testvalue", fn["GOMPLATE_TEST_ENV"])
		assert.Equal(t, os.Getenv("USER"), fn["USER"])
	})

	t.Run("should return an empty map for missing keys", func(t *testing.T) {
		ef := &EnvFuncs{}
		_, ok := ef.Env()["NONEXISTENT_VAR_12345"]
		assert.False(t, ok)
	})

	t.Run("mutation should not affect subsequent calls", func(t *testing.T) {
		ef := &EnvFuncs{}
		ef.Env()["GOMPLATE_TEST_MUTATION"] = "modified"

		assert.NotContains(t, ef.Env(), "GOMPLATE_TEST_MUTATION")
	})

	t.Run("gets updated environment variables", func(t *testing.T) {
		ef := &EnvFuncs{}

		fn := ef.Env()
		assert.Equal(t, "testvalue", fn["GOMPLATE_TEST_ENV"])
		t.Setenv("GOMPLATE_TEST_ENV", "modified")
		t.Setenv("NEW_ENV_VAR", "newvalue")

		fn = ef.Env()
		assert.Equal(t, "modified", fn["GOMPLATE_TEST_ENV"])
		assert.Equal(t, "newvalue", fn["NEW_ENV_VAR"])
	})
}

func TestEnvHasEnv(t *testing.T) {
	t.Setenv("GOMPLATE_TEST_HASENV", "somevalue")

	ef := &EnvFuncs{}

	assert.True(t, ef.HasEnv("GOMPLATE_TEST_HASENV"))

	assert.False(t, ef.HasEnv("NONEXISTENT_VAR_67890"))

	t.Setenv("GOMPLATE_TEST_EMPTY", "")
	assert.True(t, ef.HasEnv("GOMPLATE_TEST_EMPTY"))
}
