package domain

import (
	"log"

	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultInjectionService struct {
	Logger         *log.Logger
	IsFaultEnabled bool
	FaultsEnabled  map[Fault]any
}

func NewFaultInjectionService(logger *log.Logger, faultConfig models.FaultConfig) *FaultInjectionService {
	isEnabled := faultConfig.Enabled
	faultMap := map[Fault]any{}

	if faultConfig.Error != nil {
		faultMap[&ErrorFault{Config: &faultConfig}] = faultConfig.Error
	}

	if faultConfig.Latency != nil {
		faultMap[&LatencyFault{Config: &faultConfig}] = faultConfig.Error
	}

	return &FaultInjectionService{
		logger,
		isEnabled,
		faultMap,
	}
}

func (s *FaultInjectionService) ProcessFault(faultConfig models.FaultConfig) *FaultResponse {
	if !s.shouldApply(faultConfig) {
		return &FaultResponse{
			Applied: false,
		}
	}

	for fault := range s.FaultsEnabled {

		details := fault.InjectFault()

		s.Logger.Print("Fault Injected ", details.Message, " with status Code ", details.StatusCode)

		if details.ShouldTerminate {
			return &details
		}
	}
	return &FaultResponse{
		Applied: true,
	}
}

func (s *FaultInjectionService) shouldApply(faultConfig models.FaultConfig) bool {
	return faultConfig.Enabled
}
