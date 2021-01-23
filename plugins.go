package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

func bindPlugins(ctx context.Context, cfg *config.Config, funcMap template.FuncMap) error {
	for k, v := range cfg.Plugins {
		plugin := &plugin{
			ctx:     ctx,
			name:    k,
			path:    v,
			timeout: cfg.PluginTimeout,
			stderr:  cfg.Stderr,
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
	timeout    time.Duration
	ctx        context.Context
	stderr     io.Writer
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

	ctx, cancel := context.WithTimeout(p.ctx, p.timeout)
	defer cancel()
	c := exec.CommandContext(ctx, name, a...)
	c.Stdin = nil
	c.Stderr = p.stderr
	outBuf := &bytes.Buffer{}
	c.Stdout = outBuf

	start := time.Now()
	err := c.Start()
	if err != nil {
		return nil, err
	}

	// make sure all signals are propagated
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		select {
		case sig := <-sigs:
			// Pass signals to the sub-process
			if c.Process != nil {
				// nolint: gosec
				_ = c.Process.Signal(sig)
			}
		case <-ctx.Done():
		}
	}()

	err = c.Wait()
	elapsed := time.Since(start)

	if ctx.Err() != nil {
		err = fmt.Errorf("plugin timed out after %v: %w", elapsed, ctx.Err())
	}

	return outBuf.String(), err
}
