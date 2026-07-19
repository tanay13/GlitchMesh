// Package server — hot reload e2e test.
//
// This test verifies the full round-trip:
//  1. Start the fsnotify watcher on a temp YAML file (faults disabled).
//  2. Fire a proxy request — it passes through to the backend (200 OK).
//  3. Rewrite the YAML on disk to enable an error fault (503).
//  4. Wait for the watcher to pick up the change (< 500 ms).
//  5. Fire another proxy request — it is now terminated with 503 by the fault.
//
// Run with: go test -v -race -run TestHotReload ./internal/dataplane/server/...
package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

func TestHotReload_ConfigChangePickedUpWithinDeadline(t *testing.T) {
	InitRouter()

	// Backend: just returns 200 so we can distinguish "fault terminated" from "passed through".
	backendHits := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHits++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	// ── Write initial YAML (faults disabled) ────────────────────────────────
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "proxy.yaml")

	yamlDisabled := `service:
  - name: "hot-svc"
    url: "` + backend.URL + `/"`  + `
    fault:
      enabled: false
`
	if err := os.WriteFile(yamlPath, []byte(yamlDisabled), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	jsonPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(jsonPath,
		[]byte(`{"env": {"yaml_file_path": "`+yamlPath+`"}}`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	if _, err := config.Load(jsonPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := config.ProxyLoad(); err != nil {
		t.Fatalf("initial ProxyLoad: %v", err)
	}

	// ── Start hot-reload watcher ─────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := config.StartWatcher(ctx, yamlPath); err != nil {
		t.Fatalf("StartWatcher: %v", err)
	}

	time.Sleep(30 * time.Millisecond) // ensure watcher goroutine is ready

	// ── First request — should pass through (faults disabled) ───────────────
	req1 := httptest.NewRequest(http.MethodGet, "/redirect/hot-svc/api/ping", nil)
	w1 := httptest.NewRecorder()
	ProxyHandler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want 200 (faults disabled)", w1.Code)
	}
	if backendHits != 1 {
		t.Errorf("first request: backendHits = %d, want 1", backendHits)
	}

	// ── Rewrite YAML with error fault enabled ────────────────────────────────
	initialGen := config.GetConfigGeneration()

	yamlEnabled := `service:
  - name: "hot-svc"
    url: "` + backend.URL + `/"`  + `
    fault:
      enabled: true
      probability: 0
      priority:
        - error
      types:
        error:
          statuscode: 503
          message: "hot-reload-fault"
`
	if err := os.WriteFile(yamlPath, []byte(yamlEnabled), 0644); err != nil {
		t.Fatalf("write yaml (enabled): %v", err)
	}

	// ── Wait for watcher to pick up the change (up to 500 ms) ───────────────
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if config.GetConfigGeneration() > initialGen {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if config.GetConfigGeneration() <= initialGen {
		t.Fatal("hot reload did not happen within 500ms of file change")
	}

	// ── Second request — should now be faulted (503) ─────────────────────────
	req2 := httptest.NewRequest(http.MethodGet, "/redirect/hot-svc/api/ping", nil)
	w2 := httptest.NewRecorder()
	ProxyHandler(w2, req2)

	if w2.Code != http.StatusServiceUnavailable {
		t.Errorf("second request: status = %d, want 503 (fault active after hot reload)", w2.Code)
	}

	// Backend should NOT have been hit by the faulted request.
	if backendHits != 1 {
		t.Errorf("backend was hit %d times total; second request should have been fault-terminated", backendHits)
	}
}

// TestHotReload_AdminOverrideNotClobberedByReload verifies that a live admin
// API override for a service is NOT overwritten when the YAML file is reloaded.
func TestHotReload_AdminOverrideNotClobberedByReload(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "proxy.yaml")
	yaml := `service:
  - name: "override-svc"
    url: "` + backend.URL + `/"`  + `
    fault:
      enabled: false
`
	os.WriteFile(yamlPath, []byte(yaml), 0644)
	jsonPath := filepath.Join(dir, "config.json")
	os.WriteFile(jsonPath, []byte(`{"env": {"yaml_file_path": "`+yamlPath+`"}}`), 0644)
	config.Load(jsonPath)
	config.ProxyLoad()

	// Apply an admin API override (error fault).
	override := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{"error"},
		Types:       map[string]models.FaultConfig{"error": {StatusCode: 503, Message: "override-fault"}},
	}
	if err := config.SetServiceOverride("override-svc", override); err != nil {
		t.Fatalf("SetServiceOverride: %v", err)
	}
	defer config.DeleteServiceOverride("override-svc")

	// Trigger a hot reload (rewrite YAML with same content).
	config.ProxyLoad() // simulates watcher triggering a reload

	// Override should still be present after reload.
	svc := config.GetEffectiveServiceConfig("override-svc")
	if svc == nil {
		t.Fatal("GetEffectiveServiceConfig returned nil")
	}
	if !svc.Fault.Enabled {
		t.Error("admin override was clobbered by hot-reload; Fault.Enabled should still be true")
	}
	if svc.Fault.Types["error"].StatusCode != 503 {
		t.Errorf("StatusCode = %d; admin override should survive hot-reload", svc.Fault.Types["error"].StatusCode)
	}
}
