package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseConfigYaml verifies YAML parsing of proxy configuration files.
// Uses temp files so tests don't depend on project's test.yaml existing.
func TestParseConfigYaml(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		wantErr      bool
		wantServices int // how many services should be parsed
	}{
		{
			name: "valid single service config",
			yaml: `
service:
  - name: "test-service"
    url: "http://localhost:8080/"
    fault:
      enabled: true
      probability: 0.5
      priority: ["error"]
      types:
        error:
          statuscode: 500
          message: "boom"
`,
			wantErr:      false,
			wantServices: 1,
		},
		{
			name: "valid multi-service config",
			yaml: `
service:
  - name: "svc-a"
    url: "http://localhost:8080/"
    fault:
      enabled: false
  - name: "svc-b"
    url: "http://localhost:8081/"
    fault:
      enabled: true
      priority: ["latency"]
      types:
        latency:
          delay: 1000
`,
			wantErr:      false,
			wantServices: 2,
		},
		{
			name:    "invalid yaml — should error",
			yaml:    `{{{ not valid yaml`,
			wantErr: true,
		},
		{
			name:         "empty yaml — parses to zero services",
			yaml:         ``,
			wantErr:      false,
			wantServices: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp YAML file with the test content
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			proxy, err := ParseConfigYaml(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfigYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(proxy.Service) != tt.wantServices {
				t.Errorf("service count = %d, want %d", len(proxy.Service), tt.wantServices)
			}
		})
	}
}

// TestParseConfigYaml_FileNotFound verifies proper error when file doesn't exist.
func TestParseConfigYaml_FileNotFound(t *testing.T) {
	_, err := ParseConfigYaml("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// TestParseConfigYaml_EmptyPath verifies behavior with empty file path.
// The function calls log.Fatalf for empty paths, so we can't easily test
// it without refactoring. This test documents the current behavior.
// TODO: refactor ParseConfigYaml to return error instead of log.Fatalf for empty path

// TestParseConfigYaml_ServiceFieldsParsed verifies that individual
// service fields (name, url, fault config) are parsed correctly.
func TestParseConfigYaml_ServiceFieldsParsed(t *testing.T) {
	yaml := `
service:
  - name: "payment-svc"
    url: "http://localhost:9090/"
    fault:
      enabled: true
      probability: 0.75
      priority: ["latency", "error"]
      types:
        latency:
          delay: 3000
        error:
          statuscode: 503
          message: "service unavailable"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	os.WriteFile(path, []byte(yaml), 0644)

	proxy, err := ParseConfigYaml(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(proxy.Service) != 1 {
		t.Fatalf("expected 1 service, got %d", len(proxy.Service))
	}

	svc := proxy.Service[0]

	// Verify basic fields
	if svc.Name != "payment-svc" {
		t.Errorf("Name = %q, want %q", svc.Name, "payment-svc")
	}
	if svc.Url != "http://localhost:9090/" {
		t.Errorf("Url = %q, want %q", svc.Url, "http://localhost:9090/")
	}

	// Verify fault config
	if !svc.Fault.Enabled {
		t.Error("Fault.Enabled should be true")
	}
	if svc.Fault.Probability != 0.75 {
		t.Errorf("Probability = %f, want 0.75", svc.Fault.Probability)
	}

	// Verify priority order
	if len(svc.Fault.Priority) != 2 || svc.Fault.Priority[0] != "latency" || svc.Fault.Priority[1] != "error" {
		t.Errorf("Priority = %v, want [latency, error]", svc.Fault.Priority)
	}

	// Verify fault type configs
	latCfg := svc.Fault.Types["latency"]
	if latCfg.Delay != 3000 {
		t.Errorf("latency.Delay = %d, want 3000", latCfg.Delay)
	}
	errCfg := svc.Fault.Types["error"]
	if errCfg.StatusCode != 503 {
		t.Errorf("error.StatusCode = %d, want 503", errCfg.StatusCode)
	}
	if errCfg.Message != "service unavailable" {
		t.Errorf("error.Message = %q, want %q", errCfg.Message, "service unavailable")
	}
}
