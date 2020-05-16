package datasource

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEnv(t *testing.T) {
	content := []byte(`hello world`)
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	defer os.Unsetenv("HELLO_UNIVERSE")

	ctx := context.Background()
	e := &Env{}

	testdata := []struct {
		url string
	}{
		{"env:HELLO_WORLD"},
		{"env:/HELLO_WORLD"},
		{"env:///HELLO_WORLD"},
		{"env:HELLO_WORLD?foo=bar"},
		{"env:///HELLO_WORLD?foo=bar"},
	}

	for _, d := range testdata {
		url := mustParseURL(d.url)
		actual, err := e.Read(ctx, url)
		assert.NoError(t, err)
		assert.Equal(t, content, actual.Bytes)
	}
}
