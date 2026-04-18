package domain

import (
	"context"
	"log"
	"math/rand"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/metrics"
	"github.com/tanay13/GlitchMesh/internal/models"
)

// FaultInjector is a stateless service that processes fault configurations
// and applies faults based on priority and probability rules.
type FaultInjector struct{}

func (fi *FaultInjector) ProcessFault(ctx context.Context, faultConfig models.Fault) *FaultResponse {
	if !shouldApply(faultConfig) {
		return &FaultResponse{
			Applied: false,
		}
	}

	faults := buildFaultList(faultConfig)

	for _, fault := range faults {
		details := fault.InjectFault(ctx)

		metrics.RegisteredMetrics[constants.FAULT_METRICS].Increment(constants.TOTAL_FAULTS_INJECTED, 1)
		metrics.RegisteredMetrics[constants.FAULT_METRICS].Increment(details.Fault, 1)

		log.Printf("Fault applied: %s (Status: %d)", details.Message, details.StatusCode)

		if details.ShouldTerminate {
			return &details
		}
	}

	return &FaultResponse{
		Applied: true,
	}
}

func shouldApply(faultConfig models.Fault) bool {
	if !faultConfig.Enabled {
		return false
	}

	if faultConfig.Probability == 0 {
		return true
	}

	return rand.Float64() < faultConfig.Probability
}

// buildFaultList creates the fault implementations for the given config
// and returns them ordered by the configured priority.
func buildFaultList(faultConfig models.Fault) []Fault {
	faultMap := map[string]Fault{
		constants.ERROR:   &ErrorFault{Config: &faultConfig},
		constants.LATENCY: &LatencyFault{Config: &faultConfig},
		constants.TIMEOUT: &TimeoutFault{Config: &faultConfig},
	}

	faultList := make([]Fault, 0, len(faultConfig.Priority))
	for _, name := range faultConfig.Priority {
		if f, ok := faultMap[name]; ok {
			faultList = append(faultList, f)
		}
	}

	return faultList
}
