package datafs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tfs "gotest.tools/v3/fs"
)

func TestWDFS_ReadOps(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	memfs, _ := mem.NewFS()

	_ = memfs.Mkdir("tmp", 0o777)
	_ = hackpadfs.WriteFullFile(memfs, "tmp/foo", []byte("hello world"), 0o777)
	_ = hackpadfs.WriteFullFile(memfs, "tmp/one.txt", []byte("one"), 0o644)
	_ = hackpadfs.WriteFullFile(memfs, "tmp/two.txt", []byte("two"), 0o644)
	_ = hackpadfs.WriteFullFile(memfs, "tmp/three.txt", []byte("three"), 0o644)
	_ = memfs.Mkdir("tmp/sub", 0o777)
	_ = hackpadfs.WriteFullFile(memfs, "tmp/sub/bar", []byte("goodnight moon"), 0o777)

	fsys := WrapWdFS(memfs).(*wdFS)

	f, err := fsys.Open("/tmp/foo")
	require.NoError(t, err)

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(b))

	fi, err := fs.Stat(fsys, "/tmp/sub/bar")
	require.NoError(t, err)
	assert.True(t, fi.Mode().IsRegular())

	b, err = fs.ReadFile(fsys, "/tmp/sub/bar")
	require.NoError(t, err)
	assert.Equal(t, "goodnight moon", string(b))

	des, err := fs.ReadDir(fsys, "/tmp")
	require.NoError(t, err)
	assert.Len(t, des, 5)

	// note the relative path here, a requirement of fsys.Sub
	subfs, err := fs.Sub(fsys, "tmp/sub")
	require.NoError(t, err)

	b, err = fs.ReadFile(subfs, "bar")
	require.NoError(t, err)
	assert.Equal(t, "goodnight moon", string(b))
}

func TestWDFS_WriteOps(t *testing.T) {
	// this test is backed by the real filesystem so we can test permissions
	// and have some confidence it'll run on Windows
	tmpDir := tfs.NewDir(t, "gomplate-wdfs-test")
	tmpPath := tmpDir.Path()
	vol := filepath.VolumeName(tmpPath)
	if vol != "" {
		tmpPath = tmpPath[len(vol):]
	} else if tmpPath[0] == '/' {
		vol = "/"
		tmpPath = tmpPath[1:]
	}

	var osfsys fs.FS
	var err error
	if vol != "/" {
		osfsys, err = osfs.NewFS().SubVolume(vol)
		require.NoError(t, err)
	} else {
		osfsys = osfs.NewFS()
	}

	osfsys, err = hackpadfs.Sub(osfsys, tmpPath)
	require.NoError(t, err)

	fsys := &wdFS{
		vol:  vol,
		fsys: osfsys,
	}

	err = fsys.Mkdir("/tmp", 0o700)
	require.NoError(t, err, "failed to create /tmp: %q", tmpDir.Path())

	// use os.Stat to make sure the directory was created in the right place
	fi, err := os.Stat(filepath.Join(vol, tmpPath, "tmp"))
	require.NoError(t, err)
	assert.True(t, fi.Mode().IsDir())

	err = hackpadfs.WriteFullFile(fsys, "/tmp/foo", []byte("hello world"), 0o600)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/tmp/one.txt", []byte("one"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/tmp/two.txt", []byte("two"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/tmp/three.txt", []byte("three"), 0o644)
	require.NoError(t, err)

	err = fsys.MkdirAll("/tmp/sub", 0o777)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/tmp/sub/bar", []byte("goodnight moon"), 0o777)
	require.NoError(t, err)

	b, err := fs.ReadFile(fsys, "/tmp/foo")
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(b))

	b, err = fs.ReadFile(fsys, "/tmp/sub/bar")
	require.NoError(t, err)
	assert.Equal(t, "goodnight moon", string(b))

	err = fsys.Chmod("/tmp/foo", 0o444)
	require.NoError(t, err)

	// check permissions
	fi, err = fsys.Stat("/tmp/foo")
	require.NoError(t, err)
	assert.True(t, fi.Mode().IsRegular())
	assert.Equal(t, "0444", fmt.Sprintf("%#o", fi.Mode().Perm()))

	// now delete it
	err = fsys.Remove("/tmp/foo")
	require.NoError(t, err)

	// and check that it's gone
	_, err = fsys.Stat("/tmp/foo")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func skipWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping non-Windows test")
	}
}

func skipNonWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows test")
	}
}

func TestResolveLocalPath_NonWindows(t *testing.T) {
	skipWindows(t)

	wd, _ := os.Getwd()
	wd = wd[1:]

	testdata := []struct {
		path     string
		expected string
	}{
		{"/tmp/foo", "tmp/foo"},
		{"tmp/foo", wd + "/tmp/foo"},
		{"./tmp/foo", wd + "/tmp/foo"},
		{"tmp/../foo", wd + "/foo"},
	}

	for _, td := range testdata {
		root, path := ResolveLocalPath(td.path)
		assert.Equal(t, "/", root)
		assert.Equal(t, td.expected, path)
	}
}

func TestResolveLocalPath_Windows(t *testing.T) {
	skipNonWindows(t)

	wd, _ := os.Getwd()
	volname := filepath.VolumeName(wd)
	wd = wd[len(volname)+1:]
	wd = filepath.ToSlash(wd)

	testdata := []struct {
		path     string
		expRoot  string
		expected string
	}{
		{"C:/tmp/foo", "C:", "tmp/foo"},
		{"D:\\tmp\\foo", "D:", "tmp/foo"},
		{"/tmp/foo", volname, "tmp/foo"},
		{"tmp/foo", volname, wd + "/tmp/foo"},
		{"./tmp/foo", volname, wd + "/tmp/foo"},
		{"tmp/../foo", volname, wd + "/foo"},
	}

	for _, td := range testdata {
		td := td
		t.Run(td.path, func(t *testing.T) {
			root, path := ResolveLocalPath(td.path)
			assert.Equal(t, td.expRoot, root)
			assert.Equal(t, td.expected, path)
		})
	}
}

func TestWdFS_ResolveLocalPath_NonWindows(t *testing.T) {
	skipWindows(t)

	wd, _ := os.Getwd()
	wd = wd[1:]

	testdata := []struct {
		path     string
		expected string
	}{
		{"/tmp/foo", "tmp/foo"},
		{"tmp/foo", wd + "/tmp/foo"},
		{"./tmp/foo", wd + "/tmp/foo"},
		{"tmp/../foo", wd + "/foo"},
	}

	fsys := &wdFS{}

	for _, td := range testdata {
		root, path := fsys.resolveLocalPath(td.path)
		assert.Equal(t, "/", root)
		assert.Equal(t, td.expected, path)
	}
}

func TestWdFS_ResolveLocalPath_Windows(t *testing.T) {
	skipNonWindows(t)

	wd, _ := os.Getwd()
	volname := filepath.VolumeName(wd)
	wd = wd[len(volname)+1:]
	wd = filepath.ToSlash(wd)

	testdata := []struct {
		path     string
		expRoot  string
		expected string
	}{
		{"C:/tmp/foo", "C:", "tmp/foo"},
		{"D:\\tmp\\foo", "D:", "tmp/foo"},
		{"/tmp/foo", volname, "tmp/foo"},
		{"tmp/foo", volname, wd + "/tmp/foo"},
		{"./tmp/foo", volname, wd + "/tmp/foo"},
		{"tmp/../foo", volname, wd + "/foo"},
		{`\\?\C:\tmp\foo`, "//?/C:", "tmp/foo"},
		{`\\somehost\share\foo\bar`, "//somehost/share", "foo/bar"},
		{`//?/C:/tmp/foo`, "//?/C:", "tmp/foo"},
		{`//somehost/share/foo/bar`, "//somehost/share", "foo/bar"},
	}

	fsys := &wdFS{vol: volname}

	for _, td := range testdata {
		td := td
		t.Run(td.path, func(t *testing.T) {
			root, path := fsys.resolveLocalPath(td.path)
			assert.Equal(t, td.expRoot, root)
			assert.Equal(t, td.expected, path)
		})
	}
}
