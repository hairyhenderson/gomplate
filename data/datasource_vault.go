package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/vault"
)

func readVault(ctx context.Context, source *Source, args ...string) (data []byte, err error) {
	if source.vc == nil {
		source.vc, err = vault.New(source.URL)
		if err != nil {
			return nil, err
		}
		err = source.vc.Login()
		if err != nil {
			return nil, err
		}
	}

	params, p, err := parseDatasourceURLArgs(source.URL, args...)
	if err != nil {
		return nil, err
	}

	source.mediaType = jsonMimetype
	switch {
	case len(params) > 0:
		data, err = source.vc.Write(p, params)
	case strings.HasSuffix(p, "/"):
		source.mediaType = jsonArrayMimetype
		data, err = source.vc.List(p)
	default:
		data, err = source.vc.Read(p)
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no value found for path %s", p)
	}

	return data, nil
}
