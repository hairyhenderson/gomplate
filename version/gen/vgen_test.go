package main

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	testdata := []struct {
		desc, latest string
		expected     string
	}{
		{"v1.0.0", "", "1.0.0"},
		{"v1.0.0-1-gabcdef0", "v1.0.0", "1.0.1-1-gabcdef0"},
		{"v1.0.0-1-gabcdef0", "v2.3.4", "2.3.5-1-gabcdef0"},
		{"v1.0.0+123", "v2.3.4", "2.3.5+123"},
	}

	for _, td := range testdata {
		var l *semver.Version
		if td.latest != "" {
			l = semver.MustParse(td.latest)
		}

		ver, err := version(semver.MustParse(td.desc), l)
		require.NoError(t, err)
		require.Equal(t, td.expected, ver.String())
	}
}
