package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v5/internal/datafs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	fsys := fstest.MapFS{}
	fsp := fsimpl.NewMux()
	fsp.Add(datafs.EnvFS)
	fsp.Add(datafs.StdinFS)
	fsp.Add(datafs.WrappedFSProvider(fsys, "mem", ""))
	ctx := datafs.ContextWithFSProvider(context.Background(), fsp)

	// no options - built-in function
	tr := NewRenderer(RenderOptions{})
	out := &bytes.Buffer{}
	err := tr.Render(ctx, "test", "{{ `hello world` | toUpper }}", out)
	require.NoError(t, err)
	assert.Equal(t, "HELLO WORLD", out.String())

	// with datasource and context
	hu, _ := url.Parse("stdin:")
	wu, _ := url.Parse("env:WORLD")

	t.Setenv("WORLD", "world")

	tr = NewRenderer(RenderOptions{
		Context: map[string]DataSource{
			"hi": {URL: hu},
		},
		Datasources: map[string]DataSource{
			"world": {URL: wu},
		},
	})
	ctx = datafs.ContextWithStdin(ctx, strings.NewReader("hello"))
	out = &bytes.Buffer{}
	err = tr.Render(ctx, "test", `{{ .hi | toUpper }} {{ (ds "world") | toUpper }}`, out)
	require.NoError(t, err)
	assert.Equal(t, "HELLO WORLD", out.String())

	// with a nested template
	nu, _ := url.Parse("nested.tmpl")
	fsys["nested.tmpl"] = &fstest.MapFile{Data: []byte(
		`<< . | toUpper >>`)}

	tr = NewRenderer(RenderOptions{
		Templates: map[string]DataSource{
			"nested": {URL: nu},
		},
		LDelim: "<<",
		RDelim: ">>",
	})
	out = &bytes.Buffer{}
	err = tr.Render(ctx, "test", `<< template "nested" "hello" >>`, out)
	require.NoError(t, err)
	assert.Equal(t, "HELLO", out.String())

	// errors contain the template name
	tr = NewRenderer(RenderOptions{})
	err = tr.Render(ctx, "foo", `{{ bogus }}`, &bytes.Buffer{})
	assert.ErrorContains(t, err, "template: foo:")
}

// countingFS counts how many times each file is opened
type countingFS struct {
	fs.FS
	opens map[string]int
}

func (c countingFS) Open(name string) (fs.File, error) {
	c.opens[name]++
	return c.FS.Open(name)
}

func TestRenderTemplates_NestedReadOnce(t *testing.T) {
	// relative paths are resolved from the working directory
	t.Chdir("/")

	fsys := countingFS{
		FS: fstest.MapFS{
			"nested.tmpl": {Data: []byte(`{{ . | toUpper }}`)},
		},
		opens: map[string]int{},
	}
	ctx := datafs.ContextWithFSProvider(context.Background(),
		datafs.WrappedFSProvider(fsys, "mem", ""))

	nu, _ := url.Parse("nested.tmpl")
	tr := NewRenderer(RenderOptions{
		Templates: map[string]DataSource{"nested": {URL: nu}},
	})

	render := func(names ...string) {
		t.Helper()

		templates := make([]Template, len(names))
		for i, name := range names {
			templates[i] = Template{
				Name:   name,
				Text:   `{{ template "nested" "` + name + `" }}`,
				Writer: &bytes.Buffer{},
			}
		}

		require.NoError(t, tr.RenderTemplates(ctx, templates))

		for i, name := range names {
			assert.Equal(t, strings.ToUpper(name), templates[i].Writer.(*bytes.Buffer).String())
		}
	}

	render("one")
	opens := fsys.opens["nested.tmpl"]
	require.NotZero(t, opens)

	// more templates, and more calls, don't read the nested template again
	render("two", "three", "four")
	render("five")
	assert.Equal(t, opens, fsys.opens["nested.tmpl"])
}

func TestReadNestedTemplates_Canceled(t *testing.T) {
	t.Chdir("/")

	fsys := fstest.MapFS{
		"nested.tmpl": {Data: []byte(`hello`)},
	}
	ctx, cancel := context.WithCancel(datafs.ContextWithFSProvider(
		context.Background(), datafs.WrappedFSProvider(fsys, "mem", "")))
	cancel()

	nu, _ := url.Parse("nested.tmpl")
	r := &renderer{nested: map[string]DataSource{"nested": {URL: nu}}}

	_, err := r.readNestedTemplates(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, r.nestedTmpls)
}

//// examples

func ExampleRenderer() {
	ctx := context.Background()

	// create a new template renderer
	tr := NewRenderer(RenderOptions{})

	// render a template to stdout
	err := tr.Render(ctx, "mytemplate",
		`{{ "hello, world!" | toUpper }}`,
		os.Stdout)
	if err != nil {
		fmt.Println("gomplate error:", err)
	}

	// Output:
	// HELLO, WORLD!
}

func ExampleRenderer_manyTemplates() {
	ctx := context.Background()

	// create a new template renderer
	tr := NewRenderer(RenderOptions{})

	templates := []Template{
		{
			Name:   "one.tmpl",
			Text:   `contents of {{ tmpl.Path }}`,
			Writer: &bytes.Buffer{},
		},
		{
			Name:   "two.tmpl",
			Text:   `{{ "hello world" | toUpper }}`,
			Writer: &bytes.Buffer{},
		},
		{
			Name:   "three.tmpl",
			Text:   `1 + 1 = {{ math.Add 1 1 }}`,
			Writer: &bytes.Buffer{},
		},
	}

	// render the templates
	err := tr.RenderTemplates(ctx, templates)
	if err != nil {
		panic(err)
	}

	for _, t := range templates {
		fmt.Printf("%s: %s\n", t.Name, t.Writer.(*bytes.Buffer).String())
	}

	// Output:
	// one.tmpl: contents of one.tmpl
	// two.tmpl: HELLO WORLD
	// three.tmpl: 1 + 1 = 2
}

func fakeJSONServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Luna","breed":"Siberian Husky","age":4}`))
	}))
}

func ExampleRenderer_datasources() {
	ctx := context.Background()

	srv := fakeJSONServer()
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	tr := NewRenderer(RenderOptions{
		Context: map[string]DataSource{
			"pet": {URL: u},
		},
	})

	err := tr.Render(ctx, "jsontest",
		`{{ .pet.name }} is a {{ .pet.age }} year old {{ .pet.breed }}`,
		os.Stdout)
	if err != nil {
		panic(err)
	}

	// Output:
	// Luna is a 4 year old Siberian Husky
}

func BenchmarkRenderTemplates_nested(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{{ . | toUpper }}`))
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/nested.tmpl")
	opts := RenderOptions{
		Templates: map[string]DataSource{"nested": {URL: u}},
	}

	ctx := context.Background()
	templates := make([]Template, 50)

	for b.Loop() {
		// a new renderer each time, as for a single gomplate run
		tr := NewRenderer(opts)

		for i := range templates {
			templates[i] = Template{
				Name:   fmt.Sprintf("t%d", i),
				Text:   `{{ template "nested" "hello" }}`,
				Writer: &bytes.Buffer{},
			}
		}

		err := tr.RenderTemplates(ctx, templates)
		if err != nil {
			b.Fatal(err)
		}
	}
}
