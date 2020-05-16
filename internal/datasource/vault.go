package datasource

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/vault"
)

// Vault -
type Vault struct {
	vc *vault.Vault
}

var _ Reader = (*Vault)(nil)

func (v *Vault) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	if v.vc == nil {
		v.vc, err = vault.New(url)
		if err != nil {
			return data, err
		}
		err = v.vc.Login()
		if err != nil {
			return data, err
		}
	}

	params, p, err := v.parseParams(url, args)
	if err != nil {
		return data, err
	}

	data.MediaType = jsonMimetype
	switch {
	case len(params) > 0:
		data.Bytes, err = v.vc.Write(p, params)
	case strings.HasSuffix(p, "/"):
		data.MediaType = jsonArrayMimetype
		data.Bytes, err = v.vc.List(p)
	default:
		data.Bytes, err = v.vc.Read(p)
	}
	if err != nil {
		return data, err
	}

	if len(data.Bytes) == 0 {
		return Data{}, fmt.Errorf("no value found for path %s", p)
	}
	return data, nil
}

func (v *Vault) parseParams(sourceURL *url.URL, args []string) (params map[string]interface{}, p string, err error) {
	p = sourceURL.Path
	params = make(map[string]interface{})
	for key, val := range sourceURL.Query() {
		params[key] = strings.Join(val, " ")
	}

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, "", err
		}

		if parsed.Path != "" {
			p = p + "/" + parsed.Path
		}

		for key, val := range parsed.Query() {
			params[key] = strings.Join(val, " ")
		}
	}
	return params, p, nil
}
