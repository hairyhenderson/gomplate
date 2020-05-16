package datasource

import (
	"context"
	"errors"
	"net/url"

	"github.com/hairyhenderson/gomplate/v3/libkv"
)

// BoltDB -
type BoltDB struct {
	kv *libkv.LibKV
}

var _ Reader = (*BoltDB)(nil)

func (b *BoltDB) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	if b.kv == nil {
		b.kv, err = libkv.NewBoltDB(url)
		if err != nil {
			return data, err
		}
	}

	if len(args) != 1 {
		return data, errors.New("missing key")
	}
	p := args[0]

	data.Bytes, err = b.kv.Read(p)
	if err != nil {
		return data, err
	}

	return data, nil
}
