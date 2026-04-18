package domain

import (
	"context"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

func TestLatencyFault_InjectFault(t *testing.T) {
	tests := []struct {
		name           string
		delay          int
		ctxTimeout     time.Duration
		wantTerminate  bool
		wantStatusCode int
		wantFault      string
	}{
		{
			name:           "injects latency and continues request",
			delay:          50,
			ctxTimeout:     0,
			wantTerminate:  false,
			wantStatusCode: 200,
			wantFault:      constants.LATENCY,
		},
		{
			name:           "client disconnects during latency injection",
			delay:          5000,
			ctxTimeout:     10 * time.Millisecond,
			wantTerminate:  true,
			wantStatusCode: 499,
			wantFault:      constants.LATENCY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &models.Fault{
				Types: map[string]models.FaultConfig{
					constants.LATENCY: {Delay: tt.delay},
				},
			}
			f := &LatencyFault{Config: cfg}

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
			if resp.Fault != tt.wantFault {
				t.Errorf("Fault = %q, want %q", resp.Fault, tt.wantFault)
			}
			if !resp.Applied {
				t.Error("Applied should be true for injected fault")
			}

			if tt.ctxTimeout == 0 {
				expectedDelay := time.Duration(tt.delay) * time.Millisecond
				tolerance := 20 * time.Millisecond
				if elapsed < expectedDelay-tolerance {
					t.Errorf("elapsed %v shorter than expected delay %v", elapsed, expectedDelay)
				}
			}
		})
	}
}
