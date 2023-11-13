package funcs

import (
	"context"

	"github.com/Masterminds/semver/v3"
)

// CreateSemverFuncs -
func CreateSemverFuncs(ctx context.Context) map[string]interface{} {
	ns := &SemverFuncs{ctx}
	return map[string]interface{}{
		"semver": func() interface{} { return ns },
	}
}

// SemverFuncs -
type SemverFuncs struct {
	ctx context.Context
}

// Semver -
func (SemverFuncs) Semver(version string) (*semver.Version, error) {
	return semver.NewVersion(version)
}

// CheckConstraint -
func (SemverFuncs) CheckConstraint(constraint, in string) (bool, error) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	v, err := semver.NewVersion(in)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}
