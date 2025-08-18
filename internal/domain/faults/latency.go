package domain

import (
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

func (f *LatencyFault) InjectFault() FaultResponse {
	time.Sleep(time.Duration(f.Config.Types[constants.LATENCY].Delay) * time.Millisecond)

	return FaultResponse{
		true,
		"",
		false,
		http.StatusOK,
		FAULT_INJECTED,
		nil,
	}

}
