package domain

import "github.com/tanay13/GlitchMesh/internal/models"

type ErrorFault struct {
	Config *models.FaultConfig
}

const (
	ERROR_INJECTED_MSG = "Error Injected"
)

func (f *ErrorFault) InjectFault() FaultResponse {
	return FaultResponse{
		true,
		"",
		true,
		f.Config.Error.StatusCode,
		ERROR_INJECTED_MSG,
		nil,
	}
}
