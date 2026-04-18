// Package testutil provides shared test helpers and factory functions
// used across all test layers (unit, integration, e2e).
// Import this in any _test.go file to create consistent test fixtures.
package testutil

import (
	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

// NewFaultConfig creates a fault config for testing with sensible defaults.
// Use faultType constants like constants.LATENCY, constants.ERROR, constants.TIMEOUT.
// When enabled=true and probability=0, faults always apply (probability 0 means "no probability filter").
func NewFaultConfig(faultType string, enabled bool) models.Fault {
	return models.Fault{
		Enabled:     enabled,
		Probability: 0, // 0 = always apply when enabled
		Priority:    []string{faultType},
		Types: map[string]models.FaultConfig{
			constants.LATENCY: {Delay: 50},                                          // 50ms — fast enough for tests
			constants.ERROR:   {StatusCode: 500, Message: "test error"},              // standard server error
			constants.TIMEOUT: {TimeoutDuration: 50, StatusCode: 504, Message: ""},   // 50ms timeout
		},
	}
}

// NewMultiFaultConfig creates a config with multiple fault types in priority order.
// Example: NewMultiFaultConfig(true, "latency", "error") applies latency first, then error.
func NewMultiFaultConfig(enabled bool, priorities ...string) models.Fault {
	return models.Fault{
		Enabled:     enabled,
		Probability: 0,
		Priority:    priorities,
		Types: map[string]models.FaultConfig{
			constants.LATENCY: {Delay: 50},
			constants.ERROR:   {StatusCode: 500, Message: "test error"},
			constants.TIMEOUT: {TimeoutDuration: 50, StatusCode: 504},
		},
	}
}

// NewTestProxy creates a proxy config with a single service.
// url should be the httptest.Server URL (e.g., backend.URL + "/").
func NewTestProxy(name, url string, fault models.Fault) *models.Proxy {
	return &models.Proxy{
		Service: []models.ServiceConfig{
			{Name: name, Url: url, Fault: fault},
		},
	}
}

// NewTestProxyMultiService creates a proxy config with multiple services.
func NewTestProxyMultiService(services ...models.ServiceConfig) *models.Proxy {
	return &models.Proxy{
		Service: services,
	}
}
