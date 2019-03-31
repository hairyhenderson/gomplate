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

	_, err := openOutFile("/tmp/foo", 0644, false)
	assert.NoError(t, err)
	i, err := fs.Stat("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), i.Mode())

	defer func() { Stdout = os.Stdout }()
	Stdout = &nopWCloser{&bytes.Buffer{}}

	f, err := openOutFile("-", 0644, false)
	assert.NoError(t, err)
	assert.Equal(t, Stdout, f)
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

	tmpl := &tplate{name: "foo", targetPath: "/out/outfile"}
	err := tmpl.addTarget()
	assert.NoError(t, err)
	assert.NotNil(t, tmpl.target)
}

func TestGatherTemplates(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo", []byte("bar"), 0600)

	afero.WriteFile(fs, "in/1", []byte("foo"), 0644)
	afero.WriteFile(fs, "in/2", []byte("bar"), 0644)
	afero.WriteFile(fs, "in/3", []byte("baz"), 0644)

	templates, err := gatherTemplates(&Config{})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)

	templates, err = gatherTemplates(&Config{
		Input: "foo",
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "foo", templates[0].contents)
	assert.Equal(t, Stdout, templates[0].target)

	templates, err = gatherTemplates(&Config{
		Input:       "foo",
		OutputFiles: []string{"out"},
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "out", templates[0].targetPath)
	assert.Equal(t, os.FileMode(0644), templates[0].mode)
	info, err := fs.Stat("out")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode())
	fs.Remove("out")

	templates, err = gatherTemplates(&Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].contents)
	assert.NotEqual(t, Stdout, templates[0].target)
	assert.Equal(t, os.FileMode(0600), templates[0].mode)
	info, err = fs.Stat("out")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode())
	fs.Remove("out")

	templates, err = gatherTemplates(&Config{
		InputFiles:  []string{"foo"},
		OutputFiles: []string{"out"},
		OutMode:     "755",
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "bar", templates[0].contents)
	assert.NotEqual(t, Stdout, templates[0].target)
	assert.Equal(t, os.FileMode(0755), templates[0].mode)
	info, err = fs.Stat("out")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode())
	fs.Remove("out")

	templates, err = gatherTemplates(&Config{
		InputDir:  "in",
		OutputDir: "out",
	})
	assert.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.Equal(t, "foo", templates[0].contents)
	fs.Remove("out")
}

func TestProcessTemplates(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo", []byte("bar"), 0600)

	afero.WriteFile(fs, "in/1", []byte("foo"), 0644)
	afero.WriteFile(fs, "in/2", []byte("bar"), 0640)
	afero.WriteFile(fs, "in/3", []byte("baz"), 0644)

	afero.WriteFile(fs, "existing", []byte(""), 0644)

	testdata := []struct {
		templates []*tplate
		contents  []string
		modes     []os.FileMode
		targets   []io.WriteCloser
	}{
		{},
		{
			templates: []*tplate{{name: "<arg>", contents: "foo", targetPath: "-", mode: 0644}},
			contents:  []string{"foo"},
			modes:     []os.FileMode{0644},
			targets:   []io.WriteCloser{Stdout},
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
	for _, in := range testdata {
		actual, err := processTemplates(in.templates)
		assert.NoError(t, err)
		assert.Len(t, actual, len(in.templates))
		for i, a := range actual {
			assert.Equal(t, in.contents[i], a.contents)
			assert.Equal(t, in.templates[i].mode, a.mode)
			if len(in.targets) > 0 {
				assert.Equal(t, in.targets[i], a.target)
			}
			if in.templates[i].targetPath != "-" {
				info, err := fs.Stat(in.templates[i].targetPath)
				assert.NoError(t, err)
				assert.Equal(t, os.FileMode(in.modes[i]), info.Mode())
			}
		}
		fs.Remove("out")
	}
}

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
		w := &bufferCloser{&bytes.Buffer{}}
		opened := false
		f := newEmptySkipper(func() (io.WriteCloser, error) {
			t.Logf("I got called %#v", w)
			opened = true
			return w, nil
		})
		n, err := f.Write(d.in)
		assert.NoError(t, err)
		assert.Equal(t, len(d.in), n)
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

type bufferCloser struct {
	*bytes.Buffer
}

func (b *bufferCloser) Close() error {
	return nil
}
