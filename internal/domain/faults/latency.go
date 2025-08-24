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
	time.Sleep(time.Duration(f.Config.Types[constants.LATENCY].Delay) * time.Millisecond)

	return FaultResponse{
		Applied:         true,
		TargetUrl:       "",
		ShouldTerminate: false,
		StatusCode:      http.StatusOK,
		Message:         FAULT_INJECTED,
		Body:            nil,
		ContextErr:      ctx.Err(),
	}
}
