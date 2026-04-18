package domain

import (
	"context"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

func TestTimeoutFault_InjectFault(t *testing.T) {
	tests := []struct {
		name            string
		timeoutDuration int
		statusCode      int
		message         string
		ctxTimeout      time.Duration
		wantStatusCode  int
		wantTerminate   bool
		wantMessage     string
	}{
		{
			name:            "normal timeout fires — deadline exceeded",
			timeoutDuration: 50,
			statusCode:      504,
			message:         "",
			ctxTimeout:      0,
			wantStatusCode:  504,
			wantTerminate:   true,
			wantMessage:     TIMEOUT_FAULT_INJECTED,
		},
		{
			name:            "custom message on timeout",
			timeoutDuration: 50,
			statusCode:      500,
			message:         "gateway timeout",
			ctxTimeout:      0,
			wantStatusCode:  500,
			wantTerminate:   true,
			wantMessage:     "gateway timeout",
		},
		{
			name:            "client disconnects before timeout fires",
			timeoutDuration: 5000,
			statusCode:      504,
			message:         "",
			ctxTimeout:      10 * time.Millisecond,
			wantStatusCode:  499,
			wantTerminate:   true,
			wantMessage:     "Client disconnected during timeout simulation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &models.Fault{
				Types: map[string]models.FaultConfig{
					constants.TIMEOUT: {
						TimeoutDuration: tt.timeoutDuration,
						StatusCode:      tt.statusCode,
						Message:         tt.message,
					},
				},
			}
			f := &TimeoutFault{Config: cfg}

			// Set up context — optionally with early cancellation
			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			start := time.Now()
			resp := f.InjectFault(ctx)
			elapsed := time.Since(start)

			if resp.ShouldTerminate != tt.wantTerminate {
				t.Errorf("ShouldTerminate = %v, want %v", resp.ShouldTerminate, tt.wantTerminate)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
			if resp.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", resp.Message, tt.wantMessage)
			}
			if resp.Fault != constants.TIMEOUT {
				t.Errorf("Fault = %q, want %q", resp.Fault, constants.TIMEOUT)
			}

			// Verify the blocking duration is roughly right.
			// Normal timeout: should block for ~timeoutDuration.
			// Client disconnect: should unblock early (~ctxTimeout).
			if tt.ctxTimeout == 0 {
				expectedBlock := time.Duration(tt.timeoutDuration) * time.Millisecond
				tolerance := 30 * time.Millisecond
				if elapsed < expectedBlock-tolerance {
					t.Errorf("elapsed %v shorter than expected block %v", elapsed, expectedBlock)
				}
			}
		})
	}
}
