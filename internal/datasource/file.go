package datasource

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/pkg/errors"
)

// File -
type File struct {
	fs afero.Fs
}

var _ Reader = (*File)(nil)

func (f *File) Read(ctx context.Context, u *url.URL, args ...string) (data *Data, err error) {
	if f.fs == nil {
		f.fs = afero.NewOsFs()
	}

	_, p, err := f.parseFileParams(u, args)
	if err != nil {
		return nil, err
	}

	// make sure we can access the file
	i, err := f.fs.Stat(p)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't stat %s", p)
	}

	data = newData(u, args)

	if strings.HasSuffix(p, string(filepath.Separator)) {
		if i.IsDir() {
			data.MType = jsonArrayMimetype
			data.Bytes, err = f.readFileDir(p)
			return data, err
		}
		return nil, errors.Errorf("%s is not a directory", p)
	}

	file, err := f.fs.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't open %s", p)
	}

	data.Bytes, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", p)
	}
	return data, nil
}

func (f *File) readFileDir(p string) ([]byte, error) {
	names, err := afero.ReadDir(f.fs, p)
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

func (f *File) parseFileParams(sourceURL *url.URL, args []string) (params map[string]interface{}, p string, err error) {
	p = filepath.FromSlash(sourceURL.Path)
	params = make(map[string]interface{})
	for key, val := range sourceURL.Query() {
		params[key] = strings.Join(val, " ")
	}

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, "", err
		}

		if parsed.Path != "" {
			p = filepath.Join(p, parsed.Path)
		}

		for key, val := range parsed.Query() {
			params[key] = strings.Join(val, " ")
		}
	}
	return params, p, nil
}
