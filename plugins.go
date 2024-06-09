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

	"github.com/hairyhenderson/gomplate/v4/conv"
)

// bindPlugins creates custom plugin functions for each plugin specified by
// the config, and adds them to the given funcMap. Uses the configuration's
// PluginTimeout as the default plugin Timeout. Errors if a function name is
// duplicated.
func bindPlugins(ctx context.Context, cfg *Config, funcMap template.FuncMap) error {
	for k, v := range cfg.Plugins {
		if _, ok := funcMap[k]; ok {
			return fmt.Errorf("function %q is already bound, and can not be overridden", k)
		}

		// default the timeout to the one in the config
		timeout := cfg.PluginTimeout
		if v.Timeout != 0 {
			timeout = v.Timeout
		}

		funcMap[k] = PluginFunc(ctx, v.Cmd, PluginOpts{
			Timeout: timeout,
			Pipe:    v.Pipe,
			Stderr:  cfg.Stderr,
			Args:    v.Args,
		})
	}

	return nil
}

// PluginOpts are options for controlling plugin function execution
type PluginOpts struct {
	// Stderr can be set to redirect the plugin's stderr to a custom writer.
	// Defaults to os.Stderr.
	Stderr io.Writer

	// Args are additional arguments to pass to the plugin. These precede any
	// arguments passed to the plugin function at runtime.
	Args []string

	// Timeout is the maximum amount of time to wait for the plugin to complete.
	// Defaults to 5 seconds.
	Timeout time.Duration

	// Pipe indicates whether the last argument should be piped to the plugin's
	// stdin (true) or processed as a commandline argument (false)
	Pipe bool
}

// PluginFunc creates a template function that runs an external process - either
// a shell script or commandline executable.
func PluginFunc(ctx context.Context, cmd string, opts PluginOpts) func(...interface{}) (interface{}, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	plugin := &plugin{
		ctx:     ctx,
		path:    cmd,
		args:    opts.Args,
		timeout: timeout,
		pipe:    opts.Pipe,
		stderr:  stderr,
	}

	return plugin.run
}

// plugin represents a custom function that binds to an external process to be executed
type plugin struct {
	ctx     context.Context
	stderr  io.Writer
	path    string
	args    []string
	timeout time.Duration
	pipe    bool
}

// builds a command that's appropriate for running scripts
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
	a = append(p.args, a...)

	name, a := p.buildCommand(a)

	ctx, cancel := context.WithTimeout(p.ctx, p.timeout)
	defer cancel()

	var stdin *bytes.Buffer
	if p.pipe && len(a) > 0 {
		stdin = bytes.NewBufferString(a[len(a)-1])
		a = a[:len(a)-1]
	}

	c := exec.CommandContext(ctx, name, a...)
	if stdin != nil {
		c.Stdin = stdin
	}

	c.Stderr = p.stderr
	outBuf := &bytes.Buffer{}
	c.Stdout = outBuf

	start := time.Now()
	err := c.Start()
	if err != nil {
		return nil, fmt.Errorf("starting command: %w", err)
	}

	// make sure all signals are propagated
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		select {
		case sig := <-sigs:
			// Pass signals to the sub-process
			if c.Process != nil {
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
