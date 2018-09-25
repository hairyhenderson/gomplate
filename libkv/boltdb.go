package libkv

import (
	"net/url"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/pkg/errors"
)

// NewBoltDB - initialize a new BoltDB datasource handler
func NewBoltDB(u *url.URL) (*LibKV, error) {
	boltdb.Register()

	config, err := setupBoltDB(u.Fragment)
	if err != nil {
		return nil, err
	}
	kv, err := libkv.NewStore(store.BOLTDB, []string{u.Path}, config)
	if err != nil {
		return nil, errors.Wrapf(err, "BoltDB store creation failed")
	}
	return &LibKV{kv}, nil
}

func setupBoltDB(bucket string) (*store.Config, error) {
	if bucket == "" {
		return nil, errors.New("missing bucket - must specify BoltDB bucket in URL fragment")
	}

	t := conv.MustParseInt(env.Getenv("BOLTDB_TIMEOUT"), 10, 16)
	return &store.Config{
		Bucket:            bucket,
		ConnectionTimeout: time.Duration(t) * time.Second,
		PersistConnection: conv.Bool(env.Getenv("BOLTDB_PERSIST")),
	}, nil
}
