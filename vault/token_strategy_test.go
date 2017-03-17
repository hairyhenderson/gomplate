package vault

import (
	"os"
	"path"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenStrategy_FromEnvVar(t *testing.T) {
	token := "deadbeef"

	os.Setenv("VAULT_TOKEN", token)
	defer os.Unsetenv("VAULT_TOKEN")

	auth := NewTokenStrategy()
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenStrategy_FromFileGivenNoEnvVar(t *testing.T) {
	token := "deadbeef"

	fs := memfs.Create()
	err := vfs.MkdirAll(fs, homeDir(), 0777)
	assert.NoError(t, err)
	f, err := vfs.Create(fs, path.Join(homeDir(), ".vault-token"))
	assert.NoError(t, err)
	f.Write([]byte(token))

	auth := NewTokenStrategy(fs)
	assert.Equal(t, token, auth.Token)
}

func TestNewTokenStrategy_NilGivenNoVarOrFile(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	assert.Nil(t, NewTokenStrategy(memfs.Create()))
}

func TestGetToken_Token(t *testing.T) {
	expected := "foo"
	auth := &TokenStrategy{expected}
	actual, err := auth.GetToken(nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRevokable_TokenStrategy(t *testing.T) {
	strat := &TokenStrategy{}
	assert.False(t, strat.Revokable())
}
