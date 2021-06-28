package datasources

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"golang.org/x/crypto/ssh/testdata"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSplitRepoPath(t *testing.T) {
	t.Parallel()

	u := mustParseURL("http://example.com//foo")
	assert.Equal(t, "//foo", u.Path)
	parts := strings.SplitN(u.Path, "//", 2)
	assert.Equal(t, 2, len(parts))
	assert.DeepEqual(t, []string{"", "foo"}, parts)

	data := []struct {
		in         string
		repo, path string
	}{
		{"/hairyhenderson/gomplate//docs-src/content/functions/aws.yml", "/hairyhenderson/gomplate", "/docs-src/content/functions/aws.yml"},
		{"/hairyhenderson/gomplate.git", "/hairyhenderson/gomplate.git", "/"},
		{"/", "/", "/"},
		{"/foo//file.txt", "/foo", "/file.txt"},
		{"/home/foo/repo//file.txt", "/home/foo/repo", "/file.txt"},
		{"/repo", "/repo", "/"},
		{"/foo//foo", "/foo", "/foo"},
		{"/foo//foo/bar", "/foo", "/foo/bar"},
		{"/foo/bar", "/foo/bar", "/"},
		{"/foo//bar", "/foo", "/bar"},
		{"//foo/bar", "", "/foo/bar"},
		{"/foo//bar/baz", "/foo", "/bar/baz"},
		{"/foo/bar//baz", "/foo/bar", "/baz"},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%d:(%q)==(%q,%q)", i, d.in, d.repo, d.path), func(t *testing.T) {
			t.Parallel()

			repo, path, err := splitRepoPath(d.in)
			assert.NilError(t, err)
			assert.Equal(t, d.repo, repo)
			assert.Equal(t, d.path, path)
		})
	}
}

// func TestReadGitRepo(t *testing.T) {
// 	g := &gitRequester{}
// 	fs := setupGitRepo(t)
// 	fs, err := fs.Chroot("/repo")
// 	assert.NilError(t, err)

// 	_, _, err = g.read(fs, "/bogus")
// 	assert.ErrorContains(t, err, "can't stat /bogus")

// 	mtype, out, err := g.read(fs, "/foo")
// 	assert.NilError(t, err)
// 	assert.Equal(t, `["bar"]`, string(out))
// 	assert.Equal(t, jsonArrayMimetype, mtype)

// 	mtype, out, err = g.read(fs, "/foo/bar")
// 	assert.NilError(t, err)
// 	assert.Equal(t, `["hi.txt"]`, string(out))
// 	assert.Equal(t, jsonArrayMimetype, mtype)

// 	mtype, out, err = g.read(fs, "/foo/bar/hi.txt")
// 	assert.NilError(t, err)
// 	assert.Equal(t, `hello world`, string(out))
// 	assert.Equal(t, "", mtype)
// }

var testHashes = map[string]string{}

func setupGitRepo(t *testing.T) billy.Filesystem {
	fs := memfs.New()
	fs.MkdirAll("/repo/.git", os.ModeDir)
	repo, _ := fs.Chroot("/repo")
	dot, _ := repo.Chroot("/.git")
	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	r, err := git.Init(s, repo)
	assert.NilError(t, err)

	// config needs to be created after setting up a "normal" fs repo
	// this is possibly a bug in src-d/git-go?
	c, err := r.Config()
	assert.NilError(t, err)
	s.SetConfig(c)
	assert.NilError(t, err)

	w, err := r.Worktree()
	assert.NilError(t, err)

	repo.MkdirAll("/foo/bar", os.ModeDir)
	f, err := repo.Create("/foo/bar/hi.txt")
	assert.NilError(t, err)
	_, err = f.Write([]byte("hello world"))
	assert.NilError(t, err)
	_, err = w.Add(f.Name())
	assert.NilError(t, err)
	hash, err := w.Commit("initial commit", &git.CommitOptions{Author: &object.Signature{}})
	assert.NilError(t, err)

	ref, err := r.CreateTag("v1", hash, nil)
	assert.NilError(t, err)
	testHashes["v1"] = hash.String()

	branchName := plumbing.NewBranchReferenceName("mybranch")
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchName,
		Hash:   ref.Hash(),
		Create: true,
	})
	assert.NilError(t, err)

	f, err = repo.Create("/secondfile.txt")
	assert.NilError(t, err)
	_, err = f.Write([]byte("another file\n"))
	assert.NilError(t, err)
	n := f.Name()
	_, err = w.Add(n)
	assert.NilError(t, err)
	hash, err = w.Commit("second commit", &git.CommitOptions{
		Author: &object.Signature{
			Name: "John Doe",
		},
	})
	ref = plumbing.NewHashReference(branchName, hash)
	assert.NilError(t, err)
	testHashes["mybranch"] = ref.Hash().String()

	// make the repo dirty
	_, err = f.Write([]byte("dirty file"))
	assert.NilError(t, err)

	// set up a bare repo
	fs.MkdirAll("/bare.git", os.ModeDir)
	fs.MkdirAll("/barewt", os.ModeDir)
	repo, _ = fs.Chroot("/barewt")
	dot, _ = fs.Chroot("/bare.git")
	s = filesystem.NewStorage(dot, nil)

	r, err = git.Init(s, repo)
	assert.NilError(t, err)

	w, err = r.Worktree()
	assert.NilError(t, err)

	f, err = repo.Create("/hello.txt")
	assert.NilError(t, err)
	f.Write([]byte("hello world"))
	w.Add(f.Name())
	_, err = w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NilError(t, err)

	return fs
}

func overrideFSLoader(fs billy.Filesystem) {
	l := server.NewFilesystemLoader(fs)
	client.InstallProtocol("file", server.NewClient(l))
}

func TestOpenFileRepo(t *testing.T) {
	ctx := context.Background()
	repoFS := setupGitRepo(t)
	g := &gitRequester{}

	overrideFSLoader(repoFS)
	defer overrideFSLoader(osfs.New(""))

	fs, _, err := g.clone(ctx, mustParseURL("git+file:///repo"), 0)
	assert.NilError(t, err)

	f, err := fs.Open("/foo/bar/hi.txt")
	assert.NilError(t, err)
	b, _ := ioutil.ReadAll(f)
	assert.Equal(t, "hello world", string(b))

	_, repo, err := g.clone(ctx, mustParseURL("git+file:///repo#master"), 0)
	assert.NilError(t, err)

	ref, err := repo.Reference(plumbing.NewBranchReferenceName("master"), true)
	assert.NilError(t, err)
	assert.Equal(t, "refs/heads/master", ref.Name().String())

	_, repo, err = g.clone(ctx, mustParseURL("git+file:///repo#refs/tags/v1"), 0)
	assert.NilError(t, err)

	ref, err = repo.Head()
	assert.NilError(t, err)
	assert.Equal(t, testHashes["v1"], ref.Hash().String())

	_, repo, err = g.clone(ctx, mustParseURL("git+file:///repo/#mybranch"), 0)
	assert.NilError(t, err)

	ref, err = repo.Head()
	assert.NilError(t, err)
	assert.Equal(t, "refs/heads/mybranch", ref.Name().String())
	assert.Equal(t, testHashes["mybranch"], ref.Hash().String())
}

func TestOpenBareFileRepo(t *testing.T) {
	ctx := context.Background()
	repoFS := setupGitRepo(t)
	g := &gitRequester{}

	overrideFSLoader(repoFS)
	defer overrideFSLoader(osfs.New(""))

	fs, _, err := g.clone(ctx, mustParseURL("git+file:///bare.git"), 0)
	assert.NilError(t, err)

	f, err := fs.Open("/hello.txt")
	assert.NilError(t, err)
	b, _ := ioutil.ReadAll(f)
	assert.Equal(t, "hello world", string(b))
}

func TestReadGit(t *testing.T) {
	repoFS := setupGitRepo(t)

	overrideFSLoader(repoFS)
	defer overrideFSLoader(osfs.New(""))

	r := &gitRequester{}
	ctx := context.Background()

	u := mustParseURL("git+file:///bare.git//hello.txt")
	resp, err := r.Request(ctx, u, nil)
	assert.NilError(t, err)
	assert.Assert(t, resp.Body != nil)

	defer resp.Body.Close()

	assert.Assert(t, resp.ContentLength == 11)

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "hello world", string(b))

	u = mustParseURL("git+file:///bare.git")
	resp, err = r.Request(ctx, u, nil)
	assert.NilError(t, err)
	assert.Equal(t, "application/array+json", resp.ContentType)

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, `["hello.txt"]`, string(b))
}

func TestGitAuth(t *testing.T) {
	g := &gitRequester{}
	a, err := g.auth(mustParseURL("git+file:///bare.git"))
	assert.NilError(t, err)
	assert.Equal(t, nil, a)

	a, err = g.auth(mustParseURL("git+https://example.com/foo"))
	assert.NilError(t, err)
	assert.Assert(t, is.Nil(a))

	a, err = g.auth(mustParseURL("git+https://user:swordfish@example.com/foo"))
	assert.NilError(t, err)
	assert.DeepEqual(t, &http.BasicAuth{Username: "user", Password: "swordfish"}, a)

	os.Setenv("GIT_HTTP_PASSWORD", "swordfish")
	defer os.Unsetenv("GIT_HTTP_PASSWORD")
	a, err = g.auth(mustParseURL("git+https://user@example.com/foo"))
	assert.NilError(t, err)
	assert.DeepEqual(t, &http.BasicAuth{Username: "user", Password: "swordfish"}, a)
	os.Unsetenv("GIT_HTTP_PASSWORD")

	os.Setenv("GIT_HTTP_TOKEN", "mytoken")
	defer os.Unsetenv("GIT_HTTP_TOKEN")
	a, err = g.auth(mustParseURL("git+https://user@example.com/foo"))
	assert.NilError(t, err)
	assert.DeepEqual(t, &http.TokenAuth{Token: "mytoken"}, a)
	os.Unsetenv("GIT_HTTP_TOKEN")

	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Log("no SSH_AUTH_SOCK - skipping ssh agent test")
	} else {
		a, err = g.auth(mustParseURL("git+ssh://git@example.com/foo"))
		assert.NilError(t, err)
		sa, ok := a.(*ssh.PublicKeysCallback)
		assert.Equal(t, true, ok)
		assert.Equal(t, "git", sa.User)
	}

	key := string(testdata.PEMBytes["ed25519"])
	os.Setenv("GIT_SSH_KEY", key)
	defer os.Unsetenv("GIT_SSH_KEY")
	a, err = g.auth(mustParseURL("git+ssh://git@example.com/foo"))
	assert.NilError(t, err)
	ka, ok := a.(*ssh.PublicKeys)
	assert.Equal(t, true, ok)
	assert.Equal(t, "git", ka.User)
	os.Unsetenv("GIT_SSH_KEY")

	key = base64.StdEncoding.EncodeToString(testdata.PEMBytes["ed25519"])
	os.Setenv("GIT_SSH_KEY", key)
	defer os.Unsetenv("GIT_SSH_KEY")
	a, err = g.auth(mustParseURL("git+ssh://git@example.com/foo"))
	assert.NilError(t, err)
	ka, ok = a.(*ssh.PublicKeys)
	assert.Equal(t, true, ok)
	assert.Equal(t, "git", ka.User)
	os.Unsetenv("GIT_SSH_KEY")
}

func TestRefFromURL(t *testing.T) {
	t.Parallel()
	g := &gitRequester{}
	data := []struct {
		url, expected string
	}{
		{"git://localhost:1234/foo/bar.git//baz", ""},
		{"git+http://localhost:1234/foo/bar.git//baz", ""},
		{"git+ssh://localhost:1234/foo/bar.git//baz", ""},
		{"git+file:///foo/bar.git//baz", ""},
		{"git://localhost:1234/foo/bar.git//baz#master", "refs/heads/master"},
		{"git+http://localhost:1234/foo/bar.git//baz#mybranch", "refs/heads/mybranch"},
		{"git+ssh://localhost:1234/foo/bar.git//baz#refs/tags/foo", "refs/tags/foo"},
		{"git+file:///foo/bar.git//baz#mybranch", "refs/heads/mybranch"},
	}

	for _, d := range data {
		out := g.refFromURL(mustParseURL(d.url))
		assert.Equal(t, plumbing.ReferenceName(d.expected), out)
	}
}
