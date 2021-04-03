package gomplate

import "time"

// Metrics tracks interesting basic metrics around gomplate executions. Warning: experimental!
// This may change in breaking ways without warning. This is not subject to any semantic versioning guarantees!
var Metrics *MetricsType

// MetricsType - Warning: experimental! This may change in breaking ways without warning.
// This is not subject to any semantic versioning guarantees!
type MetricsType struct {
	// times for rendering each template
	RenderDuration map[string]time.Duration
	// time it took to gather templates
	GatherDuration time.Duration
	// time it took to render all templates
	TotalRenderDuration time.Duration

	TemplatesGathered  int
	TemplatesProcessed int
	Errors             int
}

func newMetrics() *MetricsType {
	return &MetricsType{
		RenderDuration: make(map[string]time.Duration),
	}
}
