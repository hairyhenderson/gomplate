package integration

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tfs "gotest.tools/v3/fs"
)

func setupGomplateignoreTest(t *testing.T) func(inFileOps ...tfs.PathOp) *tfs.Dir {
	basedir := "gomplate-gomplateignore-tests"

	inBuilder := func(inFileOps ...tfs.PathOp) *tfs.Dir {
		tmpDir := tfs.NewDir(t, basedir,
			tfs.WithDir("in", inFileOps...),
			tfs.WithDir("out"),
		)
		t.Cleanup(tmpDir.Remove)
		return tmpDir
	}

	return inBuilder
}

func execute(t *testing.T, ignoreContent string, inFileOps ...tfs.PathOp) ([]string, error) {
	return executeOpts(t, ignoreContent, []string{}, inFileOps...)
}

func executeOpts(t *testing.T, ignoreContent string, opts []string, inFileOps ...tfs.PathOp) ([]string, error) {
	inBuilder := setupGomplateignoreTest(t)

	inFileOps = append(inFileOps, tfs.WithFile(".gomplateignore", ignoreContent))
	tmpDir := inBuilder(inFileOps...)

	argv := make([]string, 0, len(opts)+4)
	argv = append(argv, opts...)
	argv = append(argv,
		"--input-dir", tmpDir.Join("in"),
		"--output-dir", tmpDir.Join("out"),
	)
	o, e, err := cmd(t, argv...).run()
	assertSuccess(t, o, e, err, "")

	files := []string{}

	fsys := os.DirFS(tmpDir.Join("out") + "/")
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != "" && !d.IsDir() {
			path = filepath.FromSlash(path)
			files = append(files, path)
		}
		return nil
	})

	sort.Strings(files)

	return files, err
}

func TestGomplateignore_Simple(t *testing.T) {
	files, err := execute(t, `# all dot files
.*
*.log`,
		tfs.WithFile("foo.log", "..."),
		tfs.WithFile("rain.txt", "..."))

	require.NoError(t, err)
	assert.Equal(t, []string{"rain.txt"}, files)
}

func fromSlashes(paths ...string) []string {
	for i, v := range paths {
		paths[i] = filepath.FromSlash(v)
	}
	return paths
}

func TestGomplateignore_Folder(t *testing.T) {
	files, err := execute(t, `.gomplateignore
f[o]o/bar
!foo/bar/tool`,
		tfs.WithDir("foo",
			tfs.WithDir("bar",
				tfs.WithDir("tool",
					tfs.WithFile("lex.txt", "..."),
				),
				tfs.WithFile("1.txt", "..."),
			),
			tfs.WithDir("tar",
				tfs.WithFile("2.txt", "..."),
			),
		),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"foo/bar/tool/lex.txt", "foo/tar/2.txt"), files)
}

func TestGomplateignore_Root(t *testing.T) {
	files, err := execute(t, `.gomplateignore
/1.txt`,
		tfs.WithDir("sub",
			tfs.WithFile("1.txt", "..."),
			tfs.WithFile("2.txt", "..."),
		),
		tfs.WithFile("1.txt", "..."),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"sub/1.txt", "sub/2.txt"), files)
}

func TestGomplateignore_Exclusion(t *testing.T) {
	files, err := execute(t, `.gomplateignore
/e*.txt
!/e2.txt
en/e3.txt
!`,
		tfs.WithFile("!", "xxx"),
		tfs.WithFile("e1.txt", "xxx"),
		tfs.WithFile("e2.txt", "xxx"),
		tfs.WithFile("e3.txt", "xxx"),
		tfs.WithDir("en",
			tfs.WithFile("e1.txt", "xxx"),
			tfs.WithFile("e2.txt", "xxx"),
			tfs.WithFile("e3.txt", "xxx"),
		),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"!", "e2.txt", "en/e1.txt", "en/e2.txt"), files)
}

func TestGomplateignore_Nested(t *testing.T) {
	files, err := execute(t, `inner/foo.md`,
		tfs.WithDir("inner",
			tfs.WithDir("inner2",
				tfs.WithFile(".gomplateignore", "moss.ini\n!/jess.ini"),
				tfs.WithFile("jess.ini", "xxx"),
				tfs.WithFile("moss.ini", "xxx")),
			tfs.WithFile(".gomplateignore", "*.lst\njess.ini"),
			tfs.WithFile("2.lst", "xxx"),
			tfs.WithFile("foo.md", "xxx"),
		),
		tfs.WithFile("1.txt", "xxx"),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(".gomplateignore", "1.txt",
		"inner/.gomplateignore",
		"inner/inner2/.gomplateignore",
		"inner/inner2/jess.ini"), files)
}

func TestGomplateignore_ByName(t *testing.T) {
	files, err := execute(t, `.gomplateignore
world.txt`,
		tfs.WithDir("aa",
			tfs.WithDir("a1",
				tfs.WithDir("a2",
					tfs.WithFile("hello.txt", "..."),
					tfs.WithFile("world.txt", "...")),
				tfs.WithFile("hello.txt", "..."),
				tfs.WithFile("world.txt", "...")),
			tfs.WithFile("hello.txt", "..."),
			tfs.WithFile("world.txt", "...")),
		tfs.WithDir("bb",
			tfs.WithFile("hello.txt", "..."),
			tfs.WithFile("world.txt", "...")),
		tfs.WithFile("hello.txt", "..."),
		tfs.WithFile("world.txt", "..."),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"aa/a1/a2/hello.txt", "aa/a1/hello.txt",
		"aa/hello.txt", "bb/hello.txt", "hello.txt"), files)
}

func TestGomplateignore_BothName(t *testing.T) {
	files, err := execute(t, `.gomplateignore
loss.txt
!2.log
`,
		tfs.WithDir("loss.txt",
			tfs.WithFile("1.log", "xxx"),
			tfs.WithFile("2.log", "xxx")),
		tfs.WithDir("foo",
			tfs.WithFile("loss.txt", "xxx"),
			tfs.WithFile("bare.txt", "xxx")),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"foo/bare.txt", "loss.txt/2.log"), files)
}

func TestGomplateignore_LeadingSpace(t *testing.T) {
	files, err := execute(t, `.gomplateignore
  what.txt
!inner/  what.txt
*.log
!  dart.log
`,
		tfs.WithDir("inner",
			tfs.WithFile("  what.txt", "xxx"),
			tfs.WithFile("  dart.log", "xxx")),
		tfs.WithDir("inner2",
			tfs.WithFile("  what.txt", "xxx")),
		tfs.WithFile("  what.txt", "xxx"),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"inner/  dart.log", "inner/  what.txt"), files)
}

func TestGomplateignore_WithExcludes(t *testing.T) {
	files, err := executeOpts(t, `.gomplateignore
*.log
`, []string{
		"--exclude", "crash.bin",
		"--exclude", "rules/*.txt",
		"--exclude", "sprites/*.ini",
	},
		tfs.WithDir("logs",
			tfs.WithFile("archive.zip", "x"),
			tfs.WithFile("engine.log", "x"),
			tfs.WithFile("skills.log", "x")),
		tfs.WithDir("rules",
			tfs.WithFile("index.csv", "x"),
			tfs.WithFile("fire.txt", "x"),
			tfs.WithFile("earth.txt", "x")),
		tfs.WithDir("sprites",
			tfs.WithFile("human.csv", "x"),
			tfs.WithFile("demon.xml", "x"),
			tfs.WithFile("alien.ini", "x")),
		tfs.WithFile("manifest.json", "x"),
		tfs.WithFile("crash.bin", "x"),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"logs/archive.zip", "manifest.json", "rules/index.csv",
		"sprites/demon.xml", "sprites/human.csv"), files)
}

func TestGomplateignore_WithIncludes(t *testing.T) {
	files, err := executeOpts(t, `.gomplateignore
*.log
`, []string{
		"--include", "rules/*",
		"--exclude", "rules/*.txt",
	},
		tfs.WithDir("logs",
			tfs.WithFile("archive.zip", "x"),
			tfs.WithFile("engine.log", "x"),
			tfs.WithFile("skills.log", "x")),
		tfs.WithDir("rules",
			tfs.WithFile("index.csv", "x"),
			tfs.WithFile("fire.txt", "x"),
			tfs.WithFile("earth.txt", "x")),
		tfs.WithFile("manifest.json", "x"),
		tfs.WithFile("crash.bin", "x"),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes("rules/index.csv"), files)
}

func TestGomplateignore_WithExcludeProcessing(t *testing.T) {
	files, err := executeOpts(t, `.gomplateignore
*.log
`, []string{
		"--exclude-processing", "crash.bin",
		"--exclude-processing", "log/*.zip",
		"--exclude", "rules/*.txt",
		"--exclude", "sprites/*.ini",
	},
		tfs.WithDir("logs",
			tfs.WithFile("archive.zip", "xxx"),
			tfs.WithFile("engine.log", "xxx"),
			tfs.WithFile("skills.log", "xxx")),
		tfs.WithDir("rules",
			tfs.WithFile("index.csv", "xxx"),
			tfs.WithFile("fire.txt", "xxx"),
			tfs.WithFile("earth.txt", "xxx")),
		tfs.WithDir("sprites",
			tfs.WithFile("human.csv", "xxx"),
			tfs.WithFile("demon.xml", "xxx"),
			tfs.WithFile("alien.ini", "xxx")),
		tfs.WithFile("manifest.json", "xxx"),
		tfs.WithFile("crash.bin", "xxx"),
	)

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"crash.bin", "logs/archive.zip", "manifest.json", "rules/index.csv",
		"sprites/demon.xml", "sprites/human.csv"), files)
}
