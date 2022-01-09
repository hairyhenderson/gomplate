package iohelpers

import (
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowsFileMode(t *testing.T) {
	data := []struct {
		in, expected os.FileMode
	}{
		{0o000, 0o000},
		{0o100, 0o000},
		{0o111, 0o000},
		{0o123, 0o000},
		{0o177, 0o000},
		{0o400, 0o444},
		{0o412, 0o444},
		{0o467, 0o444},
		{0o542, 0o444},
		{0o200, 0o666},
		{0o211, 0o666},
		{0o300, 0o666},
		{0o644, 0o666},
		{0o600, 0o666},
		{0o755, 0o666},
		{0o777, 0o666},
	}
	for _, d := range data {
		actual := windowsFileMode(d.in)
		assert.Equal(t, fmt.Sprintf("%o", d.expected), fmt.Sprintf("%o", actual))
		assert.Equal(t, d.expected, actual)
	}

	// directories are always 0777
	assert.Equal(t, 0o777|fs.ModeDir, windowsFileMode(0o755|fs.ModeDir))
}
