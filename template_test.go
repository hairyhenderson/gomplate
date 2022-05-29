package gomplate

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenOutFile(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)

	cfg := &config.Config{Stdout: &bytes.Buffer{}}
	f, err := openOutFile("/tmp/foo", 0755, 0644, false, nil, false)
	assert.NoError(t, err)

	wc, ok := f.(io.WriteCloser)
	assert.True(t, ok)
	err = wc.Close()
	assert.NoError(t, err)

	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), i.Mode())

	out := &bytes.Buffer{}

	f, err = openOutFile("-", 0755, 0644, false, out, false)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Stdout, f)
}

func TestLoadContents(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()

	afero.WriteFile(fs, "foo", []byte("contents"), 0644)

	tmpl := &tplate{name: "foo"}
	b, err := tmpl.loadContents(nil)
	assert.NoError(t, err)
	assert.Equal(t, "contents", string(b))
}

func TestGatherTemplates(t *testing.T) {
	ctx := context.Background()

	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo", []byte("bar"), 0600)

	afero.WriteFile(fs, "in/1", []byte("foo"), 0644)
	afero.WriteFile(fs, "in/2", []byte("bar"), 0644)
	afero.WriteFile(fs, "in/3", []byte("baz"), 0644)

	cfg := &config.Config{
		Stdin:  &bytes.Buffer{},
		Stdout: &bytes.Buffer{},
	}
	cfg.ApplyDefaults()
	templates, err := gatherTemplates(ctx, cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)

	cfg = &config.Config{
		Input:  "foo",
		Stdout: &bytes.Buffer{},
	}
	cfg.ApplyDefaults()
	templates, err = gatherTemplates(ctx, cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].contents)
	assert.Equal(t, cfg.Stdout, templates[0].target)

	templates, err = gatherTemplates(ctx, &config.Config{
		Input:       "foo",
		OutputFiles: []string{"out"},
	}, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), templates[0].mode)

	// out file is created only on demand
	_, err = fs.Stat("out")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = templates[0].target.Write([]byte("hello world"))
	assert.NoError(t, err)

	info, err := fs.Stat("out")
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), info.Mode())
	fs.Remove("out")

	cfg = &config.Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		Stdout:      &bytes.Buffer{},
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].contents)
	assert.NotEqual(t, cfg.Stdout, templates[0].target)
	assert.Equal(t, os.FileMode(0600), templates[0].mode)

	_, err = templates[0].target.Write([]byte("hello world"))
	assert.NoError(t, err)

	info, err = fs.Stat("out")
	assert.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0600), info.Mode())
	fs.Remove("out")

	cfg = &config.Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		OutMode:     "755",
		Stdout:      &bytes.Buffer{},
	}
	templates, err = gatherTemplates(ctx, cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].contents)
	assert.NotEqual(t, cfg.Stdout, templates[0].target)
	assert.Equal(t, iohelpers.NormalizeFileMode(0755), templates[0].mode)

	_, err = templates[0].target.Write([]byte("hello world"))
	assert.NoError(t, err)

	info, err = fs.Stat("out")
	assert.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0755), info.Mode())
	fs.Remove("out")

	templates, err = gatherTemplates(ctx, &config.Config{
		InputDir:  "in",
		OutputDir: "out",
	}, simpleNamer("out"))
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].contents)
	fs.Remove("out")
}

func TestCreateOutFile(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("in", 0755)

	_, err := createOutFile("in", 0755, 0644, false)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)
}
