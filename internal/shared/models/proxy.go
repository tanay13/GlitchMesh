package models

import (
	"fmt"
	"strings"
)

// KnownFaultTypes is the set of fault type names the injector understands.
// Kept here (not in constants) to avoid an import cycle with shared/constants.
var KnownFaultTypes = map[string]bool{
	"latency":         true,
	"error":           true,
	"timeout":         true,
	"connection_drop": true,
}

type Proxy struct {
	Service []ServiceConfig
}

type ServiceConfig struct {
	Name  string
	Url   string
	Fault Fault
}

type Fault struct {
	Enabled     bool
	Priority    []string
	Probability float64
	Types       map[string]FaultConfig
}

type FaultConfig struct {
	StatusCode      int
	Message         string
	Delay           int
	TimeoutDuration int
	DropRate        float64 `yaml:"droprate"`
	DropInterval    int     `yaml:"dropinterval"`
}

// Validate checks that a Fault config is internally consistent
func (f *Fault) Validate() error {
	if f.Probability < 0 || f.Probability > 1 {
		return fmt.Errorf("fault.probability must be between 0.0 and 1.0, got %.4f", f.Probability)
	}

	var errs []string

	// Every type key must be a known fault type
	for typeName := range f.Types {
		if !KnownFaultTypes[typeName] {
			errs = append(errs, fmt.Sprintf("unknown fault type %q in types", typeName))
		}
	}

	// Every priority entry must have a corresponding Types entry
	for _, p := range f.Priority {
		if !KnownFaultTypes[p] {
			errs = append(errs, fmt.Sprintf("unknown fault type %q in priority", p))
		} else if _, ok := f.Types[p]; !ok {
			errs = append(errs, fmt.Sprintf("priority entry %q has no matching entry in types", p))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid fault config: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Validate checks a Proxy config, validates each service's fault and rejects duplicate names
func (p *Proxy) Validate() error {
	seen := make(map[string]bool, len(p.Service))
	var errs []string

	for i, svc := range p.Service {
		if svc.Name == "" {
			errs = append(errs, fmt.Sprintf("service[%d]: name must not be empty", i))
			continue
		}
		if seen[svc.Name] {
			errs = append(errs, fmt.Sprintf("service[%d]: duplicate service name %q", i, svc.Name))
		}
		seen[svc.Name] = true

		if err := svc.Fault.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("service %q: %v", svc.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid proxy config: %s", strings.Join(errs, "; "))
	}
	return nil
}
