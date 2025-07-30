package utils

import (
	"fmt"
	"net/http"

	"github.com/tanay13/GlitchMesh/internal/models"
)

func GetServiceConfig(serviceName string, proxyConfig *models.Proxy) *models.ServiceConfig {
	for _, config := range proxyConfig.Service {
		if config.Name == serviceName {
			return &config
		}
	}
	return nil
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, message)
}
