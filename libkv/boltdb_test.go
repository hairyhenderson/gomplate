package libkv

import (
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
)

var spyLogFatalMsg string

func restoreLogFatal() {
	logFatal = log.Fatal
}

func mockLogFatal(args ...interface{}) {
	spyLogFatalMsg = (args[0]).(string)
	panic(spyLogFatalMsg)
}

func setupMockLogFatal() {
	logFatal = mockLogFatal
	spyLogFatalMsg = ""
}

func TestSetupBoltDB(t *testing.T) {
	defer restoreLogFatal()
	setupMockLogFatal()
	assert.Panics(t, func() {
		setupBoltDB("")
	})

	expectedConfig := &store.Config{Bucket: "foo"}
	actualConfig := setupBoltDB("foo")
	assert.Equal(t, expectedConfig, actualConfig)

	expectedConfig = &store.Config{
		Bucket:            "bar",
		ConnectionTimeout: 42 * time.Second,
	}
	os.Setenv("BOLTDB_TIMEOUT", "42")
	defer os.Unsetenv("BOLTDB_TIMEOUT")
	actualConfig = setupBoltDB("bar")
	assert.Equal(t, expectedConfig, actualConfig)

	expectedConfig = &store.Config{
		Bucket:            "bar",
		ConnectionTimeout: 42 * time.Second,
		PersistConnection: true,
	}
	os.Setenv("BOLTDB_PERSIST", "true")
	defer os.Unsetenv("BOLTDB_PERSIST")
	actualConfig = setupBoltDB("bar")
	assert.Equal(t, expectedConfig, actualConfig)
}

func TestNewBoltDB(t *testing.T) {
	u, _ := url.Parse("boltdb:///bolt.db")
	defer restoreLogFatal()
	setupMockLogFatal()
	assert.Panics(t, func() {
		NewBoltDB(u)
	})
}
