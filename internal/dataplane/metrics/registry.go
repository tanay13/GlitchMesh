package metrics

import "github.com/tanay13/GlitchMesh/internal/shared/constants"

var RegisteredMetrics = map[string]Metrics{
	constants.FAULT_METRICS: NewFaultMetrics(),
}
