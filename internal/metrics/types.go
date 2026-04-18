package metrics

import (
	"sync"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultMetrics struct {
	mu     sync.RWMutex
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
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.Faults[metric] += value
}

func (fm *FaultMetrics) Set(metric string, value int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.Faults[metric] = value
}

func (fm *FaultMetrics) Get() map[string]int {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	// Return a copy to prevent race conditions on the caller side
	result := make(map[string]int, len(fm.Faults))
	for k, v := range fm.Faults {
		result[k] = v
	}
	return result
}

func (fm *FaultMetrics) Snapshot() models.FaultsMetricsSnapshot {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return models.FaultsMetricsSnapshot{
		TotalFaultsInjected: fm.Faults[constants.TOTAL_FAULTS_INJECTED],
		LatencyCount:        fm.Faults[constants.LATENCY],
		TimeoutCount:        fm.Faults[constants.TIMEOUT],
		ErrorCount:          fm.Faults[constants.ERROR],
	}
}
