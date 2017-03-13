package vault

import (
	"fmt"
	"io/ioutil"
	"log"
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
func NewTokenAuthStrategy(fsOverrides ...vfs.Filesystem) *TokenAuthStrategy {
	var fs vfs.Filesystem
	if len(fsOverrides) == 0 {
		fs = vfs.OS()
	} else {
		fs = fsOverrides[0]
	}

	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		return &TokenAuthStrategy{token}
	}
	if token := getTokenFromFile(fs); token != "" {
		return &TokenAuthStrategy{token}
	}
	return nil
}

// GetToken - return the token
func (a *TokenAuthStrategy) GetToken(addr *url.URL) (string, error) {
	return a.Token, nil
}

func (a *TokenAuthStrategy) String() string {
	return fmt.Sprintf("token: %s", a.Token)
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

// RevokeToken - no-op for this strategy since we didn't request this token in-band
func (a *TokenAuthStrategy) RevokeToken(addr *url.URL) error {
	return nil
}
