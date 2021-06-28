package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
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

	r := &fileRequester{}
	ctx := config.WithFileSystem(context.Background(), fs)

	_, err := r.Request(ctx, mustParseURL("file:///bogus"), nil)
	assert.Error(t, err)

	testdata := []struct {
		url       string
		mediatype string
		body      []byte
		length    int64
	}{
		{"file:///tmp/foo", textMimetype, content, cLen},
		{"file:///tmp/partial/foo.txt", textMimetype, content, cLen},
		{"file:///tmp/partial/", jsonArrayMimetype, dirListing, dirLen},
		{"file:///tmp/partial/?type=application/json", jsonMimetype, dirListing, dirLen},
		{"file:///tmp/partial/foo.txt?type=application/json", jsonMimetype, content, cLen},
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
