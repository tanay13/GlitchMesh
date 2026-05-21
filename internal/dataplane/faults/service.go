package faults

import (
	"context"
	"log"

	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

type FaultService struct {
	faultInjector *FaultInjector
	logger        *log.Logger
}

func NewFaultService(faultInjector *FaultInjector, logger *log.Logger) *FaultService {

	return &FaultService{
		faultInjector,
		logger,
	}
}

func (fs *FaultService) ApplyFault(ctx context.Context, faultConfig models.Fault) *FaultResponse {

	response := fs.faultInjector.ProcessFault(ctx, faultConfig)

	return response
}
