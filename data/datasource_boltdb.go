package data

import (
	"github.com/hairyhenderson/gomplate/v3/libkv"
	"github.com/pkg/errors"
)

func readBoltDB(source *Source, args ...string) (data []byte, err error) {
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
