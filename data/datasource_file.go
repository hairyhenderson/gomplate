package data

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/pkg/errors"
)

func readFile(source *Source, args ...string) ([]byte, error) {
	if source.fs == nil {
		source.fs = afero.NewOsFs()
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
	}

	// make sure we can access the file
	i, err := source.fs.Stat(p)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't stat %s", p)
	}

	if strings.HasSuffix(p, string(filepath.Separator)) {
		source.mediaType = jsonArrayMimetype
		if i.IsDir() {
			return readFileDir(source, p)
		}
		return nil, errors.Errorf("%s is not a directory", p)
	}

	f, err := source.fs.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't open %s", p)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", p)
	}
	return b, nil
}

func readFileDir(source *Source, p string) ([]byte, error) {
	names, err := afero.ReadDir(source.fs, p)
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
