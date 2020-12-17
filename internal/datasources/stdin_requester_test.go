package datasources

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStdin(t *testing.T) {
	ctx := context.Background()

	r := &stdinRequester{strings.NewReader("foo")}
	u := mustParseURL("stdin:///foo?type=foo/bar")

	out, err := r.Request(ctx, u, nil)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(out.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), b)
	assert.Equal(t, "foo/bar", out.ContentType)
}
