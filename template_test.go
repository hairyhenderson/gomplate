package gomplate

import (
	"bytes"
	"fmt"
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

	cfg := &config.Config{}
	f, err := openOutFile(cfg, "/tmp/foo", 0755, 0644, false)
	assert.NoError(t, err)

	wc, ok := f.(io.WriteCloser)
	assert.True(t, ok)
	err = wc.Close()
	assert.NoError(t, err)

	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0644), i.Mode())

	cfg.Stdout = &bytes.Buffer{}

	f, err = openOutFile(cfg, "-", 0755, 0644, false)
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
	templates, err := gatherTemplates(cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)

	cfg = &config.Config{
		Input:  "foo",
		Stdout: &bytes.Buffer{},
	}
	cfg.ApplyDefaults()
	templates, err = gatherTemplates(cfg, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].contents)
	assert.Equal(t, cfg.Stdout, templates[0].target)

	templates, err = gatherTemplates(&config.Config{
		Input:       "foo",
		OutputFiles: []string{"out"},
	}, nil)
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "out", templates[0].targetPath)
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
	templates, err = gatherTemplates(cfg, nil)
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
	templates, err = gatherTemplates(cfg, nil)
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

	templates, err = gatherTemplates(&config.Config{
		InputDir:  "in",
		OutputDir: "out",
	}, simpleNamer("out"))
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].contents)
	fs.Remove("out")
}

func TestProcessTemplates(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo", []byte("bar"), iohelpers.NormalizeFileMode(0600))

	afero.WriteFile(fs, "in/1", []byte("foo"), iohelpers.NormalizeFileMode(0644))
	afero.WriteFile(fs, "in/2", []byte("bar"), iohelpers.NormalizeFileMode(0640))
	afero.WriteFile(fs, "in/3", []byte("baz"), iohelpers.NormalizeFileMode(0644))

	afero.WriteFile(fs, "existing", []byte(""), iohelpers.NormalizeFileMode(0644))

	cfg := &config.Config{
		Stdout: &bytes.Buffer{},
	}
	testdata := []struct {
		templates []*tplate
		contents  []string
		modes     []os.FileMode
		targets   []io.Writer
	}{
		{},
		{
			templates: []*tplate{{name: "<arg>", contents: "foo", targetPath: "-", mode: 0644}},
			contents:  []string{"foo"},
			modes:     []os.FileMode{0644},
			targets:   []io.Writer{cfg.Stdout},
		},
		{
			templates: []*tplate{{name: "<arg>", contents: "foo", targetPath: "out", mode: 0644}},
			contents:  []string{"foo"},
			modes:     []os.FileMode{0644},
		},
		{
			templates: []*tplate{{name: "foo", targetPath: "out", mode: 0600}},
			contents:  []string{"bar"},
			modes:     []os.FileMode{0600},
		},
		{
			templates: []*tplate{{name: "foo", targetPath: "out", mode: 0755}},
			contents:  []string{"bar"},
			modes:     []os.FileMode{0755},
		},
		{
			templates: []*tplate{
				{name: "in/1", targetPath: "out/1", mode: 0644},
				{name: "in/2", targetPath: "out/2", mode: 0640},
				{name: "in/3", targetPath: "out/3", mode: 0644},
			},
			contents: []string{"foo", "bar", "baz"},
			modes:    []os.FileMode{0644, 0640, 0644},
		},
		{
			templates: []*tplate{
				{name: "foo", targetPath: "existing", mode: 0755},
			},
			contents: []string{"bar"},
			modes:    []os.FileMode{0644},
		},
		{
			templates: []*tplate{
				{name: "foo", targetPath: "existing", mode: 0755, modeOverride: true},
			},
			contents: []string{"bar"},
			modes:    []os.FileMode{0755},
		},
	}
	for i, in := range testdata {
		in := in
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual, err := processTemplates(cfg, in.templates)
			assert.NoError(t, err)
			assert.Len(t, actual, len(in.templates))
			for i, a := range actual {
				current := in.templates[i]
				assert.Equal(t, in.contents[i], a.contents)
				assert.Equal(t, current.mode, a.mode)
				if len(in.targets) > 0 {
					assert.Equal(t, in.targets[i], a.target)
				}
				if current.targetPath != "-" && current.name != "<arg>" {
					_, err = current.loadContents(nil)
					assert.NoError(t, err)

					n, err := current.target.Write([]byte("hello world"))
					assert.NoError(t, err)
					assert.Equal(t, 11, n)

					info, err := fs.Stat(current.targetPath)
					assert.NoError(t, err)
					assert.Equal(t, iohelpers.NormalizeFileMode(in.modes[i]), info.Mode())
				}
			}
			fs.Remove("out")
		})
	}
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
