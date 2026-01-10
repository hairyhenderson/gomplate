package gomplate

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/url"
	"os"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenOutFile(t *testing.T) {
	memfs, _ := mem.NewFS()
	fsys := datafs.WrapWdFS(memfs)

	_ = hackpadfs.Mkdir(fsys, "/tmp", 0o777)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	f, err := openOutFile(ctx, "/tmp/foo", 0o755, 0o644, false, nil)
	require.NoError(t, err)

	_, err = f.Write([]byte("hello world"))
	require.NoError(t, err)

	wc, ok := f.(io.WriteCloser)
	assert.True(t, ok)
	err = wc.Close()
	require.NoError(t, err)

	i, err := hackpadfs.Stat(fsys, "/tmp/foo")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o644), i.Mode())

	out := &bytes.Buffer{}

	f, err = openOutFile(ctx, "-", 0o755, 0o644, false, out)
	require.NoError(t, err)

	_, err = f.Write([]byte("hello world"))
	require.NoError(t, err)
	assert.Equal(t, "hello world", out.String())
}

func TestGatherTemplates(t *testing.T) {
	// chdir to root so we can use relative paths
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	fsys, _ := mem.NewFS()

	_ = hackpadfs.WriteFullFile(fsys, "foo", []byte("bar"), 0o600)
	_ = hackpadfs.Mkdir(fsys, "in", 0o777)
	_ = hackpadfs.WriteFullFile(fsys, "in/1", []byte("foo"), 0o644)
	_ = hackpadfs.WriteFullFile(fsys, "in/2", []byte("bar"), 0o644)
	_ = hackpadfs.WriteFullFile(fsys, "in/3", []byte("baz"), 0o644)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	cfg := &Config{
		Stdin:  &bytes.Buffer{},
		Stdout: &bytes.Buffer{},
	}
	cfg.applyDefaults()
	templates, err := gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)

	buf := &bytes.Buffer{}
	cfg = &Config{
		Input:  "foo",
		Stdout: buf,
	}
	cfg.applyDefaults()
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].Text)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)
	assert.Equal(t, "hello world", buf.String())

	templates, err = gatherTemplates(ctx, &Config{
		Input:       "foo",
		OutputFiles: []string{"out"},
	}, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)

	// out file is created only on demand
	_, err = hackpadfs.Stat(fsys, "out")
	require.ErrorIs(t, err, fs.ErrNotExist)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)

	info, err := hackpadfs.Stat(fsys, "out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o644), info.Mode())
	_ = hackpadfs.Remove(fsys, "out")

	buf = &bytes.Buffer{}
	cfg = &Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		Stdout:      buf,
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].Text)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)
	// negative test - we should not be writing to stdout
	assert.NotEqual(t, "hello world", buf.String())

	info, err = hackpadfs.Stat(fsys, "out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o600), info.Mode())
	hackpadfs.Remove(fsys, "out")

	buf = &bytes.Buffer{}
	cfg = &Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		OutMode:     "755",
		Stdout:      buf,
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].Text)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)
	// negative test - we should not be writing to stdout
	assert.NotEqual(t, "hello world", buf.String())

	info, err = hackpadfs.Stat(fsys, "out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o755), info.Mode())
	hackpadfs.Remove(fsys, "out")

	templates, err = gatherTemplates(ctx, &Config{
		InputDir:  "in",
		OutputDir: "out",
	}, simpleNamer("out"))
	require.NoError(t, err)
	require.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].Text)
	hackpadfs.Remove(fsys, "out")
}

func TestCreateOutFile(t *testing.T) {
	fsys, _ := mem.NewFS()
	_ = hackpadfs.Mkdir(fsys, "in", 0o755)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	_, err := createOutFile(ctx, "in", 0o755, 0o644, false)
	require.Error(t, err)

	var pathErr *fs.PathError
	assert.ErrorAs(t, err, &pathErr)
}

func TestParseNestedTemplates(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	// in-memory test filesystem
	fsys := fstest.MapFS{
		"foo.t": {Data: []byte("hello world"), Mode: 0o600},
	}

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	// simple test with single template
	u, _ := url.Parse("foo.t")
	nested := map[string]DataSource{"foo": {URL: u}}

	tmpl, _ := template.New("root").Parse(`{{ template "foo" }}`)

	r := &renderer{nested: nested}

	err := r.parseNestedTemplates(ctx, tmpl)
	require.NoError(t, err)

	out := bytes.Buffer{}
	err = tmpl.Execute(&out, nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", out.String())

	// test with directory of templates
	fsys["dir/"] = &fstest.MapFile{Mode: 0o777 | os.ModeDir}
	fsys["dir/foo.t"] = &fstest.MapFile{Data: []byte("foo"), Mode: 0o600}
	fsys["dir/bar.t"] = &fstest.MapFile{Data: []byte("bar"), Mode: 0o600}

	u, _ = url.Parse("dir/")
	nested["dir"] = DataSource{URL: u}

	tmpl, _ = template.New("root").Parse(`{{ template "dir/foo.t" }} {{ template "dir/bar.t" }}`)

	r = &renderer{nested: nested}
	err = r.parseNestedTemplates(ctx, tmpl)
	require.NoError(t, err)

	out = bytes.Buffer{}
	err = tmpl.Execute(&out, nil)
	require.NoError(t, err)
	assert.Equal(t, "foo bar", out.String())
}
