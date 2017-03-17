package vault

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/blang/vfs"
)

// TokenStrategy - a pass-through strategy for situations where we already
// have a Vault token.
type TokenStrategy struct {
	Token string
}

// NewTokenStrategy - Try to create a new TokenStrategy. If we can't
// nil will be returned.
func NewTokenStrategy(fsOverrides ...vfs.Filesystem) *TokenStrategy {
	var fs vfs.Filesystem
	if len(fsOverrides) == 0 {
		fs = vfs.OS()
	} else {
		fs = fsOverrides[0]
	}

	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		return &TokenStrategy{token}
	}
	if token := getTokenFromFile(fs); token != "" {
		return &TokenStrategy{token}
	}
	return nil
}

// GetToken - return the token
func (a *TokenStrategy) GetToken(addr *url.URL) (string, error) {
	return a.Token, nil
}

// Revokable -
func (a *TokenStrategy) Revokable() bool {
	return false
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	log.Fatal(`Neither HOME nor USERPROFILE environment variables are set!
		I can't figure out where the current user's home directory is!`)
	return ""
}

func getTokenFromFile(fs vfs.Filesystem) string {
	f, err := fs.OpenFile(path.Join(homeDir(), ".vault-token"), os.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(b)
}
