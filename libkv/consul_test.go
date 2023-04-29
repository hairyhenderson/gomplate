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
	defer os.Unsetenv("CONSUL_HTTP_SSL")
	os.Setenv("CONSUL_HTTP_SSL", "true")

	u, _ := url.Parse("consul://")
	expected := &url.URL{Host: "localhost:8500", Scheme: "https"}
	actual, err := consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	u, _ = url.Parse("consul+http://myconsul.server")
	expected = &url.URL{Host: "myconsul.server", Scheme: "http"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Setenv("CONSUL_HTTP_SSL", "false")
	u, _ = url.Parse("consul+https://myconsul.server:1234")
	expected = &url.URL{Host: "myconsul.server:1234", Scheme: "https"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	os.Unsetenv("CONSUL_HTTP_SSL")
	u, _ = url.Parse("consul://myconsul.server:2345")
	expected = &url.URL{Host: "myconsul.server:2345", Scheme: "http"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	u, _ = url.Parse("consul://myconsul.server:3456/foo/bar/baz")

	expected = &url.URL{Host: "myconsul.server:3456", Scheme: "http"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	defer os.Unsetenv("CONSUL_HTTP_ADDR")
	os.Setenv("CONSUL_HTTP_ADDR", "https://foo:8500")

	// given URL takes precedence over env var
	expected = &url.URL{Host: "myconsul.server:3456", Scheme: "http"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	u, _ = url.Parse("consul://")

	defer os.Unsetenv("CONSUL_HTTP_SSL")
	os.Setenv("CONSUL_HTTP_SSL", "true")

	// TLS enabled, HTTP_ADDR is set, URL has no host and ambiguous scheme
	expected = &url.URL{Host: "foo:8500", Scheme: "https"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	defer os.Unsetenv("CONSUL_HTTP_ADDR")
	os.Setenv("CONSUL_HTTP_ADDR", "localhost:8501")
	expected = &url.URL{Host: "localhost:8501", Scheme: "https"}
	actual, err = consulURL(u)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
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

	defer os.Unsetenv("CONSUL_TLS_SERVER_NAME")
	defer os.Unsetenv("CONSUL_CACERT")
	defer os.Unsetenv("CONSUL_CAPATH")
	defer os.Unsetenv("CONSUL_CLIENT_CERT")
	defer os.Unsetenv("CONSUL_CLIENT_KEY")
	os.Setenv("CONSUL_TLS_SERVER_NAME", expected.Address)
	os.Setenv("CONSUL_CACERT", expected.CAFile)
	os.Setenv("CONSUL_CAPATH", expected.CAPath)
	os.Setenv("CONSUL_CLIENT_CERT", expected.CertFile)
	os.Setenv("CONSUL_CLIENT_KEY", expected.KeyFile)

	assert.Equal(t, expected, setupTLS())

	expected.InsecureSkipVerify = false
	defer os.Unsetenv("CONSUL_HTTP_SSL_VERIFY")
	os.Setenv("CONSUL_HTTP_SSL_VERIFY", "true")
	assert.Equal(t, expected, setupTLS())

	expected.InsecureSkipVerify = true
	os.Setenv("CONSUL_HTTP_SSL_VERIFY", "false")
	assert.Equal(t, expected, setupTLS())
}

func TestConsulConfig(t *testing.T) {
	expectedConfig := &store.Config{}

	actualConfig, err := consulConfig(false)
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actualConfig)

	defer os.Unsetenv("CONSUL_TIMEOUT")
	os.Setenv("CONSUL_TIMEOUT", "10")
	expectedConfig = &store.Config{
		ConnectionTimeout: 10 * time.Second,
	}

	actualConfig, err = consulConfig(false)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, actualConfig)

	os.Unsetenv("CONSUL_TIMEOUT")
	expectedConfig = &store.Config{
		TLS: &tls.Config{MinVersion: tls.VersionTLS13},
	}

	actualConfig, err = consulConfig(true)
	require.NoError(t, err)
	assert.NotNil(t, actualConfig.TLS)
	actualConfig.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
	assert.Equal(t, expectedConfig, actualConfig)
}
