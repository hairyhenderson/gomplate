// +build !windows

package gomplate

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

// like ioutil.NopCloser(), except for io.WriteClosers...
type nopWCloser struct {
	io.Writer
}

func (n *nopWCloser) Close() error {
	return nil
}
func TestReadInput(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	f, _ = fs.Create("/tmp/unreadable")
	_, _ = f.Write([]byte("foo"))

	actual, err := readInput("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", actual)

	defer func() { stdin = os.Stdin }()
	stdin = ioutil.NopCloser(bytes.NewBufferString("bar"))

	actual, err = readInput("-")
	assert.NoError(t, err)
	assert.Equal(t, "bar", actual)

	_, err = readInput("bogus")
	assert.Error(t, err)
}

func TestOpenOutFile(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)

	_, err := openOutFile("/tmp/foo", 0644)
	assert.NoError(t, err)
	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), i.Mode())

	defer func() { stdout = os.Stdout }()
	stdout = &nopWCloser{&bytes.Buffer{}}

	f, err := openOutFile("-", 0644)
	assert.NoError(t, err)
	assert.Equal(t, stdout, f)
}

func TestOpenOutFileWithNoneDefaultPerms(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)

	_, err := openOutFile("/tmp/foo", 0444)
	assert.NoError(t, err)
	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0444), i.Mode())

	defer func() { stdout = os.Stdout }()
	stdout = &nopWCloser{&bytes.Buffer{}}

	f, err := openOutFile("-", 0444)
	assert.NoError(t, err)
	assert.Equal(t, stdout, f)
}

func TestOpenOutFileWithExecuteBit(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)

	_, err := openOutFile("/tmp/foo", 0544)
	assert.NoError(t, err)
	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0544), i.Mode())

	defer func() { stdout = os.Stdout }()
	stdout = &nopWCloser{&bytes.Buffer{}}

	f, err := openOutFile("-", 0544)
	assert.NoError(t, err)
	assert.Equal(t, stdout, f)
}

func TestInList(t *testing.T) {
	list := []string{}
	assert.False(t, inList(list, ""))

	list = nil
	assert.False(t, inList(list, ""))

	list = []string{"foo", "baz", "qux"}
	assert.False(t, inList(list, "bar"))

	list = []string{"foo", "bar", "baz"}
	assert.True(t, inList(list, "bar"))
}

func TestExecuteCombinedGlob(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.MkdirAll("/tmp/one", 0777)
	_ = fs.MkdirAll("/tmp/two", 0777)
	_ = fs.MkdirAll("/tmp/three", 0777)
	afero.WriteFile(fs, "/tmp/one/a", []byte("file a"), 0644)
	afero.WriteFile(fs, "/tmp/two/b", []byte("file b"), 0644)
	afero.WriteFile(fs, "/tmp/three/c", []byte("file c"), 0644)

	excludes, err := executeCombinedGlob([]string{"/tmp/o*/*", "/*/*/b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"/tmp/one/a", "/tmp/two/b"}, excludes)
}

func TestWalkDir(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()

	_, _, err := walkDir("/indir", "/outdir", nil)
	assert.Error(t, err)

	_ = fs.MkdirAll("/indir/one", 0777)
	_ = fs.MkdirAll("/indir/two", 0777)
	afero.WriteFile(fs, "/indir/one/foo", []byte("foo"), 0644)
	afero.WriteFile(fs, "/indir/one/bar", []byte("bar"), 0644)
	afero.WriteFile(fs, "/indir/two/baz", []byte("baz"), 0644)

	in, out, err := walkDir("/indir", "/outdir", []string{"/*/two"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"/indir/one/bar", "/indir/one/foo"}, in)
	assert.Equal(t, []*outFile{&outFile{"/outdir/one/bar", 0644}, &outFile{"/outdir/one/foo", 0644}}, out)
}

func TestLoadContents(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()

	afero.WriteFile(fs, "foo", []byte("contents"), 0644)

	tmpl := &tplate{name: "foo"}
	err := tmpl.loadContents()
	assert.NoError(t, err)
	assert.Equal(t, "contents", tmpl.contents)
}

func TestAddTarget(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()

	tmpl := &tplate{name: "foo"}
	err := tmpl.addTarget("/out/outfile", 0644)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl.target)
}

func TestGatherTemplates(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo", []byte("bar"), 0644)

	afero.WriteFile(fs, "in/1", []byte("foo"), 0744)
	afero.WriteFile(fs, "in/2", []byte("bar"), 0754)
	afero.WriteFile(fs, "in/3", []byte("baz"), 0640)

	templates, err := gatherTemplates(&Config{})
	assert.NoError(t, err)
	assert.Len(t, templates, 0)

	templates, err = gatherTemplates(&Config{
		Input: "foo",
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].contents)
	assert.Equal(t, stdout, templates[0].target)

	templates, err = gatherTemplates(&Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].contents)
	assert.NotEqual(t, stdout, templates[0].target)

	templates, err = gatherTemplates(&Config{
		InputDir:  "in",
		OutputDir: "out",
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].contents)

	f1Stat, err := fs.Stat("in/1")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0744), f1Stat.Mode().Perm())

	f2Stat, err := fs.Stat("in/2")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0754), f2Stat.Mode().Perm())

	f3Stat, err := fs.Stat("in/3")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0640), f3Stat.Mode().Perm())
}
