// +build !windows

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"log"

	"path"

	"github.com/stretchr/testify/assert"
)

func TestDatasourceDir(t *testing.T) {
	outDir, err := ioutil.TempDir("test/files/datasource-dir", "out-temp-")
	assert.Nil(t, err)
	defer (func() {
		if cerr := os.RemoveAll(outDir); cerr != nil {
			log.Fatalf("Error while removing temporary directory %s : %v", outDir, cerr)
		}
	})()

	data := NewData([]string{"test/files/datasource-dir/ds"}, []string{})
	gomplate := NewGomplate(data, "{{", "}}")
	err = processInputFiles(
		"",
		[]string{"test/files/datasource-dir/in/test.txt"},
		[]string{path.Join(outDir, "out.txt")},
		gomplate)
	assert.NoError(t, err)

	out, err := ioutil.ReadFile(filepath.Join(outDir, "out.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "eins-deux", string(out))
}

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

	src, err := ParseSource("config=test/files/input-dir/config.yml")
	assert.Nil(t, err)

	data := &Data{
		Sources: map[string]*Source{"config": src},
	}
	gomplate := NewGomplate(data, "{{", "}}")
	err = processInputDir(filepath.Join("test", "files", "input-dir", "in"), outDir, gomplate)
	assert.Nil(t, err)

	top, err := ioutil.ReadFile(filepath.Join(outDir, "top.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "eins", string(top))

	inner, err := ioutil.ReadFile(filepath.Join(outDir, "inner/nested.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "zwei", string(inner))
}
