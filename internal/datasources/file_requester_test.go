package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestFileRequest(t *testing.T) {
	content := []byte(`hello world`)
	cLen := int64(len(content))
	dirListing := []byte(`["bar.txt","baz.txt","foo.txt"]`)
	dirLen := int64(len(dirListing))

	fs := afero.NewMemMapFs()

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write(content)

	_ = fs.Mkdir("/tmp/partial", 0777)
	f, _ = fs.Create("/tmp/partial/foo.txt")
	_, _ = f.Write(content)
	_, _ = fs.Create("/tmp/partial/bar.txt")
	_, _ = fs.Create("/tmp/partial/baz.txt")

	ctx := context.Background()
	r := &fileRequester{fs}

	_, err := r.Request(ctx, mustParseURL("file:///bogus"), nil)
	assert.Error(t, err)

	testdata := []struct {
		url       string
		mediatype string
		length    int64
		body      []byte
	}{
		{"file:///tmp/foo", textMimetype, cLen, content},
		{"file:///tmp/partial/foo.txt", textMimetype, cLen, content},
		{"file:///tmp/partial/", jsonArrayMimetype, dirLen, dirListing},
		{"file:///tmp/partial/?type=application/json", jsonMimetype, dirLen, dirListing},
		{"file:///tmp/partial/foo.txt?type=application/json", jsonMimetype, cLen, content},
	}

	for i, d := range testdata {
		t.Run(fmt.Sprintf("%s_%d", d.url, i), func(t *testing.T) {
			actual, err := r.Request(ctx, mustParseURL(d.url), nil)
			assert.NoError(t, err)
			if actual != nil {
				b, err := ioutil.ReadAll(actual.Body)
				assert.NoError(t, err)
				assert.Equal(t, d.body, b)
				assert.Equal(t, d.length, actual.ContentLength)
				assert.Equal(t, d.mediatype, actual.ContentType)
			}
		})
	}
}
