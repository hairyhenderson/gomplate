package data

import (
	"context"

	"github.com/hairyhenderson/gomplate/v3/internal/deprecated"
	"github.com/hairyhenderson/gomplate/v3/libkv"
	"github.com/pkg/errors"
)

// Deprecated: don't use
func readBoltDB(ctx context.Context, source *Source, args ...string) (data []byte, err error) {
	deprecated.WarnDeprecated(ctx, "boltdb support is deprecated and will be removed in a future major version of gomplate")
	if source.kv == nil {
		source.kv, err = libkv.NewBoltDB(source.URL)
		if err != nil {
			return nil, err
		}
	}

	if len(args) != 1 {
		return nil, errors.New("missing key")
	}
	p := args[0]

	data, err = source.kv.Read(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}
