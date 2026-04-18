package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
	"github.com/tanay13/GlitchMesh/internal/testutil"
)

// Layer 2: Integration Tests
// These tests wire together multiple components (router → service → faults)
// and test real HTTP request/response flows using httptest.

// TestHomeHandler verifies the root endpoint returns a welcome message.
func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HomeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "Hi from GlitchMesh!" {
		t.Errorf("body = %q, want %q", w.Body.String(), "Hi from GlitchMesh!")
	}
}

// TestMetricsHandler verifies the /metrics endpoint returns valid JSON
// with the expected structure.
func TestMetricsHandler(t *testing.T) {

	InitRouter()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp models.Metrics
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v\nbody: %s", err, w.Body.String())
	}
}

// TestProxyHandler_InvalidPaths verifies that malformed URLs return 400.
// These don't need a backend server since they fail at URL validation.
func TestProxyHandler_InvalidPaths(t *testing.T) {
	InitRouter()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "empty after redirect prefix",
			path:       "/redirect/",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service name only, no endpoint",
			path:       "/redirect/myservice",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "completely wrong path (no /redirect/ prefix)",
			path:       "/something-else",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			ProxyHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d\nbody: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// TestProxyHandler_UnknownService verifies that requests to unconfigured
// services return 500 with an error message.
func TestProxyHandler_UnknownService(t *testing.T) {
	InitRouter()

	// Set up config with only "known-svc"
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"known-svc",
		"http://localhost:9999/",
		models.Fault{Enabled: false},
	))

	// Request for "unknown-svc" which isn't configured
	req := httptest.NewRequest(http.MethodGet, "/redirect/unknown-svc/api/test", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for unknown service", w.Code)
	}
}

// TestProxyHandler_ErrorFaultTerminatesRequest verifies that when an error
// fault is configured and enabled, the proxy returns the configured error
// status code WITHOUT forwarding the request to the backend.
func TestProxyHandler_ErrorFaultTerminatesRequest(t *testing.T) {
	InitRouter()

	// Track whether the backend was actually contacted
	backendHit := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Configure service with error fault enabled
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"test-svc",
		backend.URL+"/",
		testutil.NewFaultConfig(constants.ERROR, true),
	))

	req := httptest.NewRequest(http.MethodGet, "/redirect/test-svc/api/data", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	// Error fault should terminate the request with configured status code
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 (error fault configured with 500)", w.Code)
	}

	// Backend should NOT have been hit since error fault terminates early
	if backendHit {
		t.Error("backend was contacted, but error fault should have terminated the request")
	}
}

// TestProxyHandler_NoFaults_ProxiesToBackend verifies that with faults disabled,
// the proxy correctly forwards requests to the backend and returns its response.
func TestProxyHandler_NoFaults_ProxiesToBackend(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("backend received path = %q, want /api/users", r.URL.Path)
		}
		w.Header().Set("X-Backend", "test")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	}))
	defer backend.Close()

	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"user-svc",
		backend.URL+"/",
		models.Fault{Enabled: false},
	))

	req := httptest.NewRequest(http.MethodGet, "/redirect/user-svc/api/users", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	if w.Header().Get("X-Backend") != "test" {
		t.Error("backend response headers not proxied")
	}

	if w.Body.String() != `{"users": []}` {
		t.Errorf("body = %q, want %q", w.Body.String(), `{"users": []}`)
	}
}

// TestProxyHandler_ForwardsQueryParams verifies that query parameters
// from the original request are forwarded to the backend.
func TestProxyHandler_ForwardsQueryParams(t *testing.T) {
	InitRouter()

	var receivedQuery string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"api-svc",
		backend.URL+"/",
		models.Fault{Enabled: false},
	))

	req := httptest.NewRequest(http.MethodGet, "/redirect/api-svc/search?q=hello&limit=10", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	if receivedQuery != "q=hello&limit=10" {
		t.Errorf("query = %q, want %q", receivedQuery, "q=hello&limit=10")
	}
}

// TestProxyHandler_LatencyFault_RequestStillCompletes verifies that
// latency faults add delay but still forward the request to the backend
// (unlike error faults which terminate).
func TestProxyHandler_LatencyFault_RequestStillCompletes(t *testing.T) {
	InitRouter()

	backendHit := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHit = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("delayed but delivered"))
	}))
	defer backend.Close()

	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"slow-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0,
			Priority:    []string{constants.LATENCY},
			Types: map[string]models.FaultConfig{
				constants.LATENCY: {Delay: 50}, // 50ms delay
			},
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/redirect/slow-svc/api/data", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	if !backendHit {
		t.Error("backend was not contacted — latency fault should delay, not terminate")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
