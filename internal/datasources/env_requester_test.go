package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvRequest(t *testing.T) {
	content := []byte(`hello world`)
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	defer os.Unsetenv("HELLO_UNIVERSE")

	testdata := []struct {
		url string
	}{
		{"env:HELLO_WORLD"},
		{"env:/HELLO_WORLD"},
		{"env:///HELLO_WORLD"},
		{"env:HELLO_WORLD?foo=bar"},
		{"env:///HELLO_WORLD?foo=bar"},
	}

	ctx := context.Background()
	r := &envRequester{}

	for i, d := range testdata {
		t.Run(fmt.Sprintf("%s_%d", d.url, i), func(t *testing.T) {
			actual, err := r.Request(ctx, mustParseURL(d.url), nil)
			assert.NoError(t, err)
			if actual != nil {
				b, err := ioutil.ReadAll(actual.Body)
				assert.NoError(t, err)
				assert.Equal(t, content, b)
				assert.Equal(t, int64(len(content)), actual.ContentLength)
				assert.Equal(t, textMimetype, actual.ContentType)
			}
		})
	}

}
