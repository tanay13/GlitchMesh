package utils

import (
	"testing"

	"github.com/tanay13/GlitchMesh/internal/models"
)

// TestParseURLParts verifies URL splitting with bounds checking.
// The function should return an error for malformed inputs instead of panicking.
func TestParseURLParts(t *testing.T) {
	tests := []struct {
		name         string
		parts        []string
		wantService  string
		wantEndpoint string
		wantErr      bool
	}{
		{
			name:         "valid two-part URL",
			parts:        []string{"auth-service", "api/users"},
			wantService:  "auth-service",
			wantEndpoint: "api/users",
			wantErr:      false,
		},
		{
			name:         "endpoint with nested path",
			parts:        []string{"payments", "api/v1/charge"},
			wantService:  "payments",
			wantEndpoint: "api/v1/charge",
			wantErr:      false,
		},
		{
			name:    "empty slice — should error, not panic",
			parts:   []string{},
			wantErr: true,
		},
		{
			name:    "single element — missing endpoint",
			parts:   []string{"service-only"},
			wantErr: true,
		},
		{
			name:    "empty service name",
			parts:   []string{"", "endpoint"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, ep, err := ParseURLParts(tt.parts)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURLParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if svc != tt.wantService {
					t.Errorf("service = %q, want %q", svc, tt.wantService)
				}
				if ep != tt.wantEndpoint {
					t.Errorf("endpoint = %q, want %q", ep, tt.wantEndpoint)
				}
			}
		})
	}
}

// TestGetServiceConfig verifies service lookup by name in the proxy config.
// Should return nil when the service doesn't exist.
func TestGetServiceConfig(t *testing.T) {
	proxy := &models.Proxy{
		Service: []models.ServiceConfig{
			{Name: "auth", Url: "http://localhost:8081/"},
			{Name: "users", Url: "http://localhost:8082/"},
			{Name: "payments", Url: "http://localhost:8083/"},
		},
	}

	tests := []struct {
		name    string
		service string
		wantNil bool
		wantUrl string
	}{
		{
			name:    "finds existing service",
			service: "auth",
			wantNil: false,
			wantUrl: "http://localhost:8081/",
		},
		{
			name:    "finds another service",
			service: "payments",
			wantNil: false,
			wantUrl: "http://localhost:8083/",
		},
		{
			name:    "returns nil for unknown service",
			service: "unknown-service",
			wantNil: true,
		},
		{
			name:    "returns nil for empty name",
			service: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetServiceConfig(tt.service, proxy)

			if (cfg == nil) != tt.wantNil {
				t.Errorf("GetServiceConfig(%q) nil = %v, want %v", tt.service, cfg == nil, tt.wantNil)
				return
			}
			if cfg != nil && cfg.Url != tt.wantUrl {
				t.Errorf("Url = %q, want %q", cfg.Url, tt.wantUrl)
			}
		})
	}
}

// TestGetServiceConfig_EmptyProxy ensures no panic when proxy has no services.
func TestGetServiceConfig_EmptyProxy(t *testing.T) {
	proxy := &models.Proxy{Service: []models.ServiceConfig{}}

	cfg := GetServiceConfig("anything", proxy)
	if cfg != nil {
		t.Error("expected nil for empty proxy config")
	}
}
