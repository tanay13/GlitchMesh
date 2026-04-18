package domain

import (
	"context"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

// TestErrorFault_InjectFault verifies that the error fault:
// 1. Always terminates the request (ShouldTerminate = true)
// 2. Returns the configured status code
// 3. Uses the custom message when provided, falls back to default otherwise
func TestErrorFault_InjectFault(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		message       string // custom message in config
		wantMessage   string // expected message in response
		wantTerminate bool
	}{
		{
			name:          "returns custom error message",
			statusCode:    500,
			message:       "service exploded",
			wantMessage:   "service exploded",
			wantTerminate: true,
		},
		{
			name:          "falls back to default message when config message is empty",
			statusCode:    503,
			message:       "",
			wantMessage:   ERROR_INJECTED_MSG, // "Error Injected"
			wantTerminate: true,
		},
		{
			name:          "supports 429 rate limit status",
			statusCode:    429,
			message:       "rate limited",
			wantMessage:   "rate limited",
			wantTerminate: true,
		},
		{
			name:          "supports 502 bad gateway",
			statusCode:    502,
			message:       "upstream down",
			wantMessage:   "upstream down",
			wantTerminate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &models.Fault{
				Types: map[string]models.FaultConfig{
					constants.ERROR: {
						StatusCode: tt.statusCode,
						Message:    tt.message,
					},
				},
			}
			f := &ErrorFault{Config: cfg}

			resp := f.InjectFault(context.Background())

			if resp.ShouldTerminate != tt.wantTerminate {
				t.Errorf("ShouldTerminate = %v, want %v", resp.ShouldTerminate, tt.wantTerminate)
			}
			if resp.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", resp.Message, tt.wantMessage)
			}
			if resp.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.statusCode)
			}
			if resp.Fault != constants.ERROR {
				t.Errorf("Fault = %q, want %q", resp.Fault, constants.ERROR)
			}
			if !resp.Applied {
				t.Error("Applied should be true for injected fault")
			}
		})
	}
}
