package domain

import (
	"context"
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type TimeoutFault struct {
	Config *models.Fault
}

const (
	TIMEOUT_FAULT_INJECTED = "Connection Timeout Fault Injected"
)

func (f *TimeoutFault) InjectFault(ctx context.Context) FaultResponse {
	timeoutDuration := time.Duration(f.Config.Types[constants.TIMEOUT].TimeoutDuration) * time.Millisecond
	timer := time.NewTimer(timeoutDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		message := TIMEOUT_FAULT_INJECTED
		if f.Config.Types[constants.TIMEOUT].Message != "" {
			message = f.Config.Types[constants.TIMEOUT].Message
		}

		return FaultResponse{
			Fault:           constants.TIMEOUT,
			Applied:         true,
			ShouldTerminate: true,
			StatusCode:      f.Config.Types[constants.TIMEOUT].StatusCode,
			Message:         message,
		}

	case <-ctx.Done():
		return FaultResponse{
			Fault:           constants.TIMEOUT,
			Applied:         true,
			ShouldTerminate: true,
			StatusCode:      499,
			Message:         "Client disconnected during timeout simulation",
			ContextErr:      ctx.Err(),
		}
	}
}
