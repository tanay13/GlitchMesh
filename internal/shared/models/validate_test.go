package models

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Fault.Validate() tests
// ---------------------------------------------------------------------------

func TestFaultValidate_ValidConfig(t *testing.T) {
	f := Fault{
		Enabled:     true,
		Probability: 0.5,
		Priority:    []string{"latency", "error"},
		Types: map[string]FaultConfig{
			"latency": {Delay: 100},
			"error":   {StatusCode: 500, Message: "boom"},
		},
	}
	if err := f.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestFaultValidate_ZeroProbabilityIsValid(t *testing.T) {
	f := Fault{
		Enabled:     true,
		Probability: 0, // means "always apply"
		Priority:    []string{"latency"},
		Types: map[string]FaultConfig{
			"latency": {Delay: 50},
		},
	}
	if err := f.Validate(); err != nil {
		t.Errorf("probability=0 should be valid, got: %v", err)
	}
}

func TestFaultValidate_NegativeProbability(t *testing.T) {
	f := Fault{Probability: -0.1}
	if err := f.Validate(); err == nil {
		t.Error("expected error for negative probability")
	}
}

func TestFaultValidate_ProbabilityAboveOne(t *testing.T) {
	f := Fault{Probability: 1.5}
	if err := f.Validate(); err == nil {
		t.Error("expected error for probability > 1")
	}
}

func TestFaultValidate_UnknownFaultTypeInTypes(t *testing.T) {
	f := Fault{
		Probability: 0,
		Types:       map[string]FaultConfig{"banana": {StatusCode: 418}},
	}
	err := f.Validate()
	if err == nil {
		t.Fatal("expected error for unknown type key")
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Errorf("error should mention the bad key, got: %v", err)
	}
}

func TestFaultValidate_UnknownFaultTypeInPriority(t *testing.T) {
	f := Fault{
		Probability: 0,
		Priority:    []string{"latency", "unicorn"},
		Types: map[string]FaultConfig{
			"latency": {Delay: 50},
		},
	}
	err := f.Validate()
	if err == nil {
		t.Fatal("expected error for unknown priority entry")
	}
	if !strings.Contains(err.Error(), "unicorn") {
		t.Errorf("error should mention 'unicorn', got: %v", err)
	}
}

func TestFaultValidate_PriorityEntryMissingFromTypes(t *testing.T) {
	f := Fault{
		Probability: 0,
		Priority:    []string{"latency", "error"},
		Types: map[string]FaultConfig{
			"latency": {Delay: 50},
			// "error" is intentionally absent from Types
		},
	}
	err := f.Validate()
	if err == nil {
		t.Fatal("expected error when priority entry has no Types entry")
	}
	if !strings.Contains(err.Error(), "error") {
		t.Errorf("error should mention 'error', got: %v", err)
	}
}

func TestFaultValidate_EmptyFaultIsValid(t *testing.T) {
	// A disabled fault with no priority / types is fine
	f := Fault{Enabled: false}
	if err := f.Validate(); err != nil {
		t.Errorf("empty disabled fault should be valid, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Proxy.Validate() tests
// ---------------------------------------------------------------------------

func TestProxyValidate_Valid(t *testing.T) {
	p := Proxy{
		Service: []ServiceConfig{
			{
				Name: "svc-a",
				Url:  "http://localhost:8080/",
				Fault: Fault{
					Enabled:     true,
					Probability: 0.5,
					Priority:    []string{"latency"},
					Types:       map[string]FaultConfig{"latency": {Delay: 100}},
				},
			},
		},
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestProxyValidate_DuplicateServiceName(t *testing.T) {
	p := Proxy{
		Service: []ServiceConfig{
			{Name: "svc-a", Fault: Fault{}},
			{Name: "svc-a", Fault: Fault{}}, // duplicate
		},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for duplicate service name")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestProxyValidate_EmptyServiceName(t *testing.T) {
	p := Proxy{
		Service: []ServiceConfig{
			{Name: "", Fault: Fault{}},
		},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestProxyValidate_PropagatesFaultError(t *testing.T) {
	p := Proxy{
		Service: []ServiceConfig{
			{
				Name: "bad-svc",
				Fault: Fault{
					Probability: 2.0, // invalid
				},
			},
		},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected Proxy.Validate to propagate Fault.Validate error")
	}
	if !strings.Contains(err.Error(), "bad-svc") {
		t.Errorf("error should mention service name, got: %v", err)
	}
}
