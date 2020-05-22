package data

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

func readStdin(source *Source, args ...string) ([]byte, error) {
	if stdin == nil {
		stdin = os.Stdin
	}
	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", stdin)
	}
	return b, nil
}
