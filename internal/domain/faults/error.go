package domain

import (
	"context"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

type ErrorFault struct {
	Config *models.Fault
}

const (
	ERROR_INJECTED_MSG = "Error Injected"
)

func (f *ErrorFault) InjectFault(ctx context.Context) FaultResponse {
	message := ERROR_INJECTED_MSG

	if f.Config.Types[constants.ERROR].Message != "" {
		message = f.Config.Types[constants.ERROR].Message
	}

	return FaultResponse{
		Applied:         true,
		TargetUrl:       "",
		ShouldTerminate: true,
		StatusCode:      f.Config.Types[constants.ERROR].StatusCode,
		Message:         message,
		Body:            nil,
		ContextErr:      ctx.Err(),
	}
}
