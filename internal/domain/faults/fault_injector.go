package domain

import (
	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultInjector struct {
	IsFaultEnabled bool
	FaultsEnabled  map[Fault]any
}

func NewFaultInjector(faultConfig models.FaultConfig) *FaultInjector {
	isEnabled := faultConfig.Enabled
	faultMap := map[Fault]any{}

	if faultConfig.Error != nil {
		faultMap[&ErrorFault{Config: &faultConfig}] = faultConfig.Error
	}

	if faultConfig.Latency != nil {
		faultMap[&LatencyFault{Config: &faultConfig}] = faultConfig.Latency
	}

	return &FaultInjector{
		isEnabled,
		faultMap,
	}
}

func (fi *FaultInjector) ProcessFault(faultConfig models.FaultConfig) *FaultResponse {
	if !fi.shouldApply(faultConfig) {
		return &FaultResponse{
			Applied: false,
		}
	}

	injector := NewFaultInjector(faultConfig)

	for fault := range injector.FaultsEnabled {

		details := fault.InjectFault()

		if details.ShouldTerminate {
			return &details
		}
	}
	return &FaultResponse{
		Applied: true,
	}
}

func (fi *FaultInjector) shouldApply(faultConfig models.FaultConfig) bool {
	return faultConfig.Enabled
}
