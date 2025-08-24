package service

import (
	"context"
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

func (fs *FaultService) ApplyFault(ctx context.Context, faultConfig models.Fault) *domain.FaultResponse {

	response := fs.faultInjector.ProcessFault(ctx, faultConfig)

	return response
}
