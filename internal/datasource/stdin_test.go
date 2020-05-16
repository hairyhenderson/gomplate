package datasource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStdin(t *testing.T) {
	ctx := context.Background()
	s := &Stdin{strings.NewReader("foo")}

	out, err := s.Read(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, Data{Bytes: []byte("foo")}, out)

	s = &Stdin{errorReader{}}
	_, err = s.Read(ctx, nil)
	assert.Error(t, err)
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}
