package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/models"
)

// TestLoad verifies that JSON config files are parsed correctly.
// Uses temp files so tests are isolated from project config.
func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		wantYAMLPath string
	}{
		{
			name:         "valid config with yaml path",
			content:      `{"env": {"yaml_file_path": "./services.yaml"}}`,
			wantErr:      false,
			wantYAMLPath: "./services.yaml",
		},
		{
			name:    "invalid JSON — should error",
			content: `{broken json`,
			wantErr: true,
		},
		{
			name:         "empty env — parses with zero values",
			content:      `{"env": {}}`,
			wantErr:      false,
			wantYAMLPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test content to a temp file
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			cfg, err := Load(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Fatal("Load() returned nil config without error")
				}
				if cfg.Env.YAML_FILE_PATH != tt.wantYAMLPath {
					t.Errorf("YAML_FILE_PATH = %q, want %q", cfg.Env.YAML_FILE_PATH, tt.wantYAMLPath)
				}
			}
		})
	}
}

// TestLoad_FileNotFound verifies proper error message on missing file.
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.json")
	if err == nil {
		t.Error("Load() should error for missing file")
	}
}

// TestSetProxyConfigForTesting verifies that the test helper
// correctly sets the proxy config accessible via GetProxyConfig.
func TestSetProxyConfigForTesting(t *testing.T) {
	// GetProxyConfig before setting — should be nil if nothing loaded
	// (depends on test order, so just test the set/get round-trip)

	testProxy := createTestProxyConfig()
	SetProxyConfigForTesting(testProxy)

	got := GetProxyConfig()
	if got == nil {
		t.Fatal("GetProxyConfig() returned nil after SetProxyConfigForTesting")
	}
	if len(got.Service) != 1 {
		t.Errorf("service count = %d, want 1", len(got.Service))
	}
	if got.Service[0].Name != "test-svc" {
		t.Errorf("service name = %q, want %q", got.Service[0].Name, "test-svc")
	}
}

// createTestProxyConfig is a helper to build a minimal proxy config for tests.
func createTestProxyConfig() *models.Proxy {
	return &models.Proxy{
		Service: []models.ServiceConfig{
			{
				Name: "test-svc",
				Url:  "http://localhost:9999/",
				Fault: models.Fault{
					Enabled: false,
				},
			},
		},
	}
}
