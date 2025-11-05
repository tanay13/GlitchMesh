package metrics

import (
	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultMetrics struct {
	Faults map[string]int
}

func NewFaultMetrics() *FaultMetrics {
	return &FaultMetrics{
		Faults: map[string]int{
			constants.TOTAL_FAULTS_INJECTED: 0,
			constants.ERROR:                 0,
			constants.LATENCY:               0,
			constants.TIMEOUT:               0,
		},
	}
}

func (fm *FaultMetrics) Increment(metric string, value int) {
	fm.Faults[metric] += value
}

func (fm *FaultMetrics) Set(metric string, value int) {
	fm.Faults[metric] = value
}

func (fm *FaultMetrics) Get() map[string]int {
	return fm.Faults
}

func (fm *FaultMetrics) Snapshot() models.FaultsMetricsSnapshot {
	return models.FaultsMetricsSnapshot{
		TotalFaultsInjected: fm.Faults[constants.TOTAL_FAULTS_INJECTED],
		LatencyCount:        fm.Faults[constants.LATENCY],
		TimeoutCount:        fm.Faults[constants.TIMEOUT],
		ErrorCount:          fm.Faults[constants.ERROR],
	}
}
