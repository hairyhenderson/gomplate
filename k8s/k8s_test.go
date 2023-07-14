package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test IsHealthy with a kubernetes pod spec in running mode
func TestIsHealthySvc(t *testing.T) {
	r := GetHealth(TestHealthy)
	assert.Equal(t, true, r.OK)
	assert.Equal(t, "Healthy", r.Status)
}

func TestIsHealthyPod(t *testing.T) {
	r := GetHealth(TestUnhealthy)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "Degraded", r.Status)
}

func TestIsHealthyAppset(t *testing.T) {
	r := GetHealth(TestLuaStatus)
	assert.Equal(t, false, r.OK)
	assert.Equal(t, "Degraded", r.Status)
	assert.Equal(t, "found less than two generators, Merge requires two or more", r.Message)
}

func TestIsHealthySvcPending(t *testing.T) {
	r := GetHealth(TestProgressing)
	assert.Equal(t, true, r.OK)
	assert.Equal(t, "Progressing", r.Status)
}
