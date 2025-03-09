package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindPlugins(t *testing.T) {
	ctx := context.Background()
	fm := template.FuncMap{}
	cfg := &Config{
		Plugins: map[string]PluginConfig{},
	}
	err := bindPlugins(ctx, cfg, fm)
	require.NoError(t, err)
	assert.EqualValues(t, template.FuncMap{}, fm)

	cfg.Plugins = map[string]PluginConfig{"foo": {Cmd: "bar"}}
	err = bindPlugins(ctx, cfg, fm)
	require.NoError(t, err)
	assert.Contains(t, fm, "foo")

	err = bindPlugins(ctx, cfg, fm)
	assert.ErrorContains(t, err, "already bound")
}

func TestBuildCommand(t *testing.T) {
	ctx := context.Background()
	data := []struct {
		name, path string
		args       []string
		expected   []string
	}{
		{"foo", "foo", nil, []string{"foo"}},
		{"foo", "foo", []string{"bar"}, []string{"foo", "bar"}},
		{"foo", "foo.bat", nil, []string{"cmd.exe", "/c", "foo.bat"}},
		{"foo", "foo.cmd", []string{"bar"}, []string{"cmd.exe", "/c", "foo.cmd", "bar"}},
		{"foo", "foo.ps1", []string{"bar", "baz"}, []string{"pwsh", "-File", "foo.ps1", "bar", "baz"}},
	}
	for _, d := range data {
		p := &plugin{
			ctx:  ctx,
			path: d.path,
		}
		name, args := p.buildCommand(d.args)
		actual := append([]string{name}, args...)
		assert.EqualValues(t, d.expected, actual)
	}
}

func TestRun(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	p := &plugin{
		ctx:     ctx,
		timeout: 500 * time.Millisecond,
		stderr:  stderr,
		path:    "echo",
	}
	out, err := p.run("foo")
	require.NoError(t, err)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "foo", strings.TrimSpace(out.(string)))

	p = &plugin{
		ctx:     ctx,
		timeout: 500 * time.Millisecond,
		stderr:  stderr,
		path:    "echo",
		args:    []string{"foo", "bar"},
	}
	out, err = p.run()
	require.NoError(t, err)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "foo bar", strings.TrimSpace(out.(string)))

	p = &plugin{
		ctx:     ctx,
		timeout: 500 * time.Millisecond,
		stderr:  stderr,
		path:    "echo",
		args:    []string{"foo", "bar"},
	}
	out, err = p.run("baz", "qux")
	require.NoError(t, err)
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "foo bar baz qux", strings.TrimSpace(out.(string)))
}

func ExamplePluginFunc() {
	ctx := context.Background()

	// PluginFunc creates a template function that runs an arbitrary command.
	f := PluginFunc(ctx, "echo", PluginOpts{})

	// The function can be used in a template, but here we'll just run it
	// directly. This is equivalent to running 'echo foo bar'
	out, err := f("foo", "bar")
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// foo bar
}

func ExamplePluginFunc_with_template() {
	ctx := context.Background()

	f := PluginFunc(ctx, "echo", PluginOpts{})

	// PluginFunc is intended for use with gomplate, but can be used in any
	// text/template by adding it to the FuncMap.
	tmpl := template.New("new").Funcs(template.FuncMap{"echo": f})

	tmpl, err := tmpl.Parse(`{{ echo "baz" "qux" }}`)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, nil)
	if err != nil {
		panic(err)
	}

	// Output:
	// baz qux
}
