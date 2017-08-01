package main

import (
	"log"
	"net/url"

	"github.com/blang/vfs"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/hairyhenderson/gomplate/env"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// LibKV -
type LibKV struct {
	store store.Store
	fs    vfs.Filesystem
}

// NewLibKV - instantiate a new
func NewLibKV(url *url.URL) *LibKV {
	var sourceType store.Backend
	var client string

	sourceType = ""
	client = ""

	if url.Scheme == "consul" {
		consul.Register()
		sourceType = store.CONSUL
		client = env.Getenv("CONSUL_HTTP_ADDR", "localhost:8500")
	}
	if url.Scheme == "etcd" {
		etcd.Register()
		sourceType = store.ETCD
		client = env.Getenv("ETCD_ADDR", "localhost:2379")
	}
	if url.Scheme == "zk" {
		zookeeper.Register()
		sourceType = store.ZK
		client = env.Getenv("ZK_ADDR", "localhost:2181")
	}
	if url.Scheme == "boltdb" {
		boltdb.Register()
		sourceType = store.BOLTDB
		client = url.Path
		if client == "" {
			client = env.Getenv("BOLTDB_PATH", "")
		}
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
