package data

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStdin(t *testing.T) {
	ctx := context.Background()

	defer func() {
		stdin = nil
	}()
	stdin = strings.NewReader("foo")
	out, err := readStdin(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	stdin = errorReader{}
	_, err = readStdin(ctx, nil)
	assert.Error(t, err)
}
