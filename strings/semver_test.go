package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemverCompare(t *testing.T) {
	v, err := SemverCompare("1.2.3", "1.2.3")
	assert.NoError(t, err)
	assert.Equal(t, true, v)
}
