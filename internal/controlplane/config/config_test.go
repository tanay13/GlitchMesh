package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

// TestLoad verifies that JSON config files are parsed correctly.
// Uses temp files so tests are isolated from project config.
func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantErr      bool
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

// ---------------------------------------------------------------------------
// Override layer tests
// ---------------------------------------------------------------------------

func TestSetAndGetServiceOverride(t *testing.T) {
	// Set up a base proxy config.
	SetProxyConfigForTesting(&models.Proxy{
		Service: []models.ServiceConfig{
			{Name: "svc-a", Url: "http://backend/", Fault: models.Fault{Enabled: false}},
		},
	})

	override := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{"error"},
		Types:       map[string]models.FaultConfig{"error": {StatusCode: 503, Message: "override"}},
	}

	if err := SetServiceOverride("svc-a", override); err != nil {
		t.Fatalf("SetServiceOverride failed: %v", err)
	}

	got := GetEffectiveServiceConfig("svc-a")
	if got == nil {
		t.Fatal("GetEffectiveServiceConfig returned nil")
	}
	if !got.Fault.Enabled {
		t.Error("override should have Enabled=true")
	}
	if got.Fault.Types["error"].StatusCode != 503 {
		t.Errorf("expected StatusCode=503, got %d", got.Fault.Types["error"].StatusCode)
	}
}

func TestDeleteServiceOverride_RevertsToBase(t *testing.T) {
	SetProxyConfigForTesting(&models.Proxy{
		Service: []models.ServiceConfig{
			{Name: "svc-b", Url: "http://backend/", Fault: models.Fault{Enabled: false}},
		},
	})

	override := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{"error"},
		Types:       map[string]models.FaultConfig{"error": {StatusCode: 500}},
	}
	if err := SetServiceOverride("svc-b", override); err != nil {
		t.Fatalf("SetServiceOverride: %v", err)
	}

	DeleteServiceOverride("svc-b")

	got := GetEffectiveServiceConfig("svc-b")
	if got == nil {
		t.Fatal("GetEffectiveServiceConfig returned nil after reset")
	}
	if got.Fault.Enabled {
		t.Error("expected base config (Enabled=false) after override deleted")
	}
}

func TestGetEffectiveServiceConfig_UnknownService(t *testing.T) {
	SetProxyConfigForTesting(&models.Proxy{Service: []models.ServiceConfig{}})
	got := GetEffectiveServiceConfig("does-not-exist")
	if got != nil {
		t.Error("expected nil for unknown service")
	}
}

func TestSetServiceOverride_InvalidFaultRejected(t *testing.T) {
	err := SetServiceOverride("svc-x", models.Fault{Probability: 99})
	if err == nil {
		t.Error("expected error for invalid fault override")
	}
}

func TestGetConfigGeneration_IncreasesOnReload(t *testing.T) {
	before := GetConfigGeneration()

	// Write a minimal valid YAML to a temp file and do a full Load + ProxyLoad cycle.
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "proxy.yaml")
	yaml := `service:
  - name: "reload-svc"
    url: "http://localhost:8080/"
    fault:
      enabled: false
`
	if err := os.WriteFile(yamlPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	jsonPath := filepath.Join(dir, "config.json")
	jsonContent := `{"env": {"yaml_file_path": "` + yamlPath + `"}}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	if _, err := Load(jsonPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := ProxyLoad(); err != nil {
		t.Fatalf("ProxyLoad: %v", err)
	}

	after := GetConfigGeneration()
	if after <= before {
		t.Errorf("generation should have increased: before=%d after=%d", before, after)
	}
}

// ---------------------------------------------------------------------------
// Hot reload (fsnotify watcher) integration test
// ---------------------------------------------------------------------------

// TestStartWatcher_PicksUpFileChange writes a YAML with faults disabled,
// starts the watcher, rewrites the file with faults enabled, and asserts that
// GetProxyConfig reflects the new config within 500ms.
func TestStartWatcher_PicksUpFileChange(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "proxy.yaml")
	jsonPath := filepath.Join(dir, "config.json")

	writeYAML := func(enabled bool) {
		t.Helper()
		enabledStr := "false"
		if enabled {
			enabledStr = "true"
		}
		yaml := `service:
  - name: "watch-svc"
    url: "http://localhost:8080/"
    fault:
      enabled: ` + enabledStr + `
`
		if err := os.WriteFile(yamlPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("writeYAML: %v", err)
		}
	}

	// Initial YAML: faults disabled.
	writeYAML(false)

	jsonContent := `{"env": {"yaml_file_path": "` + yamlPath + `"}}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if _, err := Load(jsonPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := ProxyLoad(); err != nil {
		t.Fatalf("initial ProxyLoad: %v", err)
	}

	initialGen := GetConfigGeneration()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := StartWatcher(ctx, yamlPath); err != nil {
		t.Fatalf("StartWatcher: %v", err)
	}

	// Rewrite YAML with faults enabled.
	time.Sleep(50 * time.Millisecond) // slight pause so watcher is ready
	writeYAML(true)

	// Poll for up to 500ms for the generation counter to increase.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if GetConfigGeneration() > initialGen {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if GetConfigGeneration() <= initialGen {
		t.Fatalf("generation did not increase after file change (initial=%d, current=%d)",
			initialGen, GetConfigGeneration())
	}

	// Verify the new config is actually in memory.
	cfg := GetProxyConfig()
	if cfg == nil {
		t.Fatal("GetProxyConfig returned nil after hot reload")
	}
	if len(cfg.Service) == 0 || !cfg.Service[0].Fault.Enabled {
		t.Error("hot-reload did not pick up Fault.Enabled=true from rewritten YAML")
	}
}

// TestRegisterReloadCallback fires on hot reload.
func TestRegisterReloadCallback(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "proxy.yaml")
	jsonPath := filepath.Join(dir, "config.json")

	yaml := `service:
  - name: "cb-svc"
    url: "http://localhost:8080/"
    fault:
      enabled: false
`
	os.WriteFile(yamlPath, []byte(yaml), 0644)
	os.WriteFile(jsonPath, []byte(`{"env": {"yaml_file_path": "`+yamlPath+`"}}`), 0644)
	Load(jsonPath)

	var callbackGen int64
	called := make(chan struct{}, 1)
	RegisterReloadCallback(func(gen int64) {
		callbackGen = gen
		select {
		case called <- struct{}{}:
		default:
		}
	})

	ProxyLoad()

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("reload callback was not invoked after ProxyLoad()")
	}

	if callbackGen != GetConfigGeneration() {
		t.Errorf("callback gen=%d, want %d", callbackGen, GetConfigGeneration())
	}
}
