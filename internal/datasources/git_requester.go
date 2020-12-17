package datasources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/base64"
	"github.com/hairyhenderson/gomplate/v3/env"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type gitRequester struct{}

func (r *gitRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	depth := 1
	if u.Scheme == "git+file" {
		// we can't do shallow clones for filesystem repos apparently
		depth = 0
	}

	repoPath, path, err := r.splitRepoPath(u.Path)
	if err != nil {
		return nil, err
	}

	repoURL := cloneURL(u)
	repoURL.Path = repoPath

	fs, _, err := r.clone(ctx, repoURL, depth)
	if err != nil {
		return nil, err
	}

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("can't stat %s: %w", path, err)
	}

	resp := &Response{}

	hint := ""

	if fi.IsDir() || strings.HasSuffix(path, string(filepath.Separator)) {
		hint = jsonArrayMimetype
		b, err := r.readDir(fs, path)
		if err != nil {
			return nil, fmt.Errorf("failed to list directory %q: %w", path, err)
		}

		resp.ContentLength = int64(len(b))
		resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	} else {
		resp.Body, err = fs.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return nil, fmt.Errorf("can't open %q: %w", path, err)
		}

		resp.ContentLength = fi.Size()
	}

	resp.ContentType, err = mimeType(u, hint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Split the git repo path from the subpath, delimited by "//"
func (r *gitRequester) splitRepoPath(repopath string) (repo, subpath string, err error) {
	parts := strings.SplitN(repopath, "//", 2)
	switch len(parts) {
	case 1:
		subpath = "/"
	case 2:
		subpath = "/" + parts[1]

		i := strings.LastIndex(repopath, subpath)
		repopath = repopath[:i-1]
	}
	return repopath, subpath, err
}

func (r *gitRequester) refFromURL(u *url.URL) plumbing.ReferenceName {
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
func (r *gitRequester) clone(ctx context.Context, repoURL *url.URL, depth int) (billy.Filesystem, *git.Repository, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	// preserve repoURL by cloning it
	u := cloneURL(repoURL)

	auth, err := r.auth(u)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasPrefix(u.Scheme, "git+") {
		scheme := u.Scheme[len("git+"):]
		u.Scheme = scheme
	}

	ref := r.refFromURL(u)
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
		return r.clone(ctx, u, depth)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("git clone for %v failed: %w", repoURL, err)
	}
	return fs, repo, nil
}

func (r *gitRequester) readDir(fs billy.Filesystem, path string) ([]byte, error) {
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
func (r *gitRequester) auth(u *url.URL) (auth transport.AuthMethod, err error) {
	user := u.User.Username()
	switch u.Scheme {
	case "git+http", "git+https":
		if pass, ok := u.User.Password(); ok {
			auth = &githttp.BasicAuth{Username: user, Password: pass}
		} else if pass := env.Getenv("GIT_HTTP_PASSWORD"); pass != "" {
			auth = &githttp.BasicAuth{Username: user, Password: pass}
		} else if tok := env.Getenv("GIT_HTTP_TOKEN"); tok != "" {
			// note docs on TokenAuth - this is rarely to be used
			auth = &githttp.TokenAuth{Token: tok}
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
