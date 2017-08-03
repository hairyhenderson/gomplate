package libkv

import (
	"net/url"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/hairyhenderson/gomplate/env"
	consulapi "github.com/hashicorp/consul/api"
)

// NewConsul - instantiate a new Consul datasource handler
func NewConsul(u *url.URL) *LibKV {
	consul.Register()
	c := consulURL(u)
	config := consulConfig(c.Scheme == "https")
	kv, err := libkv.NewStore(store.CONSUL, []string{c.String()}, config)
	if err != nil {
		logFatal("Consul setup failed", err)
	}
	return &LibKV{kv}
}

// -- converts a gomplate datasource URL into a usable Consul URL
func consulURL(u *url.URL) *url.URL {
	c, _ := url.Parse(env.Getenv("CONSUL_HTTP_ADDR"))
	if c.Scheme == "" {
		c.Scheme = u.Scheme
	}
	switch c.Scheme {
	case "consul+http", "http":
		c.Scheme = "http"
	case "consul+https", "https":
		c.Scheme = "https"
	case "consul":
		if mustParseBool(env.Getenv("CONSUL_HTTP_SSL")) {
			c.Scheme = "https"
		} else {
			c.Scheme = "http"
		}
	}

	if c.Host == "" && u.Host == "" {
		c.Host = "localhost:8500"
	} else if c.Host == "" {
		c.Host = u.Host
	}

	return c
}

func consulConfig(useTLS bool) *store.Config {
	t := mustParseInt(env.Getenv("CONSUL_TIMEOUT"))
	config := &store.Config{
		ConnectionTimeout: time.Duration(t) * time.Second,
	}
	if useTLS {
		tconf := setupTLS("CONSUL")
		var err error
		config.TLS, err = consulapi.SetupTLSConfig(tconf)
		if err != nil {
			logFatal("TLS Config setup failed", err)
		}
	}
	return config
}

func setupTLS(prefix string) *consulapi.TLSConfig {
	tlsConfig := &consulapi.TLSConfig{
		Address:  env.Getenv(prefix + "_TLS_SERVER_NAME"),
		CAFile:   env.Getenv(prefix + "_CACERT"),
		CAPath:   env.Getenv(prefix + "_CAPATH"),
		CertFile: env.Getenv(prefix + "_CLIENT_CERT"),
		KeyFile:  env.Getenv(prefix + "_CLIENT_KEY"),
	}
	if v := env.Getenv(prefix + "_HTTP_SSL_VERIFY"); v != "" {
		verify := mustParseBool(v)
		tlsConfig.InsecureSkipVerify = !verify
	}
	return tlsConfig
}
