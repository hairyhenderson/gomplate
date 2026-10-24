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

	matches, err := fsys.Glob("/tmp/*.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"/tmp/one.txt", "/tmp/three.txt", "/tmp/two.txt"}, matches)
}

func TestWDFS_Glob(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tmpDir := tfs.NewDir(t, "gomplate-wdfs-glob-test")
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	var osfsys fs.FS
	vol := filepath.VolumeName(tmpDir.Path())
	if vol != "" {
		osfsys, err = osfs.NewFS().SubVolume(vol)
		require.NoError(t, err)
	} else {
		vol = "/"
		osfsys = osfs.NewFS()
	}

	fsys := &wdFS{
		vol:  vol,
		fsys: osfsys,
	}

	dirPath := filepath.Join(tmpDir.Path(), "testdir")
	require.NoError(t, os.MkdirAll(dirPath, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dirPath, "a.txt"), []byte("a"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dirPath, "b.txt"), []byte("b"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dirPath, "c.log"), []byte("c"), 0o600))

	// Absolute path pattern
	absPattern := filepath.ToSlash(filepath.Join(dirPath, "*.txt"))
	matches, err := fsys.Glob(absPattern)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.ToSlash(filepath.Join(dirPath, "a.txt")),
		filepath.ToSlash(filepath.Join(dirPath, "b.txt")),
	}, matches)

	// Native path separator pattern
	nativePattern := filepath.FromSlash(filepath.Join(dirPath, "*.txt"))
	matches, err = fsys.Glob(nativePattern)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.FromSlash(filepath.Join(dirPath, "a.txt")),
		filepath.FromSlash(filepath.Join(dirPath, "b.txt")),
	}, matches)

	// Working-directory-relative pattern
	require.NoError(t, os.Chdir(dirPath))
	matches, err = fsys.Glob("*.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt", "b.txt"}, matches)

	// Single file without wildcards
	matches, err = fsys.Glob("a.txt")
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt"}, matches)

	// Non-matching pattern
	matches, err = fsys.Glob("*.nonexistent")
	require.NoError(t, err)
	assert.Empty(t, matches)

	// Malformed pattern
	_, err = fsys.Glob("[]")
	assert.Error(t, err)
}

func TestAdaptVolumeRoot(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "C:/", adaptVolumeRoot("C:"))
	assert.Equal(t, "c:/", adaptVolumeRoot("c:"))
	assert.Equal(t, "D:/", adaptVolumeRoot("D:"))
	assert.Equal(t, "C:/foo", adaptVolumeRoot("C:/foo"))
	assert.Equal(t, "/tmp", adaptVolumeRoot("/tmp"))
	assert.Equal(t, ".", adaptVolumeRoot("."))
}

func TestWDFS_Glob_StatWithoutRead(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("chmod 0200 does not restrict read on Windows")
	}

	dir := t.TempDir()
	filePath := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("secret"), 0o200))

	fsys := &wdFS{
		vol:  "/",
		fsys: osfs.NewFS(),
	}

	matches, err := fsys.Glob(filePath)
	require.NoError(t, err)
	assert.Equal(t, []string{filePath}, matches)
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
	require.ErrorIs(t, err, fs.ErrNotExist)

	// make sure we can write to a subfs
	subfs, err := fs.Sub(fsys, "tmp")
	require.NoError(t, err)
	require.NotNil(t, subfs)

	// this is no longer a wdFS so we need to make sure not to use absolute
	// paths - the path is relative to the root of the subfs
	err = hackpadfs.WriteFullFile(subfs, "foo", []byte("hello world"), 0o600)
	require.NoError(t, err)

	b, err = fs.ReadFile(subfs, "foo")
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(b))
}

func skipWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == osWindows {
		t.Skip("skipping non-Windows test")
	}
}

func skipNonWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS != osWindows {
		t.Skip("skipping Windows test")
	}
}

func TestResolveLocalPath_NonWindows(t *testing.T) {
	skipWindows(t)

	wd, _ := os.Getwd()
	fsys := &wdFS{vol: "/", fsys: osfs.NewFS()}

	wd = wd[1:]

	testdata := []struct {
		path     string
		expected string
	}{
		{"/tmp/foo", "tmp/foo"},
		{"tmp/foo", wd + "/tmp/foo"},
		{"./tmp/foo", wd + "/tmp/foo"},
		{"tmp/../foo", wd + "/foo"},
		{"/", "."},
	}

	for _, td := range testdata {
		t.Run(td.path, func(t *testing.T) {
			root, path, err := ResolveLocalPath(fsys, td.path)
			require.NoError(t, err)
			assert.Equal(t, "/", root)
			assert.Equal(t, td.expected, path)
		})
	}
}

func TestResolveLocalPath_Windows(t *testing.T) {
	skipNonWindows(t)

	wd, _ := os.Getwd()
	volname := filepath.VolumeName(wd)
	wd = filepath.ToSlash(wd)

	fsys := &wdFS{vol: volname, fsys: osfs.NewFS()}

	wd = wd[len(volname)+1:]

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
		t.Run(td.path, func(t *testing.T) {
			root, path, err := ResolveLocalPath(fsys, td.path)
			require.NoError(t, err)
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
		{"/", "."},
	}

	for _, td := range testdata {
		root, path, err := resolveLocalPath("/", td.path)
		require.NoError(t, err)
		assert.Equal(t, "/", root)
		assert.Equal(t, td.expected, path)
	}
}

func TestWdFS_ResolveLocalPath_Windows(t *testing.T) {
	skipNonWindows(t)

	wd, _ := os.Getwd()
	volname := filepath.VolumeName(wd)
	wd = filepath.ToSlash(wd)
	wd = wd[len(volname)+1:]

	testdata := []struct {
		path     string
		expRoot  string
		expected string
	}{
		{"C:/tmp/foo", "C:", "tmp/foo"},
		{`D:\tmp\foo`, "D:", "tmp/foo"},
		{"/tmp/foo", volname, "tmp/foo"},
		{"tmp/foo", volname, wd + "/tmp/foo"},
		{"./tmp/foo", volname, wd + "/tmp/foo"},
		{"tmp/../foo", volname, wd + "/foo"},
		{`\\?\C:\tmp\foo`, "C:", "tmp/foo"},
		{`\\somehost\share\foo\bar`, "//somehost/share", "foo/bar"},
		{`//?/C:/tmp/foo`, "C:", "tmp/foo"},
		{`//somehost/share/foo/bar`, "//somehost/share", "foo/bar"},
	}

	for _, td := range testdata {
		t.Run(td.path, func(t *testing.T) {
			root, path, err := resolveLocalPath(volname, td.path)
			require.NoError(t, err)
			assert.Equal(t, td.expRoot, root)
			assert.Equal(t, td.expected, path)
		})
	}
}

func TestWin32PathType(t *testing.T) {
	testdata := []struct {
		path     string
		expected winPathtype
	}{
		{"", winPathUnknown},
		{`\`, winPathRooted},
		{`\\`, winPathUncAbsolute},
		{`x`, winPathRelative},
		{`x:`, winPathDriveRelative},
		{"C:/tmp/foo", winPathDriveAbsolute},
		{`D:\tmp\foo`, winPathDriveAbsolute},
		{"/tmp/foo", winPathRooted},
		{"tmp/foo", winPathRelative},
		{"./tmp/foo", winPathRelative},
		{"tmp/../foo", winPathRelative},
		{`\\?\C:\tmp\foo`, winPathLocalDevice},
		{`\\somehost\share\foo\bar`, winPathUncAbsolute},
		{`//./C:/tmp/foo`, winPathLocalDevice},
		{`//./pipe/foo`, winPathLocalDevice},
		{`//./COM2`, winPathLocalDevice},
		{`\\.`, winPathRootLocalDevice},
		{`//?`, winPathRootLocalDevice},
		{`/??/C:/`, winPathNT},
		{`/??/UNC/server/foo`, winPathNT},
	}

	for _, td := range testdata {
		t.Run(td.path, func(t *testing.T) {
			assert.Equal(t, td.expected, win32PathType(td.path))
		})
	}
}

func TestWorkingVolume(t *testing.T) {
	vol, err := workingVolume()
	require.NoError(t, err)

	if runtime.GOOS != osWindows {
		assert.Equal(t, "/", vol)
		return
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, filepath.VolumeName(wd), vol)
}

var (
	resolveSinkRoot string
	resolveSinkPath string
)

func BenchmarkResolveLocalPath(b *testing.B) {
	name := "/tmp/foo/bar.txt"
	if runtime.GOOS == osWindows {
		name = `C:\tmp\foo\bar.txt`
	}

	for b.Loop() {
		resolveSinkRoot, resolveSinkPath, _ = ResolveLocalPath(nil, name)
	}
}
