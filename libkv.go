package main

import (
	"crypto/tls"
	"log"
	"net/url"
	"strconv"
	"time"

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
	options := store.Config{}

	if url.Scheme == "consul" {
		consul.Register()
		sourceType = store.CONSUL
		client = env.Getenv("CONSUL_HTTP_ADDR", "localhost:8500")
		if timeout := env.Getenv("CONSUL_TIMEOUT", ""); timeout != "" {
			num, err := strconv.ParseInt(timeout, 10, 16)
			if err != nil {
				logFatal("consul timeout error", err)
			}
			options.ConnectionTimeout = time.Duration(num) * time.Second
		}
		if ssl := env.Getenv("CONSUL_HTTP_SSL", ""); ssl != "" {
			enabled, err := strconv.ParseBool(ssl)
			if err != nil {
				logFatal("consul ssl error", err)
			}
			if enabled {
				options.TLS = &tls.Config{}
			}
		}
	}
	if url.Scheme == "etcd" {
		etcd.Register()
		sourceType = store.ETCD
		client = env.Getenv("ETCD_ADDR", "localhost:2379")
		if timeout := env.Getenv("ETCD_TIMEOUT", ""); timeout != "" {
			num, err := strconv.ParseInt(timeout, 10, 16)
			if err != nil {
				logFatal("etcd timeout error", err)
			}
			options.ConnectionTimeout = time.Duration(num) * time.Second
		}
		if ssl := env.Getenv("ETCD_TLS", ""); ssl != "" {
			enabled, err := strconv.ParseBool(ssl)
			if err != nil {
				logFatal("consul ssl error", err)
			}
			if enabled {
				options.TLS = &tls.Config{}
			}
		}
		options.Username = env.Getenv("ETCD_USERNAME", "")
		options.Password = env.Getenv("ETCD_PASSWORD", "")
	}
	if url.Scheme == "zk" {
		zookeeper.Register()
		sourceType = store.ZK
		client = env.Getenv("ZK_ADDR", "localhost:2181")
		if timeout := env.Getenv("ZK_TIMEOUT", ""); timeout != "" {
			num, err := strconv.ParseInt(timeout, 10, 16)
			if err != nil {
				logFatal("zk timeout error", err)
			}
			options.ConnectionTimeout = time.Duration(num) * time.Second
		}
	}
	if url.Scheme == "boltdb" {
		boltdb.Register()
		sourceType = store.BOLTDB
		client = url.Path
		if client == "" {
			client = env.Getenv("BOLTDB_DATABASE", "")
		}
		if url.Fragment != "" {
			options.Bucket = url.Fragment
		}
		if options.Bucket == "" {
			options.Bucket = env.Getenv("BOLTDB_BUCKET", "")
		}
		if options.Bucket == "" {
			logFatal("boltdb missing bucket")
		}
		if timeout := env.Getenv("BOLTDB_TIMEOUT", ""); timeout != "" {
			num, err := strconv.ParseInt(timeout, 10, 16)
			if err != nil {
				logFatal("boltdb timeout error", err)
			}
			options.ConnectionTimeout = time.Duration(num) * time.Second
		}
		if persist := env.Getenv("BOLTDB_PERSIST", ""); persist != "" {
			enabled, err := strconv.ParseBool(persist)
			if err != nil {
				logFatal("boltdb persist error", err)
			}
			options.PersistConnection = enabled
		}
	}

	if client == "" {
		logFatal("missing client location")
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
