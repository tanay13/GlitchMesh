// Package e2e contains end-to-end tests that exercise the full GlitchMesh
// request pipeline: HTTP request → router → proxy service → fault injector
// → multiple faults in sequence → backend (or termination).
//
// Unlike unit tests (which test one component in isolation) and integration
// tests (which test a few components together), e2e tests simulate real
// usage scenarios with multiple fault types configured simultaneously.
//
// Run with: go test -v -race ./internal/router/e2e_test.go ./internal/router/router.go
// Or from project root: make test-e2e
package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/config"
	"github.com/tanay13/GlitchMesh/internal/constants"
	"github.com/tanay13/GlitchMesh/internal/models"
	"github.com/tanay13/GlitchMesh/internal/testutil"
)

// ===========================================================================
// E2E Scenario: Latency + Error (latency first, then terminate with error)
// This tests the fault chain: add delay → THEN kill the request.
// Real-world analogy: simulate a slow-then-failing downstream service.
// ===========================================================================

func TestE2E_LatencyThenError(t *testing.T) {
	InitRouter()

	backendHit := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHit = true // should never be reached
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Priority: latency runs first (adds 50ms delay), then error terminates
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"slow-fail-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0, // always apply
			Priority:    []string{constants.LATENCY, constants.ERROR},
			Types: map[string]models.FaultConfig{
				constants.LATENCY: {Delay: 50},                            // 50ms delay
				constants.ERROR:   {StatusCode: 500, Message: "degraded"}, // then terminate
			},
		},
	))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/redirect/slow-fail-svc/api/data", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)
	elapsed := time.Since(start)

	// Should be terminated by error fault
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}

	// Backend should NOT be hit — error fault terminated before proxy
	if backendHit {
		t.Error("backend was contacted — error fault should terminate before reaching backend")
	}

	// Should have taken at least the latency delay (50ms)
	if elapsed < 40*time.Millisecond {
		t.Errorf("elapsed %v — latency fault should have added at least 40ms delay", elapsed)
	}

	// Verify error message in response body
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "degraded" {
		t.Errorf("error message = %q, want %q", body["error"], "degraded")
	}
}

// ===========================================================================
// E2E Scenario: Error + Latency (error first = latency never runs)
// Tests that chain terminates immediately on first terminating fault.
// ===========================================================================

func TestE2E_ErrorTerminatesBeforeLatency(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Error is first — should terminate immediately, 5s latency should NEVER run
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"fast-fail-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0,
			Priority:    []string{constants.ERROR, constants.LATENCY},
			Types: map[string]models.FaultConfig{
				constants.ERROR:   {StatusCode: 503, Message: "circuit open"},
				constants.LATENCY: {Delay: 5000}, // 5s — would be obvious if it ran
			},
		},
	))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/redirect/fast-fail-svc/api/check", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)
	elapsed := time.Since(start)

	// Error terminates
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}

	// Should finish FAST — latency never ran
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed %v — error should terminate chain, 5s latency should not run", elapsed)
	}
}

// ===========================================================================
// E2E Scenario: Timeout fault (blocks then terminates)
// Simulates a service that accepts the connection but never responds.
// ===========================================================================

func TestE2E_TimeoutFault(t *testing.T) {
	InitRouter()

	backendHit := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"hanging-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0,
			Priority:    []string{constants.TIMEOUT},
			Types: map[string]models.FaultConfig{
				constants.TIMEOUT: {TimeoutDuration: 60, StatusCode: 504}, // 60ms timeout
			},
		},
	))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/redirect/hanging-svc/api/slow", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)
	elapsed := time.Since(start)

	// Timeout terminates the request
	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d, want 504 Gateway Timeout", w.Code)
	}

	// Backend NOT hit — timeout terminated before forwarding
	if backendHit {
		t.Error("backend was contacted — timeout should terminate before reaching backend")
	}

	// Should have blocked for ~60ms (the timeout duration)
	if elapsed < 40*time.Millisecond {
		t.Errorf("elapsed %v — timeout fault should block for ~60ms", elapsed)
	}
}

// ===========================================================================
// E2E Scenario: Latency + Timeout (latency delays, then timeout terminates)
// Simulates slow AND hanging service.
// ===========================================================================

func TestE2E_LatencyThenTimeout(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Total delay: 50ms latency + 60ms timeout = ~110ms blocking minimum
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"very-slow-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0,
			Priority:    []string{constants.LATENCY, constants.TIMEOUT},
			Types: map[string]models.FaultConfig{
				constants.LATENCY: {Delay: 50},
				constants.TIMEOUT: {TimeoutDuration: 60, StatusCode: 504},
			},
		},
	))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/redirect/very-slow-svc/api/data", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)
	elapsed := time.Since(start)

	// Timeout terminates after latency delay
	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d, want 504", w.Code)
	}

	// Should have blocked for at least latency + timeout (~110ms)
	if elapsed < 90*time.Millisecond {
		t.Errorf("elapsed %v — expected latency(50ms) + timeout(60ms) = ~110ms", elapsed)
	}
}

// ===========================================================================
// E2E Scenario: Probability-based faults (only X% of requests get faults)
// Fires 1000 requests with 50% probability — verifies statistical distribution.
// ===========================================================================

func TestE2E_ProbabilityFault_StatisticalDistribution(t *testing.T) {
	InitRouter()

	// Count how many requests reach the backend (no fault applied)
	var mu sync.Mutex
	backendHits := 0

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		backendHits++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// 50% probability — roughly half requests should get the error fault
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"flaky-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0.5, // 50% of requests get faulted
			Priority:    []string{constants.ERROR},
			Types: map[string]models.FaultConfig{
				constants.ERROR: {StatusCode: 500, Message: "flaky"},
			},
		},
	))

	total := 1000
	for i := 0; i < total; i++ {
		req := httptest.NewRequest(http.MethodGet, "/redirect/flaky-svc/api/check", nil)
		w := httptest.NewRecorder()
		ProxyHandler(w, req)
	}

	// With p=0.5 and 1000 requests, backend should be hit ~500 times.
	// Allow 35%-65% range to avoid flaky tests.
	hitRatio := float64(backendHits) / float64(total)
	if hitRatio < 0.35 || hitRatio > 0.65 {
		t.Errorf("backend hit ratio = %.2f — expected ~0.50 (±15%%) for 50%% probability fault", hitRatio)
	}
}

// ===========================================================================
// E2E Scenario: Disabled faults — all requests pass through cleanly
// Verifies that fault config with enabled=false has zero effect.
// ===========================================================================

func TestE2E_DisabledFaults_AllRequestsPassThrough(t *testing.T) {
	InitRouter()

	backendHits := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendHits++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	// All fault types configured but Enabled=false — should have zero effect
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"stable-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     false, // disabled
			Probability: 1.0,   // would always fire if enabled
			Priority:    []string{constants.ERROR, constants.LATENCY, constants.TIMEOUT},
			Types: map[string]models.FaultConfig{
				constants.ERROR:   {StatusCode: 500, Message: "should not see this"},
				constants.LATENCY: {Delay: 5000},
				constants.TIMEOUT: {TimeoutDuration: 5000, StatusCode: 504},
			},
		},
	))

	total := 10
	for i := 0; i < total; i++ {
		req := httptest.NewRequest(http.MethodGet, "/redirect/stable-svc/api/ping", nil)
		w := httptest.NewRecorder()
		ProxyHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want 200 (faults are disabled)", i, w.Code)
		}
	}

	if backendHits != total {
		t.Errorf("backend hit %d times, want %d — all requests should pass through", backendHits, total)
	}
}

// ===========================================================================
// E2E Scenario: Multiple services with different fault configs
// Verifies that fault configs are correctly isolated per-service.
// ===========================================================================

func TestE2E_MultipleServices_IndependentFaultConfigs(t *testing.T) {
	InitRouter()

	// Backend A: always reachable
	backendA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("service-a"))
	}))
	defer backendA.Close()

	// Backend B: should never be reached (service-b has error fault)
	backendBHit := false
	backendB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendBHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer backendB.Close()

	// Two services: A has no faults, B has terminating error fault
	config.SetProxyConfigForTesting(testutil.NewTestProxyMultiService(
		models.ServiceConfig{
			Name:  "service-a",
			Url:   backendA.URL + "/",
			Fault: models.Fault{Enabled: false},
		},
		models.ServiceConfig{
			Name: "service-b",
			Url:  backendB.URL + "/",
			Fault: models.Fault{
				Enabled:     true,
				Probability: 0,
				Priority:    []string{constants.ERROR},
				Types: map[string]models.FaultConfig{
					constants.ERROR: {StatusCode: 500, Message: "b is down"},
				},
			},
		},
	))

	// Request to service-a should succeed
	reqA := httptest.NewRequest(http.MethodGet, "/redirect/service-a/api/ping", nil)
	wA := httptest.NewRecorder()
	ProxyHandler(wA, reqA)

	if wA.Code != http.StatusOK {
		t.Errorf("service-a status = %d, want 200", wA.Code)
	}
	if wA.Body.String() != "service-a" {
		t.Errorf("service-a body = %q, want 'service-a'", wA.Body.String())
	}

	// Request to service-b should be terminated by error fault
	reqB := httptest.NewRequest(http.MethodGet, "/redirect/service-b/api/ping", nil)
	wB := httptest.NewRecorder()
	ProxyHandler(wB, reqB)

	if wB.Code != http.StatusInternalServerError {
		t.Errorf("service-b status = %d, want 500", wB.Code)
	}
	if backendBHit {
		t.Error("service-b backend was hit — error fault should have terminated the request")
	}
}

// ===========================================================================
// E2E Scenario: Concurrent requests under fault injection
// Verifies thread safety: multiple goroutines firing through the proxy
// simultaneously with faults enabled should not race or deadlock.
// ===========================================================================

func TestE2E_ConcurrentRequests_NoRaceCondition(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Latency fault so goroutines overlap in time (not sequential)
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"concurrent-svc",
		backend.URL+"/",
		models.Fault{
			Enabled:     true,
			Probability: 0.5, // only half get faulted
			Priority:    []string{constants.LATENCY},
			Types: map[string]models.FaultConfig{
				constants.LATENCY: {Delay: 20}, // 20ms — forces goroutines to overlap
			},
		},
	))

	var wg sync.WaitGroup
	goroutines := 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/redirect/concurrent-svc/api/task", nil)
			w := httptest.NewRecorder()
			ProxyHandler(w, req)
			// No assertion on status — just verifying no race/panic
		}()
	}

	wg.Wait()
	// If we reach here without -race flagging anything, thread safety holds.
	// Run this test with: go test -race ./internal/router/...
}

// ===========================================================================
// E2E Scenario: Metrics accumulate correctly across multiple faulted requests
// Verifies that the metrics endpoint reflects the actual faults injected.
// ===========================================================================

func TestE2E_MetricsAccumulate(t *testing.T) {
	InitRouter()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Always inject error fault
	config.SetProxyConfigForTesting(testutil.NewTestProxy(
		"tracked-svc",
		backend.URL+"/",
		testutil.NewFaultConfig(constants.ERROR, true),
	))

	// Fire 5 requests — all should get error fault
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/redirect/tracked-svc/api/op", nil)
		w := httptest.NewRecorder()
		ProxyHandler(w, req)
	}

	// Check metrics endpoint reflects the injected faults
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	MetricsHandler(metricsW, metricsReq)

	var metricsResp models.Metrics
	if err := json.NewDecoder(metricsW.Body).Decode(&metricsResp); err != nil {
		t.Fatalf("failed to decode metrics: %v", err)
	}

	// At least 5 errors should be recorded (may be more if other tests ran first)
	if metricsResp.FaultMetrics.ErrorCount < 5 {
		t.Errorf("ErrorCount = %d, want >= 5", metricsResp.FaultMetrics.ErrorCount)
	}
	if metricsResp.FaultMetrics.TotalFaultsInjected < 5 {
		t.Errorf("TotalFaultsInjected = %d, want >= 5", metricsResp.FaultMetrics.TotalFaultsInjected)
	}
}
