// +build !windows

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"log"

	"github.com/hairyhenderson/gomplate/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInput(t *testing.T) {
	actual, err := readInputs("foo", nil)
	assert.Nil(t, err)
	assert.Equal(t, "foo", actual[0])

	// stdin is "" because during tests it's given /dev/null
	actual, err = readInputs("", []string{"-"})
	assert.Nil(t, err)
	assert.Equal(t, "", actual[0])

	actual, err = readInputs("", []string{"process_test.go"})
	assert.Nil(t, err)
	thisFile, _ := os.Open("process_test.go")
	expected, _ := ioutil.ReadAll(thisFile)
	assert.Equal(t, string(expected), actual[0])
}

func TestInputDir(t *testing.T) {
	outDir, err := ioutil.TempDir(filepath.Join("test", "files", "input-dir"), "out-temp-")
	assert.Nil(t, err)
	defer (func() {
		if cerr := os.RemoveAll(outDir); cerr != nil {
			log.Fatalf("Error while removing temporary directory %s : %v", outDir, cerr)
		}
	})()

	src, err := data.ParseSource("config=test/files/input-dir/config.yml")
	assert.Nil(t, err)

	d := &data.Data{
		Sources: map[string]*data.Source{"config": src},
	}
	gomplate := NewGomplate(d, "{{", "}}")
	err = processInputDir(filepath.Join("test", "files", "input-dir", "in"), outDir, gomplate)
	assert.Nil(t, err)

	top, err := ioutil.ReadFile(filepath.Join(outDir, "top.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "eins", string(top))

	inner, err := ioutil.ReadFile(filepath.Join(outDir, "inner/nested.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "zwei", string(inner))
}

func TestInputDir_WithIgnorefile_Simple(t *testing.T) {
	templateDir := filepath.Join("test", "files", "ignorefile")
	outDir, err := ioutil.TempDir(templateDir, "out-temp-")
	require.NoError(t, err)

	d := &data.Data{
		Sources: map[string]*data.Source{},
	}
	gomplate := NewGomplate(d, "{{", "}}")

	inDir := filepath.Join(templateDir)

	process := func(dir string) {
		err = processInputDir(filepath.Join(inDir, dir), outDir, gomplate)
		require.NoError(t, err)
	}

	clean := func() {
		if cerr := os.RemoveAll(outDir); cerr != nil {
			log.Fatalf("Error while removing temporary directory %s : %v", outDir, cerr)
		}
	}

	process("simple")
	defer clean()
	verifyFileExists(t, outDir, "site.xml")
	verifyFileExists(t, outDir, "inner/kill.json")
	verifyFileNotExists(t, outDir, "inner/hello.txt")
	verifyFileNotExists(t, outDir, "inner/hello.txt")
}

func TestInputDir_WithIgnorefile_Inherit(t *testing.T) {
	templateDir := filepath.Join("test", "files", "ignorefile")
	outDir, err := ioutil.TempDir(templateDir, "out-temp-")
	require.NoError(t, err)

	d := &data.Data{
		Sources: map[string]*data.Source{},
	}
	gomplate := NewGomplate(d, "{{", "}}")

	inDir := filepath.Join(templateDir)

	process := func(dir string) {
		err = processInputDir(filepath.Join(inDir, dir), outDir, gomplate)
		require.NoError(t, err)
	}

	clean := func() {
		if cerr := os.RemoveAll(outDir); cerr != nil {
			log.Fatalf("Error while removing temporary directory %s : %v", outDir, cerr)
		}
	}

	process("inherit")
	defer clean()
	verifyFileExists(t, outDir, "lv1/e.txt")
	verifyFileNotExists(t, outDir, "lv1/lv2/e.txt")
	verifyFileNotExists(t, outDir, ".gomplateignore")
	verifyFileNotExists(t, outDir, "a.txt")
	verifyFileNotExists(t, outDir, "b.txt")
	verifyFileNotExists(t, outDir, "c.txt")
	verifyFileNotExists(t, outDir, "lv1/.gomplateignore")
	verifyFileNotExists(t, outDir, "lv1/lv2/.gomplateignore")
}

func verifyFileNotExists(t *testing.T, basedir, file string) {
	file = filepath.Join(basedir, file)
	stat, err := os.Stat(file)
	assert.True(t, os.IsNotExist(err),
		file+" exists", stat)
}

func verifyFileExists(t *testing.T, basedir, file string) {
	file = filepath.Join(basedir, file)
	stat, _ := os.Stat(file)
	assert.True(t, stat != nil,
		file+" not exists")
}
