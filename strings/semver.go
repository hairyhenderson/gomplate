package strings

import (
	sv2 "github.com/Masterminds/semver/v3"
)

func SemverCompare(constraint, version string) (bool, error) {
	c, err := sv2.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	v, err := sv2.NewVersion(version)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}

func Semver(version string) (*sv2.Version, error) {
	return sv2.NewVersion(version)
}
