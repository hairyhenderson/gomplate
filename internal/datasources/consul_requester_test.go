package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

type testKV struct {
	err   error
	body  []byte
	items []string
}

var _ kvStore = (*testKV)(nil)

func (k *testKV) Login() error {
	return nil
}

func (k *testKV) Logout() {}

func (k *testKV) Read(path string) ([]byte, error) {
	return k.body, k.err
}

func (k *testKV) List(path string) ([]byte, error) {
	return json.Marshal(k.items)
}

func TestConsulStoreKey(t *testing.T) {
	testdata := []struct {
		in  *url.URL
		out string
	}{
		{mustParseURL(""), ""},
		{mustParseURL("consul://example.com/tmp/foo.db"), "consul://example.com"},
		{mustParseURL("consul://example.com/tmp/foo.db"), "consul://example.com"},
	}

	for _, d := range testdata {
		d := d
		t.Run(fmt.Sprintf("%q==%q", d.in, d.out), func(t *testing.T) {
			out := consulStoreKey(d.in)
			assert.Equal(t, d.out, out)
		})
	}
}

func TestConsulRequest(t *testing.T) {
	kv := &testKV{body: []byte("hello world")}
	u := mustParseURL("consul://example.com/foo/bar")

	r := &consulRequester{
		kv: map[string]kvStore{
			consulStoreKey(u): kv,
		},
	}

	ctx := context.Background()

	resp, err := r.Request(ctx, u, nil)
	assert.NilError(t, err)
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello world", string(b))

	u = mustParseURL("consul://example.com/")
	r.kv[consulStoreKey(u)] = &testKV{items: []string{"foo", "bar", "baz"}}

	resp, err = r.Request(ctx, u, nil)
	assert.NilError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, `["foo","bar","baz"]`, string(b))

	u = mustParseURL("consul://example.com/sub/")
	r.kv[consulStoreKey(u)] = &testKV{items: []string{"foo", "bar", "baz"}}
	resp, err = r.Request(ctx, u, nil)
	assert.NilError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, `["foo","bar","baz"]`, string(b))
}
