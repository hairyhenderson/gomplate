package integration

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

func setupDatasourcesGitTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithDir("repo",
			fs.WithFiles(map[string]string{
				"config.json": `{"foo": {"bar": "baz"}}`,
				"jsonfile":    `{"foo": {"bar": "baz"}}`,
			}),
			fs.WithDir("subdir",
				fs.WithFiles(map[string]string{
					"foo.txt":  "hello world",
					"bar.json": `{"qux": "quux"}`,
				}),
			),
		),
	)
	t.Cleanup(tmpDir.Remove)

	repoPath := tmpDir.Join("repo")

	icmd.RunCommand("git", "init", repoPath).
		Assert(t, icmd.Expected{Out: "Initialized empty Git repository"})
	icmd.RunCmd(icmd.Command("git", "add", "config.json"), icmd.Dir(repoPath)).Assert(t, icmd.Expected{})
	icmd.RunCmd(icmd.Command("git", "add", "jsonfile"), icmd.Dir(repoPath)).Assert(t, icmd.Expected{})
	icmd.RunCmd(icmd.Command("git", "add", "subdir"), icmd.Dir(repoPath)).Assert(t, icmd.Expected{})
	icmd.RunCmd(icmd.Command("git", "commit", "-m", "Initial commit"), icmd.Dir(repoPath)).Assert(t, icmd.Expected{})

	return tmpDir
}

func startGitDaemon(t *testing.T) string {
	tmpDir := setupDatasourcesGitTest(t)

	pidDir := fs.NewDir(t, "gomplate-inttests-pid")
	t.Cleanup(pidDir.Remove)

	port, addr := freeport(t)
	gitDaemon := icmd.Command("git", "daemon",
		"--verbose",
		"--port="+strconv.Itoa(port),
		"--base-path="+tmpDir.Path(),
		"--pid-file="+pidDir.Join("git.pid"),
		"--export-all",
		tmpDir.Join("repo", ".git"),
	)
	gitDaemon.Stdin = nil
	gitDaemon.Stdout = &bytes.Buffer{}
	gitDaemon.Dir = tmpDir.Path()
	result := icmd.StartCmd(gitDaemon)

	t.Cleanup(func() {
		err := result.Cmd.Process.Kill()
		require.NoError(t, err)

		_, err = result.Cmd.Process.Wait()
		require.NoError(t, err)

		result.Assert(t, icmd.Expected{ExitCode: 0})
	})

	// give git time to start
	time.Sleep(500 * time.Millisecond)

	return addr
}

func TestDatasources_GitFileDatasource(t *testing.T) {
	tmpDir := setupDatasourcesGitTest(t)

	u := filepath.ToSlash(tmpDir.Join("repo"))

	// on Windows the path will start with a volume, but we need a 'file:///'
	// prefix for the URL to be properly interpreted
	u = path.Join("/", u)
	o, e, err := cmd(t,
		"-d", "config=git+file://"+u+"//config.json",
		"-i", `{{ (datasource "config").foo.bar }}`,
	).run()
	assertSuccess(t, o, e, err, "baz")

	// subpath beginning with // is an antipattern, but should work for
	// backwards compatibility, params from subpath are used
	o, e, err = cmd(t,
		"-d", "repo=git+file://"+u,
		"-i", `{{ (datasource "repo" "//jsonfile?type=application/json" ).foo.bar }}`,
	).run()
	assertSuccess(t, o, e, err, "baz")

	// subpath beginning with // is an antipattern, but should work for
	// backwards compatibility
	o, e, err = cmd(t,
		"-d", "repo=git+file://"+u,
		"-i", `{{ (datasource "repo" "//config.json" ).foo.bar }}`,
	).run()
	assertSuccess(t, o, e, err, "baz")

	// subdir in datasource URL, relative subpath
	o, e, err = cmd(t,
		"-d", "repo=git+file://"+u+"//subdir/",
		"-i", `{{ include "repo" "foo.txt" }}`,
	).run()
	assertSuccess(t, o, e, err, "hello world")

	// ds URL ends with /, relative subpath beginning with .// is preferred
	o, e, err = cmd(t,
		"-d", "repo=git+file://"+u+"/",
		"-i", `{{ include "repo" ".//subdir/foo.txt" }}`,
	).run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestDatasources_GitDatasource(t *testing.T) {
	if isWindows {
		t.Skip("not going to run git daemon on Windows")
	}

	addr := startGitDaemon(t)

	o, e, err := cmd(t,
		"-c", "config=git://"+addr+"/repo//config.json",
		"-i", `{{ .config.foo.bar}}`,
	).run()
	assertSuccess(t, o, e, err, "baz")
}

func TestDatasources_GitHTTPDatasource(t *testing.T) {
	o, e, err := cmd(t,
		"-c", "short=git+https://github.com/git-fixtures/basic//json/short.json",
		"-i", `{{ .short.glossary.title}}`,
	).run()
	assertSuccess(t, o, e, err, "example glossary")

	// and one with a default branch of 'main'
	o, e, err = cmd(t,
		"-c", "data=git+https://github.com/hairyhenderson/git-fixtures.git//small_test.json",
		"-i", `{{ .data.foo}}`,
	).run()
	assertSuccess(t, o, e, err, "bar")
}

func TestDatasources_GitSSHDatasource(t *testing.T) {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		t.Skip("SSH Agent not running")
	}
	o, e, err := cmd(t,
		"-c", "short=git+ssh://git@github.com/git-fixtures/basic//json/short.json",
		"-i", `{{ .short.glossary.title}}`,
	).run()
	assertSuccess(t, o, e, err, "example glossary")
}
