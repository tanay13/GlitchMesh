package domain

import (
	"log"

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

	return &FaultInjector{
		isEnabled,
		faultMap,
	}
}

func (fi *FaultInjector) ProcessFault(faultConfig models.Fault) *FaultResponse {

	injector := NewFaultInjector(faultConfig)

	fi.IsFaultEnabled = injector.IsFaultEnabled
	fi.Faults = injector.Faults

	if !fi.shouldApply(faultConfig) {
		return &FaultResponse{
			Applied: false,
		}
	}

	faults := fi.getFaults(faultConfig.Types)

	for _, fault := range faults {

		details := fault.InjectFault()

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
	return faultConfig.Enabled
}

func (fi *FaultInjector) getFaults(faults map[string]models.FaultConfig) []Fault {

	faultList := make([]Fault, 0)

	for name, _ := range faults {
		faultList = append(faultList, fi.Faults[name])
	}

	return faultList

}
