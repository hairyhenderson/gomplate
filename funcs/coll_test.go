package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	c := CollNS()

	_, err := c.Flatten()
	assert.Error(t, err)

	_, err = c.Flatten(42)
	assert.Error(t, err)

	out, err := c.Flatten([]interface{}{1, []interface{}{[]int{2}, 3}})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, 2, 3}, out)

	out, err = c.Flatten(1, []interface{}{1, []interface{}{[]int{2}, 3}})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, []int{2}, 3}, out)
}
