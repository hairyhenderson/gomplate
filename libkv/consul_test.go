package libkv

import (
	"crypto/tls"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/docker/libkv/store"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsulURL(t *testing.T) {
	os.Unsetenv("CONSUL_HTTP_SSL")

	t.Run("consul scheme, CONSUL_HTTP_SSL set to true", func(t *testing.T) {
		t.Setenv("CONSUL_HTTP_SSL", "true")

		u, _ := url.Parse("consul://")
		expected := &url.URL{Host: "localhost:8500", Scheme: "https"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("consul+http scheme", func(t *testing.T) {
		u, _ := url.Parse("consul+http://myconsul.server")
		expected := &url.URL{Host: "myconsul.server", Scheme: "http"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("consul+https scheme, CONSUL_HTTP_SSL set to false", func(t *testing.T) {
		t.Setenv("CONSUL_HTTP_SSL", "false")

		u, _ := url.Parse("consul+https://myconsul.server:1234")
		expected := &url.URL{Host: "myconsul.server:1234", Scheme: "https"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("consul scheme, CONSUL_HTTP_SSL unset", func(t *testing.T) {
		u, _ := url.Parse("consul://myconsul.server:2345")
		expected := &url.URL{Host: "myconsul.server:2345", Scheme: "http"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("consul scheme, ignore path", func(t *testing.T) {
		u, _ := url.Parse("consul://myconsul.server:3456/foo/bar/baz")
		expected := &url.URL{Host: "myconsul.server:3456", Scheme: "http"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("given URL takes precedence over env var", func(t *testing.T) {
		t.Setenv("CONSUL_HTTP_ADDR", "https://foo:8500")

		u, _ := url.Parse("consul://myconsul.server:3456/foo/bar/baz")
		expected := &url.URL{Host: "myconsul.server:3456", Scheme: "http"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("TLS enabled, HTTP_ADDR is set, URL has no host and ambiguous scheme", func(t *testing.T) {
		t.Setenv("CONSUL_HTTP_ADDR", "https://foo:8500")
		t.Setenv("CONSUL_HTTP_SSL", "true")

		u, _ := url.Parse("consul://")
		expected := &url.URL{Host: "foo:8500", Scheme: "https"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("TLS enabled, HTTP_ADDR is set without scheme, URL has no host and ambiguous scheme", func(t *testing.T) {
		t.Setenv("CONSUL_HTTP_ADDR", "localhost:8501")
		t.Setenv("CONSUL_HTTP_SSL", "true")

		u, _ := url.Parse("consul://")
		expected := &url.URL{Host: "localhost:8501", Scheme: "https"}
		actual, err := consulURL(u)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestConsulAddrFromEnv(t *testing.T) {
	in := ""

	_, err := consulAddrFromEnv("bogus:url:xxx")
	assert.Error(t, err)

	addr, err := consulAddrFromEnv(in)
	require.NoError(t, err)
	assert.Empty(t, addr)

	addr, err = consulAddrFromEnv("https://foo:8500")
	require.NoError(t, err)
	assert.Equal(t, &url.URL{Scheme: "https", Host: "foo:8500"}, addr)

	addr, err = consulAddrFromEnv("foo:8500")
	require.NoError(t, err)
	assert.Equal(t, &url.URL{Host: "foo:8500"}, addr)
}

func TestSetupTLS(t *testing.T) {
	expected := &consulapi.TLSConfig{
		Address:  "address",
		CAFile:   "cafile",
		CAPath:   "ca/path",
		CertFile: "certfile",
		KeyFile:  "keyfile",
	}

	t.Setenv("CONSUL_TLS_SERVER_NAME", expected.Address)
	t.Setenv("CONSUL_CACERT", expected.CAFile)
	t.Setenv("CONSUL_CAPATH", expected.CAPath)
	t.Setenv("CONSUL_CLIENT_CERT", expected.CertFile)
	t.Setenv("CONSUL_CLIENT_KEY", expected.KeyFile)

	assert.Equal(t, expected, setupTLS())

	t.Run("CONSUL_HTTP_SSL_VERIFY is true", func(t *testing.T) {
		expected.InsecureSkipVerify = false
		t.Setenv("CONSUL_HTTP_SSL_VERIFY", "true")
		assert.Equal(t, expected, setupTLS())
	})

	t.Run("CONSUL_HTTP_SSL_VERIFY is false", func(t *testing.T) {
		expected.InsecureSkipVerify = true
		t.Setenv("CONSUL_HTTP_SSL_VERIFY", "false")
		assert.Equal(t, expected, setupTLS())
	})
}

func TestConsulConfig(t *testing.T) {
	t.Run("default ", func(t *testing.T) {
		expectedConfig := &store.Config{}

		actualConfig, err := consulConfig(false)
		require.NoError(t, err)

		assert.Equal(t, expectedConfig, actualConfig)
	})

	t.Run("with CONSUL_TIMEOUT", func(t *testing.T) {
		t.Setenv("CONSUL_TIMEOUT", "10")
		expectedConfig := &store.Config{
			ConnectionTimeout: 10 * time.Second,
		}

		actualConfig, err := consulConfig(false)
		require.NoError(t, err)
		assert.Equal(t, expectedConfig, actualConfig)
	})

	t.Run("with TLS", func(t *testing.T) {
		expectedConfig := &store.Config{
			TLS: &tls.Config{MinVersion: tls.VersionTLS13},
		}
		actualConfig, err := consulConfig(true)
		require.NoError(t, err)
		assert.NotNil(t, actualConfig.TLS)
		actualConfig.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
		assert.Equal(t, expectedConfig, actualConfig)
	})
}
