package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
)

func readFile(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	if source.fs == nil {
		fsp := datafs.FSProviderFromContext(ctx)
		fsys, err := fsp.New(source.URL)
		if err != nil {
			return nil, fmt.Errorf("filesystem provider for %q unavailable: %w", source.URL, err)
		}
		source.fs = fsys
	}

	p := filepath.FromSlash(source.URL.Path)

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, err
		}

		if parsed.Path != "" {
			p = filepath.Join(p, parsed.Path)
		}

		// reset the media type - it may have been set by a parent dir read
		source.mediaType = ""
	}

	isDir := strings.HasSuffix(p, string(filepath.Separator))
	if strings.HasSuffix(p, string(filepath.Separator)) {
		p = p[:len(p)-1]
	}

	// make sure we can access the file
	i, err := fs.Stat(source.fs, p)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", p, err)
	}

	if isDir {
		source.mediaType = jsonArrayMimetype
		if i.IsDir() {
			return readFileDir(source, p)
		}
		return nil, fmt.Errorf("%s is not a directory", p)
	}

	b, err := fs.ReadFile(source.fs, p)
	if err != nil {
		return nil, fmt.Errorf("readFile %s: %w", p, err)
	}
	return b, nil
}

func readFileDir(source *Source, p string) ([]byte, error) {
	names, err := fs.ReadDir(source.fs, p)
	if err != nil {
		return nil, err
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
