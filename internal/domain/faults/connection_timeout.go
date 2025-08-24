package domain

import (
	"context"
	"net/http"
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

	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)

	defer cancel()

	_ = <-ctx.Done()

	message := TIMEOUT_FAULT_INJECTED

	if f.Config.Types[constants.TIMEOUT].Message != "" {
		message = f.Config.Types[constants.TIMEOUT].Message
	}

	return FaultResponse{
		Applied:         true,
		TargetUrl:       "",
		ShouldTerminate: true,
		StatusCode:      http.StatusOK,
		Message:         message,
		Body:            nil,
		ContextErr:      ctx.Err(),
	}
}
