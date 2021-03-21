package datasources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"
)

type fileRequester struct {
}

func (r *fileRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	fs := config.FileSystemFromContext(ctx)
	if fs == nil {
		fs = afero.NewOsFs()
	}

	p := filepath.FromSlash(u.Path)
	// make sure we can access the file
	i, err := fs.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %q: %w", p, err)
	}

	resp := &Response{ContentLength: i.Size()}

	// paths that explicitly end with "/" should be directories
	if strings.HasSuffix(p, string(filepath.Separator)) && !i.IsDir() {
		return nil, fmt.Errorf("failed to read %q: not a directory", p)
	}

	hint := ""
	if i.IsDir() {
		hint = jsonArrayMimetype
		var b []byte
		b, err = r.readFileDir(ctx, fs, p)
		if err != nil {
			return nil, fmt.Errorf("failed to list directory %q: %w", p, err)
		}
		resp.ContentLength = int64(len(b))
		resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	} else {
		resp.Body, err = fs.OpenFile(p, os.O_RDONLY, 0)
		if err != nil {
			return nil, fmt.Errorf("can't open %q: %w", p, err)
		}
		resp.ContentLength = i.Size()
	}
	resp.ContentType, err = mimeType(u, hint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *fileRequester) readFileDir(ctx context.Context, fs afero.Fs, p string) ([]byte, error) {
	names, err := afero.ReadDir(fs, p)
	if err != nil {
		return nil, fmt.Errorf("readDir failed: %w", err)
	}
	files := make([]string, len(names))
	for i, v := range names {
		files[i] = v.Name()
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(files); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	// chop off the newline added by the json encoder
	return b[:len(b)-1], nil
}
