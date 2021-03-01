package integration

import (
	"os"
	"path/filepath"
	"sort"

	. "gopkg.in/check.v1"

	"github.com/spf13/afero"
	tassert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/fs"
)

type GomplateignoreSuite struct {
	inBuilder func(inFileOps ...fs.PathOp)
	tmpDir    *fs.Dir
}

var _ = Suite(&GomplateignoreSuite{})

func (s *GomplateignoreSuite) SetUpTest(c *C) {
	const basedir = "gomplate-gomplateignore-tests"
	s.inBuilder = func(inFileOps ...fs.PathOp) {
		s.tmpDir = fs.NewDir(c, basedir,
			fs.WithDir("in", inFileOps...),
			fs.WithDir("out"),
		)
	}
}

func (s *GomplateignoreSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *GomplateignoreSuite) execute(c *C, ignoreContent string, inFileOps ...fs.PathOp) {
	s.executeOpts(c, ignoreContent, []string{}, inFileOps...)
}

func (s *GomplateignoreSuite) executeOpts(c *C, ignoreContent string, opts []string, inFileOps ...fs.PathOp) {
	inFileOps = append(inFileOps, fs.WithFile(".gomplateignore", ignoreContent))
	s.inBuilder(inFileOps...)

	argv := []string{}
	argv = append(argv, opts...)
	argv = append(argv,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-dir", s.tmpDir.Join("out"),
	)
	o, e, err := cmdTest(c, argv...)
	assertSuccess(c, o, e, err, "")
}

func (s *GomplateignoreSuite) collectOutFiles() (files []string, err error) {
	files = []string{}

	fs := afero.NewBasePathFs(afero.NewOsFs(), s.tmpDir.Join("out"))
	afero.Walk(fs, "", func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			err = werr
			return nil
		}
		if path != "" && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	sort.Strings(files)
	return
}

func (s *GomplateignoreSuite) TestGomplateignore_Simple(c *C) {
	s.execute(c, `# all dot files
.*
*.log`,
		fs.WithFile("empty.log", ""),
		fs.WithFile("rain.txt", ""))

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, []string{"rain.txt"}, files)
}

func fromSlashes(paths ...string) []string {
	for i, v := range paths {
		paths[i] = filepath.FromSlash(v)
	}
	return paths
}

func (s *GomplateignoreSuite) TestGomplateignore_Folder(c *C) {
	s.execute(c, `.gomplateignore
f[o]o/bar
!foo/bar/tool`,
		fs.WithDir("foo",
			fs.WithDir("bar",
				fs.WithDir("tool",
					fs.WithFile("lex.txt", ""),
				),
				fs.WithFile("1.txt", ""),
			),
			fs.WithDir("tar",
				fs.WithFile("2.txt", ""),
			),
		),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"foo/bar/tool/lex.txt", "foo/tar/2.txt"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_Root(c *C) {
	s.execute(c, `.gomplateignore
/1.txt`,
		fs.WithDir("sub",
			fs.WithFile("1.txt", ""),
			fs.WithFile("2.txt", ""),
		),
		fs.WithFile("1.txt", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"sub/1.txt", "sub/2.txt"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_Exclusion(c *C) {
	s.execute(c, `.gomplateignore
/e*.txt
!/e2.txt
en/e3.txt
!`,
		fs.WithFile("!", ""),
		fs.WithFile("e1.txt", ""),
		fs.WithFile("e2.txt", ""),
		fs.WithFile("e3.txt", ""),
		fs.WithDir("en",
			fs.WithFile("e1.txt", ""),
			fs.WithFile("e2.txt", ""),
			fs.WithFile("e3.txt", ""),
		),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"!", "e2.txt", "en/e1.txt", "en/e2.txt"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_Nested(c *C) {
	s.execute(c, `inner/foo.md`,
		fs.WithDir("inner",
			fs.WithDir("inner2",
				fs.WithFile(".gomplateignore", "moss.ini\n!/jess.ini"),
				fs.WithFile("jess.ini", ""),
				fs.WithFile("moss.ini", "")),
			fs.WithFile(".gomplateignore", "*.lst\njess.ini"),
			fs.WithFile("2.lst", ""),
			fs.WithFile("foo.md", ""),
		),
		fs.WithFile("1.txt", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(".gomplateignore", "1.txt",
		"inner/.gomplateignore",
		"inner/inner2/.gomplateignore",
		"inner/inner2/jess.ini"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_ByName(c *C) {
	s.execute(c, `.gomplateignore
world.txt`,
		fs.WithDir("aa",
			fs.WithDir("a1",
				fs.WithDir("a2",
					fs.WithFile("hello.txt", ""),
					fs.WithFile("world.txt", "")),
				fs.WithFile("hello.txt", ""),
				fs.WithFile("world.txt", "")),
			fs.WithFile("hello.txt", ""),
			fs.WithFile("world.txt", "")),
		fs.WithDir("bb",
			fs.WithFile("hello.txt", ""),
			fs.WithFile("world.txt", "")),
		fs.WithFile("hello.txt", ""),
		fs.WithFile("world.txt", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"aa/a1/a2/hello.txt", "aa/a1/hello.txt",
		"aa/hello.txt", "bb/hello.txt", "hello.txt"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_BothName(c *C) {
	s.execute(c, `.gomplateignore
loss.txt
!2.log
`,
		fs.WithDir("loss.txt",
			fs.WithFile("1.log", ""),
			fs.WithFile("2.log", "")),
		fs.WithDir("foo",
			fs.WithFile("loss.txt", ""),
			fs.WithFile("bare.txt", "")),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"foo/bare.txt", "loss.txt/2.log"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_LeadingSpace(c *C) {
	s.execute(c, `.gomplateignore
  what.txt
!inner/  what.txt
*.log
!  dart.log
`,
		fs.WithDir("inner",
			fs.WithFile("  what.txt", ""),
			fs.WithFile("  dart.log", "")),
		fs.WithDir("inner2",
			fs.WithFile("  what.txt", "")),
		fs.WithFile("  what.txt", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"inner/  dart.log", "inner/  what.txt"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_WithExcludes(c *C) {
	s.executeOpts(c, `.gomplateignore
*.log
`, []string{
		"--exclude", "crash.bin",
		"--exclude", "rules/*.txt",
		"--exclude", "sprites/*.ini",
	},
		fs.WithDir("logs",
			fs.WithFile("archive.zip", ""),
			fs.WithFile("engine.log", ""),
			fs.WithFile("skills.log", "")),
		fs.WithDir("rules",
			fs.WithFile("index.csv", ""),
			fs.WithFile("fire.txt", ""),
			fs.WithFile("earth.txt", "")),
		fs.WithDir("sprites",
			fs.WithFile("human.csv", ""),
			fs.WithFile("demon.xml", ""),
			fs.WithFile("alien.ini", "")),
		fs.WithFile("manifest.json", ""),
		fs.WithFile("crash.bin", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes(
		"logs/archive.zip", "manifest.json", "rules/index.csv",
		"sprites/demon.xml", "sprites/human.csv"), files)
}

func (s *GomplateignoreSuite) TestGomplateignore_WithIncludes(c *C) {
	s.executeOpts(c, `.gomplateignore
*.log
`, []string{
		"--include", "rules/*",
		"--exclude", "rules/*.txt",
	},
		fs.WithDir("logs",
			fs.WithFile("archive.zip", ""),
			fs.WithFile("engine.log", ""),
			fs.WithFile("skills.log", "")),
		fs.WithDir("rules",
			fs.WithFile("index.csv", ""),
			fs.WithFile("fire.txt", ""),
			fs.WithFile("earth.txt", "")),
		fs.WithFile("manifest.json", ""),
		fs.WithFile("crash.bin", ""),
	)

	files, err := s.collectOutFiles()
	tassert.NoError(c, err)
	tassert.Equal(c, fromSlashes("rules/index.csv"), files)
}
