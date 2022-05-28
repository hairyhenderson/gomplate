package data

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStdin(t *testing.T) {
	ctx := context.Background()

	ctx = ContextWithStdin(ctx, strings.NewReader("foo"))
	out, err := readStdin(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	ctx = ContextWithStdin(ctx, errorReader{})
	_, err = readStdin(ctx, nil)
	assert.Error(t, err)
}
