package data

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

// stdin - for overriding in tests
var stdin io.Reader

func readStdin(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", stdin)
	}
	return b, nil
}
