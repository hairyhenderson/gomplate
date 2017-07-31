package main

import (
	"log"
	"strings"

	"github.com/blang/vfs"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// LibKV -
type LibKV struct {
	store store.Store
	fs    vfs.Filesystem
}

// NewLibKV - instantiate a new
func NewLibKV(url string) *LibKV {
	env := &Env{}

	var sourceType store.Backend
	var client string

	sourceType = ""
	client = ""

	if strings.HasPrefix(url, "consul:") {
		consul.Register()
		sourceType = store.CONSUL
		client = env.Getenv("CONSUL_HTTP_ADDR", "localhost:8500")
	}
	if strings.HasPrefix(url, "etcd:") {
		etcd.Register()
		sourceType = store.ETCD
		client = env.Getenv("ETCD_ADDR", "localhost:2379")
	}
	if strings.HasPrefix(url, "zk:") {
		zookeeper.Register()
		sourceType = store.ZK
		client = env.Getenv("ZK_ADDR", "localhost:2181")
	}
	if strings.HasPrefix(url, "boltdb:") {
		boltdb.Register()
		sourceType = store.BOLTDB
		client = env.Getenv("BOLTDB_PATH", "")
	}

	kv, err := libkv.NewStore(
		sourceType,
		[]string{client},
		&store.Config{},
	)
	if err != nil {
		logFatal("Cannot create store", err)
	}

	return &LibKV{kv, nil}
}

// Login -
func (kv *LibKV) Login() error {
	return nil
}

// Logout -
func (kv *LibKV) Logout() {
}

// Read -
func (kv *LibKV) Read(path string) ([]byte, error) {
	data, err := kv.store.Get(path)
	if err != nil {
		return nil, err
	}

	return data.Value, nil
}
