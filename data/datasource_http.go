package data

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func buildURL(base *url.URL, args ...string) (*url.URL, error) {
	if len(args) == 0 {
		return base, nil
	}
	p, err := url.Parse(args[0])
	if err != nil {
		return nil, errors.Wrapf(err, "bad sub-path %s", args[0])
	}
	return base.ResolveReference(p), nil
}

func readHTTP(source *Source, args ...string) ([]byte, error) {
	if source.hc == nil {
		source.hc = &http.Client{Timeout: time.Second * 5}
	}
	u, err := buildURL(source.URL, args...)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = source.header
	res, err := source.hc.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		err := errors.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, source.URL, string(body))
		return nil, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		source.mediaType = mediatype
	}
	return body, nil
}

func parseHeaderArgs(headerArgs []string) (map[string]http.Header, error) {
	headers := make(map[string]http.Header)
	for _, v := range headerArgs {
		ds, name, value, err := splitHeaderArg(v)
		if err != nil {
			return nil, err
		}
		if _, ok := headers[ds]; !ok {
			headers[ds] = make(http.Header)
		}
		headers[ds][name] = append(headers[ds][name], strings.TrimSpace(value))
	}
	return headers, nil
}

func splitHeaderArg(arg string) (datasourceAlias, name, value string, err error) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		err = errors.Errorf("Invalid datasource-header option '%s'", arg)
		return "", "", "", err
	}
	datasourceAlias = parts[0]
	name, value, err = splitHeader(parts[1])
	return datasourceAlias, name, value, err
}

func splitHeader(header string) (name, value string, err error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		err = errors.Errorf("Invalid HTTP Header format '%s'", header)
		return "", "", err
	}
	name = http.CanonicalHeaderKey(parts[0])
	value = parts[1]
	return name, value, nil
}
