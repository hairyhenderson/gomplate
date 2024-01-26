package config

import (
	"strings"

	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
)

func parseTemplateArg(value string) (alias string, ds DataSource, err error) {
	alias, u, _ := strings.Cut(value, "=")
	if u == "" {
		u = alias
	}

	ds.URL, err = urlhelpers.ParseSourceURL(u)

	return alias, ds, err
}
