// run as 'go run ./version/gen/vgen.go'

package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func main() {
	descVer, err := describedVersion()
	if err != nil {
		log.Fatal(err)
	}

	latest, err := latestTag()
	if err != nil {
		log.Fatal(err)
	}

	ver := version(descVer, latest)

	fmt.Println(ver.String())
}

func describedVersion() (*semver.Version, error) {
	desc, err := runCmd("git describe --always")
	if err != nil {
		return nil, fmt.Errorf("git describe failed: %w", err)
	}

	ver, err := semver.NewVersion(desc)
	if err != nil {
		return nil, err
	}

	return ver, nil
}

func version(descVer, latest *semver.Version) *semver.Version {
	ver := *descVer
	if ver.Prerelease() != "" {
		ver = ver.IncPatch().IncPatch()
		ver, _ = ver.SetPrerelease(descVer.Prerelease())
		ver, _ = ver.SetMetadata(descVer.Metadata())
	} else if ver.Metadata() != "" {
		ver = ver.IncPatch()
		ver, _ = ver.SetMetadata(descVer.Metadata())
	}

	// if we're on a release tag already, we're done
	if descVer.Equal(&ver) {
		return descVer
	}

	// If 'latest' is greater than 'ver', we need to skip to the next patch.
	// If 'latest' is already a prerelease (i.e. if v5.0.0-pre was tagged),
	// we should use
	// If 'latest' is a prerelease, it's the same logic, except that we don't
	// want to increment the patch version.
	if latest.GreaterThan(&ver) || latest.Prerelease() != "" {
		v := *latest
		if v.Prerelease() == "" {
			v = v.IncPatch()
		}
		v, _ = v.SetPrerelease(ver.Prerelease())
		v, _ = v.SetMetadata(ver.Metadata())

		ver = v
	}

	return &ver
}

func latestTag() (*semver.Version, error) {
	// get the latest tag
	tags, err := runCmd("git tag --list v*")
	if err != nil {
		return nil, fmt.Errorf("git tag failed: %w", err)
	}

	// find the latest tag
	var latest *semver.Version
	for _, tag := range strings.Split(tags, "\n") {
		ver, err := semver.NewVersion(tag)
		if err != nil {
			return nil, fmt.Errorf("parsing tag %q failed: %w", tag, err)
		}

		if latest == nil || ver.GreaterThan(latest) {
			latest = ver
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no tags found")
	}

	return latest, nil
}

func runCmd(c string) (string, error) {
	parts := strings.Split(c, " ")
	//nolint:gosec
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
