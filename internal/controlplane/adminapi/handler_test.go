package adminapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

// setupTestProxy sets a simple in-memory proxy config for tests.
func setupTestProxy(t *testing.T) {
	t.Helper()
	config.SetProxyConfigForTesting(&models.Proxy{
		Service: []models.ServiceConfig{
			{
				Name: "svc-alpha",
				Url:  "http://alpha/",
				Fault: models.Fault{
					Enabled:     false,
					Probability: 0,
				},
			},
			{
				Name: "svc-beta",
				Url:  "http://beta/",
				Fault: models.Fault{
					Enabled:     true,
					Probability: 1.0,
					Priority:    []string{"error"},
					Types:       map[string]models.FaultConfig{"error": {StatusCode: 500, Message: "beta error"}},
				},
			},
		},
	})
	// Clear any overrides from previous tests.
	config.DeleteServiceOverride("svc-alpha")
	config.DeleteServiceOverride("svc-beta")
}

// ─── Auth tests ───────────────────────────────────────────────────────────────

func TestAuth_Rejected_WhenTokenWrong(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret123")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	req.Header.Set("Authorization", "Bearer wrongtoken")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuth_Rejected_WhenHeaderMissing(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret123")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuth_Allowed_WhenTokenCorrect(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret123")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuth_Skipped_WhenNoTokenConfigured(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	// No Authorization header — should pass in dev mode.
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (dev mode, no token)", w.Code)
	}
}

// ─── GET /admin/services ──────────────────────────────────────────────────────

func TestGetServices_ReturnsList(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp serviceListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Services) != 2 {
		t.Errorf("services count = %d, want 2", len(resp.Services))
	}
}

func TestGetServices_MethodNotAllowed(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/services", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

// ─── PATCH /admin/services/{name}/faults ─────────────────────────────────────

func TestPatchFaults_AppliesOverride(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	override := models.Fault{
		Enabled:     true,
		Probability: 0,
		Priority:    []string{"error"},
		Types:       map[string]models.FaultConfig{"error": {StatusCode: 503, Message: "admin override"}},
	}
	body, _ := json.Marshal(override)

	req := httptest.NewRequest(http.MethodPatch, "/admin/services/svc-alpha/faults",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	// Verify override is reflected in GetEffectiveServiceConfig.
	svc := config.GetEffectiveServiceConfig("svc-alpha")
	if svc == nil {
		t.Fatal("effective config is nil after override")
	}
	if svc.Fault.Types["error"].StatusCode != 503 {
		t.Errorf("effective StatusCode = %d, want 503", svc.Fault.Types["error"].StatusCode)
	}
}

func TestPatchFaults_UnknownService_404(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	body, _ := json.Marshal(models.Fault{})
	req := httptest.NewRequest(http.MethodPatch, "/admin/services/does-not-exist/faults",
		bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestPatchFaults_InvalidFault_400(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// Probability out of range → validation error.
	override := models.Fault{Probability: 5.0}
	body, _ := json.Marshal(override)
	req := httptest.NewRequest(http.MethodPatch, "/admin/services/svc-alpha/faults",
		bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestPatchFaults_BadJSON_400(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPatch, "/admin/services/svc-alpha/faults",
		bytes.NewReader([]byte("{broken json")))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ─── POST /admin/services/{name}/faults/reset ────────────────────────────────

func TestFaultReset_RevertsOverride(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	// Set an override first.
	override := models.Fault{
		Enabled: true, Probability: 0,
		Priority: []string{"error"},
		Types:    map[string]models.FaultConfig{"error": {StatusCode: 503}},
	}
	config.SetServiceOverride("svc-alpha", override)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/services/svc-alpha/faults/reset", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	// After reset, effective config should be base YAML (Enabled=false).
	svc := config.GetEffectiveServiceConfig("svc-alpha")
	if svc == nil {
		t.Fatal("GetEffectiveServiceConfig returned nil after reset")
	}
	if svc.Fault.Enabled {
		t.Error("after reset, Enabled should be false (base YAML)")
	}
}

func TestFaultReset_UnknownService_404(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/services/nope/faults/reset", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// ─── GET /admin/config/diff ───────────────────────────────────────────────────

func TestConfigDiff_NoDrift(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/config/diff", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp diffResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Drifted) != 0 {
		t.Errorf("drifted count = %d, want 0 (no overrides set)", len(resp.Drifted))
	}
}

func TestConfigDiff_ShowsDriftedServices(t *testing.T) {
	os.Unsetenv("ADMIN_TOKEN")
	setupTestProxy(t)

	override := models.Fault{
		Enabled: true, Probability: 0,
		Priority: []string{"error"},
		Types:    map[string]models.FaultConfig{"error": {StatusCode: 503}},
	}
	config.SetServiceOverride("svc-alpha", override)
	defer config.DeleteServiceOverride("svc-alpha")

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/config/diff", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp diffResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Drifted) != 1 {
		t.Errorf("drifted count = %d, want 1", len(resp.Drifted))
	}
	if resp.Drifted[0].ServiceName != "svc-alpha" {
		t.Errorf("drifted service = %q, want 'svc-alpha'", resp.Drifted[0].ServiceName)
	}
}
