package vault

import (
	"os"
	"os/user"
	"path"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenAuthStrategy_FromEnvVar(t *testing.T) {
	token := "deadbeef"

	os.Setenv("VAULT_TOKEN", token)
	defer os.Unsetenv("VAULT_TOKEN")

	auth := NewTokenAuthStrategy()
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenAuthStrategy_FromFileGivenNoEnvVar(t *testing.T) {
	token := "deadbeef"
	u, err := user.Current()
	assert.NoError(t, err)

	fs := memfs.Create()
	err = vfs.MkdirAll(fs, u.HomeDir, 0777)
	assert.NoError(t, err)
	f, err := vfs.Create(fs, path.Join(u.HomeDir, ".vault-token"))
	assert.NoError(t, err)
	f.Write([]byte(token))

	auth := NewTokenAuthStrategy(fs)
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenAuthStrategy_NilGivenNoVarOrFile(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	assert.Nil(t, NewTokenAuthStrategy(memfs.Create()))
}

func TestGetToken_Token(t *testing.T) {
	expected := "foo"
	auth := &TokenAuthStrategy{expected}
	actual, err := auth.GetToken(nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
