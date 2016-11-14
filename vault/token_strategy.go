package vault

import (
	"fmt"
	"github.com/blang/vfs"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
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

// Revokable -
func (a *TokenAuthStrategy) Revokable() bool {
	return false
}

func getTokenFromFile(fs vfs.Filesystem) string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	f, err := fs.OpenFile(path.Join(u.HomeDir, ".vault-token"), os.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}
	return string(b)
}
