package faults

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/tanay13/GlitchMesh/internal/controlplane/config"
	"github.com/tanay13/GlitchMesh/internal/dataplane/metrics"
	"github.com/tanay13/GlitchMesh/internal/shared/constants"
	"github.com/tanay13/GlitchMesh/internal/shared/utils"
)

type ConnectionRegistry struct {
	mu    sync.Mutex
	conns map[string][]net.Conn
}

var Registry = &ConnectionRegistry{
	conns: make(map[string][]net.Conn),
}

func (r *ConnectionRegistry) Register(service string, conn net.Conn) net.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()

	wrapped := &trackedConn{
		Conn:     conn,
		service:  service,
		registry: r,
	}
	r.conns[service] = append(r.conns[service], wrapped)
	return wrapped
}

func (r *ConnectionRegistry) Unregister(service string, conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	conns := r.conns[service]
	for i, c := range conns {
		if c == conn {
			r.conns[service] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

type trackedConn struct {
	net.Conn
	service  string
	registry *ConnectionRegistry
	once     sync.Once
}

func (c *trackedConn) Close() error {
	var err error
	// we do this because we want connection to only close once
	c.once.Do(func() {
		c.registry.Unregister(c.service, c)
		err = c.Conn.Close()
	})
	return err
}

func StartBackgroundDropper(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				r := Registry
				proxyConfig := config.GetProxyConfig()
				if proxyConfig == nil {
					continue
				}

				r.mu.Lock()
				// Gather services that have connection drop configured and have active connections
				for serviceName, conns := range r.conns {
					if len(conns) == 0 {
						continue
					}

					serviceConfig := utils.GetServiceConfig(serviceName, proxyConfig)
					if serviceConfig == nil || !serviceConfig.Fault.Enabled {
						continue
					}

					hasDropFault := false
					for _, p := range serviceConfig.Fault.Priority {
						if p == constants.CONNECTION_DROP {
							hasDropFault = true
							break
						}
					}

					if !hasDropFault {
						continue
					}

					dropCfg, ok := serviceConfig.Fault.Types[constants.CONNECTION_DROP]
					dropRate := 0.1 // default 10%
					if ok && dropCfg.DropRate > 0 {
						dropRate = dropCfg.DropRate
					}

					var connsToClose []net.Conn
					for _, conn := range conns {
						if rand.Float64() < dropRate {
							connsToClose = append(connsToClose, conn)
						}
					}

					// Close connections outside registry mutex to prevent deadlocks/blocking
					if len(connsToClose) > 0 {
						go func(toClose []net.Conn, svc string) {
							for _, c := range toClose {
								c.Close()
								// Increment drop metrics
								metrics.RegisteredMetrics[constants.FAULT_METRICS].Increment(constants.TOTAL_FAULTS_INJECTED, 1)
								metrics.RegisteredMetrics[constants.FAULT_METRICS].Increment(constants.CONNECTION_DROP, 1)
							}
						}(connsToClose, serviceName)
					}
				}
				r.mu.Unlock()

			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
