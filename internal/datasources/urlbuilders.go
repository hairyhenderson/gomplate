package datasources

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func lookupURLBuilder(scheme string) (b urlBuilder, err error) {
	switch scheme {
	case "stdin", "env", "merge":
		b = &noopURLBuilder{}
	case "boltdb":
		b = &boltDBURLBuilder{}
	case "http", "https":
		b = &httpURLBuilder{}
	case "file",
		"vault", "vault+http", "vault+https",
		"consul", "consul+http", "consul+https":
		b = &vaultURLBuilder{}
	case "aws+sm", "aws+smp":
		b = &awssmURLBuilder{}
	case "s3", "gs":
		b = &blobURLBuilder{}
	case "git", "git+file", "git+http", "git+https", "git+ssh":
		b = &gitURLBuilder{}
	}

	if b != nil {
		return b, nil
	}
	return nil, fmt.Errorf("no URL builder found for scheme %s (not registered?)", scheme)
}

func cloneURL(u *url.URL) *url.URL {
	out := *u
	return &out
}

type noopURLBuilder struct{}

func (b *noopURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	return u, nil
}

type boltDBURLBuilder struct{}

func (b *boltDBURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("missing key - must provide key argument for boltdb datasources")
	}

	out := cloneURL(u)

	p, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("bad key argument %q: %w", args[0], err)
	}

	q := u.Query()
	q.Set("key", p.Path)
	for k, vs := range p.Query() {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	out.RawQuery = q.Encode()

	return out, nil
}

type httpURLBuilder struct{}

func (b *httpURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) == 0 {
		return u, nil
	}

	if len(args) >= 2 {
		return nil, fmt.Errorf("too many args for building %q URL: found %d",
			u.Scheme, len(args))
	}

	p, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("bad sub-path %q: %w", args[0], err)
	}

	out := u.ResolveReference(p)
	if p.RawQuery == "" {
		out.RawQuery = u.RawQuery
	} else if u.RawQuery != "" {
		// merge the query params
		q := u.Query()
		for k, vs := range p.Query() {
			for _, v := range vs {
				q.Set(k, v)
			}
		}
		out.RawQuery = q.Encode()
	}

	if u.Fragment != "" {
		out.Fragment = u.Fragment
		out.RawFragment = u.RawFragment
	}

	return out, nil
}

type vaultURLBuilder struct{}

func (b *vaultURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) >= 2 {
		return nil, fmt.Errorf("too many args for building %q URL: found %d",
			u.Scheme, len(args))
	}

	out := *u

	if out.Path == "" && u.Opaque != "" {
		out.Path = u.Opaque
	}

	if len(args) != 1 {
		return &out, nil
	}

	p, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("bad sub-path %q: %w", args[0], err)
	}
	if p.Path != "" {
		out.Path = path.Join(out.Path, p.Path)
		if strings.HasSuffix(p.Path, "/") {
			out.Path += "/"
		}
	}

	if p.RawQuery == "" {
		return &out, nil
	}

	// merge the query params
	q := u.Query()
	for k, vs := range p.Query() {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	out.RawQuery = q.Encode()

	return &out, nil
}

type awssmURLBuilder struct{}

func (b *awssmURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) >= 2 {
		return nil, fmt.Errorf("too many args for building %q URL: found %d",
			u.Scheme, len(args))
	}

	out := &url.URL{
		Scheme:   u.Scheme,
		Path:     u.Path,
		RawQuery: u.RawQuery,
	}

	if out.Path == "" && u.Opaque != "" {
		out.Path = u.Opaque
	}

	if len(args) != 1 {
		return out, nil
	}

	p, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("bad sub-path %q: %w", args[0], err)
	}
	if p.Path != "" {
		out.Path = path.Join(out.Path, p.Path)
		if strings.HasSuffix(p.Path, "/") {
			out.Path += "/"
		}
	}

	if p.RawQuery == "" {
		out.RawQuery = u.RawQuery
	} else if u.RawQuery != "" {
		// merge the query params
		q := u.Query()
		for k, vs := range p.Query() {
			for _, v := range vs {
				q.Set(k, v)
			}
		}
		out.RawQuery = q.Encode()
	}

	return out, nil
}

type blobURLBuilder struct{}

func (b *blobURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) >= 2 {
		return nil, fmt.Errorf("too many args for building %q URL: found %d",
			u.Scheme, len(args))
	}

	out := &url.URL{
		Host:     u.Host,
		Scheme:   u.Scheme,
		Path:     u.Path,
		RawQuery: u.RawQuery,
	}

	if len(args) != 1 {
		return out, nil
	}

	p, err := url.Parse(args[0])
	if err != nil {
		return nil, fmt.Errorf("bad sub-path %q: %w", args[0], err)
	}
	if p.Path != "" {
		out.Path = path.Join(out.Path, p.Path)
		if strings.HasSuffix(p.Path, "/") {
			out.Path += "/"
		}
	}

	if p.RawQuery == "" {
		out.RawQuery = u.RawQuery
	} else if u.RawQuery != "" {
		// merge the query params
		q := u.Query()
		for k, vs := range p.Query() {
			for _, v := range vs {
				q.Set(k, v)
			}
		}
		out.RawQuery = q.Encode()
	}

	return out, nil
}

type gitURLBuilder struct{}

func (b *gitURLBuilder) BuildURL(u *url.URL, args ...string) (*url.URL, error) {
	if len(args) >= 2 {
		return nil, fmt.Errorf("too many args for building %q URL: found %d",
			u.Scheme, len(args))
	}

	if len(args) == 0 {
		return u, nil
	}

	out := cloneURL(u)

	subPath := "/"
	parts := strings.SplitN(out.Path, "//", 2)
	if len(parts) == 2 {
		subPath = "/" + parts[1]

		i := strings.LastIndex(out.Path, subPath)
		out.Path = out.Path[:i-1]
	}

	argURL, err := b.parseArgURL(args[0])
	if err != nil {
		return nil, err
	}

	repo, argpath := b.parseArgPath(u.Path, argURL.Path)
	out.Path = path.Join(out.Path, repo)
	subPath = path.Join(subPath, argpath)

	// join the repo path and the subpath together again, separated with a "//"
	if subPath != "/" {
		out.Path = out.Path + "/" + subPath
	}

	out.RawQuery = b.parseQuery(u, argURL)

	if argURL.Fragment != "" {
		out.Fragment = argURL.Fragment
	}

	return out, nil
}

func (b gitURLBuilder) parseArgURL(arg string) (u *url.URL, err error) {
	if strings.HasPrefix(arg, "//") {
		u, err = url.Parse(arg[1:])
		u.Path = "/" + u.Path
	} else {
		u, err = url.Parse(arg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse arg %s: %w", arg, err)
	}
	return u, err
}

func (b gitURLBuilder) parseQuery(orig, arg *url.URL) string {
	q := orig.Query()
	pq := arg.Query()
	for k, vs := range pq {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return q.Encode()
}

// parseArgPath -
func (b gitURLBuilder) parseArgPath(orig, arg string) (repo, subpath string) {
	// if the original path already specified a repo and subpath, the whole
	// arg is interpreted as subpath
	if strings.Contains(orig, "//") || strings.HasPrefix(arg, "//") {
		return "", arg
	}

	parts := strings.SplitN(arg, "//", 2)
	repo = parts[0]
	if len(parts) == 2 {
		subpath = "/" + parts[1]
	}
	return repo, subpath
}
