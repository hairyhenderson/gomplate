package libkv

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	yaml "github.com/hairyhenderson/yaml"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/vault"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	http  = "http"
	https = "https"

	// environment variables which aren't used by the consul client
	consulVaultRoleEnv  = "CONSUL_VAULT_ROLE"
	consulVaultMountEnv = "CONSUL_VAULT_MOUNT"
	consulTimeoutEnv    = "CONSUL_TIMEOUT"
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

	token, err := consulTokenFromVault()
	if err != nil {
		return nil, fmt.Errorf("failed to set Consul Vault token: %w", err)
	}

	if token != "" {
		// set CONSUL_HTTP_TOKEN before creating the client
		// nolint: gosec
		_ = os.Setenv(consulapi.HTTPTokenEnvName, token)
	}

	kv, err := libkv.NewStore(store.CONSUL, []string{c.String()}, config)
	if err != nil {
		return nil, fmt.Errorf("consul setup failed: %w", err)
	}

	return &LibKV{kv}, nil
}

func consulTokenFromVault() (string, error) {
	role := env.Getenv(consulVaultRoleEnv)
	if role == "" {
		return "", nil
	}

	client, err := vault.New(nil)
	if err != nil {
		return "", err
	}

	err = client.Login()
	defer client.Logout()
	if err != nil {
		return "", err
	}

	mount := env.Getenv(consulVaultMountEnv, "consul")
	path := fmt.Sprintf("%s/creds/%s", mount, role)

	data, err := client.Read(path)
	if err != nil {
		return "", fmt.Errorf("vault auth failed: %w", err)
	}

	decoded := make(map[string]interface{})
	err = yaml.Unmarshal(data, &decoded)
	if err != nil {
		return "", fmt.Errorf("YAML unmarshal: %w", err)
	}

	token := decoded["token"].(string)

	return token, nil
}

// consulAddrFromEnv parses the given address as either a URL or a host:port
// pair. Given no schema, the URL will need to have a schema set separately.
func consulAddrFromEnv(addr string) (*url.URL, error) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) < 2 {
		// temporary schema so it parses correctly
		addr = "temp://" + addr
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "temp" {
		u.Scheme = ""
	}

	return u, nil
}

// consulURL gets the Consul URL from either the given URL or the
// CONSUL_HTTP_ADDR environment variable. The given URL takes precedence.
func consulURL(u *url.URL) (c *url.URL, err error) {
	if u.Host == "" {
		addrEnv := env.Getenv(consulapi.HTTPAddrEnvName)
		c, err = consulAddrFromEnv(addrEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", addrEnv, err)
		}

		if c.Scheme == "" {
			c.Scheme = u.Scheme
		}
	} else {
		// We don't want the full URL here, just the scheme and host
		c = &url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}
	}

	switch c.Scheme {
	case "consul+http", http:
		c.Scheme = http
	case "consul+https", https:
		c.Scheme = https
	case "consul":
		if conv.Bool(env.Getenv(consulapi.HTTPSSLEnvName)) {
			c.Scheme = https
		} else {
			c.Scheme = http
		}
	}

	if c.Host == "" && u.Host == "" {
		c.Host = "localhost:8500"
	}

	return c, nil
}

func consulConfig(useTLS bool) (*store.Config, error) {
	t := conv.MustAtoi(env.Getenv(consulTimeoutEnv))
	config := &store.Config{
		ConnectionTimeout: time.Duration(t) * time.Second,
	}

	if useTLS {
		tconf := setupTLS()

		var err error
		config.TLS, err = consulapi.SetupTLSConfig(tconf)
		if err != nil {
			return nil, fmt.Errorf("TLS config setup failed: %w", err)
		}
	}

	return config, nil
}

func setupTLS() *consulapi.TLSConfig {
	tlsConfig := consulapi.TLSConfig{
		Address:  env.Getenv(consulapi.HTTPTLSServerName),
		CAFile:   env.Getenv(consulapi.HTTPCAFile),
		CAPath:   env.Getenv(consulapi.HTTPCAPath),
		CertFile: env.Getenv(consulapi.HTTPClientCert),
		KeyFile:  env.Getenv(consulapi.HTTPClientKey),
	}

	if v := env.Getenv(consulapi.HTTPSSLVerifyEnvName); v != "" {
		verify := conv.Bool(v)
		tlsConfig.InsecureSkipVerify = !verify
	}
	return &tlsConfig
}
