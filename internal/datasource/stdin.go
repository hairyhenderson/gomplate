package datasource

import (
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

// Stdin -
type Stdin struct {
	// for overriding in tests
	in io.Reader
}

var _ Reader = (*Stdin)(nil)

func (s *Stdin) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	if s.in == nil {
		s.in = os.Stdin
	}

	data.Bytes, err = ioutil.ReadAll(s.in)
	if err != nil {
		return data, errors.Wrapf(err, "Can't read %s", s.in)
	}

	return data, nil
}
