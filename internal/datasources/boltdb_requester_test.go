package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBoltDBStoreKey(t *testing.T) {
	testdata := []struct {
		in  *url.URL
		out string
	}{
		{mustParseURL(""), ""},
		{mustParseURL("boltdb:thisisopaque"), "boltdb:thisisopaque"},
		{mustParseURL("boltdb:///tmp/foo.db"), "boltdb:///tmp/foo.db"},
		{mustParseURL("boltdb:///tmp/foo.db#bucket1"), "boltdb:///tmp/foo.db#bucket1"},
		{mustParseURL("boltdb:///tmp/foo.db?type=foo/bar#bucket1"), "boltdb:///tmp/foo.db#bucket1"},
		{mustParseURL("boltdb:///tmp/foo.db?key=1&type=foo/bar#bucket1"), "boltdb:///tmp/foo.db#bucket1"},
	}

	for _, d := range testdata {
		d := d
		t.Run(fmt.Sprintf("%q==%q", d.in, d.out), func(t *testing.T) {
			out := storeKey(d.in)
			assert.Equal(t, d.out, out)
		})
	}
}

func TestBoltDBRequest(t *testing.T) {
	kv := &testKV{body: []byte("hello world")}
	u := mustParseURL("boltdb:///foo.db?key=foo")

	r := &boltDBRequester{
		kv: map[string]kvStore{
			storeKey(u): kv,
		},
	}

	ctx := context.Background()

	resp, err := r.Request(ctx, u, nil)
	assert.NilError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "text/plain", resp.ContentType)
	assert.Equal(t, int64(11), resp.ContentLength)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello world", string(b))
}
