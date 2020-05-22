package data

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStdin(t *testing.T) {
	defer func() {
		stdin = nil
	}()
	stdin = strings.NewReader("foo")
	out, err := readStdin(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	stdin = errorReader{}
	_, err = readStdin(nil)
	assert.Error(t, err)
}
