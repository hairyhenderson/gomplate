package datasource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/base64"
	"github.com/hairyhenderson/gomplate/v3/env"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Git -
type Git struct{}

var _ Reader = (*Git)(nil)

func (g *Git) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	repoURL, path, err := g.parseGitPath(url, args...)
	if err != nil {
		return data, err
	}

	depth := 1
	if url.Scheme == "git+file" {
		// we can't do shallow clones for filesystem repos apparently
		depth = 0
	}

	fs, _, err := g.clone(ctx, repoURL, depth)
	if err != nil {
		return data, err
	}

	data.MediaType, data.Bytes, err = g.read(fs, path)
	return data, err
}

func (g *Git) parseArgURL(arg string) (u *url.URL, err error) {
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

func (g *Git) parseQuery(orig, arg *url.URL) string {
	q := orig.Query()
	pq := arg.Query()
	for k, vs := range pq {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return q.Encode()
}

func (g *Git) parseArgPath(u *url.URL, arg string) (repo, p string) {
	// if the source URL already specified a repo and subpath, the whole
	// arg is interpreted as subpath
	if strings.Contains(u.Path, "//") || strings.HasPrefix(arg, "//") {
		return "", arg
	}

	parts := strings.SplitN(arg, "//", 2)
	repo = parts[0]
	if len(parts) == 2 {
		p = "/" + parts[1]
	}
	return repo, p
}

// Massage the URL and args together to produce the repo to clone,
// and the path to read.
// The path is delimited from the repo by '//'
func (g *Git) parseGitPath(u *url.URL, args ...string) (out *url.URL, p string, err error) {
	if u == nil {
		return nil, "", fmt.Errorf("parseGitPath: no url provided (%v)", u)
	}
	// copy the input url so we can modify it
	out = cloneURL(u)

	parts := strings.SplitN(out.Path, "//", 2)
	switch len(parts) {
	case 1:
		p = "/"
	case 2:
		p = "/" + parts[1]

		i := strings.LastIndex(out.Path, p)
		out.Path = out.Path[:i-1]
	}

	if len(args) > 0 {
		argURL, err := g.parseArgURL(args[0])
		if err != nil {
			return nil, "", err
		}
		repo, argpath := g.parseArgPath(u, argURL.Path)
		out.Path = path.Join(out.Path, repo)
		p = path.Join(p, argpath)

		out.RawQuery = g.parseQuery(u, argURL)

		if argURL.Fragment != "" {
			out.Fragment = argURL.Fragment
		}
	}
	return out, p, err
}

func (g *Git) refFromURL(u *url.URL) plumbing.ReferenceName {
	switch {
	case strings.HasPrefix(u.Fragment, "refs/"):
		return plumbing.ReferenceName(u.Fragment)
	case u.Fragment != "":
		return plumbing.NewBranchReferenceName(u.Fragment)
	default:
		return plumbing.ReferenceName("")
	}
}

// clone a repo for later reading through http(s), git, or ssh. u must be the URL to the repo
// itself, and must have any file path stripped
func (g *Git) clone(ctx context.Context, repoURL *url.URL, depth int) (billy.Filesystem, *git.Repository, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	// preserve repoURL by cloning it
	u := cloneURL(repoURL)

	auth, err := g.auth(u)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasPrefix(u.Scheme, "git+") {
		scheme := u.Scheme[len("git+"):]
		u.Scheme = scheme
	}

	ref := g.refFromURL(u)
	u.Fragment = ""
	u.RawQuery = ""

	opts := &git.CloneOptions{
		URL:           u.String(),
		Auth:          auth,
		Depth:         depth,
		ReferenceName: ref,
		SingleBranch:  true,
		Tags:          git.NoTags,
	}
	repo, err := git.CloneContext(ctx, storer, fs, opts)
	if u.Scheme == "file" && err == transport.ErrRepositoryNotFound && !strings.HasSuffix(u.Path, ".git") {
		// maybe this has a `.git` subdirectory...
		u = cloneURL(repoURL)
		u.Path = path.Join(u.Path, ".git")
		return g.clone(ctx, u, depth)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("git clone for %v failed: %w", repoURL, err)
	}
	return fs, repo, nil
}

// read - reads the provided path out of a git repo
func (g *Git) read(fs billy.Filesystem, path string) (string, []byte, error) {
	fi, err := fs.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("can't stat %s: %w", path, err)
	}
	if fi.IsDir() || strings.HasSuffix(path, string(filepath.Separator)) {
		out, err := g.readDir(fs, path)
		return jsonArrayMimetype, out, err
	}

	f, err := fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", nil, fmt.Errorf("can't open %s: %w", path, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", nil, fmt.Errorf("can't read %s: %w", path, err)
	}

	return "", b, nil
}

func (g *Git) readDir(fs billy.Filesystem, path string) ([]byte, error) {
	names, err := fs.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read dir %s: %w", path, err)
	}
	files := make([]string, len(names))
	for i, v := range names {
		files[i] = v.Name()
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(files); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	// chop off the newline added by the json encoder
	return b[:len(b)-1], nil
}

/*
auth methods:
- ssh named key (no password support)
	- GIT_SSH_KEY (base64-encoded) or GIT_SSH_KEY_FILE (base64-encoded, or not)
- ssh agent auth (preferred)
- http basic auth (for github, gitlab, bitbucket tokens)
- http token auth (bearer token, somewhat unusual)
*/
func (g *Git) auth(u *url.URL) (auth transport.AuthMethod, err error) {
	user := u.User.Username()
	switch u.Scheme {
	case "git+http", "git+https":
		if pass, ok := u.User.Password(); ok {
			auth = &http.BasicAuth{Username: user, Password: pass}
		} else if pass := env.Getenv("GIT_HTTP_PASSWORD"); pass != "" {
			auth = &http.BasicAuth{Username: user, Password: pass}
		} else if tok := env.Getenv("GIT_HTTP_TOKEN"); tok != "" {
			// note docs on TokenAuth - this is rarely to be used
			auth = &http.TokenAuth{Token: tok}
		}
	case "git+ssh":
		k := env.Getenv("GIT_SSH_KEY")
		if k != "" {
			var key []byte
			key, err = base64.Decode(k)
			if err != nil {
				key = []byte(k)
			}
			auth, err = ssh.NewPublicKeys(user, key, "")
		} else {
			auth, err = ssh.NewSSHAgentAuth(user)
		}
	}
	return auth, err
}
