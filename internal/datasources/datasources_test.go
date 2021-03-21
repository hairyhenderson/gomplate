package datasources

import (
	"context"
	"net/url"
	"runtime"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadDataSource(t *testing.T) {
	contents := "hello world"
	fname := "foo.txt"
	fs := afero.NewMemMapFs()

	ctx := config.WithFileSystem(context.Background(), fs)

	var uPath string
	var f afero.File
	if runtime.GOOS == "windows" {
		_ = fs.Mkdir("C:\\tmp", 0777)
		f, _ = fs.Create("C:\\tmp\\" + fname)
		uPath = "C:/tmp/" + fname
	} else {
		_ = fs.Mkdir("/tmp", 0777)
		f, _ = fs.Create("/tmp/" + fname)
		uPath = "/tmp/" + fname
	}
	_, _ = f.Write([]byte(contents))

	ds := config.DataSource{
		URL: &url.URL{Scheme: "file", Path: uPath},
	}

	ct, b, err := ReadDataSource(ctx, ds)
	assert.NoError(t, err)
	assert.Equal(t, contents, string(b))
	assert.Equal(t, textMimetype, ct)
}
