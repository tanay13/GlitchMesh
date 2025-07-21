package utils

import "github.com/tanay13/GlitchMesh/internal/models"

func GetServiceConfig(serviceName string, proxyConfig *models.Proxy) *models.ServiceConfig {
	for _, config := range proxyConfig.Service {
		if config.Name == serviceName {
			return &config
		}
	}
	return nil
}
