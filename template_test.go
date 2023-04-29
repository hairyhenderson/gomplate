package gomplate

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenOutFile(t *testing.T) {
	origfs := aferoFS
	defer func() { aferoFS = origfs }()
	aferoFS = afero.NewMemMapFs()
	_ = aferoFS.Mkdir("/tmp", 0777)

	cfg := &config.Config{Stdout: &bytes.Buffer{}}
	f, err := openOutFile("/tmp/foo", 0755, 0644, false, nil, false)
	require.NoError(t, err)

	wc, ok := f.(io.WriteCloser)
	assert.True(t, ok)
	err = wc.Close()
	require.NoError(t, err)

	i, err := aferoFS.Stat("/tmp/foo")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), i.Mode())

	out := &bytes.Buffer{}

	f, err = openOutFile("-", 0755, 0644, false, out, false)
	require.NoError(t, err)
	assert.Equal(t, cfg.Stdout, f)
}

func TestGatherTemplates(t *testing.T) {
	ctx := context.Background()

	origfs := aferoFS
	defer func() { aferoFS = origfs }()
	aferoFS = afero.NewMemMapFs()
	afero.WriteFile(aferoFS, "foo", []byte("bar"), 0600)

	afero.WriteFile(aferoFS, "in/1", []byte("foo"), 0644)
	afero.WriteFile(aferoFS, "in/2", []byte("bar"), 0644)
	afero.WriteFile(aferoFS, "in/3", []byte("baz"), 0644)

	cfg := &config.Config{
		Stdin:  &bytes.Buffer{},
		Stdout: &bytes.Buffer{},
	}
	cfg.ApplyDefaults()
	templates, err := gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)

	cfg = &config.Config{
		Input:  "foo",
		Stdout: &bytes.Buffer{},
	}
	cfg.ApplyDefaults()
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].Text)
	assert.Equal(t, cfg.Stdout, templates[0].Writer)

	templates, err = gatherTemplates(ctx, &config.Config{
		Input:       "foo",
		OutputFiles: []string{"out"},
	}, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	// assert.Equal(t, iohelpers.NormalizeFileMode(0644), templates[0].mode)

	// out file is created only on demand
	_, err = aferoFS.Stat("out")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)

	info, err := aferoFS.Stat("out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), info.Mode())
	aferoFS.Remove("out")

	cfg = &config.Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		Stdout:      &bytes.Buffer{},
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].Text)
	assert.NotEqual(t, cfg.Stdout, templates[0].Writer)
	// assert.Equal(t, os.FileMode(0600), templates[0].mode)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)

	info, err = aferoFS.Stat("out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0600), info.Mode())
	aferoFS.Remove("out")

	cfg = &config.Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		OutMode:     "755",
		Stdout:      &bytes.Buffer{},
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].Text)
	assert.NotEqual(t, cfg.Stdout, templates[0].Writer)
	// assert.Equal(t, iohelpers.NormalizeFileMode(0755), templates[0].mode)

	_, err = templates[0].Writer.Write([]byte("hello world"))
	require.NoError(t, err)

	info, err = aferoFS.Stat("out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0755), info.Mode())
	aferoFS.Remove("out")

	templates, err = gatherTemplates(ctx, &config.Config{
		InputDir:  "in",
		OutputDir: "out",
	}, simpleNamer("out"))
	require.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].Text)
	aferoFS.Remove("out")
}

func TestCreateOutFile(t *testing.T) {
	origfs := aferoFS
	defer func() { aferoFS = origfs }()
	aferoFS = afero.NewMemMapFs()
	_ = aferoFS.Mkdir("in", 0755)

	_, err := createOutFile("in", 0755, 0644, false)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)
}

func TestParseNestedTemplates(t *testing.T) {
	ctx := context.Background()

	// in-memory test filesystem
	fsys := fstest.MapFS{
		"foo.t": {Data: []byte("hello world"), Mode: 0o600},
	}
	fsp := fsimpl.WrappedFSProvider(fsys, "file")
	ctx = ContextWithFSProvider(ctx, fsp)

	// simple test with single template
	u, _ := url.Parse("file:///foo.t")
	nested := config.Templates{"foo": {URL: u}}

	tmpl, _ := template.New("root").Parse(`{{ template "foo" }}`)

	err := parseNestedTemplates(ctx, nested, tmpl)
	require.NoError(t, err)

	out := bytes.Buffer{}
	err = tmpl.Execute(&out, nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", out.String())

	// test with directory of templates
	fsys["dir/foo.t"] = &fstest.MapFile{Data: []byte("foo"), Mode: 0o600}
	fsys["dir/bar.t"] = &fstest.MapFile{Data: []byte("bar"), Mode: 0o600}

	u, _ = url.Parse("file:///dir/")
	nested["dir"] = config.DataSource{URL: u}

	tmpl, _ = template.New("root").Parse(`{{ template "dir/foo.t" }} {{ template "dir/bar.t" }}`)

	err = parseNestedTemplates(ctx, nested, tmpl)
	require.NoError(t, err)

	out = bytes.Buffer{}
	err = tmpl.Execute(&out, nil)
	require.NoError(t, err)
	assert.Equal(t, "foo bar", out.String())
}
