package coll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	out, err := Index(map[string]interface{}{
		"foo": "bar", "baz": "qux",
	}, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", out)

	out, err = Index(map[string]interface{}{
		"foo": "bar", "baz": "qux", "quux": "corge",
	}, "foo", 2)
	assert.NoError(t, err)
	assert.Equal(t, byte('r'), out)

	out, err = Index([]interface{}{"foo", "bar", "baz"}, 2)
	assert.NoError(t, err)
	assert.Equal(t, "baz", out)

	out, err = Index([]interface{}{"foo", "bar", "baz"}, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, byte('z'), out)

	// error cases
	out, err = Index([]interface{}{"foo", "bar", "baz"}, 0, 1, 2)
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = Index(nil, 0)
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = Index("foo", nil)
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = Index(map[interface{}]string{nil: "foo", 2: "bar"}, "baz")
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = Index([]int{}, 0)
	assert.Error(t, err)
	assert.Nil(t, out)
}
