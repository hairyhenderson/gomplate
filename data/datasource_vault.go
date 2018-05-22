package data

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/vault"
)

func parseVaultParams(sourceURL *url.URL, args []string) (params map[string]interface{}, p string, err error) {
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

func readVault(source *Source, args ...string) ([]byte, error) {
	if source.VC == nil {
		source.VC = vault.New(source.URL)
		source.VC.Login()
	}

	params, p, err := parseVaultParams(source.URL, args)
	if err != nil {
		return nil, err
	}

	var data []byte

	source.Type = jsonMimetype
	if len(params) > 0 {
		data, err = source.VC.Write(p, params)
	} else if strings.HasSuffix(p, "/") {
		source.Type = jsonArrayMimetype
		data, err = source.VC.List(p)
	} else {
		data, err = source.VC.Read(p)
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.Errorf("no value found for path %s", p)
	}

	return data, nil
}
