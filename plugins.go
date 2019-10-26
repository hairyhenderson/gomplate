package gomplate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/env"
)

func bindPlugins(plugins []string, funcMap template.FuncMap) error {
	for _, p := range plugins {
		plugin, err := newPlugin(p)
		if err != nil {
			return err
		}
		if _, ok := funcMap[plugin.name]; ok {
			return fmt.Errorf("function %q is already bound, and can not be overridden", plugin.name)
		}
		funcMap[plugin.name] = plugin.run
	}
	return nil
}

// plugin represents a custom function that binds to an external process to be executed
type plugin struct {
	name, path string
}

func newPlugin(value string) (*plugin, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) < 2 {
		return nil, errors.New("plugin requires both name and path")
	}

	p := &plugin{
		name: parts[0],
		path: parts[1],
	}
	return p, nil
}

// builds a command that's appropriate for running scripts
// nolint: gosec
func (p *plugin) buildCommand(a []string) (name string, args []string) {
	switch filepath.Ext(p.path) {
	case ".ps1":
		a = append([]string{"-File", p.path}, a...)
		return findPowershell(), a
	case ".cmd", ".bat":
		a = append([]string{"/c", p.path}, a...)
		return "cmd.exe", a
	default:
		return p.path, a
	}
}

// finds the appropriate powershell command for the platform - prefers
// PowerShell Core (`pwsh`), but on Windows if it's not found falls back to
// Windows PowerShell (`powershell`).
func findPowershell() string {
	if runtime.GOOS != "windows" {
		return "pwsh"
	}

	_, err := exec.LookPath("pwsh")
	if err != nil {
		return "powershell"
	}
	return "pwsh"
}

func (p *plugin) run(args ...interface{}) (interface{}, error) {
	a := conv.ToStrings(args...)

	name, a := p.buildCommand(a)

	t, err := time.ParseDuration(env.Getenv("GOMPLATE_PLUGIN_TIMEOUT", "5s"))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	c := exec.CommandContext(ctx, name, a...)
	c.Stdin = nil
	c.Stderr = os.Stderr
	outBuf := &bytes.Buffer{}
	c.Stdout = outBuf

	// make sure all signals are propagated
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		// Pass signals to the sub-process
		sig := <-sigs
		if c.Process != nil {
			// nolint: gosec
			_ = c.Process.Signal(sig)
		}
	}()
	start := time.Now()
	err = c.Run()
	elapsed := time.Since(start)

	if ctx.Err() != nil {
		err = fmt.Errorf("plugin timed out after %v: %w", elapsed, ctx.Err())
	}

	return outBuf.String(), err
}
