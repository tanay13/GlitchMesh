package service

import (
	"log"

	domain "github.com/tanay13/GlitchMesh/internal/domain/faults"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type FaultService struct {
	faultInjector *domain.FaultInjector
	logger        *log.Logger
}

func NewFaultService(faultInjector *domain.FaultInjector, logger *log.Logger) *FaultService {

	return &FaultService{
		faultInjector,
		logger,
	}
}

func (fs *FaultService) ApplyFault(faultConfig models.FaultConfig) *domain.FaultResponse {

	response := fs.faultInjector.ProcessFault(faultConfig)

	if response.Applied {
		fs.logger.Printf("Fault applied: %s (Status: %d)", response.Message, response.StatusCode)
	}

	return response
}
