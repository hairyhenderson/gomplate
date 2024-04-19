package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test IsHealthy with a kubernetes pod spec in running mode
func TestIsHealthySvc(t *testing.T) {
	r := GetHealth(TestHealthySvc)
	assert.Equal(t, true, r.OK)
	assert.Equal(t, "Running", r.Status)
	assert.Equal(t, "healthy", r.Health)
}

func TestIsHealthyPod(t *testing.T) {
	r := GetHealth(TestUnhealthy)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "CrashLoopBackOff", r.Status)
	assert.Equal(t, "unhealthy", r.Health)
}

func TestIsHealthyCertificate(t *testing.T) {
	r := GetHealth(TestHealthyCertificate)
	assert.Equal(t, true, r.OK)
	assert.Equal(t, "Issued", r.Status)
	assert.Equal(t, "healthy", r.Health)

	r = GetHealth(TestDegradedCertificate)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "Issuing", r.Status)
	assert.Equal(t, "unknown", r.Health)
}

func TestIsHealthyAppset(t *testing.T) {
	r := GetHealth(TestLuaStatus)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "unhealthy", r.Health)
	assert.Equal(t, "found less than two generators, Merge requires two or more", r.Message)
}

func TestIsHealthySvcPending(t *testing.T) {
	r := GetHealth(TestProgressing)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "Creating", r.Status)
	assert.Equal(t, "unknown", r.Health)
}
