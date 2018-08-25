package libkv

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
)

func TestSetupBoltDB(t *testing.T) {
	_, err := setupBoltDB("")
	assert.Error(t, err)

	expectedConfig := &store.Config{Bucket: "foo"}
	actualConfig, err := setupBoltDB("foo")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, actualConfig)

	expectedConfig = &store.Config{
		Bucket:            "bar",
		ConnectionTimeout: 42 * time.Second,
	}
	os.Setenv("BOLTDB_TIMEOUT", "42")
	defer os.Unsetenv("BOLTDB_TIMEOUT")
	actualConfig, err = setupBoltDB("bar")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, actualConfig)

	expectedConfig = &store.Config{
		Bucket:            "bar",
		ConnectionTimeout: 42 * time.Second,
		PersistConnection: true,
	}
	os.Setenv("BOLTDB_PERSIST", "true")
	defer os.Unsetenv("BOLTDB_PERSIST")
	actualConfig, err = setupBoltDB("bar")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, actualConfig)
}

func TestNewBoltDB(t *testing.T) {
	u, _ := url.Parse("boltdb:///bolt.db")
	_, err := NewBoltDB(u)
	assert.Error(t, err)
}
