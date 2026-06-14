package faults

import (
	"context"

	"github.com/tanay13/GlitchMesh/internal/shared/constants"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

type ConnectionDropFault struct {
	Config *models.Fault
}

func (f *ConnectionDropFault) InjectFault(ctx context.Context) FaultResponse {
	// ConnectionDropFault returns Applied: true and ShouldTerminate: false.
	// This indicates that the connection drop fault is registered for this request,
	// but preprocessing doesn't terminate it. The socket is dropped asynchronously
	// in the background wrapper.
	return FaultResponse{
		Fault:           constants.CONNECTION_DROP,
		Applied:         true,
		ShouldTerminate: false,
		Message:         "Connection drop fault registered",
	}
}
