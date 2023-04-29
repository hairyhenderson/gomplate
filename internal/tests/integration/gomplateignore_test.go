package integration

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/fs"
)

func setupGomplateignoreTest(t *testing.T) func(inFileOps ...fs.PathOp) *fs.Dir {
	basedir := "gomplate-gomplateignore-tests"

	inBuilder := func(inFileOps ...fs.PathOp) *fs.Dir {
		tmpDir := fs.NewDir(t, basedir,
			fs.WithDir("in", inFileOps...),
			fs.WithDir("out"),
		)
		t.Cleanup(tmpDir.Remove)
		return tmpDir
	}

	return inBuilder
}

func execute(t *testing.T, ignoreContent string, inFileOps ...fs.PathOp) ([]string, error) {
	return executeOpts(t, ignoreContent, []string{}, inFileOps...)
}

func executeOpts(t *testing.T, ignoreContent string, opts []string, inFileOps ...fs.PathOp) ([]string, error) {
	inBuilder := setupGomplateignoreTest(t)

	inFileOps = append(inFileOps, fs.WithFile(".gomplateignore", ignoreContent))
	tmpDir := inBuilder(inFileOps...)

	argv := []string{}
	argv = append(argv, opts...)
	argv = append(argv,
		"--input-dir", tmpDir.Join("in"),
		"--output-dir", tmpDir.Join("out"),
	)
	o, e, err := cmd(t, argv...).run()
	assertSuccess(t, o, e, err, "")

	files := []string{}

	fs := afero.NewBasePathFs(afero.NewOsFs(), tmpDir.Join("out"))
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

	return files, err
}

func TestGomplateignore_Simple(t *testing.T) {
	files, err := execute(t, `# all dot files
.*
*.log`,
		fs.WithFile("empty.log", ""),
		fs.WithFile("rain.txt", ""))

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

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"foo/bar/tool/lex.txt", "foo/tar/2.txt"), files)
}

func TestGomplateignore_Root(t *testing.T) {
	files, err := execute(t, `.gomplateignore
/1.txt`,
		fs.WithDir("sub",
			fs.WithFile("1.txt", ""),
			fs.WithFile("2.txt", ""),
		),
		fs.WithFile("1.txt", ""),
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

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(
		"!", "e2.txt", "en/e1.txt", "en/e2.txt"), files)
}

func TestGomplateignore_Nested(t *testing.T) {
	files, err := execute(t, `inner/foo.md`,
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

	require.NoError(t, err)
	assert.Equal(t, fromSlashes(".gomplateignore", "1.txt",
		"inner/.gomplateignore",
		"inner/inner2/.gomplateignore",
		"inner/inner2/jess.ini"), files)
}

func TestGomplateignore_ByName(t *testing.T) {
	files, err := execute(t, `.gomplateignore
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
		fs.WithDir("loss.txt",
			fs.WithFile("1.log", ""),
			fs.WithFile("2.log", "")),
		fs.WithDir("foo",
			fs.WithFile("loss.txt", ""),
			fs.WithFile("bare.txt", "")),
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
		fs.WithDir("inner",
			fs.WithFile("  what.txt", ""),
			fs.WithFile("  dart.log", "")),
		fs.WithDir("inner2",
			fs.WithFile("  what.txt", "")),
		fs.WithFile("  what.txt", ""),
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

	require.NoError(t, err)
	assert.Equal(t, fromSlashes("rules/index.csv"), files)
}
