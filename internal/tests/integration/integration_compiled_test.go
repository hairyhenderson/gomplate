//+build integration

package integration

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"gotest.tools/v3/icmd"
)

func gomplateBin() string {
	ext := ""
	if isWindows {
		ext = ".exe"
	}
	return filepath.Join(build.Default.GOPATH, "src", "github.com",
		"hairyhenderson", "gomplate", "bin", "gomplate"+ext)
}

func (c *command) run() (o, e string, err error) {
	cmd := icmd.Command(gomplateBin(), c.args...)
	cmd.Dir = c.dir
	cmd.Stdin = strings.NewReader(c.stdin)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOMPLATE_LOG_FORMAT=simple")
	for k, v := range c.env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	result := icmd.RunCmd(cmd)
	if result.Error != nil {
		result.Error = fmt.Errorf("%w: %s", result.Error, result.Stderr())
	}
	return result.Stdout(), result.Stderr(), result.Error
}
