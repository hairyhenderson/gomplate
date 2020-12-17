package datasources

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/libkv"
)

type consulRequester struct {
	kv map[string]kvStore
}

// kvStore -
type kvStore interface {
	Login() error
	Logout()
	Read(path string) ([]byte, error)
	List(path string) ([]byte, error)
}

// consulStoreKey indexes a consul libkv client by scheme and host
func consulStoreKey(u *url.URL) string {
	k := *u
	k.RawQuery = ""
	k.Path = ""
	return k.String()
}

func (r *consulRequester) Request(ctx context.Context, u *url.URL, header http.Header) (resp *Response, err error) {
	kv, ok := r.kv[consulStoreKey(u)]
	if !ok {
		kv, err = libkv.NewConsul(u)
		if err != nil {
			return nil, err
		}

		err = kv.Login()
		if err != nil {
			return nil, err
		}

		// cache the client for potential later use
		r.kv[consulStoreKey(u)] = kv
	}

	var read func(string) ([]byte, error)

	hint := ""

	if strings.HasSuffix(u.Path, "/") {
		hint = jsonArrayMimetype
		read = kv.List
	} else {
		read = kv.Read
	}

	data, err := read(u.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read from consul key %q: %w", u.Path, err)
	}

	resp = &Response{
		Body:          ioutil.NopCloser(bytes.NewReader(data)),
		ContentLength: int64(len(data)),
	}

	resp.ContentType, err = mimeType(u, hint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
