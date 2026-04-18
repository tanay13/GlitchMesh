package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ===========================================================================
// Layer 2/3: Integration Tests for the HTTP proxy client
// These tests spin up real httptest servers as mock backends and verify
// that ProxyRequest correctly forwards requests and responses.
// ===========================================================================

// TestProxyRequest_ForwardsResponseBody verifies that the full response body
// from the backend is forwarded to the client.
func TestProxyRequest_ForwardsResponseBody(t *testing.T) {
	// Set up a mock backend returning a known JSON body
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","data":[1,2,3]}`))
	}))
	defer backend.Close()

	// Create a test request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	statusCode, err := ProxyRequest(w, req, backend.URL+"/test")

	if err != nil {
		t.Fatalf("ProxyRequest returned error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("returned status = %d, want 200", statusCode)
	}

	// Verify body was forwarded
	body := w.Body.String()
	if body != `{"status":"ok","data":[1,2,3]}` {
		t.Errorf("body = %q, want JSON payload", body)
	}
}

// TestProxyRequest_ForwardsResponseHeaders verifies that response headers
// from the backend are copied to the client response.
func TestProxyRequest_ForwardsResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.Header().Set("X-Request-Id", "abc-123")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ProxyRequest(w, req, backend.URL+"/test")

	// Check that custom headers from backend were forwarded
	if w.Header().Get("X-Custom-Header") != "test-value" {
		t.Errorf("X-Custom-Header = %q, want %q", w.Header().Get("X-Custom-Header"), "test-value")
	}
	if w.Header().Get("X-Request-Id") != "abc-123" {
		t.Errorf("X-Request-Id = %q, want %q", w.Header().Get("X-Request-Id"), "abc-123")
	}
}

// TestProxyRequest_ForwardsStatusCode verifies that non-200 status codes
// from the backend are forwarded correctly (e.g., 404, 503).
func TestProxyRequest_ForwardsStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
		{"201 Created", http.StatusCreated},
		{"204 No Content", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer backend.Close()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			ProxyRequest(w, req, backend.URL+"/test")

			// The recorder should have the backend's status code
			if w.Code != tt.statusCode {
				t.Errorf("response status = %d, want %d", w.Code, tt.statusCode)
			}
		})
	}
}

// TestProxyRequest_ForwardsRequestBody verifies that POST/PUT request bodies
// are forwarded to the backend.
func TestProxyRequest_ForwardsRequestBody(t *testing.T) {
	var receivedBody string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body that was forwarded
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Create a POST request with a JSON body
	requestBody := `{"username":"testuser","email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ProxyRequest(w, req, backend.URL+"/test")

	if receivedBody != requestBody {
		t.Errorf("backend received body = %q, want %q", receivedBody, requestBody)
	}
}

// TestProxyRequest_BackendUnreachable verifies that the proxy returns
// StatusBadGateway (502) when the backend is down.
func TestProxyRequest_BackendUnreachable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Use a URL where nothing is listening
	statusCode, err := ProxyRequest(w, req, "http://localhost:1/unreachable")

	if err == nil {
		t.Error("expected error for unreachable backend")
	}
	if statusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 (Bad Gateway)", statusCode)
	}
}

// TestProxyRequest_ForwardsHTTPMethod verifies that the HTTP method
// (GET, POST, PUT, DELETE) is forwarded to the backend.
func TestProxyRequest_ForwardsHTTPMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var receivedMethod string
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer backend.Close()

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			ProxyRequest(w, req, backend.URL+"/test")

			if receivedMethod != method {
				t.Errorf("backend received method = %q, want %q", receivedMethod, method)
			}
		})
	}
}
