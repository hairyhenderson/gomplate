package data

import (
	"strings"

	"github.com/hairyhenderson/gomplate/libkv"
)

func readConsul(source *Source, args ...string) (data []byte, err error) {
	if source.kv == nil {
		source.kv, err = libkv.NewConsul(source.URL)
		if err != nil {
			return nil, err
		}
		err = source.kv.Login()
		if err != nil {
			return nil, err
		}
	}

	p := source.URL.Path
	if len(args) == 1 {
		p = strings.TrimRight(p, "/") + "/" + args[0]
	}

	if strings.HasSuffix(p, "/") {
		source.mediaType = jsonArrayMimetype
		data, err = source.kv.List(p)
	} else {
		data, err = source.kv.Read(p)
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}
