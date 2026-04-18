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

func CopyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func ParseURLParts(urlParts []string) (string, string, error) {
	if len(urlParts) < 2 {
		return "", "", fmt.Errorf("invalid URL format: expected at least 2 parts, got %d", len(urlParts))
	}
	if urlParts[0] == "" {
		return "", "", fmt.Errorf("service name cannot be empty")
	}
	return urlParts[0], urlParts[1], nil
}
