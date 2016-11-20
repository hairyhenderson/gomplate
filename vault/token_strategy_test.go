package vault

import (
	"os"
	"path"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func getEnvStub(env string) string {
	return ""
}

func TestNewTokenAuthStrategy_FromEnvVar(t *testing.T) {
	token := "deadbeef"
	auth, err := NewTokenAuthStrategy(func(env string) string {
		if env == "VAULT_TOKEN" {
			return token
		}
		return ""
	})
	assert.Nil(t, err)
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenAuthStrategy_FromFileGivenNoEnvVar(t *testing.T) {
	token := "deadbeef"
	fs := memfs.Create()
	home, _ := homeDir(os.Getenv)
	err := vfs.MkdirAll(fs, home, 0777)
	assert.NoError(t, err)
	f, err := vfs.Create(fs, path.Join(home, ".vault-token"))
	assert.NoError(t, err)
	f.Write([]byte(token))

	auth, err := NewTokenAuthStrategy(os.Getenv, fs)
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenAuthStrategy_NilGivenNoVarOrFile(t *testing.T) {
	auth, err := NewTokenAuthStrategy(getEnvStub, memfs.Create())
	assert.Error(t, err)
	assert.Nil(t, auth)
}

func TestGetToken_Token(t *testing.T) {
	expected := "foo"
	auth := &TokenAuthStrategy{expected}
	actual, err := auth.GetToken(nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
