package faults

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/shared/constants"
	"github.com/tanay13/GlitchMesh/internal/shared/models"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	net.Conn
	closed bool
	mu     sync.Mutex
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestConnectionDropFault_InjectFault(t *testing.T) {
	fault := &ConnectionDropFault{}
	resp := fault.InjectFault(context.Background())

	if !resp.Applied {
		t.Error("expected Applied to be true")
	}
	if resp.ShouldTerminate {
		t.Error("expected ShouldTerminate to be false for connection drop")
	}
	if resp.Fault != constants.CONNECTION_DROP {
		t.Errorf("expected fault type %s, got %s", constants.CONNECTION_DROP, resp.Fault)
	}
}

func TestConnectionRegistry_RegisterAndUnregister(t *testing.T) {
	registry := &ConnectionRegistry{
		conns: make(map[string][]net.Conn),
	}

	mc := &mockConn{}
	wrapped := registry.Register("test-service", mc)

	registry.mu.Lock()
	conns, ok := registry.conns["test-service"]
	registry.mu.Unlock()

	if !ok || len(conns) != 1 {
		t.Fatalf("expected 1 connection registered, got %v", conns)
	}

	// Closing the connection should automatically unregister it
	wrapped.Close()

	registry.mu.Lock()
	conns, ok = registry.conns["test-service"]
	registry.mu.Unlock()

	if len(conns) != 0 {
		t.Errorf("expected 0 connections registered after Close, got %v", conns)
	}
}

func TestConnectionRegistry_BackgroundDropper(t *testing.T) {
	// Set up config
	proxyCfg := &models.Proxy{
		Service: []models.ServiceConfig{
			{
				Name: "test-service",
				Url:  "http://localhost:8080/",
				Fault: models.Fault{
					Enabled:  true,
					Priority: []string{constants.CONNECTION_DROP},
					Types: map[string]models.FaultConfig{
						constants.CONNECTION_DROP: {
							DropRate: 1.0, // Drop everything
						},
					},
				},
			},
		},
	}
	config.SetProxyConfigForTesting(proxyCfg)

	// Registry we will use
	registry := Registry
	registry.mu.Lock()
	registry.conns = make(map[string][]net.Conn)
	registry.mu.Unlock()

	mc := &mockConn{}
	_ = registry.Register("test-service", mc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartBackgroundDropper(ctx)

	// Wait up to 500ms for the dropper loop to run and drop the connection
	ok := false
	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)
		if mc.IsClosed() {
			ok = true
			break
		}
	}

	if !ok {
		t.Error("expected connection to be closed by background dropper")
	}

	registry.mu.Lock()
	conns := registry.conns["test-service"]
	registry.mu.Unlock()

	if len(conns) != 0 {
		t.Errorf("expected connection to be removed from registry, got %d", len(conns))
	}
}
