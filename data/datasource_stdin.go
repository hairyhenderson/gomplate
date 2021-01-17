package data

import (
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

// stdin - for overriding in tests
var stdin io.Reader

func readStdin(source *Source, args ...string) ([]byte, error) {
	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", stdin)
	}
	return b, nil
}
