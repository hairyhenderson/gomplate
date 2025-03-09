package coll

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	out, err := Index(map[string]any{
		"foo": "bar", "baz": "qux",
	}, "foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", out)

	out, err = Index(map[string]any{
		"foo": "bar", "baz": "qux", "quux": "corge",
	}, "foo", 2)
	require.NoError(t, err)
	assert.Equal(t, byte('r'), out)

	out, err = Index([]any{"foo", "bar", "baz"}, 2)
	require.NoError(t, err)
	assert.Equal(t, "baz", out)

	out, err = Index([]any{"foo", "bar", "baz"}, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, byte('z'), out)

	// error cases
	out, err = Index([]any{"foo", "bar", "baz"}, 0, 1, 2)
	require.Error(t, err)
	assert.Nil(t, out)

	out, err = Index(nil, 0)
	require.Error(t, err)
	assert.Nil(t, out)

	out, err = Index("foo", nil)
	require.Error(t, err)
	assert.Nil(t, out)

	out, err = Index(map[any]string{nil: "foo", 2: "bar"}, "baz")
	require.Error(t, err)
	assert.Nil(t, out)

	out, err = Index([]int{}, 0)
	require.Error(t, err)
	assert.Nil(t, out)
}
