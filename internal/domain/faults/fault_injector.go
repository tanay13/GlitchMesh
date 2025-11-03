package domain

import (
	"context"
	"log"
	"math/rand"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultInjector struct {
	IsFaultEnabled bool
	Faults         map[string]Fault
}

func NewFaultInjector(faultConfig models.Fault) *FaultInjector {
	isEnabled := faultConfig.Enabled
	faultMap := make(map[string]Fault)

	faultMap[constants.ERROR] = &ErrorFault{Config: &faultConfig}

	faultMap[constants.LATENCY] = &LatencyFault{Config: &faultConfig}

	faultMap[constants.TIMEOUT] = &TimeoutFault{Config: &faultConfig}

	return &FaultInjector{
		isEnabled,
		faultMap,
	}
}

func (fi *FaultInjector) ProcessFault(ctx context.Context, faultConfig models.Fault) *FaultResponse {

	injector := NewFaultInjector(faultConfig)

	fi.IsFaultEnabled = injector.IsFaultEnabled
	fi.Faults = injector.Faults

	if !fi.shouldApply(faultConfig) {
		return &FaultResponse{
			Applied: false,
		}
	}

	faults := fi.getFaults(faultConfig.Priority)

	for _, fault := range faults {

		details := fault.InjectFault(ctx)

		log.Printf("Fault applied: %s (Status: %d)", details.Message, details.StatusCode)

		if details.ShouldTerminate {
			return &details
		}
	}
	return &FaultResponse{
		Applied: true,
	}
}

func (fi *FaultInjector) shouldApply(faultConfig models.Fault) bool {

	if !faultConfig.Enabled {
		return false
	}

	if faultConfig.Probability == 0 {
		return true
	}
	randomFloat := rand.Float64()
	return randomFloat < faultConfig.Probability
}

func (fi *FaultInjector) getFaults(faultPriority []string) []Fault {

	faultList := make([]Fault, 0)

	for _, name := range faultPriority {
		faultList = append(faultList, fi.Faults[name])
	}

	return faultList

}
