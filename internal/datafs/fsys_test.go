package datafs

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSourceURL(t *testing.T) {
	expected := &url.URL{
		Scheme:   "http",
		Host:     "example.com",
		Path:     "/foo.json",
		RawQuery: "bar",
	}
	u, err := ParseSourceURL("http://example.com/foo.json?bar")
	require.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{Scheme: "", Path: ""}
	u, err = ParseSourceURL("")
	require.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{Scheme: "stdin"}
	u, err = ParseSourceURL("-")
	require.NoError(t, err)
	assert.EqualValues(t, expected, u)

	// behviour change in v4 - return relative if it's relative
	expected = &url.URL{Path: "./foo/bar.json"}
	u, err = ParseSourceURL("./foo/bar.json")
	require.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{Scheme: "file", Path: "/absolute/bar.json"}
	u, err = ParseSourceURL("/absolute/bar.json")
	require.NoError(t, err)
	assert.EqualValues(t, expected, u)
}
