package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type LatencyFault struct {
	Config *models.Fault
}

const (
	FAULT_INJECTED = "Latency Fault Injected"
)

func (f *LatencyFault) InjectFault(ctx context.Context) FaultResponse {
	delay := time.Duration(f.Config.Types[constants.LATENCY].Delay) * time.Millisecond

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return FaultResponse{
			Applied:         true,
			ShouldTerminate: false,
			StatusCode:      http.StatusOK,
			Message:         FAULT_INJECTED,
			Body:            nil,
			ContextErr:      nil,
		}

	case <-ctx.Done():
		return FaultResponse{
			Applied:         true,
			ShouldTerminate: true,
			StatusCode:      499,
			Message:         "Client disconnected during latency injection",
			Body:            nil,
			ContextErr:      ctx.Err(),
		}
	}
}
