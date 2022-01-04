package iohelpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllWhitespace(t *testing.T) {
	testdata := []struct {
		in       []byte
		expected bool
	}{
		{[]byte(" "), true},
		{[]byte("foo"), false},
		{[]byte("   \t\n\n\v\r\n"), true},
		{[]byte("   foo   "), false},
	}

	for _, d := range testdata {
		assert.Equal(t, d.expected, allWhitespace(d.in))
	}
}

func TestEmptySkipper(t *testing.T) {
	testdata := []struct {
		in    []byte
		empty bool
	}{
		{[]byte(" "), true},
		{[]byte("foo"), false},
		{[]byte("   \t\n\n\v\r\n"), true},
		{[]byte("   foo   "), false},
	}

	for _, d := range testdata {
		w := newBufferCloser(&bytes.Buffer{})
		opened := false
		f, ok := NewEmptySkipper(func() (io.Writer, error) {
			opened = true
			return w, nil
		}).(*emptySkipper)

		assert.True(t, ok)
		n, err := f.Write(d.in)
		require.NoError(t, err)
		assert.Equal(t, len(d.in), n)
		err = f.Close()
		require.NoError(t, err)
		if d.empty {
			assert.Nil(t, f.w)
			assert.False(t, opened)
		} else {
			assert.NotNil(t, f.w)
			assert.True(t, opened)
			assert.EqualValues(t, d.in, w.Bytes())
		}
	}
}

func newBufferCloser(b *bytes.Buffer) *bufferCloser {
	return &bufferCloser{b, false}
}

type bufferCloser struct {
	*bytes.Buffer

	closed bool
}

func (b *bufferCloser) Close() error {
	b.closed = true
	return nil
}

func TestSameSkipper(t *testing.T) {
	testdata := []struct {
		in   []byte
		out  []byte
		same bool
	}{
		{[]byte(" "), []byte(" "), true},
		{[]byte("foo"), []byte("foo"), true},
		{[]byte("foo"), nil, false},
		{[]byte("foo"), []byte("bar"), false},
		{[]byte("foobar"), []byte("foo"), false},
		{[]byte("foo"), []byte("foobar"), false},
	}

	for _, d := range testdata {
		d := d
		t.Run(fmt.Sprintf("in:%q/out:%q/same:%v", d.in, d.out, d.same), func(t *testing.T) {
			r := bytes.NewBuffer(d.out)
			w := newBufferCloser(&bytes.Buffer{})
			opened := false
			f, ok := SameSkipper(r, func() (io.WriteCloser, error) {
				opened = true
				return w, nil
			}).(*sameSkipper)
			assert.True(t, ok)

			n, err := f.Write(d.in)
			require.NoError(t, err)
			assert.Equal(t, len(d.in), n)
			err = f.Close()
			require.NoError(t, err)
			if d.same {
				assert.Nil(t, f.w)
				assert.False(t, opened)
				assert.Empty(t, w.Bytes())
			} else {
				assert.NotNil(t, f.w)
				assert.True(t, opened)
				assert.EqualValues(t, d.in, w.Bytes())
			}
		})
	}
}

func TestLazyWriteCloser(t *testing.T) {
	w := newBufferCloser(&bytes.Buffer{})
	opened := false
	l, ok := LazyWriteCloser(func() (io.WriteCloser, error) {
		opened = true
		return w, nil
	}).(*lazyWriteCloser)
	assert.True(t, ok)

	assert.False(t, opened)
	assert.Nil(t, l.w)
	assert.False(t, w.closed)

	p := []byte("hello world")
	n, err := l.Write(p)
	require.NoError(t, err)
	assert.True(t, opened)
	assert.Equal(t, 11, n)

	err = l.Close()
	require.NoError(t, err)
	assert.True(t, w.closed)

	// test error propagation
	l = LazyWriteCloser(func() (io.WriteCloser, error) {
		return nil, os.ErrNotExist
	}).(*lazyWriteCloser)

	assert.Nil(t, l.w)

	p = []byte("hello world")
	_, err = l.Write(p)
	assert.Error(t, err)

	err = l.Close()
	assert.Error(t, err)
}

// TODO: uncomment this and fix the import cycle!
// func TestWrite(t *testing.T) {
// 	oldwd, _ := os.Getwd()
// 	defer os.Chdir(oldwd)

// 	rootDir := tfs.NewDir(t, "gomplate-test")
// 	t.Cleanup(rootDir.Remove)

// 	// we want to use a real filesystem here, so we can test interactions with
// 	// the current working directory
// 	fsys := datafs.WrapWdFS(osfs.NewFS())

// 	newwd := rootDir.Join("the", "path", "we", "want")
// 	badwd := rootDir.Join("some", "other", "dir")
// 	hackpadfs.MkdirAll(fsys, newwd, 0o755)
// 	hackpadfs.MkdirAll(fsys, badwd, 0o755)
// 	newwd, _ = filepath.EvalSymlinks(newwd)
// 	badwd, _ = filepath.EvalSymlinks(badwd)

// 	err := os.Chdir(newwd)
// 	require.NoError(t, err)

// 	err = WriteFile(fsys, "/foo", []byte("Hello world"))
// 	assert.Error(t, err)

// 	rel, err := filepath.Rel(newwd, badwd)
// 	require.NoError(t, err)
// 	err = WriteFile(fsys, rel, []byte("Hello world"))
// 	assert.Error(t, err)

// 	foopath := filepath.Join(newwd, "foo")
// 	err = WriteFile(fsys, foopath, []byte("Hello world"))
// 	require.NoError(t, err)

// 	out, err := fs.ReadFile(fsys, foopath)
// 	require.NoError(t, err)
// 	assert.Equal(t, "Hello world", string(out))

// 	err = WriteFile(fsys, foopath, []byte("truncate"))
// 	require.NoError(t, err)

// 	out, err = fs.ReadFile(fsys, foopath)
// 	require.NoError(t, err)
// 	assert.Equal(t, "truncate", string(out))

// 	foopath = filepath.Join(newwd, "nonexistant", "subdir", "foo")
// 	err = WriteFile(fsys, foopath, []byte("Hello subdirranean world!"))
// 	require.NoError(t, err)

// 	out, err = fs.ReadFile(fsys, foopath)
// 	require.NoError(t, err)
// 	assert.Equal(t, "Hello subdirranean world!", string(out))
// }

func TestAssertPathInWD(t *testing.T) {
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)

	err := assertPathInWD("/tmp")
	assert.Error(t, err)

	err = assertPathInWD(filepath.Join(oldwd, "subpath"))
	require.NoError(t, err)

	err = assertPathInWD("subpath")
	require.NoError(t, err)

	err = assertPathInWD("./subpath")
	require.NoError(t, err)

	err = assertPathInWD(filepath.Join("..", "bogus"))
	assert.Error(t, err)

	err = assertPathInWD(filepath.Join("..", "..", "bogus"))
	assert.Error(t, err)

	base := filepath.Base(oldwd)
	err = assertPathInWD(filepath.Join("..", base))
	require.NoError(t, err)
}
