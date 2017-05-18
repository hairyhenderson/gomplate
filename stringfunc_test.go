package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAll(t *testing.T) {
	sf := &stringFunc{}

	assert.Equal(t, "Replaced",
		sf.replaceAll("Orig", "Replaced", "Orig"))
	assert.Equal(t, "ReplacedReplaced",
		sf.replaceAll("Orig", "Replaced", "OrigOrig"))
}
