//+build integration

package integration

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type GitDatasourcesSuite struct {
	tmpDir          *fs.Dir
	pidDir          *fs.Dir
	gitDaemonAddr   string
	gitDaemonResult *icmd.Result
}

var _ = Suite(&GitDatasourcesSuite{})

func (s *GitDatasourcesSuite) SetUpSuite(c *C) {
	s.pidDir = fs.NewDir(c, "gomplate-inttests-pid")
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"config.json": `{"foo": {"bar": "baz"}}`,
		}),
	)

	repoPath := s.tmpDir.Join("repo")

	result := icmd.RunCommand("git", "init", repoPath)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Initialized empty Git repository"})

	result = icmd.RunCommand("mv", s.tmpDir.Join("config.json"), repoPath)
	result.Assert(c, icmd.Expected{ExitCode: 0})

	result = icmd.RunCmd(icmd.Command("git", "add", "config.json"), icmd.Dir(repoPath))
	result.Assert(c, icmd.Expected{ExitCode: 0})

	result = icmd.RunCmd(icmd.Command("git", "commit", "-m", "Initial commit"), icmd.Dir(repoPath))
	result.Assert(c, icmd.Expected{ExitCode: 0})
}

func (s *GitDatasourcesSuite) startGitDaemon() {
	var port int
	port, s.gitDaemonAddr = freeport()
	gitDaemon := icmd.Command("git", "daemon",
		"--verbose",
		"--port="+strconv.Itoa(port),
		"--base-path="+s.tmpDir.Path(),
		"--pid-file="+s.pidDir.Join("git.pid"),
		"--export-all",
		s.tmpDir.Join("repo", ".git"),
	)
	gitDaemon.Dir = s.tmpDir.Path()
	s.gitDaemonResult = icmd.StartCmd(gitDaemon)
}

func (s *GitDatasourcesSuite) TearDownSuite(c *C) {
	defer s.tmpDir.Remove()
	defer s.pidDir.Remove()

	if s.gitDaemonResult != nil {
		err := killByPidFile(s.pidDir.Join("git.pid"))
		handle(c, err)

		s.gitDaemonResult.Cmd.Wait()

		s.gitDaemonResult.Assert(c, icmd.Expected{ExitCode: 0})
	}
}

func (s *GitDatasourcesSuite) TestGitFileDatasource(c *C) {
	u := filepath.ToSlash(s.tmpDir.Join("repo"))
	result := icmd.RunCommand(GomplateBin,
		"-d", "config=git+file://"+u+"//config.json",
		"-i", `{{ (datasource "config").foo.bar }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "repo=git+file://"+u,
		"-i", `{{ (datasource "repo" "//config.json?type=application/json" ).foo.bar }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "repo=git+file://"+u,
		"-i", `{{ (datasource "repo" "//config.json" ).foo.bar }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})
}

func (s *GitDatasourcesSuite) TestGitDatasource(c *C) {
	if isWindows {
		c.Skip("not going to run git daemon on Windows")
	}
	s.startGitDaemon()
	time.Sleep(500 * time.Millisecond)

	result := icmd.RunCommand(GomplateBin,
		"-c", "config=git://"+s.gitDaemonAddr+"/repo//config.json",
		"-i", `{{ .config.foo.bar}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})
}

func (s *GitDatasourcesSuite) TestGitHTTPDatasource(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-c", "short=git+https://github.com/git-fixtures/basic//json/short.json",
		"-i", `{{ .short.glossary.title}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "example glossary"})
}

func (s *GitDatasourcesSuite) TestGitSSHDatasource(c *C) {
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		c.Skip("SSH Agent not running")
	}
	result := icmd.RunCommand(GomplateBin,
		"-c", "short=git+ssh://git@github.com/git-fixtures/basic//json/short.json",
		"-i", `{{ .short.glossary.title}}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "example glossary"})
}
