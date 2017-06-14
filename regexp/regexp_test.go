package regexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	assert.Equal(t, "-T-T-", Replace("a(x*)b", "T", "-ab-axxb-"))
	assert.Equal(t, "--xx-", Replace("a(x*)b", "$1", "-ab-axxb-"))
	assert.Equal(t, "---", Replace("a(x*)b", "$1W", "-ab-axxb-"))
	assert.Equal(t, "-W-xxW-", Replace("a(x*)b", "${1}W", "-ab-axxb-"))

	assert.Equal(t, "Turing, Alan", Replace("(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)", "${last}, ${first}", "Alan Turing"))
}

func TestMatch(t *testing.T) {
	assert.True(t, Match(`^[a-z]+\[[0-9]+\]$`, "adam[23]"))
	assert.True(t, Match(`^[a-z]+\[[0-9]+\]$`, "eve[7]"))
	assert.False(t, Match(`^[a-z]+\[[0-9]+\]$`, "Job[48]"))
	assert.False(t, Match(`^[a-z]+\[[0-9]+\]$`, "snakey"))
}
