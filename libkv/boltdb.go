package libkv

import (
	"net/url"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/hairyhenderson/gomplate/typeconv"
)

// NewBoltDB - initialize a new BoltDB datasource handler
func NewBoltDB(u *url.URL) *LibKV {
	boltdb.Register()

	config := setupBoltDB(u.Fragment)
	kv, err := libkv.NewStore(store.BOLTDB, []string{u.Path}, config)
	if err != nil {
		logFatal("BoltDB store creation failed", err)
	}
	return &LibKV{kv}
}

func setupBoltDB(bucket string) *store.Config {
	if bucket == "" {
		logFatal("missing bucket - must specify BoltDB bucket in URL fragment")
	}

	t := typeconv.MustParseInt(env.Getenv("BOLTDB_TIMEOUT"), 10, 16)
	return &store.Config{
		Bucket:            bucket,
		ConnectionTimeout: time.Duration(t) * time.Second,
		PersistConnection: typeconv.MustParseBool(env.Getenv("BOLTDB_PERSIST")),
	}
}
