package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMath(t *testing.T) {
	assert.Equal(t, int64(10), AddInt(1, 2, 3, 4))
	assert.Equal(t, int64(12), MulInt(3, 4, 1))
}
