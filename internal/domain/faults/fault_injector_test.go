package domain

import (
	"context"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
)

func TestFaultInjector_ProcessFault(t *testing.T) {
	tests := []struct {
		name        string
		config      models.Fault
		wantApplied bool
	}{
		{
			name: "disabled fault — skipped entirely",
			config: models.Fault{
				Enabled:  false,
				Priority: []string{constants.ERROR},
				Types: map[string]models.FaultConfig{
					constants.ERROR: {StatusCode: 500, Message: "boom"},
				},
			},
			wantApplied: false,
		},
		{
			name: "probability 0 means no filter — always applies",
			config: models.Fault{
				Enabled:     true,
				Probability: 0,
				Priority:    []string{constants.ERROR},
				Types: map[string]models.FaultConfig{
					constants.ERROR: {StatusCode: 500, Message: "boom"},
				},
			},
			wantApplied: true,
		},
		{
			name: "probability 1.0 — always applies",
			config: models.Fault{
				Enabled:     true,
				Probability: 1.0,
				Priority:    []string{constants.ERROR},
				Types: map[string]models.FaultConfig{
					constants.ERROR: {StatusCode: 500, Message: "boom"},
				},
			},
			wantApplied: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := &FaultInjector{}
			resp := fi.ProcessFault(context.Background(), tt.config)

			if resp.Applied != tt.wantApplied {
				t.Errorf("Applied = %v, want %v", resp.Applied, tt.wantApplied)
			}
		})
	}
}

// TestFaultInjector_ErrorTerminatesChain verifies that when an error fault
// is first in priority, it terminates the chain and latency never runs.
// This is important: error returns ShouldTerminate=true, so the loop exits early.
func TestFaultInjector_ErrorTerminatesChain(t *testing.T) {
	config := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{constants.ERROR, constants.LATENCY},
		Types: map[string]models.FaultConfig{
			constants.ERROR:   {StatusCode: 500, Message: "fail fast"},
			constants.LATENCY: {Delay: 5000},
		},
	}

	fi := &FaultInjector{}
	resp := fi.ProcessFault(context.Background(), config)

	if !resp.ShouldTerminate {
		t.Error("expected ShouldTerminate=true when error fault is first")
	}
	if resp.Fault != constants.ERROR {
		t.Errorf("Fault = %q, want %q — error should fire, not latency", resp.Fault, constants.ERROR)
	}
	if resp.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode)
	}
}

// Verifies that when latency is first, the delay happens, then the error fires and terminates.
func TestFaultInjector_LatencyThenError(t *testing.T) {
	config := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{constants.LATENCY, constants.ERROR}, // latency first
		Types: map[string]models.FaultConfig{
			constants.LATENCY: {Delay: 50},                           // 50ms — fast
			constants.ERROR:   {StatusCode: 503, Message: "delayed"}, // then error
		},
	}

	fi := &FaultInjector{}
	resp := fi.ProcessFault(context.Background(), config)

	// Error should still terminate (it runs second but still terminates)
	if !resp.ShouldTerminate {
		t.Error("expected ShouldTerminate=true — error runs after latency")
	}
	if resp.Fault != constants.ERROR {
		t.Errorf("Fault = %q, want %q — error should be the terminating fault", resp.Fault, constants.ERROR)
	}
}

func TestFaultInjector_UnknownFaultInPriority(t *testing.T) {
	config := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{"nonexistent"}, // no matching fault type
		Types:       map[string]models.FaultConfig{},
	}

	fi := &FaultInjector{}
	resp := fi.ProcessFault(context.Background(), config)

	// Should complete without panic, with Applied=true (loop ran, nothing terminated)
	if !resp.Applied {
		t.Error("expected Applied=true even with no matching faults")
	}
	if resp.ShouldTerminate {
		t.Error("should not terminate when no faults matched")
	}
}

func TestShouldApply(t *testing.T) {
	tests := []struct {
		name   string
		config models.Fault
		want   bool
	}{
		{
			name:   "disabled — never applies",
			config: models.Fault{Enabled: false, Probability: 0},
			want:   false,
		},
		{
			name:   "enabled with probability 0 — always applies",
			config: models.Fault{Enabled: true, Probability: 0},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldApply(tt.config)
			if got != tt.want {
				t.Errorf("shouldApply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldApply_ProbabilisticBehavior(t *testing.T) {
	config := models.Fault{Enabled: true, Probability: 0.5}

	applied := 0
	iterations := 10000

	for i := 0; i < iterations; i++ {
		if shouldApply(config) {
			applied++
		}
	}

	ratio := float64(applied) / float64(iterations)
	if ratio < 0.40 || ratio > 0.60 {
		t.Errorf("probability ratio = %.2f, expected ~0.50 (applied %d/%d)", ratio, applied, iterations)
	}
}
