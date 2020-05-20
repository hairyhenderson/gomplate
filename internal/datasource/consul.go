package datasource

import (
	"context"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/libkv"
)

// Consul -
type Consul struct {
	kv *libkv.LibKV
}

var _ Reader = (*Consul)(nil)

func (c *Consul) Read(ctx context.Context, url *url.URL, args ...string) (data *Data, err error) {
	if c.kv == nil {
		c.kv, err = libkv.NewConsul(url)
		if err != nil {
			return nil, err
		}
		err = c.kv.Login()
		if err != nil {
			return nil, err
		}
	}

	p := url.Path
	if len(args) == 1 {
		p = strings.TrimRight(p, "/") + "/" + args[0]
	}

	data = newData(url, args)

	if strings.HasSuffix(p, "/") {
		data.MType = jsonArrayMimetype
		data.Bytes, err = c.kv.List(p)
	} else {
		data.Bytes, err = c.kv.Read(p)
	}

	return data, err
}
