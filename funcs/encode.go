package funcs

import (
	"context"
	"net/url"
)

type EncodeFuncs struct {
}

func CreateEncodeFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &EncodeFuncs{}
	f["urlencode"] = ns.URLEncode
	f["urldecode"] = ns.URLDecode
	return f
}

func (t EncodeFuncs) URLEncode(input string) string {
	return url.QueryEscape(input)
}

func (t EncodeFuncs) URLDecode(input string) (string, error) {
	return url.QueryUnescape(input)
}
