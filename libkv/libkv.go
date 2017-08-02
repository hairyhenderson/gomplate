package libkv

import (
	"crypto/tls"
	"errors"
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
	consulapi "github.com/hashicorp/consul/api"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// LibKV -
type LibKV struct {
	store store.Store
	fs    vfs.Filesystem
}

type SetupDetails struct {
	sourceType store.Backend
	client     string
	options    *store.Config
}

// New - instantiate a new
func New(url *url.URL) *LibKV {
	var s *SetupDetails

	if url.Scheme == "consul" {
		setup, err := setupConsul(url)
		if err != nil {
			logFatal("consul setup error", err)
		}
		s = setup
	}
	if url.Scheme == "etcd" {
		setup, err := setupEtcd(url)
		if err != nil {
			logFatal("etcd setup error", err)
		}
		s = setup
	}
	if url.Scheme == "zk" {
		setup, err := setupZookeeper(url)
		if err != nil {
			logFatal("zk setup error", err)
		}
		s = setup
	}
	if url.Scheme == "boltdb" {
		setup, err := setupBoltDB(url)
		if err != nil {
			logFatal("boltdb setup error", err)
		}
		s = setup
	}

	if s.client == "" {
		logFatal("missing client location")
	}

	kv, err := libkv.NewStore(
		s.sourceType,
		[]string{s.client},
		s.options,
	)
	if err != nil {
		logFatal("Cannot create store", err)
	}

	return &LibKV{kv, nil}
}

func setupConsul(url *url.URL) (*SetupDetails, error) {
	setup := &SetupDetails{}
	consul.Register()
	setup.sourceType = store.CONSUL
	setup.client = env.Getenv("CONSUL_HTTP_ADDR", "localhost:8500")
	setup.options = &store.Config{}
	if timeout := env.Getenv("CONSUL_TIMEOUT", ""); timeout != "" {
		num, err := strconv.ParseInt(timeout, 10, 16)
		if err != nil {
			return nil, err
		}
		setup.options.ConnectionTimeout = time.Duration(num) * time.Second
	}
	if ssl := env.Getenv("CONSUL_HTTP_SSL", ""); ssl != "" {
		enabled, err := strconv.ParseBool(ssl)
		if err != nil {
			return nil, err
		}
		if enabled {
			config, err := setupTLS("CONSUL")
			if err != nil {
				return nil, err
			}
			setup.options.TLS = config
		}
	}
	return setup, nil
}

func setupEtcd(url *url.URL) (*SetupDetails, error) {
	setup := &SetupDetails{}
	etcd.Register()
	setup.sourceType = store.ETCD
	setup.client = env.Getenv("ETCD_ADDR", "localhost:2379")
	setup.options = &store.Config{}
	if timeout := env.Getenv("ETCD_TIMEOUT", ""); timeout != "" {
		num, err := strconv.ParseInt(timeout, 10, 16)
		if err != nil {
			return nil, err
		}
		setup.options.ConnectionTimeout = time.Duration(num) * time.Second
	}
	if ssl := env.Getenv("ETCD_TLS", ""); ssl != "" {
		enabled, err := strconv.ParseBool(ssl)
		if err != nil {
			return nil, err
		}
		if enabled {
			config, err := setupTLS("ETCD")
			if err != nil {
				return nil, err
			}
			setup.options.TLS = config
		}
	}
	setup.options.Username = env.Getenv("ETCD_USERNAME", "")
	setup.options.Password = env.Getenv("ETCD_PASSWORD", "")
	return setup, nil
}

func setupZookeeper(url *url.URL) (*SetupDetails, error) {
	setup := &SetupDetails{}
	zookeeper.Register()
	setup.sourceType = store.ZK
	setup.client = env.Getenv("ZK_ADDR", "localhost:2181")
	setup.options = &store.Config{}
	if timeout := env.Getenv("ZK_TIMEOUT", ""); timeout != "" {
		num, err := strconv.ParseInt(timeout, 10, 16)
		if err != nil {
			return nil, err
		}
		setup.options.ConnectionTimeout = time.Duration(num) * time.Second
	}
	return setup, nil
}

func setupBoltDB(url *url.URL) (*SetupDetails, error) {
	setup := &SetupDetails{}
	boltdb.Register()
	setup.sourceType = store.BOLTDB
	setup.client = url.Path
	setup.options = &store.Config{}
	setup.options.Bucket = url.Fragment
	if setup.options.Bucket == "" {
		return nil, errors.New("missing bucket")
	}
	if timeout := env.Getenv("BOLTDB_TIMEOUT", ""); timeout != "" {
		num, err := strconv.ParseInt(timeout, 10, 16)
		if err != nil {
			return nil, err
		}
		setup.options.ConnectionTimeout = time.Duration(num) * time.Second
	}
	if persist := env.Getenv("BOLTDB_PERSIST", ""); persist != "" {
		enabled, err := strconv.ParseBool(persist)
		if err != nil {
			return nil, err
		}
		setup.options.PersistConnection = enabled
	}
	return setup, nil
}

func setupTLS(prefix string) (*tls.Config, error) {
	tlsConfig := &consulapi.TLSConfig{}

	if v := env.Getenv(prefix+"_TLS_SERVER_NAME", ""); v != "" {
		tlsConfig.Address = v
	}
	if v := env.Getenv(prefix+"_CACERT", ""); v != "" {
		tlsConfig.CAFile = v
	}
	if v := env.Getenv(prefix+"_CAPATH", ""); v != "" {
		tlsConfig.CAPath = v
	}
	if v := env.Getenv(prefix+"_CLIENT_CERT", ""); v != "" {
		tlsConfig.CertFile = v
	}
	if v := env.Getenv(prefix+"_CLIENT_KEY", ""); v != "" {
		tlsConfig.KeyFile = v
	}
	if v := env.Getenv(prefix+"_HTTP_SSL_VERIFY", ""); v != "" {
		verify, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		if !verify {
			tlsConfig.InsecureSkipVerify = true
		}
	}

	config, err := consulapi.SetupTLSConfig(tlsConfig)
	if err != nil {
		return nil, err
	}

	return config, nil
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
