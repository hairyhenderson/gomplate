package datasources

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hairyhenderson/gomplate/v3/libkv"
)

type boltDBRequester struct {
	kv map[string]kvStore
}

func storeKey(u *url.URL) string {
	k := *u
	k.RawQuery = ""
	return k.String()
}

func (r *boltDBRequester) Request(ctx context.Context, u *url.URL, header http.Header) (resp *Response, err error) {
	if r.kv == nil {
		r.kv = map[string]kvStore{}
	}

	k := storeKey(u)
	kv, ok := r.kv[k]
	if !ok {
		kv, err = libkv.NewBoltDB(u)
		if err != nil {
			return nil, err
		}
		r.kv[k] = kv
	}

	key := u.Query().Get("key")
	data, err := kv.Read(key)
	if err != nil {
		return nil, err
	}

	resp = &Response{
		Body:          ioutil.NopCloser(bytes.NewReader(data)),
		ContentLength: int64(len(data)),
	}

	resp.ContentType, err = mimeType(u, "")
	if err != nil {
		return nil, err
	}

	return resp, nil
}
