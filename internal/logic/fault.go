package logic

import (
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

func FaultInjection(faultConfig models.FaultConfig) {
	switch faultConfig.Type {
	case constants.LATENCY:
		time.Sleep(time.Duration(faultConfig.Value) * time.Millisecond)
	}
}
