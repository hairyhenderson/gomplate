package azure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hairyhenderson/gomplate/v4/env"
)

// defaultEndpoint is the IMDS endpoint at Azure
// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service
const defaultEndpoint = "http://169.254.169.254"

var (
	// co is a ClientOptions populated from the environment.
	co ClientOptions
	// coInit ensures that `co` is only set once.
	coInit sync.Once
)

// ClientOptions contains various user-specifiable options for a MetaClient.
type ClientOptions struct {
	Timeout time.Duration
}

// GetClientOptions - Centralised reading of AZURE_TIMEOUT
func GetClientOptions() ClientOptions {
	coInit.Do(func() {
		timeout := env.Getenv("AZURE_TIMEOUT")
		if timeout == "" {
			timeout = "500"
		}

		t, err := strconv.Atoi(timeout)
		if err != nil {
			panic(fmt.Errorf("invalid AZURE_TIMEOUT value '%s' - must be an integer: %w", timeout, err))
		}

		co.Timeout = time.Duration(t) * time.Millisecond
	})
	return co
}

// MetaClient is used to access metadata accessible via the Azure compute instance
// metadata service.
type MetaClient struct {
	client   *http.Client
	cache    map[string]string
	endpoint string
	options  ClientOptions
}

// NewMetaClient constructs a new MetaClient with the given ClientOptions. If the environment
// contains a variable named `AZURE_META_ENDPOINT`, the client will address that, if not the
// value of `DefaultEndpoint` is used.
func NewMetaClient(options ClientOptions) *MetaClient {
	endpoint := env.Getenv("AZURE_META_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	return &MetaClient{
		cache:    make(map[string]string),
		endpoint: endpoint,
		options:  options,
	}
}

// Meta retrieves a value from the Azure Instance Metadata Service, returning the given default
// if the service is unavailable or the requested URL does not exist.
func (c *MetaClient) Meta(ctx context.Context, key, format, apiVersion string, def ...string) (string, error) {
	url := c.endpoint + "/metadata/instance/" + key + "?api-version=" + apiVersion + "&format=" + format
	return c.retrieveMetadata(ctx, url, def...)
}

// retrieveMetadata executes an HTTP request to the Azure Instance Metadata Service with the
// correct headers set, and extracts the returned value.
func (c *MetaClient) retrieveMetadata(ctx context.Context, url string, def ...string) (string, error) {
	if value, ok := c.cache[url]; ok {
		return value, nil
	}

	if c.client == nil {
		timeout := c.options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}

		// Call IMDS through proxy is not supported
		// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service#proxies
		c.client = &http.Client{Timeout: timeout, Transport: &http.Transport{Proxy: nil}}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return returnDefault(def), nil
	}
	request.Header.Add("Metadata", "true")

	resp, err := c.client.Do(request)
	if err != nil {
		return returnDefault(def), nil
	}

	// nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return returnDefault(def), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}
	value := strings.TrimSpace(string(body))
	c.cache[url] = value

	return value, nil
}

// returnDefault returns the first element of the given slice (often taken from varargs)
// if there is one, or returns an empty string if the slice has no elements.
func returnDefault(def []string) string {
	if len(def) > 0 {
		return def[0]
	}
	return ""
}
