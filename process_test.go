package main

import (
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"log"

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

func TestInputDirSerial(t *testing.T) {
	inputDirTest(t, false, serialProcessor{})
}

func TestInputDirParallel(t *testing.T) {
	inputDirTest(t, true, parallelProcessor{})
}

func inputDirTest(t *testing.T, parallel bool, p processor) {
	outDir, err := ioutil.TempDir("test/files/input-dir", "out-temp-")
	assert.Nil(t, err)
	defer (func() {
		if cerr := os.RemoveAll(outDir); cerr != nil {
			log.Fatalf("Error while removing temporary directory %s : %v", outDir, cerr)
		}
	})()

	src, err := ParseSource("config=test/files/input-dir/config.yml")
	assert.Nil(t, err)

	data := &Data{
		Sources:  map[string]*Source{"config": src},
		parallel: parallel,
	}
	gomplate := NewGomplate(data, "{{", "}}")

	err = p.processInputDir("test/files/input-dir/in", outDir, gomplate)
	assert.Nil(t, err)

	top, err := ioutil.ReadFile(filepath.Join(outDir, "top.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "eins", string(top))

	inner, err := ioutil.ReadFile(filepath.Join(outDir, "inner/nested.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "zwei", string(inner))
}
