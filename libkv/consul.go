package libkv

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"

	// XXX: replace once https://github.com/go-yaml/yaml/issues/139 is solved
	yaml "gopkg.in/hairyhenderson/yaml.v2"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/hairyhenderson/gomplate/vault"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	http  = "http"
	https = "https"
)

// NewConsul - instantiate a new Consul datasource handler
func NewConsul(u *url.URL) (*LibKV, error) {
	consul.Register()
	c, err := consulURL(u)
	if err != nil {
		return nil, err
	}
	config, err := consulConfig(c.Scheme == https)
	if err != nil {
		return nil, err
	}
	if role := env.Getenv("CONSUL_VAULT_ROLE", ""); role != "" {
		mount := env.Getenv("CONSUL_VAULT_MOUNT", "consul")

		var client *vault.Vault
		client, err = vault.New(nil)
		if err != nil {
			return nil, err
		}
		err = client.Login()
		defer client.Logout()
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("%s/creds/%s", mount, role)

		var data []byte
		data, err = client.Read(path)
		if err != nil {
			return nil, errors.Wrapf(err, "vault consul auth failed")
		}

		decoded := make(map[string]interface{})
		err = yaml.Unmarshal(data, &decoded)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to unmarshal object")
		}

		token := decoded["token"].(string)

		// nolint: gosec
		_ = os.Setenv("CONSUL_HTTP_TOKEN", token)
	}
	var kv store.Store
	kv, err = libkv.NewStore(store.CONSUL, []string{c.String()}, config)
	if err != nil {
		return nil, errors.Wrapf(err, "Consul setup failed")
	}
	return &LibKV{kv}, nil
}

// -- converts a gomplate datasource URL into a usable Consul URL
func consulURL(u *url.URL) (*url.URL, error) {
	addrEnv := env.Getenv("CONSUL_HTTP_ADDR")
	c, err := url.Parse(addrEnv)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid URL '%s' in CONSUL_HTTP_ADDR", addrEnv)
	}
	if c.Scheme == "" {
		c.Scheme = u.Scheme
	}
	switch c.Scheme {
	case "consul+http", http:
		c.Scheme = http
	case "consul+https", https:
		c.Scheme = https
	case "consul":
		if conv.Bool(env.Getenv("CONSUL_HTTP_SSL")) {
			c.Scheme = https
		} else {
			c.Scheme = http
		}
	}

	if c.Host == "" && u.Host == "" {
		c.Host = "localhost:8500"
	} else if c.Host == "" {
		c.Host = u.Host
	}

	return c, nil
}

func consulConfig(useTLS bool) (*store.Config, error) {
	t := conv.MustAtoi(env.Getenv("CONSUL_TIMEOUT"))
	config := &store.Config{
		ConnectionTimeout: time.Duration(t) * time.Second,
	}
	if useTLS {
		tconf := setupTLS("CONSUL")
		var err error
		config.TLS, err = consulapi.SetupTLSConfig(tconf)
		if err != nil {
			return nil, errors.Wrapf(err, "TLS Config setup failed")
		}
	}
	return config, nil
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
		verify := conv.Bool(v)
		tlsConfig.InsecureSkipVerify = !verify
	}
	return tlsConfig
}
