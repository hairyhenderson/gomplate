package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"io"
	"os/user"
	"regexp"
	"strconv"
)

// == Direct input processing ========================================

func processInputFiles(stringTemplate string, input []string, output []string, owner string, g *Gomplate) error {
	input, err := readInputs(stringTemplate, input)
	if err != nil {
		return err
	}

	if len(output) == 0 {
		output = []string{"-"}
	}

	for n, input := range input {
		if err := renderTemplate(g, input, output[n], owner); err != nil {
			return err
		}
	}
	return nil
}

// == Recursive input dir processing ======================================

func processInputDir(input string, output string, owner string, g *Gomplate) error {
	input = filepath.Clean(input)
	output = filepath.Clean(output)

	// prepare input and output directories
	if err := prepareDirectories(input, output, owner); err != nil {
		return err
	}

	// read directory
	entries, err := ioutil.ReadDir(input)
	if err != nil {
		return err
	}

	// process or dive in again
	for _, entry := range entries {
		nextInPath := filepath.Join(input, entry.Name())
		nextOutPath := filepath.Join(output, entry.Name())

		if entry.IsDir() {
			err := processInputDir(nextInPath, nextOutPath, owner, g)
			if err != nil {
				return err
			}
		} else {
			inString, err := readInput(nextInPath)
			if err != nil {
				return err
			}
			if err := renderTemplate(g, inString, nextOutPath, owner); err != nil {
				return err
			}
		}
	}
	return nil
}

// == File handling ================================================

func prepareDirectories(input string, output string, owner string) error {
	// assert that input path exists
	si, err := statDir(input)
	if err != nil {
		return err
	}

	// ensure output directory
	if err = os.MkdirAll(output, si.Mode()); err != nil {
		return err
	}

	// check owner
	if owner != "" {
		if err = chown(output, owner); err != nil {
			return err
		}
	}
	return nil
}

func statDir(dir string) (os.FileInfo, error) {
	si, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !si.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return si, nil
}

func readInputs(input string, files []string) ([]string, error) {
	if input != "" {
		return []string{input}, nil
	}
	if len(files) == 0 {
		files = []string{"-"}
	}
	ins := make([]string, len(files))

	for n, filename := range files {
		inString, err := readInput(filename)
		if err != nil {
			return nil, err
		}
		ins[n] = inString
	}
	return ins, nil
}

func readInput(filename string) (string, error) {
	var err error
	var inFile *os.File
	if filename == "-" {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(filename)
		if err != nil {
			return "", fmt.Errorf("failed to open %s\n%v", filename, err)
		}
		defer checkClose(inFile, &err)
	}
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = fmt.Errorf("read failed for %s\n%v", filename, err)
		return "", err
	}
	return string(bytes), nil
}

func openOutFile(filename string) (out *os.File, err error) {
	if filename == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func chown(filename string, owner string) error {
	r := regexp.MustCompile("^([^:]+)(?::(.+))?$")
	found := r.FindStringSubmatch(owner)
	var uid, gid int
	var err error
	if len(found) == 0 {
		return fmt.Errorf("invalid owner specification '%s'. Must match 'uid:gid'", owner)
	}
	uid, err = extractID(found[1], "user", func(u *user.User) string { return u.Uid })
	if err != nil {
		return err
	}
	if len(found) == 2 || found[2] == "" {
		gid = os.Getgid()
	} else {
		gid, err = extractID(found[2], "group", func(u *user.User) string { return u.Gid })
		if err != nil {
			return err
		}
	}
	return os.Chown(filename, uid, gid)
}

type userExtractor func(u *user.User) string

func extractID(name string, label string, extractor userExtractor) (int, error) {
	id, err := strconv.Atoi(name)
	if err != nil {
		u, err := user.Lookup(name)
		if err != nil {
			return 0, fmt.Errorf("cannot lookup %s %s : %v", label, name, err)
		}
		id, err = strconv.Atoi(extractor(u))
		if err != nil {
			return 0, fmt.Errorf("%sid %s is not numeric: %v", label, extractor(u), err)
		}
	}
	return id, nil
}

func checkClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}
