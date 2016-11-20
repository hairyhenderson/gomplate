package vault

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/blang/vfs"
)

// TokenAuthStrategy - a pass-through strategy for situations where we already
// have a Vault token.
type TokenAuthStrategy struct {
	Token string
}

// NewTokenAuthStrategy - Try to create a new TokenAuthStrategy. If we can't
// nil will be returned.
func NewTokenAuthStrategy(getenv GetenvFunc, fsOverrides ...vfs.Filesystem) (*TokenAuthStrategy, error) {
	var fs vfs.Filesystem
	if len(fsOverrides) == 0 {
		fs = vfs.OS()
	} else {
		fs = fsOverrides[0]
	}

	if token := getenv("VAULT_TOKEN"); token != "" {
		return &TokenAuthStrategy{token}, nil
	}
	token, err := getTokenFromFile(getenv, fs)
	if err != nil {
		return nil, err
	}
	return &TokenAuthStrategy{token}, nil
}

// GetToken - return the token
func (a *TokenAuthStrategy) GetToken(addr *url.URL) (string, error) {
	return a.Token, nil
}

func (a *TokenAuthStrategy) String() string {
	return fmt.Sprintf("token: %s", a.Token)
}

// Revokable -
func (a *TokenAuthStrategy) Revokable() bool {
	return false
}

func homeDir(getenv GetenvFunc) (string, error) {
	if home := getenv("HOME"); home != "" {
		return home, nil
	}
	if home := getenv("USERPROFILE"); home != "" {
		return home, nil
	}
	return "", fmt.Errorf("Cannot detect user home directory (HOME/USERPROFILE)!")
}

func getTokenFromFile(getenv GetenvFunc, fs vfs.Filesystem) (string, error) {
	home, err := homeDir(getenv)
	if err != nil {
		return "", err
	}
	f, err := fs.OpenFile(path.Join(home, ".vault-token"), os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
