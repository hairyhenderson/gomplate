package datasource

import "net/url"

//nolint: interfacer
func cloneURL(u *url.URL) *url.URL {
	out, _ := url.Parse(u.String())
	return out
}

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}
