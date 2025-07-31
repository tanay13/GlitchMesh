package logic

import (
	"net/http"
	"strconv"
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
	"github.com/tanay13/GlitchMesh/internal/utils"
)

func FaultInjection(w http.ResponseWriter, faultConfig models.FaultConfig) (bool, string, string) {
	switch faultConfig.Type {
	case constants.LATENCY:
		time.Sleep(time.Duration(faultConfig.Value) * time.Millisecond)
		return false, faultConfig.Type, strconv.Itoa(faultConfig.Value)
	case constants.ERROR:
		utils.WriteJSONError(w, faultConfig.Value, "fault injected")
		return true, faultConfig.Type, strconv.Itoa(faultConfig.Value)
	}
	return false, faultConfig.Type, strconv.Itoa(faultConfig.Value)
}
