package data

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/server"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"

	"golang.org/x/crypto/ssh/testdata"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParseArgPath(t *testing.T) {
	t.Parallel()
	g := gitsource{}

	data := []struct {
		url        string
		arg        string
		repo, path string
	}{
		{"git+file:///foo//foo",
			"/bar",
			"", "/bar"},
		{"git+file:///foo//bar",
			"/baz//qux",
			"", "/baz//qux"},
		{"git+https://example.com/foo",
			"/bar",
			"/bar", ""},
		{"git+https://example.com/foo",
			"//bar",
			"", "//bar"},
		{"git+https://example.com/foo//bar",
			"//baz",
			"", "//baz"},
		{"git+https://example.com/foo",
			"/bar//baz",
			"/bar", "/baz"},
		{"git+https://example.com/foo?type=t",
			"/bar//baz",
			"/bar", "/baz"},
		{"git+https://example.com/foo#master",
			"/bar//baz",
			"/bar", "/baz"},
		{"git+https://example.com/foo",
			"//bar",
			"", "//bar"},
		{"git+https://example.com/foo?type=t",
			"//baz",
			"", "//baz"},
		{"git+https://example.com/foo?type=t#v1",
			"//bar",
			"", "//bar"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:(%q,%q)==(%q,%q)", i, d.url, d.arg, d.repo, d.path), func(t *testing.T) {
			t.Parallel()
			u, _ := url.Parse(d.url)
			repo, path := g.parseArgPath(u, d.arg)
			assert.Equal(t, d.repo, repo)
			assert.Equal(t, d.path, path)
		})
	}
}

func TestParseGitPath(t *testing.T) {
	t.Parallel()
	g := gitsource{}
	_, _, err := g.parseGitPath(nil)
	assert.ErrorContains(t, err, "")

	u := mustParseURL("http://example.com//foo")
	assert.Equal(t, "//foo", u.Path)
	parts := strings.SplitN(u.Path, "//", 2)
	assert.Equal(t, 2, len(parts))
	assert.DeepEqual(t, []string{"", "foo"}, parts)

	data := []struct {
		url        string
		args       []string
		repo, path string
	}{
		{"git+https://github.com/hairyhenderson/gomplate//docs-src/content/functions/aws.yml",
			nil,
			"git+https://github.com/hairyhenderson/gomplate",
			"/docs-src/content/functions/aws.yml"},
		{"git+ssh://github.com/hairyhenderson/gomplate.git",
			nil,
			"git+ssh://github.com/hairyhenderson/gomplate.git",
			"/"},
		{"https://github.com",
			nil,
			"https://github.com",
			"/"},
		{"git://example.com/foo//file.txt#someref",
			nil,
			"git://example.com/foo#someref", "/file.txt"},
		{"git+file:///home/foo/repo//file.txt#someref",
			nil,
			"git+file:///home/foo/repo#someref", "/file.txt"},
		{"git+file:///repo",
			nil,
			"git+file:///repo", "/"},
		{"git+file:///foo//foo",
			nil,
			"git+file:///foo", "/foo"},
		{"git+file:///foo//foo",
			[]string{"/bar"},
			"git+file:///foo", "/foo/bar"},
		{"git+file:///foo//bar",
			// in this case the // is meaningless
			[]string{"/baz//qux"},
			"git+file:///foo", "/bar/baz/qux"},
		{"git+https://example.com/foo",
			[]string{"/bar"},
			"git+https://example.com/foo/bar", "/"},
		{"git+https://example.com/foo",
			[]string{"//bar"},
			"git+https://example.com/foo", "/bar"},
		{"git+https://example.com//foo",
			[]string{"/bar"},
			"git+https://example.com", "/foo/bar"},
		{"git+https://example.com/foo//bar",
			[]string{"//baz"},
			"git+https://example.com/foo", "/bar/baz"},
		{"git+https://example.com/foo",
			[]string{"/bar//baz"},
			"git+https://example.com/foo/bar", "/baz"},
		{"git+https://example.com/foo?type=t",
			[]string{"/bar//baz"},
			"git+https://example.com/foo/bar?type=t", "/baz"},
		{"git+https://example.com/foo#master",
			[]string{"/bar//baz"},
			"git+https://example.com/foo/bar#master", "/baz"},
		{"git+https://example.com/foo",
			[]string{"/bar//baz?type=t"},
			"git+https://example.com/foo/bar?type=t", "/baz"},
		{"git+https://example.com/foo",
			[]string{"/bar//baz#master"},
			"git+https://example.com/foo/bar#master", "/baz"},
		{"git+https://example.com/foo",
			[]string{"//bar?type=t"},
			"git+https://example.com/foo?type=t", "/bar"},
		{"git+https://example.com/foo",
			[]string{"//bar#master"},
			"git+https://example.com/foo#master", "/bar"},
		{"git+https://example.com/foo?type=t",
			[]string{"//bar#master"},
			"git+https://example.com/foo?type=t#master", "/bar"},
		{"git+https://example.com/foo?type=t#v1",
			[]string{"//bar?type=j#v2"},
			"git+https://example.com/foo?type=t&type=j#v2", "/bar"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:(%q,%q)==(%q,%q)", i, d.url, d.args, d.repo, d.path), func(t *testing.T) {
			t.Parallel()
			u, _ := url.Parse(d.url)
			repo, path, err := g.parseGitPath(u, d.args...)
			assert.NilError(t, err)
			assert.Equal(t, d.repo, repo.String())
			assert.Equal(t, d.path, path)
		})
	}
}

func TestReadGitRepo(t *testing.T) {
	g := gitsource{}
	fs := setupGitRepo(t)
	fs, err := fs.Chroot("/repo")
	assert.NilError(t, err)

	_, _, err = g.read(fs, "/bogus")
	assert.ErrorContains(t, err, "can't stat /bogus")

	mtype, out, err := g.read(fs, "/foo")
	assert.NilError(t, err)
	assert.Equal(t, `["bar"]`, string(out))
	assert.Equal(t, jsonArrayMimetype, mtype)

	mtype, out, err = g.read(fs, "/foo/bar")
	assert.NilError(t, err)
	assert.Equal(t, `["hi.txt"]`, string(out))
	assert.Equal(t, jsonArrayMimetype, mtype)

	mtype, out, err = g.read(fs, "/foo/bar/hi.txt")
	assert.NilError(t, err)
	assert.Equal(t, `hello world`, string(out))
	assert.Equal(t, "", mtype)
}

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
	ctx := context.TODO()
	repoFS := setupGitRepo(t)
	g := gitsource{}

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
	ctx := context.TODO()
	repoFS := setupGitRepo(t)
	g := gitsource{}

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

	s := &Source{
		Alias: "hi",
		URL:   mustParseURL("git+file:///bare.git//hello.txt"),
	}
	b, err := readGit(s)
	assert.NilError(t, err)
	assert.Equal(t, "hello world", string(b))

	s = &Source{
		Alias: "hi",
		URL:   mustParseURL("git+file:///bare.git"),
	}
	b, err = readGit(s)
	assert.NilError(t, err)
	assert.Equal(t, "application/array+json", s.mediaType)
	assert.Equal(t, `["hello.txt"]`, string(b))
}

func TestGitAuth(t *testing.T) {
	g := gitsource{}
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
	g := gitsource{}
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
