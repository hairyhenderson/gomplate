package main

import (
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"log"

	"fmt"

	"github.com/docker/libcontainer/user"
	"github.com/stretchr/testify/assert"
)

func TestReadInput(t *testing.T) {
	actual, err := readInputs("foo", nil)
	assert.Nil(t, err)
	assert.Equal(t, "foo", actual[0])

	// stdin is "" because during tests it's given /dev/null
	actual, err = readInputs("", []string{"-"})
	assert.Nil(t, err)
	assert.Equal(t, "", actual[0])

	actual, err = readInputs("", []string{"main_test.go"})
	assert.Nil(t, err)
	thisFile, _ := os.Open("main_test.go")
	expected, _ := ioutil.ReadAll(thisFile)
	assert.Equal(t, string(expected), actual[0])
}

func removeDir(dir string) {
	if cerr := os.RemoveAll(dir); cerr != nil {
		log.Fatalf("Error while removing temporary directory %s : %v", dir, cerr)
	}
}

func TestInputDir(t *testing.T) {
	outDir, err := ioutil.TempDir("test/files/input-dir", "out-temp-")
	assert.NoError(t, err)
	defer removeDir(outDir)

	src, err := ParseSource("config=test/files/input-dir/config.yml")
	assert.NoError(t, err)

	data := &Data{
		Sources: map[string]*Source{"config": src},
	}
	gomplate := NewGomplate(data, "{{", "}}")
	err = processInputDir("test/files/input-dir/in", outDir, "", gomplate)
	assert.NoError(t, err)

	top, err := ioutil.ReadFile(filepath.Join(outDir, "top.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "eins", string(top))

	inner, err := ioutil.ReadFile(filepath.Join(outDir, "inner/nested.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "zwei", string(inner))
}

func TestChown(t *testing.T) {
	testDir, err := ioutil.TempDir(".", "test-temp")
	assert.NoError(t, err)
	defer removeDir(testDir)

	uid, gid := os.Getuid(), os.Getgid()
	err = chown(testDir, fmt.Sprintf("%d:%d", uid, gid))
	assert.NoError(t, err)

	err = chown(testDir, fmt.Sprintf("%d", uid))
	assert.NoError(t, err)

	// The following lookups might fail e.g. on OSX
	u, err := user.LookupUid(uid)
	if err != nil {
		g, err := user.LookupGid(gid)
		if err != nil {
			err = chown(testDir, fmt.Sprintf("%s:%s", u.Name, g.Name))
			assert.NoError(t, err)

			err = chown(testDir, fmt.Sprintf("%s", u.Name))
			assert.NoError(t, err)

			err = chown(testDir, fmt.Sprintf("%s", "blub"))
			assert.Error(t, err)
		}
	}
}
